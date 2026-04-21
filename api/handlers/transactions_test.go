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

func TestTransactionsHandler_GetTransactionsByCursor_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting transactions by cursor",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionsByCursor(
		context.Background(),
		&humatypes.TransactionGeneralListByCursorGetInput{Cursor: "", Limit: 1, Direction: database.Next},
	)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "Transactions by cursor not found")
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

func TestTransactionsHandler_GetTransactionBasic_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting transaction basic",
	}
	handler := handlers.NewTransactionsHandler(&db, "gnoland")
	response, err := handler.GetTransactionBasic(context.Background(), &humatypes.TransactionGetInput{TxHash: "invalid"})

	assert.Error(t, err)
	assert.Nil(t, response)
}
