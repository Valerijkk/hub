# Observability / Ops (logs, metrics, traces, health, deployments)

> READ-ONLY аудит зоны A14 · 2026-07-26 · ветка `cursor-grok-4.5`  
> Объём: `backend/internal/platform/{observability,middleware,config,nats}`, `backend/cmd/server/wiring_{bootstrap,fiber,run,infra,messenger}.go`, `backend/cmd/hub-mail`, `backend/internal/{auth/magiclink,status}`, `deployments/{docker-compose*.yml,monitoring/**,.env.example,seaweedfs-s3.json,livekit.yaml,Caddyfile.prod,k3s/monitoring}`.

## Что проверено

| Область | Фокус |
|---------|--------|
| Логи | JSON `slog`, уровни, access-log, PII (email/OTP), query strip, stack traces |
| Метрики | Prometheus `/metrics`, CIDR allowlist, HTTP/Timed collectors, alerts |
| Трейсинг | OTEL InitTracing, HTTP/consumer spans, NATS propagation, prod env wiring |
| Health | `/api/health`, `/api/status`, Docker healthcheck, live vs ready |
| Graceful shutdown | `signal.NotifyContext`, `app.Shutdown`, cleanups LIFO, hub-mail |
| Deployments | secrets, resource limits, restart, volumes, local↔prod diffs, monitoring overlay |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A14-001 | P1 | `backend/cmd/server/wiring_messenger.go:33-74` (+ `deployments/docker-compose.prod.yml:221-225`) | `/api/health` пингует PG/Redis/Mongo и отдаёт **503** при любой деградации; комментарий «K8s liveness»; тот же URL — Docker `healthcheck` backend | Liveness, завязанный на зависимости → при кратком сбое Redis/PG оркестратор убивает здоровый процесс → restart storm / потеря in-flight. Это семантика **readiness**, не liveness | Разделить: `/api/live` (process up → 200) и `/api/ready` (deps → 200/503). Docker/K8s: liveness=`/live`, readiness=`/ready` |
| A14-002 | P1 | `deployments/seaweedfs-s3.json:7-8` (+ `docker-compose.prod.yml:99-102`; evidence `config.go:233-238`) | Prod монтирует committed `seaweedfs-s3.json` с `hub`/`hubsecret`; `config.validate` **запрещает** эти S3-дефолты при `DEV_MODE=false` | Backend с валидным prod `.env` не сможет аутентифицироваться в SeaweedFS (identity только hub/hubsecret), либо оператор вынужден оставить запрещённые ключи | Генерировать s3.json из env при старте (как LiveKit yaml) **или** volume с templated secrets; убрать hardcoded secret из git |
| A14-003 | P1 | `deployments/docker-compose.prod.yml:74-81` (+ `nats/client.go:38-41`) | NATS JetStream в prod без auth (`nats://nats:4222`); клиент коннектится без credentials | Любой контейнер в сети compose читает/пишет EVENTS stream (outbox, notifications, webhooks) | Включить NATS auth/nkey/TLS; URL с credentials; сеть/ACL |
| A14-004 | P1 | `deployments/docker-compose.monitoring.yml:25-26,46-47,59-60,91-92` | Prometheus/Alertmanager/Loki/Grafana публикуют порты на хост (`9090`,`9093`,`3100`,`3001`) | На VPS с public IP метрики, логи и admin UI доступны без reverse-proxy/ACL | Убрать `ports:` (только internal network) или bind `127.0.0.1:` + VPN/SSH tunnel |
| A14-005 | P1 | `backend/internal/platform/config/config.go:147` (+ `docker-compose.prod.yml:206`; `prometheus.go:71-88`) | `METRICS_ALLOWED_CIDRS` по умолчанию пуст → `/metrics` без IP-фильтра | Любой peer в Docker/K8s сети scrape'ит метрики (cardinality + info disclosure). NetworkPolicy есть только как optional K3s-манифест | Prod default: CIDR overlay/Prometheus; enforce non-empty в `ENVIRONMENT=production`; применить `k3s/monitoring/networkpolicy-metrics.yaml` |
| A14-006 | P2 | `backend/internal/auth/magiclink.go:92-96` (+ `service_login.go:61`) | В логах полный `email`; при `resend == nil` — ещё и OTP `code` | PII/OTP в stdout → Loki/Promtail (retention 7d, `auth_enabled: false`) = утечка учётных данных и нарушение минимизации данных | Хэшировать/маскировать email; OTP никогда в лог (даже dev — stderr secure channel); pipeline redact в Promtail |
| A14-007 | P2 | `backend/cmd/server/wiring_fiber.go:41` | `fiberrecover` с `EnableStackTrace: true` всегда | Stack traces в access/error логах на каждый panic — шум + потенциальная утечка путей/данных запроса в централизованные логи | `EnableStackTrace: cfg.Server.DevMode` (или только non-production) |
| A14-008 | P2 | `backend/internal/platform/middleware/logging.go:33-40` | Access-log через `slog.Info` **без** `InfoContext`; нет `trace_id`/`span_id` | Корреляция HTTP↔OTEL↔Timed ломается: в Loki нельзя join по trace; request_id есть, trace — нет | `slog.InfoContext(c.UserContext(), …)` + поля из `trace.SpanFromContext` |
| A14-009 | P2 | `deployments/docker-compose.prod.yml:176-207` (+ `.env.example:63-64`; `otel.go:42-49`) | В prod backend env **нет** `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_SERVICE_NAME` | `InitTracing` no-op без endpoint → HTTP/consumer/NATS spans мёртвы в prod, хотя код и `.env.example` обещают OBS | Прокинуть OTEL_* в compose; добавить Tempo/collector в monitoring overlay или задокументировать external exporter |
| A14-010 | P2 | `backend/cmd/server/wiring_run.go:34` (+ отсутствие `stop_grace_period` в `docker-compose.prod.yml`) | `app.Shutdown()` без timeout/context; compose не задаёт `stop_grace_period` | Зависший client/WS может держать stop дольше Docker default 10s → SIGKILL mid-drain; cleanups/trace flush не успевают | `ShutdownWithContext` (или deadline 15–25s) + `stop_grace_period: 30s` на backend/hub-mail |
| A14-011 | P2 | `backend/cmd/server/wiring_messenger.go:26-31` | Mongo connect без `addCleanup(Disconnect)` (см. также A8-007) | PG/Redis/NATS закрываются LIFO; Mongo pool живёт до kill — неполный graceful shutdown | `d.addCleanup(func() { _ = mongoDB.Client().Disconnect(...) })` до process exit |
| A14-012 | P2 | `deployments/.env.example:21` (+ `config.go:19-21,234-236`) | Example задаёт `S3_ACCESS_KEY=hub` — это **dev-default**, отвергаемый `validate()` | Копирование `.env.example` → prod start fail или соблазн ослабить validate | `S3_ACCESS_KEY=CHANGE_ME_NON_DEFAULT_KEY`; синхронизировать с A14-002 |
| A14-013 | P2 | `deployments/docker-compose.monitoring.yml:84-85` (+ `loki/loki-config.yml:4`) | Grafana `GF_SECURITY_ADMIN_PASSWORD` default `changeme`; Loki `auth_enabled: false` | Вместе с A14-004 — открытый admin + unauthenticated log store | Обязательный сильный пароль без default; Loki auth / network isolation; не публиковать 3100/3001 |
| A14-014 | P2 | `deployments/docker-compose.monitoring.yml:71` | Promtail монтирует `/var/run/docker.sock:ro` | Компрометация Promtail ≈ read metadata/logs всех контейнеров (часто step к host escape) | Rootless/journald scrape, ограничить sock proxy, или sidecar log driver без sock |
| A14-015 | P2 | `backend/cmd/hub-mail/main.go:42` | `_ = app.Shutdown()` — ошибка drain игнорируется; нет JSON logger/metrics/tracing | Несимметричный ops-контракт vs main backend; сбой shutdown незаметен | Логировать err; тот же slog JSON + optional health уже есть; timeout shutdown |
| A14-016 | P3 | `backend/internal/platform/observability/timing.go:27` (+ `wiring_bootstrap.go:21`) | `Timed` пишет **Info** на каждый success; уровень захардкожен `LevelInfo`, нет `LOG_LEVEL` | Hot-path (`tracker.ListIssues`, `audit.RecordHTTPRequest`, …) раздувает Loki/диск | Success → Debug; Error остаётся; `LOG_LEVEL` из env |
| A14-017 | P3 | `backend/cmd/server/wiring_messenger.go:71-74` (+ `Caddyfile.prod:6-8`) | Публичный `/api/*` проксирует health; тело отдаёт `services: {postgres,redis,mongo}` | Инфодискложер топологии/аутэжа для анонимов (удобно атакующему) | На внешнем health — только `status`; детали deps — internal/admin |
| A14-018 | P3 | `deployments/docker-compose.yml:7-84` vs `docker-compose.prod.yml` | Local: нет `restart`/`mem_limit`/`cpus`, Redis без пароля, Mongo без healthcheck, порты БД наружу | Ожидаемо для dev; риск «поднимим compose.yml на сервере» как prod | Явный banner в шапке local-файла; CI/docs: только `*.prod.yml` для серверов |
| A14-019 | P3 | `wiring_messenger.go:34-65` vs `77-86` | `/api/health` не пингует NATS; `/api/status` — пингует | Ложное «ok» при мёртвом NATS (outbox/consumers down), пока status page уже `outage` | Включить NATS в ready-probe (после split A14-001) |

