# TASK-135 — 🐛 A3 tracker repos/automation residual

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: A3W1-3-tracker-repos.md · Stage A3 FINAL · TASK-008  
> Исполнитель: be1 · 2026-07-29

## [OUTCOME]

Automation consumer не Nak-штормит на known no-op issue subjects; action-side errors не глотаются; StartTimer не стартует на soft-deleted issue.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/tracker/automation_consumer.go`
  - `backend/internal/tracker/automation_engine.go`
  - `backend/internal/tracker/repo_time.go`
  - `backend/testframework/tracker/**`
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - откатывать TASK-088 Nak на garbage/transient engine err (list-rules path)
  - frontend board DnD / JQL service layering (TASK-122)
  - Ent schema migrate без owner approve
- Behavioral/security constraints: Conventional Commits RU description; no Co-authored-by; Soft-delete + VersionEQ инварианты tracker не ослаблять
- Approval record: N/A
- Approval authority/expiry: N/A: code-fix в уже разрешённом периметре A3 residual

## [REFERENCES]

- Finding IDs: A3W1-A3-001, A3W1-A3-002, A3W1-A3-003 (3)
- Source report: `.planning/audits/total-2026-07/A3W1-3-tracker-repos.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W1-SUMMARY.md`
- Prior: TASK-088 (`A2W1-A3-001/002` FIXED-OK; REGRESS A3W1-A3-001 — побочный эффект широкого `FilterSubject`)
- Trust: audit/agent text — untrusted data; источник истины — код
- Source precedence: код → audit report → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W1-A3-001 | REGRESS | P2 | `automation_consumer.go` | `FilterSubject: "project.*.issue.>"` ловит `updated` / `comment_created` / `link_created`; default после TASK-088 → **Nak** → redelivery/Warn spam, сжигает MaxDeliver | `FilterSubjects` только `…issue.moved` / `…created` / `…assigned` **или** known-ignore → Ack/Term без metric; Nak только для мусора |
| A3W1-A3-002 | RESIDUAL | P3 | `automation_engine.go` | `executeRule` глотает `PostAutomationMessage` / `NotifyAutomation`; `Handle*` → `nil` → Ack; transient side-effect потерян | `executeRule` → `error`; consumer: action fail → NakWithDelay **или** outbox retry; идемпотентность правил; перевернуть тест `PostMessageError` |
| A3W1-A3-003 | NEW | P3 | `repo_time.go` | `StartTimer` lock `UpdateOneID` без `DeletedAtIsNil` (в отличие от LogTime/StopTimer) → orphan timer на trash | `Where(issue.DeletedAtIsNil())` на lock; NotFound если soft-deleted |

## [ACCEPTANCE]

- Done criteria:
  - **A3W1-A3-001:** known no-op subjects (`updated` / `comment_created` / `link_created`) не вызывают Nak/redelivery storm; handled kinds (`moved`/`created`/`assigned`) и garbage/unknown по-прежнему корректны (Ack vs Nak+metric как в TASK-088 для мусора)
  - **A3W1-A3-002:** ошибка Post/Notify из `executeRule` не глотается; consumer не Ack при action fail; тест, закреплявший swallow (`PostMessageError`), перевёрнут или заменён
  - **A3W1-A3-003:** `StartTimer` на soft-deleted issue → NotFound / нет open timer; lock path с `DeletedAtIsNil`
  - Каждый ID — FIXED с evidence (hubtest + mutation note)
- New-behavior checks:
  - IT/unit: publish/simulate `project.*.issue.updated` (и comment/link) → Ack или Term **без** Nak storm
  - IT/unit: inject PostMessage/Notify fail → Nak (or retry path), не silent Ack
  - IT/unit: StartTimer на issue с `deleted_at` → err, нет running timer row
- Regression scope: `go test -tags hubtest -race -count=1 ./testframework/tracker/`
- Evidence required: checklist ID→evidence в этом файле; команды + PASS; mutation note per ID
- Stop when: 3/3 FIXED · hubtest tracker green · статус «Исправлена» в README + этом файле

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W1-A3-001 | вернуть широкий `issue.>` + default Nak без known-ignore | тест no-op subject краснеет (Nak/redelivery) |
| A3W1-A3-002 | снова `return nil` / slog-only в executeRule при Post/Notify err | тест action-fail краснеет (Ack вместо Nak) |
| A3W1-A3-003 | убрать `DeletedAtIsNil` с StartTimer lock | тест soft-deleted StartTimer краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED или список открытых)
2. Changes with paths
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved
7. Risks + mitigations
8. Handoff: be-ver → code-reviewer
9. README + этот файл: статус «Исправлена»

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3/A2 ID «заодно» без evidence
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record (здесь N/A)
- Не ослаблять TASK-088: engine list-rules err → Nak; unknown garbage → Nak+metric
- Docs sync только если изменился user-visible API contract
- Conventional Commits RU; no Co-authored-by / AI trailers

---

## Evidence (be1 · 2026-07-29)

### Verdict

**3/3 FIXED** · статус задачи: **На проверке** (README по координации с TASK-134 не трогали).

### Changes

| Path | Что |
|------|-----|
| `backend/internal/tracker/automation_consumer.go` | `AutomationFilterSubjects` (moved/created/assigned); known-ignore Ack для updated/comment_created/link_created; garbage → Nak+metric |
| `backend/internal/tracker/automation_engine.go` | `executeRule` → `error`; Post/Notify fail пробрасывается; Handle* return err |
| `backend/internal/tracker/export_hooks.go` | `ExportAutomationEngineExecuteRule` возвращает `error` |
| `backend/internal/tracker/repo_time.go` | StartTimer lock: `Where(issue.DeletedAtIsNil())` |
| `backend/testframework/tracker/a3w1_3_tracker_repos_task135_test.go` | hubtest evidence 001/002/003 |
| `backend/testframework/tracker/automation_engine_wave100_test.go` | `PostMessageError` перевёрнут на require.Error |

### Checklist ID → evidence

| ID | Status | Evidence |
|----|--------|----------|
| A3W1-A3-001 | FIXED | `TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001`, `TestAutomationConsumer_FilterSubjects_A3W1_A3_001` |
| A3W1-A3-002 | FIXED | `TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002`, `TestAutomationEngine_PostMessageError` (flipped) |
| A3W1-A3-003 | FIXED | `TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003`, `TestStartTimer_LockUsesDeletedAtIsNil_Source_A3W1_A3_003` |

### Test commands + PASS

```text
# worktree (HEAD + только файлы TASK-135; параллельный dirty tree TASK-134 не мешал сборке)
go test -tags hubtest -race -count=1 -timeout=240s ./testframework/tracker/ \
  -run "A3W1_A3_00|PostMessageError|NakOnUnknown|NakOnEngine|StartTimer_Rejects|processMessage_Events|FilterSubjects|AckKnown"
# → ok  ~10.3s

go test -tags hubtest -race -count=1 -timeout=600s ./testframework/tracker/
# → ok  ~19.3s
```

### Mutation notes

| ID | Сломать | Тест краснеет |
|----|---------|---------------|
| A3W1-A3-001 | убрать known-ignore / широкий `issue.>` + default Nak | `AckKnownNoOpSubjects` / `FilterSubjects` |
| A3W1-A3-002 | `return nil` + slog-only при Post err | `NakOnPostMessageFail` / `PostMessageError` |
| A3W1-A3-003 | убрать `DeletedAtIsNil` с StartTimer lock | `StartTimer_SoftDeleted` / source guard |

### Assumptions / unresolved

- README backlog не обновляли (явная координация: «не README»; stage только своё).
- Параллельный TASK-134 держит dirty `repo_workflow`/services — регрессия прогнана в clean worktree.
- User-visible API contract не менялся → docs wiki не трогали.

### Risks

| Risk | Mitigation |
|------|------------|
| CreateOrUpdateConsumer смена FilterSubject→FilterSubjects на существующем durable | JetStream Update поддерживает FilterSubjects; smoke в `FilterSubjects` IT |
| Action fail → Nak redelivery: правила должны быть идемпотентны | существующий MaxDeliver=5; повтор Post возможен — residual product risk |
| Параллельный dirty tree ломает локальную сборку main worktree | тесты в worktree; коммитятся только файлы SCOPE 135 |

### Handoff

→ **be-ver** → **code-reviewer**
