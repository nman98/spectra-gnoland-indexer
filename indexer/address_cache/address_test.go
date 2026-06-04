package addresscache

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"sort"
	"testing"
)

type mockDB struct {
	// record calls
	findExistingCalls []struct {
		addresses  []string
		chain      string
		validators bool
	}
	insertCalls []struct {
		addresses  []string
		chain      string
		validators bool
	}
	getAllCalls []struct {
		chain      string
		validators bool
	}

	// behavior controls
	existing map[string]int32
	// insert errors by attempt index (global)
	insertErrors []error
	// per-address insert errors when called one by one
	perAddressInsertError map[string]error
}

func (m *mockDB) FindExistingAccounts(ctx context.Context, addresses []string, chainName string, searchValidators bool) (map[string]int32, error) {
	m.findExistingCalls = append(m.findExistingCalls, struct {
		addresses  []string
		chain      string
		validators bool
	}{append([]string(nil), addresses...), chainName, searchValidators})

	res := map[string]int32{}
	for _, a := range addresses {
		if id, ok := m.existing[a]; ok {
			res[a] = id
		}
	}
	return res, nil
}

func (m *mockDB) InsertAddresses(ctx context.Context, addresses []string, chainName string, insertValidators bool) error {
	m.insertCalls = append(m.insertCalls, struct {
		addresses  []string
		chain      string
		validators bool
	}{append([]string(nil), addresses...), chainName, insertValidators})

	// global attempt-based error simulation
	idx := len(m.insertCalls) - 1
	if idx < len(m.insertErrors) && m.insertErrors[idx] != nil {
		return m.insertErrors[idx]
	}

	// if called one-by-one, use per-address errors
	if len(addresses) == 1 {
		if err, ok := m.perAddressInsertError[addresses[0]]; ok && err != nil {
			return err
		}
	}

	// simulate successful insert by adding to existing with new ids
	for _, a := range addresses {
		if _, ok := m.existing[a]; !ok {
			// assign a synthetic id: len(existing)+1
			m.existing[a] = int32(len(m.existing) + 1)
		}
	}
	return nil
}

func (m *mockDB) GetAllAddresses(ctx context.Context, chainName string, searchValidators bool, highestIndex *int32) (map[string]int32, int32, error) {
	m.getAllCalls = append(m.getAllCalls, struct {
		chain      string
		validators bool
	}{chainName, searchValidators})

	// return current snapshot
	// copy map to avoid aliasing in tests
	out := make(map[string]int32, len(m.existing))
	var max int32
	for k, v := range m.existing {
		out[k] = v
		if v > max {
			max = v
		}
	}
	return out, max, nil
}

func newCacheForTest(t *testing.T, existing map[string]int32, loadValidators bool) (*AddressCache, *mockDB) {
	t.Helper()
	m := &mockDB{existing: map[string]int32{}}
	maps.Copy(m.existing, existing)
	// constructor loads all existing for the flag
	c := NewAddressCache("chain", m, loadValidators)
	return c, m
}

func sorted(keys map[string]int32) []string {
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestAddressSolver_NoOpWhenAllCached(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{"a": 1, "b": 2}, false)
	before := sorted(cache.address)
	cache.AddressSolver([]string{"a", "b"}, "chain", false, 2, nil)
	after := sorted(cache.address)

	if !reflect.DeepEqual(before, after) {
		t.Fatalf("cache changed but should not; before=%v after=%v", before, after)
	}
	if len(m.insertCalls) != 0 {
		t.Fatalf("expected no inserts, got %d", len(m.insertCalls))
	}
}

func TestAddressSolver_SyncExistingFromDB(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{"a": 1}, false)
	// cache initially has {a:1}
	cache.AddressSolver([]string{"a", "b"}, "chain", false, 1, nil)

	if cache.GetAddress("a") != 1 {
		t.Fatalf("expected a to remain 1")
	}
	if cache.GetAddress("b") == 0 {
		t.Fatalf("expected b to be cached after solver")
	}
	// Should have attempted to insert only missing ones (b)
	if len(m.insertCalls) == 0 || len(m.insertCalls[0].addresses) != 1 || m.insertCalls[0].addresses[0] != "b" {
		t.Fatalf("expected insert for [b], got %+v", m.insertCalls)
	}
}

