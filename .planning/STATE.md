---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: test-rebase-minimal-merge-candidate
stopped_at: final Git audit and remote fast-forward publication
last_updated: "2026-08-23T00:00:00.000+03:00"
last_activity: 2026-08-23 · minimal Docker gates PASS; hosted CI off; publication pending
progress:
    total_phases: 125
    tasks_completed: 125
    goals_verified: 63
    phases_percent: 100
    phases_percent_note: >-
      100% относится только к историческим 125 фазам milestone v4 и не означает
      production-ready. Активная программа test-rebase — фазы 128-141; Phase
      128 IN_PROGRESS, остальные PLANNED. Эти 14 фаз не входят в счётчик выше.
    notes: >-
      tasks_completed = checkbox/legacy completed_phases claim.
      goals_verified = count of classification==VERIFIED_PASS in
      PHASE-STATUS-MATRIX.md per-phase table (column classification only;
      method documented in matrix header «How to count goals_verified»).
      Recount 2026-07-28 (TASK-004): 63 VERIFIED_PASS / 5 VERIFIED_PARTIAL /
      2 CODE_GATE_DEPLOY_BLOCKED / 54 NO_VERIFICATION = 124 dirs.
      Prior STATE value 55 was wrong (naive full-file / notes drift).
      phases_percent remains tasks_completed/total_phases — NOT goals_verified.
      Matrix scan date 2026-07-27; VERIFICATION for 126/127 added 2026-07-28
      not yet folded into matrix (no silent re-write of phase verdicts).
deploy_blockers:
  - id: PG-03
    aka: FA-06
    title: Secret rotation (Resend/Telegram P1 + inventory)
    status: open
    blocks: production_cutover
    does_not_block: feature_dev_july_rebuild_planning
    phase_refs:
      - "63-production-gate (VERIFICATION: code PASS, deploy blocked)"
      - "67-production-gate (VERIFICATION: code PASS, deploy blocked)"
    docs:
      - .planning/ops/production/PG-03-SECRET-ROTATION.md
      - .planning/ops/production/FA-06-EXECUTION-LOG.md
      - docs/runbooks/fa06-secret-rotation.md
    progress:
      phase_0_git_scan: done
      phase_1_rotate: pending
      phase_2_smoke: pending
      phase_3_cto_signoff: pending
coverage:
    overall_internal_statements: 82.0373
    measured: "2026-08-07"
    status: historical_non_blocking_baseline
    canonical_baseline: TASK-203_discovery_only_founder_cutover_supersedes_gate
    baseline_2026_06_22: 83.0
    delta_pp: -0.9627
    gate_floor_82: historical_pass_not_migration_gate
inventory_rotated_2026_08_13:
    source_revision: b9f690da5cdcf322fc5b28a8d96c71a422407154
    ledger_sha256: 83b6a31149807d21441a2ede91dec0b36b4a4396c4c0ceb58308b58878f5ad89
    rows_identical_to_2026_08_07: true
    reason: reconstruction-cut-lineage-and-crlf-polluted-hashes
    cutover_manifest_sha256: 85dd85ca38923a0211eff6110bef797274289c2ef453ec66364118100dff7bd5
    cutover_manifest_batches: 173
inventory_measured_2026_08_07:
    source_revision: 5708671f00411c7c9e9dbee88b725de2092cbacd
    ledger_rows: 1049
    ledger_declared_tests: 4246
    ledger_sha256: 72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511
    backend_testframework_test_go: 831
    backend_testframework_declared_tests: 3672
    backend_other_test_go: 1
    backend_other_declared_tests: 2
    backend_total_test_go: 832
    backend_total_declared_tests: 3674
    backend_genuine_unit_files: 595
    backend_hook_free_unit_files: 350
    backend_hook_dependent_protected_unit_files: 245
    backend_component_files: 237
    backend_component_api_files: 31
    backend_component_integration_files: 206
    vitest_frontend_test_files: 180
    vitest_literal_cases: 519
    vitest_frontend_unit_files: 156
    vitest_static_contract_files: 24
    frontend_support_files: 3
    playwright_specs: 20
    playwright_declared_cases: 53
    playwright_support_files: 14
    hook_source_files: 43
    hook_symbols: 582
    hook_links: 892
    protected_unit_hook_files: 42
    protected_unit_hook_symbols: 434
    protected_unit_hook_links: 653
    target_python_tests: 0
    ent_entity_schemas: 87
    zustand_stores: 12
    go_mod: "1.25.0"
    branch: "cursor-grok-4.5"
