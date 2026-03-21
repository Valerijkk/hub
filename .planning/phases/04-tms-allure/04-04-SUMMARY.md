---
phase: 04-tms-allure
plan: 04
subsystem: tms
tags: [linking, coverage, reporting, import, export, excelize, maroto, pdf]
dependency_graph:
  requires: [04-03]
  provides: [bidirectional-test-issue-linking, coverage-report, trends, import-export]
  affects: [tms-service, tms-handler, tms-repository, issue-schema, frontend-tms]
tech_stack:
  added: [excelize/v2, maroto/v2]
  patterns: [shared-ent-client-for-cross-module, multi-format-import-export, field-auto-mapping]
key_files:
  created:
    - backend/internal/tms/importer.go
    - backend/internal/tms/exporter.go
    - frontend/src/modules/tms/components/coverage-report.tsx
    - frontend/src/modules/tms/components/trends-chart.tsx
    - frontend/src/modules/tms/components/import-dialog.tsx
  modified:
    - backend/ent/schema/issue.go
    - backend/internal/tms/repository.go
    - backend/internal/tms/ent_repository.go
    - backend/internal/tms/service.go
    - backend/internal/tms/handler.go
    - backend/internal/tms/dto.go
    - backend/internal/tms/service_test.go
    - frontend/src/api/tms.ts
    - frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx
decisions:
  - Shared Ent client for cross-module bug creation (no tracker import)
  - excelize v2 for Excel import/export, maroto v2 for PDF generation
  - Auto-detect field mapping for CSV/Excel with Russian+English header support
metrics:
  duration: 15min
  completed: "2026-03-21T02:42:00Z"
---

# Phase 04 Plan 04: Cross-Module Linking, Reporting, Import/Export Summary

Bidirectional test-to-issue linking with shared Ent client bug creation, coverage/trends/comparison reports, and multi-format import/export (CSV/JSON/XML/Excel/PDF)

## Completed Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Cross-module linking + coverage + reporting + bug creation | a0e4b96 | 14 files (schema, service, handler, repository, dto, tests, frontend) |
| 2 | Import/export (CSV, JSON, XML, Excel, PDF) | a28a101 | 5 files (importer, exporter, tests, go.mod/sum) |

## What Was Built

### Backend

**Bidirectional test-to-issue linking:**
- Added `linked_test_cases` backref edge on Issue schema (From TestCase.linked_issues)
- Repository methods: GetTestCaseLinkedIssues, GetIssueLinkedTestCases, GetProjectIssuesWithTestCases
- Service methods: LinkTestCaseToIssue, UnlinkTestCaseFromIssue, GetTestCaseLinks
- Handler endpoints: GET/POST/DELETE /cases/:caseKey/links

**Bug creation from failed test:**
- CreateBugIssue repository method creates Issue with type=bug via shared Ent client (no tracker import)
- CreateBugFromFailedResult service method auto-titles "FAILED: {test case title}"
- Handler endpoint: POST /runs/:runId/results/:resultId/bug

**Coverage, trends, comparison reports:**
- GetCoverageReport: issues with linked test cases + last result status per test case
- GetTestTrends: pass/fail/skip/broken counts and pass rate for last N runs
- GetRunComparison: side-by-side diff showing changed statuses between two runs
- Handler endpoints: GET /reports/coverage, /reports/trends, /reports/comparison

**Import/export:**
- ImportFromCSV, ImportFromJSON, ImportFromXML with auto-detect field mapping (Russian + English)
- ImportFromExcel using excelize v2 with sheet selection and header detection
- ExportTestCasesToCSV/JSON/XML/Excel and ExportTestRunToPDF/CSV
- Excel export: bold headers, auto column widths, alternating row colors
- PDF export using maroto v2: title, stats table, results table with status coloring
- Handler endpoints: POST /tms/import (multipart), GET /tms/export, GET /runs/:runId/export
- Validation with row-level error reporting

### Frontend

**API client (tms.ts):**
- linkTestCaseToIssue, unlinkTestCaseFromIssue, getTestCaseLinks
- createBugFromResult
- getCoverageReport, getTestTrends, getRunComparison
- importTestCases (multipart), exportTestCases, exportTestRunReport

**Components:**
- CoverageReport: expandable table showing issues with test coverage percentage and status bars
- TrendsChart: horizontal stacked status bars for each run with pass rate
- ImportDialog: drag-and-drop file upload, auto format detection, result summary with validation errors

**Test case detail page:**
- Link/unlink issue functionality with inline input and remove buttons

### Tests (23 total, all passing)

- TestLinkTestCaseToIssue: link, verify, unlink, verify empty
- TestCreateBugFromFailedResult: create bug from failed result, verify title
- TestGetCoverageReport: verify report aggregation with linked test cases
- TestImportFromCSV: parse 3 rows with auto-detected headers
- TestAutoDetectMapping: verify Russian+English header detection
- TestExportToCSV: export and re-parse to verify roundtrip
- TestValidateImportData: verify row-level validation errors

## Deviations from Plan

None - plan executed as written.

## Decisions Made

1. **Shared Ent client for cross-module bug creation**: CreateBugIssue uses ent.Client directly to create Issue entities without importing the tracker package, avoiding circular dependencies
2. **excelize v2 for Excel**: Provides full XLSX read/write with styling (bold headers, alternating row colors, auto column widths)
3. **maroto v2 for PDF**: Generates structured PDF reports with title, stats table, and results table with status color coding
4. **Auto-detect field mapping**: Supports both Russian and English column headers for import (e.g., "Title"/"Название", "Steps"/"Шаги")

## Self-Check: PASSED
