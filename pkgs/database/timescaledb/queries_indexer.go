package timescaledb

import (
	"context"
)

// FindExistingAccounts finds which addresses already exist in the database.
// Used by the account cache to determine which ones to insert.
func (t *TimescaleDb) FindExistingAccounts(
	ctx context.Context,
	addresses []string,
	chainName string,
	searchValidators bool,
) (map[string]int32, error) {
	addressesMap := make(map[string]int32)
	query := ""
	if searchValidators {
		query = `
	SELECT address, id
	FROM gno_validators
	WHERE chain_name = $1
	AND address = ANY($2)
	`
	} else {
		query = `
		SELECT address, id
		FROM gno_addresses
		WHERE chain_name = $1
		AND address = ANY($2)
		`
	}
	rows, err := t.pool.Query(ctx, query, chainName, addresses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		var id int32
		err := rows.Scan(&address, &id)
		if err != nil {
			return nil, err
		}
		addressesMap[address] = id
	}
	return addressesMap, nil
}

// GetAllAddresses returns all addresses for a chain, starting after highestIndex.
// Used at program init to load existing accounts and their IDs.
func (t *TimescaleDb) GetAllAddresses(
	ctx context.Context,
	chainName string,
	searchValidators bool,
	highestIndex *int32,
) (map[string]int32, int32, error) {
	addressesMap := make(map[string]int32)
	var maxIndex int32 = 0
	if highestIndex != nil {
		maxIndex = *highestIndex
	}
	query := ""
	if searchValidators {
		query += `
		SELECT address, id
		FROM gno_validators
		WHERE chain_name = $1
		AND id > $2
		`
	} else {
		query += `
		SELECT address, id
		FROM gno_addresses
		WHERE chain_name = $1
		AND id > $2
		`
	}
	rows, err := t.pool.Query(ctx, query, chainName, maxIndex)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		var id int32
		err := rows.Scan(&address, &id)
		if err != nil {
			return nil, 0, err
		}
		addressesMap[address] = id
		if id > maxIndex {
			maxIndex = id
		}
	}
	return addressesMap, maxIndex, nil
}

// CheckCurrentDatabaseName returns the name of the database the pool is connected to.
func (t *TimescaleDb) CheckCurrentDatabaseName(ctx context.Context) (string, error) {
	query := `
	SELECT current_database()
	`
	row := t.pool.QueryRow(ctx, query)
	var currentDb string
	err := row.Scan(&currentDb)
	if err != nil {
		return "", err
	}
	return currentDb, nil
}

// GetLastBlockHeight returns the highest indexed block height for the given chain.
func (t *TimescaleDb) GetLastBlockHeight(ctx context.Context, chainName string) (uint64, error) {
	query := `
	SELECT
	coalesce(
	    (SELECT
			height
		FROM blocks
		WHERE chain_name = $1
		ORDER BY height DESC
		LIMIT 1
		), 0) AS height
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var lastBlockHeight uint64
	err := row.Scan(&lastBlockHeight)
	if err != nil {
		return 0, err
	}
	return lastBlockHeight, nil
}
