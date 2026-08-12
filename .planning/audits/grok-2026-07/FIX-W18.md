# FIX-W18 — Block F handler polish (stream B)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be2  
**Правила:** same checklist as W17 · fix only CLEAR P1/P2 · 1 commit / finding · no schema · no push master · time-box

Домены: `tms`, `tracker`, `sprint`, `calls`, `notification`, `search`, `bases`, `storage`, `ai`.

Checklist: authz scoped · input validation · pagination limits · status 400/401/403/404/409 · `{error,request_id}` / AppError · 404 vs 403 foreign · no PII in logs · idempotency where dangerous.

---

## Итог fixes

| ID | Sev | Status | SHA | Notes |
|----|-----|--------|-----|-------|
| A4-001 | P1 | FIXED | `27f0374` | JQL `assignee` → `HasAssigneeWith(user.IDEQ)` |
| A4-002 | P1 | FIXED | `31002d7` + WEAK `c3bc077` | Ctrl+K shards: `deleted_at IS NULL` (+ files `status=confirmed`); **runtime Postgres** soft-delete exclusion |
| A4-006 | P1 | FIXED | `06c1856` + WEAK `1b19d13` | `ConfirmUpload` only `pending` → 409 if already confirmed; **HTTP** re-confirm |
| A4-007 | P2 | FIXED | `3074dca` | Download/`GetActiveByID` only `confirmed` |
| A2-008 | P2 | FIXED | `adbc03e` | `GetActiveCall` requires `IsChannelMember` → 403 |
| Block F board/label | P2 | FIXED | `dbd03cf` | `GetBoardData`/`DeleteLabel` scoped to `:projectId` |
| A3-016 | P2 | FIXED | `0a3a3ad` + scope `8e0c9f2` | `StartSprint` fail-closed on `FindActiveSprint` DB error; **2026-07-28:** `GetSprintInProject` before activate (RT-005 sibling) |
| A3-015 | P2 | FIXED | `bc5d61a` | `CheckWIPLimit` fail-closed on missing status / count error |
| A3-017 | P2 | FIXED | `369f852` | `ExecuteTestResult` sets `ExecutedByID` |

### WEAK close (2026-07-28 · qa-auto2)

| ID | Gap closed | Test SHA | Autotest |
|----|------------|----------|----------|
| A4-002 | Source-grep / fixture only; runtime soft-delete unproven | `c3bc077` | `TestPostgresSearch_ExcludesSoftDeleted_A4_002` |
| A4-006 | Service path only; HTTP 409 unproven | `1b19d13` | `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006` |

Prod fixes remain `31002d7` / `06c1856`; StartSprint project-scope fix `8e0c9f2`.

---

## Coverage matrix (sampled handlers)

Legend: **PASS** = checklist OK · **GAP** = residual (not clear P1/P2 / deferred) · **FIXED** = closed this wave.

### search

| Handler / path | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|----------------|-------|----------|------|--------|----------|---------|---------|-------------|
| JQL `ToEntPredicates` assignee | PASS project scope upstream | **FIXED** A4-001 | n/a | n/a | PASS | n/a | PASS | n/a |
| Ctrl+K `searchIssues/TC/projects/files` | PASS membership | PASS q | PASS limit | PASS | PASS | n/a | PASS | n/a · **FIXED** A4-002 soft-delete |
| Saved filters create/list | PASS membership (A4-003/004 prior) | PASS | PASS | PASS 403 | PASS | PASS | PASS | n/a |

Residual GAP: A4-005 Ollama↔SafeDial (arch) · A4-016 JQL OR · A4-018 Meili SafeDial — deferred.

### storage

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `ConfirmUpload` | PASS file project RBAC | PASS UUID/body | n/a | **FIXED** 409 re-confirm | PASS | PASS 404 | PASS | **FIXED** A4-006 pending-only |
| `GetDownloadURL` | PASS project view | PASS UUID | n/a | PASS 404 pending | PASS | PASS | PASS | **FIXED** A4-007 confirmed-only |
| Presign upload | PASS upload perm | PASS | n/a | PASS | PASS | n/a | PASS | n/a |

Residual GAP: A4-008 NF-09 presign size (ADR) — deferred.

### calls

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `GetActiveCall` | **FIXED** A2-008 member | PASS UUID | n/a | **FIXED** 403 | PASS | PASS | PASS | n/a |
| `CreateCall` / chat / recording / RaiseHand | PASS (A2-002…006 prior) | PASS | n/a | PASS 403 | PASS | PASS | PASS | n/a |

