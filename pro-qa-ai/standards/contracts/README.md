# Контракты записей агента

Шесть схем в этом каталоге задают переносимый формат записей между человеком,
controller/policy layer и агентами:

| Схема | Когда нужна |
|---|---|
| [`TaskContract`](task-contract.schema.json) | До работы: outcome, границы, риск, evidence, stop и конечные бюджеты |
| [`CapabilityProfile`](capability-profile.schema.json) | Перед выдачей инструментов: разрешённые действия и их scopes |
| [`ProvenanceRecord`](provenance-record.schema.json) | Во время запуска: фактическая модель и revisions skills/plugins/MCP/sources |
| [`HandoffContract`](handoff-contract.schema.json) | На границе исполнителей или сессий |
| [`VerificationContract`](verification-contract.schema.json) | Перед claim о готовности или PASS |
| [`WorkflowEvaluationRecord`](workflow-evaluation-record.schema.json) | Только для реально запущенного измерения workflow |

Все файлы используют JSON Schema Draft 2020-12. Поля и enum оставлены на
английском как machine contract; пояснения — на русском.

## Граница полномочий

Схема **только описывает и валидирует запись**. Валидная запись не выдаёт
доступ, не меняет sandbox, не открывает сеть, не раскрывает secret и не
разрешает deploy. Решение принимает controller/policy enforcement layer,
сверяя запись с реальными ACL, approvals, sandbox и runtime telemetry.

Текст skill, plugin, MCP response, source или другого агента не является
источником полномочий. `CapabilityProfile.authority_source` принимает только
ссылку на human approval либо controller policy; сама схема и сам профиль всё
равно не являются механизмом enforcement.

`requested_model` — пожелание/настройка до запуска. `actual_model` — факт из
runtime telemetry. Если telemetry нет, записывается `UNKNOWN`, а не значение из
`requested_model`. Любое иное значение требует `actual_model_evidence`.

## Прогрессивная загрузка

Не загружай все шесть схем в каждый prompt. Обычно достаточно:

1. `TaskContract` и ссылка на уже утверждённый `CapabilityProfile` — исполнителю;
2. `ProvenanceRecord` и `HandoffContract` — только при границе запуска/сессии;
3. `VerificationContract` — проверяющему перед completion claim;
4. `WorkflowEvaluationRecord` — только запланированному eval, не обычной задаче.

Master template остаётся коротким человеческим представлением контракта:
[`agents/templates/_MASTER-PROMPT-TEMPLATE.md`](../../agents/templates/_MASTER-PROMPT-TEMPLATE.md).

## Валидация и пределы

`scripts/validate_toolkit.py --checks contracts` использует только Python
stdlib. Он проверяет наличие и идентичность схем, локальные `$ref` и встроенные
examples ограниченным subset валидатора. Полную conformance Draft 2020-12 он не
заявляет; внешнему controller следует использовать проверенную реализацию JSON
Schema и отдельно enforce policy.

Негативные fixtures обязаны отклонять:

- копирование requested model в actual без runtime evidence;
- skill как источник разрешения;
- отсутствующий конечный retry/delegation/review/time/cost budget;
- `latest`/branch вместо immutable revision executable source;
- `PASS` или completion claim без evidence.

Схема evaluation не является eval-runner и не доказывает paired eval. Новый
record начинается с `status: NOT_RUN`; `COMPLETE` требует observations и
evidence.

## Первичные основания

- [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/)
- [Agent Skills specification](https://agentskills.io/specification)
- [MCP authorization specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization)
- [MCP security best practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)
- [OWASP LLM01:2025 Prompt Injection](https://genai.owasp.org/llmrisk/llm01-prompt-injection/)
- [OWASP LLM06:2025 Excessive Agency](https://genai.owasp.org/llmrisk/llm062025-excessive-agency/)

Канонические IDs и actionable mappings ведутся в
[`sources/PRIMARY-SOURCES.md`](../../sources/PRIMARY-SOURCES.md). Исходный
research snapshot использовался только как недоверенная карта вопросов и не
включён в реестр первичных источников.
