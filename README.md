# mongo-adapter

A lightweight protocol adapter for MongoDB, written in Go.  
Acts as a transparent proxy for MongoDB clients and tools, enabling features like traffic filtering, statistics, and custom protocol processing.

## Features

- **Transparent MongoDB proxy:** Connects MongoDB clients to your actual MongoDB service with no client-side changes.
- **Protocol filtering:** Supports filtering or modifying MongoDB commands in real time.
- **Traffic statistics:** Collects access and performance stats for audit or optimization.
- **Compatible with MongoDB tools:** Works with Compass, mongosh, Studio 3T, and others.
- **Written in Go:** Simple, fast, and easy to deploy.

## Getting Started

### 1. Clone the repo

```sh
git clone https://github.com/finishy1995/mongo-adapter.git
cd mongo-adapter
```

### 2. Build

```sh
go build -o mongo-adapter main.go
```

### 3. Run

```sh
./mongo-adapter \
  -loglevel=DEBUG \
  -uri="mongodb://user:pass@your-mongodb-host:27017/db" \
  -listen="127.0.0.1:27017"
```

**Command line options:**

- `-loglevel` — Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). Default: `DEBUG`
- `-uri` — MongoDB connection URI (**required**)
- `-listen` — Local listen address. Default: `127.0.0.1:27017`

### 4. Connect

Point your MongoDB client/tool at your `mongo-adapter` listen address (default `127.0.0.1:27017`).

## Project Structure

- `main.go` — Application entry point
- `network/` — Network server code
- `protocol/` — MongoDB protocol parsing and handling
- `processor/` — Command processing and filtering logic
- `library/` — Logging and utilities
- `types/` — Type definitions

## License

Apache 2.0
