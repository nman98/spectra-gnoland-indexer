package routes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterTransactionsRoutes(api huma.API, h *handlers.TransactionsHandler) {
	huma.Get(api, "/transactions", h.GetTransactionsByCursor,
		func(op *huma.Operation) {
			op.Summary = "Get Transactions"
			op.Description = `Retrieve transactions using cursor-based pagination.
			Without a cursor you will fetch the latest data. To acquire older data, supply the cursor
			returned by a previous response. The cursor is a base64url-encoded string in the form
			timestamp|tx_hash.`
		})
	huma.Get(api, "/transactions/{tx_hash}", h.GetTransactionBasic,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction"
			op.Description = "Retrieve basic transaction data by its hash (base64url encoded)."
		})
	huma.Get(api, "/transactions/{tx_hash}/messages", h.GetTransactionMessage,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Messages"
			op.Description = "Retrieve all messages contained within a transaction by its hash."
		})
	huma.Get(api, "/transactions/stats/count/recent", h.GetTotalTxCount24h,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Count (Last 24h)"
			op.Description = "Retrieve the total transaction count for the last 24 hours."
		})
	huma.Get(api, "/transactions/stats/count/daily", h.GetTotalTxCountByDate,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Count by Day"
			op.Description = "Retrieve the transaction count per day within the given date range. Max range is 30 days."
		})
	huma.Get(api, "/transactions/stats/count/hourly", h.GetTotalTxCountByHour,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Count by Hour"
			op.Description = "Retrieve the transaction count per hour within the given datetime range. Max range is 7 days."
		})
	huma.Get(api, "/transactions/stats/volume/daily", h.GetVolumeByDate,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Volume by Day"
			op.Description = "Retrieve the transaction volume grouped by denom per day. Max range is 30 days."
		})
	huma.Get(api, "/transactions/stats/volume/hourly", h.GetVolumeByHour,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Volume by Hour"
			op.Description = "Retrieve the transaction volume grouped by denom per hour. Max range is 7 days."
		})
}
