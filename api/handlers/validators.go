package handlers

import (
	"context"
	"fmt"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/danielgtaylor/huma/v2"
)

type ValidatorsHandler struct {
	db        ValidatorDbHandler
	chainName string
}

func NewValidatorsHandler(db ValidatorDbHandler, chainName string) *ValidatorsHandler {
	return &ValidatorsHandler{db: db, chainName: chainName}
}

// GetValidatorSigning24h returns the signing performance of a validator over the last 24 hours
func (h *ValidatorsHandler) GetValidatorSigning24h(
	ctx context.Context,
	input *humatypes.ValidatorSigning24hGetInput,
) (*humatypes.ValidatorSigning24hGetOutput, error) {
	if input.ValidatorAddress == "" {
		return nil, huma.Error400BadRequest("validator_address is required", nil)
	}
	signing, err := h.db.GetValidatorSigning24h(ctx, input.ValidatorAddress, h.chainName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Signing data for validator %s not found", input.ValidatorAddress), err)
	}
	return &humatypes.ValidatorSigning24hGetOutput{Body: signing}, nil
}

// GetValidatorSigningByHour returns the per-hour signing performance of a validator within the given datetime range
func (h *ValidatorsHandler) GetValidatorSigningByHour(
	ctx context.Context,
	input *humatypes.ValidatorSigningByHourGetInput,
) (*humatypes.ValidatorSigningByHourGetOutput, error) {
	if input.ValidatorAddress == "" {
		return nil, huma.Error400BadRequest("validator_address is required", nil)
	}
	if !input.StartTimestamp.Before(input.EndTimestamp) {
		return nil, huma.Error400BadRequest("start_timestamp must be before end_timestamp", nil)
	}
	if input.EndTimestamp.Sub(input.StartTimestamp) > 24*time.Hour*7 { // 7 days
		return nil, huma.Error400BadRequest("end_timestamp must be within 7 days of start_timestamp", nil)
	}

	signing, err := h.db.GetValidatorSigningByHour(
		ctx, input.ValidatorAddress, h.chainName, input.StartTimestamp, input.EndTimestamp, input.SortOrder,
	)
	if err != nil {
		return nil, huma.Error404NotFound(
			fmt.Sprintf("Signing data for validator %s in the given time range not found", input.ValidatorAddress), err)
	}
	return &humatypes.ValidatorSigningByHourGetOutput{Body: signing}, nil
}
