# FIX-W20 — Program finale (RT chains + scoreboard)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27 · **Remediation recheck:** 2026-07-28  
**Режим:** W20 READ-ONLY product code; 2026-07-28 wave = WEAK evidence + targeted residual fixes (see PROGRAM summary)  
**Источники:** `19-red-team.md`, `FIX-W0`…`FIX-W19`, `00-SUMMARY.md`, live code re-read, `.planning/{STATE,COVERAGE}.md`, `planning-audit-2026-07/cov-2026-07-27-*`

Companion: [`PROGRAM-W4-W20-SUMMARY.md`](./PROGRAM-W4-W20-SUMMARY.md)

---

## 1. RT chains (A19)

Re-verified against current `backend/internal/**` on `cursor-grok-4.5` (HEAD at report time ≈ `e71f007` + prior fix commits). Status relative to the **named attack chain** in `19-red-team.md`.

| RT | Severity | Status | Evidence (code) | Closing commit(s) | Residual gap |
|----|----------|--------|-----------------|-------------------|--------------|
| **RT-001** Pending→JWT | CRITICAL | **CLOSED** | `auth/service_login.go` — `u.Status != user.StatusActive` → deny before `CreateSession`; `auth/middleware.go` `userIsDeactivated` = any non-`active`; refresh path same | `f8e17f9` (A1-001) | Dev-only `dev/handler` CreateSession bypass (out of attacker model / `-tags release`) |
| **RT-002** Messenger search all channels | CRITICAL | **CLOSED** | `messenger/service_search.go` — `channel_id` → `IsMember`; else `ListUserChannels` allowlist before search | `449c2c3` (A2-001); W17 matrix PASS | — |
| **RT-003** CRM PII mutate + invoice | CRITICAL | **CLOSED** | **Writes:** `authorizeContactWrite` / `authorizeDealWrite`. **Reads:** `ListContacts` / `GetPipeline` owner-scoped (`created_by` / `owner_id`); `IsSystemAdmin` → org-wide | W2 authorize* + RT-003 list scope | — |
| **RT-004** Cross-project role IDOR | CRITICAL | **CLOSED** | `rbac/custom_roles.go` — `GetRoleProjectID` + `ErrRoleProjectMismatch` on Update/Delete/Assign | `48bedbb` (A1-002/003) | — |
| **RT-005** CompleteSprint steals issues | CRITICAL | **CLOSED** | All service paths use `GetSprintInProject(projectID, sprintID)`: Complete/Start/Get/Update/Retro/Burndown/AddIssues | `1ac8669` (Complete); `8e0c9f2` (Start); `716c823` (Get/Update/Retro/Burndown/AddIssues + hubtests) | — |
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
| **CLOSED** | **12** |
| **PARTIAL** | **0** |
| **ALIVE** (full named chain) | **0** |
| Original exploitability | 8.5/10 |
| Post-FIX residual (honest) | **~1.5/10** (schema/ops residuals, not RT) |

**RT-005 closed 2026-07-28:** siblings scoped in `716c823` after StartSprint `8e0c9f2` / CompleteSprint `1ac8669`.

### 1.1 RT-003 — CLOSED (product decision 2026-07-28)

**RT-003 closed 2026-07-28:** product decision — non-admin List/Pipeline owner-scoped; admin/founder org-wide.

| Endpoint | Gate | Scope |
|----------|------|-------|
| `GET /api/crm/contacts` → `ListContacts(userID)` | `authMw` + service scope | Member: `created_by`; admin/founder: org-wide |
| `GET /api/crm/pipeline` → `GetPipeline(userID)` → `ListDeals` | same | Member: `owner_id`; admin/founder: org-wide |

Writes remain owner/admin via `authorize*Write`. Product policy (user-approved): personal CRM for members + CRM-lite admin view for system admin/founder — closes former ADR-blocked org dump without inventing `PermCrm*` / project filters.

**Verdict:** RT-003 **CLOSED**.

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
| **Remaining P0-class residual** | **0** | RT-003 list scope CLOSED (owner / admin org-wide) |
| **Remaining clear P1** | **~15–25** | Mix of deferred FE (A10-*), integrations residuals (A5-012+), Block F leftovers, A3-001-siblings |

### Original P0 map → status

