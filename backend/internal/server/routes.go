package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/raziel-aleman/go-todo-app/internal/auth"
	m "github.com/raziel-aleman/go-todo-app/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/markbates/goth/gothic"
)

type ctxKey string

const providerKey ctxKey = "provider"

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "All"},
		AllowedMethods:   []string{"GET", "PATCH", "POST"},
		AllowCredentials: true,
	}))

	r.Use(middleware.Logger)

	r.Get("/health", s.healthHandler)

	r.Get("/auth/{provider}/callback", s.getAuthCallbackHandler)

	r.Get("/auth/{provider}", s.getAuthLoginHandler)

	r.Get("/auth/logout/{provider}", s.getAuthLogoutHandler)

	r.Get("/auth/validate", s.validateUserSessionHandler)

	r.Get("/api/todos", auth.RequireAuth(s.getAllTodosHandler))

	r.Post("/api/todos", auth.RequireAuth(s.createTodoHandler))

	r.Patch("/api/todos/{id}/done", auth.RequireAuth(s.markTodoDoneHandler))

	r.Patch("/api/todos/{id}/edit", auth.RequireAuth(s.editTodoHandler))

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}

func (s *Server) getAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", chi.URLParam(r, "provider"))
	r.URL.RawQuery = q.Encode()

	// // make provider available to the handler
	// provider := chi.URLParam(r, "provider")
	// r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	gothic.BeginAuthHandler(w, r)
}

func (s *Server) getAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// make provider available to the handler
	providerValue := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), providerKey, providerValue))

	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, "Error completing user auth\n", r)
		return
	}

	sessionId, err := auth.StoreUserSession(w, r, user)

	if err != nil {
		log.Println(err)
		return
	}

	msg, err := s.db.SaveUser(user, sessionId)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println(msg)
	}

	http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
}

func (s *Server) getAuthLogoutHandler(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)

	err := auth.RemoveUserSession(w, r)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Location", "http://localhost:3000/login")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) getAllTodosHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetUserSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	userId, err := s.db.IsSessionIdValid(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	rows, _ := s.db.GetAll(userId)

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) markTodoDoneHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetUserSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	userId, err := s.db.IsSessionIdValid(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	paramId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(paramId)

	if err != nil {
		log.Fatalf("Invalid ID. Err: %v", err)
	}

	s.db.MarkDone(int64(id))

	rows, _ := s.db.GetAll(userId)

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) createTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetUserSession(r)
	if err != nil {
		log.Fatalln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	userId, err := s.db.IsSessionIdValid(sessionId)
	if err != nil {
		log.Fatalln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var body m.NewTodo
	json.NewDecoder(r.Body).Decode(&body)

	_, err = s.db.Create(body, userId)

	if err != nil {
		log.Fatalf("error creating new todo. Err: %v", err)
	}

	rows, _ := s.db.GetAll(userId)

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) editTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetUserSession(r)
	if err != nil {
		log.Fatalln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	userId, err := s.db.IsSessionIdValid(sessionId)
	if err != nil {
		log.Fatalln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	paramId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(paramId)

	if err != nil {
		log.Fatalf("Invalid ID. Err: %v", err)
	}

	var body m.NewTodo
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	err = s.db.Edit(id, body, userId)

	if err != nil {
		log.Fatal(err)
	}

	rows, _ := s.db.GetAll(userId)

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) validateUserSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetUserSession(r)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{})
		w.Write(jsonResp)
		return
	}

	userId, err := s.db.IsSessionIdValid(sessionId)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{})
		w.Write(jsonResp)
		return
	}

	log.Println("session id validated")

	jsonResp, err := json.Marshal(map[string]string{"userId": userId})
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	w.Write(jsonResp)
}
