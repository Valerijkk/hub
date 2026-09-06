# One approved backend 4 GiB attempt

> **For agentic workers:** Use superpowers:subagent-driven-development for bounded independent review. Root alone owns this temporary resource configuration, exact Docker execution, restoration, evidence and Git.

**Goal:** Execute one founder-approved backend-unit attempt with a hard4GiB/no-swap ceiling and learn whether the confirmed Ent compiler OOM is resolved.

**Architecture:** Temporarily change only the backend service's two memory scalars and the runner's matching cgroup assertion constant. Preserve all command/stage/test/policy/receipt integrity checks. Restore exact original source bytes after completion or abort. A separate canonical observer policy change is not authorized by this plan.

**Tech Stack:** Existing root Docker Compose, task-owned backend-unit image, Go1.25.11, unchanged serial build/vet/unit/race launcher.

**Spec:** User's 2026-09-06 reply «разрешаю» approves the preceding explicit request: temporary4GiB/no-swap for one backend QA container. It supersedes only the earlier UNAPPROVED status of this attempt, not Jitsi/infra or permanent QA policy.

## Global constraints

- Base fe87277207ed7ff7ea158219bd1963707e1512a9, calls-feature-codex; only approved origin https://github.com/Valerijkk/hub.git. Immediate push after each commit; verify SHA.
- One attempt; no retry at2GiB or additional4GiB attempt on failure without a new bounded decision/authority.
- Root owns temporary changes to docker-compose.local.yml/backend-unit and testframework/docker/backend_unit/runtime.sh/CGROUP_MEMORY_BYTES only; exactly3 scalar replacements.
- CPU6, no swap, read-only, pids256, GOFLAGS=-p=1, GOPROXY=off, unchanged GOGC/GOMEMLIMIT/default toolchain, tmpfs64/768/384MiB, cache watermark512MiB and report64MiB remain unchanged.
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

## Task1: Controlled resource attempt

- [x] Read prior actual OOM evidence; approve only the single memory-variable experiment. No repeated RED is necessary: three prior2GiB attempts already failed.
- [x] Independent read-only review of internal and observer memory contracts.
- [ ] Wait for an admissible resource window. Initial12:19:53UTC preflight: active Mail frontend-unit, free3625508KiB, disk0%/queue0; no active build. Attempt NOT_RUN and no temporary source edit.
- [ ] Apply exact3 scalar substitutions immediately before execution; inspect scoped diff/hash. Root self-check: only two named paths, no command or assertion weakening.
- [ ] Run only: docker compose -f docker-compose.local.yml run --rm backend-unit.
- [ ] Inspect exact container engine/image/start/cgroup limits/worker commands; capture Docker lifecycle events including OOM/die/destroy and bounded resource samples. Never print secrets or all env.
- [ ] Preserve unique partial/complete report directory without overwriting prior attempts. Read exact stage statuses, real Go failures and completed/skipped coverage. Nonzero infrastructure abort is not a source/test failure.
- [ ] Restore both files after process completion/owned abort; compare exact original hashes.
- [ ] Independent review of actual evidence; docs-sync and exact-path commit/immediate push. Record Go diagnostic result separately from canonical observer, product, API/media readiness.

## Independent review boundary

Terra/high reviewed the initial3-scalar proposal: in-container assertions and serializer follow the shared constant, but canonical observer rejects4GiB.
Exact consumers: testframework/observers/backend_unit/observe.py:122,615;
contracts.py:157,872; sealed testframework/policies/backend-unit-observer.v1.json:174.
lease.py:26 binds Compose and runtime.sh as current build inputs, so restoration also changes the canonical manifest.
Direct backend-unit-receipt packet validation is not authoritative Docker host-limit acceptance.
Root ruling: retain observer safeguards; the authorized one-off attempt is diagnostic only. No canonical GREEN or downstream acceptance follows from it.

## Root self-review

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
