package domains

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	http.Redirect(w, r, "/login", http.StatusFound)
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

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	_, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal("Could not connect to redis")
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	password := r.FormValue("password")

	if email == "" || name == "" {
		http.Error(w, "Email and name cannot be empty", http.StatusBadRequest)
		return
	}

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (email, name, password) VALUES (?, ?, ?)", email, name, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func isValidEmail(email string) bool {
	// Simple regex for email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
