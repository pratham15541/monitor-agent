# Configuration Forensics & Parameter Reference

## 1. Overview of Configuration Sources

Configuration across the Monitor Agent stack is loaded from four primary sources:
1. **Spring Boot YAML / Environment**: `backend/src/main/resources/application.yml`.
2. **Go Agent JSON Config File**: Disk-persisted JSON file (`config.json` or `~/.monitor-agent.json`).
3. **Frontend Runtime Environment**: `.env` and Next.js public variables (`NEXT_PUBLIC_*`).
4. **Load Tester YAML Configuration**: Config passed via `--config <file.yml>`.

---

## 2. Spring Boot Backend Configuration (`application.yml`)

| Property | Environment Variable | Default Value | Type | Description & Validation |
| :--- | :--- | :--- | :--- | :--- |
| `spring.datasource.url` | None | `jdbc:postgresql://postgres:5432/monitor` | String | PostgreSQL / TimescaleDB JDBC connection URL. |
| `spring.datasource.username`| None | `monitor` | String | Database user. |
| `spring.datasource.password`| None | `monitor` | String | Database password. |
| `server.port` | `SERVER_PORT` | `8080` | Integer | HTTP/WebSocket listen port. |
| `app.jwt.secret` | `JWT_SECRET` | `uZ9kYx7mFJm6yK3pQvT8cW2LrN5sD4eHjA1bC6xZpQ0` | String | HS256 HMAC Secret Key. Must be >= 32 bytes or prefixed with `base64:`. Fails startup if < 32 bytes. |
| `app.jwt.issuer` | `JWT_ISSUER` | `monitor-tool` | String | JWT issuer claim (`iss`). |
| `app.jwt.expirationMinutes`| `JWT_EXP_MINUTES` | `60` | Long | Expiration time of user authentication tokens. |
| `app.cors.allowed-origins` | `CORS_ALLOWED_ORIGINS`| `http://localhost:3000` | String | Comma-separated list of allowed browser origins. |
| `app.rateLimit.windowSeconds`| `RATE_LIMIT_WINDOW`| `60` | Long | Fixed window duration for edge rate limiter. |
| `app.rateLimit.maxRequests` | `RATE_LIMIT_MAX` | `120` | Integer | Maximum HTTP requests per IP+URI per window. |
| `app.metrics.retentionDays` | `METRIC_RETENTION_DAYS`| `30` | Integer | Retention window for time-series metrics. |
| `app.metrics.detailRetentionDays`| `METRIC_DETAIL_RETENTION_DAYS`| `7` | Integer | Retention window for detailed snapshots. |

---

## 3. Go Monitor Agent Configuration (`config.json`)

### File Resolution Rules
1. `MONITOR_AGENT_CONFIG` environment variable (if non-empty).
2. On Windows: `%ProgramData%\MonitorAgent\config.json` (defaults to `C:\ProgramData\MonitorAgent\config.json`).
3. On Linux/macOS: `$HOME/.monitor-agent.json`.

### Schema & Fields

```json
{
  "serverUrl": "http://localhost:8080",
  "token": "7b89d44e-128a-4d43-982d-114498ec5123",
  "deviceId": "9c3e2182-3d8b-4977-bc6d-0e42a98f411b"
}
```

| Field | Source | Written By | Description | Failure Behavior |
| :--- | :--- | :--- | :--- | :--- |
| `serverUrl` | CLI flag / command | `monitor-agent install --server`, `set-url` | Base URL of backend REST & WebSocket. | Defaults to empty string. WebSocket connection fails if empty. |
| `token` | CLI flag / command | `monitor-agent install --token`, `set-token` | Company API token used for registration and WebSocket authentication. | `service.StartWorker` calls `logrus.Fatal()` if token is empty. |
| `deviceId` | Server registration | `service.RegisterIfNeeded()` | Unique UUID assigned by backend. | If empty, agent loops and retries registration. |

---

## 4. Next.js Frontend Configuration (`.env`)

| Variable | Default Value | Consumed In | Description |
| :--- | :--- | :--- | :--- |
| `NEXT_PUBLIC_API_BASE` | `http://localhost:8080` | `lib/api.ts` | Base REST API endpoint for browser fetch requests. |
| `NEXT_PUBLIC_WS_URL` | `http://localhost:8080/ws` | `devices/[id]/page.tsx` | SockJS / STOMP endpoint URL. Normalized to `http`/`https` for SockJS. |

---

## 5. Environment-Specific Differences (Dev vs Docker vs Prod)

| Dimension | Local Development | Docker Compose (`backend/docker-compose.yml`) | Production Recommendations |
| :--- | :--- | :--- | :--- |
| **Database Host** | `localhost:5432` | `postgres:5432` (`container_name: monitor-postgres`) | Managed PostgreSQL / TimescaleDB cluster |
| **JWT Secret** | Default in-memory key | Environment variable | Secure 64-byte random string stored in secrets manager |
| **CORS Origins** | `http://localhost:3000` | `http://localhost:3000` | Restricted to production domain (e.g. `https://monitor.example.com`) |
| **Agent Executable** | Built via `go build` | Built via `scripts/build_agent.sh` | Cross-compiled releases via GitHub Actions with signed checksums |
| **Logging** | Console stdout | Container stdout / docker logs | Structured JSON shipping to centralized log aggregator (ELK, Loki) |\n