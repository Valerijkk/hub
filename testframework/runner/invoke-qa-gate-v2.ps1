Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

# The public gate owns raw argument parsing so PowerShell's parameter binder
# cannot emit a host-dependent failure before the fixed exit-64 boundary.
$rawInvocationArguments = @($args)
if ($rawInvocationArguments.Count -ne 6 -or
    [string]$rawInvocationArguments[0] -cne '-Target' -or
    [string]$rawInvocationArguments[2] -cne '-Task' -or
    [string]$rawInvocationArguments[4] -cne '-Run' -or
    (@($rawInvocationArguments | Where-Object { $null -eq $_ -or ([string]$_).IndexOf([char]0) -ge 0 }).Count -ne 0) -or
    [string]::IsNullOrEmpty([string]$rawInvocationArguments[1]) -or
    [string]::IsNullOrEmpty([string]$rawInvocationArguments[3]) -or
    [string]::IsNullOrEmpty([string]$rawInvocationArguments[5])) {
    [Console]::Error.WriteLine('E_B0_DIRECT_INVOCATION_FORBIDDEN')
    exit 64
}
$Target = [string]$rawInvocationArguments[1]
$Task = [string]$rawInvocationArguments[3]
$Run = [string]$rawInvocationArguments[5]

try {
    . (Join-Path $PSScriptRoot 'run-task-v2.ps1')
    $result = Invoke-HubQaRunnerV2PublicRequest -Target $Target -Task $Task -Run $Run
    $result | ConvertTo-Json -Depth 8 -Compress
    exit 0
}
catch {
    [Console]::Error.WriteLine([string]$_.Exception.Message)
    exit 64
}
