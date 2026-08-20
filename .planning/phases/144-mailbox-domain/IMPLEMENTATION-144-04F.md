# 144-04F — immutable sync-page policy source receipt

## 143-03B-T immutable source acceptance and conditional 143-04C gate (2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; 144-04G still requires separate docs authority because its repository port is absent. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## Status

`STATIC_ACCEPTED / EXACT_SHA_PASS` for immutable source commit
`4715f939f24e5b5c2d3f80a38d204babe68bdd34` only. This uncommitted documentation
working tree is not source authority; it records the independently reviewed immutable receipt.

| Field | Immutable value |
|---|---|
| Commit | `4715f939f24e5b5c2d3f80a38d204babe68bdd34` |
| Parent | `7ac8714e99db2935ce290ad67836e1309cc05c78` |
| Tree | `6461894a90c4340bf96118ed1a242516c02b48c7` |
| Subject | `feat(mail): add sync page policy` |
| Source author | Terra/xhigh |
| Provenance | dispatcher-enforced, no fallback; telemetry `UNKNOWN` unless exposed |
| Reviews | Contract/high, Security/xhigh, ABI/Lifecycle high; precommit and immutable exact-SHA PASS |
| Review findings | C/I/M `0/0/0` |
| Runtime | launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN` |

## Exact test-first manifest and blobs

No substitution, extra path or reordering belongs to this source receipt:

1. `backend/tests/mail/mailbox_events_test.go` — `50a41cefb539d9b840a43da9e1ff34cce6c8653f`
2. `backend/tests/mail/mailbox_sync_policy_test.go` — `b6410709e4fb8166c28a5c502cfba535eef08d06`
3. `backend/internal/mail/mailbox_events.go` — `a3017b0a6064d8fc1017e37e2d4f675ee8732945`
4. `backend/internal/mail/mailbox_sync_policy.go` — `acd642fcdb6290af79b7fa8e5aa36a5036576f47`

## Accepted static behavior

- The pure classifier consumes exact-one change/enumeration page plus caller-supplied folders,
  projections, destroyed state, labels, Hub IDs, prior state, provenance and event reservations. It
  performs no DB, Ent, network, clock, global-state, generation or migration work and returns only
  deep-copied projection/event intents.
- First-seen inbound/unknown requires nonnil valid `ReceivedAt`. Strictly after `EventCutoff` it
  requires nonzero `ReceivedEventID`, persists that marker and emits exactly one
  `mail.received.v1`. Before/equal cutoff accepts only zero reservation, persists a nil marker and
  emits no received event; a historical/equal reservation is `invalid`. Stable confirmed owner
  submission/draft never emits received.
- Projection `FolderIDs` nil or empty is `invalid`; only cardinality greater than
  `MaxEngineFolderItems` is `too_large`. The overflow preflight runs before page byte traversal and
  allocation. Change and enumeration oracles cover nil, empty and over-max with a zero output
  result on every rejection.
- Event-builder changes are limited to max 100 for `mail.received.v1.folder_ids`,
  `mail.message.changed.v1.folder_ids` and `mail.message.changed.v1.label_ids`; 100 accepts and 101
  rejects, with event subjects, field names and versions unchanged.

## Frozen limits

- Page: `Requested` 1..100, `Maximum == 100`, exact `Returned`; change live+destroyed <=100,
  enumeration messages <=100. Page projection bytes <=8,388,608; each projection <=1,048,576.
- Projection collections: each To/CC/BCC/Reply-To <=200, references <=100, folder IDs 1..100,
  reserved keywords <=200 and attachments <=50. Caller-supplied folders, label catalog,
  previous/desired sets and event reservations are each <=100.
- Scalars: opaque ID <=1,024 bytes, cursor <=4,096, message <=52,428,800, attachment
  <=26,214,400, header <=998, snippet <=4,096 and MIME field <=1,020 bytes.
- All collection overflow is `too_large`; structurally missing/nil/empty required state is
  `invalid`; duplicates and account mismatch retain their dedicated safe error classes. Bounds are
  checked before proportional traversal/allocation.

## Lifecycle and exclusions

Every clause in the combined exact17 packet that says Mail exact4 is dirty, `UNCOMMITTED`,
`UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is
`CONSUMED/HISTORICAL`. The current Mail source truth is the immutable receipt above. No automatic
Mail next source follows: 144-04G remains `NOT_AUTHORIZED / BLOCKED` pending separate docs authority
because its required repository port is absent.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains
source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted.
143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is
`STATIC_ACCEPTED`. Existing 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation,
migration and launcher readiness remain blocked or `NOT_READY`; all executable evidence remains
`NOT_RUN`. No raw iTIP, service/API/UI, product, phase, release, generation or migration authority
is granted by this receipt.
