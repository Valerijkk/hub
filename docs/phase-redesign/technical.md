# Техническая документация -- Hub v2.0 Redesign

## Архитектура дизайн-системы

Hub v2.0 Redesign -- исключительно фронтенд-проект. Архитектура данных, API и бэкенд остаются без изменений. Весь визуальный слой контролируется через CSS-переменные в `app.css`, которые каскадируются через Tailwind CSS v4 `@theme inline` во все компоненты.

### Поток данных (без изменений)

```
next-themes -> .dark class на <html>
     |
     v
CSS cascade -> :root / .dark переменные
     |
     v
@theme inline -> Tailwind утилиты (bg-*, text-*, border-*)
     |
     v
Компоненты React -> className с Tailwind утилитами
```

### Ключевые файлы

| Файл | Назначение |
|------|-----------|
| `frontend/src/app.css` | Единый источник истины для дизайн-токенов, типографики, анимаций, базовых стилей |
| `frontend/index.html` | Google Fonts CDN подключение шрифтов |
| `frontend/src/components/ui/*.tsx` | 10+ Shadcn/ui компонентов (рестайлированы) |
| `frontend/src/components/layout/app-shell.tsx` | Layout shell (AppShell) |
| `frontend/src/components/layout/sidebar.tsx` | Sidebar с gradient edge и active glow |
| `frontend/src/components/layout/topbar.tsx` | Topbar с backdrop blur |
| `frontend/src/modules/**/*.tsx` | Per-page модули (все 17 страниц) |

### Принципы

1. **Семантические токены** -- никаких хардкоженных hex-значений в компонентах; только CSS-переменные
2. **Cascade-first** -- 60-70% визуальных изменений происходят автоматически через каскад при замене токенов
3. **CSS transitions для 95%** -- `motion` v12 только для layout animations и mount/unmount; все hover/focus/press через CSS
4. **Zero new dependencies** -- весь визуальный слой реализуется существующим стеком

---

## Миграция токенов

### Старая система (Shadcn HSL defaults)

```css
/* До: Shadcn стандартные HSL-токены */
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 100%;
  --primary: 222.2 47.4% 11.2%;
  --muted: 210 40% 96.1%;
  /* ... стандартные Shadcn переменные */
}
```

### Новая система (Semantic token palette)

```css
/* После: Семантическая 5-уровневая система */
:root {
  /* Surfaces (5 levels) */
  --void: #FFFFFF;
  --base: #F8F9FA;
  --raised: #F1F3F5;
  --card: #FFFFFF;
  --input: #F8F9FA;

  /* Text (4 levels) */
  --t1: #111827;
  --t2: #374151;
  --t3: #6B7280;
  --t4: #9CA3AF;

  /* Borders (3 levels) */
  --b1: rgba(0, 0, 0, 0.06);
  --b2: rgba(0, 0, 0, 0.10);
  --b3: rgba(0, 0, 0, 0.15);

  /* ... + radius, shadows, easing */
}

.dark {
  /* Surfaces -- exact mockup values */
  --void: #09090B;
  --base: #0C0C0F;
  --raised: #141419;
  --card: #18181B;
  --input: #1E1E23;

  /* Text */
  --t1: #F4F4F5;
  --t2: #A1A1AA;
  --t3: #71717A;
  --t4: #52525B;

  /* Borders */
  --b1: rgba(255, 255, 255, 0.06);
  --b2: rgba(255, 255, 255, 0.10);
  --b3: rgba(255, 255, 255, 0.15);
}
```

### Маппинг на Shadcn

Новые семантические токены маппятся на стандартные Shadcn-переменные через `@theme inline`:

| Shadcn переменная | Маппинг на новый токен |
|-------------------|----------------------|
| `--background` | `var(--void)` |
| `--foreground` | `var(--t1)` |
| `--card` | `var(--card)` |
| `--card-foreground` | `var(--t1)` |
| `--muted` | `var(--raised)` |
| `--muted-foreground` | `var(--t3)` |
| `--border` | `var(--b1)` |
| `--input` | `var(--input)` |
| `--primary` | `var(--accent)` |
| `--primary-foreground` | `#FFFFFF` |

Это обеспечивает обратную совместимость: все существующие Shadcn-компоненты, использующие `bg-card`, `text-muted-foreground`, `border-border` и т.д., автоматически получают новые цвета без изменения кода.

---

## Шрифты

### Конфигурация

```html
<!-- index.html -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
```

