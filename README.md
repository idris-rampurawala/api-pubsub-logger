# API Pub/Sub Logger

A demonstration project showcasing how to log API requests and responses to Google Cloud Pub/Sub for later analysis and debugging.

## Blog Post

For a detailed explanation of this implementation, read the full blog post:
**[Coming Soon - Link will be added here]**

## Overview

This project demonstrates a pattern for logging all API requests and responses to Google Cloud Pub/Sub. The middleware captures:
- Request/Response bodies (with sensitive data masking)
- HTTP method and URL
- Response status codes
- Request duration
- Request IDs for tracing
- User IDs from headers

All logs are published to a Pub/Sub topic which can then be consumed by subscribers to store in BigQuery or other analytics platforms.

## Features

- **Middleware-based logging**: Automatic logging of all API requests
- **Sensitive data masking**: Automatically redacts email, phone numbers, and other sensitive fields
- **Request ID tracking**: Unique ID for each request for distributed tracing
- **API versioning**: Extracts and logs API version (v1, v2) and route names
- **Selective route skipping**: Skip logging for health checks and other endpoints
- **Local development**: Uses Pub/Sub emulator for local testing
- **Test-driven development**: Comprehensive unit tests for middleware and utilities

## Prerequisites

- Go 1.21 or higher
- Google Cloud SDK (for Pub/Sub emulator)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd api-pubsub-logger
```

2. Install dependencies:
```bash
make install
```

3. Copy the example environment file:
```bash
cp .env.example .env
```

## Running Locally

### Step 1: Start the Pub/Sub Emulator

In terminal 1:
```bash
make local-pubsub
```

This starts the Pub/Sub emulator on `localhost:8085`.

### Step 2: Create the Topic

In terminal 2:
```bash
make local-pubsub-create-topic
```

### Step 3: Start the API Server

```bash
export PUBSUB_EMULATOR_HOST=localhost:8085
make run
```

The API server will start on `http://localhost:8080`.

### Step 4: Subscribe to View Logs (Optional)

In terminal 3:
```bash
make local-pubsub-subscribe
```

This will display all API log events as they are published.

## Testing the API

### Get all items
```bash
curl http://localhost:8080/v1/items
```

### Create a new item
```bash
curl -X POST http://localhost:8080/v1/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-123" \
  -d '{
    "name": "Test Item",
    "description": "A test item",
    "email": "test@example.com",
    "phone_number": "+1-555-1234"
  }'
```

### Health check (not logged)
```bash
curl http://localhost:8080/health
```

### Run example script
```bash
./examples.sh
```

## Running Tests

The project includes comprehensive unit tests for middleware and utility functions:

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...
```

## Project Structure

```
.
├── cmd/
│   ├── api/
│   │   └── main.go                    # API server entry point
│   └── pubsub/
│       └── main.go                    # Pub/Sub CLI utility for managing topics
│
├── pkg/
│   └── logger/
│       ├── api_log.go                 # APILogEvent model
│       └── item.go                    # Item model
│
├── internal/
│   ├── http/
│   │   ├── handler.go                 # Handler struct with dependencies
│   │   ├── router.go                  # Route definitions with versioning
│   │   ├── handlers/
│   │   │   ├── health.go              # Health check handler
│   │   │   └── items.go               # Items CRUD handlers
│   │   └── middleware/
│   │       ├── logger.go              # API logging middleware
│   │       ├── logger_test.go         # Logging middleware tests
│   │       ├── requestid.go           # Request ID middleware
│   │       ├── requestid_test.go      # Request ID middleware tests
│   │       ├── userid.go              # User ID middleware
│   │       └── userid_test.go         # User ID middleware tests
│   │
│   ├── pubsub/
│   │   ├── client.go                  # Pub/Sub client implementation
│   │   └── interface.go               # Publisher interface
│   │
│   └── utils/
│       ├── context.go                 # Context helpers (user ID)
│       ├── context_test.go            # Context helpers tests
│       ├── mask.go                    # Sensitive data masking
│       ├── mask_test.go               # Masking tests
│       ├── requestid.go               # Request ID generation
│       └── requestid_test.go          # Request ID tests
│
├── .env.example                       # Example environment configuration
├── .gitignore                         # Git ignore rules
├── examples.sh                        # Example API requests script
├── go.mod                             # Go module definition
├── go.sum                             # Go dependencies checksums
├── LICENSE                            # MIT License
├── Makefile                           # Build and development commands
└── README.md                          # This file
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ADDR` | Server listen address | `:8080` |
| `SERVICE_NAME` | Service name in logs | `api-pubsub-logger` |
| `GOOGLE_CLOUD_PROJECT` | GCP project ID | `demo-project` |
| `PUBSUB_TOPIC` | Pub/Sub topic name | `api-log-events` |
| `PUBSUB_EMULATOR_HOST` | Pub/Sub emulator address | `localhost:8085` |

## Available Make Commands

- `make build` - Build the API server and CLI tools
- `make run` - Run the API server
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make install` - Install dependencies
- `make local-pubsub` - Start the Pub/Sub emulator
- `make local-pubsub-create-topic` - Create the API log topic
- `make local-pubsub-list-topics` - List all topics
- `make local-pubsub-subscribe` - Subscribe and listen to the topic
- `make local-pubsub-delete-subscriptions` - Delete all subscriptions

## How It Works

1. **Request arrives**: The API receives an HTTP request
2. **Middleware chain**: Request passes through middleware:
   - `RequestIDMiddleware`: Adds unique request ID
   - `UserIDMiddleware`: Extracts user ID from headers
   - `LoggingMiddleware`: Captures request/response data
3. **Handler executes**: Business logic processes the request
4. **Response captured**: Middleware captures the response
5. **Route extraction**: Extracts API version (v1, v2) and route name from the URL
6. **Data masking**: Sensitive fields are redacted
7. **Pub/Sub publish**: Log event is published asynchronously with all metadata
8. **Response sent**: Original response sent to client

## License

MIT

