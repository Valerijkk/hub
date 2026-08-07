# Верификация: приложения

> Часть плейбука. Канон и оглавление — [`../verification.md`](../verification.md).

Шаблоны отчётов и справочные материалы.

---

## 9. Приложения

### 9.1 Единый чек-лист сессии верификатора

```
ПОДГОТОВКА
[ ] Загрузил все docs/ + tz_constraints + свой *_VER_skill + этот playbook
[ ] Получил артефакт + описание исполнителя простым языком
[ ] Открыл STATE.md/COVERAGE.md/ROADMAP.md для контекста
[ ] Для кода — открыл сам diff/файлы, не только summary

ВОСПРОИЗВЕДЕНИЕ
[ ] Прошёл сценарий из требований самостоятельно
[ ] Запустил команды/пересчитал числа сам (не принял чужой вывод)

СВЕРКА
[ ] Соответствие docs/ (spec/architecture/vision)
[ ] Соответствие TZ (mandatory соблюдены, forbidden отсутствуют)
[ ] Прошёл пофактовый чек-лист своей роли (раздел 3)

РЕГРЕССИЯ
[ ] Green профильный прогон / /qa-gate
[ ] Лог приложен в отчёт
[ ] При FAIL — test-triage выполнен

ВЕРДИКТ
[ ] Замечания классифицированы по severity, с доказательствами
[ ] verified: true|false + YYYY-MM-DD HH:MM
[ ] Отчёт (раздел 5) заполнен; skill-лог дополнен
[ ] CRITICAL/HIGH → возврат; только MEDIUM/LOW → возможен PASS

ЗАВЕРШЕНИЕ
[ ] git status чист от coverage-мусора (rule 45)
[ ] Обновил статус верификации у исполнителя и у себя
```

### 9.2 Команды воспроизведения (сводка)

```bash
# Backend
cd backend && go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./...
go test -race -count=3 ./internal/<module>/...
go test -cover ./internal/<module>/...

# Backend Docker (канонично)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --cpus 2.0 --memory 4g --network deployments_default -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."

# Frontend
cd frontend && node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit && npm run lint && npx vitest run
npx playwright test --repeat-each=2

# Gate / audit
/qa-gate
/coverage-audit
/test-triage
```

### 9.3 Матрица роль → артефакт → GSD → блокирующая severity

| Верификатор | Артефакт | Primary GSD | Что критично |
|-------------|----------|-------------|--------------|
| ba-ver | vision_and_scope | gsd-verify-work | SMART, AC, MoSCoW-дубли, drift KPI |
| sa-ver | system_specification, wiki/API | gsd-verify-work, gsd-audit-uat | ERD-инварианты, формат ошибок, пагинация, counts |
| arch-ver | architecture, ADR | gsd-verify-work, gsd-secure-phase | TZ, SOLID/KISS, stale C4, security-арх |
| be-ver | backend код | gsd-verify-work, gsd-code-review | layering, security-инварианты, -race, triage |
| fe-ver | frontend код | gsd-verify-work, gsd-ui-review | Rules of Hooks, a11y, XSS, key, lint |
| qa-auto-ver | автотесты | gsd-validate-phase | детерминизм, flaky, sleep, покрытие |
| qa-manual-ver | test_plan/TMS | gsd-validate-phase | AC↔TC, техники тест-дизайна, трассируемость |
| product-ver | фича vs vision | gsd-verify-work | scope, KPI, JTBD, регресс сценариев |
| designer-ver | UI | gsd-ui-review | токены, a11y, отзывчивость, состояния |
| code-reviewer | diff PR | gsd-code-review | баги, security, паттерны, TZ, docs sync |

### 9.4 Глоссарий верификации

