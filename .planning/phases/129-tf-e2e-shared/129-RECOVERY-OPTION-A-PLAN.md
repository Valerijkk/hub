---
phase: 129-tf-e2e-shared
plan: recovery-option-a
type: execute
wave: recovery
depends_on: [recovery-docs-4fffc564b05d55a176e33cafcb9630f1cda68ffd]
status: PLAN_COMMIT_READY_ONLY__PREGENESIS_G0_ACTIVATION_BLOCKS_PACKAGE_A
autonomous: false
requirements: [TASK-204, TASK-219]
files_modified:
  - .planning/phases/129-tf-e2e-shared/129-RECOVERY-OPTION-A-PLAN.md
must_haves:
  truths:
    - "v1 remains byte-identical at SHA-256 85dd85ca38923a0211eff6110bef797274289c2ef453ec66364118100dff7bd5 and all 173 batches remain planned."
    - "The successor covers only the 34 Phase 129-04 deletion batches and exact 241 staged paths; preserved bulk moves never become receipts."
    - "G0 activation is a separately accepted immutable PRE-GENESIS input. Package A embeds its literal path, raw bytes digest, and fields; only that accepted input derives the recovery execution gate READY."
    - "Every deletion has one mutation commit, then later owner receipt and independent verdict; no later batch advances before prior PASS."
  prohibitions:
    - "No cutover-manifest mutation, fabricated v1 receipt/verdict, all-241 deletion commit, push, or merge."
    - "No host test/build/validator, direct Docker/Compose, pro-qa-ai access, or unrelated staging."
---

<objective>
Create a forward-only, fail-closed successor authority for staged Phase 129-04
legacy deletions without altering immutable v1 history or fabricating evidence.
It records the observed 241-path index state as future serial work only; it
does not delete, implement the runner, or claim runtime evidence.

This recovery plan is the only artifact permitted to persist now. The separate
`129-G0-ASSETS-RUNNER-PLAN.md` must first exist, pass independent review, and
produce a committed accepted G0 activation artifact containing literal accepted
digests and values. G0 activation is immutable PRE-GENESIS input, not a
recovery event. Only then may Package A create genesis/schema/docs by copying
those exact values, with no placeholders. Package B is TDD only after Package
A. Package C is validator implementation only after B's accepted red evidence.
Until those gates, the deletion cohort remains staged/uncommitted/unaccepted
and Phase 129 remains `RECOVERY_BLOCKED`.
</objective>

<execution_context>
@AGENTS.md
@.cursor/REGULATION.md
@.cursor/rules/07-agent-safety.mdc
@.cursor/rules/50-git.mdc
@.cursor/rules/55-docs-sync.mdc
@.planning/audits/test-rebase/RECOVERY-OPTION-A-2026-08-14.md
@.planning/audits/test-rebase/cutover-manifest.json
@.planning/audits/test-rebase/validate_cutover_manifest.py
@.planning/phases/129-tf-e2e-shared/129-04-PLAN.md
@.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md
@testframework/PACKAGE-CONTRACT.md
@testframework/runner/README.md
</execution_context>

## Fixed recovery baseline

| Datum | Required value |
|---|---|
| Recovery docs commit | `4fffc564b05d55a176e33cafcb9630f1cda68ffd`; forward record, not a receipt. |
| Recovered committed baseline | `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`. |
| Backend bulk move | `794ae3ccb72c75a39b18d90ef4d39c683ac31eff`, `OBSERVED_RECOVERY_UNACCEPTED`. |
| Frontend bulk move | `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`, `OBSERVED_RECOVERY_UNACCEPTED`. |
| Immutable v1 | `.planning/audits/test-rebase/cutover-manifest.json`, SHA-256 `85dd85ca38923a0211eff6110bef797274289c2ef453ec66364118100dff7bd5`, 173 all `planned`. |
| Staged cohort | 241 paths in the 34 v1 `129-04-*` batches, orders 46--79; ordered path-list SHA-256 `a31d03efe9939f23d6b266fbfb5454f071f0529827d125e4d9e89779b780b180`. |

Genesis copies the literal v1 IDs, paths, order and counts plus literal values
from the committed accepted G0 activation artifact. The counts in order are
exactly `8,15,12,6,12,4,7,5,10,10,2,2,8,1,7,5,4,8,15,2,4,9,11,3,7,9,1,9,5,6,9,17,4,4`.
The successor is additive: neither v1 bytes nor its historical states,
receipts, or verifier evidence changes.

## Closed authority and validator boundary

