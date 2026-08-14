Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

function Get-HubQaEnvelopeUtf8 {
    return New-Object System.Text.UTF8Encoding($false, $true)
}

function Get-HubQaEnvelopeSha256 {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)
    $hash = [System.Security.Cryptography.SHA256]::Create().ComputeHash($Bytes)
    return 'sha256:' + ([System.BitConverter]::ToString($hash).Replace('-', '').ToLowerInvariant())
}

function Assert-HubQaEnvelopeKeys {
    param([Parameter(Mandatory = $true)][System.Collections.IDictionary]$Object, [Parameter(Mandatory = $true)][string[]]$Expected)
    $actual = @($Object.Keys)
    if ($actual.Count -ne $Expected.Count) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    for ($i = 0; $i -lt $Expected.Count; $i++) {
        if ([string]$actual[$i] -cne [string]$Expected[$i]) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    }
}

function ConvertFrom-HubQaEnvelopeCanonicalJson {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes, [switch]$AllowNonCanonical)

    if ($Bytes.Length -eq 0 -or ($Bytes.Length -ge 3 -and $Bytes[0] -eq 0xef -and $Bytes[1] -eq 0xbb -and $Bytes[2] -eq 0xbf)) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    try { $text = (Get-HubQaEnvelopeUtf8).GetString($Bytes) } catch { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $cursor = 0
    $length = $text.Length
    $skip = { while ($cursor -lt $length -and ([int][char]$text[$cursor] -in 0x20,0x09,0x0a,0x0d)) { $cursor++ } }.GetNewClosure()
    $parseString = $null
    $parseValue = $null
    $parseArray = $null
    $parseObject = $null
    $parseString = {
        if ($cursor -ge $length -or $text[$cursor] -ne '"') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        $cursor++
        $builder = New-Object System.Text.StringBuilder
        while ($cursor -lt $length) {
            $code = [int][char]$text[$cursor]
            $cursor++
            if ($code -eq 0x22) { return $builder.ToString() }
            if ($code -lt 0x20) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            if ($code -ne 0x5c) {
                if ($code -ge 0xd800 -and $code -le 0xdbff) {
                    if ($cursor -ge $length -or ([int][char]$text[$cursor]) -lt 0xdc00 -or ([int][char]$text[$cursor]) -gt 0xdfff) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                    [void]$builder.Append([char]$code); [void]$builder.Append($text[$cursor]); $cursor++; continue
                }
                if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                [void]$builder.Append([char]$code); continue
            }
            if ($cursor -ge $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            $escape = $text[$cursor]; $cursor++
            switch ($escape) {
                '"' { [void]$builder.Append('"'); continue }
                '\' { [void]$builder.Append('\'); continue }
                '/' { [void]$builder.Append('/'); continue }
                'b' { [void]$builder.Append([char]0x08); continue }
                'f' { [void]$builder.Append([char]0x0c); continue }
                'n' { [void]$builder.Append([char]0x0a); continue }
                'r' { [void]$builder.Append([char]0x0d); continue }
                't' { [void]$builder.Append([char]0x09); continue }
                'u' {
                    if (($cursor + 4) -gt $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                    $hex = $text.Substring($cursor, 4)
                    if ($hex -cnotmatch '^[0-9a-fA-F]{4}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                    $unit = [Convert]::ToInt32($hex, 16); $cursor += 4
                    if ($unit -ge 0xd800 -and $unit -le 0xdbff) {
                        if (($cursor + 6) -gt $length -or $text[$cursor] -ne '\' -or $text[$cursor + 1] -ne 'u') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                        $lowHex = $text.Substring($cursor + 2, 4)
                        if ($lowHex -cnotmatch '^[0-9a-fA-F]{4}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                        $low = [Convert]::ToInt32($lowHex, 16)
                        if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                        [void]$builder.Append([char]$unit); [void]$builder.Append([char]$low); $cursor += 6; continue
                    }
                    if ($unit -ge 0xdc00 -and $unit -le 0xdfff) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
                    [void]$builder.Append([char]$unit); continue
                }
                default { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            }
        }
        throw 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    }.GetNewClosure()
    $parseArray = {
        if ($cursor -ge $length -or $text[$cursor] -ne '[') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        $cursor++; . $skip
        $items = New-Object 'System.Collections.Generic.List[object]'
        if ($cursor -lt $length -and $text[$cursor] -eq ']') { $cursor++; return [pscustomobject][ordered]@{ Type='array'; Value=[object[]]$items.ToArray() } }
        while ($true) {
            [void]$items.Add((. $parseValue)); . $skip
            if ($cursor -ge $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            if ($text[$cursor] -eq ']') { $cursor++; break }
            if ($text[$cursor] -ne ',') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            $cursor++; . $skip
        }
        return [pscustomobject][ordered]@{ Type='array'; Value=[object[]]$items.ToArray() }
    }.GetNewClosure()
    $parseObject = {
        if ($cursor -ge $length -or $text[$cursor] -ne '{') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        $cursor++; . $skip
        $members = New-Object System.Collections.Specialized.OrderedDictionary
        if ($cursor -lt $length -and $text[$cursor] -eq '}') { $cursor++; return [pscustomobject][ordered]@{ Type='object'; Value=$members } }
        while ($true) {
            $name = . $parseString
            if ($members.Contains($name)) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            . $skip
            if ($cursor -ge $length -or $text[$cursor] -ne ':') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            $cursor++; . $skip; $members.Add($name, (. $parseValue)); . $skip
            if ($cursor -ge $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            if ($text[$cursor] -eq '}') { $cursor++; break }
            if ($text[$cursor] -ne ',') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            $cursor++; . $skip
            if ($cursor -ge $length -or $text[$cursor] -ne '"') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        }
        return [pscustomobject][ordered]@{ Type='object'; Value=$members }
    }.GetNewClosure()
    $parseValue = {
        . $skip
        if ($cursor -ge $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        switch ($text[$cursor]) {
            '{' { return (. $parseObject) }
            '[' { return (. $parseArray) }
            '"' { return [pscustomobject][ordered]@{ Type='string'; Value=(. $parseString) } }
            'n' { if (($cursor + 4) -le $length -and $text.Substring($cursor,4) -ceq 'null') { $cursor += 4; return [pscustomobject][ordered]@{ Type='null'; Value=$null } } }
            't' { if (($cursor + 4) -le $length -and $text.Substring($cursor,4) -ceq 'true') { $cursor += 4; return [pscustomobject][ordered]@{ Type='boolean'; Value=$true } } }
            'f' { if (($cursor + 5) -le $length -and $text.Substring($cursor,5) -ceq 'false') { $cursor += 5; return [pscustomobject][ordered]@{ Type='boolean'; Value=$false } } }
        }
        if ($AllowNonCanonical -and $text.Substring($cursor) -cmatch '^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?') {
            $number = $Matches[0]
            $cursor += $number.Length
            return [pscustomobject][ordered]@{ Type='number'; Value=$number }
        }
        if ($text[$cursor] -cmatch '[0-9]') {
            $start = $cursor
            if ($text[$cursor] -eq '0') { $cursor++ } else { while ($cursor -lt $length -and $text[$cursor] -cmatch '[0-9]') { $cursor++ } }
            $number = $text.Substring($start, $cursor - $start)
            if ($number.Length -gt 1 -and $number[0] -eq '0') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
            try { return [pscustomobject][ordered]@{ Type='integer'; Value=[Int64]::Parse($number, [Globalization.CultureInfo]::InvariantCulture) } } catch { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        }
        throw 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    }.GetNewClosure()
    $root = . $parseValue; . $skip
    if ($cursor -ne $length) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $serializeString = {
        param([string]$Value)
        $builder = New-Object System.Text.StringBuilder; [void]$builder.Append('"')
        foreach ($character in $Value.ToCharArray()) {
            $code = [int][char]$character
            switch ($code) {
                0x22 { [void]$builder.Append('\"'); continue }
                0x5c { [void]$builder.Append('\\'); continue }
                0x08 { [void]$builder.Append('\b'); continue }
                0x0c { [void]$builder.Append('\f'); continue }
                0x0a { [void]$builder.Append('\n'); continue }
                0x0d { [void]$builder.Append('\r'); continue }
                0x09 { [void]$builder.Append('\t'); continue }
            }
            if ($code -lt 0x20) { [void]$builder.Append(('\u00' + $code.ToString('X2'))); continue }
            [void]$builder.Append($character)
        }
        [void]$builder.Append('"'); return $builder.ToString()
    }.GetNewClosure()
    $serializeNode = $null
    $serializeNode = {
        param([object]$Node)
        switch ([string]$Node.Type) {
            'object' { $items = foreach ($key in @($Node.Value.Keys)) { (& $serializeString ([string]$key)) + ':' + (& $serializeNode $Node.Value[$key]) }; return '{' + ($items -join ',') + '}' }
            'array' { return '[' + ((@($Node.Value) | ForEach-Object { & $serializeNode $_ }) -join ',') + ']' }
            'string' { return & $serializeString ([string]$Node.Value) }
            'integer' { return ([Int64]$Node.Value).ToString([Globalization.CultureInfo]::InvariantCulture) }
            'boolean' { if ([bool]$Node.Value) { return 'true' }; return 'false' }
            'null' { return 'null' }
            default { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
        }
    }.GetNewClosure()
    $canonicalBytes = (Get-HubQaEnvelopeUtf8).GetBytes((& $serializeNode $root))
    if (-not $AllowNonCanonical -and -not ([System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($Bytes, $canonicalBytes))) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return $root
}

function ConvertTo-HubQaEnvelopeCanonicalJson {
    param([Parameter(Mandatory = $true)][object]$Value)
    $escape = {
        param([string]$Text)
        $builder = New-Object System.Text.StringBuilder; [void]$builder.Append('"')
        foreach ($character in $Text.ToCharArray()) {
            $code = [int][char]$character
            switch ($code) {
                0x22 { [void]$builder.Append('\"'); continue }
                0x5c { [void]$builder.Append('\\'); continue }
                0x08 { [void]$builder.Append('\b'); continue }
                0x0c { [void]$builder.Append('\f'); continue }
                0x0a { [void]$builder.Append('\n'); continue }
                0x0d { [void]$builder.Append('\r'); continue }
                0x09 { [void]$builder.Append('\t'); continue }
            }
            if ($code -lt 0x20) { [void]$builder.Append(('\u00' + $code.ToString('X2'))); continue }
            [void]$builder.Append($character)
        }
        [void]$builder.Append('"'); return $builder.ToString()
    }.GetNewClosure()
    $convert = $null
    $convert = {
        param([object]$Item)
        if ($null -eq $Item) { return 'null' }
        if ($Item -is [System.Collections.IDictionary]) { $members = foreach ($key in @($Item.Keys)) { (& $escape ([string]$key)) + ':' + (& $convert $Item[$key]) }; return '{' + ($members -join ',') + '}' }
        if ($Item -is [System.Collections.IEnumerable] -and -not ($Item -is [string])) { return '[' + ((@($Item) | ForEach-Object { & $convert $_ }) -join ',') + ']' }
        if ($Item -is [string]) { return & $escape $Item }
        if ($Item -is [bool]) { if ($Item) { return 'true' }; return 'false' }
        if ($Item -is [byte] -or $Item -is [int16] -or $Item -is [int32] -or $Item -is [int64] -or $Item -is [uint16] -or $Item -is [uint32]) { return ([Int64]$Item).ToString([Globalization.CultureInfo]::InvariantCulture) }
        throw 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    }.GetNewClosure()
    return & $convert $Value
}

# The canonical parser above deliberately retains scalar types.  Do not replace
# this validator with ConvertFrom-Json: it is the single closed, duplicate-aware
# value gate used by both producers and the later activation verifier.
function Get-HubQaEnvelopeNodeObject {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ([string]$Node.Type -cne 'object') { throw $Code }
    return [System.Collections.IDictionary]$Node.Value
}

function Get-HubQaEnvelopeNodeArray {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ([string]$Node.Type -cne 'array') { throw $Code }
    return @($Node.Value)
}

function Get-HubQaEnvelopeNodeString {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ([string]$Node.Type -cne 'string') { throw $Code }
    return [string]$Node.Value
}

function Get-HubQaEnvelopeNodeInteger {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ([string]$Node.Type -cne 'integer') { throw $Code }
    return [Int64]$Node.Value
}

function Assert-HubQaEnvelopeNodeNull {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ([string]$Node.Type -cne 'null') { throw $Code }
}

function Assert-HubQaEnvelopeUtc {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    $value = Get-HubQaEnvelopeNodeString -Node $Node -Code $Code
    if ($value -cnotmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,7})?Z$') { throw $Code }
    try { return [DateTime]::ParseExact($value, 'yyyy-MM-ddTHH:mm:ss.FFFFFFFZ', [Globalization.CultureInfo]::InvariantCulture, [Globalization.DateTimeStyles]::AssumeUniversal -bor [Globalization.DateTimeStyles]::AdjustToUniversal) } catch { throw $Code }
}

function Test-HubQaEnvelopeDigest { param([string]$Value) return $Value -cmatch '^sha256:[a-f0-9]{64}$' }
function Test-HubQaEnvelopeCommit { param([string]$Value) return $Value -cmatch '^[a-f0-9]{40}$' }
function Test-HubQaEnvelopeRepoPath { param([string]$Value) return $Value -cmatch '^(?!/)(?![A-Za-z]:)(?!.*(?:^|/)\.{1,2}(?:/|$))(?!.*//)[^\x00-\x1f\\]+$' }

function Convert-HubQaEnvelopeNodeToValue {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    switch ([string]$Node.Type) {
        'object' { $value = [ordered]@{}; foreach ($key in @($Node.Value.Keys)) { $value[[string]$key] = Convert-HubQaEnvelopeNodeToValue -Node $Node.Value[$key] -Code $Code }; return $value }
        'array' { return @($Node.Value | ForEach-Object { Convert-HubQaEnvelopeNodeToValue -Node $_ -Code $Code }) }
        'string' { return [string]$Node.Value }
        'integer' { return [Int64]$Node.Value }
        'boolean' { return [bool]$Node.Value }
        'null' { return $null }
        default { throw $Code }
    }
}

function Get-HubQaEnvelopeNodeCollectionDigest {
    param([Parameter(Mandatory = $true)][object]$Node, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    [byte[]]$bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value (Convert-HubQaEnvelopeNodeToValue -Node $Node -Code $Code)))
    return Get-HubQaEnvelopeSha256 -Bytes $bytes
}

function Assert-HubQaEnvelopePathRecordArray {
    param([Parameter(Mandatory = $true)][object]$Node, [Parameter(Mandatory = $true)][string]$StateKey, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    $entries = Get-HubQaEnvelopeNodeArray -Node $Node -Code $Code
    $previous = $null
    foreach ($entry in $entries) {
        $value = Get-HubQaEnvelopeNodeObject -Node $entry -Code $Code
        Assert-HubQaEnvelopeKeys -Object $value -Expected @('path',$StateKey,'blob_sha256')
        $path = Get-HubQaEnvelopeNodeString -Node $value['path'] -Code $Code
        $state = Get-HubQaEnvelopeNodeString -Node $value[$StateKey] -Code $Code
        $digest = Get-HubQaEnvelopeNodeString -Node $value['blob_sha256'] -Code $Code
        if (-not (Test-HubQaEnvelopeRepoPath -Value $path) -or $state -cnotmatch '^[A-Z?][A-Z0-9?]{0,7}$' -or -not (Test-HubQaEnvelopeDigest -Value $digest) -or ($null -ne $previous -and [string]::CompareOrdinal($previous, $path) -ge 0)) { throw $Code }
        $previous = $path
    }
    return $entries
}

function Assert-HubQaSnapshotEnvelopePayload {
    param(
        [Parameter(Mandatory = $true)][object]$Node,
        [Parameter(Mandatory = $true)][ValidateSet('snapshot','envelope')][string]$Kind,
        [Parameter(Mandatory = $true)][string]$ExpectedTask,
        [Parameter(Mandatory = $true)][string]$ExpectedRun,
        [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    )

    $value = Get-HubQaEnvelopeNodeObject -Node $Node -Code $Code
    if ($Kind -eq 'envelope') {
        Assert-HubQaEnvelopeKeys -Object $value -Expected @('schema','producer','envelope_id','contract','target','task','run','owner_nonce','repository_snapshot_sha256','report_root','limits','cleanup_contract','created_at_utc')
        if ((Get-HubQaEnvelopeNodeString -Node $value['schema'] -Code $Code) -cne 'hub.qa.run-envelope/v1' -or (Get-HubQaEnvelopeNodeString -Node $value['producer'] -Code $Code) -cne 'hub.qa.run-envelope-producer/v1' -or (Get-HubQaEnvelopeNodeString -Node $value['envelope_id'] -Code $Code) -cne ($ExpectedTask + '-envelope-' + $ExpectedRun) -or (Get-HubQaEnvelopeNodeString -Node $value['contract'] -Code $Code) -cne 'hub.qa-runner/v2' -or (Get-HubQaEnvelopeNodeString -Node $value['target'] -Code $Code) -cne 'g0-planning-selftest' -or (Get-HubQaEnvelopeNodeString -Node $value['task'] -Code $Code) -cne $ExpectedTask -or (Get-HubQaEnvelopeNodeString -Node $value['run'] -Code $Code) -cne $ExpectedRun -or (Get-HubQaEnvelopeNodeString -Node $value['owner_nonce'] -Code $Code) -cnotmatch '^[0-9a-f]{32}$' -or -not (Test-HubQaEnvelopeDigest -Value (Get-HubQaEnvelopeNodeString -Node $value['repository_snapshot_sha256'] -Code $Code)) -or (Get-HubQaEnvelopeNodeString -Node $value['report_root'] -Code $Code) -cne ('testframework/reports/' + $ExpectedTask + '-' + $ExpectedRun) -or (Get-HubQaEnvelopeNodeString -Node $value['cleanup_contract'] -Code $Code) -cne 'hub.qa.exact-cleanup/v1') { throw $Code }
        $limits = Get-HubQaEnvelopeNodeObject -Node $value['limits'] -Code $Code
        Assert-HubQaEnvelopeKeys -Object $limits -Expected @('command_seconds','target_seconds','inspect_seconds','outer_seconds','cleanup_seconds','orchestration_slack_seconds')
        if ((Get-HubQaEnvelopeNodeInteger -Node $limits['command_seconds'] -Code $Code) -ne 180 -or (Get-HubQaEnvelopeNodeInteger -Node $limits['target_seconds'] -Code $Code) -ne 780 -or (Get-HubQaEnvelopeNodeInteger -Node $limits['inspect_seconds'] -Code $Code) -ne 30 -or (Get-HubQaEnvelopeNodeInteger -Node $limits['outer_seconds'] -Code $Code) -ne 840 -or (Get-HubQaEnvelopeNodeInteger -Node $limits['cleanup_seconds'] -Code $Code) -ne 30 -or (Get-HubQaEnvelopeNodeInteger -Node $limits['orchestration_slack_seconds'] -Code $Code) -ne 30) { throw $Code }
        [void](Assert-HubQaEnvelopeUtc -Node $value['created_at_utc'] -Code $Code)
        return $value
    }

    Assert-HubQaEnvelopeKeys -Object $value -Expected @('schema','producer','snapshot_id','captured_for','task','run','envelope_id','role_claim','self_attested','source_revision','head','index','worktree','commits','ancestry','control','evidence','lifecycle')
    if ((Get-HubQaEnvelopeNodeString -Node $value['schema'] -Code $Code) -cne 'hub.qa.repository-snapshot/v1' -or (Get-HubQaEnvelopeNodeString -Node $value['producer'] -Code $Code) -cne 'hub.qa.repository-snapshot-producer/v1' -or (Get-HubQaEnvelopeNodeString -Node $value['snapshot_id'] -Code $Code) -cne ($ExpectedTask + '-snapshot-' + $ExpectedRun + '-000-G0_BOOTSTRAP') -or (Get-HubQaEnvelopeNodeString -Node $value['task'] -Code $Code) -cne $ExpectedTask -or (Get-HubQaEnvelopeNodeString -Node $value['run'] -Code $Code) -cne $ExpectedRun -or (Get-HubQaEnvelopeNodeString -Node $value['envelope_id'] -Code $Code) -cne ($ExpectedTask + '-envelope-' + $ExpectedRun) -or (Get-HubQaEnvelopeNodeString -Node $value['role_claim'] -Code $Code) -cne 'hub.qa-bootstrap/v1' -or [string]$value['self_attested'].Type -cne 'boolean' -or -not [bool]$value['self_attested'].Value) { throw $Code }
    $captured = Get-HubQaEnvelopeNodeObject -Node $value['captured_for'] -Code $Code
    Assert-HubQaEnvelopeKeys -Object $captured -Expected @('kind','ordinal','subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256')
    if ((Get-HubQaEnvelopeNodeString -Node $captured['kind'] -Code $Code) -cne 'G0_BOOTSTRAP' -or (Get-HubQaEnvelopeNodeInteger -Node $captured['ordinal'] -Code $Code) -ne 0) { throw $Code }
    foreach ($field in @('subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256')) { Assert-HubQaEnvelopeNodeNull -Node $captured[$field] -Code $Code }
    $source = Get-HubQaEnvelopeNodeObject -Node $value['source_revision'] -Code $Code
    Assert-HubQaEnvelopeKeys -Object $source -Expected @('commit','tree','captured_at_utc')
    if ((Get-HubQaEnvelopeNodeString -Node $source['commit'] -Code $Code) -cne '61b38573682d7c01e6d9381babc5085537d7990f' -or -not (Test-HubQaEnvelopeCommit -Value (Get-HubQaEnvelopeNodeString -Node $source['tree'] -Code $Code))) { throw $Code }; [void](Assert-HubQaEnvelopeUtc -Node $source['captured_at_utc'] -Code $Code)
    $head = Get-HubQaEnvelopeNodeObject -Node $value['head'] -Code $Code
    Assert-HubQaEnvelopeKeys -Object $head -Expected @('commit','tree','event')
    if (-not (Test-HubQaEnvelopeCommit -Value (Get-HubQaEnvelopeNodeString -Node $head['commit'] -Code $Code)) -or -not (Test-HubQaEnvelopeCommit -Value (Get-HubQaEnvelopeNodeString -Node $head['tree'] -Code $Code))) { throw $Code }
    $event = Get-HubQaEnvelopeNodeObject -Node $head['event'] -Code $Code; Assert-HubQaEnvelopeKeys -Object $event -Expected @('event_id','event_type','captured_at_utc'); foreach ($field in @('event_id','event_type','captured_at_utc')) { Assert-HubQaEnvelopeNodeNull -Node $event[$field] -Code $Code }
    foreach ($pair in @(@('index','paths','status'),@('worktree','scoped_paths','state'))) {
        $collection = Get-HubQaEnvelopeNodeObject -Node $value[$pair[0]] -Code $Code; Assert-HubQaEnvelopeKeys -Object $collection -Expected @($pair[1],'count','sha256')
        $entries = Assert-HubQaEnvelopePathRecordArray -Node $collection[$pair[1]] -StateKey $pair[2] -Code $Code
        if ((Get-HubQaEnvelopeNodeInteger -Node $collection['count'] -Code $Code) -ne $entries.Count -or (Get-HubQaEnvelopeNodeString -Node $collection['sha256'] -Code $Code) -cne (Get-HubQaEnvelopeNodeCollectionDigest -Node $collection[$pair[1]] -Code $Code)) { throw $Code }
    }
    $commits = Get-HubQaEnvelopeNodeArray -Node $value['commits'] -Code $Code
    if ($commits.Count -ne 1) { throw $Code }
    $commit = Get-HubQaEnvelopeNodeObject -Node $commits[0] -Code $Code; Assert-HubQaEnvelopeKeys -Object $commit -Expected @('commit','parent','tree','changed_paths')
    if ((Get-HubQaEnvelopeNodeString -Node $commit['commit'] -Code $Code) -cne (Get-HubQaEnvelopeNodeString -Node $head['commit'] -Code $Code) -or (Get-HubQaEnvelopeNodeString -Node $commit['tree'] -Code $Code) -cne (Get-HubQaEnvelopeNodeString -Node $head['tree'] -Code $Code) -or -not (Test-HubQaEnvelopeCommit -Value (Get-HubQaEnvelopeNodeString -Node $commit['parent'] -Code $Code))) { throw $Code }; [void](Assert-HubQaEnvelopePathRecordArray -Node $commit['changed_paths'] -StateKey 'status' -Code $Code)
    $ancestry = Get-HubQaEnvelopeNodeObject -Node $value['ancestry'] -Code $Code; Assert-HubQaEnvelopeKeys -Object $ancestry -Expected @('edges'); $edges = Get-HubQaEnvelopeNodeArray -Node $ancestry['edges'] -Code $Code
    if ($edges.Count -ne 1) { throw $Code }; $edge = Get-HubQaEnvelopeNodeObject -Node $edges[0] -Code $Code; Assert-HubQaEnvelopeKeys -Object $edge -Expected @('child','parent')
    if ((Get-HubQaEnvelopeNodeString -Node $edge['child'] -Code $Code) -cne (Get-HubQaEnvelopeNodeString -Node $commit['commit'] -Code $Code) -or (Get-HubQaEnvelopeNodeString -Node $edge['parent'] -Code $Code) -cne (Get-HubQaEnvelopeNodeString -Node $commit['parent'] -Code $Code)) { throw $Code }
    $control = Get-HubQaEnvelopeNodeObject -Node $value['control'] -Code $Code; Assert-HubQaEnvelopeKeys -Object $control -Expected @('g0_plan_sha256','bootstrap_contract_sha256','runner_contract_sha256','dependency_bundle_manifest_sha256','source_policy_sha256','build_manifest_sha256','images_lock_sha256')
    foreach ($field in @('g0_plan_sha256','bootstrap_contract_sha256','runner_contract_sha256','dependency_bundle_manifest_sha256','source_policy_sha256','build_manifest_sha256')) { if (-not (Test-HubQaEnvelopeDigest -Value (Get-HubQaEnvelopeNodeString -Node $control[$field] -Code $Code))) { throw $Code } }; Assert-HubQaEnvelopeNodeNull -Node $control['images_lock_sha256'] -Code $Code
    $evidence = Get-HubQaEnvelopeNodeObject -Node $value['evidence'] -Code $Code; Assert-HubQaEnvelopeKeys -Object $evidence -Expected @('positive_green_path','positive_green_sha256','candidate_lock_path','candidate_lock_sha256','cleanup_path','cleanup_sha256')
    foreach ($field in @('positive_green_path','candidate_lock_path','cleanup_path')) { if (-not (Test-HubQaEnvelopeRepoPath -Value (Get-HubQaEnvelopeNodeString -Node $evidence[$field] -Code $Code))) { throw $Code } }; foreach ($field in @('positive_green_sha256','candidate_lock_sha256','cleanup_sha256')) { if (-not (Test-HubQaEnvelopeDigest -Value (Get-HubQaEnvelopeNodeString -Node $evidence[$field] -Code $Code))) { throw $Code } }
    $lifecycle = Get-HubQaEnvelopeNodeObject -Node $value['lifecycle'] -Code $Code; Assert-HubQaEnvelopeKeys -Object $lifecycle -Expected @('state','started_at_utc','finished_at_utc')
    if ((Get-HubQaEnvelopeNodeString -Node $lifecycle['state'] -Code $Code) -cne 'CAPTURED') { throw $Code }; $started = Assert-HubQaEnvelopeUtc -Node $lifecycle['started_at_utc'] -Code $Code; $finished = Assert-HubQaEnvelopeUtc -Node $lifecycle['finished_at_utc'] -Code $Code; if ($started -gt $finished) { throw $Code }
    return $value
}

function Get-HubQaEnvelopeRepositoryRoot {
    $root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
    return Resolve-HubQaEnvelopeSafeDirectory -Path $root -Root $root
}

function Resolve-HubQaEnvelopeSafeDirectory {
    param([Parameter(Mandatory = $true)][string]$Path, [Parameter(Mandatory = $true)][string]$Root, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    try {
        $rootFull = [IO.Path]::GetFullPath($Root).TrimEnd('\\')
        $full = [IO.Path]::GetFullPath($Path).TrimEnd('\\')
        if ($full -cne $rootFull -and -not $full.StartsWith($rootFull + '\\', [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
        $item = Get-Item -LiteralPath $full -Force -ErrorAction Stop
        if (-not $item.PSIsContainer) { throw $Code }
        $walk = $item
        while ($null -ne $walk) {
            $walk.Refresh()
            if (-not $walk.Exists -or (($walk.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0)) { throw $Code }
            if ($walk.FullName.TrimEnd('\\') -ceq $rootFull) { return $item.FullName.TrimEnd('\\') }
            $walk = $walk.Parent
        }
    } catch { throw $Code }
    throw $Code
}

function Initialize-HubQaEnvelopeWin32FileApis {
    if ($null -ne ('HubQaEnvelopeWin32FileNative' -as [type])) { return }
    Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Text;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;
[StructLayout(LayoutKind.Sequential)] public struct HubQaEnvelopeByHandleFileInformation {
 public uint FileAttributes; public System.Runtime.InteropServices.ComTypes.FILETIME CreationTime; public System.Runtime.InteropServices.ComTypes.FILETIME LastAccessTime; public System.Runtime.InteropServices.ComTypes.FILETIME LastWriteTime; public uint VolumeSerialNumber; public uint FileSizeHigh; public uint FileSizeLow; public uint NumberOfLinks; public uint FileIndexHigh; public uint FileIndexLow;
}
public static class HubQaEnvelopeWin32FileNative {
 public const uint GENERIC_READ=0x80000000, GENERIC_WRITE=0x40000000, SYNCHRONIZE=0x00100000, FILE_READ_ATTRIBUTES=0x80, FILE_SHARE_READ=1, FILE_SHARE_WRITE=2, CREATE_NEW=1, OPEN_EXISTING=3, FILE_ATTRIBUTE_DIRECTORY=0x10, FILE_ATTRIBUTE_NORMAL=0x80, FILE_ATTRIBUTE_REPARSE_POINT=0x400, FILE_FLAG_BACKUP_SEMANTICS=0x02000000, FILE_FLAG_OPEN_REPARSE_POINT=0x00200000;
 public const uint NT_FILE_OPEN=1, NT_FILE_CREATE=2, NT_FILE_OPEN_IF=3, NT_FILE_DIRECTORY_FILE=0x1, NT_FILE_SYNCHRONOUS_IO_NONALERT=0x20, NT_FILE_NON_DIRECTORY_FILE=0x40, NT_FILE_OPEN_REPARSE_POINT=0x00200000;
 private const uint OBJ_CASE_INSENSITIVE=0x40;
 [StructLayout(LayoutKind.Sequential)] private struct HubQaUnicodeString { public ushort Length; public ushort MaximumLength; public IntPtr Buffer; }
 [StructLayout(LayoutKind.Sequential)] private struct HubQaObjectAttributes { public int Length; public IntPtr RootDirectory; public IntPtr ObjectName; public uint Attributes; public IntPtr SecurityDescriptor; public IntPtr SecurityQualityOfService; }
 [StructLayout(LayoutKind.Sequential)] private struct HubQaIoStatusBlock { public IntPtr Status; public UIntPtr Information; }
 [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true, EntryPoint="CreateFileW")] public static extern SafeFileHandle CreateFileW(string path,uint access,uint share,IntPtr security,uint disposition,uint flags,IntPtr template);
 [DllImport("kernel32.dll", SetLastError=true)] [return:MarshalAs(UnmanagedType.Bool)] public static extern bool GetFileInformationByHandle(SafeFileHandle handle,out HubQaEnvelopeByHandleFileInformation information);
 [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)] public static extern uint GetFinalPathNameByHandle(SafeFileHandle handle,StringBuilder path,uint length,uint flags);
 [DllImport("ntdll.dll")] private static extern int NtCreateFile(out SafeFileHandle fileHandle,uint desiredAccess,ref HubQaObjectAttributes objectAttributes,out HubQaIoStatusBlock ioStatusBlock,IntPtr allocationSize,uint fileAttributes,uint shareAccess,uint createDisposition,uint createOptions,IntPtr eaBuffer,uint eaLength);
 [DllImport("ntdll.dll")] private static extern uint RtlNtStatusToDosError(int status);
 public static SafeFileHandle CreateRelative(SafeFileHandle parent,string name,uint desiredAccess,uint shareAccess,uint createDisposition,uint createOptions) {
  if (parent == null || parent.IsInvalid || parent.IsClosed) throw new ArgumentException("invalid parent");
  if (String.IsNullOrEmpty(name) || name.IndexOf('\\') >= 0 || name.IndexOf('/') >= 0) throw new ArgumentException("invalid relative name");
  int byteLength=Encoding.Unicode.GetByteCount(name);
  if (byteLength > UInt16.MaxValue-2) throw new ArgumentException("relative name too long");
  IntPtr buffer=IntPtr.Zero, unicodePointer=IntPtr.Zero;
  bool parentReference=false;
  try {
   buffer=Marshal.StringToHGlobalUni(name);
   HubQaUnicodeString unicode=new HubQaUnicodeString { Length=(ushort)byteLength, MaximumLength=(ushort)(byteLength+2), Buffer=buffer };
   unicodePointer=Marshal.AllocHGlobal(Marshal.SizeOf(typeof(HubQaUnicodeString)));
   Marshal.StructureToPtr(unicode,unicodePointer,false);
   parent.DangerousAddRef(ref parentReference);
   HubQaObjectAttributes attributes=new HubQaObjectAttributes { Length=Marshal.SizeOf(typeof(HubQaObjectAttributes)), RootDirectory=parent.DangerousGetHandle(), ObjectName=unicodePointer, Attributes=OBJ_CASE_INSENSITIVE, SecurityDescriptor=IntPtr.Zero, SecurityQualityOfService=IntPtr.Zero };
   HubQaIoStatusBlock statusBlock;
   SafeFileHandle result;
   uint fileAttributes=(createOptions & NT_FILE_DIRECTORY_FILE)!=0 ? FILE_ATTRIBUTE_DIRECTORY : FILE_ATTRIBUTE_NORMAL;
   int status=NtCreateFile(out result,desiredAccess,ref attributes,out statusBlock,IntPtr.Zero,fileAttributes,shareAccess,createDisposition,createOptions,IntPtr.Zero,0);
   if (status < 0) {
    if (result != null) result.Dispose();
    throw new Win32Exception((int)RtlNtStatusToDosError(status));
   }
   if (result == null || result.IsInvalid) {
    if (result != null) result.Dispose();
    throw new Win32Exception("NtCreateFile returned an invalid handle");
   }
   return result;
  } finally {
   if (parentReference) parent.DangerousRelease();
   if (unicodePointer != IntPtr.Zero) Marshal.FreeHGlobal(unicodePointer);
   if (buffer != IntPtr.Zero) Marshal.FreeHGlobal(buffer);
  }
 }
}
'@
}

function ConvertFrom-HubQaEnvelopeFinalPath {
    param([Parameter(Mandatory = $true)][string]$Path)
    if ($Path.StartsWith('\\?\UNC\', [StringComparison]::OrdinalIgnoreCase)) { return '\\' + $Path.Substring(8) }
    if ($Path.StartsWith('\\?\', [StringComparison]::Ordinal)) { return $Path.Substring(4) }
    return $Path
}

function Get-HubQaEnvelopeHandleFinalPath {
    param([Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    [int]$capacity = 1024
    while ($true) {
        $buffer = New-Object Text.StringBuilder $capacity
        [uint32]$written = [HubQaEnvelopeWin32FileNative]::GetFinalPathNameByHandle($Handle, $buffer, [uint32]$capacity, [uint32]0)
        if ($written -eq 0 -or $capacity -gt 32768) { throw $Code }
        if ($written -lt $capacity) { return ConvertFrom-HubQaEnvelopeFinalPath -Path $buffer.ToString() }
        $capacity = [int]$written + 1
    }
}

function Get-HubQaEnvelopeHandleIdentity {
    param([Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    $info = New-Object HubQaEnvelopeByHandleFileInformation
    if (-not [HubQaEnvelopeWin32FileNative]::GetFileInformationByHandle($Handle, [ref]$info)) { throw $Code }
    return [pscustomobject]@{ Attributes=[uint32]$info.FileAttributes; Volume=[uint32]$info.VolumeSerialNumber; IndexHigh=[uint32]$info.FileIndexHigh; IndexLow=[uint32]$info.FileIndexLow; Size=[Int64]((([uint64]$info.FileSizeHigh -shl 32) -bor [uint64]$info.FileSizeLow)); WriteHigh=[Int32]$info.LastWriteTime.dwHighDateTime; WriteLow=[uint32]$info.LastWriteTime.dwLowDateTime }
}

function Assert-HubQaEnvelopeSameHandleIdentity {
    param([Parameter(Mandatory = $true)][object]$Actual, [Parameter(Mandatory = $true)][object]$Expected, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    foreach ($name in @('Attributes','Volume','IndexHigh','IndexLow','Size','WriteHigh','WriteLow')) { if ($Actual.$name -ne $Expected.$name) { throw $Code } }
}

function Assert-HubQaEnvelopeSameStableFileIdentity {
    param([Parameter(Mandatory = $true)][object]$Actual, [Parameter(Mandatory = $true)][object]$Expected, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    foreach ($name in @('Volume','IndexHigh','IndexLow','Size')) { if ($Actual.$name -ne $Expected.$name) { throw $Code } }
}

function Test-HubQaEnvelopeSafePathComponent {
    param([Parameter(Mandatory = $true)][string]$Value)
    if ([string]::IsNullOrEmpty($Value) -or $Value -ceq '.' -or $Value -ceq '..' -or $Value.EndsWith('.') -or $Value.EndsWith(' ') -or $Value.IndexOfAny([char[]]'<>:"/\|?*') -ge 0) { return $false }
    foreach ($character in $Value.ToCharArray()) { if ([int]$character -lt 32) { return $false } }
    if ($Value -cmatch '^(?i:(?:con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\..*)?)$') { return $false }
    return $true
}

function Assert-HubQaEnvelopeHeldDirectoryChain {
    param([Parameter(Mandatory = $true)][object]$Chain, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ($null -eq $Chain.Entries -or @($Chain.Entries).Count -lt 1) { throw $Code }
    foreach ($entry in @($Chain.Entries)) {
        $handle = $entry.Handle
        if ($null -eq $handle -or $handle.IsInvalid -or $handle.IsClosed) { throw $Code }
        $actual = Get-HubQaEnvelopeHandleIdentity -Handle $handle -Code $Code
        if (($actual.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or ($actual.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) { throw $Code }
        foreach ($field in @('Volume','IndexHigh','IndexLow')) { if ($actual.$field -ne $entry.Identity.$field) { throw $Code } }
        $actualPath = [IO.Path]::GetFullPath((Get-HubQaEnvelopeHandleFinalPath -Handle $handle -Code $Code)).TrimEnd('\')
        if (-not [string]::Equals($actualPath, [string]$entry.ExpectedPath, [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
    }
}

function Close-HubQaEnvelopeHeldDirectoryChain {
    param([object]$Chain)
    if ($null -eq $Chain -or $null -eq $Chain.Entries) { return }
    $entries = @($Chain.Entries)
    for ($index = $entries.Count - 1; $index -ge 0; $index--) {
        $handle = $entries[$index].Handle
        if ($null -ne $handle -and -not $handle.IsClosed) { $handle.Dispose() }
    }
}

function Open-HubQaEnvelopeHeldDirectoryPath {
    param([Parameter(Mandatory = $true)][string]$Root, [Parameter(Mandatory = $true)][AllowEmptyString()][string]$RelativeDirectoryPath, [switch]$CreateLeaf, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    Initialize-HubQaEnvelopeWin32FileApis
    $parts = @()
    if ([string]::IsNullOrEmpty($RelativeDirectoryPath)) {
        if ($CreateLeaf) { throw $Code }
    } else {
        if (-not (Test-HubQaEnvelopeRepoPath -Value $RelativeDirectoryPath)) { throw $Code }
        $parts = @($RelativeDirectoryPath -split '/')
    }
    foreach ($part in $parts) { if (-not (Test-HubQaEnvelopeSafePathComponent -Value $part)) { throw $Code } }
    $rootPath = [IO.Path]::GetFullPath($Root).TrimEnd('\')
    $entries = New-Object Collections.ArrayList
    try {
        $rootHandle = [HubQaEnvelopeWin32FileNative]::CreateFileW($rootPath, ([HubQaEnvelopeWin32FileNative]::FILE_READ_ATTRIBUTES -bor [HubQaEnvelopeWin32FileNative]::SYNCHRONIZE), ([HubQaEnvelopeWin32FileNative]::FILE_SHARE_READ -bor [HubQaEnvelopeWin32FileNative]::FILE_SHARE_WRITE), [IntPtr]::Zero, [HubQaEnvelopeWin32FileNative]::OPEN_EXISTING, ([HubQaEnvelopeWin32FileNative]::FILE_FLAG_BACKUP_SEMANTICS -bor [HubQaEnvelopeWin32FileNative]::FILE_FLAG_OPEN_REPARSE_POINT), [IntPtr]::Zero)
        if ($null -eq $rootHandle -or $rootHandle.IsInvalid) { if ($null -ne $rootHandle) { $rootHandle.Dispose() }; throw $Code }
        $rootEntry = [pscustomobject][ordered]@{ Handle=$rootHandle; Identity=$null; ExpectedPath=$null }
        [void]$entries.Add($rootEntry)
        $rootIdentity = Get-HubQaEnvelopeHandleIdentity -Handle $rootHandle -Code $Code
        if (($rootIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or ($rootIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) { throw $Code }
        $physicalRoot = [IO.Path]::GetFullPath((Get-HubQaEnvelopeHandleFinalPath -Handle $rootHandle -Code $Code)).TrimEnd('\')
        $rootEntry.Identity = $rootIdentity
        $rootEntry.ExpectedPath = $physicalRoot
        $parentHandle = $rootHandle
        $parentPhysicalPath = $physicalRoot
        for ($index = 0; $index -lt $parts.Count; $index++) {
            $component = [string]$parts[$index]
            $disposition = [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN
            if ($CreateLeaf -and $index -eq $parts.Count - 1) { $disposition = [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN_IF }
            $directoryHandle = [HubQaEnvelopeWin32FileNative]::CreateRelative($parentHandle, $component, ([HubQaEnvelopeWin32FileNative]::FILE_READ_ATTRIBUTES -bor [HubQaEnvelopeWin32FileNative]::SYNCHRONIZE), ([HubQaEnvelopeWin32FileNative]::FILE_SHARE_READ -bor [HubQaEnvelopeWin32FileNative]::FILE_SHARE_WRITE), $disposition, ([HubQaEnvelopeWin32FileNative]::NT_FILE_DIRECTORY_FILE -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_SYNCHRONOUS_IO_NONALERT -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN_REPARSE_POINT))
            $directoryEntry = [pscustomobject][ordered]@{ Handle=$directoryHandle; Identity=$null; ExpectedPath=$null }
            [void]$entries.Add($directoryEntry)
            $directoryIdentity = Get-HubQaEnvelopeHandleIdentity -Handle $directoryHandle -Code $Code
            if (($directoryIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or ($directoryIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) { throw $Code }
            $expectedDirectoryPath = [IO.Path]::GetFullPath((Join-Path $parentPhysicalPath $component)).TrimEnd('\')
            $actualDirectoryPath = [IO.Path]::GetFullPath((Get-HubQaEnvelopeHandleFinalPath -Handle $directoryHandle -Code $Code)).TrimEnd('\')
            if (-not [string]::Equals($actualDirectoryPath, $expectedDirectoryPath, [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
            $directoryEntry.Identity = $directoryIdentity
            $directoryEntry.ExpectedPath = $expectedDirectoryPath
            $parentHandle = $directoryHandle
            $parentPhysicalPath = $expectedDirectoryPath
        }
        $entryArray = [object[]]($entries.ToArray())
        $chain = [pscustomobject][ordered]@{ RootPath=$physicalRoot; RelativePath=$RelativeDirectoryPath; DirectoryPath=$parentPhysicalPath; DirectoryHandle=$parentHandle; Entries=$entryArray }
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $chain -Code $Code
        return $chain
    } catch {
        $partial = [pscustomobject]@{ Entries=[object[]]($entries.ToArray()) }
        Close-HubQaEnvelopeHeldDirectoryChain -Chain $partial
        throw $Code
    }
}

function Open-HubQaEnvelopeHeldDirectoryChain {
    param([Parameter(Mandatory = $true)][string]$Root, [Parameter(Mandatory = $true)][string]$RelativePath, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if (-not (Test-HubQaEnvelopeRepoPath -Value $RelativePath)) { throw $Code }
    $parts = @($RelativePath -split '/')
    if ($parts.Count -lt 1) { throw $Code }
    foreach ($part in $parts) { if (-not (Test-HubQaEnvelopeSafePathComponent -Value $part)) { throw $Code } }
    $leafName = [string]$parts[$parts.Count - 1]
    $directoryRelativePath = ''
    if ($parts.Count -gt 1) { $directoryRelativePath = [string](@($parts[0..($parts.Count - 2)]) -join '/') }
    $directoryChain = Open-HubQaEnvelopeHeldDirectoryPath -Root $Root -RelativeDirectoryPath $directoryRelativePath -Code $Code
    try {
        $fullPath = [IO.Path]::GetFullPath((Join-Path ([string]$directoryChain.DirectoryPath) $leafName))
        $chain = [pscustomobject][ordered]@{ RootPath=[string]$directoryChain.RootPath; RelativePath=$RelativePath; FullPath=$fullPath; LeafName=$leafName; ParentHandle=$directoryChain.DirectoryHandle; Entries=[object[]]($directoryChain.Entries) }
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $chain -Code $Code
        return $chain
    } catch {
        Close-HubQaEnvelopeHeldDirectoryChain -Chain $directoryChain
        throw $Code
    }
}

function Resolve-HubQaEnvelopeSafeContainedFilePath {
    param([Parameter(Mandatory = $true)][string]$Root, [Parameter(Mandatory = $true)][string]$RelativePath, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if (-not (Test-HubQaEnvelopeRepoPath -Value $RelativePath)) { throw $Code }
    $rootFull = Resolve-HubQaEnvelopeSafeDirectory -Path $Root -Root $Root -Code $Code
    $full = [IO.Path]::GetFullPath((Join-Path $rootFull $RelativePath.Replace('/', '\\')))
    if (-not $full.StartsWith($rootFull + '\\', [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
    $parent = Split-Path -Parent $full
    [void](Resolve-HubQaEnvelopeSafeDirectory -Path $parent -Root $rootFull -Code $Code)
    return $full
}

function Get-HubQaEnvelopeSafeContainedBytes {
    param([Parameter(Mandatory = $true)][string]$Root, [Parameter(Mandatory = $true)][string]$RelativePath, [Int64]$MaximumBytes = 67108864, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK', [object]$HeldDirectoryChain)
    Initialize-HubQaEnvelopeWin32FileApis
    if ($null -eq $HeldDirectoryChain) {
        $full = Resolve-HubQaEnvelopeSafeContainedFilePath -Root $Root -RelativePath $RelativePath -Code $Code
        $rootFull = Resolve-HubQaEnvelopeSafeDirectory -Path $Root -Root $Root -Code $Code
        $handle = [HubQaEnvelopeWin32FileNative]::CreateFileW($full, [HubQaEnvelopeWin32FileNative]::GENERIC_READ, [HubQaEnvelopeWin32FileNative]::FILE_SHARE_READ, [IntPtr]::Zero, [HubQaEnvelopeWin32FileNative]::OPEN_EXISTING, [HubQaEnvelopeWin32FileNative]::FILE_FLAG_OPEN_REPARSE_POINT, [IntPtr]::Zero)
    } else {
        if ([string]$HeldDirectoryChain.RelativePath -cne $RelativePath) { throw $Code }
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $HeldDirectoryChain -Code $Code
        $full = [string]$HeldDirectoryChain.FullPath
        $rootFull = [string]$HeldDirectoryChain.RootPath
        $handle = [HubQaEnvelopeWin32FileNative]::CreateRelative($HeldDirectoryChain.ParentHandle, [string]$HeldDirectoryChain.LeafName, ([HubQaEnvelopeWin32FileNative]::GENERIC_READ -bor [HubQaEnvelopeWin32FileNative]::SYNCHRONIZE), [HubQaEnvelopeWin32FileNative]::FILE_SHARE_READ, [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN, ([HubQaEnvelopeWin32FileNative]::NT_FILE_NON_DIRECTORY_FILE -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_SYNCHRONOUS_IO_NONALERT -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN_REPARSE_POINT))
    }
    if ($null -eq $handle -or $handle.IsInvalid) { if ($null -ne $handle) { $handle.Dispose() }; throw $Code }
    try {
        $before = Get-HubQaEnvelopeHandleIdentity -Handle $handle -Code $Code
        if (($before.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -ne 0 -or ($before.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0 -or $before.Size -lt 0 -or $before.Size -gt $MaximumBytes) { throw $Code }
        $final = Get-HubQaEnvelopeHandleFinalPath -Handle $handle -Code $Code
        $finalFull = [IO.Path]::GetFullPath($final)
        if ($null -eq $HeldDirectoryChain) {
            if (-not $finalFull.StartsWith($rootFull + '\\', [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
        } elseif (-not [string]::Equals($finalFull, [IO.Path]::GetFullPath($full), [StringComparison]::OrdinalIgnoreCase) -or -not $finalFull.StartsWith($rootFull + '\\', [StringComparison]::OrdinalIgnoreCase)) { throw $Code }
        $stream = New-Object IO.FileStream($handle, [IO.FileAccess]::Read, 65536, $false)
        try { $memory = New-Object IO.MemoryStream; try { $stream.CopyTo($memory); [byte[]]$bytes = $memory.ToArray(); Assert-HubQaEnvelopeSameHandleIdentity -Actual (Get-HubQaEnvelopeHandleIdentity -Handle $handle -Code $Code) -Expected $before -Code $Code } finally { $memory.Dispose() } } finally { $stream.Dispose() }
        if ($bytes.Length -ne $before.Size) { throw $Code }
        if ($null -ne $HeldDirectoryChain) { Assert-HubQaEnvelopeHeldDirectoryChain -Chain $HeldDirectoryChain -Code $Code }
        return [pscustomobject][ordered]@{ Bytes=$bytes; Identity=$before; FullPath=$full }
    } finally { if (-not $handle.IsClosed) { $handle.Dispose() } }
}

function Write-HubQaEnvelopeSafeCreateNew {
    param([Parameter(Mandatory = $true)][string]$Root, [Parameter(Mandatory = $true)][string]$RelativePath, [Parameter(Mandatory = $true)][byte[]]$Bytes, [string]$Code = 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    Initialize-HubQaEnvelopeWin32FileApis
    $chain = Open-HubQaEnvelopeHeldDirectoryChain -Root $Root -RelativePath $RelativePath -Code $Code
    try {
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $chain -Code $Code
        $handle = [HubQaEnvelopeWin32FileNative]::CreateRelative($chain.ParentHandle, [string]$chain.LeafName, ([HubQaEnvelopeWin32FileNative]::GENERIC_WRITE -bor [HubQaEnvelopeWin32FileNative]::SYNCHRONIZE), [uint32]0, [HubQaEnvelopeWin32FileNative]::NT_FILE_CREATE, ([HubQaEnvelopeWin32FileNative]::NT_FILE_NON_DIRECTORY_FILE -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_SYNCHRONOUS_IO_NONALERT -bor [HubQaEnvelopeWin32FileNative]::NT_FILE_OPEN_REPARSE_POINT))
        if ($null -eq $handle -or $handle.IsInvalid) { if ($null -ne $handle) { $handle.Dispose() }; throw $Code }
        try {
            $final = Get-HubQaEnvelopeHandleFinalPath -Handle $handle -Code $Code
            $expectedFull = [IO.Path]::GetFullPath([string]$chain.FullPath)
            $rootFull = [string]$chain.RootPath
            if (-not [string]::Equals([IO.Path]::GetFullPath($final), $expectedFull, [StringComparison]::OrdinalIgnoreCase) -or -not ([IO.Path]::GetFullPath($final).StartsWith($rootFull + '\\', [StringComparison]::OrdinalIgnoreCase))) { throw $Code }
            $stream = New-Object IO.FileStream($handle, [IO.FileAccess]::Write, 65536, $false)
            try { $stream.Write($Bytes, 0, $Bytes.Length); $stream.Flush($true); $writtenIdentity = Get-HubQaEnvelopeHandleIdentity -Handle $handle -Code $Code; if (($writtenIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_DIRECTORY) -ne 0 -or ($writtenIdentity.Attributes -band [HubQaEnvelopeWin32FileNative]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0 -or $writtenIdentity.Size -ne $Bytes.Length) { throw $Code } } finally { $stream.Dispose() }
        } finally { if (-not $handle.IsClosed) { $handle.Dispose() } }
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $chain -Code $Code
        $reopened = Get-HubQaEnvelopeSafeContainedBytes -Root $Root -RelativePath $RelativePath -MaximumBytes ([Math]::Max([Int64]$Bytes.Length, [Int64]1)) -Code $Code -HeldDirectoryChain $chain
        Assert-HubQaEnvelopeSameStableFileIdentity -Actual $reopened.Identity -Expected $writtenIdentity -Code $Code
        if (-not [System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($Bytes, $reopened.Bytes)) { throw $Code }
        Assert-HubQaEnvelopeHeldDirectoryChain -Chain $chain -Code $Code
        return $reopened
    } finally { Close-HubQaEnvelopeHeldDirectoryChain -Chain $chain }
}

function New-HubQaRunEnvelope {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$BootstrapPrivateContext,
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$EnvelopePrivateContext
    )

    Assert-HubQaEnvelopeKeys -Object $BootstrapPrivateContext -Expected @('Task','Run','RoleClaim','OwnerNonce','OwnerMarkerRawBytes','OwnerMarkerSha256')
    Assert-HubQaEnvelopeKeys -Object $EnvelopePrivateContext -Expected @('SnapshotPath','SnapshotRawBytes','SnapshotSha256')
    $task = [string]$BootstrapPrivateContext['Task']; $run = [string]$BootstrapPrivateContext['Run']; $roleClaim = [string]$BootstrapPrivateContext['RoleClaim']; $ownerNonce = [string]$BootstrapPrivateContext['OwnerNonce']
    [byte[]]$ownerMarkerBytes = $BootstrapPrivateContext['OwnerMarkerRawBytes']
    $ownerMarkerSha256 = [string]$BootstrapPrivateContext['OwnerMarkerSha256']
    if ($task -cne 'g0-assets-runner' -or $task -cnotmatch '^[a-z][a-z0-9-]{2,31}$' -or $run -cnotmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or $roleClaim -cne 'hub.qa-bootstrap/v1' -or $ownerNonce -cnotmatch '^[0-9a-f]{32}$' -or $null -eq $ownerMarkerBytes -or $ownerMarkerBytes.Length -eq 0 -or $ownerMarkerSha256 -cnotmatch '^sha256:[a-f0-9]{64}$' -or (Get-HubQaEnvelopeSha256 -Bytes $ownerMarkerBytes) -cne $ownerMarkerSha256) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    [byte[]]$snapshotBytes = $EnvelopePrivateContext['SnapshotRawBytes']
    $reportRelative = 'testframework/reports/' + $task + '-' + $run
    $snapshotRelative = $reportRelative + '/snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json'
    if ([string]$EnvelopePrivateContext['SnapshotPath'] -cne $snapshotRelative -or $null -eq $snapshotBytes -or $snapshotBytes.Length -eq 0 -or [string]$EnvelopePrivateContext['SnapshotSha256'] -cnotmatch '^sha256:[a-f0-9]{64}$' -or (Get-HubQaEnvelopeSha256 -Bytes $snapshotBytes) -cne [string]$EnvelopePrivateContext['SnapshotSha256']) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $snapshot = ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $snapshotBytes
    $snapshotValue = Assert-HubQaSnapshotEnvelopePayload -Node $snapshot -Kind snapshot -ExpectedTask $task -ExpectedRun $run
    if ($snapshotValue['schema'].Value -cne 'hub.qa.repository-snapshot/v1' -or $snapshotValue['producer'].Value -cne 'hub.qa.repository-snapshot-producer/v1' -or $snapshotValue['snapshot_id'].Value -cne ($task + '-snapshot-' + $run + '-000-G0_BOOTSTRAP') -or $snapshotValue['task'].Value -cne $task -or $snapshotValue['run'].Value -cne $run -or $snapshotValue['envelope_id'].Value -cne ($task + '-envelope-' + $run) -or $snapshotValue['role_claim'].Value -cne $roleClaim -or $snapshotValue['self_attested'].Value -ne $true) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $capturedFor = $snapshotValue['captured_for']
    if ([string]$capturedFor.Type -cne 'object') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $capturedForValue = [System.Collections.IDictionary]$capturedFor.Value
    Assert-HubQaEnvelopeKeys -Object $capturedForValue -Expected @('kind','ordinal','subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256')
    if ($capturedForValue['kind'].Value -cne 'G0_BOOTSTRAP' -or $capturedForValue['ordinal'].Type -cne 'integer' -or [int64]$capturedForValue['ordinal'].Value -ne 0) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    foreach ($field in @('subject_path','subject_sha256','subject_commit_sha','predecessor_event_path','predecessor_event_sha256','predecessor_event_commit_sha','pre_event_snapshot_path','pre_event_snapshot_sha256')) { if ([string]$capturedForValue[$field].Type -cne 'null') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' } }
    $evidence = $snapshotValue['evidence']
    if ([string]$evidence.Type -cne 'object') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $evidenceValue = [System.Collections.IDictionary]$evidence.Value
    Assert-HubQaEnvelopeKeys -Object $evidenceValue -Expected @('positive_green_path','positive_green_sha256','candidate_lock_path','candidate_lock_sha256','cleanup_path','cleanup_sha256')
    foreach ($pair in @(@('positive_green_path','selftests/g0-planning-positive-green.v1.json'),@('candidate_lock_path','bootstrap/images.lock.v3.json'),@('cleanup_path','cleanup.json'))) {
        if ([string]$evidenceValue[$pair[0]].Type -cne 'string' -or [string]$evidenceValue[$pair[0]].Value -cne ($reportRelative + '/' + $pair[1])) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    }
    foreach ($field in @('positive_green_sha256','candidate_lock_sha256','cleanup_sha256')) { if ([string]$evidenceValue[$field].Type -cne 'string' -or [string]$evidenceValue[$field].Value -cnotmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' } }
    $root = Get-HubQaEnvelopeRepositoryRoot
    $envelopeRelative = $reportRelative + '/envelopes/' + $task + '-envelope-' + $run + '.run-envelope.v1.json'
    try { [byte[]]$storedSnapshotBytes = (Get-HubQaEnvelopeSafeContainedBytes -Root $root -RelativePath $snapshotRelative -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK').Bytes } catch { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    if (-not [System.Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($storedSnapshotBytes, $snapshotBytes)) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $envelopeParentChain = Open-HubQaEnvelopeHeldDirectoryPath -Root $root -RelativeDirectoryPath ($reportRelative + '/envelopes') -CreateLeaf -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    try { Assert-HubQaEnvelopeHeldDirectoryChain -Chain $envelopeParentChain -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK' } finally { Close-HubQaEnvelopeHeldDirectoryChain -Chain $envelopeParentChain }
    $payload = [ordered]@{
        schema = 'hub.qa.run-envelope/v1'
        producer = 'hub.qa.run-envelope-producer/v1'
        envelope_id = $task + '-envelope-' + $run
        contract = 'hub.qa-runner/v2'
        target = 'g0-planning-selftest'
        task = $task
        run = $run
        owner_nonce = $ownerNonce
        repository_snapshot_sha256 = [string]$EnvelopePrivateContext['SnapshotSha256']
        report_root = $reportRelative
        limits = [ordered]@{ command_seconds = 180; target_seconds = 780; inspect_seconds = 30; outer_seconds = 840; cleanup_seconds = 30; orchestration_slack_seconds = 30 }
        cleanup_contract = 'hub.qa.exact-cleanup/v1'
        created_at_utc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [Globalization.CultureInfo]::InvariantCulture)
    }
    [byte[]]$bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $payload))
    $parsed = ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $bytes
    [void](Assert-HubQaSnapshotEnvelopePayload -Node $parsed -Kind envelope -ExpectedTask $task -ExpectedRun $run)
    $written = Write-HubQaEnvelopeSafeCreateNew -Root $root -RelativePath $envelopeRelative -Bytes $bytes -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    return [pscustomobject][ordered]@{ EnvelopePath = $envelopeRelative; EnvelopeRawBytes = [byte[]]$written.Bytes; EnvelopeSha256 = Get-HubQaEnvelopeSha256 -Bytes $written.Bytes; EnvelopeId = $payload.envelope_id }
}

if ($MyInvocation.InvocationName -ne '.') { throw 'E_B0_DIRECT_INVOCATION_FORBIDDEN' }
