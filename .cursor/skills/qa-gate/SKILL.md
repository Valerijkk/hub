---
name: qa-gate
description: Sole public QA authority for closed G0 routing; future Python jobs remain NOT_READY.
disable-model-invocation: true
---

# /qa-gate — sole public QA authority

`/qa-gate` is Hub's only public QA execution authority. A caller never invokes the private runner, Docker/Compose, a legacy runner, or an internal command directly.

Invoke `/qa-gate` for QA before handoff or PR. Until every acceptance gate listed below is complete, that public request returns `NOT_READY` without private-runner or Docker execution.

## Current status

- G0-A is static-contract material only: `STATIC_CONTRACT_REVIEWED / PRIVATE_CONTRACT_NOT_RUN`.
- C1 Cycle-6 scoped static review: `STATIC_SCOPED_REVIEWED`; whole-slice
  review: `PENDING`. Cycle-6 source:
  `D0932A7943367DFB07660B0937E79C1A90EB347D40FCFB892D280778F2B433DB`;
  Cycle-6 package:
  `8E00ADFA4E5C54AD023B76C267F6CA2DBBE34573C6ED0875AD9811EF0ECB32CA`.
- C1 execution status: `PS5.1_PARSE_NOT_RUN / PRIVATE_CONTRACT_NOT_RUN`.
- G0-C source documentation does not imply private-contract, image, Docker, target, runtime, or PASS acceptance.
- Launcher/package: `NOT_READY`.
- Runtime: `NOT_RUN`.
- Target results and PASS evidence: absent.

Until independent G0-B, G0-C, G0-S, G0-B01, Docker-only authority, and the later acceptance tasks are accepted, every otherwise valid G0 request returns `NOT_READY` before the private runner and Docker. Static review, source files, or image declarations do not remove this gate.

## Public request

Caller supplies exactly:

- `target` — one closed G0 target below;
- `task` — lowercase `[a-z0-9][a-z0-9-]{0,23}`;
- `run` — independently validated lowercase `[a-z0-9][a-z0-9-]{0,23}`.

`/qa-gate` injects literal contract `hub.qa-runner/v1`. The caller cannot supply or override a contract, executable, script path, arguments, Compose file, image, timeout profile, dependency, fixture, resource inventory, selector, project name, container name, or raw command.

Fixed derived identities are:

- Compose project: `hubqa-<task>-<run>`;
- container: `<project>-<target>`;
- report root: `testframework/reports/<task>-<run>/`.

## Closed routing table

Exactly three G0 targets may become executable only after every named gate is accepted:

| Public target | Fixed responsibility |
|---|---|
| `backend-unit` | baked backend build, vet, genuine unit, and race sequence |
| `frontend-unit` | baked frontend typecheck, lint, build, and genuine unit sequence |
| `backend-openapi` | baked deterministic two-pass OpenAPI generation and exact three-file comparison |

Declared future jobs are not G0 targets: `python-contract`, `python-api`, `python-integration`, `python-e2e-smoke`, `python-e2e-full`, and `python-load`. They return `NOT_READY` without invoking G0 or Docker. Missing, unknown, sentinel, test, private, malformed, or extra request input is rejected before G0 with intended exit `64`; no partial or best-effort dispatch exists.

## Validation and private dispatch boundary

Apply this order without side effects:

1. Require exactly `target`, `task`, and `run`.
2. Return `NOT_READY` for a declared future Python job; reject missing, extra, unknown, sentinel, test, and private input with intended exit `64`.
3. Validate `task` and `run` independently.
4. Construct the private request with exactly injected `contract`, validated `target`, `Task`, and `Run`.
5. Confirm the named G0 source/runtime gates. Until then, return `NOT_READY` without starting the private runner or Docker.
6. Only after those gates, hand the request internally to canonical private `testframework/runner/run-task.ps1` through fixed inbox Windows PowerShell 5.1 authority at `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`.

