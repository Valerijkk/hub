[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Failing-first A0 source contract. It is static-only and remains NOT_RUN until
# the separately founder-released hub.qa-bootstrap/v1 transaction executes the
# image-baked Python fixture. This file never invokes Docker or a runtime.

function Assert-SourceContains {
    param([string]$Source, [string]$Needle, [string]$Identity)

    if ($Source.IndexOf($Needle, [StringComparison]::Ordinal) -lt 0) {
        throw "G0 image contract is missing $Identity: $Needle"
    }
}

function Assert-SourceExcludes {
    param([string]$Source, [string]$Needle, [string]$Identity)

    if ($Source.IndexOf($Needle, [StringComparison]::Ordinal) -ge 0) {
        throw "G0 image contract contains forbidden $Identity: $Needle"
    }
}

function Assert-SourceOrder {
    param([string]$Source, [string[]]$Needles, [string]$Identity)

    $offset = -1
    foreach ($needle in $Needles) {
        $offset = $Source.IndexOf($needle, $offset + 1, [StringComparison]::Ordinal)
        if ($offset -lt 0) { throw "$Identity is missing or reorders $needle" }
    }
}

function Get-FunctionBody {
    param([string]$Source, [string]$Name)

    $pattern = '(?ms)^function ' + [regex]::Escape($Name) + ' \{\r?\n(?<body>.*?)(?=^function |\z)'
    $match = [regex]::Match($Source, $pattern)
    if (-not $match.Success) { throw "G0 image contract is missing function $Name." }
    return $match.Groups['body'].Value
}

function Assert-ExactStringArray {
    param([object]$Actual, [string[]]$Expected, [string]$Identity)

    $values = @($Actual)
    if ($values.Count -ne $Expected.Count) { throw "$Identity has an unexpected count." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if (-not [StringComparer]::Ordinal.Equals([string]$values[$index], $Expected[$index])) {
            throw "$Identity differs at index $index."
        }
    }
}

function Assert-ExactObjectKeys {
    param([object]$Value, [string[]]$Expected, [string]$Identity)

    $actual = @($Value.PSObject.Properties.Name)
    Assert-ExactStringArray -Actual $actual -Expected $Expected -Identity $Identity
}

function Read-StrictUtf8File {
    param([string]$Path, [string]$Identity)

    [byte[]]$bytes = [IO.File]::ReadAllBytes($Path)
    if ($bytes.Length -eq 0 -or ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf)) {
        throw "$Identity is empty or has a UTF-8 BOM."
    }
    try { return ([Text.UTF8Encoding]::new($false, $true)).GetString($bytes) } catch { throw "$Identity is not strict UTF-8." }
}

function Get-ComposeServiceBody {
    param([string]$Compose, [string]$Service)

    $pattern = '(?ms)^  ' + [regex]::Escape($Service) + ':\r?\n(?<body>.*?)(?=^  [a-z0-9-]+:\r?\n|\z)'
    $match = [regex]::Match($Compose, $pattern)
    if (-not $match.Success) { throw "Bootstrap Compose is missing service $Service." }
    return $match.Groups['body'].Value
}

$repositoryRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\..'))
$dockerRoot = Join-Path $repositoryRoot 'testframework\docker'
$targets = [ordered]@{
    'backend-unit' = [ordered]@{
        Dockerfile = 'Dockerfile.backend-unit'
        Entrypoint = 'entrypoints/backend-unit.sh'
        EntrypointPath = '/opt/hub/bin/backend-unit'
        ComposeService = 'backend-unit'
        LocalImageLockField = 'targets.backend-unit.local_image_id'
        DockerCmd = @('--hub-qa-target=backend-unit')
        TargetAction = @(
            'go build ./...'
            'go vet ./...'
            'go test -json -count=1 -timeout=150s ./tests/...'
            'go test -json -race -count=1 -timeout=150s ./tests/...'
        )
        BootstrapSelectable = $false
    }
    'frontend-unit' = [ordered]@{
        Dockerfile = 'Dockerfile.frontend-unit'
        Entrypoint = 'entrypoints/frontend-unit.sh'
        EntrypointPath = '/opt/hub/bin/frontend-unit'
        ComposeService = 'frontend-unit'
        LocalImageLockField = 'targets.frontend-unit.local_image_id'
        DockerCmd = @('--hub-qa-target=frontend-unit')
        TargetAction = @(
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/typescript/bin/tsc -b'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/@biomejs/biome/bin/biome check ./src'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/vite/bin/vite.js build --outDir /reports/frontend-unit/vite-dist'
            '/usr/local/bin/node /opt/hub/source/frontend/node_modules/vitest/vitest.mjs run --config /opt/hub/config/vitest.frontend-unit.config.ts --reporter=json'
        )
        BootstrapSelectable = $false
    }
    'backend-openapi' = [ordered]@{
        Dockerfile = 'Dockerfile.backend-openapi'
        Entrypoint = 'entrypoints/backend-openapi.sh'
        EntrypointPath = '/opt/hub/bin/backend-openapi'
        ComposeService = 'backend-openapi'
        LocalImageLockField = 'targets.backend-openapi.local_image_id'
        DockerCmd = @('--hub-qa-target=backend-openapi')
        TargetAction = @(
            '/opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-1 --parseDependency --parseInternal'
            '/opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-2 --parseDependency --parseInternal'
            'sha256 manifest comparison and exact three-path synchronization'
        )
        BootstrapSelectable = $false
    }
    'g0-planning-selftest' = [ordered]@{
        Dockerfile = 'Dockerfile.g0-planning-selftest'
        Entrypoint = 'entrypoints/g0-planning-selftest.sh'
        EntrypointPath = '/opt/hub/bin/g0-planning-selftest'
        ComposeService = 'g0-planning-selftest'
        LocalImageLockField = 'targets.g0-planning-selftest.local_image_id'
        DockerCmd = @('--hub-qa-target=g0-planning-selftest')
        TargetAction = @(
            'image-baked Python stdlib fixture imports the shared A0 verifier and admission modules directly'
            'fixed valid and negative dependency-bundle fixtures prove verifier reasons and no success receipt after rejection'
        )
        BootstrapSelectable = $true
    }
}

