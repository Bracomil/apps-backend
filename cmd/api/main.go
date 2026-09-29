package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/db"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/utils"
	"github.com/bracomil/bracomil-internal-api-back/cmd/router"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	fiberLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	env := "/Users/isaque/Repositories/Bracomil/bracomil-api/back/.env"
	godotenv.Load(env)

	log := logger.New(utils.GetEnvInt("LOG_MODE", 1))
	log.Info("Starting service using", env)

	// Setup fiber api
	app := fiber.New(fiber.Config{BodyLimit: 5 * 1024 * 1024})
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},                       // Lista de domínios permitidos
		AllowMethods:     []string{"GET, POST, PUT, DELETE, OPTIONS"},             // Métodos permitidos
		AllowHeaders:     []string{"Origin, Content-Type, Accept, Authorization"}, // Headers permitidos
		AllowCredentials: true,                                                    // Permite cookies/autenticação
		MaxAge:           3600,                                                    // Cache do preflight por 1 hora
	}))
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "America/Sao_Paulo",
	}))

	db, err := setupDatabase(log)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := router.Setup(log, db, app); err != nil {
		log.Fatal(err)
	}

	port := utils.GetEnv("PORT", "3000")
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}

func setupDatabase(log *logger.Logger) (*db.DB, error) {
	// Format for connection url to database
	databaseURLFormat := utils.GetEnv("DATABASE_URL_FORMAT", "")
	if databaseURLFormat == "" {
		return nil, errors.New("DATABASE_URL_FORMAT not set")
	}
	// Database access user
	databaseUser := utils.GetEnv("POSTGRES_USER", "")
	if databaseURLFormat == "" {
		return nil, errors.New("POSTGRES_USER not set")
	}
	// Database access user password - set to "" after using
	databasePwd := utils.GetEnv("POSTGRES_PASSWORD", "")
	if databaseURLFormat == "" {
		return nil, errors.New("POSTGRES_PASSWORD not set")
	}
	defer func() { databasePwd = "" }()
	// Database name
	databaseName := utils.GetEnv("POSTGRES_DB", "")
	if databaseURLFormat == "" {
		return nil, errors.New("POSTGRES_DB not set")
	}

	databaseURL := fmt.Sprintf(databaseURLFormat, databaseUser, databasePwd, databaseName)

	// Roda migrações antes de conectar
	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatal("migrate: %v", err)
	}

	return db.New(
		context.Background(),
		db.DefaultConfig(databaseURL),
	)
}
