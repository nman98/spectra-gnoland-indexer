package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

type ValidatorSigning24hGetInput struct {
	ValidatorAddress string `path:"validator_address" doc:"Validator consensus address" required:"true"`
}

type ValidatorSigning24hGetOutput struct {
	Body *database.ValidatorSigning
}

type ValidatorSigningByHourGetInput struct {
	ValidatorAddress string             `path:"validator_address" doc:"Validator address" required:"true" example:"g16jqn9e738pwenxpseasr49sj3axcyd37262wal"  minLength:"40" maxLength:"40"`
	StartTimestamp   time.Time          `query:"start_timestamp" doc:"Start datetime (inclusive)" format:"date-time" required:"true" example:"2026-03-12T00:00:00Z"`
	EndTimestamp     time.Time          `query:"end_timestamp" doc:"End datetime (inclusive)" format:"date-time" required:"true" example:"2026-03-13T00:00:00Z"`
	SortOrder        database.SortOrder `query:"sort_order" doc:"Sort order for results" enum:"asc,desc" default:"desc"`
}

type ValidatorSigningByHourGetOutput struct {
	Body []*database.ValidatorSigning
}

// Placeholder, there are no params
type ValidatorsListGetInput struct{}
type ValidatorListGetOutput struct {
	Body *database.ValidatorList
}
