---
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: planning-audit-pw-in-progress
stopped_at: Planning audit 2026-07 PW1тАУPW5 ┬╖ coverage remeasured ┬╖ phase matrix live
last_updated: "2026-07-27T00:45:00.000Z"
last_activity: 2026-07-27 ┬╖ planning-audit PW waves ┬╖ coverage 81.9% (2026-07-27)
progress:
    total_phases: 125
    tasks_completed: 125
    goals_verified: 55
    phases_percent: 100
    notes: >-
      tasks_completed = checkbox/legacy completed_phases claim.
      goals_verified = VERIFIED_PASS in PHASE-STATUS-MATRIX.md (scanned 2026-07-27):
      55 PASS / 5 PARTIAL / 2 CODE_GATE_DEPLOY_BLOCKED / 52 NO_VERIFICATION
      among 114 dirs under .planning/phases/ (01тАУ13 live in milestones/).
      phases_percent remains tasks_completed/total_phases тАФ NOT goals_verified.
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
      - .planning/production/PG-03-SECRET-ROTATION.md
      - .planning/production/FA-06-EXECUTION-LOG.md
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
    zustand_stores: 13
    go_mod: "1.25.0"
spec_coverage:
    fr_total: 28
    fr_with_phase: 9
    ratio: "9/28"
    decimal: 0.321
    percent: 32.1
    measured: "2026-07-27"
    source: .planning/planning-audit-2026-07/SPEC-COVERAGE.md
---

# ╨б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╨┐╤А╨╛╨╡╨║╤В╨░ Hub

## ╨Ш╤Б╤В╨╛╤А╨╕╤П session notes

╨Я╨╛╨╗╨╜╤Л╨╣ ╨░╤А╤Е╨╕╨▓ ╨╖╨░╨╝╨╡╤В╨╛╨║ **╨┤╨╛ 2026-07-23 ╨▓╨║╨╗╤О╤З╨╕╤В╨╡╨╗╤М╨╜╨╛** ╨╕ ╤Б╤В╨░╤А╤Л╤Е completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](history/STATE-HISTORY-THROUGH-2026-07-23.md).
╨Э╨╕╨╢╨╡ тАФ ╤В╨╛╨╗╤М╨║╨╛ live-╤Б╨╕╨│╨╜╨░╨╗ ╨┤╨╗╤П resume.

## Session note — 2026-07-28 (be1 · A11-004 + RT-005 AddIssues FK)

- A11-004: removed dead `LoginWithTelegram` / `LoginWithTelegramUser` / `TelegramAuthRequest`; Magic Link only; notify bind untouched.
- RT-005: `AddIssuesToSprint` / `SetIssuesSprint` scope issues via `HasProjectWith(projectID)`; reject foreign UUIDs.
- Tests: auth login/facade TG deleted; `TestAddIssuesToSprint_RejectsForeignIssue_RT005` + handler route FK check.

## Session note — 2026-07-28 (be2 · A3-005 + RT-003)

- A3-005: StartTimer Exist→Create under tx + SELECT … FOR UPDATE on parent issue; concurrent Postgres test.
- RT-003: ListContacts / GetPipeline owner-scoped for members; system admin/founder org-wide (product decision).
- Tests: start_timer_concurrent_a3005_test.go, list_scope_rt003_test.go; FIX-W20 RT-003 → CLOSED.
- Docs: docs/wiki/API.md, docs/TECHNICAL.md CRM access notes.

## Session note тАФ 2026-07-28 (be1 ┬╖ RT-005 sibling close Get/Update/Retro)

- Scoped `GetSprint` / `UpdateSprint` / `UpdateRetroNotes` / `GetBurndown` / `AddIssuesToSprint` (+ poker start gate) via `GetSprintInProject(projectID, sprintID)`.
- Already scoped: `CompleteSprint`, `StartSprint` (`8e0c9f2`).
- IDOR hubtests: `Test*RejectsForeignProject_RT005` for each newly scoped service method.
- Schema unchanged.

## Session note тАФ 2026-07-28 (be2 ┬╖ A13-006 residual batch counts)

- `GetSprintIssueCounts` (Ent `GROUP BY sprint_issues`) + `ListSprints` preload тАФ no N├Ч `GetSprintIssueCount`.
- Tests: `TestListSprints_BatchIssueCounts_A13_006` (тЙд1тАУ2 batch, 0 single), `TestGetSprintIssueCounts_Ent_A13_006`.
- Docs: `.planning/audits/grok-2026-07/FIX-W14.md` residual close.

## Session note тАФ 2026-07-28 (be2 ┬╖ close W19 WEAK A6-006 / A7-007 / A7-008)

- A6-006: CancelLeave bind errors тЖТ 400; HTTP version=0 approve/cancel.
- A7-007: exact `pg_restore failed` / `mongorestore failed` without DSN/CombinedOutput.
- A7-008: broadcast dry-run excludes deactivated (not only pending).
- Docs: `docs/TESTING.md` TC-ADMIN-06/07, TC-HR-01; `.planning/audits/grok-2026-07/FIX-W19.md` WEAK close.

## Session note тАФ 2026-07-28 (docs-writer ┬╖ post-audit WEAK closure wave)

