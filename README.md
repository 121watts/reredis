# Reredis

A lightweight Redis-compatible key-value store written in Go with real-time WebSocket capabilities and distributed clustering support.

## Features

- **Redis Protocol Compatibility**: Supports core Redis commands (SET, GET, DEL)
- **Real-time Updates**: WebSocket server for live data synchronization
- **Distributed Clustering**: Redis-compatible clustering with automatic slot distribution
- **Web Interface**: React frontend for visualizing key-value data
- **Concurrent Architecture**: Goroutine-per-connection model for high performance

## Architecture

Reredis consists of three main components:

1. **TCP Server** (port 6379+): Implements Redis protocol for client connections
2. **HTTP/WebSocket Server** (port 8080+): Provides real-time updates and web interface  
3. **Cluster Manager**: Handles distributed hash slot assignment and node discovery

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
│   ├── store/            # In-memory key-value store
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

## Roadmap

- [ ] Data persistence to disk
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