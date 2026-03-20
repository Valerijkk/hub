# Requirements: Hub

**Defined:** 2026-03-20
**Core Value:** Unified environment for development teams where task tracker, testing, communication, and calls are connected in one interface -- no switching between a dozen separate tools.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Authentication & Users

- [ ] **AUTH-01**: User can sign in via Magic Link (email) without password
- [ ] **AUTH-02**: User can sign in via Telegram without password
- [ ] **AUTH-03**: Admin can configure registration mode: open or invite-only
- [ ] **AUTH-04**: User can edit profile: name, avatar, email, Telegram handle, position, department, phone, timezone, language, birthday
- [ ] **AUTH-05**: System supports two roles: Admin and Member
- [ ] **AUTH-06**: User can have multiple active sessions (desktop + mobile), view session list, and terminate individual sessions
- [ ] **AUTH-07**: User receives notification when a new session is created on their account
- [ ] **AUTH-08**: New user goes through onboarding flow: fill profile, overview tour, choose theme/language, create or join a project

### Projects

- [ ] **PROJ-01**: Admin and Member can create projects; creation auto-generates a tracker board, TMS space, project chat, channel, and dashboard
- [ ] **PROJ-02**: Project Owner can manage participant roles, workflow/statuses, teams, and access controls
- [ ] **PROJ-03**: Projects have fixed roles (Owner, Developer, Tester, Viewer) plus custom roles with granular RBAC permissions
- [ ] **PROJ-04**: Users can be invited via link; teams (groups of people) can be added as a unit
- [ ] **PROJ-05**: Project templates available: Empty, Software Dev, QA, Kanban Simple, plus custom templates and saving a project as template
- [ ] **PROJ-06**: Projects can be archived (read-only alternative to deletion)
- [ ] **PROJ-07**: Deleted projects go to trash (30-day retention) with option to export backup before permanent deletion

### Task Tracker

- [ ] **TASK-01**: 9 issue templates: Bug, Task, Story, Epic, Sub-task, Improvement, Tech Debt, Spike, Release
- [ ] **TASK-02**: Issues support fixed fields plus custom fields
- [ ] **TASK-03**: Unlimited nesting depth for parent/child task hierarchy
- [ ] **TASK-04**: Issue links: parent/child, blocks, relates to, duplicates, clones, causes, plus custom link types
- [ ] **TASK-05**: Project-prefixed sequential numbering (HUB-1, MOBILE-1); prefix set at project creation
- [ ] **TASK-06**: Project Owner can create custom workflow with statuses and transitions
- [ ] **TASK-07**: Comments on issues support @mentions, internal comments (role-restricted visibility), quoting, and reactions
- [ ] **TASK-08**: Full change log on every issue (who changed what and when)
- [ ] **TASK-09**: Time tracking: estimate, actual, start/stop timer
- [ ] **TASK-10**: Labels at project level with colors, label groups; unified label system shared across issues, test cases, and messages
- [ ] **TASK-11**: File attachments on issues with image/PDF/video preview, Ctrl+V paste, drag-and-drop upload

### Kanban

- [ ] **KANB-01**: Kanban board columns map to workflow statuses
- [ ] **KANB-02**: WIP limits per column
- [ ] **KANB-03**: Swimlanes (by assignee, priority, or type)
- [ ] **KANB-04**: Multiple boards per project
- [ ] **KANB-05**: Board filters: my tasks, bugs only, by priority

### Scrum

- [ ] **SCRM-01**: Create sprints with start/end dates, drag tasks from backlog, start/complete sprint
- [ ] **SCRM-02**: Burndown chart and velocity chart
- [ ] **SCRM-03**: Sprint retrospective notes attached to sprint
- [ ] **SCRM-04**: Auto-carry incomplete tasks to next sprint
- [ ] **SCRM-05**: Planning Poker: Fibonacci scale, votes hidden until reveal

### Messenger

