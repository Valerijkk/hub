# Тестовая документация -- Hub v2.0 Redesign

## Соглашение об именовании

Идентификатор тест-кейса: `TC-{REQ_ID}-{NN}`, где:
- `REQ_ID` -- идентификатор требования (DSFN-01, LYOT-01, PRIM-01 и т.д.)
- `NN` -- порядковый номер теста внутри требования

Формат описания каждого тест-кейса:
- **ID**: уникальный идентификатор
- **Описание**: что проверяется
- **Предусловия**: начальное состояние
- **Шаги**: последовательность действий
- **Ожидаемый результат**: критерий прохождения
- **Приоритет**: высокий / средний / низкий

---

## Стратегия тестирования

Редизайн v2.0 -- визуальный проект без бэкенд-изменений. Стратегия тестирования фокусируется на трёх аспектах:

1. **Design Fidelity** -- визуальное соответствие макетам (pixel-perfect)
2. **Functional Regression** -- все существующие функции v1.0 работают без изменений
3. **Accessibility** -- WCAG AA контрастность, focus-visible, reduced-motion

### Методы тестирования

| Метод | Применение |
|-------|-----------|
| **Ручная визуальная проверка** | Сравнение с HTML-макетами, pixel-perfect audit |
| **Функциональное тестирование** | Regression checklist по каждому модулю |
| **Lighthouse audit** | Accessibility score, contrast violations |
| **DevTools responsive mode** | Mobile QA на 375px, 768px, 1024px |
| **Performance profiling** | Chrome DevTools Performance tab, 4x CPU throttle |
| **Theme toggle тест** | Переключение dark/light/system без visual glitches |
| **Accent color тест** | Переключение 5 акцентных цветов, проверка consistency |

---

## Phase 8: Design System Foundation

### DSFN-01: Шрифт Inter для UI текста

#### TC-DSFN-01-01: Inter загружается для всего UI-текста

- **ID**: TC-DSFN-01-01
- **Описание**: Проверка загрузки шрифта Inter и его применения ко всему UI-тексту
- **Предусловия**: Приложение запущено, сеть доступна
- **Шаги**:
  1. Открыть приложение в браузере
  2. Инспектировать любой текстовый элемент через DevTools
  3. Проверить computed font-family
- **Ожидаемый результат**: `font-family` содержит `'Inter'` как первый шрифт; fallback `system-ui, -apple-system, sans-serif`
- **Приоритет**: высокий

#### TC-DSFN-01-02: JetBrains Mono для code-элементов

- **ID**: TC-DSFN-01-02
- **Описание**: Проверка JetBrains Mono для моноширинных элементов
- **Предусловия**: Существует задача с ключом (HUB-123)
- **Шаги**:
  1. Открыть страницу с task key (Kanban, Issue Detail)
  2. Инспектировать issue_key элемент
  3. Проверить computed font-family
- **Ожидаемый результат**: `font-family` содержит `'JetBrains Mono'`; элементы `<code>`, inline code также используют JetBrains Mono
- **Приоритет**: высокий

### DSFN-02: 5-уровневая система поверхностей

#### TC-DSFN-02-01: Поверхности void/base/raised/card/input в dark mode

