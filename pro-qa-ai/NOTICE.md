# Notices and provenance

## Ownership

Portable Test Automation AI Quality Layer распространяется по
[`LICENSE`](LICENSE). Packaging не даёт прав на external redistribution.

## GSD inspiration

Концепции lifecycle, atomic planning, wave execution, goal-backward verification
и handoff сформулированы заново по мотивам идей из GSD lineage. Проверенные
репозитории нельзя считать точным источником локального текста без revision
evidence:

- legacy: https://github.com/gsd-build/get-shit-done — его текущий README
  перенаправляет разработку в https://github.com/open-gsd/get-shit-done-redux;
- related current source reviewed separately: https://github.com/open-gsd/gsd-core;
- обе проверенные линии заявляют MIT, но точный repository/revision локального
  original inventory остаётся `UNKNOWN`;
- legacy notice, если его substantial portion копируется: Copyright (c) 2025
  Lex Christopherson.

Не заявляется, что текущий тулкит — official GSD distribution либо производная
конкретного current repository. Portable workflows переписаны и не представлены
как verbatim upstream files. Поэтому `UNKNOWN` не блокирует adapted ideas с этой
attribution. Он **блокирует packaging**, если
`contains_substantial_copied_gsd_material=true`: до external redistribution
нужны точный immutable commit SHA, проверенная license и все required notices.

## Project-derived material

Generated profiles и вручную поставленные review examples по умолчанию
исключаются из portable package. Project-derived example может быть
redistributed только когда его card фиксирует:

- verified provenance;
- completed redaction;
- known source license/permission;
- `reuse.scope: portable`;
- explicit human approval.

Материал со статусом `UNKNOWN`, `project-only` или unapproved не portable.

## Primary-source registry

Внешние standards/vendor documentation, использованные для вывода controls,
индексируются в [`sources/PRIMARY-SOURCES.md`](sources/PRIMARY-SOURCES.md).
Ссылки и access dates — evidence исследования, а не лицензии на копирование
source prose.
