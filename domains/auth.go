package domains

import (
	"context"
	"fmt"
	"github.com/go-session/session/v3"
	"github.com/google/uuid"
	"log"
	"net/http"
	"tantieme/helpers"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	db := helpers.GetConnectionManager().GetConnection("sqlite3", "auth")

	result, err := db.Query("SELECT username, password FROM users WHERE username = ? AND password = ?", username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.Next() {
		http.Redirect(w, r, "/#unauthorized", http.StatusFound)
		return
	}

	cookie := uuid.New()
	store.Set(cookie.String(), username)

	err = store.Save()
	if err != nil {
		log.Fatal("Could not save session to redis")
	}

	http.SetCookie(w, &http.Cookie{Name: "tantieme-session", Value: cookie.String(), MaxAge: 86_400, HttpOnly: true})
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	cookie := r.Header.Get("Cookie")
	store.Delete(cookie)
	err = store.Save()
	if err != nil {
		log.Fatal("Could not save session to redis")
	}

	http.SetCookie(w, &http.Cookie{Name: "tantieme-session", Value: "", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/", http.StatusFound)
}

func GetAuthenticatedUsername(w http.ResponseWriter, r *http.Request) (string, bool) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	cookie, err := r.Cookie("tantieme-session")
	if err != nil {
		return "", false
	}
	username, ok := store.Get(cookie.Value)

	fmt.Println(cookie, username, ok)

	return fmt.Sprintf("%s", username), ok
}
