package keystore

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
)

type KeyStore struct {
	mu   sync.RWMutex
	keys map[[32]byte]int
	db   *timescaledb.TimescaleDb
}

func NewKeyStore(db *timescaledb.TimescaleDb) *KeyStore {
	return &KeyStore{
		keys: make(map[[32]byte]int),
		db:   db,
	}
}

// GetKeyLimit returns the RPM limit for a given key hash.
// Returns 0, false if the key is not found or inactive.
func (ks *KeyStore) GetKeyLimit(hash [32]byte) (int, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	limit, ok := ks.keys[hash]
	return limit, ok
}

// Refresh loads all active API keys from the database into memory.
func (ks *KeyStore) Refresh(ctx context.Context) error {
	keys, err := ks.db.GetAllApiKeysWithLimits(ctx)
	if err != nil {
		return err
	}
	ks.mu.Lock()
	ks.keys = keys
	ks.mu.Unlock()
	return nil
}

// StartPeriodicRefresh spawns a background goroutine that refreshes the
// key cache at the given interval. Errors are logged but do not stop the loop.
func (ks *KeyStore) StartPeriodicRefresh(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := ks.Refresh(ctx); err != nil {
				log.Printf("keystore: periodic refresh failed: %v", err)
			}
			cancel()
		}
	}()
}
