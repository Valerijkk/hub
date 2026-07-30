# C-W19 — Red-team: tracker / sprint / project / tms / bases

> TASK-008 · Stage C · Wave 19 · READ-ONLY · 2026-07-29  
> Агент: независимый red-team (`cursor-grok-4.5`) · **не автор** фиксов TASK-134…142  
> HEAD проверен: `617af223` · Audit baseline A3: `ee353287` (ancestor OK)  
> Scope: `tracker` · `sprint` · `project` · `tms` · `bases`  
> Источники A3: `A3W1-1`…`A3W1-5`, `A3W2-1`…`A3W2-5`  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 не закрывать**

---

## Вердикт

| Вопрос | Ответ |
|--------|--------|
| A3W1*/A3W2* open ID (CSV/DEDUP) → не в TASK-124…176? | **0** — 36 ID → TASK-134…142 |
| P0/P1 orphans в периметре A3? | **0** (P1 только A3W1-2-001 → TASK-134; open P0 не было) |
| A2 residual, живой на HEAD, **вне** TASK-124…176? | **1** — `A2W1-A5-003` (RemoveTeamMember last-member) |
| Комбо soft-delete × list caps × IDOR labels/channels/FileIDs | Исторические **MITIGATED** на HEAD |
| Trash resurrect / import partial | Project resurrect **MITIGATED**; TestRail Partial **MITIGATED**; Allure — **1 NEW residual** |
| Ordinary member vs foreign project | Классический cross-project IDOR **закрыт** |

**Итог:** постановка бэклога по A3 open ID периметра **полная** (DEDUP). Код HEAD несёт remediation TASK-134…142 (маркеры `A3W1-*` / `A3W2-*`). Остаётся **A2 residual без TASK** + **Allure silent-partial** + process drift «На проверке». **TASK-008 не закрывать** без независимой W18-мутации.

---

## 1. Gaps: A3W1* / A3W2* → TASK-124…176

### 1.1 Маппинг open A3 → backlog

| Источник | Open ID (A3) | Sev (на момент A3) | Backlog | Coverage |
|----------|--------------|--------------------|---------|----------|
| A3W1-1 handlers | — | — | — | N/A (0 open) |
| A3W1-2 services | A3W1-2-001…006 | P1×1 / P2×2 / P3×3 | [TASK-134](../../backlog/TASK-134-2026-07-29-a3-1-2-tracker-services.md) | ✅ 6/6 |
| A3W1-3 repos | A3W1-A3-001…003 | P2 / P3×2 | [TASK-135](../../backlog/TASK-135-2026-07-29-a3-1-3-tracker-repos.md) | ✅ 3/3 |
| A3W1-4 sprint | A3W1-A4-001…003 | P2 / P3×2 | [TASK-136](../../backlog/TASK-136-2026-07-29-a3-1-4-sprint.md) | ✅ 3/3 |
| A3W1-5 project | A3W1-A5-001…006 | P2×2 / P3×4 | [TASK-137](../../backlog/TASK-137-2026-07-29-a3-1-5-project.md) | ✅ 6/6 |
| A3W2-1 handlers | A3W2-1-001 | P3 | [TASK-138](../../backlog/TASK-138-2026-07-29-a3-2-1-tms-handlers.md) | ✅ |
| A3W2-2 services | A3W2-2-001…006 | P2×4 / P3×2 | [TASK-139](../../backlog/TASK-139-2026-07-29-a3-2-2-tms-services.md) | ✅ 6/6 |
| A3W2-3 repos | A3W2-3-001…003 | P2 / P3×2 | [TASK-140](../../backlog/TASK-140-2026-07-29-a3-2-3-tms-repos.md) | ✅ 3/3 |
| A3W2-4 import | A3W2-4-001…005 | P2×2 / P3×3 | [TASK-141](../../backlog/TASK-141-2026-07-29-a3-2-4-tms-import.md) | ✅ 5/5 |
| A3W2-5 bases | A3W2-A5-001…003 | P3×3 | [TASK-142](../../backlog/TASK-142-2026-07-29-a3-2-5-bases.md) | ✅ 3/3 |
| **Σ A3 CSV** | **36** | | **9 TASK** | **missed = 0** |

