# Model registry

`registry.template.yaml` — это schema, а не утверждение, что какая-то модель
доступна. Создавайте runtime-local registry из текущего каталога и храните
snapshot timestamp, restrictions, capability evidence, decision logs и
fallback logs.

Правила:

- Точные model IDs берутся только из наблюдаемого runtime catalog.
- Рейтинги capability, отличные от `unknown`, требуют source.
- Project evaluations фиксируют setup/date/metric; несвязанные public benchmarks
  сами по себе не доказывают fit.
- Core agent prompts никогда не импортируют model IDs из этого registry.
- Selection и fallback следуют `playbooks/model-routing.md`.
