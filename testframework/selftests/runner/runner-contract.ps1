# G0-A is a source-only local-image contract harness. No helper is invoked from this file.

function Get-HubQaG0ComplexityBudget {
    [CmdletBinding()]
    param()

    return [pscustomobject]@{
        ContractState                = 'STATIC_CONTRACT_REVIEWED'
        PrivateContractState         = 'PRIVATE_CONTRACT_NOT_RUN'
        LocalImageLockSchema          = 'hub.qa.local-images-lock/v3'
        PrivateDependencyContract      = [pscustomobject][ordered]@{
            OperationCount = 4
            Operations     = [ordered]@{
                InspectLockedImage = [pscustomobject][ordered]@{
                    Parameters = @()
                    Result     = @('Verified', 'ExitCode', 'DeadlineLayer', 'ImageId', 'Entrypoint', 'EntrypointSha256', 'Cmd', 'CmdSha256', 'LabelsProjectionSha256')
                    VerifiedResultRequires = @('ImageId', 'Entrypoint', 'EntrypointSha256', 'Cmd', 'CmdSha256', 'LabelsProjectionSha256')
                }
                RunFixedTarget = [pscustomobject][ordered]@{
                    Parameters = @()
                    Result     = @('ExitCode', 'TimedOut', 'DeadlineLayer', 'StreamCompleted', 'Stdout', 'Stderr', 'StdoutBytes', 'StderrBytes')
                    ByteFields = @('StdoutBytes byte[]', 'StderrBytes byte[]')
                }
                ListOwnedResources = [pscustomobject][ordered]@{
                    Parameters  = @('[int64]DeadlineTick')
                    ResourceType = 'hub.qa.resource/v1'
                    Resource     = @('RecordType', 'Id', 'Kind', 'Name', 'Project')
                }
                CleanupOwnedResources = [pscustomobject][ordered]@{
                    Parameters           = @('[object[]]OwnedResources', '[int64]DeadlineTick')
                    SelectedResourceType = 'hub.qa.owned-resource/v1'
                    Resource             = @('RecordType', 'Id', 'Kind', 'Name', 'Project')
                    Result               = @('Succeeded', 'ExitCode', 'TimedOut')
                }
            }
            DockerNullRule = 'null Entrypoint or Cmd is rejected; only explicit empty JSON arrays are hashable'
            CanonicalHashes = [ordered]@{
                EmptyArray = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
                EmptyObject = 'sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a'
            }
        }
        MutableStateBudget            = [pscustomobject]@{
            Families                 = @('task receipt')
            ReceiptRule              = 'first non-zero exit is immutable'
            AdditionalFamiliesRequire = 're-slicing before implementation'
            PreMutationSequence      = @(
                'request validation'
                'canonical-path validation'
                'duplicate-aware local-image-lock schema validation'
                'top-level control, source, and build binding validation'
                'three-base inventory attestation validation'
                'three-target local-image provenance validation'
            )
            PreMutationFailure       = [pscustomobject]@{
                IntendedExit          = 64
                DependencyCalls       = 0
                ReportsTreeManifest   = 'unchanged'
            }
        }
        HelperBudget                  = [pscustomobject]@{
            FrozenHelpers             = @(
                'Get-HubQaG0ComplexityBudget'
                'Get-HubQaG0StaticCaseDefinitions'
                'Assert-HubQaOwnedFixturePath'
                'New-HubQaContractFixtureWorkspace'
                'Reset-HubQaContractFixtureCase'
                'Get-HubQaContractFixtureTreeManifest'
                'Remove-HubQaContractFixtureWorkspace'
            )
            Count                     = 7
            NewHelpers                = 0
            AdditionalHelpersRequire  = 're-slicing before implementation'
        }
        BranchBudget                  = [pscustomobject]@{
            FrozenExecutableFamilies  = @(
                'request validation'
                'canonical-path validation'
                'local-image-lock validation'
                'receipt finalization'
                'typed cleanup selection'
            )
            DeclarativeFamilies       = @(
                'atomic staging collision'
                'foreign or preexisting staging'
                'final collision'
                'capped source-control controls and invalid object'
                'dirty or raced source-manifest difference'
                'all baked-source images'
                'provenance mismatch'
            )
            FrozenExecutableCount     = 5
            DeclarativeCount          = 7
            NewExecutableBranches     = 0
            AdditionalExecutableBranchesRequire = 're-slicing before implementation'
        }
        SideEffectBudget              = [pscustomobject]@{
            ActualOperations          = @()
            Count                     = 0
            Rule                      = 'G0-A contains no product path, copy, file-system, process, or network operation'
            FutureOperations          = 'owned by a separately sliced implementation contract'
        }
        ChildProcessBudget            = [pscustomobject]@{
            ActualSites               = @()
            Count                     = 0
            Rule                      = 'inert descriptors are declarations, not process operations'
            FutureProductionSites     = 'outside G0-A ownership'
        }
        CleanupBudget                 = [pscustomobject]@{
            ActualRoutes              = @()
            Count                     = 0
            DeclarativeRule           = 'typed owned selection only; recursive staging deletion is forbidden'
            FutureRoutes              = 'one separately sliced typed cleanup route with zero-owned proof'
        }
    }
}