### CSS-переменные

```css
@theme inline {
  --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', ui-monospace, monospace;
}
```

### Применение

| Шрифт | Вес | Применение |
|-------|-----|-----------|
| Inter 400 | Regular | Body text, descriptions |
| Inter 500 | Medium | Labels, secondary headings |
| Inter 600 | Semibold | Buttons, nav items, card titles |
| Inter 700 | Bold | Page headings |
| Inter 800-900 | Extra/Black | Hero text (Dashboard greeting) |
| JetBrains Mono 400 | Regular | Code blocks, inline code |
| JetBrains Mono 500 | Medium | Task keys (HUB-123), entity IDs |

### Решение: Google Fonts CDN vs Self-hosted

**Выбрано:** Google Fonts CDN. Hub -- внутренний инструмент для 10-50 пользователей. CDN обеспечивает кеширование между сайтами, отсутствие build-оверхеда и простоту обновления. Self-hosted WOFF2 c `size-adjust` fallback -- более правильный подход для публичных продуктов, но оверкилл для данного кейса.

---

## Тёмная тема

Значения тёмной темы взяты непосредственно из 11 HTML-макетов. Каждый токен имеет точное соответствие в макете.

### Поверхности

| Токен | Значение | Источник в макете |
|-------|----------|------------------|
| `--void` | `#09090B` | Body background |
| `--base` | `#0C0C0F` | Sidebar background, content area |
| `--raised` | `#141419` | Hover surfaces, secondary panels |
| `--card` | `#18181B` | Cards, widgets, panels |
| `--input` | `#1E1E23` | Input fields, textareas, selects |

### Текст

| Токен | Значение | Контраст на `--card` |
|-------|----------|---------------------|
| `--t1` | `#F4F4F5` | 13.5:1 (AAA) |
| `--t2` | `#A1A1AA` | 6.2:1 (AA) |
| `--t3` | `#71717A` | 4.0:1 (AA large text) |
| `--t4` | `#52525B` | 2.7:1 (decorative/disabled) |

### Границы

| Токен | Значение |
|-------|----------|
| `--b1` | `rgba(255, 255, 255, 0.06)` |
| `--b2` | `rgba(255, 255, 255, 0.10)` |
| `--b3` | `rgba(255, 255, 255, 0.15)` |

### Glow-цвета (status-specific)

| Токен | Значение | Применение |
|-------|----------|-----------|
| `--gB` | `rgba(59, 130, 246, 0.15)` | Blue glow (info, links) |
| `--gG` | `rgba(34, 197, 94, 0.15)` | Green glow (success, online) |
| `--gR` | `rgba(239, 68, 68, 0.15)` | Red glow (error, critical) |
| `--gA` | `rgba(251, 191, 36, 0.15)` | Amber glow (warning) |
| `--gP` | `rgba(168, 85, 247, 0.15)` | Purple glow (special) |

---

## Светлая тема

Светлая тема не имеет макетов -- все значения derived из тёмной палитры с инверсией контрастности и проверкой WCAG AA.

### Поверхности

| Токен | Значение | Принцип derivation |
|-------|----------|-------------------|
| `--void` | `#FFFFFF` | Максимально светлый фон |
| `--base` | `#F8F9FA` | Subtle grey для разделения зон |
| `--raised` | `#F1F3F5` | Hover/secondary panels |
| `--card` | `#FFFFFF` | Чистый белый для карточек |
| `--input` | `#F8F9FA` | Subtle grey для полей ввода |

### Текст

| Токен | Значение | Контраст на `--card` |
|-------|----------|---------------------|
| `--t1` | `#111827` | 16.1:1 (AAA) |
| `--t2` | `#374151` | 9.7:1 (AAA) |
| `--t3` | `#6B7280` | 5.0:1 (AA) |
| `--t4` | `#9CA3AF` | 2.8:1 (decorative/disabled) |

### Границы

| Токен | Значение |
|-------|----------|
| `--b1` | `rgba(0, 0, 0, 0.06)` |
| `--b2` | `rgba(0, 0, 0, 0.10)` |
| `--b3` | `rgba(0, 0, 0, 0.15)` |

### WCAG AA верификация

Все текстовые токены на всех поверхностях проверены на соответствие WCAG AA:
- **t1 на void/base/card:** >= 15:1 (AAA)
- **t2 на void/base/card:** >= 9:1 (AAA)
- **t3 на void/base/card:** >= 4.5:1 (AA)
- **t4:** Используется только для disabled/decorative элементов (не требует AA)
- **UI components (borders, icons):** >= 3:1 (AA)

