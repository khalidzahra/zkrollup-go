package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"zkrollup/pkg/core"
	"zkrollup/pkg/sequencer"
)

func main() {
	config := core.DefaultConfig()

	// Get port from environment variable or use default
	port, err := strconv.Atoi(os.Getenv("SEQUENCER_PORT"))
	if err != nil {
		log.Fatalf("Failed to parse sequencer port, using default port: %v", err)
		port = config.SequencerPort
	}

	// Get bootstrap peers from environment variable
	var bootstrapPeers []string
	if peers := os.Getenv("BOOTSTRAP_PEERS"); peers != "" {
		bootstrapPeers = strings.Split(peers, ",")
	}

	// Check if this node is a leader
	isLeader := os.Getenv("IS_LEADER") == "true"

	// Initialize sequencer
	seq, err := sequencer.NewSequencer(config, port, bootstrapPeers, isLeader)
	if err != nil {
		log.Fatalf("Failed to create sequencer: %v", err)
	}

	if err := seq.Start(); err != nil {
		log.Fatalf("Failed to start sequencer: %v", err)
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	seq.Stop()
}
