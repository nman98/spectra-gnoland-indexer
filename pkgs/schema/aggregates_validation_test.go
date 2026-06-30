package schema

import (
	"reflect"
	"slices"
	"sort"
	"strconv"
	"testing"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

// allowedAggTagKeys is the set of struct-tag keys the continuous-aggregate
// generator understands (mt = column alias, fn = SQL expression, gb = GROUP BY
// position). Anything else is a typo that would be silently ignored.
var allowedAggTagKeys = map[string]bool{
	"mt": true,
	"fn": true,
	"gb": true,
}

// TestAggregateTagsAreValid guards the reflection-driven continuous-aggregate
// layer (mt/fn/gb tags), the sibling of TestTableTagsAreValid for plain tables.
// A mistyped tag here would otherwise produce a silently wrong materialized view.
func TestAggregateTagsAreValid(t *testing.T) {
	for _, agg := range AllAggregates() {
		testTable(t, agg)
	}
}

func testTable(t *testing.T, agg dbinit.ContinuousAggregateDefinition) {
	rt := reflect.TypeOf(agg)
	t.Run(agg.TableName(), func(t *testing.T) {
		if rt.Kind() != reflect.Struct {
			t.Fatalf("expected struct, got %s", rt.Kind())
		}
		if agg.TableName() == "" {
			t.Error("TableName is empty")
		}
		if agg.FromTable() == "" {
			t.Error("FromTable is empty")
		}

		fieldCount := 0
		var groupByIdx []int

		for field := range rt.Fields() {
			testColumn(t, &fieldCount, &groupByIdx, field)
		}

		// GROUP BY positions must be a contiguous 0..n-1 with no gaps or dupes;
		// aggGroupBy fills a slice by index, so a gap leaves an empty column.
		sort.Ints(groupByIdx)
		for i, idx := range groupByIdx {
			if idx != i {
				t.Errorf("gb positions are not contiguous from 0: got %v", groupByIdx)
				break
			}
		}

		// The generated column/function/group-by lists must be consistent.
		columnValidation(t, agg, fieldCount, groupByIdx)
	})
}

func testColumn(
	t *testing.T,
	fieldCount *int,
	groupByIdx *[]int,
	field reflect.StructField,
) {
	if !field.IsExported() {
		return
	}
	*fieldCount++

	// Every column must declare its materialized alias.
	if mt, ok := field.Tag.Lookup("mt"); !ok || mt == "" {
		t.Errorf("field %s: missing or empty `mt` tag", field.Name)
	}

	// Reject unknown tag keys (catches typos like `m_t`/`gp`).
	for _, key := range structTagKeys(string(field.Tag)) {
		if !allowedAggTagKeys[key] {
			t.Errorf("field %s: unknown struct-tag key %q", field.Name, key)
		}
	}

	// GROUP BY positions must be non-negative integers.
	if gb, ok := field.Tag.Lookup("gb"); ok {
		idx, err := strconv.Atoi(gb)
		if err != nil || idx < 0 {
			t.Errorf("field %s: gb tag %q is not a non-negative integer", field.Name, gb)
			return
		}
		*groupByIdx = append(*groupByIdx, idx)
	}
}

func columnValidation(
	t *testing.T,
	agg dbinit.ContinuousAggregateDefinition,
	fieldCount int,
	groupByIdx []int,
) {
	if got := len(agg.TableColumns()); got != fieldCount {
		t.Errorf("TableColumns returned %d entries, expected %d (one per field)", got, fieldCount)
	}
	if slices.Contains(agg.TableColumns(), "") {
		t.Errorf("TableColumns contains an empty alias: %v", agg.TableColumns())
	}
	if got := len(agg.TableFunctions()); got != fieldCount {
		t.Errorf("TableFunctions returned %d entries, expected %d (one per field)", got, fieldCount)
	}
	if got := len(agg.GroupBy()); got != len(groupByIdx) {
		t.Errorf("GroupBy returned %d entries, expected %d", got, len(groupByIdx))
	}
	if slices.Contains(agg.GroupBy(), "") {
		t.Errorf("GroupBy has an empty entry (gap in gb positions): %v", agg.GroupBy())
	}
}
