[CmdletBinding()]
param(
    [switch]$Describe
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Private hub.qa-bootstrap/v1 source contract. Package A0 only defines this
# closed future behavior. It never builds, invokes Docker, or publishes a
# repository lock. Package B supplies the internally validated context.

$script:Utf8 = [Text.UTF8Encoding]::new($false, $true)
$script:SourceBaseline = '7b5d5501cefab34f9aa5661e4f8c38e71474f9ce'
$script:BindingFiles = [ordered]@{
    dependency = 'testframework/docker/dependency-bundle.manifest.v1.json'
    source = 'testframework/docker/source-policy.v2.json'
    build = 'testframework/docker/build-manifest.v3.json'
}
$script:FixedTargets = [ordered]@{
    'backend-unit' = [ordered]@{
        dockerfile = 'testframework/docker/Dockerfile.backend-unit'
        entrypoint_path = '/opt/hub/bin/backend-unit'
        compose_service = 'backend-unit'
        local_image_lock_field = 'targets.backend-unit.local_image_id'
        docker_cmd = @('--hub-qa-target=backend-unit')
        target_action = @(
            'go build ./...'
            'go vet ./...'
            'go test -json -count=1 -timeout=150s ./tests/...'
            'go test -json -race -count=1 -timeout=150s ./tests/...'
        )
        bootstrap_selectable = $false
    }
    'frontend-unit' = [ordered]@{
        dockerfile = 'testframework/docker/Dockerfile.frontend-unit'
        entrypoint_path = '/opt/hub/bin/frontend-unit'
        compose_service = 'frontend-unit'
        local_image_lock_field = 'targets.frontend-unit.local_image_id'
        docker_cmd = @('--hub-qa-target=frontend-unit')
        target_action = @(
            '/opt/hub/frontend/node_modules/.bin/tsc -b'
            '/opt/hub/frontend/node_modules/.bin/biome check ./src'
            '/opt/hub/frontend/node_modules/.bin/vite build --outDir /reports/frontend-unit/vite-dist'
            '/opt/hub/frontend/node_modules/.bin/vitest run --config /opt/hub/config/vitest.frontend-unit.config.ts --reporter=json'
        )
        bootstrap_selectable = $false
    }
    'backend-openapi' = [ordered]@{
        dockerfile = 'testframework/docker/Dockerfile.backend-openapi'
        entrypoint_path = '/opt/hub/bin/backend-openapi'
        compose_service = 'backend-openapi'
        local_image_lock_field = 'targets.backend-openapi.local_image_id'
        docker_cmd = @('--hub-qa-target=backend-openapi')
        target_action = @(
            '/opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-1 --parseDependency --parseInternal'
            '/opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-2 --parseDependency --parseInternal'
            'sha256 manifest comparison and exact three-path synchronization'
        )
        bootstrap_selectable = $false
    }
    'g0-planning-selftest' = [ordered]@{
        dockerfile = 'testframework/docker/Dockerfile.g0-planning-selftest'
        entrypoint_path = '/opt/hub/bin/g0-planning-selftest'
        compose_service = 'g0-planning-selftest'
        local_image_lock_field = 'targets.g0-planning-selftest.local_image_id'
        docker_cmd = @('--hub-qa-target=g0-planning-selftest')
        target_action = @(
            'image-baked Python stdlib fixture imports the shared A0 verifier and admission modules directly'
            'fixed valid and negative dependency-bundle fixtures prove verifier reasons and no success receipt after rejection'
        )
        bootstrap_selectable = $true
    }
}
$script:CandidateChildKeys = [string[]]@(
    'local_image_id',
    'local_image_id_sha256',
    'entrypoint_path',
    'entrypoint_sha256',
    'command_sha256',
    'compose_service',
    'runtime_status'
)

class HubQaUtf8BytePathComparer : System.Collections.Generic.IComparer[string] {
    [int] Compare([string]$Left, [string]$Right) {
        if ([Object]::ReferenceEquals($Left, $Right)) { return 0 }
        if ($null -eq $Left) { return -1 }
        if ($null -eq $Right) { return 1 }
        [byte[]]$leftBytes = ([Text.UTF8Encoding]::new($false, $true)).GetBytes($Left)
        [byte[]]$rightBytes = ([Text.UTF8Encoding]::new($false, $true)).GetBytes($Right)
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

function Get-HubQaSha256 {
    param([byte[]]$Bytes)

    $algorithm = [Security.Cryptography.SHA256]::Create()
    try {
        return 'sha256:' + ([BitConverter]::ToString($algorithm.ComputeHash($Bytes))).Replace('-', '').ToLowerInvariant()
    } finally {
        $algorithm.Dispose()
    }
}

function Assert-HubQaDigest {
    param([string]$Value, [string]$Identity)

    if ($Value -notmatch '^sha256:[0-9a-f]{64}$') { throw "$Identity must be sha256 plus 64 lowercase hexadecimal characters." }
}

function Assert-HubQaSha1 {
    param([string]$Value, [string]$Identity)

    if ($Value -notmatch '^[0-9a-f]{40}$') { throw "$Identity must be 40 lowercase hexadecimal characters." }
}

function Resolve-HubQaCanonicalDirectory {
    param([string]$Path, [string]$Identity)

    $directory = [IO.DirectoryInfo]::new([IO.Path]::GetFullPath($Path))
    while ($null -ne $directory) {
        $directory.Refresh()
        if (-not $directory.Exists -or ($directory.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw "$Identity is missing, noncanonical, or contains a reparse point: $Path"
        }
        $directory = $directory.Parent
    }
    return [IO.Path]::GetFullPath($Path).TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar)
}

function Assert-HubQaContainedPath {
    param([string]$Path, [string]$Root, [string]$Identity)

    $fullPath = [IO.Path]::GetFullPath($Path)
    $fullRoot = [IO.Path]::GetFullPath($Root).TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar)
    $prefix = $fullRoot + [IO.Path]::DirectorySeparatorChar
    if (-not $fullPath.StartsWith($prefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw "$Identity escapes the closed root."
    }
    return $fullPath
}

function Read-HubQaRegularUtf8File {
    param([string]$Path, [string]$Root, [string]$Identity)

    $fullPath = Assert-HubQaContainedPath -Path $Path -Root $Root -Identity $Identity
    $file = [IO.FileInfo]::new($fullPath)
    $file.Refresh()
    if (-not $file.Exists -or ($file.Attributes -band [IO.FileAttributes]::Directory) -ne 0 -or ($file.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw "$Identity is not a regular file."
    }
    $parent = $file.Directory
    while ($null -ne $parent -and $parent.FullName.Length -ge $Root.Length) {
        $parent.Refresh()
        if (-not $parent.Exists -or ($parent.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw "$Identity has a reparse ancestor." }
        if ([StringComparer]::OrdinalIgnoreCase.Equals($parent.FullName, $Root)) { break }
        $parent = $parent.Parent
    }
    [byte[]]$bytes = [IO.File]::ReadAllBytes($fullPath)
    if ($bytes.Length -eq 0 -or ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf)) {
        throw "$Identity is empty or has a UTF-8 BOM."
    }
    try { $text = $script:Utf8.GetString($bytes) } catch { throw "$Identity is not strict UTF-8." }
    return [pscustomobject]@{ Path = $fullPath; Bytes = $bytes; Text = $text; Sha256 = Get-HubQaSha256 -Bytes $bytes }
}

function ConvertFrom-HubQaClosedJson {
    param([string]$Text, [string]$Identity)

    try { return $Text | ConvertFrom-Json -ErrorAction Stop } catch { throw "$Identity is not valid JSON." }
}

function Get-HubQaMember {
    param([object]$Object, [string]$Name, [string]$Identity)

    if ($null -eq $Object -or $null -eq $Object.PSObject.Properties[$Name]) { throw "$Identity is missing $Name." }
    return $Object.PSObject.Properties[$Name].Value
}

function Assert-HubQaExactPropertyNames {
    param([object]$Object, [string[]]$Expected, [string]$Identity)

    $actual = @($Object.PSObject.Properties.Name)
    if ($actual.Count -ne $Expected.Count) { throw "$Identity has an unexpected property count." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if (-not [StringComparer]::Ordinal.Equals([string]$actual[$index], $Expected[$index])) { throw "$Identity has a missing, extra, or reordered property." }
    }
}

function Assert-HubQaExactStringArray {
    param([object]$Actual, [string[]]$Expected, [string]$Identity)

    $values = @($Actual)
    if ($values.Count -ne $Expected.Count) { throw "$Identity has an unexpected array count." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if (-not [StringComparer]::Ordinal.Equals([string]$values[$index], $Expected[$index])) { throw "$Identity differs at index $index." }
    }
}

function ConvertTo-HubQaJcsString {
    param([string]$Value)

    if ($null -eq $Value) { throw 'JCS string cannot be null.' }
    $builder = New-Object Text.StringBuilder
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
        [void]$builder.Append([char]$code)
    }
    [void]$builder.Append('"')
    return $builder.ToString()
}

function Get-JcsArraySha256 {
    param([object]$Value)

    if ($null -eq $Value -or $Value -is [string]) { throw 'JCS array must be a non-null array of strings.' }
    $parts = New-Object Collections.Generic.List[string]
    foreach ($item in @($Value)) {
        if ($item -isnot [string]) { throw 'JCS array contains a non-string member.' }
        $parts.Add((ConvertTo-HubQaJcsString -Value ([string]$item)))
    }
    return Get-HubQaSha256 -Bytes $script:Utf8.GetBytes(('[' + ($parts -join ',') + ']'))
}

function Get-HubQaGitBlobBytes {
    param([string]$Repository, [string]$BlobSha1, [string]$Identity)

    Assert-HubQaSha1 -Value $BlobSha1 -Identity "$Identity Git blob SHA-1"
    if ($Repository.Contains('"')) { throw "$Identity Git repository path is unsafe." }
    $startInfo = New-Object Diagnostics.ProcessStartInfo
    $startInfo.FileName = 'git'
    $startInfo.WorkingDirectory = $Repository
    $startInfo.Arguments = ('cat-file blob {0}' -f $BlobSha1)
    $startInfo.UseShellExecute = $false
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $process = New-Object Diagnostics.Process
    $process.StartInfo = $startInfo
    if (-not $process.Start()) { throw "$Identity could not start Git blob reader." }
    $memory = New-Object IO.MemoryStream
    $buffer = New-Object byte[] 65536
    try {
        while (($read = $process.StandardOutput.BaseStream.Read($buffer, 0, $buffer.Length)) -gt 0) {
            $memory.Write($buffer, 0, $read)
        }
        $stderr = $process.StandardError.ReadToEnd()
        $process.WaitForExit()
        if ($process.ExitCode -ne 0) { throw "$Identity Git blob read failed: $stderr" }
        return $memory.ToArray()
    } finally {
        $memory.Dispose()
        $process.Dispose()
    }
}

function Get-HubQaGitTreeEntry {
    param([string]$Repository, [string]$RepositoryRelativePath)

    # This gate uses git ls-tree and git cat-file object data only; the
    # worktree and index (including the staged legacy deletions) are ignored.
    if ($RepositoryRelativePath -notmatch '^[A-Za-z0-9._/-]+$' -or $RepositoryRelativePath.StartsWith('/') -or $RepositoryRelativePath.Contains('..')) {
        throw 'E_SOURCE_TREE_INVALID: source-policy path is unsafe.'
    }
    $lines = @(& git -C $Repository ls-tree $script:SourceBaseline -- $RepositoryRelativePath 2>$null)
    if ($LASTEXITCODE -ne 0 -or $lines.Count -ne 1) { throw "E_SOURCE_TREE_INVALID: missing Git tree entry $RepositoryRelativePath." }
    if ([string]$lines[0] -notmatch '^(?<mode>[0-7]{6}) (?<type>blob|tree|commit) (?<object>[0-9a-f]{40})\t(?<path>.+)$') {
        throw "E_SOURCE_TREE_INVALID: malformed Git tree entry $RepositoryRelativePath."
    }
    if (-not [StringComparer]::Ordinal.Equals([string]$Matches.path, $RepositoryRelativePath)) {
        throw "E_SOURCE_TREE_INVALID: Git tree path mismatch $RepositoryRelativePath."
    }
    if ($Matches.mode -ne '100644' -or $Matches.type -ne 'blob') {
        throw "E_SOURCE_TREE_INVALID: source entry must be a mode 100644 blob $RepositoryRelativePath."
    }
    Assert-HubQaSha1 -Value ([string]$Matches.object) -Identity "Git tree object $RepositoryRelativePath"
    return [pscustomobject]@{ Mode = [string]$Matches.mode; BlobSha1 = [string]$Matches.object }
}

function Get-HubQaGitObjectSourceTree {
    [CmdletBinding()]
    param([object]$Controls)

    $policy = $Controls.source.Value
    Assert-HubQaExactPropertyNames -Object $policy -Expected @('schema', 'source_commit', 'source_tree_sha256', 'allowlist') -Identity 'source policy'
    if ($policy.schema -cne 'hub.qa.local-source-policy/v2' -or $policy.source_commit -cne $script:SourceBaseline) {
        throw 'E_SOURCE_TREE_INVALID: source policy baseline is not closed.'
    }
    Assert-HubQaDigest -Value ([string]$policy.source_tree_sha256) -Identity 'source policy source tree digest'
    if ($null -eq $policy.allowlist -or $policy.allowlist -is [string]) { throw 'E_SOURCE_TREE_INVALID: source policy allowlist is invalid.' }
    $root = Get-HubQaSourceRoot
    $commitType = @(& git -C $root cat-file -t $script:SourceBaseline 2>$null)
    if ($LASTEXITCODE -ne 0 -or $commitType.Count -ne 1 -or [string]$commitType[0] -ne 'commit') {
        throw 'E_SOURCE_TREE_INVALID: recovery source commit is unavailable.'
    }

    $seenPaths = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    $linesByPath = @{}
    foreach ($allowlistEntry in @($policy.allowlist)) {
        Assert-HubQaExactPropertyNames -Object $allowlistEntry -Expected @('path', 'blob_sha1') -Identity 'source policy allowlist entry'
        $path = [string]$allowlistEntry.path
        $declaredBlobSha1 = [string]$allowlistEntry.blob_sha1
        if (-not $seenPaths.Add($path)) { throw "E_SOURCE_TREE_INVALID: duplicate or unlisted source path $path." }
        Assert-HubQaSha1 -Value $declaredBlobSha1 -Identity "source policy blob SHA-1 $path"
        $treeEntry = Get-HubQaGitTreeEntry -Repository $root -RepositoryRelativePath $path
        if (-not [StringComparer]::Ordinal.Equals($treeEntry.BlobSha1, $declaredBlobSha1)) {
            throw "E_SOURCE_TREE_INVALID: Git blob SHA-1 mismatch $path."
        }
        [byte[]]$blobBytes = Get-HubQaGitBlobBytes -Repository $root -BlobSha1 $treeEntry.BlobSha1 -Identity "source policy blob $path"
        $line = $path + "`t100644`t" + $treeEntry.BlobSha1 + "`t" + (Get-HubQaSha256 -Bytes $blobBytes) + "`n"
        if ($linesByPath.ContainsKey($path)) { throw "E_SOURCE_TREE_INVALID: UTF-8 source path collision $path." }
        $linesByPath[$path] = $line
    }
    if ($linesByPath.Count -eq 0) { throw 'E_SOURCE_TREE_INVALID: source policy allowlist is empty.' }
    $orderedPaths = [string[]]@($linesByPath.Keys)
    [Array]::Sort($orderedPaths, [HubQaUtf8BytePathComparer]::new())
    $canonical = New-Object Text.StringBuilder
    foreach ($path in $orderedPaths) { [void]$canonical.Append([string]$linesByPath[$path]) }
    $computed = Get-HubQaSha256 -Bytes $script:Utf8.GetBytes($canonical.ToString())
    if ($computed -cne [string]$policy.source_tree_sha256 -or $computed -cne [string]$Controls.build.Value.source_tree_sha256) {
        throw 'E_SOURCE_TREE_DIGEST_MISMATCH: canonical Git object source tree digest disagrees with A1 controls.'
    }
    return [pscustomobject]@{ SourceTreeSha256 = $computed; SourceCommit = $script:SourceBaseline; Entries = $linesByPath.Count }
}

function Get-HubQaSourceRoot {
    $dockerRoot = Resolve-HubQaCanonicalDirectory -Path $PSScriptRoot -Identity 'source-image-attestor directory'
    return Resolve-HubQaCanonicalDirectory -Path (Join-Path $dockerRoot '..\..') -Identity 'repository root derived from source-image-attestor'
}

function Get-HubQaBootstrapImageContract {
    [CmdletBinding()]
    param()

    return [ordered]@{
        schema = 'hub.qa.bootstrap-image-attestation/v1'
        status = 'NOT_RUN'
        execution_authority = 'hub.qa-bootstrap/v1'
        source_baseline = $script:SourceBaseline
        platform = [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }
        dependency_binding = [ordered]@{
            required_files = @($script:BindingFiles.Values)
            missing_or_unbound_verdict = 'NOT_RUN'
            publication_before_binding = 'forbidden'
        }
        candidate_lock = [ordered]@{
            template = 'testframework/reports/<task-run>/bootstrap/images.lock.v3.json'
            repository_publication = 'forbidden'
            create_semantics = 'create-new-after-four-runner-produced-local-inspections'
        }
        targets = $script:FixedTargets
    }
}

function Assert-HubQaExactBoolean {
    param([object]$Actual, [bool]$Expected, [string]$Identity)

    if ($Actual -isnot [bool] -or [bool]$Actual -ne $Expected) {
        throw "$Identity must be the exact Boolean value $Expected."
    }
}

function Read-HubQaBootstrapControl {
    param([string]$Root)

    $controlPath = Join-Path $Root 'testframework\docker\bootstrap-control.v1.json'
    $raw = Read-HubQaRegularUtf8File -Path $controlPath -Root $Root -Identity 'fixed bootstrap control'
    $control = ConvertFrom-HubQaClosedJson -Text $raw.Text -Identity 'fixed bootstrap control'
    Assert-HubQaExactPropertyNames -Object $control -Expected @('schema', 'status', 'bootstrap_contract', 'platform', 'candidate_lock', 'runtime', 'targets') -Identity 'fixed bootstrap control'
    if ($control.schema -cne 'hub.qa.bootstrap-control/v1' -or $control.status -cne 'NOT_RUN' -or $control.bootstrap_contract -cne 'hub.qa-bootstrap/v1') {
        throw 'Fixed bootstrap control identity is not closed.'
    }
    Assert-HubQaExactPropertyNames -Object $control.platform -Expected @('os', 'architecture', 'variant') -Identity 'fixed bootstrap control platform'
    if ($control.platform.os -cne 'linux' -or $control.platform.architecture -cne 'amd64' -or $null -ne $control.platform.variant) {
        throw 'Fixed bootstrap control platform is not linux/amd64.'
    }
    Assert-HubQaExactPropertyNames -Object $control.candidate_lock -Expected @('path_template', 'repository_publication', 'runtime_status') -Identity 'fixed bootstrap control candidate lock'
    if ($control.candidate_lock.path_template -cne 'testframework/reports/<task-run>/bootstrap/images.lock.v3.json' -or
        $control.candidate_lock.repository_publication -cne 'forbidden' -or
        $control.candidate_lock.runtime_status -cne 'READY_FOR_G0_BOOTSTRAP_ONLY') {
        throw 'Fixed bootstrap control candidate lock contract is invalid.'
    }
    Assert-HubQaExactPropertyNames -Object $control.runtime -Expected @('network_mode', 'read_only', 'cpus', 'mem_limit', 'pids_limit', 'cap_drop', 'no_new_privileges', 'pull_policy', 'only_writable_mount') -Identity 'fixed bootstrap control runtime'
    if ($control.runtime.network_mode -cne 'none' -or $control.runtime.read_only -isnot [bool] -or -not [bool]$control.runtime.read_only -or
        [double]$control.runtime.cpus -ne 2.0 -or $control.runtime.mem_limit -cne '2g' -or [int]$control.runtime.pids_limit -ne 256 -or
        $control.runtime.no_new_privileges -isnot [bool] -or -not [bool]$control.runtime.no_new_privileges -or
        $control.runtime.pull_policy -cne 'never' -or $control.runtime.only_writable_mount -cne '/reports') {
        throw 'Fixed bootstrap control runtime contract is invalid.'
    }
    Assert-HubQaExactStringArray -Actual $control.runtime.cap_drop -Expected @('ALL') -Identity 'fixed bootstrap control capability drop'
    if (@($control.targets.PSObject.Properties.Name).Count -ne $script:FixedTargets.Count) {
        throw 'Fixed bootstrap control target count is not closed.'
    }
    foreach ($targetName in $script:FixedTargets.Keys) {
        $expected = $script:FixedTargets[$targetName]
        $property = $control.targets.PSObject.Properties[$targetName]
        if ($null -eq $property) { throw "Fixed bootstrap control lacks $targetName." }
        $target = $property.Value
        Assert-HubQaExactPropertyNames -Object $target -Expected @('dockerfile', 'entrypoint_path', 'compose_service', 'local_image_lock_field', 'docker_cmd', 'target_action', 'bootstrap_selectable') -Identity "fixed bootstrap control target $targetName"
        if ($target.dockerfile -cne $expected.dockerfile -or $target.entrypoint_path -cne $expected.entrypoint_path -or
            $target.compose_service -cne $expected.compose_service -or $target.local_image_lock_field -cne $expected.local_image_lock_field) {
            throw "Fixed bootstrap control target mapping differs: $targetName"
        }
        Assert-HubQaExactStringArray -Actual $target.docker_cmd -Expected @($expected.docker_cmd) -Identity "fixed bootstrap control Docker Cmd $targetName"
        Assert-HubQaExactStringArray -Actual $target.target_action -Expected @($expected.target_action) -Identity "fixed bootstrap control target action $targetName"
        Assert-HubQaExactBoolean -Actual $target.bootstrap_selectable -Expected ([bool]$expected.bootstrap_selectable) -Identity "fixed bootstrap control bootstrap selection $targetName"
    }
    if ($raw.Text.IndexOf('g0-admission-helper', [StringComparison]::Ordinal) -ge 0) {
        throw 'Fixed bootstrap control must exclude the ephemeral admission helper.'
    }
    return [pscustomobject]@{ Raw = $raw; Value = $control }
}

function Read-HubQaBoundControls {
    param([string]$Root)

    $records = [ordered]@{}
    foreach ($key in $script:BindingFiles.Keys) {
        $relativePath = [string]$script:BindingFiles[$key]
        $path = Join-Path $Root $relativePath
        if (-not [IO.File]::Exists($path)) { return $null }
        $record = Read-HubQaRegularUtf8File -Path $path -Root $Root -Identity "A1 $key control"
        $records[$key] = [pscustomobject]@{ Raw = $record; Value = ConvertFrom-HubQaClosedJson -Text $record.Text -Identity "A1 $key control" }
    }
    return $records
}

function Assert-HubQaBoundControlContract {
    param([object]$Controls, [object]$BootstrapControl)

    $dependency = $Controls.dependency.Value
    $source = $Controls.source.Value
    $build = $Controls.build.Value
    Assert-HubQaExactPropertyNames -Object $source -Expected @('schema', 'source_commit', 'source_tree_sha256', 'allowlist') -Identity 'source policy'
    Assert-HubQaExactPropertyNames -Object $build -Expected @('schema', 'source_commit', 'source_tree_sha256', 'source_policy_sha256', 'dependency_bundle_sha256', 'platform', 'targets') -Identity 'build manifest'
    if ($source.schema -cne 'hub.qa.local-source-policy/v2' -or $build.schema -cne 'hub.qa.local-images-build-manifest/v3') {
        throw 'A1 source or build control schema mismatch.'
    }
    if ($source.source_commit -cne $script:SourceBaseline -or $build.source_commit -cne $script:SourceBaseline) {
        throw 'A1 controls do not bind the immutable recovery source baseline.'
    }
    Assert-HubQaDigest -Value ([string]$source.source_tree_sha256) -Identity 'source policy tree digest'
    Assert-HubQaDigest -Value ([string]$build.source_tree_sha256) -Identity 'build manifest tree digest'
    Assert-HubQaDigest -Value ([string]$build.source_policy_sha256) -Identity 'build manifest source policy digest'
    Assert-HubQaDigest -Value ([string]$build.dependency_bundle_sha256) -Identity 'build manifest dependency bundle digest'
    if ($build.source_tree_sha256 -cne $source.source_tree_sha256 -or $build.source_policy_sha256 -cne $Controls.source.Raw.Sha256 -or $build.dependency_bundle_sha256 -cne $Controls.dependency.Raw.Sha256) {
        throw 'A1 raw control bytes or source-tree bindings disagree.'
    }
    if ($dependency.schema -cne 'hub.qa.dependency-bundle/v1' -or $dependency.source_commit -cne $script:SourceBaseline) {
        throw 'Dependency manifest source baseline mismatch.'
    }
    Assert-HubQaExactPropertyNames -Object $build.platform -Expected @('os', 'architecture', 'variant') -Identity 'build manifest platform'
    if ($build.platform.os -cne 'linux' -or $build.platform.architecture -cne 'amd64' -or $null -ne $build.platform.variant) {
        throw 'Build manifest platform mismatch.'
    }
    if (@($build.targets.PSObject.Properties.Name).Count -ne $script:FixedTargets.Count) { throw 'Build manifest target count is not closed.' }
    foreach ($targetName in $script:FixedTargets.Keys) {
        if ($null -eq $build.targets.PSObject.Properties[$targetName]) { throw "Build manifest lacks $targetName." }
        $target = $build.targets.PSObject.Properties[$targetName].Value
        $controlTarget = Get-HubQaMember -Object $BootstrapControl.Value.targets -Name $targetName -Identity 'fixed bootstrap control target'
        $expected = $script:FixedTargets[$targetName]
        Assert-HubQaExactPropertyNames -Object $target -Expected @('dockerfile', 'dockerfile_sha256', 'entrypoint_path', 'entrypoint_sha256', 'docker_entrypoint', 'docker_cmd', 'compose_service', 'local_image_lock_field', 'target_action', 'bootstrap_selectable') -Identity "build manifest target $targetName"
        Assert-HubQaDigest -Value ([string]$target.dockerfile_sha256) -Identity "build manifest Dockerfile digest $targetName"
        Assert-HubQaDigest -Value ([string]$target.entrypoint_sha256) -Identity "build manifest entrypoint digest $targetName"
        if ($target.dockerfile -cne $expected.dockerfile -or $target.entrypoint_path -cne $expected.entrypoint_path -or
            $target.compose_service -cne $expected.compose_service -or $target.local_image_lock_field -cne $expected.local_image_lock_field -or
            $target.dockerfile -cne $controlTarget.dockerfile -or $target.entrypoint_path -cne $controlTarget.entrypoint_path -or
            $target.compose_service -cne $controlTarget.compose_service -or $target.local_image_lock_field -cne $controlTarget.local_image_lock_field) {
            throw "Build manifest target mapping is not bound to fixed control: $targetName"
        }
        Assert-HubQaExactStringArray -Actual $target.docker_entrypoint -Expected @([string]$expected.entrypoint_path) -Identity "build manifest Docker Entrypoint $targetName"
        Assert-HubQaExactStringArray -Actual $target.docker_cmd -Expected @($expected.docker_cmd) -Identity "build manifest Docker Cmd $targetName"
        Assert-HubQaExactStringArray -Actual $target.target_action -Expected @($expected.target_action) -Identity "build manifest target action $targetName"
        Assert-HubQaExactStringArray -Actual $target.docker_entrypoint -Expected @($controlTarget.entrypoint_path) -Identity "build manifest control Docker Entrypoint $targetName"
        Assert-HubQaExactStringArray -Actual $target.docker_cmd -Expected @($controlTarget.docker_cmd) -Identity "build manifest control Docker Cmd $targetName"
        Assert-HubQaExactStringArray -Actual $target.target_action -Expected @($controlTarget.target_action) -Identity "build manifest control target action $targetName"
        Assert-HubQaExactBoolean -Actual $target.bootstrap_selectable -Expected ([bool]$expected.bootstrap_selectable) -Identity "build manifest bootstrap selection $targetName"
        Assert-HubQaExactBoolean -Actual $target.bootstrap_selectable -Expected ([bool]$controlTarget.bootstrap_selectable) -Identity "build manifest control bootstrap selection $targetName"
    }
}

function Test-HubQaBootstrapImageBindings {
    [CmdletBinding()]
    param()

    $root = Get-HubQaSourceRoot
    $controls = Read-HubQaBoundControls -Root $root
    if ($null -eq $controls) {
        return [pscustomobject]@{ Verdict = 'NOT_RUN'; ReasonCode = 'E_BINDING_DATA_UNAVAILABLE'; Controls = $null }
    }
    try {
        $bootstrapControl = Read-HubQaBootstrapControl -Root $root
        Assert-HubQaBoundControlContract -Controls $controls -BootstrapControl $bootstrapControl
        return [pscustomobject]@{ Verdict = 'READY_FOR_G0_BOOTSTRAP_ONLY'; ReasonCode = $null; Controls = $controls; BootstrapControl = $bootstrapControl }
    } catch {
        return [pscustomobject]@{ Verdict = 'NOT_RUN'; ReasonCode = 'E_A1_CONTROL_BINDING_INVALID'; Controls = $null }
    }
}

function Test-HubQaByteArrayEqual {
    param([byte[]]$Left, [byte[]]$Right)

    if ($Left.Length -ne $Right.Length) { return $false }
    for ($index = 0; $index -lt $Left.Length; $index++) {
        if ($Left[$index] -ne $Right[$index]) { return $false }
    }
    return $true
}

function Assert-HubQaPrivateBootstrapContext {
    param([hashtable]$BootstrapPrivateContext)

    $expectedKeys = [string[]]@('Task', 'Run', 'RoleClaim', 'OwnerNonce', 'OwnerMarkerRawBytes', 'OwnerMarkerSha256', 'RunnerInspectionRawBytes', 'RunnerInspectionSha256')
    if ($BootstrapPrivateContext.Count -ne $expectedKeys.Count) { throw 'Runner-private bootstrap context has an unexpected field count.' }
    foreach ($name in $expectedKeys) {
        if (-not $BootstrapPrivateContext.ContainsKey($name)) { throw "Runner-private bootstrap context lacks $name." }
    }
    if ([string]$BootstrapPrivateContext.Task -cne 'g0-assets-runner') { throw 'Runner-private bootstrap context task is not closed.' }
    if ([string]$BootstrapPrivateContext.Run -notmatch '^[a-z0-9][a-z0-9-]{7,63}$') { throw 'Runner-private bootstrap context run grammar is invalid.' }
    if ([string]$BootstrapPrivateContext.RoleClaim -cne 'hub.qa-bootstrap/v1') { throw 'Runner-private bootstrap context role claim is invalid.' }
    if ([string]$BootstrapPrivateContext.OwnerNonce -notmatch '^[0-9a-f]{32}$') { throw 'Runner-private bootstrap context owner nonce is invalid.' }
    if ($BootstrapPrivateContext.OwnerMarkerRawBytes -isnot [byte[]] -or $BootstrapPrivateContext.RunnerInspectionRawBytes -isnot [byte[]]) {
        throw 'Runner-private bootstrap context must retain raw receipt and owner-marker bytes.'
    }
    Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext.OwnerMarkerSha256) -Identity 'runner-private owner marker digest'
    Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext.RunnerInspectionSha256) -Identity 'runner-private inspection receipt digest'
    if ((Get-HubQaSha256 -Bytes $BootstrapPrivateContext.OwnerMarkerRawBytes) -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.RunnerInspectionRawBytes) -cne [string]$BootstrapPrivateContext.RunnerInspectionSha256) {
        throw 'Runner-private bootstrap context raw-byte digest binding is invalid.'
    }
}

