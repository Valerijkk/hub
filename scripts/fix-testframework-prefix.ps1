# Re-apply module symbol prefixing to backend/testframework/<module>/*_test.go
param(
    [Parameter(Mandatory = $true)]
    [string]$Module
)

$ErrorActionPreference = "Stop"

function Get-ModuleExtraSymbols {
    param([string]$Mod)
    $map = @{
        tracker = @(
            "TrackerService", "NewTrackerService", "NewHandler", "Handler", "Repository", "NewEntRepository", "EntRepository",
            "IssueResponse", "IssueBriefResponse", "CreateIssueInput", "UpdateIssueInput", "ListIssuesParams",
            "CreateCommentInput", "UpdateCommentInput", "CreateCustomFieldInput",
            "CreateAutomationRuleInput", "UpdateAutomationRuleInput", "AutomationService", "NewAutomationService",
            "AutomationEngine", "NewAutomationEngine", "AutomationConsumer", "NewAutomationConsumer", "WorkflowEngine", "WIPChecker",
            "CreateIssueRequest", "UpdateIssueRequest", "MoveIssueRequest", "ListIssuesRequest",
            "CreateCommentRequest", "UpdateCommentRequest", "CreateLabelRequest", "CreateBoardRequest",
            "AutomationRuleResponse", "CreateAutomationRuleRequest", "UpdateAutomationRuleRequest",
            "CommentResponse", "BoardResponse", "PaginatedResponse", "ErrNotFound", "ErrConflict",
            "IssueRepository", "CommentRepository", "WorkflowRepository", "AutomationRepository",
            "LabelResponse", "TimeLogResponse", "ActivityLogResponse", "CustomFieldDefinitionResponse",
            "IssueLinkResponse", "WorkflowStatusResponse", "WorkflowTransitionResponse", "IssueCreatedEvent", "IssueMovedEvent", "IssueAssignedEvent", "IssueUpdatedEvent", "IssueDeletedEvent",
            "TriggerIssueCreated", "TriggerStatusChanged", "TriggerIssueAssigned", "ActionPostToChannel", "ActionNotifyUser",
            "IssueMessageLinkResponse", "CreateIssueMessageLinkRequest", "CreateIssueLinkRequest",
            "CreateLabelGroupRequest", "LogTimeRequest", "CreateWorkflowTransitionRequest",
            "CreateCustomFieldRequest", "ChecklistItem", "CreateWorkflowStatusRequest"
        )
        tms = @(
            "TMSService", "NewTMSService", "NewHandler", "Handler", "Repository", "NewEntRepository",
            "NewCombinedRepository", "EntRepository", "CombinedRepository", "ErrNotFound",
            "CreateTestCaseInput", "UpdateTestCaseInput", "ListTestCasesParams", "CreateSuiteInput",
            "UpdateSuiteInput", "CreateTestRunInput", "ListTestRunsParams", "TestRunStats",
            "CreateTestResultInput", "UpdateTestResultInput", "AllureResult",
            "TestCaseExportData", "TestRunExportData", "NewMatcher", "AllureLabel",
            "MatchByAutomationID", "MatchByCaseKey",
            "MongoRepository", "CreateEnvironmentInput", "UpdateEnvironmentInput", "NewTMSProjectProvisioner",
            "ImportFromJSON", "ImportFromXML", "UpdateTestResultRequest", "TestCaseResponse", "TestRunResponse",
            "TestRunDetailResponse", "SuiteResponse", "ComponentResponse", "EnvironmentResponse",
            "CreateTestRunRequest", "TestCaseVersionResponse", "WebhookResponse",
            "JUnitTestCase", "JUnitFailure", "JUnitSkipped", "NewMongoRepository",
            "CreateTestCaseRequest", "UpdateTestCaseRequest", "ListTestCasesRequest",
            "CreateTestResultRequest", "ListTestRunsRequest",
            "AddTestCasesToRunRequest", "AssignTestersRequest", "CreateEnvironmentRequest", "CreateSuiteRequest"
        )
        project = @(
            "ProjectService", "NewProjectService", "NewHandler", "Handler", "EntRepository", "NewEntRepository",
            "Repository", "TemplateService", "NewTemplateService", "TemplateHandler", "NewTemplateHandler",
            "ErrNotFound", "ErrConflict", "ProjectResponse", "MemberResponse", "InviteResponse", "TeamResponse",
            "UpdateProjectRequest", "CreateProjectRequest", "AddMemberRequest",
            "GenerateInviteCode", "TemplateExport"
        )
        messenger = @(
            "MessengerService", "NewMessengerService", "NewHandler", "Handler", "Repository",
            "NewEntRepository", "NewMongoRepository", "NewCombinedRepository", "EntRepository",
            "MongoRepository", "CombinedRepository", "ErrNotFound", "MessageResponse", "MessageHookEvent",
            "PinnedMessageResponse", "ProjectCreatedPayload", "NewAutoChannelService", "AutoChannelService",
            "SendMessageRequest", "EditMessageRequest", "CreateChannelRequest", "ChannelResponse",
            "CreateChannelInviteInput", "CreatePollInput", "UpdateChannelInput",
            "AttachmentRef", "ReactionEntry", "ForwardMessageRequest", "ChannelMemberResponse",
            "BookmarkResponse", "InviteValidationResponse", "ChannelInviteResponse",
            "CreateInviteLinkRequest", "PromoteToIssueRequest", "VotePollRequest", "PollResponse",
            "PromoteToIssueResponse", "NewPresenceService", "NewUnreadService"
        )
        github = @(
            "GitHubService", "NewGitHubService", "NewGitHubServiceWithHTTPClient", "NewHandler", "Handler",
            "Repository", "NewEntRepository", "EntRepository", "ErrNotFound",
            "LinkPreviewService", "NewLinkPreviewService",
            "CreateIntegrationInput", "CreateGitHubLinkInput",
            "CreateIntegrationRequest", "CreateBranchRequest", "CreateBranchResponse",
            "IntegrationResponse", "GitHubLinkResponse", "LinkPreviewResponse",
            "WorkflowRunResponse", "WorkflowRunDetailResponse", "WorkflowJobResponse",
            "ValidateWebhookSignature"
        )
        gitlab = @(
            "GitLabService", "NewGitLabService", "NewGitLabServiceWithHTTPClient", "NewHandler", "Handler",
            "Repository", "NewEntRepository", "EntRepository", "ErrNotFound",
            "CreateIntegrationInput", "CreateIntegrationRequest", "CreateBranchRequest", "CreateBranchResponse",
            "IntegrationResponse", "PipelineResponse", "TriggerPipelineRequest", "TriggerPipelineResponse",
            "WebhookPushEvent", "WebhookMergeRequestEvent", "WebhookPipelineEvent",
            "WebhookProject", "WebhookCommit", "WebhookMergeRequest"
        )
        apigateway = @(
            "APIKeyService", "NewAPIKeyService", "NewEntRepository", "EntRepository",
            "APIKeyRepository", "WebhookRepository", "Repository",
            "Handler", "NewHandler", "WebhookService", "NewWebhookService",
            "WebhookDispatcher", "NewWebhookDispatcher", "APIKeyAuth",
            "CreateWebhookDeliveryInput", "UpdateWebhookEndpointInput",
            "CreateAPIKeyRequest", "APIKeyResponse", "CreateAPIKeyResponse",
            "CreateWebhookRequest", "UpdateWebhookRequest", "WebhookResponse", "WebhookDeliveryResponse"
        )
        backup = @(
            "BackupService", "NewBackupService", "NewEntRepository", "EntRepository", "Repository",
            "Handler", "NewHandler", "BackupResponse", "RestoreRequest"
        )
        search = @(
            "SearchService", "NewSearchService", "NewHandler", "Handler", "Repository", "NewEntRepository", "EntRepository",
            "CreateSavedFilterInput", "SearchResponse", "SearchResult", "SearchRequest",
            "SavedFilterDTO", "SavedFilterCreateRequest", "SavedFilterShareRequest", "SavedFilterListResponse",
            "ParseJQL", "JQLQuery", "JQLCondition", "JQLOrderBy", "ToEntPredicates",
            "NewMeiliIndexerFromEnv", "MeiliIndexer", "EnsureSearchIndexes"
        )
        calls = @(
            "CallService", "NewCallService", "NewHandler", "Handler", "NewEntRepository", "EntRepository",
            "CallRoomService", "NewCallRoomService", "CallEventPublisher", "NewCallEventPublisher",
            "CallChatService", "NewCallChatService", "CallChatMessage", "ChatRepository", "NewChatRepository",
            "RecordingService", "NewRecordingService", "ScheduledCallService", "NewScheduledCallService",
            "CallResponse", "RecordingResponse", "RecordingDownloadResponse", "ScheduledCallResponse",
            "GuestJoinResponse", "CallParticipantResponse", "ConflictResponse",
            "CreateCallRequest", "JoinCallRequest", "RaiseHandRequest", "GuestJoinRequest",
            "CreateScheduledCallRequest", "UpdateScheduledCallRequest", "ErrCallForbidden"
        )
        sprint = @(
            "SprintService", "NewSprintService", "PokerService", "NewPokerService", "NewHandler", "Handler",
            "Repository", "NewEntRepository", "EntRepository", "ErrNotFound",
            "CreateSprintRequest", "UpdateSprintRequest", "RetroNotesRequest", "VoteRequest",
            "StartPokerRequest", "MoveIssueToSprintRequest",
            "SprintResponse", "BurndownPoint", "VelocityPoint", "PokerSessionResponse"
        )
        dashboard = @(
            "DashboardService", "NewDashboardService", "NewHandler", "Handler",
            "Repository", "NewEntRepository", "EntRepository", "ErrNotFound",
            "LayoutItem", "LayoutDTO", "SaveLayoutRequest", "PersonalData", "ProjectStats", "AdminStats",
            "TaskSummary", "CallSummary", "RunSummary", "DataPoint", "ActivityEntry", "ProjectRank",
            "NewDashboardProjectProvisioner", "DashboardProjectProvisioner"
        )
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
        crm = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "ContactResponse", "DealResponse", "PipelineResponse", "PipelineStageResponse",
            "CreateContactRequest", "UpdateContactRequest", "CreateDealRequest", "UpdateDealRequest",
            "DealStage", "PipelineStages",
            "StageLead", "StageQualified", "StageProposal", "StageNegotiation", "StageWon", "StageLost",
            "ErrInvalidInput", "ErrVersionConflict", "ErrProjectForbidden",
            "ExportMapServiceError"
        )
        hr = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "LeaveRequestResponse", "CreateLeaveRequest", "ReviewLeaveRequest",
            "LeaveType", "LeaveStatus",
            "LeaveVacation", "LeaveSick", "LeavePersonal", "LeaveOther",
            "StatusPending", "StatusApproved", "StatusRejected", "StatusCancelled",
            "ErrForbidden", "ErrInvalidInput", "ErrVersionConflict",
            "ExportMapServiceError", "ExportDisplayName", "ExportParseDateRange", "ExportIsValidLeaveType"
        )
        intranet = @(
            "Service", "NewService", "Handler", "NewHandler",
            "Repository", "NewEntRepository", "EntRepository",
            "NewsPostResponse", "NewsListResponse", "CreateNewsRequest", "UpdateNewsRequest",
            "CreateKudoRequest", "KudoResponse", "KudoListResponse",
            "DirectoryUserResponse", "DirectoryResponse", "DirectoryDepartmentGroup",
            "CoffeeEnrollmentResponse", "CoffeePartnerResponse", "CoffeePairResponse",
            "CoffeeRoundResponse", "CoffeeMyMatchResponse", "CoffeeStatsResponse",
            "ErrForbidden", "ErrInvalidInput", "ErrVersionConflict", "ErrSelfKudo", "ErrInvalidKudo", "ErrUserNotFound",
            "ExportMapServiceError", "ExportToDirectoryUserResponse", "ExportDisplayUserName",
            "ExportToCoffeePartner", "ExportPairKey"
        )
        imex = @(
            "ImExService", "NewImExService", "Handler", "NewHandler", "Repository", "EntRepository", "NewEntRepository",
            "FieldMapping", "ImportPreviewRow", "ImportPreviewResponse", "ImportResultResponse", "ExportFilter",
            "IssueExportRow", "CreateIssueInput", "JiraIssue", "JiraFields", "JiraNamedField", "JiraUser",
            "JiraComments", "JiraComment", "JiraSearchResponse", "IsNotFound"
        )
        ai = @(
            "Service", "NewService", "Handler", "NewHandler", "AdminHandler", "NewAdminHandler", "UserHandler", "NewUserHandler",
            "Repository", "EntRepository", "NewEntRepository", "MessengerBot", "NewMessengerBot",
            "ChatAdapter", "OpenAIAdapter", "NewOpenAIAdapter", "AnthropicAdapter", "NewAnthropicAdapter",
            "YandexGPTAdapter", "NewYandexGPTAdapter", "GigaChatAdapter", "NewGigaChatAdapter",
            "ChatMessage", "ChatCompletionRequest", "ChatCompletionResponse", "StreamDelta", "StreamStartResponse",
            "JQLAssistRequest", "JQLAssistResponse", "ProviderResponse", "CreateProviderRequest", "UpdateProviderRequest",
            "TestProviderResponse", "ModelsListResponse", "UsageTodayResponse", "UserPreferenceResponse",
            "UpdateUserPreferenceRequest", "UsageSummaryResponse", "UsageHistoryResponse", "AdminSettingsResponse",
            "UpdateAdminSettingsRequest", "BotInfoResponse", "CreateProviderInput", "UpdateProviderInput",
            "ModelInfo", "UserPreferenceRecord", "ReviewContext", "ComposerSuggestRequest", "ComposerSuggestResponse",
            "SaveUserPreferenceInput", "UsageHistoryDay", "AuditRepoLogger", "NewAuditRepoLogger",
            "SearchDocProvider", "NewSearchDocProvider", "StreamPublisher", "WSStreamPublisher", "NewWSStreamPublisher",
            "TrackerIssueLookup", "NewTrackerIssueLookup", "IssueLookup", "CodeReviewProvider",
            "UnifiedCodeReviewProvider", "NewUnifiedCodeReviewProvider", "DocSearchProvider",
            "SettingsStore", "EntSettingsStore", "NewEntSettingsStore", "ProviderRPMLimiter",
            "RedisProviderRPMLimiter", "NewRedisProviderRPMLimiter", "NoopProviderRPMLimiter",
            "ErrProviderRateLimited", "ErrContentPolicyViolation", "ErrDailyQuotaExceeded", "ErrExternalAIConsentRequired",
            "CheckContentPolicy", "DefaultDailyRequestLimit", "AuditLogger",
            "IsYandexProvider", "DefaultBotSystemPrompt", "EnsureBotUser", "ErrAllProvidersFailed", "ErrNoProvider"
        )
    }
    if ($map.ContainsKey($Mod)) { return $map[$Mod] }
    return @()
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$dstDir = Join-Path $repoRoot "backend/testframework/$Module"
$importPath = "github.com/valerii/hub/backend/internal/$Module"

