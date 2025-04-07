package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/rs/zerolog/log"

	"zkrollup/pkg/p2p"
	"zkrollup/pkg/state"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 9100, "Port to run the client on")
	peerAddr := flag.String("peer", "", "Address of a sequencer node to connect to")
	flag.Parse()

	if *peerAddr == "" {
		log.Fatal().Msg("Please provide a peer address using the -peer flag")
	}

	// Create a P2P node for the client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := p2p.NewNode(ctx, *port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create P2P node")
	}
	defer node.Close()

	// Set up protocol handlers for the client
	node.SetupProtocols(&p2p.ProtocolHandlers{
		OnTransaction: func(tx *state.Transaction) error {
			log.Info().Msg("Received transaction acknowledgement")
			return nil
		},
		OnBatch: func(batch *state.Batch) error {
			log.Info().Msg("Received batch update")
			return nil
		},
		OnConsensus: func(msg []byte) error {
			log.Info().Msg("Received consensus message")
			return nil
		},
	})

	// Connect to the sequencer
	log.Info().Str("peer", *peerAddr).Msg("Connecting to sequencer")

	// Try connecting a few times with a delay between attempts
	connected := false
	for attempts := 0; attempts < 5; attempts++ {
		if err := node.Connect(ctx, *peerAddr); err != nil {
			log.Warn().Err(err).Int("attempt", attempts+1).Msg("Failed to connect to sequencer, retrying...")
			time.Sleep(time.Second * 2)
			continue
		}
		connected = true
		break
	}

	if !connected {
		log.Fatal().Msg("Failed to connect to sequencer after multiple attempts")
	}

	log.Info().Msg("Successfully connected to sequencer")

	// Generate some test accounts
	accounts := generateTestAccounts(5)
	log.Info().Int("count", len(accounts)).Msg("Generated test accounts")

	// Display account information
	for i, acc := range accounts {
		fmt.Printf("Account %d: Address: %x, Balance: %s\n", i, acc.Address, acc.Balance.String())
	}

	// Send test transactions
	for i := 0; i < 10; i++ {
		// Create a random transaction
		tx := createRandomTransaction(accounts)

		// Ensure proper handling of zero values for consistent message hash computation
		if tx.Amount.Sign() == 0 {
			log.Info().Msg("Transaction has zero amount - ensuring proper formatting for consistent hash computation")
		}

		// Ensure consistent nonce format between keygen, circuit, and transaction processing
		nonceStr := fmt.Sprintf("%d", tx.Nonce)
		log.Info().Msg(fmt.Sprintf("Using nonce string format '%s' for consistent hash computation", nonceStr))

		// Broadcast the transaction
		log.Info().
			Str("from", fmt.Sprintf("%x", tx.From)).
			Str("to", fmt.Sprintf("%x", tx.To)).
			Str("amount", tx.Amount.String()).
			Str("nonce_str", nonceStr). // Log nonce as string for consistency with circuit
			Msg("Sending transaction")

		if err := node.BroadcastTransaction(ctx, &tx); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast transaction")
			continue
		}

		log.Info().Msg("Transaction sent successfully")

		// Wait a bit between transactions
		time.Sleep(time.Second * 2)
	}

	// Keep the client running to receive responses
	fmt.Println("Client is running. Press Ctrl+C to exit.")
	select {}
}

// generateTestAccounts creates test accounts with random addresses and balances
func generateTestAccounts(count int) []*state.Account {
	accounts := make([]*state.Account, count)

	for i := 0; i < count; i++ {
		var addr [20]byte
		rand.Read(addr[:])

		// Generate a random balance between 100 and 1000
		max := big.NewInt(1000)
		min := big.NewInt(100)
		diff := big.NewInt(0).Sub(max, min)
		randInt, _ := rand.Int(rand.Reader, diff)
		balance := big.NewInt(0).Add(min, randInt)

		accounts[i] = &state.Account{
			Address: addr,
			Balance: balance,
			Nonce:   0,
		}
	}

	return accounts
}

// createRandomTransaction creates a random transaction between two accounts
func createRandomTransaction(accounts []*state.Account) state.Transaction {
	// Select random from and to accounts
	fromIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(accounts))))
	toIdx := fromIdx
	for toIdx.Cmp(fromIdx) == 0 {
		toIdx, _ = rand.Int(rand.Reader, big.NewInt(int64(len(accounts))))
	}

	from := accounts[fromIdx.Int64()]
	to := accounts[toIdx.Int64()]

	// Generate a random amount, occasionally using zero to test zero-value handling
	var amount *big.Int
	zeroChance, _ := rand.Int(rand.Reader, big.NewInt(10))
	if zeroChance.Cmp(big.NewInt(2)) < 0 { // 20% chance of zero amount
		amount = big.NewInt(0)
		fmt.Printf("Creating a zero-value transaction to test special zero handling\n")
	} else {
		max := big.NewInt(0).Set(from.Balance)
		if max.Cmp(big.NewInt(50)) > 0 {
			max = big.NewInt(50)
		}
		min := big.NewInt(1)

		// Ensure we have a valid range
		if max.Cmp(min) <= 0 {
			// If balance is too low, just use 1
			amount = big.NewInt(1)
		} else {
			diff := big.NewInt(0).Sub(max, min)
			randInt, _ := rand.Int(rand.Reader, diff)
			amount = big.NewInt(0).Add(min, randInt)
		}
	}

	// Create the transaction
	tx := state.Transaction{
		From:   from.Address,
		To:     to.Address,
		Amount: amount,
		Nonce:  from.Nonce,
	}

	// Update the account nonce
	from.Nonce++

	// Reduce sender's balance to track it locally (to avoid insufficient balance errors)
	from.Balance = big.NewInt(0).Sub(from.Balance, amount)

	// Increase receiver's balance
	to.Balance = big.NewInt(0).Add(to.Balance, amount)

	return tx
}
