# TASK-114 — 🔒 Audit/GDPR: limits/layering/purge tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**
> Источник: A2W6 audit residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W11-be2.md` · 12/14 PROVEN · 2 N/A (A4-004/A4-008)

## [OUTCOME]

W6 audit/GDPR residual вне TASK-090: cascade, numeric limits fail-closed, layering, mutation tests.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/audit/**
  - testframework/audit/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - TASK-079 без нужды
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W6-A4-002..006/008, A2W6-A5-001/003..009 (14)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 14 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

> Сверено по HEAD 2026-07-30 · be2 W7 slot 2 · IDs: A2W6-A4-002..006/008 + A2W6-A5-001/003..009 (14) · статус **На проверке**

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W6-A4-002 | `service.go:204-206` — убрать `n <= 0` / Atoi fail → `ErrInvalidResourceLimitValue`; `CheckLimit` `:230` / `:235-236` — empty unknown / bad Atoi → **allow** (`return true`) | `TestAuditService_UpdateResourceLimit_RejectsNonPositive_A2W6A4002` (+ `CheckLimit_InvalidValueDenies` / `UnknownEmptyDenies`) | service |
| A2W6-A4-003 | `service.go:530` / `repository.go:135` — cascade **без** `WithTx` (последовательные repo calls); mid-fail HardDelete оставляет prior deletes | `TestDeleteUserGDPR_CascadeRollback_A2W6A4003` (Ent) | service + Ent WithTx |
| A2W6-A4-004 | `retention.go:40-45` — вернуть поле `*ent.Client` в `RetentionService` (вместо `RetentionStore`) | N/A runtime RED (нет layering AST-теста); evidence = struct DI | service layering |
| A2W6-A4-005 | `middleware.go:78-83` — убрать `defer recover` из recorder goroutine | `TestHTTPAuditMiddleware_RecorderPanicRecovered_A2W6A4005` | middleware |
| A2W6-A4-006 | `handler.go:430-432` — мапить `ErrCannotDeleteFounder` → 500 / не Forbidden | `TestHandler_DeleteUserGDPR_FounderForbidden_A2W6A4006` | handler |
| A2W6-A4-008 | N/A runtime — godoc `ErrCannotDeleteFounder` / `DeleteUserGDPR` (`service.go:20-22`, `:513-514`) снова «self-erase разрешён» | N/A code test; evidence = godoc sync | docs/godoc |
| A2W6-A5-001 | `service.go:384-392` — убрать ветки `purge("crm_*")` / `hubbases` из `PurgeExpired` | `TestAuditService_PurgeExpired_CapturesCutoffs_A2W6A5001` (mock callCounts) | service |
| A2W6-A5-003 | `repository.go:332` / `:425` / `:363` / `:394` — убрать `DeletedAtGT(cutoff)` из ListDeleted* (issue/project/tc/file) | `TestRepository_Issue_RetentionHidesOlderThan30Days` (+ Project/TestCase/File) | repository (Ent) |
| A2W6-A5-004 | `service.go:353-357` / SoftDelete+Restore switch — убрать `case "crm_contact"|"crm_deal"|"hubbase"` | `TestAuditService_SoftDelete_RecordsAllTypes_A2W6A5004` (+ Restore) | service (mock call map) |
| A2W6-A5-005 | `repository.go` PurgeExpired* — soft-only / не HardDelete (строка остаётся) | `TestRepository_PurgeExpiredTestCases` / `PurgeExpiredFiles` / `PurgeExpiredProjects` (`ent.IsNotFound`) | repository (Ent) |
| A2W6-A5-006 | `handler` ListTrash — игнор `project_id` query (не парсить / не передавать в repo) | `TestHandler_ListTrash_CapturesProjectID_A2W6A5006` | handler |
| A2W6-A5-007 | `repository.go:452` / `:462` — Restore без `ClearDeletedAt` | `TestRestoreIssue_ClearsDeletedAt_W6A5006` / `TestRestoreFile_ClearsDeletedAt_W6A5006` (+ SoftDeleteAndRestore*) | repository (Ent) |
| A2W6-A5-008 | `service.go:497-502` — `ReassignTasks` return `0, nil` без `ReassignAssignedIssues` | `TestAuditService_ReassignTasks` | service (mock counters) |
| A2W6-A5-009 | `service.go` DeleteUserGDPR success path — skip cascade steps / HardDelete | `TestAuditService_DeleteUserGDPR_SuccessCascadeCalls_A2W6A5009` | service (mock cascade) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A4-002 / A5-001/004/008/009 | mock / service | Guard в service — mock OK |
| A4-003 | Ent + fail HardDelete wrap | **Ent OK** (Mongo-in-tx residual → A3W6-4-001, вне TASK-114) |
| A4-005/006 | Fiber/handler | OK |
| A5-003/005/007 | EntRepository | **Ent OK** |
| A5-006 | handler + mock ListDeleted | OK |
| A4-004 / A4-008 | layering / godoc | нет runtime RED |

`layer_mismatch IDs: (none)` — mock-only IDs бьют service-guards намеренно; Ent-путь A4-003 / A5-003/005/007 покрыт.

### Gaps

- **A2W6-A4-004** — нет теста, краснеющего при возврате `*ent.Client` в `RetentionService` (только code review / compile hygiene).
- **A2W6-A4-008** — godoc-only; нельзя PROVEN runtime.
- **A2W6-A4-003** Mongo half (eraser внутри tx) — residual A3W6-4-001; Ent rollback PROVEN в task114; не fake-PROVEN cross-store.
- TASK-114 **без** Evidence checklist в файле — verifier W8 обязан подтвердить green + mutation отдельно.

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
