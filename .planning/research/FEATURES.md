# Feature Research

**Domain:** Team Collaboration Hub (Task Tracker + Messenger + TMS + Voice Calls)
**Researched:** 2026-03-20
**Confidence:** HIGH

## Feature Landscape

Analysis based on: Jira, Linear, Plane.so, Huly (trackers); Slack, Mattermost, Zulip (messengers); TestRail, Allure TestOps (TMS); Google Meet, LiveKit (calls). Hub's core differentiator is cross-module traceability in a single self-hosted product.

---

### Table Stakes (Users Expect These)

Features users assume exist. Missing any of these = users immediately look for alternatives.

#### Task Tracker — Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Issue CRUD with rich text | Every tracker has this since Bugzilla | LOW | Plate editor already chosen; support bold/italic/code/lists/links |
| Custom fields (text, number, date, select) | Jira/Linear/Plane all support; teams need domain-specific data | MEDIUM | Start with 5-6 field types, extensible schema |
| Issue types (Bug, Task, Story, Epic) | Universal taxonomy across all trackers | LOW | Configurable per project, icon + color |
| Assignee, priority, status, due date | Core triage fields in every tracker | LOW | Priority: Critical/High/Medium/Low/None |
| Kanban board with drag-and-drop | Expected since Trello (2011); Jira/Linear/Plane all have it | MEDIUM | pragmatic-dnd chosen; columns = workflow statuses |
| Customizable workflow (statuses + transitions) | Jira popularized this; Linear simplified it; both expected | MEDIUM | Visual editor for status transitions; default templates |
| Sprint management (create, plan, execute, close) | Standard Scrum support; Jira/Plane/Huly all have it | MEDIUM | Backlog view, sprint view, auto-carry-over of incomplete items |
| Burndown chart | Basic Scrum artifact every PM expects | MEDIUM | Tremor for rendering; data = story points over time |
| Comments with @mentions | Every tracker since Jira 2.0 | LOW | Link to messenger mentions system |
| Activity log / history | "Who changed what when" — audit necessity | LOW | Store diffs per field change |
| File attachments | Universal expectation | LOW | MinIO storage, presigned URLs, image preview |
| Filters and saved views | Jira/Linear/Plane all provide | MEDIUM | Status/assignee/priority/type filters + save + share |
| Global search (Ctrl+K) | Linear made this the gold standard; now expected everywhere | MEDIUM | Search across issues, messages, test cases, users |
| Labels/tags with colors | Every tracker supports categorization | LOW | Project-scoped, reusable across modules |
| Parent-child relationships | Sub-tasks are universal; Jira/Linear/Plane all have hierarchies | MEDIUM | Unlimited nesting with Ent graph traversal |
| Issue linking (blocks, relates, duplicates) | Jira standard; teams need dependency tracking | LOW | Bidirectional links with type |
| Bulk operations | Jira/Linear allow multi-select + batch update | MEDIUM | Select multiple issues, change status/assignee/priority |
| Project-prefixed IDs (HUB-123) | Jira convention, universally expected | LOW | Auto-incrementing counter per project |

#### Messenger — Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Direct messages (1:1) | Every messenger since ICQ | LOW | WebSocket-based real-time |
| Group chats | Basic collaboration unit | LOW | Create, add/remove members, leave |
| Channels (public + private) | Slack standard since 2013; Mattermost/Zulip/Huly all have | LOW | Join/leave, topic, description |
| Threaded replies | Slack popularized; now expected in all team messengers | MEDIUM | Thread opens in side panel or inline; unread tracking per thread |
| Message editing and deletion | Slack/Mattermost standard | LOW | Edit window (or unlimited for own messages), soft delete |
| Emoji reactions | Universal since Slack; social feedback without noise | LOW | Standard emoji set, click to toggle |
| Rich text formatting | Markdown-style: bold, italic, code, lists | LOW | Reuse Plate or simpler markdown parser |
| File sharing + image preview | Every messenger supports drag-and-drop files | LOW | MinIO + thumbnail generation |
| Typing indicators | Real-time presence signal; Slack/Mattermost have it | LOW | WebSocket broadcast per channel |
| Online/offline status | Presence awareness is table stakes | LOW | Redis TTL for heartbeat; auto-away after idle |
| Unread counts and badges | Must know where new messages are | MEDIUM | Per-channel and per-thread unread counters |
| Message search | Finding past conversations is essential | MEDIUM | Full-text search in MongoDB; filters by channel/user/date |
| Link previews (OG meta) | Slack standard; users expect URL unfurling | MEDIUM | Server-side fetch of OG tags, cache result |
| Pin and bookmark messages | Slack standard for important reference messages | LOW | Per-channel pinned list; personal bookmarks |

