# Runbook FA-06 — ручная ротация секретов (Resend + Telegram)

> **Кто выполняет:** владелец (пользователь) лично — решение GZ-SEC-01 (`.planning/GRAY-ZONES-DECISIONS-2026-07-03.md` §1).
> **Почему:** ключи `RESEND_API_KEY` и `TELEGRAM_BOT_TOKEN` засветились в git-истории (finding INF-03, `docs/security_guidelines.md`). Репо приватный — BFG-purge истории **не делаем** (GZ-SEC-02), достаточно ротации. После ротации — тег `v2.5.0`.
> **Связанные материалы:** полный чек-лист по всем секретам — `.planning/PG-03-SECRET-ROTATION.md`; интерактивные скрипты — `bash scripts/fa06.sh help`; журнал выполнения — `.planning/FA-06-EXECUTION-LOG.md`.
> **Главное правило:** значения секретов НИКОГДА не вставлять в тикеты, чат, коммиты и логи.

## Обзор процедуры

| Шаг | Что делаем | Время |
|-----|------------|-------|
| 0 | Подготовка: доступы + git-скан | ~10 мин |
| 1 | Новый Resend API key | ~5 мин |
| 2 | Новый Telegram bot token (BotFather) | ~5 мин |
| 3 | Обновление `.env` / K3s Secrets / GitHub Secrets | ~10 мин |
| 4 | Перезапуск сервисов | ~5 мин |
| 5 | Проверка: тест-письмо + тест-уведомление | ~10 мин |
| 6 | Отзыв старых ключей | ~5 мин |
| 7 | Sign-off + журнал | ~5 мин |

Порядок важен: **сначала выпускаем новые ключи и разворачиваем их, потом отзываем старые** — иначе будет окно недоступности почты/уведомлений. Исключение — Telegram: `/revoke` в BotFather отзывает старый токен одновременно с выпуском нового (см. шаг 2), поэтому Telegram-уведомления будут недоступны от `/revoke` до перезапуска сервисов на шаге 4 — это ожидаемо и допустимо (минуты).

---

## Шаг 0 — Подготовка

