package decoder

import (
	"fmt"
	"math/big"
	"time"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/jackc/pgx/v5/pgtype"
)

// NewDecodedMsg creates a new DecodedMsg struct.
//
// Parameters:
//   - encodedTx: the encoded transaction
//
// Returns:
//   - *DecodedMsg: the decoded message
//   - error: an error if the decoding fails
//
// The method will not throw an error if the decoded message is not found, it will just return nil.
func NewDecodedMsg(encodedTx string) *DecodedMsg {
	decoder := NewDecoder(encodedTx)
	basicData, messages, err := decoder.GetMessageFromStdTx()
	if err != nil {
		return nil
	}
	return &DecodedMsg{
		BasicData: basicData,
		Messages:  messages,
	}
}

// GetBasicData returns the basic data of the decoded message
//
// Returns:
//   - BasicTxData: the basic data of the decoded message
//
// The method will not throw an error if the basic data is not found, it will just return nil
func (dm *DecodedMsg) GetBasicData() BasicTxData {
	return dm.BasicData
}

// GetMessages returns the messages of the decoded message.
//
// Returns:
//   - []map[string]any: the messages of the decoded message
//
// The method will not throw an error if the messages are not found, it will just return nil.
func (dm *DecodedMsg) GetMessages() []map[string]any {
	return dm.Messages
}

// GetMsgTypes returns the message types of the decoded message.
//
// Returns:
//   - []string: the message types of the decoded message
//
// The method will not throw an error if the message types are not found, it will just return nil.
func (dm *DecodedMsg) GetMsgTypes() []string {
	msgTypes := make([]string, 0, len(dm.Messages))
	for _, message := range dm.Messages {
		msgTypes = append(msgTypes, message["msg_type"].(string))
	}
	return msgTypes
}

// GetSigners returns the signers of the decoded message.
//
// Returns:
//   - []string: the signers of the decoded message
//
// The method will not throw an error if the signers are not found, it will just return nil.
func (dm *DecodedMsg) GetSigners() []string {
	return dm.BasicData.Signers
}

// GetMemo returns the memo of the decoded message.
//
// Returns:
//   - string: the memo of the decoded message
//
// The method will not throw an error if the memo is not found, it will just return nil.
func (dm *DecodedMsg) GetMemo() string {
	return dm.BasicData.Memo
}

// GetFee returns the fee of the decoded message.
//
// Returns:
//   - dataTypes.Amount: the fee of the decoded message
//
// The method will not throw an error if the fee is not found, it will just return nil.
func (dm *DecodedMsg) GetFee() dataTypes.Amount {
	return dm.BasicData.Fee
}

// GetTotalMsgCount returns the total message count of the decoded message.
//
// Returns:
//   - int: the total message count of the decoded message
//
// The method will not throw an error if the total message count is not found, it will just return 0.
func (dm *DecodedMsg) GetTotalMsgCount() int {
	return dm.BasicData.TotalMsgCount
}

// CollectAllAddresses extracts all unique addresses from the decoded message
// This includes signers and all addresses from individual messages
func (dm *DecodedMsg) CollectAllAddresses() []string {
	addressSet := make(map[string]struct{})

	// Add signers from transaction
	for _, signer := range dm.BasicData.Signers {
		addressSet[signer] = struct{}{}
	}

	// Add addresses from each message
	for _, msgMap := range dm.Messages {
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "bank_msg_send":
			if fromAddr, ok := msgMap["from_address"].(string); ok {
				addressSet[fromAddr] = struct{}{}
			}
			if toAddr, ok := msgMap["to_address"].(string); ok {
				addressSet[toAddr] = struct{}{}
			}

		case "bank_msg_multi_send":
			if distinctAddresses, ok := msgMap["distinct_addresses"].(map[string]struct{}); ok {
				for addr := range distinctAddresses {
					addressSet[addr] = struct{}{}
				}
			}

		case "vm_msg_call":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = struct{}{}
			}

		case "vm_msg_add_package":
			if creator, ok := msgMap["creator"].(string); ok {
				addressSet[creator] = struct{}{}
			}

		case "vm_msg_run":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = struct{}{}
			}
		}
	}

	// Convert set to slice
	addresses := make([]string, 0, len(addressSet))
	for addr := range addressSet {
		addresses = append(addresses, addr)
	}

	return addresses
}

