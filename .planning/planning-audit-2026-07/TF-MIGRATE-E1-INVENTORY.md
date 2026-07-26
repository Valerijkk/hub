# TF-MIGRATE E1 — Inventory: `backend/testframework` → `backend/internal`

**Дата замера:** 2026-07-27  
**Источник решения:** [`02-TEST-FRAMEWORK-MIGRATION.md`](02-TEST-FRAMEWORK-MIGRATION.md) §Э3  
**Статус:** inventory only — **перенос файлов не начат**  
**Модель:** Cursor Grok 4.5 (cursor-grok-4.5-high)

---

## HARD CONSTRAINT (gate)

> **Этап Э1 миграции** (каркас Python `/tests`, см. `02-TEST-FRAMEWORK-MIGRATION.md`) **не начинать**, пока не существует **A1 coverage baseline** (gate **2026-07-27**).
>
> Без зафиксированного coverage baseline любое перемещение 515 `*_test.go` делает потерю покрытия необнаружимой.
>
> В [`01-PLANNING-FIXES.md`](01-PLANNING-FIXES.md) перезамер coverage описан как **A3** («перестать выдавать протухшее покрытие за текущее»). Для этого workstream gate именуется **A1 coverage baseline** — тот же смысл: честный замер на текущем HEAD **до** Э1.

Связанный риск: `02-TEST-FRAMEWORK-MIGRATION.md` §Риски — «Потеря покрытия при переезде».

---

## Замер NOW

| Метрика | Значение |
|---|---|
| Файлов `backend/testframework/**/*_test.go` | **515** |
| Дата/способ | 2026-07-27, `Get-ChildItem -Recurse -Filter '*_test.go'` / Python `rglob` |
| Расхождение с `02-TEST-FRAMEWORK-MIGRATION.md` | там указано **508**; актуальный замер — **515** (+7 с момента черновика) |
| Стиль пакетов | **515/515** — `package <domain>_test` (black-box / external test) |
| Пакетов-доменов с тестами | **42** (включая `platform/*`, `integrations/onec`, `env`) |

> AGENTS.md baseline «515 files · measured 2026-07-27» совпадает с этим замером.

---

## Правило маппинга target path

| Current | Target |
|---|---|
| `backend/testframework/<domain>/<file>_test.go` | `backend/internal/<domain>/<file>_test.go` |
| `backend/testframework/platform/<sub>/<file>_test.go` | `backend/internal/platform/<sub>/<file>_test.go` |
| `backend/testframework/onec/<file>_test.go` | `backend/internal/integrations/onec/<file>_test.go` |
| `backend/testframework/env/builder_test.go` | **исключение** — это helper testframework, не прод-домен (см. ниже) |

Не-тестовые пакеты `testframework/{factory,mongotest,steps,env}` в эту таблицу **не входят** (не `*_test.go`), кроме `env/builder_test.go`. Их судьба — Э5/фикстуры Python или `backend/internal/testutil/`.

---

## Сводка по доменам

| Domain | Files | Target root |
|---|---:|---|
| accounting | 8 | `backend/internal/accounting/` |
| ai | 30 | `backend/internal/ai/` |
| apigateway | 12 | `backend/internal/apigateway/` |
| audit | 19 | `backend/internal/audit/` |
| auth | 28 | `backend/internal/auth/` |
| backup | 8 | `backend/internal/backup/` |
| bases | 10 | `backend/internal/bases/` |
| calendar | 2 | `backend/internal/calendar/` |
| calls | 21 | `backend/internal/calls/` |
| crm | 10 | `backend/internal/crm/` |
| dashboard | 6 | `backend/internal/dashboard/` |
| dev | 3 | `backend/internal/dev/` |
| env (test helper — not production domain) | 1 | `backend/internal/testutil/env/` *или* shared helper до Python `/tests` |
| github | 17 | `backend/internal/github/` |
| gitlab | 9 | `backend/internal/gitlab/` |
| hr | 7 | `backend/internal/hr/` |
| imex | 8 | `backend/internal/imex/` |
| integrations/onec | 4 | `backend/internal/integrations/onec/` |
| intranet | 8 | `backend/internal/intranet/` |
| mail | 5 | `backend/internal/mail/` |
| marketplace | 7 | `backend/internal/marketplace/` |
| messenger | 41 | `backend/internal/messenger/` |
| notification | 36 | `backend/internal/notification/` |
| platform/config | 1 | `backend/internal/platform/config/` |
| platform/cron | 1 | `backend/internal/platform/cron/` |
| platform/crypto | 1 | `backend/internal/platform/crypto/` |
| platform/errors | 2 | `backend/internal/platform/errors/` |
| platform/http | 3 | `backend/internal/platform/http/` |
| platform/middleware | 1 | `backend/internal/platform/middleware/` |
| platform/nats | 3 | `backend/internal/platform/nats/` |
| platform/observability | 7 | `backend/internal/platform/observability/` |
| platform/outbox | 4 | `backend/internal/platform/outbox/` |
| platform/telegram | 1 | `backend/internal/platform/telegram/` |
| project | 28 | `backend/internal/project/` |
| rbac | 7 | `backend/internal/rbac/` |
| search | 25 | `backend/internal/search/` |
| sprint | 9 | `backend/internal/sprint/` |
| status | 2 | `backend/internal/status/` |
| storage | 9 | `backend/internal/storage/` |
| tms | 52 | `backend/internal/tms/` |
| tracker | 50 | `backend/internal/tracker/` |
| ws | 9 | `backend/internal/ws/` |
| **TOTAL** | **515** | |

Топ по объёму: **tms (52)**, **tracker (50)**, **messenger (41)**, **notification (36)**, **ai (30)**, **auth/project (28)**.

---

## Особые случаи

1. **`env`** — `builder.go` + `builder_test.go` живут в `testframework/env` как fluent Ent/sqlite fixture builder. Прод-пакета `internal/env` нет. Target: `internal/testutil/env/` **или** поглощение фикстурами Python `/tests` (Э5). Не класть в случайный домен.
2. **`onec`** → `internal/integrations/onec/` (не `internal/onec/`).
3. **`status`** → `internal/status/` (прод-пакет есть: handler/service/dto; `export_hooks` нет).
4. **Platform без `export_hooks`:** `cron`, `crypto`, `errors`, `nats`, `telegram` — тесты уже на публичном API; после co-locate hooks не появятся.
5. После переезда: сменить `package <domain>_test` → `package <domain>` (white-box) **или** оставить black-box в том же каталоге — от этого зависит судьба `export_hooks.go` (см. [E2](TF-MIGRATE-E2-EXPORT-HOOKS.md)).
6. Build-тег `hubtest` сегодня нужен, потому что `Export*` существуют только под `//go:build hubtest`. После white-box co-locate тег на юнитах обычно снимается (критерий Э3: `go test ./internal/...`).

---

## Полная карта: current → target → domain

