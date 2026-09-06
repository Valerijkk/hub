# Approved backend 4 GiB resource diagnostics

> **For agentic workers:** Use superpowers:subagent-driven-development for bounded independent review. Root alone owns this temporary resource configuration, exact Docker execution, restoration, evidence and Git.

**Goal:** Diagnose the backend resource blocker within the founder-approved hard4GiB/no-swap ceiling. Attempt1 tested memory; the separately bounded attempt2 below tests scratch capacity. Neither is canonical acceptance.

**Architecture:** Attempt1 temporarily changed the backend service's two memory scalars and matching cgroup assertion constant. Attempt2 additionally pairs the scratch tmpfs capacity with its exact assertion, as frozen below. Preserve all command/stage/test/policy/receipt integrity checks. Restore exact original source bytes after completion or abort. A separate canonical observer policy change is not authorized by this plan.

**Tech Stack:** Existing root Docker Compose, task-owned backend-unit image, Go1.25.11, unchanged serial build/vet/unit/race launcher.

**Spec:** User's 2026-09-06 reply «разрешаю» approves the preceding explicit request: temporary4GiB/no-swap for one backend QA container. It supersedes only the earlier UNAPPROVED status of this attempt, not Jitsi/infra or permanent QA policy.

## Global constraints

- Base fe87277207ed7ff7ea158219bd1963707e1512a9, calls-feature-codex; only approved origin https://github.com/Valerijkk/hub.git. Immediate push after each commit; verify SHA.
- No retry at2GiB. First4GiB attempt is complete and infra-aborted; the separate root decision below permits exactly one further scratch-capacity experiment, one container at a time, within the same4GiB/no-swap ceiling. No automatic third attempt.
- Root owns temporary changes only to docker-compose.local.yml/backend-unit and testframework/docker/backend_unit/runtime.sh; exactly3 scalar replacements for attempt1, exactly5 for attempt2.
- CPU6, no swap, read-only, pids256, GOFLAGS=-p=1, GOPROXY=off, unchanged GOGC/GOMEMLIMIT/default toolchain, /tmp64MiB, GOCACHE768MiB, cache watermark512MiB and report64MiB remain unchanged. GOTMPDIR384MiB in attempt1 becomes1GiB only in attempt2.
- Existing image sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9; no build/pull/install, host QA, alternate command, cleanup/prune, Mail/Calendar/pro-qa-ai access or changes.
- Start only without active adjacent builds or QA jobs and with documented memory headroom. Conservative host-only admission needs6GiB available RAM (4GiB cap plus2GiB reserve); if a separate read-only Docker-host memory observation establishes already allocated available RAM, root must record that evidence before refining admission.
- Stop only our exact owned container on observed OOM, sustained dangerous memory/disk pressure, no progress or30min attempt bound; preserve partial artifacts. Do not alter test timeout150s, tests or stages.
- Authoritative observer still seals2GiB. Explicit one-off evidence is EXECUTED_NONCANONICAL / NOT_ACCEPTED at that gate, even if Go tests pass. It cannot unlock CALL-DUR-08C2's canonical acceptance. Do not patch/bypass observers, policy seals, resource validators or current-source manifest checks.
- Requested independent reviewer Terra/high; actual UNKNOWN. No child agents, reviewer edits or duplicate execution. This is a diagnostic resource atom, not product implementation.

## Frozen edit and restoration

Only:
~~~diff
 # docker-compose.local.yml, within backend-unit, immediately after cpus:6.0
-    mem_limit: 2g
-    memswap_limit: 2g
+    mem_limit: 4g
+    memswap_limit: 4g
 # testframework/docker/backend_unit/runtime.sh
-CGROUP_MEMORY_BYTES=2147483648
+CGROUP_MEMORY_BYTES=4294967296
~~~

Default restore hashes:
- Compose E8DA0EEC4F99FE3B437C885B3B19019F27D24958D477152176F7210CB4A56733.
- runtime.sh C4A8BD83E75A0F133EF939496C4DD07644C3E21B83FA68F0C76A383B334FBCF4.

Unchanged launcher hashes:
- entrypoint7B361D57728C83F5A0C161C6E92AEC2140E6646AB8508A3E03BF89F290D7CB66
- plan.sh574D0651DD395870B900CFCA68A42B823A2E4A416EFE7CD5E8D37C7A53A1DF49
- execute.shB93682138610F9D3ED60010E8AE8B29F761FD94DD03E55DE778CF98D40AD4661
- receipt.sh2013FC844AE3EBD38EDB7C99712F8E8A6667CE5ACC73F9275B1ED48E6920E79E