#### TMS — Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Test case CRUD (steps + expected result) | Core TMS functionality since TestRail | MEDIUM | Step editor with ordering, expected/actual result |
| Folder/suite hierarchy | TestRail/Allure TestOps both organize by tree | MEDIUM | Nested folders with drag-and-drop reordering |
| Test runs (manual execution) | Core TMS workflow: create run, execute, record pass/fail | MEDIUM | Run = snapshot of selected cases; per-case status |
| Test run results with evidence | Screenshot/comment per step is TestRail standard | MEDIUM | Attach files to individual step results |
| Pass/Fail/Blocked/Skipped statuses | Standard test execution statuses | LOW | Configurable status set with colors |
| Link test case to issue | Bidirectional traceability is core TMS value | MEDIUM | Many-to-many relationship; create bug from failed case |
| Import test cases (CSV, Excel) | Migration path from existing tools is essential | MEDIUM | Column mapping UI; preview before import |
| Basic reporting (pass rate, run history) | TestRail/Allure TestOps both provide | MEDIUM | Aggregate stats per run, trend over time |

#### Voice Calls — Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Audio call (multi-participant) | Basic conferencing; Google Meet/Zoom/Teams | HIGH | LiveKit handles SFU; up to 50 participants |
| Mute/unmute | Universal call control | LOW | LiveKit SDK provides this out of the box |
| Screen sharing | Google Meet/Zoom standard | MEDIUM | LiveKit native support; select screen/window |
| Join via link | One-click join without install | LOW | Room token generation, URL-based join |
| Participant list | See who's in the call | LOW | LiveKit room state subscription |

#### Cross-Module — Table Stakes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| User authentication | Login/signup is foundational | MEDIUM | Magic Link + Telegram auth (passwordless) |
| Role-based access control | Multi-user products need permissions | HIGH | System roles (Admin, Member) + project roles (Owner, Dev, Tester, Viewer) + custom RBAC |
| Notifications (in-app) | Must know about assignments, mentions, updates | MEDIUM | Bell icon, notification center, mark read/unread |
| User profiles | Name, avatar, contact info | LOW | Standard profile page + settings |
| Project creation and membership | Organizational unit for all modules | MEDIUM | Create project = auto-create tracker + TMS + channel + dashboard |
| Dark/light theme | Modern UI expectation | LOW | CSS variables + Tailwind dark mode |
| i18n (Russian + English) | Bilingual requirement from PROJECT.md | MEDIUM | react-intl or i18next; all strings externalized |

---

### Differentiators (Competitive Advantage)

