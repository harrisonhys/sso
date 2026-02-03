package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
)

// ConfigHandler handles system configuration endpoints
type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// GetAllConfigs returns all system configurations
func (h *ConfigHandler) GetAllConfigs(c *fiber.Ctx) error {
	configs, err := h.configService.GetAllConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch configurations",
		})
	}

	return c.JSON(configs)
}

// GetConfig returns a specific configuration
func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	key := c.Params("key")

	config, err := h.configService.GetConfig(c.Context(), key)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "configuration not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch configuration",
		})
	}

	return c.JSON(config)
}

// UpdateConfig updates a configuration
func (h *ConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	var req struct {
		Value string `json:"value"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	config, err := h.configService.UpdateConfig(c.Context(), key, req.Value)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update configuration",
		})
	}

	return c.JSON(config)
}