function Get-HubQaBootstrapContextPaths {
    param([hashtable]$BootstrapPrivateContext, [string]$Root)

    Assert-HubQaPrivateBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext
    $reportsRoot = Assert-HubQaContainedPath -Path (Join-Path $Root 'testframework\reports') -Root $Root -Identity 'reports root'
    $taskRunName = [string]$BootstrapPrivateContext.Task + '-' + [string]$BootstrapPrivateContext.Run
    $taskRunRoot = Assert-HubQaContainedPath -Path (Join-Path $reportsRoot $taskRunName) -Root $reportsRoot -Identity 'task-owned report root'
    $bootstrapRoot = Assert-HubQaContainedPath -Path (Join-Path $taskRunRoot 'bootstrap') -Root $taskRunRoot -Identity 'task-owned bootstrap root'
    return [pscustomobject]@{
        ReportsRoot = $reportsRoot
        TaskRunRoot = $taskRunRoot
        BootstrapRoot = $bootstrapRoot
        OwnerMarkerPath = Join-Path $bootstrapRoot 'owner-marker.v1.json'
        InspectionPath = Join-Path $bootstrapRoot 'local-image-inspect.v1.json'
        CandidatePath = Join-Path $bootstrapRoot 'images.lock.v3.json'
        ReservationPath = Join-Path $bootstrapRoot ('.image-attestor-reservation-' + [string]$BootstrapPrivateContext.OwnerNonce + '.json')
        StagingRoot = Join-Path $bootstrapRoot ('.image-attestor-staging-' + [string]$BootstrapPrivateContext.OwnerNonce)
    }
}

