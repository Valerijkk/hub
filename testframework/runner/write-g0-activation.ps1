Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

. (Join-Path $PSScriptRoot 'new-run-envelope.ps1')

$script:HubQaActivationPlanPath = '.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md'
$script:HubQaActivationReleasePath = '.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json'
$script:HubQaActivationAcceptancePath = '.planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json'
$script:HubQaActivationOutputPath = '.planning/audits/test-rebase/g0-activation.v1.json'
$script:HubQaActivationTargets = @('backend-unit','frontend-unit','backend-openapi','g0-planning-selftest')
$script:HubQaActivationSourceCommit = '61b38573682d7c01e6d9381babc5085537d7990f'
$script:HubQaActivationA0Commit = '56a288dd51e261bbb91ed693d7b58a8a97994faf'
$script:HubQaActivationA02Paths = @('testframework/docker/dependency-bundle.schema.v1.json','testframework/docker/admission/verify_local_dependency_bundle.py','testframework/docker/admit-local-dependency-assets.ps1','testframework/docker/entrypoints/backend-unit.sh','testframework/docker/vitest.frontend-unit.config.ts','testframework/docker/bootstrap-control.v1.json','testframework/docker/source-image-attestor.ps1','testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1','testframework/selftests/runner/g0-image-contract.ps1')
$script:HubQaActivationA02OverlapPaths = @('testframework/docker/dependency-bundle.schema.v1.json','testframework/docker/admission/verify_local_dependency_bundle.py','testframework/docker/admit-local-dependency-assets.ps1','testframework/docker/bootstrap-control.v1.json','testframework/docker/source-image-attestor.ps1','testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1','testframework/selftests/runner/g0-image-contract.ps1')
$script:HubQaActivationA01RemainingPaths = @('testframework/docker/admission/admit_local_dependency_assets.py','testframework/docker/Dockerfile.backend-unit','testframework/docker/Dockerfile.frontend-unit','testframework/docker/Dockerfile.backend-openapi','testframework/docker/entrypoints/frontend-unit.sh','testframework/docker/runner-fixtures/g0-planning-selftest.py')

function Get-HubQaActivationRoot {
    try { return Get-HubQaEnvelopeRepositoryRoot } catch { throw 'E_B0_ACTIVATION_INPUT' }
}

function Test-HubQaActivationCommit {
    param([string]$Value)
    return $Value -cmatch '^[a-f0-9]{40}$'
}

function Test-HubQaActivationDigest {
    param([string]$Value)
    return $Value -cmatch '^sha256:[a-f0-9]{64}$'
}

function Test-HubQaActivationRepoPath {
    param([string]$Value)
    return $Value -cmatch '^(?!/)(?![A-Za-z]:)(?!.*(?:^|/)\.{1,2}(?:/|$))(?!.*//)[^\x00-\x1f\\]+$'
}

function ConvertTo-HubQaActivationGitArgument {
    param([Parameter(Mandatory = $true)][string]$Value)
    if ($Value.IndexOf('"') -ge 0 -or $Value.IndexOf("`r") -ge 0 -or $Value.IndexOf("`n") -ge 0) { throw 'E_B0_ACTIVATION_INPUT' }
    return '"' + $Value + '"'
}

function Invoke-HubQaActivationGitBytes {
    param(
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [ValidateRange(1,30)][int]$TimeoutSeconds = 30
    )
    $wallBudgetMilliseconds = [int]($TimeoutSeconds * 1000)
    $wallClock = [Diagnostics.Stopwatch]::StartNew()
    $remainingWallBudget = {
        $remaining = [int64]$wallBudgetMilliseconds - [int64]$wallClock.ElapsedMilliseconds
        if ($remaining -le 0) { return [int]0 }
        return [int]$remaining
    }.GetNewClosure()
    $root = Get-HubQaActivationRoot
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = 'git'
    $psi.Arguments = '-C ' + (ConvertTo-HubQaActivationGitArgument -Value $root) + ' ' + (($Arguments | ForEach-Object { ConvertTo-HubQaActivationGitArgument -Value $_ }) -join ' ')
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.CreateNoWindow = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $psi
    try {
        if (-not $process.Start()) { throw 'E_B0_ACTIVATION_INPUT' }
        $stream = New-Object IO.MemoryStream
        try {
            $stdoutTask = $process.StandardOutput.BaseStream.CopyToAsync($stream)
            $stderrTask = $process.StandardError.ReadToEndAsync()
            $remaining = & $remainingWallBudget
            if ($remaining -le 0 -or -not $process.WaitForExit($remaining)) {
                try { $process.Kill() } catch {}
                $remaining = & $remainingWallBudget
                if ($remaining -gt 0) { [void]$process.WaitForExit($remaining) }
                throw 'E_B0_ACTIVATION_INPUT: exit 124'
            }
            $remaining = & $remainingWallBudget
            if ($remaining -le 0 -or -not $stdoutTask.Wait($remaining)) {
                try { $process.Kill() } catch {}
                $remaining = & $remainingWallBudget
                if ($remaining -gt 0) { [void]$process.WaitForExit($remaining) }
                throw 'E_B0_ACTIVATION_INPUT: exit 124'
            }
            $remaining = & $remainingWallBudget
            if ($remaining -le 0 -or -not $stderrTask.Wait($remaining)) {
                try { $process.Kill() } catch {}
                $remaining = & $remainingWallBudget
                if ($remaining -gt 0) { [void]$process.WaitForExit($remaining) }
                throw 'E_B0_ACTIVATION_INPUT: exit 124'
            }
            if ($process.ExitCode -ne 0) { throw 'E_B0_ACTIVATION_INPUT' }
            return [byte[]]$stream.ToArray()
        } finally { $stream.Dispose() }
    } finally { $process.Dispose() }
}

function Get-HubQaActivationGitText {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)
    try { $value = (Get-HubQaEnvelopeUtf8).GetString((Invoke-HubQaActivationGitBytes -Arguments $Arguments)) } catch { throw 'E_B0_ACTIVATION_INPUT' }
    return $value.TrimEnd("`r", "`n")
}

function Get-HubQaActivationGitBlob {
    param([Parameter(Mandatory = $true)][string]$Commit, [Parameter(Mandatory = $true)][string]$Path)
    if (-not (Test-HubQaActivationCommit -Value $Commit) -or -not (Test-HubQaActivationRepoPath -Value $Path)) { throw 'E_B0_ACTIVATION_INPUT' }
    try { return Invoke-HubQaActivationGitBytes -Arguments @('cat-file','blob',($Commit + ':' + $Path)) } catch { throw 'E_B0_ACTIVATION_INPUT' }
}

function Test-HubQaActivationGitBlobExists {
    param([Parameter(Mandatory = $true)][string]$Commit, [Parameter(Mandatory = $true)][string]$Path)
    if (-not (Test-HubQaActivationCommit -Value $Commit) -or -not (Test-HubQaActivationRepoPath -Value $Path)) { return $false }
    try {
        [void](Invoke-HubQaActivationGitBytes -Arguments @('cat-file','-e',($Commit + ':' + $Path)))
        return $true
    } catch {
        return $false
    }
}

function Assert-HubQaActivationCommitExists {
    param([Parameter(Mandatory = $true)][string]$Commit, [string]$Code = 'E_B0_ACTIVATION_INPUT')
    if (-not (Test-HubQaActivationCommit -Value $Commit)) { throw $Code }
    try { [void](Invoke-HubQaActivationGitBytes -Arguments @('cat-file','-e',($Commit + '^{commit}'))) } catch { throw $Code }
}

