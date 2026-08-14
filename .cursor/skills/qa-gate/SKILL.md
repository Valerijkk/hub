---
name: qa-gate
description: Sole public QA authority for the closed v2 target/task/run route; Python jobs remain NOT_READY.
disable-model-invocation: true
---

# /qa-gate — sole public QA authority

`/qa-gate` is Hub's only public QA execution authority. Its adapter accepts the
exact raw argument vector `-Target <target> -Task <task> -Run <run>`, derives
the request fields `target`, `task`, and `run`, then injects contract
`hub.qa-runner/v2`. A caller must never invoke a private runner, bootstrap,
Docker/Compose command, legacy runner, or internal command directly.

Current status is `NOT_READY`: package and launcher are `NOT_READY`, private
contract execution is `PRIVATE_CONTRACT_NOT_RUN`, and runtime evidence is
`NOT_RUN`. Source review and static contracts are not Docker, image, cleanup,
or test acceptance.

## Closed public v2 route

The public request contains exactly:

- `target`
- `task`, matching `^[a-z][a-z0-9-]{2,31}$`
- `run`, matching `^[a-z0-9][a-z0-9-]{7,63}$`

The caller cannot choose a contract, private context, bootstrap role, owner
nonce, source/ingress/report path, executable, argument vector, image,
Compose file, timeout, dependency, environment, resource inventory, or any
other selector. Missing, extra, unknown, sentinel, test, private, or malformed
input rejects with intended exit `64` before a private runner or Docker action.
There is no v1 bridge, compatibility runner, best-effort dispatch, retry, or
fallback.

The fixed G0 target identities are:

| Target | Current public result |
|---|---|
| `backend-unit` | `NOT_READY` pending accepted C1 authority |
| `frontend-unit` | `NOT_READY` pending accepted C1 authority |
| `backend-openapi` | `NOT_READY` pending accepted C1 authority |
| `g0-planning-selftest` | reject: bootstrap-private fixture only |

The future Python targets `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full`, and `python-load`
return `NOT_READY` without entering a private runner or Docker. They are not
G0 fallback targets.

## Bootstrap is not a public route

`hub.qa-bootstrap/v1` is only the later founder-released, local-only
transaction. It derives fixed `g0-assets-runner` task/run values from the
accepted Package-R artifact and invokes only the internal
`g0-planning-selftest` closure. A direct script call, a public planning-selftest
request, serialized role/private-context transport, a continuation, or a
duplicate closure call rejects before Docker.

The bootstrap-private admission context has the frozen 28-field A0.1 archive
shape as amended by A0.2. Each A1 asset has exactly the ordered fields
`asset_id`, `targets`, `relative_path`, `sha256`, `bytes`, `archive_format`,
`member_count`, `expanded_bytes`, `layout_sha256`, `license_spdx`,
`license_evidence_sha256`, and `provenance_sha256`.
Its separate post-build image-attestation context has the frozen 11-field
shape. Both are in-memory-only internal capability data; neither is public
input or runtime evidence.

The Package-B candidate holds every accepted private-source file with
read-only/no-write/no-delete sharing while importing and rechecking its exact
bytes. Before any admission or attestor entry point, it installs and verifies
B-owned replacements for the imported Docker-client resolver, Git
blob/tree/source-tree primitives, and admission Docker primitive. Their active
call graph reaches Git/Docker only through the bounded Windows Job Object
runner, which creates each child suspended, assigns it to a kill-on-close Job,
and resumes it only afterwards.
Inert historical process text in the accepted A0.2/A0.1 blobs is not a callable
fallback.

The envelope producer is imported before the snapshot producer. The snapshot
producer itself has no dot-source, Git, or process path: bootstrap captures the
exact 16-field Git/control evidence through the bounded runner before sealing
command receipts, requiring HEAD to equal the release commit before and after
the ordered index/worktree capture. It performs exact cleanup, then passes a
closed 24-field context of captured values and retained receipt bytes to the
pure producer.

## Docker-only execution boundary

Windows host execution of tests, validators, linters, build/vet, Vite,
backend/frontend processes, pytest, Playwright, or Locust is forbidden. All
future execution happens only in task-owned Docker after the required authority
is accepted. Direct Docker/Compose snippets, registry/network access, source
binds, runtime installs, legacy launchers, and direct private-runner invocation
are not fallbacks.

