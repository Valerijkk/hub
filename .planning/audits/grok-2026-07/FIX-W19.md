# FIX-W19 — Block F handler polish (stream A / business+integrations)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be1  
**Правила:** same checklist as W17 · fix only CLEAR P1/P2 · 1 commit / finding · no schema · no push master · time-box

Домены: `crm`, `hr`, `accounting`, `calendar`, `github`, `gitlab`, `imex`, `marketplace`, `intranet`, `mail`, `dashboard`, `backup`, `apigateway`, `status`, `dev`.

Checklist: authz scoped · input validation · pagination limits · status 400/401/403/404/409 · `{error,request_id}` / AppError · 404 vs 403 foreign · no PII in logs · idempotency where dangerous.

---

## Итог fixes

| ID | Sev | Status | SHA | Notes |
|----|-----|--------|-----|-------|
| A7-002 | P1 | FIXED | `2adcb1e` | `/dashboard/admin` → `auth.IsSystemAdmin` (founder\|\|admin) |
| A5-002 | P1 | FIXED | `2686bb4` | imex import/export + PDF: `RequirePermission` (view/create) |
| A6-005 | P1 | FIXED | `0823cc2` | calendar leave: own + shared-project members; no email in title |
| A7-005 | P1 | FIXED | `780212c` | mail `SendAlert` real Resend HTTP when key set (no silent stub) |
| A7-008 | P2 | FIXED | `31ee093` | broadcast `StatusEQ(active)` only |
| A7-007 | P2 | FIXED | `8083a60` | restore/backup HTTP + dump errors: generic client msg; stderr logged |
| A7-009 | P2 | FIXED | `8083a60` | pre-restore backup fail → abort restore |
| A6-006 | P2 | FIXED | `beaf382` | HR approve/reject/cancel require `version >= 1`; bind errors → 400 |

### WEAK close (2026-07-28 · be2)

| ID | Gap closed | Test/fix SHA | Autotest |
|----|------------|--------------|----------|
| A6-006 | CancelLeave swallowed bind errors; HTTP version<1 unproven | `cfc4959` | `TestHandler_ApproveLeave_VersionZero_A6_006`, `TestHandler_CancelLeave_BindAndVersion_A6_006` + CancelLeave bind→400 |
| A7-007 | Dump path only `Contains("pg_restore")` — not mutation-proof vs embedding CombinedOutput/DSN | `fbc20a1` | `TestRestorePostgres_DoesNotEmbedDumpOutput_A7_007`, `TestRestoreMongo_DoesNotEmbedDumpOutput_A7_007` |
| A7-008 | Broadcast fixture only pending — `StatusNEQ(pending)` would still pass | `ab580a3` | `TestAdminService_Broadcast` + deactivated user |

Prod fixes remain `beaf382` / `8083a60` / `31ee093`; this wave closes WEAK with prod guard + mutation-resistant tests.

---

## Coverage matrix (sampled handlers)

Legend: **PASS** = checklist OK · **GAP** = residual · **FIXED** = this wave · **PRIOR** = earlier FIX-W*.

### crm

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Contact/Deal CRUD | PRIOR A6-001 owner/admin | PASS stage | GAP unbounded list A6-009 | PASS | PASS | PASS | PASS | PRIOR A6-003 VersionEQ |
| CreateDeal `won` | PASS | PASS | n/a | PASS | PASS | n/a | PASS | GAP A6-007 invoice on create |

### hr

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Approve/Reject | PASS admin | **FIXED** version≥1 | n/a | **FIXED** 400/409 | PASS | PASS 404 vs 409 | PASS | **FIXED** A6-006 lock |
| CancelLeave | PASS owner | **FIXED** version≥1 | n/a | PASS 403 foreign | PASS | PASS | PASS | **FIXED** |

### accounting

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Invoice pay/create | PRIOR A6-002/004 admin+tx | PASS | PASS limit | PASS | PASS | PASS | PASS | PRIOR atomic pay |
| Journal | PASS admin | GAP AccountType A6-013 | PASS | GAP 500 enum | PASS | n/a | PASS | PASS tx |

