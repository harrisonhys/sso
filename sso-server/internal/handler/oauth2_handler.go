package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
)

// OAuth2Handler handles OAuth2 endpoints
type OAuth2Handler struct {
	authzService   *service.OAuth2AuthorizationService
	tokenService   *service.OAuth2TokenService
	clientService  *service.OAuth2ClientService
	consentService *service.OAuth2ConsentService
	userRepo       *repository.UserRepository
}

// NewOAuth2Handler creates a new OAuth2Handler
func NewOAuth2Handler(
	authzService *service.OAuth2AuthorizationService,
	tokenService *service.OAuth2TokenService,
	clientService *service.OAuth2ClientService,
	consentService *service.OAuth2ConsentService,
	userRepo *repository.UserRepository,
) *OAuth2Handler {
	return &OAuth2Handler{
		authzService:   authzService,
		tokenService:   tokenService,
		clientService:  clientService,
		consentService: consentService,
		userRepo:       userRepo,
	}
}

// Authorize handles GET /oauth2/authorize
func (h *OAuth2Handler) Authorize(c *fiber.Ctx) error {
	// Parse query parameters
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	// Validate required parameters
	if responseType != "code" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "unsupported_response_type",
			"error_description": "Only authorization code flow is supported",
		})
	}

	if clientID == "" || redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	// Validate client and redirect URI
	valid, err := h.clientService.ValidateRedirectURI(c.Context(), clientID, redirectURI)
	if err != nil || !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_client",
			"error_description": "Invalid client_id or redirect_uri",
		})
	}

	// Check if user is authenticated
	userID := c.Locals("user_id")
	if userID == nil {
		// Redirect to login with return URL
		returnURL := c.OriginalURL()
		return c.Redirect("/login.html?return_url=" + returnURL)
	}

	userIDStr := userID.(string)

	// Parse scopes
	scopes := parseScopes(scope)

	// Check if user has already consented
	hasConsent, missingScopes, err := h.consentService.CheckConsent(c.Context(), userIDStr, clientID, scopes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	// If consent is missing, show consent screen
	if !hasConsent || len(missingScopes) > 0 {
		// Render consent screen
		return c.Render("oauth2/consent", fiber.Map{
			"client_id":             clientID,
			"redirect_uri":          redirectURI,
			"scope":                 scope,
			"state":                 state,
			"code_challenge":        codeChallenge,
			"code_challenge_method": codeChallengeMethod,
			"requested_scopes":      scopes,
		})
	}

	// User has consented, generate authorization code
	var challengePtr, methodPtr *string
	if codeChallenge != "" {
		challengePtr = &codeChallenge
		if codeChallengeMethod == "" {
			method := "S256"
			methodPtr = &method
		} else {
			methodPtr = &codeChallengeMethod
		}
	}

	code, err := h.authzService.CreateAuthorizationCode(
		c.Context(),
		clientID,
		userIDStr,
		redirectURI,
		scopes,
		challengePtr,
		methodPtr,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	// Redirect back to client with code
	redirectURL := redirectURI + "?code=" + code
	if state != "" {
		redirectURL += "&state=" + state
	}

	return c.Redirect(redirectURL)
}

// AuthorizeConsent handles POST /oauth2/authorize/consent
func (h *OAuth2Handler) AuthorizeConsent(c *fiber.Ctx) error {
	// Get user ID from auth middleware
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}
	userIDStr := userID.(string)

	// Parse form data
	type ConsentRequest struct {
		ClientID            string `form:"client_id"`
		RedirectURI         string `form:"redirect_uri"`
		Scope               string `form:"scope"`
		State               string `form:"state"`
		CodeChallenge       string `form:"code_challenge"`
		CodeChallengeMethod string `form:"code_challenge_method"`
		Approve             string `form:"approve"`
	}

	var req ConsentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	// If user denied consent
	if req.Approve != "true" {
		redirectURL := req.RedirectURI + "?error=access_denied"
		if req.State != "" {
			redirectURL += "&state=" + req.State
		}
		return c.Redirect(redirectURL)
	}

	// Parse scopes
	scopes := parseScopes(req.Scope)

	// Save consent
	if err := h.consentService.GrantConsent(c.Context(), userIDStr, req.ClientID, scopes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	// Generate authorization code
	var challengePtr, methodPtr *string
	if req.CodeChallenge != "" {
		challengePtr = &req.CodeChallenge
		if req.CodeChallengeMethod == "" {
			method := "S256"
			methodPtr = &method
		} else {
			methodPtr = &req.CodeChallengeMethod
		}
	}

	code, err := h.authzService.CreateAuthorizationCode(
		c.Context(),
		req.ClientID,
		userIDStr,
		req.RedirectURI,
		scopes,
		challengePtr,
		methodPtr,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	// Redirect back to client with code
	redirectURL := req.RedirectURI + "?code=" + code
	if req.State != "" {
		redirectURL += "&state=" + req.State
	}

	return c.Redirect(redirectURL)
}

// Token handles POST /oauth2/token
func (h *OAuth2Handler) Token(c *fiber.Ctx) error {
	// Parse form data
	grantType := c.FormValue("grant_type")
	clientID := c.FormValue("client_id")
	clientSecret := c.FormValue("client_secret")

	// Validate client credentials
	_, err := h.clientService.ValidateClient(c.Context(), clientID, clientSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "invalid_client",
			"error_description": "Invalid client credentials",
		})
	}

	switch grantType {
	case "authorization_code":
		return h.handleAuthorizationCodeGrant(c, clientID)
	case "refresh_token":
		return h.handleRefreshTokenGrant(c, clientID)
	case "client_credentials":
		return h.handleClientCredentialsGrant(c, clientID)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "unsupported_grant_type",
			"error_description": "Grant type not supported",
		})
	}
}

