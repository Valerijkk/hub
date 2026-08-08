# Аналитика: приложения

> Часть плейбука. Канон и оглавление — [`../analysis.md`](../analysis.md).

Шаблоны и справочные материалы.

---

## 10. Приложения

### 10.1 Полный чек-лист сессии аналитика

```
ПОДГОТОВКА
[ ] Загрузил все docs/ (vision, spec, architecture, stack, style, security, tz, test_plan,
    competitive_*, user_needs)
[ ] Прочитал свой skill (BA/SA/ARCH) и этот playbook
[ ] Проверил split в STATE.md (для ba1/ba2/sa1/sa2 — свой домен/модуль и диапазоны)
[ ] Проверил актуальный STATE.md, ROADMAP.md, COVERAGE.md

РАБОТА
[ ] Выявил требование из источников (не выдумал)
[ ] Оформил по шаблону своей роли (US/FR/ADR)
[ ] Проставил трассировку вверх и вниз
[ ] Сверил числа/контракты с кодом и другими docs
[ ] Проверил соответствие TZ (mandatory/forbidden)
[ ] Зафиксировал OPEN-гипотезы и вопросы к PO (если есть)

ЗАВЕРШЕНИЕ (DoD)
[ ] Docs sync: STATE.md + связанные docs
[ ] Обновил COVERAGE.md (SA обязательно)
[ ] Дополнил лог в skill-файле
[ ] Написал описание простым языком (2–5 предложений)
[ ] Передал верификатору, verified: false → *-ver
[ ] Убрал временные артефакты (rule 45), git status чист от мусора
```

### 10.2 Глоссарий (ядро)

| Термин | Значение в Hub |
|--------|----------------|
| Goal (G-XX) | SMART бизнес-цель в vision_and_scope |
| User Story (US-XX) | атомарное требование «Как/хочу/чтобы» |
| AC | Acceptance Criteria, проверяемые условия готовности (Gherkin) |
| FR (`PREFIX-NN`) | функциональное требование, поведение модуля |
| NFR (NFR-NN) | нефункциональное требование, числовое ограничение качества |
| MoSCoW | Must/Should/Could/Won't приоритизация |
| JTBD | Jobs To Be Done, «работа», которую нанимают продукт выполнить |
| Персона (P-XX) | обобщённый пользователь с ролью и потребностями |
| ERD | модель сущностей и связей (Ent/PostgreSQL) |
| DFD | диаграмма потока данных через слои |
| ADR | Architecture Decision Record |
| C4 | модель context/container/component |
| TD-XX | элемент реестра технического долга |
| DoR / DoD | Definition of Ready / Definition of Done |
| Трассируемость | цепочка Goal→US→FR→API→код→TC→autotest |
| Gap | разрыв между обещанным и реализованным/покрытым |
| Spike | короткое исследование неопределённости |
| Гипотеза (H-XX) | помеченное допущение с планом проверки |
| Optimistic lock | поле version для контроля конкурентной правки |
| Soft delete | логическое удаление, retention 30 дней |
| TZ | tz_constraints, mandatory/forbidden ограничения |

### 10.3 Соответствие ролей и GSD-skills (аналитика)

| Роль | Primary GSD | Secondary GSD |
|------|-------------|---------------|
| ba1/ba2 | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-new-milestone`, `gsd-review-backlog`, `gsd-capture` |
| sa1/sa2 | Hub `/coverage-audit`, `gsd-graphify` | `gsd-audit-uat`, `gsd-map-codebase` |
| arch | `gsd-ai-integration-phase`, `gsd-ui-phase` | `gsd-plan-phase`, `gsd-secure-phase` |
| ba-ver/sa-ver/arch-ver | `gsd-verify-work` | `gsd-audit-uat` / `gsd-secure-phase` |

### 10.4 Мини-шаблоны для быстрого копирования

US:

```
US-<M|S|C><NN> | Как <персона>, хочу <действие>, чтобы <ценность> | AC: <ссылка на Gherkin>
Priority: Must|Should|Could | Goal: G-XX | Персона: P-XX
```

Gherkin AC:

```gherkin
Scenario: <имя>
  Given <предусловие>
  When <действие>
  Then <результат>
