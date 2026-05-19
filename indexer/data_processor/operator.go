package dataprocessor

import (
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/decoder"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

var l = logger.Get()

// Constructor function for the DataProcessor struct.
//
// Parameters:
//   - db: the database connection interface
//   - addressCache: the address cache interface
//   - validatorCache: the validator cache interface
//   - chainName: the name of the chain string
//   - batchSize: the batch size for the tx hash cache
//
// Returns:
//   - *DataProcessor: the data processor
//
// The method will not throw an error if the data processor is not found, it will just return nil.
func NewDataProcessor(
	db Database,
	addressCache AddressCache,
	validatorCache AddressCache,
	chainName string,
	batchSize int,
) *DataProcessor {
	return &DataProcessor{
		dbPool:         db,
		addressCache:   addressCache,
		validatorCache: validatorCache,
		chainName:      chainName,
		txHashCache:    make(map[string]int64, batchSize),
	}
}

// ProcessValidatorAddresses is a method to process the validator addresses from a slice of blocks.
// It will process the validator addresses from the blocks and store them in a map[string]struct{}.
// Then extract the addresses from the map[string]struct{} and insert them into the address cache.
//
// Parameters:
//   - blocks: a slice of blocks
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - nil
//
// The method will not throw an error if the validator addresses are not found, it will just return nil.
func (d *DataProcessor) ProcessValidatorAddresses(
	blocks []*rpcClient.BlockResponse,
	fromHeight uint64,
	toHeight uint64,
) {
	var mu sync.Mutex
	addressesMap := make(map[string]struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))

	// Process blocks concurrently to extract addresses
	for _, block := range blocks {
		go processPrecommits(&mu, addressesMap, &wg, block)
	}

	wg.Wait()

	// Extract unique addresses from map[string]struct{}
	addresses := extractAddresses(addressesMap)

	// Retry 3 times just for the sake of it
	d.validatorCache.AddressSolver(addresses, d.chainName, true, 3, nil)
	l.Info().
		Msgf(
			"Validator addresses processed from %d to %d", fromHeight, toHeight,
		)
}

// extractAddresses is a helper function to extract the addresses from a map[string]struct{}
// it will extract the addresses from the map[string]struct{} and return a slice of strings.
// If the addresses are not found, it will just return an empty slice.
//
// Parameters:
//   - addressesMap: a map[string]struct{}
//
// Returns:
//   - a slice of strings
func extractAddresses(addressesMap map[string]struct{}) []string {
	mapSize := len(addressesMap)
	addresses := make([]string, mapSize)
	idx := 0
	for address := range addressesMap {
		addresses[idx] = address
		idx++
	}
	return addresses
}

func processPrecommits(
	mu *sync.Mutex,
	addressesMap map[string]struct{},
	wg *sync.WaitGroup,
	block *rpcClient.BlockResponse,
) {
	defer wg.Done()

	// Process precommits
	precommits := block.Result.Block.LastCommit.Precommits
	for _, precommit := range precommits {
		if precommit != nil {
			mu.Lock()
			addressesMap[precommit.ValidatorAddress] = struct{}{}
			mu.Unlock()
		}
	}

	// Process proposer
	proposer := block.Result.Block.Header.ProposerAddress
	mu.Lock()
	addressesMap[proposer] = struct{}{}
	mu.Unlock()
}

// ProcessBlocks is a "swarm" method to process the blocks from a slice of blocks
// it will process the blocks using async workers and store them directly into a result slice
// it will then insert the blocks into the database.
//
// Parameters:
//   - blocks: a slice of blocks
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - nil
//
// The method will not throw an error if the blocks are not found, it will just return nil.
func (d *DataProcessor) ProcessBlocks(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64) {
	// Preallocate slice to avoid growing allocations
	blockAmount := len(blocks)
	blocksData := make([]sqlDataTypes.Blocks, blockAmount)
	wg := sync.WaitGroup{}
	wg.Add(blockAmount)

	for idx, block := range blocks {
		go d.processBlock(idx, block, &wg, blocksData)
	}

	wg.Wait()

	// add multiplier for the timeout depending on the block amount
	timeout := 10*time.Second + (time.Duration(blockAmount) * time.Second / 5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := d.dbPool.InsertBlocks(ctx, blocksData)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to insert blocks: %v", err,
			)
	}
	l.Info().
		Msgf(
			"Blocks processed from %d to %d", fromHeight, toHeight,
		)
}