Канон: [`A3-DEDUP-2026-07-29.md`](A3-DEDUP-2026-07-29.md) (unmapped 0 для 195).

### 1.2 Orphan вне DEDUP (A2 residual, жив на A3 + HEAD)

| ID | Sev | Evidence A3 | В TASK-124…176? | HEAD |
|----|-----|-------------|-----------------|------|
| **A2W1-A5-003** | P3 | `A3W1-5-project.md` — active residual; TASK-137 явно **исключает** | **Нет** | `RemoveTeamMember` без guard last-member (`project/service.go:624-632`); `requireTeamMember` → orphan team навсегда |

→ Это **единственный** gap «не попало в TASK-124…176» в периметре. Не P0/P1; DEDUP считал только A3 CSV open (195), A2 residual **не** вошёл в кластеры.

### 1.3 Процессный gap (не finding-orphan)

| Симптом | Evidence |
|---------|----------|
| Файлы TASK-134…142: evidence FIXED / «На проверке» | frontmatter + checklist в backlog |
| Prod уже с A3-маркерами | §2–3 |
| README может держать «Открыта» | `.planning/backlog/README.md` |

---

## 2. Findings (Stage C / HEAD `617af223`)

| ID | Severity | file:line | Evidence | Уже в backlog? |
|----|----------|-----------|----------|----------------|
| **C-W19-TT-001** | **P3** | `project/service.go:624-632` (+ wave122 empty-team OK) | `RemoveTeamMember`: любой member удаляет любого; нет `TeamMemberCount≤1` → 400. Пустая команда → никто не проходит `requireTeamMember` → **orphan team** (AddTeamMember / AddTeamToProject мертвы) | **Orphan A2W1-A5-003** — **NEW TASK нужен** (не 124…176) |
| **C-W19-TT-002** | **P2** | `tms/service_import.go:131-133`, `:173-177`, `:112-166` | Allure ExistingRun: (1) `StoreAllureResults` err → только slog, импорт идёт; (2) `buildAllureResultInputs` при `matchErr` → `continue` (в т.ч. transient DB после A2W2-4-003 fail-closed matcher); (3) ответ = `GetTestRunDetail` **без** `Partial` (в отличие от TestRail/CSV `ImportResult.Partial`). Итог: **тихий partial** / пустой PG-run при DB blip матчера + возможный Mongo/PG drift | **NEW** (не закрыто TASK-141; A2W2-4-004 покрыл только CSV/TestRail Partial) |
| C-W19-TT-003 | P3 | process | TASK-134…142 «На проверке» при prod FIXED — ложный signal до независимой мутации | process (как C-W19 messenger R3) |
| A3W1-2-001…006 | — | labels/group/channel/links/comments | `CountLabelsInProject` на Create+Update; `FindLabelGroupInProject`; `ensureActionChannelInProject`; AppError wrap; comment 404≠409 | **TASK-134** |
| A3W1-A3-001…003 | — | automation FilterSubjects / executeRule / StartTimer | `AutomationFilterSubjects` moved/created/assigned; StartTimer `DeletedAtIsNil` | **TASK-135** |
| A3W1-A4-001…003 | — | Complete incomplete / metrics / retro OL | `FindIncompleteIssues`+move `DeletedAtIsNil`; counts/burndown/velocity; Retro `VersionEQ` | **TASK-136** |
| A3W1-A5-001…006 | — | resurrect / hard-delete / CreateTeam tx | Archive/Unarchive reject deleted; `RestoreFromTrash`; PermanentDelete admin+trash; CreateTeamWithCreator tx | **TASK-137** |
| A3W2-*-* | — | suite OL / AssignTesters / FileIDs / import tx / caps / TestRail / hubbase | маркеры A3W2-* в prod + testframework Task13x/14x | **TASK-138…142** |

---

