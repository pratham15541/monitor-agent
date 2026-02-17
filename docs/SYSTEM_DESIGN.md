# System Design

This document describes the current system design, scale expectations based on code defaults,
operational cost drivers, and scaling options.

## Current Architecture

- Frontend: Next.js dashboard for authentication, device inventory, live metrics, detailed snapshots, and commands.
- Backend: Spring Boot REST API + STOMP WebSocket broker with JWT auth and rate limiting.
- Database: PostgreSQL storing companies, devices, metrics, and detailed snapshots.
- Agent: Go service that registers devices, streams metrics, collects detailed snapshots, and executes remote commands.

High-level flow:

- Users authenticate and access the dashboard.
- Agents register with API tokens and stream metrics via STOMP.
- Backend persists metrics and broadcasts live updates to UI subscribers.
- Detailed snapshots are sent in batches and delivered to the UI.
- Commands are routed from UI to agents over STOMP topics.

## Technologies Used

- Backend: Java 21, Spring Boot, STOMP/WebSocket, JWT, PostgreSQL
- Frontend: Next.js, React, Tailwind CSS, shadcn/ui
- Agent: Go, gopsutil, gorilla/websocket, kardianos/service

## Current Scale Expectations (Based on Defaults)

These are conservative, code-based expectations for a single backend instance:

- Metric sampling is adaptive, with a minimum 1-5 second interval and batch size of 10.
- Detailed snapshots are sent every 30 seconds by default.
- Metrics are persisted and broadcast on each batch flush.

A reasonable baseline for a single-node deployment:

- ~500-2,000 devices per backend instance at low to moderate traffic.
- ~5-20k metrics per minute total, depending on CPU load patterns.
- Detailed snapshot payloads are the main bandwidth and storage driver.

These values are estimates; validate with load tests.

## Cost Drivers

- Database storage and IOPS for metrics and detailed snapshots.
- WebSocket connection count and fan-out on the backend.
- Bandwidth for detailed snapshots, especially logs and services output.

Practical cost controls:

- Reduce detailed snapshot frequency or payload size.
- Retention policies for metrics and snapshots.
- Compression at the WebSocket or proxy layer.

## Scaling Strategy (Current Stack)

### Backend

- Add horizontal replicas behind a load balancer.
- Use sticky sessions or a shared message broker if required for WS routing.
- Consider external STOMP broker (RabbitMQ) for higher fan-out.
- Move rate limiting to API gateway for centralized policy.

### Database

- Add indexes on deviceId and createdAt for metrics and details.
- Partition tables by time (weekly/monthly) for retention.
- Use read replicas for dashboard queries.

### Agent

- Stagger detailed metrics intervals per device to reduce burst load.
- Keep adaptive sampling to reduce spikes at scale.

### Frontend

- Use pagination or windowing for large device lists.
- Cache metrics and details in the API layer if needed.

## Operational Scaling Checklist

- Add health checks, liveness probes, and autoscaling rules.
- Centralize logging and metrics for backend and agents.
- Implement retention policies (TTL) for metrics and details.
- Add request tracing for command and snapshot flows.

## Potential Improvements (Optional)

These are optional, non-breaking ideas to improve scale or reliability.

- Replace in-memory rate limiter with Redis-based limiter.
- Use Kafka or NATS for metrics ingestion.
- Move STOMP broker to RabbitMQ for large fan-out.
- Add TimescaleDB or ClickHouse for metrics storage.
- Add S3-compatible storage for large log snapshots.

## If Changing Languages or Infra (Optional)

- Backend: Go or Node.js with a dedicated WS gateway.
- DB: TimescaleDB for time-series data or ClickHouse for analytics.
- Infra: Kubernetes for autoscaling and rolling updates.
- Agent: Add gRPC streaming for lower overhead.

## Notes

- These recommendations assume current code behavior and no production load testing.
- Validate with realistic device counts and payload sizes.
