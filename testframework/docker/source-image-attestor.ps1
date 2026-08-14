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
$script:SourceBaseline = '61b38573682d7c01e6d9381babc5085537d7990f'
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
            'go test -tags hubtest -json -count=1 -timeout=150s ./tests/...'
            'go test -tags hubtest -json -race -count=1 -timeout=150s ./tests/...'
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
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/typescript/bin/tsc -b'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/@biomejs/biome/bin/biome check ./src'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/vite/bin/vite.js build --outDir /reports/frontend-unit/vite-dist'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/vitest/vitest.mjs run --config /opt/hub/config/vitest.frontend-unit.config.ts --reporter=json'
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

function ConvertFrom-HubQaStrictCanonicalJson {
    param([string]$Text, [string]$Identity)

    # This is deliberately a private PS5.1 parser.  Ordinary PowerShell JSON
    # conversion coalesces duplicate keys and cannot establish receipt or A1
    # control authority.  The parser accepts no insignificant whitespace and
    # preserves object-member order for the closed-schema checks below.
    if ([string]::IsNullOrEmpty($Text)) { throw "$Identity is empty." }
    $cursor = 0
    $length = $Text.Length
    $allowBootstrapControlFormatting = $Identity -ceq 'fixed bootstrap control'
    $parseValue = $null
    $parseArray = $null
    $parseObject = $null
    $skipWhitespace = {
        if (-not $allowBootstrapControlFormatting) { return }
        while ($cursor -lt $length -and [int][char]$Text[$cursor] -le 0x20) { $cursor++ }
    }
    $parseString = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '"') { throw "$Identity has an invalid JSON string." }
        $cursor++
        $builder = New-Object Text.StringBuilder
        while ($cursor -lt $length) {
            $code = [int][char]$Text[$cursor]
            $cursor++
            if ($code -eq 0x22) { return $builder.ToString() }
            if ($code -lt 0x20) { throw "$Identity string has a control character." }
            if ($code -eq 0x5c) {
                if ($cursor -ge $length) { throw "$Identity string has a trailing escape." }
                $escape = $Text[$cursor]
                $cursor++
                switch ($escape) {
                    '"' { [void]$builder.Append('"'); continue }
                    '\' { [void]$builder.Append('\'); continue }
                    'b' { [void]$builder.Append([char]0x08); continue }
                    'f' { [void]$builder.Append([char]0x0c); continue }
                    'n' { [void]$builder.Append([char]0x0a); continue }
                    'r' { [void]$builder.Append([char]0x0d); continue }
                    't' { [void]$builder.Append([char]0x09); continue }
                    'u' {
                        if (($cursor + 4) -gt $length) { throw "$Identity string has a truncated unicode escape." }
                        $hex = $Text.Substring($cursor, 4)
                        if ($hex -cnotmatch '^00[0-9A-F]{2}$') { throw "$Identity string has a noncanonical unicode escape." }
                        [void]$builder.Append([char][Convert]::ToInt32($hex, 16))
                        $cursor += 4
                        continue
                    }
                    default { throw "$Identity string has a forbidden escape." }
                }
            }
            if ($code -ge 0xd800 -and $code -le 0xdbff) {
                if ($cursor -ge $length) { throw "$Identity string has an unpaired high surrogate." }
                $low = [int][char]$Text[$cursor]
                if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw "$Identity string has an unpaired high surrogate." }
                [void]$builder.Append([char]$code)
                [void]$builder.Append([char]$low)
                $cursor++
                continue
            }
            if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw "$Identity string has an unpaired low surrogate." }
            [void]$builder.Append([char]$code)
        }
        throw "$Identity string is unterminated."
    }
    $parseArray = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '[') { throw "$Identity has an invalid JSON array." }
        $cursor++
        . $skipWhitespace
        $items = New-Object 'System.Collections.Generic.List[object]'
        if ($cursor -lt $length -and $Text[$cursor] -eq ']') {
            $cursor++
            return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
        }
        while ($true) {
            [void]$items.Add((. $parseValue))
            . $skipWhitespace
            if ($cursor -ge $length) { throw "$Identity array is unterminated." }
            if ($Text[$cursor] -eq ']') { $cursor++; break }
            if ($Text[$cursor] -ne ',') { throw "$Identity array has an invalid separator or whitespace." }
            $cursor++
            . $skipWhitespace
        }
        return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
    }
    $parseObject = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '{') { throw "$Identity has an invalid JSON object." }
        $cursor++
        . $skipWhitespace
        $members = [ordered]@{}
        if ($cursor -lt $length -and $Text[$cursor] -eq '}') {
            $cursor++
            return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
        }
        while ($true) {
            $name = . $parseString
            if ($members.Contains($name)) { throw "$Identity has a duplicate object member." }
            . $skipWhitespace
            if ($cursor -ge $length -or $Text[$cursor] -ne ':') { throw "$Identity object has an invalid member separator or whitespace." }
            $cursor++
            . $skipWhitespace
            $members[$name] = . $parseValue
            . $skipWhitespace
            if ($cursor -ge $length) { throw "$Identity object is unterminated." }
            if ($Text[$cursor] -eq '}') { $cursor++; break }
            if ($Text[$cursor] -ne ',') { throw "$Identity object has an invalid separator or whitespace." }
            $cursor++
            . $skipWhitespace
            if ($cursor -ge $length -or $Text[$cursor] -ne '"') { throw "$Identity object has an invalid member name." }
        }
        return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
    }
    $parseValue = {
        . $skipWhitespace
        if ($cursor -ge $length) { throw "$Identity has an incomplete JSON value." }
        $character = $Text[$cursor]
        if ([int][char]$character -le 0x20) { throw "$Identity has insignificant whitespace." }
        switch ($character) {
            '{' { return (. $parseObject) }
            '[' { return (. $parseArray) }
            '"' { return [pscustomobject][ordered]@{ Type = 'string'; Value = (. $parseString) } }
            'n' {
                if (($cursor + 4) -le $length -and $Text.Substring($cursor, 4) -ceq 'null') { $cursor += 4; return [pscustomobject][ordered]@{ Type = 'null'; Value = $null } }
                throw "$Identity has an invalid literal."
            }
            't' {
                if (($cursor + 4) -le $length -and $Text.Substring($cursor, 4) -ceq 'true') { $cursor += 4; return [pscustomobject][ordered]@{ Type = 'boolean'; Value = $true } }
                throw "$Identity has an invalid literal."
            }
            'f' {
                if (($cursor + 5) -le $length -and $Text.Substring($cursor, 5) -ceq 'false') { $cursor += 5; return [pscustomobject][ordered]@{ Type = 'boolean'; Value = $false } }
                throw "$Identity has an invalid literal."
            }
            default {
                if ($character -notmatch '^[0-9]$') { throw "$Identity has an invalid JSON scalar." }
                $start = $cursor
                while ($cursor -lt $length -and $Text[$cursor] -match '^[0-9]$') { $cursor++ }
                if ($allowBootstrapControlFormatting -and $cursor -lt $length -and $Text[$cursor] -eq '.') {
                    $cursor++
                    $fractionStart = $cursor
                    while ($cursor -lt $length -and $Text[$cursor] -match '^[0-9]$') { $cursor++ }
                    if ($fractionStart -eq $cursor) { throw "$Identity has an invalid bootstrap-control decimal." }
                    $decimalText = $Text.Substring($start, $cursor - $start)
                    try { $decimal = [Double]::Parse($decimalText, [Globalization.CultureInfo]::InvariantCulture) } catch { throw "$Identity has an invalid bootstrap-control decimal." }
                    return [pscustomobject][ordered]@{ Type = 'decimal'; Value = $decimal }
                }
                $digits = $Text.Substring($start, $cursor - $start)
                if (($digits.Length -gt 1 -and $digits[0] -eq '0') -or $digits.Length -gt 19) { throw "$Identity has a noncanonical integer." }
                try { $integer = [Int64]::Parse($digits, [Globalization.CultureInfo]::InvariantCulture) } catch { throw "$Identity has an invalid integer." }
                return [pscustomobject][ordered]@{ Type = 'integer'; Value = $integer }
            }
        }
    }
    $node = . $parseValue
    . $skipWhitespace
    if ($cursor -ne $length -or [string]$node.Type -ne 'object') { throw "$Identity has trailing data, whitespace, or no top-level object." }
    $materialize = $null
    $materialize = {
        param([object]$Current)
        switch ([string]$Current.Type) {
            'object' {
                $result = [ordered]@{}
                foreach ($name in @($Current.Value.Keys)) { $result[[string]$name] = & $materialize $Current.Value[$name] }
                return [pscustomobject]$result
            }
            'array' {
                $result = New-Object 'System.Collections.Generic.List[object]'
                foreach ($item in @($Current.Value)) { [void]$result.Add((& $materialize $item)) }
                return [object[]]$result.ToArray()
            }
            default { return $Current.Value }
        }
    }
    return & $materialize $node
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

