package humatypes

type GetAvgBlockProdTimeInput struct{}

type GetTotalAddressesCountInput struct{}

type GetAvgBlockBody struct {
	AvgBlockProdTime float64 `json:"avg_block_prod_time" doc:"Average block production time"`
}

type GetTotalAddressesCountBody struct {
	TotalAddressesCount int32 `json:"total_addresses_count" doc:"Total addresses count"`
}

type GetTotalAddressesCountOutput struct {
	Body GetTotalAddressesCountBody
}

type GetAvgBlockProdTimeOutput struct {
	Body GetAvgBlockBody
}
