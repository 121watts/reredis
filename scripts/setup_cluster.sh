#!/bin/bash

# Reredis Cluster Setup Script
# This script starts multiple Reredis nodes and connects them into a cluster

set -e

echo "🚀 Setting up Reredis Cluster..."

# Configuration
NODES=3
BASE_TCP_PORT=6379
BASE_HTTP_PORT=9080
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PIDS_FILE="$SCRIPT_DIR/cluster_pids.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up cluster nodes...${NC}"
    if [[ -f "$PIDS_FILE" ]]; then
        while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                echo "Stopping process $pid"
                kill "$pid" 2>/dev/null || true
            fi
        done < "$PIDS_FILE"
        rm -f "$PIDS_FILE"
    fi
    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

# Setup cleanup trap
trap cleanup EXIT INT TERM

# Check if binary exists
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_PATH="$PROJECT_ROOT/reredis"

if [[ ! -f "$BINARY_PATH" ]]; then
    echo -e "${BLUE}🔨 Building Reredis binary...${NC}"
    cd "$PROJECT_ROOT"
    go build -o reredis ./cmd/reredis/
    echo -e "${GREEN}✅ Build complete${NC}"
fi

# Clear any existing PID file
> "$PIDS_FILE"

echo -e "${BLUE}🏗️  Starting $NODES cluster nodes...${NC}"

# Start nodes
for i in $(seq 0 $((NODES-1))); do
    TCP_PORT=$((BASE_TCP_PORT + i))
    HTTP_PORT=$((BASE_HTTP_PORT + i))
    
    echo -e "${YELLOW}📡 Starting node $((i+1))/${NODES} on TCP:$TCP_PORT HTTP:$HTTP_PORT${NC}"
    
    # Start the node in background
    "$BINARY_PATH" -port="$TCP_PORT" -http-port="$HTTP_PORT" &
    NODE_PID=$!
    echo "$NODE_PID" >> "$PIDS_FILE"
    
    echo -e "${GREEN}✅ Node $((i+1)) started (PID: $NODE_PID)${NC}"
    
    # Small delay to ensure clean startup
    sleep 1
done

echo -e "\n${BLUE}⏳ Waiting for nodes to initialize...${NC}"
sleep 3

# Connect nodes using CLUSTER MEET commands
echo -e "\n${BLUE}🔗 Connecting nodes to form cluster...${NC}"

# Connect all nodes to the first node
for i in $(seq 1 $((NODES-1))); do
    TARGET_PORT=$((BASE_TCP_PORT + i))
    echo -e "${YELLOW}🤝 Connecting node on port $TARGET_PORT to cluster...${NC}"
    
    # Send CLUSTER MEET command using netcat (compatible with our simple protocol parser)
    if echo "CLUSTER MEET 127.0.0.1 $TARGET_PORT" | nc -w 1 127.0.0.1 "$BASE_TCP_PORT" | grep -q "OK"; then
        echo "✅ Node connected successfully"
    else
        echo "⚠️ Connection attempt completed (may need manual verification)"
    fi
    
    sleep 1
done

echo -e "\n${GREEN}🎉 Cluster setup complete!${NC}"
echo -e "\n${BLUE}📊 Cluster Information:${NC}"
echo "┌─────────┬──────────┬───────────┬─────────────────────────────────┐"
echo "│ Node    │ TCP Port │ HTTP Port │ Status                          │"
echo "├─────────┼──────────┼───────────┼─────────────────────────────────┤"

for i in $(seq 0 $((NODES-1))); do
    TCP_PORT=$((BASE_TCP_PORT + i))
    HTTP_PORT=$((BASE_HTTP_PORT + i))
    
    if [[ $i -eq 0 ]]; then
        STATUS="Primary (Web UI)"
    else
        STATUS="Secondary"
    fi
    
    printf "│ Node %-2d │ %-8d │ %-9d │ %-31s │\n" $((i+1)) "$TCP_PORT" "$HTTP_PORT" "$STATUS"
done

echo "└─────────┴──────────┴───────────┴─────────────────────────────────┘"

echo -e "\n${BLUE}🌐 Access Points:${NC}"
echo -e "  • ${GREEN}Web Dashboard:${NC} http://localhost:$BASE_HTTP_PORT"
echo -e "  • ${GREEN}WebSocket:${NC} ws://localhost:$BASE_HTTP_PORT/ws"
echo -e "  • ${GREEN}Redis Protocol:${NC} redis://localhost:$BASE_TCP_PORT"
echo -e "\n${YELLOW}📝 Note:${NC} If you have a dev server on :8080, connect to ws://localhost:$BASE_HTTP_PORT/ws instead"

echo -e "\n${BLUE}🧪 Test Commands:${NC}"
echo "  # Test Redis protocol:"
if command -v redis-cli &> /dev/null; then
    echo "  redis-cli -p $BASE_TCP_PORT SET test_key test_value"
    echo "  redis-cli -p $BASE_TCP_PORT GET test_key"
    echo "  redis-cli -p $BASE_TCP_PORT CLUSTER INFO"
else
    echo "  echo 'SET test_key test_value' | nc 127.0.0.1 $BASE_TCP_PORT"
    echo "  echo 'GET test_key' | nc 127.0.0.1 $BASE_TCP_PORT"
fi

echo -e "\n${BLUE}🔧 Cluster Commands:${NC}"
echo "  # Add data to test slot distribution:"
echo "  for i in {1..100}; do redis-cli -p $BASE_TCP_PORT SET \"key\$i\" \"value\$i\"; done"

echo -e "\n${YELLOW}💡 Tips:${NC}"
echo "  • Open the Web Dashboard to see the cluster topology"
echo "  • Add keys to see them distributed across nodes"
echo "  • The cluster will auto-initialize when 3 nodes are connected"
echo "  • Use Ctrl+C to stop all nodes"

echo -e "\n${GREEN}🎯 Cluster is ready! Press Ctrl+C to stop all nodes.${NC}"

# Keep script running and show live status
while true; do
    sleep 10
    
    # Check if all processes are still running
    RUNNING=0
    while read -r pid; do
        if kill -0 "$pid" 2>/dev/null; then
            ((RUNNING++))
        fi
    done < "$PIDS_FILE"
    
    if [[ $RUNNING -eq 0 ]]; then
        echo -e "\n${RED}❌ All cluster nodes have stopped${NC}"
        break
    elif [[ $RUNNING -lt $NODES ]]; then
        echo -e "\n${YELLOW}⚠️  Warning: Only $RUNNING/$NODES nodes running${NC}"
    fi
done