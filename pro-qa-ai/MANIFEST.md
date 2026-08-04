---
manifest_version: 1
product: pro-qa-ai
toolkit_version: 0.4.0-beta.1
status: foundation-ready
owner: automation-quality
layer: test-automation-ai-quality
source_registry: sources/PRIMARY-SOURCES.md
validation_config: scripts/validation_manifest.json
validation_status: STRUCTURAL_PASS
last_validation: 2026-08-04
provenance: official primary sources and repository inspection on 2026-07-26

---

# Манифест — pro-qa-ai

## Идентичность

Этот дистрибутив — **pro-qa-ai**: AI quality layer для инженеров
автоматизации тестирования. Он даёт переносимые операционные инструкции, структурную валидацию,
провенанс и правила владения вокруг уже существующей среды автоматизации. Продуктовая карта: [TOUCHSTONE.md](TOUCHSTONE.md).

Это не scaffold тестового фреймворка. Он не выбирает runners, не создаёт
архитектуру фреймворка и не владеет проектными командами и путями.

Статус — **foundation-ready**. `STRUCTURAL_PASS` = дерево и валидатор прошли
структурные проверки. Это **не** значит eval-validated или GA.

Последнее структурное доказательство на 2026-08-04 (перезамер после
схлопывания слоя ревью):

- валидация исходного тулкита: 0 ошибок / 0 предупреждений, exit 0;
- тесты скриптов: 41 passed / 1 skipped, 42 всего (создание symlink
  недоступно; эквивалентные mocked-тесты Windows reparse проходят);
- переносимый пакет: 181 файл, SHA-256 инвентаря записан.
  Сгенерированный `context/projects/hub/`, локальные прогоны и сырые
  треды ревью в пакет не входят — проверено поломкой, а не чтением
  конфига.

Детали команд и проверка целостности состава — [VERSION.md](VERSION.md).

Метрики эффективности остаются `NOT_RUN`: парный замер не прогонялся.

## Владение слоями

| Слой | Содержимое | Ответственный владелец | Граница approval |
|---|---|---|---|
| Portable operating content | `agents/`, `skills/`, `standards/`, `workflows/` | automation quality | framework и security review для изменений контракта |
| Project context/knowledge | `context/`, `standards/patterns/`, `standards/team-review/` | владелец фреймворка целевого репозитория | generated facts и ручные examples требуют provenance/redaction/reuse scope |
| Structural validation | `scripts/` | automation quality | security утверждает secret/injection patterns и exceptions |
| Source registry | `sources/` | automation quality | domain owner ревьюит изменённые mappings |
| Management controls | `docs/` | engineering management | approval governance, data, provider и rollout |
| Version and release | `VERSION.md`, `MANIFEST.md` | automation quality lead | release gates и accountable sign-off |

## Инвентарь quality layer

- [Context bootstrap](skills/bootstrap-test-context.md) и [контракты профилей](context/projects/README.md)
- [Pattern cards](standards/patterns/README.md) и [треды ревью коллег](standards/team-review/README.md)
- [Production recipes](workflows/recipes/README.md)
- [Structural validator](scripts/validate_toolkit.py)
- [Portable packager](scripts/package_portable.py) и [политика упаковки](scripts/package_manifest.json)
- [Mechanical context inventory](scripts/context_inventory.py), его
  [config](scripts/context_inventory_config.json) и
  [schema](scripts/context_inventory.schema.json)
- [Конфигурация валидации](scripts/validation_manifest.json)
- [Unit-тесты валидатора](scripts/tests/test_validate_toolkit.py)
- [Primary sources](sources/PRIMARY-SOURCES.md)
- Управленческие документы: [обзор](docs/EXECUTIVE-OVERVIEW-RU.md),
  [демо для руководителя](docs/LEADERSHIP-DEMO-RU.md),
  [техническое демо](docs/DEMO.md), [метрики ROI](docs/ROI-METRICS.md),
  [правила владения](docs/GOVERNANCE.md)
- [Version](VERSION.md)

## Контракт структурной валидации

Stdlib-валидатор проверяет:

- ограниченный approved UTF-8 text, лимиты per-file/total/count, containment и
  отклонение symlink/junction/Windows reparse;
