# External contract for testframework/runner/invoke-qa-gate.ps1.
# This selftest is authored-first and intentionally not executed on the Windows host.

$fixtureCatalogPath = Join-Path $PSScriptRoot 'fixtures.ps1'
. $fixtureCatalogPath
$publicLauncherFixtures = Get-HubQaPublicLauncherFixtureCatalog

function New-HubQaPublicLauncherTestCopy {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [ValidateSet('Directory', 'Linked')]
        [string]$GitMarker
    )

    $sourceRepositoryRoot = [System.IO.Directory]::GetParent([System.IO.Directory]::GetParent([System.IO.Directory]::GetParent($PSScriptRoot).FullName).FullName).FullName
    $sourcePublicLauncher = Join-Path $sourceRepositoryRoot 'testframework\runner\invoke-qa-gate.ps1'
    $sourcePrivateRunner = Join-Path $sourceRepositoryRoot 'testframework\runner\run-task.ps1'
    if (-not [System.IO.File]::Exists($sourcePublicLauncher) -or -not [System.IO.File]::Exists($sourcePrivateRunner)) {
        throw 'The public-launcher selftest sources are absent.'
    }

    $repositoryRoot = Join-Path $TestDrive ('hub-qa-public-' + [System.Guid]::NewGuid().ToString('N'))
    $runnerRoot = Join-Path $repositoryRoot 'testframework\runner'
    $dockerRoot = Join-Path $repositoryRoot 'testframework\docker'
    $reportsRoot = Join-Path $repositoryRoot 'testframework\reports'
    [void][System.IO.Directory]::CreateDirectory($runnerRoot)
    [void][System.IO.Directory]::CreateDirectory($dockerRoot)
    [void][System.IO.Directory]::CreateDirectory($reportsRoot)

    $gitMarkerPath = Join-Path $repositoryRoot '.git'
    if ($GitMarker -ceq 'Directory') {
        [void][System.IO.Directory]::CreateDirectory($gitMarkerPath)
    }
    else {
        [System.IO.File]::WriteAllText($gitMarkerPath, [string]$publicLauncherFixtures.GitMarkers.Linked.Contents, [System.Text.UTF8Encoding]::new($false))
    }

    $publicLauncherPath = Join-Path $runnerRoot 'invoke-qa-gate.ps1'
    $privateRunnerPath = Join-Path $runnerRoot 'run-task.ps1'
    [System.IO.File]::Copy($sourcePublicLauncher, $publicLauncherPath)
    [System.IO.File]::Copy($sourcePrivateRunner, $privateRunnerPath)

    $sourceText = [System.IO.File]::ReadAllText($publicLauncherPath)
    $entryMarker = '# HUB-QA-PUBLIC-LAUNCHER-ENTRY-BLOCK'
    $markerOffset = $sourceText.IndexOf($entryMarker, [System.StringComparison]::Ordinal)
    if ($markerOffset -lt 0) {
        throw 'The public launcher has no explicit entry marker.'
    }
    $entryOffset = $markerOffset + $entryMarker.Length
    $testableText = $sourceText.Substring(0, $entryOffset) + [Environment]::NewLine + 'return' + [Environment]::NewLine
    $testablePath = Join-Path $runnerRoot 'invoke-qa-gate.testable.ps1'
    [System.IO.File]::WriteAllText($testablePath, $testableText, [System.Text.UTF8Encoding]::new($false))

    return [pscustomobject][ordered]@{
        RepositoryRoot      = $repositoryRoot
        RunnerRoot          = $runnerRoot
        DockerRoot          = $dockerRoot
        ReportsRoot         = $reportsRoot
        PublicLauncherPath  = $publicLauncherPath
        PrivateRunnerPath   = $privateRunnerPath
        TestableLauncherPath = $testablePath
    }
}

