package middlewares

import (
	"strings"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/auth"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/permissions"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/user"
	"github.com/gofiber/fiber/v3"
)

type AuthorizationMiddleware struct {
	jwt      *auth.JWTIssuer
	userRepo user.Repository
	permRepo permissions.Repository
}

func NewAuthorizationMiddleware(jwt *auth.JWTIssuer, userRepo user.Repository, permRepo permissions.Repository) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		jwt:      jwt,
		userRepo: userRepo,
		permRepo: permRepo,
	}
}

// Authenticate the user requesting from the API
func (m *AuthorizationMiddleware) RequireAuthentication(c fiber.Ctx) {
	// Checar presença do token
	authToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if authToken == "" {
		c.SendStatus(fiber.StatusUnauthorized)
		return
	}
	// Validar token
	claims, err := m.jwt.Parse(authToken)
	if err != nil {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err})
		return
	}
	if claims.Id == 0 {
		c.SendStatus(fiber.StatusUnauthorized)
		return
	}
	// Fazer buscas no banco para preencher o RequestContext
	user, err := m.userRepo.GetByID(c.Context(), claims.Id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
		return
	}
	permissions, err := m.permRepo.ListByUserID(c.Context(), user.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
		return
	}
	// Colocar user no locals para as requests poderem receber os dados do banco sem fazer outra request
	rc := &RequestContext{
		User:        user,
		Permissions: permissions,
	}
	SetRequestContext(c, rc)

	c.Next()
}
