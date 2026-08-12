# TF-MIGRATE E2 — `export_hooks.go` after co-locate

**Дата анализа:** 2026-07-27  
**Вход:** [`02-TEST-FRAMEWORK-MIGRATION.md`](02-TEST-FRAMEWORK-MIGRATION.md) §Э3 + [E1 inventory](TF-MIGRATE-E1-INVENTORY.md)  
**Статус:** analysis only — **удаление hooks / перенос тестов не начаты**  
**Модель:** Cursor Grok 4.5 (cursor-grok-4.5-high)

---

## HARD CONSTRAINT (gate)

> **Этап Э1 миграции** (каркас Python `/tests`) **не начинать**, пока не существует **A1 coverage baseline** (gate **2026-07-27**).
>
> Анализ ниже **не** разрешает трогать `export_hooks.go` в коде до того же gate + до фактического co-locate юнитов (Э3). Иначе можно снести швы, пока тесты ещё black-box в `testframework/`, и сломать сборку `-tags hubtest`.
>
> В [`01-PLANNING-FIXES.md`](01-PLANNING-FIXES.md) перезамер coverage = **A3**. В этом workstream тот же gate называется **A1 coverage baseline**.

---

## Контекст: зачем `export_hooks` существуют

| Факт | Значение |
|---|---|
| Файлов `export_hooks.go` под `backend/internal/**` | **35** |
| Все с `//go:build hubtest` | **да (35/35)** |
| `Export*` функций | **444** |
| `Export*` типов | **11** (в auth/calls/dev/notification/search) |
| Тесты сегодня | **515/515** `package <domain>_test` (external) — см. E1 |
| Файлов тестов с вызовами `Export*` | **218** (~1301 match lines) |
| Cross-package consumers (`otherPkg.Export…` из чужого домена) | **0 найдено** |

`export_hooks.go` — единственный способ для **external** тестов (`package X_test`) дотянуться до **unexported** символов того же пакета без нарушения Go visibility. Под `-tags hubtest` они компилируются; в release-сборке файла нет.

Прод-методы вроде `TMSService.ExportTestCases` / `ImExService.ExportIssuesCSV` — **не** hubtest-hooks; в этот анализ не входят.

---

## Целевая модель после co-locate (Э3)

Предположение миграции (`02-…` §Э3): юниты переезжают в `backend/internal/<domain>/` и становятся **white-box** — `package <domain>` (тот же пакет, что прод-код).

Тогда:

```text
package auth_test  +  auth.ExportMagicLinkTTL()
        ↓ Э3
package auth       +  magicLinkTTL   // прямой доступ, hook не нужен
```

Альтернатива: оставить часть сюит **black-box** (`package X_test` рядом с кодом) — тогда соответствующие `Export*` **MUST remain** (или тесты переписываются на публичный API).

**Рекомендация для Hub:** white-box для юнитов (обязанность разработчика), black-box оставить только там, где проверяется публичный контракт handler/DTO. По умолчанию — **DELETE всех 35 файлов** после завершения white-box переезда домена.

---

## Вердикт сводка

| Класс | Count | Действие |
|---|---:|---|
| **DELETE candidates** — целые `export_hooks.go` после white-box co-locate домена | **35 / 35** | удалить файл, когда 0 remaining `package X_test` в домене нуждаются в `Export*` |
| **DELETE candidates** — тонкие `Export*` bridges (wrap unexported func/const/field/converter) | **~396** | заменить прямым вызовом / полем в том же пакете |
| **KEEP-as-test-helper (logic)** — DI / test-double seams (~48) | **~48** | **не** оставлять как hubtest prod-файл; перенести логику в `*_test.go` / `testing_support_test.go` того же пакета |
| **KEEP as hubtest `export_hooks.go`** | **0** (целевое состояние) | допустим только **временно** на время смешанного black/white домена или если сознательно оставляют `package X_test` |
| Домены **без** hooks (уже public-API) | cron, crypto, errors, nats, telegram, status + `env` helper | hooks не появятся; E2 N/A |

### Когда KEEP (исключения)

1. **Переходный dual-mode:** часть тестов ещё в `testframework/` (`X_test`), часть уже в `internal/` — hooks живут, пока не опустеет external suite домена.
2. **Сознательный black-box:** конкретный файл остаётся `package X_test` и зовёт unexported — тогда точечно оставить нужные `Export*` (лучше сузить файл, чем держать 444 символа).
3. **Не найдено:** кросс-доменных потребителей `other.Export*` — аргумент «оставить hooks для соседей» сейчас не подтверждён.

