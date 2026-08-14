Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

# hub.qa-bootstrap/v1 is a founder-released, bootstrap-private entrypoint.  It
# is intentionally a dot-source-only library; invoking this file is not a
# substitute for the committed Package-R authority.
$rawInvocationArguments = @($MyInvocation.UnboundArguments)
if ($MyInvocation.InvocationName -ne '.' -or $rawInvocationArguments.Count -ne 0) {
    [Console]::Error.WriteLine('E_B0_DIRECT_INVOCATION_FORBIDDEN')
    exit 64
}

$script:HubQaBootstrapTask = 'g0-assets-runner'
$script:HubQaBootstrapSourceCommit = '61b38573682d7c01e6d9381babc5085537d7990f'
$script:HubQaBootstrapReleasePath = '.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json'
$script:HubQaBootstrapRunPattern = '^[a-z0-9][a-z0-9-]{7,63}$'
$script:HubQaBootstrapTaskPattern = '^[a-z][a-z0-9-]{2,31}$'
$script:HubQaBootstrapFixedTargets = @('backend-unit', 'frontend-unit', 'backend-openapi', 'g0-planning-selftest')
$script:HubQaBootstrapFirstExit = 0
$script:HubQaBootstrapOwnedResources = New-Object 'System.Collections.Generic.List[object]'
$script:HubQaBootstrapOwnedContainers = New-Object 'System.Collections.Generic.List[object]'
$script:HubQaBootstrapOwnedImageIds = @{}
$script:HubQaBootstrapOwnedContainerIds = @{}
$script:HubQaBootstrapCommandSeconds = 180
$script:HubQaBootstrapTargetSeconds = 780
$script:HubQaBootstrapInspectSeconds = 30
$script:HubQaBootstrapOuterSeconds = 840
$script:HubQaBootstrapCleanupSeconds = 30
$script:HubQaBootstrapOrchestrationSlackSeconds = 30
$script:HubQaBootstrapMaxIngressAssetBytes = [int64]2147483648
$script:HubQaBootstrapMaxArchiveMembers = [int64]200000
$script:HubQaBootstrapMaxExpandedAssetBytes = [int64]4294967296
$script:HubQaBootstrapMaxControlBytes = [int64]4194304
$script:HubQaBootstrapRuntimeStopwatch = $null
$script:HubQaBootstrapTargetBudgetStopwatch = $null
$script:HubQaBootstrapCleanupBudgetStopwatch = $null
$script:HubQaBootstrapCommandReceipts = New-Object 'System.Collections.Generic.List[object]'
$script:HubQaBootstrapAcceptedA0Commit = '8626c38d3d6a086f44aadddd1e3440e73cd8dd1a'
$script:HubQaBootstrapA01Paths = @(
    'testframework/docker/dependency-bundle.schema.v1.json',
    'testframework/docker/admission/verify_local_dependency_bundle.py',
    'testframework/docker/admission/admit_local_dependency_assets.py',
    'testframework/docker/admit-local-dependency-assets.ps1',
    'testframework/docker/Dockerfile.backend-unit',
    'testframework/docker/Dockerfile.frontend-unit',
    'testframework/docker/Dockerfile.backend-openapi',
    'testframework/docker/entrypoints/frontend-unit.sh',
    'testframework/docker/runner-fixtures/g0-planning-selftest.py',
    'testframework/docker/bootstrap-control.v1.json',
    'testframework/docker/source-image-attestor.ps1',
    'testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1',
    'testframework/selftests/runner/g0-image-contract.ps1'
)
$script:HubQaBootstrapA02Paths = @(
    'testframework/docker/dependency-bundle.schema.v1.json',
    'testframework/docker/admission/verify_local_dependency_bundle.py',
    'testframework/docker/admit-local-dependency-assets.ps1',
    'testframework/docker/entrypoints/backend-unit.sh',
    'testframework/docker/vitest.frontend-unit.config.ts',
    'testframework/docker/bootstrap-control.v1.json',
    'testframework/docker/source-image-attestor.ps1',
    'testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1',
    'testframework/selftests/runner/g0-image-contract.ps1'
)
$script:HubQaBootstrapA02OverlapPaths = @(
    'testframework/docker/dependency-bundle.schema.v1.json',
    'testframework/docker/admission/verify_local_dependency_bundle.py',
    'testframework/docker/admit-local-dependency-assets.ps1',
    'testframework/docker/bootstrap-control.v1.json',
    'testframework/docker/source-image-attestor.ps1',
    'testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1',
    'testframework/selftests/runner/g0-image-contract.ps1'
)
$script:HubQaBootstrapA01RemainingPaths = @(
    'testframework/docker/admission/admit_local_dependency_assets.py',
    'testframework/docker/Dockerfile.backend-unit',
    'testframework/docker/Dockerfile.frontend-unit',
    'testframework/docker/Dockerfile.backend-openapi',
    'testframework/docker/entrypoints/frontend-unit.sh',
    'testframework/docker/runner-fixtures/g0-planning-selftest.py'
)
$script:HubQaBootstrapCReceiptKeys = @(
    'nonce_consumption', 'owner_marker', 'local_base_inspect', 'helper_base_verification',
    'helper_image_attestation', 'ingress_copy', 'helper_admission', 'helper_cleanup',
    'source_tree_verification', 'local_image_inspection', 'candidate_lock',
    'planning_selftest_negative_red', 'planning_selftest_positive_green', 'command_receipts',
    'run', 'cleanup'
)
$script:HubQaBootstrapA1Paths = @(
    'testframework/docker/dependency-bundle.manifest.v1.json',
    'testframework/docker/source-policy.v2.json',
    'testframework/docker/build-manifest.v3.json'
)
$script:HubQaBootstrapBPaths = @(
    'testframework/runner/contracts/hub.qa-bootstrap.v1.json',
    'testframework/runner/invoke-qa-bootstrap-v1.ps1',
    'testframework/runner/invoke-qa-gate-v2.ps1',
    'testframework/runner/run-task-v2.ps1',
    'testframework/runner/new-repository-snapshot.ps1',
    'testframework/runner/new-run-envelope.ps1',
    'testframework/runner/write-g0-activation.ps1',
    'testframework/schemas/repository-snapshot.schema.v1.json',
    'testframework/schemas/run-envelope.schema.v1.json',
    'testframework/selftests/runner/g0-runner-v2-contract.ps1',
    'testframework/runner/README.md',
    '.cursor/skills/qa-gate/SKILL.md'
)

# Package-R and the accepted private sources must be read safely before the
# A0 implementation can be imported.  Keep the minimal native handle surface
# here so bootstrap authority never starts with a path-based ReadAllBytes.
if ($null -eq ('HubQaG0NativeFile' -as [type])) {
    Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Diagnostics;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;

public static class HubQaG0NativeFile
{
    public const UInt32 GENERIC_READ = 0x80000000;
    public const UInt32 FILE_SHARE_READ = 0x00000001;
    public const UInt32 FILE_SHARE_WRITE = 0x00000002;
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

public sealed class HubQaG0AssignedProcess : IDisposable
{
    private const UInt32 CREATE_SUSPENDED = 0x00000004;
    private const UInt32 CREATE_NO_WINDOW = 0x08000000;
    private const UInt32 STARTF_USESTDHANDLES = 0x00000100;
    private const UInt32 HANDLE_FLAG_INHERIT = 0x00000001;
    private const UInt32 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;
    private const Int32 JobObjectExtendedLimitInformation = 9;

    [StructLayout(LayoutKind.Sequential)]
    private struct SECURITY_ATTRIBUTES
    {
        public Int32 nLength;
        public IntPtr lpSecurityDescriptor;
        [MarshalAs(UnmanagedType.Bool)] public bool bInheritHandle;
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    private struct STARTUPINFO
    {
        public Int32 cb;
        public string lpReserved;
        public string lpDesktop;
        public string lpTitle;
        public UInt32 dwX;
        public UInt32 dwY;
        public UInt32 dwXSize;
        public UInt32 dwYSize;
        public UInt32 dwXCountChars;
        public UInt32 dwYCountChars;
        public UInt32 dwFillAttribute;
        public UInt32 dwFlags;
        public UInt16 wShowWindow;
        public UInt16 cbReserved2;
        public IntPtr lpReserved2;
        public IntPtr hStdInput;
        public IntPtr hStdOutput;
        public IntPtr hStdError;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct PROCESS_INFORMATION
    {
        public IntPtr hProcess;
        public IntPtr hThread;
        public UInt32 dwProcessId;
        public UInt32 dwThreadId;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_LIMIT_INFORMATION
    {
        public Int64 PerProcessUserTimeLimit;
        public Int64 PerJobUserTimeLimit;
        public UInt32 LimitFlags;
        public UIntPtr MinimumWorkingSetSize;
        public UIntPtr MaximumWorkingSetSize;
        public UInt32 ActiveProcessLimit;
        public UIntPtr Affinity;
        public UInt32 PriorityClass;
        public UInt32 SchedulingClass;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IO_COUNTERS
    {
        public UInt64 ReadOperationCount;
        public UInt64 WriteOperationCount;
        public UInt64 OtherOperationCount;
        public UInt64 ReadTransferCount;
        public UInt64 WriteTransferCount;
        public UInt64 OtherTransferCount;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION
    {
        public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation;
        public IO_COUNTERS IoInfo;
        public UIntPtr ProcessMemoryLimit;
        public UIntPtr JobMemoryLimit;
        public UIntPtr PeakProcessMemoryUsed;
        public UIntPtr PeakJobMemoryUsed;
    }

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CreatePipe(out IntPtr readPipe, out IntPtr writePipe, ref SECURITY_ATTRIBUTES attributes, UInt32 size);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetHandleInformation(IntPtr handle, UInt32 mask, UInt32 flags);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool CreateProcessW(
        string applicationName,
        StringBuilder commandLine,
        IntPtr processAttributes,
        IntPtr threadAttributes,
        bool inheritHandles,
        UInt32 creationFlags,
        IntPtr environment,
        string currentDirectory,
        ref STARTUPINFO startupInfo,
        out PROCESS_INFORMATION processInformation);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateJobObjectW(IntPtr attributes, string name);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetInformationJobObject(IntPtr job, Int32 informationClass, ref JOBOBJECT_EXTENDED_LIMIT_INFORMATION information, UInt32 informationLength);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern UInt32 ResumeThread(IntPtr thread);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool TerminateJobObject(IntPtr job, UInt32 exitCode);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool TerminateProcess(IntPtr process, UInt32 exitCode);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);

    public System.Diagnostics.Process Process { get; private set; }
    public SafeFileHandle StandardOutputRead { get; private set; }
    public SafeFileHandle StandardErrorRead { get; private set; }
    public bool JobClosed { get { return _job == IntPtr.Zero; } }

    private IntPtr _job;

    private HubQaG0AssignedProcess() { }

    private static void CloseIfValid(ref IntPtr handle)
    {
        if (handle != IntPtr.Zero && handle != new IntPtr(-1))
        {
            CloseHandle(handle);
            handle = IntPtr.Zero;
        }
    }

    public static HubQaG0AssignedProcess Start(string applicationName, string commandLine, string currentDirectory)
    {
        if (String.IsNullOrWhiteSpace(applicationName) || String.IsNullOrWhiteSpace(commandLine) || String.IsNullOrWhiteSpace(currentDirectory))
            throw new InvalidOperationException("E_B0_PROCESS_JOB");

        SECURITY_ATTRIBUTES security = new SECURITY_ATTRIBUTES();
        security.nLength = Marshal.SizeOf(typeof(SECURITY_ATTRIBUTES));
        security.bInheritHandle = true;
        IntPtr stdoutRead = IntPtr.Zero, stdoutWrite = IntPtr.Zero;
        IntPtr stderrRead = IntPtr.Zero, stderrWrite = IntPtr.Zero;
        IntPtr stdinRead = IntPtr.Zero, stdinWrite = IntPtr.Zero;
        IntPtr job = IntPtr.Zero;
        PROCESS_INFORMATION information = new PROCESS_INFORMATION();
        System.Diagnostics.Process managedProcess = null;
        SafeFileHandle managedStdout = null;
        SafeFileHandle managedStderr = null;
        bool created = false;
        bool assigned = false;
        try
        {
            if (!CreatePipe(out stdoutRead, out stdoutWrite, ref security, 0) ||
                !CreatePipe(out stderrRead, out stderrWrite, ref security, 0) ||
                !CreatePipe(out stdinRead, out stdinWrite, ref security, 0))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            if (!SetHandleInformation(stdoutRead, HANDLE_FLAG_INHERIT, 0) ||
                !SetHandleInformation(stderrRead, HANDLE_FLAG_INHERIT, 0) ||
                !SetHandleInformation(stdinWrite, HANDLE_FLAG_INHERIT, 0))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());

            job = CreateJobObjectW(IntPtr.Zero, null);
            if (job == IntPtr.Zero) throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            JOBOBJECT_EXTENDED_LIMIT_INFORMATION limits = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
            limits.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
            if (!SetInformationJobObject(job, JobObjectExtendedLimitInformation, ref limits, (UInt32)Marshal.SizeOf(typeof(JOBOBJECT_EXTENDED_LIMIT_INFORMATION))))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());

            STARTUPINFO startup = new STARTUPINFO();
            startup.cb = Marshal.SizeOf(typeof(STARTUPINFO));
            startup.dwFlags = STARTF_USESTDHANDLES;
            startup.hStdInput = stdinRead;
            startup.hStdOutput = stdoutWrite;
            startup.hStdError = stderrWrite;
            StringBuilder mutableCommandLine = new StringBuilder(commandLine);
            if (!CreateProcessW(applicationName, mutableCommandLine, IntPtr.Zero, IntPtr.Zero, true, CREATE_SUSPENDED | CREATE_NO_WINDOW, IntPtr.Zero, currentDirectory, ref startup, out information))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            created = true;
            managedProcess = System.Diagnostics.Process.GetProcessById((Int32)information.dwProcessId);
            if (!AssignProcessToJobObject(job, information.hProcess))
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            assigned = true;

            CloseIfValid(ref stdoutWrite);
            CloseIfValid(ref stderrWrite);
            CloseIfValid(ref stdinRead);
            CloseIfValid(ref stdinWrite);
            managedStdout = new SafeFileHandle(stdoutRead, true);
            stdoutRead = IntPtr.Zero;
            managedStderr = new SafeFileHandle(stderrRead, true);
            stderrRead = IntPtr.Zero;
            if (ResumeThread(information.hThread) == UInt32.MaxValue)
                throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
            HubQaG0AssignedProcess result = new HubQaG0AssignedProcess();
            result.Process = managedProcess;
            managedProcess = null;
            result.StandardOutputRead = managedStdout;
            managedStdout = null;
            result.StandardErrorRead = managedStderr;
            managedStderr = null;
            result._job = job;
            job = IntPtr.Zero;
            return result;
        }
        catch
        {
            if (assigned && job != IntPtr.Zero) TerminateJobObject(job, 124);
            // This branch can touch only a CREATE_SUSPENDED process that was
            // never assigned and never resumed. It is setup rollback, not a
            // timeout/runtime fallback; every resumed process is job-owned.
            else if (created && information.hProcess != IntPtr.Zero) TerminateProcess(information.hProcess, 124);
            throw;
        }
        finally
        {
            CloseIfValid(ref information.hThread);
            CloseIfValid(ref information.hProcess);
            CloseIfValid(ref stdoutRead);
            CloseIfValid(ref stdoutWrite);
            CloseIfValid(ref stderrRead);
            CloseIfValid(ref stderrWrite);
            CloseIfValid(ref stdinRead);
            CloseIfValid(ref stdinWrite);
            CloseIfValid(ref job);
            if (managedProcess != null) managedProcess.Dispose();
            if (managedStdout != null) managedStdout.Dispose();
            if (managedStderr != null) managedStderr.Dispose();
        }
    }

    public void Terminate(UInt32 exitCode)
    {
        if (_job == IntPtr.Zero || !TerminateJobObject(_job, exitCode))
            throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
    }

    public void CloseJob()
    {
        if (_job == IntPtr.Zero) return;
        IntPtr closing = _job;
        _job = IntPtr.Zero;
        if (!CloseHandle(closing)) throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());
    }

    public void Dispose()
    {
        if (StandardOutputRead != null) StandardOutputRead.Dispose();
        if (StandardErrorRead != null) StandardErrorRead.Dispose();
        if (Process != null) Process.Dispose();
        if (_job != IntPtr.Zero)
        {
            IntPtr closing = _job;
            _job = IntPtr.Zero;
            CloseHandle(closing);
        }
    }
}
'@
}
if ($null -eq ('HubQaG0NativeFile' -as [type]) -or $null -eq ('HubQaG0AssignedProcess' -as [type])) {
    throw 'E_B0_PROCESS_JOB'
}

function New-HubQaOwnedDuplicateFileHandle {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    $process = [HubQaG0NativeFile]::GetCurrentProcess()
    $duplicate = $null
    if (-not [HubQaG0NativeFile]::DuplicateHandle(
        $process, $Handle, $process, [ref]$duplicate, [uint32]0, $false,
        [HubQaG0NativeFile]::DUPLICATE_SAME_ACCESS
    )) { throw $Identity }
    if ($null -eq $duplicate -or $duplicate.IsInvalid) {
        if ($null -ne $duplicate) { $duplicate.Dispose() }
        throw $Identity
    }
    return $duplicate
}

function Get-HubQaBootstrapRepositoryRoot {
    $root = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
    $directory = [System.IO.DirectoryInfo]::new($root)
    $directory.Refresh()
    if (-not $directory.Exists -or (($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) {
        throw 'E_B0_RELEASE_AUTHORITY'
    }
    return $directory.FullName.TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
}

function Get-HubQaBootstrapSha256 {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][byte[]]$Bytes)

    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        return 'sha256:' + ([System.BitConverter]::ToString($algorithm.ComputeHash($Bytes)).Replace('-', '').ToLowerInvariant())
    }
    finally {
        $algorithm.Dispose()
    }
}

function Test-HubQaBootstrapBytesEqual {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][byte[]]$Left,
        [Parameter(Mandatory = $true)][byte[]]$Right
    )

    if ($Left.Length -ne $Right.Length) { return $false }
    for ($index = 0; $index -lt $Left.Length; $index++) {
        if ($Left[$index] -ne $Right[$index]) { return $false }
    }
    return $true
}

function Assert-HubQaBootstrapExactKeys {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Value,
        [Parameter(Mandatory = $true)][string[]]$Expected,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($null -eq $Value) { throw $Code }
    $actual = if ($Value -is [System.Collections.IDictionary]) {
        @($Value.Keys)
    }
    else {
        @($Value.PSObject.Properties | ForEach-Object { $_.Name })
    }
    if ($actual.Count -ne $Expected.Count) { throw $Code }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ($actual[$index] -cne $Expected[$index]) { throw $Code }
    }
}

function Read-HubQaBootstrapRawFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$RelativePath,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($RelativePath -notmatch '^[A-Za-z0-9._/-]+$' -or $RelativePath.StartsWith('/') -or $RelativePath.Contains('..')) {
        throw $Code
    }
    $rootFull = [System.IO.Path]::GetFullPath($Root).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $path = [System.IO.Path]::GetFullPath((Join-Path $rootFull ($RelativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))))
    $opened = Open-HubQaBootstrapSafeRegularFile -Path $path -Root $rootFull -MaximumBytes $script:HubQaBootstrapMaxControlBytes -Code $Code
    try {
        $raw = Read-HubQaBootstrapHeldFileBytes -Opened $opened -MaximumBytes $script:HubQaBootstrapMaxControlBytes -Code $Code
        [byte[]]$bytes = [byte[]]$raw.Bytes
        if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf) { throw $Code }
        try { $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($bytes) }
        catch { throw $Code }
        return [pscustomobject][ordered]@{
            Path = [string]$opened.FinalPath
            Bytes = $bytes
            Text = $text
            Sha256 = [string]$raw.Sha256
        }
    }
    finally {
        Close-HubQaBootstrapSafeRegularFile -Opened $opened
    }
}

function Read-HubQaBootstrapHeldFileBytes {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Opened,
        [Parameter(Mandatory = $true)][int64]$MaximumBytes,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $pre = Get-HubQaBootstrapHandleObservation -Handle $Opened.Handle -Code $Code
    Assert-HubQaBootstrapSameHandleObservation -Expected $Opened.Observation -Actual $pre -Code $Code
    if ([uint64]$pre.Length -lt 1 -or [uint64]$pre.Length -gt [uint64]$MaximumBytes -or [uint64]$pre.Length -gt [uint64][int]::MaxValue) { throw $Code }
    [byte[]]$bytes = New-Object byte[] ([int]$pre.Length)
    $duplicate = New-HubQaOwnedDuplicateFileHandle -Handle $Opened.Handle -Identity $Code
    $stream = $null
    try {
        $stream = [System.IO.FileStream]::new($duplicate, [System.IO.FileAccess]::Read, 65536, $false)
        $duplicate = $null
        $offset = 0
        while ($offset -lt $bytes.Length) {
            $read = $stream.Read($bytes, $offset, $bytes.Length - $offset)
            if ($read -le 0) { throw $Code }
            $offset += $read
        }
        if ($stream.ReadByte() -ne -1) { throw $Code }
        $post = Get-HubQaBootstrapHandleObservation -Handle $Opened.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $pre -Actual $post -Code $Code
        return [pscustomobject][ordered]@{
            Bytes = $bytes
            Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
            Observation = $post
        }
    }
    finally {
        if ($null -ne $stream) { $stream.Dispose() }
        elseif ($null -ne $duplicate) { $duplicate.Dispose() }
    }
}

function ConvertFrom-HubQaBootstrapClosedJson {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Text,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $objectKeySets = New-Object 'System.Collections.Generic.Stack[object]'
    $index = 0
    while ($index -lt $Text.Length) {
        $character = $Text[$index]
        if ([char]::IsWhiteSpace($character)) { throw $Code }
        if ($character -eq '"') {
            $start = $index
            $index++
            $innerEscaped = $false
            while ($index -lt $Text.Length) {
                if ($innerEscaped) { $innerEscaped = $false; $index++; continue }
                if ($Text[$index] -eq '\') { $innerEscaped = $true; $index++; continue }
                if ($Text[$index] -eq '"') { break }
                if ([int][char]$Text[$index] -lt 0x20) { throw $Code }
                $index++
            }
            if ($index -ge $Text.Length) { throw $Code }
            $rawString = $Text.Substring($start, $index - $start + 1)
            $next = $index + 1
            if ($next -lt $Text.Length -and $Text[$next] -eq ':') {
                if ($objectKeySets.Count -eq 0 -or $null -eq $objectKeySets.Peek()) { throw $Code }
                try { $key = $rawString | ConvertFrom-Json -ErrorAction Stop } catch { throw $Code }
                if (-not $objectKeySets.Peek().Add([string]$key)) { throw $Code }
            }
            $index++
            continue
        }
        if ($character -eq '{') {
            $objectKeySets.Push([System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal))
        }
        elseif ($character -eq '[') {
            $objectKeySets.Push($null)
        }
        elseif ($character -eq '}' -or $character -eq ']') {
            if ($objectKeySets.Count -eq 0) { throw $Code }
            [void]$objectKeySets.Pop()
        }
        $index++
    }
    if ($objectKeySets.Count -ne 0) { throw $Code }
    try {
        $value = $Text | ConvertFrom-Json -ErrorAction Stop
        $value = ConvertTo-HubQaBootstrapClosedJsonValue -Value $value -Code $Code
        # This pre-import parser protects Package-R and A1 controls before the
        # later shared codec may be dot-sourced.  Its accepted subset is the
        # exact compact JSON serialization PowerShell will re-emit: no BOM,
        # no whitespace, no duplicate keys, no alternate escaping, and no
        # numeric spelling aliases can silently become release authority.
        if (($value | ConvertTo-Json -Depth 64 -Compress) -cne $Text) { throw $Code }
        return $value
    }
    catch {
        throw $Code
    }
}

function ConvertTo-HubQaBootstrapClosedJsonValue {
    [CmdletBinding()]
    param(
        [AllowNull()][object]$Value,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($null -eq $Value) { return $null }
    if ($Value -is [string] -or $Value -is [bool]) { return $Value }
    if ($Value -is [int64]) { return $Value }
    if (
        $Value -is [byte] -or $Value -is [sbyte] -or
        $Value -is [int16] -or $Value -is [uint16] -or
        $Value -is [int32] -or $Value -is [uint32]
    ) { return [int64]$Value }
    if ($Value -is [uint64]) {
        if ([uint64]$Value -gt [uint64][int64]::MaxValue) { throw $Code }
        return [int64]$Value
    }
    if ($Value -is [System.Array]) {
        [object[]]$items = New-Object object[] $Value.Count
        for ($itemIndex = 0; $itemIndex -lt $Value.Count; $itemIndex++) {
            $items[$itemIndex] = ConvertTo-HubQaBootstrapClosedJsonValue -Value $Value[$itemIndex] -Code $Code
        }
        Write-Output -NoEnumerate $items
        return
    }
    if ($Value -is [pscustomobject]) {
        foreach ($property in @($Value.PSObject.Properties)) {
            $property.Value = ConvertTo-HubQaBootstrapClosedJsonValue -Value $property.Value -Code $Code
        }
        return $Value
    }
    # ConvertFrom-Json can otherwise surface Double/Decimal values.  The
    # authority subset has only signed integral JSON numbers, normalized to
    # Int64 before any field validator observes them.
    throw $Code
}

function Invoke-HubQaBootstrapGit {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $gitCommand = Get-Command git -CommandType Application -ErrorAction SilentlyContinue
    if ($null -eq $gitCommand -or [string]::IsNullOrWhiteSpace([string]$gitCommand.Source)) { throw $Code }
    $result = Invoke-HubQaBootstrapProcess -Name 'git' -FileName ([string]$gitCommand.Source) -Arguments (@('-C', $Root) + $Arguments) -TimeoutSeconds $script:HubQaBootstrapInspectSeconds -Code $Code
    if ([bool]$result.TimedOut) { throw 'E_B0_TIMEOUT_124' }
    if ([int]$result.ExitCode -ne 0) { throw $Code }
    if ([string]::IsNullOrEmpty([string]$result.StandardOutput)) { return @() }
    return @(([string]$result.StandardOutput -split "`r?`n") | Where-Object { $_.Length -gt 0 })
}

function Get-HubQaBootstrapAttestorGitBlobBytes {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Repository,
        [Parameter(Mandatory = $true)][string]$BlobSha1,
        [Parameter(Mandatory = $true)][string]$Identity
    )

    if ($BlobSha1 -notmatch '^[a-f0-9]{40}$' -or [string]::IsNullOrWhiteSpace($Identity)) { throw 'E_SOURCE_TREE_INVALID' }
    return Get-HubQaBootstrapGitObjectBytes -Root $Repository -ObjectName $BlobSha1 -Code 'E_SOURCE_TREE_INVALID'
}

function Get-HubQaBootstrapAttestorGitTreeEntry {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Repository,
        [Parameter(Mandatory = $true)][string]$RepositoryRelativePath
    )

    if ($RepositoryRelativePath -notmatch '^[A-Za-z0-9._/-]+$' -or $RepositoryRelativePath.StartsWith('/') -or $RepositoryRelativePath.Contains('..')) {
        throw 'E_SOURCE_TREE_INVALID: source-policy path is unsafe.'
    }
    $lines = @(Invoke-HubQaBootstrapGit -Root $Repository -Arguments @('ls-tree', $script:HubQaBootstrapSourceCommit, '--', $RepositoryRelativePath) -Code 'E_SOURCE_TREE_INVALID')
    if ($lines.Count -ne 1 -or [string]$lines[0] -notmatch '^(?<mode>[0-7]{6}) (?<type>blob|tree|commit) (?<object>[0-9a-f]{40})\t(?<path>.+)$') {
        throw "E_SOURCE_TREE_INVALID: malformed or missing Git tree entry $RepositoryRelativePath."
    }
    if (
        -not [System.StringComparer]::Ordinal.Equals([string]$Matches.path, $RepositoryRelativePath) -or
        [string]$Matches.mode -cne '100644' -or [string]$Matches.type -cne 'blob'
    ) { throw "E_SOURCE_TREE_INVALID: source entry is not a closed mode-100644 blob $RepositoryRelativePath." }
    return [pscustomobject][ordered]@{ Mode = [string]$Matches.mode; BlobSha1 = [string]$Matches.object }
}

