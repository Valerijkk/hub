# Composer retry single-flight and retained recovery controls

Status: UNIT_ACCEPTED in the composed canonical Docker gate: TypeScript
and 197 files / 758 tests PASS. Plan, oracle, source and receipt reviews
C/I/M 0/0/0. Two new mock-boundary cases and four REAL-Toaster cases pass.
The intended RED and failed first post-source run are retained below.
Baseline calls-feature-codex d10f1dda. Parallel edit atom owns different files;
preserve it. One failed logical SEND recovery behavior, not a send redesign.

## Confirmed source and root decision

attemptPendingSend checks the pending-map object and existing auth/revoke
gates but lacks a per-pending transport lock. Two Retry actions can dispatch
two HTTP requests with the same immutable client_nonce before either settles.
The existing nonce and server idempotency contract must remain unchanged.

Root additionally checked the locked Sonner2.0.7 source (package-lock.json
node_modules/sonner) instead of assuming mocked callbacks equal actual UI.
Its action handler closes a toast unless the callback prevents default.
Its built-in cancel handler returns without calling the callback when the
toast is non-dismissible. Current Hub toast is dismissible:false, so existing
Retry/Cancel mock assertions miss two actual interaction failures.

Primary source: https://raw.githubusercontent.com/emilkowalski/sonner/v2.0.7/src/index.tsx
Relevant upstream source lines446-451 and465-470. This source is evidence
about the library boundary, not authority to change Hub requirements.

Root chooses to retain indefinite, non-dismissible failed-send notifications.
Retry must prevent automatic action dismissal; successful send or explicit
cancel owns removal. Keep explicit Cancel working as a native button element
provided in Sonner's supported React-element cancel slot, bypassing only its
built-in dismissible check. Do not make failed sends swipe-dismissible, add
a new modal, change copy or retain hidden orphan pending records.

## Exact ownership

- frontend/src/components/messenger/message-composer.tsx:
  PendingMessageSend transport flag, exact attempt admission/finally,
  Retry action keep-open behavior and native cancel element only;
- frontend/tests/messenger/message-composer-retry-nonce.test.tsx:
  focused single-flight regressions and stronger runtime toast-boundary
  helper adaptation for the real native cancel element; preserve all prior
  semantic assertions for snapshots, nonce, drafts, identity and invalidation;
- frontend/tests/messenger/message-composer-retry-toast.test.tsx:
  new genuine unit with REAL MessageComposer and REAL Sonner Toaster,
  mocked HTTP/unrelated expensive children only;
- root owns plan and same-session domain/status/test docs.

No edit/delete/search/thread-source overlap, shared toast component changes,
backend/API/store/query schema/dependency/lockfile/config/launcher changes,
Mail/Calendar/pro-qa-ai. No generic retry framework or broad composer rewrite.
Target <=40 changed production lines. Return BLOCKED for decomposition if
this cannot remain a single bounded recovery-control behavior.

## Frozen behavior

1. Each pending object/nonce has at most one active transport attempt,
   initial or retry. Establish its lock synchronously after current
   map/authority admission and before await; repeated action is a no-op.
   Release only that record in finally after settlement. Discarded old record
   cleanup cannot unlock or retarget a different logical send.
2. Keep existing immutable payload, nonce, principal/channel/thread binding,
   explicit retry/cancel, fresh auth/revoke guards, draft/attachment clearing
   and exact successful invalidation semantics. No global lock prevents a
   distinct allowed new logical send while a failed-send retry is pending.
3. Retry action prevents Sonner's automatic dismissal, including a duplicate
   click while locked. Real React event is supported; existing direct unit
   callback invocation without an event must not throw. The toast remains
   available through a failed retry. Success still invokes existing discard.
4. Cancel remains the exact Russian label, native type=button, in the
   supported React-element cancel slot, with data-button/data-cancel styling.
   It invokes only discardPendingSend for its owned record. Keep
   dismissible:false and duration:Infinity; no new close/swipe cancellation.
5. Cancel during an in-flight retry hides the notification and discards the
   local pending record. Already-sent HTTP is not cancelled or undone; late
   success/failure has no stale invalidation or new failure toast.
6. No automatic retries, extra toasts per attempt, changed nonce, replacement
   of newer draft, optimistic server acknowledgement or claim that client
   locking substitutes for backend idempotency.

## Test-first requirements

Extend the existing real-composer/mock-toast transport oracle with:
- initial failure then two same-tick Retry invocations while deferred retry
  is pending => exactly2 sends total, identical payload/nonce;
- rejected retry releases its own guard so a later explicit Retry sends once;
- simultaneous distinct logical send stays independent; old cancelled
  attempt settlement cannot block/unlock/retarget its record;
- action event.preventDefault is called while request starts and on duplicate
  retained click; no-event legacy callback remains safe.