$attestor = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'source-image-attestor.ps1') -Identity 'source image attestor'
foreach ($token in @(
    'Get-HubQaBootstrapImageContract',
    'Test-HubQaBootstrapImageBindings',
    'Invoke-HubQaBootstrapImageAttestation',
    'Assert-HubQaPrivateBootstrapContext',
    'RunnerInspectionRawBytes',
    'RunnerInspectionSha256',
    'LocalBaseReceiptRawBytes',
    'LocalBaseReceiptSha256',
    'LocalBaseReceiptNonce',
    'OwnerMarkerRawBytes',
    'OwnerMarkerSha256',
    'owner-marker.v1.json',
    'owner_marker_sha256',
    'receipt_nonce',
    'Read-HubQaBootstrapControl',
    'bootstrap-control.v1.json',
    'local_image_lock_field',
    'target_action',
    'bootstrap_selectable',
    'New-HubQaCandidateReservation',
    'Assert-HubQaRunnerOwnedBootstrapRoot',
    'pre-existing runner-owned bootstrap root',
    'hub.qa.bootstrap-owner/v1',
    '$taskRunName = [string]$BootstrapPrivateContext.Task + ''-'' + [string]$BootstrapPrivateContext.Run',
    'Resolve-HubQaCanonicalDirectory -Path $Paths.TaskRunRoot',
    'Resolve-HubQaCanonicalDirectory -Path $Paths.BootstrapRoot',
    'Get-HubQaGitObjectSourceTree',
    'git ls-tree',
    'git cat-file',
    'E_SOURCE_TREE_INVALID',
    'E_SOURCE_TREE_DIGEST_MISMATCH',
    '$PSScriptRoot',
    '7b5d5501cefab34f9aa5661e4f8c38e71474f9ce',
    'dependency-bundle.manifest.v1.json',
    'source-policy.v2.json',
    'build-manifest.v3.json',
    'local-image-inspect.v1.json',
    'g0-admission-helper.admission.v1.json',
    'Read-HubQaHelperAdmissionReceipt',
    'Read-HubQaLocalBaseReceipt',
    'Assert-HubQaDependencyBundleContract',
    'Assert-HubQaAdmissionBaseBinding',
    'Get-HubQaDependencyBundleProjectionSha256',
    'ConvertFrom-HubQaStrictCanonicalJson',
    'CreateFileW',
    'SafeFileHandle',
    'GetFileInformationByHandle',
    'GetFinalPathNameByHandle',
    'FILE_FLAG_OPEN_REPARSE_POINT',
    'Get-HubQaSafeStreamedFilePass',
    'Assert-HubQaSameWin32FileIdentity',
    'ReadFile',
    'Get-HubQaStreamedRegularFileDigest',
    'E_ASSET_ARCHIVE_INVALID',
    'E_RECEIPT_BINDING_INVALID',
    'linux-held-nofollow-fd/v1',
    'source_observations',
    'staged_targets',
    'descriptor_contract',
    "@('schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases', 'local_base_receipt_sha256', 'receipt_nonce', 'verdict')",
    "@('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256')",
    "@('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256')",
    "@('base_id', 'consumers', 'local_image_id', 'manifest_digest')",
    "@('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256')",
    "@('os', 'architecture', 'variant')",
    "'go-module-cache'",
    "@('backend-unit', 'backend-openapi')",
    "'go-race-toolchain'",
    "@('backend-unit')",
    "'frontend-node-modules'",
    "@('frontend-unit')",
    "'swag-binary'",
    "@('backend-openapi')",
    "'bootstrap-helper-python'",
    "@('g0-admission-helper', 'g0-planning-selftest')",
    "'posix-pax-tar'",
    "@('Task', 'Run', 'RoleClaim', 'OwnerNonce', 'OwnerMarkerRawBytes', 'OwnerMarkerSha256', 'RunnerInspectionRawBytes', 'RunnerInspectionSha256', 'LocalBaseReceiptRawBytes', 'LocalBaseReceiptSha256', 'LocalBaseReceiptNonce')",
    "@('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'descriptor_contract', 'source_observations', 'staged_targets')",
    "@('opened', 'post_hash', 'post_raw_scan', 'post_tarfile', 'pre_stage', 'post_stage')",
    "@('target', 'relative_tar_path', 'bytes', 'sha256', 'create_new', 'fsync', 'safe_reopen', 'rehash_sha256')",
    'post_raw_scan',
    'post_tarfile',
    'safe_reopen',
    'rehash_sha256',
    'FileMode]::CreateNew',
    'Get-JcsArraySha256',
    'READY_FOR_G0_BOOTSTRAP_ONLY',
    'testframework/reports/<task-run>/bootstrap/images.lock.v3.json'
)) {
    Assert-SourceContains -Source $attestor -Needle $token -Identity 'attestor closed-control token'
}
foreach ($forbidden in @(
    '[string]$RepositoryRoot',
    '[string]$CandidatePath',
    '[hashtable]$TargetAttestations',
    '[object[]]$TargetRecords',
    '[hashtable]$BootstrapContext',
    'param([string]$Path, [string]$Root, [hashtable]$BootstrapContext',
    '& docker',
    'docker image inspect',
    'Assert-BaseBinding',
    'ConvertFrom-Json',
    '[IO.Directory]::CreateDirectory($Paths.BootstrapRoot)',
    'testframework/docker/images.lock.v3.json'
)) {
    Assert-SourceExcludes -Source $attestor -Needle $forbidden -Identity 'attestor caller-controlled or retired capability'
}
Assert-SourceContains -Source $attestor -Needle '$staged.relative_tar_path -cne (''bootstrap/build-context/'' + $targetName + ''/bindings/'' + [string]$asset.relative_path)' -Identity 'attestor exact bindings staging path'
Assert-SourceContains -Source $attestor -Needle '$manifestAsset = $manifestAssets[$assetIndex]' -Identity 'receipt asset manifest-order binding'
Assert-SourceContains -Source $attestor -Needle '$asset.asset_id -isnot [string] -or $asset.asset_id -cne $manifestAsset.asset_id' -Identity 'receipt asset reordered-entry rejection'
Assert-SourceExcludes -Source $attestor -Needle 'Get-HubQaDependencyAsset -DependencyManifest $Controls.dependency.Value -AssetId ([string]$asset.asset_id)' -Identity 'receipt asset lookup-by-ID acceptance'
Assert-SourceContains -Source $attestor -Needle "Assert-HubQaExactPropertyNames -Object `$actual[`$index] -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest') -Identity 'helper admission base'" -Identity 'receipt base four-field public projection'
Assert-SourceContains -Source $attestor -Needle "Assert-HubQaExactPropertyNames -Object `$expected[`$index] -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity 'dependency manifest base'" -Identity 'manifest base seven-field source contract'
Assert-SourceExcludes -Source $attestor -Needle "Assert-HubQaExactPropertyNames -Object `$actual[`$index] -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity 'helper admission base'" -Identity 'legacy receipt base seven-field contract'
Assert-SourceOrder -Source $attestor -Needles @(
    '[void](Assert-HubQaRunnerOwnedBootstrapRoot -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext)',
    '[void](Read-HubQaLocalBaseReceipt -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls)',
    '$admission = Read-HubQaHelperAdmissionReceipt -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls -BootstrapControl $binding.BootstrapControl',
    '$reservation = New-HubQaCandidateReservation -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext',
    '$inspection = Read-HubQaRunnerInspection -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls'
) -Identity 'attestor runner ownership, reservation, inspection order'
Assert-SourceOrder -Source $attestor -Needles @(
    '$dependency = $Controls.dependency.Value',
    'Assert-HubQaDependencyBundleContract -Dependency $dependency',
    'Assert-HubQaExactPropertyNames -Object $source'
) -Identity 'attestor dependency contract before source or base bindings'
Assert-SourceOrder -Source $attestor -Needles @(
    '$staged.relative_tar_path -cne (''bootstrap/build-context/'' + $targetName + ''/bindings/'' + [string]$asset.relative_path)',
    '$relativeComponents = ([string]$staged.relative_tar_path).Split(''/'')',
    '$stagedPath = $Paths.TaskRunRoot'
) -Identity 'attestor bindings staging path before contained re-root'
Assert-SourceContains -Source $attestor -Needle 'for ($assetIndex = 0; $assetIndex -lt $manifestAssets.Count; $assetIndex++)' -Identity 'receipt asset loop has dedicated index'
Assert-SourceContains -Source $attestor -Needle '$asset = $assets[$assetIndex]' -Identity 'receipt asset lookup uses dedicated asset index'
Assert-SourceContains -Source $attestor -Needle '$manifestAsset = $manifestAssets[$assetIndex]' -Identity 'receipt manifest lookup uses dedicated asset index'
Assert-SourceContains -Source $attestor -Needle 'for ($targetIndex = 0; $targetIndex -lt $expectedTargets.Count; $targetIndex++)' -Identity 'receipt staged-target loop has distinct index'
Assert-SourceContains -Source $attestor -Needle '$staged = $stagedTargets[$targetIndex]' -Identity 'receipt staged-target lookup uses target index'
Assert-SourceContains -Source $attestor -Needle '$targetName = [string]$expectedTargets[$targetIndex]' -Identity 'receipt expected target lookup uses target index'
Assert-SourceOrder -Source $attestor -Needles @(
    'for ($assetIndex = 0; $assetIndex -lt $manifestAssets.Count; $assetIndex++)',
    '$manifestAsset = $manifestAssets[$assetIndex]',
    'for ($targetIndex = 0; $targetIndex -lt $expectedTargets.Count; $targetIndex++)',
    '$staged = $stagedTargets[$targetIndex]'
) -Identity 'receipt asset and staged-target index scopes'
foreach ($field in @('bytes', 'member_count', 'expanded_bytes')) {
    Assert-SourceContains -Source $attestor -Needle ('Assert-HubQaPositiveInt64 -Value $asset.' + $field + ' -Identity "helper admission asset $assetIndex ' + $field + '"') -Identity "receipt asset strict Int64 $field guard"
    Assert-SourceContains -Source $attestor -Needle ('Assert-HubQaPositiveInt64 -Value $manifestAsset.' + $field + ' -Identity "dependency manifest asset $assetIndex ' + $field + '"') -Identity "manifest asset strict Int64 $field guard"
}
Assert-SourceContains -Source $attestor -Needle 'Assert-HubQaPositiveInt64 -Value $staged.bytes -Identity "helper admission staged target $assetIndex/$targetIndex bytes"' -Identity 'staged target strict Int64 bytes guard'
Assert-SourceContains -Source $attestor -Needle 'if ($Value -isnot [Int64] -or [Int64]$Value -le 0)' -Identity 'strict positive Int64 type and range rejection'
$archiveIdentityFunction = Get-FunctionBody -Source $attestor -Name 'Assert-HubQaArchiveIdentity'
foreach ($field in @('device', 'inode', 'mode', 'size', 'mtime_ns')) {
    Assert-SourceContains -Source $archiveIdentityFunction -Needle ('$Identity.' + $field + ' -isnot [Int64]') -Identity "archive identity actual $field strict Int64 rejection"
}
Assert-SourceOrder -Source $archiveIdentityFunction -Needles @(
    '$Identity.device -isnot [Int64]',
    '($Identity.mode -band 0xf000)',
    '$Identity.$field -ne $Expected.$field'
) -Identity 'archive identity types precede range and equality checks'
Assert-SourceExcludes -Source $archiveIdentityFunction -Needle 'try { [void][Int64]$Identity.$field }' -Identity 'archive identity castability-only acceptance'
Assert-SourceExcludes -Source $archiveIdentityFunction -Needle '[Int64]$Identity.$field' -Identity 'archive identity post-validation cast acceptance'
$helperReceiptFunction = Get-FunctionBody -Source $attestor -Name 'Read-HubQaHelperAdmissionReceipt'
Assert-SourceContains -Source $helperReceiptFunction -Needle 'Assert-HubQaArchiveIdentity -Identity $observation.descriptor -Expected $expectedIdentity -ExpectedBytes $asset.bytes -IdentityName "helper admission $phase descriptor"' -Identity 'receipt descriptor identity receives validated receipt bytes'
Assert-SourceContains -Source $helperReceiptFunction -Needle 'Assert-HubQaArchiveIdentity -Identity $observation.pathname -Expected $openedDescriptor -ExpectedBytes $asset.bytes -IdentityName "helper admission $phase pathname"' -Identity 'receipt pathname identity receives validated receipt bytes'
Assert-SourceExcludes -Source $helperReceiptFunction -Needle 'ExpectedBytes ([Int64]$asset.bytes)' -Identity 'receipt observation castability acceptance'
$strictJsonFunction = Get-FunctionBody -Source $attestor -Name 'ConvertFrom-HubQaStrictCanonicalJson'
Assert-SourceContains -Source $strictJsonFunction -Needle '$digits.Length -gt 19' -Identity 'strict JSON parser accepts the Int64 19-digit boundary'
Assert-SourceContains -Source $strictJsonFunction -Needle 'try { $integer = [Int64]::Parse($digits, [Globalization.CultureInfo]::InvariantCulture) } catch { throw "$Identity has an invalid integer." }' -Identity 'strict JSON parser rejects overflowing 19-digit integers through Int64 parsing'
Assert-SourceExcludes -Source $strictJsonFunction -Needle '$digits.Length -gt 18' -Identity 'strict JSON parser rejects the valid 19-digit Int64 boundary'
Assert-SourceContains -Source $attestor -Needle 'param([string]$Path, [string]$Root, [string]$Identity, [Int64]$ExpectedBytes, [Int64]$HardCapBytes, [object]$ExpectedProof = $null)' -Identity 'bounded streamed pass declared byte bounds'
Assert-SourceContains -Source $attestor -Needle 'if ($opened.Proof.Size -ne $ExpectedBytes -or $opened.Proof.Size -gt $HardCapBytes)' -Identity 'streamed pass rejects observed length before reading'
Assert-SourceContains -Source $attestor -Needle '$maximumReadableBytes = $ExpectedBytes + 1' -Identity 'streamed pass caps read at expected plus one'
Assert-SourceContains -Source $attestor -Needle 'if ($length -gt $ExpectedBytes)' -Identity 'streamed pass rejects overflow immediately'
Assert-SourceContains -Source $attestor -Needle 'streamed archive exceeds the expected byte bound.' -Identity 'streamed overflow negative reason'
Assert-SourceContains -Source $attestor -Needle 'param([string]$Path, [string]$Root, [string]$Identity, [Int64]$ExpectedBytes, [Int64]$HardCapBytes)' -Identity 'two-pass streamed digest declares byte bounds'
Assert-SourceContains -Source $attestor -Needle 'Get-HubQaSafeStreamedFilePass -Path $Path -Root $Root -Identity $Identity -ExpectedBytes $ExpectedBytes -HardCapBytes $HardCapBytes' -Identity 'first streamed pass receives expected byte cap'
Assert-SourceContains -Source $attestor -Needle 'Get-HubQaSafeStreamedFilePass -Path $Path -Root $Root -Identity $Identity -ExpectedBytes $ExpectedBytes -HardCapBytes $HardCapBytes -ExpectedProof $first.Proof' -Identity 'second streamed pass receives expected byte cap'
Assert-SourceContains -Source $attestor -Needle 'Get-HubQaStreamedRegularFileDigest -Path $stagedPath -Root $Paths.TaskRunRoot -Identity ''helper admission staged archive'' -ExpectedBytes ([Int64]$asset.bytes) -HardCapBytes ([Int64]$manifestAsset.bytes)' -Identity 'indexed asset bytes bind staged archive bounded rehash'
Assert-SourceOrder -Source $attestor -Needles @(
    '$maximumReadableBytes = $ExpectedBytes + 1',
    '$readCount = [uint32][Math]::Min([Int64]$buffer.Length, $maximumReadableBytes - $length)',
    '[HubQaWin32FileNative]::ReadFile'
) -Identity 'streamed byte cap is applied before ReadFile'

