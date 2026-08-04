# Наблюдаемость: метрики и трассировка

> Часть плейбука. Канон и оглавление — [`../observability.md`](../observability.md).

Замер времени и объёмов, OpenTelemetry.

---

## 4. Замер времени и метрики (требование)

Наблюдаемость Hub **обязана** включать замер длительности критичных операций и экспорт метрик. Это не опция: без таймингов нельзя проверить SLO (§4.5) и нельзя настроить алёрты по latency (§9).

### 4.1 Базовый паттерн: замер длительности через defer

Минимальный, всегда доступный приём — засечь время в начале и залогировать в `defer`:

```go
func (s *Service) HeavyOp(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		// duration_ms логируется ВСЕГДА, даже при раннем return/ошибке
		slog.InfoContext(ctx, "heavy op done",
			"module", "tms",
			"entity_id", id,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}()

	// ... тяжёлая работа ...
	return nil
}
```

Почему `defer`, а не строка в конце: при раннем `return` (ошибка, ветвление) финальный лог с длительностью не выполнится, если он написан просто последней строкой. `defer` гарантирует замер на **любом** пути выхода.

Bad — замер, который теряется при ошибке:

```go
func (s *Service) HeavyOp(ctx context.Context) error {
	start := time.Now()
	if err := s.step(ctx); err != nil {
		return err // duration НЕ залогирован -> не видим латентность неудачных вызовов
	}
	slog.InfoContext(ctx, "done", "dur_ms", time.Since(start).Milliseconds())
	return nil
}
```

Важно: `dur_ms` для **неудачных** операций так же ценен, как для успешных — часто именно медленные таймауты предшествуют ошибке. Замеряйте оба исхода.

### 4.2 Предпочтительный путь: `observability.Timed`

Для критичных операций используйте готовую обёртку `observability.Timed` / `TimedVoid` (`backend/internal/platform/observability/timing.go`). Она делает три вещи разом: **лог длительности + метрика + span трейса**.

```10:37:backend/internal/platform/observability/timing.go
// Timed выполняет fn и логирует длительность в миллисекундах.
// Ошибки логируются на уровне Error, успех — Info.
func Timed[T any](ctx context.Context, operation string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := startOperationSpan(ctx, operation)
	start := time.Now()
	result, err := fn(ctx)
	ms := time.Since(start).Milliseconds()
	recordOperation(operation, ms, err != nil)
	finishOperationSpan(span, err)
	attrs := []any{
		"operation", operation,
		"duration_ms", ms,
	}
	if err != nil {
		slog.ErrorContext(ctx, "operation failed", append(attrs, "error", err)...)
		return result, err
	}
	slog.InfoContext(ctx, "operation completed", attrs...)
	return result, nil
}

// TimedVoid — вариант Timed для функций без возвращаемого значения.
func TimedVoid(ctx context.Context, operation string, fn func(context.Context) error) error {
	_, err := Timed(ctx, operation, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}
```

Что даёт `Timed` автоматически:

1. Открывает child-span OTel (если трейсинг включён) — `startOperationSpan`.
2. Замеряет длительность.
3. Пишет метрику через `recordOperation` (in-memory `Snapshot()` **и** Prometheus histogram/counter).
4. Закрывает span, помечая ошибку при провале.
5. Логирует `operation completed` (Info) или `operation failed` (Error) с `duration_ms`.

Good — критичная операция обёрнута в `Timed`:

```go
func (s *Service) GenerateReport(ctx context.Context, projectID uuid.UUID) (*Report, error) {
	// имя operation стабильно и низкокардинально -> хорошая метка метрики/span
	return observability.Timed(ctx, "tms.generate_report",
		func(ctx context.Context) (*Report, error) {
			return s.buildReport(ctx, projectID)
		},
	)
}
```

Good — `TimedVoid` для операции без результата:

```go
func (w *Worker) flushOutbox(ctx context.Context) error {
	return observability.TimedVoid(ctx, "outbox.flush", func(ctx context.Context) error {
		return w.publishPending(ctx)
	})
}
```

