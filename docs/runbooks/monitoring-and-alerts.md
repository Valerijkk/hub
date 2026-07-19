# Runbook — мониторинг и алерты

> **Кто выполняет:** дежурный. **Когда:** прилетел алерт, либо нужно проверить состояние системы.
> **Источник правды по правилам:** [`deployments/monitoring/prometheus/alerts/hub.yml`](../../deployments/monitoring/prometheus/alerts/hub.yml) — если правило изменилось там, обнови этот документ в том же PR.

---

## 1. Стек наблюдаемости

| Компонент | Роль | Конфиг |
|-----------|------|--------|
| **Prometheus** | Сбор метрик, вычисление алертов | `deployments/monitoring/prometheus.yml`, правила — `prometheus/alerts/hub.yml` |
| **Alertmanager** | Маршрутизация/группировка алертов, уведомления | `monitoring/alertmanager/alertmanager.yml` (прод-пример: `alertmanager.prod.example.yml`) |
| **Loki** | Хранение логов | `monitoring/loki/loki-config.yml` |
| **Promtail** | Доставка логов контейнеров в Loki | `monitoring/promtail/promtail-config.yml` |
| **Grafana** | Дашборды поверх Prometheus + Loki | дашборд `monitoring/grafana/dashboards/hub-observability.json` |

Запуск overlay-стека мониторинга поверх прода:

```bash
cd /opt/hub/deployments
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
```

**Scrape-таргет:** job `hub-backend`, `backend:3000/metrics` (`prometheus.yml`). Метрики отдаёт `backend/internal/platform/observability/prometheus.go`.

> `TODO(ops)`: зафиксировать URL Grafana, учётку доступа и куда Alertmanager реально шлёт уведомления (канал/почта) — из репозитория это не выводится.

---

## 2. Ключевые метрики приложения

| Метрика | Смысл | Где используется |
|---------|-------|------------------|
| `hub_http_requests_total{status}` | Счётчик HTTP-запросов по статусам | Алерт 5xx-rate |
| `hub_http_request_duration_milliseconds_bucket` | Гистограмма латентности | Алерт p95 |
| `hub_operation_calls_total{operation,result}` | Доменные операции и их исход | Алерт error-rate по операциям |
| `up{job="hub-backend"}` | Доступность scrape-таргета | Алерт «метрики недоступны» |

---

## 3. Алерты: что значит и что делать

> Все пороги ниже — фактические из `alerts/hub.yml`. Изменение порога в проде без обновления этой таблицы — рассинхрон документации (запрещено `.cursor/rules/55-docs-sync.mdc`).

### 3.1 `HubBackendMetricsDown` — severity **critical**

**Правило:** `up{job="hub-backend"} == 0` в течение **2 минут**.
**Значит:** Prometheus не может забрать `/metrics` с `backend:3000`. Либо backend лежит, либо сеть/скрейп сломаны.

☐ Проверить, жив ли backend вообще:

```bash
docker compose -f docker-compose.prod.yml ps backend
curl -sf http://127.0.0.1:3000/api/health | head -c 200; echo
```

☐ Если health не отвечает → это инцидент сервиса: [incident-response.md](incident-response.md) §4.1. Алерт вторичен.
☐ Если health отвечает, а метрики нет — проблема в скрейпе:

```bash
docker compose -f docker-compose.prod.yml exec prometheus wget -qO- http://backend:3000/metrics | head -5
```

☐ Пусто/недоступно из контейнера Prometheus → сетевая проблема между сервисами (проверить, что оба в одной compose-сети).
☐ `TODO(ops)`: если применяется NetworkPolicy K3s (`deployments/k3s/monitoring/networkpolicy-metrics.yaml`) — проверить, что она не блокирует scrape.

**Ложное срабатывание:** во время деплоя backend перезапускается — алерт с `for: 2m` может сработать на долгом рестарте. Сверься с временем деплоя ([deploy-and-rollback.md](deploy-and-rollback.md)); если совпало и сервис поднялся — инцидента нет, но запиши факт (частые срабатывания = слишком долгий рестарт, стоит разобраться).

### 3.2 `HubHighHTTP5xxRate` — severity warning

**Правило:** доля 5xx > **1%** от всех запросов, окно 5 минут, `for: 5m`.
**Значит:** приложение systematically отдаёт ошибки. 1% — это уже заметно пользователям.

☐ Определить, какие эндпоинты сыпятся — по логам backend за окно алерта:

```bash
docker compose -f docker-compose.prod.yml logs --since=15m backend | grep -iE '"status":5[0-9]{2}|level=error' | tail -50
```

