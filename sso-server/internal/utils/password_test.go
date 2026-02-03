package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestComparePassword(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	// Test correct password
	err := ComparePassword(hash, password)
	if err != nil {
		t.Errorf("Expected password to match, got error: %v", err)
	}

	// Test incorrect password
	err = ComparePassword(hash, "WrongPassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}

func TestValidatePassword(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "Valid password",
			password: "ValidPass123!",
			wantErr:  nil,
		},
		{
			name:     "Too short",
			password: "Short1!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "No uppercase",
			password: "lowercase123!",
			wantErr:  ErrPasswordNoUppercase,
		},
		{
			name:     "No lowercase",
			password: "UPPERCASE123!",
			wantErr:  ErrPasswordNoLowercase,
		},
		{
			name:     "No number",
			password: "NoNumbers!",
			wantErr:  ErrPasswordNoNumber,
		},
		{
			name:     "No special",
			password: "NoSpecial123",
			wantErr:  ErrPasswordNoSpecial,
		},
		{
			name:     "Common password",
			password: "Iloveyou1!",
			wantErr:  ErrPasswordCommon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, policy)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateRandomToken(t *testing.T) {
	length := 32

	token1, err := GenerateRandomToken(length)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if len(token1) != length {
		t.Errorf("Expected token length %d, got %d", length, len(token1))
	}

	// Generate another token to ensure randomness
	token2, _ := GenerateRandomToken(length)
	if token1 == token2 {
		t.Error("Tokens should be different")
	}
}