```

FR:

```
REQ-ID: <PREFIX>-<NN> | Source: US-<NN> | Technical: <endpoint/schema/constraint> | NFR: NFR-<NN>
```

NFR:

```
NFR-<NN> | <требование> | <числовой порог + способ проверки>
```

API endpoint:

```
### <METHOD> /api/<path>
Auth: <Bearer|X-API-Key>; RBAC: <право>
Request: { ... }
Responses: 2xx {...}; 4xx {"error","request_id"}; 409/422 при конфликте/бизнес-правиле
Source FR: <PREFIX>-<NN>
```

ADR: см. §5.4. Гипотеза: см. §7.3. Spike: см. §7.4.

### 10.5 Ссылки

- Верификация: `.cursor/playbooks/verification.md`
- Регламент: `.cursor/REGULATION.md`
- GSD-матрица: `.cursor/GSD-INTEGRATION.md`
- TZ: `docs/tz_constraints.md` (+ `.cursor/rules/60-tz-constraints.mdc`)
- Docs sync: `.cursor/rules/55-docs-sync.mdc`
- Регрессия: `.cursor/rules/56-regression-verify.mdc`
- Агенты: `.cursor/agents/{ba1,ba2,sa1,sa2,arch}.md`

### 10.6 Сквозной worked-пример: фича «flaky-аналитика тест-кейсов»

Ниже показан полный проход одной фичи через все аналитические слои — как эталон того, что
именно попадает в каждый артефакт. Пример условный, но соответствует стилю проекта.

#### Слой 1 — BA (`vision_and_scope.md`)

```markdown
## Бизнес-цели (SMART) — дополнение
| ID   | Цель | Измеримость | Срок |
|------|------|-------------|------|
| G-02 | Сквозная traceability и качество тестирования | доля flaky-кейсов, стабилизированных за спринт, ≥ 30% | v5 |

## KPI — дополнение
| Метрика | Целевое значение | Текущий статус |
|---------|------------------|----------------|
| Помеченные flaky-кейсы | видимость 100% нестабильных | планируется (US-S16) |

## Портреты пользователей (используется существующая персона)
| Персона | Роль | Потребности | Ключевые сценарии |
|---------|------|-------------|-------------------|
| Мария, QA Lead | Тестировщик | видеть нестабильные кейсы, приоритизировать стабилизацию | фильтр flaky, история нестабильности |

## User Stories — дополнение
### Should Have
| ID | История | Acceptance Criteria |
|----|---------|---------------------|
| US-S16 | Как QA-инженер, хочу помечать кейс flaky и видеть историю нестабильности, чтобы приоритизировать стабилизацию | флаг flaky в карточке; фильтр «только flaky»; автоподсветка при 3+ исходах за 30 дней |
| US-S17 | Как тимлид, хочу видеть долю flaky по проекту на дашборде, чтобы контролировать качество | виджет «flaky rate» на проектном дашборде; тренд за 4 спринта |

JTBD (Мария): Когда мой автотест то падает, то проходит без изменений кода, я хочу быстро
находить такие кейсы, чтобы стабилизировать их до релиза.
```

Описание BA простым языком:

```
Добавлены две Should-US про flaky-кейсы: пометка/история (US-S16) и дашборд-метрика (US-S17).
Персона — Мария (QA Lead), цель G-02. AC покрывают happy path, фильтр и автоподсветку.
MoSCoW сверён, дублей нет. Передаю ba-ver, далее SA для декомпозиции.
```

#### Слой 2 — SA (`system_specification.md`, `wiki/API.md`, `COVERAGE.md`)

```markdown
## Функциональные требования (TMS) — дополнение
| ID | Source | Поведение | NFR link |
|----|--------|-----------|----------|
| TMS-14 | US-S16 | пометка кейса flaky; счётчик нестабильных прогонов за 30д; фильтр flaky | NFR-14, NFR-24 |
| TMS-15 | US-S17 | агрегат flaky-rate по проекту для дашборда | NFR-01 |

