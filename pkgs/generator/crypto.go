package generator

import (
	"encoding/hex"
	"math/rand"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// CryptoGenerator handles authentic cryptographic key and address generation
type CryptoGenerator struct {
	rand *rand.Rand
}

// NewCryptoGenerator creates a new crypto generator with a random seed
func NewCryptoGenerator(seed int64) *CryptoGenerator {
	return &CryptoGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// KeyPair represents a complete key pair with derived address
type KeyPair struct {
	PrivateKey    crypto.PrivKey
	PublicKey     crypto.PubKey
	Address       crypto.Address
	AddressBech32 string
	PubKeyBech32  string
	PrivKeyHex    string
}

// GenerateKeyPair creates a new secp256k1 key pair with all derived values
func (cg *CryptoGenerator) GenerateKeyPair() *KeyPair {
	// Generate a new private key using secp256k1
	privKey := secp256k1.GenPrivKey()

	// Derive the public key from the private key
	pubKey := privKey.PubKey()

	// Derive the address from the public key
	address := pubKey.Address()

	return &KeyPair{
		PrivateKey:    privKey,
		PublicKey:     pubKey,
		Address:       address,
		AddressBech32: crypto.AddressToBech32(address),
		PubKeyBech32:  crypto.PubKeyToBech32(pubKey),
		PrivKeyHex:    hex.EncodeToString(privKey.Bytes()),
	}
}

// GenerateAuthenticAddress creates a proper Gno bech32 address from a real key pair
func (cg *CryptoGenerator) GenerateAuthenticAddress() string {
	keyPair := cg.GenerateKeyPair()
	return keyPair.AddressBech32
}

// GenerateAuthenticPubKey creates a proper Gno bech32 public key
func (cg *CryptoGenerator) GenerateAuthenticPubKey() string {
	keyPair := cg.GenerateKeyPair()
	return keyPair.PubKeyBech32
}

// GenerateKeyPairPool creates a pool of key pairs for reuse in testing
// This is more efficient than generating new keys for every address needed
func (cg *CryptoGenerator) GenerateKeyPairPool(size int) []*KeyPair {
	pool := make([]*KeyPair, size)
	for i := range size {
		pool[i] = cg.GenerateKeyPair()
	}
	return pool
}

// AddressFromPool selects a random address from a pre-generated pool
func (cg *CryptoGenerator) AddressFromPool(pool []*KeyPair) string {
	if len(pool) == 0 {
		// Fallback to generating a new address if pool is empty
		return cg.GenerateAuthenticAddress()
	}
	return pool[cg.rand.Intn(len(pool))].AddressBech32
}

// PubKeyFromPool selects a random public key from a pre-generated pool
func (cg *CryptoGenerator) PubKeyFromPool(pool []*KeyPair) string {
	if len(pool) == 0 {
		// Fallback to generating a new pubkey if pool is empty
		return cg.GenerateAuthenticPubKey()
	}
	return pool[cg.rand.Intn(len(pool))].PubKeyBech32
}

// ValidateAddress checks if an address follows proper Gno bech32 format
func ValidateAddress(address string) bool {
	// Basic validation: should start with "g1" and be 40 characters total
	if len(address) != 40 || address[:2] != "g1" {
		return false
	}

	// Try to parse it back to ensure it's valid bech32
	_, err := crypto.AddressFromBech32(address)
	return err == nil
}

// ValidatePubKey checks if a public key follows proper Gno bech32 format
func ValidatePubKey(pubkey string) bool {
	// Basic validation: should start with "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq"
	expectedPrefix := "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq"
	if len(pubkey) < len(expectedPrefix) || pubkey[:len(expectedPrefix)] != expectedPrefix {
		return false
	}

	// Try to parse it back to ensure it's valid bech32
	_, err := crypto.PubKeyFromBech32(pubkey)
	return err == nil
}

// GenerateSignature creates a dummy signature (for testing purposes)
// In real applications, you'd sign actual transaction data
func (cg *CryptoGenerator) GenerateSignature(keyPair *KeyPair, data []byte) []byte {
	// For testing purposes, we can create a dummy signature
	// In production, you would use: keyPair.PrivateKey.Sign(data)
	signature := make([]byte, 64) // secp256k1 signatures are 64 bytes
	cg.rand.Read(signature)
	return signature
}
