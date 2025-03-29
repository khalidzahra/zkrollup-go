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
	BatchSize        uint64
	ProofGeneration  bool
	StateDBPath      string

	// ZK-SNARK configuration
	CircuitFile      string
	ProvingKeyFile   string
	VerifyingKeyFile string
	MerkleTreeDepth  int
}

func DefaultConfig() *Config {
	return &Config{
		EthereumRPC:     "http://localhost:8545",
		ChainID:         1337, // Local network
		SequencerPort:   9000,
		BatchSize:       100,
		ProofGeneration: true,
		StateDBPath:     "./statedb",
	}
}