func TestAddressSolver_BatchInsertWithRetry(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{}, false)
	// fail first attempt, succeed second
	m.insertErrors = []error{errors.New("temp"), nil}

	cache.AddressSolver([]string{"x", "y"}, "chain", false, 2, nil)

	if cache.GetAddress("x") == 0 || cache.GetAddress("y") == 0 {
		t.Fatalf("expected x and y to be cached after retry path")
	}
	if len(m.insertCalls) != 2 {
		t.Fatalf("expected 2 batch insert attempts, got %d", len(m.insertCalls))
	}
}

func TestAddressSolver_OneByOneFallback(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{}, false)
	// cause all batch attempts to fail; retryAttempts=1 ensures we hit fallback on first/last attempt
	m.insertErrors = []error{errors.New("batch fail")}
	// one-by-one: fail for "bad", succeed for "good"
	m.perAddressInsertError = map[string]error{"bad": errors.New("fail one")}

	one := true
	cache.AddressSolver([]string{"good", "bad"}, "chain", false, 1, &one)

	if cache.GetAddress("good") == 0 {
		t.Fatalf("expected good to be cached from one-by-one path")
	}
	// bad may still be absent if all inserts failed for it
	if cache.GetAddress("bad") != 0 {
		// it could have been inserted if fetch found it; but our mock only inserts on success, so expect 0
		t.Fatalf("expected bad to not be cached")
	}
}

func TestAddressSolver_RespectsInsertValidatorsFlag(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{}, true)

	cache.AddressSolver([]string{"val1"}, "chain", true, 1, nil)

	if len(m.insertCalls) != 1 || m.insertCalls[0].validators != true {
		t.Fatalf("expected insert with validators=true, got %+v", m.insertCalls)
	}
}

func TestAddressSolver_EmptyInput(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{"a": 1}, false)
	cache.AddressSolver([]string{}, "chain", false, 2, nil)
	if len(m.insertCalls) != 0 || len(m.findExistingCalls) != 0 {
		t.Fatalf("expected no DB calls on empty input")
	}
}

func TestAddressSolver_DoesNotOverwriteExistingIDs(t *testing.T) {
	cache, _ := newCacheForTest(t, map[string]int32{"a": 10}, false)
	// attempt to reinsert same address
	cache.AddressSolver([]string{"a"}, "chain", false, 1, nil)
	if got := cache.GetAddress("a"); got != 10 {
		t.Fatalf("expected a to remain 10, got %d", got)
	}
}

func TestAddressSolver_HonorsRetryAttempts(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{}, false)
	// cause failures for all attempts
	m.insertErrors = []error{errors.New("1"), errors.New("2"), errors.New("3")}
	cache.AddressSolver([]string{"a"}, "chain", false, 3, nil)
	if len(m.insertCalls) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(m.insertCalls))
	}
}

func TestAddressSolver_CachesAfterPartialSuccess(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{}, false)
	// First batch fails, second succeeds; both x and y will be added in success path
	m.insertErrors = []error{errors.New("temp"), nil}

	cache.AddressSolver([]string{"x", "y"}, "chain", false, 2, nil)

	if cache.GetAddress("x") == 0 || cache.GetAddress("y") == 0 {
		t.Fatalf("expected x and y to be cached after success")
	}
}

func TestAddressSolver_SkipFetchWhenNothingToAdd(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{"a": 1}, false)
	// provide only already-known address so newAddresses empty -> early return
	cache.AddressSolver([]string{"a"}, "chain", false, 1, nil)

	if len(m.findExistingCalls) != 0 || len(m.insertCalls) != 0 {
		t.Fatalf("expected no DB calls when nothing to add")
	}
}
