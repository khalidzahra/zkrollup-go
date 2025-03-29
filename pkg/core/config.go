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
	VerificationKey  string
	ProvingKey       string
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
