package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewTimescaleDb is a constructor function that creates a new TimescaleDb instance
//
// Parameters:
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - *TimescaleDb: the TimescaleDb instance
//
// The method will not throw an error if the TimescaleDb is not found, it will just return nil
func NewTimescaleDb(config DatabasePoolConfig) *TimescaleDb {
	pool, err := connectToDb(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

// NewTimescaleDbSetup is a constructor function that creates a new TimescaleDb instance
// Should be used to create the database and switch to it
func NewTimescaleDbSetup(config DatabasePoolConfig) *TimescaleDb {
	pool, err := setupConnection(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

// connectToDb is a internal function that connects to the database for the indexer
//
// Parameters:
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - *pgxpool.Pool: the database connection pool
//
// The method will not throw an error if the database connection pool is not found, it will just return nil
func connectToDb(config DatabasePoolConfig) (*pgxpool.Pool, error) {
	parseConfig, err := pgxpool.ParseConfig(
		fmt.Sprintf(
			`host=%s port=%d user=%s password=%s 
			dbname=%s sslmode=%s pool_max_conns=%d 
			pool_min_conns=%d pool_max_conn_lifetime=%s 
			pool_max_conn_idle_time=%s pool_health_check_period=%s 
			pool_max_conn_lifetime_jitter=%s`,
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Dbname,
			config.Sslmode,
			config.PoolMaxConns,
			config.PoolMinConns,
			config.PoolMaxConnLifetime,
			config.PoolMaxConnIdleTime,
			config.PoolHealthCheckPeriod,
			config.PoolMaxConnLifetimeJitter))
	if err != nil {
		return nil, err
	}

	// Register custom types and enforce UTC for every connection in the pool
	parseConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "SET timezone TO 'UTC'"); err != nil {
			return fmt.Errorf("failed to set session timezone to UTC: %w", err)
		}

		dataTypeNames := sql_data_types.CustomTypeNames()

		for _, typeName := range dataTypeNames {
			dataType, err := conn.LoadType(ctx, typeName)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(dataType)
		}

		return nil
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// setupConnection is a internal function that connects to the database
// should only be used as a part of the set up process
// to be used to create the database and switch to it
func setupConnection(config DatabasePoolConfig) (*pgxpool.Pool, error) {
	parseConfig, err := pgxpool.ParseConfig(
		fmt.Sprintf(
			`host=%s port=%d user=%s password=%s 
			dbname=%s sslmode=%s pool_max_conns=%d 
			pool_min_conns=%d pool_max_conn_lifetime=%s 
			pool_max_conn_idle_time=%s pool_health_check_period=%s 
			pool_max_conn_lifetime_jitter=%s`,
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Dbname,
			config.Sslmode,
			config.PoolMaxConns,
			config.PoolMinConns,
			config.PoolMaxConnLifetime,
			config.PoolMaxConnIdleTime,
			config.PoolHealthCheckPeriod,
			config.PoolMaxConnLifetimeJitter))
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// create a new database with the given name
//
// Parameters:
//   - db: the database connection pool
//   - dbname: the name of the database to create
//
// Returns:
//   - error: an error if the creation fails
func CreateDatabase(db *TimescaleDb, dbname string) error {
	_, err := db.pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil && !strings.Contains(err.Error(), fmt.Sprintf("database %s already exists", dbname)) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// Switch to the database with the given name
// this is used to switch to the database after creating it
// most of the time when the postgres server is running, it will be in the "postgres" database
// only to be used initiating command
//
// Parameters:
//   - db: the database connection pool
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - error: an error if the switching fails
//
// TODO: the devs could integrate the indexer within already existing timescale db
// so remove hard coded dbname gnoland anywhere else in the project
func SwitchDatabase(db *TimescaleDb, config DatabasePoolConfig, dbname string) error {
	// Close the current connection
	db.pool.Close()

	// Create a new config with the target database name
	newConfig := config
	newConfig.Dbname = dbname

	// Create a new connection to the target database
	newPool, err := setupConnection(newConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", dbname, err)
	}

	// Replace the old pool with the new one
	db.pool = newPool
	return nil
}

// GetPool returns the database connection pool
//
// Parameters:
//   - db: the database connection pool
//
// Returns:
//   - *pgxpool.Pool: the database connection pool
func (db *TimescaleDb) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *TimescaleDb) Close() {
	db.pool.Close()
}
