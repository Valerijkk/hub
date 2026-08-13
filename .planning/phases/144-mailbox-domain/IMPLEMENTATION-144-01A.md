# Task 144-01A — static acceptance evidence

> **Status:** STATIC_ACCEPTED · RUNTIME_NOT_RUN · not product-accepted
> **Implementation commit:** `9c4aba438e8a65c67b4b375636f34416978f8775`
> **Fix commit / accepted head:** `43d421ef0fc86dbb98330cf68bfd398fc3b40b0d`
> **Reviewed semantic contract:** `b3fb6f483b1335c4ef9ec3efbb0ff49c6d94ac3d`

## Scope and provenance

- Authoring packet: `/root/mail_144_01a_terra` and its scoped fix turn.
- Requested and reported actual authoring route: exact `gpt-5.6-terra` / `xhigh`; no
  fallback.
- Independent exact review used `gpt-5.6-terra` / `high` in a distinct context.
- Owned product/test files:
  - `backend/internal/mail/mailbox_engine.go`
  - `backend/internal/mail/mailbox_types.go`
  - `backend/internal/mail/mailbox_errors.go`
  - `backend/tests/mail/mailbox_engine_contract_test.go`

## Accepted behavior

- one public engine-neutral `MailEngine` boundary covering capability/health,
  provisioning, folders/messages/blobs, incremental and enumeration pages, drafts,
  submissions, Sieve and spam-training outcomes;
- immutable value DTOs with explicit bounded page requests and typed lost-response
  reconciliation results;
- stable sentinel-backed error taxonomy that exposes retryability without provider text,
  identifiers, content or secrets;
- external-package compile/shape oracles for the full interface and result unions.

## Review lifecycle

The first independent review found that an adversarial unknown `EngineErrorKind` was copied
into `Error()`. Commit `43d421ef` replaced interpolation with a closed allowlist; unknown
kinds now return the fixed text `mail engine error`, with an adversarial oracle. A proposed
non-forgeable origin marker was rejected because trusted external adapters must construct
the public enum and Go has no package-friend boundary. The exact re-review returned PASS
with no remaining P0, P1 or P2.

Blob preallocation remains intentionally outside this value contract: a real adapter cannot
be accepted until the request/stream cap is frozen. That is recorded as a founder question,
not hidden as a completed safety guarantee.

## Exact accepted source blobs

| Path | Blob |
|---|---|
| `backend/internal/mail/mailbox_engine.go` | `069bedab47651d4aba66e349227d141e4c287974` |
| `backend/internal/mail/mailbox_types.go` | `40fe8093847ebdb66d9413098c9af05fa1095344` |
| `backend/internal/mail/mailbox_errors.go` | `beb90bba840280522d61a058f7979fe91ba43c90` |
| `backend/tests/mail/mailbox_engine_contract_test.go` | `a17b00d6613d9aebba878a2f18dabc49e604667f` |

## Verification verdict

- exact owned-path and source review: PASS;
- protected-path inventory: zero changes;
- hostile error-content and interface-shape oracles: present and independently reviewed;
- executable tests, build, lint, validators and Docker jobs: `NOT_RUN`;
- accepted Docker launcher: `NOT_READY`.

Docker Desktop being available does not authorize direct Docker/Compose execution. This
packet is statically accepted only; it is not runtime-verified, complete, stable or
production-ready.
