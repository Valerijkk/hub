# G0-A is a source-only contract harness. No helper is invoked from this file.

function Get-HubQaG0ComplexityBudget {
    [CmdletBinding()]
    param()

    return [pscustomobject]@{
        ContractState                    = 'STATIC_CONTRACT_REVIEWED'
        PrivateContractState             = 'PRIVATE_CONTRACT_NOT_RUN'
        MutableStateBudget               = [pscustomobject]@{
            Families                     = @('task receipt')
            ReceiptRule                  = 'first non-zero exit is immutable'
            AdditionalFamiliesRequire    = 're-slicing before implementation'
            PreMutationSequence          = @(
                'request validation'
                'canonical-path validation'
                'duplicate-aware three-key lock-schema validation'
            )
            PreMutationFailure           = [pscustomobject]@{
                IntendedExit              = 64
                DependencyCalls           = 0
                ReportsTreeManifest       = 'unchanged'
            }
        }
        HelperBudget                     = [pscustomobject]@{
            FrozenHelpers                = @(
                'Get-HubQaG0ComplexityBudget'
                'Get-HubQaG0StaticCaseDefinitions'
                'Assert-HubQaOwnedFixturePath'
                'New-HubQaContractFixtureWorkspace'
                'Reset-HubQaContractFixtureCase'
                'Get-HubQaContractFixtureTreeManifest'
                'Remove-HubQaContractFixtureWorkspace'
            )
            Count                        = 7
            AdditionalHelpersRequire     = 're-slicing before implementation'
        }
        BranchBudget                     = [pscustomobject]@{
            Families                     = @(
                'request validation'
                'canonical-path validation'
                'lock-schema validation'
                'receipt finalization'
                'typed cleanup selection'
            )
            AdditionalFamiliesRequire    = 're-slicing before implementation'
        }
        SideEffectBudget                 = [pscustomobject]@{
            BeforeValidationFailure      = 'none'
            ReportsTreeBeforeAdapter     = 'unchanged'
            FutureWriteCopySites         = @(
                [pscustomobject]@{
                    Owner                 = 'New-HubQaContractFixtureWorkspace'
                    Operation             = 'copy'
                    CanonicalSource       = 'testframework/runner/run-task.ps1'
                    Destination           = 'fixture-owned temporary repository'
                }
            )
            Count                        = 1
            AdditionalWriteCopySitesRequire = 're-slicing before implementation'
        }
        ChildProcessBudget               = [pscustomobject]@{
            FutureProductionSites        = @('RunFixedTarget')
            Count                        = 1
            AdditionalSitesRequire       = 're-slicing before implementation'
        }
        CleanupBudget                    = [pscustomobject]@{
            FutureRoutes                 = @('try/finally -> typed owned/foreign selection -> cleanup owned set -> zero-owned proof')
            Count                        = 1
            AdditionalRoutesRequire      = 're-slicing before implementation'
        }
    }
}

