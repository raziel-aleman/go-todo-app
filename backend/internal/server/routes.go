package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/raziel-aleman/go-todo-app/internal/auth"
	"github.com/raziel-aleman/go-todo-app/internal/models"

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
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "All"},
		AllowedMethods: []string{"GET", "PATCH", "POST"},
	}))

	r.Use(middleware.Logger)

	r.Get("/health", s.healthHandler)

	r.Get("/auth/{provider}/callback", s.getAuthCallbackFunction)

	r.Get("/auth/{provider}", s.getAuthLoginFunction)

	r.Get("/auth/logout/{provider}", s.getAuthLogoutFunction)

	r.Get("/api/todos", auth.RequireAuth(s.getAllHandler))

	r.Patch("/api/todos/{id}/done", auth.RequireAuth(s.markDoneHandler))

	r.Post("/api/todos", auth.RequireAuth(s.createHandler))

	r.Patch("/api/todos/{id}/edit", auth.RequireAuth(s.editHandler))

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}

func (s *Server) getAuthLoginFunction(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", chi.URLParam(r, "provider"))
	r.URL.RawQuery = q.Encode()

	// // make provider available to the handler
	// provider := chi.URLParam(r, "provider")
	// r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	gothic.BeginAuthHandler(w, r)
}

func (s *Server) getAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	// make provider available to the handler
	providerValue := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), providerKey, providerValue))

	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, "Error completing user auth\n", r)
		return
	}

	fmt.Println(user.Name)

	err = auth.StoreUserSession(w, r, user)
	if err != nil {
		log.Println(err)
		return
	}

	//s.db.SaveUser(user)

	http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
}

func (s *Server) getAuthLogoutFunction(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)

	auth.RemoveUserSession(w, r)

	w.Header().Set("Location", "http://localhost:3000/login")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) getAllHandler(w http.ResponseWriter, r *http.Request) {
	rows, _ := s.db.GetAll()

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) markDoneHandler(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(paramId)

	if err != nil {
		log.Fatalf("Invalid ID. Err: %v", err)
	}

	s.db.MarkDone(int64(id))

	rows, _ := s.db.GetAll()

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) createHandler(w http.ResponseWriter, r *http.Request) {
	var body models.NewTodo
	json.NewDecoder(r.Body).Decode(&body)

	id, err := s.db.Create(body.Title, body.Description)

	fmt.Println(id)

	if err != nil {
		log.Fatalf("error creating new todo. Err: %v", err)
	}

	rows, _ := s.db.GetAll()

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) editHandler(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(paramId)

	if err != nil {
		log.Fatalf("Invalid ID. Err: %v", err)
	}

	var body models.NewTodo
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, err = s.db.Edit(id, body.Title, body.Description)

	if err != nil {
		log.Fatal(err)
	}

	rows, _ := s.db.GetAll()

	jsonResp, err := json.Marshal(rows)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
