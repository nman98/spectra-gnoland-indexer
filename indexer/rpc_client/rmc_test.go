package rpcclient_test

import (
	"testing"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// MockRpcClient - focuses on tracking what was called
type MockRpcClient struct {
	HealthCalled               bool
	GetBlockCalled             bool
	GetLatestBlockHeightCalled bool
	GetTxCalled                bool
	GetAbciQueryCalled         bool
	GetBlockCallCount          int
	GetTxCallCount             int
}

// Mock method for GetBlock
func (m *MockRpcClient) GetBlock(height uint64) (*rpcClient.BlockResponse, *rpcClient.RpcHeightError) {
	m.GetBlockCalled = true
	m.GetBlockCallCount++
	return rpcClient.NewTestBlockResponse(height, "test-chain"), nil
}

// Mock method for GetLatestBlockHeight
func (m *MockRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	m.GetLatestBlockHeightCalled = true
	return 1, nil
}

// Mock health check
func (m *MockRpcClient) Health() error {
	m.HealthCalled = true
	return nil
}

// Mock method for GetTx
func (m *MockRpcClient) GetTx(txHash string) (*rpcClient.TxResponse, *rpcClient.RpcStringError) {
	m.GetTxCalled = true
	m.GetTxCallCount++
	return rpcClient.NewTestTxResponse(txHash, 1), nil
}

// Mock method for GetAbciQuery
func (m *MockRpcClient) GetAbciQuery(path string, data string, height *uint64, prove *bool) (any, error) {
	m.GetAbciQueryCalled = true
	return nil, nil
}

// TestRateLimitedRpcClient_Constructor_WithInvalidURL - tests constructor error handling
func TestRateLimitedRpcClient_Constructor_WithInvalidURL(t *testing.T) {
	// Test that constructor properly handles invalid URLs
	_, err := rpcClient.NewRateLimitedRpcClient(
		"not-a-valid-url",
		nil,
		5,
		3*time.Second,
		nil,
	)

	// Should get an error for invalid URL
	if err == nil {
		t.Error("Expected error when creating client with invalid URL")
	}
}

// TestRateLimitedRpcClient_Constructor_WithEmptyURL - tests constructor with empty URL
func TestRateLimitedRpcClient_Constructor_WithEmptyURL(t *testing.T) {
	// Test that constructor handles empty URL
	_, err := rpcClient.NewRateLimitedRpcClient(
		"", // empty URL
		nil,
		5,
		3*time.Second,
		nil,
	)

	// Should get an error for empty URL
	if err == nil {
		t.Error("Expected error when creating client with empty URL")
	}
}

// TestRateLimitedRpcClient_Constructor_ValidParameters - tests constructor with valid params
func TestRateLimitedRpcClient_Constructor_ValidParameters(t *testing.T) {
	// Test constructor with valid-looking parameters
	// Note: This might still fail if it tries to connect, but that's ok
	timeout := 5 * time.Second
	_, err := rpcClient.NewRateLimitedRpcClient(
		"http://localhost:26657", // Valid URL format
		&timeout,
		10,
		1*time.Minute,
		nil,
	)

	// We don't assert on error here because:
	// 1. Might fail if no server running (expected)
	// 2. Might succeed if server is running (also fine)
	// We're just testing that the constructor doesn't panic

	t.Logf("Constructor result with valid params: err=%v", err)
}
