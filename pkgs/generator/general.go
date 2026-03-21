package generator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	eventsProto "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"google.golang.org/protobuf/proto"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// The generator is used for creating synthetic data for the integration test
//
// Due to the current state of gnoland in constant state of development,
// the program will need to make synthetic data for integration test.
//
// All of the packages are fictitious, because the network reset again and
// until I can get a better look at some of the packages this will just be like a placeholder.
//
// The code does contain some of the implementation meant for training the zstandard and protobuf experiments.
// For now this will serve for the integration purposes.

type TxEvents struct {
	Events []eventsProto.Event `json:"Events"`
}

// Data generators
type DataGenerator struct {
	rand        *rand.Rand
	cryptoGen   *CryptoGenerator
	keyPairPool []*KeyPair
}

func NewDataGenerator(size int) *DataGenerator {
	// totally random seed
	seed := time.Now().UnixNano()
	cryptoGen := NewCryptoGenerator(seed)

	return &DataGenerator{
		rand:        rand.New(rand.NewSource(seed)),
		cryptoGen:   cryptoGen,
		keyPairPool: cryptoGen.GenerateKeyPairPool(size),
	}
}

// GenerateAddress generates an authentic Gno bech32 address from real cryptographic keys
func (g *DataGenerator) GenerateAddress() string {
	return g.cryptoGen.AddressFromPool(g.keyPairPool)
}

// GenerateAuthenticAddress generates a completely new authentic address (slower)
func (g *DataGenerator) GenerateAuthenticAddress() string {
	return g.cryptoGen.GenerateAuthenticAddress()
}

// GenerateAuthenticPubKey generates a completely new authentic public key (slower)
func (g *DataGenerator) GenerateAuthenticPubKey() string {
	return g.cryptoGen.GenerateAuthenticPubKey()
}

// GeneratePubKey generates an authentic Gno bech32 public key from the pool
func (g *DataGenerator) GeneratePubKey() string {
	return g.cryptoGen.PubKeyFromPool(g.keyPairPool)
}

// GetRandomKeyPair returns a random key pair from the pool for advanced usage
func (g *DataGenerator) GetRandomKeyPair() *KeyPair {
	if len(g.keyPairPool) == 0 {
		return g.cryptoGen.GenerateKeyPair()
	}
	return g.keyPairPool[g.rand.Intn(len(g.keyPairPool))]
}

type Amount struct {
	Amount int64
	Denom  string
}

// Generate amount with ugnot suffix
func (g *DataGenerator) GenerateAmount() Amount {
	// from 0.001 to 10K GNOT because 1 GNOT = 1000000 ugnot
	amount := int64(g.rand.Intn(10000000000) + 1000)

	return Amount{
		Amount: int64(amount),
		Denom:  "ugnot",
	}
}

// Generate amount string
func (g *DataGenerator) GenerateAmountString() string {
	return fmt.Sprintf("%d ugnot", g.GenerateAmount().Amount)
}

// Generate bytes value
func (g *DataGenerator) GenerateBytesValue() string {
	bytes := g.rand.Intn(50000) + 100 // 100 to 50K bytes
	return fmt.Sprintf("%d bytes", bytes)
}

// Generate blocks hash
func (g *DataGenerator) GenerateBlockHash() string {
	data := make([]byte, 33)

	for i := range 33 {
		data[i] = byte((g.rand.Intn(1000000)*31 + i*17) % 256) // try to generate value that can be encoded to base64
	}
	result := base64.StdEncoding.EncodeToString(data)
	return result
}

// Generate package path variants
func (g *DataGenerator) GeneratePackagePath() string {
	packages := []string{
		"gno.land/r/demo/profile",
		"gno.land/r/demo/board",
		"gno.land/r/demo/users",
		"gno.land/r/gnoland/blog",
		"gno.land/r/gnoland/home",
		"gno.land/p/demo/avl",
		"gno.land/p/demo/ufmt",
		"gno.land/r/dex/trade",
	}
	return packages[g.rand.Intn(len(packages))]
}

func (g *DataGenerator) GetAllBech32Addresses() []string {
	addresses := make([]string, 0, len(g.keyPairPool))
	for _, a := range g.keyPairPool {
		addresses = append(addresses, a.AddressBech32)
	}
	return addresses
}

