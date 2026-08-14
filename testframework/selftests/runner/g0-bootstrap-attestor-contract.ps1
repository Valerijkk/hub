Set-StrictMode -Version Latest

function Read-PowerShellSyntax {
    param([string]$Path, [string]$Identity)

    $tokens = $null
    $errors = $null
    $ast = [System.Management.Automation.Language.Parser]::ParseFile($Path, [ref]$tokens, [ref]$errors)
    if (@($errors).Count -ne 0) { throw "$Identity is not parseable PowerShell." }
    return [pscustomobject]@{ Ast = $ast; Tokens = @($tokens) }
}

function Get-TopLevelPowerShellAssignment {
    param([object]$Syntax, [string]$Left, [string]$Identity)

    $matches = @($Syntax.Ast.FindAll({
        param($node)
        if ($node -isnot [System.Management.Automation.Language.AssignmentStatementAst] -or $node.Left.Extent.Text -cne $Left) { return $false }
        $ancestor = $node.Parent
        while ($null -ne $ancestor) {
            if ($ancestor -is [System.Management.Automation.Language.FunctionDefinitionAst] -or
                $ancestor -is [System.Management.Automation.Language.IfStatementAst] -or
                $ancestor -is [System.Management.Automation.Language.LoopStatementAst] -or
                $ancestor -is [System.Management.Automation.Language.TryStatementAst] -or
                $ancestor -is [System.Management.Automation.Language.TrapStatementAst] -or
                $ancestor -is [System.Management.Automation.Language.ScriptBlockExpressionAst]) { return $false }
            $ancestor = $ancestor.Parent
        }
        return $true
    }, $true))
    if ($matches.Count -ne 1) { throw "$Identity must have exactly one active top-level assignment." }
    return $matches[0]
}

function Get-PowerShellFunctionAst {
    param([object]$Syntax, [string]$Name, [string]$Identity)

    $matches = @($Syntax.Ast.FindAll({
        param($node)
        $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq $Name
    }, $true))
    if ($matches.Count -ne 1) { throw "$Identity must have exactly one function $Name." }
    return $matches[0]
}

function Get-AstStringValues {
    param([System.Management.Automation.Language.Ast]$Ast)

    return @($Ast.FindAll({ param($node) $node -is [System.Management.Automation.Language.StringConstantExpressionAst] }, $true) |
        ForEach-Object { $_.Value })
}

function Get-NonCommentPowerShellTokenText {
    param([object]$Syntax, [System.Management.Automation.Language.Ast]$Ast)

    return @($Syntax.Tokens | Where-Object {
        $_.Extent.StartOffset -ge $Ast.Extent.StartOffset -and
        $_.Extent.EndOffset -le $Ast.Extent.EndOffset -and
        $_.Kind -notin @(
            [System.Management.Automation.Language.TokenKind]::Comment,
            [System.Management.Automation.Language.TokenKind]::NewLine,
            [System.Management.Automation.Language.TokenKind]::EndOfInput
        )
    } | ForEach-Object { $_.Text })
}

function Assert-ExactStringArray {
    param([object]$Actual, [string[]]$Expected, [string]$Identity)

    $values = @($Actual)
    if ($values.Count -ne $Expected.Count) { throw "$Identity has an unexpected count." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if (-not [StringComparer]::Ordinal.Equals([string]$values[$index], $Expected[$index])) { throw "$Identity differs at index $index." }
    }
}

function Assert-ExactObjectKeys {
    param([object]$Value, [string[]]$Expected, [string]$Identity)

    Assert-ExactStringArray -Actual @($Value.PSObject.Properties.Name) -Expected $Expected -Identity $Identity
}

