package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sso-project/sso-server/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTService handles JWT token operations
type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// UserClaims represents JWT claims for a user
type UserClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(issuer string, accessTTL, refreshTTL time.Duration) (*JWTService, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &JWTService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

// GenerateAccessToken creates a new access token
func (s *JWTService) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()

	// Extract roles and permissions
	roles := make([]string, len(user.Roles))
	permissions := make([]string, 0)

	for i, role := range user.Roles {
		roles[i] = role.Name
		for _, perm := range role.Permissions {
			permissions = append(permissions, perm.Name)
		}
	}

	claims := UserClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GenerateRefreshToken creates a new refresh token (simpler claims)
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateCustomToken generates a JWT token with custom claims (for OAuth2)
func (s *JWTService) GenerateCustomToken(customClaims map[string]interface{}) (string, error) {
	// Create token with custom claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(customClaims))

	// Sign token with private key
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ValidateRefreshToken validates a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", ErrInvalidToken
}

// ExportPublicKey exports the public key in PEM format
func (s *JWTService) ExportPublicKey() (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", err
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), nil
}

// ExportPrivateKey exports the private key in PEM format (for backup/recovery)
func (s *JWTService) ExportPrivateKey() (string, error) {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(s.privateKey)

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return string(privKeyPEM), nil
}
