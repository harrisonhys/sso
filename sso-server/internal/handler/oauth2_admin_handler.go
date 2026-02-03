package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/service"
)

// OAuth2AdminHandler handles OAuth2 client administration endpoints
type OAuth2AdminHandler struct {
	clientService  *service.OAuth2ClientService
	consentService *service.OAuth2ConsentService
}

// NewOAuth2AdminHandler creates a new OAuth2AdminHandler
func NewOAuth2AdminHandler(
	clientService *service.OAuth2ClientService,
	consentService *service.OAuth2ConsentService,
) *OAuth2AdminHandler {
	return &OAuth2AdminHandler{
		clientService:  clientService,
		consentService: consentService,
	}
}

// RegisterClient handles POST /admin/oauth2/clients
func (h *OAuth2AdminHandler) RegisterClient(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}
	userIDStr := userID.(string)

	var req service.RegisterClientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Set owner
	req.OwnerUserID = &userIDStr

	// Register client
	client, clientSecret, err := h.clientService.RegisterClient(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"client":        client,
		"client_secret": clientSecret, // ONLY returned on creation
		"message":       "Client created successfully. Save the client_secret - it will not be shown again!",
	})
}

// GetClients handles GET /admin/api/oauth2-clients
func (h *OAuth2AdminHandler) GetClients(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	clients, total, err := h.clientService.ListClients(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch clients",
		})
	}

	return c.JSON(fiber.Map{
		"clients": clients,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetClient handles GET /admin/oauth2/clients/:client_id
func (h *OAuth2AdminHandler) GetClient(c *fiber.Ctx) error {
	clientID := c.Params("client_id")

	client, err := h.clientService.GetClient(c.Context(), clientID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "client not found",
		})
	}

	return c.JSON(client)
}

// RegenerateSecret handles POST /admin/oauth2/clients/:client_id/regenerate-secret
func (h *OAuth2AdminHandler) RegenerateSecret(c *fiber.Ctx) error {
	clientID := c.Params("client_id")

	newSecret, err := h.clientService.RegenerateClientSecret(c.Context(), clientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"client_secret": newSecret,
		"message":       "Client secret regenerated. Save it now - it will not be shown again!",
	})
}

// RevokeClient handles DELETE /admin/oauth2/clients/:client_id
func (h *OAuth2AdminHandler) RevokeClient(c *fiber.Ctx) error {
	clientID := c.Params("client_id")

	if err := h.clientService.RevokeClient(c.Context(), clientID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Client revoked successfully",
	})
}

// GetUserConsents handles GET /user/oauth2/consents
func (h *OAuth2AdminHandler) GetUserConsents(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}
	userIDStr := userID.(string)

	consents, err := h.consentService.GetUserConsents(c.Context(), userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve consents",
		})
	}

	return c.JSON(consents)
}

// RevokeConsent handles DELETE /user/oauth2/consents/:client_id
func (h *OAuth2AdminHandler) RevokeConsent(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}
	userIDStr := userID.(string)
	clientID := c.Params("client_id")

	if err := h.consentService.RevokeConsent(c.Context(), userIDStr, clientID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Consent revoked successfully",
	})
}
