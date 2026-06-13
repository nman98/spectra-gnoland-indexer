package timescaledb

import (
	"context"
	"fmt"
)

type KeyParams struct {
	RpmLimit int
	Name     string
	Prefix   string
	Hash     [32]byte
}

type ApiKeyListItem struct {
	Prefix   string
	Name     string
	RpmLimit int
	IsActive bool
}

func (t *TimescaleDb) InsertApiKey(
	ctx context.Context, params KeyParams) error {
	result, err := t.pool.Exec(ctx, `
		INSERT INTO api_keys (prefix, hash, name, rpm_limit)
		VALUES ($1, $2, $3, $4)
		`, params.Prefix, params.Hash[:], params.Name, params.RpmLimit)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) GetAllApiKeys(ctx context.Context) ([][32]byte, error) {
	query := `
		SELECT hash
		FROM api_keys
		WHERE is_active = true
		`
	rows, err := t.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apiKeys := make([][32]byte, 0)
	for rows.Next() {
		var hashSlice []byte
		err := rows.Scan(&hashSlice)
		if err != nil {
			return nil, err
		}
		var hash [32]byte
		copy(hash[:], hashSlice)
		apiKeys = append(apiKeys, hash)
	}
	return apiKeys, nil
}

func (t *TimescaleDb) GetAllApiKeysWithLimits(ctx context.Context) (map[[32]byte]int, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT hash, rpm_limit
		FROM api_keys
		WHERE is_active = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[[32]byte]int)
	for rows.Next() {
		var hashSlice []byte
		var rpmLimit int
		if err := rows.Scan(&hashSlice, &rpmLimit); err != nil {
			return nil, err
		}
		var hash [32]byte
		copy(hash[:], hashSlice)
		keys[hash] = rpmLimit
	}
	return keys, rows.Err()
}

func (t *TimescaleDb) ListApiKeys(ctx context.Context) ([]ApiKeyListItem, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT prefix, name, rpm_limit, is_active
		FROM api_keys
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ApiKeyListItem
	for rows.Next() {
		var item ApiKeyListItem
		if err := rows.Scan(&item.Prefix, &item.Name, &item.RpmLimit, &item.IsActive); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (t *TimescaleDb) DisableKeyByName(ctx context.Context, name string) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = false
		WHERE name = $1
		`, name)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no key found with name %q", name)
	}
	return nil
}

func (t *TimescaleDb) EnableKeyByName(ctx context.Context, name string) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = true
		WHERE name = $1
		`, name)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no key found with name %q", name)
	}
	return nil
}

func (t *TimescaleDb) DisableKey(ctx context.Context, hash [32]byte) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = false
		WHERE hash = $1
		`, hash[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) EnableKey(ctx context.Context, hash [32]byte) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = true
		WHERE hash = $1
		`, hash[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) AdjustRpmLimit(
	ctx context.Context, hash [32]byte, rpmLimit int,
) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET rpm_limit = $1
		WHERE hash = $2
		`, rpmLimit, hash[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
