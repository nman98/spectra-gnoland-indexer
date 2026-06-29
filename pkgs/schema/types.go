package schema

import (
	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
	"github.com/jackc/pgx/v5/pgtype"
)

type Amount struct {
	Amount pgtype.Numeric `db:"amount" dbtype:"NUMERIC"`
	Denom  string         `db:"denom" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Amount struct
func (amt Amount) TypeName() string {
	return "amount"
}

// GetSpecialTypeInfo returns the special type info for the Amount struct
func (amt Amount) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(amt, amt.TypeName())
}

type Attribute struct {
	Key   string `db:"key" dbtype:"TEXT"`
	Value string `db:"value" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Attribute struct
func (att Attribute) TypeName() string {
	return "attribute"
}

// GetSpecialTypeInfo returns the special type info for the Attribute struct
func (att Attribute) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(att, att.TypeName())
}

type Event struct {
	AtType     string      `db:"at_type" dbtype:"TEXT"`
	Type       string      `db:"type" dbtype:"TEXT"`
	Attributes []Attribute `db:"attributes" dbtype:"attribute[]"`
	PkgPath    string      `db:"pkg_path" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Event struct
func (e Event) TypeName() string {
	return "event"
}

// GetSpecialTypeInfo returns the special type info for the Event struct
func (e Event) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(e, e.TypeName())
}

type DataTypes interface {
	TableName(valTable bool) string
	TypeName() string
	GetTableInfo() (*dbinit.TableInfo, error)
}

// DBSpecialType is an interface for structs that represent custom database types
type DBSpecialType interface {
	GetSpecialTypeInfo() (*dbinit.SpecialType, error)
	TypeName() string
}

// Returns the names of every custom type
// to be used to register the types with pgx
func CustomTypeNames() []string {
	return []string{
		// Amount
		Amount{}.TypeName(),
		// Amount[]
		string("_" + Amount{}.TypeName()),
		// Attribute
		Attribute{}.TypeName(),
		// Attribute[]
		string("_" + Attribute{}.TypeName()),
		// Event
		Event{}.TypeName(),
		// Event[]
		string("_" + Event{}.TypeName()),
		// chain_name it is a type enum so it doesn't have it's struct type
		"chain_name",
	}
}