foreach ($targetName in $targets.Keys) {
    $target = $targets[$targetName]
    $dockerfilePath = Join-Path $dockerRoot $target.Dockerfile
    $entrypointPath = Join-Path $dockerRoot $target.Entrypoint
    if (-not [IO.File]::Exists($dockerfilePath)) { throw "Missing Dockerfile for $targetName" }
    if (-not [IO.File]::Exists($entrypointPath)) { throw "Missing entrypoint for $targetName" }

    $dockerfile = Read-StrictUtf8File -Path $dockerfilePath -Identity "$targetName Dockerfile"
    $entrypoint = Read-StrictUtf8File -Path $entrypointPath -Identity "$targetName entrypoint"
    if ($targetName -ceq 'g0-planning-selftest') {
        Assert-SourceContains -Source $dockerfile -Needle 'FROM hub-qa-bootstrap-helper-python:local' -Identity "$targetName A1-bound helper base reference"
    } else {
        Assert-SourceContains -Source $dockerfile -Needle 'FROM sha256:' -Identity "$targetName pinned local base reference"
    }
    Assert-SourceContains -Source $dockerfile -Needle $target.EntrypointPath -Identity "$targetName entrypoint binding"
    Assert-SourceContains -Source $dockerfile -Needle ('CMD ["' + $target.DockerCmd[0] + '"]') -Identity "$targetName fixed Docker Cmd"
    Assert-SourceContains -Source $entrypoint -Needle $target.DockerCmd[0] -Identity "$targetName fixed Docker Cmd acceptance"
    Assert-SourceContains -Source $entrypoint -Needle 'first_exit' -Identity "$targetName first-exit handling"
    Assert-SourceContains -Source $entrypoint -Needle '/reports' -Identity "$targetName report-only output"
    foreach ($forbidden in @('backend/testframework', 'testframework/stages', 'curl ', 'wget ', 'npm install', 'go get', 'docker ')) {
        Assert-SourceExcludes -Source $entrypoint -Needle $forbidden -Identity "$targetName forbidden runtime capability"
    }
}

