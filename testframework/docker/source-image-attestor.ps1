[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [ValidatePattern('^[0-9a-f]{40}$')]
    [string]$ControlCommit
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Private P2 materializer only. It is not a qa-gate target, does not attest
# itself, and is never a target-runner dependency operation.

class HubQaUtf8BytePathComparer : System.Collections.Generic.IComparer[string] {
    [int] Compare([string]$left, [string]$right) {
        if ([Object]::ReferenceEquals($left, $right)) { return 0 }
        if ($null -eq $left) { return -1 }
        if ($null -eq $right) { return 1 }

        $encoding = [Text.UTF8Encoding]::new($false, $true)
        [byte[]]$leftBytes = $encoding.GetBytes($left)
        [byte[]]$rightBytes = $encoding.GetBytes($right)
        $sharedLength = [Math]::Min($leftBytes.Length, $rightBytes.Length)
        for ($index = 0; $index -lt $sharedLength; $index++) {
            $difference = [int]$leftBytes[$index] - [int]$rightBytes[$index]
            if ($difference -lt 0) { return -1 }
            if ($difference -gt 0) { return 1 }
        }
        if ($leftBytes.Length -lt $rightBytes.Length) { return -1 }
        if ($leftBytes.Length -gt $rightBytes.Length) { return 1 }
        return 0
    }
}

$utf8 = [Text.UTF8Encoding]::new($false, $true)
$controlAssetPaths = [string[]]@(
    'testframework/docker/Dockerfile.backend-openapi',
    'testframework/docker/Dockerfile.backend-unit',
    'testframework/docker/Dockerfile.frontend-unit',
    'testframework/docker/build-manifest.v2.json',
    'testframework/docker/compose.unit.yaml',
    'testframework/docker/source-policy.v1.json'
)
$controlLabelKeys = [string[]]@(
    'hub.qa.control-commit',
    'hub.qa.control-bundle-sha256'
)
$script:controlBlobRequests = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)

function Resolve-CanonicalNonReparseDirectory {
    param([string]$Path, [string]$Identity)

    $item = Get-Item -LiteralPath $Path -Force -ErrorAction Stop
    if (-not $item.PSIsContainer) { throw "$Identity must be a directory: $Path" }

    $cursor = [IO.DirectoryInfo]$item
    while ($null -ne $cursor) {
        if ($cursor.Attributes -band [IO.FileAttributes]::ReparsePoint) {
            throw "$Identity contains a reparse point: $($cursor.FullName)"
        }
        $cursor = $cursor.Parent
    }
    return [IO.Path]::GetFullPath($item.FullName)
}

function Assert-ContainedPath {
    param([string]$Path, [string]$Root, [string]$Identity)

    $fullPath = [IO.Path]::GetFullPath($Path)
    $fullRoot = [IO.Path]::GetFullPath($Root)
    $rootPrefix = $fullRoot.TrimEnd('\', '/') + [IO.Path]::DirectorySeparatorChar
    if (-not $fullPath.StartsWith($rootPrefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw "$Identity escapes its declared root: $Path"
    }
    return $fullPath
}

function Test-FilesystemObjectExists {
    param([string]$Path)

    try {
        [IO.File]::GetAttributes($Path) | Out-Null
        return $true
    } catch [IO.FileNotFoundException] {
        return $false
    } catch [IO.DirectoryNotFoundException] {
        return $false
    }
}

function Assert-FilesystemObjectAbsent {
    param([string]$Path, [string]$Identity)

    if (Test-FilesystemObjectExists -Path $Path) {
        throw "$Identity already exists and must never be replaced: $Path"
    }
}

function Assert-ExactLowerSha1 {
    param([string]$Value, [string]$Identity)

    if ($Value -notmatch '^[0-9a-f]{40}$') { throw "$Identity must be exactly 40 lowercase hexadecimal characters." }
}

function Assert-ExactStringSequence {
    param([AllowNull()][object]$Actual, [string[]]$Expected, [string]$Identity)

    $actualValues = @($Actual)
    if ($actualValues.Count -ne $Expected.Count) { throw "$Identity count mismatch." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if (-not [StringComparer]::Ordinal.Equals([string]$actualValues[$index], $Expected[$index])) {
            throw "$Identity differs at index $index."
        }
    }
}

function Get-ByteSha256 {
    param([byte[]]$Bytes)

    $sha = [Security.Cryptography.SHA256]::Create()
    try {
        return ([BitConverter]::ToString($sha.ComputeHash($Bytes))).Replace('-', '').ToLowerInvariant()
    } finally {
        $sha.Dispose()
    }
}

function ConvertTo-JcsString {
    param([AllowNull()][object]$Value)

    if ($null -eq $Value -or -not ($Value -is [string])) { throw 'JCS string value must be a non-null string.' }
    $text = [string]$Value
    $builder = New-Object Text.StringBuilder
    [void]$builder.Append('"')
    for ($index = 0; $index -lt $text.Length; $index++) {
        $code = [int][char]$text[$index]
        if ($code -eq 0x22) { [void]$builder.Append('\"'); continue }
        if ($code -eq 0x5c) { [void]$builder.Append('\\'); continue }
        if ($code -eq 0x08) { [void]$builder.Append('\b'); continue }
        if ($code -eq 0x09) { [void]$builder.Append('\t'); continue }
        if ($code -eq 0x0a) { [void]$builder.Append('\n'); continue }
        if ($code -eq 0x0c) { [void]$builder.Append('\f'); continue }
        if ($code -eq 0x0d) { [void]$builder.Append('\r'); continue }
        if ($code -lt 0x20) { [void]$builder.Append(('\u{0:x4}' -f $code)); continue }
        if ($code -ge 0xd800 -and $code -le 0xdbff) {
            if (($index + 1) -ge $text.Length) { throw 'JCS string contains an unpaired high surrogate.' }
            $low = [int][char]$text[$index + 1]
            if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw 'JCS string contains an unpaired high surrogate.' }
            [void]$builder.Append([char]$code)
            [void]$builder.Append([char]$low)
            $index++
            continue
        }
        if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw 'JCS string contains an unpaired low surrogate.' }
        [void]$builder.Append([char]$code)
    }
    [void]$builder.Append('"')
    return $builder.ToString()
}

function Get-JcsSortedStrings {
    param([AllowEmptyCollection()][string[]]$Values)

    $ordered = [string[]]@($Values)
    [Array]::Sort($ordered, [StringComparer]::Ordinal)
    return $ordered
}

