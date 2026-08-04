# Дизайн: токены и тема

> Часть плейбука. Канон и оглавление — [`../design.md`](../design.md).

Переменные, палитра, типографика, отступы, тени, анимации, состояния.

---

## 2.1 Два имени одного и того же: `--hub-*` и реальные токены

В правиле `.cursor/rules/30-frontend.mdc` дизайн-система описана **семантическими** именами вида
`--hub-surface-raised`, `--hub-text-primary`, `--hub-border`. Это удобная для чтения ментальная модель.

В реальном коде (`frontend/src/app.css`) те же токены реализованы **краткими** именами: `--raised`,
`--t1`, `--b1`. Обе нотации описывают одну систему. Таблица соответствия — источник истины при написании
спеки и при верстке.

> **Правило для `designer`:** в спеке допускается писать семантические `--hub-*` имена для читабельности,
> но **обязательно** рядом указывать реальный токен из `app.css`. `fe` в коде использует **только реальные**
> токены/утилиты (`bg-raised`, `var(--raised)`, `text-t1`), потому что именно они существуют в теме.

### Таблица соответствия — Surfaces (поверхности)

| Семантика (`--hub-*`) | Реальный токен | Utility-класс | Light | Dark | Назначение |
|---|---|---|---|---|---|
| `--hub-surface-void` | `--void` | `bg-void` | `#ffffff` | `#08090b` | Самый глубокий фон (за всем) |
| `--hub-surface-base` | `--base` | `bg-base` / `bg-background` | `#f8f9fa` | `#0d0e12` | Базовый фон приложения (body) |
| `--hub-surface-raised` | `--raised` | `bg-raised` | `#f1f3f5` | `#141519` | Приподнятые блоки, secondary, muted |
| `--hub-surface-card` | `--card` | `bg-card` | `#ffffff` | `#191a20` | Карточки, поповеры-контент |
| `--hub-surface-input` | `--input` | `bg-[var(--input)]` | `#f1f3f5` | `#1e2028` | Поля ввода |
| — | `--sidebar` | `bg-sidebar` | `#1a1d21` | `#0b0c0f` | Сайдбар (всегда тёмный) |

### Таблица соответствия — Text (текст)

| Семантика (`--hub-*`) | Реальный токен | Utility-класс | Light | Dark | Назначение |
|---|---|---|---|---|---|
| `--hub-text-primary` | `--t1` | `text-t1` / `text-foreground` | `#111827` | `#f2f3f5` | Основной текст, заголовки |
| `--hub-text-secondary` | `--t2` | `text-t2` | `#4b5563` | `#9da0a8` | Вторичный текст, подписи |
| `--hub-text-muted` | `--t3` | `text-t3` | `#9ca3af` | `#a1a7b3` | Приглушённый, placeholder |
| `--hub-text-disabled` | `--t4` | `text-t4` | `#6b7280` | `#8b919c` | Disabled, наименее важное |

> Внимание: в текущей теме `--t3` в dark (`#a1a7b3`) светлее, чем `--t4` (`#8b919c`). Семантика «muted»
> vs «disabled» задаётся ролью, а не абсолютной яркостью. Для disabled-состояния дополнительно снижают
> `opacity` (см. §2.9).

### Таблица соответствия — Borders (границы)

| Семантика (`--hub-*`) | Реальный токен | Utility-класс | Light | Dark | Назначение |
|---|---|---|---|---|---|
| `--hub-border-subtle` | `--b1` | `border-b1` / `border-border` | `rgba(0,0,0,.06)` | `rgba(255,255,255,.045)` | Тонкая граница по умолчанию |
| `--hub-border` | `--b2` | `border-b2` | `rgba(0,0,0,.10)` | `rgba(255,255,255,.08)` | Граница hover / выделения |
| `--hub-border-strong` | `--b3` | `border-b3` | `rgba(0,0,0,.16)` | `rgba(255,255,255,.13)` | Сильная граница, разделители-акценты |
## 2.2 Цветовая система (семантические цвета)

