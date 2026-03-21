package synthetic

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"math/rand/v2"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/generator"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var (
	// september 1st 2025, time is UTC at midnight
	baseTimestamp = time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	// 1 block per 5 seconds, real chain could be different
	blockProductionRate = time.Second * 5
)

// SyntheticQueryOperator implements orchestrator.QueryOperator interface
// but returns generated synthetic data instead of querying real RPC
type SyntheticQueryOperator struct {
	generator        *generator.DataGenerator
	chainID          string
	currentHeight    uint64
	signedValidators []string

	// Pre-generated data for consistent testing
	blocks       map[uint64]*rpcClient.BlockResponse
	transactions map[string]*rpcClient.TxResponse
	commits      map[uint64]*rpcClient.CommitResponse
	// response maker
	responseMaker *ResponseMaker
}

// NewSyntheticQueryOperator creates a new synthetic query operator.
// It generates data only for the given [fromHeight, toHeight] range,
// so callers should create one instance per chunk to bound memory usage.
func NewSyntheticQueryOperator(chainID string, fromHeight uint64, toHeight uint64) *SyntheticQueryOperator {
	gen := generator.NewDataGenerator(500)

	// Generate a consistent set of validator addresses for all blocks
	numValidators := 50 //capped to 50
	genVal := generator.NewDataGenerator(50)
	valAddr := genVal.GetAllBech32Addresses()
	validators := make([]string, 0, numValidators)
	for i := range numValidators {
		validators = append(validators, valAddr[i])
	}

	sq := &SyntheticQueryOperator{
		generator:        gen,
		chainID:          chainID,
		currentHeight:    toHeight,
		signedValidators: validators,
		blocks:           make(map[uint64]*rpcClient.BlockResponse),
		transactions:     make(map[string]*rpcClient.TxResponse),
		commits:          make(map[uint64]*rpcClient.CommitResponse),
		responseMaker:    NewResponseMaker(gen),
	}

	sq.preGenerateData(fromHeight, toHeight)
	return sq
}

