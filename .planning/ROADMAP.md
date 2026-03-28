# Roadmap: Hub

## Overview

Hub предоставляет единую платформу для команд разработки, объединяя пять интегрированных модулей (аутентификация/проекты, трекер задач, мессенджер, TMS, голосовые звонки) и сквозные функции (уведомления, поиск, дашборды) на общем фундаменте. Дорожная карта следует цепочке зависимостей: сначала базовая инфраструктура, затем четыре основных модуля в порядке ценности и зависимостей (трекер, мессенджер, TMS, звонки), далее сквозные агрегирующие функции, и наконец доработка и расширенные возможности. Каждая фаза поставляет завершённую, проверяемую функциональность, основанную на предыдущих фазах.

## Milestones

- :white_check_mark: **v1.0 MVP** - Phases 1-7 (shipped 2026-03-21)
- :construction: **v2.0 Hub Redesign** - Phases 8-13 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>v1.0 MVP (Phases 1-7) - SHIPPED 2026-03-21</summary>

- [x] **Phase 1: Foundation & Infrastructure** - Аутентификация, проекты, RBAC, инфраструктура реального времени, хранилище, среда разработки
- [x] **Phase 2: Task Tracker** - CRUD задач, Kanban-доски, рабочие процессы, комментарии, лог активности, учёт времени (completed 2026-03-21)
- [x] **Phase 3: Messenger** - Личные сообщения, каналы, треды, реакции, голосовые сообщения, присутствие, поиск
- [x] **Phase 4: TMS & Allure** - Тест-кейсы, тестовые прогоны, загрузка Allure-отчётов, связь с задачами, отчётность, импорт/экспорт
- [x] **Phase 5: Voice Calls** - Аудиозвонки, демонстрация экрана, управление звонками, эфемерный чат (completed 2026-03-21)
- [x] **Phase 6: Notifications, Search & Dashboards** - Уведомления в приложении, глобальный поиск, персональные/проектные/административные дашборды, аудит
- [x] **Phase 7: Polish & Extended Features** - Спринты, интеграция с GitHub, запись звонков, интернационализация, темы, PWA, API, резервное копирование (completed 2026-03-21)

</details>

### v2.0 Hub Redesign (In Progress)

**Milestone Goal:** Pixel-perfect редизайн всего фронтенда Hub по 11 HTML-макетам без ломания существующего функционала

- [ ] **Phase 8: Design System Foundation** - CSS-токены, шрифты, !important-очистка, тёмная и светлая палитры, радиусы, тени, анимации
- [ ] **Phase 9: Layout Shell** - Сайдбар, топбар, AppShell с новым дизайном
- [ ] **Phase 10: UI Primitives** - Все Shadcn-компоненты рестайлированы под новый дизайн-язык
- [ ] **Phase 11: Core Mockup Pages** - Login, Dashboard, Messenger, Kanban
- [ ] **Phase 12: Remaining Mockup Pages + Unmocked Pages** - Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash + Projects, Sprint, JQL, Project Settings, Project Dashboard
- [ ] **Phase 13: Polish & QA** - Финальный визуальный QA, доступность, мобильная проверка, очистка мёртвого CSS

## Phase Details

<details>
<summary>v1.0 MVP Phase Details (Phases 1-7)</summary>

### Phase 1: Foundation & Infrastructure
**Goal**: Пользователи могут аутентифицироваться, создавать проекты с командами и ролями, а платформа имеет работающую инфраструктуру реального времени, файловое хранилище и CI-пайплайн — субстрат, от которого зависит каждый модуль
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, PROJ-01, PROJ-02, PROJ-03, PROJ-04, PROJ-06, PROJ-07, UI-01, UI-07, UI-08
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может войти через Magic Link по email и через Telegram, попасть на рабочую оболочку дашборда и выйти
  2. Пользователь может создать проект и увидеть автоматически сгенерированные доску трекера, пространство TMS, чат проекта и заглушки каналов
  3. Владелец проекта может приглашать пользователей по ссылке, назначать роли (Owner/Developer/Tester/Viewer), и ограничения по ролям применяются
  4. Файлы можно загружать перетаскиванием или вставкой, они сохраняются в SeaweedFS с рабочим предпросмотром
  5. WebSocket-соединение доставляет события реального времени из NATS в браузер (проверено тестовым событием)