## Нефункциональные требования — дополнение
| ID | Требование | Значение |
|----|------------|----------|
| NFR-31 | Пересчёт flaky_count | при submit результата, идемпотентно, в транзакции |

## Модель данных — дополнение (Ent: TestCase)
- +flaky bool default false
- +flaky_count int default 0
- индекс (project_id, flaky); индекс (project_id, deleted_at)
- сохранены инварианты: id UUID v4, version int, created_at/updated_at UTC, deleted_at nullable
```

API-контракты (`wiki/API.md`):

```
### PATCH /api/tms/cases/:id
Auth: Bearer JWT; RBAC: tms.case.update
Request: { "flaky": true }
Responses:
  200 { "id":"uuid", "flaky":true, "flaky_count":0, "version":4 }
  403 { "error":"forbidden", "request_id":"uuid" }
  409 { "error":"version conflict", "request_id":"uuid" }
Source FR: TMS-14

### GET /api/tms/cases?flaky=true&offset=0&limit=50
Auth: Bearer JWT; RBAC: tms.case.read
Response 200: { "items":[...], "total":N, "offset":0, "limit":50 }
Ошибки: 400 при limit>100 {"error","request_id"}
Source FR: TMS-14
```

DFD (пересчёт flaky_count):

```
[hub-cli push-results] → POST /tms/runs/:id/results → [Auth] → [RBAC tms.result.submit]
  → [TMSService.SubmitResult] → [TMSRepository.Tx] → [PostgreSQL: TestResult + пересчёт TestCase.flaky_count]
  → [Outbox.Publish hub.tms.case.flaky_changed] → [Dashboard aggregate] / [Notification]
```

Строка COVERAGE.md:

```
| G-02 | US-S16 | 3 | TMS-14 | NFR-24 | PATCH /tms/cases/:id, GET /tms/cases?flaky | tms/handler_cases.go, service.go | TC-TMS-14 | testframework/tms/flaky_test.go | TBD | реализация pending |
```

#### Слой 3 — ARCH (`architecture.md`, ADR)

Вопрос границ: пересчёт flaky_count при submit — inline в TMSService или через event?
ARCH решает ADR:

```markdown
# ADR-042: Пересчёт flaky_count синхронно в транзакции submit

- Статус: Accepted
- Дата: 2026-07-02
- Связанные: US-S16, FR TMS-14, NFR-31, NFR-01

## Контекст
flaky_count нужно обновлять при каждом submit результата. Варианты: (1) синхронно в той же
транзакции; (2) асинхронно через NATS-consumer.

## Рассмотренные варианты
1. Синхронно в Tx — простота, консистентность; минимум задержки при малом числе результатов.
2. Асинхронно через event — развязка, но eventual consistency и сложнее отладка.

## Решение
Синхронный пересчёт в транзакции submit (вариант 1). Дашборд-агрегат (TMS-15) — асинхронно
через outbox-event hub.tms.case.flaky_changed, т.к. это read-model.

## Последствия
- Плюс: карточка кейса всегда консистентна; фильтр flaky точен сразу.
- Минус: submit чуть тяжелее (один UPDATE + пересчёт), в рамках NFR-01.
- Границы: tms владеет flaky_count; dashboard только читает через event (не тянет Ent tms).

## Проверка соответствия
tms/service.go SubmitResult в одной Tx обновляет TestResult и TestCase.flaky_count;
dashboard подписан на hub.tms.case.flaky_changed, не импортирует tms/ent.
```

C4 Component дополнение:

```
Модуль: tms (дополнение)
  service.SubmitResult → repository.Tx(TestResult+TestCase.flaky_count)
  publishes hub.tms.case.flaky_changed (outbox)
Модуль: dashboard
  subscribes hub.tms.case.flaky_changed → обновляет read-model flaky-rate (граница соблюдена)
