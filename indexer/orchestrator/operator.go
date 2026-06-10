package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataprocessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

const (
	Live     = "live"
	Historic = "historic"
)

var l = logger.Get()

func NewOrchestrator(
	runningMode string,
	config *config.Config,
	chainName string,
	db DatabaseHeight,
	gnoRpcClient GnolandRpcClient,
	dataProcessor DataProcessor,
	queryOperator QueryOperator,
) *Orchestrator {
	if runningMode != Live && runningMode != Historic {
		panic("invalid running mode, please choose between live and historic")
	}
	return &Orchestrator{
		runningMode:             runningMode,
		config:                  config,
		chainName:               chainName,
		db:                      db,
		gnoRpcClient:            gnoRpcClient,
		dataProcessor:           dataProcessor,
		queryOperator:           queryOperator,
		isProcessing:            false,
		currentProcessingHeight: 0,
	}
}

func (or *Orchestrator) HistoricProcess(
	fromHeight uint64,
	toHeight uint64,
	compressEvents bool) {
	l.Info().Msgf("Starting historic process from %d to %d", fromHeight, toHeight)
	startTime := time.Now()

	// Track processing state
	or.isProcessing = true
	or.currentProcessingHeight = fromHeight
	defer func() {
		or.isProcessing = false
		l.Info().Msgf("Historic processing completed at height %d", or.currentProcessingHeight)
	}()

	for startHeight := fromHeight; startHeight <= toHeight; {
		chunkEndHeight := min(startHeight+or.config.MaxBlockChunkSize-1, toHeight)

		l.Info().Msgf("Processing chunk from %d to %d", startHeight, chunkEndHeight)

		// Update current processing height
		or.currentProcessingHeight = startHeight

		// Process the chunk
		err := or.processChunk(startHeight, chunkEndHeight, compressEvents)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Err(err).
				Msgf(
					"Error processing chunk %d-%d", startHeight, chunkEndHeight,
				)

		}

		// Always advance to next chunk, regardless of whether blocks were found
		// Update processing height to end of chunk
		or.currentProcessingHeight = chunkEndHeight
		startHeight = chunkEndHeight + 1
	}

	totalDuration := time.Since(startTime)
	l.Info().Msgf("Historic process completed from %d to %d in %v", fromHeight, toHeight, totalDuration)
}

func (or *Orchestrator) LiveProcess(ctx context.Context, skipInitialDbCheck bool, compressEvents bool) {
	l.Info().Msg("Starting live block processing")

	var lastProcessedHeight uint64
	var err error

	// Track our current processing state for potential cleanup
	or.isProcessing = true
	or.currentProcessingHeight = 0
	defer func() {
		or.isProcessing = false
		l.Info().Msgf("Live processing stopped at height %d", or.currentProcessingHeight)
	}()

	// Initial setup - get starting height
	if !skipInitialDbCheck {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		lastProcessedHeight, err = or.db.GetLastBlockHeight(ctx, or.chainName)
		if err != nil {
			l.Error().
				Caller().Stack().Err(err).Msgf("Failed to get last block height from database: %v", err)
			l.Info().Msg("Database might not work or hasn't been initialized with necessary table schemas.")
			l.Fatal().Msg("Shutting down indexer...")
		}
		l.Info().Msgf("Retrieved last processed height from database: %d", lastProcessedHeight)
	} else {
		// Get latest block height from chain
		latestHeight, rpcErr := or.gnoRpcClient.GetLatestBlockHeight()
		if rpcErr != nil {
			l.Error().
				Caller().Stack().Err(rpcErr).Msgf("Failed to get latest block height from chain: %v", rpcErr)
			return
		}
		lastProcessedHeight = latestHeight
		l.Info().Msgf("Starting from latest chain height: %d (skipping database check)", lastProcessedHeight)
	}

	or.currentProcessingHeight = lastProcessedHeight
	lastProgressTime := time.Now()

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("Live process interrupted by context cancellation")
			or.saveProcessingState(lastProcessedHeight, "live_interrupted")
			return
		default:
		}

		// Get the latest block height from the chain
		latestHeight, rpcErr := or.gnoRpcClient.GetLatestBlockHeight()
		if rpcErr != nil {
			l.Error().
				Caller().
				Stack().
				Err(rpcErr).
				Msgf("Error fetching latest block height")
			time.Sleep(or.config.LivePooling)
			continue
		}

		blocksBehind := latestHeight - lastProcessedHeight

		// If caught up, wait and continue
		if blocksBehind <= 0 {
			l.Info().Msgf("Caught up to height %d. Waiting %d seconds...", latestHeight, or.config.LivePooling/time.Second)
			time.Sleep(or.config.LivePooling)
			continue
		}

		// Adjust chunk size based on how far behind we are
		currentChunkSize := min(blocksBehind, or.config.MaxBlockChunkSize)

		chunkStart := lastProcessedHeight + 1
		chunkEnd := min(chunkStart+currentChunkSize-1, latestHeight)

		l.Info().Msgf("Processing live chunk %d-%d (behind by %d blocks)", chunkStart, chunkEnd, blocksBehind)

		// Update current processing height
		or.currentProcessingHeight = chunkStart

		// Process this chunk
		err = or.processChunk(chunkStart, chunkEnd, compressEvents)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Err(err).
				Msgf("Error processing live chunk %d-%d", chunkStart, chunkEnd)
			time.Sleep(or.config.LivePooling)
			continue
		}

		// Update progress
		lastProcessedHeight = chunkEnd
		or.currentProcessingHeight = chunkEnd
		or.updateProgressMetrics(chunkStart, chunkEnd, blocksBehind, &lastProgressTime)

		// Small delay to prevent overwhelming the API
		time.Sleep(50 * time.Millisecond)
	}
}

