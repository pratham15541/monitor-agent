# Failure Scenarios, Crash Consistency & Recovery

## 1. Failure Scenario Matrix

| Failure Mode | Direct Consequence | Agent Behavior | Backend Behavior | Frontend Behavior | Data Loss / Duplicate Risk |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Agent Process Crash** | Telemetry and command loops cease immediately. | Process terminates. Service manager restarts if configured. | `DeviceStatusScheduler` marks device `OFFLINE` after 30s. | Real-time chart freezes. Status badge turns red `OFFLINE`. | In-flight unsent batch in memory is lost. |
| **Backend Process Crash** | All WebSocket connections drop. REST calls return connection refused. | `Dial` and `postJSON` fail. Loops sleep 3s–5s and retry indefinitely. | Process dies. Tomcat closes sockets. | UI shows "Websocket not connected". RealtimeState = "disconnected". Auto-reconnect attempts every 5s. | In-memory metric batches in agent lost. Detail loop accumulates snapshots in memory. |
| **Database Outage / Unreachable** | Backend cannot execute queries. | HTTP POST calls receive HTTP 500. STOMP message ingestion fails silently or logs errors. | Spring `@Transactional` rolls back. Exception handler logs error. | Data fetching fails with error banner: "Request failed with 500". | Ingestion transactions roll back. Ingested data rejected. |
| **Network Split (Agent <-> Backend)** | TCP socket stalls or resets. HTTP calls time out (10s). | Metrics loop catches error, closes socket, drops in-memory batch, sleeps 3s, redials. | Session terminated. `DeviceStatusScheduler` marks device `OFFLINE` after 30s. | Device marked `OFFLINE`. | **Data Loss**: Dropped metric batches. **Memory Risk**: Growing detail snapshot slice in agent. |
| **Slow Backend / DB Latency Spike** | Ingestion duration > sampling interval. | TCP send buffers fill. Detail HTTP requests exceed 10s timeout. | HikariCP pool exhaustion possible. Request threads block. | Dashboard queries slow down or time out. | Timeouts cause dropped batches on agent side. |
| **Agent Machine Power Loss** | Immediate host shutdown. | None (hard kill). | Backend marks `OFFLINE` after 30s. | Status switches to `OFFLINE`. | Buffered metrics in agent RAM lost. |

---

## 2. Crash Consistency Analysis

The following diagram analyzes crash consistency at each state boundary of a metric batch:

```text
[1. Sample Collected]
       │  (State: In Go memory slice)
       │  *CRASH HERE*: Sample lost cleanly. No duplicates.
       ▼
[2. Batch Formed (N=10)]
       │  (State: In Go memory slice)
       │  *CRASH HERE*: Entire 10-item batch lost. No duplicates.
       ▼
[3. STOMP SEND Frame Written to Socket]
       │  (State: In transit in OS kernel socket buffers)
       │  *CRASH HERE*: Frame may or may not reach backend.
       ▼
[4. Backend Receives Frame & Begins @Transactional]
       │  (State: Tomcat worker thread, active DB transaction)
       │  *CRASH HERE*: Transaction rolls back cleanly. Data not committed.
       ▼
[5. Database Commits (metricRepository.saveAll)]
       │  (State: Committed to TimescaleDB write-ahead log & table)
       │  *CRASH HERE*: Metrics are persisted.
       ▼
[6. Broker Broadcasts to /topic/device/{id}]
       │  (State: Sent to connected browser sessions)
       │  *CRASH HERE*: Database has data, but UI did not receive live event.
       ▼
[7. Frontend Receives STOMP Frame & Updates React State]
```

### Crash Invariants
1. **No Write-Ahead Logging on Agent**: The Go agent maintains no local SQLite, BadgerDB, or WAL log file. Any crash of the agent host or daemon results in the immediate, unrecoverable loss of all metrics buffered since the last successful flush.
2. **Atomic Ingestion per Batch**: Metric batches ingested by `AgentService.saveMetricsBatch()` are wrapped in `@Transactional`. Either all metrics in the batch are committed to PostgreSQL, or none are.
3. **No Distributed Commit Coordination**: The agent sends STOMP `SEND` frames asynchronously without waiting for a STOMP `RECEIPT` frame. The agent does not know whether the backend successfully committed the batch to the database.

