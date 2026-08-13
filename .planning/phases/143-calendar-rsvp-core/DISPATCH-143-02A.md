[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только детерминированную
сериализацию iTIP `REQUEST` / `CANCEL` / `REPLY` для Phase 143.

- User outcome: Hub строит canonical RFC 5545/5546 payload без wall-clock drift и без
  зависимости от mail engine.
- Completion: сначала external-package units, затем минимальная реализация; exact CRLF,
  escaping/folding, METHOD/UID/SEQUENCE/DTSTAMP, timed/all-day, timezone, organizer,
  attendees, recurrence id и status проходят oracle. Если exact runtime/контракт не
  совпадает — `BLOCKED` до записи.
- Non-goals: parser, MIME, Ent, repository/service/handler, NATS, frontend, migration,
  изменение legacy `RenderICalendar`.

[КОНТЕКСТ И OWNERSHIP]

- Frozen sources: `MAIL-SPEC.md` §§2, 3.11–3.13, 4; `ADR-MAIL-01` Calendar policy;
  `docs/mail/openapi.yaml` calendar schemas; Phase 143 `DISPATCH-INDEX.md`.
- Allowed mutations only:
  - `backend/internal/calendar/itip_types.go`
  - `backend/internal/calendar/itip_serialize.go`
  - `backend/tests/calendar/itip_serialize_test.go`
- Do not modify `ical.go`; its `time.Now()` legacy export is out of this packet.
- You are not alone. Do not revert others; stop on dirty overlap.
- Do not read/touch protected QA/test-rebase paths listed in root `AGENTS.md`.
- Go 1.25; Russian comments for exported identifiers; no new dependency; no global state;
  no filesystem/network/DB/broker/listener/wall clock.

[FROZEN PUBLIC SHAPE]

Define the smallest value model needed by the serializer, including typed method/status/
participant enums and explicit temporal union. Public entry point:

```go
func SerializeITIP(message ITIPMessage) ([]byte, error)
```

`ITIPMessage` must carry its own non-zero UTC `Stamp`; do not inject or call a clock inside
serialization. Timed events carry start/end instants plus a validated IANA timezone name;
all-day events carry date strings with exclusive end and no instants. Occurrence REPLY/
CANCEL carries canonical `RecurrenceID`; series messages do not.

[REQUIRED ORACLES]

Tests use `package calendar_test` and in-memory values only:

1. `REQUEST` emits VCALENDAR + VEVENT, matching `METHOD:REQUEST` at both contract layers.
2. `CANCEL` carries incremented SEQUENCE and `STATUS:CANCELLED`.
3. `REPLY` contains exactly one responding attendee and preserves organizer/event fields.
4. Timed values preserve UTC instants and TZID semantics across a DST fixture.
5. All-day uses `VALUE=DATE` and exclusive DTEND; no inferred midnight timestamp.
6. UID, DTSTAMP, SEQUENCE, ORGANIZER, ATTENDEE/PARTSTAT/ROLE/RSVP, RRULE, EXDATE and
   optional RECURRENCE-ID serialize deterministically.
7. Text escaping covers backslash, semicolon, comma and newline; UTF-8 content is retained.
8. Content lines fold at the RFC octet boundary without splitting a UTF-8 code point and
   continue with one space; all line endings are CRLF and output ends CRLF.
9. Invalid temporal union, method/field combination, email, timezone, sequence or missing
   required field returns a stable typed/sentinel validation error.
10. Two calls with the same value are byte-identical.

[EXECUTION / PROVENANCE]

- TDD: write failing units first, then minimal implementation.
- Host tests/build/lint/validators are forbidden. Direct Docker/Compose is not fallback.
  If accepted launcher is unavailable, return `NOT_RUN` and static evidence only.
- No commit/push/PR/network/dependency install/generated code.
- max_depth 0; max_fan_out 0; max retries per root cause 2; wall clock 45m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none.
- Orchestrator supplies and records exact clean HEAD at dispatch.

[ОТВЕТ]

На русском: summary; changed files + exact public types/signature; covered cases; verification
with `NOT_RUN` if applicable; risks; requested/actual model evidence and remaining unknowns.
