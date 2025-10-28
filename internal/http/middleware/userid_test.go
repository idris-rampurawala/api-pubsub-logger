package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-pubsub-logger/internal/utils"
)

func TestUserIDMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expectInContext bool
	}{
		{
			name:            "extracts user ID from header",
			userID:          "user-123",
			expectInContext: true,
		},
		{
			name:            "handles missing user ID header",
			userID:          "",
			expectInContext: true, // Context will have empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var contextUserID string
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextUserID = utils.GetUserID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := UserIDMiddleware(testHandler)
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if tt.expectInContext {
				if contextUserID != tt.userID {
					t.Errorf("Expected user ID in context = %v, got %v", tt.userID, contextUserID)
				}
			}
		})
	}
}

func TestUserIDMiddleware_PropagatesContext(t *testing.T) {
	expectedUserID := "test-user-456"
	var receivedUserID string

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserID = utils.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := UserIDMiddleware(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", expectedUserID)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if receivedUserID != expectedUserID {
		t.Errorf("Expected user ID in context = %v, got %v", expectedUserID, receivedUserID)
	}
}

func TestUserIDMiddleware_WithMultipleHeaders(t *testing.T) {
	// Test that middleware handles multiple headers correctly
	var receivedUserID string

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserID = utils.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := UserIDMiddleware(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "first-user")
	req.Header.Add("X-User-ID", "second-user") // Add another value

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should get the first value
	if receivedUserID != "first-user" {
		t.Errorf("Expected user ID = %v, got %v", "first-user", receivedUserID)
	}
}