Package A creates canonical UTF-8/no-BOM immutable genesis/schema and
event/schema, and consumes the single shared G0 repository-snapshot schema and
producer at `testframework/schemas/repository-snapshot.schema.v1.json`. Later
facts are one immutable recovery event file per commit and G0-produced
snapshots. The implementation API is exactly:

```text
validate_bundle(genesis_path, v1_path, g0_activation_path, events_dir, snapshots_dir) -> ValidationResult
```

There is no `policy_path`, policy file, or live repository input. Genesis
closes all path policy, expected batch contracts, G0 input digests, transition
rules, roles and snapshot rules. Validator code is stdlib-only and neither
calls Git/shell/Docker/Compose/network/runner nor reads a live worktree/index.

### Immutable genesis contract

The top-level keys are exactly:

```text
schema_version, producer, authority_id, authority_kind, created_from,
immutable_v1, observed_recovery, index_genesis_projection, execution_gate,
g0_activation_contract, batches, event_contract, snapshot_contract,
authority_digest_rule
```

All objects use `additionalProperties: false`. `schema_version` is
`hub.phase129.recovery-successor-genesis/v1`; `authority_kind` is
`forward-recovery-deletion-genesis`; `producer` has `name`, `version`,
`contract_sha256`. `created_from` has `recovery_docs_commit`,
`recovered_baseline_commit`, `recovered_baseline_tree`, `source_revision`,
`created_at_utc`; `immutable_v1` has `path`, `sha256`, `batch_count`,
`all_batch_states`; observed moves have `commit_sha`, `tree_sha`, `status`.
`index_genesis_projection` has `delete_path_count`, `ordered_paths_sha256`,
`orders`, `paths`. Each `batches[]` entry has `id`, `order`, `source_plan`,
`task_id`, `operation_count`, `paths`, `paths_sha256`.

`execution_gate` is immutable historical truth derived only from the accepted
PRE-GENESIS G0 activation record and starts exactly:

```json
{"g0_activation":"ACCEPTED","g0_runner_dependency":"READY","ci_root_discovery":"UNBOUND","post_cutover_runtime":"NOT_RUN","deletion_execution":"READY"}
```

`READY` authorizes only the narrow serial deletion protocol below. It does not
accept historic cutover batches, bind CI root discovery, or claim runtime
evidence.

`authority_digest_rule` is `sha256(exact_utf8_bytes)` and no artifact
self-hashes. `event_contract` closes event types/file naming/transition table,
serial rule, rollback terminality, role requirements. `snapshot_contract`
closes the shared G0 schema path/digest, producer name/version, accepted-G0
requirements, and required kinds.

`g0_activation_contract` has exactly the top-level/path/value mapping copied
from `.planning/audits/test-rebase/g0-activation.v1.json`:

```text
schema, activation_id, g0_plan_path, required_plan_commit, founder_release_path,
founder_release_sha256, founder_release_commit, repository_commit,
bootstrap_contract, bootstrap_contract_path, bootstrap_contract_sha256,
runner_contract, runner_contract_path, runner_contract_sha256,
dependency_bundle_manifest_path, dependency_bundle_manifest_sha256,
source_policy_path, source_policy_sha256, build_manifest_path,
build_manifest_sha256, images_lock_path, images_lock_sha256,
target_entrypoints, planning_selftest, repository_snapshot_schema_path,
repository_snapshot_schema_sha256, run_envelope_schema_path,
run_envelope_schema_sha256, repository_snapshot_producer_path,
repository_snapshot_producer_sha256, run_envelope_producer_path,
run_envelope_producer_sha256, report_cleanup_contract, acceptance_verdict,
acceptance_commit, created_at_utc
```

Package A copies every field as a literal value and also records the activation
path plus SHA-256 of its exact raw UTF-8 bytes. In particular,
`founder_release_path`, `founder_release_sha256`, and
`founder_release_commit` are copied from the accepted activation record without
reconstructing a release mapping; they must identify the exact earlier
Package-R artifact that C consumed. The validator reads no live Git: it checks
that raw activation bytes match the embedded digest and that the literal mapping
exactly matches the immutable genesis. Package A is invalid if the activation
schema, acceptance verdict, Package-R path/digest/commit binding, byte digest,
or any copied value is absent, placeholder, guessed, or drifted.

`repository_commit` is the post-bootstrap **C1 lock commit**: the separately
committed forward package that contains the generated accepted
`images.lock.v3.json` bytes after bootstrap. It is neither the earlier source
package baseline nor the later activation-record commit. Every
`target_entrypoints` child has exactly `local_image_id`,
`local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`,
`command_sha256`, `compose_service`, and `runtime_status`; the first two names
and values match the corresponding record in `images.lock.v3.json` exactly.
`image_id` and `image_id_sha256` are forbidden aliases. The activation schema,
fixture, and validator must reject a wrong C1 commit, a wrong target set, an
alias, or any lock mapping drift.

