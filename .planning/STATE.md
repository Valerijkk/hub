---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: planning-audit-pw-in-progress
stopped_at: Planning audit 2026-07 PW1–PW5 · coverage remeasured · phase matrix live
last_updated: "2026-07-27T00:45:00.000Z"
last_activity: 2026-07-27 · planning-audit PW waves · coverage 81.9% (2026-07-27)
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

## Session note — 2026-07-28 (be2 · TASK-110 calls recording/room/guest)

- A2W3-5-001..003 FIXED: StopEgress на fail CreateRecording; finalizeCreate rollback; invite base62×12.
- IT: `recording_room_guest_task110_test.go`; `go test -tags hubtest ./testframework/calls/` PASS.

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
