---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: planning-audit-pw-in-progress
stopped_at: Planning audit 2026-07 PW1–PW5 · coverage remeasured · phase matrix live
last_updated: "2026-07-29T17:10:00.000Z"
last_activity: 2026-07-29 · be2 TASK-181 C-W19 webhook soft-delete deactivate · На проверке
progress:
    total_phases: 125
    tasks_completed: 125
    goals_verified: 63
    phases_percent: 100
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
    overall_internal_statements: 81.9
    measured: "2026-07-27"
    baseline_2026_06_22: 83.0
    delta_pp: -1.1
    gate_floor_82: miss
inventory_measured_2026_07_27:
    frontend_route_tsx: 72
    backend_testframework_test_go: 515
    ent_entity_schemas: 82
    zustand_stores: 12
    go_mod: "1.25.0"
spec_coverage:
    fr_total: 28
    fr_with_phase: 9
    ratio: "9/28"
    decimal: 0.321
    percent: 32.1
    measured: "2026-07-27"
    source: .planning/audits/planning-audit-2026-07/SPEC-COVERAGE.md
---

# Состояние проекта Hub

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

- Политика: человекочитаемые инструкции `prompt-engineering/` — **только RU**
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

- `prompt-engineering/` переработан из каталога bare prompts в переносимую
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
- Model routing applied: `team-lead` = GPT-5.6 Sol Pro
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
| Backend tests | **515** files в `backend/testframework/` (measured 2026-07-27; was 499 on 2026-07-26) · **0** в `internal/*_test.go` |
| Playwright E2E | **20** specs (`testframework/stages/e2e/domains/`) |
| Frontend Vitest | **99/99** pass · 37 files `testframework/stages/unit/frontend-test/domains/` · 0 in `src/` |
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

## PO Decisions (locked)

См.: `.planning/archive/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

- **AUTH:** Magic Link only · org table v2.5 · founder role · Telegram только bind/delivery
- **NOTIF:** TG bind · hub-mail Phase 15.2 · email marketing v2.6
- **Phase 15 exit:** AUTH+NOTIF+E2E+WCAG+run_failed+Biome+LiveKit CI · coverage ≥50% ✅
- **TZ rewrite:** Gantt, webcam, marketing, no-code, accounting — IN (by phase)
- **Coverage target:** ≥82% overall · security 6/6 ≥80% ✅

---

## Test Framework / Wave C — live summary

- **TF Wave B:** COMPLETE — backend tests только в `testframework/` (**515** files,
  measured 2026-07-27; was 499 on 2026-07-26 / earlier migration note had 495) · `internal/` = **0** ·
  Vitest **99/99** · E2E **20** specs · CI `-tags hubtest`.
- **Wave C FA-01–FA-12:** COMPLETE кроме **FA-06** (PG-03 secret rotation, manual).
  FA-10 qa-gate PASS. Детали: `.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md`,
  аудиты `.planning/audits/2026-06/`. Полные таблицы закрытия — в
  [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md).

---

## Pending Todos

- **AI quality layer:** empty-repo install dry-run → paired eval pilot; ROI остаётся `NOT_RUN` до замера
- **PE language:** инструкции `prompt-engineering/` только RU (машинные id/enum без перевода)
- **FA-06 / PG-03:** Phase 0 ✅ · scripts ready · **Phase 1–3 manual** (единственный prod-блокер ops) · runbook `docs/runbooks/fa06-secret-rotation.md`
- **Wave 1:** merged пользователем 2026-07-11; policy остаётся — агенты не пушат/не мержат master
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

**Last session:** 2026-07-26 (AI quality layer · RU instructions · STATE compact)
**Stopped at:** structural/package validators PASS · tests 55 PASS / 1 SKIP ·
package 172 files · smoke PARTIAL · paired eval/ROI NOT_RUN
**Resume:** empty-repo portable install dry-run → paired baseline-vs-assisted
pilot; FA-06 manual; затем CAL-SYNC / QA-HARDENING / Redesign по приоритету.

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
