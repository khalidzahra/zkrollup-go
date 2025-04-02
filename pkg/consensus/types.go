package consensus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"zkrollup/pkg/state"
)

// MessageType represents different types of consensus messages
type MessageType int

const (
	PrePrepare MessageType = iota
	Prepare
	Commit
	ViewChange
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
	default:
		return "Unknown"
	}
}

// ConsensusMessage represents a message in the PBFT consensus protocol
type ConsensusMessage struct {
	Type      MessageType  `json:"type"`
	View      int64        `json:"view"`       // Current view number
	Sequence  int64        `json:"sequence"`   // Sequence number for this consensus round
	BatchHash string       `json:"batch_hash"` // Hash of the batch being proposed
	NodeID    string       `json:"node_id"`    // ID of the node sending this message
	Timestamp time.Time    `json:"timestamp"`
	Signature []byte       `json:"signature"`       // Signature of the message
	Batch     *state.Batch `json:"batch,omitempty"` // Only included in PrePrepare
}

// Hash returns the SHA256 hash of the message's contents
func (m *ConsensusMessage) Hash() string {
	// Exclude signature from hash calculation
	temp := m.Signature
	m.Signature = nil
	defer func() { m.Signature = temp }()

	data, err := json.Marshal(m)
	if err != nil {
		return ""
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
	PrePrepareMsg *ConsensusMessage
}

// NewConsensusState creates a new consensus state
func NewConsensusState(view, sequence int64, batch *state.Batch) *ConsensusState {
	batchData, _ := json.Marshal(batch)
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

// HasQuorum returns true if the number of messages received is greater than 2f+1
// where f is the maximum number of faulty nodes the system can tolerate
func HasQuorum(count int, totalNodes int) bool {
	f := (totalNodes - 1) / 3
	return count >= 2*f+1
}