### Что НЕ KEEP в prod

- Зеркала типов (`ExportRefreshPayload`, `ExportTelegramMsg`, …) — в white-box использовать настоящий unexported type.
- Геттеры полей (`ExportHandlerService`, `ExportCallServiceRepo`, …) — прямой доступ к полю.
- Const/TTL/prefix exporters — прямые unexported const.
- Converter wrappers (`ExportToIssueResponse`, …) — вызов unexported `toIssueResponse`.

---

## Классификация helper-типов

| Тип шва | Пример | После white-box |
|---|---|---|
| Thin func wrap | `ExportParseCORSOrigins` → `parseCORSOrigins` | **DELETE** |
| Const / config getter | `ExportMagicLinkTTL` | **DELETE** |
| Private field accessor | `ExportAuthServiceRedis` | **DELETE** |
| Method wrap | `ExportRateLimiterAllow` | **DELETE** |
| DTO converter wrap | `ExportToIssueResponse` | **DELETE** |
| Type mirror | `ExportRefreshPayload` | **DELETE** (use real type) |
| Test ctor / DI seam | `ExportNewSessionServiceForJWTTest`, `ExportResendEmailClientInjectHTTP` | **DELETE file**; **KEEP logic** in `*_test.go` if still useful |
| Fake repos | `ExportNewMemChatRepository`, `ExportNewChatRepositoryFailingInsert` | same — test-only helpers |

---

## Порядок безопасного удаления (на домен)

```text
1. A1 coverage baseline зафиксирован (gate)
2. Домен: git mv testframework/<d>/*.go → internal/<d>/
3. package <d>_test → package <d>; починить импорты
4. Заменить Package.ExportFoo(...) на foo(...) / прямые поля
5. go test ./internal/<d>/...  (без -tags hubtest, если hooks больше не нужны)
6. Удалить internal/<d>/export_hooks.go
7. Сверить coverage домена с A1 baseline slice
```

Не удалять hooks **глобально** одним коммитом: tms/tracker/ai/notification — крупные; резать по доменам.

---

## Per-file delete vs keep

### `backend/internal/accounting/export_hooks.go`

- **Domain:** `accounting`
- **Export funcs:** 3 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportParseEntryDate, ExportFiberStatusFromError`

### `backend/internal/ai/export_hooks.go`

- **Domain:** `ai`
- **Export funcs:** 48 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportNewOpenAIAdapterWithClient` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewYandexGPTAdapterWithClient` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewGigaChatAdapterWithClient` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewAnthropicAdapterWithClient` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportServiceSetGigaAdapter` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportServiceSetAnthropicAdapter` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportServiceSetYandexAdapter` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportParseBotCommand, ExportBuildSummarizePrompt, ExportBuildTestCasesPrompt, ExportBuildExplainPrompt, ExportBuildTranslatePrompt, ExportBuildDocsPrompt, ExportBuildSQLPrompt, ExportBuildHelpPrompt, ExportExtractPRURL, ExportExtractIssueKey, ExportGithubPRToReview, ExportGitlabMRToReview, ExportBuildReviewPrompt, ExportParseYandexCredentials, ExportParseGigaCredentials, ExportTruncate, ExportUserIDFromCtx, ExportSecondsUntilMidnightUTC, ExportNormalizeModels, ExportFilterProvidersByConsent, ExportConsentErrorForModel, ExportValidateChatRequest, ExportIsRetryableProviderError, ExportIsOfflineProvider, ExportExtractUserPrompt, ExportMarkdownToPlateJSON, ExportBuildChatRequest, ExportExtractAssistantText, ExportSplitAnthropicMessages, ExportAnthropicTextFromBlocks, ExportNormalizeComposerCompletion, ExportServiceResolveProviders, ExportServiceEnforceQuota, ExportNewMessengerBotForShouldRespond, ExportMessengerBotShouldRespond, ExportBotCommandName, ExportBotCommandArgs, ExportSettingBotSystemPrompt, ExportNewOpenAIAdapterWithClient, ExportNewYandexGPTAdapterWithClient, ExportNewGigaChatAdapterWithClient, ExportNewAnthropicAdapterWithClient, ExportYandexCompletionURL, ExportGigaChatAccessToken, ExportServiceSetGigaAdapter, ExportServiceSetAnthropicAdapter, ExportServiceSetYandexAdapter, ExportOpenAIAdapterClient`

### `backend/internal/apigateway/export_hooks.go`

- **Domain:** `apigateway`
- **Export funcs:** 17 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportSetWebhookHTTPClient` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportResetWebhookHTTPClient` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportHashKey, ExportSignPayload, ExportContainsEvent, ExportWebhookSendWebhook, ExportSetWebhookHTTPClient, ExportResetWebhookHTTPClient, ExportMaxWebhookResponseBodyBytes, ExportParseID, ExportGetUserID, ExportHandlerWebhookService, ExportWebhookDispatcherHandleMessage, ExportWebhookDispatcherDeliverWithRetry, ExportWebhookDispatcherInitJobs, ExportWebhookDispatcherJobAvailable, ExportWebhookDispatcherDispatchEvent, ExportErrWebhookQueueFull, ExportWebhookFilterSubjects`

