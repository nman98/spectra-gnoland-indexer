package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cli"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Only set up OTel when an endpoint is explicitly configured.
	// Without this guard the OTLP exporter dials lazily and prints a
	// "connection refused" error on shutdown even when no collector is running.
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		if exp, err := otlploghttp.New(ctx); err == nil {
			provider := sdklog.NewLoggerProvider(
				sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
			)
			global.SetLoggerProvider(provider)
			defer provider.Shutdown(ctx) //nolint:errcheck
		}
	}

	cmd := cli.RootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		logger.Get().Fatal().Err(err).Msg("failed to execute command")
	}
}
