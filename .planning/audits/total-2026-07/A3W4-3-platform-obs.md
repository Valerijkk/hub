# A3 W4 агент 3 — platform observability (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W4-3-platform-obs.md](A2W4-3-platform-obs.md) · Baseline Stage A: [A4-3-platform-obs.md](A4-3-platform-obs.md)  
> Периметр: **14** файлов (7 prod + 7 tests; `metrics_allow_middleware_test` **removed**) · Прочитано: **14/14** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignment: [W4-ASSIGNMENTS.md](W4-ASSIGNMENTS.md) § A3  
> Фокус: metrics allowlist/deny · OTel export · Prometheus scrape · PII labels · consumer cancel

> **Вне списка assignment (сверка FIXED-OK / sanitize):** `route_label.go`, `operation_label.go`,
> `pii_task043_test.go`, `operation_label_task065_test.go`, `baggage_task065_test.go` —
> беглый просмотр; сильные PII/unmatched regression живут там. Wiring CIDR —
> `wiring_fiber.go` + `config.validate` (вне слота).

## Сверка с A2 (A2W4-3-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W4-3-001 | FIXED-OK | закрыт | `W4-ASSIGNMENTS.md` A3: deleted `metrics_allow_middleware_test.go` убран; добавлен `metrics_route_deny_task059_test.go` (TASK-112 / P3-BE-07) |

**REGRESS:** 0

