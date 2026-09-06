# Canonical frontend-unit worker budget under the fixed2GiB ceiling

Status: UNIT_ACCEPTED in composed758-test Docker GREEN; publication pending.
Scope: one resource-envelope behavior of the existing exact frontend-unit
Compose service. This separate root-owned QA atom does not change any feature
oracle, test inclusion, assertion, timeout, command, image, dependency or RAM.
It explicitly supersedes only the no-QA-config non-goal of the two pending
edit/composer atoms; their product/source/oracle contracts remain frozen.

## Evidence and decision

Published d10f1dda had a composed728-case PASS. New edit/composer test-first
run has29 intended failures plus7 UNKNOWN outcomes. After source fixes and
C/I/M0/0/0 reviews, the first candidate reports740PASS12FAIL/752(expected758),
3whole-run errors,296.50s. It is NOT_ACCEPTED; exact8receipts:
testframework/reports/frontend-unit-edit-composer-attempt1-20260906/.
Two errors are Vitest worker onTaskUpdate timeouts; one is async i18n callback
after environment teardown. At01:26 the owned container has137PIDs,
1.815/2GiB,CPU535.30%,BlockIO0/0; no OOM event. This does not establish one
root cause for every failure or an observed exact Vitest worker count.

Root verified the locked Vitest3.2.4 primary source. Its forks/threads pools
default to available CPU count minus one; its resolver accepts the four
VITEST_MIN/MAX_THREADS/FORKS variables. Existing QA config specifies no worker
budget. A CPU-time quota and a RAM ceiling are not a declared Vitest budget.
Root chooses an explicit maximum of2 and minimum of1 for both supported pool
families, scoped only to frontend-unit in docker-compose.local.yml. Both
families are bounded without guessing which default pool the existing image
selected. This is a persistent canonical service resource limit, not an
ad-hoc CLI override, alternate runner or selective test bypass.

Primary sources:
- https://raw.githubusercontent.com/vitest-dev/vitest/v3.2.4/packages/vitest/src/node/config/resolveConfig.ts (466-518)
- https://raw.githubusercontent.com/vitest-dev/vitest/v3.2.4/packages/vitest/src/node/pools/forks.ts (63-76)
- https://raw.githubusercontent.com/vitest-dev/vitest/v3.2.4/packages/vitest/src/node/pools/threads.ts (42-55)

## Exact ownership and invariant

After independent plan preflight, one Terra/xhigh atomic author provides only
a tiny patch to docker-compose.local.yml/frontend-unit:
environment VITEST_MAX_THREADS:"2",VITEST_MIN_THREADS:"1",
VITEST_MAX_FORKS:"2",VITEST_MIN_FORKS:"1", with a short resource-budget comment.
Root applies it. No shared anchor or other service/environment changes.
Keep2GiB/no-swap,6CPU,readonly,pids limit,volumes,networks,entrypoint and
empty command identical. No locale/product source or oracle edits.
No host QA, new process runner, Docker exec test probe, build/pull/install,
image replacement, cache/volume sharing, cleanup/prune or Mail intervention.
Existing task-owned image remains sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Prior feature runs retain their original unbounded-pool provenance and
cannot be relabelled as having used this envelope.

## Acceptance / negative gate

Use only docker compose -f docker-compose.local.yml run --rm frontend-unit.
Preflight adjacent heavy jobs/resources; one root-owned workload at a time.
Inspect actual container environment for only these four nonsecret keys,
limits/image and Docker process-name tree/stats while Vitest is active.
Do not print all environment values. A process/PID count is not by itself
proof of exact worker count; source+selected pool evidence must qualify it.

Required final result: TypeScriptPASS; all197files/758tests present and PASS;
zero unhandled errors; engine/Composeexit0; no OOM; frozen feature source/
oracle hashes unchanged and no skipped tests. Report actual resource samples,
lifecycle events and exact eight small receipts. No retry-on-green counting,
no timeout inflation, no broad test exclusions, no claimed browser/API/media.
If resource-bounded run still fails, preserve failures and triage the first
concrete witness; no feature/assertion change based on a guessed load cause.
Retry1/fix2/rereview2 for this distinct resource atom, no endless rerun loop.
Independent source/diff review must confirm only service resource settings.
Same-session STATE + TESTING/test_plan/COVERAGE + stack_and_tools docs sync.
No module completion or Go/Jitsi unblock is inferred; Go4GiB is unapproved.

## Publication

Only accepted exact paths, then immediate calls-feature-codex push to origin
https://github.com/Valerijkk/hub.git and remote SHA verification. If tested as
a composed tree, label it as such rather than inventing isolated commit
runtime provenance. Root chooses safe atomic publication order after GREEN.

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
