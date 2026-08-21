# 143-04I — inbound iTIP REPLY immutable source receipt

## Status and lifecycle

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`.

This is a non-authoritative dirty docs receipt for immutable source commit
`46f4d6024e6c5475cdd943af9979c4ffadc5474a` (sole parent
`04dbde1a14e9499462dbc1ea8784583f83f0b40f`, tree
`e3ce6097822f360a04717d0dce61597810ca2046`, subject
`feat(calendar): add inbound reply service`). Source Contract/Security/Lifecycle
precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; provenance is
Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high,
`dispatcher-enforced`, no fallback, telemetry `UNKNOWN`.

Until one atomic manifest-equal exact10 receipt commit and the immutable receipt
Contract/Security/Lifecycle triple PASS, the source is not `STATIC_ACCEPTED`.
Only after that activation is source `STATIC_ACCEPTED / EXACT_SHA_PASS`; lower
04I authority/source-unimplemented routes then become `CONSUMED/HISTORICAL`.

## Immutable source exact3

The receipt order is authoritative:

| Order | Path | Blob |
|---:|---|---|
| 1 | NEW `backend/tests/calendar/event_service_inbound_reply_test.go` | `e0c4e64d9d704ebde24c87d9569590cde69d9115` |
| 2 | MODIFY `backend/internal/calendar/event_repository.go` | `c5fc1df5bf0a690db199b9df922c5463d27ab2a1` |
| 3 | NEW `backend/internal/calendar/event_service_inbound_reply.go` | `81814a334c17df9dbd24da58a6fb95a776924670` |

## Exact10 docs receipt manifest

The required atomic receipt commit diff is set-equal to this ordered manifest:

| Order | Change | Path |
|---:|---|---|
| 1 | MODIFY | `.planning/STATE.md` |
| 2 | MODIFY | `.planning/phases/143-calendar-rsvp-core/.continue-here.md` |
| 3 | MODIFY | `.planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md` |
| 4 | MODIFY | `.planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md` |
| 5 | MODIFY | `.planning/phases/143-calendar-rsvp-core/PLAN.md` |
| 6 | MODIFY | `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md` |
| 7 | MODIFY | `.planning/phases/143-calendar-rsvp-core/REVIEW-143-04I-INBOUND-REPLY-AUTHORITY.md` |
| 8 | MODIFY | `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md` |
| 9 | MODIFY | `docs/mail/CONTRACT-CHANGELOG.md` |
| 10 | NEW | `.planning/phases/143-calendar-rsvp-core/IMPLEMENTATION-143-04I.md` |

No source, schema, OpenAPI, QA, Mail specification, persistence contract or
testframework path belongs to this receipt commit.

## Residual blockers

Runtime remains `NOT_RUN`. The launcher/testframework external handoff,
`143-04F-I`, `143-03E`, `143-04F-Base`, Ent/generated source,
generation/migration, HTTP/API/UI and NATS/iTIP transport remain blocked or
not ready. No next Calendar packet is selected or authorized.
