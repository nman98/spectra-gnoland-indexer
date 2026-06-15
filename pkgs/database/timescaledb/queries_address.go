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

// selectCols is the shared projection/join for address transaction queries.
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

	return t.queryAddressTxs(
		ctx, accountId, chainName, fromTimestamp, toTimestamp, cursor, limit, direction,
	)
}

func (t *TimescaleDb) GetTotalAddressesCount(ctx context.Context, chainName string) (int32, error) {
	query := `
	SELECT
	id::int4
	FROM gno_addresses
	WHERE chain_name = $1
	ORDER BY id DESC
	LIMIT 1
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var maxId int32
	err := row.Scan(&maxId)
	if err != nil {
		return 0, err
	}
	return maxId, nil
}

// queryAddressTxs runs both query modes for GetAddressTxs. When fromTimestamp and
// toTimestamp are non-nil a timestamp BETWEEN predicate is added; otherwise it is
// pure cursor/keyset mode. Output is always newest-first: direction=next walks toward
// older rows (DESC), direction=prev walks toward newer rows (ASC scan, then reversed).
// Fetches limit+1 to determine hasMore without a COUNT query.
func (t *TimescaleDb) queryAddressTxs(
	ctx context.Context,
	accountId int32,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	cursor *string,
	limit *uint64,
	direction database.Direction,
) (*[]database.AddressTx, bool, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	fetchLimit := *limit + 1

	hasCursor, blockHeight, decodedTxHash, err := decodeCursor(cursor)
	if err != nil {
		return nil, false, err
	}

	// arg appends a query argument and returns its positional placeholder ($1, $2, ...),
	// so the SQL stays correct regardless of which optional predicates are present.
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	conds := []string{
		"at.address = " + arg(accountId),
		"at.chain_name = " + arg(chainName),
	}

	var (
		order   string
		reverse bool
	)
	switch direction {
	case database.Next:
		if hasCursor {
			conds = append(conds, fmt.Sprintf("(tg.block_height, id.tx_hash) < (%s, %s)", arg(blockHeight), arg(decodedTxHash)))
		}
		order = "ORDER BY tg.block_height DESC, id.tx_hash DESC"
	case database.Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		conds = append(conds, fmt.Sprintf("(tg.block_height, id.tx_hash) > (%s, %s)", arg(blockHeight), arg(decodedTxHash)))
		order = "ORDER BY tg.block_height ASC, id.tx_hash ASC"
		reverse = true
	default:
		return nil, false, fmt.Errorf("invalid direction: %q", direction)
	}

	if fromTimestamp != nil && toTimestamp != nil {
		conds = append(conds, fmt.Sprintf("at.timestamp BETWEEN %s AND %s", arg(*fromTimestamp), arg(*toTimestamp)))
	}

	query := selectCols +
		" WHERE " + strings.Join(conds, " AND ") +
		" " + order +
		" LIMIT " + arg(fetchLimit)

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, false, err
	}
	result, hasMore := finalizeAddressTxs(addressTxs, *limit, reverse)
	return result, hasMore, nil
}

// decodeCursor reports whether a usable cursor is present and, if so, returns its
// decoded (block_height, tx_hash) seek key.
func decodeCursor(cursor *string) (bool, uint64, []byte, error) {
	if cursor == nil || *cursor == "" {
		return false, 0, nil, nil
	}
	blockHeight, txHash, err := unmarshalCursorParam(*cursor)
	if err != nil {
		return false, 0, nil, err
	}
	decodedTxHash, err := base64.URLEncoding.Strict().DecodeString(txHash)
	if err != nil {
		return false, 0, nil, fmt.Errorf("error decoding cursor tx hash: %w", err)
	}
	return true, blockHeight, decodedTxHash, nil
}

// finalizeAddressTxs trims the limit+1 probe row, reports hasMore, and restores
// newest-first ordering for reverse (prev) scans.
func finalizeAddressTxs(addressTxs *[]database.AddressTx, limit uint64, reverse bool) (*[]database.AddressTx, bool) {
	hasMore := uint64(len(*addressTxs)) > limit
	if hasMore {
		trimmed := (*addressTxs)[:limit]
		addressTxs = &trimmed
	}
	if reverse {
		reverseSlice(*addressTxs)
	}
	return addressTxs, hasMore
}

// reverseSlice reverses s in place.
func reverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
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