### `backend/internal/audit/export_hooks.go`

- **Domain:** `audit`
- **Export funcs:** 18 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportNewAuditConsumerForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportProcessEvent, ExportNewAuditConsumerForTest, ExportDaysRemaining, ExportConsumerMapSubjectToAction, ExportExtractStringField, ExportMapCRUDSuffix, ExportExtractUUIDField, ExportSlowRequestThresholdMS, ExportRetentionServicePurgeOldAllureResults, ExportRetentionServicePurgeOldMessages, ExportRetentionServiceGetRetentionDays, ExportRetentionServicePurgeOldCallRecordings, ExportDefaultChatRetentionDays, ExportDefaultAllureRetentionDays, ExportDefaultCallRecordingRetention, ExportShouldSkipHTTPAudit, ExportStripQuery, ExportRepositoryEntClient`

### `backend/internal/auth/export_hooks.go`

- **Domain:** `auth`
- **Export funcs:** 29 · **Export types:** 1
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportResendEmailClientInjectHTTP` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewTelegramBotAuthForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewTelegramBotAuthForTestWithBot` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportSessionSetNATS` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportTelegramHandleLogin, ExportHandlerService, ExportMagicLinkVerifyMax, ExportMagicLinkRateMax, ExportMagicLinkTTL, ExportMagicLinkFailsPrefix, ExportAuthServiceMagicLink, ExportAuthServiceRedis, ExportUserStatusCacheDeactivated, ExportUserStatusCacheTTL, ExportResendEmailClientAPIKey, ExportResendEmailClientFromEmail, ExportResendEmailClientInjectHTTP, ExportTGCodePrefix, ExportTGFailsPrefix, ExportGenerateTGCode, ExportNewTelegramBotAuthForTest, ExportNewTelegramBotAuthForTestWithBot, ExportTelegramRecordVerifyFailure, ExportHashRefreshToken, ExportRefreshRedisKey, ExportRefreshRedisKeyStored, ExportSessionStoreRefreshToken, ExportSessionInvalidateUserSessions, ExportSessionSetNATS, ExportSessionPublishSessionCreated, ExportGenerateCode, ExportNewSessionServiceForJWTTest, ExportSessionGenerateAccessToken`
- **Export types (delete; use real unexported types in same package):**
  `ExportRefreshPayload`

### `backend/internal/backup/export_hooks.go`

- **Domain:** `backup`
- **Export funcs:** 18 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportBackupSetURLs` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportBackupSetMaxBackups` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportBackupSetMaintenanceMode` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportBackupSetS3` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportToBackupResponse, ExportStrPtr, ExportBackupExtractZIP, ExportBackupMarkFailed, ExportBackupLifecycleCtx, ExportNewBackupServiceForURLTest, ExportBackupSetURLs, ExportBackupDumpMongo, ExportBackupDumpPostgres, ExportBackupRestorePostgres, ExportBackupRestoreMongo, ExportBackupCreateZIP, ExportBackupIsNotFound, ExportBackupSetMaxBackups, ExportBackupCleanupOldBackups, ExportBackupSetMaintenanceMode, ExportBackupSetS3, ExportBackupS3Bucket`

### `backend/internal/bases/export_hooks.go`

