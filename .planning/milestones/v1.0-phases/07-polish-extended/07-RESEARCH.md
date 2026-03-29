# Phase 7: Polish & Extended Features - Research

**Researched:** 2026-03-21
**Domain:** Sprints, GitHub integration, call recording, i18n, theming, PWA, REST API, backups, import/export, Kanban enhancements
**Confidence:** HIGH

## Summary

Phase 7 is the final phase of the Hub project, adding depth to all existing modules rather than creating new ones. It covers 32 requirements across 10 functional areas: Scrum/sprints (SCRM-01..05), Kanban enhancements (KANB-02..03), GitHub integration (GHUB-01..05), call recording and scheduling (CALL-02, CALL-04..05), messenger link previews (MSG-11), TMS test environments (TMS-07), JQL-like search (SRCH-03), public REST API with Swagger (API-01..04), import/export (IMEX-01, IMEX-03), backups (BKUP-01..03), project templates (PROJ-05), and UI polish -- i18n, theming, mobile layout, PWA (UI-02..04, UI-06).

The codebase already has all core modules (auth, tracker, messenger, TMS, calls, dashboards, notifications, search, audit) operational from Phases 1-6. Phase 7 extends these modules with new Ent schemas (Sprint, ScheduledCall, GitHubLink, ProjectTemplate, APIKey, Webhook, BackupRecord, Environment), new backend services (sprint, github, recording, backup, api-keys, webhook dispatcher, import/export), and frontend enhancements (i18n provider, theme provider, PWA manifest, mobile responsive layouts, Swagger UI mount). The architecture patterns are well-established -- handler/service/repository layers, NATS events, outbox pattern, CombinedRepository for PG+Mongo -- and Phase 7 follows them without introducing new architectural concepts.

**Primary recommendation:** Decompose Phase 7 into 8-10 plans grouped by functional area (Scrum, GitHub, Recording, API/CLI, Import/Export, Backups, i18n/Theme/PWA, Kanban/UI polish), with i18n infrastructure in an early plan since it touches all existing UI code. Use the same handler/service/repository pattern established in Phases 1-6 for all new backend features.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PROJ-05 | Project templates: Empty, Software Dev, QA, Kanban Simple + custom + save as template | Ent schema ProjectTemplate with JSONB config snapshot; clone project structure on create |
| KANB-02 | WIP limits on Kanban columns | Add wip_limit field to WorkflowStatus Ent schema; enforce in MoveIssue service |
| KANB-03 | Swimlanes by assignee, priority, or type | Frontend-only grouping on board data; add swimlane_mode to Board schema |
| SCRM-01 | Sprint CRUD: create with dates, drag from backlog, start/complete | New Sprint Ent schema + SprintService; sprint_id optional FK on Issue |
| SCRM-02 | Burndown chart + Velocity chart | Raw SQL daily snapshot aggregation; Recharts on frontend |
| SCRM-03 | Sprint retrospective notes | JSONB retro_notes field on Sprint schema |
| SCRM-04 | Auto-carry incomplete tasks to next sprint | SprintService.CompleteSprint moves issues with non-done status |
| SCRM-05 | Planning Poker: Fibonacci, hidden votes until reveal | Real-time via NATS; PokerSession in-memory or Redis; WS events |
| MSG-11 | Link previews: tasks, test cases, runs, GitHub PR/Issue, external URL (OG meta) | otiai10/opengraph for OG parsing; internal entity preview via DB lookup |
| TMS-07 | Test environments: label (dev/staging/prod), configs (browser/OS), comparison, filtered history | New Environment Ent schema; environment_id FK on TestRun |
| CALL-02 | Scheduled + recurring calls (daily standup) | ScheduledCall Ent schema + teambition/rrule-go for recurrence; cron trigger |
| CALL-04 | Recording: audio + screen share, notification, save to SeaweedFS, link in channel | LiveKit Egress API via server-sdk-go EgressClient; S3 upload to SeaweedFS |
| CALL-05 | Recordings bound to project (Recordings section) | Recording Ent schema with project_id FK; list/download endpoints |
| GHUB-01 | PR/commit links on tasks with preview card | GitHubLink Ent schema; webhook handler parses payload |
| GHUB-02 | Auto-link: "HUB-123" in commit message links to task | Regex `\bHUB-\d+\b` on push webhook payload; word boundary matching |
| GHUB-03 | PR status on task (open/merged/closed) | Webhook updates GitHubLink.status on PR events |
| GHUB-04 | Auto-status change on PR merge | Configurable target status per project; WorkflowEngine.ValidateTransition |
| GHUB-05 | Create branch from task (hub-123-fix-login) | GitHub API via go-github; generate branch name from issue key + title slug |
| SRCH-03 | JQL-like query language | Custom lexer/parser in Go; AST to Ent predicates; subset of Jira JQL syntax |
| API-01 | REST API with Swagger documentation | swaggo/swag + gofiber/contrib/v3/swaggo; annotate existing handlers |
| API-02 | API key generation in user profile | APIKey Ent schema (hashed key); API key auth middleware alongside JWT |
| API-03 | Webhooks (POST to external URL on events) | WebhookEndpoint Ent schema; JetStream consumer dispatches HTTP POST |
| API-04 | hub-cli: CLI tool for CI/CD (push-results, create-bug, update-task) | Go CLI with cobra; authenticated via API key; HTTP client to Hub API |
| IMEX-01 | Universal CSV/JSON task import with field mapping | Import service with configurable field map; preview + confirm flow |
| IMEX-03 | Export: tasks (CSV/Excel), runs (PDF), sprints (PDF), full project backup (ZIP) | excelize v2 for Excel; maroto v2 for PDF (reuse from Phase 4); archive/zip |
| BKUP-01 | Weekly automatic backup (PG + MongoDB + SeaweedFS) | Go backup service: exec pg_dump + mongodump + SeaweedFS filer export; cron |
| BKUP-02 | Keep last N backups (configurable) | Retention policy in SystemSettings; cleanup after each backup |
| BKUP-03 | Restore from backup via admin panel | Admin endpoint: stop services, pg_restore + mongorestore; dangerous operation with confirmation |
| UI-02 | Mobile layout: bottom nav, simplified Kanban (list view), mobile dashboards, mobile calls | Tailwind responsive breakpoints; separate mobile components where needed |
| UI-03 | Dark/light/system theme + custom accent color | CSS variables in @theme; next-themes for switching; accent via data-attribute |
| UI-04 | Russian + English with profile switch | i18next + react-i18next; namespace-per-module JSON files; backend error messages stay Russian |
| UI-06 | PWA: install on phone, offline cache, Web Push notifications | vite-plugin-pwa 1.x; Workbox generateSW; Web Push via notification service |
</phase_requirements>

