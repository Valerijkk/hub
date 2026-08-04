# Дизайн: спека и передача во фронтенд

> Часть плейбука. Канон и оглавление — [`../design.md`](../design.md).

Формат текстовой спеки, крайние случаи, разбивка задач для fe.

---

## 6.1 Почему текстовая спека

В Hub нет Figma. Дизайн-контракт живёт как **Markdown-спека** (`DESIGN-SPEC-<screen>.md`) плюс, при
необходимости, HTML-мокап через `gsd-sketch`. Текст — это то, что `fe` может точно реализовать, а
`designer-ver`/`fe-ver` — проверить.
## 6.2 Обязательные разделы спеки

Каждая дизайн-спека содержит ровно эти разделы:

```
# DESIGN-SPEC: <название экрана / компонента>

## 1. Контекст
- Связанные US/AC: US-Mxx, ...
- Персона и сценарий: кто и зачем открывает экран
- Точка входа / выхода (роут, откуда попадает пользователь)

## 2. Структура (layout)
- ASCII-схема раскладки (см. 6.3)
- Дерево компонентов: какие Shadcn-примитивы и в какой вложенности
- Сетка/колонки, breakpoints поведение

## 3. Токены
- Поверхности: какие --void/--base/--raised/--card/--input где
- Текст: --t1/--t2/--t3 по элементам
- Границы, радиусы (--rN), тени (--shN), spacing (шкала)
- Акцент: где --primary, где статичные семантики

## 4. Состояния
- default / hover / active / focus / disabled для каждого интерактива
- loading (скелетон), empty (EmptyState текст), error (сообщение + retry)

## 5. Поведение
- Что происходит по клику/вводу/drag
- Async: оптимистичные обновления, инвалидация запросов
- Навигация, hotkeys, фокус-менеджмент

## 6. Доступность
- Контраст ключевых пар (рассчитан)
- aria-*, labels, порядок фокуса, keyboard-map
- Цели нажатия

## 7. Edge-cases
- Длинный текст / переполнение
- 0 / 1 / много элементов
- Оффлайн / медленная сеть
- Права (RBAC): что видит Viewer vs Owner

## 8. Разбивка для fe (кто что делает)
- fe1: <файлы/компоненты>
- fe2: <файлы/компоненты>
- точки синхронизации
```
## 6.3 Пример ASCII-схемы раскладки

Раскладку описывают ASCII-диаграммой — она однозначна и не требует Figma:

```
+--------------------------------------------------------------+
| PageHeader (sticky, page-header-blur)                        |
|  [H1 text-2xl t1]              [Filter chips] [+ Button pri.] |
+--------------------------------------------------------------+
| Toolbar: [Search input h-42] [Tabs] [View toggle]            |
+------------------+-------------------------------------------+
| Sidebar (aside)  | Content (main)                            |
| bg-sidebar       |  Card grid (gap-4)                        |
| w-64 lg: only    |  +----------+ +----------+ +----------+   |
|                  |  | Card r4  | | Card r4  | | Card r4  |   |
|                  |  | sh1      | |          | |          |   |
|                  |  +----------+ +----------+ +----------+   |
|                  |  EmptyState (py-12) if no data           |
+------------------+-------------------------------------------+
```
## 6.4 Пример дерева компонентов

```
<ErrorBoundary>
  <PageHeader>
    <h1 className="text-2xl font-semibold text-t1">Sprint Backlog</h1>
    <Button variant="default">+ Задача</Button>   // Ctrl+N
  </PageHeader>
  <div className="flex gap-4">
    <aside className="hidden lg:block w-64">
      <FilterPanel />                              // fe2
    </aside>
    <main className="flex-1">
      {isLoading && <SkeletonList count={6} />}    // animate-shimmer
      {isEmpty && <EmptyState icon={Inbox} title="Пусто" description="Создайте задачу" />}
      {isError && <InlineError onRetry={refetch} />}
      {data && <IssueList items={data} />}         // fe1
    </main>
  </div>
</ErrorBoundary>
```
## 6.5 GSD-инструменты дизайна

`designer`/`fe` используют GSD-скиллы для дизайна:

| Скилл | Когда | Что даёт |
|---|---|---|
| `gsd-sketch` | Быстрый набросок нового экрана | Одноразовый HTML-мокап, мульти-вариант, до фиксации спеки |
| `gsd-sketch-wrap-up` | После sketch | Упаковывает находки в persistent skill |
| `gsd-ui-phase` | Фаза с фронтендом | Генерирует UI-SPEC.md (дизайн-контракт фазы) — база для DESIGN-SPEC |
| `gsd-ui-review` | После реализации | Ретроспективный 6-pillar visual-аудит реализованного UI |

Поток: `gsd-sketch` (варианты) → зафиксировать DESIGN-SPEC → `fe` реализует → `gsd-ui-review` / `fe-ver`.

### Good / Bad — спека

**Good:** спека содержит все 8 разделов, ASCII-раскладку, дерево компонентов с конкретными Shadcn,
рассчитанный контраст и разбивку fe1/fe2. `fe` реализует без вопросов.

**Bad:** спека — «сделать красиво как в Linear, тёмная тема». Нет токенов, нет состояний, нет edge-cases.
`fe` домысливает, результат непредсказуем, `designer-ver` не может верифицировать.

### Чек-лист спеки (§6)

