package synthetic

import (
	"strconv"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/generator"
)

type ResponseMaker struct {
	generator *generator.DataGenerator
}

func NewResponseMaker(generator *generator.DataGenerator) *ResponseMaker {
	return &ResponseMaker{
		generator: generator,
	}
}

type GenBlockInput struct {
	Height          uint64
	ChainID         string
	Timestamp       time.Time
	TxsRaw          []string
	ProposerAddress string
	SignedValidators []string
}

func (rm *ResponseMaker) GenerateBlockResponse(input GenBlockInput) *rpcClient.BlockResponse {
	blockHash := rm.generator.GenerateBlockHash()
	lastBlockHash := rm.generator.GenerateBlockHash()
	partsHash := ""
	lastPartsHash := ""
	lastCommitHash := ""
	dataHash := ""
	validatorsHash := ""
	nextValidatorsHash := ""
	consensusHash := ""
	appHash := ""
	lastResultsHash := ""

	// Build minimal precommits so ProcessValidatorAddresses can discover
	// all validator addresses without needing any RNG calls.
	precommits := make([]*rpcClient.Precommit, 0, len(input.SignedValidators))
	for i, validator := range input.SignedValidators {
		precommits = append(precommits, &rpcClient.Precommit{
			ValidatorAddress: validator,
			ValidatorIndex:   strconv.Itoa(i),
		})
	}

	block := rpcClient.BlockResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: rpcClient.BlockResult{
			BlockMeta: rpcClient.BlockMeta{
				BlockID: rpcClient.BlockID{
					Hash: blockHash,
					Parts: rpcClient.Parts{
						Total: "1", Hash: partsHash},
				},
				Header: rpcClient.BlockHeader{
					Version:         "1.0.0",
					ChainID:         input.ChainID,
					Height:          strconv.FormatUint(input.Height, 10),
					Time:            input.Timestamp,
					NumTxs:          "0",
					TotalTxs:        "0",
					AppVersion:      "1.0.0",
					ProposerAddress: input.ProposerAddress,
					LastBlockID:     rpcClient.BlockID{Hash: lastBlockHash, Parts: rpcClient.Parts{Total: "1", Hash: lastPartsHash}},
					LastCommitHash:     lastCommitHash,
					DataHash:           dataHash,
					ValidatorsHash:     validatorsHash,
					NextValidatorsHash: nextValidatorsHash,
					ConsensusHash:      consensusHash,
					AppHash:            appHash,
					LastResultsHash:    lastResultsHash,
				},
			},
			Block: rpcClient.BlockInfo{
				Header: rpcClient.BlockHeader{
					Version:            "1.0.0",
					ChainID:            input.ChainID,
					Height:             strconv.FormatUint(input.Height, 10),
					Time:               input.Timestamp,
					NumTxs:             "0",
					TotalTxs:           "0",
					AppVersion:         "1.0.0",
					ProposerAddress:    input.ProposerAddress,
					LastBlockID:        rpcClient.BlockID{Hash: lastBlockHash, Parts: rpcClient.Parts{Total: "1", Hash: lastPartsHash}},
					LastCommitHash:     lastCommitHash,
					DataHash:           dataHash,
					ValidatorsHash:     validatorsHash,
					NextValidatorsHash: nextValidatorsHash,
					ConsensusHash:      consensusHash,
					AppHash:            appHash,
					LastResultsHash:    lastResultsHash,
				},
				Data: rpcClient.BlockData{
					Txs: &input.TxsRaw,
				},
				LastCommit: rpcClient.LastCommit{
					BlockID: rpcClient.BlockID{
						Hash: lastBlockHash,
						Parts: rpcClient.Parts{
							Total: "1", Hash: lastPartsHash},
					},
					Precommits: precommits,
				},
			},
		},
	}
	return &block
}

type GenTransactionInput struct {
	TxRaw  string
	TxHash string
	Height uint64
	Events *generator.TxEvents
}