| Термин | Значение |
|--------|----------|
| Воспроизведение | самостоятельное прохождение сценария/запуск/пересчёт верификатором |
| Независимость | верификатор не получает готовых ответов и не доверяет «на слово» |
| Вердикт | verified: true\|false + дата/время |
| Severity | CRITICAL/HIGH/MEDIUM/LOW — уровень воздействия дефекта |
| Baseline | принятое состояние проекта; CRITICAL/HIGH туда не пускают |
| Регрессия | профильный прогон тестов/gate, обязателен с логом |
| Test triage | протокол «функционал → код vs тест» перед фиксом упавшего теста |
| REVIEW/VERIFICATION.md | отчёт верификации по шаблону раздела 5 |
| Drift | расхождение docs vs docs или docs vs код (числа, контракты) |
| Green + лог | зелёный прогон с приложенным фактическим выводом |
| Skill-лог | таблица проверок в `*_VER_skill.md` |

### 9.5 Сквозной пример: верификация фичи «flaky-аналитика»

Показывает, как разные верификаторы проверяют один и тот же сквозной результат (см. worked-пример
в `analysis.md` §10.6).

```
ba-ver (vision, US-S16/S17):
  - Воспроизвёл: прошёл таблицу US, проверил формат «Как/хочу/чтобы», AC (3 вида).
  - Сверка: MoSCoW без дублей; KPI «flaky rate» — [MEDIUM] нет источника факта.
  - Вердикт: verified: false → добавить источник KPI. После правки → true 2026-07-02 12:10.

sa-ver (spec, wiki/API):
  - Воспроизвёл: пересчитал tms routes по RegisterRoutes; проверил ERD TestCase.
  - Замечания: [CRITICAL] нет version у TestCase (NFR-24); [HIGH] GET /tms/cases?flaky без limit≤100.
  - Вердикт: verified: false → исправить. После правки + пересчёт → true 2026-07-02 13:05.

arch-ver (ADR-042, C4):
  - Воспроизвёл: сверил границы (dashboard не импортирует tms/ent); проверил ADR-структуру.
  - Замечания: нет (2 варианта, последствия, проверка соответствия присутствуют).
  - Вердикт: verified: true 2026-07-02 13:40.

be-ver (tms/service.go, handler_cases.go):
  - Воспроизвёл: go test -race -count=1 ./internal/tms/... → ok (лог приложен).
  - Сверка: layering ок; 409 при version-конфликте есть; ошибки {"error","request_id"}.
  - Замечания: нет CRITICAL/HIGH.
  - Вердикт: verified: true 2026-07-02 15:20 (лог регрессии в отчёте).

qa-auto-ver (testframework/tms/flaky_test.go):
  - Воспроизвёл: go test -race -count=3 → зелёно; grep sleep → [MEDIUM] time.Sleep(2s).
  - Трассировка: тест ссылается на TC-TMS-14 — ок.
  - Вердикт: verified: true с MEDIUM (заменить sleep на poll) — не блокирует.

qa-manual-ver (test_plan):
  - Сверка AC↔TC: happy/фильтр/409 покрыты; границы limit (0,1,100,101) присутствуют.
  - Вердикт: verified: true 2026-07-02 16:00.

product-ver (фича vs vision):
  - Прошёл как персона Мария: пометка/фильтр/автоподсветка → AC выполнены; scope соблюдён.
  - Вердикт: verified: true 2026-07-02 16:30.
```

### 9.8 Эталонные отчёты FAIL по ролям

Ниже — заполненные отчёты по шаблону раздела 5, показывающие ожидаемый уровень конкретики.

#### be-ver FAIL