- [ ] Присутствуют все 8 разделов.
- [ ] Есть ASCII-раскладка и дерево компонентов с конкретными Shadcn-примитивами.
- [ ] Все токены названы (реальные из `app.css`).
- [ ] Описаны состояния (5 интерактивных + loading/empty/error).
- [ ] Рассчитан контраст, задан keyboard-map и aria.
- [ ] Перечислены edge-cases и RBAC-видимость.
- [ ] Работа разбита между fe1/fe2 без пересечения файлов.

---

# 7. Верификация дизайна (`designer-ver`)

`designer-ver` — независимая нейросеть, проверяющая артефакт `designer` по коду и чек-листам. Не ставит
`verified: true` без прохождения всех блоков ниже. Итог — отчёт с явными PASS/FAIL по пунктам.
## 7.1 Блок «Консистентность токенов»

- [ ] Каждый токен из спеки существует в `frontend/src/app.css` (grep по имени переменной).
- [ ] Нет HEX-хардкода вместо токена (кроме обоснованных исключений вроде статусных `.badge-*`).
- [ ] Primary-элементы используют `--primary` (accent-aware), а не жёсткий `--blue`.
- [ ] Радиусы из `--r1..--r4`/`--pill`; тени из `--sh1..--sh4`; glow из `--gB/gP/gG/gR`.
- [ ] Spacing по 4px-шкале; нет произвольных `p-[13px]` без причины.
## 7.2 Блок «Доступность»

- [ ] Контраст пар «текст/фон» рассчитан и ≥ 4.5:1 (текст) / ≥ 3:1 (крупный/UI).
- [ ] Focus-индикаторы описаны и видимы; глобальный `focus-visible` не удалён.
- [ ] Информация не только цветом (есть текст/иконка-дубликат).
- [ ] Полный keyboard-map; Radix-поведение не изменено.
- [ ] Цели нажатия ≥ 44px (touch); labels/aria/alt заданы.
## 7.3 Блок «Отзывчивость»

- [ ] Заданы breakpoints и поведение на mobile/tablet/desktop.
- [ ] Mobile-first: базовые классы мобильные.
- [ ] Модалки/таблицы имеют мобильную стратегию (Sheet / карточки / горизонтальный скролл).
- [ ] Нет фиксированных ширин, ломающих узкие экраны.
## 7.4 Блок «Соответствие дизайн-системе»

- [ ] Использованы существующие Shadcn-примитивы, где возможно.
- [ ] Расширение только `className`/`cva`; нет `style={}`/`!important`.
- [ ] Пустые состояния через `EmptyState`, ошибки через `ErrorBoundary`.
- [ ] Экран визуально консистентен с соседними модулями.
- [ ] Учтён `prefers-reduced-motion`; dark-тема — основная и корректна.
## 7.5 Формат отчёта `designer-ver`

```
# DESIGN-VER: <экран> — <дата HH:MM>

verified: false            # true только при 0 FAIL

## Токены:        PASS/FAIL (детали)
## a11y:          PASS/FAIL (контраст-таблица)
## Responsive:    PASS/FAIL
## Дизайн-система: PASS/FAIL

## Findings
- [CRITICAL] ...
- [HIGH] ...
- [MINOR] ...

## Вердикт
- Причина отказа / условия для verified: true
```

> `verified: true` ставится только при отсутствии CRITICAL/HIGH и приложенном обосновании. Иначе спека
> возвращается `designer` с конкретными findings. См. цикл верификации `team-workflow.md` §7.

### Чек-лист `designer-ver` (§7)

- [ ] Пройдены все 4 блока (токены, a11y, responsive, дизайн-система).
- [ ] Контраст рассчитан, а не «на глаз».
- [ ] Findings классифицированы (CRITICAL/HIGH/MINOR).
- [ ] Вердикт с явным `verified: true/false` и датой.

---

# 8. Definition of Done для дизайн-артефакта

Дизайн-артефакт (спека или обновление дизайн-системы) считается **Done**, когда выполнены **все** пункты:

| # | Критерий | Ответственный |
|---|---|---|
| 1 | Спека содержит все 8 разделов (§6.2) | `designer` |
| 2 | Все токены реальны и существуют в `app.css` | `designer` |
| 3 | Описаны 5 интерактивных состояний + loading/empty/error | `designer` |
| 4 | Контраст ключевых пар рассчитан и проходит AA | `designer` |
| 5 | Keyboard-map, aria, цели нажатия заданы | `designer` |
| 6 | Breakpoints и mobile-поведение описаны | `designer` |
| 7 | Edge-cases и RBAC-видимость перечислены | `designer` |
| 8 | Работа разбита между fe1/fe2 без конфликтов | `designer` |
| 9 | `designer-ver` вернул `verified: true` (0 CRITICAL/HIGH) | `designer-ver` |
| 10 | Если менялся `app.css` — регрессия frontend green (`tsc --noEmit`, `vitest`, `lint`) | `fe` / `fe-ver` |
| 11 | Docs-sync: обновлены `SHORTCUTS.md`/`QUICKSTART.md` при новом hotkey/flow; `STATE.md` | `designer` / `docs-writer` |
| 12 | Handoff в `fe` написан (2–5 предложений + артефакт) | `designer` |

> Дизайн-артефакт **не** уходит в реализацию, пока пункты 1–9 не закрыты. Реализация (`fe`) закрывает
> 10–12 совместно с `fe-ver`. Финальную приёмку делает `team-lead`.

### DoD-гейт (быстрый)

