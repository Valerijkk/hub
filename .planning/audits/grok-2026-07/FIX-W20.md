# FIX-W20 — Program finale (RT chains + scoreboard)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Режим:** READ-ONLY product code (no mutation this wave)  
**Источники:** `19-red-team.md`, `FIX-W0`…`FIX-W19`, `00-SUMMARY.md`, live code re-read, `.planning/{STATE,COVERAGE}.md`, `planning-audit-2026-07/cov-2026-07-27-*`

Companion: [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md)

---

## 1. RT chains (A19)

Re-verified against current `backend/internal/**` on `cursor-grok-4.5` (HEAD at report time ≈ `e71f007` + prior fix commits). Status relative to the **named attack chain** in `19-red-team.md`.

| RT | Severity | Status | Evidence (code) | Closing commit(s) | Residual gap |
|----|----------|--------|-----------------|-------------------|--------------|
| **RT-001** Pending→JWT | CRITICAL | **CLOSED** | `auth/service_login.go` — `u.Status != user.StatusActive` → deny before `CreateSession`; `auth/middleware.go` `userIsDeactivated` = any non-`active`; refresh path same | `f8e17f9` (A1-001) | Dev-only `dev/handler` CreateSession bypass (out of attacker model / `-tags release`) |
| **RT-002** Messenger search all channels | CRITICAL | **CLOSED** | `messenger/service_search.go` — `channel_id` → `IsMember`; else `ListUserChannels` allowlist before search | `449c2c3` (A2-001); W17 matrix PASS | — |
| **RT-003** CRM PII mutate + invoice | CRITICAL | **PARTIAL** (**ADR-blocked** residual) | **Writes CLOSED:** `authorizeContactWrite` / `authorizeDealWrite` on Update/Delete (`crm/service.go`) | W2 CRM authorize* | **Reads still org-wide (ADR-blocked):** `ListContacts` / `GetPipeline` dump all contacts/deals to any JWT (`authMw` only; no `PermCrm*`). See §1.1 |
| **RT-004** Cross-project role IDOR | CRITICAL | **CLOSED** | `rbac/custom_roles.go` — `GetRoleProjectID` + `ErrRoleProjectMismatch` on Update/Delete/Assign | `48bedbb` (A1-002/003) | — |
| **RT-005** CompleteSprint steals issues | CRITICAL | **PARTIAL** | **Named chain CLOSED:** `CompleteSprint` → `GetSprintInProject(projectID, sprintID)` (`sprint/service.go:111`) | `1ac8669` (A3-001) | **Sibling IDOR ALIVE:** `GetSprint` / `UpdateSprint` / `StartSprint` / `UpdateRetroNotes` / `GetBurndown` still resolve by `sprintId` only (handler ignores path `projectId` except Complete/Start conflict check). `StartSprint` can activate a foreign sprint UUID under project A RBAC |
| **RT-006** Imex export without RBAC | HIGH | **CLOSED** | `imex/handler.go` — import/export/PDF wrapped in `rbac.RequirePermission` | `2686bb4` (A5-002); PDF project scope `0d5d862` (A5-010) | — |
| **RT-007** Tracker issueKey IDOR | HIGH | **CLOSED** | `tracker/repo_issues.go` `FindIssueByKey(ctx, projectID, key)` + `HasProjectWith` | `58a5f56` (A3-002) | — |
| **RT-008** Calls Create/chat/recording | HIGH | **CLOSED** | `CreateCall`/`JoinCall` → `IsChannelMember`; chat `authorizeCallChatAccess`; recording `authorizeCallRecordingAccess`; `GetActiveCall` member (`adbc03e`) | `c4b1dc2` `591fa0d` `ef72607` `cb80edd` `2b78dd3` `adbc03e` | — |
| **RT-009** Push hijack / unsubscribe DoS | HIGH | **CLOSED** | `notification/push.go` — Update/Delete `Endpoint` **and** `UserID`; `ErrSubscriptionForbidden` | `a90fb5a` (A2-005) | Needs endpoint knowledge (XSS/shared device) — mitigated for IDOR rebind |
| **RT-010** Calendar HR leave leak | HIGH | **CLOSED** | `calendar/service.go` + repo — leave scoped to self + shared-project members; no email in title | `0823cc2` (A6-005) | — |
| **RT-011** Dashboard project stats IDOR | HIGH | **CLOSED** | `dashboard/service.go` `GetProjectStats` → `IsProjectMember` → `ErrForbidden` | `c5edd70` (A7-001) | — |
| **RT-012** GitHub CreateBranch PAT | HIGH | **CLOSED** | Route `PermTaskEdit`; `FindIssueByKeyInProject` + `GetIntegrationInProject` | `2ffd60f` (A5-006); GitLab twin `72f55e7` | — |