// Event type templates
var eventTemplates = map[string]func(*DataGenerator) eventsProto.Event{
	"ProfileFieldCreated": func(g *DataGenerator) eventsProto.Event {
		fieldTypes := []string{"StringField", "IntField", "BoolField", "AddressField"}
		displayNames := []string{"alice", "bob", "charlie", "noderunner", "validator", "testuser"}

		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "ProfileFieldCreated",
			Attributes: []*eventsProto.Attribute{
				{
					Key: "FieldType",
					Value: &eventsProto.Attribute_StringValue{
						StringValue: fieldTypes[g.rand.Intn(len(fieldTypes))]}},
				{
					Key: "DisplayName",
					Value: &eventsProto.Attribute_StringValue{
						StringValue: displayNames[g.rand.Intn(len(displayNames))] + fmt.Sprintf("%d", g.rand.Intn(1000))}},
			},
			PkgPath: proto.String("gno.land/r/demo/profile"),
		}
	},

	"StorageDeposit": func(g *DataGenerator) eventsProto.Event {
		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "StorageDeposit",
			Attributes: []*eventsProto.Attribute{
				{
					Key:   "Deposit",
					Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAmountString()}},
				{
					Key:   "Storage",
					Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateBytesValue()}},
			},
			PkgPath: proto.String(g.GeneratePackagePath()),
		}
	},

	// Not a real event but let's test having an nil pkg path and having int64 value
	"Transfer": func(g *DataGenerator) eventsProto.Event {
		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "Transfer",
			Attributes: []*eventsProto.Attribute{
				{Key: "from", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAddress()}},
				{Key: "to", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAddress()}},
				{Key: "amount", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAmountString()}},
				{Key: "block_height", Value: &eventsProto.Attribute_Int64Value{Int64Value: int64(g.rand.Intn(100000))}},
			},
			PkgPath: nil,
		}
	},

	"BoardCreated": func(g *DataGenerator) eventsProto.Event {
		boardNames := []string{"general", "development", "trading", "governance", "support"}
		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "BoardCreated",
			Attributes: []*eventsProto.Attribute{
				{Key: "BoardID", Value: &eventsProto.Attribute_Int64Value{Int64Value: int64(g.rand.Intn(10000))}},
				{Key: "Name", Value: &eventsProto.Attribute_StringValue{StringValue: boardNames[g.rand.Intn(len(boardNames))]}},
				{Key: "Creator", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAddress()}},
			},
			PkgPath: proto.String("gno.land/r/demo/board"),
		}
	},

	"PostCreated": func(g *DataGenerator) eventsProto.Event {
		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "PostCreated",
			Attributes: []*eventsProto.Attribute{
				{Key: "PostID", Value: &eventsProto.Attribute_Int64Value{Int64Value: int64(g.rand.Intn(100000))}},
				{Key: "BoardID", Value: &eventsProto.Attribute_Int64Value{Int64Value: int64(g.rand.Intn(10000))}},
				{Key: "Author", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAddress()}},
				{Key: "Title", Value: &eventsProto.Attribute_StringValue{StringValue: fmt.Sprintf("Post Title %d", g.rand.Intn(1000))}},
			},
			PkgPath: proto.String("gno.land/r/demo/board"),
		}
	},

	// Not a real event but it will be used for training or part of integration tests
	"Swap": func(g *DataGenerator) eventsProto.Event {
		return eventsProto.Event{
			AtType: "/tm.GnoEvent",
			Type:   "Swap",
			Attributes: []*eventsProto.Attribute{
				{Key: "SwapPoolID", Value: &eventsProto.Attribute_Int64Value{Int64Value: int64(g.rand.Intn(100000))}},
				{Key: "FromToken", Value: &eventsProto.Attribute_StringValue{StringValue: "ugnot"}},
				{Key: "ToToken", Value: &eventsProto.Attribute_StringValue{StringValue: "usdt"}},
				{Key: "FromAmount", Value: &eventsProto.Attribute_StringValue{StringValue: g.GenerateAmountString()}},
			},
			PkgPath: proto.String("gno.land/r/dex/swap"),
		}
	},
}

