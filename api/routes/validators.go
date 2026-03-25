package routes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterValidatorsRoutes(api huma.API, h *handlers.ValidatorsHandler) {
	huma.Get(api, "/validators/{validator_address}/signing/recent", h.GetValidatorSigning24h,
		func(op *huma.Operation) {
			op.Summary = "Get Validator Signing (Last 24h)"
			op.Description = "Retrieve the signing performance of a validator over the last 24 hours."
		})
	huma.Get(api, "/validators/{validator_address}/signing/hourly", h.GetValidatorSigningByHour,
		func(op *huma.Operation) {
			op.Summary = "Get Validator Signing by Hour"
			op.Description = "Retrieve the per-hour signing performance of a validator within the given datetime range. Max range is 7 days."
		})
	huma.Get(api, "/validators/list", h.GetValidatorsList,
		func(op *huma.Operation) {
			op.Summary = "Get List of All Validators"
			op.Description = `Retrieve the list of the validator address that were recorded by the indexer.
				Additional info, the indexer only records the validators that signed at least one block.`
		})
}
