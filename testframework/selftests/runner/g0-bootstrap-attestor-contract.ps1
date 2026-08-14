Set-StrictMode -Version Latest

# Static source contract only. Package C is the first authorized runtime proof.
$script:PreBootstrapStatus = 'NOT_READY/NOT_RUN'
$root = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\..'))
$paths = [ordered]@{
    Launcher = Join-Path $root 'testframework\docker\admit-local-dependency-assets.ps1'
    Schema = Join-Path $root 'testframework\docker\dependency-bundle.schema.v1.json'
    Verifier = Join-Path $root 'testframework\docker\admission\verify_local_dependency_bundle.py'
    Admission = Join-Path $root 'testframework\docker\admission\admit_local_dependency_assets.py'
    Fixture = Join-Path $root 'testframework\docker\runner-fixtures\g0-planning-selftest.py'
    HelperDockerfile = Join-Path $root 'testframework\docker\Dockerfile.g0-admission-helper'
    HelperEntrypoint = Join-Path $root 'testframework\docker\entrypoints\g0-admission-helper.sh'
}

foreach ($path in $paths.Values) {
    if (-not [System.IO.File]::Exists($path)) {
        throw "G0 dependency source contract is missing: $path"
    }
}

$launcher = [System.IO.File]::ReadAllText($paths.Launcher)
$schema = [System.IO.File]::ReadAllText($paths.Schema)
$verifier = [System.IO.File]::ReadAllText($paths.Verifier)
$admission = [System.IO.File]::ReadAllText($paths.Admission)
$fixture = [System.IO.File]::ReadAllText($paths.Fixture)
$dockerfile = [System.IO.File]::ReadAllText($paths.HelperDockerfile)
$entrypoint = [System.IO.File]::ReadAllText($paths.HelperEntrypoint)

foreach ($required in @(
    'Invoke-HubQaG0AdmissionFromBootstrapContext',
    'PrivateIngressCopyReceipt',
    'PrivateIngressOrderedPaths',
    'PrivateIngressByteCounts',
    'PrivateIngressSourceRoot',
    'PrivateIngressDestinationRoot',
    'PrivateOfflineIngressHandoffId',
    'PrivateOfflineIngressOwnerMarkerSha256',
    'PrivateDependencyBundleManifest',
    'PrivateLocalBaseReceipt',
    'PrivateLocalBaseReceiptSha256',
    'PrivateLocalBaseReceiptNonce',
    'PrivateHelperBuildTag',
    'PrivateHelperImageAttestation',
    'PrivateHelperImageAttestationSha256',
    'PrivateHelperLocalImageId',
    'PrivateBootstrapHelperPythonLocalImageId',
    'PrivateBootstrapHelperPythonLocalImageIdSha256',
    'Test-HubQaByteArrayEqual',
    'Assert-HubQaIngressCopyReceipt',
    'Assert-HubQaClosedCopiedIngressTree',
    'Assert-HubQaPrivateHelperImageAttestation',
    'g0-admission-helper.local-image-attestation.v1.json',
    'GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles)',
    'Docker\Docker\resources\bin\docker.exe',
    'HUB_QA_PRIVATE_OWNER_MARKER_SHA256',
    'HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256',
    'OwnerMarkerSha256 -ceq $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256',
    "'--pull', 'never'",
    '''--name'', $closed.HelperContainerName',
    '''--label'', $closed.HelperOwnerLabel',
    'hub.qa.owner-marker-sha256=',
    '$BootstrapPrivateContext.PrivateHelperLocalImageId',
    'dst=/work',
    'NOT_READY/NOT_RUN'
)) {
    if ($launcher.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Host launcher does not expose required closed bootstrap token: $required"
    }
}

foreach ($forbidden in @(
    'LocalBaseInventory',
    'RepositoryRoot',
    'CandidatePath',
    'TargetAttestations',
    'HandoffManifestSha256',
    'SourceIngressRoot',
    'Assert-BaseBinding',
    'verify-local-dependency-bundle.ps1',
    'docker compose',
    'pwsh',
    "'run', '--rm'",
    'C:\Program Files\Docker\Docker\resources\bin\docker.exe',
    'Get-ChildItem'
)) {
    if ($launcher.IndexOf($forbidden, [System.StringComparison]::OrdinalIgnoreCase) -ge 0) {
        throw "Host launcher contains forbidden public or legacy capability: $forbidden"
    }
}

if ($launcher.TrimStart().StartsWith('param(', [System.StringComparison]::OrdinalIgnoreCase)) {
    throw 'Host launcher must not expose a script-level public parameter surface.'
}

foreach ($required in @(
    '"targets"',
    '"consumers"',
    'every closed object retains its declared key order',
    '"asset_id",',
    '"targets",',
    '"base_id",',
    '"consumers",'
)) {
    if ($schema.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Dependency schema omits required closed mapping token: $required"
    }
}
foreach ($forbidden in @(
    'owner_marker_sha256: str | None',
    'handoff_owner_marker_sha256: str | None'
)) {
    if ($admission.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0) {
        throw "Shared Python admission wrapper retains an optional owner-marker binding: $forbidden"
    }
}

