package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/l1"
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

	// CRS Ceremony related fields
	crsManager      *l1.CRSManager     // L1 CRS Manager client
	ptauState       *PTauCeremonyState // Current Powers of Tau ceremony state
	ptauStateLock   sync.RWMutex       // Lock for ptauState
	crsCeremonyDir  string             // Directory to store CRS ceremony files
	currentEpoch    int64              // Current epoch number
	crsCeremonyDone chan bool          // Channel to signal when CRS ceremony is complete
}

// NewPBFT creates a new PBFT consensus instance
func NewPBFT(node *p2p.Node, nodeID string, isLeader bool) *PBFT {
	ctx, cancel := context.WithCancel(context.Background())
	return &PBFT{
		node:            node,
		nodeID:          nodeID,
		view:            0,
		sequence:        0,
		states:          make(map[string]*ConsensusState),
		isLeader:        isLeader,
		totalNodes:      1, // Will be updated as nodes join
		decidedBatch:    make(chan *state.Batch),
		ctx:             ctx,
		cancel:          cancel,
		nodeIDs:         []string{nodeID}, // Initialize with self
		crsCeremonyDir:  filepath.Join(os.TempDir(), "zkrollup", "crs"),
		currentEpoch:    0,
		crsCeremonyDone: make(chan bool),
	}
}

// SetCRSManager sets the L1 CRS Manager client
func (p *PBFT) SetCRSManager(crsManager *l1.CRSManager) {
	p.crsManager = crsManager
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

	// Create CRS ceremony directory if it doesn't exist
	if err := os.MkdirAll(p.crsCeremonyDir, 0755); err != nil {
		log.Error().Err(err).Str("dir", p.crsCeremonyDir).Msg("Failed to create CRS ceremony directory")
	}

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

	// Handle CRS ceremony messages
	switch msg.Type {
	case CRSCeremonyStart:
		return p.handleCRSCeremonyStart(&msg)
	case CRSContribution:
		return p.handleCRSContribution(&msg)
	case CRSCeremonyComplete:
		return p.handleCRSCeremonyComplete(&msg)
	}

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
				Msg("Received consensus message for unknown batch")
			return fmt.Errorf("unknown batch hash: %s", msg.BatchHash)
		}
	}

	// Process message based on type
	switch msg.Type {
	case PrePrepare:
		// Validate the pre-prepare message
		if state.Phase != PrePrepare {
			return fmt.Errorf("received pre-prepare message in wrong phase")
		}

		// Store the pre-prepare message
		state.PrePrepareMsg = &msg

		// Send prepare message
		prepare := &ConsensusMessage{
			Type:      Prepare,
			View:      p.view,
			Sequence:  msg.Sequence,
			BatchHash: msg.BatchHash,
			NodeID:    p.nodeID,
			Timestamp: time.Now(),
		}

		log.Info().Str("batch_hash", msg.BatchHash).Msg("Sending prepare message")

		// Add our prepare message to the state
		state.PrepareCount[p.nodeID] = true

		// Broadcast the prepare message
		if err := p.broadcast(prepare); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast prepare message")
			return fmt.Errorf("failed to broadcast prepare: %v", err)
		}

		state.Phase = Prepare

	case Prepare:
		// Validate the prepare message
		if state.Phase > Prepare {
			return nil // Already moved past prepare phase
		}

		// Add the prepare message to the state
		state.PrepareCount[msg.NodeID] = true

		// Check if we have enough prepare messages to move to commit phase
		if len(state.PrepareCount) >= 2*(p.totalNodes/3)+1 && !state.SentCommit {
			// Send commit message
			commit := &ConsensusMessage{
				Type:      Commit,
				View:      p.view,
				Sequence:  msg.Sequence,
				BatchHash: msg.BatchHash,
				NodeID:    p.nodeID,
				Timestamp: time.Now(),
			}

			log.Info().Str("batch_hash", msg.BatchHash).Msg("Sending commit message")

			// Add our commit message to the state
			state.CommitCount[p.nodeID] = true
			state.SentCommit = true

			// Broadcast the commit message
			if err := p.broadcast(commit); err != nil {
				log.Error().Err(err).Msg("Failed to broadcast commit message")
				return fmt.Errorf("failed to broadcast commit: %v", err)
			}

			state.Phase = Commit
		}

	case Commit:
		// Add the commit message to the state
		state.CommitCount[msg.NodeID] = true

		// Check if we have enough commit messages to decide
		if len(state.CommitCount) >= 2*(p.totalNodes/3)+1 && !state.Decided {
			log.Info().Str("batch_hash", msg.BatchHash).Msg("Batch decided")
			state.Decided = true

			// If we're the leader, we should rotate leadership
			if p.isLeader {
				nextLeader := p.rotateLeader()
				log.Info().Str("next_leader", nextLeader).Msg("Rotating leadership")

				// Send leader rotation message
				rotation := &ConsensusMessage{
					Type:       LeaderRotation,
					View:       p.view,
					Sequence:   p.sequence,
					BatchHash:  "",
					NodeID:     p.nodeID,
					Timestamp:  time.Now(),
					NextLeader: nextLeader,
				}

				if err := p.broadcast(rotation); err != nil {
					log.Error().Err(err).Msg("Failed to broadcast leader rotation message")
				}

				// Update our leader status
				p.isLeader = (nextLeader == p.nodeID)
				p.view++
			}

			// Send the batch to the decided channel
			p.decidedBatch <- state.Batch
		}
	}

	return nil
}

