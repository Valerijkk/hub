# W10 Assignments — routes + modules

> Стадия A · read-only · TASK-008

## A1 — routes+modules part 1 (29)
Отчёт: `A10-1-fe-routes-modules.md`

- `frontend/src/modules/accounting/accounting-page.tsx`
- `frontend/src/modules/admin/backups.tsx`
- `frontend/src/modules/bases/bases-page.tsx`
- `frontend/src/modules/calendar/unified-calendar.tsx`
- `frontend/src/modules/calls/calendar-agenda-view.tsx`
- `frontend/src/modules/calls/calendar-day-view.tsx`
- `frontend/src/modules/calls/calendar-month-view.tsx`
- `frontend/src/modules/calls/calendar-types.ts`
- `frontend/src/modules/calls/calendar-week-view.tsx`
- `frontend/src/modules/calls/call-calendar.tsx`
- `frontend/src/modules/calls/create-call-dialog.tsx`
- `frontend/src/modules/calls/recording-controls.tsx`
- `frontend/src/modules/calls/recordings-list.tsx`
- `frontend/src/modules/calls/scheduled-calls.tsx`
- `frontend/src/modules/crm/crm-page.tsx`
- `frontend/src/modules/crm/deals-pipeline.tsx`
- `frontend/src/modules/hr/leave-requests.tsx`
- `frontend/src/modules/intranet/intranet-hub.tsx`
- `frontend/src/modules/intranet/kudos-feed.tsx`
- `frontend/src/modules/intranet/news-feed.tsx`
- `frontend/src/modules/intranet/org-directory.tsx`
- `frontend/src/modules/intranet/random-coffee.tsx`
- `frontend/src/modules/marketplace/marketplace-page.tsx`
- `frontend/src/modules/messenger/link-preview.tsx`
- `frontend/src/modules/messenger/message-bubble.tsx`
- `frontend/src/modules/settings/ai-privacy.tsx`
- `frontend/src/modules/settings/api-keys.tsx`
- `frontend/src/modules/settings/appearance.tsx`
- `frontend/src/modules/settings/project-templates.tsx`

## A2 — routes+modules part 2 (29)
Отчёт: `A10-2-fe-routes-modules.md`

- `frontend/src/modules/settings/webhooks.tsx`
- `frontend/src/modules/status/status-page.tsx`
- `frontend/src/modules/tms/autotest-setup.tsx`
- `frontend/src/modules/tms/components/case-filters.tsx`
- `frontend/src/modules/tms/components/case-table-row.tsx`
- `frontend/src/modules/tms/components/case-tree.tsx`
- `frontend/src/modules/tms/components/coverage-report.tsx`
- `frontend/src/modules/tms/components/create-case-dialog.tsx`
- `frontend/src/modules/tms/components/environment-comparison.tsx`
- `frontend/src/modules/tms/components/flaky-tests-report.tsx`
- `frontend/src/modules/tms/components/import-dialog.tsx`
- `frontend/src/modules/tms/components/run-execution-panel.tsx`
- `frontend/src/modules/tms/components/run-header.tsx`
- `frontend/src/modules/tms/components/run-results-table.tsx`
- `frontend/src/modules/tms/components/step-editor.tsx`
- `frontend/src/modules/tms/components/test-case-metadata.tsx`
- `frontend/src/modules/tms/components/traceability-matrix.tsx`
- `frontend/src/modules/tms/components/trends-chart.tsx`
- `frontend/src/modules/tms/components/version-history.tsx`
- `frontend/src/modules/tms/environments.tsx`
- `frontend/src/modules/tms/trigger-ci-dialog.tsx`
- `frontend/src/modules/tracker/burndown-chart.tsx`
- `frontend/src/modules/tracker/gantt-chart.tsx`
- `frontend/src/modules/tracker/github-links.tsx`
- `frontend/src/modules/tracker/gitlab-pipeline-status.tsx`
- `frontend/src/modules/tracker/import-export.tsx`
- `frontend/src/modules/tracker/issue-description.tsx`
- `frontend/src/modules/tracker/issue-header.tsx`
- `frontend/src/modules/tracker/issue-message-links-panel.tsx`

## A3 — routes+modules part 3 (29)
Отчёт: `A10-3-fe-routes-modules.md`

