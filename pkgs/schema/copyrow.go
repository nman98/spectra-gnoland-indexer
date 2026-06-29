package schema

import "github.com/jackc/pgx/v5/pgtype"

// Insertable is implemented by every struct inserted through the generic COPY FROM
// path. A slice of Insertable handed to the database layer must be homogeneous
// (all the same table); the layer reads the table name and columns from the first
// element. TestCopyRowMatchesColumns guarantees CopyRow stays aligned with TableColumns.
type Insertable interface {
	TableName() string
	TableColumns() []string
	CopyRow() []any
}

// AsInsertable boxes a typed slice into []Insertable for the generic insert path.
func AsInsertable[T Insertable](rows []T) []Insertable {
	out := make([]Insertable, len(rows))
	for i, r := range rows {
		out[i] = r
	}
	return out
}

// pgArray converts a Go slice into a pgx typed array for COPY FROM inserts.
// Handles composite types and scalar arrays; a nil slice becomes a NULL array.
func pgArray[T any](v []T) pgtype.Array[T] {
	if v == nil {
		return pgtype.Array[T]{Valid: false}
	}
	return pgtype.Array[T]{
		Elements: v,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}},
		Valid:    true,
	}
}

// CopyRow returns the column values for a single row in the exact order reported
// by TableColumns. The two MUST stay aligned; TestCopyRowMatchesColumns enforces
// this so a struct field reorder can never silently misalign a COPY insert.

func (b Blocks) CopyRow() []any {
	return []any{
		b.Hash,
		b.Height,
		b.Timestamp,
		b.ChainID,
		b.ChainName,
	}
}

func (vbs ValidatorBlockSigning) CopyRow() []any {
	return []any{
		vbs.BlockHeight,
		vbs.Timestamp,
		vbs.Proposer,
		pgArray(vbs.SignedVals),
		vbs.ChainName,
	}
}

func (tg TransactionGeneral) CopyRow() []any {
	return []any{
		tg.TxId,
		tg.ChainName,
		tg.Timestamp,
		tg.BlockHeight,
		pgArray(tg.MsgTypes),
		pgArray(tg.TxEvents),
		tg.TxEventsCompressed,
		tg.CompressionOn,
		tg.GasUsed,
		tg.GasWanted,
		tg.FeeAmount,
		tg.FeeDenom,
		tg.Success,
		tg.ErrorLog,
	}
}

func (at AddressTx) CopyRow() []any {
	return []any{
		at.Address,
		at.TxId,
		at.ChainName,
		at.Timestamp,
	}
}

func (ms MsgSend) CopyRow() []any {
	return []any{
		ms.TxId,
		ms.Timestamp,
		ms.ChainName,
		ms.FromAddress,
		ms.ToAddress,
		pgArray(ms.Amount),
		pgArray(ms.Signers),
		ms.MessageCounter,
	}
}

func (mms MsgMultiSend) CopyRow() []any {
	return []any{
		mms.TxId,
		mms.Timestamp,
		mms.ChainName,
		mms.Direction,
		mms.AddressId,
		pgArray(mms.Coins),
		pgArray(mms.Signers),
		mms.MessageCounter,
	}
}

func (mc MsgCall) CopyRow() []any {
	return []any{
		mc.TxId,
		mc.Timestamp,
		mc.ChainName,
		mc.Caller,
		mc.PkgPath,
		mc.FuncName,
		mc.Args,
		pgArray(mc.Send),
		pgArray(mc.MaxDeposit),
		pgArray(mc.Signers),
		mc.MessageCounter,
	}
}

func (map_ MsgAddPackage) CopyRow() []any {
	return []any{
		map_.TxId,
		map_.Timestamp,
		map_.ChainName,
		map_.Creator,
		map_.PkgPath,
		map_.PkgName,
		pgArray(map_.PkgFileNames),
		pgArray(map_.Send),
		pgArray(map_.MaxDeposit),
		pgArray(map_.Signers),
		map_.MessageCounter,
	}
}

func (mr MsgRun) CopyRow() []any {
	return []any{
		mr.TxId,
		mr.Timestamp,
		mr.ChainName,
		mr.Caller,
		mr.PkgPath,
		mr.PkgName,
		pgArray(mr.PkgFileNames),
		pgArray(mr.Send),
		pgArray(mr.MaxDeposit),
		pgArray(mr.Signers),
		mr.MessageCounter,
	}
}

func (cs MsgAuthCrSession) CopyRow() []any {
	return []any{
		cs.TxId,
		cs.Timestamp,
		cs.ChainName,
		cs.Creator,
		cs.SessionKey,
		cs.ExpiresAt,
		pgArray(cs.SpendLimit),
		pgArray(cs.AllowPaths),
		cs.SpendPeriod,
		pgArray(cs.Signers),
		cs.MessageCounter,
	}
}

func (rs MsgAuthRvSession) CopyRow() []any {
	return []any{
		rs.TxId,
		rs.Timestamp,
		rs.ChainName,
		rs.Creator,
		rs.SessionKey,
		pgArray(rs.Signers),
		rs.MessageCounter,
	}
}

func (ras MsgAuthRvAllSessions) CopyRow() []any {
	return []any{
		ras.TxId,
		ras.Timestamp,
		ras.ChainName,
		ras.Creator,
		pgArray(ras.Signers),
		ras.MessageCounter,
	}
}