function Get-HubQaBootstrapAttestorGitObjectSourceTree {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][object]$Controls)

    $policy = $Controls.source.Value
    Assert-HubQaExactPropertyNames -Object $policy -Expected @('schema', 'source_commit', 'source_tree_sha256', 'allowlist') -Identity 'source policy'
    if ($policy.schema -cne 'hub.qa.local-source-policy/v2' -or $policy.source_commit -cne $script:HubQaBootstrapSourceCommit) {
        throw 'E_SOURCE_TREE_INVALID: source policy baseline is not closed.'
    }
    Assert-HubQaDigest -Value ([string]$policy.source_tree_sha256) -Identity 'source policy source tree digest'
    if ($null -eq $policy.allowlist -or $policy.allowlist -is [string]) { throw 'E_SOURCE_TREE_INVALID: source policy allowlist is invalid.' }
    $root = Get-HubQaSourceRoot
    $commitType = @(Invoke-HubQaBootstrapGit -Root $root -Arguments @('cat-file', '-t', $script:HubQaBootstrapSourceCommit) -Code 'E_SOURCE_TREE_INVALID')
    if ($commitType.Count -ne 1 -or [string]$commitType[0] -cne 'commit') { throw 'E_SOURCE_TREE_INVALID: recovery source commit is unavailable.' }

    $seenPaths = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    $linesByPath = @{}
    foreach ($allowlistEntry in @($policy.allowlist)) {
        Assert-HubQaExactPropertyNames -Object $allowlistEntry -Expected @('path', 'blob_sha1') -Identity 'source policy allowlist entry'
        $path = [string]$allowlistEntry.path
        $declaredBlobSha1 = [string]$allowlistEntry.blob_sha1
        if (-not $seenPaths.Add($path) -or $declaredBlobSha1 -notmatch '^[a-f0-9]{40}$') { throw "E_SOURCE_TREE_INVALID: duplicate or invalid source path $path." }
        $treeEntry = Get-HubQaBootstrapAttestorGitTreeEntry -Repository $root -RepositoryRelativePath $path
        if (-not [System.StringComparer]::Ordinal.Equals([string]$treeEntry.BlobSha1, $declaredBlobSha1)) { throw "E_SOURCE_TREE_INVALID: Git blob SHA-1 mismatch $path." }
        [byte[]]$blobBytes = Get-HubQaBootstrapAttestorGitBlobBytes -Repository $root -BlobSha1 $treeEntry.BlobSha1 -Identity "source policy blob $path"
        $line = $path + "`t100644`t" + $treeEntry.BlobSha1 + "`t" + (Get-HubQaSha256 -Bytes $blobBytes) + "`n"
        if ($linesByPath.ContainsKey($path)) { throw "E_SOURCE_TREE_INVALID: UTF-8 source path collision $path." }
        $linesByPath[$path] = $line
    }
    if ($linesByPath.Count -eq 0) { throw 'E_SOURCE_TREE_INVALID: source policy allowlist is empty.' }
    [string[]]$orderedPaths = @($linesByPath.Keys)
    [array]::Sort($orderedPaths, [HubQaUtf8BytePathComparer]::new())
    $canonical = New-Object System.Text.StringBuilder
    foreach ($path in $orderedPaths) { [void]$canonical.Append([string]$linesByPath[$path]) }
    $computed = Get-HubQaSha256 -Bytes $script:Utf8.GetBytes($canonical.ToString())
    if ($computed -cne [string]$policy.source_tree_sha256 -or $computed -cne [string]$Controls.build.Value.source_tree_sha256) {
        throw 'E_SOURCE_TREE_DIGEST_MISMATCH: canonical Git object source tree digest disagrees with A1 controls.'
    }
    return [pscustomobject][ordered]@{ SourceTreeSha256 = $computed; SourceCommit = $script:HubQaBootstrapSourceCommit; Entries = $linesByPath.Count }
}

function Get-HubQaBootstrapGitObjectBytes {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$ObjectName,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $gitCommand = Get-Command git -CommandType Application -ErrorAction SilentlyContinue
    if ($null -eq $gitCommand -or [string]::IsNullOrWhiteSpace([string]$gitCommand.Source)) { throw $Code }
    $arguments = @('cat-file', 'blob', $ObjectName)
    $operationBudgetMilliseconds = Get-HubQaBootstrapEffectiveTimeoutMilliseconds -NominalSeconds $script:HubQaBootstrapInspectSeconds
    $operationStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $nativeProcess = $null
    $stdoutStream = $null
    $stderrStream = $null
    $stdoutMemory = New-Object System.IO.MemoryStream
    $stderrMemory = New-Object System.IO.MemoryStream
    $timedOut = $false
    $startedAtUtc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    try {
        $nativeProcess = Start-HubQaBootstrapAssignedProcess -FileName ([string]$gitCommand.Source) -Arguments $arguments -WorkingDirectory $Root -Code $Code
        $process = $nativeProcess.Process
        $stdoutStream = [System.IO.FileStream]::new($nativeProcess.StandardOutputRead, [System.IO.FileAccess]::Read, 65536, $false)
        $stderrStream = [System.IO.FileStream]::new($nativeProcess.StandardErrorRead, [System.IO.FileAccess]::Read, 65536, $false)
        $stdoutTask = $stdoutStream.CopyToAsync($stdoutMemory)
        $stderrTask = $stderrStream.CopyToAsync($stderrMemory)
        $waitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds -ReserveForTermination
        if (-not $process.WaitForExit($waitMilliseconds)) {
            $timedOut = $true
            Stop-HubQaBootstrapProcessJobWithinDeadline -NativeProcess $nativeProcess -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        }
        else {
            $nativeProcess.CloseJob()
        }
        $stdoutWaitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        if (-not $stdoutTask.Wait($stdoutWaitMilliseconds)) { throw 'E_B0_TIMEOUT_124' }
        $stderrWaitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        if (-not $stderrTask.Wait($stderrWaitMilliseconds)) { throw 'E_B0_TIMEOUT_124' }
        [byte[]]$stdoutBytes = $stdoutMemory.ToArray()
        [byte[]]$stderrBytes = $stderrMemory.ToArray()
        Add-HubQaBootstrapCommandReceiptBytes -Name 'git-cat-file' -Arguments $arguments -ExitCode ([int]$process.ExitCode) -TimedOut $timedOut -StartedAtUtc $startedAtUtc -FinishedAtUtc ([DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)) -StandardOutputBytes $stdoutBytes -StandardErrorBytes $stderrBytes
        if ($timedOut) { throw 'E_B0_TIMEOUT_124' }
        if ($process.ExitCode -ne 0) { throw $Code }
        return $stdoutBytes
    }
    finally {
        try {
            if ($null -ne $nativeProcess -and -not $nativeProcess.JobClosed) {
                if (-not $nativeProcess.Process.HasExited) {
                    Stop-HubQaBootstrapProcessJobWithinDeadline -NativeProcess $nativeProcess -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
                }
                else { $nativeProcess.CloseJob() }
            }
        }
        finally {
            $operationStopwatch.Stop()
            if ($null -ne $stdoutStream) { $stdoutStream.Dispose() }
            if ($null -ne $stderrStream) { $stderrStream.Dispose() }
            $stdoutMemory.Dispose()
            $stderrMemory.Dispose()
            if ($null -ne $nativeProcess) { $nativeProcess.Dispose() }
        }
    }
}

function Get-HubQaBootstrapCommitPaths {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($Commit -notmatch '^[a-f0-9]{40}$') { throw $Code }
    $parent = Get-HubQaBootstrapActualFirstParent -Root $Root -Commit $Commit -Code $Code
    return @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('diff-tree', '--no-commit-id', '--name-only', '--no-renames', '-r', $parent, $Commit) -Code $Code)
}

function Get-HubQaBootstrapActualFirstParent {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($Commit -notmatch '^[a-f0-9]{40}$') { throw $Code }
    $parents = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('show', '-s', '--format=%P', $Commit) -Code $Code)
    if ($parents.Count -ne 1) { throw $Code }
    $parentTokens = @(([string]$parents[0]).Split(' ', [System.StringSplitOptions]::RemoveEmptyEntries))
    if ($parentTokens.Count -lt 1 -or [string]$parentTokens[0] -notmatch '^[a-f0-9]{40}$') { throw $Code }
    return [string]$parentTokens[0]
}

function Assert-HubQaBootstrapExactCommitPaths {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string[]]$Expected,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $actual = @(Get-HubQaBootstrapCommitPaths -Root $Root -Commit $Commit -Code $Code)
    if ($actual.Count -ne $Expected.Count) { throw $Code }
    foreach ($path in $Expected) {
        if ($actual -cnotcontains $path) { throw $Code }
    }
}

function Assert-HubQaBootstrapA01FirstParent {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$A01Commit
    )

    Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit $A01Commit -Expected $script:HubQaBootstrapA01Paths -Code 'E_B0_RELEASE_AUTHORITY'
    $firstParent = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('rev-list', '--first-parent', $A01Commit) -Code 'E_B0_RELEASE_AUTHORITY')
    if ($firstParent -cnotcontains $script:HubQaBootstrapAcceptedA0Commit) { throw 'E_B0_RELEASE_AUTHORITY' }
}

function Assert-HubQaBootstrapA02Authority {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$A01Commit,
        [Parameter(Mandatory = $true)][string]$A02Commit
    )

    if ($A01Commit -cne '56a288dd51e261bbb91ed693d7b58a8a97994faf' -or $A02Commit -notmatch '^[a-f0-9]{40}$') {
        throw 'E_B0_RELEASE_AUTHORITY'
    }
    $a02FirstParentHistory = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('rev-list', '--first-parent', $A02Commit) -Code 'E_B0_RELEASE_AUTHORITY')
    if ($a02FirstParentHistory -cnotcontains $A01Commit) { throw 'E_B0_RELEASE_AUTHORITY' }
    $actualParent = Get-HubQaBootstrapActualFirstParent -Root $Root -Commit $A02Commit -Code 'E_B0_RELEASE_AUTHORITY'
    $actualParentHistory = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('rev-list', '--first-parent', $actualParent) -Code 'E_B0_RELEASE_AUTHORITY')
    if ($actualParentHistory -cnotcontains $script:HubQaBootstrapSourceCommit) { throw 'E_B0_RELEASE_AUTHORITY' }
    Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit $A02Commit -Expected $script:HubQaBootstrapA02Paths -Code 'E_B0_RELEASE_AUTHORITY'
}

function Get-HubQaBootstrapGitTreeEntry {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $entry = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('ls-tree', $Commit, '--', $Path) -Code $Code)
    if ($entry.Count -ne 1 -or [string]$entry[0] -notmatch '^([0-7]{6}) blob ([a-f0-9]{40})\t(.+)$') { throw $Code }
    if ([string]$Matches[1] -cne '100644' -or [string]$Matches[3] -cne $Path) { throw $Code }
    return [pscustomobject][ordered]@{ Mode = [string]$Matches[1]; Blob = [string]$Matches[2]; Path = [string]$Matches[3] }
}

function Assert-HubQaBootstrapCurrentPackageBytesAndModes {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string[]]$Paths,
        [Parameter(Mandatory = $true)][string]$Code,
        [switch]$SkipExactCommitScope
    )

    # A0.2 is frozen as a nine-path commit and A0.1 as a thirteen-path
    # commit.  Their current-byte selection is deliberately 7+6, so the
    # release authority proves each complete allowlist once and this reader
    # proves only the selected current blobs without falsely shrinking either
    # accepted commit's scope.
    if (-not $SkipExactCommitScope) {
        Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit $Commit -Expected $Paths -Code $Code
    }
    foreach ($path in $Paths) {
        $treeEntry = Get-HubQaBootstrapGitTreeEntry -Root $Root -Commit $Commit -Path $path -Code $Code
        $indexEntry = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('ls-files', '-s', '--', $path) -Code $Code)
        if ($indexEntry.Count -ne 1 -or [string]$indexEntry[0] -notmatch '^([0-7]{6}) ([a-f0-9]{40}) [0-3]\t(.+)$') { throw $Code }
        if ([string]$Matches[1] -cne $treeEntry.Mode -or [string]$Matches[2] -cne $treeEntry.Blob -or [string]$Matches[3] -cne $path) { throw $Code }
        # Compare the accepted raw Git blob to the index blob, then require a
        # clean worktree diff.  This retains raw-byte authority while avoiding
        # a false mismatch from Git's configured CRLF checkout conversion.
        $indexed = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName (':' + $path) -Code $Code
        if ($indexed.Length -gt 4194304) { throw $Code }
        $accepted = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ($Commit + ':' + $path) -Code $Code
        if (-not (Test-HubQaBootstrapBytesEqual -Left $indexed -Right $accepted)) { throw $Code }
        [void](Invoke-HubQaBootstrapGit -Root $Root -Arguments @('diff', '--no-ext-diff', '--exit-code', $Commit, '--', $path) -Code $Code)
    }
}

function Assert-HubQaBootstrapCurrentAcceptedSources {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release
    )

    Assert-HubQaBootstrapCurrentPackageBytesAndModes -Root $Root -Commit ([string]$Release.Value.a02_commit) -Paths $script:HubQaBootstrapA02OverlapPaths -Code 'E_B0_PRIVATE_SOURCE' -SkipExactCommitScope
    Assert-HubQaBootstrapCurrentPackageBytesAndModes -Root $Root -Commit ([string]$Release.Value.a0_commit) -Paths $script:HubQaBootstrapA01RemainingPaths -Code 'E_B0_PRIVATE_SOURCE' -SkipExactCommitScope
    Assert-HubQaBootstrapCurrentPackageBytesAndModes -Root $Root -Commit ([string]$Release.Value.a1_commit) -Paths $script:HubQaBootstrapA1Paths -Code 'E_B0_PRIVATE_SOURCE'
    Assert-HubQaBootstrapCurrentPackageBytesAndModes -Root $Root -Commit ([string]$Release.Value.b_commit) -Paths $script:HubQaBootstrapBPaths -Code 'E_B0_PRIVATE_SOURCE'
}

function Get-HubQaBootstrapAuthorityCommitForPath {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($script:HubQaBootstrapA02Paths -ccontains $Path) {
        return [string]$Release.Value.a02_commit
    }
    if ($script:HubQaBootstrapA01Paths -ccontains $Path) {
        return [string]$Release.Value.a0_commit
    }
    if ($script:HubQaBootstrapBPaths -ccontains $Path) {
        return [string]$Release.Value.b_commit
    }
    return $script:HubQaBootstrapAcceptedA0Commit
}

function Assert-HubQaBootstrapReleaseAuthority {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string]$Root)

    $releaseRaw = Read-HubQaBootstrapRawFile -Root $Root -RelativePath $script:HubQaBootstrapReleasePath -Code 'E_B0_RELEASE_AUTHORITY'
    $release = ConvertFrom-HubQaBootstrapClosedJson -Text $releaseRaw.Text -Code 'E_B0_RELEASE_AUTHORITY'
    $releaseKeys = @(
        'schema', 'release_id', 'g0_plan_path', 'g0_plan_commit', 'a0_commit', 'a02_commit', 'a1_commit', 'b_commit',
        'task', 'run', 'local_only', 'network', 'offline_ingress_handoff_id',
        'offline_ingress_owner_marker_sha256', 'dependency_bundle_manifest_sha256', 'source_commit',
        'expires_at_utc', 'one_shot_nonce', 'founder_approval', 'created_at_utc'
    )
    Assert-HubQaBootstrapExactKeys -Value $release -Expected $releaseKeys -Code 'E_B0_RELEASE_AUTHORITY'
    if (
        [string]$release.schema -cne 'hub.qa.g0-bootstrap-founder-release/v1' -or
        [string]$release.release_id -notmatch '^[a-z][a-z0-9-]{2,63}$' -or
        [string]$release.task -cne $script:HubQaBootstrapTask -or
        [string]$release.run -notmatch $script:HubQaBootstrapRunPattern -or
        $release.local_only -isnot [bool] -or -not [bool]$release.local_only -or
        [string]$release.network -cne 'disabled' -or
        [string]$release.source_commit -cne $script:HubQaBootstrapSourceCommit -or
        [string]$release.offline_ingress_handoff_id -notmatch $script:HubQaBootstrapRunPattern -or
        [string]$release.offline_ingress_owner_marker_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$release.dependency_bundle_manifest_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$release.expires_at_utc -notmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,7})?Z$' -or
        [string]$release.one_shot_nonce -notmatch '^[0-9a-f]{32}$' -or
        [string]$release.founder_approval -cne 'EXPLICIT_LOCAL_ONLY_ONE_SHOT' -or
        [string]$release.created_at_utc -notmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,7})?Z$'
    ) {
        throw 'E_B0_RELEASE_AUTHORITY'
    }
    if ([string]$release.release_id -cne ('g0-bootstrap-founder-release-' + [string]$release.run)) { throw 'E_B0_RELEASE_AUTHORITY' }
    try {
        $expiry = [DateTime]::Parse([string]$release.expires_at_utc, [System.Globalization.CultureInfo]::InvariantCulture, [System.Globalization.DateTimeStyles]::AdjustToUniversal)
    }
    catch {
        throw 'E_B0_RELEASE_AUTHORITY'
    }
    if ($expiry.ToUniversalTime() -le [DateTime]::UtcNow) { throw 'E_B0_RELEASE_AUTHORITY' }

    $head = [string](Invoke-HubQaBootstrapGit -Root $Root -Arguments @('rev-parse', 'HEAD') -Code 'E_B0_RELEASE_AUTHORITY')[0]
    if ($head -notmatch '^[a-f0-9]{40}$') { throw 'E_B0_RELEASE_AUTHORITY' }
    $changed = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('diff-tree', '--no-commit-id', '--name-only', '-r', 'HEAD') -Code 'E_B0_RELEASE_AUTHORITY')
    if ($changed.Count -ne 1 -or [string]$changed[0] -cne $script:HubQaBootstrapReleasePath) { throw 'E_B0_RELEASE_AUTHORITY' }
    $headBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ('HEAD:' + $script:HubQaBootstrapReleasePath) -Code 'E_B0_RELEASE_AUTHORITY'
    if (-not (Test-HubQaBootstrapBytesEqual -Left $headBytes -Right $releaseRaw.Bytes)) { throw 'E_B0_RELEASE_AUTHORITY' }

    foreach ($commitName in @('g0_plan_commit', 'a0_commit', 'a02_commit', 'a1_commit', 'b_commit')) {
        $commit = [string]$release.$commitName
        if ($commit -notmatch '^[a-f0-9]{40}$') { throw 'E_B0_RELEASE_AUTHORITY' }
        [void](Invoke-HubQaBootstrapGit -Root $Root -Arguments @('merge-base', '--is-ancestor', $commit, $head) -Code 'E_B0_RELEASE_AUTHORITY')
    }
    $planPath = [string]$release.g0_plan_path
    if ($planPath -cne '.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md') { throw 'E_B0_RELEASE_AUTHORITY' }
    Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit ([string]$release.g0_plan_commit) -Expected @($planPath) -Code 'E_B0_RELEASE_AUTHORITY'
    $committedPlanBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName (([string]$release.g0_plan_commit) + ':' + $planPath) -Code 'E_B0_RELEASE_AUTHORITY'
    $headPlanBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ('HEAD:' + $planPath) -Code 'E_B0_RELEASE_AUTHORITY'
    if (-not (Test-HubQaBootstrapBytesEqual -Left $committedPlanBytes -Right $headPlanBytes)) { throw 'E_B0_RELEASE_AUTHORITY' }
    Assert-HubQaBootstrapA01FirstParent -Root $Root -A01Commit ([string]$release.a0_commit)
    Assert-HubQaBootstrapA02Authority -Root $Root -A01Commit ([string]$release.a0_commit) -A02Commit ([string]$release.a02_commit)
    Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit ([string]$release.a1_commit) -Expected $script:HubQaBootstrapA1Paths -Code 'E_B0_RELEASE_AUTHORITY'
    Assert-HubQaBootstrapExactCommitPaths -Root $Root -Commit ([string]$release.b_commit) -Expected $script:HubQaBootstrapBPaths -Code 'E_B0_RELEASE_AUTHORITY'

    return [pscustomobject][ordered]@{
        Value = $release
        Raw = $releaseRaw
        Commit = $head
        G0PlanRawBytes = $headPlanBytes
        G0PlanSha256 = Get-HubQaBootstrapSha256 -Bytes $headPlanBytes
    }
}

function New-HubQaBootstrapNonce {
    $bytes = New-Object byte[] 16
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $rng.GetBytes($bytes)
        return ([System.BitConverter]::ToString($bytes).Replace('-', '').ToLowerInvariant())
    }
    finally {
        $rng.Dispose()
    }
}

function Get-HubQaBootstrapNonceConsumptionRoot {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string]$Root)

    # The ledger is deliberately outside any per-run report root.  Build its
    # fixed path one component at a time and prove every resolved directory is
    # physical before it can contain an authority receipt.
    $current = Get-HubQaBootstrapPhysicalDirectoryRoot -Path $Root -Code 'E_B0_RELEASE_AUTHORITY'
    foreach ($segment in @('testframework', 'reports', 'g0-assets-runner-nonce-consumptions')) {
        $candidate = Join-Path $current $segment
        if ([System.IO.File]::Exists($candidate)) { throw 'E_B0_RELEASE_AUTHORITY' }
        if (-not [System.IO.Directory]::Exists($candidate)) {
            [void][System.IO.Directory]::CreateDirectory($candidate)
        }
        $physical = Get-HubQaBootstrapPhysicalDirectoryRoot -Path $candidate -Code 'E_B0_RELEASE_AUTHORITY'
        $expected = [System.IO.Path]::GetFullPath((Join-Path $current $segment)).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
        if (-not [string]::Equals($physical, $expected, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'E_B0_RELEASE_AUTHORITY' }
        $current = $physical
    }
    return $current
}

function ConvertFrom-HubQaBootstrapNonceConsumption {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][byte[]]$RawBytes,
        [Parameter(Mandatory = $true)][string]$ExpectedReleaseId,
        [Parameter(Mandatory = $true)][string]$ExpectedPath
    )

    try {
        $node = ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes $RawBytes
        $value = Get-HubQaEnvelopeNodeObject -Node $node -Code 'E_B0_RELEASE_AUTHORITY'
        Assert-HubQaEnvelopeKeys -Object $value -Expected @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status')
        $schema = Get-HubQaEnvelopeNodeString -Node $value['schema'] -Code 'E_B0_RELEASE_AUTHORITY'
        $task = Get-HubQaEnvelopeNodeString -Node $value['task'] -Code 'E_B0_RELEASE_AUTHORITY'
        $run = Get-HubQaEnvelopeNodeString -Node $value['run'] -Code 'E_B0_RELEASE_AUTHORITY'
        $releaseId = Get-HubQaEnvelopeNodeString -Node $value['release_id'] -Code 'E_B0_RELEASE_AUTHORITY'
        $releaseCommit = Get-HubQaEnvelopeNodeString -Node $value['release_commit'] -Code 'E_B0_RELEASE_AUTHORITY'
        $releaseRawSha256 = Get-HubQaEnvelopeNodeString -Node $value['release_raw_sha256'] -Code 'E_B0_RELEASE_AUTHORITY'
        $oneShotNonce = Get-HubQaEnvelopeNodeString -Node $value['one_shot_nonce'] -Code 'E_B0_RELEASE_AUTHORITY'
        $consumedAtUtc = Assert-HubQaEnvelopeUtc -Node $value['consumed_at_utc'] -Code 'E_B0_RELEASE_AUTHORITY'
        $status = Get-HubQaEnvelopeNodeString -Node $value['status'] -Code 'E_B0_RELEASE_AUTHORITY'
        if (
            $schema -cne 'hub.qa.g0-nonce-consumption/v1' -or
            $task -cne $script:HubQaBootstrapTask -or
            $run -cnotmatch $script:HubQaBootstrapRunPattern -or
            $releaseId -cnotmatch '^[a-z][a-z0-9-]{2,63}$' -or
            $releaseId -cne ('g0-bootstrap-founder-release-' + $run) -or
            $releaseId -cne $ExpectedReleaseId -or
            [System.IO.Path]::GetFileName($ExpectedPath) -cne ($releaseId + '.nonce-consumption.v1.json') -or
            $releaseCommit -cnotmatch '^[a-f0-9]{40}$' -or
            -not (Test-HubQaEnvelopeDigest -Value $releaseRawSha256) -or
            $oneShotNonce -cnotmatch '^[0-9a-f]{32}$' -or
            $status -cne 'CONSUMED' -or
            $null -eq $consumedAtUtc
        ) { throw 'E_B0_RELEASE_AUTHORITY' }
        return [pscustomobject][ordered]@{ Task = $task; Run = $run; ReleaseId = $releaseId; ReleaseCommit = $releaseCommit; ReleaseRawSha256 = $releaseRawSha256; OneShotNonce = $oneShotNonce; Status = $status }
    }
    catch {
        throw 'E_B0_RELEASE_AUTHORITY'
    }
}

function New-HubQaBootstrapNonceConsumptionReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release
    )

    $releaseId = [string]$Release.Value.release_id
    $run = [string]$Release.Value.run
    $releaseNonce = [string]$Release.Value.one_shot_nonce
    if ($releaseId -cnotmatch '^[a-z][a-z0-9-]{2,63}$' -or $releaseId -cne ('g0-bootstrap-founder-release-' + $run) -or $run -cnotmatch $script:HubQaBootstrapRunPattern -or $releaseNonce -cnotmatch '^[0-9a-f]{32}$' -or [string]$Release.Commit -cnotmatch '^[a-f0-9]{40}$' -or -not (Test-HubQaEnvelopeDigest -Value ([string]$Release.Raw.Sha256))) { throw 'E_B0_RELEASE_AUTHORITY' }
    $ledgerRoot = Get-HubQaBootstrapNonceConsumptionRoot -Root $Root
    $ledgerDirectory = [System.IO.DirectoryInfo]::new($ledgerRoot)
    $ledgerDirectory.Refresh()
    if (-not $ledgerDirectory.Exists -or (($ledgerDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_B0_RELEASE_AUTHORITY' }
    if (@($ledgerDirectory.GetDirectories()).Count -ne 0) { throw 'E_B0_RELEASE_AUTHORITY' }
    $seenReleaseIds = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    foreach ($entry in @($ledgerDirectory.GetFiles())) {
        $entry.Refresh()
        if (-not $entry.Exists -or (($entry.Attributes -band [System.IO.FileAttributes]::Directory) -ne 0) -or (($entry.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) -or $entry.Name -cnotmatch '^([a-z][a-z0-9-]{2,63})\.nonce-consumption\.v1\.json$') { throw 'E_B0_RELEASE_AUTHORITY' }
        $entryReleaseId = [string]$Matches[1]
        if (-not $seenReleaseIds.Add($entryReleaseId)) { throw 'E_B0_RELEASE_AUTHORITY' }
        $raw = Read-HubQaBootstrapHeldBoundedIngressControl -Path $entry.FullName -Root $ledgerRoot -Code 'E_B0_RELEASE_AUTHORITY'
        $record = ConvertFrom-HubQaBootstrapNonceConsumption -RawBytes ([byte[]]$raw.Bytes) -ExpectedReleaseId $entryReleaseId -ExpectedPath $entry.FullName
        if ([string]$record.OneShotNonce -ceq $releaseNonce) { throw 'E_B0_RELEASE_AUTHORITY' }
    }
    if ($seenReleaseIds.Contains($releaseId)) { throw 'E_B0_RELEASE_AUTHORITY' }

    $relativePath = 'testframework/reports/g0-assets-runner-nonce-consumptions/' + $releaseId + '.nonce-consumption.v1.json'
    $record = [ordered]@{
        schema = 'hub.qa.g0-nonce-consumption/v1'
        task = $script:HubQaBootstrapTask
        run = $run
        release_id = $releaseId
        release_commit = [string]$Release.Commit
        release_raw_sha256 = [string]$Release.Raw.Sha256
        one_shot_nonce = $releaseNonce
        consumed_at_utc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
        status = 'CONSUMED'
    }
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $written = Write-HubQaEnvelopeSafeCreateNew -Root $Root -RelativePath $relativePath -Bytes $bytes -Code 'E_B0_RELEASE_AUTHORITY'
    if (-not (Test-HubQaBootstrapBytesEqual -Left $bytes -Right ([byte[]]$written.Bytes))) { throw 'E_B0_RELEASE_AUTHORITY' }
    [void](ConvertFrom-HubQaBootstrapNonceConsumption -RawBytes ([byte[]]$written.Bytes) -ExpectedReleaseId $releaseId -ExpectedPath ([string]$written.FullPath))
    return [pscustomobject][ordered]@{ Path = $relativePath; RawBytes = [byte[]]$written.Bytes; Sha256 = Get-HubQaBootstrapSha256 -Bytes ([byte[]]$written.Bytes); Value = $record }
}

function Write-HubQaBootstrapCreateNew {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][byte[]]$Bytes,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $stream = $null
    try {
        $stream = [System.IO.File]::Open($Path, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
        $stream.Write($Bytes, 0, $Bytes.Length)
        $stream.Flush($true)
    }
    catch {
        throw $Code
    }
    finally {
        if ($null -ne $stream) { $stream.Dispose() }
    }
}

function New-HubQaBootstrapReportRoot {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Run,
        [Parameter(Mandatory = $true)][string]$Nonce
    )

    if ($Run -notmatch $script:HubQaBootstrapRunPattern -or $Nonce -notmatch '^[0-9a-f]{32}$') { throw 'E_B0_TASK_RUN' }
    $reportsRoot = Join-Path $Root 'testframework\reports'
    [System.IO.Directory]::CreateDirectory($reportsRoot) | Out-Null
    $reportRoot = Join-Path $reportsRoot ($script:HubQaBootstrapTask + '-' + $Run)
    if ([System.IO.Directory]::Exists($reportRoot) -or [System.IO.File]::Exists($reportRoot)) { throw 'E_B0_ONE_SHOT_REJECTED' }
    [System.IO.Directory]::CreateDirectory((Join-Path $reportRoot 'bootstrap')) | Out-Null
    [System.IO.Directory]::CreateDirectory((Join-Path $reportRoot 'ingress')) | Out-Null
    [System.IO.Directory]::CreateDirectory((Join-Path $reportRoot 'commands')) | Out-Null
    [System.IO.Directory]::CreateDirectory((Join-Path $reportRoot 'selftests')) | Out-Null
    $markerBytes = [System.Text.UTF8Encoding]::new($false).GetBytes(('{"schema":"hub.qa.bootstrap-owner/v1","task":"{0}","run":"{1}","owner_nonce":"{2}"}' -f $script:HubQaBootstrapTask, $Run, $Nonce))
    $markerPath = Join-Path $reportRoot 'bootstrap\owner-marker.v1.json'
    Write-HubQaBootstrapCreateNew -Path $markerPath -Bytes $markerBytes -Code 'E_B0_ONE_SHOT_REJECTED'
    return [pscustomobject][ordered]@{
        ReportRoot = $reportRoot
        MarkerPath = $markerPath
        MarkerRawBytes = $markerBytes
        MarkerSha256 = Get-HubQaBootstrapSha256 -Bytes $markerBytes
    }
}

function Assert-HubQaBootstrapPrivateProcessPrimitiveContract {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$RelativePath,
        [Parameter(Mandatory = $true)][byte[]]$Bytes
    )

    try { $text = [System.Text.UTF8Encoding]::new($false, $true).GetString($Bytes) }
    catch { throw 'E_B0_PRIVATE_SOURCE' }
    $requiredFragments = if ($RelativePath -ceq 'testframework/docker/source-image-attestor.ps1') {
        @(
            'function Get-HubQaGitBlobBytes {',
            ('$process.StandardError.ReadTo' + 'End()'),
            ('$process.WaitFor' + 'Exit()'),
            'function Get-HubQaGitTreeEntry {',
            ('$lines = @(' + '& ' + 'git -C $Repository ls-tree'),
            'function Get-HubQaGitObjectSourceTree {',
            ('$commitType = @(' + '& ' + 'git -C $root cat-file -t')
        )
    }
    elseif ($RelativePath -ceq 'testframework/docker/admit-local-dependency-assets.ps1') {
        @(
            'function Get-HubQaDockerClientPath {',
            'function Invoke-HubQaG0AdmissionFromBootstrapContext {',
            ('& ' + '$BootstrapPrivateContext.PrivateDockerClientPath @dockerArguments')
        )
    }
    else { return }
    foreach ($fragment in $requiredFragments) {
        $first = $text.IndexOf($fragment, [System.StringComparison]::Ordinal)
        if ($first -lt 0 -or $text.IndexOf($fragment, $first + $fragment.Length, [System.StringComparison]::Ordinal) -ge 0) {
            throw 'E_B0_PRIVATE_SOURCE'
        }
    }
}

function Import-HubQaBootstrapPrivateSources {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release
    )

    # The release names exact accepted A0.2/A0.1 and B blobs.  Verify every
    # private source file against those objects before any file is dot-sourced;
    # unrelated staged legacy deletions are deliberately outside these paths.
    Assert-HubQaBootstrapCurrentAcceptedSources -Root $Root -Release $Release
    $sourceFiles = @(
        [pscustomobject]@{ RelativePath = 'testframework/docker/source-image-attestor.ps1'; Describe = $true },
        [pscustomobject]@{ RelativePath = 'testframework/docker/admit-local-dependency-assets.ps1'; Describe = $false },
        [pscustomobject]@{ RelativePath = 'testframework/runner/run-task-v2.ps1'; Describe = $false },
        [pscustomobject]@{ RelativePath = 'testframework/runner/new-run-envelope.ps1'; Describe = $false },
        [pscustomobject]@{ RelativePath = 'testframework/runner/new-repository-snapshot.ps1'; Describe = $false }
    )
    $held = New-Object 'System.Collections.Generic.List[object]'
    try {
        foreach ($source in $sourceFiles) {
            $relativePath = [string]$source.RelativePath
            $commit = Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path $relativePath -Code 'E_B0_PRIVATE_SOURCE'
            [byte[]]$acceptedBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ($commit + ':' + $relativePath) -Code 'E_B0_PRIVATE_SOURCE'
            if ($acceptedBytes.Length -lt 1 -or $acceptedBytes.Length -gt $script:HubQaBootstrapMaxControlBytes) { throw 'E_B0_PRIVATE_SOURCE' }
            $path = [System.IO.Path]::GetFullPath((Join-Path $Root $relativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar)))
            $opened = Open-HubQaBootstrapSafeRegularFile -Path $path -Root $Root -MaximumBytes $script:HubQaBootstrapMaxControlBytes -Code 'E_B0_PRIVATE_SOURCE'
            try {
                $raw = Read-HubQaBootstrapHeldFileBytes -Opened $opened -MaximumBytes $script:HubQaBootstrapMaxControlBytes -Code 'E_B0_PRIVATE_SOURCE'
                if (-not (Test-HubQaBootstrapBytesEqual -Left ([byte[]]$raw.Bytes) -Right $acceptedBytes)) { throw 'E_B0_PRIVATE_SOURCE' }
                Assert-HubQaBootstrapPrivateProcessPrimitiveContract -RelativePath $relativePath -Bytes ([byte[]]$raw.Bytes)
                [void]$held.Add([pscustomobject][ordered]@{
                    Path = [string]$opened.FinalPath
                    RelativePath = $relativePath
                    Describe = [bool]$source.Describe
                    Handle = $opened.Handle
                    AncestorHandles = @($opened.AncestorHandles)
                    Observation = $opened.Observation
                    ShareMode = [string]$opened.ShareMode
                    AcceptedSha256 = Get-HubQaBootstrapSha256 -Bytes $acceptedBytes
                })
                $opened = $null
            }
            finally {
                if ($null -ne $opened) { Close-HubQaBootstrapSafeRegularFile -Opened $opened }
            }
        }
        return @($held.ToArray())
    }
    catch {
        foreach ($source in @($held.ToArray())) { Close-HubQaBootstrapSafeRegularFile -Opened $source }
        throw
    }
}

function ConvertTo-HubQaBootstrapWindowsCommandLineArgument {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string]$Value)

    if ($Value.Length -gt 0 -and $Value -notmatch '[\s"]') { return $Value }
    $builder = New-Object System.Text.StringBuilder
    [void]$builder.Append('"')
    $backslashes = 0
    for ($index = 0; $index -lt $Value.Length; $index++) {
        $character = $Value[$index]
        if ($character -eq '\') { $backslashes++; continue }
        if ($character -eq '"') {
            for ($slashIndex = 0; $slashIndex -lt (($backslashes * 2) + 1); $slashIndex++) { [void]$builder.Append('\') }
            [void]$builder.Append('"')
            $backslashes = 0
            continue
        }
        for ($slashIndex = 0; $slashIndex -lt $backslashes; $slashIndex++) { [void]$builder.Append('\') }
        $backslashes = 0
        [void]$builder.Append($character)
    }
    for ($slashIndex = 0; $slashIndex -lt ($backslashes * 2); $slashIndex++) { [void]$builder.Append('\') }
    [void]$builder.Append('"')
    return $builder.ToString()
}

function Assert-HubQaBootstrapOuterDeadline {
    if ($null -eq $script:HubQaBootstrapRuntimeStopwatch -or -not $script:HubQaBootstrapRuntimeStopwatch.IsRunning) {
        throw 'E_B0_TIMEOUT_124'
    }
    Assert-HubQaBootstrapDeadline -Stopwatch $script:HubQaBootstrapRuntimeStopwatch -Seconds $script:HubQaBootstrapOuterSeconds
}

function Assert-HubQaBootstrapCleanupAndOuterDeadline {
    Assert-HubQaBootstrapOuterDeadline
    if ($null -eq $script:HubQaBootstrapCleanupBudgetStopwatch -or -not $script:HubQaBootstrapCleanupBudgetStopwatch.IsRunning) {
        throw 'E_B0_TIMEOUT_124'
    }
    Assert-HubQaBootstrapDeadline -Stopwatch $script:HubQaBootstrapCleanupBudgetStopwatch -Seconds $script:HubQaBootstrapCleanupSeconds
}

function Get-HubQaBootstrapEffectiveTimeoutMilliseconds {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][int]$NominalSeconds,
        [switch]$CountsAgainstTarget,
        [switch]$CleanupOperation
    )

    if ($NominalSeconds -lt 1) { throw 'E_B0_TIMEOUT_124' }
    $milliseconds = [int64]$NominalSeconds * 1000
    if ($null -ne $script:HubQaBootstrapRuntimeStopwatch) {
        $reserveSeconds = if ($CleanupOperation) { 0 } else { $script:HubQaBootstrapCleanupSeconds + $script:HubQaBootstrapOrchestrationSlackSeconds }
        $remainingOuter = [int64][Math]::Floor((($script:HubQaBootstrapOuterSeconds - $reserveSeconds) - $script:HubQaBootstrapRuntimeStopwatch.Elapsed.TotalSeconds) * 1000)
        if ($remainingOuter -lt 1) { throw 'E_B0_TIMEOUT_124' }
        if ($remainingOuter -lt $milliseconds) { $milliseconds = $remainingOuter }
    }
    if ($CountsAgainstTarget -and $null -ne $script:HubQaBootstrapTargetBudgetStopwatch) {
        $remainingTarget = [int64][Math]::Floor(($script:HubQaBootstrapTargetSeconds - $script:HubQaBootstrapTargetBudgetStopwatch.Elapsed.TotalSeconds) * 1000)
        if ($remainingTarget -lt 1) { throw 'E_B0_TIMEOUT_124' }
        if ($remainingTarget -lt $milliseconds) { $milliseconds = $remainingTarget }
    }
    if ($CleanupOperation -and $null -ne $script:HubQaBootstrapCleanupBudgetStopwatch) {
        $remainingCleanup = [int64][Math]::Floor(($script:HubQaBootstrapCleanupSeconds - $script:HubQaBootstrapCleanupBudgetStopwatch.Elapsed.TotalSeconds) * 1000)
        if ($remainingCleanup -lt 1) { throw 'E_B0_TIMEOUT_124' }
        if ($remainingCleanup -lt $milliseconds) { $milliseconds = $remainingCleanup }
    }
    if ($milliseconds -lt 1 -or $milliseconds -gt [int]::MaxValue) { throw 'E_B0_TIMEOUT_124' }
    return [int]$milliseconds
}

function Get-HubQaBootstrapOperationRemainingMilliseconds {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Diagnostics.Stopwatch]$Stopwatch,
        [Parameter(Mandatory = $true)][int]$BudgetMilliseconds,
        [switch]$ReserveForTermination
    )

    [int64]$remaining = [int64]$BudgetMilliseconds - [int64]$Stopwatch.ElapsedMilliseconds
    if ($remaining -lt 1) { throw 'E_B0_TIMEOUT_124' }
    if ($ReserveForTermination) {
        [int64]$reserve = [Math]::Min(5000, [Math]::Max(1, [Math]::Floor($remaining / 10)))
        $remaining -= $reserve
        if ($remaining -lt 1) { throw 'E_B0_TIMEOUT_124' }
    }
    if ($remaining -gt [int]::MaxValue) { $remaining = [int]::MaxValue }
    return [int]$remaining
}

function Start-HubQaBootstrapAssignedProcess {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$FileName,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$WorkingDirectory,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if (-not [System.IO.File]::Exists($FileName) -or -not [System.IO.Directory]::Exists($WorkingDirectory) -or $FileName.IndexOf([char]0) -ge 0) { throw $Code }
    $commandLine = (ConvertTo-HubQaBootstrapWindowsCommandLineArgument -Value $FileName) + ' ' + (($Arguments | ForEach-Object { ConvertTo-HubQaBootstrapWindowsCommandLineArgument -Value ([string]$_) }) -join ' ')
    try {
        return [HubQaG0AssignedProcess]::Start($FileName, $commandLine, $WorkingDirectory)
    }
    catch { throw $Code }
}

function Stop-HubQaBootstrapProcessJobWithinDeadline {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][HubQaG0AssignedProcess]$NativeProcess,
        [Parameter(Mandatory = $true)][System.Diagnostics.Stopwatch]$Stopwatch,
        [Parameter(Mandatory = $true)][int]$BudgetMilliseconds
    )

    try {
        if (-not $NativeProcess.Process.HasExited) { $NativeProcess.Terminate([uint32]124) }
        $NativeProcess.CloseJob()
        $remaining = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $Stopwatch -BudgetMilliseconds $BudgetMilliseconds
        if (-not $NativeProcess.Process.WaitForExit($remaining)) { throw 'E_B0_TIMEOUT_124' }
    }
    catch {
        throw 'E_B0_TIMEOUT_124'
    }
}

function Add-HubQaBootstrapCommandReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][int]$ExitCode,
        [Parameter(Mandatory = $true)][bool]$TimedOut,
        [Parameter(Mandatory = $true)][string]$StartedAtUtc,
        [Parameter(Mandatory = $true)][string]$FinishedAtUtc,
        [Parameter(Mandatory = $true)][string]$StandardOutput,
        [Parameter(Mandatory = $true)][string]$StandardError
    )

    if ([string]::IsNullOrWhiteSpace($Name) -or $Arguments.Count -eq 0) { throw 'E_B0_ROLE_TRANSPORT' }
    [void]$script:HubQaBootstrapCommandReceipts.Add([ordered]@{
        name = $Name
        argv = @($Arguments)
        exit_code = $ExitCode
        timed_out = $TimedOut
        started_at_utc = $StartedAtUtc
        finished_at_utc = $FinishedAtUtc
        stdout_sha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes($StandardOutput))
        stderr_sha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes($StandardError))
    })
}

function Add-HubQaBootstrapCommandReceiptBytes {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][int]$ExitCode,
        [Parameter(Mandatory = $true)][bool]$TimedOut,
        [Parameter(Mandatory = $true)][string]$StartedAtUtc,
        [Parameter(Mandatory = $true)][string]$FinishedAtUtc,
        [Parameter(Mandatory = $true)][byte[]]$StandardOutputBytes,
        [Parameter(Mandatory = $true)][byte[]]$StandardErrorBytes
    )

    if ([string]::IsNullOrWhiteSpace($Name) -or $Arguments.Count -eq 0) { throw 'E_B0_ROLE_TRANSPORT' }
    [void]$script:HubQaBootstrapCommandReceipts.Add([ordered]@{
        name = $Name
        argv = @($Arguments)
        exit_code = $ExitCode
        timed_out = $TimedOut
        started_at_utc = $StartedAtUtc
        finished_at_utc = $FinishedAtUtc
        stdout_sha256 = Get-HubQaBootstrapSha256 -Bytes $StandardOutputBytes
        stderr_sha256 = Get-HubQaBootstrapSha256 -Bytes $StandardErrorBytes
    })
}

function Invoke-HubQaBootstrapProcess {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$FileName,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][int]$TimeoutSeconds,
        [Parameter(Mandatory = $true)][string]$Code,
        [string]$Name = 'process',
        [switch]$CountsAgainstTarget,
        [switch]$CleanupOperation
    )

    $invalidArguments = @($Arguments | Where-Object { $null -eq $_ -or ([string]$_).IndexOf([char]0) -ge 0 })
    if (-not [System.IO.File]::Exists($FileName) -or $TimeoutSeconds -lt 1 -or $Arguments.Count -eq 0 -or $invalidArguments.Count -ne 0) { throw $Code }
    $workingDirectory = Get-HubQaBootstrapRepositoryRoot
    $nativeProcess = $null
    $stdoutStream = $null
    $stderrStream = $null
    $stdoutMemory = New-Object System.IO.MemoryStream
    $stderrMemory = New-Object System.IO.MemoryStream
    $timedOut = $false
    $startedAtUtc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    $operationBudgetMilliseconds = Get-HubQaBootstrapEffectiveTimeoutMilliseconds -NominalSeconds $TimeoutSeconds -CountsAgainstTarget:$CountsAgainstTarget -CleanupOperation:$CleanupOperation
    $operationStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $nativeProcess = Start-HubQaBootstrapAssignedProcess -FileName $FileName -Arguments $Arguments -WorkingDirectory $workingDirectory -Code $Code
        $process = $nativeProcess.Process
        $stdoutStream = [System.IO.FileStream]::new($nativeProcess.StandardOutputRead, [System.IO.FileAccess]::Read, 65536, $false)
        $stderrStream = [System.IO.FileStream]::new($nativeProcess.StandardErrorRead, [System.IO.FileAccess]::Read, 65536, $false)
        $stdoutTask = $stdoutStream.CopyToAsync($stdoutMemory)
        $stderrTask = $stderrStream.CopyToAsync($stderrMemory)
        $waitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds -ReserveForTermination
        if (-not $process.WaitForExit($waitMilliseconds)) {
            $timedOut = $true
            Stop-HubQaBootstrapProcessJobWithinDeadline -NativeProcess $nativeProcess -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        }
        else { $nativeProcess.CloseJob() }
        $stdoutWaitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        if (-not $stdoutTask.Wait($stdoutWaitMilliseconds)) { throw 'E_B0_TIMEOUT_124' }
        $stderrWaitMilliseconds = Get-HubQaBootstrapOperationRemainingMilliseconds -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
        if (-not $stderrTask.Wait($stderrWaitMilliseconds)) { throw 'E_B0_TIMEOUT_124' }
        [byte[]]$stdoutBytes = $stdoutMemory.ToArray()
        [byte[]]$stderrBytes = $stderrMemory.ToArray()
        Add-HubQaBootstrapCommandReceiptBytes -Name $Name -Arguments $Arguments -ExitCode ([int]$process.ExitCode) -TimedOut $timedOut -StartedAtUtc $startedAtUtc -FinishedAtUtc ([DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)) -StandardOutputBytes $stdoutBytes -StandardErrorBytes $stderrBytes
        if ($timedOut) { throw 'E_B0_TIMEOUT_124' }
        try {
            $stdout = [System.Text.UTF8Encoding]::new($false, $true).GetString($stdoutBytes)
            $stderr = [System.Text.UTF8Encoding]::new($false, $true).GetString($stderrBytes)
        }
        catch { throw $Code }
        $result = [pscustomobject][ordered]@{
            ExitCode = [int]$process.ExitCode
            TimedOut = $timedOut
            StandardOutput = [string]$stdout
            StandardError = [string]$stderr
        }
        return $result
    }
    finally {
        try {
            if ($null -ne $nativeProcess -and -not $nativeProcess.JobClosed) {
                if (-not $nativeProcess.Process.HasExited) {
                    Stop-HubQaBootstrapProcessJobWithinDeadline -NativeProcess $nativeProcess -Stopwatch $operationStopwatch -BudgetMilliseconds $operationBudgetMilliseconds
                }
                else { $nativeProcess.CloseJob() }
            }
        }
        finally {
            $operationStopwatch.Stop()
            if ($null -ne $stdoutStream) { $stdoutStream.Dispose() }
            if ($null -ne $stderrStream) { $stderrStream.Dispose() }
            $stdoutMemory.Dispose()
            $stderrMemory.Dispose()
            if ($null -ne $nativeProcess) { $nativeProcess.Dispose() }
        }
    }
}

function Invoke-HubQaBootstrapDockerProbe {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$DockerClientPath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$Code,
        [switch]$CleanupOperation
    )

    $result = Invoke-HubQaBootstrapProcess -Name 'docker-inspect' -FileName $DockerClientPath -Arguments $Arguments -TimeoutSeconds $script:HubQaBootstrapInspectSeconds -Code $Code -CountsAgainstTarget:(-not $CleanupOperation) -CleanupOperation:$CleanupOperation
    if ([bool]$result.TimedOut) { throw 'E_B0_TIMEOUT_124' }
    return $result
}

function Invoke-HubQaBootstrapDocker {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$DockerClientPath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$Code,
        [int]$TimeoutSeconds = $script:HubQaBootstrapCommandSeconds,
        [switch]$CleanupOperation
    )

    $result = Invoke-HubQaBootstrapProcess -Name 'docker' -FileName $DockerClientPath -Arguments $Arguments -TimeoutSeconds $TimeoutSeconds -Code $Code -CountsAgainstTarget:(-not $CleanupOperation) -CleanupOperation:$CleanupOperation
    if ([bool]$result.TimedOut) { throw 'E_B0_TIMEOUT_124' }
    if ([int]$result.ExitCode -ne 0) { throw $Code }
    if ([string]::IsNullOrEmpty([string]$result.StandardOutput)) { return @() }
    return @(([string]$result.StandardOutput -split "`r?`n" | Where-Object { $_.Length -gt 0 }))
}

function Assert-HubQaBootstrapDockerImageAbsent {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$DockerClientPath,
        [Parameter(Mandatory = $true)][string]$Tag,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $DockerClientPath -Arguments @('image', 'inspect', '--format', '{{.Id}}', $Tag) -Code $Code
    if ([int]$probe.ExitCode -ne 1 -or -not [string]::IsNullOrEmpty([string]$probe.StandardOutput)) { throw $Code }
}

function Add-HubQaBootstrapOwnedImage {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Tag,
        [Parameter(Mandatory = $true)][string]$ImageId,
        [Parameter(Mandatory = $true)][string]$OwnerMarkerSha256,
        [Parameter(Mandatory = $true)][bool]$RequiresOwnerLabel
    )

    if ($Tag.Length -eq 0 -or $ImageId -notmatch '^sha256:[a-f0-9]{64}$' -or $OwnerMarkerSha256 -notmatch '^sha256:[a-f0-9]{64}$') {
        throw 'E_B0_IMAGE_BUILD'
    }
    [void]$script:HubQaBootstrapOwnedResources.Add([pscustomobject][ordered]@{
        Tag = $Tag
        ImageId = $ImageId
        OwnerMarkerSha256 = $OwnerMarkerSha256
        RequiresOwnerLabel = $RequiresOwnerLabel
    })
}

function Add-HubQaBootstrapOwnedContainer {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$DockerClientPath,
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$OwnerMarkerSha256
    )

    if ($Name.Length -eq 0 -or $OwnerMarkerSha256 -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_CLEANUP' }
    $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $DockerClientPath -Arguments @('container', 'inspect', '--format', '{{.Id}}', $Name) -Code 'E_B0_CLEANUP'
    if ([int]$probe.ExitCode -ne 1 -or -not [string]::IsNullOrEmpty([string]$probe.StandardOutput)) { throw 'E_B0_CLEANUP' }
    [void]$script:HubQaBootstrapOwnedContainers.Add([pscustomobject][ordered]@{
        Name = $Name
        ContainerId = $null
        OwnerMarkerSha256 = $OwnerMarkerSha256
    })
}

function Update-HubQaBootstrapOwnedContainerIdentity {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$DockerClientPath,
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $entry = @($script:HubQaBootstrapOwnedContainers | Where-Object { [string]$_.Name -ceq $Name })
    if ($entry.Count -ne 1) { throw $Code }
    $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $DockerClientPath -Arguments @('container', 'inspect', '--format', '{{.Id}}|{{index .Config.Labels "hub.qa.owner-marker-sha256"}}', $Name) -Code $Code
    if ([int]$probe.ExitCode -ne 0) { return }
    $line = ([string]$probe.StandardOutput).Trim()
    if ($line -notmatch '^(sha256:)?([a-f0-9]{64})\|sha256:[a-f0-9]{64}$') { throw $Code }
    if ([string]$line.Split('|')[1] -cne [string]$entry[0].OwnerMarkerSha256) { throw $Code }
    $entry[0].ContainerId = [string]$line.Split('|')[0]
}

function New-HubQaBootstrapAdmissionDependencyBinding {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release
    )

    # The helper needs only the A1-committed dependency-manifest bytes to
    # inspect bases and admit the copied ingress. Full current-HEAD A1 control
    # binding remains deliberately later: after admission helper cleanup and
    # before source verification/target construction.
    $a1Commit = [string]$Release.Value.a1_commit
    if ($a1Commit -notmatch '^[a-f0-9]{40}$') { throw 'BLOCKED_EXTERNAL_INGRESS' }
    [byte[]]$rawBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ($a1Commit + ':testframework/docker/dependency-bundle.manifest.v1.json') -Code 'BLOCKED_EXTERNAL_INGRESS'
    if ($rawBytes.Length -eq 0 -or ($rawBytes.Length -ge 3 -and $rawBytes[0] -eq 0xef -and $rawBytes[1] -eq 0xbb -and $rawBytes[2] -eq 0xbf)) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    try { $rawText = [System.Text.UTF8Encoding]::new($false, $true).GetString($rawBytes) } catch { throw 'BLOCKED_EXTERNAL_INGRESS' }
    $raw = [pscustomobject][ordered]@{
        Bytes = $rawBytes
        Text = $rawText
        Sha256 = Get-HubQaBootstrapSha256 -Bytes $rawBytes
    }
    $value = ConvertFrom-HubQaBootstrapClosedJson -Text $raw.Text -Code 'BLOCKED_EXTERNAL_INGRESS'
    Assert-HubQaBootstrapExactKeys -Value $value -Expected @('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256') -Code 'BLOCKED_EXTERNAL_INGRESS'
    if (
        [string]$value.schema -cne 'hub.qa.dependency-bundle/v1' -or
        [string]$value.bundle_id -notmatch '^[a-z0-9][a-z0-9-]{7,63}$' -or
        [string]$value.source_commit -cne $script:HubQaBootstrapSourceCommit -or
        [string]$value.bundle_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        $value.assets -isnot [System.Array] -or $value.bases -isnot [System.Array]
    ) {
        throw 'BLOCKED_EXTERNAL_INGRESS'
    }
    Assert-HubQaBootstrapExactKeys -Value $value.platform -Expected @('os', 'architecture', 'variant') -Code 'BLOCKED_EXTERNAL_INGRESS'
    if ([string]$value.platform.os -cne 'linux' -or [string]$value.platform.architecture -cne 'amd64' -or $null -ne $value.platform.variant) {
        throw 'BLOCKED_EXTERNAL_INGRESS'
    }
    $expectedAssets = @(
        [pscustomobject]@{ Id = 'go-module-cache'; Path = 'assets/go-module-cache.tar'; Targets = @('backend-unit', 'backend-openapi') },
        [pscustomobject]@{ Id = 'go-race-toolchain'; Path = 'assets/go-race-toolchain.tar'; Targets = @('backend-unit') },
        [pscustomobject]@{ Id = 'frontend-node-modules'; Path = 'assets/frontend-node-modules.tar'; Targets = @('frontend-unit') },
        [pscustomobject]@{ Id = 'swag-binary'; Path = 'assets/swag-binary.tar'; Targets = @('backend-openapi') }
    )
    if ($value.assets.Count -ne $expectedAssets.Count) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    for ($assetIndex = 0; $assetIndex -lt $expectedAssets.Count; $assetIndex++) {
        $asset = $value.assets[$assetIndex]
        Assert-HubQaBootstrapA1Asset -Asset $asset -Code 'BLOCKED_EXTERNAL_INGRESS'
        if (
            [string]$asset.asset_id -cne [string]$expectedAssets[$assetIndex].Id -or
            [string]$asset.relative_path -cne [string]$expectedAssets[$assetIndex].Path -or
            ((@($asset.targets) -cjoin ',') -cne (@($expectedAssets[$assetIndex].Targets) -cjoin ','))
        ) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    }
    $expectedBaseIds = @('go-backend-unit', 'go-backend-openapi', 'frontend-unit', 'bootstrap-helper-python')
    if ($value.bases.Count -ne $expectedBaseIds.Count) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    for ($baseIndex = 0; $baseIndex -lt $expectedBaseIds.Count; $baseIndex++) {
        $base = $value.bases[$baseIndex]
        Assert-HubQaBootstrapExactKeys -Value $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Code 'BLOCKED_EXTERNAL_INGRESS'
        $consumers = @($base.consumers)
        if (
            [string]$base.base_id -cne $expectedBaseIds[$baseIndex] -or
            $base.consumers -isnot [System.Array] -or $consumers.Count -lt 1 -or
            (@($consumers | Select-Object -Unique).Count -ne $consumers.Count) -or
            (@($consumers | Where-Object { [string]$_ -cnotin @('backend-unit', 'frontend-unit', 'backend-openapi', 'g0-admission-helper', 'g0-planning-selftest') }).Count -ne 0) -or
            [string]$base.local_image_id -notmatch '^sha256:[a-f0-9]{64}$' -or
            [string]$base.manifest_digest -notmatch '^sha256:[a-f0-9]{64}$' -or
            [string]::IsNullOrWhiteSpace([string]$base.license_spdx) -or
            [string]$base.license_evidence_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
            [string]$base.provenance_sha256 -notmatch '^sha256:[a-f0-9]{64}$'
        ) { throw 'BLOCKED_EXTERNAL_INGRESS' }
        if ([string]$base.base_id -ceq 'bootstrap-helper-python' -and (($consumers -cjoin ',') -cne 'g0-admission-helper,g0-planning-selftest')) {
            throw 'BLOCKED_EXTERNAL_INGRESS'
        }
    }
    return [pscustomobject][ordered]@{
        Controls = [pscustomobject][ordered]@{
            dependency = [pscustomobject][ordered]@{
                Raw = $raw
                Value = $value
            }
        }
    }
}

