# MUTATION W13 slot 1 refill — TASK-160 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-160-2026-07-29-a3-6-4-audit-core.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-160 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W6-4-001 | **PROVEN** | `TestDeleteUserGDPR_HardDeleteFailKeepsMongoIntact_A3W64_001` (+ `TestDeleteUserGDPR_MessageEraserFailAfterCommit_A3W64_001`) |
| A3W6-4-002 | **PROVEN** | `TestHandler_BanUser_SentinelValidation_A3W64_002` (+ `TestBanUser_SentinelIdentity_A3W64_002`) |
| A3W6-4-003 | **PROVEN** | `TestHandler_NoEntImport_A3W64_003` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W6-4-001` (order proven via mock eraser call count; guard in prod `service.go`)

## Per-ID evidence

### A3W6-4-001 — Mongo eraser after Ent Commit

```text
BASELINE TestDeleteUserGDPR_HardDeleteFailKeepsMongoIntact_A3W64_001
         TestDeleteUserGDPR_MessageEraserFailAfterCommit_A3W64_001 -> PASS (~0.8s)

BREAK: service.go DeleteUserGDPR — вернуть DeleteMessagesBySender внутрь WithTx
       до HardDelete/Commit; убрать post-Commit erase (+ eraseRetry path)
-> FAIL HardDeleteFail: eraser.calls not empty
       ("mongo erase must not run before successful Ent Commit")
-> FAIL MessageEraserFailAfterCommit: hardDeleted len 0
       (eraser fail aborts tx before HardDelete)
RESTORE -> PASS (~1.0s)
```

Note (mock≠Ent): тест бьёт AuditService + stub `MessageEraser` + mock cascade repo —
не dual-store Ent/Mongo IT. Kill на реальном prod guard в `service.go` → RED именованных тестов.

### A3W6-4-002 — Ban sentinels + errors.Is

```text
BASELINE TestHandler_BanUser_SentinelValidation_A3W64_002
         TestBanUser_SentinelIdentity_A3W64_002 -> PASS

BREAK: handler.go BanUser — errors.Is → err.Error() == "…"
       service.go — сменить текст ErrBanReasonRequired / ErrCannotBanSelf
-> FAIL empty_reason / self_ban: expect 400 got 500
-> FAIL SentinelIdentity: Equal text mismatch
RESTORE -> PASS
```

### A3W6-4-003 — requireAdmin без ent import

```text
BASELINE TestHandler_NoEntImport_A3W64_003 -> PASS

BREAK: handler.go — import ent + ent.IsNotFound в requireAdmin
       (вместо errors.Is(err, ErrUserNotFound))
-> FAIL assert.NotEqual ent import path
RESTORE -> PASS
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/audit/ -run `
  "TestDeleteUserGDPR_HardDeleteFailKeepsMongoIntact_A3W64_001|TestDeleteUserGDPR_MessageEraserFailAfterCommit_A3W64_001|TestHandler_BanUser_SentinelValidation_A3W64_002|TestBanUser_SentinelIdentity_A3W64_002|TestHandler_NoEntImport_A3W64_003"
# FINAL GREEN PASS ~1.0s
```

## Gaps

- A3W6-4-001: **mock≠Ent** documented — PROVEN на service-order via mock call count; нет отдельного Ent+Mongo IT на partial-erasure. Не NOT_PROVEN (kill recipe на prod service поймана).
- Docs-only: N/A

## Handoff

- Evidence: этот файл · TASK-160 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git status` на `backend/internal/audit/{service,handler}.go` — clean; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
