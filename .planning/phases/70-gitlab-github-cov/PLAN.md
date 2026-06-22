# Phase 70 — GitLab/GitHub Coverage (SEC-04)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

1. **SEC-04 / GH-01:** `gitlab/` and `github/` **48% / 45% → ≥50%**
2. Cover `EntRepository` (0%) and `RegisterRoutes` (0%) for both modules
3. Refresh overall COV baseline

## Deliverables

### GitLab
- `ent_repository_test.go` — integration CRUD, webhook secrets cache, FindIssueByKey
- `mr_review_test.go` — parseGitLabMRURL, getString
- `handler_extended_test.go` — RegisterRoutes + RBAC owner flow (create/list/delete/pipelines/webhook)

### GitHub
- `ent_repository_test.go` — integration CRUD, links, MarkTestCasesOutdatedByFiles, FindDoneStatus
- `handler_extended_test.go` — RegisterRoutes + RBAC (integrations CRUD, workflows, webhook)

## Results

| Metric | Before | After |
|--------|--------|-------|
| `gitlab/` | 48.0% | **68.5%** |
| `github/` | 45.1% | **67.8%** |
| Overall `internal/` | 58.1% | **59.2%** |

## Success criteria

- [x] gitlab ≥50%
- [x] github ≥50%
- [x] RegisterRoutes covered both modules
- [x] EntRepository covered both modules
- [x] all gitlab/github tests pass
