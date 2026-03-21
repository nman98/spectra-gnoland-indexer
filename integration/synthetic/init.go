package synthetic

import (
	"context"
	"log"
	"runtime"
	"sync"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/orchestrator"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

// SyntheticIntegrationTestConfig holds configuration for synthetic integration tests
type SyntheticIntegrationTestConfig struct {
	DatabaseConfig database.DatabasePoolConfig
	ChainID        string
	FromHeight     uint64
	ToHeight       uint64
}

// RunSyntheticIntegrationTest runs a full integration test using synthetic data.
// Data is generated and inserted in chunks so that only one chunk's worth of
// synthetic blocks/transactions/commits lives in memory at a time.
func RunSyntheticIntegrationTest(testConfig *SyntheticIntegrationTestConfig) error {
	log.Printf("Starting synthetic integration test from height %d to %d", testConfig.FromHeight, testConfig.ToHeight)

	// Initialize database
	db := database.NewTimescaleDb(testConfig.DatabaseConfig)
	log.Printf("Connected to database successfully")

	// Initialize address caches (required by data processor). These are shared
	// across all chunks so accumulated address state carries forward correctly.
	wg := sync.WaitGroup{}
	wg.Add(2)
	var validatorCache *addressCache.AddressCache
	var addrCache *addressCache.AddressCache
	go func() {
		defer wg.Done()
		validatorCache = addressCache.NewAddressCache(testConfig.ChainID, db, true)
	}()
	go func() {
		defer wg.Done()
		addrCache = addressCache.NewAddressCache(testConfig.ChainID, db, false)
	}()
	wg.Wait()
	log.Printf("Initialized address caches")

	// Initialize data processor
	dataProc := dataProcessor.NewDataProcessor(db, addrCache, validatorCache, testConfig.ChainID)
	log.Printf("Initialized data processor")

	orchConfig := &config.Config{
		MaxBlockChunkSize:       500,
		MaxTransactionChunkSize: 1000,
	}

	chunkSize := orchConfig.MaxBlockChunkSize
	totalChunks := (testConfig.ToHeight - testConfig.FromHeight + chunkSize) / chunkSize
	currentChunk := uint64(0)

	// Process height range in chunks. A fresh SyntheticQueryOperator is created
	// for each chunk so its in-memory maps hold at most one chunk's data at a time.
	// Once the orchestrator finishes processing a chunk and the operator goes out
	// of scope the GC can reclaim that memory before the next chunk is generated.
	for chunkStart := testConfig.FromHeight; chunkStart <= testConfig.ToHeight; chunkStart += chunkSize {
		chunkEnd := min(chunkStart+chunkSize-1, testConfig.ToHeight)
		currentChunk++

		log.Printf("Generating chunk %d/%d: blocks %d to %d", currentChunk, totalChunks, chunkStart, chunkEnd)

		syntheticQueryOp := NewSyntheticQueryOperator(testConfig.ChainID, chunkStart, chunkEnd)

		mockDbHeight := &MockDatabaseHeight{lastHeight: chunkStart - 1}
		mockGnoRpc := &MockGnolandRpcClient{latestHeight: chunkEnd}

		orch := orchestrator.NewOrchestrator(
			"historic",
			orchConfig,
			testConfig.ChainID,
			mockDbHeight,
			mockGnoRpc,
			dataProc,
			syntheticQueryOp,
		)

		orch.HistoricProcess(chunkStart, chunkEnd, false)

		log.Printf("Chunk %d/%d complete, freeing synthetic data", currentChunk, totalChunks)

		// Drop the references so the GC can reclaim the chunk's maps before
		// the next iteration allocates new ones.
		// We force GC here to try to speed up the process as much as
		// possible while also trying not to store much data in RAM.
		syntheticQueryOp = nil //nolint:all
		orch = nil
		runtime.GC()
	}

	log.Printf("Synthetic integration test completed successfully!")
	return nil
}

// Mock implementations for the interfaces that orchestrator needs

// MockDatabaseHeight implements the DatabaseHeight interface
type MockDatabaseHeight struct {
	lastHeight uint64
}

func (m *MockDatabaseHeight) GetLastBlockHeight(ctx context.Context, chainName string) (uint64, error) {
	return m.lastHeight, nil
}

// MockGnolandRpcClient implements the GnolandRpcClient interface
type MockGnolandRpcClient struct {
	latestHeight uint64
}

func (m *MockGnolandRpcClient) GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError) {
	return m.latestHeight, nil
}
