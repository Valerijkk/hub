# Message row pin command lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Follow the checkbox steps. Root alone owns canonical Docker QA, docs sync and Git publication.

**Goal:** Pinning from a message row has one current in-flight command, handled errors and exact current-user/channel server refresh.

**Architecture:** Add one local row-pin command hook beside MessageItem and thin pending wiring. Retain the existing pinsQueryKey and server refetch; do not alter the separate PinnedMessages unpin handler or introduce a global action framework.

**Tech Stack:** Existing React 19, TypeScript, Zustand, TanStack Query, Vitest/RTL and toastApiError; no dependencies.

**Spec:** docs/TECHNICAL.md POST /messenger/channels/:channelId/messages/:messageId/pin and GET /messenger/channels/:channelId/pins; docs/system_specification.md channel_removed barrier; source-confirmed row-pin finding in 2026-09-05-communications-completion-queue.md. This frozen lifecycle supplement does not define new backend rights.

Status: UNIT_ACCEPTED; independent final review C/I/M 0/0/0. Canonical
TypeScript + 200 files / 835 tests PASS, 26 pin cases, exit 0, 117.97s.
Published ac918830e2aa0d7082297d9913be0fd750013fb5 with immediate authorized
push and exact remote SHA equality verified. API/browser/full module
acceptance remains NOT_RUN. The earlier checkpoints below are chronological
history, superseded by the final GREEN receipt at the end of this file.

Corrected oracle applied:
frontend/tests/messenger/message-pin-lifecycle.test.tsx,423physical lines,
26cases, SHA256 E021C6C9C7EC0A4C3895D5197FA7ACFA014BDCDB777EB52CDFF0D0F062E2C65D.
Root read the full file and verified production/Compose and the six old
row-cache cases unchanged. Independent oracle review C/I/M0/0/0,
READY_FOR_RED. Root canonical full200file/835test RED started03:58UTC.
Live container8eff96ce39fc,engine51e861947c87c878995195af3d8990808fb77122730e9a7b598980d4576c72c5,
started2026-09-06T03:58:28.940206244Z; sameimage286783..., inspected2GiB/
no-swap,6CPU,readonly,MAX2/MIN1 workers declaration. First sample at Vitest
startup200.5MiB/134.45percentCPU/IO0B; workers not yet spawned in that top.
Preflight activeMailBuilds0,onlyidleMailBuildKit,15269976KiB freeRAM,disk0/queue0.
No source/Compose edits,build,pull,install or assertion/timeout changes.
Outcome pending; do not mark expected RED as completed evidence.
Prerequisite publication is now bound: reaction0d562ee855abfe1180646fe81a5a1ab8008cd475,
bookmark1877984e15462a4cc05ecb8cead2ca5810ac8ce2, both pushed and exact remote
verified. Current pin baseline is1877984e, production unchanged from the
accepted composed199file/809test GREEN. Root authorized ONLY new pin oracle
application before its independent review and RED; no pin production yet.

Draft oracle review corrections before application: compact repeated fixture
boilerplate to honor550lines, keep supported distinct parameterized titles,
isolate channel ABA without changing message ID, and prove explicit retry's
exact second POST plus single exact invalidation/formatter count. No existing
test assertion or case is removed. Author's corrected count is438lines before
the final retry-count assertion; root must measure the applied file.
Source discovery baseline c0900e622964bd48541e8929b680801e82f550d0,
calls-feature-codex, linked worktree. Reaction and bookmark source atoms
precede this production writer. Recheck actual published HEAD and source
hashes before the pin oracle RED; do not present c0900e62 as that future gate.

## Global constraints

- One behavior: MessageItem row togglePin lifecycle only, not panel unpin.
- No optimistic pin state/list update, synthetic PinnedMessage from boolean,
  server rights, API payload, query-key, WS or global command abstraction.
- Row pin requires can_pin only; can_list_pins is NOT a new row prerequisite.
  PinnedMessages legitimately requires both and remains unchanged.
