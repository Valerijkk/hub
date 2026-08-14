# Task 144-03A-Address — static acceptance evidence

> **Status:** `STATIC_ACCEPTED` · `EXACT_SHA_PASS` · `RUNTIME_NOT_RUN`
> **Atomic commit:** `41564d003946ddb78df18efca166ca86e4055405`; parent `d659c67ce657fb33da391819edbdb954454be4b9`; tree `13a97333f08df6946e65766d3223c1ede85e1f50`
> **Subject and scope:** `feat(mail): add bounded address policy`; exactly five paths.

## Scope and provenance

- Authoring route: requested exact `gpt-5.6-terra` / `xhigh`; accepted native dispatch,
  `provenance_source: dispatcher-enforced`, fallback `none`; actual model and reasoning
  telemetry: `UNKNOWN` unless exposed.
- Independent static reviews: requested exact `gpt-5.6-terra` / `high` in three distinct
  contexts; accepted native dispatch, `provenance_source: dispatcher-enforced`, fallback
  `none`; actual model and reasoning telemetry: `UNKNOWN` unless exposed.
- Documentation sync route: requested `gpt-5.6-terra` / `medium`,
  `provenance_source: dispatcher-enforced`; actual telemetry: `UNKNOWN` unless exposed.
- Owned committed code/test paths, outside this evidence document, are
  `backend/internal/mail/address_policy.go` and
  `backend/tests/mail/address_policy_test.go`.

## Exact candidate blobs

| Path | Committed Git blob |
|---|---|
| `backend/internal/mail/address_policy.go` | `fe7472d6a5bf9baa140fef95352cbbf206f49612` |
| `backend/tests/mail/address_policy_test.go` | `67f35c39dfbbfdd92573f52103798e15de922b37` |

The parent and tree identify the immutable accepted boundary. The code/test blobs above are the
committed blobs reviewed for that exact commit.

## Static-conformant behavior

The candidate validates UTF-8 email input, rejects raw values above 320 bytes before IDNA,
forbids Unicode whitespace/control and header injection, requires exactly one `@`, and permits
only ASCII RFC 5322 `atext` dot-atom local parts of 1..64 bytes. It converts the domain through
`idna.Lookup.ToASCII`, ASCII-lowercases it, enforces a 1..253-byte domain with valid 1..63-byte
lowercase ASCII labels, then returns a lowercase canonical email no longer than 320 bytes.

Display names are UTF-8, control-free, bounded to 160 runes and 640 bytes, and preserved exactly
without trim, NFC, collapsing, or case folding. Set normalization admits a configured maximum of
1..200, preserves caller input and result order, produces a concrete non-nil empty success,
rejects overcount and canonical duplicates, and yields no partial result on any failure. It
accepts `.test`, `.invalid`, and synthetic `u-*` fixtures without DNS or network access.

Only `invalid`, `too_large`, and `duplicate` are public address error kinds; their typed
sentinels and text are stable and content-free. This packet does not change OpenAPI or
persistence.

## Independent static reviews

Three independent exact-commit Terra/high reviews reported `EXACT_SHA_PASS`:

| Review | Result | C/I/M |
|---|---|---|
| Oracle conformance | `EXACT_SHA_PASS` | `0/0/0` |
| Specification conformance | `EXACT_SHA_PASS` | `0/0/0` |
| Security review | `EXACT_SHA_PASS` | `0/0/0` |

All report Spec/Quality/Security/Evidence PASS. These are exact-commit static reviews only; they
do not establish executable, phase, product, or runtime acceptance.

## Remaining gates and next action

- This packet is `STATIC_ACCEPTED`; it is not executable, phase, product, or runtime `PASS`.
- Thread is narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`; C/I/M `0/0/0`), static only.
- Folder source+oracle are narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692`
  (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Their exact 20-path docs candidate awaits three exact-SHA reviews;
  after PASS root only freezes/adjudicates `144-03B-Rule`. No execution. Its dependency on accepted
  Address is satisfied; its callable/encoding/oracle contract is
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` with C/I/M `0/0/0`.
  No Folder production/test work is authorized before the gates stated above.
- The task-owned launcher is `NOT_READY`; compile, unit, race, build, lint, validators,
  Docker, and runtime are all `NOT_RUN`. No direct host or Docker substitute is valid evidence.
- `144-01B` remains `BLOCKED`; broader Phase 142/cross-contract acceptance and Phase 148 are
  pending. This evidence does not accept the phase, product, persistence, OpenAPI, or runtime.
> **144-03B Folder post-source lifecycle (2026-08-14):** source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692` (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only, with exactly `backend/internal/mail/folder_tree.go` blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b` and `backend/tests/mail/folder_tree_test.go` blob `efe0db0071cb22ce4b026c635157f0fb08885d64`. Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high exact-SHA reviews are `EXACT_SHA_PASS`, C/I/M `0/0/0`. This exact 20-path docs candidate (the frozen prior 19-path lifecycle manifest plus its new evidence file, including the three stale-gate evidence documents) awaits its own Lifecycle/Evidence high, Security/Dependency xhigh and Exact/Boundary high exact-SHA reviews. Launcher `NOT_READY`; compile, unit, race, build, lint, validators, Docker and runtime `NOT_RUN`; 04B reparent, 144-01B, Phase 142 and Phase 148 remain blocked. Author provenance Terra/medium, `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. After candidate reviews PASS, root only freezes/adjudicates `144-03B-Rule`; no Rule TDD or production authorization.