## Standard Stack

### Core (Phase 7 additions to existing stack)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| i18next | 25.x | i18n core | De facto standard for React i18n; namespace splitting, lazy loading, interpolation |
| react-i18next | 15.x | React i18n bindings | useTranslation hook; React 19 compatible; Trans component for JSX interpolation |
| i18next-browser-languagedetector | 8.x | Auto-detect browser language | Standard companion; detects from navigator, cookie, localStorage |
| vite-plugin-pwa | 1.x | PWA service worker + manifest | Zero-config Workbox integration for Vite; generates SW and manifest |
| next-themes | 0.4.x | Theme switching (dark/light/system) | Works with Tailwind v4 CSS variables; SSR-safe; localStorage persistence |
| swaggo/swag | latest | Swagger doc generation | Standard Go Swagger generator; annotates handlers with comments |
| gofiber/contrib/v3/swaggo | latest | Fiber Swagger UI mount | Official Fiber v3 Swagger integration from contrib package |
| otiai10/opengraph/v2 | 2.x | OG meta tag parser | Fetch + parse OpenGraph data from URLs for link previews |
| teambition/rrule-go | latest | RRULE recurrence rules | RFC 5545 compliant; iCalendar rrule parsing for recurring calls |
| google/go-github/v68 | 68.x | GitHub API client | Official Go client for GitHub REST API; branch creation, PR queries |
| cobra | 1.x | CLI framework | Standard Go CLI framework for hub-cli tool |
| excelize/v2 | 2.x | Excel read/write | Already used in Phase 4; reuse for task export |
| maroto/v2 | 2.x | PDF generation | Already used in Phase 4; reuse for sprint/run reports |

