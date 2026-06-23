package decoder

import (
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
)

var l = logger.Get()

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
		// It should never happen, but we need some way to halt the program.
		l.Fatal().Stack().Msgf("failed to decode message %s", err)
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

		case "auth_msg_create_session":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = struct{}{}
			}
			if sessionKey, ok := msgMap["session_key"].(string); ok {
				addressSet[sessionKey] = struct{}{}
			}

		case "auth_msg_revoke_session":
			if caller, ok := msgMap["caller"].(string); ok {
				addressSet[caller] = struct{}{}
			}
			if sessionKey, ok := msgMap["session_key"].(string); ok {
				addressSet[sessionKey] = struct{}{}
			}

		case "auth_msg_revoke_all_sessions":
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
	MsgSend              []dataTypes.MsgSend
	MsgMultiSend         []dataTypes.MsgMultiSend
	MsgCall              []dataTypes.MsgCall
	MsgAddPkg            []dataTypes.MsgAddPackage
	MsgRun               []dataTypes.MsgRun
	MsgAuthCrSession     []dataTypes.MsgAuthCrSession
	MsgAuthRvSession     []dataTypes.MsgAuthRvSession
	MsgAuthRvAllSessions []dataTypes.MsgAuthRvAllSessions
}

// Merge appends all message slices from other into g.
func (g *DbMessageGroups) Merge(other *DbMessageGroups) {
	g.MsgSend = append(g.MsgSend, other.MsgSend...)
	g.MsgMultiSend = append(g.MsgMultiSend, other.MsgMultiSend...)
	g.MsgCall = append(g.MsgCall, other.MsgCall...)
	g.MsgAddPkg = append(g.MsgAddPkg, other.MsgAddPkg...)
	g.MsgRun = append(g.MsgRun, other.MsgRun...)
	g.MsgAuthCrSession = append(g.MsgAuthCrSession, other.MsgAuthCrSession...)
	g.MsgAuthRvSession = append(g.MsgAuthRvSession, other.MsgAuthRvSession...)
	g.MsgAuthRvAllSessions = append(g.MsgAuthRvAllSessions, other.MsgAuthRvAllSessions...)
}

// AddressEntry holds the data needed to populate the address_tx table for a single message.
type AddressEntry struct {
	Addresses *dataTypes.TxAddresses
	ChainName string
	Timestamp time.Time
	MsgType   string
}

// AllAddressEntries returns one AddressEntry per message across all groups.
func (g *DbMessageGroups) AllAddressEntries() []AddressEntry {
	entries := make([]AddressEntry, 0)
	for _, m := range g.MsgSend {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgMultiSend {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgCall {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgAddPkg {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgRun {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgAuthCrSession {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgAuthRvSession {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	for _, m := range g.MsgAuthRvAllSessions {
		entries = append(entries, AddressEntry{m.GetAllAddresses(), m.ChainName, m.Timestamp, m.TableName()})
	}
	return entries
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
		MsgSend:              make([]dataTypes.MsgSend, 0),
		MsgMultiSend:         make([]dataTypes.MsgMultiSend, 0),
		MsgCall:              make([]dataTypes.MsgCall, 0),
		MsgAddPkg:            make([]dataTypes.MsgAddPackage, 0),
		MsgRun:               make([]dataTypes.MsgRun, 0),
		MsgAuthCrSession:     make([]dataTypes.MsgAuthCrSession, 0),
		MsgAuthRvSession:     make([]dataTypes.MsgAuthRvSession, 0),
		MsgAuthRvAllSessions: make([]dataTypes.MsgAuthRvAllSessions, 0),
	}

	cvt := converter{
		msgMap:          nil,
		txId:            txId,
		chainName:       chainName,
		addressResolver: addressResolver,
		timestamp:       timestamp,
		signerIds:       signerIds,
	}
	for _, msgMap := range dm.Messages {
		cvt.msgMap = msgMap
		msgType, ok := msgMap["msg_type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid msg_type")
		}

		switch msgType {
		case "bank_msg_send":
			msg, err := cvt.toMsgSend()
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_send: %w", err)
			}
			dbGroups.MsgSend = append(dbGroups.MsgSend, *msg)

		case "bank_msg_multi_send":
			msgs, err := cvt.toMsgMultiSend()
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_multi_send: %w", err)
			}
			dbGroups.MsgMultiSend = append(dbGroups.MsgMultiSend, msgs...)

		case "vm_msg_call":
			msg, err := cvt.toMsgCall()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_call: %w", err)
			}
			dbGroups.MsgCall = append(dbGroups.MsgCall, *msg)

		case "vm_msg_add_package":
			msg, err := cvt.toMsgAddPackage()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_add_package: %w", err)
			}
			dbGroups.MsgAddPkg = append(dbGroups.MsgAddPkg, *msg)

		case "vm_msg_run":
			msg, err := cvt.toMsgRun()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_run: %w", err)
			}
			dbGroups.MsgRun = append(dbGroups.MsgRun, *msg)

		case "auth_msg_create_session":
			msg, err := cvt.toMsgCrSession()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_cr_session: %w", err)
			}
			dbGroups.MsgAuthCrSession = append(dbGroups.MsgAuthCrSession, *msg)

		case "auth_msg_revoke_session":
			msg, err := cvt.toMsgRvSession()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_rv_session: %w", err)
			}
			dbGroups.MsgAuthRvSession = append(dbGroups.MsgAuthRvSession, *msg)

		case "auth_msg_revoke_all_sessions":
			msg, err := cvt.toMsgRvAllSessions()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_rv_all_sessions: %w", err)
			}
			dbGroups.MsgAuthRvAllSessions = append(dbGroups.MsgAuthRvAllSessions, *msg)

		default:
			return nil, fmt.Errorf("unknown message type: %s", msgType)
		}
	}

	return dbGroups, nil
}
