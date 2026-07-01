package decoder

import (
	"math/big"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jackc/pgx/v5/pgtype"
)

// convCtx carries the per-message context needed to build database rows.
type convCtx struct {
	txId           int64
	chainName      string
	timestamp      time.Time
	resolver       AddressResolver
	signerIds      []int32
	messageCounter int16
}

// msgCodec is the single source of truth for one transaction message type: it
// reports the addresses the message references (so they can be batch-resolved to
// ids) and builds the database row(s) once a resolver is available. A message
// type lives in exactly one registration below; nothing else enumerates types.
type msgCodec interface {
	addresses(msg std.Msg) []string
	convert(msg std.Msg, c convCtx) ([]s.Message, error)
}

type registryEntry struct {
	typeName string
	codec    msgCodec
}

// registry maps each concrete amino message type to its codec.
var registry = map[reflect.Type]registryEntry{}

// lookup finds the codec for a decoded message by its concrete type.
func lookup(msg std.Msg) (registryEntry, bool) {
	e, ok := registry[reflect.TypeOf(msg)]
	return e, ok
}

// typedCodec adapts strongly-typed address/convert funcs into a msgCodec so each
// registration is a pair of closures with the type assertion confined here.
type typedCodec[T std.Msg] struct {
	addrs func(T) []string
	conv  func(T, convCtx) ([]s.Message, error)
}

func (tc typedCodec[T]) addresses(msg std.Msg) []string { return tc.addrs(msg.(T)) }

func (tc typedCodec[T]) convert(msg std.Msg, c convCtx) ([]s.Message, error) {
	return tc.conv(msg.(T), c)
}

// register wires a message type into the registry. typeName must equal the
// schema row's TableName so GetMsgTypes and the message label stay consistent.
func register[T std.Msg](
	typeName string,
	addrs func(T) []string,
	conv func(T, convCtx) ([]s.Message, error),
) {
	var zero T
	registry[reflect.TypeOf(zero)] = registryEntry{
		typeName: typeName,
		codec:    typedCodec[T]{addrs: addrs, conv: conv},
	}
}

// coinsToAmounts converts amino coins to schema amounts. An empty input yields a
// non-nil empty slice (matching the previous behaviour, so COPY writes an empty
// array rather than NULL).
func coinsToAmounts(coins std.Coins) []s.Amount {
	amounts := make([]s.Amount, len(coins))
	for i, coin := range coins {
		amounts[i] = s.Amount{
			Amount: pgtype.Numeric{Int: big.NewInt(coin.Amount), Valid: true},
			Denom:  coin.Denom,
		}
	}
	return amounts
}

// sanitizeUTF8 replaces any invalid UTF-8 byte sequences in s with the Unicode
// replacement character.
//
// For some reason there is a possibility that some VM arguments can be an invalid UTF-8
// char so the database might not be able to insert the data.
// So for safety all of string related data in the VM messages will use this function.
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "�")
}

// sanitizeUTF8Slice applies sanitizeUTF8 to each element of a string slice.
func sanitizeUTF8Slice(strs []string) []string {
	out := make([]string, len(strs))
	for i, str := range strs {
		out[i] = sanitizeUTF8(str)
	}
	return out
}

