package user

import (
	"context"
	"errors"
	"sync"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/permissions"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	mu       sync.Mutex
	log      *logger.Logger
	userRepo Repository
	permRepo permissions.Repository
}

func NewHandler(log *logger.Logger, userRepo Repository, permRepo permissions.Repository) *Handler {
	return &Handler{
		mu:       sync.Mutex{},
		log:      log,
		userRepo: userRepo,
		permRepo: permRepo,
	}
}

func (h *Handler) GrantPermissions(c fiber.Ctx) error {
	var req GrantPermissionsRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	response, err := h.grantPermissions(c.Context(), req.Email, req.Permissions)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "error granting permissions")
		}
		return fiber.NewError(fiber.StatusBadRequest, "error granting permissions")
	}

	var status int
	switch {
	case response.Failures == response.Total:
		status = fiber.StatusBadRequest
	case response.Successes == response.Total:
		status = fiber.StatusOK
	default:
		status = fiber.StatusMultiStatus
	}

	return c.Status(status).JSON(response)
}

func (h *Handler) grantPermissions(ctx context.Context, email string, permissions []string) (*GrantPermissionsResponse, error) {
	response := &GrantPermissionsResponse{}
	user, err := h.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrNotFound
	}

	wg := sync.WaitGroup{}
	for _, perm := range permissions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { response.Total++ }()
			permission, err := h.permRepo.GetByName(ctx, perm)
			if err != nil {
				h.log.Error("grant failed", "permission", perm, "err", err)
				h.mu.Lock()
				response.Results = append(response.Results, GrantPermissionsResponseItem{
					Permission: perm,
					Granted:    false,
				})
				h.mu.Unlock()
				response.Failures++
				return
			}
			if err := h.permRepo.Grant(ctx, user.ID, permission.ID); err != nil {
				h.mu.Lock()
				response.Results = append(response.Results, GrantPermissionsResponseItem{
					Permission: perm,
					Granted:    false,
				})
				h.mu.Unlock()
				response.Failures++
				return
			}
			h.mu.Lock()
			response.Results = append(response.Results, GrantPermissionsResponseItem{
				Permission: perm,
				Granted:    true,
			})
			h.mu.Unlock()
			response.Successes++
		}()
	}
	wg.Wait()

	return response, nil
}