**Plans**: 9 plans

Plans:
- [x] 01-01: Среда разработки, структура проекта, CI-пайплайн, Docker Compose со всеми сервисами
- [x] 01-02: Схемы базы данных (Ent + MongoDB), настройка NATS, Redis, интеграция с SeaweedFS
- [x] 01-03: Модуль аутентификации (Magic Link + вход через Telegram, сессии, RBAC-мидлвара)
- [x] 01-04: Модуль проектов (CRUD, членство, роли, команды, архив/корзина)
- [x] 01-05: WebSocket gateway (NATS bridge) + frontend API client, stores, hooks, permissions
- [x] 01-07: React UI оболочка (маршрутизация, sidebar, topbar, login, проекты, файлы)
- [x] 01-06: Документация фазы 1 (бизнес-дока, тестовая дока, техническая дока)
- [x] 01-08: [GAP] Подключение модуля хранилища в main.go, исправление auth middleware в storage handler
- [x] 01-09: [GAP] Реализация CRUD команд (teams) вместо заглушек 501

### Phase 2: Task Tracker
**Goal**: Пользователи могут управлять работой над проектом через полнофункциональный трекер задач с Kanban-досками, настраиваемыми рабочими процессами, расширенными комментариями и учётом времени
**Depends on**: Phase 1
**Requirements**: TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-06, TASK-07, TASK-08, TASK-09, TASK-10, TASK-11, KANB-01, KANB-04, KANB-05, UI-09
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может создавать задачи всех 9 типов с описаниями в формате rich text, вложениями файлов и ID с префиксом проекта (HUB-123)
  2. Пользователь может перетаскивать задачи на Kanban-доске между колонками, соответствующими статусам рабочего процесса, с обновлениями в реальном времени, видимыми другим пользователям
  3. Владелец проекта может настраивать рабочий процесс, добавляя/удаляя статусы и определяя допустимые переходы
  4. Пользователь может создавать иерархии родитель-потомок и межзадачные связи (блокирует, связан с, дублирует и т.д.)
  5. Пользователь может комментировать с @упоминаниями и реакциями, просматривать полную историю изменений и вести учёт времени с таймером старт/стоп
**Plans**: 5 plans

Plans:
- [x] 02-01-PLAN.md — Ent-схемы трекера (15 схем), Repository interface, DTOs, events
- [x] 02-02-PLAN.md — Issue CRUD backend (handler/service/repository), workflow engine, тесты
- [x] 02-03-PLAN.md — Frontend API client, Zustand stores, Plate editor wrapper
- [x] 02-04-PLAN.md — Kanban-доска (drag-and-drop), списки задач, детальный вид, комментарии, настройки
- [x] 02-05-PLAN.md — Документация фазы 2 (бизнес, тестовая, техническая) на русском

### Phase 3: Messenger
**Goal**: Пользователи могут общаться в реальном времени через личные сообщения, каналы и треды с форматированным текстом, реакциями, голосовыми сообщениями и индикаторами присутствия
**Depends on**: Phase 1, Phase 2
**Requirements**: MSG-01, MSG-02, MSG-03, MSG-04, MSG-05, MSG-06, MSG-07, MSG-08, MSG-09, MSG-10, MSG-12, MSG-13
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может отправлять и получать сообщения в реальном времени через личные сообщения, групповые чаты, публичные и приватные каналы
  2. Пользователь может отвечать в тредах, ставить реакции эмодзи, форматировать текст и записывать голосовые сообщения
  3. Каналы проекта создаются автоматически при создании проекта; каналы только для чтения работают для сообщений ботов
  4. Пользователь может редактировать/удалять свои сообщения, закреплять, добавлять в закладки и искать
  5. Статус онлайн/офлайн/отошёл виден в реальном времени и счётчики непрочитанных корректно обновляются
