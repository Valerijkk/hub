# Dev Mail/Calendar Demo Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the local development stack demonstrate a deterministic Mail mailbox, one canonical Calendar invitation, Calendar occurrences in the existing aggregate, and scheduling from a Mail message without using Stalwart or QA-only facilities.

**Architecture:** A `!qa && !release` selector composes one in-memory `mail.FakeMailEngine`, ordinary Mail workers, and the existing Calendar delivery chain. `internal/devdemo` remains provider-neutral behind `MailDemoPort`; a server adapter owns FakeMailEngine, Ent control-plane reconciliation, bounded sync, and lifecycle observation. Calendar occurrences extend the aggregate under event-detail authorization, while Mail opens existing `CreateCallDialog` with resolved initial participant IDs.

**Tech Stack:** Go 1.25, Fiber v3, Ent/PostgreSQL, NATS JetStream, React, TypeScript, TanStack Query/Router, Vitest/RTL, Docker Compose.

**Spec:** `docs/superpowers/specs/2026-08-22-mail-calendar-production-integration-design.md` plus the frozen root contract in this task.

## Global Constraints

- `MAILBOX_DEMO_ENABLED` is an explicit boolean selector compiled only under `!qa && !release`; the `qa || release` stub is inert, and startup is fatal unless `DEV_MODE=true`.
- QA has priority: do not alter `backend/cmd/server/wiring_calendar_mail_qa.go`, `backend/internal/qamailcalendar/**`, QA routes, recovery tokens, or any QA build behavior.
- Preserve normal Stalwart semantics and configuration; common composition may be extracted and reused where required. The demo performs no external network, DNS, SMTP, S3, ClamAV, upload staging, or scanner work.
- `internal/devdemo` must not own concrete MailEngine, FakeMailEngine, Ent Mail repository, or server worker. The one runtime FakeMailEngine is the source of demo Mail content. Compose, reply, and forward are attachment-less; only seeded-message attachment downloads read fake blobs.
- Seed deterministically and idempotently for only `pm@hub.local` and `dev1@hub.local`; direct Ent writes are limited to prerequisite organization/account/identity control-plane rows. Do not directly write Mail folders, threads, messages, links, attachments, inbound receipts, or `calendar_event_id`.
- Provide inbox, sent, drafts, archive, trash, and junk through a narrow Mail-package FakeMailEngine capability, then project them with bounded `MailboxSyncService` work.
- Create exactly one Calendar invitation through `CalendarInviteDelivery` -> `CalendarMailDeliveryService` -> `MailboxSubmissionService` -> dispatcher/finalizer; never fabricate correlation rows or identifiers.
- Extend `GET /api/calendar/events` additively with canonical CalendarEvent occurrences, exact detail-equivalent authorization, recurrence and `exdates`, cancelled omission, linked `ScheduledCall` de-duplication, canonical direct URL, and bounded-query loading without N+1.
- Do not read or change `pro-qa-ai/**`. Preserve the dirty `docker-compose.local.yml`; its only planned edit is one backend environment entry.
- Write tests in TDD order, but do not run Docker after individual micro-steps. Final execution is strictly sequential: `backend-unit`, then `frontend-unit`, then one local Compose build/up/manual-readiness session. Run no host validators, tests, builds, linters, Vite, Playwright, pytest, or app processes.
- Production migration runner and externally exposed Stalwart SMTP are explicitly outside this packet and remain separate follow-up work after local manual-readiness evidence.

---

## File Structure

