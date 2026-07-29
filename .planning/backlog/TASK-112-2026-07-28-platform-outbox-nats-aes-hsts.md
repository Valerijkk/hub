# TASK-112 — 🐛 Platform: NATS/HSTS/AES-AAD/RL/webhook tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке  
> Источник: A2W4 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

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