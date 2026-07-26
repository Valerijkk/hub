# NFR-24 — Classification of Ent schemas for optimistic locking

**Date measured:** 2026-07-27  
**Scope:** `backend/ent/schema/*.go` (read-only)  
**NFR wording** (`docs/system_specification.md`):

> NFR-24 | Optimistic locking | `version` int на каждой изменяемой Ent

**Related TZ:** `docs/tz_constraints.md` / `.cursor/rules/60-tz-constraints.mdc` — optimistic locking via `version` on mutable Ent schemas.

**Decision status:** owner options only — **no recommendation chosen** in this document.

---

## 1. Measurement (NOW)

| Metric | Value | Notes |
|---|---:|---|
| `.go` files in `backend/ent/schema/` | **83** | includes `helpers.go` (not an entity) |
| Entity schemas (exclude `helpers.go`) | **82** | honest universe for NFR-24 |
| Schemas with `field.Int("version")` | **14** | see §3 |
| Schemas using `VersionMixin` / any Mixin | **0** | `helpers.go` has no mixin; no `Mixin()` on schemas |
| Prior audit claim (`00-PLANNING-AUDIT.md` §4.1) | 17 / 83 | **stale** — remeasure shows **14** with `version`; 83 counted helpers |

### How measured

```powershell
# 2026-07-27, PowerShell, backend/ent/schema
(Get-ChildItem *.go | Where-Object Name -ne 'helpers.go').Count
Get-ChildItem *.go | Where-Object Name -ne 'helpers.go' |
  Where-Object { Select-String -Path $_.FullName -Pattern 'field\.Int\("version"\)' -Quiet }
Select-String -Path *.go -Pattern 'Mixin|VersionMixin'
```

`TestCaseVersion.version_number` is a **snapshot sequence**, not optimistic-lock `version` — not counted.

---

## 2. Classification rubric

Every entity schema is assigned exactly one class:

| Class | Meaning |
|---|---|
| **mutable_needs_version** | Created and later **updated in place** by users/admins; concurrent edit is plausible → NFR-24 applies |
| **append_only** | Insert-once (logs, deliveries, snapshots, telemetry). In-place merge of content fields is not a product path |
| **immutable_link** | Join / association / pin / vote: create + delete (or create-only); no meaningful content PATCH |
| **unclear** | Lifecycle / counter / status updates exist, but optimistic lock value is debatable — **owner judgment** |

**Honest denominator for NFR-24 compliance** = count of `mutable_needs_version` only.  
Append-only and immutable links are out of scope for optimistic locking. Unclear rows are **excluded from the denominator** until the owner reclassifies them (listed so they are not silently “green”).

---

## 3. Schemas that already have `version`

| Schema | Ent type | `version` |
|---|---|---|
| crmcontact.go | CrmContact | yes |
| crmdeal.go | CrmDeal | yes |
| hubbase.go | HubBase | yes |
| hubcolumn.go | HubColumn | yes |
| hubrecord.go | HubRecord | yes |
| hubtable.go | HubTable | yes |
| invoice.go | Invoice | yes |
| issue.go | Issue | yes |
| journalentry.go | JournalEntry | yes |
| leaverequest.go | LeaveRequest | yes |
| ledgeraccount.go | LedgerAccount | yes |
| marketplacelisting.go | MarketplaceListing | yes |
| newspost.go | NewsPost | yes |
| testcase.go | TestCase | yes |

**14 / 82** entity schemas. **None** via VersionMixin (mixin does not exist in tree today).

---

## 4. Full classification table (82 schemas)

