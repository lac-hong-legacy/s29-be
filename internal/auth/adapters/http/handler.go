package http

import (
	"fmt"
	"net/url"
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

type OAuthInitRequest struct {
	Provider    string `json:"provider" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

type OAuthResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
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

// @Summary Initiate OAuth Login
// @Description Start OAuth login flow for mobile apps
// @Tags Auth
// @Accept json
// @Produce json
// @Param oauthRequest body OAuthInitRequest true "OAuth Init Request"
// @Success 200 {object} OAuthResponse
// @Router /api/v1/auth/oauth/init [post]
func (h *AuthHandler) InitiateOAuth(c *fiber.Ctx) error {
	var request OAuthInitRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	if request.Provider != "google" && request.Provider != "facebook" {
		jsonResponse.ResponseBadRequest(c, "Invalid provider. Supported: google, facebook")
		return nil
	}

	if request.RedirectURI == "" {
		request.RedirectURI = "s29app://oauth/callback"
	}

	authURL, state, err := h.authService.InitiateOAuthFlow(request.Provider, request.RedirectURI, request.State)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	response := &OAuthResponse{
		AuthURL: authURL,
		State:   state,
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}

// @Summary Handle OAuth Callback
// @Description Handle OAuth callback and complete authentication
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth Provider"
// @Param code query string true "Authorization Code"
// @Param state query string false "State parameter"
// @Success 200 {object} domain.LoginResponse
// @Router /api/v1/auth/oauth/{provider}/callback [get]
func (h *AuthHandler) HandleOAuthCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	code := c.Query("code")
	state := c.Query("state")

	if provider == "" || code == "" {
		jsonResponse.ResponseBadRequest(c, "Missing provider or authorization code")
		return nil
	}

	if provider != "google" && provider != "facebook" {
		jsonResponse.ResponseBadRequest(c, "Invalid provider. Supported: google, facebook")
		return nil
	}

	response, err := h.authService.HandleOAuthCallback(provider, code, state)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	// For mobile apps, redirect to deep link with tokens
	redirectURL := fmt.Sprintf("s29app://auth/success?access_token=%s&token_type=%s&expires_in=%d",
		url.QueryEscape(response.AccessToken),
		response.TokenType,
		response.ExpiresIn)

	c.Redirect(redirectURL, 302)
	return nil
}

// @Summary Complete OAuth Login
// @Description Complete OAuth login flow for mobile apps (POST alternative to callback)
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth Provider"
// @Param callbackRequest body OAuthCallbackRequest true "OAuth Callback Request"
// @Success 200 {object} domain.LoginResponse
// @Router /api/v1/auth/oauth/{provider}/complete [post]
func (h *AuthHandler) CompleteOAuthLogin(c *fiber.Ctx) error {
	provider := c.Params("provider")
	var request OAuthCallbackRequest

	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	if provider != "google" && provider != "facebook" {
		jsonResponse.ResponseBadRequest(c, "Invalid provider. Supported: google, facebook")
		return nil
	}

	response, err := h.authService.HandleOAuthCallback(provider, request.Code, request.State)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, response)
	return nil
}
