package timescaledb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	dictloader "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/dict_loader"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/cosmos/gogoproto/proto"
	"github.com/jackc/pgx/v5"
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

// txSelectCols is the shared full-transaction projection. Its column order must stay
// in lockstep with txScanDest.
const txSelectCols = `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee_amount,
	tx.fee_denom,
	tx.msg_types,
	tx.success,
	tx.error_log
	FROM transaction_general tx
	JOIN tx_hash_id id ON tx.tx_id = id.tx_id AND tx.chain_name = id.chain_name
`

// txScanDest returns the scan destinations for txSelectCols, in matching column order.
func txScanDest(t *database.FullTxData) []any {
	return []any{
		&t.TxHash,
		&t.Timestamp,
		&t.BlockHeight,
		&t.TxEvents,
		&t.TxEventsCompressed,
		&t.CompressionOn,
		&t.GasUsed,
		&t.GasWanted,
		&t.Fee.Amount,
		&t.Fee.Denom,
		&t.MsgTypes,
		&t.Success,
		&t.ErrorLog,
	}
}

// scanTransactionRows drains a result set produced by txSelectCols into decoded
// transactions.
func scanTransactionRows(rows pgx.Rows) ([]*database.Transaction, error) {
	defer rows.Close()
	transactions := make([]*database.Transaction, 0)
	for rows.Next() {
		transaction := &database.FullTxData{}
		if err := rows.Scan(txScanDest(transaction)...); err != nil {
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

// GetTransaction gets the transaction for a given transaction hash.
func (t *TimescaleDb) GetTransaction(ctx context.Context, txHash string, chainName string) (*database.Transaction, error) {
	query := txSelectCols + `
	WHERE id.tx_hash = decode($1, 'base64')
	AND tx.chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var transaction database.FullTxData
	err := row.Scan(txScanDest(&transaction)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("transaction %s: %w", txHash, database.ErrNotFound)
		}
		log.Println("error getting transaction", err)
		return nil, err
	}
	tx, err := transaction.ToTransaction(decodeEvents)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GetLastXTransactions gets the last x transactions for a given chain name.
func (t *TimescaleDb) GetLastXTransactions(
	ctx context.Context,
	chainName string,
	x uint64,
	sortOrder *database.SortOrder,
) ([]*database.Transaction, error) {
	if sortOrder == nil {
		sortOrder = new(database.SortOrder)
		*sortOrder = database.SortOrderDesc
	}
	order := sortOrder.SQL()
	query := fmt.Sprintf(txSelectCols+`
	WHERE tx.chain_name = $1
	ORDER BY tx.timestamp %s
	LIMIT $2
	`, order)
	rows, err := t.pool.Query(ctx, query, chainName, x)
	if err != nil {
		return nil, err
	}
	transactions, err := scanTransactionRows(rows)
	if err != nil {
		return nil, err
	}
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions for chain %q: %w", chainName, database.ErrNotFound)
	}
	return transactions, nil
}

// GetTransactionsByOffset returns transactions by offset. Used only for zstd dictionary training.
func (t *TimescaleDb) GetTransactionsByOffset(
	ctx context.Context,
	chainName string,
	limit uint64,
	offset uint64,
) ([]*database.Transaction, error) {
	query := `
	SELECT
	encode(id.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee_amount,
	tx.fee_denom,
	tx.msg_types,
	tx.success,
	tx.error_log
	FROM transaction_general tx
	JOIN tx_hash_id id ON tx.tx_id = id.tx_id AND tx.chain_name = id.chain_name
	WHERE tx.chain_name = $1
	ORDER BY tx.timestamp DESC
	LIMIT $2 OFFSET $3
	`
	rows, err := t.pool.Query(ctx, query, chainName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*database.Transaction, 0)
	for rows.Next() {
		transaction := &database.FullTxData{}
		err := rows.Scan(
			&transaction.TxHash,
			&transaction.Timestamp,
			&transaction.BlockHeight,
			&transaction.TxEvents,
			&transaction.GasUsed,
			&transaction.GasWanted,
			&transaction.Fee.Amount,
			&transaction.Fee.Denom,
			&transaction.MsgTypes,
			&transaction.Success,
			&transaction.ErrorLog,
		)
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
// The response is always ordered newest-first. direction=Next walks toward older rows;
// direction=Prev walks toward newer rows (ASC scan, reversed before return).
// Fetches limit+1 internally; hasMore indicates another page exists in the walked direction.
func (t *TimescaleDb) GetTransactionsByRange(
	ctx context.Context,
	chainName string,
	cursor string,
	limit uint64,
	direction database.Direction,
) ([]*database.Transaction, bool, error) {
	if limit == 0 || limit > 100 {
		return nil, false, fmt.Errorf("limit must be between 1 and 100")
	}

	hasCursor, blockHeight, decodedTxHash, err := decodeCursor(&cursor)
	if err != nil {
		return nil, false, err
	}

	fetchLimit := limit + 1
	var (
		query   string
		args    []any
		reverse bool
	)

	switch direction {
	case database.Next:
		if hasCursor {
			query = txSelectCols + `
			WHERE tx.chain_name = $1
			AND (tx.block_height, id.tx_hash) < ($2, $3)
			ORDER BY tx.block_height DESC, id.tx_hash DESC
			LIMIT $4
			`
			args = []any{chainName, blockHeight, decodedTxHash, fetchLimit}
		} else {
			query = txSelectCols + `
			WHERE tx.chain_name = $1
			ORDER BY tx.block_height DESC, id.tx_hash DESC
			LIMIT $2
			`
			args = []any{chainName, fetchLimit}
		}
	case database.Prev:
		if !hasCursor {
			return nil, false, fmt.Errorf("prev direction requires a cursor")
		}
		query = txSelectCols + `
		WHERE tx.chain_name = $1
		AND (tx.block_height, id.tx_hash) > ($2, $3)
		ORDER BY tx.block_height ASC, id.tx_hash ASC
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
	transactions, err := scanTransactionRows(rows)
	if err != nil {
		return nil, false, err
	}

	hasMore := uint64(len(transactions)) > limit
	if hasMore {
		transactions = transactions[:limit]
	}
	if reverse {
		reverseSlice(transactions)
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

func decodeEvents(txEvents []byte) (*[]database.Event, error) {
	if len(txEvents) == 0 {
		return &[]database.Event{}, nil
	}
	decompressed, err := decompressEvents(txEvents)
	if err != nil {
		return nil, err
	}
	txEventsProto, err := protoUnmarshal(decompressed)
	if err != nil {
		return nil, err
	}
	events := make([]database.Event, 0, len(txEventsProto.Events))
	for _, event := range txEventsProto.Events {
		attributes := make([]database.Attribute, 0, len(event.Attributes))
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
			attributes = append(attributes, database.Attribute{
				Key:   attribute.Key,
				Value: value,
			})
		}
		pkgPath := ""
		if event.PkgPath != nil {
			pkgPath = *event.PkgPath
		}
		events = append(events, database.Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: attributes,
			PkgPath:    pkgPath,
		})
	}
	return &events, nil
}