$backendUnitDockerfile = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'Dockerfile.backend-unit') -Identity 'backend unit Dockerfile'
foreach ($token in @(
    'ADD bindings/assets/go-module-cache.tar /opt/hub/offline/go-module-cache/',
    'ADD bindings/assets/go-race-toolchain.tar /opt/hub/offline/go-race-toolchain/',
    'COPY --chmod=0555 entrypoints/backend-unit.sh /opt/hub/bin/backend-unit',
    'ENV CGO_ENABLED=1 CC=/opt/hub/offline/go-race-toolchain/bin/cc',
    'GOPROXY=off',
    'GOSUMDB=off',
    'GOMODCACHE=/opt/hub/offline/go-module-cache/pkg/mod'
)) {
    Assert-SourceContains -Source $backendUnitDockerfile -Needle $token -Identity 'backend unit archive and race ABI binding'
}
foreach ($forbidden in @('COPY bindings/assets/go-module-cache/', 'COPY bindings/assets/go-race-toolchain/', 'RUN ', ' tar ', 'apt', 'apk', 'yum', 'dnf', 'curl', 'wget', 'go install', 'go get')) {
    Assert-SourceExcludes -Source $backendUnitDockerfile -Needle $forbidden -Identity 'backend unit forbidden archive extraction or installation'
}

