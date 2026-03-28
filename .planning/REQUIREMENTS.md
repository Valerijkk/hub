# Requirements: Hub v2.0 Redesign

**Defined:** 2026-03-27
**Core Value:** Pixel-perfect редизайн всего фронтенда Hub по 11 HTML-макетам без ломания функционала

## v2.0 Requirements

### Design System Foundation (DSFN)

- [x] **DSFN-01**: User sees Inter font for all UI text and JetBrains Mono for code/monospace elements
- [x] **DSFN-02**: User sees new 5-level surface palette (void/base/raised/card/input) in dark mode
- [x] **DSFN-03**: User sees 4-level text hierarchy (t1/t2/t3/t4) with correct contrast ratios
- [x] **DSFN-04**: User sees 3-level border system (b1/b2/b3) with correct opacity values
- [x] **DSFN-05**: User sees new radius system (6/8/12/16px + pill) applied consistently
- [x] **DSFN-06**: User sees 4-level shadow system (sh1-sh4) on cards, modals, dropdowns
- [x] **DSFN-07**: User sees spring/bounce easing on hover/focus/enter animations
- [x] **DSFN-08**: All !important declarations removed from app.css without visual regression
- [x] **DSFN-09**: User can select accent color (red/blue/green/purple/orange) and see it applied consistently

### Theme Support (THME)

- [x] **THME-01**: User sees dark theme matching mockup palette exactly
- [x] **THME-02**: User sees light theme derived from dark tokens with consistent design language
- [x] **THME-03**: User can switch between dark/light/system themes without visual glitches
- [x] **THME-04**: Light theme passes WCAG AA contrast ratios (4.5:1 text, 3:1 UI)

### Layout Shell (LYOT)

- [x] **LYOT-01**: User sees redesigned sidebar with gradient edge line and new background
- [x] **LYOT-02**: User sees active navigation item with accent glow indicator
- [x] **LYOT-03**: User sees redesigned topbar with backdrop blur and gradient bottom edge
- [x] **LYOT-04**: Sidebar collapse/expand works without visual regression
- [x] **LYOT-05**: Mobile navigation works without regression (bottom nav, overlay sidebar)

### UI Primitives (PRIM)

- [x] **PRIM-01**: User sees redesigned Button with gradient primary, ghost/secondary variants, glow hover
- [x] **PRIM-02**: User sees redesigned Card with gradient bg, top-edge highlight, hover glow
- [x] **PRIM-03**: User sees redesigned Input with blue focus glow ring and subtle scale on focus
- [x] **PRIM-04**: User sees redesigned Badge with status-specific glow backgrounds
- [x] **PRIM-05**: User sees redesigned Dialog/Popover with entrance animation and backdrop blur
- [x] **PRIM-06**: User sees redesigned Tabs with active indicator matching section color
- [x] **PRIM-07**: User sees redesigned DropdownMenu matching new design language
- [x] **PRIM-08**: User sees redesigned ScrollArea with slim 5px scrollbar
- [x] **PRIM-09**: User sees redesigned Avatar with online indicator glow
- [x] **PRIM-10**: User sees redesigned Tooltip matching new palette and radius

### Login Page (LOGN)

- [x] **LOGN-01**: User sees split layout — animated gradient left, form right
- [x] **LOGN-02**: User sees floating animated feature cards on left panel
- [x] **LOGN-03**: User sees email→code flow with redesigned inputs and code fields
- [x] **LOGN-04**: User sees mesh gradient background animation
- [x] **LOGN-05**: Login flow (magic link + Telegram) works without regression

### Dashboard Page (DASH)

- [x] **DASH-01**: User sees 3-column widget grid with redesigned cards
- [x] **DASH-02**: User sees greeting section with date and gradient text
- [x] **DASH-03**: User sees quick action buttons with new styling
- [x] **DASH-04**: User sees staggered entrance animations on widgets
- [x] **DASH-05**: Dashboard widget drag-and-drop works without regression

### Messenger Page (MSGR)

- [x] **MSGR-01**: User sees redesigned 3-panel layout (sidebar + channel list 280px + chat)
- [x] **MSGR-02**: User sees redesigned channel list with search and filter buttons
- [x] **MSGR-03**: User sees redesigned message items with new avatar, text, metadata styling
- [x] **MSGR-04**: User sees redesigned message composer with glassmorphism effect
- [x] **MSGR-05**: User sees redesigned reactions, typing indicator, reply bar
- [x] **MSGR-06**: All messenger functionality works without regression (send, react, thread, pin, forward, voice, search, polls, stickers, GIFs)

### Kanban Page (KANB)

- [x] **KANB-01**: User sees redesigned kanban columns with new header styling
- [x] **KANB-02**: User sees redesigned issue cards with priority dots, assignee avatars, glow hover
- [x] **KANB-03**: User sees redesigned filter and view options bar
- [x] **KANB-04**: Kanban drag-and-drop works without regression

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

