# A2 W4 агент 3 — platform observability

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W4-ASSIGNMENTS.md` A3 · Stage A: `A4-3-platform-obs.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/observability/consumer.go` | 35 | 0 | чисто |
| `backend/internal/platform/observability/export_hooks.go` | 31 | 0 | чисто |
| `backend/internal/platform/observability/metrics.go` | 81 | 0 | чисто |
| `backend/internal/platform/observability/metrics_allowlist.go` | 51 | 0 | чисто |
| `backend/internal/platform/observability/otel.go` | 166 | 0 | чисто |
| `backend/internal/platform/observability/prometheus.go` | 119 | 0 | чисто |
| `backend/internal/platform/observability/timing.go` | 34 | 0 | чисто |
| `backend/testframework/platform/observability/metrics_allow_middleware_test.go` | — | 0 | **удалён** (периметр A: мёртвый middleware; замена: allowlist + route_deny) |
| `backend/testframework/platform/observability/metrics_allowlist_test.go` | 56 | 0 | чисто |
| `backend/testframework/platform/observability/metrics_test.go` | 41 | 0 | чисто |
| `backend/testframework/platform/observability/otel_test.go` | 111 | 0 | чисто |
| `backend/testframework/platform/observability/prometheus_test.go` | 82 | 0 | чисто |
| `backend/testframework/platform/observability/timing_test.go` | 35 | 0 | чисто |
| `backend/testframework/platform/observability/trace_propagation_test.go` | 114 | 0 | чисто |

Файлов в списке: 14. Прочитано: 13 + подтверждение удаления 1. Пропущено: 0.  
Доп. (вне списка, сверка фиксов): `route_label.go`, `operation_label.go`, `baggage_task065_test.go`, `pii_task043_test.go`, `metrics_route_deny_task059_test.go`.

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W4-3-001 | NEW | P3 | `W4-ASSIGNMENTS.md` A3 vs disk | Assignment всё ещё ссылает `metrics_allow_middleware_test.go`, файла нет | Дрейф периметра A2; агенты могут считать покрытие неполным | Обновить assignments / INDEX: файл removed с P3-BE-07 |

Активных дефектов продукта в 14 файлах **не найдено** (кроме дрейфа assignment).

## FIXED-OK

| Stage A ID | Sev | Статус | Доказательство |
|------------|-----|--------|----------------|
| W4-A3-001 | P2 | FIXED-OK | нет `http.target`; только `http.route` |
| W4-A3-002 | P2 | FIXED-OK | `httpRouteLabel` → `unmatched` |
| W4-A3-003 | P2 | FIXED-OK | `recordRedactedSpanError` |
| W4-A3-004 | P2 | FIXED-OK | empty CIDR → 403; sentinel `*` |
| W4-A3-005 | P3 | FIXED-OK | propagator без Baggage |
| W4-A3-006 | P3 | FIXED-OK | Timed success → Debug |
| W4-A3-007 | P3 | FIXED-OK | MetricsAllowMiddleware удалён |
| W4-A3-008 | P3 | FIXED-OK | `sanitizeOperationLabel` → `other` |
| W4-A3-009 | P3 | FIXED-OK | route/PII тесты расширены |
| W4-A3-010 | P3 | FIXED-OK | deny CIDR / ResetMetrics покрытие |

## Чисто

| Инвариант | Статус |
|-----------|--------|
| Metrics CIDR | fail-closed empty; `*` явный |
| OTEL path PII | template only |
| Prom cardinality unmatched | `unmatched` |
| Timing noise | Debug on success |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **14 / 14 / 0** |
| Активных | **1** (NEW 1 · assignment drift P3) |
| FIXED-OK | **10** |
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 |
