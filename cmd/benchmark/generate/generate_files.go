package main

import (
	"log"
	"os"
	zkCrypto "zkrollup/pkg/crypto"
)

func main() {
	// Generate key pairs
	prover, err := zkCrypto.NewProver()
	if err != nil {
		log.Fatalf("Failed to create prover: %v", err)
	}

	{
		f, err := os.Create("zkrollup.vk")
		if err != nil {
			log.Fatalf("Failed to create verifying key file: %v", err)
		}
		_, err = prover.VerifyingKey.WriteRawTo(f)
		if err != nil {
			log.Fatalf("Failed to write verifying key: %v", err)
		}
	}
	{
		f, err := os.Create("zkrollup.pk")
		if err != nil {
			log.Fatalf("Failed to create proving key file: %v", err)
		}
		_, err = prover.ProvingKey.WriteRawTo(f)
		if err != nil {
			log.Fatalf("Failed to write proving key: %v", err)
		}
	}

	{
		f, err := os.Create("contract.sol")
		if err != nil {
			log.Fatalf("Failed to create contract file: %v", err)
		}
		err = prover.VerifyingKey.ExportSolidity(f)
		if err != nil {
			log.Fatalf("Failed to export verifying key: %v", err)
		}
	}
}
