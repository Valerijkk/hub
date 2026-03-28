# Domain Pitfalls: Hub UI Redesign (Milestone 2)

**Domain:** Pixel-perfect visual redesign of existing React + Tailwind CSS 4 + Shadcn/ui application
**Researched:** 2026-03-27
**Confidence:** HIGH (codebase analysis of 235 source files + app.css audit + community patterns + official docs)

## Critical Pitfalls

Mistakes that cause rewrites, broken functionality, or extended development cycles.

---

### Pitfall 1: The `!important` Minefield in app.css (56 Declarations)

**What goes wrong:** The current `app.css` contains **56 `!important` declarations** targeting Radix data attributes (`[data-radix-select-trigger]`, `[data-radix-select-content]`, etc.), native form elements (`input:not([type="checkbox"])...`), and Radix dropdown/popover/menu content wrappers. When introducing new design tokens or restyling Shadcn components, these `!important` overrides create an unwinnable specificity war. New Tailwind utilities and component-level styles silently lose to the global `!important` rules.

**Why it happens:** The `!important` rules were added incrementally during initial development to force consistency across Radix-powered components. They target extremely broad selectors -- all inputs, all selects, all popper content wrappers. Any new styling applied through Tailwind classes or Shadcn variant props on those same elements gets overridden without warning.

**Consequences:**
- New button/input/select styles from redesign mockups will not apply without adding yet more `!important`, escalating the war
- Every Radix-based component (Dialog, Popover, DropdownMenu, Select, Tooltip) has its styling locked by global overrides -- restyling requires finding and removing the correct `!important` first
- Theme switching (light/dark) breaks when `!important` pins hardcoded `box-shadow` values like `rgba(59, 130, 246, 0.1)` that assume a specific theme
- Third-party components (Plate editor, Tremor charts, LiveKit UI) that use standard form elements inherit these forced styles

**Prevention:**
1. Phase 1 must begin by auditing and systematically removing all `!important` declarations
2. Move Radix-specific overrides into their respective Shadcn component files (`select.tsx`, `popover.tsx`, `dropdown-menu.tsx`) using the `cn()` merge utility
3. Use Tailwind's `@layer components` for custom utilities instead of global CSS with `!important`
4. Establish a zero-`!important` policy for the redesigned codebase. If `!important` is needed, the architecture is wrong
5. Test Plate editor, Tremor charts, and LiveKit call UI after every global CSS removal -- they are the most likely to have silently depended on these overrides

**Detection:** `grep -c "!important" app.css` after each phase. Count must decrease monotonically toward zero.

**Phase:** Phase 1 (Foundation). Must be completed before any component restyling begins.

---

### Pitfall 2: Breaking Event Handlers and State by Restructuring JSX

**What goes wrong:** Developer changes a component's DOM structure to match the mockup (wrapping elements in new divs, moving buttons, restructuring flex layouts, changing scroll containers) and silently breaks: click handlers that relied on event bubbling, `ref` attachments for scroll position or focus, form submission flows, drag-and-drop drop zones (pragmatic-drag-and-drop), keyboard navigation chains, or conditional rendering gates.

**Why it happens:** Visual redesign feels like "just CSS work" but mockups often imply structural changes: a button moves from inside a card to a floating action bar, a sidebar becomes a slide-over panel, scrollable areas change nesting. Each structural change risks detaching behavior from the DOM node it was bound to. The current app has 235 TSX/TS files with significant interactive logic.

**Consequences:**
- Kanban drag-and-drop stops working (pragmatic-dnd relies on specific DOM container relationships and drop zone configuration)
- Messenger infinite scroll breaks when the scroll container element changes or gets wrapped
- Modal/dialog focus trapping breaks when Radix Portal mount points shift
- Form elements disconnect from their `<form>` parent
- Click-outside detection in dropdowns breaks when wrapper elements change
- Sidebar collapse/expand animation breaks when parent layout restructures

**Prevention:**
1. **Golden rule: restyle with classes first, restructure JSX only when absolutely necessary.** If the mockup can be achieved with different Tailwind classes on existing elements, do not touch JSX structure
2. When JSX changes are required, manually test ALL interactive behaviors on that page before marking complete
3. Keep a per-page interactive behavior checklist (below)
4. Use `data-testid` attributes on functional elements, not structural wrappers

**Detection:** Run Playwright E2E tests after every component change. If no E2E tests exist for a page, write smoke tests before restyling it.

