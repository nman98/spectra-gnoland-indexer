package cli

import (
	"time"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the indexer",
	Long:  `Run the indexer in either live or historic mode.`,
}

func init() {
	// Add subcommands
	runCmd.AddCommand(liveCmd)
	runCmd.AddCommand(historicCmd)

	// Persistent flags that apply to all run subcommands (live and historic)
	runCmd.PersistentFlags().StringP("config", "c", "config.yml", "config file path")
	runCmd.PersistentFlags().IntP("max-req-per-window", "m", 10000000, "max requests per window")
	runCmd.PersistentFlags().DurationP("rate-limit-window", "r", 1*time.Minute, "rate limit window")
	runCmd.PersistentFlags().DurationP("timeout", "t", 20*time.Second, "timeout")
	runCmd.PersistentFlags().BoolP("compress-events", "e", false, "compress events")
}
