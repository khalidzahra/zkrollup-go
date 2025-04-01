package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	gnarkEddsa "github.com/consensys/gnark/std/signature/eddsa"
	"github.com/consensys/gnark/test"
)

func TestTransactionCircuit(t *testing.T) {
	// Create test witness
	witness, err := createTestWitness()
	if err != nil {
		t.Fatal(err)
	}

	// Create empty circuit
	var circuit TransactionCircuit

	// Run test
	assert := test.NewAssert(t)
	assert.ProverSucceeded(&circuit, witness, test.WithCurves(ecc.BN254), test.WithBackends(backend.GROTH16))
}

func createTestWitness() (*TransactionCircuit, error) {
	// Generate key pairs
	senderPrivKey, err := eddsa.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	senderPubKey := &senderPrivKey.PublicKey

	// Convert public key to frontend format
	fromPubKey, err := publicKeyToFrontend(senderPubKey)
	if err != nil {
		return nil, err
	}

	// Transaction parameters
	amount := big.NewInt(100)
	nonce := big.NewInt(1)
	balance := big.NewInt(1000)

	// Prepare message to sign
	msg := prepareMessage(senderPubKey, senderPubKey, amount, nonce)

	// Sign with MIMC hash
	hFunc := hash.MIMC_BN254.New()
	sigBytes, err := senderPrivKey.Sign(msg, hFunc)
	if err != nil {
		fmt.Println("Error signing message")
		fmt.Println(err)
		return nil, err
	}

	// Decode signature
	sig := new(eddsa.Signature)
	if _, err = sig.SetBytes(sigBytes); err != nil {
		return nil, err
	}

	// Convert signature to frontend format
	gnarkSig, err := signatureToFrontend(sig)
	if err != nil {
		return nil, err
	}

	return &TransactionCircuit{
		FromPubKey: *fromPubKey,
		ToPubKey:   *fromPubKey,
		Amount:     amount.String(),
		Nonce:      nonce.String(),
		Signature:  *gnarkSig,
		Balance:    balance.String(),
	}, nil
}

func publicKeyToFrontend(pubKey *eddsa.PublicKey) (*gnarkEddsa.PublicKey, error) {
	var x, y fr.Element
	x.Set(&pubKey.A.X)
	y.Set(&pubKey.A.Y)

	return &gnarkEddsa.PublicKey{
		A: twistededwards.Point{
			X: x.String(),
			Y: y.String(),
		},
	}, nil
}

func signatureToFrontend(sig *eddsa.Signature) (*gnarkEddsa.Signature, error) {
	var rx, ry, s fr.Element

	// Convert R point (X,Y)
	rx.Set(&sig.R.X)
	ry.Set(&sig.R.Y)

	// Convert S scalar
	sBigInt := new(big.Int).SetBytes(sig.S[:])
	s.SetBigInt(sBigInt)

	return &gnarkEddsa.Signature{
		R: twistededwards.Point{
			X: rx.String(),
			Y: ry.String(),
		},
		S: s.String(),
	}, nil
}

func prepareMessage(fromPubKey, toPubKey *eddsa.PublicKey, amount, nonce *big.Int) []byte {
	// Use MIMC hash to match circuit's message computation
	hFunc := hash.MIMC_BN254.New()

	// Write public key coordinates
	fromX := fromPubKey.A.X.Bytes()
	fromY := fromPubKey.A.Y.Bytes()
	toX := toPubKey.A.X.Bytes()
	toY := toPubKey.A.Y.Bytes()

	hFunc.Write(fromX[:])
	hFunc.Write(fromY[:])
	hFunc.Write(toX[:])
	hFunc.Write(toY[:])

	// Convert amount and nonce to fr.Element
	var amountFr, nonceFr fr.Element
	amountFr.SetBigInt(amount)
	nonceFr.SetBigInt(nonce)

	// Write transaction details as fr.Element bytes
	amountBytes := amountFr.Bytes()
	nonceBytes := nonceFr.Bytes()
	hFunc.Write(amountBytes[:])
	hFunc.Write(nonceBytes[:])

	return hFunc.Sum(nil)
}
