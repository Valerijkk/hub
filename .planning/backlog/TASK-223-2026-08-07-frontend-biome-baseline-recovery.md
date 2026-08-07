# TASK-223 — 🧹 Frontend Biome baseline recovery

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **На проверке** · Фаза: **128**

## [OUTCOME]

`npm run lint` возвращён в green отдельной контролируемой волной: 303 format
ошибки исправлены механически, а hook/a11y/import diagnostics разобраны вручную
без изменения поведения.

## [ROLE]

- Owner: `fe2` — formatting allowlist и ручные semantic fixes.
- Verifier / handoff: `fe-ver` — review behavior-sensitive diagnostics и полный
  frontend gate.

## [SCOPE AND CONSTRAINTS]

- Baseline: 303 format errors, 3 `useExhaustiveDependencies`, 2
  `useAriaPropsSupportedByRole`, 2 `organizeImports`, 4
  `useSemanticElements` warnings.
- Общий `--write` запрещён до allowlist/diff preview; dependency warnings нельзя
  автоподавлять или исправлять фиктивными deps.
- Depends on: TASK-222, чтобы массовое форматирование не конфликтовало с её
  четырёхфайловым allowlist.

## [ACCEPTANCE]

- `npm run lint`, `npx tsc --noEmit`, полный Vitest — green.
- Каждый non-format diagnostic имеет evidence-backed fix либо явный допустимый
  config decision; blanket ignore не добавлен.
- Formatting и semantic fixes разделены атомарными commits; reviewable diff и
  rollback сохранены.
