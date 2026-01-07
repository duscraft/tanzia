package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateCSRFToken(t *testing.T) {
	token1, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken failed: %v", err)
	}

	if len(token1) == 0 {
		t.Error("Token should not be empty")
	}

	token2, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken failed: %v", err)
	}

	if token1 == token2 {
		t.Error("Tokens should be unique")
	}
}

func TestCSRFManagerCreateAndValidate(t *testing.T) {
	cm := NewCSRFManager()
	sessionID := "test-session-123"

	token, err := cm.CreateToken(sessionID)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if !cm.ValidateToken(sessionID, token) {
		t.Error("ValidateToken should return true for valid token")
	}

	if cm.ValidateToken(sessionID, "invalid-token") {
		t.Error("ValidateToken should return false for invalid token")
	}

	if cm.ValidateToken("other-session", token) {
		t.Error("ValidateToken should return false for wrong session")
	}
}

func TestCSRFManagerGetToken(t *testing.T) {
	cm := NewCSRFManager()
	sessionID := "test-session-456"

	emptyToken := cm.GetToken(sessionID)
	if emptyToken != "" {
		t.Error("GetToken should return empty for non-existent session")
	}

	token, _ := cm.CreateToken(sessionID)

	retrievedToken := cm.GetToken(sessionID)
	if retrievedToken != token {
		t.Error("GetToken should return the created token")
	}
}

func TestCSRFManagerInvalidate(t *testing.T) {
	cm := NewCSRFManager()
	sessionID := "test-session-789"

	token, _ := cm.CreateToken(sessionID)

	if !cm.ValidateToken(sessionID, token) {
		t.Error("Token should be valid before invalidation")
	}

	cm.InvalidateToken(sessionID)

	if cm.ValidateToken(sessionID, token) {
		t.Error("Token should be invalid after invalidation")
	}
}

func TestCSRFManagerTokenExpiry(t *testing.T) {
	cm := NewCSRFManager()
	sessionID := "test-session-expiry"

	token, _ := cm.CreateToken(sessionID)

	cm.mu.Lock()
	cm.tokens[sessionID].createdAt = time.Now().Add(-25 * time.Hour)
	cm.mu.Unlock()

	if cm.ValidateToken(sessionID, token) {
		t.Error("Expired token should not be valid")
	}

	if cm.GetToken(sessionID) != "" {
		t.Error("GetToken should return empty for expired token")
	}
}

func TestGetCSRFTokenFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		setupReq   func(*http.Request)
		wantToken  string
	}{
		{
			name: "from form value",
			setupReq: func(r *http.Request) {
				r.Form = map[string][]string{
					CSRFFormFieldName: {"form-token"},
				}
			},
			wantToken: "form-token",
		},
		{
			name: "from header",
			setupReq: func(r *http.Request) {
				r.Header.Set("X-CSRF-Token", "header-token")
			},
			wantToken: "header-token",
		},
		{
			name: "from cookie",
			setupReq: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  CSRFCookieName,
					Value: "cookie-token",
				})
			},
			wantToken: "cookie-token",
		},
		{
			name:      "no token",
			setupReq:  func(r *http.Request) {},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			tt.setupReq(req)

			got := GetCSRFTokenFromRequest(req)
			if got != tt.wantToken {
				t.Errorf("GetCSRFTokenFromRequest() = %v, want %v", got, tt.wantToken)
			}
		})
	}
}

func TestGetSessionIDFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	sessionID := GetSessionIDFromRequest(req)
	if sessionID != "" {
		t.Error("Should return empty when no cookie")
	}

	req.AddCookie(&http.Cookie{
		Name:  "tanzia-session",
		Value: "my-session-id",
	})

	sessionID = GetSessionIDFromRequest(req)
	if sessionID != "my-session-id" {
		t.Errorf("Expected 'my-session-id', got %q", sessionID)
	}
}

func TestSetCSRFCookie(t *testing.T) {
	w := httptest.NewRecorder()
	token := "test-csrf-token"

	SetCSRFCookie(w, token)

	resp := w.Result()
	cookies := resp.Cookies()

	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("CSRF cookie not set")
	}

	if csrfCookie.Value != token {
		t.Errorf("Cookie value = %q, want %q", csrfCookie.Value, token)
	}

	if csrfCookie.Secure != true {
		t.Error("Cookie should be secure")
	}

	if csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Error("Cookie should have SameSite=Strict")
	}
}