Existing helper currently asserts an object cancel {label,onClick}. Adapt it
to assert the emitted native button type, accessible label and callback,
then expose a normalized callback to unchanged semantic tests. This is a
documented correction of a mocked library-boundary assumption, not removal
of cancellation or dismissible assertions. No source-text/AST tests.

New real-Toaster genuine runtime cases must actually CLICK rendered controls:
- Retry remains available after a deferred retry rejection and a later
  explicit retry can recover; multiple pending clicks do not duplicate sends;
- Cancel on a non-dismissible failure really removes it without retry;
- Cancel during retry consumes its late rejection/success without resurrecting
  the toast or invalidating a newer scope.
Mock target composer/Toaster is forbidden in those cases. Use current
package/RTL, no user-event install, fake source transport, external service,
browser process or timer/timeout inflation. Wait for actual public UI
witnesses and explicit deferred settlement; account for existing dismissal
animation without fixed sleeps. Preserve all other Messenger tests.

Root owns canonical combined RED/GREEN:
docker compose -f docker-compose.local.yml run --rm frontend-unit
Existing image,2GiB/no-swap,serial workload checks; no host QA, rebuild/pull,
dependency install or concurrent heavy Mail work. Independent Terra/high
review, explicit Terra/xhigh test-first writer, retry1/fix2/rereview2.
Oracle first; no production implementation before intended valid RED.
Accepted exact commit immediately pushes only calls-feature-codex to
origin https://github.com/Valerijkk/hub.git with remote SHA verification.
Native browser/media/API/storage/module acceptance remains NOT_RUN.

## Oracle preflight

Real-Toaster cleanup now dismisses every toast.getToasts() entry by its ID
inside act before cleanup. Sonner2.0.7 no-ID dismissal does not itself mark
all global records dismissed. This is a pre-RED fixture-isolation correction;
replay into a later Toaster or a resulting runtime failure was not observed.
Primary source: https://raw.githubusercontent.com/emilkowalski/sonner/v2.0.7/src/state.ts
Independent final narrow rereview C/I/M0/0/0, ready for intended RED.
The existing mock helper now requires the contracted native cancel element;
old test failures at that boundary will be intentional, not evidence that
the entire previously accepted728-case baseline stayed green in RED.

## First composed RED (2026-09-06)

Canonical docker compose -f docker-compose.local.yml run --rm frontend-unit,
published baseline d10f1dda with only the frozen test additions. TypeScript
PASS;197files/758tests,722PASS36FAIL,81.19s,Compose exit1. Targeted evidence:
18 edit assertions and11 composer assertions expose missing frozen behavior.
Composer's7 mock-boundary failures include existing cases strengthened to
require the actual native cancel slot; the4 real-Toaster failures exercise
rendered buttons, including duplicate request and non-working cancellation.

The run also has SIX15000ms timeouts (edit positive, delete allowed, pinned
capability, pinned identity, thread wiring, TMS member) and ONE extra
pinned-identity API-count failure. Their cause is UNKNOWN; a timeout cascade
is only a hypothesis. These7 failures are not accepted behavioral RED and
must not be hidden or attributed to production changes (none were applied).
No baseline-wide PASS is claimed. No tests, timeouts or launcher were relaxed.

Engine2eaa5a2a0f12b1751b8fd1be740cf7c362bc53025cdfdf1d010f802436476810,
outer worktree-master-frontend-unit-run-a173b2592157,
start2026-09-06T01:00:29.178276651Z, observed image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Actual2GiB/no-swap,6CPU,readonly; create/start/die/destroy captured,exit1
after98s,no OOM event in captured lifecycle; container absent afterward.
No adjacent active build/service at preflight/finish; previous Mail build
ended with its own Error state, not a successful-build claim. Root did not
stop/change it. Resource preflight freeRAM13035156KiB,disk0%/queue0.

Eight exact current receipts preserved:
testframework/reports/frontend-unit-edit-composer-red-20260906/.
Pre/post source+oracle SHA256 values match:
- edit oracle A8D8B50C0F6CF3C382DD498C477DF1761A6696A315D612E8E3E9BA2BDEB57D1B;
- composer mock oracle72BCBE4B49DCC6A64EB737A15D459610FEDDB413288F9334283111CB8C1E2006;
- real-Toaster oracle98A82978BDD4A820DEEE0D65D3C37048AF2F8DCDB4709389286B0E9711478488;
- MessageItem9F9B0EC52F9DA93F404E2EF217BBE41940A7C4F7D5F1F872CB48644A9789735A;
- MessageActionsF8E882F64B0CC22B2F62871215E44FD7E82ABF04B5C5A640E02D997F982F5FD3;
- MessageComposerC5C019B2687CC0405C2B6A6281FD5E70872CD33A8B82B00AFBEDD5FDA05386C4.
Independent triage requested before source dispatch. Full GREEN, source
review and same-session docs sync remain required for acceptance/publication.

## Source candidate checkpoint

