# Наблюдаемость: справочник

> Часть плейбука. Канон и оглавление — [`../observability.md`](../observability.md).

Приложения.

---

## 12. Приложения

### 12.1 Словарь обязательных ключей полей

Единый словарь ключей — чтобы фильтры в Loki работали одинаково во всех модулях. Не изобретайте синонимы (`uid` vs `user_id`).

| Ключ | Тип | Значение | Пример |
|---|---|---|---|
| `service` | string | Имя сервиса (глобально) | `hub-backend` |
| `module` | string | Домен | `tracker`, `messenger`, `auth`, `calls`, `tms`, `ws`, `notification` |
| `request_id` | string | Сквозной id запроса | UUID |
| `user_id` | string | Инициатор действия | UUID |
| `issue_id` / `channel_id` / `message_id` / `call_id` | string | Конкретная сущность | UUID |
| `entity_id` + `entity_type` | string | Обобщённая сущность (audit/outbox) | `<uuid>` + `issue` |
| `operation` | string | Имя `Timed`-операции | `tracker.create_issue` |
| `duration_ms` / `dur_ms` | int64 | Длительность | `142` |
| `status` | int | HTTP-статус | `200`, `500` |
| `method` / `path` / `route` | string | HTTP-атрибуты | `POST`, `/api/issues`, `/api/issues/:id` |
| `error` | error | Ошибка (цепочка `%w`) | `create issue: ent: ...` |
| `subject` | string | NATS subject | `hub.issues.created` |

Соглашение: предпочитайте `duration_ms` (как в `Timed`/`Logger`); `dur_ms` допустим в ручных замерах, но лучше единообразно `duration_ms`.

### 12.2 Переменные окружения наблюдаемости

| Переменная | Назначение | Дефолт | Кто ставит |
|---|---|---|---|
| `OTEL_SERVICE_NAME` | Имя сервиса в логах/трейсах | `hub-backend` | devops |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP-коллектор; пусто = трейсинг выключен | пусто | devops |
| `ENVIRONMENT` | `production`/`staging`/`development` | `development` | devops |
| `DEV_MODE` | Небезопасные dev-эндпоинты; **запрещён в prod** | `false` | devops |
| `LOG_LEVEL` (рекомендация) | Уровень логов | `info` | devops |
| Metrics CIDR allowlist | Ограничение `/metrics` | пусто (открыто) | devops |

### 12.3 Шпаргалка для `be1`/`be2`

```go
import (
	"log/slog"
	"github.com/valerii/hub/backend/internal/platform/observability"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// 1. событие
slog.InfoContext(ctx, "issue created", "module", "tracker", "issue_id", id, "user_id", uid)

// 2. критичная операция (лог + метрика + span)
res, err := observability.Timed(ctx, "tracker.create_issue", func(ctx context.Context) (*Issue, error) {
	return repo.Create(ctx, req)
})

// 3. ошибка на границе (один раз)
if err != nil {
	slog.ErrorContext(ctx, "create issue failed", "module", "tracker", "error", err)
	return apperrors.NewInternal("failed to create issue")
}

// 4. ожидаемый исход -> типизированная ошибка, без Error-лога
if errors.Is(err, tracker.ErrNotFound) {
	return apperrors.NewNotFound("issue not found")
}
```

Запреты: `fmt.Println`, `log.Printf`, `log.Fatal`, `panic()`, секреты в логах, лог в цикле, log-and-return.

### 12.4 Шпаргалка для `fe1`/`fe2`

```tsx
// 1. обернуть модуль
<RouteErrorBoundary module="Трекер">
	<TrackerModule />
</RouteErrorBoundary>

// 2. dev-only лог
if (import.meta.env.DEV) console.debug("[tracker] state", state);

// 3. показать request_id в ошибке
toast.error("Ошибка", { description: err.requestId ? `Код: ${err.requestId}` : undefined });
```

Запреты: `console.log` в проде, токены/PII в консоли или URL.

### 12.5 Шпаргалка для `qa`

