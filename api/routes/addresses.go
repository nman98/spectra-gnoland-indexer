package routes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterAddressesRoutes(api huma.API, h *handlers.AddressHandler, m *handlers.InMemoryHandler) {
	huma.Get(api, "/addresses/{address}/transactions", h.GetAddressTxs,
		func(op *huma.Operation) {
			op.Summary = "Get Address Transactions"
			op.Description = `Retrieve transactions for a given address. Two query modes are supported:

			1. Timestamp range: specify both from_timestamp and to_timestamp. The response is ordered by sort_order.
			2. Cursor pagination: omit the timestamps. The response is always newest-first. Omit the cursor to fetch
			   the latest page. Use direction="next" with next_cursor to load older rows, and direction="prev" with
			   prev_cursor to load newer rows. Cursors have the form "<block_height>|<tx_hash_base64url>".`
		})
	huma.Get(api, "/addresses/stats/active/daily", h.GetDailyActiveAccount,
		func(op *huma.Operation) {
			op.Summary = "Get Daily Active Addresses"
			op.Description = "Retrieve the number of daily active addresses within the given date range."
		})
	huma.Get(api, "/addresses/stats/total", m.GetTotalAddressesCount,
		func(op *huma.Operation) {
			op.Summary = "Get Total Addresses Count"
			op.Description = "Retrieve the total number of addresses."
		})
}