Root uses apply_patch for both application and restoration, never checkout/reset or shell writes. Do not commit transient4GiB files as a permanent default.

## Task1: Controlled resource attempt (historical, ended)

- [x] Read prior actual OOM evidence; approve only the single memory-variable experiment. No repeated RED is necessary: three prior2GiB attempts already failed.
- [x] Independent read-only review of internal and observer memory contracts.
- [x] Wait for an admissible resource window. Initial12:19:53UTC preflight blocked on Mail QA/RAM; recorded guest-memory refinement and12:32:25UTC preflight later admitted the run.
- [x] Apply exact3 scalar substitutions immediately before execution; inspect scoped diff/hash. Root self-check: only two named paths, no command or assertion weakening.
- [x] Run only: docker compose -f docker-compose.local.yml run --rm backend-unit.
- [x] Inspect exact container engine/image/start/cgroup limits/worker commands; capture Docker lifecycle events including OOM/die/destroy and bounded resource samples. Never print secrets or all env.
- [x] Preserve unique partial/complete report directory without overwriting prior attempts. Read exact stage statuses, real Go failures and completed/skipped coverage. Nonzero infrastructure abort is not a source/test failure.
- [x] Restore both files after process completion/owned abort; compare exact original hashes.
- [x] Independent review of actual evidence; Go diagnostic result separated from canonical observer, product, API/media readiness. Publication is tracked below.

## Independent review boundary

Terra/high reviewed the initial3-scalar proposal: in-container assertions and serializer follow the shared constant, but canonical observer rejects4GiB.
Exact consumers: testframework/observers/backend_unit/observe.py:122,615;
contracts.py:157,872; sealed testframework/policies/backend-unit-observer.v1.json:174.
lease.py:26 binds Compose and runtime.sh as current build inputs, so restoration also changes the canonical manifest.
Direct backend-unit-receipt packet validation is not authoritative Docker host-limit acceptance.
Root ruling: retain observer safeguards; the authorized one-off attempt is diagnostic only. No canonical GREEN or downstream acceptance follows from it.

## Initial root self-review (attempt1)

One behavior and3 scalars; user authority is literal and temporary; only one exact QA service. Resource and provenance limits are explicit. No tests or policy sealed defaults are changed. No additional implementation or external authority is inferred. Resume automatically when the recorded workload/resource preconditions become true; do not request the same4GiB permission again.

## Docker-host memory refinement — 2026-09-06

Read-only observation of already-running docker-desktop WSL /proc/meminfo:
MemTotal15962444KiB, MemAvailable14316644KiB, Cached11516576KiB,
AnonPages798752KiB. Windows Available5407MiB and vmmem working set13633871872B
are not the guest's available-memory ceiling; substantial allocated guest
RAM is reclaimable file cache. Root refines admission to guest MemAvailable
>=6GiB plus Windows Available>=2GiB, no active adjacent build/QA and quiet
disk. No cache eviction, WSL settings or swap state was changed. Existing
guest swap use is not task swap: the new container must independently prove
memory.swap.max=0 and matching Docker Memory/MemorySwap4GiB.
At12:23:36UTC a new Mail build was active, so no source edit/run occurred.

## Attempt1 result and restoration — 2026-09-06

Published authority checkpoint9d42e76a4190cd239a4829013a3ae9c00f4b912a;
immediate push succeeded, one read-only SHA retry after network timeout
confirmed equality. Exact backend-unit command ran from the existing image.
Outer worktree-master-backend-unit-run-155619b88d71, engine
c9f4c49b91d6a6e752d2f959ab780003b017f477afc329af15e33ebc584fbb6e,
started12:33:02.213440007Z; Docker die75/destroy at12:36:17UTC after195s,
Compose exit1. Closed observed events contain no OOM. No container remains.
Actual4GiB/no-swap6CPUreadonlypids256, p1/offline/default Go env confirmed.
Samples478.6MiB/1.342GiB are not peaks; disk1%/queue0. No root build/pull/install.