Independent mixed-RED triage C0/I2/M0 authorized exact writers on29 intended
failures; seven UNKNOWN failures remain full-acceptance blockers. Read-only
follow-up did not establish a common fixture/load cause. No hidden timeout
increase, assertion relaxation or unrelated source repair is authorized.
Root applied the27-changed-line composer candidate after payload preflight;
independent source review C/I/M0/0/0. Hash:
19E70B07E65482FE63E512FBB7D06DBCE8A307FFDC3D1C4052B78B1065CC2762.
Each record owns its inFlight flag; actual Retry prevents default even on
duplicates, and native Cancel discards only its record. Source review is
not executable acceptance. Frozen oracle hashes remain unchanged.

## First post-source attempt (NOT_ACCEPTED)

The unchanged canonical command ran2026-09-06T01:22:11.265763406Z in engine
79f98d4b8c15ce5c2c621bf930cffb10e8f2b391339259e7c91e56b59813a0b3,
outer worktree-master-frontend-unit-run-de7744cef331. Actual image remains
286783...218,2GiB/no-swap,6CPU,readonly. TypeScript PASS; Vitest reports
197files but only752tests versus expected758:740PASS12FAIL,3unhandled
errors,296.50s. Compose/engineexit1after314s,create/start/die/destroy
captured,no OOM event. All eight source/oracle hashes match pre/post.

This is NOT_GREEN. All11 mock-composer cases passed,23/24 edit passed and
3/4 real-Toaster passed; edit's first positive timed out and real Retry's
last disappearance witness failed. Other timeouts spanned unrelated suites.
The3 whole-run errors were two worker onTaskUpdate RPC timeouts and a
post-teardown React/i18n callback ReferenceError(window is not defined).
The missing6 tests and worker errors preclude acceptance of the run.

At01:26 owned container used535.30%CPU,1.815/2GiB,137PIDs,BlockIO0/0.
Disk sample0%/queue0,freeRAM14001180KiB. Adjacent Mail qa-runtime/backend
builds started only then, roughly4minutes into an already slow run; their
later overlap cannot explain the initial slowdown by itself. No other
container was stopped or changed. Source/expectations/timeouts/runner stay
unchanged; no blanket memory/load/cascade diagnosis is proven.

Independent triage C0/I3/M0 advises no feature-source change from these
symptoms: real-toast removal/edit first-witness causes UNKNOWN; async i18n
after environment teardown is a TEST/runtime-lifecycle failure. Final
unhandled-free full-count GREEN remains required. Preserve exact8receipts:
testframework/reports/frontend-unit-edit-composer-attempt1-20260906/.
One unchanged retry may capture actual worker count and resource state after
adjacent heavy jobs finish; do not loop or relax the canonical gate.

## Bounded composed GREEN (2026-09-06)

UNIT_ACCEPTED, independent Terra/high receipt review C/I/M0/0/0. Exact
canonical docker compose -f docker-compose.local.yml run --rm frontend-unit:
TypeScriptPASS;197testfiles/758testsPASS,114.95s,failed/pending/todo0.
JSON411 counts suite groups, NOT files. All24edit,11mock-composer and4real-
Toaster cases pass. Previously unknown test outcomes pass in this matrix;
the6missing cases are present again. No unhandled/worker/window/timeout
failure headlines remain; existing non-failing React/description warnings
are not claimed eliminated. Earlier failed receipts remain preserved.

Engine d154c7aa8fae7b890663606482381047a9591450035b413373cf512743c90faf,
outer worktree-master-frontend-unit-run-445e24bfa4fe,
start2026-09-06T01:39:37.938536298Z, same image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Inspected2GiB/no-swap,6CPU,readonly and all four MAX2/MIN1 worker env values.
Docker top while Vitest was active showed main node(vitest), two explicit
node(vitest1)/node(vitest2) children and esbuild. Create/start/top/die/destroy
captured,engine/Composeexit0after131s,no OOM event; container absent afterward.
A later stats attempt found the already-removed container, so bounded-run
peak RSS/CPU is NOT_CAPTURED. Do not present the prior1.815GiB sample as this
run's usage. Preflight freeRAM11271184KiB,disk0%/queue0,CPU9%; no adjacent
active build/service then, and root never modified Mail resources.

Eight exact receipts:
testframework/reports/frontend-unit-edit-composer-bounded-green-20260906/.
Pre/post compose, five feature sources/locales and three oracle hashes
match. Compose E8DA0EEC4F99FE3B437C885B3B19019F27D24958D477152176F7210CB4A56733.
All feature/oracle hashes remain the previously recorded candidate hashes.
No build,pull,install,hostQA,new exclusions,timeout/assertion changes or RAM
increase. The only execution delta is the declared service-local worker
budget. Workload was not a controlled experiment, so this does not prove
universal causation or permanent elimination of every earlier timeout.

This evidence covers the COMPOSED working tree (resource + edit + composer),
not each forthcoming atomic publication commit in isolation. Browser/API/
storage/media, whole modules, Go4GiB exception and Jitsi activation remain
outside this gate and NOT_RUN/UNBOUND.