- содержащиеся relative Markdown link targets и GitHub-style heading fragments;
- базовый YAML-like синтаксис frontmatter и обязательные frontmatter paths;
- обязательные поля Agent Skills, синтаксис naming и равенство name-to-parent;
- inferred prompt IDs для numbered файлов prompt-library, включая duplicates и gaps;
- unresolved placeholders в generated profiles при допуске template slots;
- явные инструкции, требующие раскрытия private reasoning;
- неизвестные GSD route names относительно configurable route registry;
- secret-like patterns с redacted diagnostics;
- project-specific leakage в portable core по reviewed deny и allow rules;
- резолв путей в inline-code (бэктики), а не только markdown-ссылок;
- trailing whitespace;
- поля provenance тулкита и generated-context.

Валидатор намеренно уже полного Markdown parser, YAML parser, secret scanner или
reference implementation Agent Skills. Его пределы должны оставаться видимыми
на release review.

PowerShell:

```powershell
py -3 -B .\pro-qa-ai\scripts\validate_toolkit.py `
  --root .\pro-qa-ai

py -3 -B -m unittest discover `
  -s .\pro-qa-ai\scripts\tests `
  -p "test_*.py" `
  -v
```

Exit `0` обязателен от первой команды на точном release distribution. Успех
unit-тестов сам по себе недостаточен.

## Configurable registries

`scripts/validation_manifest.json` — машиночитаемый источник политики для:

- обязательных frontmatter и skill fields;
- prompt-ID inference;
- состояний template versus generated-profile;
- context-manifest fields;
- forbidden reasoning-disclosure patterns;
- known GSD routes;
- общих secret-like patterns, сканируемых по материалу validator/config/test
  без широких path exclusions;
- portable-core includes, deny patterns и узких allow rules;
- toolkit manifest fields.

Любое allow rule — это policy exception и требует reason и review. В будущем
процессе также следует фиксировать approver и expiry при добавлении allow rule.

## Generated context и provenance

Нейтральные templates живут только в `context/templates/`; они могут держать
placeholder slots только при `status: template`. Generated
`context/projects/<profile>/` каталоги никогда не являются входами пакета.

Каждый generated `context/projects/<profile>/` должен содержать `CONTEXT-MANIFEST.yaml` и восемь документированных Markdown-артефактов. Generated Markdown frontmatter требует:

- `context_schema`;
- `profile_slug`, равный каталогу профиля;
- non-template `status`;
- `generated_at`;
- `snapshot_commit`;
- отсутствие unresolved placeholder slots.

YAML-манифест — запись provenance. Валидация требует identity/status профиля,
declared roots, scan status и trust boundary, отдельных metadata/project
command-use flags, filesystem policy и per-item source/output metadata.
Complete profiles требуют typed SHA-256, size и timestamps; документированные
invocation/filesystem-inventory records alone могут использовать явный
обоснованный `N/A`. Legacy aggregate `commands_executed` отклоняется. Generated
manifest может использовать явный `UNKNOWN` по контракту context, пока partial,
но не может сохранять template placeholders.

Bootstrap и refresh принимают только verified inventory от `context_inventory.py`
для точных declared project/test roots. AI-интерпретация ограничена
перечисленными `files[]`; произвольный file discovery вне trust boundary.

Legacy top-level context documents deprecated: read-only compatibility
артефакты. Bootstrap и refresh их не пишут. Generated legacy profiles по-прежнему
требуют заполненных identity/date, provenance, source revision, отсутствия
placeholders и своего legacy context manifest.

## Граница portability

Portable core может содержать generic roles, workflows, quality controls и
placeholders. Project facts принадлежат только generated context или явно
идентифицированному source-specific bridge. Правила leakage запрещают имена
source-project и представительные source-only paths в portable core, с узкими
reviewed exceptions.

Скопированные дистрибутивы должны заменить или перегенерировать context и не
должны копировать credentials, private runbooks, production endpoints, raw
transcripts или local eval records.

## Статус evaluation

Все bundled metrics и result records начинаются как `NOT_RUN` с null values.
Шаблоны, критерии оценки и успешное demo — не measured effectiveness evidence.

Paired pilot должен измерять first-pass acceptance, human review minutes,
defects, hallucinated paths и commands, scope escapes, traceability, assertion
sensitivity, flaky risk, false positives, tokens, tool calls и latency, где
доступно.

## Release gate

Продвижение из pre-release требует:

1. проходящих validator tests;
2. validator exit `0` на точном дистрибутиве;
3. approved source, security, framework и governance review;
4. paired pilot evidence, удовлетворяющего predeclared gates;
5. задокументированных limitations, exceptions и rollback;
6. accountable release approval.

Пока все шесть пунктов не подтверждены evidence, заявления о completion или
production-readiness не авторизованы.