| File | Responsibility |
|---|---|
| `backend/internal/devdemo/mail_runtime_config.go` | Pure publicly testable strict environment parsing; no server dependency. |
| `backend/internal/devdemo/mail_port.go` | Narrow provider-neutral `MailDemoPort`, inputs, and result shape. |
| `backend/cmd/server/wiring_calendar_mail_demo.go` | `!qa && !release` selector, dev-only runtime composition, worker lifecycle, and source seeding trigger. |
| `backend/cmd/server/wiring_calendar_mail_demo_stub.go` | `qa || release` no-op selector so QA/release cannot enable demo runtime. |
| `backend/cmd/server/devdemo_mail_adapter.go` | Concrete MailDemoPort adapter: fake source load, Ent control-plane reconciliation, bounded sync, and invitation observation. |
| `backend/internal/mail/mailbox_fake_seed.go` | Narrow FakeMailEngine seed operation for the six standard fake folders. |
| `backend/internal/mail/mailbox_draft_attachment_materializer_demo.go` | `!qa && !release` attachment-less materializer rejecting non-empty attachment IDs. |
| `backend/cmd/server/wiring_calendar_mail.go` | Calls the demo selector after common authenticated Mail routes and before the Stalwart configuration path. |
| `backend/internal/devdemo/mailbox_dataset.go` | Fixed PM/dev1 fake-engine folders, messages, and one seeded blob; validates deterministic source data. |
| `backend/internal/devdemo/mailbox_seeder.go` | Ensures only permitted control-plane Ent rows, loads the fake source, and drives bounded projection synchronization. |
| `backend/internal/devdemo/seeder.go` | Extends the existing `/api/dev/seed` result by delegating Mail reconciliation to the new source/worker boundary. |
| `backend/internal/calendar/event_aggregate.go` | Actor-authorized, batched CalendarEvent occurrence read port and conversion to the legacy aggregate response. |
| `backend/internal/calendar/service.go` | Merges canonical occurrences with legacy calls/sprints/issues/leaves and suppresses the linked-call duplicate. |
| `backend/internal/calendar/dto.go` | Adds the additive `event` aggregate kind and direct Calendar-event URL representation. |
| `frontend/src/modules/mail/mail-schedule-action.tsx` | Resolves recipients and opens the existing call dialog. |
| `frontend/src/routes/_authenticated/mail/$folderKey/thread/$threadId.tsx` | Mounts the schedule action beside existing message actions. |
| `frontend/src/modules/calls/create-call-dialog.tsx` | Backward-compatible optional `initialParticipantIds` and `initialTitle`; retains availability/create/navigation. |
| `frontend/src/locales/en/mail.json` and `frontend/src/locales/ru/mail.json` | Localized labels, pending, unavailable, and validation text for the added dialog. |
| `backend/tests/mail/*_test.go`, `backend/tests/calendar/*_test.go`, `frontend/tests/mail/*` | Genuine isolated Go and Vitest/RTL tests for selector, source/projection, aggregate reads, and schedule action. |
| `docs/wiki/QUICKSTART.md`, `docs/TECHNICAL.md`, `docs/TESTING.md` | Document dev-only activation, observable demo dataset, one Docker verification service, and residual production limits. |

### Task 1: Freeze the dev-only runtime selector

**Files:**
- Create: `backend/cmd/server/wiring_calendar_mail_demo.go`
- Create: `backend/cmd/server/wiring_calendar_mail_demo_stub.go`
- Create: `backend/internal/devdemo/mail_runtime_config.go`
- Create: `backend/internal/devdemo/mail_port.go`
- Modify: `backend/cmd/server/wiring_calendar_mail.go:44-160`
- Modify: `docker-compose.local.yml:backend service environment block containing DEV_MODE`
- Test: `backend/tests/dev/mail_runtime_config_test.go`
- Test: `backend/tests/dev/calendar_seeder_test.go`

**Interfaces:**
- Consumes: `serverDeps`, public `devdemo.ParseMailRuntimeConfig`, existing Calendar demo seeder.
- Produces: server-only `wireDevMailCalendarDemoRuntime`; `internal/devdemo` exports strict parser and `MailDemoPort` but has no concrete Mail/Ent/server dependency.

- [ ] **Step 1: Write the failing selector unit test**

```go
func TestParseMailRuntimeConfig(t *testing.T) {

	t.Run("requires explicit true selector and DEV_MODE", func(t *testing.T) {

		config, err := devdemo.ParseMailRuntimeConfig(func(key string) (string, bool) {
			return map[string]string{"MAILBOX_DEMO_ENABLED": "true", "DEV_MODE": "true"}[key], true
		})
		require.NoError(t, err)
		require.True(t, config.Enabled)
	})

	for _, value := range []string{"", "1", "TRUE", "false"} {
		t.Run(value, func(t *testing.T) {
			_, err := devdemo.ParseMailRuntimeConfig(func(key string) (string, bool) {
				if key == "MAILBOX_DEMO_ENABLED" { return value, true }
				return "true", true
			})
			require.Error(t, err)
		})
	}
}
```

- [ ] **Step 2: Implement the compile-time split and strict parser**

