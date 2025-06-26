#!/bin/bash

# Reredis Cluster Seed Script
# This script populates the cluster with mock data to demonstrate slot distribution

set -e

echo "🌱 Seeding Reredis Cluster with Mock Data..."

# Configuration
BASE_TCP_PORT=6379
NODES=3
TOTAL_KEYS=1000

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if nc (netcat) is available
if ! command -v nc &> /dev/null; then
    echo -e "${RED}❌ netcat (nc) not found. This tool is required for communication.${NC}"
    echo "  # Install on macOS:"
    echo "  brew install netcat"
    echo "  # Install on Ubuntu/Debian:"
    echo "  apt-get install netcat-openbsd"
    exit 1
fi

# Test connection to primary node
echo -e "${BLUE}🔍 Testing connection to cluster...${NC}"
if ! echo "SET _ping_test ping" | nc -w 1 127.0.0.1 $BASE_TCP_PORT > /dev/null 2>&1; then
    echo -e "${RED}❌ Cannot connect to cluster on port $BASE_TCP_PORT${NC}"
    echo -e "${YELLOW}💡 Make sure the cluster is running: ./scripts/setup_cluster.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Connected to cluster${NC}"

# Function to handle MOVED redirects when setting keys
set_key_with_redirect() {
    local key="$1"
    local value="$2"
    local port="$BASE_TCP_PORT"
    local max_redirects=3
    local redirect_count=0
    
    while [[ $redirect_count -lt $max_redirects ]]; do
        local response=$(echo "SET \"$key\" \"$value\"" | nc -w 1 127.0.0.1 $port 2>/dev/null)
        
        if [[ "$response" == "+OK"* ]]; then
            # Success!
            return 0
        elif [[ "$response" == "-MOVED"* ]]; then
            # Extract the new port from MOVED response: -MOVED 12345 127.0.0.1:6380
            local new_port=$(echo "$response" | grep -o '127\.0\.0\.1:[0-9]*' | cut -d: -f2)
            if [[ -n "$new_port" ]]; then
                port="$new_port"
                ((redirect_count++))
            else
                # Malformed MOVED response, give up
                break
            fi
        else
            # Some other error, give up
            break
        fi
    done
    
    # If we get here, we failed to set the key
    return 1
}

# Data categories for realistic mock data
declare -a CATEGORIES=("user" "product" "session" "cache" "config" "metrics" "logs")
declare -a USER_NAMES=("alice" "bob" "charlie" "diana" "eve" "frank" "grace" "henry" "iris" "jack" "kate" "liam" "mia" "noah" "olivia" "paul" "quinn" "ruby" "sam" "tina")
declare -a PRODUCTS=("laptop" "phone" "tablet" "monitor" "keyboard" "mouse" "headphones" "webcam" "speaker" "printer")
declare -a COUNTRIES=("us" "uk" "ca" "de" "fr" "jp" "au" "br" "in" "cn")

