# Roadmap: Hub

## Overview

Hub delivers a unified platform for development teams by building five integrated modules (auth/projects, task tracker, messenger, TMS, voice calls) plus cross-cutting features (notifications, search, dashboards) on a shared foundation. The roadmap follows dependency chains: foundation infrastructure first, then the four core modules in order of value and dependency (tracker, messenger, TMS, calls), then cross-cutting aggregation features, and finally polish and extended capabilities. Each phase delivers a complete, verifiable capability that builds on the previous phases.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation & Infrastructure** - Auth, projects, RBAC, real-time infra, storage, dev environment
- [ ] **Phase 2: Task Tracker** - Issue CRUD, Kanban boards, workflows, comments, activity log, time tracking
- [ ] **Phase 3: Messenger** - DMs, channels, threads, reactions, voice messages, presence, search
- [ ] **Phase 4: TMS & Allure** - Test cases, test runs, Allure ingestion, issue linking, reporting, import/export
- [ ] **Phase 5: Voice Calls** - Audio calls, screen sharing, call management, ephemeral chat
- [ ] **Phase 6: Notifications, Search & Dashboards** - In-app notifications, global search, personal/project/admin dashboards, audit
- [ ] **Phase 7: Polish & Extended Features** - Sprints, GitHub integration, recording, i18n, themes, PWA, API, backups

## Phase Details

### Phase 1: Foundation & Infrastructure
**Goal**: Users can authenticate, create projects with teams and roles, and the platform has working real-time infrastructure, file storage, and CI pipeline -- the substrate every module depends on
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, PROJ-01, PROJ-02, PROJ-03, PROJ-04, PROJ-06, PROJ-07, UI-01, UI-07, UI-08
**Success Criteria** (what must be TRUE):
  1. User can sign in via Magic Link email and via Telegram, land on a working dashboard shell, and sign out
  2. User can create a project and see auto-generated tracker board, TMS space, project chat, and channel stubs
  3. Project Owner can invite users by link, assign roles (Owner/Developer/Tester/Viewer), and role-restricted actions are enforced
  4. Files can be uploaded via drag-and-drop or paste and are stored in SeaweedFS with working previews
  5. WebSocket connection delivers real-time events from NATS to the browser (verified with a test event)
**Plans**: TBD

Plans:
- [ ] 01-01: Dev environment, project structure, CI pipeline, Docker Compose with all services
- [ ] 01-02: Database schemas (Ent + MongoDB), NATS setup, Redis, SeaweedFS integration
- [ ] 01-03: Auth module (Magic Link + Telegram login, sessions, RBAC middleware)
- [ ] 01-04: Projects module (CRUD, membership, roles, teams, archive/trash)
- [ ] 01-05: Frontend shell (React app, routing, sidebar, top bar, error handling, WebSocket gateway)

### Phase 2: Task Tracker
**Goal**: Users can manage project work through a full-featured issue tracker with Kanban boards, custom workflows, rich comments, and time tracking
**Depends on**: Phase 1
**Requirements**: TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-06, TASK-07, TASK-08, TASK-09, TASK-10, TASK-11, KANB-01, KANB-04, KANB-05, UI-09
**Success Criteria** (what must be TRUE):
  1. User can create issues of all 9 types with rich text descriptions, file attachments, and project-prefixed IDs (HUB-123)
  2. User can drag issues on a Kanban board between columns that map to workflow statuses, with real-time updates visible to other users
  3. Project Owner can customize workflow by adding/removing statuses and defining allowed transitions
  4. User can create parent-child hierarchies and cross-issue links (blocks, relates to, duplicates, etc.)
  5. User can comment with @mentions and reactions, view full change history, and track time with start/stop timer
**Plans**: TBD

Plans:
- [ ] 02-01: Issue CRUD backend (Ent schemas, 9 types, custom fields, hierarchy, linking, numbering)
- [ ] 02-02: Workflow engine (custom statuses, transitions, validation)
- [ ] 02-03: Kanban board frontend (drag-and-drop, multiple boards, filters, real-time sync)
- [ ] 02-04: Comments, activity log, time tracking, labels, attachments

