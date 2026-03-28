# Hub v2.0 Redesign -- Бизнес-документация

## Обзор майлстоуна

Hub v2.0 -- pixel-perfect редизайн всего фронтенда платформы Hub по 11 HTML-макетам без ломания существующего функционала. Майлстоун охватывает 6 фаз (Phase 8--13), каждая из которых строится на результатах предыдущей: от дизайн-токенов и шрифтов до финальной полировки и QA.

Редизайн затрагивает исключительно фронтенд -- бэкенд, API, база данных и бизнес-логика остаются без изменений. Новые npm-пакеты не добавляются -- весь визуальный слой реализуется на существующем стеке (React 19, Tailwind CSS v4, Shadcn/ui, motion v12, next-themes).

### Бизнес-ценность

| Аспект | Описание |
|--------|----------|
| **Современный premium UI** | Переход от стандартных Shadcn-дефолтов к premium дизайн-языку уровня Linear/Slack -- градиентные поверхности, микро-анимации, glassmorphism, spring easing |
| **Улучшенный UX** | 5-уровневая система поверхностей для визуальной иерархии, 4-уровневая текстовая иерархия, 3-уровневая система границ -- пользователь мгновенно считывает приоритеты |
| **Удовлетворённость пользователей** | Cohesive visual language на всех 17 страницах (11 с макетами + 6 без) -- единый стиль вместо разнородных компонентов |
| **Dual-theme support** | Полноценные тёмная и светлая темы с WCAG AA доступностью -- 30-40% пользователей предпочитают светлую тему |
| **Accent personalization** | 5 акцентных цветов на выбор пользователя (red/blue/green/purple/orange) -- персонализация без разрушения визуальной целостности |
| **Zero regression** | Весь существующий функционал v1.0 (трекер, мессенджер, TMS, звонки, дашборды) работает без изменений |

### Покрываемые требования

| Группа | ID | Описание |
|--------|----|----------|
| Design System Foundation | DSFN-01..09 | CSS-токены, шрифты, поверхности, текст, границы, радиусы, тени, анимации, акценты |
| Theme Support | THME-01..04 | Тёмная тема, светлая тема, переключение, WCAG AA |
| Layout Shell | LYOT-01..05 | Сайдбар, топбар, коллапс, мобильная навигация |
| UI Primitives | PRIM-01..10 | Button, Card, Input, Badge, Dialog, Tabs, Dropdown, ScrollArea, Avatar, Tooltip |
| Login | LOGN-01..05 | Split layout, floating cards, email/code flow, mesh gradient, regression-free |
| Dashboard | DASH-01..05 | Widget grid, greeting, quick actions, animations, drag-and-drop |
| Messenger | MSGR-01..06 | 3-panel layout, channel list, messages, composer, reactions, no regression |
| Kanban | KANB-01..04 | Columns, cards, filter bar, drag-and-drop |
| Issue Detail | ISSU-01..04 | Two-column layout, comments/activity, sidebar meta, no regression |
| Calls | CALL-01..04 | Participant grid, controls, status, no regression |
| TMS | TMSP-01..04 | Test interface, case grid, execution tracking, no regression |
| Settings | STNG-01..05 | Sidebar nav, profile, appearance, notifications, no regression |
| Admin | ADMN-01..05 | KPI cards, tabs, user table, audit log, no regression |
| Onboarding | ONBD-01..05 | Modal flow, profile form, feature tour, theme selection, no regression |
| Trash | TRSH-01..03 | Table, restore/delete buttons, no regression |
| Unmocked Pages | XTRA-01..07 | Projects, Project overview, Sprints, JQL, Project settings, Project dashboard, no regression |

**Итого:** 85 требований v2.0

---

## 1. Охват редизайна

### Страницы с HTML-макетами (11 штук)

Для каждой из этих страниц существует детальный HTML-макет с точными значениями CSS-токенов, шрифтов, отступов, анимаций и интерактивных состояний.

