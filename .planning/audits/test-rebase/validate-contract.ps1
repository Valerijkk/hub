[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$RunNonceBytes = [byte[]]::new(16)
[Security.Cryptography.RandomNumberGenerator]::Fill($RunNonceBytes)
$RunNonce = ([Convert]::ToHexString($RunNonceBytes)).ToLowerInvariant()
$RunId = "$PID-$([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds())-$RunNonce"
$TaskRun = "phase128-contract-$RunId"
$ReportRelative = "testframework/reports/$TaskRun"
$ReportRoot = Join-Path $RepoRoot $ReportRelative
$ContainerName = "hub-task204-contract-$RunId"
$SentinelCreateContainer = "hub-task204-sentinel-create-$RunId"
$SentinelCleanupContainer = "hub-task204-sentinel-cleanup-$RunId"
$TimeoutSentinelContainer = "hub-task204-timeout-$RunId"
$OwnedContainers = @(
    $ContainerName,
    $SentinelCreateContainer,
    $SentinelCleanupContainer,
    $TimeoutSentinelContainer
)
$TempBase = [System.IO.Path]::GetTempPath().TrimEnd('\', '/')
$SentinelLeaf = "hub-task204-bind-$RunId"
$SentinelHostRoot = Join-Path $TempBase $SentinelLeaf
$ImageRef = 'mcr.microsoft.com/playwright:v1.61.0-noble'
# Fail closed until an independently obtained digest is reviewed and sealed in
# this versioned wrapper. Environment/CLI overrides are intentionally absent.
$ExpectedImageDigest = 'sha256:753f831a898523662c36011748ba0729b59838745d577c61c91287fe1778bfe1'
# These four values require a separate, independently reviewed local-tool seal.
# No PATH lookup or environment override is permitted as a fallback.
$SealedGitCliPath = 'C:\Program Files\Git\bin\git.exe'
$SealedGitCliSha256 = '544548de8c9bed98eb90531729a2955f554e7009338788b68777258ea61867e4'
$SealedDockerCliPath = 'C:\Program Files\Docker\Docker\resources\bin\docker.exe'
$SealedDockerCliSha256 = 'f6e3fb4ce2342f45c680656e9ae02a47c641efc10ba52eeb4e89f379f6befc6c'
$InventoryValidator = '.planning/audits/test-rebase/inventory.py'
$Validator = '.planning/audits/test-rebase/validate_test_architecture_contract.py'
$Builder = '.planning/audits/test-rebase/build_cutover_manifest.py'
$CutoverValidator = '.planning/audits/test-rebase/validate_cutover_manifest.py'
$BatchApplier = '.planning/audits/test-rebase/apply_cutover_batch.py'
$InventoryAuthority = '.planning/audits/test-rebase/inventory.json'
$InventorySchema = '.planning/audits/test-rebase/inventory.schema.json'
$FounderManifest = '.planning/audits/test-rebase/founder-greenfield-cutover-manifest.json'
$ArchitectureContract = '.planning/audits/test-rebase/test-architecture-contract.json'
$ArchitectureAdr = '.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md'
$EnvelopeAllowlist = @(
    $InventoryValidator,
    $Validator,
    $Builder,
    $CutoverValidator,
    $BatchApplier,
    '.planning/audits/test-rebase/validate-contract.ps1',
    $InventoryAuthority,
    $InventorySchema,
    $FounderManifest,
    $ArchitectureContract,
    $ArchitectureAdr
)
$Utf8NoBom = [System.Text.UTF8Encoding]::new($false)
$Stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
$script:FirstExit = $null
$CleanupExit = 1
$TimeoutExit = -1
$MainDurationMs = 0
$AllResourcesAbsent = $false

Add-Type -TypeDefinition @'
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
using Microsoft.Win32.SafeHandles;

public sealed class HubProcessResult {
    public int ExitCode;
    public string Stdout = "";
    public string Stderr = "";
    public bool HostTimedOut;
    public bool OutputLimitExceeded;
    public bool TerminationGraceExceeded;
}

public sealed class HubPinnedFile : IDisposable {
    internal SafeFileHandle FileHandle;
    public string Identity { get; internal set; }
    public string Digest { get; internal set; }
    public byte[] Bytes { get; internal set; }
    public void Dispose() {
        if (FileHandle != null) FileHandle.Dispose();
        FileHandle = null;
        Bytes = null;
    }
}

public static class HubHandleBoundary {
    [StructLayout(LayoutKind.Sequential)]
    private struct IO_STATUS_BLOCK { public IntPtr Status; public UIntPtr Information; }
    [StructLayout(LayoutKind.Sequential)]
    private struct UNICODE_STRING { public ushort Length, MaximumLength; public IntPtr Buffer; }
    [StructLayout(LayoutKind.Sequential)]
    private struct OBJECT_ATTRIBUTES {
        public int Length; public IntPtr RootDirectory; public IntPtr ObjectName;
        public uint Attributes; public IntPtr SecurityDescriptor; public IntPtr SecurityQualityOfService;
    }
    [StructLayout(LayoutKind.Sequential)]
    private struct BY_HANDLE_FILE_INFORMATION {
        public uint FileAttributes, CreationTimeLow, CreationTimeHigh, LastAccessTimeLow, LastAccessTimeHigh;
        public uint LastWriteTimeLow, LastWriteTimeHigh, VolumeSerialNumber, FileSizeHigh, FileSizeLow;
        public uint NumberOfLinks, FileIndexHigh, FileIndexLow;
    }
    [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)]
    private static extern SafeFileHandle CreateFileW(string path, uint access, uint share, IntPtr security, uint disposition, uint flags, IntPtr template);
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool GetFileInformationByHandle(SafeFileHandle handle, out BY_HANDLE_FILE_INFORMATION info);
    [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)]
    private static extern uint GetFinalPathNameByHandleW(SafeFileHandle handle, StringBuilder path, uint length, uint flags);
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool ReadFile(SafeFileHandle handle, byte[] buffer, uint count, out uint read, IntPtr overlapped);
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool WriteFile(SafeFileHandle handle, byte[] buffer, uint count, out uint written, IntPtr overlapped);
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool SetFilePointerEx(SafeFileHandle handle, long distance, out long position, uint method);
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool FlushFileBuffers(SafeFileHandle handle);
    [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)]
    private static extern bool CreateDirectoryW(string path, IntPtr securityAttributes);
    [DllImport("kernel32.dll")]
    private static extern IntPtr GetCurrentProcess();
    [DllImport("kernel32.dll", SetLastError=true)]
    private static extern bool DuplicateHandle(IntPtr sourceProcess, SafeFileHandle sourceHandle,
        IntPtr targetProcess, out SafeFileHandle targetHandle, uint desiredAccess,
        bool inheritHandle, uint options);
    [DllImport("ntdll.dll")]
    private static extern int NtCreateFile(out SafeFileHandle handle, uint access, ref OBJECT_ATTRIBUTES attributes,
        out IO_STATUS_BLOCK status, IntPtr allocationSize, uint fileAttributes, uint shareAccess,
        uint createDisposition, uint createOptions, IntPtr eaBuffer, uint eaLength);
    [DllImport("ntdll.dll")]
    private static extern uint RtlNtStatusToDosError(int status);

    private const uint GENERIC_READ = 0x80000000, GENERIC_WRITE = 0x40000000;
    private const uint SYNCHRONIZE = 0x00100000, FILE_READ_ATTRIBUTES = 0x80;
    private const uint FILE_GENERIC_READ = 0x00120089;
    private const uint FILE_OPEN = 1, FILE_CREATE = 2, FILE_OPEN_IF = 3;
    private const uint FILE_DIRECTORY_FILE = 0x1, FILE_NON_DIRECTORY_FILE = 0x40;
    private const uint FILE_SYNCHRONOUS_IO_NONALERT = 0x20, FILE_OPEN_REPARSE_POINT = 0x00200000;
    private const uint OBJ_CASE_INSENSITIVE = 0x40, OBJ_DONT_REPARSE = 0x1000;
    private const uint FILE_ATTRIBUTE_REPARSE_POINT = 0x400, FILE_ATTRIBUTE_DIRECTORY = 0x10;

    private static void CheckLeaf(string leaf) {
        if (String.IsNullOrEmpty(leaf) || leaf == "." || leaf == ".." || leaf.IndexOfAny(new[]{'\\','/',':','\0'}) >= 0)
            throw new IOException("unsafe relative artifact leaf");
    }
    private static SafeFileHandle OpenRelative(SafeFileHandle root, string leaf, uint disposition, bool directory, uint access) {
        CheckLeaf(leaf);
        IntPtr buffer = Marshal.StringToHGlobalUni(leaf);
        IntPtr unicodePtr = IntPtr.Zero;
        try {
            UNICODE_STRING unicode = new UNICODE_STRING { Length=(ushort)(leaf.Length*2), MaximumLength=(ushort)((leaf.Length+1)*2), Buffer=buffer };
            unicodePtr = Marshal.AllocHGlobal(Marshal.SizeOf<UNICODE_STRING>());
            Marshal.StructureToPtr(unicode, unicodePtr, false);
            OBJECT_ATTRIBUTES attributes = new OBJECT_ATTRIBUTES {
                Length=Marshal.SizeOf<OBJECT_ATTRIBUTES>(), RootDirectory=root.DangerousGetHandle(), ObjectName=unicodePtr,
                Attributes=OBJ_CASE_INSENSITIVE|OBJ_DONT_REPARSE, SecurityDescriptor=IntPtr.Zero, SecurityQualityOfService=IntPtr.Zero
            };
            IO_STATUS_BLOCK status;
            SafeFileHandle handle;
            uint options=(directory?FILE_DIRECTORY_FILE:FILE_NON_DIRECTORY_FILE)|FILE_SYNCHRONOUS_IO_NONALERT|FILE_OPEN_REPARSE_POINT;
            int result=NtCreateFile(out handle, access|SYNCHRONIZE|FILE_READ_ATTRIBUTES, ref attributes, out status, IntPtr.Zero,
                0x80, directory?0x7u:0x1u, disposition, options, IntPtr.Zero, 0);
            if (result < 0) throw new Win32Exception((int)RtlNtStatusToDosError(result), "handle-relative create/open failed: "+leaf);
            Validate(handle, directory, !directory);
            return handle;
        } finally {
            if (unicodePtr != IntPtr.Zero) Marshal.FreeHGlobal(unicodePtr);
            Marshal.FreeHGlobal(buffer);
        }
    }
    private static BY_HANDLE_FILE_INFORMATION Validate(SafeFileHandle handle, bool directory, bool singleLink) {
        if (handle == null || handle.IsInvalid) throw new IOException("invalid boundary handle");
        BY_HANDLE_FILE_INFORMATION info;
        if (!GetFileInformationByHandle(handle, out info)) throw new Win32Exception(Marshal.GetLastWin32Error(), "handle identity failed");
        if ((info.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0) throw new IOException("reparse handle rejected");
        if (directory != ((info.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0)) throw new IOException("handle type mismatch");
        if (singleLink && info.NumberOfLinks != 1) throw new IOException("artifact must be single-link");
        return info;
    }
    private static string Format(BY_HANDLE_FILE_INFORMATION info) {
        return info.VolumeSerialNumber.ToString("x8")+":"+info.FileIndexHigh.ToString("x8")+info.FileIndexLow.ToString("x8");
    }
    private static string NormalizeFinalPath(string path) {
        if(path.StartsWith(@"\\?\UNC\",StringComparison.OrdinalIgnoreCase))path=@"\\"+path.Substring(8);
        else if(path.StartsWith(@"\\?\",StringComparison.OrdinalIgnoreCase))path=path.Substring(4);
        return Path.GetFullPath(path).TrimEnd(Path.DirectorySeparatorChar,Path.AltDirectorySeparatorChar);
    }
    private static void RequireCanonicalPathHandle(SafeFileHandle handle, string requested) {
        var buffer=new StringBuilder(32768);uint count=GetFinalPathNameByHandleW(handle,buffer,(uint)buffer.Capacity,0);
        if(count==0||count>=(uint)buffer.Capacity)throw new Win32Exception(Marshal.GetLastWin32Error(),"final path query failed");
        string expected=NormalizeFinalPath(requested), actual=NormalizeFinalPath(buffer.ToString());
        if(!String.Equals(expected,actual,StringComparison.OrdinalIgnoreCase))throw new IOException("path contains an ancestor alias/reparse redirect");
    }
    public static SafeFileHandle OpenDirectoryPath(string path) {
        var handle=CreateFileW(path, FILE_READ_ATTRIBUTES|SYNCHRONIZE, 0x7, IntPtr.Zero, 3, FILE_DIRECTORY_FILE|FILE_OPEN_REPARSE_POINT|0x02000000, IntPtr.Zero);
        Validate(handle, true, false); RequireCanonicalPathHandle(handle,path); return handle;
    }
    public static SafeFileHandle OpenDirectoryRelative(SafeFileHandle root, string leaf) { return OpenRelative(root, leaf, FILE_OPEN, true, FILE_READ_ATTRIBUTES); }
    public static SafeFileHandle OpenOrCreateDirectoryRelative(SafeFileHandle root, string leaf) { return OpenRelative(root, leaf, FILE_OPEN_IF, true, FILE_READ_ATTRIBUTES); }
    public static SafeFileHandle CreateDirectoryRelative(SafeFileHandle root, string leaf) { return OpenRelative(root, leaf, FILE_CREATE, true, FILE_READ_ATTRIBUTES); }
    public static string DirectoryIdentity(SafeFileHandle handle) { return Format(Validate(handle, true, false)); }
    private static string Hash(byte[] bytes) { using(var sha=SHA256.Create())return BitConverter.ToString(sha.ComputeHash(bytes)).Replace("-","").ToLowerInvariant(); }
    public static string GitBlobObjectId(byte[] bytes, int hexLength) {
        if(bytes==null)throw new IOException("Git blob bytes are absent");
        if(hexLength!=40&&hexLength!=64)throw new IOException("unsupported Git object format");
        byte[] header=Encoding.ASCII.GetBytes("blob "+bytes.Length.ToString(System.Globalization.CultureInfo.InvariantCulture)+"\0");
        byte[] payload=new byte[header.Length+bytes.Length];
        Buffer.BlockCopy(header,0,payload,0,header.Length);Buffer.BlockCopy(bytes,0,payload,header.Length,bytes.Length);
        using(HashAlgorithm algorithm=hexLength==40?(HashAlgorithm)SHA1.Create():SHA256.Create())
            return BitConverter.ToString(algorithm.ComputeHash(payload)).Replace("-","").ToLowerInvariant();
    }
    private static byte[] ReadAll(SafeFileHandle handle, int maxBytes) {
        var info=Validate(handle,false,true);ulong size=((ulong)info.FileSizeHigh<<32)|info.FileSizeLow;
        if(size>(ulong)maxBytes||size>Int32.MaxValue)throw new IOException("pinned file exceeds byte budget");
        long position;if(!SetFilePointerEx(handle,0,out position,0))throw new Win32Exception(Marshal.GetLastWin32Error(),"pinned file rewind failed");
        byte[] result=new byte[(int)size];int offset=0;
        while(offset<result.Length){int wanted=Math.Min(65536,result.Length-offset);byte[] chunk=new byte[wanted];uint read;
            if(!ReadFile(handle,chunk,(uint)wanted,out read,IntPtr.Zero))throw new Win32Exception(Marshal.GetLastWin32Error(),"pinned file read failed");
            if(read==0)break;Buffer.BlockCopy(chunk,0,result,offset,(int)read);offset+=(int)read;}
        if(offset!=result.Length)throw new IOException("pinned file short read");return result;
    }
    private static void WriteAll(SafeFileHandle handle, byte[] bytes) {
        long position;if(!SetFilePointerEx(handle,0,out position,0))throw new Win32Exception(Marshal.GetLastWin32Error(),"pinned file rewind failed");
        int offset=0;while(offset<bytes.Length){int wanted=Math.Min(65536,bytes.Length-offset);byte[] chunk=new byte[wanted];Buffer.BlockCopy(bytes,offset,chunk,0,wanted);uint written;
            if(!WriteFile(handle,chunk,(uint)wanted,out written,IntPtr.Zero))throw new Win32Exception(Marshal.GetLastWin32Error(),"pinned file write failed");
            if(written==0)throw new IOException("pinned file short write");offset+=(int)written;}
        if(!FlushFileBuffers(handle))throw new Win32Exception(Marshal.GetLastWin32Error(),"pinned file fsync failed");
    }
    public static HubPinnedFile PinFilePath(string path, int maxBytes) {
        var handle=CreateFileW(path,GENERIC_READ,0x1,IntPtr.Zero,3,FILE_OPEN_REPARSE_POINT,IntPtr.Zero);
        try {var info=Validate(handle,false,true);RequireCanonicalPathHandle(handle,path);byte[] bytes=ReadAll(handle,maxBytes);
            return new HubPinnedFile{FileHandle=handle,Identity=Format(info),Digest=Hash(bytes),Bytes=bytes};}
        catch {if(handle!=null)handle.Dispose();throw;}
    }
    public static HubPinnedFile CreatePinnedArtifact(SafeFileHandle root, string leaf, byte[] bytes) {
        var writer=OpenRelative(root,leaf,FILE_CREATE,false,GENERIC_READ|GENERIC_WRITE);
        SafeFileHandle reader=null;
        try {WriteAll(writer,bytes);byte[] actual=ReadAll(writer,bytes.Length);string digest=Hash(actual);var written=Validate(writer,false,true);
            if(digest!=Hash(bytes))throw new IOException("pinned artifact changed during seal");
            IntPtr process=GetCurrentProcess();if(!DuplicateHandle(process,writer,process,out reader,FILE_GENERIC_READ,false,0))throw new Win32Exception(Marshal.GetLastWin32Error(),"read-only pin duplication failed");
            var reopened=Validate(reader,false,true);
            if(Format(reopened)!=Format(written)||Hash(ReadAll(reader,bytes.Length))!=digest)throw new IOException("pinned artifact identity changed while downgrading handle");
            writer.Dispose();writer=null;
            return new HubPinnedFile{FileHandle=reader,Identity=Format(reopened),Digest=digest,Bytes=actual};}
        catch {if(reader!=null)reader.Dispose();if(writer!=null)writer.Dispose();throw;}
    }
    public static string VerifyPinned(HubPinnedFile pinned) {
        if(pinned==null||pinned.FileHandle==null||pinned.FileHandle.IsInvalid||pinned.FileHandle.IsClosed)throw new IOException("pinned file handle is not held");
        var info=Validate(pinned.FileHandle,false,true);byte[] bytes=ReadAll(pinned.FileHandle,pinned.Bytes.Length);string digest=Hash(bytes);
        if(Format(info)!=pinned.Identity||digest!=pinned.Digest)throw new IOException("pinned file identity/digest changed");return digest;
    }
    public static string RunPinnedAbaSentinel(string tempRoot, string nonce) {
        if(String.IsNullOrEmpty(nonce)||nonce.Length!=32)throw new IOException("ABA sentinel nonce rejected");
        foreach(char value in nonce)if(!((value>='0'&&value<='9')||(value>='a'&&value<='f')))throw new IOException("ABA sentinel nonce rejected");
        string directory=Path.Combine(tempRoot,"hub-phase128-aba-"+nonce);
        if(!CreateDirectoryW(directory,IntPtr.Zero))throw new Win32Exception(Marshal.GetLastWin32Error(),"exclusive ABA sentinel directory create failed");
        SafeFileHandle root=null;HubPinnedFile source=null;
        try {root=OpenDirectoryPath(directory);source=CreatePinnedArtifact(root,"source.bin",Encoding.ASCII.GetBytes("immutable-executed-byte\n"));
            CreateArtifact(root,"attacker.bin",Encoding.ASCII.GetBytes("attacker-byte\n"));
            bool moveRejected=false;try{File.Move(Path.Combine(directory,"source.bin"),Path.Combine(directory,"parked.bin"));}catch(IOException){moveRejected=true;}
            if(!moveRejected)throw new IOException("pinned execution ABA sentinel allowed path rename");
            bool replaceRejected=false;try{File.Replace(Path.Combine(directory,"attacker.bin"),Path.Combine(directory,"source.bin"),Path.Combine(directory,"backup.bin"));}catch(IOException){replaceRejected=true;}
            if(!replaceRejected)throw new IOException("pinned execution ABA sentinel allowed swap/replace");
            string digest=VerifyPinned(source);source.Dispose();source=null;root.Dispose();root=null;
            File.Delete(Path.Combine(directory,"attacker.bin"));File.Delete(Path.Combine(directory,"source.bin"));Directory.Delete(directory,false);return digest;}
        catch {if(source!=null)source.Dispose();if(root!=null)root.Dispose();throw;}
    }
    public static string CreateArtifact(SafeFileHandle root, string leaf, byte[] bytes) {
        using(var pinned=CreatePinnedArtifact(root,leaf,bytes))return pinned.Identity+"|"+pinned.Digest;
    }
    public static string ReadArtifact(SafeFileHandle root, string leaf, string identity, string digest, int maxBytes) {
        using(var handle=OpenRelative(root, leaf, FILE_OPEN, false, GENERIC_READ)) {
            var info=Validate(handle,false,true); ulong size=((ulong)info.FileSizeHigh<<32)|info.FileSizeLow;
            if(Format(info)!=identity || size>(ulong)maxBytes) throw new IOException("artifact identity/size changed");
            byte[] bytes=new byte[(int)size]; int offset=0;
            using(var stream=new FileStream(handle,FileAccess.Read)) while(offset<bytes.Length){int count=stream.Read(bytes,offset,bytes.Length-offset);if(count==0)break;offset+=count;}
            if(offset!=bytes.Length) throw new IOException("artifact short read");
            string actual; using(var sha=SHA256.Create()) actual=BitConverter.ToString(sha.ComputeHash(bytes)).Replace("-","").ToLowerInvariant();
            if(actual!=digest) throw new IOException("artifact digest changed");
            return new UTF8Encoding(false,true).GetString(bytes);
        }
    }
    private sealed class Capture { public MemoryStream Bytes=new MemoryStream(); public volatile bool Limit; }
    private static async Task CaptureAsync(Stream stream, Capture capture, int limit) {
        byte[] buffer=new byte[8192];
        while(true){int count=await stream.ReadAsync(buffer,0,buffer.Length).ConfigureAwait(false);if(count==0)break;
            int allowed=Math.Max(0,Math.Min(count,limit-(int)capture.Bytes.Length));if(allowed>0)capture.Bytes.Write(buffer,0,allowed);if(allowed<count)capture.Limit=true;}
    }
    public static HubProcessResult RunBounded(string file, string[] arguments, int timeoutMs, int killGraceMs, int outputLimit, string[] environment) {
        if(String.IsNullOrEmpty(file)||!Path.IsPathRooted(file))throw new IOException("subprocess executable must be an authenticated absolute path");
        var result=new HubProcessResult{ExitCode=125}; var info=new ProcessStartInfo(file){UseShellExecute=false,CreateNoWindow=true,RedirectStandardOutput=true,RedirectStandardError=true};
        info.Environment.Clear();
        foreach(string assignment in environment){
            if(assignment==null)throw new IOException("child environment entry is absent");
            int separator=assignment.IndexOf('=');
            if(separator<=0)throw new IOException("child environment entry is malformed");
            string key=assignment.Substring(0,separator),value=assignment.Substring(separator+1);
            if(key.IndexOf('=')>=0||key.IndexOf('\0')>=0||value.IndexOf('\0')>=0)throw new IOException("child environment entry is unsafe");
            if(info.Environment.ContainsKey(key))throw new IOException("duplicate child environment key");
            info.Environment.Add(key,value);
        }
        foreach(var argument in arguments)info.ArgumentList.Add(argument);
        using(var process=new Process{StartInfo=info}){
            if(!process.Start())throw new IOException("process did not start");
            var stdout=new Capture();var stderr=new Capture();var outTask=CaptureAsync(process.StandardOutput.BaseStream,stdout,outputLimit);var errTask=CaptureAsync(process.StandardError.BaseStream,stderr,outputLimit);
            var clock=Stopwatch.StartNew();bool termination=false;
            while(!process.WaitForExit(100)){if(stdout.Limit||stderr.Limit){result.OutputLimitExceeded=true;termination=true;break;}if(clock.ElapsedMilliseconds>=timeoutMs){result.HostTimedOut=true;termination=true;break;}}
            if(stdout.Limit||stderr.Limit){result.OutputLimitExceeded=true;termination=true;}
            if(termination){try{process.Kill(true);}catch{}if(!process.WaitForExit(killGraceMs)){result.TerminationGraceExceeded=true;result.ExitCode=125;}else result.ExitCode=result.HostTimedOut?124:126;}
            else result.ExitCode=process.ExitCode;
            if(process.HasExited && Task.WaitAll(new[]{outTask,errTask},killGraceMs)){
                result.Stdout=new UTF8Encoding(false,true).GetString(stdout.Bytes.ToArray());result.Stderr=new UTF8Encoding(false,true).GetString(stderr.Bytes.ToArray());
            } else {result.TerminationGraceExceeded=true;result.ExitCode=125;result.Stderr="bounded output drain/termination grace exceeded\n";}
            return result;
        }
    }
}
'@

$script:ReportIdentity = $null
$script:ReportsParentIdentity = $null
$script:ArtifactIdentities = @{}
$script:ArtifactDigests = @{}
$script:ArtifactAttempts = @{}
$script:PinnedFiles = [Collections.ArrayList]::new()
$script:PinnedArtifacts = @{}
$script:ExecutionSnapshot = [ordered]@{}
$script:AuthoritySnapshot = [ordered]@{}
$script:SourceEnvelope = [ordered]@{}
$script:ExternalTools = [ordered]@{}
$script:ClearedInheritedToolOverrides = @()
$script:GitCliPath = $null
$script:DockerCliPath = $null
$script:GitChildEnvironment = @()
$script:DockerChildEnvironment = @()
$script:AbaSentinelDigest = $null
$script:RepoHandle = $null
$script:TestframeworkHandle = $null
$script:ReportsParentHandle = $null
$script:ReportHandle = $null

function Assert-ReportBoundary {
    if ($null -eq $script:ReportHandle -or $script:ReportHandle.IsInvalid -or $script:ReportHandle.IsClosed) { throw 'Report boundary handle is not held.' }
    if ([HubHandleBoundary]::DirectoryIdentity($script:ReportHandle) -ne $script:ReportIdentity) { throw 'Held report boundary identity changed.' }
    if ([HubHandleBoundary]::DirectoryIdentity($script:ReportsParentHandle) -ne $script:ReportsParentIdentity) { throw 'Held reports parent identity changed.' }
    $ParentPathHandle = [HubHandleBoundary]::OpenDirectoryPath($ReportsParent)
    $ReportPathHandle = [HubHandleBoundary]::OpenDirectoryPath($ReportRoot)
    try {
        if ([HubHandleBoundary]::DirectoryIdentity($ParentPathHandle) -ne $script:ReportsParentIdentity) { throw 'Reports parent path no longer names held directory.' }
        if ([HubHandleBoundary]::DirectoryIdentity($ReportPathHandle) -ne $script:ReportIdentity) { throw 'Report path no longer names held directory.' }
    }
    finally { $ReportPathHandle.Dispose(); $ParentPathHandle.Dispose() }
}

function Write-NewTextArtifact {
    param([Parameter(Mandatory=$true)][string]$Path, [Parameter(Mandatory=$true)][string]$Text)
    Assert-ReportBoundary
    if ([IO.Path]::GetFullPath([IO.Path]::GetDirectoryName($Path)) -ne [IO.Path]::GetFullPath($ReportRoot)) {
        throw "Artifact escapes task-run report boundary: $Path"
    }
    $Leaf = [IO.Path]::GetFileName($Path)
    $Seal = [HubHandleBoundary]::CreateArtifact($script:ReportHandle, $Leaf, $Utf8NoBom.GetBytes($Text)).Split('|')
    if ($Seal.Count -ne 2) { throw "Artifact seal malformed: $Path" }
    $script:ArtifactIdentities[$Path] = $Seal[0]
    $script:ArtifactDigests[$Path] = $Seal[1]
}

function Get-BoundedGitScalar {
    param([Parameter(Mandatory=$true)][string[]]$Arguments)
    if (-not $script:GitCliPath) { throw 'Authenticated Git CLI is not pinned.' }
    $Result = [HubHandleBoundary]::RunBounded($script:GitCliPath, $Arguments, 30000, 5000, 4096, $script:GitChildEnvironment)
    if (
        $Result.ExitCode -ne 0 -or
        $Result.HostTimedOut -or
        $Result.OutputLimitExceeded -or
        $Result.TerminationGraceExceeded -or
        $Result.Stderr
    ) {
        throw "Bounded Git scalar command failed: $($Arguments -join ' ')"
    }
    $Value = $Result.Stdout.Trim()
    if (-not $Value -or $Value.Contains("`n") -or $Value.Contains("`r")) {
        throw 'Bounded Git command did not return one scalar.'
    }
    return $Value
}

function Add-PinnedExternalTool {
    param(
        [Parameter(Mandatory=$true)][ValidateSet('git','docker')][string]$Name,
        [Parameter(Mandatory=$true)][string]$AbsolutePath,
        [Parameter(Mandatory=$true)][string]$ExpectedSha256
    )
    if (
        $AbsolutePath.StartsWith('NOT_READY:', [StringComparison]::Ordinal) -or
        -not [IO.Path]::IsPathRooted($AbsolutePath) -or
        [IO.Path]::GetExtension($AbsolutePath) -ine '.exe'
    ) {
        throw "NOT_READY: seal an independently reviewed absolute $Name CLI .exe path."
    }
    if ($ExpectedSha256 -notmatch '^[0-9a-f]{64}$') {
        throw "NOT_READY: seal the independently reviewed $Name CLI SHA-256."
    }
    $Pinned = [HubHandleBoundary]::PinFilePath($AbsolutePath, 268435456)
    if ($Pinned.Digest -ne $ExpectedSha256) {
        $Pinned.Dispose()
        throw "Sealed $Name CLI digest mismatch."
    }
    [void]$script:PinnedFiles.Add($Pinned)
    $script:ExternalTools[$Name] = [ordered]@{
        absolute_path = $AbsolutePath
        identity = $Pinned.Identity
        sha256 = $Pinned.Digest
    }
    return $AbsolutePath
}

function Invoke-AbsoluteToolBoundarySentinel {
    $Rejected = $false
    try {
        [void][HubHandleBoundary]::RunBounded('git', @('--version'), 1000, 1000, 1024, @())
    }
    catch {
        if ($_.Exception.Message -eq 'subprocess executable must be an authenticated absolute path') {
            $Rejected = $true
        }
    }
    if (-not $Rejected) { throw 'Relative/PATH subprocess sentinel was not rejected.' }
}

function Assert-GitReplacementDefenseContract {
    param([Parameter(Mandatory=$true)][string]$InnerScript)
    if ($script:GitChildEnvironment -notcontains 'GIT_NO_REPLACE_OBJECTS=1') {
        throw 'Outer Git child environment does not disable replacement objects.'
    }
    if ($InnerScript -notmatch '(?m)^export GIT_NO_REPLACE_OBJECTS=1$') {
        throw 'Inner guard script does not disable replacement objects.'
    }
}

function Write-NewPinnedTextArtifact {
    param([Parameter(Mandatory=$true)][string]$Path, [Parameter(Mandatory=$true)][string]$Text)
    Assert-ReportBoundary
    if ([IO.Path]::GetFullPath([IO.Path]::GetDirectoryName($Path)) -ne [IO.Path]::GetFullPath($ReportRoot)) {
        throw "Pinned artifact escapes task-run report boundary: $Path"
    }
    $Pinned = [HubHandleBoundary]::CreatePinnedArtifact(
        $script:ReportHandle,
        [IO.Path]::GetFileName($Path),
        $Utf8NoBom.GetBytes($Text)
    )
    [void]$script:PinnedFiles.Add($Pinned)
    $script:PinnedArtifacts[$Path] = $Pinned
    $script:ArtifactIdentities[$Path] = $Pinned.Identity
    $script:ArtifactDigests[$Path] = $Pinned.Digest
}

function Add-PinnedEnvelopeSnapshot {
    param(
        [Parameter(Mandatory=$true)][string]$RelativePath,
        [Parameter(Mandatory=$true)][string]$Role,
        [bool]$Executed = $false,
        [string]$ContainerPath = ''
    )
    Assert-ReportBoundary
    if ($RelativePath -notin $EnvelopeAllowlist) {
        throw "Source envelope path is outside the fixed guard allowlist: $RelativePath"
    }
    if ($ContainerPath -and $ContainerPath -ne "/workspace/$RelativePath") {
        throw "Authority container path is not the canonical read path: $RelativePath"
    }
    $SourcePath = Join-Path $RepoRoot ($RelativePath -replace '/', '\')
    $SourcePin = [HubHandleBoundary]::PinFilePath($SourcePath, 16777216)
    [void]$script:PinnedFiles.Add($SourcePin)
    $TreeSpec = "$SourceRevision`:$RelativePath"
    $ExpectedBlob = Get-BoundedGitScalar @('-C', $RepoRoot, 'rev-parse', '--verify', $TreeSpec)
    if ($ExpectedBlob -notmatch '^[0-9a-f]{40}([0-9a-f]{24})?$') {
        throw "Source envelope path has no canonical committed Git blob: $RelativePath"
    }
    $ActualBlob = [HubHandleBoundary]::GitBlobObjectId($SourcePin.Bytes, $ExpectedBlob.Length)
    if (
        $ActualBlob -ne $ExpectedBlob
    ) {
        throw "Pinned source envelope bytes are not the exact committed Git blob: $RelativePath"
    }
    $Leaf = [IO.Path]::GetFileName($RelativePath)
    $SnapshotPath = Join-Path $ReportRoot $Leaf
    $SnapshotPin = [HubHandleBoundary]::CreatePinnedArtifact($script:ReportHandle, $Leaf, $SourcePin.Bytes)
    if ($SnapshotPin.Digest -ne $SourcePin.Digest) {
        $SnapshotPin.Dispose()
        throw "Source envelope snapshot digest differs from pinned source: $RelativePath"
    }
    [void]$script:PinnedFiles.Add($SnapshotPin)
    $script:PinnedArtifacts[$SnapshotPath] = $SnapshotPin
    $script:ArtifactIdentities[$SnapshotPath] = $SnapshotPin.Identity
    $script:ArtifactDigests[$SnapshotPath] = $SnapshotPin.Digest
    $Record = [ordered]@{
        role = $Role
        source_path = $RelativePath
        git_blob = $ExpectedBlob
        source_identity = $SourcePin.Identity
        sha256 = $SnapshotPin.Digest
        snapshot_path = "$ReportRelative/$Leaf"
        container_path = $(if ($ContainerPath) { $ContainerPath } else { $null })
        consumed_live = [bool]$ContainerPath
    }
    $script:SourceEnvelope[$RelativePath] = $Record
    if ($Executed) { $script:ExecutionSnapshot[$Leaf] = $Record }
    if ($ContainerPath) { $script:AuthoritySnapshot[$RelativePath] = $Record }
    return "$ReportRelative/$Leaf"
}

function Invoke-PinnedAbaSentinel {
    $script:AbaSentinelDigest = [HubHandleBoundary]::RunPinnedAbaSentinel($TempBase, $RunNonce)
}

function Assert-AllPinnedFiles {
    foreach ($Pinned in $script:PinnedFiles) {
        [void][HubHandleBoundary]::VerifyPinned($Pinned)
    }
}

function Set-FirstFailure {
    param([int]$Code)
    if ($Code -ne 0 -and $null -eq $script:FirstExit) {
        $script:FirstExit = $Code
    }
}

function Get-ArtifactAttempt {
    param([Parameter(Mandatory=$true)][string]$Key)
    $Attempt = 1 + [int]($script:ArtifactAttempts[$Key])
    $script:ArtifactAttempts[$Key] = $Attempt
    return $Attempt
}

function Invoke-DockerProcess {
    param(
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [Parameter(Mandatory = $true)][int]$TimeoutSeconds
    )
    Assert-ReportBoundary
    if (-not $script:DockerCliPath) { throw 'Authenticated Docker CLI is not pinned.' }
    $Result = [HubHandleBoundary]::RunBounded($script:DockerCliPath, $Arguments, $TimeoutSeconds * 1000, 10000, 16777216, $script:DockerChildEnvironment)
    $Stderr = $Result.Stderr
    if ($Result.OutputLimitExceeded) { $Stderr += "OUTPUT_LIMIT_EXCEEDED`n" }
    if ($Result.TerminationGraceExceeded) { $Stderr += "TERMINATION_GRACE_EXCEEDED`n" }
    Write-NewTextArtifact -Path $StdoutPath -Text $Result.Stdout
    Write-NewTextArtifact -Path $StderrPath -Text $Stderr
    return $Result.ExitCode
}

function Get-ArtifactSha256 {
    param([Parameter(Mandatory = $true)][string]$Path)
    Assert-ReportBoundary
    if (-not $script:ArtifactIdentities.ContainsKey($Path) -or -not $script:ArtifactDigests.ContainsKey($Path)) { throw "Artifact was not created by this task run: $Path" }
    [void][HubHandleBoundary]::ReadArtifact($script:ReportHandle, [IO.Path]::GetFileName($Path), $script:ArtifactIdentities[$Path], $script:ArtifactDigests[$Path], 16777216)
    return $script:ArtifactDigests[$Path]
}

function Get-OptionalArtifactSha256 {
    param([Parameter(Mandatory = $true)][string]$Path)
    Assert-ReportBoundary
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return Get-ArtifactSha256 $Path
}

function Read-TaskTextArtifact {
    param([Parameter(Mandatory = $true)][string]$Path)
    Assert-ReportBoundary
    if (-not $script:ArtifactIdentities.ContainsKey($Path)) { throw "Artifact was not created by this task run: $Path" }
    return [HubHandleBoundary]::ReadArtifact($script:ReportHandle, [IO.Path]::GetFileName($Path), $script:ArtifactIdentities[$Path], $script:ArtifactDigests[$Path], 16777216)
}

function Remove-TaskOwnedContainer {
    param([Parameter(Mandatory = $true)][string]$Name)
    $Attempt = Get-ArtifactAttempt "remove-$Name"
    $InspectOut = Join-Path $ReportRoot "$Name.remove-$Attempt.inspect.stdout.log"
    $InspectErr = Join-Path $ReportRoot "$Name.remove-$Attempt.inspect.stderr.log"
    $InspectExit = Invoke-DockerProcess `
        -Arguments @('container', 'inspect', '--format', '{{.Id}}|{{ index .Config.Labels "hub.task-run" }}|{{ index .Config.Labels "hub.owner" }}', $Name) `
        -StdoutPath $InspectOut `
        -StderrPath $InspectErr `
        -TimeoutSeconds 30
    if ($InspectExit -eq 0) {
        $Identity = @((Read-TaskTextArtifact $InspectOut).Trim() -split '\|', 3)
        if (
            $Identity.Count -ne 3 -or
            $Identity[0] -notmatch '^[0-9a-f]{64}$' -or
            $Identity[1] -ne $TaskRun -or
            $Identity[2] -ne 'phase128'
        ) {
            throw "Refusing cleanup of non-owned container: $Name"
        }
        $ImmutableContainerId = $Identity[0]
        $RemoveExit = Invoke-DockerProcess `
            -Arguments @('rm', '--force', $ImmutableContainerId) `
            -StdoutPath (Join-Path $ReportRoot "$Name.remove-$Attempt.rm.stdout.log") `
            -StderrPath (Join-Path $ReportRoot "$Name.remove-$Attempt.rm.stderr.log") `
            -TimeoutSeconds 30
        if ($RemoveExit -ne 0) {
            throw "Exact ID cleanup failed: $Name id=$ImmutableContainerId ($RemoveExit)"
        }
    }
    elseif ($InspectExit -ne 1) {
        throw "Container ownership inspection failed: $Name ($InspectExit)"
    }
    else {
        $InspectMessage = Read-TaskTextArtifact $InspectErr
        if ($InspectMessage -notmatch 'No such (object|container)') {
            throw "Container was not proven absent: $Name"
        }
        $DaemonExit = Invoke-DockerProcess `
            -Arguments @('info', '--format', '{{.ID}}') `
            -StdoutPath (Join-Path $ReportRoot "$Name.remove-$Attempt.daemon.stdout.log") `
            -StderrPath (Join-Path $ReportRoot "$Name.remove-$Attempt.daemon.stderr.log") `
            -TimeoutSeconds 30
        if ($DaemonExit -ne 0) {
            throw "Docker daemon liveness was not proven before treating container as absent: $Name"
        }
    }
}

function Assert-ContainerAbsent {
    param([Parameter(Mandatory = $true)][string]$Name)
    $Attempt = Get-ArtifactAttempt "absent-$Name"
    $Out = Join-Path $ReportRoot "$Name.absent-$Attempt.stdout.log"
    $Err = Join-Path $ReportRoot "$Name.absent-$Attempt.stderr.log"
    $Exit = Invoke-DockerProcess `
        -Arguments @('container', 'inspect', $Name) `
        -StdoutPath $Out `
        -StderrPath $Err `
        -TimeoutSeconds 30
    if ($Exit -eq 0) {
        throw "Task-owned container still exists: $Name"
    }
    if ($Exit -ne 1) {
        throw "Container absence inspection failed: $Name ($Exit)"
    }
    $InspectMessage = Read-TaskTextArtifact $Err
    if ($InspectMessage -notmatch 'No such (object|container)') {
        throw "Container was not proven absent: $Name"
    }
    $DaemonExit = Invoke-DockerProcess `
        -Arguments @('info', '--format', '{{.ID}}') `
        -StdoutPath (Join-Path $ReportRoot "$Name.absent-$Attempt.daemon.stdout.log") `
        -StderrPath (Join-Path $ReportRoot "$Name.absent-$Attempt.daemon.stderr.log") `
        -TimeoutSeconds 30
    if ($DaemonExit -ne 0) {
        throw "Docker daemon liveness was not proven after absence check: $Name"
    }
}

function Remove-TaskRunLabeledContainers {
    $ListOut = Join-Path $ReportRoot 'task-run-label-remove-list.stdout.log'
    $ListErr = Join-Path $ReportRoot 'task-run-label-remove-list.stderr.log'
    $ListExit = Invoke-DockerProcess `
        -Arguments @('ps', '-aq', '--filter', "label=hub.task-run=$TaskRun", '--filter', 'label=hub.owner=phase128') `
        -StdoutPath $ListOut `
        -StderrPath $ListErr `
        -TimeoutSeconds 30
    if ($ListExit -ne 0) { throw "Task-run label cleanup enumeration failed: $ListExit" }
    $RawIds = (Read-TaskTextArtifact $ListOut).Trim()
    $Ids = @($RawIds -split '\s+' | Where-Object { $_ })
    if ($Ids.Count -gt 16) { throw 'Task-run label cleanup exceeds 16-container budget.' }
    foreach ($Id in $Ids) {
        if ($Id -notmatch '^[0-9a-f]{12,64}$') { throw "Invalid labeled container ID: $Id" }
        $InspectOut = Join-Path $ReportRoot "label-$Id.inspect.stdout.log"
        $InspectErr = Join-Path $ReportRoot "label-$Id.inspect.stderr.log"
        $InspectExit = Invoke-DockerProcess `
            -Arguments @('container', 'inspect', '--format', '{{ index .Config.Labels "hub.task-run" }}|{{ index .Config.Labels "hub.owner" }}', $Id) `
            -StdoutPath $InspectOut `
            -StderrPath $InspectErr `
            -TimeoutSeconds 30
        if ($InspectExit -ne 0 -or (Read-TaskTextArtifact $InspectOut).Trim() -ne "$TaskRun|phase128") {
            throw "Labeled container ownership changed before removal: $Id"
        }
        $RemoveExit = Invoke-DockerProcess `
            -Arguments @('rm', '--force', $Id) `
            -StdoutPath (Join-Path $ReportRoot "label-$Id.rm.stdout.log") `
            -StderrPath (Join-Path $ReportRoot "label-$Id.rm.stderr.log") `
            -TimeoutSeconds 30
        if ($RemoveExit -ne 0) { throw "Labeled container removal failed: $Id ($RemoveExit)" }
    }
}

function Assert-NoTaskRunLabeledContainers {
    $Out = Join-Path $ReportRoot 'task-run-label-scan.stdout.log'
    $Err = Join-Path $ReportRoot 'task-run-label-scan.stderr.log'
    $Exit = Invoke-DockerProcess `
        -Arguments @('ps', '-aq', '--filter', "label=hub.task-run=$TaskRun") `
        -StdoutPath $Out `
        -StderrPath $Err `
        -TimeoutSeconds 30
    if ($Exit -ne 0) { throw "Task-run label enumeration failed: $Exit" }
    $Remaining = (Read-TaskTextArtifact $Out).Trim()
    if ($Remaining) { throw "Task-run labeled containers remain after exact cleanup: $Remaining" }
}

if ($ExpectedImageDigest -notmatch '^sha256:[0-9a-f]{64}$') {
    throw 'NOT_READY: seal the independently reviewed Phase 128 image digest in validate-contract.ps1.'
}
$WindowsRoot = [Environment]::GetFolderPath([Environment+SpecialFolder]::Windows)
if (-not [IO.Path]::IsPathRooted($WindowsRoot) -or -not [IO.Directory]::Exists($WindowsRoot)) {
    throw 'Trusted Windows root could not be established for the minimal child environment.'
}
$script:ClearedInheritedToolOverrides = @(
    [Environment]::GetEnvironmentVariables().Keys |
        ForEach-Object { [string]$_ } |
        Where-Object { $_ -match '^(?i:GIT_|DOCKER_)' } |
        Sort-Object -Unique
)
$script:GitChildEnvironment = @(
    "SystemRoot=$WindowsRoot",
    "WINDIR=$WindowsRoot",
    "TEMP=$ReportRoot",
    "TMP=$ReportRoot",
    "HOME=$ReportRoot",
    "USERPROFILE=$ReportRoot",
    'LANG=C',
    'LC_ALL=C',
    'GIT_CONFIG_NOSYSTEM=1',
    'GIT_CONFIG_GLOBAL=NUL',
    'GIT_CONFIG_COUNT=0',
    'GIT_OPTIONAL_LOCKS=0',
    'GIT_NO_REPLACE_OBJECTS=1',
    'GIT_TERMINAL_PROMPT=0'
)
$script:DockerChildEnvironment = @(
    "SystemRoot=$WindowsRoot",
    "WINDIR=$WindowsRoot",
    "TEMP=$ReportRoot",
    "TMP=$ReportRoot",
    "HOME=$ReportRoot",
    "USERPROFILE=$ReportRoot",
    "DOCKER_CONFIG=$ReportRoot",
    'DOCKER_HOST=npipe:////./pipe/docker_engine',
    'DOCKER_CLI_HINTS=false'
)
Invoke-AbsoluteToolBoundarySentinel
$script:GitCliPath = Add-PinnedExternalTool 'git' $SealedGitCliPath $SealedGitCliSha256
$script:DockerCliPath = Add-PinnedExternalTool 'docker' $SealedDockerCliPath $SealedDockerCliSha256
$ReportsParent = Join-Path $RepoRoot 'testframework\reports'
$script:RepoHandle = [HubHandleBoundary]::OpenDirectoryPath($RepoRoot)
$script:TestframeworkHandle = [HubHandleBoundary]::OpenDirectoryRelative($script:RepoHandle, 'testframework')
$script:ReportsParentHandle = [HubHandleBoundary]::OpenOrCreateDirectoryRelative($script:TestframeworkHandle, 'reports')
$script:ReportsParentIdentity = [HubHandleBoundary]::DirectoryIdentity($script:ReportsParentHandle)
$script:ReportHandle = [HubHandleBoundary]::CreateDirectoryRelative($script:ReportsParentHandle, $TaskRun)
$script:ReportIdentity = [HubHandleBoundary]::DirectoryIdentity($script:ReportHandle)
Assert-ReportBoundary

$SourceRevision = Get-BoundedGitScalar @('-C', $RepoRoot, 'rev-parse', '--verify', 'HEAD')
if ($SourceRevision -notmatch '^[0-9a-f]{40}([0-9a-f]{24})?$') {
    throw 'Live source revision is not a canonical Git object ID.'
}
$ExecInventory = Add-PinnedEnvelopeSnapshot $InventoryValidator 'python-executable' $true
$ExecValidator = Add-PinnedEnvelopeSnapshot $Validator 'python-executable' $true
$ExecBuilder = Add-PinnedEnvelopeSnapshot $Builder 'python-executable' $true
$ExecCutoverValidator = Add-PinnedEnvelopeSnapshot $CutoverValidator 'python-executable-and-import' $true
$ExecBatchApplier = Add-PinnedEnvelopeSnapshot $BatchApplier 'python-executable' $true
$WrapperSnapshot = Add-PinnedEnvelopeSnapshot '.planning/audits/test-rebase/validate-contract.ps1' 'host-wrapper-provenance' $true
$PinnedInventoryAuthority = Add-PinnedEnvelopeSnapshot $InventoryAuthority 'authority-ledger' $false "/workspace/$InventoryAuthority"
$PinnedInventorySchema = Add-PinnedEnvelopeSnapshot $InventorySchema 'authority-schema' $false "/workspace/$InventorySchema"
$PinnedFounderManifest = Add-PinnedEnvelopeSnapshot $FounderManifest 'authority-founder-manifest' $false "/workspace/$FounderManifest"
$PinnedArchitectureContract = Add-PinnedEnvelopeSnapshot $ArchitectureContract 'authority-architecture-contract' $false "/workspace/$ArchitectureContract"
$PinnedArchitectureAdr = Add-PinnedEnvelopeSnapshot $ArchitectureAdr 'authority-adr' $false "/workspace/$ArchitectureAdr"
$SourceRevisionAfterPin = Get-BoundedGitScalar @('-C', $RepoRoot, 'rev-parse', '--verify', 'HEAD')
if ($SourceRevisionAfterPin -ne $SourceRevision) {
    throw 'Live HEAD changed while the committed Phase 128 source envelope was pinned.'
}
Invoke-PinnedAbaSentinel
Assert-AllPinnedFiles
$OwnerMarker = [ordered]@{
    owner = 'phase128'
    source_revision = $SourceRevision
    task_run = $TaskRun
}
$OwnerMarkerPath = Join-Path $ReportRoot '.owner.json'
Write-NewTextArtifact -Path $OwnerMarkerPath -Text (($OwnerMarker | ConvertTo-Json -Depth 4) + "`n")
$ImageInspectOut = Join-Path $ReportRoot 'image-inspect.stdout.log'
$ImageInspectErr = Join-Path $ReportRoot 'image-inspect.stderr.log'
$ImageInspectExit = Invoke-DockerProcess `
    -Arguments @('image', 'inspect', '--format', '{{.Id}}', $ImageRef) `
    -StdoutPath $ImageInspectOut `
    -StderrPath $ImageInspectErr `
    -TimeoutSeconds 30
if ($ImageInspectExit -ne 0) {
    throw "Pinned local image inspection failed: $ImageRef ($ImageInspectExit)"
}
$ImageDigest = (Read-TaskTextArtifact $ImageInspectOut).Trim()
if ($ImageDigest -notmatch '^sha256:[0-9a-f]{64}$') {
    throw "Inspected image ID is not immutable SHA-256: $ImageDigest"
}
if ($ImageDigest -ne $ExpectedImageDigest) {
    throw "Sealed image digest mismatch: expected=$ExpectedImageDigest actual=$ImageDigest"
}
$InspectedAt = [DateTimeOffset]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')
$ImageLock = [ordered]@{
    actual_image_digest = $ImageDigest
    expected_image_digest = $ExpectedImageDigest
    image_ref = $ImageRef
    inspected_at = $InspectedAt
    source_revision = $SourceRevision
}
$ImageLockPath = Join-Path $ReportRoot 'image-lock.json'
Write-NewTextArtifact -Path $ImageLockPath -Text (($ImageLock | ConvertTo-Json -Depth 4) + "`n")

$DockerChecks = @"
#!/bin/sh
set -eu
export GIT_NO_REPLACE_OBJECTS=1
echo 'PROVENANCE: guard SHA-256'
sha256sum $ExecInventory $ExecValidator $ExecBuilder $ExecCutoverValidator $ExecBatchApplier $WrapperSnapshot $PinnedInventoryAuthority $PinnedInventorySchema $PinnedFounderManifest $PinnedArchitectureContract $PinnedArchitectureAdr "$ReportRelative/run.sh"
cmp $PinnedInventoryAuthority $InventoryAuthority
cmp $PinnedInventorySchema $InventorySchema
cmp $PinnedFounderManifest $FounderManifest
cmp $PinnedArchitectureContract $ArchitectureContract
cmp $PinnedArchitectureAdr $ArchitectureAdr
printf 'SOURCE_REVISION=%s\n' '$SourceRevision'
started=`$(date +%s)
set +e
timeout --signal=TERM --kill-after=10s 300s sh -lc '
set -eu
replacement_repo=/tmp/git-no-replace-sentinel
git init -q "`$replacement_repo"
original_object=`$(printf "original-byte\n" | git -C "`$replacement_repo" hash-object -w --stdin)
replacement_object=`$(printf "replacement-byte\n" | git -C "`$replacement_repo" hash-object -w --stdin)
mkdir -p "`$replacement_repo/.git/refs/replace"
printf "%s\n" "`$replacement_object" > "`$replacement_repo/.git/refs/replace/`$original_object"
test "`$(env -u GIT_NO_REPLACE_OBJECTS git -C "`$replacement_repo" cat-file -p "`$original_object")" = replacement-byte
test "`$(git -C "`$replacement_repo" cat-file -p "`$original_object")" = original-byte
rm -rf -- "`$replacement_repo"
git config --global --add safe.directory /workspace
test "`$(git rev-parse --verify HEAD)" = "$SourceRevision"
python3 -I -B $ExecInventory --check
python3 -I -B $ExecInventory --self-test
python3 -I -B $ExecValidator --manifest $FounderManifest
python3 -I -B $ExecValidator --manifest $FounderManifest --self-test --docker-bind-sentinel /docker-bind-sentinel
python3 -I -B $ExecBuilder --dry-run > /tmp/cutover-a.json
python3 -I -B $ExecBuilder --dry-run > /tmp/cutover-b.json
cmp /tmp/cutover-a.json /tmp/cutover-b.json
python3 -I -B $ExecBuilder --self-test
test ! -e .planning/audits/test-rebase/cutover-manifest.json
python3 -I -B $ExecCutoverValidator --stdin < /tmp/cutover-a.json
python3 -I -B $ExecCutoverValidator --stdin --self-test < /tmp/cutover-a.json
python3 -I -B $ExecBatchApplier --self-test
test ! -e .planning/audits/test-rebase/cutover-manifest.json
'
status=`$?
set -e
ended=`$(date +%s)
printf 'TELEMETRY: inner_seconds=%s timeout_seconds=300 kill_grace_seconds=10 exit=%s\n' "`$((ended-started))" "`$status"
exit "`$status"
"@
Assert-GitReplacementDefenseContract $DockerChecks
$RunScriptPath = Join-Path $ReportRoot 'run.sh'
Write-NewPinnedTextArtifact -Path $RunScriptPath -Text $DockerChecks.Replace("`r`n", "`n")

$CommonLimits = @(
    '--pull', 'never',
    '--label', "hub.task-run=$TaskRun",
    '--label', 'hub.owner=phase128',
    '--network', 'none',
    '--read-only',
    '--security-opt', 'no-new-privileges',
    '--cap-drop', 'ALL'
)
$CreateArgs = @(
    'run', '--rm', '--name', $SentinelCreateContainer,
    '--stop-timeout', '10', '--cpus', '0.25', '--memory', '128m',
    '--memory-swap', '128m', '--pids-limit', '32',
    '--mount', "type=bind,source=$TempBase,target=/hosttmp"
) + $CommonLimits + @(
    $ExpectedImageDigest, 'sh', '-lc',
    "set -eu; umask 077; mkdir -p '/hosttmp/$SentinelLeaf/real'; printf 'package sentinel\n' > '/hosttmp/$SentinelLeaf/real/source_test.go'; ln -s real '/hosttmp/$SentinelLeaf/linked'"
)
$TimeoutArgs = @(
    'run', '--rm', '--name', $TimeoutSentinelContainer,
    '--stop-timeout', '1', '--cpus', '0.25', '--memory', '128m',
    '--memory-swap', '128m', '--pids-limit', '32'
) + $CommonLimits + @(
    $ExpectedImageDigest, 'sh', '-lc',
    'sleep 30'
)
$MainArgs = @(
    'run', '--rm', '--name', $ContainerName,
    '--stop-timeout', '10', '--cpus', '1.0', '--memory', '1g',
    '--memory-swap', '1g', '--pids-limit', '128',
    '--tmpfs', '/tmp:rw,noexec,nosuid,size=64m',
    '--env', 'HOME=/tmp', '--env', 'PYTHONDONTWRITEBYTECODE=1',
    '--env', 'HUB_PHASE128_REPO_ROOT=/workspace',
    '--mount', "type=bind,source=$RepoRoot,target=/workspace,readonly",
    '--mount', "type=bind,source=$(Join-Path $ReportRoot ([IO.Path]::GetFileName($InventoryAuthority))),target=/workspace/$InventoryAuthority,readonly",
    '--mount', "type=bind,source=$(Join-Path $ReportRoot ([IO.Path]::GetFileName($InventorySchema))),target=/workspace/$InventorySchema,readonly",
    '--mount', "type=bind,source=$(Join-Path $ReportRoot ([IO.Path]::GetFileName($FounderManifest))),target=/workspace/$FounderManifest,readonly",
    '--mount', "type=bind,source=$(Join-Path $ReportRoot ([IO.Path]::GetFileName($ArchitectureContract))),target=/workspace/$ArchitectureContract,readonly",
    '--mount', "type=bind,source=$(Join-Path $ReportRoot ([IO.Path]::GetFileName($ArchitectureAdr))),target=/workspace/$ArchitectureAdr,readonly",
    '--mount', "type=bind,source=$SentinelHostRoot,target=/docker-bind-sentinel,readonly",
    '--workdir', '/workspace'
) + $CommonLimits + @(
    $ExpectedImageDigest, 'sh', "/workspace/$ReportRelative/run.sh"
)
$CleanupArgs = @(
    'run', '--rm', '--name', $SentinelCleanupContainer,
    '--stop-timeout', '10', '--cpus', '0.25', '--memory', '128m',
    '--memory-swap', '128m', '--pids-limit', '32',
    '--mount', "type=bind,source=$TempBase,target=/hosttmp"
) + $CommonLimits + @(
    $ExpectedImageDigest, 'sh', '-lc',
    "set -eu; rm -rf -- '/hosttmp/$SentinelLeaf'; test ! -e '/hosttmp/$SentinelLeaf'"
)
$ExecutionContract = [ordered]@{
    schema_version = 1
    task_run = $TaskRun
    source_revision = $SourceRevision
    expected_image_digest = $ExpectedImageDigest
    external_tools = $script:ExternalTools
    cleared_inherited_tool_overrides = $script:ClearedInheritedToolOverrides
    git_child_environment = $script:GitChildEnvironment
    docker_child_environment = $script:DockerChildEnvironment
    run_script_sha256 = Get-ArtifactSha256 $RunScriptPath
    executed_files = $script:ExecutionSnapshot
    authority_inputs = $script:AuthoritySnapshot
    source_envelope = $script:SourceEnvelope
    aba_sentinel_sha256 = $script:AbaSentinelDigest
    sentinel_create_arguments = $CreateArgs
    timeout_sentinel_arguments = $TimeoutArgs
    main_arguments = $MainArgs
    cleanup_arguments = $CleanupArgs
    host_timeout_seconds = 330
    host_kill_grace_seconds = 10
}
$ExecutionContractPath = Join-Path $ReportRoot 'execution-contract.json'
Write-NewTextArtifact -Path $ExecutionContractPath -Text (($ExecutionContract | ConvertTo-Json -Depth 8) + "`n")

try {
    $CreateExit = Invoke-DockerProcess `
        -Arguments $CreateArgs `
        -StdoutPath (Join-Path $ReportRoot 'sentinel-create.stdout.log') `
        -StderrPath (Join-Path $ReportRoot 'sentinel-create.stderr.log') `
        -TimeoutSeconds 60
    Set-FirstFailure $CreateExit
    if ($CreateExit -ne 0) { throw "Docker-bind sentinel creation failed: $CreateExit" }

    $TimeoutExit = Invoke-DockerProcess `
        -Arguments $TimeoutArgs `
        -StdoutPath (Join-Path $ReportRoot 'timeout-sentinel.stdout.log') `
        -StderrPath (Join-Path $ReportRoot 'timeout-sentinel.stderr.log') `
        -TimeoutSeconds 2
    if ($TimeoutExit -ne 124) {
        Set-FirstFailure $TimeoutExit
        throw "Forced timeout sentinel returned $TimeoutExit instead of 124"
    }
    Remove-TaskOwnedContainer $TimeoutSentinelContainer
    Assert-ContainerAbsent $TimeoutSentinelContainer

    $MainStarted = [DateTimeOffset]::UtcNow
    Assert-AllPinnedFiles
    $MainExit = Invoke-DockerProcess `
        -Arguments $MainArgs `
        -StdoutPath (Join-Path $ReportRoot 'architecture-contract.stdout.log') `
        -StderrPath (Join-Path $ReportRoot 'architecture-contract.stderr.log') `
        -TimeoutSeconds 330
    $MainDurationMs = [int64]([DateTimeOffset]::UtcNow - $MainStarted).TotalMilliseconds
    Assert-AllPinnedFiles
    Set-FirstFailure $MainExit
}
catch {
    if ($null -eq $script:FirstExit) { $script:FirstExit = 1 }
    Write-NewTextArtifact -Path (Join-Path $ReportRoot 'wrapper-error.log') -Text ($_.Exception.Message + "`n")
}
finally {
    try { Assert-AllPinnedFiles }
    catch { if ($null -eq $script:FirstExit) { $script:FirstExit = 1 } }
    foreach ($OwnedName in ($OwnedContainers | Where-Object { $_ -ne $SentinelCleanupContainer })) {
        try {
            Remove-TaskOwnedContainer $OwnedName
        }
        catch {
            if ($null -eq $script:FirstExit) { $script:FirstExit = 1 }
        }
    }
    try {
        $CleanupExit = Invoke-DockerProcess `
            -Arguments $CleanupArgs `
            -StdoutPath (Join-Path $ReportRoot 'sentinel-cleanup.stdout.log') `
            -StderrPath (Join-Path $ReportRoot 'sentinel-cleanup.stderr.log') `
            -TimeoutSeconds 60
        if ($CleanupExit -ne 0 -and $null -eq $script:FirstExit) {
            $script:FirstExit = $CleanupExit
        }
        Remove-TaskOwnedContainer $SentinelCleanupContainer
        Remove-TaskRunLabeledContainers
        foreach ($OwnedName in $OwnedContainers) {
            Assert-ContainerAbsent $OwnedName
        }
        Assert-NoTaskRunLabeledContainers
        $AllResourcesAbsent = $true
    }
    catch {
        if ($null -eq $script:FirstExit) { $script:FirstExit = 1 }
    }
}

$Stopwatch.Stop()
if ($null -eq $script:FirstExit) { $script:FirstExit = 0 }
foreach ($RequiredLog in @('architecture-contract.stdout.log', 'architecture-contract.stderr.log')) {
    $RequiredLogPath = Join-Path $ReportRoot $RequiredLog
    if (-not (Test-Path -LiteralPath $RequiredLogPath -PathType Leaf)) {
        Write-NewTextArtifact -Path $RequiredLogPath -Text "NOT_RUN`n"
    }
}
$CleanupProof = [ordered]@{
    all_resources_absent = $AllResourcesAbsent
    actual_image_digest = $ImageDigest
    expected_image_digest = $ExpectedImageDigest
    exit_code = $script:FirstExit
    inspected_at = [DateTimeOffset]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')
    label_scan_sha256 = Get-OptionalArtifactSha256 (Join-Path $ReportRoot 'task-run-label-scan.stdout.log')
    resources = $OwnedContainers
    labels_absent = $AllResourcesAbsent
    source_revision = $SourceRevision
    task_run = $TaskRun
}
$CleanupProofPath = Join-Path $ReportRoot 'cleanup-proof.json'
Write-NewTextArtifact -Path $CleanupProofPath -Text (($CleanupProof | ConvertTo-Json -Depth 5) + "`n")
$Evidence = [ordered]@{
    schema_version = 1
    task_run = $TaskRun
    owner = 'phase128'
    source_revision = $SourceRevision
    image_ref = $ImageRef
    expected_image_digest = $ExpectedImageDigest
    actual_image_digest = $ImageDigest
    external_tools = $script:ExternalTools
    cleared_inherited_tool_overrides = $script:ClearedInheritedToolOverrides
    git_child_environment = $script:GitChildEnvironment
    docker_child_environment = $script:DockerChildEnvironment
    image_lock_sha256 = Get-ArtifactSha256 $ImageLockPath
    owner_marker_sha256 = Get-ArtifactSha256 $OwnerMarkerPath
    execution_contract_path = "$ReportRelative/execution-contract.json"
    execution_contract_sha256 = Get-ArtifactSha256 $ExecutionContractPath
    executed_files = $script:ExecutionSnapshot
    authority_inputs = $script:AuthoritySnapshot
    source_envelope = $script:SourceEnvelope
    executed_bytes_aba_sentinel_sha256 = $script:AbaSentinelDigest
    run_script_path = "$ReportRelative/run.sh"
    run_script_sha256 = Get-ArtifactSha256 $RunScriptPath
    stdout_path = "$ReportRelative/architecture-contract.stdout.log"
    stdout_sha256 = Get-ArtifactSha256 (Join-Path $ReportRoot 'architecture-contract.stdout.log')
    stderr_path = "$ReportRelative/architecture-contract.stderr.log"
    stderr_sha256 = Get-ArtifactSha256 (Join-Path $ReportRoot 'architecture-contract.stderr.log')
    duration_ms = $MainDurationMs
    inner_timeout_seconds = 300
    host_timeout_seconds = 330
    cpu = '1.0'
    memory = '1g'
    memory_swap = '1g'
    pids_limit = 128
    network = 'none'
    read_only = $true
    timeout_sentinel_exit = $TimeoutExit
    cleanup_proof_sha256 = Get-ArtifactSha256 $CleanupProofPath
    exit_code = $script:FirstExit
    runner_contract = 'phase128-guard-only-not-hub-qa-runner-accepted'
    boundary_proof_path = "$ReportRelative/boundary-proof.json"
}
$EvidencePath = Join-Path $ReportRoot 'evidence.json'
Write-NewTextArtifact -Path $EvidencePath -Text (($Evidence | ConvertTo-Json -Depth 6) + "`n")
$ArtifactSeals = [ordered]@{}
foreach ($ArtifactPath in ($script:ArtifactIdentities.Keys | Sort-Object)) {
    $Digest = Get-ArtifactSha256 $ArtifactPath
    $ArtifactSeals[[IO.Path]::GetFileName($ArtifactPath)] = [ordered]@{
        identity = $script:ArtifactIdentities[$ArtifactPath]
        sha256 = $Digest
    }
}
$BoundaryProof = [ordered]@{
    schema_version = 1
    task_run = $TaskRun
    source_revision = $SourceRevision
    reports_parent_identity = $script:ReportsParentIdentity
    report_identity = $script:ReportIdentity
    evidence_sha256 = Get-ArtifactSha256 $EvidencePath
    artifacts = $ArtifactSeals
    proven_at = [DateTimeOffset]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')
}
$BoundaryProofPath = Join-Path $ReportRoot 'boundary-proof.json'
Write-NewTextArtifact -Path $BoundaryProofPath -Text (($BoundaryProof | ConvertTo-Json -Depth 8) + "`n")
[void](Get-ArtifactSha256 $BoundaryProofPath)
Assert-ReportBoundary
Assert-AllPinnedFiles
Write-Host "EVIDENCE: $ReportRelative/evidence.json boundary=$ReportRelative/boundary-proof.json source=$SourceRevision image=$ImageDigest exit=$script:FirstExit"
for ($PinnedIndex = $script:PinnedFiles.Count - 1; $PinnedIndex -ge 0; $PinnedIndex--) {
    $script:PinnedFiles[$PinnedIndex].Dispose()
}
$script:ReportHandle.Dispose()
$script:ReportsParentHandle.Dispose()
$script:TestframeworkHandle.Dispose()
$script:RepoHandle.Dispose()
exit $script:FirstExit
