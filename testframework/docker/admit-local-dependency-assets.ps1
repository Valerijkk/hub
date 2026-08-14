Set-StrictMode -Version Latest

# This bootstrap-private host consumer has no script-level parameter surface and
# produces no pre-bootstrap runtime evidence.
$script:PreBootstrapStatus = 'NOT_READY/NOT_RUN'
$script:HelperBuildTag = 'hub-qa-g0-admission-helper:local'
$script:SourceCommit = '7b5d5501cefab34f9aa5661e4f8c38e71474f9ce'

function Get-HubQaDockerClientPath {
    $programFiles = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles)
    if ([string]::IsNullOrWhiteSpace($programFiles)) {
        throw 'The operating system ProgramFiles location is unavailable.'
    }
    return [System.IO.Path]::GetFullPath((Join-Path $programFiles 'Docker\Docker\resources\bin\docker.exe'))
}

function Get-HubQaSha256 {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)

    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        return 'sha256:' + ([System.BitConverter]::ToString($algorithm.ComputeHash($Bytes)).Replace('-', '').ToLowerInvariant())
    }
    finally {
        $algorithm.Dispose()
    }
}

function Test-HubQaByteArrayEqual {
    param(
        [Parameter(Mandatory = $true)][byte[]]$Left,
        [Parameter(Mandatory = $true)][byte[]]$Right
    )

    if ($Left.Length -ne $Right.Length) {
        return $false
    }
    for ($index = 0; $index -lt $Left.Length; $index++) {
        if ($Left[$index] -ne $Right[$index]) {
            return $false
        }
    }
    return $true
}

function Assert-HubQaExactPropertyNames {
    param(
        [Parameter(Mandatory = $true)][object]$Object,
        [Parameter(Mandatory = $true)][string[]]$Expected,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $actual = @($Object.PSObject.Properties | ForEach-Object { $_.Name })
    if ($actual.Count -ne $Expected.Count) {
        throw "$Identity has an additional or missing property."
    }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ($actual[$index] -cne $Expected[$index]) {
            throw "$Identity property order or name is not closed."
        }
    }
}

function Read-HubQaRegularFileBytes {
    param([Parameter(Mandatory = $true)][string]$Path)

    $file = [System.IO.FileInfo]::new($Path)
    $file.Refresh()
    if (-not $file.Exists -or
        ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
        ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw 'A required bootstrap receipt or copied file is missing or nonregular.'
    }
    return [System.IO.File]::ReadAllBytes($file.FullName)
}

