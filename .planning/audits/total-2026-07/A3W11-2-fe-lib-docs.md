# A3 W11 агент 2 — fe lib+docs part 2 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W11-2-fe-lib-docs.md](A2W11-2-fe-lib-docs.md)  
> Stage A: [A11-2-fe-lib-docs.md](A11-2-fe-lib-docs.md)  
> Периметр: 19 файлов (W11-ASSIGNMENTS A2) · Прочитано: 19 · Пропущено: 0  
> Фокус: api/stores secrets · FE↔BE docs/DTO drift

## Сверка с A2 (A2W11-2-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W11-2-001 | RESIDUAL → **FIXED-OK** | закрыт | TASK-103 / ADR-WS-JWT-HANDSHAKE: `auth-store.ts:6-9,149-186` — JWT только в memory, без `persist`/`sessionStorage`; `localStorage` mirror только `{type,userId}` (`:67-82`); тест `does not persist JWT into sessionStorage (A2W11-2-001)` |
| A2W11-2-002 | FIXED-OK (уже в A2) | без регресса | TASK-104 / ADR-PAGINATION: `types.ts:1-5` + `tracker.ts:273-287` cursor list-issues; TZ/wiki cursor |
| W11-A2-001…014 (Stage A via A2) | FIXED-OK | без регресса | см. таблицу FIXED-OK ниже |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 2

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/api/github.ts` | 161 | 0 | FIXED-OK (нет дубля link-preview; secrets только в request DTO) |
| `frontend/src/api/gitlab.ts` | 139 | 0 | чисто (token/secret только request; response без echo) |
| `frontend/src/api/hr.ts` | 82 | 0 | чисто (`version` на approve/reject/cancel) |
| `frontend/src/api/intranet.ts` | 185 | 0 | FIXED-OK `updateNews` + `version` |
| `frontend/src/api/issue-message-links.ts` | 57 | 0 | чисто |
| `frontend/src/api/marketplace.ts` | 152 | 0 | FIXED-OK явные `<T>`; `version` на мутациях |
| `frontend/src/api/messenger.ts` | 613 | 0 | FIXED-OK `image_url`/`domain` + WS guard |
| `frontend/src/api/notifications.ts` | 160 | 0 | чисто (`bind_token` — intentional bind flow) |
| `frontend/src/api/projects.ts` | 160 | 1 | **NEW** — нет `version` после TASK-095 |
| `frontend/src/api/search.ts` | 117 | 0 | чисто |
| `frontend/src/api/sprint.ts` | 199 | 1 | **NEW** — нет `version` после TASK-095 |
| `frontend/src/api/status.ts` | 18 | 0 | чисто |
| `frontend/src/api/tms.ts` | 838 | 0 | FIXED-OK create/deleteEnvironment |
| `frontend/src/api/tracker.ts` | 692 | 0 | чисто (cursor + issue `version`) |
| `frontend/src/api/types.ts` | 20 | 0 | FIXED-OK dual pagination comment |
| `frontend/src/api/webhooks.ts` | 61 | 0 | чисто (secret в request; `WebhookEndpoint` без secret) |
| `frontend/src/stores/ai-store.ts` | 104 | 0 | FIXED-OK loaded/modelsError/rollback |
| `frontend/src/stores/auth-store.ts` | 197 | 0 | FIXED-OK memory JWT (A2W11-2-001) |
| `frontend/src/stores/board-store.ts` | 55 | 0 | чисто (UI state, нет async/secrets) |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W11-2-001 | NEW → **FIXED** | P1 | `projects.ts` + `index.lazy.tsx` | ~~`UpdateProjectRequest` / `ProjectResponse` без `version`~~ | ~~После TASK-095 BE: `version` required…~~ | **FIXED (TASK-130):** `version` в request+response; UI `updateProject(..., { …, version: project.version })`; Vitest `projects-sprint-version.test.ts` |
| A3W11-2-002 | NEW → **FIXED** | P2 | `sprint.ts` | ~~`UpdateSprintRequest` / `SprintResponse` без `version`~~ | ~~BE после TASK-095…~~ | **FIXED (TASK-130):** DTO sync; callers UI edit нет (API-only); tsc fixture omit → `@ts-expect-error` |

### Remediation evidence (TASK-130, 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W11-2-001 | **FIXED** | `frontend/src/api/projects.ts` — `UpdateProjectRequest.version`, `ProjectResponse.version`; `index.lazy.tsx` — `version: project.version`; test `projects-sprint-version.test.ts` |
| A3W11-2-002 | **FIXED** | `frontend/src/api/sprint.ts` — `UpdateSprintRequest.version`, `SprintResponse.version`; fixture `projects-sprint-version-required.fixture.ts` |

## FIXED-OK (сверка A2 + Stage A фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| JWT secrets (A2W11-2-001) | Memory-only Zustand; нет `persist`/`partialize`; sessionStorage `hub-auth` отсутствует; BroadcastChannel + localStorage event **без** токенов; Vitest + E2E asserts |
| Link-preview (W11-A2-001/013) | Канон в `messenger.ts:454-486` (`image_url`/`domain`); `github.ts` без дубля `/link-preview` |
| ai-store (W11-A2-002/003) | `loaded` только success; `botLoadError`/`modelsError`; `saveSelectedModel` rollback |
| auth-store sync (W11-A2-004/011/012) | `clearSessionLocal`; `setUser`/`setTokens` broadcast с `userId`; localStorage mirror |
| tms environments (W11-A2-006) | `createEnvironment` / `deleteEnvironment` |
| intranet/projects members (W11-A2-007/008) | `updateNews`+version; `updateMemberRole` |
| Pagination comment (W11-A2-009 / A2W11-2-002) | `types.ts` cursor для list-issues; offset для admin/intranet |
| WS message guard (W11-A2-010) | `buildMessageFromWsEvent(unknown)` + required id/sender_id → null |
| marketplace generics (W11-A2-014) | явные `<T>` на get/post/put |
| Integration secrets | github/gitlab: `access_token`/`webhook_secret` только в create request; response DTO без echo |
| Webhooks secrets | `secret?` только create/update request; list/get DTO без secret |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 (was 1 → FIXED TASK-130) |
| P2 | 0 (was 1 → FIXED TASK-130) |
| P3 | 0 |
| **Total active** | **0** |

| Метрика | Значение |
|---------|----------|
| Файлов | **19 / 19 / 0** |
| Активных | **0** (NEW 0 · REGRESS 0 · RESIDUAL 0; оба FIXED TASK-130) |
| FIXED-OK (A2 actives) | **1/1** (`A2W11-2-001`) |
| FIXED-OK (Stage A via A2) | **14** без регресса |
| FIXED (A3 NEW → TASK-130) | **2/2** (`A3W11-2-001`, `A3W11-2-002`) |

### Топ-3

1. ~~**A3W11-2-001 (P1, NEW)**~~ → **FIXED** TASK-130 — projects/sprint `version` DTO + UI.
2. ~~**A3W11-2-002 (P2, NEW)**~~ → **FIXED** TASK-130.
3. _(нет P0 / REGRESS)_ — JWT sessionStorage и cursor pagination **FIXED-OK**.

---

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.  
Counts: P0=0 · P1=1 · P2=1 · P3=0 · Total=2.  
Топ: A3W11-2-001.  
