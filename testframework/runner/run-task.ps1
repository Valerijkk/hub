# Private implementation for hub.qa-runner/v1. The only supported caller is /qa-gate.
# Source status: NOT_READY / NOT_RUN.

[CmdletBinding()]
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

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

class HubQaSupervisorRequest {
    [string]$Contract
    [string]$Target
    [string]$Task
    [string]$Run

    HubQaSupervisorRequest(
        [string]$contract,
        [string]$target,
        [string]$task,
        [string]$run
    ) {
        $this.Contract = $contract
        $this.Target = $target
        $this.Task = $task
        $this.Run = $run
    }
}

class HubQaSupervisorDependencies {
    [scriptblock]$InspectLockedImage
    [scriptblock]$RunFixedTarget
    [scriptblock]$ListOwnedResources
    [scriptblock]$CleanupOwnedResources

    HubQaSupervisorDependencies(
        [scriptblock]$inspectLockedImage,
        [scriptblock]$runFixedTarget,
        [scriptblock]$listOwnedResources,
        [scriptblock]$cleanupOwnedResources
    ) {
        $this.InspectLockedImage = $inspectLockedImage
        $this.RunFixedTarget = $runFixedTarget
        $this.ListOwnedResources = $listOwnedResources
        $this.CleanupOwnedResources = $cleanupOwnedResources
    }
}

class HubQaRunningProcessHandle {
    [System.Diagnostics.Process]$Process
    [System.IO.Stream]$StandardOutputStream
    [System.IO.Stream]$StandardErrorStream

    HubQaRunningProcessHandle(
        [System.Diagnostics.Process]$process,
        [System.IO.Stream]$standardOutputStream,
        [System.IO.Stream]$standardErrorStream
    ) {
        $this.Process = $process
        $this.StandardOutputStream = $standardOutputStream
        $this.StandardErrorStream = $standardErrorStream
    }
}

