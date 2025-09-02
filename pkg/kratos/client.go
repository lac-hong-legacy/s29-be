package kratos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	Active    bool      `json:"active"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Identity  Identity  `json:"identity"`
}

type Identity struct {
	ID                  string                 `json:"id"`
	SchemaID            string                 `json:"schema_id"`
	SchemaURL           string                 `json:"schema_url"`
	State               string                 `json:"state"`
	StateChangedAt      time.Time              `json:"state_changed_at"`
	Traits              map[string]interface{} `json:"traits"`
	VerifiableAddresses []VerifiableAddress    `json:"verifiable_addresses"`
	RecoveryAddresses   []RecoveryAddress      `json:"recovery_addresses"`
	MetadataPublic      json.RawMessage        `json:"metadata_public"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type VerifiableAddress struct {
	ID        string    `json:"id"`
	Value     string    `json:"value"`
	Verified  bool      `json:"verified"`
	Via       string    `json:"via"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RecoveryAddress struct {
	ID        string    `json:"id"`
	Value     string    `json:"value"`
	Via       string    `json:"via"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type KratosError struct {
	Code    int                    `json:"code"`
	Status  string                 `json:"status"`
	Request string                 `json:"request"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *KratosError) Error() string {
	return fmt.Sprintf("kratos error %d: %s", e.Code, e.Message)
}

type Client struct {
	publicURL string
	adminURL  string
	client    *http.Client
}

func NewClient(publicURL, adminURL string) *Client {
	return &Client{
		publicURL: publicURL,
		adminURL:  adminURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifySession validates a session using Kratos /sessions/whoami endpoint
func (c *Client) VerifySession(sessionToken string) (*Session, error) {
	if sessionToken == "" {
		return nil, &KratosError{
			Code:    401,
			Status:  "Unauthorized",
			Message: "no session token provided",
		}
	}

	url := fmt.Sprintf("%s/sessions/whoami", c.publicURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Session-Token", sessionToken)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var kratosErr KratosError
		if err := json.Unmarshal(body, &kratosErr); err != nil {
			return nil, fmt.Errorf("invalid session: status %d", resp.StatusCode)
		}
		return nil, &kratosErr
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Validate session is active and not expired
	if !session.Active {
		return nil, &KratosError{
			Code:    401,
			Status:  "Unauthorized",
			Message: "session is not active",
		}
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, &KratosError{
			Code:    401,
			Status:  "Unauthorized",
			Message: "session has expired",
		}
	}

	return &session, nil
}

// GetTraits safely extracts traits from identity
func (i *Identity) GetTraits() map[string]interface{} {
	if i.Traits == nil {
		return make(map[string]interface{})
	}
	return i.Traits
}

// GetEmail extracts email from traits
func (i *Identity) GetEmail() string {
	traits := i.GetTraits()
	if email, ok := traits["email"].(string); ok {
		return email
	}
	return ""
}

// GetDisplayName extracts display name from traits
func (i *Identity) GetDisplayName() string {
	traits := i.GetTraits()
	if displayName, ok := traits["display_name"].(string); ok {
		return displayName
	}
	return ""
}

// GetUserType extracts user type from traits
func (i *Identity) GetUserType() string {
	traits := i.GetTraits()
	if userType, ok := traits["user_type"].(string); ok {
		return userType
	}
	return "listener" // default
}
