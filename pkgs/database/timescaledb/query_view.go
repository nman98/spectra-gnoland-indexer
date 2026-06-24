package timescaledb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
	"github.com/jackc/pgx/v5"
)

func (t *TimescaleDb) GetBlockCountByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.BlockCountByDate, error) {

	query := fmt.Sprintf(`
	SELECT date::date, block_count FROM (
	SELECT
		time_bucket_gapfill('1 day', time_bucket) as date,
		coalesce(SUM(block_count), 0) as block_count
		FROM block_counter
		WHERE chain_name = $1
		AND time_bucket >= $2 and time_bucket <= $3
		GROUP BY 1
	) sum
	ORDER BY date %s
	`, sortOrder.SQL())

	newDateTo := dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	rows, err := t.pool.Query(ctx, query, chainName, dateFrom, newDateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blockCountByDates []*database.BlockCountByDate
	for rows.Next() {
		blockCountByDate := &database.BlockCountByDate{}
		err := rows.Scan(&blockCountByDate.Date, &blockCountByDate.Count)
		if err != nil {
			return nil, err
		}
		blockCountByDates = append(blockCountByDates, blockCountByDate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blockCountByDates, nil
}

func (t *TimescaleDb) GetBlockCount24h(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	coalesce(SUM(block_count), 0) as block_count
	FROM block_counter
	WHERE chain_name = $1
	AND time_bucket >= NOW() - INTERVAL '24 hours'
	AND time_bucket < NOW()
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var blockCount int64
	err := row.Scan(&blockCount)
	if err != nil {
		return 0, err
	}
	return blockCount, nil
}

func (t *TimescaleDb) GetDailyActiveAccount(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.DailyActiveAccount, error) {
	query := fmt.Sprintf(`
	SELECT date::date, count FROM (
	SELECT
		time_bucket_gapfill('1 day', time_bucket) as date,
		coalesce(SUM(active_account_count), 0) as count
		FROM daily_active_accounts
		WHERE chain_name = $1
		AND time_bucket >= $2 AND time_bucket <= $3
		GROUP BY 1
	) sub
	ORDER BY date %s
	`, sortOrder.SQL())

	newDateTo := dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	rows, err := t.pool.Query(ctx, query, chainName, dateFrom, newDateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dailyActiveAccounts []*database.DailyActiveAccount
	for rows.Next() {
		dailyActiveAccount := &database.DailyActiveAccount{}
		err := rows.Scan(&dailyActiveAccount.Date, &dailyActiveAccount.Count)
		if err != nil {
			return nil, err
		}
		dailyActiveAccounts = append(dailyActiveAccounts, dailyActiveAccount)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dailyActiveAccounts, nil
}

func (t *TimescaleDb) GetTotalTxCount(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	coalesce(SUM(transaction_count), 0) as total_tx_count
	FROM tx_counter
	WHERE
	chain_name = $1
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var totalTxCount int64
	err := row.Scan(&totalTxCount)
	if err != nil {
		return 0, err
	}
	return totalTxCount, nil
}

func (t *TimescaleDb) GetTotalTxCount24h(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	coalesce(SUM(transaction_count), 0) as tx_count_24h
	FROM tx_counter
	WHERE
	chain_name = $1
	AND time_bucket >= NOW() - INTERVAL '24 hours'
	AND time_bucket < NOW()
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var txCount24h int64
	err := row.Scan(&txCount24h)
	if err != nil {
		return 0, err
	}
	return txCount24h, nil
}

func (t *TimescaleDb) GetTotalTxCountByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) ([]*database.TxCountDateRange, error) {

	query := fmt.Sprintf(`
	SELECT date::date, tx_count FROM (
	SELECT
		time_bucket_gapfill('1 day', time_bucket) as date,
		coalesce(SUM(transaction_count), 0) as tx_count
		FROM tx_counter
		WHERE
		chain_name = $1
		AND time_bucket >= $2 AND time_bucket <= $3
		GROUP BY 1
	) sub
	ORDER BY date %s
	`, sortOrder.SQL())

	newDateTo := dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	rows, err := t.pool.Query(ctx, query, chainName, dateFrom, newDateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txCountTimeRanges []*database.TxCountDateRange
	for rows.Next() {
		txCountTimeRange := &database.TxCountDateRange{}
		err := rows.Scan(&txCountTimeRange.Date, &txCountTimeRange.Count)
		if err != nil {
			return nil, err
		}
		txCountTimeRanges = append(txCountTimeRanges, txCountTimeRange)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txCountTimeRanges, nil
}

func (t *TimescaleDb) GetTotalTxCountByHour(
	ctx context.Context,
	chainName string,
	fromTimestamp time.Time,
	endTimestamp time.Time,
	sortOrder database.SortOrder,
) ([]*database.TxCountTimeRange, error) {
	query := fmt.Sprintf(`
	SELECT
	time_bucket_gapfill('1 hour', time_bucket) as timestamp,
	coalesce(SUM(transaction_count), 0) as tx_count
	FROM tx_counter
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY 1
	ORDER BY timestamp %s
	`, sortOrder.SQL())
	rows, err := t.pool.Query(ctx, query, chainName, fromTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txCountTimeRanges []*database.TxCountTimeRange
	for rows.Next() {
		txCountTimeRange := &database.TxCountTimeRange{}
		err := rows.Scan(&txCountTimeRange.Time, &txCountTimeRange.Count)
		if err != nil {
			return nil, err
		}
		txCountTimeRanges = append(txCountTimeRanges, txCountTimeRange)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txCountTimeRanges, nil
}

func (t *TimescaleDb) GetVolumeByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder database.SortOrder,
) (database.VolumeByDenomDaily, error) {

	query := fmt.Sprintf(`
	SELECT date::date, volume, denom
	FROM (
		SELECT
			time_bucket_gapfill('1 day', time_bucket) AS date,
			coalesce(SUM(volume), 0) AS volume,
			denom
		FROM fee_volume
		WHERE
			chain_name = $1
			AND time_bucket >= $2 AND time_bucket <= $3
			GROUP BY 1, denom
		) sub
	ORDER BY date %s
	`, sortOrder.SQL())

	newDateTo := dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	rows, err := t.pool.Query(
		ctx, query, chainName, dateFrom.Format("2006-01-02"), newDateTo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeVolumeTimeRanges = make(database.VolumeByDenomDaily)
	for rows.Next() {
		denomVolume := &database.DenomVolumeDaily{}
		denom := ""
		err := rows.Scan(&denomVolume.Date, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], *denomVolume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return feeVolumeTimeRanges, nil
}

func (t *TimescaleDb) GetVolumeByHour(
	ctx context.Context,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	sortOrder database.SortOrder,
) (database.VolumeByDenomHourly, error) {

	query := fmt.Sprintf(`
	WITH hours AS (
		SELECT generate_series(date_trunc('hour', $2::timestamptz), date_trunc('hour', $3::timestamptz), '1 hour'::interval) AS hour
	),
	denoms AS (
		SELECT DISTINCT denom FROM fee_volume WHERE chain_name = $1
	)
	SELECT
		h.hour AS time,
		COALESCE(SUM(fv.volume), 0) AS volume,
		d.denom
	FROM hours h
	CROSS JOIN denoms d
	LEFT JOIN fee_volume fv ON
		fv.time_bucket = h.hour AND
		fv.chain_name = $1 AND
		fv.denom = d.denom
	GROUP BY h.hour, d.denom
	ORDER BY h.hour %s
	`, sortOrder.SQL())
	rows, err := t.pool.Query(ctx, query, chainName, fromTimestamp, toTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeVolumeTimeRanges = make(database.VolumeByDenomHourly)
	for rows.Next() {
		denomVolume := &database.DenomVolumeHourly{}
		denom := ""
		err := rows.Scan(&denomVolume.Time, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], *denomVolume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return feeVolumeTimeRanges, nil
}

func (t *TimescaleDb) GetValidatorSigning24h(
	ctx context.Context,
	validatorAddress string,
	chainName string,
) (*database.ValidatorSigning, error) {
	query1 := `
	SELECT
	id
	FROM gno_validators
	WHERE address = $1
	AND chain_name = $2
	`
	row1 := t.pool.QueryRow(ctx, query1, validatorAddress, chainName)
	var validatorId int32
	err := row1.Scan(&validatorId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("validator %q: %w", validatorAddress, database.ErrNotFound)
		}
		return nil, fmt.Errorf("validator lookup failed: %w", err)
	}

	query2 := `
	WITH total_blocks AS (
        SELECT
            coalesce(sum(bc.block_count), 0) AS count
        FROM block_counter bc
        WHERE bc.time_bucket >= now() - INTERVAL '24 hours'
        AND bc.time_bucket <  now()
        AND bc.chain_name = $1
    )

	SELECT
    coalesce(sum(vsc.blocks_signed), 0) AS blocks_signed,
    tb.count - coalesce(sum(vsc.blocks_signed), 0) AS blocks_not_signed,
    tb.count AS block_count,
    coalesce(
        round(
            coalesce(
                sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0
            ) AS signing_rate_pct
	FROM validator_signing_counter vsc
	CROSS JOIN total_blocks tb
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.validator_id  = $2
		AND vsc.time_bucket >= now() - INTERVAL '24 hours'
		AND vsc.time_bucket <  now()
	GROUP BY tb.count;
	`

	row := t.pool.QueryRow(ctx, query2, chainName, validatorId)
	var validatorSigning database.ValidatorSigning
	err = row.Scan(
		&validatorSigning.BlocksSigned,
		&validatorSigning.BlocksMissed,
		&validatorSigning.TotalBlocks,
		&validatorSigning.SigningRate,
	)
	if err != nil {
		return nil, err
	}
	return &validatorSigning, nil
}

func (t *TimescaleDb) GetValidatorSigningByHour(
	ctx context.Context,
	validatorAddress string,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	sortOrder database.SortOrder,
) ([]*database.ValidatorSigning, error) {
	query1 := `
	SELECT
	id
	FROM gno_validators
	WHERE address = $1
	AND chain_name = $2
	`
	row1 := t.pool.QueryRow(ctx, query1, validatorAddress, chainName)
	var validatorId int32
	err := row1.Scan(&validatorId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("validator %q: %w", validatorAddress, database.ErrNotFound)
		}
		return nil, fmt.Errorf("validator lookup failed: %w", err)
	}

	query2 := fmt.Sprintf(`
	SELECT
	time_bucket_gapfill('1 hour', vsc.time_bucket) as time,
	coalesce(sum(vsc.blocks_signed), 0) as blocks_signed,
	coalesce(sum(bc.block_count), 0) - coalesce(sum(vsc.blocks_signed), 0) as blocks_not_signed,
	coalesce(sum(bc.block_count), 0) as total_blocks,
	coalesce(round(coalesce(sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0) as signing_rate_pct
	FROM validator_signing_counter vsc
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.validator_id  = $2
		AND vsc.time_bucket >= $3 AND vsc.time_bucket <= $4
	GROUP BY 1
	ORDER BY time %s
	`, sortOrder.SQL())
	rows, err := t.pool.Query(ctx, query2, chainName, validatorId, fromTimestamp, toTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var validatorSignings []*database.ValidatorSigning
	for rows.Next() {
		validatorSigning := &database.ValidatorSigning{}
		err := rows.Scan(&validatorSigning.Time, &validatorSigning.BlocksSigned, &validatorSigning.BlocksMissed, &validatorSigning.TotalBlocks, &validatorSigning.SigningRate)
		if err != nil {
			return nil, err
		}
		validatorSignings = append(validatorSignings, validatorSigning)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return validatorSignings, nil
}

// GetAllValidatorSigning24h returns signing stats for all validators over the last 24 hours.
func (t *TimescaleDb) GetAllValidatorSigning24h(
	ctx context.Context,
	chainName string,
) (database.AllValidatorSignings, error) {
	query := `
    WITH total_blocks AS (
           SELECT
               coalesce(sum(bc.block_count), 0) AS count
           FROM block_counter bc
           WHERE bc.time_bucket >= now() - INTERVAL '24 hours'
           AND bc.time_bucket <  now()
           AND bc.chain_name = $1
       )

	SELECT
	   gv.address as validator_address,
       coalesce(sum(vsc.blocks_signed), 0) AS blocks_signed,
       tb.count - coalesce(sum(vsc.blocks_signed), 0) AS blocks_not_signed,
       tb.count AS block_count,
       coalesce(
           round(
               coalesce(
                   sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0
               ) AS signing_rate
	FROM validator_signing_counter vsc
	CROSS JOIN total_blocks tb
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	JOIN gno_validators gv
		ON gv.id = vsc.validator_id
		AND gv.chain_name = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.time_bucket >= now() - INTERVAL '24 hours'
		AND vsc.time_bucket <  now()
	GROUP BY tb.count, gv.address;
	`
	rows, err := t.pool.Query(ctx, query, chainName)
	if err != nil {
		return nil, err
	}
	var result = make(database.AllValidatorSignings)
	defer rows.Close()
	for rows.Next() {
		var address string
		var entry database.ValidatorSigning
		if err := rows.Scan(
			&address,
			&entry.BlocksSigned,
			&entry.BlocksMissed,
			&entry.TotalBlocks,
			&entry.SigningRate,
		); err != nil {
			return nil, err
		}
		result[address] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
