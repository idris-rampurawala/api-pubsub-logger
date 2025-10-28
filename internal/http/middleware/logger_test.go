package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"api-pubsub-logger/internal/utils"
	"api-pubsub-logger/pkg/logger"

	"github.com/gorilla/mux"
)

// mockPubSubClient is a mock implementation of the pubsub client for testing
type mockPubSubClient struct {
	mu              sync.Mutex
	publishedEvents []logger.APILogEvent
	publishError    error
}

func (m *mockPubSubClient) PublishAPILogEvent(ctx context.Context, event logger.APILogEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedEvents = append(m.publishedEvents, event)
	return m.publishError
}

func (m *mockPubSubClient) Close() error {
	return nil
}

func (m *mockPubSubClient) getEvents() []logger.APILogEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to avoid race conditions
	events := make([]logger.APILogEvent, len(m.publishedEvents))
	copy(events, m.publishedEvents)
	return events
}

func TestExtractRouteVersionAndName(t *testing.T) {
	tests := []struct {
		name            string
		pathTemplate    string
		routeName       string
		expectedVersion string
		expectedName    string
	}{
		{
			name:            "extracts v1 version",
			pathTemplate:    "/v1/items",
			routeName:       "list_items",
			expectedVersion: "v1",
			expectedName:    "list_items",
		},
		{
			name:            "extracts v2 version",
			pathTemplate:    "/v2/users/{id}",
			routeName:       "get_user",
			expectedVersion: "v2",
			expectedName:    "get_user",
		},
		{
			name:            "handles no version",
			pathTemplate:    "/items",
			routeName:       "items",
			expectedVersion: "",
			expectedName:    "items",
		},
		{
			name:            "handles nil route",
			pathTemplate:    "",
			routeName:       "",
			expectedVersion: "",
			expectedName:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mux.NewRouter()
			if tt.pathTemplate != "" {
				r.Path(tt.pathTemplate).Name(tt.routeName)
			}

			var route *mux.Route
			if tt.pathTemplate != "" {
				req := httptest.NewRequest("GET", tt.pathTemplate, nil)
				var match mux.RouteMatch
				if r.Match(req, &match) {
					route = match.Route
				}
			}

			name, version := extractRouteVersionAndName(route)

			if name != tt.expectedName {
				t.Errorf("Expected name = %v, got %v", tt.expectedName, name)
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version = %v, got %v", tt.expectedVersion, version)
			}
		})
	}
}

func TestLoggingMiddleware_SkipsHealthCheck(t *testing.T) {
	mockClient := &mockPubSubClient{}
	serviceName := "test-service"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler := LoggingMiddleware(mockClient, serviceName)(testHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Give some time for async publishing
	time.Sleep(50 * time.Millisecond)

	// Verify no events were published
	events := mockClient.getEvents()
	if len(events) != 0 {
		t.Errorf("Expected no events for /health, got %d", len(events))
	}
}

func TestLoggingMiddleware_CapturesRequestResponse(t *testing.T) {
	mockClient := &mockPubSubClient{}
	serviceName := "test-service"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	})

	handler := LoggingMiddleware(mockClient, serviceName)(testHandler)

	requestBody := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/v1/items", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Set request ID and user ID in context
	ctx := utils.SetRequestID(req.Context(), "req-123")
	ctx = utils.SetUserID(ctx, "user-456")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Give some time for async publishing
	time.Sleep(100 * time.Millisecond)

	// Verify event was published
	events := mockClient.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]

	// Verify event fields
	if event.Service != serviceName {
		t.Errorf("Expected service = %v, got %v", serviceName, event.Service)
	}

	if event.Method != "POST" {
		t.Errorf("Expected method = POST, got %v", event.Method)
	}

	if event.URL != "/v1/items" {
		t.Errorf("Expected URL = /v1/items, got %v", event.URL)
	}

	if event.ResponseCode != http.StatusOK {
		t.Errorf("Expected response code = 200, got %v", event.ResponseCode)
	}

	if !event.RequestID.Valid || event.RequestID.String != "req-123" {
		t.Errorf("Expected request ID = req-123, got %v", event.RequestID)
	}

	if !event.UserID.Valid || event.UserID.String != "user-456" {
		t.Errorf("Expected user ID = user-456, got %v", event.UserID)
	}

	if event.Duration <= 0 {
		t.Error("Expected duration > 0")
	}
}

func TestLoggingMiddleware_CapturesVersionAndName(t *testing.T) {
	mockClient := &mockPubSubClient{}
	serviceName := "test-service"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	})

	// Create a real mux router to test route extraction
	r := mux.NewRouter()
	r.Use(LoggingMiddleware(mockClient, serviceName))

	// Add a versioned route with a name
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.Methods("GET").Path("/items").Name("list_items").Handler(testHandler)

	req := httptest.NewRequest("GET", "/v1/items", nil)
	ctx := utils.SetRequestID(req.Context(), "req-789")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Give some time for async publishing
	time.Sleep(100 * time.Millisecond)

	// Verify event was published
	events := mockClient.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]

	// Verify version is extracted
	if event.Version != "v1" {
		t.Errorf("Expected version = v1, got %v", event.Version)
	}

	// Verify name is extracted
	if event.Name != "list_items" {
		t.Errorf("Expected name = list_items, got %v", event.Name)
	}
}

func TestLoggingMiddleware_MasksSensitiveData(t *testing.T) {
	mockClient := &mockPubSubClient{}
	serviceName := "test-service"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123","email":"user@example.com","phone_number":"+1-555-1234"}`))
	})

	handler := LoggingMiddleware(mockClient, serviceName)(testHandler)

	requestBody := `{"email":"test@example.com","password":"secret123"}`
	req := httptest.NewRequest("POST", "/v1/items", bytes.NewBufferString(requestBody))
	ctx := utils.SetRequestID(req.Context(), "req-789")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Give some time for async publishing
	time.Sleep(100 * time.Millisecond)

	// Verify event was published
	events := mockClient.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]

	// Verify sensitive data is masked in request
	if event.RequestBody.Valid {
		reqBody := event.RequestBody.String
		if !bytes.Contains([]byte(reqBody), []byte("***REDACTED***")) {
			t.Error("Expected sensitive data to be masked in request body")
		}
	}

	// Verify sensitive data is masked in response
	if event.ResponseBody.Valid {
		respBody := event.ResponseBody.String
		if !bytes.Contains([]byte(respBody), []byte("***REDACTED***")) {
			t.Error("Expected sensitive data to be masked in response body")
		}
	}
}

func TestResponseRecorder(t *testing.T) {
	// Test that response recorder properly captures response
	rr := &responseRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}

	testData := []byte("test response body")
	n, err := rr.Write(testData)

	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}

	if rr.body.String() != string(testData) {
		t.Errorf("Body = %v, want %v", rr.body.String(), string(testData))
	}
}

func TestResponseRecorder_WriteHeader(t *testing.T) {
	rr := &responseRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}

	rr.WriteHeader(http.StatusCreated)

	if rr.statusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", rr.statusCode, http.StatusCreated)
	}
}
