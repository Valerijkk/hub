# Layer 04 — Team Orchestration

Источники: `.cursor/playbooks/team-workflow.md`, `.cursor/agents/team-lead.md` §5,
Pro `playbooks/wave-execution.md`, `agents/orchestrator.md`.

## Ростер Pro (portable)

| Группа | Роли (Standard agents/) |
|--------|-------------------------|
| Оркестрация | orchestrator, product-owner, architect |
| Build | backend-dev, frontend-dev, devops |
| Quality | qa-manual, qa-auto, code-reviewer, debugger, security-reviewer |
| Test-framework | test-case-generator, test-code-reviewer, test-debugger, test-documenter |
| Docs | docs-writer |

Hub имеет 32 агента с `*-ver` и параллельными потоками — это **Hub runtime**.
В Pro достаточно roster Standard + правила ниже.

## Правила волны

1. Один orchestrator на волну; downstream `inherit` модель.
2. Параллель только при disjoint file sets.
3. Общий контракт (API/схема) фиксируется **до** параллели.
4. Кросс-ревью пары до независимого verifier на critical/security.
5. Atomic commit; **запрет** AI co-author trailers (portable policy).

Extract: [`../extracts/wave-protocol.md`](../extracts/wave-protocol.md).

## DoD слоя

Handoff волны содержит: состав ролей, file allowlist, evidence bar, stop condition.
