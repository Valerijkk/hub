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
    $sha256Pattern = '^sha256:[a-f0-9]{64}$'
    $commitPattern = '^[a-f0-9]{40}$'
    $baseRepoDigestPattern = '^[^@ \t\r\n]+@sha256:[a-f0-9]{64}$'
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
    $rawProcessArtifactPaths = [ordered]@{
        compose_stdout = 'raw/compose.stdout.log'
        compose_stderr = 'raw/compose.stderr.log'
    }
    $persistedArtifacts = New-Object 'System.Collections.Generic.List[string]'
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
        ComposePath = $null
        ReportPath = $null
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

        # The v3 lock parser is deliberately local and duplicate-aware. ConvertFrom-Json
        # cannot be the authority because Windows PowerShell 5.1 silently coalesces keys.
        if (
            $null -eq $Snapshot -or
            $null -eq $Snapshot.PSObject.Properties['Bytes'] -or
            $null -eq $Snapshot.PSObject.Properties['Sha256']
        ) {
            throw 'The closed image lock snapshot is absent.'
        }
        [byte[]]$lockBytes = $Snapshot.Bytes
        if ($lockBytes.Length -eq 0 -or ($lockBytes.Length -ge 3 -and $lockBytes[0] -eq 0xef -and $lockBytes[1] -eq 0xbb -and $lockBytes[2] -eq 0xbf)) {
            throw 'The closed image lock must be nonempty no-BOM UTF-8.'
        }
        $strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
        try {
            $lockText = $strictUtf8.GetString($lockBytes)
        }
        catch {
            throw 'The closed image lock is not strict UTF-8.'
        }

        $cursor = 0
        $length = $lockText.Length
        $skipWhitespace = {
            # RFC 8259 permits precisely SP, HTAB, LF, and CR between tokens.
            while ($cursor -lt $length) {
                $code = [int][char]$lockText[$cursor]
                if ($code -ne 0x20 -and $code -ne 0x09 -and $code -ne 0x0a -and $code -ne 0x0d) {
                    break
                }
                $cursor++
            }
        }
        $parseString = $null
        $parseValue = $null
        $parseArray = $null
        $parseObject = $null
        $parseString = {
            if ($cursor -ge $length -or $lockText[$cursor] -ne '"') {
                throw 'The closed image lock has an invalid JSON string.'
            }
            $cursor++
            $builder = New-Object System.Text.StringBuilder
            while ($cursor -lt $length) {
                $character = [int][char]$lockText[$cursor]
                $cursor++
                if ($character -eq 0x22) {
                    return $builder.ToString()
                }
                if ($character -lt 0x20) {
                    throw 'The closed image lock string contains a control character.'
                }
                if ($character -eq 0x5c) {
                    if ($cursor -ge $length) {
                        throw 'The closed image lock string has a trailing escape.'
                    }
                    $escape = $lockText[$cursor]
                    $cursor++
                    switch ($escape) {
                        '"' { [void]$builder.Append('"'); continue }
                        '\' { [void]$builder.Append('\'); continue }
                        '/' { [void]$builder.Append('/'); continue }
                        'b' { [void]$builder.Append([char]0x08); continue }
                        'f' { [void]$builder.Append([char]0x0c); continue }
                        'n' { [void]$builder.Append([char]0x0a); continue }
                        'r' { [void]$builder.Append([char]0x0d); continue }
                        't' { [void]$builder.Append([char]0x09); continue }
                        'u' {
                            if (($cursor + 4) -gt $length) {
                                throw 'The closed image lock string has a truncated unicode escape.'
                            }
                            $hex = $lockText.Substring($cursor, 4)
                            if ($hex -cnotmatch '^[0-9a-fA-F]{4}$') {
                                throw 'The closed image lock string has an invalid unicode escape.'
                            }
                            $codeUnit = [Convert]::ToInt32($hex, 16)
                            $cursor += 4
                            if ($codeUnit -ge 0xd800 -and $codeUnit -le 0xdbff) {
                                if (($cursor + 6) -gt $length -or $lockText[$cursor] -ne '\' -or $lockText[$cursor + 1] -ne 'u') {
                                    throw 'The closed image lock string has an unpaired high surrogate.'
                                }
                                $lowHex = $lockText.Substring($cursor + 2, 4)
                                if ($lowHex -cnotmatch '^[0-9a-fA-F]{4}$') {
                                    throw 'The closed image lock string has an invalid low surrogate escape.'
                                }
                                $lowUnit = [Convert]::ToInt32($lowHex, 16)
                                if ($lowUnit -lt 0xdc00 -or $lowUnit -gt 0xdfff) {
                                    throw 'The closed image lock string has an unpaired high surrogate.'
                                }
                                [void]$builder.Append([char]$codeUnit)
                                [void]$builder.Append([char]$lowUnit)
                                $cursor += 6
                                continue
                            }
                            if ($codeUnit -ge 0xdc00 -and $codeUnit -le 0xdfff) {
                                throw 'The closed image lock string has an unpaired low surrogate.'
                            }
                            [void]$builder.Append([char]$codeUnit)
                            continue
                        }
                        default { throw 'The closed image lock string has an invalid escape.' }
                    }
                }
                if ($character -ge 0xd800 -and $character -le 0xdbff) {
                    if ($cursor -ge $length) {
                        throw 'The closed image lock string has an unpaired high surrogate.'
                    }
                    $lowCharacter = [int][char]$lockText[$cursor]
                    if ($lowCharacter -lt 0xdc00 -or $lowCharacter -gt 0xdfff) {
                        throw 'The closed image lock string has an unpaired high surrogate.'
                    }
                    [void]$builder.Append([char]$character)
                    [void]$builder.Append([char]$lowCharacter)
                    $cursor++
                    continue
                }
                if ($character -ge 0xdc00 -and $character -le 0xdfff) {
                    throw 'The closed image lock string has an unpaired low surrogate.'
                }
                [void]$builder.Append([char]$character)
            }
            throw 'The closed image lock string is unterminated.'
        }
        $parseArray = {
            if ($cursor -ge $length -or $lockText[$cursor] -ne '[') {
                throw 'The closed image lock has an invalid JSON array.'
            }
            $cursor++
            . $skipWhitespace
            $items = New-Object 'System.Collections.Generic.List[object]'
            if ($cursor -lt $length -and $lockText[$cursor] -eq ']') {
                $cursor++
                return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
            }
            while ($true) {
                [void]$items.Add((. $parseValue))
                . $skipWhitespace
                if ($cursor -ge $length) {
                    throw 'The closed image lock array is unterminated.'
                }
                if ($lockText[$cursor] -eq ']') {
                    $cursor++
                    break
                }
                if ($lockText[$cursor] -ne ',') {
                    throw 'The closed image lock array has an invalid separator.'
                }
                $cursor++
                . $skipWhitespace
            }
            return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
        }
        $parseObject = {
            if ($cursor -ge $length -or $lockText[$cursor] -ne '{') {
                throw 'The closed image lock has an invalid JSON object.'
            }
            $cursor++
            . $skipWhitespace
            $members = [ordered]@{}
            if ($cursor -lt $length -and $lockText[$cursor] -eq '}') {
                $cursor++
                return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
            }
            while ($true) {
                $name = . $parseString
                if ($members.Contains($name)) {
                    throw 'The closed image lock contains a duplicate object member.'
                }
                . $skipWhitespace
                if ($cursor -ge $length -or $lockText[$cursor] -ne ':') {
                    throw 'The closed image lock object has no member separator.'
                }
                $cursor++
                . $skipWhitespace
                $members[$name] = . $parseValue
                . $skipWhitespace
                if ($cursor -ge $length) {
                    throw 'The closed image lock object is unterminated.'
                }
                if ($lockText[$cursor] -eq '}') {
                    $cursor++
                    break
                }
                if ($lockText[$cursor] -ne ',') {
                    throw 'The closed image lock object has an invalid separator.'
                }
                $cursor++
                . $skipWhitespace
                if ($cursor -ge $length -or $lockText[$cursor] -ne '"') {
                    throw 'The closed image lock object has an invalid member name.'
                }
            }
            return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
        }
        $parseValue = {
            . $skipWhitespace
            if ($cursor -ge $length) {
                throw 'The closed image lock has an incomplete JSON value.'
            }
            switch ($lockText[$cursor]) {
                '{' { return (. $parseObject) }
                '[' { return (. $parseArray) }
                '"' { return [pscustomobject][ordered]@{ Type = 'string'; Value = (. $parseString) } }
                'n' {
                    if (($cursor + 4) -le $length -and $lockText.Substring($cursor, 4) -ceq 'null') {
                        $cursor += 4
                        return [pscustomobject][ordered]@{ Type = 'null'; Value = $null }
                    }
                    throw 'The closed image lock has an invalid literal.'
                }
                default { throw 'The closed image lock permits no JSON numbers or booleans.' }
            }
        }

        $root = . $parseValue
        . $skipWhitespace
        if ($cursor -ne $length -or [string]$root.Type -cne 'object') {
            throw 'The closed image lock has trailing data or no top-level object.'
        }

        $assertExactObject = {
            param([object]$Node, [string[]]$Expected, [string]$Identity, [bool]$RequireOrderedMembers)

            if ($null -eq $Node -or [string]$Node.Type -cne 'object') {
                throw ('The closed image lock ' + $Identity + ' is not an object.')
            }
            $actual = [System.Collections.IDictionary]$Node.Value
            if ($actual.Count -ne $Expected.Count) {
                throw ('The closed image lock ' + $Identity + ' has an invalid member count.')
            }
            foreach ($name in $Expected) {
                if (-not $actual.Contains($name)) {
                    throw ('The closed image lock ' + $Identity + ' is missing a required member.')
                }
            }
            if ($RequireOrderedMembers) {
                $actualNames = @($actual.Keys)
                for ($index = 0; $index -lt $Expected.Count; $index++) {
                    if ([string]$actualNames[$index] -cne [string]$Expected[$index]) {
                        throw ('The closed image lock ' + $Identity + ' has an invalid member order.')
                    }
                }
            }
            return $actual
        }.GetNewClosure()
        $requireString = {
            param([object]$Node, [string]$Identity, [string]$Pattern)

            if ($null -eq $Node -or [string]$Node.Type -cne 'string' -or [string]$Node.Value -cnotmatch $Pattern) {
                throw ('The closed image lock ' + $Identity + ' has an invalid string scalar.')
            }
            return [string]$Node.Value
        }.GetNewClosure()
        $jcsString = {
            param([string]$Value)

            if ($null -eq $Value) {
                throw 'A JCS string cannot be null.'
            }
            $builder = New-Object System.Text.StringBuilder
            [void]$builder.Append('"')
            for ($index = 0; $index -lt $Value.Length; $index++) {
                $code = [int][char]$Value[$index]
                if ($code -eq 0x22) { [void]$builder.Append('\"'); continue }
                if ($code -eq 0x5c) { [void]$builder.Append('\\'); continue }
                if ($code -eq 0x08) { [void]$builder.Append('\b'); continue }
                if ($code -eq 0x09) { [void]$builder.Append('\t'); continue }
                if ($code -eq 0x0a) { [void]$builder.Append('\n'); continue }
                if ($code -eq 0x0c) { [void]$builder.Append('\f'); continue }
                if ($code -eq 0x0d) { [void]$builder.Append('\r'); continue }
                if ($code -lt 0x20) { [void]$builder.Append(('\u{0:x4}' -f $code)); continue }
                if ($code -ge 0xd800 -and $code -le 0xdbff) {
                    if (($index + 1) -ge $Value.Length) { throw 'A JCS string has an unpaired high surrogate.' }
                    $low = [int][char]$Value[$index + 1]
                    if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw 'A JCS string has an unpaired high surrogate.' }
                    [void]$builder.Append([char]$code)
                    [void]$builder.Append([char]$low)
                    $index++
                    continue
                }
                if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw 'A JCS string has an unpaired low surrogate.' }
                [void]$builder.Append([char]$code)
            }
            [void]$builder.Append('"')
            return $builder.ToString()
        }.GetNewClosure()
        $getJcsSha256 = {
            param([string]$CanonicalJson)

            $algorithm = [System.Security.Cryptography.SHA256]::Create()
            try {
                return 'sha256:' + [System.BitConverter]::ToString($algorithm.ComputeHash([System.Text.UTF8Encoding]::new($false).GetBytes($CanonicalJson))).Replace('-', '').ToLowerInvariant()
            }
            finally {
                $algorithm.Dispose()
            }
        }.GetNewClosure()
        $readStringArray = {
            param([object]$Node, [string]$Identity, [bool]$RequireNonempty)

            if ($null -eq $Node -or [string]$Node.Type -cne 'array') {
                throw ('The closed image lock ' + $Identity + ' is not an array.')
            }
            [object[]]$items = $Node.Value
            if ($RequireNonempty -and $items.Count -eq 0) {
                throw ('The closed image lock ' + $Identity + ' is empty.')
            }
            $result = New-Object 'System.Collections.Generic.List[string]'
            foreach ($item in $items) {
                if ($null -eq $item -or [string]$item.Type -cne 'string') {
                    throw ('The closed image lock ' + $Identity + ' contains a non-string member.')
                }
                [void]$result.Add([string]$item.Value)
            }
            return [string[]]$result.ToArray()
        }.GetNewClosure()
        $readLabels = {
            param([object]$Node, [string]$Identity)

            if ($null -eq $Node -or [string]$Node.Type -cne 'object') {
                throw ('The closed image lock ' + $Identity + ' is not an object.')
            }
            if (@($Node.Value.Keys).Count -eq 0) {
                throw ('The closed image lock ' + $Identity + ' is empty.')
            }
            $labels = [ordered]@{}
            foreach ($name in @($Node.Value.Keys)) {
                $value = $Node.Value[$name]
                if ([string]::IsNullOrEmpty([string]$name) -or $null -eq $value -or [string]$value.Type -cne 'string') {
                    throw ('The closed image lock ' + $Identity + ' is not a string-to-string map.')
                }
                $labels[[string]$name] = [string]$value.Value
            }
            return $labels
        }.GetNewClosure()
        $jcsArray = {
            param([string[]]$Values)

            $encoded = New-Object 'System.Collections.Generic.List[string]'
            foreach ($value in $Values) {
                [void]$encoded.Add((& $jcsString $value))
            }
            return '[' + ($encoded -join ',') + ']'
        }.GetNewClosure()
        $jcsObject = {
            param([System.Collections.IDictionary]$Values)

            $names = [string[]]@($Values.Keys)
            [System.Array]::Sort($names, [System.StringComparer]::Ordinal)
            $encoded = New-Object 'System.Collections.Generic.List[string]'
            foreach ($name in $names) {
                [void]$encoded.Add(((& $jcsString $name) + ':' + (& $jcsString ([string]$Values[$name]))))
            }
            return '{' + ($encoded -join ',') + '}'
        }.GetNewClosure()
        $readPlatform = {
            param([object]$Node, [string]$Identity)

            $platform = & $assertExactObject $Node @('os', 'architecture', 'variant') $Identity $false
            if (
                [string]$platform['os'].Type -cne 'string' -or [string]$platform['os'].Value -cne 'linux' -or
                [string]$platform['architecture'].Type -cne 'string' -or [string]$platform['architecture'].Value -cne 'amd64' -or
                [string]$platform['variant'].Type -cne 'null'
            ) {
                throw ('The closed image lock ' + $Identity + ' is not linux/amd64/null.')
            }
            return [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }
        }.GetNewClosure()
        $readRecord = {
            param([object]$Node, [string]$Identity, [bool]$IsTarget)

            $expected = if ($IsTarget) {
                @('base_key', 'image_id', 'platform', 'rootfs_layers', 'entrypoint', 'entrypoint_sha256', 'cmd', 'cmd_sha256', 'control_commit', 'control_bundle_sha256', 'source_commit', 'source_policy_sha256', 'source_tree_manifest_sha256', 'build_manifest_sha256', 'labels_projection', 'labels_projection_sha256')
            }
            else {
                @('repo_digest', 'image_id', 'platform', 'rootfs_layers', 'entrypoint', 'entrypoint_sha256', 'cmd', 'cmd_sha256', 'labels_projection', 'labels_projection_sha256')
            }
            $record = & $assertExactObject $Node $expected $Identity $IsTarget
            $entrypoint = & $readStringArray $record['entrypoint'] ($Identity + '.entrypoint') $IsTarget
            $cmd = & $readStringArray $record['cmd'] ($Identity + '.cmd') $false
            $labels = & $readLabels $record['labels_projection'] ($Identity + '.labels_projection')
            $entrypointSha = & $requireString $record['entrypoint_sha256'] ($Identity + '.entrypoint_sha256') $sha256Pattern
            $cmdSha = & $requireString $record['cmd_sha256'] ($Identity + '.cmd_sha256') $sha256Pattern
            $labelsSha = & $requireString $record['labels_projection_sha256'] ($Identity + '.labels_projection_sha256') $sha256Pattern
            if (
                $entrypointSha -cne (& $getJcsSha256 (& $jcsArray $entrypoint)) -or
                $cmdSha -cne (& $getJcsSha256 (& $jcsArray $cmd)) -or
                $labelsSha -cne (& $getJcsSha256 (& $jcsObject $labels))
            ) {
                throw ('The closed image lock ' + $Identity + ' JCS hash does not bind its complete projection.')
            }
            $validated = [ordered]@{
                image_id = (& $requireString $record['image_id'] ($Identity + '.image_id') $sha256Pattern)
                platform = (& $readPlatform $record['platform'] ($Identity + '.platform'))
                rootfs_layers = (& $readStringArray $record['rootfs_layers'] ($Identity + '.rootfs_layers') $true)
                entrypoint = $entrypoint
                entrypoint_sha256 = $entrypointSha
                cmd = $cmd
                cmd_sha256 = $cmdSha
                labels_projection = $labels
                labels_projection_sha256 = $labelsSha
            }
            foreach ($layer in @($validated['rootfs_layers'])) {
                if ($layer -cnotmatch $sha256Pattern) {
                    throw ('The closed image lock ' + $Identity + ' has an invalid RootFS layer.')
                }
            }
            if ($IsTarget) {
                $validated['base_key'] = & $requireString $record['base_key'] ($Identity + '.base_key') '^(golang|playwright|python)$'
                $validated['control_commit'] = & $requireString $record['control_commit'] ($Identity + '.control_commit') $commitPattern
                $validated['control_bundle_sha256'] = & $requireString $record['control_bundle_sha256'] ($Identity + '.control_bundle_sha256') $sha256Pattern
                $validated['source_commit'] = & $requireString $record['source_commit'] ($Identity + '.source_commit') $commitPattern
                $validated['source_policy_sha256'] = & $requireString $record['source_policy_sha256'] ($Identity + '.source_policy_sha256') $sha256Pattern
                $validated['source_tree_manifest_sha256'] = & $requireString $record['source_tree_manifest_sha256'] ($Identity + '.source_tree_manifest_sha256') $sha256Pattern
                $validated['build_manifest_sha256'] = & $requireString $record['build_manifest_sha256'] ($Identity + '.build_manifest_sha256') $sha256Pattern
            }
            else {
                $validated['repo_digest'] = & $requireString $record['repo_digest'] ($Identity + '.repo_digest') $baseRepoDigestPattern
            }
            return $validated
        }.GetNewClosure()

        $top = & $assertExactObject $root @('schema', 'control', 'source', 'build', 'bases', 'targets') 'top-level' $true
        $schemaLiteral = & $requireString $top['schema'] 'schema' '^hub\.qa\.local-images-lock/v3$'
        if ($schemaLiteral -cne 'hub.qa.local-images-lock/v3') {
            throw 'The closed image lock schema is invalid.'
        }
        $control = & $assertExactObject $top['control'] @('commit', 'bundle_sha256', 'compose_sha256') 'control' $false
        $source = & $assertExactObject $top['source'] @('commit', 'policy_sha256', 'tree_manifest_sha256') 'source' $false
        $build = & $assertExactObject $top['build'] @('manifest_sha256') 'build' $false
        $validatedControl = [ordered]@{
            commit = (& $requireString $control['commit'] 'control.commit' $commitPattern)
            bundle_sha256 = (& $requireString $control['bundle_sha256'] 'control.bundle_sha256' $sha256Pattern)
            compose_sha256 = (& $requireString $control['compose_sha256'] 'control.compose_sha256' $sha256Pattern)
        }
        $validatedSource = [ordered]@{
            commit = (& $requireString $source['commit'] 'source.commit' $commitPattern)
            policy_sha256 = (& $requireString $source['policy_sha256'] 'source.policy_sha256' $sha256Pattern)
            tree_manifest_sha256 = (& $requireString $source['tree_manifest_sha256'] 'source.tree_manifest_sha256' $sha256Pattern)
        }
        $validatedBuild = [ordered]@{ manifest_sha256 = (& $requireString $build['manifest_sha256'] 'build.manifest_sha256' $sha256Pattern) }
        $baseObject = & $assertExactObject $top['bases'] @('golang', 'playwright', 'python') 'bases' $false
        $targetObject = & $assertExactObject $top['targets'] $fixedTargets 'targets' $false
        $targetBaseMap = [ordered]@{
            'backend-unit' = 'golang'
            'frontend-unit' = 'playwright'
            'backend-openapi' = 'python'
        }
        $validatedBases = [ordered]@{}
        foreach ($baseKey in @('golang', 'playwright', 'python')) {
            $validatedBases[$baseKey] = & $readRecord $baseObject[$baseKey] ('bases.' + $baseKey) $false
        }
        $validatedTargets = [ordered]@{}
        foreach ($targetKey in $fixedTargets) {
            $targetRecord = & $readRecord $targetObject[$targetKey] ('targets.' + $targetKey) $true
            if (
                [string]$targetRecord['base_key'] -cne [string]$targetBaseMap[$targetKey] -or
                [string]$targetRecord['control_commit'] -cne [string]$validatedControl['commit'] -or
                [string]$targetRecord['control_bundle_sha256'] -cne [string]$validatedControl['bundle_sha256'] -or
                [string]$targetRecord['source_commit'] -cne [string]$validatedSource['commit'] -or
                [string]$targetRecord['source_policy_sha256'] -cne [string]$validatedSource['policy_sha256'] -or
                [string]$targetRecord['source_tree_manifest_sha256'] -cne [string]$validatedSource['tree_manifest_sha256'] -or
                [string]$targetRecord['build_manifest_sha256'] -cne [string]$validatedBuild['manifest_sha256']
            ) {
                throw 'The closed image lock target does not bind top-level provenance and its fixed base.'
            }
            $requiredLabels = [ordered]@{
                'hub.qa.target' = $targetKey
                'hub.qa.base-image-id' = [string]$validatedBases[[string]$targetBaseMap[$targetKey]]['image_id']
                'hub.qa.control-commit' = [string]$validatedControl['commit']
                'hub.qa.control-bundle-sha256' = [string]$validatedControl['bundle_sha256']
                'hub.qa.source-commit' = [string]$validatedSource['commit']
                'hub.qa.source-policy-sha256' = [string]$validatedSource['policy_sha256']
                'hub.qa.source-tree-manifest-sha256' = [string]$validatedSource['tree_manifest_sha256']
                'hub.qa.build-manifest-sha256' = [string]$validatedBuild['manifest_sha256']
                'hub.qa.runtime-status' = 'NOT_READY'
            }
            foreach ($labelName in @($requiredLabels.Keys)) {
                if (-not $targetRecord['labels_projection'].Contains($labelName) -or [string]$targetRecord['labels_projection'][$labelName] -cne [string]$requiredLabels[$labelName]) {
                    throw 'The closed image lock target has an invalid mandatory provenance label.'
                }
            }
            $validatedTargets[$targetKey] = $targetRecord
        }
        return [pscustomobject][ordered]@{
            Control = $validatedControl
            Source = $validatedSource
            Build = $validatedBuild
            Bases = $validatedBases
            Targets = $validatedTargets
            Sha256 = [string]$Snapshot.Sha256
        }
    }.GetNewClosure()

    $AssertPrivateComposeSnapshot = {
        param([object]$Snapshot, [string]$ExpectedControlComposeSha256)

        if (
            $null -eq $Snapshot -or
            $null -eq $Snapshot.PSObject.Properties['Bytes'] -or
            $null -eq $Snapshot.PSObject.Properties['Sha256']
        ) {
            throw 'The immutable Compose snapshot is absent.'
        }
        [byte[]]$composeBytes = $Snapshot.Bytes
        if (
            $composeBytes.Length -eq 0 -or
            $ExpectedControlComposeSha256 -cnotmatch $sha256Pattern -or
            ('sha256:' + [string]$Snapshot.Sha256) -cne $ExpectedControlComposeSha256
        ) {
            throw 'The immutable Compose snapshot does not equal the closed control raw-byte pin.'
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
            [string]$ExactImageId,
            [string]$ClosedComposePath,
            [string]$ClosedProject,
            [string]$ClosedReportRoot,
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
                $arguments = @('image', 'inspect', $ExactImageId)
            }
            'run' {
                if ($null -eq $ClosedComposeEnvironment) {
                    throw 'The fixed Compose environment has not been attested.'
                }
                $arguments = @(
                    'compose', '--ansi', 'never',
                    '-f', $ClosedComposePath,
                    '--project-directory', $ClosedReportRoot,
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
        $startInfo.WorkingDirectory = $ClosedReportRoot
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
        $lockedTarget = [System.Collections.IDictionary]$dependencyContext.Images.Targets[$Request.Target]
        $frozenImageId = [string]$lockedTarget['image_id']
        $frozenReportRoot = [string]$dependencyContext.ReportPath
        $deadlineTick = [System.Diagnostics.Stopwatch]::GetTimestamp() + [int64]($inspectTimeoutMilliseconds * [System.Diagnostics.Stopwatch]::Frequency / 1000)
        $handle = & $StartFixedDockerProcess 'inspect' '' '' $frozenImageId '' '' $frozenReportRoot $null
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
            & $AssertRequiredProperty $document 'Id'
            & $AssertRequiredProperty $document 'Os'
            & $AssertRequiredProperty $document 'Architecture'
            & $AssertRequiredProperty $document 'Variant'
            & $AssertRequiredProperty $document 'RootFS'
            & $AssertRequiredProperty $document.RootFS 'Layers'
            & $AssertRequiredProperty $document 'Config'
            & $AssertRequiredProperty $document.Config 'Entrypoint'
            & $AssertRequiredProperty $document.Config 'Cmd'
            & $AssertRequiredProperty $document.Config 'Labels'
            if (
                [string]$document.Id -cne $frozenImageId -or
                [string]$document.Os -cne 'linux' -or
                [string]$document.Architecture -cne 'amd64' -or
                $null -ne $document.Variant -or
                $null -eq $document.RootFS.Layers -or
                $document.RootFS.Layers -is [string] -or
                $null -eq $document.Config.Entrypoint -or
                $document.Config.Entrypoint -is [string] -or
                $null -eq $document.Config.Cmd -or
                $document.Config.Cmd -is [string] -or
                $null -eq $document.Config.Labels
            ) {
                throw 'The selected local image inspection has an invalid closed shape.'
            }
            [object[]]$rawLayers = @($document.RootFS.Layers)
            [object[]]$rawEntrypoint = @($document.Config.Entrypoint)
            [object[]]$rawCmd = @($document.Config.Cmd)
            if ($rawLayers.Count -eq 0 -or $rawEntrypoint.Count -eq 0) {
                throw 'The selected local image has an empty required array.'
            }
            foreach ($value in @($rawLayers + $rawEntrypoint + $rawCmd)) {
                if ($null -eq $value -or $value -isnot [string]) {
                    throw 'The selected local image has a non-string array member.'
                }
            }
            [string[]]$actualLayers = $rawLayers
            [string[]]$actualEntrypoint = $rawEntrypoint
            [string[]]$actualCmd = $rawCmd
            $actualLabels = [ordered]@{}
            foreach ($property in @($document.Config.Labels.PSObject.Properties)) {
                if ([string]::IsNullOrEmpty([string]$property.Name) -or $null -eq $property.Value -or $property.Value -isnot [string]) {
                    throw 'The selected local image has a non-string label projection.'
                }
                $actualLabels[[string]$property.Name] = [string]$property.Value
            }
            $lockedLabels = [System.Collections.IDictionary]$lockedTarget['labels_projection']
            if ($actualLabels.Count -ne $lockedLabels.Count) {
                throw 'The selected local image labels do not equal the complete locked projection.'
            }
            foreach ($labelName in @($lockedLabels.Keys)) {
                if (-not $actualLabels.Contains($labelName) -or [string]$actualLabels[$labelName] -cne [string]$lockedLabels[$labelName]) {
                    throw 'The selected local image labels do not equal the complete locked projection.'
                }
            }
            $arraysEqual = {
                param([string[]]$Left, [string[]]$Right)

                if ($Left.Count -ne $Right.Count) { return $false }
                for ($arrayIndex = 0; $arrayIndex -lt $Left.Count; $arrayIndex++) {
                    if ([string]$Left[$arrayIndex] -cne [string]$Right[$arrayIndex]) { return $false }
                }
                return $true
            }.GetNewClosure()
            $jcsString = {
                param([string]$Value)

                if ($null -eq $Value) { throw 'A selected image JCS string cannot be null.' }
                $builder = New-Object System.Text.StringBuilder
                [void]$builder.Append('"')
                for ($index = 0; $index -lt $Value.Length; $index++) {
                    $code = [int][char]$Value[$index]
                    if ($code -eq 0x22) { [void]$builder.Append('\"'); continue }
                    if ($code -eq 0x5c) { [void]$builder.Append('\\'); continue }
                    if ($code -eq 0x08) { [void]$builder.Append('\b'); continue }
                    if ($code -eq 0x09) { [void]$builder.Append('\t'); continue }
                    if ($code -eq 0x0a) { [void]$builder.Append('\n'); continue }
                    if ($code -eq 0x0c) { [void]$builder.Append('\f'); continue }
                    if ($code -eq 0x0d) { [void]$builder.Append('\r'); continue }
                    if ($code -lt 0x20) { [void]$builder.Append(('\u{0:x4}' -f $code)); continue }
                    if ($code -ge 0xd800 -and $code -le 0xdbff) {
                        if (($index + 1) -ge $Value.Length) { throw 'A selected image JCS string has an unpaired high surrogate.' }
                        $low = [int][char]$Value[$index + 1]
                        if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw 'A selected image JCS string has an unpaired high surrogate.' }
                        [void]$builder.Append([char]$code)
                        [void]$builder.Append([char]$low)
                        $index++
                        continue
                    }
                    if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw 'A selected image JCS string has an unpaired low surrogate.' }
                    [void]$builder.Append([char]$code)
                }
                [void]$builder.Append('"')
                return $builder.ToString()
            }.GetNewClosure()
            $hashCanonical = {
                param([string]$CanonicalJson)

                $algorithm = [System.Security.Cryptography.SHA256]::Create()
                try {
                    return 'sha256:' + [System.BitConverter]::ToString($algorithm.ComputeHash([System.Text.UTF8Encoding]::new($false).GetBytes($CanonicalJson))).Replace('-', '').ToLowerInvariant()
                }
                finally {
                    $algorithm.Dispose()
                }
            }.GetNewClosure()
            $canonicalArray = {
                param([string[]]$Values)

                $encoded = New-Object 'System.Collections.Generic.List[string]'
                foreach ($value in $Values) { [void]$encoded.Add((& $jcsString $value)) }
                return '[' + ($encoded -join ',') + ']'
            }.GetNewClosure()
            $canonicalLabels = {
                param([System.Collections.IDictionary]$Values)

                $names = [string[]]@($Values.Keys)
                [System.Array]::Sort($names, [System.StringComparer]::Ordinal)
                $encoded = New-Object 'System.Collections.Generic.List[string]'
                foreach ($name in $names) { [void]$encoded.Add(((& $jcsString $name) + ':' + (& $jcsString ([string]$Values[$name])))) }
                return '{' + ($encoded -join ',') + '}'
            }.GetNewClosure()
            $lockedBase = [System.Collections.IDictionary]$dependencyContext.Images.Bases[[string]$lockedTarget['base_key']]
            $requiredLabels = [ordered]@{
                'hub.qa.target' = $Request.Target
                'hub.qa.base-image-id' = [string]$lockedBase['image_id']
                'hub.qa.control-commit' = [string]$dependencyContext.Images.Control['commit']
                'hub.qa.control-bundle-sha256' = [string]$dependencyContext.Images.Control['bundle_sha256']
                'hub.qa.source-commit' = [string]$dependencyContext.Images.Source['commit']
                'hub.qa.source-policy-sha256' = [string]$dependencyContext.Images.Source['policy_sha256']
                'hub.qa.source-tree-manifest-sha256' = [string]$dependencyContext.Images.Source['tree_manifest_sha256']
                'hub.qa.build-manifest-sha256' = [string]$dependencyContext.Images.Build['manifest_sha256']
                'hub.qa.runtime-status' = 'NOT_READY'
            }
            foreach ($requiredLabelName in @($requiredLabels.Keys)) {
                if (-not $actualLabels.Contains($requiredLabelName) -or [string]$actualLabels[$requiredLabelName] -cne [string]$requiredLabels[$requiredLabelName]) {
                    throw 'The selected local image does not bind the required baked provenance label.'
                }
            }
            $verified = (
                (& $arraysEqual $actualLayers ([string[]]$lockedTarget['rootfs_layers'])) -and
                (& $arraysEqual $actualEntrypoint ([string[]]$lockedTarget['entrypoint'])) -and
                (& $arraysEqual $actualCmd ([string[]]$lockedTarget['cmd'])) -and
                ([string]$lockedTarget['entrypoint_sha256'] -ceq (& $hashCanonical (& $canonicalArray $actualEntrypoint))) -and
                ([string]$lockedTarget['cmd_sha256'] -ceq (& $hashCanonical (& $canonicalArray $actualCmd))) -and
                ([string]$lockedTarget['labels_projection_sha256'] -ceq (& $hashCanonical (& $canonicalLabels $actualLabels)))
            )
            return [pscustomobject][ordered]@{
                Verified = [bool]$verified
                ExitCode = 0
                DeadlineLayer = $null
                ImageId = $frozenImageId
                Entrypoint = $actualEntrypoint
                EntrypointSha256 = [string]$lockedTarget['entrypoint_sha256']
                Cmd = $actualCmd
                CmdSha256 = [string]$lockedTarget['cmd_sha256']
                LabelsProjectionSha256 = [string]$lockedTarget['labels_projection_sha256']
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
        $frozenReportRoot = [string]$dependencyContext.ReportPath
        $frozenComposeEnvironment = [System.Collections.IDictionary]$dependencyContext.ComposeEnvironment
        $deadlineTick = [System.Diagnostics.Stopwatch]::GetTimestamp() + [int64]($outerTimeoutMilliseconds * [System.Diagnostics.Stopwatch]::Frequency / 1000)
        $handle = & $StartFixedDockerProcess 'run' '' '' '' $frozenComposePath $frozenProject $frozenReportRoot $frozenComposeEnvironment
        return (& $CompleteFixedDockerProcess $handle $deadlineTick 'outer' 'run')
    }.GetNewClosure()

    $ListOwnedResources = {
        param([int64]$DeadlineTick)

        $frozenProject = [string]$dependencyContext.Project
        $frozenReportRoot = [string]$dependencyContext.ReportPath
        $rows = New-Object 'System.Collections.Generic.List[object]'
        foreach ($resourceKind in @('container', 'network', 'volume')) {
            $handle = & $StartFixedDockerProcess 'list' $resourceKind '' '' '' $frozenProject $frozenReportRoot $null
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
        $frozenReportRoot = [string]$dependencyContext.ReportPath
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
                $handle = & $StartFixedDockerProcess 'cleanup' $resourceKind ([string]$matching[0].Id) '' '' $frozenProject $frozenReportRoot $null
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
        $testframeworkRoot = & $ResolveCanonicalContainedPath $testframeworkRoot '.' $true $true
        $runnerRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'runner' $true $false
        $dockerRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'docker' $true $false
        $lockPath = & $ResolveCanonicalContainedPath $dockerRoot 'images.lock.v3.json' $true $false
        $composePath = & $ResolveCanonicalContainedPath $dockerRoot 'compose.unit.yaml' $true $false
        $reportsRoot = & $ResolveCanonicalContainedPath $testframeworkRoot 'reports' $true $false
        $reportPath = & $ResolveCanonicalContainedPath $reportsRoot ($Request.Task + '-' + $Request.Run) $false $false
        $dependencyContext.ReportPath = $reportPath
        if ([System.IO.Directory]::Exists($reportPath) -or [System.IO.File]::Exists($reportPath)) {
            return 64
        }

        $lockSnapshot = & $ReadStableContainedFileBytes $dockerRoot 'images.lock.v3.json' 131072 $false
        $images = & $ReadValidatedImagesLock $lockSnapshot
        $dependencyContext.Images = $images
        $imagesLockSha256 = [string]$images.Sha256
        $composeSnapshot = & $ReadStableContainedFileBytes $dockerRoot 'compose.unit.yaml' 131072 $false
        $composeSha256 = [string]$composeSnapshot.Sha256
        & $AssertPrivateComposeSnapshot $composeSnapshot ([string]$images.Control['compose_sha256'])
    }
    catch {
        return 64
    }

    $reservationPath = $null
    $reservationBytes = $null
    $reservationSha256 = $null
    $stagingPath = $null
    $stagingOwnerPath = $null
    $ownerBytes = $null
    try {
        $ownerNonce = [System.Guid]::NewGuid().ToString('N')
        $finalRelativePath = $Request.Task + '-' + $Request.Run
        $stagingRelativePath = '.staging-' + $Request.Task + '-' + $Request.Run + '-' + $ownerNonce
        $reservationRelativePath = '.reservation-' + $Request.Task + '-' + $Request.Run + '-' + $ownerNonce + '.json'
        $ownerRecord = [ordered]@{
            schema = 'hub.qa-owner/v1'
            contract = $contractLiteral
            task = $Request.Task
            run = $Request.Run
            project = $project
            target = $Request.Target
            nonce = $ownerNonce
            report_relative_path = ($Request.Task + '-' + $Request.Run)
            control_commit = [string]$images.Control['commit']
            control_bundle_sha256 = [string]$images.Control['bundle_sha256']
            control_compose_sha256 = [string]$images.Control['compose_sha256']
            source_commit = [string]$images.Source['commit']
            source_policy_sha256 = [string]$images.Source['policy_sha256']
            source_tree_manifest_sha256 = [string]$images.Source['tree_manifest_sha256']
            build_manifest_sha256 = [string]$images.Build['manifest_sha256']
            target_image_id = [string]$images.Targets[$Request.Target]['image_id']
            images_lock_sha256 = $imagesLockSha256
            compose_sha256 = $composeSha256
            created_at = [DateTime]::UtcNow.ToString('o')
        }
        $ownerBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(($ownerRecord | ConvertTo-Json -Compress) + [Environment]::NewLine)
        $ownerHasher = [System.Security.Cryptography.SHA256]::Create()
        try {
            $ownerMarkerSha256 = [System.BitConverter]::ToString($ownerHasher.ComputeHash($ownerBytes)).Replace('-', '').ToLowerInvariant()
        }
        finally {
            $ownerHasher.Dispose()
        }
        $reservationRecord = [ordered]@{
            schema = 'hub.qa-report-reservation/v1'
            contract = $contractLiteral
            task = $Request.Task
            run = $Request.Run
            target = $Request.Target
            project = $project
            nonce = $ownerNonce
            staging_relative_path = $stagingRelativePath
            final_relative_path = $finalRelativePath
            owner_marker_sha256 = $ownerMarkerSha256
            compose_sha256 = $composeSha256
        }
        $reservationBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(($reservationRecord | ConvertTo-Json -Compress) + [Environment]::NewLine)
        $reservationHasher = [System.Security.Cryptography.SHA256]::Create()
        try {
            $reservationSha256 = [System.BitConverter]::ToString($reservationHasher.ComputeHash($reservationBytes)).Replace('-', '').ToLowerInvariant()
        }
        finally {
            $reservationHasher.Dispose()
        }
        if ([System.IO.Directory]::Exists($reportPath) -or [System.IO.File]::Exists($reportPath)) {
            return 64
        }
        $reservationPath = & $ResolveCanonicalContainedPath $reportsRoot $reservationRelativePath $false $false
        $stagingPath = & $ResolveCanonicalContainedPath $reportsRoot $stagingRelativePath $false $false
        $reservationStream = [System.IO.FileStream]::new(
            $reservationPath,
            [System.IO.FileMode]::CreateNew,
            [System.IO.FileAccess]::Write,
            [System.IO.FileShare]::None
        )
        try {
            $reservationStream.Write($reservationBytes, 0, $reservationBytes.Length)
            $reservationStream.Flush($true)
        }
        finally {
            $reservationStream.Dispose()
        }
        if ([System.IO.Directory]::Exists($stagingPath) -or [System.IO.File]::Exists($stagingPath)) {
            throw 'The report staging path already exists.'
        }

        # The reports root is a cooperative trusted-parent namespace. This fail-leave
        # protocol rejects every observed collision or ambiguity; it is not a claim
        # to prove ownership against a hostile filesystem-race adversary.
        [System.IO.Directory]::CreateDirectory($stagingPath) | Out-Null
        if (-not [System.IO.Directory]::Exists($stagingPath)) {
            throw 'The fresh staging reservation is not a directory.'
        }
        $stagingPath = & $ResolveCanonicalContainedPath $reportsRoot ([System.IO.Path]::GetFileName($stagingPath)) $true $false
        $stagingAttributes = [System.IO.File]::GetAttributes($stagingPath)
        if (($stagingAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or @([System.IO.Directory]::GetFileSystemEntries($stagingPath)).Count -ne 0) {
            throw 'The fresh staging reservation is not an empty canonical directory.'
        }
        $stagingOwnerPath = & $ResolveCanonicalContainedPath $stagingPath '.owner.json' $false $false
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
        if (@([System.IO.Directory]::GetFileSystemEntries($stagingPath)).Count -ne 1) {
            throw 'The fresh staging reservation has an unexpected initial state.'
        }
        $stagingComposePath = & $ResolveCanonicalContainedPath $stagingPath 'compose.unit.yaml' $false $false
        $composeStream = [System.IO.FileStream]::new(
            $stagingComposePath,
            [System.IO.FileMode]::CreateNew,
            [System.IO.FileAccess]::Write,
            [System.IO.FileShare]::None
        )
        try {
            $composeStream.Write([byte[]]$composeSnapshot.Bytes, 0, ([byte[]]$composeSnapshot.Bytes).Length)
            $composeStream.Flush($true)
        }
        finally {
            $composeStream.Dispose()
        }
        $stagingOwnerSnapshot = & $ReadStableContainedFileBytes $stagingPath '.owner.json' 65536 $false
        $stagingComposeSnapshot = & $ReadStableContainedFileBytes $stagingPath 'compose.unit.yaml' 131072 $false
        if ($stagingOwnerSnapshot.Sha256 -cne $ownerMarkerSha256) {
            throw 'The owned marker bytes do not match the reservation.'
        }
        if ($stagingComposeSnapshot.Sha256 -cne $composeSha256) {
            throw 'The owned Compose snapshot bytes do not match the reservation.'
        }
        if (('sha256:' + [string]$stagingComposeSnapshot.Sha256) -cne [string]$images.Control['compose_sha256']) {
            throw 'The owned Compose snapshot bytes do not match the closed control raw-byte pin.'
        }
        $stagingEntries = @([System.IO.Directory]::GetFileSystemEntries($stagingPath))
        if ($stagingEntries.Count -ne 2) {
            throw 'The staging directory does not contain exactly two owned files.'
        }
        $expectedStagingNames = @('.owner.json', 'compose.unit.yaml')
        foreach ($stagingEntry in $stagingEntries) {
            $stagingName = [System.IO.Path]::GetFileName($stagingEntry)
            if (
                $expectedStagingNames -cnotcontains $stagingName -or
                -not [System.IO.File]::Exists($stagingEntry) -or
                [System.IO.Directory]::Exists($stagingEntry) -or
                (([System.IO.File]::GetAttributes($stagingEntry) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)
            ) {
                throw 'The staging directory contains an unowned or unsafe child.'
            }
        }
        if ([System.IO.Directory]::Exists($reportPath) -or [System.IO.File]::Exists($reportPath)) {
            throw 'The report path appeared before staging publication.'
        }
        [System.IO.Directory]::Move($stagingPath, $reportPath)
        $reportOwned = $true
        $stagingPath = $null
        $reservationSnapshot = & $ReadStableContainedFileBytes $reportsRoot $reservationRelativePath 65536 $false
        if (
            $reservationSnapshot.Sha256 -cne $reservationSha256 -or
            [System.Convert]::ToBase64String([byte[]]$reservationSnapshot.Bytes) -cne [System.Convert]::ToBase64String([byte[]]$reservationBytes)
        ) {
            throw 'The report reservation bytes changed before release.'
        }
        [System.IO.File]::Delete($reservationPath)
        if ([System.IO.File]::Exists($reservationPath) -or [System.IO.Directory]::Exists($reservationPath)) {
            throw 'The verified report reservation was not removed.'
        }
        $publishedEntries = @([System.IO.Directory]::GetFileSystemEntries($reportPath))
        if ($publishedEntries.Count -ne 2) {
            throw 'The published report changed before raw-artifact initialization.'
        }
        $expectedPublishedNames = @('.owner.json', 'compose.unit.yaml')
        foreach ($publishedEntry in $publishedEntries) {
            $publishedName = [System.IO.Path]::GetFileName($publishedEntry)
            if (
                $expectedPublishedNames -cnotcontains $publishedName -or
                -not [System.IO.File]::Exists($publishedEntry) -or
                [System.IO.Directory]::Exists($publishedEntry) -or
                (([System.IO.File]::GetAttributes($publishedEntry) -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)
            ) {
                throw 'The published report contains an unowned or unsafe child.'
            }
        }
        $composePath = & $ResolveCanonicalContainedPath $reportPath 'compose.unit.yaml' $true $false
        $publishedComposeSnapshot = & $ReadStableContainedFileBytes $reportPath 'compose.unit.yaml' 131072 $false
        if (('sha256:' + [string]$publishedComposeSnapshot.Sha256) -cne [string]$images.Control['compose_sha256']) {
            throw 'The published Compose snapshot bytes do not match the closed control raw-byte pin.'
        }
        $rawArtifactRoot = & $ResolveCanonicalContainedPath $reportPath 'raw' $false $false
        if ([System.IO.Directory]::Exists($rawArtifactRoot) -or [System.IO.File]::Exists($rawArtifactRoot)) {
            throw 'The published raw-artifact root already exists.'
        }
        [System.IO.Directory]::CreateDirectory($rawArtifactRoot) | Out-Null
        $rawArtifactRoot = & $ResolveCanonicalContainedPath $reportPath 'raw' $true $false
        $rawAttributes = [System.IO.File]::GetAttributes($rawArtifactRoot)
        if (
            -not [System.IO.Directory]::Exists($rawArtifactRoot) -or
            ($rawAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
            @([System.IO.Directory]::GetFileSystemEntries($rawArtifactRoot)).Count -ne 0
        ) {
            throw 'The published raw-artifact root is not an empty canonical directory.'
        }
        $dependencyContext.ComposePath = $composePath
        [void]$persistedArtifacts.Add('.owner.json')
        [void]$persistedArtifacts.Add('compose.unit.yaml')
    }
    catch {
        # K01 fail-leave: no recursive delete, adoption, or cleanup of a reservation,
        # owner marker, snapshot, restricted raw artifact, or failed publication.
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

        foreach ($requiredVerifiedAttestationField in @('ImageId', 'Entrypoint', 'EntrypointSha256', 'Cmd', 'CmdSha256', 'LabelsProjectionSha256')) {
            & $AssertRequiredProperty $imageAttestation $requiredVerifiedAttestationField
        }
        if ([string]$imageAttestation.ImageId -cne [string]$images.Targets[$Request.Target]['image_id']) {
            $contractFailure = $true
            throw 'The verified image attestation result does not retain the selected locked ImageId.'
        }
        $composeEnvironment = [ordered]@{
            BACKEND_UNIT_IMAGE_ID = [string]$images.Targets['backend-unit']['image_id']
            FRONTEND_UNIT_IMAGE_ID = [string]$images.Targets['frontend-unit']['image_id']
            BACKEND_OPENAPI_IMAGE_ID = [string]$images.Targets['backend-openapi']['image_id']
            HUB_QA_BACKEND_UNIT_CONTAINER_NAME = ($project + '-backend-unit')
            HUB_QA_FRONTEND_UNIT_CONTAINER_NAME = ($project + '-frontend-unit')
            HUB_QA_BACKEND_OPENAPI_CONTAINER_NAME = ($project + '-backend-openapi')
            HUB_QA_REPORT_DIR = $reportPath
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
            # Closed raw-stream boundary: unredacted process bytes are retained only in
            # these two owned raw artifacts. Structured run, cleanup, and command
            # receipts retain exact artifact paths and metadata only; launcher logs do
            # not receive the raw bytes or decoded strings.
            $composeStdoutArtifactPath = & $ResolveCanonicalContainedPath $reportPath ([string]$rawProcessArtifactPaths['compose_stdout']) $false $false
            $composeStderrArtifactPath = & $ResolveCanonicalContainedPath $reportPath ([string]$rawProcessArtifactPaths['compose_stderr']) $false $false
            & $WriteAtomicBytes $composeStdoutArtifactPath ([byte[]]$composeResult.StdoutBytes)
            [void]$persistedArtifacts.Add([string]$rawProcessArtifactPaths['compose_stdout'])
            & $WriteAtomicBytes $composeStderrArtifactPath ([byte[]]$composeResult.StderrBytes)
            [void]$persistedArtifacts.Add([string]$rawProcessArtifactPaths['compose_stderr'])
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
                foreach ($requiredOpenApiField in @('source_commit', 'source_policy_sha256', 'source_tree_manifest_sha256', 'build_manifest_sha256', 'generated_count', 'generated_files', 'generation_manifests', 'artifact_revision')) {
                    & $AssertRequiredProperty $targetResult $requiredOpenApiField
                }
                if (
                    [string]$targetResult.source_commit -cne [string]$images.Source['commit'] -or
                    [string]$targetResult.source_policy_sha256 -cne [string]$images.Source['policy_sha256'] -or
                    [string]$targetResult.source_tree_manifest_sha256 -cne [string]$images.Source['tree_manifest_sha256'] -or
                    [string]$targetResult.build_manifest_sha256 -cne [string]$images.Build['manifest_sha256']
                ) {
                    $contractFailure = $true
                    throw 'The OpenAPI receipt does not equal the selected image baked provenance.'
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
                $revisionRows.Add(('source_commit' + [char]9 + [string]$images.Source['commit']))
                $revisionRows.Add(('source_tree_manifest_sha256' + [char]9 + [string]$images.Source['tree_manifest_sha256']))
                for ($generation = 1; $generation -le 2; $generation++) {
                    $generationRoot = & $ResolveCanonicalContainedPath $reportPath ('openapi\generation-' + $generation) $true $false
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
                    throw 'The OpenAPI artifact revision does not bind baked source provenance and both ordered manifests.'
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
            control_commit = [string]$images.Control['commit']
            control_bundle_sha256 = [string]$images.Control['bundle_sha256']
            control_compose_sha256 = [string]$images.Control['compose_sha256']
            source_commit = [string]$images.Source['commit']
            source_policy_sha256 = [string]$images.Source['policy_sha256']
            source_tree_manifest_sha256 = [string]$images.Source['tree_manifest_sha256']
            build_manifest_sha256 = [string]$images.Build['manifest_sha256']
            images_lock_sha256 = $imagesLockSha256
            compose_sha256 = $composeSha256
            target = $Request.Target
            task = $Request.Task
            run = $Request.Run
            project = $project
            container = $container
            target_image_id = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['ImageId']) { [string]$imageAttestation.ImageId } else { [string]$images.Targets[$Request.Target]['image_id'] }
            entrypoint = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['Entrypoint']) { @($imageAttestation.Entrypoint) } else { @($images.Targets[$Request.Target]['entrypoint']) }
            entrypoint_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['EntrypointSha256']) { [string]$imageAttestation.EntrypointSha256 } else { $null }
            cmd = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['Cmd']) { @($imageAttestation.Cmd) } else { @($images.Targets[$Request.Target]['cmd']) }
            cmd_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['CmdSha256']) { [string]$imageAttestation.CmdSha256 } else { $null }
            labels_projection_sha256 = if ($null -ne $imageAttestation -and $null -ne $imageAttestation.PSObject.Properties['LabelsProjectionSha256']) { [string]$imageAttestation.LabelsProjectionSha256 } else { $null }
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
            changed_files = if ($Request.Target -ceq 'backend-openapi') { @() } else { 'N/A' }
            changed_count = if ($Request.Target -ceq 'backend-openapi') { 0 } else { 'N/A' }
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