active_program:
    name: test-architecture-rebase
    phases: "128-141"
    current_phase: 129
    completed_tasks:
      - TASK-202
      - TASK-221
      - TASK-222
      - TASK-223
      - TASK-226
      - TASK-225
      - TASK-203
      - TASK-224
      - TASK-204
    archived_tasks:
      - TASK-202
    in_progress:
      - TASK-197
    in_review: []
    blocked: []
    target_layout:
      - backend/tests/<domain>/**
      - frontend/tests/<domain>/**
      - frontend/tests/support/**
      - testframework/** (single root Python project; runtime test code is Python-only)
spec_coverage:
    fr_total: 28
    fr_with_phase: 9
    ratio: "9/28"
    decimal: 0.321
    percent: 32.1
    measured: "2026-07-27"
    source: .planning/audits/planning-audit-2026-07/SPEC-COVERAGE.md
---

## Current test-rebase override — 2026-08-23

This section supersedes older recovery/pre-merge blocker notes below for the
minimal founder-approved merge slice. Fresh sequential Docker evidence:
`backend-unit` PASS (build/vet/unit/race), `frontend-unit` PASS (TypeScript
plus Vitest 338/338 suites and 419/419 tests), and `python-contract` PASS
(181/181 framework selftests). The five SUT contract tests and their five
case records were explicitly removed; Python API, integration, e2e and load
are `NOT_RUN` and are not merge gates for this slice. Hosted CI/PR automation
is disabled and GitLab pipelines are off. Production deploy remains manual
and separately gated by FA-06/PG-03 plus CTO sign-off. The 14-path target-unit curation and 241 hook-dependent deletion batches have
accepted receipts. Full legacy-root reconciliation is 349 relocations + 232
net component deletions + 241 hook-dependent deletions = 822 removed source
paths; protected Mail/Calendar paths are outside this merge. This is framework/unit evidence, not product readiness.


# Состояние проекта Hub

## Historical recovery truth — 2026-08-14 (superseded by the 2026-08-23 override above)

The status table in this historical section is retained only as recovery
provenance and must not override the current minimal-merge status above.

Founder selected **Option A**: preserve commits `794ae3cc` and `7b5d5501` and
reconcile forward only. This is a recovery record, not retroactive acceptance.

| Field | Current status |
|---|---|
| `STATIC_PATH_MAPPING` | `PASS` |
| `CUTOVER_ACCEPTANCE` | `FAIL/BLOCKED` |
| `PHASE_129` | `RECOVERY_BLOCKED` |
| `CUTOVER_MANIFEST` | `PLANNED_UNCHANGED` |
| `BATCH_RECEIPTS` / `INDEPENDENT_BATCH_VERDICTS` | `ABSENT` / `ABSENT` |
| `129-02_BULK_MOVE` / `129-03_BULK_MOVE` | `OBSERVED_RECOVERY_UNACCEPTED` / `OBSERVED_RECOVERY_UNACCEPTED` |
| `129-04_DELETE_SET` | `STAGED_UNCOMMITTED_UNACCEPTED` |
| `CI_ROOT_DISCOVERY` | `UNBOUND` |
| `LAUNCHER` / `POST_CUTOVER_RUNTIME` | `NOT_READY` / `NOT_RUN` |
| `RECOVERY_VERIFICATION` | `RECOVERY_VERIFIED_NOT_BATCH_ACCEPTED` |

The immutable v1 `cutover-manifest.json` remains byte-identical with all 173
batches `planned`; it has no receipts or verifier evidence. The index contains
241 staged legacy deletions. Staged is not deleted or accepted. No Phase 129
completion or post-cutover runtime evidence is claimed. Exact observations and
successor gates are in `audits/test-rebase/RECOVERY-OPTION-A-2026-08-14.md`.

The independent re-review accepted the recovery observation packet after W-01
was fixed and found no `HIGH` or `WARNING` finding. It accepted no v1 batch,
bulk move, staged deletion, CI binding, launcher, runtime evidence, or Phase
129 completion; `CUTOVER_ACCEPTANCE=FAIL/BLOCKED` and
`PHASE_129=RECOVERY_BLOCKED` remain unchanged.

## Session note — 2026-08-13 (ротация леджер-авторитета + sealed cutover-manifest)

- Найдены и исправлены три латентные регрессии QA-инструментария, ни разу не
  гонявшегося в Docker (VERIFICATION.md фиксировал отсутствие прогонов):
  fstat-снимок staging до записи payload в `inventory.py` (`72589442`) и в
  `build_cutover_manifest.py` (`986b2211`) — атомарная публикация и self-test
  падали всегда; текст file-budget стража vs сентинел (`b9f690da`).
- Ротация леджер-авторитета (`da2c1c4c` + `31072faa`): реконструкция
  `b8677d9d` отрезала линию revision `5708671f` (не предок HEAD); пиновые
  sha256 1027/1049 строк были CRLF-загрязнены средой исходной генерации —
  доказано байт-в-байт (LF-блоб + CRLF = старый пин). Регенерация tool 2.2.1
  на `b9f690da`: 1049/1049 строк, партиции/классификация/hook-линки
  идентичны, групповые source_paths_sha256 не менялись. Пины обновлены во
  всей цепочке (inventory `83b6a311`, founder `292b40db`, contract
  `8e9f4819`), включая ADR-маркеры и must_haves PLAN 129; восстановлен
  `scripts/reorganize-testframework-pipeline.ps1` (12/28 non-ledger).
  Независимая верификация: 3 агента, полная развёртка 1049 строк + 43 hooks
  против git-блобов, 0 расхождений, PASS_WITH_NOTES.
- TASK-204: полная Docker-цепочка GREEN на `986b2211` (python:3.12 digest
  `dd4fe98a…`, 2 CPU/2GB, network none, ext4 volume, 25s): inventory
  self-test/check, контракт-валидатор ±, builder self-test/--write,
  cutover-валидатор --check/--stdin --self-test, ручной traversal-негатив.
  `cutover-manifest.json` запечатан и закоммичен (`f00292ea`): 173 батча /
  1120 операций (move 507, delete 325, keep-blocked 288), все planned,
  sha256 `85dd85ca`, `source.repository_head=986b2211`. Evidence: лог
  `task204-docker-run-2026-08-13.log` + датированная секция VERIFICATION.md.
- Открыто по Phase 128: независимая подпись arch-ver/qa-auto-ver,
  founder-gated G128-D wrapper-прогон, honest post-cutover unit baseline.
  Ветка `claude/test-rebase-20260813`; в master ничего не пушилось.

## Session note — 2026-08-08 (test-rebase execution readiness)

- Research/GSD/Superpowers synthesis is now executable Hub policy, not a loose
  report: atomic commits `a9077796`, `b1b69124`, `724947f7`, `8220370e` add
  bounded TaskContract/context/evidence/workflow-eval contracts, conclusion-blind
  verification, separate 3-fix/5-review breakers and fail-closed Docker runner
  requirements. Independent prompt-contract re-review = **PASS**; independent
  test-contract re-review = **PASS** after `77f94e89` removed the last four
  direct Docker launcher bypasses.
- Touchstone `pro-qa-ai` is current `0.4.0-beta.3` at commit `30f75a25`, with
  `STRUCTURAL_PASS`; the Desktop portable copy is also beta.3. Paired
  effectiveness/ROI evaluation remains `NOT_RUN`, so this is not GA.
- TASK-202 имеет статус **archived/verified**; live evidence — 1,049 rows /
  4,246 declared, ledger SHA `72de6a81…` на source revision `5708671f`.
- Phase 129 разбита на семь ordered executable plans. Каждый move/delete batch
  обязан в одном atomic commit выполнить filesystem operation, переключить все
  affected CI/scripts/docs, обновить `cutover-manifest.json` и приложить
  Docker-only evidence; только independent verifier ставит `verified`.
- Synthetic task aliases удалены из execution contract: deletion и финальная
  zero-consumer hook closure принадлежат реальной TASK-219; TASK-201 разделена
  на per-batch switch и final cleanup.
- Phase 136 требует 131; Phases 134/135 формально требуют 133. Все 245
  hook-dependent genuine Go units остаются protected до TASK-234.
- Product/test files не перемещались и не удалялись; validator/inventory Python
  и `pro-qa-ai/**` не менялись.
- Touchstone `0.4.0-beta.3` остаётся foundation-ready pre-release; paired eval
  `NOT_RUN`.

## Session note — 2026-08-07 (test architecture rebase + Phase 128)

- Текущая product-truth (founder, 2026-08-12): Hub остаётся незавершённым MVP;
  ни один продуктовый модуль не принят как complete или stable. Исторические
  milestone COMPLETE не равны production acceptance; PG-03/FA-06 —
  зафиксированный ops-блокер production cutover, не индикатор готовности MVP.
- Test-план полностью rebased: 14 фаз 128–141, target
  `backend/tests/**` + `frontend/tests/**` + единый корневой Python-проект
  `testframework/**` без вложенного `testframework/python/`. Канон:
  `active/august-2026/06-TEST-ARCHITECTURE-REBASE.md`.
- TASK-202 пересчитан на текущем pinned source revision `5708671f`: 1,049 rows
  / 4,246 declared; Go total 832/3,674 (legacy root 831/3,672 + server
  1/2; 595 genuine unit + 237 component: 31 API + 206 integration), Vitest 180/519
  (156 unit + 24 static contract) + 3 support, Playwright 20/53 + 14 support,
  hook sources 43 files / 582 symbols / 892 links, target Python 0.
  Первый qa-auto-ver verdict = FAIL по четырём schema/entrypoint gaps; после
  repair повторная независимая проверка = **PASS, verified true**. Current
  ledger SHA: `72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511`.
- Docker Desktop 4.40 / Engine 28.0.4 доступен после approved daemon access.
  Начальная попытка TASK-203 была красной и сохранена как historical triage
  evidence; частичные/red profiles не принимались.
- TASK-221…225 закрыли path guards, GitLab/NATS/cron nondeterminism, frontend
  regressions, Biome debt и coverage gap отдельными проверяемыми scopes.
  Финальный backend suite green: 50 packages PASS / 0 FAIL / 3 no-test-files;
  exact merged profile **82.0373%** (21,269 / 25,926), security floors **6/6**,
  audit **85.6879%**. Порог и denominator не менялись.
- Frontend Docker observation: TypeScript PASS, Biome 370 files / 0
  diagnostics; frozen legacy Vitest **510/519 PASS, 9 FAIL** (6 absent docs
  mount, 2 nested fixture timeout, 1 stale i18n assertion). Founder cancelled
  legacy remediation; this is discovery evidence, not a cutover gate.
- TASK-224 закрыт реальным изолированным Docker baseline: smoke **3/3 PASS**;
  full без retries **48 PASS / 5 expected skip / 0 fail**. Два найденных
  production gaps (`GET /api/admin/users`, стабильный Telegram API без bot
  token) исправлены отдельными атомарными коммитами; Messenger locator flake
  устранён и full повторён без retries.
- Founder superseding decision: только genuine isolated units сохраняются и
  перемещаются в `backend/tests/**` / `frontend/tests/**`. Все legacy Go
  component/API/integration, static-contract Vitest, TS Playwright и их
  support/config/entrypoints frozen и удаляются по exact manifest, без port и
  row-parity. Historical 82% Go coverage не является migration blocker.
- Весь non-unit QA создаётся с нуля на Python прямо в `testframework/**` от
  task/FR/NFR/verified TC и risk matrix: happy, negative, authn/authz,
  boundaries, edge/corner, failure/recovery, security/a11y/performance где
  применимо. Docker-only и resource limits обязательны.
- Task-owned `hub-e2e224-*` containers, network, volume и temporary frontend
  image удалены после evidence; shared `deployments-*` оставлены работающими.
- По явному founder authorization выполнен read-only policy extraction из
  `pro-qa-ai/**`; source не менялся. Hub получил
  `.cursor/rules/07-agent-safety.mdc`, усиленные delegation/handoff и запрет
  push/PR/deploy без явного поручения. Локальные commits — атомарно allowlist-ом.

## Session note — 2026-08-07 (team-lead onboarding + model routing)

- Team-lead заонборжен по `AGENTS.md`, live `STATE`, `.planning/README.md`,
  августовскому плану и `.cursor` regulation/playbooks. Канон следующей работы
  не изменён: resume с TASK-193; до старта выбрать единый coverage baseline.
- Добавлен `.cursor/playbooks/model-routing.md`: role defaults всех 32 агентов,
  effort/escalation/fallback и cost snapshot. `team-lead` = Sol/Extra High;
  Terra = основной engineering runtime; Luna = механика; Spark = optional preview.
- Downstream frontmatter сохранён `inherit` для portable запуска; фактический
  runtime/reason/fallback/catalog status обязателен в delegation/handoff.
- `pro-qa-ai/**` закреплён как protected external zone Claude Code Opus 5 под
  управлением founder. Внутренние Hub agents не читают её по умолчанию, не
  меняют и не включают в волны; только явный founder handoff конкретного файла.
- Синхронизированы ADR, AGENTS, REGULATION, team workflow, agent descriptions и
  ownership rules. Product code/tests и `pro-qa-ai/**` не менялись.

## Session note — 2026-07-30 (Touchstone Standard + Pro packaging)

- **Touchstone Standard** (`pro-qa-ai/`): продукт-идентичность `TOUCHSTONE.md`,
  версия **`0.2.0-rc.1`** foundation-ready, CHANGELOG, обновлены README/START-HERE/INSTALL/MANIFEST/VERSION.
  Structural PASS сохранён; effectiveness/ROI по-прежнему **NOT_RUN** (честный GA-gate).
  (prompt-contract, master-delegate, wave-protocol, mutation-evidence, backlog-task);
  источники: GSD inventory, PE Standard, `.cursor/` (contract/team/verify), `.planning/` (backlog/STATE/mutation).
- Синхрон: `.planning/README.md`, `.cursor/rules/70-pro-qa-ai.mdc`,
  `REGULATION.md` §20, `playbooks/README.md`, `AGENTS.md`.
- Код продукта / validator scripts не менялись. Коммитов нет (ожидание запроса пользователя).

## Session note — 2026-07-30 (qa-auto1 · TASK-008 hubtest close)

- Full hubtest **PASS**: `go test -tags hubtest -count=1 -timeout=600s ./testframework/...` → **50** ok · **0** FAIL · ~62.2 s. Evidence: [`HUBTEST-2026-07-30-task008.md`](audits/total-2026-07/HUBTEST-2026-07-30-task008.md).
- Prior FAIL (project FA11 second-member mock · storage UUID `entity_id`) — ROOT=test, fixed; packages green.
- [`00-FINAL.md`](audits/total-2026-07/00-FINAL.md) → вердикт **A**; TASK-008 / README → **Закрыта**; Close control → **DONE**.
- Residual: overall COV 80.2% vs 81.9% baseline (не gate 1–4). Коммитов нет.

## Session note — 2026-07-30 (docs-writer · TASK-008 W20 FINAL draft)

- Собран draft [`.planning/audits/total-2026-07/00-FINAL.md`](audits/total-2026-07/00-FINAL.md): вердикт **B** (superseded — hubtest re-run PASS → **A**, см. note выше).
- Gates на тот момент: README **0** «На проверке» ✅ · Stage C drain ✅ · messenger/rbac ✅ · hubtest **FAIL** (project FA11 + storage entity_id).
- Коммитов нет (инструкция слота).

## Session note — 2026-07-30 (qa-auto1 · TASK-184 C-W19-TT-001 last-member · W13 slot 4)

- **FIXED** на HEAD: `RemoveTeamMember` — `TeamMemberCount≤1` → 409 Conflict (`cannot remove the last member of a team`). Prior be1 «FIXED»/На проверке не было на HEAD (W13 GAP NOT_PROVEN).
- **Тесты GREEN:** `TestRemoveTeamMember_RejectsLastMember_TASK184`, `TestRemoveTeamMember_AllowsWhenMultiple_TASK184`; wave122 → Conflict; wave124 multi OK.
- **Mutation PROVEN:** убрать guard → оба TASK184 RED (`err==nil`); restore → GREEN. Report: `MUTATION-2026-07-30-W13-qa1-task184-fix.md`. Leftover MUTATED: нет.
- Docs: `docs/TESTING.md` TC-PROJ-11/12; API.md уже 409. Коммит не создавался.
- Статус: **Закрыта**.

## Session note — 2026-07-29 (be1 · TASK-184 C-W19-TT-001 last-member)

- be1 session claimed FIXED / На проверке — **не на HEAD** (подтверждено W13 qa-auto1 GAP 2026-07-30). Реализация + mutate — qa-auto1 refill ниже.

## Session note — 2026-07-29 (be1 · TASK-191 EA-001 AutomationRule soft-delete)

- **FIXED** EA-001: `AutomationRule.deleted_at` + indexes; `DeleteAutomationRule` → `SetDeletedAt` (не hard `Delete`); list/count/find/update/ListEnabled → `DeletedAtIsNil`.
- **Migrate:** `backend/migrations/20260729_task191_automationrule_soft_delete.sql` marker `needs-owner-apply` — **BLOCKED waiting owner** (SQL в репо, не применён на prod).
- **Тесты:** `automation_soft_delete_task191_test.go` + `schema_automationrule_task191_test.go`; source-guard `.Delete().` banned; mutation hard Delete → RED; restore → GREEN. `./testframework/{tracker,ent}/ -run TASK191` PASS; automation regression PASS.
- Не трогали: SEC-02 / VersionEQ webhook (TASK-152), RemoveTeamMember (184), calendar (189). TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-189 BE-V4 calendar trash + draft poll)

- **FIXED** BE-V4-001/002: `ListApprovedLeaveInRange` / `ListSprintsInRange` → `DeletedAtIsNil`.
- **FIXED** BE-V4-003: `GetNewsPoll` / `VoteNewsPoll` → `requirePublishedNewsOrAuthorOrAdmin` (draft → 404 non-author).
- **Тесты:** `leave_sprint_trash_task189_test.go`, `draft_poll_idor_task189_test.go`; mutation RED×3 → restore GREEN; `./testframework/{calendar,intranet}/` hubtest+race PASS.
- Docs: `docs/wiki/API.md`, `docs/system_specification.md`. Не трогали coffee/partners (163), preview (180). TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.
## Session note — 2026-07-29 (be2 · TASK-188 ASN-001/002 PushSender + Telegram RL)

- **FIXED** ASN-001: `PushSender` → `PushSubscriptionStore` (CRUD в `Repository`); `push.go` без `*ent.Client` / ent import; wiring `NewPushSender(notifRepo, …)`.
- **FIXED** ASN-002: `ConfirmLinkCode` Redis fails `tg:linkcode:fails:<user>` max 5 / TTL 5m (как HI-04); HTTP → 429 `ErrTelegramConfirmLimited`.
- **Тесты:** `push_layering_asn001_test.go`, `telegram_linkcode_rl_asn002_test.go`; `./testframework/notification/...` PASS; mutation (убрать fails/record) → RED. NF-07 / Magic Link rate не трогали. TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (fe1 · TASK-192 FE settings/AI version OL)

- **FIXED** FE-SMA-001: webhook toggle/delete/deliveries — `version` в DTO + PUT; toast на fail.
- **FIXED** FE-SMA-002: AI privacy + ai-store — `version` в load/save; 409 → toast + refresh.
- **FIXED** FE-SMA-003 (optional): marketplace «Мои» — `mineLoading` / `listLoading` ≠ empty.
- TASK-179: `AVAILABLE_EVENTS` без `sprint.completed` — без регрессии.
- Тесты: webhooks-version-ol · ai-privacy-version-ol · marketplace-mine-loading · TASK-179 · ai-store; mutation omit version → RED, restore GREEN.
- Статус TASK-192 → **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 / README не трогали.

## Session note — 2026-07-29 (docs-writer · TASK-183 C-W19-DOC-001 inventory/AGENTS/spec)

- **FIXED** C-W19-DOC-001: shell recount HEAD `d868ba77` (`cursor-grok-4.5`) — BE `*_test.go` **812** / Vitest **179** (61 `.ts` + 118 `.tsx`) / hooks **39**; дата measure **2026-07-29**.
- Синхрон: `AGENTS.md`, `docs/{README,test_plan,TESTING,testframework_strategy,TROUBLESHOOTING,system_specification,user_needs_analysis_2026,vision_and_scope,competitive_audit}.md`.
- Anti-stale: убраны live `515` / `99/99` / Vitest `94` / HEAD `bbdaf16`; US-M06 webcam → shipped (не Phase 16 IN).
- Статус TASK-183 → **На проверке**. TASK-008 не закрывали.

## Session note — 2026-07-29 (fe2 · TASK-182 HR pending leave error≠empty)

- **FIXED** C-W19-FE-005: admin `listPendingLeave` fail → секция + QueryErrorAlert/retry (не silent empty); empty queue только success+len0.
- Тест: `domains/hr/leave-pending-error.test.tsx` 2/2 PASS · mutation (убрать `pendingError` из gate) → RED · restore GREEN.
- Статус TASK-182 → **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 не закрывать.

## Session note — 2026-07-29 (be2 · TASK-185 C-W19-TT-002 Allure ExistingRun)

- **FIXED** C-W19-TT-002 path A: ExistingRun Allure — fail-closed на Mongo `StoreAllureResults` и match infra err (≠ NotFound); порядок match→Mongo→PG; create-path Mongo best-effort без изменений.
- **Тесты:** `testframework/tms/allure_existingrun_partial_c_w19_test.go`; mutation slog-only+continue → RED; `./testframework/tms/` hubtest+race OK.
- **Docs:** `docs/wiki/API.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer. TASK-008 не закрывали.

## Session note — 2026-07-29 (be1 · TASK-177 C-W19 deactivate/ban WS kill)

- **2/2 FIXED** C-W19-001 / C-W19-003: Deactivate/Ban → `SessionService.RevokeAllUserSessions` (Redis Del + NATS `revoked`) → Bridge рвёт живой WS; kill fail → fail-closed admin error.
- Wiring: `wireAuth` → `auditSvc.SetSessionRevoker(authSvc.Sessions)`.
- Тесты: audit mock revoker + fail-closed; auth NATS/Redis; ws live NATS e2e. Mutation: remove publish → RED → restore GREEN.
- Commits: `88494530` fix · `460fa131` test · `a09bdc0e` task file.
- Docs: `docs/security_guidelines.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer. TASK-008 не закрывать.

## Session note — 2026-07-29 (fe1 · TASK-187 FE-TA pipeline sanitize + P3)

- **FIXED** FE-TA-001: `pipeline-status` — `isSafeUrl`/`sanitizeUrl` на `html_url` (unsafe → без `<a href>`).
- **FIXED** FE-TA-002: `issue-attachments` — whitelist URL; unsafe → disabled chip без href.
- **FIXED** FE-TA-003: planning-poker — `loadFailed` → `QueryErrorAlert`+retry; empty → «не найдена».
- **FIXED** FE-TA-004: automation secondary statuses/channels — `isError` + compact alert (не пустой Select).
- FE-TA-005 (token cluster) — out of scope (hygiene).
- Vitest tracker: 21 PASS (sanitize + error≠empty + source-guard `href={run.html_url}` banned). tsc OK.
- Статус: **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 не закрывали.

## Session note — 2026-07-29 (fe2 · TASK-186 calls error≠empty + a11y)

- **6/6 FIXED** (FE-MC-001…006): channels fail → alert/retry (не пустая сетка); projects/channels dialog fail → alert (empty только success+len0); popover × / channel select / remove participant / week slot a11y names.
- **Тесты:** vitest calls `error-empty` + `a11y` + recordings-noopener **9/9 PASS**; mutation `channelsError=false` → RED.
- **Статус:** На проверке → fe-ver → code-reviewer. README не трогали (scope). TASK-008 не закрывали.

## Session note — 2026-07-29 (be2 · TASK-181 C-W19-004 webhook soft-delete deactivate)

- **FIXED** C-W19-004 / C-W19-FE-004: `SoftDeleteWithCrossLinkCleanup` в той же tx `SetIsActive(false)` на webhook endpoints; `ListWebhookEndpointsByProject` + `project.DeletedAtIsNil()` (trash → пустой list).
- ListActive `HasProjectWith(DeletedAtIsNil)` не ослабляли (TASK-152 lock).
- Тесты: `testframework/apigateway/task181_c_w19_004_test.go` (runtime + source mutation).
- Docs: `docs/wiki/API.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-169 A3W8-3 HubTable/JournalEntry Cascade)

- **2/2 FIXED** A3W8-3-001/002: `entsql.OnDelete(Cascade)` на HubTable→columns/records и JournalEntry→lines.
- Migrate SQL (needs-owner-apply): `backend/migrations/20260729_task169_hubtable_journalentry_cascade.sql`.
- Тесты: `testframework/accounting/cascade_hubtable_journalentry_task169_test.go` (hard CASCADE + soft unchanged).
- `go test -tags hubtest ./testframework/accounting/ -run TASK169` PASS. Handoff: be-ver → code-reviewer. Статус: На проверке.

## Session note — 2026-07-29 (be1 · TASK-144 A3W3-2 messenger services)

- **Сделано:** 6/6 A3W3-2 FIXED — Join IsMember fail-closed, Create extra-members rollback, bookmark channel-scope, star/mute NotFound→403, ClosePoll membership до close, AddMembers/DeleteChannel AppError wrap; hubtest messenger `-race` PASS + 6/6 mutation RED.
- **Коммит:** `18b3d72c` (+ docs evidence); статус **На проверке**; README не трогали.
- **Handoff:** be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-171 A3W8-5 Team/WorkflowStatus soft + Cascades)

- **7/7 FIXED** (A3W8-5-001…007): Team/WorkflowStatus `deleted_at`+`updated_at`; soft `DeleteWorkflowStatus` + 409 при issues; OnDelete Cascade members/versions/results/deliveries/from_transitions(+to_status).
- **Hubtest:** `./testframework/ent/ -run TASK171` PASS (9); `./testframework/tracker/ -run TASK171|IdempotentSoft` PASS (3); project smoke PASS.
- **Migrate:** `20260729_task171_soft_delete_ondelete_cascade.sql` — `needs-owner-apply`.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner).

## Session note — 2026-07-29 (be1 · TASK-155 A3W5-3 ai service)

- **9/9 FIXED** (A3W5-3-001…009): UTF-8 composer trim; usage/settings upsert; issue_lookup fail-closed; consent Default(false)+Create deny; VersionEQ 409; WS sanitize; BotUserRepository; bot-flow asserts.
- **Hubtest:** `./testframework/ai/` `-race` PASS (~19s). Mutation 001 observed red→green.
- **Approval:** Atlas ALTER DEFAULT `allow_external_ai` — waiting owner (schema+Create already deny).
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали.

## Session note — 2026-07-29 (be1 · TASK-170 A3W8-4 sprint schemas)

- **2/2 FIXED** (A3W8-4-001/002): partial unique active sprint (`IndexWhere` + migrate SQL); Sprint→issues `OnDelete(SetNull)`.
- **Hubtest:** `./testframework/ent/ -run TASK170` PASS (5); `./testframework/sprint/ -run SoftDelete|SecondActive|StartSprintAtomic` PASS.
- **Migrate:** `20260729_task170_sprint_active_unique_setnull.sql` — `needs-owner-apply` + pre-migrate cleanup duplicates.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner).

## Session note — 2026-07-29 (be1 · TASK-134 A3W1-2 tracker services)

- **6/6 FIXED** (A3W1-2-001…006): UpdateIssue `CountLabelsInProject`; CreateLabel `FindLabelGroupInProject`; Create/UpdateRule `ensureActionChannelInProject`; ListThreadIssueLinks `WithIssue` eager; CreateIssueLink AppError wrap; UpdateComment 404≠409.
- **Hubtest (re-verify после crash):** focused A3W1_2_* + P1BE03/A3_011 PASS (2.1s); full `./testframework/tracker/` PASS (6.5s, `-race`). Mutation smoke A3W1-2-001: без guard → FAIL, guard восстановлен.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner). Чужой WIP (workflow soft-delete / project repo) не коммитили.

## Session note — 2026-07-29 (be1 · TASK-137 A3W1-5 project residual)

- **6/6 FIXED** (A3W1-A5-001…006): Archive/Unarchive на trash → 404 без ClearDeletedAt; PermanentDelete только deleted+system admin; CreateTeamWithCreator Ent-tx; JoinByInvite MemberCount fail-closed; template List/Export без raw `err.Error()`; import >25MB → 413.
- **Hubtest (isolated worktree):** `go test -tags hubtest -race -count=1 ./testframework/project/` PASS.
- **Docs:** `docs/wiki/API.md` — archive/unarchive/restore/permanent semantics. Статус: **На проверке** → be-ver → code-reviewer. README не трогали (owner: только SCOPE).

## Session note — 2026-07-29 (be2 · TASK-164 A3W7-3 integrations)

- **3/3 FIXED** (A3W7-3-001…003): `ListSprintIssues` + GH/GL `FindIssue*` → `DeletedAtIsNil`; `/link-preview` issue path → `IsProjectMember` + per-user cache + Fiber v3 handler/auth order.
- **Hubtest (isolated worktree):** `./testframework/imex/` `./testframework/github/` `./testframework/gitlab/` PASS; mutation 001/002/003 FAIL as expected.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner: только SCOPE).

## Session note — 2026-07-29 (be1 · TASK-155 A3W5-3 ai service)

- **9/9 FIXED** (A3W5-3-001…009): UTF-8 composer trim; IncrementUsage/settings upsert+retry; issue_lookup fail-closed; preference consent false; VersionEQ+409; WS `AI stream failed`; BotUserRepository; flow asserts (+ mention/consent fixtures).
- **Hubtest:** `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/ai/` → **PASS** (~19.6s).
- **Docs:** `docs/wiki/API.md` (preferences/providers version + consent default + stream sanitize). Статус: **На проверке** → be-ver → code-reviewer. README не трогали.

## Session note — 2026-07-29 (be2 · TASK-147 A3W3-A5 calls)

- **4/4 FIXED** (A3W3-A5-001…004): orphan LiveKit room на DB-fail CreateCall/Instant; FindConflictingCalls + soft-delete channel; nil project_id → 400; title≤255 / guest_name≤100 runes → 400.
- **Тесты:** `go test -tags hubtest -race -count=1 -timeout=300s ./testframework/calls/` PASS (в т.ч. `a3w3_a5_task147_test.go`).
- **Статус:** На проверке → be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-165 A3W7-4 storage/dashboard/backup residual)

- **5/5 FIXED** (A3W7-4-001…005): ProjectStats/Admin counts fail-closed; backup S3 delete Warn+fail-closed; GetLayout membership; SoftDelete→404; burndown AggregateProjectCompletedByDay.
- Hubtest: dashboard+storage PASS (частичный); полный backup+регрессия — re-run be-ver при стабильном `ent/`.
- Docs: `docs/wiki/API.md` (layout 403 / project+admin fail-closed). Статус задачи: **На проверке**.

## Session note — 2026-07-29 (fe1 · TASK-124 A3W10-1 routes/modules error≠empty)

- **Сделано:** secondary accounting / TemplateSelector / marketplace pending / bases detail+table — `QueryErrorAlert`; recordings download `noopener,noreferrer`. 5/5 FIXED → **На проверке**.
- **Тесты:** vitest v4 + `recordings-download-noopener` **31/31** PASS; `tsc --noEmit` PASS.
- **Docs:** backlog TASK-124 Progress; README не трогали.

## Session note — 2026-07-29 (be1 · TASK-135 A3W1-3 tracker repos residual)

- **3/3 FIXED:** A3W1-A3-001 FilterSubjects+known-no-op Ack; A3W1-A3-002 executeRule Post/Notify → error → Nak; A3W1-A3-003 StartTimer `DeletedAtIsNil`.
- hubtest tracker: focused ok 4.178s · full `./testframework/tracker/` ok 12.654s (-race).
- Статус: **На проверке** → be-ver → code-reviewer.

## Session note — 2026-07-29 (fe1 · TASK-128 A3W10-5 TMS canMutate + auth routes)

- **Сделано:** `canMutateProject` на case detail (Edit/link/unlink) и runs list/detail (`RunHeader`); `/onboarding` + `/login` `beforeLoad`; avatar `revokeObjectURL`. 5/5 FIXED → **На проверке**.
- **Тесты:** vitest tms+auth (guards/revoke/can-mutate) **36/36** PASS; `tsc --noEmit` PASS.
- **Docs:** QUICKSTART — redirect guards login/onboarding; README backlog не трогали.

## Session note — 2026-07-29 (be2 · TASK-152 A3W4-5 apigateway)

- **Сделано:** ListActive + `project.DeletedAtIsNil`; Update webhook VersionEQ/409; trim `sprint.completed` allowlist + `WebhookFilterSubjects`; CreateAPIKey name MaxLen→400; docs `wiki/API.md`; статус задачи **На проверке**.
- **Тесты:** `testframework/apigateway/task152_a3w4_5_test.go` (+ правки существующих).
- **Риск:** FE `webhooks.tsx` ещё показывает `sprint.completed` / PUT без `version` (вне SCOPE).

## Session note — 2026-07-29 (be1 · TASK-161 A3W6-5 audit tests)

- **Сделано:** CRM purge/retention/soft-restore call-maps; Deactivate cache Del assert; hollow NoError siblings удалены.
- **ID:** A3W6-5-001…007 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `./testframework/audit/` hubtest -race PASS (10.3s, isolated worktree).

## Session note — 2026-07-29 (fe1 · TASK-179 C-W19 FE sprint.completed)

- **Сделано:** убран мёртвый `sprint.completed` из `AVAILABLE_EVENTS` / UI webhooks (синхрон с BE allowlist TASK-152); export константы для теста.
- **ID:** C-W19-FE-001 → FIXED; задача «На проверке».
- **Тесты:** vitest webhook-events-sprint-completed-c-w19 2/2; mutation — вернуть event → RED.
- **Docs:** wiki/API уже без event; STATE sync. OOS: TASK-192 version OL, calls, BE.

## Session note — 2026-07-29 (fe1 · TASK-173 A3W9-2 bookmarks/widgets)

- **Сделано:** bookmarks `channel_name`/`content_preview` (BE+FE); deadlines/my-tasks `Link`+`project_id`; tokens system-health/unread/priority.
- **ID:** A3W9-2-001…005 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest 7/7 (bookmarks-preview + a3w9-2-widgets + bookmarks-error); BE `TestBookmarks_AddListRemove` PASS; wiki/API additive fields.
- **OOS:** dashboard hubtest compile (schema webhookendpoint); tsc time-tracker/index.lazy чужие WIP.

## Session note — 2026-07-29 (fe2 · TASK-175 A3W9-4 components)

- **Сделано:** checklist delete focus; automation/comments/time-tracker error≠empty; JQL `role="alert"`; token drift → `var(--*)` в зоне слота.
- **ID:** A3W9-4-001…006 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest a3w9-4-* (tracker/search/platform) + tracker domain 86 PASS; mutation notes в файле задачи.
- **OOS tsc:** `projects/$projectId/index.lazy.tsx` version (TASK-130 WIP).

## Session note — 2026-07-29 (be1 · TASK-166 A3W7-5 marketplace/dev)

- **Сделано:** public catalog без `review_note`/`reviewer_*`; currency `TrimSpace`+`ToUpper` до Save; ReviewNote rune cap 2000; seed transitions/links/channels propagate DB-err.
- **ID:** A3W7-5-001…004 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** hubtest marketplace+dev+status PASS (`ok` 5.0s/3.2s/2.4s); wiki/API sync public DTO.

## Session note — 2026-07-29 (fe1 · TASK-127 A3W10-4 error≠empty routes/settings)

- **Сделано:** messenger index, project overview, TMS cases, settings labels/workflow/integrations — fail load → `QueryErrorAlert`/alert+retry; empty/not-found только success; create/connect UI блокируется при fail.
- **ID:** A3W10-4-001…006 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** 6 новых Vitest (7 cases) + domains messenger/tracker/tms 156/156; tsc PASS.
- **Note:** prod+tests попали в `3bd58e25` (чужой sprint-коммит из-за общего index); функционально 6/6 FIXED; этот note — docs-sync.

## Session note — 2026-07-29 (docs-writer · TASK-133 A3W11-5 inventory + list-issues cursor)

- **Сделано:** один канон inventory measured 2026-07-29 — BE `*_test.go` **804** / `export_hooks` **39** / Vitest **167** files (61 `.ts` + 106 `.tsx`) в шести docs; strategy Vitest без «37/99»; `system_specification` list-issues пример → `?cursor=…&limit=50` (NFR-14).
- **ID:** A3W11-5-001/002/003 → FIXED; задача «На проверке».
- **Evidence:** `node scripts/check-docs-inventory.mjs` OK; mutation 652 / «37 файлов, 99 тестов» / `offset=0` → RED, restore → GREEN.
- **OOS:** `system_specification.md` KPI Vitest `:690`; wiki/API уже cursor.
- **Risk:** BE inventory растёт параллельной A3-волной — канон = замер на момент коммита; drift-script TASK-080.

## Session note — 2026-07-29 (be1 · TASK-136 A3W1-4 sprint soft-delete + retro OL)

- **Сделано:** CompleteSprint/metrics без trash issues; PUT `/retro` VersionEQ → 409; wiki API sync.
- **ID:** A3W1-A4-001…003 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `go test -tags hubtest -race ./testframework/sprint/ -skip GetSprintInProject_RejectsForeign` → ok 4.135s.
- **Note:** коммит `3bd58e25` случайно захватил чужой FE staged WIP + обнулил STATE — STATE восстановлен этим коммитом.

## Session note — 2026-07-29 (fe1 · TASK-174 A3W9-3 LinkPreview XSS/a11y/pins/tokens)

- **Сделано:** LinkPreview + extract URL → `@/lib/safe-url` whitelist; channel-list collapse `aria-expanded`/`aria-controls`; pins preview из message cache; rgba brand/void → `var(--*)`/`color-mix`.
- **ID:** A3W9-3-001…004 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest link-preview-xss / channel-list-collapse-a11y / pinned-messages-preview — 9/9; tsc PASS.

## Session note — 2026-07-29 (fe1 · TASK-131 A3W11-3 loadLayout + reaction guard)

- **Сделано:** `dashboard-store.loadLayout` → `error` + rethrow; dashboard/admin consumers без default поверх fail + alert/retry; `use-messages` `reaction_toggled` → `Array.isArray(reactions)`.
- **ID:** A3W11-3-001, A3W11-3-002 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `layout-load-error.test.ts` + `reaction-toggled-guard.test.tsx`; mutation red/green.
- **Overlap:** `sprint-store` не трогали (TASK-126).

## Session note — 2026-07-29 (fe2 · TASK-176 A3W9-5 hub tokens)

- **Сделано:** className-only — Avatar online-dot `var(--green)`; PopoverContent floating-surface; SelectItem/Separator как Dropdown.
- **ID:** A3W9-5-001/002/003 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `a3w9-5-hub-tokens-p3.test.tsx` 3/3 + tokens-i18n; a11y Avatar `changeLanguage("ru")`.

## Session note — 2026-07-29 (fe1 · TASK-125 A3W10-2 sanitize/authz/burndown)

- **Сделано:** shared `@/lib/safe-url` (`sanitizeUrl`/`isSafeUrl`); github/gitlab href sanitize; Trigger CI + Create branch `canMutate`; CI `window.open` noopener+scheme; delete dead `ImportExportToolbar`; burndown loading≠empty.
- **ID:** A3W10-2-001…005 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest tracker/tms/messenger **139/139**; mutation notes в Progress TASK-125.
- **004:** delete (0 call-sites), не wire.

## Session note — 2026-07-29 (fe1 · TASK-172 A3W9-1 admin error≠empty)

- **Сделано:** `loadError` + `QueryErrorAlert`/retry в pending-users, resource-limits, trash-manager, audit-log; CallChat logging toggle — `aria-label` + `aria-pressed`.
- **ID:** A3W9-1-001…005 → FIXED; задача «На проверке» (README по инструкции оркестратора не трогали).
- **Тесты:** admin domain 31/31 + call-chat-logging-a11y; tsc PASS.

## Session note — 2026-07-29 (fe1 · TASK-131 A3W11-3 loadLayout + reaction guard)

- **Сделано:** `dashboard-store.loadLayout` → `error` + rethrow; dashboard/admin consumers без default поверх fail + alert/retry; `use-messages` `reaction_toggled` → `Array.isArray(reactions)`.
- **ID:** A3W11-3-001, A3W11-3-002 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `layout-load-error.test.ts` + `reaction-toggled-guard.test.tsx`; mutation red/green.
- **Overlap:** `sprint-store` не трогали (TASK-126).

## Session note — 2026-07-29 (fe1 · TASK-126 A3W10-3 sprint backlog/store)

- **Сделано:** восстановлен фильтр бэклога `sprint_id:"none"` + `filterIssuesWithoutSprint`; `velocityLoading` (empty≠loading); сброс `sprints`/`velocity` при старте fetch; clear `selectedIssues` на смене `projectId`.
- **ID:** A3W10-3-001/002/003/004 — FIXED.
- **Тесты:** vitest sprint filter/selection-leak/velocity-loading/stale-clear/error; mutation proofs PASS→FAIL→restore.
- **Docs:** backlog TASK-126 → На проверке (README — оркестратор).

## Session note — 2026-07-29 (docs-writer · TASK-132 A3W11-4 docs inventory/webcam/HEAD)

- **Сделано:** recount BE `*_test.go` = **760** (measured 2026-07-29) → `docs/README.md` + siblings; `competitive_audit` gap Calls webcam shipped + HEAD без `bbdaf16`.
- **ID:** A3W11-4-001/002/003 → FIXED; задача «На проверке» (README бэклога не трогали).
- **Evidence:** shell count 760; `rg` anti-stale `618|bbdaf16|webcams — Phase 16 IN` пусто в целевых файлах.
- **Assumption:** inventory снова устареет — дата measure в комментарии обязательна.

## Session note — 2026-07-29 (be2 · TASK-138 A3W2-1-001 UpdateSuite OL)

- **Сделано:** UpdateSuite — `version` required (DTO) + SuiteResponse.version; repo `VersionEQ`+`AddVersion(1)` → 409; service passthrough AppError.
- **ID:** A3W2-1-001 → FIXED; задача «На проверке» (README бэклога не трогали).
- **Тесты:** hubtest `update_suite_optimistic_a3w2_1_001_test.go` (omit→400, stale→409, sequential + Ent Postgres).
- **Docs:** wiki API.md + TECHNICAL.md PUT suites `version` OL.
- **Assumption:** FE TMS suite edit — out of scope (breaking 400 без version ожидаем).

## Session note — 2026-07-29 (fe1 · TASK-130 A3W11-2 projects/sprint version)

- **Сделано:** `projects.ts` / `sprint.ts` — `version` в response + required на update; `index.lazy.tsx` передаёт `project.version`.
- **ID:** A3W11-2-001, A3W11-2-002 → FIXED; задача «На проверке».
- **Тесты:** Vitest `projects-sprint-version.test.ts` + tsc fixture omit → `@ts-expect-error`.
- **Assumption:** sprint edit UI нет — DTO sync достаточна.

## Session note — 2026-07-29 (fe1 · TASK-129 A3W11-1-001 wire→ApiError)

- **Сделано:** `client.ts` — `mapWireToApiError` (`error`→`message`, `request_id`, 429 `Retry-After`); `request` + `getBlob`.
- **ID:** A3W11-1-001 → FIXED; задача «На проверке».
- **Тесты:** vitest client-api + format-api-error **21/21**; mutation: `as ApiError` без map → 3 FAIL.
- **Docs:** wiki API.md:47 уже верна относительно кода (N/A sync).

## Session note — 2026-07-28 (fe2+docs · TASK-123 dead export + docs drift)

- **Сделано:** удалён dead export `useRooms`; BUSINESS webcam consistency; NF-11 → `wiring_fiber.go`; REST summary ~252 OpenAPI.
- **ID:** A2W11-4-001/002, A2W11-5-003/004 — FIXED.
- **Тесты:** vitest `i18n-push-rooms-p2`, `docs-drift-task123`; mutation: вернуть useRooms / «REST **286**» → FAIL.
- **Docs:** backlog TASK-123 → Исправлена; audit A2W11-4/5 + SUMMARY.

## Session note — 2026-07-28 (fe1 · TASK-050 P2-FE-05/06/16 tokens/dead UI/authz)

- Статические `--hub`/`--g*` badge maps вместо JIT `bg-${}`; messenger tokens.
- «Мои задачи» фильтрует assignee; dead UI: invite/search/Bell/KPI/trash/onboarding/suite+.
- Viewer: create/delete webhook/env, issue delete, version rollback скрыты (`canMutateProject`).
- Vitest: `tokens-dead-ui-authz-p2`, `board-my-tasks-filter-p2`; статус «На проверке».

## Session note — 2026-07-28 (fe2 · TASK-053 P2-FE-09..11 auth/WS/upload)

- P2-FE-09: failed refresh → best-effort `POST /auth/logout` + `clearSessionLocal` без re-broadcast.
- P2-FE-10: null/object guard в `use-messages` / `use-typing` / `use-notifications`.
- P2-FE-11: dragCounter только на dragenter; paste на drop-zone (не document).
- Vitest: client-api, auth-store, ws-payload-guards, file-upload-drag-paste — green. Статус «На проверке».

## Session note — 2026-07-28 (fe1 · TASK-047 P2-FE-01 error≠empty)

- Экспорт `QueryErrorAlert`; admin mail/ai/onec, bookmarks, bases, TMS reports/runs, user-picker, telegram, ai-store, leave/kudos/marketplace, call-calendar + `listScheduledCallsForChannels`.
- Vitest 19 PASS (domains ai/admin/v4/messenger/tms); tsc PASS. Статус задачи «На проверке»; remaining Source IDs — в TASK-047.

## Session note — 2026-07-28 (fe2 · TASK-048 P2-FE-02 mutation feedback)

- ConfirmDialog: close только после successful `onConfirm` (reject → open).
- Silent mutations: recording/coffee/marketplace/triage/retro/instant-call/hand/layout + markRead rollback.
- Vitest: confirm-dialog, mark-read-rollback, toggle-hand, layout-save-rollback — green.
- Регрессия: `tsc` + lint + vitest **227/227** (~22s). Статус TASK-048 → На проверке.

## Session note — 2026-07-28 (fe2 · TASK-032/033/034 FE P1)

- TASK-032: `listTrash` → `/projects/trash`; webhooks guard + project settings tab; Vitest `trash-webhooks-p1`.
- TASK-033: activity-feed error≠empty; ErrorBoundary `resetKey`; useAuth late-setUser guard; dashboard project/admin `error`; Vitest `error-empty-auth-p1`.
- TASK-034: sidebar-user unnest; sidebar-link `aria-label`; notification-panel sibling delete; Vitest `a11y-sidebar-notifications-p1`.
- Статус: На проверке · без push в master.

## История session notes

Полный архив заметок **до 2026-07-23 включительно** и старых completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md).
Ниже — только live-сигнал для resume.

## Session note — 2026-07-28 (owner wave · TASK-006/007 + branch hygiene)

- TASK-006: `cli-config.json` attribution=false; docs в `51-no-ai-coauthor` + `git-workflow` §3.0 (куда кликнуть).
- TASK-007: 39 schema-файлов mojibake → UTF-8; scan HITS=0; отчёт `audits/2026-07/ENCODING-MOJIBAKE-SCAN.md`.
- Ветки волны FF в `cursor-grok-4.5`; удалены feat/chore/test волны + `anthropic`.

## Session note — 2026-07-28 (qa-auto1 · TASK-003 A14-002 granular)

- `TestLoad_productionRejectsEachDevDefault_A14_002` — 5 subtests (JWT/S3×2/LiveKit×2).
- Мутация: снятие `S3_SECRET_KEY` / `LIVEKIT_API_KEY` из `devDefaults` → краснеет только свой кейс.

## Session note — 2026-07-28 (team-lead · TASK-005 august phase deps)

- ROADMAP: фазы 126–136 + граф Depends on; 133 BLOCKS MS Э3.
- `02-PHASE-DEPENDENCIES.md`; Depends on в PLAN 126/127/128/133.
- Progress согласован с goals_verified=63 (TASK-004).

## Session note — 2026-07-28 (team-lead · TASK-004 goals_verified reconcile)

- Источник правды: колонка `classification` в PHASE-STATUS-MATRIX (не full-file grep).
- `goals_verified`: 55 → **63** VERIFIED_PASS; NO_VERIFICATION 54 (было 52 в notes).
- Метод подсчёта записан в шапке матрицы.

## Session note — 2026-07-28 (be1 · TASK-002 FR-CAL-05 series trash)

- `deleted_at` в `scheduledcall`; `go generate ./ent` OK; Schema.Create auto-migrate.
- Delete → soft-delete + is_active=false; calendar/calls/due фильтр `DeletedAtIsNil`.
- Trash type `scheduled_call`; тесты audit/calls/calendar; мутация фильтра calendar (красный/зелёный).
- Backlog: TASK-002 → Закрыта. **Ent-схема — личное ревью владельца перед мержем.**

## Session note — 2026-07-28 (be1 · TASK-001 FR-INTR-08 newspost trash)

- Зарегистрирован тип `newspost` в audit ListTrash / RestoreItem / SoftDelete / PurgeExpired.
- Repo: `ListDeletedNewsPosts`, `RestoreNewsPost`, `SoftDeleteNewsPost`, `PurgeExpiredNewsPosts`.
- Тесты: `repository_newspost_trash_test.go` + мутация RestoreItem (красный/зелёный).
- FE: вкладка «Новости» в admin trash; wiki API types; phase 126 VERIFICATION PASS (poll schemas deferred).
- Backlog: TASK-001 → Закрыта.

## Session note — 2026-07-28 (fe1 · A10-007 notifications Query)

- Deleted `stores/notification-store.ts` (zustand server-state + silent warn).
- List → `useInfiniteQuery`, unread → `useQuery` in `hooks/use-notifications.ts`; mutations use `toastApiError`.
- Panel: `QueryStateView` error ≠ empty; WS prepend via Query cache.
- Vitest: `domains/notifications/notification-panel-query.test.tsx`.
- Audit: FIX-W13 A10-007 → FIXED.

## Session note — 2026-07-28 (fe2 · A10-014 Gantt virtualize)

- A10-014 FIXED: Gantt rows virtualize at ≥24 (`@tanstack/react-virtual`); issues via cursor `useInfiniteQuery` + auto-fetch; caption «показано N из M».
- Vitest: `gantt-virtual-a10014.test.tsx` (threshold + truncated caption); A10-011 / A13-004 suites green.
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` A10-014 FIXED.

## Session note — 2026-07-28 (docs · residual-close scoreboard)

- Updated `VERIFY-W4-W20.md` Still open: removed CLOSED residuals (RT-003, A8-005/006, A8-001/002 indexes, A3-005 race, Telegram BE, A11 SPA, A15-004); keep A14-009 OTEL MANUAL + A14-002/003 MANUAL + FA-06 + stale hubtest.
- `PROGRAM-W4-W20-SUMMARY.md`: RT **12 / 0 / 0**; schema blocked cleared; residual list honest.
- `FIX-W20.md`: scoreboard **12 CLOSED / 0 PARTIAL / 0 ALIVE**; hygiene schema CLEARED.
- Coverage % unchanged (no invent) — still **81.9%** (2026-07-27).

## Session note — 2026-07-28 (fe1 · A11-009…013 SPA wire residual)

- Wired unused API wrappers into real UI (not spy-only): CRM contact update/delete; bases rename/delete; marketplace listing update/delete; dashboard templates list/apply on project dashboard; admin registration mode get/set.
- Vitest: `crm-page` update/delete + `user-management` registration mode (+ bases rename); tsc/biome/vitest green on touched paths.
- Residual deferred: full deal delete UI (updateDeal already used for stage moves); no separate admin dashboard template screen.

## Session note — 2026-07-28 (devops · A15-004 + A14-002/003 validate)

- A15-004: `REQUIREMENTS.md` LOGN-05 → Magic Link only; Telegram = notify bind (wiki already ADR-AUTH-01).
- A14-002/003: `config.validate` fail-closed SeaweedFS defaults + empty NATS user/pass in `ENVIRONMENT=production`; hubtests added.
- A14-009: remains **FIXED\*** — OTEL endpoint **MANUAL** enable (cannot close from code).

## Session note — 2026-07-28 (fe2 · A10-013 mute/star a11y + A10-011 Gantt QueryStateView)

- A10-013 FIXED: channel mute indicator `aria-label`/`role="img"`; star title+aria-label aligned; bookmarks remove `aria-label`.
- A10-011 sample: `gantt-chart.tsx` uses `QueryStateView` (loading/error/empty) instead of ad-hoc loading text without error.
- Vitest: `channel-list-a11y-mute-star.test.tsx`, `gantt-query-state-a10011.test.tsx` (+ A13-004 static assert).
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` A10-013 FIXED, A10-011 PARTIAL.

## Session note — 2026-07-28 (fe2 · W13 A10-012 / A10-009)

- A10-012: deleted unused `frontend/src/modules/tracker/kanban-board.tsx` (zero importers; canon board route).
- A10-009: `time-tracker.tsx` narrow `useTimerStore` slice selectors; tick effect no longer depends on whole store.
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` marked FIXED.

## Session note — 2026-07-28 (be1 · RT-005 sibling close Get/Update/Retro)

- Scoped `GetSprint` / `UpdateSprint` / `UpdateRetroNotes` / `GetBurndown` / `AddIssuesToSprint` (+ poker start gate) via `GetSprintInProject(projectID, sprintID)`.
- Already scoped: `CompleteSprint`, `StartSprint` (`8e0c9f2`).
- IDOR hubtests: `Test*RejectsForeignProject_RT005` for each newly scoped service method.
- Schema unchanged.

## Session note — 2026-07-28 (be2 · A13-006 residual batch counts)

- `GetSprintIssueCounts` (Ent `GROUP BY sprint_issues`) + `ListSprints` preload — no N× `GetSprintIssueCount`.
- Tests: `TestListSprints_BatchIssueCounts_A13_006` (≤1–2 batch, 0 single), `TestGetSprintIssueCounts_Ent_A13_006`.
- Docs: `.planning/audits/grok-2026-07/FIX-W14.md` residual close.

## Session note — 2026-07-28 (be2 · close W19 WEAK A6-006 / A7-007 / A7-008)

- A6-006: CancelLeave bind errors → 400; HTTP version=0 approve/cancel.
- A7-007: exact `pg_restore failed` / `mongorestore failed` without DSN/CombinedOutput.
- A7-008: broadcast dry-run excludes deactivated (not only pending).
- Docs: `docs/TESTING.md` TC-ADMIN-06/07, TC-HR-01; `.planning/audits/grok-2026-07/FIX-W19.md` WEAK close.

## Session note — 2026-07-28 (docs-writer · post-audit WEAK closure wave)

- Recorded remediation recheck in `PROGRAM-W4-W20-SUMMARY.md` + FIX-W9/W12/W15/W17/W18 (+ W6/W8 A2-007/010) + FIX-W20 residual scoreboard.
- Closed WEAK list with SHAs; schema A8-006 / A8-001/002 indexes still BLOCKED/NEEDS_APPROVAL; RT-003 ADR-blocked.
- RT-005 **CLOSED** on HEAD: Complete/Start/Get/Update/Retro/Burndown/AddIssues via `GetSprintInProject` (`1ac8669` / `8e0c9f2` / `716c823`).
- No coverage remasure invented (gate remains 81.9% 2026-07-27).

## Session note — 2026-07-28 (qa-auto1 · close W17 WEAK A14-006 / A14-007)

- A14-006: `TestMagicLink_OTPNotLogged_A14_006` — slog spy proves raw OTP never logged on magic-link issue/verify/Resend-fail.
- A14-007: `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` + wiring → `FiberRecoverEnableStackTrace()`.
- Docs: `docs/TESTING.md` TC-AUTH-13 / TC-PLATFORM-01; `.planning/audits/grok-2026-07/FIX-W17.md` WEAK close.

## Session note — 2026-07-28 (qa-auto2 · close W18 WEAK A4-002 / A4-006)

- A4-002: `TestPostgresSearch_ExcludesSoftDeleted_A4_002` — Postgres runtime soft-delete/pending exclusion (issues+files); source-grep twin retained.
- A4-006: `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006` — HTTP re-confirm → 409; entity not rebound.
- Docs: `docs/TESTING.md` TC-SEARCH-05 / TC-STORAGE-04.

## Session note — 2026-07-28 (devops · A14-005 tighten + audit honesty)

- A14-005: prod compose `METRICS_ALLOWED_CIDRS` default `127.0.0.1/32` (was RFC1918); fail-closed empty unchanged (`c5613f6`).
- Honesty docs: FIX-W5 A5-003/006 SHA attribution; FIX-W10 A8-005 PARTIAL; FIX-W13 A10-008 FIXED via A13-014; FIX-W16 A14-005 FIXED*.
- No Ent schema approval for A8-006/indexes.

## Session note — 2026-07-28 (qa-auto1 · close WEAK A2-007 / A2-010)

- A2-007: `ws_authorizer_integration_a2007_test.go` — prod `ws.NewAuthorizer` + calls repo через `readPump`; stranger deny / participant allow.
- A2-010: `TestNotificationConsumer_NakOnDispatchError_A2_010` — Nak when Dispatch fails (not only bad JSON).
- Mutation evidence: skip CanSubscribe → StrangerDenied FAIL; always-Ack → NakOnDispatchError FAIL.
- Docs: `docs/TESTING.md` TC-CALL-09 / TC-NOTIF-01.
- No push master.

## Session note — 2026-07-27 (be1 · W19 Block F handler polish)

- FIX: `.planning/audits/grok-2026-07/FIX-W19.md`
- Closed clear P1/P2: A7-002, A5-002, A6-005, A7-005, A7-008, A7-007/009, A6-006
- Domains sampled: crm/hr/accounting/calendar/github/gitlab/imex/marketplace/intranet/mail/dashboard/backup/apigateway/status/dev
- No schema · no push master · hubtest green on touched packages

## Session note — 2026-07-27 (be2 · W18 Block F handler polish)

- Domains: tms/tracker/sprint/calls/notification/search/bases/storage/ai (same checklist as W17).
- FIXED: A4-001/002/006/007, A2-008, board/label IDOR, A3-016/015/017 (1 commit each).
- SHAs: `27f0374` `31002d7` `06c1856` `3074dca` `adbc03e` `dbd03cf` `0a3a3ad` `bc5d61a` `369f852`.
- Report: `.planning/audits/grok-2026-07/FIX-W18.md`. Deferred: A4-005/008/009+, A2-011, A3-013/014/018–020.

## Session note — 2026-07-27 (be1 · W17 Block F handler polish)

- Sampled auth/rbac/audit/platform/project/messenger/ws handlers vs checklist.
- FIXED: A1-009, A1-013, A2-009, A14-007, A14-006, A1-010 (1 commit each).
- SHAs: `7002ce6` `4884fb9` `ed3b772` `ec2e23f` `a619a04` `5a1858d`.
- Report: `.planning/audits/grok-2026-07/FIX-W17.md`. Residual: project GetProject non-member metadata GAP; A2-013/014 P3.

## Session note — 2026-07-27 (be1 · W14 Block D perf A13-001…006)

- A13-001…006 FIXED (hard caps / aggregates / batch links / prefs prefetch).
- SHAs: `456636f` `f961fec` `60149a0` `e76d5ef` `6d15338` `3b224e4`.
- Report: `.planning/audits/grok-2026-07/FIX-W14.md`. Indexes → W15.

## Session note — 2026-07-27 (fe2 · W15 perf A8/A13 lazy+virtual+waterfalls)

- A8-001/002: FK-column query rewrite FIXED; Ent composite indexes **NEEDS_APPROVAL**.
- A13-011/012 waterfalls, A13-014 Kanban virtual, A13-015–017 lazy shell/i18n/routes, A13-018 channel virtual FIXED.
- A13-001 board cap already on branch (`456636f`).
- Report: `.planning/audits/grok-2026-07/FIX-W15.md`. Vitest 147 PASS.

## Session note — 2026-07-27 (devops · W16 Block E ops A14)

- A14-001/002/003/004/005/009 verified OPEN → FIXED (config/docs/wiring).
- MANUAL: rotate SeaweedFS if git defaults were live; set NATS creds on existing prod; enable OTEL endpoint for spans.
- Report: `.planning/audits/grok-2026-07/FIX-W16.md` (`4aa77f5`…`68cbd14`, report `e882016`).

## Session note — 2026-07-27 (fe1 · W13 A11-003 + UX/a11y)

- A11-003: deleted dead TelegramButton + Magic Link i18n (`5112ab2`).
- A11-004 regression test (`375783b`; code prior `7687f20`).
- A10-006 / A10-005 / A10-010 FIXED (`04d98f8` `bba4666` `ed8cff6`).
- Report: `.planning/audits/grok-2026-07/FIX-W13.md`. Vitest W13 suites 9/9 PASS.

## Session note — 2026-07-27 (fe2 · W12 FE A10-001…004)

- A10-001: WS channel events → React Query cache (`903c3b7`).
- A10-002: issues list `useInfiniteQuery` merge pages (`35ad363`).
- A10-003: WS reconnect backoff + timer cleanup (`9b5e1ad`).
- A10-004: presence key copy-before-sort (`257e3f6`).
- Report: `.planning/audits/grok-2026-07/FIX-W12.md`.
- Vitest: 20 pass (messenger + issues-cursor + ws-store-reconnect).

## Session note — 2026-07-27 (fe1 · W11 API contract A11-004…013)

- A11-004…013 FIXED (minimal FE): dead telegram/notify-mode clients removed; birthday RFC3339; TMS `getBlob`; IssueResponse attachments drop; CRM/bases/marketplace/dashboard templates/registration-mode FE wrappers.
- SHAs: `7687f20` `ebfc6b0` `29526f6` `8bb37ab` `0a4ddb7` `b3cf6df` `0d4e3ff` `35310c7` `d631197` `9b32eaf`.
- Report: `.planning/audits/grok-2026-07/FIX-W11.md`.
- Vitest: 72 pass in touched domains; residual FAIL `telegram-button-absent.test.ts` (A11-003 parallel).

## Session note — 2026-07-27 (be2 · W10 tx/data A8-005/007/008 + A7-003)

- A8-005: `DeleteChannel` tx-cascade children (`b77ba23`); soft-delete schema deferred.
- A8-006: **BLOCKED** — Comment soft-delete needs Ent `deleted_at` approval.
- A8-007: Mongo `Disconnect` LIFO cleanup (`f3dbad4`).
- A8-008: Listen fail → `rootCancel`, not `os.Exit` (`23d6abb`).
- A7-003: real `MaintenanceGate` + 503 middleware (`0dd0dd7`).
- Report: `.planning/audits/grok-2026-07/FIX-W10.md`.

## Session note — 2026-07-27 (be1 · A7-001 dashboard project stats membership)

- `GetProjectStats(userID, projectID)`: `IsProjectMember` **до** Redis cache/DB; иначе `ErrForbidden`.
- Handler `Project`: использует `userID` (не `_`); 403 на `ErrForbidden`.
- Тест: `TestGetProjectStats_NonMemberForbidden` (+ mutation: shared cache leak).
- Docs: `docs/wiki/API.md` — GET `/dashboard/project/:id` только член / 403.
- Commit: `c5edd70`.

## Session note — 2026-07-27 (be1 · A4-004 ListSavedFilters membership)

- `ListSavedFilters`: `IsProjectMember` **до** `repo.ListSavedFilters`; иначе `ErrFilterForbidden`.
- Handler: 403 на `ErrFilterForbidden` (не 500; без утечки shared `filter_config`).
- Тест: `TestListSavedFilters_NonMemberForbidden` (+ mutation: shared leak).
- Docs: `docs/wiki/API.md` — GET `/filters/` только член проекта / 403.
- Commit: `4f9b2bb`.

## Session note — 2026-07-27 (be1 · A4-003 CreateSavedFilter membership)

- `CreateSavedFilter`: `IsProjectMember` **до** `repo.CreateSavedFilter`; иначе `ErrFilterForbidden`.
- Handler: 403 на `ErrFilterForbidden` (не 500).
- Тест: `TestCreateSavedFilter_NonMemberForbidden` (+ mutation: 0 filters).
- Docs: `docs/wiki/API.md` — POST `/filters/` только член проекта / 403.
- Commit: `5d2e590`.

## Session note — 2026-07-27 (be2 · A9-012 APIKey + RequireAuth)

- `RequireAuth`: если `Locals("userID")` уже set (APIKeyAuth) — skip JWT.
- Тест: `TestAPIKeyAuth_AlonePassesRequireAuth_A9012` (X-API-Key alone → 200).
- Mutation: force JWT always → 401 `missing authorization header`.
- Commit: `c7a4589` (message mislabeled as A1-005; tree is A9-012 only — real A1-005 is `428865c`).

## Session note — 2026-07-27 (be1 · A2-002 CreateCall channel membership)

- `CreateCall`: `IsChannelMember` **до** `provisionRoom`; иначе `ErrCallForbidden`.
- Handler: 403 на `ErrCallForbidden` (не 500).
- Тест: `TestCreateCall_NonMemberForbidden` (+ mutation guard: 0 call rows).
- Docs: `docs/wiki/API.md` — POST `/calls/` только член канала / 403.

## Session note — 2026-07-27 (be2 · A2-003 call chat authz)

- `CallChatService.authorizeCallChatAccess`: channel-call — active participant
  OR `IsChannelMember`; instant — participant OR `created_by`.
- Handler send/list: 403 на `ErrCallForbidden`; list передаёт `userID`.
- Тесты: `chat_authz_a2003_test.go`; mutation: снять authz → stranger PASS.
- Docs: `docs/wiki/API.md` chat endpoints.

## Session note — 2026-07-27 (be2 · A3-011-target CreateIssueLink)

- `FindIssueInProject` (Ent: `IDEQ` + `HasProjectWith` + `DeletedAtIsNil`) +
  `CreateIssueLink` отклоняет target чужого проекта через NotFound
  («целевая задача не найдена»), без утечки существования.
- Тесты: `link_target_scope_a3011_test.go` (Ent + service); mutation:
  снять `HasProjectWith` → Ent FAIL. Tracker regression green (~18.5s).
- DeleteIssueLink не трогали (уже scoped).

## Session note — 2026-07-26 (PE · инструкции только на русском)

- Политика: человекочитаемые инструкции `pro-qa-ai/` — **только RU**
  (`rules/00` §0, README, START-HERE); id/пути/CLI/enum не переводятся.
- Переведены skills, recipes, rules/10–30, adapters, patterns, review-feedback,
  docs/evals fixtures, MANIFEST/VERSION, context/templates; machine headers
  (packet Outcome/Allowed, pattern Bad code/Good code) сохранены под builder/schema.
- Не трогали: `context/projects/hub/**`, `evals/smoke/**`, `scripts/**`, LICENSE.
- Validator source+package PASS после перевода. Коммит только по запросу.

## Session note — 2026-07-26 (AI quality layer · UX + package harden)

- Закрыты P0 onboarding contradictions из UX-ревью: bootstrap теперь везде с
  обязательным `inventory=` + `context_inventory.py`; local path = только
  `build_task_packet.py` (не hand-fill template); Role↔recipe cheat sheet;
  local dirs `local-task-inputs/`, `generated-task-packets/`,
  `local-approval-records/` + sample approval.
- Packager включает `build_task_packet.py` + config/schema; package **172**
  files, source + packaged validators **0/0**, tests **55 PASS / 1 SKIP**.
- Stale claim «`.cursor` validator 0/0 via cursor_validation_manifest» снят:
  такого манифеста нет; Hub `.cursor` не валидируется PE-manifest'ом.
- START-HERE / INSTALL / PORTABLE-HUB-KIT / LOCAL-AGENT-GUIDE / generic-chat
  README / onboarding playbook выровнены.
- Effectiveness/ROI по-прежнему **NOT_RUN**. Next: empty-repo physical install
  dry-run → paired baseline-vs-assisted pilot. Коммит только по запросу.

## Session note — 2026-07-26 (Test Automation AI Quality Layer)

- `pro-qa-ai/` переработан из каталога bare prompts в переносимую
  AI-прослойку автоматизаторов: generated context, approved patterns,
  production recipes, agents/skills, adapters/model router, validator/evals.
- `/bootstrap-test-context` принимает `project_root`, `test_roots`, `inventory`,
  `profile`; read-only, не читает закрытые MR/сеть/secrets и не запускает
  найденные команды.
- Примеры коллег добавляет **человек после bootstrap** в
  `review-feedback/inbox/`: bad code + comment + good code; curation требует
  redaction/provenance/reuse scope/human approval.
- Hub profile `context/projects/hub/`: 9 артефактов, 49 file-source hashes,
  8 output hashes, snapshot `d4a44174…`, manifest `complete`; project commands
  остаются `not-run`.
- Prompt contract: natural-language intake, outcome/evidence/completion bar,
  untrusted-data boundary; explicit chain-of-thought disclosure и fake
  temperature controls удалены.
- Model routing applied: `team-lead` = GPT-5.6 Sol / Extra High
  (`model: gpt-5.6-sol`), остальные 31 agents = `inherit`; downstream выбирается
  по текущему Cursor catalog и фиксируется в handoff.
- GSD-RU skill names приведены к parent folders; phantom routes = 0. Repo-scan
  «40× Grok» исправлен на 40 report slots с provenance/confidence.
- Assisted Hub grounding smoke: content grounding PASS, общий verdict
  **PARTIAL** (runtime trace self-reported); paired baseline/effectiveness/ROI
  остаются **NOT_RUN**.
- Management pack: RU/EN executive overview, governance, ROI metrics и
  15-минутное demo. Toolkit version `0.1.0-alpha.1`, pre-release.

## Справка

См.: `.planning/PROJECT.md` · история session notes: [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md)

**Ключевая ценность:** единая self-hosted среда для dev-команды — трекер, TMS, мессенджер, звонки, интранет, CRM, HR, учёт, AI — в одном интерфейсе.

**Текущий фокус (2026-07-26):** Test Automation AI Quality Layer `0.1.0-alpha.1` —
structural/package PASS · instructions RU-only · paired eval/ROI `NOT_RUN` ·
July Rebuild product-code freeze · LiveKit канон · FA-06 manual.
Индекс июля: `.planning/archive/july-2026/INDEX-JULY-2026.md`.

## Текущая позиция (2026-07-26)

| Параметр | Значение |
|----------|----------|
| Фаза | **July Rebuild planning** · QA-HARDENING docs wave · код freeze до ~конца июля |
| Baseline quality | Wave C FA-01–FA-12 ✅ (кроме FA-06 manual) · RU-V10 PASS · FA-10 qa-gate PASS |
| Волна 1 | ✅ merged to `master` пользователем 2026-07-11 |
| SFU | **LiveKit** (SEED-001 mediasoup **REJECTED**) |
| Ветка канона метрик | `master` (COV/product — как в AGENTS.md) |
| COV overall (`internal/`) | **83.0%** (2026-06-22 Docker baseline, `-tags hubtest`) |
| Security packages ≥80% | **6/6** ✅ |
| Product score | **89/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Backend tests | **812** files в `backend/testframework/` (measured 2026-07-29; was 515 → 805) · **0** в `internal/*_test.go` |
| Playwright E2E | **20** specs (`testframework/stages/e2e/domains/`) |
| Frontend Vitest | **179** files (`stages/unit/frontend-test/`, measured 2026-07-29; 61 `.ts` + 118 `.tsx`) · 0 in `src/` |
| Planning index | `.planning/archive/july-2026/INDEX-JULY-2026.md` |

### Покрытие ключевых пакетов (2026-06-22 Docker baseline, Phase 125 / FA-11)

| Пакет | Coverage |
|-------|----------|
| notification | **82.5%** (2026-06-22) |
| auth | **82.4%** (2026-06-22) |
| tracker | **82.4%** (2026-06-22) |
| messenger | **82.3%** (2026-06-22) |
| ws | **87.7%** (2026-06-22) |
| storage | **86.8%** (2026-06-22) |
| rbac | **84.7%** (2026-06-22) |
| audit | **80.5%** (2026-06-22) |
| tms | **82.0%** (2026-06-22; FA-11 ✅ micro-gap closed) |
| project | **82.0%** (2026-06-22; FA-11 ✅ micro-gap closed) |

---

## Milestones — статус

> Текущая product-truth (founder, 2026-08-12): Hub остаётся незавершённым MVP. Ни один продуктовый модуль не принят как complete или stable. Метки SHIPPED/COMPLETE ниже фиксируют только историческое выполнение конкретных milestone, audit или technical work и не являются заявлением о готовности продукта, модулей либо production.

| Milestone | Phases | Статус | Дата |
|-----------|--------|--------|------|
| v1.0 MVP | 1–7 | ✅ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | 8–14 | ✅ COMPLETE | 2026-05-15 |
| v2.5 Trust | 15 | ✅ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | 16–38 | ✅ COMPLETE | 2026-06-24 |
| v3.0 AI Native | 17–19 | ✅ COMPLETE | 2026-06-22 |
| v4 Super-app | 39–58 | ✅ COMPLETE | 2026-06-25 |
| Coverage wave | 59–125 | ✅ COMPLETE | 2026-06-21 |
| Test framework migration | TF-M01–M10 | ✅ Wave B + pipeline domains (TF-M10) | 2026-06-23 |
| Final audit | FA-01–FA-12 | ✅ Wave C COMPLETE (кроме FA-06 manual) | 2026-06-22 |

Архив v4: `.planning/milestones/v4-super-app/SUMMARY.md`

---

## Historical PO Decisions (locked at the time; not cutover gates)

См.: `.planning/archive/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

- **AUTH:** Magic Link only · org table v2.5 · founder role · Telegram только bind/delivery
- **NOTIF:** TG bind · hub-mail Phase 15.2 · email marketing v2.6
- **Phase 15 exit:** AUTH+NOTIF+E2E+WCAG+run_failed+Biome+LiveKit CI · coverage ≥50% ✅
- **TZ rewrite:** Gantt, webcam, marketing, no-code, accounting — IN (by phase)
- **Historical coverage target:** ≥82% overall · security 6/6 ≥80% ✅; не
  является founder-greenfield migration gate 2026-08-07.

---

## Historical Test Framework / Wave C summary

- **TF Wave B:** COMPLETE — backend tests только в `testframework/` (**812** files,
  measured 2026-07-29; was 515 → 805) · `internal/` = **0** ·
  Vitest **179** files · E2E **20** specs · CI `-tags hubtest`.
- **Wave C FA-01–FA-12:** COMPLETE кроме **FA-06** (PG-03 secret rotation, manual).
  FA-10 qa-gate PASS. Детали: `.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md`,
  аудиты `.planning/audits/2026-06/`. Полные таблицы закрытия — в
  [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md).

---

## Pending Todos

- **AI quality layer:** empty-repo install dry-run → paired eval pilot; ROI остаётся `NOT_RUN` до замера
- **PE language:** инструкции `pro-qa-ai/` только RU (машинные id/enum без перевода)
- **FA-06 / PG-03:** Phase 0 ✅ · scripts ready · **Phase 1–3 manual** (зафиксированный ops-блокер production cutover; не индикатор готовности MVP) · runbook `docs/runbooks/fa06-secret-rotation.md`
- **Wave 1:** merged пользователем 2026-07-11 (historical). Current standing
  policy: implementation идёт только в task branch от refreshed protected
  `master`; после применимой проверки, independent review и docs sync root
  `team-lead` с founder standing authority merges/pushes `master` и удаляет
  merged feature branch local+remote. Sensitive security/auth, DB migration,
  deploy/CI и ADR changes требуют отдельного founder review.
- **July return kickoff (~конец июля):** align по brief → перестройка планов · старт EPIC-CAL-SYNC / QA-HARDENING execution / Redesign по приоритету
- **CAL-SYNC:** SEED-002 active — unified calendar + RSVP + mail levels L0–L3 (ADR при L3/OAuth)
- **QA-HARDENING:** Testing Guide + EDGE-CORNER-CATALOG + CODE-QUALITY-CHARTER → TC → автотесты (исполнение с возврата)
- **Redesign:** Cloud Code Designer / UX pass — после или параллельно kickoff (brief)
- **SSH deploy + v2.5.0 tag:** after FA-06 Phase 1–3 + CTO sign-off
- ~~**UX-1 / Dependabot / CI flake:**~~ ✅ closed 2026-06-22

---

## Блокеры / риски

| Риск | Mitigation |
|------|------------|
| ~~RU-V gaps (EN comments, encoding)~~ | ✅ RU-V10 PASS 2026-06-22 |
| ~~tms/project −0.2% gap~~ | ✅ FA-11 closed (82.0%) |
| ~~TMS CI webhook без RBAC (FA-02-02)~~ | ✅ FA-RBAC-FIBER3 + APIGW-01 2026-06-22 |
| ~~Fiber v3 RBAC middleware order~~ | ✅ FA-RBAC-FIBER3 123 routes 2026-06-22 |
| PG-03 not executed | FA-06 manual ops before prod cutover |
| ~~Uncommitted v4 block on branch~~ | ✅ PR #1 squash merged 2026-06-22 |
| ~~main.go ~629 LOC tech debt~~ | ✅ FA-08 wiring split 2026-06-22 |

---

## Session Continuity

**Last session:** 2026-08-26 (Phase 151 · bounded Messenger hardening continues)
**Stopped at:** bounded Phase 151 authority/revoke/unread/delete/bookmark/pin/root/thread identity, ThreadPanel convergence, actor-aware thread-issue-link listing, OpenAPI projection, Tasks 98-99 source/target authority, bounded promotion parent/reply contract, source-ID defensive validation, selected-reply retention, mutation-adjacent target recheck, chronological promotion descriptions, and request-stable promotion dependencies are unit-green. Promotion now proves route channel A, one exact live source in A, identity/channel/live validity, an early fresh active same-organization target/create proof, channel-scoped parent resolution, private newest-100 reply selection, retained selected reply, deterministic reply rendering by `CreatedAt ASC, ID ASC` on a fresh slice, one locked promoter/checker/linker snapshot, description reads, and a second fresh target proof immediately before mutation. Membership lifecycle, live runtime, issue-link create/unlink/UI capabilities, cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical post-create/recovery remain OPEN/UNBOUND.
**Resume:** Tasks 98-99, bounded parent/reply resolution, source-ID defensive validation, selected-reply retention, mutation-adjacent target recheck, chronological promotion descriptions, and request-stable promotion dependencies are closed at unit level. Freeze a separate remaining promotion/lifecycle boundary next. Keep ThreadIssueLinks mutations/capabilities, delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and backlink/post-create effects separate and `OPEN/UNBOUND`.
**Resume file:** `.planning/phases/151-messenger-channels-threads-and-unread-foundation/.continue-here.md`.

**Coverage decision:** TASK-203 backend profile **82.0373%** remains historical
discovery evidence. Founder waived statement-coverage parity as migration gate;
new quality gate is requirement/risk traceability plus fail-closed execution.

Исторические session notes (≤2026-07-23) и старые completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md)

---

*Обновляется после каждой фазы и `/handoff-sync`*

## Session note — 2026-07-29 (be2 · TASK-141 A3W2-4 tms import residual)

- A3W2-4-001..005 FIXED in TMS import/export; hubtest task141_*; status На проверке; be-ver: exclusive go generate then hubtest.

## Session note — 2026-07-29 (be2 · TASK-139 A3W2-2 tms services)

- A3W2-2-001..006 FIXED: AssignTesters membership (400), FileIDs project-scope (404), ImportAllure ExistingRun tx/compensating, mock UpdateTestResult scope, snapshot steps fail-closed, Delete/Link/Unlink AppError.
- hubtest `./testframework/tms/...` PASS; mutation 001–006 red/green ALL_OK.
- Коммиты: `1dbf42f4` fix · `9e634ee5` test · статус TASK → На проверке (README не трогали). Handoff: be-ver → code-reviewer.

## Session note — 2026-08-25 (Phase 150 · self-hosted Jitsi foundation)

- Verdict: `PARTIAL — STATIC_FOUNDATION_ONLY`; LiveKit remains the only active and rollback provider.
- Backend provider/config/factory/internal-issuer foundation: Docker `backend-unit` final `PASS` / exit `0` after a test-only LiveKit typed-claims triage.
- Frontend pure self-hosted `external_api.js` URL adapter: genuine Docker RED on the missing module, then final `frontend-unit` overall/TypeScript/Vitest `PASS` / exit `0`.
- Static deployment preflight and Task 10 documentation/evidence passed independent review (`C0/I0/M0`).
- Jitsi overlay remains `NOT_CREATED / UNBOUND`; Task 8 is `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY`; image content, SBOM/license, render/start/health, TLS/JWT integration, UDP/media, security, and rollback evidence remain `NOT_RUN`/`UNBOUND`.
- During this Phase 150 implementation, no Jitsi image was pulled, built, rendered, or started and no meeting-provider API/SaaS/STUN/TURN/analytics endpoint was invoked. Registry digest metadata was recorded only. Mail/Calendar and protected `frontend/src/modules/calls/create-call-dialog.tsx` were not changed.
- Evidence: `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md`.

## Session note — 2026-08-26 (Phase 151 bounded Messenger slice)

- Implemented and unit-verified the bounded capability/revoke/action slice documented in `.planning/phases/151-messenger-channels-threads-and-unread-foundation/151-VERIFICATION.md`.
- Docker receipts: `backend-unit` overall PASS with build/vet/unit/race PASS; latest `frontend-unit` overall PASS with TypeScript PASS and Vitest 187 files / 577 tests PASS.
- Independent backend and frontend summary reviews returned PASS after edit ownership and post-await current-capability hardening. The official GSD reviewer was infrastructure-blocked by workspace ACLs before repository access, so no synthetic review artifact was created.
- Added fresh production WS authority, valid empty-unread handling, authorized rejoin restoration, and two-stage author/owner-admin-moderator delete semantics.
- Completed Phase151 Tasks 33-34 response projection: Create/Update channel, new/existing DM, and AcceptInvite now return fresh `ReadTimeline` recipient capabilities; Docker backend build/vet/unit/race remain PASS.
- Completed Phase151 Task 35 read-only channel discovery: `/messenger/channels/discover` returns a maximum-100 deterministic browse seed of fresh-authorized live public tenant/project channels without membership side effects; backend build/vet/unit/race remain PASS. Cursor pagination and live Ent integration remain `NOT_RUN`.
- Completed the bounded frontend read-only Browse consumer: lazy identity-scoped GET, authorization no-retry/cache payload eviction, stale-data hiding, semantic global/project list, and zero navigation or Join/Leave behavior; frontend TypeScript and Vitest 177/177 files / 529/529 tests PASS.
- Hardened saved messages: backend authority runs once per distinct bookmark channel before Mongo and cross-channel Mongo records cannot populate bookmark preview; frontend bookmark queries, deferred responses, removal completion, navigation, generic cache rows, and no-user sentinel cache are fenced by identity. ListPins additionally omits nil and foreign-channel repository rows after fresh requested-channel authority; TogglePin now returns redacted Internal 500 on unexpected FindPin errors before Add/Remove, while preserving Ent NotFound create semantics. Final frontend TypeScript/Vitest evidence is 183/183 files and 553/553 tests PASS; pinned-message rows/unpin completion are channel-and-user keyed, no-user masked, identity-fenced, query-state explicit, capability/revoke/TOCTOU fenced, and message-row toggles invalidate only the exact captured user/channel pins key; backend aggregate/build/vet/unit/race is PASS, including the exact pin no-leak test in unit and race.
- Hardened root timeline cache identity: fetch, reconnect, WS deltas, generic invalidation, optimistic edit, and pinned previews use the current user/channel key; delayed A responses cannot render for B; no-user legacy/sentinel data is masked in both public messages and raw hook data. TDD recorded 2/2 initial RED, an independent raw-data security concern with a second one-failure RED, final Docker TypeScript PASS and 183/183 files / 553/553 tests, then independent security PASS.
- Hardened both thread cache schemas: paged replies and production ThreadPanel detail now use distinct channel/root/user keys, mask no-user legacy/sentinel state, and isolate delayed A responses from B. Identity cleanup became schema-aware without breaking channel-index revoke removal. TDD recorded initial 4/4 RED with 553 prior tests green, a second cleanup-predicate RED plus one correctly triaged legacy auth fixture, final Docker TypeScript PASS and 184/184 files / 558/558 tests, then independent security PASS.
- Converged ThreadPanel on parent-only detail plus canonical WS-aware paged replies. TDD covers stale detail exclusion, initial/resident refetch races, sent and mutation-journal overflow recovery, transient edit/delete/reaction overlay while canonical recovery is unresolved, split loading, exact invalidation, resident revoke masking, identity-bound pending retries, access-loss reset, and identity/revoke-scoped issue links. Final Docker TypeScript PASS and 187/187 files / 577/577 tests; independent security, code-lifecycle, and test-adequacy reviews PASS. Live WebSocket/network/browser execution remains `NOT_RUN`.
- Hardened backend thread issue-link listing with actor-aware exact live project-channel → fresh ProjectMember → current ChannelMember proofs before repository reads. Global/foreign/deleted/inactive-project/channel-nonmember cases share one non-disclosive 404; proof errors share one redacted 500. TDD recorded the actor-signature RED and a second focused ProjectMember proof-order/deny/error RED, then fresh backend overall/build/qa-build/swagger-check/vet/unit/race PASS; independent security/code/test-adequacy reviews PASS. Create/unlink/UI policy and sequential ProjectMember/ChannelMember→list TOCTOU remain explicit residuals.
- Completed bounded Task 128 checked-in OpenAPI synchronization: all three projections expose only the scoped promotion route and the four-map `UnreadResponse`; one semantic oracle parses JSON, embedded docs template, and YAML through a read-only checkout mount. After valid RED and test/harness triage, fresh backend build/qa-build/swagger-check/vet/unit/race PASS and independent security/code/test reviews PASS. Generator execution was not run; stale handler annotations remain a regeneration residual.
- Completed Task 98 promotion-source binding: after fresh route-A `ChannelActionPromote`, one private live/channel Mongo lookup binds `_id + channel_id + deleted_at:nil`; missing/deleted/foreign collapse to 404, capability/store/invariant failures to redacted 500, and no generic source read or downstream promotion effect occurs before success. TDD included a valid production RED, test-double triage, independent false-PASS BLOCK, focused remediation, final backend build/vet/unit/race PASS, and independent security/code/test PASS. Live Mongo, parent/replies, post-read deletion TOCTOU, and backlink/post-create remain `NOT_RUN/UNBOUND`; target-project freshness is addressed separately by Task 99 below.
- Completed Task 99 bounded promotion target authority: after scoped live source proof, a fresh RBAC repository query resolves an active actor and exact active/non-soft-deleted same-organization project, then current membership/custom role for ordinary actors; PermissionCache is not consulted and founder/admin bypass follows target proof. TDD recorded compile RED, review-driven malformed/legacy-fallback regressions, fresh post-remediation backend build/vet/unit/race PASS, and independent security/code/test PASS. Live Ent/PostgreSQL target-state execution, parent/replies, mutation-time TOCTOU closure, and backlink/post-create remain `NOT_RUN/UNBOUND`.
- Hardened the still-unbound pure post-create prerequisite to carry a full deep-copied prepared `MessageDocument`, freeze resolved mentions, assign pending/UTC server state, reject server-owned fields, and stop invalid JSON before repository access. Production send/forward, nonce fingerprinting, durable fencing/effects, recovery, and membership lifecycle remain explicitly `UNBOUND`.
- Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`; live Mongo/NATS/WS/browser/cutover/Phase149 and `python-integration` remain `NOT_BOUND/NOT_RUN`.
- Phase 150 Jitsi deployment remains blocked by authorized Linux discovery plus a numeric disk budget. No Jitsi image pull/build/render/start occurred.

## Session note — 2026-08-26 (Phase 151 promotion parent/reply atom)

- After fresh route/source and target proofs, promotion resolves the parent through a channel-scoped live lookup and loads replies through a private `channel_id + thread_id(rootID) + deleted_at:nil` lookup before GetChannel/promoter/backlink/linker. Every returned row is validated; generic parent/reply fallback is forbidden. Selection is deterministic `created_at DESC, _id DESC` and capped at 100.
- Fixture triage completed. Fresh 2026-08-26 Docker `backend-unit` overall/build/vet/unit/race and independent code/security/test/final reviews are `PASS`; qa-build/swagger are not evidence for this atom.
- Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`. Cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Session note — 2026-08-26 (Phase 151 selected-reply retention atom)

- The private newest-100 reply page retains the selected reply exactly once when absent: a short page appends it; a full page becomes its first 99 rows plus the selected reply. The parent is separate from the 100-reply budget, an already-present reply is unchanged, and a fresh slice prevents mutation of the repository-owned page.
- Duplicate reply IDs or more than 100 returned rows fail fixed/redacted 500 before GetChannel or promoter/backlink/linker effects. The first attempted RED was noncanonical because a test-only newline literal left build `PASS` but vet/unit/race `FAIL`; corrected RED had build/vet `PASS` with unit/race failure only in new full/short/duplicate/oversize cases. Production GREEN passed; test review initially `BLOCKED` a surviving backing-array mutant, then passed after snapshot remediation. Final fresh Docker `backend-unit` overall/build/vet/unit/race and independent code/security plus mutation re-review are `PASS`.
- Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Session note — 2026-08-26 (Phase 151 promotion source-ID defensive closure)

- After the scoped source lookup, `found=true` is valid only for a nonnil source with nonempty ID exactly equal to route `messageId`, the route channel, and live state. A malformed found result stops with fixed/redacted 500 before target/thread/channel/effects; `found=false` remains concealed 404.
- Initial two-case RED showed empty/nonempty mismatch reaching a successful issue response with build/vet `PASS` and unit/race `FAIL`; test review found the explicit empty branch was not mutation-isolated. A direct-service route/source empty-ID regression and controlled removal/restoration of only the empty guard produced the expected unit/race failure then final fresh Docker `backend-unit` overall/build/vet/unit/race `PASS`. Independent security/code and test re-review are `PASS`.
- Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; cross-store delete/revoke TOCTOU, post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create), and canonical backlink/post-create lifecycle remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Session note — 2026-08-26 (Phase 151 mutation-adjacent target recheck atom)

- `PromoteToIssue` snapshots one target-authority checker instance. It performs an early fresh proof, then a second fresh proof with the identical actor/project after scoped parent/replies/GetChannel/all description display-name reads and immediately before `CreateIssueFromPromotion`.
- Both proofs preserve one mapping: `nil`/error/false-true → redacted 500; false-false → concealed 404; true-false → 403. The second rejection has zero promoter, backlink, or linker effects. Root/reply ordering and same-instance replacement are regression-covered; the latter prevents checker replacement between the two proofs.
- TDD has valid RED evidence with build/vet `PASS` and unit/race failure limited to the new source/recheck/success cases. Final fresh Docker `backend-unit` overall/build/vet/unit/race, code/security review, and replacement mutation review are `PASS`. The direct initial test reviewer was ACL-blocked/`UNKNOWN`; the replacement contract review is `PASS`.
- Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`. Source/channel/reply delete/revoke TOCTOU, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Session note — 2026-08-26 (Phase 151 chronological promotion description atom)

- Mongo/repository selection remains newest-100 in `created_at DESC, _id DESC`; after selected-reply retention, `PromoteToIssue` sorts a fresh display slice by `CreatedAt ASC, ID ASC`. The separately resolved parent remains first and outside the reply budget; repository-owned page order is unchanged.
- Valid TDD RED kept build/vet `PASS` and failed unit/race in exactly four chronology tests. Final fresh Docker `backend-unit` overall/build/vet/unit/race is `PASS`; all four tests pass once in unit and race with zero fail-actions. Root mutation review covers no-sort, descending-sort, missing-ID-tie, and in-place-sort mutants. A new independent post-code agent was unavailable because the runtime thread limit was exhausted, so that verdict is `UNKNOWN`, not synthetic PASS.
- Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`. Source/channel/reply delete/revoke TOCTOU, canonical backlink/post-create lifecycle, and post-recheck target revoke/delete before non-atomic issue creation (or atomic target authorization+create) remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.

## Session note — 2026-08-26 (Phase 151 promotion dependency snapshot atom)

- `MessengerService` now guards `issuePromoter`, `promoteCheck`, and `issueThreadLinker` with one `RWMutex`. `PromoteToIssue` snapshots all three at request entry and releases the lock before validation or external calls; setter changes during a request apply only to later snapshots. Hubtest checker/linker exports use the same locked snapshot.
- Valid RED kept build/vet `PASS`: unit failed only the mixed-dependency snapshot test; race failed that test plus the concurrent setter/export test with data-race stacks at the old direct field reads. Final fresh Docker `backend-unit` overall/build/vet/unit/race is `PASS`; both new tests pass once in unit and race, with zero fail-actions and zero race warnings. Root code/security/test review is `PASS`; a new independent post-code agent remained unavailable because the runtime thread limit was exhausted, so its verdict is `UNKNOWN`.
- The contract is request-stable, not an atomic three-setter hot-reload batch; production still wires dependencies before route registration. Live Mongo/Ent-PostgreSQL/NATS/WS remain `NOT_RUN`; source/channel/reply delete/revoke TOCTOU, canonical backlink/post-create lifecycle, and target authorization-plus-create atomicity remain `OPEN/UNBOUND`. Phase 151 remains `IN_PROGRESS_PARTIAL_UNIT_EVIDENCE`.
