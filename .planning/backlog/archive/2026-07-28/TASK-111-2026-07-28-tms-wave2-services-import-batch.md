# TASK-111 — 🐛 TMS W2 residual: tx/N+1/import/bases

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта  
> Источник: A2W2 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Закрыто: 2026-07-28 · be2 · ветка `fix/tms/task-111-w2-residual`  
> Mutation: 2026-07-30 · be1 · W13 slot 1 · `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task111.md` · **12 PROVEN** / 4 N/A / 8 GAP_NOTED

## Evidence (be2 · 2026-07-28)

| ID | Status | Evidence |
|----|--------|----------|
| A2W2-1-002 | FIXED | swagger: zip убран из WebhookAllure description |
| A2W2-1-003 | FIXED | StructValidator test PUT без version → 400; wave124 шлёт version |
| A2W2-1-005 | FIXED | raw body cap 5MB → 413; TestTask111_WebhookRawBodyCap |
| A2W2-2-002 | FIXED | UpdateTestCaseWithSnapshot Ent tx |
| A2W2-2-003 | FIXED | Rollback errors.As AppError (409 не → 500) |
| A2W2-2-004 | FIXED | CreateAutomaticImportRun Ent tx + abort cleanup |
| A2W2-2-005 | FIXED | maxEpochMs в ImportAllureToExistingRun |
| A2W2-2-007 | FIXED | publish EventTestRunCompleted |
| A2W2-2-008 | FIXED | UpdateTestResult(projectID, runID, resultID) |
| A2W2-2-009 | FIXED | mock+Ent Limit=1 pagination |
| A2W2-2-010 | FIXED | buildSnapshot возвращает err |
| A2W2-2-011 | FIXED | фиксированное «некорректный allure JSON» |
| A2W2-3-001 | FIXED | Mongo InsertMany/Find SetLimit cap |
| A2W2-3-002 | FIXED | List*/Batch Limit caps |
| A2W2-3-003 | FIXED | DeleteTestCase(projectID, id) |
| A2W2-4-001 | FIXED | ParseAllure/ParseJUnit depth+len |
| A2W2-4-002 | FIXED | PDF truncate() rune-safe |
| A2W2-4-003 | FIXED | matcher automation_id/case_key !IsNotFound |
| A2W2-4-004 | FIXED | ImportResult.Partial |
| A2W2-4-005 | FIXED | export paginate ListTestCases |
| A2W2-4-006 | FIXED | matcher mock projectID denial |
| A2W2-4-007 | FIXED | TestRail handler fixed 400 msg |
| A2W2-5-001 | FIXED | SoftDeleteBase cascade tx |
| A2W2-5-002 | FIXED | unique sqlite DSN в bases tests |

Mutation: убрать VersionEQ / depth cap / body cap / cascade → краснеют TestTask111_* / importer_limits / bases cascade.

## [MUTATION CRITERIA] — verifier-ready

