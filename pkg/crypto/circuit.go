package crypto

import (
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// TransactionCircuit defines the ZK-SNARK circuit for transaction verification
type TransactionCircuit struct {
	// Public inputs
	FromPubKey eddsa.PublicKey   `gnark:",public"`
	ToPubKey   eddsa.PublicKey   `gnark:",public"`
	Amount     frontend.Variable `gnark:",public"`
	Nonce      frontend.Variable `gnark:",public"`

	// Private inputs
	Signature eddsa.Signature   `gnark:",secret"`
	Balance   frontend.Variable `gnark:",secret"`

	// Internal state
	api frontend.API
}

// Define implements the circuit logic for transaction verification
func (c *TransactionCircuit) Define(api frontend.API) error {
	c.api = api

	// 1. Verify signature
	msg := computeMessageHash(api, c.FromPubKey, c.ToPubKey, c.Amount, c.Nonce)
	hashFunc, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Create Edwards curve for signature verification
	curve, err := twistededwards.NewEdCurve(api, tedwards.BN254)
	if err != nil {
		return err
	}

	// Verify signature using EdDSA
	if err := eddsa.Verify(curve, c.Signature, msg, c.FromPubKey, &hashFunc); err != nil {
		return err
	}

	// 2. Verify sender has sufficient balance
	api.AssertIsLessOrEqual(c.Amount, c.Balance)

	return nil
}

func computeMessageHash(api frontend.API, from, to eddsa.PublicKey, amount, nonce frontend.Variable) frontend.Variable {
	hash, _ := mimc.NewMiMC(api)

	hash.Write(from.A.X)
	hash.Write(from.A.Y)
	hash.Write(to.A.X)
	hash.Write(to.A.Y)
	hash.Write(amount)
	hash.Write(nonce)

	return hash.Sum()
}