- [ ] Спека полная и верифицирована (`verified: true`).
- [ ] Токены реальны, состояния и a11y описаны.
- [ ] Регрессия green (если трогали код/`app.css`).
- [ ] Docs-sync выполнен.
- [ ] Handoff передан.

---

# 9. Антипаттерны дизайна

Сводная таблица запрещённых паттернов. Любой из них — основание для FAIL у `designer-ver`/`fe-ver`.

| # | Антипаттерн | Почему плохо | Правильно |
|---|---|---|---|
| 1 | Хардкод цвета (`#191a20`, `rgb(...)`) | Не переключается тема/accent, рассинхрон с палитрой | Токен: `bg-card`, `var(--raised)`, `text-t1` |
| 2 | Inline-стиль поверх системы (`style={{background:...}}`) | Обходит дизайн-систему, не тематизируется | `className` с токенами |
| 3 | `!important` в CSS/классах | Ломает каскад, было 56 случаев (Phase 8 вычистил) | Правильная специфичность через токены/утилиты |
| 4 | Произвольные отступы (`p-[13px]`, `mt-[7px]`) | Ломает визуальный ритм 4px-сетки | Шкала Tailwind (`p-3`, `p-4`) |
| 5 | Недоступный контраст (важный текст на `--t3`) | < 4.5:1, нечитабельно, нарушает WCAG AA | Важный текст на `--t1`/`--t2` |
| 6 | Изменение поведения Radix (пропсы, focus, data-state) | Ломает a11y и UX Shadcn/Radix | Только `className`, поведение не трогать |
| 7 | Информация только цветом (точка без подписи) | Недоступно при цветовой слепоте (WCAG 1.4.1) | Цвет + иконка/текст |
| 8 | `outline: none` без замены | Пропадает focus, невозможна клавиатурная навигация | Оставить `focus-visible` ring/outline |
| 9 | Несогласованные компоненты (свой div вместо `EmptyState`) | Визуальный рассинхрон между модулями | Переиспользовать примитивы `components/ui/*` |
| 10 | Жёсткий `blue` вместо `--primary` для primary CTA | Игнорирует выбранный accent пользователя | `variant="default"` / `bg-primary` |
| 11 | Дублирование `dark:`-хардкода поверх токена | Расходится при смене палитры | Использовать токен с готовыми light/dark |
| 12 | Спека без состояний/edge-cases | `fe` домысливает, непредсказуемый результат | Полная спека по §6.2 |
| 13 | Отключение анимаций `!important` (ломает reduced-motion) | Конфликт с `prefers-reduced-motion` политикой | Полагаться на глобальный media-query |
| 14 | Новый компонент вместо композиции существующих | Раздувает библиотеку, рассинхрон | Проверить `components/ui/*`, композировать |
| 15 | Цели нажатия < 44px на mobile | Промахи, недоступно для touch | ≥ 44×44px, зазоры между целями |

### Финальный чек-лист design.md

- [ ] Все цвета/поверхности/границы/радиусы/тени — токены из `app.css`.
- [ ] Shadcn расширен только `className`/`cva`; Radix не тронут; нет `style`/`!important`.
- [ ] 5 состояний + loading/empty/error для каждого интерактивного экрана.
- [ ] WCAG AA: контраст, focus, не-только-цвет, keyboard, цели нажатия, aria.
- [ ] Спека полная (8 разделов), верифицирована `designer-ver`.
- [ ] Регрессия green при изменении кода; docs-sync выполнен; handoff передан.

---

# Приложение A. Полный пример DESIGN-SPEC (эталон)

Ниже — эталонная дизайн-спека реального экрана. Используйте её как шаблон: копируйте структуру,
заменяйте содержимое. Все токены — реальные из `app.css`.

