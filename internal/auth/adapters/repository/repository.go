package repository

import (
	userModel "s29-be/internal/user/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) FindUserByKratosIdentityID(kratosIdentityID uuid.UUID) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("kratos_identity_id = ?", kratosIdentityID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByID(userID uint64) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByEmail(email string) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdateUserLastLogin(user *userModel.User) error {
	return r.db.Model(user).Update("last_login_at", user.LastLoginAt).Error
}
