package cli

import (
	"fmt"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var allowedLogLevels = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
}

var rootCmd = &cobra.Command{
	Use:           "indexer",
	Short:         "Spectra Gnoland Indexer",
	Long:          "A blockchain indexer for Gnoland that processes blocks and transactions.",
	Version:       Version + " (commit: " + Commit + ")",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevelStr, err := cmd.Root().PersistentFlags().GetString("log-level")
		if err != nil {
			return fmt.Errorf("failed to get log level flag: %w", err)
		}
		lvl, ok := allowedLogLevels[logLevelStr]
		if !ok {
			return fmt.Errorf("invalid log level %q: must be one of debug, info, warn, error, fatal", logLevelStr)
		}
		logger.Init(logger.Config{
			Level:       lvl,
			ServiceName: "spectra-indexer",
			Pretty:      true,
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)

	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "define log level (debug, info, warn, error, fatal)")
}

func RootCmd() *cobra.Command {
	return rootCmd
}
