package sql_data_types

import (
	"reflect"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
	"github.com/jackc/pgx/v5/pgtype"
)

// TxHashId represents a transaction hash and its associated metadata
//
// Stores:
//   - TxId (bigserial)
//   - TxHash (bytea)
//   - Timestamp (timestamptz)
//   - ChainName (chain_name)
//
// PRIMARY KEY (tx_id, timestamp. chain_name)
type TxHashId struct {
	TxId      int64     `db:"tx_id" dbtype:"bigserial" nullable:"false" primary:"true"`
	TxHash    []byte    `db:"tx_hash" dbtype:"bytea" nullable:"false" primary:"false"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
}

func (t TxHashId) TableName() string {
	return "tx_hash_id"
}

func (t TxHashId) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(t, t.TableName())
}

func (t TxHashId) TableColumns() []string {
	fields := reflect.TypeOf(t)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// Blocks represents a blockchain block with database mapping information
//
// Stores:
//   - Hash (bytea)
//   - Height (bigint)
//   - Timestamp (timestamptz)
//   - Chain ID (text)
//   - Chain Name (chain_name)
//
// PRIMARY KEY (height, timestamp, chain_name)
type Blocks struct {
	Hash      []byte    `db:"hash" dbtype:"bytea" nullable:"false" primary:"false"`
	Height    uint64    `db:"height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainID   string    `db:"chain_id" dbtype:"TEXT" nullable:"false" primary:"false"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
}

func (b Blocks) TableName() string {
	return "blocks"
}

// GetTableInfo returns the table info for the Blocks struct
func (b Blocks) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(b, b.TableName())
}

func (b Blocks) TableColumns() []string {
	fields := reflect.TypeOf(b)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// ValidatorBlockSigning represents a validator block signing with database mapping information
// Stores:
//   - Block height (bigint)
//   - Timestamp (timestamptz)
//   - Proposer (integer)
//   - Signed validators (integer all of the validators that signed the block)
//   - Chain Name (chain_name)
//
// PRIMARY KEY (block_height, timestamp, chain_name)
type ValidatorBlockSigning struct {
	BlockHeight uint64    `db:"block_height" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	Proposer    int32     `db:"proposer" dbtype:"integer" nullable:"false" primary:"false"`
	SignedVals  []int32   `db:"signed_vals" dbtype:"integer[]" nullable:"false" primary:"false"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"` // use type enum chain_name from postgres
}

// TableName returns the name of the table for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) TableName() string {
	return "validator_block_signing"
}

// GetTableInfo returns the table info for the ValidatorBlockSigning struct
func (vbs ValidatorBlockSigning) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(vbs, vbs.TableName())
}

func (vbs ValidatorBlockSigning) TableColumns() []string {
	fields := reflect.TypeOf(vbs)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// AddressTx represents a transaction with database mapping information
//
// Stores:
// - Address (integer)
// - TxId (bigint)
// - Chain ID (text)
// - Timestamp (timestamptz)
// PRIMARY KEY (address, tx_id, chain_name, timestamp)
type AddressTx struct {
	Address   int32     `db:"address" dbtype:"INTEGER" nullable:"false" primary:"true"`
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
}

// TableName returns the name of the table for the AddressTx struct
func (at AddressTx) TableName() string {
	return "address_tx"
}

// GetTableInfo returns the table info for the AddressTx struct
func (at AddressTx) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(at, at.TableName())
}
func (at AddressTx) TableColumns() []string {
	fields := reflect.TypeOf(at)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// TransactionGeneral represents a transaction general data with database mapping information
//
// Stores:
// - TxId (bigint)
// - ChainName (chain_name)
// - Timestamp (timestamptz)
// - MsgTypes (text[])
// - TxEvents (event[])
// - TxEventsCompressed (bytea)
// - GasUsed (bigint)
// - GasWanted (bigint)
// - Fee (fee)
// - Success (boolean)
//
// PRIMARY KEY (tx_id, chain_name, timestamp)
type TransactionGeneral struct {
	TxId        int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	ChainName   string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	Timestamp   time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	BlockHeight uint64    `db:"block_height" dbtype:"bigint" nullable:"false" primary:"false"`
	MsgTypes    []string  `db:"msg_types" dbtype:"TEXT[]" nullable:"false" primary:"false"`
	// Tx events in the future there should be an option to have this compressed
	// for now only store the native format but keep the option to have it compressed
	TxEvents           []Event        `db:"tx_events" dbtype:"event[]" nullable:"true" primary:"false"`
	TxEventsCompressed []byte         `db:"tx_events_compressed" dbtype:"bytea" nullable:"true" primary:"false"`
	CompressionOn      bool           `db:"compression_on" dbtype:"boolean" nullable:"false" primary:"false"`
	GasUsed            uint64         `db:"gas_used" dbtype:"bigint" nullable:"false" primary:"false"`
	GasWanted          uint64         `db:"gas_wanted" dbtype:"bigint" nullable:"false" primary:"false"`
	FeeAmount          pgtype.Numeric `db:"fee_amount" dbtype:"numeric" nullable:"false" primary:"false"`
	FeeDenom           string         `db:"fee_denom" dbtype:"text" nullable:"false" primary:"false"`
	Success            bool           `db:"success" dbtype:"boolean" nullable:"false" primary:"false" default:"false"`
	// Stored only if there is error
	ErrorLog *string `db:"error_log" dbtype:"TEXT" nullable:"true" primary:"false"`
}

// TableName returns the name of the table for the TransactionGeneral struct
func (tg TransactionGeneral) TableName() string {
	return "transaction_general"
}

// GetTableInfo returns the table info for the TransactionGeneral struct
func (tg TransactionGeneral) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(tg, tg.TableName())
}