```go
type MailRuntimeConfig struct{ Enabled bool }

func ParseMailRuntimeConfig(lookup func(string) (string, bool)) (MailRuntimeConfig, error) {
	selected, present := lookup("MAILBOX_DEMO_ENABLED")
	if !present || selected == "" { return MailRuntimeConfig{}, nil }
	if selected != "true" { return MailRuntimeConfig{}, errors.New("MAILBOX_DEMO_ENABLED must be true when set") }
	if devMode, ok := lookup("DEV_MODE"); !ok || devMode != "true" { return MailRuntimeConfig{}, errors.New("MAILBOX_DEMO_ENABLED requires DEV_MODE=true") }
	return MailRuntimeConfig{Enabled: true}, nil
}
```

Put `ParseMailRuntimeConfig` in untagged `backend/internal/devdemo/mail_runtime_config.go`. The `!qa && !release` server selector calls it; the `qa || release` stub returns false without reading env. Invoke it after the QA early return and before Stalwart config parsing. Add exactly `MAILBOX_DEMO_ENABLED: "true"` beside backend `DEV_MODE: "true"`; do not alter another Compose byte.

- [ ] **Step 3: Add neutral port and keyed Calendar contract**

```go
type MailDemoPort interface {
	ReconcileMailDemo(context.Context, MailDemoInput) (MailDemoResult, error)
}

const MailInvitationEventKey = "pm-dev1-invitation"
```

Add immutable non-empty keys to Calendar demo plans, including exactly one PM-to-dev1 plan keyed `MailInvitationEventKey`; the Calendar seed result returns the created/replayed ID map. `MailDemoPort` receives only this ID and PM/dev1 identity scope.

- [ ] **Step 4: Commit the selector/port packet**

```bash
git add backend/internal/devdemo/mail_runtime_config.go backend/internal/devdemo/mail_port.go backend/internal/devdemo/calendar_dataset.go backend/internal/devdemo/calendar_seeder.go backend/tests/dev/mail_runtime_config_test.go backend/tests/dev/calendar_seeder_test.go backend/cmd/server/wiring_calendar_mail.go backend/cmd/server/wiring_calendar_mail_demo.go backend/cmd/server/wiring_calendar_mail_demo_stub.go docker-compose.local.yml
git commit -m "feat(devdemo): add strict mail runtime contract"
```

### Task 2: Seed the fake source and materialize its Mail projection

**Files:**
- Create: `backend/internal/devdemo/mailbox_dataset.go`
- Create: `backend/cmd/server/devdemo_mail_adapter.go`
- Create: `backend/internal/mail/mailbox_fake_seed.go`
- Create: `backend/internal/mail/mailbox_draft_attachment_materializer_demo.go`
- Modify: `backend/internal/devdemo/seeder.go:17-78`
- Modify: `backend/cmd/server/wiring_calendar_mail_demo.go`
- Test: `backend/tests/dev/mailbox_dataset_test.go`
- Test: `backend/tests/mail/mailbox_fake_seed_test.go`
- Test: `backend/tests/mail/mailbox_draft_attachment_materializer_demo_test.go`

**Interfaces:**
- Consumes: `devdemo.MailDemoPort`, a server-owned FakeMailEngine, existing Mail sync/submission workers, and existing devdemo organization/calendar boundaries.
- Produces: a server-only concrete adapter; devdemo holds no FakeMailEngine, Ent repository, or worker field.

- [ ] **Step 1: Write failing fake-dataset and bounded-sync tests**

```go
func TestMailboxDemoDatasetUsesOnlyPMAndDev1(t *testing.T) {

	dataset, err := devdemo.NewMailboxDemoDataset(demoIDs("pm@hub.local", "dev1@hub.local"))
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"pm@hub.local", "dev1@hub.local"}, dataset.Addresses())
	require.Equal(t, []mail.EngineFolderRole{
		mail.EngineFolderRoleInbox, mail.EngineFolderRoleSent, mail.EngineFolderRoleDrafts,
		mail.EngineFolderRoleArchive, mail.EngineFolderRoleTrash, mail.EngineFolderRoleJunk,
	}, dataset.RequiredFolderRoles())
}

func TestDemoDraftRejectsAttachmentIDs(t *testing.T) {

	err := saveDemoDraftWithAttachmentIDs(ctx, []uuid.UUID{uuid.New()})
	require.Error(t, err)
}
```

- [x] **Step 2: Reuse common runtime composition and Fake seed-folder capability** — accepted commits `a265bb3b` (common runtime refactor) and `ababa6e9` (Fake `PutFolder`).

