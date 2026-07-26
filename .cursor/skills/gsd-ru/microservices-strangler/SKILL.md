---
name: microservices-strangler
description: Planning-only чек-лист выноса домена по ADR-ARCH-MICROSERVICES (strangler fig). Используй до кода выноса — готовит PLAN.md для arch/tech-lead.
disable-model-invocation: true
---

# microservices-strangler

Planning-only чек-лист выноса домена по **ADR-ARCH-MICROSERVICES-2026** (strangler fig).
Исполнители плана: **`arch`** + **`tech-lead`**; верификация плана: **`arch-ver`** + **`security-auditor`**.
Оркестрация: **`team-lead`**.

> Этот skill **не** запускает код выноса. Он готовит PLAN.md. Реализация — отдельными фазами после Go/No-Go и (если код freeze) после снятия freeze.

## Когда

- Планируется вынос модуля в отдельный сервис
- Пилот / следующий юнит из горизонта ADR
- Ревью «готовы ли к split» (Phase 0 hardening)

**Не использовать** для предложения вынести **tracker** без нового ADR (явно исключён из горизонта).

## Предусловия ADR (обязательны)

Канон: `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md` (accepted 2026-07-02).

| Инвариант | Суть |
|-----------|------|
| Schema-per-service | Своя PG schema + DB-user; **запрет shared tables** |
| Grants | Cross-schema JOIN/FK запрещены **грантами**, не только ревью |
| Messaging | NATS JetStream + outbox; idempotent consumers (`event_id`) |
| Sync | Минимум; >1 sync-hop = граница нарезана неверно |
| Auth | Пилот: HS256 допустим; до hub-ai — JWKS (asymmetric) |
| Observability | OTel trace-context в NATS headers — gate до пилота |
| Edge | Caddy route-flip (не nginx); rollback обязателен |
| CI | Path-filtered jobs (бюджет ~3000 мин/мес) |
| Порядок Hub | **notification → hub-ai → realtime-trio** (ws, messenger, calls) |
| Исключения | **tracker** не выносить без нового ADR; tms — условно после Go пилота |

## Горизонт выноса (Hub)

| # | Юнит | Примечание |
|---|------|------------|
| Пилот | hub-notification | минимальный blast-radius, event-consumer |
| 1 | hub-ai | BYOK / gateway; JWKS обязателен |
| 2 | hub-ws, hub-messenger, hub-calls | realtime-трио — **три** сервиса |
| Уже вне | hub-mail, meilisearch-indexer, LiveKit | не планировать заново |
| Ядро | hub-core (auth, rbac, tracker, …) | остаётся вместе |

Go/No-Go после пилота: lead-time↓, p95 API не хуже +10%, 0 инцидентов split за 2 недели.

## Чек-лист плана выноса

### A. Границы и контракты

- [ ] Домен/API boundary зафиксированы (sa + arch)
- [ ] Список REST paths, уходящих в сервис
- [ ] События `hub.<domain>.*` (publish + consume)
- [ ] CQRS-проекции вместо cross-domain read
- [ ] Contract-тесты producer↔consumer запланированы

### B. Данные

- [ ] Своя PG schema + migration Atlas/Ent
- [ ] Grants only на свою schema
- [ ] Нет cross-schema JOIN/FK в плане кода
- [ ] Стратегия сирот UUID (`*.deleted` + reconcile-job)
- [ ] Бэкапы при N схемах — решение для пилота

### C. Runtime и безопасность

- [ ] Consumer dedup по `event_id`
- [ ] JWT validate / JWKS plan (по этапу горизонта)
- [ ] Rate limits: core auth + per-service + Caddy 100/s
- [ ] SSRF/webhook инварианты TZ сохранены на границе
- [ ] `security-auditor` на boundaries в плане волны

### D. Infra / CI / rollback

- [ ] Path-filtered CI для сервиса
- [ ] Compose profile (dev split после приёмки)
- [ ] K3s skeleton (namespace, local-path PVC)
- [ ] Caddy route-flip + **rollback** процедура
- [ ] OTel headers end-to-end в критерии приёмки

### E. Docs и приёмка

- [ ] Docs: `architecture.md` C4, `DEPLOY.md`, `stack_and_tools.md`, STATE
- [ ] Success criteria измеримы (не «сервис вынесен»)
- [ ] Go/No-Go метрики в PLAN
- [ ] Phase 0 hardening (layering `*ent.Client`) учтён, если ещё не закрыт

## Полный шаблон PLAN выноса

