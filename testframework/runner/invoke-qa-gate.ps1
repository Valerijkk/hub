# Sole public Hub QA launcher. The private runner owns all execution behavior.

function Invoke-HubQaPublicLauncherCore {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [object]$Request,

        [Parameter(Mandatory = $true)]
        [object]$Roots,

        [Parameter(Mandatory = $true)]
        [System.Collections.Generic.IReadOnlyDictionary[string, string]]$Catalog,

        [Parameter(Mandatory = $true)]
        [scriptblock]$PrivateInvoker
    )

    foreach ($requiredRequestField in @('Target', 'Task', 'Run')) {
        if ($null -eq $Request.PSObject.Properties[$requiredRequestField] -or [string]::IsNullOrWhiteSpace([string]$Request.$requiredRequestField)) {
            return 64
        }
    }
    foreach ($requiredRootField in @('RepositoryRoot', 'RunnerRoot', 'DockerRoot', 'ReportsRoot', 'PublicScript', 'PrivateScript', 'ReportChild')) {
        if ($null -eq $Roots.PSObject.Properties[$requiredRootField] -or [string]::IsNullOrWhiteSpace([string]$Roots.$requiredRootField)) {
            return 64
        }
    }

    $target = [string]$Request.Target
    if (-not $Catalog.ContainsKey($target)) {
        return 64
    }
    $status = [string]$Catalog[$target]
    if ($status -ceq 'NOT_READY') {
        $record = [ordered]@{
            schema = 'hub.qa-gate/v1'
            target = $target
            task   = [string]$Request.Task
            run    = [string]$Request.Run
            status = 'NOT_READY'
        }
        [Console]::Out.WriteLine(($record | ConvertTo-Json -Compress))
        return 69
    }
    if ($status -cne 'READY') {
        return 64
    }

    try {
        return [int](& $PrivateInvoker -Contract 'hub.qa-runner/v1' -Target $target -Task ([string]$Request.Task) -Run ([string]$Request.Run))
    }
    catch {
        return 64
    }
}

# HUB-QA-PUBLIC-LAUNCHER-ENTRY-BLOCK
$rawArguments = @($args)
if ($rawArguments.Count -ne 6) {
    exit 64
}
if (
    $rawArguments[0] -isnot [string] -or
    $rawArguments[1] -isnot [string] -or
    $rawArguments[2] -isnot [string] -or
    $rawArguments[3] -isnot [string] -or
    $rawArguments[4] -isnot [string] -or
    $rawArguments[5] -isnot [string] -or
    $rawArguments[0] -cne '-Target' -or
    $rawArguments[2] -cne '-Task' -or
    $rawArguments[4] -cne '-Run'
) {
    exit 64
}
$target = [string]$rawArguments[1]
$task = [string]$rawArguments[3]
$run = [string]$rawArguments[5]
$closedTargets = @('backend-unit', 'frontend-unit', 'backend-openapi', 'python-contract', 'python-api', 'python-integration', 'python-e2e-smoke', 'python-e2e-full', 'python-load')
$idPattern = '^[a-z0-9][a-z0-9-]{0,23}\z'
if ($closedTargets -cnotcontains $target -or $task -cnotmatch $idPattern -or $run -cnotmatch $idPattern) {
    exit 64
}

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