---

## 3. Retries & Backoff Table

| Subsystem | Operation | Retry Implemented? | Retry Trigger / Conditions | Maximum Attempts | Delay & Backoff Algorithm | Failure Behavior |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Go Agent** | Device Registration (`RegisterIfNeeded`) | **Yes** | `err != nil` or HTTP non-2xx | Infinite (in background loop) | Fixed 10s–15s sleep | Logs warning, sleeps, retries on next tick. |
| **Go Agent** | Metrics WebSocket Session (`runMetricsSession`) | **Yes** | Connection drop, dial error, STOMP error | Infinite | Fixed 3s sleep | Discards batch, closes socket, sleeps 3s, redials. |
| **Go Agent** | Detailed Snapshot POST (`sendDetailedMetricsBatch`) | **Partial** | HTTP failure, network error | Infinite (appends next tick) | Appends new snapshot every 30s | Retains failed items in memory, appends next snapshot, retries combined batch. |
| **Go Agent** | Command WebSocket Session (`runCommandSession`) | **Yes** | Connection drop, dial error | Infinite | Fixed 3s sleep | Closes socket, sleeps 3s, redials, resubscribes to `/topic/agent/{id}`. |
| **Frontend** | Dashboard Device Auto-Refresh | **Yes** | `autoRefresh: true` | Infinite | Fixed 15s `setInterval` | Discards error, attempts fresh fetch on next interval. |
| **Frontend** | Detailed Snapshot Auto-Refresh | **Yes** | Active tab = "detailed" | Infinite | Fixed 15s `setInterval` | Discards error, attempts fresh fetch on next interval. |
| **Frontend** | STOMP WebSocket Client | **Yes** | Socket disconnect, network error | Infinite | Fixed 5000ms (`reconnectDelay: 5000`)| Attempts SockJS handshake every 5s. |

---

## 4. Timeout Inventory

| Subsystem | Operation / Link | Timeout Duration | Enforcement Mechanism | Consequence on Expiry |
| :--- | :--- | :--- | :--- | :--- |
| **Go Agent** | HTTP Requests (`service/client.go`) | **10 seconds** | `http.Client.Timeout = 10 * time.Second` | Aborts HTTP request, returns context deadline error. |
| **Go Agent** | Shell Command Execution (`runShellCommand`) | **30 seconds** | `context.WithTimeout(ctx, 30*time.Second)` | Process killed via SIGKILL / TerminateProcess; status = `"timeout"`. |
| **Go Agent** | Service Command Inspection (`collectServicesSnapshot`)| **5 seconds** | `context.WithTimeout(ctx, 5*time.Second)` | Process killed; returns partial or empty string. |
| **Go Agent** | System Log Fetching (`readSystemLogs`) | **5 seconds** | `context.WithTimeout(ctx, 5*time.Second)` | Process killed; returns partial or empty string. |
| **Go Agent** | Shutdown Grace Period (`cmd/run.go`) | **1 second** | `time.Sleep(1 * time.Second)` | Process exits regardless of in-flight goroutine progress. |
| **Spring Boot** | STOMP WebSocket Send Buffer | **30 seconds** | `registry.setSendTimeLimit(30 * 1000)` | Session terminated by broker if client send buffer stalls for 30s. |
| **Spring Boot** | Device Offline Detection Window | **30 seconds** | `device.getLastSeenAt().isBefore(now - 30s)` | Device status marked `OFFLINE`; broadcast event sent to UI. |
| **Frontend** | Detail Snapshot Cache TTL | **5–10 seconds**| `Date.now() - cached.fetchedAt < 10000` | Frontend skips remote HTTP request and serves from memory. |\n