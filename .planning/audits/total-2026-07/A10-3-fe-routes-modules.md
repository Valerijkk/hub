# W10 агент A3 — routes+modules part 3

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус FE: modules state · TanStack Query · a11y forms  
> Список: W10-ASSIGNMENTS.md §A3 (29 файлов).

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/modules/tracker/issues-cursor.ts | 24 | 0 | чисто |
| frontend/src/modules/tracker/issue-sidebar-meta.tsx | 254 | 2 | замечания |
| frontend/src/modules/tracker/pipeline-status.tsx | 296 | 1 | замечания |
| frontend/src/modules/tracker/planning-poker.tsx | 210 | 3 | замечания |
| frontend/src/modules/tracker/sprint-backlog.tsx | 177 | 3 | замечания |
| frontend/src/modules/tracker/sprint-board.tsx | 385 | 2 | замечания |
| frontend/src/modules/tracker/sprint-retro.tsx | 166 | 2 | замечания |
| frontend/src/modules/tracker/triage-inbox.tsx | 150 | 1 | замечания |
| frontend/src/modules/tracker/velocity-chart.tsx | 108 | 1 | замечания |
| frontend/src/routes/__root.tsx | 111 | 0 | чисто |
| frontend/src/routes/_authenticated.tsx | 64 | 0 | чисто |
| frontend/src/routes/_authenticated/accounting.lazy.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/accounting.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/audit.tsx | 23 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/dashboard.tsx | 213 | 2 | замечания |
| frontend/src/routes/_authenticated/admin/limits.tsx | 23 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/mail.tsx | 22 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/trash.tsx | 23 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/users.tsx | 23 | 0 | чисто |
| frontend/src/routes/_authenticated/bases.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/calendar.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/calls.tsx | 188 | 2 | замечания |
| frontend/src/routes/_authenticated/crm.lazy.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/crm.tsx | 6 | 0 | чисто |
| frontend/src/routes/_authenticated/dashboard.lazy.tsx | 241 | 2 | замечания |
| frontend/src/routes/_authenticated/dashboard.tsx | 3 | 0 | чисто |
| frontend/src/routes/_authenticated/help.tsx | 96 | 0 | чисто |
| frontend/src/routes/_authenticated/hr.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/intranet/directory.tsx | 19 | 0 | чисто |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W10-A3-001 | P1 | frontend/src/modules/tracker/sprint-backlog.tsx:28 | `listIssues(projectId, { limit: 100 })` без фильтра по спринту; UI пишет «задач без спринта», а `ListIssuesParams`/`IssueResponse` не содержат `sprint_id` | Бэклог показывает произвольные 100 задач проекта (в т.ч. уже в спринтах); после `moveIssueToSprint` они снова появляются в списке | API: фильтр `sprint_id=null` (или отдельный backlog endpoint) + поле в DTO; FE: `useQuery` с этим фильтром и client-side guard |
| W10-A3-002 | P1 | frontend/src/modules/tracker/planning-poker.tsx:101 | Пока `pokerSession === null` (до первого ответа / при ошибке) рендерится «Сессия не найдена или истекла» | Нет разделения loading / error / empty — ложный empty на каждый mount; ошибка сети неотличима от истечения | `loading`/`error` state (или `useQuery`); скелетон до ответа; empty только после fail/404 |
| W10-A3-003 | P2 | frontend/src/modules/tracker/velocity-chart.tsx:32 | `velocity.length === 0` сразу → «Нет завершённых спринтов»; `fetchVelocity` в store без loading/error | Пока идёт запрос (и при fail) UI врёт про empty | loading/error в store или `useQuery`; empty только при успехе с `[]` |
| W10-A3-004 | P2 | frontend/src/modules/tracker/sprint-backlog.tsx:30 | `catch` глотает ошибку; `loading=false` → empty «Бэклог пуст» | Fail API неотличим от реального пустого бэклога | error state + retry; toast / `QueryStateView` |
| W10-A3-005 | P2 | frontend/src/modules/tracker/issue-sidebar-meta.tsx:57 | Подписи «Статус» / «Приоритет» / «Story Points» — `<div>`, не связаны с `Select` (`htmlFor` / `aria-labelledby`) | a11y forms: AT не озвучивает назначение контролов | `Label htmlFor` + id на Trigger, или `aria-labelledby` на SelectTrigger |
| W10-A3-006 | P2 | frontend/src/modules/tracker/sprint-backlog.tsx:130 | Checkbox с `id`, но без `<label htmlFor>` / `aria-label` | Контрол выбора задачи без accessible name | `aria-label={`Выбрать ${issue.issue_key}`}` или label вокруг строки |
| W10-A3-007 | P2 | frontend/src/modules/tracker/sprint-board.tsx:240 | `<Button>` внутри `<Link>` (Начать/Завершить спринт) | Nested interactive: невалидный HTML, конфликт фокуса/клика/AT | Вынести действия вне Link (stopPropagation уже есть — структура всё равно button⊂a) или карточка-div + отдельный Link на заголовок |
| W10-A3-008 | P2 | frontend/src/modules/tracker/planning-poker.tsx:126 | Карточки голоса — `<button>` без `aria-pressed` / имени роли группы | a11y: выбранное значение и radio-group семантика недоступны AT | `role="radiogroup"` + `aria-pressed`/`aria-checked` на выбранной карточке |
| W10-A3-009 | P2 | frontend/src/modules/tracker/planning-poker.tsx:45 | Polling `setInterval(loadSession, 3000)` + zustand вместо Query; ошибки `catch {}` | Дублирование state (store vs server); silent fail на vote/reveal/load | `useQuery` + `refetchInterval`; toast на ошибках; invalidate после vote |
| W10-A3-010 | P2 | frontend/src/modules/tracker/triage-inbox.tsx:60 | `presetBootstrapDone.current = true` **до** успеха `createPreset.mutate()` | При fail мутации preset больше не создаётся до remount | Ставить flag в `onSuccess`; на error сбрасывать / retry |
| W10-A3-011 | P2 | frontend/src/modules/tracker/sprint-retro.tsx:51 | `handleSave` без catch UI; fail → только `saving=false` | Пользователь не видит ошибку сохранения ретро | try/catch + toast; не показывать «Сохранено» при fail |
| W10-A3-012 | P2 | frontend/src/routes/_authenticated/admin/dashboard.tsx:78 | Вкладки — кастомные `<button>` внутри `Tabs`, не `TabsTrigger`; нет `role="tab"` / `aria-selected` | a11y: табы не объявляют selected state (Radix TabsList обойдён) | Заменить на `TabsList`/`TabsTrigger` или добавить ARIA вручную |
| W10-A3-013 | P2 | frontend/src/routes/_authenticated/admin/dashboard.tsx:102 | KPI `change={+5/+2/-3}` захардкожены | Ложные тренды в админ-дашборде | Убрать change или брать с API |
| W10-A3-014 | P2 | frontend/src/routes/_authenticated/calls.tsx:117 | Классы `bg-${s.color}-500/10 text-${s.color}-500` — динамическая конкатенация | Tailwind JIT не генерирует такие классы → шаги «Как начать» без цветов | Полные class strings в объекте (`bg-blue-500/10` и т.д.) |
| W10-A3-015 | P2 | frontend/src/routes/_authenticated/calls.tsx:19 | `startInstantCall` → falsy `inviteCode`: нет toast/error | Тихий fail создания звонка | else-ветка с `toast.error` |
| W10-A3-016 | P2 | frontend/src/routes/_authenticated/dashboard.lazy.tsx:95 | Icon-only Search / Bell / «Поиск» quick-action без `onClick` (Bell ещё без `aria-label`) | Мёртвый UI + слабая a11y | Привязать к command palette / notifications или убрать; `aria-label` на Bell |
| W10-A3-017 | P3 | frontend/src/modules/tracker/pipeline-status.tsx:200 | `refetchInterval: 30_000` всегда, даже без `in_progress` | Лишние запросы CI | Interval только если есть run со status queued/in_progress |
| W10-A3-018 | P3 | frontend/src/modules/tracker/sprint-retro.tsx:69 | `setTimeout(() => setSaved(false), 2000)` без cleanup | Unmount → setState на размонтированном | clearTimeout в cleanup / ref |
| W10-A3-019 | P3 | frontend/src/modules/tracker/issue-sidebar-meta.tsx:121 | Блоки исполнитель/автор с `cursor-pointer` без handler | Ложный affordance клика | Убрать cursor-pointer или открыть picker |
| W10-A3-020 | P3 | frontend/src/routes/_authenticated/dashboard.lazy.tsx:71 | `useMemo(() => formatTime(), [])` — время застывает до remount | Устаревшие часы в шапке | Убрать memo / обновлять по interval |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 14 |
| P3 | 4 |
| **Σ** | **20** |

