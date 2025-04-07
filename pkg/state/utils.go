package state

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// CalculateTransactionHash calculates the hash of a transaction
func CalculateTransactionHash(tx Transaction) []byte {
	// Create a message that includes all transaction fields
	msg := []byte{}
	
	// Add transaction type
	msg = append(msg, byte(tx.Type))
	
	// Add from address
	msg = append(msg, tx.From[:]...)
	
	// Add to address
	msg = append(msg, tx.To[:]...)
	
	// Add amount (special handling for zero values)
	amountBytes := FormatBigInt(tx.Amount)
	msg = append(msg, amountBytes...)
	
	// Add nonce (as string to ensure consistent hash computation)
	nonceStr := FormatNonce(tx.Nonce)
	msg = append(msg, []byte(nonceStr)...)
	
	// Add data
	msg = append(msg, tx.Data...)
	
	// Add gas
	gasBytes := big.NewInt(int64(tx.Gas)).Bytes()
	msg = append(msg, gasBytes...)
	
	// Calculate hash
	return crypto.Keccak256(msg)
}

// ParseAmount parses an amount string into a big.Int
func ParseAmount(amount string) *big.Int {
	result := new(big.Int)
	_, success := result.SetString(amount, 10)
	if !success {
		return nil
	}
	return result
}

// FormatBigInt ensures consistent formatting of big.Int values
// This is critical for consistent message hash computation
func FormatBigInt(value *big.Int) []byte {
	// Special handling for zero values to ensure consistent message hash computation
	if value == nil || value.Cmp(big.NewInt(0)) == 0 {
		return []byte{0}
	}
	return value.Bytes()
}

// FormatNonce ensures consistent formatting of nonce values
// This is critical for consistent message hash computation
func FormatNonce(nonce uint64) string {
	// Convert nonce to string representation to ensure consistent message hash computation
	return big.NewInt(int64(nonce)).String()
}
