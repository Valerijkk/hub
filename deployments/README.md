# Деплой и инфраструктура Hub

Каталог `deployments/` содержит конфигурации Docker Compose, Caddy, LiveKit, мониторинга и манифесты K3s. Логика приложения здесь не меняется — только описание окружения.

## Структура

| Файл / каталог | Назначение |
|----------------|------------|
| `docker-compose.yml` | Инфраструктура для локальной разработки (БД, Redis, NATS, SeaweedFS, LiveKit, Caddy) |
| `docker-compose.prod.yml` | Полный production-стек: инфра + backend + hub-mail + frontend + Caddy |
| `docker-compose.monitoring.yml` | Overlay: Prometheus, Loki, Promtail, Alertmanager, Grafana |
| `Caddyfile` | Reverse proxy для dev (localhost) |
| `Caddyfile.prod` | Reverse proxy для production с TLS |
| `livekit.yaml` | Конфиг LiveKit для локальной разработки |
| `livekit-prod.yaml.example` | Справочник production-конфига (реальный конфиг генерируется из env) |
| `.env.example` | Шаблон переменных окружения для production |
| `seaweedfs-s3.json.example` | Справочник S3-identity; **prod генерирует s3.json из `S3_*` env** (A14-002) |
| `monitoring/` | Prometheus, Loki, Promtail, Alertmanager, Grafana provisioning |
| `k3s/` | Kubernetes-манифесты (NetworkPolicy и др.) |

## Локальная разработка

1. Поднять полный локальный стек, включая backend и frontend:

```bash
docker compose -f docker-compose.local.yml up -d
```

2. Caddy проксирует:
   - `/api/*`, `/ws` → backend `:3000`
   - `/livekit/*` → LiveKit `:7880`
   - остальное → Vite `:5173`

## Production (Docker Compose)

1. Скопировать и заполнить секреты:

```bash
cp deployments/.env.example deployments/.env
# отредактировать .env — пароли, JWT, LiveKit, CORS и т.д.
```

2. Запуск:

```bash
docker compose -f deployments/docker-compose.prod.yml --env-file deployments/.env up -d
```

3. С мониторингом:

```bash
docker compose \
  -f deployments/docker-compose.prod.yml \
  -f deployments/docker-compose.monitoring.yml \
  --env-file deployments/.env up -d
```

**Важно:**
- Production workflow `.github/workflows/deploy.yml` запускается только вручную на текущем `master` после локального Docker evidence для выбранного SHA, `FA-06/PG-03` и CTO sign-off; обычный merge не деплоит, а stale dispatch останавливается до смены checkout.
- `DEV_MODE` в prod всегда `false` (задаётся в compose).
- LiveKit-ключи только через `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` — не коммитьте `livekit-prod.yaml` с plaintext-секретами.
- SeaweedFS S3 identity — из `S3_ACCESS_KEY` / `S3_SECRET_KEY` (не committed `hub`/`hubsecret`).
- NATS в prod — с `NATS_USER` / `NATS_PASSWORD` (URL с credentials).
- Образы SeaweedFS и прочих сервисов закреплены по тегу (не `latest`).

## Caddy (production)

В `Caddyfile.prod` замените домен на свой. Caddy автоматически получит сертификат Let's Encrypt.

Маршруты:
- `/api/*` — REST API backend
- `/ws` — WebSocket (HTTP/1.1)
- `/lk/*` — LiveKit (префикс снимается)
- `/` — статика frontend (Nginx)

## Мониторинг

После overlay `docker-compose.monitoring.yml`:

| Сервис | URL (только loopback, A14-004) |
|--------|----------------|
| Grafana | http://127.0.0.1:3001 (admin / `GRAFANA_ADMIN_PASSWORD`) |
| Prometheus | http://127.0.0.1:9090 |
| Alertmanager | http://127.0.0.1:9093 |
| Loki | http://127.0.0.1:3100 |

Доступ с ноутбука: `ssh -L 3001:127.0.0.1:3001 -L 9090:127.0.0.1:9090 user@vps`.  
Порты **не** публикуются на `0.0.0.0` (раньше были открыты на public IP).

**Tracing (A14-009):** backend env `OTEL_*` прокинут в prod compose.  
Опциональный Tempo: `--profile otel` + `OTEL_EXPORTER_OTLP_ENDPOINT=http://tempo:4318`.  
Без endpoint — tracing no-op (ожидаемо).

Алерты backend описаны в `monitoring/prometheus/alerts/hub.yml`. Для production webhook скопируйте `monitoring/alertmanager/alertmanager.prod.example.yml` → `alertmanager.yml` и задайте `ALERT_WEBHOOK_SECRET` (то же значение в env backend).

## K3s

Пример NetworkPolicy для scrape `/metrics`:

```bash
kubectl apply -f deployments/k3s/monitoring/networkpolicy-metrics.yaml
```

Требуется namespace `hub`, backend с меткой `app=hub-backend`, Prometheus с `app=prometheus`.

## Безопасность

- Не коммитьте `.env` и файлы с реальными ключами.
- Ротация LiveKit: см. комментарии в `livekit-prod.yaml.example`.
- `METRICS_ALLOWED_CIDRS` ограничивает доступ к `/metrics` backend (prod compose default `127.0.0.1/32`; пусто в production → fail-start; для Prometheus overlay задайте scrape CIDR явно).
- Telegram — только уведомления; вход в Hub — Magic Link (email).

Подробнее: `docs/wiki/DEPLOY.md`, `docs/security_guidelines.md`.

## Messenger recovery deployment status (MPR-09)

As of 2026-08-31, recovery deployment is a frozen plan, not a runnable service. `backend/Dockerfile` copies only `hub-server` and `hub-mail`, production Compose has no `messenger-recovery` service, and the K3s tree has no application Deployment/Service baseline. Docker Compose is the current self-host deployment authority; K3s work is optional static manifest parity until separate cluster authority and evidence exist.

Future recovery binding is profile-disabled with empty tenant enrollment, internal-only control port and PostgreSQL/Mongo-only dependencies. Existing non-TLS internal production Compose stores are unsupported recovery endpoints. Production must provide independently named TLS-capable PostgreSQL and Mongo endpoints: PostgreSQL `require|verify-ca|verify-full` and Mongo `mongodb+srv` TLS or `mongodb://...?tls=true` with hostname/CA verification. DSNs come only from least-readable mounted secret files/platform Secret references; manifests, environment values, arguments, annotations, logs and evidence cannot contain literals. No Caddy, Mail/Calendar, Calls/Jitsi/LiveKit, Redis, NATS, S3 or product-route dependency is added.

Shared Dockerfile/Compose/workflow edits wait for the planned Mail/Calendar merge and preserve `hub-mail`. One accepted build records OCI manifest/platform-child digests, provenance, SBOM and base-image digests. Compose/K3s pin that digest; mutable tags and runtime rebuild drift are forbidden. The manual workflow pulls the digest, starts only explicitly enabled recovery with dependency recreation disabled and verifies running `RepoDigest` equality; it does not add recovery to the current backend/mail/calls force-recreate group.

Image build, Compose runtime, migration, monitoring, canary, rollback, redrive and K3s execution are `NOT_RUN/UNBOUND`. The current agent launcher authorizes product evidence only through the root QA services documented in `AGENTS.md`; static manifests cannot be promoted to production PASS.
