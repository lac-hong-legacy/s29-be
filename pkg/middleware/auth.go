// pkg/middleware/auth.go
package middleware

import (
	"s29-be/internal/auth/application"
	jsonResponse "s29-be/pkg/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	authService *application.AuthService
}

func NewAuthMiddleware(authService *application.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			jsonResponse.ResponseUnauthorized(c)
			return nil
		}

		// Extract Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			jsonResponse.ResponseBadRequest(c, "Invalid authorization header format")
			return nil
		}

		// Validate token
		claims, err := m.authService.ValidateJWT(tokenString)
		if err != nil {
			jsonResponse.ResponseUnauthorized(c)
			return nil
		}

		// Store claims in context for use in handlers
		c.Locals("user_claims", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("kratos_identity_id", claims.KratosIdentityID)
		c.Locals("user_type", claims.UserType)
		c.Locals("user_email", claims.Email)
		c.Locals("user_tier", "free") // Default to free tier, can be updated later

		return c.Next()
	}
}
