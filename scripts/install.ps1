Param(
  [string]$Repo = "pratham15541/monitor-agent",
  [string]$InstallDir = "$env:ProgramData\MonitorAgent"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $InstallDir)) {
  New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$version = $release.tag_name
if (-not $version) {
  throw "Failed to resolve latest release for $Repo"
}

$asset = "monitor-agent-windows-amd64.exe"
$url = "https://github.com/$Repo/releases/download/$version/$asset"
$out = Join-Path $InstallDir "monitor-agent.exe"

Invoke-WebRequest -Uri $url -OutFile $out

Write-Output "Installed monitor-agent to $out"
