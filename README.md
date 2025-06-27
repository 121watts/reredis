# Reredis

A lightweight Redis-compatible key-value store written in Go with real-time WebSocket capabilities and distributed clustering support.

## Features

- **Redis Protocol Compatibility**: Supports core Redis commands (SET, GET, DEL)
- **Write-Ahead Logging (WAL)**: RESP-encoded persistence for data durability
- **Real-time Updates**: WebSocket server for live data synchronization
- **Distributed Clustering**: Redis-compatible clustering with automatic slot distribution
- **Web Interface**: React frontend for visualizing key-value data
- **Concurrent Architecture**: Goroutine-per-connection model for high performance

## Architecture

Reredis consists of four main components:

1. **TCP Server** (port 6379+): Implements Redis protocol for client connections
2. **HTTP/WebSocket Server** (port 8080+): Provides real-time updates and web interface  
3. **Cluster Manager**: Handles distributed hash slot assignment and node discovery
4. **WAL System**: Write-Ahead Logging for data persistence and recovery

## Quick Start

### Single Node

```bash
# Build the project
go build ./cmd/reredis

# Start a single node
./reredis

# Connect with redis-cli
redis-cli -p 6379
127.0.0.1:6379> SET mykey "hello world"
OK
127.0.0.1:6379> GET mykey
"hello world"
```

### 3-Node Cluster

```bash
# Terminal 1: Start first node
./reredis --port=6379 --http-port=8080

# Terminal 2: Start second node  
./reredis --port=6380 --http-port=8081

# Terminal 3: Start third node
./reredis --port=6381 --http-port=8082

# Connect nodes to form cluster
redis-cli -p 6379 CLUSTER MEET 127.0.0.1 6380
redis-cli -p 6379 CLUSTER MEET 127.0.0.1 6381

# Test cluster routing
redis-cli -p 6379 SET mykey "distributed!"
# May return: MOVED 8432 127.0.0.1:6380 (if key belongs to different node)

redis-cli -p 6380 SET mykey "distributed!"
OK
```

## Web Interface

Visit `http://localhost:8080` (or 8081, 8082 for cluster nodes) to access the real-time web interface showing:

- Current key-value pairs
- Live updates as data changes
- Cluster node information

### Frontend Development

```bash
cd frontend
npm install
npm run dev  # Development server
npm run build  # Production build
```

## Commands

### Redis Commands
- `SET key value` - Store a key-value pair
- `GET key` - Retrieve a value by key  
- `DEL key` - Delete a key

### Cluster Commands
- `CLUSTER MEET ip port` - Add a node to the cluster
- `CLUSTER NODES` - List all cluster nodes (planned)
- `CLUSTER INFO` - Show cluster status (planned)

## Development

### Build and Test

```bash
# Build
go build ./cmd/reredis

# Run tests
go test ./...

# Run integration tests
go test ./cmd/reredis

# Format code
gofmt -s -w .
go vet ./...
```

### Project Structure

```
├── cmd/reredis/          # Main application
├── internal/
│   ├── cluster/          # Clustering logic
│   │   ├── manager.go    # Cluster state management
│   │   ├── node.go       # Node definitions
│   │   └── hashslot.go   # Hash slot calculation
│   ├── server/           # TCP and HTTP servers
│   │   ├── server.go     # TCP server and connection handling
│   │   ├── handler.go    # Command business logic
│   │   └── http.go       # WebSocket and HTTP server
│   ├── store/            # In-memory key-value store
│   ├── wal/              # Write-Ahead Logging
│   │   ├── encoder.go    # RESP encoding for WAL entries
│   │   └── writer.go     # WAL file writing
│   └── observer/         # WebSocket event broadcasting
├── frontend/             # React web interface
└── CLAUDE.md            # Development guidelines
```

## Clustering

Reredis implements Redis-compatible clustering:

- **16,384 hash slots** distributed across nodes
- **CRC32 hashing** for key-to-slot mapping
- **MOVED responses** for client redirection
- **Automatic slot assignment** when cluster forms
- **Manual slot migration** (planned)

### Hash Slot Distribution

For a 3-node cluster:
- Node 0: slots 0-5460
- Node 1: slots 5461-10922  
- Node 2: slots 10923-16383

Keys are automatically routed to the correct node based on their hash slot.

## Compatibility

- **Redis Protocol**: Compatible with standard Redis clients
- **Redis Clustering**: Supports MOVED redirections
- **WebSocket**: JSON-based real-time protocol
- **Go Version**: Requires Go 1.19+

## Performance

- **Concurrent**: Handles multiple connections simultaneously
- **In-Memory**: All data stored in RAM for fast access
- **Efficient**: Minimal overhead routing with O(1) slot lookups
- **Real-time**: WebSocket updates with sub-millisecond latency

## Write-Ahead Logging (WAL)

Reredis implements WAL for data durability and crash recovery:

- **RESP Format**: WAL entries use Redis protocol encoding for consistency
- **Command Logging**: SET and DEL operations are logged before execution
- **Atomic Writes**: Each WAL entry is synced to disk before responding to client
- **Per-Node WAL**: Each cluster node maintains its own WAL file

### WAL Implementation Status

✅ **Completed**
- [x] RESP encoding/decoding for WAL entries
- [x] WAL writer with atomic file operations
- [x] Command logging for SET and DEL operations
- [x] Integration with command handlers
- [x] Clean architecture separation (handlers vs I/O)

🚧 **In Progress**
- [ ] WAL reader for parsing entries
- [ ] Recovery system to replay WAL on startup
- [ ] WAL file rotation and management
- [ ] Slot-aware WAL for cluster operations

📋 **Planned WAL Features**
- [ ] WAL compaction to remove redundant entries
- [ ] Checksums for WAL integrity verification
- [ ] Async WAL writing for performance optimization
- [ ] Cross-node WAL synchronization during slot migration
- [ ] WAL-based snapshot generation
- [ ] Configurable WAL persistence policies

## Roadmap

- [x] ~~Data persistence to disk~~ (WAL implementation)
- [ ] Master-replica replication  
- [ ] Authentication and security
- [ ] Memory optimization
- [ ] Advanced cluster commands
- [ ] Metrics and monitoring
- [ ] Configuration files

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.