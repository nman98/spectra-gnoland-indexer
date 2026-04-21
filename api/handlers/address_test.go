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

func TestAddressHandler_GetAddressTxs_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Hashes must be valid std-base64 so the handler can re-encode them for the cursor.
	hash1 := base64.StdEncoding.EncodeToString([]byte("tx_hash_1"))
	hash2 := base64.StdEncoding.EncodeToString([]byte("tx_hash_2"))
	hash3 := base64.StdEncoding.EncodeToString([]byte("tx_hash_3"))

	addressTxData := []database.AddressTx{
		{Hash: hash1, Timestamp: fixedTime, MsgTypes: []string{"msg_send"}, BlockHeight: 3},
		{Hash: hash2, Timestamp: fixedTime, MsgTypes: []string{"msg_send"}, BlockHeight: 2},
		{Hash: hash3, Timestamp: fixedTime, MsgTypes: []string{"msg_send"}, BlockHeight: 1},
	}

	db := MockDatabase{
		addressTxs: map[string]*[]database.AddressTx{
			"gno_address_1": &addressTxData,
		},
	}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{
		Address:   "gno_address_1",
		Limit:     10,
		Direction: database.Next,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 3, len(response.Body.AddressTxs))
	assert.Equal(t, hash1, response.Body.AddressTxs[0].Hash)
	// No cursor supplied and hasMore=false in the mock, so no pagination fields.
	assert.False(t, response.Body.HasNext)
	assert.False(t, response.Body.HasPrev)
	assert.Nil(t, response.Body.NextCursor)
	assert.Nil(t, response.Body.PrevCursor)
}

func TestAddressHandler_GetAddressTxs_Fail(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "error getting address transactions",
	}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{
		Address:   "gno_address_1",
		Limit:     10,
		Direction: database.Next,
	})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "Address not found")
}

func TestAddressHandler_GetAddressTxs_PrevWithoutCursor(t *testing.T) {
	db := MockDatabase{}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{
		Address:   "gno_address_1",
		Limit:     10,
		Direction: database.Prev,
	})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "direction=prev requires a cursor")
}

func TestAddressHandler_GetAddressTxs_UnpairedTimestamp(t *testing.T) {
	db := MockDatabase{}
	handler := handlers.NewAddressHandler(&db, "gnoland")
	response, err := handler.GetAddressTxs(context.Background(), &humatypes.AddressGetInput{
		Address:       "gno_address_1",
		FromTimestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Limit:         10,
		Direction:     database.Next,
	})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "from_timestamp and to_timestamp must both be set or both be unset")
}
