# AUDIT BE v4 — crm / hr / intranet / calendar

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`) · branch `cursor-grok-4.5`  
> Модель: `cursor-grok-4.5`  
> Периметр: **22** production `.go` (+ **4** `export_hooks.go` hubtest-only, просмотрены, вне prod-счёта)  
> Прочитано: **22/22** · Пропущено: **0**  
> Фокус: IDOR · soft-delete · VersionEQ · AppError/`%w` wrap · RBAC  
> Baseline: [A3W7-2-v4-business.md](A3W7-2-v4-business.md) · remediation TASK-163 (A3W7-2-001…003)  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

## Вердикт

Модули CRM / HR / intranet (после TASK-163) в целом держат owner-scope, VersionEQ и soft-delete на основных путях. **Новые открытые дефекты:** calendar отдаёт soft-deleted leave и sprint; poll GET/vote не проверяет `published` (обход draft-гейта `GetNews`). A3W7-2-001…003 в HEAD **FIXED-OK**.

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 2 |
| **Total open** | **5** |

## Сверка A3W7-2 (TASK-163)

| A3 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A3W7-2-001 | FIXED-OK | закрыт | `notifyCoffeeMatch` → `slog.WarnContext` (`coffee_service.go:160–171`); вызовы из `RunCoffeeMatching` |
| A3W7-2-002 | FIXED-OK | закрыт | `req.Version < 1` → `ErrInvalidInput`; repo `newspoll.VersionEQ` (`service_polls.go:76–78`, `repo_polls.go:221–224`) |
| A3W7-2-003 | FIXED-OK | закрыт | `coffeePartnersForUser` + `Partners[]`; spoke через hub (`coffee_service.go:60–158`); DTO `Partners` |

## Прочитанные файлы

### crm (4 prod)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/crm/dto.go` | 170 | 0 | чисто |
| `backend/internal/crm/handler.go` | 194 | 0 | authMw; mapServiceError 404/403/409 |
| `backend/internal/crm/repository.go` | 248 | 0 | VersionEQ + DeletedAtIsNil + MaxCRMList=100 |
| `backend/internal/crm/service.go` | 328 | 1 | owner/admin write; RT-003 list; P3 version&lt;1→409 |
| `backend/internal/crm/export_hooks.go` | 16 | 0 | hubtest-only |

### hr (4 prod)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/hr/dto.go` | 100 | 0 | чисто |
| `backend/internal/hr/handler.go` | 165 | 0 | ErrAccessDenied≠ErrForbidden mapping |
| `backend/internal/hr/repository.go` | 214 | 0 | Version WHERE; DeletedAtIsNil; SoftDeleteLeave; leave scope |
| `backend/internal/hr/service.go` | 251 | 0 | admin approve/reject; owner cancel; slog notify/export |
| `backend/internal/hr/export_hooks.go` | 33 | 0 | hubtest-only |

### intranet (9 prod)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/intranet/coffee_dto.go` | 64 | 0 | Partners[]; без email |
| `backend/internal/intranet/coffee_repository.go` | 220 | 0 | round+pairs tx; FindUserPairsInRound All |
| `backend/internal/intranet/coffee_service.go` | 397 | 0 | notify slog FIXED; triples FIXED; admin gates |
| `backend/internal/intranet/dto.go` | 217 | 0 | poll Version обязателен (док) |
| `backend/internal/intranet/handler.go` | 526 | 0 | poll routes; layering OK |
| `backend/internal/intranet/repository.go` | 242 | 0 | news VersionEQ; SoftDeleteKudo; DeletedAtIsNil |
| `backend/internal/intranet/repo_polls.go` | 258 | 0 | poll VersionEQ; soft-delete hide; vote tx |
| `backend/internal/intranet/service.go` | 347 | 1 | author\|\|admin delete; P3 kudo notify fail-after-write |
| `backend/internal/intranet/service_polls.go` | 204 | 1 | VersionEQ FIXED; **draft poll IDOR** |
| `backend/internal/intranet/export_hooks.go` | 39 | 0 | hubtest-only |

### calendar (5 prod)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/calendar/dto.go` | 38 | 0 | чисто |
| `backend/internal/calendar/handler.go` | 108 | 0 | events+ics через ListEvents |
| `backend/internal/calendar/ical.go` | 58 | 0 | escapeICal OK |
| `backend/internal/calendar/repository.go` | 126 | 2 | **leave/sprint без DeletedAtIsNil**; issues/calls OK |
| `backend/internal/calendar/service.go` | 238 | 0 | projectIDs scope; leave title без email |
| `backend/internal/calendar/export_hooks.go` | 30 | 0 | hubtest-only |

