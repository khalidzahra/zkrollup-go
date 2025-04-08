package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"zkrollup/pkg/l1"
)

func main() {
	// Parse command line flags
	privateKey := flag.String("privatekey", "", "Private key for Ethereum account (hex format without 0x prefix)")
	rpcURL := flag.String("rpc", "http://localhost:8545", "Ethereum RPC URL")
	chainID := flag.Int64("chainid", 1337, "Ethereum chain ID")
	flag.Parse()

	// Validate private key
	if *privateKey == "" {
		log.Fatal("Private key is required. Use -privatekey flag.")
	}

	// Create L1 client config
	config := &l1.Config{
		EthereumRPC: *rpcURL,
		ChainID:     *chainID,
		PrivateKey:  *privateKey,
	}

	// Create L1 client
	client, err := l1.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create L1 client: %v", err)
	}

	// Deploy ZK-Rollup contract
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("Deploying ZK-Rollup contract to L1...")
	address, err := client.DeployContract(ctx)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	}

	fmt.Printf("ZK-Rollup contract deployed at: %s\n", address.Hex())

	// Save contract address to environment file for easy loading
	envFile := ".env.l1"
	content := fmt.Sprintf(`# ZK-Rollup L1 Configuration
ETHEREUM_RPC=%s
CHAIN_ID=%d
CONTRACT_ADDRESS=%s
L1_PRIVATE_KEY=%s
L1_ENABLED=true
L1_BATCH_SUBMIT_PERIOD=300
`, *rpcURL, *chainID, address.Hex(), *privateKey)

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		log.Printf("Warning: Failed to write environment file: %v", err)
	} else {
		fmt.Printf("Environment configuration saved to %s\n", envFile)
		fmt.Println("To use this configuration, run:")
		fmt.Printf("  source %s && go run main.go\n", envFile)
	}
}
