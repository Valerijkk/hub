Set-StrictMode -Version Latest

# This bootstrap-private host consumer has no script-level parameter surface and
# produces no pre-bootstrap runtime evidence.
# Package B owns helper cleanup and the later cleanup receipt; this A0.1 consumer
# only returns the fixed helper admission receipt to that later transaction.
$script:PreBootstrapStatus = 'NOT_READY/NOT_RUN'
$script:HelperBuildTag = 'hub-qa-g0-admission-helper:local'
$script:SourceCommit = '7b5d5501cefab34f9aa5661e4f8c38e71474f9ce'
$script:MaxBoundedControlBytes = 1048576

function Get-HubQaDockerClientPath {
    $programFiles = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles)
    if ([string]::IsNullOrWhiteSpace($programFiles)) {
        throw 'The operating system ProgramFiles location is unavailable.'
    }
    return [System.IO.Path]::GetFullPath((Join-Path $programFiles 'Docker\Docker\resources\bin\docker.exe'))
}

function Get-HubQaSha256 {
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)

    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        return 'sha256:' + ([System.BitConverter]::ToString($algorithm.ComputeHash($Bytes)).Replace('-', '').ToLowerInvariant())
    }
    finally {
        $algorithm.Dispose()
    }
}

function Test-HubQaByteArrayEqual {
    param(
        [Parameter(Mandatory = $true)][byte[]]$Left,
        [Parameter(Mandatory = $true)][byte[]]$Right
    )

    if ($Left.Length -ne $Right.Length) {
        return $false
    }
    for ($index = 0; $index -lt $Left.Length; $index++) {
        if ($Left[$index] -ne $Right[$index]) {
            return $false
        }
    }
    return $true
}

function Assert-HubQaExactPropertyNames {
    param(
        [Parameter(Mandatory = $true)][object]$Object,
        [Parameter(Mandatory = $true)][string[]]$Expected,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $actual = @($Object.PSObject.Properties | ForEach-Object { $_.Name })
    if ($actual.Count -ne $Expected.Count) {
        throw "$Identity has an additional or missing property."
    }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ($actual[$index] -cne $Expected[$index]) {
            throw "$Identity property order or name is not closed."
        }
    }
}

if ($null -eq ('HubQaG0NativeFile' -as [type])) {
    Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;

public static class HubQaG0NativeFile
{
    public const UInt32 GENERIC_READ = 0x80000000;
    public const UInt32 FILE_SHARE_READ = 0x00000001;
    public const UInt32 OPEN_EXISTING = 3;
    public const UInt32 FILE_ATTRIBUTE_DIRECTORY = 0x00000010;
    public const UInt32 FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400;
    public const UInt32 FILE_FLAG_BACKUP_SEMANTICS = 0x02000000;
    public const UInt32 FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000;
    public const UInt32 DUPLICATE_SAME_ACCESS = 0x00000002;

    [StructLayout(LayoutKind.Sequential)]
    public struct ByHandleFileInformation
    {
        public UInt32 FileAttributes;
        public UInt32 CreationTimeLow;
        public UInt32 CreationTimeHigh;
        public UInt32 LastAccessTimeLow;
        public UInt32 LastAccessTimeHigh;
        public UInt32 LastWriteTimeLow;
        public UInt32 LastWriteTimeHigh;
        public UInt32 VolumeSerialNumber;
        public UInt32 FileSizeHigh;
        public UInt32 FileSizeLow;
        public UInt32 NumberOfLinks;
        public UInt32 FileIndexHigh;
        public UInt32 FileIndexLow;
    }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern SafeFileHandle CreateFileW(
        string fileName,
        UInt32 desiredAccess,
        UInt32 shareMode,
        IntPtr securityAttributes,
        UInt32 creationDisposition,
        UInt32 flagsAndAttributes,
        IntPtr templateFile);

    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern bool GetFileInformationByHandle(
        SafeFileHandle file,
        out ByHandleFileInformation information);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern UInt32 GetFinalPathNameByHandle(
        SafeFileHandle file,
        StringBuilder path,
        UInt32 pathLength,
        UInt32 flags);

    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern bool DuplicateHandle(
        IntPtr sourceProcess,
        SafeFileHandle sourceHandle,
        IntPtr targetProcess,
        out SafeFileHandle targetHandle,
        UInt32 desiredAccess,
        bool inheritHandle,
        UInt32 options);

    [DllImport("kernel32.dll")]
    public static extern IntPtr GetCurrentProcess();
}
'@
}

function Get-HubQaHandleFileObservation {
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $information = New-Object -TypeName 'HubQaG0NativeFile+ByHandleFileInformation'
    if (-not [HubQaG0NativeFile]::GetFileInformationByHandle($Handle, [ref]$information)) {
        throw "$Identity cannot obtain handle file information: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())."
    }
    $length = ([UInt64]$information.FileSizeHigh -shl 32) -bor [UInt64]$information.FileSizeLow
    return [pscustomobject][ordered]@{
        Attributes = [UInt32]$information.FileAttributes
        VolumeSerialNumber = [UInt32]$information.VolumeSerialNumber
        FileIndexHigh = [UInt32]$information.FileIndexHigh
        FileIndexLow = [UInt32]$information.FileIndexLow
        Length = [UInt64]$length
        LastWriteTimeHigh = [UInt32]$information.LastWriteTimeHigh
        LastWriteTimeLow = [UInt32]$information.LastWriteTimeLow
    }
}