- Recorded remediation recheck in `PROGRAM-W4-W20-SUMMARY.md` + FIX-W9/W12/W15/W17/W18 (+ W6/W8 A2-007/010) + FIX-W20 residual scoreboard.
- Closed WEAK list with SHAs; schema A8-006 / A8-001/002 indexes still BLOCKED/NEEDS_APPROVAL; RT-003 ADR-blocked.
- RT-005 **CLOSED** on HEAD: Complete/Start/Get/Update/Retro/Burndown/AddIssues via `GetSprintInProject` (`1ac8669` / `8e0c9f2` / `716c823`).
- No coverage remasure invented (gate remains 81.9% 2026-07-27).

## Session note тАФ 2026-07-28 (qa-auto1 ┬╖ close W17 WEAK A14-006 / A14-007)

- A14-006: `TestMagicLink_OTPNotLogged_A14_006` тАФ slog spy proves raw OTP never logged on magic-link issue/verify/Resend-fail.
- A14-007: `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` + wiring тЖТ `FiberRecoverEnableStackTrace()`.
- Docs: `docs/TESTING.md` TC-AUTH-13 / TC-PLATFORM-01; `.planning/audits/grok-2026-07/FIX-W17.md` WEAK close.

## Session note тАФ 2026-07-28 (qa-auto2 ┬╖ close W18 WEAK A4-002 / A4-006)

- A4-002: `TestPostgresSearch_ExcludesSoftDeleted_A4_002` тАФ Postgres runtime soft-delete/pending exclusion (issues+files); source-grep twin retained.
- A4-006: `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006` тАФ HTTP re-confirm тЖТ 409; entity not rebound.
- Docs: `docs/TESTING.md` TC-SEARCH-05 / TC-STORAGE-04.

## Session note тАФ 2026-07-28 (devops ┬╖ A14-005 tighten + audit honesty)

- A14-005: prod compose `METRICS_ALLOWED_CIDRS` default `127.0.0.1/32` (was RFC1918); fail-closed empty unchanged (`c5613f6`).
- Honesty docs: FIX-W5 A5-003/006 SHA attribution; FIX-W10 A8-005 PARTIAL; FIX-W13 A10-008 FIXED via A13-014; FIX-W16 A14-005 FIXED*.
- No Ent schema approval for A8-006/indexes.

## Session note тАФ 2026-07-28 (qa-auto1 ┬╖ close WEAK A2-007 / A2-010)

- A2-007: `ws_authorizer_integration_a2007_test.go` тАФ prod `ws.NewAuthorizer` + calls repo ╤З╨╡╤А╨╡╨╖ `readPump`; stranger deny / participant allow.
- A2-010: `TestNotificationConsumer_NakOnDispatchError_A2_010` тАФ Nak when Dispatch fails (not only bad JSON).
- Mutation evidence: skip CanSubscribe тЖТ StrangerDenied FAIL; always-Ack тЖТ NakOnDispatchError FAIL.
- Docs: `docs/TESTING.md` TC-CALL-09 / TC-NOTIF-01.
- No push master.

## Session note тАФ 2026-07-27 (be1 ┬╖ W19 Block F handler polish)

- FIX: `.planning/audits/grok-2026-07/FIX-W19.md`
- Closed clear P1/P2: A7-002, A5-002, A6-005, A7-005, A7-008, A7-007/009, A6-006
- Domains sampled: crm/hr/accounting/calendar/github/gitlab/imex/marketplace/intranet/mail/dashboard/backup/apigateway/status/dev
- No schema ┬╖ no push master ┬╖ hubtest green on touched packages

## Session note тАФ 2026-07-27 (be2 ┬╖ W18 Block F handler polish)

- Domains: tms/tracker/sprint/calls/notification/search/bases/storage/ai (same checklist as W17).
- FIXED: A4-001/002/006/007, A2-008, board/label IDOR, A3-016/015/017 (1 commit each).
- SHAs: `27f0374` `31002d7` `06c1856` `3074dca` `adbc03e` `dbd03cf` `0a3a3ad` `bc5d61a` `369f852`.
- Report: `.planning/audits/grok-2026-07/FIX-W18.md`. Deferred: A4-005/008/009+, A2-011, A3-013/014/018тАУ020.

## Session note тАФ 2026-07-27 (be1 ┬╖ W17 Block F handler polish)

- Sampled auth/rbac/audit/platform/project/messenger/ws handlers vs checklist.
- FIXED: A1-009, A1-013, A2-009, A14-007, A14-006, A1-010 (1 commit each).
- SHAs: `7002ce6` `4884fb9` `ed3b772` `ec2e23f` `a619a04` `5a1858d`.
- Report: `.planning/audits/grok-2026-07/FIX-W17.md`. Residual: project GetProject non-member metadata GAP; A2-013/014 P3.

## Session note тАФ 2026-07-27 (be1 ┬╖ W14 Block D perf A13-001тАж006)

- A13-001тАж006 FIXED (hard caps / aggregates / batch links / prefs prefetch).
- SHAs: `456636f` `f961fec` `60149a0` `e76d5ef` `6d15338` `3b224e4`.
- Report: `.planning/audits/grok-2026-07/FIX-W14.md`. Indexes тЖТ W15.

## Session note тАФ 2026-07-27 (fe2 ┬╖ W15 perf A8/A13 lazy+virtual+waterfalls)

