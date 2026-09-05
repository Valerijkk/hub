# Mail occurrence RSVP design

## Purpose

Mail must expose the correct RSVP target for a calendar invitation without deriving recurrence semantics from subject, body, or client-side heuristics. A single-event invitation continues to RSVP against the series; a recurrence-instance invitation uses the instance endpoint and renders the effective attendee state returned by Calendar.

## Durable contract

`calendar_inbound_message_receipts` gains an additive, nullable snapshot pair:

- `calendar_component_scope`: enum `series | occurrence`;
- `recurrence_id`: canonical recurrence identifier for `occurrence` only.

The pair is tri-state and immutable once the terminal receipt outcome is completed:

| Normalized iTIP identity | Stored snapshot |
| --- | --- |
| unknown / unavailable | both `NULL` |
| series | `scope=series`, `recurrence_id=NULL` |
| occurrence | `scope=occurrence`, canonical `recurrence_id` |

Completion preserves an already-written equivalent snapshot and rejects an incompatible overwrite. The migration is additive and intentionally does not backfill historical rows. A successfully parsed `deferred_missing_master` receipt is necessarily a known occurrence and persists its canonical occurrence pair for durable retry, but it remains non-actionable; outbound, legacy, nonterminal, and unknown receipts remain without a target.

## Backend projection

Mail keeps `calendar_event_id` and adds the typed nullable projection:

```json
"calendar_rsvp_target": {
  "event_id": "uuid",
  "scope": "series | occurrence",
  "recurrence_id": "canonical-id | null",
  "attendee_version": "positive integer",
  "rsvp_status": "needs_action | accepted | tentative | declined"
}
```

The projection is emitted only when the correlated inbound receipt is terminal `applied`, `duplicate`, or `ignored_stale`, has a known valid snapshot, resolves an event in the same tenant, and passes the existing mailbox actor/visibility fences. Otherwise it is `null` (fail closed). For `series`, `recurrence_id` is null; for `occurrence`, it is non-empty and canonical. The existing event navigation remains master-event navigation even for an occurrence target.

Both `attendee_version` and `rsvp_status` are required projection values. For an occurrence, they are the effective occurrence-attendee override when active; otherwise they fall back to the active series attendee. No inactive, removed, ambiguous, or invalid attendee state creates a target. This gives the RSVP request its correct compare-and-set version and gives the invitation card a truthful initial status without treating a series attendee as an occurrence override.

The receipt repository writes the snapshot as part of normalized iTIP completion rather than re-parsing a message at Mail read time. Ent schema, the handwritten SQL receipt repository, migration, generated schema artifacts, correlation result, API DTO/OpenAPI contract, and Mail service projection evolve together.

## Frontend behavior

`MailCalendarInvitation` consumes only `calendar_rsvp_target`:

- `series` calls the existing series RSVP mutation;
- `occurrence` calls the existing occurrence RSVP mutation with its canonical recurrence ID;
- null hides/disabled RSVP controls according to the current invitation UI and never guesses.

After an occurrence RSVP succeeds, the UI uses Calendar's returned effective attendee state (including CAS/version result) to show the truthful response. It does not optimistically pretend a series attendee changed. "Open event" continues to open the master event.

Mail cache keys gain canonical factories for thread root, drafts, direct message, and identities so message-detail invalidation and route queries share one shape.

The client validates the complete target at the unknown-value boundary, including a canonical UUID event ID equal to the navigation event, positive safe-integer attendee version, closed RSVP status, and the exact scope/recurrence nullability pair. It fails closed for `undefined`, primitives, arrays, invalid scopes, or malformed recurrence values. Occurrence routing passes the canonical recurrence ID to the existing Calendar API, which raw-base64url-encodes the URL segment; it is not derived from display text. After a response, the component keeps Calendar's returned effective attendee/version locally and invalidates the canonical Mail messages root plus Calendar detail; it does not overwrite a shared master-event attendee cache with occurrence state.

## Invariants

1. No frontend parsing of ICS, subject, body, or timestamps decides RSVP scope.
2. No receipt snapshot is inferred from an event after the fact; the terminal receipt records the normalized delivery identity.
3. Tenant, account, actor, and mailbox visibility fences remain mandatory before projection.
4. Incomplete/deferred/legacy data is safely non-actionable rather than silently treated as a series RSVP.
5. The change is additive, backward-compatible, and does not alter outbound invitation semantics.

## Rejected alternatives

- **Always RSVP the series:** incorrectly responds for a single recurring occurrence.
- **Derive recurrence from Mail presentation data:** brittle, spoofable, and duplicates Calendar's normalization rules.
- **Resolve scope only at detail read time:** makes historical/replayed state ambiguous and weakens receipt durability.
- **Backfill old receipts:** cannot establish authoritative normalized identity and risks producing false actionable targets.