function Test-HubQaActivationAncestor {
    param([Parameter(Mandatory = $true)][string]$Ancestor, [Parameter(Mandatory = $true)][string]$Descendant)
    try {
        [void](Invoke-HubQaActivationGitBytes -Arguments @('merge-base','--is-ancestor',$Ancestor,$Descendant))
        return $true
    } catch { return $false }
}

function Test-HubQaActivationFirstParentMember {
    param([Parameter(Mandatory = $true)][string]$Commit, [Parameter(Mandatory = $true)][string]$ExpectedMember)
    if (-not (Test-HubQaActivationCommit -Value $Commit) -or -not (Test-HubQaActivationCommit -Value $ExpectedMember)) { return $false }
    try {
        $history = @(Get-HubQaActivationGitText -Arguments @('rev-list','--first-parent',$Commit) -split "`n" | ForEach-Object { $_.TrimEnd("`r") })
        return $history -ccontains $ExpectedMember
    } catch { return $false }
}

function Get-HubQaActivationCommitPaths {
    param([Parameter(Mandatory = $true)][string]$Commit)
    try {
        $text = Get-HubQaActivationGitText -Arguments @('diff-tree','--no-commit-id','--name-only','-r',$Commit)
        if ([string]::IsNullOrEmpty($text)) { return @() }
        return @($text -split "`n" | ForEach-Object { $_.TrimEnd("`r") })
    } catch { throw 'E_B0_ACTIVATION_INPUT' }
}

function Get-HubQaActivationDiffPaths {
    param([Parameter(Mandatory = $true)][string]$From, [Parameter(Mandatory = $true)][string]$To, [Parameter(Mandatory = $true)][string]$Code)
    if (-not (Test-HubQaActivationCommit -Value $From) -or -not (Test-HubQaActivationCommit -Value $To)) { throw $Code }
    try {
        $text = Get-HubQaActivationGitText -Arguments @('diff','--no-renames','--name-only',$From,$To)
        if ([string]::IsNullOrEmpty($text)) { return @() }
        return @($text -split "`n" | ForEach-Object { $_.TrimEnd("`r") })
    } catch { throw $Code }
}

function Assert-HubQaActivationExactCommitPaths {
    param([Parameter(Mandatory = $true)][string]$Commit, [Parameter(Mandatory = $true)][string[]]$Expected, [Parameter(Mandatory = $true)][string]$Code)
    $actual = @(Get-HubQaActivationCommitPaths -Commit $Commit)
    if ($actual.Count -ne $Expected.Count) { throw $Code }
    foreach ($path in $Expected) { if ($actual -cnotcontains $path) { throw $Code } }
}

function Assert-HubQaActivationA02Authority {
    param(
        [Parameter(Mandatory = $true)][string]$A0Commit,
        [Parameter(Mandatory = $true)][string]$A02Commit,
        [Parameter(Mandatory = $true)][string]$CurrentCommit,
        [Parameter(Mandatory = $true)][string]$Code
    )
    if ($A0Commit -cne $script:HubQaActivationA0Commit -or -not (Test-HubQaActivationCommit -Value $A02Commit) -or -not (Test-HubQaActivationCommit -Value $CurrentCommit) -or -not (Test-HubQaActivationFirstParentMember -Commit $A02Commit -ExpectedMember $A0Commit)) { throw $Code }
    try { $actualParent = Get-HubQaActivationGitText -Arguments @('rev-parse',($A02Commit + '^1')) } catch { throw $Code }
    if (-not (Test-HubQaActivationCommit -Value $actualParent) -or -not (Test-HubQaActivationFirstParentMember -Commit $actualParent -ExpectedMember $script:HubQaActivationSourceCommit)) { throw $Code }
    $paths = @(Get-HubQaActivationDiffPaths -From $actualParent -To $A02Commit -Code $Code)
    if ($paths.Count -ne $script:HubQaActivationA02Paths.Count) { throw $Code }
    foreach ($path in $script:HubQaActivationA02Paths) { if ($paths -cnotcontains $path) { throw $Code } }
    try {
        foreach ($path in $script:HubQaActivationA02OverlapPaths) {
            if (-not [System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals((Get-HubQaActivationGitBlob -Commit $CurrentCommit -Path $path),(Get-HubQaActivationGitBlob -Commit $A02Commit -Path $path))) { throw $Code }
        }
        foreach ($path in $script:HubQaActivationA01RemainingPaths) {
            if (-not [System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals((Get-HubQaActivationGitBlob -Commit $CurrentCommit -Path $path),(Get-HubQaActivationGitBlob -Commit $A0Commit -Path $path))) { throw $Code }
        }
    } catch { throw $Code }
}

function Assert-HubQaActivationNodeObject {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$Code)
    if ([string]$Node.Type -cne 'object') { throw $Code }
    return [System.Collections.IDictionary]$Node.Value
}

function Assert-HubQaActivationNodeKeys {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string[]]$Expected, [Parameter(Mandatory = $true)][string]$Code)
    $value = Assert-HubQaActivationNodeObject -Node $Node -Code $Code
    $actual = @($value.Keys)
    if ($actual.Count -ne $Expected.Count) { throw $Code }
    for ($index = 0; $index -lt $Expected.Count; $index++) { if ([string]$actual[$index] -cne [string]$Expected[$index]) { throw $Code } }
    return $value
}

function Get-HubQaActivationNodeString {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$Code)
    if ([string]$Node.Type -cne 'string') { throw $Code }
    return [string]$Node.Value
}

function Get-HubQaActivationNodeBoolean {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$Code)
    if ([string]$Node.Type -cne 'boolean') { throw $Code }
    return [bool]$Node.Value
}

function Get-HubQaActivationNodeNull {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$Code)
    if ([string]$Node.Type -cne 'null') { throw $Code }
}

function Get-HubQaActivationCanonicalNode {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes, [Parameter(Mandatory = $true)][string]$Code, [switch]$AllowNonCanonical)
    try {
        if ($AllowNonCanonical) { return ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $Bytes -AllowNonCanonical }
        return ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $Bytes
    } catch { throw $Code }
}

function Get-HubQaActivationPhysicalBytes {
    param([Parameter(Mandatory = $true)][string]$RelativePath, [Parameter(Mandatory = $true)][string]$ReportRoot, [Parameter(Mandatory = $true)][string]$Code)
    if (-not (Test-HubQaActivationRepoPath -Value $RelativePath) -or -not $RelativePath.StartsWith($ReportRoot + '/', [StringComparison]::Ordinal)) { throw $Code }
    try { return [byte[]](Get-HubQaEnvelopeSafeContainedBytes -Root (Get-HubQaActivationRoot) -RelativePath $RelativePath -Code $Code).Bytes } catch { throw $Code }
}

