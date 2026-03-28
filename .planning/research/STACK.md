# Technology Stack: Hub v2.0 UI Redesign

**Project:** Hub Pixel-Perfect Redesign
**Researched:** 2026-03-27
**Overall confidence:** HIGH

## Current State

The frontend already has a mature stack: React 19, Tailwind CSS v4, Shadcn/ui, `motion` (ex framer-motion), `next-themes`, `class-variance-authority`. The CSS design system in `app.css` already uses CSS custom properties with `:root` / `.dark` pattern, accent color overrides via `[data-accent]`, glassmorphism utilities, gradient classes, and basic keyframe animations.

Current fonts: **Geist + Geist Mono** (Google Fonts CDN).
Redesign fonts: **Inter + JetBrains Mono** (per all 11 HTML mockups).

## What Changes

### 1. Font Swap (ZERO new dependencies)

| Change | From | To | How |
|--------|------|----|-----|
| Primary font | Geist | Inter | Swap Google Fonts URL in `index.html` |
| Monospace font | Geist Mono | JetBrains Mono | Same URL swap |
| Font weight range | 300-700 | 400-900 | Mockups use `font-weight: 900` for headings |

**Action:** Replace the Google Fonts `<link>` in `index.html`:
```html
<!-- FROM -->
<link href="https://fonts.googleapis.com/css2?family=Geist:wght@300;400;500;600;700&family=Geist+Mono:wght@400;500&display=swap" rel="stylesheet" />

<!-- TO -->
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
```

Update `app.css` body `font-family`:
```css
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
```

Add CSS variables matching mockups:
```css
:root {
  --sans: 'Inter', -apple-system, sans-serif;
  --mono: 'JetBrains Mono', monospace;
}
```

**Why Google Fonts CDN (not self-hosted @fontsource):** The project already uses Google Fonts CDN for Geist. Self-hosting adds npm packages and build complexity for marginal gain. Google Fonts CDN has global edge caching. For an internal tool with 10-50 users, CDN latency is irrelevant. Do NOT add `@fontsource/inter` or similar packages.

**Confidence:** HIGH -- direct observation of mockup source code and current `index.html`.

### 2. CSS Custom Properties Overhaul (ZERO new dependencies)

The mockups define a different token system than the current Shadcn HSL variables. The redesign palette:

| Mockup Token | Value | Maps To |
|-------------|-------|---------|
| `--void` | `#08090B` | `--background` (dark) |
| `--base` | `#0D0E12` | Sidebar/secondary backgrounds |
| `--raised` | `#141519` | Elevated surfaces |
| `--card` | `#191A20` | `--card` (dark) |
| `--input` | `#1E2028` | `--input` (dark) |
| `--t1` | `#F2F3F5` | `--foreground` (dark) |
| `--t2` | `#8C8F97` | `--muted-foreground` (dark) |
| `--t3` | `#4E5058` | Tertiary text |
| `--t4` | (varies) | Quaternary text |
| `--b1` | `rgba(255,255,255,.045)` | `--border` (dark) |
| `--b2` | `rgba(255,255,255,.08)` | Secondary border |
| `--blue` | `#5B9CF5` | `--primary` (dark) |
| `--purple` | `#A78BFA` | Accent purple |
| `--green` | `#34D399` | Success states |
| `--red` | `#F87171` | `--destructive` (dark) |

**Strategy:** Map mockup tokens to existing Shadcn variable names in `:root` and `.dark` blocks. Add new semantic tokens where mockups introduce concepts not in current system (e.g., `--t3`, `--t4`, `--raised`, `--void`). This preserves all Shadcn component compatibility while delivering the new palette.

**The current `@theme inline` block stays.** Tailwind v4 reads `--color-*` from there. The values flow through: `:root`/`.dark` CSS vars -> `@theme inline` -> Tailwind utility classes. No Tailwind config changes needed.

**New tokens to add to `@theme inline`:**
```css
@theme inline {
  /* existing mappings stay */
  --color-void: var(--void);
  --color-raised: var(--raised);
  --color-t3: var(--t3);
  --color-t4: var(--t4);
  --color-b2: var(--b2);
}
```
This enables `bg-void`, `bg-raised`, `text-t3` utility classes automatically.

**Confidence:** HIGH -- verified by examining both current `app.css` and all 11 mockup CSS blocks.

### 3. Animation System (ALREADY INSTALLED: `motion` v12)

`motion` (formerly framer-motion) is already in `package.json` at `^12.38.0`. Currently used only in `board-column.tsx`. The redesign needs it extensively.

**Mockup animation patterns (all achievable with `motion` or pure CSS):**

