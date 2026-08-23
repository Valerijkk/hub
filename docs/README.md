# Hub — Единая база знаний

> Active QA status (2026-08-23): minimal Docker gate PASS — backend build/vet/unit/race, frontend TypeScript plus Vitest 338/338 suites and 419/419 tests, and 181/181 framework selftests. Five SUT contract tests/cases were removed by founder direction; API/integration/e2e/load are `NOT_RUN` and are not gates for this slice. Hosted CI is off.

> Все агенты AI-команды загружают эти файлы **перед любой задачей**.  
> Регламент: `.cursor/REGULATION.md` · **Docs sync:** `.cursor/rules/55-docs-sync.mdc`

---

## Обязательные документы

| Файл | Назначение | Язык |
|------|------------|------|
| [vision_and_scope.md](vision_and_scope.md) | Бизнес-цели, KPI, User Stories, MoSCoW | RU |
| [system_specification.md](system_specification.md) | FR/NFR, ERD, API-контракты | RU |
| [architecture.md](architecture.md) | C4, deployment, паттерны | RU |
| [test_plan.md](test_plan.md) | Стратегия тестирования, CI, регрессия | RU |
| [stack_and_tools.md](stack_and_tools.md) | Стек, версии, инструменты, CI/CD | RU |
| [code_style_guide.md](code_style_guide.md) | Стиль Go/TS, SOLID, запреты | RU |
| [security_guidelines.md](security_guidelines.md) | OWASP, секреты, security checklist | RU |
| [tz_constraints.md](tz_constraints.md) | TZ mandatory/forbidden (tracked mirror) | RU |
| [competitive_analysis_2026.md](competitive_analysis_2026.md) | Конкуренты, TCO, gaps | RU |
| [competitive_audit.md](competitive_audit.md) | Best practices, feature backlog | RU |
| [user_needs_analysis_2026.md](user_needs_analysis_2026.md) | Персоны, JTBD | RU |
| [GLOSSARY.md](GLOSSARY.md) | Термины, ID-конвенции (TC-/FR-/GZ-/ECC-…), модули backend | RU |

Решения о качестве кода и тестов: [`.cursor/knowledge/code-practices/README.md`](../.cursor/knowledge/code-practices/README.md)
— evidence-aware каталог hard controls, review heuristics, good/bad examples и
negative sentinels. Он дополняет, но не заменяет `code_style_guide.md`,
`security_guidelines.md` и доменные спецификации.

---

## Справочная документация

| Файл | Назначение |
|------|------------|
| [BUSINESS.md](BUSINESS.md) | Расширенная бизнес-документация |
| [TECHNICAL.md](TECHNICAL.md) | ERD, API, WebSocket, env vars, деплой |
| [TESTING.md](TESTING.md) | Тест-кейсы TC-*, матрица автоматизации |
| [testframework_strategy.md](testframework_strategy.md) | Стратегия тестфреймворка: слои Go/TS/Python, Locust, Allure→TMS, гейты |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Текущий root-Compose QA protocol, legacy quarantine и окружение |
| [runbooks/](runbooks/README.md) | **Эксплуатация прода:** инцидент, алерты, деплой/откат, БД, бэкапы, ротация секретов |
| [onboarding/](onboarding/README.md) | Онбординг-листы по ролям (BE/FE/QA-manual/QA-auto/PM+ops) |
| [wiki/HOME.md](wiki/HOME.md) | Точка входа wiki |
| [wiki/API.md](wiki/API.md) | Справочник REST API (OpenAPI **252** paths в `swagger.json`; исторический route inventory ~286 — см. таблицу в API.md) |
| [wiki/DEPLOY.md](wiki/DEPLOY.md) | Установка, Docker, CI/CD |
| [wiki/QUICKSTART.md](wiki/QUICKSTART.md) | Быстрый старт для пользователей |
| [wiki/SHORTCUTS.md](wiki/SHORTCUTS.md) | Горячие клавиши |
| [wiki/ARCHITECTURE.md](wiki/ARCHITECTURE.md) | Wiki-архитектура |

---

## Планирование и статус

