package model

import (
	"s29-be/pkg/model"
	"time"

	"github.com/google/uuid"
)

type User struct {
	model.BaseModel
	KratosIdentityID uuid.UUID  `json:"kratos_identity_id" gorm:"not null;unique;type:uuid"`
	Email            string     `json:"email" gorm:"not null;unique;size:255"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt      *time.Time `json:"last_login_at"`
}
