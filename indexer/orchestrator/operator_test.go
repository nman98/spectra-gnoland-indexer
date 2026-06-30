package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataprocessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/orchestrator"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// Mock implementations for testing orchestration logic
// Mostly testing external dependencies and their interactions
// not testing the internal logic of the orchestrator

// MockDataProcessor - focuses on tracking what was called
type MockDataProcessor struct {
	ProcessValidatorAddressesCalled bool
	ProcessBlocksCalled             bool
	ProcessTransactionsCalled       bool
	ProcessMessagesCalled           bool
	ProcessValidatorSigningsCalled  bool
	ProcessTxHashIdsCalled          bool
	ProcessMessagesError            error
}

// Mock method for ProcessValidatorAddresses
func (m *MockDataProcessor) ProcessValidatorAddresses(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64) {
	m.ProcessValidatorAddressesCalled = true
}

// Mock method for ProcessBlocks
func (m *MockDataProcessor) ProcessBlocks(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64) {
	m.ProcessBlocksCalled = true
}

// Mock method for ProcessTransactions
func (m *MockDataProcessor) ProcessTransactions(transactions []dataprocessor.TransactionsData, compressEvents bool, fromHeight uint64, toHeight uint64) {
	m.ProcessTransactionsCalled = true
}

// Mock method for ProcessMessages
func (m *MockDataProcessor) ProcessMessages(transactions []dataprocessor.TransactionsData, fromHeight uint64, toHeight uint64) error {
	m.ProcessMessagesCalled = true
	return m.ProcessMessagesError
}

// Mock method for ProcessValidatorSignings
func (m *MockDataProcessor) ProcessValidatorSignings(commits []*rpcClient.CommitResponse, fromHeight uint64, toHeight uint64) {
	m.ProcessValidatorSigningsCalled = true
}

// Mock method for ProcessTxHashIds
func (m *MockDataProcessor) ProcessTxHashIds(txData []dataprocessor.TransactionsData) {
	m.ProcessTxHashIdsCalled = true
}

// MockQueryOperator - returns minimal data
// The idea should be to count the number of calls to the method
// and check if the method is called with the correct parameters
type MockQueryOperator struct {
	ShouldReturnBlocks  bool
	CallCount           int
	ShouldReturnCommits bool
}

// Mock method for GetFromToBlocks
func (m *MockQueryOperator) GetFromToBlocks(fromHeight uint64, toHeight uint64) []*rpcClient.BlockResponse {
	m.CallCount++
	if !m.ShouldReturnBlocks {
		return []*rpcClient.BlockResponse{} // Return empty slice
	}

	// Return a single empty block
	return []*rpcClient.BlockResponse{{}}
}

// Mock method for GetFromToCommits
func (m *MockQueryOperator) GetFromToCommits(fromHeight uint64, toHeight uint64) []*rpcClient.CommitResponse {
	m.CallCount++
	if !m.ShouldReturnCommits {
		return []*rpcClient.CommitResponse{} // Return empty slice
	}

	// Return a single empty commit
	return []*rpcClient.CommitResponse{{}}
}

// Mock method for GetTransactions
func (m *MockQueryOperator) GetTransactions(txs []string) []*rpcClient.TxResponse {
	return nil // Empty transactions
}

// Mock method for GetLatestBlockHeight
func (m *MockQueryOperator) GetLatestBlockHeight() (uint64, error) {
	// any uint is fine just return something here
	return 100, nil
}

// MockDatabaseHeight
type MockDatabaseHeight struct {
	HeightToReturn uint64
	ShouldError    bool
}

// Mock method for GetLastBlockHeight
func (m *MockDatabaseHeight) GetLastBlockHeight(ctx context.Context, chainName string) (uint64, error) {
	if m.ShouldError {
		return 0, &DatabaseError{"database error"}
	}
	return m.HeightToReturn, nil
}

// MockGnolandRpcClient
type MockGnolandRpcClient struct {
	HeightToReturn uint64
}

// Mock method for GetLatestBlockHeight
func (m *MockGnolandRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	return m.HeightToReturn, nil
}

// Custom error type for testing
// This is a simple error type for testing
// It will be used to test the error handling of the orchestrator
type DatabaseError struct {
	Message string
}

func (e *DatabaseError) Error() string {
	return e.Message
}

// Helper to create test config
// still needs somewaht "synthetic" values
func createSimpleTestConfig() *config.Config {
	return &config.Config{
		MaxBlockChunkSize: 5,
		LivePooling:       time.Second,
		RpcUrl:            "http://test",
	}
}

// Tests focusing on orchestration behavior

