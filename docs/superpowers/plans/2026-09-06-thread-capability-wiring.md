# Thread message capability projection wiring

Status: UNIT_ACCEPTED in the composed2026-09-06 Docker frontend gate.
Independent Terra/high plan and source review C/I/M0/0/0.7new genuine cases
preserve the intended4fail RED; all now pass without weakened assertions.

Root oracle preflight repaired two test-only risks before execution: unrelated
MessageComposer/MessageItem are mocked in the real ChatView owner-boundary
test; all new real ThreadPanel cases await exactly two DOM row witnesses and
both Map.has(root/reply) before asserting optional projection. No missing
hook/export RED and no vacuous undefined-from-absent-row acceptance.
One prop-propagation behavior, separate from search and delete lifecycle.

## Evidence and outcome

ChatView already receives the server Channel through getChannel and passes
channel?.capabilities to timeline MessageItem. Its ThreadPanel receives only
canReply/canPost/projectId. ThreadPanel renders root and reply MessageItem
without capabilities and cannot reconstruct them from ThreadResponse.
MessageItem/MessageActions deliberately fail closed on undefined projection.
Thus loaded thread messages lose their existing action controls even when
the active channel has an authoritative projection. This is a missing wire,
not authority to infer roles or invent new permissions.

Pass the same optional current-channel projection through ThreadPanel to both
root and replies. Existing prop plus live store/revoke checks in MessageItem
stay unchanged. Unknown/loading/absent projection remains undefined, no
allow-by-default fallback. Direct callers may omit the new optional prop.

## Exact ownership and non-goals

- frontend/src/components/messenger/chat-view.tsx: only ThreadPanel
  capabilities={channel?.capabilities} wiring;
- frontend/src/components/messenger/thread-panel.tsx: ChannelCapabilities
  type import, optional documented prop/destructure, root/reply forwarding;
- frontend/tests/messenger/phase151-thread-panel.test.tsx: real ThreadPanel
  with existing mocked MessageItem boundary recording delivered capabilities;
- frontend/tests/messenger/thread-panel-channel-capabilities.test.tsx: new
  genuine ChatView owner-boundary unit with real Query/store and observed
  ThreadPanel prop; mock unrelated heavy children/hooks only;
- root owns docs/status.
No server/API/hook/store/sharedUI/message-actions/message-item behavior,
new queries, permission policy, locales, dependency/runner or Mail changes.
No pro-qa-ai access. Aim <=30 changed production lines in two files.

## Acceptance and oracle

1. Real ThreadPanel loaded canonical parent+reply delivers the exact passed
   projection to BOTH real child boundaries, not only root or composer.
2. Omitted projection stays undefined; a rerender replacing allowed with
   denied/undefined updates both children, no stale retained copy.
3. Real ChatView with active matching thread and resolved channel projection
   passes that projection to ThreadPanel; pending/absent projection does not
   fabricate permissions. Different-channel activeThread is not opened.
4. Live current-channel revoke preserves existing hidden panel behavior.
5. Preserve existing MessageItem guard tests and the full frontend regression.
   These wiring units do not claim actual backend authorization or browser
   action completion. TEST_NOT_RUN until root canonical Docker evidence.

Test-first must collect against baseline: pass extended ThreadPanel props
through a structurally assignable object/spread if needed rather than making
TypeScript fail solely because the production prop isn't declared yet.
Use installed RTL fireEvent, not unavailable user-event. Deferred HTTP cases
wait for an actual first-call/child witness before changing scope. Mocking
child boundary is allowed for parent prop-wiring oracle; do not mock target
ThreadPanel in its own tests or target ChatView in its own tests.

## Workflow and safety

Independent Terra/high source-aware plan/oracle review; Terra/xhigh sole
atomic writer. Root owns serial canonical command:
docker compose -f docker-compose.local.yml run --rm frontend-unit
Existing image2GiB/no-swap, workload precheck, no host QA/build/pull/install.
Do not mutate source during a root QA snapshot. Root accepts intended RED,
then source review, GREEN, docs-sync, exact commit and immediate authorized
calls-feature-codex push with remote equality check. max2 fixes per cause.
Root must record broader discovered policy gaps separately, not widen this
prop-only atom. All real SUT/browser/API/module acceptance remains separate.

