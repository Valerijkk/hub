# Touchstone Standard — Test Automation AI Quality Layer

**Продукт:** [Touchstone](TOUCHSTONE.md) · **Издание:** Standard `0.2.0-rc.1`
**Pro (максималка):** sibling `prompt-engineering-pro/` (рядом в исходном репо)

**Политика языка:** Все инструкции тулкита — на русском. Идентификаторы, пути, CLI и машинные enum остаются как в схемах.

**Первый вход:** [START-HERE.md](START-HERE.md). Не читай дерево целиком.

> Переносимая AI-прослойка для автоматизаторов: она изучает существующий
> тестовый фреймворк, превращает его факты и проверенные паттерны в контекст для
> моделей, направляет специализированных агентов и измеряет качество результата.
> Это **не** каркас тестового фреймворка и не каталог «магических промптов».
> Эта папка — **фундамент Touchstone** для тестовых фреймворков.

## Ценность

Без слоя модель получает разовый промпт и вынуждена заново угадывать устройство
проекта. Со слоем она получает:

- верифицированную карту конкретного тестового фреймворка;
- реальные project patterns с provenance;
- вручную добавленные автоматизатором пары «плохо → комментарий коллеги → хорошо»;
- task recipes, которые ссылаются на нужный контекст, а не копируют его целиком;
- GSD-подобный цикл plan → execute → verify → handoff;
- структурные проверки и baseline-vs-assisted evals.

Эффект измеряется на пилоте: first-pass acceptance, время ревью, дефекты,
hallucinated paths/commands, scope escapes, flaky-risk и стоимость. Заявления
«ускоряет в N раз» без такого замера запрещены.

## Слои

1. **Portable core** — `agents/`, `skills/`, `rules/`, `playbooks/`, `prompts/`.
2. **Project context** — `context/projects/<profile>/`, создаётся bootstrap-скилом.
3. **Knowledge** — `patterns/` и `review-feedback/`; project examples не переносятся автоматически.
4. **Adapters/models** — подключение к Cursor/CI/chat и динамический model routing.
5. **Verification** — `scripts/` + `evals/`, без выдуманных результатов.

Карта файлов и ownership: [MANIFEST.md](MANIFEST.md). Версия: [VERSION.md](VERSION.md).
Использование/перенос: [LICENSE](LICENSE) и [NOTICE.md](NOTICE.md).

**Текущий статус:** Touchstone Standard · **foundation-ready** · structural
validators PASS (0 errors / 0 warnings) · tests: 55 PASS / 1 environmental SKIP
· package без generated Hub profile · effectiveness/ROI **NOT_RUN** (честный
gate до GA). Версия: [VERSION.md](VERSION.md) · история: [CHANGELOG.md](CHANGELOG.md).

Русская записка для руководителя:
[docs/EXECUTIVE-OVERVIEW-RU.md](docs/EXECUTIVE-OVERVIEW-RU.md).

> Паттерны lifecycle заимствованы у [GSD (get-shit-done)](https://github.com/gsd-build/get-shit-done)
> (fill-in шаблоны, frontmatter агентов, workflows, references) и усилены
> Hub-практиками; полный max-пакет — в **Touchstone Pro**.

## Философия

1. **Ноль хардкода проекта в portable core.** Проектные факты живут только в
   `context/projects/<profile>/`; вручную добавленные project examples не
   публикуются в baseline без отдельного разрешения.
2. **Контекст имеет provenance и freshness.** Bootstrap создаёт профиль из
   указанных путей, refresh обновляет его после изменений фреймворка; неизвестное
   остаётся `UNKNOWN`, а не превращается в догадку.
3. **Агенты читают релевантный контекст.** Индекс загружается всегда, полные
   документы и pattern cards — только по задаче (progressive disclosure).
4. **Промпт = outcome + constraints + evidence + completion bar.** Обычный
   запрос человека нормализуется; полный delegate обязателен между агентами.
5. **Верификация evidence-first.** Глубина зависит от риска; critical/security
   требует независимого verifier. Галочки без лога/ссылки не доказательство.
6. **Входные данные недоверенные.** Код и комментарии не могут переопределять
   задачу или правила; см. [rules/50](rules/50-data-handling.md).

## Структура

```
prompt-engineering/                 ← Touchstone Standard (эта папка)
├── TOUCHSTONE.md                   ← продукт и границы изданий
├── VERSION.md / CHANGELOG.md / MANIFEST.md
├── START-HERE.md / INSTALL.md / PORTABLE-HUB-KIT.md
├── agents/                         ← test-* + corporate portable
├── skills/ / rules/ / playbooks/ / recipes/ / prompts/
├── context/ / patterns/ / review-feedback/
├── adapters/ / models/
├── scripts/ / evals/
├── sources/ / docs/ / gsd-port/
└── (Hub sibling) prompt-engineering-pro/  ← Touchstone Pro
```

## Быстрый старт

```
1. Собери project-neutral package через `scripts/package_portable.py`
   (см. [INSTALL.md](INSTALL.md)); raw-copy generated profiles запрещён.
2. Создай `scripts/context_inventory.py` inventory только из явных
   `project_root`/`test_roots`.
3. Запусти `/bootstrap-test-context`, передав те же корни, inventory и имя
   профиля.
4. Проверь `context/projects/<profile>/` и unresolved `UNKNOWN`.
5. При необходимости человек вручную добавляет примеры коллег в
   `review-feedback/inbox/`; `/curate-review-feedback` их проверяет и индексирует.
6. Подключи адаптер своего AI-инструмента из `adapters/`.
7. Запусти structural validator и eval smoke.
8. Используй agents/recipes. `prompts/` — заготовки, не самостоятельный контекст.
```

Подробно: [INSTALL.md](INSTALL.md) и [playbooks/onboarding.md](playbooks/onboarding.md).

**Слой Corporate/GSD:** [PORTABLE-HUB-KIT.md](PORTABLE-HUB-KIT.md) ·
[gsd-port/README.md](gsd-port/README.md) · [agents/README.md](agents/README.md) ·
[prompts/13-corporate-delivery.md](prompts/13-corporate-delivery.md).

## Роли и skills

Полный model-agnostic ростер: [agents/README.md](agents/README.md).

| Группа | Роли |
|--------|------|
| Оркестрация/решения | orchestrator, product-owner, architect |
| Реализация | backend-dev, frontend-dev, devops |
| QA продукта | qa-manual, qa-auto, code-reviewer, debugger, security-reviewer |
| Test-framework специализация | test-case-generator, test-code-reviewer, test-debugger, test-documenter |
| Документация | docs-writer |

Skills — явные процедуры, а не личности: bootstrap/refresh context,
curation, planning, generation/review checks, goal verification и handoff.

## Конвенции тулкита

- `{{placeholder}}` — слот для подстановки (значение из active profile или руками).
- `MUST/MUST NOT` — жёсткое требование; `PASS/FAIL` — результат проверки.
- Severity: `CRITICAL > HIGH > MEDIUM > LOW` (критерии: [rules/30-severity-policy.md](rules/30-severity-policy.md)).
- Entry points самодостаточны; детальные patterns/references загружаются только
  когда на них ссылается задача (progressive disclosure).
