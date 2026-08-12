# A3 W8 агент 5 — ent/schema part 5 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W8-5-schemas.md](A2W8-5-schemas.md) · Stage A: [A8-5-schemas.md](A8-5-schemas.md)  
> Assignment: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md) § A5 · Периметр: **15** · Прочитано: **15/15** · Пропущено: **0**  
> Фокус: UUID / version / soft-delete / indexes / OnDelete  
> Remediation: TASK-095 (VersionMixin) · TASK-039 (soft TestSuite/WebhookEndpoint) · TASK-068 (mojibake) · TASK-096 (peer Project Cascade; intermediate owned_child — gap)

## Сверка с A2 (A2W8-5-*)

| A2 ID | Tag | Статус | Evidence в HEAD |
|-------|-----|--------|-----------------|
| A2W8-5-001 | RESIDUAL | частично | **version** FIXED: `team.go:30-35` `VersionMixin{}` (TASK-095). **хвост:** нет `deleted_at` / `updated_at` → A3 active |
| A2W8-5-002 | FIXED-OK | закрыт | `testrun.go:62-67` `VersionMixin{}` |
| A2W8-5-003 | FIXED-OK | закрыт | `testsuite.go:41-46` `VersionMixin{}`; soft `deleted_at` уже был (TASK-039) |
| A2W8-5-004 | FIXED-OK | закрыт | `user.go:80-85` `VersionMixin{}` (lifecycle через `status`, не trash) |
| A2W8-5-005 | RESIDUAL | частично | **version** FIXED: `workflowstatus.go:42-47` `VersionMixin{}`. **хвост:** нет `deleted_at`; hard `Delete()` в `tracker/repo_workflow.go:66-71` |
| A2W8-5-006 | FIXED-OK | закрыт | `webhookendpoint.go:43-48` `VersionMixin{}` + `deleted_at` (TASK-039) |
| A2W8-5-007 | FIXED-OK | закрыт | `systemsetting.go:31-36` `VersionMixin{}` |

### Cascade FIXED-OK (Stage A / A2 уже закрытые в периметре)

| Класс | Tag | Evidence |
|-------|-----|----------|
| TestCaseVersion unique `(test_case, version_number)` | FIXED-OK | `testcaseversion.go:52-53` |
| Mojibake Allure/godoc | FIXED-OK | «Имя…» / «Идентификатор…» / «Рёбра»; grep mojibake = 0 |
| TestSuite / WebhookEndpoint soft-delete | FIXED-OK | `deleted_at` + index; DeleteSuite/DeleteWebhookEndpoint → `SetDeletedAt` |
| UUID PK + `nowUTC` | FIXED-OK | все 15 схем |
| Project→teams/test_*/webhook_endpoints/workflow_statuses Cascade | FIXED-OK (peer) | `project.go` owned_child (TASK-096); **промежуточные** To-children в A5 — NEW ниже |

