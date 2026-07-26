# GSD → Prompt Engineering port

Переносимые (project-agnostic) аналоги паттернов GSD внутри `prompt-engineering/`.

## Playbooks

| Playbook | GSD-источник | Путь |
|----------|--------------|------|
| Lifecycle фаз | plan/discuss/execute/verify | [../playbooks/gsd-lifecycle.md](../playbooks/gsd-lifecycle.md) |
| Волны агентов | execute-phase / workstreams | [../playbooks/wave-execution.md](../playbooks/wave-execution.md) |
| Handoff / STATE | progress / pause / resume | [../playbooks/handoff-state.md](../playbooks/handoff-state.md) |

## Skills

| Skill | Путь |
|-------|------|
| phase-plan | [../skills/phase-plan/SKILL.md](../skills/phase-plan/SKILL.md) |
| wave-orchestrate | [../skills/wave-orchestrate/SKILL.md](../skills/wave-orchestrate/SKILL.md) |
| verify-goal | [../skills/verify-goal/SKILL.md](../skills/verify-goal/SKILL.md) |

## Templates

| Template | Путь |
|----------|------|
| phase-kickoff | [../agents/templates/phase-kickoff.md](../agents/templates/phase-kickoff.md) |
| wave-delegate | [../agents/templates/wave-delegate.md](../agents/templates/wave-delegate.md) |
| corporate-wave | [../agents/templates/corporate-wave.md](../agents/templates/corporate-wave.md) |

## Corporate agents

См. [../agents/README.md](../agents/README.md) — orchestrator, backend-dev, frontend-dev, architect, product-owner, qa-manual, qa-auto, code-reviewer, debugger, security-reviewer, devops.

## Hub-specific (не в PE)

Для Hub используй `.cursor/skills/gsd-ru/` и `.cursor/agents/` — там ворота rule 56 и TZ. PE остаётся обезличенным.
