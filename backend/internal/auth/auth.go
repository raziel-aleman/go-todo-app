package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

const (
	key         = "super-secret-key"
	MaxAge      = 86400 * 30
	IsProd      = false
	SessionName = "user_session"
)

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
	store.Options.HttpOnly = false
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteNoneMode

	gothic.Store = store

	goth.UseProviders(
		github.New(githubClientId, githubClientSecret, githubCallbackUrl),
	)
}

func StoreUserSession(w http.ResponseWriter, r *http.Request, user goth.User) error {
	//session, _ := gothic.Store.Get(r, SessionName)

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = false
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteNoneMode

	session, _ := store.Get(r, SessionName)

	session.Values["user"] = user

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func GetUserSession(r *http.Request) (goth.User, error) {
	session, err := gothic.Store.Get(r, SessionName)
	if err != nil {
		return goth.User{}, err
	}

	u := session.Values["user"]

	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}

	return u.(goth.User), nil
}

func RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetUserSession(r)
		if err != nil {
			//log.Println("User is not authenticated*!")
			http.Redirect(w, r, "http://localhost:3000/login", http.StatusTemporaryRedirect)
			return
		}

		log.Println(session.Name)

		handlerFunc(w, r)
	}
}

func RemoveUserSession(w http.ResponseWriter, r *http.Request) error {
	session, _ := gothic.Store.Get(r, SessionName)

	session.Options.MaxAge = -1
	//session.Options.Value = ""
	session.Options.Path = "/"
	session.Options.HttpOnly = true

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
