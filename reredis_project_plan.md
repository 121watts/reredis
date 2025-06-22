# Reredis Project Plan

## 🧠 Overview
Reredis is a lightweight, Redis-like key-value store written in Go with support for TCP connections and real-time visibility through a WebSocket-enabled frontend.

---

## ✅ Phase 1: Core In-Memory Redis Clone (Go)

### Features:
- TCP server that accepts multiple connections
- Supports `SET` and `GET` commands
- Shared, thread-safe in-memory store (using `sync.Mutex`)
- Text-based protocol similar to Redis (line-delimited)

### Milestones:
- [x] Project scaffolding with idiomatic Go structure
- [x] Internal packages (`store`, `server`)
- [x] TCP listener and goroutine-per-connection handling
- [x] Implement `SET` command
- [x] Implement `GET` command
- [x] Integration tests with dynamic port allocation

---

## 🔌 Phase 2: WebSocket Observer (Go)

### Goal:
Allow a connected frontend to observe updates to the in-memory store in real time.

### Features:
- WebSocket server integrated into existing Go backend
- Broadcast updates on `SET` to all connected observers
- JSON-encoded update payloads

### Milestones:
- [ ] Define `observer` package for managing WebSocket clients
- [ ] Add WebSocket endpoint (e.g. `/ws`)
- [ ] Hook into `Set()` to trigger broadcast messages
- [ ] Write integration test for WebSocket update flow

---

## 🌐 Phase 3: Frontend Viewer (React + Tailwind)

### Goal:
Create a simple UI to view and interact with the key-value store in real time.

### Features:
- React + Vite + Tailwind CSS stack
- WebSocket client connects to Go backend
- Live table view of all keys and values
- Input to `GET`/`SET` values

### Milestones:
- [ ] Project scaffold with Vite
- [ ] Connect to `/ws` WebSocket and display state
- [ ] Add input form for `SET`/`GET` commands
- [ ] Style layout for responsiveness

---

## 🔄 Phase 4: Optional Enhancements

### Features:
- TTL support
- Persistence (AOF or RDB-style)
- LRU support
- Pub/Sub channels
- Authentication (e.g. shared secret)
- Dockerfile + docker-compose setup for dev/test

---

## 🧪 Testing Strategy

- Unit tests for `store` logic
- Integration tests for TCP command flow
- WebSocket broadcast verification
- Frontend e2e tests with mocked WebSocket server

---

## 📁 Folder Structure

```
reredis/
├── cmd/
│   └── reredis/         # Main binary
├── internal/
│   ├── server/          # TCP + WebSocket logic
│   ├── store/           # Shared key-value logic
│   └── observer/        # WebSocket broadcast clients (TBD)
├── frontend/            # React app (TBD)
└── tests/               # Integration + e2e tests (optional)
```

---

## 🏁 Current Status
- Core TCP server ✅
- SET + GET + DEL implemented ✅
- Integration tests ✅
- WebSocket + frontend work upcoming ⏳
