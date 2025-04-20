package consensus

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/p2p"
	"zkrollup/pkg/state"
)

const (
	CRSEpochDuration        = 10 * time.Second
	CRSParticipantsPerEpoch = 3
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

	// CRS Ceremony integration
	crsCeremony     *CRSCeremonyState
	crsCeremonyLock sync.Mutex
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

	p.StartCRSEpochAdvancer(CRSEpochDuration, CRSParticipantsPerEpoch, p.GetSequencerIDs)
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

	// In standalone mode, we also need to update the commit count
	if p.totalNodes <= 1 {
		log.Info().Str("batch_hash", state.BatchHash).Msg("Standalone mode: Updating commit count")
		// No need to create and broadcast a commit message, just update the state
		state.CommitCount[p.nodeID] = true
	}

	p.sequence++
	return nil
}

// ProposeCRSEpoch proposes a new CRS epoch via PBFT consensus (leader only)
func (p *PBFT) ProposeCRSEpoch(epoch CRSEpoch) error {
	msg := &ConsensusMessage{
		Type:      CRSEpochProposal,
		View:      p.view,
		Sequence:  p.sequence,
		NodeID:    p.nodeID,
		Timestamp: time.Now(),
		CRSEpoch:  &epoch,
	}
	if err := p.broadcast(msg); err != nil {
		return fmt.Errorf("failed to broadcast CRS epoch proposal: %v", err)
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

	case CRSEpochProposal:
		// Validate the epoch proposal (number, timing, participant selection)
		// For now, accept if strictly increasing and not overlapping
		current := GetCRSEpoch()
		if msg.CRSEpoch.Number <= current.Number {
			return fmt.Errorf("stale CRS epoch proposal")
		}
		if !current.StartTime.IsZero() && msg.CRSEpoch.StartTime.Before(current.StartTime.Add(current.Duration)) {
			return fmt.Errorf("CRS epoch overlaps previous epoch")
		}
		// Optionally: validate participant selection is deterministic
		SetCRSEpoch(*msg.CRSEpoch)
		log.Info().Int64("number", msg.CRSEpoch.Number).Strs("participants", msg.CRSEpoch.Participants).Msg("Committed new CRS epoch")

		// === CRS Ceremony Integration ===
		sequencers := msg.CRSEpoch.Participants
		initialCRS, _, err := GenerateInitialCRS(len(sequencers))
		if err != nil {
			return fmt.Errorf("failed to generate initial CRS: %w", err)
		}
		p.crsCeremonyLock.Lock()
		p.crsCeremony = NewCRSCeremonyState(msg.CRSEpoch.Number, sequencers, initialCRS)
		p.crsCeremonyLock.Unlock()
		// If this node is first, contribute and broadcast
		if sequencers[0] == p.nodeID {
			go p.advanceCRSCeremony()
		}
		return nil

	case CRSCeremony:
		// Handle CRS Ceremony Message
		p.crsCeremonyLock.Lock()
		ceremony := p.crsCeremony
		p.crsCeremonyLock.Unlock()
		if ceremony == nil || msg.CRSCeremony.EpochNumber != ceremony.EpochNumber {
			return nil // Ignore messages for other epochs
		}
		if msg.CRSCeremony.Step != ceremony.CurrentStep {
			return nil // Ignore out-of-order steps
		}
		if msg.CRSCeremony.ContributorID != ceremony.Participants[ceremony.CurrentStep] {
			return nil // Not the correct contributor
		}
		// Accept the contribution
		ceremony.IntermediateCRS = msg.CRSCeremony.IntermediateCRS
		ceremony.CurrentStep++
		// If ceremony complete
		if ceremony.CurrentStep >= len(ceremony.Participants) {
			log.Info().Msg("CRS ceremony complete. Final CRS ready.")
			// TODO: Anchor or store final CRS
			return nil
		}
		// If it's now this node's turn, contribute and broadcast
		if ceremony.Participants[ceremony.CurrentStep] == p.nodeID {
			go p.advanceCRSCeremony()
		}
		return nil
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

	// Self-delivery: process the message locally as if received from the network
	go func() {
		if err := p.HandleMessage(data); err != nil {
			log.Error().Err(err).Msg("Failed to process own consensus message")
		}
	}()

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

// StartCRSEpochAdvancer launches a goroutine that checks and advances the CRS epoch every interval
func (p *PBFT) StartCRSEpochAdvancer(epochDuration time.Duration, participantsPerEpoch int, getSequencers func() []string) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ticker.C:
				now := time.Now().UTC()
				epoch := GetCRSEpoch()
				if epoch.StartTime.IsZero() || now.After(epoch.StartTime.Add(epoch.Duration)) {
					// Only leader proposes
					if p.isLeader && (p.crsCeremony == nil || p.crsCeremony.Completed) {
						sequencers := getSequencers()
						log.Info().Strs("sequencers_seen", sequencers).Msg("[CRS Advancer] Sequencers visible to leader")
						seed := sha256.Sum256([]byte(fmt.Sprintf("%d-%s", now.UnixNano(), sequencers)))
						participants := SelectCRSParticipants(sequencers, participantsPerEpoch, seed[:])
						log.Info().Strs("participants", participants).Msg("[CRS Advancer] Selected CRS participants for new epoch")
						newEpoch := CRSEpoch{
							Number:       epoch.Number + 1,
							StartTime:    now,
							Duration:     epochDuration,
							Participants: participants,
						}
						_ = p.ProposeCRSEpoch(newEpoch)
					}
				}
			}
		}
	}()
}

// GetSequencerIDs returns the current list of sequencer node IDs
func (p *PBFT) GetSequencerIDs() []string {
	peerIDs := p.node.GetPeers()
	ids := make([]string, 0, len(peerIDs)+1)
	ids = append(ids, p.nodeID) // Include self
	for _, pid := range peerIDs {
		ids = append(ids, pid.String())
	}
	return ids
}

// advanceCRSCeremony is called when it's this node's turn to contribute
func (p *PBFT) advanceCRSCeremony() {
	p.crsCeremonyLock.Lock()
	ceremony := p.crsCeremony
	p.crsCeremonyLock.Unlock()
	if ceremony == nil {
		return
	}
	if !ceremony.CheckTurn(p.nodeID) {
		return
	}
	newCRS, _, proof, err := AddContribution(ceremony.IntermediateCRS)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add CRS contribution")
		return
	}
	log.Info().Str("contributor", p.nodeID).
		Int("step", ceremony.CurrentStep).
		Str("crs_hash", fmt.Sprintf("%x", sha256.Sum256(bytes.Join(newCRS, nil)))).
		Msg("CRS contribution added")
	msg := &ConsensusMessage{
		Type:      CRSCeremony,
		NodeID:    p.nodeID,
		Timestamp: time.Now(),
		CRSCeremony: &CRSCeremonyMessage{
			EpochNumber:     ceremony.EpochNumber,
			Step:            ceremony.CurrentStep,
			IntermediateCRS: newCRS,
			ContributorID:   p.nodeID,
			Signature:       nil, // TODO: Add signature
			Proof:           proof,
		},
	}
	// Broadcast and update local state
	if err := p.broadcast(msg); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast CRS ceremony message")
		return
	}
	// If this was the last contribution, log the final CRS
	if ceremony.CurrentStep == len(ceremony.Participants)-1 {
		log.Info().Str("final_crs_hash", fmt.Sprintf("%x", sha256.Sum256(bytes.Join(newCRS, nil)))).
			Msg("Final CRS generated for epoch")
	}
}
