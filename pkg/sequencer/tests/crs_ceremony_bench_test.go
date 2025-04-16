package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"fmt"
	"os"
	"testing"
	"time"
	"zkrollup/contracts/bindings"
	"zkrollup/pkg/sequencer/crsutils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func RunFullCRSCeremony(t *testing.T, n int) {
	t.Helper()

	// Generate participants and fund them
	type participant struct {
		key  *ecdsa.PrivateKey
		auth *bind.TransactOpts
		addr common.Address
	}
	participants := make([]participant, n)
	alloc := core.GenesisAlloc{}
	for i := 0; i < n; i++ {
		key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		participants[i] = participant{
			key:  key,
			auth: NewTestAuth(t, key),
			addr: addr,
		}
		alloc[addr] = core.GenesisAccount{Balance: big.NewInt(1e18)}
	}

	// Setup deployer and fund
	keyDeployer, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	deployerAddr := crypto.PubkeyToAddress(keyDeployer.PublicKey)
	alloc[deployerAddr] = core.GenesisAccount{Balance: big.NewInt(1e18)}

	backend := backends.NewSimulatedBackend(alloc, 16_000_000)
	authDeployer := NewTestAuth(t, keyDeployer)
	maxParticipants := big.NewInt(int64(n))
	roundDuration := big.NewInt(1000) // arbitrary
	addr, _, crsManager, err := bindings.DeployCRSManager(authDeployer, backend, maxParticipants, roundDuration)
	require.NoError(t, err)
	backend.Commit()
	_ = addr // silence unused warning

	// Register first participant and contribute initial CRS
	_, err = crsManager.Register(participants[0].auth)
	require.NoError(t, err)
	backend.Commit()

	const numCRSPoints = 4
	initialCRS, err := crsutils.GenerateRandomCRS(numCRSPoints)
	require.NoError(t, err)
	_, err = crsManager.ContributeCRS(participants[0].auth, initialCRS)
	require.NoError(t, err)
	backend.Commit()

	// Register remaining participants
	for i := 1; i < n; i++ {
		_, err := crsManager.Register(participants[i].auth)
		require.NoError(t, err)
		backend.Commit()
	}

	// Contribute CRS for remaining participants
	crs := initialCRS
	for i := 1; i < n; i++ {
		contribution, _, err := crsutils.TransformCRSWithRandomScalar(crs)
		require.NoError(t, err)
		_, err = crsManager.ContributeCRS(participants[i].auth, contribution)
		require.NoError(t, err)
		backend.Commit()
		crs = contribution
	}

	// Finalize
	_, err = crsManager.FinalizeCRS(participants[n-1].auth)
	require.NoError(t, err)
	backend.Commit()
}

func TestCRSCeremonyPerformance(t *testing.T) {
	participantCounts := []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
	results := make([]struct {
		Participants int
		Duration     float64 // seconds
	}, 0, len(participantCounts))

	for _, n := range participantCounts {
		t.Logf("Running CRS ceremony with %d participants...", n)
		start := time.Now()
		RunFullCRSCeremony(t, n)
		elapsed := time.Since(start).Seconds()
		results = append(results, struct {
			Participants int
			Duration     float64
		}{n, elapsed})
		t.Logf("CRS ceremony with %d participants took %.2fs", n, elapsed)
	}

	f, err := os.Create("crs_ceremony_performance.csv")
	require.NoError(t, err)
	defer f.Close()
	fmt.Fprintln(f, "participants,duration_seconds")
	for _, r := range results {
		fmt.Fprintf(f, "%d,%.4f\n", r.Participants, r.Duration)
	}
}