# Function to generate random data
generate_user_data() {
    local user_id=$1
    local name=${USER_NAMES[$((RANDOM % ${#USER_NAMES[@]}))]}
    local country=${COUNTRIES[$((RANDOM % ${#COUNTRIES[@]}))]}
    local score=$((RANDOM % 1000))
    echo "{\"id\":$user_id,\"name\":\"$name\",\"country\":\"$country\",\"score\":$score,\"created\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
}

generate_product_data() {
    local product_id=$1
    local name=${PRODUCTS[$((RANDOM % ${#PRODUCTS[@]}))]}
    local price=$((RANDOM % 2000 + 10))
    local stock=$((RANDOM % 100))
    echo "{\"id\":$product_id,\"name\":\"$name\",\"price\":$price,\"stock\":$stock,\"category\":\"electronics\"}"
}

generate_session_data() {
    local session_id=$1
    local user_id=$((RANDOM % 1000))
    local ttl=$((RANDOM % 3600 + 300))
    echo "{\"session_id\":\"$session_id\",\"user_id\":$user_id,\"expires_in\":$ttl,\"ip\":\"192.168.1.$((RANDOM % 255))\"}"
}

generate_metrics_data() {
    local metric_name=$1
    local value=$((RANDOM % 100))
    local timestamp=$(date +%s)
    echo "{\"metric\":\"$metric_name\",\"value\":$value,\"timestamp\":$timestamp,\"host\":\"node-$((RANDOM % 3 + 1))\"}"
}

# Function to add data with progress
add_data_batch() {
    local category=$1
    local count=$2
    local port=$3
    local success_count=0
    
    echo -e "${CYAN}📝 Adding $count $category records...${NC}"
    
    for ((i=1; i<=count; i++)); do
        case $category in
            "user")
                key="user:$i"
                value=$(generate_user_data $i)
                ;;
            "product")
                key="product:$i"
                value=$(generate_product_data $i)
                ;;
            "session")
                key="session:sess_$(printf "%06d" $i)"
                value=$(generate_session_data "sess_$(printf "%06d" $i)")
                ;;
            "cache")
                key="cache:page_$i"
                value="{\"page\":$i,\"content\":\"Page_${i}_content\",\"cached_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"ttl\":3600}"
                ;;
            "config")
                key="config:feature_$i"
                value="{\"enabled\":$((RANDOM % 2)),\"version\":\"1.$((RANDOM % 10))\",\"updated\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
                ;;
            "metrics")
                metric_names=("cpu_usage" "memory_usage" "disk_io" "network_rx" "network_tx" "response_time")
                metric_name=${metric_names[$((RANDOM % ${#metric_names[@]}))]}
                key="metrics:${metric_name}_$i"
                value=$(generate_metrics_data $metric_name)
                ;;
            "logs")
                levels=("INFO" "WARN" "ERROR" "DEBUG")
                level=${levels[$((RANDOM % ${#levels[@]}))]}
                key="logs:$(date +%Y%m%d)_$i"
                value="{\"level\":\"$level\",\"message\":\"Log_entry_$i\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"service\":\"api\"}"
                ;;
        esac
        
        # Add the key-value pair with MOVED redirect handling
        if set_key_with_redirect "$key" "$value"; then
            ((success_count++))
        fi
        
        # Show progress every 50 items
        if ((i % 50 == 0)); then
            echo -e "${YELLOW}  ⏳ Progress: $i/$count ($(($i * 100 / $count))%) - $success_count successful${NC}"
        fi
    done
    
    echo -e "${GREEN}  ✅ Added $success_count/$count $category records${NC}"
}

# Function to add hash tag grouped data (keys that should stay together)
add_hash_tag_data() {
    echo -e "${PURPLE}🏷️  Adding hash tag grouped data...${NC}"
    local success_count=0
    
    # User shopping carts (should be on same node as user data)
    for ((i=1; i<=50; i++)); do
        user_id=$i
        cart_key="cart:{user:$user_id}"
        cart_data="{\"user_id\":$user_id,\"items\":[{\"product_id\":$((RANDOM % 100 + 1)),\"quantity\":$((RANDOM % 5 + 1))}],\"total\":$((RANDOM % 500 + 10))}"
        if set_key_with_redirect "$cart_key" "$cart_data"; then
            ((success_count++))
        fi
        
        # User preferences (should be on same node as user data)
        pref_key="preferences:{user:$user_id}"
        pref_data="{\"theme\":\"dark\",\"language\":\"en\",\"notifications\":true,\"updated\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
        if set_key_with_redirect "$pref_key" "$pref_data"; then
            ((success_count++))
        fi
    done
    
    echo -e "${GREEN}  ✅ Added $success_count/100 hash tag grouped records${NC}"
}

# Main seeding process
echo -e "\n${BLUE}📊 Seeding Strategy:${NC}"
echo "  • Total Keys: $TOTAL_KEYS"
echo "  • Data Categories: ${#CATEGORIES[@]}"
echo "  • Keys per Category: $((TOTAL_KEYS / ${#CATEGORIES[@]}))"
echo "  • Hash Tag Groups: 50 user groups (2 keys each)"

echo -e "\n${BLUE}🚀 Starting data seeding...${NC}"

# Calculate keys per category
keys_per_category=$((TOTAL_KEYS / ${#CATEGORIES[@]}))

# Add data for each category
for category in "${CATEGORIES[@]}"; do
    add_data_batch "$category" "$keys_per_category" "$BASE_TCP_PORT"
done

# Add hash tag grouped data
add_hash_tag_data

echo -e "\n${BLUE}📈 Generating cluster statistics...${NC}"

# Calculate expected key distribution
expected_total=$((TOTAL_KEYS + 100)) # includes hash tag grouped data
echo -e "\n${GREEN}🎯 Cluster Statistics:${NC}"
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│                    CLUSTER KEY DISTRIBUTION                 │"
echo "├─────────────────────────────────────────────────────────────┤"

for ((i=0; i<NODES; i++)); do
    port=$((BASE_TCP_PORT + i))
    if echo "SET _test_connection test" | nc -w 1 127.0.0.1 $port > /dev/null 2>&1; then
        printf "│ Node %-2d (Port %-4d): %-6s keys (estimated)            │\n" $((i+1)) $port "~$((expected_total / NODES))"
    else
        printf "│ Node %-2d (Port %-4d): %-6s (offline)                  │\n" $((i+1)) $port "N/A"
    fi
done

echo "├─────────────────────────────────────────────────────────────┤"
printf "│ Total Keys Added: %-6s                                 │\n" "$expected_total"
echo "└─────────────────────────────────────────────────────────────┘"

echo -e "\n${BLUE}🔍 Sample Data Preview:${NC}"
echo "  • User data: $(echo 'GET "user:1"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT 2>/dev/null | head -c 60)..."
echo "  • Product data: $(echo 'GET "product:1"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT 2>/dev/null | head -c 60)..."
echo "  • Session data: $(echo 'GET "session:sess_000001"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT 2>/dev/null | head -c 60)..."

echo -e "\n${BLUE}🧪 Test Commands:${NC}"
echo "  # Check specific keys:"
echo "  echo 'GET \"user:1\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"
echo "  echo 'GET \"cart:{user:1}\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"
echo ""
echo "  # Test different data types:"
echo "  echo 'GET \"product:5\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"
echo "  echo 'GET \"metrics:cpu_usage_10\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"

echo -e "\n${GREEN}🎉 Cluster seeding complete!${NC}"
echo -e "${CYAN}💡 Open the Web Dashboard to see the data distributed across nodes${NC}"
echo -e "${YELLOW}🌐 Dashboard: http://localhost:8080${NC}"

# Show hash tag demonstration
echo -e "\n${PURPLE}🏷️  Hash Tag Demo:${NC}"
echo "Testing hash tag grouped keys:"
echo "  echo 'GET \"cart:{user:1}\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"
echo "  echo 'GET \"preferences:{user:1}\"' | nc -w 1 127.0.0.1 $BASE_TCP_PORT"
echo ""
echo "These keys should be stored on the same node due to hash tag grouping!"