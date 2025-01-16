package indexing

import (
	"context"
	"errors"
	"fmt"

	"github.com/censys/scan-takehome/pkg/logger"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ClientInterface interface {
	IndicesExists(indexName []string) (*esapi.Response, error)
	IndicesCreate(indexName string) (*esapi.Response, error)
	Insert(indexName string, data []byte) (*esapi.Response, error)
}

type ESIndexer struct {
	Client ClientInterface
}

func NewIndexer(client ClientInterface) *ESIndexer {
	return &ESIndexer{Client: client}
}

func (indexer *ESIndexer) CreateIndex(idxName string) error {
	existsResp, err := indexer.Client.IndicesExists([]string{idxName})
	if err != nil {
		return err
	}
	if existsResp != nil && existsResp.Body != nil {
		defer existsResp.Body.Close()
	}

	// Index does not exist - create it
	if existsResp.StatusCode == 404 {
		createResp, err := indexer.Client.IndicesCreate(idxName)
		if createResp != nil && createResp.Body != nil {
			defer createResp.Body.Close()
		}
		if err != nil {
			return err
		}
		if createResp.StatusCode != 200 {
			errMsg := fmt.Sprintf("Failed to create index: %s, error: %v\n", idxName, createResp.String())
			return errors.New(errMsg)
		}
	}
	return nil
}

func (indexer *ESIndexer) ExecuteInsert(data []byte, idxName string) error {
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		_, err := indexer.Client.Insert(idxName, data)
		if err == nil{
			return nil
		}
	}
	
	errMsg := fmt.Sprintf("Failed to insert into %s after 5 attemps, Error: %v\n", idxName, err)
	return errors.New(errMsg)
}

func (indexer *ESIndexer) Process(msgChan chan []byte, idxName string, ctx context.Context) {
	for {
		select {
		case <- ctx.Done():
			logger.Log("Context canceled: %v\n", ctx.Err())
			return
		
		case esData := <-msgChan:
			err := indexer.ExecuteInsert(esData, idxName)
			if err != nil {
				logger.Log("Failed to insert data: %v\n Error: %v\n", esData, err)
			}
		}
	}
}
