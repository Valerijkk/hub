---
phase: 07-polish-extended
verified: 2026-03-21T12:00:00Z
status: gaps_found
score: 3/5 success criteria verified
re_verification: null
gaps:
  - truth: "REST API документирован через Swagger, можно генерировать API-ключи, вебхуки срабатывают на события, и автоматические резервные копии создаются еженедельно"
    status: failed
    reason: "Swagger UI (API-01) deferred — no swag annotations, no /api/docs endpoint, no docs/ directory. Webhook dispatcher (API-03) Start() goroutine does not subscribe to NATS JetStream; dispatchEvent() exists but is never called. JQL parser (SRCH-03) is implemented and tested but no HTTP endpoint registers it — ParseJQL/ToEntPredicates are orphaned."
    artifacts:
      - path: "backend/internal/apigateway/webhook.go"
        issue: "WebhookDispatcher.Start() is a stub — spins goroutine that only waits for ctx.Done(); does not subscribe to JetStream or route any events to dispatchEvent()"
      - path: "backend/internal/search/jql.go"
        issue: "ParseJQL and ToEntPredicates exist and are tested but are not called by any handler — no /api/projects/:id/issues/jql endpoint exists"
    missing:
      - "Swagger: add swag annotations to key handlers and generate docs with `swag init`; add /api/docs route"
      - "Webhook: implement JetStream consumer in Start() that routes events to dispatchEvent()"
      - "JQL: add GET /api/projects/:projectId/issues/jql?q=... endpoint in tracker handler calling ParseJQL then ToEntPredicates"
  - truth: "Задачи и прогоны можно экспортировать в CSV/Excel/PDF/ZIP"
    status: partial
    reason: "IMEX-03 PDF export (test runs PDF, sprint PDF) was explicitly deferred in 07-05-SUMMARY. ExportRunsPDF and ExportSprintPDF are not implemented. CSV/Excel/ZIP export for issues is implemented."
    artifacts:
      - path: "backend/internal/imex/service.go"
        issue: "No ExportRunsPDF or ExportSprintPDF methods; service only has CSV, JSON, Excel (returns CSV), and ZIP exports"
    missing:
      - "Implement ExportRunsPDF (maroto v2) for test run PDF reports"
      - "Implement ExportSprintPDF (maroto v2) for sprint PDF reports"
      - "Wire /api/test-runs/:runId/export/pdf and /api/projects/:projectId/sprints/:sprintId/export/pdf endpoints"
human_verification:
  - test: "Planning Poker real-time flow"
    expected: "Participants see vote-cast events in real time without vote values; reveal shows all values with average/median"
    why_human: "Real-time WebSocket behavior with Redis + NATS cannot be verified statically"
  - test: "GitHub webhook auto-link with real repository"
    expected: "Push event with 'HUB-123' in commit message creates GitHubLink entry; PR merge triggers issue status change if merge_status_id set"
    why_human: "Requires live GitHub webhook delivery to verify end-to-end"
  - test: "PWA install on mobile"
    expected: "Browser shows 'Add to Home Screen' prompt; app opens in standalone mode"
    why_human: "PWA installability depends on browser environment and HTTPS"
  - test: "Language switching"
    expected: "Switching RU/EN in settings changes all UI labels immediately"
    why_human: "Requires visual verification across all modules; not all strings may use t()"
  - test: "Call recording download"
    expected: "Completed recording shows in /projects/:id/recordings with working download URL"
    why_human: "Requires live LiveKit Egress + SeaweedFS S3 environment"
---

# Phase 7: Polish & Extended Features Verification Report

**Phase Goal:** Платформа получает управление спринтами, интеграцию с GitHub, запись звонков, интернационализацию, темы и операционное укрепление, углубляющие каждый модуль за пределы базовых возможностей