```
# DESIGN-SPEC: Issue Detail Panel (панель задачи трекера)

## 1. Контекст
- US: US-M03 (управление задачами), US-M12 (таймтрекинг), US-S04 (GitHub-линки).
- Персона: Алексей (Developer). Открывает панель из Kanban-доски кликом по карточке
  или по hotkey Enter на выделенной карточке. Читает описание, меняет статус,
  логирует время, оставляет комментарий.
- Точка входа: роут /_authenticated/projects/$projectId/issues/$issueId
  (открывается как Sheet справа поверх доски, доска остаётся видна слева).
- Точка выхода: Esc или клик вне панели закрывает Sheet, фокус возвращается на карточку.

## 2. Структура (layout)

+----------------------------------------------------+
| Sheet (right, w-[560px], bg-card, shadow-xl, --r4) |
| +------------------------------------------------+ |
| | Header (sticky top, border-b b1)               | |
| |  [HUB-123 t3 mono]        [Status Select] [X]  | |
| |  [H2 text-xl t1: заголовок задачи]             | |
| +------------------------------------------------+ |
| | Body (ScrollArea, gap-6, p-6)                  | |
| |  Meta row: [Assignee Avatar][Priority Badge]   | |
| |            [Type Badge][Estimate]              | |
| |  Description (prose, text-sm t2)               | |
| |  Checklist (US-M03)                            | |
| |  Time tracking (US-M12): [Timer btn][Manual]   | |
| |  Linked PRs (US-S04): [GitHub badge list]      | |
| |  Comments (thread): [Avatar][text][actions]    | |
| +------------------------------------------------+ |
| | Footer (sticky bottom, border-t b1)            | |
| |  [Comment Input h-42]              [Send pri.] | |
| +------------------------------------------------+ |
+----------------------------------------------------+

Дерево компонентов:
<Sheet side="right">                         // Radix Dialog, поведение не трогать
  <SheetContent className="w-[560px] p-0">
    <header className="sticky top-0 border-b border-b1 bg-card p-6">
      <span className="font-mono text-xs text-t3">HUB-123</span>
      <StatusSelect />                        // <Select>, cva-варианты по статусу
      <SheetClose />                          // Radix, Esc + click-outside из коробки
      <h2 className="text-xl font-semibold text-t1">{title}</h2>
    </header>
    <ScrollArea className="flex-1">
      <div className="flex flex-col gap-6 p-6">
        <MetaRow />                           // Avatar + Badge (priority/type)
        <Description />                        // prose text-sm text-t2
        <Checklist />
        <TimeTracker />                        // Timer + manual input
        <LinkedPRs />                          // Badge variant="default" (blue glow)
        <CommentThread />
      </div>
    </ScrollArea>
    <footer className="sticky bottom-0 border-t border-b1 bg-card p-4">
      <div className="flex gap-2">
        <Input placeholder="Комментарий..." className="flex-1" />
        <Button variant="default">Отправить</Button>
      </div>
    </footer>
  </SheetContent>
</Sheet>

Breakpoints:
- base (mobile): Sheet становится side="bottom", высота 90vh, w-full.
- md и выше: side="right", w-[560px].
- Доска слева скрывается на mobile (Sheet поверх всего экрана).

## 3. Токены
Поверхности:
- Sheet-контейнер: bg-card (#191a20 dark), shadow-[var(--sh4)].
- Header/Footer: bg-card + border-b/border-t border-b1.
- Комментарий-инпут: bg-[var(--input)].
Текст:
- Ключ задачи HUB-123: text-t3 font-mono.
- Заголовок: text-t1 font-semibold text-xl.
- Описание/комментарии: text-t2 text-sm.
- Метаданные (дата, автор): text-t3 text-xs.
Границы/радиусы/тени:
- Разделители header/footer: border-b1.
- Радиус контейнера: --r4; бейджи: rounded-full; инпут/кнопки: --r2.
- Тень Sheet: --sh4.
Акцент:
- Кнопка "Отправить": variant="default" (--primary, accent-aware).
- Бейдж статуса приоритета critical: --red; high: --orange.
- Linked PR badge: variant="default" (--gB blue glow).

## 4. Состояния
Status Select:
- default: bg-raised border-b1 text-t2.
- hover: border-b2 text-t1.
- focus: ring-2 ring-ring ring-offset-2.
- open: Radix data-state=open (не трогать), поповер bg-card shadow-md animate-dropdown-in.
- disabled (нет прав): opacity-50 pointer-events-none.
Send button: default/hover(-translate-y-px+glow)/active(scale .96)/focus(ring)/disabled(opacity-50).
Data:
- loading: скелетоны meta+description (animate-shimmer), 3 строки комментариев-скелетонов.
- empty comments: <EmptyState icon={MessageSquare} title="Пока нет комментариев"
  description="Будьте первым" />.
- error: <InlineError message="Не удалось загрузить задачу. request_id: {id}" onRetry={refetch} />.

## 5. Поведение
- Смена статуса: оптимистичное обновление (TanStack Query), при ошибке — откат + toast.
- Таймер: Start/Stop пишет time-log; ручной ввод — модалка с полем длительности + описание.
- Комментарий: Enter отправляет, Shift+Enter — новая строка; после успеха — очистка + прокрутка вниз.
- Optimistic lock (version): при конфликте 409 — toast "Задача изменена, обновите" + refetch.
- Esc закрывает Sheet (Radix), фокус возвращается на исходную карточку доски.
- Hotkeys: 'c' фокус на комментарий, 's' открыть status select.

## 6. Доступность
Контраст (dark):
- t1 #f2f3f5 на card #191a20 => ~13:1 (PASS текст).
- t2 #9da0a8 на card #191a20 => ~5.6:1 (PASS текст).
- t3 #a1a7b3 на card => ~6.4:1 (PASS, но только метаданные).
- primary-foreground #fff на --primary blue => >4.5:1 (PASS).
Клавиатура:
- Tab: status -> close -> body-интерактивы -> comment input -> send.
- Sheet focus-trap (Radix), Esc закрывает, возврат фокуса на триггер.
- Timer/Send: Enter/Space активируют.
Aria:
- SheetContent role=dialog aria-labelledby={заголовок}.
- Close button aria-label="Закрыть панель задачи".
- Toast/ошибки в aria-live="polite".
- Статус: не только цвет — бейдж с текстом статуса.
Цели нажатия: close/timer/send >= 40px desktop, 44px mobile.

## 7. Edge-cases
- Очень длинный заголовок: line-clamp-2, полный — в title-атрибуте.
- Задача без описания: плейсхолдер "Нет описания" text-t3.
- 0 / 1 / 200 комментариев: виртуализация списка при >50.
- Нет прав на изменение (Viewer): status select и send скрыты/disabled, панель read-only.
- Оффлайн: баннер "Нет соединения", действия disabled, комментарий сохраняется в очередь.
- Удалённая задача (404): EmptyState "Задача не найдена" + кнопка "К доске".

## 8. Разбивка для fe
- fe1: Sheet-контейнер, Header, StatusSelect, Footer, comment input/send, состояния data.
- fe2: Body-секции (MetaRow, Description, Checklist, TimeTracker, LinkedPRs, CommentThread).
- Точка синхронизации: контракт пропсов IssueDetailProps (types.ts) фиксируется первым коммитом,
  оба берут типы оттуда; разные файлы — конфликтов нет.
```

