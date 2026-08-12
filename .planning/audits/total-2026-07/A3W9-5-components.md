# A3 W9 агент A5 — components part 5 (ui + user-picker) FINAL

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W9-5-components.md](A2W9-5-components.md)  
> Stage A: [A9-5-components.md](A9-5-components.md)  
> Assignment: [W9-ASSIGNMENTS.md](W9-ASSIGNMENTS.md) § A5 · Периметр: **23** · Прочитано: **23/23** · Пропущено: **0**  
> Фокус: XSS · a11y · hooks · hub tokens · Shadcn className-only

## Сверка с A2 (A2W9-5-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W9-5-001 | NEW → **RESIDUAL** | живо | `avatar.tsx:34` — всё ещё `bg-emerald-500`; эталон online — `presence-indicator.tsx:21` `bg-[var(--green)]` |
| A2W9-5-002 | NEW → **RESIDUAL** | живо | `popover.tsx:22` — `rounded-md border bg-popover text-popover-foreground`; Dropdown/Dialog/Select Trigger+Content уже на `--r2`/`border-b*`/`bg-base` |
| W9-A5-001…019 (via A2 FIXED-OK) | FIXED-OK | без регресса | EB `resetKey`+soft; Button `type`; Confirm close-on-success; Description always; CommandDialog description; user-picker error/enabled; form-field a11y; Dialog/Command/EB/Sheet i18n; CardTitle `h3`; Badge `span`; decorative `aria-hidden`; Select Trigger+Content hub tokens |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 2 · **NEW (этот слот):** 1

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/tracker/user-picker.tsx` | 164 | 0 | чисто (A/A2 fixed) |
| `frontend/src/components/ui/alert-dialog.tsx` | 140 | 0 | чисто |
| `frontend/src/components/ui/avatar.tsx` | 71 | 1 | RESIDUAL emerald |
| `frontend/src/components/ui/badge.tsx` | 37 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/button.tsx` | 60 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/card.tsx` | 86 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/command.tsx` | 136 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/confirm-dialog.tsx` | 115 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/dialog.tsx` | 128 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/dropdown-menu.tsx` | 198 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/empty-state.tsx` | 34 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/error-boundary.tsx` | 89 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/form-field.tsx` | 65 | 0 | чисто (A fixed) |
| `frontend/src/components/ui/input.tsx` | 22 | 0 | чисто |
| `frontend/src/components/ui/label.tsx` | 24 | 0 | чисто |
| `frontend/src/components/ui/popover.tsx` | 31 | 1 | RESIDUAL DS drift |
| `frontend/src/components/ui/scroll-area.tsx` | 46 | 0 | чисто |
| `frontend/src/components/ui/select.tsx` | 157 | 1 | NEW: Item/Separator tokens |
| `frontend/src/components/ui/separator.tsx` | 29 | 0 | чисто |
| `frontend/src/components/ui/sheet.tsx` | 130 | 0 | чисто |
| `frontend/src/components/ui/tabs.tsx` | 55 | 0 | чисто |
| `frontend/src/components/ui/textarea.tsx` | 22 | 0 | чисто |
| `frontend/src/components/ui/tooltip.tsx` | 30 | 0 | чисто |

Файлов в списке: 23. Прочитано: 23. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W9-5-001 | RESIDUAL (A2W9-5-001) | P3 | `avatar.tsx:34` | Online-dot: `bg-emerald-500` вместо semantic token | a11y-label (W9-A5-012) закрыт; цвет всё ещё мимо `--green` / hub palette; соседний `PresenceIndicator` уже на `var(--green)` | `bg-[var(--green)]` (shadow уже `var(--gG)`) |
| A3W9-5-002 | RESIDUAL (A2W9-5-002) | P3 | `popover.tsx:22` | PopoverContent: shadcn `rounded-md` / `bg-popover` / default `border` | Select Trigger+Content (W9-A5-018) и Dropdown выровнены; Popover — единственный floating surface на defaults → drift DS | className-only → `rounded-[var(--r2)] border border-b1 bg-base text-t1` (как Dropdown/Select) |
| A3W9-5-003 | NEW | P3 | `select.tsx:119`; `:140` | После W9-A5-018 Trigger/Content на hub tokens; **SelectItem** остался `rounded-sm focus:bg-accent`; **SelectSeparator** — `bg-muted` | Неполный sweep того же класса, что A2W9-5-002: Item/Separator визуально и hover-state расходятся с DropdownItem (`rounded-[var(--r1)]` + `focus:bg-[var(--hover)]`) | Item: `rounded-[var(--r1)] focus:bg-[var(--hover)] focus:text-t1`; Separator: `bg-b1` |

## FIXED-OK (сверка Stage A / A2 фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| XSS | `dangerouslySetInnerHTML` / `innerHTML` / `eval` = **0** во всех 23; user-picker name/email — text children (React escape) |
| Hooks | `user-picker`: `useQuery`/`useMemo`/`useState` до return; `enabled: Boolean(projectId)`; deps OK; эффектов нет |
| `any` / `@ts-ignore` | **0** |
| ErrorBoundary | `resetKey` + softReset + i18n fallback (`error-boundary.tsx:65-77,18-46`) |
| Button | `type={type ?? "button"}` только non-Slot (`button.tsx:52`) |
| ConfirmDialog | close только после successful `onConfirm`; Description всегда (`confirm-dialog.tsx:77-105`) |
| CommandDialog | Title+Description sr-only, без `aria-describedby={undefined}` (`command.tsx:36-41`) |
| user-picker | `isError`+retry; `aria-expanded`/combobox; decorative `aria-hidden` |
| form-field | `aria-describedby` + `aria-required` + sr-only required |
| Badge / Card | `span` / `h3` |
| Select Trigger+Content | `rounded-[var(--r2)] border-b1 bg-base` (W9-A5-018 держится) |
| Dropdown / Dialog / Sheet / Tooltip | hub tokens + i18n close / decorative icons |
| Radix | behavior не патчится — только className / composition |

## Чисто (кратко)

alert-dialog · badge · button · card · command · confirm · dialog · dropdown · empty-state · error-boundary · form-field · input · label · scroll-area · separator · sheet · tabs · textarea · tooltip · user-picker — open=0 в этом слоте.

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 3 |
| **Total** | **3** |

| Метрика | Значение |
|---------|----------|
| Файлов | **23 / 23 / 0** |
| Активных | **3** (NEW 1 · RESIDUAL 2 · REGRESS 0) |
| FIXED-OK (A2 actives) | **0/2** (оба RESIDUAL) |
| FIXED-OK (Stage A via A2) | **19** без регресса |

### Топ-3

1. **A3W9-5-001 (P3, RESIDUAL)** — avatar online-dot `bg-emerald-500` → `var(--green)`.
2. **A3W9-5-002 (P3, RESIDUAL)** — PopoverContent shadcn defaults vs hub floating surfaces.
3. **A3W9-5-003 (P3, NEW)** — SelectItem/Separator не дотянуты до hub tokens после W9-A5-018.

---

Файлов в списке: 23. Прочитано: 23. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=0 · P3=3 · Total=3.  
Топ: A3W9-5-001 · A3W9-5-002 · A3W9-5-003.
