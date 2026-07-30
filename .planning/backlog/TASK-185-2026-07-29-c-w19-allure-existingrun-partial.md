# 🐛 C-W19: Allure ExistingRun silent partial

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`C-W19-tracker-tms.md`](../audits/total-2026-07/C-W19-tracker-tms.md) §C-W19-TT-002  
> Исполнитель: **be2** · Finding ID: `C-W19-TT-002`  
> Mutation: [`MUTATION-2026-07-30-W13-be2-task185.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task185.md) · 2/2 PROVEN (+ GAP_NOTED mock≠Ent)

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

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · path A fail-closed

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| C-W19-TT-002-001 | `backend/internal/tms/service_import.go:142-145` — `ImportAllureToExistingRun`: `StoreAllureResults` err → `NewInternal` (не slog-only + continue) | service | `TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed` · `backend/testframework/tms/allure_existingrun_partial_c_w19_test.go` | вернуть slog-only / игнорировать Store err → 200 + PG write → 500/no-write assert RED |
| C-W19-TT-002-002 | `service_import.go:137-140` + `buildAllureResultInputs:186-188` — matchErr (infra ≠ NotFound) → `return …, matchErr` / ExistingRun → `NewInternal("…сопоставления…")` (не `continue`) | service | `TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed` · `allure_existingrun_partial_c_w19_test.go` | вернуть `continue` на matchErr → silent skip / `err==nil` → 500/сопоставления assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | TMSService + mockRepo `storeAllureErr` (не live Mongo) | **mock≠Ent** (GAP_NOTED): fail-closed на service path; Mongo Store inject через mock |
| 002 | TMSService + mockRepo `findMatchErr` (matcher infra) | **mock≠Ent** (GAP_NOTED): match fail-closed на service; не Ent matcher IT |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: C-W19-TT-002-001, C-W19-TT-002-002` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — оба ID service fail-closed (wiki note не mutation target).
- Create-path Mongo best-effort (`ImportAllureResults` Store) — **не** kill target; только ExistingRun.
- Path B (`Partial=true`) не выбран — kill = fail-closed, не Partial-флаг.

## [MUTATION EXECUTION — be2 2026-07-30 W13 slot 3]

| ID | Verdict | RED |
|----|---------|-----|
| C-W19-TT-002-001 | PROVEN + GAP_NOTED (mock≠Ent) | `TestTask185_ImportAllureExistingRun_MongoStoreFail_FailClosed` |
| C-W19-TT-002-002 | PROVEN + GAP_NOTED (mock≠Ent) | `TestTask185_ImportAllureExistingRun_MatchDBErr_FailClosed` |

- BREAK: Store slog-only → RED; matchErr `continue` → RED; restore → GREEN both.
- leftover MUTATED: **нет** · `git diff service_import.go` empty.
- Статус задачи: **Закрыта** (0 NOT_PROVEN).
