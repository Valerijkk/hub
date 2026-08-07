# PLAN: Phase 129 — Python scaffold и API vertical slice

**Статус:** PLANNED · **Оценка:** 1–2 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Неразрушительно создать единый Python-проект прямо в корне
`testframework/` (без `testframework/python/`) и
доказать один health+authenticated API flow, не меняя legacy runners.

## Depends on

- Phase 128 PASS, включая frozen inventory и coverage contract.

## Success criteria

- [ ] Python 3.13/`uv`/pytest/httpx stack зафиксирован корневым
  `testframework/pyproject.toml`, единым pytest config и lockfile; clean install PASS.
- [ ] `pytest --strict-markers` собирает ненулевой API slice и выполняет health +
  auth against approved non-production profile.
- [ ] Deliberate failing assertion делает новый runner red; secrets redacted.
- [ ] `testframework/stages/**`, `guides/**` и `pipeline/**` не удалены/не сломаны.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-193 | Non-destructive root project scaffold, pinned dependencies, strict markers, typed env и artifact redaction | qa-auto1 | qa-auto-ver | `testframework/{pyproject.toml,uv.lock,conftest.py,README.md}`, `testframework/core/**`, `testframework/reports/**` | Phase 128 | 0.75д |
| TASK-205 | Health/auth vertical slice с isolated test user, cleanup и failing-sentinel evidence | qa-auto2 | qa-auto-ver | `testframework/api/{platform,auth}/**`, `testframework/core/fixtures/auth.py`, `testframework/reports/**`, `.planning/evidence/test-rebase/129/**` | TASK-193 | 0.75д |

## Волны

- Wave 1: TASK-193.
- Wave 2: TASK-205.

## Риски

| Риск | Mitigation |
|---|---|
| Python отсутствует в PATH | Provision exact runtime; без clean install verdict BLOCKED |
| DEV auth скрывает production contract | Маркер/profile явный; acceptance проверяет публичный authenticated response |
| Scaffold затрёт root legacy | Inventory checksum check до/после; delete запрещён |

## Verification

- `qa-auto-ver`: clean install, non-zero collection, green slice, red sentinel,
  два последовательных запуска без leaked data.
- Phase 128 legacy smoke запускается без изменения entry point.
