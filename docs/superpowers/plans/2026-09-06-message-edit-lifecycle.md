# Message edit save lifecycle

Status: UNIT_ACCEPTED in the composed canonical Docker gate: TypeScript
and 197 files / 758 tests PASS, including all 24 edit lifecycle cases.
Plan, oracle, source and receipt reviews C/I/M 0/0/0. Intended RED and
the failed first post-source attempt remain preserved below.
Baseline calls-feature-codex d10f1ddaac4e47b622d2db0eae1cfe662f0c5c8b
is clean and published; composed frontend TypeScript/728 tests accepted.
This is one frontend save interaction, not a new edit permission or API.

## Source finding and decision

MessageItem has an inline MessageEditMode, not a separate editor module.
The child sets saving, but parent handleSaveEdit optimistically updates the
origin timeline and immediately closes editing before awaiting PUT. The
toolbar can then open a second edit while that PUT is pending. Rejection
only invalidates the exact origin key, with no visible error. Admission
checks ownership/can_edit but lacks message.channel_id===channelId.

Root retains current optimistic preview and immediate editor close, payload,
trim/no-op behavior and exact origin timeline invalidation on current error.
Add one item-scoped pending/request-lifecycle guard and canonical failure
feedback. The UI must deny a route/message mismatch before optimistic cache
write or transport. Server capabilities remain authoritative; own authorship
is an additional existing condition, not a replacement permission.

A committed actor/channel/message identity change resets the local edit mode
and draft before a new scope can use it. Old-scope completions must not toast,
invalidate a later scope or clear its pending state, including A-B-A.
HTTP already sent is not aborted or undone by this local guard. This is not
a server ordering/idempotency or rollback guarantee across lifecycles.

## Exact ownership and limits

Writer after test-first RED owns only:
- frontend/src/components/messenger/message-item.tsx: edit admission,
  local lifecycle/pending, save settlement and edit prop wiring;
- frontend/src/components/messenger/message-actions.tsx: optional editPending
  default false, disabling only existing Edit ActionButton;
- frontend/src/i18n/locales/en/messenger.json and ru/messenger.json:
  messageEdit.error fallback only.
Oracle first owns:
- frontend/tests/messenger/message-edit-lifecycle.test.tsx new genuine unit.
Root owns this plan and same-session STATE/COVERAGE/domain/test docs.

No DELETE, pin, reaction, forward, send/composer, API, hook/query schema,
shared error formatter, routes, stores, backend, dependencies, test launcher,
Mail/Calendar or pro-qa-ai changes. Preserve d10f1dda search and published
delete/thread wiring. No general mutation framework or broad refactor.
Target <=150 changed production lines plus two tiny locale entries;
return BLOCKED for root decomposition rather than exceeding single behavior.
No child/agent forks from writers. Root owns production patch actuation,
Docker QA, Git commits and immediate authorized push. After root payload
preflight, the oracle writer may apply exactly its own new test file using
official apply_patch; no shell file-write alternatives or other edits.

## Frozen behavior

1. Edit entry and save require mounted current row, nonempty captured actor
   and message ID, live actor equality, message.channel_id===channelId,
   existing own-message condition and both current prop/store can_edit.
   Current channel revoke denies. No role inference, new authority or
   fallback to old projection. Unknown/foreign/mismatched contexts do not
   mutate cache or invoke PUT.
2. Same-scope save has one synchronous request token established before
   optimistic write/HTTP. Keep Edit visible but disabled until that request
   settles; start/save handlers also deny duplicate entry synchronously.
   Do not disable independent Reply/Delete/other actions merely due to edit.
3. Keep immutable request snapshot user/channel/message + lifecycle identity,
   exact origin messagesQueryKey and content payload. Lifecycle references
   change in commit/layout setup/cleanup and event handlers, not render.
   One request clears only its own token. Committed scope ABA, revoke,
   current authority loss and disposal make old work obsolete.
4. Empty or unchanged trimmed text keeps existing no-PUT/no-cache/no-toast
   completion. Authorized changed text preserves the existing document JSON,
   trimmed content_text and optimistic timeline projection, then closes
   editor immediately. No new optimistic thread projection is introduced.
5. Current fulfilled PUT clears its pending state, without new success toast,
   response synthesis or new invalidation behavior.
6. Current rejected PUT invalidates only the captured identity/channel
   timeline key exactly and calls toastApiError(error,t("messageEdit.error"))
   once, with RU "Не удалось изменить сообщение" and EN
   "Unable to edit message". Shared formatter can prioritize API message;
   this contract promises a localized fallback, not translation of all errors.
   Pending clears so the user may explicitly edit/save again.
