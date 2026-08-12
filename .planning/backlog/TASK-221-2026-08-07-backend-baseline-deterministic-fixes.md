# TASK-221 — 🧪 Backend baseline deterministic fixes

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **На проверке** · Фаза: **128**

## [OUTCOME]

Одиннадцать детерминированных падений TASK-203 устранены на уровне тестовой
инфраструктуры без изменения production behavior: переносимые source/schema
guards и полностью локальный GitLab HTTP fixture.

## [ROLE]

- Owner: `be1` — минимальные test-only исправления.
- Verifier / handoff: `be-ver` — focused Docker rerun и проверка отсутствия
  внешнего I/O/ослабленных assertions.

## [SCOPE AND CONSTRAINTS]

- Allowed: десять файлов из allowlist раздела `Confirmed fix worker allowlist`
  в `.planning/debug/test-baseline-failures.md`.
- Do not touch: production repositories, NATS code/tests, CI/Compose, legacy
  paths и thresholds.
- Depends on: TASK-202 и красная попытка TASK-203 от 2026-08-07.

## [ACCEPTANCE]

- Восемь domain guards вычисляют backend module root независимо от host/Docker
  path; девять функциональных тестов проходят под `/app`.
- TASK167 сохраняет executable schema assertions без runtime-чтения
  `.planning/**`; ADR traceability не теряется.
- GitLab route test использует только injected no-network transport, сохраняет
  SSRF contract и исходный one-second sentinel.
- Focused packages green с Go 1.25, `-tags hubtest -race`; GitLab test green
  `-count=10`; NATS не менялся.

## [OUTPUT]

1. Два атомарных commits: portable guards; deterministic GitLab fixture.
2. Focused command/count/time evidence.
3. Handoff `be-ver` с branch/HEAD и dirty ownership.
