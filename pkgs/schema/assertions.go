package schema

import dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"

// Assertion here is used as a sort of a linter for the schema.
// If some method or parameter is missing it should fail to compile.

// Every persisted table must describe itself for DDL generation.
var (
	_ DBTable = GnoAddress{}
	_ DBTable = GnoValidatorAddress{}
	_ DBTable = Blocks{}
	_ DBTable = ValidatorBlockSigning{}
	_ DBTable = AddressTx{}
	_ DBTable = TransactionGeneral{}
	_ DBTable = MsgSend{}
	_ DBTable = MsgMultiSend{}
	_ DBTable = MsgCall{}
	_ DBTable = MsgAddPackage{}
	_ DBTable = MsgRun{}
	_ DBTable = MsgAuthCrSession{}
	_ DBTable = MsgAuthRvSession{}
	_ DBTable = MsgAuthRvAllSessions{}
	_ DBTable = ApiKey{}
	_ DBTable = SchemaMigration{}
	_ DBTable = TxHashId{}
)

// Rows inserted through the generic COPY path must be Insertable.
var (
	_ Insertable = Blocks{}
	_ Insertable = ValidatorBlockSigning{}
	_ Insertable = TransactionGeneral{}
	_ Insertable = AddressTx{}
	_ Insertable = MsgSend{}
	_ Insertable = MsgMultiSend{}
	_ Insertable = MsgCall{}
	_ Insertable = MsgAddPackage{}
	_ Insertable = MsgRun{}
	_ Insertable = MsgAuthCrSession{}
	_ Insertable = MsgAuthRvSession{}
	_ Insertable = MsgAuthRvAllSessions{}
)

// Decoded transaction messages must be Message (Insertable + address reporting).
// GetAllAddresses has a pointer receiver, so these assert on the pointer type.
var (
	_ Message = (*MsgSend)(nil)
	_ Message = (*MsgMultiSend)(nil)
	_ Message = (*MsgCall)(nil)
	_ Message = (*MsgAddPackage)(nil)
	_ Message = (*MsgRun)(nil)
	_ Message = (*MsgAuthCrSession)(nil)
	_ Message = (*MsgAuthRvSession)(nil)
	_ Message = (*MsgAuthRvAllSessions)(nil)
)

// Composite postgres types must describe themselves for type creation.
var (
	_ DBSpecialType = Amount{}
	_ DBSpecialType = Attribute{}
	_ DBSpecialType = Event{}
)

// Continuous aggregate views are materialized views, not plain tables: they must
// satisfy ContinuousAggregateDefinition (mt/fn/gb-driven).
var (
	_ dbinit.ContinuousAggregateDefinition = TxCounter{}
	_ dbinit.ContinuousAggregateDefinition = FeeVolume{}
	_ dbinit.ContinuousAggregateDefinition = DailyActiveAccounts{}
	_ dbinit.ContinuousAggregateDefinition = ValidatorSigningCounter{}
	_ dbinit.ContinuousAggregateDefinition = BlockCounter{}
)
