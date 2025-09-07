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

// @Summary Login
// @Description Login with Kratos session token to receive Audora JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login Request"
// @Success 200 {object} domain.LoginResponse
// @Router /api/v1/auth/login [post]
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

// @Summary Refresh Token
// @Description Refresh JWT token using current JWT and Kratos session token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshTokenRequest body RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} domain.LoginResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var request RefreshTokenRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// SECURITY FIX: Pass both the current JWT and session token for validation
	response, err := h.authService.RefreshToken(request.SessionToken)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// @Summary Get Current User
// @Description Get information about the currently authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} domain.UserInfo
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
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

// @Summary Validate Token
// @Description Validate JWT token and return user info
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} domain.UserInfo
// @Router /api/v1/auth/validate [post]
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