Bad — ручной дубляж того, что уже делает `Timed` (лог + метрика + span вручную, с риском рассинхрона):

```go
func (s *Service) GenerateReport(ctx context.Context) (*Report, error) {
	start := time.Now()
	r, err := s.buildReport(ctx)
	// забыли метрику и span; продублировали то, что даёт Timed
	slog.InfoContext(ctx, "report", "dur_ms", time.Since(start).Milliseconds())
	return r, err
}
```

### 4.3 Именование операций (кардинальность меток)

`operation` становится **меткой** Prometheus (`operation="tms.generate_report"`). Метки с высокой кардинальностью взрывают память Prometheus и делают метрики бесполезными.

Правила именования:

- Формат: `<module>.<action>` в snake_case — `tracker.create_issue`, `messenger.send_message`, `auth.issue_magic_link`.
- **Никаких переменных данных в имени**: не включайте `user_id`, `issue_id`, timestamp. Это создаёт бесконечное число меток.
- Стабильность: одно и то же имя для одной и той же логической операции во всех вызовах.

Good:

```go
observability.Timed(ctx, "messenger.send_message", fn) // конечное множество имён
```

Bad — кардинальность метки бесконечна:

```go
observability.Timed(ctx, "send_message_to_"+channelID.String(), fn) // взрыв меток Prometheus
```

### 4.4 Метрики: что уже собирается

В Hub две системы метрик, обе питаются из `recordOperation` (`observability/metrics.go`):

**A. In-memory агрегат (`Snapshot()`)** — быстрый обзор без внешних систем:

```9:18:backend/internal/platform/observability/metrics.go
// OperationMetric — агрегированная статистика по имени операции (OBS-01).
type OperationMetric struct {
	Operation   string  `json:"operation"`
	Count       int64   `json:"count"`
	ErrorCount  int64   `json:"error_count"`
	TotalMs     int64   `json:"total_ms"`
	MaxMs       int64   `json:"max_ms"`
	AvgMs       float64 `json:"avg_ms"`
	LastSeenUTC string  `json:"last_seen_utc,omitempty"`
}
```

`Snapshot()` возвращает срез по всем операциям (count, error_count, avg, max). Полезно для отладочного эндпоинта/health-дашборда без Prometheus.

**B. Prometheus** (`observability/prometheus.go`) — для дашбордов и алёртов. Зарегистрированные метрики:

| Метрика | Тип | Метки | Смысл |
|---|---|---|---|
| `hub_operation_calls_total` | Counter | `operation`, `result` (success/error) | Число вызовов `Timed()` |
| `hub_operation_duration_milliseconds` | Histogram | `operation` | Длительность операций (buckets 1..5000мс) |
| `hub_http_requests_total` | Counter | `method`, `route`, `status` | Число HTTP-запросов |
| `hub_http_request_duration_milliseconds` | Histogram | `method`, `route` | Латентность HTTP (DefBuckets) |

HTTP-метрики снимаются в `HTTPMetricsMiddleware` (по `c.Route().Path`, не по конкретному пути — снова забота о кардинальности: `/api/issues/:id`, а не `/api/issues/123`).

Счётчики (Counter) — для «сколько раз»: запросы, ошибки, отправленные сообщения. Гистограммы (Histogram) — для распределения длительностей, из которых Prometheus считает p50/p95/p99 (§4.6).

### 4.5 Эндпоинт `/metrics` и его защита

Prometheus-scrape доступен на `GET /metrics` (`MetricsHandler`). Он защищён CIDR-allowlist (`MetricsEndpoint` / `MetricsAllowMiddleware`):

```71:89:backend/internal/platform/observability/prometheus.go
// MetricsEndpoint — /metrics с опциональным CIDR allowlist (OBS-01).
func MetricsEndpoint(allowedCIDRs []string) fiber.Handler {
	scrape := MetricsHandler()
	nets := parseCIDRs(allowedCIDRs)
	denyAll := len(allowedCIDRs) > 0 && len(nets) == 0

	return func(c fiber.Ctx) error {
		if len(allowedCIDRs) > 0 {
			if denyAll {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
			ip := net.ParseIP(c.IP())
			if ip == nil || !ipAllowed(ip, nets) {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
		}
		return scrape(c)
	}
}
```

