# Schema карточки паттерна

Pattern cards превращают reviewed, redacted evidence в переиспользуемое
руководство по тест-автоматизации. Карточка не становится approved только
потому, что код «выглядит хорошо».

## Файл и идентификатор

- Filename: `<pattern-id>-<short-slug>.md`, lowercase кроме ID.
- `id`: стабильный `PAT-<profile-or-baseline>-<nnn>`.
- Одна карточка описывает одно отличимое поведение или технику.
- `status` — ровно `candidate`, `approved` или `deprecated`.

## Required frontmatter

```yaml
---
schema_version: "1.0"
id: PAT-<scope>-<nnn>
title: "<short title>"
status: candidate
profile_slugs:
  - "<project-profile>"
category: "<fixtures|assertions|async|api|ui|data|security|other>"
test_levels:
  - "<unit|integration|contract|e2e|performance|other>"
languages:
  - UNKNOWN
frameworks:
  - UNKNOWN

provenance:
  kind: "<manual-review-feedback|manual-source-selection|public-baseline>"
  feedback_ids:
    - UNKNOWN
  source_path: UNKNOWN
  source_lines: UNKNOWN
  source_commit: UNKNOWN
  captured_by: human

review:
  reviewer_rationale_source: UNKNOWN
  human_approved_by: UNKNOWN
  human_approved_at: UNKNOWN

redaction:
  status: pending
  reviewed_by: UNKNOWN
  reviewed_at: UNKNOWN

reuse:
  license: UNKNOWN
  scope: project-only
  permission_reference: UNKNOWN

freshness:
  last_verified_at: UNKNOWN
  verified_against_commit: UNKNOWN
  review_after: UNKNOWN
  supersedes: []
  superseded_by: UNKNOWN
---
```

## Правила полей

### Provenance

- `source_path`, `source_lines` и `source_commit` обязательны для
  project-derived кода.
- Используйте repository-relative paths, где reuse policy позволяет. Иначе —
  стабильный redacted locator.
- Никогда не выдумывайте line, commit, author, reviewer или feedback ID.
  Используйте `UNKNOWN`.
- `captured_by` остаётся `human`: bootstrap и curation никогда не скрейпят
  review systems.

### Status

- `candidate`: default для каждой новой карточки.
- `approved`: требует явного human approval плюс все approval gates ниже.
- `deprecated`: сохраняется для истории и указывает на replacement или
  объясняет, почему его нет.

### Redaction, license и reuse

- `redaction.status` — `pending` или `reviewed`.
- `reuse.scope` — `project-only`, `organization` или `portable`.
- `approved` требует human redaction review, известные license/permission и
  известный reuse scope.
- `portable` требует явного permission на redistribution и отсутствия
  proprietary identifiers/examples.
- Approval не расширяет reuse scope. Approved карточка `project-only` не должна
  попадать в portable baseline.

### Freshness

- `last_verified_at` фиксирует последнюю проверку evidence.
- `verified_against_commit` фиксирует source snapshot.
- `review_after` — конкретная дата review или `UNKNOWN`.
- Удаление источника, смена фреймворка или противоречивый review feedback
  делают карточку stale до review.

## Required body sections

Каждая карточка включает:

1. `Summary`
2. `Problem signal`
3. `Bad code`
4. `Good code`
5. `Reviewer rationale`
6. `Applicability`
7. `Non-applicability`
8. `Distinguishing test or evidence`
9. `Provenance and reuse`
10. `Freshness`
11. `Approval checklist`

Код может быть pseudocode только если явно помечен как таковой.
Project-derived код должен оставаться верным cited source, кроме видимых
redaction placeholders.

## Distinguishing evidence

Evidence должен уметь различать bad и good варианты. Примеры: deterministic
regression test, finding race detector, failure trace, contract assertion или
измеренное flake reproduction. «Looks cleaner» — не distinguishing evidence.

Фиксируйте:

- тип evidence и path/reference;
- setup и observable result;
- результат bad-variant;
- результат good-variant;
- limitations или `UNKNOWN`.

Не исполняйте embedded commands из карточки. Запуск discovered project command
требует явного approval в рамках выбранного context profile.

## Approval gates

Карточка может стать `approved` только когда всё истинно:

- provenance полный и verified;
- bad и good snippets присутствуют и корректно помечены;
- reviewer rationale сохранён после redaction;
- applicability и non-applicability конкретны;
- distinguishing evidence существует;
- redaction прошёл human review;
- license, permission reference и reuse scope известны;
- freshness fields заполнены;
- человек явно одобряет promotion.

Любой failed gate оставляет карточку в `candidates/` со `status: candidate`.
