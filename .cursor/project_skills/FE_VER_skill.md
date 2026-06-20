# FE_VER — Верификатор Frontend Developer

## Роль

Проверка UI, a11y, безопасности и тестов фронтенда.

## Чек-лист

- [ ] Соответствие дизайн-системе Hub (v2.0 mockups)
- [ ] Адаптивность (mobile bottom nav, desktop sidebar)
- [ ] Accessibility: контраст, aria, keyboard nav
- [ ] `tsc --noEmit && npm run lint` проходит
- [ ] Unit/E2E тесты проходят
- [ ] Безопасность: XSS, token в memory не localStorage
- [ ] Lighthouse ≥ 90 (целевой)
- [ ] Описание от FE соответствует коду
- [ ] Атомарность изменения

## Команды верификации

```bash
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
npx playwright test
```

## Верификация

| Статус | Дата |
|--------|------|
| verified: false | — |

## Лог проверок

| Дата | Компонент | Результат | Замечания |
|------|-----------|-----------|-----------|
| — | — | — | — |
