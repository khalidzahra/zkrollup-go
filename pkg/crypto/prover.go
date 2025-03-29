package crypto

import (
	"bytes"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

type Prover struct {
	provingKey   groth16.ProvingKey
	verifyingKey groth16.VerifyingKey
}

func NewProver() (*Prover, error) {
	var circuit TransactionCircuit

	// Compile circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %v", err)
	}

	// Setup proving and verifying keys
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, fmt.Errorf("failed to setup keys: %v", err)
	}

	return &Prover{
		provingKey:   pk,
		verifyingKey: vk,
	}, nil
}

func (p *Prover) GenerateProof(w *TransactionCircuit) ([]byte, error) {
	// Create witness
	witness, err := frontend.NewWitness(w, ecc.BN254.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %v", err)
	}

	// Compile circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &TransactionCircuit{})
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %v", err)
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, p.provingKey, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	// Serialize proof
	var buf bytes.Buffer
	if _, err := proof.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to serialize proof: %v", err)
	}

	return buf.Bytes(), nil
}

func (p *Prover) VerifyProof(proofBytes []byte, w *TransactionCircuit) (bool, error) {
	// Create public witness
	witness, err := frontend.NewWitness(w, ecc.BN254.ScalarField())
	if err != nil {
		return false, fmt.Errorf("failed to create witness: %v", err)
	}

	// Get public witness
	publicWitness, err := witness.Public()
	if err != nil {
		return false, fmt.Errorf("failed to get public witness: %v", err)
	}

	// Deserialize proof
	proof := groth16.NewProof(ecc.BN254)
	if _, err := proof.ReadFrom(bytes.NewReader(proofBytes)); err != nil {
		return false, fmt.Errorf("failed to deserialize proof: %v", err)
	}

	// Verify proof
	if err := groth16.Verify(proof, p.verifyingKey, publicWitness); err != nil {
		return false, nil
	}

	return true, nil
}
