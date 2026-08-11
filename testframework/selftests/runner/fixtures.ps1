function Get-HubQaFixtureCatalog {
    return [ordered]@{
        LockBytes = [ordered]@{
            Valid = '{"backend-unit":"example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","frontend-unit":"example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","backend-openapi":"example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}'
            Malformed = '{not-json'
            DuplicateBackendUnit = '{"backend-unit":"example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","backend-unit":"example.invalid/backend-unit@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","frontend-unit":"example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","backend-openapi":"example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}'
            NonStringFrontendUnit = '{"backend-unit":"example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","frontend-unit":17,"backend-openapi":"example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}'
            TagOnlyBackendUnit = '{"backend-unit":"registry/unit:latest","frontend-unit":"example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","backend-openapi":"example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}'
            MissingBackendOpenapi = '{"backend-unit":"example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","frontend-unit":"example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}'
            FourthSentinelImage = '{"backend-unit":"example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","frontend-unit":"example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","backend-openapi":"example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","sentinel-image":"example.invalid/sentinel-image@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}'
        }
        TypedImageIdentities = [ordered]@{
            BackendUnit = 'example.invalid/backend-unit@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
            FrontendUnit = 'example.invalid/frontend-unit@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'
            BackendOpenapi = 'example.invalid/backend-openapi@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'
        }
        ReceiptBytes = [ordered]@{
            Success = '{"collected":1,"passed":1,"skipped":0,"first_exit_code":0,"results":[{"name":"unit","exit_code":0}]}'
            Exit23 = '{"collected":1,"passed":0,"skipped":0,"first_exit_code":23,"results":[{"name":"unit","exit_code":23}]}'
            CommandDeadline = '{"collected":1,"passed":0,"skipped":0,"first_exit_code":124,"results":[{"name":"build","exit_code":0},{"name":"unit","exit_code":124,"deadline_layer":"command"}]}'
            TargetDeadline = '{"collected":1,"passed":0,"skipped":0,"first_exit_code":124,"results":[{"name":"build","exit_code":0},{"name":"unit","exit_code":0},{"name":"target","exit_code":124,"deadline_layer":"target"}]}'
            ZeroCollection = '{"collected":0,"passed":0,"skipped":0,"first_exit_code":0,"results":[]}'
            MalformedResult = '{not-json'
            UnexpectedSkip = '{"collected":2,"passed":1,"skipped":1,"unexpected_skip_count":1,"allowlisted_skip_count":0,"first_exit_code":0,"results":[{"name":"unit","exit_code":0}]}'
        }
        RequestCases = [ordered]@{
            InvalidContract = [ordered]@{ Contract = 'hub.qa-runner/v0'; Target = 'backend-unit'; Task = 'task-1'; Run = 'run-1' }
            InvalidTarget = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'sentinel'; Task = 'task-1'; Run = 'run-1' }
            InvalidTask = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'backend-unit'; Task = 'Bad_ID'; Run = 'run-1' }
            InvalidRun = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'backend-unit'; Task = 'task-1'; Run = '../escape' }
        }
        OwnerPathCases = [ordered]@{
            ExistingOwnerDirectory = [ordered]@{ Kind = 'existing-owner-directory'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            MismatchedOwnerMarker = [ordered]@{ Kind = 'mismatched-owner-marker'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            InjectedReparsePoint = [ordered]@{ Kind = 'injected-reparse-point'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            ResolvedOutsideRoot = [ordered]@{ Kind = 'resolved-outside-root'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
        }
        ResourceInventories = [ordered]@{
            Empty = @()
            Mixed = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Ownership = 'owned' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Ownership = 'owned' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-container-a'; Kind = 'container'; Ownership = 'foreign' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-volume-a'; Kind = 'volume'; Ownership = 'foreign' }
            )
            ForeignOnly = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-container-b'; Kind = 'container'; Ownership = 'foreign' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-network-b'; Kind = 'network'; Ownership = 'foreign' }
            )
            OwnedResidual = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Ownership = 'owned' }
            )
        }
        CleanupOutcomes = [ordered]@{
            Success = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null }
            Obstructed = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'obstructed'; ResidualOwnedCount = 1; FailureCode = 71 }
        }
        DependencyCases = [ordered]@{
            'TC-G0-001' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-001'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-002' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-002'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-003' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-003'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-003A' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-003A'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-004' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-004'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-005' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-005'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-006' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-006'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-006A' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-006A'; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null } }
            'TC-G0-009' = [pscustomobject][ordered]@{
                CaseRecordType = 'hub.qa.fixture-dependency-case/v1'
                CaseId = 'TC-G0-009'
                InjectedResourceInventory = @(
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Ownership = 'owned' },
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Ownership = 'owned' },
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-container-a'; Kind = 'container'; Ownership = 'foreign' },
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-foreign-resource/v1'; Id = 'foreign-volume-a'; Kind = 'volume'; Ownership = 'foreign' }
                )
                SelectedOwnedResources = @(
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Ownership = 'owned' },
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Ownership = 'owned' }
                )
                CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'success'; ResidualOwnedCount = 0; FailureCode = $null }
            }
            'TC-G0-010' = [pscustomobject][ordered]@{
                CaseRecordType = 'hub.qa.fixture-dependency-case/v1'
                CaseId = 'TC-G0-010'
                InjectedResourceInventory = @(
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Ownership = 'owned' }
                )
                SelectedOwnedResources = @(
                    [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-owned-resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Ownership = 'owned' }
                )
                CleanupOutcome = [pscustomobject][ordered]@{ RecordType = 'hub.qa.fixture-cleanup-outcome/v1'; Status = 'obstructed'; ResidualOwnedCount = 1; FailureCode = 71 }
            }
        }
        CaseExpectations = [ordered]@{
            'TC-G0-001' = [ordered]@{ Fixture = 'Success'; ExpectedExitCode = 0; States = @('validated', 'reported', 'cleaned') }
            'TC-G0-002' = [ordered]@{ Fixture = 'Exit23'; ExpectedExitCode = 23; States = @('validated', 'reported', 'cleaned') }
            'TC-G0-003' = [ordered]@{ Fixture = 'CommandDeadline'; ExpectedExitCode = 124; DeadlineLayer = 'command' }
            'TC-G0-003A' = [ordered]@{ Fixture = 'TargetDeadline'; ExpectedExitCode = 124; DeadlineLayer = 'target' }
            'TC-G0-004' = [ordered]@{ Fixture = 'bounded-wait'; ExpectedExitCode = 124; DeadlineLayer = 'outer' }
            'TC-G0-005' = [ordered]@{ Fixture = 'ZeroCollection'; ExpectedExitCode = 65 }
            'TC-G0-006' = [ordered]@{ Fixture = 'MalformedResult'; ExpectedExitCode = 65 }
            'TC-G0-006A' = [ordered]@{ Fixture = 'UnexpectedSkip'; ExpectedExitCode = 65 }
            'TC-G0-007' = [ordered]@{ Fixtures = @('InvalidContract', 'InvalidTarget', 'InvalidTask', 'InvalidRun', 'Malformed', 'DuplicateBackendUnit', 'NonStringFrontendUnit', 'TagOnlyBackendUnit', 'MissingBackendOpenapi', 'FourthSentinelImage'); ExpectedExitCode = 64; DependencyCalls = 0; ReportsManifest = 'unchanged'; Stage = 'pre-child-pre-adapter' }
            'TC-G0-008' = [ordered]@{ Fixtures = @('ExistingOwnerDirectory', 'MismatchedOwnerMarker', 'InjectedReparsePoint', 'ResolvedOutsideRoot'); Expected = 'original-bytes-preserved'; Stage = 'pre-child' }
            'TC-G0-009' = [ordered]@{ Fixture = 'Mixed'; SelectedOwnership = 'owned'; SelectedCount = 2; ForeignSelectedCount = 0 }
            'TC-G0-010' = [ordered]@{ Fixture = 'Obstructed'; Primary23Final = 23; Primary0Final = 71 }
        }
    }
}

function Get-HubQaInertChildDefinition {
    return [ordered]@{
        'exit-0' = [ordered]@{ Name = 'exit-0'; Kind = 'inert-child-descriptor'; ExitCode = 0 }
        'exit-23' = [ordered]@{ Name = 'exit-23'; Kind = 'inert-child-descriptor'; ExitCode = 23 }
        'bounded-wait' = [ordered]@{ Name = 'bounded-wait'; Kind = 'inert-child-descriptor'; StartTick = 0; ObservedTick = 811; Scope = 'outer-deadline-contract-only' }
    }
}
