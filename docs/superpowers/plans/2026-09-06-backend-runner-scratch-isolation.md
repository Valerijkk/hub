# Backend runner: независимый RAM scratch

## Резервный checkpoint по запросу пользователя, 2026-09-06

Пользователь запросил сохранение всех текущих изменений в своей ветке.
Этот snapshot сохраняет незавершённый ремонт QA и не является принятым
исправлением, зелёной регрессией, разрешением на merge или продуктовой приёмкой.
База до snapshot: `b32b4626631272bee3d1d425cb3faf68a9b7367f`.
Единственная разрешённая публикация: `calls-feature-codex` в
`origin https://github.com/Valerijkk/hub.git`, обычный fast-forward push.
`master`, чужие ветки, Mail/Calendar и `pro-qa-ai/**` не изменяются.

Сохраняются 26 явных изменённых/новых путей: QA runner/observer, их selftests
и fixtures, LF-policy и пять документов состояния/плана. Сырые Docker logs и
промежуточные отчёты в ignored `.superpowers/sdd/2026-09-06-backend-runner-scratch/`
остаются локальными; в Git не добавляются принудительно. Для продолжения без
них ниже сохранены точная команда, последний результат, image/resources,
непроверенные изменения и очередь. Git SHA публикации проверяется после push,
а не объявляется известным заранее.

После описанного ниже Composed9 внесены следующие изменения, для которых
новый Docker-прогон ещё **NOT_RUN**:

- Execute forwarding: добавлен узкий `_hard2g_execute_forwarding_source_is_safe`
  для трёх точных передач `"$@"` в двух текущих helper-функциях. Общий shell
  parser и callers не изменены; 30 focused cases до реализации были RED.
- Runtime lifecycle: oracle теперь проверяет полные текущие terminal case
  blocks monitor/reaper. Добавлены шесть отрицательных reset-order cases для
  `go_child_session`, `go_child_starttime`, `go_child_pgid`; сам reset-order
  predicate ещё не дополнен. Эти шесть случаев ожидают подтверждения RED.
- Observe mutations: исправлены реальные source-литералы и добавлен индекс
  мутанта в assertion. Новая диагностика ожидает запуска; известный false
  positive не скрывается переносом мутации в другую ветку.

### Точка продолжения

1. Сделать свежий read-only resource preflight по ограничениям ниже. Только
   один existing-image `docker compose -f docker-compose.local.yml run --rm python-contract`;
   без selectors, pull/build, host validators, disk scratch или изменения
   лимитов. Проверить фактические image/cgroup/tmpfs и сохранить exit/output.
2. По результатам подтвердить forwarding и current lifecycle, локализовать
   observe mutation index. После подтверждённого RED отдельно дополнить
   упорядоченные reset checks, не менять runtime ради прохождения oracle.
3. Только после focused GREEN интегрировать execute-specific admission в пять
   execute-контекстов: stage, runner aggregate, abort, C02 ownership, C02
   durable topology. Runtime/plan/receipt остаются на общем parser.
4. Отдельные оставшиеся причины: actual publication использует явные
   before/after snapshots вместо вложенных AST calls; execute oracle считает
   неполные строки как элементы tuple; budget fixture содержит lifecycle
   stubs; receipt miniature отстала от актуального helper-протокола. Каждую
   исправлять отдельным контрактом/негативным тестом, без снятия assertions.
5. Observe требует отдельной проверки early terminal guard и COMPLETE binding:
   текущие ABORTED-only assertions не доказывают обе эти ветки.

Последний выполненный framework gate остаётся **FAIL (633 PASS / 41 FAIL)**;
это результат Composed9, не результат нового snapshot. Go gate, Messenger
SUT/browser/API, Jitsi deployment и реальное media остаются NOT_RUN/неприняты.
Чек-лист модулей по-прежнему 12/27 UNIT_ACCEPTED; LiveKit активен.
Следующая фича или более широкий backend memory profile этим checkpoint
не разрешаются. Для резервного сохранения новый тяжёлый QA не запускался.

## Последняя проверка, 2026-09-06

Composed9:674tests41FAIL/633PASS39.29s,seed3443901789. Engine
`8fea4d7fa91ce5c6a0ac3644c85d52672d423412694d25eb0f8bef8068d9be92`,
start `2026-09-06T16:19:01.95963593Z`; прежний image и фактические2GiB/no-swap,
2CPU,readonly,/tmp64MiB. Preflight:10198776KiBfree,только idleBuildKit,
allbuildsCompleted,disk0/queue0; во время9988572KiBfree. Exit1.
Input root-compose hash `88734B2A0766BBD6C1DC6DD5A0EC2585C122FAB839DDC54E0502ED7C115F0C6F`.
Проходят child handshake/no-sleep gate, stream mutation и8receiptguard cases.
32 новых intentional RED:30missingforwardinghelper,2currentlifecyclecanonical;
16 lifecycledenials PASS. Девять прежних failures остаются, в том числе
observe mutant false-positive и actualpublication closure. Raw output:
`forward-lifecycle-red-output.txt`. FullframeworkFAIL; backend/Jitsi не запускались.

