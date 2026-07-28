# TASK-116 — 🔒 Auth: AvatarURL scheme + ListSessions assert

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: A2W6 auth residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## Evidence (be1 · 2026-07-28)

| ID | Статус | Evidence |
|----|--------|----------|
| A2W6-A1-003 | FIXED | `validateProfileAvatarURL` в `service_profile.go`; `TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003` |
| A2W6-A1-004 | FIXED | `TestHandler_ListSessions_MarksCurrent` decode + ровно один `is_current` = JWT sid |

**Mutation:** убрать вызов `validateProfileAvatarURL` → whitelist-тест красный; `IsCurrent: false` в handler → MarksCurrent красный.

**Тесты:** `go test -tags hubtest -race -count=1 ./testframework/auth/` → PASS (~4.2s)

## [OUTCOME]

AvatarURL только безопасные schemes; ListSessions is_current покрыт assert-тестом.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/auth/**
  - testframework/auth/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - Telegram login (TASK-102)
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W6-A1-003, A2W6-A1-004 (2)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - whitelist + test
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