$frontendDockerfile = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'Dockerfile.frontend-unit') -Identity 'frontend unit Dockerfile'
foreach ($token in @(
    'ADD bindings/assets/frontend-node-modules.tar /opt/hub/source/frontend/',
    'COPY --chmod=0555 entrypoints/frontend-unit.sh /opt/hub/bin/frontend-unit'
)) {
    Assert-SourceContains -Source $frontendDockerfile -Needle $token -Identity 'frontend local archive and native Node binding'
}
foreach ($forbidden in @('COPY bindings/assets/frontend-node-modules/', 'RUN ', ' tar ', 'apt', 'apk', 'yum', 'dnf', 'curl', 'wget', 'npm install', 'corepack')) {
    Assert-SourceExcludes -Source $frontendDockerfile -Needle $forbidden -Identity 'frontend forbidden archive extraction or installation'
}

$frontendEntrypoint = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'entrypoints\frontend-unit.sh') -Identity 'frontend unit entrypoint'
Assert-SourceContains -Source $frontendEntrypoint -Needle '[ -x /usr/local/bin/node ]' -Identity 'frontend fixed native Node presence gate'
foreach ($action in $targets['frontend-unit'].TargetAction) {
    Assert-SourceContains -Source $frontendEntrypoint -Needle $action -Identity 'frontend fixed direct-module action'
}
foreach ($forbidden in @('/.bin/', ' node ', 'node_modules/.bin', 'PATH=', 'env node')) {
    Assert-SourceExcludes -Source $frontendEntrypoint -Needle $forbidden -Identity 'frontend alternate Node dispatch'
}

