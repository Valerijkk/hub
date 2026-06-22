# Migrate backend/internal/<InternalPath>/*_test.go → backend/testframework/<TestframeworkPath>/
# Usage: .\scripts\migrate-nested-go-tests.ps1 -InternalPath integrations/onec -TestframeworkPath onec
param(
    [Parameter(Mandatory = $true)]
    [string]$InternalPath,
    [Parameter(Mandatory = $true)]
    [string]$TestframeworkPath,
    [string]$PackageName = "",
    [string]$ImportAlias = "",
    [switch]$DryRun,
    [switch]$DeleteSource,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$srcDir = Join-Path $repoRoot "backend/internal/$InternalPath"
$dstDir = Join-Path $repoRoot "backend/testframework/$TestframeworkPath"
$importPath = "github.com/valerii/hub/backend/internal/$InternalPath"

if ($PackageName -eq "") {
    $PackageName = ($InternalPath -split '/')[-1]
}

if ($ImportAlias -eq "") {
    $ImportAlias = $PackageName
}

if (-not (Test-Path $srcDir)) {
    Write-Error "Source not found: $srcDir"
}

$skipFiles = @("export_test.go", "export_hooks.go")
$testFiles = Get-ChildItem -Path $srcDir -Filter "*_test.go" | Where-Object { $skipFiles -notcontains $_.Name }

if ($testFiles.Count -eq 0) {
    Write-Host "No test files in $srcDir"
    exit 0
}

if (-not $DryRun -and -not (Test-Path $dstDir)) {
    New-Item -ItemType Directory -Path $dstDir -Force | Out-Null
}

function Get-NestedExtraSymbols {
    param([string]$Pkg)
    $map = @{
        onec = @(
            "NewService", "NewAdminHandler", "Service", "Config", "SettingsResponse", "UpdateSettingsRequest",
            "TestWebhookResponse", "LeaveWebhookPayload", "EmployeePayload", "EmployeeIDField",
            "EmployeeIDEmail", "EmployeeIDHubUUID", "EmployeeIDPhone",
            "ErrInvalidConfig", "ErrNotConfigured", "ErrWebhookFailed",
            "NewEntSettingsStore", "NewEntUserLookup"
        )
        config = @("Load")
        cron = @("Schedule")
        crypto = @("NewCipherFromKey", "NewCipherFromString", "DeriveDevKey", "Cipher")
        errors = @(
            "AppError", "NewBadRequest", "NewUnauthorized", "NewForbidden", "NewNotFound",
            "NewConflict", "NewRateLimited", "NewRateLimitedAfter", "NewInternal",
            "MsgBadRequest", "MsgUnauthorized", "MsgForbidden", "MsgNotFound", "MsgInternal", "MsgConflict"
        )
        http = @(
            "UserIDFromCtx", "ParseUUIDParam", "ParseUUIDQuery", "NewSafeHTTPClient",
            "ValidateOutboundURL", "FetchOpenGraphMetadata", "OpenGraphMeta",
            "ErrBlockedScheme", "ErrInvalidURL"
        )
        middleware = @(
            "RequestID", "Recover", "Logger", "CORS", "RateLimit", "RateLimitByUser",
            "RateLimitRedis", "RedisRateLimitConfig", "RateLimitConfig", "SecurityHeaders",
            "ErrorHandler", "ExtractEmailFromBody", "RequireAuth"
        )
        nats = @("NewClient", "Client", "StreamName", "Publish", "PublishJSON", "EnsureStreams", "Subscribe")
        observability = @(
            "ResetMetrics", "Snapshot", "Timed", "TimedVoid", "InitPrometheus", "PrometheusHandler",
            "MetricsAllowMiddleware", "InitOTel", "Shutdown"
        )
        outbox = @("NewWorker", "NewPublisher", "Worker", "Publisher")
        telegram = @("EscapeMarkdownV2")
    }
    if ($map.ContainsKey($Pkg)) { return $map[$Pkg] }
    return @()
}

function Apply-NestedPrivateTransforms {
    param([string]$Content, [string]$Alias, [string]$Pkg)

    switch ($Pkg) {
        "onec" {
            $Content = $Content -replace '\bsvc\.client\s*=', "${Alias}.ExportServiceSetClient(svc, "
            $Content = $Content -replace '\bsvc\.client\s*=\s*&http\.Client\{', "${Alias}.ExportServiceSetClient(svc, &http.Client{"
            $Content = $Content -replace '\bsvc\.post\(', "${Alias}.ExportServicePost(svc, "
            $Content = $Content -replace '\bdecodeConfig\(', "${Alias}.ExportDecodeConfig("
            $Content = $Content -replace '\bencodeConfig\(', "${Alias}.ExportEncodeConfig("
            $Content = $Content -replace '\bsettingsKey\b', "${Alias}.ExportSettingsKey()"
            $Content = $Content -replace '\bbuildLeavePayload\(', "${Alias}.ExportBuildLeavePayload("
            $Content = $Content -replace 'Config\{([^}]*)\}\.validate\(\)', '${Alias}.ExportConfigValidate(Config{$1})'
            $Content = $Content -replace '\bConfig\{Enabled: false\}\.validate\(\)', "${Alias}.ExportConfigValidate(${Alias}.Config{Enabled: false})"
            $Content = $Content -replace '\bConfig\{Enabled: true, WebhookURL: ""\}\.validate\(\)', "${Alias}.ExportConfigValidate(${Alias}.Config{Enabled: true, WebhookURL: `"`"})"
            $Content = $Content -replace '\bConfig\{Enabled: true, WebhookURL: "https://x.com", EmployeeIDField: "bad"\}\.validate\(\)', "${Alias}.ExportConfigValidate(${Alias}.Config{Enabled: true, WebhookURL: `"https://x.com`", EmployeeIDField: `"bad`"})"
        }
        "observability" {
            $Content = $Content -replace '\brecordOperation\(', "${Alias}.ExportRecordOperation("
            $Content = $Content -replace '\bparseCIDRs\(', "${Alias}.ExportParseCIDRs("
            $Content = $Content -replace '\bipAllowed\(', "${Alias}.ExportIpAllowed("
        }
        "outbox" {
            $Content = $Content -replace '\bw\.client\b', "${Alias}.ExportWorkerClient(w)"
            $Content = $Content -replace '\bw\.nats\b', "${Alias}.ExportWorkerNats(w)"
            $Content = $Content -replace '\bw\.processEvent\(', "${Alias}.ExportWorkerProcessEvent(w, "
            $Content = $Content -replace '\bw\.processBatch\(', "${Alias}.ExportWorkerProcessBatch(w, "
        }
        "http" {
            $Content = $Content -replace '\bparseOpenGraphHTML\(', "${Alias}.ExportParseOpenGraphHTML("
        }
        "middleware" {
            $Content = $Content -replace '\bparseCORSOrigins\(', "${Alias}.ExportParseCORSOrigins("
            $Content = $Content -replace '\bstripQuery\(', "${Alias}.ExportStripQuery("
            $Content = $Content -replace '\bisDevMode\(\)', "${Alias}.ExportIsDevMode()"
            $Content = $Content -replace '\bnormalizeEmail\(', "${Alias}.ExportNormalizeEmail("
            $Content = $Content -replace '\bhashIP\(', "${Alias}.ExportHashIP("
            $Content = $Content -replace '&rateLimiter\{config:\s*RateLimitConfig\{', "${Alias}.ExportNewRateLimiterForTest(${Alias}.RateLimitConfig{"
            $Content = $Content -replace '\brl\.allow\(', "${Alias}.ExportRateLimiterAllow(rl, "
            $Content = $Content -replace '\brl\.cleanup\(', "${Alias}.ExportRateLimiterCleanup(rl, "
        }
    }
    return $Content
}

function Transform-NestedTestFile {
    param([string]$Content, [string]$Pkg, [string]$Alias, [string]$ImpPath)

    # Preserve build tags and existing _test package
    if ($Content -match '(?m)^package (\w+)_test\s*$') {
        # already external test package — fix import only below
    } elseif ($Content -match "(?m)^package $Pkg\s*$") {
        $Content = $Content -replace "(?m)^package $Pkg\s*$", "package ${Pkg}_test"
    }

    $Content = $Content -replace '\bfiber\.Handler\b', '__FIBER_HANDLER__'
    $Content = $Content -replace '\bredis\.Client\b', '__REDIS_CLIENT__'
    $Content = $Content -replace '\bredis\.NewClient\b', '__REDIS_NEWCLIENT__'
    $Content = $Content -replace '\bnatsclient\.Client\b', '__NATSCLIENT_CLIENT__'
    $Content = $Content -replace '\bnatsclient\.NewClient\b', '__NATSCLIENT_NEWCLIENT__'

    $Content = Apply-NestedPrivateTransforms -Content $Content -Alias $Alias -Pkg $Pkg

    $publicSymbols = @(
        "NewHandler", "NewWorker", "NewPublisher", "Handler", "Worker", "Publisher",
        "ErrNotFound", "ErrInvalidConfig", "ErrNotConfigured", "ErrWebhookFailed"
    ) + (Get-NestedExtraSymbols -Pkg $Pkg)

    foreach ($sym in $publicSymbols) {
        $escaped = [regex]::Escape($sym)
        $pattern = "(?<!${Alias}\.)\b$escaped\b"
        $Content = [System.Text.RegularExpressions.Regex]::Replace($Content, $pattern, "${Alias}.$sym")
    }

    $Content = $Content -replace '__FIBER_HANDLER__', 'fiber.Handler'
    $Content = $Content -replace '__REDIS_CLIENT__', 'redis.Client'
    $Content = $Content -replace '__REDIS_NEWCLIENT__', 'redis.NewClient'
    $Content = $Content -replace '__NATSCLIENT_CLIENT__', 'natsclient.Client'
    $Content = $Content -replace '__NATSCLIENT_NEWCLIENT__', 'natsclient.NewClient'

    if ($Content -match "\b${Alias}\." -and $Content -notmatch [regex]::Escape($ImpPath)) {
        $importLine = if ($Alias -ne $Pkg) {
            "`t${Alias} `"$ImpPath`"`r`n"
        } else {
            "`t`"$ImpPath`"`r`n"
        }
        if ($Content -match '(?ms)(import \(\r?\n)') {
            if ($Content -notmatch [regex]::Escape($ImpPath)) {
                $Content = $Content -replace '(?ms)(import \(\r?\n)', "`${1}$importLine"
            }
        } elseif ($Content -match '(?m)^import "') {
            $Content = $Content -replace '(?m)^import "([^"]+)"', "import (`r`n$importLine`t`"`$1`")"
        } else {
            $Content = $Content -replace '(?m)^(package \w+_test\s*\r?\n)', "`${1}`r`nimport (`r`n$importLine)`r`n"
        }
    }

    return $Content
}

$migrated = 0
foreach ($file in $testFiles) {
    $dstPath = Join-Path $dstDir $file.Name
    if ((Test-Path $dstPath) -and -not $Force) {
        Write-Host "SKIP (exists): $($file.Name)"
        continue
    }

    $raw = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $transformed = Transform-NestedTestFile -Content $raw -Pkg $PackageName -Alias $ImportAlias -ImpPath $importPath

    if ($DryRun) {
        Write-Host "DRY-RUN: $($file.Name) -> $dstPath"
    } else {
        Set-Content -Path $dstPath -Value $transformed -Encoding UTF8 -NoNewline
        Write-Host "Migrated: $($file.Name)"
        if ($DeleteSource) {
            Remove-Item -Path $file.FullName -Force
            Write-Host "  Deleted: $($file.Name)"
        }
    }
    $migrated++
}

Write-Host "Done. Migrated $migrated file(s) to testframework/$TestframeworkPath"
