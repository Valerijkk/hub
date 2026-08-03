# Руководство по демо

## Objective

Показать, что пакет — AI quality layer для существующей работы по автоматизации,
а не генератор тестового фреймворка, и как structural findings и paired eval
evidence становятся reviewable management inputs.

Длительность: 20–30 минут. Не используй production secrets, customer data или
unrestricted network access.

## Preconditions

- Python 3 доступен через `py -3` или `python`.
- Репозиторий на записанной revision.
- У презентера есть approved synthetic fixture.
- Baseline и assisted model settings и tool permissions идентичны.
- Generated commands будут запускаться только в disposable sandbox.
- Аудитория понимает, что eval results остаются `NOT_RUN`, пока это демо не
  записано как controlled run.

## Part 1: structural gate

Из корня репозитория в PowerShell:

```powershell
py -3 -B .\pro-qa-ai\scripts\validate_toolkit.py `
  --root .\pro-qa-ai
```

Объясни:

- exit `0` означает отсутствие configured ERROR findings;
- exit `1` означает наличие policy findings и blocked release eligibility;
- exit `2` означает, что validator или manifest не удалось загрузить;
- secret-like findings показывают только rule ID, path и line, никогда matched
  value;
- решения allow и deny живут в `scripts/validation_manifest.json` и требуют
  review.

Для machine-readable output:

```powershell
py -3 -B .\pro-qa-ai\scripts\validate_toolkit.py `
  --root .\pro-qa-ai `
  --format json `
  --fail-on never
```

`--fail-on never` — только для inspection. Его нельзя использовать как release
gate.

## Part 2: focused unit tests

```powershell
py -3 -B -m unittest discover `
  -s .\pro-qa-ai\scripts\tests `
  -p "test_*.py" `
  -v
```

Покажи, что тесты покрывают broken links, malformed frontmatter, skill naming,
prompt-ID gaps, generated-profile provenance, forbidden reasoning disclosure,
route typos, portability leakage, secret redaction и manifest provenance.

Проходящие validator unit tests доказывают поведение валидатора; они не
доказывают чистоту содержимого тулкита и не доказывают, что AI assistance
улучшает outcomes.

## Part 3: paired task

1. Выбери один fixture из eval suite, предпочтительно
   generation или review.
2. Заполни все `TODO` и slot fields; oracle должен быть approved до любого run.
3. Запусти baseline в чистой conversation.
4. Запусти assisted arm в другой чистой conversation с versioned toolkit.
5. Сохрани visible outputs и non-sensitive tool evidence.
6. Отдай оба outputs reviewer под нейтральными метками `A` и `B`.
7. Скопируй run record template по разу на
   каждый arm и заполняй только measured fields.

Suggested audience view:

- fixture и frozen scope;
- два anonymized outputs;
- focused command evidence;
- решение reviewer и corrections;
- metric record, где unavailable values всё ещё помечены `TODO` или `NOT_RUN`.

## Part 4: security behavior

Используй synthetic prompt-injection fixture.
Покажи, что:

- instruction-like текст репозитория сообщается как untrusted data;
- нет изменений permission, network или output destination;
- safe in-scope analysis продолжается;
- privileged instruction, credential или private reasoning не воспроизводятся.

Не демонстрируй на реальном credential. Secret redaction следует доказывать
unit test и synthetic pattern, approved для демо.

## Part 5: management close

Представляй только:

- structural status и unresolved finding counts;
- paired raw observations;
- first-pass acceptance и valid defect results;
- active review minutes;
- scope и security events;
- token, tool-call и latency data, когда платформа их отдаёт;
- limitations и missing observations.

Не представляй extrapolated annual saving из одного демо. Используй
[ROI metrics](ROI-METRICS.md) после завершения approved pilot sample.

## Stop conditions

Останови действие демо, сохрани evidence и уведоми owner, если:

- появляется реальный secret или personal datum;
- tool пытается unauthorized write или network call;
- baseline и assisted conditions материально различаются;
- oracle меняется после просмотра output;
- команда затронула бы non-disposable environment;
- reviewer не может оставаться blinded.

Независимые safe sections можно продолжать только когда incident owner
подтверждает, что evidence и permissions остаются isolated.
