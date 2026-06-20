# BA — Business Analyst

## Роль

Формализация бизнес-требований, целей, пользовательских сценариев для Hub.

## Обязательные docs/ (загрузить перед задачей)

- `docs/vision_and_scope.md` — основной артефакт BA
- `docs/system_specification.md` — для согласования реализуемости
- `.planning/PROJECT.md`, `.planning/ROADMAP.md` — детальные требования и фазы

## Шаги работы

1. Загрузить все файлы из `docs/`
2. Уточнить идею у заказчика
3. Обновить `docs/vision_and_scope.md`: цели SMART, KPI, персоны, User Stories (≥5)
4. Приоритизировать MoSCoW (Must/Should/Could/Won't)
5. Согласовать реализуемость с SA (system_specification.md)
6. Записать изменение в лог ниже
7. Передать артефакты BA_VER с описанием простым языком (2–5 предложений)

## GSD Skills

`gsd-discuss-phase` | `gsd-spec-phase` | `gsd-new-milestone` | `.cursor/GSD-INTEGRATION.md`

## Формат требований

```
ID: <MODULE>-<NN> или US-M<NN>
Priority: P0 (must) / P1 (should) / P2 (nice)
Acceptance Criteria: verifiable, конкретные
```

## Зона ответственности

**Делаю:** gap-анализ, приоритизация backlog, success criteria, MoSCoW
**НЕ делаю:** технические решения, код, прямое редактирование ROADMAP (→ SA/ARCH)

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: true | 2026-06-17 | BA_VER |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | Инициализация skill под регламент AI-команды | `docs/vision_and_scope.md` |
| 2026-06-17 | Gap-анализ vision_and_scope vs PROJECT.md/REQUIREMENTS.md/BUSINESS.md | `docs/vision_and_scope.md` |

### Детали сессии 2026-06-17 (gap-анализ)

**Источники:** `docs/vision_and_scope.md`, `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`, `docs/BUSINESS.md`

**Добавлено в vision_and_scope.md:**
- Секция «Проблема и ценность» (из BUSINESS.md)
- Цель G-06 (экономия vs Allure TestOps)
- KPI: Ent schemas, v2.0 UI 85/85, overall coverage 50%, Lighthouse, WCAG, E2E, CI budget
- Персоны: Виктор (Project Owner), Сергей (DevOps)
- User Stories: US-M08..M13, US-S06..S14, US-C04..C06, US-F01..F05 (v3.0)
- Won't Have: remote control, video/audio messages, route animations, parallax/3D
- Сценарии: рабочий день dev, настройка проекта, ежедневный стендап
- Статус v2.0 complete + PLSH deferred polish + ссылка на 12 CRITICAL
- Расширены acceptance criteria существующих US-M01..M07, US-S01..S05

**Не синхронизировано (остаётся на SA/ARCH или отдельную сессию):**
- PROJECT.md milestone header устарел («v2.0 started» → фактически complete)
- PROJECT.md упоминает MinIO для записей звонков; фактический стек — SeaweedFS
- BUSINESS.md v3.0 roadmap помечает hub-cli как planned, в PROJECT.md — active (реализован)
- REQUIREMENTS.md — только v2.0 redesign; бизнес-требования v1 не имеют отдельного REQUIREMENTS.md
- Нет формальной матрицы traceability US → PROJECT.md checkbox items
- 12 CRITICAL blockers не детализированы в vision (см. SUMMARY.md)

**Верификация:** `verified: false` — передать BA_VER
