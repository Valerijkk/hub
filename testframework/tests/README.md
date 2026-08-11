# Executable suite boundary

## Belongs here

Executable pytest suites belong under one layer and then the owning domain
where applicable:

- `api/` asserts public API observable behavior.
- `e2e/` asserts public end-to-end observable behavior.
- `contract/` verifies API, schema, event, and source contracts.
- `integration/` verifies DB, broker, and storage boundaries.

## Forbidden

Native units, framework selftests, private implementation assertions in any
layer,
cross-layer helper warehouses, and imports from quarantined legacy suites are
forbidden. Native units stay in `backend/tests/**` and `frontend/tests/**`.

## First consumer

The first accepted suite consumer must start from an approved `TC-*` case and
create the narrow test first. No executable suite is added by this slice.
