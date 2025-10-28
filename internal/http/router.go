package http

import (
	"net/http"

	"api-pubsub-logger/internal/http/handlers"
	"api-pubsub-logger/internal/http/middleware"

	"github.com/gorilla/mux"
)

// Handler mounts all the handlers at the appropriate routes and adds any required middleware
func (h *Handler) Handler() http.Handler {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.UserIDMiddleware)
	r.Use(middleware.LoggingMiddleware(h.PubSubClient, h.ServiceName))

	// Health check endpoint (not logged due to skip in middleware)
	r.Methods("GET").Path("/health").Name("health").HandlerFunc(handlers.HealthCheck)

	// Version 1 API routes
	v1 := r.PathPrefix("/v1").Subrouter()

	// Items routes
	v1.Methods("GET").Path("/items").Name("list_items").HandlerFunc(handlers.GetItems)
	v1.Methods("POST").Path("/items").Name("create_item").HandlerFunc(handlers.CreateItem)

	h.router = r
	return r
}
