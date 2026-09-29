package middlewares

import "github.com/gofiber/fiber/v3"

type PermissionMiddleware struct {
}

func (m *PermissionMiddleware) RequirePermissions(perm string) fiber.Handler {
	return func(c fiber.Ctx) error {
		rc, ok := GetRequestContext(c)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "not authenticate")
		}
		for _, p := range rc.Permissions {
			if p == perm {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}
}
