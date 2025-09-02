package model

import "time"

// AfterRegistrationRequest represents the webhook payload from Kratos after user registration
type AfterRegistrationRequest struct {
	Identity       IdentityData       `json:"identity"`
	Flow           FlowData           `json:"flow"`
	RequestContext RequestContextData `json:"request_context"`
}

// IdentityData contains user identity information
type IdentityData struct {
	ID        string     `json:"id"`
	Traits    UserTraits `json:"traits"`
	SchemaID  string     `json:"schema_id"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UserTraits contains user profile and preference data
type UserTraits struct {
	Email        string          `json:"email"`
	DisplayName  string          `json:"display_name"`
	FirstName    *string         `json:"first_name"`
	LastName     *string         `json:"last_name"`
	UserType     string          `json:"user_type"`
	ArtistName   *string         `json:"artist_name"`
	Bio          *string         `json:"bio"`
	Location     *string         `json:"location"`
	ProfileImage *string         `json:"profile_image"`
	Preferences  UserPreferences `json:"preferences"`
}

// UserPreferences contains user notification and marketing preferences
type UserPreferences struct {
	EmailNotifications bool `json:"email_notifications"`
	MarketingEmails    bool `json:"marketing_emails"`
}

// FlowData contains registration flow information
type FlowData struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	ExpiresAt  time.Time `json:"expires_at"`
	IssuedAt   time.Time `json:"issued_at"`
	RequestURL string    `json:"request_url"`
}

// RequestContextData contains request metadata for logging and analytics
type RequestContextData struct {
	Method    string  `json:"method"`
	URL       string  `json:"url"`
	UserAgent *string `json:"user_agent"`
	IPAddress *string `json:"ip_address"`
	Timestamp string  `json:"timestamp"`
}
