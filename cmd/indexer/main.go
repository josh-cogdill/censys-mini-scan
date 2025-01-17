package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/indexing"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	IndexName string
	ProjectId string
	SubscriptionId string
}

func main() {	
	config, err := GetConfig()
	if err != nil {
		panic(err)
	}
	pubsubClient, err := pubsub.NewClient(context.Background(), config.ProjectId)
	if err != nil {
		panic(err)
	}
	defer pubsubClient.Close()
	consumer := indexing.NewConsumer(pubsubClient, config.SubscriptionId)

	defaultESClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	esClient := indexing.NewEsClient(defaultESClient)
	indexer := indexing.NewIndexer(esClient)

	err = indexer.CreateIndex(config.IndexName)
	if err != nil {
		panic(err)
	}

	// Channel to receive OS signals for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v\n", sig)
		cancel()
	}()

	// Channel to synchronize our message processing
	msgChan := make(chan []byte)

	log.Println("Starting Indexer.")
	// Start the indexer - reading off the msgChan
	go indexer.Process(msgChan, config.IndexName, ctx)

	log.Println("Starting Consumer.")
	// Start the consumer - writing to the msgChan
	// Consume blocks until ctx is done, or the service returns a non-retryable error
	consumer.Consume(ctx, msgChan)

	log.Println("Shutdown complete.")
}

func GetConfig() (*Config, error) {
	idxName := os.Getenv("INDEX_NAME")
	if idxName == "" {
		return nil, errors.New("Environment Variable INDEX_NAME needs to be set.")
	}
	projId := os.Getenv("PUBSUB_PROJECT_ID")
	if projId == "" {
		return nil, errors.New("Environment Variable PUBSUB_PROJECT_ID needs to be set.")
	}
	subId := os.Getenv("PUBSUB_SUBCRIPTION_ID")
	if subId == "" {
		return nil, errors.New("Environment Variable PUBSUB_SUBCRIPTION_ID needs to be set.")
	}
	return &Config{
		IndexName: idxName,
		ProjectId: projId,
		SubscriptionId: subId,
	}, nil
}

