package auth

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

const (
	//key         =  "super-secret-key"
	MaxAge      = 86400 * 14
	HttpOnly    = true
	IsProd      = true
	SessionName = "user_session"
)

var key = securecookie.GenerateRandomKey(64)

func NewAuth() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env files")
	}

	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubCallbackUrl := os.Getenv("GITHUB_CALLBACK_URL")

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = HttpOnly
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteNoneMode

	gothic.Store = store

	goth.UseProviders(
		github.New(githubClientId, githubClientSecret, githubCallbackUrl),
	)
}

func StoreUserSession(w http.ResponseWriter, r *http.Request, user goth.User) (string, error) {

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = HttpOnly
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteNoneMode

	userSession, err := store.Get(r, SessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	sessionId := uuid.NewString()
	userSession.Values["sessionId"] = sessionId

	err = userSession.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	return sessionId, nil
}

func GetUserSession(r *http.Request) (string, error) {
	store := sessions.NewCookieStore([]byte(key))

	userSession, err := store.Get(r, SessionName)
	if err != nil {
		return "", err
	}

	if sessionId, ok := userSession.Values["sessionId"]; ok {
		return sessionId.(string), nil
	} else {
		return "", errors.New("no session id found in request")
	}

	//return userSession.Values["sessionId"].(string), nil
}

func RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId, err := GetUserSession(r)
		if err != nil {
			log.Println(err, "User is not authenticated!")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		//log.Printf("session id found: %s", sessionId)

		parsedUUID, err := uuid.Parse(sessionId)

		if err != nil {
			log.Println(err, "User is not authenticated!")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("session id parsed as UUID: %s", parsedUUID)

		handlerFunc(w, r)
	}
}

func RemoveUserSession(w http.ResponseWriter, r *http.Request) error {
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(-1)
	store.Options.Path = "/"
	store.Options.HttpOnly = HttpOnly
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteNoneMode

	session, _ := store.Get(r, SessionName)

	err := session.Save(r, w)
	if err != nil {
		log.Println("could not expire client session")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
