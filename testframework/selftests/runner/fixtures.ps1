function Get-HubQaFixtureCatalog {
    $sourceCommit = '0123456789abcdef0123456789abcdef01234567'
    $sourcePolicySha256 = 'sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd'
    $sourceTreeManifestSha256 = 'sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee'
    $buildManifestSha256 = 'sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff'
    $controlCommit = 'cccccccccccccccccccccccccccccccccccccccc'
    $controlBundleSha256 = 'sha256:7777777777777777777777777777777777777777777777777777777777777777'
    $controlComposeSha256 = 'sha256:8888888888888888888888888888888888888888888888888888888888888888'

    $backendUnitLabels = [ordered]@{
        'hub.qa.target'                      = 'backend-unit'
        'hub.qa.base-image-id'               = 'sha256:1111111111111111111111111111111111111111111111111111111111111111'
        'hub.qa.control-commit'              = $controlCommit
        'hub.qa.control-bundle-sha256'       = $controlBundleSha256
        'hub.qa.source-commit'               = $sourceCommit
        'hub.qa.source-policy-sha256'        = $sourcePolicySha256
        'hub.qa.source-tree-manifest-sha256' = $sourceTreeManifestSha256
        'hub.qa.build-manifest-sha256'       = $buildManifestSha256
        'hub.qa.runtime-status'              = 'NOT_READY'
    }
    $frontendUnitLabels = [ordered]@{
        'hub.qa.target'                      = 'frontend-unit'
        'hub.qa.base-image-id'               = 'sha256:7777777777777777777777777777777777777777777777777777777777777777'
        'hub.qa.control-commit'              = $controlCommit
        'hub.qa.control-bundle-sha256'       = $controlBundleSha256
        'hub.qa.source-commit'               = $sourceCommit
        'hub.qa.source-policy-sha256'        = $sourcePolicySha256
        'hub.qa.source-tree-manifest-sha256' = $sourceTreeManifestSha256
        'hub.qa.build-manifest-sha256'       = $buildManifestSha256
        'hub.qa.runtime-status'              = 'NOT_READY'
    }
    $backendOpenapiLabels = [ordered]@{
        'hub.qa.target'                      = 'backend-openapi'
        'hub.qa.base-image-id'               = 'sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd'
        'hub.qa.control-commit'              = $controlCommit
        'hub.qa.control-bundle-sha256'       = $controlBundleSha256
        'hub.qa.source-commit'               = $sourceCommit
        'hub.qa.source-policy-sha256'        = $sourcePolicySha256
        'hub.qa.source-tree-manifest-sha256' = $sourceTreeManifestSha256
        'hub.qa.build-manifest-sha256'       = $buildManifestSha256
        'hub.qa.runtime-status'              = 'NOT_READY'
    }
    $golangBaseLabels = [ordered]@{
        'hub.qa.fixture-base' = 'golang'
    }
    $playwrightBaseLabels = [ordered]@{
        'hub.qa.fixture-base' = 'playwright'
    }
    $pythonBaseLabels = [ordered]@{
        'hub.qa.fixture-base' = 'python'
    }

    $golangBaseInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-base-inspect/v2'
        BaseKey                = 'golang'
        RepoDigest             = 'example.invalid/hubqa/golang@sha256:0000000000000000000000000000000000000000000000000000000000000000'
        ImageId                = 'sha256:1111111111111111111111111111111111111111111111111111111111111111'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @('sha256:2222222222222222222222222222222222222222222222222222222222222222')
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/usr/local/bin/docker-entrypoint.sh')
            Cmd        = @()
            Labels     = $golangBaseLabels
        }
        EntrypointSha256       = 'sha256:19d8807fdab2431c85033d5f7dd076461a0b9aa72fe0306983a16f7c5a210be9'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:ae1bee75b806d4ccf70eda0f045444f5f0594e25b2bd31045e57033b63361d8e'
    }
    $playwrightBaseInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-base-inspect/v2'
        BaseKey                = 'playwright'
        RepoDigest             = 'example.invalid/hubqa/playwright@sha256:6666666666666666666666666666666666666666666666666666666666666666'
        ImageId                = 'sha256:7777777777777777777777777777777777777777777777777777777777777777'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @('sha256:8888888888888888888888888888888888888888888888888888888888888888')
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/bin/sh')
            Cmd        = @()
            Labels     = $playwrightBaseLabels
        }
        EntrypointSha256       = 'sha256:c6b105373ceb55c28347e3d08ffec03f8c95ad6b943d9a51f230bd0d5920ea3f'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:24f3654bbc23f4735e584d5e9bdd7f35fa051c2769463fd2791fc76e422363c0'
    }
    $pythonBaseInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-base-inspect/v2'
        BaseKey                = 'python'
        RepoDigest             = 'example.invalid/hubqa/python@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'
        ImageId                = 'sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @('sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee')
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/usr/local/bin/python3')
            Cmd        = @()
            Labels     = $pythonBaseLabels
        }
        EntrypointSha256       = 'sha256:464e2225d2a732d43a202d5f6a3828ac811eaab0f0c18d55a7d57d14b58a8c1e'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:09cf4eeb582db113d7515562ddd61d6490f717899f90d8448154813b3908de6f'
    }

    $backendUnitInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-target-inspect/v2'
        Target                 = 'backend-unit'
        ImageId                = 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @(
            'sha256:1111111111111111111111111111111111111111111111111111111111111111',
            'sha256:2222222222222222222222222222222222222222222222222222222222222222'
        )
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/opt/hub/bin/backend-unit')
            Cmd        = @()
            Labels     = $backendUnitLabels
        }
        EntrypointSha256       = 'sha256:8bf55017c5d3d88a50aa953aa320eb377f351387a1c4eec3ec15c3d0380d2f54'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:f94b7fd9d382092827ea20b5c017e28dd676ee1b14f9bfffc2b09b8efb0a4910'
    }
    $frontendUnitInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-target-inspect/v2'
        Target                 = 'frontend-unit'
        ImageId                = 'sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @(
            'sha256:3333333333333333333333333333333333333333333333333333333333333333',
            'sha256:4444444444444444444444444444444444444444444444444444444444444444'
        )
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/opt/hub/bin/frontend-unit')
            Cmd        = @()
            Labels     = $frontendUnitLabels
        }
        EntrypointSha256       = 'sha256:5b52deb74a4390a996ab8a2ceb8de11628435966d88d3069f67c1fefec80b8a1'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:6c0494067bccf88c8f2f146a771e95881b5e7506889066de642b103a57d85321'
    }
    $backendOpenapiInspect = [pscustomobject][ordered]@{
        RecordType             = 'hub.qa.fixture-local-target-inspect/v2'
        Target                 = 'backend-openapi'
        ImageId                = 'sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'
        Platform               = [pscustomobject][ordered]@{ Os = 'linux'; Architecture = 'amd64'; Variant = $null }
        RootFsLayers           = @(
            'sha256:5555555555555555555555555555555555555555555555555555555555555555',
            'sha256:6666666666666666666666666666666666666666666666666666666666666666'
        )
        Config                 = [pscustomobject][ordered]@{
            Entrypoint = @('/opt/hub/bin/backend-openapi')
            Cmd        = @()
            Labels     = $backendOpenapiLabels
        }
        EntrypointSha256       = 'sha256:7ad7c94780509d74afcf2adbc940d605a61449bd899a8de0f24ec58668ae7b7f'
        CmdSha256              = 'sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945'
        LabelsProjectionSha256 = 'sha256:000a27615e45dda1720381ef30bdede916ce30eb224cbba5e9759c610be14f85'
    }

    $successfulCleanupOutcome = [pscustomobject][ordered]@{ Succeeded = $true; ExitCode = 0; TimedOut = $false }
    $obstructedCleanupOutcome = [pscustomobject][ordered]@{ Succeeded = $false; ExitCode = 71; TimedOut = $false }
    $ownedContainer = [pscustomobject][ordered]@{ RecordType = 'hub.qa.owned-resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Name = 'hubqa-task-1-run-1-backend-unit'; Project = 'hubqa-task-1-run-1' }
    $ownedNetwork = [pscustomobject][ordered]@{ RecordType = 'hub.qa.owned-resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Name = 'hubqa-task-1-run-1_default'; Project = 'hubqa-task-1-run-1' }
    $ownedResidual = [pscustomobject][ordered]@{ RecordType = 'hub.qa.owned-resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Name = 'hubqa-task-1-run-1-data'; Project = 'hubqa-task-1-run-1' }
    $mixedResourceInventory = @(
        [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Name = 'hubqa-task-1-run-1-backend-unit'; Project = 'hubqa-task-1-run-1' },
        [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Name = 'hubqa-task-1-run-1_default'; Project = 'hubqa-task-1-run-1' },
        [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-container-a'; Kind = 'container'; Name = 'foreign-container-a'; Project = 'foreign-project' },
        [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-volume-a'; Kind = 'volume'; Name = 'foreign-volume-a'; Project = 'foreign-project' }
    )
    $ownedResidualInventory = @(
        [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Name = 'hubqa-task-1-run-1-data'; Project = 'hubqa-task-1-run-1' }
    )

    return [ordered]@{
        CanonicalHashContract = [ordered]@{
            Algorithm        = 'SHA-256'
            Canonicalization = 'JCS/RFC8785'
            Encoding         = 'UTF-8'
            DigestForm       = 'sha256 plus colon plus 64 lowercase hexadecimal characters'
            Projections      = @('entrypoint JSON array', 'cmd JSON array', 'labels_projection JSON object')
            DockerNullRule   = 'null Entrypoint or Cmd is rejected; only explicit empty JSON arrays are hashable'
        }
        LockBytes = [ordered]@{
            InvalidProjectionTemplate = $invalidProjectionHashes = (@'
{"schema":"hub.qa.local-images-lock/v3","control":{"commit":"cccccccccccccccccccccccccccccccccccccccc","bundle_sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","compose_sha256":"sha256:8888888888888888888888888888888888888888888888888888888888888888"},"source":{"commit":"0123456789abcdef0123456789abcdef01234567","policy_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","tree_manifest_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"},"build":{"manifest_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"bases":{"golang":{"repo_digest":"example.invalid/hubqa/golang@sha256:0000000000000000000000000000000000000000000000000000000000000000","image_id":"sha256:1111111111111111111111111111111111111111111111111111111111111111","platform":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:2222222222222222222222222222222222222222222222222222222222222222"],"entrypoint":["/usr/local/bin/docker-entrypoint.sh"],"entrypoint_sha256":"sha256:3333333333333333333333333333333333333333333333333333333333333333","cmd":[],"cmd_sha256":"sha256:4444444444444444444444444444444444444444444444444444444444444444","labels_projection":{},"labels_projection_sha256":"sha256:5555555555555555555555555555555555555555555555555555555555555555"},"playwright":{"repo_digest":"example.invalid/hubqa/playwright@sha256:6666666666666666666666666666666666666666666666666666666666666666","image_id":"sha256:7777777777777777777777777777777777777777777777777777777777777777","platform":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:8888888888888888888888888888888888888888888888888888888888888888"],"entrypoint":["/bin/sh"],"entrypoint_sha256":"sha256:9999999999999999999999999999999999999999999999999999999999999999","cmd":[],"cmd_sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","labels_projection":{},"labels_projection_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},"python":{"repo_digest":"example.invalid/hubqa/python@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","image_id":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","platform":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"],"entrypoint":["/usr/local/bin/python3"],"entrypoint_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","cmd":[],"cmd_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","labels_projection":{},"labels_projection_sha256":"sha256:1111111111111111111111111111111111111111111111111111111111111111"}},"targets":{"backend-unit":{"base_key":"golang","image_id":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","platform":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:1111111111111111111111111111111111111111111111111111111111111111","sha256:2222222222222222222222222222222222222222222222222222222222222222"],"entrypoint":["/opt/hub/bin/backend-unit"],"entrypoint_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","cmd":[],"cmd_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","control_commit":"cccccccccccccccccccccccccccccccccccccccc","control_bundle_sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","source_commit":"0123456789abcdef0123456789abcdef01234567","source_policy_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","source_tree_manifest_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","build_manifest_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","labels_projection":{"hub.qa.target":"backend-unit","hub.qa.base-image-id":"sha256:1111111111111111111111111111111111111111111111111111111111111111","hub.qa.control-commit":"cccccccccccccccccccccccccccccccccccccccc","hub.qa.control-bundle-sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","hub.qa.source-commit":"0123456789abcdef0123456789abcdef01234567","hub.qa.source-policy-sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","hub.qa.source-tree-manifest-sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},"frontend-unit":{"base_key":"playwright","image_id":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","platform":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:3333333333333333333333333333333333333333333333333333333333333333","sha256:4444444444444444444444444444444444444444444444444444444444444444"],"entrypoint":["/opt/hub/bin/frontend-unit"],"entrypoint_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","cmd":[],"cmd_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","control_commit":"cccccccccccccccccccccccccccccccccccccccc","control_bundle_sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","source_commit":"0123456789abcdef0123456789abcdef01234567","source_policy_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","source_tree_manifest_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","build_manifest_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","labels_projection":{"hub.qa.target":"frontend-unit","hub.qa.base-image-id":"sha256:7777777777777777777777777777777777777777777777777777777777777777","hub.qa.control-commit":"cccccccccccccccccccccccccccccccccccccccc","hub.qa.control-bundle-sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","hub.qa.source-commit":"0123456789abcdef0123456789abcdef01234567","hub.qa.source-policy-sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","hub.qa.source-tree-manifest-sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"},"backend-openapi":{"base_key":"python","image_id":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","platform":{"os":"linux","architecture":{"os":"linux","architecture":"amd64","variant":null},"rootfs_layers":["sha256:5555555555555555555555555555555555555555555555555555555555555555","sha256:6666666666666666666666666666666666666666666666666666666666666666"],"entrypoint":["/opt/hub/bin/backend-openapi"],"entrypoint_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","cmd":[],"cmd_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","control_commit":"cccccccccccccccccccccccccccccccccccccccc","control_bundle_sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","source_commit":"0123456789abcdef0123456789abcdef01234567","source_policy_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","source_tree_manifest_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","build_manifest_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","labels_projection":{"hub.qa.target":"backend-openapi","hub.qa.base-image-id":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","hub.qa.control-commit":"cccccccccccccccccccccccccccccccccccccccc","hub.qa.control-bundle-sha256":"sha256:7777777777777777777777777777777777777777777777777777777777777777","hub.qa.source-commit":"0123456789abcdef0123456789abcdef01234567","hub.qa.source-policy-sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","hub.qa.source-tree-manifest-sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}}}
'@).Replace(
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"',
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"'
            ).Replace(
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"',
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"'
            ).Replace(
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"',
                '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"'
            )
            Valid = $validLock = (
                $invalidProjectionHashes.Replace(
                    '"entrypoint_sha256":"sha256:3333333333333333333333333333333333333333333333333333333333333333"',
                    '"entrypoint_sha256":"sha256:19d8807fdab2431c85033d5f7dd076461a0b9aa72fe0306983a16f7c5a210be9"'
                ).Replace(
                    '"entrypoint_sha256":"sha256:9999999999999999999999999999999999999999999999999999999999999999"',
                    '"entrypoint_sha256":"sha256:c6b105373ceb55c28347e3d08ffec03f8c95ad6b943d9a51f230bd0d5920ea3f"'
                ).Replace(
                    '"entrypoint_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"',
                    '"entrypoint_sha256":"sha256:464e2225d2a732d43a202d5f6a3828ac811eaab0f0c18d55a7d57d14b58a8c1e"'
                ).Replace(
                    '"entrypoint_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"',
                    '"entrypoint_sha256":"sha256:8bf55017c5d3d88a50aa953aa320eb377f351387a1c4eec3ec15c3d0380d2f54"'
                ).Replace(
                    '"entrypoint_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"',
                    '"entrypoint_sha256":"sha256:5b52deb74a4390a996ab8a2ceb8de11628435966d88d3069f67c1fefec80b8a1"'
                ).Replace(
                    '"entrypoint_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"',
                    '"entrypoint_sha256":"sha256:7ad7c94780509d74afcf2adbc940d605a61449bd899a8de0f24ec58668ae7b7f"'
                ).Replace(
                    '"cmd_sha256":"sha256:4444444444444444444444444444444444444444444444444444444444444444"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"cmd_sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"cmd_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"cmd_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"cmd_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"cmd_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"',
                    '"cmd_sha256":"sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"'
                ).Replace(
                    '"labels_projection":{},"labels_projection_sha256":"sha256:5555555555555555555555555555555555555555555555555555555555555555"',
                    '"labels_projection":{"hub.qa.fixture-base":"golang"},"labels_projection_sha256":"sha256:ae1bee75b806d4ccf70eda0f045444f5f0594e25b2bd31045e57033b63361d8e"'
                ).Replace(
                    '"labels_projection":{},"labels_projection_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"',
                    '"labels_projection":{"hub.qa.fixture-base":"playwright"},"labels_projection_sha256":"sha256:24f3654bbc23f4735e584d5e9bdd7f35fa051c2769463fd2791fc76e422363c0"'
                ).Replace(
                    '"labels_projection":{},"labels_projection_sha256":"sha256:1111111111111111111111111111111111111111111111111111111111111111"',
                    '"labels_projection":{"hub.qa.fixture-base":"python"},"labels_projection_sha256":"sha256:09cf4eeb582db113d7515562ddd61d6490f717899f90d8448154813b3908de6f"'
                ).Replace(
                    '"labels_projection_sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"',
                    '"labels_projection_sha256":"sha256:25d380aa426ecde22708f575b5ab6b6eaee9bae708e771cb5eeab48ba6f388d3"'
                ).Replace(
                    '"labels_projection_sha256":"sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"',
                    '"labels_projection_sha256":"sha256:58e0dbe7a1fcbd22d006f684a691201b960dc13d40ce03b3de1eac9c202be4a8"'
                ).Replace(
                    '"labels_projection_sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"',
                    '"labels_projection_sha256":"sha256:647b69a995f82af86ab293a00606455dc70001e05d202d0c411548df87b950eb"'
                ).Replace(
                    '"platform":{"os":"linux","architecture":{"os":"linux","architecture":"amd64","variant":null}',
                    '"platform":{"os":"linux","architecture":"amd64","variant":null}'
                ).Replace(
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:25d380aa426ecde22708f575b5ab6b6eaee9bae708e771cb5eeab48ba6f388d3"',
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:f94b7fd9d382092827ea20b5c017e28dd676ee1b14f9bfffc2b09b8efb0a4910"'
                ).Replace(
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:58e0dbe7a1fcbd22d006f684a691201b960dc13d40ce03b3de1eac9c202be4a8"',
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:6c0494067bccf88c8f2f146a771e95881b5e7506889066de642b103a57d85321"'
                ).Replace(
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"labels_projection_sha256":"sha256:647b69a995f82af86ab293a00606455dc70001e05d202d0c411548df87b950eb"',
                    '"hub.qa.build-manifest-sha256":"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","hub.qa.runtime-status":"NOT_READY"},"labels_projection_sha256":"sha256:000a27615e45dda1720381ef30bdede916ce30eb224cbba5e9759c610be14f85"'
                )
            )
            InvalidProjectionHashes = $validLock.Replace(
                '"labels_projection_sha256":"sha256:f94b7fd9d382092827ea20b5c017e28dd676ee1b14f9bfffc2b09b8efb0a4910"',
                '"labels_projection_sha256":"sha256:9999999999999999999999999999999999999999999999999999999999999999"'
            )
            MissingControlField = $validLock.Replace(',"compose_sha256":"sha256:8888888888888888888888888888888888888888888888888888888888888888"}', '}')
            MismatchedTargetControlCommit = [regex]::Replace($validLock, '(?s)("targets":\{"backend-unit":.*?"control_commit":)"cccccccccccccccccccccccccccccccccccccccc"', '$1"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"', 1)
            MismatchedTargetControlBundleSha256 = [regex]::Replace($validLock, '(?s)("targets":\{"backend-unit":.*?"control_bundle_sha256":)"sha256:7777777777777777777777777777777777777777777777777777777777777777"', '$1"sha256:9999999999999999999999999999999999999999999999999999999999999999"', 1)
            MissingControlCommitLabel = [regex]::Replace(
                [regex]::Replace($validLock, '(?s)("targets":\{"backend-unit":.*?),"hub\.qa\.control-commit":"cccccccccccccccccccccccccccccccccccccccc"', '$1', 1),
                '(?s)("targets":\{"backend-unit":.*?"labels_projection_sha256":)"sha256:f94b7fd9d382092827ea20b5c017e28dd676ee1b14f9bfffc2b09b8efb0a4910"',
                '$1"sha256:c7a633be78bff58a970b9ab27adcc67dc9f7d999d78ff856112fb8b49346031b"',
                1
            )
            MismatchedControlBundleLabel = [regex]::Replace(
                [regex]::Replace($validLock, '(?s)("targets":\{"backend-unit":.*?"hub\.qa\.control-bundle-sha256":)"sha256:7777777777777777777777777777777777777777777777777777777777777777"', '$1"sha256:9999999999999999999999999999999999999999999999999999999999999999"', 1),
                '(?s)("targets":\{"backend-unit":.*?"labels_projection_sha256":)"sha256:f94b7fd9d382092827ea20b5c017e28dd676ee1b14f9bfffc2b09b8efb0a4910"',
                '$1"sha256:d883c1cc0e65223801fa4a288e91b615f9c96715a72d73dd9dcb7fe1bd809cbd"',
                1
            )
            MismatchedControlComposeSha256 = $validLock.Replace('"compose_sha256":"sha256:8888888888888888888888888888888888888888888888888888888888888888"', '"compose_sha256":"sha256:9999999999999999999999999999999999999999999999999999999999999999"')
            V2SchemaRejected = $validLock.Replace('"schema":"hub.qa.local-images-lock/v3"', '"schema":"hub.qa.local-images-lock/v2"')
            Malformed = '{not-json'
            InvalidSchema = '{"schema":"hub.qa.local-images-lock/v1","targets":{}}'
            DuplicateBackendUnit = '{"schema":"hub.qa.local-images-lock/v3","targets":{"backend-unit":{},"backend-unit":{},"frontend-unit":{},"backend-openapi":{}}}'
            NonStringFrontendUnit = '{"schema":"hub.qa.local-images-lock/v3","targets":{"backend-unit":{},"frontend-unit":17,"backend-openapi":{}}}'
            TagOnlyBackendUnit = '{"schema":"hub.qa.local-images-lock/v3","targets":{"backend-unit":{"image_id":"registry/unit:latest"},"frontend-unit":{},"backend-openapi":{}}}'
            MissingBackendOpenapi = '{"schema":"hub.qa.local-images-lock/v3","targets":{"backend-unit":{},"frontend-unit":{}}}'
            FourthSentinelImage = '{"schema":"hub.qa.local-images-lock/v3","targets":{"backend-unit":{},"frontend-unit":{},"backend-openapi":{},"sentinel-image":{}}}'
        }
        LockSchemaNegativeCases = [ordered]@{
            InvalidProjectionHashes = [ordered]@{
                BaseFixture = 'LockBytes.InvalidProjectionHashes'
                Expected    = 'reject empty labels_projection and JCS projection-hash mismatches'
            }
            MissingSourceField = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'remove targets.backend-unit.source_policy_sha256'
                Expected    = 'reject exact TargetRecord schema'
            }
            MismatchedSourceField = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.source_tree_manifest_sha256 unequal to source.tree_manifest_sha256'
                Expected    = 'reject target-to-top-level source binding'
            }
            MissingBuildField = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'remove build.manifest_sha256'
                Expected    = 'reject exact build schema'
            }
            MismatchedBuildField = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.build_manifest_sha256 unequal to build.manifest_sha256'
                Expected    = 'reject target-to-top-level build binding'
            }
            MissingBaseLabel = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'remove targets.backend-unit.labels_projection.hub.qa.base-image-id'
                Expected    = 'reject missing mandatory target label'
            }
            MismatchedBaseLabel = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.labels_projection.hub.qa.base-image-id unequal to bases.golang.image_id'
                Expected    = 'reject target-to-base identity binding'
            }
            ForbiddenTargetRepoDigest = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'add targets.backend-unit.repo_digest'
                Expected    = 'reject forbidden TargetRecord field'
            }
            MissingControlField = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'remove control.compose_sha256'
                Expected    = 'reject exact three-field control schema'
            }
            MismatchedTargetControlCommit = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.control_commit unequal to control.commit'
                Expected    = 'reject target-to-control commit binding'
            }
            MismatchedTargetControlBundleSha256 = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.control_bundle_sha256 unequal to control.bundle_sha256'
                Expected    = 'reject target-to-control bundle binding'
            }
            MissingControlCommitLabel = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'remove targets.backend-unit.labels_projection.hub.qa.control-commit'
                Expected    = 'reject missing mandatory control label'
            }
            MismatchedControlBundleLabel = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set targets.backend-unit.labels_projection.hub.qa.control-bundle-sha256 unequal to control.bundle_sha256'
                Expected    = 'reject mismatched mandatory control label'
            }
            MismatchedControlComposeSha256 = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set control.compose_sha256 unequal to canonical raw Compose digest'
                Expected    = 'reject raw Compose control pin'
            }
            V2SchemaRejected = [ordered]@{
                BaseFixture = 'LockBytes.Valid'
                Mutation    = 'set schema to hub.qa.local-images-lock/v2'
                Expected    = 'reject v2 schema before adapter construction'
            }
        }
        LocalImageInspectRecords = [ordered]@{
            BackendUnit    = $backendUnitInspect
            FrontendUnit   = $frontendUnitInspect
            BackendOpenapi = $backendOpenapiInspect
        }
        LocalImageAttestorRecords = [ordered]@{
            AllBasesAndTargets = [pscustomobject][ordered]@{
                RecordType     = 'hub.qa.fixture-local-image-attestor/v2'
                BaseRecords    = @($golangBaseInspect, $playwrightBaseInspect, $pythonBaseInspect)
                TargetRecords  = @($backendUnitInspect, $frontendUnitInspect, $backendOpenapiInspect)
                BaseKeys       = @('golang', 'playwright', 'python')
                TargetKeys     = @('backend-unit', 'frontend-unit', 'backend-openapi')
                Control        = [ordered]@{ Commit = $controlCommit; BundleSha256 = $controlBundleSha256; ComposeSha256 = $controlComposeSha256 }
                TargetBaseMap  = [ordered]@{
                    'backend-unit'    = 'golang'
                    'frontend-unit'   = 'playwright'
                    'backend-openapi' = 'python'
                }
                AttestationLane = 'separate all-bases/all-targets lane; not an adapter operation'
                BaseRepoDigest  = 'inventory only; never a selector'
                TargetRepoDigest = 'forbidden'
                Observation     = 'static declaration only; no attestor is invoked'
            }
        }
        BakedSourceImageSets = [ordered]@{
            AllThree = [pscustomobject][ordered]@{
                RecordType                    = 'hub.qa.fixture-baked-source-image-set/v2'
                Targets                       = @($backendUnitInspect, $frontendUnitInspect, $backendOpenapiInspect)
                RequiredTargets               = @('backend-unit', 'frontend-unit', 'backend-openapi')
                RequiredTargetCount           = 3
                RequiredSourceCommit          = $sourceCommit
                RequiredSourcePolicySha256     = $sourcePolicySha256
                RequiredTreeManifestSha256     = $sourceTreeManifestSha256
                RequiredBuildManifestSha256    = $buildManifestSha256
                RequiredControlCommit           = $controlCommit
                RequiredControlBundleSha256     = $controlBundleSha256
                RequiredControlComposeSha256    = $controlComposeSha256
                SourceAndBuildEqualTopLevel     = $true
                TargetRepoDigestForbidden       = $true
            }
        }
        StagingCases = [ordered]@{
            AtomicCollision = [ordered]@{
                CaseRecordType     = 'hub.qa.fixture-staging-case/v1'
                CreateDisposition  = 'create-new only'
                ExistingEntry      = 'same-owned-staging-name'
                Expected           = 'reject-before-generation'
            }
            ForeignOrPreexisting = [ordered]@{
                CaseRecordType     = 'hub.qa.fixture-staging-case/v1'
                ExistingEntry      = 'foreign-or-unmarked-staging'
                Expected           = 'preserve-and-reject-before-generation'
            }
            FinalCollision = [ordered]@{
                CaseRecordType     = 'hub.qa.fixture-staging-case/v1'
                FinalEntry         = 'concurrent-final-destination'
                Expected           = 'reject-before-final-synchronization'
            }
            NoRecursiveDelete = [ordered]@{
                CaseRecordType     = 'hub.qa.fixture-staging-case/v1'
                ForbiddenOperation = 'recursive staging delete'
                Expected           = 'no-delete-route'
            }
        }
        GitControlCases = [ordered]@{
            CappedControls = [ordered]@{
                CaseRecordType    = 'hub.qa.fixture-git-control-case/v1'
                MaximumGitControls = 3
                GitControls       = @('verify-revision-object', 'detect-tracked-dirty-state', 'compare-source-manifest')
                RevisionObject    = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
                Expected          = 'bounded-local-control-contract'
            }
            InvalidObject = [ordered]@{
                CaseRecordType    = 'hub.qa.fixture-git-control-case/v1'
                MaximumGitControls = 3
                RevisionObject    = 'not-a-revision-object'
                Expected          = 'reject-before-staging'
            }
        }
        SourceManifestCases = [ordered]@{
            DirtyOrRacedDifference = [ordered]@{
                CaseRecordType       = 'hub.qa.fixture-source-manifest-case/v1'
                BeforeSha256         = 'sha256:7777777777777777777777777777777777777777777777777777777777777777'
                AfterSha256          = 'sha256:8888888888888888888888888888888888888888888888888888888888888888'
                Expected             = 'reject-before-final-synchronization'
            }
        }
        ProvenanceCases = [ordered]@{
            Mismatch = [ordered]@{
                CaseRecordType       = 'hub.qa.fixture-local-image-provenance-case/v2'
                LockedTarget         = 'targets.backend-unit'
                SelectedImageId      = 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
                MismatchedFields     = @(
                    'ImageId'
                    'Platform'
                    'RootFsLayers positional order'
                    'Config.Entrypoint'
                    'Config.Cmd'
                    'full Config.Labels versus labels_projection'
                    'EntrypointSha256'
                    'CmdSha256'
                    'LabelsProjectionSha256'
                )
                TargetRepoDigest     = 'absent and forbidden'
                ControlBinding       = 'target control_commit and control_bundle_sha256 equal top-level control; full labels include both control labels'
                SeparateAttestor     = 'all three bases plus all three targets'
                Expected             = 'reject-before-target-execution'
            }
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
            InvalidTarget   = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'sentinel'; Task = 'task-1'; Run = 'run-1' }
            InvalidTask     = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'backend-unit'; Task = 'Bad_ID'; Run = 'run-1' }
            InvalidRun      = [ordered]@{ Contract = 'hub.qa-runner/v1'; Target = 'backend-unit'; Task = 'task-1'; Run = '../escape' }
        }
        OwnerPathCases = [ordered]@{
            ExistingOwnerDirectory = [ordered]@{ Kind = 'existing-owner-directory'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            MismatchedOwnerMarker  = [ordered]@{ Kind = 'mismatched-owner-marker'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            InjectedReparsePoint   = [ordered]@{ Kind = 'injected-reparse-point'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
            ResolvedOutsideRoot    = [ordered]@{ Kind = 'resolved-outside-root'; Expected = 'reject-before-child'; OriginalBytes = 'unchanged' }
        }
        ResourceInventories = [ordered]@{
            Empty = @()
            Mixed = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-container-a'; Kind = 'container'; Name = 'hubqa-task-1-run-1-backend-unit'; Project = 'hubqa-task-1-run-1' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-network-a'; Kind = 'network'; Name = 'hubqa-task-1-run-1_default'; Project = 'hubqa-task-1-run-1' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-container-a'; Kind = 'container'; Name = 'foreign-container-a'; Project = 'foreign-project' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-volume-a'; Kind = 'volume'; Name = 'foreign-volume-a'; Project = 'foreign-project' }
            )
            ForeignOnly = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-container-b'; Kind = 'container'; Name = 'foreign-container-b'; Project = 'foreign-project' },
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'foreign-network-b'; Kind = 'network'; Name = 'foreign-network-b'; Project = 'foreign-project' }
            )
            OwnedResidual = @(
                [pscustomobject][ordered]@{ RecordType = 'hub.qa.resource/v1'; Id = 'owned-volume-residual'; Kind = 'volume'; Name = 'hubqa-task-1-run-1-data'; Project = 'hubqa-task-1-run-1' }
            )
        }
        CleanupOutcomes = [ordered]@{ Success = $successfulCleanupOutcome; Obstructed = $obstructedCleanupOutcome }
        DependencyCases = [ordered]@{
            'TC-G0-001' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-001'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-002' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-002'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-003' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-003'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-003A' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-003A'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-004' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-004'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-005' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-005'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-006' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-006'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-006A' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-006A'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-009' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-009'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = $mixedResourceInventory; SelectedOwnedResources = @($ownedContainer, $ownedNetwork); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-010' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-010'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = $ownedResidualInventory; SelectedOwnedResources = @($ownedResidual); CleanupOutcome = $obstructedCleanupOutcome }
            'TC-G0-011' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-011'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-012' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-012'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-013' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-013'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-014' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-014'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-015' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-015'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-016' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-016'; LocalInspectRecord = $backendOpenapiInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-017' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-017'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
            'TC-G0-018' = [pscustomobject][ordered]@{ CaseRecordType = 'hub.qa.fixture-dependency-case/v1'; CaseId = 'TC-G0-018'; LocalInspectRecord = $backendUnitInspect; InjectedResourceInventory = @(); SelectedOwnedResources = @(); CleanupOutcome = $successfulCleanupOutcome }
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
            'TC-G0-007' = [ordered]@{ Fixtures = @('InvalidContract', 'InvalidTarget', 'InvalidTask', 'InvalidRun', 'Malformed', 'InvalidSchema', 'V2SchemaRejected', 'DuplicateBackendUnit', 'NonStringFrontendUnit', 'TagOnlyBackendUnit', 'MissingBackendOpenapi', 'FourthSentinelImage', 'InvalidProjectionHashes', 'MissingSourceField', 'MismatchedSourceField', 'MissingBuildField', 'MismatchedBuildField', 'MissingBaseLabel', 'MismatchedBaseLabel', 'ForbiddenTargetRepoDigest', 'MissingControlField', 'MismatchedTargetControlCommit', 'MismatchedTargetControlBundleSha256', 'MissingControlCommitLabel', 'MismatchedControlBundleLabel', 'MismatchedControlComposeSha256'); ExpectedExitCode = 64; DependencyCalls = 0; ReportsManifest = 'unchanged'; Stage = 'pre-child-pre-adapter' }
            'TC-G0-008' = [ordered]@{ Fixtures = @('ExistingOwnerDirectory', 'MismatchedOwnerMarker', 'InjectedReparsePoint', 'ResolvedOutsideRoot'); Expected = 'original-bytes-preserved'; Stage = 'pre-child' }
            'TC-G0-009' = [ordered]@{ Fixture = 'Mixed'; SelectedOwnership = 'owned'; SelectedCount = 2; ForeignSelectedCount = 0 }
            'TC-G0-010' = [ordered]@{ Fixture = 'Obstructed'; Primary23Final = 23; Primary0Final = 71 }
            'TC-G0-011' = [ordered]@{ Fixture = 'AtomicCollision'; Expected = 'reject-before-generation' }
            'TC-G0-012' = [ordered]@{ Fixture = 'ForeignOrPreexisting'; Expected = 'preserve-and-reject-before-generation' }
            'TC-G0-013' = [ordered]@{ Fixture = 'FinalCollision'; Expected = 'reject-before-final-synchronization' }
            'TC-G0-014' = [ordered]@{ Fixture = 'NoRecursiveDelete'; Expected = 'no-delete-route' }
            'TC-G0-015' = [ordered]@{ Fixtures = @('CappedControls', 'InvalidObject'); MaximumGitControls = 3; Expected = 'invalid-object-reject-before-staging' }
            'TC-G0-016' = [ordered]@{ Fixture = 'DirtyOrRacedDifference'; Expected = 'reject-before-final-synchronization' }
            'TC-G0-017' = [ordered]@{ Fixtures = @('AllThree', 'AllBasesAndTargets'); RequiredBaseCount = 3; RequiredTargetCount = 3; ExpectedControlCommit = $controlCommit; ExpectedControlBundleSha256 = $controlBundleSha256; ExpectedControlComposeSha256 = $controlComposeSha256; ExpectedSourceCommit = $sourceCommit; ExpectedSourcePolicySha256 = $sourcePolicySha256; ExpectedTreeManifestSha256 = $sourceTreeManifestSha256; ExpectedBuildManifestSha256 = $buildManifestSha256; TargetRepoDigest = 'forbidden' }
            'TC-G0-018' = [ordered]@{ Fixtures = @('Mismatch', 'AllBasesAndTargets'); SelectedInspect = 'exact ImageId plus normalized platform/layers/Config and JCS hashes'; SeparateAttestor = 'all bases and targets'; Expected = 'reject-before-target-execution' }
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

function Get-HubQaPublicLauncherFixtureCatalog {
    [CmdletBinding()]
    param()

    return [pscustomobject][ordered]@{
        SourceCommit = 'cc0f3fcbf4e9a0adc70e946e739b81d21ac7da6c'
        ValidTargets = @(
            'backend-unit',
            'frontend-unit',
            'backend-openapi',
            'python-contract',
            'python-api',
            'python-integration',
            'python-e2e-smoke',
            'python-e2e-full',
            'python-load'
        )
        ValidRequest = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a')
        InvalidArgumentVectors = [ordered]@{
            Missing      = @('-Target', 'backend-unit', '-Task', 'task-2a')
            Duplicate    = @('-Target', 'backend-unit', '-Target', 'frontend-unit', '-Task', 'task-2a', '-Run', 'run-2a')
            Reordered    = @('-Task', 'task-2a', '-Target', 'backend-unit', '-Run', 'run-2a')
            Extra        = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a', '-Extra', 'value')
            Unknown      = @('-Target', 'unknown', '-Task', 'task-2a', '-Run', 'run-2a')
            Private      = @('-Contract', 'hub.qa-runner/v1', '-Task', 'task-2a', '-Run', 'run-2a')
            Contract     = @('-Target', 'backend-unit', '-Task', 'hub.qa-runner/v1', '-Run', 'run-2a')
            Path         = @('-Target', 'backend-unit', '-Task', '..', '-Run', 'run-2a')
            Image        = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a', '-Image', 'sha256-deadbeef')
            Compose      = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a', '-Compose', 'compose.unit.yaml')
            Environment  = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a', '-Environment', 'X=1')
            Timeout      = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a', '-Timeout', '1')
            Terminator   = @('--', 'backend-unit', '-Task', 'task-2a', '-Run', 'run-2a')
            TerminalLf   = @('-Target', 'backend-unit', '-Task', ('task-2a' + [char]10), '-Run', 'run-2a')
            TerminalCrLf = @('-Target', 'backend-unit', '-Task', 'task-2a', '-Run', ('run-2a' + [char]13 + [char]10))
        }
        TestCopyLayout = @(
            'repo/.git',
            'repo/testframework/runner/invoke-qa-gate.ps1',
            'repo/testframework/runner/run-task.ps1',
            'repo/testframework/docker',
            'repo/testframework/reports'
        )
        GitMarkers = [ordered]@{
            Directory = [ordered]@{ Kind = 'directory'; Contents = @() }
            Linked    = [ordered]@{ Kind = 'file'; Contents = 'gitdir: C:/outside/not-dereferenced' }
        }
        ReparseLocations = @('script', 'root', 'runner', 'docker', 'reports', 'report-child')
        EscapingPaths    = @('..', '\\server\share', 'C:relative')
        ReparseCases = [ordered]@{
            Script      = [ordered]@{ Location = 'script'; Invocation = 'public'; ExpectedExitCode = 64 }
            Root        = [ordered]@{ Location = 'root'; Invocation = 'public-and-private-boundary'; ExpectedExitCode = 64 }
            Runner      = [ordered]@{ Location = 'runner'; Invocation = 'public-and-private-boundary'; ExpectedExitCode = 64 }
            Docker      = [ordered]@{ Location = 'docker'; Invocation = 'public'; ExpectedExitCode = 64 }
            Reports     = [ordered]@{ Location = 'reports'; Invocation = 'public'; ExpectedExitCode = 64 }
            ReportChild = [ordered]@{ Location = 'report-child'; Invocation = 'public'; ExpectedExitCode = 64 }
        }
        AncestorReparse = [ordered]@{ Location = 'ancestor-above-repository'; Invocation = 'public'; ExpectedExitCode = 64 }
        ScriptLaunchCases = [ordered]@{
            DotDot        = [ordered]@{ RelativePath = 'runner\..\runner\invoke-qa-gate.ps1'; ExpectedExitCode = 64 }
            RootEscape    = [ordered]@{ RelativePath = 'runner\..\..\outside\invoke-qa-gate.ps1'; ExpectedExitCode = 64 }
            Unc           = [ordered]@{ RelativePath = '\\server\share\invoke-qa-gate.ps1'; ExpectedExitCode = 64 }
            DriveRelative = [ordered]@{ RelativePath = 'C:invoke-qa-gate.ps1'; ExpectedExitCode = 64 }
        }
        PrivateExitCodes = @(0, 23, 124, 65, 70, 71)
        NotReadyRecord   = [ordered]@{
            schema = 'hub.qa-gate/v1'
            target = 'backend-unit'
            task   = 'task-2a'
            run    = 'run-2a'
            status = 'NOT_READY'
        }
    }
}
