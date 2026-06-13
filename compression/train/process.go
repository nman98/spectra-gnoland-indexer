package train

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

// Collect collects the transactions from the database
//
// Usage:
//
// # Used to collect the transactions from the database
//
// Parameters:
//   - db: the database connection pool
//   - chainName: the name of the chain
//   - amount: the amount of transactions to collect
//
// Returns:
//   - [][]byte: the transactions events in serialized protobuf format
//   - error: if the transactions fail to collect
func CollectEvents(db *timescaledb.TimescaleDb, chainName string, amount uint64) ([][]byte, error) {
	// define the limits and offset
	if amount > 250000 {
		return nil, fmt.Errorf("amount cannot be greater than 250000")
	}
	if amount == 0 {
		return nil, fmt.Errorf("amount cannot be 0")
	}
	var limit uint64
	var goroutines int
	limit = min(amount, 100)

	goroutines = int(math.Ceil(float64(amount) / 100))
	transactions := make([]*database.Transaction, 0)
	wg := sync.WaitGroup{}
	wg.Add(goroutines)
	mu := sync.Mutex{}
	for i := 0; i < goroutines; i++ {
		go fetchTransactionBatch(i, db, chainName, limit, &transactions, &mu, &wg)
	}
	wg.Wait()
	events, err := processEvents(transactions)
	if err != nil {
		return nil, err
	}
	log.Printf("All events collected")
	log.Printf("Collected %d events", len(events))
	return events, nil
}

// BuildZstdDict builds the zstd dict from the events
func BuildZstdDict(events [][]byte) ([]byte, error) {
	const maxHistorySize = 112 << 10

	var history []byte
	for _, e := range events {
		history = append(history, e...)
		if len(history) >= maxHistorySize {
			break
		}
	}
	if len(history) > maxHistorySize {
		history = history[:maxHistorySize]
	}

	lvl := zstd.SpeedBestCompression
	dict, err := zstd.BuildDict(zstd.BuildDictOptions{
		ID:       1,
		Contents: events,
		History:  history,
		Level:    lvl,
		DebugOut: os.Stdout,
	})
	if err != nil {
		return nil, err
	}
	return dict, nil
}

// processEvents processes the events from the transactions
//
// Parameters:
//   - transactions: the transactions to process
//
// Returns:
//   - [][]byte: the processed events
//   - error: if the events fail to process
func processEvents(transactions []*database.Transaction) ([][]byte, error) {
	events := make([][]byte, 0)
	for _, transaction := range transactions {
		txEvents := transaction.TxEvents
		protoTxEvents := &events_proto.TxEvents{
			Events: make([]*events_proto.Event, 0),
		}
		if len(txEvents) > 0 {
			for _, event := range txEvents {
				protoAttrs := make([]*events_proto.Attribute, 0)
				for _, attribute := range event.Attributes {
					protoAttrs = append(protoAttrs, events_proto.NewAttributeFromString(attribute.Key, attribute.Value))
				}
				protoEv := &events_proto.Event{
					AtType:     event.AtType,
					Type:       event.Type,
					Attributes: protoAttrs,
					PkgPath:    &event.PkgPath,
				}
				protoTxEvents.Events = append(protoTxEvents.Events, protoEv)
			}
		}
		if len(protoTxEvents.Events) == 0 {
			continue
		}
		bs, err := proto.Marshal(protoTxEvents)
		if err != nil {
			log.Printf("failed to marshal tx events: %v", err)
			continue
		}
		events = append(events, bs)
	}
	return events, nil
}

// fetchTransactionBatch fetches a batch of transactions from the database
//
// Usage:
//
// # In a goroutine to fetch a batch of transactions from the database
//
// Parameters:
//   - i: the index of the batch
//   - db: the database connection pool
//   - chainName: the name of the chain
//   - limit: the limit of the transactions to fetch
//   - transactions: the transactions to append to
//   - mu: the mutex to lock the transactions
//   - wg: the wait group to wait for the goroutine to finish
func fetchTransactionBatch(
	i int,
	db *timescaledb.TimescaleDb,
	chainName string,
	limit uint64,
	transactions *[]*database.Transaction,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	offset := uint64(i) * limit
	log.Printf("getting the transactions from %s with limit %d and offset %d", chainName, limit, offset)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	txs, err := db.GetTransactionsByOffset(ctx, chainName, limit, offset)
	if err != nil {
		log.Printf("failed to get transactions from %s with limit %d and offset %d: %v", chainName, limit, offset, err)
		return
	}
	log.Printf("got %d transactions from %s with limit %d and offset %d", len(txs), chainName, limit, offset)
	mu.Lock()
	*transactions = append(*transactions, txs...)
	mu.Unlock()
}