func init() {
	register("bank_msg_send",
		func(m bank.MsgSend) []string {
			return []string{m.FromAddress.String(), m.ToAddress.String()}
		},
		func(m bank.MsgSend, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgSend{
				TxId:           c.txId,
				ChainName:      c.chainName,
				FromAddress:    c.resolver.GetAddress(m.FromAddress.String()),
				ToAddress:      c.resolver.GetAddress(m.ToAddress.String()),
				Amount:         coinsToAmounts(m.Amount),
				Signers:        c.signerIds,
				Timestamp:      c.timestamp,
				MessageCounter: c.messageCounter,
			}}, nil
		},
	)

	register("bank_msg_multi_send",
		func(m bank.MsgMultiSend) []string {
			addrs := make([]string, 0, len(m.Inputs)+len(m.Outputs))
			for _, in := range m.Inputs {
				addrs = append(addrs, in.Address.String())
			}
			for _, out := range m.Outputs {
				addrs = append(addrs, out.Address.String())
			}
			return addrs
		},
		func(m bank.MsgMultiSend, c convCtx) ([]s.Message, error) {
			rows := make([]s.Message, 0, len(m.Inputs)+len(m.Outputs))
			for _, in := range m.Inputs {
				rows = append(rows, &s.MsgMultiSend{
					TxId:           c.txId,
					Timestamp:      c.timestamp,
					ChainName:      c.chainName,
					Direction:      false,
					AddressId:      c.resolver.GetAddress(in.Address.String()),
					Coins:          coinsToAmounts(in.Coins),
					MessageCounter: c.messageCounter,
					Signers:        c.signerIds,
				})
			}
			for _, out := range m.Outputs {
				rows = append(rows, &s.MsgMultiSend{
					TxId:           c.txId,
					Timestamp:      c.timestamp,
					ChainName:      c.chainName,
					Direction:      true,
					AddressId:      c.resolver.GetAddress(out.Address.String()),
					Coins:          coinsToAmounts(out.Coins),
					MessageCounter: c.messageCounter,
					Signers:        c.signerIds,
				})
			}
			return rows, nil
		},
	)

	register("vm_msg_call",
		func(m vm.MsgCall) []string {
			return []string{m.Caller.String()}
		},
		func(m vm.MsgCall, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgCall{
				TxId:           c.txId,
				MessageCounter: c.messageCounter,
				ChainName:      c.chainName,
				Caller:         c.resolver.GetAddress(m.Caller.String()),
				Send:           coinsToAmounts(m.Send),
				PkgPath:        sanitizeUTF8(m.PkgPath),
				FuncName:       sanitizeUTF8(m.Func),
				Args:           sanitizeUTF8(strings.Join(m.Args, ",")),
				MaxDeposit:     coinsToAmounts(m.MaxDeposit),
				Signers:        c.signerIds,
				Timestamp:      c.timestamp,
			}}, nil
		},
	)

	register("vm_msg_add_package",
		func(m vm.MsgAddPackage) []string {
			return []string{m.Creator.String()}
		},
		func(m vm.MsgAddPackage, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgAddPackage{
				TxId:           c.txId,
				MessageCounter: c.messageCounter,
				ChainName:      c.chainName,
				Creator:        c.resolver.GetAddress(m.Creator.String()),
				PkgPath:        sanitizeUTF8(m.Package.Path),
				PkgName:        sanitizeUTF8(m.Package.Name),
				Send:           coinsToAmounts(m.Send),
				PkgFileNames:   sanitizeUTF8Slice(m.Package.FileNames()),
				MaxDeposit:     coinsToAmounts(m.MaxDeposit),
				Signers:        c.signerIds,
				Timestamp:      c.timestamp,
			}}, nil
		},
	)

	register("vm_msg_run",
		func(m vm.MsgRun) []string {
			return []string{m.Caller.String()}
		},
		func(m vm.MsgRun, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgRun{
				TxId:           c.txId,
				MessageCounter: c.messageCounter,
				ChainName:      c.chainName,
				Caller:         c.resolver.GetAddress(m.Caller.String()),
				PkgPath:        sanitizeUTF8(m.Package.Path),
				PkgName:        sanitizeUTF8(m.Package.Name),
				PkgFileNames:   sanitizeUTF8Slice(m.Package.FileNames()),
				Send:           coinsToAmounts(m.Send),
				MaxDeposit:     coinsToAmounts(m.MaxDeposit),
				Signers:        c.signerIds,
				Timestamp:      c.timestamp,
			}}, nil
		},
	)

	register("auth_msg_create_session",
		func(m auth.MsgCreateSession) []string {
			return []string{m.Creator.String(), m.SessionKey.Address().Bech32().String()}
		},
		func(m auth.MsgCreateSession, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgAuthCrSession{
				TxId:           c.txId,
				ChainName:      c.chainName,
				Timestamp:      c.timestamp,
				Creator:        c.resolver.GetAddress(m.Creator.String()),
				SessionKey:     c.resolver.GetAddress(m.SessionKey.Address().Bech32().String()),
				ExpiresAt:      time.Unix(m.ExpiresAt, 0).UTC(),
				SpendLimit:     coinsToAmounts(m.SpendLimit),
				SpendPeriod:    m.SpendPeriod,
				Signers:        c.signerIds,
				MessageCounter: c.messageCounter,
				AllowPaths:     m.AllowPaths,
			}}, nil
		},
	)

	register("auth_msg_revoke_session",
		func(m auth.MsgRevokeSession) []string {
			return []string{m.Creator.String(), m.SessionKey.Address().Bech32().String()}
		},
		func(m auth.MsgRevokeSession, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgAuthRvSession{
				TxId:           c.txId,
				ChainName:      c.chainName,
				Timestamp:      c.timestamp,
				Creator:        c.resolver.GetAddress(m.Creator.String()),
				SessionKey:     c.resolver.GetAddress(m.SessionKey.Address().Bech32().String()),
				Signers:        c.signerIds,
				MessageCounter: c.messageCounter,
			}}, nil
		},
	)

	register("auth_msg_revoke_all_sessions",
		func(m auth.MsgRevokeAllSessions) []string {
			return []string{m.Creator.String()}
		},
		func(m auth.MsgRevokeAllSessions, c convCtx) ([]s.Message, error) {
			return []s.Message{&s.MsgAuthRvAllSessions{
				TxId:           c.txId,
				ChainName:      c.chainName,
				Timestamp:      c.timestamp,
				Creator:        c.resolver.GetAddress(m.Creator.String()),
				Signers:        c.signerIds,
				MessageCounter: c.messageCounter,
			}}, nil
		},
	)
}