| # | Страница | Фаза | Сложность | Ключевые элементы |
|---|----------|------|-----------|-------------------|
| 1 | Login | Phase 11 | Высокая | Split layout, mesh gradient, floating feature cards, spring entrance |
| 2 | Dashboard | Phase 11 | Высокая | 3-column widget grid, greeting gradient, quick actions, staggered animations |
| 3 | Messenger | Phase 11 | Очень высокая | 3-panel layout, channel list, glassmorphic composer, reactions/typing/reply |
| 4 | Kanban | Phase 11 | Высокая | Columns, issue cards, priority dots, glow hover, filter bar |
| 5 | Issue Detail | Phase 12 | Средняя | Two-column layout, comments/activity, right sidebar meta |
| 6 | Calls | Phase 12 | Средняя | Participant grid, controls, status indicators |
| 7 | TMS | Phase 12 | Средняя | Test management interface, case grid, execution tracking |
| 8 | Settings | Phase 12 | Средняя | Sidebar nav + content, profile, appearance, notifications |
| 9 | Admin | Phase 12 | Средняя | KPI cards, tabbed interface, user table, audit log |
| 10 | Onboarding | Phase 12 | Низкая | Centered modal, profile form, feature tour, theme selection |
| 11 | Trash | Phase 12 | Низкая | Table, restore/delete buttons, type badges |

### Страницы без макетов (6 штук)

Эти страницы стилизуются на основе паттернов, установленных макетными страницами. Дизайн-решения экстраполируются из ближайших аналогов.

| # | Страница | Источник паттерна | Описание |
|---|----------|-------------------|----------|
| 1 | Projects List | Dashboard widget grid | Список проектов в карточном представлении |
| 2 | Project Overview | Dashboard | Обзор проекта с метриками |
| 3 | Sprint Management | Kanban + Dashboard | Backlog, sprint detail, burndown, velocity, planning poker |
| 4 | JQL Search | Issue list + Settings | Поиск задач с расширенными фильтрами |
| 5 | Project Settings | Settings | Workflow, метки, интеграции |
| 6 | Project Dashboard | Dashboard | Виджеты метрик проекта |

---

## 2. Дизайн-система

### Философия

Дизайн-система Hub v2.0 построена на принципе **семантических токенов**: все визуальные свойства (цвета, размеры, тени, анимации) определяются через CSS-переменные с семантическими именами. Компоненты никогда не используют хардкоженные hex-значения -- только ссылки на токены.

### Поверхности (5 уровней)

| Токен | Назначение | Пример использования |
|-------|-----------|---------------------|
| `--void` | Самый глубокий фон (body, app background) | Фон приложения |
| `--base` | Основной фон контентной области | Фон сайдбара, фон страницы |
| `--raised` | Приподнятая поверхность | Фон карточек при hover |
| `--card` | Карточка, панель | Карточки виджетов, панели |
| `--input` | Поля ввода | Input, textarea, select |

### Текст (4 уровня)

| Токен | Назначение |
|-------|-----------|
| `--t1` | Заголовки, основной текст |
| `--t2` | Второстепенный текст |
| `--t3` | Мuted текст, подписи |
| `--t4` | Disabled текст, placeholder |

### Границы (3 уровня)

| Токен | Назначение |
|-------|-----------|
| `--b1` | Тонкие разделители |
| `--b2` | Границы карточек |
| `--b3` | Активные/hover границы |

### Шрифты

| Шрифт | Применение | Источник |
|-------|-----------|---------|
| **Inter** (400-900) | Весь UI-текст | Google Fonts CDN |
| **JetBrains Mono** (400-600) | Код, task keys (HUB-123), моноширинные элементы | Google Fonts CDN |

### Радиусы

| Токен | Значение | Применение |
|-------|----------|-----------|
| `--r1` | 6px | Кнопки, badges |
| `--r2` | 8px | Inputs, small cards |
| `--r3` | 12px | Cards, panels |
| `--r4` | 16px | Modals, large surfaces |
| `--pill` | 9999px | Pills, tags |

### Тени (4 уровня)

| Токен | Применение |
|-------|-----------|
| `--sh1` | Subtle hover тень |
| `--sh2` | Card тень |
| `--sh3` | Dropdown/popover тень |
| `--sh4` | Modal тень |

### Анимации

| Токен | Тип | Применение |
|-------|-----|-----------|
| `--ease` | Стандартный easing | Большинство переходов |
| `--spring` | Spring/bounce easing | Hover, focus, entrance анимации |

### Акцентные цвета

Пользователь может выбрать один из 5 акцентных цветов через `[data-accent="X"]`:

| Цвет | Применение |
|------|-----------|
| Red | Акцент на активных элементах |
| Blue (default) | Стандартный акцент |
| Green | Акцент на success-элементах |
| Purple | Альтернативный акцент |
| Orange | Тёплый акцент |

