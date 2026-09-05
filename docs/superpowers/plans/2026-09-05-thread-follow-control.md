# Messenger thread follow control

Status: UNIT_ACCEPTED; Docker RED/GREEN and independent review complete.
Requirement: MSG151-FOL-01/02 and MSG151-CHN-02; one existing API behavior,
not a new backend feature. Owner branch calls-feature-codex.

## Contract

- Add a small ThreadFollowControl in the existing ThreadPanel header. Keep
  the thread-detail query's full ThreadResponse and render its parent field;
  keep the existing identity-bound key and independent paged replies.
- Reuse useChannels() and channelsQueryKey(userId). No new capability query,
  no ChatView unscoped-cache dependency, no permission inference from roles.
  A current authenticated user, non-revoked channel, successful non-fetching
  channel list containing this channel, and successful non-fetching thread
  detail are required before sending any mutation.
- Exact modes: absent/undefined or manual_unfollow => Follow when can_follow
  is exactly true; auto or manual_follow => Unfollow when can_unfollow is
  exactly true. Other values (including auto_follow, empty or padded strings)
  fail closed. Missing/denied capability never enables the other action.
- Use existing followThread/unfollowThread with exact channel/root IDs.
  Single-flight guard and disabled pending button prevent duplicate requests.
  Current identity/revocation is rechecked immediately before dispatch and
  after settlement. An obsolete/unmounted user/channel/root scope must not
  issue cache updates or error toasts; key the control by that scope and guard
  its lifecycle. No cancellation claim for already sent server requests.
- On success, invalidate only the origin threadDetailQueryKey and
  unreadQueryKey(userId), exact:true. No setQueryData, optimistic follow mode,
  local unread arithmetic or mutation of the Messenger store. Both changed
  true and changed false responses trigger authoritative refetch.
  Dispatch requires idle authorized resources; successful completion does not.
  After each await, only the current lifecycle, user and revoke barrier gate
  exact invalidations. A same-scope background refetch, changed follow mode or
  capability snapshot must not silently discard successful server state;
  the read endpoints still authorize their requests. This distinction was
  frozen during independent review I1 before the corrective regression.
- Keep pending until those invalidations settle. Keep the same button node
  while an owned mutation or any background detail/channel refetch is pending,
  provided existing data, identity and access remain valid; disable activation
  while either resource is refetching. Missing/denied/error/auth/revoke hides
  the control rather than trusting a stale authorization result.
  On API error preserve state, report via toastApiError with localized text,
  and allow an explicit retry. No automatic mutation retries.
- Existing Button, type=button, visible focus, compact text labels, aria-busy
  and local status announcement. New labels in ru and en. Preserve close
  button and focus-return behavior, native right/mobile panel and timeline.

Frozen messenger locale keys: threadFollow.follow = EN Follow / RU
Подписаться; threadFollow.unfollow = EN Unfollow / RU Отписаться;
threadFollow.updating = EN Updating subscription… / RU Обновление подписки…;
threadFollow.error = EN Unable to update thread subscription / RU
Не удалось изменить подписку на тред. Error/retry oracle checks both literal
localized fallback and real i18n lookup; shared toast formatter is unchanged.

## Exact atomic ownership

- New frontend/src/components/messenger/thread-follow-control.tsx.
- frontend/src/components/messenger/thread-panel.tsx: full result, guarded
  scope-keyed control mount/props, header integration only.
- frontend/src/i18n/locales/{ru,en}/messenger.json: new follow labels only.
- New frontend/tests/messenger/thread-follow-control.test.tsx: render real
  ThreadPanel; mock HTTP API functions as hermetic boundaries and unrelated
  children only. Use real QueryClient, useChannels and auth/revoke stores.
- frontend/tests/messenger/phase151-thread-panel.test.tsx and
  frontend/tests/messenger/phase151-thread-query-identity.test.tsx and
  frontend/tests/messenger/phase151-thread-panel-convergence.test.tsx: only add
  getChannels empty-result API fixture if needed. Preserve every assertion.
