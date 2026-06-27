package timescaledb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TimescaleDb is the database connection pool
type TimescaleDb struct {
	pool *pgxpool.Pool
}

// DatabasePoolConfig is the configuration for the database pool.
type DatabasePoolConfig struct {
	// Basic connection info
	User     string
	Password string
	Host     string
	Port     int
	Dbname   string
	Sslmode  string
	// SSL cert paths — only needed for verify-ca/verify-full or mutual TLS.
	// SslRootCert: CA cert to verify the server. SslCert + SslKey: client
	// certificate keypair, only when pg_hba.conf uses cert auth.
	SslRootCert string
	SslCert     string
	SslKey      string

	// Pool config
	PoolMaxConns              int
	PoolMinConns              int
	PoolMaxConnLifetime       time.Duration
	PoolMaxConnIdleTime       time.Duration
	PoolHealthCheckPeriod     time.Duration
	PoolMaxConnLifetimeJitter time.Duration
}

// NewTimescaleDb is a constructor function that creates a new TimescaleDb instance
func NewTimescaleDb(config DatabasePoolConfig) *TimescaleDb {
	pool, err := connectToDb(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

// NewTimescaleDbSetup creates a new TimescaleDb instance for the setup process.
// Should be used to create the database and switch to it.
func NewTimescaleDbSetup(config DatabasePoolConfig) *TimescaleDb {
	pool, err := setupConnection(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

// buildDSN constructs a libpq keyword=value connection string from config.
// SSL cert fields are appended only when non-empty so that modes like
// "require" or "disable" work without any cert files on disk.
func buildDSN(config DatabasePoolConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"+
			" pool_max_conns=%d pool_min_conns=%d pool_max_conn_lifetime=%s"+
			" pool_max_conn_idle_time=%s pool_health_check_period=%s"+
			" pool_max_conn_lifetime_jitter=%s",
		config.Host, config.Port, config.User, config.Password,
		config.Dbname, config.Sslmode,
		config.PoolMaxConns, config.PoolMinConns, config.PoolMaxConnLifetime,
		config.PoolMaxConnIdleTime, config.PoolHealthCheckPeriod,
		config.PoolMaxConnLifetimeJitter,
	)
	if config.SslRootCert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", config.SslRootCert)
	}
	if config.SslCert != "" {
		dsn += fmt.Sprintf(" sslcert=%s", config.SslCert)
	}
	if config.SslKey != "" {
		dsn += fmt.Sprintf(" sslkey=%s", config.SslKey)
	}
	return dsn
}

func connectToDb(config DatabasePoolConfig) (*pgxpool.Pool, error) {
	parseConfig, err := pgxpool.ParseConfig(buildDSN(config))
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

// setupConnection connects to the database for the setup process only.
func setupConnection(config DatabasePoolConfig) (*pgxpool.Pool, error) {
	parseConfig, err := pgxpool.ParseConfig(buildDSN(config))
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// CreateDatabase creates a new database with the given name.
func CreateDatabase(db *TimescaleDb, dbname string) error {
	_, err := db.pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("database %s already exists", dbname)) {
			return nil
		}
		return err
	}
	return nil
}

// SwitchDatabase switches the connection to the named database.
// Only for use during the setup process.
//
// TODO: the devs could integrate the indexer within already existing timescale db
// so remove hard coded dbname gnoland anywhere else in the project.
func SwitchDatabase(db *TimescaleDb, config DatabasePoolConfig, dbname string) error {
	db.pool.Close()

	newConfig := config
	newConfig.Dbname = dbname

	newPool, err := setupConnection(newConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", dbname, err)
	}

	db.pool = newPool
	return nil
}

// GetPool returns the underlying connection pool.
func (db *TimescaleDb) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *TimescaleDb) Close() {
	db.pool.Close()
}