| ID | Status |
|----|--------|
| A1-001 / RT-001 | CLOSED `f8e17f9` |
| A1-002/003 / RT-004 | CLOSED `48bedbb` |
| A1-004 GDPR founder | CLOSED W2 |
| A2-001 / RT-002 | CLOSED `449c2c3` |
| A3-001 / RT-005 | **CLOSED** `1ac8669` + `8e0c9f2` + `716c823` |
| A5-001 CSV mappings | CLOSED `ce3bbd2` |
| A6-001 / RT-003 | **CLOSED** (writes + owner/admin list scope) |
| A6-002 invoice pay tx | CLOSED W2 |
| A11-001 join path | CLOSED `da3fe4b` |
| A11-002 TMS FormData | CLOSED W2/W3 |
| A15-001 planning | MOSTLY CLOSED (C2 + STATE notes); mild 83.0/81.9 narrative drift |

### Honest uncovered list (do not claim fixed)

#### Closed in remediation recheck 2026-07-28 (mark off prior residuals)

| Prior residual | Status after 2026-07-28 | SHA |
|----------------|-------------------------|-----|
| A3-005 StartTimer open-timer race | **CLOSED** (tx + UPDATE row-lock on issue) | `bc06be9` |
| RT-003 CRM List/Pipeline org dump | **CLOSED** (owner scope; admin org-wide) | `5e636f2` + `573c6fc` |
| RT-005 StartSprint foreign UUID activate | **CLOSED** (sibling) | `8e0c9f2` |
| RT-005 Get/Update/Retro/Burndown/AddIssues | **CLOSED** | `716c823` |
| A2-007 Authorizer integration WEAK | **CLOSED** | `cce9265` |
| A2-010 Dispatch Nak WEAK | **CLOSED** | `612cf22` |
| A3-011 Postgres twin WEAK | **CLOSED** | `6342059` |
| A1-009 / A1-013 HTTP WEAK | **CLOSED** | `957472e` / `9764850` |
| A14-006 / A14-007 slog/DevMode WEAK | **CLOSED** | `0cbc7d2` / `4ebd8c3` |
| A4-002 / A4-006 runtime/HTTP WEAK | **CLOSED** | `c3bc077` / `1b19d13` |
| A10-001 WS→Query handler WEAK | **CLOSED** | `575106d` |
| A13-014 / A13-018 threshold Vitest WEAK | **CLOSED** | `3536e0e` / `53d72c8` |
| A14-005 metrics CIDR default (repo) | **FIXED*** (ops residual) | `c5613f6` |

#### Still open / blocked (accurate)

1. **A8-006** comment soft-delete schema — **BLOCKED** on approval.
2. **A8-001/002** Ent composite indexes — **NEEDS_APPROVAL** (query rewrites FIXED).
3. **MANUAL** A14-002/003/009 + **PG-03/FA-06** secret rotation.
4. **FE deferred** A10-007…014 class residuals (InfiniteQuery notifications, Gantt virtual, dead KanbanBoard, QueryStateView) — A10-001 / A13-014/018 product paths CLOSED with WEAK tests.
5. **Block F leftovers** A4-005+, A3-013+, A5-012+, A7-010+, A14-008.
6. **COV-01 floor** — 81.9% &lt; 82% (2026-07-27); security 6-pack not remeasured 2026-07-27/28 — **no invented remeasure**.
7. **GetProject non-member metadata** — product-grey GAP (W17).
8. **TZ AppError wording** A11-014 / docs-only.
9. **A14-005 ops residual** — compose default tightened; blanket RFC1918 in `.env` can re-open Docker peers.

---

## 6. Verdict

Grok FIX program W0–W19 closed the **critical insider Broken Access Control** surface from A19: **0 fully ALIVE RT chains**, **0 PARTIAL RT**. **RT-005** and **RT-003** CLOSED 2026-07-28 (owner/admin CRM list scope + sprint siblings). Planning/coverage truth is **81.9%** (2026-07-27), not the stale 83.0 narrative alone. Schema indexes / comment soft-delete / prod secret ops remain outside code-only waves.

**Recommended next (human / next wave):**

1. Approve A8-001/002 indexes + A8-006 `deleted_at`.
2. Ops: FA-06 + A14 MANUAL checklist.
3. Optional coverage uplift wave to restore ≥82%.

*Merge to `master` — only by user.*
