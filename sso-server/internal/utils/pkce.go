package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// PKCE (Proof Key for Code Exchange) utilities for OAuth2

// GenerateCodeVerifier generates a random code verifier for PKCE
// Length should be between 43-128 characters (256-512 bits of entropy)
func GenerateCodeVerifier() (string, error) {
	// Generate 32 random bytes (256 bits)
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode as base64url (RFC 7636)
	verifier := base64.RawURLEncoding.EncodeToString(b)
	return verifier, nil
}

// GenerateCodeChallenge generates a code challenge from a verifier using S256 method
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge
}

// VerifyCodeChallenge verifies that a code verifier matches the challenge
func VerifyCodeChallenge(verifier, challenge, method string) (bool, error) {
	if method == "" || method == "plain" {
		// Plain method: verifier must equal challenge
		return verifier == challenge, nil
	}

	if method == "S256" {
		// S256 method: hash verifier and compare
		computed := GenerateCodeChallenge(verifier)
		return computed == challenge, nil
	}

	return false, errors.New("unsupported code challenge method: " + method)
}
