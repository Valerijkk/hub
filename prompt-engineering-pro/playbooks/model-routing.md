# Playbook: model routing

## Outcome

Выбирайте модель из **фактического runtime catalog** для каждой задачи, с
воспроизводимым decision и fallback log. Core agents остаются model-agnostic.

Для активного Cursor profile default orchestrator — запись каталога,
отображаемая как **GPT-5.6 Sol Pro**. Resolve текущий model ID во время
runtime; не копируйте запомненный provider slug в core prompts.

## Sources of truth

1. Runtime catalog и account/team restrictions.
2. [`../models/registry.template.yaml`](../models/registry.template.yaml),
   заполненный из этого каталога и датированный.
3. Sourced capability evidence в registry. Rating без evidence остаётся
   `unknown`; marketing text — не benchmark.
4. Task contract: outcome, context size/classification, tools, risk, budget и
   latency need.

Cursor-specific discovery каталога:

- IDE: текущий model selector плюс team/plan restrictions.
- Cloud API: `GET /v1/models`, возвращает available IDs и supported
  parameter variants.
- Custom agents: `model: inherit` — default; конкретный ID — explicit pin и
  может быть недоступен при team/plan restrictions.

Официальные ссылки:

- https://cursor.com/docs/subagents
- https://cursor.com/docs/cloud-agent/api/endpoints#list-models
- https://cursor.com/docs/sdk/typescript

## Decision procedure

### 1. Classify the task

Фиксируйте:

- task class: orchestrate | code | review | research | debug | docs | test
- complexity: low | medium | high
- context need: small | medium | large | unknown
- required strengths: coding, review, research, long-context, tool use,
  structured output
- risk: routine | important | critical
- context data classification: public | internal | confidential | restricted
- latency и cost constraints
- independence requirement: same model allowed | different model preferred

Это routing inputs, а не утверждения о какой-либо модели.

### 2. Refresh candidates

Используйте только entries, наблюдаемые в текущем catalog snapshot. Отклоняйте
candidates, которые unavailable, blocked, не имеют required tool/mode или не
могут безопасно держать required context. Unknown required capability —
blocker для critical work и disclosed risk для lower-risk work.

Перед отправкой любого non-public context требуйте known и approved:

- provider и data processor;
- processing region;
- retention;
- training-use policy;
- allowed data classifications;
- human/policy approval source и expiry.

`unknown` блокирует internal/confidential/restricted context. Availability
модели — не data-processing approval.

### 3. Select

1. Начинайте с `inherit`: сохраняет known context orchestrator и избегает
   stale pins.
2. Override только когда catalog candidate имеет sourced evidence лучшего
   соответствия material task need (например coding, independent review,
   research, context, cost или latency).
3. Для critical review предпочитайте independent pass. Другая модель полезна
   только когда available и qualified; role/process independence обязательна
   даже при недоступности model diversity.
4. Не выбирайте по brand, recency, unsourced leaderboard или benchmark,
   измеренному в несвязанных условиях.
5. Модель, не прошедшая data-policy gate, отклоняется даже если это strongest
   technical candidate.

Runtime-supported model parameters принадлежат adapter/launcher. Используйте
только значения, advertised выбранной записью каталога; никогда не кладите их
в role prompts.

### 4. Log before launch

```yaml
decision_id: "<wave/task/run>"
catalog_snapshot_at: "<timestamp>"
task_class: "<class>"
risk: "<routine|important|critical>"
context_data_classification: "<public|internal|confidential|restricted>"
required:
  context: "<small|medium|large|unknown>"
  strengths: ["<strength>"]
  tools: ["<tool-or-mode>"]
candidates_considered:
  - model_id: "<observed catalog id>"
    fit: "<qualified|rejected|unknown>"
    evidence_refs: ["<registry evidence id>"]
selected:
  mode: "<inherit|explicit>"
  model_id: "<resolved id>"
  rationale: "<concise, evidence-based>"
  provider_processor: "<approved provider/processor>"
  region: "<approved region>"
  retention: "<approved retention>"
  training_use: "<allowed|prohibited>"
  data_policy_approval_source: "<human-or-policy record>"
  data_policy_approval_expires_at: "<timestamp-or-null>"
fallback_order: ["<resolved id>", "<inherit>", "<stop>"]
approval_needed: true
```

После launch дополняйте actual model, used fallback, reason и эффект на
verification. Никогда не логируйте secrets.

## Fallbacks

1. Preferred explicit candidate unavailable → следующий **qualified** observed
   candidate.
2. Нет лучшего qualified candidate → `inherit`.
3. Context не вмещается → split задачу или compact references; не drop inputs
   молча.
4. Critical task без qualified model/tool → stop и запросите approval или
   capability change.
5. Independent-model review unavailable → отдельный readonly pass, disclose
   same-model limitation и сохраните обычный approval gate.
6. Runtime падает на automatic fallback → зафиксируйте resolved model и
   переоцените, остаётся ли evidence valid.

## Approval boundary

Routing может выбирать только entries, authorized и для account access, и для
текущей context classification. Включение paid capacity, смена team policy,
отправка non-public context к provider/processor/region, смена retention или
training use, или ослабление independence gate требуют explicit, unexpired
human/policy approval.

## Done

- [ ] Catalog snapshot актуален и source зафиксирован
- [ ] Selected ID существует и required capabilities qualified
- [ ] Data classification и provider/region/retention/training policy approved
- [ ] Decision и fallback logs полные
- [ ] Никакой unsourced benchmark claim не повлиял на selection
- [ ] Critical independence/context/approval constraints удовлетворены