---

# Приложение B. Полный каталог токенов `app.css`

Справочник всех переменных темы с примером использования в коде. `designer` ссылается на эти имена,
`fe` пишет ровно их. Значения — dark-тема (основная); light см. `:root` в `app.css`.
## B.1 Поверхности и сайдбар

| Токен | Dark | Пример в коде |
|---|---|---|
| `--void` | `#08090b` | `className="bg-void"` |
| `--base` | `#0d0e12` | `className="bg-base"` / `bg-background` |
| `--raised` | `#141519` | `className="bg-raised"` |
| `--card` | `#191a20` | `className="bg-card"` |
| `--input` | `#1e2028` | `className="bg-[var(--input)]"` |
| `--sidebar` | `#0b0c0f` | `className="bg-sidebar"` |
| `--sidebar-t1` | `#f2f3f5` | `className="text-sidebar-foreground"` |
| `--sidebar-t2` | `#9da0a8` | `className="text-sidebar-muted"` |
| `--sidebar-t3` | `#a1a7b3` | `className="text-sidebar-subtle"` |
| `--hover` | `rgba(255,255,255,.032)` | `className="hover:bg-[var(--hover)]"` |
| `--active` | `rgba(255,255,255,.058)` | `className="active:bg-[var(--active)]"` |
## B.2 Текст и границы

| Токен | Dark | Пример |
|---|---|---|
| `--t1` | `#f2f3f5` | `text-t1` / `text-foreground` |
| `--t2` | `#9da0a8` | `text-t2` |
| `--t3` | `#a1a7b3` | `text-t3` |
| `--t4` | `#8b919c` | `text-t4` |
| `--b1` | `rgba(255,255,255,.045)` | `border-b1` / `border-border` |
| `--b2` | `rgba(255,255,255,.08)` | `border-b2` / `hover:border-b2` |
| `--b3` | `rgba(255,255,255,.13)` | `border-b3` (акцентные разделители) |
## B.3 Семантические цвета и glow

| Токен | Dark | Пример |
|---|---|---|
| `--blue` / `--blueS` | `#5b9cf5` / `#3b7ee0` | `text-[var(--blue)]`, градиенты кнопок |
| `--purple` / `--purpleS` | `#a78bfa` / `#8b6ff0` | TMS-акцент, warning-бейдж |
| `--green` / `--greenS` | `#34d399` / `#22c088` | success/passed |
| `--amber` | `#fbbf24` | warning, greeting-градиент |
| `--red` / `--redS` | `#f87171` / `#ef4444` | destructive/failed/critical |
| `--orange` | `#fb923c` | high-priority, broken |
| `--gray` | `#6b7280` | neutral/skipped |
| `--gB` | `rgba(91,156,245,.15)` | `bg-[var(--gB)]` info-бейдж |
| `--gP` | `rgba(167,139,250,.12)` | `bg-[var(--gP)]` warning-бейдж |
| `--gG` | `rgba(52,211,153,.12)` | `bg-[var(--gG)]` success-бейдж |
| `--gR` | `rgba(248,113,113,.1)` | `bg-[var(--gR)]` destructive-бейдж |
## B.4 Радиусы, тени, easing, z-index

| Токен | Значение | Пример |
|---|---|---|
| `--r1..--r4` | 6/8/12/16px | `rounded-[var(--r2)]` |
| `--pill` | 100px | `rounded-full` |
| `--sh1..--sh4` | тени по нарастанию | `shadow-sm/md/lg/xl` |
| `--ease` / `--spring` / `--bounce` | кривые | `ease-[var(--spring)]` |
| `--z-dropdown` | 50 | dropdown/поповер |
| `--z-sticky` | 100 | sticky-хедеры |
| `--z-overlay` | 200 | оверлеи |
| `--z-modal` | 300 | модалки |
| `--z-toast` | 400 | тосты |
| `--z-tooltip` | 500 | тултипы |
| `--z-noise` | 9999 | шумовой оверлей body |

> Правило z-index: используйте семантические токены (`var(--z-modal)`), а не магические числа.
> Порядок наложения: dropdown < sticky < overlay < modal < toast < tooltip.

---

# Приложение C. Рецепты по компонентам Shadcn

Для каждого частого примитива — как правильно расширять и типичный good/bad.
## C.1 Dialog (модальная форма)

**Назначение:** формы, требующие фокуса пользователя (создание задачи, приглашение).

**Good:**

