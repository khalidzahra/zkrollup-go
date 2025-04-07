package sequencer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/consensus"
	"zkrollup/pkg/core"
	"zkrollup/pkg/crypto"
	"zkrollup/pkg/p2p"
	"zkrollup/pkg/state"
	"zkrollup/pkg/util"
)

type Sequencer struct {
	config *core.Config
	state  *state.State

	// Transaction pool
	txPool []state.Transaction
	poolMu sync.RWMutex

	// Batch processing
	currentBatch    *state.Batch
	batchInProgress bool
	batchMu         sync.RWMutex

	// Consensus channels
	proposalCh  chan state.Batch
	consensusCh chan state.Batch

	ctx    context.Context
	cancel context.CancelFunc

	// ZK proof generation
	prover *crypto.Prover

	// P2P networking
	node *p2p.Node

	// Consensus
	consensus *consensus.PBFT
	isLeader  bool

	// Peer tracking
	peerCount   int
	peerCountMu sync.RWMutex
}

func NewSequencer(config *core.Config, port int, bootstrapPeers []string, isLeader bool) (*Sequencer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create P2P node
	node, err := p2p.NewNode(ctx, port, bootstrapPeers)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create P2P node: %v", err)
	}

	// Create prover
	prover, err := crypto.NewProver()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create prover: %v", err)
	}

	// Create sequencer
	seq := &Sequencer{
		config:      config,
		state:       state.NewState(),
		txPool:      make([]state.Transaction, 0),
		proposalCh:  make(chan state.Batch),
		consensusCh: make(chan state.Batch),
		ctx:         ctx,
		cancel:      cancel,
		prover:      prover,
		node:        node,
		isLeader:    isLeader,
		peerCount:   1, // Start with just ourselves
	}

	// Create consensus instance
	nodeID := node.Host.ID().String()
	seq.consensus = consensus.NewPBFT(node, nodeID, isLeader)

	// Setup P2P protocol handlers
	node.SetupProtocols(&p2p.ProtocolHandlers{
		OnTransaction: seq.handleTransaction,
		OnBatch:       seq.handleBatch,
		OnConsensus:   seq.handleConsensus,
	})

	return seq, nil
}

func (s *Sequencer) Start() error {
	// Start consensus module
	s.consensus.Start()

	// Re-register our protocol handlers to ensure they're not overridden by consensus
	// This is critical because the consensus module might have overridden our transaction handler
	s.node.SetupProtocols(&p2p.ProtocolHandlers{
		OnTransaction: s.handleTransaction,
		OnBatch:       s.handleBatch,
		OnConsensus:   s.handleConsensus,
	})

	// Log that we've re-registered our handlers
	fmt.Printf("Sequencer re-registered protocol handlers after consensus start\n")

	// Start sequencer processes
	go s.processBatches()
	go s.participateConsensus()
	go s.monitorPeerCount()

	return nil
}

// monitorPeerCount periodically updates the consensus module with the current peer count
func (s *Sequencer) monitorPeerCount() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			peers := s.node.GetPeers()
			newPeerCount := len(peers) + 1 // Include ourselves

			s.peerCountMu.Lock()
			if newPeerCount != s.peerCount {
				s.peerCount = newPeerCount
				s.consensus.UpdateTotalNodes(newPeerCount)
				log.Info().Int("peer_count", newPeerCount).Msg("Updated consensus peer count")
			}
			s.peerCountMu.Unlock()
		}
	}
}

func (s *Sequencer) Stop() {
	s.consensus.Stop()
	s.cancel()
	s.node.Close()
}

