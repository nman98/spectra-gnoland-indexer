package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

type TransactionsHandler struct {
	db        TransactionDbHandler
	chainName string
}

func NewTransactionsHandler(db TransactionDbHandler, chainName string) *TransactionsHandler {
	return &TransactionsHandler{db: db, chainName: chainName}
}

// GetTransactionBasic retrieves basic transaction details by tx hash
func (h *TransactionsHandler) GetTransactionBasic(
	ctx context.Context,
	input *humatypes.TransactionGetInput,
) (*humatypes.TransactionBasicGetOutput, error) {
	input.TxHash = strings.Trim(input.TxHash, " ")
	txHash, err := base64.URLEncoding.DecodeString(input.TxHash)
	if err != nil {
		return nil, badRequest("transaction hash is not valid base64url encoded")
	}
	txHashBase64 := base64.StdEncoding.EncodeToString(txHash)
	transaction, err := h.db.GetTransaction(ctx, txHashBase64, h.chainName)
	if err != nil {
		return nil, mapDbError(
			"GetTransaction",
			fmt.Sprintf("transaction with hash %s not found", input.TxHash),
			err,
		)
	}
	return &humatypes.TransactionBasicGetOutput{
		Body: *transaction,
	}, nil
}

// GetTransactionMessage retrieves all messages within a transaction by tx hash
func (h *TransactionsHandler) GetTransactionMessage(
	ctx context.Context,
	input *humatypes.TransactionGetInput,
) (*humatypes.TransactionMessageGetOutput, error) {
	input.TxHash = strings.Trim(input.TxHash, " ")
	txHash, err := base64.URLEncoding.DecodeString(input.TxHash)
	if err != nil {
		return nil, badRequest("transaction hash is not valid base64url encoded")
	}
	txHashBase64 := base64.StdEncoding.EncodeToString(txHash)
	response := make(map[int16]humatypes.TransactionMessage)
	msgTypes, err := h.db.GetMsgTypes(ctx, txHashBase64, h.chainName)
	if err != nil {
		return nil, mapDbError(
			"GetMsgTypes",
			fmt.Sprintf("transaction with hash %s not found", input.TxHash),
			err,
		)
	}

	for _, msgType := range msgTypes {
		switch msgType {
		case "bank_msg_send":
			if err := h.getBankSendResponse(ctx, msgType, txHashBase64, h.chainName, &response); err != nil {
				return nil, err
			}
		case "vm_msg_call":
			if err := h.getMsgCallResponse(ctx, msgType, txHashBase64, h.chainName, &response); err != nil {
				return nil, err
			}
		case "vm_msg_add_package":
			if err := h.getMsgAddPackageResponse(ctx, msgType, txHashBase64, h.chainName, &response); err != nil {
				return nil, err
			}
		case "vm_msg_run":
			if err := h.getMsgRunResponse(ctx, msgType, txHashBase64, h.chainName, &response); err != nil {
				return nil, err
			}
		default:
			// An unknown message type coming out of the database is a server-side
			// integrity problem, not something the caller can fix. Log it and
			// return a generic 500 so we don't expose the internal type name.
			return nil, internalError(
				"GetTransactionMessage",
				fmt.Errorf("unknown message type %q for tx %s", msgType, input.TxHash),
			)
		}
	}
	return &humatypes.TransactionMessageGetOutput{
		Body: response,
	}, nil
}

// GetTransactionsByCursor returns a page of transactions using keyset (cursor) pagination.
//
// The response is always newest-first: transactions[0] is the newest row on the page and
// transactions[len-1] is the oldest. Cursors are built from the (block_height, tx_hash)
// pair of boundary rows so the caller can walk the history in either direction:
//   - NextCursor points at the oldest row; use it with direction=next to load older data.
//   - PrevCursor points at the newest row; use it with direction=prev to load newer data.
func (h *TransactionsHandler) GetTransactionsByCursor(
	ctx context.Context,
	input *humatypes.TransactionGeneralListByCursorGetInput,
) (*humatypes.TransactionGeneralListByCursorGetOutput, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 25
	}
	if limit > 100 {
		return nil, badRequest("invalid limit (1..100)")
	}

	direction := input.Direction
	if direction == "" {
		direction = database.Next
	}
	if direction != database.Next && direction != database.Prev {
		return nil, badRequest("invalid direction (must be 'next' or 'prev')")
	}
	if direction == database.Prev && input.Cursor == "" {
		return nil, badRequest("direction=prev requires a cursor")
	}

	transactions, hasMore, err := h.db.GetTransactionsByRange(
		ctx, h.chainName, input.Cursor, limit, direction,
	)
	if err != nil {
		return nil, mapDbError("GetTransactionsByRange", "transactions not found", err)
	}

	body := humatypes.TransactionsRangeBody{
		Transactions: transactions,
	}
	if len(transactions) > 0 {
		newest := transactions[0]
		oldest := transactions[len(transactions)-1]
		newestCur, err := makeTxCursor(newest.BlockHeight, newest.TxHash)
		if err != nil {
			return nil, internalError("GetTransactionsByCursor.makeTxCursor", err)
		}
		oldestCur, err := makeTxCursor(oldest.BlockHeight, oldest.TxHash)
		if err != nil {
			return nil, internalError("GetTransactionsByCursor.makeTxCursor", err)
		}

		switch direction {
		case database.Next:
			body.HasNext = hasMore
			if hasMore {
				body.NextCursor = &oldestCur
			}
			// A prev page exists only when the caller supplied a cursor, since
			// the initial fetch (no cursor) already starts at the head.
			if input.Cursor != "" {
				body.HasPrev = true
				body.PrevCursor = &newestCur
			}
		case database.Prev:
			// We walked toward the head: hasMore means newer rows still remain
			// between this page and the head.
			body.HasPrev = hasMore
			if hasMore {
				body.PrevCursor = &newestCur
			}
			// A prev call implies the caller was already deeper in history, so
			// older rows always exist past the oldest row on this page.
			body.HasNext = true
			body.NextCursor = &oldestCur
		}
	}

	return &humatypes.TransactionGeneralListByCursorGetOutput{
		Body: body,
	}, nil
}

