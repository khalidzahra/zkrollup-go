#!/bin/bash

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
L1_PORT=8545
ROLLUP_PORT=9000
RPC_PORT=8081
DEMO_DIR="./demo_data"
PRIVATE_KEY="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  # Default Hardhat account
CHAIN_ID=1337

# Ensure the demo directory exists
mkdir -p $DEMO_DIR

echo -e "${BLUE}=== ZK-Rollup with L1 Integration Demo ===${NC}"
echo -e "${BLUE}This script will:${NC}"
echo -e "${BLUE}1. Start a local Ethereum node${NC}"
echo -e "${BLUE}2. Deploy the ZK-Rollup contract to L1${NC}"
echo -e "${BLUE}3. Start the ZK-Rollup node with L1 integration${NC}"
echo -e "${BLUE}4. Submit transactions to create batches${NC}"
echo -e "${BLUE}5. Verify that batches are posted to L1${NC}"
echo ""

# Check if required tools are installed
command -v npx >/dev/null 2>&1 || { echo -e "${RED}Error: npx is required but not installed. Install Node.js and npm first.${NC}" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo -e "${RED}Error: go is required but not installed.${NC}" >&2; exit 1; }

# Function to clean up processes on exit
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    # Stop and remove Docker container
    if docker ps -q -f name=hardhat-node >/dev/null 2>&1; then
        echo "Stopping and removing L1 node Docker container"
        docker stop hardhat-node >/dev/null 2>&1 || true
        docker rm hardhat-node >/dev/null 2>&1 || true
    fi
    # Kill background processes
    if [ ! -z "$ROLLUP_NODE_PID" ]; then
        echo "Stopping ZK-Rollup node (PID: $ROLLUP_NODE_PID)"
        kill $ROLLUP_NODE_PID 2>/dev/null || true
    fi
    echo -e "${GREEN}Cleanup complete.${NC}"
}

# Register the cleanup function to be called on exit
trap cleanup EXIT

# Step 1: Start a local Ethereum node
echo -e "${YELLOW}Starting local Ethereum node...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Start hardhat node using Docker
echo -e "${YELLOW}Starting Hardhat node using Docker on port $L1_PORT...${NC}"

# Check if the container already exists
if docker ps -a --format '{{.Names}}' | grep -q "hardhat-node"; then
    echo "Removing existing hardhat-node container..."
    docker rm -f hardhat-node >/dev/null 2>&1
fi

# Run the hardhat node in a Docker container
docker run -d --name hardhat-node \
    -p $L1_PORT:8545 \
    -e HARDHAT_CHAIN_ID=$CHAIN_ID \
    ethereum/client-go:latest \
    --dev \
    --http \
    --http.addr=0.0.0.0 \
    --http.port=8545 \
    --http.api=eth,net,web3,debug \
    --http.corsdomain='*' \
    --ws \
    --ws.addr=0.0.0.0 \
    --ws.port=8546 \
    --ws.api=eth,net,web3 \
    --ws.origins='*' \
    --allow-insecure-unlock \
    > $DEMO_DIR/l1_node.log 2>&1

L1_NODE_CONTAINER_ID=$(docker ps -q -f name=hardhat-node)
echo "L1 node started in Docker container: $L1_NODE_CONTAINER_ID"

# Wait for the node to start and verify it's running
echo "Waiting for L1 node to start..."
for i in {1..30}; do
    if curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' http://localhost:$L1_PORT >/dev/null 2>&1; then
        echo -e "${GREEN}L1 node started successfully.${NC}"
        break
    fi
    
    if [ $i -eq 30 ]; then
        echo -e "${RED}Failed to start L1 node. Check Docker logs for details.${NC}"
        echo "Docker logs:"
        docker logs hardhat-node | tail -n 20
        exit 1
    fi
    
    # Show progress and check if container is still running
    if ! docker ps -q -f name=hardhat-node >/dev/null 2>&1; then
        echo -e "${RED}L1 node container has stopped. Check Docker logs for details.${NC}"
        echo "Docker logs:"
        docker logs hardhat-node | tail -n 20
        exit 1
    fi
    
    echo "Waiting for L1 node to start... (attempt $i/30)"
    sleep 2
done

# Display some node info
echo "L1 node info:"
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' http://localhost:$L1_PORT | grep result

# Step 2: Deploy the ZK-Rollup contract to L1
echo -e "${YELLOW}Deploying ZK-Rollup contract to L1...${NC}"
cd /home/khalidzahra/Desktop/zkrollup-go

# Try to deploy the contract with retries
MAX_RETRIES=5
for i in $(seq 1 $MAX_RETRIES); do
    echo "Deployment attempt $i of $MAX_RETRIES..."
    if go run cmd/l1deploy/main.go -privatekey $PRIVATE_KEY -rpc http://localhost:$L1_PORT -chainid $CHAIN_ID; then
        echo -e "${GREEN}ZK-Rollup contract deployed successfully.${NC}"
        break
    else
        if [ $i -eq $MAX_RETRIES ]; then
            echo -e "${RED}Failed to deploy contract after $MAX_RETRIES attempts. Exiting.${NC}"
            exit 1
        fi
        echo -e "${YELLOW}Deployment failed. Retrying in 5 seconds...${NC}"
        sleep 5
    fi
done

# Load the environment variables from the .env.l1 file
if [ ! -f .env.l1 ]; then
    echo -e "${RED}Error: .env.l1 file not found. Contract deployment may have failed.${NC}"
    exit 1