$deriveRoots = {
    param(
        [string]$ScriptPath,
        [string]$TaskName,
        [string]$RunName
    )

    if ([string]::IsNullOrWhiteSpace($ScriptPath) -or [string]::IsNullOrWhiteSpace($TaskName) -or [string]::IsNullOrWhiteSpace($RunName)) {
        throw 'The public launcher root input is absent.'
    }
    if ($ScriptPath -match '^(\\\\|//)' -or $ScriptPath -match '(^|[\\/])\.\.([\\/]|$)') {
        throw 'The public launcher script path is not canonical.'
    }
    if (-not [System.IO.Path]::IsPathRooted($ScriptPath)) {
        throw 'The public launcher script path is not rooted.'
    }

    $scriptFull = [System.IO.Path]::GetFullPath($ScriptPath)
    if ($scriptFull -match '^(\\\\|//)' -or -not $scriptFull.EndsWith('\invoke-qa-gate.ps1', [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'The public launcher script path is invalid.'
    }
    $runnerDirectory = [System.IO.Directory]::GetParent($scriptFull)
    if ($null -eq $runnerDirectory -or $runnerDirectory.Name -cne 'runner') {
        throw 'The public launcher runner root is invalid.'
    }
    $testframeworkDirectory = [System.IO.Directory]::GetParent($runnerDirectory.FullName)
    if ($null -eq $testframeworkDirectory -or $testframeworkDirectory.Name -cne 'testframework') {
        throw 'The public launcher testframework root is invalid.'
    }
    $repositoryDirectory = [System.IO.Directory]::GetParent($testframeworkDirectory.FullName)
    if ($null -eq $repositoryDirectory) {
        throw 'The public launcher repository root is absent.'
    }

    $assertDirectory = {
        param([string]$Path)
        if (-not [System.IO.Directory]::Exists($Path)) {
            throw 'A required public launcher directory is absent.'
        }
        if (([System.IO.File]::GetAttributes($Path) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A public launcher directory is a reparse point.'
        }
    }.GetNewClosure()
    $assertFile = {
        param([string]$Path)
        if (-not [System.IO.File]::Exists($Path)) {
            throw 'A required public launcher file is absent.'
        }
        if (([System.IO.File]::GetAttributes($Path) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A public launcher file is a reparse point.'
        }
    }.GetNewClosure()
    $assertExistingAncestry = {
        param([string]$LeafPath)

        $componentPath = $LeafPath
        while ($true) {
            if (-not [System.IO.File]::Exists($componentPath) -and -not [System.IO.Directory]::Exists($componentPath)) {
                throw 'A public launcher ancestry component is absent.'
            }
            if (([System.IO.File]::GetAttributes($componentPath) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'A public launcher ancestry component is a reparse point.'
            }
            $parentDirectory = [System.IO.Directory]::GetParent($componentPath)
            if ($null -eq $parentDirectory) {
                break
            }
            $componentPath = $parentDirectory.FullName
        }
    }.GetNewClosure()

    $repositoryRoot = [System.IO.Path]::GetFullPath($repositoryDirectory.FullName).TrimEnd('\\')
    $testframeworkRoot = [System.IO.Path]::GetFullPath($testframeworkDirectory.FullName).TrimEnd('\\')
    $runnerRoot = [System.IO.Path]::GetFullPath($runnerDirectory.FullName).TrimEnd('\\')
    & $assertExistingAncestry $scriptFull
    & $assertDirectory $repositoryRoot
    & $assertDirectory $testframeworkRoot
    & $assertDirectory $runnerRoot
    & $assertFile $scriptFull

    $gitMarker = Join-Path $repositoryRoot '.git'
    if ([System.IO.Directory]::Exists($gitMarker)) {
        & $assertDirectory $gitMarker
    }
    elseif ([System.IO.File]::Exists($gitMarker)) {
        & $assertFile $gitMarker
    }
    else {
        throw 'The public launcher repository marker is absent.'
    }

    $privateScript = Join-Path $runnerRoot 'run-task.ps1'
    $dockerRoot = Join-Path $testframeworkRoot 'docker'
    $reportsRoot = Join-Path $testframeworkRoot 'reports'
    $reportChild = [System.IO.Path]::GetFullPath((Join-Path $reportsRoot ($TaskName + '-' + $RunName)))
    if (-not $reportChild.StartsWith(($reportsRoot + '\\'), [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'The public launcher report child escapes its root.'
    }
    & $assertFile $privateScript
    & $assertDirectory $dockerRoot
    & $assertDirectory $reportsRoot
    if ([System.IO.Directory]::Exists($reportChild)) {
        & $assertDirectory $reportChild
    }
    elseif ([System.IO.File]::Exists($reportChild)) {
        & $assertFile $reportChild
    }

    return [pscustomobject][ordered]@{
        RepositoryRoot = $repositoryRoot
        RunnerRoot     = $runnerRoot
        DockerRoot     = $dockerRoot
        ReportsRoot    = $reportsRoot
        PublicScript   = $scriptFull
        PrivateScript  = $privateScript
        ReportChild    = $reportChild
    }
}.GetNewClosure()

try {
    $roots = & $deriveRoots $PSCommandPath $task $run
}
catch {
    exit 64
}

$catalogBacking = New-Object 'System.Collections.Generic.Dictionary[string,string]' ([System.StringComparer]::Ordinal)
foreach ($catalogTarget in $closedTargets) {
    $catalogBacking.Add($catalogTarget, 'NOT_READY')
}
$catalog = New-Object 'System.Collections.ObjectModel.ReadOnlyDictionary[string,string]' ($catalogBacking)
$privateRunnerPath = [string]$roots.PrivateScript
$privateInvoker = {
    [CmdletBinding(PositionalBinding = $false)]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Contract,

        [Parameter(Mandatory = $true)]
        [string]$Target,

        [Parameter(Mandatory = $true)]
        [string]$Task,

        [Parameter(Mandatory = $true)]
        [string]$Run
    )

    if (
        $PSVersionTable.PSEdition -cne 'Desktop' -or
        $PSVersionTable.PSVersion.Major -ne 5 -or
        $PSVersionTable.PSVersion.Minor -ne 1 -or
        $Contract -cne 'hub.qa-runner/v1'
    ) {
        return 64
    }
    & $privateRunnerPath -Contract $Contract -Target $Target -Task $Task -Run $Run
    return [int]$LASTEXITCODE
}.GetNewClosure()

try {
    $publicExitCode = Invoke-HubQaPublicLauncherCore -Request ([pscustomobject][ordered]@{ Target = $target; Task = $task; Run = $run }) -Roots $roots -Catalog $catalog -PrivateInvoker $privateInvoker
    exit [int]$publicExitCode
}
catch {
    exit 64
}