function Invoke-HubQaSupervisor {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [HubQaSupervisorRequest]$Request
    )

    $contractLiteral = 'hub.qa-runner/v1'
    $fixedTargets = @('backend-unit', 'frontend-unit', 'backend-openapi')
    $idPattern = '^[a-z0-9][a-z0-9-]{0,23}$'
    $digestReferencePattern = '^[a-z0-9][a-z0-9.-]*(?::[0-9]+)?/[a-z0-9][a-z0-9._/-]*@sha256:[a-f0-9]{64}$'
    $lockMemberPattern = '"(?<key>[^"\\]+)"\s*:\s*"(?<value>[^"\\]+)"'
    $lockDocumentPattern = '^\s*\{\s*"[^"\\]+"\s*:\s*"[^"\\]+"\s*(?:,\s*"[^"\\]+"\s*:\s*"[^"\\]+"\s*){2}\}\s*$'
    $inspectTimeoutMilliseconds = 30000
    $outerTimeoutMilliseconds = 810000
    $cleanupTimeoutMilliseconds = 30000
    $commandDeadlineSeconds = 180
    $targetDeadlineSeconds = 780
    $allowedOpenApiPaths = @(
        'backend/docs/docs.go'
        'backend/docs/swagger.json'
        'backend/docs/swagger.yaml'
    )
    $expectedCommands = @{
        'backend-unit' = @('backend-build', 'backend-vet', 'backend-unit', 'backend-race')
        'frontend-unit' = @('frontend-typecheck', 'frontend-biome', 'frontend-build', 'frontend-unit')
        'backend-openapi' = @('openapi-generate-1', 'openapi-generate-2', 'openapi-compare')
    }

    $firstExitCode = 0
    $contractFailure = $false
    $reportFailure = $false
    $cleanupFailure = $false
    $deadlineLayer = $null
    $reportOwned = $false
    $reportPath = $null
    $sourceSha = $null
    $images = $null
    $imagesLockSha256 = $null
    $composeSnapshot = $null
    $composeSha256 = $null
    $project = $null
    $container = $null
    $composeEnvironment = $null
    $imageAttestation = $null
    $composeResult = $null
    $targetResult = $null
    $commandReceipts = @()
    $persistedArtifacts = New-Object 'System.Collections.Generic.List[string]'
    $openApiChangedFiles = New-Object 'System.Collections.Generic.List[string]'
    $startedAt = [DateTime]::UtcNow
    $runStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $cleanupRecord = [ordered]@{
        status = 'not-attempted'
        selected_owned_count = 0
        foreign_count = 0
        residual_owned_count = $null
        duration_ms = 0
        error_code = $null
    }
    $dependencyContext = [pscustomobject][ordered]@{
        Images = $null
        Project = $null
        Container = $null
        RepositoryRoot = $null
        ComposePath = $null
        ReportPath = $null
        SourceSha = $null
        ComposeEnvironment = $null
    }

    $AssertRequiredProperty = {
        param([object]$Value, [string]$Name)

        if ($null -eq $Value -or $null -eq $Value.PSObject.Properties[$Name]) {
            throw 'A required private-contract property is absent.'
        }
    }.GetNewClosure()

    $ResolveCanonicalContainedPath = {
        param(
            [string]$RootPath,
            [string]$RelativePath,
            [bool]$MustExist,
            [bool]$AllowRoot
        )

        if ([string]::IsNullOrWhiteSpace($RootPath) -or [string]::IsNullOrWhiteSpace($RelativePath)) {
            throw 'A canonical path input is absent.'
        }
        if ([System.IO.Path]::IsPathRooted($RelativePath)) {
            throw 'A caller-controlled absolute path is forbidden.'
        }

        $rootFull = [System.IO.Path]::GetFullPath($RootPath).TrimEnd('\')
        $candidateFull = [System.IO.Path]::GetFullPath((Join-Path $rootFull $RelativePath)).TrimEnd('\')
        $isRoot = $candidateFull.Equals($rootFull, [System.StringComparison]::OrdinalIgnoreCase)
        if ((-not $isRoot -or -not $AllowRoot) -and -not $candidateFull.StartsWith(($rootFull + '\'), [System.StringComparison]::OrdinalIgnoreCase)) {
            throw 'A canonical path escapes its closed root.'
        }
        if (-not [System.IO.Directory]::Exists($rootFull) -and -not [System.IO.File]::Exists($rootFull)) {
            throw 'A canonical root is absent.'
        }

        $walkPath = $candidateFull
        while ($true) {
            if ([System.IO.Directory]::Exists($walkPath) -or [System.IO.File]::Exists($walkPath)) {
                $attributes = [System.IO.File]::GetAttributes($walkPath)
                if (($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                    throw 'A canonical path contains a reparse point.'
                }
            }
            if ($walkPath.Equals($rootFull, [System.StringComparison]::OrdinalIgnoreCase)) {
                break
            }
            $parentPath = [System.IO.Directory]::GetParent($walkPath)
            if ($null -eq $parentPath) {
                throw 'A canonical path has no contained parent.'
            }
            $walkPath = $parentPath.FullName.TrimEnd('\')
        }
        if ($MustExist -and -not [System.IO.Directory]::Exists($candidateFull) -and -not [System.IO.File]::Exists($candidateFull)) {
            throw 'A required canonical path is absent.'
        }
        return $candidateFull
    }.GetNewClosure()

    $ReadValidatedImagesLock = {
        param([object]$Snapshot)

        if (
            $null -eq $Snapshot -or
            $null -eq $Snapshot.PSObject.Properties['Bytes'] -or
            $null -eq $Snapshot.PSObject.Properties['Sha256']
        ) {
            throw 'The closed image lock snapshot is absent.'
        }
        $strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
        try {
            $lockText = $strictUtf8.GetString([byte[]]$Snapshot.Bytes)
        }
        catch {
            throw 'The closed image lock is not strict UTF-8.'
        }
        if ($lockText -cnotmatch $lockDocumentPattern) {
            throw 'The closed image lock has an invalid document shape.'
        }
        try {
            $parsedDocument = $lockText | ConvertFrom-Json -ErrorAction Stop
        }
        catch {
            throw 'The closed image lock is not valid JSON.'
        }
        $memberMatches = [System.Text.RegularExpressions.Regex]::Matches($lockText, $lockMemberPattern)
        if ($memberMatches.Count -ne 3) {
            throw 'The closed image lock does not have exactly three members.'
        }
        $validatedImages = @{}
        foreach ($memberMatch in $memberMatches) {
            $key = $memberMatch.Groups['key'].Value
            $value = $memberMatch.Groups['value'].Value
            if ($fixedTargets -cnotcontains $key -or $validatedImages.ContainsKey($key)) {
                throw 'The closed image lock has an unexpected or duplicate key.'
            }
            if ($value -cnotmatch $digestReferencePattern) {
                throw 'The closed image lock has a non-digest reference.'
            }
            $validatedImages[$key] = $value
        }
        $parsedProperties = @($parsedDocument.PSObject.Properties)
        if ($parsedProperties.Count -ne 3) {
            throw 'The closed image lock parser view is not exact.'
        }
        foreach ($fixedTarget in $fixedTargets) {
            if (
                -not $validatedImages.ContainsKey($fixedTarget) -or
                $null -eq $parsedDocument.PSObject.Properties[$fixedTarget] -or
                $parsedDocument.PSObject.Properties[$fixedTarget].Value -isnot [string] -or
                [string]$parsedDocument.PSObject.Properties[$fixedTarget].Value -cne [string]$validatedImages[$fixedTarget]
            ) {
                throw 'The closed image lock is missing a fixed target.'
            }
        }
        foreach ($parsedProperty in $parsedProperties) {
            if ($fixedTargets -cnotcontains [string]$parsedProperty.Name) {
                throw 'The closed image lock parser view has an unexpected key.'
            }
        }
        return [pscustomobject][ordered]@{
            References = $validatedImages
            Sha256 = [string]$Snapshot.Sha256
        }
    }.GetNewClosure()

    $WriteAtomicBytes = {
        param([string]$DestinationPath, [byte[]]$Bytes)

        $destinationFull = [System.IO.Path]::GetFullPath($DestinationPath)
        $parentPath = [System.IO.Path]::GetDirectoryName($destinationFull)
        if ([string]::IsNullOrWhiteSpace($parentPath) -or -not [System.IO.Directory]::Exists($parentPath)) {
            throw 'An atomic write parent is absent.'
        }
        foreach ($checkedPath in @($parentPath, $destinationFull)) {
            if ([System.IO.Directory]::Exists($checkedPath) -or [System.IO.File]::Exists($checkedPath)) {
                $attributes = [System.IO.File]::GetAttributes($checkedPath)
                if (($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                    throw 'An atomic write path is a reparse point.'
                }
            }
        }

        $temporaryPath = Join-Path $parentPath ('.hubqa-write-' + [System.Guid]::NewGuid().ToString('N') + '.tmp')
        try {
            $temporaryStream = [System.IO.FileStream]::new(
                $temporaryPath,
                [System.IO.FileMode]::CreateNew,
                [System.IO.FileAccess]::Write,
                [System.IO.FileShare]::None
            )
            try {
                $temporaryStream.Write($Bytes, 0, $Bytes.Length)
                $temporaryStream.Flush($true)
            }
            finally {
                $temporaryStream.Dispose()
            }
            if ([System.IO.File]::Exists($destinationFull)) {
                [System.IO.File]::Replace($temporaryPath, $destinationFull, $null)
            }
            else {
                [System.IO.File]::Move($temporaryPath, $destinationFull)
            }
        }
        finally {
            if ([System.IO.File]::Exists($temporaryPath)) {
                [System.IO.File]::Delete($temporaryPath)
            }
        }
    }.GetNewClosure()

    $ReadStableContainedFileBytes = {
        param(
            [string]$CanonicalRoot,
            [string]$ClosedRelativePath,
            [int64]$MaximumBytes,
            [bool]$AllowEmpty
        )

        # Bounded snapshot only: pre/post reparse checks plus FileShare.Read deny write/delete sharing while read.
        if ($MaximumBytes -le 0 -or $MaximumBytes -gt [int]::MaxValue) {
            throw 'A fixed stable-snapshot byte cap is invalid.'
        }
        $beforePath = & $ResolveCanonicalContainedPath $CanonicalRoot $ClosedRelativePath $true $false
        if (-not [System.IO.File]::Exists($beforePath)) {
            throw 'A stable-snapshot input is not a file.'
        }
        $beforeAttributes = [System.IO.File]::GetAttributes($beforePath)
        if (($beforeAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A stable-snapshot input is a reparse point.'
        }
        $announcedLength = [int64]([System.IO.FileInfo]::new($beforePath).Length)
        if (
            $announcedLength -lt 0 -or
            $announcedLength -gt $MaximumBytes -or
            ($announcedLength -eq 0 -and -not $AllowEmpty)
        ) {
            throw 'A stable-snapshot input length violates its fixed cap.'
        }

        $stream = $null
        $algorithm = $null
        try {
            $stream = [System.IO.FileStream]::new(
                $beforePath,
                [System.IO.FileMode]::Open,
                [System.IO.FileAccess]::Read,
                [System.IO.FileShare]::Read
            )
            $bytes = [byte[]]::new([int]$announcedLength)
            $offset = 0
            while ($offset -lt $bytes.Length) {
                $read = $stream.Read($bytes, $offset, ($bytes.Length - $offset))
                if ($read -le 0) {
                    throw 'A stable-snapshot input ended before its announced length.'
                }
                $offset += $read
            }
            if ([int64]$stream.Length -ne $announcedLength) {
                throw 'A stable-snapshot input length changed during the read.'
            }
            $afterPath = & $ResolveCanonicalContainedPath $CanonicalRoot $ClosedRelativePath $true $false
            if ($afterPath -cne $beforePath) {
                throw 'A stable-snapshot input changed its canonical path.'
            }
            $afterAttributes = [System.IO.File]::GetAttributes($afterPath)
            if (($afterAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'A stable-snapshot input became a reparse point.'
            }
            if ([int64]([System.IO.FileInfo]::new($afterPath).Length) -ne $announcedLength) {
                throw 'A stable-snapshot input final length changed.'
            }
            $algorithm = [System.Security.Cryptography.SHA256]::Create()
            $sha256 = [System.BitConverter]::ToString($algorithm.ComputeHash($bytes)).Replace('-', '').ToLowerInvariant()
            return [pscustomobject][ordered]@{
                Path = $afterPath
                Bytes = $bytes
                Sha256 = $sha256
            }
        }
        finally {
            if ($null -ne $algorithm) {
                $algorithm.Dispose()
            }
            if ($null -ne $stream) {
                $stream.Dispose()
            }
        }
    }.GetNewClosure()

    $GetCanonicalSourceSha = {
        param([string]$RepositoryRoot)

        $canonicalRepositoryRoot = & $ResolveCanonicalContainedPath $RepositoryRoot '.' $true $true
        $gitMarker = & $ResolveCanonicalContainedPath $canonicalRepositoryRoot '.git' $true $false
        $referenceRoot = $null
        $headRelativePath = $null
        if ([System.IO.Directory]::Exists($gitMarker)) {
            $primaryAttributes = [System.IO.File]::GetAttributes($gitMarker)
            if (($primaryAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The primary git directory is unsafe.'
            }
            $referenceRoot = $gitMarker
            $headRelativePath = 'HEAD'
        }
        elseif ([System.IO.File]::Exists($gitMarker)) {
            $markerAttributes = [System.IO.File]::GetAttributes($gitMarker)
            if (($markerAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree git marker is unsafe.'
            }
            $markerText = [System.IO.File]::ReadAllText($gitMarker)
            $markerMatch = [System.Text.RegularExpressions.Regex]::Match($markerText, '^gitdir: (?<path>[^\r\n\0]+)\r?\n?$')
            if (-not $markerMatch.Success) {
                throw 'The linked-worktree git marker is malformed.'
            }
            $markerDirectoryText = $markerMatch.Groups['path'].Value
            $linkedGitDirectory = if ([System.IO.Path]::IsPathRooted($markerDirectoryText)) {
                [System.IO.Path]::GetFullPath($markerDirectoryText)
            }
            else {
                [System.IO.Path]::GetFullPath((Join-Path $canonicalRepositoryRoot $markerDirectoryText))
            }
            if (-not [System.IO.Directory]::Exists($linkedGitDirectory)) {
                throw 'The linked-worktree administrative directory is absent.'
            }
            $linkedAttributes = [System.IO.File]::GetAttributes($linkedGitDirectory)
            if (($linkedAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree administrative directory is unsafe.'
            }

            $backlinkPath = & $ResolveCanonicalContainedPath $linkedGitDirectory 'gitdir' $true $false
            $backlinkAttributes = [System.IO.File]::GetAttributes($backlinkPath)
            if (($backlinkAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree gitdir backlink is unsafe.'
            }
            $backlinkText = [System.IO.File]::ReadAllText($backlinkPath)
            $backlinkMatch = [System.Text.RegularExpressions.Regex]::Match($backlinkText, '^(?<path>[^\r\n\0]+)\r?\n?$')
            if (-not $backlinkMatch.Success) {
                throw 'The linked-worktree gitdir backlink is malformed.'
            }
            $backlinkTarget = $backlinkMatch.Groups['path'].Value
            $backlinkCanonical = if ([System.IO.Path]::IsPathRooted($backlinkTarget)) {
                [System.IO.Path]::GetFullPath($backlinkTarget)
            }
            else {
                [System.IO.Path]::GetFullPath((Join-Path $linkedGitDirectory $backlinkTarget))
            }
            if (-not $backlinkCanonical.Equals($gitMarker, [System.StringComparison]::OrdinalIgnoreCase)) {
                throw 'The linked-worktree gitdir backlink does not bind this repository marker.'
            }

            $commonMarkerPath = & $ResolveCanonicalContainedPath $linkedGitDirectory 'commondir' $true $false
            $commonMarkerAttributes = [System.IO.File]::GetAttributes($commonMarkerPath)
            if (($commonMarkerAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree common-directory marker is unsafe.'
            }
            $commonMarkerText = [System.IO.File]::ReadAllText($commonMarkerPath)
            $commonMarkerMatch = [System.Text.RegularExpressions.Regex]::Match($commonMarkerText, '^(?<path>[^\r\n\0]+)\r?\n?$')
            if (-not $commonMarkerMatch.Success) {
                throw 'The linked-worktree common-directory marker is malformed.'
            }
            $commonDirectoryText = $commonMarkerMatch.Groups['path'].Value
            if ([System.IO.Path]::IsPathRooted($commonDirectoryText)) {
                throw 'The linked-worktree common-directory marker must be relative.'
            }
            $referenceRoot = [System.IO.Path]::GetFullPath((Join-Path $linkedGitDirectory $commonDirectoryText))
            if (-not [System.IO.Directory]::Exists($referenceRoot)) {
                throw 'The linked-worktree common directory is absent.'
            }
            $commonAttributes = [System.IO.File]::GetAttributes($referenceRoot)
            if (($commonAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree common directory is unsafe.'
            }
            $worktreesRoot = [System.IO.Path]::GetFullPath((Join-Path $referenceRoot 'worktrees'))
            if (-not [System.IO.Directory]::Exists($worktreesRoot)) {
                throw 'The linked-worktree administrative root is absent.'
            }
            $worktreesAttributes = [System.IO.File]::GetAttributes($worktreesRoot)
            if (($worktreesAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw 'The linked-worktree administrative root is unsafe.'
            }
            if (-not $linkedGitDirectory.StartsWith(($worktreesRoot.TrimEnd('\') + '\'), [System.StringComparison]::OrdinalIgnoreCase)) {
                throw 'The linked-worktree administrative directory escapes the trusted worktrees root.'
            }
            $walkDirectory = $linkedGitDirectory
            while (-not $walkDirectory.Equals($worktreesRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
                $walkAttributes = [System.IO.File]::GetAttributes($walkDirectory)
                if (($walkAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                    throw 'The linked-worktree administrative chain contains a reparse point.'
                }
                $walkParent = [System.IO.Directory]::GetParent($walkDirectory)
                if ($null -eq $walkParent) {
                    throw 'The linked-worktree administrative chain has no trusted parent.'
                }
                $walkDirectory = $walkParent.FullName.TrimEnd('\')
            }
            $headRelativePath = ($linkedGitDirectory.Substring($referenceRoot.Length).TrimStart('\') + '\HEAD')
        }
        else {
            throw 'The canonical git metadata is absent.'
        }

        $headPath = & $ResolveCanonicalContainedPath $referenceRoot $headRelativePath $true $false
        $head = [System.IO.File]::ReadAllText($headPath).Trim()
        if ($head -match '^[a-f0-9]{40}$') {
            return $head
        }
        if (-not $head.StartsWith('ref: ', [System.StringComparison]::Ordinal)) {
            throw 'The canonical git HEAD is malformed.'
        }

        $slashReference = $head.Substring(5).Trim()
        if (
            [string]::IsNullOrWhiteSpace($slashReference) -or
            $slashReference.StartsWith('/') -or
            $slashReference.StartsWith('\') -or
            [System.IO.Path]::IsPathRooted($slashReference)
        ) {
            throw 'The canonical git reference is unsafe.'
        }
        $referenceSegments = @($slashReference.Split('/'))
        if ($referenceSegments.Count -lt 2 -or $referenceSegments[0] -cne 'refs') {
            throw 'The canonical git reference is not a closed refs path.'
        }
        foreach ($segment in $referenceSegments) {
            if (
                [string]::IsNullOrWhiteSpace($segment) -or
                $segment -ceq '.' -or
                $segment -ceq '..' -or
                $segment -cnotmatch '^[A-Za-z0-9][A-Za-z0-9._-]*$'
            ) {
                throw 'The canonical git reference has an unsafe segment.'
            }
        }

        $platformReference = $referenceSegments -join '\'
        $referencePath = & $ResolveCanonicalContainedPath $referenceRoot $platformReference $false $false
        if ([System.IO.File]::Exists($referencePath)) {
            $referenceSha = [System.IO.File]::ReadAllText($referencePath).Trim()
            if ($referenceSha -match '^[a-f0-9]{40}$') {
                return $referenceSha
            }
        }

        $packedRefsPath = & $ResolveCanonicalContainedPath $referenceRoot 'packed-refs' $false $false
        if ([System.IO.File]::Exists($packedRefsPath)) {
            foreach ($line in [System.IO.File]::ReadAllLines($packedRefsPath)) {
                if ($line -match '^([a-f0-9]{40})\s+([^\s]+)$' -and $Matches[2] -ceq $slashReference) {
                    return $Matches[1]
                }
            }
        }
        throw 'The canonical git reference cannot be resolved.'
    }.GetNewClosure()

    $QuoteWindowsArgument = {
        param([string]$Argument)

        if ($null -eq $Argument -or $Argument.Length -eq 0) {
            return '""'
        }
        if ($Argument -notmatch '[\s"]') {
            return $Argument
        }
        $builder = New-Object System.Text.StringBuilder
        [void]$builder.Append('"')
        $backslashes = 0
        foreach ($character in $Argument.ToCharArray()) {
            if ($character -eq '\') {
                $backslashes++
                continue
            }
            if ($character -eq '"') {
                [void]$builder.Append([string]::new([char]92, (($backslashes * 2) + 1)))
                [void]$builder.Append('"')
                $backslashes = 0
                continue
            }
            if ($backslashes -gt 0) {
                [void]$builder.Append([string]::new([char]92, $backslashes))
                $backslashes = 0
            }
            [void]$builder.Append($character)
        }
        if ($backslashes -gt 0) {
            [void]$builder.Append([string]::new([char]92, ($backslashes * 2)))
        }
        [void]$builder.Append('"')
        return $builder.ToString()
    }.GetNewClosure()

    $StartFixedDockerProcess = {
        param(
            [ValidateSet('inspect', 'run', 'list', 'cleanup')]
            [string]$Operation,

            [ValidateSet('', 'container', 'network', 'volume')]
            [string]$ResourceKind,

            [string]$ExactId,
            [string]$ImageReference,
            [string]$ClosedComposePath,
            [string]$ClosedProject,
            [string]$ClosedRepositoryRoot,
            [System.Collections.IDictionary]$ClosedComposeEnvironment
        )

        $dockerExecutable = 'C:\Program Files\Docker\Docker\resources\bin\docker.exe'
        if (-not [System.IO.File]::Exists($dockerExecutable)) {
            throw 'The fixed Docker executable is unavailable.'
        }

        $arguments = @()
        $environment = [ordered]@{}
        switch ($Operation) {
            'inspect' {
                $arguments = @('image', 'inspect', $ImageReference)
            }
            'run' {
                if ($null -eq $ClosedComposeEnvironment) {
                    throw 'The fixed Compose environment has not been attested.'
                }
                $arguments = @(
                    'compose', '--ansi', 'never',
                    '-f', $ClosedComposePath,
                    '--project-directory', $ClosedRepositoryRoot,
                    '-p', $ClosedProject,
                    'up', '--no-build', '--pull', 'never',
                    '--abort-on-container-exit', '--exit-code-from', $Request.Target,
                    $Request.Target
                )
                $environment = $ClosedComposeEnvironment
            }
            'list' {
                switch ($ResourceKind) {
                    'container' {
                        $arguments = @('container', 'ls', '--all', '--filter', ('label=com.docker.compose.project=' + $ClosedProject), '--format', '{{json .}}')
                    }
                    'network' {
                        $arguments = @('network', 'ls', '--filter', ('label=com.docker.compose.project=' + $ClosedProject), '--format', '{{json .}}')
                    }
                    'volume' {
                        $arguments = @('volume', 'ls', '--filter', ('label=com.docker.compose.project=' + $ClosedProject), '--format', '{{json .}}')
                    }
                    default {
                        throw 'A closed inventory kind is required.'
                    }
                }
            }
            'cleanup' {
                if ([string]::IsNullOrWhiteSpace($ExactId)) {
                    throw 'An exact owned resource identifier is required.'
                }
                switch ($ResourceKind) {
                    'container' {
                        $arguments = @('container', 'rm', '--force', $ExactId)
                    }
                    'network' {
                        $arguments = @('network', 'rm', $ExactId)
                    }
                    default {
                        throw 'Only closed container or network deletion is permitted.'
                    }
                }
            }
        }

        $startInfo = New-Object System.Diagnostics.ProcessStartInfo
        $startInfo.FileName = $dockerExecutable
        $startInfo.WorkingDirectory = $ClosedRepositoryRoot
        $startInfo.UseShellExecute = $false
        $startInfo.CreateNoWindow = $true
        $startInfo.RedirectStandardOutput = $true
        $startInfo.RedirectStandardError = $true
        $startInfo.EnvironmentVariables.Clear()
        foreach ($environmentKey in @($environment.Keys | Sort-Object)) {
            $startInfo.EnvironmentVariables[$environmentKey] = [string]$environment[$environmentKey]
        }
        $quotedArguments = @()
        foreach ($argument in $arguments) {
            $quotedArguments += (& $QuoteWindowsArgument ([string]$argument))
        }
        $startInfo.Arguments = $quotedArguments -join ' '

        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $startInfo
        if (-not $process.Start()) {
            $process.Dispose()
            throw 'The fixed Docker process did not start.'
        }
        return [HubQaRunningProcessHandle]::new(
            $process,
            $process.StandardOutput.BaseStream,
            $process.StandardError.BaseStream
        )
    }.GetNewClosure()

    $CompleteFixedDockerProcess = {
        param(
            [HubQaRunningProcessHandle]$Handle,
            [int64]$DeadlineTick,
            [string]$TimeoutLayer,
            [ValidateSet('inspect', 'run', 'list', 'cleanup')]
            [string]$Operation
        )

        $entryNowTick = [System.Diagnostics.Stopwatch]::GetTimestamp()
        $timedOut = ($entryNowTick -ge $DeadlineTick)
        $streamCompleted = -not $timedOut
        $forcedTermination = $timedOut
        $stdout = ''
        $stderr = ''
        $stdoutBytes = [byte[]]@()
        $stderrBytes = [byte[]]@()
        $exitCode = 65
        if ($timedOut) {
            try {
                $Handle.Process.Kill()
            }
            catch {
            }
        }
        $stdoutCap = switch ($Operation) {
            'inspect' { 131072 }
            'list' { 262144 }
            'cleanup' { 65536 }
            'run' { 4194304 }
        }
        $stderrCap = switch ($Operation) {
            'inspect' { 65536 }
            'list' { 65536 }
            'cleanup' { 65536 }
            'run' { 4194304 }
        }
        $stdoutStore = [System.IO.MemoryStream]::new()
        $stderrStore = [System.IO.MemoryStream]::new()
        $streamStates = @(
            [pscustomobject]@{
                Stream = $Handle.StandardOutputStream
                Store = $stdoutStore
                Cap = [int64]$stdoutCap
                Buffer = [byte[]]::new(4096)
                Pending = $null
                Eof = $false
                Failed = $false
                IsStdout = $true
            },
            [pscustomobject]@{
                Stream = $Handle.StandardErrorStream
                Store = $stderrStore
                Cap = [int64]$stderrCap
                Buffer = [byte[]]::new(4096)
                Pending = $null
                Eof = $false
                Failed = $false
                IsStdout = $false
            }
        )
        try {
            while (-not $timedOut) {
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                $allEof = $true
                $madeProgress = $false
                foreach ($streamState in $streamStates) {
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    if ([bool]$streamState.Eof -or [bool]$streamState.Failed) {
                        continue
                    }
                    $allEof = $false
                    if ($null -eq $streamState.Pending) {
                        $remainingCapacity = [int64]$streamState.Cap - [int64]$streamState.Store.Length
                        $readLength = if ($remainingCapacity -le 0) { 1 } else { [int][Math]::Min(4096, $remainingCapacity) }
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        $streamState.Pending = $streamState.Stream.ReadAsync($streamState.Buffer, 0, $readLength)
                    }
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    $pendingCompleted = [bool]$streamState.Pending.IsCompleted
                    if ($pendingCompleted) {
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        try {
                            $readCount = [int]$streamState.Pending.GetAwaiter().GetResult()
                        }
                        catch {
                            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                                $timedOut = $true
                                $streamCompleted = $false
                                $forcedTermination = $true
                                try {
                                    $Handle.Process.Kill()
                                }
                                catch {
                                }
                                break
                            }
                            $streamState.Failed = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            continue
                        }
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        $streamState.Pending = $null
                        $madeProgress = $true
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        if ($readCount -eq 0) {
                            $streamState.Eof = $true
                            continue
                        }
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        if (([int64]$streamState.Store.Length + [int64]$readCount) -gt [int64]$streamState.Cap) {
                            $streamState.Failed = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            continue
                        }
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        $streamState.Store.Write($streamState.Buffer, 0, $readCount)
                    }
                }

                if ($timedOut) {
                    break
                }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ($forcedTermination) {
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ($allEof) {
                    $waitNowTick = [System.Diagnostics.Stopwatch]::GetTimestamp()
                    if ($waitNowTick -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    $remainingTicks = $DeadlineTick - $waitNowTick
                    $remainingMilliseconds = [int][Math]::Min([Math]::Floor(($remainingTicks * 1000.0) / [System.Diagnostics.Stopwatch]::Frequency), [int]::MaxValue)
                    if ($remainingMilliseconds -le 0) {
                        continue
                    }
                    $waitNowTick = [System.Diagnostics.Stopwatch]::GetTimestamp()
                    if ($waitNowTick -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    $remainingTicks = $DeadlineTick - $waitNowTick
                    $remainingMilliseconds = [int][Math]::Min([Math]::Floor(($remainingTicks * 1000.0) / [System.Diagnostics.Stopwatch]::Frequency), [int]::MaxValue)
                    if ($remainingMilliseconds -le 0) {
                        continue
                    }
                    $processExited = $Handle.Process.WaitForExit($remainingMilliseconds)
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    if (-not $processExited) {
                        continue
                    }
                    break
                }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ($madeProgress) {
                    continue
                }

                $pendingTasks = New-Object 'System.Collections.Generic.List[System.Threading.Tasks.Task]'
                foreach ($streamState in $streamStates) {
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        break
                    }
                    $hasPendingRead = ($null -ne $streamState.Pending)
                    if ($hasPendingRead) {
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                            $timedOut = $true
                            $streamCompleted = $false
                            $forcedTermination = $true
                            try {
                                $Handle.Process.Kill()
                            }
                            catch {
                            }
                            break
                        }
                        if (-not [bool]$streamState.Pending.IsCompleted) {
                            [void]$pendingTasks.Add([System.Threading.Tasks.Task]$streamState.Pending)
                        }
                    }
                }
                if ($timedOut) {
                    break
                }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ($pendingTasks.Count -eq 0) {
                    continue
                }
                $waitNowTick = [System.Diagnostics.Stopwatch]::GetTimestamp()
                if ($waitNowTick -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                $remainingTicks = $DeadlineTick - $waitNowTick
                $remainingMilliseconds = [int][Math]::Min([Math]::Floor(($remainingTicks * 1000.0) / [System.Diagnostics.Stopwatch]::Frequency), [int]::MaxValue)
                $waitMilliseconds = [int][Math]::Min($remainingMilliseconds, 50)
                if ($waitMilliseconds -le 0) {
                    continue
                }
                $waitNowTick = [System.Diagnostics.Stopwatch]::GetTimestamp()
                if ($waitNowTick -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                $remainingTicks = $DeadlineTick - $waitNowTick
                $remainingMilliseconds = [int][Math]::Min([Math]::Floor(($remainingTicks * 1000.0) / [System.Diagnostics.Stopwatch]::Frequency), [int]::MaxValue)
                $waitMilliseconds = [int][Math]::Min($remainingMilliseconds, 50)
                if ($waitMilliseconds -le 0) {
                    continue
                }
                $waitIndex = [System.Threading.Tasks.Task]::WaitAny([System.Threading.Tasks.Task[]]$pendingTasks.ToArray(), $waitMilliseconds)
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    break
                }
                if ($waitIndex -lt 0) {
                    continue
                }
            }

            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
                $streamCompleted = $false
                $forcedTermination = $true
                try {
                    $Handle.Process.Kill()
                }
                catch {
                }
            }
            if (-not $timedOut) {
                $stdoutBytes = $stdoutStore.ToArray()
            }
            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
                $streamCompleted = $false
                $forcedTermination = $true
                try {
                    $Handle.Process.Kill()
                }
                catch {
                }
            }
            if (-not $timedOut) {
                $stderrBytes = $stderrStore.ToArray()
            }
            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
                $streamCompleted = $false
                $forcedTermination = $true
                try {
                    $Handle.Process.Kill()
                }
                catch {
                }
            }
            if (-not $timedOut -and $streamCompleted) {
                try {
                    $strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
                    $stdout = $strictUtf8.GetString($stdoutBytes)
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                    }
                    if (-not $timedOut) {
                        $stderr = $strictUtf8.GetString($stderrBytes)
                    }
                }
                catch {
                    $streamCompleted = $false
                }
            }
            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
                $streamCompleted = $false
                $forcedTermination = $true
                try {
                    $Handle.Process.Kill()
                }
                catch {
                }
            }
            if ($timedOut) {
                $exitCode = 124
            }
            elseif (-not $streamCompleted -or $forcedTermination) {
                $exitCode = 65
            }
            else {
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                    $timedOut = $true
                    $streamCompleted = $false
                    $forcedTermination = $true
                    try {
                        $Handle.Process.Kill()
                    }
                    catch {
                    }
                    $exitCode = 124
                }
                else {
                    $observedExitCode = [int]$Handle.Process.ExitCode
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                        $timedOut = $true
                        $streamCompleted = $false
                        $forcedTermination = $true
                        try {
                            $Handle.Process.Kill()
                        }
                        catch {
                        }
                        $exitCode = 124
                    }
                    else {
                        $exitCode = $observedExitCode
                    }
                }
            }
        }
        catch {
            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
            }
            $streamCompleted = $false
            $forcedTermination = $true
            try {
                $Handle.Process.Kill()
            }
            catch {
            }
            if (-not $timedOut -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $DeadlineTick) {
                $timedOut = $true
            }
            $exitCode = if ($timedOut) { 124 } else { 65 }
        }
        finally {
            $stdoutStore.Dispose()
            $stderrStore.Dispose()
            $Handle.StandardOutputStream.Dispose()
            $Handle.StandardErrorStream.Dispose()
            $Handle.Process.Dispose()
        }
        return [pscustomobject][ordered]@{
            ExitCode = $exitCode
            TimedOut = $timedOut
            DeadlineLayer = if ($timedOut) { $TimeoutLayer } else { $null }
            StreamCompleted = $streamCompleted
            Stdout = $stdout
            Stderr = $stderr
            StdoutBytes = $stdoutBytes
            StderrBytes = $stderrBytes
        }
    }.GetNewClosure()

    $InspectLockedImage = {
        $frozenImageReference = [string]$dependencyContext.Images.References[$Request.Target]
        $frozenRepositoryRoot = [string]$dependencyContext.RepositoryRoot
        $deadlineTick = [System.Diagnostics.Stopwatch]::GetTimestamp() + [int64]($inspectTimeoutMilliseconds * [System.Diagnostics.Stopwatch]::Frequency / 1000)
        $handle = & $StartFixedDockerProcess 'inspect' '' '' $frozenImageReference '' '' $frozenRepositoryRoot $null
        $completed = & $CompleteFixedDockerProcess $handle $deadlineTick 'inspect' 'inspect'
        foreach ($requiredInspectProcessField in @('ExitCode', 'TimedOut', 'DeadlineLayer', 'StreamCompleted', 'Stdout', 'Stderr', 'StdoutBytes', 'StderrBytes')) {
            & $AssertRequiredProperty $completed $requiredInspectProcessField
        }
        if (
            $completed.ExitCode -isnot [int] -or
            $completed.TimedOut -isnot [bool] -or
            $completed.StreamCompleted -isnot [bool] -or
            $completed.Stdout -isnot [string] -or
            $completed.Stderr -isnot [string] -or
            $completed.StdoutBytes -isnot [byte[]] -or
            $completed.StderrBytes -isnot [byte[]] -or
            ($null -ne $completed.DeadlineLayer -and $completed.DeadlineLayer -isnot [string])
        ) {
            return [pscustomobject][ordered]@{
                Verified = $false
                ExitCode = 65
                DeadlineLayer = $null
            }
        }
        $inspectExitCode = [int]$completed.ExitCode
        $inspectTimedOut = [bool]$completed.TimedOut
        $inspectDeadlineLayer = $completed.DeadlineLayer
        if (
            ($inspectTimedOut -and ($inspectExitCode -ne 124 -or [string]$inspectDeadlineLayer -cne 'inspect')) -or
            ((-not $inspectTimedOut) -and $null -ne $inspectDeadlineLayer)
        ) {
            return [pscustomobject][ordered]@{
                Verified = $false
                ExitCode = 65
                DeadlineLayer = $null
            }
        }
        if ($inspectTimedOut) {
            return [pscustomobject][ordered]@{
                Verified = $false
                ExitCode = 124
                DeadlineLayer = 'inspect'
            }
        }
        if ($inspectExitCode -ne 0 -or -not [bool]$completed.StreamCompleted) {
            return [pscustomobject][ordered]@{
                Verified = $false
                ExitCode = $inspectExitCode
                DeadlineLayer = $null
            }
        }
        try {
            $documents = @($completed.Stdout | ConvertFrom-Json)
            if ($documents.Count -ne 1) {
                throw 'The image inspection document count is invalid.'
            }
            $document = $documents[0]
            & $AssertRequiredProperty $document 'RepoDigests'
            & $AssertRequiredProperty $document 'Config'
            & $AssertRequiredProperty $document.Config 'Entrypoint'
            & $AssertRequiredProperty $document.Config 'Labels'
            $expectedEntrypoint = ('/opt/hub/bin/' + $Request.Target)
            $repoDigests = @($document.RepoDigests)
            $entrypoint = @($document.Config.Entrypoint)
            $labels = $document.Config.Labels
            $entrypointSha = [string]$labels.'com.hub.qa.entrypoint.sha256'
            $configSha = [string]$labels.'com.hub.qa.config.sha256'
            $toolAttestationSha256 = [string]$labels.'com.hub.qa.tools.sha256'
            $verified = (
                $repoDigests -ccontains $frozenImageReference -and
                $entrypoint.Count -eq 1 -and
                $entrypoint[0] -ceq $expectedEntrypoint -and
                $entrypointSha -match '^[a-f0-9]{64}$' -and
                $toolAttestationSha256 -match '^[a-f0-9]{64}$'
            )
            if ($Request.Target -ceq 'frontend-unit') {
                $verified = $verified -and ($configSha -match '^[a-f0-9]{64}$')
            }
            return [pscustomobject][ordered]@{
                Verified = [bool]$verified
                ExitCode = 0
                DeadlineLayer = $null
                ImageReference = $frozenImageReference
                Digest = $frozenImageReference.Substring($frozenImageReference.IndexOf('@sha256:') + 1)
                Entrypoint = $expectedEntrypoint
                EntrypointSha256 = $entrypointSha
                ConfigSha256 = if ($Request.Target -ceq 'frontend-unit') { $configSha } else { $null }
                ToolAttestationSha256 = $toolAttestationSha256
                ToolAttestationVerified = [bool]$verified
            }
        }
        catch {
            return [pscustomobject][ordered]@{
                Verified = $false
                ExitCode = 65
                DeadlineLayer = $null
            }
        }
    }.GetNewClosure()

    $RunFixedTarget = {
        $frozenComposePath = [string]$dependencyContext.ComposePath
        $frozenProject = [string]$dependencyContext.Project
        $frozenRepositoryRoot = [string]$dependencyContext.RepositoryRoot
        $frozenComposeEnvironment = [System.Collections.IDictionary]$dependencyContext.ComposeEnvironment
        $deadlineTick = [System.Diagnostics.Stopwatch]::GetTimestamp() + [int64]($outerTimeoutMilliseconds * [System.Diagnostics.Stopwatch]::Frequency / 1000)
        $handle = & $StartFixedDockerProcess 'run' '' '' '' $frozenComposePath $frozenProject $frozenRepositoryRoot $frozenComposeEnvironment
        return (& $CompleteFixedDockerProcess $handle $deadlineTick 'outer' 'run')
    }.GetNewClosure()

    $ListOwnedResources = {
        param([int64]$DeadlineTick)

        $frozenProject = [string]$dependencyContext.Project
        $frozenRepositoryRoot = [string]$dependencyContext.RepositoryRoot
        $rows = New-Object 'System.Collections.Generic.List[object]'
        foreach ($resourceKind in @('container', 'network', 'volume')) {
            $handle = & $StartFixedDockerProcess 'list' $resourceKind '' '' '' $frozenProject $frozenRepositoryRoot $null
            $completed = & $CompleteFixedDockerProcess $handle $DeadlineTick 'cleanup' 'list'
            if ($completed.TimedOut -or $completed.ExitCode -ne 0 -or -not [bool]$completed.StreamCompleted) {
                throw 'The bounded resource inventory failed.'
            }
            foreach ($line in @($completed.Stdout -split '\r?\n')) {
                if ([string]::IsNullOrWhiteSpace($line)) {
                    continue
                }
                try {
                    $record = $line | ConvertFrom-Json -ErrorAction Stop
                    & $AssertRequiredProperty $record 'ID'
                    if ($resourceKind -eq 'container') {
                        & $AssertRequiredProperty $record 'Names'
                    }
                    else {
                        & $AssertRequiredProperty $record 'Name'
                    }
                    $id = [string]$record.ID
                    $name = if ($resourceKind -eq 'container') { [string]$record.Names } else { [string]$record.Name }
                    $labelsText = if ($null -ne $record.PSObject.Properties['Labels']) { [string]$record.Labels } else { '' }
                    if ([string]::IsNullOrWhiteSpace($id) -or [string]::IsNullOrWhiteSpace($name)) {
                        throw 'The bounded resource inventory has an incomplete record.'
                    }
                    $hasExactProjectLabel = @($labelsText -split ',') -ccontains ('com.docker.compose.project=' + $frozenProject)
                    $rows.Add([pscustomobject][ordered]@{
                        RecordType = 'hub.qa.resource/v1'
                        Id = $id
                        Kind = [string]$resourceKind
                        Name = $name
                        Project = if ($hasExactProjectLabel) { $frozenProject } else { '' }
                    })
                }
                catch {
                    $rows.Add([pscustomobject][ordered]@{
                        RecordType = 'hub.qa.malformed-resource/v1'
                        Id = ''
                        Kind = [string]$resourceKind
                        Name = ''
                        Project = ''
                    })
                }
            }
        }
        return $rows.ToArray()
    }.GetNewClosure()

    $CleanupOwnedResources = {
        param(
            [object[]]$OwnedResources,
            [int64]$DeadlineTick
        )

        $frozenProject = [string]$dependencyContext.Project
        $frozenContainer = [string]$dependencyContext.Container
        $frozenRepositoryRoot = [string]$dependencyContext.RepositoryRoot
        $expectedNetwork = ($frozenProject + '_default')
        $seenTuples = @{}
        if ($OwnedResources.Count -gt 2) {
            throw 'The selected owned set exceeds the closed cleanup set.'
        }
        foreach ($resource in @($OwnedResources)) {
            if ($null -eq $resource) {
                throw 'The selected owned set has an absent record.'
            }
            $fields = @($resource.PSObject.Properties.Name)
            if (
                $fields.Count -ne 5 -or
                $fields -cnotcontains 'RecordType' -or
                $fields -cnotcontains 'Id' -or
                $fields -cnotcontains 'Kind' -or
                $fields -cnotcontains 'Name' -or
                $fields -cnotcontains 'Project' -or
                [string]$resource.RecordType -cne 'hub.qa.owned-resource/v1' -or
                [string]$resource.Project -cne $frozenProject -or
                [string]::IsNullOrWhiteSpace([string]$resource.Id)
            ) {
                throw 'The selected owned set is not an exact typed set.'
            }
            $isExpected = (
                ([string]$resource.Kind -ceq 'container' -and [string]$resource.Name -ceq $frozenContainer) -or
                ([string]$resource.Kind -ceq 'network' -and [string]$resource.Name -ceq $expectedNetwork)
            )
            if (-not $isExpected) {
                throw 'The selected owned set contains an unexpected identity.'
            }
            $tuple = ([string]$resource.Kind + '|' + [string]$resource.Id + '|' + [string]$resource.Name)
            if ($seenTuples.ContainsKey($tuple)) {
                throw 'The selected owned set contains a duplicate identity.'
            }
            $seenTuples[$tuple] = $true
        }

        foreach ($resourceKind in @('container', 'network')) {
            $matching = @(
                foreach ($resource in @($OwnedResources)) {
                    if ([string]$resource.Kind -ceq $resourceKind) {
                        $resource
                    }
                }
            )
            if ($matching.Count -gt 1) {
                throw 'The selected owned set has duplicate resource kinds.'
            }
            if ($matching.Count -eq 1) {
                $handle = & $StartFixedDockerProcess 'cleanup' $resourceKind ([string]$matching[0].Id) '' '' $frozenProject $frozenRepositoryRoot $null
                $completed = & $CompleteFixedDockerProcess $handle $DeadlineTick 'cleanup' 'cleanup'
                if ($completed.TimedOut -or $completed.ExitCode -ne 0 -or -not [bool]$completed.StreamCompleted) {
                    return [pscustomobject][ordered]@{
                        Succeeded = $false
                        ExitCode = [int]$completed.ExitCode
                        TimedOut = [bool]$completed.TimedOut
                    }
                }
            }
        }
        return [pscustomobject][ordered]@{
            Succeeded = $true
            ExitCode = 0
            TimedOut = $false
        }
    }.GetNewClosure()

    $NewProductionDependencies = {
        return [HubQaSupervisorDependencies]::new(
            $InspectLockedImage,
            $RunFixedTarget,
            $ListOwnedResources,
            $CleanupOwnedResources
        )
    }.GetNewClosure()

    try {
        if ($Request.Contract -cne $contractLiteral) {
            return 64
        }
        if ($fixedTargets -cnotcontains $Request.Target) {
            return 64
        }
        if ($Request.Task -cnotmatch $idPattern -or $Request.Run -cnotmatch $idPattern) {
            return 64
        }

        $project = ('hubqa-' + $Request.Task + '-' + $Request.Run)
        $container = ($project + '-' + $Request.Target)
        $dependencyContext.Project = $project
        $dependencyContext.Container = $container
        $runnerRoot = [System.IO.Path]::GetFullPath($PSScriptRoot)
        $testframeworkRoot = [System.IO.Path]::GetFullPath([System.IO.Directory]::GetParent($runnerRoot).FullName)
        $repositoryRoot = [System.IO.Path]::GetFullPath([System.IO.Directory]::GetParent($testframeworkRoot).FullName)
        $repositoryRoot = & $ResolveCanonicalContainedPath $repositoryRoot '.' $true $true
        $dependencyContext.RepositoryRoot = $repositoryRoot
        $testframeworkRoot = & $ResolveCanonicalContainedPath $repositoryRoot 'testframework' $true $false
        $runnerRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'runner' $true $false
        $dockerRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'docker' $true $false
        $lockPath = & $ResolveCanonicalContainedPath $dockerRoot 'images.lock.json' $true $false
        $composePath = & $ResolveCanonicalContainedPath $dockerRoot 'compose.unit.yaml' $true $false
        $reportsRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'reports' $true $false
        $backendDocsRoot = & $ResolveCanonicalContainedPath $repositoryRoot 'backend\docs' $true $false
        foreach ($allowedOpenApiPath in $allowedOpenApiPaths) {
            [void](& $ResolveCanonicalContainedPath $repositoryRoot $allowedOpenApiPath.Replace('/', '\') $false $false)
        }
        $reportPath = & $ResolveCanonicalContainedPath $reportsRoot ($Request.Task + '-' + $Request.Run) $false $false
        $dependencyContext.ReportPath = $reportPath
        if ([System.IO.Directory]::Exists($reportPath) -or [System.IO.File]::Exists($reportPath)) {
            return 64
        }

        $lockSnapshot = & $ReadStableContainedFileBytes $dockerRoot 'images.lock.json' 8192 $false
        $images = & $ReadValidatedImagesLock $lockSnapshot
        $dependencyContext.Images = $images
        $imagesLockSha256 = [string]$images.Sha256
        $sourceSha = & $GetCanonicalSourceSha $repositoryRoot
        $dependencyContext.SourceSha = $sourceSha
        $composeSnapshot = & $ReadStableContainedFileBytes $dockerRoot 'compose.unit.yaml' 131072 $false
        $composeSha256 = [string]$composeSnapshot.Sha256
    }
    catch {
        return 64
    }

    $stagingPath = $null
    $stagingOwnerPath = $null
    $stagingOwnerJson = $null
    $stagingOwned = $false
    try {
        $ownerNonce = [System.Guid]::NewGuid().ToString('N')
        $stagingPath = & $ResolveCanonicalContainedPath $reportsRoot ('.staging-' + $Request.Task + '-' + $Request.Run + '-' + $ownerNonce) $false $false
        [System.IO.Directory]::CreateDirectory($stagingPath) | Out-Null
        $stagingPath = & $ResolveCanonicalContainedPath $reportsRoot ([System.IO.Path]::GetFileName($stagingPath)) $true $false
        $stagingOwnerPath = & $ResolveCanonicalContainedPath $stagingPath '.owner.json' $false $false
        $ownerRecord = [ordered]@{
            schema = 'hub.qa-owner/v1'
            contract = $contractLiteral
            task = $Request.Task
            run = $Request.Run
            project = $project
            target = $Request.Target
            nonce = $ownerNonce
            report_relative_path = ($Request.Task + '-' + $Request.Run)
            source_sha = $sourceSha
            images_lock_sha256 = $imagesLockSha256
            compose_sha256 = $composeSha256
            created_at = [DateTime]::UtcNow.ToString('o')
        }
        $stagingOwnerJson = ($ownerRecord | ConvertTo-Json -Compress)
        $ownerBytes = [System.Text.UTF8Encoding]::new($false).GetBytes($stagingOwnerJson + [Environment]::NewLine)
        $ownerStream = [System.IO.FileStream]::new(
            $stagingOwnerPath,
            [System.IO.FileMode]::CreateNew,
            [System.IO.FileAccess]::Write,
            [System.IO.FileShare]::None
        )
        try {
            $ownerStream.Write($ownerBytes, 0, $ownerBytes.Length)
            $ownerStream.Flush($true)
        }
        finally {
            $ownerStream.Dispose()
        }
        $stagingOwned = $true
        $stagingComposePath = & $ResolveCanonicalContainedPath $stagingPath 'compose.unit.yaml' $false $false
        & $WriteAtomicBytes $stagingComposePath ([byte[]]$composeSnapshot.Bytes)
        $stagingComposeSnapshot = & $ReadStableContainedFileBytes $stagingPath 'compose.unit.yaml' 131072 $false
        if ($stagingComposeSnapshot.Sha256 -cne $composeSha256) {
            throw 'The owned Compose snapshot bytes do not match the validated source snapshot.'
        }
        [System.IO.Directory]::Move($stagingPath, $reportPath)
        $reportOwned = $true
        $stagingPath = $null
        $composePath = & $ResolveCanonicalContainedPath $reportPath 'compose.unit.yaml' $true $false
        $dependencyContext.ComposePath = $composePath
        [void]$persistedArtifacts.Add('.owner.json')
        [void]$persistedArtifacts.Add('compose.unit.yaml')
    }
    catch {
        if (
            $stagingOwned -and
            $null -ne $stagingPath -and
            [System.IO.Directory]::Exists($stagingPath) -and
            -not (([System.IO.File]::GetAttributes($stagingPath) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) -and
            [System.IO.File]::Exists($stagingOwnerPath)
        ) {
            try {
                $verifiedStagingPath = & $ResolveCanonicalContainedPath $reportsRoot ([System.IO.Path]::GetFileName($stagingPath)) $true $false
                $verifiedOwnerPath = & $ResolveCanonicalContainedPath $verifiedStagingPath '.owner.json' $true $false
                $verifiedOwnerSnapshot = & $ReadStableContainedFileBytes $verifiedStagingPath '.owner.json' 8192 $false
                if (
                    $verifiedStagingPath -ceq $stagingPath -and
                    $verifiedOwnerPath -ceq $stagingOwnerPath -and
                    $verifiedOwnerSnapshot.Bytes.Length -eq $ownerBytes.Length -and
                    [System.Linq.Enumerable]::SequenceEqual([byte[]]$verifiedOwnerSnapshot.Bytes, [byte[]]$ownerBytes)
                ) {
                    [System.IO.Directory]::Delete($verifiedStagingPath, $true)
                }
            }
            catch {
            }
        }
        return 64
    }

    $dependencies = $null
    try {
        $dependencies = & $NewProductionDependencies
        $imageAttestation = & $dependencies.InspectLockedImage
        foreach ($requiredAttestationField in @('Verified', 'ExitCode', 'DeadlineLayer')) {
            & $AssertRequiredProperty $imageAttestation $requiredAttestationField
        }
        if (
            $imageAttestation.Verified -isnot [bool] -or
            $imageAttestation.ExitCode -isnot [int] -or
            ($null -ne $imageAttestation.DeadlineLayer -and $imageAttestation.DeadlineLayer -isnot [string])
        ) {
            $contractFailure = $true
            throw 'The image attestation result shape is invalid.'
        }
        $attestationExitCode = [int]$imageAttestation.ExitCode
        $attestationDeadlineLayer = $imageAttestation.DeadlineLayer
        if (
            ($null -ne $attestationDeadlineLayer -and (
                [string]$attestationDeadlineLayer -cne 'inspect' -or
                $attestationExitCode -ne 124 -or
                [bool]$imageAttestation.Verified
            )) -or
            ([bool]$imageAttestation.Verified -and $attestationExitCode -ne 0)
        ) {
            $contractFailure = $true
            throw 'The image attestation result transition is invalid.'
        }
        if (-not [bool]$imageAttestation.Verified) {
            if ($firstExitCode -eq 0 -and $attestationExitCode -ne 0) {
                $firstExitCode = $attestationExitCode
                if ([string]$attestationDeadlineLayer -ceq 'inspect') {
                    $deadlineLayer = 'inspect'
                }
            }
            else {
                $contractFailure = $true
            }
            throw 'The locked image does not satisfy the private attestation contract.'
        }

        foreach ($requiredVerifiedAttestationField in @('ImageReference', 'Digest', 'Entrypoint', 'EntrypointSha256', 'ConfigSha256', 'ToolAttestationSha256', 'ToolAttestationVerified')) {
            & $AssertRequiredProperty $imageAttestation $requiredVerifiedAttestationField
        }
        if ($imageAttestation.ToolAttestationVerified -isnot [bool] -or -not [bool]$imageAttestation.ToolAttestationVerified) {
            $contractFailure = $true
            throw 'The verified image attestation result is incomplete.'
        }
        $composeEnvironment = [ordered]@{
            HUB_QA_BACKEND_UNIT_IMAGE = [string]$images.References['backend-unit']
            HUB_QA_FRONTEND_UNIT_IMAGE = [string]$images.References['frontend-unit']
            HUB_QA_BACKEND_OPENAPI_IMAGE = [string]$images.References['backend-openapi']
            HUB_QA_PROJECT = $project
            HUB_QA_CONTAINER = $container
            HUB_QA_REPORT_PATH = $reportPath
            HUB_QA_REPOSITORY_PATH = $repositoryRoot
            HUB_QA_SOURCE_SHA = $sourceSha
            COMPOSE_DISABLE_ENV_FILE = '1'
        }
        $dependencyContext.ComposeEnvironment = $composeEnvironment
        $composeResult = & $dependencies.RunFixedTarget
        foreach ($requiredRunField in @('ExitCode', 'TimedOut', 'DeadlineLayer', 'StreamCompleted', 'Stdout', 'Stderr', 'StdoutBytes', 'StderrBytes')) {
            & $AssertRequiredProperty $composeResult $requiredRunField
        }
        $composeExitCode = [int]$composeResult.ExitCode
        if ($firstExitCode -eq 0 -and $composeExitCode -ne 0) {
            $firstExitCode = $composeExitCode
            if ([bool]$composeResult.TimedOut) {
                $deadlineLayer = 'outer'
            }
        }
        if (
            ([bool]$composeResult.TimedOut -and ($composeExitCode -ne 124 -or [string]$composeResult.DeadlineLayer -cne 'outer')) -or
            ((-not [bool]$composeResult.TimedOut) -and $null -ne $composeResult.DeadlineLayer)
        ) {
            $contractFailure = $true
            throw 'The fixed target process timeout transition is invalid.'
        }
        if (-not [bool]$composeResult.StreamCompleted) {
            $contractFailure = $true
        }

        try {
            & $WriteAtomicBytes (Join-Path $reportPath 'compose.stdout.log') ([byte[]]$composeResult.StdoutBytes)
            [void]$persistedArtifacts.Add('compose.stdout.log')
            & $WriteAtomicBytes (Join-Path $reportPath 'compose.stderr.log') ([byte[]]$composeResult.StderrBytes)
            [void]$persistedArtifacts.Add('compose.stderr.log')
        }
        catch {
            $reportFailure = $true
        }

        if (-not [bool]$composeResult.TimedOut) {
            $commandsSnapshot = & $ReadStableContainedFileBytes $reportPath 'commands.json' 1048576 $false
            $targetResultSnapshot = & $ReadStableContainedFileBytes $reportPath 'target-result.json' 8388608 $false
            $strictReceiptUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
            try {
                $commandReceipts = @($strictReceiptUtf8.GetString([byte[]]$commandsSnapshot.Bytes) | ConvertFrom-Json -ErrorAction Stop)
                $targetResult = $strictReceiptUtf8.GetString([byte[]]$targetResultSnapshot.Bytes) | ConvertFrom-Json -ErrorAction Stop
            }
            catch {
                $contractFailure = $true
                throw 'The stable receipt snapshots are not valid strict UTF-8 JSON.'
            }
            [void]$persistedArtifacts.Add('commands.json')
            [void]$persistedArtifacts.Add('target-result.json')
            if ($commandReceipts.Count -eq 0) {
                $contractFailure = $true
                throw 'The common command receipt sequence is empty.'
            }

            $targetCommands = @($expectedCommands[$Request.Target])
            if ($commandReceipts.Count -gt $targetCommands.Count) {
                $contractFailure = $true
                throw 'The common command receipt sequence is too long.'
            }
            $receiptFirstExit = 0
            $pendingReceiptDeadlineLayer = $null
            $seenArtifactPaths = @{}
            for ($receiptIndex = 0; $receiptIndex -lt $commandReceipts.Count; $receiptIndex++) {
                $receipt = $commandReceipts[$receiptIndex]
                foreach ($requiredReceiptField in @('command_id', 'sequence', 'exit_code', 'deadline_layer', 'stdout_artifact', 'stderr_artifact', 'result_artifact')) {
                    & $AssertRequiredProperty $receipt $requiredReceiptField
                }
                $receiptFields = @($receipt.PSObject.Properties.Name)
                $hasUnexpectedReceiptField = $false
                foreach ($receiptField in $receiptFields) {
                    if (
                        $receiptField -cne 'command_id' -and
                        $receiptField -cne 'sequence' -and
                        $receiptField -cne 'exit_code' -and
                        $receiptField -cne 'deadline_layer' -and
                        $receiptField -cne 'stdout_artifact' -and
                        $receiptField -cne 'stderr_artifact' -and
                        $receiptField -cne 'result_artifact'
                    ) {
                        $hasUnexpectedReceiptField = $true
                    }
                }
                if (
                    $receiptFields.Count -ne 7 -or
                    $hasUnexpectedReceiptField
                ) {
                    $contractFailure = $true
                    throw 'The common command receipt artifact schema is not closed.'
                }
                if (
                    [string]$receipt.command_id -cne $targetCommands[$receiptIndex] -or
                    [int]$receipt.sequence -ne ($receiptIndex + 1)
                ) {
                    $contractFailure = $true
                    throw 'The common command receipt sequence is malformed.'
                }
                $receiptExit = [int]$receipt.exit_code
                $receiptDeadlineLayer = if ($null -eq $receipt.deadline_layer) { '' } else { [string]$receipt.deadline_layer }
                if (
                    ($receiptExit -eq 124 -and ($receiptDeadlineLayer -cne 'command' -and $receiptDeadlineLayer -cne 'target')) -or
                    ($receiptExit -ne 124 -and $receiptDeadlineLayer.Length -ne 0)
                ) {
                    $contractFailure = $true
                    throw 'The common command receipt deadline transition is invalid.'
                }
                $expectedArtifactPaths = [ordered]@{
                    stdout_artifact = ('commands/' + [string]$receipt.command_id + '.stdout.log')
                    stderr_artifact = ('commands/' + [string]$receipt.command_id + '.stderr.log')
                    result_artifact = ('commands/' + [string]$receipt.command_id + '.result.json')
                }
                foreach ($artifactField in @('stdout_artifact', 'stderr_artifact', 'result_artifact')) {
                    $artifactRelativePath = [string]$receipt.$artifactField
                    if (
                        [string]::IsNullOrWhiteSpace($artifactRelativePath) -or
                        $artifactRelativePath -cne [string]$expectedArtifactPaths[$artifactField] -or
                        $seenArtifactPaths.ContainsKey($artifactRelativePath)
                    ) {
                        $contractFailure = $true
                        throw 'A command receipt artifact path is not an exact owned command artifact.'
                    }
                    $allowEmptyArtifact = ($artifactField -ceq 'stdout_artifact' -or $artifactField -ceq 'stderr_artifact')
                    $artifactSnapshot = & $ReadStableContainedFileBytes $reportPath $artifactRelativePath 8388608 $allowEmptyArtifact
                    $seenArtifactPaths[$artifactRelativePath] = $true
                    [void]$persistedArtifacts.Add($artifactRelativePath)
                }
                if ($receiptFirstExit -eq 0 -and $receiptExit -ne 0) {
                    $receiptFirstExit = $receiptExit
                    $pendingReceiptDeadlineLayer = $receiptDeadlineLayer
                    if ($receiptIndex -ne ($commandReceipts.Count - 1)) {
                        $contractFailure = $true
                        throw 'The common command receipt sequence continues after first failure.'
                    }
                }
            }
            if ($receiptFirstExit -eq 0 -and $commandReceipts.Count -ne $targetCommands.Count) {
                $contractFailure = $true
                throw 'The successful common command receipt sequence is incomplete.'
            }
            & $AssertRequiredProperty $targetResult 'exit_code'
            if ($receiptFirstExit -ne $composeExitCode -or [int]$targetResult.exit_code -ne $composeExitCode) {
                $contractFailure = $true
                throw 'The common receipt and target result exit parity is invalid.'
            }
            if ($receiptFirstExit -eq 124 -and $firstExitCode -eq 124) {
                $deadlineLayer = $pendingReceiptDeadlineLayer
            }

            if ($Request.Target -ceq 'backend-unit' -or $Request.Target -ceq 'frontend-unit') {
                foreach ($requiredNativeField in @('collected', 'passed', 'failed', 'skipped', 'unexpected_skip_count', 'discovery_root', 'discovered_files')) {
                    & $AssertRequiredProperty $targetResult $requiredNativeField
                }
                $collected = [int]$targetResult.collected
                $passed = [int]$targetResult.passed
                $failed = [int]$targetResult.failed
                $skipped = [int]$targetResult.skipped
                $unexpectedSkips = [int]$targetResult.unexpected_skip_count
                if (
                    $collected -le 0 -or
                    ($passed + $failed + $skipped) -ne $collected -or
                    $skipped -ne 0 -or
                    $unexpectedSkips -ne 0
                ) {
                    $contractFailure = $true
                    throw 'The native target result violates collection or skip gates.'
                }
                $nativeRoot = if ($Request.Target -ceq 'backend-unit') { 'backend/tests/' } else { 'frontend/tests/' }
                if ([string]$targetResult.discovery_root -cne $nativeRoot) {
                    $contractFailure = $true
                    throw 'The native target discovery root is not canonical.'
                }
                $discoveredFiles = @($targetResult.discovered_files)
                if ($discoveredFiles.Count -eq 0) {
                    $contractFailure = $true
                    throw 'The native target discovery manifest is empty.'
                }
                foreach ($discoveredFile in $discoveredFiles) {
                    $canonicalDiscoveredFile = ([string]$discoveredFile).Replace('\', '/')
                    if (
                        -not $canonicalDiscoveredFile.StartsWith($nativeRoot, [System.StringComparison]::Ordinal) -or
                        $canonicalDiscoveredFile.Substring($nativeRoot.Length).Length -eq 0 -or
                        $canonicalDiscoveredFile.Contains('../') -or
                        $canonicalDiscoveredFile.Contains('//')
                    ) {
                        $contractFailure = $true
                        throw 'The native target discovery manifest escapes its canonical root.'
                    }
                }
            }
            else {
                foreach ($nativeNotApplicableField in @('collected', 'passed', 'failed', 'skipped', 'unexpected_skip_count', 'discovery_root', 'discovered_files')) {
                    & $AssertRequiredProperty $targetResult $nativeNotApplicableField
                    if ([string]$targetResult.$nativeNotApplicableField -cne 'N/A') {
                        $contractFailure = $true
                        throw 'The OpenAPI target must record native result fields as N/A.'
                    }
                }
                foreach ($requiredOpenApiField in @('source_sha', 'generated_count', 'generated_files', 'generation_manifests', 'artifact_revision')) {
                    & $AssertRequiredProperty $targetResult $requiredOpenApiField
                }
                if ([string]$targetResult.source_sha -cne $sourceSha) {
                    $contractFailure = $true
                    throw 'The OpenAPI receipt source revision does not equal the canonical source revision.'
                }
                $generatedFiles = @($targetResult.generated_files)
                if ([int]$targetResult.generated_count -ne 3 -or $generatedFiles.Count -ne 3) {
                    $contractFailure = $true
                    throw 'The OpenAPI generated-file list is not exact.'
                }
                for ($pathIndex = 0; $pathIndex -lt $allowedOpenApiPaths.Count; $pathIndex++) {
                    if ([string]$generatedFiles[$pathIndex] -cne $allowedOpenApiPaths[$pathIndex]) {
                        $contractFailure = $true
                        throw 'The OpenAPI generated-file allowlist is invalid.'
                    }
                }
                $generationManifests = @($targetResult.generation_manifests)
                if ($generationManifests.Count -ne 2) {
                    $contractFailure = $true
                    throw 'The OpenAPI generation-manifest count is invalid.'
                }

                $revisionRows = New-Object 'System.Collections.Generic.List[string]'
                $revisionRows.Add(('source_sha' + [char]9 + $sourceSha))
                $generationOneRoot = $null
                for ($generation = 1; $generation -le 2; $generation++) {
                    $generationRoot = & $ResolveCanonicalContainedPath $reportPath ('openapi\generation-' + $generation) $true $false
                    if ($generation -eq 1) {
                        $generationOneRoot = $generationRoot
                    }
                    $pendingDirectories = New-Object 'System.Collections.Generic.Stack[string]'
                    $pendingDirectories.Push($generationRoot)
                    $actualRelativeFiles = New-Object 'System.Collections.Generic.List[string]'
                    while ($pendingDirectories.Count -gt 0) {
                        $currentDirectory = $pendingDirectories.Pop()
                        $currentAttributes = [System.IO.File]::GetAttributes($currentDirectory)
                        if (($currentAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                            $contractFailure = $true
                            throw 'The OpenAPI staging tree contains a reparse directory.'
                        }
                        $relativeDirectory = $currentDirectory.Substring($generationRoot.Length).TrimStart('\').Replace('\', '/')
                        if (
                            $relativeDirectory.Length -gt 0 -and
                            $relativeDirectory -cne 'backend' -and
                            $relativeDirectory -cne 'backend/docs'
                        ) {
                            $contractFailure = $true
                            throw 'The OpenAPI staging tree contains a residual directory.'
                        }
                        foreach ($childDirectory in [System.IO.Directory]::GetDirectories($currentDirectory)) {
                            $childAttributes = [System.IO.File]::GetAttributes($childDirectory)
                            if (($childAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                                $contractFailure = $true
                                throw 'The OpenAPI staging tree contains a reparse descendant.'
                            }
                            $pendingDirectories.Push($childDirectory)
                        }
                        foreach ($childFile in [System.IO.Directory]::GetFiles($currentDirectory)) {
                            $childAttributes = [System.IO.File]::GetAttributes($childFile)
                            if (($childAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                                $contractFailure = $true
                                throw 'The OpenAPI staging tree contains a reparse file.'
                            }
                            $actualRelativeFiles.Add($childFile.Substring($generationRoot.Length).TrimStart('\').Replace('\', '/'))
                        }
                    }
                    $orderedActualFiles = @($actualRelativeFiles.ToArray() | Sort-Object)
                    if ($orderedActualFiles.Count -ne 3) {
                        $contractFailure = $true
                        throw 'The OpenAPI staging tree does not contain exactly three files.'
                    }
                    foreach ($allowedOpenApiPath in $allowedOpenApiPaths) {
                        if ($orderedActualFiles -cnotcontains $allowedOpenApiPath) {
                            $contractFailure = $true
                            throw 'The OpenAPI staging tree has a missing or residual file.'
                        }
                    }

                    $manifest = @($generationManifests[$generation - 1])
                    if ($manifest.Count -ne 3) {
                        $contractFailure = $true
                        throw 'The OpenAPI generation manifest does not have exactly three rows.'
                    }
                    $revisionRows.Add(('generation' + [char]9 + $generation))
                    for ($pathIndex = 0; $pathIndex -lt $allowedOpenApiPaths.Count; $pathIndex++) {
                        $manifestRow = $manifest[$pathIndex]
                        $manifestFields = @($manifestRow.PSObject.Properties.Name)
                        if (
                            $manifestFields.Count -ne 2 -or
                            $manifestFields -cnotcontains 'path' -or
                            $manifestFields -cnotcontains 'sha256' -or
                            [string]$manifestRow.path -cne $allowedOpenApiPaths[$pathIndex] -or
                            [string]$manifestRow.sha256 -cnotmatch '^[a-f0-9]{64}$'
                        ) {
                            $contractFailure = $true
                            throw 'The OpenAPI generation manifest row is malformed.'
                        }
                        $stagedSnapshot = & $ReadStableContainedFileBytes $generationRoot ([string]$manifestRow.path).Replace('/', '\') 8388608 $false
                        $actualSha = [string]$stagedSnapshot.Sha256
                        if ($actualSha -cne [string]$manifestRow.sha256) {
                            $contractFailure = $true
                            throw 'The OpenAPI staging hash does not match its manifest.'
                        }
                        if ($generation -eq 2 -and [string]$generationManifests[0][$pathIndex].sha256 -cne $actualSha) {
                            $contractFailure = $true
                            throw 'The two OpenAPI staging manifests are not deterministic.'
                        }
                        $revisionRows.Add(([string]$manifestRow.path + [char]9 + $actualSha))
                    }
                }
                $revisionBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(($revisionRows -join [Environment]::NewLine) + [Environment]::NewLine)
                $revisionAlgorithm = [System.Security.Cryptography.SHA256]::Create()
                try {
                    $artifactRevision = [System.BitConverter]::ToString($revisionAlgorithm.ComputeHash($revisionBytes)).Replace('-', '').ToLowerInvariant()
                }
                finally {
                    $revisionAlgorithm.Dispose()
                }
                if ([string]$targetResult.artifact_revision -cne $artifactRevision) {
                    $contractFailure = $true
                    throw 'The OpenAPI artifact revision does not bind the source revision and both manifests.'
                }
                if ($firstExitCode -eq 0 -and -not $contractFailure) {
                    for ($pathIndex = 0; $pathIndex -lt $allowedOpenApiPaths.Count; $pathIndex++) {
                        $relativeOpenApiPath = $allowedOpenApiPaths[$pathIndex].Replace('/', '\')
                        $publicationSource = & $ReadStableContainedFileBytes $generationOneRoot $relativeOpenApiPath 8388608 $false
                        if ($publicationSource.Sha256 -cne [string]$generationManifests[0][$pathIndex].sha256) {
                            $contractFailure = $true
                            throw 'The OpenAPI publication source changed after manifest validation.'
                        }
                        $publicationDestination = & $ReadStableContainedFileBytes $repositoryRoot $relativeOpenApiPath 8388608 $false
                        $bytesEqual = ($publicationSource.Bytes.Length -eq $publicationDestination.Bytes.Length)
                        if ($bytesEqual) {
                            for ($byteIndex = 0; $byteIndex -lt $publicationSource.Bytes.Length; $byteIndex++) {
                                if ($publicationSource.Bytes[$byteIndex] -ne $publicationDestination.Bytes[$byteIndex]) {
                                    $bytesEqual = $false
                                    break
                                }
                            }
                        }
                        if (-not $bytesEqual) {
                            $destinationPath = & $ResolveCanonicalContainedPath $repositoryRoot $relativeOpenApiPath $true $false
                            & $WriteAtomicBytes $destinationPath ([byte[]]$publicationSource.Bytes)
                            [void]$openApiChangedFiles.Add($allowedOpenApiPaths[$pathIndex])
                        }
                    }
                }
            }
        }
    }
    catch {
        if ($firstExitCode -eq 0 -and -not $contractFailure) {
            $contractFailure = $true
        }
    }
    finally {
        if ($reportOwned) {
            $cleanupStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
            $cleanupDeadlineTick = [System.Diagnostics.Stopwatch]::GetTimestamp() + [int64]($cleanupTimeoutMilliseconds * [System.Diagnostics.Stopwatch]::Frequency / 1000)
            try {
                if ($null -eq $dependencies) {
                    $cleanupFailure = $true
                    $cleanupRecord.error_code = 'dependency-factory-failed'
                    throw 'The closed cleanup dependency factory is unavailable.'
                }
                $inventory = @(& $dependencies.ListOwnedResources $cleanupDeadlineTick)
                $initialInventoryClean = $true
                $candidateByLogicalIdentity = @{}
                $seenInventoryLogicalIdentities = @{}
                $expectedNetwork = ($project + '_default')
                foreach ($resource in $inventory) {
                    $isTypedRecord = $false
                    if ($null -ne $resource) {
                        $fields = @($resource.PSObject.Properties.Name)
                        $isTypedRecord = (
                            $fields.Count -eq 5 -and
                            $fields -ccontains 'RecordType' -and
                            $fields -ccontains 'Id' -and
                            $fields -ccontains 'Kind' -and
                            $fields -ccontains 'Name' -and
                            $fields -ccontains 'Project' -and
                            [string]$resource.RecordType -ceq 'hub.qa.resource/v1' -and
                            -not [string]::IsNullOrWhiteSpace([string]$resource.Id)
                        )
                    }
                    if (-not $isTypedRecord) {
                        $cleanupFailure = $true
                        $initialInventoryClean = $false
                        continue
                    }
                    $logicalIdentity = ([string]$resource.Kind + '|' + [string]$resource.Name + '|' + [string]$resource.Project)
                    if ($seenInventoryLogicalIdentities.ContainsKey($logicalIdentity)) {
                        $cleanupFailure = $true
                        $initialInventoryClean = $false
                        [void]$candidateByLogicalIdentity.Remove(([string]$resource.Kind + '|' + [string]$resource.Name))
                        continue
                    }
                    $seenInventoryLogicalIdentities[$logicalIdentity] = $true
                    if ([string]$resource.Kind -ceq 'volume') {
                        $cleanupFailure = $true
                        $initialInventoryClean = $false
                        if ([string]$resource.Project -cne $project) {
                            $cleanupRecord.foreign_count++
                        }
                        continue
                    }
                    if ([string]$resource.Project -cne $project) {
                        $cleanupRecord.foreign_count++
                        continue
                    }
                    $ownedLogicalIdentity = ([string]$resource.Kind + '|' + [string]$resource.Name)
                    $isExpectedOwned = (
                        ([string]$resource.Kind -ceq 'container' -and [string]$resource.Name -ceq $container) -or
                        ([string]$resource.Kind -ceq 'network' -and [string]$resource.Name -ceq $expectedNetwork)
                    )
                    if (-not $isExpectedOwned) {
                        $cleanupFailure = $true
                        $initialInventoryClean = $false
                        continue
                    }
                    $candidateByLogicalIdentity[$ownedLogicalIdentity] = [pscustomobject][ordered]@{
                        RecordType = 'hub.qa.owned-resource/v1'
                        Id = [string]$resource.Id
                        Kind = [string]$resource.Kind
                        Name = [string]$resource.Name
                        Project = [string]$resource.Project
                    }
                }

                if ($initialInventoryClean) {
                    $ownedResources = New-Object 'System.Collections.Generic.List[object]'
                    foreach ($expectedLogicalIdentity in @(
                        ('container|' + $container),
                        ('network|' + $expectedNetwork)
                    )) {
                        if ($candidateByLogicalIdentity.ContainsKey($expectedLogicalIdentity)) {
                            [void]$ownedResources.Add($candidateByLogicalIdentity[$expectedLogicalIdentity])
                        }
                    }
                    $cleanupRecord.selected_owned_count = $ownedResources.Count
                    $cleanupOutcome = & $dependencies.CleanupOwnedResources ($ownedResources.ToArray()) $cleanupDeadlineTick
                    & $AssertRequiredProperty $cleanupOutcome 'Succeeded'
                    if (-not [bool]$cleanupOutcome.Succeeded) {
                        $cleanupFailure = $true
                        $cleanupRecord.error_code = 'cleanup-operation-failed'
                    }
                }
                else {
                    $cleanupFailure = $true
                    $cleanupRecord.selected_owned_count = 0
                    $cleanupRecord.error_code = 'invalid-initial-inventory'
                }

                $residualInventory = @(& $dependencies.ListOwnedResources $cleanupDeadlineTick)
                $residualSameProjectCount = 0
                foreach ($resource in $residualInventory) {
                    $isTypedResidual = $false
                    if ($null -ne $resource) {
                        $fields = @($resource.PSObject.Properties.Name)
                        $isTypedResidual = (
                            $fields.Count -eq 5 -and
                            $fields -ccontains 'RecordType' -and
                            $fields -ccontains 'Id' -and
                            $fields -ccontains 'Kind' -and
                            $fields -ccontains 'Name' -and
                            $fields -ccontains 'Project' -and
                            [string]$resource.RecordType -ceq 'hub.qa.resource/v1'
                        )
                    }
                    if (-not $isTypedResidual) {
                        $cleanupFailure = $true
                        continue
                    }
                    if ([string]$resource.Project -ceq $project) {
                        $residualSameProjectCount++
                    }
                }
                $cleanupRecord.residual_owned_count = $residualSameProjectCount
                if ($residualSameProjectCount -ne 0) {
                    $cleanupFailure = $true
                    if ($null -eq $cleanupRecord.error_code) {
                        $cleanupRecord.error_code = 'residual-same-project-resource'
                    }
                }
                $cleanupRecord.status = if ($cleanupFailure) { 'failed' } else { 'succeeded' }
            }
            catch {
                $cleanupFailure = $true
                $cleanupRecord.status = 'failed'
                if ($null -eq $cleanupRecord.error_code) {
                    $cleanupRecord.error_code = 'bounded-cleanup-route-failed'
                }
            }
            finally {
                $cleanupStopwatch.Stop()
                $cleanupRecord.duration_ms = $cleanupStopwatch.ElapsedMilliseconds
                try {
                    $cleanupBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(($cleanupRecord | ConvertTo-Json -Depth 8) + [Environment]::NewLine)
                    & $WriteAtomicBytes (Join-Path $reportPath 'cleanup.json') $cleanupBytes
                    [void]$persistedArtifacts.Add('cleanup.json')
                }
                catch {
                    $reportFailure = $true
                }
            }
        }
    }

    $runStopwatch.Stop()
    $provisionalExitCode = if ($firstExitCode -ne 0) {
        $firstExitCode
    }
    elseif ($contractFailure) {
        65
    }
    elseif ($reportFailure) {
        70
    }
    elseif ($cleanupFailure) {
        71
    }
    else {
        0
    }

    if ($reportOwned) {
        $runArtifacts = New-Object 'System.Collections.Generic.List[string]'
        foreach ($persistedArtifact in $persistedArtifacts) {
            [void]$runArtifacts.Add($persistedArtifact)
        }
        [void]$runArtifacts.Add('run.json')
        $runRecord = [ordered]@{
            schema = 'hub.qa-run/v1'
            contract = $contractLiteral
            source_sha = $sourceSha
            images_lock_sha256 = $imagesLockSha256
            compose_sha256 = $composeSha256
            target = $Request.Target
            task = $Request.Task
            run = $Request.Run
            project = $project
            container = $container
            image = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['ImageReference']) { [string]$imageAttestation.ImageReference } else { [string]$images.References[$Request.Target] }
            image_digest = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['Digest']) { [string]$imageAttestation.Digest } else { $null }
            entrypoint = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['Entrypoint']) { [string]$imageAttestation.Entrypoint } else { $null }
            entrypoint_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['EntrypointSha256']) { [string]$imageAttestation.EntrypointSha256 } else { $null }
            config_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['ConfigSha256']) { [string]$imageAttestation.ConfigSha256 } else { $null }
            tool_attestation_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['ToolAttestationSha256']) { [string]$imageAttestation.ToolAttestationSha256 } else { $null }
            tool_attestation_verified = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['ToolAttestationVerified']) { [bool]$imageAttestation.ToolAttestationVerified } else { $false }
            deadlines_seconds = [ordered]@{
                inspect = [int]($inspectTimeoutMilliseconds / 1000)
                command = $commandDeadlineSeconds
                target = $targetDeadlineSeconds
                outer = [int]($outerTimeoutMilliseconds / 1000)
                cleanup = [int]($cleanupTimeoutMilliseconds / 1000)
            }
            limits = [ordered]@{ cpus = '2.0'; memory = '2g'; pids = 256 }
            commands = @($expectedCommands[$Request.Target])
            started_at = $startedAt.ToString('o')
            finished_at = [DateTime]::UtcNow.ToString('o')
            duration_ms = $runStopwatch.ElapsedMilliseconds
            process_start_deadline_semantics = 'A synchronous Process.Start is not a hard cutoff; after first observed owner-deadline expiry, only exact-handle best-effort termination and disposal remain, with immutable exit 124 for that layer.'
            deadline_layer = $deadlineLayer
            first_exit_code = $firstExitCode
            final_exit_code = $provisionalExitCode
            collected = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['collected']) { $targetResult.collected } else { $null }
            passed = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['passed']) { $targetResult.passed } else { $null }
            failed = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['failed']) { $targetResult.failed } else { $null }
            skipped = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['skipped']) { $targetResult.skipped } else { $null }
            unexpected_skip_count = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['unexpected_skip_count']) { $targetResult.unexpected_skip_count } else { $null }
            discovery_root = if ($Request.Target -ceq 'backend-unit') { 'backend/tests/' } elseif ($Request.Target -ceq 'frontend-unit') { 'frontend/tests/' } else { 'N/A' }
            discovered_files = if ($Request.Target -ceq 'backend-openapi') { 'N/A' } elseif ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['discovered_files']) { @($targetResult.discovered_files) } else { @() }
            generated_count = if ($Request.Target -ceq 'backend-openapi') { 3 } else { 'N/A' }
            generated_files = if ($Request.Target -ceq 'backend-openapi') { @($allowedOpenApiPaths) } else { 'N/A' }
            changed_files = if ($Request.Target -ceq 'backend-openapi') { @($openApiChangedFiles.ToArray()) } else { 'N/A' }
            changed_count = if ($Request.Target -ceq 'backend-openapi') { $openApiChangedFiles.Count } else { 'N/A' }
            artifact_revision = if ($null -ne $targetResult -and $null -ne $targetResult.PSObject.Properties['artifact_revision']) { [string]$targetResult.artifact_revision } else { $null }
            status = if ($provisionalExitCode -eq 0) { 'completed-zero' } else { 'completed-nonzero' }
            artifacts = @($runArtifacts.ToArray())
            cleanup = $cleanupRecord
        }
        try {
            $runBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(($runRecord | ConvertTo-Json -Depth 20) + [Environment]::NewLine)
            & $WriteAtomicBytes (Join-Path $reportPath 'run.json') $runBytes
        }
        catch {
            $reportFailure = $true
        }
    }

    if ($firstExitCode -ne 0) {
        return [int]$firstExitCode
    }
    if ($contractFailure) {
        return 65
    }
    if ($reportFailure) {
        return 70
    }
    if ($cleanupFailure) {
        return 71
    }
    return 0
}

$privateRequest = [HubQaSupervisorRequest]::new($Contract, $Target, $Task, $Run)
$privateExitCode = Invoke-HubQaSupervisor -Request $privateRequest
exit $privateExitCode
