Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

# This file is a private library.  There is deliberately no executable
# parameter set: a public invocation has no authority to run a target or to
# transport bootstrap state.
$rawInvocationArguments = @($MyInvocation.UnboundArguments)
if ($MyInvocation.InvocationName -ne '.' -or $rawInvocationArguments.Count -ne 0) {
    [Console]::Error.WriteLine('E_B0_DIRECT_INVOCATION_FORBIDDEN')
    exit 64
}

$script:HubQaRunnerV2Contract = 'hub.qa-runner/v2'
$script:HubQaRunnerV2TaskPattern = '^[a-z][a-z0-9-]{2,31}$'
$script:HubQaRunnerV2RunPattern = '^[a-z0-9][a-z0-9-]{7,63}$'
$script:HubQaRunnerV2FixedTargets = @('backend-unit', 'frontend-unit', 'backend-openapi', 'g0-planning-selftest')
$script:HubQaRunnerV2PythonTargets = @('python-contract', 'python-api', 'python-integration', 'python-e2e-smoke', 'python-e2e-full', 'python-load')

function Get-HubQaRunnerV2Sha256 {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)

    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        return 'sha256:' + ([System.BitConverter]::ToString($algorithm.ComputeHash($Bytes)).Replace('-', '').ToLowerInvariant())
    }
    finally {
        $algorithm.Dispose()
    }
}

function Write-HubQaRunnerV2CreateNew {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][byte[]]$Bytes
    )

    $stream = [System.IO.File]::Open($Path, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
    try {
        $stream.Write($Bytes, 0, $Bytes.Length)
        $stream.Flush($true)
    }
    finally {
        $stream.Dispose()
    }
}

function Test-HubQaRunnerV2ExactContext {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $expected = @(
        'Task', 'Run', 'RoleClaim', 'OwnerNonce', 'OwnerMarkerRawBytes',
        'OwnerMarkerSha256', 'RunnerInspectionRawBytes', 'RunnerInspectionSha256',
        'LocalBaseReceiptRawBytes', 'LocalBaseReceiptSha256', 'LocalBaseReceiptNonce'
    )
    if ($BootstrapPrivateContext.Count -ne $expected.Count) {
        throw 'E_B0_ROLE_TRANSPORT'
    }
    for ($index = 0; $index -lt $expected.Count; $index++) {
        if (-not $BootstrapPrivateContext.ContainsKey($expected[$index])) {
            throw 'E_B0_ROLE_TRANSPORT'
        }
    }
    if (
        [string]$BootstrapPrivateContext.Task -cne 'g0-assets-runner' -or
        [string]$BootstrapPrivateContext.Run -notmatch $script:HubQaRunnerV2RunPattern -or
        [string]$BootstrapPrivateContext.RoleClaim -cne 'hub.qa-bootstrap/v1' -or
        [string]$BootstrapPrivateContext.OwnerNonce -notmatch '^[0-9a-f]{32}$' -or
        $BootstrapPrivateContext.OwnerMarkerRawBytes -isnot [byte[]] -or
        $BootstrapPrivateContext.RunnerInspectionRawBytes -isnot [byte[]] -or
        $BootstrapPrivateContext.LocalBaseReceiptRawBytes -isnot [byte[]] -or
        [string]$BootstrapPrivateContext.OwnerMarkerSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$BootstrapPrivateContext.RunnerInspectionSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$BootstrapPrivateContext.LocalBaseReceiptSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$BootstrapPrivateContext.LocalBaseReceiptNonce -notmatch '^[0-9a-f]{32}$' -or
        [string]$BootstrapPrivateContext.LocalBaseReceiptNonce -cne [string]$BootstrapPrivateContext.OwnerNonce -or
        (Get-HubQaRunnerV2Sha256 -Bytes ([byte[]]$BootstrapPrivateContext.OwnerMarkerRawBytes)) -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256 -or
        (Get-HubQaRunnerV2Sha256 -Bytes ([byte[]]$BootstrapPrivateContext.RunnerInspectionRawBytes)) -cne [string]$BootstrapPrivateContext.RunnerInspectionSha256 -or
        (Get-HubQaRunnerV2Sha256 -Bytes ([byte[]]$BootstrapPrivateContext.LocalBaseReceiptRawBytes)) -cne [string]$BootstrapPrivateContext.LocalBaseReceiptSha256
    ) {
        throw 'E_B0_ROLE_TRANSPORT'
    }
}

function Invoke-HubQaRunnerV2PublicRequest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Target,
        [Parameter(Mandatory = $true)][string]$Task,
        [Parameter(Mandatory = $true)][string]$Run
    )

    if (
        $Task -notmatch $script:HubQaRunnerV2TaskPattern -or
        $Run -notmatch $script:HubQaRunnerV2RunPattern
    ) {
        throw 'E_B0_TASK_RUN'
    }
    if ($Target -ceq 'g0-planning-selftest') {
        throw 'E_B0_DIRECT_INVOCATION_FORBIDDEN'
    }
    if ($script:HubQaRunnerV2PythonTargets -ccontains $Target) {
        return [pscustomobject][ordered]@{
            contract = $script:HubQaRunnerV2Contract
            target = $Target
            task = $Task
            run = $Run
            status = 'NOT_READY'
            reason_code = 'E_PYTHON_TARGET_NOT_READY'
        }
    }
    if ($script:HubQaRunnerV2FixedTargets -cnotcontains $Target) {
        throw 'E_B0_DIRECT_INVOCATION_FORBIDDEN'
    }

    # Image-lock admission is a later C1 event.  Normal public targets are
    # deliberately not a bootstrap fallback and cannot invoke Docker before it.
    return [pscustomobject][ordered]@{
        contract = $script:HubQaRunnerV2Contract
        target = $Target
        task = $Task
        run = $Run
        status = 'NOT_READY'
        reason_code = 'E_G0_C1_LOCK_NOT_READY'
    }
}
