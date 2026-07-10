# SEED-002 — Unified calendar + RSVP + mail/calls sync

> Planted: 2026-07-10 · Brief: `.planning/JULY-2026-REBUILD-BRIEF.md` §3

## Idea

Единый канон календарных событий с RSVP: созвон, созданный в мессенджере или календаре, после Accept у коллеги сразу виден в его календаре (и у организатора). Плюс уровни email-синхронизации (ICS → iCal feeds → опционально IMAP/OAuth).

## Why

Сейчас calendar — **read-агрегатор** (scheduled calls + sprints + deadlines + leave). ScheduledCall имеет participants, но **нет accept/decline**. Нет first-class event + attendee RSVP. Пользовательский JTBD: «поставил созвон → приняли → у обоих в календаре».

## Trigger

Поверхностно при:

- перестройке **конец июля 2026**;
- Calendar v2 волнах (ROADMAP W3–W4);
- любом обсуждении GZ-V4-05 (iCal) — эта идея **расширяет** его.

## Scope

**Large** — эпик (ERD + API + messenger/calendar UI + notifications + ICS).  
OAuth Google/Outlook (L3) — отдельный ADR, сейчас **противоречит** GZ-V4-05.

## Breadcrumbs

- Calendar aggregate: `backend/internal/calendar/service.go` (ListScheduledCalls + sprints + issues + leave)
- Scheduled calls: `backend/ent/schema/scheduledcall.go`, `backend/internal/calls/scheduled.go`
- Conflict API: `FindConflictingCalls`
- FE: `frontend/src/modules/calendar/unified-calendar.tsx`, `modules/calls/calendar-*.tsx`, `create-call-dialog.tsx`
- Mail: `backend/cmd/hub-mail`, admin mail settings
- Decisions: `GRAY-ZONES-DECISIONS` GZ-V4-05 (iCal only, no OAuth); US-CAL-* в `docs/vision_and_scope.md`

## Proposed levels (mail)

| Level | Scope |
|-------|-------|
| L0 | ICS в email invite/update/cancel via hub-mail |
| L1 | iCal export + external subscribe (GZ-V4-05) |
| L2 | IMAP/JMAP ingest |
| L3 | OAuth two-way Google/Microsoft — needs ADR override |

## Critical edge cases (full list in brief)

- Recurring accept series vs instance  
- Reschedule: reset RSVP vs keep accepted  
- Conflict busy/free + leave  
- Channel-default attendees  
- External/guest users out of scope v1  
- Optimistic lock on attendee RSVP  

## Next when surfaced

1. product/ba2 — US/AC RSVP + messenger/calendar create  
2. sa2 — ERD `CalendarEvent` / `CalendarAttendee`, NATS, ICS  
3. arch — `ADR-CAL-UNIFIED` (+ optional mail ADR)  
4. designer — invite inbox / accept UX (Cloud Code mockups)  
5. be/fe + qa TC на corner cases brief §3.4  