---

## Акцентные цвета

### Механизм

Акцентный цвет управляется через CSS data-атрибут `[data-accent="X"]` на `<html>`. Каждый акцент определяет набор переменных:

```css
[data-accent="blue"] {
  --accent: #3B82F6;
  --accent-hover: #2563EB;
  --accent-glow: rgba(59, 130, 246, 0.25);
  --accent-subtle: rgba(59, 130, 246, 0.10);
}

[data-accent="purple"] {
  --accent: #8B5CF6;
  --accent-hover: #7C3AED;
  --accent-glow: rgba(139, 92, 246, 0.25);
  --accent-subtle: rgba(139, 92, 246, 0.10);
}

/* Аналогично для red, green, orange */
```

### Доступные акценты

| Акцент | Primary | Hover | Glow |
|--------|---------|-------|------|
| Red | `#EF4444` | `#DC2626` | `rgba(239, 68, 68, 0.25)` |
| Blue | `#3B82F6` | `#2563EB` | `rgba(59, 130, 246, 0.25)` |
| Green | `#22C55E` | `#16A34A` | `rgba(34, 197, 94, 0.25)` |
| Purple | `#8B5CF6` | `#7C3AED` | `rgba(139, 92, 246, 0.25)` |
| Orange | `#F97316` | `#EA580C` | `rgba(249, 115, 22, 0.25)` |

### Применение акцентов

- Focus ring glow на всех interactive элементах
- Primary button background
- Active sidebar item indicator
- Selected tab indicator
- Link color
- Checkbox/radio/switch checked state

---

## Layout Shell

### AppShell

```
+--------------------------------------------------+
|                    Topbar                         |
+--------+-----------------------------------------+
|        |                                         |
| Side-  |          Content Area                   |
| bar    |          (padding: 28px)                |
|        |                                         |
| (240px |                                         |
|  or    |                                         |
|  64px  |                                         |
|  col-  |                                         |
|  laps) |                                         |
+--------+-----------------------------------------+
```

- **Desktop:** Topbar always visible (sticky top-0), sidebar слева
- **Mobile (< 768px):** Bottom navigation bar + overlay sidebar
- **Sidebar collapsed:** 64px (иконки only)
- **Sidebar expanded:** 240px

### Sidebar

Ключевые визуальные элементы:
- **Background:** `var(--base)` с subtle gradient
- **Gradient edge line:** `::after` pseudo-element на правом краю, vertical gradient с акцентным цветом
- **Active item glow:** Box-shadow с `var(--accent-glow)` + left border accent
- **Header:** Logo + workspace name
- **User section:** Avatar + имя + статус
- **Projects section:** Список проектов с иконками
- **Collapse/expand:** Smooth transition 200ms

### Topbar

Ключевые визуальные элементы:
- **Background:** `var(--base)` с `backdrop-filter: blur(24px)` и semi-transparent opacity
- **Gradient bottom edge:** `::after` pseudo-element, horizontal gradient
- **Breadcrumbs:** Текущая локация
- **Search trigger:** Ctrl+K shortcut indicator
- **User menu:** Avatar dropdown
- **Notifications bell:** С badge counter

### Mobile Navigation

- **Bottom nav bar:** 5 основных разделов (Dashboard, Tracker, Messenger, TMS, Calls)
- **Overlay sidebar:** Slide-in с backdrop overlay при свайпе/tap на hamburger
- **Breakpoints:** `md:hidden` для bottom nav, `md:flex` для desktop sidebar

---

## UI Примитивы (10 Shadcn компонентов)

### Button

| Variant | Стиль |
|---------|-------|
| Primary | Gradient background (`accent` -> `accent-hover`), white text, glow shadow on hover |
| Secondary | `var(--raised)` bg, `var(--t2)` text, border `var(--b2)` |
| Ghost | Transparent bg, `var(--t2)` text, `var(--raised)` bg on hover |
| Destructive | Red gradient, white text |

**Hover:** `translateY(-1px)`, shadow bloom, spring easing (250ms `var(--spring)`)

### Card

- **Background:** `bg-gradient-to-b from-card to-raised`
- **Top-edge highlight:** `::before` pseudo-element, 1px horizontal gradient line
- **Border:** `var(--b1)`, elevates to `var(--b2)` on hover
- **Shadow:** `var(--sh2)`, elevates to `var(--sh3)` on hover
- **Hover:** `translateY(-2px)`, border + shadow elevation, spring easing

