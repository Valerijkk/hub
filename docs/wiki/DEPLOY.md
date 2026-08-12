# Установка и деплой

> **Обновлено:** 2026-06-21 · Версии образов сверены с `deployments/docker-compose.prod.yml`

---

## Требования к серверу

| Ресурс | Минимум | Рекомендуемый |
|--------|---------|---------------|
| CPU | 2 vCPU | 4 vCPU |
| RAM | 4 GB | 8 GB |
| Disk | 40 GB SSD | 100 GB SSD |
| ОС | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |

Для e2e-full в CI: Ubuntu latest runner + Docker (postgres, redis, mongo, nats).

---

## Версии инфраструктуры (pinned)

| Сервис | Docker image | Порт(ы) |
|--------|--------------|---------|
| PostgreSQL | `postgres:17` | 5432 |
| MongoDB | `mongo:8.0` | 27017 |
| Redis | `redis:7.4-alpine` | 6379 |
| NATS JetStream | `nats:2.11` | 4222, 8222 |
| SeaweedFS | `chrislusf/seaweedfs:3.97` | 9333, 8333 |
| LiveKit | `livekit/livekit-server:v1.8` | 7880, 7881, UDP 50000–50100 |
| Caddy | `caddy:2-alpine` | 80, 443 |
| Backend | `ghcr.io/valerijkk/hub/backend:SHA` | 3000 |
| hub-mail | `ghcr.io/valerijkk/hub/backend:SHA` (entrypoint `./hub-mail`) | 3001 |
| Frontend | `ghcr.io/valerijkk/hub/frontend:SHA` | 80 |

Опционально: `deployments/docker-compose.monitoring.yml` — Prometheus v3.2.1, Loki/Grafana 11.5.2.  
Порты мониторинга bind на `127.0.0.1` только (A14-004) — доступ через SSH tunnel.

---

## Быстрый деплой (production, Docker Compose)

### 1. Docker

```bash
curl -fsSL https://get.docker.com | sh
```

### 2. Клонирование

```bash
cd /opt
git clone https://github.com/Valerijkk/hub.git
cd hub
```

### 3. Переменные окружения

```bash
cd deployments
cp .env.example .env
# Обязательно заполнить: JWT_SECRET, DATABASE_URL, INTEGRATION_ENCRYPTION_KEY,
# LIVEKIT_*, RESEND_API_KEY, CORS_ORIGINS
# DEV_MODE=false в production!
```

### 4. Запуск

```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

**LiveKit:** конфиг генерируется из env в compose (heredoc), включая `departure_timeout: 120` (INF-01).  
Не монтировать `livekit-prod.yaml` с plaintext ключами. После обновления:

```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d --force-recreate livekit
```

**OTEL / tracing (A14-009):** prod compose прокидывает `OTEL_EXPORTER_OTLP_ENDPOINT` и `OTEL_SERVICE_NAME`.  
Без endpoint — `InitTracing` no-op (код готов, spans не экспортируются).  
Варианты:
1. External collector/Tempo — задать URL в `.env`.
2. Overlay Tempo: `docker compose ... --profile otel` + `OTEL_EXPORTER_OTLP_ENDPOINT=http://tempo:4318`.

**SeaweedFS (A14-002):** S3 identity генерируется из `S3_ACCESS_KEY` / `S3_SECRET_KEY` при старте контейнера (как LiveKit).  
Не коммитьте `seaweedfs-s3.json` с реальными ключами; справочник — `seaweedfs-s3.json.example`.  
`hub` / `hubsecret` — fail-closed: `config.validate` отказывает старт при `DEV_MODE=false`.  
Residual: если эти ключи когда-либо были в живом деплое — ротировать (исторический git secret).

**NATS (A14-003):** в prod включён user/pass (`NATS_USER` / `NATS_PASSWORD`).  
`NATS_URL` для backend: `nats://USER:PASS@nats:4222`. Пароль без `@ : / ? #`.  
При `ENVIRONMENT=production` backend **не стартует**, если в `NATS_URL` нет non-empty user и password.  
После смены кредов: `docker compose ... up -d --force-recreate nats backend`.

### 5. Проверка

```bash
docker compose -f docker-compose.prod.yml ps
curl -fsS http://localhost/api/live    # liveness — всегда 200 если процесс жив
curl -fsS http://localhost/api/ready   # readiness — 200/503 по deps
```

Ожидаемый `/ready`: `{"status":"ok","services":{...}}`.  
K8s: liveness=`/api/live`, readiness=`/api/ready`. Compose `healthcheck` backend = `/api/ready`.

---

## Локальная разработка

### Полный стек в Docker

```bash
# из корня репозитория
cp .env.local.example .env.local
docker compose --env-file .env.local -f docker-compose.local.yml up --build
# UI через Caddy: http://localhost:8080
```

`DEV_MODE=true` в local compose — dev-login, `/api/dev/seed`.

### Нативный backend + Docker infra (как в CI e2e-full)

```bash
# 1. Infra
bash scripts/ci-e2e-backend.sh start-infra
bash scripts/ci-e2e-backend.sh wait-infra

# 2. Backend
export DEV_MODE=true
export DATABASE_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable
export MONGODB_URL=mongodb://hub:hub@localhost:27017/?authSource=admin
export REDIS_URL=redis://localhost:6379/0
export NATS_URL=nats://localhost:4222
export JWT_SECRET=local-dev-secret-change-in-production
bash scripts/ci-e2e-backend.sh start-backend
bash scripts/ci-e2e-backend.sh wait-backend

# 3. Frontend (Vite проксирует /api → :3000)
cd frontend && npm run dev
```

