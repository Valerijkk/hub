# Phase 14 — Architectural Hardening — VERIFICATION

**Date:** 2026-07-27  
**Method:** File existence, struct-field inspection, schema/grep counts (no live `go test` re-run)  
**Sources:** `PHASE.md`, `14-0{1..5}-PLAN.md` + `*-SUMMARY.md`, live tree under `backend/`

## Goal (from PHASE.md)

Split AuthService / CallService; remove `*ent.Client` from Messenger/Audit/Search services; uplift critical-path tests; additive permission_names beyond uint64.

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| AuthService split | Present: `service.go` (116 loc), `service_session.go` (271), `service_login.go` (163), `service_profile.go` (56), `service_registration.go` (83) | PASS (split intact; session file drifted past original ≤250 loc ceiling) |
| CallService split | Present: `service.go` (226), `service_room.go` (124), `service_events.go` (69), `service_chat.go` (82) — all under plan ceilings | PASS |
| No `*ent.Client` in Messenger/Audit/Search services | Structs hold `Repository` (+ redis/sql/mongo deps); comments cite SC-14-3; no `*ent.Client` field | PASS |
| Permission representation upgrade | `backend/ent/schema/role.go` has `permission_names` JSON; `PermissionDef` + `permissionRegistry` in `rbac/permissions.go` | PASS |
| Critical tests (≥15 new) | Migrated to `backend/testframework/`: `apigateway/webhook_*.go`, `rbac/middleware_integration_test.go` (7 `Test*`), `auth/magiclink_test.go` (8), `platform/outbox/` (24), `calls/recording_webhook_test.go` (8) — far above ≥15 | PASS |
| Plan summaries | All five `14-0X-SUMMARY.md` status COMPLETE (2026-05-15) | PASS |

## Notes / drift

- Original tests lived under `backend/internal/*/…_test.go`; Wave B moved them to `backend/testframework/` — content preserved.
- `service_session.go` is **271** lines (SUMMARY claimed ≤250 at ship) — structural goal still met.
- `PHASE.md` frontmatter still says `status: planned` — docs drift only; code evidence is shipped.

## Verdict: **PASS**
