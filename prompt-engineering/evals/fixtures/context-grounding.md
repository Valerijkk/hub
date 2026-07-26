---
fixture_id: EVAL-CTX-001
category: context-grounding
status: template
owner: automation-quality
oracle_status: TODO
---

# Context-grounding fixture

## Goal

Оценить, использует ли модель поставленные факты репозитория и provenance
вместо опоры на generic stack assumptions.

## Instantiate before use

- Filled context profile revision: `{{context_revision}}`
- Approved source paths: `{{source_paths}}`
- Canonical commands: `{{canonical_commands}}`
- Deliberate decoy paths или stale commands: `{{decoys}}`
- Question, требующий repository-specific facts: `{{task_question}}`
- Facts, намеренно отсутствующие в context: `{{unknown_facts}}`

## Prompt given to both arms

Answer the task using only the supplied context and approved repository sources. Cite the source path for each repository-specific fact. Use only canonical commands. Mark facts that cannot be established as `TODO(context)` and state what source would resolve them. Do not substitute common framework defaults.

Assisted arm дополнительно получает versioned toolkit material, определённый
eval manifest.

## Oracle

- Каждый repository-specific path, command, version, naming rule или quality
  gate трассируется к approved source.
- Stale или decoy material отклоняется по provenance и precedence.
- Missing facts остаются unknown, а не достраиваются из general knowledge.
- Ответ различает observed facts, derived conclusions и recommendations.
- Никакая команда не выполняется, когда fixture разрешает только read-only
  analysis.

## Seeded mutations

1. Добавь common default command, которого нет в репозитории.
2. Добавь stale context profile со старой source revision.
3. Переименуй один approved path, оставив правдоподобный старый path в prose.

## Scoring emphasis

Traceability, hallucinated paths или commands, first-pass acceptance, defects,
scope escapes, false positives, tool calls, tokens, latency и review minutes.