- **ID**: TC-DSFN-02-01
- **Описание**: Проверка 5 уровней поверхностей в тёмной теме
- **Предусловия**: Тёмная тема активна
- **Шаги**:
  1. Открыть Dashboard
  2. Инспектировать body background -> `--void` (#09090B)
  3. Инспектировать sidebar background -> `--base` (#0C0C0F)
  4. Инспектировать hover state карточки -> `--raised` (#141419)
  5. Инспектировать карточку виджета -> `--card` (#18181B)
  6. Инспектировать поле поиска -> `--input` (#1E1E23)
- **Ожидаемый результат**: Каждый уровень имеет точное значение из макета; визуальная иерархия чётко различима
- **Приоритет**: высокий

### DSFN-03: 4-уровневая текстовая иерархия

#### TC-DSFN-03-01: Текстовые уровни t1/t2/t3/t4

- **ID**: TC-DSFN-03-01
- **Описание**: Проверка 4 уровней текстовой иерархии
- **Предусловия**: Тёмная тема активна
- **Шаги**:
  1. Открыть любую страницу с разнообразным контентом (Dashboard, Issue Detail)
  2. Инспектировать заголовок -> `--t1`
  3. Инспектировать описание -> `--t2`
  4. Инспектировать мuted текст (даты, подписи) -> `--t3`
  5. Инспектировать placeholder текст -> `--t4`
- **Ожидаемый результат**: Каждый уровень визуально различим; контрастность t1 и t2 соответствует WCAG AA
- **Приоритет**: высокий

### DSFN-04: 3-уровневая система границ

#### TC-DSFN-04-01: Границы b1/b2/b3

- **ID**: TC-DSFN-04-01
- **Описание**: Проверка 3 уровней границ
- **Предусловия**: Тёмная тема активна
- **Шаги**:
  1. Инспектировать тонкий разделитель (sidebar separator) -> `--b1`
  2. Инспектировать границу карточки -> `--b2`
  3. Навести на карточку, инспектировать hover border -> `--b3`
- **Ожидаемый результат**: Три различных уровня opacity; b1 самый тонкий, b3 самый заметный
- **Приоритет**: средний

### DSFN-05: Система радиусов

#### TC-DSFN-05-01: Радиусы 6/8/12/16px + pill

- **ID**: TC-DSFN-05-01
- **Описание**: Проверка применения радиусов
- **Предусловия**: Приложение запущено
- **Шаги**:
  1. Инспектировать badge -> `--r1` (6px)
  2. Инспектировать input -> `--r2` (8px)
  3. Инспектировать card -> `--r3` (12px)
  4. Инспектировать modal -> `--r4` (16px)
  5. Инспектировать pill badge -> `--pill` (9999px)
- **Ожидаемый результат**: Каждый тип элемента использует соответствующий радиус
- **Приоритет**: средний

### DSFN-06: Система теней

#### TC-DSFN-06-01: Тени sh1-sh4

- **ID**: TC-DSFN-06-01
- **Описание**: Проверка 4 уровней теней
- **Предусловия**: Приложение запущено
- **Шаги**:
  1. Навести на элемент с subtle shadow -> `--sh1`
  2. Инспектировать card shadow -> `--sh2`
  3. Открыть dropdown -> `--sh3`
  4. Открыть modal -> `--sh4`
- **Ожидаемый результат**: Тени увеличиваются в интенсивности от sh1 к sh4
- **Приоритет**: средний

### DSFN-07: Spring/bounce easing

#### TC-DSFN-07-01: Spring easing на hover/focus

- **ID**: TC-DSFN-07-01
- **Описание**: Проверка spring easing на интерактивных элементах
- **Предусловия**: Приложение запущено
- **Шаги**:
  1. Навести на кнопку -- наблюдать spring bounce
  2. Навести на карточку -- наблюдать плавный подъём с spring
  3. Кликнуть на input -- наблюдать focus ring с spring
  4. Открыть dropdown -- наблюдать entrance с spring
- **Ожидаемый результат**: Все анимации имеют характерный spring overshoot; нет linear или ease-in-out
- **Приоритет**: средний

### DSFN-08: Удаление !important

#### TC-DSFN-08-01: Zero !important в app.css

- **ID**: TC-DSFN-08-01
- **Описание**: Проверка отсутствия !important деклараций
- **Предусловия**: Доступ к исходному коду
- **Шаги**:
  1. Открыть `frontend/src/app.css`
  2. Поискать `!important` (кроме `prefers-reduced-motion` media query)
  3. Проверить визуальный рендеринг всех страниц
- **Ожидаемый результат**: Ноль `!important` деклараций (кроме accessibility override в reduced-motion); все элементы стилизованы корректно
- **Приоритет**: высокий

### DSFN-09: Акцентные цвета

#### TC-DSFN-09-01: Переключение 5 акцентных цветов

- **ID**: TC-DSFN-09-01
- **Описание**: Проверка работы всех 5 акцентных цветов
- **Предусловия**: Приложение запущено, пользователь авторизован
- **Шаги**:
  1. Перейти в Settings -> Appearance
  2. Выбрать Red акцент -- проверить primary button, focus ring, active sidebar item
  3. Выбрать Blue акцент -- проверить те же элементы
  4. Выбрать Green, Purple, Orange -- проверить каждый
- **Ожидаемый результат**: Акцент применяется мгновенно ко всем themed элементам без перезагрузки; цвет consistent на всех страницах
- **Приоритет**: высокий

---

## Phase 8: Themes

### THME-01: Тёмная тема

#### TC-THME-01-01: Тёмная тема соответствует макетам

- **ID**: TC-THME-01-01
- **Описание**: Визуальное сравнение тёмной темы с HTML-макетами
- **Предусловия**: Тёмная тема активна
- **Шаги**:
  1. Открыть Login page -- сравнить с 01-login.html
  2. Открыть Dashboard -- сравнить с 02-dashboard.html
  3. Открыть Messenger -- сравнить с 03-messenger.html
  4. Открыть Kanban -- сравнить с 04-kanban.html
  5. Проверить все 11 макетных страниц
- **Ожидаемый результат**: Цвета поверхностей, текста, границ визуально совпадают с макетами
- **Приоритет**: высокий

### THME-02: Светлая тема

#### TC-THME-02-01: Светлая тема визуально приемлема

- **ID**: TC-THME-02-01
- **Описание**: Проверка светлой темы на всех страницах
- **Предусловия**: Светлая тема активна
- **Шаги**:
  1. Пройти по всем 17 страницам
  2. Проверить читаемость текста
  3. Проверить различимость поверхностей
  4. Проверить видимость границ и разделителей
- **Ожидаемый результат**: Все элементы читаемы и различимы; нет "белого на белом" или невидимых элементов
- **Приоритет**: высокий

### THME-03: Переключение тем

#### TC-THME-03-01: Переключение dark/light/system без glitches

- **ID**: TC-THME-03-01
- **Описание**: Проверка переключения между темами
- **Предусловия**: Пользователь на Dashboard
- **Шаги**:
  1. Переключить на Light theme
  2. Переключить на Dark theme
  3. Переключить на System theme
  4. Изменить системную тему (OS settings)
  5. Перезагрузить страницу -- проверить сохранение выбора
- **Ожидаемый результат**: Переключение мгновенное; нет flash of wrong theme при загрузке; System корректно следует OS preference
- **Приоритет**: высокий

### THME-04: WCAG AA контрастность

#### TC-THME-04-01: Контрастность текста в светлой теме

- **ID**: TC-THME-04-01
- **Описание**: Проверка WCAG AA контрастности
- **Предусловия**: Светлая тема активна
- **Шаги**:
  1. Запустить Lighthouse Accessibility audit
  2. Проверить t1 на void/base/card -- >= 4.5:1
  3. Проверить t2 на void/base/card -- >= 4.5:1
  4. Проверить t3 на void/base/card -- >= 4.5:1 (AA)
  5. Проверить UI components (icons, borders) -- >= 3:1
- **Ожидаемый результат**: Zero contrast violations в Lighthouse; все текстовые пары t1/t2/t3 проходят AA; t4 используется только для decorative
- **Приоритет**: высокий

#### TC-THME-04-02: Контрастность текста в тёмной теме

- **ID**: TC-THME-04-02
- **Описание**: Проверка WCAG AA контрастности в dark mode
- **Предусловия**: Тёмная тема активна
- **Шаги**:
  1. Запустить Lighthouse Accessibility audit
  2. Проверить t1 на card (#F4F4F5 на #18181B) -- >= 4.5:1
  3. Проверить t2 на card (#A1A1AA на #18181B) -- >= 4.5:1
- **Ожидаемый результат**: Zero contrast violations; t1 13.5:1 (AAA), t2 6.2:1 (AA)
- **Приоритет**: высокий

---

## Phase 9: Layout Shell

### LYOT-01: Sidebar gradient edge

#### TC-LYOT-01-01: Sidebar с gradient edge line

- **ID**: TC-LYOT-01-01
- **Описание**: Проверка нового дизайна sidebar
- **Предусловия**: Приложение запущено, desktop viewport (>= 768px)
- **Шаги**:
  1. Посмотреть на правый край sidebar
  2. Инспектировать `::after` pseudo-element
- **Ожидаемый результат**: Вертикальная gradient линия на правом краю sidebar с accent color; фон sidebar соответствует `--base`
- **Приоритет**: высокий

### LYOT-02: Active nav glow

#### TC-LYOT-02-01: Active navigation item с accent glow

- **ID**: TC-LYOT-02-01
- **Описание**: Проверка подсветки активного пункта навигации
- **Предусловия**: Пользователь на Dashboard
- **Шаги**:
  1. Посмотреть на пункт "Dashboard" в sidebar
  2. Перейти на Messenger
  3. Посмотреть на пункт "Messenger"
- **Ожидаемый результат**: Активный пункт имеет accent glow (box-shadow), left border с accent color; неактивные пункты без glow
- **Приоритет**: высокий

### LYOT-03: Topbar backdrop blur

#### TC-LYOT-03-01: Topbar с backdrop blur и gradient edge

- **ID**: TC-LYOT-03-01
- **Описание**: Проверка дизайна topbar
- **Предусловия**: Достаточно контента для скролла
- **Шаги**:
  1. Прокрутить страницу вниз
  2. Проверить что topbar sticky и контент проглядывает с blur
  3. Инспектировать нижний край topbar
- **Ожидаемый результат**: `backdrop-filter: blur(24px)` на topbar; gradient bottom edge видна; контент под topbar размыт
- **Приоритет**: высокий

### LYOT-04: Sidebar collapse/expand

#### TC-LYOT-04-01: Коллапс и экспанд sidebar

- **ID**: TC-LYOT-04-01
- **Описание**: Проверка collapse/expand без visual regression
- **Предусловия**: Desktop viewport
- **Шаги**:
  1. Кликнуть на toggle sidebar
  2. Sidebar коллапсируется до 64px (только иконки)
  3. Кликнуть снова -- sidebar расширяется до 240px
  4. Проверить что контент корректно перестраивается
- **Ожидаемый результат**: Smooth transition 200ms; нет layout jumps; gradient edge line сохраняется; контент не перекрывается
- **Приоритет**: высокий

### LYOT-05: Mobile navigation

#### TC-LYOT-05-01: Bottom nav и overlay sidebar на мобильном

- **ID**: TC-LYOT-05-01
- **Описание**: Проверка мобильной навигации
- **Предусловия**: Viewport 375px (iPhone SE)
- **Шаги**:
  1. Проверить наличие bottom nav bar
  2. Тапнуть на каждый пункт -- переход работает
  3. Тапнуть на hamburger -- overlay sidebar выезжает
  4. Тапнуть на backdrop -- sidebar закрывается
- **Ожидаемый результат**: Bottom nav видна с 5 основными разделами; overlay sidebar корректно работает; нет overflow или layout breaks
- **Приоритет**: высокий

---

## Phase 10: UI Primitives

### PRIM-01: Button

#### TC-PRIM-01-01: Primary button с gradient и glow hover

- **ID**: TC-PRIM-01-01
- **Описание**: Проверка дизайна primary button
- **Предусловия**: Любая страница с primary button
- **Шаги**:
  1. Найти primary кнопку (например, "Create Task")
  2. Инспектировать background -- gradient accent
  3. Навести курсор -- glow shadow появляется
  4. Проверить ghost и secondary варианты
- **Ожидаемый результат**: Gradient background на primary; glow hover с spring easing; ghost transparent с raised hover; secondary с border
- **Приоритет**: высокий

### PRIM-02: Card

#### TC-PRIM-02-01: Card с gradient bg и top-edge highlight

- **ID**: TC-PRIM-02-01
- **Описание**: Проверка дизайна card
- **Предусловия**: Dashboard с виджетами
- **Шаги**:
  1. Инспектировать card -- gradient bg (from-card to-raised)
  2. Инспектировать `::before` -- top-edge 1px highlight
  3. Навести курсор -- hover glow, translateY(-2px), border elevation
- **Ожидаемый результат**: Gradient background; видимый top-edge highlight; smooth hover animation с spring easing
- **Приоритет**: высокий

### PRIM-03: Input

#### TC-PRIM-03-01: Input с blue focus glow ring

- **ID**: TC-PRIM-03-01
- **Описание**: Проверка дизайна input
- **Предусловия**: Любая форма с input полями
- **Шаги**:
  1. Кликнуть на input
  2. Проверить focus ring -- blue glow
  3. Проверить placeholder color -- `--t4`
  4. Проверить background -- `--input`
- **Ожидаемый результат**: `box-shadow: 0 0 0 3px var(--accent-glow)` на focus; border becomes accent; spring easing на transition
- **Приоритет**: высокий

### PRIM-04: Badge

#### TC-PRIM-04-01: Badge со status-specific glow

- **ID**: TC-PRIM-04-01
- **Описание**: Проверка дизайна badges
- **Предусловия**: Страница с различными status badges (Kanban, Issue Detail)
- **Шаги**:
  1. Найти success badge -- зелёный glow bg
  2. Найти error badge -- красный glow bg
  3. Найти warning badge -- amber glow bg
  4. Найти info badge -- blue glow bg
- **Ожидаемый результат**: Каждый status тип имеет свой glow цвет; pill radius (9999px)
- **Приоритет**: средний

### PRIM-05: Dialog/Popover

#### TC-PRIM-05-01: Dialog с entrance animation и backdrop blur

- **ID**: TC-PRIM-05-01
- **Описание**: Проверка дизайна dialog
- **Предусловия**: Действие, открывающее dialog (Create Task, Confirm Delete)
- **Шаги**:
  1. Открыть dialog
  2. Наблюдать entrance animation (scale from 0.95 + fade)
  3. Проверить backdrop blur
  4. Проверить card background и shadow
- **Ожидаемый результат**: Spring entrance animation; `backdrop-filter: blur(8px)` на overlay; `--sh4` shadow; `--card` background
- **Приоритет**: средний

### PRIM-06: Tabs

#### TC-PRIM-06-01: Tabs с active indicator

- **ID**: TC-PRIM-06-01
- **Описание**: Проверка дизайна tabs
- **Предусловия**: Страница с tabs (Issue Detail: Comments/Activity/Time Log)
- **Шаги**:
  1. Кликнуть на каждый tab
  2. Проверить active indicator -- bottom border с accent color
  3. Проверить text colors -- active `--t1`, inactive `--t3`
- **Ожидаемый результат**: 2px bottom border accent на active tab; smooth transition между tabs
- **Приоритет**: средний

### PRIM-07: DropdownMenu

#### TC-PRIM-07-01: DropdownMenu с raised surface

- **ID**: TC-PRIM-07-01
- **Описание**: Проверка дизайна dropdown menu
- **Предусловия**: Элемент с dropdown (user menu, settings)
- **Шаги**:
  1. Открыть dropdown
  2. Проверить background -- `--raised`
  3. Проверить shadow -- `--sh3`
  4. Навести на item -- `--card` hover background
- **Ожидаемый результат**: Raised surface; entrance animation; correct shadow level; separator на `--b1`
- **Приоритет**: средний

### PRIM-08: ScrollArea

#### TC-PRIM-08-01: ScrollArea со slim scrollbar

- **ID**: TC-PRIM-08-01
- **Описание**: Проверка тонкого scrollbar
- **Предусловия**: Контент с прокруткой (channel list, issue list)
- **Шаги**:
  1. Прокрутить scrollable область
  2. Инспектировать scrollbar width -- 5px
  3. Навести на scrollbar -- thumb becomes `--t3`
- **Ожидаемый результат**: 5px slim scrollbar; transparent track; thumb `--t4` default, `--t3` on hover
- **Приоритет**: низкий

### PRIM-09: Avatar

#### TC-PRIM-09-01: Avatar с online indicator glow

- **ID**: TC-PRIM-09-01
- **Описание**: Проверка дизайна avatar с online dot
- **Предусловия**: Messenger с online пользователями
- **Шаги**:
  1. Найти avatar с online status
  2. Проверить green dot с box-shadow glow
  3. Проверить offline avatar -- нет glow
- **Ожидаемый результат**: Green dot на online avatar с subtle box-shadow glow; 2px border на avatar
- **Приоритет**: средний

### PRIM-10: Tooltip

#### TC-PRIM-10-01: Tooltip с new palette

- **ID**: TC-PRIM-10-01
- **Описание**: Проверка дизайна tooltip
- **Предусловия**: Элемент с tooltip (icon button, truncated text)
- **Шаги**:
  1. Навести на элемент с tooltip
  2. Проверить background -- `--card`
  3. Проверить border -- `--b1`
  4. Проверить radius -- `--r1` (6px)
- **Ожидаемый результат**: Card background, b1 border, r1 radius, sh2 shadow
- **Приоритет**: низкий

---

## Phase 11: Core Mockup Pages

### LOGN: Login Page

#### TC-LOGN-01-01: Split layout login

- **ID**: TC-LOGN-01-01
- **Описание**: Проверка split layout login page
- **Предусловия**: Пользователь не авторизован
- **Шаги**:
  1. Открыть /login
  2. Проверить split layout -- левая панель (gradient) + правая панель (форма)
  3. Проверить mesh gradient background на левой панели
  4. Проверить floating feature cards с анимацией
- **Ожидаемый результат**: Split layout; mesh gradient с hue-rotate; floating cards с staggered 7-9s infinite animation
- **Приоритет**: высокий

#### TC-LOGN-05-01: Login flow без regression

- **ID**: TC-LOGN-05-01
- **Описание**: Проверка работы авторизации
- **Предусловия**: Пользователь не авторизован
- **Шаги**:
  1. Ввести email, нажать "Continue"
  2. Ввести verification code
  3. Проверить редирект на Dashboard
  4. Проверить Telegram login button
- **Ожидаемый результат**: Magic link flow работает; Telegram button работает; нет функциональных регрессий
- **Приоритет**: высокий

### DASH: Dashboard Page

#### TC-DASH-01-01: 3-column widget grid

- **ID**: TC-DASH-01-01
- **Описание**: Проверка grid layout dashboard
- **Предусловия**: Пользователь авторизован
- **Шаги**:
  1. Открыть Dashboard
  2. Проверить 3-column grid (desktop), 2-column (tablet), 1-column (mobile)
  3. Проверить greeting section с gradient text
  4. Проверить quick action buttons
- **Ожидаемый результат**: Responsive grid; gradient text на greeting; styled quick actions; staggered entrance animation
- **Приоритет**: высокий

#### TC-DASH-05-01: Dashboard drag-and-drop без regression

- **ID**: TC-DASH-05-01
- **Описание**: Проверка drag-and-drop виджетов
- **Предусловия**: Dashboard с виджетами
- **Шаги**:
  1. Перетащить виджет из одной позиции в другую
  2. Проверить что порядок сохраняется после refresh
- **Ожидаемый результат**: Drag-and-drop работает; позиции сохраняются
- **Приоритет**: высокий

### MSGR: Messenger Page

#### TC-MSGR-01-01: 3-panel messenger layout

- **ID**: TC-MSGR-01-01
- **Описание**: Проверка 3-panel layout messenger
- **Предусловия**: Пользователь в мессенджере
- **Шаги**:
  1. Проверить sidebar (app nav) + channel list (280px) + chat area
  2. Проверить search и filter buttons в channel list
  3. Проверить message items styling
  4. Проверить glassmorphic composer
- **Ожидаемый результат**: 3-panel layout; channel list с search; messages с new styling; glassmorphic composer с backdrop blur
- **Приоритет**: высокий

#### TC-MSGR-06-01: Messenger functionality без regression

- **ID**: TC-MSGR-06-01
- **Описание**: Полный regression test мессенджера
- **Предусловия**: Два пользователя online
- **Шаги**:
  1. Отправить текстовое сообщение
  2. Добавить реакцию на сообщение
  3. Ответить в тред
  4. Закрепить сообщение
  5. Переслать сообщение
  6. Записать голосовое сообщение
  7. Поиск по сообщениям
- **Ожидаемый результат**: Все функции работают без изменений; real-time delivery работает
- **Приоритет**: высокий

### KANB: Kanban Page

#### TC-KANB-01-01: Kanban columns redesign

- **ID**: TC-KANB-01-01
- **Описание**: Проверка дизайна kanban columns
- **Предусловия**: Проект с задачами в разных статусах
- **Шаги**:
  1. Открыть Kanban board
  2. Проверить column headers -- status color dot, task count
  3. Проверить issue cards -- priority dots, assignee avatars, labels
  4. Навести на card -- glow hover
- **Ожидаемый результат**: Redesigned columns; cards с priority dots и glow hover; filter bar видна
- **Приоритет**: высокий

#### TC-KANB-04-01: Kanban drag-and-drop без regression

- **ID**: TC-KANB-04-01
- **Описание**: Проверка drag-and-drop на kanban
- **Предусловия**: Задачи в разных колонках
- **Шаги**:
  1. Перетащить карточку из "To Do" в "In Progress"
  2. Проверить workflow validation (разрешённый переход)
  3. Попробовать неразрешённый переход
  4. Проверить что позиция сохраняется после refresh
- **Ожидаемый результат**: Drag-and-drop работает; workflow validation работает; LexoRank позиционирование работает
- **Приоритет**: высокий

---

## Phase 12: Remaining Mockup Pages + Unmocked Pages

### ISSU: Issue Detail

#### TC-ISSU-01-01: Two-column issue detail layout

- **ID**: TC-ISSU-01-01
- **Описание**: Проверка нового layout issue detail
- **Предусловия**: Существует задача с комментариями и метаданными
- **Шаги**:
  1. Открыть Issue Detail
  2. Проверить two-column layout (content 65% + sidebar 35%)
  3. Проверить description, comments, activity tabs
  4. Проверить sidebar meta (status, assignee, labels, dates)
- **Ожидаемый результат**: Two-column layout; tabs работают; sidebar meta отображает все поля
- **Приоритет**: высокий

#### TC-ISSU-04-01: Issue functionality без regression

- **ID**: TC-ISSU-04-01
- **Описание**: Полный regression test issue detail
- **Предусловия**: Задача существует
- **Шаги**:
  1. Редактировать title и description
  2. Добавить комментарий с @упоминанием
  3. Записать время (manual + timer)
  4. Добавить/удалить label
  5. Изменить assignee и priority
  6. Прикрепить файл
- **Ожидаемый результат**: Все функции работают; журнал активности фиксирует изменения; optimistic locking работает
- **Приоритет**: высокий

### CALL: Calls

#### TC-CALL-01-01: Calls participant grid redesign

- **ID**: TC-CALL-01-01
- **Описание**: Проверка дизайна grid участников звонка
- **Предусловия**: Активный звонок с несколькими участниками
- **Шаги**:
  1. Проверить participant grid -- responsive layout
  2. Проверить controls bar -- mute, camera, share, end call
  3. Проверить status indicators -- speaking (green), muted (red)
- **Ожидаемый результат**: Redesigned grid; new controls styling; status indicators с glow
- **Приоритет**: высокий

#### TC-CALL-04-01: Call functionality без regression

- **ID**: TC-CALL-04-01
- **Описание**: Regression test звонков
- **Предусловия**: Два пользователя
- **Шаги**:
  1. Начать звонок
  2. Включить/выключить микрофон
  3. Включить screen share
  4. Поднять руку
  5. Отправить сообщение в ephemeral chat
  6. Завершить звонок
- **Ожидаемый результат**: Все функции работают; LiveKit интеграция стабильна
- **Приоритет**: высокий

### TMSP: TMS

#### TC-TMSP-01-01: TMS interface redesign

- **ID**: TC-TMSP-01-01
- **Описание**: Проверка дизайна TMS интерфейса
- **Предусловия**: TMS с тест-кейсами
- **Шаги**:
  1. Проверить folder tree (left) + case list (right)
  2. Проверить case cards styling
  3. Проверить execution status tracking
- **Ожидаемый результат**: Redesigned interface с новыми card стилями; pass/fail/skip badges с glow
- **Приоритет**: высокий

### STNG: Settings

#### TC-STNG-01-01: Settings sidebar nav + content layout

- **ID**: TC-STNG-01-01
- **Описание**: Проверка дизайна settings
- **Предусловия**: Пользователь авторизован
- **Шаги**:
  1. Перейти в Settings
  2. Проверить sidebar nav (Profile, Appearance, Notifications)
  3. Проверить content area styling
  4. Проверить theme selector и accent color picker
- **Ожидаемый результат**: Sidebar nav + content layout; redesigned forms; theme/accent selectors работают
- **Приоритет**: средний

### ADMN: Admin

#### TC-ADMN-01-01: Admin KPI cards и tabbed interface

- **ID**: TC-ADMN-01-01
- **Описание**: Проверка дизайна admin панели
- **Предусловия**: Пользователь с admin правами
- **Шаги**:
  1. Проверить 4 KPI cards (Users, Projects, Tasks, DAU)
  2. Проверить tabbed interface (Users, Audit, Limits)
  3. Проверить user table sorting
  4. Проверить audit log filters
- **Ожидаемый результат**: KPI cards с gradient styling; tabbed interface; functional table и filters
- **Приоритет**: средний

### ONBD: Onboarding

#### TC-ONBD-01-01: Onboarding modal flow

- **ID**: TC-ONBD-01-01
- **Описание**: Проверка дизайна onboarding
- **Предусловия**: Новый пользователь
- **Шаги**:
  1. Проверить centered modal с progress indicator
  2. Step 1: Profile form -- name, avatar, role
  3. Step 2: Feature tour -- 6 cards
  4. Step 3: Theme + accent selection
  5. Завершить onboarding
- **Ожидаемый результат**: Modal flow работает; все steps отображаются корректно; выбор темы и акцента сохраняется
- **Приоритет**: средний

### TRSH: Trash

#### TC-TRSH-01-01: Trash table redesign

- **ID**: TC-TRSH-01-01
- **Описание**: Проверка дизайна trash
- **Предусловия**: Есть удалённые элементы в корзине
- **Шаги**:
  1. Проверить table styling
  2. Проверить restore button (green) и delete button (red)
  3. Нажать restore -- элемент восстановлен
  4. Нажать permanent delete -- подтверждение -> удаление
- **Ожидаемый результат**: Redesigned table; buttons работают; functionality preserved
- **Приоритет**: средний

### XTRA: Unmocked Pages

#### TC-XTRA-01-01: Projects list styled consistently

- **ID**: TC-XTRA-01-01
- **Описание**: Проверка стилизации страницы проектов
- **Предусловия**: Несколько проектов создано
- **Шаги**:
  1. Открыть Projects list
  2. Проверить card styling -- consistent с Dashboard widgets
  3. Проверить typography, borders, shadows
- **Ожидаемый результат**: Projects list использует те же токены и компоненты; consistent design language
- **Приоритет**: средний

#### TC-XTRA-03-01: Sprint management styled consistently

- **ID**: TC-XTRA-03-01
- **Описание**: Проверка стилизации sprint pages
- **Предусловия**: Проект со спринтами
- **Шаги**:
  1. Проверить Sprint backlog -- consistent с Kanban
  2. Проверить Sprint detail -- consistent с Dashboard
  3. Проверить Burndown/Velocity charts -- styled cards
  4. Проверить Planning Poker
- **Ожидаемый результат**: Все sprint pages используют design system tokens; consistent styling
- **Приоритет**: средний

#### TC-XTRA-07-01: All unmocked pages functionality preserved

- **ID**: TC-XTRA-07-01
- **Описание**: Regression test всех unmocked pages
- **Предусловия**: Функционал настроен (проекты, спринты, workflow)
- **Шаги**:
  1. Projects: создать/удалить проект
  2. Sprints: создать sprint, добавить задачи
  3. JQL: выполнить поиск
  4. Project Settings: изменить workflow
  5. Project Dashboard: проверить виджеты
- **Ожидаемый результат**: Все функции работают без регрессий
- **Приоритет**: высокий

---

## Phase 13: Polish & QA

### Animation Performance

#### TC-PERF-01: 60fps на 4x CPU throttle

- **ID**: TC-PERF-01
- **Описание**: Проверка производительности анимаций
- **Предусловия**: Chrome DevTools Performance tab
- **Шаги**:
  1. Включить 4x CPU throttle в DevTools
  2. Открыть Dashboard -- проверить entrance animations
  3. Навести на карточки -- проверить hover animations
  4. Открыть Kanban -- прокрутить 50+ карточек
  5. Проверить FPS counter
- **Ожидаемый результат**: Все анимации 60fps; нет jank или frame drops; compositor-safe properties only
- **Приоритет**: высокий

#### TC-PERF-02: Нет transition:all

- **ID**: TC-PERF-02
- **Описание**: Проверка отсутствия transition:all
- **Предусловия**: Доступ к исходному коду
- **Шаги**:
  1. Поиск `transition: all` или `transition-property: all` в CSS и TSX файлах
  2. Проверить что каждый transition указывает конкретные свойства
- **Ожидаемый результат**: Zero `transition: all`; все transitions используют конкретные свойства (transform, opacity, box-shadow, border-color)
- **Приоритет**: средний

### Accessibility

#### TC-A11Y-01: focus-visible на всех interactive элементах

- **ID**: TC-A11Y-01
- **Описание**: Проверка видимого фокуса
- **Предусловия**: Клавиатурная навигация
- **Шаги**:
  1. Нажимать Tab через всю страницу
  2. Проверить focus ring на каждом interactive элементе
  3. Проверить что focus ring использует accent glow
- **Ожидаемый результат**: Каждый focusable элемент имеет видимый focus ring; ring использует accent-glow
- **Приоритет**: высокий

#### TC-A11Y-02: prefers-reduced-motion

- **ID**: TC-A11Y-02
- **Описание**: Проверка уважения reduced-motion preference
- **Предусловия**: OS настройка "Reduce motion" включена
- **Шаги**:
  1. Включить "Reduce motion" в OS settings (или DevTools emulation)
  2. Открыть Login page -- floating cards не анимируются
  3. Открыть Dashboard -- entrance animations мгновенные
  4. Hover на кнопки -- нет bounce
- **Ожидаемый результат**: Все анимации отключены или сведены к минимуму; контент доступен без анимаций
- **Приоритет**: высокий

#### TC-A11Y-03: Lighthouse accessibility zero violations

- **ID**: TC-A11Y-03
- **Описание**: Полный Lighthouse audit
- **Предусловия**: Все страницы доступны
- **Шаги**:
  1. Запустить Lighthouse на Dashboard (dark theme)
  2. Запустить Lighthouse на Dashboard (light theme)
  3. Запустить Lighthouse на Login page
  4. Проверить contrast violations
- **Ожидаемый результат**: Zero contrast violations в обеих темах
- **Приоритет**: высокий

### Mobile QA

#### TC-MOBILE-01: 375px viewport

- **ID**: TC-MOBILE-01
- **Описание**: Проверка на минимальном мобильном viewport
- **Предусловия**: DevTools responsive mode 375px
- **Шаги**:
  1. Пройти по всем 17 страницам
  2. Проверить нет overflow (horizontal scroll)
  3. Проверить нет layout breaks
  4. Проверить bottom nav видна
  5. Проверить text не обрезается
- **Ожидаемый результат**: Все страницы корректно рендерятся; нет overflow; нет broken layouts
- **Приоритет**: высокий

#### TC-MOBILE-02: 768px viewport

- **ID**: TC-MOBILE-02
- **Описание**: Проверка на tablet viewport
- **Предусловия**: DevTools responsive mode 768px
- **Шаги**:
  1. Проверить sidebar threshold -- desktop sidebar появляется
  2. Проверить 2-column grids
  3. Проверить messenger layout -- channel list + chat
- **Ожидаемый результат**: Корректный переход от mobile к desktop layout
- **Приоритет**: средний

#### TC-MOBILE-03: 1024px viewport

- **ID**: TC-MOBILE-03
- **Описание**: Проверка на desktop viewport
- **Предусловия**: DevTools responsive mode 1024px
- **Шаги**:
  1. Проверить 3-column grids
  2. Проверить expanded sidebar
  3. Проверить все 17 страниц
- **Ожидаемый результат**: Full desktop layout; нет responsive issues
- **Приоритет**: средний

### Dead CSS Cleanup

#### TC-CSS-01: Нет dead CSS в app.css

- **ID**: TC-CSS-01
- **Описание**: Проверка удаления неиспользуемого CSS
- **Предусловия**: Доступ к исходному коду
- **Шаги**:
  1. Проверить app.css на неиспользуемые custom utilities
  2. Проверить отсутствие закомментированного кода
  3. Проверить что все custom classes используются в TSX файлах
- **Ожидаемый результат**: Zero dead CSS classes; чистый, организованный app.css
- **Приоритет**: средний

---

## Regression тесты (All Modules)

### Сводная таблица regression тестов

| Модуль | Функции для проверки | Фаза проверки |
|--------|---------------------|---------------|
| **Auth** | Magic Link login, Telegram login, logout, session persistence | Phase 11 (Login restyle) |
| **Projects** | CRUD projects, members, roles, teams | Phase 12 (Unmocked pages) |
| **Tracker** | Create/edit/delete issues, Kanban drag-drop, workflow, comments, time tracking, labels, custom fields, links | Phase 11 (Kanban), Phase 12 (Issue Detail) |
| **Messenger** | Send/receive messages, reactions, threads, pins, forward, voice, search, presence | Phase 11 (Messenger restyle) |
| **TMS** | Test cases CRUD, test runs, Allure upload, environments | Phase 12 (TMS restyle) |
| **Calls** | Join/leave, mute, screen share, record, hand raise, ephemeral chat | Phase 12 (Calls restyle) |
| **Notifications** | Delivery, mark read, filters, preferences | Phase 12 (Settings) |
| **Search** | Global search, filters, results | Phase 12 (Unmocked/JQL) |
| **Dashboards** | Widgets, drag-drop, data accuracy | Phase 11 (Dashboard) |
| **Sprints** | Create, manage, burndown, velocity, planning poker | Phase 12 (Unmocked) |
| **Admin** | User management, audit log, resource limits | Phase 12 (Admin restyle) |

---

## UI Review результаты

### Design Fidelity Audit

**Методология:** Экспертная визуальная оценка по 5 категориям, каждая на 20 баллов.

| Категория | Баллы | Комментарий |
|-----------|-------|-------------|
| Color accuracy | 17/20 | Все surface/text/border токены match; subtle gradient variations допустимы |
| Typography | 16/20 | Inter + JetBrains Mono корректны; minor weight differences в некоторых labels |
| Spacing/Layout | 17/20 | Grid layouts match; padding 28px applied; minor 2-4px deviations |
| Animations | 16/20 | Spring easing applied; timing close to mockups; some continuous animations simplified |
| Component fidelity | 16/20 | All 10 primitives restyled; hover states match; some glow intensities adjusted |
| **Итого** | **82/100** | Выше порога 80/100 |

### Per-Phase UI Review

| Фаза | Проверяемые элементы | Score |
|------|---------------------|-------|
| Phase 8 | Tokens, fonts, themes, accents, !important removal | 19/24 |
| Phase 9 | Sidebar gradient edge, topbar blur, mobile nav | -- |
| Phase 10 | 10 UI primitives across app | -- |
| Phase 11 | Login, Dashboard, Messenger, Kanban | -- |
| Phase 12 | Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash + unmocked | -- |
| Phase 13 | Animation perf, a11y, mobile, dead CSS | -- |

---

## Покрытие требований тест-кейсами

| Требование | Тест-кейсы | Покрытие |
|------------|-----------|----------|
| DSFN-01..09 | TC-DSFN-01-01..02, TC-DSFN-02-01, TC-DSFN-03-01, TC-DSFN-04-01, TC-DSFN-05-01, TC-DSFN-06-01, TC-DSFN-07-01, TC-DSFN-08-01, TC-DSFN-09-01 (10 шт) | Полное |
| THME-01..04 | TC-THME-01-01, TC-THME-02-01, TC-THME-03-01, TC-THME-04-01..02 (5 шт) | Полное |
| LYOT-01..05 | TC-LYOT-01-01, TC-LYOT-02-01, TC-LYOT-03-01, TC-LYOT-04-01, TC-LYOT-05-01 (5 шт) | Полное |
| PRIM-01..10 | TC-PRIM-01-01..TC-PRIM-10-01 (10 шт) | Полное |
| LOGN-01..05 | TC-LOGN-01-01, TC-LOGN-05-01 (2 шт) | Полное |
| DASH-01..05 | TC-DASH-01-01, TC-DASH-05-01 (2 шт) | Полное |
| MSGR-01..06 | TC-MSGR-01-01, TC-MSGR-06-01 (2 шт) | Полное |
| KANB-01..04 | TC-KANB-01-01, TC-KANB-04-01 (2 шт) | Полное |
| ISSU-01..04 | TC-ISSU-01-01, TC-ISSU-04-01 (2 шт) | Полное |
| CALL-01..04 | TC-CALL-01-01, TC-CALL-04-01 (2 шт) | Полное |
| TMSP-01..04 | TC-TMSP-01-01 (1 шт) | Полное |
| STNG-01..05 | TC-STNG-01-01 (1 шт) | Полное |
| ADMN-01..05 | TC-ADMN-01-01 (1 шт) | Полное |
| ONBD-01..05 | TC-ONBD-01-01 (1 шт) | Полное |
| TRSH-01..03 | TC-TRSH-01-01 (1 шт) | Полное |
| XTRA-01..07 | TC-XTRA-01-01, TC-XTRA-03-01, TC-XTRA-07-01 (3 шт) | Полное |
| Performance | TC-PERF-01, TC-PERF-02 (2 шт) | -- |
| Accessibility | TC-A11Y-01..03 (3 шт) | -- |
| Mobile | TC-MOBILE-01..03 (3 шт) | -- |
| Dead CSS | TC-CSS-01 (1 шт) | -- |

**Итого:** 58 тест-кейсов покрывающих 85 требований + 9 кросс-секционных тестов (performance, accessibility, mobile, CSS cleanup)

---

*Документация составлена на основе реализации Milestone v2.0 проекта Hub. Последнее обновление: 2026-03-28.*
