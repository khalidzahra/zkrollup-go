#!/bin/bash

echo "Building ZK-Rollup components..."
go build -o bin/sequencer main.go
go build -o bin/client cmd/client/main.go

# Make sure the bin directory exists
mkdir -p bin

# Clean up any previous data
rm -rf data
mkdir -p data

# Create directory for node data
mkdir -p data

# Start the leader node
echo "Starting leader node..."
IS_LEADER=true SEQUENCER_PORT=9000 ./bin/sequencer > data/leader.log 2>&1 &
LEADER_PID=$!

# Give the leader node some time to initialize before checking logs
sleep 3

# Wait for leader to initialize and extract its ID
echo "Waiting for leader node to initialize..."
for i in {1..30}; do
    # Use strings to extract text content from potentially binary log
    strings data/leader.log > data/leader_text.log
    
    if grep -q '"id":"' data/leader_text.log; then
        NODE_ID=$(grep -o '"id":"[^"]*"' data/leader_text.log | head -n 1 | cut -d '"' -f 4)
        if [ ! -z "$NODE_ID" ]; then
            echo "Found leader node ID: $NODE_ID"
            break
        fi
    fi
    
    if [ $i -eq 30 ]; then
        echo "Timed out waiting for leader node to start"
        exit 1
    fi
    
    echo -n "."
    sleep 1
done

# Construct the leader's address
LEADER_ADDR="/ip4/127.0.0.1/tcp/9000/p2p/$NODE_ID"
echo "Leader node address constructed as: $LEADER_ADDR"

# Start a follower node
echo "Starting follower node..."
IS_LEADER=false SEQUENCER_PORT=9001 BOOTSTRAP_PEERS=$LEADER_ADDR ./bin/sequencer > data/follower1.log 2>&1 &
FOLLOWER1_PID=$!

# Give the first follower some time to initialize
sleep 2

# Start another follower node
echo "Starting another follower node..."
IS_LEADER=false SEQUENCER_PORT=9002 BOOTSTRAP_PEERS=$LEADER_ADDR ./bin/sequencer > data/follower2.log 2>&1 &
FOLLOWER2_PID=$!

# Wait for followers to initialize and connect
echo "Waiting for follower nodes to connect..."
for i in {1..30}; do
    # Extract text content from potentially binary logs
    strings data/follower1.log > data/follower1_text.log
    strings data/follower2.log > data/follower2_text.log
    
    if grep -q "Connected to peer" data/follower1_text.log && grep -q "Connected to peer" data/follower2_text.log; then
        echo $'\nFollower nodes connected successfully'
        break
    fi
    
    if [ $i -eq 30 ]; then
        echo $'\nTimed out waiting for follower nodes to connect'
        exit 1
    fi
    
    echo -n "."
    sleep 1
done

# Wait a bit more to ensure all nodes are fully initialized
sleep 5

# Start the client and send test transactions
echo "Starting client and sending test transactions..."
./bin/client -port 9100 -peer $LEADER_ADDR > data/client.log 2>&1 &
CLIENT_PID=$!

# Give the client some time to connect and start sending transactions
sleep 2

# Wait for transactions to be processed
echo "Processing transactions..."
for i in {1..60}; do
    # Extract text content from potentially binary logs
    strings data/leader.log > data/leader_text.log
    strings data/follower1.log > data/follower1_text.log
    
    # Check if consensus messages are being processed
    if grep -q "Received consensus message" data/leader_text.log && \
       grep -q "Broadcasting prepare message" data/follower1_text.log && \
       grep -q "Received decided batch from consensus" data/leader_text.log; then
        echo $'\nConsensus appears to be working!'
        break
    fi
    
    if [ $i -eq 60 ]; then
        echo $'\nTimed out waiting for consensus to complete'
    fi
    
    echo -n "."
    sleep 1
done

# Give a bit more time for final processing
sleep 5

# Display logs (using strings to handle binary data)
echo "=== Leader Node Log ==="
strings data/leader.log | grep -E "consensus|transaction|protocol|peer" | tail -n 30

echo "=== Follower Node 1 Log ==="
strings data/follower1.log | grep -E "consensus|transaction|protocol|peer" | tail -n 30

echo "=== Follower Node 2 Log ==="
strings data/follower2.log | grep -E "consensus|transaction|protocol|peer" | tail -n 30

echo "=== Client Log ==="
strings data/client.log | grep -E "transaction|peer|connect" | tail -n 30

# Check if consensus messages were exchanged
echo "\n=== Checking for consensus messages ==="
CONSENSUS_MSGS=$(strings data/leader.log data/follower1.log data/follower2.log | grep -c "consensus")
echo "Found $CONSENSUS_MSGS consensus-related log entries"

# Check for zero value handling (from our memory)
echo "\n=== Checking for zero value handling ==="
ZERO_HANDLING=$(strings data/leader.log data/follower1.log | grep -c "zero")
echo "Found $ZERO_HANDLING zero value handling log entries"

# Cleanup
echo "Cleaning up..."
kill $LEADER_PID $FOLLOWER1_PID $FOLLOWER2_PID $CLIENT_PID

echo "Test complete!"
