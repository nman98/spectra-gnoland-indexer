//go:build integration
// +build integration

package integration_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/integration/synthetic"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"go.yaml.in/yaml/v4"
)

type TestConfig struct {
	Host                      string        `yaml:"host"`
	Port                      int           `yaml:"port"`
	User                      string        `yaml:"user"`
	Password                  string        `yaml:"password"`
	Dbname                    string        `yaml:"dbname"`
	Sslmode                   string        `yaml:"sslmode"`
	PoolMaxConns              int           `yaml:"pool_max_conns"`
	PoolMinConns              int           `yaml:"pool_min_conns"`
	PoolMaxConnLifetime       time.Duration `yaml:"pool_max_conn_lifetime"`
	PoolMaxConnIdleTime       time.Duration `yaml:"pool_max_conn_idle_time"`
	PoolHealthCheckPeriod     time.Duration `yaml:"pool_health_check_period"`
	PoolMaxConnLifetimeJitter time.Duration `yaml:"pool_max_conn_lifetime_jitter"`
	ChainID                   string        `yaml:"chain_id"`
	FromHeight                uint64        `yaml:"from_height"`
	ToHeight                  uint64        `yaml:"to_height"`
}

func TestHistoricSyntheticIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Starting synthetic integration test...")
	startTime := time.Now()

	config := loadTestConfig(t)
	t.Logf("Test configuration loaded")
	t.Logf("Chain ID: %s", config.ChainID)
	t.Logf("Processing blocks %d to %d", config.FromHeight, config.ToHeight)

	err := synthetic.RunSyntheticIntegrationTest(&config)
	if err != nil {
		t.Fatalf("Synthetic integration test failed: %v", err)
	}

	duration := time.Since(startTime)
	t.Logf("Test completed successfully in %v", duration)

	// Verify database state
	verifyDatabaseState(t, config)
}

func loadTestConfig(t *testing.T) synthetic.SyntheticIntegrationTestConfig {
	t.Helper()

	// Try to load from YAML config file first
	yamlFile, err := os.ReadFile("test_config.yml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	var config TestConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	return synthetic.SyntheticIntegrationTestConfig{
		DatabaseConfig: database.DatabasePoolConfig{
			Host:                      config.Host,
			Port:                      config.Port,
			User:                      config.User,
			Password:                  config.Password,
			Dbname:                    config.Dbname,
			Sslmode:                   config.Sslmode,
			PoolMaxConns:              config.PoolMaxConns,
			PoolMinConns:              config.PoolMinConns,
			PoolMaxConnLifetime:       config.PoolMaxConnLifetime,
			PoolMaxConnIdleTime:       config.PoolMaxConnIdleTime,
			PoolHealthCheckPeriod:     config.PoolHealthCheckPeriod,
			PoolMaxConnLifetimeJitter: config.PoolMaxConnLifetimeJitter,
		},
		ChainID:    config.ChainID,
		FromHeight: config.FromHeight,
		ToHeight:   config.ToHeight,
	}
}

func verifyDatabaseState(t *testing.T, config synthetic.SyntheticIntegrationTestConfig) {
	t.Helper()

	// Initialize database connection
	db := database.NewTimescaleDb(config.DatabaseConfig)
	pool := db.GetPool()

	ctx := context.Background()

	// Verify blocks were inserted
	var blockCount int64
	query := `SELECT COUNT(*) FROM blocks WHERE chain_id = $1 AND height >= $2 AND height <= $3`
	err := pool.QueryRow(ctx, query, config.ChainID, config.FromHeight, config.ToHeight).Scan(&blockCount)
	if err != nil {
		t.Fatalf("Failed to query block count: %v", err)
	}

	expectedBlocks := int64(config.ToHeight - config.FromHeight + 1)
	if blockCount != expectedBlocks {
		t.Errorf("Expected %d blocks, got %d", expectedBlocks, blockCount)
	} else {
		t.Logf("Verified %d blocks in database", blockCount)
	}

	// Verify transactions were inserted
	var txCount int64
	txQuery := `SELECT COUNT(*) FROM transaction_general WHERE chain_name = $1 AND block_height >= $2 AND block_height <= $3`
	err = pool.QueryRow(ctx, txQuery, config.ChainID, config.FromHeight, config.ToHeight).Scan(&txCount)
	if err != nil {
		t.Fatalf("Failed to query transaction count: %v", err)
	}

	t.Logf("Verified %d transactions in database", txCount)

	// Verify precommits were inserted
	var precommitCount int64
	precommitQuery := `SELECT COUNT(*) FROM validator_block_signing WHERE chain_name = $1 AND block_height >= $2 AND block_height <= $3`
	err = pool.QueryRow(ctx, precommitQuery, config.ChainID, config.FromHeight, config.ToHeight).Scan(&precommitCount)
	if err != nil {
		t.Fatalf("Failed to query precommit count: %v", err)
	}

	t.Logf("Verified %d precommits in database", precommitCount)

	// Log summary
	log.Printf("Database verification complete:")
	log.Printf("\tBlocks: %d", blockCount)
	log.Printf("\tTransactions: %d", txCount)
	log.Printf("\tPrecommits: %d", precommitCount)
}
