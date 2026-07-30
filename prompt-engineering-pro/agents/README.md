# Роли разработки — Touchstone Pro

Двенадцать model-agnostic role contracts для **разработки продукта**.
Это не тестовые роли: агенты тестового фреймворка (`test-*`) живут в
Standard и сюда не дублируются.

Файлы не назначают провайдера, model ID и sampling-параметры — выбор
модели идёт через [`../../prompt-engineering/playbooks/model-routing.md`](../../prompt-engineering/playbooks/model-routing.md).
Подключение к Cursor — через [`../adapters/cursor/`](../adapters/cursor/):
raw markdown в `.cursor/agents/` копировать нельзя, нужен frontmatter и wrapper.

## Ростер

| Агент | Назначение | Режим |
|-------|------------|-------|
| [orchestrator](orchestrator.md) | Нормализация цели, волны, routing, goal-backward приёмка | координация |
| [product-owner](product-owner.md) | Scope, SMART, MoSCoW, KPI, явные Won't | docs/decision |
| [architect](architect.md) | C4, ADR, выбор фреймворка, границы модулей | docs/decision |
| [backend-dev](backend-dev.md) | Серверная реализация, unit/integration regression | write |
| [frontend-dev](frontend-dev.md) | UI/UX, a11y, frontend regression | write |
| [qa-manual](qa-manual.md) | Тест-дизайн продукта, TC, UAT, ручные чек-листы | test docs |
| [qa-auto](qa-auto.md) | Автотесты продукта по принятым TC, трассируемость | write |
| [code-reviewer](code-reviewer.md) | Defect-first pre-merge review продуктового diff | readonly |
| [debugger](debugger.md) | Evidence-based root cause и triage | diagnose; fix по разрешению |
| [security-reviewer](security-reviewer.md) | Независимый security review | readonly |
| [devops](devops.md) | CI/CD, runtime, smoke, rollback | write |
| [docs-writer](docs-writer.md) | Product/ops документация, сверка с фактами | docs |

## Граница с Standard

| Объект работы | Роль | Издание |
|---------------|------|---------|
| Продукт, который тестируют | `qa-manual`, `qa-auto` | Pro |
| Сам тестовый фреймворк | `test-case-generator`, `test-debugger` | Standard |
| Продуктовые и ops-документы | `docs-writer` | Pro |
| Документация тестового фреймворка | `test-documenter` | Standard |
| Продуктовый diff | `code-reviewer` | Pro |
| Код тестового фреймворка | `test-code-reviewer` | Standard |

Пары не взаимозаменяемы. Если `code-reviewer` ревьюит код фреймворка —
это ошибка маршрутизации, а не «тоже сойдёт».

## Шаблоны запуска

| Шаблон | Когда |
|--------|-------|
| [`templates/phase-kickoff.md`](templates/phase-kickoff.md) | Старт фазы |
| [`templates/wave-delegate.md`](templates/wave-delegate.md) | Делегирование в волне |
| [`templates/corporate-wave.md`](templates/corporate-wave.md) | Корпоративная волна с несколькими ролями |

Базовый мастер-контракт — в Standard:
[`_MASTER-PROMPT-TEMPLATE.md`](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Переносимое ядро того же контракта — [`../extracts/master-delegate.md`](../extracts/master-delegate.md).

## Общий контракт

Действует контракт Standard (`agents/README.md` §Общий контракт) плюс два
усиления Pro:

1. **Evidence по пирамиде L0–L4** — [`../layers/05-verification-mutation.md`](../layers/05-verification-mutation.md).
   Галочка в отчёте (L0) доказательством не считается.
2. **Фикс с тестом требует мутации** — [`../extracts/mutation-evidence.md`](../extracts/mutation-evidence.md).
   Вердикты: `PROVEN`, `mutation_failed_to_build`, `layer_mismatch`, `NOT_PROVEN`.