Features that set Hub apart. These align with the core value: "unified environment without switching between a dozen tools."

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Unified platform (tracker + messenger + TMS + calls)** | No competitor except Huly combines all four; Huly lacks TMS. This is THE differentiator. | N/A (architectural) | Integration between modules is the product, not individual features |
| **Deep cross-module linking** | Mention HUB-123 in chat and get a rich preview card; create a bug from a failed test case that auto-links; see test coverage on an issue. No tool does all three. | MEDIUM | Smart entity references across all modules |
| **Allure results ingestion (CI pipeline)** | Allure TestOps costs $30K/yr. Hub accepts allure-results JSON for free, matches to test cases, builds reports. | HIGH | Custom parser, matching algorithm, history tracking |
| **Auto-channels per project** | When you create a project, #project-hub channel appears. CI events, test run results, and sprint changes post automatically. Huly does this; Jira+Slack requires Zapier. | MEDIUM | Event-driven NATS messages to auto-channel |
| **Traceability matrix (requirement to test case to code)** | TestRail has basic linking; Allure TestOps links to automation. Hub links issue to test case to GitHub PR to CI result in one view. | HIGH | Multi-entity graph query; dedicated matrix UI |
| **Planning Poker built-in** | External tool (PlanningPoker.com) for most teams. Built-in = no context switch. | MEDIUM | WebSocket room, Fibonacci scale, reveal mechanic |
| **Voice-only calls (no video)** | Deliberate constraint. Lower bandwidth, simpler UX, faster to join. Teams doing daily standups don't need video. | LOW (LiveKit handles it) | Simpler than full video; screen share covers visual needs |
| **Telegram bot for auth + notifications** | Magic Link via Telegram is frictionless. Bot also pushes notifications. No email setup needed. | MEDIUM | telego bot; deep links for auth flow |
| **Read-only CI channels** | #project-ci auto-posts build/test results. No manual copy-paste from Jenkins/GitHub Actions. | LOW | Webhook receiver posts formatted messages |
| **BDD test cases (Given/When/Then)** | TestRail is steps-only. Allure TestOps supports BDD but it's paid. Hub supports both formats. | LOW | Alternative editor mode for test case body |
| **Test case versioning with rollback** | TestRail versions test cases but rollback is clunky. Hub makes it explicit: view history, diff, restore. | MEDIUM | Store snapshots on edit; diff view |
| **Call recording with project binding** | Google Meet recordings go to Drive (disconnected). Hub recordings are attached to the project, linked to a channel message. | HIGH | LiveKit Egress, MinIO storage, metadata linking |
| **JQL-like query language** | Jira power users love JQL. Linear has no equivalent. Offering this for advanced users is a differentiator among mid-size tools. | HIGH | Parser + AST + query execution; can defer to v1.x |
| **Custom workflow per project** | Jira does this but it's admin-heavy. Linear is opinionated (no custom). Hub: project owner builds workflows visually. | MEDIUM | Visual status/transition editor |
| **Quiet hours with urgent override** | Slack has DND but no urgency bypass. Hub lets urgent notifications punch through quiet hours. | LOW | Notification priority flag + user schedule |
| **Project templates (save + reuse)** | Jira recently added this. Hub allows saving an entire project config (workflow, roles, channels, TMS structure) as a template. | MEDIUM | Serialize project config to template; apply on create |
| **Scheduled + recurring calls** | Google Calendar integration is typical. Hub has built-in scheduling with recurrence (daily standup). | MEDIUM | Cron-like schedule, auto-create room, reminder notification |
| **Voice messages in chat** | Telegram popularized this. Slack added it. Useful for async communication. | MEDIUM | MediaRecorder API, MinIO storage, waveform preview |

---

