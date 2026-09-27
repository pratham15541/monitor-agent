# Performance, Resource Utilization & Scalability Analysis

## 1. Subsystem Performance Profiles

### A. Go Monitor Agent
- **CPU Footprint**:
  - Metric sampling: `gopsutil/v3/cpu.Percent(0, false)` takes a 0-duration snapshot using OS kernel delta counters (negligible CPU impact: < 0.2% on modern cores).
  - Adaptive sampling throttles collection from 5s down to 1s only when CPU usage exceeds 90%, preventing the monitor from starving the host under extreme load.
  - Process table gathering (`collectProcessMetrics`): Iterates over all active PIDs in `/proc` (Linux) or via Toolhelp32 snapshot (Windows), fetching RSS, VMS, cmdline, and threads. On hosts with >1,000 running processes, this is the most CPU-intensive collector (approx 15ms–40ms CPU time per 30-second interval).
- **Memory Footprint**:
  - Baseline idle agent RSS: ~12MB–25MB.
  - During detail snapshot serialization: Allocates JSON string buffers for ~200 sockets and active processes (~2MB–5MB transient heap allocations, quickly garbage collected).
  - **Memory Leak Hazard**: In `service/metrics_detail_ws.go`, if HTTP POST fails, `batch` slice is not cleared. Under prolonged disconnects, memory grows monotonically.
- **Network I/O**:
  - Real-time metrics: 1 STOMP frame every 5s (~1.2KB payload) = ~240 bytes/second egress bandwidth.
  - Detail snapshots: 1 HTTP POST every 30s (~150KB–500KB JSON payload) = ~5KB–16KB/second egress bandwidth.

### B. Spring Boot Backend
- **Thread Concurrency**:
  - Ingestion throughput is constrained by HikariCP database pool size (default: 10 connections) and Tomcat worker threads (default: 200).
  - Ingesting a batch of 10 metrics requires:
    1. Resolving device record by UUID (`deviceRepository.findById`).
    2. Batch insert of 10 entities into TimescaleDB chunk table (`metricRepository.saveAll`).
    3. Update `device.last_seen_at` and `device.status`.
    4. SimpleBroker WebSocket fan-out to connected subscribers.
  - Transaction duration per batch: ~2ms–8ms on SSD storage.
- **Database Storage Growth**:
  - Each metric row occupies ~72 bytes on disk.
  - 1 device generating 12 metrics/minute = 17,280 rows/day = ~1.24 MB/day.
  - 1,000 devices = 17.28 million rows/day = ~1.24 GB/day (uncompressed).
  - TimescaleDB native compression (if configured) can reduce time-series footprint by 85–90%.
  - Detailed snapshots: 1 device = 2,880 snapshots/day @ 200KB avg = ~576 MB/day. 1,000 devices = ~576 GB/day. **Detailed snapshots are the single largest storage and bandwidth cost driver in the entire system.**

---

## 2. Scalability Architecture & Bottleneck Modeling

The following table evaluates architectural behavior from 1 to 10,000 concurrently active agents on a single backend node:

| Fleet Size | Metric Ingestion Rate | Detail Ingestion Rate | WebSocket Connections | Primary Bottleneck | System Behavior & Failure Mode |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1 Agent** | ~2 metrics / sec | 2 snapshots / min | 2 persistent WS | None | Sub-millisecond response times. Negligible resource usage. |
| **10 Agents** | ~20 metrics / sec | 20 snapshots / min | 20 persistent WS | None | Minimal database IOPS. Total memory ~500MB JVM heap. |
| **100 Agents** | ~200 metrics / sec | 200 snapshots / min | 200 persistent WS | DB connection pool / disk writes | HikariCP pool (10 conns) handles load smoothly. DB writes ~12MB/min. |
| **1,000 Agents** | ~2,000 metrics / sec | 2,000 snapshots / min | 2,000 persistent WS | **`DeviceStatusScheduler` full table scan** & Detail storage | **CRITICAL BOTTLENECK**: `DeviceStatusScheduler` executes `findAll()` on 1,000 rows every 30s. Detail ingestion hits ~10MB/sec HTTP ingress. |
| **10,000 Agents**| ~20,000 metrics / sec| 20,000 snapshots / min| 20,000 persistent WS | **Tomcat thread pool & SimpleBroker memory** | **SYSTEM COLLAPSE**: Single JVM Tomcat exceeds max threads. In-memory `SimpleBroker` exhausts heap. PostgreSQL connection pool starves. `findAll()` locks scheduler. |

---

## 3. Scale-Out & Production Hardening Roadmap

1. **Replace In-Memory SimpleBroker with External Message Broker**:
   - Transition Spring WebSocket configuration to use `enableStompBrokerRelay("/topic")` backed by a RabbitMQ or Apache ActiveMQ cluster. This enables multi-node horizontal scaling of the Spring Boot backend behind an AWS ALB or Nginx reverse proxy.
2. **Optimize `DeviceStatusScheduler`**:
   - Replace `deviceRepository.findAll()` with a targeted indexed query:
     ```sql
     UPDATE device 
     SET status = 'OFFLINE' 
     WHERE status = 'ONLINE' 
       AND last_seen_at < (NOW() - INTERVAL '30 seconds') 
     RETURNING id, hostname;
     ```
   - This eliminates full table scans, avoids hydrating thousands of JPA entities into JVM memory, and executes as an atomic single-statement SQL update.
3. **Partition Detail Snapshot Storage**:
   - Transition large `details_json` blobs out of relational PostgreSQL storage into S3 / MinIO object storage, storing only an S3 URI / hash in the relational table.\n