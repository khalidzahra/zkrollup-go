package crypto

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
	"github.com/consensys/gnark/std/algebra/twistededwards"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
)

// TransactionCircuit defines the ZK-SNARK circuit for transaction verification
type TransactionCircuit struct {
	// Public inputs
	FromPubKey   eddsa.PublicKey   `gnark:",public"`
	ToPubKey     eddsa.PublicKey   `gnark:",public"`
	Amount       frontend.Variable  `gnark:",public"`
	Nonce        frontend.Variable  `gnark:",public"`
	OldStateRoot frontend.Variable  `gnark:",public"`
	NewStateRoot frontend.Variable  `gnark:",public"`

	// Private inputs
	Signature    eddsa.Signature   `gnark:",private"`
	FromBalance  frontend.Variable `gnark:",private"`
	ToBalance    frontend.Variable `gnark:",private"`
	MerkleProof  []frontend.Variable `gnark:",private"`

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
	api.AssertIsLessOrEqual(c.Amount, c.FromBalance)

	// 3. Verify state transition
	newFromBalance := api.Sub(c.FromBalance, c.Amount)
	newToBalance := api.Add(c.ToBalance, c.Amount)

	// 4. Verify merkle proof for old state
	verifyMerkleProof(api, c.MerkleProof, c.OldStateRoot)

	// 5. Verify new state root
	computedNewRoot := computeNewStateRoot(api, c.MerkleProof, newFromBalance, newToBalance)
	api.AssertIsEqual(computedNewRoot, c.NewStateRoot)

	return nil
}

func computeMessageHash(api frontend.API, from eddsa.PublicKey, to eddsa.PublicKey, amount, nonce frontend.Variable) frontend.Variable {
	hash, _ := mimc.NewMiMC(api)
	
	hash.Write(from.A.X)
	hash.Write(from.A.Y)
	hash.Write(to.A.X)
	hash.Write(to.A.Y)
	hash.Write(amount)
	hash.Write(nonce)
	
	return hash.Sum()
}

func verifyMerkleProof(api frontend.API, proof []frontend.Variable, root frontend.Variable) {
	// Simplified merkle proof verification
	current := proof[0]
	hash, _ := mimc.NewMiMC(api)

	for i := 1; i < len(proof); i++ {
		hash.Reset()
		hash.Write(current)
		hash.Write(proof[i])
		current = hash.Sum()
	}

	api.AssertIsEqual(current, root)
}

func computeNewStateRoot(api frontend.API, proof []frontend.Variable, newFromBalance, newToBalance frontend.Variable) frontend.Variable {
	// Simplified new state root computation
	hash, _ := mimc.NewMiMC(api)
	hash.Write(newFromBalance)
	hash.Write(newToBalance)
	current := hash.Sum()

	for i := 1; i < len(proof); i++ {
		hash.Reset()
		hash.Write(current)
		hash.Write(proof[i])
		current = hash.Sum()
	}

	return current
}
