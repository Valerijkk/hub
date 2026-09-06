# Panel unpin command lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Root alone owns canonical Docker QA, docs and publication.

**Goal:** Removing a pin in the open panel has one request owner, handled current failure, scoped exact refetch and obsolete-settlement isolation.

**Architecture:** A separate usePanelUnpinCommand hook preserves panel-specific list+pin authority. One request per panel; no per-pin Map or generalized command framework. Existing GET/query UI and accepted row-pin hook remain unchanged.

**Tech Stack:** Existing React, Zustand, TanStack Query, i18next, Vitest/RTL and toastApiError; no dependencies.

**Spec:** docs/TECHNICAL.md POST pin/GET pins, docs/system_specification.md channel_removed barrier, source-confirmed panel-unpin entry in the completion queue. No new server roles or permission vocabulary.

Status: UNIT_ACCEPTED / PUBLISHED; independent source and final receipt
review C/I/M0/0/0. Commit25c58baaa001911c61f700e0cbfdb00dd99441d5 was
immediately pushed to the approved origin calls-feature-codex; remote SHA
equality verified on2026-09-06. This is not full-module acceptance.
The following prerequisite/oracle-stage notes are historical.
Forward prerequisite published as 74c5c492a41998213fcf7cf1556fdf6333c7ee42,
immediate push and exact remote SHA equality verified on 2026-09-06.
READY_FOR_ORACLE at that clean base: TypeScript + 201 files / 871 tests PASS.
Panel production remains unchanged until accepted RED. Root batching ruling:
the independent bookmark-removal oracle may share the same canonical full
RED/GREEN run; disjoint ownership and per-file failure triage remain mandatory.
If both frozen oracles are present, collection is 203 files / 915 tests.
No tests change during a running gate; no broader product acceptance follows.

Discovery baseline ac918830e2aa0d7082297d9913be0fd750013fb5. Bind exact published forward prerequisite HEAD before this oracle is applied. Forward acceptance is not presumed.

## Global constraints and ownership

- One behavior: PinnedMessages.handleUnpin command. Do not change row pin, forward, bookmarks/reactions, API, cache keys/shapes, previews, GET enablement, filtering, visibility, loading/error/empty states or server policy.
- New local frontend/src/components/messenger/use-panel-unpin-command.ts and focused frontend/src/components/messenger/pinned-messages.tsx adapter only. sum(added + deleted) <= 250 production lines across both files.
- New frontend/tests/messenger/message-panel-unpin-lifecycle.test.tsx only: <=600 physical lines, 34 genuine cases. All old tests unchanged.
- Preserve 12 cases: phase151-pinned-capabilities (5), phase151-pinned-identity-boundary (4), phase151-pinned-query-state (3). Preserve a11y-names-p2 direct trigger-name oracle.
- The latter identity/query-state API mocks export only getChannelPins/togglePin. Hook must use direct .capabilities booleans, not introduce canChannel imports or widen old fixtures.
- Reuse messagePin.error locale key and canonical formatter precedence; no locale changes.
- Root owns this plan, completion queue, STATE/COVERAGE and affected technical/testing/wiki docs.
- Explicit Terra/xhigh atomic writers, Terra/high reviewers; model/provider UNKNOWN unless observed, no fallback. No children. You are not alone; one writer per exact file, do not revert others.
- No pro-qa-ai or Mail/Calendar reads/edits, no other-container changes, no host tests/build/linters/validators.
- Root canonical command exactly: docker compose -f docker-compose.local.yml run --rm frontend-unit.
- Serial full Docker gate using existing image, 2 GiB/no-swap, 6 CPU/two workers, unchanged assertions/timeouts, no build/pull/install. Wait for competing heavy Mail builds to finish without altering them.
- Exact-path commit, immediate push only calls-feature-codex to origin https://github.com/Valerijkk/hub.git, verify SHA. No master/merge/force/tags.
- Go 2 GiB compile and Jitsi infra blockers remain recorded; do not retry or infer authority.

