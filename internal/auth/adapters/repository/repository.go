// internal/auth/adapters/repository/repository.go
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

// FindUserByKratosIdentityID finds a user by their Kratos identity ID
func (r *AuthRepository) FindUserByKratosIdentityID(kratosIdentityID uuid.UUID) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("kratos_identity_id = ?", kratosIdentityID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID finds a user by their internal ID
func (r *AuthRepository) FindUserByID(userID uint64) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail finds a user by their email
func (r *AuthRepository) FindUserByEmail(email string) (*userModel.User, error) {
	var user userModel.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserLastLogin updates the user's last login timestamp
func (r *AuthRepository) UpdateUserLastLogin(user *userModel.User) error {
	return r.db.Model(user).Update("last_login_at", user.LastLoginAt).Error
}

// DeactivateUser deactivates a user account
func (r *AuthRepository) DeactivateUser(userID uint64) error {
	return r.db.Model(&userModel.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

// ActivateUser activates a user account
func (r *AuthRepository) ActivateUser(userID uint64) error {
	return r.db.Model(&userModel.User{}).Where("id = ?", userID).Update("is_active", true).Error
}
