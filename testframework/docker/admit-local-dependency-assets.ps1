param(
    [Parameter(Mandatory = $true)][string]$BaseImageId, [Parameter(Mandatory = $true)][string]$CacheRoot,
    [Parameter(Mandatory = $true)][string]$HandoffManifestSha256, [Parameter(Mandatory = $true)][string]$SourceIngressRoot,
    [Parameter(Mandatory = $true)][string]$SourceBlobsManifestSha256, [Parameter(Mandatory = $true)][string]$TaskRunId,
    [Parameter(Mandatory = $true)][string]$OwnerNonce,
    [Parameter(Mandatory = $true)][string]$EvidenceRoot
)

Set-StrictMode -Version Latest

$script:CanonicalProfileBytes = [System.Text.Encoding]::UTF8.GetBytes('{"schema":"hub.qa.local-dependency-admission/v1","platform":{"os":"linux","architecture":"amd64","variant":null},"container_name_template":"hub-qa-local-dependency-admission-{task_run_id}-{owner_nonce_prefix12}","cpus":2,"memory":"2g","pids_limit":256,"timeout_seconds":300,"network_mode":"none","read_only":true,"cap_drop":["ALL"],"no_new_privileges":true,"command":["/opt/hub/admission/admit-local-dependency-assets"],"args":[],"mounts":{"cache":"/cache:ro","ingress":"/ingress:ro","evidence":"/evidence:rw","derived":"/derived:rw"}}')
$script:CanonicalProfilePath = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\.cursor\skills\qa-gate\profiles\local-dependency-admission.v1.json'))

function private:Assert-ExactInputs {
    param(
        [string]$BaseImageId, [string]$CacheRoot, [string]$HandoffManifestSha256,
        [string]$SourceIngressRoot, [string]$SourceBlobsManifestSha256, [string]$TaskRunId, [string]$OwnerNonce,
        [string]$EvidenceRoot
    )

    foreach ($Value in @($BaseImageId, $CacheRoot, $SourceIngressRoot, $TaskRunId, $OwnerNonce, $EvidenceRoot)) {
        if ([string]::IsNullOrWhiteSpace($Value)) { throw 'D0 requires all eight mandatory inputs.' }
    }
    if ($TaskRunId -notmatch '^[a-z0-9][a-z0-9-]{0,62}$') { throw 'TaskRunId must match the frozen D0 grammar.' }
    if ($OwnerNonce -notmatch '^[0-9a-f]{32}$') { throw 'OwnerNonce must match the frozen D0 grammar.' }
    foreach ($ManifestIdentity in @($HandoffManifestSha256, $SourceBlobsManifestSha256)) {
        if ($ManifestIdentity -notmatch '^[0-9a-f]{64}$') { throw 'Manifest identities must match the frozen D0 grammar.' }
    }
}

function private:Assert-ProfileByteEquality {
    $actualBytes = [System.IO.File]::ReadAllBytes($script:CanonicalProfilePath)
    if ($actualBytes.Length -ne $script:CanonicalProfileBytes.Length) { throw 'The local dependency admission profile is not byte-identical to D0.' }
    for ($index = 0; $index -lt $actualBytes.Length; $index++) {
        if ($actualBytes[$index] -ne $script:CanonicalProfileBytes[$index]) { throw 'The local dependency admission profile is not byte-identical to D0.' }
    }
}