```go
func (adapter *devdemoMailAdapter) ReconcileMailDemo(ctx context.Context, input devdemo.MailDemoInput) (devdemo.MailDemoResult, error) {
	// Reconcile permitted control-plane records, load fixed source, perform explicit bounded sync, and observe lifecycle.
}
```

The accepted common runtime composition and `mailbox_fake_seed.go` Fake `PutFolder` primitive are reused. Pending adapter work allocates one fake engine in the selector with deterministic UTC clock and opaque IDs, loads fixed PM/dev1 source, writes Ent only for permitted control-plane domain/account/identity prerequisites, and drives normal bounded sync. Folders/threads/messages/links/attachments arrive through projection. The `!qa && !release` attachment materializer rejects any non-empty attachment IDs; seeded received attachment uses fake `GetBlob`. Do not construct scanner, staging, S3, or external client.

- [ ] **Step 3: Commit the demo source/projection packet**

```bash
git add backend/internal/devdemo/mailbox_dataset.go backend/tests/dev/mailbox_dataset_test.go backend/cmd/server/devdemo_mail_adapter.go backend/cmd/server/wiring_calendar_mail_demo.go backend/internal/mail/mailbox_fake_seed.go backend/internal/mail/mailbox_draft_attachment_materializer_demo.go backend/tests/mail/mailbox_fake_seed_test.go backend/tests/mail/mailbox_draft_attachment_materializer_demo_test.go
git commit -m "feat(mail): seed deterministic dev mailbox through sync"
```

### Task 3: Drive one real Calendar invitation through the Mail delivery chain

**Files:**
- Modify: `backend/internal/devdemo/mailbox_dataset.go`
- Modify: `backend/internal/devdemo/mailbox_seeder.go`
- Modify: `backend/cmd/server/wiring_calendar_mail_demo.go`
- Test: `backend/tests/mail/mailbox_demo_calendar_delivery_test.go`
- Test: `backend/tests/calendar/calendar_invite_delivery_test.go`

**Interfaces:**
- Consumes: existing `wireCalendarOutboundRuntime`, `calendar.NewCalendarInviteDeliveryConsumer`, `mail.NewCalendarMailDeliveryService`, `mail.NewMailboxSubmissionDispatcher`, and `mail.NewMailboxSubmissionFinalizer`.
- Produces: the concrete `MailDemoPort` observes the one Calendar ID from `MailInvitationEventKey` only after ordinary delivery/finalization creates its linked Mail message.

- [ ] **Step 1: Write a failing delivery-chain test**

```go
func TestMailboxDemoInvitationUsesCanonicalDeliveryChain(t *testing.T) {

	eventID := seededEventIDs[devdemo.MailInvitationEventKey]
	require.NotEqual(t, uuid.Nil, eventID)
	require.True(t, repository.HasExactlyOneCalendarInviteDelivery(eventID))
	require.True(t, repository.HasFinalizedSubmissionForCalendarEvent(eventID))
	require.True(t, repository.HasReverseCorrelationFromFinalizedMessage(eventID))
	require.False(t, repository.HasSyntheticCalendarCorrelation(eventID))
}
```

- [ ] **Step 2: Implement one bounded chain, not a correlation shortcut**

```go
eventID, ok := input.CalendarEventIDs[devdemo.MailInvitationEventKey]
if !ok || eventID == uuid.Nil { return devdemo.MailDemoResult{}, errMissingDemoInvitation }
```

Start the existing outbound consumer/worker and submission dispatcher/finalizer with the same fake engine used by sync. The one keyed PM-to-dev1 event is created through Calendar service/outbox, admitted by `CalendarInviteDeliveryConsumer`, materialized, submitted, dispatched, and finalized normally. Its iTIP mail is attachment-less. Use explicit bounded observation; missing, duplicate, failed, or unresolved status is seed failure. Do not directly write `calendar_event_id`, Mail projection, inbound receipt, CalendarInviteDelivery, or MailSubmission state.

- [ ] **Step 3: Commit the canonical invitation packet**

```bash
git add backend/internal/devdemo/mailbox_dataset.go backend/internal/devdemo/mailbox_seeder.go backend/cmd/server/wiring_calendar_mail_demo.go backend/tests/mail/mailbox_demo_calendar_delivery_test.go backend/tests/calendar/calendar_invite_delivery_test.go
git commit -m "feat(calendar): deliver canonical dev invite through mail"
```

