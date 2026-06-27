package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
	"golang.org/x/term"
)

// parseCommonFlags extracts and validates common database flags
func parseCommonFlags(cmd *cobra.Command, defaultDbName string) (*dbParams, error) {
	params := &dbParams{}

	params.host, _ = cmd.Flags().GetString("db-host")
	params.port, _ = cmd.Flags().GetInt("db-port")
	params.user, _ = cmd.Flags().GetString("db-user")
	params.sslMode, _ = cmd.Flags().GetString("ssl-mode")
	params.name, _ = cmd.Flags().GetString("db-name")
	params.sslRootCert, _ = cmd.Flags().GetString("ssl-rootcert")
	params.sslCert, _ = cmd.Flags().GetString("ssl-cert")
	params.sslKey, _ = cmd.Flags().GetString("ssl-key")

	// Apply environment variable fallbacks (for CI/CD)
	if params.host == "" {
		if envHost := os.Getenv("DB_HOST"); envHost != "" {
			params.host = envHost
		}
	}
	if params.port == 0 {
		if envPort := os.Getenv("DB_PORT"); envPort != "" {
			if _, err := fmt.Sscanf(envPort, "%d", &params.port); err != nil {
				return nil, fmt.Errorf("failed to scan port: %v", err)
			}
		}
	}
	if params.user == "" {
		if envUser := os.Getenv("DB_USER"); envUser != "" {
			params.user = envUser
		}
	}
	if params.name == "" {
		if envDbName := os.Getenv("DB_NAME"); envDbName != "" {
			params.name = envDbName
		}
	}
	if params.sslRootCert == "" {
		params.sslRootCert = os.Getenv("DB_SSLROOTCERT")
	}
	if params.sslCert == "" {
		params.sslCert = os.Getenv("DB_SSLCERT")
	}
	if params.sslKey == "" {
		params.sslKey = os.Getenv("DB_SSLKEY")
	}

	// Apply defaults if still empty
	if params.sslMode == "" {
		params.sslMode = "disable"
	}
	if params.host == "" {
		params.host = "localhost"
	}
	if params.port == 0 {
		params.port = 5432
	}
	if params.user == "" {
		params.user = defaultDBUser
	}
	if params.name == "" {
		params.name = defaultDbName
	}

	// Validate
	if !slices.Contains(allowedSslModes, params.sslMode) {
		return nil, fmt.Errorf("invalid ssl mode: %s", params.sslMode)
	}
	if params.port < 1 || params.port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", params.port)
	}

	return params, nil
}

// promptPassword prompts user for password input or reads from environment
func promptPassword() (string, error) {
	// First check if password is provided via environment variable (for CI/CD)
	if envPassword := os.Getenv("DB_PASSWORD"); envPassword != "" {
		return envPassword, nil
	}

	// Interactive mode: prompt user for password
	fmt.Print("Enter the database password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	fmt.Println()
	return string(bytePassword), nil
}

// createDatabaseConfig creates a DatabasePoolConfig from dbParams
func (p *dbParams) createDatabaseConfig() timescaledb.DatabasePoolConfig {
	return timescaledb.DatabasePoolConfig{
		Host:                      p.host,
		Port:                      p.port,
		User:                      p.user,
		Dbname:                    p.name,
		Password:                  p.password,
		Sslmode:                   p.sslMode,
		SslRootCert:               p.sslRootCert,
		SslCert:                   p.sslCert,
		SslKey:                    p.sslKey,
		PoolMaxConns:              10,
		PoolMinConns:              1,
		PoolMaxConnLifetime:       10 * time.Minute,
		PoolMaxConnIdleTime:       5 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 1 * time.Minute,
	}
}

func createConfig(overwrite bool, fileName string) error {
	if fileName == "" {
		fileName = "config.yml"
	}
	absolutePath, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	fileName = absolutePath

	cfg := config.Config{
		RpcUrl:                    "http://localhost:26657",
		PoolMaxConns:              50,
		PoolMinConns:              10,
		PoolMaxConnLifetime:       5 * time.Minute,
		PoolMaxConnIdleTime:       5 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 1 * time.Minute,
		LivePooling:               5 * time.Second,
		MaxBlockChunkSize:         50,
		MaxTransactionChunkSize:   100,
		ChainName:                 "gnoland",
		RetryAmount:               &[]int{6}[0],
		Pause:                     &[]int{3}[0],
		PauseTime:                 &[]time.Duration{15 * time.Second}[0],
		ExponentialBackoff:        &[]time.Duration{2 * time.Second}[0],
	}

	yamlFile, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if _, err = os.Stat(fileName); err == nil {
		if overwrite {
			if writeErr := os.WriteFile(fileName, yamlFile, 0644); writeErr != nil {
				return fmt.Errorf("failed to overwrite config file: %w", writeErr)
			}
		} else {
			return fmt.Errorf("config file already exists, use --overwrite to overwrite it")
		}
	} else if os.IsNotExist(err) {
		if writeErr := os.WriteFile(fileName, yamlFile, 0644); writeErr != nil {
			return fmt.Errorf("failed to create config file: %w", writeErr)
		}
	} else {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	return nil
}
