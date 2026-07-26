# Approved pattern cards

Карточки здесь имеют `status: approved` и удовлетворяют всем gates из
[`../SCHEMA.md`](../SCHEMA.md).

Требуется перед promotion:

- явное human approval;
- verified provenance включая source path, lines и commit;
- bad/good code плюс reviewer rationale;
- конкретные applicability и non-applicability;
- distinguishing test или evidence;
- завершённый human redaction review;
- известные license, permission reference и reuse scope;
- актуальные freshness metadata.

Approval не даёт право на redistribution. Карточки `project-only` и
`organization` остаются overlays и должны исключаться из portable exports.
Карточка с любым обязательным полем `UNKNOWN` возвращается в
[`../candidates/`](../candidates/).

Deprecated cards могут оставаться здесь со `status: deprecated`, чтобы
существующие profile indexes находили replacement и rationale.
