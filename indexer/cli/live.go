package cli

import (
	mainOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_operator"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Run the indexer in live mode",
	Long: `Runs the spectra indexer, listening to any new blocks and processing them.
	It will check the database for the last processed height and start from there.

	In the events the database is empty, it will start from block height 1. This can be used
	to sync up the database to the latest block height.

	However if you do not need previous data, you can run the live mode with the skip-db-check flag set to true.
	Afterwards you can run live mode normal without the skip-db-check flag.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()
		l.Info().Msg("running in live mode")

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			l.Error().Err(err).Msg("failed to get config path")
			return err
		}

		maxRequestsPerWindow, err := cmd.Flags().GetInt("max-req-per-window")
		if err != nil {
			l.Error().Err(err).Msg("failed to get max requests per window")
			return err
		}
		rateLimitWindow, err := cmd.Flags().GetDuration("rate-limit-window")
		if err != nil {
			l.Error().Err(err).Msg("failed to get rate limit window")
			return err
		}
		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			l.Error().Err(err).Msg("failed to get timeout")
			return err
		}
		compressEvents, err := cmd.Flags().GetBool("compress-events")
		if err != nil {
			l.Error().Err(err).Msg("failed to get compress events")
			return err
		}

		skipDbCheck, err := cmd.Flags().GetBool("skip-db-check")
		if err != nil {
			l.Error().Err(err).Msg("failed to get skip db check")
			return err
		}

		rateLimitFlags := mainTypes.RpcFlags{
			RequestsPerWindow: maxRequestsPerWindow,
			TimeWindow:        rateLimitWindow,
			Timeout:           timeout,
		}

		runningFlags := mainTypes.RunningFlags{
			RunningMode:        "live",
			SkipInitialDbCheck: skipDbCheck,
			CompressEvents:     compressEvents,
			FromHeight:         0,
			ToHeight:           0,
		}

		l.Info().Msg("indexer started")
		if compressEvents {
			l.Warn().Msg("compress events is enabled, this is experimental and it might slow down the data processing speed")
		}
		mainOperator.InitMainOperator(configPath, ".", rateLimitFlags, runningFlags)
		return nil
	},
}

func init() {
	liveCmd.Flags().BoolP("skip-db-check", "s", false, "skip initial database check")
}
