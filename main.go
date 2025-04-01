package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"zkrollup/pkg/core"
	"zkrollup/pkg/sequencer"
)

func main() {
	config := core.DefaultConfig()

	// Initialize sequencer
	seq, err := sequencer.NewSequencer(config)
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
