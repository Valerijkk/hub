# Горячие клавиши

> **Обновлено:** 2026-06-21 · Реализация: TanStack Router + command palette

---

## Глобальные

| Клавиша | Действие | Где |
|---------|----------|-----|
| `Ctrl+K` / `Cmd+K` | Глобальный поиск (Command Palette) | Везде в authenticated layout |
| `Ctrl+N` / `Cmd+N` | Создать новую задачу | Проект / доска |
| `?` | Справка по shortcuts (если включено) | `shortcuts-dialog.tsx` |

### Command Palette (`Ctrl+K`)

- **Файл:** `frontend/src/components/search/command-palette.tsx`
- **API:** `GET /api/search?q=...&type=...&limit=...`
- **Debounce:** 300 ms
- **Группы:** задачи, тест-кейсы, пользователи, проекты, каналы, файлы
- **Открытие:** клавиша или кнопка 🔍 в topbar
- **E2E:** `testframework/stages/e2e/domains/navigation/command-palette.spec.ts`

---

## Навигация

| Клавиша | Действие |
|---------|----------|
| `Esc` | Закрыть модал / palette / dropdown |
| `Tab` / `Shift+Tab` | Переход между focusable элементами (WCAG) |

---

## Мессенджер

| Клавиша | Действие |
|---------|----------|
| `Enter` | Отправить сообщение |
| `Shift+Enter` | Новая строка в composer |
| `Ctrl+V` | Вставить файл из буфера (если поддерживается браузером) |
| `@` | Упоминание пользователя (в composer) |

Ветка `calls-feature-codex`: управление подпиской в заголовке треда использует
обычную фокусируемую кнопку, отдельный глобальный shortcut не добавлен.
Unit-проверка подтверждает `type=button`, блокировку повторного нажатия и
сохранение кнопки при обновлении; реальный браузерный keyboard/focus-прогон
ещё `NOT_RUN`. [Текущая проверка](../superpowers/plans/2026-09-05-thread-follow-control.md).

Отметка прочтения ответов отдельного shortcut не имеет: она использует видимый
конец ответа в активном окне. При ошибке обычной кнопкой можно «Повторить
отметку прочитанного»; требуется новое наблюдение видимого ответа. Unit-проверка
кнопки и lifecycle пройдена; реальная browser/a11y-приёмка ещё NOT_RUN.
[Контракт](../superpowers/plans/2026-09-05-thread-read-receipts.md).

---

## Редактор задач и комментариев (Plate)

| Клавиша | Действие |
|---------|----------|
| `Ctrl+B` | Жирный текст |
| `Ctrl+I` | Курсив |
| `Ctrl+E` | Inline-код |
| `` ``` `` | Блок кода |
| `@` | Упоминание пользователя |

---

## Kanban / трекер

| Действие | Ввод |
|----------|------|
| Перемещение карточки | Drag & drop (pragmatic-drag-and-drop) |
| Быстрое создание | `Ctrl+N` на странице доски |

---

## Звонки (LiveKit)

| Элемент | Управление |
|---------|------------|
| Mute/unmute | UI-кнопка (клавиатура — через focus) |
| Screen share | UI-кнопка |
| Raise hand | UI-кнопка → `POST /api/calls/:id/raise-hand` |

---

## Accessibility

- Все интерактивные элементы доступны с клавиатуры (Tab)
- Focus ring видимый (Tailwind `focus-visible`)
- E2E a11y: `testframework/stages/e2e/domains/platform/a11y.spec.ts` (axe-core, WCAG 2.1 AA target)

---

## Добавление нового shortcut

1. Реализовать в компоненте (`useEffect` + `keydown` или Radix shortcut)
2. Добавить строку в этот файл
3. При user-visible flow — обновить `docs/vision_and_scope.md` если US
4. E2E при критическом пути (как `Ctrl+K`)

---
*2026-06-21 — Ctrl+K verified in CI e2e-full*
