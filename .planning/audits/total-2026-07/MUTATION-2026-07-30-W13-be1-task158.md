# MUTATION W13 slot 1 — TASK-158 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-158-2026-07-29-a3-6-2-auth-tests.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task158.py`  
> Raw: `_mut-batch-w13-task158.json`  
> Коммит не создавался (инструкция слота).

## Verdict

**4/4 PROVEN** · TASK-158 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W6-2-001 | **PROVEN** | `TestLoginWithMagicLink_DeactivatedUserDenied` |
| A3W6-2-002 | **PROVEN** | `TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder` |
| A3W6-2-003 | **PROVEN** | `TestInvalidateUserSessions_ClearsMultiple` |
| A3W6-2-004 | **PROVEN** | `TestResendEmailClient_SendMagicLinkEmail_Success` |

`layer_mismatch IDs: (none)` — 001/002/003 service; 004 HTTP client (`resend.go`).

## Per-ID evidence

### A3W6-2-001 — login deny non-Active (deactivated)

```text
BASELINE TestLoginWithMagicLink_DeactivatedUserDenied -> PASS (~5.5s)
BREAK: service_login.go — `u.Status != StatusActive` → `u.Status == StatusPending`
-> FAIL ErrorIs ErrAccountNotActive got nil (deactivated получил JWT) (~5.9s)
RESTORE -> PASS (~4.9s)
```

### A3W6-2-002 — firstUserMu + tx founder race (SEC-06)

```text
BASELINE TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder -> PASS (~4.7s)
BREAK: service_registration.go CreateFirstOrMemberUser — убрать firstUserMu + bypass WithEntTx
      (прямой s.repo.CreateFirstOrMemberUser)
-> FAIL require.NoError: create user: database table is locked: users (~5.7s)
RESTORE -> PASS (~5.1s)
```

Note: RED mode на sqlite in-memory — lock/race без сериализации, не обязательно `Founder>1`.
Именованный тест всё равно краснеет при снятии guard (mutex+tx). Dual-Founder assert
остаётся живым на GREEN path.

### A3W6-2-003 — Redis Del after InvalidateUserSessions

```text
BASELINE TestInvalidateUserSessions_ClearsMultiple -> PASS (~5.2s)
BREAK: service_session.go invalidateUserSessions — skip pipe.Del/Exec (DB delete оставить)
-> FAIL mr.Exists(ExportRefreshRedisKey) — redis refresh key must be deleted (×2) (~7.3s)
RESTORE -> PASS (~6.2s)
```

### A3W6-2-004 — Resend body to/code

```text
BASELINE TestResendEmailClient_SendMagicLinkEmail_Success -> PASS (~6.6s)
BREAK: resend.go SendMagicLinkEmail — To=[] + HTML без code
-> FAIL Contains toEmail / Contains code (~7.5s)
RESTORE -> PASS (~6.4s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/auth/ -run \
  "TestLoginWithMagicLink_DeactivatedUserDenied|TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder|TestInvalidateUserSessions_ClearsMultiple|TestResendEmailClient_SendMagicLinkEmail_Success"
# BASELINE PASS ~1.3s package

python .planning/audits/total-2026-07/_run_mutations_w13_task158.py
# exit 0 · 4/4 PROVEN · leftover_mutated=false · ~73s
```

## Gaps

- none runnable — all 4 criteria PROVEN  
- 002 RED symptom = sqlite table lock (не dual Founder count); kill recipe «Founder>1» —
  эквивалентно по смыслу (concurrent path без mutex/tx ломает тест); noted, не GAP_NOTED

## Handoff

- Evidence: этот файл · TASK-158 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на auth prod files — empty; no `MUTATED` leftover  
- Prod code unchanged this slot (test-only close already FIXED 2026-07-29; mutation prove only)
