# Phase 136 — TF CI cutover + delete legacy

**Status:** PLANNED  
**Migration stage:** Э7  
**Estimate:** 3–4 person-days  
**Branch (suggested):** `feat/qa/tf-ci-cleanup`

## Goal

Перевести `.github/workflows/ci.yml` (и related) на новую схему: юниты языком
слоя, Python-фреймворк отдельной job. Удалить legacy:

- `testframework/` (root)
- `backend/testframework/` (если ещё не пуст после 132)
- `frontend/e2e/` / TS Playwright node_modules paths

Источник: [`02-TEST-FRAMEWORK-MIGRATION.md`](../../planning-audit-2026-07/02-TEST-FRAMEWORK-MIGRATION.md) §Э7

## Depends on

- **130** + **131** — Python E2E parity 20/20
- **132** — backend units colocated (or residual plan closed)
- **133** — frontend units colocated
- Soft: **134**, **135** jobs optional/manual wired

## Deliverables

- [ ] CI jobs: `backend-unit`, `frontend-unit`, `python-integration`, `e2e-smoke` (Python), `e2e-full` (Python)
- [ ] Coverage gate path updated to `./internal/...`
- [ ] Remove old pipelines / `link-testframework-node-modules` if unused
- [ ] Docs sync: `AGENTS.md`, `testframework/PIPELINE.md` → `tests/README.md`, `docs/test_plan.md`
- [ ] Delete legacy directories in final commit(s)

## Out of scope

- Changing coverage floor policy (open question)
- FA-06 / deploy blockers

## Success criteria

- [ ] Green CI on PR without TS Playwright
- [ ] No references to deleted paths in docs/CI
- [ ] qa-gate skill/commands updated

## Risks

| Risk | Mitigation |
|------|------------|
| Minutes budget (student GH Actions) | Keep smoke small; full e2e on schedule/main |
| Broken links in agents/rules | Repo-wide grep before delete |
