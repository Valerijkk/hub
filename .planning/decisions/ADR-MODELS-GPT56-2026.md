# ADR-MODELS-GPT56-2026: Переезд AI-команды на GPT-5.6 Sol / Sol Pro + Codex

- **Status:** accepted, amended 2026-07-26 (dynamic downstream routing)
- **Дата решения:** 2026-07-18
- **Runtime routing applied:** **2026-07-26**; model catalog/profile проверяется при запуске
- **Автор:** user + docs-sync · **Верифицируют:** team-lead, product-ver
- **Связанные:** `.cursor/playbooks/README.md`, `team-workflow.md`, `.cursor/REGULATION.md`, `AGENTS.md`, `.cursor/agents/*`

---

## Контекст

До июля 2026 AI-команда Hub работала на Claude:

| Роль | Модель (as-is) |
|------|----------------|
| `team-lead` | Fable 5 (thinking-high) |
| `team-lead-opus` | Opus 4.8 @ 1M (эскалация) |
| Все остальные (~28) | Sonnet 5 (thinking-high) |

Изначально пользователь планировал переезд в августе. Решением 2026-07-26
frontmatter/routing активированы раньше: Sol Pro orchestrator + dynamic downstream.

## Решение (целевой ростер моделей)

| Уровень | Модель | Кто | Когда |
|---------|--------|-----|-------|
| Оркестратор | **GPT-5.6 Sol Pro** | `team-lead` | applied 2026-07-26 |
| Исполнители + верификаторы | **Dynamic / inherit** | tech-lead, product*, ba*, sa*, arch*, designer*, be*, fe*, qa*, devops, security, debugger, code-reviewer, docs-writer | routed by orchestrator |
| Супер-мелкие задачи | **Lowest sufficient available model** | `codex-micro` slot | routed by orchestrator |

## Amendment 2026-07-26 — Sol Pro orchestrator + dynamic routing

Пользователь уточнил целевую политику:

1. **Оркестратор фиксирован:** GPT-5.6 Sol Pro.
2. **Исполнители не прибиты навсегда к одному семейству.** Оркестратор выбирает
   любую фактически доступную в Cursor модель по характеру задачи, её текущим
   сильным/слабым сторонам, риску, размеру контекста, стоимости и задержке.
3. **Portable agents используют `inherit`.** Конкретный model ID и параметры
   принадлежат launcher/model-router, а не ролевому prompt body.
4. **Каталог моделей проверяется на момент запуска.** Имена из ADR — policy
   defaults, не доказательство доступности конкретного slug.
5. **Выбор воспроизводим:** handoff фиксирует выбранную модель, причину,
   fallback и фактический статус запуска. Неуспешный запуск не считается run.
6. **Доступ ≠ разрешение на данные:** для non-public контекста обязательны
   approved provider/processor, region, retention, training-use policy, allowed
   data classification и unexpired human/policy approval. `UNKNOWN` блокирует
   передачу internal/confidential/restricted данных.

Рекомендуемая маршрутизация:

| Класс | Default | Альтернатива |
|-------|---------|-------------|
| Оркестрация, конфликт требований, финальная приёмка | GPT-5.6 Sol Pro | остановка/вопрос пользователю |
| Долгая coding-задача | актуальная coding-specialized модель Cursor | GPT-5.6 Sol |
| Анализ/архитектура/security | модель с сильным reasoning и достаточным контекстом | GPT-5.6 Sol Pro |
| Быстрый механический проход | доступная fast-модель | GPT-5.6 Sol |
| Независимая верификация | по возможности другая подходящая модель/агент | человек для high-impact |

Это amendment отменяет запрет на downstream-модели других провайдеров из
старого пункта «после cutover не выбирать Claude без override». Ограничение
остаётся для роли оркестратора: она закреплена за GPT-5.6 Sol Pro.

Техническая реализация хранится в
`pro-qa-ai/models/registry.template.yaml` и
`pro-qa-ai/workflows/playbooks/model-routing.md`. Frontmatter миграция выполнена
2026-07-26; explicit downstream override всё равно требует живого catalog check.

### Что меняется относительно as-is

1. **Уходит статическая эскалация Fable → Opus.** Оркестратор — Sol Pro;
   `team-lead-opus` остаётся deprecated redirect.
2. **Ролевые агенты используют `inherit`.** Orchestrator может выбрать любую
   реально доступную модель после catalog-backed оценки.
3. **`codex-micro` — имя micro-slot, не model pin.** Для него выбирается lowest
   sufficient available coding model.

### Что НЕ меняется

- Состав ролей (32 agent files), RACI, шаблон делегирования team-lead §5,
  цикличное качество (rule 56), docs-sync (55), TZ-инварианты.
- Playbooks и база знаний `docs/*` — модель-агностичны.

## Cursor frontmatter (applied 2026-07-26)

| Агент | `model:` | Runtime |
|-------|----------|---------|
| `team-lead` | `gpt-5.6-sol` | Pro mode разрешается текущим Cursor catalog/profile |
| 31 downstream agents | `inherit` | parent override только после routing + log |
| `team-lead-opus` | `inherit` | deprecated redirect |

## Когда что запускать

| Класс задачи | Агент / модель |
|--------------|----------------|
| Волна, декомпозиция, приёмка, конфликт приоритетов | `team-lead` → **Sol Pro** |
| Фича / спека / тесты / ревью / ADR | соответствующая роль + routed model |
| «Поправь опечатку», «добавь одну строку в .gitignore», «переименуй символ в одном файле» | `codex-micro` slot + lowest sufficient model |

Антипаттерн: гонять Sol Pro / полный team-lead на typo; выбирать модель по
бренду/памяти вместо task fit и текущего catalog.

## Applied checklist

1. `[done]` `team-lead` frontmatter → `gpt-5.6-sol`; Pro — runtime profile.
2. `[done]` 31 downstream frontmatter → `inherit`.
3. `[done]` `team-lead-opus.md` → deprecated redirect.
4. `[done]` AGENTS/REGULATION/playbooks synchronized.
5. `[pending]` Representative model-routing/eval pilot; не подменять structural smoke.
6. `[pending]` Session note/acceptance после полного validation и diff review.

## Последствия

- Стоимость/латентность: Sol Pro дороже Sol — только на оркестрации.
- A/B Claude (Fable/Opus) остаётся историческим обоснованием до-августовской
  политики. После amendment downstream-модель выбирает Sol Pro orchestrator по
  актуальному catalog; отдельный пользовательский override нужен только для
  смены самого orchestrator или high-impact решения.

---

*Accepted 2026-07-18 · amended 2026-07-26 · frontmatter cutover после catalog validation*
