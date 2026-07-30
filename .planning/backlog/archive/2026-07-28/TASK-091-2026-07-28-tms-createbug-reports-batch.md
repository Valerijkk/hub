# TASK-091 — 🐛 TMS: createBug runId binding + reports batch

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта** · MUTATION-2026-07-30-W4-be1  
> Источник: A2W2 · Stage A2 dedup · TASK-008  
> Исполнитель: qa-auto2 (rework Ent IT) · Verifier: be1 (W4 slot 1) · Handoff: be-ver

> **ВОЗВРАЩЕНА qa-auto1 2026-07-30 (STAGE-C W2 slot 4) · `test_does_not_catch` + `layer_mismatch`.**
>
> 1. **Guard в EntRepository:** `GetTestResultInRun` —
>    `backend/internal/tms/ent_repository.go:1231-1243` —
>    `testresult.HasRunWith(testrun.IDEQ(runID))` (A2W2-1-001/004).
> 2. **Тест только mock:** `TestCreateBugFromFailedResult_ForeignRunDenied_TASK091`
>    (`backend/testframework/tms/task091_createbug_reports_test.go`) —
>    `newMockRepo()`; scoping в `mockRepository.GetTestResultInRun` —
>    копия логики, не прод-Ent.
> 3. **Доказательство (до rework):** сняли `HasRunWith` из EntRepository →
>    mock-тест **остался PASS**. Prod tree restored.
> 4. W4 mutation била `service_test_cases.go` (GetTestResult vs InRun) —
>    ловит wiring сервиса, **не** SQL-scope Ent. Эталон: TASK-137.
>
> **Rework (qa-auto2 2026-07-30):** Ent/Postgres IT
> `TestEntRepository_GetTestResultInRun_RejectsForeignRun_TASK091` —
> result ∈ run A, CreateBug/GetTestResultInRun с run B того же project →
> NotFound / «не найден»; mutation: убрать `HasRunWith` → RED (mock PASS).
> Batch-тесты A2W2-2-001 оставлены. Mock wiring-тест сохранён.

## [OUTCOME]

createBug берёт runId из URL и сверяет result∈run; report endpoints без N+1 last-result на каждый case.

## [ROLE]

