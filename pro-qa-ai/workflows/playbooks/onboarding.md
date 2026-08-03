# Playbook: онбординг тулкита в новый проект

> Цель: от копирования папки до рабочего конвейера агентов. Обычно 1 сессия.

## Шаги

```
1. КОПИРОВАНИЕ
   Собрать target через scripts/package_portable.py; generated profiles не копировать.
   Канонические заготовки берутся только из context/templates/.

2. BOOTSTRAP (главный шаг)
   Сначала scripts/context_inventory.py на явные project/test roots
   (output вне этих roots). Затем skills/bootstrap-test-context/SKILL.md
   с обязательными project_root, test_roots, inventory, profile.
   Не обнаруживать test roots самостоятельно.
   Выход: context/projects/<profile>/ (manifest + 8 Markdown artifacts) + report.
   Гейт: каждый факт имеет SOURCE-MAP evidence или literal UNKNOWN;
   найденные команды записаны verbatim со статусом not-run.

3. РАЗРЕШЕНИЕ UNKNOWN
   Каждый UNKNOWN — либо подтверждён источником при следующем refresh, либо
   осознанно оставлен; human notes живут только в BEGIN/END MANUAL blocks.
   Расхождения «факт кода ↔ style guide» — решает владелец, решение в канон.

4. ПОДКЛЮЧЕНИЕ АГЕНТОВ
   Не копировать agents/*.md как runtime config.
   Выбрать adapter из adapters/:
   - Cursor → adapters/cursor/ (render wrapper + frontmatter)
   - generic chat → adapters/generic-chat/
   - CI → adapters/ci/
   Проверить readonly/write permissions и approval boundary.

4a. MODEL REGISTRY
   Снять snapshot актуального runtime catalog в копию
   models/registry.template.yaml; настроить model-routing.md.
   Cursor orchestrator: resolve текущую запись GPT-5.6 Sol Pro;
   downstream wrappers по умолчанию model: inherit.

5. SMOKE-ТЕСТ КОНВЕЙЕРА
   a. test-code-reviewer + agents/templates/review-request.md — на diff с тестами.
   b. test-case-generator + agents/templates/test-case-generation.md — на одну фичу со спекой.
   c. test-verification — на выход шага b.
   d. anti-ai-check — на весь репозиторий (первый прогон даст исторический долг —
      зафиксировать как baseline, не чинить всё сразу).
   e. naming-check — на одну тестовую директорию.
   f. orchestrator → docs-only wave с routing/fallback log.
   Гейт: context загружен; форматы соблюдены; permissions и routing подтверждены.

6. ФИКСАЦИЯ БАЗОВОЙ ЛИНИИ
   Исторический долг из шагов 5d/5e — списком в трекер (не блокирует новую работу).
   Новые правила действуют с этого момента: новый код — по канону.

7. ИНТЕГРАЦИЯ С РЕГЛАМЕНТОМ ПРОЕКТА
   Ссылка на тулкит из правил проекта (Cursor rules / CLAUDE.md / AGENTS.md) +
   правило «тестовый код — через конвейер тулкита».
```

## Критерии готовности

- [ ] Profile report: manifest `complete|partial`, все 8 docs существуют
- [ ] Каждый populated fact имеет source ID; неизвестное = `UNKNOWN`
- [ ] Smoke: 6/6 сценариев отработали по форматам
- [ ] Adapter wrappers и readonly permissions проверены
- [ ] Catalog snapshot + orchestrator/downstream routing проверены
- [ ] Baseline долга зафиксирован
- [ ] Регламент проекта ссылается на тулкит

## Типовые грабли

| Симптом | Причина | Фикс |
|---------|---------|------|
| Агенты дают общие советы вместо проектных | active profile не загружен в сессию агента | Подключить `context/projects/<profile>/` |
| naming-check заваливает всё легаси | exceptions не зафиксированы | Обновить `PATTERN-INDEX.md` через отдельный curation gate |
| Команда из profile не работает | `COMMANDS.md` хранит discovery, а не доказательство запуска | Получить отдельное approval, прогнать и записать execution evidence |
| Верификация резает все TC как untraceable | Генерили без документации (из кода) | Найти/написать спеку или явный exploratory-режим |
