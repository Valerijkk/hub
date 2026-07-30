# MUTATION W9 slot 1 — TASK-090 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost · Criteria: `.planning/backlog/archive/2026-07-28/TASK-090-2026-07-28-audit-purge-error-contract.md` `[MUTATION CRITERIA]`
> Коммит не создавался (инструкция слота).

## Verdict

**Runnable PROVEN: 3/3** (A2W6-A4-001, A2W6-A5-002, A2W6-A4-007)  
Статус TASK-090 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A2W6-A4-001 | **PROVEN** | `TestAuditService_PurgeExpired_PropagatesError` | `PurgeExpired` → `return nil` после slog (не `errors.Join`) | service |
| A2W6-A5-002 | **PROVEN** | `TestAuditService_PurgeExpired_PropagatesError` | тот же guard (swallow purge err) | service |
| A2W6-A4-007 | **PROVEN** | `TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007` + `TestHandler_RestoreItem_Fixed500Message_A2W6A4007` | List + Restore → `fiber.NewError(500, err.Error())` | handler |

Production restored (`git checkout --` после каждого BREAK). `git diff` на `service.go` / `handler.go` пуст.

## Per-ID evidence

### A2W6-A4-001 — Purge fail-closed Join

```text
BASELINE → PASS (~3.2s) all three named tests

BREAK: service.go PurgeExpired — return nil вместо errors.Join(errs...)
→ FAIL TestAuditService_PurgeExpired_PropagatesError
   (An error is expected but got nil)
RESTORE (git checkout -- backend/internal/audit/service.go)
→ PASS same test (final package ~3.3s)
```

### A2W6-A5-002 — Swallow purge err (same guard)

```text
Same BREAK as A4-001 (criteria: тот же guard service.go Join path).
→ FAIL TestAuditService_PurgeExpired_PropagatesError (got nil)
RESTORE → PASS
Not fake-PROVEN: independent criteria ID, shared mutation anchor + shared RED test as documented in TASK-090.
```

### A2W6-A4-007 — Admin 500 fixed RU message

```text
BREAK: handler.go ListAuditLogs + RestoreItem — fiber.NewError(500, err.Error())
→ FAIL TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007
   (body contains leak "hunter2"; missing fixed RU message)
→ FAIL TestHandler_RestoreItem_Fixed500Message_A2W6A4007
   (body contains leak "token=abc"; missing fixed RU message)
RESTORE (git checkout -- backend/internal/audit/handler.go)
→ PASS both (~3.3s)
```

## Commands

```text
cd backend
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/audit/ \
  -run "TestAuditService_PurgeExpired_PropagatesError|TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007|TestHandler_RestoreItem_Fixed500Message_A2W6A4007"
→ ok  ~3.2–3.3s (baseline + final restore)
```

## Assumptions / unresolved

- Коммит не создавался.
- A4-001 и A5-002 делят один guard и один тест — оба PROVEN одним BREAK/RED/RESTORE циклом (как в criteria).
- Gaps из TASK (Ent SQL purge_*, residual admin 500 paths) не мутировались — вне runnable criteria.
- `layer_mismatch IDs: (none)`

## Handoff

TASK-090 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
