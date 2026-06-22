# Phase 122 — TMS + Project ≥82%

> Branch: `feat/v2.5-trust` · Docker · `-tags hubtest`

## Goal

| Package | Baseline | Target |
|---------|----------|--------|
| tms | 79.3% | ≥82% |
| project | 80.4% | ≥82% |

## Tasks

- [x] W1: `handler_wave122_test.go` — suite/run HTTP, ReopenTestRun service
- [x] W2: `ent_repository_wave122_test.go` — UpdateSuite, MoveSuite, ReopenTestRun, UpdateTestRunStats
- [x] W3: `project/handler_wave122_test.go` — unarchive, invites, trash
- [x] W4: `project/service_wave122_test.go` — member role, trash, restore
- [x] W5: Docker verify — tms **80.2%**, project **81.0%** (phase 124 closes gap to 82%)

## Verify

```bash
go test -tags hubtest -cover ./internal/tms/... ./internal/project/...
```
