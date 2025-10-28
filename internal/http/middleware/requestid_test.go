package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-pubsub-logger/internal/utils"
)

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		existingID       string
		expectGenerated  bool
		expectInResponse bool
	}{
		{
			name:             "generates new request ID when not provided",
			existingID:       "",
			expectGenerated:  true,
			expectInResponse: true,
		},
		{
			name:             "uses existing request ID from header",
			existingID:       "existing-req-id-123",
			expectGenerated:  false,
			expectInResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks if request ID is in context
			var contextRequestID string
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextRequestID = utils.GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with middleware
			handler := RequestIDMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingID != "" {
				req.Header.Set("X-Request-ID", tt.existingID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Check response header
			responseRequestID := rr.Header().Get("X-Request-ID")
			if tt.expectInResponse && responseRequestID == "" {
				t.Error("Expected X-Request-ID in response header, but got empty")
			}

			// Check context value
			if contextRequestID == "" {
				t.Error("Expected request ID in context, but got empty")
			}

			// If existing ID was provided, verify it's used
			if !tt.expectGenerated {
				if contextRequestID != tt.existingID {
					t.Errorf("Expected context request ID = %v, got %v", tt.existingID, contextRequestID)
				}
				if responseRequestID != tt.existingID {
					t.Errorf("Expected response request ID = %v, got %v", tt.existingID, responseRequestID)
				}
			}

			// If generated, verify it's valid
			if tt.expectGenerated {
				if len(contextRequestID) != 32 {
					t.Errorf("Expected generated request ID length = 32, got %d", len(contextRequestID))
				}
				if responseRequestID != contextRequestID {
					t.Error("Response request ID should match context request ID")
				}
			}
		})
	}
}

func TestRequestIDMiddleware_PropagatesContext(t *testing.T) {
	var receivedRequestID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequestID = utils.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestIDMiddleware(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if receivedRequestID != "test-123" {
		t.Errorf("Expected request ID in context = %v, got %v", "test-123", receivedRequestID)
	}
}
