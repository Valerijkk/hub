# Phase 10 -- UI Review

**Audited:** 2026-03-28
**Baseline:** CONTEXT.md design contract (PRIM-01 through PRIM-10) + HTML mockups in redesign folder
**Screenshots:** not captured (no dev server detected on ports 3000, 5173, 8080)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | Primitive components carry no user-facing copy; accessibility labels present where needed |
| 2. Visuals | 4/4 | Strong visual hierarchy with gradient buttons, card highlights, glow focus states |
| 3. Color | 3/4 | Consistent token usage but hardcoded rgba glow values in button/input/tabs instead of CSS variables |
| 4. Typography | 3/4 | Clean scale (xs, sm, base, lg, 2xl) but font-weight spread is wide (medium, semibold, bold) |
| 5. Spacing | 4/4 | Tailwind scale used consistently; arbitrary values only for spec-mandated dimensions |
| 6. Experience Design | 3/4 | Disabled, focus, hover, active states all covered; avatar online indicator present; missing glow on avatar indicator per spec |

**Overall: 21/24**

---

## Top 3 Priority Fixes

1. **Avatar online indicator missing glow shadow** -- CONTEXT.md spec says `box-shadow 0 0 8px var(--gG)` but the implementation uses only `bg-emerald-500` with no glow shadow -- add `shadow-[0_0_8px_var(--gG)]` to the online indicator span in `avatar.tsx:22`
2. **Hardcoded rgba glow values instead of CSS variables** -- Button (`button.tsx:12,14`), Input (`input.tsx:11`), and Tabs (`tabs.tsx:32`) use inline `rgba(91,156,245,...)` instead of referencing `var(--blue)` or a dedicated glow token -- if the accent color changes these will be stale; extract to CSS custom properties like `--glow-blue` in app.css
3. **Select and Textarea not restyled** -- `select.tsx` and `textarea.tsx` still use default shadcn classes (`rounded-md`, `border-input`, `bg-transparent`, `bg-background`) while all 10 target primitives use design tokens -- these will look visually inconsistent on any page that mixes them with restyled components

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

These are primitive UI components, not page-level views. They contain no user-facing copy strings like "Submit" or "Click Here". The one text element present -- the dialog close button -- correctly uses `<span className="sr-only">Close</span>` for screen reader accessibility (dialog.tsx:47). No generic labels, no placeholder copy issues.

### Pillar 2: Visuals (4/4)

Strong visual design language established across all 10 primitives:

- **Button**: Gradient primary with glow shadow creates clear CTA hierarchy. Active scale(0.96) provides tactile feedback. Six well-differentiated variants (default, destructive, outline, secondary, ghost, link).
- **Card**: Gradient background (card-to-raised) with `::before` top-edge highlight is a distinctive design signature. Hover lift (-translate-y-0.5) adds depth.
- **Input**: Blue focus glow ring is distinctive and aligned with mockups. Subtle scale(1.005) on focus is a nice micro-interaction.
- **Dialog**: Backdrop blur(8px) at 65% opacity creates proper depth. Spring animation (250ms) for entrance/exit.
- **Tabs**: Underline active indicator with primary color and accent glow shadow is clean and functional.
- **DropdownMenu**: Consistent item hover pattern with var(--hover) across all item types.
- **Avatar**: Online indicator with pulse animation is well-implemented. Uses data-attribute toggle pattern.
- **Tooltip**: Clean, minimal surface treatment with proper z-indexing.

Accessibility: Dialog close button has sr-only label (dialog.tsx:47). Avatar online dot has aria-hidden (avatar.tsx:23). Icon-only close button is accessible.

### Pillar 3: Color (3/4)

**Token usage is strong.** All 10 restyled components consistently use design system CSS variables:
- Surface tokens: `bg-base`, `bg-raised`, `bg-[var(--card)]`, `bg-[var(--input)]`, `bg-[var(--hover)]`
- Text tokens: `text-t1`, `text-t2`, `text-t3`
- Border tokens: `border-b1`, `border-b2`, `border-b3`
- Semantic colors: `var(--blue)`, `var(--blueS)`, `var(--red)`, `var(--redS)`
- Glow backgrounds: `var(--gB)`, `var(--gR)`, `var(--gG)`, `var(--gP)` in badge variants

**Issue: Hardcoded rgba glow values.** Four instances of hardcoded rgba that duplicate the blue accent color:
- `button.tsx:12` -- `rgba(91,156,245,0.25)` and `rgba(91,156,245,0.4)` for primary glow
- `button.tsx:14` -- `rgba(239,68,68,0.2)` and `rgba(239,68,68,0.35)` for destructive glow
- `input.tsx:11` -- `rgba(91,156,245,0.3)` and `rgba(91,156,245,0.06)` for focus border/shadow
- `tabs.tsx:32` -- `rgba(91,156,245,0.15)` for active tab glow

These are acceptable for now since Tailwind arbitrary values cannot use `var()` with alpha channel easily, but they will break if accent colors are changed. A dedicated glow token set (e.g., `--glow-blue-sm`, `--glow-blue-md`) would be more maintainable.

**No hardcoded hex colors** found in the 10 audited component files.

