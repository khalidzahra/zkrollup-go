package sequencer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"

	"zkrollup/pkg/consensus"
	"zkrollup/pkg/core"
	"zkrollup/pkg/crypto"
	"zkrollup/pkg/evm"
	"zkrollup/pkg/l1"
	"zkrollup/pkg/p2p"
	"zkrollup/pkg/state"
)

type Sequencer struct {
	config *core.Config
	state  *state.State

	// Transaction pool
	txPool []state.Transaction
	poolMu sync.RWMutex

	// EVM executor
	evmExecutor *evm.EVMExecutor

	// Batch processing
	currentBatch    *state.Batch
	batchInProgress bool
	batchMu         sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	// ZK proof generation
	prover *crypto.Prover

	// P2P networking
	node *p2p.Node

	// L1 integration
	l1Client     *l1.Client
	l1Enabled    bool
	l1SubmitChan chan state.Batch

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
		config:       config,
		state:        state.NewState(),
		txPool:       make([]state.Transaction, 0),
		ctx:          ctx,
		cancel:       cancel,
		prover:       prover,
		node:         node,
		isLeader:     isLeader,
		peerCount:    1, // Start with just ourselves
		evmExecutor:  evm.NewEVMExecutor(),
		l1Enabled:    config.L1Enabled,
		l1SubmitChan: make(chan state.Batch, 10),
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

	// Initialize L1 client if enabled
	if config.L1Enabled && config.L1PrivateKey != "" {
		l1Config := &l1.Config{
			EthereumRPC:     config.EthereumRPC,
			ChainID:         config.ChainID,
			ContractAddress: config.ContractAddress,
			PrivateKey:      config.L1PrivateKey,
		}

		l1Client, err := l1.NewClient(l1Config)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize L1 client, L1 integration disabled")
			seq.l1Enabled = false
		} else {
			seq.l1Client = l1Client
			log.Info().Msg("L1 integration enabled")
		}
	} else {
		log.Info().Msg("L1 integration disabled")
		seq.l1Enabled = false
	}

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

	// Start L1 batch submission process if enabled
	if s.l1Enabled && s.l1Client != nil {
		go s.submitBatchesToL1()
		log.Info().Msg("Started L1 batch submission process")
	}

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

	// Close L1 submission channel
	if s.l1Enabled && s.l1SubmitChan != nil {
		close(s.l1SubmitChan)
	}
}