```markdown
# PLAN: strangler — <domain> → hub-<name>

## Цель
Вынести <domain> в отдельный деплой-юнит с обратимым route-flip; монолит продолжает работать до flip.

## ADR constraints
- Schema-per-service, no shared tables
- NATS+outbox, idempotent consumers
- Caddy route-flip + rollback
- Порядок горизонта: …

## Won't
- Не трогать tracker / …
- Не big-bang multi-service

## Success criteria
- [ ] Schema + grants only
- [ ] Consumers dedup green (test)
- [ ] Route-flip staging PASS
- [ ] Rollback drill PASS
- [ ] p95 baseline documented
- [ ] Docs architecture + DEPLOY + STATE

## Атомарные задачи (пример порядка)
| # | Задача | Исп. | Depends | Verifier |
|---|--------|------|---------|----------|
| 1 | Boundary + events spec | sa/arch | — | sa-ver/arch-ver |
| 2 | Schema + grants | be/devops | #1 | be-ver |
| 3 | Outbox publish wiring | be | #2 | be-ver |
| 4 | Consumer + dedup | be | #3 | be-ver |
| 5 | Compose/K3s skeleton | devops | #2 | — |
| 6 | Caddy route-flip + rollback | devops | #4,#5 | security-auditor |
| 7 | Contract tests | qa-auto | #4 | qa-auto-ver |
| 8 | Docs sync | docs-writer | #6 | — |
| 9 | Go/No-Go report | team-lead+product | all | product-ver |

## Волны
- Wave A: #1
- Wave B: #2–#3 (после A)
- Wave C: #4–#5 parallel (карта файлов)
- Wave D: #6–#8
- Wave E: #9 goal-verify

## Risks
| Риск | Mitigation |
|------|------------|
| Распределённый монолит (shared DB) | grants + ревью arch-ver |
| Потеря трейсов | OTel headers gate |
| CI минуты | path filters |

## Rollback
1. Caddy revert route
2. Оставить consumer dual-run / drain
3. Критерий: p95 и error rate в норме
```

## Worked example (Hub): пилот notification

**Почему пилот:** чистый event-consumer, минимальный blast-radius (ADR §3).

| Пункт плана | Hub-якорь |
|-------------|-----------|
| События | subjects `hub.<domain>.*`, outbox `platform/outbox` |
| Messaging | NATS JetStream — единственный pub/sub (TZ) |
| Edge | Caddy в `deployments/` — не nginx |
| Уже вынесено (не путать) | hub-mail, meilisearch-indexer |
| Не включать в пилот | tracker, realtime-трио, hub-ai |
| Docs после реализации | architecture.md, DEPLOY.md, STATE |
| Приёмка | Go/No-Go композит ADR |

Делегирование задач волны — только через [wave-run](../wave-run/SKILL.md) / `team-lead.md` §5.2 с rule 56.

## DoD (плана, не кода)

- [ ] Все секции чек-листа A–E закрыты или явно Won't
- [ ] Атомарные задачи без BE+FE+Infra в одной строке
- [ ] Rollback и Go/No-Go в PLAN
- [ ] tracker не в scope без нового ADR
- [ ] security-auditor и arch-ver в owners
- [ ] Docs-sync задачи включены (rule 55)
- [ ] PLAN готов к `phase-plan` refinement → `wave-run`

## Anti-patterns

| Антипаттерн | Почему | Что вместо |
|-------------|--------|------------|
| Big-bang «вынесем всё» | ADR запрещает | strangler по одному |
| Shared tables «на время» | распределённый монолит | schema-per-service сразу |
| Вынос tracker «заодно» | вне горизонта | новый ADR + драйвер |
| Route-flip без rollback | прод-ловушка | drill rollback в criteria |
| gRPC «сразу» | избыточно на старте ADR | внутренний REST |
| Redis pub/sub вместо NATS | TZ forbidden | JetStream + outbox |
| Код выноса в docs-only freeze без решения | нарушает freeze | только PLAN, или ждать kickoff |
| План без OTel gate | невидимые инциденты | headers = предусловие |

## Связь

- ADR: `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md`
- Roadmap: `.planning/roadmaps/MICROSERVICES-ROADMAP.md` (если есть)
- Arch kit: `ai-expansion/04-PORTABLE-HUB-ARCH.md`, `10-MICROSERVICES-PLAYBOOK.md`
- Шаблон PE: `prompt-engineering/agents/templates/microservices-strangler-plan.md`
- Следующие skills: [phase-plan](../phase-plan/SKILL.md) → [wave-run](../wave-run/SKILL.md)
- Индекс: [gsd-ru/README.md](../README.md)
