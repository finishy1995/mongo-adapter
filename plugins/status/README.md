# Status Plugin for mongo-adapter

This plugin provides real-time statistics on request processing time and success rate for the Mongo protocol proxy (`mongo-adapter`). It is designed for high-concurrency environments and outputs up-to-date performance and reliability metrics.

## Features

- **Processing Time Statistics**: Tracks the total, average, minimum, and maximum time taken to process messages (in milliseconds).
- **Success Rate Monitoring**: Monitors the total number of requests, successful requests, and calculates the success ratio.
- **Real-time Logging**: Outputs a summary line every second, showing the latest status.
- **Configurable Retention**: Keeps statistics for a configurable sliding window (default 600 seconds), after which all counters are reset.
- **High Concurrency Safety**: Uses atomic operations and mutexes for lock-free or low-latency concurrent updates.