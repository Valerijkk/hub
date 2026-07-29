# TASK-086 — ♻️ Webhook delivery: context + X-Hub-Delivery

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W3
> Источник: A2W4-5-002 · Stage A2 dedup · TASK-008

## [OUTCOME]

Исходящий webhook HTTP использует request context и стабильный X-Hub-Delivery; отмена/timeout не зависает.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/apigateway/webhook.go
  - backend/testframework/apigateway/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - HMAC skip / SSRF guards
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W4-5-002/003 · false-close TASK-043 · webhook.go:254-315
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - NewRequestWithContext
  - X-Hub-Delivery UUID на попытку
  - тест отмены ctx
  - Мутация: NewRequest без ctx → тест краснеет
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
