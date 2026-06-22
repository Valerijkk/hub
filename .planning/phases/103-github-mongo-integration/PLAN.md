# Phase 103 — GitHub Push + Mongo Integration Tests

**Branch:** `feat/v2.5-trust`  
**Depends on:** Phase 102  
**Goal:** Push boundary packages (github 75%→80%+) and add mongo integration tests (skip without MongoDB)

## Scope

| Package | Before | Target | Wave |
|---------|--------|--------|------|
| `internal/github/` | 75.0% | ≥78% | cipher, workflow detail, preview external |
| `internal/tracker/` | 75.0% | ≥75% | comment update/delete handlers |
| `internal/calls/chat_mongo.go` | 0% | covered when Mongo available | integration |
| `internal/search/searchMessages` | 8% | covered when Mongo available | integration |
| `internal/messenger/mongo_repository` | 0% | existing test, skip without Mongo | — |

## Tasks

### Wave 103a — github repository + service

- [x] `encryptField` round-trip + ciphertext in DB
- [x] `ListAllWebhookSecrets` cache path
- [x] `GetWorkflowRunDetail` httptest GitHub API

### Wave 103b — github handler + preview

- [x] `GetWorkflowRunDetail` HTTP integration
- [x] `resolveExternalPreview` blocked URL + example.com (network)
- [x] CreateBranch / ListGitHubLinks edge cases

### Wave 103c — tracker handlers

- [x] UpdateComment + DeleteComment flow
- [x] Invalid JSON / invalid comment ID / GetActivity

### Wave 103d — mongo integration (skip if no `MONGODB_TEST_URL`)

- [x] `calls/chat_mongo_wave103_test.go`
- [x] `search/service_wave103_mongo_test.go`
- [x] Existing `messenger/mongo_repository_test.go` unchanged

## Deferred (Phase 104+)

- Mongo coverage in CI (`MONGODB_TEST_URL` in docker-compose.test.yml job)
- Fix flaky `TestHandler_RegisterRoutes_IntegrationsCRUD` (real GitHub API on ListWorkflowRuns)
- Commit + PR (on user request)

## Success criteria

- [x] github ≥78%
- [x] COV overall ≥78.2%
- [x] race PASS on touched packages (mongo tests skip gracefully)
