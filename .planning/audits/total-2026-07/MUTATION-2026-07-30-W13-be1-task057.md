# MUTATION W13 slot 1 refill — TASK-057 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-057-be-p3-silent-error-swallow.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**5/7 PROVEN** (из них **2 PROVEN + GAP_NOTED** mock≠Ent) · **2 GAP** · leftover MUTATED: **нет**  
TASK-057 → **Закрыта**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A3-008 | **PROVEN** | `TestUpdateIssue_PublishFail_LogsWarn_TASK057` |
| W1-A4-012 | **PROVEN** + **GAP_NOTED** | `TestListSprints_CountsError_FailClosed_TASK057` |
| W1-A4-010 | **GAP** | *(нет dedicated TASK-057 RED)* — Ent Count-fail inject |
| W3-A3-010 | **PROVEN** | `TestGetAll_CorruptCounter_LogsWarn_TASK057` |
| W2-A4-013 | **PROVEN** | `TestExportRunResultsToCSV_WriteError_TASK057` |
| W6-A4-015 | **PROVEN** + **GAP_NOTED** | `TestBanUser` / `audit_log_failure_after_ban` |
| W7-A5-007 | **GAP** | *(нет dedicated TASK-057 RED)* — admin-check err→500 |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-012, W6-A4-015` · `docs-only IDs: (none)`

## Per-ID evidence

### W1-A3-008 — `logOutboxPublishErr` slog on Publish fail

```text
BASELINE TestUpdateIssue_PublishFail_LogsWarn_TASK057 -> PASS (~1.5s)

BREAK: tracker/repository.go logOutboxPublishErr —
  убрать slog.WarnContext("outbox publish failed", …)
-> FAIL outbox_publish_slog_task057_test.go:57
   expected slog warn on Publish err, got: (empty)
RESTORE -> PASS (~1.2s)
```

Guard: `backend/internal/tracker/repository.go:290-296` — Ent repo + failing outbox Publish.

### W1-A4-012 — ListSprints GetSprintIssueCounts err → Internal

```text
BASELINE TestListSprints_CountsError_FailClosed_TASK057 -> PASS

BREAK: sprint/service.go ListSprints —
  silent counts=make(map…) вместо NewInternal(…счётчиков…)
-> FAIL list_fail_closed_task057_test.go:48
   An error is expected but got nil.
RESTORE -> PASS
```

Guard: `backend/internal/sprint/service.go:72-76` — SprintService.

**GAP_NOTED** mock≠Ent: kill на service path PROVEN через `errBatchCountsRepo`; Ent GROUP BY Count-fail IT отсутствует (criteria flag).

### W1-A4-010 — velocity Count err → return err

```text
GAP — нет dedicated TASK-057 RED (нужен Ent Count-fail inject).
Guard в коде: sprint/repository.go:568-578 return nil, fmt.Errorf(…velocity committed count…).
```

### W3-A3-010 — unread ParseInt fail → slog + skip

```text
BASELINE TestGetAll_CorruptCounter_LogsWarn_TASK057 -> PASS

BREAK: messenger/unread.go GetAll —
  coerce n=0 без slog (вместо Warn + continue)
-> FAIL unread_corrupt_task057_test.go:37
   corrupt field must be skipped, not coerced to 0
   + slog assert: expected "unread counter corrupt"
RESTORE -> PASS
```

Guard: `backend/internal/messenger/unread.go:77-83` — UnreadService + miniredis.

### W2-A4-013 — CSV Write/Flush/Error → fmt.Errorf

```text
BASELINE TestExportRunResultsToCSV_WriteError_TASK057 -> PASS

BREAK: tms/exporter.go ExportRunResultsToCSV —
  _ = w.Write / ignore Flush Error
-> FAIL exporter_write_err_task057_test.go:37
   An error is expected but got nil.
RESTORE -> PASS
```

Guard: `backend/internal/tms/exporter.go:172-188` — pure exporter (no DB).

### W6-A4-015 — BanUser CreateAuditLog fail → return err

```text
BASELINE TestBanUser/audit_log_failure_after_ban -> PASS

BREAK: audit/service.go BanUser —
  swallow CreateAuditLog err (_ = err)
-> FAIL service_security_test.go:369
   An error is expected but got nil. (sub: audit_log_failure_after_ban)
RESTORE -> PASS
```

Guard: `backend/internal/audit/service.go:456-459` — AuditService.

**GAP_NOTED** mock≠Ent: kill на service path PROVEN через `mockAuditRepo.createErr`; Ent CreateAuditLog-fail IT отсутствует (criteria flag).

### W7-A5-007 — IsSystemAdmin err → wrap (не Forbidden)

```text
GAP — нет dedicated TASK-057 RED.
Guard в коде: marketplace/service.go:287-291 fmt.Errorf("check system admin: %w", err).
```

## Commands

```powershell
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
cd backend
go test -tags hubtest -count=1 -timeout=180s `
  -run "TestUpdateIssue_PublishFail_LogsWarn_TASK057$|TestListSprints_CountsError_FailClosed_TASK057$|TestGetAll_CorruptCounter_LogsWarn_TASK057$|TestExportRunResultsToCSV_WriteError_TASK057$|TestBanUser$" `
  ./testframework/tracker/ ./testframework/sprint/ ./testframework/messenger/ ./testframework/tms/ ./testframework/audit/
# MUTATED: FAIL ×5 named; FINAL restore: ok … ~0.9–1.4s each
```

## Leftover check

`Select-String MUTATED` on `tracker/repository.go` / `sprint/service.go` / `messenger/unread.go` / `tms/exporter.go` / `audit/service.go` → empty.
