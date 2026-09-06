# Message forward command lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Follow the checkboxes. Root owns Docker QA, docs and Git publication.

**Goal:** One current forward command per dialog session; handled current error and explicit retry; obsolete completions cannot close or unlock a newer session.

**Architecture:** One local hook owns committed scope, synchronous request identity and temporary store subscriptions. ForwardDialog preserves target filtering and existing UI, delegating only the command. No generic framework or backend rights change.

**Tech Stack:** Existing React, Zustand, TypeScript, Vitest/RTL, i18next and toastApiError. No dependencies.

**Baseline:** calls-feature-codex ac918830e2aa0d7082297d9913be0fd750013fb5, linked worktree-master, exact origin SHA verified, clean worktree. Prior canonical frontend-unit TypeScript + 200 files / 835 tests PASS is baseline only, not this atom's RED/GREEN.

**Spec:** docs/TECHNICAL.md forwarding endpoint; source-confirmed ForwardDialog gap in the communications completion queue; existing source CanForward / target CanPost policies remain authoritative. This lifecycle supplement is a bounded reliability fix under founder autonomous authority.

Current status: UNIT_ACCEPTED; TypeScript + 201 files / 871 tests PASS,
36 forward cases, 123.07s. Independent source/final receipt C/I 0/0;
minor existing warning disclosure retained. Published 74c5c492a41998213fcf7cf1556fdf6333c7ee42;
immediate push and exact remote SHA equality verified.
API/browser/full module acceptance remains NOT_RUN. Earlier checkpoints
below are chronological history, superseded by the final GREEN section.

Historical resume: 2026-09-06, c1990858b572c1cf470c67410d6960db229583cf.
Only workflow/docs commits followed ac918830; source/oracle hashes unchanged.
Canonical RED is complete (final section below); source stage is authorized.
Requested source runtime: gpt-6-astra/high under current model-routing; no fallback.

Historical pre-oracle status: ROOT_CONTRACT_FROZEN; independent plan review C/I/M 0/0/1,
READY_FOR_ORACLE. Minor budget wording clarified: sum(added + deleted) <= 250
across production paths only. One oracle writer may proceed; production
remains unchanged until independent oracle review and root RED triage.

Oracle now applied: 483 physical lines, 36 new cases; existing capability
fixture +21 explicit actor-only lines, all five cases/assertions preserved.
Root verified full readback equals corrected author payload. SHA256:
- oracle 01207AD0B3B34707D83B7D5771015D451618C94EB1D599C8070829C0AF30E077
- fixture 657C4EEAA8907D3514153369FCC1ABE43BE6C79CC18D18C2E50191EEFAE3180A
Production diff is empty, Compose unchanged. Independent oracle review
C/I/M 0/0/0 READY_FOR_RED after fix1. Canonical RED started 05:04 UTC;
outcome pending, not expected-failure evidence yet. SDD brief/report/full review package
are in .superpowers/sdd/2026-09-06-message-forward-lifecycle/.
Author default helper ACL/actuator issue prevented writes; root required
actual patch delivery then applied it with official apply_patch, no ACL
changes. Unsupported primitive/tuple titles and weak retry completion counts
were corrected by the author BEFORE first application, no assertion removal.

Review I1 was addressed in the same callback case: real flushSync/public
same-scope re-entry dispatches a second deferred request, verifies pending
after the first settles and exact two false callbacks, then settles second.
No 37th case/internal hook mock. Initial 467 physical vs451 nonempty line
count discrepancy was a Measure-Object -Line distinction, not receipt drift.

RED runtime snapshot: outer48fe94fa4d33, engine
ec4e1e0189842be806df2befad2069a1cb5c3f35947a304c68fb3438cab726fc,
started2026-09-06T05:04:07.209611319Z, unchanged image286783...;
actual Memory=MemorySwap2147483648,NanoCpus6000000000,readonly,MAX2/MIN1.
Initial top tsc only,857.5MiB sample/137.41percentCPU/IO0B. Not a peak.
Preflight activeMailBuildCount0,latestCompleted04:57:39.099049135Z;
9464916KiB freeRAM,disk0/queue0. Six healthy low-CPU Mail services and idle
BuildKit left untouched. No source/test/Compose edits during this run.

## Constraints and ownership

- One behavior: ForwardDialog forward command only. MessageItem, MessageActions, row pin, panel unpin, reactions and bookmarks remain unchanged.
- No API/server role, capability vocabulary, payload, cache/query/WS changes, optimistic insertion, auto-retry, abort/rollback or cross-dialog/server idempotency guarantee.
- Existing Russian labels, preview, target ordering, filtering, effectiveOpen and Dialog callback visibility checks remain.
- Production at most 250 added/deleted lines total:
  - NEW frontend/src/components/messenger/use-message-forward-command.ts
  - frontend/src/components/messenger/forward-dialog.tsx
  - frontend/src/i18n/locales/en/messenger.json: messageForward.error only
  - frontend/src/i18n/locales/ru/messenger.json: same
