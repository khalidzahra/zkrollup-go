package crypto

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// Prover handles proof generation and verification
type Prover struct {
	provingKey   groth16.ProvingKey
	verifyingKey groth16.VerifyingKey
	r1cs         constraint.ConstraintSystem
}

// NewProver creates a new prover with the necessary keys
func NewProver() (*Prover, error) {
	// Create a new circuit
	var circuit TransactionCircuit

	// Compile the circuit
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %v", err)
	}

	// Setup the proving and verifying keys
	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		return nil, fmt.Errorf("failed to setup keys: %v", err)
	}

	return &Prover{
		provingKey:   pk,
		verifyingKey: vk,
		r1cs:         r1cs,
	}, nil
}

// CreateWitness creates a witness for the circuit
func (p *Prover) CreateWitness(
	fromPubKey eddsa.PublicKey,
	toPubKey eddsa.PublicKey,
	amount *big.Int,
	nonce *big.Int,
	signature eddsa.Signature,
	balance *big.Int,
) (*TransactionCircuit, error) {
	amountVar := frontend.Variable(amount.String())
	balanceVar := frontend.Variable(balance.String())
	nonceVar := frontend.Variable(nonce.String())

	// Create circuit witness
	witness := &TransactionCircuit{
		FromPubKey: fromPubKey,
		ToPubKey:   toPubKey,
		Amount:     amountVar,
		Nonce:      nonceVar,
		Signature:  signature,
		Balance:    balanceVar,
	}

	return witness, nil
}

// GenerateProof generates a proof for the given witness
func (p *Prover) GenerateProof(w *TransactionCircuit) ([]byte, []byte, error) {
	// Create witness
	witness, err := frontend.NewWitness(w, ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create witness: %v", err)
	}

	// Generate proof
	proof, err := groth16.Prove(p.r1cs, p.provingKey, witness)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	// Serialize the proof
	var buf bytes.Buffer
	_, err = proof.WriteTo(&buf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize proof: %v", err)
	}

	// Get public witness
	publicWitness, err := witness.Public()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get public witness: %v", err)
	}
	// Serialize the public witness
	publicWitnessBuf := new(bytes.Buffer)
	_, err = publicWitness.WriteTo(publicWitnessBuf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize public witness: %v", err)
	}

	return buf.Bytes(), publicWitnessBuf.Bytes(), nil
}

// VerifyProof verifies a proof against the given witness
func (p *Prover) VerifyProof(proofBytes, publicWitnessBytes []byte) (bool, error) {
	// Create public witness
	publicWitness, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return false, fmt.Errorf("failed to create witness: %v", err)
	}

	publicWitness.ReadFrom(bytes.NewReader(publicWitnessBytes))

	// Deserialize the proof
	proof := groth16.NewProof(ecc.BN254)
	_, err = proof.ReadFrom(bytes.NewReader(proofBytes))
	if err != nil {
		return false, fmt.Errorf("failed to deserialize proof: %v", err)
	}

	// Verify the proof
	err = groth16.Verify(proof, p.verifyingKey, publicWitness)
	if err != nil {
		return false, fmt.Errorf("proof verification failed: %v", err)
	}

	return true, nil
}
