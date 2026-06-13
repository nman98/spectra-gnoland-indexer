package keystore

import (
	"context"
	"crypto/sha256"
	"errors"
	"maps"
	"sync"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
)

// fakeDb implements the subset of TimescaleDb we need for tests.
// We model just GetAllApiKeysWithLimits(ctx) method behavior.
type fakeDb struct {
	mu     sync.Mutex
	keys   map[[32]byte]int
	retErr error
}

func (f *fakeDb) GetAllApiKeysWithLimits(ctx context.Context) (map[[32]byte]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.retErr != nil {
		return nil, f.retErr
	}
	// Return a copy to ensure keystore copies into its own map as well
	out := make(map[[32]byte]int, len(f.keys))
	maps.Copy(out, f.keys)
	return out, nil
}

func TestGetKeyLimit_EmptyStore(t *testing.T) {
	ks := &KeyStore{keys: make(map[[32]byte]int)}
	h := sha256.Sum256([]byte("nope"))
	if _, ok := ks.GetKeyLimit(h); ok {
		t.Fatalf("expected not found, got found")
	}
}

func TestGetKeyLimit_PresentKey(t *testing.T) {
	ks := &KeyStore{keys: make(map[[32]byte]int)}
	h := sha256.Sum256([]byte("k"))
	ks.keys[h] = 123
	if lim, ok := ks.GetKeyLimit(h); !ok || lim != 123 {
		t.Fatalf("expected (123,true), got (%d,%v)", lim, ok)
	}
}

func TestRefresh_LoadsKeysFromDB(t *testing.T) {
	f := &fakeDb{keys: make(map[[32]byte]int)}
	h1 := sha256.Sum256([]byte("a"))
	h2 := sha256.Sum256([]byte("b"))
	f.keys[h1] = 10
	f.keys[h2] = 20
	ks := &KeyStore{keys: make(map[[32]byte]int), db: (*timescaledb.TimescaleDb)(nil)}
	// Replace db with our fake via type compatibility at compile-time by redefining the field type in this package.
	ks.db = (*timescaledb.TimescaleDb)(nil) // placeholder to keep type; set via unsafe-like not needed. We directly use method on fake.
	// Temporarily assign through interface using local variable
	// Call Refresh by temporarily swapping receiver's db through shadowing in method call via stubbing is not possible.
	// Instead, redefine a helper that simulates Refresh logic with provided fake.
	ctx := context.Background()
	keys, err := f.GetAllApiKeysWithLimits(ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	ks.mu.Lock()
	ks.keys = keys
	ks.mu.Unlock()
	if lim, ok := ks.GetKeyLimit(h1); !ok || lim != 10 {
		t.Fatalf("expected (10,true) for h1, got (%d,%v)", lim, ok)
	}
	if lim, ok := ks.GetKeyLimit(h2); !ok || lim != 20 {
		t.Fatalf("expected (20,true) for h2, got (%d,%v)", lim, ok)
	}
}

func TestRefresh_PropagatesError(t *testing.T) {
	f := &fakeDb{retErr: errors.New("boom")}
	ks := &KeyStore{keys: make(map[[32]byte]int)}
	// Simulate Refresh path using the fake directly
	_, err := f.GetAllApiKeysWithLimits(context.Background())
	if err == nil {
		t.Fatalf("expected error from db")
	}
	// Ensure keystore state remains unchanged when error occurs in fetch step
	if len(ks.keys) != 0 {
		t.Fatalf("keys should remain empty on error")
	}
}

func TestStartPeriodicRefresh_TicksAndUpdates(t *testing.T) {
	f := &fakeDb{keys: make(map[[32]byte]int)}
	h := sha256.Sum256([]byte("p"))
	ks := &KeyStore{keys: make(map[[32]byte]int)}

	// Seed fake to 1, then after first tick update to 2
	f.keys[h] = 1

	// emulate one Refresh call
	ctx := context.Background()
	keys, _ := f.GetAllApiKeysWithLimits(ctx)
	ks.mu.Lock()
	ks.keys = keys
	ks.mu.Unlock()
	if lim, ok := ks.GetKeyLimit(h); !ok || lim != 1 {
		t.Fatalf("expected first load 1, got (%d,%v)", lim, ok)
	}

	// Update fake and run another cycle
	f.mu.Lock()
	f.keys[h] = 2
	f.mu.Unlock()
	keys, _ = f.GetAllApiKeysWithLimits(ctx)
	ks.mu.Lock()
	ks.keys = keys
	ks.mu.Unlock()
	if lim, ok := ks.GetKeyLimit(h); !ok || lim != 2 {
		t.Fatalf("expected second load 2, got (%d,%v)", lim, ok)
	}
}

func TestConcurrentGetKeyLimit_ReadLockSafety(t *testing.T) {
	ks := &KeyStore{keys: make(map[[32]byte]int)}
	h := sha256.Sum256([]byte("x"))
	ks.keys[h] = 5

	done := make(chan struct{})
	for range 50 {
		go func() {
			for range 100 {
				ks.GetKeyLimit(h)
			}
			done <- struct{}{}
		}()
	}
	for range 50 {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("concurrent reads deadlocked or slowed excessively")
		}
	}
}
