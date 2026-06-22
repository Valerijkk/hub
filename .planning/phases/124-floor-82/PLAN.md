# Phase 124 — Floor ≥82% (tracker / tms / project)

> Branch: `feat/v2.5-trust` · Docker + `-tags hubtest`

## Goal

| Package | Baseline | Target | Result |
|---------|----------|--------|--------|
| tracker | 81.3% | ≥82% | ✅ **82.4%** |
| tms | 81.7% | ≥82% | 🔄 **81.8%** (+0.1%) |
| project | 81.0% | ≥82% | 🔄 **81.7%** (+0.7%) |

## Tasks

- [x] W1: `tracker/repository_wave124_test.go` — CreateIssue all fields, ListIssues filters
- [x] W2: `tracker/handler_wave124_test.go` — DeleteWorkflowTransition invalid ID
- [x] W3: `tms/service_wave124_test.go` — ent rollback, link, update all fields
- [x] W4: `tms/ent_repository_wave124_test.go` — CreateBugIssue, ReopenTestRun
- [x] W5: `project/repository_wave124_test.go` — org + outbox Create
- [x] W6: `project/service_wave124_test.go` — invite edges, team, restore
- [x] W7: Docker verify

## Note

tms/project остаются на 81.7–81.8% — gap ≤0.3% закрывается в Wave C (FA-01) или отдельной micro-phase при TF migration.