Composed8: 626 tests,14 FAIL/612 PASS,28.00s,seed1773387413. Engine
`1bac36996398b3e9fb27375b851afa0ad3d8247b1a8d65502fb1b8fe0e7a9fae`,
start `2026-09-06T15:50:37.535785887Z`. Существующий image
`sha256:c942b71372fa6094a1370b175386af25def385a6678bbf0ab04cdd5b50364e16`;
inspect подтвердил2GiB/no-swap,2CPU,readonly,/tmp64MiB. Preflight:
6303052KiB free, только idle BuildKit, все сборки Completed,disk0/queue0;
во время прогона6150356KiB free. Это samples, не peak/final OOM telemetry.

Composed6:616/18F/598P; Composed7:618/18F/600P. После независимого review
в8 проходят оба bounded provenance walker, printf escape, terminal AST owners,
lexical/control cases. Receipt guard отдельно подтвердил canonical false-negative
и небезопасный bare-line false-positive; шесть других отрицательных случаев PASS.
Исправление guard и event-driven child diagnostic пока post-source NOT_RUN.
Raw outputs сохранены в task workspace; fullframeworkFAIL, без исключения тестов,
повышения ресурсного профиля, новой Go-попытки или Jitsi activation.

## Область изменения

После двух ограниченных Go-диагностик пользователь разрешил исправить
QA-запуск и продолжать работу. База: `b32b4626631272bee3d1d425cb3faf68a9b7367f`,
ветка `calls-feature-codex`. Это исправление тестовой инфраструктуры, не
приёмка Messenger, Calls, Jitsi или всего backend.

`stage_scratch` переносится из `GOTMPDIR` в фиксированный
`/tmp/.backend-unit-stage-$stage.XXXXXXXX`. Go WORK, package plans и
исполняемые файлы остаются в `/var/tmp/hub-go`. Четыре замены префиксов
согласуют создание scratch, монитор stop-marker и проверку завершённого
stage. Проверки UID/GID, прав, identity, отсутствия путей и cleanup сохранены.

Используется уже существующий `/tmp`: RAM tmpfs 64 MiB, noexec/nosuid/nodev.
Максимальная заданная сумма двух captures — 25 MiB + 4 MiB + два sentinel
байта; пакеты выполняются последовательно. Новых mount, профилей памяти,
лимитов output, schemas, observer/lease API и SSD scratch нет.

Это разрывает прямую конкуренцию Go WORK и диагностики за одну файловую
систему. Достаточная ёмкость Go scratch и переживающая любой abort диагностика
этим не доказаны. Вопрос постоянного backend-профиля 4 GiB/no-swap остаётся
отдельным; текущие Compose/observer ограничения не изменены.

## Проверка и честные границы

Единственная команда: `docker compose -f docker-compose.local.yml run --rm python-contract`.
Образ повторно использован без сборки/скачивания:
`sha256:c942b71372fa6094a1370b175386af25def385a6678bbf0ab04cdd5b50364e16`.

| Прогон | Результат | Что он доказывает |
|---|---|---|
| RED-1 | 604 tests, 80 FAIL / 524 PASS, 32.39s | Первичная fixture имела ранний отказ; не окончательный oracle |
| Временная shell trace | 81 FAIL / 523 PASS, 31.72s | Локализован отказ synthetic identity; диагностический, не приёмочный прогон |
| RED-2 | 81 FAIL / 523 PASS, 32.22s | Все 7 scratch-сценариев падают на старом source; legacy path действительно проходит старый guard |
| После source + reviewed fixtures | 38 FAIL / 566 PASS, 28.44s | Все scratch-сценарии проходят; 43 прежних падения устранены, новых failing IDs нет; полный gate FAIL |
| Identity/terminal RED + fixture bundle | 607 tests, 33 FAIL / 574 PASS, 27.13s | Числовой parser и normal/recovery nested alias подтверждены до source-правок |
| Identity/terminal GREEN + shell-extractor RED | 610 tests, 24 FAIL / 586 PASS, 26.00s | Все observer и scratch/identity cases проходят; 21 прежний static failure + 3 новых intended RED; полный gate FAIL |

RED-2: seed `3744348604`, engine
`cf9fb8a86cb601c26e05299abbf63bef4d18771202ef55ec6839cbdd45201964`,
start `2026-09-06T13:29:51.88550756Z`. Инспекция подтвердила 2 GiB RAM,
2 GiB memory+swap (без swap), 2 CPU, readonly и `/tmp` 64 MiB. Это параметры
Python-проверки, не backend-профиля. Перед запуском соседние QA/build завершились;
их ресурсы root не останавливал. Engine-terminal events RED-2 не зафиксированы.

Независимый source/oracle review: Terra/xhigh (requested; actual UNKNOWN),
SPEC PASS / QUALITY PASS, C/I/M 0/0/0. Runtime acceptance отдельно.
Точные локальные материалы сохранены в
`.superpowers/sdd/2026-09-06-backend-runner-scratch/`.