### Phase 3: Messenger
**Goal**: Users can communicate in real time through DMs, channels, and threads with rich text, reactions, voice messages, and presence indicators
**Depends on**: Phase 1, Phase 2 (for cross-module entity references in chat)
**Requirements**: MSG-01, MSG-02, MSG-03, MSG-04, MSG-05, MSG-06, MSG-07, MSG-08, MSG-09, MSG-10, MSG-12, MSG-13
**Success Criteria** (what must be TRUE):
  1. User can send and receive messages in real time via DMs, group chats, public channels, and private channels
  2. User can reply in threads, react with emoji, format text (bold/italic/code/lists), and record voice messages
  3. Project channels are auto-created when a project is created; read-only channels work for bot posts
  4. User can edit/delete own messages, pin messages, bookmark messages, and search across all conversations
  5. Online/offline/away status is visible in real time and unread counts update correctly across channels
**Plans**: TBD

Plans:
- [ ] 03-01: Chat backend (MongoDB schemas, channel/DM CRUD, message delivery via NATS, read receipts)
- [ ] 03-02: Channel management (public/private, auto-channels, read-only, membership, permissions)
- [ ] 03-03: Chat frontend (message list with virtualization, rich text input, threads, reactions, voice messages)
- [ ] 03-04: Presence system, unread counts, message search, pins, bookmarks, forwarding

### Phase 4: TMS & Allure Integration
**Goal**: Users can manage test cases, execute manual and automated test runs, ingest Allure CI results, and see bidirectional links between tests and issues
**Depends on**: Phase 1, Phase 2 (for issue linking)
**Requirements**: TMS-01, TMS-02, TMS-03, TMS-04, TMS-05, TMS-06, TMS-08, TMS-09, TMS-10, ALLU-01, ALLU-02, ALLU-03, IMEX-02
**Success Criteria** (what must be TRUE):
  1. User can create test cases in classic (steps + expected) and BDD (Given/When/Then) formats with folder hierarchy and versioning
  2. User can create and execute manual test runs with pass/fail/blocked/skipped per step, attach screenshots and comments per result
  3. Allure results JSON from CI are ingested via API, matched to test cases by ID or name, and appear in run history
  4. Test cases link bidirectionally to issues: user can create a bug from a failed test, and coverage is tracked
  5. Reports show pass rates, trends, flaky test detection, and run comparisons at Allure-quality level
**Plans**: TBD

Plans:
- [ ] 04-01: Test case CRUD backend (classic + BDD, folders, tags, components, versioning, Ent schemas)
- [ ] 04-02: Test runs (manual execution, evidence, statuses, assignment, re-open)
- [ ] 04-03: Allure integration (parser for allure-results JSON, API endpoint, matching, run history, Telegram notifications)
- [ ] 04-04: Cross-module linking (test-to-issue bidirectional, coverage tracking, GitHub test linking), reporting, import/export

### Phase 5: Voice Calls
**Goal**: Users can join audio calls with screen sharing for standups, demos, and ad-hoc discussions
**Depends on**: Phase 1 (auth, projects), Phase 3 (call events posted to messenger channels)
**Requirements**: CALL-01, CALL-03, CALL-06
**Success Criteria** (what must be TRUE):
  1. User can start or join an audio call with up to 50 participants via a link or from within a channel
  2. User can mute/unmute, raise hand, and share their screen or a specific window (multiple simultaneous shares supported)
  3. Ephemeral call chat is available during the call and disappears after the call ends
  4. Call start/end events appear in the associated messenger channel
**Plans**: TBD

Plans:
- [ ] 05-01: LiveKit integration (server setup, room management, token generation, Caddy config)
- [ ] 05-02: Call UI (participant list, mute/unmute, raise hand, screen share picker)
- [ ] 05-03: Call lifecycle (start/join/leave/end events, NATS integration, ephemeral chat, messenger notifications)