### Task 4: Add canonical Calendar occurrences to the existing aggregate — accepted `2d98fef8`

**Files:**
- Modify: `backend/internal/calendar/repository.go`
- Modify: `backend/internal/calendar/dto.go:9-36`
- Modify: `backend/internal/calendar/service.go:36-91`
- Modify: `backend/internal/calendar/event_ent_read.go`
- Test: `backend/tests/calendar/calendar_event_aggregate_test.go`

**Interfaces:**
- Consumes: `EventEntRepository` and the existing `Service.ListEvents(context.Context, uuid.UUID, time.Time, time.Time)` aggregate endpoint.
- Produces: `type CalendarEventAggregateRepository interface { ListCalendarEventOccurrences(context.Context, uuid.UUID, time.Time, time.Time) ([]CalendarEventAggregateOccurrence, error) }` and `func calendarEventAggregateResponse(CalendarEventAggregateOccurrence) EventResponse`.

- [x] **Step 1: Write aggregate tests** — accepted in `2d98fef8`.

```go
func TestListEventsAddsAuthorizedCanonicalOccurrences(t *testing.T) {

	response, err := service.ListEvents(ctx, dev1ID, from, to)
	require.NoError(t, err)
	require.Contains(t, response.Events, EventResponse{
		ID: "event-" + eventID.String() + "-2026-09-01T08:00:00Z",
		Type: EventTypeEvent, URL: "/calendar/events/" + eventID.String(),
	})
}

func TestListEventsOmitsCancelledAndLinkedScheduledCallDuplicate(t *testing.T) {

	response, err := service.ListEvents(ctx, actorID, from, to)
	require.NoError(t, err)
	require.NotContains(t, ids(response.Events), "event-"+cancelledID.String()+"-2026-09-01T08:00:00Z")
	require.Equal(t, 1, countStartsAt(response.Events, linkedOccurrenceStart))
}
```

- [x] **Step 2: Implement one batched, authorization-equivalent occurrence read** — accepted in `2d98fef8`.

```go
type CalendarEventAggregateOccurrence struct {
	EventID uuid.UUID
	StartsAt time.Time
	EndsAt time.Time
	Title string
	Description string
	AllDay bool
	ScheduledCallID *uuid.UUID
}

func (repository *EventEntRepository) ListCalendarEventOccurrences(ctx context.Context, actorID uuid.UUID, from, to time.Time) ([]CalendarEventAggregateOccurrence, error) {
	// one actor organization lookup; one event/attendee/exception bounded query plan;
	// then expand RRULE and apply exdates/occurrence cancellation in memory.
}
```

Use the same active-organization plus owner/organizer/active-attendee predicate as `ReadCalendarEvent`. Load only non-deleted, non-cancelled master events and their relevant master exclusions/cancelled exceptions in bounded batches, then expand recurrence only within `[from,to]`; apply `exdates` and omit cancelled occurrences. Return each occurrence with optional `ScheduledCallID`. In `Service.ListEvents`, build a set of linked call IDs before `expandCallEvents`, omit those legacy call projections, append canonical occurrences as `EventTypeEvent`, and retain the existing total ordering. The URL is exactly `/calendar/events/<lowercase-event-uuid>`; it must not be derived from Mail correlation. Do not query one event per occurrence, and do not expose attendee, iTIP, provider, receipt, or Mail data.

- [x] **Step 3: Commit the Calendar aggregate packet** — `2d98fef8`.

```bash
git add backend/internal/calendar/repository.go backend/internal/calendar/dto.go backend/internal/calendar/service.go backend/internal/calendar/event_ent_read.go backend/tests/calendar/calendar_event_aggregate_test.go
git commit -m "feat(calendar): show canonical events in aggregate"
```

### Task 5: Schedule a Calendar event from Mail recipients — accepted `21192503` and `aee705f4`

**Files:**
- Create: `frontend/src/modules/mail/mail-schedule-action.tsx`
- Modify: `frontend/src/routes/_authenticated/mail/$folderKey/thread/$threadId.tsx`
- Modify: `frontend/src/modules/calls/create-call-dialog.tsx`
- Modify: `frontend/src/locales/en/mail.json`
- Modify: `frontend/src/locales/ru/mail.json`
- Test: `frontend/tests/mail/mail-schedule-action.test.tsx`
- Test: `frontend/tests/calls/create-call-dialog.test.tsx`