### Shared G0 repository snapshots and Git-less evidence

Recovery creates no duplicate repository-snapshot schema or producer. Every
snapshot is produced and validated against the single frozen G0 contract at
`testframework/schemas/repository-snapshot.schema.v1.json`. Its top-level and
child contract is used without recovery-side aliases or additions:

```text
schema, producer, snapshot_id, captured_for, task, run, envelope_id,
role_claim, self_attested, source_revision, head, index, worktree, commits,
ancestry, control, evidence, lifecycle
```

Recovery permits only these `captured_for.kind` values:

```text
GENESIS_BASELINE, OWNER_MUTATION_AFTER, OWNER_RECEIPT_AFTER,
INDEPENDENT_VERDICT_AFTER, ROLLBACK_MUTATION_AFTER, ROLLBACK_RECEIPT_AFTER,
ROLLBACK_VERDICT_AFTER
```

`captured_for` has exactly these child fields, with no recovery-specific alias
or extension:

```text
kind, ordinal, subject_path, subject_sha256, subject_commit_sha,
predecessor_event_path, predecessor_event_sha256, predecessor_event_commit_sha,
pre_event_snapshot_path, pre_event_snapshot_sha256
```

Recovery consumes the same shared G0 `captured_for` semantics without alias or
extension:

| `kind` | `ordinal` | `subject_path` / `subject_sha256` | `subject_commit_sha` | predecessor event triple | pre-event snapshot pair |
|---|---:|---|---|---|---|
| `G0_BOOTSTRAP` | `0` | task-owned bootstrap `run.json` path/digest | `null` (report not committed) | `null` | `null` |
| `GENESIS_BASELINE` | `0` | committed `recovery-successor-genesis.json` path/digest | genesis commit | `null` | `null` |
| `OWNER_MUTATION_AFTER` | batch event ordinal | `null` (mutation is a commit, not an event file) | deletion mutation commit | preceding event or `null` for first batch | genesis baseline snapshot for first batch or preceding event's after-snapshot thereafter |
| `OWNER_RECEIPT_AFTER` | current | owner-receipt event path/digest | containing commit | preceding event or `null` for first owner receipt | current `OWNER_MUTATION_AFTER` snapshot |
| `INDEPENDENT_VERDICT_AFTER` | current | independent-verdict event path/digest | containing commit | current owner-receipt event | `OWNER_RECEIPT_AFTER` snapshot |
| `ROLLBACK_MUTATION_AFTER` | batch event ordinal | `null` (mutation is a commit, not an event file) | rollback mutation commit | FAIL independent-verdict event | FAIL `INDEPENDENT_VERDICT_AFTER` snapshot |
| `ROLLBACK_RECEIPT_AFTER` | current | rollback-receipt event path/digest | containing commit | FAIL verdict event | `ROLLBACK_MUTATION_AFTER` snapshot |
| `ROLLBACK_VERDICT_AFTER` | current | rollback-verdict event path/digest | containing commit | rollback receipt event | `ROLLBACK_RECEIPT_AFTER` snapshot |

In the table, a `null` predecessor triple or pre-event snapshot pair means every
child field in that triple or pair is JSON `null`. All non-null path, digest,
and commit values follow the exact repository-relative path, no-BOM raw-bytes
SHA-256, and lowercase 40-hex commit rules of the shared G0 contract.

The canonical bundle-relative report paths are deterministic and exhaustive:

```text
snapshots/baseline.repository-snapshot.json
snapshots/batch-<two-digit-batch-order>-<batch-id>/ordinal-<four-digit-event-ordinal>-<captured_for.kind>.repository-snapshot.json
```

The baseline is the first path. A mutation-after snapshot uses the ordinal of
the immediately following receipt event. Every after-event snapshot uses the
ordinal of the event it attests. Thus one ordinal can have both a mutation
snapshot and its following receipt-after snapshot, but their required `kind`
produces distinct canonical paths. A snapshot may reference only already
written event bytes and prior snapshot bytes by these repository-relative
paths/digests; it never names a future event or snapshot. Required snapshots
are one baseline, one owner-mutation snapshot per batch mutation, one
after-event snapshot per owner receipt and independent verdict, plus the
corresponding rollback-mutation, receipt, and verdict snapshots when the
failed suffix path is used.