function Assert-HubQaSameHandleFileObservation {
    param(
        [Parameter(Mandatory = $true)][object]$Expected,
        [Parameter(Mandatory = $true)][object]$Actual,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    foreach ($name in @('Attributes', 'VolumeSerialNumber', 'FileIndexHigh', 'FileIndexLow', 'Length', 'LastWriteTimeHigh', 'LastWriteTimeLow')) {
        if ($Expected.$name -ne $Actual.$name) {
            throw "$Identity changed while its held handle was open."
        }
    }
}

function Assert-HubQaBoundedRegularControlObservation {
    param(
        [Parameter(Mandatory = $true)][object]$Observation,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if (($Observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_DIRECTORY) -ne 0 -or
        ($Observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) {
        throw "$Identity is nonregular or a reparse point."
    }
    if ($Observation.Length -lt 1 -or $Observation.Length -gt $script:MaxBoundedControlBytes) {
        throw "$Identity is outside the fixed bounded control-read limit."
    }
}

function New-HubQaOwnedDuplicateFileHandle {
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $process = [HubQaG0NativeFile]::GetCurrentProcess()
    $duplicate = $null
    if (-not [HubQaG0NativeFile]::DuplicateHandle(
        $process,
        $Handle,
        $process,
        [ref]$duplicate,
        [UInt32]0,
        $false,
        [HubQaG0NativeFile]::DUPLICATE_SAME_ACCESS
    )) {
        throw "$Identity cannot duplicate its held control handle: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())."
    }
    if ($null -eq $duplicate -or $duplicate.IsInvalid) {
        if ($null -ne $duplicate) {
            $duplicate.Dispose()
        }
        throw "$Identity duplicated an invalid control handle."
    }
    return $duplicate
}

function ConvertFrom-HubQaExtendedWindowsPath {
    param([Parameter(Mandatory = $true)][string]$Path)

    if ($Path.StartsWith('\\?\UNC\', [StringComparison]::OrdinalIgnoreCase)) {
        return '\\' + $Path.Substring(8)
    }
    if ($Path.StartsWith('\\?\', [StringComparison]::OrdinalIgnoreCase)) {
        return $Path.Substring(4)
    }
    return $Path
}

function Get-HubQaFinalPathNameByHandle {
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $capacity = 32768
    $builder = New-Object Text.StringBuilder $capacity
    $written = [HubQaG0NativeFile]::GetFinalPathNameByHandle($Handle, $builder, [UInt32]$capacity, [UInt32]0)
    if ($written -eq 0 -or $written -ge $capacity) {
        throw "$Identity cannot obtain a bounded final handle path: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())."
    }
    return [IO.Path]::GetFullPath((ConvertFrom-HubQaExtendedWindowsPath -Path $builder.ToString()))
}

function Get-HubQaPhysicalDirectoryPath {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $handle = [HubQaG0NativeFile]::CreateFileW(
        [IO.Path]::GetFullPath($Path),
        [HubQaG0NativeFile]::GENERIC_READ,
        [HubQaG0NativeFile]::FILE_SHARE_READ,
        [IntPtr]::Zero,
        [HubQaG0NativeFile]::OPEN_EXISTING,
        ([HubQaG0NativeFile]::FILE_FLAG_BACKUP_SEMANTICS -bor [HubQaG0NativeFile]::FILE_FLAG_OPEN_REPARSE_POINT),
        [IntPtr]::Zero
    )
    if ($null -eq $handle -or $handle.IsInvalid) {
        throw "$Identity cannot open its fixed root handle: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())."
    }
    try {
        $observation = Get-HubQaHandleFileObservation -Handle $handle -Identity $Identity
        if (($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or
            ($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) {
            throw "$Identity root is nonphysical or a reparse point."
        }
        return Get-HubQaFinalPathNameByHandle -Handle $handle -Identity $Identity
    }
    finally {
        $handle.Dispose()
    }
}

function Assert-HubQaNoReparseAncestors {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $directory = [IO.DirectoryInfo]::new((Split-Path -Parent $Path))
    $rootDirectory = [IO.Path]::GetFullPath($Root).TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar)
    while ($null -ne $directory) {
        $directory.Refresh()
        if (-not $directory.Exists -or ($directory.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw "$Identity has a missing or reparse ancestor."
        }
        if ([string]::Equals($directory.FullName.TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar), $rootDirectory, [StringComparison]::OrdinalIgnoreCase)) {
            return
        }
        $directory = $directory.Parent
    }
    throw "$Identity does not terminate at its fixed root."
}

function Assert-HubQaCanonicalContainedControlPath {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $fullPath = [IO.Path]::GetFullPath($Path)
    $fullRoot = [IO.Path]::GetFullPath($Root).TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar)
    $prefix = $fullRoot + [IO.Path]::DirectorySeparatorChar
    if (-not $fullPath.StartsWith($prefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw "$Identity escapes its fixed control root."
    }
    $relativePath = $fullPath.Substring($prefix.Length)
    if ($relativePath.Length -eq 0 -or $relativePath -match '(^|[\\/])[^\\/]*:' -or $relativePath.Split([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar) -contains '..') {
        throw "$Identity is not a canonical fixed control path."
    }
    Assert-HubQaNoReparseAncestors -Path $fullPath -Root $fullRoot -Identity $Identity
    $physicalRoot = Get-HubQaPhysicalDirectoryPath -Path $fullRoot -Identity $Identity
    return [pscustomobject][ordered]@{
        FullPath = $fullPath
        PhysicalRoot = $physicalRoot
        ExpectedPhysicalPath = [IO.Path]::GetFullPath((Join-Path $physicalRoot $relativePath))
    }
}

function Open-HubQaPhysicalRegularFileHandle {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $contained = Assert-HubQaCanonicalContainedControlPath -Path $Path -Root $Root -Identity $Identity
    $handle = [HubQaG0NativeFile]::CreateFileW(
        $contained.FullPath,
        [HubQaG0NativeFile]::GENERIC_READ,
        [HubQaG0NativeFile]::FILE_SHARE_READ,
        [IntPtr]::Zero,
        [HubQaG0NativeFile]::OPEN_EXISTING,
        [HubQaG0NativeFile]::FILE_FLAG_OPEN_REPARSE_POINT,
        [IntPtr]::Zero
    )
    if ($null -eq $handle -or $handle.IsInvalid) {
        throw "$Identity cannot open its fixed control handle: $([Runtime.InteropServices.Marshal]::GetLastWin32Error())."
    }
    try {
        $observation = Get-HubQaHandleFileObservation -Handle $handle -Identity $Identity
        Assert-HubQaBoundedRegularControlObservation -Observation $observation -Identity $Identity
        $finalPath = Get-HubQaFinalPathNameByHandle -Handle $handle -Identity $Identity
        if (-not [string]::Equals($finalPath, $contained.ExpectedPhysicalPath, [StringComparison]::OrdinalIgnoreCase) -or
            -not $finalPath.StartsWith($contained.PhysicalRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
            throw "$Identity final handle path is not the exact contained fixed control path."
        }
        return [pscustomobject][ordered]@{
            Handle = $handle
            Observation = $observation
            FinalPath = $finalPath
        }
    }
    catch {
        $handle.Dispose()
        throw
    }
}

function Read-HubQaHeldBoundedControlBytes {
    param(
        [Parameter(Mandatory = $true)][object]$Opened,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $preRead = Get-HubQaHandleFileObservation -Handle $Opened.Handle -Identity $Identity
    Assert-HubQaSameHandleFileObservation -Expected $Opened.Observation -Actual $preRead -Identity $Identity
    Assert-HubQaBoundedRegularControlObservation -Observation $preRead -Identity $Identity
    [byte[]]$bytes = New-Object byte[] ([int]$preRead.Length)
    [byte[]]$chunk = New-Object byte[] ([Math]::Min(65536, $bytes.Length))
    $duplicate = New-HubQaOwnedDuplicateFileHandle -Handle $Opened.Handle -Identity $Identity
    $stream = $null
    try {
        $stream = [IO.FileStream]::new($duplicate, [IO.FileAccess]::Read, 65536, $false)
        $duplicate = $null
        $offset = 0
        while ($offset -lt $preRead.Length) {
            $requested = [Math]::Min($chunk.Length, $bytes.Length - $offset)
            $read = $stream.Read($chunk, 0, $requested)
            if ($read -le 0) {
                throw "$Identity ended before its held bounded length."
            }
            [Buffer]::BlockCopy($chunk, 0, $bytes, $offset, $read)
            $offset += $read
        }
        if ($stream.ReadByte() -ne -1) {
            throw "$Identity exceeded its held bounded length."
        }
        $postRead = Get-HubQaHandleFileObservation -Handle $Opened.Handle -Identity $Identity
        Assert-HubQaSameHandleFileObservation -Expected $preRead -Actual $postRead -Identity $Identity
        Assert-HubQaBoundedRegularControlObservation -Observation $postRead -Identity $Identity
        return $bytes
    }
    finally {
        if ($null -ne $stream) {
            $stream.Dispose()
        }
        elseif ($null -ne $duplicate) {
            $duplicate.Dispose()
        }
    }
}

function Read-HubQaBoundedControlBytes {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)]
        [ValidateSet(
            'ingress-owner-marker',
            'a1-dependency-manifest',
            'copied-dependency-manifest',
            'ingress-copy-receipt',
            'local-base-receipt',
            'helper-image-attestation',
            'helper-admission-receipt'
        )]
        [string]$Control
    )

    $binding = switch ($Control) {
        'ingress-owner-marker' { [pscustomobject]@{ Path = $Closed.IngressOwnerMarkerPath; Root = $Closed.IngressRoot }; break }
        'a1-dependency-manifest' { [pscustomobject]@{ Path = (Join-Path $PSScriptRoot 'dependency-bundle.manifest.v1.json'); Root = $PSScriptRoot }; break }
        'copied-dependency-manifest' { [pscustomobject]@{ Path = $Closed.CopiedManifestPath; Root = $Closed.IngressRoot }; break }
        'ingress-copy-receipt' { [pscustomobject]@{ Path = $Closed.IngressCopyReceiptPath; Root = $Closed.ReportRoot }; break }
        'local-base-receipt' { [pscustomobject]@{ Path = $Closed.LocalBaseReceiptPath; Root = $Closed.ReportRoot }; break }
        'helper-image-attestation' { [pscustomobject]@{ Path = $Closed.HelperImageAttestationPath; Root = $Closed.ReportRoot }; break }
        'helper-admission-receipt' { [pscustomobject]@{ Path = $Closed.ReceiptPath; Root = $Closed.ReportRoot }; break }
        default { throw 'The bootstrap control selector is not closed.' }
    }
    $opened = $null
    try {
        $opened = Open-HubQaPhysicalRegularFileHandle -Path $binding.Path -Root $binding.Root -Identity $Control
        return Read-HubQaHeldBoundedControlBytes -Opened $opened -Identity $Control
    }
    finally {
        if ($null -ne $opened) {
            $opened.Handle.Dispose()
        }
    }
}

function Assert-HubQaCanonicalJsonString {
    param(
        [Parameter(Mandatory = $true)][string]$Escape,
        [AllowEmptyString()][string]$UnicodeEscape,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($Escape -ceq 'u') {
        if ($UnicodeEscape -cnotmatch '^00[0-9A-F]{2}$') {
            throw "$Identity has a noncanonical unicode escape."
        }
        $code = [Convert]::ToInt32($UnicodeEscape, 16)
        if ($code -gt 0x1f -or $code -in @(0x08, 0x09, 0x0a, 0x0c, 0x0d)) {
            throw "$Identity has a noncanonical unicode escape for a short JSON escape."
        }
        return
    }
    if ($Escape -cnotin @('"', '\', 'b', 'f', 'n', 'r', 't')) {
        throw "$Identity has a forbidden JSON string escape."
    }
}

function Assert-HubQaCanonicalJsonNumber {
    param(
        [Parameter(Mandatory = $true)][string]$Digits,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($Digits -notmatch '^(0|[1-9][0-9]*)$' -or $Digits.Length -gt 19) {
        throw "$Identity has a noncanonical JSON integer."
    }
}

function ConvertFrom-HubQaCanonicalJsonValue {
    param(
        [Parameter(Mandatory = $true)][string]$Text,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $cursor = 0
    $length = $Text.Length
    $parseValue = $null
    $parseArray = $null
    $parseObject = $null
    $parseString = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '"') { throw "$Identity has an invalid JSON string." }
        $cursor++
        $builder = New-Object Text.StringBuilder
        while ($cursor -lt $length) {
            $code = [int][char]$Text[$cursor]
            $cursor++
            if ($code -eq 0x22) { return $builder.ToString() }
            if ($code -lt 0x20) { throw "$Identity string has a control character." }
            if ($code -eq 0x5c) {
                if ($cursor -ge $length) { throw "$Identity string has a trailing escape." }
                $escape = [string]$Text[$cursor]
                $cursor++
                if ($escape -ceq 'u') {
                    if (($cursor + 4) -gt $length) { throw "$Identity string has a truncated unicode escape." }
                    $hex = $Text.Substring($cursor, 4)
                    Assert-HubQaCanonicalJsonString -Escape $escape -UnicodeEscape $hex -Identity $Identity
                    [void]$builder.Append([char][Convert]::ToInt32($hex, 16))
                    $cursor += 4
                    continue
                }
                Assert-HubQaCanonicalJsonString -Escape $escape -Identity $Identity
                switch ($escape) {
                    '"' { [void]$builder.Append('"'); continue }
                    '\' { [void]$builder.Append('\'); continue }
                    'b' { [void]$builder.Append([char]0x08); continue }
                    'f' { [void]$builder.Append([char]0x0c); continue }
                    'n' { [void]$builder.Append([char]0x0a); continue }
                    'r' { [void]$builder.Append([char]0x0d); continue }
                    't' { [void]$builder.Append([char]0x09); continue }
                }
            }
            if ($code -ge 0xd800 -and $code -le 0xdbff) {
                if ($cursor -ge $length) { throw "$Identity string has an unpaired high surrogate." }
                $low = [int][char]$Text[$cursor]
                if ($low -lt 0xdc00 -or $low -gt 0xdfff) { throw "$Identity string has an unpaired high surrogate." }
                [void]$builder.Append([char]$code)
                [void]$builder.Append([char]$low)
                $cursor++
                continue
            }
            if ($code -ge 0xdc00 -and $code -le 0xdfff) { throw "$Identity string has an unpaired low surrogate." }
            [void]$builder.Append([char]$code)
        }
        throw "$Identity string is unterminated."
    }
    $parseArray = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '[') { throw "$Identity has an invalid JSON array." }
        $cursor++
        $items = New-Object 'System.Collections.Generic.List[object]'
        if ($cursor -lt $length -and $Text[$cursor] -eq ']') {
            $cursor++
            return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
        }
        while ($true) {
            [void]$items.Add((. $parseValue))
            if ($cursor -ge $length) { throw "$Identity array is unterminated." }
            if ($Text[$cursor] -eq ']') { $cursor++; break }
            if ($Text[$cursor] -ne ',') { throw "$Identity array has an invalid separator or whitespace." }
            $cursor++
        }
        return [pscustomobject][ordered]@{ Type = 'array'; Value = [object[]]$items.ToArray() }
    }
    $parseObject = {
        if ($cursor -ge $length -or $Text[$cursor] -ne '{') { throw "$Identity has an invalid JSON object." }
        $cursor++
        $members = [ordered]@{}
        if ($cursor -lt $length -and $Text[$cursor] -eq '}') {
            $cursor++
            return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
        }
        while ($true) {
            $name = . $parseString
            if ($members.Contains($name)) { throw "$Identity has a duplicate JSON object key." }
            if ($cursor -ge $length -or $Text[$cursor] -ne ':') { throw "$Identity object has an invalid member separator or whitespace." }
            $cursor++
            $members[$name] = . $parseValue
            if ($cursor -ge $length) { throw "$Identity object is unterminated." }
            if ($Text[$cursor] -eq '}') { $cursor++; break }
            if ($Text[$cursor] -ne ',') { throw "$Identity object has an invalid separator or whitespace." }
            $cursor++
            if ($cursor -ge $length -or $Text[$cursor] -ne '"') { throw "$Identity object has an invalid member name." }
        }
        return [pscustomobject][ordered]@{ Type = 'object'; Value = $members }
    }
    $parseValue = {
        if ($cursor -ge $length) { throw "$Identity has an incomplete JSON value." }
        $character = $Text[$cursor]
        if ([int][char]$character -le 0x20) { throw "$Identity has insignificant whitespace." }
        switch ($character) {
            '{' { return (. $parseObject) }
            '[' { return (. $parseArray) }
            '"' { return [pscustomobject][ordered]@{ Type = 'string'; Value = (. $parseString) } }
            'n' {
                if (($cursor + 4) -le $length -and $Text.Substring($cursor, 4) -ceq 'null') { $cursor += 4; return [pscustomobject][ordered]@{ Type = 'null'; Value = $null } }
                throw "$Identity has an invalid literal."
            }
            't' {
                if (($cursor + 4) -le $length -and $Text.Substring($cursor, 4) -ceq 'true') { $cursor += 4; return [pscustomobject][ordered]@{ Type = 'boolean'; Value = $true } }
                throw "$Identity has an invalid literal."
            }
            'f' {
                if (($cursor + 5) -le $length -and $Text.Substring($cursor, 5) -ceq 'false') { $cursor += 5; return [pscustomobject][ordered]@{ Type = 'boolean'; Value = $false } }
                throw "$Identity has an invalid literal."
            }
            default {
                if ($character -notmatch '^[0-9]$') { throw "$Identity has an invalid JSON scalar." }
                $start = $cursor
                while ($cursor -lt $length -and $Text[$cursor] -match '^[0-9]$') { $cursor++ }
                $digits = $Text.Substring($start, $cursor - $start)
                Assert-HubQaCanonicalJsonNumber -Digits $digits -Identity $Identity
                try { $integer = [Int64]::Parse($digits, [Globalization.CultureInfo]::InvariantCulture) } catch { throw "$Identity has an invalid JSON integer." }
                return [pscustomobject][ordered]@{ Type = 'integer'; Value = $integer }
            }
        }
    }
    $node = . $parseValue
    if ($cursor -ne $length -or [string]$node.Type -ne 'object') { throw "$Identity has trailing data, whitespace, or no top-level object." }
    $materialize = $null
    $materialize = {
        param([object]$Current)
        switch ([string]$Current.Type) {
            'object' {
                $result = [ordered]@{}
                foreach ($name in @($Current.Value.Keys)) { $result[[string]$name] = & $materialize $Current.Value[$name] }
                return [pscustomobject]$result
            }
            'array' {
                $result = New-Object 'System.Collections.Generic.List[object]'
                foreach ($item in @($Current.Value)) { [void]$result.Add((& $materialize $item)) }
                return [object[]]$result.ToArray()
            }
            default { return $Current.Value }
        }
    }
    return & $materialize $node
}

function ConvertFrom-HubQaStrictCanonicalJson {
    param(
        [Parameter(Mandatory = $true)][byte[]]$Bytes,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($Bytes.Length -ge 3 -and $Bytes[0] -eq 0xEF -and $Bytes[1] -eq 0xBB -and $Bytes[2] -eq 0xBF) {
        throw "$Identity is UTF-8 with a forbidden BOM."
    }
    try {
        $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($Bytes)
        return ConvertFrom-HubQaCanonicalJsonValue -Text $text -Identity $Identity
    }
    catch {
        throw "$Identity is not strict canonical UTF-8 JSON."
    }
}

function Assert-HubQaExactStringArray {
    param(
        [Parameter(Mandatory = $true)][object]$Actual,
        [Parameter(Mandatory = $true)][string[]]$Expected,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($Actual -isnot [System.Array] -or $Actual.Count -ne $Expected.Count) {
        throw "$Identity has an unexpected string-array shape."
    }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ($Actual[$index] -isnot [string] -or $Actual[$index] -cne $Expected[$index]) {
            throw "$Identity has an unexpected string-array value or order."
        }
    }
}

function Assert-HubQaPrivateDependencyBundleManifest {
    param(
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $manifest = ConvertFrom-HubQaStrictCanonicalJson -Bytes $BootstrapPrivateContext.PrivateDependencyBundleManifest -Identity 'runner-held A1 dependency manifest'
    Assert-HubQaExactPropertyNames -Object $manifest -Expected @('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256') -Identity 'runner-held A1 dependency manifest'
    if ($manifest.schema -cne 'hub.qa.dependency-bundle/v1' -or
        $manifest.source_commit -cne $BootstrapPrivateContext.PrivateSourceCommit -or
        $manifest.bundle_id -isnot [string] -or [string]::IsNullOrWhiteSpace($manifest.bundle_id)) {
        throw 'The runner-held A1 dependency manifest identity is not closed.'
    }
    Assert-HubQaDigest -Value ([string]$manifest.bundle_sha256) -Identity 'runner-held A1 dependency manifest bundle digest'
    Assert-HubQaExactPropertyNames -Object $manifest.platform -Expected @('os', 'architecture', 'variant') -Identity 'runner-held A1 dependency manifest platform'
    if ($manifest.platform.os -cne 'linux' -or $manifest.platform.architecture -cne 'amd64' -or $null -ne $manifest.platform.variant) {
        throw 'The runner-held A1 dependency manifest platform is not closed.'
    }
    $expectedAssets = @(
        [pscustomobject]@{ Id = 'go-module-cache'; Targets = @('backend-unit', 'backend-openapi'); Path = 'assets/go-module-cache.tar' },
        [pscustomobject]@{ Id = 'go-race-toolchain'; Targets = @('backend-unit'); Path = 'assets/go-race-toolchain.tar' },
        [pscustomobject]@{ Id = 'frontend-node-modules'; Targets = @('frontend-unit'); Path = 'assets/frontend-node-modules.tar' },
        [pscustomobject]@{ Id = 'swag-binary'; Targets = @('backend-openapi'); Path = 'assets/swag-binary.tar' }
    )
    if ($manifest.assets -isnot [System.Array] -or $manifest.assets.Count -ne $expectedAssets.Count) {
        throw 'The runner-held A1 dependency manifest asset count is not closed.'
    }
    for ($index = 0; $index -lt $expectedAssets.Count; $index++) {
        $asset = $manifest.assets[$index]
        $expected = $expectedAssets[$index]
        Assert-HubQaExactPropertyNames -Object $asset -Expected @('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity 'runner-held A1 dependency manifest asset'
        Assert-HubQaExactStringArray -Actual $asset.targets -Expected $expected.Targets -Identity 'runner-held A1 dependency manifest asset targets'
        if ($asset.asset_id -cne $expected.Id -or $asset.relative_path -cne $expected.Path -or $asset.archive_format -cne 'posix-pax-tar' -or
            $asset.bytes -isnot [Int64] -or $asset.bytes -lt 1 -or $asset.member_count -isnot [Int64] -or $asset.member_count -lt 1 -or
            $asset.expanded_bytes -isnot [Int64] -or $asset.expanded_bytes -lt 1) {
            throw 'The runner-held A1 dependency manifest asset mapping is not closed.'
        }
        foreach ($digestName in @('sha256', 'layout_sha256', 'license_evidence_sha256', 'provenance_sha256')) {
            Assert-HubQaDigest -Value ([string]$asset.$digestName) -Identity ('runner-held A1 dependency manifest asset ' + $digestName)
        }
    }
    $expectedBases = @(
        [pscustomobject]@{ Id = 'go-backend-unit'; Consumers = @('backend-unit') },
        [pscustomobject]@{ Id = 'go-backend-openapi'; Consumers = @('backend-openapi') },
        [pscustomobject]@{ Id = 'frontend-unit'; Consumers = @('frontend-unit') },
        [pscustomobject]@{ Id = 'bootstrap-helper-python'; Consumers = @('g0-admission-helper', 'g0-planning-selftest') }
    )
    if ($manifest.bases -isnot [System.Array] -or $manifest.bases.Count -ne $expectedBases.Count) {
        throw 'The runner-held A1 dependency manifest base count is not closed.'
    }
    for ($index = 0; $index -lt $expectedBases.Count; $index++) {
        $base = $manifest.bases[$index]
        $expected = $expectedBases[$index]
        Assert-HubQaExactPropertyNames -Object $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Identity 'runner-held A1 dependency manifest base'
        Assert-HubQaExactStringArray -Actual $base.consumers -Expected $expected.Consumers -Identity 'runner-held A1 dependency manifest base consumers'
        if ($base.base_id -cne $expected.Id) {
            throw 'The runner-held A1 dependency manifest base mapping is not closed.'
        }
        foreach ($digestName in @('local_image_id', 'manifest_digest', 'license_evidence_sha256', 'provenance_sha256')) {
            Assert-HubQaDigest -Value ([string]$base.$digestName) -Identity ('runner-held A1 dependency manifest base ' + $digestName)
        }
    }
    return $manifest
}

function Assert-HubQaPrivateLocalBaseReceiptChildren {
    param(
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$DependencyManifest
    )

    $expectedBaseIds = @('go-backend-unit', 'go-backend-openapi', 'frontend-unit', 'bootstrap-helper-python')
    if ($BaseReceipt.bases -isnot [System.Array] -or $BaseReceipt.bases.Count -ne $expectedBaseIds.Count) {
        throw 'The runner-held local-base receipt child count is not closed.'
    }
    for ($index = 0; $index -lt $expectedBaseIds.Count; $index++) {
        $base = $BaseReceipt.bases[$index]
        Assert-HubQaExactPropertyNames -Object $base -Expected @('base_id', 'local_image_id', 'manifest_digest') -Identity 'runner-held local-base receipt base'
        if ($base.base_id -cne $expectedBaseIds[$index]) {
            throw 'The runner-held local-base receipt base order is not closed.'
        }
        Assert-HubQaDigest -Value ([string]$base.local_image_id) -Identity 'runner-held local-base receipt base local image ID'
        Assert-HubQaDigest -Value ([string]$base.manifest_digest) -Identity 'runner-held local-base receipt base manifest digest'
        foreach ($field in @('base_id', 'local_image_id', 'manifest_digest')) {
            if ($base.$field -cne $DependencyManifest.bases[$index].$field) {
                throw 'The runner-held local-base receipt base is not closed over the A1 manifest.'
            }
        }
    }
}

function Assert-HubQaHelperAdmissionReceiptChildren {
    param(
        [Parameter(Mandatory = $true)][object]$Receipt,
        [Parameter(Mandatory = $true)][object]$DependencyManifest
    )

    $expectedAssets = @(
        [pscustomobject]@{ Id = 'go-module-cache'; Targets = @('backend-unit', 'backend-openapi'); Path = 'assets/go-module-cache.tar' },
        [pscustomobject]@{ Id = 'go-race-toolchain'; Targets = @('backend-unit'); Path = 'assets/go-race-toolchain.tar' },
        [pscustomobject]@{ Id = 'frontend-node-modules'; Targets = @('frontend-unit'); Path = 'assets/frontend-node-modules.tar' },
        [pscustomobject]@{ Id = 'swag-binary'; Targets = @('backend-openapi'); Path = 'assets/swag-binary.tar' }
    )
    if ($Receipt.assets -isnot [System.Array] -or $Receipt.assets.Count -ne $expectedAssets.Count) {
        throw 'The helper admission receipt asset count is not closed.'
    }
    $phases = @('opened', 'post_hash', 'post_raw_scan', 'post_tarfile', 'pre_stage', 'post_stage')
    foreach ($index in 0..($expectedAssets.Count - 1)) {
        $asset = $Receipt.assets[$index]
        $expected = $expectedAssets[$index]
        Assert-HubQaExactPropertyNames -Object $asset -Expected @('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256', 'descriptor_contract', 'source_observations', 'staged_targets') -Identity 'helper admission receipt asset'
        Assert-HubQaExactStringArray -Actual $asset.targets -Expected $expected.Targets -Identity 'helper admission receipt asset targets'
        if ($asset.asset_id -cne $expected.Id -or $asset.relative_path -cne $expected.Path -or $asset.archive_format -cne 'posix-pax-tar' -or
            $asset.descriptor_contract -cne 'linux-held-nofollow-fd/v1' -or $asset.bytes -isnot [Int64] -or $asset.bytes -lt 1 -or
            $asset.member_count -isnot [Int64] -or $asset.member_count -lt 1 -or $asset.expanded_bytes -isnot [Int64] -or $asset.expanded_bytes -lt 1) {
            throw 'The helper admission receipt asset mapping is not closed.'
        }
        $manifestAsset = $DependencyManifest.assets[$index]
        foreach ($field in @('sha256', 'bytes', 'member_count', 'expanded_bytes', 'layout_sha256')) {
            if ($asset.$field -ne $manifestAsset.$field) {
                throw 'The helper admission receipt asset is not closed over the A1 manifest.'
            }
        }
        foreach ($digestName in @('sha256', 'layout_sha256')) {
            Assert-HubQaDigest -Value ([string]$asset.$digestName) -Identity ('helper admission receipt asset ' + $digestName)
        }
        Assert-HubQaExactPropertyNames -Object $asset.source_observations -Expected $phases -Identity 'helper admission receipt source observations'
        foreach ($phase in $phases) {
            $observation = $asset.source_observations.$phase
            Assert-HubQaExactPropertyNames -Object $observation -Expected @('descriptor', 'pathname') -Identity 'helper admission receipt source observation'
            foreach ($identityName in @('descriptor', 'pathname')) {
                $identity = $observation.$identityName
                Assert-HubQaExactPropertyNames -Object $identity -Expected @('device', 'inode', 'mode', 'size', 'mtime_ns') -Identity 'helper admission receipt source observation identity'
                foreach ($field in @('device', 'inode', 'mode', 'size', 'mtime_ns')) {
                    if ($identity.$field -isnot [Int64] -or $identity.$field -lt 0) {
                        throw 'The helper admission receipt source observation identity is not a closed nonnegative integer tuple.'
                    }
                }
            }
            foreach ($field in @('device', 'inode', 'mode', 'size', 'mtime_ns')) {
                if ($observation.descriptor.$field -ne $observation.pathname.$field) {
                    throw 'The helper admission receipt descriptor and pathname identities disagree.'
                }
            }
        }
        if ($asset.staged_targets -isnot [System.Array] -or $asset.staged_targets.Count -ne $expected.Targets.Count) {
            throw 'The helper admission receipt staged-target count is not closed.'
        }
        for ($targetIndex = 0; $targetIndex -lt $expected.Targets.Count; $targetIndex++) {
            $staged = $asset.staged_targets[$targetIndex]
            Assert-HubQaExactPropertyNames -Object $staged -Expected @('target', 'relative_tar_path', 'bytes', 'sha256', 'create_new', 'fsync', 'safe_reopen', 'rehash_sha256') -Identity 'helper admission receipt staged target'
            if ($staged.target -cne $expected.Targets[$targetIndex] -or
                $staged.relative_tar_path -cne ('bootstrap/build-context/' + $expected.Targets[$targetIndex] + '/bindings/' + $expected.Path) -or
                $staged.bytes -ne $asset.bytes -or $staged.sha256 -cne $asset.sha256 -or $staged.rehash_sha256 -cne $asset.sha256 -or
                $staged.create_new -ne $true -or $staged.fsync -ne $true -or $staged.safe_reopen -ne $true) {
                throw 'The helper admission receipt staged target is not closed over the admitted asset.'
            }
        }
    }
    $expectedBases = @(
        [pscustomobject]@{ Id = 'go-backend-unit'; Consumers = @('backend-unit') },
        [pscustomobject]@{ Id = 'go-backend-openapi'; Consumers = @('backend-openapi') },
        [pscustomobject]@{ Id = 'frontend-unit'; Consumers = @('frontend-unit') },
        [pscustomobject]@{ Id = 'bootstrap-helper-python'; Consumers = @('g0-admission-helper', 'g0-planning-selftest') }
    )
    if ($Receipt.bases -isnot [System.Array] -or $Receipt.bases.Count -ne $expectedBases.Count) {
        throw 'The helper admission receipt base count is not closed.'
    }
    for ($index = 0; $index -lt $expectedBases.Count; $index++) {
        $base = $Receipt.bases[$index]
        Assert-HubQaExactPropertyNames -Object $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest') -Identity 'helper admission receipt base'
        if ($base.base_id -cne $expectedBases[$index].Id) {
            throw 'The helper admission receipt base order is not closed.'
        }
        Assert-HubQaExactStringArray -Actual $base.consumers -Expected $expectedBases[$index].Consumers -Identity 'helper admission receipt base consumers'
        Assert-HubQaDigest -Value ([string]$base.local_image_id) -Identity 'helper admission receipt base local image ID'
        Assert-HubQaDigest -Value ([string]$base.manifest_digest) -Identity 'helper admission receipt base manifest digest'
        foreach ($field in @('base_id', 'local_image_id', 'manifest_digest')) {
            if ($base.$field -cne $DependencyManifest.bases[$index].$field) {
                throw 'The helper admission receipt base is not closed over the A1 manifest.'
            }
        }
    }
}

function Assert-HubQaDigest {
    param([Parameter(Mandatory = $true)][string]$Value, [Parameter(Mandatory = $true)][string]$Identity)

    if ($Value -notmatch '^sha256:[0-9a-f]{64}$') {
        throw "$Identity is not a lowercase SHA-256 digest."
    }
}

function Assert-HubQaPrivateBootstrapContext {
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $required = @(
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
    if ($BootstrapPrivateContext.Keys.Count -ne $required.Count) {
        throw 'Bootstrap private context has an additional or missing field.'
    }
    foreach ($name in $required) {
        if (-not $BootstrapPrivateContext.ContainsKey($name)) {
            throw 'Bootstrap private context has an additional or missing field.'
        }
    }

    $dockerClientPath = Get-HubQaDockerClientPath
    if ($BootstrapPrivateContext.Task -cne 'g0-assets-runner' -or
        $BootstrapPrivateContext.Run -notmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or
        $BootstrapPrivateContext.PrivateOfflineIngressHandoffId -notmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or
        $BootstrapPrivateContext.PrivateSourceCommit -cne $script:SourceCommit -or
        $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -notmatch '^[0-9a-f]{32}$') {
        throw 'Bootstrap private context has an invalid closed task, run, or source identity.'
    }
    foreach ($digestName in @(
        'OwnerMarkerSha256',
        'PrivateIngressCopyReceiptSha256',
        'PrivateIngressDigestMapSha256',
        'PrivateOfflineIngressOwnerMarkerSha256',
        'PrivateDependencyBundleManifestSha256',
        'PrivateLocalBaseReceiptSha256',
        'PrivateHelperImageAttestationSha256',
        'PrivateBootstrapHelperPythonLocalImageIdSha256'
    )) {
        Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext[$digestName]) -Identity $digestName
    }
    if ($BootstrapPrivateContext.OwnerMarkerSha256 -ceq $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256) {
        throw 'Bootstrap private context must keep distinct task and offline-ingress owner-marker bindings.'
    }
    if ($BootstrapPrivateContext.PrivateDockerClientPath -cne $dockerClientPath -or
        $BootstrapPrivateContext.PrivateHelperBuildTag -cne $script:HelperBuildTag) {
        throw 'Bootstrap private context has a substituted Docker client or helper build tag.'
    }
    foreach ($pathName in @('PrivateReportRoot', 'PrivateIngressRoot', 'PrivateIngressSourceRoot', 'PrivateIngressDestinationRoot')) {
        if ($BootstrapPrivateContext[$pathName] -isnot [string] -or [string]::IsNullOrWhiteSpace($BootstrapPrivateContext[$pathName])) {
            throw "Bootstrap private context has an invalid $pathName path binding."
        }
    }
    foreach ($bytesName in @('PrivateIngressCopyReceipt', 'PrivateDependencyBundleManifest', 'PrivateLocalBaseReceipt', 'PrivateHelperImageAttestation')) {
        if ($BootstrapPrivateContext[$bytesName] -isnot [byte[]] -or
            $BootstrapPrivateContext[$bytesName].Length -lt 1 -or
            $BootstrapPrivateContext[$bytesName].Length -gt $script:MaxBoundedControlBytes) {
            throw "Bootstrap private context lacks bounded $bytesName raw bytes."
        }
    }
    if ($BootstrapPrivateContext.PrivateIngressDigestMap -isnot [System.Collections.IDictionary] -or
        $BootstrapPrivateContext.PrivateIngressByteCounts -isnot [System.Collections.IDictionary] -or
        $BootstrapPrivateContext.PrivateIngressOrderedPaths -isnot [System.Array] -or
        $BootstrapPrivateContext.PrivateIngressOrderedPaths.Count -lt 2) {
        throw 'Bootstrap private context lacks the ordered ingress copy bindings.'
    }
    if ((Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateIngressCopyReceipt) -cne $BootstrapPrivateContext.PrivateIngressCopyReceiptSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateDependencyBundleManifest) -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateLocalBaseReceipt) -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        (Get-HubQaSha256 -Bytes $BootstrapPrivateContext.PrivateHelperImageAttestation) -cne $BootstrapPrivateContext.PrivateHelperImageAttestationSha256) {
        throw 'Bootstrap private context raw receipt, manifest, or helper-attestation digest does not match its held bytes.'
    }
    foreach ($imageIdName in @('PrivateHelperLocalImageId', 'PrivateBootstrapHelperPythonLocalImageId')) {
        Assert-HubQaDigest -Value ([string]$BootstrapPrivateContext[$imageIdName]) -Identity $imageIdName
    }
    if ((Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageId))) -cne
        $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageIdSha256) {
        throw 'Bootstrap private context helper-base local image ID does not match its held digest.'
    }

    $comparison = [System.StringComparison]::OrdinalIgnoreCase
    $reportRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateReportRoot)
    $ingressRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateIngressRoot)
    $expectedIngressRoot = [System.IO.Path]::Combine($reportRoot, 'ingress', 'dependency-bundle')
    $destinationRoot = [System.IO.Path]::GetFullPath([string]$BootstrapPrivateContext.PrivateIngressDestinationRoot)
    if (-not [string]::Equals($ingressRoot, $expectedIngressRoot, $comparison) -or
        -not [string]::Equals($destinationRoot, $expectedIngressRoot, $comparison)) {
        throw 'Bootstrap private ingress destination is not the fixed task-owned copied root.'
    }
    foreach ($path in @($reportRoot, $ingressRoot)) {
        $item = [System.IO.DirectoryInfo]::new($path)
        $item.Refresh()
        if (-not $item.Exists -or
            ($item.Attributes -band [System.IO.FileAttributes]::Directory) -eq 0 -or
            ($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'Bootstrap private report or copied ingress root is not a physical directory.'
        }
    }
    return [pscustomobject][ordered]@{
        ReportRoot = $reportRoot
        IngressRoot = $ingressRoot
        ReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'g0-admission-helper.admission.v1.json')
        IngressCopyReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'ingress-copy-receipt.v1.json')
        IngressOwnerMarkerPath = [System.IO.Path]::Combine($ingressRoot, 'handoff-owner.v1.json')
        CopiedManifestPath = [System.IO.Path]::Combine($ingressRoot, 'dependency-bundle.manifest.v1.json')
        LocalBaseReceiptPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'local-base-inspect.v1.json')
        HelperImageAttestationPath = [System.IO.Path]::Combine($reportRoot, 'bootstrap', 'g0-admission-helper.local-image-attestation.v1.json')
        HelperContainerName = 'hub-qa-' + [string]$BootstrapPrivateContext.Task + '-' + [string]$BootstrapPrivateContext.Run + '-admission-helper'
        HelperOwnerLabel = 'hub.qa.owner-marker-sha256=' + [string]$BootstrapPrivateContext.OwnerMarkerSha256
    }
}

function Assert-HubQaSafeIngressRelativePath {
    param([Parameter(Mandatory = $true)][string]$RelativePath)

    if ($RelativePath.Length -eq 0 -or $RelativePath.Contains('\') -or $RelativePath.StartsWith('/') -or $RelativePath.Split('/') -contains '..') {
        throw 'A runner-held ingress path is unsafe.'
    }
    foreach ($part in $RelativePath.Split('/')) {
        if ($part.Length -eq 0 -or $part -ceq '.') {
            throw 'A runner-held ingress path is unsafe.'
        }
    }
}

function Assert-HubQaClosedIngressMetadata {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $seenFiles = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    $assetPaths = [System.Collections.Generic.List[string]]::new()
    $expectedAssetPaths = @(
        'assets/go-module-cache.tar',
        'assets/go-race-toolchain.tar',
        'assets/frontend-node-modules.tar',
        'assets/swag-binary.tar'
    )
    foreach ($relativePath in $BootstrapPrivateContext.PrivateIngressOrderedPaths) {
        if ($relativePath -isnot [string]) {
            throw 'The runner-held ordered ingress paths contain a non-string item.'
        }
        Assert-HubQaSafeIngressRelativePath -RelativePath $relativePath
        if (-not $seenFiles.Add($relativePath)) {
            throw 'The runner-held ordered ingress paths contain a duplicate item.'
        }
        if (-not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains($relativePath) -or
            -not $BootstrapPrivateContext.PrivateIngressByteCounts.Contains($relativePath)) {
            throw 'The runner-held ordered ingress paths are not closed over the digest and byte maps.'
        }
        $expectedDigest = $BootstrapPrivateContext.PrivateIngressDigestMap[$relativePath]
        $expectedBytes = $BootstrapPrivateContext.PrivateIngressByteCounts[$relativePath]
        Assert-HubQaDigest -Value ([string]$expectedDigest) -Identity 'runner-held ingress digest'
        if (($expectedBytes -isnot [int] -and $expectedBytes -isnot [long]) -or $expectedBytes -lt 1) {
            throw 'The runner-held ingress byte count is invalid.'
        }
        if ($relativePath.StartsWith('assets/', [System.StringComparison]::Ordinal)) {
            $assetPaths.Add($relativePath)
        }
        elseif ($relativePath -cne 'handoff-owner.v1.json' -and
            $relativePath -cne 'dependency-bundle.manifest.v1.json' -and
            $relativePath -notmatch '^evidence/licenses/(assets|bases)/[^/]+\.txt$' -and
            $relativePath -notmatch '^evidence/provenance/(assets|bases)/[^/]+\.json$') {
            throw 'The runner-held ingress path is outside the closed control and asset layout.'
        }
    }
    if ($BootstrapPrivateContext.PrivateIngressDigestMap.Count -ne $seenFiles.Count -or
        $BootstrapPrivateContext.PrivateIngressByteCounts.Count -ne $seenFiles.Count) {
        throw 'The runner-held ingress digest or byte map count is not closed.'
    }
    if (-not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains('handoff-owner.v1.json') -or
        $BootstrapPrivateContext.PrivateIngressDigestMap['handoff-owner.v1.json'] -cne $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256 -or
        -not $BootstrapPrivateContext.PrivateIngressDigestMap.Contains('dependency-bundle.manifest.v1.json') -or
        $BootstrapPrivateContext.PrivateIngressDigestMap['dependency-bundle.manifest.v1.json'] -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256) {
        throw 'The runner-held ingress map does not bind the copied handoff owner and dependency manifest bytes.'
    }
    if ($assetPaths.Count -ne $expectedAssetPaths.Count) {
        throw 'The runner-held ingress metadata has an unexpected asset path count.'
    }
    for ($index = 0; $index -lt $expectedAssetPaths.Count; $index++) {
        if ($assetPaths[$index] -cne $expectedAssetPaths[$index]) {
            throw 'The runner-held ingress metadata has an unexpected asset path order or name.'
        }
    }

    $relativeLines = [System.Collections.Generic.List[string]]::new()
    foreach ($relativePath in $BootstrapPrivateContext.PrivateIngressOrderedPaths) {
        $expectedDigest = $BootstrapPrivateContext.PrivateIngressDigestMap[$relativePath]
        $relativeLines.Add($relativePath + "`t" + $expectedDigest + "`n")
    }
    $digestMapBytes = [System.Text.Encoding]::UTF8.GetBytes([string]::Concat($relativeLines))
    if ((Get-HubQaSha256 -Bytes $digestMapBytes) -cne $BootstrapPrivateContext.PrivateIngressDigestMapSha256) {
        throw 'The ordered private ingress digest map does not match its held digest.'
    }
}

function Assert-HubQaIngressOwnerMarker {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $bytes = Read-HubQaBoundedControlBytes -Closed $Closed -Control 'ingress-owner-marker'
    if ((Get-HubQaSha256 -Bytes $bytes) -cne $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256) {
        throw 'The copied ingress owner marker is not bound to the runner-held owner-marker digest.'
    }
    $owner = ConvertFrom-HubQaStrictCanonicalJson -Bytes $bytes -Identity 'copied ingress owner marker'
    Assert-HubQaExactPropertyNames -Object $owner -Expected @(
        'schema', 'handoff_id', 'source_commit', 'dependency_bundle_manifest_sha256', 'owner', 'released_at_utc'
    ) -Identity 'copied ingress owner marker'
    if ($owner.schema -cne 'hub.qa.g0-offline-ingress-owner/v1' -or
        $owner.handoff_id -cne $BootstrapPrivateContext.PrivateOfflineIngressHandoffId -or
        $owner.source_commit -cne $BootstrapPrivateContext.PrivateSourceCommit -or
        $owner.dependency_bundle_manifest_sha256 -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        $owner.owner -isnot [string] -or [string]::IsNullOrWhiteSpace($owner.owner) -or
        $owner.released_at_utc -isnot [string] -or [string]::IsNullOrWhiteSpace($owner.released_at_utc)) {
        throw 'The copied ingress owner marker is not closed over the private bootstrap context.'
    }
}

function Assert-HubQaIngressCopyReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $actualBytes = Read-HubQaBoundedControlBytes -Closed $Closed -Control 'ingress-copy-receipt'
    if (-not (Test-HubQaByteArrayEqual -Left $actualBytes -Right $BootstrapPrivateContext.PrivateIngressCopyReceipt)) {
        throw 'The on-disk ingress-copy receipt is not byte-identical to the runner-held receipt.'
    }
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Bytes $BootstrapPrivateContext.PrivateIngressCopyReceipt -Identity 'runner-held ingress-copy receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'task', 'run', 'owner_marker_sha256', 'offline_ingress_handoff_id', 'source_commit',
        'source_owner_marker_sha256', 'dependency_bundle_manifest_sha256', 'source_root', 'destination_root',
        'paths', 'digest_map_sha256', 'local_base_receipt_sha256'
    ) -Identity 'runner-held ingress-copy receipt'
    if ($receipt.schema -cne 'hub.qa.g0-ingress-copy-receipt/v1' -or
        $receipt.task -cne $BootstrapPrivateContext.Task -or
        $receipt.run -cne $BootstrapPrivateContext.Run -or
        $receipt.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $receipt.offline_ingress_handoff_id -cne $BootstrapPrivateContext.PrivateOfflineIngressHandoffId -or
        $receipt.source_commit -cne $BootstrapPrivateContext.PrivateSourceCommit -or
        $receipt.source_owner_marker_sha256 -cne $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256 -or
        $receipt.dependency_bundle_manifest_sha256 -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        $receipt.source_root -cne $BootstrapPrivateContext.PrivateIngressSourceRoot -or
        $receipt.destination_root -cne $BootstrapPrivateContext.PrivateIngressDestinationRoot -or
        $receipt.digest_map_sha256 -cne $BootstrapPrivateContext.PrivateIngressDigestMapSha256 -or
        $receipt.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256) {
        throw 'The runner-held ingress-copy receipt is not closed over the private bootstrap context.'
    }
    if ($receipt.paths -isnot [System.Array] -or $receipt.paths.Count -ne $BootstrapPrivateContext.PrivateIngressOrderedPaths.Count) {
        throw 'The runner-held ingress-copy receipt path count is not closed.'
    }
    for ($index = 0; $index -lt $receipt.paths.Count; $index++) {
        $entry = $receipt.paths[$index]
        Assert-HubQaExactPropertyNames -Object $entry -Expected @('relative_path', 'bytes', 'sha256') -Identity 'runner-held ingress-copy receipt path'
        $relativePath = $BootstrapPrivateContext.PrivateIngressOrderedPaths[$index]
        if ($entry.relative_path -cne $relativePath -or
            $entry.bytes -ne $BootstrapPrivateContext.PrivateIngressByteCounts[$relativePath] -or
            $entry.sha256 -cne $BootstrapPrivateContext.PrivateIngressDigestMap[$relativePath]) {
            throw 'The runner-held ingress-copy receipt paths are not in the held ordered digest-map order.'
        }
    }
}