// DbMessageGroups holds database-ready message types with address IDs
type DbMessageGroups struct {
	MsgSend      []dataTypes.MsgSend
	MsgMultiSend []dataTypes.MsgMultiSend
	MsgCall      []dataTypes.MsgCall
	MsgAddPkg    []dataTypes.MsgAddPackage
	MsgRun       []dataTypes.MsgRun
}

// ConvertToDbMessages directly converts the decoded message maps to database-ready message types
// This method combines the previous two-step conversion into a single step for better performance
func (dm *DecodedMsg) ConvertToDbMessages(
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signers []string,
) (*DbMessageGroups, error) {
	// Convert signers to address IDs once
	signerIds := make([]int32, len(signers))
	for k, signer := range signers {
		signerIds[k] = addressResolver.GetAddress(signer)
	}

	dbGroups := &DbMessageGroups{
		MsgSend:   make([]dataTypes.MsgSend, 0),
		MsgCall:   make([]dataTypes.MsgCall, 0),
		MsgAddPkg: make([]dataTypes.MsgAddPackage, 0),
		MsgRun:    make([]dataTypes.MsgRun, 0),
	}

	for _, msgMap := range dm.Messages {
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid msg_type")
		}

		switch msgType {
		case "bank_msg_send":
			msg, err := dm.convertToDbMsgSend(msgMap, addressResolver, txId, chainName, timestamp, signerIds)
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_send: %w", err)
			}
			dbGroups.MsgSend = append(dbGroups.MsgSend, *msg)

		case "bank_msg_multi_send":
			msgs, err := dm.convertToDbMsgMultiSend(msgMap, addressResolver, txId, chainName, timestamp, signerIds)
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_multi_send: %w", err)
			}
			dbGroups.MsgMultiSend = append(dbGroups.MsgMultiSend, msgs...)

		case "vm_msg_call":
			msg, err := dm.convertToDbMsgCall(msgMap, addressResolver, txId, chainName, timestamp, signerIds)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_call: %w", err)
			}
			dbGroups.MsgCall = append(dbGroups.MsgCall, *msg)

		case "vm_msg_add_package":
			msg, err := dm.convertToDbMsgAddPackage(msgMap, addressResolver, txId, chainName, timestamp, signerIds)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_add_package: %w", err)
			}
			dbGroups.MsgAddPkg = append(dbGroups.MsgAddPkg, *msg)

		case "vm_msg_run":
			msg, err := dm.convertToDbMsgRun(msgMap, addressResolver, txId, chainName, timestamp, signerIds)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_run: %w", err)
			}
			dbGroups.MsgRun = append(dbGroups.MsgRun, *msg)

		default:
			return nil, fmt.Errorf("unknown message type: %s", msgType)
		}
	}

	return dbGroups, nil
}

