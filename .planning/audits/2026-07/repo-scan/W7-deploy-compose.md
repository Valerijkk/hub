# W7 — deploy / compose / Caddy / K3s (repo-scan 2026-07-23)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

**Роль:** devops-engineer (Grok 4.5)
**Scope:** `deployments/` — `docker-compose*.yml`, Caddy, K3s, SeaweedFS pin, LiveKit secrets
**Режим:** **readonly** — код и конфиги **не менялись**
**Дата:** 2026-07-23 01:08 UTC+3

## Verdict

| Метрика | Значение |
|--------|----------|
| CRITICAL | 0 |
| HIGH | 3 |
| MEDIUM | 8 |
| LOW / accepted | 5 |
| **verified** | **false** — открытые HIGH/MEDIUM |
| Активный prod-путь | Docker Compose (`docker-compose.prod.yml`) — K3s partial (TD-04) |

**Кратко:** инварианты TZ по **SeaweedFS pin `3.97`**, **`DEV_MODE: "false"`**, **LiveKit-секреты из env (heredoc, не mount plaintext)**, **Caddy prod `/ws` HTTP/1.1** — **PASS**. Блокеров уровня CRITICAL в текущем compose-prod нет. HIGH — слабые S3-креды в git + mount в prod, footgun NetworkPolicy для K3s, публикация портов monitoring на host.

---

## Checklist (Hub § devops / TZ)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| SeaweedFS **не** `:latest` | **PASS** | `docker-compose.yml:52`, `docker-compose.prod.yml:95` → `chrislusf/seaweedfs:3.97` |
| Остальные инфра-образы pinned | **PASS** (с оговорками) | postgres:17, mongo:8.0, redis:7.4-alpine, nats:2.11, livekit:v1.8; monitoring — semver tags |
| App-образы backend/frontend | **PARTIAL** | default `${IMAGE_*}:-…:latest` (`docker-compose.prod.yml:172,237,261`) |
| `DEV_MODE` в prod | **PASS** | литерал `DEV_MODE: "false"` (`docker-compose.prod.yml:205`) — не `${DEV_MODE}` |
| LiveKit prod: секреты не в файле репо | **PASS** | heredoc `/tmp/livekit.yaml` из `LIVEKIT_API_KEY`/`SECRET` (`prod.yml:127-152`); в tree только `.example` + dev `livekit.yaml` |
| `livekit-prod.yaml` с plaintext | **PASS** (tree) | `git ls-files` — файла нет; только `livekit-prod.yaml.example` |
| `livekit-prod.yaml` в `.gitignore` | **FAIL** | нет записи (см. W2-calls M-01) — риск повторного commit |
| Caddy = reverse proxy (не nginx) | **PASS** | `caddy:2-alpine` + `Caddyfile` / `Caddyfile.prod` |
| Prod `/ws` → HTTP/1.1 | **PASS** | `Caddyfile.prod:10-15` `versions 1.1` |
| Dev `/ws` → HTTP/1.1 | **FAIL** | `Caddyfile:14-16` — без `transport http { versions 1.1 }` |
| Prod маршруты `/api` `/ws` `/lk` SPA | **PASS** | `Caddyfile.prod` |
| Redis auth в prod | **PASS** | `--requirepass ${REDIS_PASSWORD}` |
| NATS auth в prod | **FAIL** | `nats:2.11` без credentials (`prod.yml:74-79`) |
| K3s full prod manifests | **PARTIAL** | только `k3s/monitoring/networkpolicy-metrics.yaml`; architecture TD-04 Compose-first |
| `.env` не в git | **PASS** | `.gitignore` → `deployments/.env` |
| Hardcoded prod secrets в compose | **PASS** | значения через `${…}`; исключения — см. seaweedfs-s3 / livekit.yaml (dev) |

---

## Findings

