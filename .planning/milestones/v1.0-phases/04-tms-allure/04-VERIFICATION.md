---
phase: 04-tms-allure
verified: 2026-03-21T00:00:00Z
status: gaps_found
score: 3/5 must-haves verified
gaps:
  - truth: "Backend compiles and all TMS routes are served by the running server"
    status: failed
    reason: "CombinedRepository does not implement the Repository interface — 5 methods added in plan 04-04 to repository.go and ent_repository.go were never forwarded in combined_repository.go. go build ./... fails with: *tms.CombinedRepository does not implement tms.Repository (missing method CreateBugIssue)"
    artifacts:
      - path: "backend/internal/tms/combined_repository.go"
        issue: "Missing 5 delegation methods: CreateBugIssue, GetTestCaseLinkedIssues, GetIssueLinkedTestCases, GetProjectIssuesWithTestCases, GetLastResultForTestCase"
    missing:
      - "Add the 5 missing delegation methods to CombinedRepository, each forwarding to r.ent.*"

  - truth: "Test cases are bidirectionally linked to issues visible from both sides"
    status: failed
    reason: "The bidirectional linking service methods and handler endpoints exist, and ent_repository.go implements all 5 missing methods, but because the build fails at main.go the entire TMS module cannot be served. The feature is implemented but unshippable."
    artifacts:
      - path: "backend/internal/tms/combined_repository.go"
        issue: "Missing delegation stubs for GetTestCaseLinkedIssues, GetIssueLinkedTestCases, GetProjectIssuesWithTestCases — preventing compilation"
    missing:
      - "Same fix as above: add delegation methods to CombinedRepository"
---

# Phase 4: TMS & Allure Verification Report

**Phase Goal:** Пользователи могут управлять тест-кейсами, выполнять ручные и автоматизированные тестовые прогоны, загружать результаты Allure из CI и видеть двусторонние связи между тестами и задачами
**Verified:** 2026-03-21
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #  | Truth                                                                                  | Status      | Evidence                                                                           |
|----|----------------------------------------------------------------------------------------|-------------|------------------------------------------------------------------------------------|
| 1  | Test cases can be created in classic and BDD formats with folder hierarchy             | ✓ VERIFIED  | testcase.go schema with classic/bdd enum, TestSuite self-ref tree, 6 Ent schemas  |
| 2  | Manual test runs can be executed step-by-step with evidence per step                  | ✓ VERIFIED  | ExecutionPanel in runs_.$runId.tsx calls updateTestResult with step_results/comment |
| 3  | Allure JSON results can be imported and matched to test cases                          | ✓ VERIFIED  | allure_parser.go, allure_matcher.go, ImportAllureResults service method, MongoDB   |
| 4  | Test cases are bidirectionally linked to issues (visible from both sides)              | ✗ FAILED    | Implementation exists but backend does not compile (CombinedRepository gap)        |
| 5  | Reports show pass/fail trends and flaky test detection                                 | ✓ VERIFIED  | GetTestTrends, GetFlakyTests service methods; trends-chart.tsx and coverage-report.tsx |

**Score:** 3/5 truths fully verified (4 partially implemented, 5 build-broken)

### Required Artifacts

