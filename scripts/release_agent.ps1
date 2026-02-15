Param(
  [string]$Branch = "main"
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$versionFile = Join-Path $root "VERSION"

Set-Location $root

git diff --quiet
if ($LASTEXITCODE -ne 0) {
  throw "Working tree is dirty. Commit or stash changes first."
}

& (Join-Path $root "scripts\bump_version.sh")
$version = (Get-Content $versionFile -Raw).Trim()

& (Join-Path $root "scripts\build_agent.ps1")

git add VERSION

git diff --cached --quiet
if ($LASTEXITCODE -ne 0) {
  git commit -m "chore: bump version to $version"
}

git tag "v$version"

Write-Output "Release prepared: v$version"
Write-Output "Next: git push origin $Branch; git push origin v$version"
