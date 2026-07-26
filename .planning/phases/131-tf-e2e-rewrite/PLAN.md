# Phase 131 — TF E2E rewrite specs (5)

**Status:** PLANNED  
**Migration stage:** Э2 (rewrite)  
**Estimate:** 4–5 person-days  
**Branch (suggested):** `feat/qa/tf-e2e-rewrite`

## Goal

Переписать **5 non-mechanical** specs: auth waits/session, axe fixtures, AI async waits, tracker multi-step lifecycle.
Матрица: [`TF-MIGRATE-E3-PLAYWRIGHT.md`](../../planning-audit-2026-07/TF-MIGRATE-E3-PLAYWRIGHT.md)

## Depends on

- **129** tf-e2e-shared
- **130** tf-e2e-mechanical (желательно — smoke уже зелёный; soft)

## Specs in scope

| Spec | Drivers | Target |
|------|---------|--------|
| `auth/logout.spec.ts` | waits / network | `test_logout.py` |
| `auth/onboarding-profile.spec.ts` | auth session mutate | `test_onboarding_profile.py` |
| `platform/a11y.spec.ts` | fixtures (axe) | `test_a11y.py` — **smoke** |
| `ai/ai-hub.spec.ts` | auth + waits | `test_ai_hub.py` |
| `tracker/project-lifecycle.spec.ts` | waits / multi-step | `test_project_lifecycle.py` |

## Deliverables

- [ ] 5 Python modules с паритетом
- [ ] a11y в `smoke` marker
- [ ] Явные timeout policy (no bare sleep)
- [ ] Документ known flakes (ai-hub bot 45s)

## Out of scope

- Удаление TS tree (phase 136)
- Locust

## Success criteria

- [ ] Full Python e2e 20/20 parity with TS intents
- [ ] Smoke: login + a11y + critical-gate PASS
- [ ] TS suite ещё может жить параллельно

## Risks

| Risk | Mitigation |
|------|------------|
| ai-hub provider flaky | Assert reply **or** provider hint (as TS) |
| Lifecycle trash timing | Stable testids; increase timeout only where TS had it |
