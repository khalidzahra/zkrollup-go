# ZK Rollup Implementation in Go

This project implements a ZK Rollup with a decentralized sequencer network running on a local Ethereum network.

## Getting Started

1. Install dependencies:
```bash
go mod tidy
```

2. Run a local Ethereum network (e.g., Ganache or Geth in dev mode)

3. Start the rollup node:
```bash
go run main.go
```

## TODO

- [ ] Implement ZK-SNARK circuit for transaction verification
- [ ] Add P2P networking for sequencer communication
- [ ] Deploy and integrate smart contracts
- [ ] Add proof generation and verification
- [ ] Implement state commitment and merkle tree
- [ ] Add transaction signature verification
