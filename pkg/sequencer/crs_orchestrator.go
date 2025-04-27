package sequencer

import (
	"context"
	"log"
	"time"
	"zkrollup/pkg/l1"
	"zkrollup/pkg/sequencer/crsutils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type CRSOrchestrator struct {
	CRSManager   *l1.CRSManager
	Auth         *bind.TransactOpts
	PollInterval time.Duration
	CRSSize      int // bytes
}

func NewCRSOrchestrator(crsManager *l1.CRSManager, auth *bind.TransactOpts, pollInterval time.Duration, crsSize int) *CRSOrchestrator {
	return &CRSOrchestrator{
		CRSManager:   crsManager,
		Auth:         auth,
		PollInterval: pollInterval,
		CRSSize:      crsSize,
	}
}

// Start runs the orchestration loop for interactive onchain CRS
func (o *CRSOrchestrator) Start(ctx context.Context) {
	log.Println("CRS Orchestrator started")
	for {
		select {
		case <-ctx.Done():
			log.Println("CRS Orchestrator stopped")
			return
		case <-time.After(o.PollInterval):
			o.step(ctx)
		}
	}
}

// step performs one orchestration cycle
func (o *CRSOrchestrator) step(ctx context.Context) {
	isTurn, err := o.CRSManager.CheckTurn(ctx, o.Auth.From)
	if err != nil {
		log.Printf("Error checking turn: %v", err)
		return
	}
	if !isTurn {
		return
	}

	currentCRS, err := o.CRSManager.GetCurrentCRS(ctx)
	if err != nil {
		log.Printf("Error fetching current CRS: %v", err)
		return
	}

	// transformation (BN254 EC multiplication, compressed)
	myCRS, _, err := crsutils.TransformCRSWithRandomScalar(currentCRS)
	if err != nil {
		log.Printf("Error transforming CRS: %v", err)
		return
	}

	if err := o.CRSManager.ContributeCRS(o.Auth, myCRS); err != nil {
		log.Printf("Error submitting CRS: %v", err)
		return
	}
	log.Println("Successfully contributed CRS")

	isLast, err := o.CRSManager.IsLastContributor(ctx, o.Auth.From)
	if err == nil && isLast {
		if err := o.CRSManager.FinalizeCRS(o.Auth); err != nil {
			log.Printf("Error finalizing CRS: %v", err)
		} else {
			log.Println("CRS finalized!")
		}
	}
}
