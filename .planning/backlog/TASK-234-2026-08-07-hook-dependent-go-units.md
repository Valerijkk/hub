# TASK-234 — 🔌 Decouple and retain hook-dependent Go units

> Revision: **2026-08-07 founder greenfield cutover** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: launcher and runtime are `NOT_READY` / `NOT_RUN`.
> A future versioned launcher API may be separately implemented only under
> `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No
> supported command, Docker-directory entrypoint, or direct Docker/Compose,
> host test/validator, mutable-image, or ad-hoc-cleanup fallback exists.
> Executable evidence must resolve TASK/requirement/TC/risk after acceptance.

## [OUTCOME]

All 245 genuine Go unit files that currently consume private export hooks are
preserved, refactored by domain toward public behavior or a narrow justified
test seam, and moved to `backend/tests/<domain>/**`. None is reclassified as a
component test merely to simplify deletion.

## [ROLE]

- Owners: `be1`, `be2` by non-overlapping domains.
- Verifiers: `be-ver`; `arch-ver` for any new/changed production seam.
- Depends on: pinned Phase 128 partition
  `KEEP-GO-UNIT-HOOK-DEPENDENT` and TASK-197 path conventions.

## [OWNED TARGET PATHS]

sealed `backend/tests/<domain>/**` targets plus only exact production hook-source/consumer and manifest paths listed by the Phase 129 domain batches. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Exact cohort remains 245 files / 1428 declared cases / 653 hook links until a
  current-SHA manifest update proves each completed domain subtraction.
- Each unit remains service-free: no DB, broker, network, listener, process,
  filesystem provider or Docker lifecycle is introduced.
- Preferred order: existing public behavior → package-local pure collaborator →
  narrow explicit test seam. New broad export-hook surface is forbidden.
- Production-code changes have focused native unit evidence and same-session
  docs sync; behavior change requires an accepted contract, not test convenience.
- Accepted-runner Docker-only build, focused unit, race where concurrency-sensitive, and
  current-SHA manifest validation are green per domain batch.
- Source removal, destination addition and hook-consumer ledger update are one
  atomic domain commit. Remaining rows are still protected from deletion.
- Independent verification by the stated verifiers is required for every domain
  batch and for the final zero-consumer condition before the protected closure
  can be considered complete.
- At zero remaining consumers, the final batch of the real TASK-219 card may
  delete the exact hook/support closure; no synthetic task card and no
legacy coverage/count threshold blocks completion.

## [OPERATING RULES]

Never delete a genuine unit because a hook is inconvenient. Stop on ambiguous
behavior, dirty overlapping production code or a seam that widens runtime API
without architectural justification. No push and no AI attribution.