| # | Schema file | Ent type | Has `version` | Class | Rationale (short) |
|---:|---|---|:---:|---|---|
| 1 | activitylog.go | ActivityLog | no | append_only | Activity stream; append |
| 2 | aiprovider.go | AIProvider | no | mutable_needs_version | Admin edits provider config (`updated_at`) |
| 3 | aiusagedaily.go | AIUsageDaily | no | append_only | Daily usage rollup / telemetry |
| 4 | aiuserpreference.go | AIUserPreference | no | mutable_needs_version | Per-user prefs updated in place |
| 5 | apikey.go | APIKey | no | unclear | Create / revoke; not concurrent content edit |
| 6 | auditlog.go | AuditLog | no | append_only | Immutable audit trail |
| 7 | automationrule.go | AutomationRule | no | mutable_needs_version | Rule definition edited over time |
| 8 | backuprecord.go | BackupRecord | no | append_only | Backup job record |
| 9 | board.go | Board | no | mutable_needs_version | Name / filters / swimlane config |
| 10 | bookmark.go | Bookmark | no | immutable_link | User↔message bookmark |
| 11 | call.go | Call | no | unclear | Mostly status machine active→ended |
| 12 | callparticipant.go | CallParticipant | no | immutable_link | Participant membership row |
| 13 | channel.go | Channel | no | mutable_needs_version | Channel metadata edited (`updated_at`) |
| 14 | channelinvite.go | ChannelInvite | no | unclear | Invite lifecycle / use counters |
| 15 | channelmember.go | ChannelMember | no | immutable_link | Membership join |
| 16 | coffeeenrollment.go | CoffeeEnrollment | no | unclear | Opt-in toggle, low concurrency |
| 17 | coffeepair.go | CoffeePair | no | append_only | Pairing result snapshot |
| 18 | coffeeround.go | CoffeeRound | no | unclear | Round lifecycle |
| 19 | comment.go | Comment | no | mutable_needs_version | Comments edited after create |
| 20 | component.go | Component | no | mutable_needs_version | Tracker component CRUD |
| 21 | crmcontact.go | CrmContact | **yes** | mutable_needs_version | CRM contact edits |
| 22 | crmdeal.go | CrmDeal | **yes** | mutable_needs_version | CRM deal edits |
| 23 | customfielddefinition.go | CustomFieldDefinition | no | mutable_needs_version | Field defs change over time |
| 24 | dashboardlayout.go | DashboardLayout | no | mutable_needs_version | Layout JSON updated |
| 25 | environment.go | Environment | no | mutable_needs_version | TMS environment config |
| 26 | filemetadata.go | FileMetadata | no | unclear | Status pending→confirmed→deleted; rare dual-edit |
| 27 | githubintegration.go | GitHubIntegration | no | mutable_needs_version | Integration settings |
| 28 | githublink.go | GitHubLink | no | immutable_link | Issue↔PR/commit link |
| 29 | gitlabintegration.go | GitLabIntegration | no | mutable_needs_version | Integration settings |
| 30 | gitlablink.go | GitLabLink | no | immutable_link | Issue↔MR/commit link |
| 31 | hubbase.go | HubBase | **yes** | mutable_needs_version | Bases CRUD |
| 32 | hubcolumn.go | HubColumn | **yes** | mutable_needs_version | Column schema edits |
| 33 | hubrecord.go | HubRecord | **yes** | mutable_needs_version | Record cell edits |
| 34 | hubtable.go | HubTable | **yes** | mutable_needs_version | Table schema edits |
| 35 | invite.go | Invite | no | unclear | `use_count` increments; not OL target |
| 36 | invoice.go | Invoice | **yes** | mutable_needs_version | Accounting invoice |
| 37 | issue.go | Issue | **yes** | mutable_needs_version | Core concurrent-edit surface |
| 38 | issuelink.go | IssueLink | no | immutable_link | Issue↔issue relation |
| 39 | issuemessagelink.go | IssueMessageLink | no | immutable_link | Issue↔message link |
| 40 | journalentry.go | JournalEntry | **yes** | mutable_needs_version | Journal header (has version today) |
| 41 | journalline.go | JournalLine | no | append_only | Lines created with entry; no `updated_at` |
| 42 | kudo.go | Kudo | no | append_only | One-shot recognition |
| 43 | label.go | Label | no | mutable_needs_version | Label rename/color |
| 44 | labelgroup.go | LabelGroup | no | mutable_needs_version | Group rename |
| 45 | leaverequest.go | LeaveRequest | **yes** | mutable_needs_version | HR leave request |
| 46 | ledgeraccount.go | LedgerAccount | **yes** | mutable_needs_version | Chart of accounts |
| 47 | marketplacelisting.go | MarketplaceListing | **yes** | mutable_needs_version | Listing edits |
| 48 | newspost.go | NewsPost | **yes** | mutable_needs_version | Intranet post edits |
| 49 | notification.go | Notification | no | unclear | Mark-read / dismiss updates |
| 50 | notificationpreference.go | NotificationPreference | no | mutable_needs_version | Preference PATCH |
| 51 | organization.go | Organization | no | mutable_needs_version | Org profile/settings |
| 52 | outboxevent.go | OutboxEvent | no | append_only | Outbox / pub pipeline |
| 53 | pinnedmessage.go | PinnedMessage | no | immutable_link | Pin association |
| 54 | poll.go | Poll | no | mutable_needs_version | Close / status; tied to NewsPost |
| 55 | polloption.go | PollOption | no | append_only | Options fixed at create (spec: no edit after votes) |
| 56 | pollvote.go | PollVote | no | immutable_link | One vote row |
| 57 | project.go | Project | no | mutable_needs_version | **Notable gap** — highly mutable, no `version` |
| 58 | projectmember.go | ProjectMember | no | immutable_link | Membership join; role PATCH rare (owner may promote) |
| 59 | projecttemplate.go | ProjectTemplate | no | mutable_needs_version | Template definition |
| 60 | pushsubscription.go | PushSubscription | no | unclear | Endpoint register/unregister |
| 61 | reaction.go | Reaction | no | immutable_link | Emoji reaction |
| 62 | recording.go | Recording | no | unclear | Call recording metadata |
| 63 | role.go | Role | no | mutable_needs_version | RBAC role definition |
| 64 | savedfilter.go | SavedFilter | no | mutable_needs_version | Named filter edits |
| 65 | scheduledcall.go | ScheduledCall | no | mutable_needs_version | Calendar series (FR-CAL) |
| 66 | session.go | Session | no | unclear | `last_active_at` touch; not content OL |
| 67 | sprint.go | Sprint | no | mutable_needs_version | Sprint dates/status |
| 68 | systemsetting.go | SystemSetting | no | mutable_needs_version | Key/value admin settings |
| 69 | team.go | Team | no | mutable_needs_version | Team name/metadata |
| 70 | teammember.go | TeamMember | no | immutable_link | Membership join |
| 71 | testcase.go | TestCase | **yes** | mutable_needs_version | TMS case edits |
| 72 | testcaseversion.go | TestCaseVersion | no | append_only | Immutable snapshot (`version_number`) |
| 73 | testresult.go | TestResult | no | append_only | Execution result row |
| 74 | testrun.go | TestRun | no | mutable_needs_version | Run status / metadata during execution |
| 75 | testsuite.go | TestSuite | no | mutable_needs_version | Suite structure edits |
| 76 | timelog.go | TimeLog | no | unclear | Work logs sometimes editable |
| 77 | user.go | User | no | mutable_needs_version | **Notable gap** — profile edits, no `version` |
| 78 | userblock.go | UserBlock | no | immutable_link | Block relation |
| 79 | webhookdelivery.go | WebhookDelivery | no | append_only | Delivery attempt log |
| 80 | webhookendpoint.go | WebhookEndpoint | no | mutable_needs_version | Endpoint URL/secret config |
| 81 | workflowstatus.go | WorkflowStatus | no | mutable_needs_version | Status name/category |
| 82 | workflowtransition.go | WorkflowTransition | no | mutable_needs_version | Transition rules |