### Supporting (already in project)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Recharts | (via Tremor) | Charts | Burndown + velocity charts; already used in Phase 6 dashboards |
| NATS + JetStream | 2.11.x | Events | Planning Poker real-time; webhook dispatch; recording events |
| LiveKit server-sdk-go/v2 | 2.x | Egress API | StartRoomCompositeEgress / StopEgress for call recording |
| aws-sdk-go-v2 | latest | S3 operations | Recording file storage in SeaweedFS; backup file storage |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| i18next | lingui | lingui is lighter but smaller ecosystem; i18next has better tooling and React 19 support |
| next-themes | Manual CSS class toggling | next-themes handles system preference, flash prevention, SSR; not worth hand-rolling |
| otiai10/opengraph | dyatlov/go-opengraph | dyatlov parses more OG properties but otiai10 includes HTTP fetch; simpler API |
| cobra (hub-cli) | urfave/cli | Both work; cobra has better subcommand support and auto-generated help |
| Custom JQL parser | ANTLR grammar | ANTLR is overkill for a subset JQL; hand-written recursive descent parser is simpler and faster |

**Installation:**
```bash
# Frontend additions
npm install i18next react-i18next i18next-browser-languagedetector
npm install next-themes
npm install -D vite-plugin-pwa

# Backend additions
go get github.com/swaggo/swag/cmd/swag@latest
go get github.com/gofiber/contrib/v3/swaggo
go get github.com/otiai10/opengraph/v2
go get github.com/teambition/rrule-go
go get github.com/google/go-github/v68
go get github.com/spf13/cobra
```

## Architecture Patterns

### New Ent Schemas Required

```
backend/ent/schema/
  sprint.go              # SCRM-01: Sprint entity with project, dates, status
  scheduledcall.go       # CALL-02: Scheduled/recurring call with rrule
  recording.go           # CALL-04/05: Recording file with project + call FK
  githublink.go          # GHUB-01..04: PR/commit link to issue
  githubintegration.go   # GHUB: Repo connection config per project
  projecttemplate.go     # PROJ-05: Template with JSONB config snapshot
  apikey.go              # API-02: Hashed API key with user FK
  webhookendpoint.go     # API-03: Webhook URL + events + secret
  webhookdelivery.go     # API-03: Delivery log (attempts, status)
  backuprecord.go        # BKUP-01..03: Backup metadata (path, size, date)
  environment.go         # TMS-07: Test environment config
```

### New Backend Modules

```
backend/internal/
  sprint/                # SprintService + SprintHandler + SprintRepository
  github/                # GitHubService (webhook handler + API client)
  recording/             # RecordingService (LiveKit Egress + storage)
  apigateway/            # API key auth middleware + webhook dispatcher
  backup/                # BackupService (pg_dump + mongodump + retention)
  imex/                  # ImportExportService (CSV/JSON/Excel/ZIP)
```

### Frontend Structure for i18n

```
frontend/src/
  i18n/
    index.ts             # i18next init config
    locales/
      ru/
        common.json      # Shared translations (buttons, errors, navigation)
        tracker.json     # Tracker module strings
        messenger.json   # Messenger module strings
        tms.json         # TMS module strings
        calls.json       # Calls module strings
        dashboard.json   # Dashboard strings
        settings.json    # Settings strings
      en/
        common.json
        tracker.json
        messenger.json
        tms.json
        calls.json
        dashboard.json
        settings.json
```

### Pattern 1: Sprint Lifecycle

**What:** Sprint entity with state machine (planning -> active -> completed)
**When to use:** SCRM-01, SCRM-04

```go
// Спринт: создание -> запуск -> завершение с автопереносом
type SprintService struct {
    repo    SprintRepository
    tracker *TrackerService // для перемещения задач
    nats    *natsclient.Client
}

func (s *SprintService) CompleteSprint(ctx context.Context, sprintID uuid.UUID) error {
    sprint, _ := s.repo.GetSprint(ctx, sprintID)
    // Находим незавершённые задачи
    incomplete := s.repo.FindIncompleteIssues(ctx, sprintID)
    // Ищем или создаём следующий спринт
    nextSprint := s.repo.FindOrCreateNextSprint(ctx, sprint.ProjectID)
    // Переносим задачи
    s.repo.MoveIssuesToSprint(ctx, incomplete, nextSprint.ID)
    // Обновляем статус
    s.repo.UpdateSprintStatus(ctx, sprintID, "completed")
    return nil
}
```

### Pattern 2: GitHub Webhook Handler

