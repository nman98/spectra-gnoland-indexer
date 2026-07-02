package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/config"
	"github.com/spf13/cobra"
)

// healthcheckCmd hits the API's own /health endpoint. It exists so the
// docker healthcheck can exec this binary directly (CMD form) instead of
// relying on a shell or curl, neither of which exist in the distroless image.
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check that the API server is responding",
	Long: `In the events you require a health check ping because the official Docker
API image uses distroless as a base there isn't shell pre-installed. So
instead we use the healthcheck command directly.`,
	Run: runHealthcheck,
}

func runHealthcheck(cmd *cobra.Command, args []string) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get config path: %v\n", err)
		os.Exit(1)
	}
	conf, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	client := http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/health", conf.Port)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}
}
