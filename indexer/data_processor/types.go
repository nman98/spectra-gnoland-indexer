package dataprocessor

import (
	"context"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
)

// Define interface for what DataProcessor needs from database
type Database interface {
	// InsertRows inserts a homogeneous batch of rows (all the same table) via COPY FROM.
	InsertRows(ctx context.Context, rows []s.Insertable) error
	InsertTxHashIds(ctx context.Context, txHashes []string, timestamps []time.Time, chainName string) (map[string]int64, error)
}

// Define interface for what DataProcessor needs from AddressCache
type AddressCache interface {
	AddressSolver(address []string, chainName string, insertValidators bool, retryAttempts uint8, oneByOne *bool)
	GetAddress(address string) int32
}

type DataProcessor struct {
	dbPool         Database
	addressCache   AddressCache
	validatorCache AddressCache
	chainName      string
	txHashCache    map[string]int64
}

type TransactionsData struct {
	Response    *rpcClient.TxResponse
	Timestamp   time.Time
	BlockHeight uint64
}

func (t *TransactionsData) GetSuccess() bool {
	return t.Response.Result.TxResult.ResponseBase.Error == nil
}

func (t *TransactionsData) GetTransactionErrorDetails() *string {
	if t.Response.Result.TxResult.ResponseBase.Log == "" {
		return nil
	}
	// Limit the log to 255 characters to avoid overflow.
	var log string
	if len(t.Response.Result.TxResult.ResponseBase.Log) > 255 {
		log = t.Response.Result.TxResult.ResponseBase.Log[:255]
	} else {
		log = t.Response.Result.TxResult.ResponseBase.Log
	}
	return &log
}

// Internal types for address tx mapping
type key struct {
	address   int32
	txId      int64
	chainName string
}

// messageBatch is a homogeneous batch of message rows (all the same table) paired
// with the transaction ids they belong to, retained so a failed insert can be
// traced back to tx hashes for diagnostics.
type messageBatch struct {
	rows  []s.Insertable
	txIds []int64
}

// newMessageBatch boxes a typed message slice into a messageBatch, capturing each
// row's tx id via the supplied accessor while the concrete type is still known.
func newMessageBatch[T s.Insertable](rows []T, txID func(T) int64) messageBatch {
	batch := messageBatch{
		rows:  make([]s.Insertable, len(rows)),
		txIds: make([]int64, len(rows)),
	}
	for i, r := range rows {
		batch.rows[i] = r
		batch.txIds[i] = txID(r)
	}
	return batch
}