- **Domain:** `bases`
- **Export funcs:** 3 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportNormalizeCells, ExportFiberStatusFromError`

### `backend/internal/calendar/export_hooks.go`

- **Domain:** `calendar`
- **Export funcs:** 4 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportExpandCallEvents, ExportParseRRule, ExportLeaveEvent, ExportValidateRange`

### `backend/internal/calls/export_hooks.go`

- **Domain:** `calls`
- **Export funcs:** 38 · **Export types:** 1
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportNewCallServiceForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithCfg` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithRecSvc` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithServiceAndScheduled` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithServiceAndRecSvc` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithScheduledSvc` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewHandlerWithRecSvcAndRBAC` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewRecordingServiceForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewScheduledCallServiceForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportCallServiceSetRoom` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportCallServiceSetChat` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewMemChatRepository` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewChatRepositoryFailingInsert` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewChatRepositoryFailingDelete` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportChatRepositoryInsertForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewMongoChatStore` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportToRecordingResponse, ExportToScheduledCallResponse, ExportToCallResponse, ExportParseCallID, ExportVerifyWebhookSignature, ExportFinalizeCreate, ExportResolveParticipant, ExportDefaultMaxCallParticipants, ExportMaxActiveRecordingsPerCall, ExportNewCallServiceForTest, ExportNewHandlerForTest, ExportNewHandlerWithCfg, ExportNewHandlerWithRecSvc, ExportNewHandlerWithServiceAndScheduled, ExportNewHandlerWithServiceAndRecSvc, ExportNewHandlerWithScheduledSvc, ExportNewHandlerWithRecSvcAndRBAC, ExportNewRecordingServiceForTest, ExportNewScheduledCallServiceForTest, ExportCallServiceRepo, ExportCallServiceRoom, ExportCallServicePub, ExportCallServiceChat, ExportCallServiceSetRoom, ExportCallServiceSetChat, ExportHandlerService, ExportRecordingServiceRepo, ExportRecordingServicePresignClient, ExportCallRoomServiceCfg, ExportCallEventPublisherNats, ExportNewMemChatRepository, ExportNewChatRepositoryFailingInsert, ExportNewChatRepositoryFailingDelete, ExportChatRepositoryInsertForTest, ExportChannelIDStr, ExportUserNameFromCtx, ExportGenerateGuestToken, ExportNewMongoChatStore`
- **Export types (delete; use real unexported types in same package):**
  `ExportMongoChatStore`

### `backend/internal/crm/export_hooks.go`

- **Domain:** `crm`
- **Export funcs:** 2 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportFiberStatusFromError`

### `backend/internal/dashboard/export_hooks.go`

- **Domain:** `dashboard`
- **Export funcs:** 4 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportDashboardServiceRedis, ExportProjectStatsTTL, ExportPriorityRank, ExportGetUserID`

### `backend/internal/dev/export_hooks.go`

- **Domain:** `dev`
- **Export funcs:** 9 · **Export types:** 7
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportDemoUsers, ExportDemoStatuses, ExportDemoIssues, ExportDemoTestCases, ExportDemoChannels, ExportMethodOrder, ExportMethodColor, ExportGuessModule, ExportCollectRoutes`
- **Export types (delete; use real unexported types in same package):**
  `ExportDemoUser, ExportDemoStatus, ExportDemoIssue, ExportDemoTestStep, ExportDemoTestCase, ExportDemoChannel, ExportRouteEntry`

### `backend/internal/github/export_hooks.go`

- **Domain:** `github`
- **Export funcs:** 16 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportToIntegrationResponse, ExportToGitHubLinkResponse, ExportIssueKeyRegex, ExportSlugify, ExportTruncate, ExportMatchTestCaseFileByPath, ExportParseGitHubPRURL, ExportToWorkflowRunResponse, ExportParseProjectID, ExportMaxWebhookBodyBytes, ExportEntRepositoryEncryptField, ExportEntRepositoryDecryptField, ExportNewEntRepositoryForEncryptTest, ExportLinkPreviewResolveExternalPreview, ExportGitHubServiceMarkTestCasesOutdated, ExportGitHubServiceHTTPClient`

### `backend/internal/gitlab/export_hooks.go`

- **Domain:** `gitlab`
- **Export funcs:** 6 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportToIntegrationResponse, ExportIssueKeyRegex, ExportSlugify, ExportTruncate, ExportParseGitLabMRURL, ExportMaxWebhookBodyBytes`

