# Milestone 2: Hub v2.0

**Defined:** 2026-03-22
**Depends on:** Milestone 1 (v1.0) -- all 95 requirements complete across 7 phases

## Vision

Hub v1.0 delivered a working unified platform. Milestone 2 transforms it from a functional prototype into a production-grade, AI-enhanced development platform with deep GitHub integration. Three pillars drive this milestone: (1) technical debt payoff and UI polish that make the existing features actually pleasant to use, (2) deep GitHub integration that makes Hub the single pane of glass for the entire development workflow, and (3) AI capabilities that multiply team productivity. The result is a platform that teams cannot live without, deployed on production infrastructure with CI/CD, monitoring, and zero-downtime deployments.

## Phase Overview

| Phase | Name | Complexity | Priority |
|-------|------|------------|----------|
| 8 | UI Redesign & Performance | L | P0 -- Must do first |
| 9 | Production Pipeline & DevOps | XL | P0 -- Foundation for all deploys |
| 10 | Deep GitHub Integration | XL | P1 -- Core value driver |
| 11 | AI Assistant Integration | L | P1 -- Differentiator |
| 12 | Team Productivity Features | M | P2 -- Enhancement layer |

**Phase ordering rationale:**
- Phase 8 first: fixes technical debt from v1 that makes development harder. Splitting large components and code splitting reduce build times and improve DX for all subsequent phases.
- Phase 9 second: CI/CD and Docker optimization are needed before GitHub integration (Phase 10 depends on running pipelines). Also blocks production deployment.
- Phase 10 third: GitHub integration is the highest-value feature. Requires Phase 9 (CI/CD pipelines must exist to show status). OAuth infrastructure also enables AI features.
- Phase 11 fourth: AI features build on top of existing task/TMS data and GitHub integration. Claude API calls need production monitoring (Phase 9).
- Phase 12 last: Productivity features are enhancements on top of a polished, production-deployed, AI-enhanced platform.

---

## Phase 8: UI Redesign & Performance

**Goal:** Fix all partially-implemented v1 features, split oversized components, implement code splitting, and optimize rendering performance so the app feels fast and maintainable.

**Depends on:** Milestone 1 complete
**Complexity:** L (Large)
**Priority:** P0

### Requirements

#### Fix Partially-Implemented Features

- [ ] **M2-001**: Quiet hours -- full implementation with timezone-aware scheduling, midnight crossover handling, urgent notification bypass with visual indicator
- [ ] **M2-002**: WIP limits UI -- visual warning (yellow) when column approaches limit, hard block (red) when at limit, counter badge on column header
- [ ] **M2-003**: Swimlanes UI -- collapsible swimlane rows on Kanban board, swimlane by assignee/priority/type with drag-and-drop between lanes
- [ ] **M2-004**: Accent color picker -- full HSL/oklch color picker in settings, live preview, persist to user preferences, apply via CSS custom properties
- [ ] **M2-005**: Ctrl+N shortcut -- global shortcut opens new task creation modal from any page, with project context auto-filled
- [ ] **M2-006**: PWA service worker -- proper Workbox caching strategy (stale-while-revalidate for API, cache-first for static), background sync for offline mutations, push notification subscription
- [ ] **M2-007**: Fix Ctrl+K search result ranking -- weight recent items higher, show keyboard navigation hints, fix empty state
- [ ] **M2-008**: Fix link preview reliability in messenger -- handle timeout gracefully, cache OG metadata, show loading skeleton
- [ ] **M2-009**: Fix mobile bottom navigation z-index and gesture conflicts with sidebar
- [ ] **M2-010**: Fix i18n missing translation keys -- audit all components, add fallback handling, fix pluralization rules
- [ ] **M2-011**: Fix dark theme contrast issues in TMS tables and Kanban cards
- [ ] **M2-012**: Fix Planning Poker real-time sync edge cases (late joiners, reconnection)

#### Component Splitting

