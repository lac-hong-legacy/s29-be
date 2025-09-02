package repository

import (
	model "s29-be/internal/user/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	// db.AutoMigrate(&model.User{}, &model.UserFavorites{}, &model.UserPreference{})
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUserAfterRegistration(user *model.User) (*model.User, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}
