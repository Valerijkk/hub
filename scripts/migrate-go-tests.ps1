# Migrate backend/internal/<module>/*_test.go → backend/testframework/<module>/
# Usage: .\scripts\migrate-go-tests.ps1 -Module notification [-DryRun] [-DeleteSource]
param(
    [Parameter(Mandatory = $true)]
    [string]$Module,
    [switch]$DryRun,
    [switch]$DeleteSource,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$srcDir = Join-Path $repoRoot "backend/internal/$Module"
$dstDir = Join-Path $repoRoot "backend/testframework/$Module"
$importPath = "github.com/valerii/hub/backend/internal/$Module"

if (-not (Test-Path $srcDir)) {
    Write-Error "Source module not found: $srcDir"
}

$skipFiles = @("export_test.go", "export_hooks.go")
$testFiles = Get-ChildItem -Path $srcDir -Filter "*_test.go" | Where-Object { $skipFiles -notcontains $_.Name }

if ($testFiles.Count -eq 0) {
    Write-Host "No test files to migrate in $srcDir"
    exit 0
}

if (-not $DryRun -and -not (Test-Path $dstDir)) {
    New-Item -ItemType Directory -Path $dstDir -Force | Out-Null
}

function Get-ModuleExtraSymbols {
    param([string]$Mod)
    $map = @{
        auth = @(
            "AuthService", "MagicLinkService", "LoginService", "SessionService", "ProfileService",
            "OrganizationService", "RegistrationModeService", "TelegramBotAuth",
            "MagicLinkRequest", "VerifyMagicLinkRequest", "TelegramAuthRequest", "RefreshRequest",
            "UpdateProfileRequest", "OnboardingCompleteRequest", "TokenPairResponse", "UserResponse",
            "SessionResponse", "PendingUserResponse", "OrgInviteResponse", "SetRegistrationModeRequest",
            "NewAuthService", "NewMagicLinkService", "NewLoginService", "NewSessionService",
            "NewProfileService", "NewOrganizationService", "NewRegistrationModeService",
            "NewTelegramBotAuth", "NewEntRepository", "NewResendEmailClient", "VerifyTelegramAuth",
            "RequireAuth",
            "UserStatusCacheKey", "InvalidateUserStatusCache",
            "ErrMagicLinkInvalid", "ErrMagicLinkExpired", "ErrMagicLinkRateLimited",
            "ErrMagicLinkVerifyLimited", "ErrUserNotFound", "ErrTelegramVerifyLimited", "ErrTelegramAuthExpired",
            "ErrInvalidInviteToken", "ErrRegistrationInviteOnly",
            "IsNotFound", "IsSystemAdmin", "UserToResponse", "VerifyTelegramAuth",
            "Repository", "EntRepository", "JWTClaims", "TelegramUser", "ResendEmailClient",
            "ResendEmailClient", "SessionCreatedEvent", "UserRegisteredEvent"
        )
        ws = @(
            "Hub", "Client", "Bridge", "NewHub", "NewClient", "NewBridge",
            "ServerMessage", "ClientMessage", "TypePong", "TypeSubscribe", "TypeUnsubscribe",
            "TypeMessage", "TypeError", "TypeEvent", "MessageType", "Authorizer", "RoomAuthorizer", "AuthorizerConfig",
            "NewAuthorizer", "Handler"
        )
        rbac = @(
            "RBACService", "NewRBACService", "NewHandler", "Handler", "NewCustomRoleHandler", "CustomRoleHandler",
            "Cache", "NewCache", "PermissionCache", "NewPermissionCache",
            "Permission", "PermTaskCreate", "PermTaskRead", "PermProjectAdmin", "PermProjectView",
            "PermProjectEdit", "PermProjectDelete", "PermTeamManage", "PermTaskEdit", "PermFileUpload",
            "PermTestCaseCreate", "PermTestRunExecute", "PermProjectManageMembers",
            "CustomRoleRequest", "RequirePermission", "RequireAuth", "RequireAdmin", "UserIDFromCtx",
            "ValidPermissionName", "ListAllPermissionDefs", "PermissionDef",
            "DefaultRolePermissions", "ErrNotFound", "NewEntRepository", "EntRepository", "Repository"
        )
        audit = @(
            "Service", "NewService", "NewHandler", "Handler", "AuditConsumer", "NewAuditConsumer",
            "AuditService", "NewAuditService", "Repository", "NewRepository", "AlertSeverity", "AuditEvent",
            "RetentionPolicy", "AuditLogFilters", "TrashItem", "TrashItemResponse",
            "NewRetentionService", "RetentionService", "AlertmanagerWebhookPayload", "AlertmanagerAlert",
            "AuditLogListResponse", "LimitFileMaxSizeMB", "LimitUserMaxProjects", "LimitProjectMaxStorageMB",
            "HTTPAuditMiddleware", "HTTPAuditEntry"
        )
        tracker = @("TrackerService", "NewTrackerService", "NewHandler", "Handler", "Repository", "NewEntRepository")
        tms = @("TMSService", "NewTMSService", "NewHandler", "Handler", "Repository", "NewEntRepository", "NewCombinedRepository")
        project = @("ProjectService", "NewProjectService", "NewHandler", "Handler", "TemplateService", "NewTemplateService", "TemplateHandler", "NewTemplateHandler")
        messenger = @("MessengerService", "NewMessengerService", "NewHandler", "Handler", "Repository", "NewEntRepository", "NewMongoRepository", "NewCombinedRepository")
        dev = @(
            "Handler", "NewHandler", "DevLoginRequest", "DevUserInfo",
            "BuildAPIDocsHTML", "BuildAPIDocsJSON",
            "ExportDemoUsers", "ExportDemoStatuses", "ExportDemoIssues", "ExportDemoTestCases", "ExportDemoChannels",
            "ExportMethodOrder", "ExportMethodColor", "ExportGuessModule", "ExportCollectRoutes", "ExportRouteEntry"
        )
        status = @(
            "Service", "NewService", "Handler", "NewHandler", "Pinger",
            "Response", "ComponentStatus"
        )
        mail = @(
            "Config", "LoadConfig", "Service", "NewService",
            "HubHandler", "NewHubHandler", "AdminService", "NewAdminService",
            "AdminHandler", "NewAdminHandler", "EntSettingsStore", "NewEntSettingsStore",
            "InternalTokenAuth", "SendRequest", "StatusResponse", "SettingsResponse",
            "UpdateSettingsRequest", "BroadcastRequest", "BroadcastResponse",
            "BroadcastSender", "ErrInvalidRequest", "ExportInternalTokenHeader", "ExportGetEnv"
        )
        calendar = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "EventType", "EventTypeCall", "EventTypeSprint", "EventTypeDeadline", "EventTypeLeave",
            "EventResponse", "EventsResponse", "RenderICalendar",
            "ErrInvalidRange", "ErrRangeTooWide",
            "ExportExpandCallEvents", "ExportParseRRule", "ExportLeaveEvent", "ExportValidateRange"
        )
        marketplace = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "ListingResponse", "ListingListResponse", "CreateListingRequest", "UpdateListingRequest",
            "ReviewListingRequest", "VersionRequest", "ListingCategory", "ListingStatus",
            "CategoryEquipment", "CategoryBooks", "CategoryServices", "CategoryFurniture", "CategoryOther",
            "StatusPending", "StatusApproved", "StatusRejected", "StatusSold", "StatusWithdrawn",
            "ErrNotFound", "ErrForbidden", "ErrInvalidInput", "ErrVersionConflict", "ErrNotOwner", "ErrInvalidStatus",
            "ExportMapServiceError", "ExportDisplayName", "ExportFormatMinorUnits", "ExportItoa", "ExportPad2",
            "ExportIsValidCategory", "ExportFormatPrice"
        )
        accounting = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "AccountResponse", "JournalEntryResponse", "JournalLineInput", "JournalLineResponse",
            "TrialBalanceRow", "InvoiceResponse", "IncomeStatementResponse",
            "CreateAccountRequest", "CreateJournalRequest", "CreateInvoiceRequest",
            "AccountType", "EntryStatus", "InvoiceStatus",
            "AccountAsset", "AccountLiability", "AccountEquity", "AccountRevenue", "AccountExpense",
            "StatusDraft", "StatusPosted",
            "InvoiceDraft", "InvoiceIssued", "InvoicePaid", "InvoiceCancelled",
            "ErrNotFound", "ErrForbidden", "ErrInvalidInput", "ErrUnbalanced", "ErrAlreadyPosted",
            "ErrInvoiceExists", "ErrInvalidInvoice",
            "ExportMapServiceError", "ExportParseEntryDate"
        )
        bases = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "BaseResponse", "BaseDetailResponse", "TableSummary", "TableViewResponse",
            "ColumnResponse", "RecordResponse",
            "CreateBaseRequest", "UpdateBaseRequest", "CreateTableRequest", "UpdateTableRequest",
            "CreateColumnRequest", "UpdateColumnRequest", "CreateRecordRequest", "UpdateRecordRequest",
            "ColumnType", "ColumnText", "ColumnNumber", "ColumnDate", "ColumnSelect",
            "ColumnCheckbox", "ColumnURL", "ColumnUserRef",
            "ErrNotFound", "ErrInvalidInput", "ErrVersionConflict",
            "ExportMapServiceError", "ExportNormalizeCells"
        )
    }
    if ($map.ContainsKey($Mod)) { return $map[$Mod] }
    return @()
}