> Prefer Ent/Mongo denial где guard в repo. Mock-only при Ent-guard = `layer_mismatch`.

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W2-1-002 | N/A runtime — swagger `@Description` без zip (`handler_webhook.go:41-42`) | N/A code test; evidence = godoc/swag sync | docs/swagger |
| A2W2-1-003 | DTO `UpdateTestCaseRequest` — убрать `validate:"required"` с `Version` (или StructValidator bypass) | `TestTask111_UpdateTestCase_RequiresVersion_StructValidator` | handler/DTO |
| A2W2-1-005 | `handler_webhook.go:101-105` — убрать raw-body `len > maxAllureFileBytes` → 413 | `TestTask111_WebhookRawBodyCap` | handler |
| A2W2-2-002 | `ent_repository_tx.go:40-75` — разнести CreateVersion+Update вне одной tx / убрать `VersionEQ` | `TestTask111_UpdateTestCase_SnapshotTx` (weak happy) · Prefer OL: `TestUpdateTestCase_OptimisticLock_A3_004` | Ent tx |
| A2W2-2-003 | `service_test_cases.go:289-295` — убрать `errors.As` AppError → всегда `NewInternal` (409→500) | `TestUpdateTestCase_OptimisticLock_A3_004` / Rollback 409 path | service |
| A2W2-2-004 | `ent_repository_tx.go:156-173` — CreateAutomaticImportRun без tx / early Commit до results | **GAP** — нет TASK-111 inject-fail IT; ImportAllure happy path не PROVEN | Ent tx |
| A2W2-2-005 | `service_import.go:196-199` — убрать `maxEpochMs` guard | `TestTask111_ImportAllureToExisting_EpochGuard` | service |
| A2W2-2-007 | `events.go:10-11` / `service_runs.go` PublishJSON — сменить subject ≠ `tms.testrun.completed` | `TestTask111_EventSubjectCanon` (**weak**: только const) | events |
| A2W2-2-008 | `service_runs.go:169-185` — `UpdateTestResult` без run/project scope | Prefer Ent: `TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008` · mock alone ≠ PROVEN | service + Ent |
| A2W2-2-009 | `ent_repository` ListTestCases — игнор `Limit` / `.All()` | Prefer: `TestTask111_ListTestCases_EntPagination` · mock `service_cases_extra` Len==1 — residual | Ent (+ mock residual) |
| A2W2-2-010 | `buildSnapshot` — ignore err / return nil map | bundled в SnapshotTx / corrupt snapshot paths | service |
| A2W2-2-011 | `service_import.go:41-42` — снова leak internals / старый текст err | `TestTask111_ImportAllure_FixedErrorMessage` | service |
| A2W2-3-001 | `mongo_repository.go:74-80` — убрать `maxMongoAllureResults` cap / SetLimit | **GAP** в TASK-111 file; cap-тесты TASK-140 / mongo suite | Mongo |
| A2W2-3-002 | `ent_repository.go:51-56` + `.Limit(maxList*)` — снять Limit | Prefer TASK-140 `TestTask140_*_Cap` · TASK-111 не изолирует | Ent |
| A2W2-3-003 | `ent_repository.go:369+` — `DeleteTestCase` без `projectID` Where | `TestTask111_DeleteTestCase_ProjectScoped` | Ent |
| A2W2-4-001 | `allure_parser.go:22-23` / `junit_parser.go` — убрать `assertJSON/XMLMaxDepth` | `TestTask111_ParseAllure_DepthCap` (+ `ParseJUnit_DepthCap`) | parser |
| A2W2-4-002 | `exporter.go:288-293` — truncate по bytes `s[:maxLen]` не runes | `TestTask111_ExportTruncateHelper` (+ PDF) | exporter |
| A2W2-4-003 | `allure_matcher.go` — трактовать любой err как NotFound (skip `!IsNotFound`) | `TestTask111_MatcherAutomationID_DBError` | matcher |
| A2W2-4-004 | `testrail_import.go` / `importer.go` — не ставить `Partial=true` при Errors | Prefer TASK-141 Partial tests · TASK-111 gap | service/import |
| A2W2-4-005 | export ListTestCases — снова unbounded без paginate | **GAP** — нет TestTask111 export-paginate | exporter/service |
| A2W2-4-006 | mock Find* без projectID check **или** Ent `FindTestCaseByAutomationID` без `ProjectID` | `TestTask111_MatcherCrossProjectDenial` (**mock**) | **layer_mismatch** mock≠Ent |
| A2W2-4-007 | `testrail_import.go:432-433` — снова `err.Error()` в 400 body | **GAP** — нет TestTask111 fixed-msg handler IT | handler |
| A2W2-5-001 | `bases/repository.go:146-200` — SoftDeleteBase без cascade records/columns/tables | `TestTask111_SoftDeleteBase_Cascade` | Ent bases |
| A2W2-5-002 | N/A prod — unique sqlite DSN в tests (`uuid` в `file:…`) | N/A mutation prod; hygiene evidence only | test hygiene |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 1-003 / 1-005 / 2-011 / 4-001 / 4-002 / 4-003 | handler/parser/service/matcher | Guard на том же слое — OK |
| 2-002 / 2-009 / 3-003 / 5-001 | Ent | Ent OK |
| 2-008 | Ent A3_008 prefer | mock UpdateTestResult scope ≠ prod HasRunWith |
| 4-006 | mock matcher repo | Prod guard = Ent `ProjectID` Where — **layer_mismatch** |
| 3-001 / 3-002 / 4-004 / 4-005 / 4-007 / 2-004 | partial/GAP | см. Gaps |

