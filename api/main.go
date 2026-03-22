package main

import (
	_ "embed"
	"log"
	"time"

	"github.com/spf13/cobra"
)

//go:embed public/favicon.ico
var favicon []byte

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var rootCmd = &cobra.Command{
	Use:     "api",
	Short:   "Spectra Gnoland Indexer API",
	Long:    "API for the Spectra Gnoland Indexer",
	Version: Version + " (commit: " + Commit + ")",
	Run:     runServe,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "config-api.yml", "config file path")
	rootCmd.PersistentFlags().StringP("cert-file", "t", "", "certificate file path")
	rootCmd.PersistentFlags().StringP("key-file", "k", "", "key file path")

	rootCmd.AddCommand(keyCmd)
}

func main() {
	// Important to force because all of the data is in UTC time recorded.
	time.Local = time.UTC

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
