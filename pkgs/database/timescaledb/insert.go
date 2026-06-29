package timescaledb

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// InsertAddresses inserts a slice of addresses into the database using COPY FROM.
func (t *TimescaleDb) InsertAddresses(
	ctx context.Context,
	addresses []string,
	chainName string,
	insertValidators bool,
) error {
	column_names := []string{"address", "chain_name"}
	var table_name string
	if insertValidators {
		table_name = "gno_validators"
	} else {
		table_name = "gno_addresses"
	}
	pgxSlice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return []any{addresses[i], chainName}, nil
	})
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{table_name}, column_names, pgxSlice)
	return err
}

// InsertBlocks inserts a slice of blocks into the database using COPY FROM.
func (t *TimescaleDb) InsertBlocks(ctx context.Context, blocks []s.Blocks) error {
	if len(blocks) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(blocks), func(i int) ([]any, error) {
		return []any{
			blocks[i].Hash,
			blocks[i].Height,
			blocks[i].Timestamp,
			blocks[i].ChainID,
			blocks[i].ChainName}, nil
	})

	columns := blocks[0].TableColumns()
	tableName := blocks[0].TableName()

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertValidatorBlockSignings inserts a slice of validator block signings using COPY FROM.
func (t *TimescaleDb) InsertValidatorBlockSignings(
	ctx context.Context,
	validatorBlockSigning []s.ValidatorBlockSigning,
) error {
	if len(validatorBlockSigning) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(validatorBlockSigning), func(i int) ([]any, error) {
		return []any{
			validatorBlockSigning[i].BlockHeight,
			validatorBlockSigning[i].Timestamp,
			validatorBlockSigning[i].Proposer,
			makePgxArray(validatorBlockSigning[i].SignedVals),
			validatorBlockSigning[i].ChainName}, nil
	})

	columns := validatorBlockSigning[0].TableColumns()
	tableName := validatorBlockSigning[0].TableName()

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertTransactionsGeneral inserts a slice of transaction general data using COPY FROM.
func (t *TimescaleDb) InsertTransactionsGeneral(
	ctx context.Context,
	transactionsGeneral []s.TransactionGeneral,
) error {
	if len(transactionsGeneral) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(transactionsGeneral), func(i int) ([]any, error) {
		return []any{
			transactionsGeneral[i].TxId,
			transactionsGeneral[i].ChainName,
			transactionsGeneral[i].Timestamp,
			transactionsGeneral[i].BlockHeight,
			makePgxArray(transactionsGeneral[i].MsgTypes),
			makePgxArray(transactionsGeneral[i].TxEvents),
			transactionsGeneral[i].TxEventsCompressed,
			transactionsGeneral[i].CompressionOn,
			transactionsGeneral[i].GasUsed,
			transactionsGeneral[i].GasWanted,
			transactionsGeneral[i].FeeAmount,
			transactionsGeneral[i].FeeDenom,
			transactionsGeneral[i].Success,
			transactionsGeneral[i].ErrorLog,
		}, nil
	})

	columns := transactionsGeneral[0].TableColumns()
	tableName := transactionsGeneral[0].TableName()

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertAddressTx inserts a slice of AddressTx into the database using COPY FROM.
func (t *TimescaleDb) InsertAddressTx(ctx context.Context, addresses []s.AddressTx) error {
	if len(addresses) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return []any{
			addresses[i].Address,
			addresses[i].TxId,
			addresses[i].ChainName,
			addresses[i].Timestamp,
		}, nil
	})

	columns := addresses[0].TableColumns()
	tableName := addresses[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgSend inserts a slice of MsgSend messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgSend(ctx context.Context, messages []s.MsgSend) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].FromAddress,
			messages[i].ToAddress,
			makePgxArray(messages[i].Amount),
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})

	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgCall inserts a slice of MsgCall messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgCall(ctx context.Context, messages []s.MsgCall) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Caller,
			messages[i].PkgPath,
			messages[i].FuncName,
			messages[i].Args,
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})

	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgAddPackage inserts a slice of MsgAddPackage messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAddPackage(
	ctx context.Context,
	messages []s.MsgAddPackage,
) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Creator,
			messages[i].PkgPath,
			messages[i].PkgName,
			makePgxArray(messages[i].PkgFileNames),
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})

	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgRun inserts a slice of MsgRun messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgRun(
	ctx context.Context,
	messages []s.MsgRun,
) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Caller,
			messages[i].PkgPath,
			messages[i].PkgName,
			makePgxArray(messages[i].PkgFileNames),
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})

	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgMultiSend inserts a batch of MsgMultiSend messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgMultiSend(
	ctx context.Context,
	messages []s.MsgMultiSend,
) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Direction,
			messages[i].AddressId,
			makePgxArray(messages[i].Coins),
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})

	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgAuthCrSession inserts a slice of MsgAuthCrSession messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthCrSession(ctx context.Context, messages []s.MsgAuthCrSession) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Creator,
			messages[i].SessionKey,
			messages[i].ExpiresAt,
			makePgxArray(messages[i].SpendLimit),
			makePgxArray(messages[i].AllowPaths),
			messages[i].SpendPeriod,
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})
	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgAuthRvSession inserts a slice of MsgAuthRvSession messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthRvSession(ctx context.Context, messages []s.MsgAuthRvSession) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Creator,
			messages[i].SessionKey,
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})
	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// InsertMsgAuthRvAllSessions inserts a slice of MsgAuthRvAllSessions messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthRvAllSessions(ctx context.Context, messages []s.MsgAuthRvAllSessions) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxId,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Creator,
			makePgxArray(messages[i].Signers),
			messages[i].MessageCounter,
		}, nil
	})
	columns := messages[0].TableColumns()
	tableName := messages[0].TableName()
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgxSlice)
	return err
}

// makePgxArray converts a Go slice into a pgx typed array for COPY FROM inserts.
// Handles composite types and byte arrays; works on any insertable slice type.
func makePgxArray[T any](v []T) pgtype.Array[T] {
	if v == nil {
		return pgtype.Array[T]{Valid: false}
	}

	return pgtype.Array[T]{
		Elements: v,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}},
		Valid:    true,
	}
}

func (t *TimescaleDb) InsertTxHashIds(
	ctx context.Context,
	txHashes []string,
	timestamps []time.Time,
	chainName string,
) (map[string]int64, error) {
	txLength := len(txHashes)
	timestampLength := len(timestamps)

	if txLength <= 0 || timestampLength <= 0 {
		return nil, fmt.Errorf("no tx hashes to insert")
	}
	if txLength != timestampLength {
		return nil, fmt.Errorf("tx hashes and timestamps must have the same length")
	}

	txHashBytes := make([][]byte, txLength)
	for i, hash := range txHashes {
		decoded, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			return nil, err
		}
		txHashBytes[i] = decoded
	}

	rows, err := t.pool.Query(
		ctx,
		`INSERT INTO tx_hash_id (tx_hash, timestamp, chain_name)
		 SELECT unnest($1::bytea[]), unnest($2::timestamptz[]), $3
		 RETURNING tx_hash, tx_id`,
		txHashBytes,
		timestamps,
		chainName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txHashIdMap := make(map[string]int64, txLength)
	for rows.Next() {
		var txHash []byte
		var txId int64
		if err := rows.Scan(&txHash, &txId); err != nil {
			return nil, fmt.Errorf("failed to scan tx hash id: %w", err)
		}
		txHashIdMap[base64.StdEncoding.EncodeToString(txHash)] = txId
	}
	return txHashIdMap, rows.Err()
}
