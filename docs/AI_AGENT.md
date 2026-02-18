# AI Agent Guide (No Hallucination)

This guide is designed so any AI IDE can understand the project layout and existing features without guessing. Only features present in the codebase are listed under "Current Features".

## Repository map

- backend/ Spring Boot REST + STOMP WebSocket API
- frontend/ Next.js dashboard
- monitor-agent/ Go agent for metrics and commands
- docs/ Project documentation

## Current features (verified)

Backend:

- JWT auth for user login.
- Company profile endpoint for token retrieval.
- Device inventory API.
- Metrics ingestion (REST + STOMP), including batch ingestion.
- Detailed snapshots ingestion (REST + STOMP), including batch ingestion.
- Command routing over STOMP.
- Device status broadcasting and offline detection.
- Basic per-IP/per-endpoint rate limiting.
- TimescaleDB retention policies when available; fallback scheduled deletes.

Frontend:

- Company registration and login.
- Dashboard with device search, status counts, and auto-refresh.
- Device detail view with live metrics charts.
- Detailed snapshots for processes, connections, services, and logs.
- Remote commands with result history and chunked output assembly.
- Company profile view for API token.
- Agent registration screen.

Agent:

- Device registration with company API token.
- Live metrics streaming with adaptive sampling and batching.
- Detailed snapshot collection and batch upload.
- Remote command execution (shell, service control, diagnostics, collect-details).
- Runs as a background service (Windows/Linux/macOS).

## Integration points

REST endpoints:

- /auth/register, /auth/login
- /company/me
- /devices, /devices/{deviceId}/metrics, /devices/{deviceId}/metrics-detail
- /agent/register
- /agent/metrics, /agent/metrics/batch
- /agent/metrics-detail, /agent/metrics-detail/batch

WebSocket (STOMP):

- Endpoint: /ws
- Topics: /topic/device/{deviceId}, /topic/device-status/{deviceId}, /topic/device-detail/{deviceId}, /topic/command-result/{deviceId}, /topic/agent/{deviceId}
- App destinations: /app/agent/metrics, /app/agent/metrics-batch, /app/agent/metrics-detail, /app/agent/metrics-detail-batch, /app/command/{deviceId}, /app/command-result

Authentication:

- UI uses Authorization: Bearer <jwt> for REST and STOMP CONNECT.
- Agent uses x-agent-token: <api token> for REST and STOMP CONNECT.

## Data model (current)

- Company: id, email, name, passwordHash, apiToken, createdAt
- Device: id, hostname, ipAddress, os, status, lastSeenAt, createdAt, company_id
- Metric: id, device_id, cpuUsage, memoryUsage, diskUsage, networkIn, networkOut, createdAt
- MetricDetail: id, device_id, detailsJson, createdAt

## Runtime behavior (agent)

- Registration loop retries until deviceId is stored.
- Metrics loop uses WebSocket STOMP with adaptive interval (1-5s), batch size 10, and max wait 5 seconds per batch.
- Metrics batches are sent to /app/agent/metrics-batch and broadcast to the UI on /topic/device/{deviceId}.
- Detailed metrics loop uses REST batch size 1 with max wait 30 seconds; default collection interval is 30 seconds.
- Detailed metrics batches are posted to /agent/metrics-detail/batch and broadcast to the UI on /topic/device-detail/{deviceId}.
- Command loop subscribes to /topic/agent/{deviceId} and publishes results to /app/command-result.
- Destructive shell commands are blocked by regex; command timeout is 30 seconds.

## How to avoid hallucination

- Only claim features that appear in code or README files.
- Use the repository map to find the component that owns a feature.
- If unsure, inspect the implementation before listing a capability.
- Do not assume SSH support; commands are executed locally by the agent process.

## Future plans (not implemented)

Add items here only after confirming they are not yet in code.

- Multi-region deployment and global load balancing
- Advanced alerting and notification system
- RBAC and MFA for company accounts
- External message broker for WS fan-out