function Assert-HubQaBootstrapA1Asset {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Asset,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $expected = @(
        'asset_id', 'targets', 'relative_path', 'sha256', 'bytes',
        'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256',
        'license_spdx', 'license_evidence_sha256', 'provenance_sha256'
    )
    Assert-HubQaBootstrapExactKeys -Value $Asset -Expected $expected -Code $Code
    $targets = @($Asset.targets)
    if (
        [string]$Asset.asset_id -notmatch '^[a-z][a-z0-9-]{2,63}$' -or
        $Asset.targets -isnot [System.Array] -or
        $targets.Count -eq 0 -or
        (@($targets | Select-Object -Unique).Count -ne $targets.Count) -or
        (@($targets | Where-Object { [string]$_ -cnotin @('backend-unit', 'frontend-unit', 'backend-openapi') }).Count -ne 0) -or
        [string]$Asset.relative_path -notmatch '^assets/[a-z0-9-]+\.tar$' -or
        [string]$Asset.sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        $Asset.bytes -isnot [int64] -or $Asset.bytes -lt 1 -or $Asset.bytes -gt $script:HubQaBootstrapMaxIngressAssetBytes -or
        [string]$Asset.archive_format -cne 'posix-pax-tar' -or
        $Asset.member_count -isnot [int64] -or $Asset.member_count -lt 1 -or $Asset.member_count -gt $script:HubQaBootstrapMaxArchiveMembers -or
        $Asset.expanded_bytes -isnot [int64] -or $Asset.expanded_bytes -lt 1 -or $Asset.expanded_bytes -gt $script:HubQaBootstrapMaxExpandedAssetBytes -or
        [string]$Asset.layout_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]::IsNullOrWhiteSpace([string]$Asset.license_spdx) -or
        [string]$Asset.license_evidence_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$Asset.provenance_sha256 -notmatch '^sha256:[a-f0-9]{64}$'
    ) {
        throw $Code
    }
}

function Get-HubQaBootstrapClosedDockerClientPath {
    [CmdletBinding()]
    param()

    $programFiles = [System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles)
    if ([string]::IsNullOrWhiteSpace($programFiles)) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
    $programFiles = [System.IO.Path]::GetFullPath($programFiles).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $candidate = [System.IO.Path]::GetFullPath((Join-Path $programFiles 'Docker\Docker\resources\bin\docker.exe'))
    if (-not $candidate.StartsWith($programFiles + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
    $opened = Open-HubQaBootstrapSafeRegularFile -Path $candidate -Root $programFiles -MaximumBytes 268435456 -Code 'E_B0_LOCAL_BASE_RECEIPT'
    try {
        if ([string]$opened.ShareMode -cne 'read-only-no-write-no-delete' -or -not [string]::Equals([string]$opened.FinalPath, $candidate, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw 'E_B0_LOCAL_BASE_RECEIPT'
        }
        return $candidate
    }
    finally {
        Close-HubQaBootstrapSafeRegularFile -Opened $opened
    }
}

function New-HubQaBootstrapLocalBaseReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Run,
        [Parameter(Mandatory = $true)][string]$Nonce
    )

    $manifest = $Binding.Controls.dependency.Value
    Assert-HubQaBootstrapExactKeys -Value $manifest -Expected @('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256') -Code 'E_B0_LOCAL_BASE_RECEIPT'
    $expectedBases = @('go-backend-unit', 'go-backend-openapi', 'frontend-unit', 'bootstrap-helper-python')
    $bases = @($manifest.bases)
    if ($bases.Count -ne $expectedBases.Count) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
    $dockerClient = Get-HubQaDockerClientPath
    $receiptBases = New-Object 'System.Collections.Generic.List[object]'
    $helperBaseImageId = $null
    foreach ($index in 0..($expectedBases.Count - 1)) {
        $base = $bases[$index]
        Assert-HubQaBootstrapExactKeys -Value $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Code 'E_B0_LOCAL_BASE_RECEIPT'
        if ([string]$base.base_id -cne $expectedBases[$index] -or [string]$base.local_image_id -notmatch '^sha256:[a-f0-9]{64}$' -or [string]$base.manifest_digest -notmatch '^sha256:[a-f0-9]{64}$') {
            throw 'E_B0_LOCAL_BASE_RECEIPT'
        }
        $inspectOutput = Invoke-HubQaBootstrapDocker -DockerClientPath $dockerClient -Arguments @('image', 'inspect', '--format', '{{json .}}', ([string]$base.local_image_id)) -Code 'E_B0_LOCAL_BASE_RECEIPT' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
        if ($inspectOutput.Count -ne 1) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
        try { $inspect = ([string]$inspectOutput[0]) | ConvertFrom-Json -ErrorAction Stop } catch { throw 'E_B0_LOCAL_BASE_RECEIPT' }
        if ([string]$inspect.Id -cne [string]$base.local_image_id) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
        $manifestFound = $false
        foreach ($repoDigest in @($inspect.RepoDigests)) {
            if ([string]$repoDigest -match ('@' + [regex]::Escape([string]$base.manifest_digest) + '$')) { $manifestFound = $true; break }
        }
        if (-not $manifestFound) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
        [void]$receiptBases.Add([ordered]@{
            base_id = [string]$base.base_id
            local_image_id = [string]$base.local_image_id
            manifest_digest = [string]$base.manifest_digest
        })
        if ([string]$base.base_id -ceq 'bootstrap-helper-python') {
            $helperBaseImageId = [string]$base.local_image_id
        }
    }
    if ($null -eq $helperBaseImageId) { throw 'E_B0_LOCAL_BASE_RECEIPT' }
    $record = [ordered]@{
        schema = 'hub.qa.local-base-inspect/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        owner_marker_sha256 = [string]$Report.MarkerSha256
        receipt_nonce = $Nonce
        bases = @($receiptBases.ToArray())
    }
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $path = Join-Path $Report.ReportRoot 'bootstrap\local-base-inspect.v1.json'
    Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_LOCAL_BASE_RECEIPT'
    return [pscustomobject][ordered]@{
        Path = $path
        RawBytes = $bytes
        Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
        Nonce = $Nonce
        HelperBaseImageId = $helperBaseImageId
        DockerClientPath = $dockerClient
    }
}

function New-HubQaBootstrapHelperImage {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][string]$Run
    )

    $helperBaseTag = 'hub-qa-bootstrap-helper-python:local'
    $helperImageTag = 'hub-qa-g0-admission-helper:local'
    $buildContextsRoot = Join-Path $Report.ReportRoot 'bootstrap\build-context'
    $stagingRoot = Join-Path $buildContextsRoot 'g0-admission-helper'
    [System.IO.Directory]::CreateDirectory($buildContextsRoot) | Out-Null
    if ([System.IO.Directory]::Exists($stagingRoot) -or [System.IO.File]::Exists($stagingRoot)) { throw 'E_B0_HELPER_BUILD' }
    [System.IO.Directory]::CreateDirectory($stagingRoot) | Out-Null
    try {
        # All static helper inputs are materialized from the accepted A0 Git
        # objects.  The only non-Git input is the already A1-bound raw manifest.
        $objectMap = [ordered]@{
            'Dockerfile' = 'testframework/docker/Dockerfile.g0-admission-helper'
            'dependency-bundle.schema.v1.json' = 'testframework/docker/dependency-bundle.schema.v1.json'
            'admission/verify_local_dependency_bundle.py' = 'testframework/docker/admission/verify_local_dependency_bundle.py'
            'admission/admit_local_dependency_assets.py' = 'testframework/docker/admission/admit_local_dependency_assets.py'
            'entrypoints/g0-admission-helper.sh' = 'testframework/docker/entrypoints/g0-admission-helper.sh'
        }
        $dockerfileBytes = $null
        $entrypointBytes = $null
        foreach ($destination in $objectMap.Keys) {
            $source = [string]$objectMap[$destination]
            $sourceCommit = Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path $source -Code 'E_B0_HELPER_BUILD'
            $bytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ($sourceCommit + ':' + $source) -Code 'E_B0_HELPER_BUILD'
            $destinationPath = Join-Path $stagingRoot ($destination.Replace('/', [System.IO.Path]::DirectorySeparatorChar))
            [System.IO.Directory]::CreateDirectory((Split-Path -Parent $destinationPath)) | Out-Null
            Write-HubQaBootstrapCreateNew -Path $destinationPath -Bytes $bytes -Code 'E_B0_HELPER_BUILD'
            if ($destination -ceq 'Dockerfile') { $dockerfileBytes = $bytes }
            if ($destination -ceq 'entrypoints/g0-admission-helper.sh') { $entrypointBytes = $bytes }
        }
        [System.IO.Directory]::CreateDirectory((Join-Path $stagingRoot 'controls')) | Out-Null
        Write-HubQaBootstrapCreateNew -Path (Join-Path $stagingRoot 'controls\dependency-bundle.manifest.v1.json') -Bytes $Binding.Controls.dependency.Raw.Bytes -Code 'E_B0_HELPER_BUILD'
        if ($null -eq $dockerfileBytes -or $null -eq $entrypointBytes) { throw 'E_B0_HELPER_BUILD' }

        # A target-owned tag is permitted only after the nonce-bound inspection
        # proved the exact helper base.  Existing tags are foreign/colliding.
        Assert-HubQaBootstrapDockerImageAbsent -DockerClientPath $BaseReceipt.DockerClientPath -Tag $helperBaseTag -Code 'E_B0_HELPER_BUILD'
        [void](Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('tag', $BaseReceipt.HelperBaseImageId, $helperBaseTag) -Code 'E_B0_HELPER_BUILD')
        $taggedBase = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{.Id}}', $helperBaseTag) -Code 'E_B0_HELPER_BUILD' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
        if ($taggedBase.Count -ne 1 -or [string]$taggedBase[0] -cne [string]$BaseReceipt.HelperBaseImageId) { throw 'E_B0_HELPER_BUILD' }
        Add-HubQaBootstrapOwnedImage -Tag $helperBaseTag -ImageId ([string]$BaseReceipt.HelperBaseImageId) -OwnerMarkerSha256 ([string]$Report.MarkerSha256) -RequiresOwnerLabel $false
        Assert-HubQaBootstrapDockerImageAbsent -DockerClientPath $BaseReceipt.DockerClientPath -Tag $helperImageTag -Code 'E_B0_HELPER_BUILD'
        [void](Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('build', '--platform', 'linux/amd64', '--pull=false', '--network=none', '--label', ('hub.qa.owner-marker-sha256=' + [string]$Report.MarkerSha256), '--tag', $helperImageTag, '--file', (Join-Path $stagingRoot 'Dockerfile'), $stagingRoot) -Code 'E_B0_HELPER_BUILD' -TimeoutSeconds $script:HubQaBootstrapCommandSeconds)
        $helperIdOutput = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{.Id}}', $helperImageTag) -Code 'E_B0_HELPER_BUILD' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
        if ($helperIdOutput.Count -ne 1 -or [string]$helperIdOutput[0] -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_HELPER_BUILD' }
        $helperImageId = [string]$helperIdOutput[0]
        $helperLabel = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{index .Config.Labels "hub.qa.owner-marker-sha256"}}', $helperImageId) -Code 'E_B0_HELPER_BUILD' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
        if ($helperLabel.Count -ne 1 -or [string]$helperLabel[0] -cne [string]$Report.MarkerSha256) { throw 'E_B0_HELPER_BUILD' }
        Add-HubQaBootstrapOwnedImage -Tag $helperImageTag -ImageId $helperImageId -OwnerMarkerSha256 ([string]$Report.MarkerSha256) -RequiresOwnerLabel $true
        $record = [ordered]@{
            schema = 'hub.qa.g0-admission-helper-local-image-attestation/v1'
            task = $script:HubQaBootstrapTask
            run = $Run
            owner_marker_sha256 = [string]$Report.MarkerSha256
            local_base_receipt_sha256 = [string]$BaseReceipt.Sha256
            receipt_nonce = [string]$BaseReceipt.Nonce
            helper_local_image_id = $helperImageId
            helper_local_image_id_sha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes($helperImageId))
            bootstrap_helper_python_local_image_id = [string]$BaseReceipt.HelperBaseImageId
            bootstrap_helper_python_local_image_id_sha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes([string]$BaseReceipt.HelperBaseImageId))
            helper_dockerfile_sha256 = Get-HubQaBootstrapSha256 -Bytes $dockerfileBytes
            helper_entrypoint_sha256 = Get-HubQaBootstrapSha256 -Bytes $entrypointBytes
        }
        $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
        $receiptPath = Join-Path $Report.ReportRoot 'bootstrap\g0-admission-helper.local-image-attestation.v1.json'
        Write-HubQaBootstrapCreateNew -Path $receiptPath -Bytes $bytes -Code 'E_B0_HELPER_BUILD'
        return [pscustomobject][ordered]@{
            ImageTag = $helperImageTag
            ImageId = $helperImageId
            RawBytes = $bytes
            Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
            ReceiptPath = $receiptPath
        }
    }
    finally {
        if ([System.IO.Directory]::Exists($stagingRoot)) {
            [void](Remove-HubQaBootstrapOwnedReportTree -Report $Report -RelativePath 'bootstrap/build-context/g0-admission-helper')
        }
    }
}

function New-HubQaBootstrapHelperBaseVerificationReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][string]$Run
    )

    $helperBase = @($Binding.Controls.dependency.Value.bases | Where-Object { [string]$_.base_id -ceq 'bootstrap-helper-python' })
    if ($helperBase.Count -ne 1 -or [string]$helperBase[0].local_image_id -cne [string]$BaseReceipt.HelperBaseImageId) { throw 'E_B0_HELPER_BASE_VERIFICATION' }
    $record = [ordered]@{
        schema = 'hub.qa.bootstrap-helper-base-verification/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        owner_marker_sha256 = [string]$Report.MarkerSha256
        local_base_receipt_sha256 = [string]$BaseReceipt.Sha256
        base_id = 'bootstrap-helper-python'
        local_image_id = [string]$BaseReceipt.HelperBaseImageId
        manifest_digest = [string]$helperBase[0].manifest_digest
        verdict = 'ACCEPT'
    }
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $path = Join-Path $Report.ReportRoot 'bootstrap\bootstrap-helper-python.verification.v1.json'
    Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_HELPER_BASE_VERIFICATION'
    return [pscustomobject][ordered]@{ Path = $path; RawBytes = $bytes; Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes }
}

function Get-HubQaBootstrapHandleObservation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $information = New-Object -TypeName 'HubQaG0NativeFile+ByHandleFileInformation'
    if (-not [HubQaG0NativeFile]::GetFileInformationByHandle($Handle, [ref]$information)) { throw $Code }
    return [pscustomobject][ordered]@{
        Attributes = [uint32]$information.FileAttributes
        VolumeSerialNumber = [uint32]$information.VolumeSerialNumber
        FileIndexHigh = [uint32]$information.FileIndexHigh
        FileIndexLow = [uint32]$information.FileIndexLow
        Length = ([uint64]$information.FileSizeHigh -shl 32) -bor [uint64]$information.FileSizeLow
        LastWriteTimeHigh = [uint32]$information.LastWriteTimeHigh
        LastWriteTimeLow = [uint32]$information.LastWriteTimeLow
    }
}

function Assert-HubQaBootstrapSameHandleObservation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Expected,
        [Parameter(Mandatory = $true)][object]$Actual,
        [Parameter(Mandatory = $true)][string]$Code
    )

    foreach ($field in @('Attributes', 'VolumeSerialNumber', 'FileIndexHigh', 'FileIndexLow', 'Length', 'LastWriteTimeHigh', 'LastWriteTimeLow')) {
        if ($Expected.$field -ne $Actual.$field) { throw $Code }
    }
}

