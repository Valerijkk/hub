# TASK-106 — 📐 Intranet FR-INTR-11: auto-DM vs notify

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: **Исправлена**  
> Источник: W7-A2-008 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be2 · ветка `fix/intranet/task-106-coffee-dm`

## [OUTCOME]

FR-INTR-11 реализован по ADR U05 (**notify-only**) — поведение не «полусделано».

## [ROLE]

- Role: product→be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/intranet/**
  - backend/testframework/intranet/**
  - `.planning/decisions/ADR-INTRANET-U05-COFFEE-NOTIFY-2026.md`
  - `docs/system_specification.md` (FR-INTR-11 sync)
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/` (+ SPEC-COVERAGE)
- Do not touch:
  - спам DM / CreateDM без отдельного ADR
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **апрув ADR не требуется** (owner 2026-07-28) — ADR+код+тесты в одном цикле

## [REFERENCES]

- Finding IDs: W7-A2-008 U05 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-INTRANET-U05-COFFEE-NOTIFY-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR U05 ✅ notify-only
  - код+тест выбранного варианта ✅
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record — **снято owner 2026-07-28**

## Evidence (be2 2026-07-28)

| ID | Вердикт | Evidence |
|----|---------|----------|
| W7-A2-008 / P2-BE-U05 | **FIXED** | ADR-INTRANET-U05 notify-only; `CoffeeMatchDeliveryMode`; `notifyCoffeeMatch`+slog; `coffee_notify_task106_test.go` |

**Тесты:**
```
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/intranet/ -run Task106
```

**Мутация:** убрать вызовы `s.notifyCoffeeMatch` в `RunCoffeeMatching` → краснеют
`TestCoffeeMatch_NotifyOnly_Task106` / `TestCoffeeMatch_NotifyOnly_Triple_Task106`.
