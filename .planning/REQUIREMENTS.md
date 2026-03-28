# Requirements: Hub v2.0 Redesign

**Defined:** 2026-03-27
**Core Value:** Pixel-perfect редизайн всего фронтенда Hub по 11 HTML-макетам без ломания функционала

## v2.0 Requirements

### Design System Foundation (DSFN)

- [ ] **DSFN-01**: User sees Inter font for all UI text and JetBrains Mono for code/monospace elements
- [ ] **DSFN-02**: User sees new 5-level surface palette (void/base/raised/card/input) in dark mode
- [ ] **DSFN-03**: User sees 4-level text hierarchy (t1/t2/t3/t4) with correct contrast ratios
- [ ] **DSFN-04**: User sees 3-level border system (b1/b2/b3) with correct opacity values
- [ ] **DSFN-05**: User sees new radius system (6/8/12/16px + pill) applied consistently
- [ ] **DSFN-06**: User sees 4-level shadow system (sh1-sh4) on cards, modals, dropdowns
- [ ] **DSFN-07**: User sees spring/bounce easing on hover/focus/enter animations
- [ ] **DSFN-08**: All !important declarations removed from app.css without visual regression
- [ ] **DSFN-09**: User can select accent color (red/blue/green/purple/orange) and see it applied consistently

### Theme Support (THME)

- [ ] **THME-01**: User sees dark theme matching mockup palette exactly
- [ ] **THME-02**: User sees light theme derived from dark tokens with consistent design language
- [ ] **THME-03**: User can switch between dark/light/system themes without visual glitches
- [ ] **THME-04**: Light theme passes WCAG AA contrast ratios (4.5:1 text, 3:1 UI)

### Layout Shell (LYOT)

- [ ] **LYOT-01**: User sees redesigned sidebar with gradient edge line and new background
- [ ] **LYOT-02**: User sees active navigation item with accent glow indicator
- [ ] **LYOT-03**: User sees redesigned topbar with backdrop blur and gradient bottom edge
- [ ] **LYOT-04**: Sidebar collapse/expand works without visual regression
- [ ] **LYOT-05**: Mobile navigation works without regression (bottom nav, overlay sidebar)

### UI Primitives (PRIM)

- [ ] **PRIM-01**: User sees redesigned Button with gradient primary, ghost/secondary variants, glow hover
- [ ] **PRIM-02**: User sees redesigned Card with gradient bg, top-edge highlight, hover glow
- [ ] **PRIM-03**: User sees redesigned Input with blue focus glow ring and subtle scale on focus
- [ ] **PRIM-04**: User sees redesigned Badge with status-specific glow backgrounds
- [ ] **PRIM-05**: User sees redesigned Dialog/Popover with entrance animation and backdrop blur
- [ ] **PRIM-06**: User sees redesigned Tabs with active indicator matching section color
- [ ] **PRIM-07**: User sees redesigned DropdownMenu matching new design language
- [ ] **PRIM-08**: User sees redesigned ScrollArea with slim 5px scrollbar
- [ ] **PRIM-09**: User sees redesigned Avatar with online indicator glow
- [ ] **PRIM-10**: User sees redesigned Tooltip matching new palette and radius

### Login Page (LOGN)

- [ ] **LOGN-01**: User sees split layout — animated gradient left, form right
- [ ] **LOGN-02**: User sees floating animated feature cards on left panel
- [ ] **LOGN-03**: User sees email→code flow with redesigned inputs and code fields
- [ ] **LOGN-04**: User sees mesh gradient background animation
- [ ] **LOGN-05**: Login flow (magic link + Telegram) works without regression

### Dashboard Page (DASH)

- [ ] **DASH-01**: User sees 3-column widget grid with redesigned cards
- [ ] **DASH-02**: User sees greeting section with date and gradient text
- [ ] **DASH-03**: User sees quick action buttons with new styling
- [ ] **DASH-04**: User sees staggered entrance animations on widgets
- [ ] **DASH-05**: Dashboard widget drag-and-drop works without regression

### Messenger Page (MSGR)

- [ ] **MSGR-01**: User sees redesigned 3-panel layout (sidebar + channel list 280px + chat)
- [ ] **MSGR-02**: User sees redesigned channel list with search and filter buttons
- [ ] **MSGR-03**: User sees redesigned message items with new avatar, text, metadata styling
- [ ] **MSGR-04**: User sees redesigned message composer with glassmorphism effect
- [ ] **MSGR-05**: User sees redesigned reactions, typing indicator, reply bar
- [ ] **MSGR-06**: All messenger functionality works without regression (send, react, thread, pin, forward, voice, search, polls, stickers, GIFs)

### Kanban Page (KANB)