- [ ] Есть доступ к дашборду [Resend](https://resend.com) (аккаунт, на котором заведён домен отправки).
- [ ] Есть доступ к Telegram-аккаунту, владеющему ботом Hub (диалог с [@BotFather](https://t.me/BotFather)).
- [ ] Есть SSH-доступ к prod-хосту (Docker Compose) и/или `kubectl`-доступ к кластеру K3s.
- [ ] (Опционально, для фиксации масштаба) git-скан истории с локальной машины:

```bash
git log -S 'RESEND_API_KEY' --oneline --all
git log -S 're_' --oneline --all -- '*.yml' '*.yaml' '*.env*'
git log -S 'TELEGRAM_BOT_TOKEN' --oneline --all
```

Найденные SHA запиши в журнал (`.planning/FA-06-EXECUTION-LOG.md`). Полный вариант скана по всем секретам — `.planning/PG-03-SECRET-ROTATION.md`, Phase 0.

## Шаг 1 — Новый Resend API key

1. Зайди в Resend dashboard → **API Keys**.
2. **Create API Key**: имя вида `hub-prod-2026-07`, permission **Sending access**, домен — твой домен отправки.
3. Скопируй значение (`re_...`) сразу — оно показывается один раз. Храни только в менеджере паролей.
4. **Старый ключ пока НЕ удаляй** — отзыв на шаге 6.
5. Проверь, что `RESEND_FROM_EMAIL` (и `RESEND_FROM_ALERTS`, если используется) по-прежнему соответствует верифицированному домену в Resend.

## Шаг 2 — Новый Telegram bot token

1. Открой диалог с **@BotFather** → команда `/mybots` → выбери бота Hub.
2. **API Token** → **Revoke current token**. BotFather сразу выдаст **новый** токен (`1234567890:AA...`).
3. Скопируй новый токен в менеджер паролей.
4. ⚠️ С этого момента старый токен мёртв — уведомления в Telegram не работают до завершения шага 4. Делай шаги 3–4 сразу после.

Токен используется только для notify bind/delivery (ADR-AUTH-01) — логин через Telegram в Hub отсутствует, ротация не влияет на вход пользователей.

## Шаг 3 — Обновление секретов в окружениях

### 3a. Docker Compose prod (основной путь)

На prod-хосте (файл `deployments/.env` — в `.gitignore`, никогда не коммитится):

```bash
cd <hub>/deployments
# отредактируй .env:
#   RESEND_API_KEY=<новое значение re_...>
#   TELEGRAM_BOT_TOKEN=<новый токен от BotFather>
```

Подсказка-чеклист с проверкой наличия ключей: `bash scripts/fa06.sh p1-compose`.

### 3b. K3s (если prod в кластере)

```bash
# base64 без переноса строки:
echo -n '<PLAINTEXT>' | base64 -w0

kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/RESEND_API_KEY", "value": "<BASE64_RESEND_API_KEY>"},
  {"op": "replace", "path": "/data/TELEGRAM_BOT_TOKEN", "value": "<BASE64_TELEGRAM_BOT_TOKEN>"}
]'
```

Чеклист: `bash scripts/fa06.sh p1-k3s`.

### 3c. GitHub Secrets

Repo → **Settings → Secrets and variables → Actions**. Сейчас в workflows используются только `SSH_HOST`/`SSH_USER`/`SSH_KEY` (deploy) — Resend/Telegram в GitHub Secrets **не хранятся**. Проверь список: если `RESEND_API_KEY`/`TELEGRAM_BOT_TOKEN` там когда-либо добавлялись — обнови их новыми значениями или удали за ненадобностью.

### 3d. Локальные dev-окружения

`docker-compose.local.yml` использует placeholder `${RESEND_API_KEY:-}` (INF-03 fix) — в репо значений нет. Если старые ключи лежат в твоём локальном `.env` — замени на новые или оставь пустыми (dev-режим).

## Шаг 4 — Перезапуск сервисов

### Docker Compose

```bash
cd <hub>/deployments
docker compose -f docker-compose.prod.yml --env-file .env -p hub \
  up -d --force-recreate backend hub-mail
```

### K3s

```bash
kubectl rollout restart deployment/hub-backend -n hub
kubectl rollout restart deployment/hub-mail -n hub
kubectl rollout status deployment/hub-backend -n hub --timeout=300s
```

## Шаг 5 — Проверка

Автоматизированный smoke: `bash scripts/fa06.sh p2-smoke` (задай `HUB_BASE`, `MAGIC_LINK_EMAIL`; после логина — `HUB_ACCESS_TOKEN` для расширенных проверок).

Ручная проверка (обязательные два пункта задачи):

- [ ] **Тест-письмо (Resend):** открой `/login` → введи свой email → Magic Link-письмо пришло → вход по коду успешен. Если письма нет — проверь Resend dashboard → Logs (ошибка ключа/домена).
- [ ] **Тест-уведомление (Telegram):** Settings → Notifications → Telegram: если бинд уже был — отправь тестовое уведомление (например, назначь на себя задачу) и убедись, что сообщение от бота пришло; если бинда нет — пройди привязку заново и получи подтверждение от бота.
- [ ] `curl -sS "$HUB_BASE/health"` → `"status":"ok"` (postgres/redis/mongo).

## Шаг 6 — Отзыв старых ключей

- [ ] Resend dashboard → API Keys → **Delete/Revoke** старый ключ (тот, что был в git-истории).
- [ ] Telegram: старый токен уже отозван `/revoke` на шаге 2 — дополнительных действий нет. Убедись, что нигде (CI, cron, сторонние скрипты) не остался старый токен.
- [ ] Быстрый скан рабочего дерева на «живые» значения (должно быть пусто, placeholder-ы допустимы):

```bash
git grep -E '(re_[a-zA-Z0-9]{20,}|[0-9]{8,}:[A-Za-z0-9_-]{30,})' -- ':!*.md' ':!.planning/*'
```

## Шаг 7 — Sign-off

- [ ] Отметь строки RESEND / TELEGRAM в `.planning/FA-06-EXECUTION-LOG.md` (Rotated / Verified, дата UTC, исполнитель).
- [ ] Заполни sign-off секцию в `.planning/PG-03-SECRET-ROTATION.md` (Phase 3): git-скан, ротация, smoke, отзыв старых ключей.
- [ ] Убедись, что `.env` не попал в git: `git status` чист от `deployments/.env`.
- [ ] Дальше по ops-лейну: `bash scripts/fa06.sh release-gate` → тег `v2.5.0` → SSH deploy (см. `.planning/ROADMAP-FINAL-2026-07-03.md`, Ops-лейн).

## Итоговый чек-лист (всё вместе)

- [ ] Git-скан истории выполнен, SHA зафиксированы в журнале
- [ ] Новый Resend API key выпущен (Sending access, домен верифицирован)
- [ ] Новый Telegram bot token получен через BotFather `/revoke`
- [ ] `deployments/.env` (и/или K3s `hub-app-secrets`) обновлены новыми значениями
- [ ] GitHub Secrets проверены (Resend/Telegram там не хранятся; SSH-ключи не тронуты)
- [ ] `backend` + `hub-mail` пересозданы/перезапущены
- [ ] Тест-письмо Magic Link получено, вход успешен
- [ ] Тест-уведомление Telegram получено
- [ ] Smoke `scripts/fa06.sh p2-smoke` — pass
- [ ] Старый Resend key отозван; старый Telegram token недействителен
- [ ] Рабочее дерево чисто от секретов (`git grep` пуст)
- [ ] Журнал FA-06 + sign-off PG-03 заполнены
- [ ] Готово к `v2.5.0` release-gate

---
*Владелец: пользователь (ручная ротация, GZ-SEC-01) · Подготовил: devops/docs, Волна 1 задача 1.9/1.13 · Ссылки: `scripts/fa06.sh`, `.planning/PG-03-SECRET-ROTATION.md`, `docs/security_guidelines.md` (INF-03) · 2026-07-03*
