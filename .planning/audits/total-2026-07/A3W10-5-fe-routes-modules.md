# A3 W10 агент A5 — routes+modules part 5 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W10-5-fe-routes-modules.md](A2W10-5-fe-routes-modules.md)  
> Stage A: [A10-5-fe-routes-modules.md](A10-5-fe-routes-modules.md)  
> Assignment: [W10-ASSIGNMENTS.md](W10-ASSIGNMENTS.md) § A5 · Периметр: **28** · Прочитано: **28/28** · Пропущено: **0**  
> Фокус: auth guards · XSS · modules/route stubs · Query error≠empty  
> Remediation: TASK-083 (trash hard-delete) · TASK-093 (isError cases/runs)

## Сверка с A2 (A2W10-A5-*)

| A2 ID | Tag | Статус в HEAD | Доказательство |
|-------|-----|---------------|----------------|
| A2W10-A5-001 | FIXED-OK | закрыт | `trash.tsx:59-70,475-481` — `hardDeleteMutation` → `permanentDeleteProject`; fail → `toastApiError`, без invalidate списка |
| A2W10-A5-002 | FIXED-OK | закрыт | `cases_.$caseKey.lazy.tsx:265-305` — `isLoading` / `isError`+retry / `!tc` разделены |
| A2W10-A5-003 | FIXED-OK | закрыт | `runs_.$runId.tsx:129-177` — `isLoading` / `isError`+retry / `!run` разделены |

**REGRESS:** 0 · **A2 open residual:** 0 · **NEW (этот слот):** 5

### Сверка A1 FIXED-OK cascade (подтверждение на HEAD)

| Stage A ID | Tag | Evidence |
|------------|-----|----------|
| W10-A5-001 | FIXED-OK | `listTrash` queryKey `["projects","trash"]` |
| W10-A5-003 | FIXED-OK | integrations: hint + `Link` `/projects`, нет `projectId=""` |
| W10-A5-004 | FIXED-OK | «Очистить все» отсутствует |
| W10-A5-007 | FIXED-OK | `runs.lazy.tsx:274-290` `isError` + retry |
| W10-A5-008 | FIXED-OK | profile form init в `useEffect` (`profile.lazy.tsx:85-97`) |
| W10-A5-009 | FIXED-OK | `INFO_STEP_BADGE` static (`projects/index.tsx:12-16`) |
| W10-A5-010–012 | FIXED-OK | onboarding: file input, `register("timezone")`, `useAccentColor` |
| W10-A5-013 | FIXED-OK | guest call — узкий `activeCall` shape, без `as CallResponse` |
| W10-A5-014–016 | FIXED-OK | settings Nav без placeholder; security `<Link>`; avatar `onClick` → file |
| W10-A5-018 | FIXED-OK | колонка «Удалено» |
| W10-A5-019 | FIXED-OK | `existingIds` Set до map (`runs_.$runId.tsx:303-306`) |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy.tsx` | 923 | 1 | NEW (authz UI) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx` | 10 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/environments.tsx` | 18 | 0 | чисто (thin → module) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/flaky.tsx` | 13 | 0 | чисто (thin → module) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.lazy.tsx` | 510 | 1 | NEW (authz UI) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.tsx` | 5 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx` | 366 | 1 | NEW (authz UI) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/traceability.tsx` | 19 | 0 | чисто (thin → module) |
| `frontend/src/routes/_authenticated/projects/$projectId/triage.tsx` | 13 | 0 | чисто (thin → module) |
| `frontend/src/routes/_authenticated/projects/index.tsx` | 317 | 0 | FIXED-OK |
| `frontend/src/routes/_authenticated/projects/new.tsx` | 289 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings.tsx` | 118 | 0 | FIXED-OK |
| `frontend/src/routes/_authenticated/settings/ai.tsx` | 10 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/appearance.tsx` | 14 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/hotkeys.tsx` | 66 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/integrations.tsx` | 52 | 0 | FIXED-OK |
| `frontend/src/routes/_authenticated/settings/notifications.tsx` | 20 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/profile.lazy.tsx` | 497 | 0 | FIXED-OK |
| `frontend/src/routes/_authenticated/settings/profile.tsx` | 8 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/settings/security.tsx` | 74 | 0 | FIXED-OK |
| `frontend/src/routes/_authenticated/trash.tsx` | 499 | 0 | FIXED-OK (TASK-083) |
| `frontend/src/routes/call.$code.lazy.tsx` | 327 | 0 | FIXED-OK (guest path) |
| `frontend/src/routes/call.$code.tsx` | 8 | 0 | чисто (stub) |
| `frontend/src/routes/index.tsx` | 12 | 0 | чисто (auth redirect) |
| `frontend/src/routes/login.tsx` | 169 | 1 | NEW (auth UX) |
| `frontend/src/routes/onboarding.lazy.tsx` | 925 | 2 | NEW (auth guard + blob) |
| `frontend/src/routes/onboarding.tsx` | 8 | 0 | чисто (stub; guard отсутствует в stub) |
| `frontend/src/routes/status.tsx` | 19 | 0 | чисто (public + Link login) |

