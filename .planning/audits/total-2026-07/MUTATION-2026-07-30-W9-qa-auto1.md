# MUTATION W9 slot 3 — TASK-140 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** (!= be2 fixer) · Env: localhost (hubtest Ent sqlite; POSTGRES_TEST_URL set)  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-140-2026-07-29-a3-2-3-tms-repos.md` `[MUTATION CRITERIA]`  
> Runner: `_run_mutations_w9_task140.py` · Evidence JSON: `_mut-batch-w9-task140.json`

## Verdict

**3/3 finding IDs PROVEN** (9 path-level mutations) · статус TASK-140 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W2-3-001a-byrunids | **PROVEN** | `TestTask140_ListTestResultsByRunIDs_Cap` | убрать `.Limit(maxListTestResults)` с `ListTestResultsByRunIDs` | EntRepository |
| A3W2-3-001a-withresults | **PROVEN** | `TestTask140_GetTestRunWithResults_Cap` | убрать `Limit` внутри `WithResults` (оба getter) | EntRepository |
| A3W2-3-001b-add | **PROVEN** | `TestTask140_AddTestCasesToRun_BatchErr` | убрать batch check `AddTestCasesToRun` | EntRepository |
| A3W2-3-001b-last | **PROVEN** | `TestTask140_GetLastResults_BatchErr` | убрать input cap `GetLastResultsForTestCases` | EntRepository |
| A3W2-3-001c | **PROVEN** | `TestTask140_GetLastResultsForTestCases_PerCaseNoFullHistory` | `.All` Asc history (oldest) вместо per-case `First` | EntRepository |
| A3W2-3-001d-env | **PROVEN** | `TestTask140_ListEnvironments_Cap` | убрать `Limit(maxListEnvironments)` | EntRepository |
| A3W2-3-001d-issues | **PROVEN** | `TestTask140_GetProjectIssuesWithTestCases_Cap` | убрать `Limit(maxListProjectIssues)` | EntRepository |
| A3W2-3-002 | **PROVEN** | `TestTask140_CountLabelsInProject_SkipsSoftDeleted` | убрать `label.DeletedAtIsNil()` | EntRepository |
| A3W2-3-003 | **PROVEN** | `TestTask140_LinkedEdges_SkipSoftDeleted` | убрать edge `DeletedAtIsNil` с `WithLinked*` | EntRepository |

Production restored (`git checkout -- backend/internal/tms/ent_repository.go` после каждого BREAK). Mock-gap: нет (все TestTask140_* → Ent sqlite).

## Per-ID evidence

### A3W2-3-001 — list/batch/nested caps

```text
BASELINE TestTask140_ → PASS (~104.5s cold / ~13s warm)

BREAK 001a-byrunids: ListTestResultsByRunIDs без Limit
→ FAIL TestTask140_ListTestResultsByRunIDs_Cap  (len==3 при cap=2)
RESTORE → PASS

BREAK 001a-withresults: WithResults без Limit (оба getter)
→ FAIL TestTask140_GetTestRunWithResults_Cap
RESTORE → PASS

BREAK 001b-add: AddTestCasesToRun без batch check
→ FAIL TestTask140_AddTestCasesToRun_BatchErr  (нет «слишком много»)
RESTORE → PASS

BREAK 001b-last: GetLastResults input без batch check
→ FAIL TestTask140_GetLastResults_BatchErr
RESTORE → PASS

BREAK 001c: GetLastResultsForTestCases → All Asc (oldest) вместо First
→ FAIL TestTask140_GetLastResultsForTestCases_PerCaseNoFullHistory  (status≠failed)
RESTORE → PASS

BREAK 001d-env: ListEnvironmentsByProject без Limit
→ FAIL TestTask140_ListEnvironments_Cap
RESTORE → PASS

BREAK 001d-issues: GetProjectIssuesWithTestCases без Limit
→ FAIL TestTask140_GetProjectIssuesWithTestCases_Cap
RESTORE → PASS
```

### A3W2-3-002 — CountLabels soft-delete

```text
BREAK: CountLabelsInProject без label.DeletedAtIsNil()
→ FAIL TestTask140_CountLabelsInProject_SkipsSoftDeleted  (Count=2 вместо 1)
RESTORE → PASS
```

### A3W2-3-003 — linked edges soft-delete

```text
BREAK: WithLinkedIssues / WithLinkedTestCases без DeletedAtIsNil
→ FAIL TestTask140_LinkedEdges_SkipSoftDeleted  (trash в ответе)
RESTORE → PASS
```

## Commands

```text
# baseline + per-path Break→RED→GREEN
python .planning/audits/total-2026-07/_run_mutations_w9_task140.py
→ BASELINE OK · 9/9 PROVEN · final TestTask140_ OK (~13.1s)
→ wall ~523s

# regression after restore
cd backend
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/tms/ -run TestTask140_
→ ok
```

## Assumptions / unresolved

- TestTask140_* бьют Ent sqlite (`newEntRepositoryTestEnv`), не live Postgres twin — layer = repository Ent OK; `POSTGRES_TEST_URL` выставлен, но suite не требует PG.
- 001c: мутация `First`→`All` Asc + first-seen (oldest) — семантика «latest» ломается → RED. Чистый `All`+Desc+first-seen остался бы GREEN (семантика latest сохраняется при полной загрузке) — критерий ловит regression через wrong latest, не через memory assert.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-140 **Закрыта** · evidence этот файл · README синхронизирован · mutation gate закрыт · next: qa-auto-ver / wave control.