// StartCRSCeremony initiates a new CRS ceremony
func (p *PBFT) StartCRSCeremony() error {
	if !p.isLeader {
		return fmt.Errorf("only leader can start CRS ceremony")
	}

	log.Info().Msg("Leader initiating CRS ceremony")

	// Increment epoch number
	p.currentEpoch++

	// Get a sorted copy of the node IDs to ensure consistent ordering
	p.nodeIDsLock.RLock()
	sortedNodeIDs := make([]string, len(p.nodeIDs))
	copy(sortedNodeIDs, p.nodeIDs)
	p.nodeIDsLock.RUnlock()
	sort.Strings(sortedNodeIDs)

	log.Info().Strs("participants", sortedNodeIDs).Msg("Ordered participants for CRS ceremony")

	// Create a new PTau ceremony state
	ptauState, err := NewPTauCeremonyState(p.currentEpoch, sortedNodeIDs, 12, p.crsCeremonyDir)
	if err != nil {
		return fmt.Errorf("failed to create PTau ceremony state: %v", err)
	}

	p.ptauStateLock.Lock()
	p.ptauState = ptauState
	p.ptauStateLock.Unlock()

	// Create and broadcast CRS ceremony start message
	msg := &ConsensusMessage{
		Type:         CRSCeremonyStart,
		View:         p.view,
		Sequence:     p.sequence,
		NodeID:       p.nodeID,
		Timestamp:    time.Now(),
		EpochNumber:  p.currentEpoch,
		Participants: sortedNodeIDs, // Include the ordered list of participants
	}

	log.Info().Int64("epoch", p.currentEpoch).Msg("Broadcasting CRS ceremony start message")

	// Broadcast the message
	if err := p.broadcast(msg); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast CRS ceremony start message")
		return fmt.Errorf("failed to broadcast CRS ceremony start: %v", err)
	}

	// If this node is the first participant, contribute to the ceremony
	if p.ptauState.CheckTurn(p.nodeID) {
		return p.contributeToCRSCeremony()
	}

	return nil
}

// contributeToCRSCeremony contributes to the current CRS ceremony
func (p *PBFT) contributeToCRSCeremony() error {
	p.ptauStateLock.RLock()
	if p.ptauState == nil {
		p.ptauStateLock.RUnlock()
		return fmt.Errorf("no active CRS ceremony")
	}

	if !p.ptauState.CheckTurn(p.nodeID) {
		p.ptauStateLock.RUnlock()
		return fmt.Errorf("not this node's turn to contribute")
	}
	p.ptauStateLock.RUnlock()

	// Generate random entropy for contribution
	entropy := fmt.Sprintf("%d-%s-%d", time.Now().UnixNano(), p.nodeID, p.currentEpoch)

	p.ptauStateLock.Lock()
	contributionMsg, err := p.ptauState.AddContribution(p.nodeID, entropy)
	p.ptauStateLock.Unlock()

	if err != nil {
		return fmt.Errorf("failed to add contribution: %v", err)
	}

	// Create and broadcast CRS contribution message
	msg := &ConsensusMessage{
		Type:            CRSContribution,
		View:            p.view,
		Sequence:        p.sequence,
		NodeID:          p.nodeID,
		Timestamp:       time.Now(),
		EpochNumber:     p.currentEpoch,
		ContributionMsg: contributionMsg,
		PTauFileData:    contributionMsg.PTauFileData,
	}

	log.Info().Int64("epoch", p.currentEpoch).Int("step", contributionMsg.Step).Msg("Broadcasting CRS contribution message")

	// Broadcast the message
	if err := p.broadcast(msg); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast CRS contribution message")
		return fmt.Errorf("failed to broadcast CRS contribution: %v", err)
	}

	return nil
}