- A8-001/002: FK-column query rewrite FIXED; Ent composite indexes **NEEDS_APPROVAL**.
- A13-011/012 waterfalls, A13-014 Kanban virtual, A13-015тАУ017 lazy shell/i18n/routes, A13-018 channel virtual FIXED.
- A13-001 board cap already on branch (`456636f`).
- Report: `.planning/audits/grok-2026-07/FIX-W15.md`. Vitest 147 PASS.

## Session note тАФ 2026-07-27 (devops ┬╖ W16 Block E ops A14)

- A14-001/002/003/004/005/009 verified OPEN тЖТ FIXED (config/docs/wiring).
- MANUAL: rotate SeaweedFS if git defaults were live; set NATS creds on existing prod; enable OTEL endpoint for spans.
- Report: `.planning/audits/grok-2026-07/FIX-W16.md` (`4aa77f5`тАж`68cbd14`, report `e882016`).

## Session note тАФ 2026-07-27 (fe1 ┬╖ W13 A11-003 + UX/a11y)

- A11-003: deleted dead TelegramButton + Magic Link i18n (`5112ab2`).
- A11-004 regression test (`375783b`; code prior `7687f20`).
- A10-006 / A10-005 / A10-010 FIXED (`04d98f8` `bba4666` `ed8cff6`).
- Report: `.planning/audits/grok-2026-07/FIX-W13.md`. Vitest W13 suites 9/9 PASS.

## Session note тАФ 2026-07-27 (fe2 ┬╖ W12 FE A10-001тАж004)

- A10-001: WS channel events тЖТ React Query cache (`903c3b7`).
- A10-002: issues list `useInfiniteQuery` merge pages (`35ad363`).
- A10-003: WS reconnect backoff + timer cleanup (`9b5e1ad`).
- A10-004: presence key copy-before-sort (`257e3f6`).
- Report: `.planning/audits/grok-2026-07/FIX-W12.md`.
- Vitest: 20 pass (messenger + issues-cursor + ws-store-reconnect).

## Session note тАФ 2026-07-27 (fe1 ┬╖ W11 API contract A11-004тАж013)

- A11-004тАж013 FIXED (minimal FE): dead telegram/notify-mode clients removed; birthday RFC3339; TMS `getBlob`; IssueResponse attachments drop; CRM/bases/marketplace/dashboard templates/registration-mode FE wrappers.
- SHAs: `7687f20` `ebfc6b0` `29526f6` `8bb37ab` `0a4ddb7` `b3cf6df` `0d4e3ff` `35310c7` `d631197` `9b32eaf`.
- Report: `.planning/audits/grok-2026-07/FIX-W11.md`.
- Vitest: 72 pass in touched domains; residual FAIL `telegram-button-absent.test.ts` (A11-003 parallel).

## Session note тАФ 2026-07-27 (be2 ┬╖ W10 tx/data A8-005/007/008 + A7-003)

- A8-005: `DeleteChannel` tx-cascade children (`b77ba23`); soft-delete schema deferred.
- A8-006: **BLOCKED** тАФ Comment soft-delete needs Ent `deleted_at` approval.
- A8-007: Mongo `Disconnect` LIFO cleanup (`f3dbad4`).
- A8-008: Listen fail тЖТ `rootCancel`, not `os.Exit` (`23d6abb`).
- A7-003: real `MaintenanceGate` + 503 middleware (`0dd0dd7`).
- Report: `.planning/audits/grok-2026-07/FIX-W10.md`.

## Session note тАФ 2026-07-27 (be1 ┬╖ A7-001 dashboard project stats membership)

- `GetProjectStats(userID, projectID)`: `IsProjectMember` **╨┤╨╛** Redis cache/DB; ╨╕╨╜╨░╤З╨╡ `ErrForbidden`.
- Handler `Project`: ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `userID` (╨╜╨╡ `_`); 403 ╨╜╨░ `ErrForbidden`.
- ╨в╨╡╤Б╤В: `TestGetProjectStats_NonMemberForbidden` (+ mutation: shared cache leak).
- Docs: `docs/wiki/API.md` тАФ GET `/dashboard/project/:id` ╤В╨╛╨╗╤М╨║╨╛ ╤З╨╗╨╡╨╜ / 403.
- Commit: `c5edd70`.

## Session note тАФ 2026-07-27 (be1 ┬╖ A4-004 ListSavedFilters membership)

- `ListSavedFilters`: `IsProjectMember` **╨┤╨╛** `repo.ListSavedFilters`; ╨╕╨╜╨░╤З╨╡ `ErrFilterForbidden`.
- Handler: 403 ╨╜╨░ `ErrFilterForbidden` (╨╜╨╡ 500; ╨▒╨╡╨╖ ╤Г╤В╨╡╤З╨║╨╕ shared `filter_config`).
- ╨в╨╡╤Б╤В: `TestListSavedFilters_NonMemberForbidden` (+ mutation: shared leak).
- Docs: `docs/wiki/API.md` тАФ GET `/filters/` ╤В╨╛╨╗╤М╨║╨╛ ╤З╨╗╨╡╨╜ ╨┐╤А╨╛╨╡╨║╤В╨░ / 403.
- Commit: `4f9b2bb`.

## Session note тАФ 2026-07-27 (be1 ┬╖ A4-003 CreateSavedFilter membership)