function Get-TopLevelShellRunStepCommands {
    param([string]$Source, [string]$Identity)

    $commands = New-Object System.Collections.Generic.List[string]
    $functionDepth = 0
    $controlDepth = 0
    foreach ($rawLine in ($Source -split "`r?`n")) {
        $line = $rawLine.Trim()
        if ($line.Length -eq 0 -or $line.StartsWith('#', [StringComparison]::Ordinal)) { continue }
        if ($line -match '^[A-Za-z_][A-Za-z0-9_]*\(\)\s*\{$') { $functionDepth++; continue }
        if ($line -ceq '}') { $functionDepth--; if ($functionDepth -lt 0) { throw "$Identity has unbalanced function braces." }; continue }
        if ($functionDepth -ne 0) { continue }
        if ($line -match '^(fi|done|esac)\b') { $controlDepth--; if ($controlDepth -lt 0) { throw "$Identity has unbalanced shell control flow." }; continue }
        if ($line -match '^(if|for|while|until|case)\b') {
            if ($line -notmatch ';\s*(fi|done|esac)\s*$') { $controlDepth++ }
            continue
        }
        if ($controlDepth -ne 0 -or $line -notmatch '^run_step\s+') { continue }
        $parts = @([regex]::Matches($line, "'[^']*'|[^\s]+") | ForEach-Object { $_.Value.Trim("'") })
        if ($parts.Count -lt 3 -or $parts[0] -cne 'run_step') { throw "$Identity has a malformed top-level run_step invocation." }
        $commands.Add(($parts[2..($parts.Count - 1)] -join ' '))
    }
    if ($functionDepth -ne 0 -or $controlDepth -ne 0) { throw "$Identity has unbalanced function braces or shell control flow." }
    return @($commands)
}

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
    Attestor = Join-Path $root 'testframework\docker\source-image-attestor.ps1'
    BackendEntrypoint = Join-Path $root 'testframework\docker\entrypoints\backend-unit.sh'
    FrontendConfig = Join-Path $root 'testframework\docker\vitest.frontend-unit.config.ts'
    BootstrapControl = Join-Path $root 'testframework\docker\bootstrap-control.v1.json'
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
$attestor = [System.IO.File]::ReadAllText($paths.Attestor)
$backendEntrypoint = [System.IO.File]::ReadAllText($paths.BackendEntrypoint)
$frontendConfig = [System.IO.File]::ReadAllText($paths.FrontendConfig)
$bootstrapControlRaw = [System.IO.File]::ReadAllText($paths.BootstrapControl)

