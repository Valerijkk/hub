# MUTATION W5 slot 1 — TASK-135 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost:5432 / :27017 · Criteria: `.planning/backlog/TASK-135-2026-07-29-a3-1-3-tracker-repos.md` `[MUTATION CRITERIA]`
> Коммит не создавался (инструкция слота).

## Verdict

**4/4 PROVEN** · статус TASK-135 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W1-A3-001a | **PROVEN** | `TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001` | убрать ветку `knownNoOpIssueEventKinds` → всегда unknown/Nak | consumer |
| A3W1-A3-001b | **PROVEN** | `TestAutomationConsumer_FilterSubjects_A3W1_A3_001` | `AutomationFilterSubjects` → `project.*.issue.>` + `FilterSubject` в Start | consumer |
| A3W1-A3-002 | **PROVEN** | `TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002` | Post/Notify err → `return nil` (swallow) | engine→consumer |
| A3W1-A3-003 | **PROVEN** | `TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003` (+ source) | убрать `Where(issue.DeletedAtIsNil())` с StartTimer lock | repository |

Production restored (`git checkout --` после каждого BREAK). Mock-gap: нет (001a unit processMessage; 001b JetStream FilterSubjects; 002 Ent rules + stubMessenger; 003 EntRepository).

## Per-ID evidence

### A3W1-A3-001a — knownNoOp Ack

```text
BREAK: automation_consumer.go processMessage default — drop knownNoOp branch
→ FAIL TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001
   (updated/comment_created/link_created: Ack=0 Nak=1; unknown metric bumped)
RESTORE
→ PASS same test
```

### A3W1-A3-001b — FilterSubjects allowlist

```text
BREAK: AutomationFilterSubjects = ["project.*.issue.>"] + Start FilterSubject
→ FAIL TestAutomationConsumer_FilterSubjects_A3W1_A3_001
   (var + JetStream ConsumerConfig.FilterSubjects mismatch)
RESTORE
→ PASS same test
```

### A3W1-A3-002 — Post/Notify propagate

```text
BREAK: automation_engine.go executeRule — Post/Notify err → return nil
→ FAIL TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002
   (expected Nak=1 Ack=0; got silent Ack)
RESTORE
→ PASS same test
```

### A3W1-A3-003 — StartTimer soft-delete lock

```text
BREAK: repo_time.go StartTimer — drop Where(issue.DeletedAtIsNil()) on lock
→ FAIL TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003 (err==nil, orphan timer)
→ FAIL TestStartTimer_LockUsesDeletedAtIsNil_Source_A3W1_A3_003 (chunk missing DeletedAtIsNil)
RESTORE
→ PASS both
```

Note: source-guard матчит substring `DeletedAtIsNil` в chunk StartTimer — комментарий с этим токеном маскирует guard; Ent SoftDeleted остаётся primary proof.

## Commands

```text
cd backend
# baseline subset
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/tracker/ \
  -run "TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001|TestAutomationConsumer_FilterSubjects_A3W1_A3_001|TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002|TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003|TestStartTimer_LockUsesDeletedAtIsNil_Source_A3W1_A3_003"
→ ok  ~3.4s

# after each restore + final subset
→ ok  ~3.1–3.5s per run
```

## Assumptions / unresolved

- Коммит не создавался.
- `layer_mismatch IDs: (none)` — подтверждено.

## Handoff

TASK-135 **Закрыта** · evidence этот файл · README синхронизирован.