function Assert-HubQaPrivateManifestAndBaseReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    foreach ($control in @('a1-dependency-manifest', 'copied-dependency-manifest')) {
        $actualManifest = Read-HubQaBoundedControlBytes -Closed $Closed -Control $control
        if (-not (Test-HubQaByteArrayEqual -Left $actualManifest -Right $BootstrapPrivateContext.PrivateDependencyBundleManifest)) {
            throw 'The A1 or copied dependency manifest is not byte-identical to the runner-held manifest bytes.'
        }
    }
    $dependencyManifest = Assert-HubQaPrivateDependencyBundleManifest -BootstrapPrivateContext $BootstrapPrivateContext
    $actualBaseReceipt = Read-HubQaBoundedControlBytes -Closed $Closed -Control 'local-base-receipt'
    if (-not (Test-HubQaByteArrayEqual -Left $actualBaseReceipt -Right $BootstrapPrivateContext.PrivateLocalBaseReceipt)) {
        throw 'The on-disk local-base receipt is not byte-identical to the runner-held receipt.'
    }
    $baseReceipt = ConvertFrom-HubQaStrictCanonicalJson -Bytes $BootstrapPrivateContext.PrivateLocalBaseReceipt -Identity 'runner-held local-base receipt'
    Assert-HubQaExactPropertyNames -Object $baseReceipt -Expected @('schema', 'task', 'run', 'owner_marker_sha256', 'receipt_nonce', 'bases') -Identity 'runner-held local-base receipt'
    if ($baseReceipt.schema -cne 'hub.qa.local-base-inspect/v1' -or
        $baseReceipt.task -cne $BootstrapPrivateContext.Task -or
        $baseReceipt.run -cne $BootstrapPrivateContext.Run -or
        $baseReceipt.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $baseReceipt.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $baseReceipt.bases -isnot [System.Array] -or $baseReceipt.bases.Count -ne 4) {
        throw 'The runner-held local-base receipt is not bound to the private task, owner, nonce, and base count.'
    }
    Assert-HubQaPrivateLocalBaseReceiptChildren -BaseReceipt $baseReceipt -DependencyManifest $dependencyManifest
}

