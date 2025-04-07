package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"

	"zkrollup/pkg/sequencer"
	"zkrollup/pkg/state"
)

// Server represents the JSON-RPC server for the ZK-Rollup
type Server struct {
	sequencer *sequencer.Sequencer
	port      int
	server    *http.Server
	mu        sync.RWMutex
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewServer creates a new RPC server
func NewServer(seq *sequencer.Sequencer, port int) *Server {
	return &Server{
		sequencer: seq,
		port:      port,
	}
}

// Start starts the RPC server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRPC)

	addr := fmt.Sprintf(":%d", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Info().Int("port", s.port).Msg("Starting RPC server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("RPC server error")
		}
	}()

	return nil
}

// Stop stops the RPC server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		log.Info().Msg("Stopping RPC server")
		return s.server.Close()
	}
	return nil
}

// handleRPC handles JSON-RPC requests
func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &req, -32700, "Parse error")
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Process request
	switch req.Method {
	case "rollup_getNonce":
		s.handleGetNonce(w, &req)
	case "rollup_sendTransaction":
		s.handleSendTransaction(w, &req)
	case "rollup_getBalance":
		s.handleGetBalance(w, &req)
	case "rollup_getCode":
		s.handleGetCode(w, &req)
	default:
		writeError(w, &req, -32601, "Method not found")
	}
}

// handleGetNonce handles the rollup_getNonce method
func (s *Server) handleGetNonce(w http.ResponseWriter, req *JSONRPCRequest) {
	var params []string
	if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
		writeError(w, req, -32602, "Invalid params")
		return
	}

	// Parse address
	addrStr := params[0]
	if len(addrStr) < 2 || addrStr[:2] != "0x" {
		writeError(w, req, -32602, "Address must start with 0x")
		return
	}

	addr := common.HexToAddress(addrStr)
	var address [20]byte
	copy(address[:], addr.Bytes())

	// Get account from state
	account, err := s.sequencer.GetAccount(address)
	if err != nil {
		writeError(w, req, -32603, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Return nonce
	nonce := uint64(0)
	if account != nil {
		nonce = account.Nonce
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"nonce": nonce,
		},
		ID: req.ID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

// handleSendTransaction handles the rollup_sendTransaction method
func (s *Server) handleSendTransaction(w http.ResponseWriter, req *JSONRPCRequest) {
	var params []map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
		writeError(w, req, -32602, "Invalid params")
		return
	}

	txParams := params[0]

	// Parse transaction fields
	fromStr, ok := txParams["from"].(string)
	if !ok || len(fromStr) < 2 || fromStr[:2] != "0x" {
		writeError(w, req, -32602, "Invalid from address")
		return
	}

	toStr, ok := txParams["to"].(string)
	if !ok {
		// For contract deployment, to can be empty
		toStr = "0x0000000000000000000000000000000000000000"
	}

	amountStr, ok := txParams["amount"].(string)
	if !ok {
		writeError(w, req, -32602, "Invalid amount")
		return
	}

	nonceFloat, ok := txParams["nonce"].(float64)
	if !ok {
		writeError(w, req, -32602, "Invalid nonce")
		return
	}

	gasFloat, ok := txParams["gas"].(float64)
	if !ok {
		writeError(w, req, -32602, "Invalid gas")
		return
	}

	dataStr, ok := txParams["data"].(string)
	if !ok {
		writeError(w, req, -32602, "Invalid data")
		return
	}

	sigStr, ok := txParams["signature"].(string)
	if !ok {
		writeError(w, req, -32602, "Invalid signature")
		return
	}

	typeFloat, ok := txParams["type"].(float64)
	if !ok {
		writeError(w, req, -32602, "Invalid type")
		return
	}

	// Convert addresses
	from := common.HexToAddress(fromStr)
	to := common.HexToAddress(toStr)

	// Convert data and signature
	data := common.FromHex(dataStr)
	signature := common.FromHex(sigStr)

	// Create transaction
	tx := state.Transaction{
		Type:      state.TxType(uint8(typeFloat)),
		From:      [20]byte{},
		To:        [20]byte{},
		Amount:    nil, // Will be set below
		Nonce:     uint64(nonceFloat),
		Data:      data,
		Gas:       uint64(gasFloat),
		Signature: signature,
	}

	// Copy addresses
	copy(tx.From[:], from.Bytes())
	copy(tx.To[:], to.Bytes())

	// Parse amount
	tx.Amount = state.ParseAmount(amountStr)
	if tx.Amount == nil {
		writeError(w, req, -32602, "Invalid amount format")
		return
	}

	// Add transaction to sequencer
	if err := s.sequencer.AddTransaction(tx); err != nil {
		writeError(w, req, -32603, fmt.Sprintf("Failed to add transaction: %v", err))
		return
	}

	// Calculate transaction hash
	txHash := fmt.Sprintf("0x%x", state.CalculateTransactionHash(tx))

	// Return transaction hash
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"txHash": txHash,
		},
		ID: req.ID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

// handleGetBalance handles the rollup_getBalance method
func (s *Server) handleGetBalance(w http.ResponseWriter, req *JSONRPCRequest) {
	var params []string
	if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
		writeError(w, req, -32602, "Invalid params")
		return
	}

	// Parse address
	addrStr := params[0]
	if len(addrStr) < 2 || addrStr[:2] != "0x" {
		writeError(w, req, -32602, "Address must start with 0x")
		return
	}

	addr := common.HexToAddress(addrStr)
	var address [20]byte
	copy(address[:], addr.Bytes())

	// Get account from state
	account, err := s.sequencer.GetAccount(address)
	if err != nil {
		writeError(w, req, -32603, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Return balance
	balance := "0"
	if account != nil && account.Balance != nil {
		balance = account.Balance.String()
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"balance": balance,
		},
		ID: req.ID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

// handleGetCode handles the rollup_getCode method
func (s *Server) handleGetCode(w http.ResponseWriter, req *JSONRPCRequest) {
	var params []string
	if err := json.Unmarshal(req.Params, &params); err != nil || len(params) < 1 {
		writeError(w, req, -32602, "Invalid params")
		return
	}

	// Parse address
	addrStr := params[0]
	if len(addrStr) < 2 || addrStr[:2] != "0x" {
		writeError(w, req, -32602, "Address must start with 0x")
		return
	}

	addr := common.HexToAddress(addrStr)
	var address [20]byte
	copy(address[:], addr.Bytes())

	// Get code from state
	code, err := s.sequencer.GetCode(address)
	if err != nil {
		writeError(w, req, -32603, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Return code
	codeHex := fmt.Sprintf("0x%x", code)

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"code": codeHex,
		},
		ID: req.ID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

// writeError writes a JSON-RPC error response
func writeError(w http.ResponseWriter, req *JSONRPCRequest, code int, message string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: req.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}
