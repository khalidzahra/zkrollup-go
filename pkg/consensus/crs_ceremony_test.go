package consensus

import (
	"testing"
)

func TestCRSCeremonyStateMachine(t *testing.T) {
	participants := []string{"nodeA", "nodeB", "nodeC"}
	crsSize := 3

	initialCRS, _, err := GenerateInitialCRS(crsSize)
	if err != nil {
		t.Fatalf("Failed to generate initial CRS: %v", err)
	}

	state := &CRSCeremonyState{
		EpochNumber:     1,
		Participants:    participants,
		CurrentStep:     0,
		IntermediateCRS: initialCRS,
	}

	prevCRS := state.IntermediateCRS

	// Simulate round-robin contributions
	for i, node := range participants {
		if !state.CheckTurn(node) {
			t.Fatalf("Node %s should have turn at step %d", node, i)
		}
		newCRS, _, proof, err := AddContribution(state.IntermediateCRS)
		if err != nil {
			t.Fatalf("AddContribution failed for %s: %v", node, err)
		}
		if !VerifyContributionProof(prevCRS, newCRS, proof) {
			t.Fatalf("Proof verification failed for %s's contribution", node)
		}
		state.IntermediateCRS = newCRS
		prevCRS = newCRS
		state.CurrentStep++
	}

	if state.CurrentStep != len(participants) {
		t.Fatalf("Ceremony did not complete: got step %d, want %d", state.CurrentStep, len(participants))
	}
}
