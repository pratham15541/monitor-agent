# Monitor Agent CLI

Lightweight agent that registers a device and streams metrics to the backend.

## Requirements

- Go 1.22+
- Backend URL and API token

## Quick install (Windows, GitHub Releases)

1. Download and run the installer script (PowerShell, admin required):

```powershell
powershell -ExecutionPolicy Bypass -File scripts\install.ps1 -Token YOUR_TOKEN -Server http://127.0.0.1:8080 -RepoOwner YOUR_GITHUB_ORG -RepoName YOUR_REPO_NAME
```

2. Or download directly and run manually:

```bash
curl -L -o monitor-agent.exe https://github.com/YOUR_GITHUB_ORG/YOUR_REPO_NAME/releases/latest/download/monitor-agent-windows.exe
monitor-agent.exe install --token YOUR_TOKEN --server http://127.0.0.1:8080
```

## Build from source

```bash
go build -o monitor-agent ./
```

## Run

```bash
./monitor-agent run
```

## Config

The agent stores settings in `~/.monitor-agent.json`:

```json
{
  "serverUrl": "http://127.0.0.1:8080",
  "token": "YOUR_TOKEN",
  "deviceId": "..."
}
```

## Commands

- `monitor-agent install --token <TOKEN> [--server <URL>]`
- `monitor-agent set-token <TOKEN>`
- `monitor-agent run`
- `monitor-agent deregister`
- `monitor-agent uninstall`
- `monitor-agent uninstall --service`

## Notes

- `install` saves token/server, installs the service, and starts it in the background.
- `set-token` updates the token and clears `deviceId` so it re-registers on next run.
- `deregister` clears `deviceId` locally (no backend call).
- `uninstall` deletes the local config; use `--service` to also stop/remove the service.
