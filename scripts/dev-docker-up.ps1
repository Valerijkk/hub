# Hub — локальный dev-стек в Docker (Windows / PowerShell)
# Usage: pwsh scripts/dev-docker-up.ps1 [-Build] [-Down] [-Seed]
param(
    [switch]$Build,
    [switch]$Down,
    [switch]$Seed
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

function Ensure-File($Example, $Target) {
    if (-not (Test-Path $Target)) {
        if (-not (Test-Path $Example)) {
            throw "Missing template: $Example"
        }
        Copy-Item $Example $Target
        Write-Host "Created $Target from $Example"
    }
}

Ensure-File ".env.local.example" ".env.local"
Ensure-File "docker-compose.local.yml.example" "docker-compose.local.yml"

$Compose = @(
    "docker", "compose",
    "--env-file", ".env.local",
    "-f", "docker-compose.local.yml"
)

if ($Down) {
    & @Compose @("down", "-v")
    exit $LASTEXITCODE
}

$Args = @("up", "-d")
if ($Build) { $Args += "--build" }

Write-Host "Starting Hub local stack..."
& @Compose @Args
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Waiting for backend health..."
for ($i = 0; $i -lt 60; $i++) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:3000/api/health" -UseBasicParsing -TimeoutSec 3
        if ($r.StatusCode -eq 200) { break }
    } catch { }
    Start-Sleep -Seconds 2
}

if ($Seed) {
    Write-Host "Seeding demo data..."
    Invoke-RestMethod -Method Post -Uri "http://localhost:3000/api/dev/seed" | Out-Null
}

Write-Host ""
Write-Host "Hub local dev ready:"
Write-Host "  UI:   http://localhost:8080"
Write-Host "  API:  http://localhost:3000"
Write-Host "  Seed: curl -X POST http://localhost:3000/api/dev/seed"
Write-Host "  Logs: docker compose --env-file .env.local -f docker-compose.local.yml logs -f"