Файлов в списке: 28. Прочитано: 28. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W10-5-001 | NEW | P2 | `cases_.$caseKey.lazy.tsx:392-408` (+ link/unlink `:660+`, `:720`) | «Редактировать» / bind issue всегда видны; `canMutateProject` только на rollback (`:105`) | Класс P2-FE-16: viewer/guest кликают → 403; эталон — `modules/tms/environments.tsx` + webhooks | `const canMutate = canMutateProject(project?.my_role)`; скрыть/disable Edit, link, unlink при `!canMutate` |
| A3W10-5-002 | NEW | P2 | `runs.lazy.tsx:191-247`; `runs_.$runId.tsx:217-227` | Create / Start / Complete / Reopen / Add cases без `canMutateProject` | Та же authz-UI дыра; RunHeader всегда рисует мутации | Прокинуть `canMutate` в list CTA и `RunHeader` (или гейтить handlers/`null` callbacks) |
| A3W10-5-003 | NEW | P2 | `onboarding.tsx:8`; `onboarding.lazy.tsx:923-925` | `/onboarding` вне `_authenticated`, **нет** `beforeLoad` auth | Аноним открывает мастер; API позже 401, но UI/invite flow доступны без сессии | `beforeLoad`: `!isAuthenticated` → `/login`; опц. `onboarding_completed` → `/projects` |
| A3W10-5-004 | NEW | P3 | `login.tsx:164-168` | Нет `beforeLoad` redirect для уже аутентифицированных | Лишний Magic Link UI при валидной сессии (симметрия с `index.tsx:5-10`) | `beforeLoad`: auth → `/projects` или `/onboarding` по флагу |
| A3W10-5-005 | NEW | P3 | `onboarding.lazy.tsx:369-371` | `URL.createObjectURL(file)` без `revokeObjectURL` | Утечка blob URL при смене/уходе со шага | `revoke` в cleanup `useEffect` / перед новым preview / unmount |

## FIXED-OK (сверка A2 + FE-фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| XSS | Во всех 28 файлах **нет** `dangerouslySetInnerHTML` / `innerHTML` / `eval`. User strings (case title, guest name, project name, error) — React text children. Avatar `src` = blob/`/files/…/url` |
| Auth layout | `_authenticated/*` под layout с `beforeLoad` (`_authenticated.tsx:59-63`). Public: `call/$code` (guest intentional), `status`, `login`, `index` redirect |
| Trash P1 | TASK-083: API hard-delete + fail toast; `listTrash`; isError≠empty |
| TMS error≠empty | cases detail, run detail, runs list — error UI + retry |
| Guest call | Unauth → guest form; auth → join+navigate messenger; без full CallResponse cast |
| Modules | environments/flaky/traceability/triage/ai/appearance/notifications — thin wrappers, params OK |
| Stubs | cases_/runs/profile/call/onboarding — intentional lazy stubs |
| `any` / `@ts-ignore` | **0** в слоте |

## Чисто (кратко)

- `projects/new`: key sanitize, mutation+navigate, toast errors.
- `projects/index`: `QueryStateView`, static badges, stable `key={project.id}`.
- `settings/*`: integrations без broken webhooks; security Magic Link copy + TanStack `Link`.
- `profile.lazy`: form sync в effect; avatar file input; sessions revoke.
- `hotkeys` / `status`: статика / public status + login link.

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 2 |
| **Total** | **5** |

| Метрика | Значение |
|---------|----------:|
| Файлов | 28 / 28 / 0 |
| NEW | 5 |
| REGRESS | 0 |
| RESIDUAL (A2 open) | 0 |
| FIXED-OK (A2 actives + A1 cascade) | 3 + cascade |

A2 baseline open: 3 (P1=1 · P2=2) → **все FIXED-OK**.  
A3 open Σ: **5** (P0=0 · P1=0 · P2=3 · P3=2).
