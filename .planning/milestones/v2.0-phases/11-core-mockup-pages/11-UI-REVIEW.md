# Phase 11 -- UI Review

**Audited:** 2026-03-28
**Baseline:** HTML mockups in `redesign/` folder + 11-CONTEXT.md design specs
**Screenshots:** Not captured (no dev server running)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | All labels are in Russian, domain-specific, no generic "Submit/Cancel" patterns |
| 2. Visuals | 3/4 | Strong visual hierarchy across all 4 pages; feature cards lack aria-labels on SVG icons |
| 3. Color | 3/4 | Design system tokens used consistently; 48 hardcoded hex/rgb values across 20 files (expected for mockup-matching) |
| 4. Typography | 4/4 | Deliberate scale using arbitrary values from mockups; weights are purposeful (bold, semibold, extrabold, black) |
| 5. Spacing | 3/4 | Consistent use of design tokens; many arbitrary pixel values from mockup compliance |
| 6. Experience Design | 3/4 | Loading skeletons present on all 4 pages; one CSS conflict in code input; empty states covered |

**Overall: 20/24**

---

## Top 3 Priority Fixes

1. **Duplicate width class on code verification input** -- User sees a narrow 48px-wide input instead of full-width during code entry -- Remove the leading `w-12` from `magic-link-form.tsx:147` which conflicts with trailing `w-full`
2. **Login feature card SVG icons lack accessibility attributes** -- Screen reader users get no information about card icons -- Add `aria-hidden="true"` to the SVG element in login.tsx:119 (decorative icons) or add `role="img" aria-label` for informative ones
3. **Kanban board filter bar "Сбросить" uses a `<span>` instead of `<button>`** -- Keyboard users cannot reach or activate the reset filter action -- Change `<span>` at board/index.tsx:238 to `<button type="button">` with proper focus styles

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

All four restyled pages use intentional, domain-appropriate Russian copy:

- **Login:** "Получить код", "Войти", "Войти через Telegram", "Зарегистрироваться", "Войдите в свой аккаунт" -- clear, action-oriented CTAs
- **Dashboard:** "Доброе утро/день/вечер", quick actions ("Новая задача", "Написать", "Событие", "Поиск") -- contextual and specific
- **Messenger:** "Мессенджер", "Закреплённые", "Личные", "Каналы", "Нет каналов", "Начните общение", "Написать сообщение..." -- complete state coverage
- **Kanban:** "Мои задачи", "Поиск задач...", "Сбросить", "Перетащите задачу сюда", "Не удалось переместить задачу" -- clear empty/error states

Error messages use specific language: "Введите корректный email", "Неверный код", "Неверный или истёкший код". No generic "Something went wrong" patterns found.

### Pillar 2: Visuals (3/4)

**Strengths:**
- Login split layout (55/45) creates strong focal point with mesh gradient and feature cards
- Dashboard greeting with amber-orange-pink gradient text is a clear hero element
- Messenger 3-panel layout with green accent for active state creates clear navigation hierarchy
- Kanban cards use gradient backgrounds, priority dots with glow, hover lift (-3px) for interactivity cues
- Icon-only buttons in messenger consistently use `title` attributes (35 instances across 15 files)

**Issues:**
- Login page feature card SVGs at login.tsx:119 have no `aria-hidden="true"` (decorative icons should be hidden from AT)
- The Telegram button SVG at telegram-button.tsx:48 correctly has `aria-hidden="true"` -- inconsistency with feature card SVGs
- Board column add button (board-column.tsx:79) has no title/aria-label

### Pillar 3: Color (3/4)

**Design system token usage is excellent:**
- All surface colors use CSS custom properties (`--base`, `--card`, `--raised`, `--input`, `--void`)
- Text hierarchy uses `--t1` through `--t4` consistently
- Border tokens `--b1`, `--b2`, `--b3` used throughout
- Accent colors use semantic variables (`--blue`, `--green`, `--purple`, `--amber`, `--red`)
- Glow effects use glow tokens (`--gB`, `--gG`, `--gP`, `--gR`)

**Hardcoded colors (justified by mockup compliance):**
- Login mesh gradient: `rgba(59,130,246,.18)`, `rgba(139,92,246,.14)`, `rgba(52,211,153,.08)` -- exact mockup values
- Telegram button: `#229ED9` -- brand color, appropriate to hardcode
- Issue card priority colors: `#ef4444`, `#f97316`, `#eab308`, `#3b82f6`, `#9ca3af` -- semantic priority, different from accent tokens
- Widget chart components have hardcoded colors for data visualization (expected)

**Observation:** The 48 hardcoded color instances across 20 files is higher than ideal but nearly all are justified by mockup-matching requirements or semantic constants (brand colors, priority colors, chart colors). The design system tokens handle all dynamic/themeable surfaces correctly.

### Pillar 4: Typography (4/4)

The typography scale is deliberate and maps to mockup specifications:

