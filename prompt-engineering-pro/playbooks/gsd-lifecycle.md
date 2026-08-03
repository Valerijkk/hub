# Playbook: GSD Lifecycle (project-agnostic)

## Цель

Единый жизненный цикл: orient → discuss/spec → plan → execute/test →
review → goal-verify → handoff.

## Шаги

1. **Orient** — normalise natural-language input; load generated context/pattern
   cards. No profile → `skills/bootstrap-test-context/SKILL.md`; existing stale
   or reorganized profile → `skills/refresh-test-context/SKILL.md`.
2. **Discuss/Spec** — lock one falsifiable outcome, constraints, approval
   boundary, done and stop. Ask only material questions.
3. **Route** — `../playbooks/model-routing.md`; actual catalog snapshot +
   decision/fallback log.
4. **Plan** — `skills/phase-plan/SKILL.md`: atomic artifacts, files, depends,
   success criteria, evidence and regression scope.
5. **Execute** — `skills/wave-orchestrate/SKILL.md` +
   `playbooks/wave-execution.md`.
6. **Test** — lifecycle from `corporate-delivery.md` §4. Existing
   commands/skills only: `test-verification`, `anti-ai-check`, `naming-check`.
7. **Review** — relevant readonly role; sensitive work → `security-reviewer`.
8. **Verify** — `skills/verify-goal/SKILL.md` +
   `rules/40-verification-protocol.md`.
9. **Handoff** — `playbooks/handoff-state.md` + docs owner.

## Шаблон kickoff

См. `agents/templates/phase-kickoff.md` — слоты `{{phase_name}}`, `{{goal}}`, `{{constraints}}`.

## Чек-лист

- [ ] Goal одно предложение
- [ ] Context/pattern cards и approval boundary указаны
- [ ] Планы атомарны
- [ ] Routing decision uses current catalog
- [ ] New + regression test loops закрыты
- [ ] Verification не skipped
- [ ] Handoff самодостаточен

## Антипаттерны

- Сразу кодить без plan
- «Таски закрыты» без goal-verify
- Смешивать слои в одном плане
- Требовать полный шаблон от человека вместо нормализации
- Выбирать модель по старому ID или несourced capability claim
- Менять ассерт красного теста до code/test/infra/flaky triage

## Связь с rules/40

Этапы 1–10 verification вшиты в конец каждой фазы execute.
