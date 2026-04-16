package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/config"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/keystore"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/ratelimit"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/routes"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/valkey"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

func runServe(cmd *cobra.Command, args []string) {
	var configPath string
	var err error
	var certFilePath string
	var keyFilePath string

	configPath, err = cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}
	conf, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	env, err := config.LoadEnvironment(&config.DefaultEnvFileReader{}, ".")
	if err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	certFilePath, err = cmd.Flags().GetString("cert-file")
	if err != nil {
		log.Fatalf("failed to get cert file path: %v", err)
	}
	keyFilePath, err = cmd.Flags().GetString("key-file")
	if err != nil {
		log.Fatalf("failed to get key file path: %v", err)
	}

	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.CleanPath)
	mux.Use(middleware.Compress(5, "application/json", "application/problem+json"))
	mux.Use(middleware.Heartbeat("/"))

	corsOptions := cors.Options{
		AllowedOrigins: conf.CorsAllowedOrigins,
		AllowedMethods: conf.CorsAllowedMethods,
		AllowedHeaders: conf.CorsAllowedHeaders,
		MaxAge:         conf.CorsMaxAge,
	}
	if len(corsOptions.AllowedOrigins) == 0 {
		corsOptions.AllowedOrigins = []string{"*"}
	}
	if len(corsOptions.AllowedMethods) == 0 {
		corsOptions.AllowedMethods = []string{"GET"}
	}
	if len(corsOptions.AllowedHeaders) == 0 {
		corsOptions.AllowedHeaders = []string{"Origin", "Content-Type", "Accept", "X-API-Key"}
	}
	if corsOptions.MaxAge == 0 {
		corsOptions.MaxAge = 600
	}
	mux.Use(cors.Handler(corsOptions))

	db := database.NewTimescaleDb(database.DatabasePoolConfig{
		Host:                      env.ApiDbHost,
		Port:                      env.ApiDbPort,
		User:                      env.ApiDbUser,
		Password:                  env.ApiDbPassword,
		Dbname:                    env.ApiDbName,
		Sslmode:                   env.ApiDbSslmode,
		PoolMaxConns:              env.ApiDbPoolMaxConns,
		PoolMinConns:              env.ApiDbPoolMinConns,
		PoolMaxConnLifetime:       env.ApiDbPoolMaxConnLifetime,
		PoolMaxConnIdleTime:       env.ApiDbPoolMaxConnIdleTime,
		PoolHealthCheckPeriod:     env.ApiDbPoolHealthCheckPeriod,
		PoolMaxConnLifetimeJitter: env.ApiDbPoolMaxConnLifetimeJitter,
	})

	// Initialize key store and rate limiter
	ks := keystore.NewKeyStore(db)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := ks.Refresh(ctx); err != nil {
		log.Printf("warning: initial key store refresh failed: %v", err)
	}
	cancel()

	refreshInterval := conf.KeyRefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Minute
	}
	ks.StartPeriodicRefresh(refreshInterval)

	if conf.DisableRateLimit {
		log.Printf("rate limiting disabled — skipping Valkey initialisation")
	} else {
		valkeyEnv, err := config.LoadValkeyEnvironment(&config.DefaultEnvFileReader{}, ".")
		if err != nil {
			log.Fatalf("failed to load valkey environment: %v", err)
		}

		valkeyClient, err := valkey.NewValkeyClient(valkeyEnv.Host, valkeyEnv.Port)
		if err != nil {
			log.Fatalf("failed to create valkey client: %v", err)
		}

		ipRPM := conf.IpRpmLimit
		if ipRPM == 0 {
			ipRPM = 30
		}

		rl := ratelimit.NewRateLimiter(valkeyClient, ks, ipRPM, 1*time.Minute, conf.TrustedProxies)
		// Only apply rate limiting to API routes; documentation and static
		// assets (/docs, /openapi.yaml, /favicon.ico, /) are excluded.
		rl.SetRatePaths([]string{"/v1"})
		mux.Use(rl.Middleware)
	}

	humaConfig := huma.DefaultConfig("Spectra Gnoland Indexer API", Version)
	if !conf.DisableRateLimit {
		humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"apiKey": {
				Type: "apiKey",
				Name: "X-API-Key",
				In:   "header",
				Description: `API key for authenticated access with a higher rate limit. Pass your key in the X-API-Key header.
				You can query the APi without the API key but the stricter rate limit will be applied.`,
			},
		}
		// Apply the scheme globally but make it optional (empty map = anonymous access allowed).
		humaConfig.Security = []map[string][]string{
			{"apiKey": {}},
			{},
		}
	}

	api := humachi.New(mux, humaConfig)

	openApi := api.OpenAPI()
	openApi.Info = &huma.Info{
		Title:       "Spectra Gnoland Indexer API",
		Version:     strings.TrimPrefix(Version, "v"),
		Description: "API for the Spectra Gnoland Indexer",
		Contact: &huma.Contact{
			Name:  "Cogwheel Validator",
			URL:   "https://cogwheel.zone",
			Email: "info@cogwheel.zone",
		},
		License: &huma.License{
			Name: "Apache 2.0",
			URL:  "https://github.com/Cogwheel-Validator/spectra-gnoland-indexer?tab=Apache-2.0-1-ov-file#readme",
		},
	}
	openApi.ExternalDocs = &huma.ExternalDocs{
		URL: "https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/main/docs",
	}

	blocksHandler := handlers.NewBlocksHandler(db, conf.ChainName)
	transactionsHandler := handlers.NewTransactionsHandler(db, conf.ChainName)
	addressHandler := handlers.NewAddressHandler(db, conf.ChainName)
	validatorsHandler := handlers.NewValidatorsHandler(db, conf.ChainName)
	inMemoryHandler := handlers.NewInMemoryHandler(db, conf.ChainName)

	mux.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		_, err := w.Write(favicon)
		if err != nil {
			log.Printf("failed to write favicon: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	v1Group := huma.NewGroup(api, "/v1")
	routes.RegisterBlocksRoutes(v1Group, blocksHandler, inMemoryHandler)
	routes.RegisterTransactionsRoutes(v1Group, transactionsHandler)
	routes.RegisterAddressesRoutes(v1Group, addressHandler, inMemoryHandler)
	routes.RegisterValidatorsRoutes(v1Group, validatorsHandler)
	routes.RegisterUtilsRoutes(v1Group)

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	if certFilePath != "" && keyFilePath != "" {
		log.Printf("Starting server on %s with HTTPS", addr)
		err = http.ListenAndServeTLS(addr, certFilePath, keyFilePath, mux)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	} else {
		log.Printf("Starting server on %s with HTTP", addr)
		err = http.ListenAndServe(addr, mux)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	inMemoryHandler.Stop()
	os.Exit(0)
}