Семантические цвета заданы отдельно и имеют «сильный» (`S`) вариант для градиентов и hover. Значения
различаются в light/dark для сохранения контраста.

| Токен | Light | Dark | Утилита в коде | Назначение |
|---|---|---|---|---|
| `--blue` / `--blueS` | `#3b7ee0` / `#2563eb` | `#5b9cf5` / `#3b7ee0` | `var(--blue)` | Primary / info, ссылки |
| `--purple` / `--purpleS` | `#8b6ff0` / `#7c3aed` | `#a78bfa` / `#8b6ff0` | `var(--purple)` | TMS, warning-акцент |
| `--green` / `--greenS` | `#22c088` / `#16a34a` | `#34d399` / `#22c088` | `var(--green)` | Success, passed |
| `--amber` | `#d97706` | `#fbbf24` | `var(--amber)` | Warning, greeting |
| `--red` / `--redS` | `#ef4444` / `#dc2626` | `#f87171` / `#ef4444` | `var(--red)` | Destructive, failed, critical |
| `--orange` | `#ea580c` | `#fb923c` | `var(--orange)` | High priority, broken |
| `--gray` | `#6b7280` | `#6b7280` | `var(--gray)` | Neutral, skipped |

**Accent (акцент пользователя).** `--primary` по умолчанию = `--blue`, но переопределяется атрибутом
`data-accent="red|blue|green|purple|orange"` на корне (см. `app.css`, US-C03). Поэтому:

- Для интерактивных primary-элементов используйте `--primary` / `text-primary` / `bg-primary`,
  а не жёстко `--blue`. Тогда элемент подхватит выбранный пользователем accent.
- Для статических семантик (success/error/warning) используйте конкретный цвет (`--green`, `--red`).

### Good / Bad — использование цвета

**Good:**

```tsx
// Primary CTA — подхватывает accent пользователя
<Button variant="default">Создать задачу</Button>

// Статичная семантика ошибки
<span className="text-[var(--red)]">Тест упал</span>
```

**Bad:**

```tsx
// Хардкод HEX вместо токена — не переключится тема/accent
<button style={{ background: "#3b7ee0" }}>Создать</button>

// Primary через жёсткий blue — игнорирует data-accent пользователя
<div className="bg-[#5b9cf5]">...</div>
```
## 2.3 Тёмная тема (source of truth)

Dark-тема в Hub — не «инверсия» light, а **основной** режим (макеты рисовались тёмными). Правила:

- Тема переключается классом `.dark` на корне документа. Токены переопределяются внутри `.dark {}`.
- Никогда не пишите цвет «только для тёмной темы» хардкодом — используйте токен, у которого уже есть
  значение в обоих режимах.
- Поверхности идут по нарастанию «высоты»: `--void` < `--base` < `--raised` < `--card` < `--input`.
  Более «высокая» поверхность — светлее в dark, что создаёт ощущение слоёв без границ.
- Тени в dark сильнее (`--sh*` с высокой alpha), потому что тёмный фон гасит мягкие тени.
- Есть плавный кроссфейд при переключении темы (200ms, `--ease`) — не ломайте его inline-переопределением
  `transition: none`.

### Good / Bad — тёмная тема

**Good:** карточка задаётся `bg-card` — в light станет `#ffffff`, в dark `#191a20` автоматически.

**Bad:** `className="dark:bg-[#191a20] bg-white"` — дублирует то, что уже решает токен `--card`, и
расходится при следующем изменении палитры.
## 2.4 Типографика

| Параметр | Значение | Источник |
|---|---|---|
| Sans-serif | `Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` | `--sans` |
| Monospace | `JetBrains Mono, ui-monospace, monospace` | `--mono` |
| Base font-size | `14px` | `body` в `app.css` |
| Base line-height | `1.5` | `body` |
| Заголовки | `font-semibold tracking-tight` | `@layer base h1..h6` |
| Сглаживание | `antialiased`, `optimizeLegibility` | `body` |

Шкала размеров (Tailwind) — используйте утилиты, а не произвольные `text-[13px]` без причины:

