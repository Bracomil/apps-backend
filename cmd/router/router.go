package router

import (
	"os"
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/handlers"
	"github.com/bracomil/bracomil-internal-api-back/cmd/httpclient"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/auth"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/bling"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/db"
	permission "github.com/bracomil/bracomil-internal-api-back/cmd/internal/permissions"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/user"
	"github.com/bracomil/bracomil-internal-api-back/cmd/middlewares"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/utils"
	"github.com/bracomil/bracomil-internal-api-back/cmd/tokenmanager"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/time/rate"
)

func handleHealth(c fiber.Ctx) {
	c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"health": "ok",
	})
}

func Setup(log *logger.Logger, db *db.DB, app *fiber.App) error {
	// Create handlers
	cnab400Handler := handlers.NewCnab400(log)

	// ~2.2 req/s - Abaixo do 3/s da documentação
	limiter := rate.NewLimiter(rate.Every(450*time.Millisecond), 1)

	storage := bling.NewTokenRepository(db)
	params := tokenmanager.OAuth2ManagerParams{
		ProviderURL:     "https://api.bling.com.br/Api/v3",
		RefreshTokenTTL: 30 * 24 * 60 * 60 * time.Second,
	}
	blingTokenManager, err := tokenmanager.NewOAuth2Manager(log, nil, storage, params, tokenmanager.WithLimiter(limiter))
	if err != nil {
		return err
	}
	blingTokenHandler := tokenmanager.NewHandler(log, blingTokenManager)

	baseUrl := utils.GetEnv("BLING_API_URI", "https://api.bling.com.br/Api/v3")
	blingClient := httpclient.New(baseUrl, blingTokenManager, httpclient.WithLimiter(limiter))
	blingHandler := bling.NewHandler(log, blingClient)

	// Auth Handler
	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	permissionRepository := permission.NewRepository(db)
	googleVerifier := auth.NewGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID"))
	jwtIssuer := auth.NewJWTIssuer("AUTH_SECRET", os.Getenv("HOST"), 24*time.Hour)
	authService := auth.NewService(userService, permissionRepository, googleVerifier, jwtIssuer)
	authHandler := auth.NewHandler(log, authService)

	// Auth Middleware
	authMiddleware := middlewares.NewAuthorizationMiddleware(jwtIssuer, userRepository, permissionRepository)

	// ==================================== DECLARE ROUTES ==================================== //
	// Utilities Routes
	app.Get("/health", handleHealth)
	app.Get("/temp", blingHandler.Temp)

	// Authorization route to get tokens from bling
	app.Get("/jwt/bling/authorize", blingTokenHandler.Authorize)
	// Google oauth callback route
	app.Post("/oauth/google", authHandler.GoogleLogin)

	protected := app.Group("/")
	protected.Use(authMiddleware.RequireAuthentication)
	// CNAB400 Routes
	protected.Post("/bnb/cnab400/return", cnab400Handler.ParseReturnFileForBNB)
	// Bling Routes
	protected.Get("/bling/contas/receber", blingHandler.ListReceivables)
	protected.Post("/bling/contas/receber/baixar", blingHandler.SettleBatchReceipts)
	return nil
}
