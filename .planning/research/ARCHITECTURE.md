# Architecture Research: Hub UI Redesign

**Domain:** Design system migration for React 19 + Tailwind CSS v4 + Shadcn/ui frontend
**Researched:** 2026-03-27
**Confidence:** HIGH (based on direct codebase analysis + mockup inspection)

## Current Architecture

### Design Token System (app.css)

```
@import "tailwindcss"
    |
    v
@theme inline { ... }         -- Maps CSS vars -> Tailwind color-* utilities
    |
    v
:root { ... }                  -- Light theme HSL values (~20 variables)
.dark { ... }                  -- Dark theme HSL values
[data-accent="X"] { ... }     -- 5 accent colors override --primary/--ring
.dark[data-accent="X"] { ... } -- Dark-mode accent overrides
    |
    v
Custom utility classes         -- .glass, .shadow-card, .hover-lift, .card-premium, etc.
```

**Current CSS variable namespace (Shadcn standard):**
- Surface: `--background`, `--card`, `--popover`, `--muted`, `--accent`, `--secondary`
- Text: `--foreground`, `--card-foreground`, `--popover-foreground`, etc.
- Interactive: `--primary`, `--primary-foreground`, `--destructive`, `--ring`, `--border`, `--input`
- Sidebar: `--sidebar-background`, `--sidebar-foreground`, etc. (7 vars)
- Layout: `--radius` (single value, derived into sm/md/lg/xl)

**Total: ~30 CSS variables, HSL format, no alpha channels (except dark borders).**

### Theme Infrastructure

- `next-themes`: class-based dark/light/system via `ThemeProvider`
- `AccentProvider`: `data-accent` attribute on `<html>`, 5 colors (red/blue/green/purple/orange)
- Body font: `Geist, Inter, -apple-system, ...` (system fonts, Geist not actually loaded)
- Code font: none configured (no monospace variable)

### Component Architecture

- 15 Shadcn/ui primitives in `src/components/ui/` (avatar, button, card, dialog, etc.)
- All use `cn()` (clsx + tailwind-merge) for class composition
- All reference Tailwind color utilities mapped to CSS variables (e.g., `bg-primary`, `text-muted-foreground`)
- CVA (class-variance-authority) for button variants
- ~1053 lines in `app.css` with extensive custom utility classes

### Layout Structure

```
AppShell
  +-- Sidebar (260px, always dark)
  +-- Main area
       +-- Topbar (mobile only, hidden on desktop)
       +-- Outlet (page content, px-4 py-4 / px-6)
  +-- MobileNav (bottom bar)
  +-- CallPanel (overlay)
```

---

## New Design System (from Mockups)

### CSS Variable Namespace

```css
:root {
  /* Surface elevation layers (dark values) */
  --void:    #08090B;     /* deepest background */
  --base:    #0D0E12;     /* main content area */
  --raised:  #141519;     /* elevated surfaces */
  --card:    #191A20;     /* card backgrounds */
  --input:   #1E2028;     /* input fields */
  --sidebar: #0B0C0F;     /* sidebar background */

  /* Interactive states */
  --hover:   rgba(255,255,255,.032);
  --active:  rgba(255,255,255,.058);

  /* Text hierarchy */
  --t1: #F2F3F5;   /* primary text */
  --t2: #8C8F97;   /* secondary text */
  --t3: #4E5058;   /* tertiary/placeholder */
  --t4: #363840;   /* disabled/hint */

  /* Border levels */
  --b1: rgba(255,255,255,.045);  /* subtle */
  --b2: rgba(255,255,255,.08);   /* medium */
  --b3: rgba(255,255,255,.13);   /* strong */

  /* Semantic colors */
  --blue: #5B9CF5;    --blueS: #3B7EE0;
  --purple: #A78BFA;  --purpleS: #8B6FF0;
  --green: #34D399;   --greenS: #22C088;
  --amber: #FBBF24;
  --red: #F87171;     --redS: #EF4444;
  --orange: #FB923C;
  --gray: #6B7280;

  /* Glow backgrounds */
  --gB: rgba(91,156,245,.15);   /* blue glow */
  --gP: rgba(167,139,250,.12);  /* purple glow */
  --gG: rgba(52,211,153,.12);   /* green glow */
  --gR: rgba(248,113,113,.1);   /* red glow */

  /* Radii */
  --r1: 6px;  --r2: 8px;  --r3: 12px;  --r4: 16px;  --pill: 100px;

  /* Typography */
  --sans: 'Inter', -apple-system, sans-serif;
  --mono: 'JetBrains Mono', monospace;

  /* Easing */
  --ease: cubic-bezier(.25,.1,0,1);
  --spring: cubic-bezier(.175,.885,.32,1.1);

  /* Shadows */
  --sh1: 0 1px 3px rgba(0,0,0,.5);
  --sh2: 0 4px 16px rgba(0,0,0,.4), 0 1px 4px rgba(0,0,0,.25);
  --sh3: 0 12px 40px rgba(0,0,0,.5);
}
```

