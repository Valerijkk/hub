# ADR-MODELS-GPT56-2026: Переезд AI-команды на GPT-5.6 Sol / Sol Pro + Codex

- **Status:** accepted (planned cutover) — 2026-07-18, решение пользователя
- **Дата решения:** 2026-07-18
- **Cutover:** **август 2026** (точная дата — kickoff после July Rebuild / возврата)
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

Пользователь зафиксировал **переезд в августе** на семейство ChatGPT / Codex в Cursor.

## Решение (целевой ростер моделей)

| Уровень | Модель | Кто | Когда |
|---------|--------|-----|-------|
| Оркестратор | **GPT-5.6 Sol Pro** | `team-lead` | с cutover |
| Исполнители + верификаторы | **GPT-5.6 Sol** | tech-lead, product*, ba*, sa*, arch*, designer*, be*, fe*, qa*, devops, security, debugger, code-reviewer, docs-writer | с cutover |
| Супер-мелкие задачи | **Codex (latest)** | агент `codex-micro` / inline `gsd-fast`-класс | с cutover |

### Что меняется относительно as-is

1. **Уходит двухуровневая эскалация Fable → Opus.** Оркестратор сразу на **Sol Pro**; отдельный `team-lead-opus` после cutover **deprecated** (файл остаётся как redirect → `team-lead` на Sol Pro).
2. **Все ролевые агенты** (включая `*-ver`) — единый **Sol**, без отдельного «тяжёлого» Claude-тира.
3. **Codex (последний доступный в Cursor)** — только для **супер-мелких** правок: typo, однострочный fix, rename, мелкий docs one-liner, где не нужна оркестрация волн.

### Что НЕ меняется

- Состав ролей (30 агентов), RACI, шаблон делегирования team-lead §5, цикличное качество (rule 56), docs-sync (55), TZ-инварианты.
- Playbooks и база знаний `docs/*` — модель-агностичны.

## Cursor frontmatter (целевые slug)

> Точные slug в Cursor могут уточняться на день cutover (picker names ≠ API ids). Канон ниже — **целевой**; до августа frontmatter остаётся Claude.

| Агент | `model:` (после cutover) | Примечание |
|-------|--------------------------|------------|
| `team-lead` | `gpt-5.6-sol` + **Pro** (в UI: GPT-5.6 Sol Pro) | Pro = reasoning mode / picker Pro, не отдельный API-slug «sol-pro» |
| все ролевые + `*-ver` | `gpt-5.6-sol` | UI: GPT-5.6 Sol |
| `codex-micro` | latest **Codex** в Cursor picker | только micro-задачи |
| `team-lead-opus` | — | deprecated → не запускать |

На cutover: массово обновить `model:` / description во всех `.cursor/agents/*.md` + таблицы в README / REGULATION / AGENTS / team-workflow.

## Когда что запускать

| Класс задачи | Агент / модель |
|--------------|----------------|
| Волна, декомпозиция, приёмка, конфликт приоритетов | `team-lead` → **Sol Pro** |
| Фича / спека / тесты / ревью / ADR | соответствующий ролевой агент → **Sol** |
| «Поправь опечатку», «добавь одну строку в .gitignore», «переименуй символ в одном файле» | **Codex latest** (`codex-micro` или Composer/Codex inline) — **не** team-lead |

Антипаттерн: гонять Sol Pro / полный team-lead на typo; гонять Codex на волну с 4 модулями.

## План cutover (август)

1. Подтвердить slug в актуальном Cursor model picker.
2. Обновить frontmatter всех агентов + индексные таблицы.
3. Пометить `team-lead-opus.md` deprecated banner.
4. Smoke: одна волна (ba→sa→be) на Sol + одна micro-задача на Codex.
5. Session note в `STATE.md` + строка в `AGENTS.md`.

## Последствия

- Стоимость/латентность: Sol Pro дороже Sol — только на оркестрации.
- A/B Claude (Fable/Opus) остаётся историческим обоснованием до-августовской политики; после cutover не выбирать Claude для Hub-команды без явного override пользователя.

---

*Accepted 2026-07-18 · cutover August 2026*
