---
task: PHASE-128-TASK-204
verdict: NOT_RUN
verified: false
verifier: pending-independent-docker-rerun
revision: null
verified_at: null
---

# Phase 128 test-rebase guard verification

## Current verdict

**NOT_RUN.** The current guard implementation has not received a fresh run of
the canonical Docker wrapper and has not been signed off by independent
`arch-ver` and `qa-auto-ver`. No PASS, verified-state transition, cutover
manifest, batch application, rollback execution, or Phase 128 completion is
claimed by this document.

The older TASK-202/TASK-203 observations are historical inputs only. They do
not authorize migration and are preserved separately in the sealed ledger and
baseline artifacts.

## Required acceptance evidence

The only supported runtime entrypoint is:

```powershell
C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File .planning/audits/test-rebase/validate-contract.ps1
```

A future PASS requires all of the following from one exact task run:

- immutable inspected image ID and image-lock digest;
- exact live Git source revision;
- CPU, RAM, swap, PID, network-none and read-only limits;
- inner timeout plus an independent bounded host-side watchdog whose sentinel
  exercises the outer kill path;
- held-handle-relative, atomically claimed, identity-pinned, owner-marked
  `testframework/reports/<task-run>/` evidence boundary with exclusive
  no-reparse/single-link artifacts and bounded pipe-captured process output;
- raw stdout/stderr hashes, duration and exact first exit code;
- exact run-script and host execution-argument contract digests;
- exact committed Git blob plus SHA-256 for every task-owned, read-only pinned
  executable snapshot; held-handle proof through consumption and a task-only
  ABA rename/replace sentinel;
- forced-timeout sentinel and daemon-backed proof by exact names and task-run
  label enumeration that every task-owned container is absent after cleanup;
  cleanup must capture/recheck immutable container IDs and ownership labels,
  remove only by ID, then scan both exact names and labels;
- deterministic builder dry-run equality and continued absence of the final
  repository `cutover-manifest.json`;
- independent validator plus atomic apply/rollback/builder race,
  crash-window, final-name-safe pending-journal publication,
  staged-directory/mkdir/cleanup receipt recovery,
  retry-convergence, source-collision and cleanup self-tests;
- temp-only inventory publication idempotence, appearing-target and
  post-fsync/pre-publish crash sentinels without live ledger generation, plus
  pinned-Git-tree/worktree cross-file mutation and HEAD-drift sentinels;
- independent verifier identities recorded only after inspecting the artifacts.

## Verified-state rule

Manifest JSON cannot self-declare a verified state. Every verified batch must
reference an external canonical receipt whose SHA-256 is checked from the
repository. That receipt binds the exact evidence object, operations digest,
applied/verifier commits, verifier identity/timestamp, and transition-chain
digest. Gate evidence also binds the inspected immutable image, exact command
contract, owner marker, raw output artifacts, reconciled counts, duration and
parsed cleanup proof. Aggregate evidence and cleanup artifacts are parsed and
matched to task-run, image, source, exit code and declared digests. Transition
digests must be continuous, and descriptor-safe Git object reads must prove
live source/applied/verifier commit existence and ancestry. Packed-only commit
objects currently fail closed instead of using an unsafe fallback.
The receipt graph is one-way: the batch-local verified projection excludes all
sibling batch state plus its own downstream receipt/self-consumer references,
the receipt binds that projection, and a two-batch constructibility sentinel
must prove later sibling transitions cannot stale an earlier receipt. Each
batch first transition must authenticate the exact immediate global predecessor
from the sole parent commit of its applied commit. The first bootstrap artifact
comes from the declared source commit and names a strict ancestor; later
artifacts must contain the completed verified prefix, current planned batch and
untouched planned suffix, with the prior verifier commit in their ancestry.
Global states must be monotone, and the two-batch sentinel must reject reuse of
the stale original all-planned snapshot. Adjacent within-batch transitions must
preserve exact after/before continuity.
Receipt, raw streams, image/owner locks and cleanup/evidence artifacts must
byte-match the authenticated verifier-commit tree as well as their live
descriptor reads and declared SHA-256 values.

