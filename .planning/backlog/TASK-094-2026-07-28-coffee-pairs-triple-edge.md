# TASK-094 — ☕ Coffee: Pairs[] без triple-edge дублей

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Исправлена  
> Источник: A2W7-2-N01 · Stage A2 dedup · TASK-008

## [OUTCOME]

API coffee pairs не включает рёбра троек как отдельные пары либо PairCount согласован с контрактом FR-INTR-10.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/intranet/coffee_service.go
  - backend/testframework/**/coffee*
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - менять unique constraint coffee (TASK-025)
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W7-2-N01 NEW · coffee_service.go:150-204
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Pairs[] без triple edges ИЛИ документированный контракт + тест
  - Мутация → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
