package dataprocessor_test

import (
	"context"
	"testing"
	"time"

	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// Simple Mock Database for basic testing
type MockDatabase struct {
	InsertBlocksCalled       bool
	InsertTransactionsCalled bool
	LastInsertError          error
}

func (m *MockDatabase) InsertBlocks(ctx context.Context, blocks []sqlDataTypes.Blocks) error {
	m.InsertBlocksCalled = true
	return m.LastInsertError
}

func (m *MockDatabase) InsertValidatorBlockSignings(ctx context.Context, signings []sqlDataTypes.ValidatorBlockSigning) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertTransactionsGeneral(ctx context.Context, transactions []sqlDataTypes.TransactionGeneral) error {
	m.InsertTransactionsCalled = true
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgSend(ctx context.Context, messages []sqlDataTypes.MsgSend) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgCall(ctx context.Context, messages []sqlDataTypes.MsgCall) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgAddPackage(ctx context.Context, messages []sqlDataTypes.MsgAddPackage) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgRun(ctx context.Context, messages []sqlDataTypes.MsgRun) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertAddressTx(ctx context.Context, addresses []sqlDataTypes.AddressTx) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgMultiSend(ctx context.Context, messages []sqlDataTypes.MsgMultiSend) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgAuthCrSession(ctx context.Context, messages []sqlDataTypes.MsgAuthCrSession) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgAuthRvSession(ctx context.Context, messages []sqlDataTypes.MsgAuthRvSession) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertMsgAuthRvAllSessions(ctx context.Context, messages []sqlDataTypes.MsgAuthRvAllSessions) error {
	return m.LastInsertError
}

func (m *MockDatabase) InsertTxHashIds(ctx context.Context, txHashes []string, timestamps []time.Time, chainName string) (map[string]int64, error) {
	return nil, m.LastInsertError
}

// Simple Mock AddressCache
type MockAddressCache struct {
	ReturnID int32
}

func (m *MockAddressCache) AddressSolver(addresses []string, chainName string, insertValidators bool, retryAttempts uint8, oneByOne *bool) {
	// Do nothing for testing
}

func (m *MockAddressCache) GetAddress(address string) int32 {
	return m.ReturnID
}

// Test DataProcessor constructor and basic functionality
// This test focuses on testing the constructor and simple scenarios
// without triggering the complex goroutine logic that can hang
func TestNewDataProcessor(t *testing.T) {
	// Setup simple mocks
	mockDB := &MockDatabase{}
	mockAddressCache := &MockAddressCache{ReturnID: 1}
	mockValidatorCache := &MockAddressCache{ReturnID: 2}

	// Test constructor
	dp := dataProcessor.NewDataProcessor(mockDB, mockAddressCache, mockValidatorCache, "test-chain", 100)

	// Verify constructor returns non-nil
	if dp == nil {
		t.Fatal("Expected NewDataProcessor to return non-nil DataProcessor")
	}
}

func TestDataProcessor_WithEmptyData(t *testing.T) {
	// This test focuses on testing the constructor and simple scenarios
	// without triggering the complex goroutine logic that can hang

	mockDB := &MockDatabase{}
	mockAddressCache := &MockAddressCache{ReturnID: 123}
	mockValidatorCache := &MockAddressCache{ReturnID: 456}

	dp := dataProcessor.NewDataProcessor(mockDB, mockAddressCache, mockValidatorCache, "test-chain", 100)

	// Test that we can create the processor successfully
	if dp == nil {
		t.Fatal("DataProcessor should not be nil")
	}

	// We can test that the mocks work correctly
	addressID := mockAddressCache.GetAddress("test-address")
	if addressID != 123 {
		t.Errorf("Expected address ID 123, got %d", addressID)
	}

	validatorID := mockValidatorCache.GetAddress("test-validator")
	if validatorID != 456 {
		t.Errorf("Expected validator ID 456, got %d", validatorID)
	}
}

func TestDataProcessor_DatabaseInterface(t *testing.T) {
	// Test that our mock properly implements the Database interface
	var db dataProcessor.Database = &MockDatabase{}

	// Test interface methods
	err := db.InsertBlocks(context.Background(), []sqlDataTypes.Blocks{})
	if err != nil {
		t.Errorf("InsertBlocks should not return error with nil input, got: %v", err)
	}

	err = db.InsertTransactionsGeneral(context.Background(), []sqlDataTypes.TransactionGeneral{})
	if err != nil {
		t.Errorf("InsertTransactionsGeneral should not return error with nil input, got: %v", err)
	}

	err = db.InsertAddressTx(context.Background(), []sqlDataTypes.AddressTx{})
	if err != nil {
		t.Errorf("InsertAddressTx should not return error with nil input, got: %v", err)
	}
}

func TestDataProcessor_AddressCacheInterface(t *testing.T) {
	// Test that our mock properly implements the AddressCache interface
	var cache dataProcessor.AddressCache = &MockAddressCache{ReturnID: 999}

	// Test interface methods
	id := cache.GetAddress("test")
	if id != 999 {
		t.Errorf("Expected ID 999, got %d", id)
	}

	// Test AddressSolver doesn't panic
	cache.AddressSolver([]string{"addr1", "addr2"}, "chain", false, 3, nil)
}

// Test error handling
func TestDataProcessor_WithDatabaseError(t *testing.T) {
	mockDB := &MockDatabase{
		LastInsertError: &TestError{"database connection failed"},
	}
	mockAddressCache := &MockAddressCache{}
	mockValidatorCache := &MockAddressCache{}

	dp := dataProcessor.NewDataProcessor(mockDB, mockAddressCache, mockValidatorCache, "test-chain", 100)

	// Verify constructor still works even with error-prone database
	if dp == nil {
		t.Fatal("DataProcessor should be created even with error-prone dependencies")
	}
}

// Custom error for testing
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}