// makeTxCursor encodes a (block_height, tx_hash) pair into the "<height>|<hash>" form
// used by the transactions range API. The tx hash is received as standard base64 and
// re-encoded as URL-safe base64 so the cursor is safe to pass as a query parameter.
func makeTxCursor(blockHeight uint64, txHashB64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(txHashB64)
	if err != nil {
		return "", fmt.Errorf("error decoding tx hash: %w", err)
	}
	return strconv.FormatUint(blockHeight, 10) + "|" + base64.URLEncoding.Strict().EncodeToString(raw), nil
}

// GetLastXTransactions retrieves the most recent X transactions
func (h *TransactionsHandler) GetLastXTransactions(
	ctx context.Context,
	input *humatypes.LastXTransactionsGetInput,
) (*humatypes.LastXTransactionsGetOutput, error) {
	transactions, err := h.db.GetLastXTransactions(ctx, h.chainName, input.Amount, nil)
	if err != nil {
		return nil, mapDbError("GetLastXTransactions", "no recent transactions found", err)
	}
	return &humatypes.LastXTransactionsGetOutput{Body: transactions}, nil
}

// GetTotalTxCount24h returns the total number of transactions in the last 24 hours
func (h *TransactionsHandler) GetTotalTxCount24h(
	ctx context.Context,
	input *humatypes.TotalTxCount24hGetInput,
) (*humatypes.TotalTxCount24hGetOutput, error) {
	count, err := h.db.GetTotalTxCount24h(ctx, h.chainName)
	if err != nil {
		return nil, mapDbError("GetTotalTxCount24h", "transaction count for last 24h not found", err)
	}
	body := &humatypes.TotalTxCount24hBody{Count: count}
	return &humatypes.TotalTxCount24hGetOutput{Body: body}, nil
}

// GetTotalTxCountByDate returns the transaction count per day within the given date range
func (h *TransactionsHandler) GetTotalTxCountByDate(
	ctx context.Context,
	input *humatypes.TxCountByDateGetInput,
) (*humatypes.TxCountByDateGetOutput, error) {
	startDate := input.StartDate
	endDate := input.EndDate
	if !startDate.Before(endDate.Time) {
		return nil, badRequest("start_date must be before end_date")
	}
	if endDate.Sub(startDate.Time) > 24*time.Hour*30 {
		return nil, badRequest("end_date must be within 30 days of start_date")
	}

	counts, err := h.db.GetTotalTxCountByDate(ctx, h.chainName, startDate, endDate, input.SortOrder)
	if err != nil {
		return nil, mapDbError(
			"GetTotalTxCountByDate",
			"transaction count for the given date range not found",
			err,
		)
	}
	if len(counts) == 0 {
		return nil, notFound("transaction count for the given date range not found")
	}
	return &humatypes.TxCountByDateGetOutput{Body: counts}, nil
}

// GetTotalTxCountByHour returns the transaction count per hour within the given datetime range
func (h *TransactionsHandler) GetTotalTxCountByHour(
	ctx context.Context,
	input *humatypes.TxCountByHourGetInput,
) (*humatypes.TxCountByHourGetOutput, error) {
	if !input.StartTimestamp.Before(input.EndTimestamp) {
		return nil, badRequest("start_timestamp must be before end_timestamp")
	}
	if input.EndTimestamp.Sub(input.StartTimestamp) > 24*time.Hour*7 { // 7 days
		return nil, badRequest("end_timestamp must be within 7 days of start_timestamp")
	}

	counts, err := h.db.GetTotalTxCountByHour(ctx, h.chainName, input.StartTimestamp, input.EndTimestamp, input.SortOrder)
	if err != nil {
		return nil, mapDbError(
			"GetTotalTxCountByHour",
			"transaction count for the given time range not found",
			err,
		)
	}
	if len(counts) == 0 {
		return nil, notFound("transaction count for the given time range not found")
	}
	return &humatypes.TxCountByHourGetOutput{Body: counts}, nil
}