**Plans**: 5 plans

Plans:
- [x] 03-01-PLAN.md — Бэкенд ядра мессенджера
- [x] 03-02-PLAN.md — Управление каналами
- [x] 03-03-PLAN.md — Фронтенд чата
- [x] 03-04-PLAN.md — Присутствие, непрочитанные, поиск, пины, закладки, пересылка
- [x] 03-05-PLAN.md — Документация фазы 3

### Phase 4: TMS & Allure
**Goal**: Пользователи могут управлять тест-кейсами, выполнять тестовые прогоны, загружать результаты Allure и видеть связи между тестами и задачами
**Depends on**: Phase 1, Phase 2
**Requirements**: TMS-01, TMS-02, TMS-03, TMS-04, TMS-05, TMS-06, TMS-08, TMS-09, TMS-10, ALLU-01, ALLU-02, ALLU-03, IMEX-02
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может создавать тест-кейсы в классическом и BDD форматах с иерархией папок и версионированием
  2. Пользователь может создавать и выполнять ручные тестовые прогоны с результатами по каждому шагу
  3. JSON-результаты Allure загружаются через API и сопоставляются с тест-кейсами
  4. Тест-кейсы двусторонне связаны с задачами
  5. Отчёты показывают процент прохождения, тренды, обнаружение нестабильных тестов
**Plans**: 5 plans

Plans:
- [x] 04-01-PLAN.md — Ent-схемы TMS, handler/service/repository
- [x] 04-02-PLAN.md — Тестовые прогоны + фронтенд TMS UI
- [x] 04-03-PLAN.md — Интеграция с Allure
- [x] 04-04-PLAN.md — Кросс-модульные связи, отчётность, импорт/экспорт
- [x] 04-05-PLAN.md — Документация фазы 4

### Phase 5: Voice Calls
**Goal**: Пользователи могут присоединяться к аудиозвонкам с демонстрацией экрана
**Depends on**: Phase 1, Phase 3
**Requirements**: CALL-01, CALL-03, CALL-06
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь может начать или присоединиться к аудиозвонку с участием до 50 человек
  2. Пользователь может включать/выключать микрофон, поднимать руку и демонстрировать экран
  3. Эфемерный чат звонка доступен во время звонка и исчезает после завершения
  4. События звонка отображаются в канале мессенджера
**Plans**: 4 plans

Plans:
- [x] 05-01-PLAN.md — Интеграция с LiveKit
- [x] 05-02-PLAN.md — UI звонков
- [x] 05-03-PLAN.md — Жизненный цикл звонка
- [x] 05-04-PLAN.md — Документация фазы 5

### Phase 6: Notifications, Search & Dashboards
**Goal**: Пользователи имеют уведомления, глобальный поиск и дашборды
**Depends on**: Phase 1-5
**Requirements**: NOTF-01, NOTF-02, NOTF-03, NOTF-04, SRCH-01, SRCH-02, DASH-01, DASH-02, DASH-03, AUDT-01, AUDT-02, AUDT-03, AUDT-04, UI-05
**Success Criteria** (что должно быть ИСТИННО):
  1. Пользователь получает уведомления в приложении о назначениях, @упоминаниях и сменах статуса
  2. Ctrl+K открывает глобальный поиск по всей платформе
  3. Персональный дашборд показывает задачи, непрочитанные, звонки, прогоны и дедлайны
  4. Дашборд проекта показывает метрики; административный дашборд показывает метрики системы
**Plans**: 5 plans