function Assert-HubQaActivationReleaseNonceHistory {
    param([Parameter(Mandatory = $true)][string]$Nonce, [Parameter(Mandatory = $true)][string]$CurrentRun, [string]$Code = 'E_B0_ACTIVATION_RELEASE_BINDING')
    if ($Nonce -cnotmatch '^[0-9a-f]{32}$' -or $CurrentRun -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$') { throw $Code }
    $root = Get-HubQaActivationRoot
    $reports = Join-Path $root 'testframework\reports'
    if (-not (Test-Path -LiteralPath $reports)) { return }
    try { $reportsCanonical = Resolve-HubQaEnvelopeSafeDirectory -Path $reports -Root $root -Code $Code } catch { throw $Code }
    foreach ($directory in @(Get-ChildItem -LiteralPath $reportsCanonical -Directory -Force -ErrorAction Stop)) {
        if ($directory.Name -cnotmatch '^g0-assets-runner-([a-z0-9][a-z0-9-]{7,63})$') { continue }
        [void](Resolve-HubQaEnvelopeSafeDirectory -Path $directory.FullName -Root $root -Code $Code)
        $relative = 'testframework/reports/' + $directory.Name + '/run.json'
        if (-not (Test-Path -LiteralPath (Join-Path $directory.FullName 'run.json'))) { continue }
        $node = Get-HubQaActivationCanonicalNode -Bytes (Get-HubQaActivationPhysicalBytes -RelativePath $relative -ReportRoot ('testframework/reports/' + $directory.Name) -Code $Code) -Code $Code
        $value = Get-HubQaEnvelopeNodeObject -Node $node -Code $Code
        if (-not $value.Contains('release_one_shot_nonce')) { continue }
        $used = Get-HubQaEnvelopeNodeString -Node $value['release_one_shot_nonce'] -Code $Code
        if ($used -cnotmatch '^[0-9a-f]{32}$') { throw $Code }
        if ($used -ceq $Nonce -and $directory.Name -cne ('g0-assets-runner-' + $CurrentRun)) { throw $Code }
    }
}

function Get-HubQaActivationReceiptBytes {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$ExpectedPath, [Parameter(Mandatory = $true)][string]$ReportRoot, [Parameter(Mandatory = $true)][string]$Code)
    $value = Assert-HubQaActivationNodeKeys -Node $Node -Expected @('path','raw_sha256') -Code $Code
    $path = Get-HubQaActivationNodeString -Node $value['path'] -Code $Code
    $digest = Get-HubQaActivationNodeString -Node $value['raw_sha256'] -Code $Code
    if ($path -cne ($ReportRoot + '/' + $ExpectedPath) -or -not (Test-HubQaActivationDigest -Value $digest)) { throw $Code }
    [byte[]]$bytes = Get-HubQaActivationPhysicalBytes -RelativePath $path -ReportRoot $ReportRoot -Code $Code
    if ((Get-HubQaEnvelopeSha256 -Bytes $bytes) -cne $digest) { throw $Code }
    return [pscustomobject][ordered]@{ Path=$path; RawBytes=$bytes; RawSha256=$digest }
}

function Get-HubQaActivationNonceConsumptionReceipt {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][object]$Release, [Parameter(Mandatory = $true)][string]$Code)
    $value = Assert-HubQaActivationNodeKeys -Node $Node -Expected @('path','raw_sha256') -Code $Code
    $releaseID = Get-HubQaActivationNodeString -Node $Release.Value['release_id'] -Code $Code
    $expectedPath = 'testframework/reports/g0-assets-runner-nonce-consumptions/' + $releaseID + '.nonce-consumption.v1.json'
    $path = Get-HubQaActivationNodeString -Node $value['path'] -Code $Code
    $digest = Get-HubQaActivationNodeString -Node $value['raw_sha256'] -Code $Code
    if ($path -cne $expectedPath -or -not (Test-HubQaActivationDigest -Value $digest)) { throw $Code }
    try { [byte[]]$bytes = (Get-HubQaEnvelopeSafeContainedBytes -Root (Get-HubQaActivationRoot) -RelativePath $path -Code $Code).Bytes } catch { throw $Code }
    if ((Get-HubQaEnvelopeSha256 -Bytes $bytes) -cne $digest) { throw $Code }
    $record = Get-HubQaActivationCanonicalNode -Bytes $bytes -Code $Code
    $recordValue = Assert-HubQaActivationNodeKeys -Node $record -Expected @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status') -Code $Code
    if ((Get-HubQaActivationNodeString -Node $recordValue['schema'] -Code $Code) -cne 'hub.qa.g0-nonce-consumption/v1' -or (Get-HubQaActivationNodeString -Node $recordValue['task'] -Code $Code) -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $recordValue['run'] -Code $Code) -cne $Release.Run -or (Get-HubQaActivationNodeString -Node $recordValue['release_id'] -Code $Code) -cne $releaseID -or (Get-HubQaActivationNodeString -Node $recordValue['release_commit'] -Code $Code) -cne $Release.Commit -or (Get-HubQaActivationNodeString -Node $recordValue['release_raw_sha256'] -Code $Code) -cne $Release.RawSha256 -or (Get-HubQaActivationNodeString -Node $recordValue['one_shot_nonce'] -Code $Code) -cne (Get-HubQaActivationNodeString -Node $Release.Value['one_shot_nonce'] -Code $Code) -or (Get-HubQaActivationNodeString -Node $recordValue['status'] -Code $Code) -cne 'CONSUMED') { throw $Code }
    [void](Assert-HubQaEnvelopeUtc -Node $recordValue['consumed_at_utc'] -Code $Code)
    return [pscustomobject][ordered]@{ Path=$path; RawBytes=$bytes; RawSha256=$digest }
}

function Get-HubQaActivationRelease {
    param([Parameter(Mandatory = $true)][string]$FounderReleaseCommit, [Parameter(Mandatory = $true)][string]$RequiredPlanCommit)
    $code = 'E_B0_ACTIVATION_RELEASE_BINDING'
    Assert-HubQaActivationCommitExists -Commit $FounderReleaseCommit -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $FounderReleaseCommit -Expected @($script:HubQaActivationReleasePath) -Code $code
    [byte[]]$bytes = Get-HubQaActivationGitBlob -Commit $FounderReleaseCommit -Path $script:HubQaActivationReleasePath
    $root = Get-HubQaActivationCanonicalNode -Bytes $bytes -Code $code
    $value = Assert-HubQaActivationNodeKeys -Node $root -Expected @('schema','release_id','g0_plan_path','g0_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','task','run','local_only','network','offline_ingress_handoff_id','offline_ingress_owner_marker_sha256','dependency_bundle_manifest_sha256','source_commit','expires_at_utc','one_shot_nonce','founder_approval','created_at_utc') -Code $code
    if ((Get-HubQaActivationNodeString -Node $value['schema'] -Code $code) -cne 'hub.qa.g0-bootstrap-founder-release/v1' -or (Get-HubQaActivationNodeString -Node $value['g0_plan_path'] -Code $code) -cne $script:HubQaActivationPlanPath -or (Get-HubQaActivationNodeString -Node $value['g0_plan_commit'] -Code $code) -cne $RequiredPlanCommit -or (Get-HubQaActivationNodeString -Node $value['a0_commit'] -Code $code) -cne $script:HubQaActivationA0Commit -or (Get-HubQaActivationNodeString -Node $value['task'] -Code $code) -cne 'g0-assets-runner' -or -not (Get-HubQaActivationNodeBoolean -Node $value['local_only'] -Code $code) -or (Get-HubQaActivationNodeString -Node $value['network'] -Code $code) -cne 'disabled' -or (Get-HubQaActivationNodeString -Node $value['source_commit'] -Code $code) -cne $script:HubQaActivationSourceCommit -or (Get-HubQaActivationNodeString -Node $value['founder_approval'] -Code $code) -cne 'EXPLICIT_LOCAL_ONLY_ONE_SHOT') { throw $code }
    $run = Get-HubQaActivationNodeString -Node $value['run'] -Code $code
    $handoff = Get-HubQaActivationNodeString -Node $value['offline_ingress_handoff_id'] -Code $code
    $nonce = Get-HubQaActivationNodeString -Node $value['one_shot_nonce'] -Code $code
    $releaseId = Get-HubQaActivationNodeString -Node $value['release_id'] -Code $code
    $expires = Assert-HubQaEnvelopeUtc -Node $value['expires_at_utc'] -Code $code
    $created = Assert-HubQaEnvelopeUtc -Node $value['created_at_utc'] -Code $code
    if ($releaseId -cne ('g0-bootstrap-founder-release-' + $run) -or $run -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or $handoff -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or $nonce -cnotmatch '^[0-9a-f]{32}$' -or $created -gt [DateTime]::UtcNow -or $created -ge $expires -or [DateTime]::UtcNow -ge $expires) { throw $code }
    foreach ($field in @('a0_commit','a02_commit','a1_commit','b_commit')) {
        $commit = Get-HubQaActivationNodeString -Node $value[$field] -Code $code
        if (-not (Test-HubQaActivationCommit -Value $commit) -or -not (Test-HubQaActivationAncestor -Ancestor $commit -Descendant $FounderReleaseCommit)) { throw $code }
    }
    foreach ($field in @('offline_ingress_owner_marker_sha256','dependency_bundle_manifest_sha256')) { if (-not (Test-HubQaActivationDigest -Value (Get-HubQaActivationNodeString -Node $value[$field] -Code $code))) { throw $code } }
    return [pscustomobject][ordered]@{ Value=$value; RawBytes=$bytes; RawSha256=(Get-HubQaEnvelopeSha256 -Bytes $bytes); Commit=$FounderReleaseCommit; Run=$run }
}

