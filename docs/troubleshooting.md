# Operational Troubleshooting & Diagnostics Guide

## 1. Diagnostic Matrix: Symptoms & Solutions

| Symptom | Probable Root Cause | Verification Step | Resolution |
| :--- | :--- | :--- | :--- |
| **Agent logs "Token not set" and terminates immediately** | `config.json` does not exist or has an empty `"token"` field. | Run `monitor-agent status`. | Run `monitor-agent install --token <TOKEN>` or `monitor-agent set-token <TOKEN>`. |
| **Agent continuously logs "Attempting to register device... Registration failed"** | Backend server is unreachable, or company API token is invalid/revoked. | Inspect `agent.log` for HTTP status code or dial error. Test `curl -X POST http://backend:8080/agent/register`. | Verify backend is running on target port and verify token in dashboard Company profile. |
| **Agent connected, but device remains OFFLINE in dashboard** | Time drift between agent host and backend, or metrics batch sending fails. | Check `Device.lastSeenAt` in database vs backend `now()`. | Ensure both machines are synchronized via NTP. Verify STOMP `/app/agent/metrics-batch` is receiving frames. |
| **Dashboard shows "Websocket not connected" banner** | Backend WebSocket `/ws` is down, or CORS origin rejected, or JWT expired. | Open browser DevTools Network tab. Check WebSocket handshake status. | Verify `app.cors.allowed-origins` includes frontend origin (e.g. `http://localhost:3000`). Re-login to refresh JWT. |
| **Remote command output truncated or stuck in "stream" status** | Intermediate command chunk was dropped over network, or agent timed out. | Check browser console and `agent.log`. | Check if command exceeded 30s timeout. Re-run command or execute diagnostics. |
| **Backend logs "TimescaleDB not available, falling back to scheduled deletes"** | PostgreSQL container is standard Postgres rather than TimescaleDB image. | Run `SELECT * FROM pg_extension WHERE extname = 'timescaledb';`. | Use `timescale/timescaledb:2.13.1-pg16` Docker image. |
| **Backend returns HTTP 429 Too Many Requests** | Client exceeded 120 requests within a 60-second window on user REST endpoints. | Check rate limit headers or count dashboard API calls. | Slow down manual refresh clicks. Tune `app.rateLimit.maxRequests` in `application.yml`. |

---

## 2. Agent Diagnostic Procedures

### Inspecting Local Agent Status
```bash
./monitor-agent status
```
Output verifies:
- `Config path`: Absolute path to `config.json`.
- `Log path`: Location of `agent.log`.
- `Server`: Target backend URL.
- `Token set`: Whether authentication credentials are present.
- `Device ID`: Assigned server UUID.
- `Service status`: Whether the OS background service is `running` or `stopped`.

### Viewing Live Agent Logs
```bash
# Linux
tail -f ~/.monitor-agent/agent.log

# Windows PowerShell
Get-Content -Wait C:\ProgramData\MonitorAgent\agent.log
```

### Manual Foreground Debugging
To bypass background service wrappers and observe direct console output:
```bash
./monitor-agent run
```

---

## 3. Backend Diagnostic Procedures

### Inspecting TimescaleDB Hypertables
```sql
SELECT hypertable_schema, hypertable_name, num_chunks 
FROM timescaledb_information.hypertables;
```

### Checking Retention Policies
```sql
SELECT hypertable_name, schedule_interval, config 
FROM timescaledb_information.jobs 
WHERE proc_name = 'policy_retention';
```

### Checking Active Device Statuses
```sql
SELECT id, hostname, ip_address, status, last_seen_at, now() - last_seen_at AS age 
FROM device 
ORDER BY last_seen_at DESC;
```\n