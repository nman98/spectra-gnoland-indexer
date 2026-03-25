package database

import (
	"errors"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/date"
	"github.com/shopspring/decimal"
)

// SortOrder controls the direction of time-ordered query results.

type SortOrder string

const (
	SortOrderDesc SortOrder = "desc"
	SortOrderAsc  SortOrder = "asc"
)

// SQL returns the uppercase SQL keyword for use in ORDER BY clauses.
func (s SortOrder) SQL() string {
	if s == SortOrderAsc {
		return "ASC"
	}
	return "DESC"
}

// BlockData represents the actual block data returned in the response body
type BlockData struct {
	Hash      string    `json:"hash" doc:"Block hash (base64 encoded)"`
	Height    uint64    `json:"height" doc:"Block height"`
	Timestamp time.Time `json:"timestamp" doc:"Block timestamp"`
	ChainID   string    `json:"chain_id" doc:"Chain identifier"`
	Txs       []string  `json:"txs" doc:"Transactions (base64 encoded)"`
	TxCounter int       `json:"tx_count" doc:"Number of transactions in the block"`
}

type Event struct {
	AtType     string      `json:"at_type" doc:"Event type"`
	Type       string      `json:"type" doc:"Event type"`
	Attributes []Attribute `json:"attributes" doc:"Event attributes"`
	PkgPath    string      `json:"pkg_path" doc:"Package path"`
}

type Attribute struct {
	Key   string `json:"key" doc:"Attribute key"`
	Value string `json:"value" doc:"Attribute value"`
}

type Amount struct {
	Amount string `json:"amount" doc:"Amount"`
	Denom  string `json:"denom" doc:"Denom"`
}

