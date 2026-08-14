Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

$script:HubQaSnapshotSourceCommit = '61b38573682d7c01e6d9381babc5085537d7990f'

function Get-HubQaSnapshotSha256 {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)
    return Get-HubQaEnvelopeSha256 -Bytes $Bytes
}

function Assert-HubQaSnapshotKeys {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Object,
        [Parameter(Mandatory = $true)][string[]]$Expected
    )

    $actual = @($Object.Keys)
    if ($actual.Count -ne $Expected.Count) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ([string]$actual[$index] -cne [string]$Expected[$index]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    }
}

function Get-HubQaSnapshotNodeString {
    param([Parameter(Mandatory = $true)][object]$Value)
    if ($Value -isnot [string]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return [string]$Value
}

function Assert-HubQaSnapshotDigest {
    param([Parameter(Mandatory = $true)][object]$Value)
    $digest = Get-HubQaSnapshotNodeString -Value $Value
    if ($digest -cnotmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return $digest
}

function Assert-HubQaSnapshotObjectId {
    param([Parameter(Mandatory = $true)][object]$Value)
    $objectId = Get-HubQaSnapshotNodeString -Value $Value
    if ($objectId -cnotmatch '^(?!0{40}$)[a-f0-9]{40}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return $objectId
}

function Assert-HubQaSnapshotUtc {
    param([Parameter(Mandatory = $true)][object]$Value)
    $utc = Get-HubQaSnapshotNodeString -Value $Value
    if ($utc -cnotmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,7})?Z$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    try { [void][DateTimeOffset]::Parse($utc, [Globalization.CultureInfo]::InvariantCulture, ([Globalization.DateTimeStyles]::AssumeUniversal -bor [Globalization.DateTimeStyles]::AdjustToUniversal)) } catch { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return $utc
}

function Assert-HubQaSnapshotRepoPath {
    param([Parameter(Mandatory = $true)][object]$Value)
    $path = Get-HubQaSnapshotNodeString -Value $Value
    if ($path -cnotmatch '^(?!/)(?![A-Za-z]:)(?!.*(?:^|/)\.{1,2}(?:/|$))(?!.*//)[^\x00-\x1f\\]+$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return $path
}

function Assert-HubQaSnapshotPathEntries {
    param(
        [Parameter(Mandatory = $true)][object]$Entries,
        [Parameter(Mandatory = $true)][string]$StateKey
    )

    if ($null -eq $Entries -or $Entries -isnot [System.Array]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $previous = $null
    foreach ($entry in @($Entries)) {
        if ($entry -isnot [System.Collections.IDictionary]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        Assert-HubQaSnapshotKeys -Object $entry -Expected @('path',$StateKey,'blob_sha256')
        $path = Assert-HubQaSnapshotRepoPath -Value $entry['path']
        if ($null -ne $previous -and [string]::CompareOrdinal($previous, $path) -ge 0) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        $previous = $path
        $state = Get-HubQaSnapshotNodeString -Value $entry[$StateKey]
        if ($state -cnotmatch '^[A-Z?][A-Z0-9?]{0,7}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        [void](Assert-HubQaSnapshotDigest -Value $entry['blob_sha256'])
    }
}

function Assert-HubQaSnapshotCommitEntries {
    param([Parameter(Mandatory = $true)][object]$Commits)

    if ($null -eq $Commits -or $Commits -isnot [System.Array]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    foreach ($entry in @($Commits)) {
        if ($entry -isnot [System.Collections.IDictionary]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        Assert-HubQaSnapshotKeys -Object $entry -Expected @('commit','parent','tree','changed_paths')
        $commit = Assert-HubQaSnapshotObjectId -Value $entry['commit']
        [void](Assert-HubQaSnapshotObjectId -Value $entry['tree'])
        if ($null -ne $entry['parent']) { $parent = Assert-HubQaSnapshotObjectId -Value $entry['parent']; if ($parent -ceq $commit) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' } }
        Assert-HubQaSnapshotPathEntries -Entries $entry['changed_paths'] -StateKey 'status'
    }
}

function Assert-HubQaSnapshotAncestryEntries {
    param([Parameter(Mandatory = $true)][object]$Edges)

    if ($null -eq $Edges -or $Edges -isnot [System.Array]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    foreach ($edge in @($Edges)) {
        if ($edge -isnot [System.Collections.IDictionary]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        Assert-HubQaSnapshotKeys -Object $edge -Expected @('child','parent')
        $child = Assert-HubQaSnapshotObjectId -Value $edge['child']
        $parent = Assert-HubQaSnapshotObjectId -Value $edge['parent']
        if ($child -ceq $parent) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    }
}

function Get-HubQaSnapshotCollectionDigest {
    param([Parameter(Mandatory = $true)][object]$Value)
    [byte[]]$bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $Value))
    return Get-HubQaSnapshotSha256 -Bytes $bytes
}

function Assert-HubQaSnapshotEvidence {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Context,
        [Parameter(Mandatory = $true)][string]$PathKey,
        [Parameter(Mandatory = $true)][string]$RawKey,
        [Parameter(Mandatory = $true)][string]$DigestKey,
        [Parameter(Mandatory = $true)][string]$ExpectedPath
    )

    $path = Assert-HubQaSnapshotRepoPath -Value $Context[$PathKey]
    if ($path -cne $ExpectedPath) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    if ($Context[$RawKey] -isnot [byte[]]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    [byte[]]$raw = $Context[$RawKey]
    if ($raw.Length -eq 0) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $digest = Assert-HubQaSnapshotDigest -Value $Context[$DigestKey]
    if ((Get-HubQaSnapshotSha256 -Bytes $raw) -cne $digest) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $root = Get-HubQaEnvelopeRepositoryRoot
    try { [byte[]]$stored = (Get-HubQaEnvelopeSafeContainedBytes -Root $root -RelativePath $path -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK').Bytes } catch { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    if (-not [System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($stored, $raw)) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return [ordered]@{ path = $path; raw_sha256 = $digest }
}

function New-HubQaRepositorySnapshot {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$BootstrapPrivateContext,
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$SnapshotPrivateContext
    )

    Assert-HubQaSnapshotKeys -Object $BootstrapPrivateContext -Expected @('Task','Run','RoleClaim','OwnerNonce','OwnerMarkerRawBytes','OwnerMarkerSha256')
    Assert-HubQaSnapshotKeys -Object $SnapshotPrivateContext -Expected @('PositiveGreenPath','PositiveGreenRawBytes','PositiveGreenSha256','CandidateLockPath','CandidateLockRawBytes','CandidateLockSha256','CleanupPath','CleanupRawBytes','CleanupSha256','G0PlanSha256','BootstrapContractSha256','RunnerContractSha256','DependencyBundleManifestSha256','SourcePolicySha256','BuildManifestSha256','SourceRevisionTree','HeadCommit','HeadTree','IndexPaths','WorktreePaths','Commits','Ancestry','StartedAtUtc','FinishedAtUtc')

    $task = Get-HubQaSnapshotNodeString -Value $BootstrapPrivateContext['Task']
    $run = Get-HubQaSnapshotNodeString -Value $BootstrapPrivateContext['Run']
    $role = Get-HubQaSnapshotNodeString -Value $BootstrapPrivateContext['RoleClaim']
    $nonce = Get-HubQaSnapshotNodeString -Value $BootstrapPrivateContext['OwnerNonce']
    if ($BootstrapPrivateContext['OwnerMarkerRawBytes'] -isnot [byte[]]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    [byte[]]$markerBytes = $BootstrapPrivateContext['OwnerMarkerRawBytes']
    $markerDigest = Assert-HubQaSnapshotDigest -Value $BootstrapPrivateContext['OwnerMarkerSha256']
    if ($task -cne 'g0-assets-runner' -or $task -cnotmatch '^[a-z][a-z0-9-]{2,31}$' -or $run -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or $role -cne 'hub.qa-bootstrap/v1' -or $nonce -cnotmatch '^[0-9a-f]{32}$' -or $null -eq $markerBytes -or $markerBytes.Length -eq 0 -or (Get-HubQaSnapshotSha256 -Bytes $markerBytes) -cne $markerDigest) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }

    $reportRoot = 'testframework/reports/' + $task + '-' + $run
    $positiveGreen = Assert-HubQaSnapshotEvidence -Context $SnapshotPrivateContext -PathKey 'PositiveGreenPath' -RawKey 'PositiveGreenRawBytes' -DigestKey 'PositiveGreenSha256' -ExpectedPath ($reportRoot + '/selftests/g0-planning-positive-green.v1.json')
    $candidateLock = Assert-HubQaSnapshotEvidence -Context $SnapshotPrivateContext -PathKey 'CandidateLockPath' -RawKey 'CandidateLockRawBytes' -DigestKey 'CandidateLockSha256' -ExpectedPath ($reportRoot + '/bootstrap/images.lock.v3.json')
    $cleanup = Assert-HubQaSnapshotEvidence -Context $SnapshotPrivateContext -PathKey 'CleanupPath' -RawKey 'CleanupRawBytes' -DigestKey 'CleanupSha256' -ExpectedPath ($reportRoot + '/cleanup.json')

    $control = [ordered]@{}
    foreach ($pair in @(@('g0_plan_sha256','G0PlanSha256'),@('bootstrap_contract_sha256','BootstrapContractSha256'),@('runner_contract_sha256','RunnerContractSha256'),@('dependency_bundle_manifest_sha256','DependencyBundleManifestSha256'),@('source_policy_sha256','SourcePolicySha256'),@('build_manifest_sha256','BuildManifestSha256'))) {
        $control[$pair[0]] = Assert-HubQaSnapshotDigest -Value $SnapshotPrivateContext[$pair[1]]
    }
    $control['images_lock_sha256'] = $null

    $sourceTree = Assert-HubQaSnapshotObjectId -Value $SnapshotPrivateContext['SourceRevisionTree']
    $headCommit = Assert-HubQaSnapshotObjectId -Value $SnapshotPrivateContext['HeadCommit']
    $headTree = Assert-HubQaSnapshotObjectId -Value $SnapshotPrivateContext['HeadTree']
    $startedAt = Assert-HubQaSnapshotUtc -Value $SnapshotPrivateContext['StartedAtUtc']
    $finishedAt = Assert-HubQaSnapshotUtc -Value $SnapshotPrivateContext['FinishedAtUtc']
    $startedInstant = [DateTimeOffset]::Parse($startedAt, [Globalization.CultureInfo]::InvariantCulture, ([Globalization.DateTimeStyles]::AssumeUniversal -bor [Globalization.DateTimeStyles]::AdjustToUniversal))
    $finishedInstant = [DateTimeOffset]::Parse($finishedAt, [Globalization.CultureInfo]::InvariantCulture, ([Globalization.DateTimeStyles]::AssumeUniversal -bor [Globalization.DateTimeStyles]::AdjustToUniversal))
    if ($startedInstant -gt $finishedInstant) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }

    $indexPathContext = $SnapshotPrivateContext['IndexPaths']
    $worktreePathContext = $SnapshotPrivateContext['WorktreePaths']
    $commitContext = $SnapshotPrivateContext['Commits']
    $ancestryContext = $SnapshotPrivateContext['Ancestry']
    Assert-HubQaSnapshotPathEntries -Entries $indexPathContext -StateKey 'status'
    Assert-HubQaSnapshotPathEntries -Entries $worktreePathContext -StateKey 'state'
    Assert-HubQaSnapshotCommitEntries -Commits $commitContext
    Assert-HubQaSnapshotAncestryEntries -Edges $ancestryContext
    $indexPaths = @($indexPathContext)
    $worktreePaths = @($worktreePathContext)
    $commits = @($commitContext)
    $edges = @($ancestryContext)
    if ($commits.Count -ne 1 -or $edges.Count -ne 1) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $commit = [System.Collections.IDictionary]$commits[0]
    $edge = [System.Collections.IDictionary]$edges[0]
    if ($null -eq $commit['parent'] -or [string]$commit['commit'] -cne $headCommit -or [string]$commit['tree'] -cne $headTree -or [string]$edge['child'] -cne $headCommit -or [string]$edge['parent'] -cne [string]$commit['parent']) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }

    $root = Get-HubQaEnvelopeRepositoryRoot
    $snapshotRelative = $reportRoot + '/snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json'
    $snapshotParentChain = Open-HubQaEnvelopeHeldDirectoryPath -Root $root -RelativeDirectoryPath ($reportRoot + '/snapshots') -CreateLeaf -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    try { Assert-HubQaEnvelopeHeldDirectoryChain -Chain $snapshotParentChain -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK' } finally { Close-HubQaEnvelopeHeldDirectoryChain -Chain $snapshotParentChain }
    $payload = [ordered]@{
        schema = 'hub.qa.repository-snapshot/v1'
        producer = 'hub.qa.repository-snapshot-producer/v1'
        snapshot_id = $task + '-snapshot-' + $run + '-000-G0_BOOTSTRAP'
        captured_for = [ordered]@{ kind='G0_BOOTSTRAP'; ordinal=0; subject_path=$null; subject_sha256=$null; subject_commit_sha=$null; predecessor_event_path=$null; predecessor_event_sha256=$null; predecessor_event_commit_sha=$null; pre_event_snapshot_path=$null; pre_event_snapshot_sha256=$null }
        task = $task
        run = $run
        envelope_id = $task + '-envelope-' + $run
        role_claim = $role
        self_attested = $true
        source_revision = [ordered]@{ commit=$script:HubQaSnapshotSourceCommit; tree=$sourceTree; captured_at_utc=$finishedAt }
        head = [ordered]@{ commit=$headCommit; tree=$headTree; event=[ordered]@{ event_id=$null; event_type=$null; captured_at_utc=$null } }
        index = [ordered]@{ paths=$indexPaths; count=$indexPaths.Count; sha256=(Get-HubQaSnapshotCollectionDigest -Value $indexPaths) }
        worktree = [ordered]@{ scoped_paths=$worktreePaths; count=$worktreePaths.Count; sha256=(Get-HubQaSnapshotCollectionDigest -Value $worktreePaths) }
        commits = $commits
        ancestry = [ordered]@{ edges=$edges }
        control = $control
        evidence = [ordered]@{ positive_green_path=$positiveGreen.path; positive_green_sha256=$positiveGreen.raw_sha256; candidate_lock_path=$candidateLock.path; candidate_lock_sha256=$candidateLock.raw_sha256; cleanup_path=$cleanup.path; cleanup_sha256=$cleanup.raw_sha256 }
        lifecycle = [ordered]@{ state='CAPTURED'; started_at_utc=$startedAt; finished_at_utc=$finishedAt }
    }
    [byte[]]$bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $payload))
    $parsed = ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $bytes
    [void](Assert-HubQaSnapshotEnvelopePayload -Node $parsed -Kind snapshot -ExpectedTask $task -ExpectedRun $run)
    $written = Write-HubQaEnvelopeSafeCreateNew -Root $root -RelativePath $snapshotRelative -Bytes $bytes -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    return [pscustomobject][ordered]@{ SnapshotPath=$snapshotRelative; SnapshotRawBytes=[byte[]]$written.Bytes; SnapshotSha256=(Get-HubQaSnapshotSha256 -Bytes $written.Bytes); SnapshotId=$payload.snapshot_id; EnvelopeId=$payload.envelope_id }
}

if ($MyInvocation.InvocationName -ne '.') { throw 'E_B0_DIRECT_INVOCATION_FORBIDDEN' }