- [ ] **KANB-01**: User sees redesigned kanban columns with new header styling
- [ ] **KANB-02**: User sees redesigned issue cards with priority dots, assignee avatars, glow hover
- [ ] **KANB-03**: User sees redesigned filter and view options bar
- [ ] **KANB-04**: Kanban drag-and-drop works without regression

### Issue Detail Page (ISSU)

- [ ] **ISSU-01**: User sees redesigned two-column layout (content + sidebar meta)
- [ ] **ISSU-02**: User sees redesigned issue description, comments, activity section
- [ ] **ISSU-03**: User sees redesigned right sidebar (status, assignee, deadline, related issues)
- [ ] **ISSU-04**: All issue functionality works without regression (edit, comment, time track, checklist, attachments)

### Calls Page (CALL)

- [ ] **CALL-01**: User sees redesigned video/call participant grid
- [ ] **CALL-02**: User sees redesigned call controls (mute, camera, share, end call)
- [ ] **CALL-03**: User sees redesigned participant status indicators
- [ ] **CALL-04**: All call functionality works without regression (join, mute, screen share, record, hand raise, chat)

### TMS Page (TMSP)

- [ ] **TMSP-01**: User sees redesigned test management interface
- [ ] **TMSP-02**: User sees redesigned test case grid/list with new card styling
- [ ] **TMSP-03**: User sees redesigned execution status tracking
- [ ] **TMSP-04**: All TMS functionality works without regression (cases, runs, environments, traceability)

### Settings Page (STNG)

- [ ] **STNG-01**: User sees redesigned sidebar navigation + content layout
- [ ] **STNG-02**: User sees redesigned profile section with new form styling
- [ ] **STNG-03**: User sees redesigned appearance section with theme/accent selectors
- [ ] **STNG-04**: User sees redesigned notification preferences
- [ ] **STNG-05**: All settings functionality works without regression

### Admin Page (ADMN)

- [ ] **ADMN-01**: User sees redesigned KPI cards (Users, Projects, Tasks, DAU/WAU/MAU)
- [ ] **ADMN-02**: User sees redesigned tabbed interface (Users, Audit, Resource Limits)
- [ ] **ADMN-03**: User sees redesigned user management table
- [ ] **ADMN-04**: User sees redesigned audit log with filters
- [ ] **ADMN-05**: All admin functionality works without regression

### Onboarding Page (ONBD)

- [ ] **ONBD-01**: User sees redesigned centered modal flow with progress indicator
- [ ] **ONBD-02**: User sees redesigned profile form (Step 1)
- [ ] **ONBD-03**: User sees redesigned feature tour with 6 cards (Step 2)
- [ ] **ONBD-04**: User sees redesigned theme + accent selection (Step 3)
- [ ] **ONBD-05**: Onboarding flow works without regression

### Trash Page (TRSH)

- [ ] **TRSH-01**: User sees redesigned trash table with new styling
- [ ] **TRSH-02**: User sees redesigned restore/delete buttons and type badges
- [ ] **TRSH-03**: Trash functionality works without regression (restore, permanent delete)

### Pages Without Mockups (XTRA)

- [ ] **XTRA-01**: User sees Projects list page styled consistently with design system
- [ ] **XTRA-02**: User sees Project overview page styled consistently with design system
- [ ] **XTRA-03**: User sees Sprint management pages (backlog, sprint detail, burndown, velocity, planning poker) styled consistently
- [ ] **XTRA-04**: User sees JQL search page styled consistently with design system
- [ ] **XTRA-05**: User sees Project settings pages (workflow, labels, integrations) styled consistently
- [ ] **XTRA-06**: User sees Project dashboard (widget grid) styled consistently with design system
- [ ] **XTRA-07**: All existing functionality on unmocked pages works without regression

## Future Requirements

### Polish (deferred to later iteration)

- **PLSH-01**: Per-page ambient body gradients (Dashboard: amber, Messenger: blue)
- **PLSH-02**: Skeleton loading states for all widgets
- **PLSH-03**: Smooth theme toggle transition (remove disableTransitionOnChange)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Route-level page transition animations | Layout shift bugs, productivity tool not marketing site |
| Parallax/scroll-triggered animations | Not appropriate for SaaS productivity tool |
| 3D transforms / perspective tilt on cards | Over-designed, distracting |
| Gradient buttons beyond primary CTAs | Visual noise, max 1-2 per page |
| New functionality or features | This milestone is visual-only |
| Backend changes | Frontend redesign only |
| New npm packages for styling | Existing stack covers all needs |

## Traceability

Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| — | — | — |

**Coverage:**
- v2.0 requirements: 68 total
- Mapped to phases: 0 (pending roadmap)
- Unmapped: 68

---
*Requirements defined: 2026-03-27*
*Last updated: 2026-03-27 after initial definition*