- `CreateSavedFilter`: `IsProjectMember` **╨┤╨╛** `repo.CreateSavedFilter`; ╨╕╨╜╨░╤З╨╡ `ErrFilterForbidden`.
- Handler: 403 ╨╜╨░ `ErrFilterForbidden` (╨╜╨╡ 500).
- ╨в╨╡╤Б╤В: `TestCreateSavedFilter_NonMemberForbidden` (+ mutation: 0 filters).
- Docs: `docs/wiki/API.md` тАФ POST `/filters/` ╤В╨╛╨╗╤М╨║╨╛ ╤З╨╗╨╡╨╜ ╨┐╤А╨╛╨╡╨║╤В╨░ / 403.
- Commit: `5d2e590`.

## Session note тАФ 2026-07-27 (be2 ┬╖ A9-012 APIKey + RequireAuth)

- `RequireAuth`: ╨╡╤Б╨╗╨╕ `Locals("userID")` ╤Г╨╢╨╡ set (APIKeyAuth) тАФ skip JWT.
- ╨в╨╡╤Б╤В: `TestAPIKeyAuth_AlonePassesRequireAuth_A9012` (X-API-Key alone тЖТ 200).
- Mutation: force JWT always тЖТ 401 `missing authorization header`.
- Commit: `c7a4589` (message mislabeled as A1-005; tree is A9-012 only тАФ real A1-005 is `428865c`).

## Session note тАФ 2026-07-27 (be1 ┬╖ A2-002 CreateCall channel membership)

- `CreateCall`: `IsChannelMember` **╨┤╨╛** `provisionRoom`; ╨╕╨╜╨░╤З╨╡ `ErrCallForbidden`.
- Handler: 403 ╨╜╨░ `ErrCallForbidden` (╨╜╨╡ 500).
- ╨в╨╡╤Б╤В: `TestCreateCall_NonMemberForbidden` (+ mutation guard: 0 call rows).
- Docs: `docs/wiki/API.md` тАФ POST `/calls/` ╤В╨╛╨╗╤М╨║╨╛ ╤З╨╗╨╡╨╜ ╨║╨░╨╜╨░╨╗╨░ / 403.

## Session note тАФ 2026-07-27 (be2 ┬╖ A2-003 call chat authz)

- `CallChatService.authorizeCallChatAccess`: channel-call тАФ active participant
  OR `IsChannelMember`; instant тАФ participant OR `created_by`.
- Handler send/list: 403 ╨╜╨░ `ErrCallForbidden`; list ╨┐╨╡╤А╨╡╨┤╨░╤С╤В `userID`.
- ╨в╨╡╤Б╤В╤Л: `chat_authz_a2003_test.go`; mutation: ╤Б╨╜╤П╤В╤М authz тЖТ stranger PASS.
- Docs: `docs/wiki/API.md` chat endpoints.

## Session note тАФ 2026-07-27 (be2 ┬╖ A3-011-target CreateIssueLink)

- `FindIssueInProject` (Ent: `IDEQ` + `HasProjectWith` + `DeletedAtIsNil`) +
  `CreateIssueLink` ╨╛╤В╨║╨╗╨╛╨╜╤П╨╡╤В target ╤З╤Г╨╢╨╛╨│╨╛ ╨┐╤А╨╛╨╡╨║╤В╨░ ╤З╨╡╤А╨╡╨╖ NotFound
  (┬л╤Ж╨╡╨╗╨╡╨▓╨░╤П ╨╖╨░╨┤╨░╤З╨░ ╨╜╨╡ ╨╜╨░╨╣╨┤╨╡╨╜╨░┬╗), ╨▒╨╡╨╖ ╤Г╤В╨╡╤З╨║╨╕ ╤Б╤Г╤Й╨╡╤Б╤В╨▓╨╛╨▓╨░╨╜╨╕╤П.
- ╨в╨╡╤Б╤В╤Л: `link_target_scope_a3011_test.go` (Ent + service); mutation:
  ╤Б╨╜╤П╤В╤М `HasProjectWith` тЖТ Ent FAIL. Tracker regression green (~18.5s).
- DeleteIssueLink ╨╜╨╡ ╤В╤А╨╛╨│╨░╨╗╨╕ (╤Г╨╢╨╡ scoped).

## Session note тАФ 2026-07-26 (PE ┬╖ ╨╕╨╜╤Б╤В╤А╤Г╨║╤Ж╨╕╨╕ ╤В╨╛╨╗╤М╨║╨╛ ╨╜╨░ ╤А╤Г╤Б╤Б╨║╨╛╨╝)

- ╨Я╨╛╨╗╨╕╤В╨╕╨║╨░: ╤З╨╡╨╗╨╛╨▓╨╡╨║╨╛╤З╨╕╤В╨░╨╡╨╝╤Л╨╡ ╨╕╨╜╤Б╤В╤А╤Г╨║╤Ж╨╕╨╕ `prompt-engineering/` тАФ **╤В╨╛╨╗╤М╨║╨╛ RU**
  (`rules/00` ┬з0, README, START-HERE); id/╨┐╤Г╤В╨╕/CLI/enum ╨╜╨╡ ╨┐╨╡╤А╨╡╨▓╨╛╨┤╤П╤В╤Б╤П.
