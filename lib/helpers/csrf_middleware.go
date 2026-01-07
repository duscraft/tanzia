package helpers

import (
	"log"
	"net/http"
)

func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next(w, r)
			return
		}

		sessionID := GetSessionIDFromRequest(r)
		if sessionID == "" {
			next(w, r)
			return
		}

		token := GetCSRFTokenFromRequest(r)
		if token == "" {
			log.Printf("CSRF validation failed: no token provided for session %s", sessionID)
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		csrfMgr := GetCSRFManager()
		if !csrfMgr.ValidateToken(sessionID, token) {
			log.Printf("CSRF validation failed: invalid token for session %s", sessionID)
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func CSRFProtect(handler http.HandlerFunc) http.HandlerFunc {
	return CSRFMiddleware(handler)
}

func EnsureCSRFToken(w http.ResponseWriter, r *http.Request) string {
	sessionID := GetSessionIDFromRequest(r)
	if sessionID == "" {
		return ""
	}

	csrfMgr := GetCSRFManager()

	existingToken := csrfMgr.GetToken(sessionID)
	if existingToken != "" {
		return existingToken
	}

	token, err := csrfMgr.CreateToken(sessionID)
	if err != nil {
		log.Printf("Failed to create CSRF token: %v", err)
		return ""
	}

	SetCSRFCookie(w, token)
	return token
}

