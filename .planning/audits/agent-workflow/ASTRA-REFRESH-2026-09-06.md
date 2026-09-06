# Обновление рабочего процесса — 2026-09-06

## Scope и результат импорта

Поручение: забрать правила и настройки из master через MR/PR и сверить с
официальным latest-model guide. Это настройка разработки, не миграция API Hub.
Рабочая ветка: `calls-feature-codex`; origin: `https://github.com/Valerijkk/hub.git`.

Начальный HEAD: `ac918830e2aa0d7082297d9913be0fd750013fb5`.
Refreshed master: `c08b80366ee62447059d66e47e97b5f65727ee91`.
Расхождение: feature впереди на 323 коммита, master — ровно на один.
[PR #38](https://github.com/Valerijkk/hub/pull/38): `master → calls-feature-codex`,
MERGED 2026-09-06T10:11:26Z; merge `c84fd980e35cb101da28815335233cd40c3fd8ba`.
43 файла перенесены без изменения upstream-содержимого. Локальный fast-forward
сохранил все семь WIP-файлов побайтно. Stash/reset, новые worktrees, удаления,
изменения master, global config, Mail/Calendar и protected zone не выполнялись.

## Дополнение по гайду

Три active playbook: `analysis.md`, `model-routing.md`, `prompt-contract.md`.
Убраны восемь старых routing-строк `inherit`; selector GSD отделён от explicit
tool dispatch. Уточнены автономность, приоритет инструкций, читаемость,
независимое делегирование, ограничение дисковых операций и глубина проверок.
Полные feature-шаблоны не требуются для bounded docs/config без product behavior.
Исторические ADR и принятые продуктовые контракты не переписаны.

Официальные источники прочитаны 2026-09-06:

- [Astra guidance](https://developers.openai.com/api/docs/guides/latest-model):
  prompting best practices и migration quickstart.
- [Astra model](https://developers.openai.com/api/docs/models/gpt-6-astra):
  модель и поддерживаемые reasoning levels.
- [Codex config](https://learn.chatgpt.com/docs/config-file/config-reference):
  explicit spawn overrides и concurrency, исключающая root.
- [Config basics](https://learn.chatgpt.com/docs/config-file/config-basic):
  trusted project и precedence конфигураций.

API-совместимость Astra проверялась как документационный контекст. В репозитории
этой задачей не менялись API SDK, endpoints, tools, credentials, sampling или
кэширование. Новые API-возможности не становятся TOML-настройками автоматически.

## Evidence и пределы

- Upstream scoped diff: 43 docs/config paths; `git diff --check` PASS.
- Independent import review: `/root/astra_rules_review`, requested
  `gpt-5.6-terra` / `high`; actual model UNKNOWN без provider telemetry.
  Critical 0. Замечания: loader evidence отсутствует; GSD selector неоднозначен;
  восемь legacy inherit-строк. Последние два устранены в дополнении.
- Independent review дополнения по пяти owned docs paths: C0/I0/M0,
  `READY_FOR_DOCS_COMMIT`; тот же explicit Terra/High, actual model UNKNOWN.
  Staged `git diff --cached --check` PASS; продуктовые WIP-файлы вне index.
- Product tests/build/lint: NOT_APPLICABLE этому docs/config slice; не запускались.
- Config-loader/runtime activation: NOT_RUN. Installed IDE package metadata:
  `openai.chatgpt` 26.901.22334. User config уже содержит Astra/Max/default,
  Terra/Medium, три spawned slots и max_depth=1; global config не изменялся.
- `.codex/config.toml` оставлен точной копией master. Текущая Config Reference
  не перечисляет `max` в таблице config effort и не описывает `agents.max_depth`.
  Astra model/API допускают max. Это неполное evidence совместимости клиента,
  не доказательство invalid config. Фактическая загрузка требует отдельного
  наблюдения в новой сессии; конфигурация не доказывает переключение текущего runtime.
- Top-level model/effort headers в локальных `gsd-*.toml` не найдены; install/update
  GSD или plugins не выполнялись. Их defaults не заменяют явный runtime contract.
- Публикация дополнения подтверждается Git/remote проверкой после commit;
  эта запись сама по себе не является доказательством push или runtime PASS.

## Восстановление коммуникаций

Последний GREEN product commit — `ac918830e2aa0d7082297d9913be0fd750013fb5`
(pin). ForwardDialog: RED complete, production не написан; предыдущий writer
получил usage-limit error. Исходники ForwardDialog/EN/RU/Compose не менялись.
36-case oracle, actor-only fixture и планы сохранены локально, не опубликованы
как принятая реализация. Повторять RED или весь product matrix ради изменения
правил не нужно. Сначала rebind существующего контракта к новой revision и
явный выбор runtime по текущему routing, затем source → scoped review → Docker GREEN.

Receipts: `testframework/reports/frontend-unit-forward-red-20260906/`.
Recovery: `.superpowers/sdd/2026-09-06-message-forward-lifecycle/`.
Forward plan: `docs/superpowers/plans/2026-09-06-message-forward-lifecycle.md`.
Panel plan: `docs/superpowers/plans/2026-09-06-panel-unpin-lifecycle.md`, ждёт forward.
Jitsi не активирован; LiveKit остаётся текущим provider. Disk floor/Go OOM gates
этой задачей не снимались.

SHA256 WIP до и после exact upstream import (до намеренного обновления STATE):

| Путь | SHA256 |
|---|---|
| `.planning/STATE.md` | `4F7C7490B03108F99EBA80752563ECA1F8E1E8E2C817AD0C361700301683B221` |
| `docs/superpowers/plans/2026-09-05-communications-completion-queue.md` | `7D9557A63D7C929DE82AA2E28DAA1D3E29C726529EFF493C72E060A2FF697366` |
| `docs/superpowers/plans/2026-09-06-message-pin-lifecycle.md` | `D0C11C53EBD6FE5F952AA66DAB46DD61D163D9D0AC3D1983A70923EBAE639A16` |
| `frontend/tests/messenger/phase151-message-action-capabilities.test.tsx` | `657C4EEAA8907D3514153369FCC1ABE43BE6C79CC18D18C2E50191EEFAE3180A` |
| `docs/superpowers/plans/2026-09-06-message-forward-lifecycle.md` | `0B8F6CA5732D02F853CBD8C3DC46A1970C600AE409228A44419665E120DB9CDC` |
| `docs/superpowers/plans/2026-09-06-panel-unpin-lifecycle.md` | `765B99E88AB5B93118983EDF7FFA22E46DBCC38F345105429E766631173E022C` |
| `frontend/tests/messenger/message-forward-lifecycle.test.tsx` | `01207AD0B3B34707D83B7D5771015D451618C94EB1D599C8070829C0AF30E077` |