function Assert-HubQaPositiveInt64 {
    param([object]$Value, [string]$Identity)

    if ($Value -isnot [Int64] -or [Int64]$Value -le 0) {
        throw "E_RECEIPT_BINDING_INVALID: $Identity is not a positive canonical integer."
    }
}

function Read-HubQaBootstrapControl {
    param([string]$Root)

    $controlPath = Join-Path $Root 'testframework\docker\bootstrap-control.v1.json'
    $raw = Read-HubQaRegularUtf8File -Path $controlPath -Root $Root -Identity 'fixed bootstrap control'
    $control = ConvertFrom-HubQaStrictCanonicalJson -Text $raw.Text -Identity 'fixed bootstrap control'
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
        $records[$key] = [pscustomobject]@{ Raw = $record; Value = ConvertFrom-HubQaStrictCanonicalJson -Text $record.Text -Identity "A1 $key control" }
    }
    return $records
}

function ConvertTo-HubQaDependencyBundleCanonicalJsonValue {
    param([object]$Value)

    if ($null -eq $Value) { return 'null' }
    if ($Value -is [string]) { return ConvertTo-HubQaJcsString -Value ([string]$Value) }
    if ($Value -is [bool]) {
        if ([bool]$Value) { return 'true' }
        return 'false'
    }
    if ($Value -is [Int64]) { return ([Int64]$Value).ToString([Globalization.CultureInfo]::InvariantCulture) }
    if ($Value -is [Array]) {
        $items = New-Object Collections.Generic.List[string]
        foreach ($item in @($Value)) { $items.Add((ConvertTo-HubQaDependencyBundleCanonicalJsonValue -Value $item)) }
        return '[' + ($items -join ',') + ']'
    }
    if ($Value -isnot [pscustomobject]) { throw 'Dependency manifest projection has an unsupported value type.' }
    $members = New-Object Collections.Generic.List[string]
    foreach ($property in @($Value.PSObject.Properties)) {
        $members.Add((ConvertTo-HubQaJcsString -Value ([string]$property.Name)) + ':' + (ConvertTo-HubQaDependencyBundleCanonicalJsonValue -Value $property.Value))
    }
    return '{' + ($members -join ',') + '}'
}

function Get-HubQaDependencyBundleProjectionSha256 {
    param([object]$Dependency)

    $members = New-Object Collections.Generic.List[string]
    foreach ($property in @($Dependency.PSObject.Properties)) {
        if ([string]$property.Name -ceq 'bundle_sha256') { continue }
        $members.Add((ConvertTo-HubQaJcsString -Value ([string]$property.Name)) + ':' + (ConvertTo-HubQaDependencyBundleCanonicalJsonValue -Value $property.Value))
    }
    return Get-HubQaSha256 -Bytes $script:Utf8.GetBytes(('{' + ($members -join ',') + '}'))
}