### Input

- **Background:** `var(--input)`
- **Border:** `var(--b2)`
- **Focus:** Blue glow ring (`box-shadow: 0 0 0 3px var(--accent-glow)`), border becomes `var(--accent)`
- **Placeholder:** `var(--t4)`

### Badge

- **Status-specific glow:**
  - Success: green glow bg (`var(--gG)`)
  - Error: red glow bg (`var(--gR)`)
  - Warning: amber glow bg (`var(--gA)`)
  - Info: blue glow bg (`var(--gB)`)
- **Border radius:** `var(--pill)` (9999px)

### Avatar

- **Border:** 2px solid `var(--b2)`
- **Online indicator:** Green dot с box-shadow glow
- **Sizes:** sm (24px), default (32px), lg (40px), xl (48px)

### Dialog / Popover

- **Backdrop:** `backdrop-filter: blur(8px)`, semi-transparent overlay
- **Entrance animation:** Scale from 0.95 + fade, spring easing
- **Background:** `var(--card)`
- **Border:** `var(--b1)`
- **Shadow:** `var(--sh4)`

### DropdownMenu

- **Background:** `var(--raised)`
- **Border:** `var(--b1)`
- **Shadow:** `var(--sh3)`
- **Item hover:** `var(--card)` bg
- **Separator:** `var(--b1)`

### Tabs

- **Active indicator:** Bottom border с `var(--accent)`, 2px
- **Active text:** `var(--t1)`
- **Inactive text:** `var(--t3)`
- **Hover:** `var(--t2)`

### Tooltip

- **Background:** `var(--card)`
- **Border:** `var(--b1)`
- **Shadow:** `var(--sh2)`
- **Text:** `var(--t2)`
- **Border radius:** `var(--r1)` (6px)

### ScrollArea

- **Scrollbar width:** 5px (slim)
- **Scrollbar track:** Transparent
- **Scrollbar thumb:** `var(--t4)` с `var(--t3)` on hover
- **Firefox:** `scrollbar-width: thin`, `scrollbar-color`

---

## Страницы

### 11 страниц с макетами

#### Login

- **Layout:** Split -- левая панель (animated gradient) + правая панель (форма)
- **Левая панель:** Mesh gradient background (20s hue-rotate loop), floating feature cards (стaggered 7-9s infinite animation)
- **Правая панель:** Email input -> verification code flow
- **Background:** Grid-lines pattern overlay
- **Анимации:** Spring entrance (400ms stagger), floating cards continuous

#### Dashboard

- **Layout:** 3-column responsive grid
- **Greeting section:** Gradient text (accent color), дата, время
- **Quick actions:** 4 кнопки с иконками (new task, new message, start call, run tests)
- **Widgets:** Cards с drag-and-drop (Tasks, Messages, Calls, Test Runs, Deadlines, Activity)
- **Анимации:** Staggered entrance (100ms delay per widget)

#### Messenger

- **Layout:** 3-panel (sidebar 280px + channel list + chat area)
- **Channel list:** Search bar, filter buttons, channel items с unread badges
- **Chat area:** Message list, glassmorphic composer (backdrop-filter: blur)
- **Message items:** Avatar, username, timestamp, text, reactions bar
- **Typing indicator:** Animated dots
- **Reply bar:** Quote preview с close button

#### Kanban

- **Columns:** Header с status color dot, task count, add button
- **Issue cards:** Priority color dot, title, assignee avatar, labels, glow hover
- **Filter bar:** Status, priority, assignee, type, search filters
- **Drag-and-drop:** pragmatic-drag-and-drop integration preserved

#### Issue Detail

- **Layout:** Two-column (content 65% + sidebar meta 35%)
- **Content:** Title, description (Plate editor), tabs (Comments, Activity, Time Log)
- **Sidebar:** Status, priority, assignee, reporter, labels, due date, time estimate, related issues
- **Comments:** Avatar, author, timestamp, rich-text body, reactions

#### Calls

- **Participant grid:** Responsive grid с video/audio cards
- **Controls bar:** Mute, camera, screen share, record, hand raise, end call
- **Status indicators:** Speaking (green ring), muted (red icon), hand raised (yellow)

#### TMS

- **Test case grid:** Folder tree (left) + case list (right)
- **Case card:** Title, status badge, priority, steps count, linked issues
- **Execution tracking:** Progress bar, pass/fail/skip counts

