# W3 Assignments — messenger + ws + calls

> Стадия A · read-only · TASK-008 · **131 файл** · 5 агентов

## A1 — messenger HTTP (26)
Отчёт: `A3-1-messenger-handlers.md`

### Prod (8)
backend/internal/messenger/handler.go
backend/internal/messenger/handler_channels.go
backend/internal/messenger/handler_messages.go
backend/internal/messenger/handler_presence.go
backend/internal/messenger/handler_reactions.go
backend/internal/messenger/handler_search.go
backend/internal/messenger/dto.go
backend/internal/messenger/events.go

### Tests (18)
backend/testframework/messenger/handler_channels_extra_test.go
backend/testframework/messenger/handler_flow_test.go
backend/testframework/messenger/handler_helpers_test.go
backend/testframework/messenger/handler_parse_test.go
backend/testframework/messenger/handler_wave100_test.go
backend/testframework/messenger/handler_wave106_test.go
backend/testframework/messenger/handler_wave109_test.go
backend/testframework/messenger/handler_wave112_test.go
backend/testframework/messenger/handler_wave113_test.go
backend/testframework/messenger/handler_wave114_test.go
backend/testframework/messenger/handler_wave115_test.go
backend/testframework/messenger/handler_wave116_test.go
backend/testframework/messenger/handler_wave123_test.go
backend/testframework/messenger/handler_wave93_test.go
backend/testframework/messenger/dto_test.go
backend/testframework/messenger/validation_test.go
backend/testframework/messenger/publish_typing_a2009_test.go
backend/testframework/messenger/stubs_test.go

## A2 — messenger services (28)
Отчёт: `A3-2-messenger-services.md`

### Prod (14)
backend/internal/messenger/service.go
backend/internal/messenger/service_bookmarks.go
backend/internal/messenger/service_bot.go
backend/internal/messenger/service_channel_invites.go
backend/internal/messenger/service_channels.go
backend/internal/messenger/service_dm.go
backend/internal/messenger/service_hook.go
backend/internal/messenger/service_messages.go
backend/internal/messenger/service_pins.go
backend/internal/messenger/service_polls.go
backend/internal/messenger/service_promote.go
backend/internal/messenger/service_reactions.go
backend/internal/messenger/service_search.go
backend/internal/messenger/auto_channels.go

### Tests (14)
backend/testframework/messenger/auto_channels_service_test.go
backend/testframework/messenger/auto_channels_test.go
backend/testframework/messenger/service_bot_test.go
backend/testframework/messenger/service_extended_test.go
backend/testframework/messenger/service_fanout_test.go
backend/testframework/messenger/service_mentions_test.go
backend/testframework/messenger/service_messages_test.go
backend/testframework/messenger/service_search_a2001_test.go
backend/testframework/messenger/service_test.go
backend/testframework/messenger/service_wave114_test.go
backend/testframework/messenger/service_wave115_test.go
backend/testframework/messenger/service_wave116_test.go
backend/testframework/messenger/service_wave123_test.go
backend/testframework/messenger/delete_channel_cascade_a8005_test.go

## A3 — messenger repos + presence (20)
Отчёт: `A3-3-messenger-repos.md`

### Prod (9)
backend/internal/messenger/repository.go
backend/internal/messenger/ent_repository.go
backend/internal/messenger/mongo_repository.go
backend/internal/messenger/combined_repository.go
backend/internal/messenger/mock_repository.go
backend/internal/messenger/presence.go
backend/internal/messenger/unread.go
backend/internal/messenger/export_hooks.go
backend/internal/messenger/hookfixture.go

### Tests (11)
backend/testframework/messenger/ent_repository_invites_test.go
backend/testframework/messenger/ent_repository_poll_test.go
backend/testframework/messenger/ent_repository_test.go
backend/testframework/messenger/ent_repository_wave123_test.go
backend/testframework/messenger/mongo_repository_test.go
backend/testframework/messenger/mongo_repository_wave123_test.go
backend/testframework/messenger/mongo_search_empty_scope_a2001_test.go
backend/testframework/messenger/mock_repository_test.go
backend/testframework/messenger/presence_test.go
backend/testframework/messenger/presence_wave100_test.go
backend/testframework/messenger/unread_test.go

## A4 — ws полный (15)
Отчёт: `A3-4-ws.md`

### Prod (6)
backend/internal/ws/authorizer.go
backend/internal/ws/bridge.go
backend/internal/ws/export_hooks.go
backend/internal/ws/handler.go
backend/internal/ws/hub.go
backend/internal/ws/protocol.go

### Tests (9)
backend/testframework/ws/authorizer_test.go
backend/testframework/ws/bridge_lifecycle_test.go
backend/testframework/ws/bridge_test.go
backend/testframework/ws/handler_auth_test.go
backend/testframework/ws/hub_shutdown_test.go
backend/testframework/ws/hub_test.go
backend/testframework/ws/origin_test.go
backend/testframework/ws/protocol_test.go
backend/testframework/ws/pump_test.go

## A5 — calls полный (42)
Отчёт: `A3-5-calls.md`

### Prod (16)
backend/internal/calls/chat.go
backend/internal/calls/chat_mongo.go
backend/internal/calls/dto.go
backend/internal/calls/export_hooks.go
backend/internal/calls/handler.go
backend/internal/calls/handler_calls.go
backend/internal/calls/handler_recording.go
backend/internal/calls/handler_scheduled.go
backend/internal/calls/models.go
backend/internal/calls/recording.go
backend/internal/calls/repository.go
backend/internal/calls/scheduled.go
backend/internal/calls/service.go
backend/internal/calls/service_chat.go
backend/internal/calls/service_events.go
backend/internal/calls/service_room.go

### Tests (26)
backend/testframework/calls/chat_authz_a2003_test.go
backend/testframework/calls/chat_mem_test.go
backend/testframework/calls/chat_mongo_wave103_test.go
backend/testframework/calls/chat_test.go
backend/testframework/calls/handler_flow_test.go
backend/testframework/calls/handler_rbac_recording_test.go
backend/testframework/calls/handler_recording_test.go
backend/testframework/calls/handler_routes_test.go
backend/testframework/calls/handler_scheduled_coverage_test.go
backend/testframework/calls/handler_validation_test.go
backend/testframework/calls/raise_hand_a2006_test.go
backend/testframework/calls/recording_egress_mock_test.go
backend/testframework/calls/recording_service_coverage_test.go
backend/testframework/calls/recording_test.go
backend/testframework/calls/recording_webhook_test.go
backend/testframework/calls/repository_series_trash_test.go
backend/testframework/calls/repository_test.go
backend/testframework/calls/service_chat_flow_test.go
backend/testframework/calls/service_extra_test.go
backend/testframework/calls/service_flow_test.go
backend/testframework/calls/service_recording_scheduled_test.go
backend/testframework/calls/service_room_test.go
backend/testframework/calls/service_test.go
backend/testframework/calls/webhook_events_test.go
backend/testframework/calls/ws_authorizer_integration_a2007_test.go
backend/testframework/calls/ws_call_authz_a2007_test.go

## Контроль
| A1 | A2 | A3 | A4 | A5 | Σ |
|---:|---:|---:|---:|---:|---:|
| 26 | 28 | 20 | 15 | 42 | **131** |
