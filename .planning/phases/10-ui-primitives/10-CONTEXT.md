# Phase 10: UI Primitives - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous mode)

<domain>
## Phase Boundary

Все 10 базовых Shadcn-компонентов рестайлированы под новый дизайн-язык, и любая страница, использующая эти компоненты, автоматически выглядит обновлённой.

Requirements: PRIM-01..PRIM-10

</domain>

<decisions>
## Implementation Decisions

### Button (PRIM-01)
- Primary: linear-gradient(135deg, var(--blueS), #2563EB), white text, glow shadow on hover
- Secondary/Ghost: bg transparent or var(--input), border var(--b1), hover var(--hover)
- Hover: translateY(-1px), increased shadow
- Active: scale(.96)
- Border-radius: var(--r2) (8px)
- Font: weight 700, 14px

### Card (PRIM-02)
- Background: linear-gradient(165deg, var(--card), var(--raised))
- Border: 1px solid var(--b1)
- Border-radius: var(--r4) (16px)
- ::before top-edge 1px highlight gradient
- Hover: border-color var(--b2), box-shadow var(--sh2), translateY(-2px)

### Input (PRIM-03)
- Height: 42px
- Background: var(--input)
- Border: 1px solid var(--b1)
- Border-radius: var(--r2) (8px)
- Focus: border-color rgba(91,156,245,.3), box-shadow 0 0 0 3px rgba(91,156,245,.06), transform scale(1.005)
- Placeholder: var(--t3)

### Badge (PRIM-04)
- Padding: 2px 7px
- Border-radius: var(--pill) (100px)
- Font: 10-11px, weight 700
- Status-specific glow backgrounds

### Dialog/Popover (PRIM-05)
- Overlay: rgba(0,0,0,.65) with backdrop blur(8px)
- Dialog: bg var(--base), border var(--b2), padding 28px
- Border-radius: var(--r4) (16px)
- Box-shadow: var(--sh4)
- Entrance animation: scale(.95) opacity(0) → scale(1) opacity(1), 250ms spring

### Tabs (PRIM-06)
- Padding: 7px 14px
- Font: 13px, weight 500-600
- Color: var(--t3) default, var(--t1) active
- Active: bottom 2px bar, box-shadow glow

### DropdownMenu (PRIM-07)
- Background: var(--raised)
- Border: var(--b2)
- Item hover: var(--hover) background

### ScrollArea (PRIM-08)
- Scrollbar width: 5px
- Scrollbar thumb: var(--b2), hover var(--b3)

### Avatar (PRIM-09)
- Online indicator with glow: box-shadow 0 0 8px var(--gG)

### Tooltip (PRIM-10)
- Background: var(--raised)
- Border: var(--b1)
- Font: 12px
- Border-radius: var(--r1) (6px)

### Claude's Discretion
- Exact gradient stops for button primary variant
- Light mode adaptations for glow effects
- Any additional component variants needed

</decisions>

<code_context>
## Existing Code Insights

### Key Files
- frontend/src/components/ui/button.tsx
- frontend/src/components/ui/card.tsx
- frontend/src/components/ui/input.tsx
- frontend/src/components/ui/badge.tsx
- frontend/src/components/ui/dialog.tsx
- frontend/src/components/ui/tabs.tsx
- frontend/src/components/ui/dropdown-menu.tsx
- frontend/src/components/ui/scroll-area.tsx
- frontend/src/components/ui/avatar.tsx
- frontend/src/components/ui/tooltip.tsx

### Established Patterns
- Shadcn/ui with Radix primitives
- CVA (class-variance-authority) for button/badge variants
- cn() utility for class merging
- CSS variables consumed via Tailwind utility classes

</code_context>

<specifics>
## Specific Ideas

- Каждый компонент рестайлится поштучно — только className изменения, НЕ трогать Radix props/behavior
- Все значения из макетов в папке `редизайн/`

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
