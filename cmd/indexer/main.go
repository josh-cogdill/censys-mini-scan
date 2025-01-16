package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/indexing"
	"github.com/censys/scan-takehome/pkg/logger"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

var (
	PROJECT_ID string
	SUBSCRIPTION_ID string
	INDEX_NAME string
)

func init() {
	INDEX_NAME = os.Getenv("INDEX_NAME")
	PROJECT_ID = os.Getenv("PUBSUB_PROJECT_ID")
	SUBSCRIPTION_ID = os.Getenv("PUBSUB_SUBCRIPTION_ID")
}

func main() {	
	pubsubClient, err := pubsub.NewClient(context.Background(), PROJECT_ID)
	if err != nil {
		panic(err)
	}
	defer pubsubClient.Close()
	consumer := indexing.NewConsumer(pubsubClient, SUBSCRIPTION_ID)

	defaultESClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	esClient := indexing.NewEsClient(defaultESClient)
	indexer := indexing.NewIndexer(esClient)

	err = indexer.CreateIndex(INDEX_NAME)
	if err != nil {
		panic(err)
	}

	// Channel to receive OS signals for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigChan
		logger.Log("Received signal: %v\n", sig)
		cancel()
	}()

	// Channel to synchronize our message processing
	msgChan := make(chan []byte)

	logger.Log("Starting Indexer...")
	// Start the indexer - reading off the msgChan
	go indexer.Process(msgChan, INDEX_NAME, ctx)

	logger.Log("Starting Consumer...")
	// Start the consumer - writing to the msgChan
	consumer.Consume(ctx, msgChan)

	logger.Log("Shutdown complete.")
}

