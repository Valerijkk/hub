# Clean-room portable release protocol

Этот протокол выпускает только переносимое ядро. Он не копирует generated
project context, локальные evidence/run records или структуру тестов исходного
проекта. Число файлов и статусы проверок берутся из текущего запуска и inventory;
их нельзя переписывать в документацию как вечные константы.

## 14 обязательных шагов

1. **Зафиксировать candidate.** Записать release identifier, точный source root,
   commit/dirty-state и ответственного. Непроверенные параллельные изменения не
   включать молча.

2. **Создать clean room.** Использовать новый task-owned workspace/container с
   объявленными CPU/RAM limits, без сети по умолчанию и с read-only source mount.
   Output должен находиться вне source root на локальной доверенной файловой
   системе.

3. **Проверить границы.** Подтвердить, что source и output не пересекаются,
   output отсутствует либо пуст, а parent не является symlink, junction, reparse
   point или network path.

4. **Заморозить package policy.** Review `scripts/package_manifest.json`, включая
   include/exclude/prune, required paths, лимиты и provenance declaration. Не
   подменять policy raw-copy командой.

5. **Закрыть legal/provenance gate.** Сверить `NOTICE.md` и
   `sources/PRIMARY-SOURCES.md` с первичными upstream links. Если package
   содержит substantial copied GSD code/prose, `gsd_exact_revision=UNKNOWN`
   является stop-condition: нужен immutable commit SHA и требуемые notices.
   Заново сформулированные adapted ideas разрешены при явной attribution и
   `contains_substantial_copied_gsd_material=false`.

6. **Проверить secrets.** Запустить `scripts/scan_secrets.py` для рабочего дерева
   и, в среде с доверенным Git, истории. Любая находка или безопасный отказ
   останавливает выпуск; значения в evidence не копировать.

7. **Проверить source distribution.** Запустить полный
   `scripts/validate_toolkit.py --checks all` на точном candidate. Сохранить
   command, exit code и timestamp как release evidence, не как строку в
   `MANIFEST.md`.

8. **Запустить script tests.** Выполнить discovery `scripts/tests/test_*.py` в
   clean room. Maintainability heuristics — размер функции, SOLID, DRY, mocks и
   Page Objects — остаются review guidance; сами по себе они не release gate.

9. **Сделать dry-run.** Запустить `package_portable.py --dry-run`; записать
   вычисленные в этом запуске file count и manifest SHA-256 в evidence artifact.
   Не сравнивать их со старым hard-coded числом.

10. **Собрать новый output.** Запустить packager без `--dry-run`. Он пишет только
    в уникальный sibling staging, полностью проверяет staged tree и публикует
    его одним atomic rename. Ошибка до rename не должна менять destination.

11. **Проверить состав inventory.** Убедиться, что `PACKAGE-INVENTORY.json`
    содержит только policy-selected paths. Generated `context/projects/<profile>/`
    (включая старые локальные профили любого имени), local evidence и private
    review material не должны присутствовать.

12. **Проверить exact output.** Запустить полный validator уже из опубликованного
    package root. Проверяется именно artifact, а не source directory.

13. **Проверить artifact tests и digest.** Запустить script tests из package,
    повторно вычислить/сопоставить inventory fingerprints и сохранить фактические
    результаты. Любое расхождение требует новой сборки с шага 1.

14. **Одобрить и опубликовать.** Source/security/framework/governance reviewers
    фиксируют limitations, exceptions, rollback и accountable approval. Выпуск
    идентифицируется immutable manifest digest; staging и clean-room resources
    удаляются. Без этого evidence candidate не называется production-ready.

## Stop conditions

- неизвестная точная ревизия при substantial copied upstream material;
- missing attribution/license notice;
- secret finding, unsafe filesystem object или изменение source во время сборки;
- validator/test failure либо mismatch staged/published inventory;
- generated project profile или local/private evidence в portable artifact;
- destination стал непустым или изменил identity до atomic publish.