function Get-Utf8ByteSortedStrings {
    param([AllowEmptyCollection()][string[]]$Values)

    $ordered = [string[]]@($Values)
    [Array]::Sort($ordered, [HubQaUtf8BytePathComparer]::new())
    return $ordered
}

function ConvertTo-JcsStringArray {
    param([AllowNull()][object]$Value)

    if ($null -eq $Value -or $Value -is [string]) { throw 'JCS array must be a non-null array of strings.' }
    $encoded = New-Object Collections.Generic.List[string]
    foreach ($item in @($Value)) { $encoded.Add((ConvertTo-JcsString -Value $item)) }
    return '[' + ($encoded -join ',') + ']'
}

function Get-FullLabelsProjection {
    param([AllowNull()][object]$Labels)

    if ($null -eq $Labels) { throw 'Docker Config.Labels must be a non-null object.' }
    $propertyNames = @($Labels.psobject.Properties.Name)
    if ($propertyNames.Count -eq 0) { throw 'Docker Config.Labels must not be empty.' }
    $projection = [ordered]@{}
    foreach ($name in @(Get-JcsSortedStrings -Values ([string[]]$propertyNames))) {
        $value = $Labels.$name
        if ($null -eq $value -or -not ($value -is [string])) { throw "Docker label must be a string: $name" }
        $projection[$name] = [string]$value
    }
    return $projection
}

function ConvertTo-JcsStringObject {
    param([Collections.IDictionary]$Value)

    if ($null -eq $Value -or $Value.Count -eq 0) { throw 'JCS label projection must be a nonempty object.' }
    $members = New-Object Collections.Generic.List[string]
    foreach ($name in @(Get-JcsSortedStrings -Values ([string[]]@($Value.Keys)))) {
        $members.Add(('{0}:{1}' -f (ConvertTo-JcsString -Value $name), (ConvertTo-JcsString -Value $Value[$name])))
    }
    return '{' + ($members -join ',') + '}'
}

function Get-JcsArraySha256 {
    param([AllowNull()][object]$Value)

    return 'sha256:' + (Get-ByteSha256 -Bytes $utf8.GetBytes((ConvertTo-JcsStringArray -Value $Value)))
}

function Get-JcsObjectSha256 {
    param([Collections.IDictionary]$Value)

    return 'sha256:' + (Get-ByteSha256 -Bytes $utf8.GetBytes((ConvertTo-JcsStringObject -Value $Value)))
}

function Assert-GitCommitObject {
    param([string]$Repository, [string]$Commit, [string]$Identity)

    Assert-ExactLowerSha1 -Value $Commit -Identity $Identity
    $type = @(& git -C $Repository cat-file -t $Commit 2>$null)
    if ($LASTEXITCODE -ne 0 -or $type.Count -ne 1 -or [string]$type[0] -ne 'commit') {
        throw "$Identity is not an available commit object: $Commit"
    }
}

function Get-ExactGitTreeEntry {
    param([string]$Repository, [string]$Commit, [string]$RepositoryRelativePath)

    if ($RepositoryRelativePath -notmatch '^[A-Za-z0-9._/-]+$' -or $RepositoryRelativePath.StartsWith('/') -or $RepositoryRelativePath.Contains('..')) {
        throw "Unsafe repository-relative path: $RepositoryRelativePath"
    }
    $lines = @(& git -C $Repository ls-tree $Commit -- $RepositoryRelativePath 2>$null)
    if ($LASTEXITCODE -ne 0 -or $lines.Count -ne 1) { throw "Git tree entry is unavailable: $RepositoryRelativePath" }
    if ([string]$lines[0] -notmatch '^(?<mode>[0-7]{6}) (?<type>blob|tree|commit) (?<object>[0-9a-f]{40})\t(?<path>.+)$') {
        throw "Malformed Git tree entry: $RepositoryRelativePath"
    }
    if (-not [StringComparer]::Ordinal.Equals($Matches.path, $RepositoryRelativePath)) {
        throw "Git tree path mismatch: $RepositoryRelativePath"
    }
    if ($Matches.mode -ne '100644' -or $Matches.type -ne 'blob') {
        throw "Git tree entry must be mode 100644 blob: $RepositoryRelativePath"
    }
    Assert-ExactLowerSha1 -Value $Matches.object -Identity "Git blob identity for $RepositoryRelativePath"
    return [pscustomobject]@{ Mode = $Matches.mode; BlobSha1 = $Matches.object }
}

function Read-GitBlobBytes {
    param([string]$Repository, [string]$ObjectSpec, [Int64]$MaximumBytes, [string]$Identity)

    if ($MaximumBytes -le 0) { throw "Invalid byte bound: $Identity" }
    if ($Repository.Contains('"') -or $ObjectSpec.Contains('"')) { throw "Unsafe Git argument: $Identity" }

    $processInfo = New-Object Diagnostics.ProcessStartInfo
    $processInfo.FileName = 'git'
    $processInfo.WorkingDirectory = $Repository
    $processInfo.Arguments = ('cat-file blob "{0}"' -f $ObjectSpec)
    $processInfo.UseShellExecute = $false
    $processInfo.RedirectStandardOutput = $true
    $processInfo.RedirectStandardError = $true
    $process = New-Object Diagnostics.Process
    $process.StartInfo = $processInfo
    if (-not $process.Start()) { throw "Cannot start Git blob reader: $Identity" }

    $buffer = New-Object byte[] 65536
    $memory = New-Object IO.MemoryStream
    $count = [Int64]0
    try {
        while (($read = $process.StandardOutput.BaseStream.Read($buffer, 0, $buffer.Length)) -gt 0) {
            $count += $read
            if ($count -gt $MaximumBytes) { throw "Git blob byte bound exceeded: $Identity" }
            $memory.Write($buffer, 0, $read)
        }
        $stderr = $process.StandardError.ReadToEnd()
        $process.WaitForExit()
        if ($process.ExitCode -ne 0) { throw "Git blob read failed for $Identity: $stderr" }
        return $memory.ToArray()
    } finally {
        $memory.Dispose()
        $process.Dispose()
    }
}

