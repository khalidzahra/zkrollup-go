package consensus

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// PTauCeremonyState tracks the state of an ongoing Powers of Tau ceremony
// for generating SRS compatible with Circom and SnarkJS
type PTauCeremonyState struct {
	EpochNumber  int64
	Participants []string // Ordered list of participant NodeIDs
	CurrentStep  int      // Whose turn (index in Participants)
	PTauPath     string   // Path to the current ptau file
	PowerSize    int      // Power of 2 size of the ceremony (e.g., 12 for 2^12 constraints)
	Completed    bool
	Mutex        sync.Mutex
}

// PTauContributionMessage is sent between participants during the Powers of Tau ceremony
type PTauContributionMessage struct {
	EpochNumber   int64
	Step          int
	PTauFilePath  string
	PTauFileData  []byte // The actual PTau file data
	ContributorID string
	Signature     []byte
}

// NewPTauCeremonyState creates a new Powers of Tau ceremony state for an epoch
func NewPTauCeremonyState(epochNumber int64, participants []string, powerSize int, outputDir string) (*PTauCeremonyState, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Initialize the ceremony with a new ptau file
	ptauPath := filepath.Join(outputDir, fmt.Sprintf("pot_epoch%d_0.ptau", epochNumber))

	// Use snarkjs to start a new Powers of Tau ceremony
	cmd := exec.Command("npx", "snarkjs", "powersoftau", "new", "bn128", fmt.Sprintf("%d", powerSize), ptauPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Powers of Tau: %v\n%s", err, output)
	}

	return &PTauCeremonyState{
		EpochNumber:  epochNumber,
		Participants: participants,
		CurrentStep:  0,
		PTauPath:     ptauPath,
		PowerSize:    powerSize,
		Completed:    false,
	}, nil
}

// CheckTurn returns true if the given nodeID is the current contributor
func (s *PTauCeremonyState) CheckTurn(nodeID string) bool {
	// Make sure we don't try to access an index that's out of range
	if s.CurrentStep < 0 || s.CurrentStep >= len(s.Participants) {
		return false
	}
	return s.Participants[s.CurrentStep] == nodeID
}

// AddContribution adds a contribution to the Powers of Tau ceremony
func (s *PTauCeremonyState) AddContribution(contributorID, entropy string) (*PTauContributionMessage, error) {
	// Check if it's this node's turn to contribute
	if !s.CheckTurn(contributorID) {
		return nil, fmt.Errorf("not %s's turn to contribute", contributorID)
	}

	// Create the output ptau file path
	nextStep := s.CurrentStep + 1
	outputPath := filepath.Join(filepath.Dir(s.PTauPath), fmt.Sprintf("pot_epoch%d_%d.ptau", s.EpochNumber, nextStep))

	cmd := exec.Command("npx", "snarkjs", "powersoftau", "contribute",
		s.PTauPath, outputPath,
		"--name="+fmt.Sprintf("Contribution from %s", contributorID),
		"-v",
		"-e="+entropy)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to add contribution: %v\n%s", err, output)
	}

	// Update the state
	s.PTauPath = outputPath
	s.CurrentStep = nextStep

	// Read the PTau file data
	ptauData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PTau file: %v", err)
	}

	// Create the contribution message
	msg := &PTauContributionMessage{
		EpochNumber:   s.EpochNumber,
		Step:          s.CurrentStep - 1, // The step that was just completed
		ContributorID: contributorID,
		PTauFilePath:  outputPath,
		PTauFileData:  ptauData,
	}

	return msg, nil
}

