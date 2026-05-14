---
phase: 14-architectural-hardening
plan: 03
type: execute
status: complete
requirements: [SC-3]
completed_at: 2026-05-15
commits:
  - 07d4626 refactor(messenger): remove *ent.Client from MessengerService
  - 5991c61 refactor(audit): remove *ent.Client from AuditService
  - 64d69c1 refactor(search): remove *ent.Client from SearchService
  - 55a43d5 refactor(cmd/server): wire new constructors after ent.Client removal
files_touched:
  added:
    - backend/internal/search/repository.go
  modified:
    - backend/internal/messenger/repository.go
    - backend/internal/messenger/ent_repository.go
    - backend/internal/messenger/service.go
    - backend/internal/messenger/service_channel_invites.go
    - backend/internal/messenger/service_polls.go
    - backend/internal/messenger/service_test.go
    - backend/internal/audit/repository.go
    - backend/internal/audit/service.go
    - backend/internal/audit/handler.go
    - backend/internal/search/service.go
    - backend/cmd/server/main.go
---

# Phase 14 Plan 03: Remove *ent.Client from MessengerService / AuditService / SearchService

## One-liner

Replaced direct `*ent.Client` field access in three large services with focused
Repository interfaces — Ent ORM is now confined to repository implementations,
unblocking SC-3 (no ORM leakage into business logic) and the wider goal of
making services testable without database fixtures.

## Goal

Eliminate the LO-03 "ent.Client leak in services" finding by:
1. Removing the `db *ent.Client` struct field from `MessengerService`,
   `AuditService` and `SearchService`.
2. Moving every Ent query that lived in those service files into the
   corresponding `Repository` (existing where possible, new where needed).
3. Wiring `cmd/server/main.go` with the new constructor signatures.

## Implementation

### 1. MessengerService (`07d4626`)

`MessengerService` previously held both `repo Repository` (used for most
operations) and `db *ent.Client` (used for @role mentions, typing display
name, invite links, polls + transactional voting). Three new sub-interfaces
were added to the composite `Repository`:

- `MentionsRepository` — `ListProjectMembersByRole`, `GetUserDisplayName`
- `ChannelInviteRepository` — `CreateChannelInvite`, `FindChannelInviteByToken`,
  `IncrementInviteUses`, `RevokeChannelInvite`, `ListChannelInvites`
- `PollRepository` — `CreatePollWithOptions` (Tx-aware), `GetPollWithOptions`,
  `GetPoll`, `ClosePoll`, `VotePoll` (Tx-aware), `ListVotes`

Two transactions previously opened in `service_polls.go` (`s.db.Tx(ctx)` for
`CreatePoll` and `VotePoll`) were moved inside `EntRepository` — `Tx` no
longer leaks across the service boundary. The service-level code now just
hands input data to the repo and gets back already-committed entities.

The mock repository in `service_test.go` was extended with in-memory
implementations of all new methods. All existing tests continue to pass.

### 2. AuditService (`5991c61`)

`AuditService` had a `*ent.Client db` field alongside `*Repository repo` for
historical reasons: the original Repository covered only `AuditLog`, while
everything else (resource limits, trash CRUD, user admin, GDPR cascade) was
inlined as `s.db.Foo.Query()...`. The existing Repository was extended with
24 new methods (~330 lines), grouped into three sections:

- `SystemSetting` — `GetSystemSetting`, `UpsertSystemSetting`
- `Trash` — per-entity `ListDeleted{Issues,TestCases,Files,Projects}`,
  `Restore{Issue,TestCase,File,Project}`, `SoftDelete{Issue,TestCase,File,Project}`,
  `PurgeExpired{Issues,TestCases,Files,Projects}`. The list methods return a
  normalised `TrashItem` DTO so the service can `switch entityType` without
  importing Ent.
- `User admin` — `DeactivateUser`, `DeleteUserSessions`, `Reassign*Issues`,
  `ClearUserAssignments`, `DeleteUser{Comments,Reactions,ActivityLogs,TimeLogs}`,
  `HardDeleteUser`, `IsSystemAdmin`.

The `handler.requireAdmin` middleware previously used `h.service.db.User.Get`
directly — replaced by a new `AuditService.IsSystemAdmin` method that
delegates to the repo, removing the last direct Ent access from the audit
package's HTTP layer.

### 3. SearchService (`64d69c1`)

Created `internal/search/repository.go` introducing `Repository` interface
and `EntRepository` implementation. The interface covers exactly the typed
Ent paths: `IsSystemAdmin` (HI-12 admin bypass) and full `SavedFilter` CRUD
(`Create/List/Get/Delete/UpdateShared`).

`sqlDB *sql.DB` and `mongoDB *mongo.Database` stay in `SearchService` — they
are database driver level (raw `tsvector` + `$text`), not ORM level, and a
repository abstraction would only add ceremony.

