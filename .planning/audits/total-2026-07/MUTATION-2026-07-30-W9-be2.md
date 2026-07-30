# MUTATION W9 slot 2 — TASK-117 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost · Go 1.25 · Postgres `localhost:5432`  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-117-2026-07-28-notification-mail-push-limits-tests.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**6/7 PROVEN · 1/7 N/A (layering gap)** · статус TASK-117 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W7-A1-002 | **PROVEN** | `TestHandler_PushSubscribe_InvalidEndpoint` | drop `errors.Is(..., ErrInvalidPushEndpoint)` → 400 | handler |
| W7-A1-003 | **PROVEN** | `TestService_SendAlert_*_W7_A1_003` (×3) | drop `effectivePolicy` / `reserveDailySlot` in `SendAlert` | mail service |
| W7-A1-004 | **PROVEN** | `TestAdminHandler_BroadcastFailedCount` | drop `resp.Failed++` (swallow err) | mail admin |
| W7-A1-005 | **N/A** | — | layering (`*ent.Client` / `ActiveUserEmailSource`) — нет runtime RED теста | handler DI |
| W7-A1-006 | **PROVEN** | `ContextCancelled` + `FormatsAndAttemptsSend` | panic on `ctx.Done()`; `formatTelegramMessage` → `"mutated-empty"` | telegram |
| W7-A1-007 | **PROVEN** | `SendAndQueueOverflow` + `SendTelegram_MissingUser` | `ExportNewTestTelegramSender` ignore cap; enqueue on missing user | telegram queue |
| W7-A1-008 | **PROVEN** | `ValidEndpointSendFails` | on `SendNotification` err → `DeletePushSubscription` | push sender |

Production restored (StrReplace reverse). Leftover MUTATED: **нет**.

## Per-ID evidence

### W7-A1-002 — InvalidPushEndpoint → 400

```text
BASELINE → PASS TestHandler_PushSubscribe_InvalidEndpoint
BREAK: remove ErrInvalidPushEndpoint → BadRequest branch in handler.go
→ FAIL expected: 400 (got 500)
RESTORE → PASS
```

### W7-A1-003 — alerts_enabled / daily_limit

```text
BASELINE → PASS AlertsDisabled / DailyLimit / StorePolicy
BREAK: remove effectivePolicy + reserveDailySlot from SendAlert
→ FAIL AlertsDisabled (want "mail alerts disabled", got nil)
→ FAIL DailyLimit (want "mail daily limit exceeded", got nil)
→ FAIL StorePolicy (want alerts disabled, got nil)
RESTORE → PASS ×3
```

### W7-A1-004 — Broadcast Failed count

```text
BASELINE → PASS TestAdminHandler_BroadcastFailedCount
BREAK: drop resp.Failed++ on send err
→ FAIL expected Failed=1
RESTORE → PASS
```

### W7-A1-005 — AdminHandler layering

```text
N/A — нет compile/AST/runtime теста на отсутствие *ent.Client в AdminHandler.
Evidence остаётся struct DI + var _ ActiveUserEmailSource (docs/layering).
Не fake-PROVEN.
```

### W7-A1-006 — telegram NotPanics / format

```text
BASELINE → PASS ContextCancelled + FormatsAndAttemptsSend
BREAK: panic on ctx.Done(); formatTelegramMessage → "mutated-empty"
→ FAIL ContextCancelled (NotPanics / panic value MUTATION…)
→ FAIL FormatsAndAttemptsSend ("mutated-empty" does not contain "Alert"/"Body text")
RESTORE → PASS ×2
```

### W7-A1-007 — queue overflow / missing user

```text
BASELINE → PASS SendAndQueueOverflow + SendTelegram_MissingUser
BREAK: ExportNewTestTelegramSender ignores cap (buf 100);
       sendTelegram enqueues on GetUser err
→ FAIL overflow message must be dropped
→ FAIL no telegram message for missing user
RESTORE → PASS ×2
```

### W7-A1-008 — https sub survives send fail

```text
BASELINE → PASS ValidEndpointSendFails (+ UnsafeEndpointDeleted)
BREAK: DeletePushSubscription on SendNotification err
→ FAIL valid https subscription must remain after send failure
   (UnsafeEndpointDeleted still PASS — file:// path unchanged)
RESTORE → PASS
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"

go test -tags hubtest -count=1 -timeout=180s ./testframework/mail/... ./testframework/notification/ \
  -run "TestHandler_PushSubscribe_InvalidEndpoint|TestService_SendAlert_(AlertsDisabled|DailyLimit|StorePolicy)_W7_A1_003|TestAdminHandler_BroadcastFailedCount|TestTelegramSender_sendMessage_(ContextCancelled|FormatsAndAttemptsSend)|TestTelegramSender_SendAndQueueOverflow|TestNotificationService_SendTelegram_MissingUser|TestPushSender_sendToSubscription_(ValidEndpointSendFails|UnsafeEndpointDeleted)" -v

# FINAL GREEN after restore → ok mail ~1.7s · notification ~2.0s
```

## Assumptions / unresolved

- **W7-A1-005** остаётся gap (layering без runnable RED) — per criteria + «не fake-PROVEN».
- Коммит не создавался.

## Handoff

TASK-117 **Закрыта** · evidence этот файл · README синхронизирован · be-ver → code-reviewer.
