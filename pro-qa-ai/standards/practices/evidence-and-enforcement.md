# Evidence и уровень enforcement

Полезная практика не становится обязательным законом только потому, что её
описала известная компания или книга. Сначала определить тип evidence, затем —
допустимый механизм enforcement.

## Evidence labels

| Label | Что означает | Допустимый вывод |
|---|---|---|
| `CONTROL` | Механизм прямо ограничивает класс отказа | Можно сделать hard gate после проверки false positives |
| `NORM` | Стандарт, спецификация или контракт платформы | Обязателен внутри заявленного scope/version |
| `EMP` | Эмпирическое наблюдение | Калибровать локально; не выводить причинный закон |
| `REC` | Официальная/широко используемая рекомендация | Default с documented trade-off |
| `OPINION` | Мнение автора или школы | Вопрос для review, не gate |
| `POLICY` | Локальное решение владельца | Обязательно только в явно указанном scope |
| `DISPUTED` | Практика зависит от контекста или имеет сильные возражения | Decision record, не автоматический запрет |

Company engineering blog — evidence реальной практики этой компании, а не
универсальное доказательство. Observational study показывает корреляцию в
выборке, а не magic threshold для всех языков и проектов.

## Три режима

1. **Hard control** — защищает correctness, security, data/failure boundary;
   имеет точный scope, низкий false-positive rate и negative sentinel.
2. **Hybrid** — машина собирает сигнал/evidence, reviewer принимает решение.
3. **Review guidance** — smell или архитектурная эвристика. Она задаёт вопрос,
   но не определяет качество Boolean-правилом.

## Candidate control catalog

Контроль становится active только после binding в project profile с owner,
scope, tool, version, exception/expiry и проверяемым negative sentinel.

| IDs | Candidate controls | Режим по умолчанию |
|---|---|---|
| C01–C05 | secrets; pinned third-party actions; protected branch; TS strict; typed Python core | hard/profile-bound |
| C06–C09 | Go race target; parameterized SQL; server authz; cross-tenant denial | C07–09 hard, C06 hybrid по package risk |
| C10–C16 | deadlines; finite/classified retries; idempotency rationale; pagination; structured API errors; DB invariants | hard для C10/11/14–16; hybrid C12/13 |
| C17–C20 | realistic migration path; no sleep synchronization; isolated cleanup | C18–20 hard, C17 hybrid/release gate |
| C21–C26 | declared toolchain; immutable promoted artifact; provenance; blocking high-confidence scans; dependency/license policy | hard после CI binding |
| C27–C32 | semantic a11y; critical-flow a11y; cancellation; bounded fan-out; no mutable global test state; deterministic time/random | C27/28/31 hard; C29/30/32 hybrid |
| C33–C39 | least-privilege agents; approval for destructive/external writes; untrusted input cannot elevate; controls outside prompts; same CI; independent high-risk verification; generated dependency provenance | C33–37/39 hard; C38 hybrid |
| C40–C44 | trace context; secret-safe logs; terminal async state; replay semantics; migration recovery | C40/41 hard; C42–44 hybrid/release gate |

## Review guidance, не hard gate

- 25 logical lines, cognitive complexity, nesting, module/class size;
- количество dependencies, interfaces, mocks или comments;
- duplicate-code signal;
- SOLID, DRY, Rule of Three, composition/inheritance choice;
- POM/Component Object/action layer, Repository, factory/builder;
- one assertion/one physical call per Act;
- pyramid/trophy/honeycomb ratios, coverage или mutation magic number.

Machine check для этих пунктов выдаёт warning с контекстом. Reviewer спрашивает,
уменьшает ли решение cognitive load, coupling, change radius, feedback latency
и failure/security blast radius.

## Обязательный формат нового правила

```yaml
id: CXX
evidence: CONTROL | NORM | EMP | REC | OPINION | POLICY | DISPUTED
scope: exact paths/runtimes/versions
mode: hard | hybrid | review
owner: accountable team/person
detector: exact command or review question
negative_sentinel: fixture proving detector can fail
false_positive_policy: exception fields and expiry
source_ids: [SRC-...]
```

Нет scope, owner, sentinel или source status — это candidate, а не активный
контроль. Prompt-инструкция сама по себе не является security enforcement.

## Источники

- `SRC-GITHUB-ACTIONS-SECURE-001`, `SRC-OWASP-CODE-REVIEW-001`;
- `SRC-NIST-AI600-1-001`, `SRC-MCP-SECURITY-001`;
- `SRC-GOOGLE-GOOD-TEST-001`, `SRC-GOOGLE-DAMP-001`,
  `SRC-GOOGLE-DRY-001`, `SRC-GOOGLE-TEST-SIZES-001`;
- `SRC-PLAYWRIGHT-BP-001`, `SRC-SELENIUM-POM-001`.

