# Executive overview

> Полная русская версия для руководства также: [EXECUTIVE-OVERVIEW-RU.md](EXECUTIVE-OVERVIEW-RU.md).

## Decision summary

Этот пакет — **AI quality layer для инженеров автоматизации**. Он добавляет
переиспользуемые инструкции, структурную валидацию, paired evaluations, source
provenance и governance вокруг существующей работы по test automation. Это
**не framework scaffold**: он не выбирает test runner, не генерирует архитектуру
репозитория и не заменяет владельцев фреймворка команды.

Release status: **pre-release / eval NOT_RUN**. Structural validator зелёный
для source и packaged distributions (0 errors, 0 warnings). Combined tests:
55 PASS / 1 SKIP (создание symlink недоступно). Portable package из 162 files
провалидирован; generated project profiles и local smoke evidence исключены.
Никакое заявление об effectiveness или ROI не одобрено, пока paired pilot не
измерит оба arms.

Model routing настроен, но runtime provenance сейчас **PARTIAL**: live catalog
snapshot, resolved model ID, fallback и tool trace должны сохраняться для
leadership demo или measured run. Одна конфигурация не является evidence, что
routed model действительно выполнился.

## Решаемая проблема

Команды автоматизации, использующие coding agents, сталкиваются с повторяющимися
операционными рисками:

- правдоподобные, но несуществующие paths и commands;
- тесты, которые проходят, не доказывая принятое требование;
- review findings без evidence или с избыточными false positives;
- flaky-фиксы, скрывающие timing и isolation defects;
- project-specific инструкции, протекающие в portable material;
- недоверенный контент репозитория, меняющий поведение агента;
- неясное ownership для prompts, model changes, approvals и evidence.

Слой делает эти риски видимыми и измеримыми. Он не гарантирует correctness и не
снимает human accountability.

## Deliverables

- Структурный валидатор на Python stdlib с PowerShell-friendly запуском.
- Configurable registry для route names, secret-like patterns, portability
  deny/allow rules, profile provenance и required manifests.
- Baseline-versus-assisted eval suite: generation, review, debugging/flaky
  tests, context grounding, scope control и prompt injection.
- Primary-source registry, привязанный к конкретным controls.
- Guidance по governance, demo и ROI measurement.
- Pre-release version и ownership manifest.

## Intended users

- Test-automation engineers используют assisted templates и фиксируют execution
  evidence.
- Automation quality owners поддерживают fixtures, oracles и validator policy.
- Framework owners утверждают canonical paths, commands и patterns.
- Security reviewers утверждают secret, prompt-injection, MCP, network и
  external-action controls.
- Engineering managers утверждают pilot scope и adoption на основе measured
  outcomes.

## Что меняется операционно

Перед assistance команда замораживает task fixture, repository revision,
commands, scope и acceptance oracle. Затем та же задача выполняется как
baseline и с quality layer. Blinded human ревьюит оба outputs. Решения об
adoption опираются на quality, risk, effort и cost вместе; меньшая latency или
token use не компенсируют рост defects или scope escapes.

## Pilot proposal

1. Выбери 12–24 representative tasks по шести eval categories, используя
   synthetic или approved non-sensitive data.
2. Заморозь model, settings, tool permissions, repository revision и fixture
   для каждой пары.
3. Counterbalance arm order и изолируй conversations.
4. Blind reviewers относительно arm и фиксируй active review minutes.
5. Выполняй generated tests только в approved sandboxes.
6. Отчитывай raw paired records плюс aggregate uncertainty и missing data.
7. Проведи совместный automation, security и management review.

Proposed entry criteria:

- structural validator tests проходят;
- intended toolkit distribution без unresolved ERROR findings;
- fixture oracles имеют independent owner approval;
- data classification и retention одобрены;
- sandbox и tool permissions задокументированы.

Proposed adoption criteria должны быть заданы до pilot. Как минимум: zero scope
escapes и sensitive-data incidents, no increase in critical/high defects и
measured improvement хотя бы по одному target outcome без material regression
по остальным.

## Approval boundaries

AI layer может готовить черновики тестов, reviews, debug hypotheses и
документации внутри явного allowance. Human остаётся обязательным для:

- изменений production-code и merge decisions;
- решений по security, authentication, authorization, secrets и privacy;
- CI/CD, deployment, infrastructure, destructive commands и external writes;
- изменений архитектуры тестового фреймворка или canonical commands;
- изменений eval oracles, acceptance thresholds или reported metrics;
- использования confidential data или новых model/tool providers.

Полная модель ответственности и escalation — в [Governance](GOVERNANCE.md).

## Management decision points

После pilot руководство может выбрать:

- `ADOPT`: evidence удовлетворяет всем predeclared safety и quality gates;
- `LIMITED_ADOPT`: использование только в categories, прошедших gates, с явными
  exclusions;
- `ITERATE`: устранить measured regressions и перезапустить affected pairs;
- `STOP`: risk, cost или quality не оправдывают продолжение rollout.

Ни один вариант не должен опираться на unmeasured productivity multiplier.
Контракт измерения — в [ROI metrics](ROI-METRICS.md).