- Oracle:
  - NEW frontend/tests/messenger/message-forward-lifecycle.test.tsx, at most 600 physical lines / 36 new genuine cases.
  - frontend/tests/messenger/phase151-message-action-capabilities.test.tsx: only explicit useAuthStore actor setup/reset required by the new admission guard; preserve all existing assertions and counts literally.
- Root owns this plan, completion queue, .planning/STATE.md, .planning/COVERAGE.md, docs/TECHNICAL.md, docs/TESTING.md, docs/test_plan.md, docs/wiki/HOME.md.
- All agents are not alone: one exact writer per file; no unrelated reverts, protected pro-qa-ai or Mail/Calendar reads/edits, no child agents.
- Source dispatch explicitly gpt-6-astra/high; independent review gpt-5.6-terra/high. Earlier oracle used the then-current Terra/xhigh route. Actual provider/runtime UNKNOWN without telemetry; no fallback or inherited routing.
- Host reads/edits/Git only. Root alone runs exactly:
  docker compose -f docker-compose.local.yml run --rm frontend-unit
- Serial full gate, existing image, 2 GiB/no-swap, 6 CPU, two workers, unchanged timeouts; no build/pull/install. Check adjacent Mail build pressure first; do not stop/change other containers.
- Exact-path commit then immediate push ONLY calls-feature-codex to origin https://github.com/Valerijkk/hub.git; verify SHA. No master, merge, tags, force or other ref.
- Go 2 GiB compile retry and Jitsi pull/start remain blocked by their recorded authority/resource gates.

## Existing behavior and gap

ForwardDialog currently uses state isForwarding plus try/finally. Two same-tick handlers can both observe false, failures reject without feedback, and old success/finally may close/unlock a later session. Preserve canForwardNow's live source existence/can_forward, target existence/can_post and both revoke checks, including its post-request check.

API stays forwardMessage(sourceChannelId, messageId, targetChannelId), POST /messenger/channels/:channelId/messages/:messageId/forward with target_channel_id. No separate channel role inference.

Existing phase151-message-action-capabilities has two forward cases: target filtering and source revoke in flight. Its actor fixture must become explicit before RED; this is fixture admission, not assertion relaxation.

## Frozen interface

~~~ts
type ForwardCommandOptions = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  channelId: string;
  messageId: string;
  messageChannelId: string | undefined;
  canForward: boolean;
};
export function useMessageForwardCommand(options: ForwardCommandOptions): {
  pending: boolean;
  submit: (targetChannelId: string) => Promise<boolean>;
};
~~~

The boolean is internal command acceptance: true only for current success/close; false for denied, stale or caught failure. Promise must not leak the mocked HTTP rejection to its void UI caller.

## Frozen lifecycle

1. A useLayoutEffect creates a fresh scope object for every committed actor, open, source channel, message, message-channel binding or prop canForward change. No render-time ref writes. Cleanup invalidates scope AND its request and releases that request's subscriptions. A-B-A never revives identity. Set pending false for the new scope.
2. Keep latest onOpenChange in a separate committed callback ref. Callback identity and unrelated channels array changes must NOT create a new session or release its lock. Capture no stale callback semantics for successful current closure.
3. Admission requires mounted current scope, open, prop canForward, nonempty actor/source/message/target IDs, messageChannelId===channelId, exact live actor, existing non-revoked source can_forward and existing non-revoked target can_post. UI already excludes source from target list; do not add a new backend permission rule.
4. Before first await, synchronously claim one request object containing scope, immutable target ID and invalidated=false. All target buttons share pending; same-tick clicks on two different targets still produce one POST.
5. Install temporary useMessengerStore.subscribe and useAuthStore.subscribe listeners on the accepted request before dispatch. Each re-evaluates that request's actor/source/captured-target authority; any observed loss marks invalidated=true permanently. No setState or ref clearing inside these listeners: transient true-false-true within one React act must remain invalidated, but unrelated channel-list changes must not invalidate a still-authorized request.
6. A live authority predicate checks stores both at admission and settlement. Scope/request identity AND !invalidated are required for current effects. A separate ownership predicate (scope/request identity only) controls safe pending cleanup even when authority was lost. Do not use authority-currentness as the sole finally guard and strand pending after target loss.
7. Current success: release exactly this request (unsubscribe, ref null, pending false) BEFORE invoking the latest onOpenChange(false); return true. Old finally must not clear a new request started by a callback/reopen.
8. Current catch: toastApiError(error,t("messageForward.error")) exactly once, then return false. EN fallback argument: "Unable to forward message"; RU: "Не удалось переслать сообщение". Preserve canonical formatter precedence.
9. Every settlement and scope cleanup releases subscriptions idempotently. A stale or invalidated success/error is silent and returns false. If an invalidated request still owns this same open scope, settlement releases only its pending so explicit retry is possible. A detached old request cannot unlock a newer scope.
10. Already dispatched HTTP may still affect the server. Do not claim cancellation, rollback or durable idempotency.