- [ ] **M2-013**: Split `message-item.tsx` (677 lines) into MessageBubble, MessageActions, MessageReactions, MessageThread, VoiceMessagePlayer
- [ ] **M2-014**: Split `channel-list.tsx` (661 lines) into ChannelGroup, ChannelItem, ChannelSearch, UnreadBadge, ChannelContextMenu
- [ ] **M2-015**: Split `cases_.$caseKey.tsx` (876 lines) into TestCaseHeader, StepEditor, TestCaseHistory, TestCaseSidebar, LinkedIssues
- [ ] **M2-016**: Split `cases.tsx` (698 lines) into CaseTree, CaseFilters, CaseToolbar, CaseBulkActions
- [ ] **M2-017**: Split `runs_.$runId.tsx` (691 lines) into RunHeader, RunResults, RunExecution, RunStatistics, RunEvidence
- [ ] **M2-018**: Split `sidebar.tsx` (479 lines) into SidebarNav, SidebarProjects, SidebarUser, SidebarCollapse
- [ ] **M2-019**: Split `$issueKey.tsx` (498 lines) into IssueHeader, IssueDescription, IssueActivity, IssueSidebar, IssueRelations

#### Code Splitting & Bundle Optimization

- [ ] **M2-020**: Implement TanStack Router lazy routes (`.lazy.tsx`) for all major route files -- automatic code splitting for route components
- [ ] **M2-021**: Configure Rollup `manualChunks` in Vite config -- separate vendor chunks (react, tanstack, plate, livekit, recharts, zustand)
- [ ] **M2-022**: Lazy-load Plate editor (rich text) -- only load when user focuses editor, not on page load
- [ ] **M2-023**: Lazy-load LiveKit SDK -- only load when user joins a call
- [ ] **M2-024**: Lazy-load Recharts/Tremor -- only load on dashboard routes
- [ ] **M2-025**: Analyze bundle with `rollup-plugin-visualizer`, target main chunk under 200KB gzipped (from ~2.6MB uncompressed)
- [ ] **M2-026**: Implement `<link rel="modulepreload">` hints for predictable navigation (e.g., preload board when hovering project)

#### Performance Optimization

- [ ] **M2-027**: Add `React.memo` to all list item components (MessageItem, IssueCard, TestCaseRow, ChannelItem, NotificationItem)
- [ ] **M2-028**: Implement `@tanstack/react-virtual` for all long lists -- message list (already virtual), issue list, test case tree, activity log, notification list
- [ ] **M2-029**: Debounce search inputs (300ms), filter changes (150ms), resize observers
- [ ] **M2-030**: Implement `useDeferredValue` for heavy filter computations on Kanban board
- [ ] **M2-031**: Add `loading="lazy"` to all images, implement intersection observer for file previews
- [ ] **M2-032**: Profile and fix unnecessary re-renders in Kanban drag-and-drop (isolate drag state from card content)

### Success Criteria

1. All 12 partially-implemented features are fully functional with no visual or behavioral gaps
2. No frontend component file exceeds 300 lines; all split components have clear single responsibilities
3. Main JS bundle chunk is under 200KB gzipped; route-based chunks load on navigation
4. Lighthouse Performance score >= 80 on project board page with 100 tasks
5. Time-to-interactive under 3 seconds on 4G throttled connection

### Plans: 4 plans

- 08-01: Fix 12 partially-implemented features (quiet hours, WIP, swimlanes, accent colors, Ctrl+N, PWA, search, previews, mobile, i18n, dark theme, poker)
- 08-02: Split 7 oversized components into focused sub-components with barrel exports
- 08-03: Code splitting -- TanStack Router lazy routes, manualChunks, lazy-load heavy libs (Plate, LiveKit, Recharts)
- 08-04: Performance optimization -- React.memo, virtual lists, debounce, profiling, image lazy loading

---

## Phase 9: Production Pipeline & DevOps

**Goal:** Build a complete CI/CD pipeline, optimize Docker builds, create K3s deployment manifests, and add monitoring/observability so Hub can be deployed to production with zero-downtime updates.

**Depends on:** Phase 8 (clean codebase to build/deploy)
**Complexity:** XL (Extra Large)
**Priority:** P0

### Requirements

#### CI/CD Pipeline

