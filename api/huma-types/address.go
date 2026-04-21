package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type AddressGetInput struct {
	Address       string             `path:"address" doc:"Gno address you want to query" required:"true" minLength:"40" maxLength:"40"`
	FromTimestamp time.Time          `query:"from_timestamp" doc:"From timestamp (inclusive). Must be paired with to_timestamp for timestamp-range mode." format:"date-time"`
	ToTimestamp   time.Time          `query:"to_timestamp" doc:"To timestamp (inclusive). Must be paired with from_timestamp for timestamp-range mode." format:"date-time"`
	Limit         uint64             `query:"limit" doc:"Limit of transactions to return" min:"1" max:"100" default:"10"`
	Cursor        string             `query:"cursor" doc:"Cursor in the form '<block_height>|<tx_hash_base64url>'. Used in cursor mode only." required:"false"`
	Direction     database.Direction `query:"direction" doc:"Direction to walk from the cursor. 'next' returns older rows; 'prev' returns newer rows and requires a cursor. Used in cursor mode only." enum:"next,prev" default:"next"`
	SortOrder     database.SortOrder `query:"sort_order" doc:"Sort order for timestamp-range mode. Ignored in cursor mode (which is always newest-first)." enum:"asc,desc" default:"desc"`
}

type AddressGetOutput struct {
	Body AddressTxsBody
}

// AddressTxsBody is the response body for the address transactions endpoint.
// In timestamp-range mode only AddressTxs is populated. In cursor mode the
// pagination fields mirror the transactions range API: transactions are
// newest-first, NextCursor points at the oldest row (older page) and
// PrevCursor points at the newest row (newer page).
type AddressTxsBody struct {
	AddressTxs []database.AddressTx `json:"address_txs" doc:"Data about address transactions"`
	HasNext    bool                 `json:"has_next" doc:"True when an older page exists (cursor mode only)"`
	HasPrev    bool                 `json:"has_prev" doc:"True when a newer page exists (cursor mode only)"`
	NextCursor *string              `json:"next_cursor" doc:"Cursor to request the next (older) page"`
	PrevCursor *string              `json:"prev_cursor" doc:"Cursor to request the previous (newer) page"`
}

type DailyActiveAccountGetInput struct {
	StartDate date.Date          `query:"start_date" doc:"Start date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	EndDate   date.Date          `query:"end_date" doc:"End date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	SortOrder database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type DailyActiveAccountGetOutput struct {
	Body []*database.DailyActiveAccount
}
