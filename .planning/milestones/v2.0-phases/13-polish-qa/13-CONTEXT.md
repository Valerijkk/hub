# Phase 13: Polish & QA - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous mode)

<domain>
## Phase Boundary

Финальная визуальная проверка качества, доступность, мобильная адаптация и очистка кодовой базы — приложение визуально безупречно в обеих темах на всех размерах экрана.

Cross-cutting quality validation phase — no owned requirements, validates all Phase 8-12 work.

</domain>

<decisions>
## Implementation Decisions

### Animation Performance Audit
- All animations must run at 60fps on 4x CPU throttle
- Dashboard and Kanban are benchmark pages
- Replace any jank-causing animations with CSS-only alternatives
- motion v12 reserved for drag interactions only

### Accessibility Audit
- Zero contrast violations in both dark and light themes
- WCAG AA compliance: 4.5:1 for text, 3:1 for UI components
- Verify --t3 on light backgrounds (borderline at ~3:1)
- Focus states visible in both themes

### Mobile QA
- All pages correct at 375px, 768px, 1024px viewports
- No layout breaks or overflow
- Bottom nav works correctly
- Sidebar overlay works on mobile

### Dead CSS Cleanup
- Remove unused custom utilities from app.css
- Remove any orphaned classes from v1.0 that are no longer referenced
- Keep file size reasonable

### Theme Toggle Polish
- Smooth 200ms transition on theme switch
- Background and color transitions on key surfaces
- No flash of wrong theme on page load

### Outstanding Gaps from Prior Phases
- PRIM-09: Avatar online indicator glow (from Phase 10)
- Any TypeScript errors accumulated across phases

### Claude's Discretion
- Exact animation timing adjustments
- Which dead CSS classes to remove
- Any additional responsive fixes needed

</decisions>

<code_context>
## Existing Code Insights

### Key File
- frontend/src/app.css — all CSS utilities, design tokens, animations

### Verification Files
- .planning/phases/08-design-system-foundation/08-VERIFICATION.md
- .planning/phases/09-layout-shell/09-VERIFICATION.md
- .planning/phases/10-ui-primitives/10-VERIFICATION.md
- .planning/phases/11-core-mockup-pages/11-VERIFICATION.md
- .planning/phases/12-remaining-mockup-pages-unmocked-pages/12-VERIFICATION.md

</code_context>

<specifics>
## Specific Ideas

- Vite build must pass with zero errors
- TypeScript compilation must pass
- No console errors on any page

</specifics>

<deferred>
## Deferred Ideas

None — this is the final polish phase.

</deferred>