### Anti-Features (Deliberately NOT Building)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Video calls with webcams** | "Everyone expects video" | Massively increases bandwidth, complexity, and LiveKit resource usage. For 10-50 person teams doing standups, voice + screen share covers 90% of use cases. | Voice + screen share. If video is truly needed, teams use Google Meet/Zoom for that one meeting. |
| **Native mobile apps (iOS/Android)** | "People want App Store apps" | Two additional codebases (or React Native); app store approval delays; 6+ months extra work. PWA covers mobile use cases for internal tools. | PWA with push notifications, offline cache, installable. |
| **Email notifications** | "Standard notification channel" | Requires email infrastructure (SMTP, deliverability, spam management). Telegram bot is simpler, faster, and already integrated for auth. | Telegram bot notifications + in-app notifications. |
| **SSO/LDAP/OAuth providers** | "Enterprise needs it" | Significant auth complexity for v1. Target is 10-50 person teams, not enterprise. Magic Link + Telegram is sufficient and modern. | Magic Link + Telegram. Add OAuth in v2 if enterprise demand emerges. |
| **Gantt charts** | "Project managers expect it" | Complex to implement well; Kanban + Sprint views cover planning needs for dev teams. Gantt is more relevant to construction/waterfall. | Kanban board + sprint timeline + burndown chart. |
| **Real-time collaborative editing (Google Docs style)** | "Modern collaboration" | CRDT/OT implementation is extremely complex. Task descriptions and test cases don't need simultaneous editing. | Last-write-wins with conflict detection. Lock editing if needed. |
| **AI agent features** | "Plane and Linear are adding AI" | Requires LLM infrastructure, significant cost, unpredictable quality. Distraction from core value (integration, not AI). | None for v1. Revisit when core platform is stable. |
| **Plugin/extension marketplace** | "Jira has 3000+ plugins" | Plugin API design is a product in itself. Premature extensibility creates maintenance burden. | Well-designed REST API + webhooks. Third parties integrate externally. |
| **Time zone-aware scheduling (world clock)** | "Distributed teams need it" | Over-engineering for 10-50 person teams likely in same time zone. | Store all times in UTC, display in user's local timezone. Basic, not fancy. |
| **Wikis / knowledge base / Notion-like docs** | "Huly has it, Plane has Pages" | Separate product domain. Doing it well requires block editor, nested pages, permissions. Distracts from core tracker+messenger+TMS+calls value. | Link to external wiki (Notion, Confluence). Consider for v2+ if demand exists. |
| **Billing / subscriptions / multi-tenancy** | "SaaS needs it" | Hub is an internal tool, not SaaS. Building billing adds months of work with zero value for the target use case. | Single-tenant self-hosted deployment. |
| **Remote desktop control** | "Useful for support" | Complex, security-sensitive, niche use case. Dedicated tools (AnyDesk, TeamViewer) do this far better. | Screen share is sufficient for code review / demos. |

---

## Feature Dependencies

```
Authentication (Magic Link + Telegram)
    |
    +---> User Profiles + Roles (RBAC)
              |
              +---> Projects (create, membership, permissions)
                        |
                        +---> Task Tracker (issues, boards, sprints)
                        |         |
                        |         +---> Kanban Board
                        |         +---> Sprint Management ---> Burndown/Velocity Charts
                        |         +---> Custom Workflows
                        |         +---> Time Tracking
                        |
                        +---> Messenger (channels, DMs, threads)
                        |         |
                        |         +---> Auto-channels per project
                        |         +---> Link previews (requires entity reference system)
                        |         +---> Voice Messages
                        |
                        +---> TMS (test cases, runs, results)
                        |         |
                        |         +---> Allure Integration (requires TMS + CI webhook)
                        |         +---> Traceability Matrix (requires Tracker + TMS linking)
                        |         +---> Test Case Versioning
                        |
                        +---> Voice Calls (LiveKit rooms)
                        |         |
                        |         +---> Screen Share
                        |         +---> Call Recording (requires MinIO)
                        |         +---> Scheduled Calls
                        |
                        +---> Dashboards (requires data from all modules)
                        +---> Notifications (requires events from all modules)
                        +---> GitHub Integration (requires Tracker issues)

File Storage (MinIO)
    |
    +---> Attachments (all modules)
    +---> Call Recordings
    +---> Voice Messages
    +---> Avatars

Real-time Infrastructure (NATS + WebSocket)
    |
    +---> Messenger (message delivery)
    +---> Notifications (event distribution)
    +---> Kanban (live board updates)
    +---> Presence / Online Status
    +---> Planning Poker
```

### Dependency Notes

- **Authentication is the root dependency:** Nothing works without login. Build first.
- **Projects gate all modules:** Tracker, TMS, Messenger channels, and Calls all live within a project context. Project CRUD must precede module development.
- **Real-time infrastructure (NATS + WebSocket) is a horizontal dependency:** Messenger, notifications, live board updates, and presence all need it. Build the real-time layer early.
- **MinIO is a horizontal dependency:** Attachments, recordings, voice messages, and avatars all need object storage. Set up early.
- **Dashboards depend on all modules:** They aggregate data from tracker, TMS, and messenger. Build last.
- **Traceability matrix requires both Tracker and TMS:** Cannot build until both modules have linking capability.
- **Allure integration requires TMS + external webhook:** Depends on TMS test case model and a webhook/API endpoint for CI to push results.
- **GitHub integration requires Tracker:** Auto-linking commit messages to issues needs the issue model to exist.

