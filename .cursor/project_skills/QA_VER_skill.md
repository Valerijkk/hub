# QA_VER — Верификатор QA

## Роль

Проверка полноты test plan и связи с acceptance criteria.

## Чек-лист

- [ ] test_plan.md полон: пирамида, критерии, регрессия, инструменты
- [ ] Acceptance Criteria из vision_and_scope → тест-кейсы
- [ ] Матрица покрытия требований актуальна
- [ ] Smoke-тесты на билде проходят
- [ ] Баг-репорты корректны и актуальны
- [ ] Регрессионный набор запускается при каждом изменении
- [ ] При FAIL: triage выполнен (ROOT=code|test) — см. `.cursor/skills/test-triage/SKILL.md`

## GSD Skills

`gsd-verify-work`, `gsd-validate-phase` | Матрица: `.cursor/GSD-INTEGRATION.md`

## Команды

```bash
/qa-gate   # полный чек
/coverage-audit  # матрица требований × код × тесты
```

## Верификация

| Статус | Дата |
|--------|------|
| verified: false | — |

## Лог проверок

| Дата | Артефакт | Результат | Замечания |
|------|----------|-----------|-----------|
| 2026-06-17 | test_plan.md init | pending | — |
