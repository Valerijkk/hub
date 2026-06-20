# Установка и деплой

## Требования

| Ресурс | Минимум | Рекомендуемый |
|--------|---------|--------------|
| CPU | 2 vCPU | 4 vCPU |
| RAM | 4 GB | 8 GB |
| Disk | 40 GB SSD | 100 GB SSD |
| ОС | Ubuntu 22.04 | Ubuntu 22.04 |

## Быстрый деплой (Docker Compose)

### 1. Установить Docker
```bash
curl -fsSL https://get.docker.com | sh
```

### 2. Клонировать репозиторий
```bash
cd /opt
git clone https://github.com/Valerijkk/hub.git
cd hub
```

### 3. Настроить окружение
```bash
cd deployments
cp .env.example .env
# Заполните .env реальными значениями (JWT, DB passwords, LiveKit keys, Resend, Telegram)
```

### 4. Запустить
```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

LiveKit в production **не** монтирует `livekit.yaml`: конфиг генерируется из env vars
в `docker-compose.prod.yml` (heredoc), включая `departure_timeout: 120` (INF-01).
После обновления compose пересоздайте контейнер: `docker compose ... up -d --force-recreate livekit`.

### 5. Проверить
```bash
# Все контейнеры запущены?
docker compose -f docker-compose.prod.yml ps

# Health check
curl http://localhost/api/health
```

Локальная разработка (полный стек, `DEV_MODE=true`):
```bash
cp .env.local.example .env.local   # опционально: Resend/Telegram
docker compose --env-file .env.local -f docker-compose.local.yml up --build
# Демо-данные только локально:
curl -X POST http://localhost:3000/api/dev/seed
```

## Переменные окружения

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://hub:pass@postgres:5432/hub` |
| `MONGODB_URL` | MongoDB connection string | `mongodb://hub:pass@mongo:27017/hub?authSource=admin` |
| `REDIS_URL` | Redis connection string | `redis://redis:6379/0` |
| `NATS_URL` | NATS connection string | `nats://nats:4222` |
| `JWT_SECRET` | Секрет для JWT подписи | `random-secret-32-chars` |
| `ENVIRONMENT` | Окружение (`production` в prod) | `production` |
| `INTEGRATION_ENCRYPTION_KEY` | AES key для GitHub/GitLab tokens (32 bytes base64) | см. `.env.example` |
| `DEV_MODE` | Режим разработки (`false` в prod compose; `true` только в local compose) | `false` |
| `CORS_ORIGINS` | Разрешённые origins (через запятую) | `https://hub.example.com` |
| `RESEND_API_KEY` | API ключ Resend (для email) | `re_xxx` |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram бота | `123:ABC` |

## CI/CD (GitHub Actions)

**CI** (`.github/workflows/ci.yml`): каждый push/PR — build `-tags release`, `go test -race` (600s timeout), frontend lint/build, Playwright e2e-smoke.

**Deploy** (`.github/workflows/deploy.yml`): запускается **только после успешного CI** на `master` (`workflow_run`):
1. SSH на сервер → `git reset --hard origin/master`
2. Pull образов из GHCR (тег = SHA коммита)
3. `docker compose up --force-recreate` backend, livekit, frontend, caddy
4. Post-deploy: `curl` health endpoint
5. Без `docker system prune --volumes` (безопасный prune только dangling images)

Backend prod: `Dockerfile` собирает с `-tags release` (без `/dev/*` routes).

Workflow: `.github/workflows/deploy.yml`

Необходимые GitHub Secrets:
- `SSH_HOST` — IP сервера
- `SSH_USER` — пользователь (root)
- `SSH_KEY` — приватный SSH ключ
