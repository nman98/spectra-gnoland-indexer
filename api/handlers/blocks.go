package handlers

import (
	"context"
	"fmt"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
)

// BlocksHandler handles block-related API requests
type BlocksHandler struct {
	db        BlockDbHandler
	chainName string
}

// NewBlocksHandler creates a new blocks handler
func NewBlocksHandler(db BlockDbHandler, chainName string) *BlocksHandler {
	return &BlocksHandler{db: db, chainName: chainName}
}

// GetBlock retrieves a block by height
func (h *BlocksHandler) GetBlock(ctx context.Context, input *humatypes.BlockGetInput) (*humatypes.BlockGetOutput, error) {
	block, err := h.db.GetBlock(ctx, input.Height, h.chainName)
	if err != nil {
		return nil, mapDbError(
			"GetBlock",
			fmt.Sprintf("block at height %d not found", input.Height),
			err,
		)
	}

	return &humatypes.BlockGetOutput{Body: block}, nil
}

// Get from block height a to block height b
func (h *BlocksHandler) GetFromToBlocks(
	ctx context.Context,
	input *humatypes.FromToBlocksGetInput,
) (*humatypes.FromToBlocksGetOutput, error) {
	if input.FromHeight > input.ToHeight {
		return nil, badRequest("from_height must be less than or equal to to_height")
	}
	if input.ToHeight-input.FromHeight > 100 {
		return nil, badRequest("from_height and to_height difference must be less than 100")
	}

	blocks, err := h.db.GetFromToBlocks(ctx, input.FromHeight, input.ToHeight, h.chainName)
	if err != nil {
		return nil, mapDbError(
			"GetFromToBlocks",
			fmt.Sprintf("blocks from height %d to %d not found", input.FromHeight, input.ToHeight),
			err,
		)
	}
	if len(blocks) == 0 {
		return nil, notFound(fmt.Sprintf(
			"blocks from height %d to %d not found", input.FromHeight, input.ToHeight,
		))
	}
	return &humatypes.FromToBlocksGetOutput{Body: blocks}, nil
}

func (h *BlocksHandler) GetAllBlockSigners(
	ctx context.Context,
	input *humatypes.AllBlockSignersGetInput,
) (*humatypes.AllBlockSignersGetOutput, error) {
	blockSigners, err := h.db.GetAllBlockSigners(ctx, h.chainName, input.BlockHeight)
	if err != nil {
		return nil, mapDbError(
			"GetAllBlockSigners",
			fmt.Sprintf("block signers at height %d not found", input.BlockHeight),
			err,
		)
	}
	return &humatypes.AllBlockSignersGetOutput{Body: blockSigners}, nil
}

// Get latest block height
func (h *BlocksHandler) GetLatestBlock(ctx context.Context, _ *humatypes.LatestBlockHeightGetInput) (*humatypes.LatestBlockHeightGetOutput, error) {
	block, err := h.db.GetLatestBlock(ctx, h.chainName)
	if err != nil {
		return nil, mapDbError("GetLatestBlock", "latest block not found", err)
	}
	return &humatypes.LatestBlockHeightGetOutput{Body: block}, nil
}

// Get last x blocks
func (h *BlocksHandler) GetLastXBlocks(ctx context.Context, input *humatypes.LastXBlocksGetInput) (*humatypes.LastXBlocksGetOutput, error) {
	blocks, err := h.db.GetLastXBlocks(ctx, h.chainName, input.Amount)
	if err != nil {
		return nil, mapDbError("GetLastXBlocks", "no recent blocks found", err)
	}
	return &humatypes.LastXBlocksGetOutput{Body: blocks}, nil
}

// GetBlockCount24h returns the total number of blocks produced in the last 24 hours
func (h *BlocksHandler) GetBlockCount24h(
	ctx context.Context,
	_ *humatypes.BlockCount24hGetInput,
) (*humatypes.BlockCount24hGetOutput, error) {
	count, err := h.db.GetBlockCount24h(ctx, h.chainName)
	if err != nil {
		return nil, mapDbError("GetBlockCount24h", "block count for last 24h not found", err)
	}
	body := &humatypes.BlockCount24hGetBody{Count: count}
	return &humatypes.BlockCount24hGetOutput{Body: body}, nil
}

// GetBlockCountByDate returns the block count per day within the given date range
func (h *BlocksHandler) GetBlockCountByDate(
	ctx context.Context,
	input *humatypes.BlockCountByDateGetInput,
) (*humatypes.BlockCountByDateGetOutput, error) {
	startDate := input.StartDate
	endDate := input.EndDate
	if !startDate.Before(endDate.Time) {
		return nil, badRequest("start_date must be before end_date")
	}
	if endDate.Sub(startDate.Time) > 24*time.Hour*30 {
		return nil, badRequest("end_date must be within 30 days of start_date")
	}

	counts, err := h.db.GetBlockCountByDate(
		ctx, h.chainName, startDate, endDate, input.SortOrder,
	)
	if err != nil {
		return nil, mapDbError(
			"GetBlockCountByDate",
			"block count for the given date range not found",
			err,
		)
	}
	if len(counts) == 0 {
		return nil, notFound("block count for the given date range not found")
	}
	return &humatypes.BlockCountByDateGetOutput{Body: counts}, nil
}
