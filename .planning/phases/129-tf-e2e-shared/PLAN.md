# Phase 129 — TF E2E shared layer (fixtures / auth / POM)

**Status:** PLANNED  
**Migration stage:** Э2 (shared rewrite)  
**Estimate:** 3–4 person-days  
**Branch (suggested):** `feat/qa/tf-e2e-shared`

## Goal

Переписать shared-слой Playwright с TypeScript на Python:
fixtures, auth (seed + sessionStorage inject), env probes, POM, steps, a11y helper.

Источник: `testframework/stages/e2e/shared/**`  
План классов: [`TF-MIGRATE-E3-PLAYWRIGHT.md`](../../planning-audit-2026-07/TF-MIGRATE-E3-PLAYWRIGHT.md) §Shared

## Depends on

- **128** tf-python-scaffold

## Deliverables

- [ ] `tests/e2e/conftest.py` — page fixtures, skipUnless* → markers
- [ ] `tests/fixtures/auth.py` — port `devLogin` / `authenticatePage`
- [ ] `tests/e2e/pages/` — Login, AppShell, Messenger, Base
- [ ] `tests/e2e/steps/auth.py`
- [ ] `tests/e2e/helpers/a11y.py` — axe port (smoke-ready)
- [ ] `tests/e2e/helpers/ai.py` — API helpers stub
- [ ] Один smoke self-check: login branding via POM

## Out of scope

- Полный перенос 20 specs (phases 130–131)
- CI cutover

## Success criteria

- [ ] Python auth inject открывает `/projects` без UI login
- [ ] axe helper возвращает список violations на login page
- [ ] Документирован паритет с TS helpers

## Risks

| Risk | Mitigation |
|------|------------|
| Zustand persist shape drift | Snapshot payload keys from current `auth.ts` |
| axe package choice | Pin version; same impact filter as TS |