**Interfaces:**
- Consumes: `resolveMailSchedulingParticipants(messageId, { addresses })` and existing `CreateCallDialog`.
- Produces: `MailScheduleAction` plus backward-compatible optional `CreateCallDialog` props `initialParticipantIds?: string[]` and `initialTitle?: string`.

- [x] **Step 1: Write RTL coverage** — accepted in `21192503`.

```tsx
it("resolves recipients and opens the existing call dialog", async () => {
	render(<MailScheduleAction message={message} />);
	await user.click(screen.getByRole("button", { name: /schedule meeting/i }));
	expect(resolveMailSchedulingParticipants).toHaveBeenCalledWith(message.id, {
		addresses: [message.from.email, ...message.to.map(({ email }) => email)],
	}, expect.any(Object));
	await waitFor(() => expect(screen.getByRole("dialog")).toBeVisible());
	expect(CreateCallDialogSpy).toHaveBeenCalledWith(expect.objectContaining({ initialParticipantIds: [dev1ID], initialTitle: message.subject }), expect.anything());
});
```

- [x] **Step 2: Implement the minimal action without a calendar redesign** — Mail scheduling accepted in `21192503`; Calendar UI integration accepted in `aee705f4`.

```tsx
const addresses = uniqueCanonicalAddresses([message.from, ...message.to, ...message.cc]);
const resolved = await resolveMailSchedulingParticipants(message.id, { addresses }, { signal });
setInitialParticipantIds(resolved.participants.flatMap((item) => item.user_id ? [item.user_id] : []));
setInitialTitle(message.subject.trim());
setOpen(true);
```

Render one localized action within an expanded message card. It derives deduplicated sender/to/cc addresses, resolves them with the existing authenticated Mail endpoint, and opens `CreateCallDialog` with only resolver-returned internal user IDs and trimmed subject. Extend dialog initialization/reset to copy optional initials for the open session; callers without values stay unchanged. Do not add Calendar APIs, a generic CalendarEvent dialog, a new availability query, custom endpoint, or separate navigation: `CreateCallDialog` remains authoritative for availability, `createCalendarMeeting`, invalidation, and navigation.

- [x] **Step 3: Commit the Mail scheduling UI packet** — `21192503`.

```bash
git add frontend/src/modules/mail/mail-schedule-action.tsx frontend/src/routes/_authenticated/mail/$folderKey/thread/$threadId.tsx frontend/src/modules/calls/create-call-dialog.tsx frontend/src/locales/en/mail.json frontend/src/locales/ru/mail.json frontend/tests/mail/mail-schedule-action.test.tsx frontend/tests/calls/create-call-dialog.test.tsx
git commit -m "feat(mail): open call scheduler for recipients"
```

### Task 6: Synchronize documentation and perform one bounded Docker acceptance run

**Files:**
- Modify: `docs/wiki/QUICKSTART.md`
- Modify: `docs/TECHNICAL.md`
- Modify: `docs/TESTING.md`
- Test: `backend/tests/mail/wiring_mailbox_demo_selector_test.go`
- Test: `backend/tests/mail/mailbox_demo_sync_test.go`
- Test: `backend/tests/mail/mailbox_demo_calendar_delivery_test.go`
- Test: `backend/tests/calendar/calendar_event_aggregate_test.go`

**Interfaces:**
- Consumes: completed Tasks 1-5 and the approved Docker-only `backend-unit` service.
- Produces: an operator-readable procedure plus sequential backend-unit, frontend-unit, and local full-stack manual-readiness evidence; it does not claim migration, Stalwart SMTP, QA, or production readiness.

- [ ] **Step 1: Add exact local activation and residual-limit documentation**

```markdown
Set `DEV_MODE=true` and `MAILBOX_DEMO_ENABLED=true` only in local Compose. The server rejects any selected demo mode without `DEV_MODE=true`; QA and release builds compile an inert selector. The demo uses an in-memory FakeMailEngine and no external mail, DNS, SMTP, S3, or scanner.

The demo proves local manual-readiness only. Production migration runner and external Stalwart SMTP exposure are separate packets and remain unimplemented here.
```

Document PM/dev1, six folders, seeded fake-blob download, attachment-less compose/reply/forward, one canonical invitation/backlink, aggregate event behavior, schedule action, and idempotent `/api/dev/seed`. State that normal Stalwart semantics/configuration remain unchanged and provide no production activation command.

- [ ] **Step 2: Review documentation paths and changed source against the frozen constraints**

