package handlers

import (
	"context"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
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
		return nil, badRequest("start_date must be before end_date")
	}
	if endDate.Sub(startDate.Time) > 24*time.Hour*30 {
		return nil, badRequest("end_date must be within 30 days of start_date")
	}

	data, err := h.db.GetDailyActiveAccount(
		ctx, h.chainName, startDate, endDate, input.SortOrder,
	)
	if err != nil {
		return nil, mapDbError("GetDailyActiveAccount", "daily active account data not found", err)
	}
	if len(data) == 0 {
		return nil, notFound("daily active account data not found")
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
	fromTs, toTs, err := parseTimestampPair(input.FromTimestamp, input.ToTimestamp)
	if err != nil {
		return nil, err
	}
	timestampMode := fromTs != nil

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
			return nil, badRequest("invalid direction (must be 'next' or 'prev')")
		}
		if direction == database.Prev && cursor == nil {
			return nil, badRequest("direction=prev requires a cursor")
		}
	}

	addressTxs, hasMore, err := h.db.GetAddressTxs(
		ctx, input.Address, h.chainName, fromTs, toTs, limit, cursor, direction,
	)
	if err != nil {
		return nil, mapDbError("GetAddressTxs", "address not found", err)
	}

	body, err := buildAddressTxsBody(*addressTxs, hasMore, direction, cursor)
	if err != nil {
		return nil, err
	}
	return &humatypes.AddressGetOutput{Body: body}, nil
}

func parseTimestampPair(from, to time.Time) (*time.Time, *time.Time, error) {
	var fromPtr, toPtr *time.Time
	if !from.IsZero() {
		fromPtr = &from
	}
	if !to.IsZero() {
		toPtr = &to
	}
	if (fromPtr == nil) != (toPtr == nil) {
		return nil, nil, badRequest("from_timestamp and to_timestamp must both be set or both be unset")
	}
	return fromPtr, toPtr, nil
}

func buildAddressTxsBody(
	rows []database.AddressTx,
	hasMore bool,
	direction database.Direction,
	cursor *string,
) (humatypes.AddressTxsBody, error) {
	body := humatypes.AddressTxsBody{AddressTxs: rows}
	if len(rows) == 0 {
		return body, nil
	}

	newestCur, err := makeTxCursor(rows[0].BlockHeight, rows[0].Hash)
	if err != nil {
		return body, internalError("GetAddressTxs.makeTxCursor", err)
	}
	oldestCur, err := makeTxCursor(rows[len(rows)-1].BlockHeight, rows[len(rows)-1].Hash)
	if err != nil {
		return body, internalError("GetAddressTxs.makeTxCursor", err)
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
	return body, nil
}