| Sev | ID | Файл:строка | Gap | Рекомендация |
|-----|-----|-------------|-----|--------------|
| **HIGH** | H-01 | `seaweedfs-s3.json:7-8` + `docker-compose.prod.yml:99-102` | В репо закоммичены слабые S3-креды `hub` / `hubsecret`; **prod монтирует** этот файл `:ro`. Смена только `S3_*` в `.env` без правки JSON → auth fail; оставить defaults → известные креды из git на S3 Admin. | Вынести identities в env-генерируемый конфиг (как LiveKit heredoc) или `seaweedfs-s3.json.example` + gitignored runtime file; синхронизировать с `S3_ACCESS_KEY`/`S3_SECRET_KEY`; ротировать `hubsecret` если когда-либо деплоилось as-is. |
| **HIGH** | H-02 | `k3s/monitoring/networkpolicy-metrics.yaml:13-25` | Единственный Ingress-rule: только `app=prometheus` → `:3000`. При apply **без** других NP на pod — **deny-all** остального ingress (Caddy/frontend/clients). Path `/metrics` NP **не** фильтрует. README предлагает apply as-is. | Добавить allow-rules для ingress-controller/Caddy (или default-allow + отдельный metrics-only sidecar); либо явный WARNING «не apply без complementary policies»; path-filter — на уровне app (`METRICS_ALLOWED_CIDRS`). |
| **HIGH** | H-03 | `docker-compose.monitoring.yml:25-26,46-47,59-60,91-92` | Overlay публикует Grafana `:3001`, Prometheus `:9090`, Alertmanager `:9093`, Loki `:3100` на **host**. На VPS без firewall — scrape/UI снаружи. Default Grafana `changeme` (`:85`). | Убрать `ports:` в prod-overlay (только internal network) или bind `127.0.0.1:`; обязать `GRAFANA_ADMIN_PASSWORD`; не деплоить overlay на публичный интерфейс. |
| **MEDIUM** | M-01 | `docker-compose.prod.yml:172,237,261` | Default image tag `:latest` для backend/hub-mail/frontend, если `IMAGE_*` не задан. | Default на digest/SHA; в deploy.yml всегда прокидывать pinned SHA (уже паттерн — закрепить в compose comment + fail если unset в prod). |
| **MEDIUM** | M-02 | `.gitignore` (нет `livekit-prod.yaml`) | Файл отсутствует в tree, но **не** игнорируется → повтор SEC-03 возможен. | Добавить `deployments/livekit-prod.yaml` в `.gitignore` (канон — heredoc). |
| **MEDIUM** | M-03 | `docker-compose.prod.yml:119-122` | LiveKit публикует `7880/tcp`, `7881/tcp`, UDP `50000-50100` на host **параллельно** Caddy `/lk/*`. UDP нужен для RTC; TCP signaling — обход Caddy/TLS-терминации. | Оставить UDP; TCP 7880/7881 — только internal network, снаружи через Caddy `/lk` (или TURN/ICE docs). |
| **MEDIUM** | M-04 | `docker-compose.prod.yml:74-79` | NATS JetStream без auth/TLS внутри compose-сети. Компрометация любого контейнера = pub/sub. | `--user`/`--pass` или NKey; URL с credentials в backend env. |
| **MEDIUM** | M-05 | `Caddyfile:14-16` | Dev Caddy без `versions 1.1` на `/ws` — риск 400/426 при proxy через Caddy (prod уже исправлен). | Скопировать блок transport из `Caddyfile.prod`. |
| **MEDIUM** | M-06 | `docker-compose.monitoring.yml:63-71` | Promtail: mount `/var/run/docker.sock` — широкий доступ к Docker API с хоста. | Документировать риск; ограничить socket proxy / rootless; не на shared multi-tenant host. |
| **MEDIUM** | M-07 | `docs/architecture.md` TD-04 · `deployments/k3s/` | TZ: «K3s для prod»; факт — Compose-first, K3s = 1 NP. Drift ожиданий ops. | ADR/STATE: Compose = текущий prod-target; K3s roadmap отдельным milestone; не обещать полный K3s без манифестов. |
| **MEDIUM** | M-08 | `docker-compose.prod.yml:180` | `sslmode=disable` на Postgres URL (внутри docker network). | Accepted для single-host compose; для K3s/multi-host — require TLS + `verify-full`. |
| **LOW** | L-01 | `Caddyfile:10` vs `Caddyfile.prod:18` | Dev путь `/livekit/*`, prod `/lk/*` — DX/docs путаница. | Унифицировать префикс или явно в QUICKSTART. |
| **LOW** | L-02 | `caddy:2-alpine` (все compose) | Floating minor `2` (не `2.x.y`). | Pin `caddy:2.9.x-alpine` (или актуальный patch). |
| **LOW** | L-03 | `docker-compose.test.yml:20` | `mongo:8` vs prod/dev `mongo:8.0` — лёгкий drift. | Выровнять на `mongo:8.0`. |
| **LOW (accepted)** | A-01 | `livekit.yaml:10-12` | Dev-ключи в репо (`devkey-hub-local-…`). | Accepted local-only; prod ValidateProductionConfig режет devkey. |
| **LOW (accepted)** | A-02 | `docker-compose.yml` Redis/Seaweed без auth | Локальный dev. | Не переносить паттерн в prod (prod Redis/S3 уже с паролем/ключом). |

