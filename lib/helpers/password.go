package helpers

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcrypt cost factor - 12 is a good balance between security and performance
	bcryptCost = 12
	// Minimum password length
	minPasswordLength = 8
)

var (
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit     = errors.New("password must contain at least one digit")
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
// Returns true if they match, false otherwise
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks if a password meets strength requirements
// Returns nil if valid, or an error describing the first failed requirement
func ValidatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return ErrPasswordNoLowercase
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}

	return nil
}

// IsLegacyPassword checks if a stored password is a legacy plaintext password
// by checking if it's NOT a valid bcrypt hash
func IsLegacyPassword(storedPassword string) bool {
	// bcrypt hashes always start with "$2a$", "$2b$", or "$2y$" and are 60 chars
	if len(storedPassword) != 60 {
		return true
	}
	if len(storedPassword) >= 4 {
		prefix := storedPassword[:4]
		if prefix == "$2a$" || prefix == "$2b$" || prefix == "$2y$" {
			return false
		}
	}
	return true
}
