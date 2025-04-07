package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"zkrollup/pkg/state"
)

var (
	privateKey = flag.String("key", "", "Private key (hex format without 0x prefix)")
	rpcURL     = flag.String("rpc", "http://localhost:9000", "Rollup RPC URL")
	action     = flag.String("action", "deploy", "Action to perform: deploy, call")
	contractFile = flag.String("contract", "", "Contract bytecode file (for deploy) or address (for call)")
	method     = flag.String("method", "", "Method to call (for call action)")
	args       = flag.String("args", "", "Arguments for method call, comma separated")
	amount     = flag.String("amount", "0", "Amount to send with transaction")
	gas        = flag.Uint64("gas", 1000000, "Gas limit")
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	
	flag.Parse()
	
	if *privateKey == "" {
		log.Fatal().Msg("Private key is required")
	}
	
	// Parse private key
	privateKeyBytes, err := hex.DecodeString(*privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode private key")
	}
	
	// Derive address from private key
	privKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid private key")
	}
	
	address := crypto.PubkeyToAddress(privKey.PublicKey)
	
	log.Info().Str("address", address.Hex()).Msg("Using address")
	
	// Convert Ethereum address to rollup address format
	var rollupAddr [20]byte
	copy(rollupAddr[:], address.Bytes())
	
	// Parse amount
	amountValue := new(big.Int)
	amountValue, ok := amountValue.SetString(*amount, 10)
	if !ok {
		log.Fatal().Str("amount", *amount).Msg("Invalid amount")
	}
	
	// Create client
	client := NewRollupClient(*rpcURL)
	
	// Handle different actions
	switch *action {
	case "deploy":
		deployContract(client, rollupAddr, privateKeyBytes, privKey, amountValue)
	case "call":
		callContract(client, rollupAddr, privateKeyBytes, privKey, amountValue)
	default:
		log.Fatal().Str("action", *action).Msg("Unknown action")
	}
}

func deployContract(client *RollupClient, from [20]byte, privateKey []byte, privKey *ecdsa.PrivateKey, amount *big.Int) {
	if *contractFile == "" {
		log.Fatal().Msg("Contract file is required for deployment")
	}
	
	// Read contract bytecode
	bytecode, err := os.ReadFile(*contractFile)
	if err != nil {
		log.Fatal().Err(err).Str("file", *contractFile).Msg("Failed to read contract file")
	}
	
	// Get nonce
	nonce, err := client.GetNonce(from)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get nonce")
	}
	
	// Create transaction
	tx := state.Transaction{
		Type:   state.TxTypeContractDeploy,
		From:   from,
		To:     [20]byte{}, // Empty for contract deployment
		Amount: amount,
		Nonce:  nonce,
		Data:   bytecode,
		Gas:    *gas,
	}
	
	// Sign transaction
	signature, err := crypto.Sign(getTransactionHash(tx), privKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to sign transaction")
	}
	tx.Signature = signature
	
	// Send transaction
	err = client.SendTransaction(tx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to send transaction")
	}
	
	log.Info().Msg("Contract deployment transaction sent successfully")
}

func callContract(client *RollupClient, from [20]byte, privateKey []byte, privKey *ecdsa.PrivateKey, amount *big.Int) {
	if *contractFile == "" || *method == "" {
		log.Fatal().Msg("Contract address and method are required for contract call")
	}
	
	// Parse contract address
	contractAddr := common.HexToAddress(*contractFile)
	var to [20]byte
	copy(to[:], contractAddr.Bytes())
	
	// Get nonce
	nonce, err := client.GetNonce(from)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get nonce")
	}
	
	// Parse ABI and method arguments
	// This is a simplified implementation - in a real-world scenario, you'd need to parse the ABI
	methodSig := crypto.Keccak256([]byte(*method))[:4] // First 4 bytes of method signature
	
	// Parse arguments (simplified)
	var calldata []byte
	calldata = append(calldata, methodSig...)
	
	if *args != "" {
		// Very simplified argument handling - in reality, you'd need proper ABI encoding
		argsList := strings.Split(*args, ",")
		for _, arg := range argsList {
			// This is just a placeholder - real implementation would properly encode based on types
			calldata = append(calldata, common.Hex2Bytes(arg)...)
		}
	}
	
	// Create transaction
	tx := state.Transaction{
		Type:   state.TxTypeContractCall,
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
		Data:   calldata,
		Gas:    *gas,
	}
	
	// Sign transaction
	signature, err := crypto.Sign(getTransactionHash(tx), privKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to sign transaction")
	}
	tx.Signature = signature
	
	// Send transaction
	err = client.SendTransaction(tx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to send transaction")
	}
	
	log.Info().Msg("Contract call transaction sent successfully")
}

