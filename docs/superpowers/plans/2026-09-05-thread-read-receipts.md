# Visible thread read-through receipts Implementation Plan

> For agentic workers: use superpowers:subagent-driven-development. Root owns
> the contract, Docker gate, documentation and publication; Terra owns one
> bounded behaviour. Do not expand into backend or shared state changes.

**Goal:** Advance the existing followed-thread receipt from an answer whose
end is visible in the open, focused ThreadPanel, then refetch unread truth.

**Architecture:** A private ThreadReadReceipts controller observes end markers
owned by ThreadPanel. It uses existing identity-bound Query resources and
markThreadRead; one in-flight request and one coalesced candidate are retained.
The server remains authoritative for authorization and durable cursor CAS.

**Tech Stack:** existing React, TypeScript, TanStack Query, IntersectionObserver,
Shadcn Button and i18next; no dependencies or new global store.

**Spec:** Phase 151 151-CONTEXT.md and 151-SPEC.md (MSG151-UNR read-through,
MSG151-FOL follow modes, MSG151-CHN-02 capabilities), plus the root decision
below completing Tasks 140/144/154 of 151-PLAN.md. Baseline f1691e34.

## Global Constraints

- Docker-only canonical frontend-unit; root-only serial 2GiB/no-swap QA slot.
- No host tests/validators/build, image build/pull/install or Go retry at 2GiB.
- No Mail/Calendar, pro-qa-ai, shared UI, store, hook, API or backend edits.
- No optimistic unread arithmetic; four categories stay independent.
- Unknown/missing permissions fail closed; role inference is forbidden.
- Explicit downstream gpt-5.6-terra: xhigh implementation, high review.
- Commit exact paths with RU Conventional Commit, then push only
  calls-feature-codex to origin https://github.com/Valerijkk/hub.git.

## Root product and protocol decision

This is READ-THROUGH, not separate acknowledgements for every message.
Seeing an answer's end advances through older answers, including older
answers not individually seen. Never select a fetched-but-unseen newer
answer or claim that all older content was individually read. Per-message
or contiguous-reading semantics would require a different server contract.

Observe a one-pixel, aria-hidden, pointer-events-none absolute end marker in
each relative reply wrapper. Keep the existing MessageItem unchanged. Use
containerRef on the inner reply-list div, not the Radix root. Every end span
has data-thread-read-marker="" and data-message-id={reply.id}, with classes
pointer-events-none absolute bottom-0 left-0 h-px w-px. The controller selects
only these markers inside its own container, maps each entry back to the
CURRENT eligible replies by ID (never a captured message or DOM index), and
re-registers on reply-list or ready-scope changes. Removed/disconnected markers
and callbacks from prior observer generations cannot supply a witness.
Use
root:null, threshold:1, zero margins: viewport and ancestor overflow clipping
apply. This is geometric visibility, not eye tracking, reading-duration or
occlusion proof. Do not enable costly trackVisibility or add polling.

Only current rendered live direct replies with nonempty ID, matching channel
and thread_id, no deleted_at and a finite parsed created_at are candidates.
No parent/root-only receipt in this atom. No IntersectionObserver means no
receipt, not a latest-fetched fallback. Require document.visibilityState to
be visible and document.hasFocus() at observation and immediately at dispatch.
On blur/hidden clear unsent observations and disconnect; on focus/visible
re-observe current markers so a fresh browser callback supplies the witness.

Choose the greatest actually-observed candidate by Date.parse(created_at),
then raw ASCII ID comparison (not localeCompare). This only coalesces bounded
observations, never infers the durable server position or changes badges.
The server canonicalizes Mongo dates to milliseconds and UUID IDs are ASCII.
Maintain at most one in-flight, one greatest pending and one successful local
dedup watermark per mounted scope, not a history-sized queue/seen-ID set.
Repeated/equal/older observations after successful transport need no request.
Recheck a pending candidate remains a current eligible direct reply before
sending. No client-selected unseen latest message and no DOM-order assumption.

Before dispatch require all of:
1. Same mounted user/channel/root scope and live auth user; not revoked.
2. Current identity channels query and detail query success, idle, non-error;
   channel exists and capabilities.is_member, can_read, can_view_unread are
   exactly true. Detail parent is this nondeleted root in this channel.
3. Detail follow_mode exactly auto or manual_follow; no alias/unknown mode.
4. Current public reply query data exists, success/non-error/non-fetching;
   masked pending-recovery data is never a source of receipt candidates.
5. Focus/visibility and an observed, still-valid candidate as above.