**Verified:** 2026-03-21T12:00:00Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (Success Criteria from ROADMAP.md)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Scrum-команды могут создавать спринты, перетаскивать задачи из бэклога, проводить Planning Poker и просматривать графики burndown/velocity | VERIFIED | `backend/internal/sprint/service.go`: CreateSprint, StartSprint, CompleteSprint (auto-carry SCRM-04), GetBurndown, GetVelocity; `backend/internal/sprint/poker.go`: StartPokerSession, CastVote (Fibonacci + hidden), RevealVotes; frontend: sprint-board.tsx, sprint-backlog.tsx, burndown-chart.tsx (93 lines), velocity-chart.tsx, planning-poker.tsx |
| 2 | PR/коммиты GitHub автоматически связываются с задачами по упоминанию ID, статус PR отображается на карточках задач, и мерж PR может автоматически переводить задачу в другой статус | VERIFIED | `backend/internal/github/service.go`: issueKeyRegex `\b([A-Z]+-\d+)\b`, handlePushWebhook, handlePRWebhook with auto-status on merge (GHUB-04), CreateBranch (GHUB-05); frontend: github-links.tsx, message-bubble.tsx, link-preview.tsx; note: webhook signature validation present in ValidateWebhookSignature function but handler does not enforce it for all requests |
| 3 | Пользователи могут записывать звонки (аудио + демонстрация экрана), записи сохраняются в проекте, и ссылка публикуется в канал | VERIFIED | `backend/internal/calls/recording.go`: StartRecording calls StartRoomCompositeEgress with S3Upload (S3 path: recordings/{projectID}/{callID}.mp4), publishes NATS events; OnEgressEnded webhook handler; ListProjectRecordings; `backend/internal/calls/scheduled.go`: rrule-go for CALL-02; frontend: recording-controls.tsx, recordings-list.tsx, scheduled-calls.tsx; weekly ticker in main.go |
| 4 | Приложение работает на русском и английском языках с переключением тем (тёмная/светлая/системная + акцентный цвет) и базовой возможностью установки PWA | VERIFIED | `frontend/src/i18n/index.ts`: 14 JSON locale files (7 RU + 7 EN namespaces), imported in main.tsx; `frontend/src/components/theme-provider.tsx`: next-themes ThemeProvider + AccentProvider with 5 accent colors via data-accent; `frontend/vite.config.ts`: VitePWA plugin present; `frontend/public/manifest.webmanifest`: exists |
| 5 | REST API документирован через Swagger, можно генерировать API-ключи, вебхуки срабатывают на события, и автоматические резервные копии создаются еженедельно | FAILED | API keys VERIFIED: crypto/rand + SHA-256 + APIKeyAuth middleware wired in main.go. Weekly backup ticker VERIFIED (7*24h goroutine). Swagger MISSING: no docs/ directory, no swag annotations, no /api/docs route. Webhook dispatcher STUB: Start() doesn't subscribe to NATS. JQL ORPHANED: parser exists, untested endpoint absent. PDF export DEFERRED. |

**Score: 3/5 truths verified** (Truth 1, 2, 3 VERIFIED; Truth 4 partially verified with human-verification items; Truth 5 FAILED)

---

## Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `backend/ent/schema/sprint.go` | Sprint Ent schema | VERIFIED | Contains func (Sprint) Fields with status enum, dates, retro_notes JSONB |
| `backend/internal/sprint/service.go` | SprintService with CRUD, start, complete, auto-carry | VERIFIED | 295 lines, NewSprintService, CreateSprint, StartSprint, CompleteSprint (auto-carry), CheckWIPLimit |
| `backend/internal/sprint/poker.go` | Planning Poker via Redis + NATS | VERIFIED | 185 lines, StartPokerSession, CastVote (Fibonacci validated), RevealVotes, GetSession |
| `frontend/src/modules/tracker/burndown-chart.tsx` | Recharts burndown chart | VERIFIED | 93 lines (min_lines: 40 satisfied), Recharts AreaChart with ideal/remaining lines |
| `backend/ent/schema/githublink.go` | GitHubLink Ent schema | VERIFIED | PR/commit link schema present |
| `backend/internal/github/service.go` | GitHubService webhook + API client | VERIFIED | 387 lines, NewGitHubService, HandleWebhook, CreateBranch, issueKeyRegex |
| `backend/internal/github/preview.go` | Link preview service | VERIFIED | GetLinkPreview with Redis caching and OG metadata |
| `frontend/src/modules/messenger/link-preview.tsx` | LinkPreview component | VERIFIED | getLinkPreview API call present; renders cards |
| `backend/ent/schema/recording.go` | Recording Ent schema | VERIFIED | egress_id, status enum, file_path, edges to Call/Project/User |
| `backend/ent/schema/scheduledcall.go` | ScheduledCall Ent schema with rrule | VERIFIED | rrule field, edges to Project/Channel |
| `backend/internal/calls/recording.go` | RecordingService with LiveKit Egress | VERIFIED | 267 lines, StartRecording/StopRecording/OnEgressEnded, StartRoomCompositeEgress + S3Upload |
| `frontend/src/i18n/index.ts` | i18next initialization RU/EN | VERIFIED | 65 lines (min_lines: 30 satisfied), 14 JSON imports, LanguageDetector |
| `frontend/src/components/theme-provider.tsx` | Theme switching | VERIFIED | 109 lines (min_lines: 30 satisfied), ThemeProvider + AccentProvider |
| `frontend/src/components/mobile-nav.tsx` | Bottom navigation for mobile | VERIFIED | File exists (created per summary) |
| `frontend/vite.config.ts` | PWA plugin configuration | VERIFIED | VitePWA import and config present |
| `backend/internal/apigateway/apikey.go` | API key auth middleware + CRUD | VERIFIED | CreateAPIKey (crypto/rand), APIKeyAuth middleware, VerifyAPIKey (SHA-256) |
| `backend/internal/apigateway/webhook.go` | Webhook dispatcher | STUB | Start() goroutine only waits for ctx.Done(); dispatchEvent() method exists but is NEVER CALLED — no JetStream subscription |
| `backend/internal/backup/service.go` | Backup service (pg_dump + mongodump + S3) | VERIFIED | PerformBackup with exec.Command("pg_dump"), exec.Command("mongodump"), RestoreBackup (partial: pg_restore/mongorestore from ZIP commented as `_ = zipData`) |
| `backend/internal/search/jql.go` | JQL parser and Ent predicate converter | ORPHANED | ParseJQL and ToEntPredicates implemented and tested but called by zero HTTP endpoints |
| `cli/main.go` | hub-cli binary entry point | VERIFIED | 16 lines, calls cmd.Execute(); push_results.go, create_bug.go, update_task.go exist in cli/cmd/ |
| `backend/internal/tms/environment.go` | TMS Environment CRUD | VERIFIED | CreateEnvironment, ListEnvironments, UpdateEnvironment, DeleteEnvironment + HTTP handlers |
| `backend/internal/project/templates.go` | Project template service | VERIFIED | ListTemplates, SaveProjectAsTemplate, SeedSystemTemplates (4 system templates: Empty, Software Dev, QA, Kanban Simple) |
| `docs/phase-07/business.md` | Business documentation | VERIFIED | File exists |
| `docs/phase-07/testing.md` | Testing documentation | VERIFIED | File exists |
| `docs/phase-07/technical.md` | Technical documentation | VERIFIED | File exists |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `backend/internal/sprint/service.go` | `backend/internal/tracker/service.go` | WIPChecker interface via SetWIPChecker | VERIFIED | CheckWIPLimit called from SprintService; WIPChecker interface in tracker |
| `frontend/src/modules/tracker/sprint-board.tsx` | `/api/projects/:id/sprints` | Zustand store fetchSprints -> sprintApi.listSprints | VERIFIED | useSprintStore imported; fetchSprints called in useEffect |
| `backend/internal/sprint/poker.go` | Redis + NATS | Redis hashes `poker:{sessionID}` + NATS Publish | VERIFIED | p.redis.HSet, p.redis.Expire, p.nats.Publish("poker.vote_cast"), p.nats.Publish("poker.revealed") |
| `backend/internal/github/handler.go` | `/api/github/webhook` | POST endpoint for webhook receiver | VERIFIED | HandleWebhook route registered; eventType routed to handlePushWebhook/handlePRWebhook |
| `backend/internal/github/service.go` | `backend/internal/tracker/service.go` | FindIssueByKey + UpdateIssueStatus | VERIFIED | repo.FindIssueByKey, repo.UpdateIssueStatus called in handlePRWebhook |
| `frontend/src/modules/messenger/link-preview.tsx` | `/api/link-preview` | getLinkPreview(url) -> fetch | VERIFIED | `const result = await getLinkPreview(url)` called in component |
| `backend/internal/calls/recording.go` | LiveKit Egress API | lksdk.EgressClient.StartRoomCompositeEgress | VERIFIED | egressInfo, err := s.egressClient.StartRoomCompositeEgress(ctx, egressReq) |
| `backend/internal/calls/recording.go` | SeaweedFS S3 | S3Upload in Egress output config | VERIFIED | livekit.S3Upload with Endpoint, Bucket, AccessKey, SecretKey; path: recordings/{projectID}/{callID}.mp4 |
| `frontend/src/i18n/index.ts` | `frontend/src/main.tsx` | import i18n in app entry point | VERIFIED | `import "./i18n";` at line 5 of main.tsx |
| `frontend/src/components/theme-provider.tsx` | `frontend/src/app.css` | CSS variables + data-accent attribute | VERIFIED | ThemeProvider wraps app; data-accent set on document.documentElement |
| `backend/internal/apigateway/apikey.go` | `backend/cmd/server/main.go` | APIKeyAuth middleware registered | VERIFIED | `app.Use(apigatewaymod.APIKeyAuth(apiKeySvc))` at line 315 of main.go |
| `backend/internal/apigateway/webhook.go` | NATS JetStream | JetStream consumer dispatches to endpoints | NOT_WIRED | Start() only contains `go func() { <-ctx.Done() }()` — no JetStream Subscribe, no consumer, dispatchEvent() is never triggered |
| `backend/internal/search/jql.go` | Tracker HTTP handler | GET /api/projects/:id/issues/jql endpoint | NOT_WIRED | ParseJQL/ToEntPredicates called by zero files outside jql.go and jql_test.go |
| `backend/internal/backup/service.go` | weekly ticker in main.go | goroutine with 7*24h ticker | VERIFIED | `ticker := time.NewTicker(7 * 24 * time.Hour)` at line 326 of main.go |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| PROJ-05 | 07-05 | Project templates (Empty/SoftwareDev/QA/KanbanSimple + custom) | SATISFIED | SeedSystemTemplates seeds 4 system templates; SaveProjectAsTemplate available |
| KANB-02 | 07-01 | WIP-limits on Kanban columns | SATISFIED | wip_limit field in WorkflowStatus schema; CheckWIPLimit in SprintService; frontend WIP display |
| KANB-03 | 07-01 | Swimlanes (assignee/priority/type) | SATISFIED | swimlane_mode enum in Board schema; frontend-only grouping in kanban-board.tsx |
| SCRM-01 | 07-01 | Sprint CRUD, backlog drag-drop, start/complete | SATISFIED | Full sprint lifecycle with sprint-backlog.tsx drag-drop |
| SCRM-02 | 07-01 | Burndown and velocity charts | SATISFIED | GetBurndown (linear interpolation), GetVelocity; burndown-chart.tsx (93 lines), velocity-chart.tsx |
| SCRM-03 | 07-01 | Sprint retrospective notes | SATISFIED | UpdateRetroNotes stores JSONB; sprint-retro.tsx (3 sections) |
| SCRM-04 | 07-01 | Auto-carry incomplete issues on sprint complete | SATISFIED | CompleteSprint finds incomplete issues, FindOrCreateNextPlanningSprint, MoveIssuesToSprint |
| SCRM-05 | 07-01 | Planning Poker with Fibonacci, hidden until reveal | SATISFIED | fibonacciValues map; CastVote publishes event without value; RevealVotes publishes full votes |
| MSG-11 | 07-02 | Link previews in messenger (task/testcase/GitHub/external OG) | SATISFIED | LinkPreviewService with Redis cache; message-bubble.tsx extracts URLs and renders LinkPreview |
| TMS-07 | 07-05 | Test environments (dev/staging/prod/custom, config, history filter) | SATISFIED | Environment Ent schema with env_label enum + config JSON; full CRUD in environment.go |
| CALL-02 | 07-03 | Scheduled and recurring calls with rrule | SATISFIED | ScheduledCallService with rrule-go; ProcessDueScheduledCalls cron (60s ticker); RRULE next-occurrence calculation |
| CALL-04 | 07-03 | Recording (audio+screen), all notified, saved in SeaweedFS, link in channel | SATISFIED | StartRecording with LiveKit Egress + S3Upload; NATS event "calls.recording.started"; OnEgressEnded publishes ready event |
| CALL-05 | 07-03 | Recordings section in project | SATISFIED | ListProjectRecordings; recordings-list.tsx with date filter + download |
| GHUB-01 | 07-02 | PR and commit links to tasks with preview cards | SATISFIED | handlePRWebhook creates GitHubLink; github-links.tsx shows PR status badges |
| GHUB-02 | 07-02 | Auto-link: "HUB-123" in commit message auto-attaches | SATISFIED | issueKeyRegex `\b([A-Z]+-\d+)\b` in handlePushWebhook |
| GHUB-03 | 07-02 | PR status on task (open/merged/closed) | SATISFIED | link_type + status enum in GitHubLink; github-links.tsx status badges |
| GHUB-04 | 07-02 | Auto-status change on PR merge | SATISFIED | closed+merged branch in handlePRWebhook calls repo.UpdateIssueStatus with integration.MergeStatusID |
| GHUB-05 | 07-02 | Branch creation from task (hub-123-fix-login naming) | SATISFIED | CreateBranch uses go-github client, slugify(issueKey+title), creates ref via Git.CreateRef |
| SRCH-03 | 07-05 | JQL-like query language for advanced filtering | BLOCKED | JQL parser (ParseJQL, ToEntPredicates) is implemented and unit-tested but not reachable via any HTTP endpoint; no /issues/jql route registered in any handler |
| API-01 | 07-05 | REST API with Swagger documentation | BLOCKED | Deferred per 07-05-SUMMARY; no swag annotations added, no backend/docs/ directory, no /api/docs route |
| API-02 | 07-05 | API key generation | SATISFIED | CreateAPIKey (crypto/rand 32 bytes, base64url, SHA-256 hash); APIKeyAuth middleware; frontend api-keys.tsx |
| API-03 | 07-05 | Webhooks (POST to external URL on events) | BLOCKED | WebhookDispatcher.Start() is a no-op goroutine; dispatchEvent() exists but is never called from any event source; webhooks will never fire |
| API-04 | 07-05 | hub-cli (push-results, create-bug, update-task) | SATISFIED | cli/main.go entry; cli/cmd/{push_results,create_bug,update_task}.go exist; cobra with --api-key |
| IMEX-01 | 07-05 | Universal CSV/JSON issue import with field mapping | SATISFIED | ImportIssuesCSV with preview mode + field mapping; ImportIssuesJSON |
| IMEX-03 | 07-05 | Export: issues (CSV/Excel), test runs (PDF), sprints (PDF), project ZIP | PARTIAL | CSV/Excel export for issues implemented; full project ZIP implemented; PDF export for test runs and sprints explicitly deferred in summary (maroto not added) |
| BKUP-01 | 07-05 | Weekly automatic backup (PostgreSQL + MongoDB + SeaweedFS) | SATISFIED | 7*24h ticker goroutine in main.go calls PerformBackup; pg_dump + mongodump + ZIP |
| BKUP-02 | 07-05 | Keep last N backups (configurable) | SATISFIED | cleanupOldBackups with maxBackups=4; runs after each backup |
| BKUP-03 | 07-05 | Restore from backup via admin panel | PARTIAL | RestoreBackup downloads ZIP from S3 but pg_restore/mongorestore calls are stubbed (`_ = zipData`); restore does not actually execute against databases |
| UI-02 | 07-04 | Mobile layout: bottom nav, simplified Kanban, mobile dashboards | SATISFIED | mobile-nav.tsx (5 tabs); mobile-layout.tsx with useIsMobile hook; slide-over sidebar |
| UI-03 | 07-04 | Dark/light/system theme + custom accent color | SATISFIED | ThemeProvider (next-themes) with attribute="class"; AccentProvider with 5 colors via data-accent |
| UI-04 | 07-04 | Russian and English with profile switcher | SATISFIED | i18next with 14 JSON files; language-switcher.tsx in topbar and settings |
| UI-06 | 07-04 | PWA: install on phone, offline cache, Web Push | SATISFIED (basic) | VitePWA in vite.config.ts; manifest.webmanifest exists; service worker generated; Web Push not implemented (noted as optional) |

