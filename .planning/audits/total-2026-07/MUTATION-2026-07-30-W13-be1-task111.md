# MUTATION W13 slot 1 — TASK-111 (be1) · 2026-07-30

> Checker: **be1** · Env: host hubtest + Docker `deployments_default` for Ent/Postgres  
> Criteria: `.planning/backlog/archive/2026-07-28/TASK-111-2026-07-28-tms-wave2-services-import-batch.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**12 PROVEN** · **4 N/A** · **8 GAP_NOTED** · leftover MUTATED: **нет**  
TASK-111 → **Закрыта** (partial: runnable criteria proven; gaps documented, not fake-PROVEN)

| ID | Verdict | RED test |
|----|---------|----------|
| A2W2-1-002 | **N/A** | docs/swagger `@Description` — нет runtime guard |
| A2W2-1-003 | **PROVEN** | `TestTask111_UpdateTestCase_RequiresVersion_StructValidator` |
| A2W2-1-005 | **PROVEN** | `TestTask111_WebhookRawBodyCap` |
| A2W2-2-002 | **PROVEN** | `TestEntRepository_UpdateTestCase_OptimisticLock_A3_004` (Docker/Postgres) |
| A2W2-2-003 | **PROVEN** | `TestUpdateTestCase_OptimisticLock_A3_004` |
| A2W2-2-004 | **GAP_NOTED** | нет TASK-111 mid-fail IT на `CreateAutomaticImportRun` |
| A2W2-2-005 | **GAP_NOTED** | `TestTask111_ImportAllureToExisting_EpochGuard` — mock `BatchCreateTestResults` игнорирует `DurationMs` → mutation maxEpochMs не краснеет |
| A2W2-2-007 | **N/A** | `TestTask111_EventSubjectCanon` — const-only (weak); runtime PublishJSON не доказан |
| A2W2-2-008 | **PROVEN** | `TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008` (Docker/Postgres) |
| A2W2-2-009 | **PROVEN** | `TestTask111_ListTestCases_EntPagination` |
| A2W2-2-010 | **GAP_NOTED** | `TestTask111_UpdateTestCase_SnapshotTx` happy — ignore `buildSnapshot` err не ловится |
| A2W2-2-011 | **PROVEN** | `TestTask111_ImportAllure_FixedErrorMessage` |
| A2W2-3-001 | **GAP_NOTED** | cap-тесты в TASK-140 / mongo suite |
| A2W2-3-002 | **GAP_NOTED** | Prefer TASK-140 `TestTask140_*_Cap` |
| A2W2-3-003 | **PROVEN** | `TestTask111_DeleteTestCase_ProjectScoped` |
| A2W2-4-001 | **PROVEN** | `TestTask111_ParseAllure_DepthCap` (JUnit twin: test_does_not_catch — Unmarshal всё равно Error) |
| A2W2-4-002 | **PROVEN** | `TestTask111_ExportTruncateHelper` |
| A2W2-4-003 | **PROVEN** | `TestTask111_MatcherAutomationID_DBError` |
| A2W2-4-004 | **GAP_NOTED** | Prefer TASK-141 Partial tests |
| A2W2-4-005 | **GAP_NOTED** | нет TestTask111 export-paginate |
| A2W2-4-006 | **N/A** | **layer_mismatch** mock matcher ≠ Ent `ProjectID` Where |
| A2W2-4-007 | **GAP_NOTED** | нет TestTask111 fixed-msg handler IT |
| A2W2-5-001 | **PROVEN** | `TestTask111_SoftDeleteBase_Cascade` |
| A2W2-5-002 | **N/A** | unique sqlite DSN — test hygiene, не prod guard |

`layer_mismatch IDs: A2W2-4-006`

## Per-ID evidence (PROVEN)

### A2W2-1-003 — Version validate required

```text
BREAK: dto.go UpdateTestCaseRequest — убрать validate:"required" с Version
→ FAIL StatusBadRequest (omit version проходит Bind)
RESTORE → PASS
```

### A2W2-1-005 — webhook raw body 5MB

```text
BREAK: handler_webhook.go — убрать len(allureData) > maxAllureFileBytes → 413
→ FAIL StatusRequestEntityTooLarge
RESTORE → PASS
```

### A2W2-2-002 — Ent VersionEQ (Prefer OL)

```text
Env: docker run --network deployments_default golang:1.25-bookworm
     POSTGRES_TEST_URL=postgresql://hub:hub@postgres:5432/hub?sslmode=disable
     (host :5432 publish broken on this machine — [::1]/127.0.0.1 refused)

