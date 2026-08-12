---
name: qa-gate
description: Sole public QA authority for closed G0 routing; planned Python jobs remain NOT_READY.
disable-model-invocation: true
---

# /qa-gate — sole public QA authority

`/qa-gate` is Hub's only public QA execution authority. A caller never invokes
the private runner, Docker/Compose, a legacy runner, or an internal command
directly. Until every named acceptance gate is complete, a public request
returns `NOT_READY` without private-runner or Docker execution.

## Current admission state

- `testframework/docker/images.lock.v3.json` and the required local dependency
  bytes are absent.
- The candidate image-lock schema is `hub.qa.local-images-lock/v3`.
- G0-B parity is `G0-B_PARITY_NOT_VERIFIED`.
- Package and launcher are `NOT_READY`; private contract and runtime are
  `PRIVATE_CONTRACT_NOT_RUN` / `NOT_RUN`.
- Static source is accepted only. P7 and every execution acceptance gate remain
  `BLOCKED`; static review, files, or documentation are not image, Docker,
  target, runtime, cleanup, or PASS acceptance.

Until independent G0-B, G0-C, G0-S, G0-B01, Docker-only authority, and later
acceptance tasks are accepted, every otherwise valid G0 request returns
`NOT_READY` before the private runner and Docker.

## Founder and planning boundaries

- Windows host execution of tests, validators, linters, build/vet, Vite,
  backend/frontend processes, pytest, Playwright, or Locust is forbidden.
  The host is only for reading/editing, Git, and Docker CLI.
- Legacy `backend/testframework/**`, static-contract Vitest, TypeScript
  Playwright, their configs/linkers/runners, and direct Docker snippets are
  neither a gate nor a fallback.
- Historical 82% Go coverage and row parity are not rebase acceptance gates.
- Zero collection, an unexpected skip, or a leaked task-owned resource is a
  failure for a future accepted target; it does not authorise a legacy run.

The target catalog below describes the founder-greenfield end state only. It
does not grant execution authority or change the current `NOT_READY` /
`NOT_RUN` state.

| Catalog job | Planned target | Current routing state |
|---|---|---|
| `backend-unit` | `backend/tests/{domain}/**` | closed G0 candidate only |
| `frontend-unit` | `frontend/tests/{domain}/**`, support | closed G0 candidate only |
| `backend-openapi` | baked deterministic OpenAPI comparison | closed G0 candidate only |
| `python-contract` | `testframework/tests/contract/**` | future, `NOT_READY` |
| `python-api` | `testframework/tests/api/**` | future, `NOT_READY` |
| `python-integration` | `testframework/tests/integration/**` | future, `NOT_READY` |
| `python-e2e-smoke` | `testframework/tests/e2e/**`, marker `smoke` | future, `NOT_READY` |
| `python-e2e-full` | `testframework/tests/e2e/**`, launcher profile `full` | future, `NOT_READY` |
| `python-load` | `testframework/load/**` | future, `NOT_READY` |

## Public request and closed G0 routing

Caller supplies exactly:

- `target` — one closed G0 target;
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

Exactly these three G0 targets may become executable only after every named
gate is accepted:

| Public target | Fixed responsibility |
|---|---|
| `backend-unit` | baked backend build, vet, genuine unit, and race sequence |
| `frontend-unit` | baked frontend typecheck, lint, build, and genuine unit sequence |
| `backend-openapi` | baked deterministic two-pass OpenAPI generation and exact three-file comparison |

Declared Python catalog jobs are not G0 targets. They return `NOT_READY`
without invoking G0 or Docker. Missing, unknown, sentinel, test, private,
malformed, or extra input is rejected before G0 with intended exit `64`; no
partial or best-effort dispatch exists.

Apply validation without side effects in this order:

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
copyable runner instruction. Alternate shells, PATH/PATHEXT lookup, naked
Docker/Compose fallback, host-test fallback, runtime image pull/build/install,
ad-hoc build, automatic or in-run retry, grace period, and scheduler are
forbidden.

## Local image, control, and Compose boundary

The local-only `images.lock.v3.json` selects every base and target solely by
its complete local Docker `image_id`. A target `repo_digest` member rejects the
lock; observed target tags and RepoDigests are never selectors. A base
`repo_digest` is local-inventory/admission evidence only, checked only after
the exact local `image_id`, and never selects an image.

The v3 lock binds a literal reviewed `ControlCommit`, six mode-`100644` control
Git blobs, their closed control-bundle SHA-256, and the immutable raw
Compose-byte SHA-256. Product/source Git blobs are worktree-non-authoritative
provenance inputs, not a self-proof by the attestor. The exact local target
`image_id` plus its control pair and all eight mandatory provenance labels must
agree: target, base image ID, control commit, control bundle hash, source
commit, source-policy hash, source-tree-manifest hash, and build-manifest hash.
Entrypoint, Cmd, and complete labels projection remain separate values, each
hashable as no-BOM UTF-8 RFC 8785 JCS bytes.

The strict duplicate-aware parser rejects duplicate, missing, additional,
malformed, escaped, or reparsed lock members before report mutation or a
dependency call. An existing task-owned report path is never overwritten or
adopted.

All selected images must already exist locally as exact `linux/amd64` objects.
Registry login, pull, push, manifest lookup, remote provenance lookup, remote
cache, network download, package installation, mutable-tag selection, and
fallback image acquisition are forbidden. Missing local identity, platform, or
required dependency bytes fails closed before staging, execution, or
publication.