The shared snapshot producer is accepted G0 process evidence only. The
validator checks internal schema/digests/causality/index projection, not
cryptographic or human/organization/model/actor/Git/repository identity proof.

### Causal events and serial deletion

Each event has exactly `schema_version`, `event_id`, `event_ordinal`,
`event_type`, `batch_id`, `batch_order`, `predecessor_event_path`,
`predecessor_event_sha256`, `pre_event_snapshot_path`,
`pre_event_snapshot_sha256`, `mutation_commit_sha`, `role_assertion`,
`evidence`, `post_event_index_projection`, `created_from`.
`role_assertion` is `run_id`, `envelope_id`, `role_claim`, `self_attested`.
`evidence` is `task_report_path`, `task_report_sha256`,
`focused_evidence_sha256`, `root_evidence_sha256`, `verdict`.
`pre_event_snapshot_path` and `pre_event_snapshot_sha256` are the event's only
snapshot reference and must identify a pre-existing canonical snapshot with
matching raw bytes; they never identify the event's own after-event snapshot.
`post_event_index_projection` is `ordered_staged_delete_paths`,
`staged_delete_count`, `staged_delete_paths_sha256`.

G0 activation is immutable PRE-GENESIS input and is not an event, event state,
snapshot kind, fixture, or recovery commit. Genesis embeds its exact path, raw
bytes digest, and literal fields; only that accepted input derives the recovery
execution gate READY. No owner mutation is valid before genesis has bound the
accepted activation record.

Events start at ordinal 1. Before the first event, the owner creates a
deletion-only mutation commit and its `OWNER_MUTATION_AFTER` snapshot. The
first lawful event is then `OWNER_RECEIPT`, bound to that pre-existing mutation
commit and snapshot. After every event commit, the G0 shared producer creates
a separate canonical after-event snapshot at the exact path above. That
snapshot attests the event path, event raw-bytes SHA-256, containing commit,
predecessor event triple, and the event's pre-event snapshot pair. The next
event, if any, names that already-existing snapshot through its own
`pre_event_snapshot_*` pair. The only paths are:

```text
deletion mutation -> OWNER_MUTATION_AFTER -> OWNER_RECEIPT -> OWNER_RECEIPT_AFTER
-> INDEPENDENT_VERDICT(PASS) -> INDEPENDENT_VERDICT_AFTER

deletion mutation -> OWNER_MUTATION_AFTER -> OWNER_RECEIPT -> OWNER_RECEIPT_AFTER
-> INDEPENDENT_VERDICT(FAIL) -> INDEPENDENT_VERDICT_AFTER
-> rollback mutation -> ROLLBACK_MUTATION_AFTER -> ROLLBACK_RECEIPT
-> ROLLBACK_RECEIPT_AFTER -> ROLLBACK_INDEPENDENT_VERDICT(PASS)
-> ROLLBACK_VERDICT_AFTER -> terminal
```

An event references only prior event bytes/snapshot and pre-existing mutation
commits. The later after-event snapshot attests event bytes and containing
commit. The next event may then reference that snapshot. Final verdict
acceptance requires its separate final after-event snapshot and an independent
review; no event, snapshot, or activation can self-reference or name a future
SHA. Validator `E_SNAPSHOT_EVENT_LINK` checks, for every event: (1) the
pre-event path is canonical, exists before the event, and hashes to
`pre_event_snapshot_sha256`; (2) the resolved snapshot's `captured_for` is a
lawful prior state; (3) when an event has a predecessor, that prior snapshot's
`subject_path`, `subject_sha256`, and `subject_commit_sha` attest exactly that
predecessor event and its containing commit; (4) the required separately
written after-event snapshot has the canonical path/kind/ordinal and attests
the current event path, bytes, and containing commit; and (5) its predecessor
triple and pre-event snapshot pair mirror the current event's causal inputs.
For mutation-after snapshots, the same code requires null subject path/digest,
the mutation commit as `subject_commit_sha`, and the lawful predecessor/prior
snapshot link. The validator also rejects missing, duplicate, self, future,
noncanonical, wrong-kind, wrong-ordinal, wrong-batch, or wrong-digest links.
It separately checks the containing commit's changed path list/digest,
parent/tree values, ancestry edges and projection; no self-SHA is needed.