7. Stale completion does not toast, invalidate or clear newer work. Cleanup
   must consume rejected promises; no detached unhandled rejection or auto
   retry. It must not broaden invalidation or directly restore stale cache.
8. Local editText/isEditing reset on committed actor/channel/message change
   and capability/revoke loss, even when the component itself is reused.
   Preserve existing editor labels, textarea, Enter/Shift+Enter/Escape and
   cancel/save control semantics. Browser keyboard acceptance is NOT_RUN.

## Test-first oracle and acceptance

Render REAL MessageItem and MessageActions with real QueryClient and stores;
mock only editMessage HTTP, toastApiError boundary and unrelated expensive
children/browser primitives. Use deterministic deferred transport and explicit
settlement witnesses; no missing dependency or helper RED. Seed exact origin
timeline data when checking optimistic writes, with an unrelated identity/
channel cache to prove non-interference. Restore stores, locale and mocks.

Required positive/negative groups:
- authorized own edit sends exact trimmed payload and existing optimistic
  origin update, with no mutation of unrelated cache;
- empty/unchanged input no-op; own/unknown/other-author/capability/mismatch
  admission, including mismatch before cache/HTTP;
- deferred PUT closes editor, disables only Edit and prevents re-entry or
  duplicate saves until settlement; manual post-error edit/retry is allowed;
- current failure canonical helper and exact invalidation, RU/EN fallback;
- late success/error across committed user/channel/message A-B-A cannot
  clear a second current pending save or invalidate/toast for it;
- revoke, authority loss and unmount during deferred rejection settle safely;
- a reused component closes/reset draft on user/channel/message changes.
Do not mock the target handler or test source text/AST. Do not count a
fireEvent.keyDown as a real browser-generated click witness. Existing
728-case suite and frozen search/delete/thread oracles must stay unchanged.

Canonical RED/GREEN only:
docker compose -f docker-compose.local.yml run --rm frontend-unit
Existing image,2GiB/no-swap,serial root-owned resource checks; no parallel
Mail build/test contention, image rebuild/pull, dependency install or host QA.
Independent Terra/high plan/source review; explicit Terra/xhigh atomic oracle
and implementation. Retry1/fix2/rereview2; failure triage first distinguishes
ROOT=code/test/infra and records rejected evidence without weakening tests.
Accepted commit is immediately pushed only to calls-feature-codex at origin
https://github.com/Valerijkk/hub.git and remote SHA checked.
Product/browser/server/storage acceptance remains NOT_RUN/UNBOUND.

## Oracle preflight checkpoint

The23-case payload was reviewed before application. Root corrected one
ROOT=test defect: user A-B-A changes in one act were not committed B
evidence; the oracle now commits separately and witnesses B's absent Edit.
Whitespace also exercises the real Enter save path instead of only testing
the disabled button. Assertions/timeouts/mocks otherwise remain unchanged.
These are pre-RED test corrections, not product failures or runtime evidence.

Independent review requested an explicit unsaved-draft capability-loss case:
remove live can_edit while editing, observe close, restore authority, reopen
with original content and no API/cache effects. This24th case is applied.
Final narrow oracle rereview C/I/M0/0/0; ready for intended RED. At00:56UTC
Mail builder showed active qa-runtime and backend-unit jobs, so root did
not start our canonical frontend-unit run. No production patch is applied.

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

## Source payload preflight

Independent mixed-RED triage C0/I2/M0 authorized exact production writers
from29 intended failures; seven UNKNOWN outcomes remain acceptance blockers.
First edit payload was NOT_APPLIED: root requested smaller bounded code,
mounted/current-scope entry admission and removal of obsolete-scope exact
invalidation after optimistic cache notification. Writer revision is pending;
no oracle, assertion, timeout or existing DELETE behavior was changed.

Final revised candidate is applied:148 changed production lines in item/
actions plus6 tiny locale lines. Root preflight also restored synchronous
start/re-entry denial, exact rendered callback scope comparisons and the
immutable optimistic scope.messageId target. An equivalent two-guard
format keeps the bounded diff readable; no test or policy change.
Independent source review C/I/M0/0/0. Final source SHA256:
- item C810D9F2AB90F8335576847468857A553DE32D4A0679AA662E5BFFFB9AE684FA;
- actions7F1EBA7158BE19754BCB791ADC467C3FB995E42A98AA4FBD60B213FE944A13C9;
- en9008E618DCC7EB8AF237AB9D800766FD0E360F1F24226E22C808D0FB7CE4A9FC;
- ru34274B9361409FFE658B4F8AD2255103139A49F3521616AA85F48A3F56C24170.
Full canonical GREEN candidate started2026-09-06T01:22:11.265763406Z;
all frozen oracle hashes above remained unchanged before dispatch.

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
