package http

import (
	user_service "s29-be/internal/user/application"
	model "s29-be/internal/user/domain"
	app_error "s29-be/pkg/error"
	json_response "s29-be/pkg/json"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *user_service.UserService
}

func NewUserHandler(userService *user_service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) HandleError(c *fiber.Ctx, err error) bool {
	if err == nil {
		return false
	}

	if appErr, ok := app_error.GetAppError(err); ok {
		json_response.ResponseJSON(c, appErr.StatusCode, appErr.Message, appErr.Data)
		return true
	}

	json_response.ResponseInternalError(c, err)
	return true
}

func (h *UserHandler) AfterRegistration(c *fiber.Ctx) error {
	var request model.AfterRegistrationRequest
	if err := c.BodyParser(&request); err != nil {
		json_response.ResponseBadRequest(c, err.Error())
		return nil
	}

	userID, err := h.userService.CreateUserAfterRegistration(&request)
	if err != nil {
		h.HandleError(c, err)
		return nil
	}

	json_response.ResponseOK(c, userID)
	return nil
}
