package timescaledb

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestBuildDSNEscapesSpecialChars verifies that values containing spaces,
// single quotes, and backslashes survive libpq DSN parsing intact.
func TestBuildDSNEscapesSpecialChars(t *testing.T) {
	config := DatabasePoolConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "indexer",
		Password: `p@ss w'ord\x`,
		Dbname:       "gnoland",
		Sslmode:      "disable",
		PoolMaxConns: 10,
	}

	parsed, err := pgxpool.ParseConfig(buildDSN(config))
	if err != nil {
		t.Fatalf("ParseConfig failed for escaped DSN: %v", err)
	}

	if got := parsed.ConnConfig.Password; got != config.Password {
		t.Errorf("password not preserved: got %q, want %q", got, config.Password)
	}
	if got := parsed.ConnConfig.User; got != config.User {
		t.Errorf("user not preserved: got %q, want %q", got, config.User)
	}
	if got := parsed.ConnConfig.Database; got != config.Dbname {
		t.Errorf("dbname not preserved: got %q, want %q", got, config.Dbname)
	}
}
