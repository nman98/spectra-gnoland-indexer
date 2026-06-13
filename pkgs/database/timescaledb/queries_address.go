package timescaledb

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/jackc/pgx/v5"
)

var defaultLimit = uint64(10)

// GetAddressTxs returns the transactions involving a given address.
//
// There are two query modes:
//
//  1. Timestamp range: both fromTimestamp and toTimestamp are non-nil.
//  2. Cursor: fromTimestamp and toTimestamp are both nil. Pagination follows
//     the keyset scheme using (block_height, tx_hash) as the seek key.
func (t *TimescaleDb) GetAddressTxs(
	ctx context.Context,
	address string,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	limit *uint64,
	cursor *string,
	direction database.Direction,
) (*[]database.AddressTx, bool, error) {
	hasTsRange := fromTimestamp != nil && toTimestamp != nil
	noTsRange := fromTimestamp == nil && toTimestamp == nil

	if !hasTsRange && !noTsRange {
		return nil, false, fmt.Errorf("invalid query parameters: from_timestamp and to_timestamp must both be set or both be unset")
	}

	accountId, err := t.getAccountId(ctx, address, chainName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, false, err
		}
		return nil, false, fmt.Errorf("error getting account id: %w", err)
	}

	if hasTsRange {
		addressTxs, hasMore, err := t.getAddressTxsTimestampQuery(
			ctx, accountId, chainName, *fromTimestamp, *toTimestamp, limit, cursor, direction,
		)
		if err != nil {
			return nil, false, err
		}
		return addressTxs, hasMore, nil
	}

	return t.getAddressTxsCursorQuery(ctx, accountId, chainName, cursor, limit, direction)
}

