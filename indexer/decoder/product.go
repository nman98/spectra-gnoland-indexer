package decoder

import (
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
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
//   - s.Amount: the fee of the decoded message
//
// The method will not throw an error if the fee is not found, it will just return nil.
func (dm *DecodedMsg) GetFee() s.Amount {
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

// messageRow bundles a converted, database-ready message row with the metadata
// needed to populate the address_tx table, captured once at conversion time so
// neither insertion nor address derivation has to re-enumerate message types.
type messageRow struct {
	row       s.Message
	addresses *s.TxAddresses
	chainName string
	timestamp time.Time
}

// DbMessages is a type-erased collection of converted message rows. It replaces
// the former per-type field struct: adding a new message type requires no change
// here, only a new case in ConvertToDbMessages (the conversion is inherently
// per-type) plus a struct that satisfies schema.Message.
type DbMessages struct {
	rows []messageRow
}

// add records a single converted message row.
func (m *DbMessages) add(row s.Message, chainName string, timestamp time.Time) {
	m.rows = append(m.rows, messageRow{
		row:       row,
		addresses: row.GetAllAddresses(),
		chainName: chainName,
		timestamp: timestamp,
	})
}

// Len returns the number of message rows held.
func (m *DbMessages) Len() int { return len(m.rows) }

// Merge appends all rows from other into m.
func (m *DbMessages) Merge(other *DbMessages) {
	m.rows = append(m.rows, other.rows...)
}

// AddressEntry holds the data needed to populate the address_tx table for a single message.
type AddressEntry struct {
	Addresses *s.TxAddresses
	ChainName string
	Timestamp time.Time
	MsgType   string
}

// AddressEntries returns one AddressEntry per message row.
func (m *DbMessages) AddressEntries() []AddressEntry {
	entries := make([]AddressEntry, len(m.rows))
	for i, r := range m.rows {
		entries[i] = AddressEntry{r.addresses, r.chainName, r.timestamp, r.row.TableName()}
	}
	return entries
}

// InsertBatch is a homogeneous group of rows (all the same table) ready for the
// generic COPY insert path, alongside the tx ids each row belongs to for
// failure diagnostics.
type InsertBatch struct {
	Rows  []s.Insertable
	TxIds []int64
}

// InsertBatches groups the collected rows by destination table. Table order is
// the order each table was first seen, keeping the output deterministic.
func (m *DbMessages) InsertBatches() []InsertBatch {
	byTable := make(map[string]*InsertBatch)
	order := make([]string, 0)
	for _, r := range m.rows {
		table := r.row.TableName()
		batch, ok := byTable[table]
		if !ok {
			batch = &InsertBatch{}
			byTable[table] = batch
			order = append(order, table)
		}
		batch.Rows = append(batch.Rows, r.row)
		batch.TxIds = append(batch.TxIds, r.addresses.TxId)
	}
	batches := make([]InsertBatch, len(order))
	for i, table := range order {
		batches[i] = *byTable[table]
	}
	return batches
}

// ConvertToDbMessages directly converts the decoded message maps to database-ready message types
// This method combines the previous two-step conversion into a single step for better performance
func (dm *DecodedMsg) ConvertToDbMessages(
	addressResolver AddressResolver,
	txId int64,
	chainName string,
	timestamp time.Time,
	signers []string,
) (*DbMessages, error) {
	// Convert signers to address IDs once
	signerIds := make([]int32, len(signers))
	for k, signer := range signers {
		signerIds[k] = addressResolver.GetAddress(signer)
	}

	out := &DbMessages{}

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
			out.add(msg, chainName, timestamp)

		case "bank_msg_multi_send":
			msgs, err := cvt.toMsgMultiSend()
			if err != nil {
				return nil, fmt.Errorf("failed to convert bank_msg_multi_send: %w", err)
			}
			for i := range msgs {
				out.add(&msgs[i], chainName, timestamp)
			}

		case "vm_msg_call":
			msg, err := cvt.toMsgCall()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_call: %w", err)
			}
			out.add(msg, chainName, timestamp)

		case "vm_msg_add_package":
			msg, err := cvt.toMsgAddPackage()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_add_package: %w", err)
			}
			out.add(msg, chainName, timestamp)

		case "vm_msg_run":
			msg, err := cvt.toMsgRun()
			if err != nil {
				return nil, fmt.Errorf("failed to convert vm_msg_run: %w", err)
			}
			out.add(msg, chainName, timestamp)

		case "auth_msg_create_session":
			msg, err := cvt.toMsgCrSession()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_cr_session: %w", err)
			}
			out.add(msg, chainName, timestamp)

		case "auth_msg_revoke_session":
			msg, err := cvt.toMsgRvSession()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_rv_session: %w", err)
			}
			out.add(msg, chainName, timestamp)

		case "auth_msg_revoke_all_sessions":
			msg, err := cvt.toMsgRvAllSessions()
			if err != nil {
				return nil, fmt.Errorf("failed to convert auth_msg_auth_rv_all_sessions: %w", err)
			}
			out.add(msg, chainName, timestamp)

		default:
			return nil, fmt.Errorf("unknown message type: %s", msgType)
		}
	}

	return out, nil
}
