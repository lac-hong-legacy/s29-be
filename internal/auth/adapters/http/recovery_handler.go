package http

import (
	"log"
	"s29-be/internal/auth/domain"
	jsonResponse "s29-be/pkg/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *AuthHandler) AfterRecovery(c *fiber.Ctx) error {
	var request domain.RecoveryWebhookRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Failed to parse recovery webhook request: %v", err)
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// Parse the Kratos identity ID
	kratosIdentityID, err := uuid.Parse(request.Identity.ID)
	if err != nil {
		log.Printf("Invalid Kratos identity ID in recovery webhook: %v", err)
		jsonResponse.ResponseBadRequest(c, "Invalid identity ID")
		return nil
	}

	// Find the user in our database
	user, err := h.authService.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		log.Printf("User not found during recovery webhook: %v", err)
		// Don't fail the webhook - Kratos should still complete the recovery
		jsonResponse.ResponseOK(c, fiber.Map{"message": "Recovery processed"})
		return nil
	}

	// Log the recovery event (you might want to store this in your database)
	log.Printf("Password recovery completed for user ID: %s, Email: %s", user.ID, user.Email)

	// Optional: Update last login time or add recovery audit log
	now := time.Now()
	user.LastLoginAt = &now
	if err := h.authService.UpdateUserLastLogin(user); err != nil {
		log.Printf("Failed to update last login time after recovery: %v", err)
		// Don't fail the webhook
	}

	// Optional: Send a notification email about successful password reset
	// You could implement this with your email service

	jsonResponse.ResponseOK(c, fiber.Map{
		"message":      "Recovery webhook processed successfully",
		"user_id":      user.ID,
		"recovered_at": request.RecoveryInfo.RecoveredAt,
	})
	return nil
}

// InitiateRecovery starts the password recovery flow with OTP
func (h *AuthHandler) InitiateRecovery(c *fiber.Ctx) error {
	type RecoveryRequest struct {
		Email string `json:"email" binding:"required"`
	}

	var request RecoveryRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// Initiate recovery flow with Kratos
	recoveryFlow, err := h.authService.InitiatePasswordRecovery(request.Email)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, fiber.Map{
		"message":       "Recovery code sent if account exists",
		"flow_id":       recoveryFlow.ID,
		"requires_code": true,
	})
	return nil
}

// SetNewPassword sets a new password after OTP verification
func (h *AuthHandler) SetNewPassword(c *fiber.Ctx) error {
	type SetPasswordRequest struct {
		FlowID   string `json:"flow_id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var request SetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// Validate password strength
	if len(request.Password) < 8 {
		jsonResponse.ResponseBadRequest(c, "Password must be at least 8 characters long")
		return nil
	}

	result, err := h.authService.SetNewPassword(request.FlowID, request.Password)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, result)
	return nil
}

// VerifyRecoveryCode verifies the OTP code
func (h *AuthHandler) VerifyRecoveryCode(c *fiber.Ctx) error {
	type VerifyCodeRequest struct {
		FlowID string `json:"flow_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	var request VerifyCodeRequest
	if err := c.BodyParser(&request); err != nil {
		jsonResponse.ResponseBadRequest(c, "Invalid request: "+err.Error())
		return nil
	}

	// Validate code format (6 digits)
	if len(request.Code) != 6 {
		jsonResponse.ResponseBadRequest(c, "Code must be exactly 6 digits")
		return nil
	}

	result, err := h.authService.VerifyRecoveryCode(request.FlowID, request.Code)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	jsonResponse.ResponseOK(c, result)
	return nil
}