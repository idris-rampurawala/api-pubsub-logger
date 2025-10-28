package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"api-pubsub-logger/internal/pubsub"
	"api-pubsub-logger/internal/utils"
	"api-pubsub-logger/pkg/logger"

	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
)

// Routes to skip from logging (e.g., health checks)
var skipRoutes = map[string]struct{}{
	"GET::/health": {},
}

// responseRecorder is a wrapper for http.ResponseWriter to capture response data
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware logs HTTP requests and responses to Pub/Sub
func LoggingMiddleware(pubsubClient pubsub.Publisher, serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request URL path should be skipped
			routeKey := fmt.Sprintf("%s::%s", strings.ToUpper(r.Method), r.URL.Path)
			if _, skip := skipRoutes[routeKey]; skip {
				next.ServeHTTP(w, r)
				return
			}

			startTime := time.Now()

			// Read request body
			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // Restore the request body
			}

			// Create response recorder to capture response
			recorder := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(recorder, r)

			// Extract context values
			ctx := r.Context()
			requestID := utils.GetRequestID(ctx)
			userID := utils.GetUserID(ctx)

			// Mask sensitive data in request and response bodies
			maskedRequestBody := string(utils.MaskSensitiveData(requestBody))
			maskedResponseBody := string(utils.MaskSensitiveData(recorder.body.Bytes()))

			// Extract route version and name from mux router
			routeName, routeVersion := extractRouteVersionAndName(mux.CurrentRoute(r))

			// Create API log event
			logData := logger.APILogEvent{
				RequestID:    null.NewString(requestID, len(requestID) > 0),
				Service:      serviceName,
				Method:       r.Method,
				URL:          r.URL.String(),
				RequestBody:  null.NewString(maskedRequestBody, len(maskedRequestBody) > 0),
				ResponseBody: null.NewString(maskedResponseBody, len(maskedResponseBody) > 0),
				ResponseCode: recorder.statusCode,
				UserID:       null.NewString(userID, len(userID) > 0),
				Version:      routeVersion,
				Name:         routeName,
				CreatedAt:    startTime,
				Duration:     time.Since(startTime).Seconds(),
			}

			// Publish to Pub/Sub asynchronously using background context
			// We use context.Background() instead of the request context because
			// the request context gets canceled when the HTTP response is sent,
			// but we want the publishing to complete independently
			go sendToPubSub(context.Background(), pubsubClient, logData)
		})
	}
}

// extractRouteVersionAndName extracts the name and version from the mux route
func extractRouteVersionAndName(route *mux.Route) (string, string) {
	var name, version string
	if route != nil {
		name = route.GetName()
		pathTemplate, _ := route.GetPathTemplate()
		parts := strings.Split(strings.Trim(pathTemplate, "/"), "/")
		if len(parts) > 0 && strings.HasPrefix(parts[0], "v") {
			version = parts[0]
		}
	}
	return name, version
}

// sendToPubSub publishes log data to Pub/Sub
func sendToPubSub(ctx context.Context, client pubsub.Publisher, logData logger.APILogEvent) {
	if err := client.PublishAPILogEvent(ctx, logData); err != nil {
		log.Printf("Failed to publish API log event: %v", err)
	}
}