func (t *TimescaleDb) GetTotalAddressesCount(ctx context.Context, chainName string) (int32, error) {
	query := `
	SELECT
	MAX(id)
	FROM gno_addresses
	WHERE chain_name = $1
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var maxId int32
	err := row.Scan(&maxId)
	if err != nil {
		return 0, err
	}
	return maxId, nil
}

func (t *TimescaleDb) getAddressTxsTimestampQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	limit *uint64,
	cursor *string,
	direction database.Direction,
) (*[]database.AddressTx, bool, error) {
	if limit == nil {
		limit = &defaultLimit
	}

	fetchLimit := *limit + 1

	const selectCols = `
		SELECT
		encode(id.tx_hash, 'base64') AS tx_hash,
		at.timestamp,
		tg.msg_types,
		tg.block_height AS block_height,
		tg.success AS success,
		tg.error_log AS error_log
		FROM address_tx at
		JOIN tx_hash_id id ON at.tx_id = id.tx_id AND at.chain_name = id.chain_name
		JOIN transaction_general tg ON at.tx_id = tg.tx_id AND at.chain_name = tg.chain_name
	`

	hasCursor := cursor != nil && *cursor != ""
	var (
		query         string
		args          []any
		reverse       bool
		blockHeight   uint64
		decodedTxHash []byte
	)
	if hasCursor {
		bh, txHash, err := unmarshalCursorParam(*cursor)
		if err != nil {
			return nil, false, err
		}
		decoded, err := base64.URLEncoding.Strict().DecodeString(txHash)
		if err != nil {
			return nil, false, fmt.Errorf("error decoding cursor tx hash: %w", err)
		}
		blockHeight = bh
		decodedTxHash = decoded
	}

	switch direction {
	case database.Next:
		if hasCursor {
			query = selectCols + `
			WHERE at.address = $1
			AND at.chain_name = $2
			AND (tg.block_height, id.tx_hash) < ($3, $4)
			AND at.timestamp BETWEEN $5 AND $6
			ORDER BY tg.block_height DESC, id.tx_hash DESC
			LIMIT $7
			`
			args = []any{accountId, chainName, blockHeight, decodedTxHash, fromTimestamp, toTimestamp, fetchLimit}
		} else {
			query = selectCols + `
			WHERE at.address = $1
			AND at.chain_name = $2
			AND at.timestamp BETWEEN $3 AND $4
			ORDER BY tg.block_height DESC, id.tx_hash DESC
			LIMIT $5
			`
			args = []any{accountId, chainName, fromTimestamp, toTimestamp, fetchLimit}
		}
	case database.Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		query = selectCols + `
		WHERE at.address = $1
		AND at.chain_name = $2
		AND (tg.block_height, id.tx_hash) > ($3, $4)
		AND at.timestamp BETWEEN $5 AND $6
		ORDER BY tg.block_height ASC, id.tx_hash ASC
		LIMIT $7
		`
		args = []any{accountId, chainName, blockHeight, decodedTxHash, fromTimestamp, toTimestamp, fetchLimit}
		reverse = true
	default:
		return nil, false, fmt.Errorf("invalid direction: %q", direction)
	}

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, false, err
	}

	hasMore := uint64(len(*addressTxs)) > *limit
	if hasMore {
		trimmed := (*addressTxs)[:*limit]
		addressTxs = &trimmed
	}
	if reverse {
		rows := *addressTxs
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	return addressTxs, hasMore, nil
}

// getAddressTxsCursorQuery runs cursor/direction mode. Output is always newest-first;
// direction=next walks toward older rows (DESC), direction=prev walks toward newer rows
// (ASC scan, then reversed). Fetches limit+1 to determine hasMore without a COUNT query.
func (t *TimescaleDb) getAddressTxsCursorQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	cursor *string,
	limit *uint64,
	direction database.Direction,
) (*[]database.AddressTx, bool, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	fetchLimit := *limit + 1

	const selectCols = `
		SELECT
		encode(id.tx_hash, 'base64') AS tx_hash,
		at.timestamp,
		tg.msg_types,
		tg.block_height AS block_height,
		tg.success AS success,
		tg.error_log AS error_log
		FROM address_tx at
		JOIN tx_hash_id id ON at.tx_id = id.tx_id AND at.chain_name = id.chain_name
		JOIN transaction_general tg ON at.tx_id = tg.tx_id AND at.chain_name = tg.chain_name
	`

	hasCursor := cursor != nil && *cursor != ""
	var (
		query         string
		args          []any
		reverse       bool
		blockHeight   uint64
		decodedTxHash []byte
	)
	if hasCursor {
		bh, txHash, err := unmarshalCursorParam(*cursor)
		if err != nil {
			return nil, false, err
		}
		decoded, err := base64.URLEncoding.Strict().DecodeString(txHash)
		if err != nil {
			return nil, false, fmt.Errorf("error decoding cursor tx hash: %w", err)
		}
		blockHeight = bh
		decodedTxHash = decoded
	}

	switch direction {
	case database.Next:
		if hasCursor {
			query = selectCols + `
			WHERE at.address = $1
			AND at.chain_name = $2
			AND (tg.block_height, id.tx_hash) < ($3, $4)
			ORDER BY tg.block_height DESC, id.tx_hash DESC
			LIMIT $5
			`
			args = []any{accountId, chainName, blockHeight, decodedTxHash, fetchLimit}
		} else {
			query = selectCols + `
			WHERE at.address = $1
			AND at.chain_name = $2
			ORDER BY tg.block_height DESC, id.tx_hash DESC
			LIMIT $3
			`
			args = []any{accountId, chainName, fetchLimit}
		}
	case database.Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		query = selectCols + `
			WHERE at.address = $1
			AND at.chain_name = $2
			AND (tg.block_height, id.tx_hash) > ($3, $4)
			ORDER BY tg.block_height ASC, id.tx_hash ASC
			LIMIT $5
			`
		args = []any{accountId, chainName, blockHeight, decodedTxHash, fetchLimit}
		reverse = true
	default:
		return nil, false, fmt.Errorf("invalid direction: %q", direction)
	}

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, false, err
	}

	hasMore := uint64(len(*addressTxs)) > *limit
	if hasMore {
		trimmed := (*addressTxs)[:*limit]
		addressTxs = &trimmed
	}
	if reverse {
		rows := *addressTxs
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	return addressTxs, hasMore, nil
}

func unmarshalCursorParam(
	cursor string,
) (uint64, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid cursor: %s: expected 2 parts", cursor)
	}
	blockHeight, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("error parsing block height: %w", err)
	}
	return blockHeight, parts[1], nil
}

func (t *TimescaleDb) execAccQuery(
	ctx context.Context,
	query string,
	args []any,
) (*[]database.AddressTx, error) {
	addressTxs := make([]database.AddressTx, 0)
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var addressTx database.AddressTx
		err := rows.Scan(&addressTx.Hash, &addressTx.Timestamp, &addressTx.MsgTypes, &addressTx.BlockHeight, &addressTx.Success, &addressTx.ErrorLog)
		if err != nil {
			return nil, err
		}
		addressTxs = append(addressTxs, addressTx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &addressTxs, nil
}

func (t *TimescaleDb) getAccountId(
	ctx context.Context,
	address string,
	chainName string,
) (int32, error) {
	query := `
	SELECT id FROM gno_addresses WHERE address = $1 AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, address, chainName)
	var id int32
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("address %q: %w", address, database.ErrNotFound)
		}
		return 0, err
	}
	return id, nil
}