- [ ] **MSG-01**: Direct messages (1-on-1) and group chats
- [ ] **MSG-02**: Public channels (anyone can join) and private channels (invite-only)
- [ ] **MSG-03**: Auto-created project channels (#project-hub) plus topic channels (#hub-backend, #hub-bugs)
- [ ] **MSG-04**: Read-only channels (bot posts, users read -- e.g., #hub-ci)
- [ ] **MSG-05**: Threaded replies (reply opens a conversation branch)
- [ ] **MSG-06**: Emoji reactions on messages
- [ ] **MSG-07**: Rich text formatting: bold, italic, code, lists, links
- [ ] **MSG-08**: Edit and delete own messages; channel Owner/Admin can delete any message
- [ ] **MSG-09**: Message forwarding, pinning, and bookmarks (saved messages)
- [ ] **MSG-10**: Voice messages (audio recording in chat)
- [ ] **MSG-11**: Link previews: issues, test cases, test runs, GitHub PR/Issue plus external URL previews (OG meta)
- [ ] **MSG-12**: Global and per-channel message search
- [ ] **MSG-13**: User status: online/offline/away (auto-detected) plus custom status text

### TMS (Test Management)

- [ ] **TMS-01**: Test cases with classic format (steps + expected result) and BDD format (Given/When/Then)
- [ ] **TMS-02**: Test case structure: folders/sections, tags, components, versioning with history and rollback
- [ ] **TMS-03**: GitHub linking: match by annotation ID (@TestCase("HUB-TC-42")), display linked code, sync status, webhook on change
- [ ] **TMS-04**: Test runs: manual, automatic (Allure), mixed; re-open failed; assign to testers
- [ ] **TMS-05**: Screenshot and comment per test result step
- [ ] **TMS-06**: Bidirectional issue linking: create bug from failed test case, requirements coverage, traceability matrix
- [ ] **TMS-07**: Test environments: label (dev/staging/prod), configurations (browser/OS), cross-environment comparison, filtered history
- [ ] **TMS-08**: Allure-level reporting: trends, flaky tests, defect categories, run comparison, dashboards
- [ ] **TMS-09**: Import test cases: CSV, JSON, XML, Excel with universal field mapping
- [ ] **TMS-10**: Export test cases and results: CSV, JSON, XML, PDF

### Allure Integration

- [ ] **ALLU-01**: CI generates Allure Report HTML; Hub provides parser that reads allure-results JSON and POSTs to TMS API
- [ ] **ALLU-02**: Matching test results to test cases by Allure ID or test name
- [ ] **ALLU-03**: Run history, statistics, and Telegram notifications for test run completion

### Voice Calls

- [ ] **CALL-01**: Audio-only calls plus screen sharing (no webcams), up to 50 participants
- [ ] **CALL-02**: Scheduled and recurring calls (e.g., daily standup)
- [ ] **CALL-03**: Mute/unmute, raise hand, screen/window picker, multiple simultaneous screen shares
- [ ] **CALL-04**: Recording by button press (audio + screen share); all participants notified; saved to SeaweedFS; link posted to channel
- [ ] **CALL-05**: Recordings attached to project (dedicated Recordings section)
- [ ] **CALL-06**: Ephemeral call chat (disappears after call ends)

### Dashboards

- [ ] **DASH-01**: Personal dashboard: my tasks, unread messages, upcoming calls, test runs, deadlines, activity, quick actions, filters, chats, time tracking -- customizable widgets with drag-and-drop and resizing
- [ ] **DASH-02**: Project dashboard: task statuses, burndown, velocity, bugs, team activity, tests, trends, coverage, flaky tests, unassigned, overdue, commits -- customizable with templates (Dev/QA/Manager)
- [ ] **DASH-03**: Admin dashboard: users, projects, tasks, activity (DAU/WAU/MAU), storage usage, service health, top projects/users, error logs

### Notifications

- [ ] **NOTF-01**: Triggers: assignment, update, @mention, status change, chat message, call, test run, sprint, deadline, blocker
- [ ] **NOTF-02**: Channels: in-app (bell icon) plus Telegram bot
- [ ] **NOTF-03**: Preferences: configure by type, by channel, and by project
- [ ] **NOTF-04**: Quiet hours with urgent notifications breaking through

### GitHub Integration

- [ ] **GHUB-01**: Link PRs and commits to issues with preview cards
- [ ] **GHUB-02**: Auto-link: "HUB-123" in commit message auto-attaches to issue
- [ ] **GHUB-03**: PR status displayed on issue (open/merged/closed)
- [ ] **GHUB-04**: Auto-transition issue status when PR is merged
- [ ] **GHUB-05**: Create branch from issue (hub-123-fix-login naming)

### Filters & Search

- [ ] **SRCH-01**: Filters on issues: status, priority, assignee, type; saved and shareable filters
- [ ] **SRCH-02**: Ctrl+K global search: issues, test cases, users, projects, channels, files
- [ ] **SRCH-03**: JQL-like query language for advanced filtering

### API & Integrations

- [ ] **API-01**: REST API with Swagger documentation
- [ ] **API-02**: API key generation in user profile
- [ ] **API-03**: Webhooks (POST to external URL on events)
- [ ] **API-04**: hub-cli: CLI tool for CI/CD (push-results, create-bug, update-task)

### Import/Export

- [ ] **IMEX-01**: Universal CSV/JSON import for issues with field mapping
- [ ] **IMEX-02**: Import test cases from Excel
- [ ] **IMEX-03**: Export: issues (CSV/Excel), test runs (PDF), sprints (PDF), full project backup (ZIP)

### Audit & Security

- [ ] **AUDT-01**: Full audit log: login/logout, issues, projects, TMS, messenger -- Admin can view with filters
- [ ] **AUDT-02**: Configurable limits: file size, projects per user, storage per project -- Admin manages
- [ ] **AUDT-03**: 30-day trash for issues, test cases, projects, files; Admin/Owner can restore
- [ ] **AUDT-04**: User deletion: deactivation, task reassignment, full data deletion (GDPR option)

### Backups

- [ ] **BKUP-01**: Weekly automatic backups (PostgreSQL + MongoDB + SeaweedFS)
- [ ] **BKUP-02**: Retain last N backups (configurable)
- [ ] **BKUP-03**: Restore from backup via Admin panel

### UI/UX

- [ ] **UI-01**: Collapsible left sidebar plus top bar (search, notifications, profile)
- [ ] **UI-02**: Mobile layout: bottom navigation, simplified Kanban (list view), mobile dashboards, mobile call UI
- [ ] **UI-03**: Dark, light, and system theme plus custom accent color
- [ ] **UI-04**: Russian and English with language switcher in profile
- [ ] **UI-05**: Keyboard shortcuts: Ctrl+K (search), Ctrl+N (new issue)
- [ ] **UI-06**: PWA: install to phone, offline cache, Web Push notifications
- [ ] **UI-07**: Error handling: toast notifications, request ID, retry, offline mode, 404/500/403 pages
- [ ] **UI-08**: File attachments everywhere: any file type, image/PDF/video preview, Ctrl+V paste, drag-and-drop
- [ ] **UI-09**: Pagination: pages for issues, infinite scroll for chat, virtualization for large lists, server-side filtering

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Query

- **V2-01**: JQL-like query language with autocomplete and saved queries
- **V2-02**: Traceability matrix view (requirements to test cases to results)

### Advanced Project Management

- **V2-03**: Project templates marketplace (import/export templates across instances)
- **V2-04**: Import from Jira (project structure, issues, attachments)
- **V2-05**: Import from TestRail (test cases, suites, runs)

### Advanced Admin

- **V2-06**: Admin dashboard with DAU/WAU/MAU analytics, service health monitoring
- **V2-07**: CLI tool for CI/CD automation (hub-cli)

### Advanced TMS

- **V2-08**: Test environments with browser/OS configurations and cross-environment comparison
- **V2-09**: Custom fields on issues

### Advanced Kanban

- **V2-10**: Swimlanes by assignee, priority, or type
- **V2-11**: WIP limits per column

### PWA & Offline

- **V2-12**: Full PWA offline mode with sync on reconnect

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Video calls with webcams | Deliberate constraint -- voice + screen share covers 90% of needs |
| Native mobile apps (App Store/Google Play) | PWA is sufficient for internal tools |
| Email notifications | Telegram bot replaces this for target audience |
| Billing/subscriptions | Internal tool, not SaaS |
| SSO/LDAP/OAuth | Magic Link + Telegram sufficient for v1 |
| Gantt charts | Kanban + Scrum cover project management needs |
| Video/audio messages in messenger (beyond voice) | Excessive for v1 |
| Remote control (control someone else's screen) | High complexity, not critical |
| Real-time collaborative editing | High complexity, not critical for v1 |
| AI features | Out of scope for initial product |
| Plugin marketplace | Not needed for internal tool |
| Wiki/knowledge base | Separate concern, not core value |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Pending |
| AUTH-02 | Phase 1 | Pending |
| AUTH-03 | Phase 1 | Pending |
| AUTH-04 | Phase 1 | Pending |
| AUTH-05 | Phase 1 | Pending |
| AUTH-06 | Phase 1 | Pending |
| AUTH-07 | Phase 1 | Pending |
| AUTH-08 | Phase 1 | Pending |
| PROJ-01 | Phase 1 | Pending |
| PROJ-02 | Phase 1 | Pending |
| PROJ-03 | Phase 1 | Pending |
| PROJ-04 | Phase 1 | Pending |
| PROJ-05 | Phase 7 | Pending |
| PROJ-06 | Phase 1 | Pending |
| PROJ-07 | Phase 1 | Pending |
| TASK-01 | Phase 2 | Pending |
| TASK-02 | Phase 2 | Pending |
| TASK-03 | Phase 2 | Pending |
| TASK-04 | Phase 2 | Pending |
| TASK-05 | Phase 2 | Pending |
| TASK-06 | Phase 2 | Pending |
| TASK-07 | Phase 2 | Pending |
| TASK-08 | Phase 2 | Pending |
| TASK-09 | Phase 2 | Pending |
| TASK-10 | Phase 2 | Pending |
| TASK-11 | Phase 2 | Pending |
| KANB-01 | Phase 2 | Pending |
| KANB-02 | Phase 7 | Pending |
| KANB-03 | Phase 7 | Pending |
| KANB-04 | Phase 2 | Pending |
| KANB-05 | Phase 2 | Pending |
| SCRM-01 | Phase 7 | Pending |
| SCRM-02 | Phase 7 | Pending |
| SCRM-03 | Phase 7 | Pending |
| SCRM-04 | Phase 7 | Pending |
| SCRM-05 | Phase 7 | Pending |
| MSG-01 | Phase 3 | Pending |
| MSG-02 | Phase 3 | Pending |
| MSG-03 | Phase 3 | Pending |
| MSG-04 | Phase 3 | Pending |
| MSG-05 | Phase 3 | Pending |
| MSG-06 | Phase 3 | Pending |
| MSG-07 | Phase 3 | Pending |
| MSG-08 | Phase 3 | Pending |
| MSG-09 | Phase 3 | Pending |
| MSG-10 | Phase 3 | Pending |
| MSG-11 | Phase 7 | Pending |
| MSG-12 | Phase 3 | Pending |
| MSG-13 | Phase 3 | Pending |
| TMS-01 | Phase 4 | Pending |
| TMS-02 | Phase 4 | Pending |
| TMS-03 | Phase 4 | Pending |
| TMS-04 | Phase 4 | Pending |
| TMS-05 | Phase 4 | Pending |
| TMS-06 | Phase 4 | Pending |
| TMS-07 | Phase 7 | Pending |
| TMS-08 | Phase 4 | Pending |
| TMS-09 | Phase 4 | Pending |
| TMS-10 | Phase 4 | Pending |
| ALLU-01 | Phase 4 | Pending |
| ALLU-02 | Phase 4 | Pending |
| ALLU-03 | Phase 4 | Pending |
| CALL-01 | Phase 5 | Pending |
| CALL-02 | Phase 7 | Pending |
| CALL-03 | Phase 5 | Pending |
| CALL-04 | Phase 7 | Pending |
| CALL-05 | Phase 7 | Pending |
| CALL-06 | Phase 5 | Pending |
| DASH-01 | Phase 6 | Pending |
| DASH-02 | Phase 6 | Pending |
| DASH-03 | Phase 6 | Pending |
| NOTF-01 | Phase 6 | Pending |
| NOTF-02 | Phase 6 | Pending |
| NOTF-03 | Phase 6 | Pending |
| NOTF-04 | Phase 6 | Pending |
| GHUB-01 | Phase 7 | Pending |
| GHUB-02 | Phase 7 | Pending |
| GHUB-03 | Phase 7 | Pending |
| GHUB-04 | Phase 7 | Pending |
| GHUB-05 | Phase 7 | Pending |
| SRCH-01 | Phase 6 | Pending |
| SRCH-02 | Phase 6 | Pending |
| SRCH-03 | Phase 7 | Pending |
| API-01 | Phase 7 | Pending |
| API-02 | Phase 7 | Pending |
| API-03 | Phase 7 | Pending |
| API-04 | Phase 7 | Pending |
| IMEX-01 | Phase 7 | Pending |
| IMEX-02 | Phase 4 | Pending |
| IMEX-03 | Phase 7 | Pending |
| AUDT-01 | Phase 6 | Pending |
| AUDT-02 | Phase 6 | Pending |
| AUDT-03 | Phase 6 | Pending |
| AUDT-04 | Phase 6 | Pending |
| BKUP-01 | Phase 7 | Pending |
| BKUP-02 | Phase 7 | Pending |
| BKUP-03 | Phase 7 | Pending |
| UI-01 | Phase 1 | Pending |
| UI-02 | Phase 7 | Pending |
| UI-03 | Phase 7 | Pending |
| UI-04 | Phase 7 | Pending |
| UI-05 | Phase 6 | Pending |
| UI-06 | Phase 7 | Pending |
| UI-07 | Phase 1 | Pending |
| UI-08 | Phase 1 | Pending |
| UI-09 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 95 total
- Mapped to phases: 95
- Unmapped: 0

---
*Requirements defined: 2026-03-20*
*Last updated: 2026-03-20 after initial definition*