**What:** Receive GitHub webhooks, parse events, link to tasks
**When to use:** GHUB-01..04

```go
// GitHub webhook: валидация подписи + маршрутизация событий
func (h *GitHubHandler) HandleWebhook(c fiber.Ctx) error {
    payload := c.Body()
    signature := c.Get("X-Hub-Signature-256")
    // Валидируем HMAC-SHA256
    if !validateSignature(payload, signature, h.webhookSecret) {
        return fiber.ErrUnauthorized
    }
    eventType := c.Get("X-GitHub-Event")
    switch eventType {
    case "push":
        return h.handlePush(c.Context(), payload)
    case "pull_request":
        return h.handlePullRequest(c.Context(), payload)
    }
    return c.SendStatus(204)
}

func (h *GitHubHandler) handlePush(ctx context.Context, payload []byte) error {
    var event PushEvent
    json.Unmarshal(payload, &event)
    // Regex для нахождения task keys в коммит-сообщениях
    re := regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
    for _, commit := range event.Commits {
        matches := re.FindAllString(commit.Message, -1)
        for _, key := range matches {
            // Проверяем что задача существует
            h.service.LinkCommitToIssue(ctx, key, commit)
        }
    }
    return nil
}
```

### Pattern 3: LiveKit Egress Recording

**What:** Start/stop room recording via LiveKit Egress API
**When to use:** CALL-04, CALL-05

```go
// Запись звонка через LiveKit Egress
func (s *RecordingService) StartRecording(ctx context.Context, callID uuid.UUID, userID uuid.UUID) (*RecordingResponse, error) {
    call, _ := s.callRepo.GetCall(ctx, callID)
    // Проверяем лимит одновременных записей (макс 2)
    activeCount := s.repo.CountActiveRecordings(ctx)
    if activeCount >= s.maxConcurrent {
        return nil, errors.New("превышен лимит одновременных записей")
    }
    // Запускаем RoomComposite Egress
    fileOutput := &livekit.EncodedFileOutput{
        FileType: livekit.EncodedFileType_MP4,
        Filepath: fmt.Sprintf("recordings/%s/%s.mp4", call.ChannelID, callID),
        Output: &livekit.EncodedFileOutput_S3{
            S3: &livekit.S3Upload{
                AccessKey: s.s3Config.AccessKey,
                Secret:    s.s3Config.SecretKey,
                Endpoint:  s.s3Config.Endpoint,
                Bucket:    s.s3Config.Bucket,
            },
        },
    }
    info, err := s.egressClient.StartRoomCompositeEgress(ctx, &livekit.RoomCompositeEgressRequest{
        RoomName:  call.RoomName,
        AudioOnly: false, // аудио + screen share
        Output: &livekit.RoomCompositeEgressRequest_File{
            File: fileOutput,
        },
    })
    // Уведомляем всех участников
    s.nats.Publish("calls.recording.started", ...)
    // Сохраняем запись в БД
    return s.repo.CreateRecording(ctx, callID, userID, info.EgressId)
}
```

### Pattern 4: i18n Setup

**What:** Initialize i18next with namespace-per-module structure
**When to use:** UI-04

```typescript
// frontend/src/i18n/index.ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import ruCommon from './locales/ru/common.json';
import enCommon from './locales/en/common.json';
// ... остальные неймспейсы

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      ru: { common: ruCommon, tracker: ruTracker, ... },
      en: { common: enCommon, tracker: enTracker, ... },
    },
    fallbackLng: 'ru',
    defaultNS: 'common',
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  });
```

### Pattern 5: Theme System with Tailwind v4

**What:** CSS variables for dark/light + accent color
**When to use:** UI-03

```css
/* frontend/src/app.css */
@theme {
  --color-primary: var(--accent);
  --color-primary-foreground: var(--accent-foreground);
  --color-background: var(--bg);
  --color-foreground: var(--fg);
}

:root {
  --accent: oklch(0.637 0.237 25.331);  /* default red accent */
  --accent-foreground: #ffffff;
  --bg: #ffffff;
  --fg: #0a0a0a;
}

.dark {
  --bg: #0a0a0a;
  --fg: #fafafa;
}

/* Accent overrides via data attribute */
[data-accent="blue"]   { --accent: oklch(0.623 0.214 259.815); }
[data-accent="green"]  { --accent: oklch(0.723 0.219 149.579); }
[data-accent="purple"] { --accent: oklch(0.627 0.265 303.9); }
```

