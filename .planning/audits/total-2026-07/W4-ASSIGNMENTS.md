# W4 Assignments — platform + apigateway

> Стадия A · read-only · TASK-008

## A1 — platform http+middleware (15)
Отчёт: `A4-1-platform-http.md`

- `backend/internal/platform/http/export_hooks.go`
- `backend/internal/platform/http/helpers.go`
- `backend/internal/platform/http/opengraph.go`
- `backend/internal/platform/http/safedialer.go`
- `backend/internal/platform/middleware/cors.go`
- `backend/internal/platform/middleware/export_hooks.go`
- `backend/internal/platform/middleware/logging.go`
- `backend/internal/platform/middleware/middleware.go`
- `backend/internal/platform/middleware/ratelimit.go`
- `backend/internal/platform/middleware/requestid.go`
- `backend/internal/platform/middleware/security_headers.go`
- `backend/testframework/platform/http/helpers_test.go`
- `backend/testframework/platform/http/opengraph_test.go`
- `backend/testframework/platform/http/safedialer_test.go`
- `backend/testframework/platform/middleware/middleware_test.go`

## A2 — platform nats+outbox+redis (15)
Отчёт: `A4-2-platform-nats-outbox.md`

- `backend/internal/platform/nats/client.go`
- `backend/internal/platform/nats/export_hooks.go`
- `backend/internal/platform/nats/trace.go`
- `backend/internal/platform/outbox/export_hooks.go`
- `backend/internal/platform/outbox/publisher.go`
- `backend/internal/platform/outbox/worker.go`
- `backend/internal/platform/redis/client.go`
- `backend/testframework/platform/nats/client_integration_test.go`
- `backend/testframework/platform/nats/client_publish_test.go`
- `backend/testframework/platform/nats/client_test.go`
- `backend/testframework/platform/nats/redact_url_test.go`
- `backend/testframework/platform/outbox/outbox_test.go`
- `backend/testframework/platform/outbox/publisher_test.go`
- `backend/testframework/platform/outbox/trace_context_test.go`
- `backend/testframework/platform/outbox/worker_wave101_test.go`

## A3 — platform observability (14 → metrics_allow_middleware_test **removed** P3-BE-07 / A2W4-3-001)
Отчёт: `A4-3-platform-obs.md`

- `backend/internal/platform/observability/consumer.go`
- `backend/internal/platform/observability/export_hooks.go`
- `backend/internal/platform/observability/metrics.go`
- `backend/internal/platform/observability/metrics_allowlist.go`
- `backend/internal/platform/observability/otel.go`
- `backend/internal/platform/observability/prometheus.go`
- `backend/internal/platform/observability/timing.go`
- `backend/testframework/platform/observability/metrics_allowlist_test.go`
- `backend/testframework/platform/observability/metrics_route_deny_task059_test.go`
- `backend/testframework/platform/observability/metrics_test.go`
- `backend/testframework/platform/observability/otel_test.go`
- `backend/testframework/platform/observability/prometheus_test.go`
- `backend/testframework/platform/observability/timing_test.go`
- `backend/testframework/platform/observability/trace_propagation_test.go`

## A4 — platform config/db/crypto/... (20)
Отчёт: `A4-4-platform-core.md`

- `backend/internal/platform/config/config.go`
- `backend/internal/platform/config/export_hooks.go`
- `backend/internal/platform/cron/cron.go`
- `backend/internal/platform/crypto/aesgcm.go`
- `backend/internal/platform/database/mongodb.go`
- `backend/internal/platform/database/postgres.go`
- `backend/internal/platform/errors/errors.go`
- `backend/internal/platform/errors/messages.go`
- `backend/internal/platform/health/probes.go`
- `backend/internal/platform/lifecycle/listen.go`
- `backend/internal/platform/telegram/markdown.go`
- `backend/testframework/platform/config/config_test.go`
- `backend/testframework/platform/cron/cron_test.go`
- `backend/testframework/platform/crypto/aesgcm_test.go`
- `backend/testframework/platform/database/mongo_disconnect_a8007_test.go`
- `backend/testframework/platform/errors/errors_test.go`
- `backend/testframework/platform/errors/errors_wave101_test.go`
- `backend/testframework/platform/health/probes_test.go`
- `backend/testframework/platform/lifecycle/listen_a8008_test.go`
- `backend/testframework/platform/telegram/markdown_test.go`

## A5 — apigateway (19)
Отчёт: `A4-5-apigateway.md`

- `backend/internal/apigateway/apikey.go`
- `backend/internal/apigateway/dto.go`
- `backend/internal/apigateway/errors.go`
- `backend/internal/apigateway/export_hooks.go`
- `backend/internal/apigateway/handler.go`
- `backend/internal/apigateway/repository.go`
- `backend/internal/apigateway/webhook.go`
- `backend/testframework/apigateway/apikey_auth_test.go`
- `backend/testframework/apigateway/apikey_test.go`
- `backend/testframework/apigateway/handler_extended_test.go`
- `backend/testframework/apigateway/handler_test.go`
- `backend/testframework/apigateway/handler_webhook_test.go`
- `backend/testframework/apigateway/repository_extra_test.go`
- `backend/testframework/apigateway/repository_test.go`
- `backend/testframework/apigateway/webhook_dispatcher_apigw02_test.go`
- `backend/testframework/apigateway/webhook_dispatcher_apigw03_test.go`
- `backend/testframework/apigateway/webhook_dispatcher_wave97_test.go`
- `backend/testframework/apigateway/webhook_signing_test.go`
- `backend/testframework/apigateway/webhook_test.go`

**Σ файлов: 83**

