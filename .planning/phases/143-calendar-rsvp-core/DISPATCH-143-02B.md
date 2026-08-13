[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только bounded parsing и
canonical duplicate identity для iTIP payload после принятого 143-02A serializer.

- User outcome: входящий `REQUEST` / `CANCEL` / `REPLY` разбирается детерминированно и
  отклоняет resource-exhaustion/spoof-shaped input до domain mutation.
- Completion: failing external-package units precede implementation; parser accepts the
  02A canonical output, rejects named unsafe cases, and returns immutable domain values.
- Non-goals: MIME traversal, sender authorization, persistence/mutation, Ent, HTTP, NATS,
  mail engine, frontend, migration, changes to legacy export.

[КОНТЕКСТ И OWNERSHIP]

- Entry dependency: independently accepted 143-02A on a clean reviewed HEAD.
- Frozen sources: `MAIL-SPEC.md` §§2, 4.4, 4.4.1, 7, 8; ADR iTIP/security sections;
  `DISPATCH-INDEX.md`.
- Allowed mutations only:
  - `backend/internal/calendar/itip_parse.go`
  - `backend/internal/calendar/itip_limits.go`
  - `backend/tests/calendar/itip_parse_test.go`
  - narrow additions to `itip_types.go` only if the accepted parser result cannot be
    represented; stop instead of broadening semantics.
- Do not modify `ical.go`, serializer logic, service/repository/handler, or add dependency.
- No filesystem/network/DB/broker/listener/global state/wall clock.

[FROZEN ENTRY POINTS]

```go
type ITIPParseLimits struct {
    MaxBytes int
    MaxUnfoldedLineBytes int
    MaxProperties int
    MaxAttendees int
    MaxExDates int
}

func ParseITIP(payload []byte, limits ITIPParseLimits) (ITIPMessage, error)
func CanonicalITIPIdentity(message ITIPMessage) (ITIPIdentity, error)
func DefaultITIPParseLimits() ITIPParseLimits
```

The public identity is exactly:

```go
type ITIPIdentity struct {
    CanonicalVersion uint8
    Method ITIPMethod
    UID string
    HasRecurrenceID bool
    RecurrenceID string
    Sequence int
    PayloadHash [32]byte
}
```

`DefaultITIPParseLimits` returns exactly `{1048576, 8192, 4096, 200, 1000}` in field order.
Every supplied field must be positive and no greater than that hard maximum. Component
depth is fixed at two (`VCALENDAR` -> `VEVENT`) with exactly one actionable event. Identity
does not authorize a sender or mutate state.

The hash is the separate inbound CVE-v1 stream frozen in Mail spec 4.4.1, never
`SerializeITIP` output. It starts with `hub.calendar.itip.payload`, NUL and version byte 1,
then streams the exact ordered length-delimited frames into SHA-256. Preserve all tag,
enum, integer, string, sort, RRULE, inclusion/exclusion and version rules literally.
Occurrence REQUEST is valid inbound; do not weaken the accepted outbound serializer to
make it round-trip. All-day `VALUE=DATE` normalizes to the fixed semantic timezone `UTC`
without midnight conversion or caller context.

[REQUIRED ORACLES]

1. Parse every 02A REQUEST/CANCEL/REPLY fixture to equivalent values; all-day output
   canonicalizes the wire's zone-less dates to timezone marker UTC. Do not require the
   parser to reserialize occurrence REQUEST through 02A.
2. Unfold CRLF+space/tab correctly; reject bare-CR, invalid UTF-8 and over-limit unfolded
   lines before unbounded allocation.
3. Enforce total bytes, property count, attendee count, EXDATE count, nesting/component
   depth and exactly one actionable VEVENT.
4. Reject missing/mismatched METHOD, UID, DTSTAMP, SEQUENCE, DTSTART/DTEND union,
   organizer/attendee requirements and unsupported method.
5. Validate IANA timezone; preserve instant across a DST case; preserve all-day exclusive
   date without midnight conversion.
6. Parse quoted parameter values, escaped text, folded multibyte text, attendee PARTSTAT/
   ROLE/RSVP, RRULE/EXDATE and optional RECURRENCE-ID.
7. Equal canonical payloads yield equal SHA-256 identity; equal UID/sequence with changed
   canonical payload yields different hash for quarantine logic. A test-owned literal
   CVE-v1 byte vector is hashed with standard SHA-256 and compared with the parser identity;
   do not copy the unverified host-calculated digest from research notes.
8. Property order, physical folding, parameter case/quoting and RRULE component order are
   same-hash equivalences. Every included field mutation differs. Attendee/EXDATE order is
   normalized, while duplicates reject.
9. Accept and hash timed and all-day occurrence REQUEST. All-day identity contains the UTC
   semantic marker/date union and no inferred instant. `CanonicalVersion` is exactly 1.
10. Lower/equal/newer sequence classification is pure and explicit; parser itself does not
   decide database mutation.
11. Malformed input returns stable safe sentinels with no raw addresses/body in error
    strings. Reject unknown properties/parameters, extra components, VTIMEZONE, RDATE,
    EXRULE, TRANSP, CLASS, URL, ATTACH, SENT-BY and recurrence RANGE.
12. Fuzz-style boundary tables are deterministic ordinary units, not host fuzz execution.

Narrow `itip_types.go` additions are exactly `ITIPIdentity` plus
`ErrInvalidITIPLimits`, `ErrITIPLimitExceeded`, `ErrInvalidITIPSyntax`, and
`ErrUnsupportedITIPContent`; do not invent another public parser surface. Parse returns
fresh normalized slices and no alias into the input. Follow the exact singleton/cardinality,
parameter-default, method/status and timed/all-day wire matrix in Mail spec 4.4.1.

[EXECUTION / PROVENANCE]

- TDD first; `package calendar_test`; in-memory payloads only.
- Host tests/build/lint/validators forbidden; direct Docker/Compose not fallback. Accepted
  launcher unavailable => `NOT_RUN` plus static evidence.
- No commit/push/PR/network/dependency install/generated code.
- max_depth 0; max_fan_out 0; retries 2/root cause; wall clock 45m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none.
- Orchestrator records exact clean HEAD and verifies only allowed paths changed.

[ОТВЕТ]

На русском: summary; files/signatures; accepted/rejected cases; verification; risks;
requested/actual model provenance; remaining unknowns.