Until those checks run successfully, this frontmatter must remain
`verdict: NOT_RUN`, `verified: false`, and `verified_at: null`.

## G128-S static guard acceptance - NOT_READY (2026-08-09)

**Classification:** static source-and-worktree review only. This is neither
`STATIC_REVIEW_PASS` nor runtime `PASS`; no wrapper, Python tool, validator,
Docker command, test, or batch apply was invoked.

The required receipt cannot truthfully bind a committed full
HEAD/tool/image/authority envelope in the inspected worktree. The wrapper has
a declared immutable image digest and strong committed-blob checks for its
six execution sources, but those facts do not make the current dirty and
untracked input set a committed authority envelope.

### Literal failed assertions

| Assertion | Result | Static evidence |
|---|---|---|
| `G128-S/committed-execution-sources` | **FAIL** | Targeted `git status --porcelain=v1 -uall` reports ` M .planning/audits/test-rebase/inventory.py`, `?? .planning/audits/test-rebase/apply_cutover_batch.py`, and `?? .planning/audits/test-rebase/validate-contract.ps1`. `Add-PinnedExecutionSnapshot` compares each source worktree blob with `HEAD:<path>` and throws `Pinned execution source is not the exact committed Git blob` on mismatch, so a receipt cannot be produced from these current bytes ([validate-contract.ps1](validate-contract.ps1):365-410, 618-623). |
| `G128-S/committed-authority-inputs` | **FAIL** | The same targeted status reports ` M .planning/audits/test-rebase/inventory.json` and `?? .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json`. The wrapper snapshots only the six execution sources above; it passes the founder manifest as a live `/workspace` input to the validator and does not compare either authority JSON input with `$SourceRevision:<path>` ([validate-contract.ps1](validate-contract.ps1):365-410, 673-676). |
| `G128-S/full-HEAD-receipt-binding` | **FAIL** | `$SourceRevision` is only `git rev-parse --verify HEAD`; the inner guard repeats that commit-id comparison. Neither assertion establishes a clean worktree or rejects untracked authority inputs, and the execution-source blob guard does not cover the live JSON authority inputs ([validate-contract.ps1](validate-contract.ps1):614-623, 670-676). |
| `G128-S/runtime-image-and-receipt` | **NOT_RUN** | The source declares an expected immutable image digest and would compare Docker inspection output at runtime, but no such inspection, evidence artifact, or receipt exists in this review ([validate-contract.ps1](validate-contract.ps1):29, 645-656, 745-750). |

### Remediation boundary

No second launcher, tool edit, authority-JSON edit, batch apply, or runtime
guard is authorised by G128-S. Before the single named Docker guard may be
considered, a separately authorised change and review must make every
authority input consumed by the guard provably equal to a blob in the exact
`$SourceRevision` tree (and fail closed on modified, missing, or untracked
paths), then establish a fully committed reviewed envelope. Only a subsequent
one-shot execution through `validate-contract.ps1`, under the sealed plan
gate, may create runtime evidence; direct host/Docker fallback remains
forbidden. Until then, retain `NOT_READY`/`NOT_RUN` and do not represent this
static review as a runtime acceptance.
This receipt does not authorize execution: only G128-D may make one attempt
after an accepted G128-S; that one-shot window closes after the attempt and the
historical wrapper is forbidden after G129-L.

## G128-D prelaunch incident - NOT_READY (2026-08-09; post-G128-S)

**Attempted host command:** `pwsh -NoProfile -File .planning/audits/test-rebase/validate-contract.ps1`.

**Observed boundary:** host-shell resolution raised `CommandNotFoundException` because `pwsh` was not recognized. The failure occurred before PowerShell could load the wrapper or process its arguments.

**Absent task effects:** the wrapper-created task reports directory was absent and `.planning/audits/test-rebase/cutover-manifest.json` was absent. No wrapper, Docker process/container/network/mount, validator, report, or runtime check started. Accordingly no raw G128-D receipt exists.