func (tg TransactionGeneral) TableColumns() []string {
	fields := reflect.TypeOf(tg)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

func (tg *TransactionGeneral) GetMessageTypes() []string {
	return tg.MsgTypes
}

// MsgSend represents a bank send message
//
// Stores:
// - TxId (bigint)
// - Timestamp (timestamptz)
// - ChainName (chain_name)
// - FromAddress (integer)
// - ToAddress (integer)
// - Amount (amount[])
// - Signers (integer[])
// - MessageCounter (smallint)
//
// PRIMARY KEY (tx_id, chain_name, timestamp, message_counter)
type MsgSend struct {
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	FromAddress int32 `db:"from_address" dbtype:"INTEGER" nullable:"false" primary:"false"`
	// gno address, pull from the gno_addresses table
	ToAddress      int32    `db:"to_address" dbtype:"INTEGER" nullable:"true" primary:"false"`
	Amount         []Amount `db:"amount" dbtype:"amount[]" nullable:"false" primary:"false"`
	Signers        []int32  `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16    `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

// TableName returns the name of the table for the MsgSend struct
func (ms MsgSend) TableName() string {
	return "bank_msg_send"
}

// GetTableInfo returns the table info for the MsgSend struct
func (ms MsgSend) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ms, ms.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (ms MsgSend) TableColumns() []string {
	fields := reflect.TypeOf(ms)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// GetAllAddresses returns all the addresses that are involved in the message
// it will return the from address, to address, and signers in a single TxAddresses struct.
// This prevents duplicates by grouping all addresses for the same transaction.
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (ms *MsgSend) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(ms.TxId)
	txAddresses.AddAddress(ms.FromAddress)
	if ms.ToAddress != 0 {
		txAddresses.AddAddress(ms.ToAddress)
	}
	for _, address := range ms.Signers {
		txAddresses.AddAddress(address)
	}
	return txAddresses
}

// MsgCall represents a VM function call message
//
// Stores:
//   - TxId (bigint)
//   - Timestamp (timestamptz)
//   - ChainName (chain_name)
//   - Caller (integer)
//   - PkgPath (text)
//   - FuncName (text)
//   - Args (text)
//   - Send (amount[])
//   - MaxDeposit (amount[])
//   - Signers (integer[])
//   - MessageCounter (smallint)
//
// PRIMARY KEY (tx_id, chain_name, timestamp, message_counter)
type MsgCall struct {
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Caller         int32    `db:"caller" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath        string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	FuncName       string   `db:"func_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	Args           string   `db:"args" dbtype:"TEXT" nullable:"true" primary:"false"`
	Send           []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit     []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	Signers        []int32  `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16    `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

func (mc MsgCall) TableName() string {
	return "vm_msg_call"
}

// GetTableInfo returns the table info for the MsgCall struct
func (mc MsgCall) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(mc, mc.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (mc MsgCall) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(mc)
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the caller and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (mc *MsgCall) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(mc.TxId)
	txAddresses.AddAddress(mc.Caller)
	for _, addr := range mc.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// MsgAddPackage represents a VM package addition message
//
// Stores:
// - TxId (bigint)
// - ChainName (chain_name)
// - Creator (text)
// - PkgPath (text)
// - PkgName (text)
// - PkgFileNames (text[])
// - Send (amount[])
// - MaxDeposit (amount[])
// - Signers (integer[])
// - Timestamp (timestamptz)
// - MessageCounter (smallint)
//
// PRIMARY KEY (tx_id, chain_name, timestamp)
type MsgAddPackage struct {
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Creator      int32    `db:"creator" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath      string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgName      string   `db:"pkg_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgFileNames []string `db:"pkg_file_names" dbtype:"TEXT[]" nullable:"true" primary:"false"`
	Send         []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit   []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	// signers are the addresses that signed the transaction
	Signers        []int32 `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16   `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

func (ma MsgAddPackage) TableName() string {
	return "vm_msg_add_package"
}

// GetTableInfo returns the table info for the MsgAddPackage struct
func (ma MsgAddPackage) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ma, ma.TableName())
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (ma MsgAddPackage) TableColumns() []string {
	fields := reflect.TypeOf(ma)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the creator and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (ma *MsgAddPackage) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(ma.TxId)
	txAddresses.AddAddress(ma.Creator)
	for _, addr := range ma.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// MsgRun represents a VM package run message
//
// Stores:
// - TxId (bigint)
// - Timestamp (timestamptz)
// - ChainName (text)
// - Caller (integer)
// - PkgPath (text)
// - PkgName (text)
// - PkgFileNames (text[])
// - Send (amount[])
// - MaxDeposit (amount[])
// - Signers (integer[])
// - MessageCounter (smallint)
//
// PRIMARY KEY (tx_id, chain_name, timestamp)
type MsgRun struct {
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"true"`
	// gno address, pull from the gno_addresses table
	Caller       int32    `db:"caller" dbtype:"INTEGER" nullable:"false" primary:"false"`
	PkgPath      string   `db:"pkg_path" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgName      string   `db:"pkg_name" dbtype:"TEXT" nullable:"true" primary:"false"`
	PkgFileNames []string `db:"pkg_file_names" dbtype:"TEXT[]" nullable:"true" primary:"false"`
	Send         []Amount `db:"send" dbtype:"amount[]" nullable:"true" primary:"false"`
	MaxDeposit   []Amount `db:"max_deposit" dbtype:"amount[]" nullable:"true" primary:"false"`
	// signers are the addresses that signed the transaction
	Signers        []int32 `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16   `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (mr MsgRun) TableColumns() []string {
	columns := make([]string, 0)
	// get the fields of the struct
	fields := reflect.TypeOf(mr)
	numFields := fields.NumField()
	for i := range numFields {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

func (mr MsgRun) TableName() string {
	return "vm_msg_run"
}

// GetTableInfo returns the table info for the MsgRun struct
func (mr MsgRun) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(mr, mr.TableName())
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the caller and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (mr *MsgRun) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(mr.TxId)
	txAddresses.AddAddress(mr.Caller)
	for _, addr := range mr.Signers {
		txAddresses.AddAddress(addr)
	}
	return txAddresses
}

// MsgMultiSend represents a multi-send bank module message.
// It has a direction field to indicate if this is an output or input entry.
// This table will only store one half of the response because the message can have multiple entries per
// input/output for every address.
//
// Stores:
//   - TxId: transaction ID
//   - Timestamp: timestamp of the transaction
//   - Direction: true for output, false for input
//   - AddressId: address ID
//   - Coins: amount of coins
//
// PRIMARY KEY (tx_id, timestamp, chain_name, direction, address_id, message_counter)
type MsgMultiSend struct {
	TxId      int64     `db:"tx_id" dbtype:"bigint" nullable:"false" primary:"true"`
	Timestamp time.Time `db:"timestamp" dbtype:"timestamptz" nullable:"false" primary:"true"`
	ChainName string    `db:"chain_name" dbtype:"text" nullable:"false" primary:"true"`
	// By direction it refers to if this part is output entry or input entry.
	// True means output, false means input. It has primary to true
	Direction      bool     `db:"direction" dbtype:"boolean" nullable:"false" primary:"true"`
	AddressId      int32    `db:"address_id" dbtype:"integer" nullable:"false" primary:"true"`
	Coins          []Amount `db:"coins" dbtype:"amount" nullable:"false" primary:"false"`
	Signers        []int32  `db:"signers" dbtype:"INTEGER[]" nullable:"false" primary:"false"`
	MessageCounter int16    `db:"message_counter" dbtype:"smallint" nullable:"false" primary:"true"`
}

func (m MsgMultiSend) TableName() string {
	return "bank_msg_multi_send"
}

func (m MsgMultiSend) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(m, m.TableName())
}

// GetAllAddresses returns all the addresses that are involved in the message
// Groups the caller and signers for this transaction
//
// Returns:
//   - *TxAddresses: grouped addresses for this transaction
func (m *MsgMultiSend) GetAllAddresses() *TxAddresses {
	txAddresses := NewTxAddresses(m.TxId)
	txAddresses.AddAddress(m.AddressId)
	return txAddresses
}

// A method to get the columns of the struct
// Useful in GnoMessage interface
func (m MsgMultiSend) TableColumns() []string {
	fields := reflect.TypeOf(m)
	numFields := fields.NumField()
	columns := make([]string, numFields)
	for i := range numFields {
		field := fields.Field(i)
		columns[i] = field.Tag.Get("db")
	}
	return columns
}
