package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"api-pubsub-logger/pkg/logger"

	"github.com/google/uuid"
)

// In-memory storage for demo purposes
var items = []logger.Item{
	{
		ID:          "1",
		Name:        "Sample Item 1",
		Description: "This is a sample item for demonstration",
		Email:       "user1@example.com",
		PhoneNumber: "+1-555-0101",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	},
	{
		ID:          "2",
		Name:        "Sample Item 2",
		Description: "Another sample item",
		Email:       "user2@example.com",
		PhoneNumber: "+1-555-0102",
		CreatedAt:   time.Now().Add(-12 * time.Hour),
	},
}

// GetItems returns all items
func GetItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// CreateItem creates a new item
func CreateItem(w http.ResponseWriter, r *http.Request) {
	var req logger.CreateItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create new item
	newItem := logger.Item{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		CreatedAt:   time.Now(),
	}

	// Add to in-memory storage
	items = append(items, newItem)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