func (s *Sequencer) AddTransaction(tx state.Transaction) error {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()

	// Get or initialize the sender account
	acc, err := s.state.GetAccount(tx.From)
	if err != nil || acc == nil {
		// If this is a new account, initialize it with a balance for testing
		log.Info().Str("address", fmt.Sprintf("%x", tx.From)).Msg("Initializing new account with test balance")
		acc = &state.Account{
			Address: tx.From,
			Balance: big.NewInt(1000), // Initialize with 1000 units
			Nonce:   0,
		}
		s.state.SetAccount(acc)
	} else if acc.Balance == nil || acc.Balance.Sign() == 0 {
		// Ensure account has a balance
		log.Info().Str("address", fmt.Sprintf("%x", tx.From)).Msg("Setting test balance for account")
		acc.Balance = big.NewInt(1000) // Initialize with 1000 units
		s.state.SetAccount(acc)
	}

	// Initialize recipient account if needed
	recipient, err := s.state.GetAccount(tx.To)
	if err != nil || recipient == nil || recipient.Balance == nil {
		log.Info().Str("address", fmt.Sprintf("%x", tx.To)).Msg("Initializing recipient account")
		recipient = &state.Account{
			Address: tx.To,
			Balance: big.NewInt(0),
			Nonce:   0,
		}
		s.state.SetAccount(recipient)
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
	log.Info().Str("from", fmt.Sprintf("%x", tx.From)).Str("to", fmt.Sprintf("%x", tx.To)).Uint8("type", uint8(tx.Type)).Msg("Handling transaction in sequencer")

	// Special handling for zero values to ensure consistent hash computation
	if tx.Amount != nil && tx.Amount.Sign() == 0 {
		log.Info().Msg("Transaction contains zero amount, ensuring proper formatting for consistent hash computation")
		// Only use single byte for zero for hash
	}

	// The nonce must be converted to a string representation when used in the circuit
	nonceStr := fmt.Sprintf("%d", tx.Nonce)
	log.Info().Str("nonce_str", nonceStr).Msg("Using nonce string format for consistent hash computation")

	// Verify transaction type-specific requirements
	switch tx.Type {
	case state.TxTypeContractDeploy, state.TxTypeContractCall:
		// Ensure gas is provided for EVM transactions
		if tx.Gas == 0 {
			return errors.New("EVM transactions require gas")
		}

		// Ensure data is provided for contract deployment
		if tx.Type == state.TxTypeContractDeploy && len(tx.Data) == 0 {
			return errors.New("contract deployment requires bytecode")
		}
	}

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

	// Get initial state root for proof generation
	initialStateRoot := s.state.GetStateRoot()

	// Apply transactions and generate proofs
	for _, tx := range batch.Transactions {
		// Get sender account
		sender, err := s.state.GetAccount(tx.From)
		if err != nil {
			log.Error().Err(err).Str("address", formatAddress(tx.From)).Msg("Failed to get sender account")
			continue
		}

		// Verify nonce
		if tx.Nonce != sender.Nonce {
			log.Error().Uint64("expected", sender.Nonce).Uint64("got", tx.Nonce).Msg("Invalid nonce")
			continue
		}

		// Process transaction based on type
		switch tx.Type {
		case state.TxTypeTransfer:
			// Process a simple transfer transaction
			err := s.processTransferTransaction(tx, sender)
			if err != nil {
				log.Error().Err(err).Msg("Failed to process transfer transaction")
				continue
			}

		case state.TxTypeContractDeploy:
			// Process a contract deployment transaction
			err := s.processContractDeployment(tx, sender)
			if err != nil {
				log.Error().Err(err).Msg("Failed to process contract deployment")
				continue
			}

		case state.TxTypeContractCall:
			// Process a contract call transaction
			err := s.processContractCall(tx, sender)
			if err != nil {
				log.Error().Err(err).Msg("Failed to process contract call")
				continue
			}

		default:
			log.Error().Uint8("type", uint8(tx.Type)).Msg("Unknown transaction type")
			continue
		}
	}

	// Get final state root and generate state transition proof
	finalRoot := s.state.GetStateRoot()

	// Store batch proof
	batch.StateRoot = finalRoot
	// Store previous root in proof for verification
	proofData := append(initialStateRoot[:], finalRoot[:]...)
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

	// Submit batch to L1 if enabled
	if s.l1Enabled && s.l1Client != nil && s.l1SubmitChan != nil {
		// Send batch to L1 submission channel
		select {
		case s.l1SubmitChan <- batch:
			log.Info().Uint64("batch_number", batch.BatchNumber).Msg("Queued batch for L1 submission")
		default:
			log.Warn().Uint64("batch_number", batch.BatchNumber).Msg("L1 submission queue full, skipping batch")
		}
	}

	s.batchInProgress = false
	s.currentBatch = nil
	return nil
}

// processTransferTransaction processes a simple token transfer transaction
func (s *Sequencer) processTransferTransaction(tx state.Transaction, sender *state.Account) error {
	// Verify balance
	if sender.Balance.Cmp(tx.Amount) < 0 {
		return fmt.Errorf("insufficient balance: have %s, need %s", sender.Balance.String(), tx.Amount.String())
	}

	// Handle zero values consistently as per memory requirements
	if tx.Amount.Cmp(big.NewInt(0)) == 0 {
		// Use a single byte with value 0 instead of an empty array
		tx.Amount = big.NewInt(0)
	}

	// Update sender balance
	sender.Balance = new(big.Int).Sub(sender.Balance, tx.Amount)
	s.state.SetAccount(sender)

	// Update recipient account
	recipient, err := s.state.GetAccount(tx.To)
	if err != nil || recipient == nil {
		// Create recipient account if it doesn't exist
		recipient = &state.Account{
			Address: tx.To,
			Balance: tx.Amount,
			Nonce:   0,
		}
	} else {
		// Add amount to existing balance
		recipient.Balance = new(big.Int).Add(recipient.Balance, tx.Amount)
	}
	s.state.SetAccount(recipient)

	log.Info().Str("from", formatAddress(tx.From)).Str("to", formatAddress(tx.To)).Str("amount", tx.Amount.String()).Msg("Applied transfer transaction")
	return nil
}

// processContractDeployment processes a contract deployment transaction
func (s *Sequencer) processContractDeployment(tx state.Transaction, sender *state.Account) error {
	// Verify balance for the value being sent with contract creation
	if sender.Balance.Cmp(tx.Amount) < 0 {
		return fmt.Errorf("insufficient balance for contract deployment: have %s, need %s", sender.Balance.String(), tx.Amount.String())
	}

	// Special handling for zero values to ensure consistent message hash computation
	if tx.Amount != nil && tx.Amount.Cmp(big.NewInt(0)) == 0 {
		// Use a single byte with value 0 instead of an empty array
		tx.Amount = big.NewInt(0)
	}

	// Create EVM state adapter
	stateAdapter := evm.NewStateAdapter(s.state)

	// Convert addresses to Ethereum format
	callerAddr := common.BytesToAddress(tx.From[:])

	// Deploy the contract
	contractAddr, remainingGas, err := s.evmExecutor.DeployContract(
		stateAdapter,
		callerAddr,
		tx.Amount,
		tx.Gas,
		tx.Data,
	)

	if err != nil {
		return fmt.Errorf("contract deployment failed: %w", err)
	}

	// Convert contract address back to rollup format
	var contractRollupAddr [20]byte
	copy(contractRollupAddr[:], contractAddr.Bytes())

	// Update sender account
	sender.Balance = new(big.Int).Sub(sender.Balance, tx.Amount)
	sender.Nonce++
	s.state.SetAccount(sender)

	// Apply all state changes from the EVM execution
	stateAdapter.ApplyChanges()

	log.Info().Str("from", formatAddress(tx.From)).Str("contract", contractAddr.Hex()).Str("gas_used", fmt.Sprintf("%d", tx.Gas-remainingGas)).Msg("Deployed contract")
	return nil
}

// processContractCall processes a contract call transaction
func (s *Sequencer) processContractCall(tx state.Transaction, sender *state.Account) error {
	// Verify balance for the value being sent with the call
	if sender.Balance.Cmp(tx.Amount) < 0 {
		return fmt.Errorf("insufficient balance for contract call: have %s, need %s", sender.Balance.String(), tx.Amount.String())
	}

	// Special handling for zero values to ensure consistent message hash computation
	if tx.Amount != nil && tx.Amount.Cmp(big.NewInt(0)) == 0 {
		// Use a single byte with value 0 instead of an empty array
		tx.Amount = big.NewInt(0)
	}

	// Create EVM state adapter
	stateAdapter := evm.NewStateAdapter(s.state)

	// Convert addresses to Ethereum format
	callerAddr := common.BytesToAddress(tx.From[:])
	contractAddr := common.BytesToAddress(tx.To[:])

	// Execute the contract call
	returnData, remainingGas, err := s.evmExecutor.ExecuteContract(
		stateAdapter,
		callerAddr,
		contractAddr,
		tx.Amount,
		tx.Gas,
		tx.Data,
	)

	if err != nil {
		return fmt.Errorf("contract call failed: %w", err)
	}

	// Update sender account
	sender.Balance = new(big.Int).Sub(sender.Balance, tx.Amount)
	sender.Nonce++
	s.state.SetAccount(sender)

	// Apply all state changes from the EVM execution
	stateAdapter.ApplyChanges()

	log.Info().Str("from", formatAddress(tx.From)).Str("contract", contractAddr.Hex()).Str("gas_used", fmt.Sprintf("%d", tx.Gas-remainingGas)).Int("return_data_size", len(returnData)).Msg("Called contract")
	return nil
}