### Cascade FIXED-OK (Stage A W4-A3-* / A2 FIXED-OK)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W4-A3-001 | нет `http.target`; только `http.route` — `otel.go:115–118`, `141–144`; `pii_task043_test.go` (вне слота) |
| W4-A3-002 | `httpRouteLabel` → `unmatched` — `prometheus.go:104`; `route_label.go:12–21` (вне); `metrics_route_deny_task059_test.go:22–46` |
| W4-A3-003 | `recordRedactedSpanError` — `otel.go:139–140`, `182–183`, `188–192`; `pii_task043_test.go` (вне) |
| W4-A3-004 | empty CIDR → 403; sentinel `*` — `metrics_allowlist.go:16–27`, `MetricsAllowAllCIDRs`; `metrics_allowlist_test.go:13–22`, `25–39` |
| W4-A3-005 | propagator только `TraceContext{}` — `otel.go:82–83` |
| W4-A3-006 | Timed success → Debug — `timing.go:27` |
| W4-A3-007 | MetricsAllowMiddleware удалён; CIDR в `authorizeMetricsIP` — `metrics_allowlist.go:20–37`; `prometheus.go:74–81` |
| W4-A3-008 | `sanitizeOperationLabel` → `other` — `metrics.go:34–37`; `operation_label.go` (вне); см. RESIDUAL A3W4-3-002 |
| W4-A3-009 | route/PII тесты — `metrics_route_deny_task059_test.go` + `pii_task043` (вне) |
| W4-A3-010 | deny CIDR / ResetMetrics — `metrics_allowlist_test.go:50–68`; `metrics_route_deny_task059_test.go:51–66`; `ResetMetrics` в `metrics.go:84–89` |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/platform/observability/consumer.go` | 41 | 1 | FIXED-OK nil-safe; NEW cancel→Error via End |
| `backend/internal/platform/observability/export_hooks.go` | 41 | 0 | чисто — hubtest re-export |
| `backend/internal/platform/observability/metrics.go` | 90 | 1 | sanitize call-site; RWMutex OK; RESIDUAL UUID suffix |
| `backend/internal/platform/observability/metrics_allowlist.go` | 60 | 0 | FIXED-OK fail-closed empty / `*` / invalid→403 |
| `backend/internal/platform/observability/otel.go` | 199 | 1 | FIXED-OK route/redact/baggage; NEW cancel status |
| `backend/internal/platform/observability/prometheus.go` | 136 | 0 | FIXED-OK CIDR endpoint + unmatched route label |
| `backend/internal/platform/observability/timing.go` | 38 | 0 | FIXED-OK Debug success; err log + metrics path OK |
| `backend/testframework/platform/observability/metrics_allowlist_test.go` | 69 | 0 | чисто — empty/`*`/invalid/non-match |
| `backend/testframework/platform/observability/metrics_route_deny_task059_test.go` | 67 | 0 | FIXED-OK template≠UUID + proxy peer deny |
| `backend/testframework/platform/observability/metrics_test.go` | 48 | 0 | чисто — Snapshot / Timed records |
| `backend/testframework/platform/observability/otel_test.go` | 137 | 0 | слабо redact assert (сильнее `pii_task043` вне слота) |
| `backend/testframework/platform/observability/prometheus_test.go` | 100 | 0 | чисто — scrape + HTTP middleware |
| `backend/testframework/platform/observability/timing_test.go` | 39 | 0 | чисто — success/error/void |
| `backend/testframework/platform/observability/trace_propagation_test.go` | 139 | 0 | чисто — NATS round-trip + consumer parent |

Файлов в списке: **14**. Прочитано: **14**. Пропущено: **0**.  
Удаление `metrics_allow_middleware_test.go`: подтверждено (`Test-Path` = False).

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W4-3-001 | NEW | P3 | `otel.go:178–186` (call: `consumer.go:38–40`; также `timing.go:17–18`) | `finishOperationSpan` / `EndConsumerSpan` помечают **любой** `err != nil` как Error (`operation_error`), включая `context.Canceled` / `DeadlineExceeded` | Graceful shutdown / cancel consumer → span Error + (через Timed) `error_count`/slog.Error; шумит error-rate в Tempo/Prom при штатной отмене | `errors.Is(err, context.Canceled\|DeadlineExceeded)` → End без Error-status (Unset/Ok); metrics `failed=false` для cancel; regression: cancel Timed/consumer → status≠Error |
| A3W4-3-002 | RESIDUAL (хвост W4-A3-008) | P3 | `metrics.go:34–37` + sanitize `operation_label.go:44–51` (вне списка, пакет) | `sanitizeOperationLabel` пропускает `tracker.<uuid>`: domain в allowlist, rest = hex+`-` | W4-A3-008 закрыл bare-UUID/`user-controlled`, но `Timed(ctx, "tracker."+id, …)` → unbounded Prom/`map` cardinality; `operation_label_task065` не мутирует этот кейс | Запретить UUID-shaped сегменты / требовать `[A-Za-z][A-Za-z0-9_]*` без длинных hyphen-runs; тест `tracker.550e8400-…` → `other`; mutation: ослабить charset → FAIL |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W4-3-001 | assignment A3 без middleware-test; + `metrics_route_deny_task059_test.go` |
| W4-A3-001…007, 009…010 | см. cascade — без регрессии на HEAD |
| W4-A3-008 | частичный FIXED-OK (domain allowlist); хвост → A3W4-3-002 |

## Чисто

- **Metrics CIDR:** empty / all-invalid / non-match → 403; `*` explicit allow-all; `authorizeMetricsIP` единый path для `MetricsEndpoint`.
- **OTEL path PII:** template `http.route` only; no `http.target`; `/metrics` skip в HTTPTrace + HTTPMetrics.
- **Error redact (spans):** `handler_error` / `operation_error` без raw `err.Error()` (сильные asserts вне слота).
- **Baggage:** propagator без Baggage.
- **Prom cardinality (routes):** unmatched → `unmatched`; matched → Fiber template (`:id`).
- **Timing noise:** success Debug; error остаётся Error (с сырым err в Loki — осознанный ops-канал, не span).
- **Consumer span:** messaging attrs + optional `hub.event_id`; nil-safe End; parent из NATS extract (trace_propagation).
- **OTLP:** no endpoint → noop; batcher + `service.name`; endpoint логируется при enable (без credentials в канон-URL).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total** | **2** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 1 (`A2W4-3-001`) |
| FIXED-OK (Stage A cascade) | 10 |
| RESIDUAL | 1 (`A3W4-3-002`) |
| NEW | 1 (`A3W4-3-001`) |
| REGRESS | 0 |

**Топ IDs:** `A3W4-3-001` (consumer/Timed cancel→Error) · `A3W4-3-002` (sanitize `domain.<uuid>` cardinality).
