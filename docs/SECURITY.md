# Security Architecture, Threat Model & Vulnerability Audit

> **Notice**: This document reflects the verified implementation in the current source code. For the comprehensive security audit, see [Security Documentation](security.md).

---

## Documentation Corrections

| Old Claim / Assumption | Current Implementation (Verified from Code) | Required Correction |
| :--- | :--- | :--- |
| **Claim**: WebSocket interceptor binds auth to the session and enforces scoping. | `WebSocketAuthInterceptor` only validates `CONNECT` frames. **It does NOT validate topic permissions on `SUBSCRIBE` frames.** Any authenticated user or agent from Company A can subscribe to `/topic/device/{id}` or `/topic/agent/{id}` of Company B if they know or guess the device UUID. | Document the critical multi-tenant subscription authorization bypass vulnerability. |
| **Claim**: Company identity is derived from JWT and used to scope devices and metrics. | `DeviceController.getMetrics(deviceId)` completely omits company verification (`deviceService.getMetrics(deviceId)` does not check `companyId`). | Document the missing tenant validation on historical metrics queries. |
| **Claim**: Rate limiting is enforced per IP and request path at the API edge. | `RateLimitFilter.shouldNotFilter()` completely **exempts** `/agent`, `/agent/*`, `/ws`, and `/ws/*`. Agents pushing metrics or attackers flooding WebSocket handshakes are not throttled. | Document that edge rate limiting only protects `/auth/*`, `/devices/*`, and `/company/*`. |
| **Claim**: Destructive commands are blocked by regex. | The agent regex (`blockedCommandPattern`) uses a naive keyword blacklist (`rm`, `del`, `format`, etc.). Attackers can easily bypass this with exfiltration commands (`curl -d @/etc/shadow`), file overwrites (`echo > /file`), or encoded payloads (`echo ... | base64 -d | sh`). | Classify regex blacklist as ineffective for security containment. |

---

## Verified Security Posture & Threat Boundaries

### 1. Ingress & Transport Layer
- **Transport Security**: TLS termination is assumed at the edge/reverse proxy (Nginx/ALB). The Spring backend runs HTTP/WS on port 8080.
- **CORS Policy**: Configured in `SecurityConfig.java` and `WebSocketConfig.java` via `app.cors.allowed-origins: http://localhost:3000`. Supports comma-separated origin patterns.
- **CSRF**: Disabled (`csrf.disable()`) across all endpoints. Appropriate for stateless JWT authentication and machine-to-machine agent traffic.
- **Rate Limiting**: Enforced in `RateLimitFilter` (Fixed Window: 120 req / 60s per IP+URI). Only applies to user REST routes.

### 2. Authentication Model
- **User / Dashboard**: HMAC-SHA256 JWT issued by `/auth/login`, signed with `app.jwt.secret` (>= 32 bytes). Validated on `/devices/**` and `/company/**` via `JwtFilter`.
- **Agent**: Machine-level `apiToken` issued per company on registration. Stored in plaintext in `company.api_token`. Passed via HTTP header `x-agent-token` or STOMP `CONNECT` frame header `x-agent-token`.

---

## Critical Forensic Vulnerabilities (Summary)

1. **Broken Object Level Authorization on STOMP Subscriptions (Critical)**:
   - File: `backend/src/main/java/com/monitor/config/WebSocketAuthInterceptor.java`
   - Issue: Clients can subscribe to any device topic across tenant boundaries without ownership validation.
2. **Missing Tenant Check on Metrics REST Endpoint (High)**:
   - File: `backend/src/main/java/com/monitor/controller/DeviceController.java:33`
   - Issue: `GET /devices/{deviceId}/metrics` allows cross-tenant historical data access.
3. **Arbitrary Command Execution Bypass (High)**:
   - File: `monitor-agent/service/commands.go:45`
   - Issue: Regex blacklist is easily bypassed, allowing arbitrary remote code execution on monitored hosts.
4. **JWT Stored in Browser `localStorage` (Medium)**:
   - File: `frontend/lib/auth.ts:8`
   - Issue: Tokens are accessible to JavaScript in the browser, vulnerable to XSS.

---

## Hardening Recommendations

1. **Enforce Topic Authorization**:
   In `WebSocketAuthInterceptor.preSend()`, intercept `StompCommand.SUBSCRIBE` frames and verify caller owns the targeted `deviceId`.
2. **Authorize `GET /devices/{deviceId}/metrics`**:
   Pass `companyId` into `deviceService.getMetrics()` and verify `device.getCompany().getId().equals(companyId)`.
3. **Replace Command Blacklist with Script Allowlist**:
   Do not accept raw shell strings over WebSocket. Allow only predefined command identifiers (e.g. `CMD_RESTART_SERVICE`, `CMD_DISK_DIAGNOSTICS`).
4. **Store JWTs in `httpOnly` Cookies**:
   Transition frontend authentication from `localStorage` to secure, HTTP-only cookies.\n