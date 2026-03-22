package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

type AddressGetInput struct {
	Address       string    `path:"address" doc:"Gno address you want to query" required:"true" minLength:"40" maxLength:"40"`
	FromTimestamp time.Time `query:"from_timestamp" doc:"From timestamp (inclusive)" format:"date-time"`
	ToTimestamp   time.Time `query:"to_timestamp" doc:"To timestamp (inclusive)" format:"date-time"`
	Limit         uint64    `query:"limit" doc:"Limit of transactions to return" min:"1" max:"100" default:"10"`
	Page          uint64    `query:"page" doc:"Page of transactions to return"`
	Cursor        string    `query:"cursor" doc:"Cursor to continue from"`
}

type AddressGetOutput struct {
	Body AddressTxsBody
}

type AddressTxsBody struct {
	AddressTxs []database.AddressTx `json:"address_txs" doc:"Data about address transactions"`
	TxCount    uint64               `json:"tx_count" doc:"Total number of transactions"`
	NextCursor string               `json:"next_cursor" doc:"Next cursor that can be used in the query"`
}

type DailyActiveAccountGetInput struct {
	StartDate date.Date `query:"start_date" doc:"Start date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
	EndDate   date.Date `query:"end_date" doc:"End date (inclusive, YYYY-MM-DD)" format:"date" required:"true"`
}

type DailyActiveAccountGetOutput struct {
	Body []*database.DailyActiveAccount
}