func TestOrchestrator_HistoricProcess_CallsAllProcessors(t *testing.T) {
	// Setup mocks
	mockDataProcessor := &MockDataProcessor{}
	mockQueryOperator := &MockQueryOperator{
		ShouldReturnBlocks:  true, // Return at least one block
		ShouldReturnCommits: true, // Return at least one commit
	}
	mockDB := &MockDatabaseHeight{}
	mockRPC := &MockGnolandRpcClient{}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(
		"historic",
		createSimpleTestConfig(),
		"test-chain",
		mockDB,
		mockRPC,
		mockDataProcessor,
		mockQueryOperator,
	)

	// Test historic processing
	orch.HistoricProcess(context.Background(), 1, 5, false)

	// Verify orchestration: all processors should be called
	if !mockDataProcessor.ProcessValidatorAddressesCalled {
		t.Error("Expected ProcessValidatorAddresses to be called")
	}
	if !mockDataProcessor.ProcessBlocksCalled {
		t.Error("Expected ProcessBlocks to be called")
	}
	/* TODO: the current test setup doesn't have any transactions so this can't be properly tested.
	 * In the future this should be changed to test transaction processing.
	if !mockDataProcessor.ProcessTransactionsCalled {
		t.Error("Expected ProcessTransactions to be called")
	}
	if !mockDataProcessor.ProcessMessagesCalled {
		t.Error("Expected ProcessMessages to be called")
	}
	*/
	if !mockDataProcessor.ProcessValidatorSigningsCalled {
		t.Error("Expected ProcessValidatorSignings to be called")
	}

	// Verify query operator was called
	if mockQueryOperator.CallCount == 0 {
		t.Error("Expected QueryOperator to be called")
	}
}

// Test orchestrator history mode where it shouldn't process when no blocks are returned
func TestOrchestrator_HistoricProcess_SkipsProcessingWhenNoBlocks(t *testing.T) {
	// Setup mocks - no blocks returned
	mockDataProcessor := &MockDataProcessor{}
	mockQueryOperator := &MockQueryOperator{
		ShouldReturnBlocks:  false, // Return empty blocks
		ShouldReturnCommits: false, // Return empty commits
	}
	mockDB := &MockDatabaseHeight{}
	mockRPC := &MockGnolandRpcClient{}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(
		"historic",
		createSimpleTestConfig(),
		"test-chain",
		mockDB,
		mockRPC,
		mockDataProcessor,
		mockQueryOperator,
	)

	// Test historic processing
	orch.HistoricProcess(context.Background(), 1, 5, false)

	// Verify query was attempted
	if mockQueryOperator.CallCount == 0 {
		t.Error("Expected QueryOperator to be called")
	}

	// But no processing should happen with empty blocks
	if mockDataProcessor.ProcessValidatorAddressesCalled {
		t.Error("Expected no processing when no blocks returned")
	}
}

// Test orchestrator live mode where it should respect context
func TestOrchestrator_LiveProcess_RespectsContext(t *testing.T) {
	// Setup mocks
	mockDataProcessor := &MockDataProcessor{}
	mockQueryOperator := &MockQueryOperator{
		ShouldReturnBlocks:  true, // Return at least one block
		ShouldReturnCommits: true, // Return at least one commit
	}
	mockDB := &MockDatabaseHeight{HeightToReturn: 50}
	mockRPC := &MockGnolandRpcClient{HeightToReturn: 100}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(
		"live",
		createSimpleTestConfig(),
		"test-chain",
		mockDB,
		mockRPC,
		mockDataProcessor,
		mockQueryOperator,
	)

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel right away

	// Start time
	start := time.Now()

	// Test live processing - should return quickly due to context cancellation
	orch.LiveProcess(ctx, false, false)

	// Should return quickly (within 1 second)
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("LiveProcess took too long (%v), expected to return quickly on context cancellation", elapsed)
	}
}

// Test orchestrator live mode where it should skip db check
// skip db check would need more testing but should work the same way
// as the cosmos indexer
func TestOrchestrator_LiveProcess_SkipDbCheck(t *testing.T) {
	// Setup mocks
	mockDataProcessor := &MockDataProcessor{}
	mockQueryOperator := &MockQueryOperator{
		ShouldReturnBlocks:  true, // Return at least one block
		ShouldReturnCommits: true, // Return at least one commit
	}
	mockDB := &MockDatabaseHeight{ShouldError: true} // Database will error
	mockRPC := &MockGnolandRpcClient{HeightToReturn: 100}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(
		"live",
		createSimpleTestConfig(),
		"test-chain",
		mockDB,
		mockRPC,
		mockDataProcessor,
		mockQueryOperator,
	)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test live processing with skip DB check - should not fail even though DB errors
	orch.LiveProcess(ctx, true, false) // skipInitialDbCheck = true

	// Test passes if we get here without panic/fatal error
}

func TestOrchestrator_Constructor_ValidatesMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected NewOrchestrator to panic with invalid mode")
		}
	}()

	mockDataProcessor := &MockDataProcessor{}
	mockQueryOperator := &MockQueryOperator{
		ShouldReturnBlocks:  true, // Return at least one block
		ShouldReturnCommits: true, // Return at least one commit
	}
	mockDB := &MockDatabaseHeight{}
	mockRPC := &MockGnolandRpcClient{}

	// Should panic with invalid mode
	orchestrator.NewOrchestrator(
		"invalid-mode", // Invalid mode
		createSimpleTestConfig(),
		"test-chain",
		mockDB,
		mockRPC,
		mockDataProcessor,
		mockQueryOperator,
	)
}