function Read-ControlCommitBlobExact {
    param([string]$Repository, [string]$Commit, [string]$RepositoryRelativePath)

    Assert-ExactLowerSha1 -Value $Commit -Identity 'ControlCommit'
    if ($controlAssetPaths -notcontains $RepositoryRelativePath) {
        throw "Control path is not in the closed P0 set: $RepositoryRelativePath"
    }
    if (-not $script:controlBlobRequests.Add($RepositoryRelativePath)) {
        throw "Duplicate control blob request is forbidden: $RepositoryRelativePath"
    }

    $entry = Get-ExactGitTreeEntry -Repository $Repository -Commit $Commit -RepositoryRelativePath $RepositoryRelativePath
    [byte[]]$bytes = Read-GitBlobBytes -Repository $Repository -ObjectSpec ('{0}:{1}' -f $Commit, $RepositoryRelativePath) -MaximumBytes 4194304 -Identity $RepositoryRelativePath
    if ($bytes.Length -le 0) { throw "Control blob is empty: $RepositoryRelativePath" }
    return [pscustomobject]@{
        Path = $RepositoryRelativePath
        Mode = [string]$entry.Mode
        BlobSha1 = [string]$entry.BlobSha1
        Bytes = $bytes
        Sha256 = ('sha256:' + (Get-ByteSha256 -Bytes $bytes))
    }
}

function ConvertFrom-ControlJsonBytes {
    param([byte[]]$Bytes, [string]$Identity)

    if ($Bytes.Length -lt 2 -or ($Bytes.Length -ge 3 -and $Bytes[0] -eq 0xef -and $Bytes[1] -eq 0xbb -and $Bytes[2] -eq 0xbf)) {
        throw "$Identity must be nonempty no-BOM UTF-8 JSON."
    }
    try {
        $text = $utf8.GetString($Bytes)
        return $text | ConvertFrom-Json -ErrorAction Stop
    } catch {
        throw "Invalid control JSON: $Identity"
    }
}

function Assert-ControlTextBindings {
    param([hashtable]$ControlBlobs)

    foreach ($dockerfilePath in @(
        'testframework/docker/Dockerfile.backend-openapi',
        'testframework/docker/Dockerfile.backend-unit',
        'testframework/docker/Dockerfile.frontend-unit'
    )) {
        $text = $utf8.GetString([byte[]]$ControlBlobs[$dockerfilePath].Bytes)
        foreach ($requiredToken in @('ARG CONTROL_COMMIT', 'ARG CONTROL_BUNDLE_SHA256', 'hub.qa.control-commit', 'hub.qa.control-bundle-sha256', 'hub.qa.runtime-status="NOT_READY"')) {
            if (-not $text.Contains($requiredToken)) { throw "Control Dockerfile misses $requiredToken: $dockerfilePath" }
        }
    }

    $composeText = $utf8.GetString([byte[]]$ControlBlobs['testframework/docker/compose.unit.yaml'].Bytes)
    foreach ($requiredToken in @('images.lock.v3.json', 'strict-fixed-target-image-id-and-control-commit')) {
        if (-not $composeText.Contains($requiredToken)) { throw "Control Compose misses $requiredToken." }
    }
    if ($composeText.Contains('images.lock.v2.json')) { throw 'Control Compose must not name the v2 lock.' }
}

function Read-ValidatedControlBundle {
    param([string]$Repository, [string]$Commit)

    $blobs = @{}
    foreach ($path in $controlAssetPaths) {
        $blobs[$path] = Read-ControlCommitBlobExact -Repository $Repository -Commit $Commit -RepositoryRelativePath $path
    }
    Assert-ExactStringSequence -Actual ([string[]]@($blobs.Keys | Sort-Object)) -Expected ([string[]]@($controlAssetPaths | Sort-Object)) -Identity 'Control blob membership'

    $sourcePolicyPath = 'testframework/docker/source-policy.v1.json'
    $buildManifestPath = 'testframework/docker/build-manifest.v2.json'
    $composePath = 'testframework/docker/compose.unit.yaml'
    $policy = ConvertFrom-ControlJsonBytes -Bytes ([byte[]]$blobs[$sourcePolicyPath].Bytes) -Identity $sourcePolicyPath
    $buildManifest = ConvertFrom-ControlJsonBytes -Bytes ([byte[]]$blobs[$buildManifestPath].Bytes) -Identity $buildManifestPath

    if ([string]$policy.schema -ne 'hub.qa.local-source-policy/v1') { throw 'Unexpected source-policy schema.' }
    if ([string]$buildManifest.schema -ne 'hub.qa.local-images-build-manifest/v2') { throw 'Unexpected build-manifest schema.' }
    if ([string]$buildManifest.source_policy -ne 'source-policy.v1.json') { throw 'Build manifest must name source-policy.v1.json.' }
    Assert-ExactLowerSha1 -Value ([string]$policy.source_commit) -Identity 'Source policy source_commit'
    if ([string]$policy.source_commit -ne [string]$buildManifest.source_commit) { throw 'Source commit disagreement between control blobs.' }
    if ([string]$buildManifest.platform -ne 'linux/amd64') { throw 'Unsupported controlled build platform.' }
    if ([string]$buildManifest.control_provenance.authority -ne 'private-control-commit-only') {
        throw 'Unexpected control provenance authority.'
    }
    Assert-ExactStringSequence -Actual $buildManifest.control_provenance.paths -Expected $controlAssetPaths -Identity 'Control provenance paths'
    Assert-ExactStringSequence -Actual $buildManifest.control_provenance.label_keys -Expected $controlLabelKeys -Identity 'Control provenance label keys'
    Assert-ExactStringSequence -Actual ([string[]]@($buildManifest.base_inventory.psobject.Properties.Name | Sort-Object)) -Expected ([string[]]@('golang', 'playwright', 'python')) -Identity 'Base inventory keys'
    Assert-ExactStringSequence -Actual ([string[]]@($buildManifest.targets.psobject.Properties.Name | Sort-Object)) -Expected ([string[]]@('backend-openapi', 'backend-unit', 'frontend-unit')) -Identity 'Target keys'
    Assert-ControlTextBindings -ControlBlobs $blobs

    $controlRows = New-Object Collections.Generic.List[string]
    foreach ($path in $controlAssetPaths) {
        $blob = $blobs[$path]
        if ([string]$blob.Mode -ne '100644') { throw "Unexpected control mode: $path" }
        $controlRows.Add(("{0}`t100644`t{1}`t{2}`n" -f $path, [string]$blob.BlobSha1, [string]$blob.Sha256))
    }
    $bundleBytes = $utf8.GetBytes(($controlRows -join ''))
    return [pscustomobject]@{
        Blobs = $blobs
        SourcePolicy = $policy
        BuildManifest = $buildManifest
        BundleSha256 = ('sha256:' + (Get-ByteSha256 -Bytes $bundleBytes))
        ComposeSha256 = [string]$blobs[$composePath].Sha256
    }
}