### Design Language Differences

| Aspect | Current | New Mockups |
|--------|---------|-------------|
| Surface model | 3 levels (background/card/popover) | 5 levels (void/base/raised/card/input) |
| Text hierarchy | 2 levels (foreground/muted-foreground) | 4 levels (t1/t2/t3/t4) |
| Border system | Single `--border` | 3 levels (b1/b2/b3) |
| Radius | Single `--radius` with calc derivatives | 4 fixed values (6/8/12/16px) + pill |
| Shadows | Custom classes only | 3 design-token shadows (sh1/sh2/sh3) |
| Fonts | System fonts (Geist/Inter fallback) | Inter (loaded) + JetBrains Mono |
| Easing | CSS defaults / manual | Named tokens (ease/spring) |
| Colors | HSL semantic (primary/secondary) | Named semantic (blue/purple/green/amber) |
| Sidebar | Always-dark with separate vars | Own `--sidebar` bg color, consistent system |
| Card style | Flat with subtle border | Gradient bg + top-edge highlight + hover glow |

---

## Recommended Architecture: Design-Tokens-First Migration

### Strategy: Remap, Don't Replace

The critical insight: **Shadcn/ui components reference Tailwind utilities like `bg-card`, `text-foreground`, `border-border`**. These utilities are generated from `@theme inline` which reads CSS variables. If we change the CSS variables, all components update automatically.

**DO NOT rewrite components first.** Instead:

1. Replace CSS variable values in `app.css` to match new design system
2. Add new variables where the current namespace lacks granularity
3. Update `@theme inline` to expose new variables as Tailwind utilities
4. Then refine individual components for mockup-specific styling

### Phase 1: Design Token Foundation

#### File: `src/app.css` (modify existing)

```css
@import "tailwindcss";

@theme inline {
  /* === Existing Shadcn mappings (keep for component compat) === */
  --color-background: var(--base);
  --color-foreground: var(--t1);
  --color-card: var(--card);
  --color-card-foreground: var(--t1);
  --color-popover: var(--raised);
  --color-popover-foreground: var(--t1);
  --color-primary: var(--blue);
  --color-primary-foreground: #fff;
  --color-secondary: var(--raised);
  --color-secondary-foreground: var(--t2);
  --color-muted: var(--raised);
  --color-muted-foreground: var(--t2);
  --color-accent: var(--hover);
  --color-accent-foreground: var(--t1);
  --color-destructive: var(--red);
  --color-destructive-foreground: #fff;
  --color-border: var(--b1);
  --color-input: var(--input);
  --color-ring: var(--blue);

  /* === New design system utilities === */
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

  /* Radii — fixed values, not calc-derived */
  --radius-sm: var(--r1);   /* 6px */
  --radius-md: var(--r2);   /* 8px */
  --radius-lg: var(--r3);   /* 12px */
  --radius-xl: var(--r4);   /* 16px */

  /* Shadows as design tokens */
  --shadow-sm: var(--sh1);
  --shadow-md: var(--sh2);
  --shadow-lg: var(--sh3);
}
```