## Existing safeguards and precise gap

PinnedMessages currently checks actor, prop/live can_list_pins AND can_pin, revoke and pin.channel_id before togglePin, then current actor/rights before exact invalidation. It lacks a synchronous pending owner, caught error and committed/transient lifecycle fencing. Keep all existing checks.

Row-pin hook requires only can_pin and message-channel binding; adding panel list/open policy there would alter its accepted contract. Separate panel hook is intentionally local.

PinnedMessage fields are required strings id/message_id/channel_id/pinned_by/pinned_at, not readonly. Capture the first three values for transport identity; display/provenance fields do not authorize mutation.

Transport risk: the current frontend API is togglePin, not an idempotent unpin endpoint. A local lock does not resolve an ambiguous network failure after a server-side effect; explicit retry can toggle again. No automatic retry, durable dedupe, abort or rollback promise. Server idempotency requires a separate backend contract and its blocked Go/runtime gate.

## Frozen interface

~~~ts
type PanelUnpinOptions = {
  open: boolean;
  channelId: string;
  canListPins: boolean;
  canPin: boolean;
};
export function usePanelUnpinCommand(options: PanelUnpinOptions): {
  pending: boolean;
  submit: (pin: PinnedMessage) => Promise<void>;
};
~~~

## Frozen behavior

1. Hook owns auth selector, QueryClient, translation, committed scope and synchronous request refs. A layout effect creates a fresh scope for committed actor/open/channel/prop canListPins/canPin changes. Cleanup invalidates its scope AND request, and idempotently unsubscribes request listeners. New scope resets pending. No render-time ref writes.
2. Admission: mounted current open scope, nonempty actor/channel/pin.id/pin.message_id/pin.channel_id, exact pin.channel_id===scope.channelId, both prop permissions, exact live actor, existing non-revoked live channel with both capabilities true.
3. Claim a request synchronously before first await, capturing scope plus immutable pin ID/message ID/pin channel values. All panel Unpin buttons share pending; two same-tick clicks on different pins produce one POST. Other message-row controls remain unaffected.
4. Accepted request temporarily subscribes to auth and messenger stores before dispatch. Any observed actor/source list/source pin/revoke loss irreversibly marks it invalidated, even loss+restore within one React act. No setState/ref release inside subscription; unrelated channel changes or query rows do not reset the request.
5. Current effects require scope/request identity, !invalidated and fresh authority. Separate identity-only ownership controls pending cleanup: an invalidated but still owned request can release itself on settlement; an obsolete request cannot unlock a newer scope. Release listeners on every settlement and scope cleanup.
6. Call togglePin(capturedChannelId,capturedMessageId). Current success awaits only queryClient.invalidateQueries({queryKey:pinsQueryKey(capturedChannelId,capturedActorId),exact:true}). Explicitly narrow capturedActorId before passing its string key. Do not set throwOnError, invent PinnedMessage data from the boolean result, mutate preview/list caches or invalidate a broader key.
7. Pending lasts through the actual query invalidation/refetch. Completion of an old invalidation cannot release a newer request. Already-dispatched refetch may settle in its own key; do not claim to undo it.
8. Current error calls toastApiError(error,t("messagePin.error")) once; EN argument "Unable to update pinned message", RU "Не удалось изменить закрепление сообщения". Stale error is consumed silently. Finally safely releases only owned pending. Current failure permits later explicit retry.
9. No panel close on success; keep existing UI/list refetch. Old success/error/unmount/ABA cannot notify, begin obsolete invalidation or unlock a newer attempt.

## Adapter map

PinnedMessages retains useQuery/useQueryClient for GET and preview; retain useEffect/useState and userId/revoke/query state. Remove only now-unused useCallback/useRef/capabilitiesRef/canUnpinNow/handleUnpin/togglePin import. Instantiate hook before early return with open/channelId and the existing derived canListPins/canPin values; this preserves revoke and prop visibility behavior. Wire pending only to the existing Unpin Button.disabled and submit as its handler. No query or Dialog restructuring.

