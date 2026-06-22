package decoder

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
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

func processMsgs(
	tx *std.Tx,
	messages []map[string]any,
) error {
	for i, msg := range tx.GetMsgs() {
		if i > 32767 {
			return fmt.Errorf("transaction message count exceeds maximum: %d", i)
		}
		messageCounter := int16(i)
		switch m := msg.(type) {
		case bank.MsgSend:
			processSend(&m, messages, i, messageCounter)
		case bank.MsgMultiSend:
			processMultiSend(&m, messages, i, messageCounter)
		// VM messages
		case vm.MsgCall:
			processVmCall(&m, messages, i, messageCounter)
		case vm.MsgAddPackage:
			processVmAddPkg(&m, messages, i, messageCounter)
		case vm.MsgRun:
			processVmRun(&m, messages, i, messageCounter)
		default:
			return fmt.Errorf("unknown or unsupported message type: %T", m)
		}
	}
	return nil
}

// Local function to split the amount and denom
func extractCoins(amount std.Coins) ([]Coin, error) {
	// make a string and split it by space
	coins := make([]Coin, len(amount))
	for i, coin := range amount {
		coins[i] = Coin{
			Amount: coin.Amount,
			Denom:  coin.Denom,
		}
	}
	return coins, nil
}

func processSend(
	m *bank.MsgSend,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	// amount should have something like 1000000 ugnot we just need to split it and convert it to uint64
	amount, err := extractCoins(m.Amount)
	if err != nil {
		amount = []Coin{}
	}
	messages[i] = map[string]any{
		"msg_type":        "bank_msg_send",
		"from_address":    m.FromAddress.String(),
		"to_address":      m.ToAddress.String(),
		"amount":          amount,
		"message_counter": messageCounter,
	}
}

func processMultiSend(
	m *bank.MsgMultiSend,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	distinctAddresses := make(map[string]struct{})
	for _, input := range m.Inputs {
		distinctAddresses[input.Address.String()] = struct{}{}
	}
	for _, output := range m.Outputs {
		distinctAddresses[output.Address.String()] = struct{}{}
	}
	messages[i] = map[string]any{
		"msg_type":           "bank_msg_multi_send",
		"input":              m.Inputs,
		"output":             m.Outputs,
		"message_counter":    messageCounter,
		"distinct_addresses": distinctAddresses,
	}
}

func processVmCall(
	m *vm.MsgCall,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	caller := m.Caller.String()
	send, err := extractCoins(m.Send)
	if err != nil {
		send = []Coin{}
	}
	pkgPath := m.PkgPath
	// max deposit could be empty and there is a chance it will return an error
	// so we need to handle that
	maxDeposit, err := extractCoins(m.MaxDeposit)
	if err != nil {
		maxDeposit = []Coin{}
	}
	funcName := m.Func
	// combine the args into a string
	args := strings.Join(m.Args, ",")
	messages[i] = map[string]any{
		"msg_type":        "vm_msg_call",
		"caller":          caller,
		"pkg_path":        pkgPath,
		"func_name":       funcName,
		"args":            args,
		"send":            send,
		"max_deposit":     maxDeposit,
		"message_counter": messageCounter,
	}
}

func processVmAddPkg(
	m *vm.MsgAddPackage,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	pkgPath := m.Package.Path
	pkgName := m.Package.Name
	pkgFileNames := m.Package.FileNames()
	creator := m.Creator.String()
	send, err := extractCoins(m.Send)
	if err != nil {
		send = []Coin{}
	}
	maxDeposit, err := extractCoins(m.MaxDeposit)
	if err != nil {
		maxDeposit = []Coin{}
	}
	messages[i] = map[string]any{
		"msg_type":        "vm_msg_add_package",
		"pkg_path":        pkgPath,
		"pkg_name":        pkgName,
		"pkg_file_names":  pkgFileNames,
		"creator":         creator,
		"send":            send,
		"max_deposit":     maxDeposit,
		"message_counter": messageCounter,
	}
}

func processVmRun(
	m *vm.MsgRun,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	caller := m.Caller.String()
	pkgPath := m.Package.Path
	pkgName := m.Package.Name
	pkgFileNames := m.Package.FileNames()
	send, err := extractCoins(m.Send)
	if err != nil {
		send = []Coin{}
	}
	// max deposit could be empty and there is a chance it will return an error
	// so we need to handle that
	maxDeposit, err := extractCoins(m.MaxDeposit)
	if err != nil {
		maxDeposit = []Coin{}
	}
	messages[i] = map[string]any{
		"msg_type":        "vm_msg_run",
		"caller":          caller,
		"pkg_path":        pkgPath,
		"pkg_name":        pkgName,
		"pkg_file_names":  pkgFileNames,
		"send":            send,
		"max_deposit":     maxDeposit,
		"message_counter": messageCounter,
	}
}