### RT scoreboard

| Metric | Value |
|--------|------:|
| Chains | 12 |
| **CLOSED** | **10** |
| **PARTIAL** | **2** (RT-003 **ADR-blocked** read; RT-005 siblings) |
| **ALIVE** (full named chain) | **0** |
| Original exploitability | 8.5/10 |
| Post-FIX residual (honest) | **~3.5/10** insider (sibling sprint IDOR + org CRM read) |

**No code fix in W20:** RT-005 sibling sweep needs handler+service signature changes + hubtest mutation across Get/Update/Start/Retro/Burndown — exceeds the &lt;30min ALIVE bar while CompleteSprint (issue-theft) is already closed. Track as follow-up P1 `A3-001-siblings`.

### 1.1 RT-003 residual honesty (ADR-blocked — do not claim code-closable)

**What is still open (reads only):**

| Endpoint | Gate today | Scope |
|----------|------------|-------|
| `GET /api/crm/contacts` → `ListContacts` | `authMw` only | All non-deleted contacts (email/phone/notes) — org dump |
| `GET /api/crm/pipeline` → `GetPipeline` → `ListDeals` | `authMw` only | All non-deleted deals by stage — org dump |

Writes (Update/Delete + won→invoice) remain **CLOSED** via owner/admin `authorize*Write`. No project/member filter on list paths. Contacts have **no** `project_id`; deal `project_id` is optional (membership checked only on create/update refs).

**Why a “minimal safe scope” is not shippable without ADR:**

1. **No existing RBAC hook** — `backend/internal/rbac` has **zero** `PermCrm*` / CRM module permissions; routes are not under `/projects/:projectId`, so project RBAC cannot be “already required”.
2. **Owner-only List/Pipeline** would invent personal CRM semantics: wiki (`docs/wiki/API.md` CRM) documents org list without owner-scope (contrast: Bases explicitly owner-scoped). FE pipeline expects shared kanban. That is a **product** choice, not a bugfix.
3. **Admin-only List** (accounting pattern) would lock CRM-lite to founder/admin and break member sales workflow — also product ADR.
4. **Project-membership filter** is incomplete: contacts are not project-bound; deals without `project_id` have no project to scope — inventing rules here risks silent data loss or fake “safety”.
5. **Gray zone already open:** `GZ-SEC-13` (modular `crm:manager` + field masking) — default in `GRAY-ZONES-DECISIONS-2026-07-03.md`, **not** implemented as code yet. Closing read dump = accept **org-shared CRM** in ADR **or** implement that role model via ADR — not a silent be1 patch.
6. Pagination alone (A6-009) reduces unbounded blast but **does not** close PII visibility to every JWT member.

**Verdict:** RT-003 read residual = **ADR-blocked**. Code-only `fix(crm):` for list scope without product ADR is **out of bounds** (would invent multi-role / personal-CRM policy). Track until ADR: org-shared accept **or** `crm:manager` / owner-read scope (+ optional pagination).

---

## 2. Planning sync (drift check — no rewrite)

| Artifact | Claim | Reality (2026-07-27) | Drift? |
|----------|-------|----------------------|--------|
| `STATE.md` frontmatter | `coverage.overall_internal_statements: 81.9`, measured 2026-07-27 | Matches `planning-audit-2026-07/cov-2026-07-27-total.txt` / `COVERAGE.md` | OK |
| `STATE.md` progress | `tasks_completed: 125` / `phases_percent: 100` vs `goals_verified: 55` | Notes explain checkbox ≠ VERIFIED_PASS — intentional | OK (documented) |
| `STATE.md` body table (~L270) | Still quotes **83.0%** Docker baseline in older milestone table | Frontmatter + COVERAGE gate = **81.9%** | **Mild drift** — historical table not refreshed; prefer frontmatter / COVERAGE §gate |
| `COVERAGE.md` | Gate **81.9%** (−1.1 pp vs 83.0); floor 82% **miss** | Aligned with planning-audit report | OK |
| `ROADMAP.md` Progress | C2 (2026-07-27) extended Progress to Phase 125 | Checklist 65–125 present `[x]` | OK vs A15-001 original gap |
| `AGENTS.md` | Still highlights **83.0%** baseline in places | Concurrent with COVERAGE 81.9 gate | **Mild drift** — dual numbers; gate truth = 81.9 |
| Grok FIX session notes | STATE has W12–W19 notes | W20 note should append (this report) | Expected |