func (rm *ResponseMaker) GenerateTransactionResponse(input GenTransactionInput) *rpcClient.TxResponse {
	// because I pretty much complicated this part of the code the
	// program will now need to pull the data from proto events
	// and convert them to rpc client events
	// this sucks but it is what it is...

	rpcEvents := make([]rpcClient.Event, 0, len(input.Events.Events))
	for i := range input.Events.Events {
		event := &input.Events.Events[i]
		rpcAttributes := make([]rpcClient.EventAttribute, 0, len(event.Attributes))
		for _, attribute := range event.Attributes {
			rpcAttributes = append(rpcAttributes, rpcClient.EventAttribute{
				Key:   attribute.Key,
				Value: attribute.GetStringValue(), // most of the time it will be a string so what the hell keep it as string
			})
		}
		var pkgPath string
		if event.PkgPath != nil {
			pkgPath = *event.PkgPath
		}

		rpcEvents = append(rpcEvents, rpcClient.Event{
			AtType:  event.AtType,
			Type:    event.Type,
			Attrs:   rpcAttributes,
			PkgPath: pkgPath,
		})
	}

	txResponse := rpcClient.TxResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: rpcClient.TxResultData{
			Hash:   input.TxHash,
			Height: strconv.FormatUint(input.Height, 10),
			Index:  0,
			Tx:     input.TxRaw,
			TxResult: rpcClient.TxResult{
				ResponseBase: rpcClient.ResponseBase{
					Error:  nil,
					Data:   input.TxRaw,
					Events: rpcEvents,
				},
				GasWanted: "1000000", // Add some values for testing
				GasUsed:   "750000",
			},
		},
	}
	return &txResponse
}

type GenCommitInput struct {
	Height           uint64
	ChainID          string
	Timestamp        time.Time
	ProposerAddress  string
	SignedValidators []string
}

func (rm *ResponseMaker) GenerateCommitResponse(input GenCommitInput) *rpcClient.CommitResponse {
	// just like in the gen block we need the data for the header but not really
	// just generate the hash for block and last block
	// for others just use empty strings since this is all just to mock the similar experience
	blockHash := rm.generator.GenerateBlockHash()
	partsHash := ""
	lastCommitHash := ""
	dataHash := ""
	validatorsHash := ""
	nextValidatorsHash := ""
	consensusHash := ""
	appHash := ""
	lastResultsHash := ""

	precommits := make([]*rpcClient.Precommit, 0, len(input.SignedValidators))
	for i, validator := range input.SignedValidators {
		precommits = append(precommits, &rpcClient.Precommit{
			ValidatorAddress: validator,
			ValidatorIndex:   strconv.Itoa(i),
			Signature:        rm.generator.GenerateBlockHash(),
			Type:             2, // not sure why 2 but that is what I saw in the real data
			Height:           strconv.FormatUint(input.Height, 10),
			Round:            "0",
			BlockID:          rpcClient.BlockID{Hash: blockHash, Parts: rpcClient.Parts{Total: "1", Hash: partsHash}},
			Timestamp:        input.Timestamp,
		})
	}

	commitResponse := rpcClient.CommitResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: rpcClient.CommitResult{
			SignedHeader: rpcClient.SignedHeader{
				Header: rpcClient.BlockHeader{
					Height:             strconv.FormatUint(input.Height, 10),
					Time:               input.Timestamp,
					ChainID:            input.ChainID,
					ProposerAddress:    input.ProposerAddress,
					LastCommitHash:     lastCommitHash,
					DataHash:           dataHash,
					ValidatorsHash:     validatorsHash,
					NextValidatorsHash: nextValidatorsHash,
					ConsensusHash:      consensusHash,
					AppHash:            appHash,
					LastResultsHash:    lastResultsHash,
				},
				Commit: rpcClient.Commit{
					BlockID:    rpcClient.BlockID{Hash: blockHash, Parts: rpcClient.Parts{Total: "1", Hash: partsHash}},
					Precommits: precommits,
				},
			},
		},
	}
	return &commitResponse
}
