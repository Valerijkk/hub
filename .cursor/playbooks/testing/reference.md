# Hub QA reference

Этот файл сохраняет domain knowledge без привязки к legacy test paths. Exact
expected всегда сверяется с current `docs/`, ADR/API и task.

## Риск-якоря по доменам

| Домен | Значимые failure modes |
|---|---|
| auth | Magic Link expiry/replay/timing, rate limit, session revocation, founder/tenant boundary |
| rbac/admin | server-side deny, cross-tenant object access, audit/no forbidden side effect |
| tracker/sprint | optimistic version, invalid transition, duplicate active sprint, soft-delete visibility |
| messenger/ws | membership, read-only channel, ordering/replay, reconnect/cancellation, backpressure |
| tms | external ID idempotency, import partial failure/recovery, run transition, traceability |
| calls | LiveKit webhook authenticity, guest boundary, orphan room compensation, reconnect/recovery |
| search | user/project/channel scope, Unicode, pagination and stale index |
| notification | quiet hours/urgent policy, retry/terminal state, redaction, invalid endpoint |
| bases/accounting/CRM/HR | tenant/RBAC, optimistic version, invariants, soft-delete/retention, audit |
| integrations/webhooks | timeout, finite retry, signature, duplicate delivery, idempotency, partial failure |

Это checklist для выбора риска, не quota cases. Значения лимитов и status codes
не копировать отсюда без актуальной normative ссылки.

## Review rubric

### BLOCK

- неверный target path/job или legacy port;
- отсутствует/не разрешается task, requirement или TC;
- oracle не проверяет заявленный behavior;
- cross-tenant/authz/security risk без необходимого negative evidence;
- zero collection, unexpected skip, leaked resource;
- host-run выдан за canonical evidence;
- `verified: true` без exact Docker result;
- historical coverage/parity использованы как acceptance gate.

### HYBRID

- C06 race evidence, C09 tenant negative, C19 fixed wait, C20 cleanup,
  C29 cancellation, C31 shared state, C32 time/randomness;
- machine signal обязателен, но reviewer подтверждает scope и observable
  failure. Catalog entry сам по себе не доказывает blocking adoption.

### REVIEW

- выбор scope/size и suite mix;
- число assertions для одного behavioral concept;
- DAMP/DRY balance;
- stub/fake/mock и interaction oracle;
- helper/Component Object/Page Object;
- количество cases и применимость scenario family.

В REVIEW нельзя превращать «25 строк», SOLID, DRY, POM, mocks, Pyramid/Trophy/
Honeycomb или raw grep в автоматический defect.

## Вердикт qa-auto-ver

```text
QA VERDICT: verified: true|false
Scope: <jobs/TC/requirements>
Evidence: <source SHA, images, exact Docker commands, results, duration, artifacts, cleanup>
BLOCK: <finding or none>
HYBRID: <Cxx/NEG-xx + observable failure/evidence>
REVIEW: <contextual finding and rationale>
Unverified: <explicit residual scope>
```

## Triage reminder

Сначала проверить требуемое и наблюдаемое product behavior по docs/repro.
Затем только одна классификация: `ROOT=code`, `ROOT=test`, `ROOT=infra`,
`ROOT=flaky`, `ROOT=stop`. См.
[`test-triage`](../../skills/test-triage/SKILL.md).

## Исторический контекст

Legacy Go component/API/integration, static-contract Vitest и TypeScript
Playwright были inventory для exact delete/retain manifests. Их paths, counts,
POM/Steps, `hubtest` commands и 82% coverage не являются active guidance.

Связанные sources:

- [`docs/test_plan.md`](../../../docs/test_plan.md)
- [`test architecture rebase`](../../../.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md)
- [`15-test-automation.md`](../../knowledge/code-practices/15-test-automation.md)
- [`hard-controls.md`](../../knowledge/code-practices/controls/hard-controls.md)
- [`negative-sentinels.md`](../../knowledge/code-practices/controls/negative-sentinels.md)
- [`adoption-status.md`](../../knowledge/code-practices/controls/adoption-status.md)
