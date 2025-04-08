package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"zkrollup/pkg/core"
	"zkrollup/pkg/rpc"
	"zkrollup/pkg/sequencer"
)

func main() {
	config := core.DefaultConfig()

	// Get port from environment variable or use default
	port, err := strconv.Atoi(os.Getenv("SEQUENCER_PORT"))
	if err != nil {
		log.Printf("Failed to parse sequencer port, using default port: %v", err)
		port = config.SequencerPort
	}

	// Get RPC port from environment variable or use default
	rpcPort, err := strconv.Atoi(os.Getenv("RPC_PORT"))
	if err != nil {
		log.Printf("Failed to parse RPC port, using default port: %v", err)
		rpcPort = 8081 // Use port 8081 for the RPC server to match test script expectations
	}

	// Get bootstrap peers from environment variable
	var bootstrapPeers []string
	if peers := os.Getenv("BOOTSTRAP_PEERS"); peers != "" {
		bootstrapPeers = strings.Split(peers, ",")
	}

	// Check if this node is a leader
	isLeader := os.Getenv("IS_LEADER") == "true"

	// L1 integration configuration
	if l1Enabled := os.Getenv("L1_ENABLED"); l1Enabled == "true" {
		config.L1Enabled = true
		log.Printf("L1 integration enabled")

		// Get L1 configuration from environment variables
		if ethRPC := os.Getenv("ETHEREUM_RPC"); ethRPC != "" {
			config.EthereumRPC = ethRPC
		}

		if chainID := os.Getenv("CHAIN_ID"); chainID != "" {
			if id, err := strconv.ParseInt(chainID, 10, 64); err == nil {
				config.ChainID = id
			}
		}

		if contractAddr := os.Getenv("CONTRACT_ADDRESS"); contractAddr != "" {
			config.ContractAddress = contractAddr
		}

		if privateKey := os.Getenv("L1_PRIVATE_KEY"); privateKey != "" {
			config.L1PrivateKey = privateKey
		}

		if submitPeriod := os.Getenv("L1_BATCH_SUBMIT_PERIOD"); submitPeriod != "" {
			if period, err := strconv.Atoi(submitPeriod); err == nil {
				config.L1BatchSubmitPeriod = period
			}
		}
	}

	// Initialize sequencer
	seq, err := sequencer.NewSequencer(config, port, bootstrapPeers, isLeader)
	if err != nil {
		log.Fatalf("Failed to create sequencer: %v", err)
	}

	if err := seq.Start(); err != nil {
		log.Fatalf("Failed to start sequencer: %v", err)
	}

	// Initialize and start RPC server
	rpcServer := rpc.NewServer(seq, rpcPort)
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}

	log.Printf("ZK-Rollup node started with RPC server on port %d", rpcPort)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	rpcServer.Stop()
	seq.Stop()
}