ROOT=infra, TMPFS_ENOSPC. Partial report directory
`testframework/reports/backend-unit-c9f4c49b91d6.partial` retained unchanged:
owner, four package plans and four unpublished artifacts. Raw561-byte log
`.backend-unit-publish.8Om43RVZ` names backend/ent output `$WORK/b243/_pkg_.a`,
`/var/tmp/hub-go` importcfg and indexer linker ENOSPC. The263-byte
`.backend-unit-publish.Gmrw3nPN` reports the first two build packages FAIL.
Final console: head write ENOSPC; BACKEND_UNIT_ABORT_MONITOR. No completed
build stage or container receipt; vet/unit/race NOT_RUN, not passing tests.
Source trace: GOTMPDIR points to asserted384MiB tmpfs; stream scratch also
lives there. No evidence of SSD capacity exhaustion. Independent Terra/high
RCA confidence HIGH, no technical objection to the paired scratch adjustment.

Both temporary files restored by apply_patch and exact original SHA matches:
Compose E8DA0EEC4F99FE3B437C885B3B19019F27D24958D477152176F7210CB4A56733;
runtime C4A8BD83E75A0F133EF939496C4DD07644C3E21B83FA68F0C76A383B334FBCF4.

## Root decision: separate scratch-capacity experiment

Ruling: the user approved temporary4GiB/no-swap for one backend QA container,
not a globally consumable one-run token. The initial one-attempt stop was
root's experiment boundary. Fresh ENOSPC evidence justifies one separate
single-variable diagnostic before stopping; no new RAM ceiling, disk mount,
test/command change or canonical policy authority is inferred.

Relative to attempt1, change only GOTMPDIR capacity384MiB ->1GiB in the
Compose tmpfs and its exact matching runtime assertion. Retain all other
attempt1 parameters. Within4GiB, tmpfs ceilings now total1856MiB, leaving
nominal2240MiB for compiler/runtime; actual allocation is shared and may
still OOM. If wrong, this costs one bounded recomputation and another
preserved infra-abort, not an unbounded retry or SSD scratch fallback.

Exact additional scalar changes, alongside the same3 memory substitutions:
~~~diff
-      - "/var/tmp/hub-go:rw,nosuid,nodev,exec,size=384m,uid=10001,gid=10001,mode=1770"
+      - "/var/tmp/hub-go:rw,nosuid,nodev,exec,size=1g,uid=10001,gid=10001,mode=1770"
-TMPFS_GOTMPDIR_BYTES=402653184
+TMPFS_GOTMPDIR_BYTES=1073741824
~~~

Root alone applies/restores exactly5 scalars across the same2 paths. Fresh
resource preflight and same existing image/exact command are mandatory.
Stop on observed OOM, unsafe pressure or30min bound. No third attempt or
further capacity increase is authorized by this ruling. Independent review
technical blockers0; required root decision/provenance condition is resolved
here. The attempt remains diagnostic/noncanonical; observer/manifest limits
and no full-module acceptance remain unchanged.

- [x] Attempt2 fresh preflight, exact5-scalar application and same Docker run (ended; result below).
- [x] Preserve result/events; restore both original hashes.
- [x] Independent final evidence review.
- [ ] Exact-path docs commit, immediate approved push and remote SHA verification.

Attempt2 preflight12:42:29UTC: active builds0, guest MemAvailable13268780KiB,
WindowsAvailable6553MiB, disk18%/queue0, but adjacent Mail frontend-unit was
active. Root did not apply temporary edits or start a QA job; defaults remain
byte-exact while waiting for a fresh admissible window.

Attempt2 launch record (historical): fresh12:44:36UTC preflight admitted the run: no
active builds or QA containers, guestAvailable14304760KiB, Windows7491MiB,
disk0%/queue0. Started12:45:01.918343534Z, outer
worktree-master-backend-unit-run-2e64073e3ea2, engine
f5899ea672f93167b4c258db91acedce860d8c7e7e379cad1efa14928acaa9fd,
hostname f5899ea672f9. Unique backend-unit-f5899ea672f9.partial retained.
Actual4GiB/no-swap6CPUreadonlypids256 and64/768/1024MiB tmpfs confirmed;
image unchanged. Temporary Compose SHA256
EAC2FBB195DEFAD5AE1200D53A8AB8E456A1E6DFDB7776BBA276FE2D25D740F2;
runtime41F6FE610411A682BF20BF2CED5F32C3ED5DDBB8C44874C3413C517986EC00AC.
Independent pre-result review confirmed exactly5 scalars and no unexpected
edits or assertion weakening; its progress-record I1 is resolved here.
Initial242.8MiB and12:46:16UTC3.326GiB are samples, not peaks. At the latter
sample no OOM event, disk20%/queue2, Windows3637MiB, guest10733420KiB.
At that checkpoint no result or canonical acceptance existed. Final result
and exact restoration now follow.

