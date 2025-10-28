package pubsub

import (
	"context"

	"api-pubsub-logger/pkg/logger"
)

// Publisher defines the interface for publishing API log events
type Publisher interface {
	PublishAPILogEvent(ctx context.Context, event logger.APILogEvent) error
	Close() error
}
