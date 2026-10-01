package ofx

import (
	"path/filepath"
	"slices"
	"strings"

	ofxgo "github.com/IsaqueB/ofx-reader/ofx"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

var ACCEPTABLE_EXTENSIONS = []string{".ofx"}

type Handler struct {
	log *logger.Logger
}

func NewHandler(log *logger.Logger) *Handler {
	return &Handler{
		log: log,
	}
}

func (h *Handler) ParseFile(c fiber.Ctx) {
	fileHeader, err := c.FormFile("arquivo")
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": ErrFileNotSent,
		})
	}

	ext := filepath.Ext(fileHeader.Filename)
	if !slices.Contains(ACCEPTABLE_EXTENSIONS, strings.ToLower(ext)) {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": ErrFileNotSupported,
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": ErrFailedOpenFile,
		})
	}
	doc, err := ofxgo.Parse(file)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": ErrFailedParseFile,
		})
	}
	c.Status(fiber.StatusOK).JSON(doc)
}