function private:Resolve-CanonicalTaskPaths {
    param([string]$CacheRoot, [string]$SourceIngressRoot, [string]$TaskRunId, [string]$EvidenceRoot)

    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    $expectedCacheRoot = 'C:\Users\Valerii\Desktop\hub\.qa-local-cache\g0-v3'
    $expectedEvidenceRoot = 'C:\Users\Valerii\Desktop\hub\testframework\reports\' + $TaskRunId
    $expectedSourceIngressRoot = $expectedEvidenceRoot + '\dependency-admission\inputs'
    $rootSpecifications = @(
        [pscustomobject]@{ Name = 'CacheRoot'; Supplied = $CacheRoot; Expected = $expectedCacheRoot },
        [pscustomobject]@{ Name = 'EvidenceRoot'; Supplied = $EvidenceRoot; Expected = $expectedEvidenceRoot },
        [pscustomobject]@{ Name = 'SourceIngressRoot'; Supplied = $SourceIngressRoot; Expected = $expectedSourceIngressRoot }
    )

    foreach ($rootSpecification in $rootSpecifications) {
        foreach ($character in $rootSpecification.Supplied.ToCharArray()) {
            if ([char]::IsControl($character) -or $character -eq ',' -or $character -eq '=' -or $character -eq '/') {
                throw ($rootSpecification.Name + ' contains a forbidden mount/path character.')
            }
        }
        if ($rootSpecification.Supplied.EndsWith('\', [System.StringComparison]::Ordinal)) {
            throw ($rootSpecification.Name + ' must not have a trailing separator.')
        }
        if ($rootSpecification.Supplied.Length -lt 4 -or
            -not $rootSpecification.Supplied.StartsWith('C:\', $comparison)) {
            throw ($rootSpecification.Name + ' is not an absolute frozen-drive path.')
        }
        foreach ($segment in $rootSpecification.Supplied.Substring(3).Split([char]'\')) {
            if ($segment.Length -eq 0 -or $segment -ceq '.' -or $segment -ceq '..') {
                throw ($rootSpecification.Name + ' contains an unresolved path segment.')
            }
        }
        if (-not [string]::Equals($rootSpecification.Supplied, $rootSpecification.Expected, $comparison)) {
            throw ($rootSpecification.Name + ' does not have the frozen literal D0 spelling.')
        }
        $fullPath = [System.IO.Path]::GetFullPath($rootSpecification.Supplied)
        if (-not [string]::Equals($fullPath, $rootSpecification.Expected, $comparison)) {
            throw ($rootSpecification.Name + ' does not resolve to its frozen D0 path.')
        }
        $currentDirectory = [System.IO.DirectoryInfo]::new($rootSpecification.Expected)
        while ($null -ne $currentDirectory) {
            $currentDirectory.Refresh()
            if (-not $currentDirectory.Exists) {
                throw ($rootSpecification.Name + ' or an ancestor is not an existing directory.')
            }
            $attributes = $currentDirectory.Attributes
            if (($attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
                ($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                throw ($rootSpecification.Name + ' or an ancestor is not a regular non-reparse directory.')
            }
            $canonicalComponent = [System.IO.Path]::GetFullPath($currentDirectory.FullName)
            if (-not [string]::Equals($canonicalComponent, $currentDirectory.FullName, $comparison)) {
                throw ($rootSpecification.Name + ' has a noncanonical ancestor.')
            }
            $currentDirectory = $currentDirectory.Parent
        }
    }

    $dockerClientPath = 'C:\Program Files\Docker\Docker\resources\bin\docker.exe'
    $canonicalDockerClientPath = [System.IO.Path]::GetFullPath($dockerClientPath)
    if (-not [string]::Equals($canonicalDockerClientPath, $dockerClientPath, $comparison)) {
        throw 'The frozen Docker client path is not canonical.'
    }
    $dockerClient = [System.IO.FileInfo]::new($dockerClientPath)
    $dockerClient.Refresh()
    if (-not $dockerClient.Exists) {
        throw 'The frozen Docker client is not an existing regular file.'
    }
    $dockerAttributes = $dockerClient.Attributes
    if (($dockerAttributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
        ($dockerAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
        -not [string]::Equals([System.IO.Path]::GetFullPath($dockerClient.FullName), $dockerClientPath, $comparison)) {
        throw 'The frozen Docker client is not a canonical regular non-reparse file.'
    }
    $dockerAncestor = $dockerClient.Directory
    while ($null -ne $dockerAncestor) {
        $dockerAncestor.Refresh()
        if (-not $dockerAncestor.Exists) {
            throw 'The frozen Docker client has a missing ancestor.'
        }
        $dockerAncestorAttributes = $dockerAncestor.Attributes
        if (($dockerAncestorAttributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
            ($dockerAncestorAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
            -not [string]::Equals([System.IO.Path]::GetFullPath($dockerAncestor.FullName), $dockerAncestor.FullName, $comparison)) {
            throw 'The frozen Docker client has a noncanonical or reparse ancestor.'
        }
        $dockerAncestor = $dockerAncestor.Parent
    }

    $dependencyAdmissionRoot = $expectedEvidenceRoot + '\dependency-admission'
    $sourceBlobsRoot = $expectedSourceIngressRoot + '\source-blobs'
    $finalDerivedRoot = $dependencyAdmissionRoot + '\derived'
    return [pscustomobject]@{
        CacheRoot = $expectedCacheRoot
        EvidenceRoot = $expectedEvidenceRoot
        SourceIngressRoot = $expectedSourceIngressRoot
        DependencyAdmissionRoot = $dependencyAdmissionRoot
        SourceBlobsManifestPath = $expectedSourceIngressRoot + '\source-blobs.manifest'
        SourceBlobsRoot = $sourceBlobsRoot
        DockerClientPath = $dockerClientPath
        FinalRunPath = $dependencyAdmissionRoot + '\run.json'
        FinalAssetsPath = $dependencyAdmissionRoot + '\assets.json'
        FinalCleanupPath = $dependencyAdmissionRoot + '\cleanup.json'
        FinalDerivedRoot = $finalDerivedRoot
        FinalDerivedConfigPath = $finalDerivedRoot + '\vitest.frontend-unit.config.ts'
        FinalDerivationPath = $finalDerivedRoot + '\vitest-derivation.v1.json'
        SourceFilePaths = @(
            $sourceBlobsRoot + '\backend\go.mod',
            $sourceBlobsRoot + '\backend\go.sum',
            $sourceBlobsRoot + '\frontend\package.json',
            $sourceBlobsRoot + '\frontend\package-lock.json',
            $sourceBlobsRoot + '\frontend\vite.config.ts',
            $sourceBlobsRoot + '\frontend\src\test\setup.ts',
            $sourceBlobsRoot + '\frontend\tsconfig.json'
        )
        TaskRunId = $TaskRunId
    }
}

function private:Assert-IngressAndFinalClosure {
    param([object]$TaskPaths)

    if ($null -eq $TaskPaths) {
        throw 'D0 path preflight did not produce task paths.'
    }
    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    foreach ($finalPath in @(
        $TaskPaths.FinalRunPath,
        $TaskPaths.FinalAssetsPath,
        $TaskPaths.FinalCleanupPath,
        $TaskPaths.FinalDerivedRoot,
        $TaskPaths.FinalDerivedConfigPath,
        $TaskPaths.FinalDerivationPath
    )) {
        if ([System.IO.File]::Exists($finalPath) -or [System.IO.Directory]::Exists($finalPath)) {
            throw 'A frozen final dependency-admission path already exists.'
        }
    }

    $directorySpecifications = @(
        [pscustomobject]@{
            Path = $TaskPaths.DependencyAdmissionRoot
            Children = @([pscustomobject]@{ Name = 'inputs'; Kind = 'Directory' })
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceIngressRoot
            Children = @(
                [pscustomobject]@{ Name = 'source-blobs.manifest'; Kind = 'File' },
                [pscustomobject]@{ Name = 'source-blobs'; Kind = 'Directory' }
            )
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceBlobsRoot
            Children = @(
                [pscustomobject]@{ Name = 'backend'; Kind = 'Directory' },
                [pscustomobject]@{ Name = 'frontend'; Kind = 'Directory' }
            )
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceBlobsRoot + '\backend'
            Children = @(
                [pscustomobject]@{ Name = 'go.mod'; Kind = 'File' },
                [pscustomobject]@{ Name = 'go.sum'; Kind = 'File' }
            )
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceBlobsRoot + '\frontend'
            Children = @(
                [pscustomobject]@{ Name = 'package.json'; Kind = 'File' },
                [pscustomobject]@{ Name = 'package-lock.json'; Kind = 'File' },
                [pscustomobject]@{ Name = 'vite.config.ts'; Kind = 'File' },
                [pscustomobject]@{ Name = 'tsconfig.json'; Kind = 'File' },
                [pscustomobject]@{ Name = 'src'; Kind = 'Directory' }
            )
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceBlobsRoot + '\frontend\src'
            Children = @([pscustomobject]@{ Name = 'test'; Kind = 'Directory' })
        },
        [pscustomobject]@{
            Path = $TaskPaths.SourceBlobsRoot + '\frontend\src\test'
            Children = @([pscustomobject]@{ Name = 'setup.ts'; Kind = 'File' })
        }
    )

    foreach ($directorySpecification in $directorySpecifications) {
        $directoryPath = [System.IO.Path]::GetFullPath($directorySpecification.Path)
        $containedPrefix = $TaskPaths.DependencyAdmissionRoot + '\'
        if (-not [string]::Equals($directoryPath, $TaskPaths.DependencyAdmissionRoot, $comparison) -and
            -not $directoryPath.StartsWith($containedPrefix, $comparison)) {
            throw 'An ingress directory escapes dependency-admission containment.'
        }
        $directoryInfo = [System.IO.DirectoryInfo]::new($directoryPath)
        $directoryInfo.Refresh()
        if (-not $directoryInfo.Exists) {
            throw 'An expected ingress directory is missing.'
        }
        $directoryAttributes = $directoryInfo.Attributes
        if (($directoryAttributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
            ($directoryAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
            -not [string]::Equals([System.IO.Path]::GetFullPath($directoryInfo.FullName), $directoryPath, $comparison)) {
            throw 'An ingress directory is noncanonical, non-directory, or reparse.'
        }
        $directoryStreams = @(Get-Item -LiteralPath $directoryPath -Stream * -Force -ErrorAction Stop)
        $directoryDefaultStreamCount = 0
        foreach ($directoryStream in $directoryStreams) {
            if ($directoryStream.Stream -cne ':$DATA') {
                throw 'An ingress directory has an alternate data stream.'
            }
            $directoryDefaultStreamCount++
        }
        if ($directoryDefaultStreamCount -gt 1) {
            throw 'An ingress directory has an unexpected stream projection.'
        }

        $actualEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($directoryPath))
        if ($actualEntries.Count -ne $directorySpecification.Children.Count) {
            throw 'The ingress tree contains a missing or additional child.'
        }
        $matchedNames = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)
        foreach ($actualEntry in $actualEntries) {
            $actualFullPath = [System.IO.Path]::GetFullPath($actualEntry)
            if (-not $actualFullPath.StartsWith($directoryPath + '\', $comparison)) {
                throw 'An ingress child escapes its canonical parent.'
            }
            $matchedChild = $null
            foreach ($expectedChild in $directorySpecification.Children) {
                $expectedChildPath = [System.IO.Path]::Combine($directoryPath, $expectedChild.Name)
                if ([string]::Equals($actualFullPath, $expectedChildPath, $comparison)) {
                    if ($null -ne $matchedChild) {
                        throw 'The ingress tree contains a duplicate child identity.'
                    }
                    $matchedChild = $expectedChild
                }
            }
            if ($null -eq $matchedChild) {
                throw 'The ingress tree contains a non-allowlisted child.'
            }
            if (-not $matchedNames.Add($matchedChild.Name)) {
                throw 'The ingress tree contains a duplicate ordinal child identity.'
            }

            if ($matchedChild.Kind -ceq 'Directory') {
                $childInfo = [System.IO.DirectoryInfo]::new($actualFullPath)
                $childInfo.Refresh()
                if (-not $childInfo.Exists) {
                    throw 'An expected ingress directory child is missing.'
                }
                $childAttributes = $childInfo.Attributes
                if (($childAttributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
                    ($childAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                    throw 'An ingress directory child has the wrong type or is reparse.'
                }
            }
            else {
                $childInfo = [System.IO.FileInfo]::new($actualFullPath)
                $childInfo.Refresh()
                if (-not $childInfo.Exists) {
                    throw 'An expected ingress regular file is missing.'
                }
                $childAttributes = $childInfo.Attributes
                if (($childAttributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
                    ($childAttributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                    throw 'An ingress file child has the wrong type or is reparse.'
                }
            }
            if (-not [string]::Equals([System.IO.Path]::GetFullPath($childInfo.FullName), $actualFullPath, $comparison)) {
                throw 'An ingress child is not canonical.'
            }
            $childStreams = @(Get-Item -LiteralPath $actualFullPath -Stream * -Force -ErrorAction Stop)
            $childDefaultStreamCount = 0
            foreach ($childStream in $childStreams) {
                if ($childStream.Stream -cne ':$DATA') {
                    throw 'An ingress child has an alternate data stream.'
                }
                $childDefaultStreamCount++
            }
            if (($matchedChild.Kind -ceq 'File' -and $childDefaultStreamCount -ne 1) -or
                ($matchedChild.Kind -ceq 'Directory' -and $childDefaultStreamCount -gt 1)) {
                throw 'An ingress child has an unexpected stream projection.'
            }
        }
    }
    foreach ($finalPath in @(
        $TaskPaths.FinalRunPath,
        $TaskPaths.FinalAssetsPath,
        $TaskPaths.FinalCleanupPath,
        $TaskPaths.FinalDerivedRoot,
        $TaskPaths.FinalDerivedConfigPath,
        $TaskPaths.FinalDerivationPath
    )) {
        if ([System.IO.File]::Exists($finalPath) -or [System.IO.Directory]::Exists($finalPath)) {
            throw 'A frozen final dependency-admission path appeared during preflight.'
        }
    }
    $closingChildren = @([System.IO.Directory]::EnumerateFileSystemEntries($TaskPaths.DependencyAdmissionRoot))
    if ($closingChildren.Count -ne 1 -or
        -not [string]::Equals([System.IO.Path]::GetFullPath($closingChildren[0]), $TaskPaths.SourceIngressRoot, $comparison)) {
        throw 'The dependency-admission child closure changed during preflight.'
    }
    $closingIngress = [System.IO.DirectoryInfo]::new($TaskPaths.SourceIngressRoot)
    $closingIngress.Refresh()
    if (-not $closingIngress.Exists -or
        ($closingIngress.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
        ($closingIngress.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw 'The inputs directory identity changed during preflight.'
    }
}

function private:Initialize-NativeDirectoryApi {
    if ($null -ne ([System.Management.Automation.PSTypeName]'HubQaNativeDirectory').Type) {
        throw 'The fixed native directory type already exists and cannot be adopted.'
    }
    $source = @'
internal static class HubQaNativeDirectory {
    [System.Runtime.InteropServices.DllImport(
        "kernel32.dll",
        CharSet = System.Runtime.InteropServices.CharSet.Unicode,
        SetLastError = true)]
    [return: System.Runtime.InteropServices.MarshalAs(
        System.Runtime.InteropServices.UnmanagedType.Bool)]
    internal static extern bool CreateDirectoryW(
        string lpPathName,
        System.IntPtr lpSecurityAttributes);
}
'@
    $nativeTypes = @(Add-Type -TypeDefinition $source -Language CSharp -PassThru -ErrorAction Stop)
    if ($nativeTypes.Count -ne 1 -or $nativeTypes[0].FullName -cne 'HubQaNativeDirectory') {
        throw 'The fixed native directory type was not initialized exactly.'
    }
    return $nativeTypes[0]
}

function private:New-ExclusiveOwnedDirectory {
    param([string]$Path, [string]$ParentPath, [type]$NativeDirectoryType)

    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    $canonicalParent = [System.IO.Path]::GetFullPath($ParentPath)
    $canonicalPath = [System.IO.Path]::GetFullPath($Path)
    if (-not [string]::Equals($canonicalParent, $ParentPath, $comparison) -or
        -not [string]::Equals($canonicalPath, $Path, $comparison) -or
        -not $canonicalPath.StartsWith($canonicalParent + '\', $comparison)) {
        throw 'An owned directory path is noncanonical or escapes its parent.'
    }
    $parent = [System.IO.DirectoryInfo]::new($canonicalParent)
    $parent.Refresh()
    if (-not $parent.Exists -or
        ($parent.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
        ($parent.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw 'An owned directory parent is missing, non-directory, or reparse.'
    }
    $method = $NativeDirectoryType.GetMethod(
        'CreateDirectoryW',
        [System.Reflection.BindingFlags]::Static -bor [System.Reflection.BindingFlags]::NonPublic
    )
    if ($null -eq $method) {
        throw 'The fixed CreateDirectoryW primitive is unavailable.'
    }
    $created = [bool]$method.Invoke($null, [object[]]@($canonicalPath, [System.IntPtr]::Zero))
    if (-not $created) {
        $win32Error = [System.Runtime.InteropServices.Marshal]::GetLastWin32Error()
        if ($win32Error -eq 183) {
            throw 'The owned directory already exists and cannot be adopted.'
        }
        throw ('CreateDirectoryW failed with Win32 error ' + $win32Error + '.')
    }
    $directory = [System.IO.DirectoryInfo]::new($canonicalPath)
    $directory.Refresh()
    if (-not $directory.Exists -or
        ($directory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
        ($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
        -not [string]::Equals([System.IO.Path]::GetFullPath($directory.FullName), $canonicalPath, $comparison) -or
        @([System.IO.Directory]::EnumerateFileSystemEntries($canonicalPath)).Count -ne 0) {
        throw 'The exclusively created directory failed its immediate closure check.'
    }
    $directoryStreams = @(Get-Item -LiteralPath $canonicalPath -Stream * -Force -ErrorAction Stop); foreach ($stream in $directoryStreams) { if ($stream.Stream -cne ':$DATA') { throw 'The exclusively created directory has an alternate stream.' } }; if ($directoryStreams.Count -gt 1) { throw 'The exclusively created directory has an unexpected stream projection.' }
    return $canonicalPath
}

function private:Assert-BaseBinding {
    param([string]$BaseImageId)

    $bindingState = 'UNBOUND'
    $pinnedBaseImageId = $null
    $pinnedRecordSha256 = $null
    if ($bindingState -ceq 'UNBOUND') { return [pscustomobject]@{ ExitCode = 64; Status = 'BLOCKED'; Binding = 'UNBOUND' } }
    if ($bindingState -cne 'BOUND' -or $pinnedBaseImageId -notmatch '^sha256:[0-9a-f]{64}$' -or $pinnedRecordSha256 -notmatch '^[0-9a-f]{64}$' -or $BaseImageId -cne $pinnedBaseImageId) { throw 'The future base binding constants or supplied BaseImageId are not exact.' }
    $recordPath = 'C:\Users\Valerii\Desktop\hub\.superpowers\sdd\2026-08-09-minimal-qa-launcher\task-G0-local-dependency-assets-base-image.md'
    if ([System.IO.Path]::GetFullPath($recordPath) -cne $recordPath) { throw 'The future base identity record path is not canonical.' }
    $ancestor = [System.IO.FileInfo]::new($recordPath).Directory
    while ($null -ne $ancestor) {
        $ancestor.Refresh()
        if (-not $ancestor.Exists -or ($ancestor.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($ancestor.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or [System.IO.Path]::GetFullPath($ancestor.FullName) -cne $ancestor.FullName) { throw 'The future base identity record has an invalid ancestor.' }
        $ancestor = $ancestor.Parent
    }
    $parsed = Read-StrictDuplicateFreeJson -Path $recordPath -AllowCanonicalVariantNull
    if ((Get-ByteSha256 -Bytes $parsed.Bytes) -cne $pinnedRecordSha256) { throw 'The future base identity record bytes do not match the pinned SHA-256.' }
    $root = $parsed.Value
    $platform = $root['platform']
    $command = $root['command']
    $shapes = @([pscustomobject]@{ O=$root; K=@('schema','status','base_image_id','platform','command','local_observed_evidence_sha256') }, [pscustomobject]@{ O=$platform; K=@('os','architecture','variant') }, [pscustomobject]@{ O=$command; K=@('path','mode','source_sha256') })
    foreach ($shape in $shapes) {
        if ($shape.O -isnot [System.Collections.Specialized.OrderedDictionary] -or $shape.O.Keys.Count -ne $shape.K.Count) { throw 'The future base identity record has a wrong object shape.' }
        for ($index=0; $index -lt $shape.K.Count; $index++) {
            if ([string]$shape.O.Keys[$index] -cne $shape.K[$index]) { throw 'The future base identity record has missing, additional, or out-of-order members.' }
        }
    }
    if ($root['schema'] -cne 'hub.qa.local-dependency-admission-base-image/v1' -or $root['status'] -cne 'ACCEPTED' -or $root['base_image_id'] -cne $pinnedBaseImageId -or $platform['os'] -cne 'linux' -or $platform['architecture'] -cne 'amd64' -or $platform['variant'] -ne [System.DBNull]::Value -or $command['path'] -cne '/opt/hub/admission/admit-local-dependency-assets' -or $command['mode'] -cne '0555' -or $command['source_sha256'] -notmatch '^[0-9a-f]{64}$' -or $root['local_observed_evidence_sha256'] -notmatch '^[0-9a-f]{64}$') { throw 'The future base identity record values are not the frozen accepted binding.' }
    $canonicalJson = '{"schema":"hub.qa.local-dependency-admission-base-image/v1","status":"ACCEPTED","base_image_id":"' + $pinnedBaseImageId + '","platform":{"os":"linux","architecture":"amd64","variant":null},"command":{"path":"/opt/hub/admission/admit-local-dependency-assets","mode":"0555","source_sha256":"' + $command['source_sha256'] + '"},"local_observed_evidence_sha256":"' + $root['local_observed_evidence_sha256'] + '"}'
    $canonicalBytes = [System.Text.UTF8Encoding]::new($false, $true).GetBytes($canonicalJson)
    if ($canonicalBytes.Length -ne $parsed.Bytes.Length) { throw 'The future base identity record is not canonical no-BOM/zero-LF JSON.' }
    for ($index=0; $index -lt $canonicalBytes.Length; $index++) {
        if ($canonicalBytes[$index] -ne $parsed.Bytes[$index]) { throw 'The future base identity record bytes are not canonical.' }
    }
    return [pscustomobject]@{ ExitCode = 0; Status = 'STATIC'; Binding = 'BOUND'; BaseImageId = $pinnedBaseImageId; RecordSha256 = $pinnedRecordSha256; BoundPlatform = 'linux/amd64'; CommandPath = $command['path']; CommandSourceSha256 = $command['source_sha256']; LocalObservedEvidenceSha256 = $root['local_observed_evidence_sha256'] }
}

function private:Read-StrictDuplicateFreeJson {
    param([string]$Path, [switch]$AllowCanonicalVariantNull)

    $file = [System.IO.FileInfo]::new([System.IO.Path]::GetFullPath($Path))
    $file.Refresh()
    if (-not $file.Exists -or
        ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
        ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw 'Strict JSON input must be a regular non-reparse file.'
    }
    $streams = @(Get-Item -LiteralPath $file.FullName -Stream * -Force -ErrorAction Stop)
    if ($streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA') {
        throw 'Strict JSON input must have only its default data stream.'
    }
    $bytes = [System.IO.File]::ReadAllBytes($file.FullName)
    if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf) {
        throw 'Strict JSON input must not have a UTF-8 BOM.'
    }
    try {
        $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($bytes)
    }
    catch {
        throw 'Strict JSON input is not well-formed UTF-8.'
    }
    if ($text.Length -eq 0 -or $text[0] -ne [char]0x7b) {
        throw 'Strict JSON input must be a nonempty object.'
    }
    $stack = [System.Collections.ArrayList]::new()
    $rootObject = [System.Collections.Specialized.OrderedDictionary]::new([System.StringComparer]::Ordinal)
    [void]$stack.Add(@{ Object = $rootObject; State = 'KeyOrEnd'; PendingKey = $null; ParentKey = $null })
    $root = $null
    $index = 1
    while ($stack.Count -gt 0) {
        while ($index -lt $text.Length -and
            ($text[$index] -eq [char]0x20 -or $text[$index] -eq [char]0x09 -or
             $text[$index] -eq [char]0x0a -or $text[$index] -eq [char]0x0d)) {
            $index++
        }
        if ($index -ge $text.Length) {
            throw 'Strict JSON input ended before its object was complete.'
        }
        $context = $stack[$stack.Count - 1]
        $state = [string]$context.State
        $parseString = $false
        $completeScalar = $false
        $closeObject = $false
        $tokenValue = $null
        if ($state -ceq 'KeyOrEnd' -or $state -ceq 'Key') {
            if ($state -ceq 'KeyOrEnd' -and $text[$index] -eq [char]0x7d) {
                $index++
                $closeObject = $true
            }
            elseif ($text[$index] -eq [char]0x22) {
                $parseString = $true
            }
            else {
                throw 'Strict JSON object keys are malformed or out of order.'
            }
        }
        elseif ($state -ceq 'Colon') {
            if ($text[$index] -ne [char]0x3a) {
                throw 'Strict JSON object key is not followed by a colon.'
            }
            $index++
            $context.State = 'Value'
            continue
        }
        elseif ($state -ceq 'Value') {
            if ($text[$index] -eq [char]0x7b) {
                $context.State = 'AwaitChild'
                $child = [System.Collections.Specialized.OrderedDictionary]::new([System.StringComparer]::Ordinal)
                [void]$stack.Add(@{ Object = $child; State = 'KeyOrEnd'; PendingKey = $null; ParentKey = $context.PendingKey })
                $index++
                continue
            }
            if ($text[$index] -eq [char]0x22) {
                $parseString = $true
            }
            elseif ($text.Substring($index).StartsWith('true', [System.StringComparison]::Ordinal)) {
                $tokenValue = $true
                $index += 4
                $completeScalar = $true
            }
            elseif ($text.Substring($index).StartsWith('false', [System.StringComparison]::Ordinal)) {
                $tokenValue = $false
                $index += 5
                $completeScalar = $true
            }
            elseif ($AllowCanonicalVariantNull -and $context.PendingKey -ceq 'variant' -and $text.Substring($index).StartsWith('null', [System.StringComparison]::Ordinal)) { $tokenValue = [System.DBNull]::Value; $index += 4; $completeScalar = $true }
            elseif ($text[$index] -ge [char]0x30 -and $text[$index] -le [char]0x39) {
                $numberStart = $index
                if ($text[$index] -eq [char]0x30) {
                    $index++
                    if ($index -lt $text.Length -and $text[$index] -ge [char]0x30 -and $text[$index] -le [char]0x39) {
                        throw 'Strict JSON integers must not have a leading zero.'
                    }
                }
                else {
                    while ($index -lt $text.Length -and $text[$index] -ge [char]0x30 -and $text[$index] -le [char]0x39) {
                        $index++
                    }
                }
                [uint64]$unsignedInteger = 0
                if (-not [uint64]::TryParse(
                    $text.Substring($numberStart, $index - $numberStart),
                    [System.Globalization.NumberStyles]::None,
                    [System.Globalization.CultureInfo]::InvariantCulture,
                    [ref]$unsignedInteger
                )) {
                    throw 'Strict JSON integer is outside the supported unsigned range.'
                }
                $tokenValue = $unsignedInteger
                $completeScalar = $true
            }
            else {
                throw 'Strict JSON values may only be objects, ordinary strings, booleans, or unsigned integers.'
            }
        }
        elseif ($state -ceq 'CommaOrEnd') {
            if ($text[$index] -eq [char]0x2c) {
                $index++
                $context.State = 'Key'
                continue
            }
            if ($text[$index] -eq [char]0x7d) {
                $index++
                $closeObject = $true
            }
            else {
                throw 'Strict JSON object has malformed separation or trailing bytes.'
            }
        }
        else {
            throw 'Strict JSON parser reached an invalid object state.'
        }

        if ($parseString) {
            $index++
            $builder = [System.Text.StringBuilder]::new()
            $scalarCount = 0
            $closedString = $false
            while ($index -lt $text.Length) {
                $character = $text[$index]
                $index++
                if ($character -eq [char]0x22) {
                    $closedString = $true
                    break
                }
                if ($character -eq [char]0x5c) {
                    if ($index -ge $text.Length) {
                        throw 'Strict JSON string has an incomplete escape.'
                    }
                    $escape = $text[$index]
                    $index++
                    if ($escape -eq [char]0x22 -or $escape -eq [char]0x5c -or $escape -eq [char]0x2f) {
                        [void]$builder.Append($escape)
                        $scalarCount++
                        continue
                    }
                    if ($escape -ne [char]0x75 -or $index + 4 -gt $text.Length) {
                        throw 'Strict JSON ordinary strings reject control or malformed escapes.'
                    }
                    $hex = $text.Substring($index, 4)
                    if ($hex -notmatch '^[0-9A-Fa-f]{4}$') {
                        throw 'Strict JSON string has a malformed Unicode escape.'
                    }
                    $codeUnit = [Convert]::ToInt32($hex, 16)
                    $index += 4
                    if ($codeUnit -ge 0xd800 -and $codeUnit -le 0xdbff) {
                        if ($index + 6 -gt $text.Length -or $text[$index] -ne [char]0x5c -or $text[$index + 1] -ne [char]0x75) {
                            throw 'Strict JSON string has an unpaired high surrogate.'
                        }
                        $lowHex = $text.Substring($index + 2, 4)
                        if ($lowHex -notmatch '^[0-9A-Fa-f]{4}$') {
                            throw 'Strict JSON string has a malformed low surrogate.'
                        }
                        $lowCodeUnit = [Convert]::ToInt32($lowHex, 16)
                        if ($lowCodeUnit -lt 0xdc00 -or $lowCodeUnit -gt 0xdfff) {
                            throw 'Strict JSON string has an unpaired high surrogate.'
                        }
                        [void]$builder.Append([char]$codeUnit)
                        [void]$builder.Append([char]$lowCodeUnit)
                        $index += 6
                        $scalarCount++
                        continue
                    }
                    if ($codeUnit -ge 0xdc00 -and $codeUnit -le 0xdfff) {
                        throw 'Strict JSON string has an unpaired low surrogate.'
                    }
                    $decodedCharacter = [char]$codeUnit
                    if ([char]::IsControl($decodedCharacter)) {
                        throw 'Strict JSON ordinary strings reject Unicode control characters.'
                    }
                    [void]$builder.Append($decodedCharacter)
                    $scalarCount++
                    continue
                }
                if ([char]::IsControl($character)) {
                    throw 'Strict JSON ordinary strings reject Unicode control characters.'
                }
                if ([char]::IsHighSurrogate($character)) {
                    if ($index -ge $text.Length -or -not [char]::IsLowSurrogate($text[$index])) {
                        throw 'Strict JSON string has an unpaired high surrogate.'
                    }
                    [void]$builder.Append($character)
                    [void]$builder.Append($text[$index])
                    $index++
                    $scalarCount++
                    continue
                }
                if ([char]::IsLowSurrogate($character)) {
                    throw 'Strict JSON string has an unpaired low surrogate.'
                }
                [void]$builder.Append($character)
                $scalarCount++
            }
            if (-not $closedString -or $scalarCount -lt 1 -or $scalarCount -gt 4096) {
                throw 'Strict JSON ordinary string is unterminated or outside 1..4096 scalar values.'
            }
            $tokenValue = $builder.ToString()
            if (-not $tokenValue.IsNormalized([System.Text.NormalizationForm]::FormC)) {
                throw 'Strict JSON ordinary strings must be NFC.'
            }
            if ($state -ceq 'KeyOrEnd' -or $state -ceq 'Key') {
                if ($context.Object.Contains($tokenValue)) {
                    throw 'Strict JSON object contains a duplicate ordinal key.'
                }
                $context.PendingKey = $tokenValue
                $context.State = 'Colon'
                continue
            }
            $completeScalar = $true
        }
        if ($completeScalar) {
            $context.Object.Add($context.PendingKey, $tokenValue)
            $context.PendingKey = $null
            $context.State = 'CommaOrEnd'
            continue
        }
        if ($closeObject) {
            $completedObject = $context.Object
            $parentKey = $context.ParentKey
            $stack.RemoveAt($stack.Count - 1)
            if ($stack.Count -eq 0) {
                $root = $completedObject
                while ($index -lt $text.Length -and
                    ($text[$index] -eq [char]0x20 -or $text[$index] -eq [char]0x09 -or
                     $text[$index] -eq [char]0x0a -or $text[$index] -eq [char]0x0d)) {
                    $index++
                }
                if ($index -ne $text.Length) {
                    throw 'Strict JSON input has trailing bytes.'
                }
            }
            else {
                $parent = $stack[$stack.Count - 1]
                if ($parent.State -cne 'AwaitChild') {
                    throw 'Strict JSON parser lost its child-object state.'
                }
                $parent.Object.Add($parentKey, $completedObject)
                $parent.PendingKey = $null
                $parent.State = 'CommaOrEnd'
            }
        }
    }
    return [pscustomobject]@{ Bytes = $bytes; Value = $root }
}

function private:Get-ByteSha256 {
    param([byte[]]$Bytes)

    $sha256 = [System.Security.Cryptography.SHA256]::Create()
    try {
        return ([System.BitConverter]::ToString($sha256.ComputeHash($Bytes))).Replace('-', '').ToLowerInvariant()
    }
    finally {
        $sha256.Dispose()
    }
}

function private:New-WorkReservation {
    param([object]$TaskPaths, [string]$OwnerNonce, [object]$Deadline)

    if ($null -eq $Deadline -or $Deadline.EndTicks -isnot [long] -or $Deadline.Frequency -isnot [long] -or $Deadline.Frequency -le 0) { throw 'Work reservation requires the shared monotonic deadline.' }
    $workRelativePath = '.dependency-admission-untrusted-' + $OwnerNonce
    $reservationPath = $TaskPaths.EvidenceRoot + '\.dependency-admission-work-reservation-' + $OwnerNonce + '.json'
    $workRoot = $TaskPaths.EvidenceRoot + '\' + $workRelativePath
    $workEvidenceRoot = $workRoot + '\evidence'
    $workDerivedRoot = $workRoot + '\derived'
    $dockerConfigRoot = $workRoot + '\docker-config'
    $ownerPath = $workRoot + '\.owner.json'
    $reservationJson = '{"schema":"hub.qa.local-dependency-admission-work-reservation/v1","task_run":"' + $TaskPaths.TaskRunId + '","owner_nonce":"' + $OwnerNonce + '","work_relative_path":"' + $workRelativePath + '"}'
    $utf8 = [System.Text.UTF8Encoding]::new($false, $true)
    $reservationBytes = $utf8.GetBytes($reservationJson)
    $reservationSha256 = Get-ByteSha256 -Bytes $reservationBytes
    $ownerJson = '{"schema":"hub.qa.local-dependency-admission-work-owner/v1","task_run":"' + $TaskPaths.TaskRunId + '","owner_nonce":"' + $OwnerNonce + '","reservation_sha256":"' + $reservationSha256 + '"}'
    $plannedOwnerBytes = $utf8.GetBytes($ownerJson)
    try {
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-before-reservation') }
        $reservationStream = [System.IO.File]::Open($reservationPath, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
    }
    catch {
        $code = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }
        return [pscustomobject]@{ ExitCode = $code; FailureKind = $(if ($code -eq 124) { 'DEADLINE' } else { 'RESERVATION_CREATE' }); WorkState = $null }
    }
    $workState = [pscustomobject]@{
        TaskRunId = $TaskPaths.TaskRunId; OwnerNonce = $OwnerNonce; ReservationPath = $reservationPath
        ReservationBytes = $reservationBytes; ReservationSha256 = $reservationSha256; ReservationCreated = $true; ReservationComplete = $false
        WorkRelativePath = $workRelativePath; WorkRoot = $workRoot; WorkCreated = $false
        OwnerPath = $ownerPath; OwnerBytes = $null; OwnerCreated = $false; OwnerComplete = $false
        EvidenceRoot = $workEvidenceRoot; EvidenceCreated = $false; DerivedRoot = $workDerivedRoot; DerivedCreated = $false
        DockerConfigRoot = $dockerConfigRoot; DockerConfigCreated = $false
        AssetsPath = $workEvidenceRoot + '\assets.json'; ConfigPath = $workDerivedRoot + '\vitest.frontend-unit.config.ts'; DerivationPath = $workDerivedRoot + '\vitest-derivation.v1.json'
        InitializationComplete = $false; InitializationExitCode = 0; InitializationReason = ''
    }
    try {
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-reservation-create') }
        $reservationStream.Write($reservationBytes, 0, $reservationBytes.Length)
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-reservation-write') }
        $reservationStream.Flush($true)
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-reservation-flush') }
        $workState.ReservationComplete = $true
    }
    catch {
        $workState.InitializationExitCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }
        $workState.InitializationReason = $_.Exception.Message
    }
    finally {
        $reservationStream.Dispose()
    }
    if ($workState.InitializationExitCode -ne 0) {
        return [pscustomobject]@{ ExitCode = $workState.InitializationExitCode; FailureKind = $(if ($workState.InitializationExitCode -eq 124) { 'DEADLINE' } else { 'RESERVATION_WRITE' }); WorkState = $workState }
    }
    try {
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-before-native-init') }
        $nativeDirectoryType = Initialize-NativeDirectoryApi
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-native-init') }
        New-ExclusiveOwnedDirectory -Path $workRoot -ParentPath $TaskPaths.EvidenceRoot -NativeDirectoryType $nativeDirectoryType | Out-Null
        $workState.WorkCreated = $true
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-work-create') }
        New-ExclusiveOwnedDirectory -Path $workEvidenceRoot -ParentPath $workRoot -NativeDirectoryType $nativeDirectoryType | Out-Null
        $workState.EvidenceCreated = $true
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-evidence-create') }
        New-ExclusiveOwnedDirectory -Path $workDerivedRoot -ParentPath $workRoot -NativeDirectoryType $nativeDirectoryType | Out-Null
        $workState.DerivedCreated = $true
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-derived-create') }
        New-ExclusiveOwnedDirectory -Path $dockerConfigRoot -ParentPath $workRoot -NativeDirectoryType $nativeDirectoryType | Out-Null
        $workState.DockerConfigCreated = $true
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-docker-config-create') }
        $ownerStream = [System.IO.File]::Open($ownerPath, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
        $workState.OwnerBytes = $plannedOwnerBytes
        $workState.OwnerCreated = $true
        try {
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-owner-create') }
            $ownerStream.Write($plannedOwnerBytes, 0, $plannedOwnerBytes.Length)
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-owner-write') }
            $ownerStream.Flush($true)
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-owner-flush') }
            $workState.OwnerComplete = $true
        }
        finally { $ownerStream.Dispose() }
        $ownership = Assert-WorkOwnershipAndTree -WorkState $workState -Phase PreDocker
        if ($ownership.Classification -cne 'OWNED') { throw 'The new work reservation lost ownership during creation.' }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-before-work-return') }
        $workState.InitializationComplete = $true
    }
    catch {
        $workState.InitializationExitCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }
        $workState.InitializationReason = $_.Exception.Message
    }
    return [pscustomobject]@{ ExitCode = $workState.InitializationExitCode; FailureKind = $(if ($workState.InitializationExitCode -eq 124) { 'DEADLINE' } elseif ($workState.InitializationExitCode -eq 70) { 'WORK_INITIALIZATION' } else { '' }); WorkState = $workState }
}

function private:Assert-WorkOwnershipAndTree {
    param([object]$WorkState, [ValidateSet('PreDocker', 'PostDocker')][string]$Phase)

    try {
        $comparison = [System.StringComparison]::OrdinalIgnoreCase
        $taskEvidenceRoot = [System.IO.Path]::GetDirectoryName($WorkState.WorkRoot); $expectedRelative = '.dependency-admission-untrusted-' + $WorkState.OwnerNonce; if ($WorkState.WorkRelativePath -cne $expectedRelative -or -not [string]::Equals($WorkState.WorkRoot, $taskEvidenceRoot + '\' + $expectedRelative, $comparison) -or -not [string]::Equals($WorkState.ReservationPath, $taskEvidenceRoot + '\.dependency-admission-work-reservation-' + $WorkState.OwnerNonce + '.json', $comparison) -or -not [string]::Equals($WorkState.OwnerPath, $WorkState.WorkRoot + '\.owner.json', $comparison) -or -not [string]::Equals($WorkState.EvidenceRoot, $WorkState.WorkRoot + '\evidence', $comparison) -or -not [string]::Equals($WorkState.DerivedRoot, $WorkState.WorkRoot + '\derived', $comparison) -or -not [string]::Equals($WorkState.DockerConfigRoot, $WorkState.WorkRoot + '\docker-config', $comparison) -or -not [string]::Equals($WorkState.AssetsPath, $WorkState.EvidenceRoot + '\assets.json', $comparison) -or -not [string]::Equals($WorkState.ConfigPath, $WorkState.DerivedRoot + '\vitest.frontend-unit.config.ts', $comparison) -or -not [string]::Equals($WorkState.DerivationPath, $WorkState.DerivedRoot + '\vitest-derivation.v1.json', $comparison)) { throw 'The retained work-state path graph changed.' }
        $fileSpecifications = @([pscustomobject]@{ Path = $WorkState.ReservationPath; Bytes = $WorkState.ReservationBytes }, [pscustomobject]@{ Path = $WorkState.OwnerPath; Bytes = $WorkState.OwnerBytes })
        foreach ($fileSpecification in $fileSpecifications) {
            $file = [System.IO.FileInfo]::new([System.IO.Path]::GetFullPath($fileSpecification.Path)); $file.Refresh()
            if (-not $file.Exists -or ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($file.FullName, $fileSpecification.Path, $comparison)) { throw 'A retained ownership file is missing, noncanonical, wrong-type, or reparse.' }
            $streams = @(Get-Item -LiteralPath $file.FullName -Stream * -Force -ErrorAction Stop); if ($streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA') { throw 'A retained ownership file has an alternate stream.' }
            $actualBytes = [System.IO.File]::ReadAllBytes($file.FullName); if ($actualBytes.Length -ne $fileSpecification.Bytes.Length) { throw 'A retained ownership file changed length.' }
            for ($byteIndex = 0; $byteIndex -lt $actualBytes.Length; $byteIndex++) { if ($actualBytes[$byteIndex] -ne $fileSpecification.Bytes[$byteIndex]) { throw 'A retained ownership file changed bytes.' } }
        }
        if ((Get-ByteSha256 -Bytes $WorkState.ReservationBytes) -cne $WorkState.ReservationSha256) { throw 'The retained reservation hash relation changed.' }
        $reservation = (Read-StrictDuplicateFreeJson -Path $WorkState.ReservationPath).Value; $owner = (Read-StrictDuplicateFreeJson -Path $WorkState.OwnerPath).Value
        $reservationKeys = @('schema', 'task_run', 'owner_nonce', 'work_relative_path'); $ownerKeys = @('schema', 'task_run', 'owner_nonce', 'reservation_sha256')
        if ($reservation.Keys.Count -ne 4 -or $owner.Keys.Count -ne 4) { throw 'An ownership record has the wrong member count.' }
        for ($keyIndex = 0; $keyIndex -lt 4; $keyIndex++) { if ([string]$reservation.Keys[$keyIndex] -cne $reservationKeys[$keyIndex] -or [string]$owner.Keys[$keyIndex] -cne $ownerKeys[$keyIndex]) { throw 'An ownership record has missing, additional, or out-of-order members.' } }
        if ($reservation['schema'] -cne 'hub.qa.local-dependency-admission-work-reservation/v1' -or $reservation['task_run'] -cne $WorkState.TaskRunId -or $reservation['owner_nonce'] -cne $WorkState.OwnerNonce -or $reservation['work_relative_path'] -cne $WorkState.WorkRelativePath) { throw 'The retained reservation values changed.' }
        if ($owner['schema'] -cne 'hub.qa.local-dependency-admission-work-owner/v1' -or $owner['task_run'] -cne $WorkState.TaskRunId -or $owner['owner_nonce'] -cne $WorkState.OwnerNonce -or $owner['reservation_sha256'] -cne $WorkState.ReservationSha256) { throw 'The retained owner values changed.' }
        $evidenceEntries = [ordered]@{}; $derivedEntries = [ordered]@{}
        if ($Phase -ceq 'PostDocker') { $evidenceEntries.Add('assets.json', [pscustomobject]@{ Kind = 'File'; Required = $false }); $derivedEntries.Add('vitest.frontend-unit.config.ts', [pscustomobject]@{ Kind = 'File'; Required = $false }); $derivedEntries.Add('vitest-derivation.v1.json', [pscustomobject]@{ Kind = 'File'; Required = $false }) }
        $directorySpecifications = @(
            [pscustomobject]@{ Path = $WorkState.WorkRoot; Entries = [ordered]@{ '.owner.json' = [pscustomobject]@{ Kind = 'File'; Required = $true }; 'evidence' = [pscustomobject]@{ Kind = 'Directory'; Required = $true }; 'derived' = [pscustomobject]@{ Kind = 'Directory'; Required = $true }; 'docker-config' = [pscustomobject]@{ Kind = 'Directory'; Required = $true } } },
            [pscustomobject]@{ Path = $WorkState.EvidenceRoot; Entries = $evidenceEntries }, [pscustomobject]@{ Path = $WorkState.DerivedRoot; Entries = $derivedEntries }, [pscustomobject]@{ Path = $WorkState.DockerConfigRoot; Entries = [ordered]@{} }
        )
        foreach ($directorySpecification in $directorySpecifications) {
            $directory = [System.IO.DirectoryInfo]::new([System.IO.Path]::GetFullPath($directorySpecification.Path)); $directory.Refresh()
            if (-not $directory.Exists -or ($directory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($directory.FullName, $directorySpecification.Path, $comparison)) { throw 'An owned work directory is missing, noncanonical, wrong-type, or reparse.' }
            $directoryStreams = @(Get-Item -LiteralPath $directory.FullName -Stream * -Force -ErrorAction Stop); foreach ($stream in $directoryStreams) { if ($stream.Stream -cne ':$DATA') { throw 'An owned work directory has an alternate stream.' } }; if ($directoryStreams.Count -gt 1) { throw 'An owned work directory has an unexpected stream projection.' }
            $actualEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($directory.FullName)); if ($actualEntries.Count -gt $directorySpecification.Entries.Count) { throw 'An owned work directory has an unexpected descendant.' }
            $matchedNames = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)
            foreach ($actualEntry in $actualEntries) {
                $actualPath = [System.IO.Path]::GetFullPath($actualEntry); $name = [System.IO.Path]::GetFileName($actualPath)
                if (-not $actualPath.StartsWith($directory.FullName + '\', $comparison) -or -not $directorySpecification.Entries.Contains($name) -or -not $matchedNames.Add($name)) { throw 'An owned work child is escaped, unexpected, or duplicate.' }
                $expected = $directorySpecification.Entries[$name]; $child = Get-Item -LiteralPath $actualPath -Force -ErrorAction Stop; $attributes = $child.Attributes
                if (($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals([System.IO.Path]::GetFullPath($child.FullName), $actualPath, $comparison) -or ($expected.Kind -ceq 'Directory' -and ($attributes -band [System.IO.FileAttributes]::Directory) -eq 0) -or ($expected.Kind -ceq 'File' -and ($attributes -band [System.IO.FileAttributes]::Directory) -ne 0)) { throw 'An owned work child has the wrong identity or type.' }
                $childStreams = @(Get-Item -LiteralPath $actualPath -Stream * -Force -ErrorAction Stop); foreach ($stream in $childStreams) { if ($stream.Stream -cne ':$DATA') { throw 'An owned work child has an alternate stream.' } }; if (($expected.Kind -ceq 'File' -and $childStreams.Count -ne 1) -or ($expected.Kind -ceq 'Directory' -and $childStreams.Count -gt 1)) { throw 'An owned work child has an unexpected stream projection.' }
            }
            foreach ($expectedName in $directorySpecification.Entries.Keys) { if ($directorySpecification.Entries[$expectedName].Required -and -not $matchedNames.Contains($expectedName)) { throw 'An owned work directory is missing a required child.' } }
        }
    }
    catch { return [pscustomobject]@{ Classification = 'OWNER_LOSS'; OutputsComplete = $false; Reason = $_.Exception.Message } }
    $outputsComplete = [System.IO.File]::Exists($WorkState.AssetsPath) -and [System.IO.File]::Exists($WorkState.ConfigPath) -and [System.IO.File]::Exists($WorkState.DerivationPath)
    $classification = if ($Phase -ceq 'PostDocker' -and -not $outputsComplete) { 'ORDINARY_EXECUTION_FAILURE' } else { 'OWNED' }
    return [pscustomobject]@{ Classification = $classification; OutputsComplete = $outputsComplete; Reason = '' }
}

function private:Assert-TemporaryAssetsReceipt {
    param([string]$Path)

    $parsed = Read-StrictDuplicateFreeJson -Path $Path; $root = $parsed.Value
    if ($root -isnot [System.Collections.Specialized.OrderedDictionary]) { throw 'Temporary assets root must be an ordered object.' }
    $admission = $root['admission_base']; $assets = $root['assets']; $aggregate = $root['aggregate']; if ($admission -isnot [System.Collections.Specialized.OrderedDictionary] -or $assets -isnot [System.Collections.Specialized.OrderedDictionary] -or $aggregate -isnot [System.Collections.Specialized.OrderedDictionary]) { throw 'Temporary assets top-level objects have the wrong type.' }
    $go = $assets['go-module-cache']; $race = $assets['race-c-toolchain']; $frontend = $assets['frontend-node-modules']; $vitest = $assets['vitest-config']; $swag = $assets['swag']
    $executables = $race['executables']; $binShims = $frontend['bin_shims']; $frontendVersions = $frontend['versions']; $sourceHashes = $vitest['source_blob_sha1s']; $vitestVersions = $vitest['versions']
    $shapes = @([pscustomobject]@{ O=$root; K=@('schema','admission_base','assets','aggregate','accepted') }, [pscustomobject]@{ O=$admission; K=@('supplied_image_id','observed_image_id','observed_platform','image_id_match','platform_match') }, [pscustomobject]@{ O=$assets; K=@('go-module-cache','race-c-toolchain','frontend-node-modules','vitest-config','swag') }, [pscustomobject]@{ O=$go; K=@('archive_sha256','tree_manifest_sha256','go_version','os','architecture','go_sum_blob_sha1','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$race; K=@('archive_sha256','tree_manifest_sha256','compiler_identity','target_triple','abi','executables','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$executables; K=@('cc','gcc','ld') }, [pscustomobject]@{ O=$frontend; K=@('node_modules_sha256','npm_cache_sha256','tree_manifest_sha256','package_lock_projection_sha256','package_lock_blob_sha1','bin_shims','versions','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$binShims; K=@('biome','tsc','vite','vitest') }, [pscustomobject]@{ O=$frontendVersions; K=@('vitest','vite','typescript') }, [pscustomobject]@{ O=$vitest; K=@('derivation_manifest_sha256','config_sha256','source_blob_sha1s','versions','derivation_reproducible','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$sourceHashes; K=@('backend/go.mod','backend/go.sum','frontend/package.json','frontend/package-lock.json','frontend/vite.config.ts','frontend/src/test/setup.ts','frontend/tsconfig.json') }, [pscustomobject]@{ O=$vitestVersions; K=@('vitest','vite','typescript') }, [pscustomobject]@{ O=$swag; K=@('binary_sha256','tree_manifest_sha256','version','go_sum_blob_sha1','executable_mode','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$aggregate; K=@('observed_bytes','max_bytes','cap_valid') })
    foreach ($shape in $shapes) { if ($shape.O -isnot [System.Collections.Specialized.OrderedDictionary] -or $shape.O.Keys.Count -ne $shape.K.Count) { throw 'Temporary assets contains a wrong-type object or member count.' }; for ($i=0; $i -lt $shape.K.Count; $i++) { if ([string]$shape.O.Keys[$i] -cne $shape.K[$i]) { throw 'Temporary assets contains missing, additional, or out-of-order members.' } } }
    $strings = @([pscustomobject]@{ O=$root; F=@('schema') }, [pscustomobject]@{ O=$admission; F=@('supplied_image_id','observed_image_id','observed_platform') }, [pscustomobject]@{ O=$go; F=@('archive_sha256','tree_manifest_sha256','go_version','os','architecture','go_sum_blob_sha1') }, [pscustomobject]@{ O=$race; F=@('archive_sha256','tree_manifest_sha256','compiler_identity','target_triple','abi') }, [pscustomobject]@{ O=$executables; F=@('cc','gcc','ld') }, [pscustomobject]@{ O=$frontend; F=@('node_modules_sha256','npm_cache_sha256','tree_manifest_sha256','package_lock_projection_sha256','package_lock_blob_sha1') }, [pscustomobject]@{ O=$binShims; F=@('biome','tsc','vite','vitest') }, [pscustomobject]@{ O=$frontendVersions; F=@('vitest','vite','typescript') }, [pscustomobject]@{ O=$vitest; F=@('derivation_manifest_sha256','config_sha256') }, [pscustomobject]@{ O=$sourceHashes; F=$shapes[10].K }, [pscustomobject]@{ O=$vitestVersions; F=@('vitest','vite','typescript') }, [pscustomobject]@{ O=$swag; F=@('binary_sha256','tree_manifest_sha256','version','go_sum_blob_sha1','executable_mode') })
    foreach ($stringSet in $strings) { foreach ($field in $stringSet.F) { if ($stringSet.O[$field] -isnot [string]) { throw 'Temporary assets ordinary-string field has the wrong type.' } } }
    $sha256Sets = @([pscustomobject]@{ O=$go; F=@('archive_sha256','tree_manifest_sha256') }, [pscustomobject]@{ O=$race; F=@('archive_sha256','tree_manifest_sha256') }, [pscustomobject]@{ O=$frontend; F=@('node_modules_sha256','npm_cache_sha256','tree_manifest_sha256','package_lock_projection_sha256') }, [pscustomobject]@{ O=$vitest; F=@('derivation_manifest_sha256','config_sha256') }, [pscustomobject]@{ O=$swag; F=@('binary_sha256','tree_manifest_sha256') }); foreach ($set in $sha256Sets) { foreach ($field in $set.F) { if ($set.O[$field] -notmatch '^[0-9a-f]{64}$') { throw 'Temporary assets SHA-256 field is malformed.' } } }
    $sha1Sets = @([pscustomobject]@{ O=$go; F=@('go_sum_blob_sha1') }, [pscustomobject]@{ O=$frontend; F=@('package_lock_blob_sha1') }, [pscustomobject]@{ O=$sourceHashes; F=$shapes[10].K }, [pscustomobject]@{ O=$swag; F=@('go_sum_blob_sha1') }); foreach ($set in $sha1Sets) { foreach ($field in $set.F) { if ($set.O[$field] -notmatch '^[0-9a-f]{40}$') { throw 'Temporary assets Git SHA-1 field is malformed.' } } }
    $booleans = @([pscustomobject]@{ O=$root; F=@('accepted') }, [pscustomobject]@{ O=$admission; F=@('image_id_match','platform_match') }, [pscustomobject]@{ O=$go; F=@('records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$race; F=@('records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$frontend; F=@('records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$vitest; F=@('derivation_reproducible','records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$swag; F=@('records_valid','cap_valid','accepted') }, [pscustomobject]@{ O=$aggregate; F=@('cap_valid') }); foreach ($set in $booleans) { foreach ($field in $set.F) { if ($set.O[$field] -isnot [bool]) { throw 'Temporary assets boolean field has the wrong type.' } } }
    if ($root['schema'] -cne 'hub.qa.local-dependency-assets/v1' -or $aggregate['observed_bytes'] -isnot [uint64] -or $aggregate['max_bytes'] -isnot [uint64] -or $aggregate['observed_bytes'] -gt 5368709120 -or $aggregate['max_bytes'] -ne 5368709120) { throw 'Temporary assets schema or aggregate integers are invalid.' }
    return $parsed
}

function private:Assert-TemporaryDerivationReceipt {
    param([string]$Path, [string]$ConfigPath)

    $parsed = Read-StrictDuplicateFreeJson -Path $Path; $root = $parsed.Value; $sourceHashes = $root['source_blob_sha1s']; $versions = $root['versions']
    $sourceKeys = @('backend/go.mod','backend/go.sum','frontend/package.json','frontend/package-lock.json','frontend/vite.config.ts','frontend/src/test/setup.ts','frontend/tsconfig.json'); $versionKeys = @('vitest','vite','typescript')
    $shapes = @([pscustomobject]@{ O=$root; K=@('schema','source_blobs_manifest_sha256','source_blob_sha1s','versions','config_sha256','derivation_reproducible') }, [pscustomobject]@{ O=$sourceHashes; K=$sourceKeys }, [pscustomobject]@{ O=$versions; K=$versionKeys })
    foreach ($shape in $shapes) { if ($shape.O -isnot [System.Collections.Specialized.OrderedDictionary] -or $shape.O.Keys.Count -ne $shape.K.Count) { throw 'Temporary derivation contains a wrong-type object or member count.' }; for ($i=0; $i -lt $shape.K.Count; $i++) { if ([string]$shape.O.Keys[$i] -cne $shape.K[$i]) { throw 'Temporary derivation contains missing, additional, or out-of-order members.' } } }
    if ($root['schema'] -cne 'hub.qa.vitest-derivation/v1' -or $root['source_blobs_manifest_sha256'] -notmatch '^[0-9a-f]{64}$' -or $root['config_sha256'] -notmatch '^[0-9a-f]{64}$' -or $root['derivation_reproducible'] -isnot [bool]) { throw 'Temporary derivation scalar fields are invalid.' }
    foreach ($key in $sourceKeys) { if ($sourceHashes[$key] -isnot [string] -or $sourceHashes[$key] -notmatch '^[0-9a-f]{40}$') { throw 'Temporary derivation source hash is invalid.' } }; foreach ($key in $versionKeys) { if ($versions[$key] -isnot [string]) { throw 'Temporary derivation version is invalid.' } }
    $config = [System.IO.FileInfo]::new([System.IO.Path]::GetFullPath($ConfigPath)); $config.Refresh(); if (-not $config.Exists -or $config.Length -eq 0 -or ($config.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($config.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($config.FullName, $ConfigPath, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'Temporary derived config is missing, empty, noncanonical, wrong-type, or reparse.' }
    $streams = @(Get-Item -LiteralPath $config.FullName -Stream * -Force -ErrorAction Stop); if ($streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA') { throw 'Temporary derived config has an alternate stream.' }
    $configBytes = [System.IO.File]::ReadAllBytes($config.FullName); $configSha256 = Get-ByteSha256 -Bytes $configBytes; if ($configSha256 -cne $root['config_sha256']) { throw 'Temporary derived config does not match config_sha256.' }
    return [pscustomobject]@{ DerivationBytes = $parsed.Bytes; ConfigBytes = $configBytes; Value = $root; ConfigSha256 = $configSha256 }
}

function private:Quote-WindowsArgument {
    param([AllowEmptyString()][string]$Argument)

    if ($null -eq $Argument) { throw 'A Docker argv element cannot be null.' }
    if ($Argument.Length -eq 0) { return '""' }
    if ($Argument.IndexOf([char]0x20) -lt 0 -and $Argument.IndexOf([char]0x09) -lt 0 -and $Argument.IndexOf([char]0x22) -lt 0) { return $Argument }
    $builder = [System.Text.StringBuilder]::new(); [void]$builder.Append([char]0x22); $backslashes = 0
    foreach ($character in $Argument.ToCharArray()) {
        if ($character -eq [char]0x5c) { $backslashes++; continue }
        if ($character -eq [char]0x22) { [void]$builder.Append([char]0x5c, (2 * $backslashes) + 1); [void]$builder.Append($character); $backslashes = 0; continue }
        if ($backslashes -gt 0) { [void]$builder.Append([char]0x5c, $backslashes); $backslashes = 0 }
        [void]$builder.Append($character)
    }
    if ($backslashes -gt 0) { [void]$builder.Append([char]0x5c, 2 * $backslashes) }
    [void]$builder.Append([char]0x22); return $builder.ToString()
}

function private:Invoke-DockerPhysicalProcess {
    param([object]$TaskPaths, [object]$WorkState, [object]$Binding, [ValidateSet('create', 'start', 'wait', 'inspect', 'rm')][string]$Operation, [AllowNull()][string]$HeldId, [object]$Deadline)

    $emptyBytes = [byte[]]@()
    $argv = [string[]]@()
    $ownership = $null
    try {
        if ($null -eq $Deadline -or $Deadline.EndTicks -isnot [long] -or
            $Deadline.Frequency -isnot [long] -or $Deadline.Frequency -le 0) {
            throw 'The shared monotonic deadline is missing or malformed.'
        }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
            return [pscustomobject]@{ Operation = $Operation; Argv = $argv; CliExitCode = $null; StdoutBytes = $emptyBytes; StderrBytes = $emptyBytes; FailureExitCode = 124; FailureKind = 'DEADLINE'; ProcessStarted = $false; Ownership = $ownership }
        }
        Assert-IngressAndFinalClosure -TaskPaths $TaskPaths
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { return [pscustomobject]@{ Operation = $Operation; Argv = $argv; CliExitCode = $null; StdoutBytes = $emptyBytes; StderrBytes = $emptyBytes; FailureExitCode = 124; FailureKind = 'DEADLINE'; ProcessStarted = $false; Ownership = $ownership } }
        $phase = if ($Operation -ceq 'create' -or $Operation -ceq 'start') { 'PreDocker' } else { 'PostDocker' }
        $ownership = Assert-WorkOwnershipAndTree -WorkState $WorkState -Phase $phase
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
            return [pscustomobject]@{ Operation = $Operation; Argv = $argv; CliExitCode = $null; StdoutBytes = $emptyBytes; StderrBytes = $emptyBytes; FailureExitCode = 124; FailureKind = 'DEADLINE'; ProcessStarted = $false; Ownership = $ownership }
        }
        if ($ownership.Classification -ceq 'OWNER_LOSS') {
            throw ('Docker preflight rejected owner loss: ' + $ownership.Reason)
        }
        if ($Binding.Binding -cne 'BOUND' -or $Binding.BaseImageId -notmatch '^sha256:[0-9a-f]{64}$' -or
            $Binding.BoundPlatform -cne 'linux/amd64') {
            throw 'Docker execution requires the exact accepted BOUND identity.'
        }
        if ($Operation -ceq 'create') {
            if (-not [string]::IsNullOrEmpty($HeldId)) { throw 'Create cannot consume a held container ID.' }
        }
        elseif ($HeldId -notmatch '^[0-9a-f]{64}$') {
            throw 'A non-create Docker operation requires the retained exact container ID.'
        }
        $resolvedName = 'hub-qa-local-dependency-admission-' + $TaskPaths.TaskRunId + '-' + $WorkState.OwnerNonce.Substring(0, 12)
        $inspectFormat = '{{printf "%s\t%s\t%s\t%t\n" .Id .Image .HostConfig.NetworkMode .HostConfig.ReadonlyRootfs}}{{range .Mounts}}{{printf "%s\t%s\t%s\t%t\n" .Type .Source .Destination .RW}}{{end}}'
        if ($Operation -ceq 'create') {
            $fixedOperationArgv = @(
                'create', '--name', $resolvedName, '--platform=linux/amd64', '--pull=never',
                '--cpus', '2', '--memory', '2g', '--pids-limit', '256', '--network', 'none',
                '--read-only', '--cap-drop', 'ALL', '--security-opt', 'no-new-privileges:true',
                '--label', ('hub.qa.task-run=' + $TaskPaths.TaskRunId),
                '--label', ('hub.qa.owner-nonce=' + $WorkState.OwnerNonce),
                '--mount', ('type=bind,src=' + $TaskPaths.CacheRoot + ',dst=/cache,readonly'),
                '--mount', ('type=bind,src=' + $TaskPaths.SourceIngressRoot + ',dst=/ingress,readonly'),
                '--mount', ('type=bind,src=' + $WorkState.EvidenceRoot + ',dst=/evidence'),
                '--mount', ('type=bind,src=' + $WorkState.DerivedRoot + ',dst=/derived'),
                $Binding.BaseImageId
            )
        }
        elseif ($Operation -ceq 'start') { $fixedOperationArgv = @('start', $HeldId) }
        elseif ($Operation -ceq 'wait') { $fixedOperationArgv = @('wait', $HeldId) }
        elseif ($Operation -ceq 'inspect') { $fixedOperationArgv = @('container', 'inspect', '--format', $inspectFormat, $HeldId) }
        else { $fixedOperationArgv = @('rm', '--force', $HeldId) }
        $argv = [string[]](@('--config', $WorkState.DockerConfigRoot, '--host', 'npipe:////./pipe/docker_engine') + $fixedOperationArgv)
        foreach ($argument in $argv) {
            if ($null -eq $argument) { throw 'A fixed Docker argv element is null.' }
            foreach ($character in $argument.ToCharArray()) {
                if ([char]::IsControl($character)) { throw 'A fixed Docker argv element contains a control character.' }
            }
        }
        if ($TaskPaths.DockerClientPath -cne 'C:\Program Files\Docker\Docker\resources\bin\docker.exe' -or
            $TaskPaths.EvidenceRoot -cne [System.IO.Path]::GetFullPath($TaskPaths.EvidenceRoot)) {
            throw 'Docker execution requires the frozen client and canonical working directory.'
        }
        $client = [System.IO.FileInfo]::new($TaskPaths.DockerClientPath)
        $client.Refresh()
        if (-not $client.Exists -or ($client.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
            ($client.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
            $client.FullName -cne $TaskPaths.DockerClientPath) {
            throw 'The frozen Docker client identity changed before execution.'
        }
        $clientAncestor = $client.Directory
        while ($null -ne $clientAncestor) {
            $clientAncestor.Refresh()
            if (-not $clientAncestor.Exists -or ($clientAncestor.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
                ($clientAncestor.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
                [System.IO.Path]::GetFullPath($clientAncestor.FullName) -cne $clientAncestor.FullName) {
                throw 'A frozen Docker client ancestor changed before execution.'
            }
            $clientAncestor = $clientAncestor.Parent
        }
        $encodedArguments = foreach ($argument in $argv) { Quote-WindowsArgument -Argument $argument }
        $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
        $startInfo.UseShellExecute = $false
        $startInfo.FileName = $TaskPaths.DockerClientPath
        $startInfo.WorkingDirectory = $TaskPaths.EvidenceRoot
        $startInfo.Arguments = $encodedArguments -join ' '
        $startInfo.CreateNoWindow = $true
        $startInfo.LoadUserProfile = $false
        $startInfo.RedirectStandardOutput = $true
        $startInfo.RedirectStandardError = $true
        $startInfo.RedirectStandardInput = $false
        $startInfo.EnvironmentVariables.Clear()
        if ($startInfo.EnvironmentVariables.Count -ne 0) { throw 'The Docker process environment block is not empty.' }
    }
    catch {
        $code = if ($null -ne $Deadline -and $Deadline.EndTicks -is [long] -and [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }
        $kind = if ($code -eq 124) { 'DEADLINE' } else { 'PREFLIGHT' }
        return [pscustomobject]@{ Operation = $Operation; Argv = $argv; CliExitCode = $null; StdoutBytes = $emptyBytes; StderrBytes = $emptyBytes; FailureExitCode = $code; FailureKind = $kind; ProcessStarted = $false; Ownership = $ownership }
    }

    $cap = if ($Operation -ceq 'inspect') { 65536 } else { 4096 }
    $process = [System.Diagnostics.Process]::new()
    $stdoutMemory = [System.IO.MemoryStream]::new()
    $stderrMemory = [System.IO.MemoryStream]::new()
    $started = $false
    $failureKind = ''
    $cliExitCode = $null
    $stdoutTask = $null
    $stderrTask = $null
    $stdoutReader = $null
    $stderrReader = $null
    $stdoutStream = $null
    $stderrStream = $null
    $stdoutBytes = $emptyBytes
    $stderrBytes = $emptyBytes
    try {
        $process.StartInfo = $startInfo
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-before-process-start') }
        if (-not $process.Start()) { $failureKind = 'PROCESS_START' }
        else {
            $started = $true
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-process-start') }
            $stdoutReader = $process.StandardOutput
            $stderrReader = $process.StandardError
            $stdoutStream = $stdoutReader.BaseStream
            $stderrStream = $stderrReader.BaseStream
            $stdoutEof = $false
            $stderrEof = $false
            while ($failureKind.Length -eq 0) {
                $remainingTicks = $Deadline.EndTicks - [System.Diagnostics.Stopwatch]::GetTimestamp()
                if ($remainingTicks -le 0) { throw [System.TimeoutException]::new('deadline-before-process-observation') }
                $processExited = $process.HasExited
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-process-observation') }
                if ($processExited -and $stdoutEof -and $stderrEof) { break }
                if (-not $stdoutEof -and $null -eq $stdoutTask) {
                    $stdoutBuffer = [byte[]]::new([Math]::Min(4096, [int]($cap - $stdoutMemory.Length + 1)))
                    $stdoutTask = $stdoutStream.ReadAsync($stdoutBuffer, 0, $stdoutBuffer.Length)
                }
                if (-not $stderrEof -and $null -eq $stderrTask) {
                    $stderrBuffer = [byte[]]::new([Math]::Min(4096, [int]($cap - $stderrMemory.Length + 1)))
                    $stderrTask = $stderrStream.ReadAsync($stderrBuffer, 0, $stderrBuffer.Length)
                }
                $pending = [System.Collections.Generic.List[System.Threading.Tasks.Task]]::new()
                if ($null -ne $stdoutTask) { $pending.Add($stdoutTask) }
                if ($null -ne $stderrTask) { $pending.Add($stderrTask) }
                $remainingTicks = $Deadline.EndTicks - [System.Diagnostics.Stopwatch]::GetTimestamp()
                if ($remainingTicks -le 0) { throw [System.TimeoutException]::new('deadline-before-stream-wait') }
                $remainingMilliseconds = [Math]::Floor(($remainingTicks * 1000.0) / $Deadline.Frequency)
                if ($remainingMilliseconds -lt 1) { throw [System.TimeoutException]::new('deadline-before-stream-wait') }
                $waitMilliseconds = [int][Math]::Min(50, $remainingMilliseconds)
                if ($pending.Count -eq 0) {
                    [void]$process.WaitForExit($waitMilliseconds)
                    continue
                }
                $completedIndex = [System.Threading.Tasks.Task]::WaitAny($pending.ToArray(), $waitMilliseconds)
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-stream-wait') }
                if ($completedIndex -lt 0) { continue }
                $completedTask = $pending[$completedIndex]
                if ([object]::ReferenceEquals($completedTask, $stdoutTask)) {
                    $count = $stdoutTask.GetAwaiter().GetResult()
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-stdout-read') }
                    $stdoutTask.Dispose()
                    $stdoutTask = $null
                    if ($count -eq 0) { $stdoutEof = $true }
                    elseif ($stdoutMemory.Length + $count -gt $cap) { $failureKind = 'STDOUT_CAP' }
                    else { $stdoutMemory.Write($stdoutBuffer, 0, $count) }
                }
                else {
                    $count = $stderrTask.GetAwaiter().GetResult()
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-stderr-read') }
                    $stderrTask.Dispose()
                    $stderrTask = $null
                    if ($count -eq 0) { $stderrEof = $true }
                    elseif ($stderrMemory.Length + $count -gt $cap) { $failureKind = 'STDERR_CAP' }
                    else { $stderrMemory.Write($stderrBuffer, 0, $count) }
                }
            }
            if ($failureKind.Length -eq 0) {
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-before-exit-code') }
                $processExited = $process.HasExited
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline-after-process-exit-check') }
                if (-not $processExited) { $failureKind = 'PROCESS_STATE' }
                else {
                    $cliExitCode = $process.ExitCode
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
                        $cliExitCode = $null
                        throw [System.TimeoutException]::new('deadline-after-exit-code')
                    }
                }
            }
        }
    }
    catch {
        if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { $failureKind = 'DEADLINE' }
        elseif ($failureKind.Length -eq 0) { $failureKind = 'PROCESS_IO' }
    }
    finally {
        if ($started -and $failureKind.Length -ne 0) {
            try { $process.Kill() }
            catch { }
            if ($failureKind -cne 'DEADLINE') {
                $remainingTicks = $Deadline.EndTicks - [System.Diagnostics.Stopwatch]::GetTimestamp()
                $remainingMilliseconds = [Math]::Floor(($remainingTicks * 1000.0) / $Deadline.Frequency)
                if ($remainingMilliseconds -lt 1) { $failureKind = 'DEADLINE' }
                else {
                    $terminated = $process.WaitForExit([int][Math]::Min([int]::MaxValue, $remainingMilliseconds))
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks -or -not $terminated) { $failureKind = 'DEADLINE' }
                }
            }
            if ($failureKind -cne 'DEADLINE') {
                $pending = [System.Collections.Generic.List[System.Threading.Tasks.Task]]::new()
                if ($null -ne $stdoutTask) { $pending.Add($stdoutTask) }
                if ($null -ne $stderrTask) { $pending.Add($stderrTask) }
                if ($pending.Count -gt 0) {
                    $remainingTicks = $Deadline.EndTicks - [System.Diagnostics.Stopwatch]::GetTimestamp()
                    $remainingMilliseconds = [Math]::Floor(($remainingTicks * 1000.0) / $Deadline.Frequency)
                    if ($remainingMilliseconds -lt 1) { $failureKind = 'DEADLINE' }
                    else {
                        $tasksCompleted = [System.Threading.Tasks.Task]::WaitAll($pending.ToArray(), [int][Math]::Min([int]::MaxValue, $remainingMilliseconds))
                        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks -or -not $tasksCompleted) { $failureKind = 'DEADLINE' }
                    }
                }
            }
        }
        if ($failureKind -cne 'DEADLINE') {
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { $failureKind = 'DEADLINE' }
            else {
                $stdoutBytes = $stdoutMemory.ToArray()
                $stderrBytes = $stderrMemory.ToArray()
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
                    $stdoutBytes = $emptyBytes
                    $stderrBytes = $emptyBytes
                    $failureKind = 'DEADLINE'
                }
            }
        }
        if ($failureKind -cne 'DEADLINE') {
            if ($null -ne $stdoutTask) { $stdoutTask.Dispose() }
            if ($null -ne $stderrTask) { $stderrTask.Dispose() }
        }
        if ($null -ne $stdoutStream) { $stdoutStream.Dispose() }
        if ($null -ne $stderrStream) { $stderrStream.Dispose() }
        if ($null -ne $stdoutReader) { $stdoutReader.Dispose() }
        if ($null -ne $stderrReader) { $stderrReader.Dispose() }
        $stdoutMemory.Dispose()
        $stderrMemory.Dispose()
        $process.Dispose()
    }
    $failureCode = if ($failureKind.Length -eq 0) { 0 } elseif ($failureKind -ceq 'DEADLINE') { 124 } else { 70 }
    return [pscustomobject]@{ Operation = $Operation; Argv = $argv; CliExitCode = $cliExitCode; StdoutBytes = $stdoutBytes; StderrBytes = $stderrBytes; FailureExitCode = $failureCode; FailureKind = $failureKind; ProcessStarted = $started; Ownership = $ownership }
}

function private:Assert-IdentityCommandResult {
    param(
        [object]$Result,
        [ValidateSet('create', 'start', 'rm')][string]$Operation,
        [AllowNull()][string]$HeldId
    )

    if ($null -eq $Result -or $Result.Operation -cne $Operation) {
        throw 'The Docker identity result does not belong to the expected operation.'
    }
    if ($Result.FailureExitCode -ne 0 -or $Result.CliExitCode -ne 0) {
        throw ('The Docker ' + $Operation + ' process did not complete successfully.')
    }
    if ($null -eq $Result.StdoutBytes -or $null -eq $Result.StderrBytes -or $Result.StderrBytes.Length -ne 0) {
        throw ('The Docker ' + $Operation + ' result has a missing stream or nonempty stderr.')
    }
    if ($Operation -ceq 'create') {
        if (-not [string]::IsNullOrEmpty($HeldId)) { throw 'Create result validation cannot consume a held ID.' }
        if ($Result.StdoutBytes.Length -ne 65 -or $Result.StdoutBytes[64] -ne 10) {
            throw 'Docker create stdout is not exact 64-hex plus LF.'
        }
        for ($index = 0; $index -lt 64; $index++) {
            $value = $Result.StdoutBytes[$index]
            if (-not (($value -ge 48 -and $value -le 57) -or ($value -ge 97 -and $value -le 102))) {
                throw 'Docker create stdout contains a non-lowercase-hex byte.'
            }
        }
        $createdId = [System.Text.Encoding]::ASCII.GetString($Result.StdoutBytes, 0, 64)
        if ($createdId -notmatch '^[0-9a-f]{64}$') { throw 'Docker create did not yield an exact retained ID.' }
        return $createdId
    }
    if ($HeldId -notmatch '^[0-9a-f]{64}$') { throw 'Identity validation requires the retained create ID.' }
    $expectedBytes = [System.Text.Encoding]::ASCII.GetBytes($HeldId + "`n")
    if ($Result.StdoutBytes.Length -ne $expectedBytes.Length) {
        throw ('Docker ' + $Operation + ' stdout has the wrong length.')
    }
    for ($index = 0; $index -lt $expectedBytes.Length; $index++) {
        if ($Result.StdoutBytes[$index] -ne $expectedBytes[$index]) {
            throw ('Docker ' + $Operation + ' stdout does not equal HeldId plus LF.')
        }
    }
    return $HeldId
}

function private:Assert-WaitResult {
    param([object]$Result)

    if ($null -eq $Result -or $Result.Operation -cne 'wait') {
        throw 'The Docker wait result has the wrong operation identity.'
    }
    if ($Result.FailureExitCode -ne 0 -or $Result.CliExitCode -ne 0) {
        throw 'The Docker wait CLI did not complete successfully.'
    }
    if ($null -eq $Result.StdoutBytes -or $null -eq $Result.StderrBytes -or $Result.StderrBytes.Length -ne 0 -or
        $Result.StdoutBytes.Length -lt 2 -or $Result.StdoutBytes.Length -gt 4 -or
        $Result.StdoutBytes[$Result.StdoutBytes.Length - 1] -ne 10) {
        throw 'Docker wait streams do not match the exact decimal-plus-LF grammar.'
    }
    for ($index = 0; $index -lt $Result.StdoutBytes.Length - 1; $index++) {
        if ($Result.StdoutBytes[$index] -lt 48 -or $Result.StdoutBytes[$index] -gt 57) {
            throw 'Docker wait stdout contains a non-ASCII-decimal byte.'
        }
    }
    $text = [System.Text.Encoding]::ASCII.GetString($Result.StdoutBytes, 0, $Result.StdoutBytes.Length - 1)
    if ($text -notmatch '^(0|[1-9][0-9]{0,2})$') { throw 'Docker wait stdout is not canonical decimal.' }
    [int]$containedExitCode = -1
    if (-not [int]::TryParse($text, [System.Globalization.NumberStyles]::None, [System.Globalization.CultureInfo]::InvariantCulture, [ref]$containedExitCode) -or
        $containedExitCode -lt 0 -or $containedExitCode -gt 255) {
        throw 'Docker wait contained exit code is outside 0..255.'
    }
    return [pscustomobject]@{ ContainedExitCode = $containedExitCode; Successful = ($containedExitCode -eq 0) }
}

function private:Parse-InspectProjection {
    param([object]$Result, [string]$HeldId, [object]$Binding, [object]$TaskPaths, [object]$WorkState)

    if ($null -eq $Result -or $Result.Operation -cne 'inspect' -or $Result.FailureExitCode -ne 0 -or
        $Result.CliExitCode -ne 0) {
        throw 'The Docker inspect CLI did not complete successfully.'
    }
    if ($HeldId -notmatch '^[0-9a-f]{64}$' -or $Binding.Binding -cne 'BOUND' -or
        $Binding.BaseImageId -notmatch '^sha256:[0-9a-f]{64}$' -or $Binding.BoundPlatform -cne 'linux/amd64') {
        throw 'Inspect validation requires the retained ID and accepted linux/amd64 binding.'
    }
    if ($null -eq $Result.StdoutBytes -or $null -eq $Result.StderrBytes -or $Result.StderrBytes.Length -ne 0 -or
        $Result.StdoutBytes.Length -eq 0) {
        throw 'Docker inspect streams are missing or stderr is nonempty.'
    }
    if ($Result.StdoutBytes.Length -ge 3 -and $Result.StdoutBytes[0] -eq 0xef -and
        $Result.StdoutBytes[1] -eq 0xbb -and $Result.StdoutBytes[2] -eq 0xbf) {
        throw 'Docker inspect stdout must not contain a UTF-8 BOM.'
    }
    try { $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($Result.StdoutBytes) }
    catch { throw 'Docker inspect stdout is not strict UTF-8.' }
    foreach ($character in $text.ToCharArray()) {
        if ([char]::IsControl($character) -and $character -ne [char]0x09 -and $character -ne [char]0x0a) {
            throw 'Docker inspect stdout contains a forbidden control character.'
        }
    }
    $lines = $text.Split([char]0x0a)
    if ($lines.Count -ne 6 -or $lines[5].Length -ne 0) {
        throw 'Docker inspect stdout must be exactly five LF-terminated lines.'
    }
    $header = $lines[0].Split([char]0x09)
    if ($header.Count -ne 4 -or $header[0] -cne $HeldId -or $header[1] -cne $Binding.BaseImageId -or
        $header[2] -cne 'none' -or $header[3] -cne 'true') {
        throw 'Docker inspect header does not prove ID, image, network, and read-only rootfs.'
    }
    $mounts = [System.Collections.Generic.Dictionary[string, object]]::new([System.StringComparer]::Ordinal)
    for ($lineIndex = 1; $lineIndex -le 4; $lineIndex++) {
        $fields = $lines[$lineIndex].Split([char]0x09)
        if ($fields.Count -ne 4 -or $fields[0].Length -eq 0 -or $fields[1].Length -eq 0 -or
            $fields[2].Length -eq 0 -or ($fields[3] -cne 'true' -and $fields[3] -cne 'false')) {
            throw 'A Docker inspect mount line has the wrong four-field grammar.'
        }
        if ($mounts.ContainsKey($fields[2])) { throw 'Docker inspect contains a duplicate mount destination.' }
        $mounts.Add($fields[2], [pscustomobject]@{ Type = $fields[0]; Source = $fields[1]; Destination = $fields[2]; RW = $fields[3] })
    }
    $expected = [ordered]@{
        '/cache' = [pscustomobject]@{ Source = $TaskPaths.CacheRoot; RW = 'false' }
        '/ingress' = [pscustomobject]@{ Source = $TaskPaths.SourceIngressRoot; RW = 'false' }
        '/evidence' = [pscustomobject]@{ Source = $WorkState.EvidenceRoot; RW = 'true' }
        '/derived' = [pscustomobject]@{ Source = $WorkState.DerivedRoot; RW = 'true' }
    }
    if ($mounts.Count -ne $expected.Count) { throw 'Docker inspect contains an unexpected mount count.' }
    foreach ($destination in $expected.Keys) {
        if (-not $mounts.ContainsKey($destination)) { throw 'Docker inspect is missing an exact destination.' }
        $mount = $mounts[$destination]
        if ($mount.Type -cne 'bind' -or $mount.Source -cne $expected[$destination].Source -or
            $mount.RW -cne $expected[$destination].RW) {
            throw 'Docker inspect mount type, source, or RW state drifted.'
        }
    }
    return [pscustomobject]@{
        HeldId = $header[0]
        ObservedImageId = $header[1]
        BoundPlatform = 'linux/amd64'
        PlatformEvidence = 'accepted binding + create --platform=linux/amd64 + inspected .Image equality; not standalone Docker architecture'
        NetworkMode = 'none'
        ReadonlyRootfs = $true
        CacheMountMode = 'ro'
        SourceIngressMountMode = 'ro'
        Mounts = $mounts
    }
}
function private:Remove-HeldContainer {
    param([object]$TaskPaths, [object]$WorkState, [object]$Binding, [AllowNull()][string]$HeldId, [object]$Deadline)

    if ([string]::IsNullOrEmpty($HeldId)) {
        return [pscustomobject]@{ Attempted = $false; Removed = $true; ExitCode = 0; Reason = 'not-created'; Result = $null; ResidualProof = 'no-held-id' }
    }
    if ($HeldId -notmatch '^[0-9a-f]{64}$') {
        return [pscustomobject]@{ Attempted = $false; Removed = $false; ExitCode = 71; Reason = 'invalid-held-id'; Result = $null; ResidualProof = 'none' }
    }
    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
        return [pscustomobject]@{ Attempted = $false; Removed = $false; ExitCode = 124; Reason = 'deadline'; Result = $null; ResidualProof = 'none' }
    }
    $result = Invoke-DockerPhysicalProcess -TaskPaths $TaskPaths -WorkState $WorkState -Binding $Binding -Operation rm -HeldId $HeldId -Deadline $Deadline
    if ($result.FailureExitCode -ne 0) {
        $code = if ($result.FailureExitCode -eq 124) { 124 } else { 71 }
        return [pscustomobject]@{ Attempted = $result.ProcessStarted; Removed = $false; ExitCode = $code; Reason = ('rm-' + $result.FailureKind.ToLowerInvariant()); Result = $result; ResidualProof = 'none' }
    }
    try {
        [void](Assert-IdentityCommandResult -Result $result -Operation rm -HeldId $HeldId)
    }
    catch {
        $code = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 71 }
        return [pscustomobject]@{ Attempted = $true; Removed = $false; ExitCode = $code; Reason = 'rm-result-invalid'; Result = $result; ResidualProof = 'none' }
    }
    return [pscustomobject]@{ Attempted = $true; Removed = $true; ExitCode = 0; Reason = 'held-id-rm-confirmed'; Result = $result; ResidualProof = 'exact-rm-stdout' }
}

function private:Remove-ExactVerifiedWork {
    param(
        [object]$WorkState,
        [bool]$ContainerCleanupSatisfied,
        [bool]$RetainReservation,
        [object]$Deadline
    )

    $removedPaths = [System.Collections.Generic.List[string]]::new()
    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
        return [pscustomobject]@{ ExitCode = 124; Classification = 'ORDINARY_EXECUTION_FAILURE'; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = 'deadline'; RemovedPaths = $removedPaths.ToArray() }
    }
    $ownership = [pscustomobject]@{ Classification = 'ORDINARY_EXECUTION_FAILURE'; Reason = '' }
    if ($WorkState.InitializationComplete) {
        $ownership = Assert-WorkOwnershipAndTree -WorkState $WorkState -Phase PostDocker
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) {
            return [pscustomobject]@{ ExitCode = 124; Classification = $ownership.Classification; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = 'deadline'; RemovedPaths = $removedPaths.ToArray() }
        }
        if ($ownership.Classification -ceq 'OWNER_LOSS') {
            return [pscustomobject]@{ ExitCode = 71; Classification = 'OWNER_LOSS'; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = $ownership.Reason; RemovedPaths = $removedPaths.ToArray() }
        }
    }
    else {
        try {
            if ($RetainReservation) { throw 'A partial work state cannot retain its reservation.' }
            $comparison = [System.StringComparison]::OrdinalIgnoreCase
            $reservation = [System.IO.FileInfo]::new([System.IO.Path]::GetFullPath($WorkState.ReservationPath))
            $reservation.Refresh()
            $reservationStreams = @(Get-Item -LiteralPath $WorkState.ReservationPath -Stream * -Force -ErrorAction Stop)
            $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
            if (-not $reservation.Exists -or ($reservation.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($reservation.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($reservation.FullName, $WorkState.ReservationPath, $comparison) -or $reservationStreams.Count -ne 1 -or $reservationStreams[0].Stream -cne ':$DATA' -or $reservationBytes.Length -ne $WorkState.ReservationBytes.Length -or (Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256) { throw 'The partial reservation identity or bytes changed.' }
            for ($index = 0; $index -lt $reservationBytes.Length; $index++) { if ($reservationBytes[$index] -ne $WorkState.ReservationBytes[$index]) { throw 'The partial reservation bytes changed.' } }
            $childSpecifications = @(
                [pscustomobject]@{ Name = '.owner.json'; Path = $WorkState.OwnerPath; Created = [bool]$WorkState.OwnerCreated; Kind = 'File'; Bytes = $WorkState.OwnerBytes },
                [pscustomobject]@{ Name = 'evidence'; Path = $WorkState.EvidenceRoot; Created = [bool]$WorkState.EvidenceCreated; Kind = 'Directory'; Bytes = $null },
                [pscustomobject]@{ Name = 'derived'; Path = $WorkState.DerivedRoot; Created = [bool]$WorkState.DerivedCreated; Kind = 'Directory'; Bytes = $null },
                [pscustomobject]@{ Name = 'docker-config'; Path = $WorkState.DockerConfigRoot; Created = [bool]$WorkState.DockerConfigCreated; Kind = 'Directory'; Bytes = $null }
            )
            if (-not $WorkState.WorkCreated) {
                if ([System.IO.File]::Exists($WorkState.WorkRoot) -or [System.IO.Directory]::Exists($WorkState.WorkRoot) -or @($childSpecifications | Where-Object Created).Count -ne 0) { throw 'A reservation-only state has an unexpected work tree.' }
            }
            else {
                $work = [System.IO.DirectoryInfo]::new([System.IO.Path]::GetFullPath($WorkState.WorkRoot))
                $work.Refresh()
                $workStreams = @(Get-Item -LiteralPath $WorkState.WorkRoot -Stream * -Force -ErrorAction Stop)
                if (-not $work.Exists -or ($work.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($work.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($work.FullName, $WorkState.WorkRoot, $comparison) -or $workStreams.Count -gt 1 -or ($workStreams.Count -eq 1 -and $workStreams[0].Stream -cne ':$DATA')) { throw 'The partial work root changed identity.' }
                $actualEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($WorkState.WorkRoot))
                if ($actualEntries.Count -ne @($childSpecifications | Where-Object Created).Count) { throw 'The partial work root has unexpected children.' }
                foreach ($specification in $childSpecifications) {
                    $exists = [System.IO.File]::Exists($specification.Path) -or [System.IO.Directory]::Exists($specification.Path)
                    if ($exists -ne $specification.Created) { throw 'A partial work child disagrees with its creation state.' }
                    if (-not $specification.Created) { continue }
                    $child = Get-Item -LiteralPath $specification.Path -Force -ErrorAction Stop
                    $childStreams = @(Get-Item -LiteralPath $specification.Path -Stream * -Force -ErrorAction Stop)
                    if (($child.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals([System.IO.Path]::GetFullPath($child.FullName), $specification.Path, $comparison) -or ($specification.Kind -ceq 'File' -and (($child.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or $childStreams.Count -ne 1 -or $childStreams[0].Stream -cne ':$DATA')) -or ($specification.Kind -ceq 'Directory' -and (($child.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or $childStreams.Count -gt 1 -or ($childStreams.Count -eq 1 -and $childStreams[0].Stream -cne ':$DATA') -or @([System.IO.Directory]::EnumerateFileSystemEntries($specification.Path)).Count -ne 0))) { throw 'A partial work child changed identity, type, streams, or closure.' }
                    if ($specification.Kind -ceq 'File') {
                        $actualBytes = [System.IO.File]::ReadAllBytes($specification.Path)
                        if ($null -eq $specification.Bytes -or $actualBytes.Length -ne $specification.Bytes.Length) { throw 'The partial owner bytes changed.' }
                        for ($index = 0; $index -lt $actualBytes.Length; $index++) { if ($actualBytes[$index] -ne $specification.Bytes[$index]) { throw 'The partial owner bytes changed.' } }
                    }
                }
            }
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        }
        catch {
            $code = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 71 }
            $classification = if ($code -eq 124) { 'ORDINARY_EXECUTION_FAILURE' } else { 'OWNER_LOSS' }
            return [pscustomobject]@{ ExitCode = $code; Classification = $classification; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = $_.Exception.Message; RemovedPaths = $removedPaths.ToArray() }
        }
    }
    if (-not $ContainerCleanupSatisfied) {
        return [pscustomobject]@{ ExitCode = 71; Classification = $ownership.Classification; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = 'container-cleanup-unproved'; RemovedPaths = $removedPaths.ToArray() }
    }
    try {
        $comparison = [System.StringComparison]::OrdinalIgnoreCase
        foreach ($filePath in @($WorkState.AssetsPath, $WorkState.ConfigPath, $WorkState.DerivationPath)) {
            if (-not [System.IO.File]::Exists($filePath)) { continue }
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
            $file = [System.IO.FileInfo]::new([System.IO.Path]::GetFullPath($filePath))
            $file.Refresh()
            if (-not $file.Exists -or ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or -not [string]::Equals($file.FullName, $filePath, $comparison)) {
                throw 'An allowlisted work output changed before exact deletion.'
            }
            $streams = @(Get-Item -LiteralPath $filePath -Stream * -Force -ErrorAction Stop)
            if ($streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA') { throw 'An allowlisted work output acquired an alternate stream.' }
            [System.IO.File]::Delete($filePath)
            if ([System.IO.File]::Exists($filePath) -or [System.IO.Directory]::Exists($filePath)) { throw 'An allowlisted work output remained after exact deletion.' }
            $removedPaths.Add($filePath)
        }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        if ($WorkState.OwnerCreated) {
            $owner = [System.IO.FileInfo]::new($WorkState.OwnerPath)
            $owner.Refresh()
            if (-not $owner.Exists -or ($owner.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
                ($owner.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'The work owner changed before exact deletion.' }
            $ownerStreams = @(Get-Item -LiteralPath $WorkState.OwnerPath -Stream * -Force -ErrorAction Stop)
            if ($ownerStreams.Count -ne 1 -or $ownerStreams[0].Stream -cne ':$DATA') { throw 'The work owner streams changed before deletion.' }
            $ownerActualBytes = [System.IO.File]::ReadAllBytes($WorkState.OwnerPath)
            if ($ownerActualBytes.Length -ne $WorkState.OwnerBytes.Length) { throw 'The work owner bytes changed before deletion.' }
            for ($index = 0; $index -lt $ownerActualBytes.Length; $index++) { if ($ownerActualBytes[$index] -ne $WorkState.OwnerBytes[$index]) { throw 'The work owner bytes changed before deletion.' } }
            [System.IO.File]::Delete($WorkState.OwnerPath)
            $removedPaths.Add($WorkState.OwnerPath)
        }
        $directorySpecifications = @([pscustomobject]@{ Path = $WorkState.DockerConfigRoot; Created = $WorkState.DockerConfigCreated }, [pscustomobject]@{ Path = $WorkState.DerivedRoot; Created = $WorkState.DerivedCreated }, [pscustomobject]@{ Path = $WorkState.EvidenceRoot; Created = $WorkState.EvidenceCreated }, [pscustomobject]@{ Path = $WorkState.WorkRoot; Created = $WorkState.WorkCreated })
        foreach ($directorySpecification in $directorySpecifications) {
            if (-not $directorySpecification.Created) { continue }
            $directoryPath = $directorySpecification.Path
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
            $directory = [System.IO.DirectoryInfo]::new($directoryPath)
            $directory.Refresh()
            $directoryStreams = @(Get-Item -LiteralPath $directoryPath -Stream * -Force -ErrorAction Stop)
            if (-not $directory.Exists -or ($directory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
                ($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
                $directory.FullName -cne $directoryPath -or $directoryStreams.Count -gt 1 -or
                ($directoryStreams.Count -eq 1 -and $directoryStreams[0].Stream -cne ':$DATA') -or @([System.IO.Directory]::EnumerateFileSystemEntries($directoryPath)).Count -ne 0) {
                throw 'An owned work directory is no longer exact and empty.'
            }
            [System.IO.Directory]::Delete($directoryPath, $false)
            if ([System.IO.Directory]::Exists($directoryPath) -or [System.IO.File]::Exists($directoryPath)) { throw 'An owned work directory remained after exact deletion.' }
            $removedPaths.Add($directoryPath)
        }
        if (-not $RetainReservation) {
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
            $reservation = [System.IO.FileInfo]::new($WorkState.ReservationPath)
            $reservation.Refresh()
            if (-not $reservation.Exists -or ($reservation.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or
                ($reservation.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'The work reservation changed before exact deletion.' }
            $reservationStreams = @(Get-Item -LiteralPath $WorkState.ReservationPath -Stream * -Force -ErrorAction Stop)
            if ($reservationStreams.Count -ne 1 -or $reservationStreams[0].Stream -cne ':$DATA') { throw 'The work reservation streams changed before deletion.' }
            $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
            if ((Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256 -or
                $reservationBytes.Length -ne $WorkState.ReservationBytes.Length) { throw 'The retained reservation hash changed before deletion.' }
            for ($index = 0; $index -lt $reservationBytes.Length; $index++) {
                if ($reservationBytes[$index] -ne $WorkState.ReservationBytes[$index]) { throw 'The retained reservation bytes changed before deletion.' }
            }
            [System.IO.File]::Delete($WorkState.ReservationPath)
            if ([System.IO.File]::Exists($WorkState.ReservationPath)) { throw 'The retained reservation remained after exact deletion.' }
            $removedPaths.Add($WorkState.ReservationPath)
        }
    }
    catch {
        $code = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 71 }
        return [pscustomobject]@{ ExitCode = $code; Classification = $ownership.Classification; WorkRemoved = $false; ReservationRemoved = $false; ReservationRetained = $false; Abandoned = $true; Reason = $_.Exception.Message; RemovedPaths = $removedPaths.ToArray() }
    }
    return [pscustomobject]@{ ExitCode = 0; Classification = $ownership.Classification; WorkRemoved = $true; ReservationRemoved = (-not $RetainReservation); ReservationRetained = $RetainReservation; Abandoned = $false; Reason = 'exact-work-cleanup-complete'; RemovedPaths = $removedPaths.ToArray() }
}
function private:Publish-EligibleFinalEvidence {
    param([object]$TaskPaths, [object]$WorkState, [byte[]]$HeldAssetsBytes, [byte[]]$HeldConfigBytes, [byte[]]$HeldDerivationBytes, [object]$Deadline)

    $stagingRoot = $TaskPaths.EvidenceRoot + '\.dependency-admission-derived-staging-' + $WorkState.OwnerNonce; $stagingOwnerPath = $stagingRoot + '\.owner.json'; $stagingConfigPath = $stagingRoot + '\vitest.frontend-unit.config.ts'; $stagingDerivationPath = $stagingRoot + '\vitest-derivation.v1.json'
    $utf8 = [System.Text.UTF8Encoding]::new($false, $true); $ownerJson = '{"schema":"hub.qa.local-dependency-admission-derived-staging-owner/v1","task_run":"' + $WorkState.TaskRunId + '","owner_nonce":"' + $WorkState.OwnerNonce + '","reservation_sha256":"' + $WorkState.ReservationSha256 + '"}'; $ownerBytes = $utf8.GetBytes($ownerJson); $configSha256 = Get-ByteSha256 -Bytes $HeldConfigBytes; $derivationSha256 = Get-ByteSha256 -Bytes $HeldDerivationBytes
    $stagingCreated = $false; $ownerVerified = $false; $derivedMoved = $false; $assetsWritten = $false; $stagingRemoved = $false; $stagingAbandoned = $false; $reservationRemoved = $false; $failureCode = 0; $failureReason = ''
    try {
        if ($HeldAssetsBytes.Length -eq 0 -or $HeldConfigBytes.Length -eq 0 -or $HeldDerivationBytes.Length -eq 0) { throw 'Publication requires all three nonempty held outputs.' }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $reservation = [System.IO.FileInfo]::new($WorkState.ReservationPath)
        $reservation.Refresh()
        if (-not $reservation.Exists -or ($reservation.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($reservation.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'The retained publication reservation is not an exact regular file.' }
        $reservationStreams = @(Get-Item -LiteralPath $WorkState.ReservationPath -Stream * -Force -ErrorAction Stop)
        $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
        if ($reservationStreams.Count -ne 1 -or $reservationStreams[0].Stream -cne ':$DATA' -or $reservationBytes.Length -ne $WorkState.ReservationBytes.Length -or (Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256) { throw 'The retained publication reservation changed.' }
        for ($index = 0; $index -lt $reservationBytes.Length; $index++) { if ($reservationBytes[$index] -ne $WorkState.ReservationBytes[$index]) { throw 'The retained publication reservation bytes changed.' } }
        foreach ($finalPath in @($TaskPaths.FinalRunPath, $TaskPaths.FinalAssetsPath, $TaskPaths.FinalCleanupPath, $TaskPaths.FinalDerivedRoot, $TaskPaths.FinalDerivedConfigPath, $TaskPaths.FinalDerivationPath)) { if ([System.IO.File]::Exists($finalPath) -or [System.IO.Directory]::Exists($finalPath)) { throw 'A final path appeared before publication.' } }
        $nativeDirectoryType = ([System.Management.Automation.PSTypeName]'HubQaNativeDirectory').Type; if ($null -eq $nativeDirectoryType) { throw 'The fixed exclusive directory primitive is unavailable for publication.' }
        New-ExclusiveOwnedDirectory -Path $stagingRoot -ParentPath $TaskPaths.EvidenceRoot -NativeDirectoryType $nativeDirectoryType | Out-Null
        $stagingCreated = $true; $fileSpecifications = @([pscustomobject]@{ Path = $stagingOwnerPath; Bytes = $ownerBytes; Sha256 = (Get-ByteSha256 -Bytes $ownerBytes) }, [pscustomobject]@{ Path = $stagingConfigPath; Bytes = $HeldConfigBytes; Sha256 = $configSha256 }, [pscustomobject]@{ Path = $stagingDerivationPath; Bytes = $HeldDerivationBytes; Sha256 = $derivationSha256 })
        foreach ($fileSpecification in $fileSpecifications) {
            if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
            $stream = [System.IO.File]::Open($fileSpecification.Path, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
            try {
                $stream.Write($fileSpecification.Bytes, 0, $fileSpecification.Bytes.Length)
                $stream.Flush($true)
            }
            finally { $stream.Dispose() }
        }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
        if ($reservationBytes.Length -ne $WorkState.ReservationBytes.Length -or (Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256) { throw 'The publication reservation changed during staging.' }
        $stagingDirectory = [System.IO.DirectoryInfo]::new($stagingRoot)
        $stagingDirectory.Refresh()
        $stagingDirectoryStreams = @(Get-Item -LiteralPath $stagingRoot -Stream * -Force -ErrorAction Stop)
        if (-not $stagingDirectory.Exists -or ($stagingDirectory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($stagingDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $stagingDirectory.FullName -cne $stagingRoot -or $stagingDirectoryStreams.Count -gt 1 -or ($stagingDirectoryStreams.Count -eq 1 -and $stagingDirectoryStreams[0].Stream -cne ':$DATA')) { throw 'Publication staging changed directory identity.' }
        $actualEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($stagingRoot))
        if ($actualEntries.Count -ne 3) { throw 'Publication staging does not have the exact three-file closure.' }
        foreach ($fileSpecification in $fileSpecifications) {
            $file = [System.IO.FileInfo]::new($fileSpecification.Path)
            $file.Refresh()
            if (-not $file.Exists -or ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $file.FullName -cne $fileSpecification.Path -or -not $actualEntries.Contains($fileSpecification.Path)) { throw 'A publication staging file changed identity.' }
            $streams = @(Get-Item -LiteralPath $fileSpecification.Path -Stream * -Force -ErrorAction Stop)
            $bytes = [System.IO.File]::ReadAllBytes($fileSpecification.Path)
            if ($streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA' -or $bytes.Length -ne $fileSpecification.Bytes.Length -or (Get-ByteSha256 -Bytes $bytes) -cne $fileSpecification.Sha256) { throw 'A publication staging file changed bytes or streams.' }
        }
        $ownerVerified = $true
        [System.IO.File]::Delete($stagingOwnerPath)
        if ([System.IO.File]::Exists($stagingOwnerPath)) { throw 'The verified staging owner marker was not removed.' }
        $remainingEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($stagingRoot))
        if ($remainingEntries.Count -ne 2 -or -not $remainingEntries.Contains($stagingConfigPath) -or -not $remainingEntries.Contains($stagingDerivationPath)) { throw 'Publication staging lost its exact two-file move closure.' }
        if ([System.IO.Directory]::Exists($TaskPaths.FinalDerivedRoot) -or [System.IO.File]::Exists($TaskPaths.FinalDerivedRoot)) { throw 'The final derived destination appeared before the atomic move.' }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        [System.IO.Directory]::Move($stagingRoot, $TaskPaths.FinalDerivedRoot)
        $derivedMoved = $true; $stagingRemoved = $true
        $finalDirectory = [System.IO.DirectoryInfo]::new($TaskPaths.FinalDerivedRoot)
        $finalDirectory.Refresh()
        $finalDirectoryStreams = @(Get-Item -LiteralPath $TaskPaths.FinalDerivedRoot -Stream * -Force -ErrorAction Stop)
        if ([System.IO.Directory]::Exists($stagingRoot) -or [System.IO.File]::Exists($stagingRoot) -or -not $finalDirectory.Exists -or ($finalDirectory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($finalDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $finalDirectory.FullName -cne $TaskPaths.FinalDerivedRoot -or $finalDirectoryStreams.Count -gt 1 -or ($finalDirectoryStreams.Count -eq 1 -and $finalDirectoryStreams[0].Stream -cne ':$DATA')) { throw 'The final derived directory changed identity after the atomic move.' }
        $finalEntries = @([System.IO.Directory]::EnumerateFileSystemEntries($TaskPaths.FinalDerivedRoot))
        if ($finalEntries.Count -ne 2 -or -not $finalEntries.Contains($TaskPaths.FinalDerivedConfigPath) -or -not $finalEntries.Contains($TaskPaths.FinalDerivationPath)) { throw 'The atomically moved derived evidence failed exact path closure.' }
        foreach ($finalSpecification in @([pscustomobject]@{ Path = $TaskPaths.FinalDerivedConfigPath; Sha256 = $configSha256 }, [pscustomobject]@{ Path = $TaskPaths.FinalDerivationPath; Sha256 = $derivationSha256 })) {
            $finalFile = [System.IO.FileInfo]::new($finalSpecification.Path)
            $finalFile.Refresh()
            $finalStreams = @(Get-Item -LiteralPath $finalSpecification.Path -Stream * -Force -ErrorAction Stop)
            if (-not $finalFile.Exists -or ($finalFile.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($finalFile.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $finalFile.FullName -cne $finalSpecification.Path -or $finalStreams.Count -ne 1 -or $finalStreams[0].Stream -cne ':$DATA' -or (Get-ByteSha256 -Bytes ([System.IO.File]::ReadAllBytes($finalSpecification.Path))) -cne $finalSpecification.Sha256) { throw 'The atomically moved derived evidence failed exact identity or hash closure.' }
        }
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
        if ($reservationBytes.Length -ne $WorkState.ReservationBytes.Length -or (Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256) { throw 'The reservation changed before its post-move removal.' }
        [System.IO.File]::Delete($WorkState.ReservationPath)
        if ([System.IO.File]::Exists($WorkState.ReservationPath)) { throw 'The retained reservation remained after publication outcome.' }
        $reservationRemoved = $true
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $assetsStream = [System.IO.File]::Open($TaskPaths.FinalAssetsPath, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
        try {
            $assetsStream.Write($HeldAssetsBytes, 0, $HeldAssetsBytes.Length)
            $assetsStream.Flush($true)
        }
        finally { $assetsStream.Dispose() }
        $assetsWritten = $true
    }
    catch {
        $failureCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }; $failureReason = $_.Exception.Message
        if ($stagingCreated -and -not $derivedMoved -and $ownerVerified -and [System.Diagnostics.Stopwatch]::GetTimestamp() -lt $Deadline.EndTicks) {
            try {
                $allowed = @($stagingOwnerPath, $stagingConfigPath, $stagingDerivationPath)
                $stagingDirectory = [System.IO.DirectoryInfo]::new($stagingRoot)
                $stagingDirectory.Refresh()
                if (-not $stagingDirectory.Exists -or ($stagingDirectory.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($stagingDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $stagingDirectory.FullName -cne $stagingRoot) { throw 'Publication staging lost its owned directory identity.' }
                $entries = @([System.IO.Directory]::EnumerateFileSystemEntries($stagingRoot))
                foreach ($entry in $entries) { if (-not $allowed.Contains($entry)) { throw 'Publication staging gained a foreign entry.' } }
                foreach ($fileSpecification in $fileSpecifications) {
                    if (-not [System.IO.File]::Exists($fileSpecification.Path)) { continue }
                    $file = [System.IO.FileInfo]::new($fileSpecification.Path)
                    $file.Refresh()
                    $streams = @(Get-Item -LiteralPath $fileSpecification.Path -Stream * -Force -ErrorAction Stop)
                    $bytes = [System.IO.File]::ReadAllBytes($fileSpecification.Path)
                    if ($file.FullName -cne $fileSpecification.Path -or ($file.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0 -or ($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $streams.Count -ne 1 -or $streams[0].Stream -cne ':$DATA' -or (Get-ByteSha256 -Bytes $bytes) -cne $fileSpecification.Sha256) { throw 'Publication staging changed before safe cleanup.' }
                }
                foreach ($path in @($stagingConfigPath, $stagingDerivationPath)) {
                    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                    if ([System.IO.File]::Exists($path)) { [System.IO.File]::Delete($path) }
                }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                if ([System.IO.File]::Exists($stagingOwnerPath)) { [System.IO.File]::Delete($stagingOwnerPath) }
                if (@([System.IO.Directory]::EnumerateFileSystemEntries($stagingRoot)).Count -ne 0) { throw 'Publication staging is not empty after exact cleanup.' }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                [System.IO.Directory]::Delete($stagingRoot, $false)
                $stagingRemoved = $true
            }
            catch { $stagingAbandoned = $true }
        }
        elseif ($stagingCreated -and -not $derivedMoved) { $stagingAbandoned = $true }
        if (($derivedMoved -or $stagingRemoved) -and -not $reservationRemoved -and [System.Diagnostics.Stopwatch]::GetTimestamp() -lt $Deadline.EndTicks) {
            try {
                $reservationBytes = [System.IO.File]::ReadAllBytes($WorkState.ReservationPath)
                if ((Get-ByteSha256 -Bytes $reservationBytes) -cne $WorkState.ReservationSha256) { throw 'The retained reservation changed after staging outcome.' }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                [System.IO.File]::Delete($WorkState.ReservationPath)
                $reservationRemoved = -not [System.IO.File]::Exists($WorkState.ReservationPath)
            }
            catch { $stagingAbandoned = $true }
        }
    }
    return [pscustomobject]@{ ExitCode = $failureCode; DerivedMoved = $derivedMoved; AssetsWritten = $assetsWritten; StagingRemoved = $stagingRemoved; StagingAbandoned = $stagingAbandoned; ReservationRemoved = $reservationRemoved; StagingRoot = $stagingRoot; StagingOwnerPath = $stagingOwnerPath; FailureReason = $failureReason; FinalEvidencePreserved = $derivedMoved }
}
function private:Write-FinalHostReceipts {
    param([object]$TaskPaths, [object]$WorkState, [string]$BaseImageId, [string]$HandoffManifestSha256, [string]$SourceBlobsManifestSha256, [AllowNull()][object]$InspectProof, [string]$ContainerName, [object]$CleanupState, [datetime]$StartedAt, [int]$FirstExitCode, [object]$Deadline)

    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { return [pscustomobject]@{ ExitCode = $(if ($FirstExitCode -ne 0) { $FirstExitCode } else { 124 }); RunWritten = $false; CleanupWritten = $false; FinalEvidencePreserved = $CleanupState.FinalEvidenceAlreadyPreserved; AllOwnedRemoved = $false; Status = 'FAIL' } }
    $utf8 = [System.Text.UTF8Encoding]::new($false, $true)
    $runWritten = $false; $cleanupWritten = $false; $receiptFailureCode = 0
    $observedImageId = ''; $observedPlatform = ''; $networkMode = ''; $cacheMountMode = ''; $sourceIngressMountMode = ''
    if ($null -ne $InspectProof) {
        $observedImageId = $InspectProof.ObservedImageId
        $observedPlatform = $InspectProof.BoundPlatform
        $networkMode = $InspectProof.NetworkMode
        $cacheMountMode = $InspectProof.CacheMountMode
        $sourceIngressMountMode = $InspectProof.SourceIngressMountMode
    }
    $runStatus = if ($FirstExitCode -eq 0) { 'STATIC' } else { 'FAIL' }
    $startedText = $StartedAt.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    $finishedText = [datetime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    $runJson = '{"schema":"hub.qa.local-dependency-admission-run/v1","task_run":"' + $WorkState.TaskRunId + '","owner_nonce":"' + $WorkState.OwnerNonce + '","cache_handoff_manifest_sha256":"' + $HandoffManifestSha256 + '","source_blobs_manifest_sha256":"' + $SourceBlobsManifestSha256 + '","supplied_base_image_id":"' + $BaseImageId + '","observed_base_image_id":"' + $observedImageId + '","observed_platform":"' + $observedPlatform + '","network_mode":"' + $networkMode + '","cache_mount_mode":"' + $cacheMountMode + '","source_ingress_mount_mode":"' + $sourceIngressMountMode + '","source_mount_present":false,"started_at":"' + $startedText + '","finished_at":"' + $finishedText + '","status":"' + $runStatus + '"}'
    try {
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $finalParent = [System.IO.DirectoryInfo]::new($TaskPaths.DependencyAdmissionRoot)
        $finalParent.Refresh()
        if (-not $finalParent.Exists -or ($finalParent.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($finalParent.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $finalParent.FullName -cne $TaskPaths.DependencyAdmissionRoot) { throw 'The final receipt parent changed identity.' }
        if ([System.IO.File]::Exists($TaskPaths.FinalRunPath) -or [System.IO.Directory]::Exists($TaskPaths.FinalRunPath)) { throw 'Final run receipt appeared before CreateNew.' }
        $runBytes = $utf8.GetBytes($runJson + "`n")
        $runStream = [System.IO.File]::Open($TaskPaths.FinalRunPath, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
        try {
            $runStream.Write($runBytes, 0, $runBytes.Length)
            $runStream.Flush($true)
        }
        finally { $runStream.Dispose() }
        $runWritten = $true
    }
    catch {
        $receiptFailureCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 }
    }
    $effectiveExitCode = if ($FirstExitCode -ne 0) { $FirstExitCode } else { $receiptFailureCode }
    $finalEvidencePreserved = $CleanupState.FinalEvidenceAlreadyPreserved -or $runWritten
    $allOwnedRemoved = $CleanupState.ContainerRemoved -and $CleanupState.WorkRemoved -and $CleanupState.ReservationRemoved -and $CleanupState.StagingResolved
    if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { return [pscustomobject]@{ ExitCode = $(if ($effectiveExitCode -ne 0) { $effectiveExitCode } else { 124 }); RunWritten = $runWritten; CleanupWritten = $false; FinalEvidencePreserved = $finalEvidencePreserved; AllOwnedRemoved = $allOwnedRemoved; Status = 'FAIL' } }
    $cleanupStatus = if ($effectiveExitCode -eq 0 -and $allOwnedRemoved -and $finalEvidencePreserved) { 'STATIC' } else { 'FAIL' }
    $containerAttempted = $CleanupState.ContainerAttempted.ToString().ToLowerInvariant(); $containerRemoved = $CleanupState.ContainerRemoved.ToString().ToLowerInvariant()
    $reservationAttempted = $CleanupState.ReservationAttempted.ToString().ToLowerInvariant(); $reservationRemoved = $CleanupState.ReservationRemoved.ToString().ToLowerInvariant()
    $allOwnedRemovedText = $allOwnedRemoved.ToString().ToLowerInvariant(); $finalEvidencePreservedText = $finalEvidencePreserved.ToString().ToLowerInvariant()
    $containerReason = if (-not $CleanupState.ContainerAttempted) { 'not-created' } elseif ($CleanupState.ContainerRemoved) { 'held-id-rm-confirmed' } else { 'held-id-rm-unproved' }
    $mountReason = if (-not $CleanupState.ContainerAttempted) { 'not-attached' } elseif ($CleanupState.ContainerRemoved) { 'detached-by-held-id-rm' } else { 'detach-unproved' }
    $reservationReason = if ($CleanupState.ReservationRemoved) { 'exact-reservation-removed' } else { 'reservation-retained-or-abandoned' }
    $temporaryReason = if ($CleanupState.WorkRemoved -and $CleanupState.StagingResolved) { 'exact-temporary-resources-resolved' } else { 'temporary-resources-retained-or-abandoned' }
    $containerPath = $ContainerName
    $cachePath = $TaskPaths.CacheRoot.Replace('\', '\\'); $ingressPath = $TaskPaths.SourceIngressRoot.Replace('\', '\\')
    $evidencePath = $WorkState.EvidenceRoot.Replace('\', '\\'); $derivedPath = $WorkState.DerivedRoot.Replace('\', '\\')
    $reservationPath = $WorkState.ReservationPath.Replace('\', '\\'); $workPath = $WorkState.WorkRoot.Replace('\', '\\')
    $ownerPath = $WorkState.OwnerPath.Replace('\', '\\'); $dockerConfigPath = $WorkState.DockerConfigRoot.Replace('\', '\\')
    $publicationStagingPath = $CleanupState.PublicationStaging.Replace('\', '\\'); $publicationOwnerPath = $CleanupState.PublicationStagingOwner.Replace('\', '\\')
    $ownedJson = '{"container_name":"' + $containerPath + '","cache_mount":"' + $cachePath + '","source_ingress_mount":"' + $ingressPath + '","evidence_mount":"' + $evidencePath + '","derived_mount":"' + $derivedPath + '","staging_reservation":"' + $reservationPath + '","temporary_reservation":{"work_root":"' + $workPath + '","work_owner":"' + $ownerPath + '","docker_config":"' + $dockerConfigPath + '","publication_staging":"' + $publicationStagingPath + '","publication_staging_owner":"' + $publicationOwnerPath + '"}}'
    $containerResult = '{"attempted":' + $containerAttempted + ',"removed":' + $containerRemoved + ',"reason":"' + $containerReason + '"}'
    $mountResult = '{"attempted":' + $containerAttempted + ',"removed":' + $containerRemoved + ',"reason":"' + $mountReason + '"}'
    $reservationResult = '{"attempted":' + $reservationAttempted + ',"removed":' + $reservationRemoved + ',"reason":"' + $reservationReason + '"}'
    $temporaryResult = '{"attempted":' + (($CleanupState.WorkAttempted -or $CleanupState.StagingAttempted).ToString().ToLowerInvariant()) + ',"removed":' + (($CleanupState.WorkRemoved -and $CleanupState.StagingResolved).ToString().ToLowerInvariant()) + ',"reason":"' + $temporaryReason + '"}'
    $resultsJson = '{"container_name":' + $containerResult + ',"cache_mount":' + $mountResult + ',"source_ingress_mount":' + $mountResult + ',"evidence_mount":' + $mountResult + ',"derived_mount":' + $mountResult + ',"staging_reservation":' + $reservationResult + ',"temporary_reservation":' + $temporaryResult + '}'
    $cleanupJson = '{"schema":"hub.qa.local-dependency-admission-cleanup/v1","task_run":"' + $WorkState.TaskRunId + '","owner_nonce":"' + $WorkState.OwnerNonce + '","finally_entered":true,"owned_resources":' + $ownedJson + ',"cleanup_results":' + $resultsJson + ',"all_owned_removed":' + $allOwnedRemovedText + ',"foreign_resources_touched":false,"final_evidence_preserved":' + $finalEvidencePreservedText + ',"status":"' + $cleanupStatus + '"}'
    try {
        if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
        $finalParent.Refresh()
        if (-not $finalParent.Exists -or ($finalParent.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or ($finalParent.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or $finalParent.FullName -cne $TaskPaths.DependencyAdmissionRoot) { throw 'The final receipt parent changed before cleanup receipt.' }
        if ([System.IO.File]::Exists($TaskPaths.FinalCleanupPath) -or [System.IO.Directory]::Exists($TaskPaths.FinalCleanupPath)) { throw 'Final cleanup receipt appeared before CreateNew.' }
        $cleanupBytes = $utf8.GetBytes($cleanupJson + "`n")
        $cleanupStream = [System.IO.File]::Open($TaskPaths.FinalCleanupPath, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
        try {
            $cleanupStream.Write($cleanupBytes, 0, $cleanupBytes.Length)
            $cleanupStream.Flush($true)
        }
        finally { $cleanupStream.Dispose() }
        $cleanupWritten = $true
    }
    catch {
        if ($effectiveExitCode -eq 0) { $effectiveExitCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $Deadline.EndTicks) { 124 } else { 70 } }
    }
    if ($effectiveExitCode -ne 0) { $cleanupStatus = 'FAIL' }
    return [pscustomobject]@{ ExitCode = $effectiveExitCode; RunWritten = $runWritten; CleanupWritten = $cleanupWritten; FinalEvidencePreserved = $finalEvidencePreserved; AllOwnedRemoved = $allOwnedRemoved; Status = $cleanupStatus }
}

function private:Invoke-LocalDependencyAdmission {
    param(
        [string]$BaseImageId, [string]$CacheRoot, [string]$HandoffManifestSha256,
        [string]$SourceIngressRoot, [string]$SourceBlobsManifestSha256, [string]$TaskRunId, [string]$OwnerNonce,
        [string]$EvidenceRoot
    )

    try {
        Assert-ExactInputs @PSBoundParameters
        Assert-ProfileByteEquality
        $taskPaths = Resolve-CanonicalTaskPaths -CacheRoot $CacheRoot -SourceIngressRoot $SourceIngressRoot -TaskRunId $TaskRunId -EvidenceRoot $EvidenceRoot
        Assert-IngressAndFinalClosure -TaskPaths $taskPaths
        $binding = Assert-BaseBinding -BaseImageId $BaseImageId
    }
    catch {
        return [pscustomobject]@{ ExitCode = 64; Status = 'BLOCKED'; Binding = 'UNBOUND' }
    }
    if ($binding.Binding -ceq 'UNBOUND') { return $binding }

    $startedAt = [datetime]::UtcNow
    $frequency = [long][System.Diagnostics.Stopwatch]::Frequency
    $startedTicks = [System.Diagnostics.Stopwatch]::GetTimestamp()
    $deadline = [pscustomobject]@{ StartedTicks = $startedTicks; EndTicks = $startedTicks + [long]($frequency * 300); Frequency = $frequency }
    $firstExitCode = 0
    $workState = $null
    $heldId = $null
    $createStarted = $false
    $waitAccepted = $false
    $waitProof = $null
    $inspectProof = $null
    $heldAssetsBytes = $null
    $heldConfigBytes = $null
    $heldDerivationBytes = $null
    $publicationCandidate = $false
    $finallyEntered = $false
    $containerName = 'hub-qa-local-dependency-admission-' + $TaskRunId + '-' + $OwnerNonce.Substring(0, 12)
    $publicationStaging = $EvidenceRoot + '\.dependency-admission-derived-staging-' + $OwnerNonce
    $publicationStagingOwner = $publicationStaging + '\.owner.json'
    $removeResult = [pscustomobject]@{ Attempted = $false; Removed = $true; ExitCode = 0; Reason = 'not-created' }
    $workCleanup = $null
    $publicationResult = $null
    $receiptResult = $null
    try {
        $reservationResult = New-WorkReservation -TaskPaths $taskPaths -OwnerNonce $OwnerNonce -Deadline $deadline
        $workState = $reservationResult.WorkState
        if ($reservationResult.ExitCode -ne 0) { $firstExitCode = $reservationResult.ExitCode }
        elseif ($null -eq $workState -or -not $workState.InitializationComplete) { $firstExitCode = 70 }
        if ($firstExitCode -eq 0) {
            $createResult = Invoke-DockerPhysicalProcess -TaskPaths $taskPaths -WorkState $workState -Binding $binding -Operation create -HeldId $null -Deadline $deadline
            $createStarted = $createResult.ProcessStarted
            if ($createResult.FailureExitCode -ne 0) { $firstExitCode = $createResult.FailureExitCode }
            else {
                try { $heldId = Assert-IdentityCommandResult -Result $createResult -Operation create -HeldId $null }
                catch { $firstExitCode = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } }
            }
        }
        if ($firstExitCode -eq 0) {
            $startResult = Invoke-DockerPhysicalProcess -TaskPaths $taskPaths -WorkState $workState -Binding $binding -Operation start -HeldId $heldId -Deadline $deadline
            if ($startResult.FailureExitCode -ne 0) { $firstExitCode = $startResult.FailureExitCode }
            else {
                try { [void](Assert-IdentityCommandResult -Result $startResult -Operation start -HeldId $heldId) }
                catch { $firstExitCode = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } }
            }
        }
        if ($firstExitCode -eq 0) {
            $waitResult = Invoke-DockerPhysicalProcess -TaskPaths $taskPaths -WorkState $workState -Binding $binding -Operation wait -HeldId $heldId -Deadline $deadline
            if ($waitResult.FailureExitCode -ne 0) { $firstExitCode = $waitResult.FailureExitCode }
            else {
                try {
                    $waitProof = Assert-WaitResult -Result $waitResult
                    $waitAccepted = $true
                    if (-not $waitProof.Successful) { $firstExitCode = 70 }
                }
                catch { $firstExitCode = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } }
            }
        }
        if ($waitAccepted -and [System.Diagnostics.Stopwatch]::GetTimestamp() -lt $deadline.EndTicks) {
            $inspectResult = Invoke-DockerPhysicalProcess -TaskPaths $taskPaths -WorkState $workState -Binding $binding -Operation inspect -HeldId $heldId -Deadline $deadline
            if ($inspectResult.FailureExitCode -ne 0) {
                if ($firstExitCode -eq 0) { $firstExitCode = $inspectResult.FailureExitCode }
            }
            else {
                try { $inspectProof = Parse-InspectProjection -Result $inspectResult -HeldId $heldId -Binding $binding -TaskPaths $taskPaths -WorkState $workState }
                catch { if ($firstExitCode -eq 0) { $firstExitCode = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } } }
            }
        }
        elseif ($waitAccepted -and $firstExitCode -eq 0) { $firstExitCode = 124 }
        if ($firstExitCode -eq 0 -and $waitProof.Successful -and $null -ne $inspectProof) {
            try {
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                $postDockerOwnership = Assert-WorkOwnershipAndTree -WorkState $workState -Phase PostDocker
                if ($postDockerOwnership.Classification -ceq 'OWNER_LOSS') { throw ('Owner loss before temporary validation: ' + $postDockerOwnership.Reason) }
                if (-not $postDockerOwnership.OutputsComplete) { throw 'The container did not produce all three allowlisted temporary outputs.' }
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                $assetsReceipt = Assert-TemporaryAssetsReceipt -Path $workState.AssetsPath
                $derivationReceipt = Assert-TemporaryDerivationReceipt -Path $workState.DerivationPath -ConfigPath $workState.ConfigPath
                if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { throw [System.TimeoutException]::new('deadline') }
                $admissionBase = $assetsReceipt.Value['admission_base']
                $assetSet = $assetsReceipt.Value['assets']
                $vitestAsset = $assetSet['vitest-config']
                if ($admissionBase['supplied_image_id'] -cne $BaseImageId -or $admissionBase['observed_image_id'] -cne $inspectProof.ObservedImageId -or
                    $admissionBase['observed_platform'] -cne $inspectProof.BoundPlatform -or -not $admissionBase['image_id_match'] -or
                    -not $admissionBase['platform_match'] -or -not $assetsReceipt.Value['accepted'] -or -not $assetsReceipt.Value['aggregate']['cap_valid']) {
                    throw 'Temporary assets does not preserve the host-proved bound admission identity.'
                }
                foreach ($assetName in @('go-module-cache', 'race-c-toolchain', 'frontend-node-modules', 'vitest-config', 'swag')) {
                    if (-not $assetSet[$assetName]['records_valid'] -or -not $assetSet[$assetName]['cap_valid'] -or -not $assetSet[$assetName]['accepted']) { throw 'A temporary asset is not structurally accepted.' }
                }
                if (-not $vitestAsset['derivation_reproducible'] -or -not $derivationReceipt.Value['derivation_reproducible'] -or
                    $derivationReceipt.Value['source_blobs_manifest_sha256'] -cne $SourceBlobsManifestSha256 -or
                    $vitestAsset['config_sha256'] -cne $derivationReceipt.ConfigSha256 -or
                    $vitestAsset['derivation_manifest_sha256'] -cne (Get-ByteSha256 -Bytes $derivationReceipt.DerivationBytes)) {
                    throw 'Temporary Vitest derivation does not preserve exact receipt and config byte bindings.'
                }
                foreach ($sourceKey in @('backend/go.mod','backend/go.sum','frontend/package.json','frontend/package-lock.json','frontend/vite.config.ts','frontend/src/test/setup.ts','frontend/tsconfig.json')) {
                    if ($vitestAsset['source_blob_sha1s'][$sourceKey] -cne $derivationReceipt.Value['source_blob_sha1s'][$sourceKey]) { throw 'Temporary Vitest source-hash maps disagree.' }
                }
                foreach ($versionKey in @('vitest', 'vite', 'typescript')) {
                    if ($vitestAsset['versions'][$versionKey] -cne $derivationReceipt.Value['versions'][$versionKey]) { throw 'Temporary Vitest version maps disagree.' }
                }
                $heldAssetsBytes = $assetsReceipt.Bytes
                $heldConfigBytes = $derivationReceipt.ConfigBytes
                $heldDerivationBytes = $derivationReceipt.DerivationBytes
                $publicationCandidate = $true
            }
            catch { $firstExitCode = if ($_.Exception -is [System.TimeoutException] -or [System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } }
        }
    }
    catch {
        if ($firstExitCode -eq 0) { $firstExitCode = if ([System.Diagnostics.Stopwatch]::GetTimestamp() -ge $deadline.EndTicks) { 124 } else { 70 } }
    }
    finally {
        $finallyEntered = $true
        if ($null -ne $workState) {
            if ($workState.InitializationComplete -and $null -ne $heldId) { $removeResult = Remove-HeldContainer -TaskPaths $taskPaths -WorkState $workState -Binding $binding -HeldId $heldId -Deadline $deadline }
            elseif ($createStarted) { $removeResult = [pscustomobject]@{ Attempted = $false; Removed = $false; ExitCode = 71; Reason = 'create-started-without-retained-id' } }
            if ($removeResult.ExitCode -ne 0 -and $firstExitCode -eq 0) { $firstExitCode = $removeResult.ExitCode }
            $retainReservation = $publicationCandidate -and $removeResult.Removed -and $firstExitCode -eq 0
            $workCleanup = Remove-ExactVerifiedWork -WorkState $workState -ContainerCleanupSatisfied $removeResult.Removed -RetainReservation $retainReservation -Deadline $deadline
            if ($workCleanup.ExitCode -ne 0 -and $firstExitCode -eq 0) { $firstExitCode = $workCleanup.ExitCode }
            if ($retainReservation -and $workCleanup.ExitCode -eq 0 -and $workCleanup.ReservationRetained -and $firstExitCode -eq 0) {
                $publicationResult = Publish-EligibleFinalEvidence -TaskPaths $taskPaths -WorkState $workState -HeldAssetsBytes $heldAssetsBytes -HeldConfigBytes $heldConfigBytes -HeldDerivationBytes $heldDerivationBytes -Deadline $deadline
                if ($publicationResult.ExitCode -ne 0 -and $firstExitCode -eq 0) { $firstExitCode = $publicationResult.ExitCode }
            }
            $publicationAttempted = $null -ne $publicationResult
            $stagingResolved = if ($publicationAttempted) { $publicationResult.StagingRemoved } else { $true }
            $reservationRemoved = $workCleanup.ReservationRemoved -or ($publicationAttempted -and $publicationResult.ReservationRemoved)
            $cleanupState = [pscustomobject]@{
                ContainerAttempted = $removeResult.Attempted
                ContainerRemoved = $removeResult.Removed
                WorkAttempted = $true
                WorkRemoved = $workCleanup.WorkRemoved
                ReservationAttempted = $true
                ReservationRemoved = $reservationRemoved
                StagingAttempted = $publicationAttempted
                StagingResolved = $stagingResolved
                PublicationStaging = $publicationStaging
                PublicationStagingOwner = $publicationStagingOwner
                FinalEvidenceAlreadyPreserved = ($publicationAttempted -and $publicationResult.FinalEvidencePreserved)
            }
            $receiptResult = Write-FinalHostReceipts -TaskPaths $taskPaths -WorkState $workState -BaseImageId $BaseImageId -HandoffManifestSha256 $HandoffManifestSha256 -SourceBlobsManifestSha256 $SourceBlobsManifestSha256 -InspectProof $inspectProof -ContainerName $containerName -CleanupState $cleanupState -StartedAt $startedAt -FirstExitCode $firstExitCode -Deadline $deadline
            if ($receiptResult.ExitCode -ne 0 -and $firstExitCode -eq 0) { $firstExitCode = $receiptResult.ExitCode }
        }
    }
    $status = if ($firstExitCode -eq 0) { 'STATIC' } else { 'FAIL' }
    return [pscustomobject]@{ ExitCode = $firstExitCode; Status = $status; Binding = 'BOUND'; FinallyEntered = $finallyEntered; HeldId = $heldId; ContainerRemoved = $removeResult.Removed; WorkCleanup = $workCleanup; Publication = $publicationResult; Receipts = $receiptResult; Ready = $false; Runtime = 'NOT_RUN' }
}

if ($args.Count -ne 0) {
    throw 'D0 accepts no arguments outside its exact eight mandatory inputs.'
}

return Invoke-LocalDependencyAdmission @PSBoundParameters
