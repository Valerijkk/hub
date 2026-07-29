# 🐛 C-W19: Allure ExistingRun silent partial

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`C-W19-tracker-tms.md`](../audits/total-2026-07/C-W19-tracker-tms.md) §C-W19-TT-002  
> Исполнитель: **be2** · Finding ID: `C-W19-TT-002`

## [OUTCOME]

Allure import в ExistingRun не молчит при Mongo store fail / non-NotFound match err: либо fail-closed, либо явный `Partial`/Errors в ответе (зеркало TestRail/CSV).

## [ROLE]

- Role: be2 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/tms/service_import.go` ExistingRun Allure path; DTO Partial/Errors; `testframework/tms/**`
- Do not touch: откат Ent tx ImportResultsToExistingRun (TASK-139); TestRail Partial (TASK-141) — только зеркалировать контракт
- Commits EN/RU; no AI co-author

## [REFERENCES]

- `StoreAllureResults` err → slog only; `buildAllureResultInputs` `continue` on matchErr; response без `Partial`
- A2W2-4-004 / TASK-141 покрыли CSV/TestRail Partial, не Allure ExistingRun
- Audit: C-W19-tracker-tms.md

## [ACCEPTANCE]

- Выбрать path: (A) fail-closed на Store/match non-NotFound **или** (B) `Partial=true` + Errors/Skipped в ответе; IT на оба сценария Mongo fail / match DB err
- Mutation: вернуть slog-only + silent continue → тест RED
- Stop: FIXED · «На проверке»

## [EXECUTION — be2 2026-07-29]

- **Path:** A (fail-closed). Match → Mongo → PG; Store/match infra err → 500; PG не пишется.
- **Код:** `service_import.go` — `buildAllureResultInputs` возвращает `matchErr`; ExistingRun не slog-only на Store.
- **Тесты:** `allure_existingrun_partial_c_w19_test.go` (Mongo + match DB).
- **Mutation:** slog-only + continue → оба TestTask185_* RED; restore → green.
- **Docs:** `docs/wiki/API.md` ExistingRun fail-closed note.
- TASK-008 не закрывали.
