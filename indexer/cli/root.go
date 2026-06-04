package cli

import (
	"github.com/spf13/cobra"
)

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var RootCmd = &cobra.Command{
	Use:           "indexer",
	Short:         "Spectra Gnoland Indexer",
	Long:          "A blockchain indexer for Gnoland that processes blocks and transactions.",
	Version:       Version + " (commit: " + Commit + ")",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(setupCmd)
}
