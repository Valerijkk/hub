# Task 144-01A — static conformance evidence

> **Status:** `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` · `EXACT_SHA_PASS` · `RUNTIME_NOT_RUN`
> **Frozen contract:** Mail 0.1.8 144-01A entry amendment at `5e314fb080d420696642058ae47397e251c23207`; Thread contract remains accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`, static only
> **Atomic boundary:** `e3e8c774d734f926f77f8e600b58dc68dabbd993`; parent
> `fa0a2f2a542caadd9f5951e6f703b060b4776780`.

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
  C/I/M `0/0/0`), static only. Their exact 20-path docs candidate awaits three exact-SHA reviews;
  after PASS root only freezes/adjudicates `144-03B-Rule`. No execution.
> **144-03B Folder post-source lifecycle (2026-08-14):** source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692` (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only, with exactly `backend/internal/mail/folder_tree.go` blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b` and `backend/tests/mail/folder_tree_test.go` blob `efe0db0071cb22ce4b026c635157f0fb08885d64`. Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high exact-SHA reviews are `EXACT_SHA_PASS`, C/I/M `0/0/0`. This exact 20-path docs candidate (the frozen prior 19-path lifecycle manifest plus its new evidence file, including the three stale-gate evidence documents) awaits its own Lifecycle/Evidence high, Security/Dependency xhigh and Exact/Boundary high exact-SHA reviews. Launcher `NOT_READY`; compile, unit, race, build, lint, validators, Docker and runtime `NOT_RUN`; 04B reparent, 144-01B, Phase 142 and Phase 148 remain blocked. Author provenance Terra/medium, `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. After candidate reviews PASS, root only freezes/adjudicates `144-03B-Rule`; no Rule TDD or production authorization.
