---
name: qa-gate
description: Sole public QA authority for closed G0 routing; future Python jobs remain NOT_READY.
disable-model-invocation: true
---

# /qa-gate — sole public QA authority

`/qa-gate` is Hub's only public QA execution authority. A caller never invokes
the private runner, Docker/Compose, a legacy runner, or an internal command
directly. Until every named acceptance gate is complete, a public request
returns `NOT_READY` without private-runner or Docker execution.

## Current status

- `testframework/docker/images.lock.v3.json` and the required local dependency
  bytes are absent.
- The candidate image-lock schema is `hub.qa.local-images-lock/v3`.
- G0-B parity: `G0-B_PARITY_NOT_VERIFIED`.
- Package and launcher: `NOT_READY`.
- Private contract and runtime: `PRIVATE_CONTRACT_NOT_RUN` / `NOT_RUN`.
- Static source is accepted only; P7 and every execution acceptance gate remain
  `BLOCKED`.
- Static source and documentation do not imply image, Docker, target, runtime,
  cleanup, or PASS acceptance.

Until independent G0-B, G0-C, G0-S, G0-B01, Docker-only authority, and later
acceptance tasks are accepted, every otherwise valid G0 request returns
`NOT_READY` before the private runner and Docker.

## Public request

Caller supplies exactly:

- `target` — one closed G0 target below;
- `task` — lowercase `[a-z0-9][a-z0-9-]{0,23}`;
- `run` — independently validated lowercase `[a-z0-9][a-z0-9-]{0,23}`.

`/qa-gate` injects literal contract `hub.qa-runner/v1`. The caller cannot
supply or override a contract, executable, script path, arguments, Compose
file, image, timeout profile, dependency, fixture, resource inventory,
selector, project name, container name, or raw command.

Fixed derived identities are:

- Compose project: `hubqa-<task>-<run>`;
- container: `<project>-<target>`;
- report root: `testframework/reports/<task>-<run>/`.

## Closed routing table

Exactly three G0 targets may become executable only after every named gate is
accepted:

| Public target | Fixed responsibility |
|---|---|
| `backend-unit` | baked backend build, vet, genuine unit, and race sequence |
| `frontend-unit` | baked frontend typecheck, lint, build, and genuine unit sequence |
| `backend-openapi` | baked deterministic two-pass OpenAPI generation and exact three-file comparison |

Declared future jobs are not G0 targets: `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full`, and `python-load`.
They return `NOT_READY` without invoking G0 or Docker. Missing, unknown,
sentinel, test, private, malformed, or extra request input is rejected before
G0 with intended exit `64`; no partial or best-effort dispatch exists.

## Validation and private dispatch boundary

Apply this order without side effects:

1. Require exactly `target`, `task`, and `run`.
2. Return `NOT_READY` for a declared future Python job; reject missing, extra,
   unknown, sentinel, test, and private input with intended exit `64`.
3. Validate `task` and `run` independently.
4. Construct the private request with exactly injected `contract`, validated
   `target`, `Task`, and `Run`.
5. Confirm the named G0 source/runtime gates. Until then, return `NOT_READY`
   without starting the private runner or Docker.
6. Only after those gates, hand the request internally to the canonical private
   runner through fixed Windows PowerShell 5.1 authority.

That internal handoff is not a public command and must not be printed as a
copyable runner instruction. No alternate shell, PATH/PATHEXT lookup, naked
Docker/Compose fallback, legacy runner, host-test fallback, runtime image
pull/build/install, ad-hoc build, automatic or in-run retry, grace period, or
scheduler is permitted.

## Local image, control, and Compose boundary

The local-only `images.lock.v3.json` selects every base and target solely by
its complete local Docker `image_id`. A target `repo_digest` member rejects
the lock; observed target tags and RepoDigests are never selectors. A base
`repo_digest` is local-inventory/admission evidence only, checked only after
the exact local `image_id`, and never selects an image.