### Pattern 6: Backup Service

**What:** Automated PG + MongoDB + SeaweedFS backup
**When to use:** BKUP-01..03

```go
// BackupService выполняет резервное копирование всех хранилищ.
func (s *BackupService) PerformBackup(ctx context.Context) (*BackupRecord, error) {
    timestamp := time.Now().Format("20060102-150405")
    dir := fmt.Sprintf("/tmp/hub-backup-%s", timestamp)
    os.MkdirAll(dir, 0755)

    // 1. PostgreSQL
    pgFile := filepath.Join(dir, "postgres.sql.gz")
    cmd := exec.CommandContext(ctx, "pg_dump", s.pgURL, "--format=custom", "--file="+pgFile)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("pg_dump failed: %w", err)
    }

    // 2. MongoDB
    mongoDir := filepath.Join(dir, "mongo")
    cmd = exec.CommandContext(ctx, "mongodump", "--uri="+s.mongoURL, "--out="+mongoDir)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("mongodump failed: %w", err)
    }

    // 3. SeaweedFS (S3 sync)
    // Используем aws-sdk-go-v2 для скачивания всех объектов бакета

    // 4. Создаём ZIP архив и загружаем в SeaweedFS (backup bucket)
    zipPath := dir + ".zip"
    createZipArchive(dir, zipPath)
    s3Key := fmt.Sprintf("backups/%s.zip", timestamp)
    s.uploadToS3(ctx, zipPath, s3Key)

    // 5. Очистка старых бэкапов (retention)
    s.cleanupOldBackups(ctx)

    return s.repo.CreateBackupRecord(ctx, s3Key, fileSize(zipPath))
}
```

### Anti-Patterns to Avoid

- **i18n: Do NOT translate backend error messages.** Backend continues returning Russian error strings (as established in Phases 1-6). Only UI labels, buttons, and navigation are translated. Changing backend errors would break existing error handling.
- **Swagger: Do NOT generate OpenAPI spec from code.** Use swag comments on handlers; let `swag init` generate the spec. Do not maintain a separate OpenAPI YAML.
- **Recording: Do NOT proxy recording files through Go backend.** Use presigned URLs from SeaweedFS for download, same as all other files in the project.
- **JQL: Do NOT implement full Jira JQL.** Support a subset: `status = "Done" AND assignee = currentUser() AND priority IN ("High", "Critical") ORDER BY created DESC`. No functions beyond `currentUser()`, no nested clauses.
- **Backup: Do NOT run pg_dump/mongodump inside the main Go process.** Use `exec.Command` to call external tools. The backup process should be a separate goroutine or even a separate container in K3s.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Recurring events | Custom recurrence logic | teambition/rrule-go | RFC 5545 RRULE is deceptively complex (timezone edge cases, exclusion dates) |
| OG meta parsing | Custom HTML parser | otiai10/opengraph/v2 | OG spec has many edge cases (missing tags, redirects, encoding) |
| Theme switching | Manual class/localStorage | next-themes | Handles system preference, flash prevention, SSR hydration mismatch |
| PWA manifest + SW | Manual service worker | vite-plugin-pwa | Workbox caching strategies are complex; plugin handles precache manifest |
| Swagger generation | Manual OpenAPI YAML | swaggo/swag | Annotations stay near code; auto-generates spec on `swag init` |
| CLI framework | Manual flag parsing | spf13/cobra | Subcommands, auto-generated help, shell completion |
| GitHub API | Raw HTTP calls | google/go-github | Handles auth, pagination, rate limiting, type safety |
| Excel generation | Manual XLSX bytes | excelize/v2 | Already used in Phase 4; handles styles, sheets, formulas |
| PDF reports | Manual PDF | maroto/v2 | Already used in Phase 4; grid layout, headers, tables |

**Key insight:** Phase 7 is a breadth phase -- many small features touching different domains. The risk is not architectural complexity but time spent on solved problems. Use existing libraries aggressively.

## Common Pitfalls

