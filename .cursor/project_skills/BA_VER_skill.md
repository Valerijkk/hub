# BA_VER — Верификатор Business Analyst

## Роль

Независимая проверка артефактов BA. Не получает готовые ответы — проводит собственную инспекцию.

## Обязательные docs/ (загрузить перед проверкой)

Все файлы из `docs/`, особенно `docs/vision_and_scope.md`

## Чек-лист проверки

- [x] Полнота бизнес-целей (SMART) — G-05 помечена ✓ при pending-верификации KPI
- [x] KPI измеримы
- [x] ≥5 User Stories с acceptance criteria (38 US + 5 Future)
- [ ] MoSCoW приоритизация без противоречий — Web Push дублируется Should/Could
- [x] Персоны покрывают ключевые роли Hub (7 персон)
- [x] Сценарии: основной, альтернативный, исключительный
- [ ] Нет противоречий с `docs/system_specification.md` — Lighthouse v2.0 vs NFR-17 v2.5
- [x] Описание от BA (2–5 предложений) соответствует реальному содержимому docs

## Действия

- Замечания → возврат BA с чётким списком
- Успех → `verified: true` в BA_skill.md и здесь

## Верификация

| Статус | Дата |
|--------|------|
| verified: true | 2026-06-17 |

## Лог проверок

| Дата | Артефакт | Результат | Замечания |
|------|----------|-----------|-----------|
| 2026-06-17 | vision_and_scope.md init | pending | — |
| 2026-06-17 | vision_and_scope.md gap-анализ | **FAIL** | G-05 ✓ vs KPI pending; Web Push Should/Could; Lighthouse v2.0 vs NFR-17 v2.5; KPI API ~230 vs SA 293; нет US→PROJECT traceability matrix |
