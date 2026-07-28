# TASK-092 — 🔒 Search: filter HTTP status + Meili SafeDial

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: A2W5-A4 · Stage A2 dedup · TASK-008

## [OUTCOME]

Saved filter forbidden→403 (не 500); Meili HTTP client через NewSafeHTTPClient после ValidateOutboundURL.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/search/handler.go
  - backend/internal/search/meilisearch_indexer.go
  - backend/testframework/search/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - NF-03 soft-delete TASK-022 (закрыт)
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W5-A4-001/002/004
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ErrFilterForbidden → 403
  - SafeDial client
  - тесты + мутации
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
