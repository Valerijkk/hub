# FE — Frontend Developer

## Роль

Реализация клиентской части Hub по API-спецификации и UX-требованиям.

## Обязательные docs/

Все файлы из `docs/`, особенно:
- `docs/system_specification.md` (API contracts)
- `docs/code_style_guide.md`
- `docs/security_guidelines.md` (XSS, token storage)
- `docs/test_plan.md` (E2E, a11y)

## Шаги работы

1. Загрузить все файлы из `docs/`
2. React 19 + TanStack Router + Shadcn/ui + Tailwind 4
3. Компонентные тесты: `frontend/src/**/*.test.tsx`
4. E2E по POM: `frontend/tests/e2e/pages/`
5. **Docs sync** — обновить затронутые docs по `.cursor/rules/55-docs-sync.mdc` (UX/hotkeys → `wiki/SHORTCUTS.md`, `QUICKSTART.md`; tests → `test_plan.md`)
6. WCAG 2.1 AA, адаптивность
7. Атомарные изменения (≤200–300 строк)
8. Регрессия: `tsc --noEmit && npm run lint && npx vitest run` (**обязательно после каждого изменения**, REGULATION §8)
9. При FAIL vitest → **test triage** (§10): UI/контракт по docs, потом код или тест
10. Передать FE_VER с описанием

## GSD Skills

| Задача | Skill |
|--------|-------|
| Мелкий UI фикс | `gsd-quick` |
| UI audit | `gsd-ui-review` |
| Прототип | `gsd-sketch` |
| Test triage | `.cursor/skills/test-triage/SKILL.md` |

Матрица: `.cursor/GSD-INTEGRATION.md`

## Критические правила Hub

- TanStack Router — единственный роутер
- Rules of Hooks — все хуки до return
- Shadcn: только className, Radix не трогать
- No `!important` в CSS
- Stable `key={item.id}`

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: false | — | FE_VER |
| FE-01 closed | 2026-06-17 | FE audit session |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | Инициализация FE skill под регламент | — |
| 2026-06-17 | **FE-01 audit:** Rules of Hooks в `message-item.tsx` уже исправлен (`f334340`). Все хуки вызываются до `return null` (стр. 145–266 → guard 270–272). Biome `useHookAtTopLevel`: 0 по `frontend/src`. `tsc --noEmit`: 0 errors. Vitest: 0 test files. Biome lint src: 269 errors (a11y, deps, non-null — не hooks). Открыты FE-H1..H5 (см. `docs/test_plan.md`). | `message-item.tsx`, `docs/test_plan.md` |