if (-not (Test-Path $dstDir)) {
    Write-Error "testframework dir not found: $dstDir"
}

$publicSymbols = @(
    "NewConsumer", "NewHandler", "NewRepository", "NewNotificationService", "NewTelegramBindService",
    "NewTelegramSender", "NotificationService", "Repository", "Consumer", "Handler",
    "DispatchEvent", "TypeAssigned", "TypeStatusChanged", "TypeCommentAdded", "TypeBlockerAdded",
    "TypeMessageReceived", "TypeCallStarted", "TypeRunCompleted", "TypeNewLogin", "TypeRunFailed",
    "ErrNotFound", "TelegramBindService", "TelegramSender"
) + (Get-ModuleExtraSymbols -Mod $Module)

$regexOpts = [System.Text.RegularExpressions.RegexOptions]::None
$fixed = 0

Get-ChildItem -Path $dstDir -Filter "*_test.go" | Where-Object { $_.Name -ne 'fixtures_test.go' } | ForEach-Object {
    $Content = Get-Content $_.FullName -Raw -Encoding UTF8
    $orig = $Content

    $Content = $Content -replace '\bfiber\.Handler\b', '__FIBER_HANDLER__'
    $Content = $Content -replace '\bent\.IsNotFound\b', '__ENT_ISNOTFOUND__'
    $Content = $Content -replace '\brbac\.NewEntRepository\b', '__RBAC_NEW_ENT__'

    if ($Module -eq "imex") {
        $Content = $Content -replace '\bbuildFieldMap\(', "${Module}.ExportBuildFieldMap("
        $Content = $Content -replace '\bgetField\(', "${Module}.ExportGetField("
        $Content = $Content -replace '\bmapJiraIssueType\(', "${Module}.ExportMapJiraIssueType("
        $Content = $Content -replace '\bmapJiraPriority\(', "${Module}.ExportMapJiraPriority("
        $Content = $Content -replace '\bjiraDescriptionToText\(', "${Module}.ExportJiraDescriptionToText("
        $Content = $Content -replace '\bbuildTableRow\(', "${Module}.ExportBuildTableRow("
        $Content = $Content -replace '\bbuildTableHeader\(', "${Module}.ExportBuildTableHeader("
        $Content = $Content -replace '\bbuildInfoRow\(', "${Module}.ExportBuildInfoRow("
        $Content = $Content -replace '\bbuildHeaderRow\(', "${Module}.ExportBuildHeaderRow("
    }

    if ($Module -eq "ai") {
        $Content = $Content -replace '\bparseBotCommand\(', "${Module}.ExportParseBotCommand("
        $Content = $Content -replace '\bbuildSummarizePrompt\(', "${Module}.ExportBuildSummarizePrompt("
        $Content = $Content -replace '\bbuildTestCasesPrompt\(', "${Module}.ExportBuildTestCasesPrompt("
        $Content = $Content -replace '\bbuildExplainPrompt\(', "${Module}.ExportBuildExplainPrompt("
        $Content = $Content -replace '\bbuildTranslatePrompt\(', "${Module}.ExportBuildTranslatePrompt("
        $Content = $Content -replace '\bbuildDocsPrompt\(', "${Module}.ExportBuildDocsPrompt("
        $Content = $Content -replace '\bbuildSQLPrompt\(', "${Module}.ExportBuildSQLPrompt("
        $Content = $Content -replace '\bbuildHelpPrompt\(', "${Module}.ExportBuildHelpPrompt("
        $Content = $Content -replace '\bextractPRURL\(', "${Module}.ExportExtractPRURL("
        $Content = $Content -replace '\bextractIssueKey\(', "${Module}.ExportExtractIssueKey("
        $Content = $Content -replace '\bgithubPRToReview\(', "${Module}.ExportGithubPRToReview("
        $Content = $Content -replace '\bgitlabMRToReview\(', "${Module}.ExportGitlabMRToReview("
        $Content = $Content -replace '\bbuildReviewPrompt\(', "${Module}.ExportBuildReviewPrompt("
        $Content = $Content -replace '\bparseYandexCredentials\(', "${Module}.ExportParseYandexCredentials("
        $Content = $Content -replace '\bparseGigaCredentials\(', "${Module}.ExportParseGigaCredentials("
        $Content = $Content -replace '\btruncate\(', "${Module}.ExportTruncate("
        $Content = $Content -replace '\buserIDFromCtx\(', "${Module}.ExportUserIDFromCtx("
        $Content = $Content -replace '\bsecondsUntilMidnightUTC\(', "${Module}.ExportSecondsUntilMidnightUTC("
        $Content = $Content -replace '\bnormalizeModels\(', "${Module}.ExportNormalizeModels("
        $Content = $Content -replace '\bfilterProvidersByConsent\(', "${Module}.ExportFilterProvidersByConsent("
        $Content = $Content -replace '\bconsentErrorForModel\(', "${Module}.ExportConsentErrorForModel("
        $Content = $Content -replace '\bvalidateChatRequest\(', "${Module}.ExportValidateChatRequest("
        $Content = $Content -replace '\bisRetryableProviderError\(', "${Module}.ExportIsRetryableProviderError("
        $Content = $Content -replace '\bisOfflineProvider\(', "${Module}.ExportIsOfflineProvider("
        $Content = $Content -replace '\bextractUserPrompt\(', "${Module}.ExportExtractUserPrompt("
        $Content = $Content -replace '\bmarkdownToPlateJSON\(', "${Module}.ExportMarkdownToPlateJSON("
        $Content = $Content -replace '\bbuildChatRequest\(', "${Module}.ExportBuildChatRequest("
        $Content = $Content -replace '\bextractAssistantText\(', "${Module}.ExportExtractAssistantText("
        $Content = $Content -replace '\bsplitAnthropicMessages\(', "${Module}.ExportSplitAnthropicMessages("
        $Content = $Content -replace '\bnormalizeComposerCompletion\(', "${Module}.ExportNormalizeComposerCompletion("
        $Content = $Content -replace '\bsvc\.resolveProviders\(', "${Module}.ExportServiceResolveProviders(svc, "
        $Content = $Content -replace '\bsvc\.enforceQuota\(', "${Module}.ExportServiceEnforceQuota(svc, "
        $Content = $Content -replace '\bb\.shouldRespond\(', "${Module}.ExportMessengerBotShouldRespond(b, "
        $Content = $Content -replace '&MessengerBot\{botID:\s*(\w+)\}', "${Module}.ExportNewMessengerBotForShouldRespond(`$1)"
        $Content = $Content -replace '\banthropicText\(&anthropicResponse\{Content:\s*\[\]struct\s*\{[^}]+\}\s*\{[^}]+\}\}\)', "${Module}.ExportAnthropicTextFromBlocks(`"hello`")"
        $Content = $Content -replace 'auditmod\.ai\.NewRepository', 'auditmod.NewRepository'
        $Content = $Content -replace '\bcmd\.args\b', "${Module}.ExportBotCommandArgs(cmd)"
        $Content = $Content -replace '\bcmd\.name\b', "${Module}.ExportBotCommandName(cmd)"
        $Content = $Content -replace '\bsettingBotSystemPrompt\b', "${Module}.ExportSettingBotSystemPrompt()"
        $Content = $Content -replace '\bsvc\.giga\s*=\s*ai\.ExportNewGigaChatAdapterWithClient\((\w+)\)', "${Module}.ExportServiceSetGigaAdapter(svc, ${Module}.ExportNewGigaChatAdapterWithClient(`$1))"
        $Content = $Content -replace '\bsvc\.anthropic\s*=\s*ai\.ExportNewAnthropicAdapterWithClient\((\w+)\)', "${Module}.ExportServiceSetAnthropicAdapter(svc, ${Module}.ExportNewAnthropicAdapterWithClient(`$1))"
        $Content = $Content -replace '\ba\.client\b', "${Module}.ExportOpenAIAdapterClient(a)"
        $Content = $Content -replace '&ai\.OpenAIAdapter\{client:\s*(\w+)\}', "${Module}.ExportNewOpenAIAdapterWithClient(`$1)"
        $Content = $Content -replace '&ai\.YandexGPTAdapter\{client:\s*(\w+)\}', "${Module}.ExportNewYandexGPTAdapterWithClient(`$1)"
        $Content = $Content -replace '&ai\.GigaChatAdapter\{client:\s*(\w+)\}', "${Module}.ExportNewGigaChatAdapterWithClient(`$1)"
        $Content = $Content -replace '&ai\.AnthropicAdapter\{client:\s*(\w+)\}', "${Module}.ExportNewAnthropicAdapterWithClient(`$1)"
        $Content = $Content -replace '\byandexCompletionURL\b', "${Module}.ExportYandexCompletionURL()"
        $Content = $Content -replace '\ba\.accessToken\(', "${Module}.ExportGigaChatAccessToken(a, "
        $Content = $Content -replace '\bsvc\.giga\s*=\s*&ai\.GigaChatAdapter\{client:\s*(\w+)\}', "${Module}.ExportServiceSetGigaAdapter(svc, ${Module}.ExportNewGigaChatAdapterWithClient(`$1))"
        $Content = $Content -replace '\bsvc\.anthropic\s*=\s*&ai\.AnthropicAdapter\{client:\s*(\w+)\}', "${Module}.ExportServiceSetAnthropicAdapter(svc, ${Module}.ExportNewAnthropicAdapterWithClient(`$1))"
        $Content = $Content -replace '\bsvc\.yandex\s*=\s*&ai\.YandexGPTAdapter\{client:\s*(\w+)\}', "${Module}.ExportServiceSetYandexAdapter(svc, ${Module}.ExportNewYandexGPTAdapterWithClient(`$1))"
    }

    foreach ($sym in $publicSymbols) {
        $escaped = [regex]::Escape($sym)
        $pattern = "(?<![.\w])\b$escaped\b"
        $Content = [System.Text.RegularExpressions.Regex]::Replace($Content, $pattern, "${Module}.$sym", $regexOpts)
    }

    $Content = $Content -replace '__FIBER_HANDLER__', 'fiber.Handler'
    $Content = $Content -replace '__ENT_ISNOTFOUND__', 'ent.IsNotFound'
    $Content = $Content -replace '__RBAC_NEW_ENT__', 'rbac.NewEntRepository'

    if ($Module -eq "project") {
        if ($Content -match 'ent/project' -and $Content -notmatch 'entproject ') {
            $Content = $Content -replace '"github.com/valerii/hub/backend/ent/project"', 'entproject "github.com/valerii/hub/backend/ent/project"'
        }
        $Content = $Content -replace '\bproject\.Status', 'entproject.Status'
    }

    if ($Content -match "\b${Module}\." -and $Content -notmatch [regex]::Escape($importPath)) {
        $importBlock = "`t`"$importPath`"`r`n"
        if ($Content -match '(?ms)(import \(\r?\n)') {
            $Content = $Content -replace '(?ms)(import \(\r?\n)', "`${1}$importBlock"
        }
    }

    if ($Content -ne $orig) {
        Set-Content -Path $_.FullName -Value $Content -Encoding UTF8 -NoNewline
        Write-Host "Fixed: $($_.Name)"
        $fixed++
    }
}

Write-Host "Done. Fixed $fixed file(s) in testframework/$Module"
