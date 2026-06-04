package rpcclient

import (
	"fmt"
	"strconv"
)

// Named structs for TxResponse
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Event struct {
	AtType  string           `json:"@type"`
	Type    string           `json:"type"`
	Attrs   []EventAttribute `json:"attrs"`
	PkgPath string           `json:"pkg_path"`
}

type ResponseBase struct {
	Error  any     `json:"Error"`
	Data   string  `json:"Data"`
	Events []Event `json:"Events"`
	Log    string  `json:"Log"`
	Info   string  `json:"Info"`
}

type TxResult struct {
	ResponseBase ResponseBase `json:"ResponseBase"`
	GasWanted    string       `json:"GasWanted"`
	GasUsed      string       `json:"GasUsed"`
}

type TxResultData struct {
	Hash     string   `json:"hash"`
	Height   string   `json:"height"`
	Index    int      `json:"index"`
	TxResult TxResult `json:"tx_result"`
	Tx       string   `json:"tx"`
}

type TxResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  TxResultData  `json:"result"`
}

// Helper methods for TxResponse to safely access nested data
func (tr *TxResponse) GetHash() string {
	if tr == nil {
		return ""
	}
	return tr.Result.Hash
}

func (tr *TxResponse) GetHeight() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.Height, 10, 64)
}

func (tr *TxResponse) GetEvents() []Event {
	if tr == nil {
		return nil
	}
	return tr.Result.TxResult.ResponseBase.Events
}

func (tr *TxResponse) GetGasWanted() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.TxResult.GasWanted, 10, 64)
}

func (tr *TxResponse) GetGasUsed() (uint64, error) {
	if tr == nil {
		return 0, fmt.Errorf("TxResponse is nil")
	}
	return strconv.ParseUint(tr.Result.TxResult.GasUsed, 10, 64)
}

func (tr *TxResponse) GetTx() string {
	if tr == nil {
		return ""
	}
	return tr.Result.Tx
}

func (tr *TxResponse) GetIndex() int {
	if tr == nil {
		return 0
	}
	return tr.Result.Index
}

func (tr *TxResponse) IsValid() bool {
	return tr != nil && tr.Error == nil
}

func (tr *TxResponse) HasError() bool {
	return tr != nil && tr.Result.TxResult.ResponseBase.Error != nil
}
