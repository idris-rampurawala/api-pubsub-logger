package logger

import "time"

// Item represents a simple item in our demo API
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Email       string    `json:"email,omitempty"`        // Sensitive field for masking demo
	PhoneNumber string    `json:"phone_number,omitempty"` // Sensitive field for masking demo
	CreatedAt   time.Time `json:"created_at"`
}

// CreateItemRequest represents the request body for creating an item
type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}
