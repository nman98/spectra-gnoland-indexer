package database

import (
	"context"
	"fmt"
)

func (t *TimescaleDb) GetAllValidators(ctx context.Context, chainName string) (*ValidatorList, error) {
	query := `
	SELECT
	address
	FROM gno_validators
	ORDER BY id ASC
	`
	var validators = make([]string, 0)
	rows, err := t.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var valAddr string
		err := rows.Scan(&valAddr)
		if err != nil {
			return nil, err
		}
		validators = append(validators, valAddr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(validators) == 0 {
		return nil, fmt.Errorf("validator list: %w", ErrNotFound)
	}
	return &ValidatorList{
		ValAddresses: validators,
	}, nil
}