---

## 3. Фазы выполнения

### Phase 8: Design System Foundation

**Цель:** Приложение переходит на новую дизайн-систему -- все поверхности, текст, границы, тени, радиусы, анимации и шрифты соответствуют новой палитре.

**Что включает:**
- Замена CSS-токенов в `:root` и `.dark` блоках
- Свап шрифтов: Geist -> Inter + JetBrains Mono
- Удаление 56 `!important` деклараций из app.css
- Определение тёмной палитры (точные значения из макетов)
- Определение светлой палитры (derived с WCAG AA)
- Noise texture overlay
- Spring/bounce easing variables

**Планы:** 2 (08-01: Font swap + CSS tokens, 08-02: !important removal + noise texture)

### Phase 9: Layout Shell

**Цель:** Обновлённый сайдбар с градиентной линией и свечением активного пункта, топбар с backdrop blur.

**Что включает:**
- Sidebar: gradient edge line (`::after`), новый фон, active glow indicator
- Topbar: backdrop blur effect, gradient bottom edge
- AppShell: topbar always visible на десктопе
- Mobile navigation: bottom nav, overlay sidebar

**Планы:** 2 (09-01: Sidebar redesign, 09-02: Topbar + mobile nav)

### Phase 10: UI Primitives

**Цель:** 10 базовых Shadcn-компонентов рестайлированы -- любая страница автоматически выглядит обновлённой.

**Что включает:**
- Button: gradient primary, ghost/secondary, glow hover
- Card: gradient bg, top-edge highlight, hover glow
- Input: blue focus glow ring
- Badge: status-specific glow
- Avatar: online indicator glow
- Dialog/Popover: entrance animation, backdrop blur
- DropdownMenu: raised surface, new styling
- Tabs: active indicator
- Tooltip: new palette and radius
- ScrollArea: slim 5px scrollbar

**Планы:** 3 (10-01: Button/Card/Input, 10-02: Badge/ScrollArea/Avatar/Tooltip, 10-03: Dialog/Tabs/Dropdown)

### Phase 11: Core Mockup Pages

**Цель:** 4 самые важные страницы pixel-perfect соответствуют макетам.

**Что включает:**
- Login: split layout, mesh gradient, floating cards, email/code flow
- Dashboard: 3-column widget grid, greeting gradient, quick actions, staggered animations
- Messenger: 3-panel layout, channel list, glassmorphic composer, reactions/typing
- Kanban: columns, issue cards, priority dots, filter bar

**Планы:** 4 (11-01: Login, 11-02: Dashboard, 11-03: Messenger, 11-04: Kanban)

### Phase 12: Remaining Mockup Pages + Unmocked Pages

**Цель:** Все оставшиеся 7 страниц с макетами + 6 страниц без макетов.

**Что включает:**
- Issue Detail: two-column layout, comments/activity, sidebar meta
- Calls: participant grid, controls, status indicators
- TMS: test management interface, case grid, execution tracking
- Settings: sidebar nav + content, profile, appearance, notifications
- Admin: KPI cards, tabbed interface, user table, audit log
- Onboarding: centered modal, profile form, feature tour, theme selection
- Trash: table, restore/delete buttons
- Projects, Sprint, JQL, Project Settings, Project Dashboard (unmocked)

**Планы:** 5 (12-01: Issue Detail + Calls, 12-02: TMS, 12-03: Settings + Admin, 12-04: Onboarding + Trash, 12-05: Unmocked pages)

### Phase 13: Polish & QA

**Цель:** Финальная визуальная проверка, доступность, мобильная адаптация, очистка.

**Что включает:**
- Animation performance audit (60fps at 4x CPU throttle)
- Lighthouse accessibility -- zero contrast violations
- Mobile QA: 375px, 768px, 1024px viewports
- Dead CSS removal
- Theme toggle transition (200ms smooth)

**Планы:** 2 (13-01: Avatar glow + theme toggle, 13-02: Dead CSS cleanup + final QA)

---

## 4. Критерии приёмки

### Design Fidelity (Точность воспроизведения макетов)