**Phase:** Every phase. This is continuous discipline, not a one-time fix.

**Per-page verification checklist:**
| Page | Must Verify |
|------|-------------|
| Kanban | Drag cards between columns, drag columns, WIP limit indicators, filter toggle, swimlane switch |
| Messenger | Send message, open thread, reactions/emoji picker, infinite scroll up, voice message record, @mention autocomplete, file upload drag-and-drop |
| Calls | Join/leave call, mute/unmute, screen share start/stop, raise hand, participant list interaction, invite link copy |
| Issue Detail | Inline field editing, status change dropdown, comment submit, time tracker start/stop, file upload, parent/child link |
| TMS | Create test case, run tests, mark results, folder tree navigation, drag test cases into runs |
| Dashboard | Widget drag-and-drop reorder, widget resize, filter dropdowns, quick action buttons |
| Settings/Admin | All form submissions, toggle switches, role assignment, tab navigation |
| Login | Magic link flow, Telegram auth, form validation |
| Sidebar | Collapse/expand, project switcher, navigation highlights, mobile bottom nav |

---

### Pitfall 3: CSS Variable Atomic Swap vs. Gradual Migration

**What goes wrong:** The redesign introduces a new color palette (e.g., `#08090B` backgrounds, new radius system 6/8/12/16px, 4-level shadow system, spring animations). If variables are updated gradually (some pages use new tokens, some use old), components using old tokens look visually broken next to components using new tokens. Users see jarring color jumps navigating between pages.

**Why it happens:** Shadcn's convention uses the same global CSS variable names (`--background`, `--foreground`, `--primary`, etc.). You cannot have `--background: hsl(0 0% 100%)` for old components and `--background: hsl(220 13% 7%)` for new ones simultaneously -- they share a single `:root` declaration.

**Consequences:**
- Half the app looks "old design" and half looks "new design" for weeks during migration
- Shared components (Button, Badge, Avatar, Input) that appear on every page create version mismatch
- The accent color system (`[data-accent="red"]` overriding `--primary`) may conflict if `--primary` semantics change
- The "always-dark sidebar" (`--sidebar-*` variables in `:root`) complicates any variable restructuring

**Prevention:**
1. **Swap ALL CSS variables atomically in Phase 1.** Change `:root` and `.dark` variable blocks completely in a single commit. Accept that all pages shift colors at once
2. Then fix individual components page-by-page in subsequent phases to achieve pixel-perfect mockup match
3. Update accent color selectors (`[data-accent="red"]`, `.dark[data-accent="red"]`, etc.) in the same commit as the main variable swap
4. Keep `--sidebar-*` variables as a separate concern -- the sidebar is "always dark" in both themes and uses its own token namespace

**Detection:** Visual walkthrough of all 11+ pages immediately after the variable swap. Screenshot before/after for comparison.

**Phase:** Phase 1 exclusively. All CSS variables updated in one atomic change.

---

### Pitfall 4: Shadcn/ui Component Re-theming Breaks Radix Primitives

**What goes wrong:** Developer modifies a Shadcn component file (e.g., `dialog.tsx`, `select.tsx`, `dropdown-menu.tsx`) to match mockup styling and inadvertently removes or overrides Radix primitive props that control accessibility or behavior: `data-state` attribute styling, `data-side`/`data-align` positioning, `role` and `aria-*` attributes, the `asChild` composition pattern, `forwardRef` chains, or `Portal` wrappers.

**Why it happens:** The 15 Shadcn components in this project are copy-pasted Radix wrappers. When restyling, developers may:
- Remove the `cn()` className merge logic and hardcode classes, breaking override capability
- Override `data-state="open"` / `data-state="closed"` styling without realizing it controls visual state transitions
- Break the `React.forwardRef` chain needed for Radix's internal focus management
- Remove `<Portal>` wrappers from dialogs/popovers, causing z-index and scroll containment issues
- Change the component's root element type, breaking Radix's event handling expectations

**Consequences:**
- Dropdowns open but do not close on outside click
- Dialogs lose focus trapping (accessibility violation -- keyboard users cannot interact)
- Select components lose keyboard navigation (arrow keys, type-ahead search)
- Tooltips position incorrectly or render behind other elements
- Popovers clip inside scroll containers instead of portaling to the body