| Artifact                                              | Expected                              | Status      | Details                                                                    |
|-------------------------------------------------------|---------------------------------------|-------------|----------------------------------------------------------------------------|
| `backend/ent/schema/testcase.go`                      | TestCase schema (classic+BDD)         | ✓ VERIFIED  | Fields: case_type enum, steps JSON, full_name, external_id for Allure      |
| `backend/ent/schema/testcaseversion.go`               | Versioning snapshots                  | ✓ VERIFIED  | Exists, snapshot JSON field, edges to TestCase and User                    |
| `backend/ent/schema/testsuite.go`                     | Folder/section tree                   | ✓ VERIFIED  | Self-referencing parent edge verified                                      |
| `backend/ent/schema/testrun.go`                       | Test run entity                       | ✓ VERIFIED  | run_type enum, status machine, count fields                                |
| `backend/ent/schema/testresult.go`                    | Individual test result                | ✓ VERIFIED  | allure_uuid, is_flaky, match_strategy, step_results JSON                   |
| `backend/ent/schema/component.go`                     | Components entity                     | ✓ VERIFIED  | Exists with project and test_cases edges                                   |
| `backend/internal/tms/service.go`                     | TMS business logic                    | ✓ VERIFIED  | All required methods present: ImportAllureResults, LinkTestCaseToIssue, etc|
| `backend/internal/tms/handler.go`                     | TMS HTTP handlers                     | ✓ VERIFIED  | 22+ endpoints including allure/import, reports/coverage, reports/trends    |
| `backend/internal/tms/allure_parser.go`               | Defensive Allure JSON parser          | ✓ VERIFIED  | ParseAllureResults, extractLabelValue, allureStatusToTMS                   |
| `backend/internal/tms/allure_matcher.go`              | Multi-strategy test case matching     | ✓ VERIFIED  | 4 strategies: link > label > fullName > class+method                       |
| `backend/internal/tms/mongo_repository.go`            | MongoDB raw Allure storage            | ✓ VERIFIED  | StoreAllureResults, GetAllureResultsByRunID with index setup               |
| `backend/internal/tms/combined_repository.go`         | Dual PG+Mongo repository              | ✗ STUB      | Missing 5 methods from Repository interface — build fails                  |
| `backend/internal/tms/importer.go`                    | Multi-format import                   | ✓ VERIFIED  | CSV, JSON, XML, Excel via excelize; AutoDetectMapping with RU+EN headers   |
| `backend/internal/tms/exporter.go`                    | Multi-format export including PDF     | ✓ VERIFIED  | CSV, JSON, XML, Excel, PDF via maroto v2; full implementation              |
| `frontend/src/api/tms.ts`                             | TMS API client                        | ✓ VERIFIED  | All endpoints covered: cases, suites, runs, results, links, reports, imex  |
| `frontend/src/modules/tms/components/coverage-report.tsx` | Issue-to-test coverage matrix    | ✓ VERIFIED  | Calls getCoverageReport, expandable rows, pass rate bars                   |
| `frontend/src/modules/tms/components/trends-chart.tsx` | Pass/fail trend chart               | ✓ VERIFIED  | Calls getTestTrends, stacked status bars per run                           |

### Key Link Verification

| From                                      | To                              | Via                                   | Status       | Details                                                        |
|-------------------------------------------|---------------------------------|---------------------------------------|--------------|----------------------------------------------------------------|
| `handler.go`                              | `service.go`                    | Handler.service field                 | ✓ WIRED      | All handler methods delegate to service methods                |
| `service.go`                              | `repository.go`                 | Service.repo field                    | ✓ WIRED      | CreateTestCase, FindTestCaseBy*, etc. all use s.repo           |
| `main.go`                                 | `handler.go`                    | tmsHandler.RegisterRoutes             | ✓ WIRED      | Lines 157-165 in main.go wire CombinedRepository→Service→Handler |
| `handler.go`                              | `allure_parser.go`              | service.ImportAllureResults → ParseAllureResults | ✓ WIRED | Call chain verified in service.go line 1015                |
| `allure_matcher.go`                       | `ent_repository.go`             | repo.FindTestCaseBy* methods          | ✓ WIRED      | Matcher calls FindTestCaseByExternalID, FullName, ClassMethod  |
| `combined_repository.go`                  | `mongo_repository.go`           | mongo.StoreAllureResults delegation   | ✓ WIRED      | Lines 173-178 in combined_repository.go delegate correctly     |
| `combined_repository.go`                  | `ent_repository.go` (linking)   | CreateBugIssue + GetTestCaseLinked*   | ✗ NOT_WIRED  | 5 methods exist in ent_repository.go but never forwarded       |
| `importer.go`                             | `repository.go`                 | repo.CreateTestCase per row           | ✓ WIRED      | handleImport calls svc.repo.CreateTestCase directly (line 344) |
| `exporter.go`                             | `excelize` / `maroto`           | Library calls                         | ✓ WIRED      | ExportTestCasesToExcel uses excelize.NewFile; ExportTestRunToPDF uses maroto.New |
| `runs_.$runId.tsx`                        | `frontend/src/api/tms.ts`       | updateTestResult import               | ✓ WIRED      | Import on line 10, call on line 91 with step_results payload   |

### Requirements Coverage

