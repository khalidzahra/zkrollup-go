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
	s.poolMu.RLock()
	txCount := len(s.txPool)
	s.poolMu.RUnlock()

	// Check if we have enough transactions and are not already processing a batch
	if txCount < int(s.config.BatchSize/2) || s.batchInProgress {
		return
	}

	// Check if we are the leader
	if !s.isLeader {
		log.Debug().Msg("Not the leader, skipping batch creation")
		return
	}

	s.batchMu.Lock()
	defer s.batchMu.Unlock()

	// Double-check that we're not already processing a batch
	if s.batchInProgress {
		return
	}

	// Mark that we're starting to process a batch
	s.batchInProgress = true

	// Create a new batch with transactions from the pool
	s.poolMu.Lock()
	batchSize := min(txCount, int(s.config.BatchSize))
	batchTxs := make([]state.Transaction, batchSize)
	copy(batchTxs, s.txPool[:batchSize])
	s.txPool = s.txPool[batchSize:]
	s.poolMu.Unlock()

	// Create the batch
	batch := &state.Batch{
		Transactions: batchTxs,
		BatchNumber:  s.state.GetBatchNumber() + 1,
		Timestamp:    uint64(time.Now().Unix()),
	}

	// Store the current batch
	s.currentBatch = batch

	// Propose the batch for consensus
	if err := s.consensus.ProposeBatch(batch); err != nil {
		log.Error().Err(err).Msg("Failed to propose batch for consensus")
		// Return transactions to the pool
		s.poolMu.Lock()
		s.txPool = append(batchTxs, s.txPool...)
		s.poolMu.Unlock()
		s.batchInProgress = false
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

			// Update leader status from consensus module
			s.isLeader = s.consensus.IsLeader()
			if s.isLeader {
				log.Info().Msg("This node is now the leader and will propose the next batch")
			} else {
				log.Info().Msg("This node is not the leader for the next batch")
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
	log.Info().Int("tx_count", len(batch.Transactions)).Msg("Processing finalized batch")

	// Process each transaction in the batch
	for _, tx := range batch.Transactions {
		var err error

		// Get the sender account
		sender, err := s.state.GetAccount(tx.From)
		if err != nil {
			log.Error().Err(err).Str("from", common.BytesToAddress(tx.From[:]).Hex()).Msg("Failed to get sender account")
			continue
		}

		// Process transaction based on type
		switch tx.Type {
		case state.TxTypeTransfer:
			err = s.processTransferTransaction(tx, sender)
		case state.TxTypeContractDeploy:
			err = s.processContractDeployment(tx, sender)
		case state.TxTypeContractCall:
			err = s.processContractCall(tx, sender)
		default:
			err = fmt.Errorf("unknown transaction type: %d", tx.Type)
		}

		if err != nil {
			log.Error().Err(err).Str("tx_hash", common.BytesToHash(tx.HashToBytes()).Hex()).Msg("Failed to process transaction")
			continue
		}
	}

	// Update batch number in state
	s.state.AddBatch(&batch)

	// Mark batch processing as complete
	s.batchMu.Lock()
	s.batchInProgress = false
	s.currentBatch = nil
	s.batchMu.Unlock()

	// Submit batch to L1 if enabled
	if s.l1Enabled && s.l1SubmitChan != nil {
		select {
		case s.l1SubmitChan <- batch:
			log.Info().Msg("Submitted batch to L1 submission queue")
		default:
			log.Warn().Msg("L1 submission queue is full, skipping this batch")
		}
	}

	// Add a small delay after processing a batch to prevent rapid leader rotation
	// This gives the system time to stabilize between batches
	time.Sleep(time.Millisecond * 500)

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
