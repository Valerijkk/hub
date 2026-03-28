# Phase 8: Design System Foundation - Research

**Researched:** 2026-03-28
**Domain:** CSS design token migration, font swap, theme system, !important cleanup
**Confidence:** HIGH

## Summary

Phase 8 replaces the entire CSS variable foundation in `app.css` (1053 lines), swaps the font stack from Geist to Inter + JetBrains Mono, removes all 56 `!important` declarations, adds new design tokens (easing, shadows, radii, noise overlay), and defines a derived light theme palette. The scope is confined to 2 files (`app.css` and `index.html`) plus minor potential touch-ups in components that depended on `!important` overrides.

The mockup CSS tokens have been directly verified against 11 HTML mockups in the `redesign/` folder. All token values are concrete and documented below. The critical risk is the `!important` removal -- 56 declarations targeting selects, inputs, textareas, Radix popovers/dropdowns, and checkboxes must be removed without breaking form elements or third-party components (Plate editor, Tremor charts). The secondary risk is light theme contrast -- no light mockups exist, so derived values need WCAG AA verification.

**Primary recommendation:** Execute as a single atomic CSS variable swap followed by systematic `!important` removal. The font swap and noise overlay are independent and can be done in parallel. Test all form elements and Radix-based dropdowns after `!important` removal.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Font Loading Strategy: Google Fonts CDN. Inter weights 400, 500, 600, 700, 800, 900. JetBrains Mono weights 400, 600. font-display: swap.
- Light Theme Derivation: Inversion pattern -- white surfaces (#FFFFFF -> #F8F9FA), gray text/borders, same accents. Light shadows reduced opacity (0.05-0.15 vs 0.4-0.5 dark). Light gradients on cards minimized. Define light tokens in Phase 8, full visual QA in Phase 13.
- CSS Variable Migration: Dual mapping -- new tokens (--void, --base, --raised, --card, --input, --t1-t4, --b1-b3) + Shadcn namespace compatibility (--background, --foreground, --primary, etc.). Noise texture overlay as global body::after. Easing variables as CSS custom properties in :root. !important cleanup -- remove all 56 declarations, replace with scoped cn() in components where needed.

### Claude's Discretion
- Exact light mode hex values for surfaces and text (best-guess inversion)
- z-index scale tokens
- Scrollbar styling
- Concrete noise overlay implementation (SVG filter vs CSS)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DSFN-01 | Inter font for UI text, JetBrains Mono for code | Font swap in index.html + app.css body rule. Exact CDN URL and weights verified from mockups |
| DSFN-02 | 5-level surface palette (void/base/raised/card/input) in dark mode | All 5 hex values verified from 11 mockups. @theme inline mapping documented |
| DSFN-03 | 4-level text hierarchy (t1/t2/t3/t4) with correct contrast | Hex values verified. Light mode values need WCAG AA contrast check |
| DSFN-04 | 3-level border system (b1/b2/b3) with correct opacity | rgba values verified from mockups |
| DSFN-05 | Radius system (6/8/12/16px + pill) | Values verified. Replaces current calc-based system |
| DSFN-06 | 4-level shadow system (sh1-sh4) | sh1-sh3 in most mockups, sh4 in messenger/tms/trash for modals |
| DSFN-07 | Spring/bounce easing on hover/focus/enter animations | --ease, --spring verified in all mockups; --bounce in tms mockup |
| DSFN-08 | All !important removed without visual regression | 56 declarations mapped by line number and target element |
| DSFN-09 | Accent color selection (red/blue/green/purple/orange) applied consistently | Current accent system works via --primary override; needs value update for new palette |
| THME-01 | Dark theme matching mockup palette exactly | All dark token values extracted and verified |
| THME-02 | Light theme derived from dark tokens | Light values recommended in Architecture Patterns section |
| THME-03 | Dark/light/system switching without visual glitches | next-themes setup unchanged; only CSS variable values change |
| THME-04 | Light theme passes WCAG AA contrast ratios | Light values designed for 4.5:1 text / 3:1 UI. Must verify with axe-core |
</phase_requirements>

## Architecture Patterns

### Current app.css Structure (1053 lines)

```
Lines 1-41:     @import "tailwindcss" + @theme inline block (Shadcn -> CSS var mapping)
Lines 44-74:    :root {} (light theme, ~20 HSL variables)
Lines 77-106:   .dark {} (dark theme, ~20 HSL variables)
Lines 111-161:  Accent color selectors (5 colors x 2 themes = 10 blocks)
Lines 163-191:  Module gradient classes (.gradient-tracker, .gradient-text, etc.)
Lines 196-309:  Utility classes (.glass, .shadow-card, .hover-lift, animations)
Lines 314-348:  @layer base {} (body, headings, global transitions)
Lines 350-518:  Form element overrides with !important (THE MINEFIELD)
Lines 520-567:  Kanban, scrollbar, card-premium classes
Lines 568-1053: Badges, animations, glow effects, page-specific utilities
```

### Recommended Migration Structure

The new `app.css` should be reorganized into this order:

```
1. @import "tailwindcss"
2. @theme inline { ... }           -- Shadcn compat + new token utilities
3. :root { ... }                   -- Light theme (ALL new tokens)
4. .dark { ... }                   -- Dark theme (ALL new tokens, source of truth)
5. [data-accent="X"] selectors     -- Updated for new palette
6. .dark[data-accent="X"] selectors
7. @layer base { ... }             -- body, headings, scrollbar, noise overlay
8. Keyframe animations             -- Consolidated @keyframes
9. Utility classes                 -- Keep what's still needed, remove dead code
```

### Pattern 1: Dual Token Mapping (Shadcn Compat + New System)

The `@theme inline` block must map BOTH the existing Shadcn namespace (so `bg-card`, `text-foreground` etc. continue working in all 15 Shadcn components) AND expose new design system tokens as Tailwind utilities.

```css
@theme inline {
  /* Shadcn compatibility layer -- existing components use these */
  --color-background: var(--base);
  --color-foreground: var(--t1);
  --color-card: var(--card);
  --color-card-foreground: var(--t1);
  --color-popover: var(--raised);
  --color-popover-foreground: var(--t1);
  --color-primary: var(--primary);
  --color-primary-foreground: var(--primary-foreground);
  --color-secondary: var(--raised);
  --color-secondary-foreground: var(--t2);
  --color-muted: var(--raised);
  --color-muted-foreground: var(--t2);
  --color-accent: var(--hover);
  --color-accent-foreground: var(--t1);
  --color-destructive: var(--red);
  --color-destructive-foreground: #fff;
  --color-border: var(--b1);
  --color-input: var(--b1);
  --color-ring: var(--primary);

  /* Sidebar compat */
  --color-sidebar: var(--sidebar);
  --color-sidebar-foreground: var(--t1);
  --color-sidebar-primary: var(--t1);
  --color-sidebar-primary-foreground: var(--sidebar);
  --color-sidebar-accent: var(--active);
  --color-sidebar-accent-foreground: var(--t1);
  --color-sidebar-border: var(--b1);
  --color-sidebar-ring: var(--blue);

  /* New design system utilities -- use these in new/updated code */
  --color-void: var(--void);
  --color-base: var(--base);
  --color-raised: var(--raised);
  --color-surface-card: var(--card);
  --color-surface-input: var(--input);
  --color-t1: var(--t1);
  --color-t2: var(--t2);
  --color-t3: var(--t3);
  --color-t4: var(--t4);
  --color-b1: var(--b1);
  --color-b2: var(--b2);
  --color-b3: var(--b3);

  /* Radius -- fixed values, replacing calc-based system */
  --radius-sm: var(--r1);   /* 6px */
  --radius-md: var(--r2);   /* 8px */
  --radius-lg: var(--r3);   /* 12px */
  --radius-xl: var(--r4);   /* 16px */

  /* Shadows */
  --shadow-sm: var(--sh1);
  --shadow-md: var(--sh2);
  --shadow-lg: var(--sh3);
  --shadow-xl: var(--sh4);

  /* Fonts */
  --font-sans: var(--sans);
  --font-mono: var(--mono);
}
```

### Pattern 2: Dark Theme Token Block (Source of Truth)

Values verified directly from mockup `:root` declarations across all 11 HTML files:

```css
.dark {
  /* Surfaces */
  --void: #08090B;
  --base: #0D0E12;
  --raised: #141519;
  --card: #191A20;
  --input: #1E2028;
  --sidebar: #0B0C0F;

  /* Interactive states */
  --hover: rgba(255,255,255,.032);
  --active: rgba(255,255,255,.058);

  /* Text hierarchy */
  --t1: #F2F3F5;
  --t2: #8C8F97;
  --t3: #4E5058;
  --t4: #363840;

  /* Borders */
  --b1: rgba(255,255,255,.045);
  --b2: rgba(255,255,255,.08);
  --b3: rgba(255,255,255,.13);

  /* Semantic colors */
  --blue: #5B9CF5;    --blueS: #3B7EE0;
  --purple: #A78BFA;  --purpleS: #8B6FF0;
  --green: #34D399;   --greenS: #22C088;
  --amber: #FBBF24;
  --red: #F87171;     --redS: #EF4444;
  --orange: #FB923C;
  --gray: #6B7280;

  /* Glow backgrounds */
  --gB: rgba(91,156,245,.15);
  --gP: rgba(167,139,250,.12);
  --gG: rgba(52,211,153,.12);
  --gR: rgba(248,113,113,.1);

  /* Radii */
  --r1: 6px; --r2: 8px; --r3: 12px; --r4: 16px; --pill: 100px;

  /* Typography */
  --sans: 'Inter', -apple-system, sans-serif;
  --mono: 'JetBrains Mono', monospace;

  /* Easing */
  --ease: cubic-bezier(.25,.1,0,1);
  --spring: cubic-bezier(.175,.885,.32,1.1);
  --bounce: cubic-bezier(.34,1.56,.64,1);

  /* Shadows */
  --sh1: 0 1px 3px rgba(0,0,0,.5);
  --sh2: 0 4px 16px rgba(0,0,0,.4), 0 1px 4px rgba(0,0,0,.25);
  --sh3: 0 12px 40px rgba(0,0,0,.5);
  --sh4: 0 24px 80px rgba(0,0,0,.6), 0 4px 24px rgba(0,0,0,.4);

  /* Accent system -- keep --primary and --ring for Shadcn compat */
  --primary: var(--blue);
  --primary-foreground: #fff;
}
```

### Pattern 3: Light Theme Derived Values (Claude's Discretion)

Designed for WCAG AA compliance (4.5:1 text, 3:1 UI). Must be verified with axe-core.

```css
:root {
  /* Surfaces -- inverted: white/light gray */
  --void: #FFFFFF;
  --base: #F8F9FA;
  --raised: #F1F3F5;
  --card: #FFFFFF;
  --input: #F1F3F5;
  --sidebar: #1A1D21;  /* sidebar stays dark */

  /* Interactive states */
  --hover: rgba(0,0,0,.03);
  --active: rgba(0,0,0,.06);

  /* Text -- darker for contrast on white */
  --t1: #111827;   /* contrast ~15:1 on white */
  --t2: #4B5563;   /* contrast ~7:1 on white */
  --t3: #9CA3AF;   /* contrast ~3:1 on white -- placeholder only */
  --t4: #D1D5DB;   /* disabled text */

  /* Borders -- black-based alpha */
  --b1: rgba(0,0,0,.06);
  --b2: rgba(0,0,0,.10);
  --b3: rgba(0,0,0,.16);

  /* Semantic colors -- slightly darker for contrast on white */
  --blue: #3B7EE0;    --blueS: #2563EB;
  --purple: #8B6FF0;  --purpleS: #7C3AED;
  --green: #22C088;   --greenS: #16A34A;
  --amber: #D97706;
  --red: #EF4444;     --redS: #DC2626;
  --orange: #EA580C;
  --gray: #6B7280;

  /* Glow backgrounds -- stronger opacity for visibility on light */
  --gB: rgba(59,126,224,.12);
  --gP: rgba(139,111,240,.10);
  --gG: rgba(34,192,136,.10);
  --gR: rgba(239,68,68,.08);

  /* Radii, fonts, easing -- same as dark */
  --r1: 6px; --r2: 8px; --r3: 12px; --r4: 16px; --pill: 100px;
  --sans: 'Inter', -apple-system, sans-serif;
  --mono: 'JetBrains Mono', monospace;
  --ease: cubic-bezier(.25,.1,0,1);
  --spring: cubic-bezier(.175,.885,.32,1.1);
  --bounce: cubic-bezier(.34,1.56,.64,1);

  /* Shadows -- much lower opacity for light backgrounds */
  --sh1: 0 1px 3px rgba(0,0,0,.08);
  --sh2: 0 4px 16px rgba(0,0,0,.08), 0 1px 4px rgba(0,0,0,.04);
  --sh3: 0 12px 40px rgba(0,0,0,.12);
  --sh4: 0 24px 80px rgba(0,0,0,.15), 0 4px 24px rgba(0,0,0,.08);

  /* Accent system */
  --primary: var(--blue);
  --primary-foreground: #fff;
}
```

### Pattern 4: Accent Color Update

Current accent system uses HSL values. New palette uses hex. Update all 10 blocks:

```css
/* Light theme accents */
[data-accent="red"]    { --primary: #EF4444; --primary-foreground: #fff; --ring: #EF4444; }
[data-accent="blue"]   { --primary: #3B7EE0; --primary-foreground: #fff; --ring: #3B7EE0; }
[data-accent="green"]  { --primary: #22C088; --primary-foreground: #fff; --ring: #22C088; }
[data-accent="purple"] { --primary: #8B6FF0; --primary-foreground: #fff; --ring: #8B6FF0; }
[data-accent="orange"] { --primary: #EA580C; --primary-foreground: #fff; --ring: #EA580C; }

/* Dark theme accents */
.dark[data-accent="red"]    { --primary: #F87171; --primary-foreground: #111; --ring: #F87171; }
.dark[data-accent="blue"]   { --primary: #5B9CF5; --primary-foreground: #111; --ring: #5B9CF5; }
.dark[data-accent="green"]  { --primary: #34D399; --primary-foreground: #111; --ring: #34D399; }
.dark[data-accent="purple"] { --primary: #A78BFA; --primary-foreground: #111; --ring: #A78BFA; }
.dark[data-accent="orange"] { --primary: #FB923C; --primary-foreground: #111; --ring: #FB923C; }
```

### Pattern 5: Noise Overlay (from mockups)

Every mockup includes identical `body::after` noise texture. Implementation:

```css
body::after {
  content: '';
  position: fixed;
  inset: 0;
  z-index: 10000;
  pointer-events: none;
  opacity: .3;
  background: url("data:image/svg+xml,%3Csvg viewBox='0 0 512 512' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.03'/%3E%3C/svg%3E");
}
```

This is an inline SVG data URI using `feTurbulence` for fractal noise. The outer opacity is .3 and the inner rect opacity is .03 -- extremely subtle. `z-index: 10000` ensures it sits above all content but `pointer-events: none` makes it non-interactive.

**Important:** The z-index of 10000 could conflict with Radix portaled content. Consider whether a lower z-index (e.g., 9999) or Radix portal z-index adjustments are needed. Test dropdown/dialog visibility after adding this.

### Pattern 6: Font Swap

Replace line 12 of `index.html`:

```html
<!-- Current -->
<link href="https://fonts.googleapis.com/css2?family=Geist:wght@300;400;500;600;700&family=Geist+Mono:wght@400;500&display=swap" rel="stylesheet" />

<!-- New -->
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet" />
```

Update body font in `app.css`:
```css
body {
  font-family: var(--sans); /* 'Inter', -apple-system, sans-serif */
}
```

Note: The mockups use JetBrains Mono weights 400 and 500. CONTEXT.md says 400 and 600. Use 400 and 600 per user decision.

### Pattern 7: Z-Index Scale (Claude's Discretion)

Recommended token scale to prevent z-index wars:

```css
:root {
  --z-dropdown: 50;
  --z-sticky: 100;
  --z-overlay: 200;
  --z-modal: 300;
  --z-toast: 400;
  --z-tooltip: 500;
  --z-noise: 9999;  /* noise overlay -- below Radix portals at 10000+ */
}
```

### Pattern 8: Scrollbar Styling (Claude's Discretion)

Mockups specify 5px scrollbar. Add Firefox support:

```css
/* WebKit */
::-webkit-scrollbar { width: 5px; height: 5px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(255,255,255,.05); border-radius: 10px; }
::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,.12); }

/* Firefox */
* { scrollbar-width: thin; scrollbar-color: rgba(255,255,255,.05) transparent; }
```

### Anti-Patterns to Avoid

- **Hardcoded hex in components:** Never use `bg-[#191A20]`. Use `bg-card` or `bg-surface-card`.
- **New !important declarations:** Zero tolerance. If specificity is needed, use more specific selectors or `@layer`.
- **HSL format for new tokens:** The new system uses hex and rgba. Do not convert back to HSL.
- **Partial variable swap:** All `:root` and `.dark` tokens must change in a single commit.
- **Keeping old Shadcn var names in :root/.dark:** Old variable names (`--background`, `--foreground`, etc.) should be REMOVED from `:root`/`.dark` blocks. They are now derived from new tokens via `@theme inline`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Noise texture | Canvas-based noise generator | Inline SVG feTurbulence in body::after | Mockups already have working implementation; zero JS needed |
| Font metrics matching | Custom size-adjust calculations | Google Fonts CDN with font-display: swap | Internal tool, 10-50 users, CLS acceptable |
| Theme switching logic | Custom theme toggle | Existing next-themes setup (unchanged) | Already works, only CSS values change |
| Contrast ratio checking | Manual color math | Browser axe-core extension or Lighthouse | Automated, catches all violations at once |

## Common Pitfalls

### Pitfall 1: !important Removal Cascade

**What goes wrong:** Removing all 56 `!important` declarations at once causes form elements (inputs, selects, textareas) and Radix popovers/dropdowns to lose their styled appearance. They fall back to browser defaults or Tailwind resets.

**Why it happens:** The `!important` declarations in lines 356-518 are the ONLY styles targeting native form elements and Radix content wrappers. Without them, Shadcn component `cn()` classes become the sole styling mechanism -- but some components may not have complete styling without the global overrides.

**How to avoid:** Remove `!important` declarations in groups by target:
1. First: Remove Radix content wrapper overrides (lines 407-431) -- Shadcn components already style these
2. Second: Remove input/textarea overrides (lines 434-459) -- Shadcn Input/Textarea components handle this
3. Third: Remove select/trigger overrides (lines 356-404) -- Shadcn Select handles this
4. Fourth: Remove checkbox/radio overrides (lines 462-476)
5. After each group: verify all forms and dropdowns visually

**Warning signs:** Unstyled native `<select>` elements appearing, inputs losing border-radius, dropdowns missing backgrounds.

### Pitfall 2: Radix Portal Z-Index vs Noise Overlay

**What goes wrong:** The noise overlay at `z-index: 10000` renders on top of Radix-portaled content (dialogs, popovers, dropdowns) if their z-index is lower.

**Why it happens:** Radix portals to `document.body` and uses inline z-index values. If the noise overlay z-index is higher than Radix's, all dropdowns appear behind the grain texture.

**How to avoid:** Use `z-index: 9999` for the noise overlay instead of `10000`. Verify all Radix portaled content (Dialog, Popover, DropdownMenu, Select, Tooltip) renders above the overlay.

### Pitfall 3: Radius System Breaking Existing Components

**What goes wrong:** The current radius system uses `calc(var(--radius) - Npx)` to derive sm/md/lg/xl. Changing to fixed values (6/8/12/16px) breaks any component using `rounded-sm`/`rounded-md`/`rounded-lg`/`rounded-xl` because the semantics change.

**Why it happens:** Current: `--radius: 0.75rem` -> sm=8px, md=10px, lg=12px, xl=16px. New: sm=6px, md=8px, lg=12px, xl=16px. The sm and md values shift by 2px each.

**How to avoid:** Accept the global ripple. Fix any components where the 2px shift matters visually. Most components will look fine or better with the new values.

### Pitfall 4: Accent System Conflict

**What goes wrong:** The accent system overrides `--primary` and `--ring`. If the new design system redefines `--primary` to point at `var(--blue)` in the base theme, the accent override `[data-accent="red"] { --primary: ... }` correctly overrides it. But if any component hardcodes `var(--blue)` instead of `var(--primary)`, it ignores the user's accent preference.

**How to avoid:** Rule: Components that need to respect user accent MUST use `bg-primary`/`text-primary`/`ring-ring` Tailwind utilities, NEVER `text-blue` or hardcoded accent colors. Module-specific semantic colors (tracker=blue, messenger=green) are separate from accent and CAN use named colors.

### Pitfall 5: Light Mode Text Contrast Failure

**What goes wrong:** Derived `--t2` and `--t3` values on white/light backgrounds fail WCAG AA 4.5:1 contrast ratio.

**How to avoid:** The recommended light values (`--t2: #4B5563` at ~7:1, `--t3: #9CA3AF` at ~3:1) are designed for compliance. `--t3` at 3:1 is acceptable ONLY for large text (18px+) and placeholder text. Verify with axe-core after implementation.

### Pitfall 6: Plate Editor Style Leak

**What goes wrong:** After removing `!important` overrides on `input` and `textarea`, the Plate rich-text editor (used in Messenger and Issue descriptions) may lose its styling because it renders native editable elements that previously picked up the global overrides.

**How to avoid:** After `!important` removal, specifically test: (1) Plate editor in Messenger compose area, (2) Plate editor in Issue detail description, (3) inline editing fields. If styling breaks, add scoped styles using `.plate-editor` container class rather than reintroducing global overrides.

## Code Examples

### Exact !important Declarations to Remove (56 total, by target)

**Group 1: Select/Trigger overrides (lines 356-404) -- 24 declarations**
- `select, [data-radix-select-trigger], [role="listbox"]` -- 19 !important props
- `select option` -- 5 !important props

**Group 2: Radix content wrappers (lines 407-431) -- 10 declarations**
- `[data-radix-select-content], [data-radix-popper-content-wrapper] [role="listbox"], [data-radix-menu-content], [data-radix-dropdown-menu-content], [data-radix-popover-content]` -- 6 props
- Hover/highlighted items -- 4 props

**Group 3: Input/textarea (lines 434-459) -- 10 declarations**
- `input:not([type="checkbox"]):not(...)`, `textarea` -- 5 props
- Focus states -- 3 props
- Placeholder -- 2 props

**Group 4: Checkbox/radio (lines 462-476) -- 5 declarations**
- `button[role="checkbox"], button[role="radio"]` -- 3 props
- Checked state -- 2 props

**Group 5: Dark theme specifics (lines 498-517) -- 7 declarations**
- Dark dropdown shadows -- 1 prop
- Dark focus rings -- 3+3 props (inputs + selects combined)

### Custom CSS Classes Audit (keep vs remove)

**Keep and update to use new tokens:**
- `.glass`, `.glass-subtle` -- update rgba values to use new palette
- `.shadow-card` -- replace with `shadow-sm` token
- `.hover-lift` -- update easing to `var(--spring)`
- `.card-premium` -- update to use `var(--card)`, `var(--b1)` etc.
- `.page-header-blur` -- update to new palette colors
- `.board-page-root` -- keep for kanban layout

**Keep as-is (still needed, values are correct):**
- All `@keyframes` animations (fade-in, message-in, typing-bounce, etc.)
- `.typing-dots` -- animation utility
- `.animate-*` classes -- animation utilities
- `.scrollbar-thin` -- update to 5px per mockups
- `.table-striped` -- update to use new tokens
- `.step-connector` -- onboarding specific

**Update significantly:**
- `.badge-passed/failed/skipped/broken/blocked` -- replace hardcoded hex with new palette tokens
- `.gradient-tracker/messenger/tms/calls/dashboard/settings` -- update hex to new palette
- `.gradient-text` -- keep pattern, update colors
- `.greeting-gradient` -- update for new palette
- `.btn-glow`, `.logo-glow`, `.send-glow`, `.composer-glow` -- update blue hex to `var(--blue)`

**Potentially remove (evaluate during implementation):**
- `.module-header-gradient` + strip variants -- may be replaced by new design pattern
- `.card-accent-top` -- may be replaced by new card pattern
- `.hover-pulse` -- may not match new design language

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| HSL color variables | Hex + rgba variables | This phase | All Shadcn compat vars now derived via @theme inline |
| calc-based radius | Fixed radius tokens (r1-r4) | This phase | More predictable, matches mockup pixel values |
| Single `--border` value | 3-level border system (b1/b2/b3) | This phase | Richer visual hierarchy for borders |
| 2-level text (foreground/muted) | 4-level text (t1/t2/t3/t4) | This phase | Finer-grained text hierarchy |
| Geist font (not actually loading) | Inter + JetBrains Mono (CDN) | This phase | Real loaded fonts matching mockups |
| 56 !important declarations | Zero !important, scoped cn() | This phase | Clean specificity, no override wars |

## Open Questions

1. **Noise overlay z-index interaction with Radix portals**
   - What we know: Mockups use z-index 10000. Radix portals render at body level with their own z-index.
   - What's unclear: Exact z-index values used by Radix Portal in this project.
   - Recommendation: Start with z-index 9999 and test all dropdowns/dialogs. Adjust if needed.

2. **Plate editor after !important removal**
   - What we know: Plate renders editable divs/inputs that currently inherit global !important styles.
   - What's unclear: Whether Plate has its own styling that will take over, or if it relies on the global overrides.
   - Recommendation: Test Plate editor immediately after Group 3 (input/textarea) !important removal. Add scoped styles if needed.

3. **Light theme --t3 contrast on small text**
   - What we know: Recommended --t3 (#9CA3AF) has ~3:1 contrast on white -- passes for large text and UI only.
   - What's unclear: Whether any small (<18px) body text in the app uses t3-equivalent styling.
   - Recommendation: Reserve --t3 for placeholders and large labels only. Use --t2 for any readable small text.

## Sources

### Primary (HIGH confidence -- direct codebase and mockup analysis)
- `frontend/src/app.css` (1053 lines) -- full read, all 56 !important declarations mapped
- `frontend/index.html` -- Google Fonts CDN link verified (Geist currently)
- `frontend/src/components/theme-provider.tsx` -- next-themes setup, accent system verified
- `redesign/01-login.html` through `redesign/11-trash.html` -- CSS token values extracted and cross-verified
- `.planning/research/ARCHITECTURE.md` -- migration strategy and token mapping
- `.planning/research/PITFALLS.md` -- 16 pitfalls documented with prevention strategies

### Secondary (HIGH confidence -- project decisions)
- `.planning/phases/08-design-system-foundation/08-CONTEXT.md` -- locked decisions from user discussion
- `.planning/REQUIREMENTS.md` -- DSFN-01..09, THME-01..04 requirement definitions
- `.planning/STATE.md` -- project state and blockers

## Metadata

**Confidence breakdown:**
- Dark theme tokens: HIGH -- directly verified from 11 mockup HTML files
- Light theme tokens: MEDIUM -- derived values (no mockup), need axe-core verification
- !important audit: HIGH -- all 56 declarations mapped by line number and target
- Font swap: HIGH -- exact CDN URL from mockups
- Noise overlay: HIGH -- exact SVG data URI from mockups
- Easing tokens: HIGH -- verified from mockups (--ease, --spring from all; --bounce from TMS mockup)
- Shadow tokens: HIGH -- sh1-sh3 common across mockups; sh4 found in messenger/tms/trash mockups

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (stable -- CSS variables and fonts don't change rapidly)
