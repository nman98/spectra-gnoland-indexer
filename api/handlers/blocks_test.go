package handlers_test

import (
	"context"
	"testing"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlocksHandler_GetBlock_Success(t *testing.T) {
	db := MockDatabase{
		blocks: map[uint64]*database.BlockData{
			42: {Height: 42, Hash: "abc123"},
		},
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetBlock(context.Background(), &humatypes.BlockGetInput{Height: 42})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	require.Equal(t, uint64(42), response.Body.Height)
}

func TestBlocksHandler_GetBlock_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "pgx: timeout",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetBlock(context.Background(), &humatypes.BlockGetInput{Height: 42})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "pgx: timeout")
}

func TestBlocksHandler_GetBlock_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "block lookup",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetBlock(context.Background(), &humatypes.BlockGetInput{Height: 42})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "block at height 42 not found")
}

func TestBlocksHandler_GetFromToBlocks_Success(t *testing.T) {
	db := MockDatabase{
		blocks: map[uint64]*database.BlockData{
			42: {Height: 42, Hash: "abc123"},
			43: {Height: 43, Hash: "def456"},
		},
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetFromToBlocks(context.Background(), &humatypes.FromToBlocksGetInput{FromHeight: 42, ToHeight: 43})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.Body))
	assert.Equal(t, uint64(42), response.Body[0].Height)
	assert.Equal(t, uint64(43), response.Body[1].Height)
}

func TestBlocksHandler_GetFromToBlocks_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "relation blocks does not exist",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetFromToBlocks(context.Background(), &humatypes.FromToBlocksGetInput{FromHeight: 42, ToHeight: 43})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "relation blocks")
}

func TestBlocksHandler_GetFromToBlocks_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "block range empty",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetFromToBlocks(context.Background(), &humatypes.FromToBlocksGetInput{FromHeight: 42, ToHeight: 43})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "blocks from height 42 to 43 not found")
}

func TestBlocksHandler_GetAllBlockSigners_Success(t *testing.T) {
	db := MockDatabase{
		blockSigners: map[uint64]*database.BlockSigners{
			42: {BlockHeight: 42, Proposer: "abc123", SignedVals: []string{"val1", "val2"}},
		},
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetAllBlockSigners(context.Background(), &humatypes.AllBlockSignersGetInput{BlockHeight: 42})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	require.Equal(t, database.BlockSigners{BlockHeight: 42, Proposer: "abc123", SignedVals: []string{"val1", "val2"}}, *response.Body)
}

func TestBlocksHandler_GetAllBlockSigners_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "deadlock detected",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetAllBlockSigners(context.Background(), &humatypes.AllBlockSignersGetInput{BlockHeight: 42})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "deadlock")
}

func TestBlocksHandler_GetAllBlockSigners_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "signers lookup",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetAllBlockSigners(context.Background(), &humatypes.AllBlockSignersGetInput{BlockHeight: 42})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "block signers at height 42 not found")
}

func TestBlocksHandler_GetLatestBlockHeight_Success(t *testing.T) {
	db := MockDatabase{
		latestBlock: &database.BlockData{Height: 42, Hash: "abc123"},
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLatestBlock(context.Background(), &humatypes.LatestBlockHeightGetInput{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	require.Equal(t, database.BlockData{Height: 42, Hash: "abc123"}, *response.Body)
}

func TestBlocksHandler_GetLatestBlockHeight_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "pool exhausted",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLatestBlock(context.Background(), &humatypes.LatestBlockHeightGetInput{})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "pool exhausted")
}

func TestBlocksHandler_GetLatestBlockHeight_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "latest block lookup",
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLatestBlock(context.Background(), &humatypes.LatestBlockHeightGetInput{})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "latest block not found")
}

func TestBlocksHandler_GetLastXBlocks_Success(t *testing.T) {
	db := MockDatabase{
		blocks: map[uint64]*database.BlockData{
			42: {Height: 42, Hash: "abc123"},
			43: {Height: 43, Hash: "def456"},
		},
	}
	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLastXBlocks(context.Background(), &humatypes.LastXBlocksGetInput{Amount: 2})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.Body))
	// Just verify the blocks are present, don't do exact comparison
	// because map iteration order is random
	assert.NotEmpty(t, response.Body)
}

func TestBlocksHandler_GetLastXBlocks_InternalError(t *testing.T) {
	db := MockDatabase{
		shouldError: true,
		errorMsg:    "tls handshake failed",
	}

	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLastXBlocks(context.Background(), &humatypes.LastXBlocksGetInput{Amount: 2})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "internal server error")
	assert.NotContains(t, err.Error(), "tls handshake")
}

func TestBlocksHandler_GetLastXBlocks_NotFound(t *testing.T) {
	db := MockDatabase{
		shouldError:   true,
		notFoundError: true,
		errorMsg:      "no blocks",
	}

	handler := handlers.NewBlocksHandler(&db, "gnoland")
	response, err := handler.GetLastXBlocks(context.Background(), &humatypes.LastXBlocksGetInput{Amount: 2})

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "no recent blocks found")
}
