package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var defaultLimit = uint64(10)

// GetAddressTxs returns the transactions involving a given address.
//
// There are two query modes:
//
//  1. Timestamp range: both fromTimestamp and toTimestamp are non-nil. The result
//     is ordered by (timestamp, tx_hash) according to sortOrder. hasMore is
//     always false in this mode; the caller is expected to narrow the window if
//     it needs fewer rows.
//  2. Cursor: fromTimestamp and toTimestamp are both nil. Pagination follows the
//     same keyset scheme as the transactions range API, using (block_height, tx_hash)
//     as the seek key. The response is always ordered newest-first and the caller
//     walks history via direction ("next" for older rows, "prev" for newer rows).
//     A cursor is required for direction=prev. sortOrder is ignored in this mode.
//
// Parameters:
//   - address: the bech32-style gno address to look up
//   - chainName: the chain to query
//   - fromTimestamp, toTimestamp: inclusive bounds for timestamp mode, both nil for cursor mode
//   - limit: max rows to return (defaults to 10 when nil)
//   - cursor: encoded "<block_height>|<tx_hash_base64url>"; nil in cursor mode means "start at the head"
//   - direction: Next (older) or Prev (newer); only used in cursor mode
//   - sortOrder: order for timestamp mode; ignored in cursor mode
//
// Returns:
//   - *[]AddressTx: the page of transactions
//   - bool: hasMore, whether another page exists in the walked direction (cursor mode only)
//   - error: any query error
func (t *TimescaleDb) GetAddressTxs(
	ctx context.Context,
	address string,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	limit *uint64,
	cursor *string,
	direction Direction,
	sortOrder SortOrder,
) (*[]AddressTx, bool, error) {
	hasTsRange := fromTimestamp != nil && toTimestamp != nil
	noTsRange := fromTimestamp == nil && toTimestamp == nil

	if !hasTsRange && !noTsRange {
		return nil, false, fmt.Errorf("invalid query parameters: from_timestamp and to_timestamp must both be set or both be unset")
	}

	accountId, err := t.getAccountId(ctx, address, chainName)
	if err != nil {
		return nil, false, fmt.Errorf("error getting account id: %w", err)
	}

	if hasTsRange {
		addressTxs, err := t.getAddressTxsTimestampQuery(
			ctx, accountId, chainName, *fromTimestamp, *toTimestamp, limit, sortOrder,
		)
		if err != nil {
			return nil, false, err
		}
		return addressTxs, false, nil
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
	sortOrder SortOrder,
) (*[]AddressTx, error) {
	if limit == nil {
		limit = &defaultLimit
	}

	order := sortOrder.SQL()
	query := fmt.Sprintf(`
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp AS timestamp,
		tx.msg_types AS msg_types,
		tg.block_height AS block_height
		FROM address_tx tx
		JOIN transaction_general tg ON tx.tx_hash = tg.tx_hash AND tx.chain_name = tg.chain_name
		WHERE tx.address = $1
		AND tx.chain_name = $2
		AND tx.timestamp >= $3
		AND tx.timestamp <= $4
		ORDER BY tx.timestamp %s, tx.tx_hash %s
		LIMIT $5
	`, order, order)

	args := []any{accountId, chainName, fromTimestamp, toTimestamp, *limit}
	return t.execAccQuery(ctx, query, args)
}

// getAddressTxsCursorQuery runs the cursor/direction mode. The output is always
// newest-first; direction=next walks toward older rows (DESC scan), direction=prev
// walks toward newer rows (ASC scan, then reverse). limit+1 rows are fetched so
// we can report hasMore without issuing a second COUNT query.
func (t *TimescaleDb) getAddressTxsCursorQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	cursor *string,
	limit *uint64,
	direction Direction,
) (*[]AddressTx, bool, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	fetchLimit := *limit + 1

	const selectCols = `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types,
		tg.block_height AS block_height
		FROM address_tx tx
		JOIN transaction_general tg ON tx.tx_hash = tg.tx_hash AND tx.chain_name = tg.chain_name
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
	case Next:
		if hasCursor {
			query = selectCols + `
			WHERE tx.address = $1
			AND tx.chain_name = $2
			AND (tg.block_height, tx.tx_hash) < ($3, $4)
			ORDER BY tg.block_height DESC, tx.tx_hash DESC
			LIMIT $5
			`
			args = []any{accountId, chainName, blockHeight, decodedTxHash, fetchLimit}
		} else {
			query = selectCols + `
			WHERE tx.address = $1
			AND tx.chain_name = $2
			ORDER BY tg.block_height DESC, tx.tx_hash DESC
			LIMIT $3
			`
			args = []any{accountId, chainName, fetchLimit}
		}
	case Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		query = selectCols + `
			WHERE tx.address = $1
			AND tx.chain_name = $2
			AND (tg.block_height, tx.tx_hash) > ($3, $4)
			ORDER BY tg.block_height ASC, tx.tx_hash ASC
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
) (*[]AddressTx, error) {
	addressTxs := make([]AddressTx, 0)
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var addressTx AddressTx
		err := rows.Scan(&addressTx.Hash, &addressTx.Timestamp, &addressTx.MsgTypes, &addressTx.BlockHeight)
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
		return 0, err
	}
	return id, nil
}