```markdown
# Verification Report — be-ver
- Артефакт: backend/internal/tms (flaky feature)
- Исполнитель: be
- Верификатор: be-ver
- Дата/время: 2026-07-02 14:05
- Вердикт: verified: false

## Контекст и воспроизведение
Проверял реализацию TMS-14 (пометка flaky + фильтр). Запустил go test -race ./internal/tms/...

## Регрессия
- Команда: cd backend && go test -race -count=1 ./internal/tms/...
- Результат: PASS
- Лог:
```
ok  hub/internal/tms  9.812s
```

## Сверка с docs/TZ
Контракт GET /tms/cases?flaky сверен с wiki/API.md; NFR-13/14/24 из spec.

## Замечания
| # | Severity | Файл:строка | Ожидалось | Фактически | Действие |
|---|----------|-------------|-----------|------------|----------|
| 1 | HIGH | tms/handler_cases.go:88 | limit ограничен ≤100 | limit не валидируется | добавить проверку, вернуть |
| 2 | HIGH | tms/service.go:154 | ошибка обёрнута %w | return err без контекста | обернуть fmt.Errorf |
| 3 | MEDIUM | tms/service.go:201 | slog c контекстом | fmt.Println для отладки | убрать/заменить slog |

## Итог
FAIL: 2×HIGH блокируют. Возврат be. Регрессия зелёная, но контракт/стиль нарушены.
```

#### sa-ver FAIL

```markdown
# Verification Report — sa-ver
- Артефакт: system_specification.md + wiki/API.md (TMS-14/15)
- Верификатор: sa-ver
- Дата/время: 2026-07-02 13:00
- Вердикт: verified: false

## Воспроизведение
Независимый пересчёт tms routes по RegisterRoutes; проверка ERD TestCase.

## Замечания
| # | Severity | Раздел | Ожидалось | Фактически | Действие |
|---|----------|--------|-----------|------------|----------|
| 1 | CRITICAL | ERD TestCase | version int (NFR-24) | поле отсутствует | добавить version |
| 2 | HIGH | wiki GET /tms/cases | пагинация limit≤100 | без ограничения | описать в контракте |
| 3 | MEDIUM | inventory | tms=54 | пересчёт=52 | свести число |

## Итог
FAIL: CRITICAL (version). Возврат SA, приложить пересчёт routes.
```

#### fe-ver FAIL

```markdown
# Verification Report — fe-ver
- Артефакт: frontend/src/modules/tms/FlakyList.tsx
- Верификатор: fe-ver
- Дата/время: 2026-07-02 17:20
- Вердикт: verified: false

## Регрессия
- Команда: npx tsc --noEmit && npm run lint && npx vitest run
- Результат: PASS
- Лог: Test Files 37 passed (37) · Tests 99 passed (99)

## Замечания
| # | Severity | Файл:строка | Ожидалось | Фактически | Действие |
|---|----------|-------------|-----------|------------|----------|
| 1 | HIGH | FlakyList.tsx:24 | хуки до return | useEffect после if(!id) return | перенести хук вверх |
| 2 | HIGH | FlakyList.tsx:60 | key={item.id} | key={index} | стабильный ключ |
| 3 | MEDIUM | FlakyList.tsx:75 | aria-label на иконке-кнопке | отсутствует | добавить aria |

## Итог
FAIL: Rules of Hooks + key=index. Возврат fe. Регрессия зелёная, но дефекты корректности.
```

#### qa-auto-ver FAIL

```markdown
# Verification Report — qa-auto-ver
- Артефакт: testframework/tms/flaky_test.go
- Верификатор: qa-auto-ver
- Дата/время: 2026-07-02 18:00
- Вердикт: verified: false

## Воспроизведение
go test -race -count=3 ./internal/tms/... → 2/3 PASS, 1 FAIL (нестабильно).

## Замечания
| # | Severity | Файл:строка | Ожидалось | Фактически | Действие |
|---|----------|-------------|-----------|------------|----------|
| 1 | HIGH | flaky_test.go:33 | детерминизм | зависит от time.Now() без фиксации | заморозить время/clock |
| 2 | MEDIUM | flaky_test.go:20 | poll по условию | time.Sleep(2s) | заменить на waitFor |
| 3 | MEDIUM | flaky_test.go | ссылка на TC | нет привязки к TC-TMS-14 | добавить трассировку |

## Итог
FAIL: тест flaky (HIGH). Возврат qa. Перед фиксом — triage не требуется (ROOT=test, недетерминизм).
```

### 9.9 Верификация параллельной работы (ba1/ba2, sa1/sa2)