Identity/channel/root changes, revoke or unmount invalidate old callbacks and
clear unsent work. Use a scope-keyed controller plus lifecycle guard; A-B-A
does not resurrect an old request. Gate lost/refetch clears unsent work and
re-observes on recovery. Do not cancel or claim cancellation of sent requests.
Keep the owned in-flight guard until its refresh settles. After transport
success (advanced true OR false), current lifecycle/auth/revoke alone gate
exact invalidation of unreadQueryKey(userId), exact:true. Do not reuse idle
resource gates here: background refetch must not swallow a successful receipt.
Never invalidate all users, channel read, mention read or thread detail.

After fulfilled markThreadRead, record the local dedup watermark before
awaiting unread refresh. A rejected refresh does not undo that transport
success, show a false receipt-failed message, or replay the confirmed ID;
consume that rejection and release single-flight only when refresh settles.
The unread query retains its normal authoritative recovery lifecycle.
On markThreadRead transport error, do not advance the local success watermark
or change counts.
Pause automatic sends and show a compact localized inline error with explicit
retry. Retry clears that failure and re-observes current visible candidates;
it does not replay a stale/unseen ID. No timer, automatic retry, toast storm
or unhandled promise. Obsolete-scope errors are silent.
Frozen messenger locale keys:
- threadRead.error: EN Unable to mark thread replies as read;
  RU Не удалось отметить ответы прочитанными.
- threadRead.retry: EN Retry marking as read; RU Повторить отметку прочитанного.
Use existing Button type=button, role=alert on error, keyboard activation.

## Task 1: one ThreadPanel visible read-through behaviour

**Files:**
- Create frontend/src/components/messenger/thread-read-receipts.tsx: private
  single-consumer controller, aim <=300 production lines. If substantially
  larger, return BLOCKED for decomposition instead of general infrastructure.
- Modify frontend/src/components/messenger/thread-panel.tsx: retain the whole
  useThreadMessages result, reply container ref/end markers and guarded
  scope-keyed controller mount only. Preserve independent reply rendering,
  follow control, close/focus return, issue links and composer.
- Modify frontend/src/i18n/locales/en/messenger.json and ru/messenger.json:
  only the frozen threadRead keys.
- Create frontend/tests/messenger/thread-read-receipts.test.tsx: render real
  ThreadPanel with real QueryClient/auth/revoke/hooks; mock HTTP and browser
  observer boundary plus unrelated child components, not the target logic.

**Interfaces:** consume markThreadRead(channelId, rootMessageId, messageId),
ThreadResponse, Message[], channelsQueryKey, threadDetailQueryKey and
unreadQueryKey. Produce ThreadReadReceipts with exact props: channelId,
rootMessageId, userId, replies:Message[], containerRef:RefObject<HTMLDivElement
| null>, detail:ThreadResponse, detailReady:boolean, repliesReady:boolean.
ThreadPanel supplies ready flags from public hook data/status/fetch fields;
controller additionally checks live query/store barriers before dispatch.
Mount only with authenticated nonrevoked successful current detail, not while
initial detail loading/error. Readiness false disables observation rather
than dropping existing rendered replies. Reuse useChannels() deduplicated
identity query, never add a separate endpoint or use unscoped store channels.

- [x] Write behavioural tests first. Existing ThreadPanel imports must resolve;
  expected RED is absent markThreadRead after a visible reply, not an absent
  new-module import. A minimal representative assertion is:

  ```tsx
  render(<ThreadPanel channelId="ch-1" messageId="root-1" onClose={() => {}} />,
    { wrapper });
  await screen.findByText("reply-1");
  expect(markThreadRead).not.toHaveBeenCalled();
  // Harness emits ratio 1 for the reply end marker, only if actually observed.
  await emitVisibleReply("reply-1");
  await waitFor(() => expect(markThreadRead)
    .toHaveBeenCalledWith("ch-1", "root-1", "reply-1"));
  ```

  Matrix: no observer/no visibility/no focus; no initial eager receipt;
  ratio<1 and nonintersecting ignored; exact visible direct ID only; unknown/
  manual_unfollow/absent mode denied; false/missing capabilities, absent
  channel, detail/replies loading/error/refetch denied; malformed/deleted/
  wrong-scope replies denied; greatest visible tuple independent of DOM order,
  millisecond tie by ID, newer unseen never selected; single-flight/coalescing
  and successful duplicate suppression; advanced:false exact unread refresh;
  background refetch after send does not swallow success; blur/focus fresh
  witness; unread refresh rejection does not become receipt error or resend;
  auth/root A-B-A/unmount/revoke ignore late completion; error pause,
  localized explicit retry and no unread-store arithmetic. Isolated test
  clients cleaned up, real timers, no network or live server.
- [x] Independent preflight of plan/oracle; root rules on any contradiction.
- [x] Root runs docker compose -f docker-compose.local.yml run --rm frontend-unit
  after fresh workload check. Preserve only small current report files.