---

## MVP Definition

### Launch With (v1.0)

Minimum viable product that demonstrates the core value (unified platform with cross-module traceability).

- [ ] **Auth: Magic Link + Telegram login** -- Foundation; passwordless is simpler than password-based
- [ ] **User profiles + system roles (Admin/Member)** -- Basic identity and access control
- [ ] **Projects with membership** -- Organizational container for all modules
- [ ] **Task Tracker: CRUD, Kanban board, basic workflow** -- Core module #1; most immediate daily use
- [ ] **Issue types, priorities, assignees, labels** -- Basic triage capability
- [ ] **Parent-child issues + issue linking** -- Hierarchy is table stakes
- [ ] **Messenger: DMs, channels, threads** -- Core module #2; replaces Slack
- [ ] **Rich text, reactions, file sharing** -- Messenger table stakes
- [ ] **Auto-channel per project** -- First differentiator: project activity in one place
- [ ] **TMS: Test cases (steps), folders, manual test runs** -- Core module #3; replaces TestRail
- [ ] **Link test case to issue (bidirectional)** -- First cross-module traceability
- [ ] **Voice calls: audio + screen share** -- Core module #4; replaces Meet
- [ ] **In-app notifications** -- Must know about mentions, assignments, status changes
- [ ] **Global search (Ctrl+K)** -- Cross-module search from day one
- [ ] **File storage (MinIO) for all modules** -- Attachments, screenshots in test results
- [ ] **Dark/light theme, i18n (RU/EN)** -- UX requirements from PROJECT.md

### Add After Validation (v1.x)

Features to add once the four core modules work reliably.

- [ ] **Sprint management + burndown** -- When teams start using tracker for Scrum (not just Kanban)
- [ ] **Allure results ingestion** -- When teams have CI pipelines pushing test results
- [ ] **Telegram bot notifications** -- When in-app notifications aren't enough
- [ ] **GitHub integration (PR linking, auto-status)** -- When dev workflow tightens
- [ ] **Call recording** -- When teams want to reference past meetings
- [ ] **Custom workflows (visual editor)** -- When default workflow is too restrictive
- [ ] **Project-level RBAC (custom roles)** -- When teams need fine-grained permissions
- [ ] **BDD test cases (Given/When/Then)** -- When QA teams request alternative format
- [ ] **Test case versioning** -- When test suites mature and need history
- [ ] **Planning Poker** -- When sprint planning sessions begin
- [ ] **Scheduled/recurring calls** -- When daily standups are routine
- [ ] **Voice messages** -- When async communication patterns emerge
- [ ] **Bulk operations on issues** -- When issue volume makes one-by-one editing painful
- [ ] **Saved/shared filters** -- When teams standardize their views
- [ ] **Link previews in chat (HUB-123 cards)** -- Cross-module UX polish

### Future Consideration (v2+)

Features to defer until the platform is stable and adoption is proven.

