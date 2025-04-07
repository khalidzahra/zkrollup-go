package evm

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"zkrollup/pkg/state"
)

// StateAdapter adapts the ZK-Rollup state to the EVM StateDB interface
type StateAdapter struct {
	rollupState *state.State
	
	// Track changes that need to be applied to the rollup state
	balanceChanges map[common.Address]*big.Int
	nonceChanges   map[common.Address]uint64
	codeChanges    map[common.Address][]byte
	storageChanges map[common.Address]map[common.Hash]common.Hash
	
	// Mutex for concurrent access
	mu sync.RWMutex
}

// NewStateAdapter creates a new state adapter
func NewStateAdapter(rollupState *state.State) *StateAdapter {
	return &StateAdapter{
		rollupState:    rollupState,
		balanceChanges: make(map[common.Address]*big.Int),
		nonceChanges:   make(map[common.Address]uint64),
		codeChanges:    make(map[common.Address][]byte),
		storageChanges: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

// Convert Ethereum address to rollup address
func (s *StateAdapter) toRollupAddress(addr common.Address) [20]byte {
	var result [20]byte
	copy(result[:], addr[:])
	return result
}

// Convert rollup address to Ethereum address
func (s *StateAdapter) toEthAddress(addr [20]byte) common.Address {
	return common.BytesToAddress(addr[:])
}

// GetBalance returns the balance of the given account
func (s *StateAdapter) GetBalance(addr common.Address) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Check if we have a pending balance change
	if balance, exists := s.balanceChanges[addr]; exists {
		return balance
	}
	
	// Get from rollup state
	rollupAddr := s.toRollupAddress(addr)
	account, err := s.rollupState.GetAccount(rollupAddr)
	if err != nil || account == nil {
		return big.NewInt(0)
	}
	
	// Return a copy to avoid modifications
	return new(big.Int).Set(account.Balance)
}

// GetNonce returns the nonce of the given account
func (s *StateAdapter) GetNonce(addr common.Address) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Check if we have a pending nonce change
	if nonce, exists := s.nonceChanges[addr]; exists {
		return nonce
	}
	
	// Get from rollup state
	rollupAddr := s.toRollupAddress(addr)
	account, err := s.rollupState.GetAccount(rollupAddr)
	if err != nil || account == nil {
		return 0
	}
	return account.Nonce
}



// GetCode returns the code of the given account
func (s *StateAdapter) GetCode(addr common.Address) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Check if we have pending code changes
	if code, exists := s.codeChanges[addr]; exists {
		return code
	}
	
	// In our current implementation, code is stored in storage
	// This could be enhanced to store code directly in the account
	return []byte{}
}



// GetState returns the value of the given key in the account's storage
func (s *StateAdapter) GetState(addr common.Address, key common.Hash) common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Check if we have pending storage changes
	if storage, exists := s.storageChanges[addr]; exists {
		if value, exists := storage[key]; exists {
			return value
		}
	}
	
	// In our current implementation, we don't have direct storage access
	// This would need to be enhanced for full EVM support
	return common.Hash{}
}

// SetBalance sets the balance of the given account
func (s *StateAdapter) SetBalance(addr common.Address, amount *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Handle zero values consistently as per memory requirements
	if amount.Cmp(big.NewInt(0)) == 0 {
		amount = big.NewInt(0)
	}
	
	// Store the balance change
	s.balanceChanges[addr] = new(big.Int).Set(amount)
}

// SetNonce sets the nonce of the given account
func (s *StateAdapter) SetNonce(addr common.Address, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Store the nonce change
	s.nonceChanges[addr] = nonce
}

// SetCode sets the code of the given account
func (s *StateAdapter) SetCode(addr common.Address, code []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Store the code change
	s.codeChanges[addr] = code
}

// SetState sets the value of the given key in the account's storage
func (s *StateAdapter) SetState(addr common.Address, key common.Hash, value common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Initialize storage map if needed
	if _, exists := s.storageChanges[addr]; !exists {
		s.storageChanges[addr] = make(map[common.Hash]common.Hash)
	}
	
	// Store the storage change
	s.storageChanges[addr][key] = value
}













// SubBalance subtracts amount from the account balance
func (s *StateAdapter) SubBalance(addr common.Address, amount *big.Int) {
	balance := s.GetBalance(addr)
	newBalance := new(big.Int).Sub(balance, amount)
	s.SetBalance(addr, newBalance)
}

// AddBalance adds amount to the account balance
func (s *StateAdapter) AddBalance(addr common.Address, amount *big.Int) {
	balance := s.GetBalance(addr)
	newBalance := new(big.Int).Add(balance, amount)
	s.SetBalance(addr, newBalance)
}

// ApplyChanges applies all pending changes to the rollup state
func (s *StateAdapter) ApplyChanges() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Apply balance changes
	for addr, balance := range s.balanceChanges {
		rollupAddr := s.toRollupAddress(addr)
		account, err := s.rollupState.GetAccount(rollupAddr)
		
		// Create account if it doesn't exist
		if err != nil || account == nil {
			account = &state.Account{
				Address: rollupAddr,
				Balance: balance,
				Nonce:   0,
			}
		} else {
			account.Balance = balance
		}
		
		s.rollupState.SetAccount(account)
	}
	
	// Apply nonce changes
	for addr, nonce := range s.nonceChanges {
		rollupAddr := s.toRollupAddress(addr)
		account, err := s.rollupState.GetAccount(rollupAddr)
		
		// Create account if it doesn't exist
		if err != nil || account == nil {
			account = &state.Account{
				Address: rollupAddr,
				Balance: big.NewInt(0),
				Nonce:   nonce,
			}
		} else {
			account.Nonce = nonce
		}
		
		s.rollupState.SetAccount(account)
	}
	
	// Apply code changes
	for addr, code := range s.codeChanges {
		rollupAddr := s.toRollupAddress(addr)
		s.rollupState.SetCode(rollupAddr, code)
	}
	
	// Apply storage changes
	for addr, storage := range s.storageChanges {
		rollupAddr := s.toRollupAddress(addr)
		
		for k, v := range storage {
			var key, value [32]byte
			copy(key[:], k.Bytes())
			copy(value[:], v.Bytes())
			s.rollupState.SetStorage(rollupAddr, key, value)
		}
	}
	
	// Clear all pending changes
	s.balanceChanges = make(map[common.Address]*big.Int)
	s.nonceChanges = make(map[common.Address]uint64)
	s.codeChanges = make(map[common.Address][]byte)
	s.storageChanges = make(map[common.Address]map[common.Hash]common.Hash)
}


