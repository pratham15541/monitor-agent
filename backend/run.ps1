$ErrorActionPreference = "Stop"

Set-Location $PSScriptRoot

mvn clean package -DskipTests

docker compose down -v

docker compose up --build