#### Settings

- **Layout:** Sidebar nav (left 240px) + content (right)
- **Sections:** Profile, Appearance, Notifications, Security, Language
- **Appearance:** Theme selector (dark/light/system), accent color picker (5 colors)

#### Admin

- **KPI cards:** 4 cards top row (Users, Projects, Tasks, DAU/WAU/MAU)
- **Tabbed interface:** Users, Audit Log, Resource Limits
- **User table:** Sortable columns, role badges, status indicators
- **Audit log:** Filterable timeline

#### Onboarding

- **Layout:** Centered modal overlay
- **Step 1:** Profile form (name, avatar, role)
- **Step 2:** Feature tour (6 cards с иконками и описаниями)
- **Step 3:** Theme + accent selection

#### Trash

- **Table:** Item type badge, name, deleted date, deleted by
- **Actions:** Restore button (green), permanent delete button (red)
- **Empty state:** Иллюстрация + текст

### 6 страниц без макетов

| Страница | Паттерн-источник | Реализация |
|----------|-----------------|------------|
| Projects List | Dashboard widget grid | Card grid с project avatars, member counts, activity |
| Project Overview | Dashboard | Метрики проекта в widget cards |
| Sprint Management | Kanban + Dashboard | Backlog list, sprint board, burndown/velocity charts |
| JQL Search | Issue list + Settings | Search input + results table + saved filters |
| Project Settings | Settings | Sidebar nav (Workflow, Labels, Integrations, Members) + content |
| Project Dashboard | Dashboard | Project-specific widget grid |

---

## Анимации

### Типы и easing

| Тип | Easing | Duration | Применение |
|-----|--------|----------|-----------|
| Hover | `var(--spring)` | 250ms | Button, card, link hover states |
| Focus | `var(--spring)` | 200ms | Input focus ring, tab selection |
| Entrance | `var(--spring)` | 400ms | Page content, modal, dropdown |
| Stagger | `var(--spring)` | 100ms delay | Widget grids, list items |
| Continuous | `linear` | 7-20s | Login mesh gradient, floating cards |
| Transition | `var(--ease)` | 200ms | Theme toggle, sidebar collapse |

### Easing values

```css
:root {
  --ease: cubic-bezier(0.4, 0, 0.2, 1);
  --spring: cubic-bezier(0.34, 1.56, 0.64, 1);
}
```

### Compositor-safe properties

Все анимации используют только compositor-safe свойства для 60fps:
- `transform` (translate, scale, rotate)
- `opacity`
- `box-shadow` (с `will-change` при необходимости)
- `filter` (backdrop-filter, blur)

**Запрещено анимировать:**
- `width`, `height` (layout trigger)
- `top`, `left`, `right`, `bottom` (layout trigger)
- `padding`, `margin` (layout trigger)

### prefers-reduced-motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## Производительность

### CSS правила

| Правило | Причина |
|---------|---------|
| Нет `transition: all` | Анимирует все свойства, включая layout-triggering; используем конкретные свойства |
| Нет `!important` | 56 деклараций удалены в Phase 8; specificity решается через правильные селекторы |
| Dead CSS removed | Неиспользуемые классы и утилиты удалены в Phase 13 |
| Compositor-safe only | Только transform/opacity/box-shadow/filter для анимаций |
| `will-change` sparingly | Только на элементах с continuous animations (login mesh gradient) |

### Font loading

- `font-display: swap` -- текст видим сразу с fallback, шрифт подгружается асинхронно
- Google Fonts CDN с preconnect -- минимизация latency
- 2 шрифта, 9 весов total -- приемлемый размер для internal tool

### Bundle impact

- **Новые npm-пакеты:** 0
- **CSS size change:** Минимальный (замена значений переменных, добавление новых утилит)
- **JS size change:** Незначительный (обновлённые className в компонентах)

---

## Адаптив (Responsive)

### Breakpoints

| Breakpoint | Значение | Применение |
|-----------|----------|-----------|
| `sm` | 640px | Mobile landscape |
| `md` | 768px | Tablet / sidebar threshold |
| `lg` | 1024px | Desktop |
| `xl` | 1280px | Wide desktop |
| `2xl` | 1536px | Ultra-wide |

### Mobile-first grids