### Демо-данные

```bash
curl -X POST http://localhost:3000/api/dev/seed
# Только при DEV_MODE=true
```

---

## Переменные окружения (ключевые)

| Переменная | Описание | Пример / default |
|-----------|----------|------------------|
| `DATABASE_URL` | PostgreSQL | `postgresql://hub:pass@postgres:5432/hub` |
| `MONGODB_URL` | MongoDB | `mongodb://hub:pass@mongo:27017/?authSource=admin` |
| `MONGODB_DATABASE` | Имя БД | `hub` |
| `REDIS_URL` | Redis | `redis://redis:6379/0` |
| `NATS_URL` | NATS | `nats://nats:4222` |
| `JWT_SECRET` | JWT HS256 | random 32+ chars |
| `ENVIRONMENT` | `production` / `development` | `production` |
| `DEV_MODE` | Dev routes | **`false` в prod** |
| `CORS_ORIGINS` | Origins через запятую | `https://hub.example.com` |
| `INTEGRATION_ENCRYPTION_KEY` | AES для VCS tokens | 32 bytes base64 |
| `S3_ENDPOINT` | SeaweedFS | `http://seaweedfs:8333` |
| `RESEND_API_KEY` | Magic Link email | `re_xxx` |
| `HUB_MAIL_URL` | URL микросервиса hub-mail | `http://hub-mail:3001` |
| `HUB_MAIL_INTERNAL_TOKEN` | Shared-secret backend↔hub-mail | random |
| `RESEND_FROM_ALERTS` | From-адрес alert-писем | `alerts@hub.example.com` |
| `TELEGRAM_BOT_TOKEN` | Notify bot | `123:ABC` |
| `LIVEKIT_URL` | WebRTC | `http://livekit:7880` |
| `LIVEKIT_API_KEY` / `SECRET` | LiveKit creds | из env, не в git |

Полный список: `deployments/.env.example`, `docs/TECHNICAL.md` § Environment Variables.

---

## CI/CD (GitHub Actions)

### CI — `.github/workflows/ci.yml`

Триггеры: push/PR на `master`, `workflow_dispatch`.

| Job | Шаги |
|-----|------|
| `backend-test` | Go 1.25, Mongo 8 service, `go generate`, `go vet`, `go test -tags hubtest -race` 600s |
| `frontend-test` | Node 22, `npm ci`, `tsc`, `vitest run`, `biome check` |
| `build` | `go build -tags release`, `npm run build` |
| `e2e-smoke` | Playwright без backend: login, a11y, critical-gate |
| `e2e-full` | Infra script + backend + 16 E2E specs |

Артефакты при FAIL: `playwright-report`, `backend-e2e-log` (7 дней).

### Deploy — `.github/workflows/deploy.yml`

- Запуск: `workflow_run` после **успешного CI** на push в `master`
- GHCR images: `backend:SHA`, `frontend:SHA`
- SSH на сервер → `git reset --hard SHA` → `docker compose up --force-recreate`
- Post-deploy health check
- **Нет** `DEV_MODE=true` в workflow env

### GitHub Secrets (deploy)

| Secret | Назначение |
|--------|------------|
| `SSH_HOST` | IP сервера |
| `SSH_USER` | SSH user (например root) |
| `SSH_KEY` | Private key |

---

## Backend production image

`backend/Dockerfile`:

```dockerfile
# Сборка с -tags release — без /dev/* routes
go build -tags release -o /app/server ./cmd/server
```

---

## Caddy (production routing)

```caddyfile
hub.example.com {
  handle /api/*  { reverse_proxy backend:3000 }
  handle /ws     { reverse_proxy backend:3000 { transport http { versions 1.1 } } }
  handle /lk/*   { uri strip_prefix /lk; reverse_proxy livekit:7880 }
  handle         { reverse_proxy frontend:80 }
}
```

Файл: `deployments/Caddyfile` (или prod variant).

---

## Бэкапы

- Cron: еженедельно PG dump + Mongo dump + S3 files → ZIP в SeaweedFS
- API: `POST /api/admin/backups` (Admin)
- Restore: `POST /api/admin/backups/:id/restore`
- **Maintenance (U04 / ADR-BACKUP-MAINTENANCE-GATE-2026):** на время restore
  все реплики читают Redis-ключ `hub:backup:maintenance` и отдают 503 на
  non-health API (health/live/ready/metrics bypass). TTL lease 2h — safety
  net при crash пода без `DEL`. Multi-replica K3s **не** требует affinity.
- Retention: настраивается в admin limits

---

## Troubleshooting

| Симптом | Проверка |
|---------|----------|
| 502 на /api | `docker logs hub-backend-1`, DATABASE_URL |
| WS disconnect | Caddy WS upgrade, `CORS_ORIGINS` |
| Magic Link не приходит | `RESEND_API_KEY`, hub-mail status |
| E2E fail в CI | Артефакт `playwright-report`, `.ci-backend.log` |
| LiveKit no media | UDP ports 50000–50100, `LIVEKIT_*` keys |

---
*2026-06-21 — CI jobs, pinned images, DEV_MODE policy*
