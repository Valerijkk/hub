# TASK-079 — GDPR Mongo eraser fail-closed + integration test

> Заведено: 2026-07-28 · Автор: W18-C5 (auditor) · Приоритет: **P2** · Статус: На проверке
> Источник: TASK-008 W18 catch-up · Findings: W18-C5-001, W18-C5-002  
> Отчёт: `.planning/audits/total-2026-07/W18-CATCHUP-PRODUCT.md`  
> Rework 2026-07-28: fail-closed nil coll + Mongo DeleteMany IT (`gdpr_mongo_eraser_task079_test.go`)

## [OUTCOME]

`MongoMessageEraser` не маскирует отсутствие Mongo как успех; есть hubtest на реальном `DeleteMany` по `sender_id`, мутация фильтра краснеет.

## [ROLE]

- Role: be1
- Responsibility: fail-closed eraser + integration test на слое Mongo
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: `backend/internal/audit/gdpr_mongo.go`; `backend/testframework/audit/**`; wiring при необходимости; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest`; один логический коммит на подкласс при необходимости
- Approval record: N/A

## [REFERENCES]

- `.planning/audits/total-2026-07/W18-CATCHUP-PRODUCT.md` (W18-C5-001/002)
- `backend/internal/audit/service.go` GDPR cascade + `SetMessageEraser`
- `backend/testframework/audit/gdpr_cascade_p1be34_test.go` (stub-only сегодня)
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria:
  1. `NewMongoMessageEraser(nil)` / `coll==nil` → `DeleteMessagesBySender` возвращает **error** (не `(0,nil)`).
  2. Hubtest против Mongo: insert messages → erase by sender → DeletedCount/остаток; чужой sender не тронут.
- **Мутация (C):** (a) вернуть `(0,nil)` при nil coll → cascade/unit RED; (b) фильтр `"sender_id"`→`"user_id"` → integration count RED
- Regression: `go test -tags hubtest -race ./testframework/audit/...`
- Evidence: вывод тестов, SHA
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое (Mongo), не только stub
- Co-authored-by запрещён
