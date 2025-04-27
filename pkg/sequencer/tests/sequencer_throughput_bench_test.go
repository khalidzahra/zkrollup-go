package tests

import (
	"encoding/csv"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"zkrollup/pkg/core"
	"zkrollup/pkg/sequencer"
	"zkrollup/pkg/state"
)

func generateRandomAddress() [20]byte {
	var addr [20]byte
	rand.Read(addr[:])
	return addr
}

// generateTransactions generates n transactions, ensuring each sender's nonce starts at 1 (not 0) and increments by 1.
func generateTransactions(n int, senderCount int) []state.Transaction {
	txs := make([]state.Transaction, 0, n)
	senders := make([][20]byte, senderCount)
	for i := 0; i < senderCount; i++ {
		senders[i] = generateRandomAddress()
	}
	// Distribute transactions as evenly as possible among senders
	txsPerSender := n / senderCount
	remaining := n % senderCount
	for i, sender := range senders {
		numTx := txsPerSender
		if i < remaining {
			numTx++
		}
		for nonce := uint64(1); nonce <= uint64(numTx); nonce++ {
			txs = append(txs, state.Transaction{
				Type:   state.TxTypeTransfer,
				From:   sender,
				To:     generateRandomAddress(),
				Amount: big.NewInt(int64(rand.Intn(1000) + 1)),
				Nonce:  nonce,
				Data:   nil,
				Gas:    21000,
			})
		}
	}
	return txs
}

func TestSequencerThroughput(t *testing.T) {
	transactionCounts := []int{20000, 40000, 80000, 100000, 150000, 200000}
	results := make([]struct {
		Transactions int
		Duration     float64 // seconds
		Throughput   float64 // tx/sec
	}, 0, len(transactionCounts))

	config := core.DefaultConfig()
	seq, err := sequencer.NewSequencer(config, config.SequencerPort, nil, true)
	if err != nil {
		t.Fatalf("failed to initialize sequencer: %v", err)
	}

	senderCount := 1000 // Number of unique senders

	for _, txCount := range transactionCounts {
		txs := generateTransactions(txCount, senderCount)

		t.Logf("Processing %d transactions...", txCount)
		start := time.Now()
		for i := range txs {
			err := seq.AddTransaction(txs[i])
			if err != nil {
				t.Fatalf("failed to add transaction: %v", err)
			}
		}
		duration := time.Since(start).Seconds()
		throughput := float64(txCount) / duration
		t.Logf("Processed %d txs in %.2fs (%.0f tx/sec)", txCount, duration, throughput)

		results = append(results, struct {
			Transactions int
			Duration     float64
			Throughput   float64
		}{
			Transactions: txCount,
			Duration:     duration,
			Throughput:   throughput,
		})
	}

	csvFile := "sequencer_throughput.csv"
	f, err := os.Create(csvFile)
	if err != nil {
		t.Fatalf("failed to create csv: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"transactions", "duration_seconds", "throughput_tps"})
	for _, r := range results {
		w.Write([]string{
			fmt.Sprintf("%d", r.Transactions),
			fmt.Sprintf("%.6f", r.Duration),
			fmt.Sprintf("%.2f", r.Throughput),
		})
	}
	t.Logf("Sequencer throughput results written to %s", csvFile)
}