An owner mutation contains exactly and only current batch deletion paths;
receipt/verdict commits contain none. Genesis projection is all 241. After
owner receipt/PASS it is only later-batch paths. After FAIL, a rollback commit
must restore the failed current batch cleanly; its `ROLLBACK_MUTATION_AFTER`
snapshot therefore projects only the later-batch suffix, never the restored
current batch. That terminal state cannot retry the current batch: retry
requires a new successor genesis. Any missing/duplicate/ordinal gap, wrong predecessor,
wrong parent/tree/changed path, two in-flight batches, later-before-PASS,
projection drift or post-terminal event fails closed.

Independent verdict operational evidence requires a distinct accepted-G0
`run_id`, distinct `envelope_id`, `role_claim: independent-verifier` instead
of owner, `self_attested: false`, and later event commit/ancestry binding.
This is evidence-level operational independence only, not cryptographic proof
of actor/human/org/model identity. `E_INDEPENDENCE` covers same run, same
envelope, wrong role, and self-attested cases.

## Tasks

<tasks>
<task type="docs">
  <name>Package A — successor authority/docs</name>
  <owner>tech-lead</owner><verifier>code-reviewer</verifier>
  <files>.planning/audits/test-rebase/recovery-successor-genesis.schema.json, .planning/audits/test-rebase/recovery-successor-genesis.json, .planning/audits/test-rebase/recovery-successor-event.schema.json, .planning/audits/test-rebase/RECOVERY-SUCCESSOR-AUTHORITY.md, .planning/audits/test-rebase/RECOVERY-OPTION-A-2026-08-14.md, .planning/audits/test-rebase/README.md, .planning/audits/test-rebase/VERIFICATION.md, .planning/STATE.md, .planning/ROADMAP.md</files>
  <precondition>The separately authored `.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md` exists, is independently accepted, and `.planning/audits/test-rebase/g0-activation.v1.json` exists as its committed accepted immutable activation record. It supplies every literal contract path/digest/value and binds the single shared `testframework/schemas/repository-snapshot.schema.v1.json` plus producer. Currently false.</precondition>
  <action>Only after that precondition, create immutable schema/genesis/docs using the exact accepted G0 activation values and raw bytes digest. Consume, but do not duplicate, the G0 repository-snapshot schema/producer. Preserve `STATIC_PATH_MAPPING=PASS`, `CUTOVER_ACCEPTANCE=FAIL/BLOCKED`, `LAUNCHER=NOT_READY`, `POST_CUTOVER_RUNTIME=NOT_RUN`, `PHASE_129=RECOVERY_BLOCKED`; do not create successor events/snapshots/validator/tests/fixtures, runner, CI change or deletion.</action>
  <verify>Read-only v1 SHA/all-planned and cached 241-path digest/path checks; Package A paths must not intersect cached deletions. No validator/test.</verify>
  <done>Independent review and a literal Package A-only commit allowlist; then Package B remains blocked until Package A acceptance.</done>
