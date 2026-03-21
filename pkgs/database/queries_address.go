package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

var defaultLimit = uint64(10)

// GetAddressTxs gets the transactions for a given address for a certain time period
//
// Usage:
//
// # Used to get the transactions for a given address for a certain time period
//
// Parameters:
//   - address: the address
//   - chainName: the name of the chain
//   - fromTimestamp: the starting timestamp
//   - toTimestamp: the ending timestamp
//
// Returns:
//   - []*AddressTx: the transactions
//   - error: if the query fails
func (t *TimescaleDb) GetAddressTxs(
	ctx context.Context,
	address string,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	limit *uint64,
	page *uint64,
	cursor *string,
) (*[]AddressTx, string, uint64, error) {
	hasTsRange := fromTimestamp != nil && toTimestamp != nil
	noTsRange := fromTimestamp == nil && toTimestamp == nil

	var mode string
	switch {
	case hasTsRange:
		mode = "timestamp"
	case noTsRange && page == nil:
		mode = "cursor"
	case noTsRange && cursor == nil:
		mode = "limit_page"
	default:
		return nil, "", 0, fmt.Errorf("invalid query parameters")
	}

	accountId, err := t.getAccountId(ctx, address, chainName)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting account id: %w", err)
	}

	txCount, err := t.getTxsCount(ctx, accountId, chainName)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting tx count: %w", err)
	}

	var addressTxs *[]AddressTx
	var nextCursor string

	switch mode {
	case "timestamp":
		addressTxs, err = t.getAddressTxsTimestampQuery(
			ctx, accountId, chainName, *fromTimestamp, *toTimestamp, limit,
		)
		if err != nil {
			return nil, "", 0, err
		}
	case "cursor":
		addressTxs, nextCursor, err = t.getAddressTxsCursorQuery(
			ctx, accountId, chainName, cursor, limit, txCount,
		)
		if err != nil {
			return nil, "", 0, err
		}
	case "limit_page":
		addressTxs, err = t.getAddressTxsLimitPageQuery(
			ctx, accountId, chainName, limit, *page,
		)
		if err != nil {
			return nil, "", 0, err
		}
	}

	return addressTxs, nextCursor, txCount, nil
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
) (*[]AddressTx, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	var args []any

	query := `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		AND tx.timestamp >= $3
		AND tx.timestamp <= $4
		ORDER BY tx.timestamp DESC
		LIMIT $5
		`
	args = append(args, accountId, chainName, fromTimestamp, toTimestamp, *limit)

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return addressTxs, nil
}

func (t *TimescaleDb) getAddressTxsCursorQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	cursor *string,
	limit *uint64,
	txCount uint64,
) (*[]AddressTx, string, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	// Fetch limit+1 to detect if there are more rows; only return limit to the caller.
	fetchLimit := *limit + 1
	var query string
	var args []any

	if cursor == nil {
		query = `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		ORDER BY tx.timestamp DESC, tx.tx_hash DESC
		LIMIT $3
		`
		args = append(args, accountId, chainName, fetchLimit)
	} else {
		timestamp, txHash, err := unmarshalCursorParam(*cursor)
		if err != nil {
			return nil, "", err
		}
		decodedTxHash, err := base64.URLEncoding.Strict().DecodeString(txHash)
		if err != nil {
			return nil, "", fmt.Errorf("error decoding tx hash: %w", err)
		}
		query = `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		AND (tx.timestamp, tx.tx_hash) < ($3::timestamptz, $4)
		ORDER BY tx.timestamp DESC, tx.tx_hash DESC
		LIMIT $5
		`
		args = append(args, accountId, chainName, timestamp, decodedTxHash, fetchLimit)
	}

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, "", err
	}
	// If we got more than limit, there is a next page: return only the first limit and set nextCursor.
	if len(*addressTxs) > int(*limit) {
		page := (*addressTxs)[:int(*limit)]
		lastAddressTx := page[len(page)-1]
		nextCursor := makeCursorParam(lastAddressTx.Timestamp, lastAddressTx.Hash)
		return &page, nextCursor, nil
	}
	return addressTxs, "", nil
}

func (t *TimescaleDb) getAddressTxsLimitPageQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	limit *uint64,
	page uint64,
) (*[]AddressTx, error) {
	if limit == nil {
		limit = &defaultLimit
	}

	var query string
	var args []any

	offset := page * *limit

	query = `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.msg_types
	FROM address_tx tx
	WHERE tx.address = $1
	AND tx.chain_name = $2
	ORDER BY tx.timestamp DESC
	LIMIT $3 OFFSET $4
	`
	args = append(args, accountId, chainName, *limit, offset)

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return addressTxs, nil
}

func (t *TimescaleDb) getTxsCount(
	ctx context.Context,
	accountId int32,
	chainName string,
) (uint64, error) {
	query := `
	SELECT COUNT(*) FROM address_tx WHERE address = $1 AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, accountId, chainName)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func makeCursorParam(
	timestamp time.Time,
	txHash string,
) string {
	txHashBytes, err := base64.StdEncoding.DecodeString(txHash)
	if err != nil {
		// TODO: log error
		return "error decoding tx hash"
	}
	timestamp = timestamp.Round(time.Second)
	base64Url := base64.URLEncoding.Strict().EncodeToString(txHashBytes)
	return timestamp.Format(time.RFC3339) + "|" + base64Url
}

func unmarshalCursorParam(
	cursor string,
) (time.Time, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	timestamp, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return timestamp, parts[1], nil
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
		err := rows.Scan(&addressTx.Hash, &addressTx.Timestamp, &addressTx.MsgTypes)
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