- ╨Я╨╡╤А╨╡╨▓╨╡╨┤╨╡╨╜╤Л skills, recipes, rules/10тАУ30, adapters, patterns, review-feedback,
  docs/evals fixtures, MANIFEST/VERSION, context/templates; machine headers
  (packet Outcome/Allowed, pattern Bad code/Good code) ╤Б╨╛╤Е╤А╨░╨╜╨╡╨╜╤Л ╨┐╨╛╨┤ builder/schema.
- ╨Э╨╡ ╤В╤А╨╛╨│╨░╨╗╨╕: `context/projects/hub/**`, `evals/smoke/**`, `scripts/**`, LICENSE.
- Validator source+package PASS ╨┐╨╛╤Б╨╗╨╡ ╨┐╨╡╤А╨╡╨▓╨╛╨┤╨░. ╨Ъ╨╛╨╝╨╝╨╕╤В ╤В╨╛╨╗╤М╨║╨╛ ╨┐╨╛ ╨╖╨░╨┐╤А╨╛╤Б╤Г.

## Session note тАФ 2026-07-26 (AI quality layer ┬╖ UX + package harden)

- ╨Ч╨░╨║╤А╤Л╤В╤Л P0 onboarding contradictions ╨╕╨╖ UX-╤А╨╡╨▓╤М╤О: bootstrap ╤В╨╡╨┐╨╡╤А╤М ╨▓╨╡╨╖╨┤╨╡ ╤Б
  ╨╛╨▒╤П╨╖╨░╤В╨╡╨╗╤М╨╜╤Л╨╝ `inventory=` + `context_inventory.py`; local path = ╤В╨╛╨╗╤М╨║╨╛
  `build_task_packet.py` (╨╜╨╡ hand-fill template); RoleтЖФrecipe cheat sheet;
  local dirs `local-task-inputs/`, `generated-task-packets/`,
  `local-approval-records/` + sample approval.
- Packager ╨▓╨║╨╗╤О╤З╨░╨╡╤В `build_task_packet.py` + config/schema; package **172**
  files, source + packaged validators **0/0**, tests **55 PASS / 1 SKIP**.
- Stale claim ┬л`.cursor` validator 0/0 via cursor_validation_manifest┬╗ ╤Б╨╜╤П╤В:
  ╤В╨░╨║╨╛╨│╨╛ ╨╝╨░╨╜╨╕╤Д╨╡╤Б╤В╨░ ╨╜╨╡╤В; Hub `.cursor` ╨╜╨╡ ╨▓╨░╨╗╨╕╨┤╨╕╤А╤Г╨╡╤В╤Б╤П PE-manifest'╨╛╨╝.
- START-HERE / INSTALL / PORTABLE-HUB-KIT / LOCAL-AGENT-GUIDE / generic-chat
  README / onboarding playbook ╨▓╤Л╤А╨╛╨▓╨╜╨╡╨╜╤Л.
- Effectiveness/ROI ╨┐╨╛-╨┐╤А╨╡╨╢╨╜╨╡╨╝╤Г **NOT_RUN**. Next: empty-repo physical install
  dry-run тЖТ paired baseline-vs-assisted pilot. ╨Ъ╨╛╨╝╨╝╨╕╤В ╤В╨╛╨╗╤М╨║╨╛ ╨┐╨╛ ╨╖╨░╨┐╤А╨╛╤Б╤Г.

## Session note тАФ 2026-07-26 (Test Automation AI Quality Layer)

- `prompt-engineering/` ╨┐╨╡╤А╨╡╤А╨░╨▒╨╛╤В╨░╨╜ ╨╕╨╖ ╨║╨░╤В╨░╨╗╨╛╨│╨░ bare prompts ╨▓ ╨┐╨╡╤А╨╡╨╜╨╛╤Б╨╕╨╝╤Г╤О
  AI-╨┐╤А╨╛╤Б╨╗╨╛╨╣╨║╤Г ╨░╨▓╤В╨╛╨╝╨░╤В╨╕╨╖╨░╤В╨╛╤А╨╛╨▓: generated context, approved patterns,
  production recipes, agents/skills, adapters/model router, validator/evals.
- `/bootstrap-test-context` ╨┐╤А╨╕╨╜╨╕╨╝╨░╨╡╤В `project_root`, `test_roots`, `inventory`,
  `profile`; read-only, ╨╜╨╡ ╤З╨╕╤В╨░╨╡╤В ╨╖╨░╨║╤А╤Л╤В╤Л╨╡ MR/╤Б╨╡╤В╤М/secrets ╨╕ ╨╜╨╡ ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В
  ╨╜╨░╨╣╨┤╨╡╨╜╨╜╤Л╨╡ ╨║╨╛╨╝╨░╨╜╨┤╤Л.
- ╨Я╤А╨╕╨╝╨╡╤А╤Л ╨║╨╛╨╗╨╗╨╡╨│ ╨┤╨╛╨▒╨░╨▓╨╗╤П╨╡╤В **╤З╨╡╨╗╨╛╨▓╨╡╨║ ╨┐╨╛╤Б╨╗╨╡ bootstrap** ╨▓
  `review-feedback/inbox/`: bad code + comment + good code; curation ╤В╤А╨╡╨▒╤Г╨╡╤В
  redaction/provenance/reuse scope/human approval.
