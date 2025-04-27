package core

type Config struct {
	// Ethereum network configuration
	EthereumRPC     string
	ChainID         int64
	ContractAddress string

	// Sequencer configuration
	SequencerPort    int
	SequencerPeerKey string
	BootstrapPeers   []string

	// Rollup configuration
	BatchSize       uint64
	ProofGeneration bool
	StateDBPath     string

	// ZK-SNARK configuration
	CircuitFile      string
	ProvingKeyFile   string
	VerifyingKeyFile string
	MerkleTreeDepth  int

	// L1 integration configuration
	L1Enabled           bool
	L1PrivateKey        string
	L1BatchSubmitPeriod int // in seconds
	L1GasLimit          uint64
	L1GasPrice          int64 // in gwei
}

func DefaultConfig() *Config {
	return &Config{
		EthereumRPC:         "http://localhost:8545",
		ChainID:             1337, // Local network
		SequencerPort:       9000,
		BatchSize:           1,
		ProofGeneration:     true,
		StateDBPath:         "./statedb",
		L1Enabled:           false, // Disabled by default
		L1BatchSubmitPeriod: 300,   // 5 minutes
		L1GasLimit:          3000000,
		L1GasPrice:          20, // 20 gwei
	}
}