function Assert-DependenciesReady {
    param([object]$BuildManifest)

    if ([string]$BuildManifest.build_preflight.status -ne 'READY') {
        throw ('NOT_READY: {0}' -f [string]$BuildManifest.build_preflight.reason)
    }
    foreach ($targetName in @('backend-unit', 'frontend-unit', 'backend-openapi')) {
        if ([string]$BuildManifest.targets.$targetName.dependency_assets.status -ne 'READY') {
            throw "NOT_READY: dependency assets are not approved for $targetName"
        }
    }
}

function Test-PolicyPath {
    param([string]$Path, [object]$Policy)

    $normalized = $Path.Replace('\', '/')
    foreach ($segment in @($normalized.Split('/'))) {
        foreach ($excludedSegment in @($Policy.materialization.exclude_segments_any_depth)) {
            if ([StringComparer]::OrdinalIgnoreCase.Equals($segment, [string]$excludedSegment)) { return $false }
        }
    }
    $included = $false
    foreach ($pattern in @($Policy.include)) {
        if ($normalized -like [string]$pattern) { $included = $true; break }
    }
    if (-not $included) { return $false }
    foreach ($pattern in @($Policy.exclude)) {
        if ($normalized -like [string]$pattern) { return $false }
    }
    return $true
}

function Get-ExactCommitTree {
    param([string]$Repository, [object]$Policy)

    $commit = [string]$Policy.source_commit
    Assert-GitCommitObject -Repository $Repository -Commit $commit -Identity 'Source commit'
    $tree = @(& git -C $Repository ls-tree -r --full-tree $commit 2>$null)
    if ($LASTEXITCODE -ne 0) { throw "Declared source commit is unavailable: $commit" }

    $entriesByPath = @{}
    foreach ($line in $tree) {
        if ([string]$line -notmatch '^(?<mode>[0-7]{6}) (?<type>blob|tree|commit) (?<object>[0-9a-f]{40})\t(?<path>.+)$') {
            throw 'Unexpected Git tree record.'
        }
        $path = [string]$Matches.path
        if (-not (Test-PolicyPath -Path $path -Policy $Policy)) { continue }
        if ($Matches.mode -ne '100644' -or $Matches.type -ne 'blob') { throw "Source tree entry must be mode 100644 blob: $path" }
        if ($path.Contains(':') -or $path.StartsWith('/') -or $path.Contains('\') -or $path.Split('/') -contains '..' -or $path.Split('/') -contains '.') {
            throw "Unsafe source path: $path"
        }
        if ($entriesByPath.ContainsKey($path)) { throw "Duplicate or case-colliding source path: $path" }
        $entriesByPath[$path] = [pscustomobject]@{ path = $path; mode = $Matches.mode; object = $Matches.object }
        if ($entriesByPath.Count -gt [int]$Policy.materialization.max_files) { throw 'Source file-count bound exceeded.' }
    }
    if ($entriesByPath.Count -eq 0) { throw 'The declared source policy selected no Git files.' }

    $orderedEntries = New-Object Collections.Generic.List[object]
    foreach ($path in @(Get-Utf8ByteSortedStrings -Values ([string[]]@($entriesByPath.Keys)))) { $orderedEntries.Add($entriesByPath[$path]) }
    return @($orderedEntries)
}

function Write-BytesCreateNew {
    param([string]$Destination, [byte[]]$Bytes)

    $parent = Split-Path -Parent $Destination
    [IO.Directory]::CreateDirectory($parent) | Out-Null
    $stream = [IO.File]::Open($Destination, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
    try {
        $stream.Write($Bytes, 0, $Bytes.Length)
        $stream.Flush($true)
    } finally {
        $stream.Dispose()
    }
}

function Materialize-ExactSource {
    param([string]$Repository, [object]$Policy, [object[]]$Entries, [string]$ContextRoot)

    $sourceRoot = Assert-ContainedPath -Path (Join-Path $ContextRoot 'source') -Root $ContextRoot -Identity 'Source materialization root'
    [IO.Directory]::CreateDirectory($sourceRoot) | Out-Null
    $manifestLines = New-Object Collections.Generic.List[string]
    $totalBytes = [Int64]0
    foreach ($entry in $Entries) {
        $destination = Assert-ContainedPath -Path (Join-Path $sourceRoot $entry.path) -Root $sourceRoot -Identity 'Source materialization destination'
        [byte[]]$bytes = Read-GitBlobBytes -Repository $Repository -ObjectSpec ('{0}:{1}' -f [string]$Policy.source_commit, [string]$entry.path) -MaximumBytes ([Int64]$Policy.materialization.max_file_bytes) -Identity ([string]$entry.path)
        $totalBytes += $bytes.Length
        if ($totalBytes -gt [Int64]$Policy.materialization.max_total_bytes) { throw 'Total source byte bound exceeded.' }
        Write-BytesCreateNew -Destination $destination -Bytes $bytes
        $manifestLines.Add(("{0}`t{1}`t{2}" -f [string]$entry.path, [string]$entry.mode, (Get-ByteSha256 -Bytes $bytes)))
    }
    $canonicalManifest = ((@(Get-Utf8ByteSortedStrings -Values ([string[]]@($manifestLines))) -join "`n") + "`n")
    return [pscustomobject]@{
        Root = $sourceRoot
        Bytes = $totalBytes
        Sha256 = ('sha256:' + (Get-ByteSha256 -Bytes $utf8.GetBytes($canonicalManifest)))
    }
}

function New-ReservedStagingRoot {
    $nonce = [Guid]::NewGuid().ToString('N')
    $name = 'hub-qa-local-image-attestor-' + $nonce
    $temporaryRoot = Resolve-CanonicalNonReparseDirectory -Path ([IO.Path]::GetTempPath()) -Identity 'temporary root'
    $root = Join-Path $temporaryRoot $name
    [IO.Directory]::CreateDirectory($root) | Out-Null
    $marker = Join-Path $root '.reservation.json'
    [byte[]]$markerBytes = $utf8.GetBytes(('{' + '"owner":"hub.qa.local-image-attestor","nonce":"' + $nonce + '"}'))
    Write-BytesCreateNew -Destination $marker -Bytes $markerBytes
    return [pscustomobject]@{
        Root = [IO.Path]::GetFullPath($root)
        TemporaryRoot = $temporaryRoot
        MarkerPath = [IO.Path]::GetFullPath($marker)
        MarkerBytes = $markerBytes
        Nonce = $nonce
    }
}

function Test-ExactBytesEqual {
    param([byte[]]$Left, [byte[]]$Right)

    if ($Left.Length -ne $Right.Length) { return $false }
    for ($index = 0; $index -lt $Left.Length; $index++) {
        if ($Left[$index] -ne $Right[$index]) { return $false }
    }
    return $true
}

function Add-ExpectedDirectoryChain {
    param(
        [Collections.Generic.HashSet[string]]$ExpectedDirectories,
        [string]$Directory,
        [string]$StagingRoot
    )

    $current = [IO.Path]::GetFullPath($Directory)
    $root = [IO.Path]::GetFullPath($StagingRoot)
    while ($true) {
        [void]$ExpectedDirectories.Add($current)
        if ([StringComparer]::OrdinalIgnoreCase.Equals($current, $root)) { break }
        $parent = [IO.Path]::GetDirectoryName($current)
        if ([string]::IsNullOrEmpty($parent)) { throw 'Expected staging directory chain escaped its root.' }
        $parentPrefix = $root.TrimEnd('\', '/') + [IO.Path]::DirectorySeparatorChar
        if (-not [StringComparer]::OrdinalIgnoreCase.Equals($parent, $root) -and
            -not $parent.StartsWith($parentPrefix, [StringComparison]::OrdinalIgnoreCase)) {
            throw 'Expected staging directory chain escaped its root.'
        }
        $current = $parent
    }
}

function Remove-OwnedStagingAfterPublication {
    param(
        [object]$Reservation,
        [object[]]$SourceEntries,
        [string[]]$TargetNames
    )

    if ([string]$Reservation.Nonce -notmatch '^[0-9a-f]{32}$') { throw 'Owned staging nonce is malformed.' }
    $temporaryRoot = Resolve-CanonicalNonReparseDirectory -Path ([string]$Reservation.TemporaryRoot) -Identity 'owned staging temporary root'
    $stagingRoot = Resolve-CanonicalNonReparseDirectory -Path ([string]$Reservation.Root) -Identity 'owned staging root'
    $expectedRootName = 'hub-qa-local-image-attestor-' + [string]$Reservation.Nonce
    if (-not [StringComparer]::Ordinal.Equals([IO.Path]::GetFileName($stagingRoot), $expectedRootName)) {
        throw 'Owned staging root name does not match its nonce.'
    }
    $temporaryPrefix = $temporaryRoot.TrimEnd('\', '/') + [IO.Path]::DirectorySeparatorChar
    if (-not $stagingRoot.StartsWith($temporaryPrefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Owned staging root escaped the canonical temporary root.'
    }

    $expectedFiles = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
    $expectedDirectories = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
    [void]$expectedDirectories.Add($stagingRoot)
    Assert-ExactStringSequence -Actual $TargetNames -Expected ([string[]]@('backend-unit', 'frontend-unit', 'backend-openapi')) -Identity 'owned staging target names'
    $markerPath = Assert-ContainedPath -Path ([string]$Reservation.MarkerPath) -Root $stagingRoot -Identity 'owned staging marker'
    $expectedMarkerPath = [IO.Path]::GetFullPath((Join-Path $stagingRoot '.reservation.json'))
    if (-not [StringComparer]::OrdinalIgnoreCase.Equals($markerPath, $expectedMarkerPath)) {
        throw 'Owned staging marker path mismatch.'
    }
    [void]$expectedFiles.Add($markerPath)

    foreach ($targetName in $TargetNames) {
        if ($targetName -notmatch '^(backend-unit|frontend-unit|backend-openapi)$') { throw "Unexpected owned target name: $targetName" }
        $targetRoot = Assert-ContainedPath -Path (Join-Path $stagingRoot $targetName) -Root $stagingRoot -Identity 'owned target root'
        $sourceRoot = Assert-ContainedPath -Path (Join-Path $targetRoot 'source') -Root $stagingRoot -Identity 'owned source root'
        Add-ExpectedDirectoryChain -ExpectedDirectories $expectedDirectories -Directory $sourceRoot -StagingRoot $stagingRoot
        foreach ($knownFileName in @('Dockerfile', '.iid')) {
            $knownFile = Assert-ContainedPath -Path (Join-Path $targetRoot $knownFileName) -Root $stagingRoot -Identity 'known owned target file'
            [void]$expectedFiles.Add($knownFile)
        }
        foreach ($entry in $SourceEntries) {
            $sourceFile = Assert-ContainedPath -Path (Join-Path $sourceRoot ([string]$entry.path)) -Root $stagingRoot -Identity 'owned materialized source file'
            if (-not $expectedFiles.Add($sourceFile)) { throw "Duplicate owned source path: $sourceFile" }
            Add-ExpectedDirectoryChain -ExpectedDirectories $expectedDirectories -Directory ([IO.Path]::GetDirectoryName($sourceFile)) -StagingRoot $stagingRoot
        }
    }

    $actualFiles = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
    $actualDirectories = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
    $pendingDirectories = New-Object Collections.Generic.Stack[string]
    $pendingDirectories.Push($stagingRoot)
    while ($pendingDirectories.Count -gt 0) {
        $directory = $pendingDirectories.Pop()
        $directoryItem = Get-Item -LiteralPath $directory -Force -ErrorAction Stop
        if (-not $directoryItem.PSIsContainer -or ($directoryItem.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
            throw "Owned staging directory is not a regular non-reparse directory: $directory"
        }
        $canonicalDirectory = [IO.Path]::GetFullPath($directoryItem.FullName)
        if (-not $expectedDirectories.Contains($canonicalDirectory)) { throw "Foreign owned-staging directory: $canonicalDirectory" }
        [void]$actualDirectories.Add($canonicalDirectory)
        foreach ($child in @(Get-ChildItem -LiteralPath $canonicalDirectory -Force -ErrorAction Stop)) {
            if ($child.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw "Reparse point in owned staging: $($child.FullName)" }
            $childPath = Assert-ContainedPath -Path $child.FullName -Root $stagingRoot -Identity 'owned staging child'
            if ($child.PSIsContainer) {
                if (-not $expectedDirectories.Contains($childPath)) { throw "Foreign owned-staging directory: $childPath" }
                $pendingDirectories.Push($childPath)
            } else {
                if (-not $expectedFiles.Contains($childPath)) { throw "Foreign owned-staging file: $childPath" }
                [void]$actualFiles.Add($childPath)
            }
        }
    }
    if ($actualFiles.Count -ne $expectedFiles.Count -or $actualDirectories.Count -ne $expectedDirectories.Count) {
        throw 'Owned staging tree is missing an expected file or directory.'
    }
    foreach ($path in $expectedFiles) {
        if (-not $actualFiles.Contains($path)) { throw "Owned staging file is missing: $path" }
    }
    foreach ($path in $expectedDirectories) {
        if (-not $actualDirectories.Contains($path)) { throw "Owned staging directory is missing: $path" }
    }
    [byte[]]$expectedMarkerBytes = $utf8.GetBytes(('{' + '"owner":"hub.qa.local-image-attestor","nonce":"' + [string]$Reservation.Nonce + '"}'))
    if (-not (Test-ExactBytesEqual -Left ([byte[]]$Reservation.MarkerBytes) -Right $expectedMarkerBytes)) {
        throw 'Owned staging in-memory marker/nonce mismatch.'
    }
    [byte[]]$observedMarkerBytes = [IO.File]::ReadAllBytes($markerPath)
    if (-not (Test-ExactBytesEqual -Left $observedMarkerBytes -Right $expectedMarkerBytes)) {
        throw 'Owned staging marker/nonce mismatch.'
    }

    # Validation is complete before the first delete. Known files and known
    # directories are removed exactly and non-recursively; a foreign or changed
    # tree is abandoned rather than adopted or recursively cleaned.
    foreach ($path in $expectedFiles) {
        if (-not [StringComparer]::OrdinalIgnoreCase.Equals($path, $markerPath)) { [IO.File]::Delete($path) }
    }
    [string[]]$directoriesToDelete = @($expectedDirectories | Where-Object { -not [StringComparer]::OrdinalIgnoreCase.Equals($_, $stagingRoot) })
    while ($directoriesToDelete.Count -gt 0) {
        $maximumDepth = ($directoriesToDelete | ForEach-Object { $_.Split([IO.Path]::DirectorySeparatorChar).Count } | Measure-Object -Maximum).Maximum
        [string[]]$deepestDirectories = @($directoriesToDelete | Where-Object { $_.Split([IO.Path]::DirectorySeparatorChar).Count -eq $maximumDepth })
        [Array]::Sort($deepestDirectories, [StringComparer]::OrdinalIgnoreCase)
        foreach ($directory in $deepestDirectories) { [IO.Directory]::Delete($directory, $false) }
        $directoriesToDelete = [string[]]@($directoriesToDelete | Where-Object { $_.Split([IO.Path]::DirectorySeparatorChar).Count -ne $maximumDepth })
    }
    [IO.File]::Delete($markerPath)
    [IO.Directory]::Delete($stagingRoot, $false)
}

function Invoke-LocalImageInspect {
    param([string]$ImageReference)

    if ($ImageReference -notmatch '^sha256:[0-9a-f]{64}$') { throw "Image reference must be a complete local ImageId: $ImageReference" }
    $format = '{"id":{{json .Id}},"repo_digests":{{json .RepoDigests}},"os":{{json .Os}},"architecture":{{json .Architecture}},"variant":{{json .Variant}},"rootfs_layers":{{json .RootFS.Layers}},"entrypoint":{{json .Config.Entrypoint}},"cmd":{{json .Config.Cmd}},"labels":{{json .Config.Labels}}}'
    $json = & docker image inspect --format $format $ImageReference 2>$null
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($json)) { throw "Local image is unavailable: $ImageReference" }
    try { return $json | ConvertFrom-Json -ErrorAction Stop } catch { throw "Local image inspect output is malformed: $ImageReference" }
}

function Get-ValidatedPlatform {
    param([object]$Actual, [string]$Identity)

    if ([string]$Actual.os -ne 'linux' -or [string]$Actual.architecture -ne 'amd64') { throw "Platform mismatch: $Identity" }
    if ($null -ne $Actual.variant -and -not [string]::IsNullOrEmpty([string]$Actual.variant)) { throw "Platform variant must be null: $Identity" }
    return [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }
}

function Get-ValidatedRootfsLayers {
    param([AllowNull()][object]$Value, [string]$Identity)

    if ($null -eq $Value) { throw "RootFS layer array is null: $Identity" }
    $layers = [string[]]@($Value)
    if ($layers.Count -eq 0) { throw "RootFS layer array is empty: $Identity" }
    foreach ($layer in $layers) {
        if ($layer -notmatch '^sha256:[0-9a-f]{64}$') { throw "Invalid RootFS layer identity: $Identity" }
    }
    return $layers
}

function Get-ValidatedConfigStringArray {
    param([AllowNull()][object]$Value, [string]$Field, [string]$Identity)

    if ($null -eq $Value -or $Value -is [string]) { throw "Docker Config.$Field must be a non-null string array: $Identity" }
    foreach ($item in @($Value)) {
        if ($null -eq $item -or -not ($item -is [string])) { throw "Docker Config.$Field contains a non-string: $Identity" }
    }
    return [string[]]@($Value)
}

function New-BaseRecord {
    param([string]$BaseKey, [object]$Inventory)

    $expectedImageId = [string]$Inventory.image_id
    if ($expectedImageId -notmatch '^sha256:[0-9a-f]{64}$') { throw "Build-manifest base ImageId is invalid: $BaseKey" }
    $actual = Invoke-LocalImageInspect -ImageReference $expectedImageId
    if ([string]$actual.id -ne $expectedImageId) { throw "Local base image ID mismatch: $BaseKey" }
    $platform = Get-ValidatedPlatform -Actual $actual -Identity $BaseKey
    if ([string]$Inventory.platform.os -ne 'linux' -or [string]$Inventory.platform.architecture -ne 'amd64' -or $null -ne $Inventory.platform.variant) {
        throw "Build-manifest base platform mismatch: $BaseKey"
    }
    $layers = @(Get-ValidatedRootfsLayers -Value $actual.rootfs_layers -Identity $BaseKey)
    if (($layers -join "`n") -ne (@($Inventory.rootfs_layers) -join "`n")) { throw "Local base rootfs layers mismatch: $BaseKey" }
    $entrypoint = @(Get-ValidatedConfigStringArray -Value $actual.entrypoint -Field 'Entrypoint' -Identity $BaseKey)
    $cmd = @(Get-ValidatedConfigStringArray -Value $actual.cmd -Field 'Cmd' -Identity $BaseKey)
    if (($cmd -join "`n") -ne (@($Inventory.cmd) -join "`n")) { throw "Local base Cmd mismatch: $BaseKey" }
    $repoDigest = [string]$Inventory.repo_digest
    if (-not [string]::IsNullOrEmpty($repoDigest)) {
        if ($repoDigest -notmatch '^[A-Za-z0-9./_-]+@sha256:[0-9a-f]{64}$') { throw "Build-manifest base RepoDigest is invalid: $BaseKey" }
        if ((@($actual.repo_digests) -notcontains $repoDigest)) { throw "Local base repo digest admission failed: $BaseKey" }
    }
    $labels = Get-FullLabelsProjection -Labels $actual.labels
    return [ordered]@{
        repo_digest = $repoDigest
        image_id = [string]$actual.id
        platform = $platform
        rootfs_layers = $layers
        entrypoint = $entrypoint
        entrypoint_sha256 = Get-JcsArraySha256 -Value $entrypoint
        cmd = $cmd
        cmd_sha256 = Get-JcsArraySha256 -Value $cmd
        labels_projection = $labels
        labels_projection_sha256 = Get-JcsObjectSha256 -Value $labels
    }
}

function Invoke-LocalBuild {
    param([string]$Context, [string]$Dockerfile, [object]$BuildArguments, [string]$IidFile)

    $arguments = @('build', '--pull=false', '--network=none', '--platform', 'linux/amd64', '--iidfile', $IidFile, '-f', $Dockerfile)
    foreach ($name in @(Get-JcsSortedStrings -Values ([string[]]@($BuildArguments.Keys)))) {
        $arguments += @('--build-arg', ('{0}={1}' -f $name, $BuildArguments[$name]))
    }
    $arguments += $Context
    & docker @arguments
    if ($LASTEXITCODE -ne 0) { throw 'Local image build failed.' }
    $imageId = [Text.Encoding]::ASCII.GetString([IO.File]::ReadAllBytes($IidFile)).Trim()
    if ($imageId -notmatch '^sha256:[0-9a-f]{64}$') { throw 'Docker build did not produce a complete local image ID.' }
    return $imageId
}

function Publish-NoOverwriteLock {
    param([string]$Destination, [string]$Content)

    $directory = Split-Path -Parent $Destination
    $temporary = Join-Path $directory ('.images.lock.v3.' + [Guid]::NewGuid().ToString('N') + '.partial')
    $stream = [IO.File]::Open($temporary, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
    try {
        [byte[]]$bytes = $utf8.GetBytes($Content)
        $stream.Write($bytes, 0, $bytes.Length)
        $stream.Flush($true)
    } finally {
        $stream.Dispose()
    }
    Assert-FilesystemObjectAbsent -Path $Destination -Identity 'v3 lock destination before publication'
    [IO.File]::Move($temporary, $Destination)
}

$scriptDirectory = Resolve-CanonicalNonReparseDirectory -Path $PSScriptRoot -Identity 'testframework/docker directory'
$repositoryRoot = Resolve-CanonicalNonReparseDirectory -Path (Join-Path $scriptDirectory '..\..') -Identity 'repository root'
$lockParent = Resolve-CanonicalNonReparseDirectory -Path $scriptDirectory -Identity 'v3 lock parent'
$lockPath = Assert-ContainedPath -Path (Join-Path $lockParent 'images.lock.v3.json') -Root $lockParent -Identity 'v3 lock destination'

# P0 gates 1-4: canonical roots, absent destination, immutable ControlCommit
# objects, then closed-control validation. No Docker, staging, or source action
# occurs before these gates complete.
Assert-FilesystemObjectAbsent -Path $lockPath -Identity 'v3 lock destination'
Assert-ExactLowerSha1 -Value $ControlCommit -Identity 'ControlCommit'
if (-not [StringComparer]::Ordinal.Equals($ControlCommit, '5afb8ba750f02d390ef75295db1afa12a041d7ff')) {
    throw 'ControlCommit is not the accepted P1A control commit.'
}
Assert-GitCommitObject -Repository $repositoryRoot -Commit $ControlCommit -Identity 'ControlCommit'
$control = Read-ValidatedControlBundle -Repository $repositoryRoot -Commit $ControlCommit

# P0 gate 5 is deliberately before local inspection, source staging, and builds.
Assert-DependenciesReady -BuildManifest $control.BuildManifest

$baseRecords = [ordered]@{}
foreach ($baseKey in @('golang', 'playwright', 'python')) {
    $baseRecords[$baseKey] = New-BaseRecord -BaseKey $baseKey -Inventory $control.BuildManifest.base_inventory.$baseKey
}

$sourceEntries = Get-ExactCommitTree -Repository $repositoryRoot -Policy $control.SourcePolicy
$stagingReservation = New-ReservedStagingRoot
$stagingRoot = [string]$stagingReservation.Root
$fixedTargetNames = [string[]]@('backend-unit', 'frontend-unit', 'backend-openapi')
$targetBaseMap = [ordered]@{
    'backend-unit' = 'golang'
    'frontend-unit' = 'playwright'
    'backend-openapi' = 'python'
}
$targetRecords = [ordered]@{}
$sourceTreeManifestSha = $null
$publicationSucceeded = $false

try {
    foreach ($targetName in $fixedTargetNames) {
        $target = $control.BuildManifest.targets.$targetName
        $baseKey = [string]$targetBaseMap[$targetName]
        if ([string]$target.base_inventory_key -ne $baseKey) { throw "Target-to-base mapping mismatch: $targetName" }
        $dockerfileName = [string]$target.dockerfile
        if ($dockerfileName -notmatch '^Dockerfile\.(backend-unit|frontend-unit|backend-openapi)$') { throw "Unexpected controlled Dockerfile: $targetName" }
        $dockerfilePath = 'testframework/docker/' + $dockerfileName
        if (-not $control.Blobs.ContainsKey($dockerfilePath)) { throw "Controlled Dockerfile is outside the P0 bundle: $targetName" }

        $contextRoot = Assert-ContainedPath -Path (Join-Path $stagingRoot $targetName) -Root $stagingRoot -Identity 'Target build context'
        [IO.Directory]::CreateDirectory($contextRoot) | Out-Null
        $source = Materialize-ExactSource -Repository $repositoryRoot -Policy $control.SourcePolicy -Entries $sourceEntries -ContextRoot $contextRoot
        if ($null -eq $sourceTreeManifestSha) {
            $sourceTreeManifestSha = [string]$source.Sha256
        } elseif ($sourceTreeManifestSha -ne [string]$source.Sha256) {
            throw "Exact source materialization changed between targets: $targetName"
        }

        $contextDockerfile = Assert-ContainedPath -Path (Join-Path $contextRoot 'Dockerfile') -Root $contextRoot -Identity 'Controlled Dockerfile destination'
        Write-BytesCreateNew -Destination $contextDockerfile -Bytes ([byte[]]$control.Blobs[$dockerfilePath].Bytes)
        $iidFile = Assert-ContainedPath -Path (Join-Path $contextRoot '.iid') -Root $contextRoot -Identity 'Docker IID file'
        $imageId = Invoke-LocalBuild -Context $contextRoot -Dockerfile $contextDockerfile -IidFile $iidFile -BuildArguments ([ordered]@{
            BASE_IMAGE_ID = [string]$baseRecords[$baseKey].image_id
            SOURCE_COMMIT = [string]$control.SourcePolicy.source_commit
            SOURCE_POLICY_SHA256 = [string]$control.Blobs['testframework/docker/source-policy.v1.json'].Sha256
            SOURCE_TREE_MANIFEST_SHA256 = $sourceTreeManifestSha
            BUILD_MANIFEST_SHA256 = [string]$control.Blobs['testframework/docker/build-manifest.v2.json'].Sha256
            CONTROL_COMMIT = $ControlCommit
            CONTROL_BUNDLE_SHA256 = [string]$control.BundleSha256
        })

        $actual = Invoke-LocalImageInspect -ImageReference $imageId
        if ([string]$actual.id -ne $imageId) { throw "Target image ID mismatch: $targetName" }
        $platform = Get-ValidatedPlatform -Actual $actual -Identity $targetName
        $layers = @(Get-ValidatedRootfsLayers -Value $actual.rootfs_layers -Identity $targetName)
        $entrypoint = @(Get-ValidatedConfigStringArray -Value $actual.entrypoint -Field 'Entrypoint' -Identity $targetName)
        $cmd = @(Get-ValidatedConfigStringArray -Value $actual.cmd -Field 'Cmd' -Identity $targetName)
        if (($entrypoint -join "`n") -ne (@($target.entrypoint) -join "`n")) { throw "Target entrypoint mismatch: $targetName" }
        $labels = Get-FullLabelsProjection -Labels $actual.labels
        $requiredLabels = [ordered]@{
            'hub.qa.target' = $targetName
            'hub.qa.base-image-id' = [string]$baseRecords[$baseKey].image_id
            'hub.qa.control-commit' = $ControlCommit
            'hub.qa.control-bundle-sha256' = [string]$control.BundleSha256
            'hub.qa.source-commit' = [string]$control.SourcePolicy.source_commit
            'hub.qa.source-policy-sha256' = [string]$control.Blobs['testframework/docker/source-policy.v1.json'].Sha256
            'hub.qa.source-tree-manifest-sha256' = $sourceTreeManifestSha
            'hub.qa.build-manifest-sha256' = [string]$control.Blobs['testframework/docker/build-manifest.v2.json'].Sha256
            'hub.qa.runtime-status' = 'NOT_READY'
        }
        foreach ($labelName in $requiredLabels.Keys) {
            if (-not $labels.Contains($labelName) -or [string]$labels[$labelName] -ne [string]$requiredLabels[$labelName]) {
                throw "Required target label binding mismatch: $targetName / $labelName"
            }
        }
        $targetRecords[$targetName] = [ordered]@{
            base_key = $baseKey
            image_id = [string]$actual.id
            platform = $platform
            rootfs_layers = $layers
            entrypoint = $entrypoint
            entrypoint_sha256 = Get-JcsArraySha256 -Value $entrypoint
            cmd = $cmd
            cmd_sha256 = Get-JcsArraySha256 -Value $cmd
            control_commit = $ControlCommit
            control_bundle_sha256 = [string]$control.BundleSha256
            source_commit = [string]$control.SourcePolicy.source_commit
            source_policy_sha256 = [string]$control.Blobs['testframework/docker/source-policy.v1.json'].Sha256
            source_tree_manifest_sha256 = $sourceTreeManifestSha
            build_manifest_sha256 = [string]$control.Blobs['testframework/docker/build-manifest.v2.json'].Sha256
            labels_projection = $labels
            labels_projection_sha256 = Get-JcsObjectSha256 -Value $labels
        }
    }

    $lock = [ordered]@{
        schema = 'hub.qa.local-images-lock/v3'
        control = [ordered]@{
            commit = $ControlCommit
            bundle_sha256 = [string]$control.BundleSha256
            compose_sha256 = [string]$control.ComposeSha256
        }
        source = [ordered]@{
            commit = [string]$control.SourcePolicy.source_commit
            policy_sha256 = [string]$control.Blobs['testframework/docker/source-policy.v1.json'].Sha256
            tree_manifest_sha256 = $sourceTreeManifestSha
        }
        build = [ordered]@{ manifest_sha256 = [string]$control.Blobs['testframework/docker/build-manifest.v2.json'].Sha256 }
        bases = $baseRecords
        targets = $targetRecords
    }
    Publish-NoOverwriteLock -Destination $lockPath -Content (($lock | ConvertTo-Json -Depth 32) + "`n")
    $publicationSucceeded = $true
} finally {
    if ($publicationSucceeded) {
        try {
            Remove-OwnedStagingAfterPublication -Reservation $stagingReservation -SourceEntries $sourceEntries -TargetNames $fixedTargetNames
        } catch {
            # A cleanup mismatch abandons the remaining staging object and does
            # not replace the successful lock-publication outcome.
            $cleanupFailureMessage = 'Owned staging cleanup was abandoned: {0}' -f $_.Exception.Message
            try { [Console]::Error.WriteLine($cleanupFailureMessage) } catch { }
        }
    }
}