- Hub profile `context/projects/hub/`: 9 ╨░╤А╤В╨╡╤Д╨░╨║╤В╨╛╨▓, 49 file-source hashes,
  8 output hashes, snapshot `d4a44174тАж`, manifest `complete`; project commands
  ╨╛╤Б╤В╨░╤О╤В╤Б╤П `not-run`.
- Prompt contract: natural-language intake, outcome/evidence/completion bar,
  untrusted-data boundary; explicit chain-of-thought disclosure ╨╕ fake
  temperature controls ╤Г╨┤╨░╨╗╨╡╨╜╤Л.
- Model routing applied: `team-lead` = GPT-5.6 Sol Pro
  (`model: gpt-5.6-sol`), ╨╛╤Б╤В╨░╨╗╤М╨╜╤Л╨╡ 31 agents = `inherit`; downstream ╨▓╤Л╨▒╨╕╤А╨░╨╡╤В╤Б╤П
  ╨┐╨╛ ╤В╨╡╨║╤Г╤Й╨╡╨╝╤Г Cursor catalog ╨╕ ╤Д╨╕╨║╤Б╨╕╤А╤Г╨╡╤В╤Б╤П ╨▓ handoff.
- GSD-RU skill names ╨┐╤А╨╕╨▓╨╡╨┤╨╡╨╜╤Л ╨║ parent folders; phantom routes = 0. Repo-scan
  ┬л40├Ч Grok┬╗ ╨╕╤Б╨┐╤А╨░╨▓╨╗╨╡╨╜ ╨╜╨░ 40 report slots ╤Б provenance/confidence.
- Assisted Hub grounding smoke: content grounding PASS, ╨╛╨▒╤Й╨╕╨╣ verdict
  **PARTIAL** (runtime trace self-reported); paired baseline/effectiveness/ROI
  ╨╛╤Б╤В╨░╤О╤В╤Б╤П **NOT_RUN**.
- Management pack: RU/EN executive overview, governance, ROI metrics ╨╕
  15-╨╝╨╕╨╜╤Г╤В╨╜╨╛╨╡ demo. Toolkit version `0.1.0-alpha.1`, pre-release.

## ╨б╨┐╤А╨░╨▓╨║╨░

╨б╨╝.: `.planning/PROJECT.md` ┬╖ ╨╕╤Б╤В╨╛╤А╨╕╤П session notes: [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](history/STATE-HISTORY-THROUGH-2026-07-23.md)

**╨Ъ╨╗╤О╤З╨╡╨▓╨░╤П ╤Ж╨╡╨╜╨╜╨╛╤Б╤В╤М:** ╨╡╨┤╨╕╨╜╨░╤П self-hosted ╤Б╤А╨╡╨┤╨░ ╨┤╨╗╤П dev-╨║╨╛╨╝╨░╨╜╨┤╤Л тАФ ╤В╤А╨╡╨║╨╡╤А, TMS, ╨╝╨╡╤Б╤Б╨╡╨╜╨┤╨╢╨╡╤А, ╨╖╨▓╨╛╨╜╨║╨╕, ╨╕╨╜╤В╤А╨░╨╜╨╡╤В, CRM, HR, ╤Г╤З╤С╤В, AI тАФ ╨▓ ╨╛╨┤╨╜╨╛╨╝ ╨╕╨╜╤В╨╡╤А╤Д╨╡╨╣╤Б╨╡.

**╨в╨╡╨║╤Г╤Й╨╕╨╣ ╤Д╨╛╨║╤Г╤Б (2026-07-26):** Test Automation AI Quality Layer `0.1.0-alpha.1` тАФ
structural/package PASS ┬╖ instructions RU-only ┬╖ paired eval/ROI `NOT_RUN` ┬╖
July Rebuild product-code freeze ┬╖ LiveKit ╨║╨░╨╜╨╛╨╜ ┬╖ FA-06 manual.
╨Ш╨╜╨┤╨╡╨║╤Б ╨╕╤О╨╗╤П: `.planning/july-2026/INDEX-JULY-2026.md`.

## ╨в╨╡╨║╤Г╤Й╨░╤П ╨┐╨╛╨╖╨╕╤Ж╨╕╤П (2026-07-26)

| ╨Я╨░╤А╨░╨╝╨╡╤В╤А | ╨Ч╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|----------|----------|
| ╨д╨░╨╖╨░ | **July Rebuild planning** ┬╖ QA-HARDENING docs wave ┬╖ ╨║╨╛╨┤ freeze ╨┤╨╛ ~╨║╨╛╨╜╤Ж╨░ ╨╕╤О╨╗╤П |
| Baseline quality | Wave C FA-01тАУFA-12 тЬЕ (╨║╤А╨╛╨╝╨╡ FA-06 manual) ┬╖ RU-V10 PASS ┬╖ FA-10 qa-gate PASS |
| ╨Т╨╛╨╗╨╜╨░ 1 | тЬЕ merged to `master` ╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╨╡╨╗╨╡╨╝ 2026-07-11 |
| SFU | **LiveKit** (SEED-001 mediasoup **REJECTED**) |
| ╨Т╨╡╤В╨║╨░ ╨║╨░╨╜╨╛╨╜╨░ ╨╝╨╡╤В╤А╨╕╨║ | `master` (COV/product тАФ ╨║╨░╨║ ╨▓ AGENTS.md) |
| COV overall (`internal/`) | **83.0%** (2026-06-22 Docker baseline, `-tags hubtest`) |
| Security packages тЙе80% | **6/6** тЬЕ |
| Product score | **89/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Backend tests | **515** files ╨▓ `backend/testframework/` (measured 2026-07-27; was 499 on 2026-07-26) ┬╖ **0** ╨▓ `internal/*_test.go` |
| Playwright E2E | **20** specs (`testframework/stages/e2e/domains/`) |
| Frontend Vitest | **99/99** pass ┬╖ 37 files `testframework/stages/unit/frontend-test/domains/` ┬╖ 0 in `src/` |
| Planning index | `.planning/july-2026/INDEX-JULY-2026.md` |