```

#### Итог сквозного примера

Одна фича породила: 2 US + JTBD (BA) → 2 FR + 1 NFR + ERD-изменение + 2 контракта + DFD +
строка COVERAGE (SA) → 1 ADR + C4-дополнение (ARCH). Каждый нижний элемент ссылается на
верхний; каждый инвариант TZ (UUID, version, ошибки, пагинация, границы модулей) соблюдён.

### 10.7 Расширенный каталог good/bad

#### FR vs NFR (частая путаница)

```
GOOD FR (поведение): "TMS-14: пометка кейса flaky и фильтр flaky"
GOOD NFR (качество): "NFR-31: пересчёт flaky_count идемпотентен, в транзакции"
BAD: "NFR: сделать flaky-кейсы" — это FR, не NFR (поведение, не качество)
BAD: "FR: система должна быть быстрой" — это NFR, не FR (качество, не поведение)
```

#### US vs эпик

```
GOOD US: "Как QA, хочу помечать кейс flaky, чтобы фильтровать нестабильные"
BAD (эпик): "Как QA, хочу полноценную систему управления качеством тестов с аналитикой,
             дашбордами, автоподсветкой, экспортом и интеграцией с CI"
Разбить эпик на: пометка flaky (US-S16) + дашборд-метрика (US-S17) + экспорт (US-C) + CI (US-S)
```

#### AC: полнота видов

```
GOOD (три вида):
  happy:  When помечаю кейс flaky Then флаг сохранён, version инкрементирован
  alt:    When фильтрую по flaky Then вижу только flaky-кейсы с пагинацией
  error:  When шлю PATCH с устаревшим version Then 409 {"error","request_id"}
BAD (только happy): "флаг flaky работает" — нет границы и ошибки
```

#### ERD: инварианты

```
GOOD: TestCase{ id UUID, version int, deleted_at nullable, created_at/updated_at UTC }
BAD:  TestCase{ id "TC-42" string, нет version, hard delete } — три нарушения TZ сразу
```

#### API: коды ответов

```
GOOD: перечислены 200/201, 400, 401, 403, 404, 409 (version), 422 (бизнес), 429 (rate)
BAD:  только 200 и 500 — не описано поведение при конфликте/валидации/лимите
```

### 10.8 Частые вопросы (FAQ аналитика)

| Вопрос | Ответ |
|--------|-------|
| Куда писать новую сущность, если она между модулями двух SA? | Решает `arch` (ADR о владении), затем владеющий SA описывает ERD |
| Можно ли BA задать endpoint? | Нет, endpoint — зона SA; BA описывает поведение как AC |
| Что если US конфликтует с TZ? | US не проходит DoR; либо меняется, либо ADR на исключение |
| Нужен ли ADR на каждую библиотеку? | На значимый выбор/замену — да; на патч-обновление — нет (только stack_and_tools) |
| Кто ведёт COVERAGE.md? | SA — основной; BA добавляет строки US; BE/FE/QA заполняют Code/Tests |
| Когда spike, когда сразу FR? | Spike — при неопределённости реализуемости/выбора; иначе сразу FR |
| Как помечать незакрытые вопросы? | Гипотеза H-XX (§7.3) в разделе «Открытые вопросы» + STATE.md |
| Кто ставит verified: true? | Только соответствующий `*-ver`, никогда сам аналитик |

### 10.9 Быстрый справочник соответствий требований TZ по слоям

| Слой аналитики | Проверяемые пункты TZ |
|----------------|------------------------|
| BA | scope (что out of scope), Won't Have vs mandatory features by phase |
| SA (auth) | Magic Link only, JWT 15м, refresh 30д Redis, rate 3/10м |
| SA (data) | UUID v4, version, UTC, soft delete 30д |
| SA (api) | `{"error","request_id"}`, offset/cursor, rate limits |
| SA (real-time) | WS через Fiber, NATS JetStream only |
| SA (storage) | SeaweedFS + aws-sdk-go-v2, presigned, 25MB |
| ARCH (frontend) | TanStack Router, Tailwind 4 + Shadcn, no `!important` |
| ARCH (infra) | Compose(dev)/K3s(prod), Caddy, GitHub Actions, no DEV_MODE в prod |
| ARCH (security) | webhook signature, WS RBAC, search user-scope, ValidateOutboundURL |

### 10.10 Пошаговые операционные процедуры по ролям

#### Процедура ba1/ba2

```
1. Загрузить docs/ + свой skill + этот playbook.
2. Свериться со split в STATE.md: мой домен, мои диапазоны US/G/P.
3. Выявить требование: пройти 5 вопросов (Кто/Что/Зачем/Когда/Как поймём готово).
4. Сверить с competitive_* и user_needs_* — не изобретаю ли уже описанное.
5. Оформить: цель (SMART) при необходимости → персона/JTBD → US → AC (Gherkin) → MoSCoW.
6. Проверить: атомарность US, ≥3 вида AC, привязка к G-XX, приоритет, отсутствие TZ-нарушений.
7. Сверить глоссарий и нумерацию с параллельным BA.
8. Открытые вопросы → закрытые вопросы к PO / гипотезы H-XX.
9. Docs sync: STATE + vision. Лог в BA_skill. Описание простым языком.
10. verified:false → ba-ver. После true — handoff SA.
```

#### Процедура sa1/sa2

```
1. Загрузить docs/ + SA_skill + playbook. Свериться со split (мои модули, диапазоны FR/NFR).
2. Взять verified US от BA. Проверить их DoR (иначе вернуть BA).
3. Декомпозировать каждую US: FR (Source=US) → NFR link → ERD-изменения → API-контракт → DFD.
4. Соблюсти инварианты: UUID v4, version, UTC, soft delete; ошибки {"error","request_id"};
   пагинация offset/cursor; rate limits; коды 4xx/409/422.
