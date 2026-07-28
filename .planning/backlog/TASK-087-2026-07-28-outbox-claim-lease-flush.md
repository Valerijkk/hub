# TASK-087 — 🐛 Outbox: короткий claim lease на время flush

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: A2W4-2-001 · Stage A2 dedup · TASK-008

## [OUTCOME]

SKIP LOCKED claim не держит строку на всё время NATS publish; lease/heartbeat или publish вне длинной tx.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/platform/outbox/**
  - backend/testframework/**/outbox*
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - менять JetStream subjects без ADR
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W4-2-001 NEW · outbox/worker.go:62-196
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Документированный lease или publish after commit claim
  - тест concurrent worker
  - Мутация: hold-during-flush → тест краснеет
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
