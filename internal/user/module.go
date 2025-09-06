package user

import (
	"s29-be/internal/user/adapters/http"
	"s29-be/internal/user/adapters/repository"
	"s29-be/internal/user/application"
	ctx2 "s29-be/pkg/context"

	"github.com/gofiber/fiber/v2"
)

type UserModule struct {
	Repository *repository.UserRepository
	Service    *application.UserService
	Handler    *http.UserHandler
}

func NewUserModule(serviceContext *ctx2.ServiceContext) *UserModule {
	userRepo := repository.NewUserRepository(serviceContext.GetDB())
	userService := application.NewUserService(userRepo)
	userHandler := http.NewUserHandler(userService)

	return &UserModule{
		Repository: userRepo,
		Service:    userService,
		Handler:    userHandler,
	}
}

func (u *UserModule) RegisterRoutes(router fiber.Router) {
	internal := router.Group("/internal")
	internal.Post("/hooks/after-registration", u.Handler.AfterRegistration)
}
