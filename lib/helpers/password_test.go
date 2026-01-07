package helpers

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPass123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if len(hash) != 60 {
		t.Errorf("Expected hash length 60, got %d", len(hash))
	}

	if hash[:4] != "$2a$" {
		t.Errorf("Expected bcrypt prefix $2a$, got %s", hash[:4])
	}
}

func TestCheckPassword(t *testing.T) {
	password := "TestPass123"
	hash, _ := HashPassword(password)

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	if CheckPassword("WrongPassword1", hash) {
		t.Error("CheckPassword should return false for incorrect password")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		wantErr  error
	}{
		{"short", ErrPasswordTooShort},
		{"alllowercase1", ErrPasswordNoUppercase},
		{"ALLUPPERCASE1", ErrPasswordNoLowercase},
		{"NoDigitsHere", ErrPasswordNoDigit},
		{"ValidPass1", nil},
		{"Another$Valid2", nil},
	}

	for _, tt := range tests {
		err := ValidatePasswordStrength(tt.password)
		if err != tt.wantErr {
			t.Errorf("ValidatePasswordStrength(%q) = %v, want %v", tt.password, err, tt.wantErr)
		}
	}
}

func TestIsLegacyPassword(t *testing.T) {
	tests := []struct {
		password string
		isLegacy bool
	}{
		{"plaintext", true},
		{"short", true},
		{"$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.2OoH7x4n/SRmOy", false},
		{"$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.2OoH7x4n/SRmOy", false},
		{"$2y$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.2OoH7x4n/SRmOy", false},
		{"$1a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.2OoH7x4n/SRmOy", true},
	}

	for _, tt := range tests {
		got := IsLegacyPassword(tt.password)
		if got != tt.isLegacy {
			t.Errorf("IsLegacyPassword(%q) = %v, want %v", tt.password, got, tt.isLegacy)
		}
	}
}
