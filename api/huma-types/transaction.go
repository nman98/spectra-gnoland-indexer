package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type TransactionGetInput struct {
	// tx hash needs to be exactly 44 characters long
	TxHash string `path:"tx_hash" minLength:"44" maxLength:"44" doc:"Transaction hash (base64url encoded)" required:"true"`
}

// TransactionBasicGetOutput represents the response for basic transaction details
type TransactionBasicGetOutput struct {
	Body database.Transaction
}

// TransactionMessage represents a unified transaction message type that can be one of:
// bank_msg_send, vm_msg_call, vm_msg_add_package, or vm_msg_run
// not maybe the best implementation, but this one works for now
// to future me, if you figure out a better way to do this, please do so
// for now this is good enough
type TransactionMessage struct {
	// Common fields (always present)
	MessageType string    `json:"message_type" doc:"Type of message: bank_msg_send, vm_msg_call, vm_msg_add_package, or vm_msg_run" enum:"bank_msg_send,vm_msg_call,vm_msg_add_package,vm_msg_run"`
	TxHash      string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp   time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Signers     []string  `json:"signers" doc:"Signers (addresses)"`

	// BankSend specific fields
	FromAddress string            `json:"from_address,omitempty" doc:"From address (only for bank_msg_send)"`
	ToAddress   string            `json:"to_address,omitempty" doc:"To address (only for bank_msg_send)"`
	Amount      []database.Amount `json:"amount,omitempty" doc:"Amount (only for bank_msg_send)"`

	// MsgCall specific fields
	Caller   string `json:"caller,omitempty" doc:"Caller address (for vm_msg_call and vm_msg_run)"`
	FuncName string `json:"func_name,omitempty" doc:"Function name (only for vm_msg_call)"`
	Args     string `json:"args,omitempty" doc:"Arguments (only for vm_msg_call)"`

	// MsgAddPackage and MsgRun specific fields
	Creator      string   `json:"creator,omitempty" doc:"Creator address (only for vm_msg_add_package)"`
	PkgName      string   `json:"pkg_name,omitempty" doc:"Package name (for vm_msg_add_package and vm_msg_run)"`
	PkgFileNames []string `json:"pkg_file_names,omitempty" doc:"Package file names (for vm_msg_add_package and vm_msg_run)"`

	// Shared fields for vm_* messages
	PkgPath    string            `json:"pkg_path,omitempty" doc:"Package path (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`
	Send       []database.Amount `json:"send,omitempty" doc:"Send amount (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`
	MaxDeposit []database.Amount `json:"max_deposit,omitempty" doc:"Max deposit (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`
}

// TransactionMessageGetOutput represents the response containing all messages within a transaction
type TransactionMessageGetOutput struct {
	Body map[int16]TransactionMessage
}

type TransactionGeneralListByCursorGetInput struct {
	Cursor string `query:"cursor" doc:"Cursor to get the next set of transactions in form of timestamp|tx_hash(base64url encoded)" required:"false"`
	Limit  uint64 `query:"limit" doc:"Limit of transactions to get" required:"true" min:"1" max:"100" default:"10"`
}

type TransactionGeneralListByCursorGetOutput struct {
	Body []*database.Transaction
}

type LastXTransactionsGetInput struct {
	Amount uint64 `query:"amount" doc:"Amount of transactions to get" required:"true" min:"1" max:"100" default:"10"`
}

type LastXTransactionsGetOutput struct {
	Body []*database.Transaction
}

type TotalTxCount24hGetInput struct{}

type TotalTxCount24hGetOutput struct {
	Body int64
}

type TxCountByDateGetInput struct {
	StartDate date.Date `query:"start_date" doc:"Start date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	EndDate   date.Date `query:"end_date" doc:"End date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
}

type TxCountByDateGetOutput struct {
	Body []*database.TxCountDateRange
}

type TxCountByTimeGetOutput struct {
	Body []*database.TxCountTimeRange
}
type TxCountByHourGetInput struct {
	StartTimestamp time.Time `query:"start_timestamp" doc:"Start datetime (inclusive)" format:"date-time" required:"true"`
	EndTimestamp   time.Time `query:"end_timestamp" doc:"End datetime (inclusive)" format:"date-time" required:"true"`
}

type TxCountByHourGetOutput struct {
	Body []*database.TxCountTimeRange
}

type VolumeByDateGetInput struct {
	StartDate date.Date `query:"start_date" doc:"Start date (inclusive)" format:"date" required:"true"`
	EndDate   date.Date `query:"end_date" doc:"End date (inclusive)" format:"date" required:"true"`
}

type VolumeByDateGetOutput struct {
	Body database.VolumeByDenomDaily
}

type VolumeByHourGetInput struct {
	StartTimestamp time.Time `query:"start_timestamp" doc:"Start datetime (inclusive)" format:"date-time" required:"true"`
	EndTimestamp   time.Time `query:"end_timestamp" doc:"End datetime (inclusive)" format:"date-time" required:"true"`
}

type VolumeByHourGetOutput struct {
	Body database.VolumeByDenomHourly
}
