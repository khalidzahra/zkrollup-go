package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

// This script interacts with the deployed SimpleStorage contract
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

	// Set contract address and RPC URL
	contractAddress := "0x583fd70c0b66a6db9073c28ee35255a4c3b187c0"
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

	// Create ABI for SimpleStorage contract
	simpleStorageABI := `[{"inputs":[{"internalType":"uint256","name":"x","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"increment","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"x","type":"uint256"}],"name":"set","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"storedData","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
	parsedABI, err := abi.JSON(strings.NewReader(simpleStorageABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// Call the 'set' function with value 42
	setValue := big.NewInt(42)
	fmt.Printf("Setting value to: %s\n", setValue.String())

	// Pack the function call data
	data, err := parsedABI.Pack("set", setValue)
	if err != nil {
		log.Fatalf("Failed to pack data: %v", err)
	}

	// Create message to sign for the 'set' function
	messageToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		address.Hex(),
		contractAddress,
		"0",                           // amount as string (zero value) - must be decimal string
		strconv.FormatUint(nonce, 10), // nonce as string
		"0x"+hex.EncodeToString(data), // data
	)

	// Sign the message
	messageHash := crypto.Keccak256Hash([]byte(messageToSign))
	signature, err := crypto.Sign(messageHash.Bytes(), privKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// Create transaction parameters for the 'set' function
	txParams := map[string]interface{}{
		"from":      address.Hex(),
		"to":        contractAddress,
		"amount":    "0", // Use decimal string for amount (base 10) as expected by ParseAmount
		"nonce":     float64(nonce),
		"gas":       float64(100000),
		"data":      "0x" + hex.EncodeToString(data),
		"signature": "0x" + hex.EncodeToString(signature), // Add proper signature
		"type":      float64(1),                           // Contract call
	}

	// Send transaction
	txHash, err := sendTransaction(rpcURL, txParams)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}
	fmt.Printf("Transaction sent successfully! Hash: %s\n", txHash)

	// Increment nonce for next transaction
	nonce++

	// Now call the 'get' function to retrieve the value
	data, err = parsedABI.Pack("get")
	if err != nil {
		log.Fatalf("Failed to pack data: %v", err)
	}

	// Create message to sign for the 'get' function
	messageToSign = fmt.Sprintf("%s:%s:%s:%s:%s",
		address.Hex(),
		contractAddress,
		"0",                           // amount as string (zero value) - must be decimal string
		strconv.FormatUint(nonce, 10), // nonce as string
		"0x"+hex.EncodeToString(data), // data
	)

	// Sign the message
	messageHash = crypto.Keccak256Hash([]byte(messageToSign))
	signature, err = crypto.Sign(messageHash.Bytes(), privKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// Create transaction parameters for the 'get' function
	txParams = map[string]interface{}{
		"from":      address.Hex(),
		"to":        contractAddress,
		"amount":    "0", // Use decimal string for amount (base 10) as expected by ParseAmount
		"nonce":     float64(nonce),
		"gas":       float64(100000),
		"data":      "0x" + hex.EncodeToString(data),
		"signature": "0x" + hex.EncodeToString(signature), // Add proper signature
		"type":      float64(2),                           // Contract call (view function)
	}

	// Send transaction
	txHash, err = sendTransaction(rpcURL, txParams)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}
	fmt.Printf("Get transaction sent successfully! Hash: %s\n", txHash)
}

// getNonce gets the current nonce for an address
func getNonce(rpcURL, address string) (uint64, error) {
	// Create JSON-RPC request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "rollup_getNonce",
		"params":  []string{address},
		"id":      1,
	}

	// Send request
	respBody, err := sendJSONRPCRequest(rpcURL, req)
	if err != nil {
		return 0, err
	}

	// Print the raw response for debugging
	fmt.Printf("Raw getNonce response: %s\n", string(respBody))

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
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
	// Create JSON-RPC request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "rollup_sendTransaction",
		"params":  []map[string]interface{}{txParams},
		"id":      1,
	}

	// Send request
	respBody, err := sendJSONRPCRequest(rpcURL, req)
	if err != nil {
		return "", err
	}

	// Print the raw response for debugging
	fmt.Printf("Raw sendTransaction response: %s\n", string(respBody))

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
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

// sendJSONRPCRequest sends a JSON-RPC request to the specified URL
func sendJSONRPCRequest(url string, req map[string]interface{}) ([]byte, error) {
	// Convert request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(reqJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return respBody, nil
}