// GetVolumeByDate returns the transaction volume grouped by denom per day within the given date range
func (h *TransactionsHandler) GetVolumeByDate(ctx context.Context, input *humatypes.VolumeByDateGetInput) (*humatypes.VolumeByDateGetOutput, error) {
	if !input.StartDate.Before(input.EndDate.Time) {
		return nil, badRequest("start_date must be before end_date")
	}
	if input.StartDate.Sub(input.EndDate.Time) > 24*time.Hour*30 {
		return nil, badRequest("end_date must be within 30 days of start_date")
	}

	volume, err := h.db.GetVolumeByDate(ctx, h.chainName, input.StartDate, input.EndDate, input.SortOrder)
	if err != nil {
		return nil, mapDbError(
			"GetVolumeByDate",
			"volume for the given date range not found",
			err,
		)
	}
	if len(volume) == 0 {
		return nil, notFound("volume for the given date range not found")
	}
	return &humatypes.VolumeByDateGetOutput{Body: volume}, nil
}

// GetVolumeByHour returns the transaction volume grouped by denom per hour within the given datetime range
func (h *TransactionsHandler) GetVolumeByHour(ctx context.Context, input *humatypes.VolumeByHourGetInput) (*humatypes.VolumeByHourGetOutput, error) {
	if !input.StartTimestamp.Before(input.EndTimestamp) {
		return nil, badRequest("start_timestamp must be before end_timestamp")
	}
	if input.EndTimestamp.Sub(input.StartTimestamp) > 24*time.Hour*7 { // 7 days
		return nil, badRequest("end_timestamp must be within 7 days of start_timestamp")
	}

	volume, err := h.db.GetVolumeByHour(ctx, h.chainName, input.StartTimestamp, input.EndTimestamp, input.SortOrder)
	if err != nil {
		return nil, mapDbError(
			"GetVolumeByHour",
			"volume for the given time range not found",
			err,
		)
	}
	if len(volume) == 0 {
		return nil, notFound("volume for the given time range not found")
	}
	return &humatypes.VolumeByHourGetOutput{Body: volume}, nil
}

// Helper method that collects msg call data from the database and adds it to the response
func (h *TransactionsHandler) getMsgCallResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgCall(ctx, txHash, chainName)
	if err != nil {
		return internalError("GetMsgCall", err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      d.TxHash,
			Timestamp:   d.Timestamp,
			Signers:     d.Signers,
			Caller:      d.Caller,
			Send:        d.Send,
			PkgPath:     d.PkgPath,
			FuncName:    d.FuncName,
			Args:        d.Args,
			MaxDeposit:  d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects add package data from the database and adds it to the response
func (h *TransactionsHandler) getMsgAddPackageResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgAddPackage(ctx, txHash, chainName)
	if err != nil {
		return internalError("GetMsgAddPackage", err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       d.TxHash,
			Timestamp:    d.Timestamp,
			Signers:      d.Signers,
			Creator:      d.Creator,
			PkgPath:      d.PkgPath,
			PkgName:      d.PkgName,
			PkgFileNames: d.PkgFileNames,
			Send:         d.Send,
			MaxDeposit:   d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects msg run data from the database and adds it to the response
func (h *TransactionsHandler) getMsgRunResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetMsgRun(ctx, txHash, chainName)
	if err != nil {
		return internalError("GetMsgRun", err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType:  msgType,
			TxHash:       d.TxHash,
			Timestamp:    d.Timestamp,
			Signers:      d.Signers,
			Caller:       d.Caller,
			PkgPath:      d.PkgPath,
			PkgName:      d.PkgName,
			PkgFileNames: d.PkgFileNames,
			Send:         d.Send,
			MaxDeposit:   d.MaxDeposit,
		}
	}
	return nil
}

// Helper method that collects bank send data from the database and adds it to the response
func (h *TransactionsHandler) getBankSendResponse(
	ctx context.Context,
	msgType string,
	txHash string,
	chainName string,
	response *map[int16]humatypes.TransactionMessage,
) error {
	data, err := h.db.GetBankSend(ctx, txHash, chainName)
	if err != nil {
		return internalError("GetBankSend", err)
	}
	for _, d := range data {
		index := d.MessageCounter
		(*response)[index] = humatypes.TransactionMessage{
			MessageType: msgType,
			TxHash:      d.TxHash,
			Timestamp:   d.Timestamp,
			Signers:     d.Signers,
			FromAddress: d.FromAddress,
			ToAddress:   d.ToAddress,
			Amount:      d.Amount,
		}
	}
	return nil
}