### ╨Я╨╛╨║╤А╤Л╤В╨╕╨╡ ╨║╨╗╤О╤З╨╡╨▓╤Л╤Е ╨┐╨░╨║╨╡╤В╨╛╨▓ (2026-06-22 Docker baseline, Phase 125 / FA-11)

| ╨Я╨░╨║╨╡╤В | Coverage |
|-------|----------|
| notification | **82.5%** (2026-06-22) |
| auth | **82.4%** (2026-06-22) |
| tracker | **82.4%** (2026-06-22) |
| messenger | **82.3%** (2026-06-22) |
| ws | **87.7%** (2026-06-22) |
| storage | **86.8%** (2026-06-22) |
| rbac | **84.7%** (2026-06-22) |
| audit | **80.5%** (2026-06-22) |
| tms | **82.0%** (2026-06-22; FA-11 тЬЕ micro-gap closed) |
| project | **82.0%** (2026-06-22; FA-11 тЬЕ micro-gap closed) |

---

## Milestones тАФ ╤Б╤В╨░╤В╤Г╤Б

| Milestone | Phases | ╨б╤В╨░╤В╤Г╤Б | ╨Ф╨░╤В╨░ |
|-----------|--------|--------|------|
| v1.0 MVP | 1тАУ7 | тЬЕ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | 8тАУ14 | тЬЕ COMPLETE | 2026-05-15 |
| v2.5 Trust | 15 | тЬЕ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | 16тАУ38 | тЬЕ COMPLETE | 2026-06-24 |
| v3.0 AI Native | 17тАУ19 | тЬЕ COMPLETE | 2026-06-22 |
| v4 Super-app | 39тАУ58 | тЬЕ COMPLETE | 2026-06-25 |
| Coverage wave | 59тАУ125 | тЬЕ COMPLETE | 2026-06-21 |
| Test framework migration | TF-M01тАУM10 | тЬЕ Wave B + pipeline domains (TF-M10) | 2026-06-23 |
| Final audit | FA-01тАУFA-12 | тЬЕ Wave C COMPLETE (╨║╤А╨╛╨╝╨╡ FA-06 manual) | 2026-06-22 |

╨Р╤А╤Е╨╕╨▓ v4: `.planning/milestones/v4-super-app/SUMMARY.md`

---

## PO Decisions (locked)