### Pitfall 1: LiveKit Egress Resource Starvation
**What goes wrong:** Starting a RoomComposite recording consumes 2-6 CPUs and 4GB+ RAM (headless Chrome). On a budget K3s cluster, this starves other services.
**Why it happens:** Egress records by rendering the room in a headless Chrome browser. Developers test with 1 recording and don't discover resource impact.
**How to avoid:**
- Deploy Egress as separate service on dedicated node
- Limit concurrent recordings to 1-2 max
- Add `SYS_ADMIN` capability to Egress pod (required since v1.7.6 for Chrome sandboxing)
- Monitor Egress status; auto-terminate recordings > 2 hours
**Warning signs:** Other services OOMKilled; recordings produce empty files; silent recording failures

### Pitfall 2: GitHub Webhook Task Key False Matches
**What goes wrong:** Regex `HUB-123` matches inside words like `THUB-123` or matches `HUB-12345` when only `HUB-1234` exists.
**Why it happens:** Missing word boundary in regex; no validation that matched key actually exists in the database.
**How to avoid:**
- Use `\b([A-Z]+-\d+)\b` with word boundaries
- Validate each matched key exists via database lookup before creating link
- Handle multiple project prefixes (not just "HUB")
**Warning signs:** Phantom links appearing on tasks; links to nonexistent tasks

### Pitfall 3: i18n Key Drift
**What goes wrong:** Russian and English translation files diverge -- keys added to Russian but missing from English, or vice versa.
**Why it happens:** Manual JSON file management without validation.
**How to avoid:**
- i18next `saveMissing` option in dev to detect missing keys
- CI check that all namespaces have identical key sets in both languages
- Fallback to Russian (`fallbackLng: 'ru'`) so missing English keys show Russian instead of key strings
**Warning signs:** Raw key strings (`tracker.issue.title`) visible in English UI

### Pitfall 4: Backup Restore Destroys Active Data
**What goes wrong:** Admin clicks "Restore" without understanding it replaces all current data.
**Why it happens:** No confirmation flow; restore is irreversible.
**How to avoid:**
- Double confirmation: "This will replace ALL data" + type project name to confirm
- Create a backup of current state before restoring
- Restore to a separate database first for verification (if feasible)
- Lock the system during restore (maintenance mode)
**Warning signs:** Accidental data loss; users losing recent work after restore

### Pitfall 5: Planning Poker State Management
**What goes wrong:** Real-time voting with hidden votes requires synchronized state across all participants. Race conditions when revealing votes.
**Why it happens:** Distributed real-time state without proper synchronization.
**How to avoid:**
- Use Redis for poker session state (votes, revealed flag)
- Single atomic operation to reveal all votes (Redis MULTI/EXEC)
- NATS events for real-time UI updates
- Session timeout (auto-reveal after 5 minutes of inactivity)
**Warning signs:** Votes visible before reveal; desynchronized vote counts; stuck sessions

### Pitfall 6: PWA Offline Cache Serving Stale API Data
**What goes wrong:** Service worker caches API responses; user sees outdated data after reconnecting.
**Why it happens:** Default Workbox caching strategy caches everything including API calls.
**How to avoid:**
- Use `generateSW` with `navigateFallback` for app shell only
- Exclude `/api/*` from service worker caching (network-only for API)
- Cache only static assets (JS, CSS, images, fonts)
- Use TanStack Query's built-in cache for API data (not SW cache)
**Warning signs:** Stale data after going online; API errors not surfacing; duplicate notifications

## Code Examples

### JQL-like Query Parser (SRCH-03)

```go
// Простой рекурсивный парсер для JQL-подобных запросов.
// Поддерживает: field = "value", field IN ("a", "b"), AND, OR, ORDER BY
type JQLParser struct {
    tokens []Token
    pos    int
}

type JQLFilter struct {
    Field    string
    Operator string   // =, !=, IN, NOT IN
    Values   []string
}

type JQLQuery struct {
    Filters  []JQLFilter
    Logic    string // AND, OR
    OrderBy  string
    OrderDir string // ASC, DESC
}

// ParseJQL разбирает строку запроса в структуру JQLQuery.
func ParseJQL(input string) (*JQLQuery, error) {
    tokens := tokenize(input)
    parser := &JQLParser{tokens: tokens}
    return parser.parse()
}

// ToEntPredicates конвертирует JQL в предикаты Ent ORM.
func (q *JQLQuery) ToEntPredicates() []predicate.Issue {
    var preds []predicate.Issue
    for _, f := range q.Filters {
        switch f.Field {
        case "status":
            // issue.HasStatusWith(workflowstatus.NameIn(f.Values...))
        case "assignee":
            if f.Values[0] == "currentUser()" {
                // Подставляем текущего пользователя из контекста
            }
        case "priority":
            // issue.PriorityIn(...)
        }
    }
    return preds
}
```

