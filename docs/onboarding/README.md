# Hub — Онбординг по ролям

> Чек-листы первого дня и первой недели для **людей**, присоединяющихся к проекту Hub.
> Создано в рамках эпика DOCS-EXCELLENCE (DOC-05 ⚡, 2026-07-03).

Hub — self-hosted корпоративный super-app (Task Tracker + Messenger + TMS + Voice Calls + Dashboards + intranet/calendar/CRM/HR/учёт/no-code базы/AI hub). Разработку ведёт AI-команда из 32 агентов под управлением владельца проекта; люди подключаются к ролям и работают **вместе** с агентами.

## Выбери свой лист

| Роль | Файл | Кратко |
|------|------|--------|
| Backend-разработчик | [backend-developer.md](backend-developer.md) | Go 1.25 + Fiber v3 + Ent, layering, юниты — твоя обязанность |
| Frontend-разработчик | [frontend-developer.md](frontend-developer.md) | React 19 + TS + TanStack + Tailwind 4/Shadcn, юниты Vitest |
| Ручной QA | [qa-manual.md](qa-manual.md) | Бизнес-дока → TC в `docs/TESTING.md`, тест-дизайн, верификация доки |
| QA-автоматизатор | [qa-automation.md](qa-automation.md) | Автотесты по верифицированным TC; целевой стек Python + Allure + Locust |
| PM / тимлид (+ админ/ops) | [pm-teamlead.md](pm-teamlead.md) | Roadmap, волны, RACI; секция ops: деплой, секреты, бэкапы |

## Общее для всех ролей (день 1)

1. **Доступы:** GitHub-репозиторий, инстанс Hub (dev), корпоративная почта (Magic Link — единственный способ входа).
2. **Локальное окружение:**

```bash
git clone <repo> && cd hub
cp .env.local.example .env.local
docker compose --env-file .env.local -f docker-compose.local.yml up --build
# UI: http://localhost:8080 · демо-данные: curl -X POST http://localhost:8080/api/dev/seed
```

В dev-режиме (`DEV_MODE=true`) доступен **dev-login** — вход без почты.

3. **База знаний:** начни с `docs/README.md` (карта документации), затем документы своей роли — они перечислены в каждом листе.
4. **Правила игры:** `.cursor/REGULATION.md` (регламент), `.cursor/playbooks/README.md` (ростер агентов и карта playbooks), `docs/tz_constraints.md` (что нельзя менять).

## Git-политика (одинакова для людей и агентов)

- Каждая новая Codex-задача начинается в короткой ветке `codex/<task>` от обновлённого `master`; Cloud Code использует свой отдельный prefix, который задаёт его worker.
- `master` — защищённый канонический baseline, не заявление о готовности продукта к production. Прямая реализация в нём запрещена.
- После применимых Docker-only проверок, независимой верификации/ревью и docs-sync `team-lead` по standing founder authority мержит routine accepted change в `master`, push-ит его и удаляет merged branch local+remote. Для security/auth, DB migrations, deploy/CI и ADR сначала нужен founder review.
- Перед handoff: регрессия своего слоя + docs-sync (`.cursor/rules/55-docs-sync.mdc`).

Детали — `.cursor/playbooks/git-workflow.md`.