// processChunk processes a single chunk of blocks for live processing
func (or *Orchestrator) processChunk(chunkStart, chunkEnd uint64, compressEvents bool) error {
	chunkStartTime := time.Now()

	// Step 1: Get blocks concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	var blocks []*rpcClient.BlockResponse
	var commits []*rpcClient.CommitResponse

	go func() {
		defer wg.Done()
		blocks = or.queryOperator.GetFromToBlocks(chunkStart, chunkEnd)
	}()
	go func() {
		defer wg.Done()
		commits = or.queryOperator.GetFromToCommits(chunkStart, chunkEnd)
	}()

	wg.Wait()

	fetchDuration := time.Since(chunkStartTime)
	l.Debug().Msgf("Chunk %d-%d fetched in %v", chunkStart, chunkEnd, fetchDuration)

	if len(blocks) == 0 && len(commits) == 0 {
		l.Info().Msgf("No valid blocks in live chunk %d-%d", chunkStart, chunkEnd)
		return nil
	}

	// Step 2: Collect all transactions from all blocks in this chunk
	allTransactions := or.collectTransactionsFromBlocks(blocks)

	l.Info().Msgf("Collected %d transactions from %d blocks in live chunk", len(allTransactions), len(blocks))

	// Step 3: Process all data concurrently
	if err := or.processAll(blocks, commits, allTransactions, compressEvents, chunkStart, chunkEnd); err != nil {
		return fmt.Errorf("failed to process live chunk %d-%d: %w", chunkStart, chunkEnd, err)
	}

	chunkDuration := time.Since(chunkStartTime)
	l.Info().Msgf("Chunk %d-%d completed in %v", chunkStart, chunkEnd, chunkDuration)

	return nil
}

// updateProgressMetrics updates and logs progress metrics for live processing
func (or *Orchestrator) updateProgressMetrics(
	chunkStart, chunkEnd, blocksBehind uint64,
	lastProgressTime *time.Time,
) {
	now := time.Now()
	timeSinceLastProgress := now.Sub(*lastProgressTime)

	// Log progress every 30 seconds or significant milestones
	if timeSinceLastProgress >= 30*time.Second || blocksBehind <= 10 {
		blocksProcessed := chunkEnd - chunkStart + 1
		l.Info().Msgf("Live progress: processed %d blocks (%d-%d), %d blocks behind, last update: %v ago",
			blocksProcessed, chunkStart, chunkEnd, blocksBehind, timeSinceLastProgress.Round(time.Second))
		*lastProgressTime = now
	}
}