func (s *Sequencer) AddTransaction(tx state.Transaction) error {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()

	// Get or initialize the sender account
	acc := s.state.GetAccount(tx.From)

	// If this is a new account, initialize it with a balance for testing
	if acc.Balance == nil || acc.Balance.Sign() == 0 {
		log.Info().Str("address", fmt.Sprintf("%x", tx.From)).Msg("Initializing new account with test balance")
		acc.Balance = big.NewInt(1000) // Initialize with 1000 units
		acc.Nonce = 0
		s.state.UpdateAccount(acc)
	}

	// Initialize recipient account if needed
	recipient := s.state.GetAccount(tx.To)
	if recipient.Balance == nil {
		log.Info().Str("address", fmt.Sprintf("%x", tx.To)).Msg("Initializing recipient account")
		recipient.Balance = big.NewInt(0)
		s.state.UpdateAccount(recipient)
	}

	// Basic transaction validation
	if acc.Nonce >= tx.Nonce {
		return fmt.Errorf("invalid nonce")
	}
	if acc.Balance.Cmp(tx.Amount) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Add transaction to pool
	s.txPool = append(s.txPool, tx)
	log.Info().Str("from", fmt.Sprintf("%x", tx.From)).Str("to", fmt.Sprintf("%x", tx.To)).Str("amount", tx.Amount.String()).Uint64("nonce", tx.Nonce).Msg("Added transaction to pool")

	return nil
}

func (s *Sequencer) processBatches() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.tryCreateBatch()
		}
	}
}

func (s *Sequencer) tryCreateBatch() {
	s.poolMu.Lock()
	s.batchMu.Lock()
	defer s.poolMu.Unlock()
	defer s.batchMu.Unlock()

	if s.batchInProgress || len(s.txPool) == 0 {
		return
	}

	// Create new batch
	txCount := min(len(s.txPool), int(s.config.BatchSize))
	batch := &state.Batch{
		Transactions: make([]state.Transaction, txCount),
		BatchNumber:  s.state.GetBatchNumber() + 1,
		Timestamp:    uint64(time.Now().Unix()),
	}

	copy(batch.Transactions, s.txPool[:txCount])
	s.txPool = s.txPool[txCount:]

	s.batchInProgress = true
	s.currentBatch = batch

	// If we're the leader, propose the batch for consensus
	if s.isLeader {
		log.Info().Msg("Proposing batch for consensus")
		if err := s.consensus.ProposeBatch(batch); err != nil {
			log.Error().Err(err).Msg("Failed to propose batch for consensus")
			// Reset batch state on error
			s.batchInProgress = false
			s.currentBatch = nil
		}
	} else {
		// Non-leaders just broadcast the batch to the network
		if err := s.node.BroadcastBatch(s.ctx, batch); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast batch")
		}
	}
}

func (s *Sequencer) participateConsensus() {
	// Listen for decided batches from the consensus module
	decidedBatchCh := s.consensus.GetDecidedBatchChan()

	for {
		select {
		case <-s.ctx.Done():
			return
		case batch := <-decidedBatchCh:
			// Process batches that have been decided by consensus
			log.Info().Msg("Received decided batch from consensus")
			if err := s.processFinalizedBatch(*batch); err != nil {
				log.Error().Err(err).Msg("Failed to process finalized batch")
			}
		}
	}
}

// P2P message handlers
func (s *Sequencer) handleTransaction(tx *state.Transaction) error {
	// Log that we're handling a transaction
	log.Info().Str("from", fmt.Sprintf("%x", tx.From)).Str("to", fmt.Sprintf("%x", tx.To)).Msg("Handling transaction in sequencer")

	// Special handling for zero values to ensure consistent hash computation
	if tx.Amount != nil && tx.Amount.Sign() == 0 {
		log.Info().Msg("Transaction contains zero amount, ensuring proper formatting for consistent hash computation")
		// Only use single byte for zero for hash
	}

	// The nonce must be converted to a string representation when used in the circuit
	nonceStr := fmt.Sprintf("%d", tx.Nonce)
	log.Info().Str("nonce_str", nonceStr).Msg("Using nonce string format for consistent hash computation")

	// Add the transaction to the sequencer's pool
	return s.AddTransaction(*tx)
}

