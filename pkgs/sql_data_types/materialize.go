package sql_data_types

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

type TxCounter struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	Count      int64     `mt:"transaction_count" fn:"count(*)"`
}

func (tc TxCounter) TableName() string {
	return "tx_counter"
}

func (tc TxCounter) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(tc, tc.TableName())
}

func (tc TxCounter) TableColumns() []string {
	return aggColumns(tc)
}

func (tc TxCounter) TableFunctions() []string {
	return aggFunctions(tc)
}

func (tc TxCounter) GroupBy() []string {
	return aggGroupBy(tc)
}

// Timescaledb continuos aggregation requires source table to be specified
// this is the source table for the continuous aggregation
func (tc TxCounter) FromTable() string {
	return "transaction_general"
}

// Specification for timescaledb continuous aggregation policy
//
// Usage:
// This is used to build the SQL for the aggregation policy.
//
// Returns:
//   - tableName: the name of the table to aggregate
//   - startOffset: the start offset for the aggregation
//   - endOffset: the end offset for the aggregation
//   - interval: the interval for the aggregation
func (tc TxCounter) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := ""
	if startOffset != nil {
		formattedStartOffset = fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	}
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return tc.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type FeeVolume struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"2"`
	FeeDenom   string    `mt:"denom" fn:"fee_denom" gb:"1"`
	FeeVolume  int64     `mt:"volume" fn:"sum(fee_amount)"`
}

func (dfv FeeVolume) TableName() string {
	return "fee_volume"
}

func (dfv FeeVolume) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dfv, dfv.TableName())
}

func (dfv FeeVolume) TableColumns() []string {
	return aggColumns(dfv)
}

func (dfv FeeVolume) TableFunctions() []string {
	return aggFunctions(dfv)
}

func (dfv FeeVolume) GroupBy() []string {
	return aggGroupBy(dfv)
}

func (dfv FeeVolume) FromTable() string {
	return "transaction_general"
}

func (dfv FeeVolume) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := ""
	if startOffset != nil {
		formattedStartOffset = fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	}
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return dfv.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type DailyActiveAccounts struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 day', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	AccCount   int64     `mt:"active_account_count" fn:"count(DISTINCT address)"`
}

func (dac DailyActiveAccounts) TableName() string {
	return "daily_active_accounts"
}

func (dac DailyActiveAccounts) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dac, dac.TableName())
}

func (dac DailyActiveAccounts) TableColumns() []string {
	return aggColumns(dac)
}

func (dac DailyActiveAccounts) TableFunctions() []string {
	return aggFunctions(dac)
}

func (dac DailyActiveAccounts) GroupBy() []string {
	return aggGroupBy(dac)
}

func (dac DailyActiveAccounts) FromTable() string {
	return "address_tx"
}

func (dac DailyActiveAccounts) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if endOffset == nil {
		d := time.Hour
		endOffset = &d
	}
	if interval == nil {
		d := 30 * time.Minute
		interval = &d
	}
	formattedStartOffset := ""
	if startOffset != nil {
		formattedStartOffset = fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	}
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return dac.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type ValidatorSigningCounter struct {
	TimeBucket  time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', vbs.timestamp)" gb:"0"`
	ChainName   string    `mt:"chain_name" gb:"1"`
	ValidatorId int32     `mt:"validator_id" gb:"2"`
	BlockSigned int64     `mt:"blocks_signed" fn:"count(*)"`
}

func (vds ValidatorSigningCounter) TableName() string {
	return "validator_signing_counter"
}

func (vds ValidatorSigningCounter) FromTable() string {
	return "validator_block_signing"
}

func (vds ValidatorSigningCounter) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(vds, vds.TableName())
}

func (vds ValidatorSigningCounter) TableColumns() []string {
	return aggColumns(vds)
}

func (vds ValidatorSigningCounter) TableFunctions() []string {
	return aggFunctions(vds)
}

func (vds ValidatorSigningCounter) GroupBy() []string {
	return aggGroupBy(vds)
}

func (vds ValidatorSigningCounter) FromTableAlias() string { return "vbs" }
func (vds ValidatorSigningCounter) LateralJoins() []dbinit.LateralJoinDef {
	return []dbinit.LateralJoinDef{
		{
			Kind:       "CROSS JOIN LATERAL",
			Expression: "unnest(vbs.signed_vals)",
			Alias:      "v_id",
			Columns:    []string{"validator_id"},
		},
	}
}

func (vds ValidatorSigningCounter) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := ""
	if startOffset != nil {
		formattedStartOffset = fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	}
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return vds.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type BlockCounter struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	BlockCount int64     `mt:"block_count" fn:"count(*)"`
}

func (dbc BlockCounter) TableName() string {
	return "block_counter"
}

func (dbc BlockCounter) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dbc, dbc.TableName())
}

func (dbc BlockCounter) TableColumns() []string {
	return aggColumns(dbc)
}

func (dbc BlockCounter) TableFunctions() []string {
	return aggFunctions(dbc)
}

func (dbc BlockCounter) GroupBy() []string {
	return aggGroupBy(dbc)
}

func (dbc BlockCounter) FromTable() string {
	return "blocks"
}

func (dbc BlockCounter) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := ""
	if startOffset != nil {
		formattedStartOffset = fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	}
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return dbc.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

func aggColumns(v any) []string {
	t := reflect.TypeOf(v)
	cols := make([]string, 0, t.NumField())
	for field := range t.Fields() {
		cols = append(cols, field.Tag.Get("mt"))
	}
	return cols
}
func aggFunctions(v any) []string {
	t := reflect.TypeOf(v)
	fns := make([]string, 0, t.NumField())
	for field := range t.Fields() {
		fn := field.Tag.Get("fn")
		if fn == "" {
			fn = "noop"
		}
		fns = append(fns, fn)
	}
	return fns
}

func aggGroupBy(v any) []string {
	t := reflect.TypeOf(v)
	type entry struct {
		idx  int
		name string
	}
	var entries []entry
	maxIdx := -1
	for f := range t.Fields() {
		idx := f.Tag.Get("gb")
		if idx == "" {
			continue
		}
		idxInt, _ := strconv.Atoi(idx)
		entries = append(entries, entry{idxInt, f.Tag.Get("mt")})
		if idxInt > maxIdx {
			maxIdx = idxInt
		}
	}
	if maxIdx < 0 {
		return nil
	}
	gb := make([]string, maxIdx+1)
	for _, e := range entries {
		gb[e.idx] = e.name
	}
	return gb
}
