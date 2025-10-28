package utils

import (
	"context"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	// Test that request ID is generated
	requestID := GenerateRequestID()
	if requestID == "" {
		t.Error("GenerateRequestID() returned empty string")
	}

	// Test that request ID has expected length (32 characters for 16 bytes in hex)
	if len(requestID) != 32 {
		t.Errorf("GenerateRequestID() length = %d, want 32", len(requestID))
	}

	// Test that multiple calls generate different IDs
	requestID2 := GenerateRequestID()
	if requestID == requestID2 {
		t.Error("GenerateRequestID() generated same ID twice")
	}
}

func TestSetRequestID(t *testing.T) {
	ctx := context.Background()
	testID := "test-request-id-123"

	newCtx := SetRequestID(ctx, testID)

	// Verify the context is not nil
	if newCtx == nil {
		t.Fatal("SetRequestID() returned nil context")
	}

	// Verify we can retrieve the value
	retrievedID := GetRequestID(newCtx)
	if retrievedID != testID {
		t.Errorf("GetRequestID() = %v, want %v", retrievedID, testID)
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "returns request ID from context",
			setup: func() context.Context {
				return SetRequestID(context.Background(), "test-id-456")
			},
			expected: "test-id-456",
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
				return context.WithValue(context.Background(), requestIDKey, nil)
			},
			expected: "",
		},
		{
			name: "returns empty string for wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), requestIDKey, 12345)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := GetRequestID(ctx)
			if result != tt.expected {
				t.Errorf("GetRequestID() = %v, want %v", result, tt.expected)
			}
		})
	}
}
