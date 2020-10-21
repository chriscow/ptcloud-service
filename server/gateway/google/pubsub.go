package google

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// Publish publishes a message
func Publish(topicID string, message []byte) error {

	ctx := context.Background()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	topic, err := client.createTopic(topicID)
	if err != nil {
		return err
	}

	// Publish is asynchronous. It will never block
	result := topic.Publish(ctx, &pubsub.Message{
		Data: message,
	})

	_, err = result.Get(ctx) // blocks until success or error

	return err // return any error or nil if success.
}

// SubscribeCallback ....
type SubscribeCallback func(ctx context.Context, msg *pubsub.Message) (bool, error)

// Subscribe ...
func Subscribe(topicID string, callback SubscribeCallback) error {
	ctx := context.Background()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	topic, err := client.createTopic(topicID)
	if err != nil {
		return err
	}

	sub, err := client.createSubscription("strucim", topic)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	var mu sync.Mutex
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Printf("[pubsub.go] Got message: %q\n", string(msg.Data))

		mu.Lock()
		defer mu.Unlock()

		ok, err := callback(ctx, msg)
		if err != nil || !ok {
			cancel()
		}
	})

	return err
}

type pubSubClient struct {
	psclient *pubsub.Client
}

// getClient creates a pubsub client
func getClient(ctx context.Context) (*pubSubClient, error) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credsPath))
	if err != nil {
		log.Printf("Error when creating pubsub client. Err: %v", err)
		return nil, err
	}
	return &pubSubClient{psclient: client}, nil
}

// topicExists checks if a given topic exists
func (client *pubSubClient) topicExists(topicName string) (bool, error) {
	topic := client.psclient.Topic(topicName)
	return topic.Exists(context.Background())
}

// createTopic creates a topic if a topic name does not exist or returns one
// if it is already present
func (client *pubSubClient) createTopic(topicName string) (*pubsub.Topic, error) {
	topicExists, err := client.topicExists(topicName)
	if err != nil {
		log.Printf("Could not check if topic exists. Error: %+v", err)
		return nil, err
	}
	var topic *pubsub.Topic

	if !topicExists {
		topic, err = client.psclient.CreateTopic(context.Background(), topicName)
		if err != nil {
			log.Printf("Could not create topic. Err: %+v", err)
			return nil, err
		}
	} else {
		topic = client.psclient.Topic(topicName)
	}

	return topic, nil
}

// deleteTopic Deletes a topic
func (client *pubSubClient) deleteTopic(topicName string) error {
	return client.psclient.Topic(topicName).Delete(context.Background())
}

// createSubscription creates the subscription to a topic
func (client *pubSubClient) createSubscription(subscriptionName string, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	subscription := client.psclient.Subscription(subscriptionName)

	subscriptionExists, err := subscription.Exists(context.Background())
	if err != nil {
		log.Printf("Could not check if subscription %s exists. Err: %v", subscriptionName, err)
		return nil, err
	}

	if !subscriptionExists {

		cfg := pubsub.SubscriptionConfig{
			Topic: topic,
			// The subscriber has a configurable, limited amount of time -- known as the ackDeadline -- to acknowledge
			// the outstanding message. Once the deadline passes, the message is no longer considered outstanding, and
			// Cloud Pub/Sub will attempt to redeliver the message.
			AckDeadline: 60 * time.Second,
		}

		subscription, err = client.psclient.CreateSubscription(context.Background(), subscriptionName, cfg)
		if err != nil {
			log.Printf("Could not create subscription %s. Err: %v", subscriptionName, err)
			return nil, err
		}
		subscription.ReceiveSettings = pubsub.ReceiveSettings{
			// This is the maximum amount of messages that are allowed to be processed by the callback function at a time.
			// Once this limit is reached, the client waits for messages to be acked or nacked by the callback before
			// requesting more messages from the server.
			MaxOutstandingMessages: 100,
			// This is the maximum amount of time that the client will extend a message's deadline. This value should be
			// set as high as messages are expected to be processed, plus some buffer.
			MaxExtension: 10 * time.Second,
		}
	}
	return subscription, nil
}
