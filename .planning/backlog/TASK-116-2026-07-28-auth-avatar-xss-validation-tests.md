# TASK-116 — 🔒 Auth: AvatarURL scheme + ListSessions assert

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: A2W6 auth residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## Evidence (be1 · 2026-07-28)

| ID | Статус | Evidence |
|----|--------|----------|
| A2W6-A1-003 | FIXED | `validateProfileAvatarURL` в `service_profile.go`; `TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003` |
| A2W6-A1-004 | FIXED | `TestHandler_ListSessions_MarksCurrent` decode + ровно один `is_current` = JWT sid |

**Тесты:** `go test -tags hubtest -race -count=1 ./testframework/auth/` → PASS (~4.2s)

## Mutation prove (be1 · W8 slot 3 · 2026-07-30)

| ID | Verdict | Evidence |
|----|---------|----------|
| A2W6-A1-003 | **PROVEN** | drop `validateProfileAvatarURL` → RED named test → restore GREEN |
| A2W6-A1-004 | **PROVEN** | `IsCurrent: s.ID != currentSessionID` → RED named test → restore GREEN |

Отчёт: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W8-be1.md`

## [MUTATION CRITERIA] — verifier-ready

> Сверено по HEAD 2026-07-30 · be2 W7 slot 2 · статус **На проверке**

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W6-A1-003 | `backend/internal/auth/service_profile.go:37-38` (и/или `:46-47`) — убрать вызов `validateProfileAvatarURL`; **или** `:57-80` сделать always-nil | `TestProfileService_UpdateProfile_AvatarURL_SchemeWhitelist_A2W6_A1_003` | service |
| A2W6-A1-004 | `backend/internal/auth/handler.go:350` — `IsCurrent: false` (или `s.ID != currentSessionID`) | `TestHandler_ListSessions_MarksCurrent` | handler |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 003 | ProfileService + repo (Ent/sqlite env) | Guard в service → whitelist OK |
| 004 | Fiber handler ListSessions | JWT sid → `is_current` OK |

`layer_mismatch IDs: (none)`

### Gaps

- _(none)_ — оба ID имеют Break→RED якорь на prod guard.

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
- **Доказательство мутацией** — см. **`[MUTATION CRITERIA]`** выше (verifier-ready, 2026-07-30).

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