// processBlock is a helper method to process a block and store it at a pre-allocated slice.
func (d *DataProcessor) processBlock(
	idx int,
	block *rpcClient.BlockResponse,
	wg *sync.WaitGroup,
	blocksData []sqlDataTypes.Blocks,
) {
	defer wg.Done()
	hash, err := base64.StdEncoding.DecodeString(block.Result.BlockMeta.BlockID.Hash)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to decode block hash %s: %v", block.Result.BlockMeta.BlockID.Hash, err,
			)
		return
	}
	height, err := strconv.ParseUint(block.Result.Block.Header.Height, 10, 64)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to parse block height %s: %v", block.Result.Block.Header.Height, err,
			)
		return
	}
	blocksData[idx] = sqlDataTypes.Blocks{
		Hash:      hash,
		Height:    height,
		Timestamp: block.Result.Block.Header.Time,
		ChainID:   block.Result.Block.Header.ChainID,
		ChainName: d.chainName,
	}
}

// ProcessTxHashIds is a function to process the tx hash ids and store them in the tx hash cache.
// It is used to store transaction ids that will be used later when inserting.
//
// Parameters:
//	- txData - transactions data from RPC client
func (d *DataProcessor) ProcessTxHashIds(
	txData []TransactionsData,
) {
	lenData := len(txData)
	txHashes := make([]string, lenData)
	timestamps := make([]time.Time, lenData)
	for idx, tx := range txData {
		txHashes[idx] = tx.Response.GetHash()
		timestamps[idx] = tx.Timestamp
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	txHashIds, err := d.dbPool.InsertTxHashIds(ctx, txHashes, timestamps, d.chainName)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to insert tx hash ids: %v", err,
			)
		return
	}
	clear(d.txHashCache)
	maps.Copy(d.txHashCache, txHashIds)
}

// ProcessTransactions is a swarm method to process the transactions from a map of transactions and timestamps.
// It will process the transactions using async workers and collect them in a pre allocated slice
// it will then insert the transactions into the database.
//
// Parameters:
//   - transactions: a map of transactions and timestamps
//   - compressEvents: if true, compress the events
//
// Returns:
//   - nil
//
// The method will not throw an error if the transactions are not found, it will just return nil
func (d *DataProcessor) ProcessTransactions(
	transactions []TransactionsData,
	compressEvents bool,
	fromHeight uint64,
	toHeight uint64) {

	// Preallocate slice to avoid growing allocations
	transactionAmount := len(transactions)
	transactionsData := make([]sqlDataTypes.TransactionGeneral, transactionAmount)
	valid := make([]bool, transactionAmount)
	wg := sync.WaitGroup{}
	wg.Add(transactionAmount)

	for idx, transaction := range transactions {
		go d.processTransaction(idx, transaction, &wg, &valid[idx], transactionsData, compressEvents)
	}

	wg.Wait()

	// Collect only the entries that were successfully processed
	result := make([]sqlDataTypes.TransactionGeneral, 0, transactionAmount)
	for idx, ok := range valid {
		if ok {
			result = append(result, transactionsData[idx])
		}
	}

	// Add multiplier for the timeout depending on the transaction amount
	timeout := 10*time.Second + (time.Duration(len(result)) * time.Second / 5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := d.dbPool.InsertTransactionsGeneral(ctx, result); err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to insert transactions: %v", err,
			)
		return
	}
	l.Info().
		Msgf(
			"Transactions processed from %d to %d", fromHeight, toHeight,
		)
}

// processTransaction is a helper method to process a transaction and store it at a pre-allocated index.
// No mutex is needed: each goroutine owns its own slot in the pre-allocated slice.
// valid is set to true only when all steps succeed, allowing the caller to filter failed entries.
func (d *DataProcessor) processTransaction(
	idx int,
	transaction TransactionsData,
	wg *sync.WaitGroup,
	valid *bool,
	transactionsData []sqlDataTypes.TransactionGeneral,
	compressEvents bool,
) {
	defer wg.Done()
	txResult := transaction.Response.Result.TxResult

	decodedMsg := decoder.NewDecodedMsg(transaction.Response.Result.Tx)

	fee := decodedMsg.GetFee()
	msgTypes := decodedMsg.GetMsgTypes()

	txId, ok := d.txHashCache[transaction.Response.GetHash()]
	if !ok {
		l.Error().
			Caller().
			Stack().Msgf("Transaction hash not found in cache: %s", transaction.Response.GetHash())
		return
	}

	gasWanted, err := strconv.ParseUint(txResult.GasWanted, 10, 64)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to parse gas wanted %s: %v", txResult.GasWanted, err,
			)
		return
	}
	gasUsed, err := strconv.ParseUint(txResult.GasUsed, 10, 64)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to parse gas used %s: %v", txResult.GasUsed, err,
			)
		return
	}

	events, err := EventSolver(transaction.Response, compressEvents)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to solve events: %v", err,
			)
		return
	}

	// if transaction is successful, mark it as such
	success := transaction.GetSuccess()
	var errLog *string
	if !success {
		errLog = transaction.GetTransactionErrorDetails()
	}

	transactionsData[idx] = sqlDataTypes.TransactionGeneral{
		TxId:           	txId,
		ChainName:          d.chainName,
		Timestamp:          transaction.Timestamp,
		BlockHeight:        transaction.BlockHeight,
		MsgTypes:           msgTypes,
		TxEvents:           events.GetNativeEvents(),
		TxEventsCompressed: events.GetCompressedData(),
		CompressionOn:      events.IsCompressed(),
		GasUsed:            gasUsed,
		GasWanted:          gasWanted,
		FeeAmount:          fee.Amount,
		FeeDenom:           fee.Denom,
		Success:            success,
		ErrorLog:           errLog,
	}
	*valid = true
}

