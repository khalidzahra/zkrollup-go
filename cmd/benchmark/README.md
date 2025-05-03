# ZK-Rollup Proof Verification Benchmark

This tool benchmarks the performance and cost of ZK proof generation and verification in the ZK-Rollup implementation.

## Overview

The benchmark tool measures:

1. **Proof Generation Time**: How long it takes to generate ZK proofs
2. **Proof Size**: The size of the generated proofs in bytes
3. **Local Verification Time**: How long it takes to verify proofs locally
4. **Simulated Gas Costs**: Gas costs for verifying proofs on-chain, measured using EVM simulation

## Consistent Message Hash Computation

This benchmark tool ensures consistent message hash computation between keygen, circuit, and transaction processing by:

1. Using the same format for all inputs to the hash function
2. Properly handling zero values by using a single byte with value 0 instead of an empty array
3. Converting the nonce to a consistent string representation

## Usage

```bash
go run main.go [flags]
```

### Flags

- `-iterations int`: Number of iterations for benchmarking (default: 10)
- `-proof-sizes string`: Comma-separated list of proof sizes to benchmark in bytes (default: "128,256,512,1024")
- `-rpc string`: Ethereum RPC URL (default: "http://localhost:8545")
- `-key string`: Private key in hex format without 0x prefix (not required for local benchmarking)
- `-simulate bool`: Use EVM simulation for gas measurements (default: true)
- `-vk string`: Path to the verifying key file (default: "generate/zkrollup.vk")
- `-pk string`: Path to the proving key file (default: "generate/zkrollup.pk")
- `-contract string`: Path to the verifier contract file (default: "generate/contract.sol")

### Examples

Run the benchmark with default settings:
```bash
go run main.go
```

Run the benchmark with 20 iterations:
```bash
go run main.go -iterations 20
```

Benchmark specific proof sizes:
```bash
go run main.go -proof-sizes "256,512,1024,2048"
```

Use custom key files and contract:
```bash
go run main.go -vk /path/to/verifying.key -pk /path/to/proving.key -contract /path/to/contract.sol
```

## Key Generation

Before running the benchmark, ensure that you have generated the necessary proving and verifying keys using the key generation tool:

```bash
go run generate/generate_files.go
```

This will create:
- `zkrollup.vk`: The verifying key file
- `zkrollup.pk`: The proving key file
- `contract.sol`: The Solidity verifier contract

## EVM Simulation

The benchmark tool uses EVM simulation to provide accurate gas cost measurements for on-chain verification. This approach:

1. Creates a simulated Ethereum blockchain environment
2. Deploys the verifier contract from the specified contract file
3. Executes verification transactions against the contract
4. Measures the actual gas used by these transactions

This provides more accurate cost estimates than static approximations, as it accounts for the actual execution of the verification logic in the EVM.

## Output

The benchmark outputs:

1. **Individual Results**: For each iteration, showing proof size, generation time, verification time, and gas usage
2. **Average Results**: Average proof size, generation time, verification time, and gas usage across all iterations
3. **Gas Costs by Proof Size**: A table showing simulated gas usage and ETH cost for different proof sizes

## Implementation Details

The benchmark:

1. Generates key pairs for signing and verification
2. Creates transaction parameters (amount, balance, nonce)
3. Signs the transaction data
4. Creates a witness for the ZK circuit
5. Generates a ZK proof
6. Verifies the proof locally
7. Simulates on-chain verification using a local EVM instance

## Important Notes

- The simulated gas costs are based on the current EVM implementation and may change with future Ethereum upgrades
- Actual costs may vary depending on the specific circuit implementation and Ethereum network conditions
- The benchmark ensures consistent message hash computation by properly handling zero values and using consistent formats

## Future Improvements

- Deploy an actual verifier contract instead of simulating the verification
- Compare different proving systems (Groth16, Plonk, etc.)
- Benchmark different circuit complexities
