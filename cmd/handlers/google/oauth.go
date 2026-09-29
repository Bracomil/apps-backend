package google

// import (
// 	"encoding/json"

// 	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
// 	"github.com/gofiber/fiber/v3"
// )

// const OAUTH_CERTIFICATE_URL = "https://www.googleapis.com/oauth2/v1/certs"

// type Handler struct {
// 	log *logger.Logger
// }

// func NewHandler(log *logger.Logger) *Handler {
// 	return &Handler{
// 		log: log,
// 	}
// }

// func (h *Handler) Login(c fiber.Ctx) {
// 	// csrf_cookie := c.Cookies("g_csrf_token")
// 	// csrf_header := c.GetHeaders()["g_csrf_token"]
// 	// if csrf_cookie == "" || len(csrf_header) < 1 || csrf_header[0] == "" || csrf_cookie != csrf_header[0] {
// 	// 	c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 	// 		"error": "csrf verification failed",
// 	// 	})
// 	// 	return
// 	// }

// 	var body GoogleOAuthCredential
// 	if err := json.Unmarshal(c.BodyRaw(), &body); err != nil {
// 		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 		return
// 	}

// 	if body.Token == "" {
// 		c.SendStatus(fiber.StatusBadRequest)
// 		return
// 	}
// 	h.login(c, body.Token)
// }
