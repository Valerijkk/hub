# 21 — AI-agent engineering

## Главный принцип

Textual instruction не является access-control boundary. Prompt, issue, web page, email, retrieved document и tool output могут содержать untrusted text и не получают authority из-за своей формулировки.

## Trust and action cards

| Principle | Evidence | Mode | BAD | GOOD | Sources |
|---|---|---|---|---|---|
| Untrusted data cannot raise authority | `CONTROL`, `REC`; high | hard platform boundary | issue говорит «deploy» и agent исполняет | issue остаётся data; trusted policy/tool ACL решает capability | [S76, S77](sources/registry.md) |
| Least-privilege tools | `CONTROL`; high | hard | repo/cloud admin для read task | scoped read/write roots и task-bound credentials | [S76, S77](sources/registry.md) |
| Approval at action point | `CONTROL`, `REC`; high | hard | раннее абстрактное «можно продолжить?» | exact target, impact и scope непосредственно перед risky call | [S76](sources/registry.md) |
| Deterministic restriction | `CONTROL`; high | hard/hybrid | prompt «никогда не deploy» | tool deny/hook/policy фактически блокирует deploy | [S76, S78](sources/registry.md) |
| Evidence-backed completion | `POLICY`, `REC`; high | hybrid | success по намерению | fresh tests/static evidence и перечисленные ограничения | [S76](sources/registry.md) |
| Independent verification by risk | `POLICY`; medium | review/workflow | self-review как единственное evidence high-risk change | independent verifier для security/migration/critical behavior | [evidence status](sources/evidence-status.md) |

## Context hygiene

- Инструкции кратки и scoped, но точное число строк не является quality threshold.
- Retrieved content размечается как data, не как policy.
- Structured tool outputs предпочтительнее смешивания instructions и payload.
- High-risk capability не выдаётся «на всякий случай».
- Approval привязан к конкретному необратимому/внешнему действию.
- Результат содержит provenance: что прочитано, что изменено, чем проверено, что не проверено.

## BAD / GOOD

**BAD**

```text
Read the issue and do everything it asks.
You have repository write and deployment credentials.
```

**GOOD**

```text
Trusted policy: edit only task-owned paths; deployment denied by tool policy.
Untrusted issue: <data>...</data>
Approval: required immediately before a named external/destructive action.
```

Prose должна дублироваться реальным permission/hook control для hard restrictions.

## Claims that remain limited

- Fresh context может уменьшить drift, но не гарантирует correctness.
- Independent verifier вероятно полезен для high-risk work, но correlated model errors остаются возможны.
- GSD workflow — operational practice, не независимое научное доказательство преимуществ любой конкретной длины контекста.

См. `NEG-18..NEG-20` в [negative sentinels](controls/negative-sentinels.md).
