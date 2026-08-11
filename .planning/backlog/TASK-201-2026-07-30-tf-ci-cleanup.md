# TASK-201 — 🔁 Per-batch entrypoint switch and final CI cleanup

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Evidence resolves requirement `TASK-201`, stable cutover TC IDs and consumer-drift/removed-path risks; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Every Phase 129 move/delete batch changes its affected CI/scripts/docs in the
same atomic batch, so no committed state points at a removed path.  After the
last TASK-234/TASK-219 batch, the final cleanup makes only `backend/tests`,
`frontend/tests` and root Python entrypoints authoritative.

## [ROLE]

- Owner: `devops-engineer`; verifier: `qa-auto-ver`.
- Batch-switch work depends on the matching TASK-197, TASK-198, TASK-219 or
  TASK-234 batch and runs inside that same atomic commit.
- Final cleanup depends on TASK-193, TASK-197, TASK-198, TASK-219 and TASK-234.

## [OWNED TARGET PATHS]

only exact affected `.github/workflows/**`, `.gitlab-ci.yml`, `scripts/**`, docs consumers and manifest records enumerated by each Phase 129 batch; final cleanup uses the same sealed consumer set. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- **Work package 1 — per-batch switch:** before each commit, enumerate all
  executable and documentation consumers of every moved/deleted path; edit the
  affected consumers in that commit; validate `removed_path_reference_count=0`
  and record the exact edits under the same batch in `cutover-manifest.json`.
- **Work package 2 — final cleanup:** after TASK-234 and the final zero-consumer
  TASK-219 closure, remove transitional selectors without creating an entrypoint
  under `testframework/docker/**`. A separately approved runner slice may later
  integrate CI through the sole future versioned API under `testframework/runner/**`.
- No intermediate commit may invoke a path deleted by that commit.  A remaining
  protected TASK-234 path must be selected explicitly and must never be deleted.
- Required lanes fail on zero collection/unexpected skip and sentinel failure.
- No old path/reference remains in executable config or documentation in scope.
- Clean Docker run green; images/containers/volumes/networks cleaned or TTL-handoff.
- `cutover-manifest.json` has one `verified` record per applied batch, including
  pre/post reference counts, declarative execution requirements, image digest,
  source SHA and evidence URI.
