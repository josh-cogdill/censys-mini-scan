package indexing

import (
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/logger"
	"github.com/censys/scan-takehome/pkg/scanning"
)


type Consumer struct {
	Client *pubsub.Client
	Sub pubsub.Subscription
}

func NewConsumer(client *pubsub.Client, subscriptionId string) *Consumer {
	return &Consumer{
		Client: client, 
		Sub: *client.Subscription(subscriptionId),
	}
}

func (c *Consumer) Consume(ctx context.Context, msgChan chan []byte) {
	err := c.Sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		// By default, Receive uses multiple goroutines, so we'll take care of the processing here
		esData, err := GetESData(m)
		if err != nil {
			logger.Log("Failed to parse message: %v, Error: %v\n", m, err)
			// Call Nack for faster re-delivery of the message
			m.Nack()
		} else {
			// Fan in our processed data
			msgChan <- esData
			m.Ack()
		}
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}

func GetESData(m *pubsub.Message) ([]byte, error) {
	var serviceData scanning.ServiceData

	// Unmarshal takes care of our different data versions. See pkg/scanning/types.go
	err := json.Unmarshal(m.Data, &serviceData)
	if err != nil {
		return nil, err
	}

	logger.Log("Processing message: %v\n", serviceData)

	// Re-serialize for DB insert
	serialized, err := json.Marshal(serviceData) 
	if err != nil {
		return nil, err
	}
	return serialized, nil
}