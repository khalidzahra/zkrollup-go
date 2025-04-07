package util

import (
	"fmt"
	"math/big"
)

// FormatBigInt ensures consistent formatting of big.Int values
// Specifically handling zero values properly by returning a single byte with value 0
// instead of an empty byte array
func FormatBigInt(value *big.Int) []byte {
	if value.Sign() == 0 {
		// Return a single byte with value 0 for zero values
		fmt.Printf("Handling zero value specially: using []byte{0} instead of empty array\n")
		return []byte{0}
	}

	// For non-zero values, use the standard bytes representation
	return value.Bytes()
}

// FormatNonce ensures consistent nonce format
// The nonce must be converted to a string representation when used in the circuit
func FormatNonce(nonce uint64) string {
	fmt.Printf("Converting nonce %d to string format for consistent hash computation\n", nonce)
	return fmt.Sprintf("%d", nonce)
}

// GetNonceForHash returns a consistent string representation of the nonce for hash computation
func GetNonceForHash(nonce uint64) string {
	return fmt.Sprintf("%d", nonce)
}

// FormatAmount ensures consistent amount format for the circuit
func FormatAmount(amount *big.Int) string {
	return amount.String()
}