// convertToDbMsgSend converts a map data type directly to a database-ready MsgSend struct
func (dm *DecodedMsg) convertToDbMsgSend(
	msgMap map[string]any,
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signerIds []int32,
) (*dataTypes.MsgSend, error) {
	messageCounter, ok := msgMap["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	fromAddress, ok := msgMap["from_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing from_address")
	}

	toAddress, ok := msgMap["to_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing to_address")
	}

	// Convert amount from []Coin to dataTypes.Amount
	coinAmount, ok := msgMap["amount"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing amount")
	}

	amount := make([]dataTypes.Amount, len(coinAmount))
	for j, amt := range coinAmount {
		bigInt := big.NewInt(amt.Amount)
		amount[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &dataTypes.MsgSend{
		TxId:           txId,
		ChainName:      chainName,
		FromAddress:    addressResolver.GetAddress(fromAddress),
		ToAddress:      addressResolver.GetAddress(toAddress),
		Amount:         amount,
		Signers:        signerIds,
		Timestamp:      timestamp,
		MessageCounter: messageCounter,
	}, nil
}

func (dm *DecodedMsg) convertToDbMsgMultiSend(
	msgMap map[string]any,
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signerIds []int32,
) ([]dataTypes.MsgMultiSend, error) {
	messageCounter, ok := msgMap["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	input, ok := msgMap["input"].([]bank.Input)
	if !ok {
		return nil, fmt.Errorf("missing input")
	}

	output, ok := msgMap["output"].([]bank.Output)
	if !ok {
		return nil, fmt.Errorf("missing output")
	}

	msgMultiSend := make([]dataTypes.MsgMultiSend, 0, len(input)+len(output))

	for _, in := range input {
		coins := make([]dataTypes.Amount, len(in.Coins))
		for i, coin := range in.Coins {
			bigInt := big.NewInt(coin.Amount)
			coins[i].Amount = pgtype.Numeric{Int: bigInt, Valid: true}
			coins[i].Denom = coin.Denom
		}
		multiSend := dataTypes.MsgMultiSend{
			TxId:           txId,
			Timestamp:      timestamp,
			ChainName:      chainName,
			Direction:      false,
			AddressId:      addressResolver.GetAddress(in.Address.String()),
			Coins:          coins,
			MessageCounter: messageCounter,
			Signers:        signerIds,
		}
		msgMultiSend = append(msgMultiSend, multiSend)
	}

	for _, ou := range output {
		coins := make([]dataTypes.Amount, len(ou.Coins))
		for i, coin := range ou.Coins {
			bigInt := big.NewInt(coin.Amount)
			coins[i].Amount = pgtype.Numeric{Int: bigInt, Valid: true}
			coins[i].Denom = coin.Denom
		}
		multiSend := dataTypes.MsgMultiSend{
			TxId:           txId,
			Timestamp:      timestamp,
			ChainName:      chainName,
			Direction:      true,
			AddressId:      addressResolver.GetAddress(ou.Address.String()),
			Coins:          coins,
			MessageCounter: messageCounter,
			Signers:        signerIds,
		}
		msgMultiSend = append(msgMultiSend, multiSend)
	}
	return msgMultiSend, nil
}

// convertToDbMsgCall converts a map data type directly to a database-ready MsgCall struct
func (dm *DecodedMsg) convertToDbMsgCall(
	msgMap map[string]any,
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signerIds []int32,
) (*dataTypes.MsgCall, error) {
	messageCounter, ok := msgMap["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	caller, ok := msgMap["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	funcName, ok := msgMap["func_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing func_name")
	}

	argsStr, ok := msgMap["args"].(string)
	if !ok {
		return nil, fmt.Errorf("missing args")
	}

	// Convert send from []Coin to dataTypes.Amount
	coinSend, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	send := make([]dataTypes.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to dataTypes.Amount
	coinMaxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]dataTypes.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &dataTypes.MsgCall{
		TxId:           txId,
		MessageCounter: messageCounter,
		ChainName:      chainName,
		Caller:         addressResolver.GetAddress(caller),
		Send:           send,
		PkgPath:        pkgPath,
		FuncName:       funcName,
		Args:           argsStr,
		MaxDeposit:     maxDeposit,
		Signers:        signerIds,
		Timestamp:      timestamp,
	}, nil
}

// convertToDbMsgAddPackage converts a map data type directly to a database-ready MsgAddPackage struct
func (dm *DecodedMsg) convertToDbMsgAddPackage(
	msgMap map[string]any,
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signerIds []int32,
) (*dataTypes.MsgAddPackage, error) {
	messageCounter, ok := msgMap["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	creator, ok := msgMap["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := msgMap["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	// Convert send from []Coin to dataTypes.Amount
	coinSend, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	send := make([]dataTypes.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to dataTypes.Amount
	coinMaxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]dataTypes.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &dataTypes.MsgAddPackage{
		TxId:           txId,
		MessageCounter: messageCounter,
		ChainName:      chainName,
		Creator:        addressResolver.GetAddress(creator),
		PkgPath:        pkgPath,
		PkgName:        pkgName,
		Send:           send,
		MaxDeposit:     maxDeposit,
		Signers:        signerIds,
		Timestamp:      timestamp,
	}, nil
}

// convertToDbMsgRun converts a map data type directly to a database-ready MsgRun struct
func (dm *DecodedMsg) convertToDbMsgRun(
	msgMap map[string]any,
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signerIds []int32,
) (*dataTypes.MsgRun, error) {
	messageCounter, ok := msgMap["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	caller, ok := msgMap["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := msgMap["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := msgMap["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	// Convert send from []Coin to dataTypes.Amount
	coinSend, ok := msgMap["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	send := make([]dataTypes.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to dataTypes.Amount
	coinMaxDeposit, ok := msgMap["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]dataTypes.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = dataTypes.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &dataTypes.MsgRun{
		TxId:           txId,
		MessageCounter: messageCounter,
		ChainName:      chainName,
		Caller:         addressResolver.GetAddress(caller),
		PkgPath:        pkgPath,
		PkgName:        pkgName,
		Send:           send,
		MaxDeposit:     maxDeposit,
		Signers:        signerIds,
		Timestamp:      timestamp,
	}, nil
}