Plans:
- [x] 06-01-PLAN.md — Система уведомлений
- [x] 06-02-PLAN.md — Глобальный поиск
- [x] 06-03-PLAN.md — Дашборды
- [x] 06-04-PLAN.md — Журнал аудита, лимиты, корзина, GDPR
- [x] 06-05-PLAN.md — Документация фазы 6

### Phase 7: Polish & Extended Features
**Goal**: Спринты, интеграция с GitHub, запись звонков, i18n, темы, PWA, API, бэкапы
**Depends on**: Phase 1-6
**Requirements**: PROJ-05, KANB-02, KANB-03, SCRM-01, SCRM-02, SCRM-03, SCRM-04, SCRM-05, MSG-11, TMS-07, CALL-02, CALL-04, CALL-05, GHUB-01, GHUB-02, GHUB-03, GHUB-04, GHUB-05, SRCH-03, API-01, API-02, API-03, API-04, IMEX-01, IMEX-03, BKUP-01, BKUP-02, BKUP-03, UI-02, UI-03, UI-04, UI-06
**Success Criteria** (что должно быть ИСТИННО):
  1. Scrum-команды могут создавать спринты с Planning Poker и графиками burndown/velocity
  2. PR/коммиты GitHub автоматически связываются с задачами по ID
  3. Пользователи могут записывать звонки
  4. Приложение работает на русском и английском с переключением тем и PWA
  5. REST API документирован, API-ключи, вебхуки, автоматические бэкапы
**Plans**: 6 plans

Plans:
- [x] 07-01-PLAN.md — Управление спринтами
- [x] 07-02-PLAN.md — Интеграция с GitHub
- [x] 07-03-PLAN.md — Запись звонков, запланированные/повторяющиеся звонки
- [x] 07-04-PLAN.md — Интернационализация, темизация, мобильная раскладка, PWA, превью ссылок
- [x] 07-05-PLAN.md — REST API, API-ключи, вебхуки, hub-cli, импорт/экспорт, шаблоны, бэкапы, WIP, swimlanes, окружения, JQL
- [x] 07-06-PLAN.md — Документация фазы 7

</details>

### Phase 8: Design System Foundation
**Goal**: Приложение переходит на новую дизайн-систему — все поверхности, текст, границы, тени, радиусы, анимации и шрифты соответствуют новой палитре, а тёмная и светлая темы работают корректно
**Depends on**: Phase 7 (v1.0 complete)
**Requirements**: DSFN-01, DSFN-02, DSFN-03, DSFN-04, DSFN-05, DSFN-06, DSFN-07, DSFN-08, DSFN-09, THME-01, THME-02, THME-03, THME-04
**Success Criteria** (what must be TRUE):
  1. User sees Inter font for all UI text and JetBrains Mono for code elements across the entire application
  2. User sees the new 5-level surface palette (void/base/raised/card/input), 4-level text hierarchy, and 3-level border system in dark mode matching mockup values
  3. User can switch between dark/light/system themes without visual glitches, and light theme passes WCAG AA contrast ratios
  4. User can select accent color and see it applied consistently across all themed surfaces
  5. No !important declarations remain in app.css, and spring/bounce easing is applied on hover/focus/enter animations
**Plans**: 2 plans

Plans:
- [x] 08-01-PLAN.md — Font swap + CSS token foundation (surfaces, text, borders, radii, shadows, easing, accents, dual Shadcn mapping)
- [x] 08-02-PLAN.md — !important removal (56 declarations) + noise texture overlay

**UI hint**: yes

### Phase 9: Layout Shell
**Goal**: Пользователь видит обновлённый сайдбар с градиентной линией и свечением активного пункта, топбар с backdrop blur, и вся оболочка работает на всех размерах экрана
**Depends on**: Phase 8
**Requirements**: LYOT-01, LYOT-02, LYOT-03, LYOT-04, LYOT-05
**Success Criteria** (what must be TRUE):
  1. User sees redesigned sidebar with gradient edge line, new background, and active navigation item with accent glow indicator
  2. User sees redesigned topbar with backdrop blur effect and gradient bottom edge
  3. Sidebar collapse/expand works without visual regression at all viewport sizes
  4. Mobile navigation (bottom nav, overlay sidebar) works without regression
