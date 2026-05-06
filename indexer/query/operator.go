package query

import (
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/retry"
	rc "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

var l = logger.Get()

var (
	defaultRetryAmount        = 6
	defaultPause              = 3
	defaultPauseTime          = 15 * time.Second
	defaultExponentialBackoff = 2 * time.Second
)

// NewQueryOperator creates a new query operator
func NewQueryOperator(
	rpcClient RpcClient,
	retryAmount *int,
	pause *int,
	pauseTime *time.Duration,
	exponentialBackoff *time.Duration,
) *QueryOperator {
	if retryAmount == nil || *retryAmount == 0 {
		retryAmount = &defaultRetryAmount
	}
	if pause == nil || *pause == 0 {
		pause = &defaultPause
	}
	if pauseTime == nil || *pauseTime == 0 {
		pauseTime = &defaultPauseTime
	}
	if exponentialBackoff == nil || *exponentialBackoff == 0 {
		exponentialBackoff = &defaultExponentialBackoff
	}
	return &QueryOperator{
		rpcClient:          rpcClient,
		retryAmount:        *retryAmount,
		pause:              *pause,
		pauseTime:          *pauseTime,
		exponentialBackoff: *exponentialBackoff,
	}
}

// A swarm method to get blocks from a to b chain height inclusive
// This is a fan out method that launches async workers for each block and wait to get the results
// The order of the blocks is not guaranteed but it shouldn't matter because at the end of the process
// the indexer should store them all together as one huge slice of blocks, so the order is not important
// the speed is what matters here.
//
// Parameters:
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - []*rpcClient.BlockResponse: returns the slice of block responses
//
// The method will not throw an error if the block is missing, not found or there is some query error,
// it will just return nil for the block.
//
// Example:
//
//	var blocks []*rpcClient.BlockResponse
//	blocks = q.GetFromToBlocks(1, 50)
//	for _, block := range blocks {
//		fmt.Println(block.Height)
//	}
//
// The method will not throw an error if the block is missing, not found or there is some query error,
// it will just return nil for the block.
//
// Example:
//
//	var blocks []*rpcClient.BlockResponse
//	blocks = q.GetFromToBlocks(1, 50)
//	for _, block := range blocks {
//		fmt.Println(block.Height)
//	}
func (q *QueryOperator) GetFromToBlocks(fromHeight uint64, toHeight uint64) []*rc.BlockResponse {
	diff := toHeight - fromHeight + 1 // example from 1 to 50 means 50 blocks so +1 is needed because 100-51+1=50
	if diff < 1 {
		return nil
	}

	// Pre-allocate with exact size
	blocks := make([]*rc.BlockResponse, diff)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(int(diff))

	// Launch goroutines to get the blocks
	for i := range diff {
		height := fromHeight + i
		idx := i // Capture index
		go func(height uint64, idx int) {
			block, err := q.rpcClient.GetBlock(height)
			if err != nil {
				// Use retry mechanism with callback pattern
				retry.RetryWithContext(
					q.retryAmount,
					q.pause,
					q.pauseTime,
					q.exponentialBackoff,
					func(args ...any) (*rc.BlockResponse, error) {
						h := args[0].(uint64)
						result, rpcErr := q.rpcClient.GetBlock(h)
						if rpcErr != nil {
							return nil, rpcErr
						}
						return result, nil
					},
					func(result *rc.BlockResponse) {
						mu.Lock()
						blocks[idx] = result
						mu.Unlock()
						wg.Done()
					},
					func(retryErr error) {
						l.Error().
							Caller().
							Stack().
							Err(retryErr).
							Msgf("failed to get block %d after retries", height)
						mu.Lock()
						blocks[idx] = nil
						mu.Unlock()
						wg.Done()
					},
					height,
				)
				return
			}
			mu.Lock()
			blocks[idx] = block
			mu.Unlock()
			wg.Done()
		}(height, int(idx))
	}

	wg.Wait()
	return blocks
}

func (q *QueryOperator) GetFromToCommits(fromHeight uint64, toHeight uint64) []*rc.CommitResponse {
	diff := toHeight - fromHeight + 1
	if diff < 1 {
		return nil
	}

	commits := make([]*rc.CommitResponse, diff)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(int(diff))

	// Launch goroutines to get the commits
	for i := range diff {
		height := fromHeight + i
		idx := i // Capture index
		go func(height uint64, idx int) {
			commit, err := q.rpcClient.GetCommit(height)
			if err != nil {
				// Use retry mechanism with callback pattern
				retry.RetryWithContext(
					q.retryAmount,
					q.pause,
					q.pauseTime,
					q.exponentialBackoff,
					func(args ...any) (*rc.CommitResponse, error) {
						h := args[0].(uint64)
						result, rpcErr := q.rpcClient.GetCommit(h)
						if rpcErr != nil {
							return nil, rpcErr
						}
						return result, nil
					},
					func(result *rc.CommitResponse) {
						mu.Lock()
						commits[idx] = result
						mu.Unlock()
						wg.Done()
					},
					func(retryErr error) {
						l.Error().
							Caller().
							Stack().
							Err(retryErr).
							Msgf("failed to get commit %d after retries", height)
						mu.Lock()
						commits[idx] = nil
						mu.Unlock()
						wg.Done()
					},
					height,
				)
				return
			}
			mu.Lock()
			commits[idx] = commit
			mu.Unlock()
			wg.Done()
		}(height, int(idx))
	}

	wg.Wait()
	return commits
}

// A swarm method to get transactions from a slice of tx hashes
// This is a fan out method that lauches async workers for each tx and wait to get the resaults
// the indexer should store them all together as one huge slice of transactions,
//
// Parameters:
//   - txs: a slice of tx hashes
//
// Returns:
//   - []*rpcClient.TxResponse: returns the slice of transaction responses
//
// The method will not throw an error if the transaction is missing, not found or there is some query error,
// it will just return nil for the transaction.
//
// Example:
//
//	var transactions []*rpcClient.TxResponse
//	transactions = q.GetTransactions([]string{"tx_hash_1", "tx_hash_2", "tx_hash_3"})
//	for _, transaction := range transactions {
//		fmt.Println(transaction.Hash)
//	}
func (q *QueryOperator) GetTransactions(txs []string) []*rc.TxResponse {
	nTxs := len(txs)

	if nTxs < 1 {
		return nil
	}

	// Pre allocate with exact size
	transactions := make([]*rc.TxResponse, nTxs)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(nTxs)

	// Launch goroutines to get the transactions
	for idx, tx := range txs {
		go func(idx int, tx string) {
			txResponse, err := q.rpcClient.GetTx(tx)
			if err != nil {
				// Use retry mechanism with callback pattern
				retry.RetryWithContext(
					q.retryAmount,
					q.pause,
					q.pauseTime,
					q.exponentialBackoff,
					func(args ...any) (*rc.TxResponse, error) {
						txHash := args[0].(string)
						result, rpcErr := q.rpcClient.GetTx(txHash)
						if rpcErr != nil {
							return nil, rpcErr
						}
						return result, nil
					},
					func(result *rc.TxResponse) {
						mu.Lock()
						transactions[idx] = result
						mu.Unlock()
						wg.Done()
					},
					func(retryErr error) {
						l.Error().
							Caller().
							Stack().
							Err(retryErr).
							Msgf("failed to get tx %s after retries", tx)
						mu.Lock()
						transactions[idx] = nil
						mu.Unlock()
						wg.Done()
					},
					tx,
				)
				return
			}
			mu.Lock()
			transactions[idx] = txResponse
			mu.Unlock()
			wg.Done()
		}(idx, tx)
	}

	wg.Wait()
	return transactions
}

func (q *QueryOperator) GetLatestBlockHeight() (uint64, error) {
	result, err := q.rpcClient.GetLatestBlockHeight()
	if err != nil {
		return 0, err
	}
	return result, nil
}