```css
/* Dashboard widget grid */
grid-cols-1          /* mobile: 1 column */
sm:grid-cols-2       /* tablet: 2 columns */
lg:grid-cols-3       /* desktop: 3 columns */

/* Messenger layout */
/* mobile: full-width channel list OR chat (toggled) */
/* md: 280px channel list + chat */

/* Kanban columns */
/* mobile: horizontal scroll */
/* md: flex-nowrap with min-width per column */
```

### Критические viewport sizes

| Viewport | Проверяемые элементы |
|----------|---------------------|
| 375px (iPhone SE) | Все страницы, sidebar overlay, bottom nav, single column layouts |
| 768px (iPad) | Sidebar collapse/expand threshold, 2-column grids |
| 1024px (Desktop) | Full 3-column layout, expanded sidebar |

### Sidebar responsive behavior

```
< 768px:  Sidebar hidden, bottom nav visible, overlay sidebar on demand
>= 768px: Sidebar visible (collapsible), topbar visible, no bottom nav
```

---

## Решения и обоснования

### CSS Variables vs Tailwind Config

| Критерий | CSS Variables `:root`/`.dark` | tailwind.config.js |
|----------|-------------------------------|-------------------|
| Hot-reload | Мгновенный (CSS cascade) | Требует rebuild |
| Theme switching | Runtime (class toggle) | Compile-time |
| Dynamic values | Поддержка | Статичные |

**Выбрано:** CSS Variables через `@theme inline` -- стандарт Tailwind v4, поддерживает runtime theme switching.

### Google Fonts CDN vs Self-hosted WOFF2

| Критерий | CDN | Self-hosted |
|----------|-----|------------|
| Cache | Cross-site caching | Per-site only |
| Build | Zero config | Requires font files in assets |
| CLS prevention | `font-display: swap` | `size-adjust` fallback |
| Offline | Requires network | Works offline |

**Выбрано:** CDN -- Hub внутренний инструмент, всегда online, CDN проще и уже используется в проекте.

### motion v12 vs CSS-only animations

| Критерий | motion v12 | CSS-only |
|----------|-----------|----------|
| Layout animations | Excellent | Impossible |
| Mount/unmount | AnimatePresence | display: none hack |
| Hover/focus/press | Overkill | Native, zero JS |
| Bundle cost | ~25KB | 0 |

**Выбрано:** Гибрид -- CSS transitions для 95% интерактивности (hover, focus, press), motion v12 только для layout animations и mount/unmount (page entrance, staggered lists).

### !important removal strategy

56 `!important` деклараций были в `app.css`. Стратегия удаления:
1. Аудит каждой декларации -- определение, какие компоненты она защищает
2. Перенос Radix-specific overrides в индивидуальные Shadcn component files через `cn()`
3. Увеличение специфичности через правильные селекторы вместо `!important`
4. Верификация каждого удаления -- visual regression check

---

## Известные ограничения

### Светлая тема -- extrapolated values

Светлая тема не имеет HTML-макетов. Все значения derived из тёмной палитры с инверсией. Возможны отклонения от "идеального" дизайна, которые требуют ручной корректировки при получении макетов светлой темы.

**Митигация:** WCAG AA контрастность верифицирована; визуальная приемлемость проверена на всех 17 страницах.

### Unmocked pages -- pattern extrapolation

6 страниц без макетов стилизованы по паттернам из макетных страниц. Дизайн-решения могут не совпадать с видением дизайнера.

**Митигация:** Каждая unmocked страница использует точно те же компоненты и токены, что и ближайший аналог с макетом.

### Plate editor style isolation

Rich-text editor Plate v49 рендерит собственный DOM с Slate-специфичными классами. После удаления глобальных `!important` overrides, стили Plate могут конфликтовать с новой палитрой.

**Митигация:** CSS scoping через `.plate-editor` container class; отдельные переопределения для Slate elements.

### LiveKit component styling

LiveKit SDK рендерит собственные UI-компоненты для видеозвонков. Глобальные изменения Button/Avatar могут "просачиваться" в LiveKit DOM.

**Митигация:** CSS scoping, тестирование UI звонков после каждого глобального изменения стилей.

### Font loading CLS

Google Fonts CDN с `font-display: swap` может вызывать кратковременный CLS при первой загрузке (fallback font -> Inter transition).

**Митигация:** `font-display: swap` минимизирует FOIT; для internal tool с постоянными пользователями шрифт кешируется после первого визита.

---

*Документация составлена на основе реализации Milestone v2.0 проекта Hub. Последнее обновление: 2026-03-28.*
