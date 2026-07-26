# Playbook: corporate delivery

## Outcome

Провести natural-language product или engineering request через scope, design,
implementation, test automation, review, goal verification и handoff без
hidden context и silent side effects.

Используйте только стадии, нужные задаче. One-file fix не требует полной
ceremonial wave; security, architecture, migration и cross-role изменения —
требуют.

## 0. Intake and routing

Owner: `orchestrator`.

1. Нормализуйте запрос через
   `agents/templates/_MASTER-PROMPT-TEMPLATE.md`.
2. Загрузите релевантные generated context/pattern cards. Profile отсутствует →
   явный `bootstrap-test-context`; stale profile → `refresh-test-context`.
   Отсутствующая approved card остаётся disclosed gap или идёт в manual
   curation — никогда не изобретайте её.
3. Зафиксируйте outcome, scope, do-not-touch, evidence, done/stop и approval
   boundary.
4. Примените `playbooks/model-routing.md` к текущему runtime catalog.
5. Задавайте только вопросы, ответы на которые существенно меняют scope, risk,
   permissions или acceptance. Продолжайте независимую работу вокруг
   localized blocker.

Artifact: нормализованный task contract + routing decision/fallback log.

## 1. Product and architecture gates

Запускайте условно:

- `product-owner`: scope, SMART/MoSCoW, success metrics или явные Won't
  изменились.
- `architect`: module/framework options, API/data boundaries, ADR, deployment
  topology или irreversible trade-off изменились.

Framework decisions требуют primary-source compatibility evidence. Unsourced
benchmark claims не считаются evidence.

Approval gates:

- Owner approval для изменений scope/KPI/Won't.
- Architecture owner approval для accepted ADR или migration boundary.
- Security owner approval для exception/accepted risk.

Artifacts: product decision; ADR/boundary contract.

## 2. Plan

Owner: `orchestrator`; process: `skills/phase-plan/SKILL.md`.

Каждый atomic plan фиксирует owner, files, do-not-touch, dependencies,
expected artifact, new-behavior checks, regression scope, evidence, approval
boundary и stopping condition. Waves стройте только когда parallel file
overlap равен нулю.

Artifact: executable plan + wave map.

## 3. Execute waves

Process:

- `skills/wave-orchestrate/SKILL.md`
- `playbooks/wave-execution.md`
- `agents/templates/wave-delegate.md`

Типичные owners:

- `backend-dev` / `frontend-dev` — implementation плюс локальные new/regression
  loops
- `devops` — CI/runtime changes со smoke и rollback
- `debugger` — root cause до фикса, когда поведение неопределённо

Каждый агент возвращает outcome first, paths, evidence, assumptions, risks,
handoff и Verification. Model fallback — часть evidence, не hidden runtime
detail.

## 4. Test-automation lifecycle

Это portable GSD-like loop для нового или изменённого поведения:

```text
requirements/spec
  → qa-manual or test-case-generator
  → skills/test-verification/SKILL.md (independent)
  → skills/phase-plan/SKILL.md for non-trivial automation packages
  → qa-auto implementation
  → targeted test loop until green
  → affected regression loop until green
  → skills/anti-ai-check/SKILL.md
  → skills/naming-check/SKILL.md
  → test-code-reviewer
  → skills/verify-goal/SKILL.md
  → test-documenter/docs-writer handoff
```

Правила:

1. Нет automation от unapproved или invented expectation.
2. Упавший тест triage как code/test/infra/flaky до смены assertions;
   используйте `debugger` или `test-debugger` и `playbooks/debugging.md`.
3. Test implementation мапится на реальный TC/requirement и observable result.
4. Нет retry/sleep/skip как постоянного flaky fix.
5. New-behavior и affected-regression loops оба доходят до green, иначе
   outcome = FAIL.
6. Full project gate запускается на single-task handoff или wave integration
   согласно context card целевого проекта.

## 5. Review and security

1. `code-reviewer` делает readonly defect-first review продуктового diff.
2. `test-code-reviewer` ревьюит изменённый код тест-фреймворка.
3. `security-reviewer` обязателен для sensitive surfaces и accepted-risk
   изменений.
4. Findings возвращаются owning implementation role; после фиксов повторяется
   полный relevant gate.

Никакой открытый CRITICAL/HIGH finding не переносится в PASS молча. Accepted
risk требует owner, reason, due date/ticket и approval.

## 6. Goal verification

Owner: independent verifier или `orchestrator`; process:
`skills/verify-goal/SKILL.md`.

Для каждого success criterion приложите path, command/output, observed
behavior или approved decision. Verdict — PASS, FAIL или PARTIAL. Completed
tasks без promised outcome — не PASS.

## 7. Docs and handoff

- `docs-writer`: product/API/ops/delivery documentation
- `test-documenter`: документация тест-фреймворка и traceability
- `playbooks/handoff-state.md`: state note, blockers и next action

Handoff включает source revision, changed paths, checks, routing/fallback
log, open risks, approval status и один actionable next step.

## Stop conditions

Остановитесь и верните текущий artifact, когда:

- delegated done criteria и verification gates пройдены;
- material product/architecture/security choice требует approval;
- required evidence или runtime capability недоступны;
- следующее действие destructive, external, production-facing или outside
  scope;
- дальнейшая работа — adjacent polish, а не часть outcome.

## Delivery DoD

- [ ] Outcome и approval boundary явны
- [ ] Context/pattern cards и decisions referenced
- [ ] Model decisions используют actual catalog и имеют fallback logs
- [ ] New behavior и affected regression green или явно FAIL
- [ ] Required code/test/security reviews закрыты
- [ ] Goal verification мапит каждый criterion на evidence
- [ ] Documentation и handoff синхронизированы
