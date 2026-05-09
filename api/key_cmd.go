package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/key"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var allowedSslModes = []string{"disable", "require", "verify-ca", "verify-full", "allow", "prefer"}

type keyDbParams struct {
	host     string
	port     int
	user     string
	name     string
	password string
	sslMode  string
}

func (p *keyDbParams) createDatabaseConfig() database.DatabasePoolConfig {
	return database.DatabasePoolConfig{
		Host:                      p.host,
		Port:                      p.port,
		User:                      p.user,
		Dbname:                    p.name,
		Password:                  p.password,
		Sslmode:                   p.sslMode,
		PoolMaxConns:              5,
		PoolMinConns:              1,
		PoolMaxConnLifetime:       5 * time.Minute,
		PoolMaxConnIdleTime:       2 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 30 * time.Second,
	}
}

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "API key management",
	Long:  "Create, disable, enable, and list API keys for rate-limited access.",
}

func init() {
	keyCmd.AddCommand(keyCreateCmd)
	keyCmd.AddCommand(keyDisableCmd)
	keyCmd.AddCommand(keyEnableCmd)
	keyCmd.AddCommand(keyListCmd)

	for _, cmd := range []*cobra.Command{keyCreateCmd, keyDisableCmd, keyEnableCmd, keyListCmd} {
		cmd.Flags().StringP("db-host", "b", "", "Database host (default: localhost, env: KEY_DB_HOST)")
		cmd.Flags().IntP("db-port", "p", 0, "Database port (default: 5432, env: KEY_DB_PORT)")
		cmd.Flags().StringP("db-user", "u", "", "Database user (default: keymgr, env: KEY_DB_USER)")
		cmd.Flags().StringP("db-name", "d", "", "Database name (default: gnoland, env: KEY_DB_NAME)")
		cmd.Flags().StringP("ssl-mode", "s", "", "SSL mode (default: disable)")
	}

	keyCreateCmd.Example = "	api key create my-app 10000"
	keyDisableCmd.Example = "	api key disable my-app"
	keyEnableCmd.Example = "	api key enable my-app"
}

func parseKeyDbFlags(cmd *cobra.Command) (*keyDbParams, error) {
	params := &keyDbParams{}

	params.host, _ = cmd.Flags().GetString("db-host")
	params.port, _ = cmd.Flags().GetInt("db-port")
	params.user, _ = cmd.Flags().GetString("db-user")
	params.sslMode, _ = cmd.Flags().GetString("ssl-mode")
	params.name, _ = cmd.Flags().GetString("db-name")

	if params.host == "" {
		if v := os.Getenv("KEY_DB_HOST"); v != "" {
			params.host = v
		}
	}
	if params.port == 0 {
		if v := os.Getenv("KEY_DB_PORT"); v != "" {
			if _, err := fmt.Sscanf(v, "%d", &params.port); err != nil {
				return nil, fmt.Errorf("failed to scan port: %w", err)
			}
		}
	}
	if params.user == "" {
		if v := os.Getenv("KEY_DB_USER"); v != "" {
			params.user = v
		}
	}
	if params.name == "" {
		if v := os.Getenv("KEY_DB_NAME"); v != "" {
			params.name = v
		}
	}

	if params.host == "" {
		params.host = "localhost"
	}
	if params.port == 0 {
		params.port = 5432
	}
	if params.user == "" {
		params.user = "keymgr"
	}
	if params.name == "" {
		params.name = "gnoland"
	}
	if params.sslMode == "" {
		params.sslMode = "disable"
	}

	if !slices.Contains(allowedSslModes, params.sslMode) {
		return nil, fmt.Errorf("invalid ssl mode: %s", params.sslMode)
	}
	if params.port < 1 || params.port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", params.port)
	}

	return params, nil
}

func keyPromptPassword() (string, error) {
	if v := os.Getenv("KEY_DB_PASSWORD"); v != "" {
		return v, nil
	}
	fmt.Print("Enter database password: ")
	pw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	return string(pw), nil
}

func connectKeyDb(cmd *cobra.Command) (*database.TimescaleDb, error) {
	params, err := parseKeyDbFlags(cmd)
	if err != nil {
		return nil, err
	}
	params.password, err = keyPromptPassword()
	if err != nil {
		return nil, err
	}
	return database.NewTimescaleDbSetup(params.createDatabaseConfig()), nil
}

var keyCreateCmd = &cobra.Command{
	Use:   "create NAME RPM_LIMIT",
	Short: "Create a new API key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" {
			return fmt.Errorf("NAME must be non-empty")
		}
		rpmLimit, err := strconv.Atoi(args[1])
		if err != nil || rpmLimit <= 0 {
			return fmt.Errorf("RPM_LIMIT must be a positive integer")
		}

		rawKey, prefix, hash, err := key.GenerateApiKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}

		db, err := connectKeyDb(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err = db.InsertApiKey(ctx, database.KeyParams{
			RpmLimit: rpmLimit,
			Name:     name,
			Prefix:   prefix,
			Hash:     hash,
		})
		if err != nil {
			return fmt.Errorf("failed to insert API key: %w", err)
		}

		log.Printf("API key created successfully")
		log.Printf("\tName:      %s", name)
		log.Printf("\tPrefix:    %s", prefix)
		log.Printf("\tRPM Limit: %d", rpmLimit)
		log.Printf("\tKey:       %s", rawKey)
		log.Printf("")
		log.Printf("Save this key now! It cannot be retrieved later.")

		return nil
	},
}

var keyDisableCmd = &cobra.Command{
	Use:   "disable NAME",
	Short: "Disable an API key by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" {
			return fmt.Errorf("NAME must be non-empty")
		}

		db, err := connectKeyDb(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := db.DisableKeyByName(ctx, name); err != nil {
			return fmt.Errorf("failed to disable key: %w", err)
		}

		log.Printf("API key %q disabled", name)
		return nil
	},
}

var keyEnableCmd = &cobra.Command{
	Use:   "enable NAME",
	Short: "Enable an API key by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" {
			return fmt.Errorf("NAME must be non-empty")
		}

		db, err := connectKeyDb(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := db.EnableKeyByName(ctx, name); err != nil {
			return fmt.Errorf("failed to enable key: %w", err)
		}

		log.Printf("API key %q enabled", name)
		return nil
	},
}

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := connectKeyDb(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		keys, err := db.ListApiKeys(ctx)
		if err != nil {
			return fmt.Errorf("failed to list keys: %w", err)
		}

		if len(keys) == 0 {
			log.Println("No API keys found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, err = fmt.Fprintln(w, "PREFIX\tNAME\tRPM LIMIT\tACTIVE")
		if err != nil {
			return fmt.Errorf("failed to write to tabwriter: %w", err)
		}
		_, err = fmt.Fprintln(w, "------\t----\t---------\t------")
		if err != nil {
			return fmt.Errorf("failed to write to tabwriter: %w", err)
		}
		for _, k := range keys {
			_, err = fmt.Fprintf(w, "%s\t%s\t%d\t%v\n", k.Prefix, k.Name, k.RpmLimit, k.IsActive)
			if err != nil {
				return fmt.Errorf("failed to write to tabwriter: %w", err)
			}
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush tabwriter: %w", err)
		}

		return nil
	},
}
