package evm

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
)

// EVMExecutor handles EVM execution in the ZK-Rollup
type EVMExecutor struct {}

// NewEVMExecutor creates a new EVM executor
func NewEVMExecutor() *EVMExecutor {
	return &EVMExecutor{}
}

// StateDB interface for our simplified EVM to interact with the rollup state
type StateDB interface {
	GetBalance(common.Address) *big.Int
	SetBalance(common.Address, *big.Int)
	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)
	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)
	GetState(common.Address, common.Hash) common.Hash
	SetState(common.Address, common.Hash, common.Hash)
	SubBalance(common.Address, *big.Int)
	AddBalance(common.Address, *big.Int)
	ApplyChanges() // Apply all pending changes to the rollup state
}

// ExecuteContract executes a smart contract call
// This is a simplified implementation for the ZK-Rollup
func (e *EVMExecutor) ExecuteContract(
	stateDB StateDB,
	caller common.Address,
	contract common.Address,
	value *big.Int,
	gas uint64,
	input []byte,
) ([]byte, uint64, error) {
	// Check if the contract exists
	code := stateDB.GetCode(contract)
	if len(code) == 0 {
		return nil, 0, errors.New("contract not found")
	}

	// Check if caller has sufficient balance
	callerBalance := stateDB.GetBalance(caller)
	if callerBalance.Cmp(value) < 0 {
		return nil, 0, errors.New("insufficient balance")
	}

	// Handle zero values consistently as per memory requirements
	transferAmount := new(big.Int).Set(value)
	if transferAmount.Cmp(big.NewInt(0)) == 0 {
		transferAmount = big.NewInt(0)
	}

	// Transfer value from caller to contract
	stateDB.SubBalance(caller, transferAmount)
	stateDB.AddBalance(contract, transferAmount)

	// In a real implementation, we would execute the contract code here
	// For now, we'll just return a simple response
	returnData := []byte("contract executed")

	// Apply state changes
	stateDB.ApplyChanges()

	// Calculate gas used (simplified)
	gasUsed := gas / 2 // Assume half the gas was used
	remaining := gas - gasUsed

	log.Info().Str("caller", caller.Hex()).Str("contract", contract.Hex()).Msg("Contract executed")
	return returnData, remaining, nil
}

// DeployContract deploys a new smart contract
// This is a simplified implementation for the ZK-Rollup
func (e *EVMExecutor) DeployContract(
	stateDB StateDB,
	caller common.Address,
	value *big.Int,
	gas uint64,
	code []byte,
) (common.Address, uint64, error) {
	// Check if caller has sufficient balance
	callerBalance := stateDB.GetBalance(caller)
	if callerBalance.Cmp(value) < 0 {
		return common.Address{}, 0, errors.New("insufficient balance")
	}

	// Generate contract address (simplified)
	// In Ethereum, this would be based on sender and nonce
	nonce := stateDB.GetNonce(caller)
	contractAddr := crypto.CreateAddress(caller, nonce)

	// Handle zero values consistently as per memory requirements
	transferAmount := new(big.Int).Set(value)
	if transferAmount.Cmp(big.NewInt(0)) == 0 {
		transferAmount = big.NewInt(0)
	}

	// Transfer value from caller to contract
	stateDB.SubBalance(caller, transferAmount)
	stateDB.AddBalance(contractAddr, transferAmount)

	// Store contract code
	stateDB.SetCode(contractAddr, code)

	// Increment nonce
	stateDB.SetNonce(caller, nonce+1)

	// Apply state changes
	stateDB.ApplyChanges()

	// Calculate gas used (simplified)
	gasUsed := gas / 2 // Assume half the gas was used
	remaining := gas - gasUsed

	log.Info().Str("caller", caller.Hex()).Str("contract", contractAddr.Hex()).Msg("Contract deployed successfully")
	return contractAddr, remaining, nil
}

// FormatBigInt ensures consistent formatting of big.Int values
// This is critical for consistent message hash computation
func FormatBigInt(value *big.Int) []byte {
	// Special handling for zero values to ensure consistent message hash computation
	if value.Cmp(big.NewInt(0)) == 0 {
		return []byte{0}
	}
	return value.Bytes()
}

// FormatNonce ensures consistent formatting of nonce values
// This is critical for consistent message hash computation
func FormatNonce(nonce uint64) string {
	// Convert nonce to string representation to ensure consistent message hash computation
	return fmt.Sprintf("%d", nonce)
}