Правила для `devops`:

- В проде **обязательно** задать allowlist (CIDR сети Prometheus/мониторинга). Пустой список = открыто всем — допустимо только в изолированной dev-сети.
- `/metrics` исключён из access-лога и трейс/метрик middleware (`if c.Path() == metricsPath { return c.Next() }`), чтобы scrape не засорял собственные метрики. При добавлении новых middleware учитывайте это исключение.

### 4.6 p50 / p95 / p99 и целевые SLO

Средняя длительность (`avg`) **лжёт**: один медленный запрос из тысячи не поднимет среднее, но испортит опыт этого пользователя. Поэтому мерим **перцентили**.

- **p50** (медиана) — типичный пользователь.
- **p95** — «почти худший» опыт; основной SLO-показатель Hub.
- **p99** — хвост; ловит редкие деградации, таймауты, GC-паузы.

Из гистограмм Prometheus перцентили считаются через `histogram_quantile`:

```promql
# p95 латентности HTTP по маршруту за 5 минут
histogram_quantile(0.95,
  sum(rate(hub_http_request_duration_milliseconds_bucket[5m])) by (le, route)
)

# p99 длительности бизнес-операции
histogram_quantile(0.99,
  sum(rate(hub_operation_duration_milliseconds_bucket[5m])) by (le, operation)
)
```

**Целевые SLO Hub** (источник — `docs/TESTING.md`, раздел производительности):

| Показатель | Цель | Метрика для проверки |
|---|---|---|
| API response time (p95) | **< 200 мс** | `hub_http_request_duration_milliseconds` |
| WebSocket message latency | **< 100 мс** | замер доставки в WS-слое |
| Kanban board load (100 задач) | **< 500 мс** | `tracker.*` операции + фронт |
| Messenger: загрузка канала (1000 сообщений) | **< 300 мс** | `messenger.*` операции + фронт |

Правило: если операция входит в один из SLO выше — она **обязана** быть обёрнута в `Timed` (или иметь HTTP-метрику через middleware), иначе SLO нечем проверять.

### 4.7 Где мерить

| Место | Как | Что покрывает |
|---|---|---|
| HTTP-граница | `HTTPMetricsMiddleware` (авто) | Латентность/счётчики всех запросов по маршруту |
| Критичные бизнес-операции | `observability.Timed(ctx, "module.action", fn)` | Отчёты, импорт/экспорт, генерация, тяжёлые выборки |
| Фоновые задачи | `TimedVoid` в теле job | outbox flush, cron-задачи, консьюмеры |
| Внешние вызовы | `Timed` вокруг вызова интеграции | GitHub/GitLab/Telegram/LiveKit/SMTP латентность |
| WS-доставка | ручной `start/defer` с `dur_ms` | latency доставки сообщения (SLO < 100 мс) |

Не мерьте **всё подряд**: тривиальные getter'ы и мгновенные операции не нуждаются в `Timed` — это лишний шум логов Info. Мерьте то, что: (а) входит в SLO, (б) обращается к БД/сети, (в) исторически бывало медленным.

### 4.8 Запись таймингов прогонов тестов

Тайминги — это не только рантайм, но и тесты. По `docs/TESTING.md` и `.cursor/rules/56-regression-verify.mdc` при регрессии фиксируйте длительность прогонов, чтобы отслеживать деградацию CI и находить медленные тесты.

Практика для `qa`:

- Прогон бэкенда: `go test -race -count=1 -timeout=600s ./...` — Go сам печатает время по пакетам; логируйте суммарное время в `.planning/COVERAGE.md`/отчёт сессии при заметном росте.
- Для точечного профилирования медленного теста используйте `-v` и `t.Log` с временем этапов (§7).
- Целевые пороги производительности из `docs/TESTING.md` (§4.6) проверяются перф-тестами; их результаты — часть DoD (§10).

