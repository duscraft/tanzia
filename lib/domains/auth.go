package domains

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/duscraft/tanzia/lib/helpers"

	"github.com/go-session/session/v3"
	"github.com/google/uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userID string
	result, err := db.Query("SELECT id FROM users WHERE email = $1 AND password = $2", email, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = result.Close() }()

	if !result.Next() {
		http.Redirect(w, r, "/login#unauthorized", http.StatusFound)
		return
	}

	if err := result.Scan(&userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := uuid.New()
	store.Set(cookie.String(), userID)

	if err := store.Save(); err != nil {
		log.Printf("Session save error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tanzia-session",
		Value:    cookie.String(),
		MaxAge:   86_400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	cookie, err := r.Cookie("tanzia-session")
	if err == nil {
		store.Delete(cookie.Value)
	}

	if err := store.Save(); err != nil {
		log.Printf("Session save error: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tanzia-session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func GetAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Printf("Session error: %v", err)
		return "", false
	}

	cookie, err := r.Cookie("tanzia-session")
	if err != nil {
		return "", false
	}
	id, ok := store.Get(cookie.Value)

	return fmt.Sprintf("%s", id), ok
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	_, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
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

	_, err = db.Exec("INSERT INTO users (email, name, password, is_premium) VALUES ($1, $2, $3, $4)", email, name, password, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
