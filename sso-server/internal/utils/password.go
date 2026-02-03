package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

var (
	ErrPasswordTooShort    = errors.New("password is too short")
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber    = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial   = errors.New("password must contain at least one special character")
	ErrPasswordCommon      = errors.New("password is too common")
	ErrPasswordReused      = errors.New("password was recently used")
	ErrPasswordInvalid     = errors.New("invalid password")
)

// PasswordPolicy represents password complexity requirements
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword compares a password with its hash
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePassword checks if password meets policy requirements
func ValidatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return ErrPasswordTooShort
	}

	if policy.RequireUpper {
		hasUpper := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return ErrPasswordNoUppercase
		}
	}

	if policy.RequireLower {
		hasLower := false
		for _, char := range password {
			if unicode.IsLower(char) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return ErrPasswordNoLowercase
		}
	}

	if policy.RequireNumber {
		hasNumber := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return ErrPasswordNoNumber
		}
	}

	if policy.RequireSpecial {
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
		if !hasSpecial {
			return ErrPasswordNoSpecial
		}
	}

	// Check against common passwords
	if isCommonPassword(password) {
		return ErrPasswordCommon
	}

	return nil
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// isCommonPassword checks if password is in common passwords list
func isCommonPassword(password string) bool {
	// Simple check - in production, use a comprehensive list
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "iloveyou", "master", "sunshine", "ashley",
		"bailey", "passw0rd", "shadow", "123123", "654321",
		"superman", "qazwsx", "michael", "football",
	}

	// Convert to lowercase and check exact match first
	passwordLower := strings.ToLower(password)
	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	// Check if password contains common words
	for _, common := range commonPasswords {
		if strings.Contains(passwordLower, common) && len(common) > 4 {
			return true
		}
	}

	return false
}
