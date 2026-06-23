package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type TransactionGetInput struct {
	// tx hash needs to be exactly 44 characters long
	TxHash string `path:"tx_hash" minLength:"44" maxLength:"44" doc:"Transaction hash (base64 encoded; base64url is also accepted)" required:"true"`
}

// TransactionBasicGetOutput represents the response for basic transaction details
type TransactionBasicGetOutput struct {
	Body database.Transaction
}

type MultiSendEntry struct {
	Address string            `json:"address" doc:"Address"`
	Coins   []database.Amount `json:"coins" doc:"Coins"`
}

// TransactionMessage represents a unified transaction message type.
type TransactionMessage struct {
	// Common fields (always present)
	MessageType string    `json:"message_type" doc:"Type of message" enum:"bank_msg_send,bank_msg_multi_send,vm_msg_call,vm_msg_add_package,vm_msg_run,auth_msg_create_session,auth_msg_revoke_session,auth_msg_revoke_all_sessions"`
	TxHash      string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp   time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Signers     []string  `json:"signers" doc:"Signers (addresses)"`

	// BankSend specific fields
	FromAddress string            `json:"from_address,omitempty" doc:"From address (only for bank_msg_send)"`
	ToAddress   string            `json:"to_address,omitempty" doc:"To address (only for bank_msg_send)"`
	Amount      []database.Amount `json:"amount,omitempty" doc:"Amount (only for bank_msg_send)"`

	// BankMultiSend specific fields
	Inputs  []MultiSendEntry `json:"inputs,omitempty" doc:"Input entries (only for bank_msg_multi_send)"`
	Outputs []MultiSendEntry `json:"outputs,omitempty" doc:"Output entries (only for bank_msg_multi_send)"`

	// MsgCall specific fields
	Caller   string `json:"caller,omitempty" doc:"Caller address (for vm_msg_call and vm_msg_run)"`
	FuncName string `json:"func_name,omitempty" doc:"Function name (only for vm_msg_call)"`
	Args     string `json:"args,omitempty" doc:"Arguments (only for vm_msg_call)"`

	// MsgAddPackage, MsgRun, and auth message specific fields
	Creator      string   `json:"creator,omitempty" doc:"Creator address (for vm_msg_add_package and auth messages)"`
	PkgName      string   `json:"pkg_name,omitempty" doc:"Package name (for vm_msg_add_package and vm_msg_run)"`
	PkgFileNames []string `json:"pkg_file_names,omitempty" doc:"Package file names (for vm_msg_add_package and vm_msg_run)"`

	// Shared fields for vm_* messages
	PkgPath    string            `json:"pkg_path,omitempty" doc:"Package path (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`
	Send       []database.Amount `json:"send,omitempty" doc:"Send amount (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`
	MaxDeposit []database.Amount `json:"max_deposit,omitempty" doc:"Max deposit (for vm_msg_call, vm_msg_add_package, and vm_msg_run)"`

	// Auth message specific fields
	SessionKey  string            `json:"session_key,omitempty" doc:"Session key hex (for auth_msg_create_session and auth_msg_revoke_session)"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" doc:"Session expiry (only for auth_msg_create_session)"`
	SpendLimit  []database.Amount `json:"spend_limit,omitempty" doc:"Spend limit (only for auth_msg_create_session)"`
	SpendPeriod *int64            `json:"spend_period,omitempty" doc:"Spend period in seconds; 0 means infinite (only for auth_msg_create_session)"`
}

// TransactionMessageGetOutput represents the response containing all messages within a transaction
type TransactionMessageGetOutput struct {
	Body map[int16]TransactionMessage
}

type TransactionGeneralListByCursorGetInput struct {
	Cursor    string             `query:"cursor" doc:"Cursor in the form '<block_height>|<tx_hash_base64url>'. Omit to fetch the most recent page." required:"false"`
	Limit     uint64             `query:"limit" doc:"Number of transactions to return" required:"false" min:"1" max:"100" default:"25"`
	Direction database.Direction `query:"direction" doc:"Direction to walk from the cursor. 'next' returns older rows; 'prev' returns newer rows and requires a cursor." enum:"next,prev" default:"next"`
}

// TransactionsRangeBody is the response body for cursor-based transaction
// listings. Transactions are always returned newest-first. NextCursor points at
// the oldest row on the page (use it to load an older page) while PrevCursor
// points at the newest row (use it to load a newer page).
type TransactionsRangeBody struct {
	Transactions []*database.Transaction `json:"transactions" doc:"Transactions on this page, newest-first"`
	HasNext      bool                    `json:"has_next" doc:"True when an older page exists"`
	HasPrev      bool                    `json:"has_prev" doc:"True when a newer page exists"`
	NextCursor   *string                 `json:"next_cursor" doc:"Cursor to request the next (older) page"`
	PrevCursor   *string                 `json:"prev_cursor" doc:"Cursor to request the previous (newer) page"`
}

type TransactionGeneralListByCursorGetOutput struct {
	Body TransactionsRangeBody
}

type LastXTransactionsGetInput struct {
	Amount uint64 `query:"amount" doc:"Amount of transactions to get" required:"true" min:"1" max:"100" default:"10"`
}

type LastXTransactionsGetOutput struct {
	Body []*database.Transaction
}

type TotalTxCount24hGetInput struct{}
type TotalTxCount24hBody struct {
	Count int64 `json:"count" doc:"The count of tx that occured in the last 24 hours"`
}
type TotalTxCount24hGetOutput struct {
	Body *TotalTxCount24hBody
}

type TxCountByDateGetInput struct {
	StartDate date.Date          `query:"start_date" doc:"Start date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	EndDate   date.Date          `query:"end_date" doc:"End date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	SortOrder database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type TxCountByDateGetOutput struct {
	Body []*database.TxCountDateRange
}

type TxCountByTimeGetOutput struct {
	Body []*database.TxCountTimeRange
}
type TxCountByHourGetInput struct {
	StartTimestamp time.Time          `query:"start_timestamp" doc:"Start datetime (inclusive)" format:"date-time" required:"true"`
	EndTimestamp   time.Time          `query:"end_timestamp" doc:"End datetime (inclusive)" format:"date-time" required:"true"`
	SortOrder      database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type TxCountByHourGetOutput struct {
	Body []*database.TxCountTimeRange
}

type VolumeByDateGetInput struct {
	StartDate date.Date          `query:"start_date" doc:"Start date (inclusive)" format:"date" required:"true"`
	EndDate   date.Date          `query:"end_date" doc:"End date (inclusive)" format:"date" required:"true"`
	SortOrder database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type VolumeByDateGetOutput struct {
	Body database.VolumeByDenomDaily
}

type VolumeByHourGetInput struct {
	StartTimestamp time.Time          `query:"start_timestamp" doc:"Start datetime (inclusive)" format:"date-time" required:"true"`
	EndTimestamp   time.Time          `query:"end_timestamp" doc:"End datetime (inclusive)" format:"date-time" required:"true"`
	SortOrder      database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type VolumeByHourGetOutput struct {
	Body database.VolumeByDenomHourly
}