function Assert-HubQaDependencyBundleContract {
    param([object]$Dependency)

    Assert-HubQaExactPropertyNames -Object $Dependency -Expected @('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256') -Identity 'dependency manifest'
    if ($Dependency.schema -isnot [string] -or $Dependency.schema -cne 'hub.qa.dependency-bundle/v1') { throw 'Dependency manifest schema is not closed.' }
    if ($Dependency.bundle_id -isnot [string] -or $Dependency.bundle_id -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$') { throw 'Dependency manifest bundle ID is invalid.' }
    if ($Dependency.source_commit -isnot [string] -or $Dependency.source_commit -cne $script:SourceBaseline) { throw 'Dependency manifest source baseline mismatch.' }
    if ($Dependency.bundle_sha256 -isnot [string]) { throw 'Dependency manifest bundle digest is not a string.' }
    Assert-HubQaDigest -Value ([string]$Dependency.bundle_sha256) -Identity 'dependency manifest bundle digest'

    Assert-HubQaExactPropertyNames -Object $Dependency.platform -Expected @('os', 'architecture', 'variant') -Identity 'dependency manifest platform'
    if ($Dependency.platform.os -isnot [string] -or $Dependency.platform.architecture -isnot [string] -or
        $Dependency.platform.os -cne 'linux' -or $Dependency.platform.architecture -cne 'amd64' -or $null -ne $Dependency.platform.variant) {
        throw 'Dependency manifest platform is not linux/amd64/null.'
    }

    $expectedAssets = @(
        [pscustomobject][ordered]@{ asset_id = 'go-module-cache'; targets = @('backend-unit', 'backend-openapi'); relative_path = 'assets/go-module-cache.tar' },
        [pscustomobject][ordered]@{ asset_id = 'go-race-toolchain'; targets = @('backend-unit'); relative_path = 'assets/go-race-toolchain.tar' },
        [pscustomobject][ordered]@{ asset_id = 'frontend-node-modules'; targets = @('frontend-unit'); relative_path = 'assets/frontend-node-modules.tar' },
        [pscustomobject][ordered]@{ asset_id = 'swag-binary'; targets = @('backend-openapi'); relative_path = 'assets/swag-binary.tar' }
    )
    $assets = @($Dependency.assets)
    if ($Dependency.assets -isnot [Array] -or $assets.Count -ne $expectedAssets.Count) { throw 'Dependency manifest asset array is not closed.' }
    $allowedSpdx = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach ($license in @('0BSD', 'Apache-2.0', 'Artistic-2.0', 'BSD-2-Clause', 'BSD-3-Clause', 'CC0-1.0', 'EPL-2.0', 'ISC', 'MIT', 'MPL-2.0', 'PSF-2.0', 'Python-2.0', 'Unicode-DFS-2016', 'Zlib')) { [void]$allowedSpdx.Add($license) }
    for ($index = 0; $index -lt $expectedAssets.Count; $index++) {
        $asset = $assets[$index]
        $expected = $expectedAssets[$index]
        Assert-HubQaExactPropertyNames -Object $asset -Expected @('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity "dependency manifest asset $index"
        if ($asset.asset_id -isnot [string] -or $asset.relative_path -isnot [string] -or $asset.archive_format -isnot [string] -or
            $asset.asset_id -cne $expected.asset_id -or $asset.relative_path -cne $expected.relative_path -or $asset.archive_format -cne 'posix-pax-tar') {
            throw "Dependency manifest asset $index mapping is not closed."
        }
        if ($asset.targets -isnot [Array]) { throw "Dependency manifest asset $index targets are not an array." }
        Assert-HubQaExactStringArray -Actual $asset.targets -Expected @($expected.targets) -Identity "dependency manifest asset $index targets"
        foreach ($field in @('sha256', 'layout_sha256', 'license_evidence_sha256', 'provenance_sha256')) {
            if ($asset.$field -isnot [string]) { throw "Dependency manifest asset $index $field is not a digest string." }
            Assert-HubQaDigest -Value ([string]$asset.$field) -Identity "dependency manifest asset $index $field"
        }
        foreach ($field in @('bytes', 'member_count', 'expanded_bytes')) {
            if ($asset.$field -isnot [Int64] -or [Int64]$asset.$field -le 0) { throw "Dependency manifest asset $index $field is not a positive canonical integer." }
        }
        if ($asset.license_spdx -isnot [string] -or -not $allowedSpdx.Contains([string]$asset.license_spdx)) { throw "Dependency manifest asset $index SPDX value is not closed." }
    }

    $expectedBases = @(
        [pscustomobject][ordered]@{ base_id = 'go-backend-unit'; consumers = @('backend-unit') },
        [pscustomobject][ordered]@{ base_id = 'go-backend-openapi'; consumers = @('backend-openapi') },
        [pscustomobject][ordered]@{ base_id = 'frontend-unit'; consumers = @('frontend-unit') },
        [pscustomobject][ordered]@{ base_id = 'bootstrap-helper-python'; consumers = @('g0-admission-helper', 'g0-planning-selftest') }
    )
    $bases = @($Dependency.bases)
    if ($Dependency.bases -isnot [Array] -or $bases.Count -ne $expectedBases.Count) { throw 'Dependency manifest base array is not closed.' }
    for ($index = 0; $index -lt $expectedBases.Count; $index++) {
        $base = $bases[$index]
        $expected = $expectedBases[$index]
        Assert-HubQaExactPropertyNames -Object $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity "dependency manifest base $index"
        if ($base.base_id -isnot [string] -or $base.base_id -cne $expected.base_id) { throw "Dependency manifest base $index mapping is not closed." }
        if ($base.consumers -isnot [Array]) { throw "Dependency manifest base $index consumers are not an array." }
        Assert-HubQaExactStringArray -Actual $base.consumers -Expected @($expected.consumers) -Identity "dependency manifest base $index consumers"
        foreach ($field in @('local_image_id', 'manifest_digest', 'license_evidence_sha256', 'provenance_sha256')) {
            if ($base.$field -isnot [string]) { throw "Dependency manifest base $index $field is not a digest string." }
            Assert-HubQaDigest -Value ([string]$base.$field) -Identity "dependency manifest base $index $field"
        }
        if ($base.license_spdx -isnot [string] -or -not $allowedSpdx.Contains([string]$base.license_spdx)) { throw "Dependency manifest base $index SPDX value is not closed." }
    }

    if ((Get-HubQaDependencyBundleProjectionSha256 -Dependency $Dependency) -cne [string]$Dependency.bundle_sha256) {
        throw 'Dependency manifest non-self bundle digest mismatches its canonical projection.'
    }
}

function Assert-HubQaBoundControlContract {
    param([object]$Controls, [object]$BootstrapControl)

    $dependency = $Controls.dependency.Value
    Assert-HubQaDependencyBundleContract -Dependency $dependency
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
        $expectedBuildTargetKeys = @('dockerfile', 'dockerfile_sha256', 'entrypoint_path', 'entrypoint_sha256', 'docker_entrypoint', 'docker_cmd', 'compose_service', 'local_image_lock_field', 'target_action', 'bootstrap_selectable')
        if ($targetName -ceq 'frontend-unit') {
            $expectedBuildTargetKeys += @('vitest_config_path', 'vitest_config_sha256')
        }
        Assert-HubQaExactPropertyNames -Object $target -Expected $expectedBuildTargetKeys -Identity "build manifest target $targetName"
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
        if ($targetName -ceq 'frontend-unit') {
            $vitestConfigRelative = 'testframework/docker/vitest.frontend-unit.config.ts'
            if ($target.vitest_config_path -isnot [string] -or $target.vitest_config_path -cne $vitestConfigRelative) {
                throw 'Build manifest frontend Vitest config path is not closed.'
            }
            Assert-HubQaDigest -Value ([string]$target.vitest_config_sha256) -Identity 'build manifest frontend Vitest config digest'
            $vitestConfig = Read-HubQaRegularUtf8File -Path (Join-Path $Root $vitestConfigRelative) -Root $Root -Identity 'frontend Vitest config'
            if ($vitestConfig.Sha256 -cne $target.vitest_config_sha256) {
                throw 'Build manifest frontend Vitest config byte binding mismatch.'
            }
        }
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

    $expectedKeys = [string[]]@('Task', 'Run', 'RoleClaim', 'OwnerNonce', 'OwnerMarkerRawBytes', 'OwnerMarkerSha256', 'RunnerInspectionRawBytes', 'RunnerInspectionSha256', 'LocalBaseReceiptRawBytes', 'LocalBaseReceiptSha256', 'LocalBaseReceiptNonce')
    if ($BootstrapPrivateContext.Count -ne $expectedKeys.Count) { throw 'Runner-private bootstrap context has an unexpected field count.' }
    foreach ($name in $expectedKeys) {
        if (-not $BootstrapPrivateContext.ContainsKey($name)) { throw "Runner-private bootstrap context lacks $name." }
    }
    if ([string]$BootstrapPrivateContext.Task -cne 'g0-assets-runner') { throw 'Runner-private bootstrap context task is not closed.' }
    if ([string]$BootstrapPrivateContext.Run -notmatch '^[a-z0-9][a-z0-9-]{7,63}$') { throw 'Runner-private bootstrap context run grammar is invalid.' }
    if ([string]$BootstrapPrivateContext.RoleClaim -cne 'hub.qa-bootstrap/v1') { throw 'Runner-private bootstrap context role claim is invalid.' }
    if ([string]$BootstrapPrivateContext.OwnerNonce -notmatch '^[0-9a-f]{32}$') { throw 'Runner-private bootstrap context owner nonce is invalid.' }
    if ($BootstrapPrivateContext.OwnerMarkerRawBytes -isnot [byte[]] -or $BootstrapPrivateContext.RunnerInspectionRawBytes -isnot [byte[]] -or $BootstrapPrivateContext.LocalBaseReceiptRawBytes -isnot [byte[]]) {
        throw 'Runner-private bootstrap context must retain raw owner-marker, image-receipt, and local-base-receipt bytes.'
    }
    Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext.OwnerMarkerSha256) -Identity 'runner-private owner marker digest'
    Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext.RunnerInspectionSha256) -Identity 'runner-private inspection receipt digest'
    Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext.LocalBaseReceiptSha256) -Identity 'runner-private local-base receipt digest'
    if ([string]$BootstrapPrivateContext.LocalBaseReceiptNonce -notmatch '^[0-9a-f]{32}$') { throw 'Runner-private bootstrap context local-base receipt nonce is invalid.' }
    if ((Get-HubQaSha256 -Bytes $BootstrapPrivateContext.OwnerMarkerRawBytes) -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.RunnerInspectionRawBytes) -cne [string]$BootstrapPrivateContext.RunnerInspectionSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.LocalBaseReceiptRawBytes) -cne [string]$BootstrapPrivateContext.LocalBaseReceiptSha256) {
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
        LocalBaseReceiptPath = Join-Path $bootstrapRoot 'local-base-inspect.v1.json'
        AdmissionPath = Join-Path $bootstrapRoot 'g0-admission-helper.admission.v1.json'
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
    $marker = ConvertFrom-HubQaStrictCanonicalJson -Text $ownerMarker.Text -Identity 'runner-owned bootstrap owner marker'
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
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Text $raw.Text -Identity 'runner-produced local image inspection receipt'
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

function Initialize-HubQaWin32FileApis {
    if ($null -ne ('HubQaWin32FileNative' -as [type])) { return }
    Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;

[StructLayout(LayoutKind.Sequential)]
public struct HubQaByHandleFileInformation {
    public uint FileAttributes;
    public System.Runtime.InteropServices.ComTypes.FILETIME CreationTime;
    public System.Runtime.InteropServices.ComTypes.FILETIME LastAccessTime;
    public System.Runtime.InteropServices.ComTypes.FILETIME LastWriteTime;
    public uint VolumeSerialNumber;
    public uint FileSizeHigh;
    public uint FileSizeLow;
    public uint NumberOfLinks;
    public uint FileIndexHigh;
    public uint FileIndexLow;
}

public static class HubQaWin32FileNative {
    public const uint GENERIC_READ = 0x80000000;
    public const uint FILE_SHARE_READ = 0x00000001;
    public const uint OPEN_EXISTING = 3;
    public const uint FILE_ATTRIBUTE_DIRECTORY = 0x00000010;
    public const uint FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400;
    public const uint FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000;

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true, EntryPoint = "CreateFileW")]
    public static extern SafeFileHandle CreateFileW(string path, uint access, uint share, IntPtr security, uint disposition, uint flags, IntPtr template);

    [DllImport("kernel32.dll", SetLastError = true)]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool GetFileInformationByHandle(SafeFileHandle handle, out HubQaByHandleFileInformation information);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern uint GetFinalPathNameByHandle(SafeFileHandle handle, StringBuilder path, uint length, uint flags);

    [DllImport("kernel32.dll", SetLastError = true)]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool ReadFile(SafeFileHandle handle, byte[] buffer, uint count, out uint read, IntPtr overlapped);
}
'@
}

