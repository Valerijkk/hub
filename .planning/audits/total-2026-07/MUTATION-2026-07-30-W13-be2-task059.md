# MUTATION W13 slot 3 refill — TASK-059 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/legacy/TASK-059-be-p3-mutation-proof-gaps.md` `[MUTATION CRITERIA]`  
> Findings: **10** runnable IDs (P3-BE-03)  
> Коммит не создавался. leftover MUTATED: **нет**.

## Verdict

**9/10 PROVEN** · **W1-A4-015 GAP_NOTED** (mock≠Ent) · **0 NOT_PROVEN** · TASK-059 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A4-015 | **GAP_NOTED** | `TestAddIssuesToSprint_RejectsMissingIssueMapping_W1A4015` (Ent kill GREEN; mock control RED) |
| W2-A5-009 | **PROVEN** | `TestGetTableView_CrossBaseTable_NotFound` |
| W4-A2-007 | **PROVEN** | `TestPublisher_PublishInTx_RollbackLeavesNoEvents` |
| W4-A3-009 | **PROVEN** | `TestHTTPMetricsMiddleware_RouteIsTemplate_NotRawUUID` |
| W4-A3-010 | **PROVEN** | `TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer` |
| W5-A4-012 | **PROVEN** | `TestToEntPredicates_ORUnionVsAND_EntCount_W5A4012` |
| W5-A1-011 | **PROVEN** | `TestUserHandler_CreateProvider_RedactsAPIKey_W5A1011` |
| W5-A3-019 | **PROVEN** | `TestAuditRepoLogger_LogAIChat_NoPromptInDetails_W5A3019` |
| W6-A5-006 | **PROVEN** | `TestRestoreIssue_ClearsDeletedAt_W6A5006` |
| W7-A5-016 | **PROVEN** | `TestHandler_DeleteListing_Forbidden_W7A5016` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-015` · `docs-only IDs: (none)` · `GAP IDs: W1-A4-015`

## Per-ID evidence

### W1-A4-015 — Ent `SetIssuesSprint` n≠len → `ErrIssueNotInProject` (mock≠Ent)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/sprint/ \
    -run TestAddIssuesToSprint_RejectsMissingIssueMapping_W1A4015$
  -> PASS (~0.84s)

BREAK (EntRepository correct layer — REAL prod):
  backend/internal/sprint/repository.go SetIssuesSprint
  — убрать Count/n!=len + HasProjectWith на Update
  — MUTATED W1-A4-015

  -> PASS (~0.90s)  ← Exact RED остаётся GREEN: тест бьёт mock, не Ent

RESTORE SetIssuesSprint
  -> PASS (~0.77s)

CONTROL (mock layer — не prod):
  backend/testframework/sprint/service_test.go mock SetIssuesSprint
  — silent-skip missing issueProject
  -> FAIL (~0.83s)
       expected error, got nil
  RESTORE mock → PASS (~0.77s)

git diff sprint/repository.go + service_test.go → empty
MUTATED marker → отсутствует
```

**GAP_NOTED (mock≠Ent):** kill реального Ent-guard не краснит Exact RED; RED доказан только на mock control. Sibling Ent IT — вне этого Exact path.

### W2-A5-009 — `HasBaseWith` cross-base table

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/bases/ \
    -run TestGetTableView_CrossBaseTable_NotFound$
  -> PASS (~0.62s)

BREAK (EntRepository correct layer):
  backend/internal/bases/repository.go GetTable + GetTableMeta
  — убрать hubtable.HasBaseWith(...)
  — MUTATED W2-A5-009

  -> FAIL (~0.67s)
       Expected ErrNotFound but got nil

RESTORE HasBaseWith
  -> PASS (~0.65s)

git diff repository.go → empty
```

Layer OK: Ent + HTTP GetTableView.

### W4-A2-007 — `PublishInTx` + Rollback → 0 rows

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/platform/outbox/ \
    -run TestPublisher_PublishInTx_RollbackLeavesNoEvents$
  -> PASS (~0.69s)

BREAK (outbox Publisher correct layer):
  backend/internal/platform/outbox/publisher.go PublishInTx
  — Save + tx.Commit() внутри PublishInTx
  — MUTATED W4-A2-007

  -> FAIL (~0.71s)
       Rollback: sql: transaction has already been committed or rolled back

RESTORE return err after Save
  -> PASS (~0.76s)

git diff publisher.go → empty
```

Layer OK: Ent tx outbox.

### W4-A3-009 — `httpRouteLabel` template (не raw UUID)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/platform/observability/ \
    -run TestHTTPMetricsMiddleware_RouteIsTemplate_NotRawUUID$
  -> PASS (~1.09s)