That internal handoff is not a public command and must not be printed as a copyable runner instruction. No alternate shell, `pwsh`, PATH/PATHEXT lookup, naked Docker/Compose fallback, legacy runner, host-test fallback, runtime image pull/build/install, ad-hoc build, automatic or in-run retry, grace period, or scheduler is permitted. The canonical fixed backend/frontend build jobs in the closed targets remain required.

## Fixed deadline, process, and cleanup contract

`/qa-gate` owns only routing and task/run identity. It never schedules, terminates, extends, automatically retries within a run, or reinterprets a deadline, and it returns the supervisor's immutable first-cause exit unchanged.

| Layer | Exclusive owner | Fixed bound and rule |
|---|---|---|
| fixed command | image-baked entrypoint | 180 seconds; image-owned |
| fixed target | image-baked entrypoint | 780 seconds; image-owned |
| image inspect/attestation | private supervisor | 30 seconds; distinct layer |
| fixed target outer process | private supervisor | 810 seconds |
| inventory, exact-ID cleanup, and zero proof | private supervisor | 30 seconds aggregate |

After report ownership, if image inspect/attestation reaches its 30-second expiry, it freezes `first_exit_code = 124` and `deadline_layer = inspect`, then skips `RunFixedTarget`, Compose target start, and Compose-environment start. It does not skip the mandatory sole aggregate 30-second cleanup `finally`: `ListOwnedResources` → pure exact owned/foreign selector → `CleanupOwnedResources` only when the complete initial inventory is clean → residual `ListOwnedResources` zero proof. If the fixed target outer process reaches its 810-second expiry, it likewise freezes first exit `124` before that same mandatory cleanup. Cleanup and reporting may record their own categorical outcomes but cannot overwrite frozen `124`.

At the first observed private process expiry, completion performs only best-effort termination of that exact handle, makes no further wait, drain, scheduling, read, progress, EOF, stream, or child-exit observation, and disposes the process and streams through the existing path. The bounds have no additive grace, automatic or in-run retry, or duplicate scheduler.

The private supervisor contract has exactly one physical Process constructor. It is shared only by four closed operations: `InspectLockedImage`, `RunFixedTarget`, `ListOwnedResources`, and `CleanupOwnedResources`. These are the only ordinary Docker operations, with at most ten Docker invocations in total; no general command abstraction exists. Cleanup is one closed, exact-ID, foreign-preserving sequence: inventory, typed owned/foreign selection, cleanup of the owned set only, then zero-owned containers, networks, and volumes proof.

## Evidence contract

Only the accepted private boundary may report task-owned evidence:

- `commands.json`: actual ordered per-command receipt sequence, which may stop at the first failed or timed-out command;
- for each actual command ID, closed receipts `commands/<id>.stdout.log`, `commands/<id>.stderr.log`, and `commands/<id>.result.json`;
- `run.json`: final immutable first-exit, deadline layer, outcome, and artifact manifest; its `commands` field is the fixed expected command-ID list; and
- `.owner.json`, target artifacts when produced, and `cleanup.json`.

The G0-B filename convention is future `NOT_VERIFIED` until independent G0-B review. It is neither G0-B acceptance evidence nor C1 approval. `run.json` lists only artifacts actually persisted and contained beneath the owned report root; it contains no environment dump or secret value.

## Result and triage

Report only the status and task-owned evidence produced by the accepted private boundary. Do not infer runtime or PASS from source, static review, files, image declarations, or an absent error.

If an authorized future target fails, do not immediately edit code or tests. Apply `.cursor/skills/test-triage/SKILL.md`, establish whether the product or test is wrong from current docs, and obtain any separately required correction/execution authority. A fresh explicit public `/qa-gate` request is then allowed; it is a new run, never a continuation or automatic retry of the failed run. Never fall back to a direct or legacy command.

## Related authority

- Docker-only and execution authority: `AGENTS.md` and `.cursor/REGULATION.md`.
- Failure classification: `.cursor/skills/test-triage/SKILL.md`.
- G0 contract and gates: `.superpowers/sdd/2026-08-09-minimal-qa-launcher/2026-08-09-minimal-qa-launcher.md`.
