package model

import "time"

type AfterRegistrationRequest struct {
	Identity       IdentityData       `json:"identity"`
	Flow           FlowData           `json:"flow"`
	RequestContext RequestContextData `json:"request_context"`
}

type IdentityData struct {
	ID        string     `json:"id"`
	Traits    UserTraits `json:"traits"`
	SchemaID  string     `json:"schema_id"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UserTraits struct {
	Email        string          `json:"email"`
	DisplayName  string          `json:"display_name"`
	YearOfBirth  int             `json:"year_of_birth"`
}

type FlowData struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	ExpiresAt  time.Time `json:"expires_at"`
	IssuedAt   time.Time `json:"issued_at"`
	RequestURL string    `json:"request_url"`
}

type RequestContextData struct {
	Method    string  `json:"method"`
	URL       string  `json:"url"`
	UserAgent *string `json:"user_agent"`
	IPAddress *string `json:"ip_address"`
	Timestamp string  `json:"timestamp"`
}