function Invoke-HubQaPublicLauncherProcess {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$ScriptPath,

        [Parameter(Mandatory)]
        [string[]]$Arguments,

        [string]$WorkingDirectory
    )

    $hostPath = Join-Path $PSHOME 'powershell.exe'
    if (-not [System.IO.File]::Exists($hostPath)) {
        throw 'The Windows PowerShell 5.1 host is required for this selftest.'
    }
    $escapedArguments = @('-NoProfile', '-NonInteractive', '-ExecutionPolicy', 'Bypass', '-File', $ScriptPath) + @($Arguments)
    $commandLine = (@($escapedArguments | ForEach-Object { '"' + ([string]$_).Replace('"', '\"') + '"' }) -join ' ')
    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $hostPath
    $startInfo.Arguments = $commandLine
    if (-not [string]::IsNullOrWhiteSpace($WorkingDirectory)) {
        $startInfo.WorkingDirectory = $WorkingDirectory
    }
    $startInfo.UseShellExecute = $false
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo
    [void]$process.Start()
    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()
    $process.WaitForExit()
    return [pscustomobject][ordered]@{
        ExitCode = [int]$process.ExitCode
        Stdout   = $stdout
        Stderr   = $stderr
    }
}

function New-HubQaReadOnlyTargetCatalog {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Target,

        [Parameter(Mandatory)]
        [string]$Status
    )

    $backing = New-Object 'System.Collections.Generic.Dictionary[string,string]' ([System.StringComparer]::Ordinal)
    $backing.Add($Target, $Status)
    return New-Object 'System.Collections.ObjectModel.ReadOnlyDictionary[string,string]' ($backing)
}

function Get-HubQaFixtureTreeFingerprint {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$RootPath
    )

    return @(
        Get-ChildItem -LiteralPath $RootPath -Force -Recurse |
            ForEach-Object { $_.FullName.Substring($RootPath.Length) + '|' + [int][System.IO.File]::GetAttributes($_.FullName) } |
            Sort-Object
    ) -join "`n"
}

function New-HubQaFixtureDirectoryReparse {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$LinkPath,

        [Parameter(Mandatory)]
        [string]$TargetPath
    )

    [void](New-Item -ItemType Junction -Path $LinkPath -Target $TargetPath -ErrorAction Stop)
    if (([System.IO.File]::GetAttributes($LinkPath) -band [System.IO.FileAttributes]::ReparsePoint) -eq 0) {
        throw 'The required isolated-copy directory reparse fixture was not created.'
    }
}

function New-HubQaFixtureFileReparse {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$LinkPath,

        [Parameter(Mandatory)]
        [string]$TargetPath
    )

    [void](New-Item -ItemType SymbolicLink -Path $LinkPath -Target $TargetPath -ErrorAction Stop)
    if (([System.IO.File]::GetAttributes($LinkPath) -band [System.IO.FileAttributes]::ReparsePoint) -eq 0) {
        throw 'The required isolated-copy file reparse fixture was not created.'
    }
}