function Assert-HubQaPrivateHelperImageAttestation {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $actualBytes = Read-HubQaBoundedControlBytes -Closed $Closed -Control 'helper-image-attestation'
    if (-not (Test-HubQaByteArrayEqual -Left $actualBytes -Right $BootstrapPrivateContext.PrivateHelperImageAttestation)) {
        throw 'The on-disk helper image attestation is not byte-identical to the runner-held attestation.'
    }
    $attestation = ConvertFrom-HubQaStrictCanonicalJson -Bytes $BootstrapPrivateContext.PrivateHelperImageAttestation -Identity 'runner-held helper image attestation'
    Assert-HubQaExactPropertyNames -Object $attestation -Expected @(
        'schema', 'task', 'run', 'owner_marker_sha256', 'local_base_receipt_sha256', 'receipt_nonce',
        'helper_local_image_id', 'helper_local_image_id_sha256',
        'bootstrap_helper_python_local_image_id', 'bootstrap_helper_python_local_image_id_sha256',
        'helper_dockerfile_sha256', 'helper_entrypoint_sha256'
    ) -Identity 'runner-held helper image attestation'
    if ($attestation.schema -cne 'hub.qa.g0-admission-helper-local-image-attestation/v1' -or
        $attestation.task -cne $BootstrapPrivateContext.Task -or
        $attestation.run -cne $BootstrapPrivateContext.Run -or
        $attestation.owner_marker_sha256 -cne $BootstrapPrivateContext.OwnerMarkerSha256 -or
        $attestation.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        $attestation.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $attestation.helper_local_image_id -cne $BootstrapPrivateContext.PrivateHelperLocalImageId -or
        $attestation.bootstrap_helper_python_local_image_id -cne $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageId -or
        $attestation.bootstrap_helper_python_local_image_id_sha256 -cne $BootstrapPrivateContext.PrivateBootstrapHelperPythonLocalImageIdSha256) {
        throw 'The runner-held helper image attestation is not bound to the private bootstrap task, receipt, nonce, or local image IDs.'
    }
    foreach ($digestName in @(
        'owner_marker_sha256', 'local_base_receipt_sha256', 'helper_local_image_id_sha256',
        'bootstrap_helper_python_local_image_id_sha256', 'helper_dockerfile_sha256', 'helper_entrypoint_sha256'
    )) {
        Assert-HubQaDigest -Value ([string]$attestation.$digestName) -Identity ('helper image attestation ' + $digestName)
    }
    if ((Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$attestation.helper_local_image_id))) -cne $attestation.helper_local_image_id_sha256 -or
        (Get-HubQaSha256 -Bytes ([System.Text.Encoding]::UTF8.GetBytes([string]$attestation.bootstrap_helper_python_local_image_id))) -cne $attestation.bootstrap_helper_python_local_image_id_sha256) {
        throw 'The runner-held helper image attestation local image ID digest does not match its exact full image ID.'
    }
}

