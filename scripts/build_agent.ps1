Param(
  [string]$Output = ""
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$versionFile = Join-Path $root "VERSION"

if (-not (Test-Path $versionFile)) {
  throw "VERSION file not found at $versionFile"
}

$version = (Get-Content $versionFile -Raw).Trim()
if (-not $version) {
  throw "VERSION is empty"
}

if (-not $Output) {
  $Output = Join-Path $root "monitor-agent\monitor-agent.exe"
}

Push-Location (Join-Path $root "monitor-agent")
try {
  go build -ldflags "-X monitor-agent/cmd.Version=$version" -o $Output .
  Write-Output "Built $Output (version $version)"
} finally {
  Pop-Location
}