**Status:** G128-D remains `NOT_READY`; runtime is `NOT_RUN`, not `FAIL`. This incident does not alter frontmatter `verified: false` or the G128-S `STATIC_REVIEW_PASS` below.

**One-shot boundary:** this was not the named Docker-owned G128-D attempt, so the one-shot attempt remains unconsumed. There will be no retry or interpreter fallback under this incident record. The former PowerShell 7 prescription is superseded by accepted G128-WPS51: install nothing, bind the preinstalled signed Windows PowerShell 5.1 inbox component by its exact absolute path and host lock, then require renewed G128-E/G128-S acceptance before the one permitted Docker-owned attempt. Unqualified `powershell.exe`, `pwsh`, PATH/PATHEXT discovery and every fallback remain forbidden.

## G128-S superseding static receipt - STATIC_REVIEW_PASS (2026-08-09)

**Provenance clarification (fix round 1):** G128-S does not claim to have rerun any Git check or independently proved a commit-tree relation. Its use of `committed` for the G128-E envelope is a consumption of the accepted G128-E provenance receipt: atomic commit `4fc2d21459991bf0e66eebe79902d77bc5fa6082`; cached pre-commit audit scope of exactly five writable paths; commit output of five files; empty targeted post-commit status for those five paths; and `git ls-tree` listing all five. The independent G128-E architecture PASS further established the same-HEAD relation for six read-only dependencies and all eleven envelope hashes, while G128-E security PASS established the wrapper/tool boundary. This G128-S review independently matches only the present bytes of the eleven named files to the SHA-256 values below; that byte comparison alone is not represented as new Git provenance.

**Scope:** static review of the committed G128-E source envelope only.
**Source revision:** `4fc2d21459991bf0e66eebe79902d77bc5fa6082`.
**Runtime:** still `NOT_RUN`; this receipt does not change the frontmatter
`verdict: NOT_RUN`, `verified: false`, or `verified_at: null`.

This supersedes the preceding G128-S `NOT_READY` result only for the historical
pre-G128-E dirty/untracked envelope described there. That evidence is retained
unchanged. The current source bytes match the supplied committed G128-E seals;
this read-only review deliberately did not invoke Git to re-resolve the commit
tree.

### Sealed current envelope

| Path | SHA-256 |
|---|---|
| `.planning/audits/test-rebase/validate-contract.ps1` | `f1e5aaea97b38d0b32442b2d5256fea3e5e28b213df53002c67de1d5718a0ea4` |
| `.planning/audits/test-rebase/apply_cutover_batch.py` | `44004ec557af849974b551b4eeb185ddbc399d7440e26b6387b0183e6e793156` |
| `.planning/audits/test-rebase/founder-greenfield-cutover-manifest.json` | `df503a7f136aed70f0a9aaed2a576350f692574b0fd9f5506bd4b7f5fed0376a` |
| `.planning/audits/test-rebase/inventory.py` | `4b8a21b76acdf541c2c576f817b3df16d6c98b0b66caf6be1cca605fe5142361` |
| `.planning/audits/test-rebase/inventory.json` | `72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511` |
| `.planning/audits/test-rebase/inventory.schema.json` | `4ea14f9ad23cf3a0c98a21158802409a73224f86e1355c7647d392fdff65f4ce` |
| `.planning/audits/test-rebase/test-architecture-contract.json` | `62eb27ecd93164cf404639e6c8b85d14843613da2d34ea4da08101a673bf1335` |
| `.planning/audits/test-rebase/validate_test_architecture_contract.py` | `05f5515720074bb2aaf2d429d2658911263fc4f0fbf598cf0b54f0f09cc2035d` |
| `.planning/audits/test-rebase/build_cutover_manifest.py` | `84f00f095899b683784887fd8d1668a8b7edbb44d22b8fc21e406acd79da61f9` |
| `.planning/audits/test-rebase/validate_cutover_manifest.py` | `677331b75b8b2f032a8f77dfea854c0384796401602816c07e7c35da7a402610` |
| `.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md` | `6c0198ead0fc1afa5e696d552106bfa6e5d99e497a6987a7ece5d1219666659a` |

