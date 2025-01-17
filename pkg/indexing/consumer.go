package indexing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"

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
	// Receive blocks until ctx is done, or the service returns a non-retryable error.
	// The standard way to terminate a Receive is to cancel its context
	err := c.Sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		// Because Receive calls this function concurrently from multiple goroutines, I'll take care of processing here
		esData, err := GetESData(m)
		if err != nil {
			log.Printf("Failed to parse message: %v, Error: %v\n", m, err)
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

type ESData struct {
	Ip          string      `json:"ip"`
	Port        uint32      `json:"port"`
	Service     string      `json:"service"`
	Timestamp   int64       `json:"timestamp"`
	Response	string		`json:"response"`
}

func GetESData(m *pubsub.Message) ([]byte, error) {
	var esData ESData

	// Unmarshal handles the different data versions.
	err := json.Unmarshal(m.Data, &esData)
	if err != nil {
		return nil, err
	}

	logger.Log("Processing message: %v\n", esData)

	// Re-serialize for DB insert
	serialized, err := json.Marshal(esData) 
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

// Custom unmarshal to handle data versions V1 and V2
func (esData *ESData) UnmarshalJSON(data []byte) error {
	var scan scanning.Scan
	err := json.Unmarshal(data, &scan)
	if err != nil {
		return err
	}

	esData.Ip, esData.Port, esData.Service, esData.Timestamp = scan.Ip, scan.Port, scan.Service, scan.Timestamp

	if scan.DataVersion == scanning.V1 {
		dataMap, ok := scan.Data.(map[string]interface{})
		if !ok {
			return errors.New("Failed to parse scan data")
		}
		respBytes, ok := dataMap["response_bytes_utf8"]
		if !ok {
			return errors.New("Failed to find response_bytes_utf8")
		}

		str, ok := respBytes.(string)
		if !ok {
			return errors.New("Failed to parse response_bytes_utf8")
		}

		decodedBytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}
		esData.Response = string(decodedBytes)
	} else {
		str := scan.Data.(map[string]interface{})["response_str"].(string)
		esData.Response = str
	}
	return nil
}