---

## Anti-Patterns Found

| File | Pattern | Severity | Impact |
|------|---------|----------|--------|
| `backend/internal/apigateway/webhook.go` | `Start()` goroutine only waits for ctx.Done() — no JetStream subscription, dispatchEvent() unreachable | Blocker | API-03 requirement effectively not met: webhooks will never fire on any platform event |
| `backend/internal/backup/service.go:144` | `_ = zipData // pg_restore + mongorestore из ZIP` — restore downloads backup but discards data without executing restore commands | Blocker | BKUP-03 restore does not actually restore databases |
| `backend/internal/search/jql.go` | ParseJQL/ToEntPredicates exist but are never called from any handler (only from jql_test.go) | Blocker | SRCH-03 not reachable by API consumers |
| `backend/internal/github/handler.go:61-63` | `_ = signature` — webhook signature validation bypassed in handler even though ValidateWebhookSignature function exists | Warning | Security gap: any actor can send forged webhook events |

---

## Human Verification Required

### 1. Planning Poker Real-Time Flow

**Test:** Open two browser windows as different users. Start a poker session from one window. Vote from both. Observe that vote values are hidden. Click Reveal.
**Expected:** Both windows update in real time; votes hidden until reveal; after reveal all values and stats visible.
**Why human:** Real-time WS/NATS flow cannot be verified statically.