## Implementation map

- Hook owns auth selector, scope/request/callback refs and pending, store predicates/subscriptions and submit.
- ForwardDialog keeps useMessengerStore UI selectors and existing filtering/effectiveOpen; remove obsolete useState/useCallback/Channel/forwardMessage imports and handler. Instantiate hook with message.id/message.channel_id and existing props. Wire pending to existing buttons and onClick={() => void submit(channel.id)}.
- Locale keys only as above.
- Production code starts only after root full RED and independent failure triage.

## Oracle inventory — 36 new cases

Real ForwardDialog plus Zustand and existing i18n; mock HTTP/formatter and unrelated presentation only. Do not import/mock the future hook, force internals, read source files or globally swallow unhandled errors. Exact assertions for requests, formatter, callback count and pending.

- 1 success: exact source/message/target args, one close, pending across HTTP; unrelated channels list update does not release lock.
- 2 EN/RU current failure and explicit retry: exact error/fallback, once formatter, exact second request, successful close once.
- 1 same-tick two-target click: one POST, all target buttons disabled.
- 5 invalid ID/binding admissions: absent actor, empty source, empty message, mismatched message channel, empty target ID.
- 6 fresh denial cases: source absent/can_forward lost/revoked; target absent/can_post lost/revoked. Use captured live UI handler in a same-act store transition where applicable; do not assert hidden buttons are clickable.
- 8 loss/restore cases times success/error: prop canForward, source can_forward, target can_post, target revoke. Store loss/restore must occur within one act for the latter three, proving subscription invalidation even without an intermediate React commit. Restore target controls, observe silent settlement then explicit retry admitted.
- 8 committed ABA cases times success/error: actor, source channel with same message ID and matching binding, message ID, dialog open false/true. Separate committed act/rerender transitions; issue a newer request, settle old first and prove newer pending remains until its own success. No accidental second dimension change except necessary source binding.
- 2 unmount success/error: no toast/close or whole-run unhandled rejection.
- 1 current callback replacement: rerender with new onOpenChange only while pending; no duplicate dispatch, success calls latest callback once, old never.
- 2 batched actor A-B-A success/error: no intermediate React commit; subscription invalidation, silent old settlement and explicit retry.

Count: 1+2+1+5+6+8+8+2+1+2=36. All five old cases remain, including the two direct forward cases. New total expected 201 files / 871 tests if no unrelated changes.

Use independent deferred promises, deterministic real act boundaries, no sleep/timeout bumps. Local window rejection spies are supplemental; canonical Vitest whole-run unhandled gate is required. Any RED precondition failure is reported as such, not proof that later assertions executed.

Concrete same-tick oracle anchor (real rendered target buttons):

~~~ts
act(() => { targetButtons[0].click(); targetButtons[1].click(); });
expect(forwardMessage).toHaveBeenCalledTimes(1);
expect(forwardMessage).toHaveBeenCalledWith("source", "message-1", "target-a");
expect(targetButtons[0]).toBeDisabled();
expect(targetButtons[1]).toBeDisabled();
~~~

Reference command sequencing (predicates/ref implementation follows the
frozen lifecycle above, not a new public API):

~~~ts
try {
  await forwardMessage(scope.channelId, scope.messageId, request.targetId);
  if (!isCurrent(request)) return false;
  releaseOwned(request);
  callbackRef.current(false);
  return true;
} catch (error) {
  if (isCurrent(request)) toastApiError(error, t("messageForward.error"));
  return false;
} finally {
  releaseOwned(request);
}
~~~

## Task 1: Scoped forward command

- [x] Independent plan review C/I clear.
- [x] One Terra/xhigh oracle writer applies only standalone tests and permitted actor fixture.
- [x] Root reads full oracle, measures cases/lines/hashes; independent oracle review C/I clear.
- [x] Root canonical full RED, preserve eight receipts; independent source/fixture/runtime triage.
- [x] One explicitly routed source writer applies only reviewed bounded patch.
- [x] Root/independent source review C/I clear; oracle hash unchanged.
- [x] Root full canonical GREEN, exact receipts and pre/post hashes, actual runtime samples.
- [x] Final independent receipt review and root docs sync; C/I0/0, warnings disclosed.
- [x] Exact commit/immediate push/remote SHA; continue unblocked queue.

