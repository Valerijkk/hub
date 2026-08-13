# Task 143-03F — inbound-REPLY receipt remediation and docs-sync record

> **Status:** SOURCE_FIX_STATICALLY_CONFORMANT · DOCS_REVIEW_AND_COMMIT_PENDING
> **Baseline dependency:** 143-03B `61c19f9f0dee5a406005aeacd4af056dcebc6a36`; its current attendee source includes the allowed 03F inverse-edge remediation.
> **Current boundary:** uncommitted source/test remediation at clean base `dd346e318b9ef2eab390f7509b06eb412a65d730`; no implementation SHA is claimed.

## Scope and provenance

- Authoring route: requested `gpt-5.6-terra` / `xhigh`, dispatcher-enforced; actual telemetry `UNKNOWN`.
- Independent review route: requested `gpt-5.6-terra` / `high`, dispatcher-enforced; actual telemetry `UNKNOWN`.
- Scoped production files only:
  - `backend/ent/schema/calendarinboundreplyreceipt.go`
  - narrow inverse edge in `backend/ent/schema/calendarattendee.go`
- Test file: `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go`.
- No generated Ent output, migration, API, or other production file changes are owned by this packet.

## Static conformance and traceability

Fresh review reopened the earlier packet at Critical/Important/Minor `0/2/0`: the attendee inbound-receipts ownership lacked `OnDelete(Cascade)`, and the 32-byte SHA-256 constraint was not an exact database CHECK. The current source remediation is statically conformant:

- required immutable attendee foreign key;
- non-negative immutable event sequence;
- required payload hash version with minimum one;
- required exact 32-byte SHA-256 payload hash, including named database CHECK `calendar_inbound_reply_receipts_payload_hash_length` = `octet_length(payload_hash) = 32`;
- closed immutable `applied|observed_no_change` outcome and immutable creation time;
- unique attendee plus sequence plus version plus hash;
- one additive attendee inverse edge, with `OnDelete(Cascade)` only on `inbound_reply_receipts` and all prior fields, edges and indexes preserved;
- no raw sender, MIME, HTTP key/response, engine id, event or organizer mutation edge;
- a genuine external Ent-metadata unit written first for the exact-length CHECK and owned-child-Cascade schema risks; its static expected RED is recorded, while runtime RED/GREEN remains `NOT_RUN`.

Current content hashes are:

| Path | SHA-256 |
|---|---|
| `backend/ent/schema/calendarattendee.go` | `34B5EBA8010C6C1BC29506C1BF073B3D77788451525680A164A138DE825F3B82` |
| `backend/ent/schema/calendarinboundreplyreceipt.go` | `22DF57A9E6AA6F5E87141B07DD8BE2F4C6337137464226D2CDC05FF0109C2BE2` |
| `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go` | `AEDEE0337C5FDBC65B307D18F2DE036DB30C363EF54E02539E0DA121186A1472` |

Sender-bound hash computation, the maximum-eight distinct-hash cap and retention remain service-owned behavior and are not claimed by this schema packet.

## Review and acceptance boundary

The fresh independent review returned `SPEC PASS` and `QUALITY PASS` with Critical/Important/Minor `0/0/1`; the sole Minor was documentation sync, now complete. Dispatcher-enforced provenance is recorded for both requested routes; actual telemetry remains `UNKNOWN`.

The current uncommitted commit boundary is pending. The source fix is statically conformant, and commit readiness is subject to independent docs review and the root exact commit. 143-03F is not `STATIC_ACCEPTED` until those gates complete; no future commit SHA is claimed. A later docs-only SHA sync may be required after that commit.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the accepted launcher is `NOT_READY`.
