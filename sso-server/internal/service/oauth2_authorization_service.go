package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/utils"
)

var (
	ErrInvalidPKCE         = errors.New("invalid PKCE challenge/verifier")
	ErrAuthCodeExpired     = errors.New("authorization code has expired")
	ErrAuthCodeUsed        = errors.New("authorization code has already been used")
	ErrMismatchRedirectURI = errors.New("redirect URI does not match")
)

// OAuth2AuthorizationService handles OAuth2 authorization flow
type OAuth2AuthorizationService struct {
	codeRepo     *repository.OAuth2CodeRepository
	clientRepo   *repository.OAuth2ClientRepository
	consentRepo  *repository.OAuth2ConsentRepository
	tokenService *OAuth2TokenService
	codeExpiry   time.Duration
	enforcePKCE  bool
}

// NewOAuth2AuthorizationService creates a new OAuth2AuthorizationService
func NewOAuth2AuthorizationService(
	codeRepo *repository.OAuth2CodeRepository,
	clientRepo *repository.OAuth2ClientRepository,
	consentRepo *repository.OAuth2ConsentRepository,
	tokenService *OAuth2TokenService,
	codeExpiry time.Duration,
	enforcePKCE bool,
) *OAuth2AuthorizationService {
	return &OAuth2AuthorizationService{
		codeRepo:     codeRepo,
		clientRepo:   clientRepo,
		consentRepo:  consentRepo,
		tokenService: tokenService,
		codeExpiry:   codeExpiry,
		enforcePKCE:  enforcePKCE,
	}
}

// CreateAuthorizationCode generates a new authorization code
func (s *OAuth2AuthorizationService) CreateAuthorizationCode(
	ctx context.Context,
	clientID, userID, redirectURI string,
	scopes []string,
	codeChallenge, codeChallengeMethod *string,
) (string, error) {
	// Validate client
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return "", err
	}

	if !client.IsActive {
		return "", errors.New("client is not active")
	}

	// Validate redirect URI
	valid, err := s.clientRepo.ValidateRedirectURI(ctx, clientID, redirectURI)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", errors.New("invalid redirect URI")
	}

	// Validate scopes
	if len(scopes) > 0 {
		valid, err := s.clientRepo.ValidateScopes(ctx, clientID, scopes)
		if err != nil {
			return "", err
		}
		if !valid {
			return "", errors.New("invalid scopes for this client")
		}
	}

	// Enforce PKCE for public clients
	if client.IsPublic && codeChallenge == nil {
		if s.enforcePKCE {
			return "", errors.New("PKCE required for public clients")
		}
	}

	// Generate authorization code
	code := generateAuthCode()

	authCode := &models.OAuth2AuthorizationCode{
		ID:                  uuid.New().String(),
		Code:                code,
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		Scopes:              scopes,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(s.codeExpiry),
		Used:                false,
	}

	if err := s.codeRepo.Create(ctx, authCode); err != nil {
		return "", err
	}

	return code, nil
}

// ExchangeCodeForTokens exchanges an authorization code for access/refresh tokens
func (s *OAuth2AuthorizationService) ExchangeCodeForTokens(
	ctx context.Context,
	code, clientID, redirectURI string,
	codeVerifier *string,
) (*TokenResponse, error) {
	// Validate code
	authCode, err := s.codeRepo.ValidateCode(ctx, code, clientID, redirectURI)
	if err != nil {
		return nil, err
	}

	// Verify PKCE if challenge was provided
	if authCode.CodeChallenge != nil {
		if codeVerifier == nil {
			return nil, errors.New("code verifier required")
		}

		method := "S256"
		if authCode.CodeChallengeMethod != nil {
			method = *authCode.CodeChallengeMethod
		}

		valid, err := utils.VerifyCodeChallenge(*codeVerifier, *authCode.CodeChallenge, method)
		if err != nil || !valid {
			return nil, ErrInvalidPKCE
		}
	}

	// Mark code as used (prevent replay attacks)
	if err := s.codeRepo.MarkAsUsed(ctx, code); err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, tokenModel, err := s.tokenService.GenerateAccessToken(
		ctx,
		clientID,
		&authCode.UserID,
		authCode.Scopes,
	)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(
		ctx,
		clientID,
		authCode.UserID,
		authCode.Scopes,
		tokenModel.ID,
	)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenService.accessTokenExpiry.Seconds()),
		RefreshToken: refreshToken,
		Scope:        scopesToString(authCode.Scopes),
	}, nil
}

// ClientCredentialsGrant handles the client credentials grant flow
func (s *OAuth2AuthorizationService) ClientCredentialsGrant(
	ctx context.Context,
	clientID string,
	scopes []string,
) (*TokenResponse, error) {
	// Validate client
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if !client.IsActive {
		return nil, errors.New("client is not active")
	}

	// Client credentials flow doesn't have a user
	// Validate scopes
	if len(scopes) > 0 {
		valid, err := s.clientRepo.ValidateScopes(ctx, clientID, scopes)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, errors.New("invalid scopes for this client")
		}
	}

	// Generate access token (no user ID, no refresh token)
	accessToken, _, err := s.tokenService.GenerateAccessToken(
		ctx,
		clientID,
		nil, // No user for client credentials
		scopes,
	)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenService.accessTokenExpiry.Seconds()),
		Scope:       scopesToString(scopes),
	}, nil
}

// Helper functions

func generateAuthCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func scopesToString(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}