### API Key Authentication Middleware (API-02)

```go
// APIKeyAuth проверяет API ключ из заголовка X-API-Key.
// Работает параллельно с JWT auth -- если JWT отсутствует, проверяет API key.
func APIKeyAuth(repo APIKeyRepository) fiber.Handler {
    return func(c fiber.Ctx) error {
        apiKey := c.Get("X-API-Key")
        if apiKey == "" {
            return fiber.ErrUnauthorized
        }
        // Хешируем ключ для сравнения (ключи хранятся как SHA-256 хеш)
        hash := sha256Hash(apiKey)
        keyRecord, err := repo.FindByHash(c.Context(), hash)
        if err != nil {
            return fiber.ErrUnauthorized
        }
        // Обновляем last_used_at
        go repo.UpdateLastUsed(context.Background(), keyRecord.ID)
        // Устанавливаем userID в контекст (как JWT middleware)
        c.Locals("userID", keyRecord.UserID.String())
        return c.Next()
    }
}
```

### Webhook Dispatcher (API-03)

```go
// WebhookDispatcher слушает JetStream и рассылает вебхуки на внешние URL.
func (d *WebhookDispatcher) Start(ctx context.Context) error {
    consumer, _ := d.js.CreateOrUpdateConsumer(ctx, "EVENTS", jetstream.ConsumerConfig{
        Durable:   "webhook-dispatcher",
        AckPolicy: jetstream.AckExplicitPolicy,
    })
    cons, _ := consumer.Consume(func(msg jetstream.Msg) {
        endpoints := d.repo.FindEndpointsForEvent(ctx, msg.Subject())
        for _, ep := range endpoints {
            // Подписываем payload HMAC-SHA256
            signature := hmacSign(msg.Data(), ep.Secret)
            resp, err := d.httpClient.Post(ep.URL, msg.Data(), map[string]string{
                "X-Hub-Signature-256": signature,
                "X-Hub-Event":        msg.Subject(),
            })
            // Логируем доставку
            d.repo.LogDelivery(ctx, ep.ID, resp.StatusCode, err)
            if err == nil && resp.StatusCode < 300 {
                msg.Ack()
            }
            // Retry с exponential backoff при ошибке
        }
    })
    return nil
}
```

### WIP Limit Enforcement (KANB-02)

