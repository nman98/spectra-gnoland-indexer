package decoder

import (
	"fmt"
	"math/big"
	"time"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/jackc/pgx/v5/pgtype"
)

// A converter is used to convert map data types to database-ready structs
type converter struct {
	msgMap          map[string]any
	txId            int64
	chainName       string
	addressResolver AddressResolver
	timestamp       time.Time
	signerIds       []int32
}

// convertToDbMsgSend converts a map data type directly to a database-ready MsgSend struct
func (c *converter) toMsgSend() (*s.MsgSend, error) {
	data := c.msgMap
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	fromAddress, ok := data["from_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing from_address")
	}

	toAddress, ok := data["to_address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing to_address")
	}

	// Convert amount from []Coin to s.Amount
	coinAmount, ok := data["amount"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing amount")
	}

	amount := make([]s.Amount, len(coinAmount))
	for j, amt := range coinAmount {
		bigInt := big.NewInt(amt.Amount)
		amount[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &s.MsgSend{
		TxId:           c.txId,
		ChainName:      c.chainName,
		ToAddress:      c.addressResolver.GetAddress(toAddress),
		FromAddress:    c.addressResolver.GetAddress(fromAddress),
		Amount:         amount,
		Signers:        c.signerIds,
		Timestamp:      c.timestamp,
		MessageCounter: messageCounter,
	}, nil
}

func (c *converter) toMsgMultiSend() ([]s.MsgMultiSend, error) {
	data := c.msgMap
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	input, ok := data["input"].([]bank.Input)
	if !ok {
		return nil, fmt.Errorf("missing input")
	}

	output, ok := data["output"].([]bank.Output)
	if !ok {
		return nil, fmt.Errorf("missing output")
	}

	msgMultiSend := make([]s.MsgMultiSend, 0, len(input)+len(output))

	for _, in := range input {
		coins := make([]s.Amount, len(in.Coins))
		for i, coin := range in.Coins {
			bigInt := big.NewInt(coin.Amount)
			coins[i].Amount = pgtype.Numeric{Int: bigInt, Valid: true}
			coins[i].Denom = coin.Denom
		}
		multiSend := s.MsgMultiSend{
			TxId:           c.txId,
			Timestamp:      c.timestamp,
			ChainName:      c.chainName,
			Direction:      false,
			AddressId:      c.addressResolver.GetAddress(in.Address.String()),
			Coins:          coins,
			MessageCounter: messageCounter,
			Signers:        c.signerIds,
		}
		msgMultiSend = append(msgMultiSend, multiSend)
	}

	for _, ou := range output {
		coins := make([]s.Amount, len(ou.Coins))
		for i, coin := range ou.Coins {
			bigInt := big.NewInt(coin.Amount)
			coins[i].Amount = pgtype.Numeric{Int: bigInt, Valid: true}
			coins[i].Denom = coin.Denom
		}
		multiSend := s.MsgMultiSend{
			TxId:           c.txId,
			Timestamp:      c.timestamp,
			ChainName:      c.chainName,
			Direction:      true,
			AddressId:      c.addressResolver.GetAddress(ou.Address.String()),
			Coins:          coins,
			MessageCounter: messageCounter,
			Signers:        c.signerIds,
		}
		msgMultiSend = append(msgMultiSend, multiSend)
	}
	return msgMultiSend, nil
}

// convertToDbMsgCall converts a map data type directly to a database-ready MsgCall struct
func (c *converter) toMsgCall() (*s.MsgCall, error) {
	data := c.msgMap
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	caller, ok := data["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := data["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	funcName, ok := data["func_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing func_name")
	}

	argsStr, ok := data["args"].(string)
	if !ok {
		return nil, fmt.Errorf("missing args")
	}

	// Convert send from []Coin to s.Amount
	coinSend, ok := data["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	send := make([]s.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to s.Amount
	coinMaxDeposit, ok := data["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]s.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &s.MsgCall{
		TxId:           c.txId,
		MessageCounter: messageCounter,
		ChainName:      c.chainName,
		Caller:         c.addressResolver.GetAddress(caller),
		Send:           send,
		PkgPath:        pkgPath,
		FuncName:       funcName,
		Args:           argsStr,
		MaxDeposit:     maxDeposit,
		Signers:        c.signerIds,
		Timestamp:      c.timestamp,
	}, nil
}

// convertToDbMsgAddPackage converts a map data type directly to a database-ready MsgAddPackage struct
func (c *converter) toMsgAddPackage() (*s.MsgAddPackage, error) {
	data := c.msgMap
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	creator, ok := data["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}

	pkgPath, ok := data["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := data["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	// Convert send from []Coin to s.Amount
	coinSend, ok := data["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	pkgFileNames, ok := data["pkg_file_names"].([]string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_file_names")
	}

	send := make([]s.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to s.Amount
	coinMaxDeposit, ok := data["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]s.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	return &s.MsgAddPackage{
		TxId:           c.txId,
		MessageCounter: messageCounter,
		ChainName:      c.chainName,
		Creator:        c.addressResolver.GetAddress(creator),
		PkgPath:        pkgPath,
		PkgName:        pkgName,
		Send:           send,
		PkgFileNames:   pkgFileNames,
		MaxDeposit:     maxDeposit,
		Signers:        c.signerIds,
		Timestamp:      c.timestamp,
	}, nil
}

// convertToDbMsgRun converts a map data type directly to a database-ready MsgRun struct
func (c *converter) toMsgRun() (*s.MsgRun, error) {
	data := c.msgMap
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	caller, ok := data["caller"].(string)
	if !ok {
		return nil, fmt.Errorf("missing caller")
	}

	pkgPath, ok := data["pkg_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_path")
	}

	pkgName, ok := data["pkg_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_name")
	}

	// Convert send from []Coin to s.Amount
	coinSend, ok := data["send"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing send")
	}

	send := make([]s.Amount, len(coinSend))
	for j, amt := range coinSend {
		bigInt := big.NewInt(amt.Amount)
		send[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	// Convert maxDeposit from []Coin to s.Amount
	coinMaxDeposit, ok := data["max_deposit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing max_deposit")
	}

	maxDeposit := make([]s.Amount, len(coinMaxDeposit))
	for j, amt := range coinMaxDeposit {
		bigInt := big.NewInt(amt.Amount)
		maxDeposit[j] = s.Amount{
			Amount: pgtype.Numeric{Int: bigInt, Valid: true},
			Denom:  amt.Denom,
		}
	}

	pkgFileNames, ok := data["pkg_file_names"].([]string)
	if !ok {
		return nil, fmt.Errorf("missing pkg_file_names")
	}

	return &s.MsgRun{
		TxId:           c.txId,
		MessageCounter: messageCounter,
		ChainName:      c.chainName,
		Caller:         c.addressResolver.GetAddress(caller),
		PkgPath:        pkgPath,
		PkgName:        pkgName,
		PkgFileNames:   pkgFileNames,
		Send:           send,
		MaxDeposit:     maxDeposit,
		Signers:        c.signerIds,
		Timestamp:      c.timestamp,
	}, nil
}

func (c *converter) toMsgCrSession() (*s.MsgAuthCrSession, error) {
	data := c.msgMap
	creator, ok := data["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}
	sessionKey, ok := data["session_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing session_key")
	}
	expiresAt, ok := data["expires_at"].(time.Time)
	if !ok {
		return nil, fmt.Errorf("missing expires_at")
	}

	spendLimit, ok := data["spend_limit"].([]Coin)
	if !ok {
		return nil, fmt.Errorf("missing spend_limit")
	}

	spendLimitAmount := make([]s.Amount, len(spendLimit))
	for i, coin := range spendLimit {
		amt := big.NewInt(coin.Amount)
		spendLimitAmount[i] = s.Amount{
			Amount: pgtype.Numeric{Int: amt, Valid: true},
			Denom:  coin.Denom,
		}
	}

	spendPeriod, ok := data["spend_period"].(int64)
	if !ok {
		return nil, fmt.Errorf("missing spend_period")
	}

	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	allowPaths, ok := data["allow_paths"].([]string)
	if !ok {
		return nil, fmt.Errorf("missing allow_path")
	}

	return &s.MsgAuthCrSession{
		TxId:           c.txId,
		ChainName:      c.chainName,
		Timestamp:      c.timestamp,
		Creator:        c.addressResolver.GetAddress(creator),
		SessionKey:     c.addressResolver.GetAddress(sessionKey),
		ExpiresAt:      expiresAt,
		SpendLimit:     spendLimitAmount,
		SpendPeriod:    spendPeriod,
		Signers:        c.signerIds,
		MessageCounter: messageCounter,
		AllowPaths:     allowPaths,
	}, nil
}

func (c *converter) toMsgRvSession() (*s.MsgAuthRvSession, error) {
	data := c.msgMap
	creator, ok := data["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}
	sessionKey, ok := data["session_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing session_key")
	}
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	return &s.MsgAuthRvSession{
		TxId:           c.txId,
		ChainName:      c.chainName,
		Timestamp:      c.timestamp,
		Creator:        c.addressResolver.GetAddress(creator),
		SessionKey:     c.addressResolver.GetAddress(sessionKey),
		Signers:        c.signerIds,
		MessageCounter: messageCounter,
	}, nil
}

func (c *converter) toMsgRvAllSessions() (*s.MsgAuthRvAllSessions, error) {
	data := c.msgMap
	creator, ok := data["creator"].(string)
	if !ok {
		return nil, fmt.Errorf("missing creator")
	}
	messageCounter, ok := data["message_counter"].(int16)
	if !ok {
		return nil, fmt.Errorf("missing message_counter")
	}

	return &s.MsgAuthRvAllSessions{
		TxId:           c.txId,
		ChainName:      c.chainName,
		Timestamp:      c.timestamp,
		Creator:        c.addressResolver.GetAddress(creator),
		Signers:        c.signerIds,
		MessageCounter: messageCounter,
	}, nil
}