function Assert-HubQaRunnerOwnedBootstrapRoot {
    param([pscustomobject]$Paths, [hashtable]$BootstrapPrivateContext)

    # The runner creates this exact task-run report namespace before it calls
    # the attestor.  The attestor must only consume it; it must never adopt or
    # create a report or bootstrap root on the runner's behalf.
    [void](Resolve-HubQaCanonicalDirectory -Path $Paths.ReportsRoot -Identity 'pre-existing reports root')
    [void](Resolve-HubQaCanonicalDirectory -Path $Paths.TaskRunRoot -Identity 'pre-existing task-owned report root')
    [void](Resolve-HubQaCanonicalDirectory -Path $Paths.BootstrapRoot -Identity 'pre-existing runner-owned bootstrap root')
    $ownerMarker = Read-HubQaRegularUtf8File -Path $Paths.OwnerMarkerPath -Root $Paths.TaskRunRoot -Identity 'runner-owned bootstrap owner marker'
    if (-not (Test-HubQaByteArrayEqual -Left $ownerMarker.Bytes -Right $BootstrapPrivateContext.OwnerMarkerRawBytes) -or
        $ownerMarker.Sha256 -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256) {
        throw 'Runner-owned bootstrap owner marker does not match the held private context.'
    }
    $marker = ConvertFrom-HubQaClosedJson -Text $ownerMarker.Text -Identity 'runner-owned bootstrap owner marker'
    Assert-HubQaExactPropertyNames -Object $marker -Expected @('schema', 'task', 'run', 'owner_nonce') -Identity 'runner-owned bootstrap owner marker'
    if ($marker.schema -cne 'hub.qa.bootstrap-owner/v1' -or
        $marker.task -cne $BootstrapPrivateContext.Task -or
        $marker.run -cne $BootstrapPrivateContext.Run -or
        $marker.owner_nonce -cne $BootstrapPrivateContext.OwnerNonce) {
        throw 'Runner-owned bootstrap owner marker identity is not closed.'
    }
    return $ownerMarker
}