## Concrete oracle inventory — 34 new genuine cases

Real PinnedMessages, Zustand, QueryClient and QueryObserver/actual component query; mock only HTTP/formatter/unrelated presentation. No future-hook import, forced refs, source reading, timeout bumps or global rejection swallowing.

- 1 positive exact toggle/key; deep independent own/other-user/other-channel/timeline snapshots prove no optimistic mutation; no panel close.
- 2 EN/RU current failure plus explicit retry, exact formatter argument/count and second POST.
- 1 same-tick two different pins: one request, all Unpin buttons pending.
- 5 ID/binding denials: no actor, empty channel, empty pin ID, empty message ID, foreign pin.channel_id. Use same-act captured controls where UI otherwise hides; never click a detached hidden control.
- 4 fresh live denials: absent channel, can_list_pins false, can_pin false, revoked.
- 6 committed actor/channel/open-session ABA times success/error: new request started before old settlement; newer pending protected.
- 8 prop can_list_pins/prop can_pin/live can_list_pins/live can_pin loss+restore times success/error. Store loss+restore within one act proves subscriber memory; prop changes commit separately. Reopen after list-prop loss if the actual UI closed. Silent old settlement, explicit retry.
- 2 revoke+restore success/error; respect existing Dialog close/visibility effects.
- 2 unmount success/error, no toast or unhandled runner error.
- 1 real deferred QueryObserver invalidation: pending spans refetch; prior-scope invalidation cannot unlock newer pending.
- 2 batched actor A-B-A success/error without intermediate React commit; stale settlement and explicit retry.

Count 1+2+1+5+4+6+8+2+2+1+2=34. Baseline full count must be rebound after forward publication; do not invent a panel total now.

~~~ts
act(() => {
  unpinButtons[0].click();
  unpinButtons[1].click();
});
expect(togglePin).toHaveBeenCalledTimes(1);
expect(unpinButtons[0]).toBeDisabled();
expect(unpinButtons[1]).toBeDisabled();
~~~

Local window unhandled spies are supplemental; full Vitest exit/raw headings are required. Report a RED failure at a precondition honestly; it does not execute later assertions.

## Task 1: Panel command

- [x] Independent contract review C/I clear; bind actual forward prerequisite publication.
- [x] One Terra/xhigh oracle author, root full read/hash and independent review.
- [x] Root full canonical RED, preserve eight receipts and triage source vs fixture/runtime.
- [x] One Terra/xhigh source writer for only two production paths within budget (234/250).
- [x] Independent source/oracle review; no weakened assertions or oracle drift (C/I/M0/0/0).
- [x] Root full canonical GREEN, exact provenance/resource evidence (final independent receipt review recorded below).
- [x] Docs sync, exact source commit/immediate authorized push/remote equality; remaining source work follows separate gates.

Every dispatch max20minutes, depth/fanout0, max3 fixes/root cause /5 review rounds /1 transient retry. If bounds or authority do not fit, return BLOCKED for root decomposition. No full Messenger/API/storage/browser or Jitsi readiness claim from these units.

## Canonical combined RED — 2026-09-06

Panel oracle34cases/440physical lines,
SHA21CE66893687CB7C40C8B1B94A6B015487E031F9FDE422C7DB3EC4A14F7B73C8;
root full read and two fixture-only corrections before independent
READY_FOR_RED. No old tests or production changed. Canonical TypeScript
PASS;203files201PASS/2FAIL,915tests887PASS/28FAIL,0pending/todo;exit1,
126.19s,start11:13:25UTC. Panel14PASS/20FAIL, bookmark2PASS/8FAIL;
all201oldfilesPASS. Nine raw unhandled rejections are the uncaught panel
handler's currentEN/RU, prop/live list/pin, revoke, unmount and batched actor
failures. Independent triage agrees ROOT=code per command, C/I0/0,
READY_FOR_SOURCE; this is source-stage authority, not product acceptance.