**A15-001 (planning P0):** largely mitigated by planning-audit C2 + STATE notes; residual is narrative dual-coverage (83.0 vs 81.9), not missing phases.

---

## 3. Coverage

| Source | Overall `internal/` statements | Notes |
|--------|-------------------------------:|-------|
| **Canonical remeasure** | **81.9%** | `planning-audit-2026-07/cov-2026-07-27-REPORT.md` / `cov-2026-07-27-total.txt` (Docker, `-tags hubtest`, 2026-07-27) |
| Previous baseline | **83.0%** | 2026-06-22 Docker baseline |
| Delta | **−1.1 pp** | Below 82% COV-01 floor |
| W20 this session | **NOT_REMEASURED** | Cite planning-audit 81.9%; do not invent package % |

Security 6-pack: last full statement pack still **2026-06-22** (≥80% each); 2026-07-27 run did not refresh comparable per-package statement table (func-mean proxy only).

---

## 4. Hygiene — remaining DEFERRED / BLOCKED / MANUAL

### Schema / approval

| ID | Status | Notes |
|----|--------|-------|
| **A8-006** | **BLOCKED** | Comment hard-delete; TZ soft-delete needs `deleted_at` on `ent/schema/comment.go` (FIX-W10) |
| **A8-001** indexes | **NEEDS_APPROVAL** | Query rewrite FIXED (`0ce8cd7`); composite index `(project_issues, deleted_at)` pending (FIX-W15) |
| **A8-002** indexes | **NEEDS_APPROVAL** | Query rewrite FIXED (`443a569`); TC FK index pending (FIX-W15) |
| A8-005 soft-delete Channel | PARTIAL | Tx cascade FIXED (`b77ba23`); `deleted_at` not added |

### MANUAL ops (repo FIXED*, ops pending)

| ID | Repo | MANUAL |
|----|------|--------|
| **A14-002** | `7c03388` SeaweedFS secrets out of defaults | Rotate S3 if `hub`/`hubsecret` ever deployed |
| **A14-003** | `53f0436` NATS auth + redact | Set `NATS_USER`/`NATS_PASSWORD` on existing prod + recreate |
| **A14-009** | `68cbd14` OTEL wired | Set OTEL endpoint to enable spans |
| **PG-03 / FA-06** | — | Resend/Telegram secret rotation — sole prod cutover blocker (unchanged) |

### FE / product DEFERRED (from FIX-W11/W13/W17–19)

| IDs | Notes |
|-----|-------|
| A10-007…009, A10-011, A10-012, A10-014 | InfiniteQuery notifications, Kanban/Gantt virtual, timer selectors, dead KanbanBoard, QueryStateView rollout |
| A10-013 | PARTIAL a11y labels |
| A11-014 | TZ error-shape wording vs live AppError |
| A15-004 | REQ LOGN-05 / wiki drift |
| BE `LoginWithTelegram*` / notify_mode write API | Dead FE removed; BE DTO/API deferred |
| A4-005/008/009…012, A2-011, A3-013/014/018–020 | Block F time-box residuals (W18) |
| A6-007…009, A5-012…015, A7-010…012 | W19 out of time-box |
| A14-008 | access-log without InfoContext (P2 obs) |

### Dead code / deps

No new obvious dead-deps sweep in W20. Prior: FE telegram client / notify_mode client / TelegramButton removed (W11/W13). Residual unused FE wrappers (teams/trash/RBAC/imex HTTP) still **DEFERRED** product work (FIX-W11).

---

## 5. 261 findings scoreboard (approx)

