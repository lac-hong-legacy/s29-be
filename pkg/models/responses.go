package models

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2023-08-31T12:00:00Z"`
	Version   string `json:"version" example:"1.0.0"`
}

// PingResponse represents the ping response
type PingResponse struct {
	Message   string `json:"message" example:"pong"`
	Timestamp string `json:"timestamp" example:"2023-08-31T12:00:00Z"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Internal server error"`
	Code    int    `json:"code" example:"500"`
	Details string `json:"details,omitempty" example:"Additional error details"`
}
