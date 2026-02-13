param(
    [Parameter(Mandatory = $true)]
    [string]$Token,

    [string]$Server = "",
    [string]$RepoOwner = "YOUR_GITHUB_ORG",
    [string]$RepoName = "YOUR_REPO_NAME",
    [string]$AssetName = "monitor-agent-windows.exe",
    [string]$InstallDir = "$env:ProgramFiles\MonitorAgent",
    [string]$ExeName = "monitor-agent.exe"
)

$ErrorActionPreference = "Stop"

function Test-Admin {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-Admin)) {
    Write-Warning "This installer needs Administrator privileges to install the service. Re-run in an elevated PowerShell."
    exit 1
}

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$releaseUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
$release = Invoke-RestMethod -Uri $releaseUrl
$asset = $release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1

if (-not $asset) {
    Write-Error "Asset not found: $AssetName in latest release."
    exit 1
}

$downloadPath = Join-Path $InstallDir $ExeName
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $downloadPath

Write-Host "Downloaded $AssetName to $downloadPath"

$installArgs = @("install", "--token", $Token)
if ($Server -ne "") {
    $installArgs += @("--server", $Server)
}

& $downloadPath @installArgs

Write-Host "Install complete. Service should be running in the background."
