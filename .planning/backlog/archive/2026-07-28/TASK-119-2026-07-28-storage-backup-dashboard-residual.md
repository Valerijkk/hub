# TASK-119 — 🔒 Storage/backup/dashboard residual hygiene

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-30-W11-be1  

> Источник: W7 storage residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Sanitize backup errors; dashboard fail-visible; handler 404; argv secrets; layering — W7-A4 residual.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{storage,backup,dashboard}/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - NF-09/MaintenanceGate ADR (099/100)
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A4-007..009/011..013/017/018 (8)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 8 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record

## [MUTATION CRITERIA] — verifier-ready

> W7 slot 3 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| W7-A4-007 | `backup/sanitize.go:13-33` / `service.go:462-465` — `markFailed` пишет `err.Error()` вместо `sanitizeBackupErrorMessage(err)` | `TestMarkFailed_SanitizesErrorMessage_W7A4007` | service |
| W7-A4-008 | `dashboard/repository.go:277-291` — вернуть `Project.Query().WithIssues().All()` вместо `GroupBy`+`Aggregate` | `TestListTopProjectsByIssueCount_Aggregate_W7A4008` | repository |
| W7-A4-009 | `dashboard/service.go:68-71` — после `ListMyOpenTasks` err снова `slog` + continue / `return data, nil` | `TestGetPersonalDashboard_FailVisible_W7A4009` | service |
| W7-A4-011 | `backup/handler.go:102-104` / `:132-134` — убрать `errors.Is(err, ErrNotFound)` → всегда 500 | `TestHandler_RestoreDelete_NotFound404_W7A4011` | handler |
| W7-A4-012 | wiring/test: `adminMw = Next()` вместо `rbac.RequireAdmin` | `TestHandler_RequireAdmin_Denial_W7A4012` | handler+rbac |
| W7-A4-013 | `backup/sanitize.go:38-50` — `--dbname=` с raw URI (password в argv); mongo: `--uri=` вместо `--config=` | `TestPgDumpArgsEnv_NoPasswordInArgv_W7A4013`; `TestMongoConfig_NoPasswordInArgv_W7A4013` | service/sanitize |
| W7-A4-017 | `dashboard/service.go:302+` / `handler.go:114` — `RequireSystemAdmin` → `return nil` (или handler→repo.GetUser) | `TestRequireSystemAdmin_Layering_Task062` | service |
| W7-A4-018 | `backup/repository.go:64-67` / restore path — писать note в `ErrorMessage` вместо `SetRestoredAt` | `TestEntRepository_SetRestoredAtAndDelete` (+ `service_wave100` assert ErrorMessage nil) | repository |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 007 | Ent BackupRecord + export MarkFailed | Sanitize в service — Ent OK |
| 008 | EntRepository aggregate | Guard в repo — Ent OK; **weak:** малый N — WithIssues().All() может дать те же counts |
| 009 | service + closed client | Fail-visible Personal — OK; **хвост** ProjectStats soft-fail → A3W7-4-001 (вне 8 ID) |
| 011 | Fiber handler | ErrNotFound→404 — OK |
| 012 | real `rbac.RequireAdmin` + Ent user roles | Не stub Next — OK |
| 013 | export hooks unit (без exec) | Argv/env contract — OK |
| 017 | service RequireSystemAdmin | Layering — OK |
| 018 | Ent SetRestoredAt | Column vs ErrorMessage — Ent OK |

`layer_mismatch IDs: (none)`

### Gaps

- **W7-A4-008:** semantic counts alone ≠ proof of GroupBy; W8 verifier: confirm RED after WithIssues().All() **или** добавить source-assert на `GroupBy`/`Aggregate`.
- **W7-A4-009 residual:** Personal fail-visible закрыт; `GetProjectStats`/admin counters soft-fail → A3W7-4-001 (не в scope TASK-119).
- Evidence checklist в файле задачи отсутствует (только audit FIXED-OK) — W8 должен сверить hubtest green отдельно.
