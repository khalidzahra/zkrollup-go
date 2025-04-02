package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"zkrollup/pkg/core"
	"zkrollup/pkg/sequencer"
)

func main() {
	config := core.DefaultConfig()

	port, err := strconv.Atoi(os.Getenv("SEQUENCER_PORT"))
	if err != nil {
		log.Fatalf("Failed to parse sequencer port, using default port: %v", err)
		port = config.SequencerPort
	}

	// Initialize sequencer with P2P port
	seq, err := sequencer.NewSequencer(config, port)
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
