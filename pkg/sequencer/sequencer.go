package sequencer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zkrollup/pkg/core"
	"zkrollup/pkg/crypto"
	"zkrollup/pkg/state"
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
}

func NewSequencer(config *core.Config) (*Sequencer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	prover, err := crypto.NewProver()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create prover: %v", err)
	}

	return &Sequencer{
		config:      config,
		state:       state.NewState(),
		txPool:      make([]state.Transaction, 0),
		proposalCh:  make(chan state.Batch),
		consensusCh: make(chan state.Batch),
		ctx:         ctx,
		cancel:      cancel,
		prover:      prover,
	}, nil
}

func (s *Sequencer) Start() error {
	go s.processBatches()
	go s.participateConsensus()
	return nil
}

func (s *Sequencer) Stop() {
	s.cancel()
}

func (s *Sequencer) AddTransaction(tx state.Transaction) error {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()

	// Basic transaction validation
	acc := s.state.GetAccount(tx.From)
	if acc.Nonce >= tx.Nonce {
		return fmt.Errorf("invalid nonce")
	}
	if acc.Balance.Cmp(tx.Amount) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	s.txPool = append(s.txPool, tx)
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

	// Propose batch to network
	s.proposalCh <- *batch
}

func (s *Sequencer) participateConsensus() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case batch := <-s.consensusCh:
			s.processFinalizedBatch(batch)
		}
	}
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

		// Create circuit assignment
		assignment := &crypto.TransactionCircuit{
			Amount:  tx.Amount.String(),
			Balance: from.Balance.String(),
		}

		// Generate and verify proof
		proof, err := s.prover.GenerateProof(assignment)
		if err != nil {
			return fmt.Errorf("failed to generate proof: %v", err)
		}

		// Verify proof
		valid, err := s.prover.VerifyProof(proof, assignment)
		if err != nil {
			return fmt.Errorf("failed to verify proof: %v", err)
		}

		if !valid {
			return fmt.Errorf("invalid transaction proof")
		}

		// Store proof
		tx.Signature = proof

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

	s.batchInProgress = false
	s.currentBatch = nil
	return nil
}
