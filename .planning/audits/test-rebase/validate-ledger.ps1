[CmdletBinding()]
param(
    [Parameter()]
    [string]$Python
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$inventory = Join-Path $PSScriptRoot 'inventory.py'
if (-not (Test-Path -LiteralPath $inventory -PathType Leaf)) {
    throw "Inventory tool not found: $inventory"
}

$pythonExecutable = $null
$pythonPrefix = @()

if ($Python) {
    if (Test-Path -LiteralPath $Python -PathType Leaf) {
        $pythonExecutable = (Resolve-Path -LiteralPath $Python).Path
    } else {
        $command = Get-Command $Python -ErrorAction SilentlyContinue
        if ($command) {
            $pythonExecutable = $command.Source
        }
    }
}

if (-not $pythonExecutable) {
    $profiles = @()
    if ($env:USERPROFILE) {
        $profiles += $env:USERPROFILE
    }
    $platformProfile = [Environment]::GetFolderPath('UserProfile')
    if ($platformProfile) {
        $profiles += $platformProfile
    }
    $bundledCandidates = foreach ($userProfile in ($profiles | Select-Object -Unique)) {
        Join-Path $userProfile '.cache/codex-runtimes/codex-primary-runtime/dependencies/python/python.exe'
        Join-Path $userProfile '.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python3'
        Join-Path $userProfile '.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python'
    }
    foreach ($candidate in $bundledCandidates) {
        if (Test-Path -LiteralPath $candidate -PathType Leaf) {
            $pythonExecutable = $candidate
            break
        }
    }
}

if (-not $pythonExecutable) {
    foreach ($name in @('python3', 'python')) {
        $command = Get-Command $name -ErrorAction SilentlyContinue
        if ($command) {
            $pythonExecutable = $command.Source
            break
        }
    }
}

if (-not $pythonExecutable) {
    $launcher = Get-Command 'py' -ErrorAction SilentlyContinue
    if ($launcher) {
        $pythonExecutable = $launcher.Source
        $pythonPrefix = @('-3')
    }
}

if (-not $pythonExecutable) {
    throw 'Python 3 not found. Pass -Python, set PATH, or install the bundled Codex runtime.'
}

Write-Host "Ledger validator: $pythonExecutable"
& $pythonExecutable @pythonPrefix $inventory --check
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

& $pythonExecutable @pythonPrefix $inventory --self-test
exit $LASTEXITCODE