### Static boundary assertions

| Assertion | Result | Evidence |
|---|---|---|
| Fixed invocation and external-tool provenance | **PASS** | The wrapper has empty `param()` and no override path. It rejects relative executables, clears inherited child environments, pins no-reparse single-link absolute Git/Docker executables, verifies their SHA-256 values, and records their identities. Current read-only hashes match sealed Git `544548de8c9bed98eb90531729a2955f554e7009338788b68777258ea61867e4` and Docker `f6e3fb4ce2342f45c680656e9ae02a47c641efc10ba52eeb4e89f379f6befc6c` ([validate-contract.ps1](validate-contract.ps1):301-440, 727-756). |
| Committed 11-path source and authority closure | **PASS** | Every fixed allowlisted path is opened by a held no-reparse, single-link handle; its bytes are compared with `SourceRevision:<path>`, copied to a held snapshot, and rechecked through final cleanup. The wrapper snapshots six executable sources and five authority inputs, then rejects HEAD drift after pinning ([validate-contract.ps1](validate-contract.ps1):470-532, 766-786, 1103-1114). |
| Replacement and environment defense | **PASS** | Outer Git uses only the minimal environment with `GIT_NO_REPLACE_OBJECTS=1`; the inner script exports it, asserts both declarations, and runs a tmpfs-only replacement-ref sentinel before the live HEAD check or validators ([validate-contract.ps1](validate-contract.ps1):443-450, 727-741, 822-869). |
| Canonical read-only authority use | **PASS** | Each authority snapshot is overlay-mounted at exactly `/workspace/<canonical path>`, read-only, and compared to the pinned copy before validation. The repository bind is read-only; `/tmp` is the only writable main-container filesystem ([validate-contract.ps1](validate-contract.ps1):822-861, 899-915). |
| No cutover, publish, SUT, or unsealed write | **PASS** | The inner command vector is limited to inventory `--check`/`--self-test`, architecture validation, builder `--dry-run`/`--self-test`, independent validation, and applier `--self-test`. It asserts absence of the final `cutover-manifest.json` before and after the builder/applier checks; no `--apply`, `--generate`, publish, SUT, or direct host fallback is exposed ([validate-contract.ps1](validate-contract.ps1):847-861). |
| Limits, timeout, and exact cleanup | **PASS** | All guard containers have fixed image digest, no network, read-only root, dropped capabilities, CPU/RAM/PID limits, timeout sentinel, bounded host process handling, and ownership-checked exact-ID/label cleanup with final absence scans ([validate-contract.ps1](validate-contract.ps1):301-325, 589-701, 873-924, 958-1012). |

### One-shot authorisation and expiry

`STATIC_REVIEW_PASS` is not a runtime pass and does not reserve or consume an
attempt. Under Phase 128 plan tasks 128-S and 128-D, only one named,
Docker-owned G128-D historical-guard attempt may occur after this receipt. That
window closes after the attempt and the wrapper is forbidden after G129-L;
direct host/Docker fallback remains prohibited. The wrapper is an execution
boundary, not a scheduler, so any claim that the one attempt occurred remains
invalid until the separate G128-D receipt exists.

## 2026-08-13 — ротация леджер-авторитета и Docker-цепочка TASK-204

Этот раздел — датированная evidence-запись; вердикт фронтматтера выше
(NOT_RUN для канонического wrapper-прогона G128-D) он не меняет.

### Контекст и фиксы инструментария

Реконструкция репозитория (`b8677d9d`) создала параллельную git-линию:
recorded revision `5708671f` перестал быть предком HEAD, и fail-closed
`inventory.py --check` корректно уходил в RED по ancestry. Дополнительно
вскрыты и исправлены три никогда не гонявшиеся регрессии инструментария:

- `72589442` — fstat-снимок staging до записи payload в
  `inventory.py::write_inventory_atomically` (публикация и `--self-test`
  падали всегда);