```go
// t.Log вместо fmt/slog
t.Logf("created id=%s in %s", id, elapsed)

// информативное падение
t.Fatalf("status = %d, want %d; body=%s", got, want, body)

// проверка метрики
observability.ResetMetrics()
_, _ = svc.Op(ctx)
snap := observability.Snapshot() // проверить Count/ErrorCount
```

Артефакты покрытия — в `${TMPDIR}`, не в `backend/`; не коммитить.

### 12.6 Шпаргалка для `devops`

```bash
# prod: JSON-логи (дефолт), уровень через env (рекомендация)
ENVIRONMENT=production
LOG_LEVEL=info

# опциональный трейсинг
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=hub-backend

# мониторинг-стек
docker compose -f deployments/docker-compose.monitoring.yml up -d
```

```promql
# ключевые алёрты
# 5xx error rate > 1%
(sum(rate(hub_http_requests_total{status=~"5.."}[5m])) / sum(rate(hub_http_requests_total[5m]))) > 0.01
# SLO p95 API > 200ms
histogram_quantile(0.95, sum(rate(hub_http_request_duration_milliseconds_bucket[5m])) by (le)) > 200
```

Обязательно: CIDR-allowlist на `/metrics`; ротация логов (`max-size`/`max-file` или Loki retention).

### 12.7 Карта источников (реальные файлы кода)

| Тема | Файл |
|---|---|
| Инициализация логгера | `backend/cmd/server/wiring_bootstrap.go` |
| Access-лог + stripQuery (ME-03) | `backend/internal/platform/middleware/logging.go` |
| RequestID middleware | `backend/internal/platform/middleware/requestid.go` |
| Централизованный ErrorHandler | `backend/internal/platform/middleware/middleware.go` |
| AppError (`{code,message,request_id}`) | `backend/internal/platform/errors/errors.go` |
| Timed/TimedVoid | `backend/internal/platform/observability/timing.go` |
| In-memory метрики / Snapshot | `backend/internal/platform/observability/metrics.go` |
| Prometheus метрики + /metrics + HTTP middleware | `backend/internal/platform/observability/prometheus.go` |
| /metrics CIDR allowlist | `backend/internal/platform/observability/metrics_allowlist.go` |
| OpenTelemetry (InitTracing, HTTPTraceMiddleware, spans) | `backend/internal/platform/observability/otel.go` |
| Тестовые хуки (hubtest) | `backend/internal/platform/observability/export_hooks.go` |
| Frontend ErrorBoundary | `frontend/src/components/ui/error-boundary.tsx` |
| Frontend RouteErrorBoundary | `frontend/src/components/route-error-boundary.tsx` |

### 12.8 Глоссарий

| Термин | Значение |
|---|---|
| Столпы наблюдаемости | Logs (события), Metrics (агрегаты), Traces (путь запроса) |
| Structured logging | Логи как пары ключ-значение (JSON), а не свободный текст |
| `request_id` | Сквозной идентификатор запроса; мост между логами, ответом и трейсом |
| Correlation | Связывание событий разных подсистем по общему id |
| Cardinality (кардинальность) | Число уникальных значений метки метрики; высокая — опасна для Prometheus |
| p50/p95/p99 | Перцентили латентности (медиана/почти худший/хвост) |
| SLO | Целевой уровень качества (например, p95 API < 200 мс) |
| Span | Узел трейса — одна операция с длительностью и атрибутами |
| Propagation | Проброс trace-контекста между сервисами (W3C `traceparent`) |
| Sampling | Логирование/трассировка части событий для экономии |
| OTLP | OpenTelemetry Protocol — формат экспорта телеметрии |
| Error boundary | React-компонент, перехватывающий ошибки рендера поддерева |

### 12.9 Быстрый принцип на одну строку

> **Логируй событие один раз на границе, структурированно, с `request_id` и `module`, без секретов; меряй критичное через `Timed`; метрики и логи обязательны, трейсинг — усиление; уровни не путай; в цикле не логируй.**

---

*Playbook основан на реальном коде Hub на 2026-07-02. При изменении кода наблюдаемости (`backend/internal/platform/observability/*`, `middleware/*`) — обновить соответствующие разделы (`.cursor/rules/55-docs-sync.mdc`).*