- Role: qa-auto2 (Ent IT rework) / be2 (prod guard already present)
- Responsibility: regression/mutation-ready тест **на слое EntRepository**
- Handoff to: qa-auto-ver → be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/tms/handler_test_cases.go`
  - `backend/internal/tms/service_test_cases.go`
  - `backend/internal/tms/service_reports.go`
  - `backend/internal/tms/ent_repository.go` (HasRunWith scope — не ослаблять)
  - `backend/testframework/tms/**`
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - TMS IDOR TASK-015 (уже закрыт)
  - ослаблять `HasRunWith` в других query без отдельной задачи
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W2-1-001/004 · A2W2-2-001
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Эталон reopen: TASK-137 · STAGE-C W2 slot 4
- Prior false-close: MUTATION-2026-07-29-W4 (service-layer only)
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - URL :runId + ownership check **доказан Ent IT** (не только mock)
  - batch IDIn / join для reports (уже Ent — сохранить green)
  - тесты + мутации: снятие `HasRunWith` → Ent-тест RED
- Gate перед «На проверке»: нет layer_mismatch mock≠Ent для A2W2-1-001/004
- Regression: `go test -tags hubtest -race -count=1 ./testframework/tms/`
- Stop when: Ent IT + mutation prove · статус «На проверке» в README

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A2W2-1-001/004 | убрать `HasRunWith` из `EntRepository.GetTestResultInRun` | Ent IT foreign-run → RED (баг создаётся / result найден) |
| A2W2-2-001 | вернуть N+1 `GetLastResultForTestCase` в coverage | batch counter / Ent batch тест RED |

## [MUTATION CRITERIA] — verifier-ready

> PROVEN 2026-07-30 qa-auto2 · независимо подтверждено be1 2026-07-30 (W4) · слой = реальный Ent guard на Postgres

| ID | что ломать | какой тест | слой |
|----|------------|------------|------|
| A2W2-1-001/004 | `backend/internal/tms/ent_repository.go` `GetTestResultInRun` — убрать `testresult.HasRunWith(testrun.IDEQ(runID))` из Where | `TestEntRepository_GetTestResultInRun_RejectsForeignRun_TASK091` | repository |
| A2W2-2-001a | вернуть цикл `GetLastResultForTestCase` в `GetCoverageReport` (убрать batch) | `TestCoverageReport_UsesBatchLastResults_TASK091` | service |
| A2W2-2-001b | сломать batch `GetLastResultsForTestCases` / `ListTestResultsByRunIDs` (N+1 или drop IDIn) | `TestGetLastResultsForTestCases_Batch_TASK091` / `TestListTestResultsByRunIDs_Batch_TASK091` | repository |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A2W2-1-001/004 | **EntRepository + Postgres** | Guard в `GetTestResultInRun`; mock wiring оставлен, mutation на mock не ловит |
| A2W2-2-001a | mockRepo + service | batch counter — mock OK |
| A2W2-2-001b | EntRepository (sqlite env) | batch IDIn — Ent OK |

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note (Ent HasRunWith)
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
- Не закрывать снова только mock-mutation service→GetTestResult

## [EVIDENCE] be1 2026-07-30 (W4 slot 1 · independent verify)

### Verdict: **PROVEN** · статус **Закрыта** · MUTATION-2026-07-30-W4-be1

Независимая мутация (rewriter qa-auto2; verifier = be1 ≠ qa-auto2).

| Шаг | Результат |
|-----|-----------|
| BREAK: убрать `testresult.HasRunWith(testrun.IDEQ(runID))` из `GetTestResultInRun` | **FAIL** `TestEntRepository_GetTestResultInRun_RejectsForeignRun_TASK091` — `foreign-run result must be NotFound, got <nil>` (0.58s) |
| RESTORE `HasRunWith` | **PASS** тот же тест (0.58s) · `ent_repository.go` diff clean |

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"

# MUTATED
go test -tags hubtest -count=1 -timeout=180s ./testframework/tms/ -run RejectsForeignRun_TASK091 -v
→ FAIL (1.793s)
   foreign-run result must be NotFound, got <nil>

# RESTORED
go test -tags hubtest -count=1 -timeout=180s ./testframework/tms/ -run RejectsForeignRun_TASK091 -v
→ PASS (1.661s)
ok  github.com/valerii/hub/backend/testframework/tms
```

- Слой: **EntRepository + Postgres** (не mock)
- Prod tree restored; `git diff -- backend/internal/tms/ent_repository.go` empty
- A2W2-2-001 batch IDs не трогались в этом слоте

---

## [EVIDENCE] qa-auto2 2026-07-30 (Ent IT rework)

### Verdict: **FIXED** · A2W2-1-001/004 Ent-mutation PROVEN → статус **На проверке**

### Mutation prove A2W2-1-001/004

```text
BREAK: убрать HasRunWith из EntRepository.GetTestResultInRun
→ FAIL TestEntRepository_GetTestResultInRun_RejectsForeignRun_TASK091
   (foreign-run result must be NotFound, got <nil>)
→ PASS TestCreateBugFromFailedResult_ForeignRunDenied_TASK091  (mock — не ловит)

RESTORE HasRunWith
→ PASS TestEntRepository_GetTestResultInRun_RejectsForeignRun_TASK091
→ PASS ./testframework/tms/ -run TASK091 (5 tests)
```

### Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
go test -tags hubtest -count=1 -timeout=180s ./testframework/tms/ -run TASK091 -v
→ PASS (5 tests)
```

### Paths changed

- `backend/testframework/tms/task091_createbug_reports_test.go` — Ent/Postgres IT
- `.planning/backlog/TASK-091-…` + `README.md` — MUTATION CRITERIA + статус
- `ent_repository.go` — только temporary mutation (восстановлен, diff clean)
