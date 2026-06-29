package decoder

import (
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type BasicTxData struct {
	TxHash []byte
	// gno addresses
	Signers       []string
	Memo          string
	Fee           s.Amount
	TotalMsgCount int
}

// AddressResolver interface to make the code testable and flexible
// the iterface is related to the type struct AddressCache.
//
// The only method needed is GetAddress which returns the address id
type AddressResolver interface {
	GetAddress(address string) int32
}

// DecodedMsg holds the basic transaction data and the decoded, strongly-typed
// messages. Per-message handling (address extraction, conversion to database
// rows, type label) is driven by the codec registry rather than re-decoding into
// an untyped map, so each message type is defined in exactly one place.
type DecodedMsg struct {
	BasicData BasicTxData
	Msgs      []std.Msg
}