fi
source .env.l1

# Verify that we have the required environment variables
if [ -z "$CONTRACT_ADDRESS" ] || [ -z "$L1_PRIVATE_KEY" ]; then
    echo -e "${RED}Error: Required environment variables not found in .env.l1${NC}"
    exit 1
fi

echo -e "${BLUE}Contract deployed at: $CONTRACT_ADDRESS${NC}"

# Step 3: Start the ZK-Rollup node with L1 integration
echo -e "${YELLOW}Starting ZK-Rollup node with L1 integration...${NC}"
L1_ENABLED=true \
ETHEREUM_RPC=$ETHEREUM_RPC \
CHAIN_ID=$CHAIN_ID \
CONTRACT_ADDRESS=$CONTRACT_ADDRESS \
L1_PRIVATE_KEY=$L1_PRIVATE_KEY \
L1_BATCH_SUBMIT_PERIOD=30 \
SEQUENCER_PORT=$ROLLUP_PORT \
RPC_PORT=$RPC_PORT \
IS_LEADER=true \
go run main.go > $DEMO_DIR/rollup_node.log 2>&1 &
ROLLUP_NODE_PID=$!
echo "ZK-Rollup node started with PID: $ROLLUP_NODE_PID"

# Wait for the rollup node to start
echo "Waiting for ZK-Rollup node to start..."
echo "Logs available at: $DEMO_DIR/rollup_node.log"
sleep 5

# Check if the rollup node is still running
if ! kill -0 $ROLLUP_NODE_PID 2>/dev/null; then
    echo -e "${RED}Error: ZK-Rollup node failed to start. Check logs for details.${NC}"
    echo "Last 20 lines of log:"
    tail -n 20 $DEMO_DIR/rollup_node.log
    exit 1
fi

# Display some initial log output
echo "Initial log output:"
tail -n 10 $DEMO_DIR/rollup_node.log
echo -e "${GREEN}ZK-Rollup node started successfully.${NC}"

# Step 4: Submit transactions to create batches
echo -e "${YELLOW}Submitting transactions to create batches...${NC}"

# Function to deploy a contract with retries
deploy_contract() {
    echo "Deploying a simple contract..."
    MAX_RETRIES=3
    for i in $(seq 1 $MAX_RETRIES); do
        if go run cmd/test/deploy/deploy_simple.go; then
            echo -e "${GREEN}Contract deployed successfully.${NC}"
            return 0
        else
            echo -e "${YELLOW}Contract deployment failed. Attempt $i of $MAX_RETRIES.${NC}"
            if [ $i -eq $MAX_RETRIES ]; then
                echo -e "${RED}Failed to deploy contract after $MAX_RETRIES attempts.${NC}"
                return 1
            fi
            sleep 5
        fi
    done
}

# Function to interact with a contract with retries
interact_with_contract() {
    echo "Interacting with the deployed contract..."
    MAX_RETRIES=3
    for i in $(seq 1 $MAX_RETRIES); do
        if go run cmd/test/interact/interact_contract.go; then
            echo -e "${GREEN}Contract interaction successful.${NC}"
            return 0
        else
            echo -e "${YELLOW}Contract interaction failed. Attempt $i of $MAX_RETRIES.${NC}"
            if [ $i -eq $MAX_RETRIES ]; then
                echo -e "${RED}Failed to interact with contract after $MAX_RETRIES attempts.${NC}"
                return 1
            fi
            sleep 5
        fi
    done
}

# Deploy and interact with contracts to generate batches
for i in {1..3}; do
    echo -e "${BLUE}Batch creation iteration $i${NC}"
    if deploy_contract; then
        if interact_with_contract; then
            echo "Waiting for batch to be processed..."
            sleep 30
        else
            echo -e "${YELLOW}Skipping batch wait due to interaction failure.${NC}"
            sleep 10
        fi
    else
        echo -e "${YELLOW}Skipping interaction due to deployment failure.${NC}"
        sleep 10
    fi
    
    # Check if the rollup node is still running
    if ! kill -0 $ROLLUP_NODE_PID 2>/dev/null; then
        echo -e "${RED}Error: ZK-Rollup node is no longer running. Check logs for details.${NC}"
        echo "Last 20 lines of rollup log:"
        tail -n 20 $DEMO_DIR/rollup_node.log
        exit 1
    fi
done

# Step 5: Verify that batches are posted to L1
echo -e "${YELLOW}Verifying batches on L1...${NC}"
echo "Checking the rollup node logs for batch submissions..."
grep -i "submitted batch to l1" $DEMO_DIR/rollup_node.log || echo "No batch submissions found in logs yet."

# Create a symbolic link to the logs directory for easier access
ln -sf "$(realpath $DEMO_DIR)" ./logs

echo -e "${GREEN}Demo completed successfully!${NC}"
echo -e "${BLUE}You can check the logs in the ./logs directory (symbolic link to $DEMO_DIR):${NC}"
echo "  - L1 node log: ./logs/l1_node.log"
echo "  - ZK-Rollup node log: ./logs/rollup_node.log"
echo ""
echo -e "${YELLOW}The demo is still running. Press Ctrl+C to stop all processes and clean up.${NC}"

# Keep the script running until user interrupts
while true; do
    sleep 10
    echo -e "${BLUE}Checking for new batch submissions...${NC}"
    grep -i "submitted batch to l1" $DEMO_DIR/rollup_node.log | tail -5 || echo "No new batch submissions found."
done