- No automatic retry, success toast, abort/rollback guarantee or cross-row/tab
  duplicate guarantee. Already dispatched HTTP may still affect the server.
- Preserve reaction/bookmark/edit/delete/SEND and every old test assertion.
- Host reads/edits/Git only. Root QA command is exactly:
  docker compose -f docker-compose.local.yml run --rm frontend-unit
- Existing image,2GiB/no-swap,6CPU,two workers, full tests/timeouts unchanged.
  No build/pull/install; no competing heavy Mail work during root QA.
- No pro-qa-ai/** or Mail/Calendar access/changes and no other containers.
- Root commits exact paths and immediately pushes only calls-feature-codex
  to origin https://github.com/Valerijkk/hub.git; verify remote SHA.
- Explicit Terra/xhigh writer, Terra/high read-only reviewer; actual runtime
  model/provider UNKNOWN without telemetry, no silent fallback or children.
- One writer at a time for shared MessageItem/Actions/locales; root owns docs.

## Existing safeguards and precise gap

At c0900e62 MessageItem.handlePin captures actor/channel, checks nonempty
message ID/current live actor/prop and live can_pin/revoke, calls togglePin
and exact-invalidates pinsQueryKey(capturedChannelId,capturedUserId) after
a second live check. Preserve these safeguards. It lacks a catch, synchronous
lock, message/channel admission binding and committed A-B-A scope identity.
There is no row pending prop. PinnedMessages.handleUnpin has separate
list+pin and pin.channel_id safeguards; its similar command gap is deferred.
GET loading/error/empty UI is already implemented, not missing work.

Existing six phase151-message-row-pin-cache cases preserve exact key,
no-user denial and identity/DTO/live/revoke success fences. Panel-specific
phase151-pinned-identity-boundary, phase151-pinned-capabilities and
phase151-pinned-query-state suites remain regression, not new row evidence.

## Frozen interface and behavior

NEW frontend/src/components/messenger/use-message-pin-command.ts:

```ts
type PinCommandOptions = {
  channelId: string;
  messageId: string;
  messageChannelId: string | undefined;
  canPin: boolean;
};
export function useMessagePinCommand(options: PinCommandOptions): {
  pending: boolean;
  submit: () => Promise<void>;
};
```

1. Subscribe to auth userId, this channel live can_pin and revoke. Require
   nonempty actor/channel/message, messageChannelId===channelId, prop canPin,
   live can_pin===true, not revoked and a current mounted committed scope.
   Other authors and can_list_pins=false are allowed when can_pin permits.
2. useLayoutEffect creates a fresh scope object for each committed actor,
   channel,message,binding,prop/live can_pin or revoke transition. Cleanup
   invalidates that exact object and request ref. Reset local pending for
   each new scope. No ref writes during render. A-B-A cannot revive tokens.
3. Synchronously claim requestRef before first await. Duplicate returns
   without HTTP. Only the row Pin control disables; unrelated actions and
   other mounted rows remain usable. Capture IDs for the request.
4. Call togglePin(scope.channelId,scope.messageId), preserving API order.
   Current success awaits exact invalidation of
   pinsQueryKey(scope.channelId,scope.userId). Do not derive or insert data
   from returned {pinned:boolean}, invalidate another key or change label/state.
5. Currentness before effects and final settlement requires scope AND
   request object identity plus live auth/capability/revoke checks. Current
   pending spans the awaited invalidation. Keep TanStack default errors;
   do not set throwOnError. Old success/error/invalidation completion cannot
   notify, invalidate on behalf of an obsolete request or unlock a new one.
6. Current caught error calls toastApiError(error,t("messagePin.error")) once.
   EN fallback argument "Unable to update pinned message".
   RU fallback argument "Не удалось изменить закрепление сообщения".
   Canonical formatter precedence remains unchanged; this does not promise
   that fallback overrides a server or Error message. No new error policy.
7. Consume old rejection silently after actor/channel/message A-B-A,
   prop/live loss/restore, revoke or unmount. Current failure unlocks for a
   later explicit retry; no automatic request.

```ts
const request = { scope };
requestRef.current = request;
setPending(true);
try {
  await togglePin(scope.channelId, scope.messageId);
  if (isCurrent(request)) {
    await queryClient.invalidateQueries({
      queryKey: pinsQueryKey(scope.channelId, scope.userId),
      exact: true,
    });
  }
} catch (error) {
  if (isCurrent(request)) toastApiError(error, t("messagePin.error"));
} finally {
  if (isCurrent(request)) {
    requestRef.current = null;
    setPending(false);
  }
}
```

The scope effect separately invalidates obsolete refs; isCurrent must check
object identity and live stores, not just equal actor/channel string values.

Adapters:
- MessageItem instantiates hook with safeMessageId,message.channel_id and
  capabilities?.can_pin===true. Replace only handlePin with submit, wire
  pinPending to MessageActions; remove togglePin/pinsQueryKey imports and
  now-unused canPinNow plus its exclusively-used capabilitiesRef.
  Verify those helpers still have no other consumers after prior atoms.
- MessageActions adds optional pinPending?:boolean default false and passes
  it only to existing Pin ActionButton.disabled. No new visibility policy.
- Add messagePin.error in both messenger locale JSONs only.
- Import pinsQueryKey from ./pinned-messages without changing that module.

## Ownership and task

Production (at most250 added/deleted lines total):
- NEW frontend/src/components/messenger/use-message-pin-command.ts
- frontend/src/components/messenger/message-item.tsx: row-pin adapter only
- frontend/src/components/messenger/message-actions.tsx: pinPending only
- frontend/src/i18n/locales/en/messenger.json: messagePin.error only
- frontend/src/i18n/locales/ru/messenger.json: same

Oracle:
- NEW frontend/tests/messenger/message-pin-lifecycle.test.tsx only
- Keep all six existing row-cache cases and all panel tests unchanged.
- At most550 physical lines/26 concrete new cases; genuine real row/Actions,
  Zustand, QueryClient/QueryObserver; mock HTTP/formatter/unrelated children
  only. No future-hook import, owner mock or forcing internal state.

Root docs: this plan, completion queue,.planning/STATE.md,.planning/COVERAGE.md,
docs/TECHNICAL.md,docs/TESTING.md,docs/test_plan.md,docs/wiki/HOME.md.

### Task 1: Scoped row-pin mutation

- [x] Independent plan review C/I clear; no production writer until prerequisite
      reaction/bookmark ownership is released and canonical pin RED is triaged.
- [x] Write failing oracle against the actual published prerequisite HEAD.
      Freeze its hash; independently review every new case.
- [x] Root run the full canonical frontend-unit, preserve receipts and triage
      intended missing behavior vs type/fixture/runtime failure. No host QA.
- [x] One Terra/xhigh writer implements only hook/adapters/locales above.
- [x] Independent source+oracle review; fix C/I within bounds, no weakened tests.
- [x] Root full canonical GREEN, no missing/failed/pending/unhandled cases;
      inspect resource metadata immediately while the container is alive.
- [x] Docs sync, exact-path commit and immediate authorized push; verify SHA
      and continue the separate panel/unblocked queue.

Concrete oracle inventory (26 cases max):
- 1 positive deferred success: exact args/key, independent deep snapshots of
  own/other-user/other-channel pin and timeline caches; no optimistic writes;
  foreign author and can_list_pins=false do not block row can_pin.
- 2 current RU/EN rejection+explicit retry, exactly one formatter call with
  exact error/fallback, no success invalidation before retry.
- 1 same-tick duplicate plus independent row; pending affects only Pin.
- 7 denied admissions: no actor, empty message ID, mismatched message channel,
  prop denied, absent live capability, live denied, revoked. Observe actual
  row removal for revoke; do not invent a still-mounted control.
- 6 actor/channel/message A-B-A success/error, separate committed act steps.
  Old settlement must not unlock the newer pending command or invalidate/toast.
- 8 prop/live loss+restore/revoke/unmount times success/error.
- 1 real QueryObserver deferred refetch: pending spans invalidation; an old
  invalidation settlement cannot unlock the new scope's pending command.

```ts
const pinButton = screen.getByRole("button", { name: "Закрепить сообщение" });
act(() => { pinButton.click(); pinButton.click(); });
expect(togglePin).toHaveBeenCalledTimes(1);
expect(pinButton).toBeDisabled();
await act(async () => post.resolve({ pinned: true }));
expect(invalidate).toHaveBeenCalledWith({
  queryKey: pinsQueryKey("channel-a", "user-a"),
  exact: true,
});
```

Deep snapshots must be independent before seeding, not aliases. Promise
observers may await mocked HTTP settlement, but must not swallow command
rejections globally. Local window rejection spies are supplemental; root
Vitest whole-run unhandled-error result is mandatory.

## Acceptance limits

Production candidate206 added/deleted lines, including137-line local hook,
applied after root rejected one UNAPPLIED TypeScript mismatch: boolean
isCurrent does not narrow actorId:string|undefined for pinsQueryKey.
Writer added capturedUserId with explicit guard, no cast or key change.
Independent source/oracle review C/I/M0/0/0 READY_FOR_GREEN; oracle unchanged.
Full canonical200file/835test GREEN now running at04:14UTC; no changes to
source/tests/Compose while it runs. Preflight no active Mail build, only
idle BuildKit,13774768KiB freeRAM,disk0/queue0; no build/pull/install.

Pre-run hashes:
- PinHook FBAD71F82D606B5CB39D8AB90AEF7A138F92BF73D4619A9C7027ECD9403B0EFB
- MessageItem0B5A4DE72067E848002C1C820ECFDA1975E2D1CB1FD21A4C09DFB45519506E58
- MessageActionsC2F8D8B81F382F252E5696F3F17282FFB24BA77B6CF154BB8D7AA911651FAFE7
- EN4358AB10E4BE082A3103CE3513180E570D99F6503065315603F078EB0DF79025
- RU B47C2A4074CD225F0C6F94122A7600BB85748FC818CF4F7A910D9BB892905BD2
- OracleE021C6C9C7EC0A4C3895D5197FA7ACFA014BDCDB777EB52CDFF0D0F062E2C65D
- ComposeE8DA0EEC4F99FE3B437C885B3B19019F27D24958D477152176F7210CB4A56733

## Canonical baseline RED — 2026-09-06

Command: docker compose -f docker-compose.local.yml run --rm frontend-unit.
Published source baseline1877984e15462a4cc05ecb8cead2ca5810ac8ce2; no product,
Compose, old-oracle, assertion or timeout changes during run. New pin oracle
SHA256 remains E021C6C9C7EC0A4C3895D5197FA7ACFA014BDCDB777EB52CDFF0D0F062E2C65D.
TypeScript PASS;200 files199PASS/1FAIL;835 tests821PASS/14FAIL;
0pending/todo;6 whole-run unhandled errors; exit1;119.95s,start03:58:47UTC.
Only the new pin oracle fails:26cases12PASS/14FAIL. All old suites pass,
including the six row-pin cache cases and accepted reaction/bookmark atoms.

Failure map:
- Two current EN/RU rejections do not call the formatter.
- Same-tick clicks produce3 requests instead of2 for two independent rows.
- A message from another channel still dispatches one request.
- Six committed actor/channel/message ABA cases fail at missing pending
  before old settlement; their later stale-result assertions are NOT yet
  executed evidence.
- Three old successes invalidate after prop/live capability restore or unmount.
- Real QueryObserver invalidation case fails its missing pending witness.
Six unhandled strings map exactly to current EN/RU and late prop/live
loss+restore,revoke,unmount HTTP rejections. Local window spies alone did
not detect these whole-run handler errors. Existing warnings/mock stderr
remain; do not call stderr clean or treat waitFor assertion timeouts as an
infrastructure diagnosis without their actual source failure context.
Independent final RED triage C/I/M0/0/0 READY_FOR_PRODUCTION. All14 failures
and6 unhandled origins match the frozen source gaps; no type/fixture or
old-suite blocker. Root dispatched one Terra/xhigh production writer for
the exact local hook/adapters/locales, leaving this oracle unchanged.

Preserved eight receipts, new path without overwrite:
testframework/reports/frontend-unit-pin-red-20260906/
{result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log}.

Outer worktree-master-frontend-unit-run-8eff96ce39fc; engine
51e861947c87c878995195af3d8990808fb77122730e9a7b598980d4576c72c5;
started2026-09-06T03:58:28.940206244Z; unchangedimage
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Actual inspected Memory=MemorySwap2147483648,NanoCpus6000000000,readonly;
MAX2/MIN1 threads/forks. Startup top precedes workers; subsequent top
observed main Vitest plus two named workers and esbuild. Samples200.5/
492.8MiB,CPU134.45/257.38percent,IO0B/0B; no peak/p95 measurement.
Engine completion/OOM events NOT_CAPTURED; no no-OOM claim from absent data.
Preflight activeMailBuilds0,only idleMailBuildKit,15269976KiB freeRAM,disk0/
queue0. No build/pull/install, other-container changes or concurrent root QA.

This is a frontend genuine-unit boundary, not API/storage/browser or full
Messenger acceptance. Existing successful receipts are not this future atom's
GREEN. Max1 transient retry,3 fixes/rootcause,5 rereview rounds. Stop on
unexpected shared-file conflict, missing oracle, destructive or authority
need; record blocker and continue independent approved work. Do not re-run
blocked Go at2GiB or activate/pull Jitsi without its separate authorization.

## Canonical final GREEN — 2026-09-06

Root command: docker compose -f docker-compose.local.yml run --rm frontend-unit.
TypeScript PASS; 200 files / 835 tests PASS, all 26 pin cases; zero failed,
pending/todo; target, TypeScript, Vitest and Compose exit 0. Start 04:14:37 UTC,
duration 117.97s. JSON's 417 suites count includes describe groups, not files.
There is no numUnhandledErrors property; no whole-run Unhandled Errors /
Unhandled Rejection / Uncaught Exception headings in raw logs, and exit 0
confirms the runner gate. Existing React act warnings and mock stderr remain.

All seven pre-run source/oracle/Compose hashes above match after the run.
Oracle unchanged from intended RED; all old row/panel suites remain intact.
Independent final review C/I/M 0/0/0, UNIT_ACCEPTED for this frontend atom.

Eight receipts preserved in a new directory without overwrite:
testframework/reports/frontend-unit-pin-green-20260906/
{result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log}.

Outer worktree-master-frontend-unit-run-c6d794bcdb32; engine
d67ab9f11460d04590bb786e8b0ffc5addcd5772bbc3a36873defaf351e5a312;
started 2026-09-06T04:14:20.528837123Z; unchanged image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Inspected Memory=MemorySwap 2147483648, NanoCpus 6000000000, readonlyRootfs;
MAX2/MIN1 threads/forks. Initial top contained tsc only; subsequent top showed
main Vitest plus two named workers and esbuild. Samples 865.5/299.4 MiB,
CPU 117.55/236.49 percent, IO 0B/0B; no peak or p95 measurement.
Engine completion/OOM state NOT_CAPTURED, not a measured no-OOM claim.
Preflight active Mail builds 0, only idle Mail BuildKit; 13774768 KiB free RAM,
disk 0/queue 0. No build/pull/install, timeout changes or other-container work.
