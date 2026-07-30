---
name: wave-orchestrate
description: Запустить запланированную волну агентов с disjoint file maps, структурированными промптами делегирования и собранным evidence. Использовать после phase-plan, когда PLAN готов и задачи могут идти параллельно без конфликтов путей.
disable-model-invocation: true
---

# Оркестрация волны (`wave-orchestrate`)

Выполнить одну волну из PLAN: делегировать агентам с явным scope, собрать outputs, разрешить конфликты, затем передать в `verify-goal`.

## Когда использовать

- PLAN существует с задачами текущей волны и file map
- Пересечение файлов внутри волны = **0**
- У каждой задачи есть owner и verifier (или задокументированный N/A для docs-only)

## Когда остановиться

- Обнаружено пересечение файлов → **остановиться**; вернуться в `phase-plan`
- В промпте агента нет scope/evidence/stopping condition → не запускать; починить шаблон делегирования
- Конфликт в merged outputs → остановить волну; replan или сериализовать конфликтующие задачи

## Контекст (загрузить сначала)

| Нужно | Primary | Legacy fallback |
|------|---------|-----------------|
| PLAN | User/path из выхода phase-plan | — |
| Delegation | `agents/templates/wave-delegate.md` | — |
| Commands / regression | `context/projects/<profile>/COMMANDS.md`, `CI-PIPELINE.md` | PROJECT-CONTEXT, TESTFRAMEWORK-MAP |
| Pattern/context cards | `PATTERN-INDEX.md` + `patterns/approved/` | PATTERNS-LIBRARY |

## Outcome-first шаги

1. **Загрузить срез волны** из PLAN (task IDs, owners, files only / do not touch, depends удовлетворены).
2. **Проверка конфликтов** — пересечение `files only` внутри волны должно быть пустым.
3. **Делегировать** — один заполненный `wave-delegate.md` на агента (не free-text «do X»). Обязательные слоты:
   - Outcome (одно проверяемое предложение для этой задачи)
   - Files only / do not touch
   - Acceptance: done definition, проверки new-behavior, `{{regression_scope}}` из profile
   - Evidence required; stop when критерии выполнены или blocked
   - Output schema §1–7 включая блок Verification ([rules/40](../../../prompt-engineering/rules/40-verification-protocol.md))
4. **Параллелизм** — по умолчанию ≤4 агента; больше только с явным disjoint map.
5. **Собрать** — каждый агент возвращает outcome, paths touched, evidence, open questions, блок Verification.
6. **Sync** — если один path изменён дважды или contracts конфликтуют → остановиться; не мержить молча.
7. **Регрессия** — прогнать wave-level regression scope один раз на integration (не N дубликатов полного gate на агента, если PLAN не требует).
8. **Handoff** — `verify-goal` по success criteria волны; verifiers получают своё scoped delegation, если PLAN их назначает.

## Минимум делегирования (из шаблона)

Использовать полный шаблон `agents/templates/wave-delegate.md`. Каждый промпт должен включать:

- `[OUTCOME]` — фальсифицируемый для этой задачи
- `[SCOPE AND CONSTRAINTS]` — files, out of scope, approval boundary
- `[ACCEPTANCE]` — done, new-behavior checks, regression scope, evidence, **stop when**
- `[OUTPUT SCHEMA]` — включая Verification (rules/40); без private reasoning traces

**Не** включать chain-of-thought instructions или temperature settings в промпты делегирования.

## Чек-лист sync волны

```markdown
## Wave sync — {{wave_id}}

- [ ] All delegations used wave-delegate template
- [ ] Zero file overlap within wave
- [ ] Each agent returned Verification block
- [ ] Domain tests green (logs attached) or NOT_RUN with reason
- [ ] No unresolved path conflicts
- [ ] Ready for verify-goal: {{criteria IDs from PLAN}}
```

## Условия остановки

- Все задачи волны соответствуют done definition → переходить к `verify-goal`
- Любой открытый CRITICAL из review/security → остановиться; fix или replan
- Нет evidence для заявленного PASS → остановиться; перезапустить задачу с явным запросом evidence

## Handoff

→ `verify-goal` с PLAN success criteria для этой волны
→ При FAIL: целевое re-delegation или replan через `phase-plan`
→ При PASS: `playbooks/handoff-state.md`, если проект использует formal handoff

## Анти-паттерны

- Два агента на одном файле
- Волна без sync point или verify-goal
- Принятие «looks good» без блока Verification
- Изобретённые regression commands не из profile

## Ссылки

- Planning: `../phase-plan/SKILL.md`
- Acceptance: `../verify-goal/SKILL.md`
- Playbook: `../../playbooks/wave-execution.md`
- Verification: [rules/40](../../../prompt-engineering/rules/40-verification-protocol.md)