---

## 5. Aggregate counts (2026-07-27)

| Class | Count | Of which have `version` |
|---|---:|---:|
| **mutable_needs_version** | **45** | **14** |
| append_only | 12 | 0 |
| immutable_link | 13 | 0 |
| unclear | 12 | 0 |
| **Total entity schemas** | **82** | **14** |

### Honest NFR-24 denominator

| Metric | Value |
|---|---:|
| Denominator (`mutable_needs_version`) | **45** |
| Numerator (mutable ∧ has `version`) | **14** |
| Coverage of mutable schemas | **14 / 45 ≈ 31.1%** |
| Naive “14 / 82” (misleading) | 17.1% |
| Prior audit “17 / 83” | superseded by this measure |

### Mutable schemas missing `version` (31)

`aiprovider`, `aiuserpreference`, `automationrule`, `board`, `channel`, `comment`, `component`, `customfielddefinition`, `dashboardlayout`, `environment`, `githubintegration`, `gitlabintegration`, `label`, `labelgroup`, `notificationpreference`, `organization`, `poll`, `project`, `projecttemplate`, `role`, `savedfilter`, `scheduledcall`, `sprint`, `systemsetting`, `team`, `testrun`, `testsuite`, `user`, `webhookendpoint`, `workflowstatus`, `workflowtransition`.