- Root owns this plan, .planning/STATE.md and affected usage documentation.

No backend, store, API, ChatView, shared hook, dependency, Docker runner,
generated artifact, Mail/Calendar or protected-zone changes.

## Verification sequence

- [x] Terra/xhigh writes only tests first. Expected behavioral RED: existing
  ThreadPanel lacks named Follow/Unfollow buttons; no missing-module RED.
- [x] Root runs canonical frontend-unit in 2GiB/no-swap, existing image,
  sequential with no other heavy QA. Preserve fixed report files before reuse.
- [x] Terra/xhigh implements only the frozen behavior after observed RED.
- [x] Canonical frontend-unit GREEN (TypeScript plus full Vitest), then
  independent Terra/high review and bounded corrections with a repeated gate.
- [x] Docs sync in the same session. Publication uses an exact-path commit
  followed immediately by the authorized branch push; the resulting commit
  and remote equality are recorded in the terminal handoff, not a self-hash.

Unit matrix: absent/manual_unfollow Follow; auto/manual_follow Unfollow;
unknown mode, missing/false capability, absent channel, no user, revocation,
loading/error/refetch deny; exact route and duplicate click; changed=false;
server mode change after refetch; API error/retry; late auth switch/revocation/
root switch no cache/toast effects; only exact origin detail/unread refresh,
no local snapshot arithmetic; keyboard button activation and close focus
regression. Test no live server or database. Browser/media remain NOT_RUN.

Baseline prerequisite: FE-QA-2G 914df3d2 pushed, TypeScript PASS and Vitest
189 files/601 tests PASS inside actual 2GiB/no-swap Docker container.
Backend Go acceptance remains blocked by confirmed Ent compilation OOM;
the requested 4GiB Go-only limit has not been approved.

## Execution evidence

Seed behavioral RED: canonical frontend-unit outer9174a10aedfc, engine
d04d84a03b0eb0b118f078e85d83425009dba65f49cc15fa6c0c65663a00f927,
TypeScript PASS; Vitest601PASS/1FAIL (602cases,190files,60.26s). The only
failure is the new Follow button lookup, not an import/harness error.
Docker closed events start1788634101, die1 at1788634181 after79s, destroy;
noOOM in that interval. Container already auto-removed before inspect, so
no per-run inspect/peak-memory claim. Compose source retains the previously
verified2GiB/no-swap envelope. Prior baseline files copied to
testframework/reports/frontend-unit-before-thread-red-20260905/.
Expanded behavioral matrix and independent preflight review are in progress;
production implementation has not started. Fresh user continuation restored
agent dispatch after the prior session's thread limit; review is not waived.
Independent Terra/high preflight C/I/M0/1/0 identified the authenticated
convergence test's additional real getChannels boundary. Root verified and
froze only its empty-result mock fixture in the exact ownership above;
all its existing realtime/cache assertions must remain unchanged. The
expanded oracle worker owns that four-line addition before the next RED.
Expanded RED: engine50416107f6bbf3d67082c7b358377bb621d35ed354175cd5e3cde37744e63042,
outer1a21891de3e4, TypeScriptPASS; Vitest603PASS25FAIL628cases65.78s.
The prior189files/601tests remainPASS. Verified2GiB/no-swap/6CPU/readonly;
Dockerexit1 after82s with noOOM; reports preserved in
frontend-unit-thread-matrix-red-20260905. Independent test review then found
one non-exact localized-error assertion; corrected EN/RU parameterization
was re-reviewed and I1 closed, with no remaining reported findings.
Final preproduction RED: outera92bfeea2876,
engine90aaf631677d45dfee1bc3196d142ca051df74beec049f20c5a0368f3844b9cd,
start2026-09-05T19:24:44.581629827Z, inspected2GiB/no-swap/6CPU/readonly.
TypeScriptPASS; Vitest603PASS26FAIL629cases190files54.89s; Composeexit1.
Only the new follow-control file fails; missing channel observer/control is
the expected absent integration, not an import/harness failure. Root now
authorizes the exact production atom; no module/browser/Jitsi acceptance.

