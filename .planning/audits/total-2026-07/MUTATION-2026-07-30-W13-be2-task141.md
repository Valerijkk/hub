# MUTATION W13 slot 3 — TASK-141 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-141-2026-07-29-a3-2-4-tms-import.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task141.py`  
> Коммит не создавался.

## Verdict

**5/5 PROVEN** · TASK-141 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W2-4-001 | **PROVEN** | `TestTask141_ExportTestCases_StepsFilled` |
| A3W2-4-002 | **PROVEN** | `TestTask141_ParseTestRailXML_DepthCap` · `TestTask141_ParseTestRailCSV_RecordsCap` |
| A3W2-4-003 | **PROVEN** | `TestTask141_HandleImport_FixedParseError` |
| A3W2-4-004 | **PROVEN** | `TestTask141_ExportTestCasesToCSV_FlushError` |
| A3W2-4-005 | **PROVEN** | `TestTask141_ImportTestRail_CreateSuiteFail_Partial` |

`layer_mismatch IDs: (none)` — guards в service/parser/exporter; mock≠Ent намеренно.

## Per-ID evidence

### A3W2-4-001 — export Steps fill

```text
BASELINE → PASS TestTask141_ExportTestCases_StepsFilled (~6.4s)
BREAK: importer.go handleExportTestCases — Steps: "" (без formatTestStepsForExport)
→ FAIL (Steps empty)
RESTORE → PASS (~8.4s)
```

### A3W2-4-002 — TestRail depth/records caps

```text
BASELINE XML → PASS ParseTestRailXML_DepthCap
BREAK: убрать assertXMLMaxDepth в parseTestRailXML
→ FAIL
RESTORE → PASS

BASELINE CSV → PASS ParseTestRailCSV_RecordsCap
BREAK: убрать maxImportRecords check в parseTestRailCSV
→ FAIL
RESTORE → PASS
```

### A3W2-4-003 — fixed parse error (no leak)

```text
BASELINE → PASS HandleImport_FixedParseError (~6.2s)
BREAK: return NewBadRequest("ошибка парсинга файла: " + err.Error())
→ FAIL (~1.1s)
RESTORE → PASS (~6.6s)
```

### A3W2-4-004 — CSV Flush w.Error()

```text
BASELINE → PASS ExportTestCasesToCSV_FlushError (~6.1s)
BREAK: ExportTestCasesToCSV — Flush без w.Error()
→ FAIL (~8.5s)
RESTORE → PASS (~6.7s)
```

### A3W2-4-005 — CreateSuite fail Partial/Errors

```text
BASELINE → PASS ImportTestRail_CreateSuiteFail_Partial (~6.3s)
BREAK: CreateSuite err → continue без Errors++/ValidationErrors
→ FAIL (~8.4s)
RESTORE → PASS (~6.9s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=300s ./testframework/tms/ -run "TestTask141_"
# BASELINE PASS ~6.1s · FINAL PASS ~6.5s

# driver
python .planning/audits/total-2026-07/_run_mutations_w13_task141.py
# raw: _mut-batch-w13-task141.json
```

## Gaps

- none runnable — all 5 criteria PROVEN  
- 001 не бьёт HTTP Export handler напрямую — service path (criteria note) OK

## Handoff

- Evidence: этот файл · TASK-141 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на `importer.go` / `testrail_import.go` / `exporter.go` — empty
