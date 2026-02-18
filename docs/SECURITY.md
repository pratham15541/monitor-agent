# Security Overview

This document explains how security is managed across the frontend, backend, and agent. It reflects the current implementation.

## Frontend security

- Auth flow uses JWT from `/auth/login`.
- JWT is stored in localStorage and attached to API requests.
- WebSocket authentication uses the Authorization header on STOMP CONNECT.
- Company API tokens are handled in-memory in UI flows and can be stored locally for convenience.

## Backend security

- JWT auth protects user endpoints.
- Company identity is derived from the JWT and used to scope devices and metrics.
- Agent access is authenticated with `x-agent-token` in REST requests and STOMP CONNECT.
- WebSocket interceptor validates JWT or agent token at CONNECT and binds auth to the session.
- Rate limiting is enforced per IP and request path at the API edge.
- JWT secret must be at least 32 bytes; base64 secrets are supported via `base64:` prefix.

## Agent security

- Agent authenticates with company API token.
- Agent registers once, then stores the device ID locally.
- Agent WebSocket sessions use `x-agent-token` during STOMP CONNECT.
- Remote commands are received only on the device-specific topic.
- Shell commands are executed locally by the agent process (no SSH in current code).

## Command safety (current)

- Destructive commands are blocked by a regex (rm, del, format, etc.).
- Shell command execution is limited to 30 seconds timeout.
- Output is chunked and streamed back to the dashboard.

## Operational guidance

- Keep JWT secret and API tokens private; rotate if leaked.
- Restrict CORS origins to trusted dashboard domains.
- Use HTTPS for API and WebSocket endpoints in production.
- Run the agent with least privilege needed for service/command execution.

## Known limitations (current)

- JWT is stored in localStorage (susceptible to XSS if present).
- No MFA or password policy enforcement is implemented in code.
- No per-command authorization policies beyond company scoping.
