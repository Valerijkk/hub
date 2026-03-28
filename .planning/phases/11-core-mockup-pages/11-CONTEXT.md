# Phase 11: Core Mockup Pages - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous mode)

<domain>
## Phase Boundary

Четыре самые важные и сложные страницы (Login, Dashboard, Messenger, Kanban) pixel-perfect соответствуют HTML-макетам, а вся существующая функциональность работает без регрессий.

Requirements: LOGN-01..05, DASH-01..05, MSGR-01..06, KANB-01..04

</domain>

<decisions>
## Implementation Decisions

### Login Page (LOGN-01..05)
- Split layout: 55% left (animated gradient/mesh bg), 45% right (form)
- Floating animated feature cards on left panel (staggered 7-9s infinite float)
- Mesh gradient background: animated hue-rotate 20s
- Email→Code flow: redesigned inputs + 48x58px code fields (JetBrains Mono, center aligned)
- Telegram button preserved
- Magic link flow must work without regression

### Dashboard Page (DASH-01..05)
- 3-column widget grid (repeat(3, 1fr), gap 16px)
- Greeting section with gradient text and current date
- Quick action buttons with new button styling
- Staggered entrance animations on widgets (enterUp keyframe, 40-160ms delays)
- Widget drag-and-drop (react-grid-layout) must work without regression

### Messenger Page (MSGR-01..06)
- 3-panel layout: main sidebar (260px) + channel list (280px) + chat (flex)
- Channel list: search input + filter buttons (All/Unread/DMs/Channels)
- Message items: 38px avatar, 14px text, 10.5px metadata, hover bg var(--hover)
- Composer: glassmorphism effect (backdrop blur)
- Reactions: 13px font, border, rounded, hover scale
- Typing indicator dots animation
- ALL messenger functionality preserved: send, react, thread, pin, forward, voice, search, polls, stickers, GIFs

### Kanban Page (KANB-01..04)
- Multi-column board with redesigned headers
- Issue cards: priority dots with glow, assignee avatars, hover translateY(-2px) + shadow
- Filter/view options bar
- pragmatic-drag-and-drop MUST work without regression — only restyle with CSS, minimize JSX changes

### Claude's Discretion
- Exact animation timing values where not specified in mockups
- Responsive breakpoints for widget grid (2 cols at 1200px and below)
- Plate editor styling within messenger (scoped to not break rich text)
- Any additional CSS utilities needed for page-specific patterns

</decisions>

<code_context>
## Existing Code Insights

### Key Files — Login
- frontend/src/routes/login.tsx (or similar login route)

### Key Files — Dashboard
- frontend/src/routes/_authenticated/dashboard.tsx (lazy-loaded)
- frontend/src/modules/dashboard/ (widget components)
- frontend/src/stores/dashboard-store.ts

### Key Files — Messenger
- frontend/src/routes/_authenticated/messenger/ (layout + channel routes)
- frontend/src/modules/messenger/ (chat components, composer, messages)
- frontend/src/components/messenger/ (channel list, etc.)
- frontend/src/stores/messenger-store.ts

### Key Files — Kanban
- frontend/src/routes/_authenticated/projects/$projectId/board/
- frontend/src/modules/tracker/ (kanban board, issue cards)
- frontend/src/stores/board-store.ts

### Mockup Sources (read for exact CSS values)
- редизайн/01-login (1).html
- редизайн/02-dashboard (2).html
- редизайн/03-messenger (2).html
- редизайн/04-kanban.html

</code_context>

<specifics>
## Specific Ideas

- Макеты в папке `редизайн/` — единственный источник правды для pixel-perfect match
- Плоть от плоти рестайл — только CSS classes, НЕ трогаем state management, event handlers, API calls
- Messenger: НЕ трогать Plate editor internals, только обернуть стилями
- Kanban: НЕ трогать pragmatic-drag-and-drop конфигурацию, только визуал карт и колонок

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
