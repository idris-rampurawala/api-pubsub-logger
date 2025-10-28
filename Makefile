.PHONY: all build run clean test help

# Configuration
GOOGLE_CLOUD_PROJECT ?= demo-project
PUBSUB_TOPIC ?= api-log-events
PUBSUB_EMULATOR_HOST ?= localhost:8085

export GOOGLE_CLOUD_PROJECT
export PUBSUB_TOPIC
export PUBSUB_EMULATOR_HOST

# Build the API server
build:
	@go build -o dist/api-server cmd/api/main.go
	@go build -o dist/pubsub-cli cmd/pubsub/main.go

# Run the API server
run:
	@go run cmd/api/main.go

# Clean build artifacts
clean:
	@rm -rf dist

# Run tests
test:
	@go test -race -cover ./...

# Install dependencies
install:
	@go mod download
	@go mod tidy

# Local Pub/Sub emulator commands
local-pubsub:
	@echo "Starting Pub/Sub emulator on $(PUBSUB_EMULATOR_HOST)..."
	@gcloud beta emulators pubsub start \
		--project=$(GOOGLE_CLOUD_PROJECT) \
		--host-port=$(PUBSUB_EMULATOR_HOST)

local-pubsub-create-topic:
	@PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) go run cmd/pubsub/main.go create-topic \
		--project-id=$(GOOGLE_CLOUD_PROJECT) \
		--topic=$(PUBSUB_TOPIC)

local-pubsub-list-topics:
	@PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) go run cmd/pubsub/main.go list-topics \
		--project-id=$(GOOGLE_CLOUD_PROJECT)

local-pubsub-subscribe:
	@PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) go run cmd/pubsub/main.go subscribe-topic \
		--project-id=$(GOOGLE_CLOUD_PROJECT) \
		--topic=$(PUBSUB_TOPIC)

local-pubsub-delete-subscriptions:
	@PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) go run cmd/pubsub/main.go delete-all-subscriptions \
		--project-id=$(GOOGLE_CLOUD_PROJECT)

# Help command
help:
	@echo "Available commands:"
	@echo "  make build                           - Build the API server and CLI tools"
	@echo "  make run                             - Run the API server"
	@echo "  make clean                           - Clean build artifacts"
	@echo "  make test                            - Run tests"
	@echo "  make install                         - Install dependencies"
	@echo ""
	@echo "Pub/Sub Emulator commands:"
	@echo "  make local-pubsub                    - Start the Pub/Sub emulator"
	@echo "  make local-pubsub-create-topic       - Create the API log topic"
	@echo "  make local-pubsub-list-topics        - List all topics"
	@echo "  make local-pubsub-subscribe          - Subscribe and listen to the topic"
	@echo "  make local-pubsub-delete-subscriptions - Delete all subscriptions"