Panel20:2formatter,1duplicate,3emptyIDs,6ABA pending witnesses,4permission
restore-success invalidations,1revoke-success invalidation,1unmount-success
invalidation,1real-refetch pending,1batched-actor-success invalidation.
SixABA cases dispatchrequest2, then fail disabled BEFORE old settlement;
later isolation assertions NOT_EXECUTED. Real QueryObserver invalidation
starts, but its disabled failure precedes the later ABA/finally witness.
Prop/live test title omits outcome display; third-column row data actually
runs both outcomes. No post-RED test changes are authorized by this note.

Eight receipts: `testframework/reports/frontend-unit-panel-actions-red-20260906/`
(result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log).
One permission-review timeout occurred before execution; one allowed retry
launched the sole runner `worktree-master-frontend-unit-run-872844fdca3b`,
engine07542988fd019512e18a53dc366cc3964facfe6a88a4e1dfea6e31899ad8b496,
started2026-09-06T11:13:07.778600029Z. Actual2GiB/no-swap6CPUreadonly,
two workers; image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Samples365.1/474.6MiB,CPU170.97/215.86%,IO0B; not peaks. Host disk second
sample0%/queue0; freeRAM3,839,812KiB at first runtime sample. Mail QA stack
untouched. Disposable container absent after exit; engine completion/OOM
NOT_CAPTURED. Build/lint/API/browser/storage NOT_RUN. Oracle/source/Compose
hashes match before/after; raw RED retained without overwrite.

## Canonical composed GREEN — 2026-09-06

At base74c5c492 plus both reviewed panel/bookmark atoms, root ran exactly
`docker compose -f docker-compose.local.yml run --rm frontend-unit`.
TypeScript PASS;203 files /915 tests PASS, including34 panel and10 bookmark
cases; exit0,136.52s,start11:37:40UTC. Failed/pending/todo and whole-run
unhandled0; JSON423 nested suites is not the file count. All later ABA and
refetch assertions unreached in RED now passed. Existing act/DialogContent/
mock warnings remain; build/lint/API/browser/storage/media NOT_RUN.

Panel production +178/-56=234/250, hook168 physical lines. Source review
Terra/high C/I/M0/0/0, spec and quality compliant. Author Terra/xhigh; actual
runtimes UNKNOWN. Hook SHA EF2DF7F657B8A728DA614A696AC3B4124182F433B22DA127CFAC795D34108BB5;
adapter ED7CFFF80838C46B039FEB1B54F0607A12AFEF3DC7D0F3ACDDF05FC3E5D52C81.
Frozen oracle, all four old panel/bookmark fixtures and Compose hashes match
before/after. No source/test/config changes during execution.

Eight preserved receipts at
`testframework/reports/frontend-unit-panel-actions-green-20260906/`:
result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log.
RED retained separately. Root detailed provenance/reviews are in
`.superpowers/sdd/2026-09-06-panel-actions/green-receipt.md` and source-review.md.
Separate source commits use this composed candidate evidence; intermediate
single-atom commits are not independently executed.

Runner `worktree-master-frontend-unit-run-ceb04593a18b`, engine
1c862c2fb437daaa680b4dcb7b9947be589d13afdaf5048802de15193bcd7066,
started2026-09-06T11:37:22.127807906Z; same existing image286783701fc3….
Actual2GiB/no-swap6CPUreadonly, two workers; samples511.6/288.8MiB,
CPU233.84/182.49%,IO0B; host disk3%/4%,queue0. Samples are not peaks.
Final independent receipt review: UNIT_ACCEPTED / READY_FOR_PUBLICATION,
C/I/M0/0/0; raw counts and absence of whole-run unhandled confirmed.
Disposable runner absent after exit; engine completion/OOM NOT_CAPTURED.
No root build/pull/install or Mail container changes. Preflight11:36:43 had
no active Mail build and4297192KiB freeRAM. Mail independently began a
python-tests build11:39:14-11:39:53 during the tail; continuous exclusivity
is not claimed. Later Mail builds were active; root starts no further QA.