function Get-HubQaBootstrapFinalPathByHandle {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][Microsoft.Win32.SafeHandles.SafeFileHandle]$Handle,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $capacity = 32768
    $builder = New-Object System.Text.StringBuilder $capacity
    $written = [HubQaG0NativeFile]::GetFinalPathNameByHandle($Handle, $builder, [uint32]$capacity, [uint32]0)
    if ($written -eq 0 -or $written -ge $capacity) { throw $Code }
    $path = $builder.ToString()
    if ($path.StartsWith('\\?\UNC\', [System.StringComparison]::OrdinalIgnoreCase)) { $path = '\\' + $path.Substring(8) }
    elseif ($path.StartsWith('\\?\', [System.StringComparison]::OrdinalIgnoreCase)) { $path = $path.Substring(4) }
    return [System.IO.Path]::GetFullPath($path)
}

function Get-HubQaBootstrapPhysicalDirectoryRoot {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $handle = [HubQaG0NativeFile]::CreateFileW(
        [System.IO.Path]::GetFullPath($Path), [HubQaG0NativeFile]::GENERIC_READ,
        [HubQaG0NativeFile]::FILE_SHARE_READ, [IntPtr]::Zero, [HubQaG0NativeFile]::OPEN_EXISTING,
        ([HubQaG0NativeFile]::FILE_FLAG_BACKUP_SEMANTICS -bor [HubQaG0NativeFile]::FILE_FLAG_OPEN_REPARSE_POINT), [IntPtr]::Zero
    )
    if ($null -eq $handle -or $handle.IsInvalid) { throw $Code }
    try {
        $observation = Get-HubQaBootstrapHandleObservation -Handle $handle -Code $Code
        if (($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or ($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) { throw $Code }
        return (Get-HubQaBootstrapFinalPathByHandle -Handle $handle -Code $Code).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    }
    finally { $handle.Dispose() }
}

function Open-HubQaBootstrapSafeRegularFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][int64]$MaximumBytes,
        [Parameter(Mandatory = $true)][string]$Code,
        [switch]$AllowEmpty
    )

    if ($MaximumBytes -lt 1) { throw $Code }
    $fullRoot = [System.IO.Path]::GetFullPath($Root).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $fullPath = [System.IO.Path]::GetFullPath($Path)
    if (-not $fullPath.StartsWith($fullRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
    $directoryPaths = New-Object 'System.Collections.Generic.List[string]'
    $parent = [System.IO.DirectoryInfo]::new((Split-Path -Parent $fullPath))
    while ($null -ne $parent) {
        $candidateDirectory = $parent.FullName.TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
        [void]$directoryPaths.Add($candidateDirectory)
        if ([string]::Equals($candidateDirectory, $fullRoot, [System.StringComparison]::OrdinalIgnoreCase)) { break }
        $parent = $parent.Parent
    }
    if ($null -eq $parent) { throw $Code }
    [string[]]$orderedDirectories = @($directoryPaths.ToArray())
    [array]::Reverse($orderedDirectories)
    $ancestorHandles = New-Object 'System.Collections.Generic.List[object]'
    $handle = $null
    try {
        $physicalRoot = $null
        foreach ($directoryPath in $orderedDirectories) {
            $directoryHandle = [HubQaG0NativeFile]::CreateFileW(
                $directoryPath, [HubQaG0NativeFile]::GENERIC_READ,
                ([HubQaG0NativeFile]::FILE_SHARE_READ -bor [HubQaG0NativeFile]::FILE_SHARE_WRITE),
                [IntPtr]::Zero, [HubQaG0NativeFile]::OPEN_EXISTING,
                ([HubQaG0NativeFile]::FILE_FLAG_BACKUP_SEMANTICS -bor [HubQaG0NativeFile]::FILE_FLAG_OPEN_REPARSE_POINT), [IntPtr]::Zero
            )
            if ($null -eq $directoryHandle -or $directoryHandle.IsInvalid) {
                if ($null -ne $directoryHandle) { $directoryHandle.Dispose() }
                throw $Code
            }
            try {
                $directoryObservation = Get-HubQaBootstrapHandleObservation -Handle $directoryHandle -Code $Code
                if (($directoryObservation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_DIRECTORY) -eq 0 -or ($directoryObservation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0) { throw $Code }
                $directoryFinal = (Get-HubQaBootstrapFinalPathByHandle -Handle $directoryHandle -Code $Code).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
                if ($null -eq $physicalRoot) {
                    if (-not [string]::Equals($directoryFinal, $fullRoot, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
                    $physicalRoot = $directoryFinal
                }
                else {
                    $directoryRelative = $directoryPath.Substring(($fullRoot + [System.IO.Path]::DirectorySeparatorChar).Length)
                    $expectedDirectoryFinal = [System.IO.Path]::GetFullPath((Join-Path $physicalRoot $directoryRelative)).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
                    if (-not [string]::Equals($directoryFinal, $expectedDirectoryFinal, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
                }
                [void]$ancestorHandles.Add($directoryHandle)
                $directoryHandle = $null
            }
            finally {
                if ($null -ne $directoryHandle) { $directoryHandle.Dispose() }
            }
        }
        $relative = $fullPath.Substring(($fullRoot + [System.IO.Path]::DirectorySeparatorChar).Length)
        $expectedFinal = [System.IO.Path]::GetFullPath((Join-Path $physicalRoot $relative))
        # FILE_SHARE_READ only is part of the execution binding: opening fails
        # if any pre-existing handle has write/delete access, and no writer or
        # path replacement can be opened until this returned handle closes.
        $handle = [HubQaG0NativeFile]::CreateFileW(
            $fullPath, [HubQaG0NativeFile]::GENERIC_READ, [HubQaG0NativeFile]::FILE_SHARE_READ,
            [IntPtr]::Zero, [HubQaG0NativeFile]::OPEN_EXISTING, [HubQaG0NativeFile]::FILE_FLAG_OPEN_REPARSE_POINT, [IntPtr]::Zero
        )
        if ($null -eq $handle -or $handle.IsInvalid) { throw $Code }
        $observation = Get-HubQaBootstrapHandleObservation -Handle $handle -Code $Code
        if (($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_DIRECTORY) -ne 0 -or ($observation.Attributes -band [HubQaG0NativeFile]::FILE_ATTRIBUTE_REPARSE_POINT) -ne 0 -or ((-not $AllowEmpty) -and [uint64]$observation.Length -lt 1) -or [uint64]$observation.Length -gt [uint64]$MaximumBytes) { throw $Code }
        $finalPath = Get-HubQaBootstrapFinalPathByHandle -Handle $handle -Code $Code
        if (-not [string]::Equals($finalPath, $expectedFinal, [System.StringComparison]::OrdinalIgnoreCase) -or -not $finalPath.StartsWith($physicalRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
        return [pscustomobject][ordered]@{ Handle = $handle; AncestorHandles = @($ancestorHandles.ToArray()); Observation = $observation; FinalPath = $finalPath; Root = $physicalRoot; ShareMode = 'read-only-no-write-no-delete' }
    }
    catch {
        if ($null -ne $handle) { $handle.Dispose() }
        for ($ancestorIndex = $ancestorHandles.Count - 1; $ancestorIndex -ge 0; $ancestorIndex--) { $ancestorHandles[$ancestorIndex].Dispose() }
        throw
    }
}

function Close-HubQaBootstrapSafeRegularFile {
    [CmdletBinding()]
    param([AllowNull()][object]$Opened)

    if ($null -eq $Opened) { return }
    if ($null -ne $Opened.Handle -and -not $Opened.Handle.IsClosed) { $Opened.Handle.Dispose() }
    $ancestors = @($Opened.AncestorHandles)
    for ($ancestorIndex = $ancestors.Count - 1; $ancestorIndex -ge 0; $ancestorIndex--) {
        if ($null -ne $ancestors[$ancestorIndex] -and -not $ancestors[$ancestorIndex].IsClosed) { $ancestors[$ancestorIndex].Dispose() }
    }
}

function Get-HubQaBootstrapHeldStreamDigest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Opened,
        [Parameter(Mandatory = $true)][int64]$ExpectedBytes,
        [Parameter(Mandatory = $true)][string]$Code,
        [switch]$AllowEmpty
    )

    $pre = Get-HubQaBootstrapHandleObservation -Handle $Opened.Handle -Code $Code
    Assert-HubQaBootstrapSameHandleObservation -Expected $Opened.Observation -Actual $pre -Code $Code
    if ($ExpectedBytes -lt 0 -or ((-not $AllowEmpty) -and $ExpectedBytes -lt 1) -or [uint64]$pre.Length -ne [uint64]$ExpectedBytes) { throw $Code }
    $duplicate = New-HubQaOwnedDuplicateFileHandle -Handle $Opened.Handle -Identity $Code
    $stream = $null
    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        $stream = [System.IO.FileStream]::new($duplicate, [System.IO.FileAccess]::Read, 65536, $false)
        $duplicate = $null
        $buffer = New-Object byte[] 65536
        [int64]$total = 0
        while (($read = $stream.Read($buffer, 0, $buffer.Length)) -gt 0) {
            [void]$algorithm.TransformBlock($buffer, 0, $read, $buffer, 0)
            $total += $read
            if ($total -gt $ExpectedBytes) { throw $Code }
        }
        [void]$algorithm.TransformFinalBlock($buffer, 0, 0)
        $post = Get-HubQaBootstrapHandleObservation -Handle $Opened.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $pre -Actual $post -Code $Code
        if ($total -ne $ExpectedBytes) { throw $Code }
        return [pscustomobject][ordered]@{ Bytes = $total; Sha256 = 'sha256:' + ([System.BitConverter]::ToString($algorithm.Hash).Replace('-', '').ToLowerInvariant()) }
    }
    finally {
        $algorithm.Dispose()
        if ($null -ne $stream) { $stream.Dispose() } elseif ($null -ne $duplicate) { $duplicate.Dispose() }
    }
}

function Read-HubQaBootstrapHeldBoundedIngressControl {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $opened = Open-HubQaBootstrapSafeRegularFile -Path $Path -Root $Root -MaximumBytes 1048576 -Code $Code
    $duplicate = $null
    $stream = $null
    try {
        $pre = Get-HubQaBootstrapHandleObservation -Handle $opened.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $opened.Observation -Actual $pre -Code $Code
        [byte[]]$bytes = New-Object byte[] ([int]$pre.Length)
        $duplicate = New-HubQaOwnedDuplicateFileHandle -Handle $opened.Handle -Identity $Code
        $stream = [System.IO.FileStream]::new($duplicate, [System.IO.FileAccess]::Read, 65536, $false)
        $duplicate = $null
        $offset = 0
        while ($offset -lt $bytes.Length) {
            $read = $stream.Read($bytes, $offset, $bytes.Length - $offset)
            if ($read -le 0) { throw $Code }
            $offset += $read
        }
        if ($stream.ReadByte() -ne -1) { throw $Code }
        $post = Get-HubQaBootstrapHandleObservation -Handle $opened.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $pre -Actual $post -Code $Code
        return [pscustomobject][ordered]@{ Bytes = $bytes; Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes; FinalPath = $opened.FinalPath }
    }
    finally {
        if ($null -ne $stream) { $stream.Dispose() } elseif ($null -ne $duplicate) { $duplicate.Dispose() }
        Close-HubQaBootstrapSafeRegularFile -Opened $opened
    }
}

function Copy-HubQaBootstrapSafeHeldFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$SourcePath,
        [Parameter(Mandatory = $true)][string]$SourceRoot,
        [Parameter(Mandatory = $true)][string]$DestinationPath,
        [Parameter(Mandatory = $true)][string]$DestinationRoot,
        [Parameter(Mandatory = $true)][int64]$ExpectedBytes,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $source = Open-HubQaBootstrapSafeRegularFile -Path $SourcePath -Root $SourceRoot -MaximumBytes $script:HubQaBootstrapMaxIngressAssetBytes -Code $Code
    $input = $null
    $output = $null
    $duplicate = $null
    $algorithm = [System.Security.Cryptography.SHA256]::Create()
    try {
        $pre = Get-HubQaBootstrapHandleObservation -Handle $source.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $source.Observation -Actual $pre -Code $Code
        $effectiveExpectedBytes = if ($ExpectedBytes -eq -1) { [int64]$pre.Length } else { $ExpectedBytes }
        if ($effectiveExpectedBytes -lt 1 -or [uint64]$pre.Length -ne [uint64]$effectiveExpectedBytes) { throw $Code }
        $destinationRootFull = [System.IO.Path]::GetFullPath($DestinationRoot).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
        $destinationFull = [System.IO.Path]::GetFullPath($DestinationPath)
        if (-not $destinationFull.StartsWith($destinationRootFull + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
        [System.IO.Directory]::CreateDirectory((Split-Path -Parent $destinationFull)) | Out-Null
        [void](Get-HubQaBootstrapPhysicalDirectoryRoot -Path $destinationRootFull -Code $Code)
        $duplicate = New-HubQaOwnedDuplicateFileHandle -Handle $source.Handle -Identity $Code
        $input = [System.IO.FileStream]::new($duplicate, [System.IO.FileAccess]::Read, 65536, $false)
        $duplicate = $null
        $output = [System.IO.File]::Open($destinationFull, [System.IO.FileMode]::CreateNew, [System.IO.FileAccess]::Write, [System.IO.FileShare]::Read)
        $buffer = New-Object byte[] 65536
        [int64]$total = 0
        while (($read = $input.Read($buffer, 0, $buffer.Length)) -gt 0) {
            $total += $read
            if ($total -gt $effectiveExpectedBytes) { throw $Code }
            [void]$algorithm.TransformBlock($buffer, 0, $read, $buffer, 0)
            $output.Write($buffer, 0, $read)
        }
        [void]$algorithm.TransformFinalBlock($buffer, 0, 0)
        $output.Flush($true)
        $post = Get-HubQaBootstrapHandleObservation -Handle $source.Handle -Code $Code
        Assert-HubQaBootstrapSameHandleObservation -Expected $pre -Actual $post -Code $Code
        if ($total -ne $effectiveExpectedBytes) { throw $Code }
        $sha256 = 'sha256:' + ([System.BitConverter]::ToString($algorithm.Hash).Replace('-', '').ToLowerInvariant())
        $output.Dispose(); $output = $null
        $destination = Open-HubQaBootstrapSafeRegularFile -Path $destinationFull -Root $destinationRootFull -MaximumBytes $script:HubQaBootstrapMaxIngressAssetBytes -Code $Code
        try {
            $rehash = Get-HubQaBootstrapHeldStreamDigest -Opened $destination -ExpectedBytes $effectiveExpectedBytes -Code $Code
            if ($rehash.Sha256 -cne $sha256) { throw $Code }
        }
        finally { Close-HubQaBootstrapSafeRegularFile -Opened $destination }
        return [pscustomobject][ordered]@{ Bytes = $total; Sha256 = $sha256; SourceFinalPath = $source.FinalPath; DestinationFinalPath = $destinationFull }
    }
    finally {
        $algorithm.Dispose()
        if ($null -ne $output) { $output.Dispose() }
        if ($null -ne $input) { $input.Dispose() } elseif ($null -ne $duplicate) { $duplicate.Dispose() }
        Close-HubQaBootstrapSafeRegularFile -Opened $source
    }
}

function Copy-HubQaBootstrapFounderIngress {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][string]$Run
    )

    $handoffId = [string]$Release.Value.offline_ingress_handoff_id
    $sourceRelativeRoot = '.qa-local-ingress/g0-assets/' + $handoffId
    $sourceRoot = [System.IO.Path]::GetFullPath((Join-Path $Root ($sourceRelativeRoot.Replace('/', [System.IO.Path]::DirectorySeparatorChar))))
    $sourceDirectory = [System.IO.DirectoryInfo]::new($sourceRoot)
    $sourceDirectory.Refresh()
    if (-not $sourceDirectory.Exists -or (($sourceDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    [void](Get-HubQaBootstrapPhysicalDirectoryRoot -Path $sourceRoot -Code 'BLOCKED_EXTERNAL_INGRESS')
    $ownerControl = Read-HubQaBootstrapHeldBoundedIngressControl -Path (Join-Path $sourceRoot 'handoff-owner.v1.json') -Root $sourceRoot -Code 'BLOCKED_EXTERNAL_INGRESS'
    $manifestControl = Read-HubQaBootstrapHeldBoundedIngressControl -Path (Join-Path $sourceRoot 'dependency-bundle.manifest.v1.json') -Root $sourceRoot -Code 'BLOCKED_EXTERNAL_INGRESS'
    foreach ($controlBytes in @([byte[]]$ownerControl.Bytes, [byte[]]$manifestControl.Bytes)) {
        if ($controlBytes.Length -eq 0 -or ($controlBytes.Length -ge 3 -and $controlBytes[0] -eq 0xef -and $controlBytes[1] -eq 0xbb -and $controlBytes[2] -eq 0xbf)) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    }
    try {
        $ownerText = [System.Text.UTF8Encoding]::new($false, $true).GetString([byte[]]$ownerControl.Bytes)
        $manifestText = [System.Text.UTF8Encoding]::new($false, $true).GetString([byte[]]$manifestControl.Bytes)
    }
    catch { throw 'BLOCKED_EXTERNAL_INGRESS' }
    $owner = [pscustomobject][ordered]@{ Bytes = [byte[]]$ownerControl.Bytes; Text = $ownerText; Sha256 = [string]$ownerControl.Sha256 }
    $manifest = [pscustomobject][ordered]@{ Bytes = [byte[]]$manifestControl.Bytes; Text = $manifestText; Sha256 = [string]$manifestControl.Sha256 }
    if (
        $owner.Sha256 -cne [string]$Release.Value.offline_ingress_owner_marker_sha256 -or
        $manifest.Sha256 -cne [string]$Release.Value.dependency_bundle_manifest_sha256 -or
        -not (Test-HubQaBootstrapBytesEqual -Left $manifest.Bytes -Right $Binding.Controls.dependency.Raw.Bytes)
    ) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    $ownerValue = ConvertFrom-HubQaBootstrapClosedJson -Text $owner.Text -Code 'BLOCKED_EXTERNAL_INGRESS'
    Assert-HubQaBootstrapExactKeys -Value $ownerValue -Expected @('schema', 'handoff_id', 'source_commit', 'dependency_bundle_manifest_sha256', 'owner', 'released_at_utc') -Code 'BLOCKED_EXTERNAL_INGRESS'
    if ([string]$ownerValue.schema -cne 'hub.qa.g0-offline-ingress-owner/v1' -or [string]$ownerValue.handoff_id -cne $handoffId -or [string]$ownerValue.source_commit -cne $script:HubQaBootstrapSourceCommit -or [string]$ownerValue.dependency_bundle_manifest_sha256 -cne $manifest.Sha256) {
        throw 'BLOCKED_EXTERNAL_INGRESS'
    }
    $manifestValue = ConvertFrom-HubQaBootstrapClosedJson -Text $manifest.Text -Code 'BLOCKED_EXTERNAL_INGRESS'
    Assert-HubQaBootstrapExactKeys -Value $manifestValue -Expected @('schema', 'bundle_id', 'source_commit', 'platform', 'assets', 'bases', 'bundle_sha256') -Code 'BLOCKED_EXTERNAL_INGRESS'
    $paths = New-Object 'System.Collections.Generic.List[string]'
    [void]$paths.Add('handoff-owner.v1.json')
    [void]$paths.Add('dependency-bundle.manifest.v1.json')
    foreach ($asset in @($manifestValue.assets)) {
        Assert-HubQaBootstrapA1Asset -Asset $asset -Code 'BLOCKED_EXTERNAL_INGRESS'
        [void]$paths.Add([string]$asset.relative_path)
        [void]$paths.Add(('evidence/licenses/assets/' + [string]$asset.asset_id + '.txt'))
        [void]$paths.Add(('evidence/provenance/assets/' + [string]$asset.asset_id + '.json'))
    }
    foreach ($base in @($manifestValue.bases)) {
        Assert-HubQaBootstrapExactKeys -Value $base -Expected @('base_id', 'consumers', 'local_image_id', 'manifest_digest', 'license_spdx', 'license_evidence_sha256', 'provenance_sha256') -Code 'BLOCKED_EXTERNAL_INGRESS'
        [void]$paths.Add(('evidence/licenses/bases/' + [string]$base.base_id + '.txt'))
        [void]$paths.Add(('evidence/provenance/bases/' + [string]$base.base_id + '.json'))
    }
    $expectedPaths = @($paths.ToArray())
    if (($expectedPaths | Select-Object -Unique).Count -ne $expectedPaths.Count) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    $actualFiles = New-Object 'System.Collections.Generic.List[string]'
    foreach ($file in $sourceDirectory.GetFiles('*', [System.IO.SearchOption]::AllDirectories)) {
        if (($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'BLOCKED_EXTERNAL_INGRESS' }
        [void]$actualFiles.Add($file.FullName.Substring($sourceRoot.Length).TrimStart('\').Replace('\', '/'))
    }
    if ($actualFiles.Count -ne $expectedPaths.Count) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    foreach ($expectedPath in $expectedPaths) {
        if ($actualFiles -cnotcontains $expectedPath) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    }
    $expectedDirectories = New-Object 'System.Collections.Generic.HashSet[string]' ([System.StringComparer]::Ordinal)
    foreach ($expectedPath in $expectedPaths) {
        $parentPath = [System.IO.Path]::GetDirectoryName($expectedPath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))
        while (-not [string]::IsNullOrEmpty($parentPath)) {
            [void]$expectedDirectories.Add($parentPath.Replace('\', '/'))
            $parentPath = [System.IO.Path]::GetDirectoryName($parentPath)
        }
    }
    $actualDirectories = New-Object 'System.Collections.Generic.List[string]'
    foreach ($directory in $sourceDirectory.GetDirectories('*', [System.IO.SearchOption]::AllDirectories)) {
        if (($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'BLOCKED_EXTERNAL_INGRESS' }
        [void]$actualDirectories.Add($directory.FullName.Substring($sourceRoot.Length).TrimStart('\').Replace('\', '/'))
    }
    if ($actualDirectories.Count -ne $expectedDirectories.Count) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    foreach ($directoryPath in $actualDirectories) {
        if (-not $expectedDirectories.Contains($directoryPath)) { throw 'BLOCKED_EXTERNAL_INGRESS' }
    }
    $destinationRoot = Join-Path $Report.ReportRoot 'ingress\dependency-bundle'
    if ([System.IO.Directory]::Exists($destinationRoot) -or [System.IO.File]::Exists($destinationRoot)) { throw 'E_B0_ONE_SHOT_REJECTED' }
    [System.IO.Directory]::CreateDirectory($destinationRoot) | Out-Null
    [void](Get-HubQaBootstrapPhysicalDirectoryRoot -Path $destinationRoot -Code 'E_B0_INGRESS_COPY')
    $digestMap = [ordered]@{}
    $byteCounts = [ordered]@{}
    foreach ($relativePath in $expectedPaths) {
        if ($relativePath -notmatch '^[A-Za-z0-9._/-]+$' -or $relativePath.StartsWith('/') -or $relativePath.Contains('..')) { throw 'BLOCKED_EXTERNAL_INGRESS' }
        $destination = Join-Path $destinationRoot ($relativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))
        $expectedBytes = [int64]-1
        $expectedSha256 = $null
        if ($relativePath -ceq 'handoff-owner.v1.json') {
            $expectedBytes = [int64]$owner.Bytes.Length
            $expectedSha256 = [string]$owner.Sha256
        }
        elseif ($relativePath -ceq 'dependency-bundle.manifest.v1.json') {
            $expectedBytes = [int64]$manifest.Bytes.Length
            $expectedSha256 = [string]$manifest.Sha256
        }
        else {
            $manifestAsset = @($manifestValue.assets | Where-Object { [string]$_.relative_path -ceq $relativePath })
            if ($manifestAsset.Count -eq 1) {
                $expectedBytes = [int64]$manifestAsset[0].bytes
                $expectedSha256 = [string]$manifestAsset[0].sha256
            }
        }
        $copy = Copy-HubQaBootstrapSafeHeldFile -SourcePath (Join-Path $sourceRoot ($relativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))) -SourceRoot $sourceRoot -DestinationPath $destination -DestinationRoot $destinationRoot -ExpectedBytes $expectedBytes -Code 'E_B0_INGRESS_COPY'
        if ($null -ne $expectedSha256 -and [string]$copy.Sha256 -cne $expectedSha256) { throw 'BLOCKED_EXTERNAL_INGRESS' }
        $digestMap[$relativePath] = [string]$copy.Sha256
        $byteCounts[$relativePath] = [int64]$copy.Bytes
    }
    $mapText = (($expectedPaths | ForEach-Object { $_ + "`t" + [string]$digestMap[$_] + "`n" }) -join '')
    $mapSha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes($mapText))
    $receiptPaths = New-Object 'System.Collections.Generic.List[object]'
    foreach ($relativePath in $expectedPaths) {
        [void]$receiptPaths.Add([ordered]@{ relative_path = $relativePath; bytes = [int64]$byteCounts[$relativePath]; sha256 = [string]$digestMap[$relativePath] })
    }
    $record = [ordered]@{
        schema = 'hub.qa.g0-ingress-copy-receipt/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        owner_marker_sha256 = [string]$Report.MarkerSha256
        offline_ingress_handoff_id = $handoffId
        source_commit = $script:HubQaBootstrapSourceCommit
        source_owner_marker_sha256 = $owner.Sha256
        dependency_bundle_manifest_sha256 = $manifest.Sha256
        source_root = $sourceRoot
        destination_root = $destinationRoot
        paths = @($receiptPaths.ToArray())
        digest_map_sha256 = $mapSha256
        local_base_receipt_sha256 = [string]$BaseReceipt.Sha256
    }
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $receiptPath = Join-Path $Report.ReportRoot 'bootstrap\ingress-copy-receipt.v1.json'
    Write-HubQaBootstrapCreateNew -Path $receiptPath -Bytes $bytes -Code 'E_B0_INGRESS_COPY'
    return [pscustomobject][ordered]@{
        Root = $destinationRoot
        SourceRoot = $sourceRoot
        OwnerMarkerSha256 = $owner.Sha256
        ManifestRawBytes = $manifest.Bytes
        ManifestSha256 = $manifest.Sha256
        ReceiptRawBytes = $bytes
        ReceiptSha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
        DigestMap = $digestMap
        DigestMapSha256 = $mapSha256
        OrderedPaths = $expectedPaths
        ByteCounts = $byteCounts
        ReceiptPath = $receiptPath
        HandoffId = $handoffId
    }
}

function Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    $closed = Assert-HubQaPrivateBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateHelperImageAttestation -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaClosedIngressMetadata -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaIngressOwnerMarker -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaIngressCopyReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    Assert-HubQaPrivateManifestAndBaseReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    if ([System.IO.File]::Exists($closed.ReceiptPath) -or [System.IO.Directory]::Exists($closed.ReceiptPath)) { throw 'E_B0_ADMISSION_CONTEXT' }
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
    $result = Invoke-HubQaBootstrapProcess -Name 'docker' -FileName ([string]$BootstrapPrivateContext.PrivateDockerClientPath) -Arguments $dockerArguments -TimeoutSeconds $script:HubQaBootstrapCommandSeconds -Code 'E_B0_ADMISSION_CONTEXT' -CountsAgainstTarget
    if ([int]$result.ExitCode -ne 0) { throw 'E_B0_ADMISSION_CONTEXT' }
    Assert-HubQaHelperAdmissionReceipt -Closed $closed -BootstrapPrivateContext $BootstrapPrivateContext
    return $closed.ReceiptPath
}

function Invoke-HubQaBootstrapAdmission {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$HelperImage,
        [Parameter(Mandatory = $true)][object]$Ingress,
        [Parameter(Mandatory = $true)][string]$Run
    )

    # The admission context intentionally has neither RoleClaim nor raw owner
    # marker/image-inspection fields.  It is a separate, non-serializable A0
    # context used once before the helper is removed.
    [hashtable]$AdmissionPrivateContext = @{
        Task = $script:HubQaBootstrapTask
        Run = $Run
        OwnerMarkerSha256 = [string]$Report.MarkerSha256
        PrivateReportRoot = [string]$Report.ReportRoot
        PrivateIngressRoot = [string]$Ingress.Root
        PrivateIngressCopyReceipt = [byte[]]$Ingress.ReceiptRawBytes
        PrivateIngressCopyReceiptSha256 = [string]$Ingress.ReceiptSha256
        PrivateIngressDigestMap = $Ingress.DigestMap
        PrivateIngressDigestMapSha256 = [string]$Ingress.DigestMapSha256
        PrivateIngressOrderedPaths = [string[]]$Ingress.OrderedPaths
        PrivateIngressByteCounts = $Ingress.ByteCounts
        PrivateIngressSourceRoot = [string]$Ingress.SourceRoot
        PrivateIngressDestinationRoot = [string]$Ingress.Root
        PrivateOfflineIngressHandoffId = [string]$Ingress.HandoffId
        PrivateOfflineIngressOwnerMarkerSha256 = [string]$Ingress.OwnerMarkerSha256
        PrivateDependencyBundleManifest = [byte[]]$Ingress.ManifestRawBytes
        PrivateDependencyBundleManifestSha256 = [string]$Ingress.ManifestSha256
        PrivateSourceCommit = $script:HubQaBootstrapSourceCommit
        PrivateLocalBaseReceipt = [byte[]]$BaseReceipt.RawBytes
        PrivateLocalBaseReceiptSha256 = [string]$BaseReceipt.Sha256
        PrivateLocalBaseReceiptNonce = [string]$BaseReceipt.Nonce
        PrivateDockerClientPath = [string]$BaseReceipt.DockerClientPath
        PrivateHelperBuildTag = 'hub-qa-g0-admission-helper:local'
        PrivateHelperImageAttestation = [byte[]]$HelperImage.RawBytes
        PrivateHelperImageAttestationSha256 = [string]$HelperImage.Sha256
        PrivateHelperLocalImageId = [string]$HelperImage.ImageId
        PrivateBootstrapHelperPythonLocalImageId = [string]$BaseReceipt.HelperBaseImageId
        PrivateBootstrapHelperPythonLocalImageIdSha256 = Get-HubQaBootstrapSha256 -Bytes ([System.Text.UTF8Encoding]::new($false).GetBytes([string]$BaseReceipt.HelperBaseImageId))
    }
    if ($AdmissionPrivateContext.Keys.Count -ne 28) { throw 'E_B0_ADMISSION_CONTEXT' }
    $receiptPath = Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext -BootstrapPrivateContext $AdmissionPrivateContext
    if ([string]$receiptPath -cne (Join-Path $Report.ReportRoot 'bootstrap\g0-admission-helper.admission.v1.json')) { throw 'E_B0_ADMISSION_CONTEXT' }
    $receiptRaw = Read-HubQaBootstrapHeldBoundedIngressControl -Path $receiptPath -Root $Report.ReportRoot -Code 'E_B0_ADMISSION_CONTEXT'
    $receipt = ConvertFrom-HubQaStrictCanonicalJson -Bytes ([byte[]]$receiptRaw.Bytes) -Identity 'helper admission receipt for target build binding'
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases',
        'local_base_receipt_sha256', 'receipt_nonce', 'verdict'
    ) -Identity 'helper admission receipt for target build binding'
    if ($receipt.schema -cne 'hub.qa.g0-dependency-admission-receipt/v1' -or $receipt.source_commit -cne $script:HubQaBootstrapSourceCommit -or $receipt.manifest_sha256 -cne $Ingress.ManifestSha256 -or $receipt.local_base_receipt_sha256 -cne $BaseReceipt.Sha256 -or $receipt.receipt_nonce -cne $BaseReceipt.Nonce -or $receipt.verdict -cne 'ACCEPT') {
        throw 'E_B0_ADMISSION_CONTEXT'
    }
    return [pscustomobject][ordered]@{
        Context = $AdmissionPrivateContext
        Receipt = [pscustomobject][ordered]@{ Path = $receiptPath; RawBytes = [byte[]]$receiptRaw.Bytes; Sha256 = [string]$receiptRaw.Sha256; Value = $receipt }
    }
}

function Invoke-HubQaBootstrapHelperCleanup {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$HelperImage,
        [Parameter(Mandatory = $true)][string]$Run
    )

    if ($null -eq $script:HubQaBootstrapCleanupBudgetStopwatch) { $script:HubQaBootstrapCleanupBudgetStopwatch = [System.Diagnostics.Stopwatch]::new() }
    if (-not $script:HubQaBootstrapCleanupBudgetStopwatch.IsRunning) { $script:HubQaBootstrapCleanupBudgetStopwatch.Start() }
    try {
        $containerName = 'hub-qa-' + $script:HubQaBootstrapTask + '-' + $Run + '-admission-helper'
        $removed = New-Object 'System.Collections.Generic.List[string]'
        $ownedContainer = @($script:HubQaBootstrapOwnedContainers | Where-Object { [string]$_.Name -ceq $containerName })
        if ($ownedContainer.Count -ne 1) { throw 'E_B0_HELPER_CLEANUP' }
        try {
        $containerProbe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('container', 'inspect', '--format', '{{.Id}}|{{index .Config.Labels "hub.qa.owner-marker-sha256"}}', $containerName) -Code 'E_B0_HELPER_CLEANUP' -CleanupOperation
        if ([int]$containerProbe.ExitCode -eq 0) {
            $containerLine = ([string]$containerProbe.StandardOutput).Trim()
            if ($containerLine -notmatch '^(sha256:)?[a-f0-9]{64}\|sha256:[a-f0-9]{64}$' -or [string]$containerLine.Split('|')[1] -cne [string]$Report.MarkerSha256) { throw 'E_B0_HELPER_CLEANUP' }
            $observedId = [string]$containerLine.Split('|')[0]
            if ($null -ne $ownedContainer[0].ContainerId -and [string]$ownedContainer[0].ContainerId -cne $observedId) { throw 'E_B0_HELPER_CLEANUP' }
            $ownedContainer[0].ContainerId = $observedId
        }
        elseif ([int]$containerProbe.ExitCode -ne 1) { throw 'E_B0_HELPER_CLEANUP' }
        [void](Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('rm', '--force', $containerName) -Code 'E_B0_HELPER_CLEANUP' -TimeoutSeconds $script:HubQaBootstrapCleanupSeconds -CleanupOperation)
        [void]$removed.Add($containerName)
        }
        catch {
        if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { throw }
        # A successful --rm-less helper can still be absent after a daemon
        # cleanup.  Prove absence with inspect rather than adopting anything.
        $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('container', 'inspect', $containerName) -Code 'E_B0_HELPER_CLEANUP' -CleanupOperation
        if ([int]$probe.ExitCode -ne 1) { throw 'E_B0_HELPER_CLEANUP' }
        }
        foreach ($ownedImage in @(
        [pscustomobject]@{ Tag = [string]$HelperImage.ImageTag; ImageId = [string]$HelperImage.ImageId; RequiresOwnerLabel = $true },
        [pscustomobject]@{ Tag = 'hub-qa-bootstrap-helper-python:local'; ImageId = [string]$BaseReceipt.HelperBaseImageId; RequiresOwnerLabel = $false }
        )) {
            try {
            $format = if ($ownedImage.RequiresOwnerLabel) { '{{.Id}}|{{index .Config.Labels "hub.qa.owner-marker-sha256"}}' } else { '{{.Id}}' }
            $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', $format, $ownedImage.Tag) -Code 'E_B0_HELPER_CLEANUP' -CleanupOperation
            if ([int]$probe.ExitCode -eq 0) {
                $line = ([string]$probe.StandardOutput).Trim()
                if ($ownedImage.RequiresOwnerLabel) {
                    if ($line -notmatch '^sha256:[a-f0-9]{64}\|sha256:[a-f0-9]{64}$' -or [string]$line.Split('|')[0] -cne $ownedImage.ImageId -or [string]$line.Split('|')[1] -cne [string]$Report.MarkerSha256) { throw 'E_B0_HELPER_CLEANUP' }
                }
                elseif ($line -cne $ownedImage.ImageId) { throw 'E_B0_HELPER_CLEANUP' }
            }
            elseif ([int]$probe.ExitCode -ne 1) { throw 'E_B0_HELPER_CLEANUP' }
            [void](Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'rm', '--force', $ownedImage.Tag) -Code 'E_B0_HELPER_CLEANUP' -TimeoutSeconds $script:HubQaBootstrapCleanupSeconds -CleanupOperation)
            [void]$removed.Add($ownedImage.Tag)
            }
            catch {
            if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { throw }
            $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', $ownedImage.Tag) -Code 'E_B0_HELPER_CLEANUP' -CleanupOperation
            if ([int]$probe.ExitCode -ne 1) { throw 'E_B0_HELPER_CLEANUP' }
            }
        }
        $record = [ordered]@{
            schema = 'hub.qa.g0-admission-helper-cleanup/v1'
            task = $script:HubQaBootstrapTask
            run = $Run
            owner_marker_sha256 = [string]$Report.MarkerSha256
            status = 'removed'
            resources = @($removed.ToArray())
        }
        $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
        $path = Join-Path $Report.ReportRoot 'bootstrap\g0-admission-helper.cleanup.v1.json'
        Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_HELPER_CLEANUP'
        return [pscustomobject][ordered]@{
            Path = $path
            RawBytes = $bytes
            Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
        }
    }
    finally {
        if ($script:HubQaBootstrapCleanupBudgetStopwatch.IsRunning) { $script:HubQaBootstrapCleanupBudgetStopwatch.Stop() }
    }
}

function New-HubQaBootstrapSourceTreeReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Run
    )

    try {
        $sourceTree = Get-HubQaBootstrapAttestorGitObjectSourceTree -Controls $Binding.Controls
        if (
            [string]$sourceTree.SourceCommit -cne $script:HubQaBootstrapSourceCommit -or
            [string]$sourceTree.SourceTreeSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
            [int]$sourceTree.Entries -lt 1
        ) { throw 'E_B0_SOURCE_TREE_PREBUILD' }
        $record = [ordered]@{
            schema = 'hub.qa.source-tree-verification/v1'
            producer = 'hub.qa-bootstrap/v1'
            task = $script:HubQaBootstrapTask
            run = $Run
            owner_marker_sha256 = [string]$Report.MarkerSha256
            source_commit = [string]$sourceTree.SourceCommit
            source_tree_sha256 = [string]$sourceTree.SourceTreeSha256
            entry_count = [int]$sourceTree.Entries
            source_policy_sha256 = [string]$Binding.Controls.source.Raw.Sha256
            build_manifest_sha256 = [string]$Binding.Controls.build.Raw.Sha256
            created_at_utc = [DateTime]::UtcNow.ToString('o')
        }
        $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
        $path = Join-Path $Report.ReportRoot 'bootstrap\source-tree-verification.v1.json'
        Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_SOURCE_TREE_PREBUILD'
        return [pscustomobject][ordered]@{
            Path = $path
            RawBytes = $bytes
            Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
            Value = $record
        }
    }
    catch {
        throw 'E_B0_SOURCE_TREE_PREBUILD'
    }
}

function Copy-HubQaBootstrapGitObjectToContext {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string]$SourcePath,
        [Parameter(Mandatory = $true)][string]$ContextRoot,
        [Parameter(Mandatory = $true)][string]$DestinationPath,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($SourcePath -notmatch '^[A-Za-z0-9._/-]+$' -or $DestinationPath -notmatch '^[A-Za-z0-9._/-]+$' -or $SourcePath.Contains('..') -or $DestinationPath.Contains('..')) {
        throw $Code
    }
    $destination = [System.IO.Path]::GetFullPath((Join-Path $ContextRoot ($DestinationPath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))))
    $prefix = [System.IO.Path]::GetFullPath($ContextRoot).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    if (-not $destination.StartsWith($prefix + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw $Code }
    [System.IO.Directory]::CreateDirectory((Split-Path -Parent $destination)) | Out-Null
    $bytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName ($Commit + ':' + $SourcePath) -Code $Code
    Write-HubQaBootstrapCreateNew -Path $destination -Bytes $bytes -Code $Code
}

function Copy-HubQaBootstrapGitTreeToContext {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Commit,
        [Parameter(Mandatory = $true)][string]$SourceRoot,
        [Parameter(Mandatory = $true)][string]$ContextRoot,
        [Parameter(Mandatory = $true)][string]$DestinationRoot,
        [Parameter(Mandatory = $true)][string]$Code
    )

    if ($Commit -notmatch '^[a-f0-9]{40}$' -or $SourceRoot -notmatch '^[A-Za-z0-9._/-]+$' -or $DestinationRoot -notmatch '^[A-Za-z0-9._/-]+$') { throw $Code }
    $entries = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('ls-tree', '-r', '--full-tree', $Commit, '--', $SourceRoot) -Code $Code)
    if ($entries.Count -eq 0) { throw $Code }
    foreach ($entry in $entries) {
        if ([string]$entry -notmatch '^100644 blob ([a-f0-9]{40})\t(.+)$') { throw $Code }
        $sourcePath = [string]$Matches[2]
        $prefix = $SourceRoot.TrimEnd('/') + '/'
        if (-not $sourcePath.StartsWith($prefix, [System.StringComparison]::Ordinal) -or $sourcePath.Contains('..') -or $sourcePath.Contains('\')) { throw $Code }
        $relative = $sourcePath.Substring($prefix.Length)
        if ([string]::IsNullOrWhiteSpace($relative)) { throw $Code }
        Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit $Commit -SourcePath $sourcePath -ContextRoot $ContextRoot -DestinationPath ($DestinationRoot.TrimEnd('/') + '/' + $relative) -Code $Code
    }
}

function Assert-HubQaBootstrapStagedArchives {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$AdmissionReceipt,
        [Parameter(Mandatory = $true)][string]$Target,
        [Parameter(Mandatory = $true)][string]$ContextRoot
    )

    $receipt = $AdmissionReceipt.Value
    Assert-HubQaExactPropertyNames -Object $receipt -Expected @(
        'schema', 'bundle_id', 'source_commit', 'manifest_sha256', 'assets', 'bases',
        'local_base_receipt_sha256', 'receipt_nonce', 'verdict'
    ) -Identity 'helper admission receipt before target build'
    if ($receipt.schema -cne 'hub.qa.g0-dependency-admission-receipt/v1' -or $receipt.source_commit -cne $script:HubQaBootstrapSourceCommit -or $receipt.manifest_sha256 -cne $Binding.Controls.dependency.Raw.Sha256 -or $receipt.verdict -cne 'ACCEPT') { throw 'E_ASSET_ARCHIVE_INVALID' }
    $held = New-Object 'System.Collections.Generic.List[object]'
    try {
        foreach ($asset in @($Binding.Controls.dependency.Value.assets)) {
            $applies = @($asset.targets) -ccontains $Target
            if (-not $applies) { continue }
            $relativePath = [string]$asset.relative_path
            if ($relativePath -notmatch '^assets/[a-z0-9-]+\.tar$' -or [int64]$asset.bytes -lt 1 -or [string]$asset.sha256 -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_ASSET_ARCHIVE_INVALID' }
            $receiptAsset = @($receipt.assets | Where-Object { $_.relative_path -ceq $relativePath })
            if ($receiptAsset.Count -ne 1) { throw 'E_ASSET_ARCHIVE_INVALID' }
            Assert-HubQaExactPropertyNames -Object $receiptAsset[0] -Expected @('asset_id', 'targets', 'relative_path', 'sha256', 'bytes', 'archive_format', 'member_count', 'expanded_bytes', 'layout_sha256') -Identity 'helper admission receipt asset before target build'
            if ($receiptAsset[0].asset_id -cne $asset.asset_id -or ((@($receiptAsset[0].targets) -cjoin ',') -cne (@($asset.targets) -cjoin ',')) -or $receiptAsset[0].sha256 -cne $asset.sha256 -or [int64]$receiptAsset[0].bytes -ne [int64]$asset.bytes -or $receiptAsset[0].archive_format -cne 'posix-pax-tar' -or $receiptAsset[0].layout_sha256 -cne $asset.layout_sha256) { throw 'E_ASSET_ARCHIVE_INVALID' }
            $staged = Join-Path $ContextRoot ('bindings\' + $relativePath.Replace('/', '\'))
            $opened = Open-HubQaBootstrapSafeRegularFile -Path $staged -Root $ContextRoot -MaximumBytes ([int64]$asset.bytes) -Code 'E_ASSET_ARCHIVE_INVALID'
            [void]$held.Add($opened)
            $rehash = Get-HubQaBootstrapHeldStreamDigest -Opened $opened -ExpectedBytes ([int64]$asset.bytes) -Code 'E_ASSET_ARCHIVE_INVALID'
            if ([string]$rehash.Sha256 -cne [string]$asset.sha256) { throw 'E_ASSET_ARCHIVE_INVALID' }
        }
        return @($held.ToArray())
    }
    catch {
        foreach ($opened in @($held.ToArray())) { Close-HubQaBootstrapSafeRegularFile -Opened $opened }
        throw
    }
}

function New-HubQaBootstrapFixedBuildContext {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$AdmissionReceipt,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Target
    )

    if ($script:HubQaBootstrapFixedTargets -cnotcontains $Target) { throw 'E_B0_IMAGE_BUILD' }
    $contextsRoot = Join-Path $Report.ReportRoot 'bootstrap\build-context'
    [System.IO.Directory]::CreateDirectory($contextsRoot) | Out-Null
    $contextRoot = Join-Path $contextsRoot $Target
    if ($Target -ceq 'g0-planning-selftest') {
        if ([System.IO.File]::Exists($contextRoot) -or [System.IO.Directory]::Exists($contextRoot)) { throw 'E_B0_IMAGE_BUILD' }
        [System.IO.Directory]::CreateDirectory($contextRoot) | Out-Null
    }
    $contextDirectory = [System.IO.DirectoryInfo]::new($contextRoot)
    $contextDirectory.Refresh()
    if (-not $contextDirectory.Exists -or (($contextDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_ASSET_ARCHIVE_INVALID' }

    [object[]]$stagedArchiveHandles = @()
    if ($Target -ne 'g0-planning-selftest') { $stagedArchiveHandles = @(Assert-HubQaBootstrapStagedArchives -Binding $Binding -AdmissionReceipt $AdmissionReceipt -Target $Target -ContextRoot $contextRoot) }
    try {
        $dockerfileSource = 'testframework/docker/Dockerfile.' + $Target
        $dockerfileCommit = Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path $dockerfileSource -Code 'E_B0_IMAGE_BUILD'
        Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit $dockerfileCommit -SourcePath $dockerfileSource -ContextRoot $contextRoot -DestinationPath 'Dockerfile' -Code 'E_B0_IMAGE_BUILD'

        if ($Target -eq 'g0-planning-selftest') {
            [System.IO.Directory]::CreateDirectory((Join-Path $contextRoot 'controls')) | Out-Null
            foreach ($control in @('dependency-bundle.manifest.v1.json', 'source-policy.v2.json', 'build-manifest.v3.json')) {
                $controlKey = if ($control -like 'dependency-*') { 'dependency' } elseif ($control -like 'source-*') { 'source' } else { 'build' }
                Write-HubQaBootstrapCreateNew -Path (Join-Path $contextRoot ('controls\' + $control)) -Bytes ([byte[]]$Binding.Controls.$controlKey.Raw.Bytes) -Code 'E_B0_IMAGE_BUILD'
            }
            foreach ($source in @(
                'testframework/docker/dependency-bundle.schema.v1.json',
                'testframework/docker/admission/verify_local_dependency_bundle.py',
                'testframework/docker/admission/admit_local_dependency_assets.py',
                'testframework/docker/entrypoints/g0-planning-selftest.sh',
                'testframework/docker/runner-fixtures/g0-planning-selftest.py'
            )) {
                $sourceCommit = Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path $source -Code 'E_B0_IMAGE_BUILD'
                Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit $sourceCommit -SourcePath $source -ContextRoot $contextRoot -DestinationPath $source -Code 'E_B0_IMAGE_BUILD'
            }
            return [pscustomobject][ordered]@{ ContextRoot = $contextRoot; StagedArchiveHandles = @($stagedArchiveHandles) }
        }

        [System.IO.Directory]::CreateDirectory((Join-Path $contextRoot 'bindings')) | Out-Null
        foreach ($control in @('dependency-bundle.manifest.v1.json', 'source-policy.v2.json', 'build-manifest.v3.json')) {
            $controlKey = if ($control -like 'dependency-*') { 'dependency' } elseif ($control -like 'source-*') { 'source' } else { 'build' }
            Write-HubQaBootstrapCreateNew -Path (Join-Path $contextRoot ('bindings\' + $control)) -Bytes ([byte[]]$Binding.Controls.$controlKey.Raw.Bytes) -Code 'E_B0_IMAGE_BUILD'
        }
        if ($Target -eq 'frontend-unit') {
            Copy-HubQaBootstrapGitTreeToContext -Root $Root -Commit $script:HubQaBootstrapSourceCommit -SourceRoot 'frontend' -ContextRoot $contextRoot -DestinationRoot 'source/frontend' -Code 'E_B0_IMAGE_BUILD'
            Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit (Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path 'testframework/docker/entrypoints/frontend-unit.sh' -Code 'E_B0_IMAGE_BUILD') -SourcePath 'testframework/docker/entrypoints/frontend-unit.sh' -ContextRoot $contextRoot -DestinationPath 'entrypoints/frontend-unit.sh' -Code 'E_B0_IMAGE_BUILD'
            Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit (Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path 'testframework/docker/vitest.frontend-unit.config.ts' -Code 'E_B0_IMAGE_BUILD') -SourcePath 'testframework/docker/vitest.frontend-unit.config.ts' -ContextRoot $contextRoot -DestinationPath 'config/vitest.frontend-unit.config.ts' -Code 'E_B0_IMAGE_BUILD'
        }
        else {
            Copy-HubQaBootstrapGitTreeToContext -Root $Root -Commit $script:HubQaBootstrapSourceCommit -SourceRoot 'backend' -ContextRoot $contextRoot -DestinationRoot 'source/backend' -Code 'E_B0_IMAGE_BUILD'
            $entrypointSource = 'testframework/docker/entrypoints/' + $Target + '.sh'
            Copy-HubQaBootstrapGitObjectToContext -Root $Root -Commit (Get-HubQaBootstrapAuthorityCommitForPath -Release $Release -Path $entrypointSource -Code 'E_B0_IMAGE_BUILD') -SourcePath $entrypointSource -ContextRoot $contextRoot -DestinationPath ('entrypoints/' + $Target + '.sh') -Code 'E_B0_IMAGE_BUILD'
        }
        return [pscustomobject][ordered]@{ ContextRoot = $contextRoot; StagedArchiveHandles = @($stagedArchiveHandles) }
    }
    catch {
        foreach ($opened in $stagedArchiveHandles) { Close-HubQaBootstrapSafeRegularFile -Opened $opened }
        throw
    }
}

function Invoke-HubQaBootstrapFourFixedImageBuilds {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$AdmissionReceipt,
        [Parameter(Mandatory = $true)][object]$SourceTreeReceipt,
        [Parameter(Mandatory = $true)][string]$Run
    )

    if ($null -eq $SourceTreeReceipt -or -not [System.IO.File]::Exists($SourceTreeReceipt.Path)) {
        throw 'E_B0_SOURCE_TREE_PREBUILD'
    }
    $imageIds = @{}
    foreach ($target in $script:HubQaBootstrapFixedTargets) {
        $buildContext = New-HubQaBootstrapFixedBuildContext -Root $Root -Release $Release -Binding $Binding -AdmissionReceipt $AdmissionReceipt -Report $Report -Target $target
        $tag = ('hub-qa-' + $script:HubQaBootstrapTask + '-' + $Run + '-' + $target)
        try {
            Assert-HubQaBootstrapDockerImageAbsent -DockerClientPath $BaseReceipt.DockerClientPath -Tag $tag -Code 'E_B0_IMAGE_BUILD'
            [void](Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('build', '--platform', 'linux/amd64', '--pull=false', '--network=none', '--label', ('hub.qa.owner-marker-sha256=' + [string]$Report.MarkerSha256), '--tag', $tag, '--file', (Join-Path $buildContext.ContextRoot 'Dockerfile'), $buildContext.ContextRoot) -Code 'E_B0_IMAGE_BUILD' -TimeoutSeconds $script:HubQaBootstrapCommandSeconds)
            $result = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{.Id}}', $tag) -Code 'E_B0_IMAGE_BUILD' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
            if ($result.Count -ne 1 -or [string]$result[0] -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_IMAGE_BUILD' }
            $ownerLabel = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{index .Config.Labels "hub.qa.owner-marker-sha256"}}', ([string]$result[0])) -Code 'E_B0_IMAGE_BUILD' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
            if ($ownerLabel.Count -ne 1 -or [string]$ownerLabel[0] -cne [string]$Report.MarkerSha256) { throw 'E_B0_IMAGE_BUILD' }
            Add-HubQaBootstrapOwnedImage -Tag $tag -ImageId ([string]$result[0]) -OwnerMarkerSha256 ([string]$Report.MarkerSha256) -RequiresOwnerLabel $true
            $imageIds[$target] = [string]$result[0]
        }
        finally {
            foreach ($opened in @($buildContext.StagedArchiveHandles)) { Close-HubQaBootstrapSafeRegularFile -Opened $opened }
        }
    }
    return $imageIds
}

function New-HubQaBootstrapLocalImageInspectionReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][hashtable]$ImageIds,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][string]$Run
    )

    if ($ImageIds.Keys.Count -ne $script:HubQaBootstrapFixedTargets.Count) { throw 'E_B0_IMAGE_INSPECTION' }
    $targets = [ordered]@{}
    foreach ($target in $script:HubQaBootstrapFixedTargets) {
        if (-not $ImageIds.ContainsKey($target) -or [string]$ImageIds[$target] -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_IMAGE_INSPECTION' }
        $inspectOutput = Invoke-HubQaBootstrapDocker -DockerClientPath $BaseReceipt.DockerClientPath -Arguments @('image', 'inspect', '--format', '{{json .}}', ([string]$ImageIds[$target])) -Code 'E_B0_IMAGE_INSPECTION' -TimeoutSeconds $script:HubQaBootstrapInspectSeconds
        if ($inspectOutput.Count -ne 1) { throw 'E_B0_IMAGE_INSPECTION' }
        try { $inspect = ([string]$inspectOutput[0]) | ConvertFrom-Json -ErrorAction Stop } catch { throw 'E_B0_IMAGE_INSPECTION' }
        if ([string]$inspect.Id -cne [string]$ImageIds[$target]) { throw 'E_B0_IMAGE_INSPECTION' }
        $labels = $inspect.Config.Labels
        if ($null -eq $labels -or [string]$labels.'hub.qa.target' -cne $target -or [string]$labels.'hub.qa.execution-authority' -cne 'hub.qa-bootstrap/v1' -or [string]$labels.'hub.qa.network' -cne 'none' -or [string]$labels.'hub.qa.runtime-status' -cne 'NOT_RUN') {
            throw 'E_B0_IMAGE_INSPECTION'
        }
        $targets[$target] = [ordered]@{
            local_image_id = [string]$inspect.Id
            platform = [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }
            entrypoint = @($inspect.Config.Entrypoint)
            cmd = @($inspect.Config.Cmd)
            labels = [ordered]@{
                'hub.qa.target' = [string]$labels.'hub.qa.target'
                'hub.qa.execution-authority' = [string]$labels.'hub.qa.execution-authority'
                'hub.qa.network' = [string]$labels.'hub.qa.network'
                'hub.qa.runtime-status' = [string]$labels.'hub.qa.runtime-status'
            }
        }
    }
    $record = [ordered]@{
        schema = 'hub.qa.local-image-inspect/v1'
        producer = 'hub.qa-bootstrap/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        source_commit = $script:HubQaBootstrapSourceCommit
        owner_marker_sha256 = [string]$Report.MarkerSha256
        receipt_nonce = [string]$BaseReceipt.Nonce
        controls = [ordered]@{
            dependency_bundle_sha256 = [string]$Binding.Controls.dependency.Raw.Sha256
            source_policy_sha256 = [string]$Binding.Controls.source.Raw.Sha256
            build_manifest_sha256 = [string]$Binding.Controls.build.Raw.Sha256
        }
        targets = $targets
    }
    # The post-build attestor and the planning closure consume these raw bytes.
    # Produce them with the same duplicate-aware canonical codec used by the
    # closed consumer below; ConvertTo-Json is not an authority format.
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $path = Join-Path $Report.ReportRoot 'bootstrap\local-image-inspect.v1.json'
    Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_IMAGE_INSPECTION'
    return [pscustomobject][ordered]@{
        Path = $path
        RawBytes = $bytes
        Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
    }
}

function Invoke-HubQaBootstrapPostBuildAttestation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$Inspection,
        [Parameter(Mandatory = $true)][string]$Run
    )

    # This exact, non-serialized key set is distinct from the admission
    # context.  A0 binds raw owner, post-build inspection, and earlier
    # nonce-bound local-base receipt bytes without reusing admission fields.
    [hashtable]$ImageAttestationPrivateContext = @{
        Task = $script:HubQaBootstrapTask
        Run = $Run
        RoleClaim = 'hub.qa-bootstrap/v1'
        OwnerNonce = [string]$BaseReceipt.Nonce
        OwnerMarkerRawBytes = [byte[]]$Report.MarkerRawBytes
        OwnerMarkerSha256 = [string]$Report.MarkerSha256
        RunnerInspectionRawBytes = [byte[]]$Inspection.RawBytes
        RunnerInspectionSha256 = [string]$Inspection.Sha256
        LocalBaseReceiptRawBytes = [byte[]]$BaseReceipt.RawBytes
        LocalBaseReceiptSha256 = [string]$BaseReceipt.Sha256
        LocalBaseReceiptNonce = [string]$BaseReceipt.Nonce
    }
    if ($ImageAttestationPrivateContext.Keys.Count -ne 11) { throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD' }
    $attestation = Invoke-HubQaBootstrapImageAttestation -BootstrapPrivateContext $ImageAttestationPrivateContext
    if (
        $null -eq $attestation -or
        [string]$attestation.Verdict -cne 'READY_FOR_G0_BOOTSTRAP_ONLY' -or
        [string]$attestation.CandidateLockPath -cne (Join-Path $Report.ReportRoot 'bootstrap\images.lock.v3.json') -or
        [string]$attestation.CandidateLockSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        [string]$attestation.SourceTreeSha256 -cne [string]$Binding.Controls.source.Value.source_tree_sha256
    ) {
        throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD'
    }
    return [pscustomobject][ordered]@{
        Context = $ImageAttestationPrivateContext
        CandidatePath = [string]$attestation.CandidateLockPath
        CandidateSha256 = [string]$attestation.CandidateLockSha256
    }
}

function Get-HubQaBootstrapRunnerInspectionFromPrivateContext {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][hashtable]$BootstrapPrivateContext)

    # The lexical capability is a fixed, in-memory eleven-field record.  Check
    # it before interpreting any raw receipt bytes; no serialized role claim is
    # accepted as a substitute for this exact record.
    Test-HubQaRunnerV2ExactContext -BootstrapPrivateContext $BootstrapPrivateContext
    $inspection = $null
    try {
        $inspectionNode = ConvertFrom-HubQaEnvelopeCanonicalJson -Bytes ([byte[]]$BootstrapPrivateContext.RunnerInspectionRawBytes)
        $inspectionValue = Get-HubQaEnvelopeNodeObject -Node $inspectionNode -Code 'E_B0_SCHEMA_VERSION'
        Assert-HubQaEnvelopeKeys -Object $inspectionValue -Expected @(
            'schema', 'producer', 'task', 'run', 'source_commit', 'owner_marker_sha256', 'receipt_nonce', 'controls', 'targets'
        )
        $schema = Get-HubQaEnvelopeNodeString -Node $inspectionValue['schema'] -Code 'E_B0_SCHEMA_VERSION'
        $producer = Get-HubQaEnvelopeNodeString -Node $inspectionValue['producer'] -Code 'E_B0_SCHEMA_VERSION'
        $task = Get-HubQaEnvelopeNodeString -Node $inspectionValue['task'] -Code 'E_B0_SCHEMA_VERSION'
        $run = Get-HubQaEnvelopeNodeString -Node $inspectionValue['run'] -Code 'E_B0_SCHEMA_VERSION'
        $sourceCommit = Get-HubQaEnvelopeNodeString -Node $inspectionValue['source_commit'] -Code 'E_B0_SCHEMA_VERSION'
        $ownerMarkerSha256 = Get-HubQaEnvelopeNodeString -Node $inspectionValue['owner_marker_sha256'] -Code 'E_B0_SCHEMA_VERSION'
        $receiptNonce = Get-HubQaEnvelopeNodeString -Node $inspectionValue['receipt_nonce'] -Code 'E_B0_SCHEMA_VERSION'
        if ($schema -cne 'hub.qa.local-image-inspect/v1' -or $producer -cne 'hub.qa-bootstrap/v1' -or $sourceCommit -cne $script:HubQaBootstrapSourceCommit -or -not (Test-HubQaEnvelopeDigest -Value $ownerMarkerSha256) -or $receiptNonce -cnotmatch '^[0-9a-f]{32}$') { throw 'E_B0_SCHEMA_VERSION' }

        $controlsValue = Get-HubQaEnvelopeNodeObject -Node $inspectionValue['controls'] -Code 'E_B0_SCHEMA_VERSION'
        Assert-HubQaEnvelopeKeys -Object $controlsValue -Expected @('dependency_bundle_sha256','source_policy_sha256','build_manifest_sha256')
        $controls = [ordered]@{}
        foreach ($name in @('dependency_bundle_sha256','source_policy_sha256','build_manifest_sha256')) {
            $digest = Get-HubQaEnvelopeNodeString -Node $controlsValue[$name] -Code 'E_B0_SCHEMA_VERSION'
            if (-not (Test-HubQaEnvelopeDigest -Value $digest)) { throw 'E_B0_SCHEMA_VERSION' }
            $controls[$name] = $digest
        }

        $targetsValue = Get-HubQaEnvelopeNodeObject -Node $inspectionValue['targets'] -Code 'E_B0_SCHEMA_VERSION'
        Assert-HubQaEnvelopeKeys -Object $targetsValue -Expected $script:HubQaBootstrapFixedTargets
        $targets = [ordered]@{}
        foreach ($targetName in $script:HubQaBootstrapFixedTargets) {
            $targetValue = Get-HubQaEnvelopeNodeObject -Node $targetsValue[$targetName] -Code 'E_B0_SCHEMA_VERSION'
            Assert-HubQaEnvelopeKeys -Object $targetValue -Expected @('local_image_id','platform','entrypoint','cmd','labels')
            $imageId = Get-HubQaEnvelopeNodeString -Node $targetValue['local_image_id'] -Code 'E_B0_SCHEMA_VERSION'
            if ($imageId -cnotmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_SCHEMA_VERSION' }
            $platformValue = Get-HubQaEnvelopeNodeObject -Node $targetValue['platform'] -Code 'E_B0_SCHEMA_VERSION'
            Assert-HubQaEnvelopeKeys -Object $platformValue -Expected @('os','architecture','variant')
            if ((Get-HubQaEnvelopeNodeString -Node $platformValue['os'] -Code 'E_B0_SCHEMA_VERSION') -cne 'linux' -or (Get-HubQaEnvelopeNodeString -Node $platformValue['architecture'] -Code 'E_B0_SCHEMA_VERSION') -cne 'amd64') { throw 'E_B0_SCHEMA_VERSION' }
            Assert-HubQaEnvelopeNodeNull -Node $platformValue['variant'] -Code 'E_B0_SCHEMA_VERSION'
            $entrypoint = Get-HubQaEnvelopeNodeArray -Node $targetValue['entrypoint'] -Code 'E_B0_SCHEMA_VERSION'
            $command = Get-HubQaEnvelopeNodeArray -Node $targetValue['cmd'] -Code 'E_B0_SCHEMA_VERSION'
            foreach ($argument in @($entrypoint) + @($command)) { [void](Get-HubQaEnvelopeNodeString -Node $argument -Code 'E_B0_SCHEMA_VERSION') }
            $labelsValue = Get-HubQaEnvelopeNodeObject -Node $targetValue['labels'] -Code 'E_B0_SCHEMA_VERSION'
            Assert-HubQaEnvelopeKeys -Object $labelsValue -Expected @('hub.qa.target','hub.qa.execution-authority','hub.qa.network','hub.qa.runtime-status')
            $labels = [ordered]@{}
            foreach ($labelName in @('hub.qa.target','hub.qa.execution-authority','hub.qa.network','hub.qa.runtime-status')) { $labels[$labelName] = Get-HubQaEnvelopeNodeString -Node $labelsValue[$labelName] -Code 'E_B0_SCHEMA_VERSION' }
            if ($labels['hub.qa.target'] -cne $targetName -or $labels['hub.qa.execution-authority'] -cne 'hub.qa-bootstrap/v1' -or $labels['hub.qa.network'] -cne 'none' -or $labels['hub.qa.runtime-status'] -cne 'NOT_RUN') { throw 'E_B0_SCHEMA_VERSION' }
            $targets[$targetName] = [ordered]@{ local_image_id = $imageId; platform = [ordered]@{ os = 'linux'; architecture = 'amd64'; variant = $null }; entrypoint = @($entrypoint | ForEach-Object { Get-HubQaEnvelopeNodeString -Node $_ -Code 'E_B0_SCHEMA_VERSION' }); cmd = @($command | ForEach-Object { Get-HubQaEnvelopeNodeString -Node $_ -Code 'E_B0_SCHEMA_VERSION' }); labels = $labels }
        }
        $inspection = [pscustomobject][ordered]@{ schema = $schema; producer = $producer; task = $task; run = $run; source_commit = $sourceCommit; owner_marker_sha256 = $ownerMarkerSha256; receipt_nonce = $receiptNonce; controls = $controls; targets = $targets }
    }
    catch {
        throw 'E_B0_SCHEMA_VERSION'
    }
    if ([string]$inspection.task -cne [string]$BootstrapPrivateContext.Task -or [string]$inspection.run -cne [string]$BootstrapPrivateContext.Run -or [string]$inspection.owner_marker_sha256 -cne [string]$BootstrapPrivateContext.OwnerMarkerSha256 -or [string]$inspection.receipt_nonce -cne [string]$BootstrapPrivateContext.OwnerNonce) { throw 'E_B0_ROLE_TRANSPORT' }
    return $inspection
}

function Assert-HubQaBootstrapDeadline {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Diagnostics.Stopwatch]$Stopwatch,
        [Parameter(Mandatory = $true)][int]$Seconds
    )

    if ($Stopwatch.Elapsed.Ticks -ge ([System.TimeSpan]::FromSeconds($Seconds)).Ticks) { throw 'E_B0_TIMEOUT_124' }
}

function Write-HubQaBootstrapCommandReceipts {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Run,
        [Parameter(Mandatory = $true)][string]$ReleaseNonce
    )

    if ($Run -notmatch $script:HubQaBootstrapRunPattern -or $ReleaseNonce -notmatch '^[0-9a-f]{32}$' -or $script:HubQaBootstrapCommandReceipts.Count -eq 0) {
        throw 'E_B0_COMMAND_RECEIPTS'
    }
    $commands = @($script:HubQaBootstrapCommandReceipts.ToArray())
    foreach ($command in $commands) {
        Assert-HubQaBootstrapExactKeys -Value $command -Expected @('name', 'argv', 'exit_code', 'timed_out', 'started_at_utc', 'finished_at_utc', 'stdout_sha256', 'stderr_sha256') -Code 'E_B0_COMMAND_RECEIPTS'
        # Expected absence probes deliberately retain Docker's exit=1.  The
        # receipt is invocation evidence, not an assertion that every probe
        # succeeded; timed-out or otherwise unhandled failures cannot reach
        # this post-positive writer.
        if ([string]$command.name -notmatch '^[a-z0-9-]{1,64}$' -or @($command.argv).Count -eq 0 -or (@($command.argv | Where-Object { $null -eq $_ -or ([string]$_).IndexOf([char]0) -ge 0 }).Count -ne 0) -or [bool]$command.timed_out -or [string]$command.started_at_utc -notmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{7}Z$' -or [string]$command.finished_at_utc -notmatch '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{7}Z$' -or [string]$command.stdout_sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or [string]$command.stderr_sha256 -notmatch '^sha256:[a-f0-9]{64}$') {
            throw 'E_B0_COMMAND_RECEIPTS'
        }
    }
    $record = [ordered]@{
        schema = 'hub.qa.command-receipts/v1'
        producer = 'hub.qa-bootstrap/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        owner_marker_sha256 = [string]$Report.MarkerSha256
        release_one_shot_nonce = $ReleaseNonce
        commands = $commands
        created_at_utc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    }
    $path = Join-Path $Report.ReportRoot 'commands\command-receipts.v1.json'
    [System.IO.Directory]::CreateDirectory((Split-Path -Parent $path)) | Out-Null
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_COMMAND_RECEIPTS'
    return [pscustomobject][ordered]@{
        Path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/commands/command-receipts.v1.json'
        RawBytes = $bytes
        Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
    }
}

function Assert-HubQaBootstrapCReceiptChain {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Receipts,
        [Parameter(Mandatory = $true)][string]$Run,
        [Parameter(Mandatory = $true)][string]$ReleaseId
    )

    if ($Run -notmatch $script:HubQaBootstrapRunPattern -or $ReleaseId -notmatch '^[a-z][a-z0-9-]{2,63}$') { throw 'E_B0_C_RECEIPT_CHAIN' }
    $actual = @($Receipts.Keys)
    if ($actual.Count -ne $script:HubQaBootstrapCReceiptKeys.Count) { throw 'E_B0_C_RECEIPT_CHAIN' }
    $reportRoot = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/'
    $expectedPaths = @(
        'testframework/reports/g0-assets-runner-nonce-consumptions/' + $ReleaseId + '.nonce-consumption.v1.json',
        $reportRoot + 'bootstrap/owner-marker.v1.json',
        $reportRoot + 'bootstrap/local-base-inspect.v1.json',
        $reportRoot + 'bootstrap/bootstrap-helper-python.verification.v1.json',
        $reportRoot + 'bootstrap/g0-admission-helper.local-image-attestation.v1.json',
        $reportRoot + 'bootstrap/ingress-copy-receipt.v1.json',
        $reportRoot + 'bootstrap/g0-admission-helper.admission.v1.json',
        $reportRoot + 'bootstrap/g0-admission-helper.cleanup.v1.json',
        $reportRoot + 'bootstrap/source-tree-verification.v1.json',
        $reportRoot + 'bootstrap/local-image-inspect.v1.json',
        $reportRoot + 'bootstrap/images.lock.v3.json',
        $reportRoot + 'selftests/g0-planning-negative-red.v1.json',
        $reportRoot + 'selftests/g0-planning-positive-green.v1.json',
        $reportRoot + 'commands/command-receipts.v1.json',
        $reportRoot + 'run.json',
        $reportRoot + 'cleanup.json'
    )
    for ($index = 0; $index -lt $script:HubQaBootstrapCReceiptKeys.Count; $index++) {
        $key = $script:HubQaBootstrapCReceiptKeys[$index]
        if ($actual[$index] -cne $key) { throw 'E_B0_C_RECEIPT_CHAIN' }
        $value = $Receipts[$key]
        Assert-HubQaBootstrapExactKeys -Value $value -Expected @('path', 'raw_sha256') -Code 'E_B0_C_RECEIPT_CHAIN'
        if ([string]$value.path -cne $expectedPaths[$index] -or [string]$value.raw_sha256 -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_C_RECEIPT_CHAIN' }
    }
}

function New-HubQaBootstrapCReceiptChain {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$NonceConsumption,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$BaseReceipt,
        [Parameter(Mandatory = $true)][object]$HelperBaseVerification,
        [Parameter(Mandatory = $true)][object]$HelperImage,
        [Parameter(Mandatory = $true)][object]$Ingress,
        [Parameter(Mandatory = $true)][object]$Admission,
        [Parameter(Mandatory = $true)][object]$HelperCleanup,
        [Parameter(Mandatory = $true)][object]$SourceTreeReceipt,
        [Parameter(Mandatory = $true)][object]$Inspection,
        [Parameter(Mandatory = $true)][object]$PostBuild,
        [Parameter(Mandatory = $true)][object]$PlanningSelftest,
        [Parameter(Mandatory = $true)][object]$CommandReceipts,
        [Parameter(Mandatory = $true)][object]$RunReceipt,
        [Parameter(Mandatory = $true)][object]$Cleanup,
        [Parameter(Mandatory = $true)][string]$Run
    )

    $chain = [ordered]@{
        nonce_consumption = [ordered]@{ path = [string]$NonceConsumption.Path; raw_sha256 = [string]$NonceConsumption.Sha256 }
        owner_marker = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/owner-marker.v1.json'; raw_sha256 = [string]$Report.MarkerSha256 }
        local_base_inspect = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/local-base-inspect.v1.json'; raw_sha256 = [string]$BaseReceipt.Sha256 }
        helper_base_verification = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/bootstrap-helper-python.verification.v1.json'; raw_sha256 = [string]$HelperBaseVerification.Sha256 }
        helper_image_attestation = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/g0-admission-helper.local-image-attestation.v1.json'; raw_sha256 = [string]$HelperImage.Sha256 }
        ingress_copy = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/ingress-copy-receipt.v1.json'; raw_sha256 = [string]$Ingress.ReceiptSha256 }
        helper_admission = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/g0-admission-helper.admission.v1.json'; raw_sha256 = [string]$Admission.Receipt.Sha256 }
        helper_cleanup = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/g0-admission-helper.cleanup.v1.json'; raw_sha256 = [string]$HelperCleanup.Sha256 }
        source_tree_verification = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/source-tree-verification.v1.json'; raw_sha256 = [string]$SourceTreeReceipt.Sha256 }
        local_image_inspection = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/local-image-inspect.v1.json'; raw_sha256 = [string]$Inspection.Sha256 }
        candidate_lock = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/bootstrap/images.lock.v3.json'; raw_sha256 = [string]$PostBuild.CandidateSha256 }
        planning_selftest_negative_red = [ordered]@{ path = [string]$PlanningSelftest.NegativeRedPath; raw_sha256 = [string]$PlanningSelftest.NegativeRedSha256 }
        planning_selftest_positive_green = [ordered]@{ path = [string]$PlanningSelftest.PositiveGreenPath; raw_sha256 = [string]$PlanningSelftest.PositiveGreenSha256 }
        command_receipts = [ordered]@{ path = [string]$CommandReceipts.Path; raw_sha256 = [string]$CommandReceipts.Sha256 }
        run = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/run.json'; raw_sha256 = [string]$RunReceipt.Sha256 }
        cleanup = [ordered]@{ path = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/cleanup.json'; raw_sha256 = [string]$Cleanup.Sha256 }
    }
    Assert-HubQaBootstrapCReceiptChain -Receipts $chain -Run $Run -ReleaseId ([string]$Release.Value.release_id)
    return $chain
}

function Write-HubQaBootstrapRunReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Run,
        [Parameter(Mandatory = $true)][string]$ReleaseNonce,
        [Parameter(Mandatory = $true)][object]$NonceConsumption,
        [Parameter(Mandatory = $true)][object]$Snapshot,
        [Parameter(Mandatory = $true)][object]$Envelope
    )

    $snapshotPath = [string]$Snapshot.SnapshotPath
    $envelopePath = [string]$Envelope.EnvelopePath
    $snapshotRawBytes = [byte[]]$Snapshot.SnapshotRawBytes
    $envelopeRawBytes = [byte[]]$Envelope.EnvelopeRawBytes
    $snapshotSha256 = [string]$Snapshot.SnapshotSha256
    $envelopeSha256 = [string]$Envelope.EnvelopeSha256
    $expectedNoncePath = 'testframework/reports/g0-assets-runner-nonce-consumptions/' + [string]$NonceConsumption.Value.release_id + '.nonce-consumption.v1.json'
    if (
        $Run -notmatch $script:HubQaBootstrapRunPattern -or
        $ReleaseNonce -notmatch '^[0-9a-f]{32}$' -or
        [string]$NonceConsumption.Path -cne $expectedNoncePath -or
        $NonceConsumption.RawBytes -isnot [byte[]] -or
        [string]$NonceConsumption.Sha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        (Get-HubQaBootstrapSha256 -Bytes ([byte[]]$NonceConsumption.RawBytes)) -cne [string]$NonceConsumption.Sha256 -or
        $snapshotPath -cne ('testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json') -or
        $envelopePath -cne ('testframework/reports/' + $script:HubQaBootstrapTask + '-' + $Run + '/envelopes/' + $script:HubQaBootstrapTask + '-envelope-' + $Run + '.run-envelope.v1.json') -or
        $snapshotSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        $envelopeSha256 -notmatch '^sha256:[a-f0-9]{64}$' -or
        $snapshotRawBytes.Length -eq 0 -or
        $envelopeRawBytes.Length -eq 0 -or
        (Get-HubQaBootstrapSha256 -Bytes $snapshotRawBytes) -cne $snapshotSha256 -or
        (Get-HubQaBootstrapSha256 -Bytes $envelopeRawBytes) -cne $envelopeSha256
    ) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $consumption = ConvertFrom-HubQaBootstrapNonceConsumption -RawBytes ([byte[]]$NonceConsumption.RawBytes) -ExpectedReleaseId ([string]$NonceConsumption.Value.release_id) -ExpectedPath $expectedNoncePath
    if ([string]$consumption.Run -cne $Run -or [string]$consumption.OneShotNonce -cne $ReleaseNonce -or [string]$consumption.ReleaseCommit -cne [string]$NonceConsumption.Value.release_commit -or [string]$consumption.ReleaseRawSha256 -cne [string]$NonceConsumption.Value.release_raw_sha256) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $record = [ordered]@{
        schema = 'hub.qa.bootstrap-run/v1'
        producer = 'hub.qa-bootstrap/v1'
        task = $script:HubQaBootstrapTask
        run = $Run
        owner_marker_sha256 = [string]$Report.MarkerSha256
        release_one_shot_nonce = $ReleaseNonce
        repository_snapshot = [ordered]@{ path = $snapshotPath; raw_sha256 = $snapshotSha256 }
        run_envelope = [ordered]@{ path = $envelopePath; raw_sha256 = $envelopeSha256 }
        status = 'completed-zero'
        created_at_utc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    }
    $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
    $path = Join-Path $Report.ReportRoot 'run.json'
    Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_RUN_RECEIPT'
    return [pscustomobject][ordered]@{ Path = $path; RawBytes = $bytes; Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes }
}

function Remove-HubQaBootstrapOwnedReportTree {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][ValidateSet('ingress/dependency-bundle', 'bootstrap/build-context/g0-admission-helper')][string]$RelativePath
    )

    Assert-HubQaBootstrapCleanupAndOuterDeadline
    $reportRoot = [System.IO.Path]::GetFullPath($Report.ReportRoot).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $ownedRoot = [System.IO.Path]::GetFullPath((Join-Path $reportRoot ($RelativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar))))
    if (-not $ownedRoot.StartsWith($reportRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'E_B0_CLEANUP' }
    if (-not [System.IO.Directory]::Exists($ownedRoot)) {
        if ([System.IO.File]::Exists($ownedRoot)) { throw 'E_B0_CLEANUP' }
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        return $false
    }
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    $marker = Read-HubQaBootstrapHeldBoundedIngressControl -Path ([string]$Report.MarkerPath) -Root $reportRoot -Code 'E_B0_CLEANUP'
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    if (-not (Test-HubQaBootstrapBytesEqual -Left ([byte[]]$marker.Bytes) -Right ([byte[]]$Report.MarkerRawBytes)) -or [string]$marker.Sha256 -cne [string]$Report.MarkerSha256) { throw 'E_B0_CLEANUP' }
    $rootDirectory = [System.IO.DirectoryInfo]::new($ownedRoot)
    $rootDirectory.Refresh()
    if (-not $rootDirectory.Exists -or (($rootDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_B0_CLEANUP' }
    [void](Get-HubQaBootstrapPhysicalDirectoryRoot -Path $reportRoot -Code 'E_B0_CLEANUP')
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    $files = @($rootDirectory.GetFiles('*', [System.IO.SearchOption]::AllDirectories))
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    $directories = @($rootDirectory.GetDirectories('*', [System.IO.SearchOption]::AllDirectories) | Sort-Object { $_.FullName.Length } -Descending)
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    foreach ($file in $files) {
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        $file.Refresh()
        if (-not $file.Exists -or (($file.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_B0_CLEANUP' }
        $opened = Open-HubQaBootstrapSafeRegularFile -Path $file.FullName -Root $reportRoot -MaximumBytes $script:HubQaBootstrapMaxIngressAssetBytes -Code 'E_B0_CLEANUP'
        Close-HubQaBootstrapSafeRegularFile -Opened $opened
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        try { [System.IO.File]::Delete($file.FullName) }
        finally { Assert-HubQaBootstrapCleanupAndOuterDeadline }
    }
    foreach ($directory in $directories) {
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        $directory.Refresh()
        if (-not $directory.Exists -or (($directory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_B0_CLEANUP' }
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        try { [System.IO.Directory]::Delete($directory.FullName, $false) }
        finally { Assert-HubQaBootstrapCleanupAndOuterDeadline }
    }
    Assert-HubQaBootstrapCleanupAndOuterDeadline
    try { [System.IO.Directory]::Delete($rootDirectory.FullName, $false) }
    finally { Assert-HubQaBootstrapCleanupAndOuterDeadline }
    return $true
}

function Invoke-HubQaBootstrapExactCleanup {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$Run
    )

    if ($null -eq $script:HubQaBootstrapCleanupBudgetStopwatch) { $script:HubQaBootstrapCleanupBudgetStopwatch = [System.Diagnostics.Stopwatch]::new() }
    if (-not $script:HubQaBootstrapCleanupBudgetStopwatch.IsRunning) { $script:HubQaBootstrapCleanupBudgetStopwatch.Start() }
    $cleanupStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        $removed = New-Object 'System.Collections.Generic.List[string]'
        $cleanupFailure = $false
        try {
            $dockerClient = $null
            try {
                Assert-HubQaBootstrapCleanupAndOuterDeadline
                $dockerClient = Get-HubQaDockerClientPath
                Assert-HubQaBootstrapCleanupAndOuterDeadline
            }
            catch {
                $cleanupError = [string]$_.Exception.Message
                Assert-HubQaBootstrapCleanupAndOuterDeadline
                if ($cleanupError -ceq 'E_B0_TIMEOUT_124') { throw }
                $dockerClient = $null
            }
            if ($null -ne $dockerClient -and [System.IO.File]::Exists($dockerClient)) {
                Assert-HubQaBootstrapCleanupAndOuterDeadline
                foreach ($container in @($script:HubQaBootstrapOwnedContainers.ToArray())) {
                    Assert-HubQaBootstrapCleanupAndOuterDeadline
                    if ($null -eq $container -or [string]::IsNullOrWhiteSpace([string]$container.Name) -or [string]$container.OwnerMarkerSha256 -notmatch '^sha256:[a-f0-9]{64}$') {
                        $cleanupFailure = $true
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        continue
                    }
                    try {
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $dockerClient -Arguments @('container', 'inspect', '--format', '{{.Id}}|{{index .Config.Labels "hub.qa.owner-marker-sha256"}}', [string]$container.Name) -Code 'E_B0_CLEANUP' -CleanupOperation
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        if ([int]$probe.ExitCode -eq 0) {
                            $line = ([string]$probe.StandardOutput).Trim()
                            if ($line -notmatch '^(sha256:)?[a-f0-9]{64}\|sha256:[a-f0-9]{64}$' -or [string]$line.Split('|')[1] -cne [string]$container.OwnerMarkerSha256) { throw 'E_B0_CLEANUP' }
                            $observedId = [string]$line.Split('|')[0]
                            if ($null -ne $container.ContainerId -and [string]$container.ContainerId -cne $observedId) { throw 'E_B0_CLEANUP' }
                            $container.ContainerId = $observedId
                            Assert-HubQaBootstrapCleanupAndOuterDeadline
                            [void](Invoke-HubQaBootstrapDocker -DockerClientPath $dockerClient -Arguments @('rm', '--force', [string]$container.Name) -Code 'E_B0_CLEANUP' -TimeoutSeconds $script:HubQaBootstrapCleanupSeconds -CleanupOperation)
                            Assert-HubQaBootstrapCleanupAndOuterDeadline
                            [void]$removed.Add([string]$container.Name)
                        }
                        elseif ([int]$probe.ExitCode -ne 1) { throw 'E_B0_CLEANUP' }
                    }
                    catch {
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { throw }
                        $cleanupFailure = $true
                    }
                    Assert-HubQaBootstrapCleanupAndOuterDeadline
                }
                Assert-HubQaBootstrapCleanupAndOuterDeadline
                foreach ($resource in @($script:HubQaBootstrapOwnedResources.ToArray())) {
                    Assert-HubQaBootstrapCleanupAndOuterDeadline
                    if ($null -eq $resource -or [string]::IsNullOrWhiteSpace([string]$resource.Tag) -or [string]$resource.ImageId -notmatch '^sha256:[a-f0-9]{64}$') {
                        $cleanupFailure = $true
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        continue
                    }
                    try {
                        $format = if ([bool]$resource.RequiresOwnerLabel) { '{{.Id}}|{{index .Config.Labels "hub.qa.owner-marker-sha256"}}' } else { '{{.Id}}' }
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        $probe = Invoke-HubQaBootstrapDockerProbe -DockerClientPath $dockerClient -Arguments @('image', 'inspect', '--format', $format, [string]$resource.Tag) -Code 'E_B0_CLEANUP' -CleanupOperation
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        if ([int]$probe.ExitCode -eq 0) {
                            $line = ([string]$probe.StandardOutput).Trim()
                            if ([bool]$resource.RequiresOwnerLabel) {
                                if ($line -notmatch '^sha256:[a-f0-9]{64}\|sha256:[a-f0-9]{64}$' -or [string]$line.Split('|')[0] -cne [string]$resource.ImageId -or [string]$line.Split('|')[1] -cne [string]$resource.OwnerMarkerSha256) { throw 'E_B0_CLEANUP' }
                            }
                            elseif ($line -cne [string]$resource.ImageId) { throw 'E_B0_CLEANUP' }
                            Assert-HubQaBootstrapCleanupAndOuterDeadline
                            [void](Invoke-HubQaBootstrapDocker -DockerClientPath $dockerClient -Arguments @('image', 'rm', '--force', [string]$resource.Tag) -Code 'E_B0_CLEANUP' -TimeoutSeconds $script:HubQaBootstrapCleanupSeconds -CleanupOperation)
                            Assert-HubQaBootstrapCleanupAndOuterDeadline
                            [void]$removed.Add([string]$resource.Tag)
                        }
                        elseif ([int]$probe.ExitCode -ne 1) { throw 'E_B0_CLEANUP' }
                    }
                    catch {
                        Assert-HubQaBootstrapCleanupAndOuterDeadline
                        if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { throw }
                        $cleanupFailure = $true
                    }
                    Assert-HubQaBootstrapCleanupAndOuterDeadline
                }
                Assert-HubQaBootstrapCleanupAndOuterDeadline
            }
            Assert-HubQaBootstrapCleanupAndOuterDeadline
            $removedOwnedReportTree = Remove-HubQaBootstrapOwnedReportTree -Report $Report -RelativePath 'ingress/dependency-bundle'
            Assert-HubQaBootstrapCleanupAndOuterDeadline
            if ($removedOwnedReportTree) { [void]$removed.Add('ingress/dependency-bundle') }
        }
        catch {
            Assert-HubQaBootstrapCleanupAndOuterDeadline
            if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { throw }
            $cleanupFailure = $true
        }
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        $record = [ordered]@{
            schema = 'hub.qa.exact-cleanup/v1'
            task = $script:HubQaBootstrapTask
            run = $Run
            owner_marker_sha256 = [string]$Report.MarkerSha256
            status = if ($cleanupFailure) { 'failed' } else { 'succeeded' }
            resources = @($removed.ToArray())
            foreign_resources_adopted = 0
            reports_retained = $true
            duration_ms = [int64]$cleanupStopwatch.ElapsedMilliseconds
        }
        $bytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
        $path = Join-Path $Report.ReportRoot 'cleanup.json'
        try {
            Assert-HubQaBootstrapCleanupAndOuterDeadline
            Write-HubQaBootstrapCreateNew -Path $path -Bytes $bytes -Code 'E_B0_CLEANUP'
            Assert-HubQaBootstrapCleanupAndOuterDeadline
        }
        catch {
            $cleanupError = [string]$_.Exception.Message
            Assert-HubQaBootstrapCleanupAndOuterDeadline
            if ($cleanupError -ceq 'E_B0_TIMEOUT_124') { throw }
            $cleanupFailure = $true
        }
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        $cleanupReceipt = [pscustomobject][ordered]@{
            Succeeded = (-not $cleanupFailure)
            Path = $path
            RawBytes = $bytes
            Sha256 = Get-HubQaBootstrapSha256 -Bytes $bytes
        }
        Assert-HubQaBootstrapCleanupAndOuterDeadline
        return $cleanupReceipt
    }
    finally {
        $cleanupStopwatch.Stop()
        if ($script:HubQaBootstrapCleanupBudgetStopwatch.IsRunning) { $script:HubQaBootstrapCleanupBudgetStopwatch.Stop() }
        Assert-HubQaBootstrapDeadline -Stopwatch $script:HubQaBootstrapCleanupBudgetStopwatch -Seconds $script:HubQaBootstrapCleanupSeconds
        Assert-HubQaBootstrapOuterDeadline
    }
}

function Get-HubQaBootstrapGitSingle {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $result = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments $Arguments -Code $Code)
    if ($result.Count -ne 1 -or [string]::IsNullOrWhiteSpace([string]$result[0])) { throw $Code }
    return [string]$result[0]
}

function Get-HubQaBootstrapNameStatusEntries {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $entries = New-Object 'System.Collections.Generic.List[object]'
    $previous = $null
    foreach ($line in @(Invoke-HubQaBootstrapGit -Root $Root -Arguments $Arguments -Code $Code)) {
        $parts = ([string]$line).Split("`t", 2)
        if ($parts.Count -ne 2 -or [string]::IsNullOrWhiteSpace($parts[0]) -or $parts[1] -notmatch '^(?!/)(?!.*\\)(?!.*(?:^|/)\.\.?(/|$))[A-Za-z0-9._/-]+$') { throw $Code }
        if ($null -ne $previous -and [string]::CompareOrdinal($previous, $parts[1]) -ge 0) { throw $Code }
        $previous = $parts[1]
        [void]$entries.Add([pscustomobject][ordered]@{ Status = [string]$parts[0]; Path = [string]$parts[1] })
    }
    return @($entries.ToArray())
}

function Get-HubQaBootstrapBoundedGitBlobDigest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][string]$ObjectName,
        [Parameter(Mandatory = $true)][string]$Code
    )

    $sizeText = Get-HubQaBootstrapGitSingle -Root $Root -Arguments @('cat-file', '-s', $ObjectName) -Code $Code
    [int64]$declaredBytes = 0
    if (
        $sizeText -cnotmatch '^(?:0|[1-9][0-9]{0,9})$' -or
        -not [int64]::TryParse($sizeText, [System.Globalization.NumberStyles]::None, [System.Globalization.CultureInfo]::InvariantCulture, [ref]$declaredBytes) -or
        $declaredBytes -lt 0 -or $declaredBytes -gt $script:HubQaBootstrapMaxIngressAssetBytes -or $declaredBytes -gt [int64][int]::MaxValue
    ) { throw $Code }
    [byte[]]$bytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName $ObjectName -Code $Code
    if ([int64]$bytes.LongLength -ne $declaredBytes) { throw $Code }
    return Get-HubQaBootstrapSha256 -Bytes $bytes
}

function Get-HubQaBootstrapIndexPathRecords {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string]$Root)

    $records = New-Object 'System.Collections.Generic.List[object]'
    foreach ($entry in @(Get-HubQaBootstrapNameStatusEntries -Root $Root -Arguments @('diff', '--cached', '--name-status', '--no-renames') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK')) {
        $objectName = if ([string]$entry.Status -ceq 'D') { 'HEAD:' + [string]$entry.Path } else { ':' + [string]$entry.Path }
        $digest = Get-HubQaBootstrapBoundedGitBlobDigest -Root $Root -ObjectName $objectName -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
        [void]$records.Add([ordered]@{ path = [string]$entry.Path; status = [string]$entry.Status; blob_sha256 = $digest })
    }
    return @($records.ToArray())
}

function Get-HubQaBootstrapWorktreePathRecords {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string]$Root)

    $records = New-Object 'System.Collections.Generic.List[object]'
    foreach ($entry in @(Get-HubQaBootstrapNameStatusEntries -Root $Root -Arguments @('diff', '--name-status', '--no-renames') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK')) {
        $relativePath = [string]$entry.Path
        if ([string]$entry.Status -ceq 'D') {
            $indexDigest = Get-HubQaBootstrapBoundedGitBlobDigest -Root $Root -ObjectName (':' + $relativePath) -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
            [void]$records.Add([ordered]@{ path = $relativePath; state = [string]$entry.Status; blob_sha256 = $indexDigest })
            continue
        }
        $fullPath = [System.IO.Path]::GetFullPath((Join-Path $Root $relativePath.Replace('/', [System.IO.Path]::DirectorySeparatorChar)))
        $opened = Open-HubQaBootstrapSafeRegularFile -Path $fullPath -Root $Root -MaximumBytes $script:HubQaBootstrapMaxIngressAssetBytes -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK' -AllowEmpty
        try {
            $digest = Get-HubQaBootstrapHeldStreamDigest -Opened $opened -ExpectedBytes ([int64]$opened.Observation.Length) -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK' -AllowEmpty
            [void]$records.Add([ordered]@{ path = $relativePath; state = [string]$entry.Status; blob_sha256 = [string]$digest.Sha256 })
        }
        finally {
            Close-HubQaBootstrapSafeRegularFile -Opened $opened
        }
    }
    return @($records.ToArray())
}

function New-HubQaBootstrapSnapshotGitEvidence {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Root,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$Binding,
        [Parameter(Mandatory = $true)][object]$Report,
        [Parameter(Mandatory = $true)][object]$PostBuild
    )

    $candidate = Read-HubQaBootstrapHeldBoundedIngressControl -Path ([string]$PostBuild.CandidatePath) -Root ([string]$Report.ReportRoot) -Code 'E_B0_IMAGE_ATTESTATION_POSTBUILD'
    $expectedCandidatePath = [System.IO.Path]::GetFullPath((Join-Path ([string]$Report.ReportRoot) 'bootstrap\images.lock.v3.json'))
    if (-not [string]::Equals([string]$candidate.FinalPath, $expectedCandidatePath, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD' }
    [byte[]]$candidateBytes = [byte[]]$candidate.Bytes
    if ([string]$candidate.Sha256 -cne [string]$PostBuild.CandidateSha256) { throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD' }
    $headCommit = Get-HubQaBootstrapGitSingle -Root $Root -Arguments @('rev-parse', 'HEAD') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    if ($headCommit -cne [string]$Release.Commit) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $headTree = Get-HubQaBootstrapGitSingle -Root $Root -Arguments @('rev-parse', 'HEAD^{tree}') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $sourceTree = Get-HubQaBootstrapGitSingle -Root $Root -Arguments @('rev-parse', ($script:HubQaBootstrapSourceCommit + '^{tree}')) -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $parents = @(Invoke-HubQaBootstrapGit -Root $Root -Arguments @('show', '-s', '--format=%P', 'HEAD') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK')
    if ($parents.Count -ne 1) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $parent = [string]$parents[0]
    if ($parent -notmatch '^[a-f0-9]{40}$') { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $headChanged = New-Object 'System.Collections.Generic.List[object]'
    foreach ($entry in @(Get-HubQaBootstrapNameStatusEntries -Root $Root -Arguments @('diff-tree', '--no-commit-id', '--name-status', '--no-renames', '-r', 'HEAD') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK')) {
        $changedObject = if ([string]$entry.Status -ceq 'D') { $parent + ':' + [string]$entry.Path } else { 'HEAD:' + [string]$entry.Path }
        $digest = Get-HubQaBootstrapBoundedGitBlobDigest -Root $Root -ObjectName $changedObject -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
        [void]$headChanged.Add([ordered]@{ path = [string]$entry.Path; status = [string]$entry.Status; blob_sha256 = $digest })
    }
    [byte[]]$planBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    if ($planBytes.Length -eq 0 -or (Get-HubQaBootstrapSha256 -Bytes $planBytes) -cne [string]$Release.G0PlanSha256 -or -not (Test-HubQaBootstrapBytesEqual -Left $planBytes -Right ([byte[]]$Release.G0PlanRawBytes))) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $bootstrapContractBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:testframework/runner/contracts/hub.qa-bootstrap.v1.json' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $runnerContractBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:testframework/runner/run-task-v2.ps1' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $dependencyBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:testframework/docker/dependency-bundle.manifest.v1.json' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $sourcePolicyBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:testframework/docker/source-policy.v2.json' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    $buildManifestBytes = Get-HubQaBootstrapGitObjectBytes -Root $Root -ObjectName 'HEAD:testframework/docker/build-manifest.v3.json' -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    if (
        (Get-HubQaBootstrapSha256 -Bytes $dependencyBytes) -cne [string]$Binding.Controls.dependency.Raw.Sha256 -or
        (Get-HubQaBootstrapSha256 -Bytes $sourcePolicyBytes) -cne [string]$Binding.Controls.source.Raw.Sha256 -or
        (Get-HubQaBootstrapSha256 -Bytes $buildManifestBytes) -cne [string]$Binding.Controls.build.Raw.Sha256
    ) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    $indexPaths = @(Get-HubQaBootstrapIndexPathRecords -Root $Root)
    $worktreePaths = @(Get-HubQaBootstrapWorktreePathRecords -Root $Root)
    $headAfterCapture = Get-HubQaBootstrapGitSingle -Root $Root -Arguments @('rev-parse', 'HEAD') -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    if ($headAfterCapture -cne $headCommit -or $headAfterCapture -cne [string]$Release.Commit) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }
    return [ordered]@{
        CandidateLockPath = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + [string]$Release.Value.run + '/bootstrap/images.lock.v3.json'
        CandidateLockRawBytes = $candidateBytes
        CandidateLockSha256 = [string]$PostBuild.CandidateSha256
        G0PlanSha256 = Get-HubQaBootstrapSha256 -Bytes $planBytes
        BootstrapContractSha256 = Get-HubQaBootstrapSha256 -Bytes $bootstrapContractBytes
        RunnerContractSha256 = Get-HubQaBootstrapSha256 -Bytes $runnerContractBytes
        DependencyBundleManifestSha256 = Get-HubQaBootstrapSha256 -Bytes $dependencyBytes
        SourcePolicySha256 = Get-HubQaBootstrapSha256 -Bytes $sourcePolicyBytes
        BuildManifestSha256 = Get-HubQaBootstrapSha256 -Bytes $buildManifestBytes
        SourceRevisionTree = $sourceTree
        HeadCommit = $headCommit
        HeadTree = $headTree
        IndexPaths = $indexPaths
        WorktreePaths = $worktreePaths
        Commits = @([ordered]@{ commit = $headCommit; parent = $parent; tree = $headTree; changed_paths = @($headChanged.ToArray()) })
        Ancestry = @([ordered]@{ child = $headCommit; parent = $parent })
    }
}

function New-HubQaBootstrapSnapshotPrivateContext {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$SnapshotGitEvidence,
        [Parameter(Mandatory = $true)][object]$Release,
        [Parameter(Mandatory = $true)][object]$PlanningSelftest,
        [Parameter(Mandatory = $true)][object]$Cleanup,
        [Parameter(Mandatory = $true)][string]$StartedAtUtc,
        [Parameter(Mandatory = $true)][string]$FinishedAtUtc
    )

    $evidenceKeys = @(
        'CandidateLockPath','CandidateLockRawBytes','CandidateLockSha256',
        'G0PlanSha256','BootstrapContractSha256','RunnerContractSha256',
        'DependencyBundleManifestSha256','SourcePolicySha256','BuildManifestSha256',
        'SourceRevisionTree','HeadCommit','HeadTree','IndexPaths','WorktreePaths','Commits','Ancestry'
    )
    Assert-HubQaBootstrapExactKeys -Value $SnapshotGitEvidence -Expected $evidenceKeys -Code 'E_B0_SNAPSHOT_ENVELOPE_LINK'
    [byte[]]$candidateBytes = [byte[]]$SnapshotGitEvidence['CandidateLockRawBytes']
    [byte[]]$positiveBytes = [byte[]]$PlanningSelftest.PositiveGreenRawBytes
    [byte[]]$cleanupBytes = [byte[]]$Cleanup.RawBytes
    if (
        $candidateBytes.Length -lt 1 -or (Get-HubQaBootstrapSha256 -Bytes $candidateBytes) -cne [string]$SnapshotGitEvidence['CandidateLockSha256'] -or
        $positiveBytes.Length -lt 1 -or (Get-HubQaBootstrapSha256 -Bytes $positiveBytes) -cne [string]$PlanningSelftest.PositiveGreenSha256 -or
        $cleanupBytes.Length -lt 1 -or (Get-HubQaBootstrapSha256 -Bytes $cleanupBytes) -cne [string]$Cleanup.Sha256
    ) { throw 'E_B0_SNAPSHOT_ENVELOPE_LINK' }

    return [ordered]@{
        PositiveGreenPath = [string]$PlanningSelftest.PositiveGreenPath
        PositiveGreenRawBytes = $positiveBytes
        PositiveGreenSha256 = [string]$PlanningSelftest.PositiveGreenSha256
        CandidateLockPath = [string]$SnapshotGitEvidence['CandidateLockPath']
        CandidateLockRawBytes = $candidateBytes
        CandidateLockSha256 = [string]$SnapshotGitEvidence['CandidateLockSha256']
        CleanupPath = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + [string]$Release.Value.run + '/cleanup.json'
        CleanupRawBytes = $cleanupBytes
        CleanupSha256 = [string]$Cleanup.Sha256
        G0PlanSha256 = [string]$SnapshotGitEvidence['G0PlanSha256']
        BootstrapContractSha256 = [string]$SnapshotGitEvidence['BootstrapContractSha256']
        RunnerContractSha256 = [string]$SnapshotGitEvidence['RunnerContractSha256']
        DependencyBundleManifestSha256 = [string]$SnapshotGitEvidence['DependencyBundleManifestSha256']
        SourcePolicySha256 = [string]$SnapshotGitEvidence['SourcePolicySha256']
        BuildManifestSha256 = [string]$SnapshotGitEvidence['BuildManifestSha256']
        SourceRevisionTree = [string]$SnapshotGitEvidence['SourceRevisionTree']
        HeadCommit = [string]$SnapshotGitEvidence['HeadCommit']
        HeadTree = [string]$SnapshotGitEvidence['HeadTree']
        IndexPaths = @($SnapshotGitEvidence['IndexPaths'])
        WorktreePaths = @($SnapshotGitEvidence['WorktreePaths'])
        Commits = @($SnapshotGitEvidence['Commits'])
        Ancestry = @($SnapshotGitEvidence['Ancestry'])
        StartedAtUtc = $StartedAtUtc
        FinishedAtUtc = $FinishedAtUtc
    }
}

function Invoke-HubQaBootstrapV1 {
    [CmdletBinding()]
    param()

    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $script:HubQaBootstrapRuntimeStopwatch = $stopwatch
    $script:HubQaBootstrapTargetBudgetStopwatch = $null
    $script:HubQaBootstrapCleanupBudgetStopwatch = $null
    $script:HubQaBootstrapCommandReceipts.Clear()
    $firstExit = 0
    $release = $null
    $nonceConsumption = $null
    $report = $null
    $run = $null
    $cleanup = $null
    $cleanupCompleted = $false
    $privateSources = @()
    $startedAtUtc = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture)
    try {
        $root = Get-HubQaBootstrapRepositoryRoot
        $release = Assert-HubQaBootstrapReleaseAuthority -Root $root
        $run = [string]$release.Value.run
        $privateSources = @(Import-HubQaBootstrapPrivateSources -Root $root -Release $release)
        foreach ($privateSource in $privateSources) {
            if ([string]$privateSource.ShareMode -cne 'read-only-no-write-no-delete') { throw 'E_B0_PRIVATE_SOURCE' }
            if ([bool]$privateSource.Describe) {
                $null = . ([string]$privateSource.Path) -Describe:$false
            }
            else {
                $null = . ([string]$privateSource.Path)
            }
            $postImport = Read-HubQaBootstrapHeldFileBytes -Opened $privateSource -MaximumBytes $script:HubQaBootstrapMaxControlBytes -Code 'E_B0_PRIVATE_SOURCE'
            if ([string]$postImport.Sha256 -cne [string]$privateSource.AcceptedSha256) { throw 'E_B0_PRIVATE_SOURCE' }
        }
        # The accepted A0.2/A0.1 blobs are frozen inputs, but their historical
        # process primitives are not execution authority.  Replace all five
        # reachable names in this bootstrap scope and verify the active
        # function bodies point only at B-owned Job Object implementations.
        $boundedOverrides = [ordered]@{
            'Get-HubQaDockerClientPath' = [pscustomobject]@{
                RequiredCall = 'Get-HubQaBootstrapClosedDockerClientPath'
                Body = { return Get-HubQaBootstrapClosedDockerClientPath }
            }
            'Get-HubQaGitBlobBytes' = [pscustomobject]@{
                RequiredCall = 'Get-HubQaBootstrapAttestorGitBlobBytes'
                Body = { param([string]$Repository, [string]$BlobSha1, [string]$Identity); return Get-HubQaBootstrapAttestorGitBlobBytes -Repository $Repository -BlobSha1 $BlobSha1 -Identity $Identity }
            }
            'Get-HubQaGitTreeEntry' = [pscustomobject]@{
                RequiredCall = 'Get-HubQaBootstrapAttestorGitTreeEntry'
                Body = { param([string]$Repository, [string]$RepositoryRelativePath); return Get-HubQaBootstrapAttestorGitTreeEntry -Repository $Repository -RepositoryRelativePath $RepositoryRelativePath }
            }
            'Get-HubQaGitObjectSourceTree' = [pscustomobject]@{
                RequiredCall = 'Get-HubQaBootstrapAttestorGitObjectSourceTree'
                Body = { param([object]$Controls); return Get-HubQaBootstrapAttestorGitObjectSourceTree -Controls $Controls }
            }
            'Invoke-HubQaG0AdmissionFromBootstrapContext' = [pscustomobject]@{
                RequiredCall = 'Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext'
                Body = { param([hashtable]$BootstrapPrivateContext); return Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext -BootstrapPrivateContext $BootstrapPrivateContext }
            }
        }
        foreach ($overrideName in $boundedOverrides.Keys) {
            Set-Item -Path ('Function:\' + $overrideName) -Value $boundedOverrides[$overrideName].Body -Force
            $activeOverride = Get-Command $overrideName -CommandType Function -ErrorAction Stop
            $activeBody = [string]$activeOverride.ScriptBlock.ToString()
            $requiredCall = [string]$boundedOverrides[$overrideName].RequiredCall
            if ($activeBody.IndexOf($requiredCall, [System.StringComparison]::Ordinal) -lt 0) { throw 'E_B0_PRIVATE_SOURCE' }
        }
        $boundedOverrides = $null
        $nonceConsumption = New-HubQaBootstrapNonceConsumptionReceipt -Root $root -Release $release
        Assert-HubQaBootstrapOuterDeadline

        $nonce = New-HubQaBootstrapNonce
        $report = New-HubQaBootstrapReportRoot -Root $root -Run $run -Nonce $nonce
        $script:HubQaBootstrapTargetBudgetStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $admissionDependencyBinding = New-HubQaBootstrapAdmissionDependencyBinding -Root $root -Release $release
        $baseReceipt = New-HubQaBootstrapLocalBaseReceipt -Binding $admissionDependencyBinding -Report $report -Run $run -Nonce $nonce
        Assert-HubQaBootstrapOuterDeadline
        $helperBaseVerification = New-HubQaBootstrapHelperBaseVerificationReceipt -Binding $admissionDependencyBinding -Report $report -BaseReceipt $baseReceipt -Run $run
        $helperImage = New-HubQaBootstrapHelperImage -Root $root -Release $release -Binding $admissionDependencyBinding -Report $report -BaseReceipt $baseReceipt -Run $run
        Assert-HubQaBootstrapOuterDeadline
        $ingress = Copy-HubQaBootstrapFounderIngress -Root $root -Release $release -Binding $admissionDependencyBinding -Report $report -BaseReceipt $baseReceipt -Run $run
        $helperContainerName = 'hub-qa-' + $script:HubQaBootstrapTask + '-' + $run + '-admission-helper'
        Add-HubQaBootstrapOwnedContainer -DockerClientPath $baseReceipt.DockerClientPath -Name $helperContainerName -OwnerMarkerSha256 ([string]$report.MarkerSha256)
        $admission = Invoke-HubQaBootstrapAdmission -Release $release -Report $report -BaseReceipt $baseReceipt -HelperImage $helperImage -Ingress $ingress -Run $run
        Update-HubQaBootstrapOwnedContainerIdentity -DockerClientPath $baseReceipt.DockerClientPath -Name $helperContainerName -Code 'E_B0_ADMISSION_CONTEXT'
        $helperCleanup = Invoke-HubQaBootstrapHelperCleanup -Report $report -BaseReceipt $baseReceipt -HelperImage $helperImage -Run $run
        Assert-HubQaBootstrapOuterDeadline

        $binding = Test-HubQaBootstrapImageBindings
        if ($null -eq $binding -or [string]$binding.Verdict -cne 'READY_FOR_G0_BOOTSTRAP_ONLY' -or
            -not (Test-HubQaBootstrapBytesEqual -Left ([byte[]]$admissionDependencyBinding.Controls.dependency.Raw.Bytes) -Right ([byte[]]$binding.Controls.dependency.Raw.Bytes)) -or
            [string]$admissionDependencyBinding.Controls.dependency.Raw.Sha256 -cne [string]$binding.Controls.dependency.Raw.Sha256) {
            throw 'BLOCKED_EXTERNAL_INGRESS'
        }

        # No target image exists until this mandatory Git-object-only gate has
        # produced its raw, create-new receipt from A1-bound controls.
        $sourceTreeReceipt = New-HubQaBootstrapSourceTreeReceipt -Binding $binding -Report $report -Run $run
        $imageIds = Invoke-HubQaBootstrapFourFixedImageBuilds -Root $root -Release $release -Binding $binding -Report $report -BaseReceipt $baseReceipt -AdmissionReceipt $admission.Receipt -SourceTreeReceipt $sourceTreeReceipt -Run $run
        $inspection = New-HubQaBootstrapLocalImageInspectionReceipt -ImageIds $imageIds -Binding $binding -Report $report -BaseReceipt $baseReceipt -Run $run
        $postBuild = Invoke-HubQaBootstrapPostBuildAttestation -Binding $binding -Report $report -BaseReceipt $baseReceipt -Inspection $inspection -Run $run
        # This capability is minted only after Package-R authority, A0.1/A1
        # gates, admission, source verification, fixed-image build, and the
        # attestor have all accepted in this same bootstrap process.  No
        # dot-sourced runner function receives a Docker-capable parameter.
        [hashtable]$planningPrivateContext = [hashtable]$postBuild.Context
        Test-HubQaRunnerV2ExactContext -BootstrapPrivateContext $planningPrivateContext
        $planningInspection = Get-HubQaBootstrapRunnerInspectionFromPrivateContext -BootstrapPrivateContext $planningPrivateContext
        if (
            [string]$planningPrivateContext.OwnerNonce -cne [string]$baseReceipt.Nonce -or
            [string]$planningPrivateContext.LocalBaseReceiptNonce -cne [string]$baseReceipt.Nonce -or
            [string]$planningPrivateContext.OwnerMarkerSha256 -cne [string]$report.MarkerSha256 -or
            [string]$planningPrivateContext.LocalBaseReceiptSha256 -cne [string]$baseReceipt.Sha256 -or
            -not (Test-HubQaBootstrapBytesEqual -Left ([byte[]]$planningPrivateContext.OwnerMarkerRawBytes) -Right ([byte[]]$report.MarkerRawBytes)) -or
            -not (Test-HubQaBootstrapBytesEqual -Left ([byte[]]$planningPrivateContext.LocalBaseReceiptRawBytes) -Right ([byte[]]$baseReceipt.RawBytes))
        ) { throw 'E_B0_ROLE_TRANSPORT' }
        $privateDockerClientPath = [string]$baseReceipt.DockerClientPath
        if ([string]::IsNullOrWhiteSpace($privateDockerClientPath) -or -not [System.IO.File]::Exists($privateDockerClientPath)) { throw 'E_B0_PLANNING_SELFTEST' }
        $privateReportRoot = [string]$report.ReportRoot
        $privateReportRelative = 'testframework/reports/' + $script:HubQaBootstrapTask + '-' + $run
        $planningContainerName = 'hub-qa-' + $script:HubQaBootstrapTask + '-' + $run + '-g0-planning-selftest-container'
        Add-HubQaBootstrapOwnedContainer -DockerClientPath $privateDockerClientPath -Name $planningContainerName -OwnerMarkerSha256 ([string]$report.MarkerSha256)
        $planningOneShot = [pscustomobject]@{ Consumed = $false }
        $planningSelftestClosure = {
            if ([bool]$planningOneShot.Consumed) { throw 'E_B0_ONE_SHOT_REJECTED' }
            Test-HubQaRunnerV2ExactContext -BootstrapPrivateContext $planningPrivateContext
            $task = [string]$planningPrivateContext.Task
            $privateRun = [string]$planningPrivateContext.Run
            if ($privateRun -cne $run) { throw 'E_B0_ROLE_TRANSPORT' }
            $reportRoot = [System.IO.Path]::GetFullPath($privateReportRoot)
            $reportsRoot = [System.IO.Path]::GetFullPath((Join-Path $root 'testframework\reports')).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
            if (-not $reportRoot.StartsWith($reportsRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'E_B0_ROLE_TRANSPORT' }
            $reportDirectory = [System.IO.DirectoryInfo]::new($reportRoot)
            $reportDirectory.Refresh()
            if (-not $reportDirectory.Exists -or (($reportDirectory.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0)) { throw 'E_B0_ROLE_TRANSPORT' }
            $candidatePath = Join-Path $reportRoot 'bootstrap\images.lock.v3.json'
            $candidate = [System.IO.FileInfo]::new($candidatePath)
            $candidate.Refresh()
            if (-not $candidate.Exists -or (($candidate.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) -or $candidate.Length -lt 1 -or $candidate.Length -gt 1048576) { throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD' }
            if ($null -eq $planningInspection.targets -or $null -eq $planningInspection.targets.'g0-planning-selftest' -or [string]$planningInspection.targets.'g0-planning-selftest'.local_image_id -notmatch '^sha256:[a-f0-9]{64}$') { throw 'E_B0_IMAGE_ATTESTATION_POSTBUILD' }
            $imageId = [string]$planningInspection.targets.'g0-planning-selftest'.local_image_id
            $containerName = 'hub-qa-' + $task + '-' + $privateRun + '-g0-planning-selftest-container'
            $resultPath = Join-Path $reportRoot 'selftests\g0-planning-selftest\result.json'
            $positivePath = Join-Path $reportRoot 'selftests\g0-planning-positive-green.v1.json'
            if ([System.IO.File]::Exists($positivePath) -or [System.IO.Directory]::Exists($positivePath)) { throw 'E_B0_ONE_SHOT_REJECTED' }
            $planningOneShot.Consumed = $true
            try {
                [void](Invoke-HubQaBootstrapDocker -DockerClientPath $privateDockerClientPath -Arguments @('run', '--pull', 'never', '--platform', 'linux/amd64', '--name', $containerName, '--label', ('hub.qa.owner-marker-sha256=' + [string]$planningPrivateContext.OwnerMarkerSha256), '--network', 'none', '--read-only', '--cap-drop', 'ALL', '--security-opt', 'no-new-privileges:true', '--cpus', '2.0', '--memory', '2g', '--pids-limit', '256', '--mount', ('type=bind,src=' + $reportRoot + ',dst=/reports'), '--env', 'A1_BINDING_STATUS=BOUND', '--env', 'HUB_QA_BOOTSTRAP_EXECUTION=hub.qa-bootstrap/v1', $imageId, '--hub-qa-target=g0-planning-selftest') -Code 'E_B0_PLANNING_SELFTEST' -TimeoutSeconds $script:HubQaBootstrapCommandSeconds)
            }
            catch {
                if ([string]$_.Exception.Message -ceq 'E_B0_PLANNING_SELFTEST') { throw }
                throw 'E_B0_PLANNING_SELFTEST'
            }
            $result = [System.IO.FileInfo]::new($resultPath)
            $result.Refresh()
            if (-not $result.Exists -or (($result.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) -or $result.Length -lt 1 -or $result.Length -gt 1048576) { throw 'E_B0_PLANNING_SELFTEST' }
            [byte[]]$resultBytes = [System.IO.File]::ReadAllBytes($result.FullName)
            try {
                $resultText = [System.Text.UTF8Encoding]::new($false, $true).GetString($resultBytes).TrimEnd([char[]]"`r`n")
                $resultValue = ConvertFrom-HubQaBootstrapClosedJson -Text $resultText -Code 'E_B0_PLANNING_SELFTEST'
            }
            catch {
                throw 'E_B0_PLANNING_SELFTEST'
            }
            Assert-HubQaBootstrapExactKeys -Value $resultValue -Expected @('target', 'status', 'step', 'first_exit') -Code 'E_B0_PLANNING_SELFTEST'
            if ([string]$resultValue.target -cne 'g0-planning-selftest' -or [string]$resultValue.status -cne 'PASS' -or [string]$resultValue.step -cne 'shared-python-admission-fixtures-complete' -or [int]$resultValue.first_exit -ne 0) { throw 'E_B0_PLANNING_SELFTEST' }
            $fixtureLogPath = Join-Path $reportRoot 'selftests\g0-planning-selftest\fixture.stdout.log'
            $fixtureLog = [System.IO.FileInfo]::new($fixtureLogPath)
            $fixtureLog.Refresh()
            if (-not $fixtureLog.Exists -or (($fixtureLog.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) -or $fixtureLog.Length -lt 1 -or $fixtureLog.Length -gt 1048576) { throw 'E_B0_PLANNING_SELFTEST' }
            [byte[]]$fixtureLogBytes = [System.IO.File]::ReadAllBytes($fixtureLog.FullName)
            try {
                $fixtureText = [System.Text.UTF8Encoding]::new($false, $true).GetString($fixtureLogBytes).TrimEnd([char[]]"`r`n")
                $fixtureValue = ConvertFrom-HubQaBootstrapClosedJson -Text $fixtureText -Code 'E_B0_PLANNING_SELFTEST'
            }
            catch {
                throw 'E_B0_PLANNING_SELFTEST'
            }
            Assert-HubQaBootstrapExactKeys -Value $fixtureValue -Expected @('target', 'status', 'fixtures') -Code 'E_B0_PLANNING_SELFTEST'
            $expectedFixtures = @(
                'noncanonical-bootstrap-json', 'wrong-version-or-producer', 'short-or-foreign-run',
                'serialized-role-claim', 'direct-planning-selftest', 'release-nonce-consumption',
                'successful-run-receipt-binding', 'snapshot-envelope-drift', 'missing-prebuild-source-receipt',
                'partial-postbuild-attestation', 'postbuild-base-receipt-binding',
                'd-evidence-acceptance-contract', 'activation-r-c1-d-chain', 'first-exit-cleanup'
            )
            $fixtureNames = @($fixtureValue.fixtures)
            if ([string]$fixtureValue.target -cne 'g0-planning-selftest' -or [string]$fixtureValue.status -cne 'PASS' -or $fixtureNames.Count -ne $expectedFixtures.Count) { throw 'E_B0_PLANNING_SELFTEST' }
            for ($fixtureIndex = 0; $fixtureIndex -lt $expectedFixtures.Count; $fixtureIndex++) {
                if ([string]$fixtureNames[$fixtureIndex] -cne $expectedFixtures[$fixtureIndex]) { throw 'E_B0_PLANNING_SELFTEST' }
            }
            $negativeRecord = [ordered]@{
                schema = 'hub.qa.g0-planning-selftest-negative-red/v1'
                producer = 'hub.qa-runner/v2'
                task = $task
                run = $privateRun
                owner_marker_sha256 = [string]$planningPrivateContext.OwnerMarkerSha256
                target = 'g0-planning-selftest'
                fixture_log_path = $privateReportRelative + '/selftests/g0-planning-selftest/fixture.stdout.log'
                fixture_log_sha256 = Get-HubQaRunnerV2Sha256 -Bytes $fixtureLogBytes
                verdict = 'ACCEPT'
            }
            [byte[]]$negativeBytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $negativeRecord))
            $negativePath = Join-Path $reportRoot 'selftests\g0-planning-negative-red.v1.json'
            Write-HubQaRunnerV2CreateNew -Path $negativePath -Bytes $negativeBytes
            $record = [ordered]@{
                schema = 'hub.qa.g0-planning-selftest-positive-green/v1'
                producer = 'hub.qa-runner/v2'
                task = $task
                run = $privateRun
                owner_marker_sha256 = [string]$planningPrivateContext.OwnerMarkerSha256
                target = 'g0-planning-selftest'
                result_path = $privateReportRelative + '/selftests/g0-planning-selftest/result.json'
                result_sha256 = Get-HubQaRunnerV2Sha256 -Bytes $resultBytes
                verdict = 'ACCEPT'
            }
            [byte[]]$positiveBytes = (Get-HubQaEnvelopeUtf8).GetBytes((ConvertTo-HubQaEnvelopeCanonicalJson -Value $record))
            Write-HubQaRunnerV2CreateNew -Path $positivePath -Bytes $positiveBytes
            return [pscustomobject][ordered]@{
                ContainerName = $containerName
                NegativeRedPath = $privateReportRelative + '/selftests/g0-planning-negative-red.v1.json'
                NegativeRedRawBytes = $negativeBytes
                NegativeRedSha256 = Get-HubQaRunnerV2Sha256 -Bytes $negativeBytes
                PositiveGreenPath = $privateReportRelative + '/selftests/g0-planning-positive-green.v1.json'
                PositiveGreenRawBytes = $positiveBytes
                PositiveGreenSha256 = Get-HubQaRunnerV2Sha256 -Bytes $positiveBytes
            }
        }.GetNewClosure()
        $planningSelftest = & $planningSelftestClosure
        $planningSelftestClosure = $null
        if ([string]$planningSelftest.ContainerName -cne $planningContainerName) { throw 'E_B0_PLANNING_SELFTEST' }
        Update-HubQaBootstrapOwnedContainerIdentity -DockerClientPath $privateDockerClientPath -Name $planningContainerName -Code 'E_B0_PLANNING_SELFTEST'
        Assert-HubQaBootstrapOuterDeadline
        # Close every Git query and mutable candidate/index/worktree read into
        # memory before command-receipts seals the complete process transcript.
        Assert-HubQaBootstrapOuterDeadline
        $snapshotGitEvidence = New-HubQaBootstrapSnapshotGitEvidence -Root $root -Release $release -Binding $binding -Report $report -PostBuild $postBuild
        Assert-HubQaBootstrapOuterDeadline
        Assert-HubQaBootstrapOuterDeadline
        $commandReceipts = Write-HubQaBootstrapCommandReceipts -Report $report -Run $run -ReleaseNonce ([string]$release.Value.one_shot_nonce)
        Assert-HubQaBootstrapOuterDeadline

        try {
            Assert-HubQaBootstrapOuterDeadline
            $cleanup = Invoke-HubQaBootstrapExactCleanup -Report $report -Root $root -Run $run
            Assert-HubQaBootstrapOuterDeadline
            $cleanupCompleted = $true
            if (-not [bool]$cleanup.Succeeded) {
                if ($firstExit -eq 0) { $firstExit = 71 }
                throw 'E_B0_CLEANUP'
            }
        }
        catch {
            if ($firstExit -eq 0) {
                if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { $firstExit = 124 }
                else { $firstExit = 71 }
            }
            throw
        }

        # The sibling producers own the closed snapshot/envelope payload
        # construction.  Package B supplies only runner-held, already-created
        # evidence in the one-way order snapshot -> envelope -> run.json.
        $snapshotBootstrapPrivateContext = [ordered]@{
            Task = $script:HubQaBootstrapTask
            Run = $run
            RoleClaim = 'hub.qa-bootstrap/v1'
            OwnerNonce = [string]$baseReceipt.Nonce
            OwnerMarkerRawBytes = [byte[]]$report.MarkerRawBytes
            OwnerMarkerSha256 = [string]$report.MarkerSha256
        }
        Assert-HubQaBootstrapOuterDeadline
        $snapshotPrivateContext = New-HubQaBootstrapSnapshotPrivateContext -SnapshotGitEvidence $snapshotGitEvidence -Release $release -PlanningSelftest $planningSelftest -Cleanup $cleanup -StartedAtUtc $startedAtUtc -FinishedAtUtc ([DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [System.Globalization.CultureInfo]::InvariantCulture))
        Assert-HubQaBootstrapOuterDeadline
        Assert-HubQaBootstrapOuterDeadline
        $snapshot = New-HubQaRepositorySnapshot -BootstrapPrivateContext $snapshotBootstrapPrivateContext -SnapshotPrivateContext $snapshotPrivateContext
        Assert-HubQaBootstrapOuterDeadline
        Assert-HubQaBootstrapOuterDeadline
        $envelope = New-HubQaRunEnvelope -BootstrapPrivateContext $snapshotBootstrapPrivateContext -EnvelopePrivateContext ([ordered]@{
            SnapshotPath = [string]$snapshot.SnapshotPath
            SnapshotRawBytes = [byte[]]$snapshot.SnapshotRawBytes
            SnapshotSha256 = [string]$snapshot.SnapshotSha256
        })
        Assert-HubQaBootstrapOuterDeadline
        Assert-HubQaBootstrapOuterDeadline
        $runReceipt = Write-HubQaBootstrapRunReceipt -Report $report -Run $run -ReleaseNonce ([string]$release.Value.one_shot_nonce) -NonceConsumption $nonceConsumption -Snapshot $snapshot -Envelope $envelope
        Assert-HubQaBootstrapOuterDeadline
        Assert-HubQaBootstrapOuterDeadline
        [void](New-HubQaBootstrapCReceiptChain -Release $release -NonceConsumption $nonceConsumption -Report $report -BaseReceipt $baseReceipt -HelperBaseVerification $helperBaseVerification -HelperImage $helperImage -Ingress $ingress -Admission $admission -HelperCleanup $helperCleanup -SourceTreeReceipt $sourceTreeReceipt -Inspection $inspection -PostBuild $postBuild -PlanningSelftest $planningSelftest -CommandReceipts $commandReceipts -RunReceipt $runReceipt -Cleanup $cleanup -Run $run)
        Assert-HubQaBootstrapOuterDeadline
    }
    catch {
        if ($firstExit -eq 0) {
            if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { $firstExit = 124 }
            else { $firstExit = 65 }
        }
    }
    finally {
        if ($null -ne $report -and -not $cleanupCompleted) {
            try {
                Assert-HubQaBootstrapOuterDeadline
                $cleanup = Invoke-HubQaBootstrapExactCleanup -Report $report -Root $root -Run $run
                Assert-HubQaBootstrapOuterDeadline
                if (-not [bool]$cleanup.Succeeded -and $firstExit -eq 0) { $firstExit = 71 }
            }
            catch {
                if ($firstExit -eq 0) {
                    if ([string]$_.Exception.Message -ceq 'E_B0_TIMEOUT_124') { $firstExit = 124 }
                    else { $firstExit = 71 }
                }
            }
        }
        if ($null -ne $script:HubQaBootstrapTargetBudgetStopwatch) { $script:HubQaBootstrapTargetBudgetStopwatch.Stop() }
        $script:HubQaBootstrapTargetBudgetStopwatch = $null
        if ($null -ne $script:HubQaBootstrapCleanupBudgetStopwatch) { $script:HubQaBootstrapCleanupBudgetStopwatch.Stop() }
        $script:HubQaBootstrapCleanupBudgetStopwatch = $null
        foreach ($privateSource in @($privateSources)) {
            if ($null -ne $privateSource) { Close-HubQaBootstrapSafeRegularFile -Opened $privateSource }
        }
        $privateSources = @()
        if ($firstExit -eq 0) {
            try { Assert-HubQaBootstrapOuterDeadline }
            catch { $firstExit = 124 }
        }
        $stopwatch.Stop()
        if ($firstExit -eq 0) {
            try { Assert-HubQaBootstrapDeadline -Stopwatch $stopwatch -Seconds $script:HubQaBootstrapOuterSeconds }
            catch { $firstExit = 124 }
        }
        $script:HubQaBootstrapRuntimeStopwatch = $null
    }
    return $firstExit
}
