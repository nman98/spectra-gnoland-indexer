package handlers

import (
	"context"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
)

type AddressHandler struct {
	db        AddressDbHandler
	chainName string
}

func NewAddressHandler(db AddressDbHandler, chainName string) *AddressHandler {
	return &AddressHandler{db: db, chainName: chainName}
}

func (h *AddressHandler) GetDailyActiveAccount(
	ctx context.Context,
	input *humatypes.DailyActiveAccountGetInput,
) (*humatypes.DailyActiveAccountGetOutput, error) {
	startDate := input.StartDate
	endDate := input.EndDate
	// validate input
	if !startDate.Before(endDate.Time) {
		return nil, huma.Error400BadRequest("start_date must be before end_date", nil)
	}
	if endDate.Sub(startDate.Time) > 24*time.Hour*30 {
		return nil, huma.Error400BadRequest("end_date must be within 30 days of start_date", nil)
	}

	data, err := h.db.GetDailyActiveAccount(
		ctx, h.chainName, startDate, endDate, input.SortOrder,
	)
	if err != nil {
		return nil, huma.Error404NotFound("Daily active account data not found", err)
	}
	return &humatypes.DailyActiveAccountGetOutput{Body: data}, nil
}

// GetAddressTxs returns transactions involving a given address. Two modes are
// supported:
//
//   - Timestamp range: supply both from_timestamp and to_timestamp. sort_order
//     controls the ordering; cursor/direction are ignored.
//   - Cursor pagination: leave the timestamps empty. The response is always
//     newest-first, and the caller walks history with direction=next (older
//     rows) or direction=prev (newer rows, requires a cursor). NextCursor/
//     PrevCursor are filled in so the caller can paginate in both directions.
func (h *AddressHandler) GetAddressTxs(
	ctx context.Context,
	input *humatypes.AddressGetInput,
) (*humatypes.AddressGetOutput, error) {
	var fromTs, toTs *time.Time
	if !input.FromTimestamp.IsZero() {
		fromTs = &input.FromTimestamp
	}
	if !input.ToTimestamp.IsZero() {
		toTs = &input.ToTimestamp
	}
	if (fromTs == nil) != (toTs == nil) {
		return nil, huma.Error400BadRequest("from_timestamp and to_timestamp must both be set or both be unset", nil)
	}
	timestampMode := fromTs != nil && toTs != nil

	var limit *uint64
	if input.Limit != 0 {
		limit = &input.Limit
	}
	var cursor *string
	if input.Cursor != "" {
		cursor = &input.Cursor
	}

	direction := input.Direction
	if direction == "" {
		direction = database.Next
	}
	if !timestampMode {
		if direction != database.Next && direction != database.Prev {
			return nil, huma.Error400BadRequest("Invalid direction (must be 'next' or 'prev')", nil)
		}
		if direction == database.Prev && cursor == nil {
			return nil, huma.Error400BadRequest("direction=prev requires a cursor", nil)
		}
	}

	addressTxs, hasMore, err := h.db.GetAddressTxs(
		ctx,
		input.Address,
		h.chainName,
		fromTs,
		toTs,
		limit,
		cursor,
		direction,
		input.SortOrder,
	)
	if err != nil {
		return nil, huma.Error404NotFound("Address not found", err)
	}

	body := humatypes.AddressTxsBody{
		AddressTxs: *addressTxs,
	}

	// Pagination metadata only makes sense in cursor mode; timestamp-range
	// responses are returned as-is.
	if !timestampMode && len(*addressTxs) > 0 {
		rows := *addressTxs
		newest := rows[0]
		oldest := rows[len(rows)-1]
		newestCur, err := makeTxCursor(newest.BlockHeight, newest.Hash)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to build cursor", err)
		}
		oldestCur, err := makeTxCursor(oldest.BlockHeight, oldest.Hash)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to build cursor", err)
		}

		switch direction {
		case database.Next:
			body.HasNext = hasMore
			if hasMore {
				body.NextCursor = &oldestCur
			}
			if cursor != nil {
				body.HasPrev = true
				body.PrevCursor = &newestCur
			}
		case database.Prev:
			body.HasPrev = hasMore
			if hasMore {
				body.PrevCursor = &newestCur
			}
			body.HasNext = true
			body.NextCursor = &oldestCur
		}
	}

	return &humatypes.AddressGetOutput{
		Body: body,
	}, nil
}
