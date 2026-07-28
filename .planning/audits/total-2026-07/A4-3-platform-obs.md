# W4 агент A3 — platform observability

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: metrics allowlist (PII), otel leakage, prometheus cardinality, timing.  
> Сверка с `grok-2026-07/14-observability.md`: A14-005 (CIDR) — **mitigated** в `config.validate` для `ENVIRONMENT=production` (вне списка A3); A14-016 (Timed Info) — **residual**; A14-009 (OTEL prod env) — wiring/deploy, вне списка.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/observability/consumer.go` | 40 | 0 | чисто |
| `backend/internal/platform/observability/export_hooks.go` | 25 | 0 | чисто |
| `backend/internal/platform/observability/metrics.go` | 88 | 1 | замечания |
| `backend/internal/platform/observability/metrics_allowlist.go` | 50 | 1 | замечания |
| `backend/internal/platform/observability/otel.go` | 201 | 3 | замечания |
| `backend/internal/platform/observability/prometheus.go` | 144 | 2 | замечания |
| `backend/internal/platform/observability/timing.go` | 37 | 1 | замечания |
| `backend/testframework/platform/observability/metrics_allow_middleware_test.go` | 49 | 1 | замечания |
| `backend/testframework/platform/observability/metrics_allowlist_test.go` | 56 | 0 | чисто |
| `backend/testframework/platform/observability/metrics_test.go` | 47 | 0 | чисто |
| `backend/testframework/platform/observability/otel_test.go` | 136 | 1 | замечания |
| `backend/testframework/platform/observability/prometheus_test.go` | 99 | 1 | замечания |
| `backend/testframework/platform/observability/timing_test.go` | 38 | 1 | замечания |
| `backend/testframework/platform/observability/trace_propagation_test.go` | 138 | 0 | чисто |

Файлов в списке: 14. Прочитано: 14. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W4-A3-001 | P2 | `otel.go:118-122` | Span attribute `http.target` = `c.Path()` (сырой URL path с UUID/id сегментами) | OTEL leakage: resource ID / потенциальный PII в каждом server span → Tempo/collector; высокая cardinality атрибутов. Канон — только `http.route` (шаблон) | Убрать `http.target` или писать только шаблон `r.Path`; query уже не пишется — не добавлять |
| W4-A3-002 | P2 | `prometheus.go:110-132` | Label `route`: при пустом `c.Route().Path` fallback на `c.Path()` | Prometheus cardinality bomb: 404 / unmatched / pre-route → уникальный label на каждый `/api/.../<uuid>`; scrape раздувается, Prom OOM risk | Всегда использовать route template; unmatched → `"unmatched"` / `"unknown"`; никогда сырой path |
| W4-A3-003 | P2 | `otel.go:146-147`; `otel.go:190-192` | `span.RecordError(err)` + `SetStatus(..., err.Error())` (и `finishOperationSpan`) | Текст ошибки (SQL, email, токены в wrapped err) уходит в OTEL backend — утечка internals/PII | Status message = короткий код (`"handler_error"`); `RecordError` с redacted attrs или без raw message |
| W4-A3-004 | P2 | `prometheus.go:71-86`; `metrics_allowlist.go:9-13` | Пустой CIDR-список = **fail-open** (`/metrics` всем). Prod enforce только в `config.validate` (вне пакета) | Любой caller/`ENVIRONMENT≠production` (staging) → scrape без ACL: latency/RPS/labels disclosure (residual A14-005 / A9-015 для non-prod). Пакет сам не знает ENVIRONMENT | Fail-closed по умолчанию **или** явный `AllowAllMetricsCIDRs` + тест; staging тоже требовать CIDR |
| W4-A3-005 | P3 | `otel.go:82-85` | Propagator = `TraceContext` **+** `Baggage` | Client-controlled baggage инжектится в контекст и уходит в exporter / downstream — baggage injection / PII в baggage | Только `TraceContext{}` пока нет доверенного baggage-контракта |
| W4-A3-006 | P3 | `timing.go:23-27` | Residual **A14-016**: каждый success `Timed` → `slog.Info` | Hot-path (`tracker.ListIssues`, `audit.RecordHTTPRequest`, …) раздувает Loki/диск | Success → `Debug`; Error оставить; уважать `LOG_LEVEL` |
| W4-A3-007 | P3 | `metrics_allowlist.go:9-28` vs `prometheus.go:71-88` | `MetricsAllowMiddleware` **не используется** в `wiring_fiber` (только `MetricsEndpoint`); дубль CIDR-логики | Drift двух allowlist-реализаций; мёртвый код + тесты на не-prod path | Удалить middleware **или** mount через него; одна функция `authorizeMetricsIP` |
| W4-A3-008 | P3 | `metrics.go:33-54`; `prometheus.go:62-68` | Имя `operation` без allowlist → CounterVec/HistogramVec label + in-memory `map` | Сегодня call sites статичны; ошибочный `Timed(ctx, userID, …)` → unbounded cardinality + рост map до restart | Allowlist/prefix (`tracker.`, `auth.`, …) или hash+cap; unknown → `"other"` |
| W4-A3-009 | P3 | `prometheus_test.go:42-68`; `otel_test.go:110-136` | Нет assert: HTTP metrics/`http.route` = **шаблон**, не raw path с UUID; нет assert отсутствия `http.target` UUID | Мутация «всегда `c.Path()`» / «добавить target» **не краснеет** | Запрос `/api/items/<uuid>` на route `/api/items/:id` → scrape `route="/api/items/:id"`; span attrs без raw UUID path |
| W4-A3-010 | P3 | `metrics_allow_middleware_test.go:24-36`; `timing_test.go:9-38` | Allow-middleware: нет deny non-matching IP; тест включает `ProxyHeader` (prod wiring **без** него). `Timed_*` без `ResetMetrics` | Слабое покрытие ACL; семантика IP ≠ prod; параллельные тесты могут загрязнять global metrics | Deny-тест `10.0.0.0/8` + peer вне сети; align ProxyHeader с wiring; `ResetMetrics` в Cleanup |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `consumer.go` | Nil-safe при tracing off; attrs `messaging.*` + optional `hub.event_id`; subject в span name — текущие NATS subjects событийные (не user UUID). End → `finishOperationSpan` |
| `export_hooks.go` | hubtest-only re-export record/CIDR/tracing flag — без prod-логики |
| `metrics.go` Snapshot/Reset | RWMutex; empty operation skip; Prometheus sync через `recordPrometheusOperation`; `ResetMetrics` чистит оба слоя |
| `metrics_allowlist.go` parseCIDRs | Битый CIDR silently skip; all-invalid → denyAll (fail-closed) — корректно при non-empty input |
| `otel.go` InitTracing | No endpoint → noop shutdown; service.name fallback; batcher + resource; `/metrics` path skip в HTTPTraceMiddleware |
| `otel.go` HTTPTraceMiddleware | Prefer `c.Route().Path` для span name/route attr (хорошо); request_id только из Locals |
| `prometheus.go` collectors | Labels method/route/status и operation/result — конечные при template routes; buckets разумные; scrape handler через fasthttpadaptor |
| `prometheus.go` MetricsEndpoint | CIDR check до scrape; invalid CIDRs → 403; `/metrics` не учитывается HTTPMetricsMiddleware |
| `timing.go` Timed/TimedVoid | Span + metrics + log; ошибка пробрасывается; generic OK |
| `metrics_allowlist_test.go` | Empty allow OK; invalid CIDR deny; non-match deny; `ipAllowed` table — asserts есть |
| `metrics_test.go` | Snapshot aggregation + Timed records — asserts есть |
| `trace_propagation_test.go` | NATS header/map round-trip; consumer parent linkage; tracing-disabled nil-safe — asserts есть |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **14 / 14 / 0** |
| Находок | **10** |
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 6 |

### Топ-3

1. **W4-A3-001** — `http.target` = raw `c.Path()` → UUID/PII в OTEL spans.  
2. **W4-A3-002** — Prometheus `route` fallback на raw path → cardinality explosion на 404/unmatched.  
3. **W4-A3-003** — `err.Error()` в span status/RecordError → утечка internals в collector.

### Вердикт по фокусу аудита

| Инвариант | Статус |
|-----------|--------|
| Metrics CIDR allowlist | **Prod fail-closed** через config (вне списка); пакет API **fail-open** при пустом CIDR (W4-A3-004); dead `MetricsAllowMiddleware` (W4-A3-007) |
| PII в metrics/traces | Operation names статичны (OK); **http.target** и error text — gaps (W4-A3-001/003) |
| OTEL leakage | Path IDs + err strings + optional Baggage (W4-A3-001/003/005) |
| Prometheus cardinality | Matched routes → template (OK); unmatched Path fallback — **дырка** (W4-A3-002); operation label без cap (W4-A3-008) |
| Timing | Корректный Timed/metrics/span; Info-шум residual A14-016 (W4-A3-006) |

P0/P1 в рамках 14 файлов **не найдены**. Критичный scrape ACL в production держится валидацией config; внутри пакета observability главные риски — OTEL path/error leakage и Prom cardinality на unmatched routes.
