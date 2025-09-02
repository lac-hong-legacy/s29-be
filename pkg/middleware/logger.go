package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogEntry struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	Timestamp    time.Time         `json:"timestamp"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Status       int               `json:"status"`
	Latency      float64           `json:"latency_ms"`
	ClientIP     string            `json:"client_ip"`
	UserAgent    string            `json:"user_agent"`
	RequestBody  string            `json:"request_body,omitempty"`
	ResponseBody string            `json:"response_body,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	ServiceName  string            `json:"service_name"`
	Environment  string            `json:"environment"`
	Error        string            `json:"error,omitempty"`
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Create log entry
		logEntry := LogEntry{
			TraceID:     c.Get("X-Trace-ID", generateID()),
			SpanID:      c.Get("X-Span-ID", generateID()),
			Timestamp:   time.Now(),
			Method:      c.Method(),
			Path:        c.Path(),
			Status:      c.Response().StatusCode(),
			Latency:     float64(latency.Nanoseconds()) / 1e6, // Convert to milliseconds
			ClientIP:    c.IP(),
			UserAgent:   c.Get("User-Agent"),
			ServiceName: "s29-api",
			Environment: getEnv("APP_ENV", "development"),
		}

		// Add error if exists
		if err != nil {
			logEntry.Error = err.Error()
		}

		// Log based on environment
		if getEnv("APP_ENV", "development") == "production" {
			// JSON format for production
			logJSON, _ := json.Marshal(logEntry)
			fmt.Println(string(logJSON))
		} else {
			// Human-readable format for development
			fmt.Printf("[%s] %d - %s %s | %s | %.2fms | %s\n",
				logEntry.Timestamp.Format("2006-01-02 15:04:05"),
				logEntry.Status,
				logEntry.Method,
				logEntry.Path,
				logEntry.ClientIP,
				logEntry.Latency,
				logEntry.UserAgent,
			)
		}

		return err
	}
}

// Simple ID generator for trace/span IDs
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Helper function to get environment variables with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
