package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/p2p"
	"zkrollup/pkg/state"
)

// PBFT represents a PBFT consensus instance
type PBFT struct {
	node         *p2p.Node
	nodeID       string
	view         int64
	sequence     int64
	states       map[string]*ConsensusState // Map from batch hash to consensus state
	statesLock   sync.RWMutex
	isLeader     bool
	totalNodes   int
	decidedBatch chan *state.Batch
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewPBFT creates a new PBFT consensus instance
func NewPBFT(node *p2p.Node, nodeID string, isLeader bool) *PBFT {
	ctx, cancel := context.WithCancel(context.Background())
	return &PBFT{
		node:         node,
		nodeID:       nodeID,
		view:         0,
		sequence:     0,
		states:       make(map[string]*ConsensusState),
		isLeader:     isLeader,
		totalNodes:   1, // Will be updated as nodes join
		decidedBatch: make(chan *state.Batch),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the consensus process
func (p *PBFT) Start() {
	// Set up consensus message handler
	p.node.SetupProtocols(&p2p.ProtocolHandlers{
		OnConsensus: p.HandleMessage,
	})
}

// Stop stops the consensus process
func (p *PBFT) Stop() {
	p.cancel()
}

// ProposeBatch proposes a new batch for consensus
func (p *PBFT) ProposeBatch(batch *state.Batch) error {
	if !p.isLeader {
		return fmt.Errorf("only leader can propose batches")
	}

	log.Info().Msg("Leader proposing new batch for consensus")

	// Create consensus state for this batch
	state := NewConsensusState(p.view, p.sequence, batch)
	p.statesLock.Lock()
	p.states[state.BatchHash] = state
	p.statesLock.Unlock()

	// Create and broadcast pre-prepare message
	msg := &ConsensusMessage{
		Type:      PrePrepare,
		View:      p.view,
		Sequence:  p.sequence,
		BatchHash: state.BatchHash,
		NodeID:    p.nodeID,
		Timestamp: time.Now(),
		Batch:     batch,
	}

	log.Info().Str("batch_hash", state.BatchHash).Msg("Broadcasting pre-prepare message")

	if err := p.broadcast(msg); err != nil {
		return fmt.Errorf("failed to broadcast pre-prepare: %v", err)
	}

	p.sequence++
	return nil
}

// HandleMessage handles incoming consensus messages
func (p *PBFT) HandleMessage(data []byte) error {
	log.Info().Msg("Received consensus message")

	var msg ConsensusMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal consensus message: %v", err)
	}

	log.Info().
		Str("type", msg.Type.String()).
		Str("from", msg.NodeID).
		Str("batch_hash", msg.BatchHash).
		Msg("Processing consensus message")

	// Verify message basics
	if msg.View != p.view {
		return fmt.Errorf("message from wrong view")
	}

	p.statesLock.Lock()
	defer p.statesLock.Unlock()

	// Get or create consensus state
	state, exists := p.states[msg.BatchHash]
	if !exists {
		if msg.Type == PrePrepare {
			// Only create new state for PrePrepare messages
			log.Info().Msg("Creating new consensus state for pre-prepare message")
			state = NewConsensusState(msg.View, msg.Sequence, msg.Batch)
			state.PrePrepareMsg = &msg
			p.states[msg.BatchHash] = state
		} else {
			// For other message types, wait for PrePrepare
			log.Warn().
				Str("type", msg.Type.String()).
				Str("batch_hash", msg.BatchHash).
				Msg("Received consensus message before pre-prepare, storing for later")
			return nil
		}
	}

	// Process message based on type
	switch msg.Type {
	case PrePrepare:
		// Only non-leader nodes should process pre-prepare messages
		if p.isLeader {
			return nil
		}

		// Send prepare message
		prepare := &ConsensusMessage{
			Type:      Prepare,
			View:      p.view,
			Sequence:  msg.Sequence,
			BatchHash: msg.BatchHash,
			NodeID:    p.nodeID,
			Timestamp: time.Now(),
		}

		log.Info().Str("batch_hash", msg.BatchHash).Msg("Broadcasting prepare message")

		if err := p.broadcast(prepare); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast prepare message")
		}

	case Prepare:
		// Record prepare message
		state.PrepareCount[msg.NodeID] = true

		log.Info().
			Int("prepare_count", len(state.PrepareCount)).
			Int("total_nodes", p.totalNodes).
			Msg("Received prepare message")

		// If we have enough prepares, move to commit phase
		if len(state.PrepareCount) >= p.totalNodes-1 {
			log.Info().Msg("Received enough prepare messages, moving to commit phase")

			commit := &ConsensusMessage{
				Type:      Commit,
				View:      p.view,
				Sequence:  msg.Sequence,
				BatchHash: msg.BatchHash,
				NodeID:    p.nodeID,
				Timestamp: time.Now(),
			}

			if err := p.broadcast(commit); err != nil {
				log.Error().Err(err).Msg("Failed to broadcast commit message")
			}
		}

	case Commit:
		// Record commit message
		state.CommitCount[msg.NodeID] = true

		log.Info().
			Int("commit_count", len(state.CommitCount)).
			Int("total_nodes", p.totalNodes).
			Msg("Received commit message")

		// If we have enough commits, decide on the batch
		if len(state.CommitCount) >= p.totalNodes-1 && !state.Decided {
			log.Info().Msg("Received enough commit messages, finalizing batch")
			state.Decided = true
			p.decidedBatch <- state.Batch
			delete(p.states, msg.BatchHash) // Clean up state
		}
	}

	return nil
}

// broadcast sends a consensus message to all peers
func (p *PBFT) broadcast(msg *ConsensusMessage) error {
	log.Info().Msgf("Broadcasting consensus message: type=%s, from=%s, batch_hash=%s", msg.Type, msg.NodeID, msg.BatchHash)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus message: %v", err)
	}

	if err := p.node.BroadcastConsensus(p.ctx, data); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast consensus message")
		return err
	}

	log.Info().Msg("Successfully broadcasted consensus message")
	return nil
}

// GetDecidedBatchChan returns the channel that receives decided batches
func (p *PBFT) GetDecidedBatchChan() <-chan *state.Batch {
	return p.decidedBatch
}

// UpdateTotalNodes updates the total number of nodes in the network
func (p *PBFT) UpdateTotalNodes(count int) {
	p.totalNodes = count
}
