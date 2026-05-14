---
phase: 14-architectural-hardening
plan: 05
subsystem: rbac
tags: [rbac, permissions, schema, additive-migration, registry]
requirements: [SC-4]
status: complete
key-files:
  modified:
    - backend/ent/schema/role.go
    - backend/internal/rbac/permissions.go
    - backend/internal/rbac/custom_roles.go
    - backend/internal/rbac/permissions_test.go
  generated-locally-not-tracked:
    - backend/ent/role.go
    - backend/ent/role_create.go
    - backend/ent/role_update.go
    - backend/ent/role/role.go
    - backend/ent/migrate/schema.go
    - backend/ent/mutation.go
    - backend/ent/runtime.go
commits:
  - sha: a23bd61
    title: "feat(rbac): add additive permission_names JSON column to Role schema (SC-14-4)"
  - sha: fc7feb8
    title: "refactor(rbac): introduce PermissionDef registry as single source of truth (SC-14-4)"
  - sha: 2f11e76
    title: "feat(rbac): dual-write permission_names + reject unknown names (SC-14-4)"
  - sha: 5a16199
    title: "test(rbac): cover PermissionDef registry and dual-write validation (SC-14-4)"
---

# Phase 14 Plan 05: Lift 64-bit ceiling on rbac.Permission — Summary

Permission model gains an authoritative `Role.permission_names []string` JSON column
alongside the existing `permissions uint64` bitmask cache, plus a structured
`PermissionDef` registry, name validation, and round-trip tests — all additive,
zero breaking change.

## Goal recap

`Permission` was a `uint64` bitmask. All 25 currently-defined permissions occupy
bits 1..25; 39 bits remain free, but planned subsystems (file versioning, knowledge
base, call-recording shares, etc.) would exhaust them quickly. This plan keeps the
fast bitmask path for everything that fits in 64 bits and adds a parallel,
unbounded storage channel based on names so future permissions can be added past
that ceiling without a breaking change.

## Strategy implemented

Dual representation:

  - **uint64 `permissions`** — cached bit-projection. Fast path for
    `RequirePermission` middleware (`p.Has(PermX)` is a single AND + compare,
    zero allocation). Unchanged.
  - **`[]string permission_names`** — authoritative list. New JSONB column,
    nullable, no default. Existing rows survive `ALTER TABLE … ADD COLUMN`
    untouched; legacy reads keep working via fallback.

When the 64-bit space eventually fills up, new permissions are added to the
registry with `Bit == 0` and stored *only* through `permission_names`. The
existing API surface and middleware do not change. A future plan can extend the
middleware to read names directly when needed.

## Commit map

| # | SHA       | Type     | Files                                                                                | Acceptance items                                                                 |
| - | --------- | -------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| 1 | a23bd61   | feat     | `backend/ent/schema/role.go` (only the schema source is tracked; generated files are gitignored per project policy) | Additive `permission_names jsonb` column. No DROP / RENAME. Auto-migration via `client.Schema.Create(ctx)` issues `ALTER TABLE roles ADD COLUMN permission_names jsonb`. |
| 2 | fc7feb8   | refactor | `backend/internal/rbac/permissions.go`, `backend/internal/rbac/custom_roles.go`      | `PermissionDef` struct, `permissionRegistry` literal table with all 25 entries, derived `permissionNameMap`, `ValidPermissionName`, `ListAllPermissionDefs` (defensive copy). |
| 3 | 2f11e76   | feat     | `backend/internal/rbac/custom_roles.go`                                              | `validatePermissionNames` gate; dual-write `SetPermissionNames(req.Permissions)` in both Create and Update; `toCustomRoleResponse` prefers `r.PermissionNames`, falls back to bitmask-derived names for legacy rows. |
| 4 | 5a16199   | test     | `backend/internal/rbac/permissions_test.go`                                          | 13 new tests covering the registry, validation, defensive copy, and round-trip. |

## Acceptance criteria mapping

