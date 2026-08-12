# A3 W9 агент 2 — frontend/components part 2 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W9-2-components.md](A2W9-2-components.md)  
> Stage A: [A9-2-components.md](A9-2-components.md)  
> Периметр: 26 файлов (W9-ASSIGNMENTS §A2) · Прочитано: **26** · Пропущено: **0**  
> Фокус FE: XSS · a11y · Rules of Hooks · design tokens · Radix/Shadcn behavior intact

## Сверка с A2 (A2W9-2-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W9-2-001 | RESIDUAL | **живо** | `bookmarks-panel.tsx:78-80` — `{bm.channel_id}` / `{bm.message_id}`; FE `Bookmark` и BE `BookmarkResponse` (`dto.go:129-134`) без preview |
| A2W9-2-002 | RESIDUAL | **живо** | `deadlines.tsx:66-91` — `<div>` rows, нет `Link`; `TaskSummary` без `project_id` (BE `dto.go:37-44`, FE `dashboard.ts:26-33`) |
| A2W9-2-003 | NEW→open | **живо** | `my-tasks.tsx:73-97` — тот же non-link паттерн |
| A2W9-2-004 | NEW→open | **живо** | `deadlines.tsx:32-39,72` `red-500`/`amber-*`; `my-tasks.tsx:13-18` `PRIORITY_COLORS` raw Tailwind |

### Cascade FIXED-OK (Stage A / W9-A2-*, без регрессии)

