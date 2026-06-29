package decoder

import (
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
)

type BasicTxData struct {
	TxHash []byte
	// gno addresses
	Signers       []string
	Memo          string
	Fee           s.Amount
	TotalMsgCount int
}

type Coin struct {
	Amount int64
	Denom  string
}

// AddressResolver interface to make the code testable and flexible
// the iterface is related to the type struct AddressCache.
//
// The only method needed is GetAddress which returns the address id
type AddressResolver interface {
	GetAddress(address string) int32
}

// DecodedMsg struct to hold the basic data and messages of the decoded message.
//
// The struct contains the basic data and messages of the decoded message.
// The basic data contains the tx hash, signers, memo, and fee, while
// the messages contains the decoded messages related data.
// The messages are stored in a map with the message type as the key and the message as the value.
type DecodedMsg struct {
	BasicData BasicTxData
	Messages  []map[string]any
}
