#!/bin/bash

# Cluster Cleanup Script
# This script forcefully stops all reredis cluster nodes

echo "🧹 Cleaning up cluster nodes..."

# Kill all reredis processes
KILLED=0
for pid in $(pgrep -f "reredis -port"); do
    echo "Stopping reredis process: $pid"
    kill "$pid" 2>/dev/null && ((KILLED++))
done

# Also check for any remaining processes
pkill -f "reredis -port" 2>/dev/null && echo "Killed additional reredis processes"

# Clean up PID files
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
rm -f "$SCRIPT_DIR/cluster_pids.txt"

if [[ $KILLED -gt 0 ]]; then
    echo "✅ Stopped $KILLED cluster node(s)"
else
    echo "ℹ️  No cluster nodes were running"
fi

echo "🎯 Cluster cleanup complete!"