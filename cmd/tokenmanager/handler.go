package tokenmanager

import (
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	log     *logger.Logger
	manager *OAuth2Manager
}

func NewHandler(log *logger.Logger, manager *OAuth2Manager) *Handler {
	return &Handler{
		log:     log,
		manager: manager,
	}
}

func (h *Handler) Authorize(c fiber.Ctx) {
	code := c.Query("code")
	if code == "" {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not find code in request",
		})
	}
	state := c.Query("state")
	if err := h.manager.ValidateState(state); state != "" && err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "invalid state provided",
			"error":   err,
		})
		return
	}

	// Request tokens
	tokens, err := h.manager.RequestTokensWithCode(c.Context(), code)
	if err != nil {
		h.log.Error(err)
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not get tokens",
			"error":   err,
		})
	}

	now := time.Now()
	if err := h.manager.UpdateTokens(
		c.Context(),
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.ExpiresIn,
		now,
	); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get tokens",
			"error":   err,
		})
	}

	c.SendStatus(202)
}

func (h *Handler) validateState(state string) error {
	return nil
}
