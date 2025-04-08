package sequencer

import (
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/state"
)

// submitBatchesToL1 processes batches from the l1SubmitChan and submits them to L1
func (s *Sequencer) submitBatchesToL1() {
	// Set up ticker for periodic batch submission
	submitPeriod := time.Duration(s.config.L1BatchSubmitPeriod) * time.Second
	ticker := time.NewTicker(submitPeriod)
	defer ticker.Stop()

	log.Info().Int("period_seconds", s.config.L1BatchSubmitPeriod).Msg("Starting L1 batch submission process")

	for {
		select {
		case <-s.ctx.Done():
			log.Info().Msg("Stopping L1 batch submission process")
			return

		case batch := <-s.l1SubmitChan:
			// Process the batch and submit to L1
			if err := s.submitBatchToL1(batch); err != nil {
				log.Error().Err(err).Uint64("batch_number", batch.BatchNumber).Msg("Failed to submit batch to L1")
			} else {
				log.Info().Uint64("batch_number", batch.BatchNumber).Msg("Successfully submitted batch to L1")
			}

		case <-ticker.C:
			log.Debug().Msg("Checking for pending batches to submit to L1")
		}
	}
}

// submitBatchToL1 submits a single batch to L1
func (s *Sequencer) submitBatchToL1(batch state.Batch) error {
	if s.l1Client == nil {
		return nil
	}

	log.Info().Uint64("batch_number", batch.BatchNumber).Msg("Submitting batch to L1")

	// Generate a ZK proof for the batch
	var proof []byte
	if s.config.ProofGeneration {
		proof = batch.Proof
		log.Info().Uint64("batch_number", batch.BatchNumber).Msg("Generated proof for batch")
	} else {
		// If proof generation is disabled, use a dummy proof
		proof = []byte("dummy_proof")
		log.Warn().Uint64("batch_number", batch.BatchNumber).Msg("Using dummy proof for batch (proof generation disabled)")
	}

	// Submit the batch to L1
	err := s.l1Client.SubmitBatch(s.ctx, &batch, proof)
	if err != nil {
		return err
	}

	return nil
}