/*
	collectTransactionsFromBlocks extracts all transactions from blocks and queries them concurrently

Parameters:
  - blocks: a slice of blocks

Returns:
  - a slice of transactions

The method will not throw an error if the transactions are not found, it will just return an empty slice.
*/
func (or *Orchestrator) collectTransactionsFromBlocks(blocks []*rpcClient.BlockResponse) []dataprocessor.TransactionsData {
	// Collect all transaction hashes from all blocks
	var allTxHashes []string
	blockTxData := make([]struct {
		txHash      string
		blockHeight uint64
		timestamp   time.Time
	}, 0)

	for _, block := range blocks {
		if block == nil {
			continue
		}

		// this should not be nil but just in case
		// also we need to collect all of the tx hashes, decode the base64 to raw bytes then to sha256
		// and then sha256 to base64
		txHashes := block.GetTxHashes()
		if txHashes == nil {
			continue
		}
		for _, txHash := range txHashes {
			txHashBytes, err := base64.StdEncoding.DecodeString(txHash)
			if err != nil {
				continue
			}
			txHashSha256 := sha256.Sum256(txHashBytes)
			txHashFinal := base64.StdEncoding.EncodeToString(txHashSha256[:])
			blockHeight, err := block.GetHeight()
			if err != nil {
				l.Error().
					Caller().
					Stack().
					Err(err).
					Msgf("Failed to get block height")
				continue
			}
			allTxHashes = append(allTxHashes, txHashFinal)
			blockTxData = append(blockTxData, struct {
				txHash      string
				blockHeight uint64
				timestamp   time.Time
			}{
				txHash:      txHashFinal,
				blockHeight: blockHeight,
				timestamp:   block.GetTimestamp(),
			})
		}
	}

	if len(allTxHashes) == 0 {
		return make([]dataprocessor.TransactionsData, 0)
	}

	l.Info().Msgf("Fetching %d transactions", len(allTxHashes))

	// Query all transactions
	transactions := or.queryOperator.GetTransactions(allTxHashes)

	// Build map of transactions with their timestamps
	txData := make([]dataprocessor.TransactionsData, 0)
	for _, tx := range transactions {
		if tx != nil {
			for _, blockTx := range blockTxData {
				if blockTx.txHash == tx.GetHash() {
					txData = append(txData, dataprocessor.TransactionsData{
						Response:    tx,
						Timestamp:   blockTx.timestamp,
						BlockHeight: blockTx.blockHeight,
					})
				}
			}
		}
	}

	l.Info().Msgf("Successfully collected %d valid transactions", len(txData))
	return txData
}

// This function processes all data using optimized concurrent execution.
//
// Parameters:
//   - blocks: a slice of blocks
//   - transactions: a map of transactions and timestamps
//   - compressEvents: if true, compress the events
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - error: if processing fails
//
// The method will not throw an error if the data is not found, it will just return nil.
func (or *Orchestrator) processAll(
	blocks []*rpcClient.BlockResponse,
	commits []*rpcClient.CommitResponse,
	transactions []dataprocessor.TransactionsData,
	compressEvents bool,
	fromHeight uint64,
	toHeight uint64) error {

	phase1Done := make(chan struct{})
	phase2Done := make(chan struct{})
	hasTxs := len(transactions) > 0

	ctx := &processingContext{
		blocks:         blocks,
		commits:        commits,
		transactions:   transactions,
		compressEvents: compressEvents,
		fromHeight:     fromHeight,
		toHeight:       toHeight,
		hasTxs:         hasTxs,
	}

	go or.processPhase1(phase1Done, ctx)
	go or.processPhase2(phase1Done, phase2Done, ctx)
	<-phase2Done

	l.Info().Msgf("All processing completed successfully from %d to %d", fromHeight, toHeight)
	return nil
}

func (or *Orchestrator) processPhase1(
	phase1Done chan struct{},
	ctx *processingContext,
) {
	defer close(phase1Done)
	var wg1 sync.WaitGroup
	wg1.Go(func() {
		l.Info().Msg("Phase 1: Starting ProcessValidatorAddresses")
		or.dataProcessor.ProcessValidatorAddresses(ctx.blocks, ctx.fromHeight, ctx.toHeight)
		l.Info().Msg("Phase 1: ProcessValidatorAddresses completed")
	})

	if ctx.hasTxs {
		wg1.Go(func() {
			l.Info().Msg("Phase 1: Starting ProcessTxHashIds")
			or.dataProcessor.ProcessTxHashIds(ctx.transactions)
			l.Info().Msg("Phase 1: ProcessTxHashIds completed")
		})
	}

	wg1.Wait()
	l.Info().Msg("Phase 1: chunk processing completed")
}