function Assert-HubQaActivationPackageCommits {
    param([Parameter(Mandatory = $true)][object]$Release)
    $code = 'E_B0_ACTIVATION_RELEASE_BINDING'
    $a0 = Get-HubQaActivationNodeString -Node $Release.Value['a0_commit'] -Code $code
    $a02 = Get-HubQaActivationNodeString -Node $Release.Value['a02_commit'] -Code $code
    $a1 = Get-HubQaActivationNodeString -Node $Release.Value['a1_commit'] -Code $code
    $b = Get-HubQaActivationNodeString -Node $Release.Value['b_commit'] -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $a0 -Expected @('testframework/docker/dependency-bundle.schema.v1.json','testframework/docker/admission/verify_local_dependency_bundle.py','testframework/docker/admission/admit_local_dependency_assets.py','testframework/docker/admit-local-dependency-assets.ps1','testframework/docker/Dockerfile.backend-unit','testframework/docker/Dockerfile.frontend-unit','testframework/docker/Dockerfile.backend-openapi','testframework/docker/entrypoints/frontend-unit.sh','testframework/docker/runner-fixtures/g0-planning-selftest.py','testframework/docker/bootstrap-control.v1.json','testframework/docker/source-image-attestor.ps1','testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1','testframework/selftests/runner/g0-image-contract.ps1') -Code $code
    Assert-HubQaActivationA02Authority -A0Commit $a0 -A02Commit $a02 -CurrentCommit $a02 -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $a1 -Expected @('testframework/docker/dependency-bundle.manifest.v1.json','testframework/docker/source-policy.v2.json','testframework/docker/build-manifest.v3.json') -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $b -Expected @('testframework/runner/contracts/hub.qa-bootstrap.v1.json','testframework/runner/invoke-qa-bootstrap-v1.ps1','testframework/runner/invoke-qa-gate-v2.ps1','testframework/runner/run-task-v2.ps1','testframework/runner/new-repository-snapshot.ps1','testframework/runner/new-run-envelope.ps1','testframework/runner/write-g0-activation.ps1','testframework/schemas/repository-snapshot.schema.v1.json','testframework/schemas/run-envelope.schema.v1.json','testframework/selftests/runner/g0-runner-v2-contract.ps1','testframework/runner/README.md','.cursor/skills/qa-gate/SKILL.md') -Code $code
}

function Assert-HubQaActivationDChronology {
    param([Parameter(Mandatory = $true)][string]$AcceptanceCommit, [Parameter(Mandatory = $true)][object]$Release, [Parameter(Mandatory = $true)][string]$RepositoryCommit)
    $code = 'E_B0_D_SELF_REFERENCE'
    $planCommit = Get-HubQaActivationNodeString -Node $Release.Value['g0_plan_commit'] -Code $code
    $a0Commit = Get-HubQaActivationNodeString -Node $Release.Value['a0_commit'] -Code $code
    $a02Commit = Get-HubQaActivationNodeString -Node $Release.Value['a02_commit'] -Code $code
    $a1Commit = Get-HubQaActivationNodeString -Node $Release.Value['a1_commit'] -Code $code
    $bCommit = Get-HubQaActivationNodeString -Node $Release.Value['b_commit'] -Code $code
    $releaseCommit = [string]$Release.Commit
    $commits = @($planCommit,$a0Commit,$a02Commit,$a1Commit,$bCommit,$releaseCommit,$RepositoryCommit,$AcceptanceCommit)
    foreach ($commit in $commits) { if (-not (Test-HubQaActivationCommit -Value $commit)) { throw $code } }
    for ($left = 0; $left -lt $commits.Count; $left++) {
        for ($right = $left + 1; $right -lt $commits.Count; $right++) {
            if ($commits[$left] -ceq $commits[$right]) { throw $code }
        }
    }
    foreach ($prior in @($planCommit,$a0Commit,$a02Commit,$a1Commit,$bCommit)) {
        if (-not (Test-HubQaActivationAncestor -Ancestor $prior -Descendant $releaseCommit)) { throw $code }
    }
    if (-not (Test-HubQaActivationAncestor -Ancestor $releaseCommit -Descendant $RepositoryCommit) -or -not (Test-HubQaActivationAncestor -Ancestor $RepositoryCommit -Descendant $AcceptanceCommit)) { throw $code }
    if (Test-HubQaActivationGitBlobExists -Commit $AcceptanceCommit -Path $script:HubQaActivationOutputPath) { throw $code }
}