### Phase 6: Notifications, Search & Dashboards
**Goal**: Users have a unified notification center, can find anything instantly via global search, and see project health at a glance on dashboards
**Depends on**: Phase 1-5 (aggregates data from all modules)
**Requirements**: NOTF-01, NOTF-02, NOTF-03, NOTF-04, SRCH-01, SRCH-02, DASH-01, DASH-02, DASH-03, AUDT-01, AUDT-02, AUDT-03, AUDT-04, UI-05
**Success Criteria** (what must be TRUE):
  1. User receives in-app notifications (bell icon) for assignments, @mentions, status changes, and other configured triggers across all modules
  2. User can configure notification preferences by type, channel, and project; quiet hours suppress non-urgent notifications
  3. Ctrl+K opens global search that finds issues, test cases, users, projects, channels, and files across the entire platform
  4. Personal dashboard shows user's tasks, unreads, calls, runs, and deadlines with customizable drag-and-drop widgets
  5. Project dashboard shows task statuses, burndown, velocity, test trends, and team activity; Admin dashboard shows system-wide metrics
**Plans**: TBD

Plans:
- [ ] 06-01: Notification system (NATS consumers, triggers from all modules, in-app delivery, Telegram bot, preferences, quiet hours)
- [ ] 06-02: Global search (PostgreSQL full-text search with tsvector/GIN, cross-module indexing, Ctrl+K UI), saved/shared filters
- [ ] 06-03: Dashboards (personal, project, admin) with Tremor widgets, drag-and-drop layout, templates
- [ ] 06-04: Audit log, resource limits, trash/restore system, user deletion with GDPR compliance

### Phase 7: Polish & Extended Features
**Goal**: The platform gains sprint management, GitHub integration, call recording, internationalization, themes, and operational hardening that deepen each module beyond table stakes
**Depends on**: Phase 1-6
**Requirements**: PROJ-05, KANB-02, KANB-03, SCRM-01, SCRM-02, SCRM-03, SCRM-04, SCRM-05, MSG-11, TMS-07, CALL-02, CALL-04, CALL-05, GHUB-01, GHUB-02, GHUB-03, GHUB-04, GHUB-05, SRCH-03, API-01, API-02, API-03, API-04, IMEX-01, IMEX-03, BKUP-01, BKUP-02, BKUP-03, UI-02, UI-03, UI-04, UI-06
**Success Criteria** (what must be TRUE):
  1. Scrum teams can create sprints, drag tasks from backlog, run Planning Poker, and view burndown/velocity charts
  2. GitHub PRs/commits auto-link to issues by ID mention, PR status appears on issue cards, and merging a PR can auto-transition issue status
  3. Users can record calls (audio + screen share), recordings are saved to the project, and a link is posted to the channel
  4. The app works in Russian and English with theme switching (dark/light/system + accent color) and basic PWA install capability
  5. REST API is documented via Swagger, API keys can be generated, webhooks fire on events, and automated backups run weekly
**Plans**: TBD

Plans:
- [ ] 07-01: Sprint management (create/start/complete sprints, backlog, auto-carry, burndown, velocity, retrospective, Planning Poker)
- [ ] 07-02: GitHub integration (webhook receiver, PR/commit linking, auto-status, branch creation from issue)
- [ ] 07-03: Call recording (LiveKit Egress, storage, project binding, channel link), scheduled/recurring calls
- [ ] 07-04: i18n (RU/EN), theming (dark/light/system/accent), mobile layout, PWA, link previews in messenger
- [ ] 07-05: REST API (Swagger), API keys, webhooks, hub-cli, import/export, project templates, backups, WIP limits, swimlanes, test environments, JQL

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Infrastructure | 0/5 | Not started | - |
| 2. Task Tracker | 0/4 | Not started | - |
| 3. Messenger | 0/4 | Not started | - |
| 4. TMS & Allure | 0/4 | Not started | - |
| 5. Voice Calls | 0/3 | Not started | - |
| 6. Notifications, Search & Dashboards | 0/4 | Not started | - |
| 7. Polish & Extended Features | 0/5 | Not started | - |
