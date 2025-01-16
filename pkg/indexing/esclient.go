package indexing

import (
	"bytes"
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ESClient struct {
	Client *elasticsearch.Client
}

func NewEsClient(client *elasticsearch.Client) *ESClient {
	return &ESClient{Client: client}
}

func (c *ESClient) IndicesExists(indices []string) (*esapi.Response, error) {
	req := esapi.IndicesExistsRequest{Index: indices}
	return req.Do(context.Background(), c.Client)
}

func (c *ESClient) IndicesCreate(idxName string) (*esapi.Response, error) {
	req := esapi.IndicesCreateRequest{Index: idxName}
	return req.Do(context.Background(), c.Client)
}

func (c *ESClient) Insert(idxName string, data []byte) (*esapi.Response, error) {
	req := esapi.IndexRequest{
		Index: idxName,
		Body: bytes.NewReader(data),
	}
	return req.Do(context.Background(), c.Client)
}