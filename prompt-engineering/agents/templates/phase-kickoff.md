# Template: phase-kickoff

```markdown
[OUTCOME]
Подготовить исполнимый и проверяемый план фазы «{{phase_name}}» для цели:
{{goal}}

[ROLE]
- Role: {{role}}
- Handoff to: {{handoff_to}}

[SCOPE AND CONSTRAINTS]
- Files/resources allowed: {{files}}
- Do not touch: {{out_of_scope}}
- Constraints: {{constraints}}
- Approval boundary: {{approval_boundary}}

[REFERENCES]
- Inputs / prior decisions: {{inputs}}
- Context cards: {{context_cards}}
- Pattern cards: {{pattern_cards}}

[ACCEPTANCE]
- Atomic plans with owner, files, depends, and success criteria
- Wave map has no parallel file overlap
- Verification and test-automation lifecycle are explicit where applicable
- Evidence required: {{evidence_required}}
- Stop when: plan is executable without hidden context; material blockers are listed

[OUTPUT SCHEMA]
1. Outcome and assumptions
2. Atomic PLAN table
3. Dependencies and waves
4. Verification/test strategy
5. Risks and approval gates
6. Handoff
7. Verification (rules/40)
```