// getTransactionHash computes the hash of a transaction for signing
func getTransactionHash(tx state.Transaction) []byte {
	// This is a simplified implementation
	// In a real implementation, you'd use RLP encoding or a similar approach
	
	// Special handling for zero values to ensure consistent hash computation
	amountBytes := []byte{0}
	if tx.Amount.Sign() > 0 {
		amountBytes = tx.Amount.Bytes()
	}
	
	// Convert nonce to string for consistent hash computation
	nonceStr := fmt.Sprintf("%d", tx.Nonce)
	
	// Compute hash
	hash := crypto.Keccak256(
		tx.From[:],
		tx.To[:],
		amountBytes,
		[]byte(nonceStr),
		tx.Data,
	)
	
	return hash
}

// RollupClient is a client for interacting with the ZK-Rollup
type RollupClient struct {
	rpcURL string
	httpClient *http.Client
}

// NewRollupClient creates a new rollup client
func NewRollupClient(rpcURL string) *RollupClient {
	return &RollupClient{
		rpcURL: rpcURL,
		httpClient: &http.Client{},
	}
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetNonce gets the current nonce for an address
func (c *RollupClient) GetNonce(address [20]byte) (uint64, error) {
	// Convert address to hex string
	addrHex := fmt.Sprintf("0x%x", address)
	
	// Create RPC request
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "rollup_getNonce",
		Params:  []string{addrHex},
		ID:      1,
	}
	
	// Make the RPC call
	var resp struct {
		Nonce uint64 `json:"nonce"`
	}
	
	err := c.call(req, &resp)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get nonce from rollup, using 0 as fallback")
		return 0, nil
	}
	
	return resp.Nonce, nil
}

// SendTransaction sends a transaction to the rollup
func (c *RollupClient) SendTransaction(tx state.Transaction) error {
	// Log transaction details
	log.Info().
		Str("from", fmt.Sprintf("%x", tx.From)).
		Str("to", fmt.Sprintf("%x", tx.To)).
		Str("amount", tx.Amount.String()).
		Uint64("nonce", tx.Nonce).
		Uint64("gas", tx.Gas).
		Int("data_size", len(tx.Data)).
		Uint8("type", uint8(tx.Type)).
		Msg("Sending transaction")
	
	// Convert transaction to JSON-friendly format
	txParams := map[string]interface{}{
		"from":      fmt.Sprintf("0x%x", tx.From),
		"to":        fmt.Sprintf("0x%x", tx.To),
		"amount":    tx.Amount.String(),
		"nonce":     tx.Nonce,
		"gas":       tx.Gas,
		"data":      fmt.Sprintf("0x%x", tx.Data),
		"signature": fmt.Sprintf("0x%x", tx.Signature),
		"type":      uint8(tx.Type),
	}
	
	// Create RPC request
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "rollup_sendTransaction",
		Params:  []interface{}{txParams},
		ID:      1,
	}
	
	// Make the RPC call
	var resp struct {
		TxHash string `json:"txHash"`
	}
	
	err := c.call(req, &resp)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}
	
	log.Info().Str("txHash", resp.TxHash).Msg("Transaction sent successfully")
	return nil
}

// call makes a JSON-RPC call to the rollup node
func (c *RollupClient) call(req RPCRequest, result interface{}) error {
	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.rpcURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Make the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse response
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("failed to parse response: %w, body: %s", err, string(respBody))
	}
	
	// Check for RPC error
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %d - %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	
	// Unmarshal result
	if result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}
	
	return nil
}
