package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// Publish publishes a message
func Publish(topicID, message string) error {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	// authenticate with pubsub using service account
	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credsPath))
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	topic := client.Topic(topicID)

	// Publish is asynchronous. It will never block
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(message),
	})

	_, err = result.Get(ctx) // blocks until success or error

	return err // return any error or nil if success.
}
