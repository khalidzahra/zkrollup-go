package state

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Hash computes the hash of a transaction
func (tx *Transaction) Hash() [32]byte {
	// Create a buffer to hold all transaction data
	var buffer []byte

	// Add transaction type
	buffer = append(buffer, byte(tx.Type))

	// Add from and to addresses
	buffer = append(buffer, tx.From[:]...)
	buffer = append(buffer, tx.To[:]...)

	// Add amount (convert to bytes)
	amountBytes := tx.Amount.Bytes()
	buffer = append(buffer, amountBytes...)

	// Add nonce (convert to bytes)
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, tx.Nonce)
	buffer = append(buffer, nonceBytes...)

	// Add data
	buffer = append(buffer, tx.Data...)

	// Add gas (convert to bytes)
	gasBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasBytes, tx.Gas)
	buffer = append(buffer, gasBytes...)

	// Compute hash
	hash := sha256.Sum256(buffer)
	return hash
}

// HashToBytes converts a transaction hash to bytes
func (tx *Transaction) HashToBytes() []byte {
	hash := tx.Hash()
	return hash[:]
}

// HashToEthHash converts a transaction hash to Ethereum hash format
func (tx *Transaction) HashToEthHash() common.Hash {
	hash := tx.Hash()
	return common.BytesToHash(hash[:])
}

// SignTransaction signs a transaction with the given private key
func SignTransaction(tx *Transaction, privateKey []byte) ([]byte, error) {
	// Compute the hash of the transaction
	hash := tx.Hash()

	// Parse private key
	privKey, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	// Sign the hash using the private key
	signature, err := crypto.Sign(hash[:], privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// The signature should be 65 bytes (r, s, v)
	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	return signature, nil
}
