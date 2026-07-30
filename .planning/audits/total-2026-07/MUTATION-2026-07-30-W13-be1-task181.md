# MUTATION W13 slot 1 refill — TASK-181 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-181-2026-07-29-c-w19-webhook-softdelete-deactivate.md` `[MUTATION CRITERIA]`  
> Finding alias: audit `C-W19-004` ≡ task `C-W19-FE-004` · Коммит не создавался.

## Verdict

**2/2 PROVEN** · TASK-181 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| C-W19-FE-004-001 | **PROVEN** | `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181`; `TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181` |
| C-W19-FE-004-002 | **PROVEN** | `TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181`; `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### C-W19-FE-004-001 — SoftDelete tx deactivate endpoints

```text
BASELINE DeactivatesWebhookEndpoints + SourceGuard -> PASS (~1.0s)

BREAK: project/repository.go SoftDeleteWithCrossLinkCleanup —
  убрать SetIsActive(false) из WebhookEndpoint.Update (Update().Where().Save без Set)
-> FAIL DeactivatesWebhookEndpoints: IsActive Should be false
-> FAIL SourceGuard: Regexp SetIsActive(false) missing
RESTORE -> PASS (~0.9s)
```

Note: полный drop `WebhookEndpoint.Update` оставлял unused import `webhookendpoint` → build fail (не test RED). Kill = remove `SetIsActive(false)` call — компилируется, named tests RED.

### C-W19-FE-004-002 — ListWebhookEndpointsByProject trash filter

```text
BASELINE (после restore 001) empty list path covered in suite GREEN

BREAK: apigateway/repository.go ListWebhookEndpointsByProject —
  убрать project.DeletedAtIsNil() из HasProjectWith
-> FAIL SoftDeletedProject_Empty: Should be empty (list непуст)
-> FAIL DeactivatesWebhookEndpoints: list-by-project trash not empty
RESTORE -> PASS (~1.0s)
```

Note: SourceGuard string-Contains может остаться GREEN, если comment ещё держит токен `project.DeletedAtIsNil()` — runtime Empty asserts = канон RED для PROVEN.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/apigateway/ -run `
  "TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181|TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181|TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181"
# FINAL GREEN PASS ~1.0s
```

## Gaps

- Docs-only: **N/A**
- TASK-152 lock `ListActiveWebhookEndpoints` `HasProjectWith(DeletedAtIsNil)` — не ослабляли
- Pre-existing dirty `project/repository.go` vs HEAD (`afterCrossLinkDeleteHook` block) — не часть этого mutate; webhook guard restored; no `MUTATED` marker

## Handoff

- Evidence: этот файл · TASK-181 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `apigateway/repository.go` diff empty vs start; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