// VerifyContribution verifies a contribution message from another participant
func (s *PTauCeremonyState) VerifyContribution(msg *PTauContributionMessage) error {
	// Verify that the message is for the current epoch
	if msg.EpochNumber != s.EpochNumber {
		return fmt.Errorf("unexpected epoch: got %d, expected %d", msg.EpochNumber, s.EpochNumber)
	}

	// Allow any step number as long as it's the expected one or the next one
	// This is needed because different nodes may have different views of the current step
	if msg.Step != s.CurrentStep && msg.Step != s.CurrentStep-1 && msg.Step != s.CurrentStep+1 {
		return fmt.Errorf("unexpected step: got %d, expected one of [%d, %d, %d]",
			msg.Step, s.CurrentStep-1, s.CurrentStep, s.CurrentStep+1)
	}

	// Verify the contribution using snarkjs
	tmpFile, err := os.CreateTemp("", "ptau-*.ptau")
	if err != nil {
		return fmt.Errorf("failed to create temporary ptau file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(msg.PTauFileData); err != nil {
		return fmt.Errorf("failed to write ptau file: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command("npx", "snarkjs", "powersoftau", "verify", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to verify contribution: %v\n%s", err, output)
	}

	// If verification succeeds, update the state
	s.PTauPath = tmpFile.Name()

	if !strings.Contains(string(output), "Powers of Tau Ok!") {
		return fmt.Errorf("contribution verification failed")
	}

	return nil
}

// FinalizeCeremony finalizes the Powers of Tau ceremony by adding a random beacon
// and preparing the final PTau file for circuit-specific setup
func (s *PTauCeremonyState) FinalizeCeremony() (string, error) {
	if !s.Completed {
		return "", fmt.Errorf("ceremony not completed yet")
	}

	// Create the final PTau file path
	finalPath := filepath.Join(filepath.Dir(s.PTauPath), fmt.Sprintf("pot_epoch%d_final.ptau", s.EpochNumber))

	// Add a random beacon contribution
	// The beacon hash is a random 64-character hex string
	beaconHash := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	numIterationsExp := 10 // Number of iterations as a power of 2 (2^10 = 1024 iterations)

	log.Info().Msgf("Adding random beacon to PTau file: %s", s.PTauPath)

	// Use snarkjs to add a random beacon
	cmd := exec.Command("npx", "snarkjs", "powersoftau", "beacon",
		s.PTauPath, finalPath, beaconHash, fmt.Sprintf("%d", numIterationsExp))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to add beacon: %v\n%s", err, output)
	}

	// Update the state
	s.PTauPath = finalPath

	return finalPath, nil
}

// GenerateCircomKeys generates the proving and verification keys for a specific Circom circuit
// using the finalized Powers of Tau ceremony output
func GenerateCircomKeys(ptauPath, circuitPath, outputDir string) (string, string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Get the circuit name from the path
	circuitName := filepath.Base(circuitPath)
	circuitName = strings.TrimSuffix(circuitName, filepath.Ext(circuitName))

	// Compile the circuit
	r1csPath := filepath.Join(outputDir, circuitName+".r1cs")
	compileCmd := exec.Command("npx", "circom", circuitPath,
		"--r1cs", "--output", outputDir)

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to compile circuit: %v\n%s", err, compileOutput)
	}

	// Generate the zkey file (proving key)
	zkeyPath := filepath.Join(outputDir, circuitName+".zkey")
	zkeyCmd := exec.Command("npx", "snarkjs", "groth16", "setup",
		r1csPath, ptauPath, zkeyPath)

	zkeyOutput, err := zkeyCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate zkey: %v\n%s", err, zkeyOutput)
	}

	// Export the verification key
	vkeyPath := filepath.Join(outputDir, circuitName+"_verification_key.json")
	vkeyCmd := exec.Command("npx", "snarkjs", "zkey", "export", "verificationkey",
		zkeyPath, vkeyPath)

	vkeyOutput, err := vkeyCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to export verification key: %v\n%s", err, vkeyOutput)
	}

	return zkeyPath, vkeyPath, nil
}

// GenerateProof generates a proof for a Circom circuit using the provided inputs
func GenerateProof(zkeyPath, inputsPath, outputDir string) (string, string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Get the circuit name from the zkey path
	zkeyName := filepath.Base(zkeyPath)
	zkeyName = strings.TrimSuffix(zkeyName, filepath.Ext(zkeyName))

	// Generate the witness
	wtnsPath := filepath.Join(outputDir, zkeyName+".wtns")
	wtnsCmd := exec.Command("npx", "snarkjs", "wtns", "calculate",
		zkeyPath, inputsPath, wtnsPath)

	wtnsOutput, err := wtnsCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate witness: %v\n%s", err, wtnsOutput)
	}

	// Generate the proof
	proofPath := filepath.Join(outputDir, zkeyName+"_proof.json")
	publicPath := filepath.Join(outputDir, zkeyName+"_public.json")
	proofCmd := exec.Command("npx", "snarkjs", "groth16", "prove",
		zkeyPath, wtnsPath, proofPath, publicPath)

	proofOutput, err := proofCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate proof: %v\n%s", err, proofOutput)
	}

	return proofPath, publicPath, nil
}

// VerifyCircomProof verifies a proof generated for a Circom circuit
func VerifyCircomProof(vkeyPath, proofPath, publicPath string) (bool, error) {
	// Verify the proof
	verifyCmd := exec.Command("npx", "snarkjs", "groth16", "verify",
		vkeyPath, publicPath, proofPath)

	verifyOutput, err := verifyCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to verify proof: %v\n%s", err, verifyOutput)
	}

	// Check if the verification was successful
	if strings.Contains(string(verifyOutput), "OK") {
		return true, nil
	}

	return false, nil
}
