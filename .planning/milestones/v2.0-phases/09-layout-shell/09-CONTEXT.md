# Phase 9: Layout Shell - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous mode — design decisions from mockups)

<domain>
## Phase Boundary

Пользователь видит обновлённый сайдбар с градиентной линией и свечением активного пункта, топбар с backdrop blur, и вся оболочка работает на всех размерах экрана.

Requirements: LYOT-01, LYOT-02, LYOT-03, LYOT-04, LYOT-05

</domain>

<decisions>
## Implementation Decisions

### Sidebar Redesign
- Background: var(--sidebar) (#0B0C0F) — always dark regardless of theme
- Gradient edge line: ::after pseudo-element on right border with linear-gradient
- Active nav item: accent glow indicator using user's chosen accent color + box-shadow
- Width: 260px (matches mockups), collapsible to 64px (icons only)
- Navigation items: 14px font, 500 weight, var(--t2) color, var(--t1) when active
- User section at bottom: 56px height
- Header: 56px height with logo + brand name

### Topbar Redesign
- Height: 56px (consistent with sidebar header)
- Background: backdrop-filter: blur(24px) saturate(180%) with rgba background
- Bottom border: gradient underline (not solid border)
- z-index: 50 (sticky)
- Padding: 0 24px

### Mobile Navigation
- Bottom navigation bar preserved
- Overlay sidebar for mobile (existing behavior)
- No regression on mobile nav functionality

### Claude's Discretion
- Exact gradient values for sidebar edge line
- Exact blur values for topbar if different from mockups
- Any responsive breakpoint adjustments needed

</decisions>

<code_context>
## Existing Code Insights

### Key Files
- `frontend/src/components/layout/app-shell.tsx` — Main layout wrapper
- `frontend/src/components/layout/sidebar.tsx` — Navigation sidebar
- `frontend/src/components/layout/topbar.tsx` — Top navigation bar
- `frontend/src/components/layout/mobile-nav.tsx` — Bottom navigation for mobile
- `frontend/src/app.css` — Design tokens (already updated in Phase 8)

### Established Patterns
- Sidebar uses Tailwind utility classes with CSS variables
- next-themes class attribute on html for theme switching
- AccentProvider data-accent attribute for accent colors
- Collapsible sidebar via Zustand store (sidebar-store.ts)

### Integration Points
- AppShell wraps all authenticated routes
- Sidebar renders NavigationMenu with route links
- Topbar shows on mobile (hidden on desktop if sidebar visible)

</code_context>

<specifics>
## Specific Ideas

- Макеты из папки `редизайн/` — sidebar всегда тёмный с gradient edge на правой стороне
- Active state: amber/blue/green/purple/red accent bar на левой стороне элемента (зависит от user accent)
- Topbar: backdrop blur + gradient bottom edge, sticky
- Sidebar footer: user avatar + name + status

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>
