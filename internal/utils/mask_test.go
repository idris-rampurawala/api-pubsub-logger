package utils

import (
	"encoding/json"
	"testing"
)

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "masks email field",
			input: `{"name": "John", "email": "john@example.com"}`,
			expected: map[string]interface{}{
				"name":  "John",
				"email": "***REDACTED***",
			},
		},
		{
			name:  "masks phone_number field",
			input: `{"name": "Jane", "phone_number": "+1-555-1234"}`,
			expected: map[string]interface{}{
				"name":         "Jane",
				"phone_number": "***REDACTED***",
			},
		},
		{
			name:  "masks multiple sensitive fields",
			input: `{"name": "Bob", "email": "bob@test.com", "phone_number": "+1-555-5678", "token": "secret123"}`,
			expected: map[string]interface{}{
				"name":         "Bob",
				"email":        "***REDACTED***",
				"phone_number": "***REDACTED***",
				"token":        "***REDACTED***",
			},
		},
		{
			name:  "does not mask non-sensitive fields",
			input: `{"name": "Alice", "age": 30, "city": "NYC"}`,
			expected: map[string]interface{}{
				"name": "Alice",
				"age":  float64(30),
				"city": "NYC",
			},
		},
		{
			name:  "masks sensitive fields in nested objects",
			input: `{"user": {"name": "Test", "email": "test@example.com"}, "id": 123}`,
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "Test",
					"email": "***REDACTED***",
				},
				"id": float64(123),
			},
		},
		{
			name:  "masks sensitive fields in arrays",
			input: `{"users": [{"name": "User1", "email": "user1@test.com"}, {"name": "User2", "email": "user2@test.com"}]}`,
			expected: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"name":  "User1",
						"email": "***REDACTED***",
					},
					map[string]interface{}{
						"name":  "User2",
						"email": "***REDACTED***",
					},
				},
			},
		},
		{
			name:     "handles invalid JSON gracefully",
			input:    `{invalid json}`,
			expected: nil, // Will return original data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData([]byte(tt.input))

			if tt.expected == nil {
				// For invalid JSON, result should be original input
				if string(result) != tt.input {
					t.Errorf("Expected original input for invalid JSON, got %s", string(result))
				}
				return
			}

			var resultMap map[string]interface{}
			if err := json.Unmarshal(result, &resultMap); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Compare the results
			if !compareMaps(resultMap, tt.expected) {
				t.Errorf("MaskSensitiveData() = %v, want %v", resultMap, tt.expected)
			}
		})
	}
}

func TestMaskJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "masks map with sensitive keys",
			input: map[string]interface{}{
				"name":  "Test",
				"email": "test@example.com",
			},
			expected: map[string]interface{}{
				"name":  "Test",
				"email": "***REDACTED***",
			},
		},
		{
			name: "handles array of maps",
			input: []interface{}{
				map[string]interface{}{"email": "test1@example.com"},
				map[string]interface{}{"email": "test2@example.com"},
			},
			expected: []interface{}{
				map[string]interface{}{"email": "***REDACTED***"},
				map[string]interface{}{"email": "***REDACTED***"},
			},
		},
		{
			name:     "handles primitive types",
			input:    "plain string",
			expected: "plain string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskJSON(tt.input)

			if !compareInterfaces(result, tt.expected) {
				t.Errorf("maskJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to compare maps
func compareMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valA := range a {
		valB, exists := b[key]
		if !exists {
			return false
		}

		if !compareInterfaces(valA, valB) {
			return false
		}
	}

	return true
}

// Helper function to compare interfaces
func compareInterfaces(a, b interface{}) bool {
	switch va := a.(type) {
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return compareMaps(va, vb)
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !compareInterfaces(va[i], vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
