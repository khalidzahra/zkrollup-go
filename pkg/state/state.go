package state

import (
	"errors"
	"math/big"
	"sync"
)

// Error types
var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrCodeNotFound      = errors.New("code not found")
	ErrStorageNotFound   = errors.New("storage not found")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// TxType represents the type of transaction
type TxType uint8

const (
	TxTypeTransfer       TxType = 0
	TxTypeContractDeploy TxType = 1
	TxTypeContractCall   TxType = 2
)

// Transaction represents a transaction in the ZK-Rollup
type Transaction struct {
	Type      TxType
	From      [20]byte
	To        [20]byte
	Amount    *big.Int
	Nonce     uint64
	Data      []byte
	Gas       uint64
	Signature []byte
}

// Account represents an account in the ZK-Rollup
type Account struct {
	Address [20]byte
	Balance *big.Int
	Nonce   uint64
}

// Batch represents a batch of transactions in the ZK-Rollup
type Batch struct {
	BatchNumber  uint64
	Transactions []Transaction
	StateRoot    [32]byte
	Timestamp    uint64
	Proof        []byte // ZK proof data
}

// State represents the state of the ZK-Rollup
type State struct {
	accounts    map[[20]byte]*Account
	code        map[[20]byte][]byte
	storage     map[[20]byte]map[[32]byte][32]byte
	batches     []Batch
	batchNumber uint64
	mu          sync.RWMutex
}

// NewState creates a new state
func NewState() *State {
	return &State{
		accounts:    make(map[[20]byte]*Account),
		code:        make(map[[20]byte][]byte),
		storage:     make(map[[20]byte]map[[32]byte][32]byte),
		batches:     make([]Batch, 0),
		batchNumber: 0,
	}
}

// GetAccount retrieves an account from the state
func (s *State) GetAccount(address [20]byte) (*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[address]
	if !ok {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

// SetAccount sets an account in the state
func (s *State) SetAccount(account *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[account.Address] = account
}

// GetCode retrieves contract code from the state
func (s *State) GetCode(address [20]byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code, ok := s.code[address]
	if !ok {
		return nil, ErrCodeNotFound
	}

	return code, nil
}

// SetCode sets contract code in the state
func (s *State) SetCode(address [20]byte, code []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.code[address] = code
}

// GetStorage retrieves a storage value from the state
func (s *State) GetStorage(address [20]byte, key [32]byte) ([32]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addressStorage, ok := s.storage[address]
	if !ok {
		return [32]byte{}, ErrStorageNotFound
	}

	value, ok := addressStorage[key]
	if !ok {
		return [32]byte{}, ErrStorageNotFound
	}

	return value, nil
}

// SetStorage sets a storage value in the state
func (s *State) SetStorage(address [20]byte, key [32]byte, value [32]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.storage[address]; !ok {
		s.storage[address] = make(map[[32]byte][32]byte)
	}

	s.storage[address][key] = value
}

// GetStateRoot computes and returns the state root hash
// This is a simplified implementation for the ZK-Rollup
func (s *State) GetStateRoot() [32]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// In a real implementation, this would compute a Merkle root of the state
	// For now, we'll use a simple hash of account addresses and balances
	var stateRoot [32]byte

	// Create a simple hash based on the number of accounts and their balances
	// This is just a placeholder - a real implementation would use a proper Merkle tree
	for addr, acc := range s.accounts {
		// XOR the address bytes with the state root
		for i, b := range addr {
			stateRoot[i%32] ^= b
		}

		// If the account has a balance, incorporate it into the hash
		if acc.Balance != nil {
			balanceBytes := acc.Balance.Bytes()
			for i, b := range balanceBytes {
				stateRoot[(i+16)%32] ^= b
			}
		}

		// Incorporate nonce into the hash
		stateRoot[acc.Nonce%32] ^= byte(acc.Nonce % 256)
	}

	return stateRoot
}

// GetBatchNumber returns the current batch number
func (s *State) GetBatchNumber() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.batchNumber
}

// AddBatch adds a batch to the state
func (s *State) AddBatch(batch *Batch) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set the batch number
	batch.BatchNumber = s.batchNumber

	// Add the batch to the list
	s.batches = append(s.batches, *batch)

	// Increment the batch number
	s.batchNumber++
}
