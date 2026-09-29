package middlewares

import (
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/user"
	"github.com/gofiber/fiber/v3"
)

// RequestContext é o que o middleware injeta na request.
type RequestContext struct {
	User        *user.User
	Permissions []string
}

const localsKey = "auth.request_context"

func SetRequestContext(c fiber.Ctx, rc *RequestContext) {
	c.Locals(localsKey, rc)
}

func GetRequestContext(c fiber.Ctx) (*RequestContext, bool) {
	rc, ok := c.Locals(localsKey).(*RequestContext)
	return rc, ok
}

func MustGetRequestContext(c fiber.Ctx) *RequestContext {
	rc, ok := GetRequestContext(c)
	if !ok {
		panic("auth: request context not set; did you forget RequireAuth middleware?")
	}
	return rc
}
