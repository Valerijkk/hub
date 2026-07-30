# TASK-112 — 🐛 Platform: NATS/HSTS/AES-AAD/RL/webhook tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта  
> Источник: A2W4 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W10-be2.md` (7/9 PROVEN, 2 N/A docs)

## [OUTCOME]

Остаток W4 вне 081/086/087 закрыт: PubAck, HSTS docs/Caddy, AES AAD, RL race, webhook test path.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/platform/**
  - apigateway tests
  - deployments/caddy*
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - секреты в git
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W4-1-001..003, A2W4-2-002/003, A2W4-3-001, A2W4-4-001, A2W4-5-004/005 (9)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ID checklist + тесты/docs
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

> Сверено по HEAD 2026-07-30 · be2 W7 slot 2 · статус задачи не менялся (**На проверке**)

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W4-1-001 | `backend/internal/platform/middleware/security_headers.go:12-18` — добавить `c.Set("Strict-Transport-Security", "max-age=…")` **или** убрать HSTS из `deployments/Caddyfile.prod:5-6` | `TestSecurityHeaders_SetsExpectedHeaders` (Fiber Empty HSTS); Caddy — docs/config diff (не hubtest) | middleware (+ Caddy edge) |
| A2W4-1-002 | `backend/internal/platform/http/helpers.go:46-48` — убрать `if id == uuid.Nil` reject | `TestParseUUIDParam_RejectsNilUUID` | http helpers |
| A2W4-1-003 | `ratelimit.go:255-265` / `redisSlidingWindowLua` — ZADD **до** card check (или pipeline ZADD→ZCARD→ZRem без Lua) | `TestRateLimitRedis_LuaRejectDoesNotAddMember` | middleware Redis Lua |
| A2W4-2-002 | `backend/internal/platform/outbox/worker.go:261` — заменить `PublishMsgAck` на core `PublishMsg`/`Publish` (без PubAck) | `TestProcessEvent_PubAckFailure_StaysPending_A2W4_2_002` | outbox worker |
| A2W4-2-003 | `backend/internal/platform/nats/client.go:21-32` — выкинуть subject из `StreamSubjects` (напр. `notification.>` / `audit.>`) | `TestStreamSubjects` | nats client |
| A2W4-3-001 | N/A runtime — assignment drift: вернуть удалённый `metrics_allow_middleware_test.go` в `W4-ASSIGNMENTS.md` A3 / убрать `metrics_route_deny_task059_test.go` | N/A code test; evidence = assignment sync | docs/assignment |
| A2W4-4-001 | `backend/internal/platform/crypto/aesgcm.go:125-138` / `:159-176` — `EncryptWithAAD`/`DecryptWithAAD` игнор `aad` (всегда `nil` в Seal/Open) | `TestEncryptDecryptWithAAD_RoundtripAndMismatch` | crypto |
| A2W4-5-004 | N/A runtime — откатить уточнение CTC inbound vs outbound в `.cursor/rules/60-tz-constraints.mdc` (строка Webhook signature) | N/A code test; evidence = docs diff | docs/TZ |
| A2W4-5-005 | `backend/testframework/apigateway/webhook_test.go:199-201` — заменить `ExportWebhookDispatcherDeliverWithRetry` на 1× прямой `sendWebhook` (без retry loop) **или** prod `deliverWithRetry` без повторных попыток на 5xx | `TestSendWebhook_5xxRetries` | apigateway (test→prod deliver path) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001 | Fiber middleware + Caddyfile | Guard = «HSTS не в Fiber»; edge — config |
| 002–003 | unit helpers / miniredis Lua | OK |
| 002 outbox | Ent outbox + JS PubAck | Ent/NATS OK |
| 003 nats | unit slice assert | OK |
| 4-001 | crypto unit | OK |
| 3-001 / 5-004 | docs | нет runtime RED |
| 5-005 | stub repo + DeliverWithRetry | intentional dispatcher path |

`layer_mismatch IDs: (none)`

### Gaps

- **A2W4-3-001 / A2W4-5-004** — только docs; нельзя PROVEN runtime-мутацией.
- **A2W4-1-001** Caddy HSTS — нет automated assert на `Caddyfile.prod` (только Fiber Empty + ручной/diff config).
- TASK-112 без отдельного Progress-checklist в README-смысле уже FIXED в Evidence; verifier W8 должен Break→RED независимо от be1.

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record

## [EVIDENCE] 2026-07-28 be1

| ID | Status | Evidence |
|----|--------|----------|
| A2W4-1-001 | FIXED | HSTS в `Caddyfile.prod`; Fiber без HSTS + тест |
| A2W4-1-002 | FIXED | `TestParseUUIDParam_RejectsNilUUID` |
| A2W4-1-003 | FIXED | Redis RL Lua atomic ZADD; `TestRateLimitRedis_LuaRejectDoesNotAddMember` |
| A2W4-2-002 | FIXED | outbox → `PublishMsgAck`; `TestProcessEvent_PubAckFailure_StaysPending` |
| A2W4-2-003 | FIXED | `StreamSubjects` канон + `TestStreamSubjects` (9) |
| A2W4-3-001 | FIXED | `W4-ASSIGNMENTS.md` A3 без deleted middleware test |
| A2W4-4-001 | FIXED | `EncryptWithAAD`/`DecryptWithAAD` + mismatch test |
| A2W4-5-004 | FIXED | `60-tz-constraints.mdc`: CTC inbound github/gitlab |
| A2W4-5-005 | FIXED | `TestSendWebhook_5xxRetries` → `DeliverWithRetry` |

Tests: `go test -tags hubtest -race -count=1 ./testframework/platform/{http,middleware,nats,crypto,outbox}/ ./testframework/apigateway/` → PASS.

## [MUTATION] 2026-07-30 be2 W10 slot 1

| ID | Verdict |
|----|---------|
| A2W4-1-001 | PROVEN |
| A2W4-1-002 | PROVEN |
| A2W4-1-003 | PROVEN |
| A2W4-2-002 | PROVEN |
| A2W4-2-003 | PROVEN |
| A2W4-3-001 | N/A (docs) |
| A2W4-4-001 | PROVEN |
| A2W4-5-004 | N/A (docs) |
| A2W4-5-005 | PROVEN |

Report: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W10-be2.md` · статус **Закрыта**.