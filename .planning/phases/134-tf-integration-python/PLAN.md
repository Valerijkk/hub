# Phase 134 — Python integration layer

**Status:** PLANNED  
**Migration stage:** Э5  
**Estimate:** 5–7 person-days  
**Branch (suggested):** `feat/qa/tf-integration-python`

## Goal

Перенести в `tests/integration/` то, что по смыслу **интеграционное** (БД, NATS,
Mongo, внешние интеграции через API), а не юнит — по мере выделения из бывшего
`backend/testframework` и новых сценариев.

Layout: [`TF-MIGRATE-E4-TESTS-LAYOUT.md`](../../audits/planning-audit-2026-07/TF-MIGRATE-E4-TESTS-LAYOUT.md)

## Depends on

- **128** tf-python-scaffold (hard)
- Soft: **132** частично (чтобы не дублировать unit+integration)

## Deliverables

- [ ] Domains: auth, tracker, messenger, search, notification, bases, platform
- [ ] DB fixtures (`pg`, `mongo`) skipif
- [ ] Factories через API (`tests/fixtures/factories.py`)
- [ ] CI job draft `integration-python` (включается в 136)

## Out of scope

- Browser E2E
- Locust (135)
- Импорт Go internal из Python

## Success criteria

- [ ] ≥1 тест на домен из списка выше green
- [ ] Нет прямого SQL в e2e; integration может читать DB для assert

## Risks

| Risk | Mitigation |
|------|------------|
| Дубли с Go tests | Матрица: unit = logic; integration = wire+DB |
| Flaky Docker | Same compose as hubtest; explicit health wait |