**Prevention:**
1. **Rule: only modify `className` props and CSS, never remove Radix behavioral props** (`data-*`, `aria-*`, `role`, `asChild`, `forwardRef`, `Portal`)
2. When updating a Shadcn component, diff against the latest Shadcn source (`npx shadcn@latest diff`) to ensure no behavioral code was removed
3. Test every restyled component with keyboard-only navigation (Tab, Enter, Escape, Arrow keys)
4. Preserve the `cn()` utility everywhere -- it ensures Tailwind class merging works correctly with Radix's internal classes
5. Do not remove or restructure `React.forwardRef` wrappers

**Detection:** Keyboard-only testing after each component restyle. Tab through all interactive elements on the page.

**Phase:** Every phase where Shadcn components are touched. Most critical in Phase 2 (shared component library).

---

### Pitfall 5: Light Mode as Afterthought -- Accessibility Failures

**What goes wrong:** All 11 HTML mockups are dark-mode only. Developer implements dark mode pixel-perfectly, then "adapts" for light mode by mechanically lightening/inverting colors without verifying WCAG contrast ratios. Result: light mode text fails WCAG AA (4.5:1 for normal text, 3:1 for large text and UI components), hover states become invisible against white backgrounds, and borders disappear.

**Why it happens:** The current codebase already defines both light and dark CSS variables, but the redesign palette is fundamentally different from the current one (`#08090B` vs current `hsl(220 13% 9%)`). The existing light theme variables will not match the new design language. Generating a correct light palette from dark-only mockups requires deliberate color calculation, not visual guessing.

**Consequences:**
- `--muted-foreground` text becomes invisible on white/light backgrounds (currently `hsl(210 8% 62%)` in dark mode -- this gray is readable on dark backgrounds but fails contrast on white)
- Status badges (`.badge-passed`, `.badge-failed`, etc.) currently use hardcoded hex colors with separate dark mode overrides -- these must be recalculated for the new palette
- Glass morphism effects (`backdrop-filter: blur()`) that look premium on dark backgrounds appear washed out and amateur on light backgrounds
- Gradient text (`.gradient-text`, `.page-title`, `.greeting-gradient`) becomes unreadable when gradient endpoints don't contrast with light backgrounds
- The "always dark sidebar" creates a harsh contrast boundary against a bright white content area

**Prevention:**
1. **Design the light palette deliberately in Phase 1.** For each dark mode token, calculate the light equivalent ensuring WCAG AA compliance
2. Use automated contrast checking tools (WebAIM contrast checker, axe-core browser extension) for every text/background pair
3. Status badge colors need separate light and dark definitions -- the current pattern of having both is correct, just update values
4. Glass effects need higher background opacity and less blur radius in light mode
5. Test gradient text readability in both modes with a contrast ratio tool
6. Consider a subtle shadow or border on the sidebar-to-content boundary in light mode to soften the dark-to-light transition

**Detection:** Run Lighthouse accessibility audit targeting both themes after each phase. All contrast violations must be zero.

**Phase:** Phase 1 (token definition). The light palette must be defined alongside the dark palette, never deferred.

---

## Moderate Pitfalls

---

### Pitfall 6: Animation Performance -- Death by a Thousand Springs

**What goes wrong:** The redesign design system specifies spring/bounce animations for interactions. Developer wraps elements in `motion` (formerly framer-motion) `<motion.div>` components for cards, list items, page transitions, sidebar toggles, modals, and tooltips. Each animation is fine in isolation. But on pages with high element counts -- Dashboard (20+ widgets), Kanban (50+ cards), Messenger (100+ visible messages) -- simultaneous animations cause frame drops and jank on mid-range hardware.

**Why it happens:** Spring animations are JS-driven and run on the main thread. They are not compositor-friendly like CSS `transform`/`opacity` transitions. When 30 Kanban cards animate their `layout` prop simultaneously during a filter change, React re-renders cascade, and the main thread bottlenecks. The current codebase uses **CSS-only animations** (`@keyframes` in app.css -- `fade-in`, `slide-in-right`, `message-in`, `dropdown-in`, `modal-in`, etc.) which are compositor-friendly. Switching wholesale to JS-driven springs is a performance regression.

