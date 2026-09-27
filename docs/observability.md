# Observability & Monitoring Infrastructure

## 1. Logging Forensics

### Go Monitor Agent Logging (`service/logger.go`)
- **Logging Framework**: `github.com/sirupsen/logrus`.
- **Output Routing**: Configured via `io.MultiWriter(os.Stdout, file)`. All log entries are written simultaneously to terminal stdout and local disk (`agent.log`).
- **Log Formatting**: `&logrus.TextFormatter{FullTimestamp: true}`.
- **Log Level**: Fixed to `logrus.InfoLevel`. No CLI flag or configuration option exists to enable `DebugLevel` dynamically.
- **Structured Fields**:
  - Heartbeat: `logrus.WithFields(logrus.Fields{"deviceId": cfg.DeviceID, "server": cfg.ServerURL}).Info("Agent heartbeat")`
  - Command: `logrus.WithFields(logrus.Fields{"commandType": request.Type, "commandId": request.CommandID, "status": status, "payload": ..., "snippet": ...}).Info("Remote command executed")`
  - Websocket: `logrus.WithFields(logrus.Fields{"deviceId": cfg.DeviceID, "wsUrl": wsURL}).Info(...)`

### Spring Boot Backend Logging
- **Logging Framework**: SLF4J backed by Logback.
- **Configured Level**: Standard Spring Boot INFO level.
- **SQL Logging**: `spring.jpa.show-sql: true` and `spring.jpa.properties.hibernate.format_sql: true` are enabled in `application.yml`, dumping every generated SQL query and batch insert to standard output.
- **Scheduled Tasks**:
  - `DeviceStatusScheduler`: Uses `System.out.println("Device marked OFFLINE: " + device.getHostname())` rather than structured SLF4J logging.
  - `MetricsStorageService`: Logs TimescaleDB hypertable setup and retention cleanup row counts via SLF4J Logger.

---

## 2. Health Checks & Spring Actuator

- **Dependency**: `spring-boot-starter-actuator` is present in `backend/pom.xml`.
- **Exposure**: Actuator endpoints are not explicitly customized in `application.yml`, meaning default Spring Boot behavior applies (`/actuator/health` is exposed by default).
- **Security Interception**: Because `SecurityConfig.java` defines `.anyRequest().authenticated()`, accessing `/actuator/**` requires a valid Bearer JWT unless specifically exempted.

---

## 3. Observability Capability Matrix (Can the system answer...?)

| Operational Question | Can System Answer? | Source of Truth / Evidence | Observability Gap / Recommendation |
| :--- | :--- | :--- | :--- |
| **Is the agent alive?** | **Yes** | `Device.status` in DB (`ONLINE` vs `OFFLINE`). Agent writes 20s local heartbeat logs. | Relies on 30s polling sweep in backend. No direct agent liveness probe endpoint. |
| **Is collection working?** | **Yes** | Agent logs CPU/memory metrics and sends STOMP frames; DB increments rows. | No client-side metric counter exposed via Prometheus or StatsD. |
| **Is batching delayed?** | **No** | No metric or log tracks batch flush latency or batch fill time. | Add histogram metric measuring time from first sample to batch flush. |
| **Are requests failing?** | **Partial** | Agent logs `logrus.Error("Failed to send detailed metrics batch:", err)`. | No error metrics exported to alerting systems (e.g. Prometheus alert rules). |
| **Is backend overloaded?** | **No** | Tomcat thread pool utilization and queue depth are not monitored or exported. | Enable Actuator Prometheus exporter (`micrometer-registry-prometheus`). |
| **Is DB query slow?** | **Partial** | `show-sql: true` prints queries to console. | Lacks slow-query threshold logging or APM transaction tracing. |
| **Is UI receiving stale data?** | **Yes** | Frontend tracks `lastSeenAt` from incoming frames; switches UI status if stale. | No frontend telemetry / error reporting to Sentry or OpenTelemetry. |

---

## 4. Observability Gaps & Hardening Steps

1. **Missing Metrics Endpoint on Agent**:
   - The Go agent has no local HTTP endpoint (e.g. `:9100/metrics`) to scrape internal operational metrics (goroutine count, dropped batches, socket reconnect counts).
2. **Unstructured `System.out.println` in Production Code**:
   - `DeviceStatusScheduler.java:39` outputs via `System.out.println`. This bypasses log rotation, formatting, and MDC context propagation.
3. **No Distributed Tracing**:
   - Commands dispatched from the UI have a `commandId` (UUID), but this ID is not propagated via W3C TraceContext headers (`traceparent`) to link backend logs, agent execution, and DB events together in Jaeger or OpenTelemetry.\n