// ProcessMessages processes all messages from transactions using concurrent "swarm method"
// This method uses a two-phase concurrent approach:
// 1. Collect and resolve all addresses to IDs using concurrent workers and map[string]struct{}
// 2. Convert messages to database-ready format with address IDs using concurrent processing
//
// Parameters:
//   - transactions: a map of transactions and timestamps
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - error: if processing fails
func (d *DataProcessor) ProcessMessages(
	transactions []TransactionsData,
	fromHeight uint64,
	toHeight uint64) error {

	// Phase 1: Concurrent address collection using map[string]struct{}
	var mu sync.Mutex
	transactionAmount := len(transactions)
	allDecodedMsgs, addressesMap := transactionDecoding(&mu, transactions, transactionAmount)

	// Extract addresses from map[string]struct{} and resolve to IDs
	allAddresses := extractAddresses(addressesMap)

	if len(allAddresses) > 0 {
		d.addressCache.AddressSolver(allAddresses, d.chainName, false, 3, nil)
		l.Info().
			Msgf(
				"Resolved %d unique addresses for messages from %d to %d",
				len(allAddresses), fromHeight, toHeight,
			)
	}

	// Phase 2: Process message groups concurrently, each goroutine writes to its own index slot.
	msgResults := make([]*decoder.DbMessageGroups, transactionAmount)
	wg := sync.WaitGroup{}
	wg.Add(transactionAmount)

	for idx, transaction := range transactions {
		// Guard against index out of bounds
		if idx >= transactionAmount {
			wg.Done()
			continue
		}
		go d.processMessageGroup(idx, transaction, allDecodedMsgs[idx], &wg, msgResults)
	}

	wg.Wait()

	aggregatedDbGroups := &decoder.DbMessageGroups{
		MsgSend:   make([]sqlDataTypes.MsgSend, 0),
		MsgMultiSend: make([]sqlDataTypes.MsgMultiSend, 0),
		MsgCall:   make([]sqlDataTypes.MsgCall, 0),
		MsgAddPkg: make([]sqlDataTypes.MsgAddPackage, 0),
		MsgRun:    make([]sqlDataTypes.MsgRun, 0),
	}
	for _, result := range msgResults {
		if result != nil {
			aggregatedDbGroups.MsgSend = append(aggregatedDbGroups.MsgSend, result.MsgSend...)
			aggregatedDbGroups.MsgMultiSend = append(aggregatedDbGroups.MsgMultiSend, result.MsgMultiSend...)
			aggregatedDbGroups.MsgCall = append(aggregatedDbGroups.MsgCall, result.MsgCall...)
			aggregatedDbGroups.MsgAddPkg = append(aggregatedDbGroups.MsgAddPkg, result.MsgAddPkg...)
			aggregatedDbGroups.MsgRun = append(aggregatedDbGroups.MsgRun, result.MsgRun...)
		}
	}

	addresses := createAddressTx(aggregatedDbGroups)
	timeout := 10*time.Second + (time.Duration(len(addresses)) * time.Second / 5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := d.dbPool.InsertAddressTx(ctx, addresses)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to insert address tx: %w", err)
	}

	if err := d.insertDbMessageGroups(aggregatedDbGroups); err != nil {
		return fmt.Errorf("failed to insert optimized messages: %w", err)
	}

	l.Info().Msgf("Messages processed from %d to %d: MsgSend=%d, MsgCall=%d, MsgAddPkg=%d, MsgRun=%d, MsgMultiSend=%d",
		fromHeight, toHeight,
		len(aggregatedDbGroups.MsgSend),
		len(aggregatedDbGroups.MsgCall),
		len(aggregatedDbGroups.MsgAddPkg),
		len(aggregatedDbGroups.MsgRun),
		len(aggregatedDbGroups.MsgMultiSend),
	)

	return nil
}