**Prevention:**
1. **Keep CSS animations for common patterns** (fade-in, slide-in, scale-in, modal enter/exit). These are already defined in `app.css` and work well
2. Use `motion` library only for: drag reorder feedback, layout shifts that need FLIP, and page route transitions. Not for hover effects, tooltips, or list item enter/exit
3. Only animate `transform` and `opacity` -- never `width`, `height`, `background-color`, `border-radius`, or `box-shadow` via JS
4. Never use `motion`'s `layout` prop on list items in large lists (Kanban cards, message items, widget grids)
5. Add `@media (prefers-reduced-motion: reduce)` support -- disable all non-essential animations
6. Performance budget: 60fps on Chrome DevTools with 4x CPU throttle on Dashboard and Kanban pages

**Detection:** Chrome DevTools Performance tab profiling. Record a 5-second interaction on Kanban (drag a card) and Dashboard (scroll through widgets). Frame rate must stay above 55fps.

**Phase:** Phase 2 (animation foundation). Establish animation primitives and performance budget before per-page work.

---

### Pitfall 7: Font Loading Causes Layout Shift (CLS)

**What goes wrong:** The redesign specifies Inter (body text) + JetBrains Mono (code). Loading two font families with multiple weights (400, 500, 600, 700 for Inter; 400 for JetBrains Mono) causes FOUT (Flash of Unstyled Text): the page renders with system fonts then visibly shifts when custom fonts load. This is especially jarring on Login (large hero text) and the sidebar (narrow column where every pixel matters).

**Why it happens:** The current font stack is `"Geist", "Inter", -apple-system...` -- already declaring Inter as a fallback. The redesign likely drops Geist and makes Inter primary. JetBrains Mono is a new addition. Each font file is 50-150KB (WOFF2), and loading 6+ font files creates visible loading delays on slower connections.

**Consequences:**
- Cumulative Layout Shift (CLS) as font metrics differ between system fallback and loaded Inter/JetBrains Mono
- Sidebar text reflows when Inter loads (system font has different character widths)
- Code blocks in issue descriptions and TMS test cases resize when JetBrains Mono loads
- Users on slow connections see a 500ms-1s font swap flash

**Prevention:**
1. **Self-host fonts as WOFF2** files in the project (do not use Google Fonts CDN) to eliminate DNS lookup + external request
2. Preload only critical weights: `<link rel="preload" as="font" crossorigin>` for Inter 400 and 600 only
3. Use `font-display: swap` for Inter (body text must be immediately visible), `font-display: optional` for JetBrains Mono (if it does not load, `monospace` fallback is acceptable)
4. Use CSS `size-adjust` on the fallback system font to match Inter's metrics and minimize CLS
5. Subset fonts to Latin + Cyrillic only (the app supports Russian + English), reducing file size by 40-60%

**Detection:** Lighthouse CLS score. Target CLS < 0.05. Test on simulated "Slow 3G" in DevTools.

**Phase:** Phase 1 (design system foundation). Font loading strategy before component work.

---

### Pitfall 8: Tailwind Token Changes Ripple Globally

**What goes wrong:** The `@theme inline` block in `app.css` maps Tailwind utilities to CSS variables. Changing `--radius` from `0.75rem` to match the new design system (e.g., `0.5rem`) immediately changes every `rounded-sm`, `rounded-md`, `rounded-lg`, `rounded-xl` across all 235 source files. Same for shadow tokens, spacing, and color mappings. Pages that have not yet been migrated look subtly wrong.

**Why it happens:** Tailwind's utility classes derive from a single source of truth. The current radius system:
```
--radius-sm: calc(var(--radius) - 4px)  // 8px
--radius-md: calc(var(--radius) - 2px)  // 10px
--radius-lg: var(--radius)              // 12px
--radius-xl: calc(var(--radius) + 4px)  // 16px
```
Changing `--radius` changes all four. This is by design, but means radius changes are all-or-nothing.

**Consequences:**
- Every component's border radius shifts simultaneously -- cards, buttons, badges, inputs, avatars, dropdowns
- Custom CSS classes with hardcoded `border-radius` (e.g., `.card-premium { border-radius: 16px }`) become inconsistent with Tailwind radius utilities
- `.board-page-root` and other layout classes with hardcoded values diverge from the token system

**Prevention:**
1. **Change `--radius` once in Phase 1** and accept the global ripple. This is intentional -- faster to fix all pages at once than maintain two radius systems
2. Audit all custom CSS classes in `app.css` that hardcode `border-radius`, `box-shadow`, or color values. Update or remove them in the same commit
3. Create a mapping: old custom class --> new Tailwind equivalent. Then systematically remove custom classes
4. Each migrated component should use ONLY Tailwind utilities -- no custom CSS classes from the old system