| Requirement | Description                                                             | Status        | Evidence                                                              |
|-------------|-------------------------------------------------------------------------|---------------|-----------------------------------------------------------------------|
| TMS-01      | Classic and BDD test case formats                                       | ✓ SATISFIED   | case_type enum in schema, BDD step rendering in cases_.$caseKey.tsx   |
| TMS-02      | Folder/suite hierarchy, tags, versioning with rollback                  | ✓ SATISFIED   | TestSuite self-ref tree, TestCaseVersion snapshot, RollbackTestCase   |
| TMS-03      | GitHub linking via external_id and @TestCase annotation matching        | ✓ SATISFIED   | external_id field on TestCase schema, FindTestCaseByExternalID         |
| TMS-04      | Manual, automatic, mixed test runs with reopen and assignment           | ✓ SATISFIED   | run_type enum, state machine, ReopenTestRun, AssignTestersToRun       |
| TMS-05      | Screenshot and comment per step result                                  | ✓ SATISFIED   | step_results JSON, file_ids per StepResult, ExecutionPanel            |
| TMS-06      | Bidirectional test-issue linking, bug creation, coverage matrix         | ✗ BLOCKED     | Service+handler+ent_repo implemented; CombinedRepository missing 5 methods → build fails |
| TMS-08      | Allure-level reporting: trends, flaky, run comparison                   | ✓ SATISFIED   | GetTestTrends, GetFlakyTests, GetRunComparison; frontend chart components |
| TMS-09      | Import from CSV, JSON, XML, Excel with field mapping                    | ✓ SATISFIED   | importer.go with AutoDetectMapping, ImportFromExcel via excelize       |
| TMS-10      | Export to CSV, JSON, XML, PDF                                           | ✓ SATISFIED   | exporter.go with CSV/JSON/XML/Excel/PDF; ExportTestRunToPDF via maroto |
| ALLU-01     | CI can POST allure-results JSON and get test run back                   | ✓ SATISFIED   | ImportAllureResults in service.go, POST /tms/allure/import endpoint   |
| ALLU-02     | Match by Allure ID or name                                              | ✓ SATISFIED   | allure_matcher.go: 4 strategies link>label>fullName>class+method      |
| ALLU-03     | Run history, statistics, Telegram notification on completion            | ✓ SATISFIED   | GetTestRunHistory, publishEvent on completion via NATS                |
| IMEX-02     | Import test cases from Excel                                            | ✓ SATISFIED   | ImportFromExcel in importer.go, import-dialog.tsx frontend             |

### Anti-Patterns Found

| File                                      | Issue                                        | Severity    | Impact                                                  |
|-------------------------------------------|----------------------------------------------|-------------|---------------------------------------------------------|
| `backend/internal/tms/combined_repository.go` | Missing 5 method delegations from Repository interface | 🛑 Blocker | Backend binary cannot be built; entire TMS module is non-functional at runtime |

No TODO/FIXME/placeholder stubs found in any TMS files.
No empty handler implementations found.
No return-null stubs found in frontend TMS components.
Frontend TypeScript compiles cleanly (`npx tsc --noEmit` exits 0).

### Human Verification Required

#### 1. Test Case BDD Format Rendering

**Test:** Create a BDD test case with Given/When/Then steps, open the detail page
**Expected:** Steps display with keyword prefix (Given/When/Then) visually distinct from action text
**Why human:** UI layout verification requires visual inspection

#### 2. Allure Import End-to-End

**Test:** POST a valid allure-results JSON array to `POST /api/projects/:id/tms/allure/import` once build gap is fixed
**Expected:** Automatic test run is created; matched results show case keys; unmatched results show as "unmapped"; NATS event fires
**Why human:** Requires running infrastructure (PostgreSQL, MongoDB, NATS) and a real HTTP request

#### 3. Import Dialog Field Mapping UI

**Test:** Upload a CSV file in the ImportDialog, verify auto-detect populates mapping, then import
**Expected:** First 5 rows preview shows in table; mapping dropdowns appear per column; import shows created/skipped counts
**Why human:** Visual/interactive UI behavior

### Gaps Summary

The phase is nearly complete — 12 of 13 requirements are satisfied at the implementation level, all 5 frontend routes exist and compile, all backend service/handler methods are implemented, and the Allure pipeline is fully wired.

**The single blocking gap is structural:** When plan 04-04 added 5 cross-module linking methods to `Repository` interface and `EntRepository`, it did not forward them through `CombinedRepository`. The Go compiler enforces full interface satisfaction, so `main.go` line 163 fails at compile time with:

```
*tms.CombinedRepository does not implement tms.Repository (missing method CreateBugIssue)
```

The 5 missing methods are: `CreateBugIssue`, `GetTestCaseLinkedIssues`, `GetIssueLinkedTestCases`, `GetProjectIssuesWithTestCases`, `GetLastResultForTestCase`. Each requires one line of delegation in `combined_repository.go` (forwarding to `r.ent.*`). This is a mechanical fix with no design decisions required.

Until this is fixed, the entire TMS module (including verified features like Allure import and test run execution) cannot run in production because the server binary cannot be built.

---
_Verified: 2026-03-21_
_Verifier: Claude (gsd-verifier)_