type BankSend struct {
	MessageCounter int16     `json:"message_counter" doc:"Transaction order integer, starts from 0"`
	TxHash         string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp      time.Time `json:"timestamp" doc:"Transaction timestamp"`
	FromAddress    string    `json:"from_address" doc:"From address (addresses)"`
	ToAddress      string    `json:"to_address" doc:"To address (addresses)"`
	Amount         []Amount  `json:"amount" doc:"Amount"`
	Signers        []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgCall struct {
	MessageCounter int16     `json:"message_counter" doc:"Transaction order integer, starts from 0"`
	TxHash         string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp      time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Caller         string    `json:"caller" doc:"Caller address (addresses)"`
	Send           []Amount  `json:"send" doc:"Send amount"`
	PkgPath        string    `json:"pkg_path" doc:"Package path"`
	FuncName       string    `json:"func_name" doc:"Function name"`
	Args           string    `json:"args" doc:"Arguments"`
	MaxDeposit     []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers        []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgAddPackage struct {
	MessageCounter int16     `json:"message_counter" doc:"Transaction order integer, starts from 0"`
	TxHash         string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp      time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Creator        string    `json:"creator" doc:"Creator address (addresses)"`
	PkgPath        string    `json:"pkg_path" doc:"Package path"`
	PkgName        string    `json:"pkg_name" doc:"Package name"`
	PkgFileNames   []string  `json:"pkg_file_names" doc:"Package file names"`
	Send           []Amount  `json:"send" doc:"Send amount"`
	MaxDeposit     []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers        []string  `json:"signers" doc:"Signers (addresses)"`
}

type MsgRun struct {
	MessageCounter int16     `json:"message_counter" doc:"Transaction order integer, starts from 0"`
	TxHash         string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp      time.Time `json:"timestamp" doc:"Transaction timestamp"`
	Caller         string    `json:"caller" doc:"Caller address (addresses)"`
	PkgPath        string    `json:"pkg_path" doc:"Package path"`
	PkgName        string    `json:"pkg_name" doc:"Package name"`
	PkgFileNames   []string  `json:"pkg_file_names" doc:"Package file names"`
	Send           []Amount  `json:"send" doc:"Send amount"`
	MaxDeposit     []Amount  `json:"max_deposit" doc:"Max deposit"`
	Signers        []string  `json:"signers" doc:"Signers (addresses)"`
}

type Transaction struct {
	TxHash      string    `json:"tx_hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp   time.Time `json:"timestamp" doc:"Transaction timestamp"`
	BlockHeight uint64    `json:"block_height" doc:"Block height"`
	TxEvents    []Event   `json:"tx_events" doc:"Transaction events"`
	GasUsed     uint64    `json:"gas_used" doc:"Gas used"`
	GasWanted   uint64    `json:"gas_wanted" doc:"Gas wanted"`
	Fee         Amount    `json:"fee" doc:"Fee"`
	MsgTypes    []string  `json:"msg_types" doc:"Message types"`
}

type FullTxData struct {
	TxHash             string
	Timestamp          time.Time
	BlockHeight        uint64
	TxEvents           []Event
	TxEventsCompressed []byte
	CompressionOn      bool
	GasUsed            uint64
	GasWanted          uint64
	Fee                Amount
	MsgTypes           []string
}

func (f *FullTxData) ToTransaction(decode func([]byte) (*[]Event, error)) (*Transaction, error) {
	if decode == nil {
		return nil, errors.New("decode function is nil")
	}
	tx := &Transaction{
		TxHash:      f.TxHash,
		Timestamp:   f.Timestamp,
		BlockHeight: f.BlockHeight,
		GasUsed:     f.GasUsed,
		GasWanted:   f.GasWanted,
		Fee:         f.Fee,
		MsgTypes:    f.MsgTypes,
	}
	if f.CompressionOn {
		events, err := decode(f.TxEventsCompressed)
		if err != nil {
			return nil, err
		}
		if events == nil {
			return nil, errors.New("events are nil")
		}
		tx.TxEvents = *events
	} else {
		tx.TxEvents = f.TxEvents
	}
	return tx, nil
}

type BlockSigners struct {
	BlockHeight uint64   `json:"block_height" doc:"Block height"`
	Proposer    string   `json:"proposer" doc:"Proposer (addresses)"`
	SignedVals  []string `json:"signed_vals" doc:"Signed validators (addresses)"`
}

type AddressTx struct {
	Hash      string    `json:"hash" doc:"Transaction hash (base64 encoded)"`
	Timestamp time.Time `json:"timestamp" doc:"Transaction timestamp"`
	MsgTypes  []string  `json:"msg_types" doc:"Message types"`
}

type BlockCountByDate struct {
	Date  date.Date `json:"date" doc:"Date in YYYY-MM-DD format" format:"date"`
	Count int64     `json:"count" doc:"Block count"`
}

type DailyActiveAccount struct {
	Date  date.Date `json:"date" doc:"Date in YYYY-MM-DD format" format:"date"`
	Count int64     `json:"count" doc:"Active account count"`
}

type TxCountDateRange struct {
	Date  date.Date `json:"date" doc:"Date in YYYY-MM-DD format" format:"date"`
	Count int64     `json:"count" doc:"Transaction count"`
}

type TxCountTimeRange struct {
	Time  time.Time `json:"time" doc:"Time in timestamp format" format:"date-time"`
	Count int64     `json:"count" doc:"Transaction count"`
}

type VolumeByDenomDaily map[string][]*DenomVolumeDaily
type VolumeByDenomHourly map[string][]*DenomVolumeHourly

type DenomVolumeDaily struct {
	Date   date.Date       `json:"date" doc:"Time in date format" format:"date"`
	Volume decimal.Decimal `json:"volume" doc:"Volume"`
}

type DenomVolumeHourly struct {
	Time   time.Time       `json:"time" doc:"Time in timestamp format" format:"date-time"`
	Volume decimal.Decimal `json:"volume" doc:"Volume"`
}

type ValidatorSigning struct {
	Time         *time.Time `json:"time" doc:"Time" omitempty:"true"`
	BlocksSigned int64      `json:"blocks_signed" doc:"Blocks signed"`
	BlocksMissed int64      `json:"blocks_missed" doc:"Blocks missed"`
	TotalBlocks  int64      `json:"blocks_total" doc:"Total blocks"`
	SigningRate  float64    `json:"signing_rate" doc:"Signing rate percentage"`
}

type ValidatorList struct {
	ValAddresses []string `json:"validator_addresses" doc:"List of all validator addresses"`
}
