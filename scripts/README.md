# ZK-Rollup with L1 Integration Demo

This directory contains scripts for demonstrating the ZK-Rollup with L1 integration.

## Full Demo Script

The `run_full_demo.sh` script automates the entire process of running a ZK-Rollup with L1 integration:

1. Starts a local Ethereum node (using Hardhat)
2. Deploys the ZK-Rollup contract to L1
3. Starts the ZK-Rollup node with L1 integration enabled
4. Submits transactions to create batches
5. Verifies that batches are posted to L1

### Prerequisites

- Go 1.24 or later
- Node.js and npm (for running Hardhat)
- Hardhat (`npm install --global hardhat`)

### Usage

```bash
cd /home/khalidzahra/Desktop/zkrollup-go
./scripts/run_full_demo.sh
```

The script will create a `demo_data` directory in the project root to store logs.

### Configuration

You can modify the following variables at the top of the script to customize the demo:

- `L1_PORT`: Port for the local Ethereum node (default: 8545)
- `ROLLUP_PORT`: Port for the ZK-Rollup node (default: 9000)
- `RPC_PORT`: Port for the ZK-Rollup RPC server (default: 8081)
- `PRIVATE_KEY`: Private key for deploying contracts (default: Hardhat's first account)
- `CHAIN_ID`: Chain ID for the local Ethereum network (default: 1337)

### Stopping the Demo

Press `Ctrl+C` to stop all processes and clean up.

## Understanding the L1 Integration

The ZK-Rollup posts batches to L1 for data availability and security. Each batch contains:

1. A batch number
2. A state root (the Merkle root of the ZK-Rollup state)
3. Transaction hashes
4. A ZK proof (simplified in this demo)

When a batch is finalized by the consensus mechanism, it is queued for submission to L1. A background process then submits the batch to the L1 contract, which stores the state root and verifies the proof.

This provides a secure and verifiable record of all batches on the L1 chain, ensuring that the ZK-Rollup state can be reconstructed if needed.

## Customizing the Demo

To modify the batch submission period, change the `L1_BATCH_SUBMIT_PERIOD` environment variable in the script. The default is 30 seconds for the demo, but you might want to increase this in a production environment.

You can also modify the number of batch creation iterations by changing the loop counter in the script.
