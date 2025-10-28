package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

// Utility program to interact with a local Pub/Sub emulator

var subcommands = []string{
	"create-topic",
	"list-topics",
	"subscribe-topic",
	"delete-all-subscriptions",
}

func main() {
	createTopicCmd := flag.NewFlagSet("create-topic", flag.ExitOnError)
	createTopicProjectID := createTopicCmd.String("project-id", "", "Pub/Sub project ID")
	createTopicName := createTopicCmd.String("topic", "", "Topic to be created")

	listTopicsCmd := flag.NewFlagSet("list-topics", flag.ExitOnError)
	listTopicsProjectID := listTopicsCmd.String("project-id", "", "Pub/Sub project ID")

	subscribeTopicCmd := flag.NewFlagSet("subscribe-topic", flag.ExitOnError)
	subscribeTopicProjectID := subscribeTopicCmd.String("project-id", "", "Pub/Sub project ID")
	subscribeTopicName := subscribeTopicCmd.String("topic", "", "Topic to be subscribed")

	deleteSubscriptionsCmd := flag.NewFlagSet("delete-all-subscriptions", flag.ExitOnError)
	deleteSubscriptionsProjectID := deleteSubscriptionsCmd.String("project-id", "", "Pub/Sub project ID")

	if len(os.Args) < 2 {
		fmt.Println("Subcommand is required")
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-topic":
		if err := createTopicCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		createTopic(*createTopicProjectID, *createTopicName)
	case "list-topics":
		if err := listTopicsCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		listTopics(*listTopicsProjectID)
	case "delete-all-subscriptions":
		if err := deleteSubscriptionsCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		deleteAllSubscriptions(*deleteSubscriptionsProjectID)
	case "subscribe-topic":
		if err := subscribeTopicCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		subscribeTopic(*subscribeTopicProjectID, *subscribeTopicName)
	default:
		fmt.Printf("Unknown subcommand %s\n", os.Args[1])
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Available subcommands:")
	for _, sc := range subcommands {
		fmt.Println(" - " + sc)
	}
}

func newClient(project string) (*pubsub.Client, error) {
	if project == "" {
		return nil, errors.New("project is empty")
	}

	client, err := pubsub.NewClient(context.Background(), project)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createTopic(project, topic string) {
	if topic == "" {
		panic("topic is empty")
	}

	client, err := newClient(project)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Creating topic '%s' for project '%s'...\n", topic, project)

	t, err := client.CreateTopic(context.Background(), topic)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Topic '%s' already exists\n", topic)
			return
		}
		panic(err)
	}

	fmt.Println("Topic created:")
	fmt.Println(t)
}

func listTopics(project string) {
	client, err := newClient(project)
	if err != nil {
		panic(err)
	}

	fmt.Println("Topics:")
	topics := client.Topics(context.Background())
	for {
		topic, err := topics.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(" * " + topic.String())
	}
}

func subscribeTopic(project, topic string) {
	subscriptionName := fmt.Sprintf("%s_sub", topic)

	client, err := newClient(project)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	sub := client.Subscription(subscriptionName)
	subExists, err := sub.Exists(ctx)
	if err != nil {
		panic(err)
	}
	if subExists {
		fmt.Printf("Subscription '%s' already exists\n", subscriptionName)
	} else {
		fmt.Printf("Creating subscription '%s' for project '%s' and topic '%s'...\n", subscriptionName, project, topic)

		subConfig := pubsub.SubscriptionConfig{
			Topic:            client.Topic(topic),
			AckDeadline:      10 * time.Second,
			ExpirationPolicy: time.Duration(0),
		}

		sub, err = client.CreateSubscription(ctx, subscriptionName, subConfig)
		if err != nil {
			fmt.Printf("Failed to create subscription: %v\n", err)
			return
		}
		fmt.Printf("Subscription created: %s\n", sub.String())
	}

	fmt.Printf("Listening to topic '%s'...\n", topic)
	if err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		msg.Ack()
		fmt.Println("---")
		fmt.Println("Received message:")
		fmt.Println(string(msg.Data))
		fmt.Println("---")
	}); err != nil {
		panic(err)
	}
}

func deleteAllSubscriptions(project string) {
	client, err := newClient(project)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	subIter := client.Subscriptions(ctx)

	fmt.Printf("Deleting all subscriptions for project '%s'...\n", project)

	for {
		sub, err := subIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}

		subName := sub.String()
		fmt.Printf("Deleting subscription: %s...\n", subName)

		if err := sub.Delete(ctx); err != nil {
			fmt.Printf("Failed to delete subscription '%s': %v\n", subName, err)
		} else {
			fmt.Printf("Deleted subscription: %s\n", subName)
		}
	}
}
