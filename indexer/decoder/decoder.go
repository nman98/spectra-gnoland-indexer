package decoder

import (
	"crypto/sha256"
	"encoding/base64"
	"math/big"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jackc/pgx/v5/pgtype"
)

// Decoder is a struct that contains the encoded transaction and the decoded transaction
// It is used to decode the transaction and get the data from it
type Decoder struct {
	encodedTx string
}

// NewDecoder creates a new Decoder struct
//
// Parameters:
//   - encodedTx: base64 encoded stdTx
//
// Returns:
//   - *Decoder: new Decoder struct
func NewDecoder(encodedTx string) *Decoder {
	return &Decoder{
		encodedTx: encodedTx,
	}
}

// DecodeStdTxFromBase64 decodes a base64 encoded stdTx
//
// The function decodes the data and unmarshal it into a std.Tx struct
// The struct contains the transaction data and the messages
//
// Parameters:
//   - s: base64 encoded stdTx
//
// Returns:
//   - *std.Tx: decoded stdTx
//   - error: if the base64 decoding or unmarshaling fails
func (d *Decoder) DecodeStdTxFromBase64() (*std.Tx, error) {
	bz, err := base64.StdEncoding.DecodeString(d.encodedTx)
	if err != nil {
		return nil, err
	}
	var tx std.Tx
	if err := amino.Unmarshal(bz, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetMessageFromStdTx is a method that decodes the transaction and returns the appropriate basic tx data and messages
//
// # The use case of this function is to decode the raw tx data and gather information about the transaction
//
// Parameters:
//   - none
//
// Returns:
//   - BasicTxData: basic tx data
//   - []map[string]any: messages data in a map
//   - error: if the decoding or unmarshaling fails
func (d *Decoder) GetMessageFromStdTx() (BasicTxData, []map[string]any, error) {
	tx, err := d.DecodeStdTxFromBase64()
	if err != nil {
		return BasicTxData{}, nil, err
	}

	// Get transaction hash
	bz, err := base64.StdEncoding.DecodeString(d.encodedTx)
	if err != nil {
		return BasicTxData{}, nil, err
	}

	// Use sha256 and then we will use the hash as the primary key for the transaction
	txHash := sha256.Sum256(bz)

	signers := tx.GetSigners()
	signersString := make([]string, len(signers))
	for i, signer := range signers {
		signersString[i] = signer.String()
	}
	bigInt := big.NewInt(tx.Fee.GasFee.Amount)
	feeAmount := pgtype.Numeric{Int: bigInt, Valid: true}
	fee := dataTypes.Amount{
		Amount: feeAmount,
		Denom:  tx.Fee.GasFee.Denom,
	}

	msgCount := len(tx.GetMsgs())

	basicTxData := BasicTxData{
		TxHash:        txHash[:],
		Signers:       signersString,
		Memo:          tx.GetMemo(),
		Fee:           fee,
		TotalMsgCount: msgCount,
	}

	var messages = make([]map[string]any, msgCount)

	// Process each message in the transaction
	err = processMsgs(tx, messages)
	if err != nil {
		return BasicTxData{}, nil, err
	}
	return basicTxData, messages, nil
}
