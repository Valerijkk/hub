# TASK-117 — 📧 Notification/mail/push limits + tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке
> Источник: W7 notification · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

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

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
