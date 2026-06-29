package timescaledb

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/jackc/pgx/v5"
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

// InsertRows inserts a homogeneous batch of rows into the database using COPY FROM.
// Every element must belong to the same table; the table name and column list are
// read from the first element. CopyRow supplies each row's values in column order
// (kept aligned with TableColumns by TestCopyRowMatchesColumns in pkgs/schema).
func (t *TimescaleDb) InsertRows(ctx context.Context, rows []s.Insertable) error {
	if len(rows) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
		return rows[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{rows[0].TableName()}, rows[0].TableColumns(), pgxSlice)
	return err
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
