package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	ed "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark-crypto/signature"
	"github.com/consensys/gnark-crypto/signature/eddsa"
	"github.com/consensys/gnark/backend"
	gnarkEddsa "github.com/consensys/gnark/std/signature/eddsa"
	"github.com/consensys/gnark/test"
)

func TestCircuitConstraints(t *testing.T) {
	// Create a simple circuit
	circuit := &TransactionCircuit{}

	// Compile the circuit
	assert := test.NewAssert(t)
	assert.CheckCircuit(circuit, test.WithCurves(ecc.BN254), test.WithBackends(backend.GROTH16))
}

func TestProverSetup(t *testing.T) {
	// Create a new prover
	prover, err := NewProver()
	if err != nil {
		t.Fatalf("failed to create prover: %v", err)
	}

	// Check that the prover was created with valid keys
	if prover.provingKey == nil {
		t.Fatal("proving key is nil")
	}

	if prover.verifyingKey == nil {
		t.Fatal("verifying key is nil")
	}
}

// TestEndToEndProofGeneration tests the complete proof generation and verification process
func TestEndToEndProofGeneration(t *testing.T) {
	hFunc := hash.MIMC_BN254.New()

	senderPubKey, senderSigner, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate sender key pair: %v", err)
	}
	receiverPubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate receiver key pair: %v", err)
	}

	gSenderPubKey := gnarkEddsa.PublicKey{}
	gReceiverPubKey := gnarkEddsa.PublicKey{}

	gSenderPubKey.Assign(ed.BN254, senderPubKey.Bytes()[:32])
	gReceiverPubKey.Assign(ed.BN254, receiverPubKey.Bytes()[:32])

	amount := big.NewInt(100)
	balance := big.NewInt(200)
	nonce := big.NewInt(200)

	hFunc.Write(getBytes(fmt.Sprint(gReceiverPubKey.A.X)))
	hFunc.Write(getBytes(fmt.Sprint(gReceiverPubKey.A.Y)))
	hFunc.Write(amount.Bytes())
	hFunc.Write(balance.Bytes())
	hFunc.Write(nonce.Bytes())

	nonceHash := hFunc.Sum(nil)

	signature, err := senderSigner.Sign(nonceHash, hFunc)
	if err != nil {
		t.Fatalf("failed to sign nonce: %v", err)
	}

	gSig := gnarkEddsa.Signature{}

	gSig.Assign(ed.BN254, signature)

	prover, err := NewProver()
	if err != nil {
		t.Fatalf("failed to create prover: %v", err)
	}
	witness, err := prover.CreateWitness(
		gSenderPubKey,
		gReceiverPubKey,
		amount,
		nonce,
		gSig,
		balance,
	)
	if err != nil {
		t.Fatalf("failed to create witness: %v", err)
	}
	proof, pubWitness, err := prover.GenerateProof(witness)
	if err != nil {
		t.Fatalf("failed to generate proof: %v", err)
	}

	valid, err := prover.VerifyProof(proof, pubWitness)
	if err != nil {
		t.Fatalf("failed to verify proof: %v", err)
	}

	if !valid {
		t.Fatal("proof verification failed")
	} else {
		t.Log("proof verification succeeded")
	}
}

func GenerateKeyPair() (signature.PublicKey, signature.Signer, error) {
	privateKey, err := eddsa.New(ed.BN254, rand.Reader)
	if err != nil {
		fmt.Println("failed to create a key pair. error:", err)
		return nil, nil, err
	}
	publicKey := privateKey.Public()

	return publicKey, privateKey, nil
}

func getBytes(str string) []byte {
	trimmed := strings.Trim(str, "[]")
	parts := strings.Fields(trimmed) // Splits on whitespace
	bytes := make([]byte, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		bytes[i] = byte(val)
	}
	return bytes
}
