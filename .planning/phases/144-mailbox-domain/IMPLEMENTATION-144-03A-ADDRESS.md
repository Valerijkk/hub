# Task 144-03A-Address — pre-commit static conformance evidence

> **Status:** `STATIC_CONFORMANT` · `EXACT_ATOMIC_COMMIT_INDEPENDENT_REVIEW_PENDING` · `RUNTIME_NOT_RUN`
> **Frozen parent:** `d659c67ce657fb33da391819edbdb954454be4b9`
> **Self SHA:** unknowable until the atomic commit is created; no exact-commit verdict exists yet.

## Scope and provenance

- Authoring route: requested exact `gpt-5.6-terra` / `xhigh`; accepted native dispatch,
  `provenance_source: dispatcher-enforced`, fallback `none`; actual model and reasoning
  telemetry: `UNKNOWN` unless exposed.
- Independent static reviews: requested exact `gpt-5.6-terra` / `high` in three distinct
  contexts; accepted native dispatch, `provenance_source: dispatcher-enforced`, fallback
  `none`; actual model and reasoning telemetry: `UNKNOWN` unless exposed.
- Documentation sync route: requested `gpt-5.6-terra` / `medium`,
  `provenance_source: dispatcher-enforced`; actual telemetry: `UNKNOWN` unless exposed.
- Owned code/test candidate paths, outside this evidence document, are
  `backend/internal/mail/address_policy.go` and
  `backend/tests/mail/address_policy_test.go`.

## Exact candidate blobs

| Path | Candidate Git blob |
|---|---|
| `backend/internal/mail/address_policy.go` | `fe7472d6a5bf9baa140fef95352cbbf206f49612` |
| `backend/tests/mail/address_policy_test.go` | `67f35c39dfbbfdd92573f52103798e15de922b37` |

The candidate is intentionally pre-commit. The parent above is evidence of the starting
boundary, not a substitute for the future atomic commit SHA or an exact-SHA review.

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

Three independent exact-blob reviews of the candidate code and oracle reported PASS:

| Review | Result | C/I/M |
|---|---|---|
| Oracle conformance | PASS | `0/0/0` |
| Specification conformance | PASS | `0/0/0` |
| Security review | PASS | `0/0/0` |

These are exact-blob static reviews only. They establish neither an exact atomic commit verdict
nor executable, phase, product, or runtime acceptance.

## Remaining gates and next action

- Required next action: create the exact atomic commit for `144-03A-Address`, then perform an
  independent exact-SHA review of that commit.
- Until that review passes, the lifecycle remains `STATIC_CONFORMANT` and
  `EXACT_ATOMIC_COMMIT_INDEPENDENT_REVIEW_PENDING`; it is not `STATIC_ACCEPTED` or `PASS`.
- The task-owned launcher is `NOT_READY`; compile, unit, race, build, lint, validators,
  Docker, and runtime are all `NOT_RUN`. No direct host or Docker substitute is valid evidence.
- `144-01B` remains `BLOCKED`; broader Phase 142/cross-contract acceptance and Phase 148 are
  pending. This evidence does not accept the phase, product, persistence, OpenAPI, or runtime.
