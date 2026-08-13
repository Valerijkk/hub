# Phase 142 machine-contract remediation — 2026-08-13

> **Status:** ROOT_REMEDIATED · INDEPENDENT_ACTUAL_PROVENANCE_REREVIEW_PENDING
> **Reviewed source:** `c48046b876f429c7ab31f58fbc3ee0b960529144`
> **Requested reviewer route:** `gpt-5.6-terra` / `high`
> **Actual reviewer telemetry:** `UNKNOWN`
> **Executable validation/runtime:** `NOT_RUN`; accepted launcher `NOT_READY`

## Review disposition

| Severity | Finding | Root disposition |
|---|---|---|
| P1 | Public response objects, including `CalendarEvent` and `ErrorEnvelope`, admitted undeclared fields | Closed every operation-visible object. Thread/message composition now uses explicitly internal open field bases under closed summary/detail wrappers. |
| P1 | Deprecated legacy broadcast admitted unbounded HTML, unknown request fields and unbounded counters | Kept the already-blocked compatibility route without re-authorizing it; closed request/response, capped body at 8,388,608 UTF-8 bytes and counters at non-negative int32, and froze aggregate/dry-run invariants. Phase 148 disable/removal gate remains. |
| P1 | Calendar invite event payload admitted 500 attendees while the canonical event cap is 200 | Both invite payloads now cap unique `attendee_ids` at 200. The same audit exposed stale mail received/sent `to` caps; those now match the engine projection cap 200 too. |
| P1 | Channel-scoped due subject could name a different channel than `payload.channel_id` | Added structured `CALLS_CHANNEL_SUBJECT_PAYLOAD_MATCH`: regex capture group 1 from `/subject` must equal normalized `/payload/channel_id`; producer rejects before outbox and consumer dead-letters with zero mutation. |
| P1 | Non-admin ownership/isolation was prose-only | Added default-deny `x-hub-authz-vocabulary` v1 and exactly one declared policy id to all 56 non-admin operations; all 13 admin operations retain explicit admin RBAC. Opaque inaccessible and nonexistent ids share the non-disclosing NFR-13 `404`; the read-only account-status probe uses a distinct resource-optional self-namespace policy so `not_configured` remains representable. |
| P2 | Cursor, recovery and replay examples were absent | Added standard OpenAPI examples for complete-item cursor continuation, calendar create plus identical replay, recoverable calendar `409` headers/body and mailbox `423` retry metadata/body. |
| P2 | Policy allowed 500 recipients although three draft lists can contain at most 300 | Added `MaxEngineDraftTotalRecipientItems = 300`; OpenAPI and persistence policy maxima are 300 and the normalized combined-count rule rejects recipient 301 before engine I/O. |

## Root static evidence

- Operation inventory: 69 total; 56/56 non-admin operations have one of 14 declared
  `x-hub-authz` policy ids; 13/13 admin operations retain `x-hub-rbac`; zero absent/unknown
  policy ids.
- Object inventory: the only `type: object` blocks without a closure keyword are
  `MailThreadFields` and `MailMessageFields`. Both carry
  `x-hub-internal-composition-only: true` and are referenced only by closed wrappers/details,
  never by an operation.
- Bound inventory: event `to` and invite `attendee_ids` maxima are 200; public/persisted
  `max_recipients` maxima are 300; no Mail 0.1.8 `max_recipients` maximum 500 remains.
- Example inventory contains `bounded_page_with_continuation`,
  `identical_input_replay`, `stale_event_version` and `account_busy`.
- `git diff --check` is rerun on the complete 12-path working-tree diff and
  `git diff --cached --check` is rerun immediately before the atomic commit.
- Changed paths remain outside every protected test-rebase/QA-AI zone.

## Honesty boundary

This record is not an independent PASS. The review runtime did not expose actual model or
effort, no post-remediation independent rereview has occurred, and the project forbids host
OpenAPI/JSON Schema validation. Docker daemon availability does not replace the accepted
task-owned launcher, which remains `NOT_READY`; executable validation and runtime remain
`NOT_RUN`.
