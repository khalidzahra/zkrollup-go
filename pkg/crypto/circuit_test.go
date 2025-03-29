package crypto

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/signature/eddsa"
	"github.com/consensys/gnark/test"
)

func TestTransactionCircuit(t *testing.T) {
	assert := test.NewAssert(t)

	// Create circuit
	var circuit TransactionCircuit
	witness := &TransactionCircuit{
		FromPubKey: eddsa.PublicKey{
			A: struct {
				X frontend.Variable
				Y frontend.Variable
			}{
				X: 1,
				Y: 2,
			},
		},
		ToPubKey: eddsa.PublicKey{
			A: struct {
				X frontend.Variable
				Y frontend.Variable
			}{
				X: 3,
				Y: 4,
			},
		},
		Amount:       100,
		Nonce:        1,
		OldStateRoot: 123456,
		NewStateRoot: 789012,
		FromBalance:  1000,
		ToBalance:    500,
		MerkleProof:  []frontend.Variable{1, 2, 3, 4},
	}

	// Generate proof
	assert.SolvingSucceeded(&circuit, witness, test.WithCurves(ecc.BN254))
}