Когда два аналитика писали в один файл параллельно (см. `analysis.md` §2.3–2.4), верификатор
дополнительно проверяет швы между потоками:

```
ba-ver для vision после ba1+ba2:
[ ] Нет коллизий ID (US/G/P) между потоками — диапазоны соблюдены
[ ] Глоссарий един (нет синонимов «канал»/«беседа»)
[ ] MoSCoW сведён целиком, нет дублей приоритета на стыке доменов
[ ] Общие разделы (KPI, глоссарий) непротиворечивы, редактировались владельцем
[ ] Числа в общих метриках = сумма/сведение частей, не «на глаз»

sa-ver для spec после sa1+sa2:
[ ] Нет коллизий FR/NFR-номеров между потоками
[ ] Межмодульные связи (FK между доменами) согласованы и описаны с обеих сторон
[ ] Endpoint inventory = сумма per-module по коду (независимый пересчёт)
[ ] Общий раздел соглашений (ошибки/пагинация) един, не продублирован по-разному
[ ] Нет прямого доступа модуля A к Ent модуля B (граница)
```

Реальный класс дефектов на стыке: два SA независимо описали смежную связь по-разному, или
сумма per-module routes не сошлась с заявленным total (кейс 293 vs 285). Верификатор сводит
части сам и ловит расхождение.

### 9.10 Deep-dive: проверка security-инвариантов (для be-ver, arch-ver, code-reviewer)

Security — зона максимального внимания (CRITICAL). Пошаговая проверка ключевых инвариантов Hub:

```
1) Webhook signature (SEC-02)
   - Найти обработчик webhook (github/gitlab/livekit)
   - Убедиться: подпись body проверяется ВСЕГДА, нет ветки skip при отсутствии claim
   - Негативный тест: запрос без/с неверной подписью → отклонён
   Severity при нарушении: CRITICAL

2) WS room authorization (NF-02)
   - Подписка на room только при membership/RBAC
   - Проверить ws.Authorizer вызывается до доставки событий room
   Severity: CRITICAL

3) Search user-scope (NF-03)
   - Каждый поисковый запрос фильтруется по project/channel membership
   - Негативный тест: поиск не возвращает чужие данные
   Severity: CRITICAL

4) SSRF (outbound URL)
   - user-controlled URL идёт через SafeHTTPClient + ValidateOutboundURL
   - Нет http.DefaultClient для таких URL
   Severity: CRITICAL/HIGH

5) Секреты
   - subtle.ConstantTimeCompare для сравнения секретов (не !=)
   - нет hardcoded credentials; нет DEV_MODE=true в prod
   Severity: CRITICAL

6) Память/DoS
   - io.LimitReader вместо io.ReadAll на входных потоках
   - BodyLimit 25MB соблюдён
   Severity: HIGH
```

Для каждого инварианта верификатор ищет **и** позитивный путь (проверка есть), **и** негативный
тест (нарушение отклоняется). Отсутствие негативного теста на security-инвариант — минимум
MEDIUM, а отсутствие самой проверки — CRITICAL.

### 9.11 Deep-dive: проверка a11y/WCAG (для fe-ver, designer-ver)

```
Контраст:   текст ≥ 4.5:1, крупный текст/UI ≥ 3:1 (axe / ручная проверка токенов)
Клавиатура: все интерактивные элементы достижимы Tab; focus-visible виден; нет ловушек фокуса
ARIA:       кнопки-иконки имеют aria-label; роли соответствуют семантике; live-region для динамики
Формы:      label связан с input; ошибки объявляются программно
Порядок:    логический порядок табуляции соответствует визуальному
Семантика:  button type задан явно; ссылки vs кнопки не перепутаны
Проверка:   npx playwright test (a11y spec) / axe; Lighthouse Accessibility
```

Реальный ориентир проекта: login axe PASS, E2E a11y spec присутствует; Lighthouse — цель ≥90
(gate non-blocking сейчас, фиксировать фактический статус, не выдавать желаемое за факт).

### 9.12 Верификация в волновом исполнении (wave)