**Detection:** Track custom CSS class count in `app.css`. It should decrease with each phase.

**Phase:** Phase 1 (foundation) for token changes, then continuous cleanup.

---

### Pitfall 9: Mobile Responsiveness Regression

**What goes wrong:** Developer implements the desktop mockup pixel-perfectly, but the mobile layout (`mobile-layout.tsx`, `mobile-nav.tsx`, bottom navigation bar) breaks because new padding, font sizes, or layout containers do not account for viewports below 768px.

**Why it happens:** All 11 HTML mockups appear to be desktop-only designs. The current app has dedicated mobile components: mobile layout shell, bottom navigation, and simplified views. Desktop-focused redesign work can easily break these by:
- Changing shared components (Button, Input, Card) to sizes that overflow on mobile
- Modifying the sidebar in ways that break its mobile collapse/drawer behavior
- Adding horizontal layouts or fixed-width elements that do not wrap on narrow screens

**Prevention:**
1. After restyling each page for desktop, immediately verify at 375px, 768px, and 1024px viewports
2. Do not modify `mobile-layout.tsx` or `mobile-nav.tsx` structure unless mockups explicitly require it
3. Use Tailwind responsive prefixes (`sm:`, `md:`, `lg:`) consistently
4. Touch targets: minimum 44x44px per WCAG 2.5.8 (Target Size)
5. Test sidebar collapse behavior on each viewport size after any sidebar changes

**Detection:** DevTools responsive mode at 375px width after each page restyle.

**Phase:** Every phase. Add mobile check to per-page completion gate.

---

### Pitfall 10: Inconsistent UI During Component-by-Component Migration

**What goes wrong:** The app is used internally during the redesign. For weeks, some pages look "new" (redesigned) and others look "old" (pre-redesign). Shared components like Button, Badge, Avatar appear on every page -- if they are updated for the new design, they look wrong on pages that have not been migrated yet.

**Why it happens:** Component-by-component migration is correct for quality control (vs. big-bang rewrite), but without careful ordering, users see jarring transitions: new Login --> old Dashboard --> new Messenger --> old Issue Detail. The navigation shell (sidebar + topbar) frames every page, so if it changes design language before page content does, every page looks partially broken.

**Prevention:**
1. **Migrate in strict dependency order:**
   - Phase 1: Foundation (tokens, variables, fonts, `!important` cleanup)
   - Phase 2: Shell (sidebar + topbar -- they frame every page)
   - Phase 3: Shared components (Button, Input, Badge, Avatar, Card, Dialog, etc.)
   - Phase 4+: Pages in user flow order (Login --> Dashboard --> Messenger --> Kanban --> Issue Detail --> Calls --> TMS --> Settings --> Admin --> Onboarding --> Trash)
   - Last: Pages without mockups (Projects list, Sprint management, JQL search, Project settings, Project dashboard) -- apply the new design system consistently
2. When restyling shared components, ensure they degrade gracefully in not-yet-migrated pages (e.g., do not change Button size if old pages have tight layouts)
3. Deploy the full redesign as a single milestone release if possible, rather than incremental deploys with mixed UI

**Detection:** Visual walkthrough of entire app after each phase, checking both migrated and unmigrated pages.

**Phase:** This dictates the entire roadmap structure. Phase ordering decision.

---

## Minor Pitfalls

---

### Pitfall 11: Accent Color System Incompatibility

**What goes wrong:** The current accent system overrides `--primary` and `--ring` via `[data-accent="red"]` selectors (with separate `.dark[data-accent="red"]` variants). If the redesign redefines `--primary` to mean something other than "user-chosen accent color" (e.g., brand blue), the accent system breaks.

**Prevention:** Decide in Phase 1 whether `--primary` remains the accent-driven token or if a separate `--accent-user` token is introduced. Update all accent selectors in the same commit as the main variable swap. The current system has 10 accent selector blocks (5 colors x 2 themes) -- all must be updated.

**Phase:** Phase 1.

---

### Pitfall 12: Scrollbar Styling Cross-Browser Inconsistency

**What goes wrong:** Current scrollbar styling (4px, transparent track, rgba thumb) only works in WebKit browsers. Firefox users see default OS scrollbars which clash with the redesign. The `.scrollbar-thin` class targets only WebKit pseudo-elements.

