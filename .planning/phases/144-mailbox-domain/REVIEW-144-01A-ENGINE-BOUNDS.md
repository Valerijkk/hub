# Phase 144-01A deep review — engine payload bounds

> **Detected:** 2026-08-13 root deep static audit
> **Severity:** P1 contract boundary
> **Status:** `P1_STATICALLY_REMEDIATED_BY_0_1_8_ATOMIC_BOUNDARY` ·
> `EXACT_ATOMIC_COMMIT_INDEPENDENT_REVIEW_PENDING` · `DOWNSTREAM_BLOCKED` · `RUNTIME_NOT_RUN`
> **Frozen contract:** `docs/mail/ENGINE-CONTRACT-0.1.8.md` at
> `5e314fb080d420696642058ae47397e251c23207`

## Historical finding

The superseded 0.1.7 boundary embedded full `EngineMessageContent` in pages, accepted a
whole `GetBlob(accountID, blobID)` without a caller budget, and did not expose hard public
message/attachment ceilings. A provider response could therefore make an item-count-bounded
page consume unbounded retained memory. The historical 0.1.7 source/fix evidence through
`43d421ef` remains useful only as history; it is not the current safety gate.

## Decision record

The retained options were: (A) projection-only pages plus explicit bounded reads; (B) hydrated
pages with item and aggregate budgets; or (C) identifiers-only pages with N+1 hydration.
Option A is frozen in 0.1.8. It avoids projection body duplication without moving an adapter
or runtime-validation implementation into this DTO packet. The earlier founder-question
uncertainty is superseded by the frozen 0.1.8 contract and is no longer this packet's gate.

## Static remediation proof

The atomic seven-path boundary containing this evidence is `STATIC_CONFORMANT` against the
frozen contract. Static reviews
returned `Spec/Quality PASS` C/I/M `0/0/0`, then Security/Oracle PASS after fix round 1 and a
final C/I/M `0/0/0`. The source proof is:

1. `ListMessages`, `Changes` and `EnumerateProjection` return projection-only page values;
   `EngineMessageProjection` has no body, raw MIME or attachment-byte field.
2. `GetMessage` and `GetBlob` take explicit account-bound byte-budget requests; the closed
   DTO surface carries the 1 MiB item, 8 MiB page, 50 MiB message and 25 MiB attachment caps.
3. `EngineMessage` is exactly projection plus one authorized body. Draft and blob byte values
   remain only in their bounded read/write DTOs, not page projections.
4. `EngineSubmission` is the six-field reference-only value; it cannot carry or replace
   draft/message content.
5. `ErrEnginePayloadTooLarge` and `payload_too_large` are exact safe literals in both error
   allowlists; unknown error kinds retain the content-free fallback.
6. The external-package oracle fixes method signatures, field count/order/type, forbidden
   projection/submission fields, literal limits/enums and `errors.Is` sentinel behavior.

This is a structural source proof, not proof that a future adapter validates before allocation,
rejects an overflowing stream, or runs a fake/provider integration. Those behaviors require
their later packet and supported execution evidence.

## Remaining gate impact

The atomic seven-path boundary containing this evidence has parent
`fa0a2f2a542caadd9f5951e6f703b060b4776780`. Its own commit SHA cannot be self-recorded in the
same commit and must be captured by post-commit Git evidence; no placeholder SHA is valid.

- Independent review of that exact atomic commit is still pending.
- The accepted task-owned launcher remains `NOT_READY`; compile, unit, race, build, lint,
  validators, Docker and runtime remain `NOT_RUN`.
- `144-01B` remains blocked, and Phase 145 message/blob/change/enumeration mapping remains
  blocked until the exact-commit review and each packet's applicable supported Docker evidence
  close their gates.
- No direct Docker run can convert this static finding into runtime evidence.