Composed run: seed `275404350`, engine
`3bee3a41ce21c0ccf85939a9e8b494e9ffd7576c2b968339c7c2850bf774ab96`,
start `2026-09-06T13:51:20.848298583Z`; Compose exit1. Фактические параметры
те же: 2 GiB/no-swap, 2 CPU, readonly, `/tmp`64MiB. После прогона хеши
execute/runtime/Compose/new scratch oracle неизменны. Engine terminal event
и peak memory не зафиксированы; sample OOM=false не заменяет эти данные.
Точные stdout/failure inventory — `composed-run-1-output.txt` в SDD workspace.

Прошли все3 новых scratch tests,2 monitor,6 reaped-child и5 реальных POSIX
signal tests, receipt fixture-chain/terminal-negative cases и мигрированные
adapter API cases. Повторный recovery test прошёл исправленное сравнение
dict/list, но выявил вложенное aliasing из-за shallow copy в production:
эта ошибка требует отдельного source-атома, а не снятия assertion.

## Отдельные находки

Четвёртый составной прогон:612 tests,22 FAIL/590 PASS,28.65s,seed1712759000;
engine `6ce923951094cc8a3fd551a1c87324fe27638a942ffcbf388a15e49586f9096a`,
start `2026-09-06T14:34:51.375532372Z`. Тот же образ, проверенные2GiB/no-swap,
2CPU,readonly/tmp64MiB. Все observer cases остаются PASS; два исправленных
provenance-name cases проходят. Корректирующий shell I1 тест подтверждает
ложный допуск dangling logical opener; реальное многострочное тело проходит.
Named diagnostics локализовали отказ на первой строке документации класса
и regex, допускающем `${NAME}`, но не текущий `$NAME`. Это отдельные ошибки
test helpers, не основания менять production или исключать проверки.
Полный framework gate FAIL; raw output сохранён как `shell-i1-red-output.txt`.

Пятый составной прогон:615 tests,20 FAIL/595 PASS,28.86s,seed1651823279;
engine `8135e4ec9770d2deba3bd0376113063637f0a8c6bf4b17aad11c9676f0b7b05a`,
start `2026-09-06T15:11:12.369773651Z`. Фактически проверены прежние image,
2GiB/no-swap,2CPU,readonly/tmp64MiB. Все observer cases, newline repairs и
две miniature-source fixtures проходят; четыре helper RED подтверждены.
AST target depth65 ещё ошибочно допускается; depth1/64 обоих путей проходят.
Binding depth65 будет отдельным parameterized RED до изменения обоих walkers.
Raw output: `helper-batch-red-output.txt`; полный framework остаётся FAIL.

Операционный порядок уточнён root после четырёх подряд соседних Mail QA-прогонов:
только короткий existing-image Python selftest допускается рядом с QA при
нулевой очереди диска, отсутствии активной сборки и запасе минимум3GiB после
вычета неиспользованного cgroup-бюджета соседа и собственного полного2GiB cap.
Перед запуском: free12192704KiB, сосед1.885/8GiB, disk0%/queue0; во время:
free10707516KiB. Это замеры, не peak/continuous telemetry. При наблюдаемом
freeRAM<3GiB прекращается только свой контейнер. Тяжёлые QA/build остаются
последовательными; no disk scratch, no pull/build, профиль backend не изменён.

Третий составной прогон: seed `2496297863`, engine
`841ac5799e48ff8102fe197ffb4af3b71bd1a665b34a860e3bd06066e560925e`,
start `2026-09-06T14:16:54.374922522Z`; тот же образ и фактически проверенные
2 GiB/no-swap,2 CPU,readonly,64MiB tmpfs. Два новых source-атома прошли
независимый C/I/M0/0/0 review и свои реальные Docker регрессии: numeric grammar
без integer conversion и глубокая копия двух terminal captures. В обоих
normal/recovery путях сохранены session/stop/close/identity проверки.
Исправления fake pipe, root-handoff seam, canonical manifest drift и момента
file/directory race также прошли scoped review и Docker cases.
Оставшийся shell parser в test oracle ошибочно обрывает тело на скобках внутри
quoted awk и вложенных `|| {}`; его source пока не принят, зафиксирован RED.
Ни один тест не исключён/xfail; observer policy/Compose/ресурсные лимиты сохранены.

- 74 остальных framework failures существовали до четырёх source-замен;
  они не считаются допустимыми или исправленными автоматически.
- Observer receipt fixtures содержат несовместимые hostname/container ID.
  Их цепочка provenance исправлена отдельно без ослабления validator;
  соответствующие Docker selftests прошли, raw-byte LF policy закреплена.
- Shell numeric pattern в `_execute_identity_is_exact` исправлен отдельным
  атомом после sensitive RED. Docker прошёл 65 canonical,95 malformed и11
  arity/shape/ownership комбинаций через настоящий `/bin/sh`; type/equality
  ограничения сохранены. Это не проверка полной backend-компиляции.
- Go build/vet/unit/race, реальный Messenger SUT и Jitsi media остаются
  непринятыми. Старые успешные unit-прогоны не заменяют эти проверки.
