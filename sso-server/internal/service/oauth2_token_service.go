package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// OAuth2TokenService handles OAuth2 token generation and validation
type OAuth2TokenService struct {
	tokenRepo  *repository.OAuth2TokenRepository
	jwtService *JWTService

	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewOAuth2TokenService creates a new OAuth2TokenService
func NewOAuth2TokenService(
	tokenRepo *repository.OAuth2TokenRepository,
	jwtService *JWTService,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) *OAuth2TokenService {
	return &OAuth2TokenService{
		tokenRepo:          tokenRepo,
		jwtService:         jwtService,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// GenerateAccessToken creates a new JWT access token for OAuth2
func (s *OAuth2TokenService) GenerateAccessToken(ctx context.Context, clientID string, userID *string, scopes []string) (string, *models.OAuth2AccessToken, error) {
	// Create custom claims for OAuth2
	claims := map[string]interface{}{
		"client_id": clientID,
		"scope":     strings.Join(scopes, " "),
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(s.accessTokenExpiry).Unix(),
	}

	if userID != nil {
		claims["sub"] = *userID
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateCustomToken(claims)
	if err != nil {
		return "", nil, err
	}

	// Hash token for storage
	tokenHash := hashToken(token)

	// Store token metadata
	accessToken := &models.OAuth2AccessToken{
		ID:        uuid.New().String(),
		TokenHash: tokenHash,
		ClientID:  clientID,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(s.accessTokenExpiry),
	}

	if err := s.tokenRepo.CreateAccessToken(ctx, accessToken); err != nil {
		return "", nil, err
	}

	return token, accessToken, nil
}

// GenerateRefreshToken creates a new refresh token
func (s *OAuth2TokenService) GenerateRefreshToken(ctx context.Context, clientID, userID string, scopes []string, accessTokenID string) (string, error) {
	// Generate random refresh token
	token := generateRandomToken(32)

	refreshToken := &models.OAuth2RefreshToken{
		ID:            uuid.New().String(),
		Token:         token,
		AccessTokenID: &accessTokenID,
		ClientID:      clientID,
		UserID:        userID,
		Scopes:        scopes,
		ExpiresAt:     time.Now().Add(s.refreshTokenExpiry),
		Revoked:       false,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", err
	}

	return token, nil
}

// RefreshAccessToken issues a new access token using a refresh token
func (s *OAuth2TokenService) RefreshAccessToken(ctx context.Context, refreshTokenString, clientID string) (*TokenResponse, error) {
	// Validate refresh token
	refreshToken, err := s.tokenRepo.GetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Verify client ID matches
	if refreshToken.ClientID != clientID {
		return nil, errors.New("client ID mismatch")
	}

	// Generate new access token
	accessTokenString, accessToken, err := s.GenerateAccessToken(
		ctx,
		refreshToken.ClientID,
		&refreshToken.UserID,
		refreshToken.Scopes,
	)
	if err != nil {
		return nil, err
	}

	// Optional: Implement refresh token rotation
	// Revoke old refresh token and issue new one
	// if err := s.tokenRepo.RevokeRefreshToken(ctx, refreshTokenString); err != nil {
	// 	return nil, err
	// }

	// Generate new refresh token (rotation)
	newRefreshToken, err := s.GenerateRefreshToken(
		ctx,
		refreshToken.ClientID,
		refreshToken.UserID,
		refreshToken.Scopes,
		accessToken.ID,
	)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token after new one is created
	s.tokenRepo.RevokeRefreshToken(ctx, refreshTokenString)

	return &TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenExpiry.Seconds()),
		RefreshToken: newRefreshToken,
		Scope:        strings.Join(refreshToken.Scopes, " "),
	}, nil
}

// ValidateAccessToken validates an OAuth2 access token
func (s *OAuth2TokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*models.OAuth2AccessToken, error) {
	// Verify JWT signature and expiry
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Hash token and check database
	tokenHash := hashToken(tokenString)
	token, err := s.tokenRepo.GetAccessTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Additional validation can be done with claims
	_ = claims

	return token, nil
}

// RevokeToken revokes an access or refresh token
func (s *OAuth2TokenService) RevokeToken(ctx context.Context, token string, tokenTypeHint string) error {
	if tokenTypeHint == "refresh_token" {
		return s.tokenRepo.RevokeRefreshToken(ctx, token)
	}

	// Assume access token - hash it first
	tokenHash := hashToken(token)
	accessToken, err := s.tokenRepo.GetAccessTokenByHash(ctx, tokenHash)
	if err != nil {
		return err
	}

	return s.tokenRepo.RevokeAccessToken(ctx, accessToken.ID)
}

// RevokeAllUserTokens revokes all tokens for a user (e.g., on logout or password change)
func (s *OAuth2TokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.tokenRepo.RevokeAllUserTokens(ctx, userID)
}

// CleanupExpiredTokens removes expired tokens from database
func (s *OAuth2TokenService) CleanupExpiredTokens(ctx context.Context) error {
	if err := s.tokenRepo.DeleteExpiredAccessTokens(ctx); err != nil {
		return err
	}
	return s.tokenRepo.DeleteExpiredRefreshTokens(ctx)
}

// Helper functions

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