## 3. Combinations / attack chains

### 3.1 Запрошенные комбо

#### C-W19-TT-C1 — soft-delete × list caps × IDOR (labels / channels / FileIDs)

| Вектор | Исторический gap | HEAD | Вердикт |
|--------|------------------|------|---------|
| Foreign / trash **label** на UpdateIssue | A3W1-2-001: Update без `CountLabelsInProject` | `service_issues.go:339-347` + `repo_labels.go:64-73` (`DeletedAtIsNil`) | **MITIGATED** |
| Foreign **LabelGroup** на CreateLabel | A3W1-2-002 | `FindLabelGroupInProject` | **MITIGATED** |
| Soft-delete label + Count → attach | trash label считался «в проекте» | Count с `DeletedAtIsNil` → BadRequest | **MITIGATED** |
| Foreign / trash **channel** automation | A3W1-2-003 | `ensureActionChannelInProject` → `FindChannelInProject` + `channel.DeletedAtIsNil` (`repo_issue_message_link.go:77-84`) | **MITIGATED** |
| Foreign / trash **FileIDs** ExecuteResult | A3W2-2-002 | `CountFilesInProject` + `filemetadata.DeletedAtIsNil` (`service_runs.go:286-301`) | **MITIGATED** |
| AssignTesters non-member | A3W2-2-001 | `CountProjectMembers` → 400 | **MITIGATED** |
| List caps + soft-delete TMS edges | A3W2-3-001/002/003 | Limit/First на ByRunIDs/LastResults/AddCases; CountLabels soft-delete; linked edges `DeletedAtIsNil` | **MITIGATED** |
| Sprint Complete × trash issues | A3W1-A4-001/002 | `FindIncompleteIssues` + metrics `DeletedAtIsNil` | **MITIGATED** |
| StartTimer на trash issue | A3W1-A3-003 | lock `DeletedAtIsNil` (`repo_time.go:82`) | **MITIGATED** |

**Комбо «хуже суммы» (исторический):** soft-delete прятал ресурс из list, но UUID оставался валидным для M2M attach / FileIDs / channel_id → cross-project или trash-side-channel. На HEAD count/find helpers фильтруют project **и** `DeletedAtIsNil`.

**Residual:** сообщение «метка не принадлежит проекту» для trash-label той же проекта — UX smell, не IDOR.

#### C-W19-TT-C2 — trash resurrect

| Путь | Было | HEAD | Вердикт |
|------|------|------|---------|
| Project Archive/Unarchive → ClearDeletedAt | A3W1-A5-001 | `UpdateStatus(..., false)` **не** ClearDeletedAt; deleted → 404; только `RestoreFromTrash` (`service.go:215-266`, `:334-356`; `repository.go:240-250`) | **MITIGATED** |
| PermanentDelete active / non-admin | A3W1-A5-002 | только `status=deleted` + `IsSystemAdmin` (`:358-378`) | **MITIGATED** |
| HubBase trash 30d | A3W2-A5-001 | audit `type=hubbase` List/Restore/Purge + cascade RestoreHubBase (`audit/service.go:269-311`, `repository.go:722-762`) | **MITIGATED** |
| SoftDeleteTable orphans × SoftDeleteBase | A3W2-A5-002 | SoftDeleteTable cascade children; SoftDeleteBase all tableIDs (`bases/repository.go:146-211`, `:314+`) | **MITIGATED** |

**Ordinary member:** не может воскресить чужой/soft-deleted project через archive; PermanentDelete — admin only. HubBase restore — admin trash API.

#### C-W19-TT-C3 — import partial

