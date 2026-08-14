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
            '/opt/hub/frontend/node_modules/.bin/tsc -b'
            '/opt/hub/frontend/node_modules/.bin/biome check ./src'
            '/opt/hub/frontend/node_modules/.bin/vite build --outDir /reports/frontend-unit/vite-dist'
            '/opt/hub/frontend/node_modules/.bin/vitest run --config /opt/hub/config/vitest.frontend-unit.config.ts --reporter=json'
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
    '[IO.Directory]::CreateDirectory($Paths.BootstrapRoot)',
    'testframework/docker/images.lock.v3.json'
)) {
    Assert-SourceExcludes -Source $attestor -Needle $forbidden -Identity 'attestor caller-controlled or retired capability'
}
Assert-SourceOrder -Source $attestor -Needles @(
    '[void](Assert-HubQaRunnerOwnedBootstrapRoot -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext)',
    '$reservation = New-HubQaCandidateReservation -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext',
    '$inspection = Read-HubQaRunnerInspection -Paths $paths -BootstrapPrivateContext $BootstrapPrivateContext -Controls $binding.Controls'
) -Identity 'attestor runner ownership, reservation, inspection order'

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
    'assert_omitted_binding_rejections',
    'assert_owner_binding_probes',
    'binding-null-task-owner',
    'binding-null-handoff-owner',
    'binding-equal-owners',
    'binding-mismatch-task-owner',
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