### `backend/internal/hr/export_hooks.go`

- **Domain:** `hr`
- **Export funcs:** 5 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportDisplayName, ExportParseDateRange, ExportIsValidLeaveType, ExportFiberStatusFromError`

### `backend/internal/imex/export_hooks.go`

- **Domain:** `imex`
- **Export funcs:** 10 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportBuildFieldMap, ExportGetField, ExportMapJiraIssueType, ExportMapJiraPriority, ExportJiraDescriptionToText, ExportParseJiraJSON, ExportBuildHeaderRow, ExportBuildInfoRow, ExportBuildTableHeader, ExportBuildTableRow`

### `backend/internal/integrations/onec/export_hooks.go`

- **Domain:** `integrations/onec`
- **Export funcs:** 7 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportServiceSetClient` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportSettingsKey, ExportConfigValidate, ExportDecodeConfig, ExportEncodeConfig, ExportServiceSetClient, ExportServicePost, ExportBuildLeavePayload`

### `backend/internal/intranet/export_hooks.go`

- **Domain:** `intranet`
- **Export funcs:** 5 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportToDirectoryUserResponse, ExportDisplayUserName, ExportToCoffeePartner, ExportPairKey`

### `backend/internal/mail/export_hooks.go`

- **Domain:** `mail`
- **Export funcs:** 2 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportInternalTokenHeader, ExportGetEnv`

### `backend/internal/marketplace/export_hooks.go`

- **Domain:** `marketplace`
- **Export funcs:** 9 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportMapServiceError, ExportDisplayName, ExportFormatMinorUnits, ExportItoa, ExportPad2, ExportIsValidCategory, ExportFormatPrice, ExportValidateListingInput, ExportFiberStatusFromError`

### `backend/internal/messenger/export_hooks.go`

- **Domain:** `messenger`
- **Export funcs:** 24 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportAutoChannelServiceSetLifecycleCtx` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportNewMessengerServiceTest, ExportAutoChannelServiceHandleProjectCreated, ExportAutoChannelServiceSetLifecycleCtx, ExportAutoChannelServiceSubscription, ExportToMessageResponse, ExportToChannelResponse, ExportNewMessengerServiceEmpty, ExportToChannelInviteResponse, ExportToPollResponse, ExportSplitComma, ExportQueryInt, ExportMessengerServiceUnread, ExportMessengerServiceResolveGroupMentions, ExportParseChannelID, ExportParseMessageID, ExportPresenceOnlineKey, ExportMapKeys, ExportParseExpiresIn, ExportMessengerServiceEmitMessageSent, ExportMessengerServiceIssueThreadLinker, ExportMessengerServicePromoteCheck, ExportGenerateInviteToken, ExportInviteTokenLength, ExportBase62Chars`

### `backend/internal/notification/export_hooks.go`

- **Domain:** `notification`
- **Export funcs:** 44 · **Export types:** 1
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportSeedDebounceEntry` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewTestTelegramSender` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportSetServicePush` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewTestConsumerWithDebounce` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportHandleStatusChanged, ExportHandleMessageReceived, ExportHandleCallStarted, ExportHandleAssigned, ExportHandleBlockerAdded, ExportHandleCommentAdded, ExportHandleRunCompleted, ExportHandleNewLogin, ExportProcessEvent, ExportShouldDispatch, ExportSeedDebounceEntry, ExportDebounceLen, ExportCleanDebounce, ExportRepositoryDB, ExportBuildPushURL, ExportIsRunFailedRecipient, ExportTelegramBindRepository, ExportTelegramHandleMessage, ExportTelegramHandleUnlinkCommand, ExportTelegramHandleStartBind, ExportTelegramHandleLinkCommand, ExportTelegramHandleTestCommand, ExportNewTestTelegramSender, ExportTelegramSenderQueueReady, ExportTelegramTryRecv, ExportSendTelegram, ExportTelegramSendMessage, ExportFormatTelegramMessage, ExportSubjectToType, ExportTelegramBindSender, ExportPushSendToUser, ExportPushSendToSubscription, ExportIsQuietHours, ExportSetServicePush, ExportGetUserID, ExportParseTimeMinutes, ExportMatchSuffix, ExportNotifToDTO, ExportNewTestConsumerWithDebounce, ExportPublishToWS, ExportTGBindPrefix, ExportTGBindTTL, ExportTGLinkCodePrefix, ExportConsumerStart`
- **Export types (delete; use real unexported types in same package):**
  `ExportTelegramMsg`

