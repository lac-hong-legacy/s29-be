package http

import (
	"s29-be/internal/auth/application"
	"s29-be/internal/auth/domain"
	appError "s29-be/pkg/error"
	jsonResponse "s29-be/pkg/json"
	"s29-be/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *application.AuthService
}

type LoginRequest struct {
	SessionToken string `json:"session_token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	SessionToken string `json:"session_token" binding:"required"` // NEW: Required for security
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) HandleError(c *fiber.Ctx, err error) bool {
	if err == nil {
		return false
	}

	if appErr, ok := appError.GetAppError(err); ok {
		jsonResponse.ResponseJSON(c, appErr.StatusCode, appErr.Message, appErr.Data)
		return true
	}

	jsonResponse.ResponseInternalError(c, err)
	return true
}

// Login validates Kratos session and issues Audora JWT
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request LoginRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	response, err := h.authService.VerifySessionAndIssueJWT(request.SessionToken)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// LoginWithCookie validates Kratos session from cookie and issues Audora JWT
func (h *AuthHandler) LoginWithCookie(c *fiber.Ctx) error {
	// Extract session token from cookie
	sessionToken := c.Cookies("ory_kratos_session")
	if sessionToken == "" {
		jsonResponse.ResponseUnauthorized(c)
		return nil
	}

	response, err := h.authService.VerifySessionAndIssueJWT(sessionToken)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// FIXED: RefreshToken now validates with Kratos session
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var request RefreshTokenRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// SECURITY FIX: Pass both the current JWT and session token for validation
	response, err := h.authService.RefreshToken(request.RefreshToken, request.SessionToken)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// Alternative: RefreshTokenWithCookie for cookie-based session validation
func (h *AuthHandler) RefreshTokenWithCookie(c *fiber.Ctx) error {
	// Extract refresh token from request body
	var refreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.BodyParser(&refreshRequest); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// Extract session token from cookie
	sessionToken := c.Cookies("ory_kratos_session")
	if sessionToken == "" {
		jsonResponse.ResponseBadRequest(c, "Session cookie required for refresh")
		return nil
	}

	response, err := h.authService.RefreshToken(refreshRequest.RefreshToken, sessionToken)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// Me returns current user information from JWT
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get claims from middleware context
	claims := c.Locals("user_claims")
	if claims == nil {
		jsonResponse.ResponseUnauthorized(c)
		return nil
	}

	userInfo, err := h.authService.GetCurrentUser(claims.(*jwt.Claims))
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, userInfo)
	return nil
}

// Logout invalidates the current session (optional - mainly handled by frontend)
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Since we're using stateless JWT, logout is mainly handled client-side
	// But we can add token blacklisting here if needed in the future
	jsonResponse.ResponseOK(c, fiber.Map{"message": "Successfully logged out"})
	return nil
}

// ValidateToken endpoint for other services to validate tokens
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
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

	claims, err := h.authService.ValidateJWT(tokenString)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	userInfo := &domain.UserInfo{
		ID:               claims.UserID,
		KratosIdentityID: claims.KratosIdentityID,
		Email:            claims.Email,
		DisplayName:      claims.DisplayName,
		UserType:         claims.UserType,
		IsActive:         claims.IsActive,
	}

	jsonResponse.ResponseOK(c, userInfo)
	return nil
}