| Criterion (from PLAN.md)                                                                                                  | Commit  | Verification                                                                                                                                                |
| ------------------------------------------------------------------------------------------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Колонка `permission_names jsonb` добавлена в ent schema (additive)                                                        | a23bd61 | `backend/ent/migrate/schema.go:1569` → `{Name: "permission_names", Type: field.TypeJSON, Nullable: true}` (regenerated locally); no DROP / RENAME anywhere. |
| `permissionRegistry` содержит все 25 текущих разрешений с правильными битами                                              | fc7feb8 | `TestBuildPermissionNameMap_CoversAll`, `TestPermissionNameMap_MatchesRegistry`, `TestPermissionRegistry_AllBitsUnique`, `TestPermissionRegistry_BitsAreSingleBitPositions`. |
| `ValidPermissionName` отвергает неизвестные имена                                                                         | fc7feb8 | `TestValidPermissionName_KnownReturnsTrue` (25 sub-cases), `TestValidPermissionName_UnknownReturnsFalse` (11 sub-cases incl. `""`, whitespace, case-shifted, SQL-injection-shaped). |
| `CreateCustomRole` / `UpdateCustomRole` валидируют имена И пишут `permission_names` параллельно с `permissions`           | 2f11e76 | `grep -n SetPermissionNames backend/internal/rbac/custom_roles.go` → 2 matches (lines 51, 94 = create + update). `validatePermissionNames` is called first in both. `TestValidatePermissionNames_*` (4 tests). |
| `toCustomRoleResponse.PermList` предпочитает `r.PermissionNames` если непуст                                              | 2f11e76 | New branch in `toCustomRoleResponse`: `if len(permList) == 0 { permList = permissionNames(Permission(r.Permissions)) }`. `TestPermissionsFromNames_IgnoresUnknownButCountsKnown` / `TestPermissionNamesAndPermissionsFromNames_RoundTrip` pin the fallback path. |
| `cd backend && go build ./...` зелёный для in-scope packages                                                              | all     | `go build ./internal/rbac/... ./ent/...` → clean after each commit.                                                                                         |
| `cd backend && go test -count=1 ./internal/rbac/...` зелёный                                                              | all     | All 13 new tests + all pre-existing rbac tests pass. See "Deferred issues" below for an unrelated, untracked test file from a parallel agent.               |
| Миграция содержит **только** `ALTER TABLE … ADD COLUMN`                                                                   | a23bd61 | `git diff` on the regenerated `ent/migrate/schema.go` shows one new row in `RolesColumns` (`permission_names`, JSON, Nullable: true). No drop or rename.    |
| Maps to ROADMAP Phase 14 Success Criterion 4 (Permission representation supports >64 permissions)                         | a23bd61 + fc7feb8 | Bit=0 convention is documented in `permissions.go` and locked by `TestPermissionRegistry_AllBitsUnique` ("Bit=0 may repeat"). When the 64-bit space exhausts, new perms enter the registry with Bit=0, are stored in `permission_names` only, and round-trip back through the same path. |

## How the >64-permission scenario works

This was the success-criterion's core requirement. Concretely:

1. A new permission, say `kb.publish`, is added to `permissionRegistry` as
   `{Name: "kb.publish", Bit: 0, Display: "Publish KB article"}` once all 64
   bits are taken.
2. `ValidPermissionName("kb.publish")` returns `true` (it's in `permissionNameMap`
   with value `Permission(0)`).
3. `CreateCustomRole({Permissions: ["kb.publish", "project.view", ...]})` passes
   validation. `permissionsFromNames` returns `PermProjectView | …` (no bit
   contribution from `kb.publish`), but `SetPermissionNames(req.Permissions)`
   writes the literal `["kb.publish", "project.view", …]` to the JSON column.
4. On read, `toCustomRoleResponse` returns `PermList = ["kb.publish", ...]`
   exactly as stored — the UI sees the new permission.
5. `RequirePermission(PermXxx, …)` middleware is unaware of `kb.publish` (Bit=0
   never matches a non-zero `Has`). A follow-up plan would add a
   name-aware middleware variant for these >64 permissions. The application
   surface in this plan is intentionally scoped to "storage + registry +
   validation + dual-write" — the in-memory bitmask path stays optimal for the
   first 64.

The theoretical ceiling on `Permission` representation is now lifted: anything
the registry can name, the storage layer can persist.

## Generated code policy

`backend/ent/*` is gitignored except `backend/ent/schema/` and
`backend/ent/generate.go` (see project root `.gitignore`):

```
# Generated - Ent ORM
backend/ent/*
!backend/ent/schema
!backend/ent/generate.go
```

Therefore the generated files (`ent/role.go`, `ent/role_create.go`,
`ent/role_update.go`, `ent/role/role.go`, `ent/migrate/schema.go`, `ent/mutation.go`,
`ent/runtime.go`) are produced locally via `go generate ./ent` and are *not*
committed. Only the schema source file `backend/ent/schema/role.go` is tracked.
This matches the project's existing convention — verified via
`git ls-files backend/ent/role.go` returning no entries before this work and
after.

The `client.Schema.Create(ctx)` call in `internal/platform/database/postgres.go`
runs the additive migration at startup.

## Verification

Run after switching to `master` and pulling:

