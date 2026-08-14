# Task 144-03A-Address — static acceptance evidence

> **Status:** `STATIC_ACCEPTED` · `EXACT_SHA_PASS` · `RUNTIME_NOT_RUN`
> **Atomic commit:** `41564d003946ddb78df18efca166ca86e4055405`; parent `d659c67ce657fb33da391819edbdb954454be4b9`; tree `13a97333f08df6946e65766d3223c1ede85e1f50`
> **Subject and scope:** `feat(mail): add bounded address policy`; exactly five paths.

> **Mail lifecycle update:** 144-02D docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` + transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; contrary combined 02D/02F-next text below is historical. Sole next action: three exact 02F closure reviews; Rule/Phase145/runtime blockers remain.

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
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review
  target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. 144-02D is docs-only accepted; 144-02F's three exact closure reviews, later plan and
  Phase-145 carrier remain. No execution. Its dependency on accepted
  Address is satisfied; its callable/encoding/oracle contract is
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` with C/I/M `0/0/0`.
  No Folder production/test work is authorized before the gates stated above.
- The task-owned launcher is `NOT_READY`; compile, unit, race, build, lint, validators,
  Docker, and runtime are all `NOT_RUN`. No direct host or Docker substitute is valid evidence.
- `144-01B` remains `BLOCKED`; broader Phase 142/cross-contract acceptance and Phase 148 are
  pending. This evidence does not accept the phase, product, persistence, OpenAPI, or runtime.

> **Current Rule gate:** docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. 144-02D is docs-only accepted; accepted 02F and the later separate Rule plan/Phase-145 carrier remain required before Rule TDD/production.