При параллельном исполнении (`gsd-execute-phase`, до 4 агентов, см. GSD-INTEGRATION) каждый
исполнитель волны сопровождается верификацией своего слоя:

```
Wave N: параллельные be/fe/qa агенты
  → каждый прогоняет test-triage при FAIL
  → каждый запускает /qa-gate subset на затронутых пакетах
  → соответствующий *-ver верифицирует свой артефакт (независимо)
  → оркестратор мержит только verified: true артефакты
  → `/handoff` + `/handoff-sync` (STATE + COVERAGE)
Wave N+1: зависит от verified-артефактов Wave N
```

Правила верификации в волне:

```
[ ] Каждый параллельный артефакт проверяется своим верификатором (не «оптом»)
[ ] verified: false в любом артефакте волны блокирует мерж этого артефакта
[ ] Регрессия — на затронутых пакетах каждого артефакта, лог у каждого
[ ] Кросс-артефактные конфликты (два агента тронули один файл) → отдельная проверка швов (§9.9)
[ ] Wave N+1 не стартует, пока зависимости из Wave N не verified: true
```

### 9.13 FAQ верификатора

| Вопрос | Ответ |
|--------|-------|
| Можно ли verified: true при MEDIUM? | Да, MEDIUM/LOW не блокируют; зафиксировать и завести в backlog |
| Можно ли самому исправить опечатку и поставить PASS? | Опечатку в СВОЁМ отчёте — да; в артефакте исполнителя — нет, вернуть |
| Что если регрессия зелёная, но код нарушает контракт? | FAIL: green прогон не отменяет нарушения контракта/security |
| Что если исполнитель не согласен с severity? | Severity — по факту воздействия; решение верификатора, эскалация к аудиту при споре |
| Нужно ли перепроверять после фикса? | Да, повторная верификация обязательна (полный протокол, не только исправленный пункт) |
| Кто проверяет верификатора? | Для критичных изменений — `verifier`/аудит-цикл; вердикты датируются и логируются |
| Можно ли пропустить регрессию для docs-only? | Для docs без влияния на поведение — да; но сверка непротиворечивости обязательна |
| Что делать с flaky-тестом исполнителя? | FAIL (детерминизм — требование qa-auto-ver), ROOT=test, вернуть на стабилизацию |
| Как проверять число endpoints? | Независимый пересчёт по RegisterRoutes, не принимать заявленное |
| Куда писать вердикт? | В отчёт (раздел 5) + skill-лог + статус verified у исполнителя |

### 9.14 Расширенный good/bad каталог верификации

#### Регрессия и лог

```
GOOD: секция «Регрессия» с реальным выводом "ok hub/internal/tms 9.8s" и командой
BAD:  "тесты прошли" без команды и вывода → лог отсутствует → verified: true запрещён
```

#### Классификация severity

```
GOOD: "пропуск webhook signature → CRITICAL (SEC-02), блокирует"
BAD:  "webhook подпись бы не помешала → LOW" → занижение security-CRITICAL
```

#### Доказательство замечания

```
GOOD: "handler_cases.go:88 — limit не валидируется; ожидалось ≤100 (NFR-14); фактически принимает 100000"
BAD:  "пагинация какая-то не такая" → без file:line, без ожидалось/фактически
```

#### Воспроизведение

```
GOOD: "запустил go test -race -count=3, 1/3 упал → flaky подтверждён логом"
BAD:  "исполнитель сказал стабильно, поверил" → нет воспроизведения
```

#### Возврат vs переписывание

```
GOOD: "замечание [HIGH] с предложением io.LimitReader, возврат be"
BAD:  "сам заменил io.ReadAll на io.LimitReader и поставил PASS" → потеря независимости
```

### 9.16 Топ-дефекты по ролям (что искать в первую очередь)

Приоритизированные списки — «где статистически чаще всего прячется проблема». Верификатор
начинает проверку с них (§1.6, §2.7).