**REGRESS:** 0 · **FIXED-OK (A2 actives):** 5 (+ version-часть 001/005) · **RESIDUAL (A2):** 2 · **NEW:** 5

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/systemsetting.go` | 37 | 0 | VersionMixin FIXED-OK; KV без trash — OK |
| `backend/ent/schema/team.go` | 45 | 2 | RESIDUAL soft; NEW OnDelete members |
| `backend/ent/schema/teammember.go` | 48 | 0 | junction Unique(team,user) чисто |
| `backend/ent/schema/testcase.go` | 134 | 1 | эталон OL+trash+indexes; NEW OnDelete versions |
| `backend/ent/schema/testcaseversion.go` | 55 | 0 | unique FIXED-OK; snapshot append |
| `backend/ent/schema/testresult.go` | 92 | 0 | append-only чисто |
| `backend/ent/schema/testrun.go` | 97 | 1 | VersionMixin FIXED; NEW OnDelete results |
| `backend/ent/schema/testsuite.go` | 70 | 0 | soft+version FIXED-OK |
| `backend/ent/schema/timelog.go` | 59 | 0 | append; Issue→time_logs Cascade (peer) |
| `backend/ent/schema/user.go` | 161 | 0 | VersionMixin FIXED; status lifecycle |
| `backend/ent/schema/userblock.go` | 56 | 0 | Unique(blocker,blocked) чисто |
| `backend/ent/schema/webhookdelivery.go` | 58 | 0 | append log чисто |
| `backend/ent/schema/webhookendpoint.go` | 68 | 1 | soft+version FIXED; NEW OnDelete deliveries |
| `backend/ent/schema/workflowstatus.go` | 73 | 2 | RESIDUAL soft/hard; NEW OnDelete transitions |
| `backend/ent/schema/workflowtransition.go` | 54 | 0 | Unique(from,to); VersionMixin (TASK-095 list) |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W8-5-001 | RESIDUAL (A2W8-5-001∩soft) | P2 | `team.go:16-28` | Team: есть VersionMixin, нет `deleted_at` / `updated_at` | Mutable user-visible; TZ soft-delete / корзина. DeleteTeam API нет, но PermanentDelete project CASCADE hard-purge'ит teams без trash | `deleted_at` (+ index) + `updated_at`; soft path при dissolve/GDPR |
| A3W8-5-002 | RESIDUAL (A2W8-5-005∩soft) | P1 | `workflowstatus.go:17-40`; runtime `tracker/repo_workflow.go:66-71` | Нет `deleted_at`; `DeleteWorkflowStatus` = hard `Delete()` | Concurrent rename закрыт Mixin, но удаление необратимо; issues `Required` status → FK trap / потеря статуса; TZ trash | soft-delete + `DeletedAtIsNil` в list **или** запрет delete при issues + ADR |
| A3W8-5-003 | NEW | P2 | `team.go:43` | `To("members", TeamMember)` без `entsql.OnDelete(Cascade)` | ADR owned_child / симметрия Channel→members; Project Cascade→Team при живых TeamMember → FK trap | `Annotations(entsql.OnDelete(entsql.Cascade))` + migrate FK |
| A3W8-5-004 | NEW | P2 | `testcase.go:112` | `To("versions")` без Cascade | hard-purge TestCase (project PermanentDelete / GDPR) → orphan / FK trap на `TestCaseVersion` | Cascade на `versions` (как Issue→comments) |
| A3W8-5-005 | NEW | P2 | `testrun.go:81` | `To("results")` без Cascade | owned TestResult; Project Cascade→TestRun при живых results → FK trap | Cascade на `results` |
| A3W8-5-006 | NEW | P2 | `webhookendpoint.go:58` | `To("deliveries")` без Cascade | owned delivery log; hard-purge endpoint → FK trap | Cascade на `deliveries` |
| A3W8-5-007 | NEW | P2 | `workflowstatus.go:58` | `To("from_transitions")` без Cascade | hard Delete status при transitions → FK trap (to_status edge тоже Required) | Cascade на `from_transitions` (+ политика to_status / app purge transitions before status) |

## Матрица фокуса (UUID / version / soft-delete / OnDelete)

| Схема | UUID PK | version | deleted_at | OnDelete owned To | Класс | Вердикт |
|-------|:-------:|:-------:|:----------:|:-----------------:|-------|---------|
| SystemSetting | ✓ | ✓ Mixin | ✗ (KV OK) | — | mutable KV | FIXED-OK |
| Team | ✓ | ✓ Mixin | ✗ | members ✗ | mutable | A3W8-5-001/003 |
| TeamMember | ✓ | — | — | — | junction | чисто |
| TestCase | ✓ | ✓ Mixin | ✓ | versions ✗ | эталон+gap | A3W8-5-004 |
| TestCaseVersion | ✓ | — (`version_number`) | — | — | snapshot | unique FIXED |
| TestResult | ✓ | — | — | — | append | чисто |
| TestRun | ✓ | ✓ Mixin | ✗ (hist OK) | results ✗ | mutable | A3W8-5-005 |
| TestSuite | ✓ | ✓ Mixin | ✓ | — | mutable+trash | FIXED-OK |
| TimeLog | ✓ | — | — | — | append | Issue Cascade peer |
| User | ✓ | ✓ Mixin | — (`status`) | — | mutable | FIXED-OK |
| UserBlock | ✓ | — | — | — | junction | чисто |
| WebhookDelivery | ✓ | — | — | — | append | чисто |
| WebhookEndpoint | ✓ | ✓ Mixin | ✓ | deliveries ✗ | mutable+trash | A3W8-5-006 |
| WorkflowStatus | ✓ | ✓ Mixin | ✗ | transitions ✗ | mutable | A3W8-5-002/007 |
| WorkflowTransition | ✓ | ✓ Mixin | — | — | junction | чисто |

## Чисто (подтверждено после remediation)

- **UUID PK** `Default(uuid.New).Immutable()` — 15/15.
- **VersionMixin** на всех A2 mutable gaps + `workflowtransition` (список TASK-095 `mutableNeedsVersion`); guard `TestSchema_TASK095_VersionMixinOnMutable`.
- **Indexes:** TestCase project+`deleted_at` (A8-002); TestCaseVersion unique; TeamMember/UserBlock/WorkflowTransition Unique edges; TestSuite/WebhookEndpoint `deleted_at`; WorkflowStatus unique `(project,name)`.
- **Soft-delete OK:** TestCase, TestSuite, WebhookEndpoint; User через `status`/`pending_since` (не `deleted_at` — accepted pattern).
- **Append/junction без OL/trash:** TestResult, TimeLog, WebhookDelivery, TeamMember, UserBlock — N/A.
- **Sensitive:** WebhookEndpoint `secret` `Sensitive()` сохранён.
- **Mojibake:** FIXED (TASK-068).

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 6 |
| P3 | 0 |
| **Total open** | **7** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 actives 002–004, 006–007 + version∩001/005) | 5+ |
| RESIDUAL | 2 (`A3W8-5-001`, `A3W8-5-002`) |
| NEW | 5 (`A3W8-5-003`…`007` OnDelete) |
| REGRESS | 0 |

A2 baseline open: 7 → **FIXED-OK 5** (полностью) · **RESIDUAL 2** (soft-хвосты) · **NEW 5** · REGRESS 0 · open **7** (P1=1 · P2=6).

**Топ ID:** A3W8-5-002 (WorkflowStatus hard-delete) · A3W8-5-005 (TestRun→results OnDelete) · A3W8-5-003 (Team→members OnDelete).