function Transform-TestFile {
    param([string]$Content, [string]$Mod)

    $pkg = "${Mod}_test"
    $Content = $Content -replace "(?m)^package $Mod\s*$", "package $pkg"

    # repo.db → ExportRepositoryDB
    $Content = $Content -replace '\b(\w+)\.db\.', "${Mod}.ExportRepositoryDB(`$1)."

    # Consumer handlers
    $handlers = @(
        "handleAssigned", "handleStatusChanged", "handleCommentAdded", "handleBlockerAdded",
        "handleMessageReceived", "handleCallStarted", "handleRunCompleted", "handleNewLogin"
    )
    foreach ($h in $handlers) {
        $export = "Export" + ($h.Substring(0,1).ToUpper() + $h.Substring(1))
        if ($h -eq "handleNewLogin") {
            $Content = $Content -replace "\bc\.$h\(", "${Mod}.$export(c, "
        } else {
            $Content = $Content -replace "\bc\.$h\(", "${Mod}.$export(c, "
        }
    }

    # processEvent, shouldDispatch, cleanDebounce
    $Content = $Content -replace '\bc\.processEvent\(', "${Mod}.ExportProcessEvent(c, "
    $Content = $Content -replace '\bc\.shouldDispatch\(', "${Mod}.ExportShouldDispatch(c, "
    $Content = $Content -replace '\bc\.cleanDebounce\(\)', "${Mod}.ExportCleanDebounce(c)"
    $Content = $Content -replace '\bc\.debounce\[', "# DEBOUNCE_MANUAL: c.debounce["
    $Content = $Content -replace '\bassert\.Empty\(t, c\.debounce\)', "assert.Equal(t, 0, ${Mod}.ExportDebounceLen(c))"

    # buildPushURL, isRunFailedRecipient
    $Content = $Content -replace '\bbuildPushURL\(', "${Mod}.ExportBuildPushURL("
    $Content = $Content -replace '\bisRunFailedRecipient\(', "${Mod}.ExportIsRunFailedRecipient("

    # processEvent (audit/notification consumer)
    $Content = $Content -replace '\bconsumer\.processEvent\(', "${Mod}.ExportProcessEvent(consumer, "

    # Telegram bind
    $Content = $Content -replace '\bsvc\.repo\.db\.', "${Mod}.ExportRepositoryDB(${Mod}.ExportTelegramBindRepository(svc))."
    $Content = $Content -replace '\bsvc\.handleMessage\(', "${Mod}.ExportTelegramHandleMessage(svc, "
    $Content = $Content -replace '\bsvc\.handleUnlinkCommand\(', "${Mod}.ExportTelegramHandleUnlinkCommand(svc, "
    $Content = $Content -replace '\bsvc\.handleStartBind\(', "${Mod}.ExportTelegramHandleStartBind(svc, "
    $Content = $Content -replace '\bsvc\.handleLinkCommand\(', "${Mod}.ExportTelegramHandleLinkCommand(svc, "
    $Content = $Content -replace '\bsvc\.handleTestCommand\(', "${Mod}.ExportTelegramHandleTestCommand(svc, "

    # ws hub/bridge private methods
    $Content = $Content -replace '\bhub\.registerClient\(', "${Mod}.ExportRegisterClient(hub, "
    $Content = $Content -replace '\bhub\.removeClient\(', "${Mod}.ExportRemoveClient(hub, "
    $Content = $Content -replace '\bhub\.connectionsForUser\(', "${Mod}.ExportConnectionsForUser(hub, "
    $Content = $Content -replace '\bbridge\.handleProjectEvent\(', "${Mod}.ExportBridgeHandleProjectEvent(bridge, "
    $Content = $Content -replace '\bbridge\.handleMessengerEvent\(', "${Mod}.ExportBridgeHandleMessengerEvent(bridge, "
    $Content = $Content -replace '\bbridge\.handleCallEvent\(', "${Mod}.ExportBridgeHandleCallEvent(bridge, "
    $Content = $Content -replace '\bbridge\.handleCallChatEvent\(', "${Mod}.ExportBridgeHandleCallChatEvent(bridge, "
    $Content = $Content -replace '\bbridge\.handleAuthSessionEvent\(', "${Mod}.ExportBridgeHandleAuthSessionEvent(bridge, "
    $Content = $Content -replace '\bbridge\.handleNotificationEvent\(', "${Mod}.ExportBridgeHandleNotificationEvent(bridge, "

    # Preserve fiber.Handler before symbol prefixing
    $Content = $Content -replace '\bfiber\.Handler\b', '__FIBER_HANDLER__'

    # Telegram bot (auth module)
    $Content = $Content -replace '\btgAuth\.handleLogin\(', "${Mod}.ExportTelegramHandleLogin(tgAuth, "

    # Auth private constants and service fields
    if ($Mod -eq "auth") {
        $Content = $Content -replace '\bmagicLinkVerifyMax\b', "${Mod}.ExportMagicLinkVerifyMax()"
        $Content = $Content -replace '\bmagicLinkRateMax\b', "${Mod}.ExportMagicLinkRateMax()"
        $Content = $Content -replace '\bmagicLinkTTL\b', "${Mod}.ExportMagicLinkTTL()"
        $Content = $Content -replace '\bmagicLinkFailsPrefix\b', "${Mod}.ExportMagicLinkFailsPrefix()"
        $Content = $Content -replace '\buserStatusCacheDeactivated\b', "${Mod}.ExportUserStatusCacheDeactivated()"
        $Content = $Content -replace '\buserStatusCacheTTL\b', "${Mod}.ExportUserStatusCacheTTL()"
        $Content = $Content -replace '\bsvc\.magicLink\b', "${Mod}.ExportAuthServiceMagicLink(svc)"
        $Content = $Content -replace '\bsvc\.redis\b', "${Mod}.ExportAuthServiceRedis(svc)"
        $Content = $Content -replace '\bc\.apiKey\b', "${Mod}.ExportResendEmailClientAPIKey(c)"
        $Content = $Content -replace '\bc\.fromEmail\b', "${Mod}.ExportResendEmailClientFromEmail(c)"
        $Content = $Content -replace '\btgCodePrefix\b', "${Mod}.ExportTGCodePrefix()"
        $Content = $Content -replace '\btgFailsPrefix\b', "${Mod}.ExportTGFailsPrefix()"
        $Content = $Content -replace '\bgenerateTGCode\(\)', "${Mod}.ExportGenerateTGCode()"
        $Content = $Content -replace '\bbot\.recordVerifyFailure\(', "${Mod}.ExportTelegramRecordVerifyFailure(bot, "
        $Content = $Content -replace '\btgAuth\.handleLogin\(', "${Mod}.ExportTelegramHandleLogin(tgAuth, "
        $Content = $Content -replace '&TelegramBotAuth\{bot:\s*bot,\s*redis:\s*rdb\}', "${Mod}.ExportNewTelegramBotAuthForTestWithBot(bot, rdb)"
        $Content = $Content -replace '&TelegramBotAuth\{redis:\s*rdb\}', "${Mod}.ExportNewTelegramBotAuthForTest(rdb)"
        $Content = [System.Text.RegularExpressions.Regex]::Replace(
            $Content,
            '(?ms)(\w+)\.client = (&http\.Client\{.*?\n\t\})',
            "${Mod}.ExportResendEmailClientInjectHTTP(`$1, `$2)"
        )
        $Content = $Content.Replace('&Handler{service: &AuthService{}}', "${Mod}.NewHandler(&${Mod}.AuthService{})")
        $Content = $Content -replace '\bhashRefreshToken\b', "${Mod}.ExportHashRefreshToken"
        $Content = $Content -replace '\brefreshRedisKeyStored\b', "${Mod}.ExportRefreshRedisKeyStored"
        $Content = $Content -replace '\brefreshRedisKey\b', "${Mod}.ExportRefreshRedisKey"
        $Content = $Content -replace '\brefreshPayload\b', "${Mod}.ExportRefreshPayload"
        $Content = $Content -replace '\bsvc\.storeRefreshToken\(', "${Mod}.ExportSessionStoreRefreshToken(svc, "
        $Content = $Content -replace '\bsvc\.invalidateUserSessions\(', "${Mod}.ExportSessionInvalidateUserSessions(svc, "
        $Content = $Content -replace '\bsvc\.publishSessionCreated\(', "${Mod}.ExportSessionPublishSessionCreated(svc, "
        $Content = $Content -replace '\bsvc\.nats\s*=\s*(\w+)', "${Mod}.ExportSessionSetNATS(svc, `$1)"
        $Content = $Content -replace '\bgenerateCode\(\)', "${Mod}.ExportGenerateCode()"
        $Content = $Content -replace '\bsvc\.generateAccessToken\(', "${Mod}.ExportSessionGenerateAccessToken(svc, "
        $Content = $Content -replace '\bsvc1\.generateAccessToken\(', "${Mod}.ExportSessionGenerateAccessToken(svc1, "
        $Content = [System.Text.RegularExpressions.Regex]::Replace(
            $Content,
            '(?ms)&SessionService\{\s*jwtSecret:\s*\[\]byte\("([^"]+)"\),\s*accessTTL:\s*([^,\}]+),?\s*\}',
            { param($m) "${Mod}.ExportNewSessionServiceForJWTTest(""$($m.Groups[1].Value)"", $($m.Groups[2].Value))" }
        )
    }

    if ($Mod -eq "rbac") {
        $Content = $Content -replace '__F_SVC_CACHE__', "${Mod}.ExportRBACServiceCache(f.svc)"
        $Content = $Content -replace '\bf\.svc\.cache\b', '__F_SVC_CACHE__'
        $Content = $Content -replace '\bsvc\.cache\b', "${Mod}.ExportRBACServiceCache(svc)"
        $Content = $Content -replace '\bcacheKey\(', "${Mod}.ExportCacheKey("
        $Content = $Content -replace '\bvalidatePermissionNames\(', "${Mod}.ExportValidatePermissionNames("
        $Content = $Content -replace '\bpermissionsFromNames\(', "${Mod}.ExportPermissionsFromNames("
        $Content = $Content -replace '\bpermissionNames\(', "${Mod}.ExportPermissionNames("
        $Content = $Content -replace '\bpermissionRegistry\b', "${Mod}.ExportPermissionRegistry()"
        $Content = $Content -replace '\bbuildPermissionNameMap\(\)', "${Mod}.ExportBuildPermissionNameMap()"
    }

    if ($Mod -eq "audit") {
        $Content = [System.Text.RegularExpressions.Regex]::Replace(
            $Content,
            '&AuditConsumer\{repo:\s*(\w+)\}',
            { param($m) "${Mod}.ExportNewAuditConsumerForTest($($m.Groups[1].Value))" }
        )
        $Content = $Content -replace '&AuditConsumer\{\}', "${Mod}.ExportNewAuditConsumerForTest(nil)"
        $Content = $Content -replace '\bconsumer\.mapSubjectToAction\(', "${Mod}.ExportConsumerMapSubjectToAction(consumer, "
        $Content = $Content -replace '\bextractStringField\(', "${Mod}.ExportExtractStringField("
        $Content = $Content -replace '\bslowRequestThresholdMS\b', "${Mod}.ExportSlowRequestThresholdMS()"
        $Content = $Content -replace '\bsvc\.purgeOldAllureResults\(', "${Mod}.ExportRetentionServicePurgeOldAllureResults(svc, "
        $Content = $Content -replace '\bdaysRemaining\(', "${Mod}.ExportDaysRemaining("
    }

    if ($Mod -eq "marketplace") {
        $Content = $Content -replace '\bmapServiceError\(', "${Mod}.ExportMapServiceError("
        $Content = $Content -replace '\bdisplayName\(', "${Mod}.ExportDisplayName("
        $Content = $Content -replace '\bformatMinorUnits\(', "${Mod}.ExportFormatMinorUnits("
        $Content = $Content -replace '\bitoa\(', "${Mod}.ExportItoa("
        $Content = $Content -replace '\bpad2\(', "${Mod}.ExportPad2("
        $Content = $Content -replace '\bisValidCategory\(', "${Mod}.ExportIsValidCategory("
        $Content = $Content -replace '\bformatPrice\(', "${Mod}.ExportFormatPrice("
    }

    if ($Mod -eq "accounting") {
        $Content = $Content -replace '\bmapServiceError\(', "${Mod}.ExportMapServiceError("
        $Content = $Content -replace '\bparseEntryDate\(', "${Mod}.ExportParseEntryDate("
    }

    if ($Mod -eq "bases") {
        $Content = $Content -replace '\bmapServiceError\(', "${Mod}.ExportMapServiceError("
        $Content = $Content -replace '\bnormalizeCells\(', "${Mod}.ExportNormalizeCells("
    }

    if ($Mod -eq "ws") {
        $Content = $Content -replace '\bsplitRoom\(', "${Mod}.ExportSplitRoom("
        $Content = $Content -replace '\ballowAllAuthorizer\{\}', "${Mod}.ExportAllowAllAuthorizer()"
        $Content = $Content -replace '\breadPump\(', "${Mod}.ExportReadPump("
        $Content = $Content -replace '\bclient\.closeOnce\.Do\(func\(\) \{ close\(client\.done\) \}\)', "${Mod}.ExportClientCloseDone(client)"
        $Content = [System.Text.RegularExpressions.Regex]::Replace(
            $Content,
            '\b(\w+)\.send\b',
            "${Mod}.ExportClientSendCh(`$1)"
        )
    }

    # Handler unexported service field (before symbol prefixing — still &Handler{service:})
    $Content = $Content -replace 'return &Handler\{service:\s*([a-zA-Z_][a-zA-Z0-9_]*)\}', "return ${Mod}.NewHandler(`$1)"
    $Content = $Content -replace '&Handler\{service:\s*([a-zA-Z_][a-zA-Z0-9_]*)\}', "${Mod}.NewHandler(`$1)"
    $Content = $Content -replace '\bgot\.service\b', "${Mod}.ExportHandlerService(got)"

    # Preserve ent.IsNotFound from auth symbol prefixing
    $Content = $Content -replace '\bent\.IsNotFound\b', '__ENT_ISNOTFOUND__'
    # Preserve nats platform client from ws Client prefixing
    $Content = $Content -replace '\bnatsclient\.Client\b', '__NATSCLIENT_CLIENT__'
    $Content = $Content -replace '\bnatsclient\.NewClient\b', '__NATSCLIENT_NEWCLIENT__'
    $Content = $Content -replace '\bredis\.Client\b', '__REDIS_CLIENT__'
    $Content = $Content -replace '\bredis\.NewClient\b', '__REDIS_NEWCLIENT__'

    # Constructors and types (word boundary, avoid double-prefix)
    $publicSymbols = @(
        "NewConsumer", "NewHandler", "NewRepository", "NewNotificationService", "NewTelegramBindService",
        "NewTelegramSender", "NotificationService", "Repository", "Consumer", "Handler",
        "DispatchEvent", "TypeAssigned", "TypeStatusChanged", "TypeCommentAdded", "TypeBlockerAdded",
        "TypeMessageReceived", "TypeCallStarted", "TypeRunCompleted", "TypeNewLogin", "TypeRunFailed",
        "ErrNotFound", "TelegramBindService", "TelegramSender"
    ) + (Get-ModuleExtraSymbols -Mod $Mod)

    $regexOpts = [System.Text.RegularExpressions.RegexOptions]::None
    foreach ($sym in $publicSymbols) {
        $escaped = [regex]::Escape($sym)
        $pattern = "(?<!${Mod}\.)\b$escaped\b"
        $Content = [System.Text.RegularExpressions.Regex]::Replace($Content, $pattern, "${Mod}.$sym", $regexOpts)
    }

    $Content = $Content -replace '__FIBER_HANDLER__', 'fiber.Handler'
    $Content = $Content -replace '__ENT_ISNOTFOUND__', 'ent.IsNotFound'
    $Content = $Content -replace '__NATSCLIENT_CLIENT__', 'natsclient.Client'
    $Content = $Content -replace '__NATSCLIENT_NEWCLIENT__', 'natsclient.NewClient'
    $Content = $Content -replace '__REDIS_CLIENT__', 'redis.Client'
    $Content = $Content -replace '__REDIS_NEWCLIENT__', 'redis.NewClient'

    # Fix ent notification package conflict
    $Content = $Content -replace "${Mod}\.notification\.", "entnotif."

    # Ensure module import (always when auth. symbols present)
    if ($Content -match "\b${Mod}\." -and $Content -notmatch [regex]::Escape($importPath)) {
        $importBlock = "`t`"$importPath`"`r`n"
        if ($Content -match '(?ms)(import \(\r?\n)') {
            $Content = $Content -replace '(?ms)(import \(\r?\n)', "`${1}$importBlock"
        } elseif ($Content -match '(?m)^import "') {
            $Content = $Content -replace '(?m)^import "', "import (`r`n`t`"$importPath`",`r`n`t"
            $Content = $Content -replace '(?m)^import \(`r`n`t"([^"]+)"\)', "import (`r`n`t`"$importPath`",`r`n`t`"`$1`")"
        }
    }

    # Add entnotif alias if ent/notification imported
    if ($Content -match 'ent/notification' -and $Content -notmatch 'entnotif ') {
        $Content = $Content -replace '"github.com/valerii/hub/backend/ent/notification"', 'entnotif "github.com/valerii/hub/backend/ent/notification"'
    }

    return $Content
}

$migrated = 0
foreach ($file in $testFiles) {
    $dstPath = Join-Path $dstDir $file.Name
    if ((Test-Path $dstPath) -and $file.Name -ne "consumer_extra_test.go" -and -not $Force) {
        Write-Host "SKIP (exists): $($file.Name)"
        continue
    }

    $raw = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $transformed = Transform-TestFile -Content $raw -Mod $Module

    if ($DryRun) {
        Write-Host "DRY-RUN would migrate: $($file.Name) -> $dstPath"
    } else {
        Set-Content -Path $dstPath -Value $transformed -Encoding UTF8 -NoNewline
        Write-Host "Migrated: $($file.Name)"
        if ($DeleteSource) {
            Remove-Item -Path $file.FullName -Force
            Write-Host "  Deleted source: $($file.Name)"
        }
    }
    $migrated++
}

Write-Host "Done. Migrated $migrated file(s). Verify with Docker:"
Write-Host "  go test -tags hubtest ./testframework/$Module/... ./internal/$Module/..."