// transactionDecoding decodes all transactions and stores the decoded messages at the pre-allocated index.
func transactionDecoding(
	mu *sync.Mutex,
	transactions []TransactionsData,
	txCount int,
) ([]*decoder.DecodedMsg, map[string]struct{}) {
	decodedMsgs := make([]*decoder.DecodedMsg, txCount)
	addressesMap := make(map[string]struct{})
	wg := sync.WaitGroup{}
	wg.Add(txCount)

	for idx, transaction := range transactions {
		go decodeTx(mu, addressesMap, &wg, decodedMsgs, transaction, idx)
	}

	wg.Wait()

	return decodedMsgs, addressesMap
}

// decodeTx decodes a transaction and stores the decoded message at the pre-allocated index.
func decodeTx(
	mu *sync.Mutex,
	addressesMap map[string]struct{},
	wg *sync.WaitGroup,
	decodedMsgs []*decoder.DecodedMsg,
	transaction TransactionsData,
	idx int,
) {
	defer wg.Done()
	decodedMsg := decoder.NewDecodedMsg(transaction.Response.Result.Tx)
	if decodedMsg == nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"The transaction couldn't be decoded, tx hash: %s",
				transaction.Response.GetHash(),
			)
		return
	}
	// Collect addresses from this transaction and store in map[string]struct{}
	addresses := decodedMsg.CollectAllAddresses()
	mu.Lock()
	for _, address := range addresses {
		addressesMap[address] = struct{}{}
	}
	decodedMsgs[idx] = decodedMsg
	mu.Unlock()

}

// processMessageGroup converts a single transaction's messages into database-ready structs
// and stores the result at the pre-allocated index.
func (d *DataProcessor) processMessageGroup(
	idx int,
	transaction TransactionsData,
	decodedMsg *decoder.DecodedMsg,
	wg *sync.WaitGroup,
	results []*decoder.DbMessageGroups,
) {
	defer wg.Done()

	if decodedMsg == nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"The transaction couldn't be decoded, tx hash: %s",
				transaction.Response.GetHash(),
			)
		return
	}

	txId, ok := d.txHashCache[transaction.Response.GetHash()]
	if !ok {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Transaction hash not found in cache: %s", transaction.Response.GetHash(),
			)
		return
	}

	dbMessageGroups, err := decodedMsg.ConvertToDbMessages(
		d.addressCache, txId, d.chainName, transaction.Timestamp, decodedMsg.GetSigners(),
	)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to convert messages for tx %s: %v",
				transaction.Response.GetHash(),
				err,
			)
		return
	}

	results[idx] = dbMessageGroups
}