```go
// MoveIssue с проверкой WIP-лимита колонки.
func (s *TrackerService) MoveIssueWithWIP(ctx context.Context, issueKey string, userID uuid.UUID, req MoveIssueRequest) (*IssueResponse, error) {
    newStatusID, _ := uuid.Parse(req.StatusID)
    // Проверяем WIP лимит целевого статуса
    status, _ := s.repo.GetWorkflowStatus(ctx, newStatusID)
    if status.WIPLimit > 0 {
        currentCount := s.repo.CountIssuesByStatus(ctx, newStatusID)
        if currentCount >= status.WIPLimit {
            return nil, apperrors.NewConflict(
                fmt.Sprintf("WIP лимит колонки '%s' достигнут (%d/%d)", status.Name, currentCount, status.WIPLimit),
            )
        }
    }
    // Далее стандартная логика MoveIssue
    return s.MoveIssue(ctx, issueKey, userID, req)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| tailwind.config.js themes | CSS @theme + variables | Tailwind v4 (Jan 2025) | Theme switching via CSS only, no JS config |
| Manual service worker | vite-plugin-pwa + Workbox | vite-plugin-pwa 1.x (2025) | Zero-config PWA with precache manifest |
| framer-motion | motion (renamed) | 2024 | Import from `motion/react`, same API |
| swaggo/fiber-swagger | gofiber/contrib/v3/swaggo | Fiber v3 (2025) | Official Fiber contrib package for v3 |
| MinIO for backups | SeaweedFS | Feb 2026 | MinIO archived; SeaweedFS S3-compatible |

**Deprecated/outdated:**
- `framer-motion` package name: use `motion` instead
- `gofiber/swagger`: use `gofiber/contrib/v3/swaggo` for Fiber v3
- `tailwind.config.js`: CSS-first config in v4

## Open Questions

1. **JQL Syntax Scope**
   - What we know: Full Jira JQL is extremely complex (functions, nested clauses, field aliases)
   - What's unclear: Exact subset to support in v1
   - Recommendation: Start with `field OP value [AND|OR field OP value] [ORDER BY field ASC|DESC]`. Operators: `=`, `!=`, `IN`, `NOT IN`, `>`, `<`, `>=`, `<=`. Functions: only `currentUser()`. No nested parentheses in v1.

2. **Backup Restore Scope**
   - What we know: Full restore requires stopping all services and replacing databases
   - What's unclear: Whether partial restore (single project) is needed for v1
   - Recommendation: Full restore only in v1. Admin panel shows backup list, size, date. Restore requires maintenance mode. Single-project restore is v2.

3. **Hub-CLI Distribution**
   - What we know: CLI needs to be distributed to CI environments
   - What's unclear: Distribution method (Go binary releases vs npm package vs Docker image)
   - Recommendation: Go binary with goreleaser; publish to GitHub Releases. Users download binary or use `go install`. Docker image as alternative for CI.

4. **GitHub Integration Scope**
   - What we know: Need webhook receiver + optional GitHub API calls (branch creation)
   - What's unclear: Whether to support GitHub App (installation-level tokens) or just personal access tokens
   - Recommendation: Personal access tokens in v1 (simpler setup). Store token in project settings. GitHub App is v2 complexity.

5. **Mobile Layout Depth**
   - What we know: Need responsive layouts with bottom nav and simplified views
   - What's unclear: How simplified Kanban (list view) should look; which dashboard widgets to show on mobile
   - Recommendation: Kanban list view = sorted task list grouped by status (no columns). Mobile dashboard = subset of widgets (my tasks, unread, upcoming). No mobile drag-and-drop.

## Sources

### Primary (HIGH confidence)
- [LiveKit Egress API docs](https://docs.livekit.io/home/egress/api/) - Recording API, resource requirements
- [LiveKit Egress GitHub](https://github.com/livekit/egress) - SYS_ADMIN requirement, Docker setup
- [Fiber v3 Swagger contrib](https://docs.gofiber.io/contrib/next/swaggo/) - Official Fiber v3 Swagger integration
- [swaggo/swag](https://github.com/swaggo/swag) - Swagger doc generation from Go comments
- [i18next docs](https://www.i18next.com/) - i18n framework documentation
- [react-i18next docs](https://react.i18next.com) - React bindings, useTranslation hook
- [Tailwind CSS v4 dark mode](https://tailwindcss.com/docs/dark-mode) - CSS variable approach
- [vite-plugin-pwa docs](https://vite-pwa-org.netlify.app/) - PWA plugin configuration
- [GitHub webhook events](https://docs.github.com/en/webhooks/webhook-events-and-payloads) - Event types and payloads

### Secondary (MEDIUM confidence)
- [teambition/rrule-go](https://github.com/teambition/rrule-go) - RFC 5545 RRULE implementation for Go
- [otiai10/opengraph](https://github.com/otiai10/opengraph) - OG meta parser for Go
- [go-playground/webhooks v6](https://github.com/go-playground/webhooks) - Webhook parsing library
- [excelize](https://github.com/qax-os/excelize) - Excel read/write in Go
- [next-themes](https://github.com/pacocoursey/next-themes) - Theme switching for React
- [Tailwind v4 multi-theme guide](https://medium.com/render-beyond/build-a-flawless-multi-theme-ui-using-new-tailwind-css-v4-react-dca2b3c95510)

### Tertiary (LOW confidence)
- LiveKit Egress resource consumption numbers (2-6 CPU, 4GB RAM) - from community issues, varies by hardware
- JQL parser complexity estimates - based on Jira documentation, not validated implementation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries verified via npm/pkg.go.dev, versions confirmed
- Architecture: HIGH - follows established patterns from Phases 1-6; no new architectural concepts
- Pitfalls: HIGH - LiveKit Egress pitfall documented in initial research; GitHub webhook matching from PITFALLS.md
- i18n/Theme/PWA: HIGH - standard React patterns with well-documented libraries
- JQL parser: MEDIUM - custom implementation needed; scope unclear
- Backup/restore: MEDIUM - straightforward pg_dump/mongodump but restore safety is complex

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (stable domain, well-established patterns)
