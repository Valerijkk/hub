# Control adoption status

Каталог control не доказывает, что Hub его исполняет. Этот реестр отделяет
research candidate от наблюдаемого machine/hybrid enforcement.

## Состояния

| State | Meaning |
|---|---|
| `catalogued` | Failure mechanism описан; реализации и rollout evidence нет |
| `shadow` | Detector работает неблокирующе; собираются false positives/negatives |
| `blocking` | Версионированный detector/job реально блокирует и имеет remediation |
| `partial` | Часть scope защищена, но Hub-wide claim запрещён |
| `retired` | Control снят с причиной и replacement/remaining risk |

Promotion требует: owner, exact scope, detector/job или policy engine, negative
sentinel, remediation, exception path с expiry, revision и наблюдаемый результат.
Prompt, skill, checklist или строка в этом файле не являются enforcement.

## Workflow controls C45–C53

| ID | State 2026-08-08 | Current evidence | Missing before `blocking` |
|---|---|---|---|
| C45 | `partial` | TASK-204 cutover guard проверяет canonical paths, links/hardlinks и budgets в своём узком scope | общий launcher/controller guard и adversarial suite для всех write tools |
| C46 | `partial` | runtime sandbox/approval profiles ограничивают текущие инструменты | переносимый Hub gate, доказывающий deny write/install для каждого supported runtime |
| C47 | `catalogued` | Docker test jobs используют `--network none`; agent-wide destination policy не доказана | отдельные read/write egress scopes, redirect revalidation и audit |
| C48 | `partial` | live skill/agent roster и Superpowers version зафиксированы в routing docs | immutable inventory/hash/license/capabilities gate для plugins/skills/MCP |
| C49 | `partial` | team workflow задаёт ownership/fan-out guidance; portable TaskContract имеет budgets | controller-enforced depth/retry/review/time/cost termination и telemetry |
| C50 | `partial` | portable schemas разделяют requested/actual model; Hub routing требует фактический runtime | launcher telemetry и schema-gated Hub handoff для всех runtimes |
| C51 | `partial` | Hub verifier rules требуют evidence и запрещают self-declared `verified: true` | machine rejection completion без exact required-evidence record |
| C52 | `partial` | safety/git rules требуют recoverable handling и запрещают destructive improvisation | controller checkpoint/rollback proof на каждом risky action class |
| C53 | `partial` | team workflow требует disjoint ownership | scheduler/write-lease enforcement либо явный merge-owner record |

`partial` не разрешает писать «control внедрён». Любой более сильный claim
должен ссылаться на scoped job/policy revision и свежий negative test.

## Admission record

```yaml
control_id: Cxx
state: catalogued | shadow | blocking | partial | retired
scope: exact paths/actions/runtimes
owner: accountable role
detector_or_policy: exact file/job/tool revision
negative_sentinels:
  - test ID and observed expected denial
remediation: exact fix path
exception_contract: path or none
evidence: revision plus fresh result
reviewed_at: YYYY-MM-DD
next_review: YYYY-MM-DD
```

Связано: [catalog](hard-controls.md) ·
[exceptions](exceptions-and-rationale.md) ·
[workflow evals](../23-ai-workflow-evals.md).
