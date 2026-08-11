function New-HubQaFixtureDependencies {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [pscustomobject]$FixtureCase,

        [Parameter(Mandatory)]
        [ValidatePattern('^example\.invalid/[a-z0-9][a-z0-9._/-]*@sha256:[a-f0-9]{64}$')]
        [string]$PrevalidatedImageIdentity
    )
    $closedImageIdentity = [string]$PrevalidatedImageIdentity
    $closedInjectedResourceInventory = @($FixtureCase.InjectedResourceInventory)
    $closedSelectedOwnedResources = @($FixtureCase.SelectedOwnedResources)
    $closedCleanupOutcome = $FixtureCase.CleanupOutcome
    $closedSelectedOwnedTuples = @(
        foreach ($resource in $closedSelectedOwnedResources) {
            '{0}|{1}|{2}|{3}' -f @(
                $resource.RecordType
                $resource.Id
                $resource.Kind
                $resource.Ownership
            )
        }
    )

    $fixtureRunningProcess = [pscustomobject]@{
        Kind  = 'fixture-running-process'
        State = 'inert'
    }
    $fixtureMonotonicStartTick = [int64]0
    $fixedTargetIdentity = 'fixture-fixed-target'

    return [pscustomobject]@{
        InspectLockedImage = {
            return [pscustomobject]@{
                ImageIdentity = $closedImageIdentity
            }
        }.GetNewClosure()
        RunFixedTarget = {
            return [pscustomobject]@{
                RunningProcessDescriptor = $fixtureRunningProcess
                MonotonicStartTick        = $fixtureMonotonicStartTick
                FixedTargetIdentity       = $fixedTargetIdentity
            }
        }.GetNewClosure()
        ListOwnedResources = {
            return ,$closedInjectedResourceInventory
        }.GetNewClosure()
        CleanupOwnedResources = {
            param(
                [Parameter(Mandatory)]
                [AllowEmptyCollection()]
                [pscustomobject[]]$OwnedResources
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
                    $actualFields.Count -ne 4 -or
                    $actualFields -cnotcontains 'RecordType' -or
                    $actualFields -cnotcontains 'Id' -or
                    $actualFields -cnotcontains 'Kind' -or
                    $actualFields -cnotcontains 'Ownership'
                ) {
                    throw 'CleanupOwnedResources requires exact four-field typed resource records.'
                }
                if (
                    $resource.RecordType -cne 'hub.qa.fixture-owned-resource/v1' -or
                    $resource.Ownership -cne 'owned'
                ) {
                    throw 'CleanupOwnedResources rejects foreign or untyped resource records.'
                }

                $actualTuple = '{0}|{1}|{2}|{3}' -f @(
                    $resource.RecordType
                    $resource.Id
                    $resource.Kind
                    $resource.Ownership
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
