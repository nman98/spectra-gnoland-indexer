package handlers_test

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionsHandler_GetTransactionsByCursor_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	txHash := base64.StdEncoding.EncodeToString([]byte("tx_hash_1"))

	db := MockDatabase{
		transactions: map[string]*database.Transaction{
			txHash: {
				TxHash:      txHash,
				Timestamp:   fixedTime,
				BlockHeight: 42,
				TxEvents:    []database.Event{},
				GasUsed:     100,
				GasWanted:   100,
				Fee:         database.Amount{Amount: "100", Denom: "ugnot"},
				MsgTypes:    []string{"msg_send"},
			},
		},
	}

	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionsByCursor(
		context.Background(),
		&humatypes.TransactionGeneralListByCursorGetInput{Cursor: "", Limit: 1, Direction: database.Next},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, len(response.Body.Transactions))
	assert.Equal(t, txHash, response.Body.Transactions[0].TxHash)
	assert.False(t, response.Body.HasPrev)
	assert.Nil(t, response.Body.PrevCursor)
}

func TestTransactionsHandler_GetTransactionsByCursor_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "sql syntax error at or near SELECT",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionsByCursor(
		context.Background(),
		&humatypes.TransactionGeneralListByCursorGetInput{Cursor: "", Limit: 1, Direction: database.Next},
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "sql syntax error")
}

func TestTransactionsHandler_GetTransactionsByCursor_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "no transactions for chain",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionsByCursor(
		context.Background(),
		&humatypes.TransactionGeneralListByCursorGetInput{Cursor: "", Limit: 1, Direction: database.Next},
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "transactions not found")
}

func TestTransactionsHandler_GetTransactionsByCursor_PrevWithoutCursor(t *testing.T) {
	db := MockDatabase{}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionsByCursor(
		context.Background(),
		&humatypes.TransactionGeneralListByCursorGetInput{Cursor: "", Limit: 10, Direction: database.Prev},
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "direction=prev requires a cursor")
}

func TestTransactionsHandler_GetTransactionBasic_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	txHash := base64.RawURLEncoding.EncodeToString([]byte("tx_hash_1"))

	db := MockDatabase{
		transactions: map[string]*database.Transaction{
			txHash: {
				TxHash:      txHash,
				Timestamp:   fixedTime,
				BlockHeight: 42,
				TxEvents:    []database.Event{},
				GasUsed:     100,
				GasWanted:   100,
				Fee:         database.Amount{Amount: "100", Denom: "ugnot"},
				MsgTypes:    []string{"msg_send"},
			},
		},
	}

	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: txHash})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, txHash, response.Body.TxHash)
}

func TestTransactionsHandler_GetTransactionBasic_InternalError(t *testing.T) {
	txHash := base64.RawURLEncoding.EncodeToString([]byte("tx_hash_1"))
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "connection reset by peer",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: txHash})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "connection reset")
}

func TestTransactionsHandler_GetTransactionBasic_NotFound(t *testing.T) {
	txHash := base64.RawURLEncoding.EncodeToString([]byte("tx_hash_1"))
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "tx lookup",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: txHash})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "not found")
}

func TestTransactionsHandler_GetTransactionBasic_BadHash(t *testing.T) {
	db := MockDatabase{}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(
		context.Background(),
		// Contains a character that is invalid in URL-safe base64.
		&humatypes.TransactionGetInput{TxHash: "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"},
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "transaction hash is not valid base64url encoded")
}
