package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// This script deploys a contract to our simple RPC server
// It ensures consistent handling of zero values and nonce formats
// as required by the ZK-Rollup implementation

func main() {
	// Parse private key
	privateKeyHex := "7478e3b73c7f4741dbc94a39dcca55778dab9fcd5eec42e71ad602f2bf67e15f"
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}

	// Derive address from private key
	privKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("Using address: %s\n", address.Hex())

	// Read contract bytecode
	bytecode, err := ioutil.ReadFile("./contracts/examples/SimpleStorage.bin")
	if err != nil {
		log.Fatalf("Failed to read contract file: %v", err)
	}
	bytecodeStr := strings.TrimSpace(string(bytecode))

	// Set RPC URL
	rpcURL := "http://localhost:8081"

	// Get nonce
	nonce, err := getNonce(rpcURL, address.Hex())
	if err != nil {
		log.Printf("Failed to get nonce, using 0: %v", err)
		nonce = 0
	}
	// Increment nonce to ensure it's greater than the account's current nonce
	nonce++
	fmt.Printf("Using nonce: %d\n", nonce)

	// Create transaction message hash for signing
	messageToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		address.Hex(),
		"0x",                          // to address (empty for contract deployment)
		"0",                           // amount as string (zero value) - must be decimal string
		strconv.FormatUint(nonce, 10), // nonce as string
		"0x"+bytecodeStr,              // data
	)

	// Sign the message
	messageHash := crypto.Keccak256Hash([]byte(messageToSign))
	signature, err := crypto.Sign(messageHash.Bytes(), privKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// Create transaction parameters
	txParams := map[string]interface{}{
		"from":      address.Hex(),
		"to":        "0x",             // Empty address for contract deployment
		"amount":    "0",              // Use decimal string for amount (base 10) as expected by ParseAmount
		"gas":       float64(1000000), // Use float64 for gas as expected by the server
		"nonce":     float64(nonce),   // Use float64 for nonce as expected by the server
		"data":      "0x" + bytecodeStr,
		"signature": "0x" + hex.EncodeToString(signature), // Add signature
		"type":      float64(0),                           // Default transaction type
	}

	// Send transaction
	txHash, err := sendTransaction(rpcURL, txParams)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent successfully! Hash: %s\n", txHash)

	// Generate contract address
	contractAddr := generateContractAddress(address.Hex(), nonce)
	fmt.Printf("Contract deployed at address: %s\n", contractAddr)

	// Wait for the transaction to be processed
	fmt.Println("Waiting for transaction to be processed...")
	time.Sleep(5 * time.Second)

	// Get the contract code with retries
	var code string
	for i := 0; i < 5; i++ {
		code, err = getCode(rpcURL, contractAddr)
		if err != nil {
			log.Printf("Failed to get contract code (attempt %d/5): %v", i+1, err)
		} else if code != "0x" {
			// Contract code is available
			break
		}

		fmt.Printf("Contract code not available yet, retrying in 2 seconds (attempt %d/5)...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if code == "0x" {
		fmt.Println("Contract deployment transaction was accepted but code is not available yet.")
		fmt.Println("This might be because the transaction is still being processed or hasn't been included in a batch.")
	} else {
		fmt.Printf("Contract code: %s\n", code)
	}
}

// getNonce gets the current nonce for an address
func getNonce(rpcURL, address string) (uint64, error) {
	// Create request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "rollup_getNonce",
		"params":  []string{address},
		"id":      1,
	}

	// Send request
	respData, err := sendJSONRPCRequest(rpcURL, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Print the raw response for debugging
	fmt.Printf("Raw getNonce response: %s\n", string(respData))

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if errObj, ok := resp["error"]; ok {
		return 0, fmt.Errorf("RPC error: %v", errObj)
	}

	// Get result
	result := resp["result"]

	// Try to parse as a direct number first
	if nonceFloat, ok := result.(float64); ok {
		return uint64(nonceFloat), nil
	}

	// Try to parse as a string (hex)
	if nonceStr, ok := result.(string); ok {
		// Remove "0x" prefix if present
		if strings.HasPrefix(nonceStr, "0x") {
			nonceStr = nonceStr[2:]
		}

		// Parse hex string
		nonceInt, err := strconv.ParseUint(nonceStr, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid nonce hex format: %v", err)
		}
		return nonceInt, nil
	}

	// Try to parse as a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		nonceFloat, ok := resultMap["nonce"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid nonce format in map")
		}
		return uint64(nonceFloat), nil
	}

	return 0, fmt.Errorf("invalid result format: %v", result)
}

// sendTransaction sends a transaction to the rollup
func sendTransaction(rpcURL string, txParams map[string]interface{}) (string, error) {
	// Create request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "rollup_sendTransaction",
		"params":  []interface{}{txParams},
		"id":      1,
	}

	// Send request
	respData, err := sendJSONRPCRequest(rpcURL, req)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	// Print the raw response for debugging
	fmt.Printf("Raw sendTransaction response: %s\n", string(respData))

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if errObj, ok := resp["error"]; ok {
		return "", fmt.Errorf("RPC error: %v", errObj)
	}

	// Get result
	result := resp["result"]

	// Try to parse as a string directly (txHash)
	if txHash, ok := result.(string); ok {
		return txHash, nil
	}

	// Try to parse as a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		txHash, ok := resultMap["txHash"].(string)
		if !ok {
			return "", fmt.Errorf("invalid transaction hash format in map")
		}
		return txHash, nil
	}

	return "", fmt.Errorf("invalid result format: %v", result)
}

// getCode gets the code at an address
func getCode(rpcURL, address string) (string, error) {
	// Create request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "rollup_getCode",
		"params":  []string{address},
		"id":      1,
	}

	// Send request
	respData, err := sendJSONRPCRequest(rpcURL, req)
	if err != nil {
		return "", fmt.Errorf("failed to get code: %w", err)
	}

	// Print the raw response for debugging
	fmt.Printf("Raw getCode response: %s\n", string(respData))

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if errObj, ok := resp["error"]; ok {
		return "", fmt.Errorf("RPC error: %v", errObj)
	}

	// Get result
	result := resp["result"]

	// Try to parse as a string directly
	if codeStr, ok := result.(string); ok {
		return codeStr, nil
	}

	// Try to parse as a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		codeStr, ok := resultMap["code"].(string)
		if !ok {
			return "", fmt.Errorf("invalid code format in map")
		}
		return codeStr, nil
	}

	return "", fmt.Errorf("invalid result format: %v", result)
}

// sendJSONRPCRequest sends a JSON-RPC request to the specified URL
func sendJSONRPCRequest(url string, req map[string]interface{}) ([]byte, error) {
	// Marshal request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return respData, nil
}

// generateContractAddress generates a contract address from the sender address and nonce
func generateContractAddress(sender string, nonce uint64) string {
	// Remove 0x prefix if present
	if strings.HasPrefix(sender, "0x") {
		sender = sender[2:]
	}

	// Parse sender address
	senderBytes, _ := hex.DecodeString(sender)
	var senderAddr [20]byte
	copy(senderAddr[:], senderBytes)

	// Generate contract address (simplified version)
	// In Ethereum, this would be RLP([sender, nonce])
	data := append(senderAddr[:], byte(nonce))
	contractAddr := crypto.Keccak256(data)

	// Return the last 20 bytes as the contract address
	return fmt.Sprintf("0x%x", contractAddr[12:])
}
