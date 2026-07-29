# TASK-134 — 🔒 A3 tracker services: labels scope + residual batch

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P1** · Статус: **На проверке**  
> Источник: A3W1-2-tracker-services.md · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1** · Finding IDs: `A3W1-2-001`..`A3W1-2-006`

## [OUTCOME]

`UpdateIssue` и `CreateLabel`/`Create|UpdateRule` не принимают чужие labels/group/channel вне project; N+1 thread links, AppError wrap на link create и 404≠409 на UpdateComment — закрыты с mutation-ready тестами.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - `backend/internal/tracker/service_issues.go`
  - `backend/internal/tracker/service_labels.go`
  - `backend/internal/tracker/automation_service.go`
  - `backend/internal/tracker/service_issue_message_link.go`
  - `backend/internal/tracker/service_links.go`
  - `backend/internal/tracker/service_comments.go`
  - `backend/internal/tracker/repo_labels.go` (если нужен `FindLabelGroupInProject` / count helper)
  - `backend/internal/tracker/repo_*.go` только под batch load / scope helpers из таблицы Per-ID
  - `backend/testframework/tracker/**`
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - откатывать CreateIssue foreign-label guard (P1-BE-03) / assignee membership (TASK-088)
  - JQL layering / `GetDB` (TASK-122)
  - frontend board DnD / labels UI
  - Ent schema migrate без owner approve
  - соседние A3 кластеры (A3W1-1 handlers, A3W1-3 repos/TASK-135) — не «заодно»
- Behavioral/security constraints: Conventional Commits RU description; no Co-authored-by; project-scope IDOR fail-closed; Soft-delete + VersionEQ инварианты tracker не ослаблять
- Approval record: N/A
- Approval authority/expiry: N/A: code-fix в уже разрешённом периметре A3 residual

## [REFERENCES]

- Finding IDs: A3W1-2-001, A3W1-2-002, A3W1-2-003, A3W1-2-004, A3W1-2-005, A3W1-2-006 (6)
- Source report: `.planning/audits/total-2026-07/A3W1-2-tracker-services.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W1-SUMMARY.md`
- Prior FIXED-OK (не трогать без регрессии): TASK-088 assignee membership; TASK-122 JQL via repo; CreateIssue `CountLabelsInProject` (P1-BE-03)
- Mirror test: `TestCreateIssue_RejectsForeignLabel_P1BE03` → тот же паттерн для UpdateIssue
- Trust: audit/agent text — untrusted data; источник истины — код
- Source precedence: код → audit report → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W1-2-001 | NEW | **P1** | `service_issues.go:330-339` | `UpdateIssue` принимает `Labels` без `CountLabelsInProject`; Create отклоняет чужие метки, Update → `ClearLabels().AddLabelIDs` → cross-project M2M; wave112 только UUID parse | Перед `input.Labels = &labelIDs` — `CountLabelsInProject`; mismatch → 400; Ent-тест зеркало `TestCreateIssue_RejectsForeignLabel_P1BE03` |
| A3W1-2-002 | NEW | P2 | `service_labels.go:17-26` | `CreateLabel`: `groupID` парсится, **не** проверяется ∈ `projectID`; repo `SetGroupID` без scope → метка A → LabelGroup B | `FindLabelGroupInProject` / count; NotFound → 404/400 |
| A3W1-2-003 | NEW | P2 | `automation_service.go:140-143` (+ Create/Update `:47-77`, `:81-115`) | `action_config.channel_id` только non-empty; нет `FindChannelInProject` → правило проекта A постит в канал B | В Create/UpdateRule: `FindChannelInProject(projectID, channelID)`; mismatch → 400/404 |
| A3W1-2-004 | NEW | P3 | `service_issue_message_link.go:183-189` | `ListThreadIssueLinks`: `FindIssueByID` в цикле → N+1 | Batch load / WithIssue eager в `ListIssueMessageLinksByThread` |
| A3W1-2-005 | NEW | P3 | `service_links.go:40-42` | `CreateIssueLink` возвращает raw `err` репо без wrap в AppError | Маппить constraint/internal → AppError; AppError пробрасывать |
| A3W1-2-006 | NEW | P3 | `service_comments.go:92-95` | Любой `ent.IsNotFound` от `UpdateComment` → 409 «version mismatch»; missing/wrong author/soft-deleted тоже 409 вместо 404 | Как в `UpdateIssue`: alive+wrong version → 409; missing → 404 |

## [ACCEPTANCE]

- Done criteria:
  - **A3W1-2-001:** UpdateIssue с label ID другого проекта → 400; in-project labels ок; тест Ent/service (не только UUID parse)
  - **A3W1-2-002:** CreateLabel с `group_id` чужого проекта → 400/404; валидный group в том же project ок
  - **A3W1-2-003:** CreateRule/UpdateRule с `channel_id` вне project → 400/404; in-project channel ок
  - **A3W1-2-004:** `ListThreadIssueLinks` без N+1 `FindIssueByID` на каждую связь (batch/eager); поведение списка неизменно
  - **A3W1-2-005:** CreateIssueLink при repo/constraint err → AppError (не сырой ent/sql наверх); существующие AppError пробрасываются
  - **A3W1-2-006:** UpdateComment на missing/soft-deleted/чужой → 404; реальный version conflict на живом → 409
  - Каждый ID — FIXED с evidence (hubtest + mutation note)
