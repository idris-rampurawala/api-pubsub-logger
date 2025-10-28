package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "api-pubsub-logger/internal/http"
	"api-pubsub-logger/internal/pubsub"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Addr               string `envconfig:"ADDR" default:":8080"`
	ServiceName        string `envconfig:"SERVICE_NAME" default:"api-pubsub-logger"`
	Version            string `envconfig:"VERSION" default:"1.0.0"`
	GoogleCloudProject string `envconfig:"GOOGLE_CLOUD_PROJECT" default:"demo-project"`
	PubSubTopic        string `envconfig:"PUBSUB_TOPIC" default:"api-log-events"`
}

func main() {
	// Load configuration from environment
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting %s v%s on %s", cfg.ServiceName, cfg.Version, cfg.Addr)

	// Initialize Pub/Sub client
	ctx := context.Background()
	pubsubClient, err := pubsub.New(ctx, pubsub.Options{
		ProjectID: cfg.GoogleCloudProject,
		TopicName: cfg.PubSubTopic,
	})
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubsubClient.Close()

	log.Printf("Connected to Pub/Sub project: %s, topic: %s", cfg.GoogleCloudProject, cfg.PubSubTopic)

	// Initialize HTTP handler
	handler := httphandler.New(pubsubClient, cfg.ServiceName, cfg.Version)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server is listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
