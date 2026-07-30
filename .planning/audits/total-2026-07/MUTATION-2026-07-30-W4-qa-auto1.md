# MUTATION W4 slot 3 — TASK-152 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** (!= be2 fixer) · Env: localhost · Evidence: `_mut-batch-w4-task152.json` (+ 001/002 first-run log)
> Criteria: `.planning/backlog/TASK-152-2026-07-29-a3-4-5-apigateway.md` `[MUTATION CRITERIA]`

## Verdict

**4/4 runnable PROVEN** · A3W4-5-003 FilterSubjects **PROVEN** (есть тест) · статус TASK-152 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W4-5-001 | **PROVEN** | `TestListActiveWebhookEndpoints_SoftDeletedProject_Excluded` | убрать `HasProjectWith(project.DeletedAtIsNil())` из `ListActiveWebhookEndpoints` | repository |
| A3W4-5-002 | **PROVEN** | `TestUpdateWebhookEndpoint_VersionEQ_ConflictAndSuccess` | убрать `VersionEQ(input.Version)` из `UpdateWebhookEndpoint` | repository |
| A3W4-5-003 | **PROVEN** (allowlist) | `TestValidateWebhookEvents_SprintCompleted_Rejected` | вернуть `"sprint.completed": {}` в `allowedWebhookEvents` | service |
| A3W4-5-003 | **PROVEN** (FilterSubjects) | `TestWebhookFilterSubjects_CoversWebhookEvents` | вернуть `"sprint.>"` в `WebhookFilterSubjects` | nats |
| A3W4-5-004 | **PROVEN** | `TestCreateAPIKey_NameMaxLen_400` | `> 255` → `> 999999` в CreateAPIKey | handler |

Production restored (`git checkout --` после каждого BREAK). Mock-gap: нет (001/002 — Ent sqlite repo; 003 — allowlist + FilterSubjects snapshot; 004 — Fiber handler).

## Per-ID evidence

### A3W4-5-001 — soft-delete project filter

```text
BREAK: repository.go — drop HasProjectWith(project.DeletedAtIsNil())
→ FAIL TestListActiveWebhookEndpoints_SoftDeletedProject_Excluded  (red_ok=true)
RESTORE
→ PASS same test  (green_ok=true)
```

### A3W4-5-002 — VersionEQ optimistic lock

```text
BREAK: repository.go — drop VersionEQ(input.Version)
→ FAIL TestUpdateWebhookEndpoint_VersionEQ_ConflictAndSuccess  (stale last-write-wins)
RESTORE
→ PASS same test
```

### A3W4-5-003 — trim path (allowlist + FilterSubjects)

Путь фикса: **trim** (не publish). Критерий MUTATION CRITERIA — allowlist; FilterSubjects — отдельный gap-check.

```text
BREAK allowlist: webhook.go — add "sprint.completed"
→ FAIL TestValidateWebhookEvents_SprintCompleted_Rejected
RESTORE → PASS

BREAK FilterSubjects: platform/nats/client.go — add "sprint.>"
→ FAIL TestWebhookFilterSubjects_CoversWebhookEvents
RESTORE → PASS
```

FilterSubjects **не N/A**: тест `TestWebhookFilterSubjects_CoversWebhookEvents` ловит мёртвый `sprint.>` — Break→RED→GREEN PROVEN.

### A3W4-5-004 — API key name MaxLen → 400

```text
BREAK: handler.go — RuneCount > 999999
→ FAIL TestCreateAPIKey_NameMaxLen_400  (ожидали 400)
RESTORE
→ PASS same test
```

## Commands

```text
cd backend
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/apigateway/
→ ok  ~6.4s

# related subset after restore
go test -tags hubtest -race -count=1 -run '^Test(ListActiveWebhookEndpoints_SoftDeleted|UpdateWebhookEndpoint_VersionEQ_Conflict|ValidateWebhookEvents_Sprint|CreateAPIKey_NameMaxLen|WebhookFilterSubjects_Covers)' ./testframework/apigateway/
→ ok  ~2.9s
```

## Assumptions / unresolved

- FE `webhooks.tsx` ещё может показывать `sprint.completed` / PUT без `version` — OOS (be2 Progress); mutation не трогала FE.
- Коммит не создавался (инструкция слота).

## Handoff

TASK-152 **Закрыта** · evidence этот файл · README синхронизирован.