| Requirement | Phase | Status |
|-------------|-------|--------|
| DSFN-01 | Phase 8 | Complete |
| DSFN-02 | Phase 8 | Complete |
| DSFN-03 | Phase 8 | Complete |
| DSFN-04 | Phase 8 | Complete |
| DSFN-05 | Phase 8 | Complete |
| DSFN-06 | Phase 8 | Complete |
| DSFN-07 | Phase 8 | Complete |
| DSFN-08 | Phase 8 | Complete |
| DSFN-09 | Phase 8 | Complete |
| THME-01 | Phase 8 | Complete |
| THME-02 | Phase 8 | Complete |
| THME-03 | Phase 8 | Complete |
| THME-04 | Phase 8 | Complete |
| LYOT-01 | Phase 9 | Complete |
| LYOT-02 | Phase 9 | Complete |
| LYOT-03 | Phase 9 | Complete |
| LYOT-04 | Phase 9 | Complete |
| LYOT-05 | Phase 9 | Complete |
| PRIM-01 | Phase 10 | Complete |
| PRIM-02 | Phase 10 | Complete |
| PRIM-03 | Phase 10 | Complete |
| PRIM-04 | Phase 10 | Complete |
| PRIM-05 | Phase 10 | Complete |
| PRIM-06 | Phase 10 | Complete |
| PRIM-07 | Phase 10 | Complete |
| PRIM-08 | Phase 10 | Complete |
| PRIM-09 | Phase 10 | Complete |
| PRIM-10 | Phase 10 | Complete |
| LOGN-01 | Phase 11 | Complete |
| LOGN-02 | Phase 11 | Complete |
| LOGN-03 | Phase 11 | Complete |
| LOGN-04 | Phase 11 | Complete |
| LOGN-05 | Phase 11 | Complete |
| DASH-01 | Phase 11 | Complete |
| DASH-02 | Phase 11 | Complete |
| DASH-03 | Phase 11 | Complete |
| DASH-04 | Phase 11 | Complete |
| DASH-05 | Phase 11 | Complete |
| MSGR-01 | Phase 11 | Complete |
| MSGR-02 | Phase 11 | Complete |
| MSGR-03 | Phase 11 | Complete |
| MSGR-04 | Phase 11 | Complete |
| MSGR-05 | Phase 11 | Complete |
| MSGR-06 | Phase 11 | Complete |
| KANB-01 | Phase 11 | Complete |
| KANB-02 | Phase 11 | Complete |
| KANB-03 | Phase 11 | Complete |
| KANB-04 | Phase 11 | Complete |
| ISSU-01 | Phase 12 | Pending |
| ISSU-02 | Phase 12 | Pending |
| ISSU-03 | Phase 12 | Pending |
| ISSU-04 | Phase 12 | Pending |
| CALL-01 | Phase 12 | Pending |
| CALL-02 | Phase 12 | Pending |
| CALL-03 | Phase 12 | Pending |
| CALL-04 | Phase 12 | Pending |
| TMSP-01 | Phase 12 | Pending |
| TMSP-02 | Phase 12 | Pending |
| TMSP-03 | Phase 12 | Pending |
| TMSP-04 | Phase 12 | Pending |
| STNG-01 | Phase 12 | Pending |
| STNG-02 | Phase 12 | Pending |
| STNG-03 | Phase 12 | Pending |
| STNG-04 | Phase 12 | Pending |
| STNG-05 | Phase 12 | Pending |
| ADMN-01 | Phase 12 | Pending |
| ADMN-02 | Phase 12 | Pending |
| ADMN-03 | Phase 12 | Pending |
| ADMN-04 | Phase 12 | Pending |
| ADMN-05 | Phase 12 | Pending |
| ONBD-01 | Phase 12 | Pending |
| ONBD-02 | Phase 12 | Pending |
| ONBD-03 | Phase 12 | Pending |
| ONBD-04 | Phase 12 | Pending |
| ONBD-05 | Phase 12 | Pending |
| TRSH-01 | Phase 12 | Pending |
| TRSH-02 | Phase 12 | Pending |
| TRSH-03 | Phase 12 | Pending |
| XTRA-01 | Phase 12 | Pending |
| XTRA-02 | Phase 12 | Pending |
| XTRA-03 | Phase 12 | Pending |
| XTRA-04 | Phase 12 | Pending |
| XTRA-05 | Phase 12 | Pending |
| XTRA-06 | Phase 12 | Pending |
| XTRA-07 | Phase 12 | Pending |

**Coverage:**
- v2.0 requirements: 85 total
- Mapped to phases: 85 (Phase 8: 13, Phase 9: 5, Phase 10: 10, Phase 11: 20, Phase 12: 37)
- Phase 13 (Polish & QA): cross-cutting quality validation, no owned requirements
- Unmapped: 0

---
*Requirements defined: 2026-03-27*
*Last updated: 2026-03-27 after roadmap creation*
