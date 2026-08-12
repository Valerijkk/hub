# TASK-102 — 📐 Telegram login surface: ADR-AUTH-01 cleanup

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W4  
> Источник: A2W6 Telegram · ADR-AUTH-01 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be1 · ветка `fix/auth/task-102-telegram-login`

## [OUTCOME]

Telegram не является login: API/тесты login удалены; остаётся только notification bind (ADR-AUTH-01).

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/auth/**
  - testframework/auth/**
  - frontend auth telegram widgets если есть
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать notification bind
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **owner апрув не требуется** (делегация 2026-07-28) — Magic Link only / Telegram = bind only

## [REFERENCES]

- Finding IDs: A2W6-A2-001, A2W6-A1-001/002/005/006, A2W6-A2-002/003 (7)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - login endpoints/tests gone or 410 ✅ (deleted)
  - positive Telegram-login test removed ✅
  - bind-only path green ✅ (notification не тронут)
  - Каждый ID из списка — FIXED с evidence ✅
- Stop when: green + статус «Исправлена» в README ✅

## [OUTPUT SCHEMA]

### Checklist ID → evidence

| ID | Status | Evidence |
|----|--------|----------|
| A2W6-A1-001 | FIXED | deleted `telegram_bot.go` (`handleLogin`/`VerifyCode`/`tglogin:`) |
| A2W6-A1-002 | FIXED | deleted `telegram.go` (`VerifyTelegramAuth`) |
| A2W6-A1-005 | FIXED | `events.go` Method comment → magic_link only |
| A2W6-A1-006 | FIXED | `GetUserByTelegramID` удалён из auth Repository |
| A2W6-A2-001 | FIXED | deleted `HandleLogin_StoresCode`; denial `TestAuth_NoTelegramLoginSurface_Task102` |
| A2W6-A2-002 | FIXED | login suite удалён из `resend_telegram_test.go` + coverage/polling |
| A2W6-A2-003 | FIXED | deleted widget/WebApp `VerifyTelegramAuth_*` tests |

### Mutation note

Вернуть `VerifyTelegramAuth` / `tglogin:` / `TelegramBotAuth` в `internal/auth` или позитивный login-тест в `testframework/auth` → краснеет `TestAuth_NoTelegramLoginSurface_Task102` / `TestAuth_NoPositiveTelegramLoginTests_Task102`.

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