Артефакты покрытия/профилей **не коммитим** (`.cursor/rules/45-artifacts-cleanup.mdc`): пишите профили в системный temp, не в `backend/`.

---

<a name="5-трейсинг-opentelemetry"></a>
## 5. Трейсинг (OpenTelemetry)

### 5.1 Что уже есть в коде

Hub **уже** имеет OpenTelemetry-трейсинг (OTel v1.40.0 по `docs/stack_and_tools.md`), реализованный в `backend/internal/platform/observability/otel.go`. Он **опциональный**: включается только при заданном OTLP-эндпоинте.

Инициализация (`InitTracing`) регистрирует global `TracerProvider` **только** если задан `OTEL_EXPORTER_OTLP_ENDPOINT` (или `cfg.OTLPEndpoint`):

```42:49:backend/internal/platform/observability/otel.go
func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
```

Ключевые факты:

- Экспорт по **OTLP HTTP** (`otlptracehttp`). Эндпоинт — коллектор (Grafana Tempo/Jaeger/OTel Collector).
- Пропагация — `TraceContext` + `Baggage` (W3C). Это значит, что `traceparent`-заголовок принимается от клиента и пробрасывается дальше.
- Флаг `tracingActive` (atomic) хранит, включён ли реальный провайдер. `TracingEnabled()` возвращает его. Если трейсинг выключен, span'ы **не создаются** (нулевой оверхед).
- Имя сервиса: `cfg.ServiceName` → `OTEL_SERVICE_NAME` → `hub-backend`.
- Возвращается функция `Shutdown`, которую bootstrap регистрирует в cleanup с таймаутом 5 сек (graceful flush трейсов при остановке).

### 5.2 Серверные span'ы: `HTTPTraceMiddleware`

Каждый HTTP-запрос оборачивается в server-span:

```97:127:backend/internal/platform/observability/otel.go
// HTTPTraceMiddleware — server span на запрос и propagation context.
func HTTPTraceMiddleware() fiber.Handler {
	tracer := otel.Tracer(httpTracerName)
	prop := otel.GetTextMapPropagator()

	return func(c fiber.Ctx) error {
		if c.Path() == metricsPath {
			return c.Next()
		}

		carrier := fiberHeaderCarrier{c: c}
		ctx := prop.Extract(c.UserContext(), carrier)

		route := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			route = r.Path
		}
		spanName := c.Method() + " " + route

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
			attribute.String("http.target", c.Path()),
		)
		if requestID, ok := c.Locals("requestID").(string); ok && requestID != "" {
			span.SetAttributes(attribute.String("hub.request_id", requestID))
		}

		c.SetUserContext(ctx)
```

Важные детали:

- Span называется `METHOD /route` по **шаблону** маршрута (низкая кардинальность), не по конкретному пути.
- В span кладётся `hub.request_id` — так трейс склеивается с логами и ответом пользователю. Это мост между тремя столпами.
- `prop.Extract` вытаскивает входящий trace-контекст: если запрос пришёл от другого сервиса/фронта с `traceparent`, span станет частью распределённого трейса.
- `c.SetUserContext(ctx)` — важно: дальше по стеку `c.UserContext()` уже несёт span. Поэтому `Timed` (§4.2), вызванный внутри handler, создаст **child**-span автоматически.

### 5.3 Child-span'ы бизнес-операций

`Timed`/`TimedVoid` создают дочерние span'ы через `startOperationSpan`, но **только** если трейсинг включён:

```178:195:backend/internal/platform/observability/otel.go
// startOperationSpan — child span при активном tracing.
func startOperationSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	if !TracingEnabled() {
		return ctx, nil
	}
	return otel.Tracer(operationTracerName).Start(ctx, operation)
}

func finishOperationSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
```

Следствие: **обёртка `Timed` — это и метрика, и лог, и span одновременно.** Обернув критичную операцию в `Timed`, вы бесплатно получаете узел в трейсе. Отдельно вызывать OTel API почти никогда не нужно.