**Font sizes in use (Phase 11 files):**
- `text-[10px]` / `text-[10.5px]` / `text-[11px]` / `text-[11.5px]` -- metadata, timestamps, section labels
- `text-xs` (12px) -- secondary labels, badge text, descriptions
- `text-[13px]` / `text-[13.5px]` -- card titles, widget headers, message author
- `text-sm` (14px) -- body text, channel names
- `text-[15px]` -- inputs, buttons, chat header, tagline
- `text-[30px]` -- login tagline "Единая платформа"
- `text-[34px]` -- dashboard greeting
- `text-4xl` (36px) -- "Hub" logo text

**Font weights:**
- `font-medium` -- timestamps, secondary text
- `font-semibold` -- labels, channel names, buttons
- `font-bold` -- headings, card titles, author names
- `font-extrabold` -- channel header icon
- `font-black` -- hero text, greeting, logo

The scale is larger than typical (8+ sizes) but each is sourced directly from mockup CSS values. Weights serve clear hierarchy purposes. `tracking` values are also mockup-precise (-0.04em, -0.01em, 0.04em, 0.08em).

### Pillar 5: Spacing (3/4)

**Consistent patterns:**
- Page padding: `p-6` (dashboard content), `px-[22px]` (chat area) -- mockup-driven
- Card padding: `p-6` (login feature cards), `p-5` (widget content), `p-16` (kanban cards via CSS class)
- Gaps: `gap-5` (login card grid), `gap-2.5` (quick actions, kanban cards), `gap-[11px]` (channel items)
- Section spacing: `mb-7` (dividers), `mb-9` (logo area), `mb-2.5` (greeting)

**Arbitrary values from mockup compliance:**
- `px-[18px]`, `py-[11px]`, `h-[50px]`, `h-[58px]`, `h-[38px]`, `w-[42px]`, `gap-[7px]`, `gap-[9px]`
- These are exact mockup values and justified by pixel-perfect requirement

**Minor concern:** The density of arbitrary spacing values makes future maintenance harder, but this is an inherent tradeoff of mockup-matching vs. design-token-only approach. The CSS utility classes (`kanban-card`, `dash-widget`, `dash-qa-btn`) help by encapsulating complex spacing in reusable classes.

### Pillar 6: Experience Design (3/4)

**Loading states (present on all 4 pages):**
- Login: Submit button shows spinner SVG + "Отправка..." / "Проверка..." text
- Dashboard: Full skeleton grid with `animate-shimmer` for 6 placeholder widgets
- Messenger: Channel list skeleton (5 shimmer rows), chat loading skeleton (5 message placeholders)
- Kanban: Full skeleton with filter bar + column placeholders

**Empty states (covered):**
- Messenger channel list: "Нет каналов" with icon
- Chat view: "Начните общение" / "Напишите первое сообщение в этом канале"
- Kanban column: "Перетащите задачу сюда" with dashed border

**Error handling:**
- Magic link form: try/catch with `toast.error()` for specific messages
- Telegram auth: `toast.error("Неверный или истёкший код")`
- Kanban move: Optimistic update with rollback on error + `toast.error("Не удалось переместить задачу")`
- Messenger send: `toast.error(message)` with specific error extraction
- Channel delete uses `confirm()` dialog (native browser) -- not ideal but functional

**Bug found:**
- `magic-link-form.tsx:147`: Code verification input has conflicting `className="w-12 ... w-full"`. The `w-12` (48px) appears first and may override `w-full` depending on CSS specificity. The mockup spec calls for individual digit inputs at 48x58px, but the implementation uses a single full-width input. The `w-12` is a leftover from the plan's code digit input spec that wasn't properly adapted.

**Interaction patterns:**
- Disabled states on submit buttons (`disabled={loading || !isEmailValid}`, `disabled={loading || !code.trim()}`)
- Hover states with translateY, scale, glow effects across all pages
- Active/pressed states with `active:scale-[.97]`
- Drag-and-drop preserved on kanban and dashboard widgets
- Context menu for channel mute (right-click)

---

## Registry Safety

Registry audit skipped -- no third-party registries detected in UI-SPEC. shadcn uses official registry only.

---

## Files Audited

- `frontend/src/routes/login.tsx` -- Login page (split layout, mesh gradient, feature cards)
- `frontend/src/components/auth/magic-link-form.tsx` -- Email/code form with mockup inputs
- `frontend/src/components/auth/telegram-button.tsx` -- Telegram login button
- `frontend/src/routes/_authenticated/dashboard.lazy.tsx` -- Dashboard page (greeting, quick actions)
- `frontend/src/components/dashboard/widget-grid.tsx` -- Widget grid with drag-and-drop
- `frontend/src/components/messenger/channel-list.tsx` -- Channel list panel (280px)
- `frontend/src/components/messenger/chat-view.tsx` -- Chat view with header, messages, typing
- `frontend/src/components/messenger/message-item.tsx` -- Message rendering with reactions
- `frontend/src/components/messenger/message-composer.tsx` -- Glassmorphic composer
- `frontend/src/routes/_authenticated/projects/$projectId/board/index.tsx` -- Kanban filter bar
- `frontend/src/components/tracker/board-column.tsx` -- Kanban column (borderless)
- `frontend/src/components/tracker/issue-card.tsx` -- Issue card with priority dots
- `frontend/src/app.css` -- All page-specific CSS utilities and keyframes