// GetFromToBlocks implements the QueryOperator interface by returning synthetic blocks
func (sq *SyntheticQueryOperator) GetFromToBlocks(fromHeight uint64, toHeight uint64) []*rpcClient.BlockResponse {
	diff := toHeight - fromHeight + 1
	if diff < 1 {
		return nil
	}

	blocks := make([]*rpcClient.BlockResponse, 0, diff)
	for height := fromHeight; height <= toHeight; height++ {
		block := sq.getBlock(height)
		if block != nil {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// GetTransactions implements the QueryOperator interface by returning synthetic transactions
func (sq *SyntheticQueryOperator) GetTransactions(txHashes []string) []*rpcClient.TxResponse {
	if len(txHashes) < 1 {
		return nil
	}

	transactions := make([]*rpcClient.TxResponse, 0, len(txHashes))
	for _, hash := range txHashes {
		tx := sq.getTransaction(hash)
		if tx != nil {
			transactions = append(transactions, tx)
		}
	}
	return transactions
}

// GetFromToCommits implements the QueryOperator interface by returning synthetic commits
func (sq *SyntheticQueryOperator) GetFromToCommits(fromHeight uint64, toHeight uint64) []*rpcClient.CommitResponse {
	diff := toHeight - fromHeight + 1
	if diff < 1 {
		return nil
	}

	commits := make([]*rpcClient.CommitResponse, 0, diff)

	for height := fromHeight; height <= toHeight; height++ {
		commits = append(commits, sq.getCommit(height))
	}
	return commits
}

// GetLatestBlockHeight implements the QueryOperator interface
func (sq *SyntheticQueryOperator) GetLatestBlockHeight() (uint64, error) {
	return sq.currentHeight, nil
}

// preGenerateData creates a consistent dataset of blocks, transactions and commits
func (sq *SyntheticQueryOperator) preGenerateData(fromHeight uint64, maxHeight uint64) {
	startTime := time.Now()
	// Generate blocks from height start to maxHeight
	for height := fromHeight; height <= maxHeight; height++ {
		sq.createSynthBlock(height)
		sq.createCommit(height)

		// Log progress every 100 blocks
		if height%100 == 0 {
			log.Printf("Generated %d/%d blocks (%.1f%%)", height, maxHeight, float64(height-fromHeight+1)/float64(maxHeight-fromHeight+1)*100)
		}
	}
	log.Printf("Pre-generated data for blocks from %d to %d in %v", fromHeight, maxHeight, time.Since(startTime))
	log.Printf("Total transactions generated: %d", len(sq.transactions))
	log.Printf("Total commits generated: %d", len(sq.commits))
}

// getBlock returns existing block or creates a new one
func (sq *SyntheticQueryOperator) getBlock(height uint64) *rpcClient.BlockResponse {
	if block, ok := sq.blocks[height]; ok {
		return block
	}
	// until I test this let it throw an error and shut down the program
	log.Fatal("block not found")
	return nil
}

// createBlock generates a synthetic block for the given height
func (sq *SyntheticQueryOperator) createSynthBlock(height uint64) (*rpcClient.BlockResponse, []*rpcClient.TxResponse) {
	// Generate 0-10 transactions per block (weighted toward fewer)
	numTxs := 0
	randomVal := rand.Float32() // Use deterministic for now, can be made random later
	switch {
	case randomVal < 0.3: // 30% chance of 0 transactions
		numTxs = 0
	case randomVal > 0.3 && randomVal < 0.6: // 30% chance of 1-2 transactions
		numTxs = 1 + int(height%2)
	case randomVal > 0.6 && randomVal < 0.85: // 25% chance of 3-5 transactions
		numTxs = 3 + int(height%3)
	default: // 15% chance of 6-10 transactions
		numTxs = 6 + int(height%5)
	}

	// Generate transactions for this block
	var txResponses []*rpcClient.TxResponse
	txRaws := make([]string, numTxs)   // Store raw tx data (for block)
	txHashes := make([]string, numTxs) // Store hashes (for lookup)
	for i := 0; i < numTxs; i++ {
		// we will need to generate a full transaction here, later it should be processed
		// by the indexer, hopefully...
		txEvents, stdTx := sq.generator.GenerateTransaction()
		// now it should encode the data and then the bytes need to be encoded to base64
		bz, err := amino.Marshal(stdTx)
		if err != nil {
			log.Fatal(err)
		}
		base64Encoded := base64.StdEncoding.EncodeToString(bz)

		// to get the "nice hash" decode the txRaw to base64 and then to sha256
		txRawBytes, err := base64.StdEncoding.DecodeString(base64Encoded)
		if err != nil {
			log.Fatal(err)
		}
		txHash := sha256.Sum256(txRawBytes)
		txHashString := base64.StdEncoding.EncodeToString(txHash[:])

		txRaws[i] = base64Encoded  // Store raw tx data in block
		txHashes[i] = txHashString // Store hash for transaction lookup

		txResponse := sq.createTransaction(txHashString, height, base64Encoded, &txEvents)
		txResponses = append(txResponses, txResponse)
	}

	blockTimestamp := baseTimestamp.Add(time.Duration(height-1) * blockProductionRate)

	// Create the block using existing synthetic response maker
	blockInput := GenBlockInput{
		Height:    height,
		ChainID:   sq.chainID,
		Timestamp: blockTimestamp,
		TxsRaw:    txRaws,
	}

	block := sq.responseMaker.GenerateBlockResponse(blockInput)
	sq.blocks[height] = block

	return block, txResponses
}

// getOrCreateTransaction returns existing transaction or creates a new one
func (sq *SyntheticQueryOperator) getTransaction(txHash string) *rpcClient.TxResponse {
	if tx, ok := sq.transactions[txHash]; ok {
		return tx
	}
	log.Fatal("transaction not found")
	return nil
}

// createTransaction generates a synthetic transaction
func (sq *SyntheticQueryOperator) createTransaction(
	txHash string,
	height uint64,
	txRaw string,
	txEvents *generator.TxEvents,
) *rpcClient.TxResponse {

	txInput := GenTransactionInput{
		TxRaw:  txRaw,
		TxHash: txHash,
		Height: height,
		Events: txEvents,
	}

	tx := sq.responseMaker.GenerateTransactionResponse(txInput)
	sq.transactions[txHash] = tx

	return tx
}

// createCommit generates a synthetic commit for the given height
func (sq *SyntheticQueryOperator) createCommit(height uint64) *rpcClient.CommitResponse {
	commitInput := GenCommitInput{
		Height:           height,
		ChainID:          sq.chainID,
		Timestamp:        baseTimestamp.Add(time.Duration(height-1) * blockProductionRate),
		ProposerAddress:  sq.signedValidators[height%uint64(len(sq.signedValidators))], // random validator
		SignedValidators: sq.signedValidators,
	}
	commit := sq.responseMaker.GenerateCommitResponse(commitInput)
	sq.commits[height] = commit
	return commit
}

func (sq *SyntheticQueryOperator) getCommit(height uint64) *rpcClient.CommitResponse {
	if commit, ok := sq.commits[height]; ok {
		return commit
	}
	log.Fatal("commit not found")
	return nil
}

// SetCurrentHeight allows updating the current height for live testing scenarios
func (sq *SyntheticQueryOperator) SetCurrentHeight(height uint64) {
	sq.currentHeight = height
}

// AddBlock allows adding a new block at the current height for live simulation
func (sq *SyntheticQueryOperator) AddBlock() (*rpcClient.BlockResponse, []*rpcClient.TxResponse) {
	sq.currentHeight++
	return sq.createSynthBlock(sq.currentHeight)
}

// GetGeneratedBlocks returns all pre-generated blocks
func (sq *SyntheticQueryOperator) GetGeneratedBlocks() map[uint64]*rpcClient.BlockResponse {
	return sq.blocks
}

// GetGeneratedTransactions returns all pre-generated transactions
func (sq *SyntheticQueryOperator) GetGeneratedTransactions() map[string]*rpcClient.TxResponse {
	return sq.transactions
}