```tsx
<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent className="max-w-lg">
    <DialogHeader>
      <DialogTitle className="text-xl text-t1">Новая задача</DialogTitle>
      <DialogDescription className="text-t2">Заполните поля ниже</DialogDescription>
    </DialogHeader>
    <form className="flex flex-col gap-4">
      <FormField label="Заголовок"><Input /></FormField>
      <FormField label="Описание"><Textarea /></FormField>
    </form>
    <DialogFooter>
      <Button variant="ghost" onClick={() => setOpen(false)}>Отмена</Button>
      <Button variant="default" type="submit">Создать</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

**Bad:** переопределять `onEscapeKeyDown`/`onPointerDownOutside` без причины (ломает ожидаемое закрытие),
или ставить `!p-0 !bg-black`. Отмена не должна быть `type="submit"`.
## C.2 Select (выпадающий список)

**Good:** статусный select с cva-вариантами цвета по значению, только `className`. Radix управляет
клавиатурой (стрелки, typeahead) — не дублируйте `onKeyDown`.

**Bad:** самодельный `<div>`-дропдаун вместо `<Select>` — теряется клавиатура и a11y.
## C.3 DropdownMenu (меню действий)

**Good:** меню «...» действий строки таблицы; пункты `DropdownMenuItem`, деструктивные — `text-[var(--red)]`;
разделитель `DropdownMenuSeparator`. `align="end"`.

**Bad:** popover c ручными кнопками вместо `DropdownMenu` — нет roving-tabindex и Esc.
## C.4 Tooltip

**Good:** тултип на icon-only кнопке дополняет (не заменяет) `aria-label`. Задержка появления из Radix.

**Bad:** класть в тултип критичную информацию, доступную только по hover (недоступно с клавиатуры/touch).
## C.5 Command (Ctrl+K)

**Good:** глобальный поиск US-M07 через `Command`; группы результатов (Задачи/Люди/Проекты/Файлы),
user-scope фильтрация на бэке; Enter открывает, стрелки навигируют.

**Bad:** рендерить все сущности без серверной user-scope фильтрации (нарушение NF-03) и без виртуализации.
## C.6 Sheet / Popover

**Good:** `Sheet` для боковых панелей (детали задачи, фильтры), `Popover` для лёгких контекстных
поповеров (выбор даты, emoji). На mobile Sheet — `side="bottom"`.

**Bad:** использовать `Dialog` там, где нужен неблокирующий `Popover`, — лишний оверлей и focus-trap.
## C.7 Форма (FormField + Label + Input/Textarea)

**Good:** каждое поле в `FormField` c `<Label>`; ошибка под полем текстом + красная граница; кнопка
submit `type="submit"`, отмена `type="button"`.

**Bad:** placeholder вместо label (исчезает при вводе, недоступно screen-reader); ошибка только цветом рамки.

### Чек-лист рецептов (Прил. C)

- [ ] Выбран правильный примитив под задачу (Dialog vs Sheet vs Popover vs DropdownMenu).
- [ ] Расширение только `className`/`cva`; Radix-поведение сохранено.
- [ ] Форма: label + сообщение об ошибке текстом + корректные `type`.
- [ ] Icon-only кнопки имеют `aria-label` (+ Tooltip как дополнение).

---

# Приложение D. Каталог анимаций и когда применять

Все ниже — из `app.css`. Используйте готовые утилиты, не пишите новые keyframes без нужды.

| Утилита / keyframe | Длительность | Когда применять | Когда НЕ применять |
|---|---|---|---|
| `animate-fade-in` | 0.3s | Появление контента блока | На каждый ре-рендер списка |
| `animate-fade-in-up` | 0.4s | Каскадный вход карточек | На текст, который читают |
| `animate-shimmer` | 1.5s loop | Скелетоны загрузки | После загрузки (убрать) |
| `animate-dropdown-in` | 150ms | Поповеры/дропдауны | На модалки (у них своя) |
| `animate-scale-in` | 200ms | Бейдж уведомлений | На крупные блоки |
| `animate-pulse-online` | 2s loop | Онлайн-индикатор presence | На статичные точки |
| `animate-speaking` | 1s loop | Ring говорящего в звонке | Когда участник молчит |
| `animate-blink` | 1s loop | REC-индикатор записи | На обычный контент |
| `kanban-enter` | 0.5s | Вход карточек доски | На drag (там свои transform) |
| `dash-enter` (+2/+3) | 0.4s каскад | Виджеты дашборда | На частые обновления виджета |
| `msg-enter-x` | 0.4s | Вход элементов списка каналов | На каждое новое сообщение |
| `login-form-animate` | 0.5s | Форма логина | Вне логина |

**Правила анимаций:**

- Всегда уважать `prefers-reduced-motion` (глобальный media-query уже гасит всё).
- Не зацикливать анимации в зоне чтения/ввода.
- Длительность интеракций ≤ 200ms; вход контента ≤ 500ms.
- Не анимировать `width/height/top/left` (рефлоу) — предпочитать `transform`/`opacity`.

---

# Приложение E. A11y-рецепты по паттернам
## E.1 Форма

- Каждое поле: `<Label htmlFor>` + `<Input id>`; обязательные — с `aria-required`.
- Ошибка: `aria-invalid` + `aria-describedby` на текст ошибки; текст, не только цвет.
- Группировка связанных полей: `<fieldset>` + `<legend>`.
## E.2 Модалка/Sheet

- `role="dialog"` (Radix), `aria-labelledby` на заголовок, `aria-describedby` на описание.
- Focus-trap внутри, Esc закрывает, фокус возвращается на триггер (Radix из коробки — не ломать).
## E.3 Таблица/список

- Семантическая `<table>` c `<th scope>` или ARIA grid для интерактивных.
- Навигация ↑/↓; выделенная строка — `aria-selected`; сортировка — `aria-sort`.
- Виртуализация не должна ломать порядок чтения screen-reader.
## E.4 Меню/дропдаун

- `DropdownMenu` даёт `role=menu`/`menuitem`, roving tabindex, typeahead — не дублировать вручную.
- Деструктивные действия визуально (цвет + иконка) и по тексту отличимы.
## E.5 Тосты/уведомления

- Контейнер тостов — `aria-live="polite"` (ошибки — `assertive`).
- Тост не исчезает слишком быстро для чтения; критичный — с кнопкой действия/закрытия.

### Чек-лист a11y-рецептов (Прил. E)

- [ ] Формы: label+ошибка текстом+aria-invalid.
- [ ] Модалки: labelledby+focus-trap+Esc+возврат фокуса.
- [ ] Списки/таблицы: семантика+клавиатура+aria-selected/sort.
- [ ] Меню: нативное поведение Radix сохранено.
- [ ] Тосты: aria-live, достаточное время, действие.

---

# Приложение F. Responsive-рецепты по модулям

| Модуль | Mobile (base) | Tablet (`md:`) | Desktop (`lg:`+) |
|---|---|---|---|
| Tracker/Kanban | Одна колонка, свайп между статусами, FAB «+» | 2–3 колонки со скроллом | Все колонки, `board-page-root` полная высота, drag&drop |
| Messenger | Список каналов ИЛИ чат (стек), назад-навигация | Список + чат две панели | + панель тредов/деталей справа |
| TMS | Список кейсов карточками | Таблица кейсов | Таблица + панель прогона |
| Calls | Одно активное видео + свайп участников | Сетка 2×2 | Сетка N×N + панель чата/участников |
| Dashboard | Виджеты в один столбец | 2 колонки | Drag&drop сетка виджетов |
| Detail (задача/кейс) | `Sheet side="bottom"` 90vh | `Sheet side="right"` w-[480px] | `Sheet side="right"` w-[560px] |

**Общие правила:**

- Навигация: bottom navigation на mobile (US-C05), сайдбар на desktop.
- Таблицы: на mobile — карточки или горизонтальный скролл, никогда не «сжатая» таблица.
- Модалки → `Sheet side="bottom"` на mobile для удобства большого пальца.
- Тестировать на 360px (узкий телефон), 768px (планшет), 1280px (ноутбук), 1536px+ (широкий).

---

# Приложение G. Матрица контраста (расширенная, dark)

Ориентировочные соотношения для типовых пар. `designer` пересчитывает точные значения для новых пар.

| Текст | Фон | Контраст (≈) | Вердикт |
|---|---|---|---|
| `--t1 #f2f3f5` | `--base #0d0e12` | ~15:1 | PASS (любой текст) |
| `--t1 #f2f3f5` | `--card #191a20` | ~13:1 | PASS (любой текст) |
| `--t1 #f2f3f5` | `--raised #141519` | ~14:1 | PASS |
| `--t2 #9da0a8` | `--base #0d0e12` | ~6.5:1 | PASS (обычный текст) |
| `--t2 #9da0a8` | `--card #191a20` | ~5.6:1 | PASS (обычный текст) |
| `--t3 #a1a7b3` | `--card #191a20` | ~6.4:1 | PASS, но только метаданные/placeholder |
| `--t4 #8b919c` | `--card #191a20` | ~4.7:1 | граничный — только крупный/декоративный |
| `--blue #5b9cf5` | `--base #0d0e12` | ~6:1 | PASS для ссылок/иконок |
| `#ffffff` | `--primary (blue #5b9cf5)` | ~2.8:1 | ВНИМАНИЕ: primary-foreground в dark = `#111`, не белый |