function Get-HubQaG0StaticCaseDefinitions {
    [CmdletBinding()]
    param()

    return @(
        [pscustomobject]@{
            Id       = 'TC-G0-001'
            Intent   = 'valid lock, success receipt, empty inventory, successful cleanup'
            Expected = [pscustomobject]@{
                IntendedExit        = 0
                Inventory            = 'empty'
                Cleanup              = 'successful'
                Observation          = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-002'
            Intent   = 'first exit 23 remains immutable after later transitions'
            Expected = [pscustomobject]@{
                PrimaryExit          = 23
                FinalExit            = 23
                Observation          = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-003'
            Intent   = 'synthetic command-deadline receipt has exit 124 and no child mode'
            Expected = [pscustomobject]@{
                IntendedExit        = 124
                ChildMode           = 'none'
                Receipt             = 'synthetic command-deadline'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-003A'
            Intent   = 'completed prior receipt plus target deadline has exit 124 and no child mode'
            Expected = [pscustomobject]@{
                IntendedExit        = 124
                ChildMode           = 'none'
                Receipt             = 'completed prior receipt plus target deadline'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-004'
            Intent   = 'inert bounded-wait descriptor plus already-expired monotonic tick creates no process'
            Expected = [pscustomobject]@{
                WaitDescriptor      = 'bounded and inert'
                MonotonicTick       = 'already expired'
                ProcessCreation     = 0
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-005'
            Intent   = 'zero collection intends final exit 65'
            Expected = [pscustomobject]@{
                Collected           = 0
                IntendedFinalExit   = 65
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-006'
            Intent   = 'literal malformed result {not-json intends final exit 65'
            Expected = [pscustomobject]@{
                Result              = '{not-json'
                IntendedFinalExit   = 65
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-006A'
            Intent   = 'collected 2, passed 1, skipped 1, empty skip allowlist intends final exit 65'
            Expected = [pscustomobject]@{
                Collected           = 2
                Passed              = 1
                Skipped             = 1
                SkipAllowlist       = @()
                IntendedFinalExit   = 65
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-007'
            Intent   = 'invalid contract, target, Task, Run, and lock variants reject before adapter construction or child handling'
            Fixtures = [pscustomobject]@{
                InvalidTagOnlyLockFixture = 'LockBytes.TagOnlyBackendUnit'
                SchemaValidLock     = 'example.invalid/hubqa/g0@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
                InvalidVariants     = @('contract', 'target', 'Task', 'Run', 'lock')
            }
            Expected = [pscustomobject]@{
                IntendedExit        = 64
                DependencyCalls     = 0
                ReportsTreeManifest = 'unchanged'
                AdapterFactory      = 'not constructed'
                InspectLockedImage  = 'not reached'
                RunFixedTarget      = 'not reached'
                ChildHandling       = 'not reached'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-008'
            Intent   = 'existing directory, mismatched owner marker, reparse point, and resolved-outside path are rejected'
            Expected = [pscustomobject]@{
                Rejections          = @('existing directory', 'mismatched owner marker', 'reparse point', 'resolved-outside path')
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-009'
            Intent   = 'exactly two owned and two foreign inventory records; only owned typed records are eligible for inert cleanup'
            Expected = [pscustomobject]@{
                OwnedRecords        = 2
                ForeignRecords      = 2
                EligibleForCleanup  = 'owned typed records only'
                Cleanup             = 'inert declaration only'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-010'
            Intent   = 'primary 23 plus cleanup failure keeps 23; primary 0 plus cleanup failure intends 71'
            Expected = [pscustomobject]@{
                Primary23CleanupFailure = 23
                Primary0CleanupFailure  = 71
                Observation              = 'static source-review intent only'
            }
        }
    )
}

function Assert-HubQaOwnedFixturePath {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$CandidatePath,

        [Parameter(Mandatory)]
        [string]$OwnedRoot,

        [Parameter(Mandatory)]
        [string]$ExpectedOwnerMarker
    )

    return [pscustomobject]@{
        Helper              = 'Assert-HubQaOwnedFixturePath'
        CandidatePath       = $CandidatePath
        OwnedRoot           = $OwnedRoot
        ExpectedOwnerMarker = $ExpectedOwnerMarker
        FutureChecks        = @(
            'canonicalize candidate and owned root before comparison'
            'reject candidate root escape'
            'reject any reparse path'
            'reject resolved-outside path'
            'reject owner-marker mismatch'
        )
        Observation         = 'static declaration only; no path inspection performed'
    }
}

function New-HubQaContractFixtureWorkspace {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$CaseId,

        [Parameter(Mandatory)]
        [string]$ProductRoot
    )

    return [pscustomobject]@{
        Helper                    = 'New-HubQaContractFixtureWorkspace'
        CaseId                    = $CaseId
        ProductRoot               = $ProductRoot
        TemporaryRootContract     = [pscustomobject]@{
            OperatingSystemRoot   = 'operating-system temporary root'
            ClosedPrefix          = 'hubqa-g0-contract-'
            OwnerMarker           = 'hubqa-g0-contract-owner'
            RejectRootEscape      = $true
            RejectReparsePaths    = $true
            RejectMarkerMismatch  = $true
        }
        RepositoryLayoutContract  = @(
            'canonical temporary repository layout'
            'canonical product source resolution'
            'copy testframework/runner/run-task.ps1 into owned temporary layout'
        )
        Observation               = 'static declaration only; no workspace is created or copied'
    }
}

function Reset-HubQaContractFixtureCase {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$WorkspacePath,

        [Parameter(Mandatory)]
        [string]$CaseId
    )

    return [pscustomobject]@{
        Helper                    = 'Reset-HubQaContractFixtureCase'
        WorkspacePath             = $WorkspacePath
        CaseId                    = $CaseId
        FutureResetContract       = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'reset only the selected closed fixture case'
        )
        Observation               = 'static declaration only; no fixture reset is performed'
    }
}

function Get-HubQaContractFixtureTreeManifest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$WorkspacePath
    )

    return [pscustomobject]@{
        Helper                    = 'Get-HubQaContractFixtureTreeManifest'
        WorkspacePath             = $WorkspacePath
        FutureManifestContract    = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'record typed owned and foreign inventory separately'
        )
        Observation               = 'static declaration only; no tree is inspected'
    }
}

function Remove-HubQaContractFixtureWorkspace {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$WorkspacePath,

        [Parameter(Mandatory)]
        [string]$ExpectedOwnerMarker
    )

    return [pscustomobject]@{
        Helper                    = 'Remove-HubQaContractFixtureWorkspace'
        WorkspacePath             = $WorkspacePath
        ExpectedOwnerMarker       = $ExpectedOwnerMarker
        FutureCleanupContract     = 'try/finally -> typed owned/foreign selection -> cleanup owned set -> zero-owned proof'
        FutureCleanupChecks       = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'leave foreign records ineligible for cleanup'
        )
        Observation               = 'static declaration only; no workspace is removed'
    }
}
