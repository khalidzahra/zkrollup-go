package sequencer

import (
	"errors"
	"math/big"

	"github.com/rs/zerolog/log"
	"zkrollup/pkg/state"
)

// GetAccount retrieves an account from the state
func (s *Sequencer) GetAccount(address [20]byte) (*state.Account, error) {
	// Special handling for zero values to ensure consistent message hash computation
	if isZeroAddress(address) {
		return &state.Account{
			Address: address,
			Balance: big.NewInt(0),
			Nonce:   0,
		}, nil
	}

	account, err := s.state.GetAccount(address)
	if err != nil {
		if errors.Is(err, state.ErrAccountNotFound) {
			// Return a new empty account if not found
			return &state.Account{
				Address: address,
				Balance: big.NewInt(0),
				Nonce:   0,
			}, nil
		}
		log.Error().Err(err).Str("address", formatAddress(address)).Msg("Failed to get account")
		return nil, err
	}

	return account, nil
}

// GetCode retrieves contract code from the state
func (s *Sequencer) GetCode(address [20]byte) ([]byte, error) {
	// Special handling for zero values
	if isZeroAddress(address) {
		return []byte{}, nil
	}

	code, err := s.state.GetCode(address)
	if err != nil {
		if errors.Is(err, state.ErrCodeNotFound) {
			// Return empty code if not found
			return []byte{}, nil
		}
		log.Error().Err(err).Str("address", formatAddress(address)).Msg("Failed to get code")
		return nil, err
	}

	return code, nil
}

// GetStorage retrieves a storage value from the state
func (s *Sequencer) GetStorage(address [20]byte, key [32]byte) ([32]byte, error) {
	// Special handling for zero values
	if isZeroAddress(address) {
		return [32]byte{}, nil
	}

	value, err := s.state.GetStorage(address, key)
	if err != nil {
		if errors.Is(err, state.ErrStorageNotFound) {
			// Return zero value if not found
			return [32]byte{}, nil
		}
		log.Error().Err(err).
			Str("address", formatAddress(address)).
			Str("key", formatBytes(key[:])).
			Msg("Failed to get storage")
		return [32]byte{}, err
	}

	return value, nil
}

// Helper functions

// isZeroAddress checks if an address is the zero address
func isZeroAddress(address [20]byte) bool {
	for _, b := range address {
		if b != 0 {
			return false
		}
	}
	return true
}

// formatAddress formats an address as a hex string
func formatAddress(address [20]byte) string {
	return "0x" + formatBytes(address[:])
}

// formatBytes formats bytes as a hex string
func formatBytes(b []byte) string {
	hex := make([]byte, len(b)*2)
	for i, v := range b {
		hex[i*2] = "0123456789abcdef"[v>>4]
		hex[i*2+1] = "0123456789abcdef"[v&0x0f]
	}
	return string(hex)
}
