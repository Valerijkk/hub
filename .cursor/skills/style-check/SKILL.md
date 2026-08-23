---
name: style-check
description: Evidence-based review с раздельными BLOCK/HYBRID/REVIEW. Raw search собирает кандидатов, но не превращает 25 lines, SOLID/DRY/POM/mocks или grep hit в автоматический defect.
disable-model-invocation: true
---

# /style-check — evidence-based style and risk review

Цель — найти конкретные correctness/security/maintainability failures без
cargo-cult gates. Канон evidence:
`.cursor/knowledge/code-practices/{00-principles.md,01-evidence-model.md}`,
controls: `hard-controls.md`, `review-guidelines.md`,
`negative-sentinels.md`, `adoption-status.md`.

Tests, validators, linters, build/vet и app processes выполняются только в
fixed `docker compose -f docker-compose.local.yml run --rm <exact-service>`. Host `rg` допустим
только как read-only просмотр кандидатов; его output не является canonical
validator PASS/FAIL и не доказывает defect.

## 1. Классификация

### BLOCK

Только когда есть:

1. точный locked policy или control реально `blocking` в adoption status;
2. конкретный failure class и применимый scope;
3. machine/runtime evidence, отличающий violation от допустимого случая;
4. exact path:line, remediation и exception path.

Примеры: confirmed committed secret (`C01`), exploitable SQL construction
(`C07`/`NEG-12`), confirmed missing server authz (`C08`/`NEG-13`), founder
Docker-only или test-target violation. Catalog row без adoption evidence не BLOCK.

### HYBRID

Machine signal обязателен, семантику подтверждает reviewer. Finding содержит
`Cxx` и/или `NEG-xx`, observable failure, evidence и remaining uncertainty.
Примеры: timeout/retry/idempotency (`C10–C13`, `NEG-01/09/16`), cleanup/shared
state (`C20/C31`, `NEG-07`), cancellation/concurrency (`C29/C30`,
`NEG-10/11`), arbitrary test sleep (`C19`, `NEG-03`).

### REVIEW

Context-dependent design question. Нужен change/failure model и concrete
rationale, но raw shape не блокирует:

- function/module size, включая `>25 logical lines`;
- SOLID, DRY, YAGNI, composition/inheritance/DI;
- число assertions и test length;
- DAMP/DRY balance;
- stub/fake/spy/mock и interaction assertions;
- helper/Component Object/Page Object;
- Pyramid/Trophy/Honeycomb suite mix;
- comment wording, abstraction и naming.

## 2. Workflow

1. Определить changed scope, contract, owner и risk.
2. Прочитать [anti-ai reference](references/anti-ai.md).
3. Собрать raw candidates read-only; не считать count verdict.
4. Проверить каждый candidate в контексте code, callers, tests и docs.
5. Связать finding с locked policy, `Cxx`/`NEG-xx` или review question.
6. Назвать observable failure и supporting evidence; если их нет — оставить
   question/telemetry gap, не defect.
7. Проверить adoption status перед словами `hard`, `blocking`, `enforced`.
8. Machine gates запустить только Docker-only; при запрете — `NOT_RUN`.

## 3. Required review areas

- secrets, injection, authn/authz/tenant isolation, redaction;
- timeout, finite retry, idempotency, cancellation, bounded concurrency;
- ownership/lifetime, mutable global state, cleanup;
- API/error/schema/event compatibility;
- accessible semantic controls and critical flow evidence;
- test determinism, observable oracle, false-confidence sentinel;
- agent authority/provenance where agent workflow changes.

Hub deprecated technology (`minio-go`, Gorilla WebSocket) — candidate until
confirmed in active production path and contradicted by current architecture.
Fixture, generated, migration or historical references are not automatic defects.

## 4. Forbidden reviewer shortcuts

- `rg`/grep hit = finding;
- 25 lines = split;
- SOLID = interface/class quota;
- DRY = merge similar syntax;
- one behavior = one assert;
- mock/POM absent = defect, or present = quality;
- Pyramid/Trophy/Honeycomb ratio = gate;
- catalogued control = implemented control;
- style preference overrides correctness/security/data integrity.

## 5. Report

```text
STYLE REVIEW: PASS|FINDINGS|NOT_RUN
Scope: <paths/change contract>
Docker validators: <exact evidence or NOT_RUN>

BLOCK
- [Cxx|POLICY] path:line — observable failure; evidence; remediation

HYBRID
- [Cxx/NEG-xx] path:line — signal; semantic evidence; uncertainty; owner

REVIEW
- path:line — question tied to change/failure model; recommendation

Dismissed candidates
- pattern — why false positive/allowed context
```

`PASS` означает отсутствие подтверждённых findings в просмотренном scope, а не
Hub-wide compliance. Raw candidate count можно приложить отдельно, не как score.