☐ В Grafana/Loki — разбивка 5xx по маршрутам (дашборд `hub-observability`).
☐ Один эндпоинт → баг в конкретном домене; все эндпоинты → отказ общей зависимости (БД, Redis, NATS) → [database-operations.md](database-operations.md) §4.
☐ Совпало с деплоем → ⚠️ откат ([deploy-and-rollback.md](deploy-and-rollback.md) §4).
☐ Не совпало ни с чем → искать внешнюю зависимость (S3/SeaweedFS, LiveKit, Resend) и ресурсы хоста.

**Порог осознанно низкий:** 1% 5xx в норме быть не должно. Если алерт «шумит» на легитимных 5xx (например, health-проверки внешнего сервиса) — чинить источник, а не поднимать порог.

### 3.3 `HubHighHTTPLatencyP95` — severity warning

**Правило:** p95 латентности > **2000 мс**, окно 5 минут, `for: 5m`.
**Значит:** каждый двадцатый запрос дольше 2 секунд — пользователи это чувствуют.

☐ Проверить нагрузку и ресурсы:

```bash
docker stats --no-stream
```

☐ Проверить БД как самого частого виновника медленных ответов: [database-operations.md](database-operations.md) §3 (медленные запросы, блокировки).
☐ Проверить, не деградировала ли внешняя зависимость (таймауты к S3/LiveKit/почте дают рост p95 без роста ошибок).
☐ Проверить, не появился ли N+1 после релиза — классическая причина «после деплоя стало медленно» (см. `prompt-engineering/prompts/11-performance-security-testing.md` #14).

**Важно:** p95 растёт и от нагрузки, и от деградации. Отличать по RPS: выросли оба — вероятно, нагрузка; вырос только p95 — деградация.

### 3.4 `HubOperationErrorRate` — severity warning

**Правило:** для конкретной `operation` доля `result="error"` > **5%**, окно 5 минут, `for: 10m`.
**Значит:** доменная операция стабильно падает. Алерт несёт метку `{{ $labels.operation }}` — она сразу указывает домен.

☐ Взять `operation` из тела алерта — это точка входа в диагностику.
☐ Логи по этой операции:

```bash
docker compose -f docker-compose.prod.yml logs --since=20m backend | grep -i '<operation>' | tail -50
```

☐ Определить характер ошибок: отказ зависимости, ошибка валидации (тогда, возможно, изменился клиент), или баг после релиза.
☐ `for: 10m` намеренно длиннее остальных — операция может кратковременно шуметь; сработавший алерт означает устойчивую проблему, не всплеск.

---

## 4. Регулярные проверки (не по алерту)

| Периодичность | Что проверить | Как |
|---------------|---------------|-----|
| Ежедневно | Все контейнеры `Up`, health зелёный | `docker compose ps` + `curl /api/health` |
| Ежедневно | Место на диске | `df -h` (>85% — планировать чистку, не ждать инцидента) |
| Еженедельно | Тренд p95 и 5xx в Grafana | дашборд `hub-observability` |
| Еженедельно | Сработавшие алерты за неделю | Alertmanager: были ли проигнорированы? |
| Ежемесячно | Бэкапы восстанавливаются | [backup-restore.md](backup-restore.md) §4 — **проверка восстановлением**, не наличием файла |
| Ежемесячно | Свежесть образов/зависимостей, CVE | Dependabot PR, `docker compose images` |

---

## 5. Работа с алертами как с сигналом

- **Алерт сработал и оказался ложным** — это дефект правила, а не «норма». Либо чинить источник шума, либо править правило в `alerts/hub.yml` **с обоснованием в PR**. Молча игнорировать — путь к тому, что алерты перестают читать.
- **Инцидент случился без алерта** — пробел мониторинга. По итогам постмортема ([incident-response.md](incident-response.md) §7) добавить правило, которое поймает **класс** проблемы.
- **Порог меняется** — обновить `alerts/hub.yml` **и** §3 этого документа в одном PR (docs-sync, rule 55).

---

## 6. Добавление нового алерта

☐ Сформулировать: какой пользовательский симптом ловим (не «метрика X выросла», а «пользователи не могут Y»).
☐ Выбрать метрику и порог; порог обосновать данными (текущий baseline из Grafana), не «на глаз».
☐ Задать `for:` — защита от всплесков; чем дороже ложное срабатывание, тем длиннее.
☐ Добавить правило в `deployments/monitoring/prometheus/alerts/hub.yml` с русскоязычными `summary`/`description` (как у существующих).
☐ Проверить синтаксис и что правило вообще может сработать:

```bash
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml exec prometheus \
  promtool check rules /etc/prometheus/alerts/hub.yml
```

☐ Описать в §3 этого документа: что значит, что делать по шагам, какие ложные срабатывания возможны.
☐ Перезагрузить конфиг Prometheus (без рестарта, если включён `--web.enable-lifecycle` — `TODO(ops)` подтвердить):

```bash
curl -X POST http://127.0.0.1:9090/-/reload
```

**Правило:** алерт без runbook-секции «что делать» — это будильник без инструкции. Не добавлять.
