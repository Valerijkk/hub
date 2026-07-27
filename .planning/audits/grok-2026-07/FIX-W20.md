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

## 4. Hygiene — remaining DEFERRED / MANUAL (schema cleared)

### Schema / approval — CLEARED (residual-close 2026-07-28)

| ID | Status | Notes |
|----|--------|-------|
| **A8-006** | **CLOSED** | Comment `deleted_at` + soft-delete (`5c74170` / `9f4ce5b` / `c40c621`) |
| **A8-001** indexes | **CLOSED** | Query rewrite + composite index (`0ce8cd7` / `6031985`) |
| **A8-002** indexes | **CLOSED** | Query rewrite + composite index (`443a569` / `6031985`) |
| **A8-005** | **CLOSED** | Channel soft-delete (`25028b5` / `293dee7` / `14c7929`) |

### MANUAL ops (repo FIXED*, ops pending)

| ID | Repo | MANUAL |
|----|------|--------|
| **A14-002** | `7c03388` + fail-closed `f626397` | Rotate S3 if `hub`/`hubsecret` ever deployed |
| **A14-003** | `53f0436` + fail-closed `f626397` | Set `NATS_USER`/`NATS_PASSWORD` on existing prod + recreate |
| **A14-009** | `68cbd14` OTEL wired | Set `OTEL_EXPORTER_OTLP_ENDPOINT` to enable spans (**true residual**) |
| **PG-03 / FA-06** | — | Resend/Telegram secret rotation — sole prod cutover blocker (unchanged) |

### FE / product residuals (post residual-close)

| IDs | Notes |
|-----|-------|
| A11-009…013 | **CLOSED** — SPA wire (`68f9b0b`…`cbfae6c`) |
| A15-004 | **CLOSED** — REQUIREMENTS Magic Link only (`d8e787b`) |
| BE `LoginWithTelegram*` | **CLOSED** — service purge (`8ab8f76`) |
| A10-007…014 class | Partial closes on a11y/QueryState/kanban dead code; deeper InfiniteQuery/Gantt virt may remain P2 |
| A11-014 | TZ error-shape wording vs live AppError (docs) |
| A4-005/008/009…012, A2-011, A3-013/014/018–020 | Block F time-box residuals (W18) |
| A6-007…009, A5-012…015, A7-010…012 | W19 out of time-box |
| A14-008 | access-log without InfoContext (P2 obs) |

### Dead code / deps

FE telegram client / notify_mode / TelegramButton removed (W11/W13); BE login surface purged (`8ab8f76`). Unused FE wrappers (teams/trash/RBAC/imex HTTP) remain optional product debt.

---

## 5. 261 findings scoreboard (approx)

Canonical unique count from `00-SUMMARY.md`: **262** (user shorthand «261» ≈ same set; 12 P0 + 77 P1 + 120 P2 + 53 P3).

| Bucket | Approx | Notes |
|--------|-------:|-------|
| Unique findings (deduped) | **262** | `00-SUMMARY.md` |
| Closed in FIX-W0…W19 (IDs with FIXED/CLOSED in reports) | **~110–140** | Regex over FIX reports under-counts (~81) where waves used ✅/SHA tables without keyword; commit volume W1–W19 supports **≥120** discrete closes |
| Still open / deferred / blocked / P2–P3 untouched | **~100–140** | Mostly P2/P3 polish, FE debt, Block F leftovers, ops MANUAL (schema A8 cleared) |
| **Remaining P0 (security/product)** | **0** full ALIVE | All 12 original P0 primary paths closed; RT **12 / 0 / 0** |
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

#### Still open (accurate — residual-close 2026-07-28)

1. **A14-009 OTEL** — FIXED\* repo; **MANUAL** enable endpoint (`68cbd14`).
2. **A14-002/003** — FIXED\* fail-closed; **MANUAL** rotate/recreate if old defaults deployed.
3. **PG-03 / FA-06** secret rotation — sole prod cutover blocker.
4. **Hubtest gate artifact** — STALE vs HEAD; re-run required.
5. **Block F leftovers** A4-005+, A3-013+, A5-012+, A7-010+, A14-008 (P2/P3).
6. **COV-01 floor** — 81.9% &lt; 82% (2026-07-27); security 6-pack not remeasured 2026-07-27/28 — **no invented remeasure**.
7. **GetProject non-member metadata** — product-grey GAP (W17).
8. **TZ AppError wording** A11-014 / docs-only.
9. **A14-005 ops residual** — compose default tightened; blanket RFC1918 in `.env` can re-open Docker peers.

Schema A8-005/006 + A8-001/002 indexes, RT-003, A3-005 race, Telegram BE purge, A11 SPA wire, A15-004 — **CLOSED** (not listed above).

---

## 6. Verdict

Grok FIX program W0–W19 + residual-close closed the **critical insider Broken Access Control** surface from A19: **12 CLOSED / 0 PARTIAL / 0 ALIVE** RT chains. **RT-005** and **RT-003** CLOSED 2026-07-28. Schema soft-delete/indexes CLOSED. Planning/coverage truth is **81.9%** (2026-07-27), not the stale 83.0 narrative alone. True residuals are ops MANUAL (A14-009 OTEL, A14-002/003, FA-06) + stale hubtest artifact.

**Recommended next (human / next wave):**

1. Re-run hubtest gate on HEAD.
2. Ops: FA-06 + A14-009 OTEL enable (+ A14-002/003 rotate if needed).
3. Optional coverage uplift wave to restore ≥82% (remeasure only — do not invent %).

*Merge to `master` — only by user.*
