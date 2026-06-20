# Hub — Единая база знаний

> Все агенты AI-команды загружают эти файлы перед любой задачей.
> Регламент: `.cursor/REGULATION.md` · **Docs sync после изменений:** `.cursor/rules/55-docs-sync.mdc`

## Общие файлы (обязательные)

| Файл | Назначение |
|------|------------|
| [vision_and_scope.md](vision_and_scope.md) | Бизнес-цели, KPI, User Stories, MoSCoW |
| [system_specification.md](system_specification.md) | FR/NFR, ERD, API контракты |
| [architecture.md](architecture.md) | C4, deployment, паттерны |
| [test_plan.md](test_plan.md) | Стратегия тестирования, регрессия |
| [stack_and_tools.md](stack_and_tools.md) | Стек, версии, CI/CD |
| [code_style_guide.md](code_style_guide.md) | Стиль кода Go/TS |
| [security_guidelines.md](security_guidelines.md) | OWASP, секреты, чек-лист |
| [competitive_audit.md](competitive_audit.md) | Best practices конкурентов, feature backlog |
| [tz_constraints.md](tz_constraints.md) | TZ mandatory/forbidden (tracked mirror) |
| [user_needs_analysis_2026.md](user_needs_analysis_2026.md) | Персоны, JTBD, потребности пользователей |

## Детальная документация (справочная)

| Файл | Назначение |
|------|------------|
| [BUSINESS.md](BUSINESS.md) | Расширенная бизнес-документация |
| [TECHNICAL.md](TECHNICAL.md) | Полная техническая спецификация |
| [TESTING.md](TESTING.md) | Детали тестового покрытия |
| [wiki/API.md](wiki/API.md) | Справочник API |
| [wiki/ARCHITECTURE.md](wiki/ARCHITECTURE.md) | Wiki-архитектура |
| [wiki/DEPLOY.md](wiki/DEPLOY.md) | Деплой |

## Планирование

| Путь | Назначение |
|------|------------|
| `.planning/PROJECT.md` | Полный список требований |
| `.planning/ROADMAP.md` | Фазы и milestones |
| `.planning/STATE.md` | Текущий статус |
| `.planning/PLATFORM-VISION-v4.md` | Corporate super-app vision |
| `.planning/AI-PROVIDER-HUB-SPEC.md` | Multi-model AI gateway |
| `.planning/MICROSERVICES-ROADMAP.md` | Service extraction plan |
| `.planning/FEATURE-BACKLOG-2026.md` | FB + MOD + AI backlog |
| `.planning/EXECUTIVE-REPORT-2026.md` | Strategy summary |
| `.planning/DISCUSS-ROADMAP-LOCKED.md` | **PO decisions LOCKED** 2026-06-21 |
| `.planning/decisions/ADR-*.md` | AUTH, NOTIF, TZ rewrite ADRs |

## Git / CI hooks

| Platform | Files |
|----------|-------|
| GitHub | `.github/workflows/ci.yml`, `pr-guard.yml`, `dependabot.yml`, issue/PR templates |
| GitLab | `.gitlab-ci.yml` (mirror CI) |

**Local:** `.cursor/` в `.gitignore` — rules/skills держать локально или копировать из wiki.

---
*2026-06-21 — v4 vision + CI hooks*
