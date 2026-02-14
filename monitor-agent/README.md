# Monitor Agent CLI

Lightweight agent that registers a device, streams live metrics, collects
detailed snapshots, and executes remote commands.

## Features

- Auto registration using company API token
- Live metrics stream (CPU, memory, disk, network) over STOMP/WebSocket
- Detailed snapshots: processes, connections, memory, services, logs
- Remote commands: shell, service control, diagnostics
- Runs as a background service (Windows/Linux/macOS via kardianos/service)

## Requirements

- Go 1.22+
- Backend URL and API token

## Build from source

```bash
go build -o monitor-agent ./
```

## Install as a Service

```bash
./monitor-agent install --token YOUR_TOKEN --server http://127.0.0.1:8080
```

This stores config, installs the service, and starts it.

## Run in Foreground

```bash
./monitor-agent run
```

## Config

Default paths:

- Windows: C:\ProgramData\MonitorAgent\config.json
- Linux/macOS: ~/.monitor-agent.json

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

## Commands

- monitor-agent install --token <TOKEN> [--server <URL>]
- monitor-agent set-token <TOKEN>
- monitor-agent set-url <URL>
- monitor-agent run
- monitor-agent deregister
- monitor-agent uninstall
- monitor-agent uninstall --service

## Networking

- POST /agent/register
- WebSocket /ws (x-agent-token header)
- STOMP send to /app/agent/metrics
- POST /agent/metrics-detail
- STOMP subscribe /topic/agent/{deviceId} and publish /app/command-result

## Notes

- install saves token/server, installs the service, and starts it.
- set-token and set-url reset deviceId so it re-registers on next run.
- deregister clears deviceId locally (no backend call).
- uninstall deletes local config; use --service to stop/remove the service.