| Путь | Назначение |
|------|------------|
| `.planning/PROJECT.md` | Полный список требований |
| `.planning/ROADMAP.md` | Фазы и milestones; live test-rebase 128–141 + исторические фазы |
| **`.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md`** | **Актуальный roadmap: волны Track F/P, TF-PY, DOCS-EXCELLENCE** |
| `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` | **Реестр решений по ~90 серым зонам** |
| `.planning/STATE.md` | **Текущий статус сессии** |
| `.planning/audits/2026-06/AUDIT-2026-06-23.md` | Полный аудит проекта (2026-06-23) |
| `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` | Анализ глубины модулей + бэклог расширения |
| `.planning/research/MARKET-AND-NEEDS-2026-07-02.md` | Рынок/конкуренты/потребности сотрудников |
| `.planning/COVERAGE.md` | Матрица requirement → code → test |
| `.planning/decisions/specs/PLATFORM-VISION-v4.md` | Corporate super-app vision |
| `.planning/decisions/ADR-*.md` | AUTH, NOTIF, TZ rewrite, **ARCH-MICROSERVICES-2026** |

---

## Git / CI

| Платформа | Файлы |
|-----------|-------|
| GitHub | `.github/workflows/ci.yml`, `deploy.yml`, `pr-guard.yml` |
| GitLab | `.gitlab-ci.yml` (mirror gates) |

**Версии в CI (2026-06-21):** Go **1.25**, Node **22**, Mongo **8** (service), Postgres/Redis/NATS в e2e-full via `scripts/ci-e2e-backend.sh`.

**Команда агентов:** `.cursor/` **трекается в git** (решение 2026-07-03): агенты (`.cursor/agents/`, 32 файла), playbooks (`.cursor/playbooks/`, 11 + README), правила, скилы. Индекс: `.cursor/playbooks/README.md`.

---

## Быстрый вход в проверки — только Docker

> Founder lock 2026-08-07: на Windows host запрещены tests, validators,
> linters, build/vet, Vite, Playwright, pytest/Locust и тестовые app-процессы.
> Host используется только для чтения/редактирования, Git и Docker CLI.

Единственная QA-точка входа — root Compose: `docker compose -f docker-compose.local.yml run --rm <exact-service>`.
Exact services: `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full`, `python-load`.
Они profile-explicit/disposable; integration/e2e используют изолированную QA-инфраструктуру.
Исторические legacy QA entrypoints архивны и не являются fallback; используй только `docker compose -f docker-compose.local.yml run --rm <service>`.

Test-rebase 128–141: legacy suites frozen и не являются gate; целевые
пути — `backend/tests/**` (только Go unit), `frontend/tests/**`
(только Vitest/RTL unit) и единый Python-проект прямо в
`testframework/`: `src/hub_testkit/**`, `selftests/**`, `cases/**`,
`schemas/**`, `policies/**`, `load/**`, `docker/**`, `reports/**`; SUT suite
sources outside the minimal slice do not become merge gates.
Fresh sequential Docker evidence for the minimal slice: `backend-unit` PASS
(build/vet/unit/race), `frontend-unit` PASS (TypeScript plus 338/338 suites and
419/419 tests), `python-contract` PASS (181/181 framework selftests). The five
SUT contract tests and five associated case records were explicitly deleted.
API/integration/e2e/load are `NOT_RUN` and are not merge gates for this slice.
Hosted GitHub CI/PR automation is manual-only and GitLab pipelines are off.
Accepted receipts cover the 14-path target-unit curation and 241
hook-dependent deletion batch. Full legacy-root reconciliation is 349
relocations + 232 net component deletions + 241 hook-dependent deletions = 822
removed source paths; protected Mail/Calendar paths stay outside the merge. Evidence:
`.planning/audits/test-rebase/FINAL-MERGE-EVIDENCE-2026-08-23.md`.

---
*2026-08-23 — minimal unit/framework Docker gate PASS; hosted CI off ·
2026-08-08 — Docker-only safe service contract, executable Phase 129 cutover/status ·
2026-08-07 — root Python testframework contract и current-vs-target split ·
2026-07-19 — GLOSSARY, TROUBLESHOOTING, runbooks/, актуализация фактов и счётчиков*
