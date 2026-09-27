# Architecture Decision Records (ADRs) Reverse-Engineered from Source

This document captures the explicit and implicit architectural decisions reflected in the current implementation.

---

## ADR-01: Dual-Protocol Ingestion (WebSocket STOMP for Metrics + HTTP REST for Snapshots)

- **Status**: Implemented in current source.
- **Context**: The monitoring agent collects two classes of telemetry: lightweight high-frequency numeric counters (CPU, memory, disk, network) every 1–5 seconds, and heavyweight diagnostics (process trees, sockets, service outputs, logs) every 30 seconds.
- **Decision**:
  - Stream numeric metrics over a persistent STOMP WebSocket connection (`/app/agent/metrics-batch`).
  - Transmit heavyweight snapshots over standard HTTP/1.1 POST requests (`/agent/metrics-detail/batch`).
- **Consequences**:
  - *Positive*: Protects the WebSocket frame channel from buffer bloat and head-of-line blocking caused by multi-megabyte snapshot payloads.
  - *Negative*: Requires managing two distinct network channels and authentication paths (STOMP frames with `x-agent-token` vs HTTP headers with `x-agent-token`).

---

## ADR-02: Unified Go Binary Serving Both CLI and Background Service

- **Status**: Implemented in current source.
- **Context**: Monitored machines need both automated daemon execution (managed by `systemd`, Windows SCM, or `launchd`) and interactive administrator commands (`install`, `set-token`, `status`).
- **Decision**:
  - Compile a single unified Go binary (`monitor-agent`).
  - At process boot, invoke `kardianos/service.Interactive()`. If false, launch the background daemon supervisor; if true, delegate to Cobra CLI.
- **Consequences**:
  - *Positive*: Dramatically simplifies distribution, deployment, and packaging: single binary download per OS/architecture.
  - *Negative*: Both daemon and CLI share the same binary dependencies and size.

---

## ADR-03: In-Memory SimpleBroker over External Message Bus

- **Status**: Implemented in current source.
- **Context**: The backend must distribute real-time metric updates and command streams between agents and browser dashboards.
- **Decision**:
  - Use Spring Boot's built-in `SimpleBroker` prefixing `/topic` without configuring an external broker relay (e.g. RabbitMQ or Kafka).
- **Consequences**:
  - *Positive*: Zero external messaging infrastructure required for single-node development and small deployments.
  - *Negative*: Prevents horizontal clustering of backend nodes. If multiple Spring Boot instances run behind a round-robin load balancer, an agent connected to Instance A cannot stream metrics to a browser connected to Instance B.

---

## ADR-04: Dynamic Schema Alteration and Hypertables via `MetricsStorageService`

- **Status**: Implemented in current source.
- **Context**: Time-series telemetry requires high write throughput and automated chunk retention, but standard JPA/Hibernate tools lack native awareness of TimescaleDB hypertables.
- **Decision**:
  - Let Hibernate generate standard tables (`ddl-auto: update`).
  - On `ApplicationReadyEvent`, execute raw SQL DDL to drop primary keys, alter timestamps to `TIMESTAMPTZ`, and invoke TimescaleDB's `create_hypertable()` and `add_retention_policy()`.
  - Fall back to a scheduled cron job (`0 30 2 * * *`) if the TimescaleDB extension is missing.
- **Consequences**:
  - *Positive*: Out-of-the-box support for TimescaleDB with zero manual database migration scripts. Transparent degradation on standard PostgreSQL.
  - *Negative*: Dropping primary keys at runtime violates standard JPA expectations (entities operate without DB-enforced PKs). DDL execution on application startup slows boot and can conflict in multi-node clusters.

---

## ADR-05: In-Memory Fixed-Window Edge Rate Limiting with Ingestion Exemption

- **Status**: Implemented in current source.
- **Context**: The backend REST API must be protected against brute-force authentication and denial-of-service query spam.
- **Decision**:
  - Implement a custom Servlet filter (`RateLimitFilter`) using an in-memory `ConcurrentHashMap`.
  - Enforce a 60-second window with 120 requests limit per IP+URI.
  - Explicitly exempt all `/agent/**` and `/ws/**` paths from rate limiting.
- **Consequences**:
  - *Positive*: Protects user authentication and device query endpoints from browser abuse. Ensures agent telemetry streams are never throttled or dropped by the edge rate limiter.
  - *Negative*: Agent endpoints are completely unprotected against volume floods. Rate limiter state is local to a single JVM node and the in-memory map has no eviction policy.\n