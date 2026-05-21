package dbinit

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

var l = logger.Get()

// Hypertable management for TimescaleDB
//
// This package supports both modern (2.19.3+) and legacy TimescaleDB versions:
// - Modern: Uses CREATE TABLE WITH (tsdb.hypertable, ...) syntax
// - Legacy: Uses the 3-step process with create_hypertable() function
//
// The system automatically detects the TimescaleDB version and uses the appropriate method.
//
// Recommended TimescaleDB versions:
// - Community Edition 2.19.3+ (TimescaleDB)
// - Cloud edition (Tiger Data)
//

// ConvertToHypertables is a method that converts the given table names to hypertables.
//
// This function will only start this process however the whole process will run through the 3 steps.
// This is first step in the process.
// Parameters:
// - tableNames: a slice of table names to convert to hypertables
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program.
//
// The function will only set the hypertable chunk to 1 week, this is pretty much the default
// however this interval should be defined mostly by developer and system specs.
// The indexed data on a weekly basis must not be more than 25% of system RAM memory.
// However depending on the data and the system specs this might need to be adjusted to be shorter or longer.
// This is all in alpha stage and might be adjusted later. For now it will be hard coded to 1 week since this
// was an optimal setting for the cosmos indexer.
// For addition info search the Tiger Data for more info.
func (init *DBInitializer) ConvertToHypertables(tableNames []string) {
	for _, tableName := range tableNames {
		sql := fmt.Sprintf("SELECT create_hypertable('%s', 'timestamp', chunk_time_interval => INTERVAL '1 weeks')", tableName)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Msgf(
					"failed to convert table %s to hypertable: %v", tableName, err,
				)
		}
	}
}

// AlterCompressionSegments is a method that alters the compression segments for the given tables
//
// This function will only start this process however the whole process will run through the 3 steps
// This is second step in the process
// Parameters:
// - tables: a map of table names to their columns
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// The function will only set the compression segments to the given columns.
// The columns will be hard encoded for now depending on the table.
func (init *DBInitializer) AlterCompressionSegments(tables map[string][]string) {
	for tableName, columns := range tables {
		columnsString := strings.Join(columns, ", ")
		sql := fmt.Sprintf(
			`
			ALTER TABLE %s SET (
				timescaledb.enable_columnstore,
				timescaledb.segmentby = %s,
				timescaledb.orderby = 'timestamp DESC'
			);
			`, tableName, columnsString)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Msgf(
					"failed to alter compression segments for table %s: %v", tableName, err,
				)
		}
	}
}

// AddCompressionPolicy is a method that adds the columnstore policy for the given tables.
//
// This function will only start this process however the whole process will run through the 3 steps
// This is third step in the process
// Parameters:
// - tableNames: a slice of table names to add the columnstore policy to
//
// Returns:
// - nil: if the program has a problem it will call log.Fatalf which will exit the program
//
// TThis function specifies the columnstore policy
func (init *DBInitializer) AddColumnstorePolicy(tableNames []string) {
	for _, tableName := range tableNames {
		sql := fmt.Sprintf(
			`
			CALL add_columnstore_policy('%s', INTERVAL '1 week');
			`, tableName)
		_, err := init.pool.Exec(context.Background(), sql)
		if err != nil {
			l.Error().
				Caller().
				Stack().
				Msgf(
					"failed to add columnstore policy for table %s: %v", tableName, err,
				)
		}
	}
}