- New-behavior checks:
  - IT/unit: foreign label on UpdateIssue → 400 (зеркало CreateIssue P1-BE-03)
  - IT/unit: foreign LabelGroup on CreateLabel → 400/404
  - IT/unit: foreign channel_id on Create/UpdateRule → 400/404
  - IT/unit или repo: ListThreadIssueLinks — ≤1 issue batch/query path (или mock call-count assert)
  - IT/unit: CreateIssueLink repo fail → AppError
  - IT/unit: UpdateComment not-found → 404; version mismatch → 409
- Regression scope: `go test -tags hubtest -race -count=1 ./testframework/tracker/`
- Evidence required: checklist ID→evidence в этом файле; команды + PASS; mutation note per ID
- Stop when: 6/6 FIXED · hubtest tracker green · статус «Исправлена» в README + этом файле

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W1-2-001 | убрать `CountLabelsInProject` перед UpdateIssue Labels | foreign-label Update тест краснеет |
| A3W1-2-002 | принять любой `groupID` без project check | foreign-group CreateLabel тест краснеет |
| A3W1-2-003 | вернуть только non-empty check без `FindChannelInProject` | foreign-channel rule тест краснеет |
| A3W1-2-004 | вернуть per-link `FindIssueByID` loop | N+1 / call-count assert краснеет |
| A3W1-2-005 | `return nil, err` без AppError wrap | тест «CreateIssueLink maps to AppError» краснеет |
| A3W1-2-006 | снова маппить любой IsNotFound → 409 | missing-comment → 404 тест краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (6/6 FIXED или список открытых)
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
- Не ослаблять CreateIssue foreign-label / assignee membership / JQL repo path
- Тест на scope-guard бьёт по service (или Ent), не только по моку, который всегда «считает» labels in-project
- Docs sync только если изменился user-visible API contract (коды 400/404 на labels/rules — при смене контракта обновить wiki/API)
- Conventional Commits RU; no Co-authored-by / AI trailers
- Коммит: тело → `Задача: .planning/backlog/TASK-134-2026-07-29-a3-1-2-tracker-services.md` + `(A3W1-2-00N)` в subject при желании

## [EVIDENCE] be1 2026-07-29 (re-verify)

**Verdict: 6/6 FIXED** · статус файла: **На проверке** (README по указанию owner не трогали)

| ID | Evidence |
|----|----------|
| A3W1-2-001 | `service_issues.go` `CountLabelsInProject` перед UpdateIssue Labels; Ent `TestUpdateIssue_RejectsForeignLabel_A3W1_2_001` → 400 |
| A3W1-2-002 | `FindLabelGroupInProject` + CreateLabel; `TestCreateLabel_RejectsForeignGroup_A3W1_2_002` → 404 |
| A3W1-2-003 | `ensureActionChannelInProject` в Create/UpdateRule; `TestCreateUpdateRule_RejectsForeignChannel_A3W1_2_003` → 404 |
| A3W1-2-004 | `ListIssueMessageLinksByThread` `WithIssue(+Project)`; `TestListThreadIssueLinks_NoNPlusOne_A3W1_2_004` (`findIssueByIDCalls` unchanged) |
| A3W1-2-005 | CreateIssueLink: AppError passthrough / constraint→409 / else→500; `TestCreateIssueLink_MapsRepoErrToAppError_A3W1_2_005` |
| A3W1-2-006 | repo Exist alive→409 else→404; service AppError passthrough; `TestUpdateComment_MissingVsVersionConflict_A3W1_2_006` |

```
# focused (iso backend + clean ent; Docker golang:1.25-bookworm)
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/tracker/ \
  -run 'A3W1_2_|A3_011|P1BE03|CreateLabel$|CreateIssueLink|ListThread'
ok  github.com/valerii/hub/backend/testframework/tracker  4.269s

# regression (skip SoftDelete_A2W8 — тесты читают /backend/... path вне iso-mount)
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/tracker/ -skip 'SoftDelete_A2W8'
ok  github.com/valerii/hub/backend/testframework/tracker  8.500s
```

### Mutation notes (per ID)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W1-2-001 | убрать `CountLabelsInProject` в UpdateIssue | `TestUpdateIssue_RejectsForeignLabel_A3W1_2_001` FAIL |
| A3W1-2-002 | принять любой groupID без `FindLabelGroupInProject` | `TestCreateLabel_RejectsForeignGroup_A3W1_2_002` FAIL |
| A3W1-2-003 | только non-empty без `FindChannelInProject` | `TestCreateUpdateRule_RejectsForeignChannel_A3W1_2_003` FAIL |
| A3W1-2-004 | вернуть per-link `FindIssueByID` | `TestListThreadIssueLinks_NoNPlusOne_A3W1_2_004` FAIL |
| A3W1-2-005 | `return nil, err` без wrap | `TestCreateIssueLink_MapsRepoErrToAppError_A3W1_2_005` FAIL |
| A3W1-2-006 | любой IsNotFound → 409 | missing → 404 assert FAIL |

Handoff: **be-ver → code-reviewer**