## Bounds / handoff

Each dispatch max 20 minutes, max depth/fanout 0, max 3 fixes per root cause / 5 review rounds / 1 transient retry. If one behavior cannot fit 250 changed production lines, return BLOCKED for root decomposition before writing. No automatic extra abstraction or new authority.

Stop only that atom on unexpected shared-file conflict, missing oracle, runtime/model unavailable, forbidden execution or scope/authority need. Record exact evidence and continue independent approved work. Browser/API/server persistence/full Messenger acceptance remain NOT_RUN until separately proven.

## Canonical baseline RED — 2026-09-06

Root canonical frontend-unit: TypeScript PASS; 201 files200PASS/1FAIL;
871tests848PASS/23FAIL,0pending/todo;8 whole-run unhandled rejections;exit1.
Start05:04:24UTC,duration122.11s. New oracle36cases13PASS/23FAIL; all old
200 suites pass. Source/Compose/oracle/fixture hashes unchanged through run.

Failure map:2missingformatter,1same-tick duplicate(2POSTs not1),5ID/binding
admissions,4stale loss/restore successes,8ABA enabled preconditions,1unmount
success,1callback re-entry missing second request,1batchedactor success.
ABA failures occur BEFORE newer request dispatch; later stale/finally checks
are not yet executed evidence. The callback test similarly fails before
its second-request pending witness. Eight raw unhandled strings exactly
map to EN/RU current, late prop/source/target/revoke, unmount and batchedactor.
These source-handler rejections are NOT fixture or infrastructure failures.
Independent final RED triage C/I/M0/0/1 opens source writer only, not release.

Preserved eight receipts, no overwrite:
testframework/reports/frontend-unit-forward-red-20260906/
{result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log}.

Runtime limits/image/start/preflight are recorded above. Later top observed
main Vitest plus two named workers; second sample508.4MiB,CPU234.39percent,
IO0B. Samples857.5/508.4MiB are not peak/p95. Engine completion/OOM state
NOT_CAPTURED; no OOM conclusion. Existing warnings/mock stderr remain.
No build/pull/install, source/test/Compose edits during run or Mail changes.

## Canonical source GREEN — 2026-09-06

At base c1990858 plus the bounded source patch, root ran exactly
`docker compose -f docker-compose.local.yml run --rm frontend-unit`.
TypeScript PASS; 201 files / 871 tests PASS, including all 36 forward cases;
exit 0, 123.07s, start 10:43:07 UTC. Failed/pending/todo and whole-run
unhandled errors: 0. JSON's 419 nested suites are not the file count.
Existing mock/act stderr remains; matching rejection words in test titles
are not runner error headings. Build/lint/API/browser/storage NOT_RUN.

Independent source review: C/I 0/0, READY_FOR_DOCKER_GREEN; production diff
183 added + 56 deleted = 239/250. Source requested Astra/high, reviewer
Terra/high; actual runtimes UNKNOWN. The frozen oracle/fixture and all four
source hashes match the source report before/after GREEN; Compose unchanged.
Final independent receipt review: READY_FOR_UNIT_PUBLICATION, C/I0/0;
minor pre-existing DialogContent description/act/mock warnings disclosed.
Publication is still pending. Root checked the disposable container is
absent after exit0; no adjacent containers were stopped or changed.

Preflight: no active Mail build; free RAM 11,911,456 KiB; disk sample 16%,
queue 0; existing image, no download/build/install. Outer container
`worktree-master-frontend-unit-run-5263838871da`; engine
`a4d5091d47ea9b89bc57a4c5a7ec5cf84766149f0cf5933c09ff84d3f7e28e6e`,
started 2026-09-06T10:42:49.516364242Z. Image
`sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218`.
Actual memory=swap limit 2,147,483,648 bytes; NanoCpus 6,000,000,000;
read-only root; min1/max2 for both worker pools; two named Vitest workers
observed. Samples 536.3/410.2 MiB, CPU159.11/235.79%, block IO0B;
not peak/p95. Engine completion/OOM state NOT_CAPTURED.

Eight receipts retained without overwrite in
`testframework/reports/frontend-unit-forward-green-20260906/`:
result.json, typescript.result.json, typescript.stdout.log,
typescript.stderr.log, vitest.result.json, vitest.report.json,
vitest.stdout.log, vitest.stderr.log. RED receipts remain alongside.