</task>
<task type="tdd-blocked">
  <name>Package B — exhaustive failing bundle tests</name>
  <owner>qa-auto1</owner><verifier>qa-auto-ver</verifier>
  <precondition>G0 plan/static packages/bootstrap/evidence/activation are independently accepted and Package A genesis/schema/docs are independently accepted; valid immutable G0 activation record exists through accepted `/qa-gate` planning-selftest. Currently false.</precondition>
  <files>.planning/audits/test-rebase/test_validate_recovery_successor_authority.py, .planning/audits/test-rebase/fixtures/recovery-successor-authority/**</files>
  <action>One canonical valid bundle and descriptors below. Its immutable G0 activation contains the literal accepted `founder_release_path`, `founder_release_sha256`, and `founder_release_commit` mapping as well as every other copied field. `materialize_bundle(valid_bundle, descriptor, tmp_path)` writes canonical genesis/v1/immutable-G0-activation/events/snapshots directory bytes, applies exactly one descriptor mutation, preserves all untouched bytes, returns `(genesis_path, v1_path, g0_activation_path, events_dir, snapshots_dir)`.</action>
  <verify>Only accepted runner `/qa-gate` target after G0; current `NOT_RUN`.</verify>
</task>
<task type="implementation-blocked">
  <name>Package C — fail-closed validator</name>
  <owner>qa-auto2</owner><verifier>qa-auto-ver + code-reviewer</verifier>
  <precondition>Package B accepted red evidence through G0; currently false.</precondition>
  <files>.planning/audits/test-rebase/validate_recovery_successor_authority.py</files>
  <action>Implement only stdlib read-only `validate_bundle(genesis_path, v1_path, g0_activation_path, events_dir, snapshots_dir) -> ValidationResult`, checking every named fixture failure. No Git/shell/container/network/live-index access.</action>
  <verify>Only accepted `/qa-gate` selftest; valid passes and every negative emits exact code.</verify>
</task>
<task type="handoff">
  <name>Package D — G0 then later one-batch plans</name>
  <owner>tech-lead</owner><verifier>qa-auto-ver</verifier>
  <action>The separately authored G0 plan must have its own per-package literal allowlists and deliver a committed, independently accepted immutable activation output artifact with literal accepted bootstrap contract, dependency bundle/digest, source policy, build manifest, images lock, target entrypoints, outer runner, planning selftest, shared snapshot/envelope schemas and producers, reports/cleanup, and every activation value. Only then may Package A begin. A later plan owns exactly one deletion mutation/receipt/verdict/snapshot.</action>
  <done>STOP after this plan's independent review and its separate planning commit. Then enforce the order reviewed-plan commit -> accepted G0 plan/static packages/bootstrap/evidence/activation -> Package A -> Package B -> Package C -> later one-batch deletion. No all-241 or parallel deletion is authorized.</done>
</task>
</tasks>

## Package B exhaustive fixtures

`valid-bundle.json` is canonical and test `test_accepts_valid_bundle` passes.
Every negative file is a single mutation descriptor with exactly `schema`,
`target`, `operation`, `value`, `expected_error`; `schema` equals
`hub.phase129.recovery-successor-fixture-mutation/v1`; `target` is a
bundle-relative JSON path or virtual file; `operation` is `replace`, `delete`,
`append`, `duplicate`, `reorder`, `malform`, or `noncanonical`; `value` is JSON
or literal malformed bytes; `expected_error` is stable `E_*`.

| Fixture | Test function | Error |
|---|---|---|
| `invalid-malformed-json.json` | `test_rejects_malformed_json` | `E_JSON_MALFORMED` |
| `invalid-noncanonical-json.json` | `test_rejects_noncanonical_json` | `E_JSON_CANONICAL` |
| `invalid-unknown-key.json` | `test_rejects_unknown_key` | `E_SCHEMA_UNKNOWN_KEY` |
| `invalid-v1-digest.json` | `test_rejects_v1_digest_drift` | `E_V1_DIGEST` |
| `invalid-v1-state.json` | `test_rejects_v1_not_all_planned` | `E_V1_STATE` |
| `invalid-batch-missing.json` | `test_rejects_missing_batch` | `E_BATCH_PARTITION` |
| `invalid-batch-duplicate.json` | `test_rejects_duplicate_batch` | `E_BATCH_PARTITION` |
| `invalid-batch-order.json` | `test_rejects_reordered_batch` | `E_BATCH_ORDER` |
| `invalid-path-digest.json` | `test_rejects_path_digest_mismatch` | `E_PATH_DIGEST` |
| `invalid-initial-projection.json` | `test_rejects_initial_genesis_projection` | `E_INDEX_PROJECTION` |
| `invalid-prefix-projection.json` | `test_rejects_per_prefix_pass_projection` | `E_INDEX_PROJECTION` |
| `invalid-fail-rollback-projection.json` | `test_rejects_fail_rollback_projection` — rejects any rollback-after projection that restages the restored current batch instead of the later-batch suffix | `E_INDEX_PROJECTION` |
| `invalid-snapshot-schema-producer.json` | `test_rejects_snapshot_schema_or_producer` | `E_SNAPSHOT_CONTRACT` |
| `invalid-snapshot-missing-required.json` | `test_rejects_missing_required_snapshot` | `E_SNAPSHOT_REQUIRED` |
| `invalid-snapshot-duplicate-captured-for.json` | `test_rejects_duplicate_snapshot_captured_for` | `E_SNAPSHOT_DUPLICATE` |
| `invalid-snapshot-event-link.json` | `test_rejects_bad_snapshot_to_event_link` — rejects a self/future/missing/noncanonical pre-event snapshot, wrong raw digest, unlawful prior captured-for state, mutation snapshot with non-null subject bytes, or an absent/mismatched deterministic after-event snapshot | `E_SNAPSHOT_EVENT_LINK` |
| `invalid-event-bytes-commit.json` | `test_rejects_event_bytes_or_containing_commit_binding` | `E_EVENT_BINDING` |
| `invalid-event-missing-file.json` | `test_rejects_missing_required_event_file` | `E_EVENT_FILE_REQUIRED` |
| `invalid-event-duplicate.json` | `test_rejects_duplicate_event_or_ordinal` | `E_EVENT_DUPLICATE` |
| `invalid-mutation-metadata.json` | `test_rejects_mutation_parent_tree_or_changed_paths` | `E_MUTATION_COMMIT` |
| `invalid-ancestry.json` | `test_rejects_ancestry` | `E_ANCESTRY` |
| `invalid-worktree-scope.json` | `test_rejects_worktree_scope` | `E_WORKTREE_SCOPE` |
| `invalid-g0-activation-digest.json` | `test_rejects_g0_activation_digest` | `E_G0_ACTIVATION_DIGEST` |
| `invalid-g0-activation-schema.json` | `test_rejects_g0_activation_schema` — rejects `image_id` aliases, missing `local_image_id`/`local_image_id_sha256`, or a target set that differs from the v3 image lock | `E_G0_ACTIVATION_SCHEMA` |
| `invalid-g0-activation-acceptance.json` | `test_rejects_g0_activation_acceptance` | `E_G0_ACTIVATION_ACCEPTANCE` |
| `invalid-g0-activation-path-binding.json` | `test_rejects_g0_activation_path_binding` — rejects a `repository_commit` other than the post-bootstrap C1 lock commit or a target-entrypoint value that drifts from that commit's v3 image lock | `E_G0_ACTIVATION_PATH_BINDING` |
| `invalid-g0-activation-founder-release-binding.json` | `test_rejects_g0_activation_founder_release_binding` — rejects a missing, reordered, wrong-path, wrong-raw-digest, or wrong-commit `founder_release_path`/`founder_release_sha256`/`founder_release_commit` mapping | `E_G0_ACTIVATION_PATH_BINDING` |
| `invalid-event-predecessor.json` | `test_rejects_event_predecessor` | `E_EVENT_CHAIN` |
| `invalid-event-ordinal.json` | `test_rejects_event_ordinal` | `E_EVENT_ORDER` |
| `invalid-later-before-pass.json` | `test_rejects_later_batch_before_pass` | `E_SERIAL_ORDER` |
| `invalid-two-inflight.json` | `test_rejects_two_inflight_batches` | `E_SERIAL_ORDER` |
| `invalid-owner-allowlist.json` | `test_rejects_owner_allowlist` | `E_MUTATION_ALLOWLIST` |
| `invalid-receipt.json` | `test_rejects_owner_receipt` | `E_OWNER_RECEIPT` |
| `invalid-independence-same-run.json` | `test_rejects_independence_same_run` | `E_INDEPENDENCE` |
| `invalid-independence-same-envelope.json` | `test_rejects_independence_same_envelope` | `E_INDEPENDENCE` |
| `invalid-independence-role.json` | `test_rejects_independence_wrong_role` | `E_INDEPENDENCE` |
| `invalid-independence-self-attested.json` | `test_rejects_independence_self_attested` | `E_INDEPENDENCE` |
| `invalid-verdict-transition.json` | `test_rejects_verdict_transition` | `E_VERDICT_TRANSITION` |
| `invalid-rollback-binding.json` | `test_rejects_rollback_binding` | `E_ROLLBACK_BINDING` |
| `invalid-terminal-event.json` | `test_rejects_terminal_event` | `E_TERMINAL` |

## Exact commit templates and STOP

This plan is persisted/reviewed before Package A and is separate from it:

```text
git commit --only .planning/phases/129-tf-e2e-shared/129-RECOVERY-OPTION-A-PLAN.md -m "plan(test-rebase): harden Option A recovery authority"
```

The separately authored G0 plan must itself be committed and independently
accepted, with its static packages, bootstrap, evidence, and immutable
`g0-activation.v1.json` output accepted by literal allowlists, before this
Package A-only command is permitted:

```text
git commit --only .planning/audits/test-rebase/recovery-successor-genesis.schema.json .planning/audits/test-rebase/recovery-successor-genesis.json .planning/audits/test-rebase/recovery-successor-event.schema.json .planning/audits/test-rebase/RECOVERY-SUCCESSOR-AUTHORITY.md .planning/audits/test-rebase/RECOVERY-OPTION-A-2026-08-14.md .planning/audits/test-rebase/README.md .planning/audits/test-rebase/VERIFICATION.md .planning/STATE.md .planning/ROADMAP.md -m "docs(test-rebase): establish successor recovery authority"
```

Package B commits exactly test source and every fixture in this list:
`valid-bundle.json`, `invalid-malformed-json.json`, `invalid-noncanonical-json.json`,
`invalid-unknown-key.json`, `invalid-v1-digest.json`, `invalid-v1-state.json`,
`invalid-batch-missing.json`, `invalid-batch-duplicate.json`,
`invalid-batch-order.json`, `invalid-path-digest.json`,
`invalid-initial-projection.json`, `invalid-prefix-projection.json`,
`invalid-fail-rollback-projection.json`, `invalid-snapshot-schema-producer.json`,
`invalid-snapshot-missing-required.json`, `invalid-snapshot-duplicate-captured-for.json`,
`invalid-snapshot-event-link.json`, `invalid-event-bytes-commit.json`,
`invalid-event-missing-file.json`, `invalid-event-duplicate.json`,
`invalid-mutation-metadata.json`, `invalid-ancestry.json`, `invalid-worktree-scope.json`,
`invalid-g0-activation-digest.json`, `invalid-g0-activation-schema.json`,
`invalid-g0-activation-acceptance.json`, `invalid-g0-activation-path-binding.json`,
`invalid-g0-activation-founder-release-binding.json`, `invalid-event-predecessor.json`,
`invalid-event-ordinal.json`, `invalid-later-before-pass.json`,
`invalid-two-inflight.json`, `invalid-owner-allowlist.json`, `invalid-receipt.json`,
`invalid-independence-same-run.json`, `invalid-independence-same-envelope.json`,
`invalid-independence-role.json`, `invalid-independence-self-attested.json`,
`invalid-verdict-transition.json`, `invalid-rollback-binding.json`,
`invalid-terminal-event.json`.

The following is the authoritative Package B literal `git commit --only`
allowlist and supersedes any illustrative template below. Omission of any path
is a blocked package:

```text
git commit --only .planning/audits/test-rebase/test_validate_recovery_successor_authority.py .planning/audits/test-rebase/fixtures/recovery-successor-authority/valid-bundle.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-malformed-json.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-noncanonical-json.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-unknown-key.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-v1-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-v1-state.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-missing.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-duplicate.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-order.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-path-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-initial-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-prefix-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-fail-rollback-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-schema-producer.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-missing-required.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-duplicate-captured-for.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-event-link.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-bytes-commit.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-missing-file.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-duplicate.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-mutation-metadata.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-ancestry.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-worktree-scope.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-schema.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-acceptance.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-path-binding.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-predecessor.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-ordinal.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-later-before-pass.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-two-inflight.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-owner-allowlist.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-receipt.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-same-run.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-same-envelope.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-role.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-self-attested.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-verdict-transition.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-rollback-binding.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-terminal-event.json -m "test(test-rebase): specify successor authority failures"
```

Both displayed Package-B commands are amended in-place by inserting exactly
`.planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-founder-release-binding.json`
immediately after `invalid-g0-activation-path-binding.json`. That path is part
of the same atomic Package-B allowlist; omission is `BLOCKED`, and it must not
be committed separately.

```text
git commit --only .planning/audits/test-rebase/test_validate_recovery_successor_authority.py .planning/audits/test-rebase/fixtures/recovery-successor-authority/valid-bundle.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-malformed-json.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-noncanonical-json.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-unknown-key.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-v1-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-v1-state.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-missing.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-duplicate.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-batch-order.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-path-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-initial-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-prefix-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-fail-rollback-projection.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-schema-producer.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-missing-required.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-duplicate-captured-for.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-snapshot-event-link.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-bytes-commit.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-missing-file.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-duplicate.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-mutation-metadata.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-ancestry.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-worktree-scope.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-digest.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-schema.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-acceptance.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-g0-activation-path-binding.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-predecessor.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-event-ordinal.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-later-before-pass.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-two-inflight.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-owner-allowlist.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-receipt.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-same-run.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-same-envelope.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-role.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-independence-self-attested.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-verdict-transition.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-rollback-binding.json .planning/audits/test-rebase/fixtures/recovery-successor-authority/invalid-terminal-event.json -m "test(test-rebase): specify successor authority failures"
git commit --only .planning/audits/test-rebase/validate_recovery_successor_authority.py -m "feat(test-rebase): add fail-closed recovery authority validator"
```

Before and after each allowed commit, read-only inspection proves the cached cohort remains exactly
241 paths at `a31d03efe9939f23d6b266fbfb5454f071f0529827d125e4d9e89779b780b180`.
No `git add -A`, reset, restore, clean, checkout, rebase, all-241 commit, push
or merge. Run only `git diff --check` and `git status --short` for this plan.
STOP after independent plan review and this separate planning commit. Valid G0
activation is mandatory before Package A or any later package/deletion. The
enforced order is: reviewed-plan commit -> accepted G0 plan/static
packages/bootstrap/evidence/activation -> Package A genesis/schema/docs ->
Package B red tests -> Package C validator -> later one-batch deletion. No
later package may start early.
