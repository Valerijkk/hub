# Hub — Единая база знаний

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
| [competitive_audit.md](competitive_audit.md) | Конкуренты, feature backlog | RU |
| [user_needs_analysis_2026.md](user_needs_analysis_2026.md) | Персоны, JTBD | RU |

---

## Справочная документация

| Файл | Назначение |
|------|------------|
| [BUSINESS.md](BUSINESS.md) | Расширенная бизнес-документация |
| [TECHNICAL.md](TECHNICAL.md) | ERD, API, WebSocket, env vars, деплой |
| [TESTING.md](TESTING.md) | Тест-кейсы TC-*, матрица автоматизации |
| [testframework_strategy.md](testframework_strategy.md) | Стратегия тестфреймворка: слои Go/TS/Python, Locust, Allure→TMS, гейты |
| [onboarding/](onboarding/README.md) | Онбординг-листы по ролям (BE/FE/QA-manual/QA-auto/PM+ops) |
| [wiki/HOME.md](wiki/HOME.md) | Точка входа wiki |
| [wiki/API.md](wiki/API.md) | Справочник REST API (~286 endpoints) |
| [wiki/DEPLOY.md](wiki/DEPLOY.md) | Установка, Docker, CI/CD |
| [wiki/QUICKSTART.md](wiki/QUICKSTART.md) | Быстрый старт для пользователей |
| [wiki/SHORTCUTS.md](wiki/SHORTCUTS.md) | Горячие клавиши |
| [wiki/ARCHITECTURE.md](wiki/ARCHITECTURE.md) | Wiki-архитектура |

---

## Планирование и статус

| Путь | Назначение |
|------|------------|
| `.planning/PROJECT.md` | Полный список требований |
| `.planning/ROADMAP.md` | Фазы и milestones (исторический) |
| **`.planning/ROADMAP-FINAL-2026-07-03.md`** | **Актуальный roadmap: волны Track F/P, TF-PY, DOCS-EXCELLENCE** |
| `.planning/GRAY-ZONES-DECISIONS-2026-07-03.md` | **Реестр решений по ~90 серым зонам** |
| `.planning/STATE.md` | **Текущий статус сессии** |
| `.planning/AUDIT-2026-06-23.md` | Полный аудит проекта (2026-06-23) |
| `.planning/EXPANSION-ANALYSIS-2026-07-02.md` | Анализ глубины модулей + бэклог расширения |
| `.planning/research/MARKET-AND-NEEDS-2026-07-02.md` | Рынок/конкуренты/потребности сотрудников |
| `.planning/COVERAGE.md` | Матрица requirement → code → test |
| `.planning/PLATFORM-VISION-v4.md` | Corporate super-app vision |
| `.planning/decisions/ADR-*.md` | AUTH, NOTIF, TZ rewrite, **ARCH-MICROSERVICES-2026** |

---

## Git / CI

| Платформа | Файлы |
|-----------|-------|
| GitHub | `.github/workflows/ci.yml`, `deploy.yml`, `pr-guard.yml`, `dependabot.yml` |
| GitLab | `.gitlab-ci.yml` (mirror gates) |

**Версии в CI (2026-06-21):** Go **1.25**, Node **22**, Mongo **8** (service), Postgres/Redis/NATS в e2e-full via `scripts/ci-e2e-backend.sh`.

**Команда агентов:** `.cursor/` **трекается в git** (решение 2026-07-03): агенты (`.cursor/agents/`, 30 ролей), playbooks (`.cursor/playbooks/`, 10 файлов), правила, скилы. Индекс: `.cursor/playbooks/README.md`.

---

## Быстрые команды

```bash
# Полный backend regression
cd backend && go generate ./ent && go build -tags release ./... \
  && go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/...

# Frontend regression
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run

# Локальный стек
docker compose --env-file .env.local -f docker-compose.local.yml up --build
```

Skill: `/qa-gate` — останавливается на первой FAIL.

---
*2026-07-03 — реорганизация AI-команды (30 агентов + playbooks), ADR-MICROSERVICES, ROADMAP-FINAL, DOCS-EXCELLENCE*