**Prevention:** Add Firefox scrollbar properties alongside WebKit:
```css
* { scrollbar-width: thin; scrollbar-color: var(--border) transparent; }
```
Test in both Chrome and Firefox after implementation.

**Phase:** Phase 1 (base styles).

---

### Pitfall 13: Third-Party Component Styling Leaks

**What goes wrong:** The Plate rich-text editor, LiveKit call UI, and Tremor chart components render their own DOM with inputs, selects, and popovers. The current global CSS overrides (targeting all `input`, `select`, `[data-radix-*]` elements with `!important`) leak into these components. After removing `!important` and changing global styles, third-party components may look broken in unexpected ways.

**Prevention:**
1. Scope form element overrides to exclude third-party containers using `:not()` selectors or wrapper class scoping
2. Test Plate editor rendering after any global input/textarea style change
3. Test Tremor chart tooltips and legends after any Radix popover/tooltip style change
4. Test LiveKit participant panel after any global button/avatar style change
5. Use Tailwind's `@layer` system to ensure component-level styles take precedence over base overrides

**Phase:** Phase 1 (during `!important` cleanup) and phases containing Messenger (Plate), Calls (LiveKit), Dashboard (Tremor).

---

### Pitfall 14: `backdrop-filter: blur()` GPU Memory on Stacked Layers

**What goes wrong:** The current design already uses glass morphism in 4 places (`.glass`, `.glass-subtle`, `.page-header-blur`, `.composer-glow`). The redesign mockups may add more blur layers. When multiple `backdrop-filter` elements are visible simultaneously (e.g., sticky header + modal overlay + tooltip), GPU memory pressure causes rendering lag on older hardware and all mobile devices.

**Prevention:** Limit `backdrop-filter` to maximum 2-3 visible instances at any time. Use opaque backgrounds as fallback: `@supports not (backdrop-filter: blur(1px))`. Profile on a throttled GPU (DevTools rendering settings).

**Phase:** Phase 2 (overlay/modal components).

---

### Pitfall 15: Z-Index Wars Between New and Existing Layers

**What goes wrong:** The redesign introduces new overlay patterns (modal redesign, toast redesign, tooltip redesign). Without a defined z-index scale, toasts appear behind modals, tooltips render under the sidebar, and dropdown menus clip behind the sticky page header (`z-index: 40` currently).

**Prevention:** Define a z-index scale as design tokens in Phase 1:
```
--z-dropdown: 50
--z-sticky: 100
--z-overlay: 200
--z-modal: 300
--z-toast: 400
--z-tooltip: 500
```
Radix components manage their own z-index through Portals -- do not override them with global rules. Remove any global z-index overrides on Radix popper content.

**Phase:** Phase 1 (design token definition).

---

### Pitfall 16: Custom CSS Class Accumulation (Dead Code)

**What goes wrong:** `app.css` currently contains 40+ custom utility classes (`.card-premium`, `.hover-lift`, `.glass`, `.gradient-tracker`, `.badge-passed`, `.page-title`, `.widget-card`, `.greeting-gradient`, etc.). As components are migrated to use Tailwind utilities directly, these old classes are not removed, creating dead CSS that confuses developers and inflates bundle size.

**Prevention:** After migrating each component, grep for the old custom classes. If unused, remove them from `app.css`. Maintain a checklist of custom classes and their migration status.

