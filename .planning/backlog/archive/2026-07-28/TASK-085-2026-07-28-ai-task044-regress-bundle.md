# TASK-085 — ♻️ AI: закрыть false-close TASK-044 (RPM/AssistJQL/stream)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W4
> Источник: A2W5 REGRESS · Stage A2 dedup · TASK-008

## [OUTCOME]

AI rate-limit и stream-пути fail-closed по контракту TASK-040/044: Redis error не пропускает RPM; AssistJQL мапит quota/RPM; stream без 90s whole-request kill после первого delta.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/ai/**
  - backend/cmd/server/wiring_audit_ai.go
  - backend/testframework/ai/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - Ollama SafeDial ADR U01
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W5 REGRESS · provider_rpm.go:47-48 return nil
- handler AssistJQL vs SuggestComposer
- TASK-044 / TASK-040
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - RPM: Redis err → rate-limit/503, не Allow
  - AssistJQL: sentinel map как SuggestComposer; AIRateLimit gaps
  - stream timeout/fallback после delta; code_review nil deny
  - Мутация: return nil на Redis err → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился

## [REWORK 2026-07-29 be1]

**Исправлена.** Evidence (worktree `fix/ai/task-085-regress-bundle` ← `cursor-grok-4.5`):

- RPM: Redis err → `ErrProviderRPMUnavailable` (не Allow); ZRem on reject; unique ZSET member
- AssistJQL: sentinel map как SuggestComposer; AIRateLimit на admin + `/ai/my`
- Stream: SSE client без `Client.Timeout`; no provider switch после первого content-delta
- code_review: `access==nil` → deny
- Тесты: `backend/testframework/ai/task085_regress_test.go` + правки unified/scope
- Mutation: `return nil` на Redis err → `TestRedisProviderRPMLimiter_FailClosedOnRedisErr` RED
- Регрессия: `go test -tags hubtest -race -count=1 ./testframework/ai/` PASS (~15.7s)