The later local-only bootstrap uses the admitted immutable archive topology
`bootstrap/build-context/<target>/bindings/<relative_path>`, performs the
Git-object source-tree gate before target-image builds, and retains evidence in
one direction: snapshot first, envelope second with the only raw link to the
snapshot, and `run.json` third as the pair aggregator. It preserves the first
non-zero exit, freezes a deadline as `124`, and cleans only exact owned
resources while preserving foreign resources and retained reports.

All process waits are bounded: Git/inspection `30s`, individual command `180s`,
one cumulative target budget `780s`, cleanup `30s`, and outer budget `840s`
including `30s` orchestration slack. Every child wait is clipped to the
remaining applicable budgets; timeout maps to `124` without overriding an
earlier non-zero exit.

The active outer guard fails closed when its runtime Stopwatch is missing or
stopped and compares elapsed ticks with the exact 840-second limit. It runs
before and after snapshot Git evidence, command-receipt sealing, exact cleanup,
snapshot context/snapshot, envelope, `run.json`, and the C chain. Command
receipts intentionally precede cleanup. A final success fence in `finally`
runs before the outer Stopwatch stops. The stopped elapsed ticks are checked
again before runtime authority is cleared; either expiry maps to `124`.

The 30-second cleanup budget is cumulative active cleanup time: helper cleanup
starts and stops the shared Stopwatch in `finally`, and exact cleanup later
restarts and stops that same Stopwatch. Intervening builds and inspections do
not age the cleanup budget. Both cleanup layers immediately rethrow
`E_B0_TIMEOUT_124`; success and fallback cleanup catches retain exit `124`
instead of general cleanup exit `71`. Exact cleanup applies the combined
outer-plus-cleanup guard around resource enumeration and every iteration,
owned file/directory deletion, Docker probe/removal, report-tree removal, and
cleanup-receipt creation. One unconditional post-delete guard may also serve as
the terminal iteration fence. Exact cleanup finally stops and rechecks the
cumulative cleanup clock against 30 seconds, then checks the still-live outer
clock.

Immediately after Package-R validation and before report-root/runtime creation,
bootstrap scans the fixed physical nonce ledger and create-new writes
`testframework/reports/g0-assets-runner-nonce-consumptions/<release_id>.nonce-consumption.v1.json`.
The nine-field `CONSUMED` receipt is durable, raw-digest bound to R, and never
cleanup-owned. After positive-green the bootstrap writes canonical
`commands/command-receipts.v1.json` with top-level `schema`, `producer`, `task`,
`run`, `owner_marker_sha256`, `release_one_shot_nonce`, `commands`, and
`created_at_utc`. Each command is exactly `name`, `argv`, `exit_code`,
`timed_out`, `started_at_utc`, `finished_at_utc`, `stdout_sha256`, and
`stderr_sha256`. The exact ordered 16-key C chain is `nonce_consumption`,
`owner_marker`, `local_base_inspect`, `helper_base_verification`,
`helper_image_attestation`, `ingress_copy`, `helper_admission`,
`helper_cleanup`, `source_tree_verification`, `local_image_inspection`,
`candidate_lock`, `planning_selftest_negative_red`,
`planning_selftest_positive_green`, `command_receipts`, `run`, and `cleanup`;
each child is `{path, raw_sha256}`.
The ten-field completed-zero `run.json` binds that nonce and the raw
snapshot/envelope path-digest pairs; C1/D/activation bind the same receipt and
ordered `a0_commit`, `a02_commit`, `a1_commit`, `b_commit` authority.

## Current blockers

- A1 is `BLOCKED_EXTERNAL_INGRESS`: exact offline dependency bytes and their
  approved provenance/license bindings have not been admitted.
- Package B remains a static candidate until its independent package acceptance
  is recorded; static review does not constitute private-contract execution.
- Package R is `BLOCKED`: it needs accepted A0.1, the exact nine-path A0.2
  successor, A1, and Package B plus the founder's explicit one-shot local-only
  approval.
- Until Package R and the later C/C1/D acceptance chain exist, `/qa-gate`
  returns `NOT_READY` before private execution or Docker.

When a future accepted target fails, use
`.cursor/skills/test-triage/SKILL.md` before deciding whether code or the test
is at fault. Do not infer a pass from static source, an image declaration, a
missing error, or zero collection.

## Planned routing declarations

These declarations preserve future Python ownership only. They do not create
or authorize a direct runner or Docker route.

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