## Чисто

- **JSON slog** с полем `service` (`wiring_bootstrap.go:16-23`); access-log **stripQuery** против JWT в `?token=` (`logging.go:11-19,35`).
- **Prometheus**: HTTP + Timed collectors; `/metrics` до rate-limit (`wiring_fiber.go:46-47`); CIDR deny-all при битых CIDR (`prometheus.go:75-85`).
- **OTEL**: optional InitTracing + HTTP middleware + consumer spans + NATS inject/extract; shutdown cleanup 5s (`otel.go`, `wiring_bootstrap.go:81-95`).
- **Graceful root**: единый `signal.NotifyContext` → cancel outbox/cron/consumers/WS; LIFO `runCleanups` (`wiring_bootstrap.go:66-79`, `wiring_deps.go:124-127`).
- **Prod compose hygiene**: `restart: unless-stopped`, `mem_limit`/`cpus` + `deploy.resources`, named volumes для PG/Mongo/Redis/NATS/Seaweed/Caddy, `DEV_MODE: "false"`, secrets из env (не plaintext в yaml), pinned SeaweedFS `3.97`.
- **Monitoring**: scrape config, alert rules (5xx/latency/ops/up), Loki retention 7d, Promtail JSON parse `request_id`/`level`.
- **Caddy.prod** не проксирует `/metrics` наружу (только `/api/*`, `/ws`, `/lk/*`) — снижает риск A14-005 с edge.

## Оценка зоны: 6/10

**Обоснование:** Observability-стек в коде зрелый (JSON logs, Prometheus, optional OTEL, CIDR hook, monitoring overlay, rootCtx shutdown). Оценка снижена **ops-дырами prod-compose**: health=liveness с 503 по deps, SeaweedFS secret file vs validate, NATS без auth, открытые порты monitoring, metrics allowlist пуст по умолчанию, PII/OTP в логах, OTEL не прокинут в prod env. После split live/ready, templated S3 creds, NATS auth и закрытия monitoring ports зона может выйти на ~8/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 5 |
| P2 | 10 |
| P3 | 4 |
| **Total** | **19** |
| **Score** | **6/10** |