- [ ] **M2-033**: GitHub Actions workflow -- monorepo path filtering (backend/**, frontend/**, cli/**, deployments/**) to run only affected jobs
- [ ] **M2-034**: Backend CI job -- `go vet`, `go test ./...`, `golangci-lint run`, build binary, cache Go modules
- [ ] **M2-035**: Frontend CI job -- `tsc --noEmit`, `vitest run`, `vite build`, cache node_modules
- [ ] **M2-036**: CLI CI job -- `go test`, build cross-platform binaries (linux/amd64, darwin/arm64, windows/amd64)
- [ ] **M2-037**: Docker build job -- build and push images to GitHub Container Registry (ghcr.io) on main branch merge
- [ ] **M2-038**: Deploy job -- SSH/kubectl apply to K3s cluster on tag push, with rollback on health check failure
- [ ] **M2-039**: PR checks -- required status checks (lint, test, build) before merge, with Codecov coverage reporting

#### Docker Optimization

- [ ] **M2-040**: Backend Dockerfile multi-stage build -- builder stage (Go compile) + runtime stage (distroless/static), target image < 30MB
- [ ] **M2-041**: Frontend Dockerfile multi-stage build -- node builder (vite build) + Caddy/nginx static server, target image < 25MB
- [ ] **M2-042**: Docker layer caching optimization -- `go.sum` / `package-lock.json` copied before source for dependency caching
- [ ] **M2-043**: Docker Compose production profile -- separate from dev, with resource limits, health checks, restart policies
- [ ] **M2-044**: `.dockerignore` cleanup -- exclude tests, docs, .planning, node_modules, .git

#### K3s Deployment

- [ ] **M2-045**: Kubernetes manifests (Kustomize) for all services: backend, frontend, PostgreSQL, MongoDB, Redis, NATS, SeaweedFS, LiveKit, Caddy
- [ ] **M2-046**: ConfigMaps for non-sensitive config, Secrets for credentials (DB passwords, JWT secret, API keys)
- [ ] **M2-047**: Horizontal Pod Autoscaler for backend (CPU 70% threshold, min 2, max 5 replicas)
- [ ] **M2-048**: PersistentVolumeClaims for PostgreSQL, MongoDB, SeaweedFS data directories
- [ ] **M2-049**: Ingress configuration via Traefik (K3s default) with TLS termination
- [ ] **M2-050**: Resource requests and limits for all pods (backend: 256Mi/512Mi RAM, 250m/500m CPU)
- [ ] **M2-051**: Liveness and readiness probes on all services

#### Monitoring & Observability

- [ ] **M2-052**: `/health` endpoint -- readiness (DB, Redis, NATS connected) and liveness (process alive) checks
- [ ] **M2-053**: `/metrics` endpoint -- Prometheus format via `fiberprometheus/v3`: request count, latency histogram, error rate, in-progress requests
- [ ] **M2-054**: Custom business metrics -- active WebSocket connections, NATS message throughput, active calls, online users
- [ ] **M2-055**: Structured logging with `slog` -- JSON format in production, correlation ID per request (X-Request-ID), log level via env var
- [ ] **M2-056**: Grafana dashboards (JSON provisioning) -- API latency, error rate, WebSocket connections, DB query time, NATS lag
- [ ] **M2-057**: Loki log aggregation -- collect from all pods, label by service, query via Grafana
- [ ] **M2-058**: Alert rules -- 5xx rate > 1%, p99 latency > 2s, DB connection pool exhausted, NATS consumer lag > 1000

#### Database Migration Strategy

- [ ] **M2-059**: Atlas versioned migrations -- generate migration files from Ent schema diff, apply via `atlas migrate apply`
- [ ] **M2-060**: Migration CI check -- `atlas migrate lint` in PR pipeline to catch destructive changes
- [ ] **M2-061**: Migration rollback strategy -- down migrations for each up migration, tested in CI
- [ ] **M2-062**: MongoDB migration scripts -- versioned JavaScript migration files for schema changes (indexes, collection renames)

#### Zero-Downtime Deployment

- [ ] **M2-063**: Rolling update strategy in K3s -- `maxSurge: 1, maxUnavailable: 0` for zero-downtime
- [ ] **M2-064**: Graceful shutdown in Go backend -- drain WebSocket connections, finish in-flight requests (30s timeout), deregister from NATS
- [ ] **M2-065**: Database migration runs as Kubernetes Job before deployment, not during pod startup
- [ ] **M2-066**: Health check gate -- new pods must pass readiness probe before old pods terminate

### Success Criteria

1. `git push` to main triggers full CI pipeline (lint, test, build, Docker push) completing in under 10 minutes
2. Tag push (`v2.x.x`) triggers automated deployment to K3s with zero user-visible downtime
3. `/health` and `/metrics` endpoints are accessible; Grafana dashboard shows real-time API metrics
4. Docker images are under 30MB (backend) and 25MB (frontend)
5. Atlas migrations are lint-checked in CI; destructive changes are flagged and require manual approval

### Plans: 5 plans

- 09-01: GitHub Actions CI workflows -- path-filtered jobs for backend, frontend, CLI with caching
- 09-02: Docker multi-stage builds -- optimized Dockerfiles, .dockerignore, production compose profile
- 09-03: K3s Kustomize manifests -- deployments, services, ingress, PVCs, HPA, probes
- 09-04: Monitoring stack -- health/metrics endpoints, fiberprometheus, slog correlation IDs, Grafana dashboards, Loki, alerts
- 09-05: Atlas versioned migrations + zero-downtime deployment strategy (rolling updates, graceful shutdown, migration jobs)

---

## Phase 10: Deep GitHub Integration

**Goal:** Transform Hub's GitHub integration from webhook-only to full bidirectional OAuth-based integration with CI/CD visibility, code review status, and issue sync.

**Depends on:** Phase 9 (CI/CD pipeline must exist for pipeline status display)
**Complexity:** XL (Extra Large)
**Priority:** P1

### Requirements

#### GitHub App & OAuth

- [ ] **M2-067**: Register Hub as a GitHub App (not OAuth App) -- fine-grained permissions, installation-level access, not tied to user account
- [ ] **M2-068**: GitHub App installation flow -- org owner installs app, selects repositories, Hub stores installation ID and access tokens
- [ ] **M2-069**: User-level GitHub OAuth for personal actions (commenting on PRs, creating branches) -- "Connect GitHub Account" in user settings
- [ ] **M2-070**: Token management -- short-lived installation tokens (1h), refresh automatically, encrypted storage in PostgreSQL
- [ ] **M2-071**: Repository selection UI -- after GitHub App install, show list of accessible repos, allow project-to-repo mapping

#### CI/CD Pipeline Visibility

- [ ] **M2-072**: Fetch GitHub Actions workflow runs for mapped repositories -- show status (queued/in_progress/completed) on project dashboard
- [ ] **M2-073**: Pipeline status widget on project dashboard -- last 10 runs with commit message, branch, status, duration, link to GitHub
- [ ] **M2-074**: Pipeline status badge on task cards -- if task's branch has a running/failed/passed pipeline, show inline indicator
- [ ] **M2-075**: Real-time pipeline updates via GitHub App webhooks (`workflow_run`, `check_run` events)
- [ ] **M2-076**: Pipeline failure notification -- when CI fails on a branch linked to a task, notify task assignee

#### Test Report Integration

- [ ] **M2-077**: Pull test reports from GitHub Actions artifacts -- download artifact ZIP, parse JUnit XML / Allure results JSON
- [ ] **M2-078**: Auto-import test results into TMS after successful pipeline run -- match by `@TestCase` annotation or test name
- [ ] **M2-079**: Test result trend from CI -- show pass rate trend per branch, detect regressions across runs
- [ ] **M2-080**: Link test failures to tasks -- when a test linked to a task fails in CI, add comment to task with failure details

#### Code & Diff Integration

- [ ] **M2-081**: Show files changed in PR directly in Hub task view -- file list with additions/deletions count
- [ ] **M2-082**: Inline diff viewer for PR files linked to a task -- syntax-highlighted, collapsible file diffs
- [ ] **M2-083**: Commit history timeline on task -- show all commits mentioning task ID with message, author, date

#### GitHub Issues Sync

- [ ] **M2-084**: Bidirectional issue sync -- Hub task creates GitHub Issue, GitHub Issue creates Hub task (configurable per project)
- [ ] **M2-085**: Status mapping configuration -- map Hub workflow statuses to GitHub Issue labels (e.g., "In Progress" -> label:in-progress)
- [ ] **M2-086**: Comment sync -- comments on Hub task appear on GitHub Issue and vice versa (with attribution)
- [ ] **M2-087**: Conflict resolution -- last-write-wins with sync log showing all changes and their source

#### Code Review Integration

- [ ] **M2-088**: Show PR review status on task card -- approved / changes_requested / pending, with reviewer names
- [ ] **M2-089**: Review request notification -- when PR linked to task gets review request, notify task watcher in Hub
- [ ] **M2-090**: Branch protection status -- show if branch is protected, required checks passing, merge-ready status on task

### Success Criteria

1. Org owner can install Hub GitHub App, select repositories, and see pipeline status on project dashboard within 5 minutes
2. PR review status (approved/changes_requested/pending) appears on linked task cards in real-time
3. Test results from GitHub Actions are automatically imported into TMS after pipeline completion
4. Bidirectional issue sync works: creating a task in Hub creates a GitHub Issue, and changes sync both ways
5. Code changes (files changed, diff) for a PR are viewable directly in the Hub task detail page

### Plans: 5 plans

- 10-01: GitHub App registration, installation flow, OAuth user connection, token management, repository mapping UI
- 10-02: CI/CD pipeline visibility -- fetch workflow runs, dashboard widget, task badge, real-time webhook updates, failure notifications
- 10-03: Test report integration -- artifact download, JUnit/Allure parsing, auto-import to TMS, regression detection
- 10-04: Code integration -- PR diff viewer, commit timeline, file changes on task view
- 10-05: GitHub Issues bidirectional sync, code review status, branch protection visibility

---

## Phase 11: AI Assistant Integration

**Goal:** Add AI-powered capabilities using Claude API for test generation, code review suggestions, smart search, task estimation, sprint summaries, and bug enrichment.

**Depends on:** Phase 10 (GitHub integration needed for code review AI), Phase 9 (monitoring needed for API cost tracking)
**Complexity:** L (Large)
**Priority:** P1

### Requirements

#### AI Infrastructure

- [ ] **M2-091**: Anthropic Claude API integration via official Go SDK (`anthropic-sdk-go`) -- API key management in admin settings, usage tracking
- [ ] **M2-092**: AI request queue via NATS JetStream -- rate limiting (60 req/min), retry with backoff, cost tracking per request
- [ ] **M2-093**: AI response caching in Redis -- cache identical prompts for 1 hour, invalidate on context change
- [ ] **M2-094**: AI settings per project -- enable/disable AI features, model selection (Haiku for fast/cheap, Sonnet for quality), usage budget

#### Test Case Generation

- [ ] **M2-095**: Generate test cases from task description -- user clicks "Generate Tests" on a task, AI creates test cases (classic steps + BDD) based on task title, description, and acceptance criteria
- [ ] **M2-096**: Generate test cases from code changes -- when PR is linked to task, AI analyzes diff and suggests test cases for changed functionality
- [ ] **M2-097**: AI test review -- analyze existing test cases for coverage gaps, suggest missing edge cases, flag redundant tests
- [ ] **M2-098**: Generated tests go to "AI Suggested" folder in TMS, require human review before moving to active suite

#### AI Code Review

- [ ] **M2-099**: AI review suggestions on PR diffs -- when user views PR in Hub, option to "Ask AI for Review" analyzes diff and provides inline comments
- [ ] **M2-100**: Bug risk analysis -- AI flags code patterns likely to introduce bugs (error handling gaps, race conditions, missing null checks)
- [ ] **M2-101**: AI review comments posted to task activity feed (not directly to GitHub PR -- user decides what to forward)

#### AI-Powered Search

- [ ] **M2-102**: Natural language search -- user types "tasks assigned to me that are overdue" or "test cases for login feature" in Ctrl+K, AI translates to JQL/filters
- [ ] **M2-103**: Semantic search across tasks, test cases, messages -- vector embeddings stored in PostgreSQL pgvector extension
- [ ] **M2-104**: "Ask about project" -- conversational AI that answers questions like "what is the test coverage for authentication module?" by querying Hub data

#### Smart Estimation

- [ ] **M2-105**: AI task estimation -- based on historical time tracking data (similar tasks by type, labels, description length), suggest story points and time estimate
- [ ] **M2-106**: Estimation confidence indicator -- HIGH/MEDIUM/LOW based on amount of historical data available for similar tasks
- [ ] **M2-107**: Sprint planning assistant -- given backlog items and team velocity, suggest which tasks fit in next sprint

#### Sprint & Retrospective AI

- [ ] **M2-108**: AI sprint retrospective summary -- analyze completed tasks, time tracking, bug rate, test results, and generate structured summary (what went well, what didn't, action items)
- [ ] **M2-109**: AI standup digest -- daily summary of what each team member did yesterday and what's planned today (from task updates)

#### Bug Report Enhancement

- [ ] **M2-110**: Similar issue detection -- when creating a bug, AI searches existing tasks and shows potential duplicates with similarity score
- [ ] **M2-111**: Suggested assignee -- based on code ownership (GitHub blame data), recent task history, and current workload
- [ ] **M2-112**: Auto-priority suggestion -- based on affected component, severity keywords, and historical priority distribution
- [ ] **M2-113**: Bug enrichment -- AI adds structured fields (steps to reproduce, expected/actual behavior) from unstructured bug description

### Success Criteria

1. User can click "Generate Tests" on any task and receive relevant test cases within 10 seconds (using Claude Haiku)
2. Natural language search resolves "my overdue bugs" to correct JQL filter and returns accurate results
3. AI estimation provides story point suggestions with confidence levels for tasks with sufficient historical data
4. Sprint retrospective AI summary covers completed work, blockers, test quality, and suggests action items
5. AI usage is tracked per project with budget controls; admin can see total API costs

### Plans: 5 plans

- 11-01: AI infrastructure -- Anthropic SDK integration, NATS request queue, Redis caching, settings UI, cost tracking
- 11-02: Test case generation -- from task descriptions, from PR diffs, coverage gap analysis, AI suggested folder
- 11-03: AI code review + AI-powered search -- PR diff analysis, natural language JQL, semantic search (pgvector)
- 11-04: Smart estimation + sprint AI -- historical analysis, estimation confidence, sprint planning assistant, retro summary, standup digest
- 11-05: Bug enhancement -- duplicate detection, suggested assignee, auto-priority, description enrichment

---

## Phase 12: Team Productivity Features

**Goal:** Add analytics, resource management, cross-project views, bulk operations, and notification digests that make Hub indispensable for team leads and managers.

**Depends on:** Phase 10 (GitHub data for analytics), Phase 11 (AI for velocity predictions)
**Complexity:** M (Medium)
**Priority:** P2

### Requirements

#### Time Reports & Analytics

- [ ] **M2-114**: Time report per team member -- daily/weekly/monthly breakdown of logged time by project, task type, and label
- [ ] **M2-115**: Time report per project -- total effort, estimated vs actual comparison, time per workflow status
- [ ] **M2-116**: Team time report -- aggregated view with ability to drill down to individual members
- [ ] **M2-117**: Export time reports to CSV/PDF with customizable date range and filters

#### Resource Allocation

- [ ] **M2-118**: Resource allocation heatmap -- show team members on Y axis, days on X axis, color-coded by workload (green/yellow/red based on assigned task points vs capacity)
- [ ] **M2-119**: Capacity planning -- each member sets weekly capacity (hours or story points), system shows available vs assigned
- [ ] **M2-120**: Overload alerts -- notification when a team member is assigned more than 120% of their capacity

#### Sprint Predictions

- [ ] **M2-121**: Sprint velocity prediction -- based on last 5 sprints, predict likely completion range (optimistic/pessimistic/expected) for current sprint
- [ ] **M2-122**: Burndown prediction line -- extend burndown chart with projected completion based on current velocity
- [ ] **M2-123**: Scope creep indicator -- track tasks added mid-sprint, show impact on burndown

#### Custom Dashboard Widgets

- [ ] **M2-124**: Widget API -- define custom widgets with data source (JQL query), visualization type (number, chart, table, list), and refresh interval
- [ ] **M2-125**: Widget gallery -- browse and add widgets from a catalog, including community-shared widget configs
- [ ] **M2-126**: Widget sharing -- export/import widget configurations as JSON between projects

#### Notification Digest

- [ ] **M2-127**: Daily digest email via Resend -- summary of yesterday's activity: tasks assigned/completed, messages, test results, upcoming deadlines
- [ ] **M2-128**: Daily digest Telegram message -- same content as email, formatted for Telegram with inline buttons to open Hub
- [ ] **M2-129**: Digest preferences -- choose delivery time (default 9:00 AM user timezone), channels (email/Telegram/both/none), content sections to include

#### Cross-Project Views

- [ ] **M2-130**: "My Work" global page -- all tasks assigned to current user across all projects, sortable by due date, priority, project
- [ ] **M2-131**: Cross-project search -- search tasks and test cases across all accessible projects in one query
- [ ] **M2-132**: Global activity feed -- combined activity from all user's projects, filterable by type

#### Bulk Operations

- [ ] **M2-133**: Bulk status change -- select multiple tasks on board/list view, change status in one action
- [ ] **M2-134**: Bulk assign -- select multiple tasks, assign to team member(s)
- [ ] **M2-135**: Bulk label -- add/remove labels from multiple tasks at once
- [ ] **M2-136**: Bulk move -- move selected tasks to different sprint/project
- [ ] **M2-137**: Bulk selection UI -- checkbox on each task card/row, "Select All" on page, floating action bar showing count and available actions

### Success Criteria

1. Team lead can view time reports per person and project with drill-down, exported as PDF
2. Resource allocation heatmap accurately shows team workload based on assigned story points vs capacity
3. Sprint velocity prediction displays optimistic/pessimistic/expected completion lines on burndown chart
4. User can view all their tasks across all projects on a single "My Work" page
5. Bulk operations (status, assign, label) work on 50+ selected tasks without performance degradation

### Plans: 5 plans

- 12-01: Time reports and analytics -- per person, per project, per team, export to CSV/PDF
- 12-02: Resource allocation heatmap, capacity planning, overload alerts
- 12-03: Sprint predictions (velocity, burndown projection, scope creep) + custom dashboard widgets API
- 12-04: Notification digest (email + Telegram) + cross-project views (My Work, global search, activity feed)
- 12-05: Bulk operations -- multi-select UI, bulk status/assign/label/move with floating action bar

---

## Dependencies Graph

```
Phase 8 (UI Polish)
    |
    v
Phase 9 (DevOps/CI/CD)
    |
    +---> Phase 10 (GitHub Integration)
    |         |
    |         v
    |     Phase 11 (AI) <--- also depends on Phase 9
    |         |
    v         v
Phase 12 (Team Productivity) <--- depends on Phase 10 + Phase 11
```

**Critical path:** Phase 8 -> Phase 9 -> Phase 10 -> Phase 11 -> Phase 12

**Parallelization opportunity:** Phase 10 and Phase 11 infrastructure (M2-091 through M2-094) can begin in parallel once Phase 9 is done, but Phase 11's AI code review (M2-099) depends on Phase 10's diff viewer.

---

## Requirements Summary

| ID Range | Phase | Count |
|----------|-------|-------|
| M2-001 to M2-032 | Phase 8: UI Redesign & Performance | 32 |
| M2-033 to M2-066 | Phase 9: Production Pipeline & DevOps | 34 |
| M2-067 to M2-090 | Phase 10: Deep GitHub Integration | 24 |
| M2-091 to M2-113 | Phase 11: AI Assistant Integration | 23 |
| M2-114 to M2-137 | Phase 12: Team Productivity Features | 24 |
| **Total** | **Milestone 2** | **137** |

---

## Risk Register

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| GitHub App rate limits hit during heavy sync | High | Medium | Cache aggressively, use conditional requests (ETags), batch API calls |
| Claude API costs exceed budget | Medium | Medium | Haiku for fast tasks, budget caps per project, cache identical prompts |
| K3s cluster instability on CloudForBox | High | Low | Fallback to Docker Compose deploy, K3s auto-healing, regular backups |
| pgvector extension not available on managed PG | Medium | Low | Use cosine similarity on JSONB arrays as fallback, pgvector is widely supported |
| Large component refactor breaks existing features | High | Medium | Phase 8 runs first, comprehensive Playwright E2E tests before and after split |
| Atlas migration conflicts in parallel development | Medium | Medium | CI lint check, single migration branch, review before merge |

---

## Estimated Effort

Based on Milestone 1 velocity (40 plans, ~8min avg per plan):

| Phase | Plans | Estimated Duration |
|-------|-------|-------------------|
| Phase 8 | 4 plans | ~45 min |
| Phase 9 | 5 plans | ~60 min |
| Phase 10 | 5 plans | ~55 min |
| Phase 11 | 5 plans | ~55 min |
| Phase 12 | 5 plans | ~45 min |
| **Total** | **24 plans** | **~4.5 hours** |

---

*Defined: 2026-03-22*
*Based on: Milestone 1 complete (95 requirements, 7 phases, 40 plans)*