> Важный вывод из таблицы: в dark-теме кнопки accent используют `--primary-foreground: #111` (тёмный
> текст на светлом accent), а не белый. Не задавайте `text-white` на accent-кнопке в dark вручную.

### Финальные напоминания design.md

- Токены — единственный источник значений. HEX в компоненте = баг.
- Shadcn/Radix: `className`/`cva` only; поведение не трогать; без `style`/`!important`.
- 5 состояний + loading/empty/error всегда.
- WCAG AA: контраст, focus, не-только-цвет, keyboard, 44px, aria.
- Спека полная (8 разделов) → `designer-ver` `verified: true` → `fe` → `fe-ver` → `team-lead`.

---

# Приложение H. Модульные акценты, градиенты и спец-классы

`app.css` содержит готовые модульные утилиты. Не изобретайте новые — используйте эти.
## H.1 Градиенты и «gradient text»

| Класс | Что делает | Где применять |
|---|---|---|
| `.gradient-tms` | `linear-gradient(135deg, #a78bfa, #8b6ff0)` | Акцент модуля TMS |
| `.gradient-text` | Blue→Purple→Red клип на текст | Логотип, ключевые заголовки |
| `.greeting-gradient` | `--t1 → --t3` клип | Приветствие |
| `.dash-greeting-gradient` | amber→orange→pink | Приветствие дашборда (DASH-02) |

**Правило:** gradient-text применять точечно (логотип, 1 заголовок), не на body-текст — страдает контраст
и читаемость.
## H.2 Карточные и layout-утилиты

| Класс | Назначение |
|---|---|
| `.kanban-card` (+`-critical`/`-high`/`-done`) | Карточка задачи: градиент фон, hover-подъём, priority-бордер слева |
| `.kanban-fchip` | Фильтр-чип на доске (pill, hover-подъём) |
| `.dash-widget` | Виджет дашборда (r4, верхний хайлайт, hover) |
| `.dash-qa-btn` | Кнопка быстрого действия дашборда |
| `.admin-kpi` | KPI-карточка админки (верхний градиент-хайлайт) |
| `.tms-run-card` / `.tms-precond-card` | Карточки TMS с верхним хайлайтом |
| `.page-header-blur` | Sticky-хедер с backdrop-blur |
| `.glass` | Стеклянный оверлей (login, модалки) |
| `.board-page-root` | Полная высота Kanban-доски |
| `.scrollbar-thin` | Тонкий скроллбар для колонок |
## H.3 Glow-утилиты