### 2. GitHub Webhook Auto-Linking (Live)

**Test:** Configure a GitHub integration with a real repository. Push a commit with "HUB-123" in the message.
**Expected:** A GitHubLink entry appears on the HUB-123 task in the UI.
**Why human:** Requires live GitHub webhook delivery to localhost/production URL.

### 3. PWA Install Prompt

**Test:** Open the app in Chrome on mobile (HTTPS required). Trigger install prompt.
**Expected:** Browser shows install banner; app opens in standalone mode after install.
**Why human:** PWA installability depends on browser heuristics, HTTPS environment, and real mobile device.

### 4. Language Switching Visual Coverage

**Test:** Switch language from RU to EN in Settings > Appearance. Navigate through all major modules.
**Expected:** All visible UI labels in sidebar, topbar, tracker, messenger, TMS, calls, and settings are in English.
**Why human:** Not all components may have been migrated to use t(). The plan explicitly scoped this to "sidebar, topbar, settings, login and core buttons" — other modules need visual inspection.

### 5. Call Recording Download

**Test:** Start a call, click Record, stop recording, wait for processing, navigate to project Recordings section, click Download.
**Expected:** File downloads correctly with valid audio content.
**Why human:** Requires live LiveKit Egress integration + SeaweedFS S3.

---