| Pattern | Mockup Usage | Implementation |
|---------|-------------|----------------|
| Spring easing | `cubic-bezier(.175,.885,.32,1.1)` | CSS `--spring` variable, no JS needed |
| Smooth easing | `cubic-bezier(.25,.1,0,1)` | CSS `--ease` variable, no JS needed |
| Card float | `translateY` oscillation, 7-9s infinite | Pure CSS `@keyframes` |
| Mesh shift | Hue-rotate background, 20s infinite | Pure CSS `@keyframes` |
| Enter-up | Fade + translateY on mount | `motion` `initial`/`animate` or CSS `@keyframes` |
| Hover lift | `translateY(-6px)` + shadow bloom | Pure CSS `:hover` transition |
| Click scale | `scale(.97)` on `:active` | Pure CSS `:active` |
| Staggered children | Sequential widget appearance | `motion` `staggerChildren` |
| Progress bar fill | Width animation from 0 | CSS `@keyframes` or `motion` |

**Decision: CSS-first, `motion` for orchestrated/interactive animations.**

- Simple transitions (hover, active, focus): Pure CSS with `--ease` and `--spring` variables
- Mount/unmount animations: `motion` `AnimatePresence` + `motion.div`
- Staggered lists: `motion` `staggerChildren`
- Layout animations: `motion` `layoutId` (already have the library)
- Infinite ambient (float, mesh): Pure CSS `@keyframes`

**CSS easing variables to add:**
```css
:root {
  --ease: cubic-bezier(.25, .1, 0, 1);
  --spring: cubic-bezier(.175, .885, .32, 1.1);
}
```

**Do NOT add:** `react-spring`, `gsap`, `animejs`, `@formkit/auto-animate`, `framer-motion` (old name -- `motion` is the current package and already installed).

**Confidence:** HIGH -- `motion` v12 is already installed, all mockup animations are achievable with CSS + `motion`.

### 4. Glassmorphism & Visual Effects (ZERO new dependencies)

All effects in the mockups are pure CSS:

| Effect | CSS Technique | Already in `app.css`? |
|--------|--------------|----------------------|
| Glass panels | `backdrop-filter: blur(20px)` + semi-transparent bg | YES (`.glass` class exists) |
| Glow on focus | `box-shadow: 0 0 0 3px rgba(color, .08), 0 0 20px rgba(color, .04)` | Partially (focus-gradient exists) |
| Gradient borders | `::before` pseudo with gradient background | NO -- add as utility |
| Noise texture overlay | SVG filter via `body::after` | NO -- add as utility |
| Gradient text | `background-clip: text` + `-webkit-text-fill-color` | YES (`.gradient-text` exists) |
| Dot glow on hover | `box-shadow: 0 0 8px currentColor` | NO -- add as utility |
| Card top highlight | `::before` with horizontal gradient line | NO -- add as utility |
| Grid lines background | Linear gradient grid + radial mask | NO -- add as utility |

**New CSS utilities to add to `app.css`:**
```css
/* Noise texture overlay (from mockup body::after) */
.noise-overlay::after { content: ''; position: fixed; inset: 0; pointer-events: none; opacity: .3; background: url("data:image/svg+xml,..."); }

/* Gradient top border on cards */
.gradient-border-top::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 1px; background: linear-gradient(90deg, transparent, rgba(255,255,255,.06), transparent); }

/* Dot/indicator glow */
.dot-glow { transition: box-shadow 200ms; }
.dot-glow:hover { box-shadow: 0 0 8px currentColor; }

/* Grid lines background (login page) */
.grid-lines { background-image: linear-gradient(rgba(255,255,255,.015) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.015) 1px, transparent 1px); background-size: 64px 64px; }
```

**Confidence:** HIGH -- all effects verified in mockup source. Pure CSS, no libraries needed.

### 5. Tailwind CSS v4 Configuration (NO PLUGINS NEEDED)

Tailwind v4 uses CSS-first configuration via `@theme`. The current setup is correct and sufficient.

**What needs updating in `@theme inline`:**

- New radius values: mockups use `6px`, `8px`, `12px`, `16px`, `100px` (pill)
- Current radius: `calc(var(--radius) - 4px)` etc. -- switch to explicit values matching mockups
- Add shadow presets if needed via `--shadow-*` tokens

**Updated radius tokens:**
```css
@theme inline {
  --radius-sm: 6px;   /* was: calc(var(--radius) - 4px) */
  --radius-md: 8px;   /* was: calc(var(--radius) - 2px) */
  --radius-lg: 12px;  /* was: var(--radius) */
  --radius-xl: 16px;  /* was: calc(var(--radius) + 4px) */
  --radius-pill: 100px; /* NEW */
}
```

**Do NOT add Tailwind plugins:**
- `@tailwindcss/typography` -- not needed, mockups don't use prose
- `tailwindcss-animate` -- not needed with Tailwind v4 (has built-in `animate-*` + we have `motion`)
- `tailwind-scrollbar` -- mockups use native scrollbar styling via CSS

**Confidence:** HIGH -- Tailwind v4 CSS-first config is already in place.

## Recommended Stack Additions: NONE

The entire redesign requires **zero new npm packages**. Everything is achievable with the existing stack:

| Need | Solution | Already Have |
|------|----------|-------------|
| Fonts | Google Fonts CDN swap | YES (just change URL) |
| Animations (simple) | CSS custom properties + keyframes | YES |
| Animations (complex) | `motion` v12 | YES (`^12.38.0`) |
| Theme switching | `next-themes` | YES (`^0.4.6`) |
| CSS variables | Native CSS custom properties | YES |
| Glassmorphism | `backdrop-filter` CSS | YES |
| Gradient effects | CSS gradients | YES |
| Component variants | `class-variance-authority` | YES (`^0.7.1`) |
| Class merging | `tailwind-merge` + `clsx` | YES |
| Dark/light mode | `:root` / `.dark` CSS + `next-themes` | YES |

### Libraries to Leverage More (Already Installed)

| Library | Version | Current Usage | Redesign Usage |
|---------|---------|---------------|----------------|
| `motion` | ^12.38.0 | 1 file (board-column) | Page transitions, staggered lists, mount animations, layout animations |
| `class-variance-authority` | ^0.7.1 | Component variants | Extended variants for new design states |
| `next-themes` | ^0.4.6 | Theme toggle | Same, ensure CSS variables work for both themes |
| `tailwind-merge` | latest | Class merging | Same |
| `lucide-react` | ^0.577.0 | Icons | Same icons, possibly resized per mockups (18px standard, 22px for featured) |

## Integration Notes

### Shadcn/ui Compatibility

The redesign changes CSS variables, not component logic. Shadcn components use `bg-background`, `text-foreground`, `bg-card`, etc. -- these map to CSS custom properties. Changing the property values automatically reskins all Shadcn components.

**Critical:** Keep the Shadcn variable naming convention (`--background`, `--foreground`, `--card`, etc.). Add new variables alongside, don't rename existing ones.

### `next-themes` Integration

Current: works with `.dark` class on `<html>`. Mockups are dark-only but the project needs light + dark. Strategy:
1. Redesign dark theme first (directly from mockups)
2. Derive light theme by creating complementary values for every CSS variable
3. `next-themes` continues toggling `.dark` class -- no changes needed to the library usage

### Font Loading Performance

`font-display: swap` already set via `&display=swap` param. `<link rel="preconnect">` already in `index.html`. No FOUT concerns for internal tool.

### Accent Color System

Current `[data-accent="blue"]` etc. overrides `--primary`. This stays. Update the accent HSL values to match the new palette. The mockup blue (`#5B9CF5` ~ `hsl(214, 88%, 66%)`) differs from current dark blue (`hsl(217, 91%, 60%)`).

## What NOT to Add

| Library | Why NOT |
|---------|---------|
| `@fontsource/inter` / `@fontsource/jetbrains-mono` | Google Fonts CDN is already the pattern, self-hosting adds complexity for zero benefit on internal tool |
| `react-spring` | `motion` already installed, does everything needed |
| `gsap` | Overkill for UI animations, `motion` + CSS covers all mockup patterns |
| `@formkit/auto-animate` | Too magic, less control than `motion` |
| `tailwindcss-animate` | Tailwind v4 has built-in animations, plus we have `motion` |
| `@tailwindcss/typography` | No prose/markdown rendering in redesign scope |
| `tailwind-scrollbar` | Native CSS `::-webkit-scrollbar` sufficient |
| `styled-components` / `emotion` | Tailwind + CSS vars is the established pattern |
| `postcss-preset-env` | Tailwind v4 handles modern CSS |
| `open-props` | Already have custom design tokens, adding another system creates confusion |
| `radix-colors` | Already have Shadcn color system, don't mix token systems |
| Any CSS-in-JS | Against established Tailwind pattern |

## Implementation Checklist

1. **Swap fonts** in `index.html` (Inter weights 400-900, JetBrains Mono weights 400-600)
2. **Update `app.css`** body font-family to Inter
3. **Add CSS easing variables** (`--ease`, `--spring`, `--sans`, `--mono`)
4. **Overhaul `:root` and `.dark`** color values to match mockup palette
5. **Add new semantic tokens** (`--void`, `--raised`, `--t3`, `--t4`, `--b2`) to both CSS and `@theme inline`
6. **Update radius tokens** to explicit pixel values matching mockups
7. **Add new CSS utility classes** (noise overlay, gradient borders, dot glow, grid lines, card top highlight)
8. **Keep all existing Shadcn variable names** for component compatibility
9. **Extend `motion` usage** for page transitions and staggered animations
10. **Derive light theme** values complementary to the new dark palette

## Sources

- `frontend/index.html` line 12 -- current Google Fonts CDN link (Geist)
- `frontend/src/app.css` -- full CSS design system (380+ lines, HSL variables, glass/gradient utilities)
- `frontend/package.json` -- `motion ^12.38.0` already installed, no gaps
- All 11 HTML mockups in `redesign/` -- CSS analysis: Inter + JetBrains Mono fonts, `--void`/`--base`/`--raised`/`--card`/`--input` palette, `--ease`/`--spring` easing, `backdrop-filter` glassmorphism, gradient borders, noise textures
- Tailwind CSS v4 `@theme inline` pattern -- verified working in current `app.css`