// Generate synthetic transactions for integration tests
func (g *DataGenerator) GenerateTransaction() (TxEvents, std.Tx) {
	// some transactions do not have any events like bank send, other vm msgs might have events
	// but until some real events are available use the event templates
	// declare the transaction type first
	// bank send should be maybe 40% of the time?
	// vm. MsgCall and MsgRun should be 50% (25/25) of the time?
	// vm. MsgAddPackage should be 10% of the time since it is only used to create and update smart contracts?

	var transactionType string
	randomNum := g.rand.Float32()
	switch {
	case randomNum < 0.4:
		transactionType = "bank_send"
	case randomNum > 0.4 && randomNum < 0.9:
		// generate another number to decide between msg call and msg run
		rn := g.rand.Float32()
		switch {
		case rn > 0.0 && rn <= 0.5:
			transactionType = "vm_msg_call"
		case rn > 0.5 && rn <= 1.0:
			transactionType = "vm_msg_run"
		}
	case randomNum > 0.9 && randomNum <= 1.0:
		transactionType = "vm_msg_add_package"
	default:
		// fallback to bank send
		transactionType = "bank_send"
	}

	// Random number of events (1-5 per transaction, weighted toward 1-2)
	// if the transaction type is bank send, then there will be no events
	numEvents := 0
	if transactionType != "bank_send" {
		numEvents = 1
		if g.rand.Float32() < 0.3 {
			numEvents = 2
		}
		if g.rand.Float32() < 0.1 {
			numEvents = g.rand.Intn(3) + 3 // 3-5 events (rare)
		}
	}

	events := make([]eventsProto.Event, numEvents)
	eventTypes := []string{"ProfileFieldCreated", "StorageDeposit", "Transfer", "BoardCreated", "PostCreated", "Swap"}

	for i := 0; i < numEvents; i++ {
		eventType := eventTypes[g.rand.Intn(len(eventTypes))]
		events[i] = eventTemplates[eventType](g)
	}

	// now let's create the std.Tx with all of it's components
	pubKey, err := crypto.PubKeyFromBech32(g.GeneratePubKey())
	if err != nil {
		return TxEvents{Events: events}, std.Tx{}
	}
	tx := std.Tx{
		Fee: std.Fee{
			GasFee: std.Coin{
				Amount: g.GenerateAmount().Amount,
				Denom:  g.GenerateAmount().Denom,
			},
			GasWanted: 1000000,
		},
		Memo: "test memo",
		Signatures: []std.Signature{
			{PubKey: pubKey, Signature: []byte("signature")},
		},
		Msgs: []std.Msg{g.genMsgData(transactionType)},
	}
	return TxEvents{Events: events}, tx
}

func (g *DataGenerator) genMsgData(transactionType string) std.Msg {
	switch transactionType {
	case "bank_send":
		fromAddress, err := crypto.AddressFromString(g.GenerateAddress())
		if err != nil {
			return nil
		}
		toAddress, err := crypto.AddressFromString(g.GenerateAddress())
		if err != nil {
			return nil
		}
		return bank.MsgSend{
			FromAddress: fromAddress,
			ToAddress:   toAddress,
			Amount: []std.Coin{
				{Amount: int64(g.GenerateAmount().Amount), Denom: g.GenerateAmount().Denom},
			},
		}
	case "vm_msg_call":
		caller, err := crypto.AddressFromString(g.GenerateAddress())
		if err != nil {
			return nil
		}
		return vm.MsgCall{
			Caller:  caller,
			PkgPath: g.GeneratePackagePath(),
			Func:    g.GenerateFuncName(),
			Args:    []string{g.GenerateArgs()},
			Send: []std.Coin{
				{Amount: g.GenerateAmount().Amount, Denom: g.GenerateAmount().Denom},
			},
		}
	case "vm_msg_run":
		caller, err := crypto.AddressFromString(g.GenerateAddress())
		if err != nil {
			return nil
		}
		fileNames := g.GeneratePackageFileName()
		files := make([]*std.MemFile, len(fileNames))
		for i := 0; i < len(fileNames); i++ {
			files[i] = &std.MemFile{
				Name: fileNames[i],
				Body: "content",
			}
		}
		return vm.MsgRun{
			Caller: caller,
			Package: &std.MemPackage{
				Path:  g.GeneratePackagePath(),
				Name:  g.GeneratePackageName(),
				Files: files,
			},
			Send: []std.Coin{
				{Amount: int64(g.GenerateAmount().Amount), Denom: g.GenerateAmount().Denom},
			},
			MaxDeposit: []std.Coin{
				{Amount: g.GenerateAmount().Amount, Denom: g.GenerateAmount().Denom},
			},
		}
	case "vm_msg_add_package":
		creator, err := crypto.AddressFromString(g.GenerateAddress())
		if err != nil {
			return nil
		}
		fileNames := g.GeneratePackageFileName()
		files := make([]*std.MemFile, len(fileNames))
		for i := 0; i < len(fileNames); i++ {
			files[i] = &std.MemFile{
				Name: fileNames[i],
				Body: "content",
			}
		}
		return vm.MsgAddPackage{
			Creator: creator,
			Package: &std.MemPackage{
				Path:  g.GeneratePackagePath(),
				Name:  g.GeneratePackageName(),
				Files: files,
			},
			Send: []std.Coin{
				{Amount: g.GenerateAmount().Amount, Denom: g.GenerateAmount().Denom},
			},
			MaxDeposit: []std.Coin{
				{Amount: g.GenerateAmount().Amount, Denom: g.GenerateAmount().Denom},
			},
		}
	}
	return nil
}

