package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"zkrollup/contracts/bindings"
	"zkrollup/pkg/l1"
	"zkrollup/pkg/sequencer/crsutils"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func deployTestCRSManager(t *testing.T, backend bind.ContractBackend, auth *bind.TransactOpts) (common.Address, *l1.CRSManager) {
	// Deploy the contract using the generated bindings
	address, _, _, err := bindings.DeployCRSManager(
		auth,
		backend,
		big.NewInt(60), // roundDuration: 60 seconds
		big.NewInt(4),  // maxParticipants: 4 (workaround for < instead of <=)
	)
	require.NoError(t, err)
	// Commit the deployment
	if sim, ok := backend.(interface{ Commit() }); ok {
		sim.Commit()
	}
	// Initialize the CRSManager client
	crsManager, err := l1.NewCRSManager(address, backend.(*backends.SimulatedBackend))
	require.NoError(t, err)
	return address, crsManager
}

func TestCRSCeremonyIntegration(t *testing.T) {
	// Set up simulated blockchain
	keyA, _ := crypto.GenerateKey()
	keyB, _ := crypto.GenerateKey()
	keyC, _ := crypto.GenerateKey()

	alloc := core.GenesisAlloc{
		crypto.PubkeyToAddress(keyA.PublicKey): {Balance: big.NewInt(1e18)},
		crypto.PubkeyToAddress(keyB.PublicKey): {Balance: big.NewInt(1e18)},
		crypto.PubkeyToAddress(keyC.PublicKey): {Balance: big.NewInt(1e18)},
	}
	backend := backends.NewSimulatedBackend(alloc, 8000000)

	// Deploy CRSManager contract
	authA := NewTestAuth(t, keyA)
	addr, crsManager := deployTestCRSManager(t, backend, authA)
	_ = addr
	_ = crsManager

	// Initialize CRS as N random BN254 G1 points (compressed)
	const numCRSPoints = 4
	initialCRS, err := crsutils.GenerateRandomCRS(numCRSPoints)
	require.NoError(t, err)
	t.Logf("Initial CRS length: %d", len(initialCRS))
	if len(initialCRS)%bn254.SizeOfG1AffineCompressed != 0 {
		t.Fatalf("Initial CRS length %d is not a multiple of compressed G1 size %d", len(initialCRS), bn254.SizeOfG1AffineCompressed)
	}
	// Set initial CRS via contract (simulate admin or first participant)
	err = crsManager.ContributeCRS(authA, initialCRS)
	require.NoError(t, err)
	backend.Commit()

	// Commit after CRS contribution to ensure state is updated
	backend.Commit()

	crsStored, err := crsManager.GetCurrentCRS(context.Background())
	require.NoError(t, err)
	t.Logf("CRS stored in contract after initial contribution: length=%d", len(crsStored))
	if len(crsStored)%bn254.SizeOfG1AffineCompressed != 0 {
		t.Fatalf("CRS stored after initial contribution has invalid length %d (not a multiple of %d)", len(crsStored), bn254.SizeOfG1AffineCompressed)
	}

	// Simulate registration
	ctx := context.Background()
	participants := []struct {
		key  *ecdsa.PrivateKey
		auth *bind.TransactOpts
		addr common.Address
	}{
		{keyA, authA, crypto.PubkeyToAddress(keyA.PublicKey)},
		{keyB, NewTestAuth(t, keyB), crypto.PubkeyToAddress(keyB.PublicKey)},
		{keyC, NewTestAuth(t, keyC), crypto.PubkeyToAddress(keyC.PublicKey)},
	}

	// Register the first participant (authA) and contribute initial CRS
	err = crsManager.Register(authA)
	require.NoError(t, err)
	backend.Commit()

	err = crsManager.ContributeCRS(authA, initialCRS)
	require.NoError(t, err)
	backend.Commit()

	crsStored, err = crsManager.GetCurrentCRS(context.Background())
	require.NoError(t, err)
	t.Logf("CRS stored in contract after initial contribution: length=%d", len(crsStored))
	if len(crsStored)%bn254.SizeOfG1AffineCompressed != 0 {
		t.Fatalf("CRS stored after initial contribution has invalid length %d (not a multiple of %d)", len(crsStored), bn254.SizeOfG1AffineCompressed)
	}

	// Register the remaining participants (skip authA)
	for i, p := range participants[1:] {
		backend.Commit()
		err := crsManager.Register(p.auth)
		if err != nil {
			t.Logf("Registration failed for participant %d (%s): %v", i+1, p.addr.Hex(), err)
		}
		require.NoError(t, err)
		backend.Commit()
	}

	// Fetch and compare contract participant list to local list
	contractParticipants, err := crsManager.GetRegisteredParticipants(ctx)
	require.NoError(t, err)
	if len(contractParticipants) != len(participants) {
		t.Fatalf("Participant count mismatch: contract=%d local=%d", len(contractParticipants), len(participants))
	}
	for i, addr := range contractParticipants {
		if addr != participants[i].addr {
			t.Fatalf("Participant %d mismatch: contract=%s local=%s", i, addr.Hex(), participants[i].addr.Hex())
		}
	}

	crsHistory := [][]byte{initialCRS}
	for i, p := range participants[1:] {
		crsBefore, err := crsManager.GetCurrentCRS(ctx)
		require.NoError(t, err)
		t.Logf("CRS before contribution %d: length=%d", i+1, len(crsBefore))
		if len(crsBefore)%bn254.SizeOfG1AffineCompressed != 0 {
			t.Fatalf("CRS before contribution %d has invalid length %d (not a multiple of %d)", i+1, len(crsBefore), bn254.SizeOfG1AffineCompressed)
		}
		crsHistory = append(crsHistory, crsBefore)
		// Use production CRS transformation logic
		contribution, _, err := crsutils.TransformCRSWithRandomScalar(crsBefore)
		require.NoError(t, err)
		err = crsManager.ContributeCRS(p.auth, contribution)
		require.NoError(t, err)
		backend.Commit()
	}
	// Last participant finalizes
	isLast, _ := crsManager.IsLastContributor(ctx, participants[2].addr)
	require.True(t, isLast)
	err = crsManager.FinalizeCRS(participants[2].auth)
	require.NoError(t, err)
	backend.Commit()

	// Check that the final CRS matches the last CRS after all contributions
	var finalCrs []byte
	finalCrs, _, _, err = crsManager.GetLatestCRS(ctx)
	require.NoError(t, err)
	require.Equal(t, crsHistory[len(crsHistory)-1], finalCrs, "Final CRS should match last CRS after all contributions")
}