Run: `git diff --check -- docs/wiki/QUICKSTART.md docs/TECHNICAL.md docs/TESTING.md backend/cmd/server backend/internal/devdemo backend/internal/calendar frontend/src/modules/mail frontend/src/routes/_authenticated/mail`

Expected: exit 0; the diff contains no whitespace errors and no `pro-qa-ai/**` path.

- [ ] **Step 3: Run backend verification alone in Docker**

Run: `docker compose -f docker-compose.local.yml run --rm backend-unit`

Expected: PASS for configured build/vet/unit/race, including new devdemo/Mail/Calendar contracts. Retain full output and do not run another heavy job concurrently.

- [ ] **Step 4: Run frontend verification alone after backend completes**

Run: `docker compose -f docker-compose.local.yml run --rm frontend-unit`

Expected: PASS for configured TypeScript/Vitest, including Mail action and CreateCallDialog default tests. Retain full output; do not overlap backend work.

- [ ] **Step 5: Run one local full-stack manual-readiness session after both unit jobs complete**

Run: `docker compose -f docker-compose.local.yml up --build -d`

Expected: stack starts with `DEV_MODE=true` and `MAILBOX_DEMO_ENABLED=true`. In one sequential logged session: call `/api/dev/seed`; dev-login as PM then dev1; inspect Mail folders/message detail; download seeded attachment; verify compose/reply/forward attachment-less behavior; inspect invitation/backlink; invoke Schedule meeting and complete existing call dialog; inspect `/api/calendar/events` recurrence/exdate/cancelled behavior and no linked-call duplicate; open canonical Calendar detail. Record actor/result. This is local full-stack evidence, not production evidence.

- [ ] **Step 6: Commit documentation and evidence references**

```bash
git add docs/wiki/QUICKSTART.md docs/TECHNICAL.md docs/TESTING.md
git commit -m "docs(mail): describe local demo readiness"
```

## Acceptance Criteria

- [ ] In `!qa && !release`, only explicit `MAILBOX_DEMO_ENABLED=true` plus `DEV_MODE=true` starts the demo runtime; selected mode without dev mode terminates startup; QA/release build stubs remain inert.
- [ ] Common composition may be reused while normal Stalwart semantics/configuration remain unchanged; the Compose diff is exactly backend `MAILBOX_DEMO_ENABLED: "true"`.
- [ ] `internal/devdemo` uses `MailDemoPort`, receives Calendar IDs by immutable plan key, and has no concrete MailEngine/Ent/server dependency.
- [ ] PM and dev1 receive deterministic, replay-safe Mail data containing inbox, sent, drafts, archive, trash, and junk; messages/folders/links/attachments were projected through bounded `MailboxSyncService`, and only seeded attachment download reads a fake blob.
- [ ] No QA controller/route/token, fake scanner, upload staging, external network client, direct Mail projection write, synthetic correlation, direct inbound receipt, or direct Calendar event-ID write exists in the dev path.
- [ ] Exactly one PM-to-dev1 canonical invitation reaches Mail through the ordinary invite admission, materialization, submission dispatch, finalization, and reverse-correlation chain.
- [ ] `GET /api/calendar/events` retains its response envelope and legacy kinds while additively returning authorized canonical occurrences with recurrence/exdates, cancelled omission, canonical direct URLs, and no duplicate for a linked `ScheduledCall`.
- [ ] The Mail reader resolves recipients then opens existing `CreateCallDialog` using backward-compatible optional initials; the dialog remains authoritative for availability, `createCalendarMeeting`, and navigation.
- [ ] Documentation limits the result to local manual-readiness, and evidence runs sequentially as backend-unit, frontend-unit, then local Compose build/up/manual checks for PM and dev1.

## Residual Limits

- Production migration runner is not introduced, invoked, or verified by this plan.
- External Stalwart SMTP exposure is not introduced, configured, or verified by this plan.
- Generated OpenAPI output is a separate documentation/generator synchronization note: regenerate and review it only when a completed API-surface packet requires it; no generated OpenAPI claim is made by the accepted runtime, aggregate, or UI commits.
- Local manual-readiness plus Docker/unit/full-stack evidence does not claim production, deployment, shared-database migration, external SMTP, or QA runtime acceptance.

## Execution Handoff

Autonomous subagent-driven execution is selected. Execute Tasks 1-6 in order with an independent review after each committed task; preserve frozen constraints and run the final Docker/manual stages strictly sequentially.
