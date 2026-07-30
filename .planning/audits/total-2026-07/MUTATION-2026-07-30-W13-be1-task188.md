# MUTATION W13 slot 1 refill — TASK-188 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-188-2026-07-29-audit-be-push-telegram-rl.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**2/2 PROVEN** · TASK-188 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| ASN-001 | **PROVEN** | `TestPushSender_NoEntImport_ASN001` |
| ASN-002 | **PROVEN** | `TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### ASN-001 — PushSender без `*ent.Client` / ent import

```text
BASELINE TestPushSender_NoEntImport_ASN001 -> PASS (~1.0s)

BREAK: notification/push.go —
  добавить import "github.com/valerii/hub/backend/ent" + var _ *ent.Client
-> FAIL push_layering_asn001_test.go:31
   Should not be: "\"github.com/valerii/hub/backend/ent\""
   Messages: push.go must not import ent (ASN-001); use PushSubscriptionStore
RESTORE -> PASS (~0.9s)
```

Guard: `backend/internal/notification/push.go:43-76` — `PushSubscriptionStore`, без ent import.

### ASN-002 — ConfirmLinkCode fails-counter → ErrTelegramConfirmLimited / HTTP 429

```text
BASELINE TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002 -> PASS (~1.0s)

BREAK: telegram_bind.go ConfirmLinkCode —
  убрать failsKey ≥ tgLinkCodeVerifyMax check + recordLinkCodeFailure на invalid
-> FAIL telegram_linkcode_rl_asn002_test.go:67
   6-я попытка должна вернуть ErrTelegramConfirmLimited, got invalid or expired link code
   (+ correct code during «lockout» succeeds — err expected nil)
RESTORE -> PASS (~1.0s)
```

Guard: `telegram_bind.go:111-147` · `recordLinkCodeFailure` · handler map → 429.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/notification/ -run `
  "TestPushSender_NoEntImport_ASN001|TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002"
# FINAL GREEN PASS ~1.0s
```

## Gaps

- Docs-only: **N/A**
- Do not touch: NF-07 `ValidateOutboundURL`; Magic Link `magiclink:fails:` contract

## Handoff

- Evidence: этот файл · TASK-188 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after BREAK; `git diff` на `push.go` / `telegram_bind.go` — empty; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
