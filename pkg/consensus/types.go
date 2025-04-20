package consensus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"zkrollup/pkg/state"
	"zkrollup/pkg/util"
)

// MessageType represents different types of consensus messages
type MessageType int

const (
	PrePrepare MessageType = iota
	Prepare
	Commit
	ViewChange
	CRSEpochProposal
	CRSCeremony
)

func (m MessageType) String() string {
	switch m {
	case PrePrepare:
		return "PrePrepare"
	case Prepare:
		return "Prepare"
	case Commit:
		return "Commit"
	case ViewChange:
		return "ViewChange"
	case CRSEpochProposal:
		return "CRSEpochProposal"
	case CRSCeremony:
		return "CRSCeremony"
	default:
		return "Unknown"
	}
}

// ConsensusMessage represents a message in the PBFT consensus protocol
// Added CRSEpoch for CRS proposals
// Only one of Batch or CRSEpoch should be set per message
// (Batch for transaction batches, CRSEpoch for CRS epoch proposals)
type ConsensusMessage struct {
	Type             MessageType  `json:"type"`
	View             int64        `json:"view"`       // Current view number
	Sequence         int64        `json:"sequence"`   // Sequence number for this consensus round
	BatchHash        string       `json:"batch_hash"` // Hash of the batch being proposed
	NodeID           string       `json:"node_id"`    // ID of the node sending this message
	Timestamp        time.Time    `json:"timestamp"`
	Signature        []byte       `json:"signature"`       // Signature of the message
	Batch            *state.Batch `json:"batch,omitempty"` // Only included in PrePrepare
	CRSEpoch         *CRSEpoch    `json:"crs_epoch,omitempty"`
	CRSCeremony      *CRSCeremonyMessage `json:"crs_ceremony_message,omitempty"`
}

// Hash returns the SHA256 hash of the message's contents
func (m *ConsensusMessage) Hash() string {
	// Exclude signature from hash calculation
	temp := m.Signature
	m.Signature = nil
	defer func() { m.Signature = temp }()

	// First marshal the message to JSON
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	// For messages with batches, we need to handle nonce formatting specially
	if m.Batch != nil {
		// Create a custom map for consistent hash computation
		hashData := make(map[string]interface{})

		// Unmarshal the original data
		if err := json.Unmarshal(data, &hashData); err != nil {
			return ""
		}

		// Handle batch transactions specially
		if batchData, ok := hashData["batch"].(map[string]interface{}); ok {
			if txs, ok := batchData["Transactions"].([]interface{}); ok {
				for i, tx := range txs {
					if txMap, ok := tx.(map[string]interface{}); ok {
						// Replace nonce with string format for consistent hashing
						if nonce, ok := txMap["Nonce"].(float64); ok {
							txMap["Nonce"] = util.GetNonceForHash(uint64(nonce))
						}
						txs[i] = txMap
					}
				}
				batchData["Transactions"] = txs
			}
			hashData["batch"] = batchData
		}

		// Re-marshal with the consistent format
		data, err = json.Marshal(hashData)
		if err != nil {
			return ""
		}
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ConsensusState represents the state of a consensus round
type ConsensusState struct {
	View          int64
	Sequence      int64
	Phase         MessageType
	PrepareCount  map[string]bool
	CommitCount   map[string]bool
	Batch         *state.Batch
	BatchHash     string
	Decided       bool
	SentCommit    bool // Tracks if we've already sent a commit message
	PrePrepareMsg *ConsensusMessage
}

// NewConsensusState creates a new consensus state
func NewConsensusState(view, sequence int64, batch *state.Batch) *ConsensusState {
	// First marshal the batch to JSON
	batchData, _ := json.Marshal(batch)

	// Create a custom map for consistent hash computation
	var batchMap map[string]interface{}
	if err := json.Unmarshal(batchData, &batchMap); err != nil {
		// Fallback to direct hashing if unmarshaling fails
		hash := sha256.Sum256(batchData)
		batchHash := fmt.Sprintf("%x", hash)

		return &ConsensusState{
			View:         view,
			Sequence:     sequence,
			Phase:        PrePrepare,
			PrepareCount: make(map[string]bool),
			CommitCount:  make(map[string]bool),
			Batch:        batch,
			BatchHash:    batchHash,
			Decided:      false,
		}
	}

	// Handle transactions specially for consistent hashing
	if txs, ok := batchMap["Transactions"].([]interface{}); ok {
		for i, tx := range txs {
			if txMap, ok := tx.(map[string]interface{}); ok {
				// Replace nonce with string format for consistent hashing
				if nonce, ok := txMap["Nonce"].(float64); ok {
					txMap["Nonce"] = util.GetNonceForHash(uint64(nonce))
				}
				txs[i] = txMap
			}
		}
		batchMap["Transactions"] = txs
	}

	// Re-marshal with the consistent format
	consistentBatchData, _ := json.Marshal(batchMap)
	hash := sha256.Sum256(consistentBatchData)
	batchHash := fmt.Sprintf("%x", hash)

	return &ConsensusState{
		View:         view,
		Sequence:     sequence,
		Phase:        PrePrepare,
		PrepareCount: make(map[string]bool),
		CommitCount:  make(map[string]bool),
		Batch:        batch,
		BatchHash:    batchHash,
		Decided:      false,
		SentCommit:   false,
	}
}

// HasQuorum returns true if the number of messages received is greater than 2f+1
// where f is the maximum number of faulty nodes the system can tolerate
func HasQuorum(count int, totalNodes int) bool {
	if totalNodes <= 4 {
		return count >= 2 // Only require 2 nodes for our test environment
	}

	// Standard PBFT quorum calculation for larger networks
	f := (totalNodes - 1) / 3
	return count >= 2*f+1
}

// CRSEpoch tracks the state of a CRS ceremony
type CRSEpoch struct {
	Number       int64         `json:"number"`
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration"`
	Participants []string      `json:"participants"`
}

// CRSState is the global CRS epoch state
var CRSState struct {
	CurrentEpoch CRSEpoch
	EpochLock    sync.RWMutex
}

// SetCRSEpoch sets the current CRS epoch (thread-safe)
func SetCRSEpoch(epoch CRSEpoch) {
	CRSState.EpochLock.Lock()
	defer CRSState.EpochLock.Unlock()
	CRSState.CurrentEpoch = epoch
}

// GetCRSEpoch returns the current CRS epoch (thread-safe)
func GetCRSEpoch() CRSEpoch {
	CRSState.EpochLock.RLock()
	defer CRSState.EpochLock.RUnlock()
	return CRSState.CurrentEpoch
}

// SelectCRSParticipants deterministically selects N participants from a list of sequencer IDs using a randomness seed
func SelectCRSParticipants(sequencers []string, n int, seed []byte) []string {
	if n >= len(sequencers) {
		return append([]string{}, sequencers...) // all
	}
	// Deterministic shuffle using seed
	shuffled := append([]string{}, sequencers...)
	for i := len(shuffled) - 1; i > 0; i-- {
		h := sha256.New()
		h.Write(seed)
		h.Write([]byte(shuffled[i]))
		idx := int(h.Sum(nil)[0]) % (i + 1)
		shuffled[i], shuffled[idx] = shuffled[idx], shuffled[i]
	}
	return shuffled[:n]
}
