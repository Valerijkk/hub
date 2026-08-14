# Task 144-01A — static conformance evidence

> **Status:** `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` · `EXACT_SHA_PASS` · `RUNTIME_NOT_RUN`
> **Frozen contract:** Mail 0.1.8 144-01A entry amendment at `5e314fb080d420696642058ae47397e251c23207`; Thread contract remains accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`, static only
> **Atomic boundary:** `e3e8c774d734f926f77f8e600b58dc68dabbd993`; parent
> `fa0a2f2a542caadd9f5951e6f703b060b4776780`.

> **Mail lifecycle update:** correction commit `fcb5b84fed4a417e736ddb4053256e494a0c25f6` (parent `55e36b7a1a3516d07b2baf390a5742a24ec8fbe5`, tree `5a0637f2306ff42295adea5560793e2a8f4a294a`, subject `docs(mail): correct receipt and rule invariants`, exact 28 docs) is `COMMITTED_UNACCEPTED_AT_FCB5B84F` / `EXACT_SHA_TRANSITION_REQUIRED` / `IMPLEMENTATION_BLOCKED`. Contract Terra/high and Security Terra/xhigh PASS `0/0/0`; Lifecycle Terra/high FAIL `0/1/0` solely stale committed precommit lifecycle; dispatcher-enforced, no fallback, telemetry `UNKNOWN`. Former correction/precommit/root-commit-next instructions are `CONSUMED/HISTORICAL`. Next: transition review, exact 28-doc transition commit without self-SHA, Lifecycle exact-SHA re-review of fcb plus transition and trace Contract/Security, then separate 02F docs-only sync only after all PASS; 144-01B/Rule/Phase145/runtime blockers remain.

## Scope and provenance

- Authoring route: exact `gpt-5.6-terra` / `xhigh`, fallback `none`, accepted native
  dispatch, `provenance_source: dispatcher-enforced`; actual model and reasoning telemetry:
  `UNKNOWN`.
- Independent review route: exact `gpt-5.6-terra` / `high` in distinct contexts, fallback
  `none`, accepted native dispatch, `provenance_source: dispatcher-enforced`; actual model
  and reasoning telemetry: `UNKNOWN`.
- The four owned product/test paths are:
  - `backend/internal/mail/mailbox_engine.go`
  - `backend/internal/mail/mailbox_types.go`
  - `backend/internal/mail/mailbox_errors.go`
  - `backend/tests/mail/mailbox_engine_contract_test.go`

## Static-conformant 0.1.8 boundary

The atomic seven-path boundary implements the closed 0.1.8 structural port only. It replaces hydrated list,
change and enumeration pages with `EngineMessageProjection` pages; moves page cursor ownership
to the exact request/page values; adds account-bound full-message/blob byte-budget requests;
and keeps `EngineMessage` as one projection plus one authorized body.

The atomic boundary freezes the exact projection, attachment, body, draft and reconciliation field
grammar; adds the 1 MiB projection-item, 8 MiB projection-page, 50 MiB message and 25 MiB
attachment hard limits and their related collection/string limits; and makes submission the
six-field reference-only value for a confirmed draft. It adds the closed body-format,
priority and attachment-disposition enums, and preserves all three exact projection-origin
literals, including `confirmed_owner_draft`.

`ErrEnginePayloadTooLarge` and `EngineErrorPayloadTooLarge` are literal-safe: the kind is in
both `EngineError.Error` and `EngineError.Unwrap` allowlists, maps only to `mail engine payload
too large`, and unknown kinds still return `mail engine error` and unwrap to `nil`.

## Atomic-boundary source blobs

| Path | Git blob in atomic boundary |
|---|---|
| `backend/internal/mail/mailbox_engine.go` | `fcb32e22413af56e3c74d26064edad7098765fac` |
| `backend/internal/mail/mailbox_types.go` | `3d1f91226eec2612fe9a4a83ebef03fc66c5514e` |
| `backend/internal/mail/mailbox_errors.go` | `328daf1e141dab9ea5ffe589227192c8de05ebf7` |
| `backend/tests/mail/mailbox_engine_contract_test.go` | `51b42a7933e4e2f8ab6da7c2830bdd23526d1c2b` |

The external-package oracle statically checks method conformance, exact field count/order/type,
absence of body/bytes in projections and content in submissions, literal limits/enums, and
`errors.Is` for the payload sentinel. It is source evidence only until a supported current-SHA
Docker job exists.

## Review lifecycle

The exact-commit independent Terra/high review returned `EXACT_SHA_PASS`: Spec/Quality/Security/
Evidence PASS, C/I/M `0/0/0`. Requested routes are dispatcher-enforced and actual telemetry is
`UNKNOWN`; `UNKNOWN` alone is not a blocker for an accepted native dispatch. Any exposed mismatch,
rejection, omitted model/effort or fallback remains blocking.

The 0.1.7 implementation commit `9c4aba438e8a65c67b4b375636f34416978f8775`, fix commit
`43d421ef0fc86dbb98330cf68bfd398fc3b40b0d`, and semantic contract
`b3fb6f483b1335c4ef9ec3efbb0ff49c6d94ac3d` are historical and superseded for this packet.
The historical fix correctly changed adversarial unknown error kinds to the closed literal
fallback, but it did not implement the 0.1.8 projection/budget/reference-only amendment.

| Historical superseded path | Historical Git blob |
|---|---|
| `backend/internal/mail/mailbox_engine.go` | `069bedab47651d4aba66e349227d141e4c287974` |
| `backend/internal/mail/mailbox_types.go` | `40fe8093847ebdb66d9413098c9af05fa1095344` |
| `backend/internal/mail/mailbox_errors.go` | `beb90bba840280522d61a058f7979fe91ba43c90` |
| `backend/tests/mail/mailbox_engine_contract_test.go` | `a17b00d6613d9aebba878a2f18dabc49e604667f` |

## Verification and remaining gates

- Source status: `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993`; owned-path,
  exact-field, safe-error and oracle reviews are static only.
- Accepted task-owned launcher: `NOT_READY`.
- Compile, unit, race, build, lint, validators, Docker and runtime: all `NOT_RUN`.
- Docker Desktop availability does not authorize a direct Docker/Compose substitute.
- The packet is not runtime-verified, complete, stable or production-ready.
- `144-01B` remains `BLOCKED` under stricter Task3 authority until its applicable supported
  current-SHA Docker evidence closes; Phase 145 remains independently gated.
- `144-03A-Address` is separately `STATIC_ACCEPTED` at
  `41564d003946ddb78df18efca166ca86e4055405`. Thread callable/encoding/oracle contract remains
  accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; its exact 19-path boundary is
  `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`. Folder contract `a565` and plan `a98` are
  accepted; Folder source+oracle are now narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692`
  (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review
  target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. Before any later plan or
  Phase-145 carrier, the active 28-doc correction requires transition review, one exact transition
  commit without self-SHA, fresh Lifecycle exact-SHA review of fcb plus transition with trace Contract/Security as needed, all exact reviews PASS, then separate 02F docs-only acceptance sync. No execution.

> **Current Rule gate:** the three Terra review routes PASS C/I/M `0/0/0` with dispatcher-enforced provenance, no fallback, and telemetry `UNKNOWN` unless exposed. 144-02D is docs-only accepted; `55e36b7a` 02F closure is historical and the live 28-carrier correction requires transition review, an exact transition commit without self-SHA, fresh Lifecycle exact-SHA review of fcb plus transition with trace Contract/Security as needed, all exact reviews PASS, then separate 02F docs-only acceptance sync. No Rule source/test/TDD/schema/runtime/product/release authority.
