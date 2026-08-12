# MUTATION W13 slot 3 refill — TASK-177 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-177-2026-07-29-c-w19-deactivate-ban-ws-kill.md` `[MUTATION CRITERIA]`  
> Findings: **C-W19-001**, **C-W19-003** (2 IDs)  
> Коммит не создавался. Prefer LiveNATS PROVEN (stub mock≠Ent не sole).

## Verdict

**2/2 PROVEN** · TASK-177 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| C-W19-001 | **PROVEN** | `TestDeactivateUser_LiveNATS_UnregistersWS_C_W19` (+ unit `TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19`) |
| C-W19-003 | **PROVEN** | `TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003` · `TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: C-W19-001 stub call-count only (not used as sole PROVEN)` · `docs-only IDs: (none)`

## Per-ID evidence

### C-W19-001 — Deactivate → Redis+NATS revoked → WS unregister

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/ws/ -run TestDeactivateUser_LiveNATS_UnregistersWS_C_W19
  go test -tags hubtest -count=1 ./testframework/auth/ -run TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19
  -> PASS (~1.1s / ~1.1s)

BREAK (prefer LiveNATS path):
  backend/internal/auth/service_session.go RevokeAllUserSessions
  — after invalidateUserSessions: skip both publishSessionEnded branches
    (empty sid + per-session); return nil only
  — comment: MUTATED C-W19-001

  -> FAIL TestDeactivateUser_LiveNATS_UnregistersWS_C_W19 (2.08s)
       Condition never satisfied — WS must close after Deactivate
  -> FAIL TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19 (2.08s)
       expected NATS publish on auth.session.<uid>.revoked

RESTORE publishSessionEnded branches (original)
  -> PASS LiveNATS (~1.0s) · PASS RevokeAll unit (~1.0s)

git diff service_session.go → empty
MUTATED marker → отсутствует
```

Canon PROVEN = LiveNATS E2E (real SessionService + NATS + WS Bridge). Stub
`TestDeactivateUser_CallsSessionRevoker_*` = call-count only (mock≠Ent) — not sole.

### C-W19-003 — fail-closed admin error on session-kill fail

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/audit/ \
    -run "TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003|TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003"
  -> PASS (~1.1s)

BREAK:
  backend/internal/audit/service.go deactivateUser :422-428
  — replace both `return fmt.Errorf("revoke sessions on deactivate: %w", err)`
    with slog-only / best-effort (continue to InvalidateUserStatusCache + nil)
  — MUTATED C-W19-003 comments

  -> FAIL TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003
       An error is expected but got nil
  -> FAIL TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003
       An error is expected but got nil

RESTORE fail-closed returns
  -> PASS (~1.1s)

git diff service.go → empty
MUTATED marker → отсутствует
```

Layer OK: fail-closed is service-level; Ent N/A for mockAuditRepo/stubSessionRevoker.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

# 001 LiveNATS + RevokeAll unit
go test -tags hubtest -count=1 -timeout=60s ./testframework/ws/ \
  -run TestDeactivateUser_LiveNATS_UnregistersWS_C_W19
go test -tags hubtest -count=1 -timeout=60s ./testframework/auth/ \
  -run TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19

# 003 fail-closed
go test -tags hubtest -count=1 -timeout=60s ./testframework/audit/ \
  -run "TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003|TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003"
# BASELINE PASS · BREAK FAIL · RESTORE PASS (each ID)
```

## Gaps

- Stub `*_CallsSessionRevoker_*` not used as sole PROVEN (criteria flag mock≠Ent).
- Separate Ban LiveNATS not required — Ban shares `deactivateUser` kill path; Deactivate LiveNATS covers it.

## Handoff

- Evidence: этот файл · TASK-177 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored; no leftover MUTATED; no commit
