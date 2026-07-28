# A2 W9 агент 5 — frontend/components part 5 (ui + user-picker)

> TASK-008 · **Стадия A2** · read-only · 2026-07-28  
> Зеркало: `W9-ASSIGNMENTS.md` §A5 · Stage A: `A9-5-components.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/tracker/user-picker.tsx` | 157 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/alert-dialog.tsx` | 127 | 0 | чисто |
| `frontend/src/components/ui/avatar.tsx` | 63 | 1 | замечания |
| `frontend/src/components/ui/badge.tsx` | 32 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/button.tsx` | 55 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/card.tsx` | 78 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/command.tsx` | 127 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/confirm-dialog.tsx` | 112 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/dialog.tsx` | 115 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/dropdown-menu.tsx` | 181 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/empty-state.tsx` | 33 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/error-boundary.tsx` | 79 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/form-field.tsx` | 62 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/input.tsx` | 19 | 0 | чисто |
| `frontend/src/components/ui/label.tsx` | 20 | 0 | чисто |
| `frontend/src/components/ui/popover.tsx` | 25 | 1 | замечания |
| `frontend/src/components/ui/scroll-area.tsx` | 42 | 0 | чисто |
| `frontend/src/components/ui/select.tsx` | 145 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/separator.tsx` | 26 | 0 | чисто |
| `frontend/src/components/ui/sheet.tsx` | 115 | 0 | чисто |
| `frontend/src/components/ui/tabs.tsx` | 48 | 0 | чисто |
| `frontend/src/components/ui/textarea.tsx` | 19 | 0 | чисто |
| `frontend/src/components/ui/tooltip.tsx` | 23 | 0 | чисто |

Файлов в списке: 23. Прочитано: 23. Пропущено: 0.

## Перепроверка Stage A (`W9-A5-*`)

| Старый ID | Tag | Комментарий |
|-----------|-----|-------------|
| W9-A5-001 | **FIXED-OK** | `resetKey` + soft reset; app-shell `resetKey={pathname}` |
| W9-A5-002 | **FIXED-OK** | Button `type={type ?? "button"}` (не Slot) |
| W9-A5-003 | **FIXED-OK** | confirm: close только после successful `onConfirm` |
| W9-A5-004 | **FIXED-OK** | всегда Description |
| W9-A5-005 | **FIXED-OK** | CommandDialog без `aria-describedby={undefined}` |
| W9-A5-006 | **FIXED-OK** | user-picker `isError` + retry |
| W9-A5-007 | **FIXED-OK** | `enabled: Boolean(projectId)` |
| W9-A5-008 | **FIXED-OK** | form-field `aria-describedby` + `aria-required` |
| W9-A5-009…011 | **FIXED-OK** | Dialog/Command/EB i18n |
| W9-A5-012 | **FIXED-OK*** | online sr-only label при `data-online` (*dot всё ещё emerald — NEW) |
| W9-A5-013 | **FIXED-OK** | CardTitle → `h3` |
| W9-A5-014 | **FIXED-OK** | Badge → `span`, без мёртвого focus ring |
| W9-A5-015…017,019 | **FIXED-OK** | decorative `aria-hidden` / required |
| W9-A5-018 | **FIXED-OK** | Select hub tokens (`rounded-[var(--r2)]`, `border-b1`) |

## Находки (open)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W9-5-001 | NEW | P3 | `avatar.tsx:34` | Online-dot: `bg-emerald-500` вместо semantic token | a11y W9-A5-012 закрыт, но цвет обходит `--green` / hub palette | `bg-[var(--green)]` |
| A2W9-5-002 | NEW | P3 | `popover.tsx:22` | PopoverContent: shadcn `rounded-md` / `bg-popover` / default border | Select выровнен (W9-A5-018); Popover остался на defaults — drift DS | className-only → `--r2` / `border-b1` / `bg-base` |

## Чисто

ErrorBoundary reset · Button type · ConfirmDialog async · Command/Dialog i18n · form-field a11y · user-picker error/enabled · Badge/Card semantics · Select tokens · Radix behavior не патчится · `any`/`@ts-ignore` = 0.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **23 / 23 / 0** |
| Stage A перепроверено | 19 |
| FIXED-OK | **19** |
| RESIDUAL | 0 |
| REGRESS | 0 |
| NEW | 2 |
| Open Σ | **2** (P3:2) |