Avatar uses `bg-emerald-500` (a Tailwind preset) instead of `var(--green)` or `var(--greenS)` -- minor inconsistency at avatar.tsx:22.

### Pillar 4: Typography (3/4)

**Font sizes in use across the 10 components:**
- `text-xs` -- tooltip, dropdown shortcut
- `text-sm` -- button, dropdown items, dialog description, card description, input file, label
- `text-[11px]` -- badge (arbitrary, matches spec "10-11px")
- `text-[13px]` -- tabs trigger (arbitrary, matches spec "13px")
- `text-base` -- input
- `text-lg` -- dialog title
- `text-2xl` -- card title

That is 7 distinct sizes. However, `text-[11px]` and `text-[13px]` are spec-mandated values that fall between standard Tailwind stops, so the effective scale is reasonable (xs/sm/base/lg/2xl + 2 intentional custom sizes).

**Font weights in use:**
- `font-medium` -- label, input file, tabs trigger (inactive)
- `font-semibold` -- card title, dialog title, dropdown label, tabs trigger (active)
- `font-bold` -- button, badge

Three weights is within acceptable range for a design system. The distinction is clear: bold for interactive elements (buttons, badges), semibold for headings/labels, medium for body content.

### Pillar 5: Spacing (4/4)

All spacing uses the Tailwind scale consistently:
- Padding: `p-6` (card header/content/footer), `p-7` (dialog content = 28px per spec), `p-1` (dropdown/menu)
- Horizontal: `px-3` (input, tooltip, dropdown items), `px-4` (button default), `px-8` (button large), `px-3.5` (tabs)
- Vertical: `py-2` (input, button), `py-1.5` (tooltip, dropdown items)
- Gap: `gap-1` (tabs list), `gap-2` (button content, dropdown items), `gap-4` (dialog content)
- Space: `space-y-1.5` (card header, dialog header)

**Arbitrary spacing values** are spec-mandated and justified:
- `px-[7px]`, `py-[2px]` in badge -- matches spec "2px 7px padding"
- `py-[7px]` in tabs -- matches spec "7px 14px padding"
- `h-[42px]` in input -- matches spec "42px height"
- `w-[5px]`, `h-[5px]` in scroll-area -- matches spec "5px scrollbar"
- `p-[1px]` in scrollbar -- minimal padding for thumb

No rogue arbitrary spacing found.

### Pillar 6: Experience Design (3/4)

**State coverage is comprehensive:**
- **Disabled**: All interactive components have `disabled:pointer-events-none disabled:opacity-50`
- **Focus**: Button and input use `focus-visible` with ring/glow patterns. Dialog close, dropdown items all have focus styles.
- **Hover**: Button has translateY + shadow increase. Card lifts with shadow. Dropdown items use `var(--hover)` bg. Tabs text brightens.
- **Active**: Button has `active:scale-[.96]`
- **Open/Closed**: Dialog and dropdown have enter/exit animations with fade and zoom

**Missing per spec:**
- Avatar online indicator lacks `box-shadow: 0 0 8px var(--gG)` glow. The spec (PRIM-09 in CONTEXT.md) explicitly calls for this glow but the implementation only has `bg-emerald-500` with no shadow.
- No confirmation pattern for destructive actions (expected at page level, not primitive level -- not a deduction).

**Observation:** Select (`select.tsx`) and Textarea (`textarea.tsx`) were not in scope for Phase 10 but remain un-restyled. They use old shadcn defaults (`rounded-md`, `border-input`, `bg-transparent`) which will clash visually with the restyled primitives when used on the same page. This is a downstream risk for Phases 11-13.

---

## Files Audited

| File | Status | Tokens Used |
|------|--------|-------------|
| `frontend/src/components/ui/button.tsx` | Restyled | var(--r2), var(--spring), gradients, glow shadows |
| `frontend/src/components/ui/card.tsx` | Restyled | var(--r4), gradient bg, ::before highlight |
| `frontend/src/components/ui/input.tsx` | Restyled | var(--r2), var(--input), var(--ease), blue focus glow |
| `frontend/src/components/ui/badge.tsx` | Restyled | pill shape, glow variants (gB, gR, gG, gP) |
| `frontend/src/components/ui/dialog.tsx` | Restyled | var(--r4), backdrop blur, spring animation, bg-base |
| `frontend/src/components/ui/tabs.tsx` | Restyled | underline indicator, var(--ease), primary border |
| `frontend/src/components/ui/dropdown-menu.tsx` | Restyled | var(--r1), var(--r2), var(--hover), bg-raised |
| `frontend/src/components/ui/scroll-area.tsx` | Restyled | 5px scrollbar, bg-b2/b3 thumb |
| `frontend/src/components/ui/avatar.tsx` | Restyled | online indicator, pulse animation |
| `frontend/src/components/ui/tooltip.tsx` | Restyled | var(--r1), bg-raised, border-b1 |
| `frontend/src/components/ui/select.tsx` | NOT in scope | Still uses default shadcn classes |
| `frontend/src/components/ui/textarea.tsx` | NOT in scope | Still uses default shadcn classes |
| `frontend/src/app.css` | Reference | Design tokens, keyframes |