Describe 'Hub QA public launcher contract' {
    It 'rejects every raw CLI closure violation before private work or report mutation' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        foreach ($case in $publicLauncherFixtures.InvalidArgumentVectors.GetEnumerator()) {
            $result = Invoke-HubQaPublicLauncherProcess -ScriptPath $fixture.PublicLauncherPath -Arguments @([string[]]$case.Value)
            $result.ExitCode | Should -Be 64
            $result.Stdout | Should -Be ''
            $result.Stderr | Should -Be ''
            @(Get-ChildItem -LiteralPath $fixture.ReportsRoot -Force).Count | Should -Be 0
        }
    }

    It 'returns one content-free NOT_READY record for every closed current or future target' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        foreach ($target in $publicLauncherFixtures.ValidTargets) {
            $result = Invoke-HubQaPublicLauncherProcess -ScriptPath $fixture.PublicLauncherPath -Arguments @('-Target', $target, '-Task', 'task-2a', '-Run', 'run-2a')
            $result.ExitCode | Should -Be 69
            $result.Stderr | Should -Be ''
            $lines = @($result.Stdout -split "`r?`n" | Where-Object { $_.Length -gt 0 })
            $lines.Count | Should -Be 1
            $record = $lines[0] | ConvertFrom-Json
            @($record.PSObject.Properties.Name) | Should -Be @('schema', 'target', 'task', 'run', 'status')
            $record.schema | Should -Be 'hub.qa-gate/v1'
            $record.target | Should -Be $target
            $record.task | Should -Be 'task-2a'
            $record.run | Should -Be 'run-2a'
            $record.status | Should -Be 'NOT_READY'
            @(Get-ChildItem -LiteralPath $fixture.ReportsRoot -Force).Count | Should -Be 0
        }
    }

    It 'accepts both regular and linked-worktree Git root markers without dereferencing gitdir' {
        foreach ($marker in @('Directory', 'Linked')) {
            $fixture = New-HubQaPublicLauncherTestCopy -GitMarker $marker
            $result = Invoke-HubQaPublicLauncherProcess -ScriptPath $fixture.PublicLauncherPath -Arguments @($publicLauncherFixtures.ValidRequest)
            $result.ExitCode | Should -Be 69
            $result.Stdout | Should -Match '"status":"NOT_READY"'
            @(Get-ChildItem -LiteralPath $fixture.ReportsRoot -Force).Count | Should -Be 0
        }
    }

    It 'exposes only the marked core seam and delegates a ready request with the exact fixed private arguments' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        . $fixture.TestableLauncherPath
        $calls = New-Object 'System.Collections.Generic.List[object]'
        $privateInvoker = {
            param($Contract, $Target, $Task, $Run)
            $calls.Add([pscustomobject][ordered]@{ Contract = $Contract; Target = $Target; Task = $Task; Run = $Run })
            return 23
        }.GetNewClosure()
        $request = [pscustomobject][ordered]@{ Target = 'backend-unit'; Task = 'task-2a'; Run = 'run-2a' }
        $roots = [pscustomobject][ordered]@{ RepositoryRoot = $fixture.RepositoryRoot; RunnerRoot = $fixture.RunnerRoot; DockerRoot = $fixture.DockerRoot; ReportsRoot = $fixture.ReportsRoot; PublicScript = $fixture.PublicLauncherPath; PrivateScript = $fixture.PrivateRunnerPath; ReportChild = (Join-Path $fixture.ReportsRoot 'task-2a-run-2a') }
        $exitCode = Invoke-HubQaPublicLauncherCore -Request $request -Roots $roots -Catalog (New-HubQaReadOnlyTargetCatalog -Target 'backend-unit' -Status 'READY') -PrivateInvoker $privateInvoker
        $exitCode | Should -Be 23
        $calls.Count | Should -Be 1
        $calls[0].Contract | Should -Be 'hub.qa-runner/v1'
        $calls[0].Target | Should -Be 'backend-unit'
        $calls[0].Task | Should -Be 'task-2a'
        $calls[0].Run | Should -Be 'run-2a'
    }

    It 'preserves every private exit result including the private report-collision rejection' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        . $fixture.TestableLauncherPath
        $request = [pscustomobject][ordered]@{ Target = 'backend-unit'; Task = 'task-2a'; Run = 'run-2a' }
        $roots = [pscustomobject][ordered]@{ RepositoryRoot = $fixture.RepositoryRoot; RunnerRoot = $fixture.RunnerRoot; DockerRoot = $fixture.DockerRoot; ReportsRoot = $fixture.ReportsRoot; PublicScript = $fixture.PublicLauncherPath; PrivateScript = $fixture.PrivateRunnerPath; ReportChild = (Join-Path $fixture.ReportsRoot 'task-2a-run-2a') }
        foreach ($expectedExit in @($publicLauncherFixtures.PrivateExitCodes + 64)) {
            $privateInvoker = { param($Contract, $Target, $Task, $Run) return $expectedExit }.GetNewClosure()
            $actualExit = Invoke-HubQaPublicLauncherCore -Request $request -Roots $roots -Catalog (New-HubQaReadOnlyTargetCatalog -Target 'backend-unit' -Status 'READY') -PrivateInvoker $privateInvoker
            $actualExit | Should -Be $expectedExit
        }
    }

    It 'rejects every isolated-copy reparse location before work and preserves its report tree' {
        foreach ($case in $publicLauncherFixtures.ReparseCases.GetEnumerator()) {
            $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
            switch ([string]$case.Value.Location) {
                'script' {
                    $physicalPath = $fixture.PublicLauncherPath + '.physical'
                    [System.IO.File]::Move($fixture.PublicLauncherPath, $physicalPath)
                    New-HubQaFixtureFileReparse -LinkPath $fixture.PublicLauncherPath -TargetPath $physicalPath
                }
                'root' {
                    $physicalPath = $fixture.RepositoryRoot + '.physical'
                    [System.IO.Directory]::Move($fixture.RepositoryRoot, $physicalPath)
                    New-HubQaFixtureDirectoryReparse -LinkPath $fixture.RepositoryRoot -TargetPath $physicalPath
                }
                'runner' {
                    $physicalPath = $fixture.RunnerRoot + '.physical'
                    [System.IO.Directory]::Move($fixture.RunnerRoot, $physicalPath)
                    New-HubQaFixtureDirectoryReparse -LinkPath $fixture.RunnerRoot -TargetPath $physicalPath
                }
                'docker' {
                    $physicalPath = $fixture.DockerRoot + '.physical'
                    [System.IO.Directory]::Move($fixture.DockerRoot, $physicalPath)
                    New-HubQaFixtureDirectoryReparse -LinkPath $fixture.DockerRoot -TargetPath $physicalPath
                }
                'reports' {
                    $physicalPath = $fixture.ReportsRoot + '.physical'
                    [System.IO.Directory]::Move($fixture.ReportsRoot, $physicalPath)
                    New-HubQaFixtureDirectoryReparse -LinkPath $fixture.ReportsRoot -TargetPath $physicalPath
                }
                'report-child' {
                    $physicalPath = Join-Path $fixture.RepositoryRoot 'outside-report-child'
                    [void][System.IO.Directory]::CreateDirectory($physicalPath)
                    New-HubQaFixtureDirectoryReparse -LinkPath (Join-Path $fixture.ReportsRoot 'task-2a-run-2a') -TargetPath $physicalPath
                }
                default {
                    throw 'The reparse fixture location is not closed.'
                }
            }

            $reportsBefore = Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot
            $publicResult = Invoke-HubQaPublicLauncherProcess -ScriptPath $fixture.PublicLauncherPath -Arguments @($publicLauncherFixtures.ValidRequest)
            $publicResult.ExitCode | Should -Be 64
            (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore
            if ([string]$case.Value.Invocation -ceq 'public-and-private-boundary') {
                $privateResult = Invoke-HubQaPublicLauncherProcess -ScriptPath $fixture.PrivateRunnerPath -Arguments @('-Contract', 'hub.qa-runner/v1', '-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a')
                $privateResult.ExitCode | Should -Be 64
                (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore
            }
        }
    }

    It 'rejects a public copy launched through a reparse ancestor above the repository root' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        $ancestorPhysical = Join-Path $TestDrive ('ancestor-physical-' + [System.Guid]::NewGuid().ToString('N'))
        [void][System.IO.Directory]::CreateDirectory($ancestorPhysical)
        $physicalRepository = Join-Path $ancestorPhysical 'repo'
        [System.IO.Directory]::Move($fixture.RepositoryRoot, $physicalRepository)
        $ancestorLink = Join-Path $TestDrive ('ancestor-link-' + [System.Guid]::NewGuid().ToString('N'))
        New-HubQaFixtureDirectoryReparse -LinkPath $ancestorLink -TargetPath $ancestorPhysical
        $publicPath = Join-Path $ancestorLink 'repo\testframework\runner\invoke-qa-gate.ps1'
        $reportsRoot = Join-Path $ancestorLink 'repo\testframework\reports'
        $reportsBefore = Get-HubQaFixtureTreeFingerprint -RootPath $reportsRoot
        $result = Invoke-HubQaPublicLauncherProcess -ScriptPath $publicPath -Arguments @($publicLauncherFixtures.ValidRequest)
        $result.ExitCode | Should -Be $publicLauncherFixtures.AncestorReparse.ExpectedExitCode
        (Get-HubQaFixtureTreeFingerprint -RootPath $reportsRoot) | Should -Be $reportsBefore
    }

    It 'rejects dot-dot and root-escape isolated-copy script launches before report mutation' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        $dotDotPath = Join-Path $fixture.RunnerRoot '..\runner\invoke-qa-gate.ps1'
        $reportsBefore = Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot
        $dotDotResult = Invoke-HubQaPublicLauncherProcess -ScriptPath $dotDotPath -Arguments @($publicLauncherFixtures.ValidRequest)
        $dotDotResult.ExitCode | Should -Be $publicLauncherFixtures.ScriptLaunchCases.DotDot.ExpectedExitCode
        (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore

        $outsideRoot = Join-Path $fixture.RepositoryRoot '..\outside-launch-root'
        $outsideRunner = Join-Path $outsideRoot 'runner'
        [void][System.IO.Directory]::CreateDirectory($outsideRunner)
        [System.IO.File]::Copy($fixture.PublicLauncherPath, (Join-Path $outsideRunner 'invoke-qa-gate.ps1'))
        $outsideResult = Invoke-HubQaPublicLauncherProcess -ScriptPath (Join-Path $outsideRunner 'invoke-qa-gate.ps1') -Arguments @($publicLauncherFixtures.ValidRequest)
        $outsideResult.ExitCode | Should -Be $publicLauncherFixtures.ScriptLaunchCases.RootEscape.ExpectedExitCode
        (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore
    }

    It 'rejects UNC-form and drive-relative isolated-copy script launches before report mutation' {
        $fixture = New-HubQaPublicLauncherTestCopy -GitMarker Directory
        $reportsBefore = Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot
        $verbatimUncForm = '\\?\' + $fixture.PublicLauncherPath
        $uncResult = Invoke-HubQaPublicLauncherProcess -ScriptPath $verbatimUncForm -Arguments @($publicLauncherFixtures.ValidRequest)
        $uncResult.ExitCode | Should -Be $publicLauncherFixtures.ScriptLaunchCases.Unc.ExpectedExitCode
        (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore

        $volumeRoot = [System.IO.Path]::GetPathRoot($fixture.PublicLauncherPath)
        $driveRelativeForm = $volumeRoot.Substring(0, 2) + $fixture.PublicLauncherPath.Substring($volumeRoot.Length)
        $driveRelativeResult = Invoke-HubQaPublicLauncherProcess -ScriptPath $driveRelativeForm -Arguments @($publicLauncherFixtures.ValidRequest) -WorkingDirectory $volumeRoot
        $driveRelativeResult.ExitCode | Should -Be $publicLauncherFixtures.ScriptLaunchCases.DriveRelative.ExpectedExitCode
        (Get-HubQaFixtureTreeFingerprint -RootPath $fixture.ReportsRoot) | Should -Be $reportsBefore
    }

    It 'keeps public operational code closed and pins both manifest source commits to the immutable base' {
        $sourceRepositoryRoot = [System.IO.Directory]::GetParent([System.IO.Directory]::GetParent([System.IO.Directory]::GetParent($PSScriptRoot).FullName).FullName).FullName
        $publicSource = [System.IO.File]::ReadAllText((Join-Path $sourceRepositoryRoot 'testframework\runner\invoke-qa-gate.ps1'))
        $privateSource = [System.IO.File]::ReadAllText((Join-Path $sourceRepositoryRoot 'testframework\runner\run-task.ps1'))
        $publicSource | Should -Match ([regex]::Escape('$PSCommandPath'))
        $privateSource | Should -Match ([regex]::Escape('$PSCommandPath'))
        $publicSource | Should -Not -Match ([regex]::Escape('$PSScriptRoot'))
        $privateSource | Should -Not -Match ([regex]::Escape('$PSScriptRoot'))
        $publicSource | Should -Match ([regex]::Escape("`$privateScript = Join-Path `$runnerRoot 'run-task.ps1'"))
        $publicSource | Should -Match '\\z'
        $privateSource | Should -Match '\\z'
        foreach ($forbiddenOperationalToken in @('Start-Process', 'System.Diagnostics.Process', 'docker.exe', 'docker compose', 'Remove-Item', 'CreateDirectory', 'FileStream', 'powershell.exe', '$env:', '$env:PATH', '$env:PATHEXT', 'GetEnvironmentVariable', 'PATHEXT', 'Get-Command', 'Invoke-Expression')) {
            $publicSource | Should -Not -Match ([regex]::Escape($forbiddenOperationalToken))
        }
        foreach ($manifestPath in @('testframework\docker\build-manifest.v2.json', 'testframework\docker\source-policy.v1.json')) {
            $manifest = [System.IO.File]::ReadAllText((Join-Path $sourceRepositoryRoot $manifestPath)) | ConvertFrom-Json
            $manifest.source_commit | Should -Be 'cc0f3fcbf4e9a0adc70e946e739b81d21ac7da6c'
            $manifest.PSObject.Properties.Name | Should -Not -Contain 'future_commit'
            $manifest.PSObject.Properties.Name | Should -Not -Contain 'self_commit'
        }
    }
}
