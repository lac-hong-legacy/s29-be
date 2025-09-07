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

	kratosIdentityID, err := uuid.Parse(request.Identity.ID)
	if err != nil {
		log.Printf("Invalid Kratos identity ID in recovery webhook: %v", err)
		jsonResponse.ResponseBadRequest(c, "Invalid identity ID")
		return nil
	}

	user, err := h.authService.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		log.Printf("User not found during recovery webhook: %v", err)
		jsonResponse.ResponseOK(c, fiber.Map{"message": "Recovery processed"})
		return nil
	}

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