// handleCRSCeremonyStart handles a CRS ceremony start message
func (p *PBFT) handleCRSCeremonyStart(msg *ConsensusMessage) error {
	log.Info().Int64("epoch", msg.EpochNumber).Strs("participants", msg.Participants).Msg("Received CRS ceremony start message")

	// Ensure we have participants list
	if len(msg.Participants) == 0 {
		return fmt.Errorf("received CRS ceremony start message without participants list")
	}

	// Create a new PTau ceremony state using the participants list from the message
	ptauState, err := NewPTauCeremonyState(msg.EpochNumber, msg.Participants, 12, p.crsCeremonyDir)
	if err != nil {
		return fmt.Errorf("failed to create PTau ceremony state: %v", err)
	}

	p.ptauStateLock.Lock()
	p.ptauState = ptauState
	p.currentEpoch = msg.EpochNumber
	p.ptauStateLock.Unlock()

	// If this node is the first participant, contribute to the ceremony
	if p.ptauState.CheckTurn(p.nodeID) {
		return p.contributeToCRSCeremony()
	}

	return nil
}

// handleCRSContribution handles a CRS contribution message
func (p *PBFT) handleCRSContribution(msg *ConsensusMessage) error {
	log.Info().Int64("epoch", msg.EpochNumber).Str("from", msg.NodeID).Msg("Received CRS contribution message")

	// Check if we have an active CRS ceremony
	p.ptauStateLock.RLock()
	if p.ptauState == nil {
		p.ptauStateLock.RUnlock()
		return fmt.Errorf("no active CRS ceremony")
	}
	p.ptauStateLock.RUnlock()

	// Verify the contribution
	p.ptauStateLock.Lock()
	if p.ptauState == nil || p.currentEpoch != msg.EpochNumber {
		p.ptauStateLock.Unlock()
		return fmt.Errorf("no matching CRS ceremony for epoch %d", msg.EpochNumber)
	}

	// Verify the contribution
	if err := p.ptauState.VerifyContribution(msg.ContributionMsg); err != nil {
		p.ptauStateLock.Unlock()
		return fmt.Errorf("failed to verify contribution: %v", err)
	}

	// Save the contribution
	tauFile, err := os.Create(p.ptauState.PTauPath)
	if err != nil {
		return fmt.Errorf("failed to create PTau file: %v", err)
	}
	if _, err := tauFile.Write(msg.PTauFileData); err != nil {
		return fmt.Errorf("failed to write PTau file: %v", err)
	}
	tauFile.Close()
	fmt.Printf("Saved PTau file to %s\n", p.ptauState.PTauPath)

	// Update the ceremony state
	p.ptauState.CurrentStep = msg.ContributionMsg.Step + 1

	// Check if the ceremony is complete
	isComplete := p.ptauState.CurrentStep >= len(p.ptauState.Participants)
	if isComplete {
		p.ptauState.Completed = true
	}

	// Check if it's this node's turn to contribute
	isMyTurn := p.ptauState.CheckTurn(p.nodeID)
	p.ptauStateLock.Unlock()

	// If it's this node's turn, contribute to the ceremony
	if isMyTurn {
		p.contributeToCRSCeremony()
	}

	// If the ceremony is complete and this node is the leader, finalize it
	p.ptauStateLock.RLock()
	isComplete = p.ptauState.CurrentStep >= len(p.ptauState.Participants)
	p.ptauStateLock.RUnlock()

	log.Info().Msgf("isComplete: %t, isLeader: %t", isComplete, p.isLeader)
	if isComplete && p.isLeader {
		log.Info().Msgf("Finalizing CRS ceremony for epoch %d", p.currentEpoch)
		return p.finalizeCRSCeremony()
	}

	return nil
}

