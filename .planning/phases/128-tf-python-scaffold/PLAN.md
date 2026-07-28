# Phase 128 — TF Python scaffold (`/tests`)

**Status:** PLANNED  
**Migration stage:** Э1  
**Estimate:** 2–3 person-days  
**Branch (suggested):** `feat/qa/tf-python-scaffold`

## Goal

Поднять каркас корневого Python-проекта `/tests` по дизайну
[`TF-MIGRATE-E4-TESTS-LAYOUT.md`](../../audits/planning-audit-2026-07/TF-MIGRATE-E4-TESTS-LAYOUT.md):
`pyproject.toml`, корневой `conftest.py` (env + DB + auth), пустые пакеты
`integration/`, `e2e/`, `load/`, `fixtures/`.

**Критерий:** один integration-тест проходит end-to-end против Docker local stack
(`health` + authenticated API call).

## Hard dependency

- **BLOCKED on A3 coverage baseline** — [`01-PLANNING-FIXES.md`](../../audits/planning-audit-2026-07/01-PLANNING-FIXES.md) §A3  
  (прогнать coverage gate, обновить `COVERAGE.md` с датой).  
  Без baseline нельзя начинать миграцию — см. [`02-TEST-FRAMEWORK-MIGRATION.md`](../../audits/planning-audit-2026-07/02-TEST-FRAMEWORK-MIGRATION.md).

## Depends on

- A3 (planning-audit) — **hard**
- Docker compose local/test stack доступен

## Deliverables

- [ ] `tests/pyproject.toml` + README
- [ ] `tests/conftest.py` — `api_url`, `backend_available`, `seed_dev`, `dev_tokens`, `auth_headers`
- [ ] `tests/fixtures/{auth,env,db,seed}.py` stubs
- [ ] `tests/integration/platform/test_health.py` — green
- [ ] `tests/integration/auth/test_dev_login.py` — green
- [ ] Документировать команды install / pytest

## Out of scope

- Перенос E2E specs
- Удаление `testframework/`
- Locust scenarios (только пустая `load/` папка)

## Success criteria

- [ ] A3 baseline зафиксирован до merge этой фазы
- [ ] `pytest tests/integration -q` green locally
- [ ] Старый `testframework/` не удалён

## Risks

| Risk | Mitigation |
|------|------------|
| Dev-login only in DEV_MODE | Document; skipif when unavailable |
| Windows vs Linux path/CI | Use pytest from `tests/` as cwd |
