package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
)

// NewTestAuth creates a TransactOpts for testing with a given key and chain ID 1337.
func NewTestAuth(t *testing.T, key *ecdsa.PrivateKey) *bind.TransactOpts {
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	require.NoError(t, err)
	auth.GasLimit = 8_000_000
	auth.Context = context.Background()
	return auth
}
