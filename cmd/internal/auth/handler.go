package auth

import (
	"errors"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/user"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	log     *logger.Logger
	service *Service
}

func NewHandler(log *logger.Logger, service *Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

type GoogleLoginRequest struct {
	Credential string `json:"credential"`
}

func (h *Handler) GoogleLogin(c fiber.Ctx) error {
	var req GoogleLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Credential == "" {
		return fiber.NewError(fiber.StatusBadRequest, "credential is required")
	}

	result, err := h.service.LoginWithGoogle(c.Context(), req.Credential)
	if err != nil {
		// Erros de negócio
		if errors.Is(err, user.ErrNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		// Erro inesperado
		h.log.Error("google login failed", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}

	return c.JSON(result)
}
