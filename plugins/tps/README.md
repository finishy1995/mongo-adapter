# TPS Plugin for mongo-adapter

This plugin is used to collect TPS (Transactions Per Second) statistics for the Mongo protocol proxy (`mongo-adapter`) under high concurrency scenarios, providing real-time output and a historical time-window summary.

## Features

- **Real-time statistics**: Outputs the global TPS for each operation type (find, insert, update, delete) every second.
- **Per-collection distribution**: Uses `collection+operation+timestamp` as the key, supporting statistics by collection and operation.
- **Historical retention**: Configurable data retention window (default 600 seconds); outdated data is automatically cleaned up.
- **High concurrency safety**: Implemented with `sync.Map + atomic` for lock-free, high-performance multi-threaded writes.