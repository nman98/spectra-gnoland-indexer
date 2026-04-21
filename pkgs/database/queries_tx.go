package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	dictloader "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/dict_loader"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/cosmos/gogoproto/proto"
	"github.com/klauspost/compress/zstd"
)

var dictBytes = dictloader.LoadDict()
var zstdDict = zstd.WithDecoderDicts(dictBytes)
var zstdReader *zstd.Decoder

var l = logger.Get()

func init() {
	var err error
	zstdReader, err = zstd.NewReader(nil, zstdDict)
	if err != nil {
		l.Fatal().Caller().Stack().Err(err).Msg("failed to initialize zstd reader")
	}
}

// GetTransaction gets the transaction for a given transaction hash
//
// Usage:
//
// # Used to get the transaction for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - *Transaction: the transaction
//   - error: if the query fails
func (t *TimescaleDb) GetTransaction(ctx context.Context, txHash string, chainName string) (*Transaction, error) {
	query := `
	SELECT 
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.tx_hash = decode($1, 'base64')
	AND tx.chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var transaction FullTxData
	err := row.Scan(
		&transaction.TxHash,
		&transaction.Timestamp,
		&transaction.BlockHeight,
		&transaction.TxEvents,
		&transaction.TxEventsCompressed,
		&transaction.CompressionOn,
		&transaction.GasUsed,
		&transaction.GasWanted,
		&transaction.Fee,
		&transaction.MsgTypes,
	)
	if err != nil {
		log.Println("error getting transaction", err)
		return nil, err
	}
	tx, err := transaction.ToTransaction(decodeEvents)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GetLastXTransactions gets the last x transactions from the database for a given chain name
//
// Usage:
//
// # Used to get the last x transactions from the database for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - x: the number of transactions to get
func (t *TimescaleDb) GetLastXTransactions(
	ctx context.Context,
	chainName string,
	x uint64,
	sortOrder *SortOrder,
) ([]*Transaction, error) {
	// The only usage for this is for the transaction queried by cursor, it shouldn't be allowed to be
	// queried on the endpoint for the last x transactions for now.
	if sortOrder == nil {
		sortOrder = new(SortOrder)
		*sortOrder = SortOrderDesc
	}
	order := sortOrder.SQL()
	query := fmt.Sprintf(`
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.chain_name = $1
	ORDER BY tx.timestamp %s
	LIMIT $2
	`, order)
	rows, err := t.pool.Query(ctx, query, chainName, x)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &FullTxData{}
		err := rows.Scan(
			&transaction.TxHash,
			&transaction.Timestamp,
			&transaction.BlockHeight,
			&transaction.TxEvents,
			&transaction.TxEventsCompressed,
			&transaction.CompressionOn,
			&transaction.GasUsed,
			&transaction.GasWanted,
			&transaction.Fee,
			&transaction.MsgTypes)
		if err != nil {
			return nil, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByOffset gets the transactions by offset for a given chain name
//
// Usage:
//
// # Used to get the transactions by offset for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - limit: the limit of the transactions to get
//   - offset: the offset of the transactions to get
//
// Additional info:
//
// This part of the logic won't be officially present in the API. It's current only usage
// is with training the zstd dictionary. You are welcome to modify this function and add it to the API.
func (t *TimescaleDb) GetTransactionsByOffset(
	ctx context.Context,
	chainName string,
	limit uint64,
	offset uint64,
) ([]*Transaction, error) {
	query := `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.chain_name = $1
	ORDER BY tx.timestamp DESC
	LIMIT $2 OFFSET $3
	`
	rows, err := t.pool.Query(ctx, query, chainName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &FullTxData{}
		err := rows.Scan(&transaction.TxHash, &transaction.Timestamp, &transaction.BlockHeight, &transaction.TxEvents, &transaction.GasUsed, &transaction.GasWanted, &transaction.Fee, &transaction.MsgTypes)
		if err != nil {
			return nil, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByRange returns a page of transactions using keyset (cursor) pagination
// over the composite key (block_height, tx_hash).
//
// The response is always ordered newest-first (DESC by block_height, tx_hash). The caller
// distinguishes whether it wants older rows ("next") or newer rows ("prev"):
//
//   - direction == Next: walk toward older rows. Without a cursor it returns the latest
//     page. With a cursor it returns rows strictly older than the cursor.
//   - direction == Prev: walk toward newer rows. A cursor is required; the query is run
//     ASC and the result is reversed before returning so the caller still sees newest-first.
//
// The function fetches limit+1 rows internally and truncates. The returned hasMore flag
// indicates that another page exists in the direction that was walked.
//
// Parameters:
//   - chainName: the name of the chain
//   - cursor: encoded cursor "<block_height>|<tx_hash_base64url>"; empty means "start"
//   - limit: number of transactions to return (1..100)
//   - direction: Next or Prev
//
// Returns:
//   - []*Transaction: the page, newest-first
//   - bool: hasMore, true if another page exists in the walked direction
//   - error: if the query fails
func (t *TimescaleDb) GetTransactionsByRange(
	ctx context.Context,
	chainName string,
	cursor string,
	limit uint64,
	direction Direction,
) ([]*Transaction, bool, error) {
	if limit == 0 || limit > 100 {
		return nil, false, fmt.Errorf("limit must be between 1 and 100")
	}

	const selectCols = `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	`

	hasCursor := cursor != ""
	var (
		query         string
		args          []any
		reverse       bool
		blockHeight   uint64
		decodedTxHash []byte
	)

	if hasCursor {
		bh, txHash, err := unmarshalCursorParam(cursor)
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

	fetchLimit := limit + 1

	switch direction {
	case Next:
		if hasCursor {
			query = selectCols + `
			WHERE tx.chain_name = $1
			AND (tx.block_height, tx.tx_hash) < ($2, $3)
			ORDER BY tx.block_height DESC, tx.tx_hash DESC
			LIMIT $4
			`
			args = []any{chainName, blockHeight, decodedTxHash, fetchLimit}
		} else {
			query = selectCols + `
			WHERE tx.chain_name = $1
			ORDER BY tx.block_height DESC, tx.tx_hash DESC
			LIMIT $2
			`
			args = []any{chainName, fetchLimit}
		}
	case Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		query = selectCols + `
		WHERE tx.chain_name = $1
		AND (tx.block_height, tx.tx_hash) > ($2, $3)
		ORDER BY tx.block_height ASC, tx.tx_hash ASC
		LIMIT $4
		`
		args = []any{chainName, blockHeight, decodedTxHash, fetchLimit}
		reverse = true
	default:
		return nil, false, fmt.Errorf("invalid direction: %q", direction)
	}

	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	transactions := make([]*Transaction, 0, fetchLimit)
	for rows.Next() {
		transaction := &FullTxData{}
		err := rows.Scan(
			&transaction.TxHash,
			&transaction.Timestamp,
			&transaction.BlockHeight,
			&transaction.TxEvents,
			&transaction.TxEventsCompressed,
			&transaction.CompressionOn,
			&transaction.GasUsed,
			&transaction.GasWanted,
			&transaction.Fee,
			&transaction.MsgTypes,
		)
		if err != nil {
			return nil, false, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, false, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := uint64(len(transactions)) > limit
	if hasMore {
		transactions = transactions[:limit]
	}
	if reverse {
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
		}
	}
	return transactions, hasMore, nil
}

func decompressEvents(txEvents []byte) ([]byte, error) {
	decompressed, err := zstdReader.DecodeAll(txEvents, nil)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func protoUnmarshal(rawData []byte) (*events_proto.TxEvents, error) {
	txEvents := &events_proto.TxEvents{}
	err := proto.Unmarshal(rawData, txEvents)
	if err != nil {
		return nil, err
	}
	return txEvents, nil
}

func decodeEvents(txEvents []byte) (*[]Event, error) {
	if len(txEvents) == 0 {
		return &[]Event{}, nil
	}
	decompressed, err := decompressEvents(txEvents)
	if err != nil {
		return nil, err
	}
	txEventsProto, err := protoUnmarshal(decompressed)
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(txEventsProto.Events))
	for _, event := range txEventsProto.Events {
		attributes := make([]Attribute, 0, len(event.Attributes))
		for _, attribute := range event.Attributes {
			var value string
			switch v := attribute.Value.(type) {
			case *events_proto.Attribute_StringValue:
				value = v.StringValue
			case *events_proto.Attribute_Int64Value:
				value = strconv.FormatInt(v.Int64Value, 10)
			case *events_proto.Attribute_BoolValue:
				value = strconv.FormatBool(v.BoolValue)
			case *events_proto.Attribute_DoubleValue:
				value = strconv.FormatFloat(v.DoubleValue, 'g', -1, 64)
			default:
				value = ""
			}
			attributes = append(attributes, Attribute{
				Key:   attribute.Key,
				Value: value,
			})
		}
		pkgPath := ""
		if event.PkgPath != nil {
			pkgPath = *event.PkgPath
		}
		events = append(events, Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: attributes,
			PkgPath:    pkgPath,
		})
	}
	return &events, nil
}
