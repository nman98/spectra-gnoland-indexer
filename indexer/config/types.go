package config

import (
	"time"
)

type Environment struct {
	Host    string `env:"DB_HOST" envDefault:"localhost"`
	Port    int    `env:"DB_PORT" envDefault:"5432"`
	User    string `env:"DB_USER" envDefault:"postgres"`
	Sslmode     string `env:"DB_SSLMODE" envDefault:"disable"`
	SslRootCert string `env:"DB_SSLROOTCERT" envDefault:""`
	SslCert     string `env:"DB_SSLCERT" envDefault:""`
	SslKey      string `env:"DB_SSLKEY" envDefault:""`
	// do not use password default unless for development or testing!!!
	Password string `env:"DB_PASSWORD" envDefault:"12345678"`
	Dbname   string `env:"DB_NAME" envDefault:"gnoland"`
}

type Config struct {
	RpcUrl                    string        `yaml:"rpc"`
	UserAgent                 *string       `yaml:"user_agent"`
	PoolMaxConns              int           `yaml:"pool_max_conns"`
	PoolMinConns              int           `yaml:"pool_min_conns"`
	PoolMaxConnLifetime       time.Duration `yaml:"pool_max_conn_lifetime"`
	PoolMaxConnIdleTime       time.Duration `yaml:"pool_max_conn_idle_time"`
	PoolHealthCheckPeriod     time.Duration `yaml:"pool_health_check_period"`
	PoolMaxConnLifetimeJitter time.Duration `yaml:"pool_max_conn_lifetime_jitter"`
	LivePooling               time.Duration `yaml:"live_pooling"`
	MaxBlockChunkSize         uint64        `yaml:"max_block_chunk_size"`
	MaxTransactionChunkSize   uint64        `yaml:"max_transaction_chunk_size"`
	ChainName                 string        `yaml:"chain_name"`
	// retry options are optional, so we need to set them to a pointer
	RetryAmount        *int           `yaml:"retry_amount"`
	Pause              *int           `yaml:"pause"`
	PauseTime          *time.Duration `yaml:"pause_time"`
	ExponentialBackoff *time.Duration `yaml:"exponential_backoff"`
}