- [x] Only after observed intended RED, implement minimal controller and parent
  glue. Protocol sketch (not an additional public abstraction):

  ```text
  observer callback -> greatest(valid fully-intersecting entries)
    -> merge one pending candidate -> pump if idle and dispatch gates pass
  pump -> markThreadRead -> same scope? exact unread invalidation -> dedup
    -> release single flight -> pump one remaining valid pending candidate
  failure -> pause + localized retry -> fresh observer witness, never timer
  cleanup -> increment lifecycle, disconnect, clear unsent work/listeners
  ```

- [x] Independent spec-aware/conclusion-blind Terra/high review, bounded fixes.
- [x] Canonical Docker GREEN for TypeScript and full genuine Vitest suite.
  Browser layout/visibility and backend/API/storage remain NOT_RUN until their
  separate real Docker SUT gates; unit simulation is not E2E acceptance.
- [x] Root updates this plan, STATE, COVERAGE, system specification, test plan,
  TESTING and affected QUICKSTART/SHORTCUTS in the same session; exact commit
  and immediate authorized push, verify remote equality.

## Design evidence and limits

mongo_unread.go uses phase151MongoCanonicalTime and tuple CAS; dto.go stores
MessageDocument.CreatedAt as bson created_at; mongo_repository.go assigns
UUID message IDs. ListThreadReplies is timestamp-descending only and WS may
append, so DOM index is not a receipt ordering oracle. Public thread hook
masks data during pending-recovery overflow; that barrier must survive UI glue.
Primary references: [MDN intersection clipping](https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API#clipping_and_the_intersection_rectangle)
and [MongoDB BSON Date](https://www.mongodb.com/docs/manual/reference/bson-types/#date).

Root self-review: this single task covers the above UI addendum, not all of
Phase151. No external API, backend permission change or separate read category
is introduced. Exact existing signature and aliases verified at baseline.
GSD state/ledger are maintained by ordinary file edits: host Node GSD helpers
are not an exception to Docker-only. User already selected autonomous SDD;
no repeated execution-choice approval is needed.

Status: UNIT_ACCEPTED after independent review and canonical Docker GREEN;
published d65da37da94c97702be18fed351a0f9294fd4b34 to the authorized origin
calls-feature-codex with remote equality verified. Browser/API/storage/product
acceptance NOT_RUN.

Independent Terra/high preflight C/I/M0/2/0: root accepted I-01 exact marker
registration and current-reply lookup, and I-02 separating fulfilled receipt
transport from unread refresh rejection. Both are incorporated above before
production, with an additional refresh-rejection oracle. Review did not inspect
the then-absent test file and is not runtime acceptance.

## RED / triage checkpoint

Final pre-RED oracle review and its scoped re-review: C/I/M0/0/0, requested
Terra/high, actual model telemetry UNKNOWN. Root ran only canonical
frontend-unit using existing image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Outer b7055d96ed0e, engine
e740b18b114ea3a293fc3a116d9a9887c936a67cc2df1df540187190a86045e4,
started 2026-09-05T20:47:27.272708197Z; inspected2GiB/no-swap/6CPU/readonly.
TypeScriptPASS; Vitest638PASS/23FAIL,661cases/191files,73.70s; Compose exit1.
The inspected container is absent after the run. Its closed lifecycle events
were not returned by subsequent bounded queries, so engine-die/OOM provenance
is UNKNOWN for this RED, not a claimed no-OOM receipt.
Only eight small current result/typescript/vitest artifacts were preserved in
testframework/reports/frontend-unit-thread-read-red-20260905; old Biome/Vite
files, caches and tmpfs content were not copied.

Of31 new cases,17 failed for the intended missing receipt/observer behavior.
One failed earlier in a mixed fixture: an empty ID makes the REAL existing
isCachedMessage/isMessageCache reject the whole public page, hiding the valid
control reply. ROOT=test, not product code. The correction removes that row
only from the mixed positive-control case, resets its reused HTTP call counts,
and tests the empty-ID masked page separately. Existing fail-closed hook and
all positive receipt assertions remain unchanged. Final32-case oracle hash:
86245FD8BFEFF4211063D3B3AD0E4DF7116B6CE8F1BBF05C9371A8CE766DAED8.
Independent scoped fixture/hook review C/I/M0/0/0. The unchanged positive
read-through anchor has observed RED; fixture corrections and the separate
negative will next run with production GREEN, not a redundant whole-suite RED.

Five unchanged admin/call-dialog/composer cases timed out at15s (actual18s).
Four also timed out in the earlier follow race RED; all five passed in its
final GREEN, before this test existed. Independent read-only triage found no
direct test-file imports but cannot prove worker/process isolation from these
reports; actual cause remains UNKNOWN (flaky/infra candidates). No assertion,
timeout, worker profile, dependency or service selection was changed.
Full canonical GREEN must include these cases unchanged; no product code is
being changed on the basis of those timeouts.

Pre-run checks: all adjacent Mail builds completed, no visible heavy test job;
resident healthy Mail services stayed untouched, low CPU; free RAM7092096KiB,
C: free301897482240bytes, one disk sample0%/queue0. These are snapshots, not
peak-memory/load or future-exclusivity guarantees. No build/pull/install ran.

## Initial production review / race RED (2026-09-06 local)

Independent review plus root challenge identified four bounded source points:
missing scope-keyed mount; missing pump-time fulfilled-watermark check;
observation eligibility not explicitly following capability/follow/root gates;
truthy parent.thread_id not rejected. Reviewer's original exact A-B-A before
passive-cleanup timing claim was withdrawn as not reproduced; do not present
it as runtime proof. Root freezes keyed mount plus tiny committed lifecycle
layout setup/cleanup, leaving browser observation passive. Ref/lifecycle
mutation during render is not the remedy. Latest visible data refs should
reflect committed props. The layout lifecycle decision follows
[React effect timing](https://react.dev/reference/react/useLayoutEffect).
Marker-free position watermark is optional bounded hygiene, not a leak claim.

Added four independently reviewed genuine regression cases; oracle hash
16891C6CE73EA0F98467973DF9700B8A283643A6816F4C1D741624022BCA4C91.
Same-props RTL rerenders after denied/recovered cache writes commit both gate
states instead of letting query notification batching elide the transition.
Canonical frontend-unit engine
064fc9a929d8f7b9df10a85c1d40cbad877423dac031893344f2a9ef204e760e,
outer d8c7e9ea13a9, same existing image286783...218, started
2026-09-05T21:37:28.317431341Z. Inspected2GiB/no-swap/6CPU/readonly.
TypeScriptPASS; Vitest662PASS4FAIL/666cases,191files,60.33s. Compose/engine
exit1, engine duration77s, create/start/die/destroy events captured, no OOM
event observed in the subscribed lifecycle. Eight small current artifacts:
testframework/reports/frontend-unit-thread-read-race-red-20260906.

Two intended REDs confirm ROOT=code: equal/older pending sends twice and
nested parent allows a receipt. Two remaining failures are in the new
focus/refetch recovery oracle; source versus premature observer witness is
being triaged separately. All unchanged tests pass in this run. Both new
gate-loss cases pass already, so their explicit source gate is contract
hardening, not a claimed reproduced runtime failure. No suite timeout,
worker profile, assertions or non-task source were relaxed.

Root triage of the other two failures: ROOT=test, independently confirmed.
emitReply silently returns when no current marker is registered; resolving
refetch or dispatching native focus does not synchronously complete a passive
observer effect. Only these two oracle cases now capture the prior observer,
wait for a distinct active registered marker, then emit one fresh ratio1
witness. Focus events are act-wrapped. No production synchronization, retry,
timeout change, repeated witness emission or weakened receipt assertion.
The separate two confirmed source bugs and explicit lifecycle/gate remedies
were applied by their original owner via root patch actuator. Re-review and
canonical GREEN remain pending.

## Final bounded acceptance (2026-09-06 local)

Independent final source review closes all four findings: C/I/M0/0/0,
requested Terra/high, actual model/provider telemetry UNKNOWN. Root accepted
the two condition-based observer oracle corrections without changing their
assertions or production timing. Canonical frontend-unit final GREEN:
- image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218;
- engine c68613e366df5157482f6b65693457156387b4fc3e78229b320e7c4e83a8364e,
  outer009b8bd3609e, start2026-09-05T21:46:46.537466365Z;
- inspected2GiB/no-swap/6CPU/readonly, Compose and engine exit0;
- TypeScriptPASS,191files/666tests PASS, zero failed/pending,58.20s;
- create/start/die/destroy captured, engine75s, no OOM event in lifecycle;
  container absent after completion.

Unchanged pre/post-run hashes:
- controller EED29ACDA9E95956046530746624458CF9DD5B629BEF65492552D91C1592ECAE;
- parent 6FFEF3ECADFAC6823AE0971C13CB607788519D4AD7410F93EBB213CD0F5F0409;
-36-case oracle 96E4B5EB95318DEFC848417683076B449384CDD371245D9BD5040AF962897A3E.
Eight current small report files are retained in
testframework/reports/frontend-unit-thread-read-green-20260906.
No source edits, build/pull/install or other QA job ran during this gate.
Mail services had ended; existing Mail builder history showed completed builds
and was left untouched. Pre-run free RAM7243204KiB, disk15%/queue0: snapshots,
not a claim of continuous machine resource telemetry.

This accepts ONLY the visible read-through UI atom with its36 genuine units,
not backend transport durability, real browser geometry, the Messenger module,
Jitsi activation, or communications end-to-end. Single retained watermark
marker is bounded-one state; no history queue or unbounded leak was found.
