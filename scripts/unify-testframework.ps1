# TF-M09: consolidate frontend + e2e tests under root testframework/
# Go tests MUST stay in backend/testframework/ (Go internal package rule).
# Usage: powershell -ExecutionPolicy Bypass -File scripts/unify-testframework.ps1
$ErrorActionPreference = "Stop"
$Root = git rev-parse --show-toplevel
Set-Location $Root

Write-Host "Frontend: frontend/testframework -> testframework/frontend"
if (Test-Path "frontend/testframework") {
    if (Test-Path "testframework/frontend") {
        throw "testframework/frontend already exists"
    }
    git mv frontend/testframework testframework/frontend
}

Write-Host "Update vite.config.ts (manual if script already ran)"
Write-Host "Run: cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run"
Write-Host "Go tests: keep backend/testframework/ — cannot move to separate module (internal/ imports)."
Write-Host "Done."