func (h *OAuth2Handler) handleAuthorizationCodeGrant(c *fiber.Ctx, clientID string) error {
	code := c.FormValue("code")
	redirectURI := c.FormValue("redirect_uri")
	codeVerifier := c.FormValue("code_verifier")

	if code == "" || redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	var verifierPtr *string
	if codeVerifier != "" {
		verifierPtr = &codeVerifier
	}

	tokenResp, err := h.authzService.ExchangeCodeForTokens(
		c.Context(),
		code,
		clientID,
		redirectURI,
		verifierPtr,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
	}

	return c.JSON(tokenResp)
}

func (h *OAuth2Handler) handleRefreshTokenGrant(c *fiber.Ctx, clientID string) error {
	refreshToken := c.FormValue("refresh_token")

	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	tokenResp, err := h.tokenService.RefreshAccessToken(c.Context(), refreshToken, clientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
	}

	return c.JSON(tokenResp)
}

func (h *OAuth2Handler) handleClientCredentialsGrant(c *fiber.Ctx, clientID string) error {
	scope := c.FormValue("scope")
	scopes := parseScopes(scope)

	tokenResp, err := h.authzService.ClientCredentialsGrant(c.Context(), clientID, scopes)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_scope",
			"error_description": err.Error(),
		})
	}

	return c.JSON(tokenResp)
}

// Revoke handles POST /oauth2/revoke
func (h *OAuth2Handler) Revoke(c *fiber.Ctx) error {
	token := c.FormValue("token")
	tokenTypeHint := c.FormValue("token_type_hint")

	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	if err := h.tokenService.RevokeToken(c.Context(), token, tokenTypeHint); err != nil {
		// Don't reveal if token existed or not
		return c.SendStatus(fiber.StatusOK)
	}

	return c.SendStatus(fiber.StatusOK)
}

// UserInfo handles GET /oauth2/userinfo
func (h *OAuth2Handler) UserInfo(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "invalid_token",
			"error_description": "Missing or invalid Authorization header",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	accessToken, err := h.tokenService.ValidateAccessToken(c.Context(), tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "invalid_token",
			"error_description": "Token is invalid, expired, or revoked",
		})
	}

	// Get user details
	if accessToken.UserID == nil {
		// Client credentials flow doesn't have user
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "Token does not belong to a user",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), *accessToken.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	// OpenID Connect Standard Claims
	userInfo := fiber.Map{
		"sub":            user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"updated_at":     user.UpdatedAt.Unix(),
	}

	return c.JSON(userInfo)
}

// Helper functions

func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	return strings.Fields(scopeStr)
}