Residual GAP: A2-011 invite entropy — deferred (crypto + rate-limit design).

### tracker

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `GetBoardData` | PASS RBAC + **FIXED** board∈project | PASS UUID | PASS col cap A13-001 | PASS 404 foreign | PASS | **FIXED** 404 | PASS | n/a |
| `DeleteLabel` | PASS RBAC + **FIXED** label∈project | PASS UUID | n/a | PASS 404 foreign | PASS | **FIXED** 404 | PASS | n/a |
| Issue CRUD by key | PASS (A3-002 prior) | PASS | PASS | PASS | PASS | PASS | PASS | version lock |

Residual GAP: A3-013 MoveIssue version · A3-018 HasMore off-by-one — deferred.

### sprint

| Handler / svc | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------------|-------|----------|------|--------|----------|---------|---------|-------------|
| `StartSprint` | PASS RBAC + **FIXED** `GetSprintInProject` (`8e0c9f2`) | PASS UUID | n/a | PASS 409 conflict · **FIXED** A3-016 fail-closed | PASS | PASS | PASS | n/a |
| `CheckWIPLimit` | via move path | PASS | n/a | PASS 409 · **FIXED** A3-015 fail-closed | PASS | PASS | PASS | n/a |
| `CompleteSprint` | PASS project scope (A3-001) | PASS | n/a | PASS | PASS | PASS | PASS | n/a |

### tms

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| `ExecuteTestResult` | PASS run∈project (A3-008) | PASS | n/a | PASS | PASS | PASS | PASS | **FIXED** A3-017 ExecutedBy |
| TC/run CRUD | PASS (A3-003…012 prior) | PASS | PASS | PASS | PASS | PASS | PASS | A3-004 VersionEQ prior |

Residual GAP: A3-014 snapshot tx · A3-020 flaky N+1 — deferred.

### notification

| Handler / path | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|----------------|-------|----------|------|--------|----------|---------|---------|-------------|
| Push update/delete | PASS user-scope (A2-005) | PASS endpoint | n/a | PASS | PASS | PASS | PASS | n/a |
| JetStream consumer | n/a | n/a | n/a | n/a | n/a | n/a | PASS | PASS Nak (A2-010) |
| Dispatch prefs | n/a | n/a | n/a | n/a | n/a | n/a | PASS | PASS prefetch A13-005 |

No new clear P1/P2 in sample this wave.

### bases

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Base/table/record CRUD | PASS owner-scope | PASS | **PASS** cap A13-002 | PASS 403/404 | PASS | PASS | PASS | version |

Residual GAP: A4-013 CreateBase tx · cell allowlist already PASS — deferred tx polish.

### ai

| Handler / path | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|----------------|-------|----------|------|--------|----------|---------|---------|-------------|
| Chat / BYOK adapters | PASS user | PASS | n/a | PASS | PASS | n/a | PASS (no key log) | GAP quota races A4-010/011 |
| Offline/Ollama | n/a | SafeDial | n/a | GAP A4-005 | PASS | n/a | PASS | n/a |

Residual GAP: A4-005/009/010/011/012 — not clear one-line P1/P2 in time-box; need design.

---

## Tests run

```
go test -tags hubtest -race -count=1 ./testframework/search/ -run "A4_001|A4_002|Assignee"
go test -tags hubtest -race -count=1 ./testframework/storage/ -run "A4_006|A4_007|ConfirmUpload"
go test -tags hubtest -race -count=1 ./testframework/calls/ -run "GetActiveCall|A2_008"
go test -tags hubtest -race -count=1 ./testframework/tracker/ -run "BlockF|GetBoardData|DeleteLabel"
go test -tags hubtest -race -count=1 ./testframework/sprint/ -run "A3_015|A3_016|WIPLimit|StartSprint"
go test -tags hubtest -race -count=1 ./testframework/tms/ -run "A3_017|ExecuteTestResult"
→ PASS (domain suites)
```

## Out of scope / deferred

| Item | Why |
|------|-----|
| A4-005 Ollama vs SafeDial | Architectural allowlist — not a one-line fix |
| A4-008 NF-09 | Needs ADR / SeaweedFS constraint |
| A4-009…012 AI stream/quota | Larger reliability; not clear P1 IDOR |
| A3-013/014/018/019/020 | Larger tx/perf — later |
| A2-011 invite entropy | Crypto + rate-limit design |
| Schema / indexes | Forbidden this wave |

*Merge в `master` — только пользователем.*