Файлов prod: **22**. Прочитано: **22**. hubtest hooks: **4** (не в prod-счёте).

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| BE-V4-001 | NEW | P2 | `calendar/repository.go:116–125` | `ListApprovedLeaveInRange` **без** `leaverequest.DeletedAtIsNil()` | HR-репо фильтрует trash; calendar после GDPR soft-delete leave продолжает показывать отсутствие в `/events` и `.ics` (PII + TZ soft-delete) | Добавить `leaverequest.DeletedAtIsNil()` как в `hr/repository.go:92`; IT: soft-delete leave → calendar empty; mutation: снять фильтр → FAIL |
| BE-V4-002 | NEW | P2 | `calendar/repository.go:74–87` | `ListSprintsInRange` **без** `sprintEnt.DeletedAtIsNil()` | У `Sprint` есть `deleted_at` (schema); sprint-модуль везде фильтрует trash; calendar отдаёт soft-deleted спринты в UI/iCal | `Where(..., sprintEnt.DeletedAtIsNil())`; IT + source guard / mutation |
| BE-V4-003 | NEW | P2 | `intranet/service_polls.go:14–61` | `GetNewsPoll` / `VoteNewsPoll` не проверяют `post.Published` | `GetNews` для member: unpublished → `ErrNotFound`; poll/vote по UUID обходят гейт — чтение вопроса/опций и голос по черновику | Как GetNews: если `!post.Published` и не admin → `ErrNotFound`; IT draft UUID member 404; author/admin OK |
| BE-V4-004 | NEW | P3 | `crm/service.go:91–92`, `189–190` | `version < 1` → `ErrVersionConflict` (409) | News/HR/poll: omit/0 → **400** `ErrInvalidInput`; CRM → **409** — drift контракта optimistic lock | `return nil, ErrInvalidInput` (или отдельный sentinel→400); handler уже мапит InvalidInput |
| BE-V4-005 | NEW | P3 | `intranet/service.go:331–334` | `CreateKudo`: fail `NotifyKudosReceived` → `return error` после успешного Save | Клиент получает 500, kudo уже в БД; coffee/HR — best-effort + `slog.WarnContext` | Как coffee: `slog.WarnContext` + вернуть resp; не откатывать UX после commit |

## Чисто (подтверждено)

### IDOR / RBAC
- **CRM:** `authorizeContactWrite` / `authorizeDealWrite` (owner \|\| system admin/founder); `listOwnerScope` RT-003; `validateDealRefs` → contact owner-scope (не чужой UUID на сделку).
- **HR:** pending/approve/reject — `requireAdmin`; cancel — только requester (`ErrAccessDenied`); `ListApprovedLeaveInRange` viewer+shared projects.
- **Intranet news:** create/update/admin-list — admin; delete — author\|\|admin; draft GET — 404 для member.
- **Coffee:** enrollment/match — self; stats/rounds/run — admin.
- **Calendar:** только проекты membership; leave не org-wide (A6-005) — **кроме** trash-gap BE-V4-001.

### Soft-delete
- CRM contact/deal: `DeletedAtIsNil` на get/list/update; SoftDelete* после authorize.
- HR leave: все query + SoftDeleteLeave; Version WHERE на approve/reject/cancel.
- Intranet news: SoftDeleteNewsPost + audit tx; poll hide через `HasNewsPostWith(DeletedAtIsNil)`; SoftDeleteKudo.
- Calendar calls/issues: `DeletedAtIsNil` OK; **sprint/leave — BE-V4-001/002**.

### VersionEQ
- CRM contact/deal: `VersionEQ` + `AddVersion(1)`.
- HR: `leaverequest.Version(version)` в WHERE approve/reject/cancel.
- News post + NewsPoll: `VersionEQ`; service `version < 1` → 400.
- Calendar: read-only агрегатор — OL N/A.

### AppError / wrap
- Sentinel + `errors.Is` в handlers → `fiber.NewError` (контракт ErrorHandler → `{"error","request_id"}`).
- Repo/service: `fmt.Errorf("…: %w", err)` на infra-путях; known → sentinel без утечки ent.
- CRM list/pipeline: всегда 500 на неизвестное (mapServiceError не нужен — нет sentinel с list).
- Отклонение: BE-V4-005 (error после успешной записи).

## Матрица фокуса (сводка)

| Модуль | IDOR/RBAC | Soft-delete | VersionEQ | AppError wrap |
|--------|-----------|-------------|-----------|---------------|
| crm | OK | OK | OK (P3 status code) | OK |
| hr | OK | OK (repo) | OK | OK (slog) |
| intranet | **P2 draft poll** | OK news/kudo | OK post+poll | **P3 kudo notify** |
| calendar | OK scope | **P2 leave+sprint** | N/A | OK |

## Рекомендации remediation (не исполнять в этом слоте)

1. **P2 calendar soft-delete** — один коммит `fix(calendar): …` + IT в `testframework/calendar/`.
2. **P2 draft poll** — `fix(intranet): …` + IT vote/get unpublished.
3. **P3** CRM version 400 + kudo notify best-effort — опционально hygiene / follow-up TASK.

## Статус TASK-008

- Этот слот: **READ-ONLY audit complete**.
- **TASK-008 не закрывать** (программа тотального аудита; нет mutation/COV gate).
- Коммит отчёта — по решению owner/team-lead; агент **не** коммитит.

---

*File-auditor · 2026-07-29 · `617af223` · RU*