This approach means:
- `bg-card` continues working in existing components but now renders `#191A20` (dark) instead of `hsl(220 11% 12%)`
- `text-muted-foreground` now maps to `--t2` (#8C8F97)
- New utilities like `bg-void`, `text-t3`, `border-b2` become available for new/refined markup
- No component code changes needed for baseline visual update

#### Variable definitions (`:root` and `.dark`)

The mockups are dark-only. The mapping strategy:

```css
/* Dark theme — new design system (primary) */
.dark, :root[data-force-dark] {
  --void: #08090B;
  --base: #0D0E12;
  --raised: #141519;
  --card: #191A20;
  --input: #1E2028;
  --sidebar-bg: #0B0C0F;
  --hover: rgba(255,255,255,.032);
  --active: rgba(255,255,255,.058);
  --t1: #F2F3F5;
  --t2: #8C8F97;
  --t3: #4E5058;
  --t4: #363840;
  --b1: rgba(255,255,255,.045);
  --b2: rgba(255,255,255,.08);
  --b3: rgba(255,255,255,.13);
  --blue: #5B9CF5;
  --blueS: #3B7EE0;
  --purple: #A78BFA;
  --green: #34D399;
  --amber: #FBBF24;
  --red: #F87171;
  --orange: #FB923C;
  --r1: 6px; --r2: 8px; --r3: 12px; --r4: 16px;
  --ease: cubic-bezier(.25,.1,0,1);
  --spring: cubic-bezier(.175,.885,.32,1.1);
  --sh1: 0 1px 3px rgba(0,0,0,.5);
  --sh2: 0 4px 16px rgba(0,0,0,.4), 0 1px 4px rgba(0,0,0,.25);
  --sh3: 0 12px 40px rgba(0,0,0,.5);
}

/* Light theme — derived from dark with inverted palette */
:root {
  --void: #FFFFFF;
  --base: #F8F9FA;
  --raised: #F1F3F5;
  --card: #FFFFFF;
  --input: #F1F3F5;
  --sidebar-bg: #1A1D21;  /* sidebar stays dark */
  --hover: rgba(0,0,0,.03);
  --active: rgba(0,0,0,.06);
  --t1: #111827;
  --t2: #6B7280;
  --t3: #9CA3AF;
  --t4: #D1D5DB;
  --b1: rgba(0,0,0,.06);
  --b2: rgba(0,0,0,.10);
  --b3: rgba(0,0,0,.16);
  --blue: #3B7EE0;
  --blueS: #2563EB;
  --purple: #8B6FF0;
  --green: #22C088;
  --amber: #D97706;
  --red: #EF4444;
  --orange: #EA580C;
  --r1: 6px; --r2: 8px; --r3: 12px; --r4: 16px;
  --ease: cubic-bezier(.25,.1,0,1);
  --spring: cubic-bezier(.175,.885,.32,1.1);
  --sh1: 0 1px 3px rgba(0,0,0,.08);
  --sh2: 0 4px 16px rgba(0,0,0,.08), 0 1px 4px rgba(0,0,0,.04);
  --sh3: 0 12px 40px rgba(0,0,0,.12);
}
```

### Phase 2: Typography + Fonts

#### File: `index.html` (modify)

Add Google Fonts import:
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

#### File: `src/app.css` (modify body rule)

```css
body {
  font-family: var(--sans);  /* 'Inter', -apple-system, sans-serif */
  font-size: 14px;
  line-height: 1.5;
}
```

Add to `@theme inline`:
```css
--font-sans: var(--sans);
--font-mono: var(--mono);
```

### Phase 3: Accent Color Remap

Current accent system overrides `--primary` via `[data-accent="X"]`. This continues working because the `@theme inline` maps `--color-primary: var(--blue)` by default, and accent overrides replace `--primary`.

**Change needed:** Rename accent internal mapping to use new semantic colors:

```css
[data-accent="blue"] {
  --primary-hue: var(--blue);
  --ring-hue: var(--blue);
}
[data-accent="purple"] {
  --primary-hue: var(--purple);
  --ring-hue: var(--purple);
}
/* etc. */
```

But simpler: **keep the current accent mechanism as-is.** It already works by overriding `--primary` and `--ring`. The new design system's `--blue`/`--purple`/etc. are semantic module colors (tracker=blue, messenger=green), not accent colors. These are separate concerns.

### Phase 4: Layout Shell Redesign

Mockup layout structure:

```
.app (flex, h-screen)
  +-- .sb (260px, sidebar-bg, border-right: b1)
  |     +-- .sb-h (56px header, logo + brand)
  |     +-- .sb-nav (flex-1, scrollable)
  |     +-- .sb-u (user section, border-top)
  +-- .mn (flex-1, flex-col)
        +-- .topbar (56px, backdrop-blur, border-bottom: b1)
        +-- .page (flex-1, overflow-y: auto, padding: 28px)
```

**Key differences from current:**
1. Topbar is always visible (current: mobile-only)
2. Topbar has backdrop blur + gradient bottom edge
3. Sidebar has decorative gradient edge (`::after`)
4. Content padding: 28px (current: 16px/24px)
5. Background uses `--base` not `--background`

### Phase 5: Component-by-Component Refinement

After tokens are in place, individual Shadcn/ui components get refined:

**High-priority component updates:**
1. `Button` — add gradient variants, glow effects, spring animations
2. `Card` — gradient backgrounds (`linear-gradient(165deg, var(--card), var(--raised))`), top-edge highlights, hover glow
3. `Input` — use `--input` bg, `--b1` border, blue focus glow
4. `Badge` — status colors with glow backgrounds (`--gB`, `--gR`, etc.)
5. `Dialog` — modal entrance animation, backdrop blur

**Medium-priority:**
6. `Tabs` — active indicator with amber glow (sidebar style)
7. `DropdownMenu` — entrance animation, hover states
8. `ScrollArea` — 5px slim scrollbar
9. `Avatar` — gradient backgrounds, online indicator with glow
10. `Tooltip` — match new popover style

---

## Component Boundaries

| Component | Responsibility | Files Modified | Files Created |
|-----------|---------------|----------------|---------------|
| Design tokens | CSS variables, `@theme inline` | `src/app.css` | None |
| Typography | Font loading, base styles | `index.html`, `src/app.css` | None |
| AppShell | Layout structure, topbar visibility | `src/components/layout/app-shell.tsx` | None |
| Sidebar | New styling, gradient edge, active indicator | `src/components/layout/sidebar.tsx`, `sidebar-link.tsx` | None |
| Topbar | Always visible, backdrop blur, gradient edge | `src/components/layout/topbar.tsx` | None |
| UI primitives | Button, Card, Input, Badge updates | `src/components/ui/*.tsx` | None |
| Page layouts | Per-page styling to match mockups | `src/modules/**/*.tsx` | None |
| Animations | Entrance animations, spring easing | `src/app.css` | None |

**Key insight: Zero new files needed for the design system foundation.** All changes modify existing files. The design token layer, typography, and layout updates all live in `app.css` and existing components.

---

## Data Flow for Theming

```
User selects theme (light/dark/system)
  |
  v
next-themes sets class="dark" on <html>
  |
  v
CSS cascade resolves :root vs .dark variables
  |
  v
@theme inline maps vars to Tailwind utilities
  |
  v
All bg-*, text-*, border-* classes update automatically
  |
  v
User selects accent color
  |
  v
AccentProvider sets data-accent="X" on <html>
  |
  v
[data-accent="X"] overrides --primary and --ring
  |
  v
All bg-primary, text-primary references update
```

**This data flow does not change.** The redesign only changes the VALUES inside the CSS variables, not the mechanism. This is the strongest argument for design-tokens-first approach.

---

## Patterns to Follow

### Pattern 1: Surface Elevation

**What:** Use the 5-level elevation system instead of arbitrary backgrounds.
**When:** Any container, card, section, or overlay needs a background.

```
void    → page-level background, deepest layer
base    → main content area
raised  → elevated panels, secondary surfaces
card    → cards, dialogs, distinct containers
input   → form inputs, textareas, code blocks
```

```tsx
// Good
<div className="bg-void">
  <main className="bg-base">
    <Card className="bg-card" />
  </main>
</div>

// Bad — arbitrary colors
<div className="bg-[#0D0E12]">
```

### Pattern 2: Text Hierarchy

**What:** Use the 4-level text system for visual hierarchy.
**When:** Any text needs to establish importance.

```
t1  → primary content, headings, names
t2  → secondary info, descriptions, labels
t3  → placeholders, timestamps, hints
t4  → disabled text, barely visible dividers
```

```tsx
<h2 className="text-t1 font-bold">Project Name</h2>
<p className="text-t2">Last updated 2 hours ago</p>
<span className="text-t3">No description</span>
```

### Pattern 3: Border Levels

**What:** Use 3 border levels for visual depth.
**When:** Drawing lines between sections or around containers.

```
b1  → subtle separator (between list items)
b2  → medium border (card outlines, hover state)
b3  → strong border (focus rings, active elements)
```

### Pattern 4: Gradient Cards (Mockup Style)

**What:** Cards use subtle gradient backgrounds with top-edge highlights.
**When:** Interactive cards (widgets, project cards, kanban items).

```tsx
<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-xl
  relative overflow-hidden hover:border-b2 hover:shadow-md transition-all">
  {/* Top edge highlight */}
  <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r
    from-transparent via-white/[.04] to-transparent" />
  {children}
</div>
```

### Pattern 5: Spring Animations

**What:** Use custom easing for interactive elements.
**When:** Hover transforms, entrance animations.

```css
.spring-hover {
  transition: all 250ms var(--spring);
}
.spring-hover:hover {
  transform: translateY(-3px);
}
```

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Hardcoded Colors

**What:** Using hex/rgb values directly in components.
**Why bad:** Breaks theme switching, unmaintainable.
**Instead:** Always reference CSS variables via Tailwind utilities.

```tsx
// Bad
<div style={{ background: '#191A20' }}>
// Bad
<div className="bg-[#191A20]">

// Good
<div className="bg-card">
```

### Anti-Pattern 2: Parallel Variable Systems

**What:** Keeping both old Shadcn HSL variables AND new hex variables active.
**Why bad:** Confusion about which to use, double maintenance, visual inconsistencies.
**Instead:** Map Shadcn semantic names to new design tokens in `@theme inline`. One source of truth.

### Anti-Pattern 3: Component-First Migration

**What:** Rewriting individual page components to match mockups before establishing tokens.
**Why bad:** Each component invents its own color values, inconsistency, and need to redo everything when tokens are finalized.
**Instead:** Tokens first. After `app.css` is updated, 60-70% of the visual change happens automatically through the cascade.

### Anti-Pattern 4: !important Overrides

**What:** Current `app.css` has extensive `!important` on form elements and Radix overrides.
**Why bad:** Creates specificity wars, makes component customization harder.
**Instead:** During redesign, remove `!important` declarations and use proper specificity via Tailwind's `@layer` or more specific selectors.

### Anti-Pattern 5: Global Element Selectors for Styling

**What:** Current `app.css` has `input`, `select`, `textarea` with broad `!important` overrides.
**Why bad:** Fights with Shadcn component styles, causes unexpected behavior.
**Instead:** Style through component props and Tailwind classes. Remove global form element overrides.

---

## Migration Sequence (Build Order)

```
Phase 1: Foundation (no visual breakage)
  1.1  Add Inter + JetBrains Mono font loading to index.html
  1.2  Define new CSS variables (:root and .dark) in app.css
  1.3  Remap @theme inline to point at new variables
  1.4  Update body font-family
  1.5  Light theme: derive inverted values for all new vars
  >>> At this point: entire app shifts to new palette automatically

Phase 2: Layout Shell
  2.1  AppShell: add always-visible topbar on desktop
  2.2  Sidebar: new background, gradient edge, active indicator style
  2.3  Topbar: backdrop blur, gradient bottom edge
  2.4  Content area: adjust padding to 28px
  >>> At this point: chrome matches mockups

Phase 3: UI Primitives
  3.1  Button: gradient variants, glow, spring easing
  3.2  Card: gradient bg, top-edge highlight, hover glow
  3.3  Input: new focus style (blue glow, scale 1.005)
  3.4  Badge: status glow backgrounds
  3.5  Dialog/Popover: entrance animations, backdrop blur
  3.6  Remove !important overrides from form element global styles
  >>> At this point: shared components match design language

Phase 4: Page-by-Page Matching (11 mockups)
  4.1  Login page
  4.2  Dashboard (widget grid, quick actions, greeting gradient)
  4.3  Messenger (message list, composer, channel list)
  4.4  Kanban (columns, cards, swimlanes)
  4.5  Issue Detail (fields, comments, sidebar)
  4.6  Calls (participants grid, controls)
  4.7  TMS (test case list, run details)
  4.8  Settings (form sections, toggles)
  4.9  Admin (tables, analytics)
  4.10 Onboarding (step indicator, forms)
  4.11 Trash (restore actions, filters)
  >>> At this point: all mockup pages match

Phase 5: Non-Mockup Pages
  5.1  Projects list
  5.2  Sprint management
  5.3  JQL search
  5.4  Project settings
  5.5  Project dashboard
  >>> Styled to match overall design language without explicit mockups

Phase 6: Polish
  6.1  Animation timing audit
  6.2  Light mode visual QA
  6.3  Mobile responsive QA
  6.4  Accessibility audit (contrast ratios)
  6.5  Clean up dead CSS classes from app.css
```

---

## Scalability Considerations

| Concern | Current (15 UI components) | At 30+ components | At full feature set |
|---------|---------------------------|---------------------|---------------------|
| Token consistency | Easy — single app.css | Enforce via Tailwind utilities only | Lint rule: no hardcoded colors |
| Dark/light parity | Manual in app.css | Same approach, more variables | Consider CSS color-mix() for auto-derivation |
| Component variants | CVA in each component | CVA continues to work | Consider extracting shared variant configs |
| Animation perf | CSS animations | Same | Monitor paint/composite layers, use will-change sparingly |

---

## Key Integration Points

### 1. Shadcn/ui Component Compatibility

All 15 existing Shadcn components use these Tailwind utilities:
- `bg-background`, `bg-card`, `bg-popover`, `bg-primary`, `bg-secondary`, `bg-muted`, `bg-accent`, `bg-destructive`
- `text-foreground`, `text-card-foreground`, `text-primary-foreground`, `text-muted-foreground`
- `border-border`, `border-input`
- `ring-ring`

**These all resolve through `@theme inline`.** Changing the `@theme inline` mappings is the single point of control. No component code changes needed for basic color updates.

### 2. Custom CSS Classes in app.css

Current `app.css` has ~50 custom classes (.glass, .card-premium, .hover-lift, .badge-passed, etc.). These use a mix of:
- Hardcoded hex values (`.badge-passed { background: #dcfce7; }`)
- CSS variable references (`.card-premium { background: var(--card); }`)

**Migration:** Replace hardcoded values with new design token references. Some classes will be replaced by new equivalents from the mockup design language.

### 3. AccentProvider / data-accent

Works independently of the design system migration. Accent overrides `--primary` and `--ring` which continue to exist in the new system. **No changes needed.**

### 4. next-themes / class="dark"

The `.dark` selector continues to work. New variables are defined under `.dark` instead of `:root` (since mockups are dark-first). **No changes needed to ThemeProvider.**

### 5. Tailwind CSS v4 @theme inline

Tailwind v4 generates utilities from `@theme inline` declarations. Adding new `--color-*` entries (void, base, raised, t1-t4, b1-b3) automatically creates:
- `bg-void`, `bg-base`, `bg-raised`
- `text-t1`, `text-t2`, `text-t3`, `text-t4`
- `border-b1`, `border-b2`, `border-b3`

**No Tailwind config file changes needed.** Everything is driven by CSS.

---

## Sources

- Direct analysis of `frontend/src/app.css` (1053 lines)
- Direct analysis of 11 HTML mockups in `redesign/` directory
- Direct analysis of `frontend/src/components/ui/*.tsx` (15 components)
- Direct analysis of `frontend/src/components/theme-provider.tsx`
- Direct analysis of `frontend/src/components/layout/app-shell.tsx`
- Direct analysis of `frontend/vite.config.ts`
- Tailwind CSS v4 `@theme inline` behavior (from `@tailwindcss/vite` plugin)