`savedFilterToDTO` still consumes `*ent.SavedFilter`. Removing Ent types from
public service signatures is explicitly out of scope per the plan's Risk
section ("это task для будущего").

### 4. `cmd/server/main.go` (`55a43d5`)

Three call sites updated:

```go
// before
messengerSvc := messengermod.NewMessengerService(messengerRepo, nc, presenceSvc, unreadSvc, db)
auditSvc     := auditmod.NewAuditService(db, auditRepo, rdb)
searchSvc    := searchmod.NewSearchService(db, rawDB, mongoDB)

// after
messengerSvc := messengermod.NewMessengerService(messengerRepo, nc, presenceSvc, unreadSvc)
auditSvc     := auditmod.NewAuditService(auditRepo, rdb)
searchRepo   := searchmod.NewEntRepository(db)
searchSvc    := searchmod.NewSearchService(searchRepo, rawDB, mongoDB)
```

Committed last and separately, per the plan's rollback guidance — each prior
commit is independently revertable, the final `main.go` commit ties them
together.

## Verification

Final build / vet / test against the working tree:

```text
$ go build ./...           → clean
$ go vet ./...             → clean
$ go test -count=1 ./internal/messenger/... ./internal/audit/... ./internal/search/...
ok  github.com/valerii/hub/backend/internal/messenger  0.954s
ok  github.com/valerii/hub/backend/internal/audit      0.663s
ok  github.com/valerii/hub/backend/internal/search     0.604s
```

Full repository test run (`go test -count=1 ./...`) — all 33 internal packages
green; no regressions outside the touched modules.

Static grep checks from the plan:

```text
*ent.Client in service files (non-comment lines):  0
s.db.X usage in service files:                     0
```

(The grep does match two comment lines that describe what was removed —
`"SC-14-3: сервис не хранит *ent.Client напрямую"` — these are intentional
documentation, not actual field declarations or method calls.)

## Acceptance Criteria Mapping

| Plan criterion                                                                 | Status |
| ------------------------------------------------------------------------------ | ------ |
| `grep "*ent.Client" backend/internal/messenger/service*.go` пусто              | OK (only comment in service.go) |
| `grep "*ent.Client" backend/internal/audit/service.go` пусто                   | OK (only comment) |
| `grep "*ent.Client" backend/internal/search/service.go` пусто                  | OK |
| `grep "s\.db\." backend/internal/{messenger,audit,search}/service*.go` пусто   | OK |
| `cd backend && go build ./...` зелёный                                         | OK |
| `cd backend && go test -count=1 ./internal/{messenger,audit,search}/...` ok    | OK |
| `cmd/server/main.go` обновлён — три call-site                                  | OK (commit `55a43d5`) |
| HTTP API behaviour unchanged                                                   | OK (no public signatures changed except constructors) |
| Maps to ROADMAP Phase 14 SC-3                                                  | OK |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 — blocking issue] `audit/handler.go` direct `h.service.db.User.Get` reference**

- **Found during:** Task 2 (AuditService) build verification
- **Issue:** Removing `db` from `AuditService` broke `handler.requireAdmin`,
  which read `h.service.db.User.Get(c.Context(), userID)` directly.
- **Fix:** Added `Repository.IsSystemAdmin` + `AuditService.IsSystemAdmin`
  wrapper; handler now calls `h.service.IsSystemAdmin(...)`. This was the
  natural extension of the plan's goal and matches the same `IsSystemAdmin`
  pattern that already exists in `search/repository.go`.
- **Files modified:** `internal/audit/handler.go`, `internal/audit/repository.go`,
  `internal/audit/service.go`
- **Commit:** part of `5991c61`

No other deviations. Plan executed as written.

## Notes

- The plan's "Risk" section flagged the poll transaction migration as the
  trickiest part. Implementation followed the prescribed pattern exactly:
  `CreatePollWithOptions` and `VotePoll` open + commit/rollback the Tx
  internally and return only post-commit data. The service no longer touches
  `ent.Tx` anywhere.
- `audit/retention.go` and `messenger/auto_channels.go` still hold an
  `*ent.Client` field, but they are explicitly out of the plan's scope —
  retention runs raw deletion cron and auto_channels uses `EntRepository`
  for the one read it needs. SC-3 targets the three large services and they
  are now clean.
- All three SC-14-3 services keep their original constructor signature
  shape (positional, no functional options) — minimum churn for downstream
  callers and for the integration tests planned in 14-04.

## Self-Check: PASSED

All commits verified present in git log:
- `07d4626` — present (refactor messenger)
- `5991c61` — present (refactor audit)
- `64d69c1` — present (refactor search)
- `55a43d5` — present (wire main.go)

All listed files modified/created exist on disk:
- `backend/internal/search/repository.go` — created
- All ten modified files — present and matching tree state
