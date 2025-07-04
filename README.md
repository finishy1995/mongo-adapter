# mongo-adapter

A lightweight, stateless protocol adapter and proxy for MongoDB, written in Go.
Designed to act as a transparent proxy between MongoDB clients and servers, it enables features like traffic filtering, statistics collection, and custom command processing—all without modifying MongoDB itself or your managed MongoDB services.

Key features:

- **Stateless and highly scalable**: Deploy multiple mongo-adapter instances behind a load balancer for horizontal scaling and high availability.
- **Offload operational tasks**: Move rate limiting, traffic monitoring, and audit logging from your MongoDB server to the proxy layer—improving security and reducing load on your primary database infrastructure.
- **Protocol bridging for non-official SDKs**: Enable legacy or unofficial MongoDB clients that lack TLS/SSL support (such as some community drivers or SDKs) to connect securely to managed MongoDB services like Atlas.
- **Powerful extension point**: Add advanced features or enforce policies without altering MongoDB source code or your cloud provider’s managed service.

This makes mongo-adapter ideal for scenarios requiring flexible traffic management, security, and observability for MongoDB, with minimal operational risk and maximum deployment flexibility.

## Architecture

Normal:

![Normal Architecture](architecture/normal.drawio.png)

Cross Region(2 regions):

![Cross Region Architecture](architecture/cross-region.drawio.png)

The same principle applies to three or more regions. The deployment of MongoDB servers can be adjusted to a 1-1-1, 2-2-1, or any other configuration suitable for multi-region setups.


## Scenarios

- [x] Solving application-side lack of TLS support, e.g., enabling Atlas usage for clients that do not support TLS
- [ ] Bridging old or unofficial drivers to connect with newer versions of MongoDB (cross-version compatibility)
- [ ] Cross-database support, e.g., executing SQL statements to operate on MongoDB
- [ ] UDP and WebSocket protocol support, enabling direct MongoDB access in weak network or browser-based scenarios
- [x] Cross-region and multi-availability-zone high availability deployment, supporting pseudo-multi-region read/write
- [ ] Migration and synchronization: enable zero-downtime migrations with dual/multi-write strategies via proxy
- [ ] Audit logging, rate limiting, and tracing (tracespan)

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
  -listen="0.0.0.0:27017"
```

**Command line options:**

- `-loglevel` — Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). Default: `INFO`
- `-uri` — MongoDB connection URI (**required**)
- `-listen` — Local listen address. Default: `0.0.0.0:27017`
- `-expose` — Expose address, for compass or other tools to discovery. Default: `127.0.0.1:27017`

### 4. Connect

Point your MongoDB client/tool at your `mongo-adapter` listen address.

## License

Apache 2.0
