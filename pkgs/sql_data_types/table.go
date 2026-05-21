package sql_data_types

import (
	"reflect"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

// TxAddresses groups all addresses involved in a single transaction
// It stores in a set like data structure to avoid duplicates
// all addresses for the same transaction hash together
type TxAddresses struct {
	TxId      int64
	Addresses map[int32]struct{}
}

// NewTxAddresses creates a new TxAddresses with the given transaction hash
func NewTxAddresses(txId int64) *TxAddresses {
	return &TxAddresses{
		TxId:      txId,
		Addresses: make(map[int32]struct{}),
	}
}

// AddAddress adds an address to the set
// this will probably overwrite the address if it already exists but there should be no duplicates
func (ta *TxAddresses) AddAddress(addressID int32) {
	ta.Addresses[addressID] = struct{}{}
}

// GetAddressList returns a slice of all address IDs.
// Returns a slice of all address IDs.
func (ta *TxAddresses) GetAddressList() []int32 {
	addresses := make([]int32, 0, len(ta.Addresses))
	for addr := range ta.Addresses {
		addresses = append(addresses, addr)
	}
	return addresses
}

// GnoAddress represents a regular Gno address with database mapping information
// Stores:
// - Address (string)
// - ID (int32)
// - Chain ID (string)
// PRIMARY KEY (id), UNIQUE (address, chain_id)
type GnoAddress struct {
	// any of the values can't be a null value and there shouldn't be any duplicates
	Address string `db:"address" dbtype:"TEXT" nullable:"false" primary:"false" unique:"true"`
	ID      int32  `db:"id" dbtype:"INTEGER GENERATED ALWAYS AS IDENTITY" nullable:"false" primary:"true"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false" unique:"true"`
}

// TableName returns the name of the table for the GnoAddress struct
func (g GnoAddress) TableName() string {
	return "gno_addresses"
}

// GetTableInfo returns the table info for the GnoAddress struct
func (g GnoAddress) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(g, g.TableName())
}

// GnoValidatorAddress represents a Gno validator address with database mapping information
// Same structure as GnoAddress but creates a separate table for validators
// Stores:
// - Address (string)
// - ID (int32)
// - Chain Name (string)
// PRIMARY KEY (id), UNIQUE (address, chain_name)
type GnoValidatorAddress struct {
	Address   string `db:"address" dbtype:"TEXT" nullable:"false" primary:"false" unique:"true"`
	ID        int32  `db:"id" dbtype:"INTEGER GENERATED ALWAYS AS IDENTITY" nullable:"false" primary:"true"`
	ChainName string `db:"chain_name" dbtype:"chain_name" nullable:"false" primary:"false" unique:"true"`
}

// TableName returns the name of the table for the GnoValidatorAddress struct
func (gv GnoValidatorAddress) TableName() string {
	return "gno_validators"
}

// GetTableInfo returns the table info for the GnoValidatorAddress struct
func (gv GnoValidatorAddress) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(gv, gv.TableName())
}

// DBTable is an interface for structs that represent database tables
type DBTable interface {
	GetTableInfo() (*dbinit.TableInfo, error)
	TableName() string
}

// An interface for Gno messages
//
// Methods:
// - TableColumns() []string: a method to get the columns of the struct
type GnoMessage interface {
	TableColumns() []string
}

type ApiKey struct {
	Id       string   `db:"id" dbtype:"UUID DEFAULT gen_random_uuid()" primary:"true" nullable:"false"`
	Prefix   string   `db:"prefix" dbtype:"VARCHAR(10)" nullable:"false" primary:"false" unique:"false"`
	Hash     [32]byte `db:"hash" dbtype:"BYTEA" nullable:"false" primary:"false" unique:"false"`
	Name     string   `db:"name" dbtype:"TEXT" nullable:"false" primary:"false" unique:"true"`
	RpmLimit int      `db:"rpm_limit" dbtype:"INT" nullable:"false" primary:"false"`
	IsActive bool     `db:"is_active" dbtype:"BOOLEAN DEFAULT TRUE" nullable:"false" primary:"false"`
}

func (ak ApiKey) TableName() string {
	return "api_keys"
}

func (ak ApiKey) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ak, ak.TableName())
}

func (ak ApiKey) TableColumns() []string {
	columns := make([]string, 0)
	fields := reflect.TypeOf(ak)
	for i := range fields.NumField() {
		field := fields.Field(i)
		columns = append(columns, field.Tag.Get("db"))
	}
	return columns
}

func AllTableNames() []string {
	tables := []DBTable{
		GnoAddress{},
		GnoValidatorAddress{},
		Blocks{},
		ValidatorBlockSigning{},
		AddressTx{},
		TransactionGeneral{},
		MsgSend{},
		MsgMultiSend{},
		MsgCall{},
		MsgAddPackage{},
		MsgRun{},
		ApiKey{},
	}
	names := make([]string, len(tables))
	for i, t := range tables {
		names[i] = t.TableName()
	}
	return names
}

func AllAggrTableNames() []string {
	tables := []DBTable{
		TxCounter{},
		FeeVolume{},
		DailyActiveAccounts{},
		ValidatorSigningCounter{},
		BlockCounter{},
	}
	names := make([]string, len(tables))
	for i, t := range tables {
		names[i] = t.TableName()
	}
	return names
}
