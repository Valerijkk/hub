# W10 агент A5 — routes+modules part 5 (remaining TMS/settings/auth)

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: remaining routes, dead code, `any`/`@ts-ignore`, loading/error, a11y.  
> ID: `W10-A5-001`…  
> Сверка: TZ soft-delete/trash 30d, TanStack Router, Magic Link only, FE Query error/empty.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy.tsx` | 881 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx` | 10 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/environments.tsx` | 18 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/flaky.tsx` | 13 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.lazy.tsx` | 493 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.tsx` | 5 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx` | 331 | 2 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/traceability.tsx` | 19 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/triage.tsx` | 13 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/index.tsx` | 310 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/new.tsx` | 289 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings.tsx` | 132 | 1 | замечания |
| `frontend/src/routes/_authenticated/settings/ai.tsx` | 10 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/appearance.tsx` | 14 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/hotkeys.tsx` | 66 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/integrations.tsx` | 39 | 1 | замечания |
| `frontend/src/routes/_authenticated/settings/notifications.tsx` | 20 | 0 | чисто |
| `frontend/src/routes/_authenticated/settings/profile.lazy.tsx` | 493 | 3 | замечания |
| `frontend/src/routes/_authenticated/settings/profile.tsx` | 8 | 0 | чисто (stub) |
| `frontend/src/routes/_authenticated/settings/security.tsx` | 74 | 1 | замечания |
| `frontend/src/routes/_authenticated/trash.tsx` | 459 | 4 | замечания |
| `frontend/src/routes/call.$code.lazy.tsx` | 329 | 1 | замечания |
| `frontend/src/routes/call.$code.tsx` | 8 | 0 | чисто (stub) |
| `frontend/src/routes/index.tsx` | 12 | 0 | чисто |
| `frontend/src/routes/login.tsx` | 169 | 0 | чисто |
| `frontend/src/routes/onboarding.lazy.tsx` | 859 | 3 | замечания |
| `frontend/src/routes/onboarding.tsx` | 8 | 0 | чисто (stub) |
| `frontend/src/routes/status.tsx` | 19 | 0 | чисто |

Файлов в списке: 28. Прочитано: 28. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W10-A5-001 | P1 | `trash.tsx:36-43` | Корзина грузит `listProjects()` (`GET /projects`) и фильтрует `status === "deleted" \| "archived"` | Backend `ListProjects` **исключает** deleted; archived только с `include_archived`. Канон — `GET /api/projects/trash` (`ListTrash`). FE-клиента нет → корзина всегда пуста для soft-deleted | Добавить `listTrash()` → `/projects/trash`; убрать фильтр по обычному list; archived — отдельно или через query |
| W10-A5-002 | P1 | `trash.tsx:439-444` | «Удалить навсегда» только `toast.success` + закрытие диалога — **нет API-вызова** | Пользователь думает, что данные уничтожены; проект остаётся в корзине до auto-purge. Нарушает ожидание TZ hard-delete | Вызвать admin/permanent endpoint (если доступен роли) **или** убрать кнопку; не показывать успех без backend |
| W10-A5-003 | P1 | `integrations.tsx:29-31` | `<WebhooksPage projectId="" />` — комментарий «нет projectId» | `listWebhooks`/`createWebhook` бьют в `/projects/${projectId}/webhooks` → `/projects//webhooks`. Секция вебхуков в Settings **сломанная** (dead/broken UI) | Убрать Webhooks из глобальных settings **или** требовать выбор проекта; webhooks только в project settings |
| W10-A5-004 | P2 | `trash.tsx:95-100` | Кнопка «Очистить все» без `onClick` | Dead UI: клик ничего не делает | Реализовать bulk-purge через API **или** удалить кнопку |
| W10-A5-005 | P2 | `cases_.$caseKey.lazy.tsx:252-272` | `if (isLoading \|\| !tc)` → shimmer; `isError` не различается | 404/сеть → бесконечный скелетон вместо error/empty | `isError` → error UI + retry; `!isLoading && !tc` → not found |
| W10-A5-006 | P2 | `runs_.$runId.tsx:124-143` | То же: `isLoading \|\| !run` без `isError` | Ошибка прогона → вечный shimmer | Как W10-A5-005 |
| W10-A5-007 | P2 | `runs.lazy.tsx:104-107`; `:254` | `useQuery` без ветки `isError` — при fail `runs=[]` → empty «Нет тестовых прогонов» | Путает сеть/403 с «нет данных» | `QueryStateView` / error empty + retry |
| W10-A5-008 | P2 | `profile.lazy.tsx:86-98` | Синхронизация формы через `setState` **во время render** (`if (user && !initialized) { setForm… }`) | Anti-pattern React 18/19; риск лишних re-render/StrictMode-сюрпризов | `useEffect([user])` или key-remount / `useQuery` `placeholderData` + reset |
| W10-A5-009 | P2 | `projects/index.tsx:121` | Динамические классы `bg-${s.color}-500/10 text-${s.color}-500` | Tailwind JIT **не** видит динамические имена → стили info-steps мёртвые | Статический map `blue→"bg-blue-500/10…"` |
| W10-A5-010 | P2 | `onboarding.lazy.tsx:411-428` | Блок «загрузка аватара» — нет `input[type=file]` / `onClick` | Dead UI: клик по кругу ничего не делает | File input + upload как в profile **или** убрать блок |
| W10-A5-011 | P2 | `onboarding.lazy.tsx:512-524` | `<select id="onb-timezone">` не в `react-hook-form`, значение не уходит в `completeOnboarding` | Dead control: timezone не сохраняется | `register("timezone")` + поле в API payload |
| W10-A5-012 | P2 | `onboarding.lazy.tsx:202-206`; `:663-678` | `accent` / `accentColors` — только локальный state, никуда не пишется | Dead feature: выбор accent визуально «работает», эффекта нет | Persist в appearance/store/CSS vars **или** убрать UI |
| W10-A5-013 | P2 | `call.$code.lazy.tsx:102-112` | `as import("@/api/calls").CallResponse` на неполном объекте (`channel_id:""`, `created_by:""`) | Type-lie; CallRoom/controls могут упасть на пустых полях | Отдельный GuestCall тип / узкий store shape без полного `CallResponse` |
| W10-A5-014 | P3 | `settings.tsx:90-114` | `NavItem` ветка `placeholder` — ни один `navItems` не передаёт `placeholder` | Dead code path | Удалить ветку **или** использовать для «скоро» |
| W10-A5-015 | P3 | `security.tsx:58-63` | Ссылка на профиль через `<a href="/settings/profile">` | Full page reload вместо TanStack `Link`; ломает SPA/scroll restore | `<Link to="/settings/profile">` |
| W10-A5-016 | P3 | `profile.lazy.tsx:162-177` | Аватар `cursor-pointer` + overlay, но клик не открывает file input (только кнопка) | Misleading a11y/UX | `onClick` на круг → `fileInputRef.click()` |
| W10-A5-017 | P3 | `profile.lazy.tsx:82`; `:42-51` | `statusDuration` = `useState("1d")` без setter; `department` в form/submit, **поля в UI нет** | Dead/half-wired state | Убрать department из state **или** добавить input; duration — константа / select |
| W10-A5-018 | P3 | `trash.tsx:160-214` | Колонка «Удалил» показывает `updated_at` (дату), не пользователя | Misleading label / dead column semantics | Переименовать в «Удалено» **или** показывать actor, если API отдаёт |
| W10-A5-019 | P3 | `runs_.$runId.tsx:270-273` | `new Set(…)` пересоздаётся **внутри** `.map` на каждый кейс | Лишняя O(n²) работа в диалоге | Вынести `existingIds` до map |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **`any` / `@ts-ignore` / `@ts-expect-error`** | Во всех 28 файлах зоны — **0** вхождений (в т.ч. settings/*, tms routes, login/onboarding/call) |
| **Rules of Hooks** | Хуки до early-return в cases/runs/profile/call; нарушений порядка нет (кроме setState-in-render в profile — W10-A5-008) |
| `cases_.$caseKey.tsx`, `runs.tsx`, `profile.tsx`, `call.$code.tsx`, `onboarding.tsx` | Намеренные TanStack code-split stubs — не dead |
| `environments.tsx`, `flaky.tsx`, `traceability.tsx`, `triage.tsx` | Thin wrappers → modules; params проброшены |
| `ai.tsx`, `appearance.tsx`, `notifications.tsx`, `status.tsx` | Thin wrappers; status + Link на login OK |
| `hotkeys.tsx` | Статическая справка; keys/labels; нет `any` |
| `new.tsx` | Валидация key, i18n toast, mutation+navigate; `generateKey`/`FromInitials` используются |
| `index.tsx` (root) | `beforeLoad` redirect auth → projects / login |
| `login.tsx` | Magic Link + invite search; DevLoginPanel; success → onboarding/projects |
| Optimistic lock | `cases_.$caseKey.lazy` передаёт `version: tc.version` в update — OK |

## Counts

| Метрика | Значение |
|---------|----------|
| Файлов | 28 |
| Находок всего | **19** |
| P0 | 0 |
| P1 | 3 |
| P2 | 10 |
| P3 | 6 |
| Файлов с находками | 12 |
| Файлов чисто | 16 |
| `any` / `@ts-ignore` | **0** |
| Dead / broken UI (явные) | trash list+purge+clear-all; settings webhooks; onboarding avatar/tz/accent |

## Топ-3

1. **W10-A5-001 (P1)** — корзина `/trash` читает `listProjects`, а deleted живут в `GET /projects/trash` → soft-deleted проекты **никогда не видны** (`trash.tsx:36-43`).
2. **W10-A5-003 (P1)** — Settings → Интеграции монтирует `WebhooksPage` с `projectId=""` → запросы на `/projects//webhooks` (`integrations.tsx:29-31`).
3. **W10-A5-002 (P1)** — «Удалить навсегда» в корзине — fake success без API (`trash.tsx:439-444`).

**Фокус-вердикт:** stubs/lazy-split и thin TMS wrappers в порядке; `any` нет. Главный риск зоны — **битая корзина** (неверный endpoint + fake hard-delete + мёртвая «Очистить все») и **мёртвые вебхуки в глобальных Settings**. Второстепенно: infinite shimmer на error (TMS detail), dead onboarding controls (avatar/tz/accent), setState-in-render в profile.
