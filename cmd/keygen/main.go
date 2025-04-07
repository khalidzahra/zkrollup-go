package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	
	// Parse flags
	flag.Parse()
	
	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate private key")
	}
	
	// Get the private key in bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)
	
	// Convert to hex string (without 0x prefix)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	
	// Get the public address
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	
	// Print the results
	fmt.Println("Generated new Ethereum key")
	fmt.Println("---------------------------")
	fmt.Printf("Private Key: %s\n", privateKeyHex)
	fmt.Printf("Address:     %s\n", address.Hex())
	fmt.Println("\nTo use this key with the ZK-Rollup EVM client:")
	fmt.Printf("./zkrollup-evm -key %s -action deploy -contract ./contracts/examples/SimpleStorage.sol\n", privateKeyHex)
}