- `frontend/src/modules/tracker/issues-cursor.ts`
- `frontend/src/modules/tracker/issue-sidebar-meta.tsx`
- `frontend/src/modules/tracker/pipeline-status.tsx`
- `frontend/src/modules/tracker/planning-poker.tsx`
- `frontend/src/modules/tracker/sprint-backlog.tsx`
- `frontend/src/modules/tracker/sprint-board.tsx`
- `frontend/src/modules/tracker/sprint-retro.tsx`
- `frontend/src/modules/tracker/triage-inbox.tsx`
- `frontend/src/modules/tracker/velocity-chart.tsx`
- `frontend/src/routes/__root.tsx`
- `frontend/src/routes/_authenticated.tsx`
- `frontend/src/routes/_authenticated/accounting.lazy.tsx`
- `frontend/src/routes/_authenticated/accounting.tsx`
- `frontend/src/routes/_authenticated/admin/audit.tsx`
- `frontend/src/routes/_authenticated/admin/dashboard.tsx`
- `frontend/src/routes/_authenticated/admin/limits.tsx`
- `frontend/src/routes/_authenticated/admin/mail.tsx`
- `frontend/src/routes/_authenticated/admin/trash.tsx`
- `frontend/src/routes/_authenticated/admin/users.tsx`
- `frontend/src/routes/_authenticated/bases.tsx`
- `frontend/src/routes/_authenticated/calendar.tsx`
- `frontend/src/routes/_authenticated/calls.tsx`
- `frontend/src/routes/_authenticated/crm.lazy.tsx`
- `frontend/src/routes/_authenticated/crm.tsx`
- `frontend/src/routes/_authenticated/dashboard.lazy.tsx`
- `frontend/src/routes/_authenticated/dashboard.tsx`
- `frontend/src/routes/_authenticated/help.tsx`
- `frontend/src/routes/_authenticated/hr.tsx`
- `frontend/src/routes/_authenticated/intranet/directory.tsx`

## A4 — routes+modules part 4 (29)
Отчёт: `A10-4-fe-routes-modules.md`

- `frontend/src/routes/_authenticated/intranet/index.tsx`
- `frontend/src/routes/_authenticated/marketplace.lazy.tsx`
- `frontend/src/routes/_authenticated/marketplace.tsx`
- `frontend/src/routes/_authenticated/messenger.tsx`
- `frontend/src/routes/_authenticated/messenger/$channelId.tsx`
- `frontend/src/routes/_authenticated/messenger/index.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/backlog.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/board/index.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/dashboard.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/gantt.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/gantt.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/index.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/issues/index.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/issues/new.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/autotest.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.tsx`

## A5 — routes+modules part 5 (28)
Отчёт: `A10-5-fe-routes-modules.md`

- `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/environments.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/flaky.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.lazy.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/tms/traceability.tsx`
- `frontend/src/routes/_authenticated/projects/$projectId/triage.tsx`
- `frontend/src/routes/_authenticated/projects/index.tsx`
- `frontend/src/routes/_authenticated/projects/new.tsx`
- `frontend/src/routes/_authenticated/settings.tsx`
- `frontend/src/routes/_authenticated/settings/ai.tsx`
- `frontend/src/routes/_authenticated/settings/appearance.tsx`
- `frontend/src/routes/_authenticated/settings/hotkeys.tsx`
- `frontend/src/routes/_authenticated/settings/integrations.tsx`
- `frontend/src/routes/_authenticated/settings/notifications.tsx`
- `frontend/src/routes/_authenticated/settings/profile.lazy.tsx`
- `frontend/src/routes/_authenticated/settings/profile.tsx`
- `frontend/src/routes/_authenticated/settings/security.tsx`
- `frontend/src/routes/_authenticated/trash.tsx`
- `frontend/src/routes/call.$code.lazy.tsx`
- `frontend/src/routes/call.$code.tsx`
- `frontend/src/routes/index.tsx`
- `frontend/src/routes/login.tsx`
- `frontend/src/routes/onboarding.lazy.tsx`
- `frontend/src/routes/onboarding.tsx`
- `frontend/src/routes/status.tsx`

**Σ файлов: 144**
