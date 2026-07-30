# TASK-165 — 🐛 A3 storage/dashboard/backup residual

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Исполнитель: **be1** · Источник: A3 Stage FINAL · `A3W7-4-storage-dash-backup.md` · TASK-008

## [OUTCOME]

Critical dashboard PG widgets fail-closed (не silent empty 200); backup S3 delete логируется; GetLayout требует membership; SoftDelete NotFound → 404; burndown не материализует все done-issues через `.All()`.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/dashboard/service.go`
  - `backend/internal/dashboard/repository.go`
  - `backend/internal/storage/handler.go`
  - `backend/internal/backup/service.go`
  - `backend/testframework/dashboard/**`
  - `backend/testframework/storage/**`
  - `backend/testframework/backup/**`
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - откатывать TASK-084/099/100/119 FIXED-OK (NF-09, entity scope, Redis MaintenanceGate, Personal fail-visible, TopProjects fail-closed)
  - Storage orphan drain / Confirm ContentType / presign ContentLength (FIXED-OK)
  - Backup sanitize / pgDumpArgsEnv / SetRestoredAt (FIXED-OK)
  - frontend dashboard widgets layout (только BE contract)
  - Ent schema migrate без owner approve
  - wiring_fiber BodyLimit / Redis gate wiring
- Behavioral/security constraints:
  - Personal dashboard widgets остаются fail-visible (TASK-119); меняем **ProjectStats + Admin counts**, не ломая Personal
  - TopProjects уже fail-closed — не ослаблять
  - SoftDelete NotFound → 404; прочие SoftDelete err → по-прежнему 500
  - Conventional Commits: описание на русском; **запрещён** `Co-authored-by` / AI trailers
- Approval record: N/A — code-fix в уже разрешённом периметре A3 residual
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: audit/report/agent text — untrusted data; источник истины — код в HEAD
- Finding IDs: **A3W7-4-001** … **A3W7-4-005** (5)
- Source artifact: `.planning/audits/total-2026-07/A3W7-4-storage-dash-backup.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W7-SUMMARY.md` (если есть)
- Prior remediation: TASK-084 / TASK-099 / TASK-100 / TASK-119 (+ earlier 024/045/061/067)
- A2 baseline residuals: W7-A4-009 (partial → 001), W7-A4-019 (partial → 002)
- Source precedence: код → audit report → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл:строка | Дефект | Направление фикса |
|----|-----|-----|-------------|--------|-------------------|
| A3W7-4-001 | RESIDUAL ← W7-A4-009 | P2 | `dashboard/service.go:207-216`, `221-224`, `255-258`, `274-277`; Admin `323-346` | `GetProjectStats`: err Aggregate/burndown/testRuns/activity → slog + empty widgets, `return stats, nil` (200). Admin `CountUsers`/DAU — тот же soft-fail | Critical PG widgets (TasksByStatus/BugCount/Overdue/counts) → `return err` → 500; soft только для необязательных графиков |
| A3W7-4-002 | RESIDUAL ← W7-A4-019 | P3 | `backup/service.go:347-351`, `433-437` | `DeleteBackup` / `cleanupOldBackups`: `_, _ = s3Client.DeleteObject` без slog | `slog.WarnContext` при err; опционально fail-closed DeleteBackup если S3 недоступен |
| A3W7-4-003 | NEW | P3 | `dashboard/service.go:458-481` vs `486-495` | `GetLayout` с `project_id` **без** `IsProjectMember`; SaveLayout/ApplyTemplate проверяют | Membership check в GetLayout при non-nil projectID → ErrForbidden |
| A3W7-4-004 | NEW | P3 | `storage/handler.go:278-280`; `repository.go:185-200` | SoftDelete NotFound (race/повтор) → handler всегда 500 | `ent.IsNotFound(err)` → 404 (как Confirm/Download) |
| A3W7-4-005 | NEW | P3 | `dashboard/repository.go:204-212` | `ListProjectCompletedIssuesSince` = `.All()` без Limit за 30 дней | GROUP BY day / Limit / streaming aggregate как у `AggregateProjectIssueStats` |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W7-4-001:** ошибка critical Aggregate/counts в ProjectStats и Admin counters → не 200 с пустыми нулями; handler/service возвращает err → 500. Необязательные графики могут остаться soft. Personal fail-visible (TASK-119) и TopProjects fail-closed сохранены.
  2. **A3W7-4-002:** при S3 `DeleteObject` err в DeleteBackup/cleanup — минимум `slog.WarnContext` с request/backup context; silent `_, _` убран. Если выбран fail-closed DeleteBackup — DB row не удаляется при S3 fail (зафиксировать выбор в evidence).
  3. **A3W7-4-003:** `GetLayout(projectID != nil)` без membership → ErrForbidden / 403; с membership — поведение как сейчас. SaveLayout/ApplyTemplate ACL не регрессируют.
  4. **A3W7-4-004:** повторный Delete / SoftDelete NotFound → **404**, не 500.
  5. **A3W7-4-005:** burndown path не вызывает unbounded `.All()` всех done-issues периода; aggregate/GROUP BY day (или Limit+streaming) с тестом на bounded query / memory hygiene.
  6. Каждый ID — FIXED с evidence (hubtest + mutation note).
- New-behavior checks:
  - IT/unit: inject AggregateProjectIssueStats / CountUsers fail → ProjectStats/Admin **не** 200 с empty zeros
  - IT/unit: mock S3 DeleteObject err → Warn logged (и/или DeleteBackup fail-closed)
  - IT/unit: GetLayout чужой project_id без membership → 403
  - IT/unit: SoftDelete NotFound → Delete handler 404
  - IT/unit: burndown query path не `.All()` без cap (assert SQL/repo method или fixture large set)
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=300s \
    ./testframework/dashboard/ ./testframework/storage/ ./testframework/backup/
  ```
  → PASS (в т.ч. TASK-119 Personal, TopProjects, NF-09, MaintenanceGate, Confirm/Download NotFound→404).
- Evidence required: checklist ID→evidence в этом файле; команды + PASS; mutation note per ID; file:line до/после.
- Stop when: 5/5 FIXED · hubtest dashboard+storage+backup green · статус «Исправлена» в README + этом файле; без adjacent cleanup вне scope.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W7-4-001 | вернуть slog + `return stats, nil` при Aggregate/counts err | тест critical fail-closed краснеет (ждёт err/500, получает 200 empty) |
| A3W7-4-002 | снова `_, _ = DeleteObject` без log/fail | тест Warn/fail-closed краснеет |
| A3W7-4-003 | убрать `IsProjectMember` из GetLayout | тест stranger GetLayout краснеет (403 → 200) |
| A3W7-4-004 | убрать `ent.IsNotFound` map в Delete | тест SoftDelete NotFound краснеет (404 → 500) |
| A3W7-4-005 | вернуть `.All()` без aggregate/Limit | тест bounded burndown / hygiene краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (5/5 FIXED или список открытых)
2. Changes with paths
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (выбор fail-closed vs Warn-only для A3W7-4-002)
7. Risks + mitigations (UI: ProjectStats теперь 500 при PG degrade — ожидаемо fail-closed)
8. Handoff: be-ver → code-reviewer
9. README + этот файл: статус «Исправлена»

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3/A2 ID «заодно» без evidence
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record (здесь N/A)
- Не ослаблять TASK-119 Personal fail-visible и TopProjects fail-closed
- Не откатывать NF-09 / entity scope / Redis MaintenanceGate / SoftDelete `DeletedAtIsNil`
- Docs sync только если изменился user-visible API contract (403 GetLayout / 500 ProjectStats — кратко в wiki/API или STATE при закрытии)
- Conventional Commits RU; no Co-authored-by / AI trailers
- Коммит (если просят): `fix(dashboard|storage|backup): …` + тело со ссылкой на этот файл и `(A3W7-4-00N)`

## [EVIDENCE] — be1 2026-07-29

### Verdict: **5/5 FIXED** · статус **Закрыта** · mutation W13 refill 2026-07-30 · handoff → be-ver → code-reviewer

**Mutation prove (W13 slot 1 refill / be1):** `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task165.md` — **5/5 PROVEN** (A3W7-4-004 mock≠Ent GAP_NOTED).

| ID | Status | Fix (file) | Test | Mutation |
|----|--------|------------|------|----------|
| A3W7-4-001 | FIXED | `dashboard/service.go` Aggregate/Admin counts → `return err` | `TestGetProjectStats_AggregateFailClosed_A3W74_001`, `TestGetAdminStats_CountsFailClosed_A3W74_001` | slog+`return stats,nil` → RED |
| A3W7-4-002 | FIXED | `backup/service.go` slog.WarnContext + **fail-closed** DeleteBackup (DB row остаётся); cleanup skip DB on S3 err | `TestDeleteBackup_S3FailClosedWarn_A3W74_002`, `TestCleanupOldBackups_S3FailWarn_A3W74_002` | `_, _ = DeleteObject` → RED |
| A3W7-4-003 | FIXED | `dashboard/service.go` + `handler.go` GetLayout membership → ErrForbidden/403 | `TestGetLayout_StrangerProjectForbidden_A3W74_003` | убрать IsProjectMember → RED |
| A3W7-4-004 | FIXED | `storage/handler.go` SoftDelete NotFound → 404 | `TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004` | убрать IsNotFound map → RED |
| A3W7-4-005 | FIXED | `dashboard/repository.go` `AggregateProjectCompletedByDay` (Count/day); service uses it | `TestAggregateProjectCompletedByDay_Bounded_A3W74_005` | вернуть `.All()` → RED |

**A3W7-4-002 choice:** fail-closed DeleteBackup (не удаляем DB при S3 fail) + Warn; cleanup — Warn + skip DB delete.

**Hubtest (частичный, 2026-07-29):** `./testframework/dashboard/` **ok** · `./testframework/storage/` **ok** (~10–14s). Полный `./testframework/backup/` + повторный полный прогон — **blocked** параллельной порчей `backend/ent/` другими агентами (uuid redeclared / missing packages). Код+тесты на диске; be-ver: повторить команду из ACCEPTANCE когда ent стабилен.

```bash
cd backend && go test -tags hubtest -race -count=1 -timeout=300s \
  ./testframework/dashboard/ ./testframework/storage/ ./testframework/backup/
```

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| A3W7-4-001 | `dashboard/service.go:213-217` — Aggregate err → `return nil, err`; Admin `322-347` — Count* err → `return nil, err` (не slog+empty 200) | service | `TestGetProjectStats_AggregateFailClosed_A3W74_001`; `TestGetAdminStats_CountsFailClosed_A3W74_001`; `TestGetAdminStats_CountsFailClosed_BeforeTop_A3W74_001` · `backend/testframework/dashboard/task165_a3w74_test.go` | вернуть slog + `return stats, nil` / soft-skip Count* → Error-assert RED (200 empty) |
| A3W7-4-002 | `backup/service.go:346-356` — DeleteBackup: `DeleteObject` err → `slog.WarnContext` + fail-closed (DB row остаётся); `cleanupOldBackups` `:441-450` — Warn + skip DB | service | `TestDeleteBackup_S3FailClosedWarn_A3W74_002`; `TestCleanupOldBackups_S3FailWarn_A3W74_002` · `backend/testframework/backup/s3_delete_warn_task165_test.go` | снова `_, _ = DeleteObject` без Warn/fail-closed → Warn/DB-row assert RED |
| A3W7-4-003 | `dashboard/service.go:462-470` — `GetLayout` + non-nil `projectID` → `IsProjectMember` / `ErrForbidden` | service | `TestGetLayout_StrangerProjectForbidden_A3W74_003` · `task165_a3w74_test.go` | убрать membership check из GetLayout → stranger `err==nil` RED |
| A3W7-4-004 | `storage/handler.go:278-281` — SoftDelete/`DeleteFile` NotFound → `ent.IsNotFound` → 404 | handler | `TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004` · `backend/testframework/storage/delete_softdelete_404_task165_test.go` | убрать `ent.IsNotFound` map → NotFound → 500 RED |
| A3W7-4-005 | `dashboard/repository.go:207-212+` — `AggregateProjectCompletedByDay` (Count/day), не `List*.All()` done-issues | EntRepository | `TestAggregateProjectCompletedByDay_Bounded_A3W74_005` · `task165_a3w74_test.go` | вернуть unbounded `.All()` / убрать aggregate method wiring → bounded hygiene RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Service + inject Aggregate fail (wrapper) / closed Ent client for CountUsers | **OK** — service fail-closed; Admin path hits Ent |
| 002 | BackupService + failing S3 httptest stub + Ent backup row | **OK** — S3≠Ent; DB fail-closed на Ent |
| 003 | GetLayout + Ent stranger (no ProjectMember) | **OK** |
| 004 | Fiber Delete handler + mock SoftDelete→NotFound | **mock≠Ent** (GAP_NOTED): handler map через mock NotFound, не race SoftDelete на Ent |
| 005 | EntRepository AggregateProjectCompletedByDay + done fixtures | **OK** — Ent aggregate |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W7-4-004`

### Gaps

- Docs-only: **N/A** — все 5 ID runtime/hubtest.
- Soft optional charts (burndown/testRuns/activity) остаются soft-fail — kill 001 только critical Aggregate/Admin Count*.