function ConvertFrom-HubQaWin32FinalPath {
    param([string]$Path)

    if ($Path.StartsWith('\\?\UNC\', [StringComparison]::OrdinalIgnoreCase)) { return '\\' + $Path.Substring(8) }
    if ($Path.StartsWith('\\?\', [StringComparison]::Ordinal)) { return $Path.Substring(4) }
    return $Path
}

function Get-HubQaWin32FinalPath {
    param([Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle, [string]$Identity)

    [int]$capacity = 1024
    while ($true) {
        $buffer = New-Object Text.StringBuilder $capacity
        [uint32]$written = [HubQaWin32FileNative]::GetFinalPathNameByHandle($Handle, $buffer, [uint32]$capacity, [uint32]0)
        if ($written -eq 0) { throw "E_ASSET_ARCHIVE_INVALID: $Identity GetFinalPathNameByHandle failed." }
        if ($written -lt $capacity) { return ConvertFrom-HubQaWin32FinalPath -Path $buffer.ToString() }
        if ($capacity -gt 32768) { throw "E_ASSET_ARCHIVE_INVALID: $Identity final path exceeds the closed limit." }
        $capacity = [int]$written + 1
    }
}

function Get-HubQaWin32FileIdentity {
    param([Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle, [string]$Identity)

    $information = New-Object HubQaByHandleFileInformation
    if (-not [HubQaWin32FileNative]::GetFileInformationByHandle($Handle, [ref]$information)) {
        throw "E_ASSET_ARCHIVE_INVALID: $Identity GetFileInformationByHandle failed."
    }
    [uint64]$size = (([uint64]$information.FileSizeHigh -shl 32) -bor [uint64]$information.FileSizeLow)
    return [pscustomobject]@{
        Attributes = [uint32]$information.FileAttributes
        VolumeSerialNumber = [uint32]$information.VolumeSerialNumber
        FileIndexHigh = [uint32]$information.FileIndexHigh
        FileIndexLow = [uint32]$information.FileIndexLow
        Size = [int64]$size
        LastWriteHigh = [int32]$information.LastWriteTime.dwHighDateTime
        LastWriteLow = [uint32]$information.LastWriteTime.dwLowDateTime
    }
}

function Assert-HubQaSameWin32FileIdentity {
    param([object]$Actual, [object]$Expected, [string]$Identity)

    foreach ($field in @('Attributes', 'VolumeSerialNumber', 'FileIndexHigh', 'FileIndexLow', 'Size', 'LastWriteHigh', 'LastWriteLow')) {
        if ($Actual.$field -ne $Expected.$field) { throw "E_ASSET_ARCHIVE_INVALID: $Identity changed identity during safe staged-tar rehash." }
    }
}

function Open-HubQaSafeContainedRegularFile {
    param([string]$Path, [string]$Root, [string]$Identity)

    Initialize-HubQaWin32FileApis
    $fullPath = Assert-HubQaContainedPath -Path $Path -Root $Root -Identity $Identity
    $parent = [IO.DirectoryInfo]::new([IO.Path]::GetDirectoryName($fullPath))
    while ($null -ne $parent -and $parent.FullName.Length -ge $Root.Length) {
        $parent.Refresh()
        if (-not $parent.Exists -or ($parent.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw "E_ASSET_ARCHIVE_INVALID: $Identity has a reparse ancestor." }
        if ([StringComparer]::OrdinalIgnoreCase.Equals($parent.FullName, $Root)) { break }
        $parent = $parent.Parent
    }
    $handle = [HubQaWin32FileNative]::CreateFileW($fullPath, [HubQaWin32FileNative]::GENERIC_READ, [HubQaWin32FileNative]::FILE_SHARE_READ, [IntPtr]::Zero, [HubQaWin32FileNative]::OPEN_EXISTING, [HubQaWin32FileNative]::FILE_FLAG_OPEN_REPARSE_POINT, [IntPtr]::Zero)
    if ($null -eq $handle -or $handle.IsInvalid) { throw "E_ASSET_ARCHIVE_INVALID: $Identity CreateFileW safe open failed." }
    try {
        $proof = Get-HubQaWin32FileIdentity -Handle $handle -Identity $Identity
        if (($proof.Attributes -band [HubQaWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -ne 0 -or ($proof.Attributes -band [HubQaWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) {
            throw "E_ASSET_ARCHIVE_INVALID: $Identity is not an existing non-reparse regular target."
        }
        $finalPath = Get-HubQaWin32FinalPath -Handle $handle -Identity $Identity
        [void](Assert-HubQaContainedPath -Path $finalPath -Root $Root -Identity "$Identity final opened path")
        return [pscustomobject]@{ Handle = $handle; Proof = $proof; FinalPath = $finalPath }
    } catch {
        $handle.Dispose()
        throw
    }
}

function Get-HubQaSafeStreamedFilePass {
    param([string]$Path, [string]$Root, [string]$Identity, [Int64]$ExpectedBytes, [Int64]$HardCapBytes, [object]$ExpectedProof = $null)

    $opened = Open-HubQaSafeContainedRegularFile -Path $Path -Root $Root -Identity $Identity
    try {
        if ($null -ne $ExpectedProof) { Assert-HubQaSameWin32FileIdentity -Actual $opened.Proof -Expected $ExpectedProof -Identity "$Identity safe reopen" }
        if ($ExpectedBytes -le 0 -or $HardCapBytes -le 0 -or $ExpectedBytes -gt $HardCapBytes -or $ExpectedBytes -ge [Int64]::MaxValue) {
            throw "E_ASSET_ARCHIVE_INVALID: $Identity has an invalid expected byte bound."
        }
        if ($opened.Proof.Size -ne $ExpectedBytes -or $opened.Proof.Size -gt $HardCapBytes) {
            throw "E_ASSET_ARCHIVE_INVALID: $Identity opened file length is outside the expected byte bound."
        }
        $algorithm = [Security.Cryptography.SHA256]::Create()
        try {
            [byte[]]$buffer = New-Object byte[] 1048576
            [Int64]$length = 0
            [Int64]$maximumReadableBytes = $ExpectedBytes + 1
            while ($true) {
                $readCount = [uint32][Math]::Min([Int64]$buffer.Length, $maximumReadableBytes - $length)
                [uint32]$read = 0
                if (-not [HubQaWin32FileNative]::ReadFile($opened.Handle, $buffer, $readCount, [ref]$read, [IntPtr]::Zero)) {
                    throw "E_ASSET_ARCHIVE_INVALID: $Identity streamed ReadFile failed."
                }
                if ($read -eq 0) { break }
                $length += [Int64]$read
                if ($length -gt $ExpectedBytes) {
                    throw "E_ASSET_ARCHIVE_INVALID: $Identity streamed archive exceeds the expected byte bound."
                }
                [void]$algorithm.TransformBlock($buffer, 0, [int]$read, $buffer, 0)
            }
            [void]$algorithm.TransformFinalBlock($buffer, 0, 0)
            $postHashProof = Get-HubQaWin32FileIdentity -Handle $opened.Handle -Identity $Identity
            Assert-HubQaSameWin32FileIdentity -Actual $postHashProof -Expected $opened.Proof -Identity "$Identity before/after hash"
            if ($length -ne $opened.Proof.Size) { throw "E_ASSET_ARCHIVE_INVALID: $Identity streamed length differs from opened identity." }
            return [pscustomobject]@{ Bytes = $length; Sha256 = 'sha256:' + ([BitConverter]::ToString($algorithm.Hash)).Replace('-', '').ToLowerInvariant(); Proof = $opened.Proof; FinalPath = $opened.FinalPath }
        } finally {
            $algorithm.Dispose()
        }
    } finally {
        $opened.Handle.Dispose()
    }
}

function Get-HubQaStreamedRegularFileDigest {
    param([string]$Path, [string]$Root, [string]$Identity, [Int64]$ExpectedBytes, [Int64]$HardCapBytes)

    $first = Get-HubQaSafeStreamedFilePass -Path $Path -Root $Root -Identity $Identity -ExpectedBytes $ExpectedBytes -HardCapBytes $HardCapBytes
    $second = Get-HubQaSafeStreamedFilePass -Path $Path -Root $Root -Identity $Identity -ExpectedBytes $ExpectedBytes -HardCapBytes $HardCapBytes -ExpectedProof $first.Proof
    Assert-HubQaSameWin32FileIdentity -Actual $second.Proof -Expected $first.Proof -Identity "$Identity safe reopen identity"
    if ($second.FinalPath -cne $first.FinalPath -or $second.Bytes -ne $first.Bytes -or $second.Sha256 -cne $first.Sha256) {
        throw "E_ASSET_ARCHIVE_INVALID: $Identity differs after safe reopen rehash."
    }
    return [pscustomobject]@{ Bytes = $second.Bytes; Sha256 = $second.Sha256 }
}

function Assert-HubQaArchiveIdentity {
    param([object]$Identity, [object]$Expected, [Int64]$ExpectedBytes, [string]$IdentityName)

    Assert-HubQaExactPropertyNames -Object $Identity -Expected @('device', 'inode', 'mode', 'size', 'mtime_ns') -Identity $IdentityName
    if ($Identity.device -isnot [Int64]) { throw "E_ASSET_ARCHIVE_INVALID: $IdentityName device must be an Int64." }
    if ($Identity.inode -isnot [Int64]) { throw "E_ASSET_ARCHIVE_INVALID: $IdentityName inode must be an Int64." }
    if ($Identity.mode -isnot [Int64]) { throw "E_ASSET_ARCHIVE_INVALID: $IdentityName mode must be an Int64." }
    if ($Identity.size -isnot [Int64]) { throw "E_ASSET_ARCHIVE_INVALID: $IdentityName size must be an Int64." }
    if ($Identity.mtime_ns -isnot [Int64]) { throw "E_ASSET_ARCHIVE_INVALID: $IdentityName mtime_ns must be an Int64." }
    if (($Identity.mode -band 0xf000) -ne 0x8000 -or $Identity.size -ne $ExpectedBytes) {
        throw "E_ASSET_ARCHIVE_INVALID: $IdentityName is not the admitted regular-file identity."
    }
    if ($null -ne $Expected) {
        foreach ($field in @('device', 'inode', 'mode', 'size', 'mtime_ns')) {
            if ($Identity.$field -ne $Expected.$field) {
                throw "E_ASSET_ARCHIVE_INVALID: $IdentityName differs from the opened descriptor identity."
            }
        }
    }
}

function Assert-HubQaAdmissionBaseBinding {
    param([object]$ReceiptBases, [object]$ManifestBases)

    $actual = @($ReceiptBases)
    $expected = @($ManifestBases)
    if ($actual.Count -ne $expected.Count) { throw 'E_RECEIPT_BINDING_INVALID: helper admission base count differs from the dependency manifest.' }
    for ($index = 0; $index -lt $expected.Count; $index++) {
        Assert-HubQaExactPropertyNames -Object $actual[$index] -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest') -Identity 'helper admission base'
        Assert-HubQaExactPropertyNames -Object $expected[$index] -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity 'dependency manifest base'
        $manifestProjection = [ordered]@{
            base_id = $expected[$index].base_id
            consumers = $expected[$index].consumers
            local_image_id = $expected[$index].local_image_id
            manifest_digest = $expected[$index].manifest_digest
        }
        if ($actual[$index].base_id -isnot [string] -or $manifestProjection.base_id -isnot [string] -or
            $actual[$index].base_id -cne $manifestProjection.base_id) {
            throw 'E_RECEIPT_BINDING_INVALID: helper admission base ID differs from the dependency manifest projection.'
        }
        if ($actual[$index].consumers -isnot [Array] -or $manifestProjection.consumers -isnot [Array]) {
            throw 'E_RECEIPT_BINDING_INVALID: helper admission base consumers are not closed arrays.'
        }
        Assert-HubQaExactStringArray -Actual $actual[$index].consumers -Expected @($manifestProjection.consumers) -Identity 'helper admission base consumers'
        foreach ($field in @('local_image_id', 'manifest_digest')) {
            if ($actual[$index].$field -isnot [string] -or $manifestProjection.$field -isnot [string] -or
                $actual[$index].$field -cne $manifestProjection.$field) {
                throw 'E_RECEIPT_BINDING_INVALID: helper admission base value differs from the dependency manifest projection.'
            }
            Assert-HubQaDigest -Value ([string]$actual[$index].$field) -Identity "helper admission base $field"
        }
    }
}

function Read-HubQaLocalBaseReceipt {
    param([pscustomobject]$Paths, [hashtable]$BootstrapPrivateContext, [object]$Controls)

    $raw = Read-HubQaRegularUtf8File -Path $Paths.LocalBaseReceiptPath -Root $Paths.TaskRunRoot -Identity 'runner-produced local-base inspection receipt'
    if (-not (Test-HubQaByteArrayEqual -Left $raw.Bytes -Right $BootstrapPrivateContext.LocalBaseReceiptRawBytes) -or
        $raw.Sha256 -cne [string]$BootstrapPrivateContext.LocalBaseReceiptSha256) {
        throw 'E_RECEIPT_BINDING_INVALID: local-base inspection receipt is not the exact raw receipt held by the post-build context.'
    }
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Text $raw.Text -Identity 'runner-produced local-base inspection receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @('schema', 'task', 'run', 'owner_marker_sha256', 'receipt_nonce', 'bases') -Identity 'runner-produced local-base inspection receipt'
    if ($receipt.schema -cne 'hub.qa.local-base-inspect/v1' -or
        $receipt.task -cne $BootstrapPrivateContext.Task -or
        $receipt.run -cne $BootstrapPrivateContext.Run -or
        $receipt.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $receipt.receipt_nonce -cne $BootstrapPrivateContext.LocalBaseReceiptNonce) {
        throw 'E_RECEIPT_BINDING_INVALID: local-base inspection receipt task, run, owner, or nonce is not closed over the held context.'
    }
    Assert-HubQaAdmissionBaseBinding -ReceiptBases $receipt.bases -ManifestBases $Controls.dependency.Value.bases
    return $receipt
}

function Read-HubQaHelperAdmissionReceipt {
    param([pscustomobject]$Paths, [hashtable]$BootstrapPrivateContext, [object]$Controls, [object]$BootstrapControl)

    $raw = Read-HubQaRegularUtf8File -Path $Paths.AdmissionPath -Root $Paths.TaskRunRoot -Identity 'helper admission receipt'
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Text $raw.Text -Identity 'helper admission receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @('schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases', 'local_base_receipt_sha256', 'receipt_nonce', 'verdict') -Identity 'helper admission receipt'
    if ($receipt.schema -cne 'hub.qa.g0-dependency-admission-receipt/v1' -or $receipt.bundle_id -cne $Controls.dependency.Value.bundle_id -or
        $receipt.source_commit -cne $script:SourceBaseline -or $receipt.manifest_sha256 -cne $Controls.dependency.Raw.Sha256 -or
        $receipt.local_base_receipt_sha256 -cne $BootstrapPrivateContext.LocalBaseReceiptSha256 -or
        $receipt.receipt_nonce -cne $BootstrapPrivateContext.LocalBaseReceiptNonce -or $receipt.verdict -cne 'ACCEPT') {
        throw 'E_RECEIPT_BINDING_INVALID: helper admission receipt identity is not bound to the closed bootstrap inputs.'
    }
    Assert-HubQaDigest -Value ([string]$receipt.local_base_receipt_sha256) -Identity 'helper admission local-base receipt digest'
    Assert-HubQaAdmissionBaseBinding -ReceiptBases $receipt.bases -ManifestBases $Controls.dependency.Value.bases

    $assets = @($receipt.assets)
    $manifestAssets = @($Controls.dependency.Value.assets)
    if ($assets.Count -ne $manifestAssets.Count) { throw 'E_RECEIPT_BINDING_INVALID: helper admission asset count differs from the dependency manifest.' }
    $seenAssetIds = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    for ($assetIndex = 0; $assetIndex -lt $manifestAssets.Count; $assetIndex++) {
        $asset = $assets[$assetIndex]
        $manifestAsset = $manifestAssets[$assetIndex]
        Assert-HubQaExactPropertyNames -Object $asset -Expected @('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'descriptor_contract', 'source_observations', 'staged_targets') -Identity 'helper admission asset'
        if ($asset.asset_id -isnot [string] -or $asset.asset_id -cne $manifestAsset.asset_id) {
            throw 'E_RECEIPT_BINDING_INVALID: helper admission asset order differs from the dependency manifest.'
        }
        if (-not $seenAssetIds.Add([string]$asset.asset_id)) { throw 'E_RECEIPT_BINDING_INVALID: helper admission receipt has a duplicate asset ID.' }
        foreach ($field in @('bytes', 'member_count', 'expanded_bytes')) {
            Assert-HubQaPositiveInt64 -Value $asset.$field -Identity "helper admission asset $assetIndex $field"
            Assert-HubQaPositiveInt64 -Value $manifestAsset.$field -Identity "dependency manifest asset $assetIndex $field"
            if ([Int64]$asset.$field -ne [Int64]$manifestAsset.$field) {
                throw 'E_RECEIPT_BINDING_INVALID: helper admission asset numeric field differs from the dependency manifest.'
            }
        }
        foreach ($field in @('asset_id', 'relative_path', 'sha256', 'archive_format', 'layout_sha256')) {
            if ([string]$asset.$field -cne [string]$manifestAsset.$field) { throw 'E_RECEIPT_BINDING_INVALID: helper admission asset differs from the dependency manifest.' }
        }
        Assert-HubQaDigest -Value ([string]$asset.sha256) -Identity 'helper admission asset digest'
        Assert-HubQaDigest -Value ([string]$asset.layout_sha256) -Identity 'helper admission asset layout digest'
        Assert-HubQaExactStringArray -Actual $asset.targets -Expected @($manifestAsset.targets) -Identity 'helper admission asset target mapping'
        if ($asset.descriptor_contract -cne 'linux-held-nofollow-fd/v1') { throw 'E_ASSET_ARCHIVE_INVALID: helper admission asset descriptor contract is not closed.' }
        Assert-HubQaExactPropertyNames -Object $asset.source_observations -Expected @('opened', 'post_hash', 'post_raw_scan', 'post_tarfile', 'pre_stage', 'post_stage') -Identity 'helper admission source observations'
        $openedDescriptor = $null
        foreach ($phase in @('opened', 'post_hash', 'post_raw_scan', 'post_tarfile', 'pre_stage', 'post_stage')) {
            $observation = Get-HubQaMember -Object $asset.source_observations -Name $phase -Identity 'helper admission source observation'
            Assert-HubQaExactPropertyNames -Object $observation -Expected @('descriptor', 'pathname') -Identity 'helper admission source observation pair'
            $expectedIdentity = if ($null -eq $openedDescriptor) { $null } else { $openedDescriptor }
            Assert-HubQaArchiveIdentity -Identity $observation.descriptor -Expected $expectedIdentity -ExpectedBytes $asset.bytes -IdentityName "helper admission $phase descriptor"
            if ($null -eq $openedDescriptor) { $openedDescriptor = $observation.descriptor }
            Assert-HubQaArchiveIdentity -Identity $observation.pathname -Expected $openedDescriptor -ExpectedBytes $asset.bytes -IdentityName "helper admission $phase pathname"
        }
        $stagedTargets = @($asset.staged_targets)
        $expectedTargets = @($manifestAsset.targets)
        if ($stagedTargets.Count -ne $expectedTargets.Count) { throw 'E_ASSET_ARCHIVE_INVALID: helper admission target staging count is incomplete.' }
        for ($targetIndex = 0; $targetIndex -lt $expectedTargets.Count; $targetIndex++) {
            $staged = $stagedTargets[$targetIndex]
            Assert-HubQaExactPropertyNames -Object $staged -Expected @('target', 'relative_tar_path', 'bytes', 'sha256', 'create_new', 'fsync', 'safe_reopen', 'rehash_sha256') -Identity 'helper admission target staging'
            $targetName = [string]$expectedTargets[$targetIndex]
            Assert-HubQaPositiveInt64 -Value $staged.bytes -Identity "helper admission staged target $assetIndex/$targetIndex bytes"
            if ($null -eq $script:FixedTargets[$targetName] -or $null -eq $BootstrapControl.Value.targets.PSObject.Properties[$targetName] -or $staged.target -cne $targetName -or
                $staged.relative_tar_path -cne ('bootstrap/build-context/' + $targetName + '/bindings/' + [string]$asset.relative_path) -or
                [Int64]$staged.bytes -ne [Int64]$asset.bytes -or $staged.sha256 -cne $asset.sha256 -or $staged.rehash_sha256 -cne $asset.sha256 -or
                $staged.create_new -isnot [bool] -or -not [bool]$staged.create_new -or $staged.fsync -isnot [bool] -or -not [bool]$staged.fsync -or
                $staged.safe_reopen -isnot [bool] -or -not [bool]$staged.safe_reopen) {
                throw 'E_ASSET_ARCHIVE_INVALID: helper admission staged target does not retain the admitted archive identity.'
            }
            $relativeComponents = ([string]$staged.relative_tar_path).Split('/')
            $stagedPath = $Paths.TaskRunRoot
            foreach ($component in $relativeComponents) {
                if ([string]::IsNullOrEmpty($component) -or $component -eq '.' -or $component -eq '..' -or $component.IndexOf([IO.Path]::DirectorySeparatorChar) -ge 0 -or $component.IndexOf([IO.Path]::AltDirectorySeparatorChar) -ge 0) {
                    throw 'E_ASSET_ARCHIVE_INVALID: helper admission staged path is not a closed relative path.'
                }
                $stagedPath = Join-Path $stagedPath $component
            }
            $streamed = Get-HubQaStreamedRegularFileDigest -Path $stagedPath -Root $Paths.TaskRunRoot -Identity 'helper admission staged archive' -ExpectedBytes ([Int64]$asset.bytes) -HardCapBytes ([Int64]$manifestAsset.bytes)
            if ($streamed.Bytes -ne [Int64]$asset.bytes -or $streamed.Sha256 -cne [string]$asset.sha256) {
                throw 'E_ASSET_ARCHIVE_INVALID: helper admission staged archive bytes do not match the admitted asset.'
            }
        }
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
    [void](Read-HubQaLocalBaseReceipt -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls)
    $admission = Read-HubQaHelperAdmissionReceipt -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls -BootstrapControl $binding.BootstrapControl
    # Reserve with create-new semantics only after runner ownership and every
    # held-descriptor archive admission/staging receipt are closed; then
    # consume the nonce/digest-bound local inspection receipt.
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
        AdmissionReceiptSchema = [string]$admission.schema
        SourceTreeSha256 = $sourceTree.SourceTreeSha256
    }
}

if ($Describe -or -not $PSBoundParameters.ContainsKey('Describe')) {
    Get-HubQaBootstrapImageContract | ConvertTo-Json -Depth 16
}
