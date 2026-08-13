# Phase 144-01A deep review — engine payload bounds

> **Detected:** 2026-08-13 root deep static audit
> **Severity:** P1 contract boundary
> **Status:** CONTRACT_AMENDMENT_REQUIRED · DOWNSTREAM_BLOCKED · RUNTIME_NOT_RUN

## What remains valid

The implementation through `43d421ef` matches the exact frozen 144-01A method surface and
its safe error taxonomy. No new source defect was found in the allowlisted `EngineError`
fix. The public interface, DTO shapes and external compile oracle remain useful evidence.

## Reopened contract gap

The current boundary does not actually bound response memory:

1. `EngineMessage` embeds `EngineMessageContent`, including recipient slices and complete
   plain/rich body strings. `ListMessages`, `Changes` and `EnumerateProjection` return up to
   100 such values, but neither the request nor DTO specifies a per-message or aggregate
   byte budget. An item-count cap alone is not a memory cap.
2. `GetBlob(accountID, blobID)` returns one complete `EngineBlob.Bytes` value and accepts no
   maximum byte request or bounded stream primitive. A provider-reported size can therefore
   drive allocation before the domain has a chance to reject it.
3. OpenAPI `max_message_bytes` and `max_attachment_bytes` have minimums but no maximums.
   Phase 145 says provider bodies/messages/blobs must be bounded before allocation, yet the
   accepted port and admin contract do not define the authoritative ceilings or over-limit
   behavior.

A malicious/compromised adapter response or an unsafe administrator value could therefore
turn a nominally bounded page into unbounded process memory. The deterministic fake cannot
serve as a meaningful conformance oracle until this is frozen.

## Decision-ready options

### Option A — recommended safe default

- projection/list/change/enumeration pages return a projection-only message DTO with no
  full body;
- authorized `GetMessage` uses an explicit request byte budget and returns one bounded
  content value;
- blob reads use either a caller-supplied byte ceiling plus exact-size rejection before
  allocation, or a bounded streaming handle whose reader cannot exceed the ceiling;
- OpenAPI/admin policy gets hard `maximum` values no larger than deployment hard caps;
- one stable content-too-large error is safe and contains no provider/body/id value.

### Option B — retain one hydrated DTO

Keep `EngineMessage` in every page, but add explicit per-item and aggregate byte budgets to
every read/page request and require the adapter to reject before body/blob allocation. This
is a smaller type split but makes projection sync pay for body hydration and increases the
memory/error matrix.

## Gate impact

Until founder questions 28, 30 and 31 are answered or the safe default is explicitly
accepted:

- 144-01A is source-conformant but no longer a fully accepted contract gate;
- 144-01B should not be authored against a surface expected to change;
- Phase 145 message/blob/change/enumeration mapping remains blocked independently of its
  license/runtime gates;
- no direct Docker run can resolve an under-specified public boundary.

No production/test source, dependency, generator, migration, runtime, test or Docker command
was changed or run by this review.