function ConvertFrom-HubQaStrictUtf8Json {
    param(
        [Parameter(Mandatory = $true)][byte[]]$Bytes,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($Bytes.Length -ge 3 -and $Bytes[0] -eq 0xEF -and $Bytes[1] -eq 0xBB -and $Bytes[2] -eq 0xBF) {
        throw "$Identity is UTF-8 with a forbidden BOM."
    }
    try {
        $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($Bytes)
        return ConvertFrom-Json -InputObject $text -ErrorAction Stop
    }
    catch {
        throw "$Identity is not strict UTF-8 JSON."
    }
}

function Assert-HubQaDigest {
    param([Parameter(Mandatory = $true)][string]$Value, [Parameter(Mandatory = $true)][string]$Identity)

    if ($Value -notmatch '^sha256:[0-9a-f]{64}$') {
        throw "$Identity is not a lowercase SHA-256 digest."
    }
}

function Assert-HubQaPrivateBootstrapContext {
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $required = @(
        'Task',
        'Run',
        'OwnerMarkerSha256',
        'PrivateReportRoot',
        'PrivateIngressRoot',
        'PrivateIngressCopyReceipt',
        'PrivateIngressCopyReceiptSha256',
        'PrivateIngressDigestMap',
        'PrivateIngressDigestMapSha256',
        'PrivateIngressOrderedPaths',
        'PrivateIngressByteCounts',
        'PrivateIngressSourceRoot',
        'PrivateIngressDestinationRoot',
        'PrivateOfflineIngressHandoffId',
        'PrivateOfflineIngressOwnerMarkerSha256',
        'PrivateDependencyBundleManifest',
        'PrivateDependencyBundleManifestSha256',
        'PrivateSourceCommit',
        'PrivateLocalBaseReceipt',
        'PrivateLocalBaseReceiptSha256',
        'PrivateLocalBaseReceiptNonce',
        'PrivateDockerClientPath',
        'PrivateHelperBuildTag',
        'PrivateHelperImageAttestation',
        'PrivateHelperImageAttestationSha256',
        'PrivateHelperLocalImageId',
        'PrivateBootstrapHelperPythonLocalImageId',
        'PrivateBootstrapHelperPythonLocalImageIdSha256'
    )
    if ($BootstrapPrivateContext.Keys.Count -ne $required.Count) {
        throw 'Bootstrap private context has an additional or missing field.'
    }
    foreach ($name in $required) {
        if (-not $BootstrapPrivateContext.ContainsKey($name)) {
            throw 'Bootstrap private context has an additional or missing field.'
        }
    }

    $dockerClientPath = Get-HubQaDockerClientPath
    if ($BootstrapPrivateContext.Task -cne 'g0-assets-runner' -or
        $BootstrapPrivateContext.Run -notmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or
        $BootstrapPrivateContext.PrivateOfflineIngressHandoffId -notmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or
        $BootstrapPrivateContext.PrivateSourceCommit -cne $script:SourceCommit -or
        $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -notmatch '^[0-9a-f]{32}$') {
        throw 'Bootstrap private context has an invalid closed task, run, or source identity.'
    }
    foreach ($digestName in @(
        'OwnerMarkerSha256',
        'PrivateIngressCopyReceiptSha256',
        'PrivateIngressDigestMapSha256',
        'PrivateOfflineIngressOwnerMarkerSha256',
        'PrivateDependencyBundleManifestSha256',
        'PrivateLocalBaseReceiptSha256',
        'PrivateHelperImageAttestationSha256',
        'PrivateBootstrapHelperPythonLocalImageIdSha256'
    )) {
        Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext[$digestName]) -Identity $digestName
    }
    if ($BootstrapPrivateContext.OwnerMarkerSha256 -ceq $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256) {
        throw 'Bootstrap private context must keep distinct task and offline-ingress owner-marker bindings.'
    }
    if ($BootstrapPrivateContext.PrivateDockerClientPath -cne $dockerClientPath -or
        $BootstrapPrivateContext.PrivateHelperBuildTag -cne $script:HelperBuildTag) {
        throw 'Bootstrap private context has a substituted Docker client or helper build tag.'
    }
    foreach ($pathName in @('PrivateReportRoot', 'PrivateIngressRoot', 'PrivateIngressSourceRoot', 'PrivateIngressDestinationRoot')) {
        if ($BootstrapPrivateContext[$pathName] -isnot [string] -or [string]::IsNullOrWhiteSpace($BootstrapPrivateContext[$pathName])) {
            throw "Bootstrap private context has an invalid $pathName path binding."
        }
    }
    foreach ($bytesName in @('PrivateIngressCopyReceipt', 'PrivateDependencyBundleManifest', 'PrivateLocalBaseReceipt', 'PrivateHelperImageAttestation')) {
        if ($BootstrapPrivateContext[$bytesName] -isnot [byte[]] -or $BootstrapPrivateContext[$bytesName].Length -lt 1) {
            throw "Bootstrap private context lacks $bytesName raw bytes."
        }
    }
    if ($BootstrapPrivateContext.PrivateIngressDigestMap -isnot [System.Collections.IDictionary] -or
        $BootstrapPrivateContext.PrivateIngressByteCounts -isnot [System.Collections.IDictionary] -or
        $BootstrapPrivateContext.PrivateIngressOrderedPaths -isnot [System.Array] -or
        $BootstrapPrivateContext.PrivateIngressOrderedPaths.Count -lt 2) {
        throw 'Bootstrap private context lacks the ordered ingress copy bindings.'
    }
    if ((Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateIngressCopyReceipt) -cne $BootstrapPrivateContext.PrivateIngressCopyReceiptSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateDependencyBundleManifest) -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateLocalBaseReceipt) -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateHelperImageAttestation) -cne $BootstrapPrivateContext.PrivateHelperImageAttestationSha256) {
        throw 'Bootstrap private context raw receipt, manifest, or helper-attestation digest does not match its held bytes.'
    }
    foreach ($imageIdName in @('PrivateHelperLocalImageId', 'PrivateBootstrapHelperPythonLocalImageId')) {
        Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext[$imageIdName]) -Identity $imageIdName
    }
    if ((Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageId))) -cne
        $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageIdSha256) {
        throw 'Bootstrap private context helper-base local image ID does not match its held digest.'
    }

    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    $reportRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateReportRoot)
    $ingressRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateIngressRoot)
    $expectedIngressRoot = [System.IO.Path]::Combine($reportRoot, 'ingress', 'dependency-bundle')
    $destinationRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateIngressDestinationRoot)
    if (-not [string]::Equals($ingressRoot, $expectedIngressRoot, $comparison) -or
        -not [string]::Equals($destinationRoot, $expectedIngressRoot, $comparison)) {
        throw 'Bootstrap private ingress destination is not the fixed task-owned copied root.'
    }
    foreach ($path in @($reportRoot, $ingressRoot)) {
        $item = [System.IO.DirectoryInfo]::new($path)
        $item.Refresh()
        if (-not $item.Exists -or
            ($item.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
            ($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'Bootstrap private report or copied ingress root is not a physical directory.'
        }
    }
    return [pscustomobject][ordered]@{
        ReportRoot = $reportRoot
        IngressRoot = $ingressRoot
        ReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'g0-admission-helper.admission.v1.json')
        IngressCopyReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'ingress-copy-receipt.v1.json')
        LocalBaseReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'local-base-inspect.v1.json')
        HelperImageAttestationPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'g0-admission-helper.local-image-attestation.v1.json')
        HelperContainerName = 'hub-qa-' + [string]$BootstrapPrivateContext.Task + '-' + [string]$BootstrapPrivateContext.Run + '-admission-helper'
        HelperOwnerLabel = 'hub.qa.owner-marker-sha256=' + [string]$BootstrapPrivateContext.OwnerMarkerSha256
    }
}