## Gaps Summary

Three functional gaps prevent full goal achievement:

**1. Swagger UI (API-01) — MISSING**
The 07-05-SUMMARY explicitly notes Swagger was deferred because it "requires swag CLI tool installation and annotation of all existing handlers." No docs/ directory exists, no swag annotations are present anywhere in the codebase, and no /api/docs route is registered. API-01 is completely absent.

**2. Webhook Dispatcher (API-03) — STUB**
The WebhookDispatcher.Start() method creates a goroutine that solely waits for context cancellation. It does not subscribe to NATS JetStream, does not consume any events, and consequently dispatchEvent() is never called. The delivery logic, HMAC signing, and retry mechanism all exist and are well-implemented — but they have no trigger. Webhooks will never fire regardless of what events occur on the platform.

**3. JQL HTTP Endpoint (SRCH-03) — ORPHANED**
The JQL recursive descent parser (ParseJQL) and Ent predicate converter (ToEntPredicates) are fully implemented and tested. However, they are called by zero production code. No HTTP endpoint registers a JQL route in the tracker handler, search handler, or main.go. The SRCH-03 requirement is not accessible to API consumers.

**Secondary gaps (partial):**

**4. PDF Export (IMEX-03) — PARTIAL**
Issue CSV/Excel/ZIP export works. Test run PDF and sprint PDF export were deferred (maroto v2 not added to go.mod). The IMEX-03 requirement is partially met.

**5. Backup Restore (BKUP-03) — STUB**
RestoreBackup downloads the ZIP from S3 and creates a safety-net backup, but the actual pg_restore and mongorestore commands are not implemented (data is read then discarded with `_ = zipData`). The admin UI for restore exists, but clicking Restore does not actually restore databases.

---

_Verified: 2026-03-21T12:00:00Z_
_Verifier: Claude (gsd-verifier)_
