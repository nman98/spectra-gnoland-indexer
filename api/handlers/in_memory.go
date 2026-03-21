package handlers

import (
	"context"
	"log"
	"time"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
)

type ConstData struct {
	AvgBlockProdTime    float64
	TotalAddressesCount int32
}

type InMemoryHandler struct {
	db        InMemoryDbHandler
	chainName string
	done      chan bool
	interval  time.Duration
	data      *ConstData
}

// NewInMemoryHandler creates a new InMemoryHandler.
// It will start the handler and initialize the data. The interval is set to 5 minutes.
//
// Parameters:
// - db: the database handler
// - chainName: the chain name
//
// Returns:
// - the InMemoryHandler
func NewInMemoryHandler(db InMemoryDbHandler, chainName string) *InMemoryHandler {
	handler := &InMemoryHandler{
		db: db, chainName: chainName, done: make(chan bool), interval: 5 * time.Minute,
		data: &ConstData{
			AvgBlockProdTime:    0,
			TotalAddressesCount: 0,
		},
	}
	handler.Start()
	return handler
}

// Start starts the InMemoryHandler.
// It will initialize the data and start the goroutine to update the data.
//
// Parameters:
// - none
//
// Returns:
// - none
func (h *InMemoryHandler) Start() {
	// initialize the data
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	defer cancel()
	avgBlockProdTime, err := h.db.GetAvgBlockProdTime(ctx, h.chainName)
	if err != nil {
		log.Printf("error getting average block production time: %v", err)
		h.data.AvgBlockProdTime = 0
		return
	}
	h.data.AvgBlockProdTime = avgBlockProdTime

	totalAddressesCount, err := h.db.GetTotalAddressesCount(ctx, h.chainName)
	if err != nil {
		log.Printf("error getting total addresses count: %v", err)
		h.data.TotalAddressesCount = 0
		return
	}
	h.data.TotalAddressesCount = totalAddressesCount

	go func() {
		for {
			select {
			case <-h.done:
				return
			case <-time.After(h.interval):
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
				defer cancel()
				avgBlockProdTime, err := h.db.GetAvgBlockProdTime(ctx, h.chainName)
				if err != nil {
					log.Printf("error getting average block production time: %v", err)
					h.data.AvgBlockProdTime = 0
					continue
				}
				h.data.AvgBlockProdTime = avgBlockProdTime
				totalAddressesCount, err := h.db.GetTotalAddressesCount(ctx, h.chainName)
				if err != nil {
					log.Printf("error getting total addresses count: %v", err)
					h.data.TotalAddressesCount = 0
					continue
				}
				h.data.TotalAddressesCount = totalAddressesCount
			}
		}
	}()
}

func (h *InMemoryHandler) Stop() {
	h.done <- true
}

func (h *InMemoryHandler) GetAvgBlockProdTime(
	ctx context.Context,
	input *humatypes.GetAvgBlockProdTimeInput,
) (*humatypes.GetAvgBlockProdTimeOutput, error) {
	return &humatypes.GetAvgBlockProdTimeOutput{
		Body: humatypes.GetAvgBlockBody{AvgBlockProdTime: h.data.AvgBlockProdTime},
	}, nil
}

func (h *InMemoryHandler) GetTotalAddressesCount(
	ctx context.Context,
	input *humatypes.GetTotalAddressesCountInput,
) (*humatypes.GetTotalAddressesCountOutput, error) {
	return &humatypes.GetTotalAddressesCountOutput{
		Body: humatypes.GetTotalAddressesCountBody{TotalAddressesCount: h.data.TotalAddressesCount},
	}, nil
}