The G0-B Compose asset remains an inert no-source-bind template until separate
acceptance. A subsequent local contract may use `pull_policy: never`, bind a
service only to its locked local target `image_id`, and mount only the
task-owned report path. Repository/source binds, checkout, working-tree reads,
Git metadata inference, and runtime installation are forbidden. Raw Compose
bytes are pinned to the reviewed control value and rechecked before staging,
execution, and publication; YAML substring parsing or transformed bytes cannot
replace that identity check.

The private candidate reserves report staging with one nonce-bound create-new
record. Before atomic publication, staging is exactly `.owner.json` plus the
raw-byte-pinned `compose.unit.yaml`; `raw/` is initialized only after verified
publication and reservation release, before any dependency construction.
Collision or ambiguity fails leave-only: no object is adopted or deleted, no
dependency starts, and no fallback route is permitted.

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
owned/foreign selection, exact cleanup only when the complete initial inventory
is clean, then residual zero-owned proof. Foreign records are preserved.
Recursive deletion, wildcard deletion, pruning, automatic staging janitors,
and deletion of a final report are forbidden.

## Evidence and triage

Only the accepted private boundary may report task-owned evidence:

- `.owner.json`;
- `commands.json`, the actual ordered receipt sequence, which may stop at the
  first failure or timeout;
- closed per-command output and result artifacts;
- `run.json`, with immutable first exit, deadline layer, outcome, and artifact
  manifest;
- target artifacts when produced; and
- `cleanup.json`.

Raw Compose stdout and stderr remain only in owned raw artifacts and are not
copied into structured receipts. Evidence contains neither a host environment
dump nor secret values. G0-B filename/baked-entrypoint parity is future
`NOT_VERIFIED`; it is not G0-B acceptance evidence or C1 approval.

Report only the status and task-owned evidence produced by the accepted private
boundary. Do not infer runtime or PASS from source, static review, files, image
declarations, or an absent error. If an authorized future target fails, apply
`.cursor/skills/test-triage/SKILL.md` before any correction. A fresh explicit
public `/qa-gate` request is a new run, never a continuation or automatic retry.

## Planned routing declarations

The following declarations preserve target ownership only. Their referenced
paths are future and non-executable until the relevant accepted phase creates
them; they never authorise a direct runner or Docker invocation.

<!-- hub.qa-routing-declaration/v1
{"schema_version":"hub.qa-routing-declaration/v1","scenario_id":"api-health","assigned_skill":".cursor/skills/qa-gate/SKILL.md","declared_routes":["testframework/src/hub_testkit/domains/health/api.py","testframework/tests/api/health/test_health.py"],"expected_routes":["testframework/src/hub_testkit/domains/health/api.py","testframework/tests/api/health/test_health.py"],"prohibited_legacy_tokens":["testframework/api/**","testframework/core/**"]}
-->
<!-- hub.qa-routing-declaration/v1
{"schema_version":"hub.qa-routing-declaration/v1","scenario_id":"integration-postgres","assigned_skill":".cursor/skills/qa-gate/SKILL.md","declared_routes":["testframework/src/hub_testkit/integration/postgres.py","testframework/tests/integration/postgres/test_transactions.py"],"expected_routes":["testframework/src/hub_testkit/integration/postgres.py","testframework/tests/integration/postgres/test_transactions.py"],"prohibited_legacy_tokens":["testframework/core/**","testframework/integration/**"]}
-->
<!-- hub.qa-routing-declaration/v1
{"schema_version":"hub.qa-routing-declaration/v1","scenario_id":"e2e-login","assigned_skill":".cursor/skills/qa-gate/SKILL.md","declared_routes":["testframework/src/hub_testkit/browser/lifecycle.py","testframework/src/hub_testkit/domains/identity/ui/components/session_form.py","testframework/src/hub_testkit/domains/identity/ui/pages/login.py","testframework/tests/e2e/identity/test_login_smoke.py"],"expected_routes":["testframework/src/hub_testkit/browser/lifecycle.py","testframework/src/hub_testkit/domains/identity/ui/components/session_form.py","testframework/src/hub_testkit/domains/identity/ui/pages/login.py","testframework/tests/e2e/identity/test_login_smoke.py"],"prohibited_legacy_tokens":["testframework/core/**","testframework/e2e/**","testframework/e2e/pages/**","testframework/e2e/support/**"]}
-->
<!-- hub.qa-routing-declaration/v1
{"schema_version":"hub.qa-routing-declaration/v1","scenario_id":"load-tracker","assigned_skill":".cursor/skills/qa-gate/SKILL.md","declared_routes":["testframework/load/profiles/steady.toml","testframework/load/users/tracker.py"],"expected_routes":["testframework/load/profiles/steady.toml","testframework/load/users/tracker.py"],"prohibited_legacy_tokens":["testframework/core/**"]}
-->

## Related authority

- Docker-only and execution authority: `AGENTS.md` and `.cursor/REGULATION.md`.
- Failure classification: `.cursor/skills/test-triage/SKILL.md`.
- G0 contract and gates:
  `.superpowers/sdd/2026-08-09-minimal-qa-launcher/2026-08-09-minimal-qa-launcher.md`.
