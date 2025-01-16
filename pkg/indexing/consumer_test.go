package indexing

import (
	"encoding/json"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/scanning"
	"github.com/stretchr/testify/assert"
)

func TestGetESData(t *testing.T) {
	var (
		ip string = "1.1.1.1"
		port uint32 = 1
		service string = "HTTP"
		timestamp int64 = time.Now().Unix()
		responsestr string = "service response: 1"
	)
	
	t.Run("Valid Scan Data Version V1", func(t *testing.T) {
		expectedDeserialized := &scanning.ServiceData{
			Ip: ip,
			Port: port,
			Service: service,
			Timestamp: timestamp,
			Response: responsestr,
		}
		expectedSerialized, err := json.Marshal(expectedDeserialized)
		assert.NoError(t, err)

		mockScanV1 := createMockScan(ip, port, service, timestamp, scanning.V1, responsestr)
		pubsubData, err := json.Marshal(mockScanV1)
		assert.NoError(t, err)

		pubsubMsg := &pubsub.Message{
			Data: pubsubData,
		}

		esData, err := GetESData(pubsubMsg)
		assert.NoError(t, err)
		assert.Equal(t, expectedSerialized, esData)
	})

	t.Run("Valid Scan Data Version V2", func(t *testing.T) {
		expectedDeserialized := &scanning.ServiceData{
			Ip: ip,
			Port: port,
			Service: service,
			Timestamp: timestamp,
			Response: responsestr,
		}
		expectedSerialized, err := json.Marshal(expectedDeserialized)
		assert.NoError(t, err)

		mockScanV2 := createMockScan(ip, port, service, timestamp, scanning.V2, responsestr)
		pubsubData, err := json.Marshal(mockScanV2)
		assert.NoError(t, err)

		pubsubMsg := &pubsub.Message{
			Data: pubsubData,
		}

		esData, err := GetESData(pubsubMsg)
		assert.NoError(t, err)
		assert.Equal(t, expectedSerialized, esData)
	})

	t.Run("Invalid json returns error", func(t *testing.T) {
		msg := &pubsub.Message{
			Data: []byte("invalid json"),
		}

		esData, err := GetESData(msg)
		assert.Error(t, err)
		assert.Nil(t, esData)
	})
}

// helpers
func createMockScan(ip string, port uint32, service string, timestamp int64, dataversion int, responsestr string) *scanning.Scan {
	mockScan := &scanning.Scan{
		Ip: ip,
		Port: port,
		Service: service,
		Timestamp: timestamp,
		DataVersion: dataversion,
	}
	if dataversion == scanning.V1 {
		mockScan.Data = &scanning.V1Data{ResponseBytesUtf8: []byte(responsestr)}
	} else {
		mockScan.Data = &scanning.V2Data{ResponseStr: responsestr}
	}
	return mockScan
}