## Attempt2 result — infra abort, original defaults restored

Started12:45:01.918343534Z; Docker die75 at12:48:48.850822047Z, destroy at
12:48:48.862405769Z, approximately226.93s. Compose exit1; session42324 is
closed and the disposable container is absent. The closed Docker event
window12:44:58Z..12:50:11.0449462Z contains die/destroy and no OOM event.
This is not a measured peak or a guarantee that all future stages fit4GiB.

Exact terminal diagnostic:
~~~text
head: error writing 'standard output': No space left on device
BACKEND_UNIT_ABORT_MONITOR
~~~

The preserved263-byte `.backend-unit-publish.ccdreVob` contains only:
~~~jsonl
{"ordinal":1,"import_path":"github.com/valerii/hub/backend/cmd/hub-mail","package_name":"main","result":"PASS","exit_code":0}
{"ordinal":2,"import_path":"github.com/valerii/hub/backend/cmd/meilisearch-indexer","package_name":"main","result":"PASS","exit_code":0}
~~~
Its SHA256 is AEB305C57EA2006C3CD860EBE530968BA0D587D3E27EC1AD2A2800608EF261DB.
Three other unpublished artifacts are empty; owner and four plans remain.
The unique nine-file partial directory is preserved without edits, rename,
overwrite or cleanup. Plans are discovery inventories, not execution proof.
No build-stage receipt or final container receipt exists; vet/unit/race
NOT_RUN. The last observed Go command was cmd/server at12:48:11UTC; its
exit code and original Go diagnostic were not captured. Do not fabricate a
server compiler failure from the monitor error or claim two packages as a
complete build/test pass.

ROOT=infra: monitor stream capture exhausted the asserted1GiB GOTMPDIR
tmpfs. Unlike attempt1, no package-level ENOSPC log survived. The specific
largest files, byte/inode consumption and underlying Go diagnostic are
UNKNOWN; no SSD capacity exhaustion is evidenced. A future diagnostic
contract needs safe, bounded scratch-usage evidence before teardown, not
another blind capacity increase. No third attempt is authorized here.

Resource samples (not peaks):242.8MiB initially,3.326GiB at12:46:16UTC,
422.4MiB at12:47:02UTC and1.433GiB at12:48:11UTC. Docker reported blockIO0B
for those samples, not total Windows disk traffic. Corresponding later
disk samples20%/queue2,5%/queue0,1%/queue0; post-exit0%/queue7 is an isolated
snapshot, not proof of sustained queue pressure. No root image build/pull,
swap/disk scratch change, dependency install or neighboring resource stop.

Root restored both files using apply_patch after exit; exact original hashes
match the restoration section. plan.sh, execute.sh, receipt.sh and entrypoint
also match all recorded original hashes. Git source diff for both temporary
paths is empty. This run is EXECUTED_NONCANONICAL / INFRA_ABORT; authoritative
observer acceptance remains NOT_ACCEPTED. CALL-DUR-08C2 stays STATIC_ACCEPTED
without executable evidence; Messenger/Calls/Jitsi product gates are unchanged.

## Independent final review and next-contract boundary

Terra/high, actual UNKNOWN: C/I/M0/1/0. No unexpected source change or
restoration defect. The remaining I1 is the next diagnostic contract, not
permission to retry or a missing product assertion repair. Confirmed source
chain: execute.sh:218 creates stage scratch under GOTMPDIR; :310 writes
reader output to that scratch; :1328 aborts/rolls back the unpublished stage.
The immediate write failure is proven; cmd/server's own result is UNKNOWN.

Before any further backend diagnostic, separately freeze:
- a bounded, integrity-bound abort record surviving stream-reader failure;
- package/stage and which stdout/stderr reader failed;
- target mount identity, bytes/inodes used/free/capacity and cgroup events;
- retained causal stream/error outside the failing GOTMPDIR, without an
  unbounded SSD log or a new scratch fallback;
- a combined budget for Go work, retained build outputs and stream captures;
- separate profile/observer/lease compatibility for any canonical acceptance.

These requirements do not authorize implementation, a third run, a new QA
service, altered policy seal or a resource increase. Root records the open
contract instead of treating C0 as full acceptance. No product source or
test was changed during these two diagnostics.
