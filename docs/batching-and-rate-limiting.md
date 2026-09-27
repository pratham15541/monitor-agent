# Batching & Rate Limiting Forensics

## 1. Batching Mechanisms

There are two primary batching mechanisms implemented in the Go Monitor Agent.

---

### Mechanism A: Real-Time Metrics Batcher

- **Source Code**: `monitor-agent/service/metrics_ws.go`
- **Producer**: `CollectMetrics()` polling OS metrics at adaptive intervals (1s, 2s, or 5s).
- **Consumer**: `sendMetricsBatch(conn, batch)` streaming over STOMP WebSocket.
- **Batch Structure**:
  ```go
  var batch []map[string]interface{}
  ```
  Each element contains:
  ```json
  {
    "deviceId": "UUID",
    "cpuUsage": 14.2,
    "memoryUsage": 55.8,
    "diskUsage": 42.1,
    "networkIn": 1048576,
    "networkOut": 2097152
  }
  ```
- **Thresholds**:
  - `metricsBatchSize = 10`
  - `metricsBatchMaxWait = 5 * time.Second`
- **Flush Condition**:
  ```go
  if len(batch) >= metricsBatchSize || time.Since(lastFlush) >= metricsBatchMaxWait {
      if err := sendMetricsBatch(conn, batch); err != nil {
          return err
      }
      batch = batch[:0]
      lastFlush = time.Now()
  }
  ```
  *Flush is triggered by: N >= 10 items OR T >= 5 seconds, whichever occurs first.*
- **Memory Ownership & Buffer Management**:
  - The slice is owned exclusively by the `runMetricsSession` stack frame.
  - On flush, `batch = batch[:0]` resets length to 0 while preserving the underlying array allocation.
- **Failure Behavior**:
  - If `sendMetricsBatch` returns an error (e.g. socket broken), `runMetricsSession` returns `err`.
  - The connection is closed.
  - **Data Loss Finding**: The in-memory slice `batch` is garbage-collected. Pending samples in the batch are **permanently dropped**.
- **Shutdown Behavior**:
  - If `stop` channel closes, `runMetricsSession` returns `nil` immediately.
  - No shutdown flush barrier is executed. Any uncommitted items in `batch` are discarded.

---

### Mechanism B: Detailed Metrics Snapshot Batcher

- **Source Code**: `monitor-agent/service/metrics_detail_ws.go`
- **Producer**: `collectDetailedMetricsPayload(cfg)` executed on an interval timer (default 30 seconds).
- **Consumer**: `sendDetailedMetricsBatch(cfg, batch)` via HTTP POST to `/agent/metrics-detail/batch`.
- **Batch Structure**:
  ```go
  var batch []map[string]interface{}
  ```
  Each element contains:
  ```json
  {
    "deviceId": "UUID",
    "details": {
      "collectedAt": "2026-09-27T12:00:00Z",
      "processes": [...],
      "connections": [...],
      "memory": {...},
      "services": {...},
      "logs": {...},
      "os": "linux"
    }
  }
  ```
- **Thresholds**:
  - `detailBatchSize = 1`
  - `detailBatchMaxWait = 30 * time.Second`
- **Flush Condition**:
  ```go
  if len(batch) >= detailBatchSize || time.Since(lastFlush) >= detailBatchMaxWait {
      if err := sendDetailedMetricsBatch(cfg, batch); err != nil {
          logrus.Error("Failed to send detailed metrics batch:", err)
      } else {
          batch = batch[:0]
          lastFlush = time.Now()
      }
  }
  ```
- **Failure Behavior & Backpressure Risk**:
  - Notice that on error, `batch = batch[:0]` is **NOT** called!
  - If the backend is down or returns HTTP 500/503, the failed payload remains in `batch`.
  - On the next 30-second tick, another ~1MB–5MB detailed payload is appended: `batch = append(batch, payload)`.
  - **Memory Leak & Unbounded Growth Finding**: Under prolonged backend outages, `batch` grows monotonically in memory, leading to potential Out-Of-Memory (OOM) termination on resource-constrained hosts.
- **Shutdown Behavior**:
  - When `<-stop` triggers, it calls `_ = sendDetailedMetricsBatch(cfg, batch)` before returning, attempting a best-effort terminal flush.

---

## 2. Rate Limiting Architecture

Rate limiting is enforced at the backend edge using a custom Servlet filter.

### Configuration & Bean Definition
- **Source Code**: `backend/src/main/java/com/monitor/config/RateLimitFilter.java`
- **Algorithm**: In-Memory Fixed Window Counter.
- **Properties**:
  - `app.rateLimit.windowSeconds: 60` (Default 60 seconds)
  - `app.rateLimit.maxRequests: 120` (Default 120 requests per window)

### Rate Limiting Key Construction
```java
private String buildKey(HttpServletRequest request) {
    String ip = request.getHeader("X-Forwarded-For");
    if (ip == null || ip.isBlank()) {
        ip = request.getRemoteAddr();
    } else {
        ip = ip.split(",")[0].trim();
    }
    return ip + "|" + request.getRequestURI();
}
```
The rate limit bucket is keyed by client IP + exact request URI.

### Enforcement Logic
```java
long now = System.currentTimeMillis();
Counter counter = counters.compute(key, (k, existing) -> {
    if (existing == null || now - existing.windowStartMs >= windowSeconds * 1000) {
        return new Counter(now, 1);
    }
    existing.count++;
    return existing;
});

if (counter.count > maxRequests) {
    response.setStatus(HttpStatus.TOO_MANY_REQUESTS.value()); // HTTP 429
    return;
}
```

### Rate Limiter Exclusions & Bypass Rules
In `RateLimitFilter.shouldNotFilter()`:
1. **Agent REST Ingestion Exemption**:
   ```java
   if (path.equals("/agent") || path.startsWith("/agent/") || path.equals("/ws") || path.startsWith("/ws/")) {
       return true;
   }
   ```
   **CRITICAL FINDING**: The rate limiter is completely disabled for `/agent/**` and `/ws/**`! Agents pushing metric or detail batches are NEVER rate-limited by the backend.
2. **Load Tester Bypass Header**:
   ```java
   String loadTester = request.getHeader("X-Load-Tester");
   if ("true".equalsIgnoreCase(loadTester)) return true;
   ```
3. **Load Tester User Agent Bypass**:
   ```java
   String userAgent = request.getHeader("User-Agent");
   return userAgent != null && userAgent.startsWith("monitor-loadtester/");
   ```

### Protection Scope Summary
- **Protected**: User / Dashboard REST endpoints (`/auth/login`, `/auth/register`, `/devices`, `/devices/{id}/metrics`, `/company/me`).
- **Unprotected**: All agent ingestion endpoints (`/agent/register`, `/agent/metrics`, `/agent/metrics/batch`, `/agent/metrics-detail`, `/agent/metrics-detail/batch`) and all STOMP WebSocket traffic (`/ws`).
- **Memory Concern**: The `counters` map (`ConcurrentHashMap<String, Counter>`) does not implement TTL eviction or LRU cleanup. Distinct IP+URI keys accumulate indefinitely.\n