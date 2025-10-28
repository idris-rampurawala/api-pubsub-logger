package utils

import "encoding/json"

// Sensitive keys that should be masked in logs
var sensitiveKeys = map[string]struct{}{
	"email":        {},
	"phone_number": {},
	"password":     {},
	"token":        {},
	"api_key":      {},
}

// MaskSensitiveData recursively masks sensitive data in JSON objects
func MaskSensitiveData(data []byte) []byte {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return data
	}

	maskedData := maskJSON(jsonData)

	result, err := json.Marshal(maskedData)
	if err != nil {
		return data
	}
	return result
}

// maskJSON handles objects (maps) and arrays recursively to mask sensitive data
func maskJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}: // Handle JSON objects
		for key, val := range v {
			if _, exists := sensitiveKeys[key]; exists {
				v[key] = "***REDACTED***"
			} else {
				v[key] = maskJSON(val)
			}
		}
	case []interface{}: // Handle JSON arrays
		for i, val := range v {
			v[i] = maskJSON(val)
		}
	}
	return data
}
