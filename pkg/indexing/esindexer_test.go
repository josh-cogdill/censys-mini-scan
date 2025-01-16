package indexing

import (
	"errors"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	RESPONSE_404 *esapi.Response = &esapi.Response{StatusCode: 404}
	RESPONSE_200 *esapi.Response = &esapi.Response{StatusCode: 200}
	RESPONSE_400 *esapi.Response = &esapi.Response{StatusCode: 400}
	MOCK_IDX string = "mockIdx"
)

func TestCreateIndex(t *testing.T) {
	t.Run("Create Index Success", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Index does not exist
		mockClient.On("IndicesExists", []string{MOCK_IDX}).Return(RESPONSE_404, nil)
		mockClient.On("IndicesCreate", MOCK_IDX).Return(RESPONSE_200, nil)

		indexer := NewIndexer(mockClient)

		// Act
		err := indexer.CreateIndex(MOCK_IDX)

		// Assert
		assert.NoError(t, err)
		mockClient.AssertCalled(t, "IndicesExists", []string{MOCK_IDX})
		mockClient.AssertCalled(t, "IndicesCreate", MOCK_IDX)
	})
	t.Run("Index Already Exists - Does Not Create", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		mockClient.On("IndicesExists", []string{MOCK_IDX}).Return(RESPONSE_200, nil)

		indexer := ESIndexer{Client: mockClient}

		// Act
		err := indexer.CreateIndex(MOCK_IDX)

		// Assert
		assert.NoError(t, err)
		mockClient.AssertCalled(t, "IndicesExists", []string{MOCK_IDX})
		mockClient.AssertNotCalled(t, "IndicesCreate", MOCK_IDX)
	})

	t.Run("Create Index Returns Bad Request", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Index does not exist
		mockClient.On("IndicesExists", []string{MOCK_IDX}).Return(RESPONSE_404, nil)
		mockClient.On("IndicesCreate", MOCK_IDX).Return(RESPONSE_400, nil)

		indexer := ESIndexer{Client: mockClient}

		// Act
		err := indexer.CreateIndex(MOCK_IDX)

		// Assert
		assert.Error(t, err)
		mockClient.AssertCalled(t, "IndicesExists", []string{MOCK_IDX})
		mockClient.AssertCalled(t, "IndicesCreate", MOCK_IDX)
	})

	t.Run("Index Exists Returns error", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Not sure what esapi response will be when an error occurs. Using StatusCode 400
		mockClient.On("IndicesExists", []string{MOCK_IDX}).Return(RESPONSE_400, errors.New("Mock Error"))

		indexer := ESIndexer{Client: mockClient}

		// Act
		err := indexer.CreateIndex(MOCK_IDX)

		// Assert
		assert.Error(t, err)
		mockClient.AssertCalled(t, "IndicesExists", []string{MOCK_IDX})
		mockClient.AssertNotCalled(t, "IndicesCreate", MOCK_IDX)
	})
}

func TestExecuteInsert(t *testing.T) {
	var mockData = []byte(`{"json": "value"}`)
	t.Run("Insert Success - One Attempt", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Mock Insert to succeed on the first attempt
		mockClient.On("Insert", MOCK_IDX, mockData).Return(RESPONSE_200, nil)

		indexer := &ESIndexer{Client: mockClient}

		// Act
		err := indexer.ExecuteInsert(mockData, MOCK_IDX)

		// Assert
		assert.NoError(t, err)
		mockClient.AssertCalled(t, "Insert", MOCK_IDX, mockData)
		mockClient.AssertNumberOfCalls(t, "Insert", 1)
	})

	t.Run("Insert Success - Second Attempt", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Fail the first attempt
		mockClient.On("Insert", MOCK_IDX, mockData).Return(RESPONSE_200, errors.New("First Error")).Once()
		mockClient.On("Insert", MOCK_IDX, mockData).Return(RESPONSE_200, nil).Once()

		indexer := &ESIndexer{Client: mockClient}

		// Act
		err := indexer.ExecuteInsert(mockData, MOCK_IDX)

		// Assert
		assert.NoError(t, err)
		mockClient.AssertCalled(t, "Insert", MOCK_IDX, mockData)
		mockClient.AssertNumberOfCalls(t, "Insert", 2)
	})

	t.Run("Insert Returns Error - Exceeded Retry Limit", func(t *testing.T) {
		// Arrange
		mockClient := new(MockClient)

		// Fail all attempts
		mockClient.On("Insert", MOCK_IDX, mockData).Return(RESPONSE_200, errors.New("Error")).Times(5)

		indexer := &ESIndexer{Client: mockClient}

		// Act
		err := indexer.ExecuteInsert(mockData, MOCK_IDX)

		// Assert
		assert.Error(t, err)
		mockClient.AssertCalled(t, "Insert", MOCK_IDX, mockData)
		mockClient.AssertNumberOfCalls(t, "Insert", 5)
	})
}


// Mock implementation of ClientInterface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) IndicesExists(indices []string) (*esapi.Response, error) {
	args := m.Called(indices)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func (m *MockClient) IndicesCreate(indexName string) (*esapi.Response, error) {
	args := m.Called(indexName)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func (m *MockClient) Insert(indexName string, data []byte) (*esapi.Response, error) {
	args := m.Called(indexName, data)
	return args.Get(0).(*esapi.Response), args.Error(1)
}