package handlers

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/IsaqueB/cnab400-bnb/cnab400"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

const NAME_BANCO_DO_NORDESTE = "B.DO NORDESTE"

var ACCEPTED_BANKS_NAME = []string{NAME_BANCO_DO_NORDESTE}

type Cnab400Handler struct {
	log *logger.Logger
}

func NewCnab400(log *logger.Logger) *Cnab400Handler {
	return &Cnab400Handler{
		log: log,
	}
}

var ACCEPTABLE_EXTENSIONS = []string{".ret", ".sai"}

type BankReturnRequest struct {
}

func (h Cnab400Handler) ParseReturnFileForBNB(c fiber.Ctx) {
	fileHeader, err := c.FormFile("arquivo_retorno")
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not get return file",
			"error":   err.Error(),
		})
		return
	}

	ext := filepath.Ext(fileHeader.Filename)
	if !slices.Contains(ACCEPTABLE_EXTENSIONS, strings.ToLower(ext)) {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "This file is not supported",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"message": "Could not open provided file",
			"error":   err.Error(),
		})
		return
	}
	defer file.Close()

	returnFile, err := cnab400.ParseReturn(file, nil)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "Was not able to parse the file provided",
			"error":   err.Error(),
		})
		return
	}

	if !slices.Contains(ACCEPTED_BANKS_NAME, returnFile.Header.BankName) {
		c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "The bank that provided this return file is not supported",
			"error":   err.Error(),
		})
		return
	}
	c.Status(fiber.StatusOK).JSON(returnFile)
}