**Plans**: 2 plans

Plans:
- [x] 09-01-PLAN.md — Sidebar redesign (background, gradient edge, active glow, header, user/projects tokens)
- [x] 09-02-PLAN.md — Topbar redesign (backdrop blur, gradient edge), AppShell topbar visibility, mobile nav tokens

**UI hint**: yes

### Phase 10: UI Primitives
**Goal**: Все 10 базовых Shadcn-компонентов рестайлированы под новый дизайн-язык, и любая страница, использующая эти компоненты, автоматически выглядит обновлённой
**Depends on**: Phase 8, Phase 9
**Requirements**: PRIM-01, PRIM-02, PRIM-03, PRIM-04, PRIM-05, PRIM-06, PRIM-07, PRIM-08, PRIM-09, PRIM-10
**Success Criteria** (what must be TRUE):
  1. User sees redesigned Button (gradient primary, ghost/secondary variants, glow hover) and Card (gradient bg, top-edge highlight, hover glow) across the app
  2. User sees redesigned Input (blue focus glow ring), Badge (status-specific glow), and Avatar (online indicator glow)
  3. User sees redesigned Dialog/Popover (entrance animation, backdrop blur), DropdownMenu, Tabs (active indicator), Tooltip, and ScrollArea (slim 5px scrollbar)
  4. All existing page functionality that uses these components continues to work without regression
**Plans**: 3 plans

Plans:
- [x] 10-01-PLAN.md — Button, Card, Input restyle (gradient primary, card gradient+highlight, input focus glow)
- [x] 10-02-PLAN.md — Badge, ScrollArea, Avatar, Tooltip restyle (status glows, 5px scrollbar, raised surfaces)
- [x] 10-03-PLAN.md — Dialog, Tabs, DropdownMenu restyle (backdrop blur, active indicator, raised menus)
**UI hint**: yes

### Phase 11: Core Mockup Pages
**Goal**: Четыре самые важные и сложные страницы (Login, Dashboard, Messenger, Kanban) pixel-perfect соответствуют HTML-макетам, а вся существующая функциональность работает без регрессий
**Depends on**: Phase 10
**Requirements**: LOGN-01, LOGN-02, LOGN-03, LOGN-04, LOGN-05, DASH-01, DASH-02, DASH-03, DASH-04, DASH-05, MSGR-01, MSGR-02, MSGR-03, MSGR-04, MSGR-05, MSGR-06, KANB-01, KANB-02, KANB-03, KANB-04
**Success Criteria** (what must be TRUE):
  1. User sees Login page with split layout (animated gradient left panel with floating feature cards, form right panel), mesh gradient background, and redesigned email/code flow -- login via magic link and Telegram works without regression
  2. User sees Dashboard with 3-column widget grid, greeting section with gradient text, quick action buttons, and staggered entrance animations -- widget drag-and-drop works without regression
  3. User sees Messenger with redesigned 3-panel layout, channel list with search/filter, message items with new styling, glassmorphic composer, and redesigned reactions/typing/reply -- all messenger functionality (send, react, thread, pin, forward, voice, search) works without regression
  4. User sees Kanban with redesigned columns, issue cards with priority dots and glow hover, filter/view options bar -- drag-and-drop works without regression
**Plans**: 4 plans

Plans:
- [x] 11-01-PLAN.md — Login page restyle (split layout, mesh gradient, floating cards, form)
- [x] 11-02-PLAN.md — Dashboard page restyle (greeting gradient, quick actions, widget grid)
- [ ] 11-03-PLAN.md — Messenger page restyle (channel list, messages, glassmorphic composer)
- [ ] 11-04-PLAN.md — Kanban page restyle (columns, issue cards, filter bar)
**UI hint**: yes

