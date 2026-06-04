package rpcclient

import (
	"fmt"
	"strconv"
	"time"
)

// Test Builder Functions for easy synthetic data creation

// NewTestBlockResponse creates a BlockResponse with sensible defaults for testing
func NewTestBlockResponse(height uint64, chainID string) *BlockResponse {
	heightStr := strconv.FormatUint(height, 10)
	timestamp := time.Now()

	return &BlockResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: BlockResult{
			BlockMeta: BlockMeta{
				BlockID: BlockID{
					Hash: fmt.Sprintf("test-block-hash-%d", height),
					Parts: Parts{
						Total: "1",
						Hash:  fmt.Sprintf("test-parts-hash-%d", height),
					},
				},
				Header: BlockHeader{
					Version:         "test-version",
					ChainID:         chainID,
					Height:          heightStr,
					Time:            timestamp,
					NumTxs:          "0",
					TotalTxs:        heightStr,
					AppVersion:      "test-app",
					ProposerAddress: "test-proposer-address",
				},
			},
			Block: BlockInfo{
				Header: BlockHeader{
					Version:         "test-version",
					ChainID:         chainID,
					Height:          heightStr,
					Time:            timestamp,
					NumTxs:          "0",
					TotalTxs:        heightStr,
					AppVersion:      "test-app",
					ProposerAddress: "test-proposer-address",
				},
				Data: BlockData{
					Txs: nil, // No transactions by default
				},
				LastCommit: LastCommit{
					BlockID: BlockID{
						Hash: fmt.Sprintf("test-prev-hash-%d", height-1),
						Parts: Parts{
							Total: "1",
							Hash:  fmt.Sprintf("test-prev-parts-%d", height-1),
						},
					},
					Precommits: nil, // No precommits by default
				},
			},
		},
	}
}

// WithTransactions adds transaction hashes to a test BlockResponse
func (br *BlockResponse) WithTransactions(txHashes []string) *BlockResponse {
	br.Result.Block.Data.Txs = &txHashes
	br.Result.Block.Header.NumTxs = strconv.Itoa(len(txHashes))
	br.Result.BlockMeta.Header.NumTxs = strconv.Itoa(len(txHashes))
	return br
}

// WithPrecommits adds precommits to a test BlockResponse
func (br *BlockResponse) WithPrecommits(precommits []*Precommit) *BlockResponse {
	br.Result.Block.LastCommit.Precommits = precommits
	return br
}

// NewTestTxResponse creates a TxResponse with sensible defaults for testing
func NewTestTxResponse(hash string, height uint64) *TxResponse {
	heightStr := strconv.FormatUint(height, 10)

	return &TxResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: TxResultData{
			Hash:   hash,
			Height: heightStr,
			Index:  0,
			TxResult: TxResult{
				ResponseBase: ResponseBase{
					Error:  nil,
					Data:   "test-data",
					Events: []Event{}, // No events by default
					Log:    "test-log",
					Info:   "test-info",
				},
				GasWanted: "100000",
				GasUsed:   "50000",
			},
			Tx: "test-tx-data",
		},
	}
}

// WithEvents adds events to a test TxResponse
func (tr *TxResponse) WithEvents(events []Event) *TxResponse {
	tr.Result.TxResult.ResponseBase.Events = events
	return tr
}

// WithError adds an error to a test TxResponse
func (tr *TxResponse) WithError(err any) *TxResponse {
	tr.Result.TxResult.ResponseBase.Error = err
	return tr
}

// NewTestEvent creates a test Event with sensible defaults
// should be part of the integration tests but keep it here for now as an idea
func NewTestEvent(eventType string, pkgPath string) Event {
	return Event{
		AtType:  "test-type",
		Type:    eventType,
		PkgPath: pkgPath,
		Attrs: []EventAttribute{
			{Key: "test-key", Value: "test-value"},
		},
	}
}

// NewTestPrecommit creates a test Precommit with sensible defaults
// should be part of the integration tests but keep it here for now as an idea
func NewTestPrecommit(validatorAddress string, height uint64) *Precommit {
	return &Precommit{
		Type:             1,
		Height:           strconv.FormatUint(height, 10),
		Round:            "0",
		BlockID:          BlockID{Hash: fmt.Sprintf("test-hash-%d", height), Parts: Parts{Total: "1", Hash: "test-parts"}},
		Timestamp:        time.Now(),
		ValidatorAddress: validatorAddress,
		ValidatorIndex:   "0",
		Signature:        "test-signature",
	}
}
