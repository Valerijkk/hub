# Phase 130 — TF E2E mechanical specs (15)

> Целевая раскладка и недельный порядок —
> [`active/august-2026/05-AUGUST-DETAILED.md`](../../active/august-2026/05-AUGUST-DETAILED.md).
> Юниты живут рядом с кодом, всё остальное — в корневой `testframework/`
> (решение владельца 2026-07-30).


**Status:** PLANNED  
**Migration stage:** Э2 (mechanical)  
**Estimate:** 4–5 person-days  
**Branch (suggested):** `feat/qa/tf-e2e-mechanical`

## Goal

Перенести **15 mechanical** Playwright specs TS → Python с паритетом сценариев.
Матрица: [`TF-MIGRATE-E3-PLAYWRIGHT.md`](../../audits/planning-audit-2026-07/TF-MIGRATE-E3-PLAYWRIGHT.md)

## Depends on

- **129** tf-e2e-shared

## Specs in scope

| Spec | Target path |
|------|-------------|
| `auth/login.spec.ts` | `tests/e2e/domains/auth/test_login.py` |
| `platform/health.spec.ts` | `tests/e2e/domains/platform/test_health.py` |
| `platform/metrics.spec.ts` | `…/test_metrics.py` |
| `platform/notifications.spec.ts` | `…/test_notifications.py` |
| `platform/critical-gate.spec.ts` | `…/test_critical_gate.py` |
| `admin/admin-ai.spec.ts` | `tests/e2e/domains/admin/test_admin_ai.py` |
| `admin/admin-audit.spec.ts` | `…/test_admin_audit.py` |
| `admin/admin-metrics.spec.ts` | `…/test_admin_metrics.py` |
| `ai/automation.spec.ts` | `tests/e2e/domains/ai/test_automation.py` |
| `messenger/messenger.spec.ts` | `tests/e2e/domains/messenger/test_messenger.py` |
| `navigation/navigation.spec.ts` | `tests/e2e/domains/navigation/test_navigation.py` |
| `navigation/shortcuts.spec.ts` | `…/test_shortcuts.py` |
| `navigation/command-palette.spec.ts` | `…/test_command_palette.py` |
| `tms/tms.spec.ts` | `tests/e2e/domains/tms/test_tms.py` |
| `v4/v4-superapp.spec.ts` | `tests/e2e/domains/v4/test_v4_superapp.py` |

## Deliverables

- [ ] Все 15 файлов с теми же test intents
- [ ] Marker `smoke` на login + critical-gate (a11y — в 131)
- [ ] Parity checklist в PR description

## Out of scope

- rewrite specs (logout, onboarding, a11y, ai-hub, project-lifecycle)
- Удаление TS originals

## Success criteria

- [ ] `pytest -m "e2e and mechanical"` (или path filter) green vs local stack
- [ ] Smoke paths 1–5 из critical-gate PASS

## Risks

| Risk | Mitigation |
|------|------------|
| Seeded DEMO projectId | Shared fixture from seed response / UI parse |
| Locale-dependent roles/names | Keep RU locators as in TS |