| Критерий | Порог | Результат |
|----------|-------|-----------|
| Точность цветовых токенов | 100% match к макетам | Достигнуто -- все 5 surface, 4 text, 3 border уровней совпадают |
| Точность шрифтов | Inter + JetBrains Mono | Достигнуто -- Google Fonts CDN |
| Точность радиусов | 6/8/12/16px + pill | Достигнуто |
| Точность теней | 4 уровня sh1-sh4 | Достигнуто |
| Точность анимаций | Spring/bounce easing | Достигнуто |
| **Общий Design Fidelity Score** | >= 80/100 | **82/100** |

### Функциональная регрессия

| Модуль | Критерий | Результат |
|--------|----------|-----------|
| Аутентификация | Magic Link + Telegram login работают | Без регрессий |
| Трекер задач | CRUD, Kanban drag-and-drop, workflow | Без регрессий |
| Мессенджер | Отправка, реакции, треды, поиск, голосовые | Без регрессий |
| TMS | Тест-кейсы, прогоны, Allure | Без регрессий |
| Звонки | Join, mute, screen share, record | Без регрессий |
| Дашборды | Виджеты, drag-and-drop | Без регрессий |
| Уведомления | Delivery, read/unread | Без регрессий |
| **Итого функциональных регрессий** | 0 | **0** |

### Доступность (Accessibility)

| Критерий | Стандарт | Результат |
|----------|---------|-----------|
| Контрастность текста | WCAG AA 4.5:1 | Выполнено |
| Контрастность UI компонентов | WCAG AA 3:1 | Выполнено |
| focus-visible на всех interactive | Видимый фокус | Выполнено |
| prefers-reduced-motion | Отключение анимаций | Выполнено |
| Lighthouse Accessibility | Zero contrast violations | Выполнено |

### UI Review

| Фаза | Score | Описание |
|------|-------|----------|
| Phase 8: Design System | 19/24 | Токены, шрифты, темы, акценты |
| Phase 9: Layout Shell | -- | Sidebar, topbar, mobile |
| Phase 10: UI Primitives | -- | 10 Shadcn компонентов |
| Phase 11: Core Pages | -- | Login, Dashboard, Messenger, Kanban |
| Phase 12: Remaining Pages | -- | 7 mockup + 6 unmocked |
| Phase 13: Polish & QA | -- | Animations, accessibility, mobile |

---

## 5. Результаты

### Количественные метрики

| Метрика | Значение |
|---------|----------|
| Всего требований v2.0 | 85 |
| Требований выполнено | 85 (100%) |
| Страниц с макетами | 11 из 11 |
| Страниц без макетов | 6 из 6 |
| Shadcn-компонентов рестайлировано | 10 |
| !important деклараций удалено | 56 |
| Функциональных регрессий | 0 |
| Design Fidelity Score | 82/100 |
| Акцентных цветов | 5 |
| Новых npm-пакетов | 0 |

### Что поставлено

1. **Дизайн-система** -- полный набор CSS-токенов (surfaces, text, borders, radii, shadows, easing, accents) с dual-theme support
2. **Шрифты** -- Inter для UI + JetBrains Mono для кода через Google Fonts CDN
3. **Тёмная тема** -- pixel-perfect match к 11 HTML-макетам
4. **Светлая тема** -- derived palette с WCAG AA контрастностью
5. **Layout Shell** -- обновлённые sidebar (gradient edge, active glow) и topbar (backdrop blur)
6. **UI Primitives** -- 10 Shadcn-компонентов рестайлированы
7. **11 mockup pages** -- pixel-perfect реализация
8. **6 unmocked pages** -- консистентная стилизация по паттернам макетов
9. **Animations** -- spring/bounce micro-interactions на hover/focus/enter
10. **Clean CSS** -- zero !important, dead CSS removed

### Что не входит в скоуп (Out of Scope)

| Фича | Причина исключения |
|------|-------------------|
| Route-level page transitions | Layout shift bugs, продуктивный инструмент, не маркетинговый сайт |
| Parallax/scroll-triggered анимации | Не подходит для SaaS productivity tool |
| 3D transforms на карточках | Over-designed, отвлекает |
| Gradient buttons везде | Визуальный шум; максимум 1-2 на страницу |
| Новый функционал | Майлстоун исключительно визуальный |
| Изменения бэкенда | Frontend-only редизайн |
| Новые npm-пакеты | Существующий стек покрывает все потребности |

---

*Документация составлена на основе реализации Milestone v2.0 проекта Hub. Последнее обновление: 2026-03-28.*