function Read-HubQaRunnerInspection {
    param([pscustomobject]$Paths, [hashtable]$BootstrapPrivateContext, [object]$Controls)

    $raw = Read-HubQaRegularUtf8File -Path $Paths.InspectionPath -Root $Paths.TaskRunRoot -Identity 'runner-produced local image inspection receipt'
    if (-not (Test-HubQaByteArrayEqual -Left $raw.Bytes -Right $BootstrapPrivateContext.RunnerInspectionRawBytes) -or
        $raw.Sha256 -cne [string]$BootstrapPrivateContext.RunnerInspectionSha256) {
        throw 'Runner inspection receipt is not the exact raw receipt held by the private bootstrap context.'
    }
    $receipt = ConvertFrom-HubQaClosedJson -Text $raw.Text -Identity 'runner-produced local image inspection receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @('schema', 'producer', 'task', 'run', 'source_commit', 'owner_marker_sha256', 'receipt_nonce', 'controls', 'targets') -Identity 'runner inspection receipt'
    if ($receipt.schema -cne 'hub.qa.local-image-inspect/v1' -or $receipt.producer -cne 'hub.qa-bootstrap/v1' -or $receipt.task -cne $BootstrapPrivateContext.Task -or $receipt.run -cne $BootstrapPrivateContext.Run -or $receipt.source_commit -cne $script:SourceBaseline -or
        $receipt.owner_marker_sha256 -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256 -or $receipt.receipt_nonce -cne $BootstrapPrivateContext.OwnerNonce) {
        throw 'Runner inspection receipt identity is not closed.'
    }
    Assert-HubQaExactPropertyNames -Object $receipt.controls -Expected @('dependency_bundle_sha256', 'source_policy_sha256', 'build_manifest_sha256') -Identity 'runner inspection controls'
    if ($receipt.controls.dependency_bundle_sha256 -cne $Controls.dependency.Raw.Sha256 -or $receipt.controls.source_policy_sha256 -cne $Controls.source.Raw.Sha256 -or $receipt.controls.build_manifest_sha256 -cne $Controls.build.Raw.Sha256) {
        throw 'Runner inspection receipt control bytes do not bind fixed A1 controls.'
    }
    if (@($receipt.targets.PSObject.Properties.Name).Count -ne $script:FixedTargets.Count) { throw 'Runner inspection target count is not closed.' }
    foreach ($targetName in $script:FixedTargets.Keys) {
        if ($null -eq $receipt.targets.PSObject.Properties[$targetName]) { throw "Runner inspection lacks $targetName." }
    }
    return $receipt
}