`layer_mismatch IDs: A2W2-4-006` (+ residual mock path A2W2-2-009 если крутить только mock)

### Gaps

- Нет runnable PROVEN для: **1-002** (docs), **2-004** (нет mid-fail IT), **2-007** (const-only), **3-001/002** (вынесены в TASK-140), **4-004/005/007**, **5-002** (test hygiene).
- SnapshotTx happy-path **не** ловит early-Commit mutation без inject hook.
- **2-005:** named test не ловит maxEpochMs — mock `BatchCreateTestResults` игнорирует `DurationMs` (`GAP_NOTED`).
- Verifier W8/W13: PROVEN только ID с runnable RED; остальное — GAP / N/A, не fake-PROVEN.

### Mutation W13 be1 (2026-07-30)

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task111.md`  
**12 PROVEN** · 4 N/A · 8 GAP_NOTED · leftover MUTATED: нет · статус **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| A2W2-1-002 | N/A | — |
| A2W2-1-003 | PROVEN | TestTask111_UpdateTestCase_RequiresVersion_StructValidator |
| A2W2-1-005 | PROVEN | TestTask111_WebhookRawBodyCap |
| A2W2-2-002 | PROVEN | TestEntRepository_UpdateTestCase_OptimisticLock_A3_004 |
| A2W2-2-003 | PROVEN | TestUpdateTestCase_OptimisticLock_A3_004 |
| A2W2-2-004 | GAP_NOTED | — |
| A2W2-2-005 | GAP_NOTED | TestTask111_ImportAllureToExisting_EpochGuard (no catch) |
| A2W2-2-007 | N/A | TestTask111_EventSubjectCanon (const-only) |
| A2W2-2-008 | PROVEN | TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008 |
| A2W2-2-009 | PROVEN | TestTask111_ListTestCases_EntPagination |
| A2W2-2-010 | GAP_NOTED | TestTask111_UpdateTestCase_SnapshotTx (no catch) |
| A2W2-2-011 | PROVEN | TestTask111_ImportAllure_FixedErrorMessage |
| A2W2-3-001 | GAP_NOTED | — |
| A2W2-3-002 | GAP_NOTED | — |
| A2W2-3-003 | PROVEN | TestTask111_DeleteTestCase_ProjectScoped |
| A2W2-4-001 | PROVEN | TestTask111_ParseAllure_DepthCap |
| A2W2-4-002 | PROVEN | TestTask111_ExportTruncateHelper |
| A2W2-4-003 | PROVEN | TestTask111_MatcherAutomationID_DBError |
| A2W2-4-004 | GAP_NOTED | — |
| A2W2-4-005 | GAP_NOTED | — |
| A2W2-4-006 | N/A | layer_mismatch |
| A2W2-4-007 | GAP_NOTED | — |
| A2W2-5-001 | PROVEN | TestTask111_SoftDeleteBase_Cascade |
| A2W2-5-002 | N/A | test hygiene |

## [OUTCOME]

Все open W2 вне TASK-091 закрыты (tx, N+1, parser caps, import limits, bases).

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tms/**
  - backend/internal/bases/**
  - testframework/tms bases
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать TASK-015/016 closures
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W2-1-002/003/005, A2W2-2-002..005/007..011, A2W2-3-*, A2W2-4-*, A2W2-5-* (24)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - чеклист 24 ID в отчёте
  - hubtest green
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
