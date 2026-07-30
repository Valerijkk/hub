# TASK-117 — 📧 Notification/mail/push limits + tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: W7 notification · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W9-be2.md` · 6/7 PROVEN · W7-A1-005 N/A (layering)

## [OUTCOME]

Push URL→400; mail daily limit; broadcast Failed count; layering; assert-free tests исправлены.

## [EVIDENCE 2026-07-28 be2]

| ID | Статус | Evidence |
|----|--------|----------|
| W7-A1-002 | FIXED | `ErrInvalidPushEndpoint` → handler 400; тест `PushSubscribe_InvalidEndpoint` |
| W7-A1-003 | FIXED | `SendAlert` enforcement `alerts_enabled`/`daily_limit` (+ store); Status из policy |
| W7-A1-004 | FIXED | `BroadcastResponse.Failed` + slog на err |
| W7-A1-005 | FIXED | `AdminHandler` без `*ent.Client`; `ActiveUserEmailSource` |
| W7-A1-006 | FIXED | assert NotPanics + format в `telegram_wave97_test.go` |
| W7-A1-007 | FIXED | queue depth / empty recv asserts в lifecycle |
| W7-A1-008 | FIXED | https-sub остаётся в БД после send fail |

Tests: `go test -tags hubtest -race ./testframework/mail/... ./testframework/notification/` PASS (~20s)

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/notification/**
  - backend/internal/mail/**
  - testframework
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - Resend secrets в git
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A1-002..008 (7)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 7 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

> Сверено по HEAD 2026-07-30 · be2 W9 slot 2 · W7-A1-002..008 · статус **Закрыта** (mutation PROVEN)

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| W7-A1-002 | `handler.go:326-328` — убрать `errors.Is(..., ErrInvalidPushEndpoint)` → 400 (упасть в 500); **или** `push.go:252-254` не оборачивать в `ErrInvalidPushEndpoint` | `TestHandler_PushSubscribe_InvalidEndpoint` | handler (+ push SaveSubscription) |
| W7-A1-003 | `mail/service.go:93-98` — убрать `effectivePolicy` / `reserveDailySlot` из `SendAlert` | `TestService_SendAlert_AlertsDisabled_W7_A1_003` (+ `DailyLimit_W7_A1_003` / `StorePolicy`) | mail service |
| W7-A1-004 | `mail/admin_broadcast.go:69-72` — убрать `resp.Failed++` (глотать err / только Sent) | `TestAdminHandler_BroadcastFailedCount` | mail admin service |
| W7-A1-005 | `mail/admin_handler.go:12-17` — вернуть поле `*ent.Client` в `AdminHandler` / Broadcast без `ActiveUserEmailSource` | N/A runtime RED (нет AST/layering теста); evidence = struct DI + `var _ ActiveUserEmailSource` | handler layering |
| W7-A1-006 | `telegram_wave97_test.go` asserts **или** prod format/cancel path panic — снять `NotPanics` / `Contains` format | `TestTelegramSender_sendMessage_ContextCancelled` (+ `FormatsAndAttemptsSend`) | notification telegram (test asserts) |
| W7-A1-007 | queue depth / TryRecv asserts — `ExportNewTestTelegramSender` без drop на overflow; missing-user всё же enqueue | `TestTelegramSender_SendAndQueueOverflow` (+ `SendTelegram_MissingUser`) | notification telegram |
| W7-A1-008 | `push.go:211-214` — на `SendNotification` err вызывать `DeletePushSubscription`; **или** `:230-232` удалять при любом ≥400 (не только 404/410) | `TestPushSender_sendToSubscription_ValidEndpointSendFails` (https must remain); file:// — `UnsafeEndpointDeleted` | push sender |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 002 | Fiber + PushSender ValidateOutboundURL | Guard service→handler map — OK |
| 003 | mail Service policy | OK (Memory/Redis slot) |
| 004 | AdminHandler Broadcast + mock sender | OK |
| 005 | layering | нет RED теста |
| 006–007 | telegram unit | assert-hardening / queue — OK |
| 008 | Ent push sub + sendToSubscription | Ent OK |

`layer_mismatch IDs: (none)` · **gap ID:** W7-A1-005 (layering без runtime mutation)

### Gaps

- **W7-A1-005** — нельзя PROVEN Break→RED без нового compile/AST теста на отсутствие `*ent.Client` в `AdminHandler`.
- **W7-A1-006** — часть доказательств = test asserts (NotPanics/format); мутация prod panic-path слабее, чем у 002/003/008.

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