Canonical unique count from `00-SUMMARY.md`: **262** (user shorthand «261» ≈ same set; 12 P0 + 77 P1 + 120 P2 + 53 P3).

| Bucket | Approx | Notes |
|--------|-------:|-------|
| Unique findings (deduped) | **262** | `00-SUMMARY.md` |
| Closed in FIX-W0…W19 (IDs with FIXED/CLOSED in reports) | **~110–140** | Regex over FIX reports under-counts (~81) where waves used ✅/SHA tables without keyword; commit volume W1–W19 supports **≥120** discrete closes |
| Still open / deferred / blocked / P2–P3 untouched | **~120–150** | Mostly P2/P3 polish, schema, FE refactors, docs |
| **Remaining P0 (security/product)** | **0** full ALIVE | All 12 original P0 primary paths closed or downgraded; residuals are PARTIAL siblings / planning narrative |
| **Remaining P0-class residual** | **2 PARTIAL RT** | RT-003 org CRM read (**ADR-blocked**, §1.1); RT-005 sprint sibling IDOR |
| **Remaining clear P1** | **~15–25** | Mix of deferred FE (A10-*), integrations residuals (A5-012+), Block F leftovers, A3-001-siblings |

### Original P0 map → status

| ID | Status |
|----|--------|
| A1-001 / RT-001 | CLOSED `f8e17f9` |
| A1-002/003 / RT-004 | CLOSED `48bedbb` |
| A1-004 GDPR founder | CLOSED W2 |
| A2-001 / RT-002 | CLOSED `449c2c3` |
| A3-001 / RT-005 | PARTIAL (CompleteSprint CLOSED; siblings open) |
| A5-001 CSV mappings | CLOSED `ce3bbd2` |
| A6-001 / RT-003 | PARTIAL **ADR-blocked** (writes CLOSED; list org-wide — §1.1) |
| A6-002 invoice pay tx | CLOSED W2 |
| A11-001 join path | CLOSED `da3fe4b` |
| A11-002 TMS FormData | CLOSED W2/W3 |
| A15-001 planning | MOSTLY CLOSED (C2 + STATE notes); mild 83.0/81.9 narrative drift |

### Honest uncovered list (do not claim fixed)

1. **Sprint sibling IDOR** — Get/Update/Start/Retro/Burndown without `GetSprintInProject` (RT-005 residual).
2. **CRM org-wide List/Pipeline** — **ADR-blocked** residual (RT-003 / A6-009): any JWT lists all contacts/deals; no `PermCrm*`; owner/admin/project filters would invent product policy (§1.1).
3. **A8-006** comment soft-delete schema — BLOCKED on approval.
4. **A8-001/002** Ent composite indexes — NEEDS_APPROVAL.
5. **MANUAL** A14-002/003/009 + **PG-03/FA-06** secret rotation.
6. **FE deferred** A10-007…014 class (virtualization, InfiniteQuery, dead KanbanBoard).
7. **Block F leftovers** A4-005+, A3-013+, A5-012+, A7-010+, A14-008.
8. **COV-01 floor** — 81.9% &lt; 82%; security 6-pack not remeasured 2026-07-27.
9. **GetProject non-member metadata** — product-grey GAP (W17).
10. **TZ AppError wording** A11-014 / docs-only.

---

## 6. Verdict

Grok FIX program W0–W19 closed the **critical insider Broken Access Control** surface from A19: **0 fully ALIVE RT chains**. Two PARTIAL residuals remain (CRM read dump, sprint non-complete IDOR). Planning/coverage truth is **81.9%** (2026-07-27), not the stale 83.0 narrative alone. Schema indexes / comment soft-delete / prod secret ops remain outside code-only waves.

**Recommended next (human / next wave):**

1. `fix(sprint):` scope Get/Update/Start/Retro/Burndown to `GetSprintInProject` + mutation tests (**closes RT-005**).
2. **Product ADR (required):** CRM list scope vs org-shared CRM / `crm:manager` (GZ-SEC-13) — **blocks** code close of RT-003 read; do not ship silent owner-scope List.
3. Approve A8-001/002 indexes + A8-006 `deleted_at`.
4. Ops: FA-06 + A14 MANUAL checklist.
5. Optional coverage uplift wave to restore ≥82%.

*Merge to `master` — only by user.*
