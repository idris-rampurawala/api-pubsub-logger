package pubsub

import (
	"context"
	"encoding/json"
	"log"

	"api-pubsub-logger/pkg/logger"

	"cloud.google.com/go/pubsub"
)

// Client is the Pub/Sub client wrapper
type Client struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// Options contains configuration options for the Pub/Sub client
type Options struct {
	ProjectID string
	TopicName string
}

// New creates a new Pub/Sub client
func New(ctx context.Context, opts Options) (*Client, error) {
	client, err := pubsub.NewClient(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(opts.TopicName)

	return &Client{
		client: client,
		topic:  topic,
	}, nil
}

// PublishAPILogEvent publishes an API log event to Pub/Sub
func (c *Client) PublishAPILogEvent(ctx context.Context, event logger.APILogEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling API log event: %v", err)
		return err
	}

	result := c.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	// Get the server-generated message ID
	_, err = result.Get(ctx)
	if err != nil {
		log.Printf("Error publishing API log event: %v", err)
		return err
	}

	return nil
}

// Close closes the Pub/Sub client
func (c *Client) Close() error {
	c.topic.Stop()
	return c.client.Close()
}
