# ADR-MODELS-GPT56-2026: Переезд AI-команды на GPT-5.6 Sol / Terra / Luna + Codex

- **Status:** accepted, amended 2026-08-07 (cost-aware role defaults + protected Opus zone)
- **Дата решения:** 2026-07-18
- **Runtime routing applied:** **2026-07-26**; role defaults уточнены **2026-08-07**;
  model catalog/profile и pricing проверяются при запуске
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
frontmatter/routing активированы раньше: Sol orchestrator + dynamic downstream.

## Решение (целевой ростер моделей)

| Уровень | Модель | Кто | Когда |
|---------|--------|-----|-------|
| Оркестратор | **GPT-5.6 Sol / Extra High** | `team-lead` | applied 2026-07-26, effort clarified 2026-08-07 |
| Исполнители + верификаторы | **Dynamic / inherit** | tech-lead, product*, ba*, sa*, arch*, designer*, be*, fe*, qa*, devops, security, debugger, code-reviewer, docs-writer | role default из `model-routing.md`, override по риску |
| Супер-мелкие задачи | **GPT-5.6 Luna / Light** | `codex-micro` slot | Spark только как доступный interactive preview |

## Amendment 2026-07-26 / 2026-08-07 — Sol orchestrator + cost-aware routing

Пользователь уточнил целевую политику:

1. **Оркестратор фиксирован:** GPT-5.6 Sol / Extra High; Ultra — только для большой делимой задачи.
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

Рекомендуемая маршрутизация и точная матрица 32 ролей вынесены в
[`.cursor/playbooks/model-routing.md`](../../.cursor/playbooks/model-routing.md).
Краткий default:

| Класс | Default | Альтернатива |
|-------|---------|-------------|
| Оркестрация, конфликт требований, финальная приёмка | GPT-5.6 Sol / Extra High | Sol / Ultra только для делимой большой волны |
| Обычная разработка, анализ, QA и review | GPT-5.6 Terra / Medium–High | Sol при высокой цене ошибки |
| Архитектура и security | GPT-5.6 Sol / Extra High | Terra / Extra High только для независимого challenge |
| Быстрый механический проход | GPT-5.6 Luna / Light–Medium | Terra при любой неоднозначности |
| Интерактивная micro-итерация | Codex Spark / Light, если доступен | Luna / Light |
| Независимая верификация | Terra / High в отдельном контексте | Sol для critical/security |

Это amendment отменяет запрет на downstream-модели других провайдеров из
старого пункта «после cutover не выбирать Claude без override». Ограничение
остаётся для роли оркестратора: она закреплена за GPT-5.6 Sol.

Операционный канон Hub хранится в
`.cursor/playbooks/model-routing.md`. Frontmatter миграция выполнена
2026-07-26; explicit downstream override всё равно требует живого catalog check.
`pro-qa-ai/**` не является реализацией роутера Hub: это защищённая внешняя зона
Claude Code Opus 5, которой управляет founder; агенты Hub её не меняют.

### Что меняется относительно as-is

1. **Уходит статическая эскалация Fable → Opus.** Оркестратор — Sol / Extra High;
   `team-lead-opus` остаётся deprecated redirect.
2. **Ролевые агенты используют `inherit`.** Orchestrator применяет role default
   из `model-routing.md` и может повысить/понизить runtime после catalog-backed оценки.
3. **`codex-micro` — имя micro-slot, не model pin.** Default — Luna / Light;
   Spark допустим только как доступный interactive preview.

### Что НЕ меняется

- Состав ролей (32 agent files), RACI, шаблон делегирования team-lead §5,
  цикличное качество (rule 56), docs-sync (55), TZ-инварианты.
- Playbooks и база знаний `docs/*` — модель-агностичны.

## Cursor frontmatter (applied 2026-07-26)

| Агент | `model:` | Runtime |
|-------|----------|---------|
| `team-lead` | `gpt-5.6-sol` | Extra High default; Ultra только для большой делимой волны |
| 31 downstream agents | `inherit` | parent override только после routing + log |
| `team-lead-opus` | `inherit` | deprecated redirect |

## Когда что запускать

| Класс задачи | Агент / модель |
|--------------|----------------|
| Волна, декомпозиция, приёмка, конфликт приоритетов | `team-lead` → **Sol / Extra High** |
| Фича / спека / тесты / ревью / ADR | соответствующая роль + routed model |
| «Поправь опечатку», «добавь одну строку в .gitignore», «переименуй символ в одном файле» | `codex-micro` + Luna / Light (Spark при доступности) |

Антипаттерн: гонять Sol / полный team-lead на typo; выбирать модель по
бренду/памяти вместо task fit и текущего catalog; назначать 5.5 по цене Sol
или новые 5.4/5.4 Mini pin перед их Codex retirement 2026-08-31.

## Applied checklist

1. `[done]` `team-lead` frontmatter → `gpt-5.6-sol`; Extra High — default effort.
2. `[done]` 31 downstream frontmatter → `inherit`.
3. `[done]` `team-lead-opus.md` → deprecated redirect.
4. `[done]` AGENTS/REGULATION/playbooks synchronized.
5. `[done]` Cost-aware role matrix + effort thresholds → `model-routing.md` (2026-08-07).
6. `[done]` `pro-qa-ai/**` закреплён как external Opus-owned protected zone.
7. `[pending]` Representative model-routing/eval pilot; не подменять docs review.

## Последствия

- Стоимость/латентность: Sol используется для самых дорогих по цене ошибки
  решений; Terra — рабочий default, Luna — механический high-volume слой.
- A/B Claude (Fable/Opus) остаётся историческим обоснованием до-августовской
  политики. После amendment downstream-модель выбирает Sol orchestrator по
  актуальному catalog; отдельный пользовательский override нужен только для
  смены самого orchestrator или high-impact решения.
- Claude Code Opus 5 работает только внутри `pro-qa-ai/**` по управлению founder.
  Результат из этой зоны попадает в Hub только через явный handoff и повторную
  проверку как недоверенный внешний артефакт.

---

*Accepted 2026-07-18 · amended 2026-07-26 / 2026-08-07 · frontmatter остаётся portable через `inherit`*