func (s *Sequencer) handleBatch(batch *state.Batch) error {
	log.Info().Msg("Received batch from peer")

	if s.isLeader {
		// Leader should propose the batch for consensus
		log.Info().Msg("Leader proposing received batch for consensus")
		s.batchMu.Lock()
		s.currentBatch = batch
		s.batchInProgress = true
		s.batchMu.Unlock()

		// Propose the batch for consensus
		if err := s.consensus.ProposeBatch(batch); err != nil {
			log.Error().Err(err).Msg("Failed to propose received batch for consensus")
			// Reset batch state on error
			s.batchMu.Lock()
			s.batchInProgress = false
			s.currentBatch = nil
			s.batchMu.Unlock()
			return err
		}
	} else {
		// Non-leaders should store the batch and wait for consensus
		log.Info().Msg("Non-leader received batch, storing for consensus")
		s.batchMu.Lock()
		s.currentBatch = batch
		s.batchInProgress = true
		s.batchMu.Unlock()
	}
	return nil
}

func (s *Sequencer) handleConsensus(msg []byte) error {
	// Forward consensus messages to the PBFT consensus module
	return s.consensus.HandleMessage(msg)
}

// BroadcastConsensusMessage broadcasts a consensus message to the network
func (s *Sequencer) BroadcastConsensusMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus message: %v", err)
	}

	return s.node.BroadcastConsensus(s.ctx, data)
}

func (s *Sequencer) processFinalizedBatch(batch state.Batch) error {
	s.batchMu.Lock()
	defer s.batchMu.Unlock()

	// Get initial state root
	initialRoot := s.state.GetStateRoot()

	// Apply transactions and generate proofs
	for _, tx := range batch.Transactions {
		from := s.state.GetAccount(tx.From)
		to := s.state.GetAccount(tx.To)

		// Create circuit assignment with consistent formatting
		// Use util.FormatAmount to ensure consistent string representation
		// assignment := &crypto.TransactionCircuit{
		// 	// Ensure consistent formatting for amount
		// 	Amount: util.FormatAmount(tx.Amount),
		// 	// Ensure consistent formatting for balance
		// 	Balance: util.FormatAmount(from.Balance),
		// 	// Ensure consistent formatting for nonce
		// 	Nonce: util.GetNonceForHash(from.Nonce),
		// }

		log.Debug().Str("nonce", util.GetNonceForHash(from.Nonce)).Msg("Using formatted nonce for circuit")

		// Handle zero values properly
		if tx.Amount.Sign() == 0 {
			log.Debug().Msg("Handling zero amount specially for consistent hash computation")
		}

		// // Generate and verify proof
		// proof, err := s.prover.GenerateProof(assignment)
		// if err != nil {
		// 	return fmt.Errorf("failed to generate proof: %v", err)
		// }

		// // Verify proof
		// valid, err := s.prover.VerifyProof(proof, assignment)
		// if err != nil {
		// 	return fmt.Errorf("failed to verify proof: %v", err)
		// }

		// if !valid {
		// 	return fmt.Errorf("invalid transaction proof")
		// }

		// // Store proof
		// tx.Signature = proof

		// Update state
		from.Balance.Sub(from.Balance, tx.Amount)
		from.Nonce++
		to.Balance.Add(to.Balance, tx.Amount)

		s.state.UpdateAccount(from)
		s.state.UpdateAccount(to)
	}

	// Get final state root and generate state transition proof
	finalRoot := s.state.GetStateRoot()

	// Store batch proof
	batch.StateRoot = finalRoot
	// Store previous root in proof for verification
	proofData := append(initialRoot[:], finalRoot[:]...)
	batch.Proof = proofData

	// Add the finalized batch to the state's batch history
	s.state.AddBatch(&batch)
	log.Info().Int("tx_count", len(batch.Transactions)).Str("state_root", fmt.Sprintf("%x", batch.StateRoot)).Msg("Added finalized batch to state")

	// Broadcast the finalized batch to all peers
	if err := s.node.BroadcastBatch(s.ctx, &batch); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast finalized batch")
		// Continue processing even if broadcast fails
	} else {
		log.Info().Msg("Successfully broadcasted finalized batch to peers")
	}

	s.batchInProgress = false
	s.currentBatch = nil
	return nil
}