Good — вложенность даёт полезный трейс:

```go
// HTTPTraceMiddleware создал server-span для запроса;
// Timed добавит child-span "tracker.create_issue" внутри него
func (s *Service) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	return observability.Timed(ctx, "tracker.create_issue",
		func(ctx context.Context) (*Issue, error) {
			return s.repo.Create(ctx, req) // выполняется внутри child-span
		},
	)
}
```

### 5.4 Пропагация в исходящих вызовах

Для входящих запросов пропагация уже настроена. Для **исходящих** (Hub → GitHub/GitLab/внешний сервис) при необходимости распределённого трейса нужно инжектить контекст в заголовки запроса:

```go
// пробрасываем trace-контекст во внешний вызов, чтобы трейс не рвался
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
resp, err := client.Do(req)
```

Напоминание безопасности: для user-controlled URL используйте `safehttp.NewSafeHTTPClient` и `ValidateOutboundURL` (SSRF), а не `http.DefaultClient` — это ортогонально трейсингу, но обязательно.

**NATS / outbox (есть в коде, 2026-07-03).** Трейс не рвётся на асинхронной границе REST → outbox → NATS → consumer:

- `Publisher.Publish/PublishInTx` (`platform/outbox/publisher.go`) сохраняют W3C trace-context из ctx в колонку `trace_context` outbox-события; worker при publish переносит его в NATS headers (`traceparent` + `Hub-Event-Id`).
- Хелперы: `natsclient.InjectTraceContext/ExtractTraceContext` (carrier поверх `nats.Header`), `TraceContextMap/ExtractTraceContextFromMap` (`platform/nats/trace.go`). Пропагатор явный (W3C TraceContext+Baggage), не глобальный — работает и без OTLP endpoint.
- Прямой publish в рамках запроса: `nc.PublishJSONContext(ctx, subject, payload)` или `PublishWithTraceContext` вместо `PublishJSON` без ctx.
- Consumer-сторона: `ctx = natsclient.ExtractTraceContext(ctx, msg.Headers())`, затем `observability.StartConsumerSpan(ctx, subject, eventID)` / `EndConsumerSpan(span, err)` — span `consume <subject>` (kind=Consumer, атрибуты `messaging.*`, `hub.event_id`). Применено в notification consumer, audit consumer, ws bridge (`traced`-обёртка).

### 5.5 Если трейсинг выключен

Когда `OTEL_EXPORTER_OTLP_ENDPOINT` не задан (типичный dev/минимальный prod):

- `InitTracing` возвращает no-op shutdown, `tracingActive=false`.
- `startOperationSpan` возвращает `(ctx, nil)` — span'ы не создаются, оверхед нулевой.
- Логи и метрики продолжают работать полностью.

Это значит: **логи и метрики — базовый уровень (всегда), трейсинг — усиление (по возможности)**. Не полагайтесь на трейсы там, где обязаны быть лог и метрика.

### 5.6 Минимальное включение (для devops)

Чтобы поднять трейсинг, достаточно окружения — код уже готов:

```bash
# указать OTLP-коллектор (например, Grafana Tempo / OTel Collector)
OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector:4318"
OTEL_SERVICE_NAME="hub-backend"
```

Коллектор поднимается опциональным `deployments/docker-compose.monitoring.yml` (Prometheus + Loki + Grafana, pinned v3.x — `docs/stack_and_tools.md`). Sampling трейсов настраивается на уровне SDK/коллектора: в проде под нагрузкой не экспортируйте 100% трейсов — используйте head/tail sampling на коллекторе, чтобы не платить за хранение.

### 5.7 Тестовый сброс

Для изоляции тестов есть `ResetTracingForTest()` и (под `//go:build hubtest`) `ExportSetTracingActiveForTest`. Используйте их в тестах наблюдаемости, чтобы включать/выключать трейсинг детерминированно и не оставлять global state между тестами.

---

<a name="6-frontend-наблюдаемость"></a>
