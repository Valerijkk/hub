# Phase 8: Design System Foundation - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Приложение переходит на новую дизайн-систему — все поверхности, текст, границы, тени, радиусы, анимации и шрифты соответствуют новой палитре из HTML-макетов в папке `редизайн/`, а тёмная и светлая темы работают корректно. Это фундамент для всех последующих фаз редизайна.

Requirements: DSFN-01..09, THME-01..04

</domain>

<decisions>
## Implementation Decisions

### Font Loading Strategy
- Google Fonts CDN (текущий паттерн, внутренний инструмент 10-50 юзеров)
- Inter: веса 400, 500, 600, 700, 800, 900
- JetBrains Mono: веса 400, 600
- font-display: swap — текст виден сразу, шрифт подменяется

### Light Theme Derivation
- Инверсия: белые поверхности (#FFFFFF → #F8F9FA), серые текст/бордеры, те же акценты
- Light mode shadows: уменьшенная opacity (0.05-0.15 vs 0.4-0.5 в dark)
- Light mode gradients на картах: убрать/минимизировать — subtler на белом фоне
- Определить light токены в Phase 8, полный visual QA в Phase 13

### CSS Variable Migration
- Dual mapping: новые токены (--void, --base, --raised, --card, --input, --t1-t4, --b1-b3) + совместимость с Shadcn namespace (--background, --foreground, --primary, etc.)
- Noise texture overlay: глобальный body::after как в макетах — subtle grain effect
- Easing variables: CSS custom properties в :root (--ease, --spring, --bounce)
- !important cleanup: удалить все 56 деклараций сразу, заменить на scoped cn() в компонентах где нужно

### Claude's Discretion
- Точные light mode hex-значения для поверхностей и текста (best-guess inversion)
- z-index scale tokens
- Scrollbar стилизация
- Конкретная реализация noise overlay (SVG filter vs CSS)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `frontend/src/app.css` — единственный файл для design tokens (1053 строки), содержит :root и .dark блоки с HSL переменными
- `frontend/index.html` — Google Fonts CDN link (текущий: Geist)
- `frontend/src/lib/utils.ts` — cn() utility для class merging
- `frontend/src/components/theme-provider.tsx` — next-themes integration (light/dark/system)
- `frontend/src/components/accent-provider.tsx` — data-accent attribute для accent colors

### Established Patterns
- Tailwind CSS v4 с @tailwindcss/vite plugin
- @theme inline в app.css для Tailwind token generation
- CSS variables в HSL формате для Shadcn/ui compatibility
- next-themes с class attribute на html element
- AccentColor context для user accent selection

### Integration Points
- Все 15+ Shadcn/ui компонентов в src/components/ui/ ссылаются на CSS variables
- AppShell, Sidebar, Topbar используют bg-sidebar, text-sidebar-foreground etc.
- Tailwind утилиты генерируются из @theme inline → CSS variables

### Key Files to Modify
- `frontend/src/app.css` — основной файл (CSS variables, @theme inline, custom utilities)
- `frontend/index.html` — Google Fonts link swap
- `frontend/src/components/theme-provider.tsx` — если нужны изменения в ThemeProvider
- `frontend/src/components/accent-provider.tsx` — если нужны изменения в акцентах

</code_context>

<specifics>
## Specific Ideas

- Макеты в папке `редизайн/` — единственный источник правды для дизайн-токенов
- Все CSS variables из макетов: --void (#08090B), --base (#0D0E12), --raised (#141519), --card (#191A20), --input (#1E2028)
- Текст: --t1 (#F2F3F5), --t2 (#8C8F97), --t3 (#4E5058), --t4 (#363840)
- Бордеры: --b1 (rgba 255,255,255,.045), --b2 (.08), --b3 (.13)
- Акценты: --blue (#5B9CF5), --purple (#A78BFA), --green (#34D399), --amber (#FBBF24), --red (#F87171)
- Радиусы: --r1 (6px), --r2 (8px), --r3 (12px), --r4 (16px), --pill (100px)
- Тени: --sh1 through --sh4 (4 уровня)
- Шрифты: --sans (Inter), --mono (JetBrains Mono)
- Easing: --ease (cubic-bezier(.25,.1,0,1)), --spring (cubic-bezier(.175,.885,.32,1.1)), --bounce (cubic-bezier(.34,1.56,.64,1))

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>
