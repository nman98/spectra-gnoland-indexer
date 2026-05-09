package mainoperator

import (
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	contextHook "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/context_hook"
	dp "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/orchestrator"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

var l = logger.Get()

// This function is not ready to be used
// this is just a placeholder
// every part of the indexer should be initialized within the main operator
func InitMainOperator(
	configPath string,
	envPath string,
	rpcFlags mainTypes.RpcFlags,
	runningFlags mainTypes.RunningFlags,
 ) {
	// load config
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		l.Fatal().Caller().Stack().Err(err).Msg("failed to load config")
	}
	// load environment
	env, err := config.LoadEnvironment(envPath)
	if err != nil {
		l.Fatal().Caller().Stack().Err(err).Msg("failed to load environment")
	}

	// get the chain name
	chainName := &conf.ChainName

	mc := initializeMajorConstructors(conf, env, *chainName, rpcFlags)

	// initialize the orchestrator
	orch := orchestrator.NewOrchestrator(
		runningFlags.RunningMode, conf, *chainName, mc.db, mc.gnoRpcClient, mc.dataProcessor, mc.queryOperator,
	)

	// Setup signal handling with proper cleanup and state dump functions
	signalHandler := contextHook.NewSignalHandler(
		func() error {
			l.Info().Msg("Starting main operator cleanup...")
			// Cleanup orchestrator first
			if err := orch.Cleanup(); err != nil {
				l.Error().Caller().Stack().Err(err).Msg("Orchestrator cleanup failed")
			}

			// Cleanup major constructors
			if err := mc.cleanup(); err != nil {
				l.Error().Caller().Stack().Err(err).Msg("Major constructors cleanup failed")
			}

			l.Info().Msg("Main operator cleanup completed")
			return nil
		},
		func() error {
			l.Info().Msg("Creating emergency state dump...")
			// Dump orchestrator state
			if err := orch.DumpState(); err != nil {
				l.Error().Caller().Stack().Err(err).Msg("Orchestrator state dump failed")
			}

			// Dump major constructors state if needed
			if err := mc.dumpState(); err != nil {
				l.Error().Caller().Stack().Err(err).Msg("Major constructors state dump failed")
			}

			l.Info().Msg("Emergency state dump completed")
			return nil
		},
	)

	// Start signal listening
	signalHandler.StartListening()
	l.Info().Msg("Signal handler started, listening for termination signals")

	// let the orchestrator do it's thing
	switch runningFlags.RunningMode {
	case "live":
		go InitDebug(signalHandler.Context())
		orch.LiveProcess(signalHandler.Context(), runningFlags.SkipInitialDbCheck, runningFlags.CompressEvents)
	case "historic":
		if runningFlags.FromHeight == 0 || runningFlags.ToHeight == 0 {
			l.Fatal().Caller().Stack().Msg("from height and to height are required for historic mode")
		} else if runningFlags.FromHeight > runningFlags.ToHeight {
			l.Fatal().Caller().Stack().Msg("from height must be less than to height")
		}
		go InitDebug(signalHandler.Context())
		// Historic processing doesn't need context cancellation in the same way,
		// but we should still respect shutdown signals
		go func() {
			<-signalHandler.Context().Done()
			l.Info().Msg("Shutdown signal received during historic processing")
		}()
		orch.HistoricProcess(runningFlags.FromHeight, runningFlags.ToHeight, runningFlags.CompressEvents)
	default:
		l.Fatal().Caller().Stack().Msg("invalid running mode, please choose between live and historic")
	}
}

// initializeDatabase is a private function to initialize the database
// it is used to initialize the database for the main operator
//
// Parameters:
//   - conf: the config
//   - env: the environment
//
// Returns:
//   - the database
//   - error if any
func initializeDatabase(conf *config.Config, env *config.Environment) *database.TimescaleDb {
	// check if the config has any null
	// if rpc is null throw an error and exit
	if conf.RpcUrl == "" {
		l.Fatal().Caller().Stack().Msg("rpc url is required")
	}
	// if pool max connections is 0 or nil set a default of 100
	if conf.PoolMaxConns == 0 {
		conf.PoolMaxConns = 100
	}
	// set to a default of 10 if not set
	if conf.PoolMinConns == 0 {
		conf.PoolMinConns = 10
	}
	// set to a default of 10 minutes if not set
	if conf.PoolMaxConnLifetime == 0 {
		conf.PoolMaxConnLifetime = 10 * time.Minute
	}
	// set to a default of 5 minutes if not set
	if conf.PoolMaxConnIdleTime == 0 {
		conf.PoolMaxConnIdleTime = 5 * time.Minute
	}
	// set to a default of 1 minute if not set
	if conf.PoolHealthCheckPeriod == 0 {
		conf.PoolHealthCheckPeriod = 1 * time.Minute
	}
	// set to a default of 1 minute if not set
	if conf.PoolMaxConnLifetimeJitter == 0 {
		conf.PoolMaxConnLifetimeJitter = 1 * time.Minute
	}

	// pull config and env data to load init the database pool

	dbConfig := database.DatabasePoolConfig{
		Host:                      env.Host,
		Port:                      env.Port,
		User:                      env.User,
		Password:                  env.Password,
		Dbname:                    env.Dbname,
		Sslmode:                   env.Sslmode,
		PoolMaxConns:              conf.PoolMaxConns,
		PoolMinConns:              conf.PoolMinConns,
		PoolMaxConnLifetime:       conf.PoolMaxConnLifetime,
		PoolMaxConnIdleTime:       conf.PoolMaxConnIdleTime,
		PoolHealthCheckPeriod:     conf.PoolHealthCheckPeriod,
		PoolMaxConnLifetimeJitter: conf.PoolMaxConnLifetimeJitter,
	}

	// no need to return error since it will throw a fatal error and exit the program
	db := database.NewTimescaleDb(dbConfig)
	return db
}