function Assert-HubQaHelperAdmissionReceipt {
    param(
        [Parameter(Mandatory = $true)][object]$Closed,
        [Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext
    )

    $bytes = Read-HubQaBoundedControlBytes -Closed $Closed -Control 'helper-admission-receipt'
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Bytes $bytes -Identity 'helper admission receipt'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases',
        'local_base_receipt_sha256', 'receipt_nonce', 'verdict'
    ) -Identity 'helper admission receipt'
    if ($receipt.schema -cne 'hub.qa.g0-dependency-admission-receipt/v1' -or
        $receipt.source_commit -cne $script:SourceCommit -or
        $receipt.manifest_sha256 -cne $BootstrapPrivateContext.PrivateDependencyBundleManifestSha256 -or
        $receipt.local_base_receipt_sha256 -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256 -or
        $receipt.receipt_nonce -cne $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce -or
        $receipt.verdict -cne 'ACCEPT' -or
        $receipt.assets -isnot [System.Array] -or $receipt.assets.Count -ne 4 -or
        $receipt.bases -isnot [System.Array] -or $receipt.bases.Count -ne 4) {
        throw 'The helper admission receipt is not bound to the verified private inputs.'
    }
    $dependencyManifest = Assert-HubQaPrivateDependencyBundleManifest -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaHelperAdmissionReceiptChildren -Receipt $receipt -DependencyManifest $dependencyManifest
}

