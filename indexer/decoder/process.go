package decoder

import (
	"fmt"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

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
		// Auth messages
		case auth.MsgCreateSession:
			processAuthCr(&m, messages, i, messageCounter)
		case auth.MsgRevokeSession:
			processAuthRv(&m, messages, i, messageCounter)
		case auth.MsgRevokeAllSessions:
			processAuthRvAll(&m, messages, i, messageCounter)

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

func processAuthCr(
	m *auth.MsgCreateSession,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	creator := m.Creator.String()
	sessionKey := m.SessionKey.Address().Bech32().String()
	// It might set it to local time so force it to UTC just in case.
	expiresAt := time.Unix(m.ExpiresAt, 0).UTC()
	spendLimit, err := extractCoins(m.SpendLimit)
	if err != nil {
		spendLimit = []Coin{}
	}
	messages[i] = map[string]any{
		"msg_type":        "auth_msg_create_session",
		"creator":         creator,
		"session_key":     sessionKey,
		"expires_at":      expiresAt,
		"allow_paths":     m.AllowPaths,
		"spend_limit":     spendLimit,
		"spend_period":    m.SpendPeriod,
		"message_counter": messageCounter,
	}
}

func processAuthRv(
	m *auth.MsgRevokeSession,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	creator := m.Creator.String()
	sessionKeyRaw := m.SessionKey.Bytes()
	messages[i] = map[string]any{
		"msg_type":        "auth_msg_revoke_session",
		"creator":         creator,
		"session_key_raw": sessionKeyRaw,
		"message_counter": messageCounter,
	}
}

func processAuthRvAll(
	m *auth.MsgRevokeAllSessions,
	messages []map[string]any,
	i int,
	messageCounter int16,
) {
	creator := m.Creator.String()
	messages[i] = map[string]any{
		"msg_type":        "auth_msg_revoke_all_sessions",
		"creator":         creator,
		"message_counter": messageCounter,
	}
}