- `b9f690da` — текст file-budget стража расходился с негативным сентинелом;
- `986b2211` — та же fstat-регрессия в
  `build_cutover_manifest.py` (destination-race сентинел падал wrong failure).

### Ротация авторитета (`da2c1c4c` + `31072faa`)

Доказано байт-в-байт: пиновые sha256 у 1027/1049 строк старого леджера были
хэшами CRLF-варианта тех же git-блобов (пример
`backend/testframework/accounting/fixtures_test.go`: LF-блоб `883e1aa1…`,
блоб+CRLF = старый пин `d8e92a58…`); блобы линий `5708671f` и `b9f690da`
байт-идентичны. Регенерация tool 2.2.1 в контейнере: 1049/1049 строк, у 1026
изменились только `sha256`+`evidence`, у 1 строки
(`testframework/stages/unit/frontend-test/domains/tms/import-multipart-api.test.ts`)
дополнительно `dependency_hints` (CRLF внутри сниппетов); у всех 43 записей
`hooks[]` ротирован `sha256`; классификация, партиции, hook-линки и групповые
`source_paths_sha256` не изменились. Пары дайджестов:

| Артефакт | Старый | Новый |
|---|---|---|
| inventory.json | `72de6a81b2bd…` @ `5708671f` | `83b6a3114980…` @ `b9f690da` |
| founder-greenfield-cutover-manifest.json | `df503a7f136a…` | `292b40dbddcc…` |
| test-architecture-contract.json | `62eb27ecd931…` | `8e9f4819f4f9…` |

Независимая верификация (3 агента, полная развёртка всех 1049 строк и 43
hooks против git-блобов): 0 расхождений, вердикты PASS_WITH_NOTES.
`scripts/reorganize-testframework-pipeline.ps1` (позиция 12/28 non-ledger
удалений) восстановлен байт-точно из старой линии (blob `0e967e22…`).

### Docker-цепочка (source revision `986b2211`)

Окружение: image `python:3.12`
(`sha256:dd4fe98ab39f91e936f8e7e7a65a3ce59ecfb11e32f9a125b3132779920ba7f7`),
контейнер `--cpus 2 --memory 2g --pids-limit 512 --network none`, workspace —
named volume (ext4), клон в нативную ФС, checkout `986b2211`. Полный лог:
`task204-docker-run-2026-08-13.log` (в этом каталоге). Результаты:

| Шаг | Результат | Время |
|---|---|---|
| inventory `--self-test` | GREEN | 0s |
| inventory `--check` | GREEN: rows=1049, LEDGER_HEAD=b9f690da | 4s |
| контракт-валидатор позитив | GREEN: move 507/1673, keep-block 245/1428, delete 297/1145, non-ledger 28 | 0s |
| контракт-валидатор `--self-test` (bind-sentinel) | GREEN (негативы RED) | 1s |
| builder `--self-test` | PASS (atomic no-replace race sentinel) | 0s |
| builder `--write` | PASS: sealed planned manifest | 1s |
| cutover-валидатор `--check` | PASS: batches=173 planned=173 | 1s |
| cutover-валидатор `--stdin --self-test` | PASS (негативная батарея RED) | 5s |
| ручной traversal-негатив | exit 1 (ожидаемо) | — |

Контейнер удалён (`--rm`); суммарно 25s. `cutover-manifest.json` перенесён в
этот каталог байт-точно: sha256
`85dd85ca38923a0211eff6110bef797274289c2ef453ec66364118100dff7bd5`,
173 батча в состоянии `planned`, `source.repository_head=986b2211`. По
конструкции валидатора live-HEAD-проверка all-planned манифеста привязана к
`986b2211`: PASS зафиксирован до коммита манифеста; после старта батчей
authority ведёт state-transition механика.

Остаются открытыми: канонический wrapper-прогон G128-D (founder-gated),
независимая подпись arch-ver/qa-auto-ver, honest post-cutover unit baseline.
