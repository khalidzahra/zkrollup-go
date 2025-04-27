package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
	nodeIDs      []string     // List of all node IDs in the network
	nodeIDsLock  sync.RWMutex // Lock for nodeIDs
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
		nodeIDs:      []string{nodeID}, // Initialize with self
	}
}

// Start starts the consensus process
func (p *PBFT) Start() {
	existingHandlers := p.node.GetProtocolHandlers()

	newHandlers := &p2p.ProtocolHandlers{
		OnTransaction: existingHandlers.OnTransaction,
		OnBatch:       existingHandlers.OnBatch,
		OnConsensus:   p.HandleMessage,
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

	// Check if we're in standalone mode (no peers)
	if p.totalNodes <= 1 {
		log.Info().Str("batch_hash", state.BatchHash).Msg("Running in standalone mode, automatically committing batch")
		// In standalone mode, we can automatically commit the batch
		state.CommitCount[p.nodeID] = true
		// Send the batch to the decided channel
		p.decidedBatch <- batch
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

	log.Info().
		Str("type", msg.Type.String()).
		Str("from", msg.NodeID).
		Str("batch_hash", msg.BatchHash).
		Msg("Processing consensus message")

	// Add the node ID to our list if it's not already there
	p.addNodeID(msg.NodeID)

	// Handle leader rotation messages separately as they don't depend on batch state
	if msg.Type == LeaderRotation {
		log.Info().Str("from", msg.NodeID).Str("next_leader", msg.NextLeader).Msg("Received leader rotation message")

		// Update our leader status based on the leader's decision
		p.isLeader = (msg.NextLeader == p.nodeID)
		p.view = msg.View + 1 // Set view to match the leader's next view

		if p.isLeader {
			log.Info().Str("node_id", p.nodeID).Int64("view", p.view).Msg("This node is now the leader based on leader rotation message")
		} else {
			log.Info().Str("node_id", p.nodeID).Str("new_leader", msg.NextLeader).Int64("view", p.view).Msg("Leadership transferred based on leader rotation message")
		}

		return nil
	}

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

			// If we're the leader, initiate leader rotation
			if p.isLeader {
				// Determine the next leader
				nextLeader := p.rotateLeader()
				log.Info().Str("next_leader", nextLeader).Msg("Broadcasting leader rotation message")

				// Send a dedicated leader rotation message
				leaderMsg := &ConsensusMessage{
					Type:       LeaderRotation,
					View:       p.view - 1, // Use the previous view since we just incremented it
					Sequence:   msg.Sequence,
					BatchHash:  msg.BatchHash,
					NodeID:     p.nodeID,
					Timestamp:  time.Now(),
					NextLeader: nextLeader,
				}

				if err := p.broadcast(leaderMsg); err != nil {
					log.Error().Err(err).Msg("Failed to broadcast leader rotation message")
				}
			}

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

// addNodeID adds a node ID to the list if it's not already there
func (p *PBFT) addNodeID(nodeID string) {
	p.nodeIDsLock.Lock()
	defer p.nodeIDsLock.Unlock()

	// Check if the node ID is already in the list
	for _, id := range p.nodeIDs {
		if id == nodeID {
			return
		}
	}

	// Add the node ID to the list
	p.nodeIDs = append(p.nodeIDs, nodeID)
	log.Info().Str("node_id", nodeID).Int("total_nodes", len(p.nodeIDs)).Msg("Added new node ID to the list")
}

// rotateLeader rotates the leader role to the next node in the list
func (p *PBFT) rotateLeader() string {
	p.nodeIDsLock.Lock()
	defer p.nodeIDsLock.Unlock()

	log.Debug().Int("total_nodes", len(p.nodeIDs)).Msg("Rotating leader")

	sort.Strings(p.nodeIDs)
	if len(p.nodeIDs) <= 1 {
		p.isLeader = true
		return p.nodeID
	}
	p.view++

	// Deterministically select the leader based on the view number
	leaderIndex := int(p.view) % len(p.nodeIDs)
	nextLeaderID := p.nodeIDs[leaderIndex]

	log.Debug().Int64("view", p.view).Str("next_leader", nextLeaderID).Msg("Rotating leader")

	// Update the leader status
	wasLeader := p.isLeader
	p.isLeader = (nextLeaderID == p.nodeID)

	if wasLeader != p.isLeader {
		if p.isLeader {
			log.Info().Str("node_id", p.nodeID).Int64("view", p.view).Msg("This node is now the leader")
		} else {
			log.Info().Str("node_id", p.nodeID).Str("new_leader", nextLeaderID).Int64("view", p.view).Msg("Leadership transferred to another node")
		}
	}

	return nextLeaderID
}

// IsLeader returns whether this node is currently the leader
func (p *PBFT) IsLeader() bool {
	return p.isLeader
}
