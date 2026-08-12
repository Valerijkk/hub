# AI Quality Layer Rework — test automation

> Дата решения: 2026-07-26
> Scope: `.cursor/`, `pro-qa-ai/`, связанная `.planning/`
> Product code: не менять
> Commit: только по отдельному запросу пользователя

## Goal

`pro-qa-ai/` становится переносимой AI-прослойкой для автоматизаторов,
которая:

1. изучает существующий тестовый фреймворк по заданным путям;
2. создаёт проверяемый project context с provenance/freshness;
3. применяет approved patterns и вручную добавленную обратную связь коллег;
4. направляет специализированных агентов по GSD-подобному lifecycle;
5. измеряет качество и стоимость результата относительно bare-prompt baseline.

Это не scaffold тестового фреймворка и не обещание «лучшего промпта» без eval.

## Решения пользователя

- Orchestrator: GPT-5.6 Sol Pro.
- Downstream model выбирает orchestrator из фактически доступных Cursor models
  по задаче, риску, контексту, latency/cost и подтверждённым возможностям.
- Закрытые MR не сканируются и не читаются автоматически.
- После bootstrap человек сам добавляет в подготовленные файлы:
  bad code, комментарий коллеги, good code и доступную provenance.
- Conservative migration: улучшать in-place; удалять/перемещать только
  доказанные дубли после link/reference проверки.

## Target layers

| Layer | Paths | Назначение |
|-------|-------|------------|
| Core | `agents/`, `skills/`, `rules/`, `playbooks/`, `recipes/` | Роли и процессы |
| Project profile | `context/projects/<profile>/` | Факты тестового фреймворка |
| Knowledge | `patterns/`, `review-feedback/` | Candidate/approved patterns |
| Runtime adapters | `adapters/`, `models/` | Cursor/CI/chat + routing |
| Verification | `scripts/`, `evals/` | Structural lint и eval regression |
| Adoption | `docs/`, `sources/`, `VERSION.md`, `MANIFEST.md` | Руководитель, governance, источники |

## Bootstrap contract

Минимальный ввод:

```text
project_root=<repo-root>
test_roots=<one-or-more-roots>
inventory=<approved-local-evidence>/context-inventory.json
profile=<project-slug>
```

Сначала `scripts/context_inventory.py` на те же roots (output вне roots).
Bootstrap читает только files из проверенного inventory и пишет только
`context/projects/<profile>/`. Он не обнаруживает test roots сам, не запускает
найденные команды без отдельного разрешения и не читает secret files.

Required outputs:

- `CONTEXT-MANIFEST.yaml`
- `PROJECT.md`
- `TEST-ARCHITECTURE.md`
- `COMMANDS.md`
- `CI-PIPELINE.md`
- `TEST-DATA.md`
- `TRACEABILITY.md`
- `SOURCE-MAP.md`
- `PATTERN-INDEX.md`

Каждый факт имеет source или `UNKNOWN`; profile фиксирует source commit SHA,
generated/verified time и refresh triggers.

## Manual colleague-feedback flow

1. Bootstrap создаёт `review-feedback/inbox/` и шаблон.
2. Человек вручную вставляет before/comment/after.
3. Curator проверяет redaction, provenance, applicability и reuse scope.
4. До human approval карточка остаётся candidate.
5. Project cards по умолчанию `project-only`; portable baseline получает только
   явно разрешённые обезличенные/абстрактные patterns.

## Prompt contract 2.0

- outcome-first;
- constraints/out-of-scope;
- required context/pattern refs;
- allowed tools и approval boundary;
- output schema;
- evidence;
- completion/stopping bar;
- краткое rationale/assumptions вместо раскрытия chain-of-thought;
- ordinary human request нормализуется, полный template обязателен только для
  delegation/handoff.

## Verification

Structural validator:

- links/frontmatter;
- skill name=folder;
- known GSD routes;
- prompt/recipe IDs;
- placeholders/profile manifest;
- explicit CoT disclosure phrases;
- secret-like patterns без печати значений;
- project leakage;
- trailing whitespace/provenance.

Eval dimensions:

- first-pass acceptance;
- human review minutes;
- defects and false positives;
- hallucinated paths/commands;
- scope escapes;
- traceability completeness;
- mutation sensitivity/flaky risk;
- token/tool/latency cost.

До реального прогона значения только `NOT_RUN`; «10x» без baseline запрещено.

## Waves

1. P0 truth/compatibility: provenance scan, skill names, GSD routes, links, CoT.
2. Context bootstrap/refresh + manual-feedback schemas.
3. Agents/recipes/model router/adapters.
4. Validator/evals/primary-source registry.
5. Management docs + portable dry-run.
6. Full sweep `.cursor/`, `pro-qa-ai/`, `.planning/`.
7. Independent review, docs sync, final acceptance.

## Acceptance

| Gate | Status | Evidence |
|------|--------|----------|
| Bootstrap создаёт profile только по roots | PASS | Hub manifest complete, 51 provenance records |
| Source mutation / closed MR access | PASS | 0 / 0; manifest policy |
| Manual example candidate→approved | NOT_RUN | templates/gates exist; human example intentionally absent |
| Portable core без project-only data | PASS | generated package excludes Hub profile/smoke/examples |
| Agent Skills contract | PASS | validators 0/0 |
| Explicit private reasoning disclosure absent | PASS | validator |
| Phantom GSD routes / broken links | PASS | `.cursor` + PE validators |
| Source and packaged validators | PASS | 0 errors / 0 warnings |
| Runtime provenance | PARTIAL | configuration valid; smoke action/model trace self-reported |
| Paired effectiveness/ROI | NOT_RUN | pilot required |
| Hub profile smoke | PARTIAL | content grounding PASS; runtime trace PARTIAL |
| Physical empty-repo installation | NOT_RUN | package build/validation PASS only |