### Топ-3
1. **W10-A3-001** — SprintBacklog не фильтрует задачи без спринта → бэклог врёт и ломает move-to-sprint UX.
2. **W10-A3-002** — Planning Poker показывает «сессия истекла» на каждый loading/error.
3. **W10-A3-003** — VelocityChart путает loading/fail с empty (тот же класс state-бага, zustand без Query).

## Чисто
Проверено: modules state / Query keys+invalidation / a11y forms / loading-empty-error / any·ts-ignore / useEffect leaks.

| Файл | Что проверено |
|------|----------------|
| issues-cursor.ts | Чистые хелперы flatten/nextCursor (фикс A10-002); state/UI нет |
| __root.tsx | QueryClient defaults (retry:1, staleTime:30s); error/404 UI; Toaster |
| _authenticated.tsx | Auth gate; Query clear + navigate на multi-tab logout/login; SW register |
| accounting.tsx / .lazy.tsx | Admin gate + lazy page + RouteErrorBoundary |
| admin/audit\|limits\|mail\|trash\|users.tsx | System-admin beforeLoad + тонкие обёртки панелей |
| bases.tsx / calendar.tsx / hr.tsx | ErrorBoundary + module page |
| crm.tsx / crm.lazy.tsx / dashboard.tsx | Eager stub + lazy code-split (A13-017) |
| help.tsx | i18n; aria-hidden на иконках; admin link gated |
| intranet/directory.tsx | validateSearch `user`; OrgDirectory props |

**any / @ts-ignore / @ts-expect-error:** в 29 файлах не найдено.  
**Хорошие паттерны:** `triage-inbox` + `QueryStateView` (loading/error/empty); `pipeline-status` Query + skeleton/error/empty; `_authenticated` Query clear на auth-sync; lazy+ErrorBoundary на accounting/crm/dashboard/calls/bases/calendar/hr.