| # | Current path | Target path | Domain |
|---|---|---|---|
| 1 | `backend/testframework/accounting/fixtures_test.go` | `backend/internal/accounting/fixtures_test.go` | accounting |
| 2 | `backend/testframework/accounting/handler_test.go` | `backend/internal/accounting/handler_test.go` | accounting |
| 3 | `backend/testframework/accounting/handler_wave108_test.go` | `backend/internal/accounting/handler_wave108_test.go` | accounting |
| 4 | `backend/testframework/accounting/handler_wave98_test.go` | `backend/internal/accounting/handler_wave98_test.go` | accounting |
| 5 | `backend/testframework/accounting/pay_concurrent_a6002_test.go` | `backend/internal/accounting/pay_concurrent_a6002_test.go` | accounting |
| 6 | `backend/testframework/accounting/service_invoice_test.go` | `backend/internal/accounting/service_invoice_test.go` | accounting |
| 7 | `backend/testframework/accounting/service_wave108_test.go` | `backend/internal/accounting/service_wave108_test.go` | accounting |
| 8 | `backend/testframework/accounting/service_wave98_test.go` | `backend/internal/accounting/service_wave98_test.go` | accounting |
| 9 | `backend/testframework/ai/adapter_anthropic_test.go` | `backend/internal/ai/adapter_anthropic_test.go` | ai |
| 10 | `backend/testframework/ai/adapter_http_test.go` | `backend/internal/ai/adapter_http_test.go` | ai |
| 11 | `backend/testframework/ai/adapter_openai_test.go` | `backend/internal/ai/adapter_openai_test.go` | ai |
| 12 | `backend/testframework/ai/adapter_wave110_test.go` | `backend/internal/ai/adapter_wave110_test.go` | ai |
| 13 | `backend/testframework/ai/adapter_yandex_test.go` | `backend/internal/ai/adapter_yandex_test.go` | ai |
| 14 | `backend/testframework/ai/admin_handler_test.go` | `backend/internal/ai/admin_handler_test.go` | ai |
| 15 | `backend/testframework/ai/admin_handler_wave110_test.go` | `backend/internal/ai/admin_handler_wave110_test.go` | ai |
| 16 | `backend/testframework/ai/admin_settings_test.go` | `backend/internal/ai/admin_settings_test.go` | ai |
| 17 | `backend/testframework/ai/audit_test.go` | `backend/internal/ai/audit_test.go` | ai |
| 18 | `backend/testframework/ai/bot_commands_test.go` | `backend/internal/ai/bot_commands_test.go` | ai |
| 19 | `backend/testframework/ai/bot_user_test.go` | `backend/internal/ai/bot_user_test.go` | ai |
| 20 | `backend/testframework/ai/code_review_test.go` | `backend/internal/ai/code_review_test.go` | ai |
| 21 | `backend/testframework/ai/code_review_unified_test.go` | `backend/internal/ai/code_review_unified_test.go` | ai |
| 22 | `backend/testframework/ai/composer_suggest_test.go` | `backend/internal/ai/composer_suggest_test.go` | ai |
| 23 | `backend/testframework/ai/content_policy_test.go` | `backend/internal/ai/content_policy_test.go` | ai |
| 24 | `backend/testframework/ai/coverage_wave81_test.go` | `backend/internal/ai/coverage_wave81_test.go` | ai |
| 25 | `backend/testframework/ai/doc_search_test.go` | `backend/internal/ai/doc_search_test.go` | ai |
| 26 | `backend/testframework/ai/fallback_test.go` | `backend/internal/ai/fallback_test.go` | ai |
| 27 | `backend/testframework/ai/handler_extended_test.go` | `backend/internal/ai/handler_extended_test.go` | ai |
| 28 | `backend/testframework/ai/handler_wave115_test.go` | `backend/internal/ai/handler_wave115_test.go` | ai |
| 29 | `backend/testframework/ai/issue_lookup_test.go` | `backend/internal/ai/issue_lookup_test.go` | ai |
| 30 | `backend/testframework/ai/messenger_bot_flow_test.go` | `backend/internal/ai/messenger_bot_flow_test.go` | ai |
| 31 | `backend/testframework/ai/messenger_bot_test.go` | `backend/internal/ai/messenger_bot_test.go` | ai |
| 32 | `backend/testframework/ai/privacy_test.go` | `backend/internal/ai/privacy_test.go` | ai |
| 33 | `backend/testframework/ai/provider_rpm_test.go` | `backend/internal/ai/provider_rpm_test.go` | ai |
| 34 | `backend/testframework/ai/repository_ent_test.go` | `backend/internal/ai/repository_ent_test.go` | ai |
| 35 | `backend/testframework/ai/service_test.go` | `backend/internal/ai/service_test.go` | ai |
| 36 | `backend/testframework/ai/settings_store_test.go` | `backend/internal/ai/settings_store_test.go` | ai |
| 37 | `backend/testframework/ai/stream_test.go` | `backend/internal/ai/stream_test.go` | ai |
| 38 | `backend/testframework/ai/user_handler_test.go` | `backend/internal/ai/user_handler_test.go` | ai |
| 39 | `backend/testframework/apigateway/apikey_auth_test.go` | `backend/internal/apigateway/apikey_auth_test.go` | apigateway |
| 40 | `backend/testframework/apigateway/apikey_test.go` | `backend/internal/apigateway/apikey_test.go` | apigateway |
| 41 | `backend/testframework/apigateway/handler_extended_test.go` | `backend/internal/apigateway/handler_extended_test.go` | apigateway |
| 42 | `backend/testframework/apigateway/handler_test.go` | `backend/internal/apigateway/handler_test.go` | apigateway |
| 43 | `backend/testframework/apigateway/handler_webhook_test.go` | `backend/internal/apigateway/handler_webhook_test.go` | apigateway |
| 44 | `backend/testframework/apigateway/repository_extra_test.go` | `backend/internal/apigateway/repository_extra_test.go` | apigateway |
| 45 | `backend/testframework/apigateway/repository_test.go` | `backend/internal/apigateway/repository_test.go` | apigateway |
| 46 | `backend/testframework/apigateway/webhook_dispatcher_apigw02_test.go` | `backend/internal/apigateway/webhook_dispatcher_apigw02_test.go` | apigateway |
| 47 | `backend/testframework/apigateway/webhook_dispatcher_apigw03_test.go` | `backend/internal/apigateway/webhook_dispatcher_apigw03_test.go` | apigateway |
| 48 | `backend/testframework/apigateway/webhook_dispatcher_wave97_test.go` | `backend/internal/apigateway/webhook_dispatcher_wave97_test.go` | apigateway |
| 49 | `backend/testframework/apigateway/webhook_signing_test.go` | `backend/internal/apigateway/webhook_signing_test.go` | apigateway |
| 50 | `backend/testframework/apigateway/webhook_test.go` | `backend/internal/apigateway/webhook_test.go` | apigateway |
| 51 | `backend/testframework/audit/consumer_lifecycle_test.go` | `backend/internal/audit/consumer_lifecycle_test.go` | audit |
| 52 | `backend/testframework/audit/consumer_misc_test.go` | `backend/internal/audit/consumer_misc_test.go` | audit |
| 53 | `backend/testframework/audit/consumer_test.go` | `backend/internal/audit/consumer_test.go` | audit |
| 54 | `backend/testframework/audit/dto_mapping_test.go` | `backend/internal/audit/dto_mapping_test.go` | audit |
| 55 | `backend/testframework/audit/gdpr_founder_a1004_test.go` | `backend/internal/audit/gdpr_founder_a1004_test.go` | audit |
| 56 | `backend/testframework/audit/handler_alerts_test.go` | `backend/internal/audit/handler_alerts_test.go` | audit |
| 57 | `backend/testframework/audit/handler_test.go` | `backend/internal/audit/handler_test.go` | audit |
| 58 | `backend/testframework/audit/handler_user_mgmt_test.go` | `backend/internal/audit/handler_user_mgmt_test.go` | audit |
| 59 | `backend/testframework/audit/middleware_test.go` | `backend/internal/audit/middleware_test.go` | audit |
| 60 | `backend/testframework/audit/repository_extra_test.go` | `backend/internal/audit/repository_extra_test.go` | audit |
| 61 | `backend/testframework/audit/repository_test.go` | `backend/internal/audit/repository_test.go` | audit |
| 62 | `backend/testframework/audit/repository_trash_test.go` | `backend/internal/audit/repository_trash_test.go` | audit |
| 63 | `backend/testframework/audit/retention_test.go` | `backend/internal/audit/retention_test.go` | audit |
| 64 | `backend/testframework/audit/service_deactivate_test.go` | `backend/internal/audit/service_deactivate_test.go` | audit |
| 65 | `backend/testframework/audit/service_limits_test.go` | `backend/internal/audit/service_limits_test.go` | audit |
| 66 | `backend/testframework/audit/service_purge_test.go` | `backend/internal/audit/service_purge_test.go` | audit |
| 67 | `backend/testframework/audit/service_security_test.go` | `backend/internal/audit/service_security_test.go` | audit |
| 68 | `backend/testframework/audit/service_test.go` | `backend/internal/audit/service_test.go` | audit |
| 69 | `backend/testframework/audit/service_trash_test.go` | `backend/internal/audit/service_trash_test.go` | audit |
| 70 | `backend/testframework/auth/handler_admin_extra_test.go` | `backend/internal/auth/handler_admin_extra_test.go` | auth |
| 71 | `backend/testframework/auth/handler_admin_test.go` | `backend/internal/auth/handler_admin_test.go` | auth |
| 72 | `backend/testframework/auth/handler_extra_test.go` | `backend/internal/auth/handler_extra_test.go` | auth |
| 73 | `backend/testframework/auth/handler_integration_test.go` | `backend/internal/auth/handler_integration_test.go` | auth |
| 74 | `backend/testframework/auth/handler_invite_coverage_test.go` | `backend/internal/auth/handler_invite_coverage_test.go` | auth |
| 75 | `backend/testframework/auth/handler_invite_test.go` | `backend/internal/auth/handler_invite_test.go` | auth |
| 76 | `backend/testframework/auth/handler_magiclink_runes_test.go` | `backend/internal/auth/handler_magiclink_runes_test.go` | auth |
| 77 | `backend/testframework/auth/handler_organization_test.go` | `backend/internal/auth/handler_organization_test.go` | auth |
| 78 | `backend/testframework/auth/handler_profile_test.go` | `backend/internal/auth/handler_profile_test.go` | auth |
| 79 | `backend/testframework/auth/handler_test.go` | `backend/internal/auth/handler_test.go` | auth |
| 80 | `backend/testframework/auth/handler_wave115_test.go` | `backend/internal/auth/handler_wave115_test.go` | auth |
| 81 | `backend/testframework/auth/magiclink_test.go` | `backend/internal/auth/magiclink_test.go` | auth |
| 82 | `backend/testframework/auth/middleware_failclosed_test.go` | `backend/internal/auth/middleware_failclosed_test.go` | auth |
| 83 | `backend/testframework/auth/middleware_test.go` | `backend/internal/auth/middleware_test.go` | auth |
| 84 | `backend/testframework/auth/repository_organization_test.go` | `backend/internal/auth/repository_organization_test.go` | auth |
| 85 | `backend/testframework/auth/repository_test.go` | `backend/internal/auth/repository_test.go` | auth |
| 86 | `backend/testframework/auth/resend_telegram_test.go` | `backend/internal/auth/resend_telegram_test.go` | auth |
| 87 | `backend/testframework/auth/service_facade_test.go` | `backend/internal/auth/service_facade_test.go` | auth |
| 88 | `backend/testframework/auth/service_login_test.go` | `backend/internal/auth/service_login_test.go` | auth |
| 89 | `backend/testframework/auth/service_organization_test.go` | `backend/internal/auth/service_organization_test.go` | auth |
| 90 | `backend/testframework/auth/service_profile_test.go` | `backend/internal/auth/service_profile_test.go` | auth |
| 91 | `backend/testframework/auth/service_registration_race_test.go` | `backend/internal/auth/service_registration_race_test.go` | auth |
| 92 | `backend/testframework/auth/service_registration_test.go` | `backend/internal/auth/service_registration_test.go` | auth |
| 93 | `backend/testframework/auth/service_session_test.go` | `backend/internal/auth/service_session_test.go` | auth |
| 94 | `backend/testframework/auth/service_test.go` | `backend/internal/auth/service_test.go` | auth |
| 95 | `backend/testframework/auth/telegram_bot_coverage_test.go` | `backend/internal/auth/telegram_bot_coverage_test.go` | auth |
| 96 | `backend/testframework/auth/telegram_bot_polling_test.go` | `backend/internal/auth/telegram_bot_polling_test.go` | auth |
| 97 | `backend/testframework/auth/telegram_widget_test.go` | `backend/internal/auth/telegram_widget_test.go` | auth |
| 98 | `backend/testframework/backup/coverage_wave80_test.go` | `backend/internal/backup/coverage_wave80_test.go` | backup |
| 99 | `backend/testframework/backup/handler_test.go` | `backend/internal/backup/handler_test.go` | backup |
| 100 | `backend/testframework/backup/repository_test.go` | `backend/internal/backup/repository_test.go` | backup |
| 101 | `backend/testframework/backup/service_extended_test.go` | `backend/internal/backup/service_extended_test.go` | backup |
| 102 | `backend/testframework/backup/service_test.go` | `backend/internal/backup/service_test.go` | backup |
| 103 | `backend/testframework/backup/service_wave100_test.go` | `backend/internal/backup/service_wave100_test.go` | backup |
| 104 | `backend/testframework/backup/service_wave98_test.go` | `backend/internal/backup/service_wave98_test.go` | backup |
| 105 | `backend/testframework/backup/service_wave99_test.go` | `backend/internal/backup/service_wave99_test.go` | backup |
| 106 | `backend/testframework/bases/fixtures_test.go` | `backend/internal/bases/fixtures_test.go` | bases |
| 107 | `backend/testframework/bases/handler_flow_test.go` | `backend/internal/bases/handler_flow_test.go` | bases |
| 108 | `backend/testframework/bases/handler_test.go` | `backend/internal/bases/handler_test.go` | bases |
| 109 | `backend/testframework/bases/handler_wave107_test.go` | `backend/internal/bases/handler_wave107_test.go` | bases |
| 110 | `backend/testframework/bases/handler_wave115_test.go` | `backend/internal/bases/handler_wave115_test.go` | bases |
| 111 | `backend/testframework/bases/rbac_test.go` | `backend/internal/bases/rbac_test.go` | bases |
| 112 | `backend/testframework/bases/repository_wave98_test.go` | `backend/internal/bases/repository_wave98_test.go` | bases |
| 113 | `backend/testframework/bases/service_test.go` | `backend/internal/bases/service_test.go` | bases |
| 114 | `backend/testframework/bases/service_wave115_test.go` | `backend/internal/bases/service_wave115_test.go` | bases |
| 115 | `backend/testframework/bases/service_wave98_test.go` | `backend/internal/bases/service_wave98_test.go` | bases |
| 116 | `backend/testframework/calendar/handler_test.go` | `backend/internal/calendar/handler_test.go` | calendar |
| 117 | `backend/testframework/calendar/service_wave98_test.go` | `backend/internal/calendar/service_wave98_test.go` | calendar |
| 118 | `backend/testframework/calls/chat_mem_test.go` | `backend/internal/calls/chat_mem_test.go` | calls |
| 119 | `backend/testframework/calls/chat_mongo_wave103_test.go` | `backend/internal/calls/chat_mongo_wave103_test.go` | calls |
| 120 | `backend/testframework/calls/chat_test.go` | `backend/internal/calls/chat_test.go` | calls |
| 121 | `backend/testframework/calls/handler_flow_test.go` | `backend/internal/calls/handler_flow_test.go` | calls |
| 122 | `backend/testframework/calls/handler_rbac_recording_test.go` | `backend/internal/calls/handler_rbac_recording_test.go` | calls |
| 123 | `backend/testframework/calls/handler_recording_test.go` | `backend/internal/calls/handler_recording_test.go` | calls |
| 124 | `backend/testframework/calls/handler_routes_test.go` | `backend/internal/calls/handler_routes_test.go` | calls |
| 125 | `backend/testframework/calls/handler_scheduled_coverage_test.go` | `backend/internal/calls/handler_scheduled_coverage_test.go` | calls |
| 126 | `backend/testframework/calls/handler_validation_test.go` | `backend/internal/calls/handler_validation_test.go` | calls |
| 127 | `backend/testframework/calls/recording_egress_mock_test.go` | `backend/internal/calls/recording_egress_mock_test.go` | calls |
| 128 | `backend/testframework/calls/recording_service_coverage_test.go` | `backend/internal/calls/recording_service_coverage_test.go` | calls |
| 129 | `backend/testframework/calls/recording_test.go` | `backend/internal/calls/recording_test.go` | calls |
| 130 | `backend/testframework/calls/recording_webhook_test.go` | `backend/internal/calls/recording_webhook_test.go` | calls |
| 131 | `backend/testframework/calls/repository_test.go` | `backend/internal/calls/repository_test.go` | calls |
| 132 | `backend/testframework/calls/service_chat_flow_test.go` | `backend/internal/calls/service_chat_flow_test.go` | calls |
| 133 | `backend/testframework/calls/service_extra_test.go` | `backend/internal/calls/service_extra_test.go` | calls |
| 134 | `backend/testframework/calls/service_flow_test.go` | `backend/internal/calls/service_flow_test.go` | calls |
| 135 | `backend/testframework/calls/service_recording_scheduled_test.go` | `backend/internal/calls/service_recording_scheduled_test.go` | calls |
| 136 | `backend/testframework/calls/service_room_test.go` | `backend/internal/calls/service_room_test.go` | calls |
| 137 | `backend/testframework/calls/service_test.go` | `backend/internal/calls/service_test.go` | calls |
| 138 | `backend/testframework/calls/webhook_events_test.go` | `backend/internal/calls/webhook_events_test.go` | calls |
| 139 | `backend/testframework/crm/fixtures_test.go` | `backend/internal/crm/fixtures_test.go` | crm |
| 140 | `backend/testframework/crm/handler_test.go` | `backend/internal/crm/handler_test.go` | crm |
| 141 | `backend/testframework/crm/handler_wave101_test.go` | `backend/internal/crm/handler_wave101_test.go` | crm |
| 142 | `backend/testframework/crm/handler_wave107_test.go` | `backend/internal/crm/handler_wave107_test.go` | crm |
| 143 | `backend/testframework/crm/handler_wave111_test.go` | `backend/internal/crm/handler_wave111_test.go` | crm |
| 144 | `backend/testframework/crm/rbac_a6001_test.go` | `backend/internal/crm/rbac_a6001_test.go` | crm |
| 145 | `backend/testframework/crm/service_invoice_test.go` | `backend/internal/crm/service_invoice_test.go` | crm |
| 146 | `backend/testframework/crm/service_test.go` | `backend/internal/crm/service_test.go` | crm |
| 147 | `backend/testframework/crm/service_wave101_test.go` | `backend/internal/crm/service_wave101_test.go` | crm |
| 148 | `backend/testframework/crm/service_wave111_test.go` | `backend/internal/crm/service_wave111_test.go` | crm |
| 149 | `backend/testframework/dashboard/handler_test.go` | `backend/internal/dashboard/handler_test.go` | dashboard |
| 150 | `backend/testframework/dashboard/repository_extra_test.go` | `backend/internal/dashboard/repository_extra_test.go` | dashboard |
| 151 | `backend/testframework/dashboard/repository_test.go` | `backend/internal/dashboard/repository_test.go` | dashboard |
| 152 | `backend/testframework/dashboard/service_extended_test.go` | `backend/internal/dashboard/service_extended_test.go` | dashboard |
| 153 | `backend/testframework/dashboard/service_test.go` | `backend/internal/dashboard/service_test.go` | dashboard |
| 154 | `backend/testframework/dashboard/service_wave107_test.go` | `backend/internal/dashboard/service_wave107_test.go` | dashboard |
| 155 | `backend/testframework/dev/apidocs_wave102_test.go` | `backend/internal/dev/apidocs_wave102_test.go` | dev |
| 156 | `backend/testframework/dev/handler_wave102_test.go` | `backend/internal/dev/handler_wave102_test.go` | dev |
| 157 | `backend/testframework/dev/seed_test.go` | `backend/internal/dev/seed_test.go` | dev |
| 158 | `backend/testframework/env/builder_test.go` | `backend/internal/testutil/env/builder_test.go  OR keep as shared test helper until Python /tests absorbs fixtures` | env (test helper — not production domain) |
| 159 | `backend/testframework/github/converters_test.go` | `backend/internal/github/converters_test.go` | github |
| 160 | `backend/testframework/github/ent_repository_test.go` | `backend/internal/github/ent_repository_test.go` | github |
| 161 | `backend/testframework/github/handler_extended_test.go` | `backend/internal/github/handler_extended_test.go` | github |
| 162 | `backend/testframework/github/handler_validation_test.go` | `backend/internal/github/handler_validation_test.go` | github |
| 163 | `backend/testframework/github/handler_wave101_test.go` | `backend/internal/github/handler_wave101_test.go` | github |
| 164 | `backend/testframework/github/handler_wave103_test.go` | `backend/internal/github/handler_wave103_test.go` | github |
| 165 | `backend/testframework/github/pr_review_test.go` | `backend/internal/github/pr_review_test.go` | github |
| 166 | `backend/testframework/github/preview_test.go` | `backend/internal/github/preview_test.go` | github |
| 167 | `backend/testframework/github/preview_wave103_test.go` | `backend/internal/github/preview_wave103_test.go` | github |
| 168 | `backend/testframework/github/repository_helpers_test.go` | `backend/internal/github/repository_helpers_test.go` | github |
| 169 | `backend/testframework/github/repository_wave101_test.go` | `backend/internal/github/repository_wave101_test.go` | github |
| 170 | `backend/testframework/github/repository_wave103_test.go` | `backend/internal/github/repository_wave103_test.go` | github |
| 171 | `backend/testframework/github/service_http_test.go` | `backend/internal/github/service_http_test.go` | github |
| 172 | `backend/testframework/github/service_mock_test.go` | `backend/internal/github/service_mock_test.go` | github |
| 173 | `backend/testframework/github/service_test.go` | `backend/internal/github/service_test.go` | github |
| 174 | `backend/testframework/github/service_wave104_test.go` | `backend/internal/github/service_wave104_test.go` | github |
| 175 | `backend/testframework/github/service_wave115_test.go` | `backend/internal/github/service_wave115_test.go` | github |
| 176 | `backend/testframework/gitlab/converters_test.go` | `backend/internal/gitlab/converters_test.go` | gitlab |
| 177 | `backend/testframework/gitlab/ent_repository_test.go` | `backend/internal/gitlab/ent_repository_test.go` | gitlab |
| 178 | `backend/testframework/gitlab/handler_extended_test.go` | `backend/internal/gitlab/handler_extended_test.go` | gitlab |
| 179 | `backend/testframework/gitlab/handler_test.go` | `backend/internal/gitlab/handler_test.go` | gitlab |
| 180 | `backend/testframework/gitlab/handler_wave115_test.go` | `backend/internal/gitlab/handler_wave115_test.go` | gitlab |
| 181 | `backend/testframework/gitlab/mr_review_test.go` | `backend/internal/gitlab/mr_review_test.go` | gitlab |
| 182 | `backend/testframework/gitlab/service_http_test.go` | `backend/internal/gitlab/service_http_test.go` | gitlab |
| 183 | `backend/testframework/gitlab/service_mock_test.go` | `backend/internal/gitlab/service_mock_test.go` | gitlab |
| 184 | `backend/testframework/gitlab/service_test.go` | `backend/internal/gitlab/service_test.go` | gitlab |
| 185 | `backend/testframework/hr/fixtures_test.go` | `backend/internal/hr/fixtures_test.go` | hr |
| 186 | `backend/testframework/hr/handler_test.go` | `backend/internal/hr/handler_test.go` | hr |
| 187 | `backend/testframework/hr/handler_wave101_test.go` | `backend/internal/hr/handler_wave101_test.go` | hr |
| 188 | `backend/testframework/hr/handler_wave111_test.go` | `backend/internal/hr/handler_wave111_test.go` | hr |
| 189 | `backend/testframework/hr/repository_ent_test.go` | `backend/internal/hr/repository_ent_test.go` | hr |
| 190 | `backend/testframework/hr/service_extended_test.go` | `backend/internal/hr/service_extended_test.go` | hr |
| 191 | `backend/testframework/hr/service_wave111_test.go` | `backend/internal/hr/service_wave111_test.go` | hr |
| 192 | `backend/testframework/imex/handler_flow_test.go` | `backend/internal/imex/handler_flow_test.go` | imex |
| 193 | `backend/testframework/imex/jira_import_flow_test.go` | `backend/internal/imex/jira_import_flow_test.go` | imex |
| 194 | `backend/testframework/imex/jira_import_test.go` | `backend/internal/imex/jira_import_test.go` | imex |
| 195 | `backend/testframework/imex/pdf_export_test.go` | `backend/internal/imex/pdf_export_test.go` | imex |
| 196 | `backend/testframework/imex/repository_pdf_test.go` | `backend/internal/imex/repository_pdf_test.go` | imex |
| 197 | `backend/testframework/imex/repository_test.go` | `backend/internal/imex/repository_test.go` | imex |
| 198 | `backend/testframework/imex/service_extended_test.go` | `backend/internal/imex/service_extended_test.go` | imex |
| 199 | `backend/testframework/imex/service_test.go` | `backend/internal/imex/service_test.go` | imex |
| 200 | `backend/testframework/intranet/coffee_repository_test.go` | `backend/internal/intranet/coffee_repository_test.go` | intranet |
| 201 | `backend/testframework/intranet/coffee_service_test.go` | `backend/internal/intranet/coffee_service_test.go` | intranet |
| 202 | `backend/testframework/intranet/dto_test.go` | `backend/internal/intranet/dto_test.go` | intranet |
| 203 | `backend/testframework/intranet/fixtures_test.go` | `backend/internal/intranet/fixtures_test.go` | intranet |
| 204 | `backend/testframework/intranet/handler_test.go` | `backend/internal/intranet/handler_test.go` | intranet |
| 205 | `backend/testframework/intranet/handler_wave91_test.go` | `backend/internal/intranet/handler_wave91_test.go` | intranet |
| 206 | `backend/testframework/intranet/service_test.go` | `backend/internal/intranet/service_test.go` | intranet |
| 207 | `backend/testframework/intranet/service_wave114_test.go` | `backend/internal/intranet/service_wave114_test.go` | intranet |
| 208 | `backend/testframework/mail/admin_handler_test.go` | `backend/internal/mail/admin_handler_test.go` | mail |
| 209 | `backend/testframework/mail/admin_service_test.go` | `backend/internal/mail/admin_service_test.go` | mail |
| 210 | `backend/testframework/mail/config_test.go` | `backend/internal/mail/config_test.go` | mail |
| 211 | `backend/testframework/mail/hub_handler_test.go` | `backend/internal/mail/hub_handler_test.go` | mail |
| 212 | `backend/testframework/mail/internal_auth_test.go` | `backend/internal/mail/internal_auth_test.go` | mail |
| 213 | `backend/testframework/marketplace/dto_wave99_test.go` | `backend/internal/marketplace/dto_wave99_test.go` | marketplace |
| 214 | `backend/testframework/marketplace/fixtures_test.go` | `backend/internal/marketplace/fixtures_test.go` | marketplace |
| 215 | `backend/testframework/marketplace/handler_test.go` | `backend/internal/marketplace/handler_test.go` | marketplace |
| 216 | `backend/testframework/marketplace/handler_wave105_test.go` | `backend/internal/marketplace/handler_wave105_test.go` | marketplace |
| 217 | `backend/testframework/marketplace/handler_wave97_test.go` | `backend/internal/marketplace/handler_wave97_test.go` | marketplace |
| 218 | `backend/testframework/marketplace/service_test.go` | `backend/internal/marketplace/service_test.go` | marketplace |
| 219 | `backend/testframework/marketplace/service_wave105_test.go` | `backend/internal/marketplace/service_wave105_test.go` | marketplace |
| 220 | `backend/testframework/messenger/auto_channels_service_test.go` | `backend/internal/messenger/auto_channels_service_test.go` | messenger |
| 221 | `backend/testframework/messenger/auto_channels_test.go` | `backend/internal/messenger/auto_channels_test.go` | messenger |
| 222 | `backend/testframework/messenger/dto_test.go` | `backend/internal/messenger/dto_test.go` | messenger |
| 223 | `backend/testframework/messenger/ent_repository_invites_test.go` | `backend/internal/messenger/ent_repository_invites_test.go` | messenger |
| 224 | `backend/testframework/messenger/ent_repository_poll_test.go` | `backend/internal/messenger/ent_repository_poll_test.go` | messenger |
| 225 | `backend/testframework/messenger/ent_repository_test.go` | `backend/internal/messenger/ent_repository_test.go` | messenger |
| 226 | `backend/testframework/messenger/ent_repository_wave123_test.go` | `backend/internal/messenger/ent_repository_wave123_test.go` | messenger |
| 227 | `backend/testframework/messenger/handler_channels_extra_test.go` | `backend/internal/messenger/handler_channels_extra_test.go` | messenger |
| 228 | `backend/testframework/messenger/handler_flow_test.go` | `backend/internal/messenger/handler_flow_test.go` | messenger |
| 229 | `backend/testframework/messenger/handler_helpers_test.go` | `backend/internal/messenger/handler_helpers_test.go` | messenger |
| 230 | `backend/testframework/messenger/handler_parse_test.go` | `backend/internal/messenger/handler_parse_test.go` | messenger |
| 231 | `backend/testframework/messenger/handler_wave100_test.go` | `backend/internal/messenger/handler_wave100_test.go` | messenger |
| 232 | `backend/testframework/messenger/handler_wave106_test.go` | `backend/internal/messenger/handler_wave106_test.go` | messenger |
| 233 | `backend/testframework/messenger/handler_wave109_test.go` | `backend/internal/messenger/handler_wave109_test.go` | messenger |
| 234 | `backend/testframework/messenger/handler_wave112_test.go` | `backend/internal/messenger/handler_wave112_test.go` | messenger |
| 235 | `backend/testframework/messenger/handler_wave113_test.go` | `backend/internal/messenger/handler_wave113_test.go` | messenger |
| 236 | `backend/testframework/messenger/handler_wave114_test.go` | `backend/internal/messenger/handler_wave114_test.go` | messenger |
| 237 | `backend/testframework/messenger/handler_wave115_test.go` | `backend/internal/messenger/handler_wave115_test.go` | messenger |
| 238 | `backend/testframework/messenger/handler_wave116_test.go` | `backend/internal/messenger/handler_wave116_test.go` | messenger |
| 239 | `backend/testframework/messenger/handler_wave123_test.go` | `backend/internal/messenger/handler_wave123_test.go` | messenger |
| 240 | `backend/testframework/messenger/handler_wave93_test.go` | `backend/internal/messenger/handler_wave93_test.go` | messenger |
| 241 | `backend/testframework/messenger/mock_repository_test.go` | `backend/internal/messenger/mock_repository_test.go` | messenger |
| 242 | `backend/testframework/messenger/mongo_repository_test.go` | `backend/internal/messenger/mongo_repository_test.go` | messenger |
| 243 | `backend/testframework/messenger/mongo_repository_wave123_test.go` | `backend/internal/messenger/mongo_repository_wave123_test.go` | messenger |
| 244 | `backend/testframework/messenger/mongo_search_empty_scope_a2001_test.go` | `backend/internal/messenger/mongo_search_empty_scope_a2001_test.go` | messenger |
| 245 | `backend/testframework/messenger/presence_test.go` | `backend/internal/messenger/presence_test.go` | messenger |
| 246 | `backend/testframework/messenger/presence_wave100_test.go` | `backend/internal/messenger/presence_wave100_test.go` | messenger |
| 247 | `backend/testframework/messenger/service_bot_test.go` | `backend/internal/messenger/service_bot_test.go` | messenger |
| 248 | `backend/testframework/messenger/service_extended_test.go` | `backend/internal/messenger/service_extended_test.go` | messenger |
| 249 | `backend/testframework/messenger/service_fanout_test.go` | `backend/internal/messenger/service_fanout_test.go` | messenger |
| 250 | `backend/testframework/messenger/service_mentions_test.go` | `backend/internal/messenger/service_mentions_test.go` | messenger |
| 251 | `backend/testframework/messenger/service_messages_test.go` | `backend/internal/messenger/service_messages_test.go` | messenger |
| 252 | `backend/testframework/messenger/service_search_a2001_test.go` | `backend/internal/messenger/service_search_a2001_test.go` | messenger |
| 253 | `backend/testframework/messenger/service_test.go` | `backend/internal/messenger/service_test.go` | messenger |
| 254 | `backend/testframework/messenger/service_wave114_test.go` | `backend/internal/messenger/service_wave114_test.go` | messenger |
| 255 | `backend/testframework/messenger/service_wave115_test.go` | `backend/internal/messenger/service_wave115_test.go` | messenger |
| 256 | `backend/testframework/messenger/service_wave116_test.go` | `backend/internal/messenger/service_wave116_test.go` | messenger |
| 257 | `backend/testframework/messenger/service_wave123_test.go` | `backend/internal/messenger/service_wave123_test.go` | messenger |
| 258 | `backend/testframework/messenger/stubs_test.go` | `backend/internal/messenger/stubs_test.go` | messenger |
| 259 | `backend/testframework/messenger/unread_test.go` | `backend/internal/messenger/unread_test.go` | messenger |
| 260 | `backend/testframework/messenger/validation_test.go` | `backend/internal/messenger/validation_test.go` | messenger |
| 261 | `backend/testframework/notification/consumer_extra_test.go` | `backend/internal/notification/consumer_extra_test.go` | notification |
| 262 | `backend/testframework/notification/consumer_test.go` | `backend/internal/notification/consumer_test.go` | notification |
| 263 | `backend/testframework/notification/consumer_wave102_test.go` | `backend/internal/notification/consumer_wave102_test.go` | notification |
| 264 | `backend/testframework/notification/consumer_wave118_test.go` | `backend/internal/notification/consumer_wave118_test.go` | notification |
| 265 | `backend/testframework/notification/consumer_wave119_test.go` | `backend/internal/notification/consumer_wave119_test.go` | notification |
| 266 | `backend/testframework/notification/consumer_wave121_test.go` | `backend/internal/notification/consumer_wave121_test.go` | notification |
| 267 | `backend/testframework/notification/consumer_wave94_test.go` | `backend/internal/notification/consumer_wave94_test.go` | notification |
| 268 | `backend/testframework/notification/consumer_wave99_test.go` | `backend/internal/notification/consumer_wave99_test.go` | notification |
| 269 | `backend/testframework/notification/coverage_wave79_test.go` | `backend/internal/notification/coverage_wave79_test.go` | notification |
| 270 | `backend/testframework/notification/fixtures_test.go` | `backend/internal/notification/fixtures_test.go` | notification |
| 271 | `backend/testframework/notification/handler_list_filter_test.go` | `backend/internal/notification/handler_list_filter_test.go` | notification |
| 272 | `backend/testframework/notification/handler_telegram_bind_test.go` | `backend/internal/notification/handler_telegram_bind_test.go` | notification |
| 273 | `backend/testframework/notification/handler_telegram_bind_wave102_test.go` | `backend/internal/notification/handler_telegram_bind_wave102_test.go` | notification |
| 274 | `backend/testframework/notification/handler_test.go` | `backend/internal/notification/handler_test.go` | notification |
| 275 | `backend/testframework/notification/handler_wave115_test.go` | `backend/internal/notification/handler_wave115_test.go` | notification |
| 276 | `backend/testframework/notification/handler_wave116_test.go` | `backend/internal/notification/handler_wave116_test.go` | notification |
| 277 | `backend/testframework/notification/handler_wave117_test.go` | `backend/internal/notification/handler_wave117_test.go` | notification |
| 278 | `backend/testframework/notification/handler_wave118_test.go` | `backend/internal/notification/handler_wave118_test.go` | notification |
| 279 | `backend/testframework/notification/handler_wave119_test.go` | `backend/internal/notification/handler_wave119_test.go` | notification |
| 280 | `backend/testframework/notification/handler_wave99_test.go` | `backend/internal/notification/handler_wave99_test.go` | notification |
| 281 | `backend/testframework/notification/push_lifecycle_test.go` | `backend/internal/notification/push_lifecycle_test.go` | notification |
| 282 | `backend/testframework/notification/push_wave119_test.go` | `backend/internal/notification/push_wave119_test.go` | notification |
| 283 | `backend/testframework/notification/push_wave95_test.go` | `backend/internal/notification/push_wave95_test.go` | notification |
| 284 | `backend/testframework/notification/repository_extra_test.go` | `backend/internal/notification/repository_extra_test.go` | notification |
| 285 | `backend/testframework/notification/service_extended_test.go` | `backend/internal/notification/service_extended_test.go` | notification |
| 286 | `backend/testframework/notification/service_test.go` | `backend/internal/notification/service_test.go` | notification |
| 287 | `backend/testframework/notification/service_wave111_test.go` | `backend/internal/notification/service_wave111_test.go` | notification |
| 288 | `backend/testframework/notification/service_wave117_test.go` | `backend/internal/notification/service_wave117_test.go` | notification |
| 289 | `backend/testframework/notification/service_wave118_test.go` | `backend/internal/notification/service_wave118_test.go` | notification |
| 290 | `backend/testframework/notification/service_wave119_test.go` | `backend/internal/notification/service_wave119_test.go` | notification |
| 291 | `backend/testframework/notification/stubs_test.go` | `backend/internal/notification/stubs_test.go` | notification |
| 292 | `backend/testframework/notification/telegram_bind_test.go` | `backend/internal/notification/telegram_bind_test.go` | notification |
| 293 | `backend/testframework/notification/telegram_bind_wave102_test.go` | `backend/internal/notification/telegram_bind_wave102_test.go` | notification |
| 294 | `backend/testframework/notification/telegram_bind_wave95_test.go` | `backend/internal/notification/telegram_bind_wave95_test.go` | notification |
| 295 | `backend/testframework/notification/telegram_lifecycle_test.go` | `backend/internal/notification/telegram_lifecycle_test.go` | notification |
| 296 | `backend/testframework/notification/telegram_wave97_test.go` | `backend/internal/notification/telegram_wave97_test.go` | notification |
| 297 | `backend/testframework/onec/fixtures_test.go` | `backend/internal/integrations/onec/fixtures_test.go` | integrations/onec |
| 298 | `backend/testframework/onec/handler_test.go` | `backend/internal/integrations/onec/handler_test.go` | integrations/onec |
| 299 | `backend/testframework/onec/service_extended_test.go` | `backend/internal/integrations/onec/service_extended_test.go` | integrations/onec |
| 300 | `backend/testframework/onec/settings_store_wave101_test.go` | `backend/internal/integrations/onec/settings_store_wave101_test.go` | integrations/onec |
| 301 | `backend/testframework/platform/config/config_test.go` | `backend/internal/platform/config/config_test.go` | platform/config |
| 302 | `backend/testframework/platform/cron/cron_test.go` | `backend/internal/platform/cron/cron_test.go` | platform/cron |
| 303 | `backend/testframework/platform/crypto/aesgcm_test.go` | `backend/internal/platform/crypto/aesgcm_test.go` | platform/crypto |
| 304 | `backend/testframework/platform/errors/errors_test.go` | `backend/internal/platform/errors/errors_test.go` | platform/errors |
| 305 | `backend/testframework/platform/errors/errors_wave101_test.go` | `backend/internal/platform/errors/errors_wave101_test.go` | platform/errors |
| 306 | `backend/testframework/platform/http/helpers_test.go` | `backend/internal/platform/http/helpers_test.go` | platform/http |
| 307 | `backend/testframework/platform/http/opengraph_test.go` | `backend/internal/platform/http/opengraph_test.go` | platform/http |
| 308 | `backend/testframework/platform/http/safedialer_test.go` | `backend/internal/platform/http/safedialer_test.go` | platform/http |
| 309 | `backend/testframework/platform/middleware/middleware_test.go` | `backend/internal/platform/middleware/middleware_test.go` | platform/middleware |
| 310 | `backend/testframework/platform/nats/client_integration_test.go` | `backend/internal/platform/nats/client_integration_test.go` | platform/nats |
| 311 | `backend/testframework/platform/nats/client_publish_test.go` | `backend/internal/platform/nats/client_publish_test.go` | platform/nats |
| 312 | `backend/testframework/platform/nats/client_test.go` | `backend/internal/platform/nats/client_test.go` | platform/nats |
| 313 | `backend/testframework/platform/observability/metrics_allow_middleware_test.go` | `backend/internal/platform/observability/metrics_allow_middleware_test.go` | platform/observability |
| 314 | `backend/testframework/platform/observability/metrics_allowlist_test.go` | `backend/internal/platform/observability/metrics_allowlist_test.go` | platform/observability |
| 315 | `backend/testframework/platform/observability/metrics_test.go` | `backend/internal/platform/observability/metrics_test.go` | platform/observability |
| 316 | `backend/testframework/platform/observability/otel_test.go` | `backend/internal/platform/observability/otel_test.go` | platform/observability |
| 317 | `backend/testframework/platform/observability/prometheus_test.go` | `backend/internal/platform/observability/prometheus_test.go` | platform/observability |
| 318 | `backend/testframework/platform/observability/timing_test.go` | `backend/internal/platform/observability/timing_test.go` | platform/observability |
| 319 | `backend/testframework/platform/observability/trace_propagation_test.go` | `backend/internal/platform/observability/trace_propagation_test.go` | platform/observability |
| 320 | `backend/testframework/platform/outbox/outbox_test.go` | `backend/internal/platform/outbox/outbox_test.go` | platform/outbox |
| 321 | `backend/testframework/platform/outbox/publisher_test.go` | `backend/internal/platform/outbox/publisher_test.go` | platform/outbox |
| 322 | `backend/testframework/platform/outbox/trace_context_test.go` | `backend/internal/platform/outbox/trace_context_test.go` | platform/outbox |
| 323 | `backend/testframework/platform/outbox/worker_wave101_test.go` | `backend/internal/platform/outbox/worker_wave101_test.go` | platform/outbox |
| 324 | `backend/testframework/platform/telegram/markdown_test.go` | `backend/internal/platform/telegram/markdown_test.go` | platform/telegram |
| 325 | `backend/testframework/project/handler_flow_test.go` | `backend/internal/project/handler_flow_test.go` | project |
| 326 | `backend/testframework/project/handler_wave110_test.go` | `backend/internal/project/handler_wave110_test.go` | project |
| 327 | `backend/testframework/project/handler_wave113_test.go` | `backend/internal/project/handler_wave113_test.go` | project |
| 328 | `backend/testframework/project/handler_wave120_test.go` | `backend/internal/project/handler_wave120_test.go` | project |
| 329 | `backend/testframework/project/handler_wave122_routes_test.go` | `backend/internal/project/handler_wave122_routes_test.go` | project |
| 330 | `backend/testframework/project/handler_wave122_test.go` | `backend/internal/project/handler_wave122_test.go` | project |
| 331 | `backend/testframework/project/provisioner_test.go` | `backend/internal/project/provisioner_test.go` | project |
| 332 | `backend/testframework/project/repository_ent_extended_test.go` | `backend/internal/project/repository_ent_extended_test.go` | project |
| 333 | `backend/testframework/project/repository_ent_test.go` | `backend/internal/project/repository_ent_test.go` | project |
| 334 | `backend/testframework/project/repository_wave124_extra_test.go` | `backend/internal/project/repository_wave124_extra_test.go` | project |
| 335 | `backend/testframework/project/repository_wave124_test.go` | `backend/internal/project/repository_wave124_test.go` | project |
| 336 | `backend/testframework/project/service_extended_test.go` | `backend/internal/project/service_extended_test.go` | project |
| 337 | `backend/testframework/project/service_fa11_test.go` | `backend/internal/project/service_fa11_test.go` | project |
| 338 | `backend/testframework/project/service_test.go` | `backend/internal/project/service_test.go` | project |
| 339 | `backend/testframework/project/service_wave101_test.go` | `backend/internal/project/service_wave101_test.go` | project |
| 340 | `backend/testframework/project/service_wave105_test.go` | `backend/internal/project/service_wave105_test.go` | project |
| 341 | `backend/testframework/project/service_wave113_test.go` | `backend/internal/project/service_wave113_test.go` | project |
| 342 | `backend/testframework/project/service_wave120_test.go` | `backend/internal/project/service_wave120_test.go` | project |
| 343 | `backend/testframework/project/service_wave122_extra_test.go` | `backend/internal/project/service_wave122_extra_test.go` | project |
| 344 | `backend/testframework/project/service_wave122_test.go` | `backend/internal/project/service_wave122_test.go` | project |
| 345 | `backend/testframework/project/service_wave124_test.go` | `backend/internal/project/service_wave124_test.go` | project |
| 346 | `backend/testframework/project/template_export_ent_test.go` | `backend/internal/project/template_export_ent_test.go` | project |
| 347 | `backend/testframework/project/template_handler_fa11_test.go` | `backend/internal/project/template_handler_fa11_test.go` | project |
| 348 | `backend/testframework/project/template_handler_routes_test.go` | `backend/internal/project/template_handler_routes_test.go` | project |
| 349 | `backend/testframework/project/template_handler_wave101_test.go` | `backend/internal/project/template_handler_wave101_test.go` | project |
| 350 | `backend/testframework/project/template_multipart_test.go` | `backend/internal/project/template_multipart_test.go` | project |
| 351 | `backend/testframework/project/templates_test.go` | `backend/internal/project/templates_test.go` | project |
| 352 | `backend/testframework/project/templates_wave105_test.go` | `backend/internal/project/templates_wave105_test.go` | project |
| 353 | `backend/testframework/rbac/cache_test.go` | `backend/internal/rbac/cache_test.go` | rbac |
| 354 | `backend/testframework/rbac/custom_roles_a1002_test.go` | `backend/internal/rbac/custom_roles_a1002_test.go` | rbac |
| 355 | `backend/testframework/rbac/handler_test.go` | `backend/internal/rbac/handler_test.go` | rbac |
| 356 | `backend/testframework/rbac/middleware_integration_test.go` | `backend/internal/rbac/middleware_integration_test.go` | rbac |
| 357 | `backend/testframework/rbac/middleware_test.go` | `backend/internal/rbac/middleware_test.go` | rbac |
| 358 | `backend/testframework/rbac/permissions_test.go` | `backend/internal/rbac/permissions_test.go` | rbac |
| 359 | `backend/testframework/rbac/service_test.go` | `backend/internal/rbac/service_test.go` | rbac |
| 360 | `backend/testframework/search/handler_flow_test.go` | `backend/internal/search/handler_flow_test.go` | search |
| 361 | `backend/testframework/search/handler_wave100_test.go` | `backend/internal/search/handler_wave100_test.go` | search |
| 362 | `backend/testframework/search/handler_wave109_test.go` | `backend/internal/search/handler_wave109_test.go` | search |
| 363 | `backend/testframework/search/handler_wave112_test.go` | `backend/internal/search/handler_wave112_test.go` | search |
| 364 | `backend/testframework/search/handler_wave116_test.go` | `backend/internal/search/handler_wave116_test.go` | search |
| 365 | `backend/testframework/search/handler_wave93_test.go` | `backend/internal/search/handler_wave93_test.go` | search |
| 366 | `backend/testframework/search/jql_predicates_extended_test.go` | `backend/internal/search/jql_predicates_extended_test.go` | search |
| 367 | `backend/testframework/search/jql_test.go` | `backend/internal/search/jql_test.go` | search |
| 368 | `backend/testframework/search/jql_wave100_test.go` | `backend/internal/search/jql_wave100_test.go` | search |
| 369 | `backend/testframework/search/jql_wave116_test.go` | `backend/internal/search/jql_wave116_test.go` | search |
| 370 | `backend/testframework/search/meilisearch_indexer_test.go` | `backend/internal/search/meilisearch_indexer_test.go` | search |
| 371 | `backend/testframework/search/repository_wave100_test.go` | `backend/internal/search/repository_wave100_test.go` | search |
| 372 | `backend/testframework/search/service_query_test.go` | `backend/internal/search/service_query_test.go` | search |
| 373 | `backend/testframework/search/service_scope_test.go` | `backend/internal/search/service_scope_test.go` | search |
| 374 | `backend/testframework/search/service_shard_errors_test.go` | `backend/internal/search/service_shard_errors_test.go` | search |
| 375 | `backend/testframework/search/service_test.go` | `backend/internal/search/service_test.go` | search |
| 376 | `backend/testframework/search/service_wave100_test.go` | `backend/internal/search/service_wave100_test.go` | search |
| 377 | `backend/testframework/search/service_wave103_mongo_test.go` | `backend/internal/search/service_wave103_mongo_test.go` | search |
| 378 | `backend/testframework/search/service_wave106_test.go` | `backend/internal/search/service_wave106_test.go` | search |
| 379 | `backend/testframework/search/service_wave108_test.go` | `backend/internal/search/service_wave108_test.go` | search |
| 380 | `backend/testframework/search/service_wave109_test.go` | `backend/internal/search/service_wave109_test.go` | search |
| 381 | `backend/testframework/search/service_wave114_test.go` | `backend/internal/search/service_wave114_test.go` | search |
| 382 | `backend/testframework/search/service_wave115_test.go` | `backend/internal/search/service_wave115_test.go` | search |
| 383 | `backend/testframework/search/service_wave116_postgres_test.go` | `backend/internal/search/service_wave116_postgres_test.go` | search |
| 384 | `backend/testframework/search/service_wave92_test.go` | `backend/internal/search/service_wave92_test.go` | search |
| 385 | `backend/testframework/sprint/handler_flow_test.go` | `backend/internal/sprint/handler_flow_test.go` | sprint |
| 386 | `backend/testframework/sprint/handler_rbac_test.go` | `backend/internal/sprint/handler_rbac_test.go` | sprint |
| 387 | `backend/testframework/sprint/handler_routes_test.go` | `backend/internal/sprint/handler_routes_test.go` | sprint |
| 388 | `backend/testframework/sprint/handler_wave113_test.go` | `backend/internal/sprint/handler_wave113_test.go` | sprint |
| 389 | `backend/testframework/sprint/repository_ent_test.go` | `backend/internal/sprint/repository_ent_test.go` | sprint |
| 390 | `backend/testframework/sprint/repository_get_in_project_a3001_test.go` | `backend/internal/sprint/repository_get_in_project_a3001_test.go` | sprint |
| 391 | `backend/testframework/sprint/service_extended_test.go` | `backend/internal/sprint/service_extended_test.go` | sprint |
| 392 | `backend/testframework/sprint/service_test.go` | `backend/internal/sprint/service_test.go` | sprint |
| 393 | `backend/testframework/sprint/service_wave114_test.go` | `backend/internal/sprint/service_wave114_test.go` | sprint |
| 394 | `backend/testframework/status/handler_test.go` | `backend/internal/status/handler_test.go` | status |
| 395 | `backend/testframework/status/service_test.go` | `backend/internal/status/service_test.go` | status |
| 396 | `backend/testframework/storage/filename_test.go` | `backend/internal/storage/filename_test.go` | storage |
| 397 | `backend/testframework/storage/handler_extended_test.go` | `backend/internal/storage/handler_extended_test.go` | storage |
| 398 | `backend/testframework/storage/handler_routes_test.go` | `backend/internal/storage/handler_routes_test.go` | storage |
| 399 | `backend/testframework/storage/handler_wave117_test.go` | `backend/internal/storage/handler_wave117_test.go` | storage |
| 400 | `backend/testframework/storage/orphan_cleanup_test.go` | `backend/internal/storage/orphan_cleanup_test.go` | storage |
| 401 | `backend/testframework/storage/repository_test.go` | `backend/internal/storage/repository_test.go` | storage |
| 402 | `backend/testframework/storage/service_ent_test.go` | `backend/internal/storage/service_ent_test.go` | storage |
| 403 | `backend/testframework/storage/service_test.go` | `backend/internal/storage/service_test.go` | storage |
| 404 | `backend/testframework/storage/service_wave117_test.go` | `backend/internal/storage/service_wave117_test.go` | storage |
| 405 | `backend/testframework/tms/allure_matcher_wave99_test.go` | `backend/internal/tms/allure_matcher_wave99_test.go` | tms |
| 406 | `backend/testframework/tms/allure_parser_test.go` | `backend/internal/tms/allure_parser_test.go` | tms |
| 407 | `backend/testframework/tms/coverage_wave_test.go` | `backend/internal/tms/coverage_wave_test.go` | tms |
| 408 | `backend/testframework/tms/create_bug_nil_tc_a3009_test.go` | `backend/internal/tms/create_bug_nil_tc_a3009_test.go` | tms |
| 409 | `backend/testframework/tms/ent_repository_batch_test.go` | `backend/internal/tms/ent_repository_batch_test.go` | tms |
| 410 | `backend/testframework/tms/ent_repository_extended_test.go` | `backend/internal/tms/ent_repository_extended_test.go` | tms |
| 411 | `backend/testframework/tms/ent_repository_list_update_test.go` | `backend/internal/tms/ent_repository_list_update_test.go` | tms |
| 412 | `backend/testframework/tms/ent_repository_test.go` | `backend/internal/tms/ent_repository_test.go` | tms |
| 413 | `backend/testframework/tms/ent_repository_wave122_extra_test.go` | `backend/internal/tms/ent_repository_wave122_extra_test.go` | tms |
| 414 | `backend/testframework/tms/ent_repository_wave122_test.go` | `backend/internal/tms/ent_repository_wave122_test.go` | tms |
| 415 | `backend/testframework/tms/ent_repository_wave124_test.go` | `backend/internal/tms/ent_repository_wave124_test.go` | tms |
| 416 | `backend/testframework/tms/environment_wave105_test.go` | `backend/internal/tms/environment_wave105_test.go` | tms |
| 417 | `backend/testframework/tms/exporter_test.go` | `backend/internal/tms/exporter_test.go` | tms |
| 418 | `backend/testframework/tms/exporter_wave95_test.go` | `backend/internal/tms/exporter_wave95_test.go` | tms |
| 419 | `backend/testframework/tms/find_by_key_a3003_test.go` | `backend/internal/tms/find_by_key_a3003_test.go` | tms |
| 420 | `backend/testframework/tms/handler_flow_test.go` | `backend/internal/tms/handler_flow_test.go` | tms |
| 421 | `backend/testframework/tms/handler_import_flow_test.go` | `backend/internal/tms/handler_import_flow_test.go` | tms |
| 422 | `backend/testframework/tms/handler_routes_test.go` | `backend/internal/tms/handler_routes_test.go` | tms |
| 423 | `backend/testframework/tms/handler_wave108_test.go` | `backend/internal/tms/handler_wave108_test.go` | tms |
| 424 | `backend/testframework/tms/handler_wave110_test.go` | `backend/internal/tms/handler_wave110_test.go` | tms |
| 425 | `backend/testframework/tms/handler_wave113_test.go` | `backend/internal/tms/handler_wave113_test.go` | tms |
| 426 | `backend/testframework/tms/handler_wave114_test.go` | `backend/internal/tms/handler_wave114_test.go` | tms |
| 427 | `backend/testframework/tms/handler_wave115_test.go` | `backend/internal/tms/handler_wave115_test.go` | tms |
| 428 | `backend/testframework/tms/handler_wave120_test.go` | `backend/internal/tms/handler_wave120_test.go` | tms |
| 429 | `backend/testframework/tms/handler_wave122_test.go` | `backend/internal/tms/handler_wave122_test.go` | tms |
| 430 | `backend/testframework/tms/handler_wave124_test.go` | `backend/internal/tms/handler_wave124_test.go` | tms |
| 431 | `backend/testframework/tms/handler_wave88_test.go` | `backend/internal/tms/handler_wave88_test.go` | tms |
| 432 | `backend/testframework/tms/handler_webhook_flow_test.go` | `backend/internal/tms/handler_webhook_flow_test.go` | tms |
| 433 | `backend/testframework/tms/handler_webhook_rbac_test.go` | `backend/internal/tms/handler_webhook_rbac_test.go` | tms |
| 434 | `backend/testframework/tms/handler_webhook_wave99_test.go` | `backend/internal/tms/handler_webhook_wave99_test.go` | tms |
| 435 | `backend/testframework/tms/helpers_test.go` | `backend/internal/tms/helpers_test.go` | tms |
| 436 | `backend/testframework/tms/import_run_scope_a3012_test.go` | `backend/internal/tms/import_run_scope_a3012_test.go` | tms |
| 437 | `backend/testframework/tms/importer_wave96_test.go` | `backend/internal/tms/importer_wave96_test.go` | tms |
| 438 | `backend/testframework/tms/junit_parser_test.go` | `backend/internal/tms/junit_parser_test.go` | tms |
| 439 | `backend/testframework/tms/mongo_repository_test.go` | `backend/internal/tms/mongo_repository_test.go` | tms |
| 440 | `backend/testframework/tms/result_scope_a3008_test.go` | `backend/internal/tms/result_scope_a3008_test.go` | tms |
| 441 | `backend/testframework/tms/rollback_a3007_test.go` | `backend/internal/tms/rollback_a3007_test.go` | tms |
| 442 | `backend/testframework/tms/service_cases_extra_test.go` | `backend/internal/tms/service_cases_extra_test.go` | tms |
| 443 | `backend/testframework/tms/service_converters_test.go` | `backend/internal/tms/service_converters_test.go` | tms |
| 444 | `backend/testframework/tms/service_reports_test.go` | `backend/internal/tms/service_reports_test.go` | tms |
| 445 | `backend/testframework/tms/service_runs_extra_test.go` | `backend/internal/tms/service_runs_extra_test.go` | tms |
| 446 | `backend/testframework/tms/service_suites_test.go` | `backend/internal/tms/service_suites_test.go` | tms |
| 447 | `backend/testframework/tms/service_test.go` | `backend/internal/tms/service_test.go` | tms |
| 448 | `backend/testframework/tms/service_wave113_test.go` | `backend/internal/tms/service_wave113_test.go` | tms |
| 449 | `backend/testframework/tms/service_wave114_test.go` | `backend/internal/tms/service_wave114_test.go` | tms |
| 450 | `backend/testframework/tms/service_wave115_test.go` | `backend/internal/tms/service_wave115_test.go` | tms |
| 451 | `backend/testframework/tms/service_wave116_test.go` | `backend/internal/tms/service_wave116_test.go` | tms |
| 452 | `backend/testframework/tms/service_wave120_test.go` | `backend/internal/tms/service_wave120_test.go` | tms |
| 453 | `backend/testframework/tms/service_wave122_extra_test.go` | `backend/internal/tms/service_wave122_extra_test.go` | tms |
| 454 | `backend/testframework/tms/service_wave124_test.go` | `backend/internal/tms/service_wave124_test.go` | tms |
| 455 | `backend/testframework/tms/testrail_import_wave96_test.go` | `backend/internal/tms/testrail_import_wave96_test.go` | tms |
| 456 | `backend/testframework/tms/update_optimistic_a3004_test.go` | `backend/internal/tms/update_optimistic_a3004_test.go` | tms |
| 457 | `backend/testframework/tracker/automation_ack_a3006_test.go` | `backend/internal/tracker/automation_ack_a3006_test.go` | tracker |
| 458 | `backend/testframework/tracker/automation_consumer_test.go` | `backend/internal/tracker/automation_consumer_test.go` | tracker |
| 459 | `backend/testframework/tracker/automation_engine_events_test.go` | `backend/internal/tracker/automation_engine_events_test.go` | tracker |
| 460 | `backend/testframework/tracker/automation_engine_test.go` | `backend/internal/tracker/automation_engine_test.go` | tracker |
| 461 | `backend/testframework/tracker/automation_engine_wave100_test.go` | `backend/internal/tracker/automation_engine_wave100_test.go` | tracker |
| 462 | `backend/testframework/tracker/automation_service_extra_test.go` | `backend/internal/tracker/automation_service_extra_test.go` | tracker |
| 463 | `backend/testframework/tracker/comment_scope_a3011_test.go` | `backend/internal/tracker/comment_scope_a3011_test.go` | tracker |
| 464 | `backend/testframework/tracker/handler_automation_flow_test.go` | `backend/internal/tracker/handler_automation_flow_test.go` | tracker |
| 465 | `backend/testframework/tracker/handler_flow_test.go` | `backend/internal/tracker/handler_flow_test.go` | tracker |
| 466 | `backend/testframework/tracker/handler_jql_flow_test.go` | `backend/internal/tracker/handler_jql_flow_test.go` | tracker |
| 467 | `backend/testframework/tracker/handler_rbac_test.go` | `backend/internal/tracker/handler_rbac_test.go` | tracker |
| 468 | `backend/testframework/tracker/handler_routes_test.go` | `backend/internal/tracker/handler_routes_test.go` | tracker |
| 469 | `backend/testframework/tracker/handler_wave103_test.go` | `backend/internal/tracker/handler_wave103_test.go` | tracker |
| 470 | `backend/testframework/tracker/handler_wave106_test.go` | `backend/internal/tracker/handler_wave106_test.go` | tracker |
| 471 | `backend/testframework/tracker/handler_wave109_test.go` | `backend/internal/tracker/handler_wave109_test.go` | tracker |
| 472 | `backend/testframework/tracker/handler_wave112_test.go` | `backend/internal/tracker/handler_wave112_test.go` | tracker |
| 473 | `backend/testframework/tracker/handler_wave114_test.go` | `backend/internal/tracker/handler_wave114_test.go` | tracker |
| 474 | `backend/testframework/tracker/handler_wave115_test.go` | `backend/internal/tracker/handler_wave115_test.go` | tracker |
| 475 | `backend/testframework/tracker/handler_wave117_test.go` | `backend/internal/tracker/handler_wave117_test.go` | tracker |
| 476 | `backend/testframework/tracker/handler_wave118_test.go` | `backend/internal/tracker/handler_wave118_test.go` | tracker |
| 477 | `backend/testframework/tracker/handler_wave119_test.go` | `backend/internal/tracker/handler_wave119_test.go` | tracker |
| 478 | `backend/testframework/tracker/handler_wave120_test.go` | `backend/internal/tracker/handler_wave120_test.go` | tracker |
| 479 | `backend/testframework/tracker/handler_wave121_test.go` | `backend/internal/tracker/handler_wave121_test.go` | tracker |
| 480 | `backend/testframework/tracker/handler_wave124_test.go` | `backend/internal/tracker/handler_wave124_test.go` | tracker |
| 481 | `backend/testframework/tracker/handler_wave95_test.go` | `backend/internal/tracker/handler_wave95_test.go` | tracker |
| 482 | `backend/testframework/tracker/handler_workflow_labels_test.go` | `backend/internal/tracker/handler_workflow_labels_test.go` | tracker |
| 483 | `backend/testframework/tracker/log_time_a3005_test.go` | `backend/internal/tracker/log_time_a3005_test.go` | tracker |
| 484 | `backend/testframework/tracker/provisioner_test.go` | `backend/internal/tracker/provisioner_test.go` | tracker |
| 485 | `backend/testframework/tracker/repository_extended_test.go` | `backend/internal/tracker/repository_extended_test.go` | tracker |
| 486 | `backend/testframework/tracker/repository_more_test.go` | `backend/internal/tracker/repository_more_test.go` | tracker |
| 487 | `backend/testframework/tracker/repository_test.go` | `backend/internal/tracker/repository_test.go` | tracker |
| 488 | `backend/testframework/tracker/repository_wave100_test.go` | `backend/internal/tracker/repository_wave100_test.go` | tracker |
| 489 | `backend/testframework/tracker/repository_wave120_test.go` | `backend/internal/tracker/repository_wave120_test.go` | tracker |
| 490 | `backend/testframework/tracker/repository_wave124_test.go` | `backend/internal/tracker/repository_wave124_test.go` | tracker |
| 491 | `backend/testframework/tracker/service_converters_extended_test.go` | `backend/internal/tracker/service_converters_extended_test.go` | tracker |
| 492 | `backend/testframework/tracker/service_converters_test.go` | `backend/internal/tracker/service_converters_test.go` | tracker |
| 493 | `backend/testframework/tracker/service_helpers_test.go` | `backend/internal/tracker/service_helpers_test.go` | tracker |
| 494 | `backend/testframework/tracker/service_issue_message_link_test.go` | `backend/internal/tracker/service_issue_message_link_test.go` | tracker |
| 495 | `backend/testframework/tracker/service_test.go` | `backend/internal/tracker/service_test.go` | tracker |
| 496 | `backend/testframework/tracker/service_wave100_test.go` | `backend/internal/tracker/service_wave100_test.go` | tracker |
| 497 | `backend/testframework/tracker/service_wave109_test.go` | `backend/internal/tracker/service_wave109_test.go` | tracker |
| 498 | `backend/testframework/tracker/service_wave112_test.go` | `backend/internal/tracker/service_wave112_test.go` | tracker |
| 499 | `backend/testframework/tracker/service_wave113_test.go` | `backend/internal/tracker/service_wave113_test.go` | tracker |
| 500 | `backend/testframework/tracker/service_wave114_test.go` | `backend/internal/tracker/service_wave114_test.go` | tracker |
| 501 | `backend/testframework/tracker/service_wave115_test.go` | `backend/internal/tracker/service_wave115_test.go` | tracker |
| 502 | `backend/testframework/tracker/service_wave117_test.go` | `backend/internal/tracker/service_wave117_test.go` | tracker |
| 503 | `backend/testframework/tracker/service_wave118_test.go` | `backend/internal/tracker/service_wave118_test.go` | tracker |
| 504 | `backend/testframework/tracker/service_wave119_test.go` | `backend/internal/tracker/service_wave119_test.go` | tracker |
| 505 | `backend/testframework/tracker/service_wave121_test.go` | `backend/internal/tracker/service_wave121_test.go` | tracker |
| 506 | `backend/testframework/tracker/service_workflow_extra_test.go` | `backend/internal/tracker/service_workflow_extra_test.go` | tracker |
| 507 | `backend/testframework/ws/authorizer_test.go` | `backend/internal/ws/authorizer_test.go` | ws |
| 508 | `backend/testframework/ws/bridge_lifecycle_test.go` | `backend/internal/ws/bridge_lifecycle_test.go` | ws |
| 509 | `backend/testframework/ws/bridge_test.go` | `backend/internal/ws/bridge_test.go` | ws |
| 510 | `backend/testframework/ws/handler_auth_test.go` | `backend/internal/ws/handler_auth_test.go` | ws |
| 511 | `backend/testframework/ws/hub_shutdown_test.go` | `backend/internal/ws/hub_shutdown_test.go` | ws |
| 512 | `backend/testframework/ws/hub_test.go` | `backend/internal/ws/hub_test.go` | ws |
| 513 | `backend/testframework/ws/origin_test.go` | `backend/internal/ws/origin_test.go` | ws |
| 514 | `backend/testframework/ws/protocol_test.go` | `backend/internal/ws/protocol_test.go` | ws |
| 515 | `backend/testframework/ws/pump_test.go` | `backend/internal/ws/pump_test.go` | ws |

---

## Следующий шаг

- **E2:** анализ `export_hooks.go` — [`TF-MIGRATE-E2-EXPORT-HOOKS.md`](TF-MIGRATE-E2-EXPORT-HOOKS.md)
- **Не делать:** физический `git mv` / создание `/tests` до A1 coverage baseline gate
- **Э3 (код):** раскладка по доменам волнами (сначала мелкие: calendar/status/mail → потом tms/tracker)

---

*Документ: planning only. Production code не изменялся.*
