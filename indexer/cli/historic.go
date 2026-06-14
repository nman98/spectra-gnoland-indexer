package cli

import (
	"fmt"

	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/spf13/cobra"
)

var historicCmd = &cobra.Command{
	Use:   "historic",
	Short: "Run the indexer in historic mode",
	Long: `Runs the spectra indexer in historic mode, processing blocks from a given height to a given height.
	The historic mode takes in starting height point and a finishing height. It should be used to 
	sync up the database to the latest block height. 
	
	It can also be useful if you want to index blockchain partially and work with data for any kind of testing
	or partial scan of the chain where you want to index from a certain height to a certain height.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()
		l.Info().Msg("running in historic mode")

		configPath, rateLimitFlags, compressEvents, err := parseCommonRunFlags(cmd)
		if err != nil {
			return err
		}

		fromHeight, err := cmd.Flags().GetUint64("from-height")
		if err != nil {
			l.Error().Err(err).Msg("failed to get from height")
			return err
		}
		toHeight, err := cmd.Flags().GetUint64("to-height")
		if err != nil {
			l.Error().Err(err).Msg("failed to get to height")
			return err
		}

		runningFlags := mainTypes.RunningFlags{
			RunningMode:        "historic",
			SkipInitialDbCheck: false,
			CompressEvents:     compressEvents,
			FromHeight:         fromHeight,
			ToHeight:           toHeight,
		}

		launchIndexer(configPath, rateLimitFlags, runningFlags)
		return nil
	},
}

func init() {
	historicCmd.Flags().Uint64P("from-height", "f", 1, "starting block height")
	historicCmd.Flags().Uint64P("to-height", "o", 1000, "ending block height")

	if err := historicCmd.MarkFlagRequired("from-height"); err != nil {
		panic(fmt.Sprintf("failed to mark from-height as required: %v", err))
	}
	if err := historicCmd.MarkFlagRequired("to-height"); err != nil {
		panic(fmt.Sprintf("failed to mark to-height as required: %v", err))
	}
}
