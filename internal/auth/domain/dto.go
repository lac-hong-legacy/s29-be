package domain

import (
	"time"

	"github.com/google/uuid"
)

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	ID               uuid.UUID `json:"id"`
	KratosIdentityID string    `json:"kratos_identity_id"`
	Email            string    `json:"email"`
	DisplayName      string    `json:"display_name"`
	UserType         string    `json:"user_type"`
	IsActive         bool      `json:"is_active"`
}

type RecoveryWebhookRequest struct {
	Identity     RecoveryIdentityData `json:"identity"`
	RecoveryInfo RecoveryInfo         `json:"recovery_info"`
}

type RecoveryIdentityData struct {
	ID       string                 `json:"id"`
	Traits   map[string]interface{} `json:"traits"`
	SchemaID string                 `json:"schema_id"`
	State    string                 `json:"state"`
}

type RecoveryInfo struct {
	FlowID         string    `json:"flow_id"`
	RecoveryMethod string    `json:"recovery_method"`
	RecoveredAt    time.Time `json:"recovered_at"`
}

type RecoveryFlow struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	ExpiresAt   string      `json:"expires_at"`
	IssuedAt    string      `json:"issued_at"`
	RequestURL  string      `json:"request_url"`
	UI          interface{} `json:"ui"`
	State       string      `json:"state"`
}

type RecoverySubmissionResult struct {
	Flow    RecoveryFlow `json:"flow"`
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Step    string       `json:"step"` // "code_sent", "code_verified", "password_set"
}