function Get-HubQaG0StaticCaseDefinitions {
    [CmdletBinding()]
    param()

    return @(
        [pscustomobject]@{
            Id       = 'TC-G0-001'
            Intent   = 'valid local-image lock, success receipt, empty inventory, successful cleanup'
            Expected = [pscustomobject]@{
                IntendedExit = 0
                Inventory    = 'empty'
                Cleanup      = 'successful'
                Observation  = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-002'
            Intent   = 'first exit 23 remains immutable after later transitions'
            Expected = [pscustomobject]@{
                PrimaryExit = 23
                FinalExit   = 23
                Observation = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-003'
            Intent   = 'synthetic command-deadline receipt has exit 124 and no child mode'
            Expected = [pscustomobject]@{
                IntendedExit = 124
                ChildMode    = 'none'
                Receipt      = 'synthetic command-deadline'
                Observation  = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-003A'
            Intent   = 'completed prior receipt plus target deadline has exit 124 and no child mode'
            Expected = [pscustomobject]@{
                IntendedExit = 124
                ChildMode    = 'none'
                Receipt      = 'completed prior receipt plus target deadline'
                Observation  = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-004'
            Intent   = 'inert bounded-wait descriptor plus already-expired monotonic tick creates no process'
            Expected = [pscustomobject]@{
                WaitDescriptor  = 'bounded and inert'
                MonotonicTick   = 'already expired'
                ProcessCreation = 0
                Observation     = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-005'
            Intent   = 'zero collection intends final exit 65'
            Expected = [pscustomobject]@{
                Collected         = 0
                IntendedFinalExit = 65
                Observation       = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-006'
            Intent   = 'literal malformed result {not-json intends final exit 65'
            Expected = [pscustomobject]@{
                Result            = '{not-json'
                IntendedFinalExit = 65
                Observation       = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-006A'
            Intent   = 'collected 2, passed 1, skipped 1, empty skip allowlist intends final exit 65'
            Expected = [pscustomobject]@{
                Collected         = 2
                Passed            = 1
                Skipped           = 1
                SkipAllowlist     = @()
                IntendedFinalExit = 65
                Observation       = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-007'
            Intent   = 'invalid request or exact v3 local-image-lock schema/control/source/build/base/label variant rejects before adapter construction or child handling'
            Fixtures = [pscustomobject]@{
                InvalidTagOnlyLockFixture = 'LockBytes.TagOnlyBackendUnit'
                Schema                     = 'hub.qa.local-images-lock/v3'
                TopLevelFields              = @('schema', 'control', 'source', 'build', 'bases', 'targets')
                ControlFields               = @('commit', 'bundle_sha256', 'compose_sha256')
                CanonicalControl            = [ordered]@{
                    Commit       = 'cccccccccccccccccccccccccccccccccccccccc'
                    BundleSha256 = 'sha256:7777777777777777777777777777777777777777777777777777777777777777'
                    ComposeSha256 = 'sha256:8888888888888888888888888888888888888888888888888888888888888888'
                }
                BaseKeys                    = @('golang', 'playwright', 'python')
                TargetKeys                  = @('backend-unit', 'frontend-unit', 'backend-openapi')
                TargetBaseMapping           = @('backend-unit->golang', 'frontend-unit->playwright', 'backend-openapi->python')
                BaseRecordFields            = @('repo_digest', 'image_id', 'platform', 'rootfs_layers', 'entrypoint', 'entrypoint_sha256', 'cmd', 'cmd_sha256', 'labels_projection', 'labels_projection_sha256')
                TargetRecordFields          = @('base_key', 'image_id', 'platform', 'rootfs_layers', 'entrypoint', 'entrypoint_sha256', 'cmd', 'cmd_sha256', 'control_commit', 'control_bundle_sha256', 'source_commit', 'source_policy_sha256', 'source_tree_manifest_sha256', 'build_manifest_sha256', 'labels_projection', 'labels_projection_sha256')
                MandatoryTargetLabels       = @('hub.qa.target', 'hub.qa.base-image-id', 'hub.qa.control-commit', 'hub.qa.control-bundle-sha256', 'hub.qa.source-commit', 'hub.qa.source-policy-sha256', 'hub.qa.source-tree-manifest-sha256', 'hub.qa.build-manifest-sha256', 'hub.qa.runtime-status')
                HashCanonicalization        = 'JCS/RFC8785 canonical UTF-8'
                InvalidLockFixtures         = @('MissingSourceField', 'MismatchedSourceField', 'MissingBuildField', 'MismatchedBuildField', 'MissingBaseLabel', 'MismatchedBaseLabel', 'ForbiddenTargetRepoDigest', 'MissingControlField', 'MismatchedTargetControlCommit', 'MismatchedTargetControlBundleSha256', 'MissingControlCommitLabel', 'MismatchedControlBundleLabel', 'MismatchedControlComposeSha256', 'V2SchemaRejected')
                InvalidVariants            = @('contract', 'target', 'Task', 'Run', 'lock')
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
                Rejections  = @('existing directory', 'mismatched owner marker', 'reparse point', 'resolved-outside path')
                Observation = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-009'
            Intent   = 'exactly two owned and two foreign inventory records; only owned typed records are eligible for inert cleanup'
            Expected = [pscustomobject]@{
                OwnedRecords       = 2
                ForeignRecords     = 2
                EligibleForCleanup = 'owned typed records only'
                Cleanup            = 'inert declaration only'
                Observation        = 'static source-review intent only'
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
        [pscustomobject]@{
            Id       = 'TC-G0-011'
            Intent   = 'atomic staging collision is rejected before generation or final synchronization'
            Expected = [pscustomobject]@{
                StagingDisposition = 'create-new only'
                FinalSynchronization = 'not reached'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-012'
            Intent   = 'foreign or preexisting staging is rejected without adopting or deleting it'
            Expected = [pscustomobject]@{
                ForeignStaging      = 'preserved'
                PreexistingStaging  = 'preserved'
                Generation          = 'not reached'
                Observation         = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-013'
            Intent   = 'final destination collision is rejected before final synchronization'
            Expected = [pscustomobject]@{
                FinalDestination = 'not overwritten'
                Observation      = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-014'
            Intent   = 'staging cleanup has no recursive delete route'
            Expected = [pscustomobject]@{
                RecursiveStagingDelete = 'forbidden'
                Observation            = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-015'
            Intent   = 'capped Git controls reject an invalid revision object'
            Expected = [pscustomobject]@{
                MaximumGitControls = 3
                InvalidObject      = 'reject-before-staging'
                Observation        = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-016'
            Intent   = 'dirty or raced source-manifest difference rejects final synchronization'
            Expected = [pscustomobject]@{
                SourceManifestEquality = 'required'
                FinalSynchronization   = 'not reached'
                Observation            = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-017'
            Intent   = 'all three target images bind the exact top-level source and build values and their mapped base identities'
            Expected = [pscustomobject]@{
                RequiredBases                 = @('golang', 'playwright', 'python')
                RequiredTargets               = @('backend-unit', 'frontend-unit', 'backend-openapi')
                RequiredBaseCount             = 3
                RequiredTargetCount           = 3
                SourceAndBuildEqualTopLevel    = $true
                TargetRepoDigest               = 'forbidden'
                BaseRepoDigest                 = 'inventory only'
                Observation                    = 'static source-review intent only'
            }
        }
        [pscustomobject]@{
            Id       = 'TC-G0-018'
            Intent   = 'selected target inspect or separate all-base/all-target attestation mismatch rejects before target execution'
            Expected = [pscustomobject]@{
                SelectedInspectComparison = @('ImageId', 'Platform', 'RootFsLayers', 'Config.Entrypoint', 'Config.Cmd', 'Config.Labels', 'projection hashes')
                SeparateAttestorScope      = 'all bases and all targets'
                FullObservedLabels         = 'equal locked labels_projection'
                Execution                  = 'not reached'
                Observation                = 'static source-review intent only'
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
        [string]$CaseId
    )

    return [pscustomobject]@{
        Helper                = 'New-HubQaContractFixtureWorkspace'
        CaseId                = $CaseId
        TemporaryRootContract = [pscustomobject]@{
            OperatingSystemRoot  = 'operating-system temporary root'
            ClosedPrefix         = 'hubqa-g0-contract-'
            OwnerMarker          = 'hubqa-g0-contract-owner'
            RejectRootEscape     = $true
            RejectReparsePaths   = $true
            RejectMarkerMismatch = $true
        }
        WorkspaceContract     = @(
            'closed inert fixture namespace'
            'no product source input'
            'no copy operation'
        )
        Observation           = 'static declaration only; no workspace is created'
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
        Helper              = 'Reset-HubQaContractFixtureCase'
        WorkspacePath       = $WorkspacePath
        CaseId              = $CaseId
        FutureResetContract = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'select only the closed fixture case'
        )
        Observation         = 'static declaration only; no fixture reset is performed'
    }
}

function Get-HubQaContractFixtureTreeManifest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$WorkspacePath
    )

    return [pscustomobject]@{
        Helper                 = 'Get-HubQaContractFixtureTreeManifest'
        WorkspacePath          = $WorkspacePath
        FutureManifestContract = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'record typed owned and foreign inventory separately'
        )
        Observation            = 'static declaration only; no tree is inspected'
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
        Helper                = 'Remove-HubQaContractFixtureWorkspace'
        WorkspacePath         = $WorkspacePath
        ExpectedOwnerMarker   = $ExpectedOwnerMarker
        FutureCleanupContract = 'typed owned selection only; recursive staging deletion forbidden; zero-owned proof required'
        FutureCleanupChecks   = @(
            'assert owned canonical workspace path'
            'reject root escape and reparse paths'
            'reject owner-marker mismatch'
            'leave foreign records ineligible for cleanup'
        )
        Observation           = 'static declaration only; no workspace is removed'
    }
}
