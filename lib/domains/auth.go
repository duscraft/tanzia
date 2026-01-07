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

	rateLimiter := helpers.GetRateLimiter()
	if rateLimiter.IsLocked(email) {
		remaining := rateLimiter.GetLockoutRemaining(email)
		minutes := int(remaining.Minutes()) + 1
		http.Redirect(w, r, fmt.Sprintf("/login#locked-%d", minutes), http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userID string
	var storedPassword string
	var needsPasswordReset bool
	result, err := db.Query("SELECT id, password, needs_password_reset FROM users WHERE email = $1", email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = result.Close() }()

	if !result.Next() {
		rateLimiter.RecordFailedAttempt(email)
		remaining := rateLimiter.GetRemainingAttempts(email)
		http.Redirect(w, r, fmt.Sprintf("/login#unauthorized-%d", remaining), http.StatusFound)
		return
	}

	if err := result.Scan(&userID, &storedPassword, &needsPasswordReset); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if helpers.IsLegacyPassword(storedPassword) {
		if password != storedPassword {
			locked := rateLimiter.RecordFailedAttempt(email)
			if locked {
				http.Redirect(w, r, "/login#locked-15", http.StatusFound)
				return
			}
			remaining := rateLimiter.GetRemainingAttempts(email)
			http.Redirect(w, r, fmt.Sprintf("/login#unauthorized-%d", remaining), http.StatusFound)
			return
		}
		hashedPassword, err := helpers.HashPassword(password)
		if err != nil {
			log.Printf("Failed to hash password during migration: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", hashedPassword, userID)
		if err != nil {
			log.Printf("Failed to update password hash: %v", err)
		}
	} else {
		if !helpers.CheckPassword(password, storedPassword) {
			locked := rateLimiter.RecordFailedAttempt(email)
			if locked {
				http.Redirect(w, r, "/login#locked-15", http.StatusFound)
				return
			}
			remaining := rateLimiter.GetRemainingAttempts(email)
			http.Redirect(w, r, fmt.Sprintf("/login#unauthorized-%d", remaining), http.StatusFound)
			return
		}
	}

	rateLimiter.ResetAttempts(email)

	if needsPasswordReset {
		http.Redirect(w, r, "/reset-password#required", http.StatusFound)
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
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	password := r.FormValue("password")
	redirect := r.FormValue("redirect")

	if email == "" || name == "" {
		http.Error(w, "Email and name cannot be empty", http.StatusBadRequest)
		return
	}

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if err := helpers.ValidatePasswordStrength(password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userID int
	err = db.QueryRow("INSERT INTO users (email, name, password, is_premium, needs_password_reset) VALUES ($1, $2, $3, $4, $5) RETURNING id", email, name, hashedPassword, false, false).Scan(&userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := uuid.New()
	store.Set(cookie.String(), fmt.Sprintf("%d", userID))

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

	if redirect == "subscribe" {
		http.Redirect(w, r, "/subscribe", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetAuthenticatedUserID(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/reset-password#mismatch", http.StatusFound)
		return
	}

	if err := helpers.ValidatePasswordStrength(newPassword); err != nil {
		http.Redirect(w, r, "/reset-password#weak", http.StatusFound)
		return
	}

	db, err := helpers.GetConnectionManager().GetConnection("postgres")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&storedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if helpers.IsLegacyPassword(storedPassword) {
		if currentPassword != storedPassword {
			http.Redirect(w, r, "/reset-password#invalid", http.StatusFound)
			return
		}
	} else {
		if !helpers.CheckPassword(currentPassword, storedPassword) {
			http.Redirect(w, r, "/reset-password#invalid", http.StatusFound)
			return
		}
	}

	hashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET password = $1, needs_password_reset = FALSE WHERE id = $2", hashedPassword, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/reset-password#success", http.StatusFound)
}
