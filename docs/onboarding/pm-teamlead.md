# Онбординг — PM / Тимлид (+ админ/ops)

> Роль: управление roadmap, волнами разработки и AI-командой Hub; в конце — секция для админа/ops.
> Главные документы: `.cursor/playbooks/team-workflow.md` (RACI, волны, handoff) + `.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` · Индекс: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Ведёшь roadmap и планирование: волны, приоритеты, scope, готовность milestone.
- Оркеструешь AI-команду (32 агента) вместе с владельцем: декомпозиция → волна → интеграция → приёмка.
- Держишь ворота качества: регрессия у исполнителей, верификация `*-ver`, полный `/qa-gate` на интеграции волны, docs-sync.
- Следишь за границами scope (`docs/tz_constraints.md` — mandatory/forbidden) и бюджетом CI-минут (~3000/мес).
- **Не мержишь в master** — это делает только владелец проекта; твоя зона — довести волну до merge-ready и передать.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Код собирать не обязательно, но продукт подними:

- [ ] Локальный стек: `docker compose --env-file .env.local -f docker-compose.local.yml up --build` → `http://localhost:8080`, dev-login, seed.
- [ ] Доступ к GitHub (PR, Actions), инстансу Hub.

**Что прочитать (в этом порядке):**

- [ ] `AGENTS.md` — статус проекта, метрики, ростер команды: стартовая точка.
- [ ] `docs/vision_and_scope.md` — бизнес-цели, KPI, User Stories.
- [ ] **`.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md`** — актуальный план: два трека (F функционал / P платформа-микросервисы), волны 1–12+, эпик DOCS-EXCELLENCE, фаза TF-PY.
- [ ] `.planning/STATE.md` — текущее состояние сессий; `.planning/COVERAGE.md` — качество.
- [ ] **`.cursor/playbooks/team-workflow.md`** — RACI, потоки фич, параллелизм, handoff.
- [ ] `.cursor/playbooks/README.md` — ростер агентов и карта playbooks.
- [ ] `.cursor/REGULATION.md` — регламент; `.cursor/rules/50-git.mdc`, `55-docs-sync.mdc`, `56-regression-verify.mdc` — ворота.
- [ ] `docs/tz_constraints.md` — что менять нельзя (mandatory/forbidden) — твой инструмент против scope creep.

## Неделя 1

- [ ] Пройди продукт по всем модулям как пользователь; сверь увиденное с KPI/US в vision_and_scope.
- [ ] Разбери одну прошедшую волну по `.planning/`: как выглядели планы, handoff, верификация, что попало в STATE.md.
- [ ] Посмотри 2–3 закрытых PR: структура (PR=фича, атомарные коммиты, rebase-merge), кто ревьюил, как выглядят логи регрессии.
- [ ] Разбери процесс волны на текущей: декомпозиция tech-lead → исполнители → верификаторы → интеграция team-lead → merge-ready.
- [ ] **Первая задача:** проведи планирование одной маленькой волны (или подволны): нарезка задач ≤200–300 строк/план, назначение исполнителей и верификаторов, карта файлов без пересечений, ворота.

## Инструменты и команды

Твой основной интерфейс — планирование и агенты; из команд полезны:

```bash
# Статус волны глазами git
git log --oneline master..feat/<branch>
gh pr list && gh pr checks <n>

# Полный гейт на интеграции волны (через агента или локально)
# см. .cursor/skills/qa-gate/SKILL.md
```

**Модель волн (roadmap 2026-07-03):** волна ≤4 агента (мягкий лимит), планы атомарные ≤200–300 строк, границы файлов между параллельными потоками назначаются заранее. Ворота каждой волны:

1. Максимальная декомпозиция до старта (tech-lead → приёмка team-lead).
2. Исполнитель гоняет регрессию своего домена; полный `/qa-gate` — один раз на интеграции волны.
3. QA-конвейер: TC → qa-manual-ver → автотесты → qa-auto-ver; юниты пишут be/fe.
4. Coverage нового кода ≥82% (security-critical ≥80%).
5. Docs-sync в той же волне.
6. Мерж в master — только по команде владельца (личное ревью: security/миграции/CI/ADR).

## Как работает AI-команда и как с ней взаимодействовать

Ростер и модели: `.cursor/playbooks/README.md`. Кому что поручать:

