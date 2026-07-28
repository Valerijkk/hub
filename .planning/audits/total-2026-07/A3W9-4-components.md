# A3 W9 агент 4 — frontend/components part 4 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W9-4-components.md](A2W9-4-components.md)  
> Stage A: [A9-4-components.md](A9-4-components.md) · Assignments: [W9-ASSIGNMENTS.md](W9-ASSIGNMENTS.md) §A4  
> Фокус: XSS · a11y · Rules of Hooks · design tokens  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с A2 (A2W9-4-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W9-4-001 | RESIDUAL → **FIXED-OK** | закрыт | `saved-filters.tsx:57-60,80-81,92-93,104-105,109-116` — `toastApiError` + `loadError` / `QueryErrorAlert` + retry |
| A2W9-4-002 | NEW → **FIXED-OK** | закрыт | `notification-panel.tsx:320` — `group-focus-within:flex` + `focus-visible:flex` на delete |
| A2W9-4-003 | NEW → **FIXED-OK** | закрыт | `notification-panel.tsx:44-68,278,302` — `var(--blue)` / `var(--purple)` / `color-mix` / `--gB` |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 6

### Stage A (`W9-A4-*`) via A2 — без регресса

| Старый ID | Tag | Комментарий HEAD |
|-----------|-----|------------------|
| W9-A4-001 | FIXED-OK | sibling buttons в `div` (не nested) |
| W9-A4-002 | FIXED-OK | activity-feed `isError` + retry |
| W9-A4-003 | FIXED-OK | jql-editor combobox / listbox / activedescendant |
| W9-A4-004 | FIXED-OK | saved-filters catch→toast (бывший RESIDUAL A2) |
| W9-A4-005 | FIXED-OK | plate `editor.tf.setValue` sync |
| W9-A4-006 | FIXED-OK | telegram `loadError` ≠ unlinked |
| W9-A4-007 | FIXED-OK | bell aria-label с unread count |
| W9-A4-008 | FIXED-OK | Share/Trash `aria-label` |
| W9-A4-009 | FIXED-OK | mobile-nav `aria-expanded` / `aria-controls` |
| W9-A4-010 | FIXED-OK | quiet hours `Label htmlFor` |
| W9-A4-011 | FIXED-OK | checklist delete `aria-label` (видимость на focus — отдельно NEW) |
| W9-A4-012 | FIXED-OK | empty copy «Нажмите Добавить файл» |
| W9-A4-013 | FIXED-OK | route-error `role="alert"` |
| W9-A4-014 | FIXED-OK | узкие DTO casts в зоне |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/mobile-layout.tsx` | 69 | 0 | чисто |
| `frontend/src/components/mobile-nav.tsx` | 233 | 0 | чисто (A fixed) |
| `frontend/src/components/notifications/notification-bell.tsx` | 59 | 1 | замечания (tokens) |
| `frontend/src/components/notifications/notification-panel.tsx` | 329 | 0 | чисто (A2 fixed) |
| `frontend/src/components/notifications/notification-settings.tsx` | 271 | 0 | чисто (A fixed) |
| `frontend/src/components/notifications/push-permission-banner.tsx` | 90 | 0 | чисто |
| `frontend/src/components/notifications/telegram-link-card.tsx` | 191 | 0 | чисто (A fixed) |
| `frontend/src/components/query-state-view.tsx` | 167 | 1 | замечания (tokens) |
| `frontend/src/components/route-error-boundary.tsx` | 106 | 1 | замечания (tokens) |
| `frontend/src/components/search/command-palette.tsx` | 279 | 0 | чисто |
| `frontend/src/components/search/jql-editor.tsx` | 479 | 0 | чисто (A fixed) |
| `frontend/src/components/search/jql-search-page.tsx` | 426 | 2 | замечания |
| `frontend/src/components/search/saved-filters.tsx` | 199 | 0 | чисто (A2 fixed) |
| `frontend/src/components/theme-provider.tsx` | 109 | 0 | чисто |
| `frontend/src/components/tracker/activity-feed.tsx` | 216 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/automation-rules-panel.tsx` | 235 | 1 | замечания |
| `frontend/src/components/tracker/board-column.tsx` | 170 | 0 | чисто |
| `frontend/src/components/tracker/checklist.tsx` | 163 | 1 | замечания |
| `frontend/src/components/tracker/comment-list.tsx` | 310 | 2 | замечания |
| `frontend/src/components/tracker/issue-attachments.tsx` | 197 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/issue-card.tsx` | 229 | 1 | замечания |
| `frontend/src/components/tracker/issue-form.tsx` | 330 | 0 | чисто* |
| `frontend/src/components/tracker/issue-form-schema.ts` | 34 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor.tsx` | 55 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor-impl.tsx` | 68 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/time-tracker.tsx` | 257 | 1 | замечания |

\* `issue-form.tsx` validation `border-red-500` / `text-red-500` — тот же класс token-drift, учтён в A3W9-4-004.

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W9-4-001 | NEW | P2 | `checklist.tsx:127-136` | Delete `×`: `opacity-0 group-hover:opacity-100` без `group-focus-within` / `focus-visible` | Тот же паттерн, что закрытый A2W9-4-002: клавиатура/SR доходят до кнопки (есть `aria-label`), но контрол невидим до hover мышью | `group-focus-within:opacity-100` и/или `focus-visible:opacity-100` (как notification delete) |
| A3W9-4-002 | NEW | P2 | `automation-rules-panel.tsx:36-38,193-196` | `useQuery` → `rules = []` по умолчанию; `isError` не читается | Сеть/500 → UI «Нет правил автоматизации» (error≠empty, класс TASK-047) | `isError` + `QueryErrorAlert`/retry; empty только при success && `length===0` |
| A3W9-4-003 | NEW | P2 | `comment-list.tsx:63-66,116-164` | Только `isLoading`; при `isError` `comments` undefined → ни empty, ни alert | Сбой списка комментариев молча: форма есть, история отсутствует без retry | `isError` + retry (паттерн `activity-feed` / `QueryStateView`) |
| A3W9-4-004 | NEW | P3 | `issue-card.tsx:30-38` (+ `issue-form.tsx:182,187`; `comment-list.tsx:215,235`; `notification-bell.tsx:31,43`; `query-state-view.tsx:84`; `route-error-boundary.tsx:66-68`; `jql-search-page.tsx:43,344`) | Raw Tailwind/`#hex` цвета вне `--hub` / semantic tokens | Token drift после FE-05 / A2W9-4-003: `text-*-500`, `bg-red-500`, `yellow-300`, `emerald-*`, `#D1D2D3` | `text-[var(--blue)]` / `bg-[var(--red)]` / semantic amber для internal; bell badge → `var(--red)` |
| A3W9-4-005 | NEW | P3 | `jql-search-page.tsx:277-280` | Блок ошибки JQL без `role="alert"` | AT может не анонсировать сбой поиска (рядом activity/route уже с alert) | `role="alert"` (+ опционально retry button) |
| A3W9-4-006 | NEW | P3 | `time-tracker.tsx:64-67,212` | `listTimeLogs` без `isError` | Fail → пустая «История» без отличия от «логов нет» | `isError` + compact alert/retry над списком |

## Фокус-чек (XSS / hooks / a11y / tokens)

| Проверка | Результат |
|----------|-----------|
| `dangerouslySetInnerHTML` / `innerHTML` / `eval` | **0** в периметре |
| Plate XSS surface | Plate JSON → `PlateContent` (не raw HTML); sync `setValue` OK |
| User text (titles, bodies, filter names) | React text nodes — OK |
| `window.open(result.url)` (command-palette) | `noopener,noreferrer`; URL с API — без client sanitize (не P0 в слоте; BE trust) |
| `any` / `@ts-ignore` / `@ts-expect-error` | **0** |
| Rules of Hooks / effect cleanup | нарушений нет (matchMedia, DnD, timer tick, plate sync, click-outside) |
| Nested interactive notifications | FIXED-OK |
| Hover-only destructive controls | **1 open** (checklist) |

## FIXED-OK (сверка A2 + Stage A)

| Зона | Проверено в HEAD |
|------|------------------|
| Notification delete a11y | sibling button + `group-focus-within` / `focus-visible` |
| saved-filters fail | toast + loadError + QueryErrorAlert |
| Notification tokens | `var(--blue|purple|…)` / `color-mix` |
| activity / telegram / JQL combobox / Plate sync | без регресса |
| mobile-nav / bell unread / quiet hours / attachments copy / route EB | без регресса |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 3 |
| **Total** | **6** |

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Активных | **6** (NEW 6 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (A2 actives) | **3/3** |
| FIXED-OK (Stage A via A2) | **14/14** без регресса |

### Топ-3

1. **A3W9-4-001 (P2, NEW)** — checklist delete только на hover (регресс-класс A2W9-4-002).
2. **A3W9-4-002 / A3W9-4-003 (P2, NEW)** — automation-rules + comment-list: error≠empty.
3. **A3W9-4-004 (P3, NEW)** — оставшийся token drift (`*-500` / hex) вне notification-panel.

---

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=3 · P3=3 · Total=6.  
Топ: A3W9-4-001 · A3W9-4-002/003 · A3W9-4-004.
