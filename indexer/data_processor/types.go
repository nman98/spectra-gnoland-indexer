package dataprocessor

import (
	"context"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	sqlDataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

// Define interface for what DataProcessor needs from database
type Database interface {
	InsertBlocks(ctx context.Context, blocks []sqlDataTypes.Blocks) error
	InsertValidatorBlockSignings(ctx context.Context, validatorBlockSignings []sqlDataTypes.ValidatorBlockSigning) error
	InsertTransactionsGeneral(ctx context.Context, transactionsGeneral []sqlDataTypes.TransactionGeneral) error
	InsertMsgSend(ctx context.Context, messages []sqlDataTypes.MsgSend) error
	InsertMsgCall(ctx context.Context, messages []sqlDataTypes.MsgCall) error
	InsertMsgAddPackage(ctx context.Context, messages []sqlDataTypes.MsgAddPackage) error
	InsertMsgRun(ctx context.Context, messages []sqlDataTypes.MsgRun) error
	InsertMsgMultiSend(ctx context.Context, messages []sqlDataTypes.MsgMultiSend) error
	InsertMsgAuthCrSession(ctx context.Context, messages []sqlDataTypes.MsgAuthCrSession) error
	InsertMsgAuthRvSession(ctx context.Context, messages []sqlDataTypes.MsgAuthRvSession) error
	InsertMsgAuthRvAllSessions(ctx context.Context, messages []sqlDataTypes.MsgAuthRvAllSessions) error
	InsertAddressTx(ctx context.Context, addresses []sqlDataTypes.AddressTx) error
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

type msgSendInserter []sqlDataTypes.MsgSend

func (m msgSendInserter) count() int {
	return len(m)
}
func (m msgSendInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgSend(ctx, []sqlDataTypes.MsgSend(m))
}

func (m msgSendInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgCallInserter []sqlDataTypes.MsgCall

func (m msgCallInserter) count() int {
	return len(m)
}
func (m msgCallInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgCall(ctx, []sqlDataTypes.MsgCall(m))
}

func (m msgCallInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAddPackageInserter []sqlDataTypes.MsgAddPackage

func (m msgAddPackageInserter) count() int {
	return len(m)
}
func (m msgAddPackageInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAddPackage(ctx, []sqlDataTypes.MsgAddPackage(m))
}
func (m msgAddPackageInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgRunInserter []sqlDataTypes.MsgRun

func (m msgRunInserter) count() int {
	return len(m)
}
func (m msgRunInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgRun(ctx, []sqlDataTypes.MsgRun(m))
}

func (m msgRunInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgMultiSendInserter []sqlDataTypes.MsgMultiSend

func (m msgMultiSendInserter) count() int {
	return len(m)
}
func (m msgMultiSendInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgMultiSend(ctx, []sqlDataTypes.MsgMultiSend(m))
}

func (m msgMultiSendInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthCrSessionInserter []sqlDataTypes.MsgAuthCrSession

func (m msgAuthCrSessionInserter) count() int { return len(m) }
func (m msgAuthCrSessionInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthCrSession(ctx, []sqlDataTypes.MsgAuthCrSession(m))
}
func (m msgAuthCrSessionInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthRvSessionInserter []sqlDataTypes.MsgAuthRvSession

func (m msgAuthRvSessionInserter) count() int { return len(m) }
func (m msgAuthRvSessionInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthRvSession(ctx, []sqlDataTypes.MsgAuthRvSession(m))
}
func (m msgAuthRvSessionInserter) getTxIds() []int64 {
	txIds := make([]int64, len(m))
	for i, msg := range m {
		txIds[i] = msg.TxId
	}
	return txIds
}

type msgAuthRvAllSessionsInserter []sqlDataTypes.MsgAuthRvAllSessions

func (m msgAuthRvAllSessionsInserter) count() int { return len(m) }
func (m msgAuthRvAllSessionsInserter) insert(ctx context.Context, db Database) error {
	return db.InsertMsgAuthRvAllSessions(ctx, []sqlDataTypes.MsgAuthRvAllSessions(m))
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