function Get-HubQaActivationDRecord {
    param([Parameter(Mandatory = $true)][string]$AcceptanceCommit, [Parameter(Mandatory = $true)][object]$Release, [Parameter(Mandatory = $true)][string]$RepositoryCommit)
    $code = 'E_B0_D_CANONICAL_JSON'
    Assert-HubQaActivationCommitExists -Commit $AcceptanceCommit -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $AcceptanceCommit -Expected @($script:HubQaActivationAcceptancePath) -Code $code
    Assert-HubQaActivationExactCommitPaths -Commit $RepositoryCommit -Expected @('testframework/docker/images.lock.v3.json') -Code 'E_B0_D_C1_BINDING'
    Assert-HubQaActivationDChronology -AcceptanceCommit $AcceptanceCommit -Release $Release -RepositoryCommit $RepositoryCommit
    [byte[]]$bytes = Get-HubQaActivationGitBlob -Commit $AcceptanceCommit -Path $script:HubQaActivationAcceptancePath
    $root = Get-HubQaActivationCanonicalNode -Bytes $bytes -Code $code
    $value = Assert-HubQaActivationNodeKeys -Node $root -Expected @('schema','acceptance_id','g0_plan_path','g0_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','founder_release','repository_commit','task','run','c_receipts','c1_lock','repository_snapshot','run_envelope','acceptance_verdict','accepted_at_utc') -Code 'E_B0_D_SCHEMA_CLOSED'
    if ((Get-HubQaActivationNodeString -Node $value['schema'] -Code $code) -cne 'hub.qa.g0-bootstrap-evidence-acceptance/v1' -or (Get-HubQaActivationNodeString -Node $value['acceptance_id'] -Code $code) -cne ('g0-bootstrap-evidence-acceptance-' + $Release.Run) -or (Get-HubQaActivationNodeString -Node $value['g0_plan_path'] -Code $code) -cne $script:HubQaActivationPlanPath -or (Get-HubQaActivationNodeString -Node $value['g0_plan_commit'] -Code $code) -cne (Get-HubQaActivationNodeString -Node $Release.Value['g0_plan_commit'] -Code $code) -or (Get-HubQaActivationNodeString -Node $value['repository_commit'] -Code $code) -cne $RepositoryCommit -or (Get-HubQaActivationNodeString -Node $value['task'] -Code $code) -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $value['run'] -Code $code) -cne $Release.Run) { throw 'E_B0_D_R_BINDING' }
    [void](Assert-HubQaEnvelopeUtc -Node $value['accepted_at_utc'] -Code 'E_B0_D_SCHEMA_CLOSED')
    foreach ($field in @('a0_commit','a02_commit','a1_commit','b_commit')) { if ((Get-HubQaActivationNodeString -Node $value[$field] -Code $code) -cne (Get-HubQaActivationNodeString -Node $Release.Value[$field] -Code $code)) { throw 'E_B0_D_R_BINDING' } }
    $releaseNode = Assert-HubQaActivationNodeKeys -Node $value['founder_release'] -Expected @('path','raw_sha256','commit') -Code 'E_B0_D_R_BINDING'
    if ((Get-HubQaActivationNodeString -Node $releaseNode['path'] -Code $code) -cne $script:HubQaActivationReleasePath -or (Get-HubQaActivationNodeString -Node $releaseNode['raw_sha256'] -Code $code) -cne $Release.RawSha256 -or (Get-HubQaActivationNodeString -Node $releaseNode['commit'] -Code $code) -cne $Release.Commit) { throw 'E_B0_D_R_BINDING' }
    $releaseCommit = Get-HubQaActivationNodeString -Node $releaseNode['commit'] -Code $code
    if (-not (Test-HubQaActivationCommit -Value $releaseCommit)) { throw 'E_B0_D_R_BINDING' }
    $receiptKeys = @('nonce_consumption','owner_marker','local_base_inspect','helper_base_verification','helper_image_attestation','ingress_copy','helper_admission','helper_cleanup','source_tree_verification','local_image_inspection','candidate_lock','planning_selftest_negative_red','planning_selftest_positive_green','command_receipts','run','cleanup')
    $receiptPaths = @('bootstrap/owner-marker.v1.json','bootstrap/local-base-inspect.v1.json','bootstrap/bootstrap-helper-python.verification.v1.json','bootstrap/g0-admission-helper.local-image-attestation.v1.json','bootstrap/ingress-copy-receipt.v1.json','bootstrap/g0-admission-helper.admission.v1.json','bootstrap/g0-admission-helper.cleanup.v1.json','bootstrap/source-tree-verification.v1.json','bootstrap/local-image-inspect.v1.json','bootstrap/images.lock.v3.json','selftests/g0-planning-negative-red.v1.json','selftests/g0-planning-positive-green.v1.json','commands/command-receipts.v1.json','run.json','cleanup.json')
    $reportRoot = 'testframework/reports/g0-assets-runner-' + $Release.Run
    $receiptsValue = Assert-HubQaActivationNodeKeys -Node $value['c_receipts'] -Expected $receiptKeys -Code 'E_B0_D_RECEIPT_LINK'
    $receipts = [ordered]@{}
    $receipts['nonce_consumption'] = Get-HubQaActivationNonceConsumptionReceipt -Node $receiptsValue['nonce_consumption'] -Release $Release -Code 'E_B0_D_RECEIPT_LINK'
    for ($index = 1; $index -lt $receiptKeys.Count; $index++) { $receipts[$receiptKeys[$index]] = Get-HubQaActivationReceiptBytes -Node $receiptsValue[$receiptKeys[$index]] -ExpectedPath $receiptPaths[$index - 1] -ReportRoot $reportRoot -Code 'E_B0_D_RECEIPT_LINK' }
    $lockNode = Assert-HubQaActivationNodeKeys -Node $value['c1_lock'] -Expected @('path','raw_sha256','commit') -Code 'E_B0_D_C1_BINDING'
    $lockPath = Get-HubQaActivationNodeString -Node $lockNode['path'] -Code $code; $lockDigest = Get-HubQaActivationNodeString -Node $lockNode['raw_sha256'] -Code $code; $lockCommit = Get-HubQaActivationNodeString -Node $lockNode['commit'] -Code $code
    if ($lockPath -cne 'testframework/docker/images.lock.v3.json' -or $lockCommit -cne $RepositoryCommit -or -not (Test-HubQaActivationDigest -Value $lockDigest)) { throw 'E_B0_D_C1_BINDING' }
    [byte[]]$lockBytes = Get-HubQaActivationGitBlob -Commit $RepositoryCommit -Path $lockPath
    if ((Get-HubQaEnvelopeSha256 -Bytes $lockBytes) -cne $lockDigest -or -not ([System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($lockBytes, $receipts['candidate_lock'].RawBytes))) { throw 'E_B0_D_C1_BINDING' }
    foreach ($pair in @(@('repository_snapshot','snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json'),@('run_envelope',('envelopes/g0-assets-runner-envelope-' + $Release.Run + '.run-envelope.v1.json')))) {
        $record = Get-HubQaActivationReceiptBytes -Node $value[$pair[0]] -ExpectedPath $pair[1] -ReportRoot $reportRoot -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
        if ($pair[0] -eq 'repository_snapshot') { $snapshot = $record } else { $envelope = $record }
    }
    $ownerRecord = Get-HubQaActivationCanonicalNode -Bytes $receipts['owner_marker'].RawBytes -Code 'E_B0_D_RECEIPT_LINK'
    $ownerValue = Assert-HubQaActivationNodeKeys -Node $ownerRecord -Expected @('schema','task','run','owner_nonce') -Code 'E_B0_D_RECEIPT_LINK'
    $ownerNonce = Get-HubQaActivationNodeString -Node $ownerValue['owner_nonce'] -Code 'E_B0_D_RECEIPT_LINK'
    if ((Get-HubQaActivationNodeString -Node $ownerValue['schema'] -Code 'E_B0_D_RECEIPT_LINK') -cne 'hub.qa.bootstrap-owner/v1' -or (Get-HubQaActivationNodeString -Node $ownerValue['task'] -Code 'E_B0_D_RECEIPT_LINK') -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $ownerValue['run'] -Code 'E_B0_D_RECEIPT_LINK') -cne $Release.Run -or $ownerNonce -cnotmatch '^[0-9a-f]{32}$') { throw 'E_B0_D_RECEIPT_LINK' }
    $runRecord = Get-HubQaActivationCanonicalNode -Bytes $receipts['run'].RawBytes -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $runValue = Assert-HubQaActivationNodeKeys -Node $runRecord -Expected @('schema','producer','task','run','owner_marker_sha256','release_one_shot_nonce','repository_snapshot','run_envelope','status','created_at_utc') -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $releaseNonce = Get-HubQaActivationNodeString -Node $Release.Value['one_shot_nonce'] -Code $code
    $nonceConsumptionRecord = Get-HubQaActivationCanonicalNode -Bytes $receipts['nonce_consumption'].RawBytes -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $nonceConsumptionValue = Assert-HubQaActivationNodeKeys -Node $nonceConsumptionRecord -Expected @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status') -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $consumedNonce = Get-HubQaActivationNodeString -Node $nonceConsumptionValue['one_shot_nonce'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $runReleaseNonce = Get-HubQaActivationNodeString -Node $runValue['release_one_shot_nonce'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    if ((Get-HubQaActivationNodeString -Node $runValue['schema'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne 'hub.qa.bootstrap-run/v1' -or (Get-HubQaActivationNodeString -Node $runValue['producer'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne 'hub.qa-bootstrap/v1' -or (Get-HubQaActivationNodeString -Node $runValue['task'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $runValue['run'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $Release.Run -or (Get-HubQaActivationNodeString -Node $runValue['owner_marker_sha256'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $receipts['owner_marker'].RawSha256 -or $runReleaseNonce -cne $releaseNonce -or $runReleaseNonce -cne $consumedNonce -or (Get-HubQaActivationNodeString -Node $runValue['status'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne 'completed-zero') { throw 'E_B0_D_SNAPSHOT_ENVELOPE_LINK' }
    [void](Assert-HubQaEnvelopeUtc -Node $runValue['created_at_utc'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK')
    foreach ($pair in @(@('repository_snapshot',$snapshot),@('run_envelope',$envelope))) {
        $link = Assert-HubQaActivationNodeKeys -Node $runValue[$pair[0]] -Expected @('path','raw_sha256') -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
        if ((Get-HubQaActivationNodeString -Node $link['path'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $pair[1].Path -or (Get-HubQaActivationNodeString -Node $link['raw_sha256'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $pair[1].RawSha256) { throw 'E_B0_D_SNAPSHOT_ENVELOPE_LINK' }
    }
    $snapshotRecord = Get-HubQaActivationCanonicalNode -Bytes $snapshot.RawBytes -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $envelopeRecord = Get-HubQaActivationCanonicalNode -Bytes $envelope.RawBytes -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $snapshotValue = Assert-HubQaSnapshotEnvelopePayload -Node $snapshotRecord -Kind snapshot -ExpectedTask 'g0-assets-runner' -ExpectedRun $Release.Run -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    $envelopeValue = Assert-HubQaSnapshotEnvelopePayload -Node $envelopeRecord -Kind envelope -ExpectedTask 'g0-assets-runner' -ExpectedRun $Release.Run -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    if ((Get-HubQaActivationNodeString -Node $envelopeValue['owner_nonce'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $ownerNonce -or (Get-HubQaActivationNodeString -Node $envelopeValue['repository_snapshot_sha256'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $snapshot.RawSha256) { throw 'E_B0_D_SNAPSHOT_ENVELOPE_LINK' }
    $snapshotEvidence = Assert-HubQaActivationNodeObject -Node $snapshotValue['evidence'] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK'
    foreach ($pair in @(@('positive_green_path','positive_green_sha256','selftests/g0-planning-positive-green.v1.json',$receipts['planning_selftest_positive_green'].RawSha256),@('candidate_lock_path','candidate_lock_sha256','bootstrap/images.lock.v3.json',$receipts['candidate_lock'].RawSha256),@('cleanup_path','cleanup_sha256','cleanup.json',$receipts['cleanup'].RawSha256))) {
        if ((Get-HubQaActivationNodeString -Node $snapshotEvidence[$pair[0]] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne ($reportRoot + '/' + $pair[2]) -or (Get-HubQaActivationNodeString -Node $snapshotEvidence[$pair[1]] -Code 'E_B0_D_SNAPSHOT_ENVELOPE_LINK') -cne $pair[3]) { throw 'E_B0_D_SNAPSHOT_ENVELOPE_LINK' }
    }
    if ((Get-HubQaActivationNodeString -Node $value['acceptance_verdict'] -Code $code) -cne 'ACCEPT') { throw 'E_B0_D_VERDICT' }
    return [pscustomobject][ordered]@{ Value=$value; Receipts=$receipts; LockBytes=$lockBytes; LockSha256=$lockDigest; Snapshot=$snapshot; Envelope=$envelope; ReportRoot=$reportRoot }
}

function Get-HubQaActivationTargets {
    param([Parameter(Mandatory = $true)][byte[]]$LockBytes, [Parameter(Mandatory = $true)][string]$RepositoryCommit)
    $code = 'E_B0_ACTIVATION_C1_BINDING'
    $lock = Get-HubQaActivationCanonicalNode -Bytes $LockBytes -Code $code -AllowNonCanonical
    $lockValue = Assert-HubQaActivationNodeKeys -Node $lock -Expected @('schema','platform','targets') -Code $code
    if ((Get-HubQaActivationNodeString -Node $lockValue['schema'] -Code $code) -cne 'hub.qa.local-images-lock/v3') { throw $code }
    $platform = Assert-HubQaActivationNodeKeys -Node $lockValue['platform'] -Expected @('os','architecture','variant') -Code $code
    if ((Get-HubQaActivationNodeString -Node $platform['os'] -Code $code) -cne 'linux' -or (Get-HubQaActivationNodeString -Node $platform['architecture'] -Code $code) -cne 'amd64') { throw $code }; Get-HubQaActivationNodeNull -Node $platform['variant'] -Code $code
    $targets = Assert-HubQaActivationNodeKeys -Node $lockValue['targets'] -Expected $script:HubQaActivationTargets -Code $code
    [byte[]]$controlBytes = Get-HubQaActivationGitBlob -Commit $RepositoryCommit -Path 'testframework/docker/bootstrap-control.v1.json'
    $control = Get-HubQaActivationCanonicalNode -Bytes $controlBytes -Code $code -AllowNonCanonical
    $controlValue = Assert-HubQaActivationNodeObject -Node $control -Code $code
    if (-not $controlValue.Contains('targets')) { throw $code }
    $controlTargets = Assert-HubQaActivationNodeObject -Node $controlValue['targets'] -Code $code
    $result = [ordered]@{}
    foreach ($targetName in $script:HubQaActivationTargets) {
        if (-not $controlTargets.Contains($targetName)) { throw $code }
        $entry = Assert-HubQaActivationNodeKeys -Node $targets[$targetName] -Expected @('local_image_id','local_image_id_sha256','entrypoint_path','entrypoint_sha256','command_sha256','compose_service','runtime_status') -Code $code
        $controlTarget = Assert-HubQaActivationNodeObject -Node $controlTargets[$targetName] -Code $code
        foreach ($field in @('entrypoint_path','compose_service','docker_cmd')) { if (-not $controlTarget.Contains($field)) { throw $code } }
        $localId = Get-HubQaActivationNodeString -Node $entry['local_image_id'] -Code $code; $localIdDigest = Get-HubQaActivationNodeString -Node $entry['local_image_id_sha256'] -Code $code; $entrypoint = Get-HubQaActivationNodeString -Node $entry['entrypoint_path'] -Code $code; $entrypointDigest = Get-HubQaActivationNodeString -Node $entry['entrypoint_sha256'] -Code $code; $commandDigest = Get-HubQaActivationNodeString -Node $entry['command_sha256'] -Code $code; $composeService = Get-HubQaActivationNodeString -Node $entry['compose_service'] -Code $code; $runtime = Get-HubQaActivationNodeString -Node $entry['runtime_status'] -Code $code
        if (-not (Test-HubQaActivationDigest -Value $localId) -or -not (Test-HubQaActivationDigest -Value $localIdDigest) -or -not (Test-HubQaActivationDigest -Value $entrypointDigest) -or -not (Test-HubQaActivationDigest -Value $commandDigest) -or $runtime -cne 'READY_FOR_G0_BOOTSTRAP_ONLY' -or $entrypoint -cne (Get-HubQaActivationNodeString -Node $controlTarget['entrypoint_path'] -Code $code) -or $composeService -cne (Get-HubQaActivationNodeString -Node $controlTarget['compose_service'] -Code $code)) { throw $code }
        if ((Get-HubQaEnvelopeSha256 -Bytes (Get-HubQaEnvelopeUtf8).GetBytes($localId)) -cne $localIdDigest) { throw $code }
        [byte[]]$entrypointBytes = Get-HubQaActivationGitBlob -Commit $RepositoryCommit -Path ('testframework/docker/entrypoints/' + $targetName + '.sh')
        if ((Get-HubQaEnvelopeSha256 -Bytes $entrypointBytes) -cne $entrypointDigest) { throw $code }
        $commandNode = $controlTarget['docker_cmd']
        if ([string]$commandNode.Type -cne 'array' -or (Get-HubQaEnvelopeSha256 -Bytes (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value @($commandNode.Value | ForEach-Object { Get-HubQaActivationNodeString -Node $_ -Code $code })))) -cne $commandDigest) { throw $code }
        $result[$targetName] = [ordered]@{ local_image_id=$localId; local_image_id_sha256=$localIdDigest; entrypoint_path=$entrypoint; entrypoint_sha256=$entrypointDigest; command_sha256=$commandDigest; compose_service=$composeService; runtime_status=$runtime }
    }
    return $result
}

function Write-HubQaG0Activation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$RequiredPlanCommit,
        [Parameter(Mandatory = $true)][string]$FounderReleaseCommit,
        [Parameter(Mandatory = $true)][string]$RepositoryCommit,
        [Parameter(Mandatory = $true)][string]$AcceptanceCommit
    )

    if (-not (Test-HubQaActivationCommit -Value $RequiredPlanCommit) -or -not (Test-HubQaActivationCommit -Value $FounderReleaseCommit) -or -not (Test-HubQaActivationCommit -Value $RepositoryCommit) -or -not (Test-HubQaActivationCommit -Value $AcceptanceCommit)) { throw 'E_B0_ACTIVATION_INPUT' }
    $head = Get-HubQaActivationGitText -Arguments @('rev-parse','HEAD')
    if ($head -cne $AcceptanceCommit -or $RequiredPlanCommit -ceq $FounderReleaseCommit -or $FounderReleaseCommit -ceq $RepositoryCommit -or $RepositoryCommit -ceq $AcceptanceCommit -or $RequiredPlanCommit -ceq $AcceptanceCommit -or -not (Test-HubQaActivationAncestor -Ancestor $RequiredPlanCommit -Descendant $FounderReleaseCommit) -or -not (Test-HubQaActivationAncestor -Ancestor $FounderReleaseCommit -Descendant $RepositoryCommit) -or -not (Test-HubQaActivationAncestor -Ancestor $RepositoryCommit -Descendant $AcceptanceCommit)) { throw 'E_B0_ACTIVATION_INPUT' }
    Assert-HubQaActivationCommitExists -Commit $RequiredPlanCommit
    Assert-HubQaActivationExactCommitPaths -Commit $RequiredPlanCommit -Expected @($script:HubQaActivationPlanPath) -Code 'E_B0_ACTIVATION_INPUT'
    Assert-HubQaActivationExactCommitPaths -Commit $RepositoryCommit -Expected @('testframework/docker/images.lock.v3.json') -Code 'E_B0_ACTIVATION_C1_BINDING'
    $release = Get-HubQaActivationRelease -FounderReleaseCommit $FounderReleaseCommit -RequiredPlanCommit $RequiredPlanCommit
    Assert-HubQaActivationPackageCommits -Release $release
    try { $d = Get-HubQaActivationDRecord -AcceptanceCommit $AcceptanceCommit -Release $release -RepositoryCommit $RepositoryCommit } catch { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
    if ((Get-HubQaActivationNodeString -Node $d.Value['founder_release'].Value['commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne $FounderReleaseCommit) { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
    try { Assert-HubQaActivationA02Authority -A0Commit (Get-HubQaActivationNodeString -Node $release.Value['a0_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -A02Commit (Get-HubQaActivationNodeString -Node $release.Value['a02_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -CurrentCommit $RepositoryCommit -Code 'E_B0_ACTIVATION_D_ACCEPTANCE' } catch { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
    try { $targets = Get-HubQaActivationTargets -LockBytes $d.LockBytes -RepositoryCommit $RepositoryCommit } catch { throw 'E_B0_ACTIVATION_C1_BINDING' }
    try {
        $snapshot = Get-HubQaActivationCanonicalNode -Bytes $d.Snapshot.RawBytes -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        $envelope = Get-HubQaActivationCanonicalNode -Bytes $d.Envelope.RawBytes -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        [void](Assert-HubQaSnapshotEnvelopePayload -Node $snapshot -Kind snapshot -ExpectedTask 'g0-assets-runner' -ExpectedRun $release.Run -Code 'E_B0_ACTIVATION_D_ACCEPTANCE')
        [void](Assert-HubQaSnapshotEnvelopePayload -Node $envelope -Kind envelope -ExpectedTask 'g0-assets-runner' -ExpectedRun $release.Run -Code 'E_B0_ACTIVATION_D_ACCEPTANCE')
        $snapshotValue = Assert-HubQaActivationNodeKeys -Node $snapshot -Expected @('schema','producer','snapshot_id','captured_for','task','run','envelope_id','role_claim','self_attested','source_revision','head','index','worktree','commits','ancestry','control','evidence','lifecycle') -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        $envelopeValue = Assert-HubQaActivationNodeKeys -Node $envelope -Expected @('schema','producer','envelope_id','contract','target','task','run','owner_nonce','repository_snapshot_sha256','report_root','limits','cleanup_contract','created_at_utc') -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        if ((Get-HubQaActivationNodeString -Node $snapshotValue['schema'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa.repository-snapshot/v1' -or (Get-HubQaActivationNodeString -Node $snapshotValue['producer'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa.repository-snapshot-producer/v1' -or (Get-HubQaActivationNodeString -Node $snapshotValue['snapshot_id'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne ('g0-assets-runner-snapshot-' + $release.Run + '-000-G0_BOOTSTRAP') -or (Get-HubQaActivationNodeString -Node $snapshotValue['task'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $snapshotValue['run'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne $release.Run -or (Get-HubQaActivationNodeString -Node $snapshotValue['role_claim'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa-bootstrap/v1' -or -not (Get-HubQaActivationNodeBoolean -Node $snapshotValue['self_attested'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -or (Get-HubQaActivationNodeString -Node $envelopeValue['schema'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa.run-envelope/v1' -or (Get-HubQaActivationNodeString -Node $envelopeValue['producer'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa.run-envelope-producer/v1' -or (Get-HubQaActivationNodeString -Node $envelopeValue['envelope_id'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne ('g0-assets-runner-envelope-' + $release.Run) -or (Get-HubQaActivationNodeString -Node $envelopeValue['contract'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa-runner/v2' -or (Get-HubQaActivationNodeString -Node $envelopeValue['target'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'g0-planning-selftest' -or (Get-HubQaActivationNodeString -Node $envelopeValue['task'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'g0-assets-runner' -or (Get-HubQaActivationNodeString -Node $envelopeValue['run'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne $release.Run -or (Get-HubQaActivationNodeString -Node $envelopeValue['owner_nonce'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cnotmatch '^[0-9a-f]{32}$' -or (Get-HubQaActivationNodeString -Node $envelopeValue['report_root'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne ('testframework/reports/g0-assets-runner-' + $release.Run) -or (Get-HubQaActivationNodeString -Node $envelopeValue['cleanup_contract'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'hub.qa.exact-cleanup/v1' -or (Get-HubQaActivationNodeString -Node $envelopeValue['repository_snapshot_sha256'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne $d.Snapshot.RawSha256 -or (Get-HubQaActivationNodeString -Node $snapshotValue['envelope_id'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne (Get-HubQaActivationNodeString -Node $envelopeValue['envelope_id'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE')) { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
        $capturedFor = Assert-HubQaActivationNodeKeys -Node $snapshotValue['captured_for'] -Expected @('kind','ordinal','subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256') -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        if ((Get-HubQaActivationNodeString -Node $capturedFor['kind'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne 'G0_BOOTSTRAP' -or [string]$capturedFor['ordinal'].Type -cne 'integer' -or [int64]$capturedFor['ordinal'].Value -ne 0) { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
        foreach ($field in @('subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256')) { Get-HubQaActivationNodeNull -Node $capturedFor[$field] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE' }
        $control = Assert-HubQaActivationNodeKeys -Node $snapshotValue['control'] -Expected @('g0_plan_sha256','bootstrap_contract_sha256','runner_contract_sha256','dependency_bundle_manifest_sha256','source_policy_sha256','build_manifest_sha256','images_lock_sha256') -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        foreach ($field in @('g0_plan_sha256','bootstrap_contract_sha256','runner_contract_sha256','dependency_bundle_manifest_sha256','source_policy_sha256','build_manifest_sha256')) { if (-not (Test-HubQaActivationDigest -Value (Get-HubQaActivationNodeString -Node $control[$field] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'))) { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' } }
        Get-HubQaActivationNodeNull -Node $control['images_lock_sha256'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        $evidence = Assert-HubQaActivationNodeKeys -Node $snapshotValue['evidence'] -Expected @('positive_green_path','positive_green_sha256','candidate_lock_path','candidate_lock_sha256','cleanup_path','cleanup_sha256') -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'
        $evidencePairs = @(@('positive_green_path','positive_green_sha256','selftests/g0-planning-positive-green.v1.json',$d.Receipts['planning_selftest_positive_green'].RawSha256),@('candidate_lock_path','candidate_lock_sha256','bootstrap/images.lock.v3.json',$d.Receipts['candidate_lock'].RawSha256),@('cleanup_path','cleanup_sha256','cleanup.json',$d.Receipts['cleanup'].RawSha256))
        foreach ($pair in $evidencePairs) { if ((Get-HubQaActivationNodeString -Node $evidence[$pair[0]] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne ($d.ReportRoot + '/' + $pair[2]) -or (Get-HubQaActivationNodeString -Node $evidence[$pair[1]] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE') -cne $pair[3]) { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' } }
    } catch { throw 'E_B0_ACTIVATION_D_ACCEPTANCE' }
    $sourcePaths = [ordered]@{ bootstrap_contract='testframework/runner/contracts/hub.qa-bootstrap.v1.json'; runner_contract='testframework/runner/run-task-v2.ps1'; dependency_bundle_manifest='testframework/docker/dependency-bundle.manifest.v1.json'; source_policy='testframework/docker/source-policy.v2.json'; build_manifest='testframework/docker/build-manifest.v3.json'; images_lock='testframework/docker/images.lock.v3.json'; repository_snapshot_schema='testframework/schemas/repository-snapshot.schema.v1.json'; run_envelope_schema='testframework/schemas/run-envelope.schema.v1.json'; repository_snapshot_producer='testframework/runner/new-repository-snapshot.ps1'; run_envelope_producer='testframework/runner/new-run-envelope.ps1' }
    $sourceDigests = [ordered]@{}
    foreach ($key in @($sourcePaths.Keys)) { $sourceDigests[$key] = Get-HubQaEnvelopeSha256 -Bytes (Get-HubQaActivationGitBlob -Commit $RepositoryCommit -Path $sourcePaths[$key]) }
    $activationRoot = Get-HubQaActivationRoot
    $outputFull = Join-Path $activationRoot ($script:HubQaActivationOutputPath.Replace('/', '\'))
    $outputParent = Split-Path -Parent $outputFull
    [void](Resolve-HubQaEnvelopeSafeDirectory -Path $outputParent -Root $activationRoot -Code 'E_B0_ACTIVATION_INPUT')
    $payload = [ordered]@{
        schema='hub.qa.g0-activation/v1'; activation_id='g0-activation-' + $release.Run; g0_plan_path=$script:HubQaActivationPlanPath; required_plan_commit=$RequiredPlanCommit; a0_commit=(Get-HubQaActivationNodeString -Node $release.Value['a0_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'); a02_commit=(Get-HubQaActivationNodeString -Node $release.Value['a02_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'); a1_commit=(Get-HubQaActivationNodeString -Node $release.Value['a1_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'); b_commit=(Get-HubQaActivationNodeString -Node $release.Value['b_commit'] -Code 'E_B0_ACTIVATION_D_ACCEPTANCE'); founder_release_path=$script:HubQaActivationReleasePath; founder_release_sha256=$release.RawSha256; founder_release_commit=$FounderReleaseCommit; nonce_consumption_path=$d.Receipts.nonce_consumption.Path; nonce_consumption_sha256=$d.Receipts.nonce_consumption.RawSha256; repository_commit=$RepositoryCommit; bootstrap_contract='hub.qa-bootstrap/v1'; bootstrap_contract_path=$sourcePaths.bootstrap_contract; bootstrap_contract_sha256=$sourceDigests.bootstrap_contract; runner_contract='hub.qa-runner/v2'; runner_contract_path=$sourcePaths.runner_contract; runner_contract_sha256=$sourceDigests.runner_contract; dependency_bundle_manifest_path=$sourcePaths.dependency_bundle_manifest; dependency_bundle_manifest_sha256=$sourceDigests.dependency_bundle_manifest; source_policy_path=$sourcePaths.source_policy; source_policy_sha256=$sourceDigests.source_policy; build_manifest_path=$sourcePaths.build_manifest; build_manifest_sha256=$sourceDigests.build_manifest; images_lock_path=$sourcePaths.images_lock; images_lock_sha256=$sourceDigests.images_lock; target_entrypoints=$targets; planning_selftest=[ordered]@{ target='g0-planning-selftest'; task='g0-assets-runner'; run=$release.Run; run_receipt_sha256=$d.Receipts.run.RawSha256; cleanup_receipt_sha256=$d.Receipts.cleanup.RawSha256; snapshot_sha256=$d.Snapshot.RawSha256; envelope_sha256=$d.Envelope.RawSha256; verdict='ACCEPT' }; repository_snapshot_schema_path=$sourcePaths.repository_snapshot_schema; repository_snapshot_schema_sha256=$sourceDigests.repository_snapshot_schema; run_envelope_schema_path=$sourcePaths.run_envelope_schema; run_envelope_schema_sha256=$sourceDigests.run_envelope_schema; repository_snapshot_producer_path=$sourcePaths.repository_snapshot_producer; repository_snapshot_producer_sha256=$sourceDigests.repository_snapshot_producer; run_envelope_producer_path=$sourcePaths.run_envelope_producer; run_envelope_producer_sha256=$sourceDigests.run_envelope_producer; report_cleanup_contract='hub.qa.exact-cleanup/v1'; acceptance_verdict='ACCEPT'; acceptance_commit=$AcceptanceCommit; created_at_utc=[DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [Globalization.CultureInfo]::InvariantCulture)
    }
    [byte[]]$outputBytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $payload))
    $written = Write-HubQaEnvelopeSafeCreateNew -Root $activationRoot -RelativePath $script:HubQaActivationOutputPath -Bytes $outputBytes -Code 'E_B0_ACTIVATION_SELF_COMMIT'
    return [pscustomobject][ordered]@{ activation_path=$script:HubQaActivationOutputPath; activation_sha256=(Get-HubQaEnvelopeSha256 -Bytes $written.Bytes); required_plan_commit=$RequiredPlanCommit; founder_release_commit=$FounderReleaseCommit; repository_commit=$RepositoryCommit; acceptance_commit=$AcceptanceCommit }
}

if ($MyInvocation.InvocationName -ne '.') { throw 'E_B0_ACTIVATION_INPUT' }
