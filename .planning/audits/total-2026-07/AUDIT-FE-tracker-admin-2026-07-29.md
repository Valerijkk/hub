# AUDIT FE — tracker modules + admin module + tracker components

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`) · branch `cursor-grok-4.5`  
> Модель: `cursor-grok-4.5`  
> Периметр: `frontend/src/modules/tracker/**/*.{ts,tsx}` · `frontend/src/modules/admin/**/*.{ts,tsx}` · `frontend/src/components/tracker/**/*.{ts,tsx}`  
> Фокус: **error≠empty** · **version optimistic lock UI** · **XSS** · **hub-tokens**  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

## Сверка с предшественниками (кратко)

| Источник | Статус в HEAD (этот периметр) |
|----------|-------------------------------|
| A3W10-2-001 (github/gitlab `href`) | **FIXED-OK** — `sanitizeUrl` в `github-links.tsx`, `gitlab-pipeline-status.tsx`; lib `@/lib/safe-url` |
| A3W10-2-005 (burndown loading≠empty) | **FIXED-OK** — loading / error / empty разделены |
| A3W9-4-002/003/006 (automation / comments / time-logs) | **FIXED-OK** — `isError` + `QueryErrorAlert` |
| A2W10-A3-002 (sprint-backlog) | **FIXED-OK** — `loadError` + alert |
| A3W9-4-001 (checklist hover-only delete) | **FIXED-OK** — `group-focus-within` / `focus-visible` |
| A3W9-1 admin panels (pending/limits/trash) | **вне периметра** (`components/admin/*`, не `modules/admin`) |

**REGRESS:** 0 · **RESIDUAL (переоткрытие A3 actives в этом слоте):** 0 · **NEW:** 5

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/modules/tracker/burndown-chart.tsx` | 172 | 0 | чисто (error/loading/empty) |
| `frontend/src/modules/tracker/gantt-chart.tsx` | 275 | 0 | чисто (`QueryStateView`) |
| `frontend/src/modules/tracker/github-links.tsx` | 375 | 1 | замечания (tokens) |
| `frontend/src/modules/tracker/gitlab-pipeline-status.tsx` | 315 | 1 | замечания (tokens; XSS FIXED) |
| `frontend/src/modules/tracker/issue-description.tsx` | 195 | 0 | чисто (Plate + zod guard) |
| `frontend/src/modules/tracker/issue-header.tsx` | 101 | 1 | замечания (tokens) |
| `frontend/src/modules/tracker/issue-message-links-panel.tsx` | 88 | 1 | замечания (tokens) |
| `frontend/src/modules/tracker/issues-cursor.ts` | 24 | 0 | чисто |
| `frontend/src/modules/tracker/issue-sidebar-meta.tsx` | 273 | 1 | замечания (tokens; OL display OK) |
| `frontend/src/modules/tracker/pipeline-status.tsx` | 304 | 2 | **NEW XSS** + tokens |
| `frontend/src/modules/tracker/planning-poker.tsx` | 238 | 2 | **NEW error≠empty** + tokens |
| `frontend/src/modules/tracker/sprint-backlog.tsx` | 204 | 1 | замечания (tokens; error≠empty OK) |
| `frontend/src/modules/tracker/sprint-board.tsx` | 405 | 1 | замечания (tokens; error≠empty OK) |
| `frontend/src/modules/tracker/sprint-retro.tsx` | 180 | 1 | замечания (tokens) |
| `frontend/src/modules/tracker/triage-inbox.tsx` | 168 | 1 | замечания (tokens; QueryStateView OK) |
| `frontend/src/modules/tracker/velocity-chart.tsx` | 154 | 0 | чисто |
| `frontend/src/modules/admin/backups.tsx` | 239 | 1 | замечания (hex gradient; error≠empty OK) |
| `frontend/src/components/tracker/activity-feed.tsx` | 216 | 0 | чисто |
| `frontend/src/components/tracker/automation-rules-panel.tsx` | 249 | 1 | **NEW** soft error≠empty (statuses/channels) |
| `frontend/src/components/tracker/board-column.tsx` | 170 | 0 | чисто |
| `frontend/src/components/tracker/checklist.tsx` | 163 | 0 | чисто (a11y delete FIXED) |
| `frontend/src/components/tracker/comment-list.tsx` | 326 | 0 | чисто |
| `frontend/src/components/tracker/issue-attachments.tsx` | 197 | 1 | **NEW XSS** defense-in-depth |
| `frontend/src/components/tracker/issue-card.tsx` | 229 | 0 | чисто (hub tokens) |
| `frontend/src/components/tracker/issue-form.tsx` | 334 | 0 | чисто (version в mutate) |
| `frontend/src/components/tracker/issue-form-schema.ts` | 34 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor.tsx` | 55 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor-impl.tsx` | 68 | 0 | чисто (Plate JSON, не HTML) |
| `frontend/src/components/tracker/time-tracker.tsx` | 270 | 0 | чисто |
| `frontend/src/components/tracker/user-picker.tsx` | 164 | 0 | чисто |

Файлов в списке: **30**. Прочитано: **30**. Пропущено: **0**.  
(`modules/admin` содержит только `backups.tsx`.)

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| FE-TA-001 | NEW | P2 | `pipeline-status.tsx:263–267` | `href={run.html_url}` **без** `sanitizeUrl` | После TASK-125 github/gitlab закрыты; GitHub Actions runs в том же классе API→DTO URL остались сырыми → `javascript:`/`data:` в href при компрометации/баге BE | `href={sanitizeUrl(run.html_url)}` + `rel="noopener noreferrer"` (уже есть) |
| FE-TA-002 | NEW | P3 | `issue-attachments.tsx:155–159` | `href={att.url \|\| …}` без whitelist | Локальный upload ставит `/files/…` (safe), но prop `attachments` с API может нести внешний URL | `sanitizeUrl(att.url \|\| fallback)`; при `"#"` — не делать `<a>` или disabled |
| FE-TA-003 | NEW | P3 | `planning-poker.tsx:49–56,123–130` | `loadFailed \|\| !pokerSession` → одна копия «Сессия не найдена или истекла» | Сеть/500 визуально неотличима от 404/expired (класс error≠empty) | Разделить: `loadFailed` → alert+retry; `!pokerSession && !loadFailed` → «не найдена» |
| FE-TA-004 | NEW | P3 | `automation-rules-panel.tsx:46–53` (+ render Select) | `statuses = []` / `channels = []` по умолчанию; `isError` не читается | Fail workflow/channels → пустые Select «как будто нет данных», не error | `isError` + compact alert; empty option только при success |
| FE-TA-005 | NEW | P3 | кластер modules: `issue-sidebar-meta.tsx:18–24`, `triage-inbox.tsx:13–18,106`, `issue-header.tsx:21–30`, `pipeline-status.tsx` / `gitlab-pipeline-status.tsx` / `github-links.tsx` (`text-*-600`, `bg-gray-*`), `planning-poker.tsx:230` / `backups.tsx:113` (`#2563EB`), `sprint-backlog`/`sprint-board` emerald/blue-500 | Raw Tailwind / hex вне `--hub` / `var(--*)` | Token drift vs design-system и эталон `issue-card` (`var(--red|blue|…)`) | Заменить на `var(--red)` / `text-t*` / `bg-card` / `from-[var(--blueS)]`; priority badges → semantic tokens как в `issue-card` |

## Фокус-чек

### error≠empty

| Зона | Результат |
|------|-----------|
| sprint-backlog / board / burndown / velocity / gantt / triage | **OK** — loading · error · empty разделены |
| github-links / message-links / backups | **OK** — `loadError` / `isError` + `QueryErrorAlert` |
| automation rules list / comments / time-logs / activity / user-picker | **OK** |
| pipeline + gitlab CI | **OK** (отдельные тексты error vs empty; retry нет — soft, не заводилось как P2) |
| **planning-poker** | **FE-TA-003** |
| **automation statuses/channels** | **FE-TA-004** |

### version optimistic lock UI

| Проверка | Результат |
|----------|-----------|
| `issue-form` update передаёт `version: initialData.version` | **FIXED-OK** (`issue-form.tsx:111–124`) |
| Detail route inject `version: issue.version` в `updateIssue` | **FIXED-OK** (`issues/$issueKey.tsx:71–77`, вне списка файлов, но wiring сайдбара) |
| Сайдбар показывает `Версия: {issue.version}` | **FIXED-OK** (`issue-sidebar-meta.tsx:269`) |
| 409 UX | Toast через `toastApiError` → `formatApiError` / i18n `common:apiErrors.409` («Конфликт — обновите…») — **достаточно** для P2; отдельной кнопки «Перезагрузить задачу» нет (не P0/P1) |
| Checklist / poker / retro / attachments | Нет entity `version` в UI — N/A |

### XSS

| Проверка | Результат |
|----------|-----------|
| `dangerouslySetInnerHTML` / `innerHTML` / `eval` | **0** в 30 файлах |
| Plate | JSON → `PlateContent`; описание — zod `parsePlateValue` |
| User text (titles, comments, labels, errors) | React text children |
| github / gitlab external `href` | **FIXED-OK** `sanitizeUrl` |
| **pipeline-status** `html_url` | **FE-TA-001 (P2)** |
| **issue-attachments** `att.url` | **FE-TA-002 (P3)** |
| Label `style={{ color: label.color }}` | CSSOM property (не HTML inject); без sanitize цвета — hygiene, не XSS HTML |

### hub-tokens

| Зона | Результат |
|------|-----------|
| `components/tracker` (card, form, comment, checklist, board) | В основном `var(--*)` / semantic |
| `modules/tracker` + CI chrome + priority maps | **FE-TA-005** drift |
| `modules/admin/backups` | В основном hub; hex в CTA gradient |

## FIXED-OK (сводка)

| Зона | Доказательство HEAD |
|------|---------------------|
| XSS github/gitlab | `sanitizeUrl` + `rel="noopener noreferrer"` |
| Burndown/velocity/backlog/gantt error≠empty | ветки loading/error/empty |
| Automation rules list / comments / time history | `isError` + `QueryErrorAlert` |
| Admin backups | `loadError` ≠ «Нет резервных копий» |
| Checklist delete a11y | `group-focus-within` / `focus-visible` |
| Optimistic lock wire | version в mutate + display + i18n 409 |
| Plate XSS surface | нет raw HTML |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 4 |
| **Total** | **5** |

| Метрика | Значение |
|---------|----------|
| Файлов | **30 / 30 / 0** |
| Активных | **5** (NEW 5 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (фокус) | error≠empty основные списки · github/gitlab XSS · OL wire · Plate |

### Топ-3

1. **FE-TA-001 (P2, NEW)** — `pipeline-status`: `href={run.html_url}` без `sanitizeUrl` (дыра после фикса github/gitlab).
2. **FE-TA-003 (P3, NEW)** — planning-poker: fail load = «сессия не найдена».
3. **FE-TA-005 (P3, NEW)** — token drift в modules/tracker (+ CI chrome).

---

Файлов в списке: 30. Прочитано: 30. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=1 · P3=4 · Total=5.  
**TASK-008: открыт** · коммит не создавался.
