---
manifest_version: 1
product: pro-qa-ai
toolkit_version: 0.4.0-beta.3
status: foundation-ready
owner: automation-quality
layer: test-automation-ai-quality
source_registry: sources/PRIMARY-SOURCES.md
validation_config: scripts/validation_manifest.json
validation_status: STRUCTURAL_PASS
last_validation: 2026-08-08
provenance: official primary sources and repository inspection through 2026-08-08

---

# Манифест — pro-qa-ai

## Идентичность

Этот дистрибутив — **pro-qa-ai**: AI quality layer для инженеров
автоматизации тестирования. Он даёт переносимые операционные инструкции, структурную валидацию,
провенанс и правила владения вокруг уже существующей среды автоматизации. Продуктовая карта: [TOUCHSTONE.md](TOUCHSTONE.md).

Это не scaffold тестового фреймворка. Он не выбирает runners, не создаёт
архитектуру фреймворка и не владеет проектными командами и путями.

Статус — **foundation-ready**. Это **не** значит eval-validated или GA.
Манифест намеренно не хранит число тестов/файлов или старый
artifact digest: после любого изменения такие значения становятся ложными.
Фактические exit codes, counts и SHA-256 записываются для каждого candidate по
[clean-room release protocol](docs/PORTABLE-RELEASE.md) и
`PACKAGE-INVENTORY.json`. `STRUCTURAL_PASS` относится только к точному
candidate, принятому указанной датой; paired effectiveness остаётся отдельным
непройденным gate.

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
- [Контракты записей](standards/contracts/README.md): task, handoff,
  verification, scoped capabilities, provenance и workflow evaluation; схемы
  декларативны и не заменяют controller enforcement
- [Поиск конфиденциальных значений](scripts/scan_secrets.py) — рабочее
  дерево и вся история; значения не печатает, историю не правит
- [Mechanical context inventory](scripts/context_inventory.py), его
  [config](scripts/context_inventory_config.json) и
  [schema](scripts/context_inventory.schema.json)
- [Конфигурация валидации](scripts/validation_manifest.json)
- [Unit-тесты валидатора](scripts/tests/test_validate_toolkit.py)
- [Primary sources](sources/PRIMARY-SOURCES.md)
- [Clean-room portable release](docs/PORTABLE-RELEASE.md)
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
- поля provenance тулкита и generated-context;
- единую форму файлов ролей и скилов: состав секций, их порядок, наличие
  чеклиста готовности;
- бюджет размера файлов инструкций по ярусам 38 / 54 / 90 КБ, включая
  требование записанной причины у каждого исключения;
- упразднённые понятия в тексте: удалённые издания, подсистемы и скрипты,
  на которые документ продолжает ссылаться;
- разнобой машинных значений: одно значение, записанное двумя способами;
- ссылки на роли, которых нет в ростере;
- пустые и разорванные вставки кода;
- повтор текста между файлами при пересечении от половины;
- переход между ролями, описанный только с одной стороны;
- упоминание выхода наружу без рядом стоящего разрешения или запрета.
- registry и встроенные examples шести JSON Schemas контрактных records через
  ограниченный Python-stdlib subset (`contracts`).

Последние семь закрывают класс, который структурные проверки не видят:
файл валиден и связен, но обещает то, чего в пакете нет, либо пересказывает
соседний файл. Все они найдены построчным чтением 2026-08-04 и после этого
автоматизированы.

Проверка повтора — не про стиль, а про цену. Скопированный абзац
расходует контекст модели дважды и расходится с оригиналом при первой
правке: чинят одну копию, вторая продолжает учить старому. Законные пары
(правило `CHECK-NNN` и тред, из которого оно выросло) внесены в список
исключений явно.

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

Продвижение из pre-release выполняется по
[14-шаговому clean-room protocol](docs/PORTABLE-RELEASE.md) и требует:

1. проходящих validator tests;
2. validator exit `0` на точном дистрибутиве;
3. approved source, legal/provenance, security, framework и governance review;
4. paired pilot evidence, удовлетворяющего predeclared gates;
5. задокументированных limitations, exceptions и rollback;
6. accountable release approval.

Пока все шесть пунктов не подтверждены evidence, заявления о completion или
production-readiness не авторизованы.