function Invoke-HubQaG0AdmissionFromBootstrapContext {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $closed = Assert-HubQaPrivateBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateHelperImageAttestation -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaClosedIngressMetadata -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaIngressOwnerMarker -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaIngressCopyReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateManifestAndBaseReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    if ([System.IO.File]::Exists($closed.ReceiptPath) -or [System.IO.Directory]::Exists($closed.ReceiptPath)) {
        throw 'The fixed helper admission receipt path already exists.'
    }
    $dockerArguments = @(
        'run', '--pull', 'never', '--name', $closed.HelperContainerName, '--label', $closed.HelperOwnerLabel,
        '--network', 'none', '--read-only', '--cap-drop', 'ALL',
        '--security-opt', 'no-new-privileges:true', '--cpus', '2.0', '--memory', '2g', '--pids-limit', '256',
        '--mount', ('type=bind,src=' + $closed.ReportRoot + ',dst=/work'),
        '--env', ('HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_SHA256=' + $BootstrapPrivateContext.PrivateLocalBaseReceiptSha256),
        '--env', ('HUB_QA_PRIVATE_LOCAL_BASE_RECEIPT_NONCE=' + $BootstrapPrivateContext.PrivateLocalBaseReceiptNonce),
        '--env', ('HUB_QA_PRIVATE_OWNER_MARKER_SHA256=' + $BootstrapPrivateContext.OwnerMarkerSha256),
        '--env', ('HUB_QA_PRIVATE_OFFLINE_INGRESS_OWNER_MARKER_SHA256=' + $BootstrapPrivateContext.PrivateOfflineIngressOwnerMarkerSha256),
        $BootstrapPrivateContext.PrivateHelperLocalImageId
    )
    & $BootstrapPrivateContext.PrivateDockerClientPath @dockerArguments
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw ('The fixed g0-admission-helper exited with code ' + $exitCode + '.')
    }
    Assert-HubQaHelperAdmissionReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    return $closed.ReceiptPath
}