---

## Detail: SeaweedFS pin

| Compose | Image | Pin |
|---------|-------|-----|
| `docker-compose.yml` | `chrislusf/seaweedfs:3.97` | ✅ |
| `docker-compose.prod.yml` | `chrislusf/seaweedfs:3.97` | ✅ (+ комментарий LO-Inf про бывший `:latest`) |
| `docker-compose.test.yml` | — | N/A (S3 не в test-стеке) |

История: `09d3d8e` / `d0fea66` — pin после запрета `:latest`. **TZ forbidden `seaweedfs:latest` — соблюдён.**

---

## Detail: LiveKit secrets

| Артефакт | Роль | Секреты |
|----------|------|---------|
| `livekit.yaml` | dev only, volume-mount | hardcoded local keys (accepted) |
| `livekit-prod.yaml.example` | справочник | `${LIVEKIT_API_KEY}` placeholders |
| `docker-compose.prod.yml` livekit service | prod канон | heredoc + env; webhook → `http://backend:3000/api/calls/webhook`; `departure_timeout: 120` (INF-01) |
| `livekit-prod.yaml` | — | **отсутствует** в tree |

→ SEC-03 класс (plaintext prod yaml) в текущем дереве **закрыт**. Остаточный риск — отсутствие gitignore (M-02).

---

## Detail: Caddy

| | Dev (`Caddyfile`) | Prod (`Caddyfile.prod`) |
|--|-------------------|-------------------------|
| Listen | `:80` | `770470.cloud4box.ru` (+ LE TLS) |
| API | `localhost:3000` | `backend:3000` |
| WS HTTP/1.1 | ❌ | ✅ |
| LiveKit path | `/livekit/*` | `/lk/*` + `strip_prefix` |
| SPA | Vite `:5173` | `frontend:80` |
| Порядок handle | fallback last ✅ | fallback last ✅ |

---

## Detail: K3s

```
deployments/k3s/
└── monitoring/networkpolicy-metrics.yaml   # единственный манифест
```

- Namespace `hub`, selector `app=hub-backend`.
- Нет Deployment/Service/Ingress/Secret манифестов приложения.
- Согласовано с `docs/architecture.md`: Compose-first, TD-04.
- **Не применять** NP в одиночку на живой кластер (H-02).

---

## Image inventory (prod compose)

| Service | Image | Note |
|---------|-------|------|
| postgres | `postgres:17` | pinned major |
| mongo | `mongo:8.0` | pinned |
| redis | `redis:7.4-alpine` | pinned |
| nats | `nats:2.11` | pinned |
| seaweedfs | `chrislusf/seaweedfs:3.97` | pinned ✅ |
| livekit | `livekit/livekit-server:v1.8` | pinned |
| caddy | `caddy:2-alpine` | floating minor (L-02) |
| backend / hub-mail | `ghcr.io/…/backend:latest` default | M-01 |
| frontend | `ghcr.io/…/frontend:latest` default | M-01 |

Monitoring overlay: prometheus v3.2.1, alertmanager v0.28.1, loki/promtail 3.4.2, grafana 11.5.2 — **pinned**.

---

## Verification

| Этап | Результат |
|------|-----------|
| 1. Scope ограничен `deployments/` | ✅ |
| 2. Readonly — правок нет | ✅ |
| 3. DEV_MODE / seaweedfs pin / LiveKit heredoc | ✅ проверено по файлам |
| 4. grep secrets / `:latest` | ✅ |
| 5. `git ls-files deployments/` | ✅ tree сверен |
| 6. Применено / smoke / rollback | N/A (audit-only) |
| 7. CI не гонялся | N/A |
| **verified** | **false** |

---

## Handoff

→ `team-lead` / волна 7 (CI · docs-drift · TZ-grep)
→ при фиксе H-01/H-02/H-03: `security-auditor` + `code-reviewer`
→ FA-06: если `hubsecret` когда-либо был в prod — ротация S3 keys в логе PG-03

**Код не менялся.** Отчёт: `.planning/audits/2026-07/repo-scan/W7-deploy-compose.md`
