# AI Agent Guide (No Hallucination)

This guide is designed so any AI IDE can understand the project layout and existing features
without guessing. Only features present in the codebase are listed under "Current Features".

## Repository Map

- backend/ Spring Boot REST + STOMP WebSocket API
- frontend/ Next.js dashboard
- monitor-agent/ Go agent for metrics and commands
- docs/ Project documentation

## Current Features (Verified)

Backend:

- JWT auth for user login
- Company profile endpoint for token retrieval
- Device inventory API
- Metrics ingestion (REST + STOMP), including batch ingestion
- Detailed snapshots ingestion (REST + STOMP), including batch ingestion
- Command routing over STOMP
- Device status broadcasting and offline detection
- Basic per-IP/per-endpoint rate limiting

Frontend:

- Company registration and login
- Dashboard with device search, status counts, and auto-refresh
- Device detail view with live metrics charts
- Detailed snapshots for processes, connections, services, and logs
- Remote commands with result history
- Company profile view for API token
- Agent registration screen

Agent:

- Device registration with company API token
- Live metrics streaming with adaptive sampling and batching
- Detailed snapshot collection and batch upload
- Remote command execution (shell, service control, diagnostics, collect-details)
- Runs as a background service (Windows/Linux/macOS)

## Integration Points

- REST endpoints: /auth/_, /company/me, /devices/_, /agent/\*
- WebSocket: /ws (STOMP)
- Topics: /topic/device/_, /topic/device-status/_, /topic/device-detail/_, /topic/command-result/_, /topic/agent/\*
- App destinations: /app/agent/metrics, /app/agent/metrics-batch, /app/agent/metrics-detail, /app/agent/metrics-detail-batch, /app/command/{deviceId}, /app/command-result

## How to Avoid Hallucination

- Only claim features that appear in code or README files.
- Use the repository map to find the component that owns a feature.
- If unsure, inspect the implementation before listing a capability.

## Future Plans (Not Implemented)

Add items here only after confirming they are not yet in code.

- Multi-region deployment and global load balancing
- Metrics retention policies and archival storage
- Advanced alerting and notification system
- RBAC and MFA for company accounts
- External message broker for WS fan-out
