package application

import (
	"s29-be/internal/user/adapters/repository"
	model "s29-be/internal/user/domain"
	baseModel "s29-be/pkg/model"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUserAfterRegistration(user *model.AfterRegistrationRequest) (*uuid.UUID, error) {
	identityID, err := uuid.Parse(user.Identity.ID)
	if err != nil {
		return nil, err
	}

	baseModelInstance, err := baseModel.NewBaseModel()
	if err != nil {
		return nil, err
	}

	userModel, err := s.userRepo.CreateUserAfterRegistration(&model.User{
		BaseModel:        *baseModelInstance,
		KratosIdentityID: identityID,
		Email:            user.Identity.Traits.Email,
		IsActive:         true,
		LastLoginAt:      nil,
	})

	if err != nil {
		return nil, err
	}

	return &userModel.ID, nil
}
