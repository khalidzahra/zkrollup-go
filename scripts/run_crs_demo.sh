#!/bin/bash

set -e  # Exit on error

# Colors for output
green='\033[0;32m'
yellow='\033[1;33m'
red='\033[0;31m'
blue='\033[0;34m'
nc='\033[0m' # No Color

# Configuration
NUM_SEQUENCERS=3
BASE_PORT=9100
DEMO_DIR="./demo_crs_data"

# Ensure the demo directory exists with absolute path
DEMO_DIR_ABS="$(cd "$(dirname "$0")/.." && pwd)/$DEMO_DIR"
mkdir -p "$DEMO_DIR_ABS"

# Use absolute paths for logs
SEQUENCER_LOGS=()
for i in $(seq 1 $NUM_SEQUENCERS); do
    SEQUENCER_LOGS[$i]="$DEMO_DIR_ABS/sequencer_$i.log"
done

# Cleanup function
cleanup() {
    echo -e "${yellow}Cleaning up...${nc}"
    for pid in "${SEQUENCER_PIDS[@]}"; do
        if [ ! -z "$pid" ]; then
            kill $pid 2>/dev/null || true
        fi
    done
    echo -e "${green}Cleanup complete.${nc}"
}
trap cleanup EXIT

cd ..
# Start sequencer nodes
echo -e "${blue}=== CRS Ceremony Demo ===${nc}"

# 1. Start the leader node (sequencer 1)
leader_port=$((BASE_PORT + 1))
leader_log="${SEQUENCER_LOGS[1]}"
echo -e "${yellow}Starting leader (sequencer 1) on port $leader_port...${nc}"
IS_LEADER=true SEQUENCER_PORT=$leader_port NODE_ID="sequencer_1" go run main.go > "$leader_log" 2>&1 &
SEQUENCER_PIDS[1]=$!
echo "Leader started with PID: ${SEQUENCER_PIDS[1]} (logs: $leader_log)"

# Wait for leader to output its multiaddress
leader_addr=""
echo -n "Waiting for leader's multiaddress..."
for i in {1..20}; do
    leader_addr=$(grep -Eo '/ip4/127.0.0.1/tcp/'"$leader_port"'/p2p/[A-Za-z0-9]+' "$leader_log" | head -n 1)
    if [ ! -z "$leader_addr" ]; then
        echo -e "${green}Found leader address: $leader_addr${nc}"
        break
    fi
    sleep 1
    echo -n "."
done
if [ -z "$leader_addr" ]; then
    echo -e "${red}\nFailed to obtain leader's multiaddress from log. Exiting.${nc}"
    exit 1
fi

# 2. Start the other sequencer nodes with BOOTSTRAP_PEERS
for i in $(seq 2 $NUM_SEQUENCERS); do
    port=$((BASE_PORT + i))
    log_file="${SEQUENCER_LOGS[$i]}"
    echo -e "${yellow}Starting sequencer $i on port $port...${nc}"
    BOOTSTRAP_PEERS="$leader_addr" SEQUENCER_PORT=$port NODE_ID="sequencer_$i" go run main.go > "$log_file" 2>&1 &
    SEQUENCER_PIDS[$i]=$!
    sleep 1
    echo "Sequencer $i started with PID: ${SEQUENCER_PIDS[$i]} (logs: $log_file)"
done

echo -e "${yellow}Waiting for CRS ceremony to complete...${nc}"
sleep 10

echo -e "${blue}=== CRS Ceremony Logs ===${nc}"
for i in $(seq 1 $NUM_SEQUENCERS); do
    log_file="${SEQUENCER_LOGS[$i]}"
    echo -e "${green}--- Sequencer $i CRS Log ---${nc}"
    grep -E "CRS contribution added|Final CRS generated for epoch" "$log_file" || echo "No CRS logs found."
    echo
    echo "(Full log: $log_file)"
done

echo -e "${green}CRS demo complete. Press Ctrl+C to exit and cleanup.${nc}"
wait