$backendOpenapiDockerfile = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'Dockerfile.backend-openapi') -Identity 'backend OpenAPI Dockerfile'
foreach ($token in @(
    'ADD bindings/assets/go-module-cache.tar /opt/hub/offline/go-module-cache/',
    'ADD bindings/assets/swag-binary.tar /opt/hub/offline/swag/',
    'COPY --chmod=0555 entrypoints/backend-openapi.sh /opt/hub/bin/backend-openapi',
    'GOPROXY=off',
    'GOSUMDB=off',
    'GOMODCACHE=/opt/hub/offline/go-module-cache/pkg/mod'
)) {
    Assert-SourceContains -Source $backendOpenapiDockerfile -Needle $token -Identity 'backend OpenAPI archive and swag binding'
}
foreach ($forbidden in @('COPY bindings/assets/go-module-cache/', 'COPY bindings/assets/swag-binary/', 'RUN ', ' tar ', 'apt', 'apk', 'yum', 'dnf', 'curl', 'wget', 'go install', 'go get')) {
    Assert-SourceExcludes -Source $backendOpenapiDockerfile -Needle $forbidden -Identity 'backend OpenAPI forbidden archive extraction or installation'
}
$backendOpenapiEntrypoint = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'entrypoints\backend-openapi.sh') -Identity 'backend OpenAPI entrypoint'
Assert-SourceContains -Source $backendOpenapiEntrypoint -Needle '[ ! -x /opt/hub/offline/swag/swag ]' -Identity 'backend OpenAPI exact swag executable gate'

