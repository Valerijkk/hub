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

### 3. Запустить
```bash
docker compose -f docker-compose.prod.yml up -d --build
```

### 4. Проверить
```bash
# Все контейнеры запущены?
docker compose -f docker-compose.prod.yml ps

# Health check
curl http://localhost/api/health
```

### 5. Засидить демо-данные (опционально)
```bash
curl -X POST http://localhost/api/dev/seed
```

## Переменные окружения

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://hub:pass@postgres:5432/hub` |
| `MONGODB_URL` | MongoDB connection string | `mongodb://hub:pass@mongo:27017` |
| `REDIS_URL` | Redis connection string | `redis://redis:6379/0` |
| `NATS_URL` | NATS connection string | `nats://nats:4222` |
| `JWT_SECRET` | Секрет для JWT подписи | `random-secret-32-chars` |
| `DEV_MODE` | Режим разработки | `true` / `false` |
| `CORS_ORIGINS` | Разрешённые origins | `http://example.com` |
| `RESEND_API_KEY` | API ключ Resend (для email) | `re_xxx` |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram бота | `123:ABC` |

## CI/CD (GitHub Actions)

При каждом push в `master`:
1. Сервер делает `git pull`
2. Docker Compose пересобирает изменённые образы
3. Перезапускает контейнеры без простоя

Workflow: `.github/workflows/deploy.yml`

Необходимые GitHub Secrets:
- `SSH_HOST` — IP сервера
- `SSH_USER` — пользователь (root)
- `SSH_KEY` — приватный SSH ключ
