package routes

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterUtilsRoutes(api huma.API) {
	huma.Post(api, "/utils/base64url/encode", handlers.ConvertFromBase64toBase64Url,
		func(op *huma.Operation) {
			op.Summary = "Encode Base64 to Base64URL"
			op.Description = "Convert a standard base64-encoded tx hash to a base64url-encoded tx hash."
		})
	huma.Post(api, "/utils/base64url/decode", handlers.ConvertFromBase64UrlToBase64,
		func(op *huma.Operation) {
			op.Summary = "Decode Base64URL to Base64"
			op.Description = "Convert a base64url-encoded tx hash to a standard base64-encoded tx hash."
		})
}
