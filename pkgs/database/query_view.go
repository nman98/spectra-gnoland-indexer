package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
)

func (t *TimescaleDb) GetBlockCountByDate(
	ctx context.Context,
	chainName string,
	dateFrom date.Date,
	dateTo date.Date,
	sortOrder SortOrder,
) ([]*BlockCountByDate, error) {

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
	var blockCountByDates []*BlockCountByDate
	for rows.Next() {
		blockCountByDate := &BlockCountByDate{}
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
	SUM(block_count) as block_count
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
	sortOrder SortOrder,
) ([]*DailyActiveAccount, error) {
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
	var dailyActiveAccounts []*DailyActiveAccount
	for rows.Next() {
		dailyActiveAccount := &DailyActiveAccount{}
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
	SUM(transaction_count) as total_tx_count
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
	SUM(transaction_count) as tx_count_24h
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
	sortOrder SortOrder,
) ([]*TxCountDateRange, error) {

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
	var txCountTimeRanges []*TxCountDateRange
	for rows.Next() {
		txCountTimeRange := &TxCountDateRange{}
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
	sortOrder SortOrder,
) ([]*TxCountTimeRange, error) {
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
	var txCountTimeRanges []*TxCountTimeRange
	for rows.Next() {
		txCountTimeRange := &TxCountTimeRange{}
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
	sortOrder SortOrder,
) (VolumeByDenomDaily, error) {

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
	var feeVolumeTimeRanges = make(VolumeByDenomDaily)
	for rows.Next() {
		denomVolume := &DenomVolumeDaily{}
		denom := ""
		err := rows.Scan(&denomVolume.Date, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], denomVolume)
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
	sortOrder SortOrder,
) (VolumeByDenomHourly, error) {

	query := fmt.Sprintf(`
	SELECT
	time_bucket_gapfill('1 hour', time_bucket) as time,
	coalesce(SUM(volume), 0) as volume,
	denom
	FROM fee_volume
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY 1, denom
	ORDER BY time %s
	`, sortOrder.SQL())
	rows, err := t.pool.Query(ctx, query, chainName, fromTimestamp, toTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeVolumeTimeRanges = make(VolumeByDenomHourly)
	for rows.Next() {
		denomVolume := &DenomVolumeHourly{}
		denom := ""
		err := rows.Scan(&denomVolume.Time, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], denomVolume)
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
) (*ValidatorSigning, error) {
	// Check if validator address exists and get its id
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
		return nil, fmt.Errorf("validator seems to not exist: %w", err)
	}

	query2 := `
	SELECT
    coalesce(sum(vsc.blocks_signed), 0) AS blocks_signed,
    coalesce(sum(bc.block_count), 0) - coalesce(sum(vsc.blocks_signed), 0) AS blocks_not_signed,
    coalesce(sum(bc.block_count), 0) AS total_blocks,
    coalesce(round(coalesce(sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0) AS signing_rate_pct
	FROM validator_signing_counter vsc
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.validator_id  = $2
		AND vsc.time_bucket >= now() - INTERVAL '24 hours'
		AND vsc.time_bucket <  now();
	`

	row := t.pool.QueryRow(ctx, query2, chainName, validatorId)
	var validatorSigning ValidatorSigning
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
	sortOrder SortOrder,
) ([]*ValidatorSigning, error) {
	// Check if validator address exists and get its id
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
		return nil, fmt.Errorf("validator seems to not exist: %w", err)
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
	var validatorSignings []*ValidatorSigning
	for rows.Next() {
		validatorSigning := &ValidatorSigning{}
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