BREAK (observability correct layer):
  backend/internal/platform/observability/route_label.go
  — return c.Path()
  — MUTATED W4-A3-009

  -> FAIL (~0.86s)
       does not contain route="/api/items/:id"
       (raw UUID in metrics label)

RESTORE httpRouteLabel template
  -> PASS (~0.90s)

git diff route_label.go → empty
```

Layer OK: Fiber HTTP metrics middleware.

### W4-A3-010 — metrics CIDR / proxy peer deny

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/platform/observability/ \
    -run TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer$
  -> PASS (~0.90s)

BREAK (observability correct layer):
  backend/internal/platform/observability/metrics_allowlist.go
  — authorizeMetricsIP always nil
  — MUTATED W4-A3-010

  -> FAIL (~0.90s)
       expected: 403 / actual: 200

RESTORE CIDR check
  -> PASS (~0.87s)

git diff metrics_allowlist.go → empty
```

Layer OK: Fiber MetricsEndpoint + X-Forwarded-For peer.

### W5-A4-012 — JQL OR → Ent union (не AND-only)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/search/ \
    -run TestToEntPredicates_ORUnionVsAND_EntCount_W5A4012$
  -> PASS (~0.71s)

BREAK (search ToEntPredicates correct layer):
  backend/internal/search/jql.go
  — ignore Logic OR (always AND)
  — MUTATED W5-A4-012

  -> FAIL (~0.70s)
       OR count expected 2, actual 0

RESTORE OR flushAnd
  -> PASS (~0.82s)

git diff jql.go → empty
```

Layer OK: Ent predicate count on sqlite hubtest.

### W5-A1-011 — `toProviderResponse` redact (нет raw `api_key`)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/ai/ \
    -run TestUserHandler_CreateProvider_RedactsAPIKey_W5A1011$
  -> PASS (~0.75s)

BREAK (AI DTO/service correct layer):
  backend/internal/ai/dto.go + service.go toProviderResponse
  — json:"api_key" + APIKey: "sk-leaked-mutated"
  — MUTATED W5-A1-011

  -> FAIL (~1.59s)
       response should not contain "api_key"

RESTORE redact-only HasAPIKey
  -> PASS (~1.00s)

git diff dto.go + service.go → empty
```

Layer OK: Fiber CreateProvider JSON.

### W5-A3-019 — `LogAIChat` details без prompt

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/ai/ \
    -run TestAuditRepoLogger_LogAIChat_NoPromptInDetails_W5A3019$
  -> PASS (~1.05s)

BREAK (AI audit logger correct layer):
  backend/internal/ai/audit.go LogAIChat
  — details["prompt"] = …
  — MUTATED W5-A3-019

  -> FAIL (~1.09s)
       audit details must not contain "prompt"

RESTORE metadata-only details
  -> PASS (~0.82s)

git diff audit.go → empty
```

Layer OK: Ent AuditLog details.

### W6-A5-006 — `RestoreIssue` → `ClearDeletedAt`

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/audit/ \
    -run TestRestoreIssue_ClearsDeletedAt_W6A5006$
  -> PASS (~0.84s)

BREAK (audit repository correct layer):
  backend/internal/audit/repository.go RestoreIssue
  — noop без ClearDeletedAt
  — MUTATED W6-A5-006

  -> FAIL (~0.99s)
       Expected DeletedAt nil, got time.Time

RESTORE ClearDeletedAt
  -> PASS

git diff repository.go → empty
```

Layer OK: Ent RestoreIssue.

### W7-A5-016 — marketplace `DeleteListing` owner/admin → 403

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/marketplace/ \
    -run TestHandler_DeleteListing_Forbidden_W7A5016$
  -> PASS

BREAK (marketplace service correct layer):
  backend/internal/marketplace/service.go DeleteListing
  — убрать owner/admin check
  — MUTATED W7-A5-016

  -> FAIL
       expected: 403 / actual: 204 (stranger delete)

RESTORE owner/admin guard
  -> PASS

git diff service.go → empty
```

Layer OK: Fiber DELETE + service RBAC.

## Residual class gaps (criteria)

Docs-only: **N/A**. Class residuals без `*_task059_test.go` (poker RBAC / soft-delete / calls recordings / AI stream denials / …) — **GAP_NOTED** вне runnable Exact RED этой TASK (не блокируют Закрыта).

## Cleanup

- `git checkout --` на все touched prod/testframework files после каждого ID
- `rg MUTATED` по touched paths → пусто
- Коммит не создавался
