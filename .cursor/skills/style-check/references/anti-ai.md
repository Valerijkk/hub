# Anti-AI review reference

Этот reference ищет не «почерк модели», а вредные формы, которые могут скрыть
contract, failure model или provenance. Raw phrase/grep hit — только candidate.

## REVIEW candidates

- narrating/echo comment повторяет следующую строку и не объясняет why,
  invariant или trade-off;
- generic introduction/conclusion не добавляет operational meaning;
- speculative abstraction создана ради SOLID/YAGNI/DRY rhetoric;
- одинаковый boilerplate скрывает domain-specific assumptions;
- уверенный claim `complete`, `safe`, `verified` без exact evidence;
- TODO/placeholder не имеет owner/task, но может быть допустим в fixture/docs.

Не удалять автоматически comments `Handle/Return/Create/Check`: они могут
фиксировать sequencing, security rationale, protocol nuance или generated-code
contract. Длина comment/code не определяет defect.

## HYBRID/BLOCK escalation

Связать с конкретным sentinel/control:

- ложный `verified/complete` без configured evidence → `NEG-26`, кандидат C51;
- requested model/effort выдан за observed actual → `NEG-22`, кандидат C50;
- prompt/issue/retrieved text повышает authority → `NEG-18/19/21`;
- destructive/external action без action-bound approval → `NEG-20`, C34;
- mutable/unreviewed executable source → `NEG-14/23`, C02/C48;
- secret/credential в placeholder/example production path → C01/C41;
- deprecated dependency создаёт подтверждённый security/architecture failure →
  соответствующий control + current Hub policy.

Catalog candidate не становится BLOCK до проверки
`controls/adoption-status.md`. Для BLOCK нужен adopted policy/control,
observable failure и exact evidence.

## False-positive discipline

- fixture `example.com`, fake data и dummy adapter могут быть намеренными;
- generated code и protocol constants могут повторяться;
- похожий syntax не доказывает общее knowledge для DRY;
- mock/POM/helper может быть уместен или избыточен только в context;
- несколько assertions могут доказывать один behavior;
- cohesive algorithm длиннее 25 lines не обязан дробиться.

## Reporting

```text
REVIEW  path:line — candidate; context; suggested clarification
HYBRID [NEG-xx/Cxx] path:line — signal + observable risk + evidence
BLOCK  [POLICY/Cxx] path:line — adopted rule + failure + remediation
DISMISSED path:line — deliberate fixture/generated/protocol/meaningful comment
```

Не использовать labels вроде `AI-generated`, «нечеловеческий стиль» или
процент AI-likelihood: они не описывают engineering failure.