function Get-HubQaTargetRecord {
    param([string]$TargetName, [object]$Inspection, [object]$Controls, [object]$BootstrapControl, [string]$Root)

    $expected = $script:FixedTargets[$TargetName]
    $controlTarget = Get-HubQaMember -Object $BootstrapControl.Value.targets -Name $TargetName -Identity 'fixed bootstrap control target'
    $actual = $Inspection.targets.PSObject.Properties[$TargetName].Value
    Assert-HubQaExactPropertyNames -Object $actual -Expected @('local_image_id', 'platform', 'entrypoint', 'cmd', 'labels') -Identity "runner inspection $TargetName"
    Assert-HubQaExactPropertyNames -Object $actual.platform -Expected @('os', 'architecture', 'variant') -Identity "runner inspection platform $TargetName"
    if ($actual.platform.os -cne 'linux' -or $actual.platform.architecture -cne 'amd64' -or $null -ne $actual.platform.variant) { throw "Runner inspection platform mismatch: $TargetName" }
    Assert-HubQaDigest -Value ([string]$actual.local_image_id) -Identity "Local image ID $TargetName"
    Assert-HubQaExactStringArray -Actual $actual.entrypoint -Expected @([string]$expected.entrypoint_path) -Identity "Entrypoint $TargetName"
    Assert-HubQaExactStringArray -Actual $actual.cmd -Expected @($expected.docker_cmd) -Identity "Docker Cmd $TargetName"
    Assert-HubQaExactPropertyNames -Object $actual.labels -Expected @('hub.qa.target', 'hub.qa.execution-authority', 'hub.qa.network', 'hub.qa.runtime-status') -Identity "runner inspection labels $TargetName"
    if ($null -eq $actual.labels.PSObject.Properties['hub.qa.target'] -or $actual.labels.'hub.qa.target' -cne $TargetName) { throw "Target label mismatch: $TargetName" }
    if ($null -eq $actual.labels.PSObject.Properties['hub.qa.execution-authority'] -or $actual.labels.'hub.qa.execution-authority' -cne 'hub.qa-bootstrap/v1') { throw "Execution authority label mismatch: $TargetName" }
    if ($actual.labels.'hub.qa.network' -cne 'none' -or $actual.labels.'hub.qa.runtime-status' -cne 'NOT_RUN') { throw "Target runtime label mismatch: $TargetName" }

    $buildTarget = Get-HubQaMember -Object $Controls.build.Value.targets -Name $TargetName -Identity 'build manifest target'
    if ($buildTarget.dockerfile -cne $expected.dockerfile -or $buildTarget.entrypoint_path -cne $expected.entrypoint_path -or
        $buildTarget.compose_service -cne $expected.compose_service -or $buildTarget.local_image_lock_field -cne $expected.local_image_lock_field -or
        $buildTarget.dockerfile -cne $controlTarget.dockerfile -or $buildTarget.entrypoint_path -cne $controlTarget.entrypoint_path -or
        $buildTarget.compose_service -cne $controlTarget.compose_service -or $buildTarget.local_image_lock_field -cne $controlTarget.local_image_lock_field) {
        throw "Build manifest target binding mismatch: $TargetName"
    }
    Assert-HubQaExactStringArray -Actual $buildTarget.docker_entrypoint -Expected @($expected.entrypoint_path) -Identity "Build manifest Docker Entrypoint $TargetName"
    Assert-HubQaExactStringArray -Actual $buildTarget.docker_cmd -Expected @($expected.docker_cmd) -Identity "Build manifest Docker Cmd $TargetName"
    Assert-HubQaExactStringArray -Actual $buildTarget.target_action -Expected @($expected.target_action) -Identity "Build manifest target action $TargetName"
    Assert-HubQaExactStringArray -Actual $buildTarget.docker_entrypoint -Expected @($controlTarget.entrypoint_path) -Identity "Build manifest control Docker Entrypoint $TargetName"
    Assert-HubQaExactStringArray -Actual $buildTarget.docker_cmd -Expected @($controlTarget.docker_cmd) -Identity "Build manifest control Docker Cmd $TargetName"
    Assert-HubQaExactStringArray -Actual $buildTarget.target_action -Expected @($controlTarget.target_action) -Identity "Build manifest control target action $TargetName"
    Assert-HubQaExactBoolean -Actual $buildTarget.bootstrap_selectable -Expected ([bool]$expected.bootstrap_selectable) -Identity "Build manifest bootstrap selection $TargetName"
    Assert-HubQaExactBoolean -Actual $buildTarget.bootstrap_selectable -Expected ([bool]$controlTarget.bootstrap_selectable) -Identity "Build manifest control bootstrap selection $TargetName"
    Assert-HubQaDigest -Value ([string]$buildTarget.dockerfile_sha256) -Identity "Build manifest Dockerfile digest $TargetName"
    Assert-HubQaDigest -Value ([string]$buildTarget.entrypoint_sha256) -Identity "Build manifest entrypoint digest $TargetName"
    $dockerBytes = Read-HubQaRegularUtf8File -Path (Join-Path $Root $expected.dockerfile) -Root $Root -Identity "Dockerfile $TargetName"
    $entrypointRelative = 'testframework/docker/entrypoints/' + $TargetName + '.sh'
    $entrypointBytes = Read-HubQaRegularUtf8File -Path (Join-Path $Root $entrypointRelative) -Root $Root -Identity "Entrypoint $TargetName"
    if ($dockerBytes.Sha256 -cne $buildTarget.dockerfile_sha256 -or $entrypointBytes.Sha256 -cne $buildTarget.entrypoint_sha256) { throw "Build manifest source byte binding mismatch: $TargetName" }
    if ((Get-JcsArraySha256 -Value $actual.entrypoint) -cne (Get-JcsArraySha256 -Value $buildTarget.docker_entrypoint) -or (Get-JcsArraySha256 -Value $actual.cmd) -cne (Get-JcsArraySha256 -Value $buildTarget.docker_cmd)) {
        throw "Docker inspection JCS command binding mismatch: $TargetName"
    }

    return [ordered]@{
        local_image_id = [string]$actual.local_image_id
        local_image_id_sha256 = Get-HubQaSha256 -Bytes $script:Utf8.GetBytes([string]$actual.local_image_id)
        entrypoint_path = [string]$expected.entrypoint_path
        entrypoint_sha256 = $entrypointBytes.Sha256
        command_sha256 = Get-JcsArraySha256 -Value $actual.cmd
        compose_service = [string]$expected.compose_service
        runtime_status = 'READY_FOR_G0_BOOTSTRAP_ONLY'
    }
}