**Phase:** Every phase. Track class count as a metric.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|---|---|---|
| **Design System Foundation** | `!important` removal cascade breaks ALL form elements immediately | Accept temporary visual regression. Fix iteratively within Phase 1. Do not ship to users until stabilized |
| **Design System Foundation** | CSS variable swap makes accent colors look wrong | Update all 10 accent selector blocks in same commit as variable swap |
| **Design System Foundation** | Font loading CLS on login page hero text | Self-host WOFF2, preload Inter 400+600, add `size-adjust` to fallback font |
| **Design System Foundation** | Light mode palette fails WCAG contrast on first attempt | Run axe-core audit immediately after defining light tokens. Iterate until zero violations |
| **Shell (Sidebar + Topbar)** | "Always dark" sidebar styling breaks when light mode variables change | Sidebar uses `--sidebar-*` namespace -- keep it isolated from `--background`/`--foreground` changes |
| **Shell (Sidebar + Topbar)** | Sidebar collapse animation breaks from layout restructuring | Test collapse/expand toggle at every viewport size after changes |
| **Shared Components** | Radix primitive breakage from Shadcn restyling | Diff against shadcn source, never remove data-* or aria-* attributes |
| **Shared Components** | Animation performance budget blown on shared hover/press effects | CSS transitions for hover/press. Reserve motion.js for layout animations only |
| **Kanban Page** | pragmatic-drag-and-drop breaks from DOM restructuring | Test drag-and-drop after EVERY JSX change on this page. Do not wrap drop zones in new containers |
| **Messenger Page** | Infinite scroll breaks, Plate editor styling leaks from global CSS changes | Scope input/textarea overrides to exclude `.plate-editor` container |
| **Calls Page** | LiveKit UI components get unwanted global styling | Scope call UI styles, test after any global button/avatar change |
| **Dashboard Page** | 20+ widget mount animations cause simultaneous jank | CSS `animation` only for widgets. No motion.js `layout` on widget grid. Stagger with `animation-delay` |
| **TMS Page** | Badge and table stripe colors break with new palette | Verify all `.badge-*` colors and `.table-striped` against new `--muted`/`--card` values in both themes |
| **Light Mode (all pages)** | Gradient text becomes unreadable on light backgrounds | Test `.gradient-text`, `.page-title`, `.greeting-gradient` in light mode with contrast checker |
| **Pages Without Mockups** | No design reference leads to inconsistent interpretation | Apply the design system tokens consistently. Match existing migrated pages for spacing, radius, shadow |
| **Mobile (all pages)** | Desktop redesign overflows or breaks mobile views | Test 375px viewport after each page. Do not restructure mobile-specific layout components |

## "Will Look Done But Actually Broken" Checklist

Before marking any page redesign as complete, verify:

- [ ] **Keyboard navigation:** Tab through every interactive element. Enter/Escape/Arrow keys work on dropdowns, dialogs, selects
- [ ] **Both themes:** Switch between light and dark mode. No invisible text, no disappearing borders, no unreadable badges
- [ ] **All accent colors:** Cycle through red/blue/green/purple/orange. Primary buttons, focus rings, and active states must look correct in each
- [ ] **Mobile viewport:** Check 375px width. No horizontal overflow, no truncated text without ellipsis, touch targets >= 44px
- [ ] **Scroll behavior:** Infinite scroll still works (messenger), scroll position preserves on navigation, sticky headers do not overlap content
- [ ] **All interactive states:** hover, focus, active, disabled, loading, error states for every interactive element
- [ ] **Third-party components:** Plate editor renders correctly, Tremor charts display correctly, LiveKit call UI functions
- [ ] **Animations:** No jank on page load, no flash on theme switch, reduced-motion preference respected
- [ ] **Real data:** Test with actual content lengths (long usernames, long task titles, multi-line comments), not just "Test" placeholder text

## Sources

- Codebase analysis: `frontend/src/app.css` (1053 lines, 56 `!important` declarations), 15 Shadcn/ui components, 235 TSX/TS files, 11 HTML mockups in `redesign/`
- [Tailwind CSS v4 Official Migration Guide](https://tailwindcss.com/blog/tailwindcss-v4) -- `@theme inline` configuration, utility class renames, opacity syntax changes
- [Motion Animation Performance Guide](https://motion.dev/docs/performance) -- compositor-safe properties, hardware acceleration, layout animation costs
- [Web Animation Performance Tier List](https://motion.dev/magazine/web-animation-performance-tier-list) -- transform/opacity as tier-1 properties
- [Font Loading Strategies (2025)](https://font-converters.com/guides/font-loading-strategies) -- FOUT/FOIT prevention, preload, font-display, size-adjust
- [WCAG Accessible Dark/Light Mode Design](https://blog.logrocket.com/how-do-you-implement-accessible-linear-design-across-light-and-dark-modes/) -- dual-theme contrast requirements, 4.5:1 AA ratio
- [Shadcn/ui Official Documentation](https://ui.shadcn.com/) -- component architecture, Radix primitive relationship, theming conventions
- [Radix UI Primitives and Shadcn Relationship](https://workos.com/blog/what-is-the-difference-between-radix-and-shadcn-ui) -- when re-theming breaks primitives

---
*Pitfalls research for: Hub UI Redesign Milestone 2 (React + Tailwind CSS 4 + Shadcn/ui)*
*Researched: 2026-03-27*