Next discovery (NOT_FROZEN, not part of this atom): markThreadRead is exported
by frontend/src/api/messenger.ts but has no production caller (only API unit
test references). Opening/reading a followed thread still needs a separately
scoped visible-read receipt integration; a working Follow button alone must
not be reported as complete Slack-like thread/unread behavior.

Post-implementation independent review: C/I/M0/1/0. I1 identified reuse of
the idle-resource dispatch guard after transport success, dropping required
invalidations during a background refetch. A deferred post-dispatch channel
refetch regression is being added before correction. The first draft's
incorrect pre-dispatch refetch ordering was rejected before application.
Initial production GREEN was deferred because the already running Mail
builder had an active backend qa-runtime build; no overlapping QA was run.

I1 Docker RED: enginef8ad6347835f34ae39fcd2edfcc47405e015c2bfb69f2f7ebb806a9630c12645,
outera59b1560e97b, TypeScriptPASS; Vitest625PASS5FAIL630cases72.50s.
The new post-dispatch refetch case fails with zero invalidation calls;
the other28 new cases pass. Four unchanged admin/composer cases hit15s
timeouts. Read-only import/mount triage found no direct path to the new
control; their runner-level cause remains UNKNOWN. No assertion, timeout,
test selection or resource profile was changed. Reports are retained in
frontend-unit-thread-race-red-20260905; engineexit1 after89s, noOOM.
Correction removes only the reused post-success resource guard, retaining
all dispatch checks and both lifecycle/identity/revoke completion barriers.

Final canonical GREEN: docker compose -f docker-compose.local.yml run --rm
frontend-unit, outeraddc0ae61dc4,
engineec2497fbfc5041c2cc404047654c2aa5d30bbbf69fa49bae62f32c2e66888087,
start2026-09-05T19:50:00.518892544Z. Inspected memory/memory-swap2147483648,
CPU6000000000, readonlytrue; existing image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
TypeScriptPASS; Vitest190files/630testsPASS54.44s, including29 new cases
and all four previously timed-out cases unchanged. Compose/engineexit0;
closed events start1788637800, die0 at1788637872 after71s, destroy, noOOM.
Preserved small reports: testframework/reports/frontend-unit-thread-follow-green-20260905/.
Only current result/typescript/vitest files belong to this receipt; old
Biome/Vite/typescript-tests files copied with the directory are historical,
not new evidence. No build, pull, install, host QA or Mail mutation occurred.
Final independent Terra/high source/spec/oracle review C/I/M0/0/0.
Requested downstream model/effort were explicit; actual telemetry UNKNOWN.

Tested SHA-256:
- thread-follow-control.tsx: 0FA792B48C5F9F5D5C1AD0F3E42902CB4DC963138A85C0AC51C4192D097FDE8E
- thread-panel.tsx: 439BEAB4A247496DF08C78A6219378E9CFF3404F6B0C6DC0ACC110D176B0C688
- thread-follow-control.test.tsx: 8E2ECFED6F33B7E63502C531DD5E6B765BDB88FAC70C59AD67379F89E3330520
- en/messenger.json: C9EAB9E649A928B01B095499FBE5412620BB7DB11AEADB273719B57ABED81B3F
- ru/messenger.json: AEB2E31D6B8182F8880A93D09BC4C9BE6620441BE7B9613133E48B15A94D7F0B

Acceptance is this unit-tested UI behavior only. Real browser keyboard/focus,
API/storage/WS end-to-end, automatic visible-read receipts, Jitsi deployment,
real join/media and module readiness remain NOT_RUN/UNBOUND. Backend Go
remains BLOCKED by2GiB Ent compiler OOM; no further2GiB retry is authorized
by this plan, and the requested4GiB Go-only exception is still unapproved.