$planningDockerfile = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'Dockerfile.g0-planning-selftest') -Identity 'planning selftest Dockerfile'
foreach ($token in @(
    'COPY controls/dependency-bundle.manifest.v1.json /opt/hub/controls/dependency-bundle.manifest.v1.json',
    'COPY controls/source-policy.v2.json /opt/hub/controls/source-policy.v2.json',
    'COPY controls/build-manifest.v3.json /opt/hub/controls/build-manifest.v3.json',
    'COPY testframework/docker/dependency-bundle.schema.v1.json /opt/hub/testframework/docker/dependency-bundle.schema.v1.json',
    'COPY testframework/docker/admission/verify_local_dependency_bundle.py /opt/hub/admission/verify_local_dependency_bundle.py',
    'COPY testframework/docker/admission/admit_local_dependency_assets.py /opt/hub/admission/admit_local_dependency_assets.py',
    'COPY testframework/docker/entrypoints/g0-planning-selftest.sh /opt/hub/bin/g0-planning-selftest',
    'COPY testframework/docker/runner-fixtures/g0-planning-selftest.py /opt/hub/testframework/docker/runner-fixtures/g0-planning-selftest.py'
)) {
    Assert-SourceContains -Source $planningDockerfile -Needle $token -Identity 'planning image closed copy set'
}
foreach ($forbidden in @('ARG ', 'apt-get', 'apk add', 'pip install', 'npm install', 'curl ', 'wget ', 'ADD http', 'source/backend', 'source/frontend', '.ps1', 'powershell', 'pwsh', 'testframework/runner/')) {
    Assert-SourceExcludes -Source $planningDockerfile -Needle $forbidden -Identity 'planning image forbidden capability'
}

$fixture = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'runner-fixtures\g0-planning-selftest.py') -Identity 'planning selftest fixture'
foreach ($token in @(
    'from verify_local_dependency_bundle import verify_bundle, BundleVerificationError, SOURCE_COMMIT',
    'from admit_local_dependency_assets import admit_bundle, AdmissionError',
    'valid-manifest',
    'duplicate-asset-id',
    'asset-hash-mismatch',
    'asset-path-escape',
    'missing-asset',
    'unbound-manifest',
    'E_DUPLICATE_ASSET_ID',
    'E_ASSET_SHA256_MISMATCH',
    'E_ASSET_PATH_ESCAPE',
    'E_ASSET_MISSING',
    'E_MANIFEST_SCHEMA',
    'success_receipt_path',
    'handoff-owner.v1.json',
    'hub.qa.g0-offline-ingress-owner/v1',
    'write_task_owner',
    'task_owner_raw',
    'handoff_owner_raw',
    'expected_task_owner_marker_sha256',
    'expected_handoff_owner_marker_sha256',
    'assert_owner_binding_rejections',
    'archive-observation-phase-hooks',
    'archive-replaced-between-passes',
    'archive-replaced-before-staging',
    'archive-pax-invalid-utf8',
    'archive-pax-prefix-cap',
    'E_RECEIPT_BINDING_INVALID',
    'TypeError',
    'dependency_bundle_manifest_sha256',
    'write_bound_manifest',
    'ingress_manifest_path',
    'BUNDLED_SCHEMA_PATH',
    'load_baked_schema',
    'assert_manifest_against_baked_schema',
    'hub.qa.dependency-bundle/v1'
)) {
    Assert-SourceContains -Source $fixture -Needle $token -Identity 'planning fixture shared-admission behavior'
}
foreach ($forbidden in @('subprocess', 'powershell', 'pwsh', '.ps1', 'testframework/runner/', 'fixture-result.json')) {
    Assert-SourceExcludes -Source $fixture -Needle $forbidden -Identity 'planning fixture forbidden runtime imitation'
}

$frontendConfig = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'vitest.frontend-unit.config.ts') -Identity 'frontend Vitest configuration'
Assert-SourceContains -Source $frontendConfig -Needle '/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}' -Identity 'dedicated frontend test include'
Assert-SourceExcludes -Source $frontendConfig -Needle 'exclude:' -Identity 'frontend test exclusion'