function Assert-HubQaSafeIngressRelativePath {
    param([Parameter(Mandatory = $true)][string]$RelativePath)

    if ($RelativePath.Length -eq 0 -or $RelativePath.Contains('\') -or $RelativePath.StartsWith('/') -or $RelativePath.Split('/') -contains '..') {
        throw 'A runner-held ingress path is unsafe.'
    }
    foreach ($part in $RelativePath.Split('/')) {
        if ($part.Length -eq 0 -or $part -ceq '.') {
            throw 'A runner-held ingress path is unsafe.'
        }
    }
}

function Assert-HubQaClosedCopiedIngressTree {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    $expectedDirectories = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    $seenFiles = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    foreach ($relativePath in $BootstrapPrivateContext.PrivateIngressOrderedPaths) {
        if ($relativePath -isnot [string]) {
            throw 'The runner-held ordered ingress paths contain a non-string item.'
        }
        Assert-HubQaSafeIngressRelativePath -RelativePath $relativePath
        if (-not $seenFiles.Add($relativePath)) {
            throw 'The runner-held ordered ingress paths contain a duplicate item.'
        }
        if (-not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains($relativePath) -or
            -not $BootstrapPrivateContext.PrivateIngressByteCounts.Contains($relativePath)) {
            throw 'The runner-held ordered ingress paths are not closed over the digest and byte maps.'
        }
        $parts = $relativePath.Split('/')
        for ($index = 0; $index -lt ($parts.Count - 1); $index++) {
            [void]$expectedDirectories.Add(($parts[0..$index] -join '/'))
        }
    }
    if ($BootstrapPrivateContext.PrivateIngressDigestMap.Count -ne $seenFiles.Count -or
        $BootstrapPrivateContext.PrivateIngressByteCounts.Count -ne $seenFiles.Count) {
        throw 'The runner-held ingress digest or byte map count is not closed.'
    }
    if (-not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains('handoff-owner.v1.json') -or
        $BootstrapPrivateContext.PrivateIngressDigestMap['handoff-owner.v1.json'] -cne $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256 -or
        -not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains('dependency-bundle.manifest.v1.json') -or
        $BootstrapPrivateContext.PrivateIngressDigestMap['dependency-bundle.manifest.v1.json'] -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256) {
        throw 'The runner-held ingress map does not bind the copied handoff owner and dependency manifest bytes.'
    }

    $actualDirectories = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    foreach ($directory in ([System.IO.DirectoryInfo]::new($Closed.IngressRoot)).GetDirectories('*', [System.IO.SearchOption]::AllDirectories)) {
        if (($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A copied ingress directory is a reparse point.'
        }
        $relativeDirectory = $directory.FullName.Substring($Closed.IngressRoot.Length).TrimStart('\').Replace('\', '/')
        [void]$actualDirectories.Add($relativeDirectory)
    }
    if ($actualDirectories.Count -ne $expectedDirectories.Count) {
        throw 'The copied ingress tree has an unexpected empty or foreign directory.'
    }
    foreach ($expectedDirectory in $expectedDirectories) {
        if (-not $actualDirectories.Contains($expectedDirectory)) {
            throw 'The copied ingress tree has an unexpected empty or foreign directory.'
        }
    }

    $actualFiles = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    foreach ($file in ([System.IO.DirectoryInfo]::new($Closed.IngressRoot)).GetFiles('*', [System.IO.SearchOption]::AllDirectories)) {
        if (($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A copied ingress file is a reparse point.'
        }
        $relativeFile = $file.FullName.Substring($Closed.IngressRoot.Length).TrimStart('\').Replace('\', '/')
        [void]$actualFiles.Add($relativeFile)
    }
    if ($actualFiles.Count -ne $seenFiles.Count) {
        throw 'The copied ingress tree has an unexpected or missing file.'
    }
    foreach ($expectedFile in $seenFiles) {
        if (-not $actualFiles.Contains($expectedFile)) {
            throw 'The copied ingress tree has an unexpected or missing file.'
        }
    }

    $relativeLines = [System.Collections.Generic.List[string]]::new()
    foreach ($relativePath in $BootstrapPrivateContext.PrivateIngressOrderedPaths) {
        $expectedDigest = $BootstrapPrivateContext.PrivateIngressDigestMap[$relativePath]
        $expectedBytes = $BootstrapPrivateContext.PrivateIngressByteCounts[$relativePath]
        Assert-HubQaDigest -Value ([string]$expectedDigest) -Identity 'runner-held ingress digest'
        if (($expectedBytes -isnot [int] -and $expectedBytes -isnot [long]) -or $expectedBytes -lt 1) {
            throw 'The runner-held ingress byte count is invalid.'
        }
        $candidate = [System.IO.Path]::GetFullPath((Join-Path $Closed.IngressRoot ($relativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))))
        if (-not $candidate.StartsWith($Closed.IngressRoot + '\', $comparison)) {
            throw 'A runner-held ingress path escapes the fixed copied root.'
        }
        $bytes = Read-HubQaRegularFileBytes -Path $candidate
        if ($bytes.Length -ne [int64]$expectedBytes -or (Get-HubQaSha256 -Bytes $bytes) -cne $expectedDigest) {
            throw 'A copied ingress file does not match the held byte count or digest map.'
        }
        $relativeLines.Add($relativePath + "`t" + $expectedDigest + "`n")
    }
    $digestMapBytes = [System.Text.Encoding]::UTF8.GetBytes([string]::Concat($relativeLines))
    if ((Get-HubQaSha256 -Bytes $digestMapBytes) -cne $BootstrapPrivateContext.PrivateIngressDigestMapSha256) {
        throw 'The ordered private ingress digest map does not match its held digest.'
    }
}

function Assert-HubQaIngressCopyReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $actualBytes = Read-HubQaRegularFileBytes -Path $Closed.IngressCopyReceiptPath
    if (-not (Test-HubQaByteArrayEqual -Left $actualBytes -Right $BootstrapPrivateContext.PrivateIngressCopyReceipt)) {
        throw 'The on-disk ingress-copy receipt is not byte-identical to the runner-held receipt.'
    }
    $receipt = ConvertFrom-HubQaStrictUtf8Json -Bytes $BootstrapPrivateContext.PrivateIngressCopyReceipt -Identity 'runner-held ingress-copy receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'task', 'run', 'owner_marker_sha256', 'offline_ingress_handoff_id', 'source_commit',
        'source_owner_marker_sha256', 'dependency_bundle_manifest_sha256', 'source_root', 'destination_root',
        'paths', 'digest_map_sha256', 'local_base_receipt_sha256'
    ) -Identity 'runner-held ingress-copy receipt'
    if ($receipt.schema -cne 'hub.qa.g0-ingress-copy-receipt/v1' -or
        $receipt.task -cne $BootstrapPrivateContext.Task -or
        $receipt.run -cne $BootstrapPrivateContext.Run -or
        $receipt.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $receipt.offline_ingress_handoff_id -cne $BootstrapPrivateContext.PrivateOfflineIngressHandoffId -or
        $receipt.source_commit -cne $BootstrapPrivateContext.PrivateSourceCommit -or
        $receipt.source_owner_marker_sha256 -cne $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256 -or
        $receipt.dependency_bundle_manifest_sha256 -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        $receipt.source_root -cne $BootstrapPrivateContext.PrivateIngressSourceRoot -or
        $receipt.destination_root -cne $BootstrapPrivateContext.PrivateIngressDestinationRoot -or
        $receipt.digest_map_sha256 -cne $BootstrapPrivateContext.PrivateIngressDigestMapSha256 -or
        $receipt.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256) {
        throw 'The runner-held ingress-copy receipt is not closed over the private bootstrap context.'
    }
    if ($receipt.paths -isnot [System.Array] -or $receipt.paths.Count -ne $BootstrapPrivateContext.PrivateIngressOrderedPaths.Count) {
        throw 'The runner-held ingress-copy receipt path count is not closed.'
    }
    for ($index = 0; $index -lt $receipt.paths.Count; $index++) {
        $entry = $receipt.paths[$index]
        Assert-HubQaExactPropertyNames -Object $entry -Expected @('relative_path', 'bytes', 'sha256') -Identity 'runner-held ingress-copy receipt path'
        $relativePath = $BootstrapPrivateContext.PrivateIngressOrderedPaths[$index]
        if ($entry.relative_path -cne $relativePath -or
            $entry.bytes -ne $BootstrapPrivateContext.PrivateIngressByteCounts[$relativePath] -or
            $entry.sha256 -cne $BootstrapPrivateContext.PrivateIngressDigestMap[$relativePath]) {
            throw 'The runner-held ingress-copy receipt paths are not in the held ordered digest-map order.'
        }
    }
}

function Assert-HubQaPrivateManifestAndBaseReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $a1ManifestPath = Join-Path $PSScriptRoot 'dependency-bundle.manifest.v1.json'
    $copiedManifestPath = Join-Path $Closed.IngressRoot 'dependency-bundle.manifest.v1.json'
    foreach ($path in @($a1ManifestPath, $copiedManifestPath)) {
        $actualManifest = Read-HubQaRegularFileBytes -Path $path
        if (-not (Test-HubQaByteArrayEqual -Left $actualManifest -Right $BootstrapPrivateContext.PrivateDependencyBundleManifest)) {
            throw 'The A1 or copied dependency manifest is not byte-identical to the runner-held manifest bytes.'
        }
    }
    $actualBaseReceipt = Read-HubQaRegularFileBytes -Path $Closed.LocalBaseReceiptPath
    if (-not (Test-HubQaByteArrayEqual -Left $actualBaseReceipt -Right $BootstrapPrivateContext.PrivateLocalBaseReceipt)) {
        throw 'The on-disk local-base receipt is not byte-identical to the runner-held receipt.'
    }
    $baseReceipt = ConvertFrom-HubQaStrictUtf8Json -Bytes $BootstrapPrivateContext.PrivateLocalBaseReceipt -Identity 'runner-held local-base receipt'
    Assert-HubQaExactPropertyNames -Object $baseReceipt -Expected @('schema', 'task', 'run', 'owner_marker_sha256', 'receipt_nonce', 'bases') -Identity 'runner-held local-base receipt'
    if ($baseReceipt.schema -cne 'hub.qa.local-base-inspect/v1' -or
        $baseReceipt.task -cne $BootstrapPrivateContext.Task -or
        $baseReceipt.run -cne $BootstrapPrivateContext.Run -or
        $baseReceipt.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $baseReceipt.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $baseReceipt.bases -isnot [System.Array] -or $baseReceipt.bases.Count -ne 4) {
        throw 'The runner-held local-base receipt is not bound to the private task, owner, nonce, and base count.'
    }
}

function Assert-HubQaPrivateHelperImageAttestation {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $actualBytes = Read-HubQaRegularFileBytes -Path $Closed.HelperImageAttestationPath
    if (-not (Test-HubQaByteArrayEqual -Left $actualBytes -Right $BootstrapPrivateContext.PrivateHelperImageAttestation)) {
        throw 'The on-disk helper image attestation is not byte-identical to the runner-held attestation.'
    }
    $attestation = ConvertFrom-HubQaStrictUtf8Json -Bytes $BootstrapPrivateContext.PrivateHelperImageAttestation -Identity 'runner-held helper image attestation'
    Assert-HubQaExactPropertyNames -Object $attestation -Expected @(
        'schema', 'task', 'run', 'owner_marker_sha256', 'local_base_receipt_sha256', 'receipt_nonce',
        'helper_local_image_id', 'helper_local_image_id_sha256',
        'bootstrap_helper_python_local_image_id', 'bootstrap_helper_python_local_image_id_sha256',
        'helper_dockerfile_sha256', 'helper_entrypoint_sha256'
    ) -Identity 'runner-held helper image attestation'
    if ($attestation.schema -cne 'hub.qa.g0-admission-helper-local-image-attestation/v1' -or
        $attestation.task -cne $BootstrapPrivateContext.Task -or
        $attestation.run -cne $BootstrapPrivateContext.Run -or
        $attestation.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $attestation.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        $attestation.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $attestation.helper_local_image_id -cne $BootstrapPrivateContext.PrivateHelperLocalImageId -or
        $attestation.bootstrap_helper_python_local_image_id -cne $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageId -or
        $attestation.bootstrap_helper_python_local_image_id_sha256 -cne $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageIdSha256) {
        throw 'The runner-held helper image attestation is not bound to the private bootstrap task, receipt, nonce, or local image IDs.'
    }
    foreach ($digestName in @(
        'owner_marker_sha256', 'local_base_receipt_sha256', 'helper_local_image_id_sha256',
        'bootstrap_helper_python_local_image_id_sha256', 'helper_dockerfile_sha256', 'helper_entrypoint_sha256'
    )) {
        Assert-HubQaDigest -Value ([string]$attestation.$digestName) -Identity ('helper image attestation ' + $digestName)
    }
    if ((Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$attestation.helper_local_image_id))) -cne $attestation.helper_local_image_id_sha256 -or
        (Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$attestation.bootstrap_helper_python_local_image_id))) -cne $attestation.bootstrap_helper_python_local_image_id_sha256) {
        throw 'The runner-held helper image attestation local image ID digest does not match its exact full image ID.'
    }
}

function Assert-HubQaHelperAdmissionReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $bytes = Read-HubQaRegularFileBytes -Path $Closed.ReceiptPath
    $receipt = ConvertFrom-HubQaStrictUtf8Json -Bytes $bytes -Identity 'helper admission receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases',
        'local_base_receipt_sha256', 'receipt_nonce', 'verdict'
    ) -Identity 'helper admission receipt'
    if ($receipt.schema -cne 'hub.qa.g0-dependency-admission-receipt/v1' -or
        $receipt.source_commit -cne $script:SourceCommit -or
        $receipt.manifest_sha256 -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        $receipt.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        $receipt.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $receipt.verdict -cne 'ACCEPT' -or
        $receipt.assets -isnot [System.Array] -or $receipt.assets.Count -ne 4 -or
        $receipt.bases -isnot [System.Array] -or $receipt.bases.Count -ne 4) {
        throw 'The helper admission receipt is not bound to the verified private inputs.'
    }
}

function Invoke-HubQaG0AdmissionFromBootstrapContext {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $closed = Assert-HubQaPrivateBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateHelperImageAttestation -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaIngressCopyReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaClosedCopiedIngressTree -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateManifestAndBaseReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    if ([System.IO.File]::Exists($closed.ReceiptPath) -or [System.IO.Directory]::Exists($closed.ReceiptPath)) {
        throw 'The fixed helper admission receipt path already exists.'
    }
    $dockerArguments = @(
        'run', '--pull', 'never', '--name', $closed.HelperContainerName, '--label', $closed.HelperOwnerLabel,
        '--network', 'none', '--read-only', '--cap-drop', 'ALL',
        '--security-opt', 'no-new-privileges:true', '--cpus', '2.0', '--memory', '2g', '--pids-limit', '256',
        '--mount', ('type=bind,src=' + $closed.ReportRoot + ',dst=/work'),
        '--env', ('HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_SHA256=' + $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256),
        '--env', ('HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_NONCE=' + $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce),
        '--env', ('HUB_QA_PRIVATE_OWNER_MARKER_SHA256=' + $BootstrapPrivateContext.OwnerMarkerSha256),
        '--env', ('HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256=' + $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256),
        $BootstrapPrivateContext.PrivateHelperLocalImageId
    )
    & $BootstrapPrivateContext.PrivateDockerClientPath @dockerArguments
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw ('The fixed g0-admission-helper exited with code ' + $exitCode + '.')
    }
    Assert-HubQaHelperAdmissionReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    return $closed.ReceiptPath
}
