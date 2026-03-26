# Hub — UI/UX Design Requirements
## Comprehensive Design Specification for All 11 Modules

> **Purpose**: This document defines the complete UI/UX requirements for Hub — a unified corporate platform combining project management, test management, messaging, video calls, and CI/CD integrations. Use it to generate high-fidelity dark-theme mockups and reference frontend code for every screen.

> **Design Philosophy**: Dark, minimalist, maximally modern. Must look expensive and premium. Inspired by Linear (tracker), Telegram (messenger), Allure TestOps (TMS), Yandex Telemost (calls), and Vercel (overall polish). Smooth micro-animations everywhere.

---

## 1. GLOBAL DESIGN SYSTEM

### 1.1 Color Palette

#### Dark Theme (Primary)
| Token | Value | Usage |
|-------|-------|-------|
| `--bg-primary` | `hsl(220, 13%, 9%)` | Main background (#141518) |
| `--bg-secondary` | `hsl(220, 11%, 12%)` | Cards, panels (#1C1E22) |
| `--bg-tertiary` | `hsl(220, 10%, 16%)` | Hover states, inputs (#252830) |
| `--bg-elevated` | `hsl(220, 11%, 14%)` | Floating elements (#212428) |
| `--text-primary` | `hsl(210, 10%, 93%)` | Main text (#ECEDEF) |
| `--text-secondary` | `hsl(210, 8%, 62%)` | Muted text (#979A9F) |
| `--text-tertiary` | `hsl(210, 6%, 45%)` | Disabled text (#6B6E73) |
| `--border-subtle` | `hsla(0, 0%, 100%, 0.06)` | Subtle borders |
| `--border-default` | `hsla(0, 0%, 100%, 0.10)` | Default borders |
| `--border-strong` | `hsla(0, 0%, 100%, 0.15)` | Emphasis borders |

#### Accent Colors
| Name | Value | Gradient | Usage |
|------|-------|----------|-------|
| Blue (Primary) | `#3B82F6` | `#3B82F6 -> #2563EB` | CTAs, links, primary actions |
| Purple | `#8B5CF6` | `#8B5CF6 -> #7C3AED` | TMS, Calls module accent |
| Green | `#10B981` | `#10B981 -> #059669` | Messenger, success states |
| Amber | `#F59E0B` | `#F59E0B -> #D97706` | Dashboard, warnings |
| Red | `#EF4444` | `#EF4444 -> #DC2626` | Destructive, errors, failed |
| Orange | `#F97316` | `#F97316 -> #EA580C` | Broken tests, attention |

#### Status Colors
| Status | Background | Text | Border |
|--------|-----------|------|--------|
| Passed/Success | `rgba(16,185,129,0.10)` | `#10B981` | `rgba(16,185,129,0.20)` |
| Failed/Error | `rgba(239,68,68,0.10)` | `#EF4444` | `rgba(239,68,68,0.20)` |
| Skipped/Muted | `rgba(107,114,128,0.10)` | `#6B7280` | `rgba(107,114,128,0.20)` |
| Broken/Warning | `rgba(249,115,22,0.10)` | `#F97316` | `rgba(249,115,22,0.20)` |
| In Progress | `rgba(59,130,246,0.10)` | `#3B82F6` | `rgba(59,130,246,0.20)` |
| Blocked | `rgba(139,92,246,0.10)` | `#8B5CF6` | `rgba(139,92,246,0.20)` |

### 1.2 Typography

| Level | Size | Weight | Line Height | Letter Spacing | Usage |
|-------|------|--------|-------------|----------------|-------|
| Display | 32px | 700 | 1.2 | -0.02em | Page hero titles |
| H1 | 24px | 700 | 1.3 | -0.015em | Page titles |
| H2 | 20px | 600 | 1.35 | -0.01em | Section headings |
| H3 | 16px | 600 | 1.4 | 0 | Card titles |
| Body | 14px | 400 | 1.5 | 0 | Main content text |
| Body Small | 13px | 400 | 1.5 | 0 | Secondary text |
| Caption | 12px | 500 | 1.4 | 0.01em | Labels, timestamps |
| Tiny | 11px | 500 | 1.3 | 0.02em | Badges, counters |
| Mono | 13px | 500 | 1.4 | 0 | Code, keys, IDs |

**Font Stack**: `"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`
**Mono Font**: `"JetBrains Mono", "Fira Code", "SF Mono", monospace`

### 1.3 Spacing & Grid

| Token | Value | Usage |
|-------|-------|-------|
| `--space-1` | 4px | Minimal gaps |
| `--space-2` | 8px | Tight element spacing |
| `--space-3` | 12px | Default element gaps |
| `--space-4` | 16px | Card padding, section gaps |
| `--space-5` | 20px | Component spacing |
| `--space-6` | 24px | Section spacing |
| `--space-8` | 32px | Major section breaks |
| `--space-10` | 40px | Page-level spacing |
| `--space-12` | 48px | Hero spacing |

**Grid**: 12-column layout, 8px base unit, 16px gutters
**Max content width**: 1280px (xl), 1024px (lg) for centered pages
**Border Radius**: 6px (small), 8px (medium), 12px (large), 16px (xlarge), 9999px (pill)

### 1.4 Shadows & Elevation

| Level | Shadow | Usage |
|-------|--------|-------|
| Level 0 | none | Flat elements |
| Level 1 | `0 1px 2px rgba(0,0,0,0.3), 0 1px 3px rgba(0,0,0,0.15)` | Cards, inputs |
| Level 2 | `0 4px 12px rgba(0,0,0,0.25), 0 1px 4px rgba(0,0,0,0.15)` | Dropdowns, popovers |
| Level 3 | `0 8px 24px rgba(0,0,0,0.35), 0 2px 8px rgba(0,0,0,0.2)` | Modals, dialogs |
| Level 4 | `0 16px 48px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)` | Floating panels |

### 1.5 Animations & Micro-interactions

**Principles**: Every state change should be animated. Transitions should feel physical and responsive. No jarring jumps.

| Animation | Duration | Easing | Trigger |
|-----------|----------|--------|---------|
| Hover lift | 200ms | `cubic-bezier(0.2, 0, 0, 1)` | Card/button hover |
| Click scale | 100ms | `ease-out` | Button press (scale 0.97) |
| Page enter | 350ms | `cubic-bezier(0.16, 1, 0.3, 1)` | Route change (fade+slide up 12px) |
| Modal open | 250ms | `cubic-bezier(0.16, 1, 0.3, 1)` | Dialog appear (scale 0.95->1 + fade) |
| Modal close | 150ms | `ease-in` | Dialog dismiss |
| Dropdown open | 150ms | `cubic-bezier(0.16, 1, 0.3, 1)` | Menu appear (scale Y 0.95->1) |
| Toast enter | 300ms | `cubic-bezier(0.21, 1.02, 0.73, 1)` | Notification slide-in from right |
| Skeleton shimmer | 1.5s | `ease-in-out infinite` | Loading placeholder |
| Stagger children | 50ms delay each | `ease-out` | List item cascade |
| Progress fill | 600ms | `cubic-bezier(0.4, 0, 0.2, 1)` | Bar/ring fill |
| Typing dots | 1.2s | `ease-in-out infinite` | Bounce cascade (3 dots) |
| Reaction float | 2.5s | `ease-out` | Emoji float upward + fade |
| Sidebar collapse | 200ms | `cubic-bezier(0.2, 0, 0, 1)` | Width 260px <-> 68px |
| Kanban drag | real-time | spring physics | Card movement with shadow |

**Hover Effects**:
- Cards: translateY(-2px) + shadow expand + border glow (accent at 20% opacity)
- Buttons: background lighten + subtle glow
- Links: color transition + underline slide-in from left
- Icons: scale(1.1) + color shift
- Avatars: ring-2 ring-accent/50 + scale(1.05)

**Scroll Effects**:
- Sticky headers: backdrop-blur-xl + border-bottom appears on scroll
- Parallax: None (keep it clean)
- Infinite scroll: Loading spinner at bottom + skeleton items

---

## 2. GLOBAL LAYOUT

### 2.1 App Shell

```
+------------------------------------------------------------------+
| Sidebar (260px)  |  Main Content Area (flex-1)                    |
| +--------------+ | +--------------------------------------------+ |
| | Logo + H     | | | Page Header (sticky, backdrop-blur)        | |
| | ------------ | | +--------------------------------------------+ |
| | Navigation   | | |                                            | |
| |  Dashboard   | | |  Page Content                              | |
| |  Messenger   | | |  (scrollable, padded 24px)                 | |
| |  Calls       | | |                                            | |
| |  Tracker     | | |                                            | |
| |  TMS         | | |                                            | |
| | ------------ | | |                                            | |
| | Projects     | | |                                            | |
| |  Project 1   | | |                                            | |
| |  Project 2   | | |                                            | |
| | ------------ | | |                                            | |
| | User Profile | | |                                            | |
| +--------------+ | +--------------------------------------------+ |
+------------------------------------------------------------------+
```

**Sidebar**: Always dark (#19171D), regardless of theme. Slim scrollbar. Collapsible to 68px with icon-only mode. Smooth 200ms transition. Active nav item has 3px left accent border + text-white + bg-white/8%.

**Main Area**: bg-primary background. Scrollable Y-axis. Content max-width varies by page.

### 2.2 Sidebar Design (CRITICAL — this is always visible)

**Header (49px)**:
- Logo: Gradient "H" badge (blue->purple) 32x32 + "Hub" text bold
- Compose button: "+" icon, rounded, accent hover glow
- Collapse toggle: ChevronLeft icon, rotates 180deg when collapsed

**Navigation Items**:
- Icon (20x20) + Label (14px medium)
- Spacing: 4px gap between items, 8px padding
- Active: `bg-white/8% text-white border-l-3 border-blue-500`
- Hover: `bg-white/6% text-white/90`
- Transition: 150ms background + color

**Project List**:
- Section title: "Проекты" (12px, uppercase, text-white/40%, letter-spacing 0.05em)
- Project item: Colored dot (4px, project accent) + name (13px)
- Hover: bg-white/6%
- Scrollable with thin scrollbar (4px, white/10%)

**User Profile (bottom)**:
- Avatar (32x32, rounded-full) + Name (13px) + Role badge (11px)
- Click: Dropdown with Settings, Profile, Appearance, Logout
- Online indicator: 8px green dot on avatar corner

### 2.3 Mobile Layout (< 768px)

- Sidebar hidden (overlay on hamburger click with backdrop blur)
- Top bar: 56px height, hamburger + logo + user avatar
- Bottom nav: 5 icons (Dashboard, Messenger, Calls, Tracker, TMS), 60px height
- Active: Filled icon + accent color + label text
- Safe area: bottom padding for iOS

---

## 3. BLOCK 1: AUTH & LOGIN

### 3.1 Login Page

**Layout**: Full-screen split
- **Left (55%)**: Animated dark gradient background with floating feature cards
  - Mesh gradient: deep blue -> purple -> dark, shifting 20s infinite
  - 4 floating cards with stagger animation (7-9s loops):
    - "Tracker" (blue icon)
    - "Messenger" (green icon)
    - "Calls" (purple icon)
    - "TMS" (violet icon)
  - Cards: glass morphism (bg-white/5%, border-white/10%, backdrop-blur-xl)
  - Cards float with subtle translateY oscillation

- **Right (45%)**: Login form
  - Centered card (max-width 400px)
  - Logo: Gradient "Hub" text (32px, blue->purple->pink)
  - Subtitle: "Войдите в свой аккаунт" (14px, text-secondary)
  - **Magic Link Form**:
    - Email input (48px height, rounded-12, bg-tertiary, border-subtle)
    - "Получить код" button (48px, full-width, gradient blue, shadow-blue/25%)
    - Code input (6 digits, large centered, spaced chars)
    - "Войти" button after code entry
  - Divider: "или" with lines
  - **Telegram Button**: Branded Telegram blue, icon + "Войти через Telegram"
  - Footer: "Нет аккаунта? Зарегистрироваться" link

**Animations**:
- Form fades in 400ms on load
- Input focus: border glows blue, slight scale(1.01)
- Button hover: gradient lightens, shadow expands
- Success: Checkmark animation, redirect

### 3.2 Onboarding (3 Steps)

**Step 1 — Profile Setup**:
- Avatar upload (click circle placeholder, 96x96)
- Name, display name, phone, position, department inputs
- Timezone and language selects
- Birthday date picker

**Step 2 — Platform Tour**:
- 6 module cards in 2x3 grid:
  - Each: icon (gradient bg) + title + short description
  - Non-interactive, informational
- Smooth stagger reveal animation

**Step 3 — Preferences**:
- Theme picker (3 cards: Light/Dark/System with preview thumbnails)
- Accent color picker (5 circles with selection ring)
- "Начать работу" primary button

**Progress Indicator**:
- 3 steps connected by line
- Active: filled circle + accent border
- Completed: checkmark in circle
- Upcoming: empty circle, muted

### 3.3 Profile Settings

- Avatar: 96x96, click-to-upload, hover overlay with camera icon
- Form fields in 2-column grid (desktop), 1-column (mobile)
- Custom status: emoji picker + text input + duration select
- 12 preset statuses with emojis
- Work hours: start/end time inputs with calculated duration badge
- Active sessions list: device icon + IP + last active + "Revoke" button
- Current session highlighted with green "Текущая" badge

---

## 4. BLOCK 2: DASHBOARD

### 4.1 Main Dashboard

**Header**:
- Greeting: "Добрый день, {Name}" in gradient text (Display size)
- Date/time below (14px, text-secondary)
- Quick actions row: 4 buttons (New Task, Write Message, Create Event, Search)
  - Each: icon + label, rounded-12, border, hover-lift

**Widget Grid**:
- react-grid-layout, 12 columns, 8px margin
- Each widget: rounded-16 card, bg-secondary, border-subtle
- Widget header: title (14px semibold) + 3-dot menu (refresh, hide)
- Drag handle: GripVertical icon in header, cursor-move
- Resize handle: SE corner, 12x12 triangle indicator
- On drag: shadow-4 + slight opacity change + z-index boost

**Widget Types** (11 total):
1. **My Tasks**: Issue list with status dots, priority indicators. Click navigates to issue.
2. **Unread Messages**: Channel list with unread counts. Badge numbers.
3. **Upcoming Calls**: Next 3 calls with time, channel name, "Join" button (accent).
4. **Task Status Chart**: Donut/pie chart with status breakdown. Hover tooltips.
5. **Burndown Chart**: Line chart, ideal vs actual. Gradient fill below line.
6. **Velocity Chart**: Bar chart, sprints comparison. Tooltip with values.
7. **Test Trends**: Stacked area chart (passed/failed/skipped). Last 10 runs.
8. **Team Activity**: Avatar + action + timestamp feed. Scrollable.
9. **System Health**: Green/yellow/red indicators for services.
10. **Storage Usage**: Circular progress with GB used / total.
11. **Deadlines**: Issue list sorted by due_date. Overdue items in red.

**Add Widget Button**: "+ Добавить виджет" bottom area, opens dropdown with available widgets.

**Templates**: Auto-apply by role (Dev/QA/Manager/Admin) on first visit.

---

## 5. BLOCK 3: MESSENGER (Reference: Telegram)

### 5.1 Channel List (Left Sidebar, 272px)

**Header**: Search input (32px, rounded-pill, bg-tertiary, magnifying glass icon)

**Action Buttons**: Row of 3 icon buttons (Bookmarks, New DM, Create Channel)

**Channel Groups**:
- **Starred** (collapsed by default if empty)
- **Direct Messages**
- **Channels**

**Channel Item (48px height)**:
- Left: Avatar (36x36) or channel icon (#/lock/group)
  - DM: User avatar with presence dot (green=online, yellow=away, gray=offline)
  - Channel: Styled icon in colored circle
- Center: Channel name (14px semibold) + last message preview (12px, text-secondary, 1 line truncate)
- Right: Timestamp (11px) + unread badge (red circle, white text, min-width 20px)
- Muted: Bell-off icon overlay, dimmed opacity
- Active: bg-accent/10%, border-l-2 accent

**Context Menu (right-click)**:
- Star/Unstar
- Mute (submenu: 1h, 2h, 8h, 1d, Forever)
- Notification mode (All, Mentions only, Nothing)
- Mark as read
- Leave channel
- Delete (if owner)

### 5.2 Chat View (Main Area)

**Chat Header (56px, sticky)**:
- Channel icon + name (16px semibold) + member count
- Right actions: Search (magnifying glass), Call (phone icon), Pin (pin icon), Info (i icon)
- Bottom border: 1px border-subtle
- backdrop-blur-xl on scroll

**Message List (scrollable)**:
- Messages in flat Slack-style layout (NO bubbles):
  ```
  Avatar (36x36)  Name (13px bold)  Timestamp (11px muted)
                  Message content (14px)
                  [Attachments / Link previews / Reactions]
  ```
- **Grouping**: Same sender within 2min = no repeated avatar/name, just content
- **Date separators**: Centered pill ("Сегодня", "Вчера", "25 марта")
- **New messages divider**: Red line + "Новые сообщения" label
- **Message hover**: bg-tertiary/50% + floating action bar appears (top-right of message)

**Message Action Bar (floating, on hover)**:
- 6 icon buttons in a row: React, Reply, Thread, Pin, Bookmark, Forward, Edit (own), Delete (own)
- Rounded-8, shadow-2, bg-elevated, border-subtle
- Quick reaction: first 6 popular emojis in a row

**Message Content Types**:
- **Text**: Markdown rendered (bold, italic, code, code blocks, links, lists, blockquotes)
- **Sticker**: Large centered emoji (72px), no background
- **GIF**: Inline image (max-height 256px), rounded-8, click to expand
- **Voice**: Waveform visualizer + play button + duration (MM:SS)
- **File**: Pill card with file icon + name + size + download button
- **Image**: Inline (max-height 256px), rounded-8, click for lightbox
- **Poll**: Card with question (bold), options (progress bars), vote counts, anonymous badge
- **Link Preview**: Card with OG image (left), title + description + domain (right)
- **Forwarded**: Thin left accent border + "Forwarded from {name}" label
- **Edited**: "(изм.)" italic text after timestamp

**Reactions Row (below message)**:
- Emoji chips: rounded-pill, border, padding 2px 8px
- Count next to emoji
- Own reaction: accent border + bg-accent/10%
- Click to toggle, hover to see who reacted

### 5.3 Message Composer (Bottom, Sticky)

**Container**: Rounded-16 border, bg-secondary, focus-within: border-accent/50% + shadow glow

**Layout (single row, items vertically centered)**:
```
[Attach] [                Text Input                    ] [Emoji] [GIF] [Sticker] [Poll] [Voice] [Send]
```

**Text Input**:
- Auto-expanding textarea (min 44px, max 200px)
- Placeholder: "Написать сообщение..." (text-tertiary)
- Supports @mention autocomplete (popup above)
- Markdown shortcuts: **bold**, *italic*, `code`

**Buttons** (each 40x36, icon 20x20, text-secondary, hover text-primary):
- Attach: Paperclip icon. Click opens file picker. Pending files show as chips above input.
- Emoji: Smiley icon. Opens 5x4 grid popover with popular emojis.
- GIF: Film icon. Opens Giphy search popover (2-column masonry grid, search at top).
- Sticker: Image icon. Opens sticker pack selector (4 packs, 4x4 grid per pack, tab navigation).
- Poll: Chart icon. Opens create poll dialog.
- Voice: Mic icon. Click enters recording mode (waveform + timer + send/cancel).
- Send: Arrow icon. Accent color when text present, muted when empty. Scale animation on click.

**Reply Preview (above input when replying)**:
- Thin accent left border + original message preview (truncated) + close X button
- Slide-down animation 150ms

**@Mention Autocomplete**:
- Popup above cursor position
- User list filtered by typed text
- Arrow key navigation, Enter to select
- Shows avatar + name + role

### 5.4 Channel Creation Dialog

- Channel name input (required)
- Description textarea (optional)
- Type selector: 3 cards (Public, Private, Group) with icons
- Member search + multi-select chips
- Create button (accent gradient)

### 5.5 Presence & Status

- **Online**: 8px green dot (bottom-right of avatar), pulsing subtle glow
- **Away**: 8px yellow dot
- **Offline**: 8px gray dot (no animation)
- **Custom Status**: Emoji + text shown in sidebar user profile + channel member list

---

## 6. BLOCK 4: CALLS (Reference: Yandex Telemost)

### 6.1 Call Calendar Page

**Tabs**: "Calendar" | "About Calls"

**Calendar Views**: Month / Week / Day / Agenda
- **Month**: Grid cells with event dots (colored by type)
- **Week**: 7-column time grid with event blocks
- **Day**: Full 24h timeline with event blocks
- **Agenda**: Chronological list of upcoming events (30 days)

**Event Block**: Rounded-6 card in accent color, title + time, click to see details
**Create Button**: "+" floating or "New Event" button, opens CreateCallDialog

### 6.2 Create Call Dialog

**Form Fields**:
- Title (text, required)
- Date + Time inputs (side by side)
- Duration quick-select: 4 pills (30min, 1h, 1.5h, 2h) — active = accent bg
- Reminder quick-select: 5 pills (None, 5min, 15min, 30min, 1h) — active = accent bg
- Channel selector (required)
- Recurrence: 4 pills (None, Daily, Weekly, Weekdays)
- Participants: Search + multi-select with conflict warnings
- Description textarea

**Conflict Warning**: Yellow card with "User X busy at this time" + conflicting call title

### 6.3 Call Room (Full-screen overlay)

**Layout**:
```
+------------------------------------------------------------------+
| Call Header: Room name | Participant count | Timer | Fullscreen   |
+------------------------------------------------------------------+
|                                                                    |
|     Screen Share Area (if active)      |  Participant Grid         |
|     or Full Participant Grid           |  (adaptive columns)       |
|                                                                    |
+------------------------------------------------------------------+
| Controls: [Mic] [Camera] [Share] [Reactions] [Chat] [Settings] [Leave] |
+------------------------------------------------------------------+
```

**Participant Tile**:
- Avatar (large, centered, 80-120px depending on grid size)
- Name (14px) below avatar
- Speaking indicator: Green ring around avatar + subtle pulse animation
- Muted: Small red mic-off badge (bottom-right corner)
- Hand raised: Bouncing hand emoji above avatar

**Adaptive Grid**:
- 1 person: 1x1 (full width, large avatar)
- 2-4: 2x2 grid
- 5-6: 3x2
- 7-9: 3x3
- 10+: 4xN with scroll

**Controls Bar (bottom, 64px)**:
- Rounded-pill buttons on dark bg
- Mic toggle: Green when on, red when off
- Camera toggle: Blue when on, gray when off (confirmation dialog to enable)
- Screen share: Purple when active
- Reactions: Opens emoji picker (hand + 5 emojis), selected emoji floats upward
- Chat: Opens right sidebar (280px) with ephemeral messages
- Settings: Opens modal (device selection, noise suppression toggle)
- Leave: Red button, always visible

**Floating Reactions**:
- Emoji + sender name
- Float upward 200px over 2.5s
- Fade out + scale down at end
- Multiple reactions can overlap

**Recording Indicator**: Red "REC" badge + timer (top-left), pulsing red dot

### 6.4 About Calls Tab

- 3-step guide (numbered circles + descriptions)
- 6 feature cards (2x3 grid):
  - Audio calls (up to 50)
  - Screen Share
  - Recording
  - Scheduling
  - Hand Raise
  - Ephemeral Chat

---

## 7. BLOCK 5: CALENDAR

### 7.1 Views

**Month View**: 7x5 grid (or 7x6), day cells with event dots. Today highlighted with accent ring. Click cell to see day events.

**Week View**: 7-column time grid (08:00-23:00 visible), current time red line. Event blocks as colored cards spanning their duration.

**Day View**: Full 24h timeline, hourly slots. Events as wide cards with title + time + participants.

**Agenda View**: Chronological list grouped by date. Each event: time + title + participants avatars + channel name.

### 7.2 Event Details Popup

- Title (bold)
- Date + time + duration
- Recurrence label (if repeating)
- Participants (avatar row)
- Channel
- Description
- Reminder setting
- Edit/Delete buttons

---

## 8. BLOCK 6: TASK TRACKER (Reference: Linear)

### 8.1 Kanban Board

**Layout**: Horizontal scrollable columns (min-height: calc(100vh - 7rem))

**Column** (280px width, flex-shrink 0):
- Header: Status name (14px semibold) + count badge + "+" add button
- Cards: Stacked vertically, 8px gap
- Drop zone: Dashed border highlight when dragging over

**Issue Card (Kanban)**:
- Rounded-8, bg-secondary, border-subtle
- Hover: translateY(-1px) + shadow expand + border-accent/20%
- **Content**:
  - Row 1: Type emoji + Issue key (mono, 12px, muted)
  - Row 2: Title (14px, max 2 lines, ellipsis)
  - Row 3: Priority dot (colored) + Story points badge + Checklist progress (X/Y)
  - Row 4: Assignee avatar (24x24) + Labels (colored dots)
- Drag: Shadow-4, scale(1.02), semi-transparent (opacity 0.9)

**Swimlanes** (optional): Group by Assignee, Priority, or Type
- Horizontal lane header with avatar/icon + name + collapse toggle

**Filter Bar (top)**:
- Search input (left)
- Status filter (multi-select dropdown)
- Assignee filter (avatar dropdown)
- Priority filter (colored dots dropdown)
- Reset filters button

### 8.2 Issue Detail Page

**Two-column layout**: Content (flex-1) | Sidebar (320px)

**Left Column**:
1. **Breadcrumb**: Project > Issues > KEY-123
2. **Title**: Inline-editable, 20px semibold. Click to edit, Enter to save, Esc to cancel.
3. **Description**: Rich text editor (Plate). Toggle between edit/view mode. Supports headings, bold, italic, code, lists, quotes, links.
4. **Attachments Section**: File list with icons + upload button. Drag-and-drop zone.
5. **Checklist**: Checkbox list with progress bar. Add/remove items. Inline editing.
6. **Tabs**: Comments | Activity
   - **Comments**: Avatar + name + time + content + reactions. Composer at bottom.
   - **Activity**: Timeline of field changes with old->new values and timestamps.

**Right Sidebar**:
- **Status**: Select with colored badge (workflow states)
- **Assignee**: Avatar + name, click to change (member search dropdown)
- **Priority**: Colored badge select (Critical=red, High=orange, Medium=yellow, Low=blue, Trivial=gray)
- **Type**: Icon + label select (Bug, Task, Story, Epic, etc.)
- **Story Points**: Number select (Fibonacci: 1,2,3,5,8,13,21)
- **Sprint**: Select (active/planning sprints)
- **Labels**: Multi-select tag chips (colored)
- **Due Date**: Date picker, overdue shown in red
- **Time Tracking**: Timer button + logged time display
- **Created/Updated**: Timestamps (muted text)
- **GitHub/GitLab Links**: PR/commit list with status icons

### 8.3 Backlog View

**Two sections**:
- **Active Sprint** (top): Sprint name + progress + issues. Drag to reorder.
- **Backlog** (bottom): Unplanned issues. Drag from backlog to sprint.

**Sprint Actions**: Start Sprint, Complete Sprint, Edit Sprint buttons

### 8.4 Issue List View

- Table with sortable columns: Key, Title, Status, Priority, Assignee, Updated
- Row hover: bg-tertiary/50%
- Click row: navigate to detail
- Keyboard navigation: arrow keys to move, Enter to open

---

## 9. BLOCK 7: TMS (Reference: Allure TestOps)

### 9.1 Test Cases Tree View

**Layout**: Split pane — Tree (300px left) | Case list (right)

**Tree (left)**:
- Folder hierarchy with expand/collapse
- Folder icon + name + case count badge
- Drag-and-drop to reorder/move
- Right-click: Rename, Delete, Create subfolder

**Case List (right)**:
- Table: Key (mono), Title, Priority (colored), Status (badge), Automation (icon)
- Row actions: Edit, Run, Delete
- Bulk actions: Select multiple, add to run, delete

### 9.2 Test Case Detail

**Header**: Case key + Title (editable) + Status badge + Priority badge

**Tabs**: Steps | Details | History | Coverage

**Steps Tab** (main):
- Table: # | Action | Expected Result
- BDD mode: Keyword (Given/When/Then) colored badges
- Add/remove/reorder rows
- Preconditions section (rich text above steps)
- Postconditions section (below steps)

**Details Tab**: Description, automation_id, full_name, labels, components, linked issues

**History Tab**: Version timeline with diff (old->new), rollback button per version

**Coverage Tab**: Linked issues with last test result status

### 9.3 Test Run List

- Card grid (3 columns)
- Card: Title + status badge + progress bar (passed/failed/skipped/broken) + stats + date + assignee avatars
- Progress bar: segmented, colored sections, animated fill

### 9.4 Test Run Execution

**Layout**: Two-column — Case list (left) | Execution panel (right)

**Case List**: Status icon + case key + title. Click to select for execution.

**Execution Panel**:
- Case title + description
- Steps table with status buttons per step (Pass/Fail/Skip)
- Comment field
- File attachment
- Overall status buttons: Passed (green), Failed (red), Broken (orange), Skipped (gray)

### 9.5 Reports

**Trends Chart**: Stacked area chart (last N runs). Colors: green/red/orange/gray. Hover tooltips.

**Coverage Matrix**: Requirements (rows) x Test Cases (columns). Cell color = last result status. Click cell for details.

**Traceability Matrix**: Same as coverage but grouped by requirement → test cases → status.

**Flaky Tests**: Table with case key + title + status history (colored dots timeline).

### 9.6 Autotest Integration Page

**Top**: TriggerCIPanel — select test cases + branch + launch pipeline button

**Sections** (numbered cards):
1. Webhook URLs (code blocks with copy button)
2. API Token guide (expandable, step-by-step with permissions table)
3. Test case ID format (PROJ-TC-N)
4. Framework examples (collapsible: Python, JS, Java)
5. CI/CD setup (collapsible: GitHub Actions, GitLab CI)
6. Matching strategies table (6 strategies, priority order)
7. Hub CI variables documentation

---

## 10. BLOCK 8: GITHUB/GITLAB INTEGRATION

### 10.1 Integration Settings Page

**Tabs**: GitHub | GitLab

**Each Tab**:
- Connected repos list (card per repo: icon + path + disconnect button)
- Connect form:
  - Owner/path input
  - Access token (password input) with expandable "How to create token" guide
  - Webhook secret input with setup instructions
  - Connect button

**Token Guides**: Expandable cards with step-by-step numbered instructions, permission tables, colored tips

### 10.2 Pipeline Status Widget (on issue detail)

- Pipeline list: status badge (colored) + branch + SHA (truncated) + duration + link
- Status icons: spinning circle (running), green check (success), red X (failed), gray dash (skipped)
- Auto-refresh every 30s with subtle pulse

### 10.3 Git Links on Issue

- PR/MR cards: status icon + title + author avatar + branch names
- Commit list: SHA (mono) + message (truncated) + author + timestamp
- Create branch button: generates `KEY-123-title-slug` format

---

## 11. BLOCK 9: NOTIFICATIONS

### 11.1 Notification Bell (header)

- Bell icon in sidebar user area or top bar
- Unread count: red circle badge (animated scale-in)
- Click: opens notification panel (popover, 380px wide, max-height 500px)

### 11.2 Notification Panel

- Header: "Уведомления" + "Прочитать все" link + "Очистить" link
- Items: Icon (type-colored) + title (14px semibold) + body (13px muted) + timestamp (11px)
- Unread: left accent border + slightly brighter bg
- Click: navigate to entity, mark as read
- Hover: bg-tertiary/50%
- Empty state: bell icon + "Нет уведомлений"
- Infinite scroll with loading spinner

### 11.3 Notification Settings

- Matrix: 9 event types x 3 channels (In-app, Telegram, Push)
- Checkboxes at intersections
- Row labels: Назначение, Упоминание, Смена статуса, Комментарий, Сообщение, Звонок, Прогон завершён, Дедлайн, Блокировка
- Column labels with icons

**Quiet Hours Section**:
- Start time + End time inputs
- Save button
- Description: "Несрочные уведомления подавляются"

### 11.4 Push Permission Banner

- Top banner (slide-down): "Включите push-уведомления" + Allow + Dismiss buttons
- Appears 3s after page load if not subscribed
- Dismissable, remembers choice

---

## 12. BLOCK 10: ADMIN PANEL

### 12.1 Admin Dashboard

- 4 KPI cards: Users, Projects, Issues, DAU/WAU/MAU
- Each card: big number + change indicator (green up, red down) + mini sparkline
- Widget grid below (same system as user dashboard but admin metrics)

### 12.2 User Management

- Table: Avatar + Name + Email + Role badge + Status badge + Last active + Actions
- Actions: Edit role, Deactivate, Reassign tasks, GDPR delete
- Search + filter by role/status
- Invite button: email input + role selector

### 12.3 Audit Log

- Table: Timestamp + Action + Entity + User avatar + IP + Details (expandable)
- Filters: Date range, Action type, Entity type, User
- Auto-refresh every 30s
- Export CSV button

### 12.4 Resource Limits

- Settings form: max file size, max projects per user, max storage per project
- Save button with confirmation

---

## 13. BLOCK 11: TRASH

### 13.1 Trash Page

**Tabs**: "Deleted Items" | "About Trash"

**Items Tab**:
- Table: Type icon + Name + Deleted by + Days remaining + Restore button
- Empty state: trash icon + "Корзина пуста"
- Admin: additional "Permanent Delete" action (red, confirmation dialog)

**About Tab**: 3-step lifecycle + 6 type cards explaining what can be trashed

---

## 14. RESPONSIVE BEHAVIOR

### Mobile (< 768px)
- Sidebar: overlay with backdrop blur
- Top bar with hamburger + logo + avatar
- Bottom navigation (5 tabs)
- Single column layouts everywhere
- Cards stack vertically
- Modals become full-screen sheets (slide up from bottom)
- Composer: simplified (hide some buttons, show on long-press)

### Tablet (768-1024px)
- Sidebar collapsible (icon-only by default)
- Two-column layouts where applicable
- Reduced widget grid columns

### Desktop (> 1024px)
- Full sidebar expanded
- All multi-column layouts active
- Maximum widget density

---

## 15. EMPTY STATES (for every section)

**Pattern**: Centered vertically, max-width 320px
- Large icon in gradient circle (48x48 in 80x80 container)
- Title (18px semibold)
- Description (14px muted)
- CTA button (optional, accent gradient)
- Decorative floating dots (subtle, animated pulse)

**Examples**:
- No projects: "Создайте первый проект" + grid icon
- No messages: "Выберите канал" + chat icon
- No tasks: "Нет задач в этом спринте" + clipboard icon
- No test runs: "Запустите первый прогон" + play icon
- Empty trash: "Корзина пуста" + trash icon
- No notifications: "Нет уведомлений" + bell icon

---

## 16. LOADING STATES

**Skeleton Pattern**: Animated shimmer (gradient sweep left-to-right, 1.5s infinite)
- Gray rectangles matching content layout
- Rounded corners matching final content
- Stagger animation (50ms delay per row)

**Specific Skeletons**:
- Issue card: 3 lines (title, meta, avatar)
- Message: Avatar circle + 2 lines
- Channel item: Icon circle + 2 lines
- Widget: Full card placeholder
- Table row: 5 column blocks

**Spinners**: Only for button actions (not full-page)
- Small: 16x16, border-2
- Default: 20x20, border-2
- Large: 32x32, border-3

---

## 17. TOAST NOTIFICATIONS

**Position**: Top-right, stacked
**Animation**: Slide in from right (300ms, spring easing)
**Auto-dismiss**: 4s (info), 6s (success), persistent (error)
**Types**:
- **Success**: Green left border, check icon
- **Error**: Red left border, X icon, "Retry" action button
- **Info**: Blue left border, info icon
- **Warning**: Orange left border, warning icon

**Style**: bg-elevated, shadow-3, rounded-12, border-subtle, max-width 380px

---

## 18. KEYBOARD SHORTCUTS

| Shortcut | Action |
|----------|--------|
| `Ctrl+N` | New issue |
| `Ctrl+K` | Quick search |
| `Ctrl+/` | Show shortcuts |
| `Esc` | Close modal/panel |
| `Enter` | Confirm/submit |
| `Arrow keys` | Navigate lists |
| `Space` | Toggle selection |

---

## 19. DELIVERABLES EXPECTED

For each block/page generate:
1. **High-fidelity mockup** (dark theme, 1440x900 desktop)
2. **Mobile mockup** (375x812 iPhone)
3. **Component states** (default, hover, active, disabled, loading, error, empty)
4. **Animation specs** (describe transitions between states)
5. **Reference React/TSX code** (using Tailwind CSS + Radix UI patterns)

**Priority order for mockups**:
1. Messenger (most used daily)
2. Kanban Board (core workflow)
3. Issue Detail (complex page)
4. Dashboard (first impression)
5. Test Run Execution (complex flow)
6. Call Room (real-time UI)
7. Login Page (first touchpoint)
8. Settings (configuration)
9. TMS Case Tree (information architecture)
10. Calendar Views (scheduling)
11. Admin Panel (management)