| Нужно | Агент |
|-------|-------|
| Декомпозиция фичи/milestone на планы | `tech-lead` |
| Оркестрация волны, приёмка | `team-lead` (эскалация — `team-lead-opus`) |
| Приоритеты, scope, KPI | `product` (+ `product-ver`) |
| US/AC | `ba1`/`ba2` → `ba-ver` |
| FR/API/ERD-спека | `sa1`/`sa2` → `sa-ver` |
| Архитектура, ADR | `arch` → `arch-ver` |
| Дизайн-спека | `designer` → `designer-ver` |
| Реализация | `be1`/`be2`, `fe1`/`fe2` → `be-ver`/`fe-ver` |
| Тест-кейсы / автотесты | `qa-manual` → `qa-manual-ver`; `qa-auto1`/`qa-auto2` → `qa-auto-ver` |
| Инфра/CI/деплой | `devops-engineer` |
| Security-аудит | `security-auditor` |
| Root-cause бага | `debugger` |
| Pre-merge ревью | `code-reviewer` |
| Docs-sync/wiki | `docs-writer` |

Принципы: каждый артефакт проходит своего верификатора (`verified: true` только при green regression + лог); дубли ролей работают параллельно без пересечения файлов; кросс-ревью внутри пары до верификатора. Git-политика: агенты и люди коммитят в фича-ветки свободно, **push/мерж в master — только владелец лично**.

---

## Секция: админ / ops

Минимум для того, кто отвечает за прод-инстанс. Главные документы: `docs/wiki/DEPLOY.md`, `docs/security_guidelines.md`, `docs/stack_and_tools.md`.

### Деплой

- **Dev:** Docker Compose (`docker-compose.local.yml`); **Prod:** K3s + Caddy (reverse proxy — не nginx). Манифесты: `deployments/`.
- CI/CD: GitHub Actions (`.github/workflows/ci.yml`, `deploy.yml`); деплой гейтится на зелёный CI. `DEV_MODE=true` в проде запрещён.
- Инфра-сервисы: PostgreSQL 17, MongoDB 8, Redis 7, NATS + JetStream, LiveKit, SeaweedFS (пиновать версию, не `latest`).

### Секреты

- Никаких секретов в репозитории: `.env` вне git, только `.env.example`; проверка diff на `secret|password|token|key` перед коммитом.
- Хранение: GitHub Secrets (CI) + env на хосте; ротация — по runbook (задача 1.9 Волны 1, FA-06 — ротация git-секретов выполняется владельцем по чек-листу).
- JWT 15 мин / refresh в Redis 30 дней — instant revocation через Redis.

### Бэкапы

- Модуль `backend/internal/backup` — ZIP-бэкапы данных; регулярность и хранение вне хоста — настрой по DEPLOY.md.
- PostgreSQL: дампы по расписанию; MongoDB (сообщения) и SeaweedFS (файлы) — отдельные политики. Восстановление отрепетируй до того, как понадобится.

### Алерты и наблюдаемость

- Логи: структурированный `slog` с request_id; playbook `observability.md`.
- Здоровье: `GET /api/health`; смотри p95 API (<200ms) и WS latency (<100ms) — SLO из `docs/testframework_strategy.md`.
- OTel-трейсинг разворачивается в Track P (M0-06); алерты на деградацию p95 и ошибки 5xx — минимальный набор.

### Чек-лист ops

- [ ] Прочитаны DEPLOY.md + security_guidelines.md.
- [ ] Доступ к прод-хосту/K3s и GitHub Secrets получен.
- [ ] Бэкапы настроены, восстановление проверено на копии.
- [ ] Ротация секретов по runbook понятна; календарь ротации заведён.
- [ ] Алерты на health/5xx/p95 настроены.

---

## Чек-лист готовности (PM/тимлид)

- [ ] Прочитаны: AGENTS.md, vision_and_scope, ROADMAP-FINAL, team-workflow.md, REGULATION.
- [ ] Понимаю два трека roadmap (F/P), модель волн и ворота.
- [ ] Знаю ростер агентов и кому что поручать; понимаю цепочку верификации.
- [ ] Понимаю QA-конвейер и почему юниты — зона разработчиков.
- [ ] Провёл планирование первой волны (декомпозиция, границы файлов, ворота).
- [ ] Знаю git-политику: волна доводится до merge-ready, мерж в master — только владелец.