| Утилита | Размер | Применение |
|---|---|---|
| `text-2xl` | 24px | Заголовок карточки (`CardTitle`) |
| `text-xl` | 20px | Заголовок секции |
| `text-lg` | 18px | Подзаголовок |
| `text-base` | 16px | Крупный body (в input по умолчанию, `md:text-sm`) |
| `text-sm` | 14px | Основной body в приложении |
| `text-xs` | 12px | Метаданные, подписи |
| `text-[11px]` | 11px | Бейджи (см. `badge.tsx`) |

**Правила:**

- Вес: `font-bold` для кнопок/бейджей (см. `button.tsx`, `badge.tsx`), `font-semibold` для заголовков,
  `font-medium` для акцентных подписей, обычный — для body.
- `tracking-tight` только на заголовках. Не «сжимайте» body-текст.
- Не более 3 весов на экране. Иерархия строится размером + весом + цветом (`--t1`/`--t2`/`--t3`),
  а не десятком разных кеглей.
## 2.5 Spacing (отступы)

Hub использует Tailwind spacing-шкалу (шаг 4px). Разрешённые значения — по шкале, произвольные
`p-[13px]` запрещены без обоснования.

| Утилита | px | Типичное применение |
|---|---|---|
| `gap-1` / `p-1` | 4 | Иконка + текст вплотную |
| `gap-2` / `p-2` | 8 | Внутри бейджа, мелкие группы |
| `gap-3` / `p-3` | 12 | Поля ввода, кнопки |
| `gap-4` / `p-4` | 16 | Стандартный внутренний отступ карточки (Kanban card = 16px) |
| `p-6` | 24 | `CardHeader` / `CardContent` (Shadcn) |
| `gap-8` / `p-8` | 32 | Разделение секций |
| `py-12` | 48 | Пустые состояния (`EmptyState`) |

**Правила плотности:** Hub — инструмент для профи-команд (10–50 человек), которые проводят в нём часы.
Плотность выше, чем в consumer-приложениях. Стандартная высота строк списков компактна, но не в ущерб
целям нажатия (см. §5.5).
## 2.6 Радиусы

| Токен | Значение | Утилита | Применение |
|---|---|---|---|
| `--r1` | 6px | `rounded-[var(--r1)]` | Мелкие элементы, `sm` |
| `--r2` | 8px | `rounded-[var(--r2)]` | Кнопки, инпуты (base) |
| `--r3` | 12px | `rounded-[var(--r3)]` | Карточки Kanban, средние блоки |
| `--r4` | 16px | `rounded-[var(--r4)]` | Крупные карточки (`Card`), виджеты |
| `--pill` | 100px | `rounded-full` | Бейджи, чипы, аватары |

> Также замаплены на Tailwind: `--radius-sm=--r1`, `--radius-md=--r2`, `--radius-lg=--r3`, `--radius-xl=--r4`.
> В спеке указывайте `--r2`/`--r4`, а не «8px», чтобы `fe` использовал токен.
## 2.7 Тени и glow

| Токен | Light | Dark | Применение |
|---|---|---|---|
| `--sh1` | `0 1px 3px rgba(0,0,0,.08)` | `0 1px 3px rgba(0,0,0,.5)` | Карточка в покое (`shadow-sm`) |
| `--sh2` | мягкая двойная | усиленная | Hover карточки, поповер (`shadow-md`) |
| `--sh3` | `0 12px 40px` | сильная | Kanban hover, модалка (`shadow-lg`) |
| `--sh4` | `0 24px 80px` | максимальная | Крупные overlay (`shadow-xl`) |

**Glow-фоны** (полупрозрачные заливки для «свечения» и бейджей):

| Токен | Назначение |
|---|---|
| `--gB` | Blue glow — info-бейдж, primary hover-свечение |
| `--gP` | Purple glow — TMS, warning-бейдж |
| `--gG` | Green glow — success-бейдж |
| `--gR` | Red glow — destructive-бейдж |

**Специальные glow-утилиты (`app.css`):** `.btn-glow`, `.logo-glow`, `.hover-pulse`. Кнопка `default`
уже несёт `shadow-[0_0_20px_rgba(91,156,245,0.25)]` — не дублируйте glow сверху.

