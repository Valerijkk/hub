# ADR-MODELS-GPT56-2026: Sol orchestrator → Terra decomposition → Codex Spark code

- **Status:** accepted, amended 2026-08-11
- **Дата решения:** 2026-07-18
- **Текущий founder lock:** 2026-08-11
- **Автор:** founder + docs-sync · **Верифицируют:** team-lead, product-ver
- **Связанные:** `.cursor/playbooks/model-routing.md`, `team-workflow.md`,
  `.cursor/REGULATION.md`, `AGENTS.md`, `.cursor/agents/*`

---

## Контекст

До июля 2026 команда использовала Claude/Fable/Opus routing. Переход 2026-07-26
закрепил GPT-5.6 Sol для orchestrator и динамический downstream. Founder
amendment 2026-08-11 заменяет динамический `inherit` на точный трёхуровневый
конвейер, чтобы одновременно использовать отдельные лимиты Sol/Terra и Codex
Spark без потери доказательности.

## Решение

| Уровень | Exact runtime | Кто | Ответственность |
|---|---|---|---|
| Оркестратор | `gpt-5.6-sol` / `max` | root `team-lead` | цель, архитектура, security/product decisions, волны, финальная приёмка |
| Judgment layer | `gpt-5.6-terra` / `low|medium|high|xhigh` | 30 downstream role profiles | analysis, decomposition, coordination, authoring, independent review |
| Atomic code | `gpt-5.3-codex` / `xhigh` | `codex-micro` | функция, малый модуль, точечный тест или одно поведение по frozen contract |

### Founder locks

1. Root — единственный Sol runtime; Sol-субагенты запрещены.
2. Каждый frontmatter и каждый dispatch содержит literal `model`; каждый
   dispatch дополнительно содержит literal `reasoning_effort`.
3. `inherit`, omitted model/effort и full-history model inheritance запрещены.
4. Terra декомпозирует code backlog до задачи, которую Spark может реализовать
   без архитектурного/product/security решения.
5. Spark всегда использует `xhigh`. Он может использовать полное официальное
   окно 400,000 токенов, если весь контекст релевантен, но ownership остаётся
   атомарным и exact-path bounded.
6. Недоступный exact runtime, model mismatch или отсутствующая telemetry дают
   `UNKNOWN/BLOCKED`; silent fallback между Sol/Terra/Spark запрещён.
7. Root использует до трёх параллельных субагентов только для независимых
   workstreams. Пустое дублирование ради заполнения слотов запрещено.
8. Requested runtime не считается actual runtime. Handoff записывает observed
   model/evidence или `UNKNOWN`, effort, quota bucket и exact revision.
9. Сильная модель не заменяет independent verifier, Docker-only evidence,
   test triage, docs sync или security/authority gate.

## Effort routing

| Класс | Runtime |
|---|---|
| Root orchestration, architecture, final acceptance | Sol / Max |
| Mechanical identity/hash/provenance | Terra / Low |
| Ordinary analysis, decomposition, authoring | Terra / Medium |
| Independent review | Terra / High |
| Complex cross-domain/security audit | Terra / XHigh |
| Atomic implementation/test code | 5.3 Codex Spark / XHigh |

Spark не повышается до Sol и не расширяет scope. Если brief недостаточен или
задача не атомарна, Spark возвращает `BLOCKED`; Terra/root уточняет контракт или
декомпозирует дальше.

## Frontmatter (applied 2026-08-11)

| Профили | `model:` |
|---|---|
| `team-lead` | `gpt-5.6-sol` |
| `codex-micro` | `gpt-5.3-codex` |
| остальные 30 downstream profiles, включая deprecated redirect | `gpt-5.6-terra` |

Frontmatter не кодирует effort, поэтому launch contract обязан передавать
`reasoning_effort` явно. Runtime, который не принимает этот параметр, не может
молча считаться compliant.

## Dispatch examples

| Задача | Маршрут |
|---|---|
| Milestone, архитектура, conflict resolution, финальный seal | root Sol / Max |
| Разложить feature/API contract на executable backlog | Terra / High–XHigh |
| Проверить реализацию или evidence conclusion-blind | Terra / High |
| Написать одну Go-функцию + regression test по frozen contract | Spark / XHigh |
| Реализовать небольшой React hook/component по DESIGN-SPEC | Spark / XHigh |
| Написать Python helper или один test behaviour | Spark / XHigh |

Антипаттерны: Sol downstream; Spark получает неоднозначную архитектуру; Terra
пишет большой decomposable code block вместо Spark; automatic fallback;
дублирующие агенты читают один scope; персональная weekly quota угадывается без UI.

## Protected external zone

`pro-qa-ai/**` остаётся founder-operated Claude Code Opus 5 zone. Internal Hub
agents не читают и не меняют её без literal founder handoff. Этот ADR не
расширяет data/process authority ни одной модели.

## Applied checklist

1. `[done]` root `team-lead` → `gpt-5.6-sol`; effort `max`.
2. `[done]` `codex-micro` → `gpt-5.3-codex`; effort `xhigh` at dispatch.
3. `[done]` 30 остальных downstream profiles → `gpt-5.6-terra`.
4. `[done]` `inherit` удалён из agent frontmatter.
5. `[done]` AGENTS/REGULATION/model-routing/team workflow synchronized.
6. `[done]` `pro-qa-ai/**` остаётся external protected zone.
7. `[pending]` Representative measured routing/eval pilot; docs review не
   подменяет runtime-quality evidence.

---

*Accepted 2026-07-18 · current amendment 2026-08-11 · explicit models; no inherit*