5. Согласовать межмодульные связи с параллельным SA и с ARCH (FK, владение сущностью).
6. Пересчитать счётчики (endpoints) по коду, не «на глаз».
7. Обновить COVERAGE.md (строки US→FR→API→Code/Tests).
8. Docs sync: STATE + spec + wiki/API. Лог в SA_skill. Описание простым языком.
9. verified:false → sa-ver. При изменении модульности/стека → handoff ARCH; иначе → BE/FE.
```

#### Процедура arch

```
1. Загрузить docs/ (вкл. stack, style, security, tz) + ARCH_skill + playbook.
2. Взять verified FR/NFR от SA. Определить, нужно ли арх-решение (границы/стек/интеграция).
3. Обновить C4 (context/container/component) синхронно с кодом (grep/go.mod/routeTree).
4. Для выбора/замены библиотеки — критерии 5.3 + ADR (§5.4).
5. Проверить атрибуты качества (безопасность/произв./масштаб/надёжность) против NFR.
6. Проверить соответствие TZ; при исключении — ADR + согласование.
7. Обновить tech debt registry (TD-XX) с evidence file:line.
8. Docs sync: STATE + architecture (+ stack_and_tools/code_style при изменении).
9. verified:false → arch-ver. Handoff BE/FE с паттернами/границами/ADR.
```

#### Процедура координации параллельных потоков (таймлайн milestone)

```
T0  Оркестратор фиксирует split (домены/модули, диапазоны ID) в STATE.md.
T1  ba1 || ba2 работают параллельно над своими доменами (US/AC/MoSCoW).
T2  Синк BA: владелец общих разделов сводит MoSCoW/KPI/глоссарий, снимает дубли.
T3  ba-ver проверяет обе части vision (независимо).
T4  sa1 || sa2 декомпозируют verified US своих модулей (FR/ERD/API).
T5  Синк SA: сведение endpoint-inventory и межмодульных связей; согласование с ARCH.
T6  sa-ver проверяет spec.
T7  arch решает cross-cutting (границы, ADR, deployment, атрибуты качества).
T8  arch-ver проверяет architecture.
T9  Handoff BE/FE; COVERAGE.md отражает готовность к разработке.
```

### 10.11 Техники выявления требований (детально)

| Техника | Когда применять | Результат |
|---------|-----------------|-----------|
| Интервью с PO (закрытые вопросы) | неоднозначность бизнес-намерения | уточнённая US/AC |
| Анализ конкурентов | поиск gap и паритета | backlog-идеи, обоснование ценности |
| Анализ персон/JTBD | понять мотивацию и контекст | привязка US к реальной работе |
| Gap-анализ по COVERAGE | найти невыполненные обещания | backlog реализации/тестов |
| Аудит документов (drift) | расхождение docs vs код | список рассинхронизаций |
| Сценарный анализ | сложные многошаговые потоки | основной/альтернативный/исключительный сценарии |
| Прототип/sketch | неясен UX | gsd-sketch, ранняя обратная связь |
| Spike | неясна реализуемость | знание для FR/ADR |

### 10.12 Связь аналитики с тестированием и верификацией

Аналитические артефакты — прямой вход для QA и верификаторов:

- Каждая AC (§3.7) должна породить как минимум один тест-кейс (`test_plan.md` / TMS) и, при
  автоматизации, один autotest (`testframework/**`). Это проверяет `qa-ver` (см.
  `verification.md`, чек-лист qa-manual-ver и qa-auto-ver).
- Три вида AC (happy/alt/error) напрямую соответствуют техникам тест-дизайна: классы
  эквивалентности, граничные значения, негативные сценарии.
- FR/NFR — критерий приёмки для `be-ver`/`fe-ver`: реализация сверяется с контрактом, а не
  «работает ли вообще».
- Матрица COVERAGE — общий язык SA и QA: строки без Tests = задача QA; строки без Code =
  задача BE/FE.
- ADR и TZ — критерий для `arch-ver` и `code-reviewer`: любое нарушение mandatory/forbidden
  без ADR блокирует.

Полное описание процесса проверки — в парном документе `.cursor/playbooks/verification.md`.

### 10.13a Каталог измеримых NFR по категориям (справочник SA)

Готовые формулировки, чтобы NFR всегда были числовыми и проверяемыми. Пороговые значения —
из `tz_constraints.md` и `system_specification.md`; при новых NFR подбирать по аналогии.

| Категория | Пример измеримой формулировки | Способ проверки |
|-----------|-------------------------------|-----------------|
| Производительность | p95 ответа CRUD < 500ms (NFR-01) | нагрузочный тест / профиль |
| Пропускная способность | 50+ конкурентных WS на инстанс (NFR-02) | стресс-тест WS |
| Лимиты | auth 3/10мин, API 100/сек, msg 60/мин, файлы 10/мин | интеграционный тест лимитера |
| Время жизни | JWT 15м, refresh 30д, soft delete 30д | проверка конфигов/поведения |
| Целостность | version int на изменяемой Ent (409 при конфликте) | тест конкурентной правки |
| Идентификаторы | UUID v4 PK, string запрещён | статическая проверка схем |
| Безопасность | webhook signature обязателен, skip запрещён | негативный тест без подписи |
| Изоляция данных | search только user-scope | тест доступа чужих данных |
| Формат | ошибки `{"error","request_id"}` | контрактный тест |
| Доступность (a11y) | WCAG 2.1 AA: контраст 4.5:1 текст, 3:1 UI | axe / Lighthouse |
| Качество фронта | Lighthouse ≥ 90 (цель) | Lighthouse CI |
| Покрытие | overall ≥ 50% (gate), security-critical ≥ 80% | coverage-прогон Docker |

#### Good / Bad формулировки NFR

```
GOOD: "NFR-32: presigned upload завершается без прохождения тела через backend; backend buffer ≤ 25MB"
BAD:  "NFR-32: загрузка файлов должна быть эффективной" (нет числа, нет проверки)
```

### 10.13b Фреймворки приоритизации backlog (справочник BA)

MoSCoW — основной механизм, но для сортировки внутри Should/Could полезны вспомогательные
метрики. Они не заменяют MoSCoW, а уточняют порядок.

| Фреймворк | Формула/суть | Когда |
|-----------|--------------|-------|
| Value vs Effort | быстрые победы = высокая ценность + низкие усилия | сортировка внутри приоритета |
| RICE | (Reach × Impact × Confidence) / Effort | сравнение конкурирующих US |
| Kano | must-be / performance / delighter | классификация ожиданий |
| Cost of Delay | ценность × срочность | что делать раньше при равном приоритете |

Шаблон оценки кандидата backlog:

```
US-кандидат: <ID>
MoSCoW: Should
Reach: <сколько персон/сценариев затрагивает>
Impact: <вклад в цель G-XX, шкала 0.25..3>
Confidence: <уверенность оценки, %>
Effort: <оценка SA/ARCH, story points>
Вывод: RICE=<число> → позиция в backlog
```

Правило: приоритизация — предложение BA; фиксация в ROADMAP — оркестратор/SA. BA не правит
ROADMAP напрямую (зона ответственности).

### 10.13 Итоговая памятка (одна страница)

```
BA:   Goal(SMART) → Персона/JTBD → US(атомарна) → AC(happy/alt/error) → MoSCoW(без дублей)
SA:   FR(Source=US) → NFR(число+проверка) → ERD(UUID/version/UTC/soft) → API({"error","request_id"},
      пагинация, коды) → DFD → COVERAGE
ARCH: C4(синхронно с кодом) → выбор библиотек(критерии+ADR) → deployment → атрибуты качества
      → TZ(mandatory/forbidden) → TD-registry
ВСЕ:  трассировка вверх/вниз · docs sync · описание простым языком · verified ставит только *-ver
НИКОГДА: расплывчатые US · нетестируемые NFR · тихие допущения · нарушение TZ без ADR ·
         drift чисел · self-verified · правка чужого артефакта
```

### 10.14 Пример возврата артефакта от верификатора (что видит аналитик)

Аналитик должен понимать, как выглядит FAIL, чтобы не доводить до него. Реальный стиль
замечаний (см. логи `ba-ver`/`sa-ver` в skill-файлах):

```
BA_VER → BA (FAIL, 2026-07-02)
Артефакт: vision_and_scope.md (US-S16..S17)
Замечания:
  [HIGH]   Web Push числится и в Should (US-S05), и в Could (US-C01) — устранить дубль MoSCoW.
  [MEDIUM] KPI «flaky rate» без источника факта — добавить ссылку/статус.
  [LOW]    US-S17 AC «тренд за 4 спринта» — уточнить, что считается спринтом (глоссарий).
Действие: вернуть BA, повторная проверка после правок.
```

```
SA_VER → SA (FAIL, 2026-07-02)
Артефакт: system_specification.md + wiki/API.md
Замечания:
  [CRITICAL] TestCase без поля version в ERD-дополнении — нарушение NFR-24 (optimistic lock).
  [HIGH]     GET /tms/cases?flaky без ограничения limit (>100) — нарушение NFR-14.
  [MEDIUM]   Endpoint inventory: заявлено tms 54, пересчёт по RegisterRoutes даёт 52 — свести.
Действие: исправить, приложить пересчёт, вернуть на проверку.
```

Аналитик исправляет по списку (не спорит «на глаз»), прикладывает доказательства (пересчёт,
diff docs), и только тогда получает `verified: true`. Подробный формат отчёта верификации —
в `.cursor/playbooks/verification.md`, раздел 5.

### 10.15 Мини-DoR по ролям (быстрая самопроверка перед handoff)

```
BA готов передать SA, если:
  US атомарна · AC(3 вида) · Goal+MoSCoW · персона · без TZ-нарушений · дубли сняты

SA готов передать BE/FE, если:
  FR(Source=US) · NFR(число) · ERD(инварианты) · контракт(ошибки/пагинация/коды) · COVERAGE обновлён

ARCH готов передать BE/FE, если:
  C4 синхронен с кодом · ADR на значимый выбор · TZ соблюдён · атрибуты качества покрывают NFR ·
  границы модулей заданы · TD зафиксирован
```

---

*Hub Analysis Playbook — эталон для ba1/ba2/sa1/sa2/arch и ba-ver/sa-ver/arch-ver.*
