# ZK-Rollup L1 Integration

This directory contains the deployment script for the ZK-Rollup L1 contract.

## Contract Deployment

The ZK-Rollup contract has been deployed to the Ethereum network with the following details:

- Contract Address: ` + address.Hex() + `
- Ethereum RPC URL: ` + *rpcURL + `
- Chain ID: ` + fmt.Sprintf("%d", *chainID) + `

## Running the ZK-Rollup with L1 Integration

To run the ZK-Rollup node with L1 integration enabled, use the following command:

```bash
source .env.l1 && go run main.go
```

This will load the environment variables from the .env.l1 file and start the ZK-Rollup node with L1 integration enabled.

## Configuration Options

The following environment variables can be used to configure the L1 integration:

- ETHEREUM_RPC: Ethereum RPC URL
- CHAIN_ID: Ethereum chain ID
- CONTRACT_ADDRESS: Address of the deployed ZK-Rollup contract
- L1_PRIVATE_KEY: Private key for the Ethereum account
- L1_ENABLED: Set to "true" to enable L1 integration
- L1_BATCH_SUBMIT_PERIOD: Period (in seconds) for submitting batches to L1