| Stage A / prior | Evidence in HEAD |
|-----------------|------------------|
| W9-A2-001 joinCall | `upcoming-calls.tsx:83-85` — `joinCall(call.id)` |
| W9-A2-002 nested interactive | `sidebar-user.tsx:60-80` — collapsed expand `button`, без Link внутри |
| W9-A2-003/005 collapsed names | `sidebar-link.tsx:63-68` — `aria-label` + unread count |
| W9-A2-004 createProject label | `sidebar.tsx:58-62` — `aria-label` на compose Link |
| W9-A2-006 presence muted | `sidebar-user.tsx:38-42` — `bg-[var(--t3)]`, title «недоступен» |
| W9-A2-007 userCollapsedIds | `sidebar-projects.tsx:153-216` — ручной collapse active |
| W9-A2-008 mobile Escape | `app-shell.tsx:69-83` — close on pathname + Escape |
| W9-A2-009 bookmarks error≠empty | `bookmarks-panel.tsx:54-64` — `QueryStateView` |
| W9-A2-011/012 files a11y | `dropzone.tsx:32-34,44`; `file-upload.tsx:69` |
| W9-A2-013 language pressed | `language-switcher.tsx:34-58` — `role="group"` + `aria-pressed` (compact) |
| W9-A2-014 charts img | burndown/task-status/test-trends/velocity — `role="img"` + summary `aria-label` |
| W9-A2-015 priority img | `my-tasks.tsx:81-82` — `role="img"` + `aria-label` |
| W9-A2-016 typed widget data | `isPersonalData` / `isProjectStats` / `isAdminStats` |
| W9-A2-017 bot-markdown timer | `bot-markdown.tsx:88-95` — `copyTimerRef` cleanup |
| W9-A2-018 bookmarks toast | `bookmarks-panel.tsx:30-32` — `toastApiError` |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 4 · **NEW (этот слот):** 1

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/dashboard/widgets/burndown-chart.tsx` | 109 | 0 | чисто (A fixed; `var(--blue)`) |
| `frontend/src/components/dashboard/widgets/deadlines.tsx` | 89 | 2 | RESIDUAL Link + tokens |
| `frontend/src/components/dashboard/widgets/my-tasks.tsx` | 93 | 2 | RESIDUAL Link + tokens |
| `frontend/src/components/dashboard/widgets/storage-usage.tsx` | 49 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/system-health.tsx` | 102 | 1 | NEW tokens / inline glow |
| `frontend/src/components/dashboard/widgets/task-status-chart.tsx` | 109 | 0 | чисто (A fixed; hex fallback palette — hygiene) |
| `frontend/src/components/dashboard/widgets/team-activity.tsx` | 58 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/test-trends.tsx` | 111 | 0 | чисто (A fixed) |
| `frontend/src/components/dashboard/widgets/unread-messages.tsx` | 44 | 0 | чисто (tokens → A3W9-2-005 вместе с system-health) |
| `frontend/src/components/dashboard/widgets/upcoming-calls.tsx` | 87 | 0 | чисто (A fixed joinCall) |
| `frontend/src/components/dashboard/widgets/velocity-chart.tsx` | 107 | 0 | чисто (A fixed) |
| `frontend/src/components/files/dropzone.tsx` | 78 | 0 | чисто (A fixed) |
| `frontend/src/components/files/file-upload.tsx` | 70 | 0 | чисто (A fixed) |
| `frontend/src/components/help/context-help-tip.tsx` | 47 | 0 | чисто (Tooltip + `aria-label`) |
| `frontend/src/components/help/shortcuts-dialog.tsx` | 59 | 0 | чисто (Shadcn Dialog className-only) |
| `frontend/src/components/language-switcher.tsx` | 86 | 0 | чисто (A fixed compact) |
| `frontend/src/components/layout/app-shell.tsx` | 187 | 0 | чисто (A fixed Escape/route) |
| `frontend/src/components/layout/sidebar.tsx` | 148 | 0 | чисто (A fixed labels) |
| `frontend/src/components/layout/sidebar-link.tsx` | 125 | 0 | чисто (A fixed a11y; decorative nav colors OK) |
| `frontend/src/components/layout/sidebar-nav.tsx` | 341 | 0 | чисто |
| `frontend/src/components/layout/sidebar-projects.tsx` | 305 | 0 | чисто (A fixed collapse; expand ≠ Link) |
| `frontend/src/components/layout/sidebar-user.tsx` | 104 | 0 | чисто (A fixed nested + presence) |
| `frontend/src/components/layout/skip-link.tsx` | 13 | 0 | чисто |
| `frontend/src/components/layout/topbar.tsx` | 86 | 0 | чисто (search `aria-label`/`aria-keyshortcuts`) |
| `frontend/src/components/messenger/bookmarks-panel.tsx` | 112 | 1 | RESIDUAL UUID preview |
| `frontend/src/components/messenger/bot-markdown.tsx` | 173 | 0 | чисто (XSS-safe text nodes; timer OK) |

Файлов в списке: **26**. Прочитано: **26**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W9-2-001 | RESIDUAL (A2W9-2-001 / W9-A2-010 / P2-FE-U03) | P2 | `bookmarks-panel.tsx:78-80` | UI показывает сырые `channel_id` / `message_id` UUID | Список закладок бесполезен; accessible name кнопки = UUID; DTO без preview | BE: `channel_name` + `content_preview` в `BookmarkResponse` **или** FE fetch message/channel; не рендерить голые UUID |
| A3W9-2-002 | RESIDUAL (A2W9-2-002 / W9-A2-019; TASK-075 false-close) | P3 | `deadlines.tsx:66-91` | Строки дедлайнов не кликабельны | Виджет без сценария «открыть задачу»; маршрут требует `projectId` (`issues/$issueKey`) | Добавить `project_id` в `TaskSummary` (BE+FE) → `Link` на `/projects/$projectId/issues/$issueKey` |
| A3W9-2-003 | RESIDUAL (A2W9-2-003) | P3 | `my-tasks.tsx:73-97` | Строки «Мои задачи» без Link | Тот же dead-UI паттерн, что deadlines | Общий helper/row-link с `project_id` + `key` |
| A3W9-2-004 | RESIDUAL (A2W9-2-004) | P3 | `deadlines.tsx:32-39,72`; `my-tasks.tsx:13-18` | Raw Tailwind `red-500` / `amber-500` / `orange-500` / `yellow-500` / `blue-400` / `slate-400` | Обход semantic/`--hub` токенов; drift light/dark | `bg-[var(--red)]` / `var(--amber)` / map приоритетов на токены |
| A3W9-2-005 | NEW | P3 | `system-health.tsx:19-36,83-101` (+ `unread-messages.tsx:24,35-36`) | Status dots: `bg-green/red/yellow-500` + inline `boxShadow: rgba(...)`; unread — `text-green-500` / `bg-amber-500` | A2 пометил файлы «чисто», но тот же class, что A2W9-2-004; glow ещё и вне DS | Semantic `var(--green/--red/--amber)` + tokenized glow; убрать raw rgba inline |

## FIXED-OK / чисто (перепроверено по фокусу)

| Зона | Проверено в HEAD |
|------|------------------|
| XSS | Нет `dangerouslySetInnerHTML` / `innerHTML` в 26 файлах. `BotMarkdown` — React text/`<strong>`/`<em>`/`<code>` nodes, не HTML inject |
| a11y (layout) | Nested interactive, collapsed labels, mobile Escape, skip-link, charts `role="img"`, dropzone/upload labels — без регресса |
| Hooks | Все hooks до early return (`BookmarksPanel`, `AppShell`, widget pure components); timer cleanup в `bot-markdown` |
| Radix/Shadcn | Tooltip/Dialog — `asChild` / className-only; Radix behavior не переопределён |
| Typed data | Widget type guards сохранены |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 4 |
| **Total** | **5** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (Stage A via A2 cascade) | **17** (W9-A2-001…009, 011…018) |
| REGRESS | 0 |
| RESIDUAL | 4 (`A3W9-2-001`…`004` ← A2W9-2-001…004) |
| NEW | 1 (`A3W9-2-005`) |

A2 baseline open: **4** → FIXED-OK **0** · RESIDUAL **4** · NEW **1** · REGRESS **0** → A3 open **5**.

### Топ-3

1. **A3W9-2-001 (P2)** — bookmarks UUID-only (контракт BE+FE).
2. **A3W9-2-002/003 (P3)** — deadlines + my-tasks без Link (`project_id` отсутствует в DTO).
3. **A3W9-2-004/005 (P3)** — raw Tailwind tokens в personal/admin widgets.

---

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=1 · P3=4 · Total=5.  
Топ: A3W9-2-001 (bookmarks UUID) · A3W9-2-002/003 (widget Links) · A3W9-2-004/005 (tokens).