func (or *Orchestrator) processPhase2(
	phase1Done chan struct{},
	phase2Done chan struct{},
	ctx *processingContext,
) {
	<-phase1Done
	defer close(phase2Done)
	var wg2 sync.WaitGroup
	wg2.Go(func() {
		l.Info().Msg("Phase 2: Starting ProcessBlocks")
		or.dataProcessor.ProcessBlocks(ctx.blocks, ctx.fromHeight, ctx.toHeight)
		l.Info().Msg("Phase 2: ProcessBlocks completed")
	})

	wg2.Go(func() {
		l.Info().Msg("Phase 2: Starting ProcessValidatorSignings")
		or.dataProcessor.ProcessValidatorSignings(ctx.commits, ctx.fromHeight, ctx.toHeight)
		l.Info().Msg("Phase 2: ProcessValidatorSignings completed")
	})

	if ctx.hasTxs {
		wg2.Go(func() {
			l.Info().Msg("Phase 2: Starting ProcessMessages")
			or.dataProcessor.ProcessMessages(ctx.transactions, ctx.fromHeight, ctx.toHeight)
			l.Info().Msg("Phase 2: ProcessMessages completed")
		})
		wg2.Go(func() {
			l.Info().Msg("Phase 2: Starting ProcessTransactions")
			or.dataProcessor.ProcessTransactions(ctx.transactions, ctx.compressEvents, ctx.fromHeight, ctx.toHeight)
			l.Info().Msg("Phase 2: ProcessTransactions completed")
		})
	}

	wg2.Wait()
	l.Info().Msg("Phase 2: chunk processing completed")
}

// saveProcessingState is a private method that saves
// the current processing state to a file.
//
// Parameters:
//   - height: the height of the processing state
//   - reason: the reason for the processing state
//
// Returns:
//   - none
//
// The method will not throw an error if the processing state is not found, it will just return nil.
func (or *Orchestrator) saveProcessingState(height uint64, reason string) {
	state := ProcessingState{
		ChainName:               or.chainName,
		RunningMode:             or.runningMode,
		IsProcessing:            or.isProcessing,
		CurrentProcessingHeight: height,
		// it would be hard to read timestamp from the height since the program might not have the
		// data so use current time
		Timestamp: time.Now(),
		Reason:    reason,
	}

	// Create state directory if it doesn't exist
	stateDir := "state_dumps"
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		l.Error().
			Caller().
			Stack().
			Err(err).
			Msgf("Failed to create state directory")
		return
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("processing_state_%s_%d.json",
		time.Now().Format("20060102_150405"), height)
	filepath := filepath.Join(stateDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Err(err).
			Msgf("Failed to marshal processing state")
		return
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		l.Error().
			Caller().
			Stack().
			Err(err).
			Msgf("Failed to write processing state")
		return
	}

	l.Info().Msgf("Processing state saved to %s", filepath)
}

// Cleanup performs cleanup operations for the orchestrator
//
// Returns:
//   - error: if cleanup fails
func (or *Orchestrator) Cleanup() error {
	l.Info().Msg("Starting orchestrator cleanup...")

	// Save current state before cleanup
	or.saveProcessingState(or.currentProcessingHeight, "cleanup_requested")
	l.Info().Msg("Orchestrator cleanup completed - state saved successfully")

	return nil
}

// DumpState creates an emergency state dump with current processing information
func (or *Orchestrator) DumpState() error {
	l.Info().Msg("Creating emergency state dump...")

	// Save processing state
	or.saveProcessingState(or.currentProcessingHeight, "emergency_dump")

	// Create additional diagnostic information
	diagnostics := map[string]any{
		"chain_name":                or.chainName,
		"running_mode":              or.runningMode,
		"is_processing":             or.isProcessing,
		"current_processing_height": or.currentProcessingHeight,
		"config": map[string]any{
			"max_block_chunk_size": or.config.MaxBlockChunkSize,
			"live_pooling":         or.config.LivePooling,
			"rpc_url":              or.config.RpcUrl,
		},
		// it would be hard to read timestamp from the height since the program might not have the data so use
		// current time
		"timestamp":   time.Now(),
		"dump_reason": "emergency_shutdown",
	}

	// Create diagnostics directory if it doesn't exist
	diagDir := "diagnostics"
	if err := os.MkdirAll(diagDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagnostics directory: %w", err)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("emergency_dump_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(diagDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write diagnostics: %w", err)
	}

	l.Info().Msgf("Emergency state dump saved to %s", filepath)
	return nil
}
