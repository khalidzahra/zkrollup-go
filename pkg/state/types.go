package state

import (
	"crypto/sha256"
	"math/big"
	"sync"
)

// Account represents a user account in the rollup
type Account struct {
	Address    [20]byte
	Balance    *big.Int
	Nonce      uint64
	PubKeyHash [32]byte
}

// Transaction represents a rollup transaction
type Transaction struct {
	From      [20]byte
	To        [20]byte
	Amount    *big.Int
	Nonce     uint64
	Signature []byte
}

// Batch represents a batch of transactions to be processed
type Batch struct {
	Transactions []Transaction
	StateRoot    [32]byte
	BatchNumber  uint64
	Timestamp    uint64
	Proof        []byte
}

// State manages the rollup state
type State struct {
	mu sync.RWMutex

	accounts    map[[20]byte]*Account
	stateTree   *MerkleTree
	batchNumber uint64
	batches     []*Batch // Store finalized batches
}

func NewState() *State {
	return &State{
		accounts:    make(map[[20]byte]*Account),
		stateTree:   NewMerkleTree(32), // 32 levels deep
		batchNumber: 0,
		batches:     make([]*Batch, 0),
	}
}

func (s *State) GetAccount(address [20]byte) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if acc, exists := s.accounts[address]; exists {
		return acc
	}
	return &Account{
		Address: address,
		Balance: big.NewInt(0),
	}
}

func (s *State) UpdateAccount(account *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[account.Address] = account

	// Update state tree
	var accountHash [32]byte
	hash := sha256.New()
	hash.Write(account.Address[:])
	hash.Write(account.Balance.Bytes())
	hash.Write(big.NewInt(int64(account.Nonce)).Bytes())
	hash.Write(account.PubKeyHash[:])
	copy(accountHash[:], hash.Sum(nil))

	// Convert address to 32 bytes
	var addressKey [32]byte
	copy(addressKey[:], account.Address[:])
	s.stateTree.Update(addressKey, accountHash)
}

func (s *State) GetBatchNumber() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.batchNumber
}

func (s *State) GetStateRoot() [32]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stateTree.GetRoot()
}

// AddBatch adds a finalized batch to the state and increments the batch number
func (s *State) AddBatch(batch *Batch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Set the batch number
	batch.BatchNumber = s.batchNumber
	
	// Add the batch to our history
	s.batches = append(s.batches, batch)
	
	// Increment the batch number for the next batch
	s.batchNumber++
}

// GetBatches returns all finalized batches
func (s *State) GetBatches() []*Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	batchesCopy := make([]*Batch, len(s.batches))
	copy(batchesCopy, s.batches)
	return batchesCopy
}
