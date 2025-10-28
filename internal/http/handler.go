package http

import (
	"net/http"

	"api-pubsub-logger/internal/pubsub"

	"github.com/gorilla/mux"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	PubSubClient pubsub.Publisher
	ServiceName  string
	Version      string
	router       *mux.Router
}

// New creates a new HTTP handler with dependencies
func New(pubsubClient pubsub.Publisher, serviceName, version string) *Handler {
	return &Handler{
		PubSubClient: pubsubClient,
		ServiceName:  serviceName,
		Version:      version,
	}
}

// GetURLParam calls gorilla's mux.Vars function to extract URL parameters
func GetURLParam(r *http.Request, key string) string {
	return mux.Vars(r)[key]
}
