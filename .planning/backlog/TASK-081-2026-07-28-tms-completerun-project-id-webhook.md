# TASK-081 — ♻️🐛 TMS CompleteTestRun: project_id в webhook payload

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: A2W4-5-001 REGRESS · Stage A2 dedup · TASK-008

## [OUTCOME]

Событие test_run.completed доставляется в project-scoped webhook: в payload есть project_id, extractEventProjectID его видит.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tms/service_runs.go
  - backend/testframework/tms/**
  - backend/testframework/apigateway/** (если IT на dispatch)
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - apigateway extractEventProjectID fail-closed логику TASK-012
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W4-5-001 · A2-DEDUP-2026-07-28
- tms/service_runs.go:363-368 — PublishJSON без project_id
- apigateway/webhook.go:220-228,393-419
- TASK-012 (fail-closed без project → drop)
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - CompleteTestRun публикует project_id (UUID строка)
  - IT/unit: webhook endpoint проекта A получает completed run проекта A; без project_id — по-прежнему fail-closed
  - Мутация C: убрать project_id из payload → тест краснеет
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
