package routes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterAddressesRoutes(api huma.API, h *handlers.AddressHandler, m *handlers.InMemoryHandler) {
	huma.Get(api, "/addresses/{address}/transactions", h.GetAddressTxs,
		func(op *huma.Operation) {
			op.Summary = "Get Address Transactions"
			op.Description = `Retrieve all transactions for a given address.
			There are 3 ways to query the transactions:

			1. By timestamp range: specify from_timestamp and to_timestamp.
			2. By cursor: omit all parameters on the first request, then use the returned next_cursor on subsequent requests.
			3. By limit and page: specify limit and page.`
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