╨б╨╝.: `.planning/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

- **AUTH:** Magic Link only ┬╖ org table v2.5 ┬╖ founder role ┬╖ Telegram ╤В╨╛╨╗╤М╨║╨╛ bind/delivery
- **NOTIF:** TG bind ┬╖ hub-mail Phase 15.2 ┬╖ email marketing v2.6
- **Phase 15 exit:** AUTH+NOTIF+E2E+WCAG+run_failed+Biome+LiveKit CI ┬╖ coverage тЙе50% тЬЕ
- **TZ rewrite:** Gantt, webcam, marketing, no-code, accounting тАФ IN (by phase)
- **Coverage target:** тЙе82% overall ┬╖ security 6/6 тЙе80% тЬЕ

---

## Test Framework / Wave C тАФ live summary

- **TF Wave B:** COMPLETE тАФ backend tests ╤В╨╛╨╗╤М╨║╨╛ ╨▓ `testframework/` (**515** files,
  measured 2026-07-27; was 499 on 2026-07-26 / earlier migration note had 495) ┬╖ `internal/` = **0** ┬╖
  Vitest **99/99** ┬╖ E2E **20** specs ┬╖ CI `-tags hubtest`.
- **Wave C FA-01тАУFA-12:** COMPLETE ╨║╤А╨╛╨╝╨╡ **FA-06** (PG-03 secret rotation, manual).
  FA-10 qa-gate PASS. ╨Ф╨╡╤В╨░╨╗╨╕: `.planning/tasks/TASKS-PHASES-121-FINAL-AUDIT.md`,
  ╨░╤Г╨┤╨╕╤В╤Л `.planning/audits/2026-06/`. ╨Я╨╛╨╗╨╜╤Л╨╡ ╤В╨░╨▒╨╗╨╕╤Ж╤Л ╨╖╨░╨║╤А╤Л╤В╨╕╤П тАФ ╨▓
  [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](history/STATE-HISTORY-THROUGH-2026-07-23.md).

---

## Pending Todos

- **AI quality layer:** empty-repo install dry-run тЖТ paired eval pilot; ROI ╨╛╤Б╤В╨░╤С╤В╤Б╤П `NOT_RUN` ╨┤╨╛ ╨╖╨░╨╝╨╡╤А╨░
- **PE language:** ╨╕╨╜╤Б╤В╤А╤Г╨║╤Ж╨╕╨╕ `prompt-engineering/` ╤В╨╛╨╗╤М╨║╨╛ RU (╨╝╨░╤И╨╕╨╜╨╜╤Л╨╡ id/enum ╨▒╨╡╨╖ ╨┐╨╡╤А╨╡╨▓╨╛╨┤╨░)
- **FA-06 / PG-03:** Phase 0 тЬЕ ┬╖ scripts ready ┬╖ **Phase 1тАУ3 manual** (╨╡╨┤╨╕╨╜╤Б╤В╨▓╨╡╨╜╨╜╤Л╨╣ prod-╨▒╨╗╨╛╨║╨╡╤А ops) ┬╖ runbook `docs/runbooks/fa06-secret-rotation.md`
- **Wave 1:** merged ╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╨╡╨╗╨╡╨╝ 2026-07-11; policy ╨╛╤Б╤В╨░╤С╤В╤Б╤П тАФ ╨░╨│╨╡╨╜╤В╤Л ╨╜╨╡ ╨┐╤Г╤И╨░╤В/╨╜╨╡ ╨╝╨╡╤А╨╢╨░╤В master
- **July return kickoff (~╨║╨╛╨╜╨╡╤Ж ╨╕╤О╨╗╤П):** align ╨┐╨╛ brief тЖТ ╨┐╨╡╤А╨╡╤Б╤В╤А╨╛╨╣╨║╨░ ╨┐╨╗╨░╨╜╨╛╨▓ ┬╖ ╤Б╤В╨░╤А╤В EPIC-CAL-SYNC / QA-HARDENING execution / Redesign ╨┐╨╛ ╨┐╤А╨╕╨╛╤А╨╕╤В╨╡╤В╤Г
- **CAL-SYNC:** SEED-002 active тАФ unified calendar + RSVP + mail levels L0тАУL3 (ADR ╨┐╤А╨╕ L3/OAuth)
- **QA-HARDENING:** Testing Guide + EDGE-CORNER-CATALOG + CODE-QUALITY-CHARTER тЖТ TC тЖТ ╨░╨▓╤В╨╛╤В╨╡╤Б╤В╤Л (╨╕╤Б╨┐╨╛╨╗╨╜╨╡╨╜╨╕╨╡ ╤Б ╨▓╨╛╨╖╨▓╤А╨░╤В╨░)
- **Redesign:** Cloud Code Designer / UX pass тАФ ╨┐╨╛╤Б╨╗╨╡ ╨╕╨╗╨╕ ╨┐╨░╤А╨░╨╗╨╗╨╡╨╗╤М╨╜╨╛ kickoff (brief)
- **SSH deploy + v2.5.0 tag:** after FA-06 Phase 1тАУ3 + CTO sign-off
- ~~**UX-1 / Dependabot / CI flake:**~~ тЬЕ closed 2026-06-22

---

## ╨С╨╗╨╛╨║╨╡╤А╤Л / ╤А╨╕╤Б╨║╨╕

| ╨а╨╕╤Б╨║ | Mitigation |
|------|------------|
| ~~RU-V gaps (EN comments, encoding)~~ | тЬЕ RU-V10 PASS 2026-06-22 |
| ~~tms/project тИТ0.2% gap~~ | тЬЕ FA-11 closed (82.0%) |
| ~~TMS CI webhook ╨▒╨╡╨╖ RBAC (FA-02-02)~~ | тЬЕ FA-RBAC-FIBER3 + APIGW-01 2026-06-22 |
| ~~Fiber v3 RBAC middleware order~~ | тЬЕ FA-RBAC-FIBER3 123 routes 2026-06-22 |
| PG-03 not executed | FA-06 manual ops before prod cutover |
| ~~Uncommitted v4 block on branch~~ | тЬЕ PR #1 squash merged 2026-06-22 |
| ~~main.go ~629 LOC tech debt~~ | тЬЕ FA-08 wiring split 2026-06-22 |

---

## Session Continuity

**Last session:** 2026-07-26 (AI quality layer ┬╖ RU instructions ┬╖ STATE compact)
**Stopped at:** structural/package validators PASS ┬╖ tests 55 PASS / 1 SKIP ┬╖
package 172 files ┬╖ smoke PARTIAL ┬╖ paired eval/ROI NOT_RUN
**Resume:** empty-repo portable install dry-run тЖТ paired baseline-vs-assisted
pilot; FA-06 manual; ╨╖╨░╤В╨╡╨╝ CAL-SYNC / QA-HARDENING / Redesign ╨┐╨╛ ╨┐╤А╨╕╨╛╤А╨╕╤В╨╡╤В╤Г.

╨Ш╤Б╤В╨╛╤А╨╕╤З╨╡╤Б╨║╨╕╨╡ session notes (тЙд2026-07-23) ╨╕ ╤Б╤В╨░╤А╤Л╨╡ completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](history/STATE-HISTORY-THROUGH-2026-07-23.md)

---

*╨Ю╨▒╨╜╨╛╨▓╨╗╤П╨╡╤В╤Б╤П ╨┐╨╛╤Б╨╗╨╡ ╨║╨░╨╢╨┤╨╛╨╣ ╤Д╨░╨╖╤Л ╨╕ `/handoff-sync`*