$compose = Read-StrictUtf8File -Path (Join-Path $dockerRoot 'compose.g0-bootstrap.yaml') -Identity 'bootstrap Compose topology'
if ([regex]::Matches($compose, '(?m)^  [a-z0-9-]+:$').Count -ne 4) { throw 'Bootstrap Compose must contain exactly four services.' }
Assert-SourceExcludes -Source $compose -Needle 'g0-admission-helper' -Identity 'ephemeral helper service'
Assert-SourceExcludes -Source $compose -Needle 'HUB_QA_REPOSITORY' -Identity 'repository source bind mount'
foreach ($targetName in $targets.Keys) {
    $target = $targets[$targetName]
    $body = Get-ComposeServiceBody -Compose $compose -Service $target.ComposeService
    foreach ($token in @('pull_policy: never', 'platform: linux/amd64', 'read_only: true', 'network_mode: none', 'cpus: 2.0', 'mem_limit: 2g', 'pids_limit: 256', 'cap_drop: ["ALL"]', 'no-new-privileges:true', 'target: /reports')) {
        Assert-SourceContains -Source $body -Needle $token -Identity "$targetName Compose hardening"
    }
    if ($target.BootstrapSelectable) {
        Assert-SourceContains -Source $body -Needle 'profiles: ["bootstrap"]' -Identity "$targetName bootstrap selection"
    } else {
        Assert-SourceContains -Source $body -Needle 'profiles: ["internal-only"]' -Identity "$targetName bootstrap exclusion"
    }
}

$controlPath = Join-Path $dockerRoot 'bootstrap-control.v1.json'
$control = (Read-StrictUtf8File -Path $controlPath -Identity 'bootstrap control') | ConvertFrom-Json -ErrorAction Stop
Assert-ExactObjectKeys -Value $control -Expected @('schema', 'status', 'bootstrap_contract', 'platform', 'candidate_lock', 'runtime', 'targets') -Identity 'bootstrap control root'
if ($control.schema -cne 'hub.qa.bootstrap-control/v1' -or $control.status -cne 'NOT_RUN' -or $control.bootstrap_contract -cne 'hub.qa-bootstrap/v1') { throw 'Bootstrap control root mismatch.' }
Assert-ExactObjectKeys -Value $control.platform -Expected @('os', 'architecture', 'variant') -Identity 'bootstrap control platform'
if ($control.platform.os -cne 'linux' -or $control.platform.architecture -cne 'amd64' -or $null -ne $control.platform.variant) { throw 'Bootstrap control platform mismatch.' }
Assert-ExactObjectKeys -Value $control.runtime -Expected @('network_mode', 'read_only', 'cpus', 'mem_limit', 'pids_limit', 'cap_drop', 'no_new_privileges', 'pull_policy', 'only_writable_mount') -Identity 'bootstrap control runtime'
if (@($control.targets.PSObject.Properties).Count -ne 4) { throw 'Bootstrap control must have exactly four targets.' }
foreach ($targetName in $targets.Keys) {
    $target = $targets[$targetName]
    $property = $control.targets.PSObject.Properties[$targetName]
    if ($null -eq $property) { throw "Bootstrap control is missing $targetName" }
    $value = $property.Value
    Assert-ExactObjectKeys -Value $value -Expected @('dockerfile', 'entrypoint_path', 'compose_service', 'local_image_lock_field', 'docker_cmd', 'target_action', 'bootstrap_selectable') -Identity "bootstrap control target $targetName"
    if ($value.dockerfile -cne ('testframework/docker/' + $target.Dockerfile)) { throw "Bootstrap Dockerfile mismatch for $targetName" }
    if ($value.entrypoint_path -cne $target.EntrypointPath) { throw "Bootstrap entrypoint mismatch for $targetName" }
    if ($value.compose_service -cne $target.ComposeService) { throw "Bootstrap Compose service mismatch for $targetName" }
    if ($value.local_image_lock_field -cne $target.LocalImageLockField) { throw "Bootstrap lock field mismatch for $targetName" }
    Assert-ExactStringArray -Actual $value.docker_cmd -Expected $target.DockerCmd -Identity "Bootstrap Docker Cmd for $targetName"
    Assert-ExactStringArray -Actual $value.target_action -Expected $target.TargetAction -Identity "Bootstrap target action for $targetName"
    if (([bool]($value.bootstrap_selectable)) -ne ([bool]($target.BootstrapSelectable))) { throw "Bootstrap selection mismatch for $targetName" }
}
Assert-SourceExcludes -Source (Read-StrictUtf8File -Path $controlPath -Identity 'bootstrap control raw') -Needle 'g0-admission-helper' -Identity 'ephemeral helper control record'

[ordered]@{
    suite = 'hub.qa.g0-image-contract/v1'
    status = 'NOT_RUN'
    reason = 'Pre-bootstrap static source contract only; runtime execution requires one separately founder-released hub.qa-bootstrap/v1 transaction.'
    targets = @($targets.Keys)
} | ConvertTo-Json -Compress