Highest product-risk gaps (multi-user edit surfaces): **`project`**, **`user`**, **`channel`**, **`comment`**, **`sprint`**, **`scheduledcall`**, **`board`**, workflow/status entities.

### Unclear (owner must resolve) — 12

`apikey`, `call`, `channelinvite`, `coffeeenrollment`, `coffeeround`, `filemetadata`, `invite`, `notification`, `pushsubscription`, `recording`, `session`, `timelog`.

---

## 6. Owner options (DO NOT choose here)

### Option 1 — Roll `VersionMixin` to all `mutable_needs_version`

Introduce an Ent `VersionMixin` (does not exist today) and apply it to all **45** mutable schemas (or at least the 31 missing). Wire services/handlers to check `version` and return **409** on conflict where updates are exposed.

| Pros | Cons |
|---|---|
| Matches current NFR-24 / TZ wording literally | Large migration: Atlas schema changes across ~31 tables |
| Uniform pattern for new schemas (mixin copy-paste) | Touches many modules; risk of regressions / CI time |
| Closes known gaps (`project`, `user`, …) | Not every mutable entity has a real concurrent-edit UX — cost may exceed benefit |
| Makes coverage gate automatable (`mutable` ∩ `version`) | Requires service-layer enforcement, not just DB column |

### Option 2 — Narrow NFR-24 wording to the real list

Rewrite NFR-24 (and TZ mirror) to an **explicit allow-list** of entities that require optimistic locking (starting from the 14 that already have `version`, plus any owner-added high-risk ones). Document that append-only / links / unclear are exempt.

| Pros | Cons |
|---|---|
| Docs match code without a multi-module code wave | Weaker global invariant — reviewers must know the list |
| Focuses engineering on surfaces with real lost-update risk | Drift risk if list not kept next to schema changes |
| Honest status reporting (no fake “every Ent”) | Does not fix `project`/`user` unless owner adds them to the list |
| Cheaper near-term under July code-freeze | Spec tables that assume “all entities have version” (Intranet/Calendar/Wiki sections) still need line edits |

---

## 7. What this document does *not* do

- Does **not** pick Option 1 or 2.
- Does **not** change product code, migrations, or NFR text.
- Does **not** reclassify `unclear` without owner input.
- Spec-only entities without schemas yet (e.g. Wiki `wiki_page`) are out of this inventory; they remain FR/NFR planning debt (see `00-PLANNING-AUDIT.md` §5).

---

## 8. Traceability

| Source | Use |
|---|---|
| `docs/system_specification.md` NFR-24 | Requirement text |
| `docs/tz_constraints.md` / `60-tz-constraints.mdc` | Mandatory optimistic lock |
| `.planning/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.1 | Prior (stale) 17/83 claim |
| `.planning/planning-audit-2026-07/01-PLANNING-FIXES.md` B3 | Asks for this owner decision |
| `backend/ent/schema/*.go` | Measurement source, 2026-07-27 |
