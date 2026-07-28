# W9 агент A5 — components part 5 (ui + user-picker)

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: hooks, a11y, `any`/`ts-ignore`, ui primitives (Shadcn className-only, Radix).  
> ID: `W9-A5-001`…  
> Сверка: style guide (`type="button"`), FE-H4 ErrorBoundary, NFR-16 a11y, TZ Shadcn className-only.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/components/tracker/user-picker.tsx` | 132 | 3 | замечания |
| `frontend/src/components/ui/alert-dialog.tsx` | 140 | 0 | чисто |
| `frontend/src/components/ui/avatar.tsx` | 57 | 1 | замечания |
| `frontend/src/components/ui/badge.tsx` | 37 | 1 | замечания |
| `frontend/src/components/ui/button.tsx` | 57 | 1 | замечания |
| `frontend/src/components/ui/card.tsx` | 86 | 1 | замечания |
| `frontend/src/components/ui/command.tsx` | 136 | 2 | замечания |
| `frontend/src/components/ui/confirm-dialog.tsx` | 88 | 2 | замечания |
| `frontend/src/components/ui/dialog.tsx` | 120 | 1 | замечания |
| `frontend/src/components/ui/dropdown-menu.tsx` | 198 | 1 | замечания |
| `frontend/src/components/ui/empty-state.tsx` | 34 | 1 | замечания |
| `frontend/src/components/ui/error-boundary.tsx` | 53 | 2 | замечания |
| `frontend/src/components/ui/form-field.tsx` | 28 | 2 | замечания |
| `frontend/src/components/ui/input.tsx` | 22 | 0 | чисто |
| `frontend/src/components/ui/label.tsx` | 24 | 0 | чисто |
| `frontend/src/components/ui/popover.tsx` | 31 | 0 | чисто |
| `frontend/src/components/ui/scroll-area.tsx` | 46 | 0 | чисто |
| `frontend/src/components/ui/select.tsx` | 157 | 1 | замечания |
| `frontend/src/components/ui/separator.tsx` | 29 | 0 | чисто |
| `frontend/src/components/ui/sheet.tsx` | 124 | 0 | чисто |
| `frontend/src/components/ui/tabs.tsx` | 55 | 0 | чисто |
| `frontend/src/components/ui/textarea.tsx` | 22 | 0 | чисто |
| `frontend/src/components/ui/tooltip.tsx` | 30 | 0 | чисто |

Файлов в списке: 23. Прочитано: 23. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W9-A5-001 | P1 | `error-boundary.tsx:18-52`; использование `app-shell.tsx:154-156` | `ErrorBoundary` после ошибки держит `hasError: true` навсегда; retry только через `window.location.reload()`. Sidebar/`Outlet` смены маршрута **не** сбрасывают state | Пользователь кликает другой раздел в sidebar (вне boundary) — контент остаётся на «Что-то пошло не так» до hard reload. Слабее `RouteErrorBoundary` (там есть `reset`) | `componentDidUpdate`: сброс при смене `children`/`location`; кнопка soft-reset `setState({hasError:false})`; либо deprecate в пользу `RouteErrorBoundary` |
| W9-A5-002 | P2 | `button.tsx:43-52` | `Button` не задаёт `type="button"` по умолчанию; HTML default для `<button>` в форме — `submit` | Style guide запрещает implicit submit. Примитив используется в ~80 файлах; кнопка без `type` внутри `<form>` случайно сабмитит форму | `type={props.type ?? (asChild ? undefined : "button")}` (не трогать Radix Slot path) |
| W9-A5-003 | P2 | `confirm-dialog.tsx:74-79` | `onClick` → `e.preventDefault()` + `Promise.resolve(onConfirm()).finally(() => onOpenChange(false))` — диалог **всегда** закрывается, в том числе при reject | Ошибка удаления/reject не видна в диалоге: UI закрылся, toast может быть проглочен/опоздать; повторный confirm невозможен без reopen | `try/await onConfirm(); onOpenChange(false)` в `catch` оставить open + опциональный `onError`; не закрывать в `finally` |
| W9-A5-004 | P2 | `confirm-dialog.tsx:59-61` | `description` опционален: при отсутствии не рендерится `AlertDialogDescription` | Radix AlertDialog ожидает Description (или явный `aria-describedby={undefined}`); иначе a11y-warning / диалог без объяснения для screen reader | Всегда рендерить Description (пустой → `sr-only` fallback) **или** на Content `aria-describedby={undefined}` |
| W9-A5-005 | P2 | `command.tsx:34-41` | `CommandDialog`: `aria-describedby={undefined}` **и** одновременно `DialogDescription` (sr-only) | Описание в DOM есть, но association с Content снята — screen reader не получает dialog description; мёртвый a11y-код | Убрать `aria-describedby={undefined}` **или** убрать `DialogDescription`; оставить одну каноническую связь |
| W9-A5-006 | P2 | `user-picker.tsx:45-49` | `useQuery` без `isError` / retry UI; при fail список пуст → `CommandEmpty` «нет результатов» | Пользователь путает сетевую ошибку с «нет участников»; assignee нельзя выбрать | Показать error empty + retry; не смешивать с `userPicker.empty` |
| W9-A5-007 | P2 | `user-picker.tsx:45-49` | Нет `enabled: Boolean(projectId)` | Пустой/`undefined` projectId → лишний запрос `listMembers` и шум в Network/ошибках | `enabled: !!projectId` |
| W9-A5-008 | P2 | `form-field.tsx:17-25` | `hint` рендерится рядом, но не связан с control через `aria-describedby` | Hint недоступен AT как описание поля (WCAG 1.3.1 / 3.3.2) | `id` на hint + clone/проп `aria-describedby` на children или wrapper API |
| W9-A5-009 | P3 | `dialog.tsx:46-47` | Close: `sr-only` текст **«Close»** (EN); в `sheet.tsx:82` — «Закрыть» | Несогласованность i18n/RU UI; EN-only для RU-аудитории | i18n ключ / единый «Закрыть» |
| W9-A5-010 | P3 | `command.tsx:38-41` | Hardcoded RU «Глобальный поиск» / описание в `CommandDialog` | Обходит i18n-слой; EN locale сломан | `useTranslation` / ключи `search.*` |
| W9-A5-011 | P3 | `error-boundary.tsx:32-46` | Fallback и кнопка hardcoded RU; нет soft-reset API | Дублирует/расходится с `RouteErrorBoundary`; нет i18n | Вынести тексты в i18n; API `reset` как у route EB |
| W9-A5-012 | P3 | `avatar.tsx:21-24` | Online-dot всегда `aria-hidden="true"` | Статус online не объявляется AT (только visual) | `aria-label` на Root при `data-online` / visually hidden text |
| W9-A5-013 | P3 | `card.tsx:32-45` | `CardTitle` — `<div>`, не heading | Нет семантики заголовка карточки для AT / outline | `h2`/`h3` или `role="heading" aria-level` |
| W9-A5-014 | P3 | `badge.tsx:31-34` | `Badge` = `<div>` с `focus:ring-*`, но не focusable | Focus-стили мёртвые; status-badge без `role` | `span` + при необходимости `role="status"`; убрать focus ring или сделать interactive |
| W9-A5-015 | P3 | `empty-state.tsx:23-26` | Decorative `Icon` без `aria-hidden` | Лишний шум для screen reader | `aria-hidden` на Icon / wrapper |
| W9-A5-016 | P3 | `dropdown-menu.tsx:35`; `:108`; `:131` | ChevronRight / Check / Circle без `aria-hidden` | Декоративные иконки читаются AT | `aria-hidden` на lucide |
| W9-A5-017 | P3 | `user-picker.tsx:74`; `:100-105` | `ChevronsUpDown` / `Check` без `aria-hidden` (у `UserRound` — есть) | Несогласованный a11y в том же компоненте | `aria-hidden` на декоративных svg |
| W9-A5-018 | P3 | `select.tsx:20`; `:76` | Токены `rounded-md` / `border-input` / `bg-popover` — shadcn defaults, не `--hub-*` / `border-b1` / `rounded-[var(--r2)]` | Drift дизайн-системы относительно Button/Input/Dialog | Выровнять className под hub tokens (className-only, Radix не трогать) |
| W9-A5-019 | P3 | `form-field.tsx:22` | `required` — только визуальный `*` без `aria-required` / sr-only «обязательно» | AT может не узнать required | `aria-required` на control + sr-only текст |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **`any` / `@ts-ignore`** | Во всех 23 файлах — **0** вхождений |
| **Rules of Hooks** | `user-picker`: hooks до return; deps `useMemo`/`useQuery` корректны; утечек `useEffect` нет (нет эффектов) |
| `alert-dialog.tsx` | Thin Radix wrap; Title/Description/Action/Cancel; focus ring; className-only |
| `input.tsx` / `textarea.tsx` | forwardRef, focus-visible, disabled; hub tokens |
| `label.tsx` | Radix Label + cva; peer-disabled |
| `popover.tsx` | Portal + Content; className-only |
| `scroll-area.tsx` | Viewport + ScrollBar + Corner |
| `separator.tsx` | `decorative=true` default — a11y OK |
| `sheet.tsx` | Dialog-based; Escape/focus-trap; `showClose` + «Закрыть»; Title/Description exports |
| `tabs.tsx` | focus-visible на Trigger/Content; Radix keyboard |
| `tooltip.tsx` | Provider/Root/Trigger/Content — стандарт |
| Shadcn invariant | Нет правок Radix behavior — только className / composition |

## Counts

| Метрика | Значение |
|---------|----------|
| Файлов | 23 |
| Находок всего | **19** |
| P0 | 0 |
| P1 | 1 |
| P2 | 7 |
| P3 | 11 |
| Файлов с находками | 14 |
| Файлов чисто | 9 |
| `any` / `@ts-ignore` | **0** |

## Топ-3

1. **W9-A5-001 (P1)** — `ErrorBoundary` липнет после ошибки: навигация через sidebar не восстанавливает UI (`error-boundary.tsx:18-52` + `app-shell`).
2. **W9-A5-002 (P2)** — `Button` без default `type="button"` → latent form-submit во всех call-sites (`button.tsx:43-52`).
3. **W9-A5-003 (P2)** — `ConfirmDialog` закрывается в `finally` даже при failed `onConfirm` (`confirm-dialog.tsx:74-79`).

**Фокус-вердикт:** UI primitives в целом аккуратные (Radix a11y, нет `any`, Shadcn className-only). Главные риски зоны — **sticky ErrorBoundary на Outlet**, **Button type footgun**, **ConfirmDialog async close-on-error**; a11y-долг — form-field hint, CommandDialog aria conflict, decorative icons.
