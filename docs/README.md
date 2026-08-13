# Hub — Единая база знаний

> Active Python QA routing: reusable modules belong under `testframework/src/hub_testkit/**`; planned SUT suites belong under `testframework/tests/{contract,api,integration,e2e}/**` and are currently absent.

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
| [mail/QA-HANDOFF.md](mail/QA-HANDOFF.md) | Контракт Mail 0.1.4 + Calendar RSVP/reliability: ADR/OpenAPI/events/security/QA; implementation `NOT_IMPLEMENTED`, runtime `NOT_RUN` |
| [Phase 147 reliability contract](../.planning/phases/147-mail-calendar-reliability/CONTRACT.md) | REQUEST/CANCEL recovery, outward REPLY, private availability, reminders, conditional L1 and CRM manual-link |
| [Phase 147 root static review](../.planning/phases/147-mail-calendar-reliability/STATIC-REVIEW.md) | 0.1.3 PASS на `8daaf0d9` стал STALE после 0.1.4 amendment; новый root/independent review PENDING, runtime `NOT_RUN` |
| [testframework_strategy.md](testframework_strategy.md) | Стратегия тестфреймворка: слои Go/TS/Python, Locust, Allure→TMS, гейты |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Текущий `/qa-gate`/`NOT_READY` протокол, legacy quarantine, окружение и exact cleanup |
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

Единственная точка входа — skill `/qa-gate`; он только маршрутизирует task и пока не может
её выполнить. Private source candidate уже существует в `testframework/runner/run-task.ps1`
и принят только статически; все target jobs обязаны использовать versioned outer launcher API
только из `testframework/runner/**` с task-owned names, CPU/RAM/PID limits, wall-clock timeout,
`reports/<task-run>` + owner marker и cleanup в `finally`. `testframework/docker/**` владеет
только immutable execution assets. `images.lock.v3.json`, dependency bytes,
`pyproject.toml`, `uv.lock`, `py.typed` и target suites отсутствуют. Package/launcher —
`NOT_READY`; `PS5.1_PARSE_NOT_RUN`, `PRIVATE_CONTRACT_NOT_RUN`, runtime — `NOT_RUN`.
Прямые Docker snippets и legacy runners не fallback.

Test-rebase 128–141: legacy suites frozen и не являются gate; целевые
пути — `backend/tests/**` (только Go unit), `frontend/tests/**`
(только Vitest/RTL unit) и единый Python-проект прямо в
`testframework/`: `src/hub_testkit/**`, planned `tests/{contract,api,integration,e2e}/**`,
`selftests/**`, `cases/**`, `schemas/**`, `policies/**`, `load/**`, `docker/**`,
`runner/**` и `reports/<task-run>/**`. Runner содержит только статически reviewed
private candidate; package, target suites и launcher —
`NOT_READY`; runtime — `NOT_RUN`. Legacy Go component/API/
integration, static-contract Vitest и TS Playwright удаляются в атомарной
Phase 129 по exact `cutover-manifest.json`, без port/parity/dual-run; каждый
delete batch переключает затронутые CI/scripts/docs в том же commit. Все 245
hook-dependent genuine Go units защищены до TASK-234. TASK-202 archived/verified;
current ledger = **1,049 rows / 4,246 declared**, backend **832 / 3,674**,
E2E baseline **48 pass / 5 expected skip / 0 fail**. Канон:
[`.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md`](../.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

---
*2026-08-13 — Mail 0.1.4 + Calendar RSVP/reliability contract index (`NOT_IMPLEMENTED` / `NOT_RUN`) ·
2026-08-08 — Docker-only safe runner contract, executable Phase 129 cutover/status ·
2026-08-07 — root Python testframework contract и current-vs-target split ·
2026-07-19 — GLOSSARY, TROUBLESHOOTING, runbooks/, актуализация фактов и счётчиков*
