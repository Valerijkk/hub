# W9 агент A4 — components part 4

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус FE: Rules of Hooks / утечки `useEffect` · a11y · `any`/`@ts-ignore` · `!important` запрещён.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/components/mobile-layout.tsx` | 70 | 0 | чисто |
| `frontend/src/components/mobile-nav.tsx` | 229 | 1 | замечания |
| `frontend/src/components/notifications/notification-bell.tsx` | 53 | 1 | замечания |
| `frontend/src/components/notifications/notification-panel.tsx` | 324 | 1 | замечания |
| `frontend/src/components/notifications/notification-settings.tsx` | 261 | 1 | замечания |
| `frontend/src/components/notifications/push-permission-banner.tsx` | 91 | 0 | чисто |
| `frontend/src/components/notifications/telegram-link-card.tsx` | 180 | 1 | замечания |
| `frontend/src/components/query-state-view.tsx` | 138 | 0 | чисто |
| `frontend/src/components/route-error-boundary.tsx` | 103 | 1 | замечания |
| `frontend/src/components/search/command-palette.tsx` | 280 | 0 | чисто |
| `frontend/src/components/search/jql-editor.tsx` | 463 | 1 | замечания |
| `frontend/src/components/search/jql-search-page.tsx` | 420 | 1 | замечания |
| `frontend/src/components/search/saved-filters.tsx` | 183 | 2 | замечания |
| `frontend/src/components/theme-provider.tsx` | 110 | 0 | чисто |
| `frontend/src/components/tracker/activity-feed.tsx` | 197 | 1 | замечания |
| `frontend/src/components/tracker/automation-rules-panel.tsx` | 238 | 1 | замечания |
| `frontend/src/components/tracker/board-column.tsx` | 171 | 0 | чисто |
| `frontend/src/components/tracker/checklist.tsx` | 163 | 1 | замечания |
| `frontend/src/components/tracker/comment-list.tsx` | 296 | 1 | замечания |
| `frontend/src/components/tracker/issue-attachments.tsx` | 186 | 1 | замечания |
| `frontend/src/components/tracker/issue-card.tsx` | 230 | 0 | чисто |
| `frontend/src/components/tracker/issue-form.tsx` | 331 | 0 | чисто |
| `frontend/src/components/tracker/issue-form-schema.ts` | 35 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor.tsx` | 56 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor-impl.tsx` | 57 | 1 | замечания |
| `frontend/src/components/tracker/time-tracker.tsx` | 258 | 0 | чисто |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 6 |
| P3 | 6 |
| **Σ** | **14** |

### Фокус-метрики

| Проверка | Результат |
|----------|-----------|
| `any` / `@ts-ignore` / `@ts-expect-error` | **0** literal matches |
| `!important` | **0** |
| `useEffect` с cleanup | mobile-layout / theme / push-banner / jql click-outside / board DnD / issue-card DnD / time-tracker — OK |
| Rules of Hooks (early return после hooks) | нарушений не найдено |

### Топ-3

1. **W9-A4-001 (P1)** — `notification-panel.tsx`: вложенный `<button>` (удалить) внутри кнопки-строки уведомления — невалидный HTML / a11y.
2. **W9-A4-002 (P1)** — `activity-feed.tsx`: ошибка `useQuery` отображается как «Нет активности» (loading/error/empty смешаны).
3. **W9-A4-003 (P2)** — `jql-editor.tsx`: автодополнение без combobox-паттерна (`listbox` / `aria-activedescendant`).

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W9-A4-001 | P1 | `notification-panel.tsx:275-311` | Строка уведомления — `<button>`, внутри ещё один `<button>` «Удалить» | Nested interactive: невалидный HTML; SR/клавиатура непредсказуемы; клик delete всплывает к parent | Внешний контейнер → `div`/`role="button"`+клавиши **или** delete вынести из строки (отдельный control вне parent button) |
| W9-A4-002 | P1 | `activity-feed.tsx:16-53` | Только `isLoading`; при `isError` `activities` undefined → ветка «Нет активности» | Тот же класс багов, что закрывал `QueryStateView`: сеть/500 ≠ empty | `QueryStateView` или явный `isError` + retry; empty только при успехе и `length===0` |
| W9-A4-003 | P2 | `jql-editor.tsx:408-458` | Input + dropdown подсказок без `role="combobox"` / `listbox` / `option` / `aria-expanded` / `aria-activedescendant` | Стрелки/Tab работают визуально, но AT не объявляет список и активный пункт (WCAG 4.1.2) | Combobox pattern: связать input↔listbox, id на option, `aria-activedescendant` |
| W9-A4-004 | P2 | `saved-filters.tsx:53-55,75-77,87-89,99-101` | `catch { }` без UI/toast на load/create/delete/share | Сбой API молча: пустой список / «кнопка не работает» | `toastApiError` + loading/error state (как в notification-settings) |
| W9-A4-005 | P2 | `plate-editor-impl.tsx:33-35` | `usePlateEditor({ value })` — value только на mount; смена пропа не синкается | Controlled/readOnly сценарии (смена комментария без remount) показывают stale JSON | `editor.tf.setValue` / key=`JSON.stringify` remount / follow Plate controlled API |
| W9-A4-006 | P2 | `telegram-link-card.tsx:30-31` | `catch { setStatus({ linked: false }) }` глотает ошибку статуса | Сбой API → UI «не привязан» вместо error; риск повторного bind | Разделить loading/error/linked; toast + retry |
| W9-A4-007 | P2 | `notification-bell.tsx:32-39` | `aria-label="Уведомления"` без числа; badge только визуальный | SR не слышит unread count (NFR-16) | `aria-label={\`Уведомления${n ? `, непрочитанных: ${n}` : ""}\`}` + `aria-live` на badge |
| W9-A4-008 | P2 | `saved-filters.tsx:120-145` | Icon-only Share/Trash: Tooltip без `aria-label` на button | Radix Tooltip ≠ accessible name; иконка без имени для AT | `aria-label` на кнопках («Расшарить» / «Удалить фильтр») |
| W9-A4-009 | P3 | `mobile-nav.tsx:201-224` | Кнопка «Ещё» без `aria-expanded` / `aria-controls` | Меню открыто/закрыто не анонсируется | `aria-expanded={moreOpen}` + id на panel |
| W9-A4-010 | P3 | `notification-settings.tsx:42-57` | Quiet hours `Input type="time"` с `id`, без `<Label htmlFor>` | Placeholder ≠ label (WCAG 3.3.2) | Label «Начало»/«Конец» или `sr-only` + `htmlFor` |
| W9-A4-011 | P3 | `checklist.tsx:127-135` | Delete `×` только `title`, нет `aria-label` | `title` слабо для SR; icon-only control | `aria-label="Удалить пункт"` |
| W9-A4-012 | P3 | `issue-attachments.tsx:178-181` | Текст «Перетащите файлы…», DnD handlers нет | Misleading empty copy; drag ничего не делает | Убрать «Перетащите» или добавить dropzone |
| W9-A4-013 | P3 | `route-error-boundary.tsx:59-99` | Fallback без `role="alert"` / `aria-live` | Ошибка маршрута может не попасть в AT | `role="alert"` на контейнер |
| W9-A4-014 | P3 | `comment-list.tsx:56,236`; `jql-search-page.tsx:102,271`; `automation-rules-panel.tsx:96,147` | `as unknown as` / `as Channel[]` / `as Record<…>` вместо узких типов | Обход type-safety (рядом с запретом голого `any`); drift API↔UI | Выровнять DTO (`body_json: Value`, typed `filter_config.jql`, typed `getChannels`) |

## Чисто

Проверено: hooks order · effect cleanup · `any`/`ts-ignore` · `!important` · loading/empty/error где применимо · базовая a11y (labels, aria-*).

| Файл | Что проверено / почему чисто |
|------|------------------------------|
| `mobile-layout.tsx` | `useIsMobile`: matchMedia + removeListener; DesktopOnly/MobileOnly через `hidden` (вне a11y tree) |
| `push-permission-banner.tsx` | timer cleanup; `aria-label` на close; type=button |
| `query-state-view.tsx` | loading `aria-busy`/`aria-live`; error `role="alert"`; разделяет loading/error/empty |
| `command-palette.tsx` | effect сброс query; `aria-label`/`aria-busy`/`aria-live`/`sr-only`; error+retry |
| `theme-provider.tsx` | accent effect sync; localStorage guard; без `any` |
| `board-column.tsx` | DnD effect cleanup; virtualizer count=0 когда off; `aria-label` на create |
| `issue-card.tsx` | draggable cleanup; клик vs drag; stable keys |
| `issue-form.tsx` | RHF hooks до return; title `aria-invalid`/`aria-describedby`; optimistic `version` на update |
| `issue-form-schema.ts` | Zod title 3–200; без any |
| `plate-editor.tsx` | lazy+Suspense; skeleton `aria-busy` |
| `time-tracker.tsx` | tick effect cleanup; narrow zustand selectors; sr-only labels на manual log |

## Чек-лист фокуса A4

```
Hooks
  [x] все хуки до conditional return
  [x] useEffect cleanup (listeners/timers/DnD/interval)
  [x] plate value sync после mount — FAIL (W9-A4-005)

a11y
  [x] nested interactive — FAIL notification-panel
  [x] combobox JQL — FAIL
  [x] unread/expanded/labels/icon-only — частичные FAIL (bell, more, quiet hours, checklist, saved-filters)
  [x] QueryStateView / command-palette — OK

Типы / CSS
  [x] any / @ts-ignore / !important — 0
  [x] as unknown as / loose casts — зафиксировано как P3 (W9-A4-014)

Loading/error/empty
  [x] activity-feed — FAIL (error→empty)
  [x] saved-filters / telegram status — silent fail
```