### `backend/internal/platform/config/export_hooks.go`

- **Domain:** `platform/config`
- **Export funcs:** 3 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportIsDevDefault, ExportGetEnvInt, ExportGetEnvBool`

### `backend/internal/platform/http/export_hooks.go`

- **Domain:** `platform/http`
- **Export funcs:** 2 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportParseOpenGraphHTML, ExportIsBlockedIP`

### `backend/internal/platform/middleware/export_hooks.go`

- **Domain:** `platform/middleware`
- **Export funcs:** 8 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportNewRateLimiterForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportParseCORSOrigins, ExportStripQuery, ExportIsDevMode, ExportNormalizeEmail, ExportHashIP, ExportNewRateLimiterForTest, ExportRateLimiterAllow, ExportRateLimiterCleanup`

### `backend/internal/platform/observability/export_hooks.go`

- **Domain:** `platform/observability`
- **Export funcs:** 4 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportSetTracingActiveForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportRecordOperation, ExportParseCIDRs, ExportIpAllowed, ExportSetTracingActiveForTest`

### `backend/internal/platform/outbox/export_hooks.go`

- **Domain:** `platform/outbox`
- **Export funcs:** 4 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportWorkerClient, ExportWorkerNats, ExportWorkerProcessEvent, ExportWorkerProcessBatch`

### `backend/internal/project/export_hooks.go`

- **Domain:** `project`
- **Export funcs:** 7 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportToProjectResponse, ExportToMemberResponse, ExportToInviteResponse, ExportTemplateExportVersion, ExportProjectServiceProvisionProject, ExportProjectServiceNotifyLinkDeleted, ExportToTemplateResponse`

### `backend/internal/rbac/export_hooks.go`

- **Domain:** `rbac`
- **Export funcs:** 8 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportRBACServiceCache, ExportCacheKey, ExportPermissionRegistry, ExportBuildPermissionNameMap, ExportPermissionNameMap, ExportPermissionsFromNames, ExportPermissionNames, ExportValidatePermissionNames`

### `backend/internal/search/export_hooks.go`

- **Domain:** `search`
- **Export funcs:** 24 · **Export types:** 1
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportSetSearchServiceRepo` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportGetUserID, ExportSavedFilterToDTO, ExportNewSearchServiceSQL, ExportNewSearchServiceRepoOnly, ExportNewSearchServiceRepoSQL, ExportNewSearchServiceFull, ExportSetSearchServiceRepo, ExportSearchIssues, ExportSearchTestCases, ExportSearchUsers, ExportSearchUsersUnscoped, ExportSearchProjects, ExportSearchFiles, ExportSearchChannels, ExportSearchMessages, ExportGetUserChannelIDs, ExportGetUserProjectIDs, ExportIsSystemAdmin, ExportUUIDSliceToPgArray, ExportDateFieldPredicate, ExportIssuesIndex, ExportNewMeiliIndexerTest, ExportMeiliIndexerHost, ExportMeiliUpsertDocuments`
- **Export types (delete; use real unexported types in same package):**
  `ExportMeiliIssueDoc`

### `backend/internal/sprint/export_hooks.go`

- **Domain:** `sprint`
- **Export funcs:** 2 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportSprintServiceToSprintResponse, ExportNilIfEmpty`

### `backend/internal/storage/export_hooks.go`

- **Domain:** `storage`
- **Export funcs:** 7 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportNewStorageServiceForTest` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportNewStorageServiceForTestWithS3` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportSanitizeFilename, ExportNewStorageServiceForTest, ExportNewStorageServiceForTestWithS3, ExportUploadURLTTL, ExportDownloadURLTTL, ExportUserIDFromCtx, ExportHandlerAuthorizeProject`

### `backend/internal/tms/export_hooks.go`

- **Domain:** `tms`
- **Export funcs:** 6 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportJUnitTestCaseToAllure, ExportParseAllureResults, ExportParseAllureResult, ExportExtractLabelValue, ExportAllureStatusToTMS, ExportParseUUIDs`

