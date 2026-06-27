package config

import "time"

type ApiConfig struct {
	// Basic connection info
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	// CORS config
	CorsAllowedOrigins []string `yaml:"cors_allowed_origins"`
	CorsAllowedMethods []string `yaml:"cors_allowed_methods"`
	CorsAllowedHeaders []string `yaml:"cors_allowed_headers"`
	CorsMaxAge         int      `yaml:"cors_max_age"`
	ChainName          string   `yaml:"chain_name"`
	// Rate limiting
	// Set DisableRateLimit to true when an upstream API gateway (Kong, AWS API
	// Gateway, Cloudflare, etc.) already handles rate limiting and you do not
	// want to run a Valkey instance. When true, IpRpmLimit, KeyRefreshInterval,
	// and TrustedProxies are ignored and the Valkey client is never initialised.
	DisableRateLimit   bool          `yaml:"disable_rate_limit"`
	IpRpmLimit         int           `yaml:"ip_rpm_limit"`
	KeyRefreshInterval time.Duration `yaml:"key_refresh_interval"`
	// TrustedProxies is the list of CIDR blocks whose X-Forwarded-For / X-Real-Ip
	// headers are trusted for real-IP resolution. Leave empty to always use
	// RemoteAddr (safe default for direct-exposure deployments).
	TrustedProxies []string `yaml:"trusted_proxies"`
}

type ApiEnv struct {
	ApiDbHost                      string        `env:"API_DB_HOST" envDefault:"localhost"`
	ApiDbPort                      int           `env:"API_DB_PORT" envDefault:"5432"`
	ApiDbUser                      string        `env:"API_DB_USER" envDefault:"postgres"`
	ApiDbPassword                  string        `env:"API_DB_PASSWORD" envDefault:"12345678"`
	ApiDbName                      string        `env:"API_DB_NAME" envDefault:"gnoland"`
	ApiDbSslmode                   string        `env:"API_DB_SSLMODE" envDefault:"disable"`
	ApiDbSslRootCert               string        `env:"API_DB_SSLROOTCERT" envDefault:""`
	ApiDbSslCert                   string        `env:"API_DB_SSLCERT" envDefault:""`
	ApiDbSslKey                    string        `env:"API_DB_SSLKEY" envDefault:""`
	ApiDbPoolMaxConns              int           `env:"API_DB_POOL_MAX_CONNS" envDefault:"50"`
	ApiDbPoolMinConns              int           `env:"API_DB_POOL_MIN_CONNS" envDefault:"10"`
	ApiDbPoolMaxConnLifetime       time.Duration `env:"API_DB_POOL_MAX_CONN_LIFETIME" envDefault:"10m"`
	ApiDbPoolMaxConnIdleTime       time.Duration `env:"API_DB_POOL_MAX_CONN_IDLE_TIME" envDefault:"5m"`
	ApiDbPoolHealthCheckPeriod     time.Duration `env:"API_DB_POOL_HEALTH_CHECK_PERIOD" envDefault:"1m"`
	ApiDbPoolMaxConnLifetimeJitter time.Duration `env:"API_DB_POOL_MAX_CONN_LIFETIME_JITTER" envDefault:"1m"`
}

type ValkeyEnv struct {
	Host string `env:"VALKEY_HOST" envDefault:"localhost"`
	Port int    `env:"VALKEY_PORT" envDefault:"6379"`
}