| Путь | Было | HEAD | Вердикт |
|------|------|------|---------|
| TestRail/CSV Partial flag | A2W2-4-004 | `ImportResult.Partial` + Errors/Skipped; CreateSuite fail → Errors++ (A3W2-4-005) | **MITIGATED** |
| ExistingRun PG half-write | A3W2-2-003 | `ImportResultsToExistingRun` Ent tx + mock compensating delete | **MITIGATED** |
| TestRail unbounded nest/rows | A3W2-4-002 | depth + `maxImportRecords` | **MITIGATED** |
| **Allure ExistingRun silent partial** | matcher fail-closed (A2) × continue в builder × нет Partial | см. **C-W19-TT-002** | **OPEN residual** |
| Allure create-path Mongo after PG | best-effort Mongo post-commit | `service_import.go:96-99` — осознанный best-effort; PG атомарнее ExistingRun Mongo-first | residual ops (не IDOR) |

---

### 3.2 Attack chains — ordinary member vs foreign project

Threat actor: authenticated **member** проекта A (не founder/admin), знает UUID ресурсов проекта B.

| # | Цепочка | Ожидание HEAD | Статус |
|---|---------|---------------|--------|
| AC-1 | `UpdateIssue(A)` + `Labels=[label_from_B]` | CountLabelsInProject → 400 | ✅ |
| AC-2 | Automation rule A + `channel_id` канала B | FindChannelInProject → 404 | ✅ |
| AC-3 | Message-link issue A ↔ channel B | FindChannelInProject + IsChannelMember | ✅ |
| AC-4 | `ExecuteTestResult` + FileID из B | CountFilesInProject → 404 | ✅ |
| AC-5 | `AssignTesters` UUID non-member / user из B | CountProjectMembers → 400 | ✅ |
| AC-6 | `GetTestRun` / CreateBug run из B при path project A | GetTestRunInProject → 404 | ✅ |
| AC-7 | Suite/Component/TC Update/Delete foreign project | HasProjectWith / projectID scope | ✅ |
| AC-8 | Sprint Complete / poker / SetIssuesSprint foreign | GetSprintInProject / IssueInProject | ✅ |
| AC-9 | Project Get/List trash / members foreign | membership + RBAC outsider 403 | ✅ |
| AC-10 | Bases CRUD чужой base | `authorizeBase` → 403/404 | ✅ |
| AC-11 | Soft-deleted issue A: mutate / StartTimer / Complete move | DeletedAtIsNil guards | ✅ |
| AC-12 | Archive soft-deleted project → resurrect | 404 / no ClearDeletedAt | ✅ |
| AC-13 | PermanentDelete без admin | Forbidden | ✅ |
| AC-14 | Remove last team member → orphan | **allowed** (C-W19-TT-001) | ⚠️ residual P3 |
| AC-15 | Allure import: Mongo down / match DB err | 200 + silent skip (C-W19-TT-002) | ⚠️ residual P2 |

**Цепочка member→foreign project IDOR:** на HEAD **не** проходит через labels/channels/FileIDs/runs/suites/sprint/bases. Остаётся abuse **внутри своей** team (empty orphan) и **Allure integrity** (не cross-project).

---

## 4. Clean checklist (что проверено и чисто)

| Зона | Проверено | Вердикт |
|------|-----------|---------|
| A3W1-1 handlers IDOR/perm/JQL | scoped ParseProjectID; CreateBoard PermTaskEdit; A2 closes | ✅ чисто (0 open A3) |
| Tracker Create/Update labels | CountLabelsInProject + soft-delete | ✅ |
| Tracker automation channel | ensureActionChannelInProject | ✅ |
| Tracker FilterSubject Nak storm | AutomationFilterSubjects narrow | ✅ |
| Sprint soft-delete metrics/complete | DeletedAtIsNil на incomplete/counts/velocity | ✅ |
| Sprint retro OL | VersionEQ + DTO version | ✅ |
| Project trash resurrect via archive | blocked | ✅ |
| Project hard-delete TZ | admin + trash only | ✅ |
| CreateTeam atomic creator | Ent-tx | ✅ |
| TMS AssignTesters / FileIDs | project membership / file scope | ✅ |
| TMS ExistingRun PG tx | ImportResultsToExistingRun | ✅ |
| TMS list caps / label count / linked trash | TASK-140 | ✅ |
| TestRail/CSV Partial + caps | TASK-141 | ✅ |
| Bases SoftDeleteTable / MaxLen / hubbase trash | TASK-142 | ✅ |
| RBAC outsider tracker/tms/sprint/project/bases | Fiber v3 order + service scope | ✅ (классический IDOR) |