BREAK: ent_repository.go UpdateTestCase — снять testcase.VersionEQ
→ FAIL stale update passes / 409 missing
RESTORE → PASS (~90s/container cold)
```

SnapshotTx happy (`TestTask111_UpdateTestCase_SnapshotTx`) **не** использован как PROVEN для early-Commit (weak).

### A2W2-2-003 — errors.As → 409

```text
BREAK: service_test_cases.go UpdateTestCase — убрать errors.As AppError → всегда NewInternal
→ FAIL expect 409, got Internal
RESTORE → PASS
```

### A2W2-2-008 — UpdateTestResult HasRunWith

```text
BREAK: ent_repository.go UpdateTestResult — снять HasRunWith(runID)
→ FAIL wrong-run update меняет status
RESTORE → PASS (Docker/Postgres)
```

### A2W2-2-009 — ListTestCases Limit

```text
BREAK: ent_repository.go ListTestCases — убрать .Limit(limit)
→ FAIL Len(items)==1 actual=3
RESTORE → PASS
```

### A2W2-2-011 — fixed allure error message

```text
BREAK: service_import.go — вернуть leak "парсинг allure-results: …"
→ FAIL Contains "некорректный allure JSON" / NotContains leak
RESTORE → PASS
```

### A2W2-3-003 — DeleteTestCase project scope

```text
BREAK: ent_repository.go DeleteTestCase — убрать testCaseProjectIDEQ(projectID)
→ FAIL foreign projectID delete succeeds
RESTORE → PASS
```

### A2W2-4-001 — ParseAllure depth

```text
BREAK: allure_parser.go — убрать assertJSONMaxDepth
→ FAIL Contains "глубин"
RESTORE → PASS

JUnit twin (junit_parser.go assertXMLMaxDepth): test_does_not_catch —
ParseJUnitXML всё равно Error («нет testsuite») без depth-сообщения.
```

### A2W2-4-002 — rune-safe truncate

```text
BREAK: exporter.go truncate — s[:maxLen] по bytes
→ FAIL ExportTruncate("абвгдеёжзий", 3) ≠ "абв..."
RESTORE → PASS
```

### A2W2-4-003 — matcher infra ≠ NotFound

```text
BREAK: allure_matcher.go automation_id label — silent-skip любого err
→ FAIL expect db timeout propagated
RESTORE → PASS
```

### A2W2-5-001 — SoftDeleteBase cascade

```text
BREAK: bases/repository.go SoftDeleteBase — только base, без cascade tables/cols/recs
→ FAIL aliveTables/Cols/Recs ≠ 0
RESTORE → PASS
```

## Gaps / N/A detail

| ID | Почему не PROVEN |
|----|------------------|
| 1-002 | swagger/godoc only |
| 2-004 | нет inject-fail IT в TASK-111 |
| 2-005 | named test не бьёт DurationMs (mock BatchCreate drops input fields) |
| 2-007 | const equality ≠ PublishJSON subject proof |
| 2-010 | happy SnapshotTx не провоцирует marshal err |
| 3-001/002 | вынесены в TASK-140 (уже закрыт mutation W9) |
| 4-004/005/007 | нет runnable TASK-111 RED / TASK-141 |
| 4-006 | layer_mismatch |
| 5-002 | test DSN hygiene |

## Commands

```text
# host (sqlite/mock + most TestTask111_*)
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
go test -tags hubtest -count=1 -timeout=180s ./testframework/tms/ -run "^TestTask111_"

# Ent/Postgres OL + result scope (host publish broken → compose network)
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -v "$HOME/go/pkg/mod:/go/pkg/mod" -w /app \
  -e GOPROXY=off -e GOSUMDB=off \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  golang:1.25-bookworm \
  go test -tags hubtest -count=1 -timeout=180s ./testframework/tms/ \
  -run "TestEntRepository_UpdateTestCase_OptimisticLock_A3_004|TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008"

# drivers
# .planning/audits/total-2026-07/_run_mutations_w13_task111.py
# .planning/audits/total-2026-07/_run_mutations_w13_task111_ent.py
```

## Leftover / side-note

- Mutation restores via file snapshots (не `git checkout` blanket) — concurrent TASK-139 `afterExistingRunBatchHook` WIP не затирался после инцидента первого прогона.
- Prod files без маркеров `MUTATED`. Diff `ent_repository_tx.go` / `export_hooks_repo.go` = TASK-139 inject hook (не leftover этой волны).

## Handoff

- Evidence: этот файл · TASK-111 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Коммит: не создавался