### calendar

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| ListEvents / export.ics | **FIXED** A6-005 leave scope | PASS range | n/a | PASS 400 range | PASS | n/a | **FIXED** no email title | n/a |
| Calls/sprints/issues | PASS projectIDs | PASS | n/a | PASS | PASS | n/a | PASS | n/a |

### github / gitlab

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Integrations CRUD / branch / webhook | PRIOR A5-003…009 | PASS | n/a | PASS | PASS | PASS | PASS | PRIOR HMAC bind |
| ListGitHubLinks | GAP A5-013 auth-only | PASS | n/a | GAP | PASS | GAP | PASS | n/a |

### imex

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Import/export/PDF | **FIXED** A5-002 RBAC | PASS file size | n/a | PASS 401/403 | PASS | PASS | PASS | PRIOR A5-001 mappings |

### marketplace

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Listing update/status | PASS owner/admin | PRIOR A5-011 version≥1 | PASS ≤100 | PASS | PASS | PASS | PASS | PASS VersionEQ |

### intranet

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| News/kudos/coffee | PASS admin gates | PASS | PASS ≤100 | PASS | PASS | PASS | PASS | PRIOR A7-006 rematch |
| Kudos notify fail | PASS | PASS | n/a | GAP A7-011 retry dup | PASS | n/a | PASS | GAP |

### mail

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| SendAlert / hub `/v1/send` | PASS internal token | PASS | n/a | **FIXED** fail on Resend err | PASS | n/a | PASS | **FIXED** A7-005 real send |
| Broadcast | PASS admin | PASS | n/a | PASS | PASS | n/a | PASS | **FIXED** A7-008 active only |

### dashboard

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Project stats | PRIOR A7-001 membership | PASS UUID | n/a | PASS 403 | PASS | PASS | PASS | n/a |
| Admin stats | **FIXED** A7-002 founder | PASS | n/a | PASS 403 member | PASS | n/a | PASS | GAP A7-010 TopProjects WithIssues |

### backup

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Restore | PASS adminMw | PASS confirm | n/a | PASS | PASS | PASS | **FIXED** A7-007 | **FIXED** A7-009 pre-backup |
| Trigger/List/Delete | PASS | PASS UUID | n/a | PASS | PASS | PASS | **FIXED** generic err | PRIOR A7-004 fail→error |

### apigateway / status / dev

| Handler | authz | validate | page | status | AppError | 404≠403 | no PII | idempotency |
|---------|-------|----------|------|--------|----------|---------|---------|-------------|
| Webhooks / API keys | PASS HMAC+RBAC+SSRF | PASS | n/a | PASS | PASS | PASS | PASS | PASS Ack after enqueue |
| `/api/status` | public aggregate | n/a | n/a | PASS 503 | PASS | n/a | PASS | n/a |
| Dev routes | PASS `!release` + DEV_MODE | PASS | n/a | PASS 404 prod | PASS | n/a | PASS | n/a |

---

## Tests run

```
go test -tags hubtest -count=1 ./testframework/dashboard/ -run Admin_
go test -tags hubtest -count=1 ./testframework/imex/
go test -tags hubtest -count=1 ./testframework/calendar/ -run A6_005
go test -tags hubtest -count=1 ./testframework/mail/
go test -tags hubtest -count=1 ./testframework/backup/
go test -tags hubtest -count=1 ./testframework/hr/
→ PASS (touched packages)
```

Mutations (prod layer): A7-002 admin-only → founder 403 FAIL; A5-002 drop export RBAC → 200 FAIL; A6-005 drop requester filter → 3 leaves FAIL. Restored → green.

---

## Deferred / out of time-box

| ID | Why |
|----|-----|
| A6-007 CreateDeal won→invoice | P2 product hook; needs accounting wiring edge-case |
| A6-008/009 invoice number / CRM list unbound | P2; schema/ADR for unique / pagination |
| A5-012…015 gitlab escape / LimitReader / empty token | P2 residual integrations |
| A5-013 GitHub links auth-only | P2 authz; conflict risk with parallel GH work |
| A7-010 TopProjects WithIssues | P2 perf; Count rewrite |
| A7-011/012 intranet kudos/coffee tx | P2 partial-success |
| Schema indexes / soft-delete | Forbidden this wave |

*Merge в `master` — только пользователем.*