// finalizeCRSCeremony finalizes the current CRS ceremony
func (p *PBFT) finalizeCRSCeremony() error {
	p.ptauStateLock.Lock()
	if p.ptauState == nil {
		p.ptauStateLock.Unlock()
		return fmt.Errorf("no active CRS ceremony")
	}

	// Mark the ceremony as completed
	p.ptauState.Completed = true

	// Finalize the ceremony
	finalPath, err := p.ptauState.FinalizeCeremony()
	if err != nil {
		p.ptauStateLock.Unlock()
		return fmt.Errorf("failed to finalize ceremony: %v", err)
	}
	p.ptauStateLock.Unlock()

	// Read the final PTau file
	finalData, err := os.ReadFile(finalPath)
	if err != nil {
		return fmt.Errorf("failed to read final PTau file: %v", err)
	}

	// Create and broadcast CRS ceremony complete message
	msg := &ConsensusMessage{
		Type:         CRSCeremonyComplete,
		View:         p.view,
		Sequence:     p.sequence,
		NodeID:       p.nodeID,
		Timestamp:    time.Now(),
		EpochNumber:  p.currentEpoch,
		PTauFileData: finalData,
	}

	log.Info().Int64("epoch", p.currentEpoch).Msg("Broadcasting CRS ceremony complete message")

	// Broadcast the message
	if err := p.broadcast(msg); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast CRS ceremony complete message")
		return fmt.Errorf("failed to broadcast CRS ceremony complete: %v", err)
	}

	// Signal that the CRS ceremony is complete
	select {
	case p.crsCeremonyDone <- true:
	default:
	}

	return nil
}

// handleCRSCeremonyComplete handles a CRS ceremony complete message
func (p *PBFT) handleCRSCeremonyComplete(msg *ConsensusMessage) error {
	log.Info().Int64("epoch", msg.EpochNumber).Msg("Received CRS ceremony complete message")

	// Save the final PTau file
	finalPath := filepath.Join(p.crsCeremonyDir, fmt.Sprintf("pot_epoch%d_final.ptau", msg.EpochNumber))
	if err := os.WriteFile(finalPath, msg.PTauFileData, 0644); err != nil {
		return fmt.Errorf("failed to save final PTau file: %v", err)
	}

	p.ptauStateLock.Lock()
	if p.ptauState != nil && p.currentEpoch == msg.EpochNumber {
		p.ptauState.Completed = true
		p.ptauState.PTauPath = finalPath
	}
	p.ptauStateLock.Unlock()

	// Signal that the CRS ceremony is complete
	select {
	case p.crsCeremonyDone <- true:
	default:
	}

	return nil
}

// GetCRSCeremonyDoneChan returns the channel that signals when a CRS ceremony is complete
func (p *PBFT) GetCRSCeremonyDoneChan() <-chan bool {
	return p.crsCeremonyDone
}

// broadcast sends a consensus message to all peers
func (p *PBFT) broadcast(msg *ConsensusMessage) error {
	// Marshal the message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus message: %v", err)
	}

	// Broadcast to all peers
	if err := p.node.BroadcastConsensus(p.ctx, data); err != nil {
		return fmt.Errorf("failed to broadcast consensus message: %v", err)
	}

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
	p.totalNodes = len(p.nodeIDs)

	log.Info().Str("node_id", nodeID).Int("total_nodes", p.totalNodes).Msg("Added new node ID to list")
}

// rotateLeader rotates the leader role to the next node in the list
func (p *PBFT) rotateLeader() string {
	p.nodeIDsLock.RLock()
	defer p.nodeIDsLock.RUnlock()

	// Sort the node IDs to ensure consistent leader rotation
	sortedNodeIDs := make([]string, len(p.nodeIDs))
	copy(sortedNodeIDs, p.nodeIDs)
	sort.Strings(sortedNodeIDs)

	// Find the current leader's position
	currentLeaderPos := -1
	for i, id := range sortedNodeIDs {
		if id == p.nodeID {
			currentLeaderPos = i
			break
		}
	}

	// If we couldn't find the current leader, default to the first node
	if currentLeaderPos == -1 {
		return sortedNodeIDs[0]
	}

	// Rotate to the next node
	nextLeaderPos := (currentLeaderPos + 1) % len(sortedNodeIDs)
	return sortedNodeIDs[nextLeaderPos]
}

// IsLeader returns whether this node is currently the leader
func (p *PBFT) IsLeader() bool {
	return p.isLeader
}