| Класс | Назначение |
|---|---|
| `.btn-glow` | Свечение primary-кнопки (уже в variant=default) |
| `.logo-glow` | Свечение логотип-бейджа |
| `.hover-pulse` | Пульс CTA при hover |
## H.4 Статусные бейджи (TMS/трекер)

Классы `.badge-passed/.badge-failed/.badge-skipped/.badge-broken/.badge-blocked` имеют отдельные значения
для light и dark (light — насыщенные HEX, dark — glow-токены). Это осознанное исключение из «только
токены»: статусы тестов имеют фиксированную семантику цвета. Новые статусы добавляйте тем же паттерном
(`.badge-<name>` + `.dark .badge-<name>`), а не inline.

### Good / Bad — модульные классы

**Good:** карточка задачи = `<div className="kanban-card kanban-card-high">` — весь визуал и hover уже
описаны, priority-бордер оранжевый через `--orange`.

**Bad:** вручную собирать градиент, hover и бордер карточки в компоненте — дубль `app.css`, рассинхрон.

---

# Приложение I. Глоссарий и FAQ
## I.1 Глоссарий

| Термин | Значение |
|---|---|
| Токен | CSS-переменная дизайн-системы (`--card`, `--t1`), единственный источник значений |
| Surface | Поверхность-фон по «высоте»: void<base<raised<card<input |
| Accent | Пользовательский primary-цвет (`data-accent`), подхватывается через `--primary` |
| cva | class-variance-authority — декларативные варианты компонента |
| Glow | Полупрозрачная цветная заливка/свечение (`--gB/gP/gG/gR`) |
| Дизайн-спека | Markdown-контракт экрана (8 разделов), заменяет Figma |
| DoD | Definition of Done — критерии готовности артефакта |
| a11y | Accessibility, целевой уровень WCAG 2.2 AA (2026: переход с 2.1) |
## I.2 FAQ

**Нужен новый цвет — что делать?** Добавить токен в `app.css` (`:root` + `.dark`), затем использовать его.
Не хардкодить HEX в компоненте.

**Shadcn-компонент выглядит не так, как в спеке — можно поправить пропсы Radix?** Нет. Только `className`
или новый `cva`-вариант. Если нужен другой поведенческий контракт — это задача для `arch`/`fe` с
согласованием, а не правка на месте.

**Как показать состояние загрузки?** Скелетоны `animate-shimmer` в форме контента. Полноэкранный спиннер —
только для самой первой загрузки приложения.

**Placeholder вместо label — ок?** Нет. Label обязателен (может быть `sr-only`), placeholder — подсказка,
не замена подписи.

**Как проверить контраст?** Рассчитать соотношение для пары «текст/фон» по значениям из `app.css`
(см. Приложение G) и вписать в спеку. `designer-ver` перепроверяет.

**Тёмная или светлая тема — что основное?** Тёмная (макеты рисовались тёмными). Но обе обязаны работать —
используйте токены с готовыми значениями в обоих режимах.

**Можно `!important`?** Нет. Phase 8 вычистил 56 случаев. Новые запрещены (`60-tz-constraints.mdc`).

### Быстрый предполётный чек-лист (перед handoff в `fe`)

- [ ] Открыл `app.css` и сверил все токены спеки.
- [ ] Проверил `components/ui/*` — не изобретаю существующее.
- [ ] 5 состояний + loading/empty/error расписаны.
- [ ] Контраст посчитан, keyboard/aria/44px заданы.
- [ ] Breakpoints и mobile-стратегия есть.
- [ ] Edge-cases + RBAC-видимость перечислены.
- [ ] Разбивка fe1/fe2 без пересечения файлов.
- [ ] Нет антипаттернов из §9.

---

# Приложение J. Онбординг нового дизайн-агента (первые 5 шагов)

Если ты `designer`/`designer-ver`/`fe` и впервые открываешь дизайн-задачу Hub — выполни по порядку:

1. **Прочитай токены.** Открой `frontend/src/app.css` целиком. Запомни поверхности (void→input),
   текст (t1→t4), границы (b1→b3), радиусы, тени, семантические цвета и accent-механику.
2. **Прочитай правила.** `.cursor/rules/30-frontend.mdc` (дизайн-система, Shadcn), `60-tz-constraints.mdc`
   (запреты: `!important`, inline, замена UI-фреймворка).
3. **Изучи примитивы.** Пробеги `frontend/src/components/ui/*` — что уже есть (Button/Card/Dialog/… ),
   как заданы cva-варианты, как расширяются `className`.
4. **Пойми аудиторию.** `docs/vision_and_scope.md` — персоны (Алексей-dev, Мария-QA, Дмитрий-lead),
   плотность и keyboard-first для профи-команд 10–50.
5. **Сверься с этим playbook.** Разделы §2 (токены), §5 (a11y), §6 (спека), §9 (антипаттерны) — держи
   открытыми во время работы.

**Правило одной страницы:** прежде чем предложить любой пиксель, ты должен уметь назвать токен, из
которого он берётся. Если токена нет — сначала обсуди добавление токена, потом рисуй.

---

> Связанные документы: `.cursor/rules/30-frontend.mdc`, `.cursor/rules/60-tz-constraints.mdc`,
> `.cursor/playbooks/team-workflow.md`, `frontend/src/app.css`, `frontend/src/components/ui/*`,
> `docs/vision_and_scope.md`, `docs/wiki/SHORTCUTS.md`.
