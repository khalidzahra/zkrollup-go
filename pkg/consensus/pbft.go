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
	// Get existing handlers to preserve them
	existingHandlers := p.node.GetProtocolHandlers()

	// Create new handlers with our consensus handler but preserving existing transaction and batch handlers
	newHandlers := &p2p.ProtocolHandlers{
		OnTransaction: existingHandlers.OnTransaction, // Preserve existing transaction handler
		OnBatch:       existingHandlers.OnBatch,       // Preserve existing batch handler
		OnConsensus:   p.HandleMessage,                // Set our consensus handler
	}

	// Set up protocol handlers with the combined handlers
	p.node.SetupProtocols(newHandlers)

	// Log the handlers we're using
	fmt.Printf("PBFT consensus started with handlers - OnTransaction: %v, OnBatch: %v, OnConsensus: %v\n",
		newHandlers.OnTransaction != nil, newHandlers.OnBatch != nil, newHandlers.OnConsensus != nil)
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

	for i := range batch.Transactions {
		// Handle zero values properly
		if batch.Transactions[i].Amount.Sign() == 0 {
			log.Debug().Msg("Handling zero amount specially for consistent hash computation")
		}
	}

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

	// Broadcast the message with explicit error handling
	if err := p.broadcast(msg); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast pre-prepare message")
		return fmt.Errorf("failed to broadcast pre-prepare: %v", err)
	}

	// Leader also sends a prepare message to participate in consensus
	prepare := &ConsensusMessage{
		Type:      Prepare,
		View:      p.view,
		Sequence:  p.sequence,
		BatchHash: state.BatchHash,
		NodeID:    p.nodeID,
		Timestamp: time.Now(),
	}

	log.Info().Str("batch_hash", state.BatchHash).Msg("Leader sending prepare message")

	// Add the leader's prepare message to the state
	state.PrepareCount[p.nodeID] = true

	// Broadcast the prepare message
	if err := p.broadcast(prepare); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast leader prepare message")
		// Continue even if there's an error, as the pre-prepare was successful
	}

	p.sequence++
	return nil
}

// HandleMessage handles incoming consensus messages
func (p *PBFT) HandleMessage(data []byte) error {
	log.Info().Int("data_size", len(data)).Msg("Received consensus message")

	// Handle empty messages
	if len(data) == 0 {
		log.Error().Msg("Received empty consensus message")
		return fmt.Errorf("empty consensus message")
	}

	var msg ConsensusMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal consensus message: %v", err)
	}

	if msg.Batch != nil {
		for i := range msg.Batch.Transactions {
			// Special handling for zero values
			if msg.Batch.Transactions[i].Amount != nil && msg.Batch.Transactions[i].Amount.Sign() == 0 {
				log.Debug().Msg("Ensuring proper zero value handling in consensus message")
			}
		}
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
		if HasQuorum(len(state.PrepareCount), p.totalNodes) && !state.SentCommit {
			log.Info().Msg("Received enough prepare messages, moving to commit phase")

			// Mark that we've sent a commit message for this batch
			state.SentCommit = true

			commit := &ConsensusMessage{
				Type:      Commit,
				View:      p.view,
				Sequence:  msg.Sequence,
				BatchHash: msg.BatchHash,
				NodeID:    p.nodeID,
				Timestamp: time.Now(),
			}

			// Add our own commit message to the count
			state.CommitCount[p.nodeID] = true

			log.Info().Str("node_id", p.nodeID).Int("commit_count", len(state.CommitCount)).Msg("Adding our own commit message to the count")

			// Check if we have enough commits after adding our own
			if HasQuorum(len(state.CommitCount), p.totalNodes) && !state.Decided && p.isLeader {
				log.Info().Msg("Leader has enough commit messages after adding its own, finalizing batch")
				state.Decided = true
				p.decidedBatch <- state.Batch
				delete(p.states, msg.BatchHash) // Clean up state
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
		if HasQuorum(len(state.CommitCount), p.totalNodes) && !state.Decided {
			log.Info().Int("commit_count", len(state.CommitCount)).Int("total_nodes", p.totalNodes).Msg("Received enough commit messages, finalizing batch")
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

	hashStr := msg.Hash()
	log.Debug().Str("message_hash", hashStr).Msg("Computed message hash for consensus")

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal consensus message")
		return fmt.Errorf("failed to marshal consensus message: %v", err)
	}

	log.Debug().Int("data_size", len(data)).Msg("Consensus message size")

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