### Phase 12: Remaining Mockup Pages + Unmocked Pages
**Goal**: Все оставшиеся 7 страниц с макетами pixel-perfect соответствуют HTML-макетам, а 6 страниц без макетов приведены к единому дизайн-языку -- вся функциональность работает без регрессий
**Depends on**: Phase 11
**Requirements**: ISSU-01, ISSU-02, ISSU-03, ISSU-04, CALL-01, CALL-02, CALL-03, CALL-04, TMSP-01, TMSP-02, TMSP-03, TMSP-04, STNG-01, STNG-02, STNG-03, STNG-04, STNG-05, ADMN-01, ADMN-02, ADMN-03, ADMN-04, ADMN-05, ONBD-01, ONBD-02, ONBD-03, ONBD-04, ONBD-05, TRSH-01, TRSH-02, TRSH-03, XTRA-01, XTRA-02, XTRA-03, XTRA-04, XTRA-05, XTRA-06, XTRA-07
**Success Criteria** (what must be TRUE):
  1. User sees redesigned Issue Detail (two-column layout, comments/activity, right sidebar meta) -- all issue functionality (edit, comment, time track, checklist, attachments) works without regression
  2. User sees redesigned Calls (participant grid, controls, status indicators), TMS (test management interface, case grid, execution tracking) -- all call and TMS functionality works without regression
  3. User sees redesigned Settings (sidebar nav + content, profile, appearance, notifications), Admin (KPI cards, tabbed interface, user table, audit log), Onboarding (centered modal, profile form, feature tour, theme selection), and Trash (table, restore/delete buttons) -- all functionality works without regression
  4. User sees Projects list, Project overview, Sprint management, JQL search, Project settings, and Project dashboard pages styled consistently with the design system -- all existing functionality on these pages works without regression
**Plans**: TBD
**UI hint**: yes

### Phase 13: Polish & QA
**Goal**: Финальная визуальная проверка качества, доступность, мобильная адаптация и очистка кодовой базы -- приложение визуально безупречно в обеих темах на всех размерах экрана
**Depends on**: Phase 12
**Requirements**: (cross-cutting quality -- validates all previous phases)
**Success Criteria** (what must be TRUE):
  1. All animations run at 60fps on 4x CPU throttle (Dashboard and Kanban benchmark pages)
  2. Lighthouse accessibility audit shows zero contrast violations in both dark and light themes across all pages
  3. All pages render correctly at 375px, 768px, and 1024px viewports without layout breaks or overflow
  4. Dead CSS classes removed from app.css; no unused custom utilities remain
  5. Theme toggle transition is smooth (200ms background/color transition on key surfaces)
**Plans**: TBD
**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 8 -> 9 -> 10 -> 11 -> 12 -> 13

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation & Infrastructure | v1.0 | 9/9 | Complete | 2026-03-21 |
| 2. Task Tracker | v1.0 | 5/5 | Complete | 2026-03-21 |
| 3. Messenger | v1.0 | 5/5 | Complete | 2026-03-21 |
| 4. TMS & Allure | v1.0 | 5/5 | Complete | 2026-03-21 |
| 5. Voice Calls | v1.0 | 4/4 | Complete | 2026-03-21 |
| 6. Notifications, Search & Dashboards | v1.0 | 5/5 | Complete | 2026-03-21 |
| 7. Polish & Extended Features | v1.0 | 6/6 | Complete | 2026-03-21 |
| 8. Design System Foundation | v2.0 | 1/2 | In Progress|  |
| 9. Layout Shell | v2.0 | 0/2 | Not started | - |
| 10. UI Primitives | v2.0 | 3/3 | Complete    | 2026-03-28 |
| 11. Core Mockup Pages | v2.0 | 2/4 | In Progress|  |
| 12. Remaining Mockup Pages + Unmocked Pages | v2.0 | 0/? | Not started | - |
| 13. Polish & QA | v2.0 | 0/? | Not started | - |