### Good / Bad — тени

**Good:** `<Card>` уже содержит `shadow-sm → hover:shadow-md`. Дизайнер указывает «состояние hover:
подъём на 2px + `--sh2`» — это ровно то, что делает компонент.

**Bad:** спека требует `box-shadow: 0 10px 30px #000` хардкодом — рассинхрон с `--sh*`, тень будет
слишком жёсткой в light-теме.
## 2.8 Easing и анимации

| Токен | Кривая | Применение |
|---|---|---|
| `--ease` | `cubic-bezier(0.25,0.1,0,1)` | Базовые переходы (цвет, фон, граница) |
| `--spring` | `cubic-bezier(0.175,0.885,0.32,1.1)` | Кнопки, карточки, вход элементов |
| `--bounce` | `cubic-bezier(0.34,1.56,0.64,1)` | Игривые микро-взаимодействия |

Готовые keyframes/утилиты в `app.css`: `animate-fade-in`, `animate-fade-in-up`, `animate-shimmer`
(скелетоны), `animate-dropdown-in`, `animate-scale-in`, `animate-pulse-online`, `kanban-enter`,
`dash-enter`, `msg-enter-x`, `login-form-animate`.

**Длительности:** 150ms — hover/цвет; 200ms — кнопки/тема; 260–280ms — карточки; 400–500ms — вход
контента. Дольше 500ms — только для декоративных фонов (login mesh 20s).

**Reduced motion:** `app.css` содержит `@media (prefers-reduced-motion: reduce)`, который гасит все
анимации до 0.01ms. Не переопределяйте это `!important`-ом в компонентах.
## 2.9 Состояния (hover / active / focus / disabled)

Каждый интерактивный элемент обязан описывать 5 состояний. Базовые паттерны из компонентов Hub:

| Состояние | Паттерн (из `button.tsx` / `input.tsx` / `card.tsx`) |
|---|---|
| default | Базовый фон/граница токенами |
| hover | Подъём `hover:-translate-y-px`, усиление тени, `hover:brightness-110`, граница `--b2` |
| active | `active:scale-[.96]` (кнопки), `active:scale-[1.02]` (Kanban card) |
| focus | `focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2`; инпут: `focus-visible:border-[rgba(91,156,245,0.3)]` + мягкий ring-shadow |
| disabled | `disabled:pointer-events-none disabled:opacity-50` |

Глобальный focus-контур задан в `app.css`: `*:focus-visible { outline: 2px solid var(--primary); outline-offset: 2px; }`. Это база доступности — не удаляйте.

### Good / Bad — состояния

**Good:** спека для кастомной кликабельной строки списка: default `bg-transparent`, hover `bg-[var(--hover)]`,
focus — глобальный ring (наследуется), active — `bg-[var(--active)]`, disabled — `opacity-50 + pointer-events-none`.

**Bad:** описан только default и hover. `fe` не знает, как выглядит focus → клавиатурная навигация
невидима → `fe-ver` заворачивает по a11y.

### Чек-лист дизайн-системы (§2)

- [ ] Все цвета/поверхности/границы — токены из `app.css` (нет HEX-хардкода).
- [ ] Primary-элементы используют `--primary` (подхватывают accent), статусы — конкретные семантики.
- [ ] Радиусы из `--r1..--r4`/`--pill`; тени из `--sh1..--sh4`; glow из `--gB/gP/gG/gR`.
- [ ] Типографика: `--sans`/`--mono`, ≤3 веса, иерархия размером+цветом.
- [ ] Spacing по 4px-шкале Tailwind.
- [ ] Для каждого интерактива описаны 5 состояний.
- [ ] Учтён `prefers-reduced-motion`, тема dark — основная.

---

# 3. Компонентная библиотека Shadcn/ui
## 2. Структура (layout)
- ASCII-схема раскладки (см. 6.3)
- Дерево компонентов: какие Shadcn-примитивы и в какой вложенности
- Сетка/колонки, breakpoints поведение
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