func (g *DataGenerator) GenerateFuncName() string {
	possibleFuncs := []string{"Swap", "Lend", "Borrow", "Repay", "Deposit", "Withdraw"}
	return possibleFuncs[g.rand.Intn(len(possibleFuncs))]
}

func (g *DataGenerator) GenerateArgs() string {
	// we will use lorem ipsum for the args
	// from what I have seen it is rare to have long arguments
	// mostly I have seen some longer ones for validator registration
	// so I think we will just generate a radnom lenght depending on the generator
	// since it is hard to know how much each transaction will have we will just generate a random length
	var length int
	var randomNum float32
	switch randomNum = g.rand.Float32(); {
	// 15% chance of having short arguments
	case randomNum <= 0.15:
		length = 12
	// 15% chance of having medium arguments
	case randomNum > 0.15 && randomNum <= 0.3:
		length = 24
	// 15% chance of having long arguments
	case randomNum > 0.3 && randomNum <= 0.45:
		length = 36
	// 30% chance of having very long arguments
	case randomNum > 0.45 && randomNum <= 0.75:
		length = 48
	// 15% chance of having very long arguments
	case randomNum > 0.75 && randomNum <= 0.9:
		length = 60
	// 9% chance of having very long arguments
	case randomNum > 0.9 && randomNum <= 0.99:
		length = 72
	// 0.9% chance of having very long arguments
	case randomNum > 0.99 && randomNum <= 0.999:
		length = 84
	// 0.1% chance of having very long arguments
	case randomNum > 0.999 && randomNum <= 1.0:
		length = 3000
	}
	loremIpsum := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
	Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, 
	quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
	Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat 
	nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia 
	deserunt mollit anim id est laborum.`
	if length > len(loremIpsum) {
		loremIpsum = strings.Repeat(loremIpsum, 10)
	}
	return loremIpsum[:length]
}

func (g *DataGenerator) GeneratePackageName() string {
	possibleNames := []string{
		"package_name", "package_name_2", "package_name_3", "package_name_4", "package_name_5",
		"package_name_6", "package_name_7", "package_name_8", "package_name_9", "package_name_10",
	}
	return possibleNames[g.rand.Intn(len(possibleNames))]
}

func (g *DataGenerator) GeneratePackageFileName() []string {
	// these are not real file names, but at least they end with .gno
	// TODO: add some real example maybe?
	possibleFileNames := []string{
		"package.gno", "main.gno", "wrapper.gno",
		"module1.gno", "module2.gno", "module3.gno", "module4.gno", "module5.gno",
		"module6.gno", "module7.gno", "module8.gno", "module9.gno", "module10.gno",
	}
	// from what I have seen it can contain multiple file names
	// so we will return a slice of file names
	numFileNames := g.rand.Intn(3) + 1
	fileNames := make([]string, numFileNames)
	for i := 0; i < numFileNames; i++ {
		fileNames[i] = possibleFileNames[g.rand.Intn(len(possibleFileNames))]
	}
	return fileNames
}

// Generate dataset for training
func GenerateTrainingDataset(numTransactions int) [][]byte {
	generator := NewDataGenerator(500)
	dataset := make([][]byte, numTransactions)

	for i := 0; i < numTransactions; i++ {
		tx, _ := generator.GenerateTransaction()
		jsonData, _ := json.Marshal(tx)
		dataset[i] = jsonData
	}

	return dataset
}
