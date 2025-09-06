// internal/auth/module.go
package auth

import (
	"os"
	"s29-be/internal/auth/adapters/http"
	"s29-be/internal/auth/adapters/repository"
	"s29-be/internal/auth/application"
	"s29-be/pkg/jwt"
	"s29-be/pkg/kratos"
	"s29-be/pkg/middleware"
	"time"

	svcContext "s29-be/pkg/context"

	"github.com/gofiber/fiber/v2"
)

type AuthModule struct {
	Repository   *repository.AuthRepository
	Service      *application.AuthService
	Handler      *http.AuthHandler
	Middleware   *middleware.AuthMiddleware
	KratosClient *kratos.Client
	JWTService   *jwt.JWTService
}

func NewAuthModule(ctx2 *svcContext.ServiceContext) *AuthModule {
	// Initialize Kratos client
	kratosPublicURL := os.Getenv("KRATOS_PUBLIC_URL")
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosPublicURL == "" {
		kratosPublicURL = "http://localhost:4433"
	}
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
	}
	kratosClient := kratos.NewClient(kratosPublicURL, kratosAdminURL)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-jwt-secret-here" // Default for development
	}
	jwtService := jwt.NewJWTService(jwtSecret, "audora-api", 24*time.Hour)

	authRepo := repository.NewAuthRepository(ctx2.GetDB())
	authService := application.NewAuthService(authRepo, kratosClient, jwtService)
	authHandler := http.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	return &AuthModule{
		Repository:   authRepo,
		Service:      authService,
		Handler:      authHandler,
		Middleware:   authMiddleware,
		KratosClient: kratosClient,
		JWTService:   jwtService,
	}
}

func (a *AuthModule) RegisterRoutes(router fiber.Router) {
	auth := router.Group("auth")
	{
		// Public endpoints
		auth.Post("/login", a.Handler.Login)                  // Login with session token from body
		auth.Post("/login/cookie", a.Handler.LoginWithCookie) // Login with session token from cookie
		auth.Post("/refresh", a.Handler.RefreshToken)         // Refresh JWT token
		auth.Post("/validate", a.Handler.ValidateToken)       // Validate token (for other services)

		// Protected endpoints
		protected := auth.Group("")
		protected.Use(a.Middleware.RequireAuth())
		{
			protected.Get("/me", a.Handler.Me)          // Get current user info
			protected.Post("/logout", a.Handler.Logout) // Logout (optional)
		}
	}

	internal := router.Group("internal/hooks")
	{
		internal.Post("/after-recovery", a.Handler.AfterRecovery)
	}
}
