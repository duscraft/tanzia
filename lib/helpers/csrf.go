package helpers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	csrfTokenLength   = 32
	CSRFCookieName    = "tanzia-csrf"
	CSRFFormFieldName = "csrf_token"
	csrfTokenExpiry   = 24 * time.Hour
)

type csrfToken struct {
	token     string
	createdAt time.Time
}

type CSRFManager struct {
	tokens map[string]*csrfToken
	mu     sync.RWMutex
}

var csrfManager *CSRFManager

func GetCSRFManager() *CSRFManager {
	if csrfManager == nil {
		csrfManager = &CSRFManager{
			tokens: make(map[string]*csrfToken),
		}
		go csrfManager.cleanup()
	}
	return csrfManager
}

func NewCSRFManager() *CSRFManager {
	return &CSRFManager{
		tokens: make(map[string]*csrfToken),
	}
}

func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (cm *CSRFManager) CreateToken(sessionID string) (string, error) {
	token, err := GenerateCSRFToken()
	if err != nil {
		return "", err
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.tokens[sessionID] = &csrfToken{
		token:     token,
		createdAt: time.Now(),
	}

	return token, nil
}

func (cm *CSRFManager) ValidateToken(sessionID, token string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stored, exists := cm.tokens[sessionID]
	if !exists {
		return false
	}

	if time.Since(stored.createdAt) > csrfTokenExpiry {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(stored.token), []byte(token)) == 1
}

func (cm *CSRFManager) InvalidateToken(sessionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.tokens, sessionID)
}

func (cm *CSRFManager) GetToken(sessionID string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stored, exists := cm.tokens[sessionID]
	if !exists {
		return ""
	}

	if time.Since(stored.createdAt) > csrfTokenExpiry {
		return ""
	}

	return stored.token
}

func SetCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		MaxAge:   int(csrfTokenExpiry.Seconds()),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

func GetCSRFTokenFromRequest(r *http.Request) string {
	token := r.FormValue(CSRFFormFieldName)
	if token != "" {
		return token
	}

	token = r.Header.Get("X-CSRF-Token")
	if token != "" {
		return token
	}

	cookie, err := r.Cookie(CSRFCookieName)
	if err == nil {
		return cookie.Value
	}

	return ""
}

func GetSessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("tanzia-session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (cm *CSRFManager) cleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for key, token := range cm.tokens {
			if now.Sub(token.createdAt) > csrfTokenExpiry {
				delete(cm.tokens, key)
			}
		}
		cm.mu.Unlock()
	}
}
