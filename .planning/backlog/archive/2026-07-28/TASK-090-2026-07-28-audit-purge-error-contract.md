# TASK-090 — 🔒 Audit: purge fail-closed + admin error contract

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W9-be1.md` · PROVEN 3/3
> Источник: A2W6 · Stage A2 dedup · TASK-008

## [OUTCOME]

Purge возвращает ошибку наружу (не fail-open); admin handler не отдаёт err.Error() в 500 body.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/audit/**
  - backend/testframework/audit/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - GDPR eraser TASK-079 без нужды
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W6-A4-001, A2W6-A5-002, A2W6-A4-007
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - purge error propagates
  - тест fail-open перевёрнут
  - fixed user messages + slog
  - Мутация: swallow purge err → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W6-A4-001 | `backend/internal/audit/service.go:368-394` — в `PurgeExpired` снова `return nil` после slog (не `errors.Join(errs...)`) | `TestAuditService_PurgeExpired_PropagatesError` | service |
| A2W6-A5-002 | тот же guard `service.go:368-394` (бывший cement fail-open тест перевёрнут) — swallow purge err | `TestAuditService_PurgeExpired_PropagatesError` | service |
| A2W6-A4-007 | `backend/internal/audit/handler.go:137-140` — снова `fiber.NewError(500, err.Error())` (и/или Restore `:263-265`) | `TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007` (+ `TestHandler_RestoreItem_Fixed500Message_A2W6A4007`) | handler |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A4-001 / A5-002 | `purgeMockRepo` → `AuditService.PurgeExpired` | Guard в **service** (Join), не Ent purge_* — mock repo **OK** |
| A4-007 | Fiber handler + mock repo leak | Guard в **handler** fixed RU message — mock **OK** |

`layer_mismatch IDs: (none)`

### Gaps

- Purge fail-closed не доказывает Ent `PurgeExpired*` SQL; только service Join path.
- A4-007: покрыты List + Restore; остальные admin 500 paths — residual coverage (мутировать List/Restore как канон).

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