---

## 5. Сводка severity (residual после remediation)

| ID | Sev | Тип | Суть |
|----|-----|-----|------|
| C-W19-TT-001 | P3 | orphan backlog | A2W1-A5-003 RemoveTeamMember last-member — **нет TASK в 124…176** |
| C-W19-TT-002 | P2 | NEW combo | Allure ExistingRun: Mongo slog-only + match `continue` + нет `Partial` |
| C-W19-TT-003 | P3 | process | README/статусы «Открыта/На проверке» при prod FIXED |

Новых **P0/P1 product IDOR** (member → foreign project) **не** найдено. Orphan A3 CSV findings → backlog: **0**. Orphan A2 residual: **1**.

---

## 6. Evidence (file:line, HEAD `617af223`)

| Проверка | Путь |
|----------|------|
| UpdateIssue labels scope | `backend/internal/tracker/service_issues.go:339-347` |
| CountLabels soft-delete | `backend/internal/tracker/repo_labels.go:64-73` |
| Automation channel | `backend/internal/tracker/automation_service.go:137-158` |
| FindChannelInProject soft-delete | `backend/internal/tracker/repo_issue_message_link.go:77-84` |
| FilterSubjects | `backend/internal/tracker/automation_consumer.go:27-33` |
| StartTimer trash | `backend/internal/tracker/repo_time.go:72-82` |
| FindIncompleteIssues | `backend/internal/sprint/repository.go:326-332` |
| Retro VersionEQ | `backend/internal/sprint/repository.go:201-209`; `dto.go:26-33` |
| Archive no resurrect | `backend/internal/project/service.go:215-266` |
| RestoreFromTrash only | `backend/internal/project/service.go:334-356` |
| PermanentDelete admin | `backend/internal/project/service.go:358-378` |
| RemoveTeamMember | `backend/internal/project/service.go:624-632` |
| FileIDs / AssignTesters | `backend/internal/tms/service_runs.go:286-341` |
| Allure ExistingRun Mongo/match | `backend/internal/tms/service_import.go:131-166`, `:169-177` |
| TestRail Partial | `backend/internal/tms/testrail_import.go:280`; `importer.go:425` |
| SoftDeleteTable cascade | `backend/internal/bases/repository.go:314+` |
| hubbase trash | `backend/internal/audit/service.go:269-311`; `repository.go:722-762` |
| Dedup map | `A3-DEDUP-2026-07-29.md` TASK-134…142 |

Мутацию prod **не** выполнял (READ-ONLY Stage C). Для «Исправлена» на TASK-134…142 нужна W18-style мутация **не** автором фикса.

---

## 7. Handoff

1. **Не закрывать TASK-008.**
2. Завести отдельный TASK (вне 124…176 или следующий номер) на **A2W1-A5-003** / C-W19-TT-001 — last-member RemoveTeamMember.
3. Рекомендуемый fix/TASK на **C-W19-TT-002:** Allure path — fail-closed на `matchErr` (не `continue` для non-NotFound), не игнорировать `StoreAllureResults` на ExistingRun **или** явный `partial`/warning в ответе (зеркало A2W2-4-004).
4. Перед «Исправлена» TASK-134…142: независимый агент — одна мутация guard на ID → красный hubtest.
5. Синхронизировать README статусы после мутации (C-W19-TT-003).

---

## Verification block

```
verified: false
reason: Stage C red-team READ-ONLY; mutation not run; TASK-008 must stay open
coverage: A3W1-1..5 + A3W2-1..5 → TASK-134…142 map 36/36; A2W1-A5-003 orphan; combos C1–C3 on HEAD
orphans_p0_p1: 0
orphans_other: 1 (A2W1-A5-003 / C-W19-TT-001)
new_task_required: true (TT-001 mandatory; TT-002 recommended P2)
```
