---
name: team-lead-opus
description: DEPRECATED compatibility redirect — используй team-lead на GPT-5.6 Sol/Extra High. Не связан с external Claude Code Opus 5 и не operational escalation slot.
model: inherit
readonly: true
---

# team-lead-opus — deprecated

> **Compatibility redirect only (2026-07-26).** Этот файл не является operational routing target.

Оркестратор и финальная инстанция — [`team-lead`](team-lead.md) на **GPT-5.6 Sol / Extra High**
(`model: gpt-5.6-sol`). Downstream-модели `team-lead` выбирает динамически по
capability/cost/risk из Cursor catalog; для высокорисковых решений — независимый
challenge-pass и `*-ver`, не отдельный agent slot.

## Что делать вместо этого файла

1. Открыть [`team-lead.md`](team-lead.md) — §§5.2, §12 (downstream routing), §19 (GSD).
2. Маршрутизацию моделей — [`playbooks/team-workflow.md` §1.1](../playbooks/team-workflow.md).
3. Исторический A/B-контекст прежней model-specific эскалации — только в
   [`.planning/decisions/ADR-MODELS-GPT56-2026.md`](../../.planning/decisions/ADR-MODELS-GPT56-2026.md);
   не использовать как operative steps.

## Запрещено

- Назначать `team-lead-opus` как обязательный escalation slot.
- Копировать устаревшие процедуры эскалации из git history в active prompts.
- Менять merge policy: финальный push/merge в `master` — только пользователь.
