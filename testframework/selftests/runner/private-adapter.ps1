function New-HubQaFixtureDependencies {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [pscustomobject]$FixtureCase
    )

    $closedSelectedTargetInspectRecord = $FixtureCase.LocalInspectRecord
    $closedInjectedResourceInventory = @($FixtureCase.InjectedResourceInventory)
    $closedSelectedOwnedResources = @($FixtureCase.SelectedOwnedResources)
    $closedCleanupOutcome = $FixtureCase.CleanupOutcome
    $closedSelectedOwnedTuples = @(
        foreach ($resource in $closedSelectedOwnedResources) {
            '{0}|{1}|{2}|{3}|{4}' -f @(
                $resource.RecordType
                $resource.Id
                $resource.Kind
                $resource.Name
                $resource.Project
            )
        }
    )

    $closedInspectResult = [pscustomobject][ordered]@{
        Verified               = $true
        ExitCode               = 0
        DeadlineLayer          = $null
        ImageId                = $closedSelectedTargetInspectRecord.ImageId
        Entrypoint             = @($closedSelectedTargetInspectRecord.Config.Entrypoint)
        EntrypointSha256       = $closedSelectedTargetInspectRecord.EntrypointSha256
        Cmd                    = @($closedSelectedTargetInspectRecord.Config.Cmd)
        CmdSha256              = $closedSelectedTargetInspectRecord.CmdSha256
        LabelsProjectionSha256 = $closedSelectedTargetInspectRecord.LabelsProjectionSha256
    }
    $closedRunResult = [pscustomobject][ordered]@{
        ExitCode        = 0
        TimedOut        = $false
        DeadlineLayer   = $null
        StreamCompleted = $true
        Stdout          = ''
        Stderr          = ''
        StdoutBytes     = [byte[]]@()
        StderrBytes     = [byte[]]@()
    }

    return [pscustomobject]@{
        InspectLockedImage = {
            return $closedInspectResult
        }.GetNewClosure()
        RunFixedTarget = {
            return $closedRunResult
        }.GetNewClosure()
        ListOwnedResources = {
            param(
                [Parameter(Mandatory)]
                [int64]$DeadlineTick
            )

            return $closedInjectedResourceInventory
        }.GetNewClosure()
        CleanupOwnedResources = {
            param(
                [Parameter(Mandatory)]
                [AllowEmptyCollection()]
                [object[]]$OwnedResources,

                [Parameter(Mandatory)]
                [int64]$DeadlineTick
            )

            if ($OwnedResources.Count -ne $closedSelectedOwnedResources.Count) {
                throw 'CleanupOwnedResources received an owned-set cardinality mismatch.'
            }

            $actualOwnedTuples = @()
            foreach ($resource in $OwnedResources) {
                if ($null -eq $resource -or $resource -isnot [pscustomobject]) {
                    throw 'CleanupOwnedResources requires typed resource records.'
                }
                $actualFields = @($resource.PSObject.Properties.Name)
                if (
                    $actualFields.Count -ne 5 -or
                    $actualFields -cnotcontains 'RecordType' -or
                    $actualFields -cnotcontains 'Id' -or
                    $actualFields -cnotcontains 'Kind' -or
                    $actualFields -cnotcontains 'Name' -or
                    $actualFields -cnotcontains 'Project'
                ) {
                    throw 'CleanupOwnedResources requires exact five-field typed resource records.'
                }
                if ($resource.RecordType -cne 'hub.qa.owned-resource/v1') {
                    throw 'CleanupOwnedResources rejects unselected resource records.'
                }

                $actualTuple = '{0}|{1}|{2}|{3}|{4}' -f @(
                    $resource.RecordType
                    $resource.Id
                    $resource.Kind
                    $resource.Name
                    $resource.Project
                )
                if ($closedSelectedOwnedTuples -cnotcontains $actualTuple) {
                    throw 'CleanupOwnedResources received an unselected resource tuple.'
                }
                if ($actualOwnedTuples -ccontains $actualTuple) {
                    throw 'CleanupOwnedResources received a duplicate resource tuple.'
                }
                $actualOwnedTuples += $actualTuple
            }

            return $closedCleanupOutcome
        }.GetNewClosure()
    }
}
