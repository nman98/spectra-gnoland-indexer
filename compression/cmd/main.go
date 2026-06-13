package main

import (
	"log"
	"os"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/compression/train"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "train",
	Short: "Train",
	Long: `Train tool for Zstd dictionary training the Spectra Gnoland Indexer.
	
	Please note that the training process will only ever be done prior to the indexer being deployed and only be
	done once, or if you want to re-train the dictionary. 

	Purpose of this cli is to collect the data from the database where it would gather all of the transaction events
	and then build the zstd dictionary from the events. By combining protobuf serialization and zstd dictionary,
	you can achieve a very efficient compression of the data. In theory you could do this manually by using zstd cli,
	however then you would need to create some logic to pull the data from the RPC endpoint and serialize the data into protobuf format.
	And then you would need to use the zstd cli to build the dictionary from the serialized data.

	However this cli is a lot more efficient since it is doing all of the work in a single step
	and it is a lot more convenient to use.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Training Zstd dictionary")
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config path: %v", err)
		}
		amount, err := cmd.Flags().GetUint64("amount")
		if err != nil {
			log.Fatalf("failed to get amount: %v", err)
		}
		chainName, err := cmd.Flags().GetString("chain-name")
		if err != nil {
			log.Fatalf("failed to get chain name: %v", err)
		}
		dictPath, err := cmd.Flags().GetString("dict-path")
		if err != nil {
			log.Fatalf("failed to get dict path: %v", err)
		}

		loadedConfig, err := train.LoadTrainingConfig(&configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		dbConfig := timescaledb.DatabasePoolConfig{
			Host:                      loadedConfig.Host,
			Port:                      loadedConfig.Port,
			User:                      loadedConfig.User,
			Password:                  loadedConfig.Password,
			Dbname:                    loadedConfig.Dbname,
			Sslmode:                   loadedConfig.Sslmode,
			PoolMaxConns:              loadedConfig.PoolMaxConns,
			PoolMinConns:              loadedConfig.PoolMinConns,
			PoolMaxConnLifetime:       loadedConfig.PoolMaxConnLifetime,
			PoolMaxConnIdleTime:       loadedConfig.PoolMaxConnIdleTime,
			PoolHealthCheckPeriod:     loadedConfig.PoolHealthCheckPeriod,
			PoolMaxConnLifetimeJitter: loadedConfig.PoolMaxConnLifetimeJitter,
		}
		db := train.InitDatabase(dbConfig)

		events, err := train.CollectEvents(db, chainName, amount)
		if err != nil {
			log.Fatalf("failed to collect events: %v", err)
		}
		dict, err := train.BuildZstdDict(events)
		if err != nil {
			log.Fatalf("failed to build zstd dictionary: %v", err)
		}
		err = os.WriteFile(dictPath, dict, 0644)
		if err != nil {
			log.Fatalf("failed to write zstd dictionary: %v", err)
		}
		log.Println("Zstd dictionary built")
		log.Println("Zstd dictionary written to: ", dictPath)
	},
}

func init() {
	rootCmd.Flags().StringP("config", "c", "", "the path to the config file")
	rootCmd.Flags().Uint64P("amount", "a", 1000, "the amount of events to collect")
	rootCmd.Flags().StringP("chain-name", "n", "gnoland", "the name of the chain")
	rootCmd.Flags().StringP("dict-path", "d", "events.zstd.bin", "the path to the zstd dictionary")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
