package utils

import (
	"context"
	"testing"
)

func TestSetUserID(t *testing.T) {
	ctx := context.Background()
	testUserID := "user-123"

	newCtx := SetUserID(ctx, testUserID)

	// Verify the context is not nil
	if newCtx == nil {
		t.Fatal("SetUserID() returned nil context")
	}

	// Verify we can retrieve the value
	retrievedID := GetUserID(newCtx)
	if retrievedID != testUserID {
		t.Errorf("GetUserID() = %v, want %v", retrievedID, testUserID)
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "returns user ID from context",
			setup: func() context.Context {
				return SetUserID(context.Background(), "user-456")
			},
			expected: "user-456",
		},
		{
			name: "returns empty string when not set",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
		{
			name: "returns empty string for nil context value",
			setup: func() context.Context {
				return context.WithValue(context.Background(), userIDKey, nil)
			},
			expected: "",
		},
		{
			name: "returns empty string for wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), userIDKey, 67890)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetUserID(ctx)
			if result != tt.expected {
				t.Errorf("GetUserID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMultipleContextValues(t *testing.T) {
	// Test that both request ID and user ID can coexist in the same context
	ctx := context.Background()
	ctx = SetRequestID(ctx, "req-123")
	ctx = SetUserID(ctx, "user-456")

	requestID := GetRequestID(ctx)
	userID := GetUserID(ctx)

	if requestID != "req-123" {
		t.Errorf("GetRequestID() = %v, want %v", requestID, "req-123")
	}

	if userID != "user-456" {
		t.Errorf("GetUserID() = %v, want %v", userID, "user-456")
	}
}