// initializeMajorConstructors is a private function to initialize the major constructors
// it is used to initialize the major constructors for the main operator
//
// Parameters:
//   - conf: the config
//   - env: the environment
//   - chainName: the chain name
//   - rpcFlags: the rpc flags
//
// Returns:
//   - the major constructors struct
func initializeMajorConstructors(
	conf *config.Config,
	env *config.Environment,
	chainName string,
	rpcFlags mainTypes.RpcFlags) *MajorConstructors {
	// initialize the rpc client
	// check the flags first
	// this is yet to be implemented but for now just set it and later fix anything
	if rpcFlags.RequestsPerWindow == 0 {
		// realistically this could be ignored
		// if this really is the case set it to 10 million since
		// this should indicate that no rate limiting is needed
		rpcFlags.RequestsPerWindow = 10000000
	}
	if rpcFlags.TimeWindow == 0 {
		// set it to a default of 1 minute
		rpcFlags.TimeWindow = 1 * time.Minute
	} else if rpcFlags.TimeWindow <= 0 {
		l.Fatal().Caller().Stack().Msg("time window must be greater than 0")
	}

	// init all of the major constructors

	// initialize the database
	db := initializeDatabase(conf, env)

	// initialize the rpc client
	gnoRpcClient, err := rpcClient.NewRateLimitedRpcClient(
		conf.RpcUrl, nil, rpcFlags.RequestsPerWindow, rpcFlags.TimeWindow,
	)
	if err != nil {
		l.Fatal().Caller().Stack().Err(err).Msg("failed to initialize rpc client")
	}

	// initialize the validator cache
	validatorCache := addressCache.NewAddressCache(chainName, db, true)

	// initialize the address cache
	addressCache := addressCache.NewAddressCache(chainName, db, false)

	// initialize the data processor
	dataProcessor := dp.NewDataProcessor(db, addressCache, validatorCache, chainName)

	// initialize the query operator
	queryOperator := query.NewQueryOperator(
		gnoRpcClient, conf.RetryAmount, conf.Pause, conf.PauseTime, conf.ExponentialBackoff,
	)

	return &MajorConstructors{
		db:             db,
		gnoRpcClient:   gnoRpcClient,
		validatorCache: validatorCache,
		addressCache:   addressCache,
		dataProcessor:  dataProcessor,
		queryOperator:  queryOperator,
	}
}

// cleanup performs cleanup operations on all major constructors
func (mc *MajorConstructors) cleanup() error {
	l.Info().Msg("Starting major constructors cleanup...")

	// Close database connection pool
	if mc.db != nil {
		l.Info().Msg("Closing database connection pool...")
		mc.db.Close()
		l.Info().Msg("Database connection pool closed successfully")
	}

	// Close RPC client (closes the rate limiter)
	if mc.gnoRpcClient != nil {
		l.Info().Msg("Closing RPC client...")
		mc.gnoRpcClient.Close()
		l.Info().Msg("RPC client closed successfully")
	}

	// Other components (caches, data processor, query operator) don't need explicit cleanup
	// as they rely on the database and RPC client connections that we've already closed
	l.Info().Msg("Address caches, data processor, and query operator don't require explicit cleanup")

	l.Info().Msg("Major constructors cleanup completed successfully")
	return nil
}

// dumpState creates a state dump of the major constructors
func (mc *MajorConstructors) dumpState() error {
	l.Info().Msg("Creating major constructors state dump...")

	// Create basic state information
	state := map[string]interface{}{
		"timestamp": time.Now(),
		"components": map[string]interface{}{
			"database":        mc.db != nil,
			"gno_rpc_client":  mc.gnoRpcClient != nil,
			"validator_cache": mc.validatorCache != nil,
			"address_cache":   mc.addressCache != nil,
			"data_processor":  mc.dataProcessor != nil,
			"query_operator":  mc.queryOperator != nil,
		},
	}

	// Add more detailed state if components support it
	if mc.gnoRpcClient != nil {
		// Try to get RPC client state if it has a method for it
		if stateProvider, ok := interface{}(mc.gnoRpcClient).(interface{ GetState() map[string]interface{} }); ok {
			state["rpc_client_state"] = stateProvider.GetState()
		}
	}

	// Create diagnostics directory if it doesn't exist
	diagDir := "diagnostics"
	if err := os.MkdirAll(diagDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagnostics directory: %w", err)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("major_constructors_dump_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(diagDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal major constructors state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write major constructors state: %w", err)
	}

	l.Info().Msgf("Major constructors state dump saved to %s", filepath)
	return nil
}
