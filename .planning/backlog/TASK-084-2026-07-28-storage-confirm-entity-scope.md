# TASK-084 — 🔒 Storage Confirm: scope для non-issue entity

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: W7-A4-002 · Stage A2 dedup · TASK-008

## [OUTCOME]

ConfirmUpload для message/avatar/voice/channel проверяет membership/ownership, а не только непустой entityID.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/storage/service.go
  - backend/internal/storage/repository*.go
  - backend/testframework/storage/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - presigned NF-09 size==0 (deferred)
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- W7-A4-002 · storage/service.go:232-256
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Для каждого allowlist типа — конкретная проверка (channel membership / user avatar self / message channel membership)
  - Denial-тесты foreign entity
  - Мутация: вернуть только entityID!="" → denial краснеет
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