- [ ] **JQL-like query language** -- Power user feature; basic filters suffice initially
- [ ] **Traceability matrix view** -- Requires mature tracker + TMS data; complex UI
- [ ] **Project templates (save + reuse)** -- Useful at scale; not needed with few projects
- [ ] **Admin dashboard (DAU/WAU, storage, health)** -- Operational tooling for production scale
- [ ] **Webhook system (outbound)** -- External integration point; API keys + REST API first
- [ ] **CLI tool (hub-cli)** -- Developer convenience; REST API covers CI/CD needs initially
- [ ] **PWA with offline mode** -- Mobile experience polish; web app works initially
- [ ] **Audit log with filters** -- Compliance feature; basic activity log in v1 is enough
- [ ] **Quiet hours + urgent override** -- Notification sophistication for larger teams
- [ ] **Import from Jira/TestRail** -- Migration tooling once there are teams to migrate
- [ ] **Custom fields on issues** -- Extensibility; fixed fields cover 80% of use cases in v1
- [ ] **Environments in TMS (dev/staging/prod)** -- Advanced TMS feature for mature testing orgs
- [ ] **Swimlanes on Kanban** -- Nice-to-have board customization
- [ ] **WIP limits** -- Process enforcement; teams can self-regulate initially

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Auth (Magic Link + Telegram) | HIGH | MEDIUM | P1 |
| User profiles + roles | HIGH | LOW | P1 |
| Projects with membership | HIGH | MEDIUM | P1 |
| Task CRUD + Kanban board | HIGH | MEDIUM | P1 |
| Custom workflow (basic) | HIGH | MEDIUM | P1 |
| Issue hierarchy + linking | HIGH | MEDIUM | P1 |
| Messenger (DMs, channels, threads) | HIGH | HIGH | P1 |
| TMS (cases, runs, results) | HIGH | HIGH | P1 |
| Voice calls + screen share | HIGH | HIGH | P1 |
| In-app notifications | HIGH | MEDIUM | P1 |
| Global search (Ctrl+K) | HIGH | MEDIUM | P1 |
| Auto-channels per project | HIGH | LOW | P1 |
| Cross-module entity linking | HIGH | MEDIUM | P1 |
| Sprint management + burndown | MEDIUM | MEDIUM | P2 |
| Allure results ingestion | HIGH | HIGH | P2 |
| GitHub integration | MEDIUM | MEDIUM | P2 |
| Telegram notifications | MEDIUM | MEDIUM | P2 |
| Call recording | MEDIUM | HIGH | P2 |
| Custom RBAC | MEDIUM | HIGH | P2 |
| Planning Poker | LOW | MEDIUM | P2 |
| BDD test cases | LOW | LOW | P2 |
| Test case versioning | MEDIUM | MEDIUM | P2 |
| JQL-like queries | LOW | HIGH | P3 |
| Traceability matrix | MEDIUM | HIGH | P3 |
| Project templates | LOW | MEDIUM | P3 |
| Admin dashboard | LOW | MEDIUM | P3 |
| CLI tool | LOW | MEDIUM | P3 |
| PWA offline mode | LOW | HIGH | P3 |
| Import from Jira/TestRail | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch (v1.0)
- P2: Should have, add after core validation (v1.x)
- P3: Nice to have, future consideration (v2+)

---

## Competitor Feature Analysis

### Task Trackers

| Feature | Jira | Linear | Plane.so | Huly | Hub Approach |
|---------|------|--------|----------|------|-------------|
| Kanban board | Full (swimlanes, WIP) | Minimal, fast | Full | Full | Full with WIP limits in v1.x |
| Scrum (sprints) | Full | Cycles (simplified) | Cycles | Sprints | Sprints with burndown in v1.x |
| Custom workflows | Extremely flexible (admin-heavy) | Opinionated (fixed) | Configurable | Configurable | Configurable per project (v1.x) |
| Custom fields | Extensive | Limited | Yes | Yes | v2 (fixed fields in v1) |
| JQL/query language | JQL (powerful) | Filter views | Filter views | Filter views | JQL-like in v2 |
| Hierarchy depth | Epic > Story > Sub-task (3 levels) | Project > Issue > Sub-issue | 3 levels | Flexible | Unlimited nesting |
| GitHub integration | Marketplace plugin | Native | Native | Native two-way sync | Native (commit-to-issue linking) |
| Time tracking | Via plugins | Built-in (basic) | No | Built-in | Built-in timer (v1) |
| AI features | Atlassian Intelligence | AI triage/summaries | AI agents | No | Not in scope |
| Speed/UX | Slow, complex | Fast, keyboard-first | Medium | Fast | Target: Linear-like speed |

### Messengers