$admissionPrivateContextKeys = @(
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
if ($admissionPrivateContextKeys.Count -ne 28) {
    throw 'The A0.1 admission private-context fixture must name the exact 28-key closed set.'
}

$sourceBaseline = '61b38573682d7c01e6d9381babc5085537d7990f'
$launcherSyntax = Read-PowerShellSyntax -Path $paths.Launcher -Identity 'host admission launcher'
$launcherSourceAssignment = Get-TopLevelPowerShellAssignment -Syntax $launcherSyntax -Left '$script:SourceCommit' -Identity 'host admission source authority'
Assert-ExactStringArray -Actual (Get-AstStringValues -Ast $launcherSourceAssignment.Right) -Expected @($sourceBaseline) -Identity 'host admission source authority'
$attestorSyntax = Read-PowerShellSyntax -Path $paths.Attestor -Identity 'source image attestor'
$attestorSourceAssignment = Get-TopLevelPowerShellAssignment -Syntax $attestorSyntax -Left '$script:SourceBaseline' -Identity 'attestor source authority'
Assert-ExactStringArray -Actual (Get-AstStringValues -Ast $attestorSourceAssignment.Right) -Expected @($sourceBaseline) -Identity 'attestor source authority'
$fixedTargetsAssignment = Get-TopLevelPowerShellAssignment -Syntax $attestorSyntax -Left '$script:FixedTargets' -Identity 'attestor fixed targets'
$fixedTargetStrings = Get-AstStringValues -Ast $fixedTargetsAssignment.Right
$backendActions = @(
    'go build ./...',
    'go vet ./...',
    'go test -tags hubtest -json -count=1 -timeout=150s ./tests/...',
    'go test -tags hubtest -json -race -count=1 -timeout=150s ./tests/...'
)
foreach ($command in $backendActions) {
    if (@($fixedTargetStrings | Where-Object { $_ -ceq $command }).Count -ne 1) { throw "Attestor fixed-target AST omits or duplicates backend action: $command" }
}
$backendShellActions = Get-TopLevelShellRunStepCommands -Source $backendEntrypoint -Identity 'backend unit entrypoint'
Assert-ExactStringArray -Actual $backendShellActions -Expected $backendActions -Identity 'backend shell and attestor action equality'

$bootstrapControl = $bootstrapControlRaw | ConvertFrom-Json -ErrorAction Stop
Assert-ExactObjectKeys -Value $bootstrapControl -Expected @('schema', 'status', 'bootstrap_contract', 'platform', 'candidate_lock', 'runtime', 'targets') -Identity 'bootstrap control root'
Assert-ExactObjectKeys -Value $bootstrapControl.targets -Expected @('backend-unit', 'frontend-unit', 'backend-openapi', 'g0-planning-selftest') -Identity 'bootstrap control targets'
Assert-ExactObjectKeys -Value $bootstrapControl.targets.'backend-unit' -Expected @('dockerfile', 'entrypoint_path', 'compose_service', 'local_image_lock_field', 'docker_cmd', 'target_action', 'bootstrap_selectable') -Identity 'bootstrap backend target'
Assert-ExactObjectKeys -Value $bootstrapControl.targets.'frontend-unit' -Expected @('dockerfile', 'entrypoint_path', 'compose_service', 'local_image_lock_field', 'docker_cmd', 'target_action', 'bootstrap_selectable') -Identity 'bootstrap frontend target'
Assert-ExactStringArray -Actual $bootstrapControl.targets.'backend-unit'.target_action -Expected $backendActions -Identity 'backend JSON and executable action equality'

$expectedFrontendConfig = (@(
    'import { defineConfig } from "vitest/config";', '', 'export default defineConfig({',
    '  root: "/opt/hub/source/frontend",', '  resolve: {', '    alias: {',
    '      "@": "/opt/hub/source/frontend/src",', '    },', '  },', '  test: {',
    '    include: ["/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}"],',
    '    environment: "jsdom",', '    globals: true,',
    '    setupFiles: ["/opt/hub/source/frontend/src/test/setup.ts"],', '  },', '});', ''
) -join "`n")
if ($frontendConfig.Replace("`r`n", "`n") -cne $expectedFrontendConfig) { throw 'Frontend Vitest config differs from the exact executable TS contract.' }

$boundContractAst = Get-PowerShellFunctionAst -Syntax $attestorSyntax -Name 'Assert-HubQaBoundControlContract' -Identity 'bound control contract'
$boundContractTokens = Get-NonCommentPowerShellTokenText -Syntax $attestorSyntax -Ast $boundContractAst
foreach ($token in @('Read-HubQaRegularUtf8File', 'vitest_config_path', 'vitest_config_sha256')) {
    if (@($boundContractTokens | Where-Object { $_ -ceq $token }).Count -lt 1) { throw "Comment-stripped bound-control function omits active token $token." }
}
$frontendBranches = @($boundContractAst.FindAll({
    param($node)
    $node -is [System.Management.Automation.Language.IfStatementAst] -and
    $node.Extent.Text -match '(?m)\$targetName\s+-ceq\s+''frontend-unit'''
}, $true))
if ($frontendBranches.Count -ne 2) { throw 'Bound control contract must have the two active frontend-only key and digest branches.' }
$frontendBranchStrings = @($frontendBranches | ForEach-Object { Get-AstStringValues -Ast $_ })
foreach ($value in @('frontend-unit', 'vitest_config_path', 'vitest_config_sha256', 'testframework/docker/vitest.frontend-unit.config.ts')) {
    if (@($frontendBranchStrings | Where-Object { $_ -ceq $value }).Count -lt 1) { throw "Frontend-only attestor branch omits $value." }
}
$frontendCommands = @($frontendBranches | ForEach-Object {
    $_.FindAll({ param($node) $node -is [System.Management.Automation.Language.CommandAst] }, $true) |
        ForEach-Object { $_.GetCommandName() }
})
if (@($frontendCommands | Where-Object { $_ -ceq 'Read-HubQaRegularUtf8File' }).Count -ne 1) { throw 'Frontend-only attestor branch must actively read the Vitest config once.' }
$frontendComparisons = @($frontendBranches | ForEach-Object {
    $_.FindAll({ param($node) $node -is [System.Management.Automation.Language.BinaryExpressionAst] }, $true) |
        ForEach-Object { ($_.Extent.Text -replace '\s+', ' ').Trim() }
})
if (@($frontendComparisons | Where-Object { $_ -ceq '$vitestConfig.Sha256 -cne $target.vitest_config_sha256' }).Count -ne 1) { throw 'Frontend-only attestor branch must actively compare the raw Vitest config digest.' }

$schemaValue = $schema | ConvertFrom-Json -ErrorAction Stop
Assert-ExactObjectKeys -Value $schemaValue.properties.source_commit -Expected @('const') -Identity 'dependency schema source commit'
if ($schemaValue.properties.source_commit.const -cne $sourceBaseline) { throw 'Dependency schema source authority mismatch.' }
$verifierSourceLines = @($verifier -split "`r?`n" | Where-Object { $_ -match '^SOURCE_COMMIT:\s+Final\s+=\s+' })
if ($verifierSourceLines.Count -ne 1 -or $verifierSourceLines[0].Trim() -cne ('SOURCE_COMMIT: Final = "' + $sourceBaseline + '"')) { throw 'Python verifier source authority is not one active top-level constant.' }

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
    'Assert-HubQaPrivateHelperImageAttestation',
    'g0-admission-helper.local-image-attestation.v1.json',
    'GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles)',
    'Docker\Docker\resources\bin\docker.exe',
    'HUB_QA_PRIVATE_OWNER_MARKER_SHA256',
    'HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256',
    '$BootstrapPrivateContext.Keys.Count -ne $required.Count',
    'Read-HubQaBoundedControlBytes',
    '$script:MaxBoundedControlBytes = 1048576',
    '[ValidateSet(',
    "'ingress-owner-marker'",
    "'a1-dependency-manifest'",
    "'copied-dependency-manifest'",
    "'ingress-copy-receipt'",
    "'local-base-receipt'",
    "'helper-image-attestation'",
    "'helper-admission-receipt'",
    'Assert-HubQaIngressOwnerMarker',
    'Assert-HubQaClosedIngressMetadata',
    "'--pull', 'never'",
    '''--name'', $closed.HelperContainerName',
    '''--label'', $closed.HelperOwnerLabel',
    'hub.qa.owner-marker-sha256=',
    '$BootstrapPrivateContext.PrivateHelperLocalImageId',
    'dst=/work',
    'g0-admission-helper.admission.v1.json',
    'NOT_READY/NOT_RUN'
)) {
    if ($launcher.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Host launcher does not expose required closed bootstrap token: $required"
    }
}

