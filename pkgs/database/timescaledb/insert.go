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

// InsertBlocks inserts a slice of blocks into the database using COPY FROM.
func (t *TimescaleDb) InsertBlocks(ctx context.Context, blocks []s.Blocks) error {
	if len(blocks) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(blocks), func(i int) ([]any, error) {
		return blocks[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{blocks[0].TableName()}, blocks[0].TableColumns(), pgxSlice)
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
		return validatorBlockSigning[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{validatorBlockSigning[0].TableName()}, validatorBlockSigning[0].TableColumns(), pgxSlice)
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
		return transactionsGeneral[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{transactionsGeneral[0].TableName()}, transactionsGeneral[0].TableColumns(), pgxSlice)
	return err
}

// InsertAddressTx inserts a slice of AddressTx into the database using COPY FROM.
func (t *TimescaleDb) InsertAddressTx(ctx context.Context, addresses []s.AddressTx) error {
	if len(addresses) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return addresses[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{addresses[0].TableName()}, addresses[0].TableColumns(), pgxSlice)
	return err
}

// InsertMsgSend inserts a slice of MsgSend messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgSend(ctx context.Context, messages []s.MsgSend) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return messages[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
	return err
}

// InsertMsgCall inserts a slice of MsgCall messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgCall(ctx context.Context, messages []s.MsgCall) error {
	if len(messages) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return messages[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
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
		return messages[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
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
		return messages[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
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
		return messages[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
	return err
}

// InsertMsgAuthCrSession inserts a slice of MsgAuthCrSession messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthCrSession(ctx context.Context, messages []s.MsgAuthCrSession) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return messages[i].CopyRow(), nil
	})
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
	return err
}

// InsertMsgAuthRvSession inserts a slice of MsgAuthRvSession messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthRvSession(ctx context.Context, messages []s.MsgAuthRvSession) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return messages[i].CopyRow(), nil
	})
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
	return err
}

// InsertMsgAuthRvAllSessions inserts a slice of MsgAuthRvAllSessions messages into the database using COPY FROM.
func (t *TimescaleDb) InsertMsgAuthRvAllSessions(ctx context.Context, messages []s.MsgAuthRvAllSessions) error {
	if len(messages) == 0 {
		return nil
	}
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return messages[i].CopyRow(), nil
	})
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{messages[0].TableName()}, messages[0].TableColumns(), pgxSlice)
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
