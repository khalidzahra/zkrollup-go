package crypto

import (
	ed "github.com/consensys/gnark-crypto/ecc/twistededwards"
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
}

// Define implements the circuit logic for transaction verification
func (c *TransactionCircuit) Define(api frontend.API) error {
	api.AssertIsLessOrEqual(c.Amount, c.Balance)

	curve, err := twistededwards.NewEdCurve(api, ed.BN254)
	if err != nil {
		return err
	}

	mimc, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	mimc.Write(c.ToPubKey.A.X, c.ToPubKey.A.Y, c.Amount, c.Balance, c.Nonce)
	msgHash := mimc.Sum()

	mimc.Reset()
	return eddsa.Verify(curve, c.Signature, msgHash, c.FromPubKey, &mimc)
}