// insertDbMessageGroups performs optimized batch insertions using address IDs
func (d *DataProcessor) insertDbMessageGroups(groups *decoder.DbMessageGroups) error {
	var insertErrors = make([]error, 0)

	msgSendCount := len(groups.MsgSend)
	msgCallCount := len(groups.MsgCall)
	msgAddPkgCount := len(groups.MsgAddPkg)
	msgRunCount := len(groups.MsgRun)
	msgMultiSendCount := len(groups.MsgMultiSend)

	// Insert DbMsgSend messages with address IDs
	if msgSendCount > 0 {
		d.insertMessage(msgSendInserter(groups.MsgSend), msgSendCount, insertErrors)
	}

	// Insert DbMsgCall messages with address IDs
	if msgCallCount > 0 {
		d.insertMessage(msgCallInserter(groups.MsgCall), msgCallCount, insertErrors)
	}

	// Insert DbMsgAddPackage messages with address IDs
	if msgAddPkgCount > 0 {
		d.insertMessage(msgAddPackageInserter(groups.MsgAddPkg), msgAddPkgCount, insertErrors)
	}

	// Insert DbMsgRun messages with address IDs
	if msgRunCount > 0 {
		d.insertMessage(msgRunInserter(groups.MsgRun), msgRunCount, insertErrors)
	}

	// Insert DbMsgMultiSend messages with address IDs
	if msgMultiSendCount > 0 {
		d.insertMessage(msgMultiSendInserter(groups.MsgMultiSend), msgMultiSendCount, insertErrors)
	}

	// Combine all errors if any occurred
	if len(insertErrors) > 0 {
		var errorMessages []string
		for _, err := range insertErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("multiple insertion errors: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

func (d *DataProcessor)insertMessage(
	msgGroups messageInserter,
	msgCount int,
	errors []error,
) {
	timeout := 10*time.Second + (time.Duration(msgCount) * time.Second / 5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := msgGroups.insert(ctx, d.dbPool)
	if err != nil {
		txIds := msgGroups.getTxIds()
		hashes := d.finxHashes(txIds)
		errors = append(errors, fmt.Errorf("failed to insert messages: %w, hashes: %v", err, hashes))
	}
}

func (d *DataProcessor) ProcessValidatorSignings(
	commits []*rpcClient.CommitResponse,
	fromHeight uint64,
	toHeight uint64) {

	commitAmount := len(commits)
	validatorData := make([]sqlDataTypes.ValidatorBlockSigning, commitAmount)
	valid := make([]bool, commitAmount)
	wg := sync.WaitGroup{}
	wg.Add(commitAmount)

	for idx, commit := range commits {
		go d.processValidatorSigning(idx, commit, &wg, &valid[idx], validatorData)
	}

	wg.Wait()

	// Collect only the entries that were successfully processed
	result := make([]sqlDataTypes.ValidatorBlockSigning, 0, commitAmount)
	for idx, ok := range valid {
		if ok {
			result = append(result, validatorData[idx])
		}
	}

	timeout := 10*time.Second + (time.Duration(len(result)) * time.Second / 5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := d.dbPool.InsertValidatorBlockSignings(ctx, result)
	cancel()
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to insert validator commit signings: %v", err,
			)
	}
	l.Info().
		Msgf("Validator commit signings processed from %d to %d", fromHeight, toHeight)
}

// processValidatorSigning processes a single commit and stores it at a pre-allocated index.
func (d *DataProcessor) processValidatorSigning(
	idx int,
	commit *rpcClient.CommitResponse,
	wg *sync.WaitGroup,
	valid *bool,
	validatorData []sqlDataTypes.ValidatorBlockSigning,
) {
	defer wg.Done()

	proposer := d.validatorCache.GetAddress(commit.GetProposerAddress())
	precommits := commit.GetSigners()
	signedVals := make([]int32, 0, len(precommits))
	for _, precommit := range precommits {
		if precommit != nil {
			signedVals = append(signedVals, d.validatorCache.GetAddress(precommit.ValidatorAddress))
		}
	}

	height, err := commit.GetHeight()
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Msgf(
				"Failed to get commit height: %v", err,
			)
		return
	}

	validatorData[idx] = sqlDataTypes.ValidatorBlockSigning{
		BlockHeight: height,
		Timestamp:   commit.GetTimestamp(),
		Proposer:    proposer,
		SignedVals:  signedVals,
		ChainName:   d.chainName,
	}
	*valid = true
}


func (d *DataProcessor) finxHashes(txIds []int64) []string {
	hashes := make([]string, 0, len(txIds))
	for hash, id := range d.txHashCache {
		if slices.Contains(txIds, id) {
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

// createAddressTx builds a flat slice of AddressTx rows from all message groups.
func createAddressTx(msgGroups *decoder.DbMessageGroups) []sqlDataTypes.AddressTx {
	seen := make(map[key]sqlDataTypes.AddressTx)
	for _, m := range msgGroups.MsgSend {
		addToAddressTx(seen, m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName())
	}
	for _, m := range msgGroups.MsgCall {
		addToAddressTx(seen, m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName())
	}
	for _, m := range msgGroups.MsgAddPkg {
		addToAddressTx(seen, m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName())
	}
	for _, m := range msgGroups.MsgRun {
		addToAddressTx(seen, m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName())
	}
	for _, m := range msgGroups.MsgMultiSend {
		addToAddressTx(seen, m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName())
	}
	result := make([]sqlDataTypes.AddressTx, 0, len(seen))
	for _, e := range seen {
		result = append(result, e)
	}
	return result
}

func addToAddressTx(
	seen map[key]sqlDataTypes.AddressTx,
	addresses *sqlDataTypes.TxAddresses,
	chainName string,
	ts time.Time,
	msgType string,
) {
	for _, addr := range addresses.GetAddressList() {
		k := key{addr, addresses.TxId, chainName}
		if _, ok := seen[k]; !ok {
			seen[k] = sqlDataTypes.AddressTx{
				Address:   addr,
				TxId:      addresses.TxId,
				ChainName: chainName,
				Timestamp: ts,
			}
		}
	}
}
