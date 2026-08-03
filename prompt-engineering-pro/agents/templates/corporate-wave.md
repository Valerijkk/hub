# Template: corporate-wave

Мульти-ролевая волна корпоративной доставки. Роли без работы в конкретной
волне не запускаются.

```markdown
[OUTCOME]
- Wave: {{wave_name}}
- Goal: {{goal}}
- Done / stop: {{done_definition}} / {{stopping_condition}}

[REFERENCES]
- Context cards: {{context_cards}}
- Pattern cards: {{pattern_cards}}
- Decisions / inputs: {{inputs}}

[WAVE MAP]
| Order | Role | Outcome | Files | Depends | Handoff |
|-------|------|---------|-------|---------|---------|
| {{n}} | {{role}} | {{agent_outcome}} | {{files}} | {{depends}} | {{handoff}} |

[ROUTING]
- Orchestrator runtime: {{orchestrator_model}}
- Downstream selection: ../../../playbooks/model-routing.md
- Default adapter behavior: inherit; override only from an observed catalog entry
- Decision/fallback log: {{routing_log}}

[BOUNDARIES]
- Do not touch: {{out_of_scope}}
- Approval boundary: {{approval_boundary}}
- Parallel entries must have zero file overlap; shared files are serialized
- Every role returns evidence and Verification
- Orchestrator accepts by goal-backward criteria, then writes handoff-state
```