## Combined Docker RED (2026-09-06 local)

Canonical command: docker compose -f docker-compose.local.yml run --rm frontend-unit
Existing image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218;
outer worktree-master-frontend-unit-run-c1effb4f13de;
engine fe485fa4ac6fb90911537ec75ac1ca6e5c8eab5c7f500b2bc184917f25ee556c,
start2026-09-05T23:10:23.769260921Z. Inspected2GiB/no-swap,6CPU,readonly;
Compose/engineexit1,79s, create/start/die/destroy observed with no OOM event,
container absent afterward. TypeScriptPASS,194files/716tests:
684PASS32FAIL,62.36s,plus3 unhandled rejections.

32 failures are18 already-known search lifecycle,10 new delete lifecycle and
4 new thread capability wiring. All prior666 assertions passed, but this is
NOT a green run: the uncaught DELETE errors make whole-run acceptance fail.
Preserved only eight current report artifacts:
testframework/reports/frontend-unit-delete-wiring-red-20260906/.
No build/pull/install/hostQA. Mail frontend/backend and data services stayed
running; its last builds were completed, no other QA container was listed.
Initial Docker event formatting failed to quote a map key; the read-only
subscription was corrected, without retrying the QA process.

Frozen source/oracle hashes stayed unchanged through the RED:
- delete oracle09A734D7C376C7680B1E390D50CF6A4EBCA0D7BED9FD9D1F6B5ECD082C0930C2;
- parent wiring oracle8403CB9492EFB9D5DE167B54F0E6FEEEBCE157557E273325D9B4BA9BD9E91AEB;
- owner wiring oracleF9E544E838517DB3D6ECA3DCD85CB1BD4ABA6B8A4B23598D54BF24617D3D2BD2;
- search oracle70E4899F92BC1560D57BAC754BF8B27C207159F1C509A8BCF24B21F9CA73D7EA.

Wiring ROOT=code: both parent/reply and ChatView owner observe undefined
instead of the provided projection in4 new cases. Missing/undefined and
other-channel negative cases remain preserved. Production patch adds only
the optional type/prop/destructure and three forwards, with no policy/query.

## Composed GREEN and acceptance (2026-09-06 local)

Canonical composed-tree gate (search + delete + thread wiring):
docker compose -f docker-compose.local.yml run --rm frontend-unit
TypeScript PASS;195 files/728 tests PASS,64.12s,Compose exit0. No whole-run
unhandled-error failure. Existing unrelated React act/description warnings
remain; this is not a clean-stderr claim. Observed engine
61238c7ed829fb461c15326296f38aa946cb1ebc330b4d99305b3a749a05536d,
outer worktree-master-frontend-unit-run-8768e627cfd5,
start2026-09-05T23:44:23.616980053Z, image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Inspected2GiB/no-swap,6CPU,readonly; container absent after completion.
Post-run Docker event lookups returned no records, so separate die/destroy
timestamps and OOM-event history are NOT_CAPTURED, not invented.
Eight current reports preserved at
testframework/reports/frontend-unit-messenger-lifecycle-green-20260906/.
No source/oracle changes during execution; recorded pre/post hashes match.
No build/pull/install/hostQA or modification of resident Mail services.
This tests the composed working tree, not each intermediate publication
commit in isolation. Browser/API/storage/media and whole modules remain
NOT_RUN/UNBOUND.

Target source hashes:
- thread-panel.tsx DA5E975D5BC8045427EEE46E9BCB19CB49C5968B6F2B956E49958465130F687A;
- chat-view.tsx 93DE131119F4598354CD887E67A64BE898526DFD3398B318675E224D7B9D460B.
The two frozen oracle hashes above stayed unchanged. Exact root/reply
projection, pending-to-resolved owner and missing/revoked/other-channel
negative cases are accepted as prop-boundary unit evidence only.
