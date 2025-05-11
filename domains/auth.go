package domains

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"tantieme/helpers"

	"github.com/go-session/session/v3"
	"github.com/google/uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userId string
	result, err := db.Query("SELECT id FROM users WHERE username = ? AND password = ?", username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !result.Next() {
		http.Redirect(w, r, "/#unauthorized", http.StatusFound)
		return
	}

	err = result.Scan(&userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := uuid.New()
	store.Set(cookie.String(), userId)

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

func GetAuthenticatedUserId(w http.ResponseWriter, r *http.Request) (string, bool) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	cookie, err := r.Cookie("tantieme-session")
	if err != nil {
		return "", false
	}
	id, ok := store.Get(cookie.Value)

	return fmt.Sprintf("%s", id), ok
}
