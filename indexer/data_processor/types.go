package dataprocessor

import (
	"context"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
)

// Define interface for what DataProcessor needs from database
type Database interface {
	InsertBlocks(ctx context.Context, blocks []s.Blocks) error
	InsertValidatorBlockSignings(ctx context.Context, validatorBlockSignings []s.ValidatorBlockSigning) error
	InsertTransactionsGeneral(ctx context.Context, transactionsGeneral []s.TransactionGeneral) error
	InsertMsgSend(ctx context.Context, messages []s.MsgSend) error
	InsertMsgCall(ctx context.Context, messages []s.MsgCall) error
	InsertMsgAddPackage(ctx context.Context, messages []s.MsgAddPackage) error
	InsertMsgRun(ctx context.Context, messages []s.MsgRun) error
	InsertMsgMultiSend(ctx context.Context, messages []s.MsgMultiSend) error
	InsertMsgAuthCrSession(ctx context.Context, messages []s.MsgAuthCrSession) error
	InsertMsgAuthRvSession(ctx context.Context, messages []s.MsgAuthRvSession) error
	InsertMsgAuthRvAllSessions(ctx context.Context, messages []s.MsgAuthRvAllSessions) error
	InsertAddressTx(ctx context.Context, addresses []s.AddressTx) error
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

type msgSendInserter []s.MsgSend

func (m msgSendInserter) count() int {
	return len(m)
}
func (m msgSendInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgSend(ctx, []s.MsgSend(m))
}

func (m msgSendInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgCallInserter []s.MsgCall

func (m msgCallInserter) count() int {
	return len(m)
}
func (m msgCallInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgCall(ctx, []s.MsgCall(m))
}

func (m msgCallInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAddPackageInserter []s.MsgAddPackage

func (m msgAddPackageInserter) count() int {
	return len(m)
}
func (m msgAddPackageInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAddPackage(ctx, []s.MsgAddPackage(m))
}
func (m msgAddPackageInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgRunInserter []s.MsgRun

func (m msgRunInserter) count() int {
	return len(m)
}
func (m msgRunInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgRun(ctx, []s.MsgRun(m))
}

func (m msgRunInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgMultiSendInserter []s.MsgMultiSend

func (m msgMultiSendInserter) count() int {
	return len(m)
}
func (m msgMultiSendInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgMultiSend(ctx, []s.MsgMultiSend(m))
}

func (m msgMultiSendInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthCrSessionInserter []s.MsgAuthCrSession

func (m msgAuthCrSessionInserter) count() int { return len(m) }
func (m msgAuthCrSessionInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthCrSession(ctx, []s.MsgAuthCrSession(m))
}
func (m msgAuthCrSessionInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthRvSessionInserter []s.MsgAuthRvSession

func (m msgAuthRvSessionInserter) count() int { return len(m) }
func (m msgAuthRvSessionInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthRvSession(ctx, []s.MsgAuthRvSession(m))
}
func (m msgAuthRvSessionInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthRvAllSessionsInserter []s.MsgAuthRvAllSessions

func (m msgAuthRvAllSessionsInserter) count() int { return len(m) }
func (m msgAuthRvAllSessionsInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthRvAllSessions(ctx, []s.MsgAuthRvAllSessions(m))
}
func (m msgAuthRvAllSessionsInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

// Interface for message inserter
type messageInserter interface {
	insert(ctx context.Context, db Database) error
	count() int
	getTxIds() []int64
}
