# Monitor Agent CLI

Lightweight agent that registers a device, streams live metrics, collects detailed snapshots, and executes remote commands. It runs as a background service on Windows, Linux, and macOS.

## Features (current)

- Auto registration using company API token.
- Live metrics stream over STOMP/WebSocket with adaptive sampling.
- Batched metric payloads (size 10, max wait 5 seconds).
- Detailed snapshots: processes, connections, memory, services, logs.
- Remote commands: shell, service control, diagnostics, collect-details.
- Command results published back to the control plane.
- Background service management via `kardianos/service`.

## Architecture

```mermaid
flowchart LR
  Agent[Monitor Agent] -->|POST /agent/register| API[Backend REST]
  Agent -->|POST /agent/metrics-detail/batch| API
  Agent <--> |STOMP /ws| WS[Backend WS]
  WS -->|/topic/agent/{deviceId}| Agent
  Agent -->|/app/command-result| WS
```

## Requirements

- Go 1.22+
- Backend URL and API token

## Build from source

Build with version injection (recommended):

```bash
# Linux/macOS
./scripts/build_agent.sh

# Windows
.\scripts/build_agent.ps1
```

Quick build (version will be "dev"):

```bash
go build -o monitor-agent ./
```

## Install as a service

```bash
./monitor-agent install --token YOUR_TOKEN --server http://127.0.0.1:8080
./monitor-agent start
```

## Run in foreground

```bash
./monitor-agent run
```

## Configuration

Default config paths:

- Windows: `C:\ProgramData\MonitorAgent\config.json`
- Linux/macOS: `~/.monitor-agent.json`

Override with:

```env
MONITOR_AGENT_CONFIG=/custom/path/config.json
```

Example config:

```json
{
  "serverUrl": "http://127.0.0.1:8080",
  "token": "YOUR_TOKEN",
  "deviceId": "..."
}
```

## CLI commands

- `monitor-agent install --token <TOKEN> [--server <URL>]`
- `monitor-agent set-token <TOKEN>`
- `monitor-agent set-url <URL>`
- `monitor-agent run`
- `monitor-agent start`
- `monitor-agent stop`
- `monitor-agent status`
- `monitor-agent version`
- `monitor-agent deregister`
- `monitor-agent uninstall [--service]`

Behavior notes:

- `install` saves token/server and installs the service.
- `set-token` and `set-url` reset `deviceId` so the next run re-registers.
- `deregister` clears the stored `deviceId` locally only.
- `uninstall --service` removes the service before deleting config.

## Networking and protocol

REST:

- POST `/agent/register`
- POST `/agent/metrics-detail/batch`

WebSocket (STOMP) `/ws`:

- CONNECT uses header `x-agent-token`.
- SEND `/app/agent/metrics-batch` with metric arrays.
- SUBSCRIBE `/topic/agent/{deviceId}` for commands.
- SEND `/app/command-result` for command results.

## Runtime loops

### Registration loop

- If `deviceId` is missing, the agent attempts `/agent/register` until it succeeds.
- The loop retries every 10 seconds on failure, and checks every 15 seconds.

### Metrics loop (WebSocket)

- Collects CPU, memory, disk, and network stats.
- Batch size 10 or 5 second max wait.
- Adaptive interval based on CPU:
  - > 90%: 1s
  - > 70%: 2s
  - else: 5s

### Detailed metrics loop (REST)

- Collects snapshots every 30 seconds.
- Sends to `/agent/metrics-detail/batch` with `x-agent-token` header.

### Command loop (WebSocket)

- Subscribes to `/topic/agent/{deviceId}`.
- Executes commands and returns results via `/app/command-result`.

## Detailed snapshot payload

Fields included in `details`:

- `processes`: name, pid, cpu/memory stats, io, status, threads.
- `connections`: network connection list (limited to 200).
- `memory`: OS memory and swap details.
- `services`: OS-specific service list output.
- `logs`: agent log tail + system logs (bounded to 16 KB).
- `os`: runtime OS string.
- `collectedAt`: RFC3339 UTC timestamp.

## Command types

| Type              | Payload            | Behavior                                          |
| ----------------- | ------------------ | ------------------------------------------------- |
| `shell`           | shell command      | Runs command (PowerShell on Windows, sh on Unix). |
| `service`         | start/stop/restart | Controls the agent service.                       |
| `diagnostics`     | ignored            | Returns JSON with runtime + latest metrics.       |
| `collect-details` | ignored            | Triggers a detailed snapshot batch send.          |

Safety checks:

- Destructive shell commands are blocked by a regex (rm, del, format, etc.).
- Shell command timeout is 30 seconds.
- Output is chunked at 12 KB and streamed as multiple results.

## Logs and services

- Agent log path is next to config (`agent.log`).
- System logs are collected using OS-native tools:
  - Windows: `wevtutil`
  - Linux: `journalctl`
  - macOS: `log show`
- Services snapshot uses `sc`, `systemctl`, or `launchctl`.

## Release process

The project uses a unified VERSION file for version management. See [../docs/VERSION_MANAGEMENT.md](../docs/VERSION_MANAGEMENT.md) for details.

To prepare a release:

```bash
# Linux/macOS
./scripts/release_agent.sh

# Windows
.\scripts/release_agent.ps1
```

This will:

1. Bump the patch version in VERSION file
2. Build the agent with injected version
3. Commit version change
4. Create git tag

Then push:

```bash
git push origin main
git push origin <tag>
```