```
be-ver (топ-5)
  1. Формат ошибки не {"error","request_id"} (NFR-13)
  2. *ent.Client / GetDB() в service/handler (layering)
  3. Отсутствие 409 при version-конфликте
  4. Незавёрнутые ошибки (return err без %w)
  5. Пропущенный негативный/граничный тест

fe-ver (топ-5)
  1. Rules of Hooks (хук после условия/return)
  2. key={index} в списках
  3. a11y: нет aria-label / контраст / focus-visible
  4. Токен в localStorage вместо памяти (XSS-риск)
  5. !important / изменение поведения Radix

sa-ver (топ-5)
  1. ERD без version/UUID/soft delete
  2. Контракт без пагинации/кодов 4xx
  3. Drift endpoint count (заявлено ≠ пересчёт)
  4. FR без Source (US)
  5. NFR нетестируемо (без числа)

ba-ver (топ-5)
  1. Дубль MoSCoW-приоритета
  2. US без AC или только happy path
  3. KPI-желаемое вместо факта
  4. Цель не SMART
  5. Коллизия ID между ba1/ba2

arch-ver (топ-5)
  1. Нарушение TZ без ADR (forbidden/убран mandatory)
  2. Stale C4/число модулей vs код
  3. ADR без вариантов/последствий
  4. Неоправданная сложность (нарушение KISS)
  5. Пропущенный security-атрибут на уровне архитектуры

qa-auto-ver (топ-5)
  1. Flaky/недетерминизм
  2. time.Sleep вместо poll
  3. Пустой ассерт (err==nil без проверки результата)
  4. Нет привязки TC↔autotest
  5. Один прогон без -count/-race

qa-manual-ver (топ-5)
  1. AC без TC (разрыв трассировки)
  2. Только happy path (нет границ/негатива)
  3. Неоднозначные шаги TC
  4. Устаревшая матрица покрытия
  5. Некорректный баг-репорт (нет шагов/факт vs ожидание)

product-ver / designer-ver (топ-5)
  1. Реализация не решает JTBD персоны
  2. Выход за scope (реализовано Won't Have)
  3. Регресс существующего сценария
  4. Рассинхрон токенов/состояний UI
  5. a11y-нарушение в новом экране
```

### 9.17 Метрики качества верификации

Верификатор оценивается не числом PASS, а качеством ловли дефектов. Ориентиры:

| Метрика | Хорошо | Плохо |
|---------|--------|-------|
| Пропущенные CRITICAL (найдены позже в аудите/prod) | 0 | >0 (ложноотрицательные) |
| Доля замечаний с доказательством | 100% | <100% (голословные) |
| verified: true с приложенным логом регрессии | 100% | <100% (нарушение §6) |
| Ложноположительные (FAIL без реальной проблемы) | минимум | частые (трение) |
| Severity корректна пост-фактум | стабильно | занижения/завышения |

Аудит-цикл (`.cursor/rules/56-regression-verify.mdc`) периодически re-verify находки против
кода — это проверка самих верификаторов. Цель до «production-ready»: CRITICAL = 0,
product score ≥ 75.

### 9.15 Соответствие разделов REGULATION

| Раздел playbook | REGULATION |
|-----------------|-----------|
| §1 Философия | §4 Принципы качества (одна нейросеть проверяет другую) |
| §2 Протокол | §3 Цикл верификации |
| §4 Severity/baseline | §4.3 регрессия обязательна |
| §6 Регрессия | §7 Обязательная регрессия |
| §6.4 Test triage | §9/§10 Test Triage |
| §7 GSD | §8 GSD Skills Integration |
| §3.11 docs sync | §11 Documentation Sync |

### 9.6 Итоговая памятка верификатора (одна страница)