The v3 lock binds a literal reviewed `ControlCommit`, six mode-`100644`
control Git blobs, their closed control-bundle SHA-256, and the immutable raw
Compose-byte SHA-256. Product/source Git blobs are worktree-non-authoritative
provenance inputs, not a self-proof by the attestor. The exact local target
`image_id` plus its control pair and all eight mandatory provenance labels
must agree: target, base image ID, control commit, control bundle hash, source
commit, source-policy hash, source-tree-manifest hash, and build-manifest hash.
Entrypoint, Cmd, and complete labels projection remain separate values, each
hashable as no-BOM UTF-8 RFC 8785 JCS bytes.

The strict duplicate-aware parser rejects duplicate, missing, additional,
malformed, escaped, or reparsed lock members before report mutation or a
dependency call. An existing task-owned report path is never overwritten or
adopted.

All selected images must already exist locally as exact `linux/amd64` objects.
There is no registry login, pull, push, manifest lookup, remote provenance
lookup, remote cache, network download, package installation, mutable-tag
selection, or fallback image acquisition. Missing local identity, platform,
or required dependency bytes fails closed before staging, execution, or
publication.

The G0-B Compose asset remains an inert no-source-bind template until its
separate acceptance. The subsequent local contract may use `pull_policy: never`,
may bind a service only to its locked local target `image_id`, and may mount
only the task-owned report path. Repository/source binds, checkout,
working-tree reads, Git metadata inference, and runtime installation are
forbidden. There is no registry or network path. Raw Compose bytes are pinned
to the reviewed control value and rechecked before staging, execution, and
publication; YAML substring parsing or transformed bytes cannot replace that
identity check.

## Fixed deadline, dependency, and cleanup contract

`/qa-gate` owns only routing and task/run identity. It never schedules,
terminates, extends, automatically retries within a run, or reinterprets a
deadline, and it returns the supervisor's immutable first-cause exit unchanged.

| Layer | Exclusive owner | Fixed bound and rule |
|---|---|---|
| fixed command | image-baked entrypoint | 180 seconds; image-owned |
| fixed target | image-baked entrypoint | 780 seconds; image-owned |
| image inspect/attestation | private supervisor | 30 seconds; distinct layer |
| fixed target outer process | private supervisor | 810 seconds |
| inventory, exact-ID cleanup, and zero proof | private supervisor | 30 seconds aggregate |

The private supervisor has exactly four closed dependency operations:

1. `InspectLockedImage`
2. `RunFixedTarget`
3. `ListOwnedResources`
4. `CleanupOwnedResources`

There is no fifth dependency operation or general command abstraction. At the
first observed deadline, exit `124` freezes for that layer; there is no
additive grace period or retry. Cleanup remains mandatory but cannot overwrite
that immutable first exit.

Cleanup is a sole exact owned-resource route: initial inventory, typed exact
owned/foreign selection, exact cleanup only when the complete initial
inventory is clean, then residual zero-owned proof. Foreign records are
preserved. Recursive deletion, wildcard deletion, pruning, automatic staging
janitors, and deletion of a final report are forbidden.

## Evidence contract

Only the accepted private boundary may report task-owned evidence:

- `.owner.json`;
- `commands.json`, the actual ordered receipt sequence, which may stop at the
  first failure or timeout;
- closed per-command output and result artifacts;
- `run.json`, with immutable first exit, deadline layer, outcome, and artifact
  manifest;
- target artifacts when produced; and
- `cleanup.json`.

Raw Compose stdout and stderr remain only in their owned raw artifacts and are
not copied into structured receipts. Evidence contains neither a host
environment dump nor secret values. G0-B filename/baked-entrypoint parity is
future `NOT_VERIFIED`; it is not G0-B acceptance evidence or C1 approval.

## Result and triage

Report only the status and task-owned evidence produced by the accepted private
boundary. Do not infer runtime or PASS from source, static review, files, image
declarations, or an absent error.

If an authorized future target fails, apply `.cursor/skills/test-triage/SKILL.md`
before any correction. A fresh explicit public `/qa-gate` request is a new run,
never a continuation or automatic retry of the failed run.

## Related authority

- Docker-only and execution authority: `AGENTS.md` and `.cursor/REGULATION.md`.
- Failure classification: `.cursor/skills/test-triage/SKILL.md`.
- G0 contract and gates:
  `.superpowers/sdd/2026-08-09-minimal-qa-launcher/2026-08-09-minimal-qa-launcher.md`.