function New-HubQaCreateNewFile {
    param([string]$Path, [byte[]]$Bytes, [string]$Identity)

    $stream = [IO.File]::Open($Path, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
    try {
        $stream.Write($Bytes, 0, $Bytes.Length)
        $stream.Flush($true)
    } catch {
        throw "$Identity could not be written with create-new semantics."
    } finally {
        $stream.Dispose()
    }
}

function New-HubQaCandidateReservation {
    param([pscustomobject]$Paths, [hashtable]$BootstrapPrivateContext)

    if ([IO.File]::Exists($Paths.CandidatePath) -or [IO.Directory]::Exists($Paths.CandidatePath)) { throw 'Candidate lock destination already exists.' }
    if ([IO.File]::Exists($Paths.ReservationPath) -or [IO.Directory]::Exists($Paths.ReservationPath) -or [IO.Directory]::Exists($Paths.StagingRoot)) { throw 'Task-owned image attestor reservation already exists.' }
    $reservation = ('{"schema":"hub.qa.image-attestor-reservation/v1","task":"' + [string]$BootstrapPrivateContext.Task + '","run":"' + [string]$BootstrapPrivateContext.Run + '","owner_nonce":"' + [string]$BootstrapPrivateContext.OwnerNonce + '","owner_marker_sha256":"' + [string]$BootstrapPrivateContext.OwnerMarkerSha256 + '"}')
    New-HubQaCreateNewFile -Path $Paths.ReservationPath -Bytes $script:Utf8.GetBytes($reservation) -Identity 'image attestor reservation'
    try {
        [IO.Directory]::CreateDirectory($Paths.StagingRoot) | Out-Null
        $markerPath = Join-Path $Paths.StagingRoot '.owner.json'
        $marker = ('{"schema":"hub.qa.image-attestor-owner/v1","reservation_sha256":"' + (Get-HubQaSha256 -Bytes $script:Utf8.GetBytes($reservation)) + '"}')
        New-HubQaCreateNewFile -Path $markerPath -Bytes $script:Utf8.GetBytes($marker) -Identity 'image attestor owner marker'
        return [pscustomobject]@{ MarkerPath = $markerPath; StagedCandidatePath = Join-Path $Paths.StagingRoot 'images.lock.v3.json' }
    } catch { throw 'Task-owned image attestor staging creation failed after reservation.' }
}

function Invoke-HubQaBootstrapImageAttestation {
    [CmdletBinding()]
    param([hashtable]$BootstrapPrivateContext)

    $binding = Test-HubQaBootstrapImageBindings
    if ($binding.Verdict -ne 'READY_FOR_G0_BOOTSTRAP_ONLY') { return $binding }
    Assert-HubQaPrivateBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext
    $root = Get-HubQaSourceRoot
    $paths = Get-HubQaBootstrapContextPaths -BootstrapPrivateContext $BootstrapPrivateContext -Root $root
    # Before the first reservation or directory/file creation, consume the
    # runner-created physical report root and its raw-byte owner marker. This
    # closes the task/run namespace and excludes reparse or foreign roots.
    [void](Assert-HubQaRunnerOwnedBootstrapRoot -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext)
    # Reserve with create-new semantics only after runner ownership is closed;
    # then consume the nonce/digest-bound local inspection receipt.
    $reservation = New-HubQaCandidateReservation -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext
    # Package B invokes this same Git-object-only gate before its first image
    # build.  Repeating it here makes candidate publication fail closed if the
    # committed recovery source tree or either A1 digest binding has drifted.
    $sourceTree = Get-HubQaGitObjectSourceTree -Controls $binding.Controls
    $inspection = Read-HubQaRunnerInspection -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls
    $targets = [ordered]@{}
    foreach ($targetName in $script:FixedTargets.Keys) {
        $targets[$targetName] = Get-HubQaTargetRecord -TargetName $targetName -Inspection $inspection -Controls $binding.Controls -BootstrapControl $binding.BootstrapControl -Root $root
    }
    $candidate = [ordered]@{
        schema = 'hub.qa.local-images-lock/v3'
        platform = [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }
        targets = $targets
    }
    $candidateBytes = $script:Utf8.GetBytes(($candidate | ConvertTo-Json -Depth 16 -Compress) + "`n")
    New-HubQaCreateNewFile -Path $reservation.StagedCandidatePath -Bytes $candidateBytes -Identity 'staged candidate lock'
    [IO.File]::Move($reservation.StagedCandidatePath, $paths.CandidatePath)
    return [pscustomobject]@{
        Verdict = 'READY_FOR_G0_BOOTSTRAP_ONLY'
        ReasonCode = $null
        CandidateLockPath = $paths.CandidatePath
        CandidateLockSha256 = Get-HubQaSha256 -Bytes $candidateBytes
        OwnerMarkerPath = $reservation.MarkerPath
        SourceTreeSha256 = $sourceTree.SourceTreeSha256
    }
}

if ($Describe -or -not $PSBoundParameters.ContainsKey('Describe')) {
    Get-HubQaBootstrapImageContract | ConvertTo-Json -Depth 16
}