foreach ($contextKey in $admissionPrivateContextKeys) {
    if ($launcher.IndexOf("'$contextKey'", [System.StringComparison]::Ordinal) -lt 0) {
        throw "Host launcher omits the closed A0.1 admission-context key: $contextKey"
    }
}

# host-launcher-forbids-asset-readallbytes: archive bytes remain helper-only.
# The host can consume only fixed small controls through a held physical handle;
# Package B owns helper cleanup and the later cleanup receipt.
foreach ($required in @(
    'Read-HubQaBoundedControlBytes',
    'Open-HubQaPhysicalRegularFileHandle',
    'Assert-HubQaCanonicalContainedControlPath',
    'Get-HubQaHandleFileObservation',
    'Read-HubQaHeldBoundedControlBytes',
    'CreateFileW',
    'FILE_FLAG_OPEN_REPARSE_POINT',
    'SafeFileHandle',
    'GetFileInformationByHandle',
    'GetFinalPathNameByHandle',
    'DuplicateHandle',
    'DUPLICATE_SAME_ACCESS',
    'New-HubQaOwnedDuplicateFileHandle',
    'Assert-HubQaBoundedRegularControlObservation',
    'ConvertFrom-HubQaStrictCanonicalJson',
    'ConvertFrom-HubQaCanonicalJsonValue',
    'Assert-HubQaCanonicalJsonString',
    'Assert-HubQaCanonicalJsonNumber',
    'Assert-HubQaPrivateDependencyBundleManifest',
    'Assert-HubQaPrivateLocalBaseReceiptChildren',
    'Assert-HubQaHelperAdmissionReceiptChildren',
    'duplicate JSON object key',
    'Package B owns helper cleanup and the later cleanup receipt'
)) {
    if ($launcher.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "host-launcher-forbids-asset-readallbytes: host control reader or strict parser omits $required"
    }
}
foreach ($forbiddenAssetRead in @(
    'Assert-HubQaClosedCopiedIngressTree',
    'Read-HubQaBoundedControlBytes -Path',
    'Read-HubQaBoundedControlBytes -Path $candidate',
    'Read-HubQaRegularFileBytes -Path $candidate',
    'GetFiles(',
    'GetDirectories(',
    'Get-Content',
    'Copy-Item',
    'Expand-Archive',
    '[System.IO.File]::ReadAllBytes(',
    'ConvertFrom-Json',
    'Get-FileHash'
)) {
    if ($launcher.IndexOf($forbiddenAssetRead, [System.StringComparison]::OrdinalIgnoreCase) -ge 0) {
        throw "host-launcher-forbids-asset-readallbytes: forbidden asset-content authority: $forbiddenAssetRead"
    }
}
if ($launcher.IndexOf('$script:MaxBoundedControlBytes = 1048576', [System.StringComparison]::Ordinal) -lt 0 -or
    $launcher.IndexOf('Read-HubQaHeldBoundedControlBytes', [System.StringComparison]::Ordinal) -lt 0) {
    throw 'host-launcher-forbids-asset-readallbytes: the held bounded control reader is not closed.'
}
$heldReaderStart = $launcher.IndexOf('function Read-HubQaHeldBoundedControlBytes', [System.StringComparison]::Ordinal)
if ($heldReaderStart -lt 0) {
    throw 'host-launcher-forbids-asset-readallbytes: held bounded reader is missing.'
}
$heldReaderEnd = $launcher.IndexOf('function Read-HubQaBoundedControlBytes', $heldReaderStart, [System.StringComparison]::Ordinal)
if ($heldReaderEnd -le $heldReaderStart) {
    throw 'host-launcher-forbids-asset-readallbytes: held bounded reader has no closed source boundary.'
}
$preReadObservation = $launcher.IndexOf('$preRead = Get-HubQaHandleFileObservation -Handle $Opened.Handle', $heldReaderStart, [System.StringComparison]::Ordinal)
$preReadRegularity = $launcher.IndexOf('Assert-HubQaBoundedRegularControlObservation -Observation $preRead', $heldReaderStart, [System.StringComparison]::Ordinal)
$allocation = $launcher.IndexOf('[byte[]]$bytes = New-Object byte[] ([int]$preRead.Length)', $heldReaderStart, [System.StringComparison]::Ordinal)
$duplicateOpen = $launcher.IndexOf('New-HubQaOwnedDuplicateFileHandle -Handle $Opened.Handle', $heldReaderStart, [System.StringComparison]::Ordinal)
$streamOpen = $launcher.IndexOf('$stream = [IO.FileStream]::new($duplicate', $heldReaderStart, [System.StringComparison]::Ordinal)
$postReadObservation = $launcher.IndexOf('$postRead = Get-HubQaHandleFileObservation -Handle $Opened.Handle', $heldReaderStart, [System.StringComparison]::Ordinal)
$postReadIdentity = $launcher.IndexOf('Assert-HubQaSameHandleFileObservation -Expected $preRead -Actual $postRead', $heldReaderStart, [System.StringComparison]::Ordinal)
$streamDispose = $launcher.IndexOf('$stream.Dispose()', $heldReaderStart, [System.StringComparison]::Ordinal)
if ($preReadObservation -lt $heldReaderStart -or $preReadObservation -ge $heldReaderEnd -or
    $preReadRegularity -lt $preReadObservation -or $preReadRegularity -ge $heldReaderEnd -or
    $allocation -lt $preReadRegularity -or $allocation -ge $heldReaderEnd -or
    $duplicateOpen -lt $allocation -or $duplicateOpen -ge $heldReaderEnd -or
    $streamOpen -lt $duplicateOpen -or $streamOpen -ge $heldReaderEnd -or
    $postReadObservation -lt $streamOpen -or $postReadObservation -ge $heldReaderEnd -or
    $postReadIdentity -lt $postReadObservation -or $postReadIdentity -ge $heldReaderEnd -or
    $streamDispose -lt $postReadIdentity -or $streamDispose -ge $heldReaderEnd) {
    throw 'host-launcher-forbids-asset-readallbytes: held reader must revalidate before allocation, read through an owned duplicate, and observe identity before disposal.'
}
$openControlStart = $launcher.IndexOf('function Open-HubQaPhysicalRegularFileHandle', [System.StringComparison]::Ordinal)
if ($openControlStart -lt 0) {
    throw 'host-launcher-forbids-asset-readallbytes: physical control-handle opener is missing.'
}
$openControlEnd = $launcher.IndexOf('function Read-HubQaHeldBoundedControlBytes', $openControlStart, [System.StringComparison]::Ordinal)
if ($openControlEnd -le $openControlStart) {
    throw 'host-launcher-forbids-asset-readallbytes: physical control-handle opener has no closed source boundary.'
}
$openObservation = $launcher.IndexOf('$observation = Get-HubQaHandleFileObservation -Handle $handle', $openControlStart, [System.StringComparison]::Ordinal)
$openRegularity = $launcher.IndexOf('Assert-HubQaBoundedRegularControlObservation -Observation $observation', $openControlStart, [System.StringComparison]::Ordinal)
$openFinalPath = $launcher.IndexOf('Get-HubQaFinalPathNameByHandle -Handle $handle', $openControlStart, [System.StringComparison]::Ordinal)
$openExactContainment = $launcher.IndexOf('final handle path is not the exact contained fixed control path.', $openControlStart, [System.StringComparison]::Ordinal)
if ($openObservation -lt $openControlStart -or $openObservation -ge $openControlEnd -or
    $openRegularity -lt $openObservation -or $openRegularity -ge $openControlEnd -or
    $openFinalPath -lt $openRegularity -or $openFinalPath -ge $openControlEnd -or
    $openExactContainment -lt $openFinalPath -or $openExactContainment -ge $openControlEnd) {
    throw 'host-launcher-forbids-asset-readallbytes: control opener must bind a regular nonreparse handle to its final contained path before return.'
}
if ($launcher.IndexOf('public const UInt32 FILE_SHARE_READ = 0x00000001;', [System.StringComparison]::Ordinal) -lt 0 -or
    $launcher.IndexOf('FILE_SHARE_WRITE', [System.StringComparison]::Ordinal) -ge 0 -or
    $launcher.IndexOf('FILE_SHARE_DELETE', [System.StringComparison]::Ordinal) -ge 0) {
    throw 'host-launcher-forbids-asset-readallbytes: control handles must deny write and delete sharing.'
}
$canonicalEscapeStart = $launcher.IndexOf('function Assert-HubQaCanonicalJsonString', [System.StringComparison]::Ordinal)
if ($canonicalEscapeStart -lt 0) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON string parser is missing.'
}
$canonicalEscapeEnd = $launcher.IndexOf('function Assert-HubQaCanonicalJsonNumber', $canonicalEscapeStart, [System.StringComparison]::Ordinal)
if ($canonicalEscapeEnd -le $canonicalEscapeStart) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON string parser has no closed source boundary.'
}
$uppercaseControlEscape = $launcher.IndexOf("`$UnicodeEscape -cnotmatch '^00[0-9A-F]{2}`$'", $canonicalEscapeStart, [System.StringComparison]::Ordinal)
$nonControlEscapeRejection = $launcher.IndexOf('$code -gt 0x1f', $canonicalEscapeStart, [System.StringComparison]::Ordinal)
$shortEscapeRejection = $launcher.IndexOf('0x08, 0x09, 0x0a, 0x0c, 0x0d', $canonicalEscapeStart, [System.StringComparison]::Ordinal)
$shortEscapeAllowlist = $launcher.IndexOf("`$Escape -cnotin @('`"', '\\', 'b', 'f', 'n', 'r', 't')", $canonicalEscapeStart, [System.StringComparison]::Ordinal)
$lowercaseControlEscape = $launcher.IndexOf("'^00[0-9a-f]{2}`$'", $canonicalEscapeStart, [System.StringComparison]::Ordinal)
if ($uppercaseControlEscape -lt $canonicalEscapeStart -or $uppercaseControlEscape -ge $canonicalEscapeEnd -or
    $nonControlEscapeRejection -lt $uppercaseControlEscape -or $nonControlEscapeRejection -ge $canonicalEscapeEnd -or
    $shortEscapeRejection -lt $nonControlEscapeRejection -or $shortEscapeRejection -ge $canonicalEscapeEnd -or
    $shortEscapeAllowlist -lt $canonicalEscapeStart -or $shortEscapeAllowlist -ge $canonicalEscapeEnd -or
    ($lowercaseControlEscape -ge $canonicalEscapeStart -and $lowercaseControlEscape -lt $canonicalEscapeEnd)) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON must allow literal Unicode and only uppercase u00 control escapes plus short escapes.'
}
$canonicalNumberStart = $launcher.IndexOf('function Assert-HubQaCanonicalJsonNumber', [System.StringComparison]::Ordinal)
if ($canonicalNumberStart -lt 0) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON integer parser is missing.'
}
$canonicalNumberEnd = $launcher.IndexOf('function ConvertFrom-HubQaCanonicalJsonValue', $canonicalNumberStart, [System.StringComparison]::Ordinal)
if ($canonicalNumberEnd -le $canonicalNumberStart) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON integer parser has no closed source boundary.'
}
$canonicalIntegerGrammar = $launcher.IndexOf("`$Digits -notmatch '^(0|[1-9][0-9]*)`$'", $canonicalNumberStart, [System.StringComparison]::Ordinal)
$canonicalNineteenDigitLimit = $launcher.IndexOf('$Digits.Length -gt 19', $canonicalNumberStart, [System.StringComparison]::Ordinal)
$canonicalLegacyEighteenDigitLimit = $launcher.IndexOf('$Digits.Length -gt 18', $canonicalNumberStart, [System.StringComparison]::Ordinal)
$canonicalTwentyDigitLimit = $launcher.IndexOf('$Digits.Length -gt 20', $canonicalNumberStart, [System.StringComparison]::Ordinal)
$canonicalIntegerValidation = $launcher.IndexOf('Assert-HubQaCanonicalJsonNumber -Digits $digits -Identity $Identity', $canonicalNumberEnd, [System.StringComparison]::Ordinal)
$int64OverflowParse = $launcher.IndexOf('[Int64]::Parse($digits, [Globalization.CultureInfo]::InvariantCulture)', $canonicalNumberEnd, [System.StringComparison]::Ordinal)
$int64OverflowCatch = $launcher.IndexOf('catch { throw "$Identity has an invalid JSON integer." }', $int64OverflowParse, [System.StringComparison]::Ordinal)
if ($canonicalIntegerGrammar -lt $canonicalNumberStart -or $canonicalIntegerGrammar -ge $canonicalNumberEnd -or
    $canonicalNineteenDigitLimit -lt $canonicalIntegerGrammar -or $canonicalNineteenDigitLimit -ge $canonicalNumberEnd -or
    ($canonicalLegacyEighteenDigitLimit -ge $canonicalNumberStart -and $canonicalLegacyEighteenDigitLimit -lt $canonicalNumberEnd) -or
    ($canonicalTwentyDigitLimit -ge $canonicalNumberStart -and $canonicalTwentyDigitLimit -lt $canonicalNumberEnd) -or
    $canonicalIntegerValidation -lt $canonicalNumberEnd -or
    $int64OverflowParse -lt $canonicalIntegerValidation -or
    $int64OverflowCatch -lt $int64OverflowParse) {
    throw 'host-launcher-forbids-asset-readallbytes: canonical JSON integers must allow exactly up to 19 unsigned decimal digits, reject 20 digits, and fail closed on Int64 overflow.'
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
    'Get-ChildItem',
    'Copy-Item',
    'Expand-Archive',
    'tar.exe',
    'tar ',
    'Add-Type -AssemblyName System.IO.Compression'
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
    '"reason_code": None',
    'archive_format',
    'member_count',
    'expanded_bytes',
    'layout_sha256',
    'O_NOFOLLOW',
    'os.fstat',
    'tarfile.open(fileobj=',
    'mode="r:"',
    'E_ASSET_ARCHIVE_INVALID',
    'go-race-toolchain.json',
    'hub.qa.go-race-toolchain-provenance/v1',
    'cc_target',
    'cgo_enabled'
)) {
    if ($verifier.IndexOf($required, [System.StringComparison]::Ordinal) -lt 0) {
        throw "Shared Python verifier omits required closed contract token: $required"
    }
}
foreach ($forbidden in @(
    'task_owner_marker_sha256: str | None',
    'handoff_owner_marker_sha256: str | None',
    'task_owner_marker_sha256 if task_owner_marker_sha256 is not None else actual_handoff_owner_marker_sha',
    'Path.read_bytes()',
    'extractall(',
    'tarfile.open(name=',
    'tarfile.open(path='
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
    'E_ASSET_ARCHIVE_INVALID',
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
    'E_MANIFEST_SCHEMA',
    'archive-lstat-regular-only',
    'archive-safe-handle-required',
    'archive-replaced-between-passes',
    'archive-replaced-before-staging',
    'archive-stream-cap-before-tar',
    'archive-terminal-eof',
    'archive-pax-closed',
    'archive-raw-tarfile-crosscheck',
    'E_ASSET_ARCHIVE_INVALID'
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