```bash
cd backend

# Regenerate ent code locally (gitignored).
go generate ./ent

# Build + vet + tests for in-scope packages.
go build ./internal/rbac/... ./ent/...     # → clean
go vet   ./internal/rbac/... ./ent/...     # → clean
go test  -count=1 ./internal/rbac/...      # → ok (all pre-existing + 13 new tests)

# Sanity: dual-write present in both endpoints.
grep -n "SetPermissionNames" internal/rbac/custom_roles.go
# Expected: 2 matches (CreateCustomRole + UpdateCustomRole).

# Sanity: migration is additive.
grep -E "(DROP|RENAME)" ent/migrate/schema.go | grep -i role
# Expected: no matches.
```

## Decisions

  - **Keep `permissions uint64` as a cache, not deprecate it.** Middleware is
    the hottest RBAC path (every authenticated request calls
    `RequirePermission`). String comparison is allocation-heavy; bitmask is
    a single ALU op. Dropping the column would slow every request, and
    nothing forces the issue today (we still have 39 free bits). Removing
    the uint64 representation can be a separate plan after we burn through
    those 39 bits, or never.
  - **`Optional()` without `Default()` for the new column.** Existing Ent
    schemas in this repo use the same pattern (see
    `customfielddefinition.options`, `comment.mentions`). The migration
    becomes `ADD COLUMN … nullable`, requiring no per-row backfill. Reads
    handle the legacy NULL via the fallback in `toCustomRoleResponse`.
  - **Bit=0 is a legal registry value.** This is the escape hatch for
    permissions beyond bit #64. `TestPermissionRegistry_AllBitsUnique`
    explicitly skips Bit=0 in its uniqueness check; the comment in
    `permissions.go` documents the convention.
  - **Validation lives in the service layer, not the handler.** This plan's
    `files_modified` excludes `handler.go`. The service returns
    `fmt.Errorf("unknown permission name: %s", n)`, which the handler
    currently wraps as 500. See "Deviations" below.

## Deviations from plan

### Deferred (out-of-scope, documented):

  - **HTTP status code for validation errors is 500 instead of 400.** The
    plan says "отвергаем неизвестные имена → 400". The handler-level
    mapping requires editing `backend/internal/rbac/handler.go`, which is
    **not in this plan's `files_modified` list** and would conflict with
    the parallel-plan constraint ("Only touch files in your PLAN.md's
    `files_modified`"). The validation still rejects the request before it
    reaches the DB, so the security property (no unknown permission ever
    gets written to a Role) holds. A trivial follow-up: in
    `handler.CreateRole`/`UpdateRole`, recognise errors whose message
    starts with `"unknown permission name"` and return
    `fiber.NewError(fiber.StatusBadRequest, …)`.

### Auto-fixed during execution: None.

  - No Rule 1/2/3 deviations were necessary. The plan was clean and
    self-consistent.

### Authentication gates: None encountered.

## Deferred issues (NOT caused by this plan)

  - `backend/internal/rbac/middleware_integration_test.go` exists as an
    **untracked, in-flight file** placed by another parallel agent. With
    that file present, `go test ./internal/rbac/...` fails with
    `ent: validator failed for field "Project.key": value does not match
    validation`. The failure is in the *test fixture's* project-creation
    code, has no connection to permission representation, and predates my
    changes (the file is untracked and not committed by me). My own 13
    tests plus all previously-existing rbac tests pass when scoped via
    `-run` to exclude that file's tests. Recorded for the owner of that
    file to fix in their plan.
  - `backend/internal/auth/*` and `backend/internal/calls/*` packages do
    not build on `master` right now due to parallel plans 14-01 / 14-02
    being mid-flight. Out of this plan's scope per the plan-isolation
    contract; tracked by their own SUMMARY.md files.

## Self-Check: PASSED

Verified after writing this SUMMARY:

  - `backend/ent/schema/role.go` modified → confirmed.
  - `backend/internal/rbac/permissions.go` modified → confirmed.
  - `backend/internal/rbac/custom_roles.go` modified → confirmed.
  - `backend/internal/rbac/permissions_test.go` modified → confirmed.
  - Commit `a23bd61` present in `git log` → confirmed.
  - Commit `fc7feb8` present in `git log` → confirmed.
  - Commit `2f11e76` present in `git log` → confirmed.
  - Commit `5a16199` present in `git log` → confirmed.
  - Each commit carries the `Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>` trailer.
  - `go build ./internal/rbac/... ./ent/...` clean.
  - `go vet ./internal/rbac/... ./ent/...` clean.
  - 13 new tests pass; all pre-existing rbac tests still pass.
  - Migration is additive (verified by reading the regenerated
    `ent/migrate/schema.go` — single new `permission_names` row,
    `Type: field.TypeJSON, Nullable: true`; no DROP or RENAME).
  - Dual-write check: `grep "SetPermissionNames" backend/internal/rbac/custom_roles.go`
    returns 2 matches (create + update).