foreach ($required in @(
    '_ASSET_KEYS',
    '"asset_id", "targets", "relative_path"',
    '_BASE_KEYS',
    '"base_id", "consumers", "local_image_id"',
    'sort_keys=False',
    'tuple(bundle.keys()) != _ROOT_KEYS',
    'E_DUPLICATE_ASSET_ID',
    'found_directories',
    'expected_directories',
    'expected_task_owner_marker_sha256: str',
    'expected_handoff_owner_marker_sha256: str',
    'def _validate_expected_owner_bindings(',
    'if task_owner_marker_sha256 == handoff_owner_marker_sha256:',
    '"verdict": "ACCEPT"',
    '"manifest_sha256": manifest_sha',
    '"assets": assets',
    '"bases": bases',
    '"reason_code": None'
)) {
    if ($verifier.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Shared Python verifier omits required closed contract token: $required"
    }
}
foreach ($forbidden in @(
    'task_owner_marker_sha256: str | None',
    'handoff_owner_marker_sha256: str | None',
    'task_owner_marker_sha256 if task_owner_marker_sha256 is not None else actual_handoff_owner_marker_sha'
)) {
    if ($verifier.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0) {
        throw "Shared Python verifier retains an optional or substituted owner-marker binding: $forbidden"
    }
}

$planReasonCodes = @(
    'E_MANIFEST_UTF8',
    'E_MANIFEST_SCHEMA',
    'E_DUPLICATE_ASSET_ID',
    'E_ASSET_PATH_ESCAPE',
    'E_ASSET_MISSING',
    'E_ASSET_BYTES_MISMATCH',
    'E_ASSET_SHA256_MISMATCH',
    'E_BUNDLE_SHA256_MISMATCH',
    'E_LICENSE_SPDX_INVALID',
    'E_LICENSE_EVIDENCE_MISSING',
    'E_PROVENANCE_EVIDENCE_MISSING',
    'E_BASE_MISSING',
    'E_BASE_ID_MISMATCH',
    'E_SOURCE_COMMIT_MISMATCH',
    'E_RECEIPT_BINDING_INVALID'
)
foreach ($code in $planReasonCodes) {
    if ($verifier.IndexOf($code, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Shared Python verifier omits plan reason code: $code"
    }
}
foreach ($forbidden in @('E_MANIFEST_UNBOUND', 'E_MANIFEST_BINDING_MISMATCH', 'E_INGRESS_TREE_INVALID', 'E_BASE_RECEIPT_BINDING')) {
    if ($verifier.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0 -or
        $admission.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0 -or
        $fixture.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0) {
        throw "Dependency source retains an unplanned reason code: $forbidden"
    }
}

foreach ($required in @(
    '_verify_bundle_or_raise',
    'HUB_QA_PRIVATE_OWNER_MARKER_SHA256',
    'HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256',
    'expected_task_owner_marker_sha256=',
    'expected_handoff_owner_marker_sha256=',
    'Path("/work")',
    'E_RECEIPT_BINDING_INVALID'
)) {
    if ($admission.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Shared Python admission wrapper omits required closed contract token: $required"
    }
}

foreach ($required in @(
    'from verify_local_dependency_bundle import verify_bundle, BundleVerificationError, SOURCE_COMMIT',
    'from admit_local_dependency_assets import admit_bundle, AdmissionError',
    '"targets": list(targets)',
    '"consumers": list(consumers)',
    'handoff-owner.v1.json',
    'dependency-bundle.manifest.v1.json',
    'owner_marker_sha256',
    'E_DUPLICATE_ASSET_ID',
    'E_MANIFEST_SCHEMA'
)) {
    if ($fixture.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Image-baked shared Python fixture omits required contract token: $required"
    }
}
foreach ($forbidden in @('sort_keys=True', '"consumers": list(targets)', 'E_MANIFEST_UNBOUND')) {
    if ($fixture.IndexOf($forbidden, [System.StringComparison]::Ordinal) -ge 0) {
        throw "Image-baked shared Python fixture retains an invalid contract token: $forbidden"
    }
}

foreach ($required in @(
    'hub-qa-bootstrap-helper-python:local',
    'dependency-bundle.schema.v1.json',
    'verify_local_dependency_bundle.py',
    'admit_local_dependency_assets.py',
    'g0-admission-helper.sh',
    'controls/dependency-bundle.manifest.v1.json',
    'ENTRYPOINT ["/opt/hub/bin/g0-admission-helper"]'
)) {
    if ($dockerfile.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Helper Dockerfile does not contain closed contract token: $required"
    }
}
foreach ($forbidden in @('ARG ', 'RUN ', 'ADD ', 'curl', 'wget', 'apt', 'apk', 'pip ', 'npm ', 'git ', 'pwsh', 'admit-local-dependency-assets.ps1')) {
    if ($dockerfile.IndexOf($forbidden, [System.StringComparison]::OrdinalIgnoreCase) -ge 0) {
        throw "Helper Dockerfile contains forbidden build capability: $forbidden"
    }
}

if ($entrypoint.IndexOf('python3 /opt/hub/admission/admit_local_dependency_assets.py', [System.StringComparison]::Ordinal) -lt 0 -or
    $entrypoint.IndexOf('"$#" -ne 0', [System.StringComparison]::Ordinal) -lt 0) {
    throw 'Helper entrypoint must execute only the fixed Python admission module and reject arguments.'
}

return [pscustomobject][ordered]@{
    status = $script:PreBootstrapStatus
    contract = 'hub.qa.g0-dependency-admission-source/v1'
}
