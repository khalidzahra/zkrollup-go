package l1

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"zkrollup/pkg/l1/contracts"
	"zkrollup/pkg/state"
)

// Client represents an Ethereum L1 client for the ZK-Rollup
type Client struct {
	ethClient      *ethclient.Client
	rollupContract *contracts.ZKRollup
	privateKey     *ecdsa.PrivateKey
	address        common.Address
	chainID        *big.Int
}

// Config represents the configuration for the L1 client
type Config struct {
	EthereumRPC     string
	ChainID         int64
	ContractAddress string
	PrivateKey      string
}

// NewClient creates a new L1 client
func NewClient(config *Config) (*Client, error) {
	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(config.EthereumRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}

	// Load private key
	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	// Get account address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Load rollup contract if address is provided
	var rollupContract *contracts.ZKRollup
	if config.ContractAddress != "" {
		contractAddress := common.HexToAddress(config.ContractAddress)
		rollupContract, err = contracts.NewZKRollup(contractAddress, ethClient)
		if err != nil {
			return nil, fmt.Errorf("failed to load rollup contract: %v", err)
		}
	}

	return &Client{
		ethClient:      ethClient,
		rollupContract: rollupContract,
		privateKey:     privateKey,
		address:        address,
		chainID:        big.NewInt(config.ChainID),
	}, nil
}

// DeployContract deploys the ZK-Rollup contract to L1
func (c *Client) DeployContract(ctx context.Context) (common.Address, error) {
	auth, err := c.getTransactOpts(ctx)
	if err != nil {
		return common.Address{}, err
	}

	// Deploy contract using the safe deployment function
	address, tx, err := contracts.DeployZKRollupSafe(auth, c.ethClient)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to deploy contract: %v", err)
	}

	log.Info().Str("tx_hash", tx.Hash().Hex()).Str("contract_address", address.Hex()).Msg("Deployed ZK-Rollup contract")
	return address, nil
}

// SubmitBatch submits a batch to the L1 contract
func (c *Client) SubmitBatch(ctx context.Context, batch *state.Batch, proof []byte) error {
	if c.rollupContract == nil {
		return fmt.Errorf("rollup contract not initialized")
	}

	auth, err := c.getTransactOpts(ctx)
	if err != nil {
		return err
	}

	// Convert batch to contract format
	batchNumber := big.NewInt(int64(batch.BatchNumber))
	stateRoot := common.BytesToHash(batch.StateRoot[:])
	txHashes := make([][32]byte, len(batch.Transactions))
	
	for i, tx := range batch.Transactions {
		txHashes[i] = tx.Hash()
	}

	// Submit batch to L1
	tx, err := c.rollupContract.SubmitBatch(auth, batchNumber, stateRoot, txHashes, proof)
	if err != nil {
		return fmt.Errorf("failed to submit batch: %v", err)
	}

	log.Info().Str("tx_hash", tx.Hash().Hex()).Uint64("batch_number", batch.BatchNumber).Msg("Submitted batch to L1")
	return nil
}

// VerifyBatch verifies a batch on L1
func (c *Client) VerifyBatch(ctx context.Context, batchNumber uint64) (bool, error) {
	if c.rollupContract == nil {
		return false, fmt.Errorf("rollup contract not initialized")
	}

	// Call the verify method
	verified, err := c.rollupContract.VerifyBatch(&bind.CallOpts{
		Context: ctx,
	}, big.NewInt(int64(batchNumber)))

	if err != nil {
		return false, fmt.Errorf("failed to verify batch: %v", err)
	}

	return verified, nil
}

// getTransactOpts creates transaction options for sending transactions
func (c *Client) getTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // No ether transfer
	auth.GasLimit = uint64(3000000) // Gas limit
	auth.GasPrice = gasPrice

	return auth, nil
}