### `backend/internal/tracker/export_hooks.go`

- **Domain:** `tracker`
- **Export funcs:** 26 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - _(none — pure thin wrappers / converters / const getters)_
- **All Export* (delete-as-hubtest-bridge):**
  `ExportParseIssueEventKind, ExportAutomationConsumerProcessMessage, ExportMaxAutomationRulesPerProject, ExportRenderAutomationTemplate, ExportValidateAutomationRuleRequest, ExportAutomationEngineExecuteRule, ExportToAutomationRuleResponse, ExportToCommentResponse, ExportToActivityLogResponse, ExportToTimeLogResponse, ExportToLabelGroupResponse, ExportToBoardResponse, ExportToIssueResponse, ExportToIssueBriefResponse, ExportToBoardDataResponse, ExportToIssueLinkResponse, ExportToCustomFieldResponse, ExportToWorkflowStatusResponse, ExportBuildFieldChanges, ExportAggregateReactions, ExportExtractPlainText, ExportBuildSwimlanes, ExportTrackerServiceWorkflow, ExportTrackerServiceMirrorIssueCommentsToThreads, ExportConfigString, ExportSwimlaneKeyLabel`

### `backend/internal/ws/export_hooks.go`

- **Domain:** `ws`
- **Export funcs:** 22 · **Export types:** 0
- **Verdict after white-box co-locate:** **DELETE file** (all bridges become direct unexported access)
- **Optional KEEP-as-test-helper (move into `*_test.go`, not hubtest prod file):**
  - `ExportAllowAllAuthorizer` — test double / DI seam; recreate as unexported helper in test file if still needed
  - `ExportResetOriginCheck` — test double / DI seam; recreate as unexported helper in test file if still needed
- **All Export* (delete-as-hubtest-bridge):**
  `ExportRegisterClient, ExportRemoveClient, ExportConnectionsForUser, ExportBridgeHandleProjectEvent, ExportBridgeHandleMessengerEvent, ExportBridgeHandleCallEvent, ExportBridgeHandleCallChatEvent, ExportBridgeHandleAuthSessionEvent, ExportBridgeHandleNotificationEvent, ExportClientSendCh, ExportSplitRoom, ExportAllowAllAuthorizer, ExportClientDone, ExportClientCloseDone, ExportBridgeSubsLen, ExportBridgeSubs, ExportReadPump, ExportClientRooms, ExportResetOriginCheck, ExportCheckOrigin, ExportWritePump, ExportWSUpgrader`


---

## Домены без `export_hooks.go`

| Path | `*_test.go` (E1) | Замечание |
|---|---:|---|
| `platform/cron` | 1 | public API |
| `platform/crypto` | 1 | public API |
| `platform/errors` | 2 | public API |
| `platform/nats` | 3 | public API |
| `platform/telegram` | 1 | public API |
| `status` | 2 | public API; hooks отсутствуют |
| `env (testframework helper)` | 1 | не прод-домен |

---

## Риски при чистке hooks

| Риск | Митигация |
|---|---|
| Удалить hooks до переезда тестов | Только после white-box в том же домене |
| Потеря DI-seams (InjectHTTP, ForTest ctors) | Сначала перенести в `*_test.go`, потом удалить `export_hooks.go` |
| Смешанный домен (часть external) | Не удалять файл, пока `rg Export` по домену = 0 |
| Coverage drift | A1 baseline + per-domain cover после каждого домена |
| Случайно удалить прод `Export*` API (imex/tms) | Искать только `export_hooks.go` / `Export` + `hubtest` |

---

## Связь с этапами

| Этап | Отношение к hooks |
|---|---|
| **Э1** Python scaffold | **blocked** на A1 coverage baseline; hooks не трогать |
| **Э3** co-locate Go units | единственное окно удаления hooks |
| **Э7** CI cleanup | убрать `-tags hubtest` из unit job, когда 0 `export_hooks.go` осталось |

---

## Итог

- **35/35** `export_hooks.go` — **DELETE candidates** после white-box co-locate соответствующего домена.
- **~48** seam-хелпера — **KEEP logic** только как test-file helpers, не как hubtest prod bridges.
- **Целевых KEEP hubtest-файлов: 0.**
- **Э1 / правки hooks в коде — стоп до A1 coverage baseline (2026-07-27 gate).**

---

*Документ: planning only. Production code не изменялся.*
