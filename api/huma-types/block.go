package humatypes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

// BlockGetInput represents the input for getting a block by height
type BlockGetInput struct {
	Height uint64 `path:"height" minimum:"1" example:"12345" doc:"Block height to retrieve" required:"true"`
}

type FromToBlocksGetInput struct {
	FromHeight uint64 `path:"from_height" minimum:"1" example:"12345" doc:"From block height" required:"true"`
	ToHeight   uint64 `path:"to_height" minimum:"1" example:"12345" doc:"To block height" required:"true"`
}

// BlockGetOutput represents the response structure for a single block
type BlockGetOutput struct {
	Body *database.BlockData
}

type FromToBlocksGetOutput struct {
	Body []*database.BlockData
}

type AllBlockSignersGetInput struct {
	BlockHeight uint64 `path:"block_height" minimum:"1" example:"12345" doc:"Block height" required:"true"`
}

type AllBlockSignersGetOutput struct {
	Body *database.BlockSigners
}

// LatestBlockHeightGetInput represents the empty input for getting the latest block height
type LatestBlockHeightGetInput struct{}

type LatestBlockHeightGetOutput struct {
	Body *database.BlockData
}

type LastXBlocksGetInput struct {
	Amount uint64 `query:"amount" doc:"Amount of blocks to get" required:"true" min:"1" max:"100" default:"10"`
}

type LastXBlocksGetOutput struct {
	Body []*database.BlockData
}

type BlockCount24hGetInput struct{}

type BlockCount24hGetBody struct {
	Count int64 `json:"count" doc:"Count of blocks in the last 24 hours"`
}

type BlockCount24hGetOutput struct {
	Body *BlockCount24hGetBody
}

type BlockCountByDateGetInput struct {
	StartDate date.Date             `query:"start_date" doc:"Start date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	EndDate   date.Date             `query:"end_date" doc:"End date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	SortOrder database.SortOrder    `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type BlockCountByDateGetOutput struct {
	Body []*database.BlockCountByDate
}