```
ПРОТОКОЛ: загрузить контекст → воспроизвести → сверить (docs+TZ) → регрессия(green+лог)
          → вердикт(verified±дата) → замечания(severity) → вернуть/пропустить

НЕЗАВИСИМОСТЬ: не доверяй на слово · запускай сам · пересчитывай сам · не переписывай

БЛОКИРУЕТ: CRITICAL, HIGH → verified: false
ПРОПУСКАЕТ: только MEDIUM/LOW (с фиксацией)

ЗАПРЕЩЕНО: verified без green+лог · озеленение без triage · проверка «на глаз» ·
           смягчение severity · замечание без доказательства · приём чужого лога

ОТЧЁТ: REVIEW/VERIFICATION.md (контекст, регрессия+лог, сверка, замечания-таблица, итог)
       + запись в *_VER_skill.md + статус verified у исполнителя
```

### 9.18 Мета-проверка отчёта верификации (само-контроль перед публикацией)

Прежде чем опубликовать вердикт, верификатор проверяет собственный отчёт:

```
[ ] Есть секция «Воспроизведение» с конкретикой (что и как проверено)
[ ] Есть секция «Регрессия» с командой и фактическим логом (для кода — обязательно)
[ ] Каждое замечание имеет severity, file:line/раздел, «ожидалось vs фактически», действие
[ ] Замечания отсортированы: CRITICAL → HIGH → MEDIUM → LOW
[ ] Вердикт согласован с замечаниями (есть CRITICAL/HIGH ⇒ verified: false)
[ ] Дата/время проставлены
[ ] Нет голословных пунктов («что-то не так») без доказательства
[ ] Отчёт самодостаточен: исполнитель исправит всё по списку без доп. вопросов
[ ] Запись продублирована в skill-лог и статус verified обновлён
```

Если хоть один пункт не выполнен — отчёт не готов, даже если проверка проведена. Плохой отчёт
= невоспроизводимая верификация = фактически «на глаз».

### 9.19 Быстрый разбор: как отличить PASS от FAIL за 30 секунд

```
СТОП-сигналы (мгновенный FAIL, дальше — полный список замечаний):
  - нарушение TZ mandatory/forbidden без ADR
  - любой security-инвариант отсутствует (webhook signature, WS RBAC, search scope, SSRF, секреты)
  - целостность данных нарушена (нет version/UUID/soft delete)
  - регрессия красная ИЛИ лога нет
  - тест flaky / недетерминирован
  - формат ошибки/пагинация нарушают контракт

ЖЁЛТЫЕ (MEDIUM, PASS возможен с фиксацией):
  - sleep в тесте, stale число в docs, отсутствие негативного TC, неоптимальный индекс

ЗЕЛЁНЫЙ (PASS):
  - чек-лист роли пройден пофактово, регрессия green+лог, TZ соблюдён, трассировка есть
```

### 9.7 Ссылки

- Аналитика (парный playbook): `.cursor/playbooks/analysis.md`
- Регламент: `.cursor/REGULATION.md` (§3 цикл, §4 принципы, §7 регрессия, §9 GSD, §10 triage)
- GSD-матрица: `.cursor/GSD-INTEGRATION.md`
- TZ: `docs/tz_constraints.md` (+ `.cursor/rules/60-tz-constraints.mdc`)
- Регрессия: `.cursor/rules/56-regression-verify.mdc`
- Docs sync: `.cursor/rules/55-docs-sync.mdc`
- Cleanup: `.cursor/rules/45-artifacts-cleanup.mdc`
- Test triage: `.cursor/skills/test-triage/SKILL.md`
- Skills: `.cursor/project_skills/BA_VER_skill.md`, `SA_VER_skill.md`, `ARCH_VER_skill.md`,
  `BE_VER_skill.md`, `FE_VER_skill.md`, `QA_VER_skill.md`
- Артефакты-cleanup: `.cursor/rules/45-artifacts-cleanup.mdc`
- Стиль кода (для be-ver/fe-ver/code-reviewer): `docs/code_style_guide.md`
- Security (для security-инвариантов): `docs/security_guidelines.md`

---

*Hub Verification Playbook — эталон для ba-ver/sa-ver/arch-ver/be-ver/fe-ver/qa-manual-ver/*
*qa-auto-ver/product-ver/designer-ver и code-reviewer.*