| Feature | Slack | Mattermost | Zulip | Huly | Hub Approach |
|---------|-------|------------|-------|------|-------------|
| Threading | Threads (secondary) | Threads (secondary) | Topics (primary, mandatory) | Basic threads | Slack-style threads (optional per message) |
| Channels | Public + private | Public + private | Streams + topics | Channels | Public + private + auto-project channels |
| Search | Powerful | Full-text | Advanced (by topic) | Basic | Full-text with filters |
| Integrations | 2500+ apps | Smaller ecosystem | Webhooks | Built-in tracker | Built-in tracker/TMS/calls (core differentiator) |
| Bots | Extensive bot platform | Webhooks + plugins | Bots | Limited | Webhook receiver + Telegram bot |
| Voice/video | Huddles (basic) | Plugin (Jitsi) | Via BigBlueButton | Built-in video | Built-in voice + screen share |
| Self-hosted | No | Yes | Yes | Yes | Yes |
| Link previews | Rich (OG + app unfurling) | Basic OG | Basic | Issue cards | Issue/test case/PR cards (cross-module) |

### TMS

| Feature | TestRail | Allure TestOps | Hub Approach |
|---------|----------|----------------|-------------|
| Manual test cases | Full (steps + expected) | Full | Full (steps + BDD) |
| Automation integration | API-based | Native (100+ frameworks) | Allure results parser (free alternative) |
| Folder hierarchy | Sections + suites | Tree structure | Nested folders |
| Test runs | Manual + automated | Manual + automated + smart | Manual + automated (Allure) + mixed |
| Reporting | Basic charts | Advanced (trends, flaky, categories) | Allure-level reports (trends, flaky, categories) |
| Versioning | Basic | Smart (auto-update) | Explicit versioning with diff + rollback |
| Issue linking | Jira integration | Jira integration | Native (same platform) -- key advantage |
| Pricing | $36/user/mo (cloud) | ~$30K/yr (enterprise) | Free (self-hosted) |
| Environments | No native support | Configuration profiles | Labeled environments with comparison |

### Voice/Calls

| Feature | Google Meet | Zoom | LiveKit (raw) | Huly | Hub Approach |
|---------|-------------|------|---------------|------|-------------|
| Video | Yes | Yes | Yes | Yes | No (deliberate: voice only) |
| Screen share | Yes | Yes | Yes | Yes | Yes (multiple simultaneous) |
| Recording | Google Drive | Cloud/local | Egress service | Yes | MinIO storage, project-bound |
| Max participants | 100-500 | 100-1000 | Scalable | Unknown | 50 (target team size) |
| Scheduling | Google Calendar | Built-in | None (raw SDK) | Basic | Built-in with recurrence (v1.x) |
| Integration | Google Workspace | Standalone | SDK/API | Built-in tracker | Built-in tracker + messenger |
| Self-hosted | No | No | Yes | Yes | Yes |
| Raise hand | Yes | Yes | SDK support | Unknown | Yes |

---

## Sources

- [Huly - Everything App for your teams](https://huly.io/)
- [Huly GitHub](https://github.com/hcengineering/platform)
- [Plane.so - AI-native project management](https://plane.so/)
- [Plane GitHub](https://github.com/makeplane/plane)
- [Jira Software Features](https://www.atlassian.com/software/jira/features)
- [Linear vs Jira comparison (efficient.app)](https://efficient.app/compare/linear-vs-jira)
- [Linear vs Jira (monday.com)](https://monday.com/blog/rnd/linear-or-jira/)
- [Mattermost vs Slack (pumble.com)](https://pumble.com/blog/mattermost-vs-slack/)
- [Zulip - Why Zulip](https://zulip.com/why-zulip/)
- [TestRail vs Allure TestOps (Capterra)](https://www.capterra.com/compare/128204-227555/TestRail-vs-Allure-TestOps)
- [Allure TestOps - TestRail alternatives](https://qameta.io/blog/testrail-alternatives/)
- [LiveKit Documentation](https://docs.livekit.io/transport/)
- [LiveKit GitHub](https://github.com/livekit/livekit)
- [LiveKit screen sharing docs](https://docs.livekit.io/transport/media/screenshare/)

---
*Feature research for: Team Collaboration Hub*
*Researched: 2026-03-20*
