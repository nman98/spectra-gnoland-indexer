package decoder

import (
	"encoding/base64"
	"sort"
	"testing"
	"time"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// mockResolver assigns a deterministic, stable id to each distinct address.
type mockResolver struct {
	ids  map[string]int32
	next int32
}

func newMockResolver() *mockResolver { return &mockResolver{ids: make(map[string]int32)} }

func (m *mockResolver) GetAddress(address string) int32 {
	if id, ok := m.ids[address]; ok {
		return id
	}
	m.next++
	m.ids[address] = m.next
	return m.next
}

// testAddr builds a distinct, deterministic address from a seed byte.
func testAddr(seed byte) crypto.Address {
	var a crypto.Address
	for i := range a {
		a[i] = seed
	}
	return a
}

func coins(amount int64, denom string) std.Coins {
	return std.Coins{std.Coin{Denom: denom, Amount: amount}}
}

func sessionPubKey() crypto.PubKey {
	return ed25519.GenPrivKeyFromSecret([]byte("registry-test-session-key")).PubKey()
}

func memPkg() *std.MemPackage {
	return &std.MemPackage{
		Name:  "foo",
		Path:  "gno.land/r/demo/foo",
		Files: []*std.MemFile{{Name: "foo.gno", Body: "package foo"}},
	}
}

// sampleMessages returns one message of every supported type, with known
// addresses so conversions and address extraction can be asserted.
func sampleMessages(sk crypto.PubKey) []std.Msg {
	return []std.Msg{
		bank.MsgSend{FromAddress: testAddr(1), ToAddress: testAddr(2), Amount: coins(100, "ugnot")},
		bank.MsgMultiSend{
			Inputs:  []bank.Input{{Address: testAddr(3), Coins: coins(50, "ugnot")}},
			Outputs: []bank.Output{{Address: testAddr(4), Coins: coins(30, "ugnot")}, {Address: testAddr(5), Coins: coins(20, "ugnot")}},
		},
		vm.MsgCall{Caller: testAddr(6), Send: coins(5, "ugnot"), PkgPath: "gno.land/r/demo/foo", Func: "Bar", Args: []string{"a", "b"}},
		vm.MsgAddPackage{Creator: testAddr(7), Package: memPkg(), Send: std.Coins{}, MaxDeposit: std.Coins{}},
		vm.MsgRun{Caller: testAddr(8), Package: memPkg(), Send: std.Coins{}, MaxDeposit: std.Coins{}},
		auth.MsgCreateSession{Creator: testAddr(9), SessionKey: sk, ExpiresAt: 1700000000, AllowPaths: []string{"*"}, SpendLimit: coins(1000, "ugnot"), SpendPeriod: 3600},
		auth.MsgRevokeSession{Creator: testAddr(10), SessionKey: sk},
		auth.MsgRevokeAllSessions{Creator: testAddr(11)},
	}
}

// expectedTypeNames is the message label of each entry in sampleMessages, in order.
var expectedTypeNames = []string{
	"bank_msg_send",
	"bank_msg_multi_send",
	"vm_msg_call",
	"vm_msg_add_package",
	"vm_msg_run",
	"auth_msg_create_session",
	"auth_msg_revoke_session",
	"auth_msg_revoke_all_sessions",
}

func TestRegistryCoversAllTypes(t *testing.T) {
	if len(registry) != len(expectedTypeNames) {
		t.Fatalf("registry has %d entries, expected %d", len(registry), len(expectedTypeNames))
	}
	for _, msg := range sampleMessages(sessionPubKey()) {
		if _, ok := lookup(msg); !ok {
			t.Errorf("message type %T is not registered", msg)
		}
	}
}

// TestDecodeRoundTrip marshals a transaction with amino and decodes it back
// through the real base64 path. This proves the registry keys match the concrete
// types amino produces (value, not pointer) and that decode-time validation
// accepts supported types.
//
// Only amino-registered message types are exercised here. bank.MsgMultiSend is
// deliberately excluded: it is not registered with amino in this gno version, so
// it can never appear in a real decoded transaction. Its conversion logic is
// still covered by the white-box tests below.
func TestDecodeRoundTrip(t *testing.T) {
	sk := sessionPubKey()
	roundTripMsgs := []std.Msg{
		bank.MsgSend{FromAddress: testAddr(1), ToAddress: testAddr(2), Amount: coins(100, "ugnot")},
		vm.MsgCall{Caller: testAddr(6), Send: coins(5, "ugnot"), PkgPath: "gno.land/r/demo/foo", Func: "Bar", Args: []string{"a", "b"}},
		vm.MsgAddPackage{Creator: testAddr(7), Package: memPkg(), Send: std.Coins{}, MaxDeposit: std.Coins{}},
		vm.MsgRun{Caller: testAddr(8), Package: memPkg(), Send: std.Coins{}, MaxDeposit: std.Coins{}},
		auth.MsgCreateSession{Creator: testAddr(9), SessionKey: sk, ExpiresAt: 1700000000, AllowPaths: []string{"*"}, SpendLimit: coins(1000, "ugnot"), SpendPeriod: 3600},
		auth.MsgRevokeSession{Creator: testAddr(10), SessionKey: sk},
		auth.MsgRevokeAllSessions{Creator: testAddr(11)},
	}
	wantTypes := []string{
		"bank_msg_send",
		"vm_msg_call",
		"vm_msg_add_package",
		"vm_msg_run",
		"auth_msg_create_session",
		"auth_msg_revoke_session",
		"auth_msg_revoke_all_sessions",
	}

	tx := std.Tx{
		Msgs: roundTripMsgs,
		Fee:  std.Fee{GasWanted: 1000, GasFee: std.Coin{Denom: "ugnot", Amount: 200}},
		Memo: "round trip",
	}
	bz, err := amino.Marshal(tx)
	if err != nil {
		t.Fatalf("amino marshal: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(bz)

	basic, msgs, err := NewDecoder(encoded).GetMessageFromStdTx()
	if err != nil {
		t.Fatalf("GetMessageFromStdTx: %v", err)
	}
	if len(msgs) != len(wantTypes) {
		t.Fatalf("decoded %d messages, expected %d", len(msgs), len(wantTypes))
	}
	if basic.TotalMsgCount != len(wantTypes) {
		t.Errorf("TotalMsgCount = %d, expected %d", basic.TotalMsgCount, len(wantTypes))
	}
	if basic.Fee.Denom != "ugnot" || basic.Fee.Amount.Int.Int64() != 200 {
		t.Errorf("fee = %d %s, expected 200 ugnot", basic.Fee.Amount.Int.Int64(), basic.Fee.Denom)
	}
	if basic.Memo != "round trip" {
		t.Errorf("memo = %q, expected %q", basic.Memo, "round trip")
	}

	dm := &DecodedMsg{BasicData: basic, Msgs: msgs}
	gotTypes := dm.GetMsgTypes()
	if len(gotTypes) != len(wantTypes) {
		t.Fatalf("GetMsgTypes returned %d, expected %d", len(gotTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if gotTypes[i] != want {
			t.Errorf("msg type[%d] = %q, expected %q", i, gotTypes[i], want)
		}
	}

	// The decoded messages must convert end-to-end without error.
	if _, err := dm.ConvertToDbMessages(newMockResolver(), 1, "c", time.Unix(0, 0).UTC(), basic.Signers); err != nil {
		t.Fatalf("ConvertToDbMessages on decoded tx: %v", err)
	}
}

// TestConvertToDbMessages checks that every message type converts to the right
// table with the expected row count (multi-send fans out to one row per
// input/output).
func TestConvertToDbMessages(t *testing.T) {
	dm := &DecodedMsg{Msgs: sampleMessages(sessionPubKey())}
	out, err := dm.ConvertToDbMessages(newMockResolver(), 42, "test-chain", time.Unix(1700000000, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("ConvertToDbMessages: %v", err)
	}

	counts := make(map[string]int)
	for _, batch := range out.InsertBatches() {
		counts[batch.Rows[0].TableName()] += len(batch.Rows)
	}

	want := map[string]int{
		"bank_msg_send":                1,
		"bank_msg_multi_send":          3,
		"vm_msg_call":                  1,
		"vm_msg_add_package":           1,
		"vm_msg_run":                   1,
		"auth_msg_create_session":      1,
		"auth_msg_revoke_session":      1,
		"auth_msg_revoke_all_sessions": 1,
	}
	if len(counts) != len(want) {
		t.Fatalf("got %d tables, expected %d: %v", len(counts), len(want), counts)
	}
	for table, n := range want {
		if counts[table] != n {
			t.Errorf("table %q row count = %d, expected %d", table, counts[table], n)
		}
	}
}

func TestMsgSendConversion(t *testing.T) {
	from, to := testAddr(1), testAddr(2)
	resolver := newMockResolver()
	dm := &DecodedMsg{Msgs: []std.Msg{
		bank.MsgSend{FromAddress: from, ToAddress: to, Amount: coins(100, "ugnot")},
	}}

	out, err := dm.ConvertToDbMessages(resolver, 7, "test-chain", time.Unix(123, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	rows := out.InsertBatches()[0].Rows
	row, ok := rows[0].(*s.MsgSend)
	if !ok {
		t.Fatalf("row is %T, expected *schema.MsgSend", rows[0])
	}
	if row.FromAddress != resolver.GetAddress(from.String()) {
		t.Errorf("FromAddress = %d, expected resolved id %d", row.FromAddress, resolver.GetAddress(from.String()))
	}
	if row.ToAddress != resolver.GetAddress(to.String()) {
		t.Errorf("ToAddress = %d, expected resolved id %d", row.ToAddress, resolver.GetAddress(to.String()))
	}
	if row.TxId != 7 || row.ChainName != "test-chain" {
		t.Errorf("TxId/ChainName = %d/%q, expected 7/test-chain", row.TxId, row.ChainName)
	}
	if len(row.Amount) != 1 || row.Amount[0].Denom != "ugnot" || row.Amount[0].Amount.Int.Int64() != 100 {
		t.Errorf("Amount = %+v, expected 100 ugnot", row.Amount)
	}
}

func TestMultiSendDirections(t *testing.T) {
	dm := &DecodedMsg{Msgs: []std.Msg{
		bank.MsgMultiSend{
			Inputs:  []bank.Input{{Address: testAddr(3), Coins: coins(50, "ugnot")}},
			Outputs: []bank.Output{{Address: testAddr(4), Coins: coins(30, "ugnot")}, {Address: testAddr(5), Coins: coins(20, "ugnot")}},
		},
	}}

	out, err := dm.ConvertToDbMessages(newMockResolver(), 1, "c", time.Unix(0, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	rows := out.InsertBatches()[0].Rows
	if len(rows) != 3 {
		t.Fatalf("got %d rows, expected 3 (1 input + 2 outputs)", len(rows))
	}
	var inputs, outputs int
	for _, r := range rows {
		ms := r.(*s.MsgMultiSend)
		if ms.Direction {
			outputs++
		} else {
			inputs++
		}
	}
	if inputs != 1 || outputs != 2 {
		t.Errorf("got %d inputs / %d outputs, expected 1 / 2", inputs, outputs)
	}
}

func TestCreateSessionConversion(t *testing.T) {
	sk := sessionPubKey()
	creator := testAddr(9)
	resolver := newMockResolver()
	dm := &DecodedMsg{Msgs: []std.Msg{
		auth.MsgCreateSession{
			Creator: creator, SessionKey: sk, ExpiresAt: 1700000000,
			AllowPaths: []string{"*"}, SpendLimit: coins(1000, "ugnot"), SpendPeriod: 3600,
		},
	}}

	out, err := dm.ConvertToDbMessages(resolver, 1, "c", time.Unix(0, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	row := out.InsertBatches()[0].Rows[0].(*s.MsgAuthCrSession)

	if row.Creator != resolver.GetAddress(creator.String()) {
		t.Errorf("creator id = %d, expected %d", row.Creator, resolver.GetAddress(creator.String()))
	}
	wantSessionID := resolver.GetAddress(sk.Address().Bech32().String())
	if row.SessionKey != wantSessionID {
		t.Errorf("session key id = %d, expected %d", row.SessionKey, wantSessionID)
	}
	if row.Creator == row.SessionKey {
		t.Error("creator and session key resolved to the same id; they are distinct addresses")
	}
	if !row.ExpiresAt.Equal(time.Unix(1700000000, 0).UTC()) {
		t.Errorf("ExpiresAt = %v, expected %v", row.ExpiresAt, time.Unix(1700000000, 0).UTC())
	}
	if row.SpendPeriod != 3600 {
		t.Errorf("SpendPeriod = %d, expected 3600", row.SpendPeriod)
	}
	if len(row.AllowPaths) != 1 || row.AllowPaths[0] != "*" {
		t.Errorf("AllowPaths = %v, expected [*]", row.AllowPaths)
	}
}

func TestMessageCounterMatchesIndex(t *testing.T) {
	// Three single-row messages; each row's counter must equal its position.
	dm := &DecodedMsg{Msgs: []std.Msg{
		bank.MsgSend{FromAddress: testAddr(1), ToAddress: testAddr(2), Amount: coins(1, "ugnot")},
		vm.MsgCall{Caller: testAddr(6), Send: std.Coins{}, MaxDeposit: std.Coins{}, PkgPath: "p", Func: "F"},
		auth.MsgRevokeAllSessions{Creator: testAddr(11)},
	}}

	out, err := dm.ConvertToDbMessages(newMockResolver(), 1, "c", time.Unix(0, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	counters := map[string]int16{}
	for _, batch := range out.InsertBatches() {
		switch row := batch.Rows[0].(type) {
		case *s.MsgSend:
			counters["send"] = row.MessageCounter
		case *s.MsgCall:
			counters["call"] = row.MessageCounter
		case *s.MsgAuthRvAllSessions:
			counters["revoke_all"] = row.MessageCounter
		}
	}
	if counters["send"] != 0 || counters["call"] != 1 || counters["revoke_all"] != 2 {
		t.Errorf("message counters = %v, expected send=0 call=1 revoke_all=2", counters)
	}
}

func TestCollectAllAddresses(t *testing.T) {
	sk := sessionPubKey()
	signer := testAddr(100).String()
	dm := &DecodedMsg{
		BasicData: BasicTxData{Signers: []string{signer}},
		Msgs: []std.Msg{
			bank.MsgSend{FromAddress: testAddr(1), ToAddress: testAddr(2), Amount: coins(1, "ugnot")},
			auth.MsgCreateSession{Creator: testAddr(9), SessionKey: sk, ExpiresAt: 1},
		},
	}

	got := dm.CollectAllAddresses()
	want := []string{
		signer,
		testAddr(1).String(),
		testAddr(2).String(),
		testAddr(9).String(),
		sk.Address().Bech32().String(),
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("collected %d addresses, expected %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("address[%d] = %q, expected %q", i, got[i], want[i])
		}
	}
}

// unknownMsg is a std.Msg that is not registered with any codec.
type unknownMsg struct{}

func (unknownMsg) Route() string                { return "x" }
func (unknownMsg) Type() string                 { return "x" }
func (unknownMsg) ValidateBasic() error         { return nil }
func (unknownMsg) GetSignBytes() []byte         { return nil }
func (unknownMsg) GetSigners() []crypto.Address { return nil }

func TestUnknownMessageTypeRejected(t *testing.T) {
	if _, ok := lookup(unknownMsg{}); ok {
		t.Fatal("lookup unexpectedly found a codec for an unregistered type")
	}

	dm := &DecodedMsg{Msgs: []std.Msg{unknownMsg{}}}
	if _, err := dm.ConvertToDbMessages(newMockResolver(), 1, "c", time.Unix(0, 0).UTC(), nil); err == nil {
		t.Error("ConvertToDbMessages accepted an unregistered message type, expected error")
	}
}
